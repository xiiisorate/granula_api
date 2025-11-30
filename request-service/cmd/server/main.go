// =============================================================================
// Request Service - Main Entry Point
// =============================================================================
// This is the main entry point for the Request Service, which handles
// expert service requests from users for BTI consultations.
//
// Port: 50059 (gRPC)
//
// =============================================================================
package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xiiisorate/granula_api/request-service/internal/config"
	grpcserver "github.com/xiiisorate/granula_api/request-service/internal/grpc"
	"github.com/xiiisorate/granula_api/request-service/internal/repository/postgres"
	"github.com/xiiisorate/granula_api/request-service/internal/service"
	requestpb "github.com/xiiisorate/granula_api/shared/gen/request/v1"
	"github.com/xiiisorate/granula_api/shared/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

const (
	serviceName     = "request-service"
	shutdownTimeout = 30 * time.Second
)

var Version = "dev"

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load configuration
	cfg := config.Load()

	// Initialize logger
	logCfg := logger.Config{
		Level:       cfg.LogLevel,
		Format:      "json",
		ServiceName: serviceName,
		Development: cfg.IsDevelopment(),
	}
	log, err := logger.New(logCfg)
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
	log.Info("starting request service",
		logger.String("version", Version),
		logger.String("env", cfg.AppEnv),
		logger.Int("port", cfg.GRPCPort),
	)

	// Connect to PostgreSQL
	pool, err := connectPostgres(ctx, cfg, log)
	if err != nil {
		log.Fatal("failed to connect to database", logger.Err(err))
	}
	defer pool.Close()

	// Initialize Notification Client (optional - continues without if unavailable)
	var notificationClient *grpcserver.NotificationClient
	notificationClient, err = grpcserver.NewNotificationClient(cfg.NotificationServiceAddr, log)
	if err != nil {
		log.Warn("notification service unavailable, notifications will be disabled",
			logger.String("address", cfg.NotificationServiceAddr),
			logger.Err(err),
		)
		notificationClient = nil
	} else {
		defer notificationClient.Close()
	}

	// Initialize layers
	requestRepo := postgres.NewRequestRepository(pool)
	requestService := service.NewRequestService(requestRepo, notificationClient, log)

	// Create gRPC server
	grpcSrv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			loggingInterceptor(log),
			recoveryInterceptor(log),
		),
	)

	// Register request service adapter (proto interface)
	requestAdapter := grpcserver.NewRequestServiceAdapter(requestService, log)
	requestpb.RegisterRequestServiceServer(grpcSrv, requestAdapter)

	// Register health check
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcSrv, healthServer)
	healthServer.SetServingStatus(serviceName, grpc_health_v1.HealthCheckResponse_SERVING)

	// Enable reflection for development
	if cfg.IsDevelopment() {
		reflection.Register(grpcSrv)
	}

	// Create listener
	listener, err := net.Listen("tcp", cfg.GRPCAddress())
	if err != nil {
		log.Fatal("failed to create listener", logger.Err(err))
	}

	// Start server
	serverErr := make(chan error, 1)
	go func() {
		log.Info("gRPC server starting", logger.String("address", cfg.GRPCAddress()))
		if err := grpcSrv.Serve(listener); err != nil {
			serverErr <- err
		}
	}()

	// Wait for shutdown signal
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		log.Error("server error", logger.Err(err))
	case sig := <-shutdown:
		log.Info("shutdown signal received", logger.String("signal", sig.String()))
	}

	// Graceful shutdown
	log.Info("initiating graceful shutdown")
	healthServer.SetServingStatus(serviceName, grpc_health_v1.HealthCheckResponse_NOT_SERVING)

	stopped := make(chan struct{})
	go func() {
		grpcSrv.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
		log.Info("server stopped gracefully")
	case <-time.After(shutdownTimeout):
		log.Warn("graceful shutdown timed out, forcing stop")
		grpcSrv.Stop()
	}

	log.Info("request service stopped")
}

func connectPostgres(ctx context.Context, cfg *config.Config, log *logger.Logger) (*pgxpool.Pool, error) {
	dsn := cfg.Postgres.DSN()
	log.Info("connecting to PostgreSQL",
		logger.String("host", cfg.Postgres.Host),
		logger.String("database", cfg.Postgres.Database),
	)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}

	poolConfig.MaxConns = int32(cfg.Postgres.MaxOpenConns)
	poolConfig.MinConns = int32(cfg.Postgres.MaxIdleConns)
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	log.Info("PostgreSQL connected successfully")
	return pool, nil
}

func loggingInterceptor(log *logger.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start)

		if err != nil {
			log.Warn("gRPC request failed",
				logger.String("method", info.FullMethod),
				logger.Int64("duration_ms", duration.Milliseconds()),
				logger.Err(err),
			)
		} else {
			log.Debug("gRPC request completed",
				logger.String("method", info.FullMethod),
				logger.Int64("duration_ms", duration.Milliseconds()),
			)
		}
		return resp, err
	}
}

func recoveryInterceptor(log *logger.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				log.Error("panic recovered", logger.F("panic", r))
				err = fmt.Errorf("internal server error")
			}
		}()
		return handler(ctx, req)
	}
}

