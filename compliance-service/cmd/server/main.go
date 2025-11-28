// Package main is the entry point for Compliance Service.
//
// Compliance Service checks floor plans and renovations against Russian building codes:
// - СНиП 31-01-2003 (residential buildings)
// - ЖК РФ (Housing Code)
// - СП 54.13330.2016 (actualized standards)
// - Regional regulations
//
// Run with: go run ./cmd/server
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xiiisorate/granula_api/compliance-service/internal/config"
	compliancegrpc "github.com/xiiisorate/granula_api/compliance-service/internal/grpc"
	"github.com/xiiisorate/granula_api/compliance-service/internal/repository/postgres"
	"github.com/xiiisorate/granula_api/compliance-service/internal/service"
	pb "github.com/xiiisorate/granula_api/shared/gen/compliance/v1"
	sharedgrpc "github.com/xiiisorate/granula_api/shared/pkg/grpc"
	"github.com/xiiisorate/granula_api/shared/pkg/logger"
)

func main() {
	// Load configuration
	cfg := config.MustLoad()

	// Initialize logger
	log, err := logger.New(logger.Config{
		Level:       cfg.Logger.Level,
		Format:      cfg.Logger.Format,
		ServiceName: cfg.Service.Name,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	log.Info("starting compliance service",
		logger.String("version", cfg.Service.Version),
		logger.Int("grpc_port", cfg.Service.GRPCPort),
	)

	// Connect to PostgreSQL
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := pgxpool.New(ctx, cfg.Database.DSN())
	if err != nil {
		log.Fatal("failed to connect to database", logger.Err(err))
	}
	defer pool.Close()

	// Ping database
	if err := pool.Ping(ctx); err != nil {
		log.Fatal("failed to ping database", logger.Err(err))
	}
	log.Info("connected to database",
		logger.String("host", cfg.Database.Host),
		logger.String("database", cfg.Database.Database),
	)

	// Initialize repositories
	ruleRepo := postgres.NewRuleRepository(pool)

	// Initialize service
	complianceService := service.NewComplianceService(ruleRepo, log)

	// Initialize gRPC server
	grpcServer, err := sharedgrpc.NewServer(sharedgrpc.ServerConfig{
		Port:             cfg.Service.GRPCPort,
		Logger:           log,
		ServiceName:      cfg.Service.Name,
		EnableReflection: true,
	})
	if err != nil {
		log.Fatal("failed to create gRPC server", logger.Err(err))
	}

	// Register compliance service
	complianceGRPCServer := compliancegrpc.NewComplianceServer(complianceService, log)
	pb.RegisterComplianceServiceServer(grpcServer.Server(), complianceGRPCServer)

	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- grpcServer.Start()
	}()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errChan:
		log.Fatal("server error", logger.Err(err))
	case sig := <-quit:
		log.Info("shutting down", logger.String("signal", sig.String()))
	}

	// Graceful shutdown
	grpcServer.Stop()
	log.Info("server stopped")
}
