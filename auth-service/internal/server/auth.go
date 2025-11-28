// Package server implements gRPC server for Auth Service.
package server

import (
	"context"

	"github.com/xiiisorate/granula_api/auth-service/internal/repository"
	"github.com/xiiisorate/granula_api/auth-service/internal/service"
	"github.com/xiiisorate/granula_api/shared/pkg/errors"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

// AuthServiceServer is the interface for Auth gRPC service.
type AuthServiceServer interface {
	Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error)
	Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error)
	ValidateToken(ctx context.Context, req *ValidateTokenRequest) (*ValidateTokenResponse, error)
	RefreshToken(ctx context.Context, req *RefreshTokenRequest) (*RefreshTokenResponse, error)
	Logout(ctx context.Context, req *LogoutRequest) (*LogoutResponse, error)
	LogoutAll(ctx context.Context, req *LogoutAllRequest) (*LogoutAllResponse, error)
}

// Request/Response types (matching proto)
type RegisterRequest struct {
	Email    string
	Password string
	Name     string
}

type RegisterResponse struct {
	User   *User
	Tokens *Tokens
}

type LoginRequest struct {
	Email    string
	Password string
	DeviceID string
}

type LoginResponse struct {
	User   *User
	Tokens *Tokens
}

type ValidateTokenRequest struct {
	AccessToken string
}

type ValidateTokenResponse struct {
	Valid  bool
	UserID string
	Email  string
	Role   string
}

type RefreshTokenRequest struct {
	RefreshToken string
}

type RefreshTokenResponse struct {
	Tokens *Tokens
}

type LogoutRequest struct {
	RefreshToken string
}

type LogoutResponse struct {
	Message string
}

type LogoutAllRequest struct{}

type LogoutAllResponse struct {
	RevokedCount int64
	Message      string
}

type User struct {
	ID            string
	Email         string
	Name          string
	Role          string
	EmailVerified bool
	AvatarURL     string
	Settings      *UserSettings
	CreatedAt     string
	UpdatedAt     string
}

type UserSettings struct {
	Language      string
	Theme         string
	Units         string
	Notifications *NotificationSettings
}

type NotificationSettings struct {
	Email     bool
	Push      bool
	Marketing bool
}

type Tokens struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    string
}

// AuthServer implements AuthServiceServer.
type AuthServer struct {
	authService *service.AuthService
}

// NewAuthServer creates a new AuthServer.
func NewAuthServer(authService *service.AuthService) *AuthServer {
	return &AuthServer{authService: authService}
}

// Register handles user registration.
func (s *AuthServer) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	result, err := s.authService.Register(&service.RegisterInput{
		Email:    req.Email,
		Password: req.Password,
		Name:     req.Name,
	})
	if err != nil {
		return nil, err
	}

	return &RegisterResponse{
		User:   userToProto(result.User),
		Tokens: tokensToProto(result),
	}, nil
}

// Login handles user login.
func (s *AuthServer) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	result, err := s.authService.Login(&service.LoginInput{
		Email:    req.Email,
		Password: req.Password,
		DeviceID: req.DeviceID,
	})
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		User:   userToProto(result.User),
		Tokens: tokensToProto(result),
	}, nil
}

// ValidateToken validates an access token.
func (s *AuthServer) ValidateToken(ctx context.Context, req *ValidateTokenRequest) (*ValidateTokenResponse, error) {
	claims, err := s.authService.ValidateToken(req.AccessToken)
	if err != nil {
		return &ValidateTokenResponse{Valid: false}, nil
	}

	return &ValidateTokenResponse{
		Valid:  true,
		UserID: claims.UserID.String(),
		Email:  claims.Email,
		Role:   claims.Role,
	}, nil
}

// RefreshToken refreshes an access token.
func (s *AuthServer) RefreshToken(ctx context.Context, req *RefreshTokenRequest) (*RefreshTokenResponse, error) {
	result, err := s.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		return nil, err
	}

	return &RefreshTokenResponse{
		Tokens: tokensToProto(result),
	}, nil
}

// Logout revokes a refresh token.
func (s *AuthServer) Logout(ctx context.Context, req *LogoutRequest) (*LogoutResponse, error) {
	if err := s.authService.Logout(req.RefreshToken); err != nil {
		return nil, err
	}

	return &LogoutResponse{Message: "Successfully logged out"}, nil
}

// LogoutAll revokes all refresh tokens for a user.
func (s *AuthServer) LogoutAll(ctx context.Context, req *LogoutAllRequest) (*LogoutAllResponse, error) {
	// Get user ID from context (set by auth middleware in gateway)
	userID, ok := ctx.Value("user_id").(uuid.UUID)
	if !ok {
		return nil, errors.Unauthenticated("user not authenticated")
	}

	count, err := s.authService.LogoutAll(userID)
	if err != nil {
		return nil, err
	}

	return &LogoutAllResponse{
		RevokedCount: count,
		Message:      "Logged out from all devices",
	}, nil
}

// RegisterAuthServiceServer registers the auth service server (stub for now).
func RegisterAuthServiceServer(s *grpc.Server, srv AuthServiceServer) {
	// Will be generated from proto
}

// Helper functions
func userToProto(user *repository.User) *User {
	return &User{
		ID:            user.ID.String(),
		Email:         user.Email,
		Name:          user.Name,
		Role:          user.Role,
		EmailVerified: user.EmailVerified,
		CreatedAt:     user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func tokensToProto(result *service.AuthResult) *Tokens {
	return &Tokens{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresAt:    result.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

