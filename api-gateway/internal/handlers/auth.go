// Package handlers contains HTTP handlers for API Gateway.
package handlers

import (
	"github.com/xiiisorate/granula_api/api-gateway/internal/config"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// RegisterInput represents registration request.
type RegisterInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// LoginInput represents login request.
type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	DeviceID string `json:"device_id,omitempty"`
}

// RefreshInput represents refresh token request.
type RefreshInput struct {
	RefreshToken string `json:"refresh_token"`
}

// LogoutInput represents logout request.
type LogoutInput struct {
	RefreshToken string `json:"refresh_token"`
}

// Register handles user registration.
// @Summary Register a new user
// @Tags auth
// @Accept json
// @Produce json
// @Param body body RegisterInput true "Registration data"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Router /auth/register [post]
func Register(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input RegisterInput
		if err := c.BodyParser(&input); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
		}

		// TODO: Call Auth Service via gRPC
		// For now, return placeholder
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"data": fiber.Map{
				"message": "Registration endpoint - gRPC integration pending",
			},
			"request_id": c.GetRespHeader("X-Request-ID"),
		})
	}
}

// Login handles user login.
// @Summary Login user
// @Tags auth
// @Accept json
// @Produce json
// @Param body body LoginInput true "Login credentials"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /auth/login [post]
func Login(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input LoginInput
		if err := c.BodyParser(&input); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
		}

		// TODO: Call Auth Service via gRPC
		return c.JSON(fiber.Map{
			"data": fiber.Map{
				"message": "Login endpoint - gRPC integration pending",
			},
			"request_id": c.GetRespHeader("X-Request-ID"),
		})
	}
}

// RefreshToken handles token refresh.
// @Summary Refresh access token
// @Tags auth
// @Accept json
// @Produce json
// @Param body body RefreshInput true "Refresh token"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /auth/refresh [post]
func RefreshToken(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input RefreshInput
		if err := c.BodyParser(&input); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
		}

		// TODO: Call Auth Service via gRPC
		return c.JSON(fiber.Map{
			"data": fiber.Map{
				"message": "Refresh endpoint - gRPC integration pending",
			},
			"request_id": c.GetRespHeader("X-Request-ID"),
		})
	}
}

// Logout handles user logout.
// @Summary Logout user
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body LogoutInput true "Refresh token"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /auth/logout [post]
func Logout(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input LogoutInput
		if err := c.BodyParser(&input); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
		}

		// TODO: Call Auth Service via gRPC
		return c.JSON(fiber.Map{
			"data": fiber.Map{
				"message": "Successfully logged out",
			},
			"request_id": c.GetRespHeader("X-Request-ID"),
		})
	}
}

// LogoutAll handles logout from all devices.
// @Summary Logout from all devices
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /auth/logout-all [post]
func LogoutAll(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Locals("userID").(uuid.UUID)

		// TODO: Call Auth Service via gRPC
		_ = userID
		return c.JSON(fiber.Map{
			"data": fiber.Map{
				"message":       "Logged out from all devices",
				"revoked_count": 0,
			},
			"request_id": c.GetRespHeader("X-Request-ID"),
		})
	}
}

