// Package config provides configuration for AI Service.
package config

import (
	"time"

	"github.com/xiiisorate/granula_api/shared/pkg/config"
)

// Config holds all configuration for AI Service.
type Config struct {
	// Service configuration
	Service ServiceConfig `mapstructure:"service"`

	// OpenRouter API configuration
	OpenRouter config.OpenRouterConfig `mapstructure:"openrouter"`

	// MongoDB configuration
	MongoDB config.MongoConfig `mapstructure:"mongodb"`

	// Worker pool configuration
	Worker WorkerConfig `mapstructure:"worker"`

	// Logger configuration
	Logger LoggerConfig `mapstructure:"logger"`
}

// ServiceConfig holds service-specific settings.
type ServiceConfig struct {
	// Name of the service.
	Name string `mapstructure:"name" default:"ai-service"`

	// Version of the service.
	Version string `mapstructure:"version" default:"1.0.0"`

	// GRPCPort is the port for gRPC server.
	GRPCPort int `mapstructure:"grpc_port" default:"50057"`

	// ShutdownTimeout for graceful shutdown.
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout" default:"30s"`
}

// WorkerConfig holds worker pool settings.
type WorkerConfig struct {
	// PoolSize is the number of concurrent workers.
	PoolSize int `mapstructure:"pool_size" default:"5"`

	// QueueSize is the job queue buffer size.
	QueueSize int `mapstructure:"queue_size" default:"100"`

	// JobTimeout is the maximum time for a single job.
	JobTimeout time.Duration `mapstructure:"job_timeout" default:"120s"`
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
		Name:            "ai-service",
		Version:         "1.0.0",
		GRPCPort:        50057,
		ShutdownTimeout: 30 * time.Second,
	}

	cfg.OpenRouter = config.OpenRouterConfig{
		BaseURL:         "https://openrouter.ai/api/v1",
		Model:           "anthropic/claude-sonnet-4",
		Timeout:         60 * time.Second,
		MaxTokens:       4096,
		Temperature:     0.7,
		MaxRetries:      3,
		RateLimitPerMin: 60,
	}

	cfg.MongoDB = config.MongoConfig{
		URI:            "mongodb://localhost:27017",
		Database:       "ai_db",
		ConnectTimeout: 10 * time.Second,
		MaxPoolSize:    100,
	}

	cfg.Worker = WorkerConfig{
		PoolSize:   5,
		QueueSize:  100,
		JobTimeout: 120 * time.Second,
	}

	cfg.Logger = LoggerConfig{
		Level:  "info",
		Format: "json",
	}

	// Load from environment with AI_ prefix
	if err := config.LoadFromEnv(cfg, "AI"); err != nil {
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
