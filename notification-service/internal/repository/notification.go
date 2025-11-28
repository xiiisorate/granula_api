// Package repository handles data access for Notification Service.
package repository

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// NotificationRepository handles notification database operations.
type NotificationRepository struct {
	db *gorm.DB
}

// NewNotificationRepository creates a new NotificationRepository.
func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// Create creates a new notification.
func (r *NotificationRepository) Create(notif *Notification) error {
	return r.db.Create(notif).Error
}

// FindByID finds a notification by ID.
func (r *NotificationRepository) FindByID(id uuid.UUID) (*Notification, error) {
	var notif Notification
	err := r.db.Where("id = ?", id).First(&notif).Error
	if err != nil {
		return nil, err
	}
	return &notif, nil
}

// FindByUserID finds notifications for a user with pagination and filters.
func (r *NotificationRepository) FindByUserID(userID uuid.UUID, limit, offset int, unreadOnly bool, notifType *NotificationType) ([]Notification, error) {
	var notifications []Notification

	query := r.db.Where("user_id = ?", userID)

	if unreadOnly {
		query = query.Where("read = ?", false)
	}

	if notifType != nil {
		query = query.Where("type = ?", *notifType)
	}

	err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&notifications).Error

	return notifications, err
}

// CountByUserID counts all notifications for a user.
func (r *NotificationRepository) CountByUserID(userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&Notification{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return count, err
}

// CountUnreadByUserID counts unread notifications for a user.
func (r *NotificationRepository) CountUnreadByUserID(userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&Notification{}).
		Where("user_id = ? AND read = ?", userID, false).
		Count(&count).Error
	return count, err
}

// CountUnreadByType counts unread notifications by type for a user.
func (r *NotificationRepository) CountUnreadByType(userID uuid.UUID) (map[string]int, error) {
	type Result struct {
		Type  string
		Count int
	}

	var results []Result
	err := r.db.Model(&Notification{}).
		Select("type, count(*) as count").
		Where("user_id = ? AND read = ?", userID, false).
		Group("type").
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	counts := make(map[string]int)
	for _, r := range results {
		counts[r.Type] = r.Count
	}
	return counts, nil
}

// MarkAsRead marks a notification as read.
func (r *NotificationRepository) MarkAsRead(id uuid.UUID) error {
	now := time.Now()
	return r.db.Model(&Notification{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"read":    true,
			"read_at": now,
		}).Error
}

// MarkAllAsRead marks all notifications as read for a user.
func (r *NotificationRepository) MarkAllAsRead(userID uuid.UUID, notifType *NotificationType, before *time.Time) (int64, error) {
	now := time.Now()

	query := r.db.Model(&Notification{}).
		Where("user_id = ? AND read = ?", userID, false)

	if notifType != nil {
		query = query.Where("type = ?", *notifType)
	}

	if before != nil {
		query = query.Where("created_at < ?", *before)
	}

	result := query.Updates(map[string]interface{}{
		"read":    true,
		"read_at": now,
	})

	return result.RowsAffected, result.Error
}

// Delete removes a notification.
func (r *NotificationRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&Notification{}, "id = ?", id).Error
}

// DeleteReadByUserID removes all read notifications for a user.
func (r *NotificationRepository) DeleteReadByUserID(userID uuid.UUID) (int64, error) {
	result := r.db.Delete(&Notification{}, "user_id = ? AND read = ?", userID, true)
	return result.RowsAffected, result.Error
}

