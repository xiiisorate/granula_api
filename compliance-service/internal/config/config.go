// Package config provides configuration for Compliance Service.
package config

import (
	"time"

	"github.com/xiiisorate/granula_api/shared/pkg/config"
)

// Config holds all configuration for Compliance Service.
type Config struct {
	// Service configuration
	Service ServiceConfig `mapstructure:"service"`

	// Database configuration
	Database config.PostgresConfig `mapstructure:"database"`

	// Logger configuration
	Logger LoggerConfig `mapstructure:"logger"`
}

// ServiceConfig holds service-specific settings.
type ServiceConfig struct {
	// Name of the service.
	Name string `mapstructure:"name" default:"compliance-service"`

	// Version of the service.
	Version string `mapstructure:"version" default:"1.0.0"`

	// GRPCPort is the port for gRPC server.
	GRPCPort int `mapstructure:"grpc_port" default:"50058"`

	// ShutdownTimeout for graceful shutdown.
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout" default:"10s"`
}

// LoggerConfig holds logger settings.
type LoggerConfig struct {
	// Level is the minimum log level (debug, info, warn, error).
	Level string `mapstructure:"level" default:"info"`

	// Format is the output format (json, console).
	Format string `mapstructure:"format" default:"json"`
}

// Load loads configuration from environment and files.
func Load() (*Config, error) {
	cfg := &Config{}

	// Set defaults
	cfg.Service = ServiceConfig{
		Name:            "compliance-service",
		Version:         "1.0.0",
		GRPCPort:        50058,
		ShutdownTimeout: 10 * time.Second,
	}

	cfg.Database = config.PostgresConfig{
		Host:         "localhost",
		Port:         5432,
		User:         "postgres",
		Password:     "",
		Database:     "compliance_db",
		SSLMode:      "disable",
		MaxOpenConns: 25,
		MaxIdleConns: 5,
		MaxLifetime:  5 * time.Minute,
	}

	cfg.Logger = LoggerConfig{
		Level:  "info",
		Format: "json",
	}

	// Load from environment with COMPLIANCE_ prefix
	if err := config.LoadFromEnv(cfg, "COMPLIANCE"); err != nil {
		return nil, err
	}

	return cfg, nil
}

// MustLoad loads configuration or panics.
func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		panic(err)
	}
	return cfg
}
