// Package service handles business logic for Auth Service.
package service

import (
	"strings"
	"time"

	"github.com/xiiisorate/granula_api/auth-service/internal/repository"
	"github.com/xiiisorate/granula_api/shared/pkg/errors"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// AuthService handles authentication business logic.
type AuthService struct {
	userRepo  *repository.UserRepository
	tokenRepo *repository.RefreshTokenRepository
	jwtSvc    *JWTService
}

// NewAuthService creates a new AuthService.
func NewAuthService(
	userRepo *repository.UserRepository,
	tokenRepo *repository.RefreshTokenRepository,
	jwtSvc *JWTService,
) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		jwtSvc:    jwtSvc,
	}
}

// RegisterInput contains registration data.
type RegisterInput struct {
	Email    string
	Password string
	Name     string
}

// LoginInput contains login data.
type LoginInput struct {
	Email    string
	Password string
	DeviceID string
}

// AuthResult contains authentication result.
type AuthResult struct {
	User         *repository.User
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

// Register registers a new user.
func (s *AuthService) Register(input *RegisterInput) (*AuthResult, error) {
	// Normalize email
	email := strings.ToLower(strings.TrimSpace(input.Email))

	// Validate email
	if !isValidEmail(email) {
		return nil, errors.InvalidArgument("email", "invalid email format")
	}

	// Check if email exists
	if s.userRepo.EmailExists(email) {
		return nil, errors.AlreadyExists("user", "email", email)
	}

	// Validate password
	if len(input.Password) < 8 {
		return nil, errors.InvalidArgument("password", "must be at least 8 characters")
	}

	// Validate name
	name := strings.TrimSpace(input.Name)
	if len(name) < 2 || len(name) > 255 {
		return nil, errors.InvalidArgument("name", "must be between 2 and 255 characters")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.Internal("failed to hash password").WithCause(err)
	}

	// Create user
	user := &repository.User{
		Email:        email,
		PasswordHash: string(hashedPassword),
		Name:         name,
		Role:         "user",
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, errors.Internal("failed to create user").WithCause(err)
	}

	// Generate tokens
	return s.generateTokens(user, nil, nil, nil)
}

// Login authenticates a user.
func (s *AuthService) Login(input *LoginInput) (*AuthResult, error) {
	// Normalize email
	email := strings.ToLower(strings.TrimSpace(input.Email))

	// Find user
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, errors.Unauthenticated("invalid email or password")
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, errors.Unauthenticated("invalid email or password")
	}

	// Generate tokens
	var deviceID *string
	if input.DeviceID != "" {
		deviceID = &input.DeviceID
	}

	return s.generateTokens(user, deviceID, nil, nil)
}

// ValidateToken validates an access token.
func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	return s.jwtSvc.ValidateToken(tokenString)
}

// RefreshToken refreshes an access token.
func (s *AuthService) RefreshToken(refreshTokenString string) (*AuthResult, error) {
	// Find token
	tokenRecord, err := s.tokenRepo.FindByToken(refreshTokenString)
	if err != nil || !tokenRecord.IsValid() {
		return nil, errors.Unauthenticated("invalid or expired refresh token")
	}

	// Revoke old token
	_ = s.tokenRepo.RevokeToken(refreshTokenString)

	// Generate new tokens
	return s.generateTokens(&tokenRecord.User, tokenRecord.DeviceID, tokenRecord.UserAgent, tokenRecord.IP)
}

// Logout revokes a refresh token.
func (s *AuthService) Logout(refreshTokenString string) error {
	return s.tokenRepo.RevokeToken(refreshTokenString)
}

// LogoutAll revokes all refresh tokens for a user.
func (s *AuthService) LogoutAll(userID uuid.UUID) (int64, error) {
	return s.tokenRepo.RevokeByUserID(userID)
}

// generateTokens generates access and refresh tokens.
func (s *AuthService) generateTokens(user *repository.User, deviceID, userAgent, ip *string) (*AuthResult, error) {
	// Generate access token
	accessToken, expiresAt, err := s.jwtSvc.GenerateAccessToken(user)
	if err != nil {
		return nil, errors.Internal("failed to generate access token").WithCause(err)
	}

	// Generate refresh token
	refreshToken, refreshExpiresAt, err := s.jwtSvc.GenerateRefreshToken(user)
	if err != nil {
		return nil, errors.Internal("failed to generate refresh token").WithCause(err)
	}

	// Save refresh token
	tokenRecord := &repository.RefreshToken{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: refreshExpiresAt,
		DeviceID:  deviceID,
		UserAgent: userAgent,
		IP:        ip,
	}

	if err := s.tokenRepo.Create(tokenRecord); err != nil {
		return nil, errors.Internal("failed to save refresh token").WithCause(err)
	}

	return &AuthResult{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	}, nil
}

// isValidEmail validates email format.
func isValidEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".") && len(email) >= 5
}

