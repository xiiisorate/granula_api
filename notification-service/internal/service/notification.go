// Package service handles business logic for Notification Service.
package service

import (
	"time"

	"github.com/xiiisorate/granula_api/notification-service/internal/repository"
	"github.com/xiiisorate/granula_api/shared/pkg/errors"

	"github.com/google/uuid"
)

// NotificationService handles notification business logic.
type NotificationService struct {
	notifRepo *repository.NotificationRepository
}

// NewNotificationService creates a new NotificationService.
func NewNotificationService(notifRepo *repository.NotificationRepository) *NotificationService {
	return &NotificationService{notifRepo: notifRepo}
}

// CreateInput contains data for creating a notification.
type CreateInput struct {
	UserID  uuid.UUID
	Type    repository.NotificationType
	Title   string
	Message string
	Data    repository.NotificationData
}

// Create creates a new notification.
func (s *NotificationService) Create(input *CreateInput) (*repository.Notification, error) {
	notif := &repository.Notification{
		UserID:  input.UserID,
		Type:    input.Type,
		Title:   input.Title,
		Message: input.Message,
		Data:    input.Data,
	}

	if err := s.notifRepo.Create(notif); err != nil {
		return nil, errors.Internal("failed to create notification").WithCause(err)
	}

	return notif, nil
}

// GetListInput contains parameters for listing notifications.
type GetListInput struct {
	UserID     uuid.UUID
	Limit      int
	Offset     int
	UnreadOnly bool
	Type       *repository.NotificationType
}

// NotificationListResult contains the list result.
type NotificationListResult struct {
	Notifications []repository.Notification
	UnreadCount   int64
	Total         int64
}

// GetList retrieves notifications for a user.
func (s *NotificationService) GetList(input *GetListInput) (*NotificationListResult, error) {
	if input.Limit <= 0 || input.Limit > 100 {
		input.Limit = 50
	}
	if input.Offset < 0 {
		input.Offset = 0
	}

	notifications, err := s.notifRepo.FindByUserID(input.UserID, input.Limit, input.Offset, input.UnreadOnly, input.Type)
	if err != nil {
		return nil, errors.Internal("failed to list notifications").WithCause(err)
	}

	unreadCount, err := s.notifRepo.CountUnreadByUserID(input.UserID)
	if err != nil {
		return nil, errors.Internal("failed to count unread notifications").WithCause(err)
	}

	total, err := s.notifRepo.CountByUserID(input.UserID)
	if err != nil {
		return nil, errors.Internal("failed to count notifications").WithCause(err)
	}

	return &NotificationListResult{
		Notifications: notifications,
		UnreadCount:   unreadCount,
		Total:         total,
	}, nil
}

// UnreadCountResult contains unread count result.
type UnreadCountResult struct {
	UnreadCount int64
	ByType      map[string]int
}

// GetUnreadCount retrieves unread notification count.
func (s *NotificationService) GetUnreadCount(userID uuid.UUID) (*UnreadCountResult, error) {
	unreadCount, err := s.notifRepo.CountUnreadByUserID(userID)
	if err != nil {
		return nil, errors.Internal("failed to count unread notifications").WithCause(err)
	}

	byType, err := s.notifRepo.CountUnreadByType(userID)
	if err != nil {
		return nil, errors.Internal("failed to count notifications by type").WithCause(err)
	}

	return &UnreadCountResult{
		UnreadCount: unreadCount,
		ByType:      byType,
	}, nil
}

// MarkAsRead marks a notification as read.
func (s *NotificationService) MarkAsRead(userID, notifID uuid.UUID) (*repository.Notification, error) {
	notif, err := s.notifRepo.FindByID(notifID)
	if err != nil {
		return nil, errors.NotFound("notification", notifID.String())
	}

	// Verify ownership
	if notif.UserID != userID {
		return nil, errors.NotFound("notification", notifID.String())
	}

	if err := s.notifRepo.MarkAsRead(notifID); err != nil {
		return nil, errors.Internal("failed to mark notification as read").WithCause(err)
	}

	notif.MarkAsRead()
	return notif, nil
}

// MarkAllAsReadInput contains parameters for marking all as read.
type MarkAllAsReadInput struct {
	UserID uuid.UUID
	Type   *repository.NotificationType
	Before *time.Time
}

// MarkAllAsRead marks all notifications as read.
func (s *NotificationService) MarkAllAsRead(input *MarkAllAsReadInput) (int64, error) {
	return s.notifRepo.MarkAllAsRead(input.UserID, input.Type, input.Before)
}

// Delete deletes a notification.
func (s *NotificationService) Delete(userID, notifID uuid.UUID) error {
	notif, err := s.notifRepo.FindByID(notifID)
	if err != nil {
		return errors.NotFound("notification", notifID.String())
	}

	// Verify ownership
	if notif.UserID != userID {
		return errors.NotFound("notification", notifID.String())
	}

	return s.notifRepo.Delete(notifID)
}

// DeleteAllRead deletes all read notifications.
func (s *NotificationService) DeleteAllRead(userID uuid.UUID) (int64, error) {
	return s.notifRepo.DeleteReadByUserID(userID)
}
