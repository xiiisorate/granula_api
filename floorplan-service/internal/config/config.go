// Package config provides configuration for Floor Plan Service.
package config

import (
	"time"

	"github.com/xiiisorate/granula_api/shared/pkg/config"
)

// Config holds all configuration for Floor Plan Service.
type Config struct {
	// Service configuration
	Service ServiceConfig `mapstructure:"service"`

	// PostgreSQL for metadata
	Postgres config.PostgresConfig `mapstructure:"postgres"`

	// MinIO/S3 for file storage
	Storage StorageConfig `mapstructure:"storage"`

	// AI Service connection
	AIService AIServiceConfig `mapstructure:"ai_service"`

	// Logger configuration
	Logger LoggerConfig `mapstructure:"logger"`
}

// ServiceConfig holds service-specific settings.
type ServiceConfig struct {
	Name            string        `mapstructure:"name" default:"floorplan-service"`
	Version         string        `mapstructure:"version" default:"1.0.0"`
	GRPCPort        int           `mapstructure:"grpc_port" default:"50053"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout" default:"30s"`
}

// StorageConfig holds file storage settings.
type StorageConfig struct {
	// MinIO/S3 endpoint
	Endpoint string `mapstructure:"endpoint" default:"localhost:9000"`

	// Access credentials
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`

	// Bucket name
	Bucket string `mapstructure:"bucket" default:"floorplans"`

	// Use SSL
	UseSSL bool `mapstructure:"use_ssl" default:"false"`

	// Region
	Region string `mapstructure:"region" default:"us-east-1"`

	// Max file size in bytes (default 50MB)
	MaxFileSize int64 `mapstructure:"max_file_size" default:"52428800"`

	// Allowed MIME types
	AllowedTypes []string `mapstructure:"allowed_types"`
}

// AIServiceConfig holds AI service connection settings.
type AIServiceConfig struct {
	Address string        `mapstructure:"address" default:"localhost:50057"`
	Timeout time.Duration `mapstructure:"timeout" default:"120s"`
}

// LoggerConfig holds logger settings.
type LoggerConfig struct {
	Level  string `mapstructure:"level" default:"info"`
	Format string `mapstructure:"format" default:"json"`
}

// Load loads configuration.
func Load() (*Config, error) {
	cfg := &Config{}

	// Set defaults
	cfg.Service = ServiceConfig{
		Name:            "floorplan-service",
		Version:         "1.0.0",
		GRPCPort:        50053,
		ShutdownTimeout: 30 * time.Second,
	}

	cfg.Postgres = config.PostgresConfig{
		Host:         "localhost",
		Port:         5432,
		User:         "postgres",
		Password:     "postgres",
		Database:     "floorplans_db",
		SSLMode:      "disable",
		MaxOpenConns: 25,
	}

	cfg.Storage = StorageConfig{
		Endpoint:    "localhost:9000",
		Bucket:      "floorplans",
		UseSSL:      false,
		Region:      "us-east-1",
		MaxFileSize: 50 * 1024 * 1024,
		AllowedTypes: []string{
			"image/jpeg",
			"image/png",
			"image/webp",
			"application/pdf",
		},
	}

	cfg.AIService = AIServiceConfig{
		Address: "localhost:50057",
		Timeout: 120 * time.Second,
	}

	cfg.Logger = LoggerConfig{
		Level:  "info",
		Format: "json",
	}

	if err := config.LoadFromEnv(cfg, "FLOORPLAN"); err != nil {
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

