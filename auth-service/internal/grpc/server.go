// =============================================================================
// Package server implements gRPC server for Auth Service.
// =============================================================================
// This package provides the gRPC interface for authentication operations
// including user registration, login, token management, and logout.
// =============================================================================
package server

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/xiiisorate/granula_api/auth-service/internal/repository"
	"github.com/xiiisorate/granula_api/auth-service/internal/service"
	authpb "github.com/xiiisorate/granula_api/shared/gen/auth/v1"
)

// AuthServer implements authpb.AuthServiceServer.
type AuthServer struct {
	authpb.UnimplementedAuthServiceServer
	authService *service.AuthService
}

// NewAuthServer creates a new AuthServer.
func NewAuthServer(authService *service.AuthService) *AuthServer {
	return &AuthServer{authService: authService}
}

// Register handles user registration.
func (s *AuthServer) Register(ctx context.Context, req *authpb.RegisterRequest) (*authpb.RegisterResponse, error) {
	// Validate input
	if req.Email == "" || req.Password == "" || req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "email, password and name are required")
	}

	result, err := s.authService.Register(&service.RegisterInput{
		Email:    req.Email,
		Password: req.Password,
		Name:     req.Name,
	})
	if err != nil {
		return nil, convertError(err)
	}

	return &authpb.RegisterResponse{
		UserId:       result.User.ID.String(),
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresIn:    int64(result.ExpiresAt.Sub(result.User.CreatedAt).Seconds()),
	}, nil
}

// Login handles user authentication.
func (s *AuthServer) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.LoginResponse, error) {
	// Validate input
	if req.Email == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password are required")
	}

	result, err := s.authService.Login(&service.LoginInput{
		Email:    req.Email,
		Password: req.Password,
		DeviceID: req.DeviceId,
	})
	if err != nil {
		return nil, convertError(err)
	}

	return &authpb.LoginResponse{
		UserId:       result.User.ID.String(),
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresIn:    int64(result.ExpiresAt.Sub(result.User.CreatedAt).Seconds()),
	}, nil
}

// ValidateToken validates an access token.
func (s *AuthServer) ValidateToken(ctx context.Context, req *authpb.ValidateTokenRequest) (*authpb.ValidateTokenResponse, error) {
	if req.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "token is required")
	}

	claims, err := s.authService.ValidateToken(req.Token)
	if err != nil {
		return &authpb.ValidateTokenResponse{Valid: false}, nil
	}

	return &authpb.ValidateTokenResponse{
		Valid:  true,
		UserId: claims.UserID.String(),
		Roles:  []string{claims.Role},
	}, nil
}

// RefreshToken refreshes an access token.
func (s *AuthServer) RefreshToken(ctx context.Context, req *authpb.RefreshTokenRequest) (*authpb.RefreshTokenResponse, error) {
	if req.RefreshToken == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh_token is required")
	}

	result, err := s.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		return nil, convertError(err)
	}

	return &authpb.RefreshTokenResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresIn:    900, // 15 minutes
	}, nil
}

// Logout revokes a refresh token.
func (s *AuthServer) Logout(ctx context.Context, req *authpb.LogoutRequest) (*authpb.LogoutResponse, error) {
	if req.RefreshToken == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh_token is required")
	}

	if err := s.authService.Logout(req.RefreshToken); err != nil {
		return nil, convertError(err)
	}

	return &authpb.LogoutResponse{Success: true}, nil
}

// LogoutAll revokes all refresh tokens for a user.
func (s *AuthServer) LogoutAll(ctx context.Context, req *authpb.LogoutAllRequest) (*authpb.LogoutAllResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
	}

	count, err := s.authService.LogoutAll(userID)
	if err != nil {
		return nil, convertError(err)
	}

	return &authpb.LogoutAllResponse{
		Success:         true,
		SessionsRevoked: int32(count),
	}, nil
}

// ChangePassword changes user password.
func (s *AuthServer) ChangePassword(ctx context.Context, req *authpb.ChangePasswordRequest) (*authpb.ChangePasswordResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.CurrentPassword == "" {
		return nil, status.Error(codes.InvalidArgument, "current_password is required")
	}
	if req.NewPassword == "" {
		return nil, status.Error(codes.InvalidArgument, "new_password is required")
	}
	if len(req.NewPassword) < 8 {
		return nil, status.Error(codes.InvalidArgument, "new_password must be at least 8 characters")
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
	}

	if err := s.authService.ChangePassword(userID, req.CurrentPassword, req.NewPassword); err != nil {
		return nil, convertError(err)
	}

	return &authpb.ChangePasswordResponse{
		Success: true,
	}, nil
}

// =============================================================================
// Helper Functions
// =============================================================================

// convertError converts domain errors to gRPC status errors.
func convertError(err error) error {
	if err == nil {
		return nil
	}

	errStr := err.Error()

	// Check for error codes in the message (format: "CODE: message")
	switch {
	case strings.Contains(errStr, "NOT_FOUND") || strings.Contains(errStr, "user not found"):
		return status.Error(codes.NotFound, "user not found")
	case strings.Contains(errStr, "ALREADY_EXISTS") || strings.Contains(errStr, "email already"):
		return status.Error(codes.AlreadyExists, "email already registered")
	case strings.Contains(errStr, "invalid password") || strings.Contains(errStr, "invalid credentials"):
		return status.Error(codes.Unauthenticated, "invalid credentials")
	case strings.Contains(errStr, "invalid or expired token") || strings.Contains(errStr, "UNAUTHENTICATED"):
		return status.Error(codes.Unauthenticated, "invalid or expired token")
	case strings.Contains(errStr, "password is too weak") || strings.Contains(errStr, "min 8 characters"):
		return status.Error(codes.InvalidArgument, "password is too weak (min 8 characters)")
	case strings.Contains(errStr, "VALIDATION") || strings.Contains(errStr, "INVALID_ARGUMENT"):
		// Extract message after ":"
		if idx := strings.Index(errStr, ": "); idx > 0 {
			return status.Error(codes.InvalidArgument, errStr[idx+2:])
		}
		return status.Error(codes.InvalidArgument, errStr)
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}

// userToProto converts repository.User to protobuf User (helper for future use).
func userToProto(user *repository.User) map[string]interface{} {
	return map[string]interface{}{
		"id":             user.ID.String(),
		"email":          user.Email,
		"name":           user.Name,
		"role":           user.Role,
		"email_verified": user.EmailVerified,
		"created_at":     user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		"updated_at":     user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
