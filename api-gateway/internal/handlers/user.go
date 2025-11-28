// Package handlers contains HTTP handlers for API Gateway.
package handlers

import (
	"github.com/xiiisorate/granula_api/api-gateway/internal/config"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// UpdateProfileInput represents profile update request.
type UpdateProfileInput struct {
	Name     *string                `json:"name,omitempty"`
	Settings *UserSettingsInput     `json:"settings,omitempty"`
}

// UserSettingsInput represents user settings update.
type UserSettingsInput struct {
	Language      *string                     `json:"language,omitempty"`
	Theme         *string                     `json:"theme,omitempty"`
	Units         *string                     `json:"units,omitempty"`
	Notifications *NotificationSettingsInput  `json:"notifications,omitempty"`
}

// NotificationSettingsInput represents notification settings update.
type NotificationSettingsInput struct {
	Email     *bool `json:"email,omitempty"`
	Push      *bool `json:"push,omitempty"`
	Marketing *bool `json:"marketing,omitempty"`
}

// ChangePasswordInput represents password change request.
type ChangePasswordInput struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// DeleteAccountInput represents account deletion request.
type DeleteAccountInput struct {
	Password string `json:"password"`
	Reason   string `json:"reason,omitempty"`
}

// GetProfile handles get profile request.
// @Summary Get current user profile
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /users/me [get]
func GetProfile(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Locals("userID").(uuid.UUID)

		// TODO: Call User Service via gRPC
		_ = userID
		return c.JSON(fiber.Map{
			"data": fiber.Map{
				"message": "GetProfile endpoint - gRPC integration pending",
			},
			"request_id": c.GetRespHeader("X-Request-ID"),
		})
	}
}

// UpdateProfile handles profile update request.
// @Summary Update current user profile
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body UpdateProfileInput true "Profile updates"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /users/me [patch]
func UpdateProfile(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Locals("userID").(uuid.UUID)

		var input UpdateProfileInput
		if err := c.BodyParser(&input); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
		}

		// TODO: Call User Service via gRPC
		_ = userID
		return c.JSON(fiber.Map{
			"data": fiber.Map{
				"message": "UpdateProfile endpoint - gRPC integration pending",
			},
			"request_id": c.GetRespHeader("X-Request-ID"),
		})
	}
}

// ChangePassword handles password change request.
// @Summary Change password
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body ChangePasswordInput true "Password change data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /users/me/password [put]
func ChangePassword(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Locals("userID").(uuid.UUID)

		var input ChangePasswordInput
		if err := c.BodyParser(&input); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
		}

		// TODO: Call User Service via gRPC
		_ = userID
		return c.JSON(fiber.Map{
			"data": fiber.Map{
				"message": "Password changed successfully",
			},
			"request_id": c.GetRespHeader("X-Request-ID"),
		})
	}
}

// DeleteAccount handles account deletion request.
// @Summary Delete account
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body DeleteAccountInput true "Deletion confirmation"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /users/me [delete]
func DeleteAccount(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Locals("userID").(uuid.UUID)

		var input DeleteAccountInput
		if err := c.BodyParser(&input); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
		}

		// TODO: Call User Service via gRPC
		_ = userID
		return c.JSON(fiber.Map{
			"data": fiber.Map{
				"message": "Account deleted successfully",
			},
			"request_id": c.GetRespHeader("X-Request-ID"),
		})
	}
}

