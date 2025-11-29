// =============================================================================
// Package config provides configuration management for Request Service.
// =============================================================================
// This package handles loading and validation of all configuration parameters
// required by the Request Service, including database connections,
// gRPC settings, and integration with other services.
//
// =============================================================================
package config

import (
	"fmt"

	"github.com/xiiisorate/granula_api/shared/pkg/config"
)

// Config holds all configuration parameters for Request Service.
type Config struct {
	// ==========================================================================
	// Application Settings
	// ==========================================================================

	// AppEnv specifies the running environment (development, staging, production).
	AppEnv string

	// LogLevel determines the minimum log level to output.
	LogLevel string

	// ==========================================================================
	// gRPC Server Settings
	// ==========================================================================

	// GRPCHost is the network interface to bind the gRPC server.
	GRPCHost string

	// GRPCPort is the TCP port for the gRPC server.
	// Default: 50059
	GRPCPort int

	// ==========================================================================
	// Database Settings (PostgreSQL)
	// ==========================================================================

	// Postgres contains all PostgreSQL connection parameters.
	Postgres config.PostgresConfig

	// ==========================================================================
	// Redis Settings
	// ==========================================================================

	// RedisURL is the connection string for Redis (for pub/sub events).
	RedisURL string

	// ==========================================================================
	// Service Discovery
	// ==========================================================================

	// NotificationServiceAddr is the gRPC address of the Notification Service.
	// Used for sending request status update notifications.
	NotificationServiceAddr string

	// WorkspaceServiceAddr is the gRPC address of the Workspace Service.
	// Used for validating workspace access.
	WorkspaceServiceAddr string
}

// Load reads configuration from environment variables.
func Load() *Config {
	return &Config{
		// Application
		AppEnv:   config.GetEnv("APP_ENV", "development"),
		LogLevel: config.GetEnv("LOG_LEVEL", "info"),

		// gRPC
		GRPCHost: config.GetEnv("GRPC_HOST", "0.0.0.0"),
		GRPCPort: config.GetEnvInt("GRPC_PORT", 50059),

		// PostgreSQL
		Postgres: config.PostgresConfig{
			Host:         config.GetEnv("POSTGRES_HOST", "localhost"),
			Port:         config.GetEnvInt("POSTGRES_PORT", 5432),
			User:         config.GetEnv("POSTGRES_USER", "granula"),
			Password:     config.GetEnv("POSTGRES_PASSWORD", "granula_secret"),
			Database:     config.GetEnv("POSTGRES_DATABASE", "requests_db"),
			SSLMode:      config.GetEnv("POSTGRES_SSL_MODE", "disable"),
			MaxOpenConns: config.GetEnvInt("POSTGRES_MAX_OPEN_CONNS", 25),
			MaxIdleConns: config.GetEnvInt("POSTGRES_MAX_IDLE_CONNS", 5),
		},

		// Redis
		RedisURL: config.GetEnv("REDIS_URL", "redis://localhost:6379"),

		// Service Discovery
		NotificationServiceAddr: config.GetEnv("NOTIFICATION_SERVICE_ADDR", "notification-service:50060"),
		WorkspaceServiceAddr:    config.GetEnv("WORKSPACE_SERVICE_ADDR", "workspace-service:50053"),
	}
}

// IsDevelopment returns true if running in development mode.
func (c *Config) IsDevelopment() bool {
	return c.AppEnv == "development"
}

// GRPCAddress returns the full gRPC server address.
func (c *Config) GRPCAddress() string {
	return fmt.Sprintf("%s:%d", c.GRPCHost, c.GRPCPort)
}
