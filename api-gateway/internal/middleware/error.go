// Package middleware contains HTTP middleware for API Gateway.
package middleware

import (
	"github.com/xiiisorate/granula_api/shared/pkg/logger"

	"github.com/gofiber/fiber/v2"
)

// ErrorHandler handles errors and returns appropriate responses.
func ErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	errorCode := "internal_error"

	// Check if it's a Fiber error
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		switch code {
		case fiber.StatusBadRequest:
			errorCode = "bad_request"
		case fiber.StatusUnauthorized:
			errorCode = "unauthorized"
		case fiber.StatusForbidden:
			errorCode = "forbidden"
		case fiber.StatusNotFound:
			errorCode = "not_found"
		case fiber.StatusConflict:
			errorCode = "conflict"
		case fiber.StatusTooManyRequests:
			errorCode = "rate_limited"
		}
	}

	// Log error using global logger
	log := logger.Global()
	log.Error("request_error",
		logger.Err(err),
		logger.Int("status", code),
		logger.String("path", c.Path()),
		logger.String("method", c.Method()),
		logger.String("request_id", c.GetRespHeader("X-Request-ID")),
	)

	// Return generic error to client
	return c.Status(code).JSON(fiber.Map{
		"error":      errorCode,
		"request_id": c.GetRespHeader("X-Request-ID"),
	})
}

