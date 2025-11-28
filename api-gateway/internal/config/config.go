// Package config handles API Gateway configuration.
package config

import (
	sharedconfig "github.com/xiiisorate/granula_api/shared/pkg/config"
)

// Config holds all configuration for API Gateway.
type Config struct {
	AppEnv   string
	LogLevel string
	Host     string
	Port     int
	
	// Service addresses
	AuthServiceAddr         string
	UserServiceAddr         string
	NotificationServiceAddr string
	
	// JWT (for validation)
	JWTSecret string
}

// Load loads configuration from environment variables.
func Load() *Config {
	return &Config{
		AppEnv:   sharedconfig.GetEnv("APP_ENV", "development"),
		LogLevel: sharedconfig.GetEnv("LOG_LEVEL", "info"),
		Host:     sharedconfig.GetEnv("HOST", "0.0.0.0"),
		Port:     sharedconfig.GetEnvInt("PORT", 8090), // Non-standard port
		
		// Service addresses
		AuthServiceAddr:         sharedconfig.GetEnv("AUTH_SERVICE_ADDR", "auth-service:50051"),
		UserServiceAddr:         sharedconfig.GetEnv("USER_SERVICE_ADDR", "user-service:50052"),
		NotificationServiceAddr: sharedconfig.GetEnv("NOTIFICATION_SERVICE_ADDR", "notification-service:50060"),
		
		// JWT
		JWTSecret: sharedconfig.GetEnv("JWT_SECRET", ""),
	}
}

