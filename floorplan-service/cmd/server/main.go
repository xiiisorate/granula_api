// Package main is the entry point for Floor Plan Service.
//
// Floor Plan Service handles:
// - Floor plan file upload and storage
// - Integration with AI Service for recognition
// - Thumbnail generation
// - Presigned URLs for secure downloads
//
// Uses PostgreSQL for metadata and MinIO/S3 for file storage.
// Run with: go run ./cmd/server
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xiiisorate/granula_api/floorplan-service/internal/config"
	fpgrpc "github.com/xiiisorate/granula_api/floorplan-service/internal/grpc"
	"github.com/xiiisorate/granula_api/floorplan-service/internal/repository/postgres"
	"github.com/xiiisorate/granula_api/floorplan-service/internal/service"
	"github.com/xiiisorate/granula_api/floorplan-service/internal/storage"
	aipb "github.com/xiiisorate/granula_api/shared/gen/ai/v1"
	pb "github.com/xiiisorate/granula_api/shared/gen/floorplan/v1"
	sharedgrpc "github.com/xiiisorate/granula_api/shared/pkg/grpc"
	"github.com/xiiisorate/granula_api/shared/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

	log.Info("starting Floor Plan service",
		logger.String("version", cfg.Service.Version),
		logger.Int("grpc_port", cfg.Service.GRPCPort),
	)

	// Connect to PostgreSQL
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	poolConfig, err := pgxpool.ParseConfig(cfg.Postgres.DSN())
	if err != nil {
		log.Fatal("failed to parse postgres config", logger.Err(err))
	}
	poolConfig.MaxConns = int32(cfg.Postgres.MaxOpenConns)

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Fatal("failed to connect to postgres", logger.Err(err))
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatal("failed to ping postgres", logger.Err(err))
	}
	log.Info("connected to PostgreSQL",
		logger.String("host", cfg.Postgres.Host),
		logger.String("database", cfg.Postgres.Database),
	)

	// Initialize MinIO storage
	minioStorage, err := storage.NewMinIOStorage(cfg.Storage, log)
	if err != nil {
		log.Fatal("failed to create MinIO storage", logger.Err(err))
	}

	if err := minioStorage.EnsureBucket(ctx); err != nil {
		log.Fatal("failed to ensure bucket", logger.Err(err))
	}
	log.Info("connected to MinIO",
		logger.String("endpoint", cfg.Storage.Endpoint),
		logger.String("bucket", cfg.Storage.Bucket),
	)

	// Connect to AI Service
	aiConn, err := grpc.NewClient(cfg.AIService.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal("failed to connect to AI service", logger.Err(err))
	}
	defer aiConn.Close()

	aiClient := aipb.NewAIServiceClient(aiConn)
	log.Info("connected to AI service", logger.String("address", cfg.AIService.Address))

	// Initialize repository
	repo := postgres.NewFloorPlanRepository(pool)

	// Initialize service
	svc := service.NewFloorPlanService(repo, minioStorage, aiClient, log)

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

	// Register service
	fpGRPCServer := fpgrpc.NewFloorPlanServer(svc, log)
	pb.RegisterFloorPlanServiceServer(grpcServer.Server(), fpGRPCServer)

	// Start server
	errChan := make(chan error, 1)
	go func() {
		errChan <- grpcServer.Start()
	}()

	// Wait for shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errChan:
		log.Fatal("server error", logger.Err(err))
	case sig := <-quit:
		log.Info("shutting down", logger.String("signal", sig.String()))
	}

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	grpcServer.Stop()

	// Close connections
	pool.Close()
	_ = aiConn.Close()

	_ = shutdownCtx

	log.Info("server stopped")
}

