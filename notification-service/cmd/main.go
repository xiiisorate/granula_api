// Package main is the entry point for Notification Service.
package main

import (
	"fmt"
	"net"

	"github.com/xiiisorate/granula_api/notification-service/internal/config"
	"github.com/xiiisorate/granula_api/notification-service/internal/repository"
	"github.com/xiiisorate/granula_api/notification-service/internal/server"
	"github.com/xiiisorate/granula_api/notification-service/internal/service"
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
	logger.Init(logger.Config{
		Level:       cfg.LogLevel,
		ServiceName: "notification-service",
		Pretty:      cfg.AppEnv != "production",
	})

	logger.Info("Starting Notification Service")

	// Connect to database
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.Name, cfg.DB.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Fatal("Failed to connect to database", err)
	}

	// Run migrations
	if err := repository.Migrate(db); err != nil {
		logger.Fatal("Failed to run migrations", err)
	}

	// Initialize repositories
	notifRepo := repository.NewNotificationRepository(db)

	// Initialize services
	notifService := service.NewNotificationService(notifRepo)

	// Create gRPC server
	grpcServer := grpc.NewServer()
	notifServer := server.NewNotificationServer(notifService)
	server.RegisterNotificationServiceServer(grpcServer, notifServer)

	// Enable reflection for debugging
	reflection.Register(grpcServer)

	// Start server
	address := fmt.Sprintf("%s:%d", cfg.GRPC.Host, cfg.GRPC.Port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		logger.Fatal("Failed to listen", err)
	}

	logger.Info(fmt.Sprintf("Notification Service listening on %s", address))

	if err := grpcServer.Serve(listener); err != nil {
		logger.Fatal("Failed to serve", err)
	}
}

