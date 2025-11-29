// Package middleware contains HTTP middleware for API Gateway.
package middleware

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/xiiisorate/granula_api/api-gateway/internal/config"
	"github.com/xiiisorate/granula_api/shared/pkg/logger"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims represents JWT token claims.
type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Role   string    `json:"role"`
	jwt.RegisteredClaims
}

// Auth creates an authentication middleware.
func Auth(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "missing authorization header")
		}

		// Check Bearer prefix
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid authorization header format")
		}

		tokenString := parts[1]

		// Validate token
		claims, err := validateToken(tokenString, cfg.JWTSecret)
		if err != nil {
			// Log detailed error for debugging
			logger.Global().Error("JWT validation failed",
				logger.Err(err),
				logger.String("path", c.Path()),
				logger.String("method", c.Method()),
			)
			return fiber.NewError(fiber.StatusUnauthorized, "invalid token")
		}

		// Set user info in context
		// Using string keys that match what handlers expect
		c.Locals("userID", claims.UserID)           // uuid.UUID for backward compatibility
		c.Locals("user_id", claims.UserID.String()) // string for new handlers
		c.Locals("email", claims.Email)
		c.Locals("role", claims.Role)

		return c.Next()
	}
}

// validateToken validates a JWT token.
func validateToken(tokenString, secret string) (*Claims, error) {
	log := logger.Global()

	// Check if secret is configured
	if len(secret) == 0 {
		log.Error("JWT secret is empty!")
		return nil, errors.New("jwt secret not configured")
	}

	// Log secret length for debugging (never log the actual secret!)
	log.Info("JWT validation attempt",
		logger.Int("secret_length", len(secret)),
		logger.Int("token_length", len(tokenString)),
	)

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Error("Invalid JWT signing method",
				logger.String("method", token.Method.Alg()),
			)
			return nil, errors.New("invalid signing method")
		}
		return []byte(secret), nil
	})

	if err != nil {
		// Log specific JWT error - use Error level to ensure visibility in production
		log.Error("JWT parse error",
			logger.Err(err),
			logger.String("error_type", fmt.Sprintf("%T", err)),
		)
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		log.Error("Token claims invalid",
			logger.Bool("claims_ok", ok),
			logger.Bool("token_valid", token.Valid),
		)
		return nil, errors.New("invalid token claims")
	}

	// Log successful parsing
	log.Info("JWT parsed successfully",
		logger.String("user_id", claims.UserID.String()),
		logger.String("email", claims.Email),
	)

	// Check expiration (jwt library already checks this, but we log it)
	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
		log.Error("Token expired",
			logger.String("expires_at", claims.ExpiresAt.Time.String()),
		)
		return nil, errors.New("token expired")
	}

	return claims, nil
}
