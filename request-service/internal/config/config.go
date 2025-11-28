// Package config handles Request Service configuration.
package config

import (
	sharedconfig "github.com/xiiisorate/granula_api/shared/pkg/config"
)

// Config holds all configuration for Request Service.
type Config struct {
	AppEnv   string
	LogLevel string
	GRPC     GRPCConfig
	DB       DatabaseConfig
}

// GRPCConfig holds gRPC server configuration.
type GRPCConfig struct {
	Host string
	Port int
}

// DatabaseConfig holds database configuration.
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

// Load loads configuration from environment variables.
func Load() *Config {
	return &Config{
		AppEnv:   sharedconfig.GetEnv("APP_ENV", "development"),
		LogLevel: sharedconfig.GetEnv("LOG_LEVEL", "info"),
		GRPC: GRPCConfig{
			Host: sharedconfig.GetEnv("GRPC_HOST", "0.0.0.0"),
			Port: sharedconfig.GetEnvInt("GRPC_PORT", 50059),
		},
		DB: DatabaseConfig{
			Host:     sharedconfig.GetEnv("DB_HOST", "localhost"),
			Port:     sharedconfig.GetEnvInt("DB_PORT", 5432),
			User:     sharedconfig.GetEnv("DB_USER", "postgres"),
			Password: sharedconfig.GetEnv("DB_PASSWORD", "postgres"),
			Name:     sharedconfig.GetEnv("DB_NAME", "requests_db"),
			SSLMode:  sharedconfig.GetEnv("DB_SSLMODE", "disable"),
		},
	}
}

