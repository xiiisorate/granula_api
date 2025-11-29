// Package main is the entry point for User Service.
package main

import (
	"fmt"
	"net"

	"github.com/xiiisorate/granula_api/user-service/internal/config"
	"github.com/xiiisorate/granula_api/user-service/internal/repository"
	"github.com/xiiisorate/granula_api/user-service/internal/server"
	"github.com/xiiisorate/granula_api/user-service/internal/service"
	"github.com/xiiisorate/granula_api/shared/pkg/logger"

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
		ServiceName: "user-service",
		Format:      "json",
		Development: cfg.AppEnv != "production",
	})
	logger.SetGlobal(log)

	log.Info("Starting User Service")

	// Connect to database
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.Name, cfg.DB.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database", logger.Err(err))
	}

	// Run migrations
	if err := repository.Migrate(db); err != nil {
		log.Fatal("Failed to run migrations", logger.Err(err))
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)

	// Initialize services
	userService := service.NewUserService(userRepo)

	// Create gRPC server
	grpcServer := grpc.NewServer()
	userServer := server.NewUserServer(userService)
	server.RegisterUserServiceServer(grpcServer, userServer)

	// Enable reflection for debugging
	reflection.Register(grpcServer)

	// Start server
	address := fmt.Sprintf("%s:%d", cfg.GRPC.Host, cfg.GRPC.Port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal("Failed to listen", logger.Err(err))
	}

	log.Info("User Service listening", logger.String("address", address))

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatal("Failed to serve", logger.Err(err))
	}
}

