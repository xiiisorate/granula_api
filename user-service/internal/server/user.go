// Package server implements gRPC server for User Service.
package server

import (
	"context"

	"github.com/xiiisorate/granula_api/user-service/internal/repository"
	"github.com/xiiisorate/granula_api/user-service/internal/service"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

// UserServiceServer is the interface for User gRPC service.
type UserServiceServer interface {
	GetProfile(ctx context.Context, req *GetProfileRequest) (*GetProfileResponse, error)
	UpdateProfile(ctx context.Context, req *UpdateProfileRequest) (*UpdateProfileResponse, error)
	ChangePassword(ctx context.Context, req *ChangePasswordRequest) (*ChangePasswordResponse, error)
	DeleteAccount(ctx context.Context, req *DeleteAccountRequest) (*DeleteAccountResponse, error)
}

// Request/Response types
type GetProfileRequest struct {
	UserID string
}

type GetProfileResponse struct {
	Profile *UserProfile
}

type UpdateProfileRequest struct {
	UserID   string
	Name     *string
	Settings *UserSettings
}

type UpdateProfileResponse struct {
	Profile *UserProfile
}

type ChangePasswordRequest struct {
	UserID          string
	CurrentPassword string
	NewPassword     string
	PasswordHash    string
}

type ChangePasswordResponse struct {
	Message string
}

type DeleteAccountRequest struct {
	UserID   string
	Password string
	Reason   string
}

type DeleteAccountResponse struct {
	Message string
}

type UserProfile struct {
	ID            string
	Email         string
	Name          string
	AvatarURL     string
	Role          string
	EmailVerified bool
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

// UserServer implements UserServiceServer.
type UserServer struct {
	userService *service.UserService
}

// NewUserServer creates a new UserServer.
func NewUserServer(userService *service.UserService) *UserServer {
	return &UserServer{userService: userService}
}

// GetProfile returns a user profile.
func (s *UserServer) GetProfile(ctx context.Context, req *GetProfileRequest) (*GetProfileResponse, error) {
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, err
	}

	profile, err := s.userService.GetProfile(userID)
	if err != nil {
		return nil, err
	}

	return &GetProfileResponse{
		Profile: profileToProto(profile),
	}, nil
}

// UpdateProfile updates a user profile.
func (s *UserServer) UpdateProfile(ctx context.Context, req *UpdateProfileRequest) (*UpdateProfileResponse, error) {
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, err
	}

	input := &service.UpdateProfileInput{
		Name: req.Name,
	}

	if req.Settings != nil {
		input.Settings = &repository.UserSettings{
			Language: req.Settings.Language,
			Theme:    req.Settings.Theme,
			Units:    req.Settings.Units,
		}
		if req.Settings.Notifications != nil {
			input.Settings.Notifications = &repository.NotificationSettings{
				Email:     req.Settings.Notifications.Email,
				Push:      req.Settings.Notifications.Push,
				Marketing: req.Settings.Notifications.Marketing,
			}
		}
	}

	profile, err := s.userService.UpdateProfile(userID, input)
	if err != nil {
		return nil, err
	}

	return &UpdateProfileResponse{
		Profile: profileToProto(profile),
	}, nil
}

// ChangePassword validates password change.
func (s *UserServer) ChangePassword(ctx context.Context, req *ChangePasswordRequest) (*ChangePasswordResponse, error) {
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, err
	}

	err = s.userService.ChangePassword(&service.ChangePasswordInput{
		UserID:          userID,
		CurrentPassword: req.CurrentPassword,
		NewPassword:     req.NewPassword,
		PasswordHash:    req.PasswordHash,
	})
	if err != nil {
		return nil, err
	}

	return &ChangePasswordResponse{
		Message: "Password changed successfully",
	}, nil
}

// DeleteAccount deletes a user account.
func (s *UserServer) DeleteAccount(ctx context.Context, req *DeleteAccountRequest) (*DeleteAccountResponse, error) {
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, err
	}

	if err := s.userService.DeleteAccount(userID); err != nil {
		return nil, err
	}

	return &DeleteAccountResponse{
		Message: "Account deleted successfully",
	}, nil
}

// RegisterUserServiceServer registers the user service server.
func RegisterUserServiceServer(s *grpc.Server, srv UserServiceServer) {
	// Will be generated from proto
}

// Helper function
func profileToProto(p *repository.UserProfile) *UserProfile {
	profile := &UserProfile{
		ID:            p.ID.String(),
		Email:         p.Email,
		Name:          p.Name,
		Role:          p.Role,
		EmailVerified: p.EmailVerified,
		CreatedAt:     p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if p.AvatarURL != nil {
		profile.AvatarURL = *p.AvatarURL
	}

	if p.Settings != nil {
		profile.Settings = &UserSettings{
			Language: p.Settings.Language,
			Theme:    p.Settings.Theme,
			Units:    p.Settings.Units,
		}
		if p.Settings.Notifications != nil {
			profile.Settings.Notifications = &NotificationSettings{
				Email:     p.Settings.Notifications.Email,
				Push:      p.Settings.Notifications.Push,
				Marketing: p.Settings.Notifications.Marketing,
			}
		}
	}

	return profile
}

