// =============================================================================
// Package server implements gRPC server for User Service.
// =============================================================================
package server

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/xiiisorate/granula_api/user-service/internal/repository"
	"github.com/xiiisorate/granula_api/user-service/internal/service"
	userpb "github.com/xiiisorate/granula_api/shared/gen/user/v1"
)

// UserServer implements userpb.UserServiceServer.
type UserServer struct {
	userpb.UnimplementedUserServiceServer
	userService *service.UserService
}

// NewUserServer creates a new UserServer.
func NewUserServer(userService *service.UserService) *UserServer {
	return &UserServer{userService: userService}
}

// GetProfile returns a user profile.
func (s *UserServer) GetProfile(ctx context.Context, req *userpb.GetProfileRequest) (*userpb.GetProfileResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
	}

	profile, err := s.userService.GetProfile(userID)
	if err != nil {
		return nil, convertError(err)
	}

	return &userpb.GetProfileResponse{
		User: profileToProto(profile),
	}, nil
}

// UpdateProfile updates a user profile.
func (s *UserServer) UpdateProfile(ctx context.Context, req *userpb.UpdateProfileRequest) (*userpb.UpdateProfileResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
	}

	input := &service.UpdateProfileInput{}
	if req.Name != "" {
		input.Name = &req.Name
	}
	// Note: AvatarURL update is handled separately via UploadAvatar

	profile, err := s.userService.UpdateProfile(userID, input)
	if err != nil {
		return nil, convertError(err)
	}

	return &userpb.UpdateProfileResponse{
		User: profileToProto(profile),
	}, nil
}

// UploadAvatar uploads a user avatar.
func (s *UserServer) UploadAvatar(ctx context.Context, req *userpb.UploadAvatarRequest) (*userpb.UploadAvatarResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	// TODO: Implement avatar upload to storage
	return &userpb.UploadAvatarResponse{
		AvatarUrl: "https://storage.granula.ru/avatars/default.png",
	}, nil
}

// ChangePassword changes user password.
func (s *UserServer) ChangePassword(ctx context.Context, req *userpb.ChangePasswordRequest) (*userpb.ChangePasswordResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
	}

	err = s.userService.ChangePassword(&service.ChangePasswordInput{
		UserID:          userID,
		CurrentPassword: req.CurrentPassword,
		NewPassword:     req.NewPassword,
	})
	if err != nil {
		return nil, convertError(err)
	}

	return &userpb.ChangePasswordResponse{
		Success: true,
	}, nil
}

// DeleteUser deletes a user account.
func (s *UserServer) DeleteUser(ctx context.Context, req *userpb.DeleteUserRequest) (*userpb.DeleteUserResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
	}

	if err := s.userService.DeleteAccount(userID); err != nil {
		return nil, convertError(err)
	}

	return &userpb.DeleteUserResponse{
		Success: true,
	}, nil
}

// =============================================================================
// Helper Functions
// =============================================================================

func convertError(err error) error {
	if err == nil {
		return nil
	}

	switch err.Error() {
	case "user not found":
		return status.Error(codes.NotFound, err.Error())
	case "invalid password":
		return status.Error(codes.Unauthenticated, "current password is incorrect")
	case "new password must be different":
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}

func profileToProto(p *repository.UserProfile) *userpb.User {
	user := &userpb.User{
		Id:            p.ID.String(),
		Email:         p.Email,
		Name:          p.Name,
		Role:          p.Role,
		EmailVerified: p.EmailVerified,
		CreatedAt:     p.CreatedAt.Unix(),
		UpdatedAt:     p.UpdatedAt.Unix(),
	}

	if p.AvatarURL != nil {
		user.AvatarUrl = *p.AvatarURL
	}

	return user
}
