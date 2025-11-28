// Package main is the entry point for Workspace Service.
package main

import (
	"fmt"
	"net"

	"github.com/xiiisorate/granula_api/shared/pkg/logger"
	"github.com/xiiisorate/granula_api/workspace-service/internal/config"
	"github.com/xiiisorate/granula_api/workspace-service/internal/repository"
	"github.com/xiiisorate/granula_api/workspace-service/internal/server"
	"github.com/xiiisorate/granula_api/workspace-service/internal/service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize logger
	log := logger.MustNew(logger.Config{
		Level:       cfg.LogLevel,
		ServiceName: "workspace-service",
		Format:      "console",
		Development: cfg.AppEnv != "production",
	})
	logger.SetGlobal(log)

	log.Info("Starting Workspace Service")

	// Connect to database
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.Name, cfg.DB.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database", logger.Err(err))
	}

	log.Info("Connected to database", logger.String("database", cfg.DB.Name))

	// Run migrations
	if err := repository.Migrate(db); err != nil {
		log.Fatal("Failed to run migrations", logger.Err(err))
	}

	log.Info("Database migrations completed")

	// Initialize repositories
	workspaceRepo := repository.NewWorkspaceRepository(db)
	memberRepo := repository.NewMemberRepository(db)

	// Initialize services
	workspaceService := service.NewWorkspaceService(workspaceRepo, memberRepo)

	// Create gRPC server
	grpcServer := grpc.NewServer()
	workspaceServer := server.NewWorkspaceServer(workspaceService)
	server.RegisterWorkspaceServiceServer(grpcServer, workspaceServer)

	// Enable reflection for debugging
	reflection.Register(grpcServer)

	// Start server
	address := fmt.Sprintf("%s:%d", cfg.GRPC.Host, cfg.GRPC.Port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal("Failed to listen", logger.Err(err))
	}

	log.Info("Workspace Service listening", logger.String("address", address))

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatal("Failed to serve", logger.Err(err))
	}
}

