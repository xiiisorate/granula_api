// =============================================================================
// Package grpc provides gRPC handlers and clients for Request Service.
// =============================================================================
// NotificationClient provides integration with Notification Service for sending
// request-related notifications to users and staff.
//
// Usage:
//
//	client, err := NewNotificationClient("notification-service:50060", log)
//	if err != nil {
//	    log.Warn("notification service unavailable", logger.Err(err))
//	}
//	// Use client in RequestService
//
// =============================================================================
package grpc

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	notificationpb "github.com/xiiisorate/granula_api/shared/gen/notification/v1"
	"github.com/xiiisorate/granula_api/shared/pkg/logger"
)

// =============================================================================
// Constants
// =============================================================================

const (
	// notificationConnTimeout is the timeout for connecting to notification service.
	notificationConnTimeout = 5 * time.Second

	// notificationCallTimeout is the default timeout for notification RPC calls.
	notificationCallTimeout = 10 * time.Second
)

// =============================================================================
// NotificationClient
// =============================================================================

// NotificationClient wraps Notification Service gRPC client.
// It provides convenient methods for sending request-related notifications.
//
// Thread Safety: Safe for concurrent use.
type NotificationClient struct {
	client notificationpb.NotificationServiceClient
	conn   *grpc.ClientConn
	log    *logger.Logger
}

// NewNotificationClient creates a new Notification Service gRPC client.
//
// Parameters:
//   - addr: gRPC address in format "host:port" (e.g., "notification-service:50060")
//   - log: Logger instance for operational logging
//
// Returns:
//   - *NotificationClient: Connected client ready for use
//   - error: Connection error if service is unavailable
//
// Note: Connection is established lazily by gRPC, but Dial verifies address format.
func NewNotificationClient(addr string, log *logger.Logger) (*NotificationClient, error) {
	if addr == "" {
		return nil, fmt.Errorf("notification service address is required")
	}

	// Create gRPC connection with options
	// Note: Using WithInsecure for internal microservice communication
	// In production with mTLS, replace with appropriate credentials
	ctx, cancel := context.WithTimeout(context.Background(), notificationConnTimeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to notification service at %s: %w", addr, err)
	}

	log.Info("connected to Notification Service",
		logger.String("address", addr),
	)

	return &NotificationClient{
		client: notificationpb.NewNotificationServiceClient(conn),
		conn:   conn,
		log:    log,
	}, nil
}

// Close closes the gRPC connection.
// Should be called when the service shuts down.
func (c *NotificationClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// =============================================================================
// Request Notification Methods
// =============================================================================

// SendRequestSubmitted sends notification when a request is submitted for review.
// Notifies the user that their request has been received.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - userID: UUID of the user who submitted the request
//   - requestID: UUID of the submitted request
//   - serviceType: Type of service requested (e.g., "consultation", "documentation")
//
// Returns:
//   - error: nil on success, error if notification failed
func (c *NotificationClient) SendRequestSubmitted(ctx context.Context, userID, requestID, serviceType string) error {
	if c == nil || c.client == nil {
		return fmt.Errorf("notification client not initialized")
	}

	ctx, cancel := context.WithTimeout(ctx, notificationCallTimeout)
	defer cancel()

	_, err := c.client.SendNotification(ctx, &notificationpb.SendNotificationRequest{
		UserId:   userID,
		Type:     notificationpb.NotificationType_NOTIFICATION_TYPE_REQUEST,
		Priority: notificationpb.NotificationPriority_NOTIFICATION_PRIORITY_NORMAL,
		Title:    "Заявка отправлена",
		Message:  "Ваша заявка успешно отправлена на рассмотрение. Мы свяжемся с вами в ближайшее время.",
		Data: map[string]string{
			"request_id":   requestID,
			"service_type": serviceType,
		},
		ActionUrl:  fmt.Sprintf("/requests/%s", requestID),
		EntityId:   requestID,
		EntityType: "request",
	})

	if err != nil {
		c.log.Warn("failed to send request_submitted notification",
			logger.Err(err),
			logger.String("user_id", userID),
			logger.String("request_id", requestID),
		)
		return fmt.Errorf("send notification: %w", err)
	}

	c.log.Debug("request_submitted notification sent",
		logger.String("user_id", userID),
		logger.String("request_id", requestID),
	)

	return nil
}

// SendRequestStatusChanged sends notification when request status changes.
// Notifies the user about the new status of their request.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - userID: UUID of the request owner
//   - requestID: UUID of the request
//   - oldStatus: Previous status (for context)
//   - newStatus: New status after the change
//   - comment: Optional comment from staff explaining the change
//
// Returns:
//   - error: nil on success, error if notification failed
func (c *NotificationClient) SendRequestStatusChanged(ctx context.Context, userID, requestID, oldStatus, newStatus, comment string) error {
	if c == nil || c.client == nil {
		return fmt.Errorf("notification client not initialized")
	}

	ctx, cancel := context.WithTimeout(ctx, notificationCallTimeout)
	defer cancel()

	// Determine title and priority based on new status
	title, priority, message := c.getStatusNotificationContent(newStatus, comment)

	_, err := c.client.SendNotification(ctx, &notificationpb.SendNotificationRequest{
		UserId:   userID,
		Type:     notificationpb.NotificationType_NOTIFICATION_TYPE_REQUEST,
		Priority: priority,
		Title:    title,
		Message:  message,
		Data: map[string]string{
			"request_id": requestID,
			"old_status": oldStatus,
			"new_status": newStatus,
		},
		ActionUrl:  fmt.Sprintf("/requests/%s", requestID),
		EntityId:   requestID,
		EntityType: "request",
	})

	if err != nil {
		c.log.Warn("failed to send request_status_changed notification",
			logger.Err(err),
			logger.String("user_id", userID),
			logger.String("request_id", requestID),
			logger.String("new_status", newStatus),
		)
		return fmt.Errorf("send notification: %w", err)
	}

	c.log.Debug("request_status_changed notification sent",
		logger.String("user_id", userID),
		logger.String("request_id", requestID),
		logger.String("new_status", newStatus),
	)

	return nil
}

// getStatusNotificationContent returns notification content based on status.
func (c *NotificationClient) getStatusNotificationContent(status, comment string) (string, notificationpb.NotificationPriority, string) {
	statusContent := map[string]struct {
		title    string
		priority notificationpb.NotificationPriority
		message  string
	}{
		"pending": {
			title:    "Заявка на рассмотрении",
			priority: notificationpb.NotificationPriority_NOTIFICATION_PRIORITY_LOW,
			message:  "Ваша заявка ожидает рассмотрения.",
		},
		"reviewing": {
			title:    "Заявка рассматривается",
			priority: notificationpb.NotificationPriority_NOTIFICATION_PRIORITY_NORMAL,
			message:  "Эксперт взял вашу заявку в работу.",
		},
		"approved": {
			title:    "Заявка одобрена",
			priority: notificationpb.NotificationPriority_NOTIFICATION_PRIORITY_HIGH,
			message:  "Ваша заявка одобрена! Ожидайте связи от эксперта.",
		},
		"rejected": {
			title:    "Заявка отклонена",
			priority: notificationpb.NotificationPriority_NOTIFICATION_PRIORITY_HIGH,
			message:  "К сожалению, ваша заявка была отклонена.",
		},
		"in_progress": {
			title:    "Заявка в работе",
			priority: notificationpb.NotificationPriority_NOTIFICATION_PRIORITY_NORMAL,
			message:  "Работа по вашей заявке началась.",
		},
		"completed": {
			title:    "Заявка выполнена",
			priority: notificationpb.NotificationPriority_NOTIFICATION_PRIORITY_HIGH,
			message:  "Работа по вашей заявке завершена!",
		},
		"cancelled": {
			title:    "Заявка отменена",
			priority: notificationpb.NotificationPriority_NOTIFICATION_PRIORITY_NORMAL,
			message:  "Ваша заявка была отменена.",
		},
	}

	content, ok := statusContent[status]
	if !ok {
		content = statusContent["pending"]
		content.title = "Статус заявки изменён"
		content.message = "Статус вашей заявки был изменён."
	}

	// Append comment if provided
	if comment != "" {
		content.message = content.message + " " + comment
	}

	return content.title, content.priority, content.message
}

// SendRequestAssigned sends notification when an expert is assigned.
// Notifies both the user and the expert about the assignment.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - userID: UUID of the request owner
//   - requestID: UUID of the request
//   - expertID: UUID of the assigned expert
//   - expertName: Display name of the expert (for user notification)
//
// Returns:
//   - error: nil on success, error if notification failed
func (c *NotificationClient) SendRequestAssigned(ctx context.Context, userID, requestID, expertID, expertName string) error {
	if c == nil || c.client == nil {
		return fmt.Errorf("notification client not initialized")
	}

	ctx, cancel := context.WithTimeout(ctx, notificationCallTimeout)
	defer cancel()

	// Notify user about expert assignment
	_, err := c.client.SendNotification(ctx, &notificationpb.SendNotificationRequest{
		UserId:   userID,
		Type:     notificationpb.NotificationType_NOTIFICATION_TYPE_REQUEST,
		Priority: notificationpb.NotificationPriority_NOTIFICATION_PRIORITY_NORMAL,
		Title:    "Назначен эксперт",
		Message:  fmt.Sprintf("К вашей заявке назначен эксперт: %s", expertName),
		Data: map[string]string{
			"request_id":  requestID,
			"expert_id":   expertID,
			"expert_name": expertName,
		},
		ActionUrl:  fmt.Sprintf("/requests/%s", requestID),
		EntityId:   requestID,
		EntityType: "request",
	})

	if err != nil {
		c.log.Warn("failed to send request_assigned notification to user",
			logger.Err(err),
			logger.String("user_id", userID),
			logger.String("request_id", requestID),
		)
	}

	// Notify expert about new assignment
	_, err = c.client.SendNotification(ctx, &notificationpb.SendNotificationRequest{
		UserId:   expertID,
		Type:     notificationpb.NotificationType_NOTIFICATION_TYPE_REQUEST,
		Priority: notificationpb.NotificationPriority_NOTIFICATION_PRIORITY_HIGH,
		Title:    "Новая заявка назначена",
		Message:  "Вам назначена новая заявка на рассмотрение.",
		Data: map[string]string{
			"request_id": requestID,
			"user_id":    userID,
		},
		ActionUrl:  fmt.Sprintf("/expert/requests/%s", requestID),
		EntityId:   requestID,
		EntityType: "request",
	})

	if err != nil {
		c.log.Warn("failed to send request_assigned notification to expert",
			logger.Err(err),
			logger.String("expert_id", expertID),
			logger.String("request_id", requestID),
		)
		return fmt.Errorf("send notification to expert: %w", err)
	}

	c.log.Debug("request_assigned notifications sent",
		logger.String("user_id", userID),
		logger.String("expert_id", expertID),
		logger.String("request_id", requestID),
	)

	return nil
}

// NotifyStaff sends notification to staff about new request.
// This is a broadcast notification to alert staff about incoming work.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - requestID: UUID of the new request
//   - serviceType: Type of service requested
//   - workspaceID: UUID of the associated workspace
//
// Returns:
//   - error: nil on success, error if notification failed
//
// Note: In production, this should fetch staff user IDs from a user service.
// For MVP, it uses a broadcast mechanism or predefined staff channel.
func (c *NotificationClient) NotifyStaff(ctx context.Context, requestID, serviceType, workspaceID string) error {
	if c == nil || c.client == nil {
		return fmt.Errorf("notification client not initialized")
	}

	ctx, cancel := context.WithTimeout(ctx, notificationCallTimeout)
	defer cancel()

	// Map service type to human-readable name
	serviceNames := map[string]string{
		"consultation":  "консультацию",
		"documentation": "оформление документов",
		"expert_visit":  "выезд эксперта",
		"full_service":  "полный комплекс услуг",
	}

	serviceName := serviceNames[serviceType]
	if serviceName == "" {
		serviceName = serviceType
	}

	// Send notification to staff broadcast channel
	// Note: Using "staff_broadcast" as a special user_id that notification service
	// can expand to all staff users. In production, implement proper broadcast.
	_, err := c.client.SendNotification(ctx, &notificationpb.SendNotificationRequest{
		UserId:   "staff_broadcast", // Special ID for staff notifications
		Type:     notificationpb.NotificationType_NOTIFICATION_TYPE_REQUEST,
		Priority: notificationpb.NotificationPriority_NOTIFICATION_PRIORITY_NORMAL,
		Title:    "Новая заявка",
		Message:  fmt.Sprintf("Поступила новая заявка на %s", serviceName),
		Data: map[string]string{
			"request_id":   requestID,
			"service_type": serviceType,
			"workspace_id": workspaceID,
		},
		ActionUrl:  fmt.Sprintf("/admin/requests/%s", requestID),
		EntityId:   requestID,
		EntityType: "request",
	})

	if err != nil {
		c.log.Warn("failed to notify staff about new request",
			logger.Err(err),
			logger.String("request_id", requestID),
		)
		return fmt.Errorf("send staff notification: %w", err)
	}

	c.log.Debug("staff notified about new request",
		logger.String("request_id", requestID),
		logger.String("service_type", serviceType),
	)

	return nil
}

// SendRequestRejected sends notification when request is rejected.
// Includes the rejection reason for the user.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - userID: UUID of the request owner
//   - requestID: UUID of the rejected request
//   - reason: Explanation for rejection
//
// Returns:
//   - error: nil on success, error if notification failed
func (c *NotificationClient) SendRequestRejected(ctx context.Context, userID, requestID, reason string) error {
	if c == nil || c.client == nil {
		return fmt.Errorf("notification client not initialized")
	}

	ctx, cancel := context.WithTimeout(ctx, notificationCallTimeout)
	defer cancel()

	message := "К сожалению, ваша заявка была отклонена."
	if reason != "" {
		message = fmt.Sprintf("К сожалению, ваша заявка была отклонена. Причина: %s", reason)
	}

	_, err := c.client.SendNotification(ctx, &notificationpb.SendNotificationRequest{
		UserId:   userID,
		Type:     notificationpb.NotificationType_NOTIFICATION_TYPE_REQUEST,
		Priority: notificationpb.NotificationPriority_NOTIFICATION_PRIORITY_HIGH,
		Title:    "Заявка отклонена",
		Message:  message,
		Data: map[string]string{
			"request_id": requestID,
			"reason":     reason,
		},
		ActionUrl:  fmt.Sprintf("/requests/%s", requestID),
		EntityId:   requestID,
		EntityType: "request",
	})

	if err != nil {
		c.log.Warn("failed to send request_rejected notification",
			logger.Err(err),
			logger.String("user_id", userID),
			logger.String("request_id", requestID),
		)
		return fmt.Errorf("send notification: %w", err)
	}

	c.log.Debug("request_rejected notification sent",
		logger.String("user_id", userID),
		logger.String("request_id", requestID),
	)

	return nil
}

// SendRequestCompleted sends notification when request is completed.
// Includes final cost if available.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - userID: UUID of the request owner
//   - requestID: UUID of the completed request
//   - finalCost: Final cost of the service (0 if not applicable)
//
// Returns:
//   - error: nil on success, error if notification failed
func (c *NotificationClient) SendRequestCompleted(ctx context.Context, userID, requestID string, finalCost int) error {
	if c == nil || c.client == nil {
		return fmt.Errorf("notification client not initialized")
	}

	ctx, cancel := context.WithTimeout(ctx, notificationCallTimeout)
	defer cancel()

	message := "Работа по вашей заявке успешно завершена!"
	if finalCost > 0 {
		message = fmt.Sprintf("Работа по вашей заявке успешно завершена! Итоговая стоимость: %d ₽", finalCost)
	}

	_, err := c.client.SendNotification(ctx, &notificationpb.SendNotificationRequest{
		UserId:   userID,
		Type:     notificationpb.NotificationType_NOTIFICATION_TYPE_REQUEST,
		Priority: notificationpb.NotificationPriority_NOTIFICATION_PRIORITY_HIGH,
		Title:    "Заявка выполнена",
		Message:  message,
		Data: map[string]string{
			"request_id": requestID,
			"final_cost": fmt.Sprintf("%d", finalCost),
		},
		ActionUrl:  fmt.Sprintf("/requests/%s", requestID),
		EntityId:   requestID,
		EntityType: "request",
	})

	if err != nil {
		c.log.Warn("failed to send request_completed notification",
			logger.Err(err),
			logger.String("user_id", userID),
			logger.String("request_id", requestID),
		)
		return fmt.Errorf("send notification: %w", err)
	}

	c.log.Debug("request_completed notification sent",
		logger.String("user_id", userID),
		logger.String("request_id", requestID),
	)

	return nil
}
