// =============================================================================
// Workspace Service - Main Entry Point
// =============================================================================
// This is the main entry point for the Workspace Service, one of the 11
// microservices in the Granula API architecture.
//
// Responsibilities:
//   - Initialize configuration from environment
//   - Set up logging with appropriate level
//   - Establish database connections with pooling
//   - Create service and repository instances
//   - Start gRPC server and handle graceful shutdown
//
// Environment Variables:
//   - APP_ENV: Application environment (development, staging, production)
//   - LOG_LEVEL: Logging level (debug, info, warn, error)
//   - GRPC_HOST: gRPC server host (default: 0.0.0.0)
//   - GRPC_PORT: gRPC server port (default: 50053)
//   - POSTGRES_DSN: PostgreSQL connection string
//   - REDIS_URL: Redis connection URL
//
// Health Check:
//   - gRPC health check enabled on same port
//   - HTTP health endpoint at /health (if enabled)
//
// Graceful Shutdown:
//   - Listens for SIGINT and SIGTERM
//   - Stops accepting new requests
//   - Waits for in-flight requests to complete
//   - Closes database connections cleanly
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
	workspacev1 "github.com/xiiisorate/granula_api/shared/gen/workspace/v1"
	"github.com/xiiisorate/granula_api/shared/pkg/logger"
	"github.com/xiiisorate/granula_api/workspace-service/internal/config"
	grpcserver "github.com/xiiisorate/granula_api/workspace-service/internal/grpc"
	"github.com/xiiisorate/granula_api/workspace-service/internal/repository/postgres"
	"github.com/xiiisorate/granula_api/workspace-service/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

// =============================================================================
// Application Constants
// =============================================================================

const (
	// serviceName is used for logging and health checks
	serviceName = "workspace-service"

	// shutdownTimeout is the maximum time to wait for graceful shutdown
	shutdownTimeout = 30 * time.Second
)

// Version is set at build time via ldflags
var Version = "dev"

// =============================================================================
// Main Function
// =============================================================================

func main() {
	// Initialize context for graceful shutdown
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
	log.Info("starting workspace service",
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

	// Initialize repository and service layers
	workspaceRepo := postgres.NewWorkspaceRepository(pool)
	workspaceService := service.NewWorkspaceService(workspaceRepo, log)

	// Create gRPC server
	grpcSrv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			loggingInterceptor(log),
			recoveryInterceptor(log),
		),
	)

	// Register workspace service
	workspaceServer := grpcserver.NewWorkspaceServer(workspaceService, log)
	workspacev1.RegisterWorkspaceServiceServer(grpcSrv, workspaceServer)

	// Register health check
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcSrv, healthServer)
	healthServer.SetServingStatus(serviceName, grpc_health_v1.HealthCheckResponse_SERVING)

	// Enable reflection for development
	if cfg.IsDevelopment() {
		reflection.Register(grpcSrv)
		log.Info("gRPC reflection enabled (development mode)")
	}

	// Create listener
	listener, err := net.Listen("tcp", cfg.GRPCAddress())
	if err != nil {
		log.Fatal("failed to create listener",
			logger.String("address", cfg.GRPCAddress()),
			logger.Err(err),
		)
	}

	// Start server in goroutine
	serverErr := make(chan error, 1)
	go func() {
		log.Info("gRPC server starting",
			logger.String("address", cfg.GRPCAddress()),
		)
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
		log.Info("shutdown signal received",
			logger.String("signal", sig.String()),
		)
	}

	// Graceful shutdown
	log.Info("initiating graceful shutdown",
		logger.Duration("timeout_ms", int64(shutdownTimeout.Milliseconds())),
	)

	// Mark as not serving
	healthServer.SetServingStatus(serviceName, grpc_health_v1.HealthCheckResponse_NOT_SERVING)

	// Stop accepting new connections
	stopped := make(chan struct{})
	go func() {
		grpcSrv.GracefulStop()
		close(stopped)
	}()

	// Wait for graceful stop or timeout
	select {
	case <-stopped:
		log.Info("server stopped gracefully")
	case <-time.After(shutdownTimeout):
		log.Warn("graceful shutdown timed out, forcing stop")
		grpcSrv.Stop()
	}

	log.Info("workspace service stopped")
}

// =============================================================================
// Database Connection
// =============================================================================

// connectPostgres establishes a connection pool to PostgreSQL.
// It verifies the connection by pinging the database.
func connectPostgres(ctx context.Context, cfg *config.Config, log *logger.Logger) (*pgxpool.Pool, error) {
	dsn := cfg.Postgres.DSN()
	log.Info("connecting to PostgreSQL",
		logger.String("host", cfg.Postgres.Host),
		logger.String("database", cfg.Postgres.Database),
	)

	// Parse config with pool settings
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}

	// Apply pool settings
	poolConfig.MaxConns = int32(cfg.Postgres.MaxOpenConns)
	poolConfig.MinConns = int32(cfg.Postgres.MaxIdleConns)
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute

	// Create pool
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	log.Info("PostgreSQL connected successfully",
		logger.Int("max_conns", cfg.Postgres.MaxOpenConns),
		logger.Int("min_conns", cfg.Postgres.MaxIdleConns),
	)

	return pool, nil
}

// maskDSN masks the password in a DSN for logging.
func maskDSN(dsn string) string {
	// Simple masking - in production use a proper parser
	return "postgres://***:***@***/***"
}

// =============================================================================
// gRPC Interceptors
// =============================================================================

// loggingInterceptor logs all gRPC requests.
func loggingInterceptor(log *logger.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		// Call handler
		resp, err := handler(ctx, req)

		// Log request
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

// recoveryInterceptor recovers from panics in handlers.
func recoveryInterceptor(log *logger.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				log.Error("panic recovered",
					logger.String("method", info.FullMethod),
					logger.F("panic", r),
				)
				err = fmt.Errorf("internal server error")
			}
		}()
		return handler(ctx, req)
	}
}
