// =============================================================================
// Package config provides configuration management for Workspace Service.
// =============================================================================
// This package handles loading and validation of all configuration parameters
// required by the Workspace Service, including database connections,
// gRPC settings, and service-specific options.
//
// Configuration is loaded from environment variables with sensible defaults
// for development environments.
//
// Example usage:
//
//	cfg := config.Load()
//	fmt.Printf("Starting server on port %d\n", cfg.GRPCPort)
//
// =============================================================================
package config

import (
	"fmt"

	"github.com/xiiisorate/granula_api/shared/pkg/config"
)

// Config holds all configuration parameters for Workspace Service.
// All fields are documented with their purpose, default values, and constraints.
type Config struct {
	// ==========================================================================
	// Application Settings
	// ==========================================================================

	// AppEnv specifies the running environment (development, staging, production).
	// Default: "development"
	// Affects: Logging verbosity, debug features, error details in responses.
	AppEnv string

	// LogLevel determines the minimum log level to output.
	// Valid values: "debug", "info", "warn", "error"
	// Default: "info"
	LogLevel string

	// ==========================================================================
	// gRPC Server Settings
	// ==========================================================================

	// GRPCHost is the network interface to bind the gRPC server.
	// Default: "0.0.0.0" (all interfaces)
	// Use "127.0.0.1" to restrict to localhost only.
	GRPCHost string

	// GRPCPort is the TCP port for the gRPC server.
	// Default: 50053
	// Must be unique across all services in the deployment.
	GRPCPort int

	// ==========================================================================
	// Database Settings (PostgreSQL)
	// ==========================================================================

	// Postgres contains all PostgreSQL connection parameters.
	// Connection string format: postgres://user:password@host:port/database?sslmode=disable
	Postgres config.PostgresConfig

	// ==========================================================================
	// Redis Settings (for caching and pub/sub)
	// ==========================================================================

	// RedisURL is the connection string for Redis.
	// Format: redis://[:password@]host:port[/database]
	// Default: "redis://localhost:6379"
	RedisURL string

	// ==========================================================================
	// Service Discovery (addresses of dependent services)
	// ==========================================================================

	// UserServiceAddr is the gRPC address of the User Service.
	// Used for fetching user details when listing workspace members.
	// Format: "host:port"
	// Default: "user-service:50052"
	UserServiceAddr string

	// NotificationServiceAddr is the gRPC address of the Notification Service.
	// Used for sending workspace-related notifications (invites, updates).
	// Format: "host:port"
	// Default: "notification-service:50060"
	NotificationServiceAddr string
}

// Load reads configuration from environment variables and returns a Config struct.
// Missing environment variables are replaced with sensible defaults suitable
// for local development.
//
// Environment Variables:
//   - APP_ENV: Application environment
//   - LOG_LEVEL: Logging level
//   - GRPC_HOST: gRPC server host
//   - GRPC_PORT: gRPC server port
//   - POSTGRES_*: PostgreSQL connection settings
//   - REDIS_URL: Redis connection URL
//   - USER_SERVICE_ADDR: User Service gRPC address
//   - NOTIFICATION_SERVICE_ADDR: Notification Service gRPC address
//
// Returns:
//   - *Config: Populated configuration struct ready for use
func Load() *Config {
	return &Config{
		// Application
		AppEnv:   config.GetEnv("APP_ENV", "development"),
		LogLevel: config.GetEnv("LOG_LEVEL", "info"),

		// gRPC
		GRPCHost: config.GetEnv("GRPC_HOST", "0.0.0.0"),
		GRPCPort: config.GetEnvInt("GRPC_PORT", 50053),

		// PostgreSQL
		Postgres: config.PostgresConfig{
			Host:         config.GetEnv("POSTGRES_HOST", "localhost"),
			Port:         config.GetEnvInt("POSTGRES_PORT", 5432),
			User:         config.GetEnv("POSTGRES_USER", "granula"),
			Password:     config.GetEnv("POSTGRES_PASSWORD", "granula_secret"),
			Database:     config.GetEnv("POSTGRES_DATABASE", "workspaces_db"),
			SSLMode:      config.GetEnv("POSTGRES_SSL_MODE", "disable"),
			MaxOpenConns: config.GetEnvInt("POSTGRES_MAX_OPEN_CONNS", 25),
			MaxIdleConns: config.GetEnvInt("POSTGRES_MAX_IDLE_CONNS", 5),
		},

		// Redis
		RedisURL: config.GetEnv("REDIS_URL", "redis://localhost:6379"),

		// Service Discovery
		UserServiceAddr:         config.GetEnv("USER_SERVICE_ADDR", "user-service:50052"),
		NotificationServiceAddr: config.GetEnv("NOTIFICATION_SERVICE_ADDR", "notification-service:50060"),
	}
}

// IsDevelopment returns true if the application is running in development mode.
// Useful for enabling debug features or verbose logging.
func (c *Config) IsDevelopment() bool {
	return c.AppEnv == "development"
}

// IsProduction returns true if the application is running in production mode.
// Useful for disabling debug features and enabling production optimizations.
func (c *Config) IsProduction() bool {
	return c.AppEnv == "production"
}

// GRPCAddress returns the full gRPC server address in "host:port" format.
// This is the address that the gRPC server will listen on.
func (c *Config) GRPCAddress() string {
	return fmt.Sprintf("%s:%d", c.GRPCHost, c.GRPCPort)
}
