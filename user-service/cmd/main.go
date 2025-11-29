// =============================================================================
// Package main is the entry point for User Service.
// =============================================================================
package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/xiiisorate/granula_api/user-service/internal/config"
	"github.com/xiiisorate/granula_api/user-service/internal/repository"
	"github.com/xiiisorate/granula_api/user-service/internal/server"
	"github.com/xiiisorate/granula_api/user-service/internal/service"
	userpb "github.com/xiiisorate/granula_api/shared/gen/user/v1"
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

	log.Info("Starting User Service",
		logger.String("env", cfg.AppEnv),
		logger.String("version", "1.0.0"),
	)

	// Connect to database
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.Name, cfg.DB.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database", logger.Err(err))
	}

	log.Info("Connected to database",
		logger.String("host", cfg.DB.Host),
		logger.String("database", cfg.DB.Name),
	)

	// Run migrations
	if err := repository.Migrate(db); err != nil {
		log.Fatal("Failed to run migrations", logger.Err(err))
	}

	log.Info("Database migrations completed")

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)

	// Initialize services
	userService := service.NewUserService(userRepo)

	// Create gRPC server
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(loggingInterceptor(log)),
	)

	// Register user service
	userServer := server.NewUserServer(userService)
	userpb.RegisterUserServiceServer(grpcServer, userServer)

	// Enable reflection for debugging
	if cfg.AppEnv != "production" {
		reflection.Register(grpcServer)
		log.Info("gRPC reflection enabled (development mode)")
	}

	// Start server
	address := fmt.Sprintf("%s:%d", cfg.GRPC.Host, cfg.GRPC.Port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal("Failed to listen", logger.Err(err))
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Info("User Service listening", logger.String("address", address))
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatal("Failed to serve", logger.Err(err))
		}
	}()

	<-quit
	log.Info("Shutting down User Service...")

	grpcServer.GracefulStop()

	log.Info("User Service stopped")
}

// loggingInterceptor returns a gRPC interceptor that logs all requests.
func loggingInterceptor(log *logger.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		log.Debug("gRPC request",
			logger.String("method", info.FullMethod),
		)

		resp, err := handler(ctx, req)

		if err != nil {
			log.Error("gRPC error",
				logger.String("method", info.FullMethod),
				logger.Err(err),
			)
		}

		return resp, err
	}
}
