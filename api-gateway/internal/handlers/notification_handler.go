// =============================================================================
// Package handlers contains HTTP handlers for API Gateway.
// =============================================================================
// Notification handlers manage user notification operations.
// =============================================================================
package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"google.golang.org/grpc"

	notifpb "github.com/xiiisorate/granula_api/shared/gen/notification/v1"
)

// NotificationHandler handles notification-related HTTP requests.
type NotificationHandler struct {
	notifClient notifpb.NotificationServiceClient
}

// NewNotificationHandler creates a new NotificationHandler with gRPC client connection.
func NewNotificationHandler(conn *grpc.ClientConn) *NotificationHandler {
	return &NotificationHandler{
		notifClient: notifpb.NewNotificationServiceClient(conn),
	}
}

// =============================================================================
// Handlers
// =============================================================================

// GetNotifications returns the user's notifications.
//
// @Summary Get notifications
// @Description Returns paginated list of user's notifications
// @Tags notifications
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param unread_only query bool false "Only unread notifications" default(false)
// @Success 200 {object} map[string]interface{} "List of notifications"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /notifications [get]
func (h *NotificationHandler) GetNotifications(c *fiber.Ctx) error {
	userIDStr, ok := c.Locals("user_id").(string)
	if !ok || userIDStr == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid user ID")
	}

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)
	unreadOnly := c.QueryBool("unread_only", false)

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	resp, err := h.notifClient.GetNotifications(ctx, &notifpb.GetNotificationsRequest{
		UserId:     userID.String(),
		UnreadOnly: unreadOnly,
	})

	if err != nil {
		return handleGRPCError(err)
	}

	notifications := make([]fiber.Map, 0, len(resp.Notifications))
	for _, n := range resp.Notifications {
		notifications = append(notifications, fiber.Map{
			"id":         n.Id,
			"type":       n.Type,
			"title":      n.Title,
			"message":    n.Message,
			"data":       n.Data,
			"read":       n.Read,
			"created_at": n.CreatedAt,
		})
	}

	var total int64
	var totalPages int64
	if resp.Pagination != nil {
		total = int64(resp.Pagination.Total)
		totalPages = int64(resp.Pagination.TotalPages)
	}

	return c.JSON(fiber.Map{
		"data": notifications,
		"pagination": fiber.Map{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": totalPages,
		},
		"request_id": c.GetRespHeader("X-Request-ID"),
	})
}

// GetUnreadCount returns the count of unread notifications.
//
// @Summary Get unread count
// @Description Returns the number of unread notifications
// @Tags notifications
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Unread count"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /notifications/count [get]
func (h *NotificationHandler) GetUnreadCount(c *fiber.Ctx) error {
	userIDStr, ok := c.Locals("user_id").(string)
	if !ok || userIDStr == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid user ID")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	resp, err := h.notifClient.GetUnreadCount(ctx, &notifpb.GetUnreadCountRequest{
		UserId: userID.String(),
	})

	if err != nil {
		return handleGRPCError(err)
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"unread_count": resp.Count,
		},
		"request_id": c.GetRespHeader("X-Request-ID"),
	})
}

// MarkAsRead marks a notification as read.
//
// @Summary Mark notification as read
// @Description Marks a specific notification as read
// @Tags notifications
// @Produce json
// @Security BearerAuth
// @Param id path string true "Notification ID"
// @Success 200 {object} map[string]interface{} "Notification marked as read"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Notification not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /notifications/{id}/read [post]
func (h *NotificationHandler) MarkAsRead(c *fiber.Ctx) error {
	_, ok := c.Locals("user_id").(string)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}

	notificationID := c.Params("id")
	if notificationID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "notification id is required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	_, err := h.notifClient.MarkAsRead(ctx, &notifpb.MarkAsReadRequest{
		NotificationId: notificationID,
	})

	if err != nil {
		return handleGRPCError(err)
	}

	return c.JSON(fiber.Map{
		"message":    "Notification marked as read",
		"request_id": c.GetRespHeader("X-Request-ID"),
	})
}

// MarkAllAsRead marks all notifications as read.
//
// @Summary Mark all notifications as read
// @Description Marks all user's notifications as read
// @Tags notifications
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "All notifications marked as read"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /notifications/read-all [post]
func (h *NotificationHandler) MarkAllAsRead(c *fiber.Ctx) error {
	userIDStr, ok := c.Locals("user_id").(string)
	if !ok || userIDStr == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid user ID")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	resp, err := h.notifClient.MarkAllAsRead(ctx, &notifpb.MarkAllAsReadRequest{
		UserId: userID.String(),
	})

	if err != nil {
		return handleGRPCError(err)
	}

	return c.JSON(fiber.Map{
		"message":       "All notifications marked as read",
		"updated_count": resp.MarkedCount,
		"request_id":    c.GetRespHeader("X-Request-ID"),
	})
}

// DeleteNotification deletes a specific notification.
//
// @Summary Delete notification
// @Description Deletes a specific notification
// @Tags notifications
// @Produce json
// @Security BearerAuth
// @Param id path string true "Notification ID"
// @Success 200 {object} map[string]interface{} "Notification deleted"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Notification not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /notifications/{id} [delete]
func (h *NotificationHandler) DeleteNotification(c *fiber.Ctx) error {
	_, ok := c.Locals("user_id").(string)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}

	notificationID := c.Params("id")
	if notificationID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "notification id is required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	_, err := h.notifClient.DeleteNotification(ctx, &notifpb.DeleteNotificationRequest{
		NotificationId: notificationID,
	})

	if err != nil {
		return handleGRPCError(err)
	}

	return c.JSON(fiber.Map{
		"message":    "Notification deleted",
		"request_id": c.GetRespHeader("X-Request-ID"),
	})
}

// DeleteAllRead deletes all read notifications.
//
// @Summary Delete all read notifications
// @Description Deletes all read notifications for the user
// @Tags notifications
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Read notifications deleted"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /notifications [delete]
func (h *NotificationHandler) DeleteAllRead(c *fiber.Ctx) error {
	userIDStr, ok := c.Locals("user_id").(string)
	if !ok || userIDStr == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid user ID")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	resp, err := h.notifClient.DeleteAllRead(ctx, &notifpb.DeleteAllReadRequest{
		UserId: userID.String(),
	})

	if err != nil {
		return handleGRPCError(err)
	}

	return c.JSON(fiber.Map{
		"message":       "Read notifications deleted",
		"deleted_count": resp.DeletedCount,
		"request_id":    c.GetRespHeader("X-Request-ID"),
	})
}
