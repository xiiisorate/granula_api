// Package handlers contains HTTP handlers for API Gateway.
package handlers

import (
	"strconv"

	"github.com/xiiisorate/granula_api/api-gateway/internal/config"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// MarkAllAsReadInput represents mark all as read request.
type MarkAllAsReadInput struct {
	Type   *string `json:"type,omitempty"`
	Before *string `json:"before,omitempty"`
}

// GetNotifications handles get notifications request.
// @Summary Get notifications list
// @Tags notifications
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit" default(50)
// @Param offset query int false "Offset" default(0)
// @Param unread_only query bool false "Unread only" default(false)
// @Param type query string false "Filter by type"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /notifications [get]
func GetNotifications(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Locals("userID").(uuid.UUID)
		limit, _ := strconv.Atoi(c.Query("limit", "50"))
		offset, _ := strconv.Atoi(c.Query("offset", "0"))
		unreadOnly := c.Query("unread_only") == "true"
		notifType := c.Query("type")

		// TODO: Call Notification Service via gRPC
		_, _, _, _ = userID, limit, offset, unreadOnly
		_ = notifType

		return c.JSON(fiber.Map{
			"data": fiber.Map{
				"notifications": []interface{}{},
				"unread_count":  0,
				"total":         0,
			},
			"request_id": c.GetRespHeader("X-Request-ID"),
		})
	}
}

// GetUnreadCount handles get unread count request.
// @Summary Get unread notifications count
// @Tags notifications
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /notifications/count [get]
func GetUnreadCount(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Locals("userID").(uuid.UUID)

		// TODO: Call Notification Service via gRPC
		_ = userID

		return c.JSON(fiber.Map{
			"data": fiber.Map{
				"unread_count": 0,
				"by_type":      map[string]int{},
			},
			"request_id": c.GetRespHeader("X-Request-ID"),
		})
	}
}

// MarkAsRead handles mark notification as read request.
// @Summary Mark notification as read
// @Tags notifications
// @Produce json
// @Security BearerAuth
// @Param id path string true "Notification ID"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /notifications/{id}/read [post]
func MarkAsRead(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Locals("userID").(uuid.UUID)
		notificationID := c.Params("id")

		// TODO: Call Notification Service via gRPC
		_, _ = userID, notificationID

		return c.JSON(fiber.Map{
			"data": fiber.Map{
				"id":      notificationID,
				"read":    true,
				"read_at": "2024-01-01T00:00:00Z",
			},
			"request_id": c.GetRespHeader("X-Request-ID"),
		})
	}
}

// MarkAllAsRead handles mark all notifications as read request.
// @Summary Mark all notifications as read
// @Tags notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body MarkAllAsReadInput false "Filter options"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /notifications/read-all [post]
func MarkAllAsRead(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Locals("userID").(uuid.UUID)

		var input MarkAllAsReadInput
		_ = c.BodyParser(&input)

		// TODO: Call Notification Service via gRPC
		_ = userID

		return c.JSON(fiber.Map{
			"data": fiber.Map{
				"marked_count": 0,
				"message":      "Notifications marked as read",
			},
			"request_id": c.GetRespHeader("X-Request-ID"),
		})
	}
}

// DeleteNotification handles delete notification request.
// @Summary Delete notification
// @Tags notifications
// @Produce json
// @Security BearerAuth
// @Param id path string true "Notification ID"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /notifications/{id} [delete]
func DeleteNotification(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Locals("userID").(uuid.UUID)
		notificationID := c.Params("id")

		// TODO: Call Notification Service via gRPC
		_, _ = userID, notificationID

		return c.JSON(fiber.Map{
			"data": fiber.Map{
				"message": "Notification deleted",
			},
			"request_id": c.GetRespHeader("X-Request-ID"),
		})
	}
}

// DeleteAllRead handles delete all read notifications request.
// @Summary Delete all read notifications
// @Tags notifications
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /notifications [delete]
func DeleteAllRead(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Locals("userID").(uuid.UUID)

		// TODO: Call Notification Service via gRPC
		_ = userID

		return c.JSON(fiber.Map{
			"data": fiber.Map{
				"deleted_count": 0,
				"message":       "Notifications deleted",
			},
			"request_id": c.GetRespHeader("X-Request-ID"),
		})
	}
}

