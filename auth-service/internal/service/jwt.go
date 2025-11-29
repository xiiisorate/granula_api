// Package service handles business logic for Auth Service.
package service

import (
	"errors"
	"time"

	"github.com/xiiisorate/granula_api/auth-service/internal/config"
	"github.com/xiiisorate/granula_api/auth-service/internal/repository"

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

// JWTService handles JWT operations.
type JWTService struct {
	secret        []byte
	accessExpire  time.Duration
	refreshExpire time.Duration
}

// NewJWTService creates a new JWTService.
func NewJWTService(cfg config.JWTConfig) *JWTService {
	return &JWTService{
		secret:        []byte(cfg.Secret),
		accessExpire:  cfg.AccessExpire,
		refreshExpire: cfg.RefreshExpire,
	}
}

// GenerateAccessToken generates an access token.
func (s *JWTService) GenerateAccessToken(user *repository.User) (string, time.Time, error) {
	expiresAt := time.Now().Add(s.accessExpire)

	claims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.secret)
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

// GenerateRefreshToken generates a refresh token.
// Each refresh token has a unique ID (JTI) to prevent duplicate key issues.
func (s *JWTService) GenerateRefreshToken(user *repository.User) (string, time.Time, error) {
	expiresAt := time.Now().Add(s.refreshExpire)

	claims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(), // Unique JWT ID to ensure uniqueness
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.secret)
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

// ValidateToken validates a token and returns claims.
func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return s.secret, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// GetAccessExpire returns access token expiration duration.
func (s *JWTService) GetAccessExpire() time.Duration {
	return s.accessExpire
}

// GetRefreshExpire returns refresh token expiration duration.
func (s *JWTService) GetRefreshExpire() time.Duration {
	return s.refreshExpire
}

