// Package main is the entry point for Branch Service.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/xiiisorate/granula_api/branch-service/internal/config"
	branchgrpc "github.com/xiiisorate/granula_api/branch-service/internal/grpc"
	"github.com/xiiisorate/granula_api/branch-service/internal/repository/mongodb"
	"github.com/xiiisorate/granula_api/branch-service/internal/service"
	pb "github.com/xiiisorate/granula_api/shared/gen/branch/v1"
	sharedgrpc "github.com/xiiisorate/granula_api/shared/pkg/grpc"
	"github.com/xiiisorate/granula_api/shared/pkg/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	cfg := config.MustLoad()

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

	log.Info("starting Branch service",
		logger.String("version", cfg.Service.Version),
		logger.Int("grpc_port", cfg.Service.GRPCPort),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mongoClient, err := mongo.Connect(ctx, options.Client().
		ApplyURI(cfg.MongoDB.URI).
		SetMaxPoolSize(cfg.MongoDB.MaxPoolSize))
	if err != nil {
		log.Fatal("failed to connect to MongoDB", logger.Err(err))
	}
	defer mongoClient.Disconnect(ctx)

	if err := mongoClient.Ping(ctx, nil); err != nil {
		log.Fatal("failed to ping MongoDB", logger.Err(err))
	}
	log.Info("connected to MongoDB", logger.String("database", cfg.MongoDB.Database))

	db := mongoClient.Database(cfg.MongoDB.Database)
	repo := mongodb.NewBranchRepository(db)

	// Initialize Scene Client (optional - continues without if unavailable)
	var sceneClient *branchgrpc.SceneClient
	sceneClient, err = branchgrpc.NewSceneClient(cfg.Services.SceneServiceAddr, log)
	if err != nil {
		log.Warn("scene service unavailable, element copying will be disabled",
			logger.String("address", cfg.Services.SceneServiceAddr),
			logger.Err(err),
		)
		sceneClient = nil
	} else {
		defer sceneClient.Close()
	}

	svc := service.NewBranchService(repo, sceneClient, log)

	grpcServer, err := sharedgrpc.NewServer(sharedgrpc.ServerConfig{
		Port:             cfg.Service.GRPCPort,
		Logger:           log,
		ServiceName:      cfg.Service.Name,
		EnableReflection: true,
	})
	if err != nil {
		log.Fatal("failed to create gRPC server", logger.Err(err))
	}

	pb.RegisterBranchServiceServer(grpcServer.Server(), branchgrpc.NewBranchServer(svc, log))

	errChan := make(chan error, 1)
	go func() { errChan <- grpcServer.Start() }()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errChan:
		log.Fatal("server error", logger.Err(err))
	case sig := <-quit:
		log.Info("shutting down", logger.String("signal", sig.String()))
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	grpcServer.Stop()
	_ = mongoClient.Disconnect(shutdownCtx)

	log.Info("server stopped")
}

