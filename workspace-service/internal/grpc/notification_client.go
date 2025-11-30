// =============================================================================
// Package grpc provides gRPC handlers and clients for Workspace Service.
// =============================================================================
// NotificationClient provides integration with Notification Service for sending
// workspace-related notifications to users.
//
// Usage:
//
//	client, err := NewNotificationClient("notification-service:50060", log)
//	if err != nil {
//	    log.Warn("notification service unavailable", logger.Err(err))
//	}
//	// Use client in WorkspaceService
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
// It provides convenient methods for sending workspace-related notifications.
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
func NewNotificationClient(addr string, log *logger.Logger) (*NotificationClient, error) {
	if addr == "" {
		return nil, fmt.Errorf("notification service address is required")
	}

	// Create gRPC connection with options
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
// Workspace Notification Methods
// =============================================================================

// SendMemberAdded sends notification when a user is added to a workspace.
// Notifies the newly added member about their new access.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - userID: UUID of the user who was added
//   - workspaceID: UUID of the workspace
//   - workspaceName: Display name of the workspace
//   - role: Role assigned to the user (viewer, editor, admin)
//   - addedByName: Name of the user who added them
//
// Returns:
//   - error: nil on success, error if notification failed
func (c *NotificationClient) SendMemberAdded(ctx context.Context, userID, workspaceID, workspaceName, role, addedByName string) error {
	if c == nil || c.client == nil {
		return fmt.Errorf("notification client not initialized")
	}

	ctx, cancel := context.WithTimeout(ctx, notificationCallTimeout)
	defer cancel()

	roleNames := map[string]string{
		"viewer": "Просмотр",
		"editor": "Редактор",
		"admin":  "Администратор",
		"owner":  "Владелец",
	}

	roleName := roleNames[role]
	if roleName == "" {
		roleName = role
	}

	message := fmt.Sprintf("%s добавил вас в проект \"%s\" с ролью \"%s\"", addedByName, workspaceName, roleName)

	_, err := c.client.SendNotification(ctx, &notificationpb.SendNotificationRequest{
		UserId:   userID,
		Type:     notificationpb.NotificationType_NOTIFICATION_TYPE_WORKSPACE,
		Priority: notificationpb.NotificationPriority_NOTIFICATION_PRIORITY_NORMAL,
		Title:    "Вы добавлены в проект",
		Message:  message,
		Data: map[string]string{
			"workspace_id":   workspaceID,
			"workspace_name": workspaceName,
			"role":           role,
			"added_by":       addedByName,
		},
		ActionUrl:  fmt.Sprintf("/workspaces/%s", workspaceID),
		EntityId:   workspaceID,
		EntityType: "workspace",
	})

	if err != nil {
		c.log.Warn("failed to send member_added notification",
			logger.Err(err),
			logger.String("user_id", userID),
			logger.String("workspace_id", workspaceID),
		)
		return fmt.Errorf("send notification: %w", err)
	}

	c.log.Debug("member_added notification sent",
		logger.String("user_id", userID),
		logger.String("workspace_id", workspaceID),
	)

	return nil
}

// SendInvitation sends notification when a user is invited to a workspace.
// The invitee receives an invitation they can accept or decline.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - inviteeEmail: Email of the invited user
//   - inviteeUserID: UUID of the invitee (if they exist in system, empty otherwise)
//   - workspaceID: UUID of the workspace
//   - workspaceName: Display name of the workspace
//   - inviterName: Name of the person sending the invitation
//   - role: Role being offered
//   - inviteToken: Token for accepting the invitation
//
// Returns:
//   - error: nil on success, error if notification failed
//
// Note: If inviteeUserID is empty, only email notification is sent.
// If inviteeUserID is provided, in-app notification is also sent.
func (c *NotificationClient) SendInvitation(ctx context.Context, inviteeEmail, inviteeUserID, workspaceID, workspaceName, inviterName, role, inviteToken string) error {
	if c == nil || c.client == nil {
		return fmt.Errorf("notification client not initialized")
	}

	ctx, cancel := context.WithTimeout(ctx, notificationCallTimeout)
	defer cancel()

	roleNames := map[string]string{
		"viewer": "Просмотр",
		"editor": "Редактор",
		"admin":  "Администратор",
	}

	roleName := roleNames[role]
	if roleName == "" {
		roleName = role
	}

	message := fmt.Sprintf("%s приглашает вас в проект \"%s\" с ролью \"%s\"", inviterName, workspaceName, roleName)

	// If user exists in system, send in-app notification
	if inviteeUserID != "" {
		_, err := c.client.SendNotification(ctx, &notificationpb.SendNotificationRequest{
			UserId:   inviteeUserID,
			Type:     notificationpb.NotificationType_NOTIFICATION_TYPE_WORKSPACE,
			Priority: notificationpb.NotificationPriority_NOTIFICATION_PRIORITY_HIGH,
			Title:    "Приглашение в проект",
			Message:  message,
			Data: map[string]string{
				"workspace_id":   workspaceID,
				"workspace_name": workspaceName,
				"role":           role,
				"inviter_name":   inviterName,
				"invite_token":   inviteToken,
			},
			ActionUrl:  fmt.Sprintf("/invites/%s", inviteToken),
			EntityId:   workspaceID,
			EntityType: "workspace_invite",
		})

		if err != nil {
			c.log.Warn("failed to send in-app invitation notification",
				logger.Err(err),
				logger.String("user_id", inviteeUserID),
				logger.String("workspace_id", workspaceID),
			)
		}
	}

	// Always send email invitation
	_, err := c.client.SendEmail(ctx, &notificationpb.SendEmailRequest{
		To:       inviteeEmail,
		Subject:  fmt.Sprintf("Приглашение в проект \"%s\"", workspaceName),
		Template: "workspace_invite",
		TemplateData: map[string]string{
			"workspace_name": workspaceName,
			"inviter_name":   inviterName,
			"role":           roleName,
			"invite_url":     fmt.Sprintf("/invites/%s", inviteToken),
		},
	})

	if err != nil {
		c.log.Warn("failed to send email invitation",
			logger.Err(err),
			logger.String("email", inviteeEmail),
			logger.String("workspace_id", workspaceID),
		)
		return fmt.Errorf("send email: %w", err)
	}

	c.log.Debug("invitation notification sent",
		logger.String("email", inviteeEmail),
		logger.String("workspace_id", workspaceID),
	)

	return nil
}

// SendMemberRemoved sends notification when a user is removed from a workspace.
// Notifies the removed member about their revoked access.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - userID: UUID of the removed user
//   - workspaceID: UUID of the workspace
//   - workspaceName: Display name of the workspace
//
// Returns:
//   - error: nil on success, error if notification failed
func (c *NotificationClient) SendMemberRemoved(ctx context.Context, userID, workspaceID, workspaceName string) error {
	if c == nil || c.client == nil {
		return fmt.Errorf("notification client not initialized")
	}

	ctx, cancel := context.WithTimeout(ctx, notificationCallTimeout)
	defer cancel()

	_, err := c.client.SendNotification(ctx, &notificationpb.SendNotificationRequest{
		UserId:   userID,
		Type:     notificationpb.NotificationType_NOTIFICATION_TYPE_WORKSPACE,
		Priority: notificationpb.NotificationPriority_NOTIFICATION_PRIORITY_NORMAL,
		Title:    "Доступ к проекту отозван",
		Message:  fmt.Sprintf("Ваш доступ к проекту \"%s\" был отозван.", workspaceName),
		Data: map[string]string{
			"workspace_id":   workspaceID,
			"workspace_name": workspaceName,
		},
		EntityId:   workspaceID,
		EntityType: "workspace",
	})

	if err != nil {
		c.log.Warn("failed to send member_removed notification",
			logger.Err(err),
			logger.String("user_id", userID),
			logger.String("workspace_id", workspaceID),
		)
		return fmt.Errorf("send notification: %w", err)
	}

	c.log.Debug("member_removed notification sent",
		logger.String("user_id", userID),
		logger.String("workspace_id", workspaceID),
	)

	return nil
}

// SendRoleChanged sends notification when a member's role is changed.
// Notifies the member about their updated permissions.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - userID: UUID of the member whose role changed
//   - workspaceID: UUID of the workspace
//   - workspaceName: Display name of the workspace
//   - oldRole: Previous role
//   - newRole: New role
//
// Returns:
//   - error: nil on success, error if notification failed
func (c *NotificationClient) SendRoleChanged(ctx context.Context, userID, workspaceID, workspaceName, oldRole, newRole string) error {
	if c == nil || c.client == nil {
		return fmt.Errorf("notification client not initialized")
	}

	ctx, cancel := context.WithTimeout(ctx, notificationCallTimeout)
	defer cancel()

	roleNames := map[string]string{
		"viewer": "Просмотр",
		"editor": "Редактор",
		"admin":  "Администратор",
	}

	newRoleName := roleNames[newRole]
	if newRoleName == "" {
		newRoleName = newRole
	}

	_, err := c.client.SendNotification(ctx, &notificationpb.SendNotificationRequest{
		UserId:   userID,
		Type:     notificationpb.NotificationType_NOTIFICATION_TYPE_WORKSPACE,
		Priority: notificationpb.NotificationPriority_NOTIFICATION_PRIORITY_NORMAL,
		Title:    "Изменение роли в проекте",
		Message:  fmt.Sprintf("Ваша роль в проекте \"%s\" изменена на \"%s\".", workspaceName, newRoleName),
		Data: map[string]string{
			"workspace_id":   workspaceID,
			"workspace_name": workspaceName,
			"old_role":       oldRole,
			"new_role":       newRole,
		},
		ActionUrl:  fmt.Sprintf("/workspaces/%s", workspaceID),
		EntityId:   workspaceID,
		EntityType: "workspace",
	})

	if err != nil {
		c.log.Warn("failed to send role_changed notification",
			logger.Err(err),
			logger.String("user_id", userID),
			logger.String("workspace_id", workspaceID),
		)
		return fmt.Errorf("send notification: %w", err)
	}

	c.log.Debug("role_changed notification sent",
		logger.String("user_id", userID),
		logger.String("workspace_id", workspaceID),
		logger.String("new_role", newRole),
	)

	return nil
}

// SendWorkspaceCreated sends notification when a workspace is created.
// This is typically just a confirmation to the owner.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - ownerID: UUID of the workspace owner
//   - workspaceID: UUID of the new workspace
//   - workspaceName: Display name of the workspace
//
// Returns:
//   - error: nil on success, error if notification failed
func (c *NotificationClient) SendWorkspaceCreated(ctx context.Context, ownerID, workspaceID, workspaceName string) error {
	if c == nil || c.client == nil {
		return fmt.Errorf("notification client not initialized")
	}

	ctx, cancel := context.WithTimeout(ctx, notificationCallTimeout)
	defer cancel()

	_, err := c.client.SendNotification(ctx, &notificationpb.SendNotificationRequest{
		UserId:   ownerID,
		Type:     notificationpb.NotificationType_NOTIFICATION_TYPE_WORKSPACE,
		Priority: notificationpb.NotificationPriority_NOTIFICATION_PRIORITY_LOW,
		Title:    "Проект создан",
		Message:  fmt.Sprintf("Проект \"%s\" успешно создан. Загрузите планировку, чтобы начать работу.", workspaceName),
		Data: map[string]string{
			"workspace_id":   workspaceID,
			"workspace_name": workspaceName,
		},
		ActionUrl:  fmt.Sprintf("/workspaces/%s", workspaceID),
		EntityId:   workspaceID,
		EntityType: "workspace",
	})

	if err != nil {
		c.log.Warn("failed to send workspace_created notification",
			logger.Err(err),
			logger.String("owner_id", ownerID),
			logger.String("workspace_id", workspaceID),
		)
		return fmt.Errorf("send notification: %w", err)
	}

	c.log.Debug("workspace_created notification sent",
		logger.String("owner_id", ownerID),
		logger.String("workspace_id", workspaceID),
	)

	return nil
}

// SendWorkspaceDeleted sends notification when a workspace is deleted.
// Notifies all members that the workspace no longer exists.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - memberUserIDs: UUIDs of all workspace members to notify
//   - workspaceID: UUID of the deleted workspace
//   - workspaceName: Display name of the deleted workspace
//
// Returns:
//   - error: nil on success, error if any notification failed
func (c *NotificationClient) SendWorkspaceDeleted(ctx context.Context, memberUserIDs []string, workspaceID, workspaceName string) error {
	if c == nil || c.client == nil {
		return fmt.Errorf("notification client not initialized")
	}

	ctx, cancel := context.WithTimeout(ctx, notificationCallTimeout*2) // Extended timeout for batch
	defer cancel()

	var lastErr error
	for _, userID := range memberUserIDs {
		_, err := c.client.SendNotification(ctx, &notificationpb.SendNotificationRequest{
			UserId:   userID,
			Type:     notificationpb.NotificationType_NOTIFICATION_TYPE_WORKSPACE,
			Priority: notificationpb.NotificationPriority_NOTIFICATION_PRIORITY_NORMAL,
			Title:    "Проект удалён",
			Message:  fmt.Sprintf("Проект \"%s\" был удалён владельцем.", workspaceName),
			Data: map[string]string{
				"workspace_id":   workspaceID,
				"workspace_name": workspaceName,
			},
			EntityId:   workspaceID,
			EntityType: "workspace",
		})

		if err != nil {
			c.log.Warn("failed to send workspace_deleted notification to member",
				logger.Err(err),
				logger.String("user_id", userID),
				logger.String("workspace_id", workspaceID),
			)
			lastErr = err
		}
	}

	if lastErr != nil {
		return fmt.Errorf("some notifications failed: %w", lastErr)
	}

	c.log.Debug("workspace_deleted notifications sent",
		logger.String("workspace_id", workspaceID),
		logger.Int("members_notified", len(memberUserIDs)),
	)

	return nil
}

