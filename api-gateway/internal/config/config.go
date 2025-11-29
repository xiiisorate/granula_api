// =============================================================================
// Package config provides configuration management for API Gateway.
// =============================================================================
// This package handles loading and validation of all configuration parameters
// required by the API Gateway, including HTTP server settings, gRPC client
// addresses, and security configurations.
//
// =============================================================================
package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all configuration parameters for API Gateway.
type Config struct {
	// ==========================================================================
	// Application Settings
	// ==========================================================================

	// AppEnv specifies the running environment (development, staging, production).
	AppEnv string

	// LogLevel determines the minimum log level to output.
	LogLevel string

	// ==========================================================================
	// HTTP Server Settings
	// ==========================================================================

	// HTTPHost is the network interface to bind the HTTP server.
	HTTPHost string

	// HTTPPort is the TCP port for the HTTP server.
	HTTPPort int

	// ReadTimeout is the maximum duration for reading the entire request.
	ReadTimeout time.Duration

	// WriteTimeout is the maximum duration before timing out writes of the response.
	WriteTimeout time.Duration

	// IdleTimeout is the maximum duration to wait for the next request.
	IdleTimeout time.Duration

	// ==========================================================================
	// CORS Settings
	// ==========================================================================

	// CORSAllowedOrigins is a list of allowed origins for CORS.
	CORSAllowedOrigins []string

	// CORSAllowCredentials indicates whether credentials are allowed.
	CORSAllowCredentials bool

	// ==========================================================================
	// Rate Limiting
	// ==========================================================================

	// RateLimitAnonymous is requests per minute for anonymous users.
	RateLimitAnonymous int

	// RateLimitAuthenticated is requests per minute for authenticated users.
	RateLimitAuthenticated int

	// RateLimitAI is requests per minute for AI endpoints.
	RateLimitAI int

	// ==========================================================================
	// JWT Settings
	// ==========================================================================

	// JWTSecret is the secret key for signing JWT tokens.
	JWTSecret string

	// JWTAccessTokenTTL is the lifetime of access tokens.
	JWTAccessTokenTTL time.Duration

	// JWTRefreshTokenTTL is the lifetime of refresh tokens.
	JWTRefreshTokenTTL time.Duration

	// ==========================================================================
	// Redis Settings (for rate limiting and caching)
	// ==========================================================================

	// RedisURL is the connection string for Redis.
	RedisURL string

	// ==========================================================================
	// Service Discovery (gRPC service addresses)
	// ==========================================================================

	// AuthServiceAddr is the gRPC address of the Auth Service.
	AuthServiceAddr string

	// UserServiceAddr is the gRPC address of the User Service.
	UserServiceAddr string

	// WorkspaceServiceAddr is the gRPC address of the Workspace Service.
	WorkspaceServiceAddr string

	// FloorPlanServiceAddr is the gRPC address of the Floor Plan Service.
	FloorPlanServiceAddr string

	// SceneServiceAddr is the gRPC address of the Scene Service.
	SceneServiceAddr string

	// BranchServiceAddr is the gRPC address of the Branch Service.
	BranchServiceAddr string

	// AIServiceAddr is the gRPC address of the AI Service.
	AIServiceAddr string

	// ComplianceServiceAddr is the gRPC address of the Compliance Service.
	ComplianceServiceAddr string

	// RequestServiceAddr is the gRPC address of the Request Service.
	RequestServiceAddr string

	// NotificationServiceAddr is the gRPC address of the Notification Service.
	NotificationServiceAddr string

	// ==========================================================================
	// Swagger Settings
	// ==========================================================================

	// SwaggerEnabled indicates whether Swagger UI is enabled.
	SwaggerEnabled bool

	// SwaggerBasePath is the base path for API documentation.
	SwaggerBasePath string
}

// Load reads configuration from environment variables.
func Load() *Config {
	return &Config{
		// Application
		AppEnv:   getEnv("APP_ENV", "development"),
		LogLevel: getEnv("LOG_LEVEL", "info"),

		// HTTP Server
		HTTPHost:     getEnv("HTTP_HOST", "0.0.0.0"),
		HTTPPort:     getEnvInt("HTTP_PORT", 8080),
		ReadTimeout:  time.Duration(getEnvInt("HTTP_READ_TIMEOUT", 30)) * time.Second,
		WriteTimeout: time.Duration(getEnvInt("HTTP_WRITE_TIMEOUT", 30)) * time.Second,
		IdleTimeout:  time.Duration(getEnvInt("HTTP_IDLE_TIMEOUT", 120)) * time.Second,

		// CORS
		CORSAllowedOrigins:   []string{getEnv("CORS_ALLOWED_ORIGINS", "*")},
		CORSAllowCredentials: getEnvBool("CORS_ALLOW_CREDENTIALS", true),

		// Rate Limiting
		RateLimitAnonymous:     getEnvInt("RATE_LIMIT_ANONYMOUS", 60),
		RateLimitAuthenticated: getEnvInt("RATE_LIMIT_AUTHENTICATED", 300),
		RateLimitAI:            getEnvInt("RATE_LIMIT_AI", 30),

		// JWT
		JWTSecret:          getEnv("JWT_SECRET", "your-super-secret-key-change-in-production"),
		JWTAccessTokenTTL:  time.Duration(getEnvInt("JWT_ACCESS_TOKEN_TTL", 15)) * time.Minute,
		JWTRefreshTokenTTL: time.Duration(getEnvInt("JWT_REFRESH_TOKEN_TTL", 7*24)) * time.Hour,

		// Redis
		RedisURL: getEnv("REDIS_URL", "redis://localhost:6379"),

		// Service Discovery
		AuthServiceAddr:         getEnv("AUTH_SERVICE_ADDR", "auth-service:50051"),
		UserServiceAddr:         getEnv("USER_SERVICE_ADDR", "user-service:50052"),
		WorkspaceServiceAddr:    getEnv("WORKSPACE_SERVICE_ADDR", "workspace-service:50053"),
		FloorPlanServiceAddr:    getEnv("FLOOR_PLAN_SERVICE_ADDR", "floorplan-service:50054"),
		SceneServiceAddr:        getEnv("SCENE_SERVICE_ADDR", "scene-service:50055"),
		BranchServiceAddr:       getEnv("BRANCH_SERVICE_ADDR", "branch-service:50056"),
		AIServiceAddr:           getEnv("AI_SERVICE_ADDR", "ai-service:50057"),
		ComplianceServiceAddr:   getEnv("COMPLIANCE_SERVICE_ADDR", "compliance-service:50058"),
		RequestServiceAddr:      getEnv("REQUEST_SERVICE_ADDR", "request-service:50059"),
		NotificationServiceAddr: getEnv("NOTIFICATION_SERVICE_ADDR", "notification-service:50060"),

		// Swagger
		SwaggerEnabled:  getEnvBool("SWAGGER_ENABLED", true),
		SwaggerBasePath: getEnv("SWAGGER_BASE_PATH", "/api/v1"),
	}
}

// IsDevelopment returns true if running in development mode.
func (c *Config) IsDevelopment() bool {
	return c.AppEnv == "development"
}

// IsProduction returns true if running in production mode.
func (c *Config) IsProduction() bool {
	return c.AppEnv == "production"
}

// HTTPAddress returns the full HTTP server address.
func (c *Config) HTTPAddress() string {
	return c.HTTPHost + ":" + strconv.Itoa(c.HTTPPort)
}

// =============================================================================
// Helper Functions
// =============================================================================

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
