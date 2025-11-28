// Package config provides configuration for Scene Service.
package config

import (
	"time"

	"github.com/xiiisorate/granula_api/shared/pkg/config"
)

// Config holds all configuration for Scene Service.
type Config struct {
	Service          ServiceConfig           `mapstructure:"service"`
	MongoDB          config.MongoConfig      `mapstructure:"mongodb"`
	ComplianceService ServiceConnectionConfig `mapstructure:"compliance_service"`
	Logger           LoggerConfig            `mapstructure:"logger"`
}

// ServiceConfig holds service-specific settings.
type ServiceConfig struct {
	Name            string        `mapstructure:"name" default:"scene-service"`
	Version         string        `mapstructure:"version" default:"1.0.0"`
	GRPCPort        int           `mapstructure:"grpc_port" default:"50054"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout" default:"30s"`
}

// ServiceConnectionConfig holds gRPC service connection settings.
type ServiceConnectionConfig struct {
	Address string        `mapstructure:"address" default:"localhost:50056"`
	Timeout time.Duration `mapstructure:"timeout" default:"30s"`
}

// LoggerConfig holds logger settings.
type LoggerConfig struct {
	Level  string `mapstructure:"level" default:"info"`
	Format string `mapstructure:"format" default:"json"`
}

// Load loads configuration.
func Load() (*Config, error) {
	cfg := &Config{}

	cfg.Service = ServiceConfig{
		Name:            "scene-service",
		Version:         "1.0.0",
		GRPCPort:        50054,
		ShutdownTimeout: 30 * time.Second,
	}

	cfg.MongoDB = config.MongoConfig{
		URI:            "mongodb://localhost:27017",
		Database:       "scenes_db",
		ConnectTimeout: 10 * time.Second,
		MaxPoolSize:    100,
	}

	cfg.ComplianceService = ServiceConnectionConfig{
		Address: "localhost:50056",
		Timeout: 30 * time.Second,
	}

	cfg.Logger = LoggerConfig{
		Level:  "info",
		Format: "json",
	}

	if err := config.LoadFromEnv(cfg, "SCENE"); err != nil {
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

