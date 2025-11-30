// Package config provides configuration loading for Granula microservices.
//
// Supports:
// - Environment variables
// - .env files
// - YAML/JSON config files
// - Type-safe configuration structs
//
// Priority (highest to lowest):
// 1. Environment variables
// 2. .env file
// 3. Config file (config.yaml)
// 4. Default values
//
// Example:
//
//	type AppConfig struct {
//	    Port int    `mapstructure:"port"`
//	    Host string `mapstructure:"host"`
//	}
//
//	var cfg AppConfig
//	if err := config.Load(&cfg, "APP"); err != nil {
//	    log.Fatal(err)
//	}
package config

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Load loads configuration into the provided struct.
// envPrefix is used to prefix environment variables (e.g., "APP" -> APP_PORT).
func Load(cfg any, envPrefix string) error {
	v := viper.New()

	// Set environment variable prefix
	if envPrefix != "" {
		v.SetEnvPrefix(envPrefix)
	}
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	// Look for config file
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.AddConfigPath("/etc/granula")

	// Try to read config file (optional)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Load .env file if exists
	if _, err := os.Stat(".env"); err == nil {
		envViper := viper.New()
		envViper.SetConfigFile(".env")
		envViper.SetConfigType("env")
		if err := envViper.ReadInConfig(); err == nil {
			for _, key := range envViper.AllKeys() {
				if !v.IsSet(key) {
					v.Set(key, envViper.Get(key))
				}
			}
		}
	}

	// Set defaults from struct tags
	setDefaults(v, cfg, "")

	// Unmarshal into config struct
	if err := v.Unmarshal(cfg); err != nil {
		return fmt.Errorf("error unmarshaling config: %w", err)
	}

	return nil
}

// LoadFromFile loads configuration from a specific file.
func LoadFromFile(cfg any, filepath string) error {
	v := viper.New()

	v.SetConfigFile(filepath)
	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("error reading config file: %w", err)
	}

	if err := v.Unmarshal(cfg); err != nil {
		return fmt.Errorf("error unmarshaling config: %w", err)
	}

	return nil
}

// LoadFromEnv loads configuration only from environment variables.
func LoadFromEnv(cfg any, envPrefix string) error {
	v := viper.New()

	if envPrefix != "" {
		v.SetEnvPrefix(envPrefix)
	}
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	// Set defaults from struct tags
	setDefaults(v, cfg, "")

	// Bind environment variables
	bindEnvs(v, cfg, envPrefix, "")

	if err := v.Unmarshal(cfg); err != nil {
		return fmt.Errorf("error unmarshaling config: %w", err)
	}

	return nil
}

// setDefaults sets default values from struct tags.
func setDefaults(v *viper.Viper, cfg any, prefix string) {
	val := reflect.ValueOf(cfg)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	typ := val.Type()

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.Field(i)

		// Get mapstructure tag
		tag := field.Tag.Get("mapstructure")
		if tag == "" || tag == "-" {
			continue
		}

		key := tag
		if prefix != "" {
			key = prefix + "." + tag
		}

		// Handle nested structs
		if fieldVal.Kind() == reflect.Struct && field.Type != reflect.TypeOf(time.Time{}) {
			setDefaults(v, fieldVal.Addr().Interface(), key)
			continue
		}

		// Set default value if specified
		if defaultVal := field.Tag.Get("default"); defaultVal != "" {
			v.SetDefault(key, defaultVal)
		}
	}
}

// bindEnvs binds environment variables to config keys.
func bindEnvs(v *viper.Viper, cfg any, envPrefix, prefix string) {
	val := reflect.ValueOf(cfg)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	typ := val.Type()

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.Field(i)

		tag := field.Tag.Get("mapstructure")
		if tag == "" || tag == "-" {
			continue
		}

		key := tag
		if prefix != "" {
			key = prefix + "." + tag
		}

		// Handle nested structs
		if fieldVal.Kind() == reflect.Struct && field.Type != reflect.TypeOf(time.Time{}) {
			bindEnvs(v, fieldVal.Addr().Interface(), envPrefix, key)
			continue
		}

		// Construct env var name
		envKey := strings.ToUpper(strings.ReplaceAll(key, ".", "_"))
		if envPrefix != "" {
			envKey = strings.ToUpper(envPrefix) + "_" + envKey
		}

		_ = v.BindEnv(key, envKey)
	}
}

// MustLoad loads configuration or panics on error.
func MustLoad(cfg any, envPrefix string) {
	if err := Load(cfg, envPrefix); err != nil {
		panic(err)
	}
}

// GetEnv returns environment variable with fallback default.
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEnvInt returns environment variable as int with fallback.
func GetEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var result int
		if _, err := fmt.Sscanf(value, "%d", &result); err == nil {
			return result
		}
	}
	return defaultValue
}

// GetEnvBool returns environment variable as bool with fallback.
func GetEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return strings.ToLower(value) == "true" || value == "1"
	}
	return defaultValue
}

// GetEnvDuration returns environment variable as duration with fallback.
func GetEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return defaultValue
}

// -----------------------------------------------------------------------------
// Common configuration types
// -----------------------------------------------------------------------------

// ServerConfig holds common server settings.
type ServerConfig struct {
	Host            string        `mapstructure:"host" default:"0.0.0.0"`
	Port            int           `mapstructure:"port" default:"8080"`
	GRPCPort        int           `mapstructure:"grpc_port" default:"50051"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout" default:"30s"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout" default:"30s"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout" default:"10s"`
}

// PostgresConfig holds PostgreSQL connection settings.
type PostgresConfig struct {
	Host         string        `mapstructure:"host" default:"localhost"`
	Port         int           `mapstructure:"port" default:"5432"`
	User         string        `mapstructure:"user" default:"postgres"`
	Password     string        `mapstructure:"password"`
	Database     string        `mapstructure:"database"`
	SSLMode      string        `mapstructure:"ssl_mode" default:"disable"`
	MaxOpenConns int           `mapstructure:"max_open_conns" default:"25"`
	MaxIdleConns int           `mapstructure:"max_idle_conns" default:"5"`
	MaxLifetime  time.Duration `mapstructure:"max_lifetime" default:"5m"`
}

// DSN returns PostgreSQL connection string.
func (c PostgresConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode,
	)
}

// MongoConfig holds MongoDB connection settings.
type MongoConfig struct {
	URI            string        `mapstructure:"uri" default:"mongodb://localhost:27017"`
	Database       string        `mapstructure:"database"`
	ConnectTimeout time.Duration `mapstructure:"connect_timeout" default:"10s"`
	MaxPoolSize    uint64        `mapstructure:"max_pool_size" default:"100"`
}

// RedisConfig holds Redis connection settings.
type RedisConfig struct {
	Addr         string        `mapstructure:"addr" default:"localhost:6379"`
	Password     string        `mapstructure:"password"`
	DB           int           `mapstructure:"db" default:"0"`
	PoolSize     int           `mapstructure:"pool_size" default:"10"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout" default:"3s"`
	WriteTimeout time.Duration `mapstructure:"write_timeout" default:"3s"`
}

// S3Config holds S3/MinIO settings.
type S3Config struct {
	Endpoint       string `mapstructure:"endpoint" default:"localhost:9000"`
	AccessKey      string `mapstructure:"access_key"`
	SecretKey      string `mapstructure:"secret_key"`
	Bucket         string `mapstructure:"bucket"`
	UseSSL         bool   `mapstructure:"use_ssl" default:"false"`
	Region         string `mapstructure:"region" default:"us-east-1"`
	ForcePathStyle bool   `mapstructure:"force_path_style" default:"true"`
}

// JWTConfig holds JWT settings.
type JWTConfig struct {
	Secret          string        `mapstructure:"secret"`
	AccessTokenTTL  time.Duration `mapstructure:"access_token_ttl" default:"15m"`
	RefreshTokenTTL time.Duration `mapstructure:"refresh_token_ttl" default:"168h"` // 7 days
	Issuer          string        `mapstructure:"issuer" default:"granula"`
}

// OpenRouterConfig holds OpenRouter API settings.
type OpenRouterConfig struct {
	APIKey          string        `mapstructure:"api_key"`
	Model           string        `mapstructure:"model" default:"anthropic/claude-sonnet-4"`
	BaseURL         string        `mapstructure:"base_url" default:"https://openrouter.ai/api/v1"`
	Timeout         time.Duration `mapstructure:"timeout" default:"60s"`
	VisionTimeout   time.Duration `mapstructure:"vision_timeout" default:"300s"` // Extended timeout for vision/multimodal
	MaxTokens       int           `mapstructure:"max_tokens" default:"4096"`
	Temperature     float64       `mapstructure:"temperature" default:"0.7"`
	MaxRetries      int           `mapstructure:"max_retries" default:"3"`
	RateLimitPerMin int           `mapstructure:"rate_limit_per_min" default:"60"`
}
