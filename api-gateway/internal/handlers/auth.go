// =============================================================================
// Package handlers contains HTTP handlers for API Gateway.
// =============================================================================
// Auth handlers manage authentication operations including registration,
// login, token refresh, and logout.
// =============================================================================
package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	authpb "github.com/xiiisorate/granula_api/shared/gen/auth/v1"
	userpb "github.com/xiiisorate/granula_api/shared/gen/user/v1"
	"github.com/xiiisorate/granula_api/shared/pkg/logger"
)

// AuthHandler handles authentication-related HTTP requests.
type AuthHandler struct {
	authClient authpb.AuthServiceClient
	userClient userpb.UserServiceClient
}

// NewAuthHandler creates a new AuthHandler with gRPC client connections.
func NewAuthHandler(authConn, userConn *grpc.ClientConn) *AuthHandler {
	return &AuthHandler{
		authClient: authpb.NewAuthServiceClient(authConn),
		userClient: userpb.NewUserServiceClient(userConn),
	}
}

// =============================================================================
// Request/Response Types
// =============================================================================

// RegisterInput represents registration request.
type RegisterInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Name     string `json:"name" validate:"required,min=2,max=100"`
}

// LoginInput represents login request.
type LoginInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
	DeviceID string `json:"device_id,omitempty"`
}

// RefreshInput represents refresh token request.
type RefreshInput struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// LogoutInput represents logout request.
type LogoutInput struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// TokenResponse represents authentication response with tokens.
type TokenResponse struct {
	UserID       string `json:"user_id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// =============================================================================
// Handlers
// =============================================================================

// Register handles user registration.
//
// @Summary Register a new user
// @Description Creates a new user account and returns authentication tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param body body RegisterInput true "Registration data"
// @Success 201 {object} map[string]interface{} "User created successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 409 {object} map[string]interface{} "Email already exists"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var input RegisterInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	// Validate input
	if input.Email == "" || input.Password == "" || input.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "email, password and name are required")
	}

	if len(input.Password) < 8 {
		return fiber.NewError(fiber.StatusBadRequest, "password must be at least 8 characters")
	}

	// Call Auth Service via gRPC
	ctx, cancel := context.WithTimeout(c.Context(), 15*time.Second)
	defer cancel()

	resp, err := h.authClient.Register(ctx, &authpb.RegisterRequest{
		Email:    input.Email,
		Password: input.Password,
		Name:     input.Name,
	})

	if err != nil {
		return handleGRPCError(err)
	}

	// Create user profile in User Service
	_, profileErr := h.userClient.CreateProfile(ctx, &userpb.CreateProfileRequest{
		UserId: resp.UserId,
		Email:  input.Email,
		Name:   input.Name,
		Role:   "user",
	})

	if profileErr != nil {
		// Log error but don't fail registration - profile can be created later
		logger.Global().Warn("failed to create user profile",
			logger.String("user_id", resp.UserId),
			logger.Err(profileErr),
		)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": TokenResponse{
			UserID:       resp.UserId,
			AccessToken:  resp.AccessToken,
			RefreshToken: resp.RefreshToken,
			ExpiresIn:    resp.ExpiresIn,
		},
		"message":    "User registered successfully",
		"request_id": c.GetRespHeader("X-Request-ID"),
	})
}

// Login handles user authentication.
//
// @Summary Login user
// @Description Authenticates user with email and password, returns tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param body body LoginInput true "Login credentials"
// @Success 200 {object} map[string]interface{} "Login successful"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 401 {object} map[string]interface{} "Invalid credentials"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var input LoginInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	if input.Email == "" || input.Password == "" {
		return fiber.NewError(fiber.StatusBadRequest, "email and password are required")
	}

	// Call Auth Service via gRPC
	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	resp, err := h.authClient.Login(ctx, &authpb.LoginRequest{
		Email:    input.Email,
		Password: input.Password,
		DeviceId: input.DeviceID,
	})

	if err != nil {
		return handleGRPCError(err)
	}

	return c.JSON(fiber.Map{
		"data": TokenResponse{
			UserID:       resp.UserId,
			AccessToken:  resp.AccessToken,
			RefreshToken: resp.RefreshToken,
			ExpiresIn:    resp.ExpiresIn,
		},
		"message":    "Login successful",
		"request_id": c.GetRespHeader("X-Request-ID"),
	})
}

// RefreshToken handles token refresh.
//
// @Summary Refresh access token
// @Description Exchanges refresh token for new access token
// @Tags auth
// @Accept json
// @Produce json
// @Param body body RefreshInput true "Refresh token"
// @Success 200 {object} map[string]interface{} "Tokens refreshed"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 401 {object} map[string]interface{} "Invalid or expired refresh token"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	var input RefreshInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	if input.RefreshToken == "" {
		return fiber.NewError(fiber.StatusBadRequest, "refresh_token is required")
	}

	// Call Auth Service via gRPC
	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	resp, err := h.authClient.RefreshToken(ctx, &authpb.RefreshTokenRequest{
		RefreshToken: input.RefreshToken,
	})

	if err != nil {
		return handleGRPCError(err)
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"access_token":  resp.AccessToken,
			"refresh_token": resp.RefreshToken,
			"expires_in":    resp.ExpiresIn,
		},
		"message":    "Tokens refreshed successfully",
		"request_id": c.GetRespHeader("X-Request-ID"),
	})
}

// Logout handles user logout.
//
// @Summary Logout user
// @Description Revokes the provided refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body LogoutInput true "Refresh token to revoke"
// @Success 200 {object} map[string]interface{} "Logout successful"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	var input LogoutInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	if input.RefreshToken == "" {
		return fiber.NewError(fiber.StatusBadRequest, "refresh_token is required")
	}

	// Call Auth Service via gRPC
	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	_, err := h.authClient.Logout(ctx, &authpb.LogoutRequest{
		RefreshToken: input.RefreshToken,
	})

	if err != nil {
		return handleGRPCError(err)
	}

	return c.JSON(fiber.Map{
		"message":    "Successfully logged out",
		"request_id": c.GetRespHeader("X-Request-ID"),
	})
}

// LogoutAll handles logout from all devices.
//
// @Summary Logout from all devices
// @Description Revokes all refresh tokens for the authenticated user
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Logged out from all devices"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /auth/logout-all [post]
func (h *AuthHandler) LogoutAll(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}

	// Call Auth Service via gRPC
	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	resp, err := h.authClient.LogoutAll(ctx, &authpb.LogoutAllRequest{
		UserId: userID.String(),
	})

	if err != nil {
		return handleGRPCError(err)
	}

	return c.JSON(fiber.Map{
		"message":       "Logged out from all devices",
		"revoked_count": resp.SessionsRevoked,
		"request_id":    c.GetRespHeader("X-Request-ID"),
	})
}

// =============================================================================
// Helper Functions
// =============================================================================

// handleGRPCError converts gRPC errors to HTTP errors.
func handleGRPCError(err error) error {
	st, ok := status.FromError(err)
	if !ok {
		return fiber.NewError(fiber.StatusInternalServerError, "internal server error")
	}

	switch st.Code() {
	case codes.InvalidArgument:
		return fiber.NewError(fiber.StatusBadRequest, st.Message())
	case codes.NotFound:
		return fiber.NewError(fiber.StatusNotFound, st.Message())
	case codes.AlreadyExists:
		return fiber.NewError(fiber.StatusConflict, st.Message())
	case codes.Unauthenticated:
		return fiber.NewError(fiber.StatusUnauthorized, st.Message())
	case codes.PermissionDenied:
		return fiber.NewError(fiber.StatusForbidden, st.Message())
	case codes.ResourceExhausted:
		return fiber.NewError(fiber.StatusTooManyRequests, st.Message())
	case codes.Unavailable:
		return fiber.NewError(fiber.StatusServiceUnavailable, "service temporarily unavailable")
	case codes.DeadlineExceeded:
		return fiber.NewError(fiber.StatusGatewayTimeout, "request timeout")
	default:
		return fiber.NewError(fiber.StatusInternalServerError, "internal server error")
	}
}
