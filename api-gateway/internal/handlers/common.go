// =============================================================================
// Package handlers provides HTTP handlers for API Gateway.
// =============================================================================
// Common utilities shared across all handlers including:
// - gRPC error handling and conversion to HTTP errors
// - Response formatting helpers
// - Input validation utilities
// =============================================================================
package handlers

import (
	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// =============================================================================
// gRPC Error Handling
// =============================================================================

// HandleGRPCError converts gRPC errors to appropriate HTTP/Fiber errors.
// This function maps gRPC status codes to HTTP status codes for consistent
// error handling across all API endpoints.
//
// Mapping:
//   - InvalidArgument → 400 Bad Request
//   - NotFound → 404 Not Found
//   - AlreadyExists → 409 Conflict
//   - Unauthenticated → 401 Unauthorized
//   - PermissionDenied → 403 Forbidden
//   - ResourceExhausted → 429 Too Many Requests
//   - Unavailable → 503 Service Unavailable
//   - DeadlineExceeded → 504 Gateway Timeout
//   - Default → 500 Internal Server Error
func HandleGRPCError(err error) error {
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

// =============================================================================
// Response Formatting Helpers
// =============================================================================

// SuccessResponseData creates a standard success response with data.
// This helper ensures consistent response format across all endpoints.
//
// Response format:
//
//	{
//	    "data": <data>,
//	    "request_id": "<request_id>"
//	}
func SuccessResponseData(c *fiber.Ctx, data interface{}) error {
	return c.JSON(fiber.Map{
		"data":       data,
		"request_id": c.GetRespHeader("X-Request-ID"),
	})
}

// SuccessResponseCreated creates a standard 201 Created response with data.
func SuccessResponseCreated(c *fiber.Ctx, data interface{}) error {
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data":       data,
		"request_id": c.GetRespHeader("X-Request-ID"),
	})
}

// SuccessResponseMessage creates a standard success response with message.
func SuccessResponseMessage(c *fiber.Ctx, message string) error {
	return c.JSON(fiber.Map{
		"message":    message,
		"request_id": c.GetRespHeader("X-Request-ID"),
	})
}

// SuccessResponseDataMessage creates a response with both data and message.
func SuccessResponseDataMessage(c *fiber.Ctx, data interface{}, message string) error {
	return c.JSON(fiber.Map{
		"data":       data,
		"message":    message,
		"request_id": c.GetRespHeader("X-Request-ID"),
	})
}

// =============================================================================
// User Context Helpers
// =============================================================================

// GetUserIDFromContext extracts user ID from the request context.
// The user ID is set by the auth middleware after JWT validation.
// Returns empty string if not found.
func GetUserIDFromContext(c *fiber.Ctx) string {
	if userID, ok := c.Locals("user_id").(string); ok {
		return userID
	}
	return ""
}

// GetUserIDFromContextRequired extracts user ID from the request context.
// Returns an error if user ID is not found.
func GetUserIDFromContextRequired(c *fiber.Ctx) (string, error) {
	userID := GetUserIDFromContext(c)
	if userID == "" {
		return "", fiber.NewError(fiber.StatusUnauthorized, "user not authenticated")
	}
	return userID, nil
}

// =============================================================================
// Pagination Helpers
// =============================================================================

// PaginationParams holds common pagination parameters.
type PaginationParams struct {
	Limit  int
	Offset int
	Page   int
}

// GetPaginationParams extracts pagination parameters from query string.
// Defaults: limit=20, offset=0 (or calculated from page)
func GetPaginationParams(c *fiber.Ctx) PaginationParams {
	limit := c.QueryInt("limit", 20)
	offset := c.QueryInt("offset", 0)
	page := c.QueryInt("page", 0)

	// If page is specified, calculate offset
	if page > 0 {
		offset = (page - 1) * limit
	}

	// Clamp values
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	return PaginationParams{
		Limit:  limit,
		Offset: offset,
		Page:   page,
	}
}

// PaginationResponse creates a standard pagination response object.
func PaginationResponse(page, limit, total int) fiber.Map {
	totalPages := 0
	if limit > 0 {
		totalPages = (total + limit - 1) / limit
	}

	return fiber.Map{
		"page":        page,
		"limit":       limit,
		"total":       total,
		"total_pages": totalPages,
	}
}
