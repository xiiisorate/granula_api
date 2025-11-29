// =============================================================================
// Package handlers contains HTTP handlers for API Gateway.
// =============================================================================
// User handlers manage user profile operations.
// =============================================================================
package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"google.golang.org/grpc"

	userpb "github.com/xiiisorate/granula_api/shared/gen/user/v1"
)

// UserHandler handles user-related HTTP requests.
type UserHandler struct {
	userClient userpb.UserServiceClient
}

// NewUserHandler creates a new UserHandler with gRPC client connection.
func NewUserHandler(conn *grpc.ClientConn) *UserHandler {
	return &UserHandler{
		userClient: userpb.NewUserServiceClient(conn),
	}
}

// =============================================================================
// Request/Response Types
// =============================================================================

// UpdateProfileInput represents profile update request.
type UpdateProfileInput struct {
	Name      string `json:"name,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

// ChangePasswordInput represents password change request.
type ChangePasswordInput struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
}

// =============================================================================
// Handlers
// =============================================================================

// GetProfile returns the current user's profile.
//
// @Summary Get user profile
// @Description Returns the authenticated user's profile information
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "User profile"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /users/me [get]
func (h *UserHandler) GetProfile(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	resp, err := h.userClient.GetProfile(ctx, &userpb.GetProfileRequest{
		UserId: userID.String(),
	})

	if err != nil {
		return handleGRPCError(err)
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"id":             resp.User.Id,
			"email":          resp.User.Email,
			"name":           resp.User.Name,
			"avatar_url":     resp.User.AvatarUrl,
			"email_verified": resp.User.EmailVerified,
			"role":           resp.User.Role,
			"created_at":     resp.User.CreatedAt,
			"updated_at":     resp.User.UpdatedAt,
		},
		"request_id": c.GetRespHeader("X-Request-ID"),
	})
}

// UpdateProfile updates the current user's profile.
//
// @Summary Update user profile
// @Description Updates the authenticated user's profile information
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body UpdateProfileInput true "Profile data to update"
// @Success 200 {object} map[string]interface{} "Profile updated"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /users/me [patch]
func (h *UserHandler) UpdateProfile(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}

	var input UpdateProfileInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	resp, err := h.userClient.UpdateProfile(ctx, &userpb.UpdateProfileRequest{
		UserId:    userID.String(),
		Name:      input.Name,
		AvatarUrl: input.AvatarURL,
	})

	if err != nil {
		return handleGRPCError(err)
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"id":         resp.User.Id,
			"email":      resp.User.Email,
			"name":       resp.User.Name,
			"avatar_url": resp.User.AvatarUrl,
			"updated_at": resp.User.UpdatedAt,
		},
		"message":    "Profile updated successfully",
		"request_id": c.GetRespHeader("X-Request-ID"),
	})
}

// ChangePassword changes the current user's password.
//
// @Summary Change password
// @Description Changes the authenticated user's password
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body ChangePasswordInput true "Password change data"
// @Success 200 {object} map[string]interface{} "Password changed"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 401 {object} map[string]interface{} "Invalid current password"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /users/me/password [put]
func (h *UserHandler) ChangePassword(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}

	var input ChangePasswordInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	if input.CurrentPassword == "" || input.NewPassword == "" {
		return fiber.NewError(fiber.StatusBadRequest, "current_password and new_password are required")
	}

	if len(input.NewPassword) < 8 {
		return fiber.NewError(fiber.StatusBadRequest, "new password must be at least 8 characters")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	_, err := h.userClient.ChangePassword(ctx, &userpb.ChangePasswordRequest{
		UserId:          userID.String(),
		CurrentPassword: input.CurrentPassword,
		NewPassword:     input.NewPassword,
	})

	if err != nil {
		return handleGRPCError(err)
	}

	return c.JSON(fiber.Map{
		"message":    "Password changed successfully",
		"request_id": c.GetRespHeader("X-Request-ID"),
	})
}

// DeleteAccount deletes the current user's account.
//
// @Summary Delete account
// @Description Permanently deletes the authenticated user's account
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Account deleted"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /users/me [delete]
func (h *UserHandler) DeleteAccount(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	_, err := h.userClient.DeleteUser(ctx, &userpb.DeleteUserRequest{
		UserId: userID.String(),
	})

	if err != nil {
		return handleGRPCError(err)
	}

	return c.JSON(fiber.Map{
		"message":    "Account deleted successfully",
		"request_id": c.GetRespHeader("X-Request-ID"),
	})
}

