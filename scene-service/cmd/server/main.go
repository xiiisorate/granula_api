// Package main is the entry point for Scene Service.
//
// Scene Service manages:
// - 3D scenes containing walls, rooms, furniture
// - Scene elements with hierarchical structure
// - Integration with Compliance Service
//
// Uses MongoDB for storage.
// Run with: go run ./cmd/server
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/xiiisorate/granula_api/scene-service/internal/config"
	scenegrpc "github.com/xiiisorate/granula_api/scene-service/internal/grpc"
	"github.com/xiiisorate/granula_api/scene-service/internal/repository/mongodb"
	"github.com/xiiisorate/granula_api/scene-service/internal/service"
	compliancepb "github.com/xiiisorate/granula_api/shared/gen/compliance/v1"
	pb "github.com/xiiisorate/granula_api/shared/gen/scene/v1"
	sharedgrpc "github.com/xiiisorate/granula_api/shared/pkg/grpc"
	"github.com/xiiisorate/granula_api/shared/pkg/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

	log.Info("starting Scene service",
		logger.String("version", cfg.Service.Version),
		logger.Int("grpc_port", cfg.Service.GRPCPort),
	)

	// Connect to MongoDB
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mongoCtx, mongoCancel := context.WithTimeout(ctx, cfg.MongoDB.ConnectTimeout)
	defer mongoCancel()

	mongoClient, err := mongo.Connect(mongoCtx, options.Client().
		ApplyURI(cfg.MongoDB.URI).
		SetMaxPoolSize(cfg.MongoDB.MaxPoolSize))
	if err != nil {
		log.Fatal("failed to connect to MongoDB", logger.Err(err))
	}
	defer mongoClient.Disconnect(ctx)

	if err := mongoClient.Ping(ctx, nil); err != nil {
		log.Fatal("failed to ping MongoDB", logger.Err(err))
	}
	log.Info("connected to MongoDB",
		logger.String("database", cfg.MongoDB.Database),
	)

	db := mongoClient.Database(cfg.MongoDB.Database)

	// Connect to Compliance Service
	complianceConn, err := grpc.NewClient(cfg.ComplianceService.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal("failed to connect to Compliance service", logger.Err(err))
	}
	defer complianceConn.Close()

	complianceClient := compliancepb.NewComplianceServiceClient(complianceConn)
	log.Info("connected to Compliance service",
		logger.String("address", cfg.ComplianceService.Address),
	)

	// Initialize repositories
	sceneRepo := mongodb.NewSceneRepository(db)
	elementRepo := mongodb.NewElementRepository(db)

	// Initialize service
	svc := service.NewSceneService(sceneRepo, elementRepo, complianceClient, log)

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
	sceneGRPCServer := scenegrpc.NewSceneServer(svc, log)
	pb.RegisterSceneServiceServer(grpcServer.Server(), sceneGRPCServer)

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

	if err := mongoClient.Disconnect(shutdownCtx); err != nil {
		log.Error("failed to disconnect MongoDB", logger.Err(err))
	}

	log.Info("server stopped")
}

