// Package config handles Auth Service configuration.
package config

import (
	"time"

	sharedconfig "github.com/xiiisorate/github.com/xiiisorate/github.com/xiiisorate/granula_api/shared/pkg/config"
)

// Config holds all configuration for Auth Service.
type Config struct {
	AppEnv   string
	LogLevel string
	GRPC     GRPCConfig
	DB       DatabaseConfig
	JWT      JWTConfig
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

// JWTConfig holds JWT configuration.
type JWTConfig struct {
	Secret        string
	AccessExpire  time.Duration
	RefreshExpire time.Duration
}

// Load loads configuration from environment variables.
func Load() *Config {
	return &Config{
		AppEnv:   sharedconfig.GetEnv("APP_ENV", "development"),
		LogLevel: sharedconfig.GetEnv("LOG_LEVEL", "info"),
		GRPC: GRPCConfig{
			Host: sharedconfig.GetEnv("GRPC_HOST", "0.0.0.0"),
			Port: sharedconfig.GetEnvInt("GRPC_PORT", 50051),
		},
		DB: DatabaseConfig{
			Host:     sharedconfig.GetEnv("DB_HOST", "localhost"),
			Port:     sharedconfig.GetEnvInt("DB_PORT", 5432),
			User:     sharedconfig.GetEnv("DB_USER", "postgres"),
			Password: sharedconfig.GetEnv("DB_PASSWORD", "postgres"),
			Name:     sharedconfig.GetEnv("DB_NAME", "auth_db"),
			SSLMode:  sharedconfig.GetEnv("DB_SSLMODE", "disable"),
		},
		JWT: JWTConfig{
			Secret:        sharedconfig.GetEnv("JWT_SECRET", ""),
			AccessExpire:  sharedconfig.GetEnvDuration("JWT_ACCESS_EXPIRE", 15*time.Minute),
			RefreshExpire: sharedconfig.GetEnvDuration("JWT_REFRESH_EXPIRE", 7*24*time.Hour),
		},
	}
}

