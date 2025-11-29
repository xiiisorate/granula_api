// Package main is the entry point for AI Service.
//
// AI Service provides:
// - Floor plan recognition from images
// - Layout variant generation
// - Interactive chat with AI assistant
// - Streaming responses for real-time UX
//
// Uses OpenRouter API for LLM integration.
// Run with: go run ./cmd/server
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/xiiisorate/granula_api/ai-service/internal/config"
	aigrpc "github.com/xiiisorate/granula_api/ai-service/internal/grpc"
	"github.com/xiiisorate/granula_api/ai-service/internal/openrouter"
	"github.com/xiiisorate/granula_api/ai-service/internal/repository/mongodb"
	"github.com/xiiisorate/granula_api/ai-service/internal/service"
	pb "github.com/xiiisorate/granula_api/shared/gen/ai/v1"
	sharedgrpc "github.com/xiiisorate/granula_api/shared/pkg/grpc"
	"github.com/xiiisorate/granula_api/shared/pkg/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

	log.Info("starting AI service",
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

	// Ping MongoDB
	if err := mongoClient.Ping(ctx, nil); err != nil {
		log.Fatal("failed to ping MongoDB", logger.Err(err))
	}
	log.Info("connected to MongoDB",
		logger.String("database", cfg.MongoDB.Database),
	)

	db := mongoClient.Database(cfg.MongoDB.Database)

	// Validate OpenRouter API key
	if cfg.OpenRouter.APIKey == "" {
		log.Warn("OpenRouter API key not set, AI features will not work")
	}

	// Initialize OpenRouter client
	openRouterClient := openrouter.NewClient(cfg.OpenRouter, log)

	// Initialize repositories
	chatRepo := mongodb.NewChatRepository(db)
	jobRepo := mongodb.NewJobRepository(db)

	// Initialize Scene Service client (optional - for context integration)
	var sceneClient *aigrpc.SceneClient
	if cfg.Service.SceneServiceAddr != "" {
		var err error
		sceneClient, err = aigrpc.NewSceneClient(cfg.Service.SceneServiceAddr, log)
		if err != nil {
			log.Warn("failed to connect to Scene Service, context integration disabled",
				logger.Err(err),
				logger.String("addr", cfg.Service.SceneServiceAddr),
			)
			sceneClient = nil
		} else {
			log.Info("connected to Scene Service",
				logger.String("addr", cfg.Service.SceneServiceAddr),
			)
			defer sceneClient.Close()
		}
	} else {
		log.Info("Scene Service integration disabled (no address configured)")
	}

	// Initialize services with Scene Service integration
	chatService := service.NewChatService(chatRepo, openRouterClient, sceneClient, log)
	recognitionService := service.NewRecognitionService(jobRepo, openRouterClient, log)
	generationService := service.NewGenerationService(jobRepo, openRouterClient, log)

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

	// Register AI service with Scene Client for context
	aiGRPCServer := aigrpc.NewAIServer(chatService, recognitionService, generationService, sceneClient, log)
	pb.RegisterAIServiceServer(grpcServer.Server(), aiGRPCServer)

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
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	grpcServer.Stop()

	// Close MongoDB
	if err := mongoClient.Disconnect(shutdownCtx); err != nil {
		log.Error("failed to disconnect MongoDB", logger.Err(err))
	}

	log.Info("server stopped")
}
