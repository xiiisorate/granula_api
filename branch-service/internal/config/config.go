// Package config provides configuration for Branch Service.
package config

import (
	"time"

	"github.com/xiiisorate/granula_api/shared/pkg/config"
)

// Config holds all configuration for Branch Service.
type Config struct {
	Service  ServiceConfig      `mapstructure:"service"`
	MongoDB  config.MongoConfig `mapstructure:"mongodb"`
	Logger   LoggerConfig       `mapstructure:"logger"`
	Services ServicesConfig     `mapstructure:"services"`
}

// ServicesConfig holds addresses of dependent services.
type ServicesConfig struct {
	// SceneServiceAddr is the gRPC address of the Scene Service.
	// Used for copying elements between branches.
	// Format: "host:port"
	// Default: "scene-service:50055"
	SceneServiceAddr string `mapstructure:"scene_service_addr" default:"scene-service:50055"`
}

// ServiceConfig holds service-specific settings.
type ServiceConfig struct {
	Name            string        `mapstructure:"name" default:"branch-service"`
	Version         string        `mapstructure:"version" default:"1.0.0"`
	GRPCPort        int           `mapstructure:"grpc_port" default:"50055"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout" default:"30s"`
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
		Name:            "branch-service",
		Version:         "1.0.0",
		GRPCPort:        50055,
		ShutdownTimeout: 30 * time.Second,
	}
	cfg.MongoDB = config.MongoConfig{
		URI:            "mongodb://localhost:27017",
		Database:       "branches_db",
		ConnectTimeout: 10 * time.Second,
		MaxPoolSize:    100,
	}
	cfg.Logger = LoggerConfig{Level: "info", Format: "json"}
	cfg.Services = ServicesConfig{
		SceneServiceAddr: "scene-service:50055",
	}

	if err := config.LoadFromEnv(cfg, "BRANCH"); err != nil {
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

