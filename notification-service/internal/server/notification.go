// Package server implements gRPC server for Notification Service.
package server

import (
	"context"
	"time"

	"github.com/xiiisorate/granula_api/notification-service/internal/repository"
	"github.com/xiiisorate/granula_api/notification-service/internal/service"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

// NotificationServiceServer is the interface for Notification gRPC service.
type NotificationServiceServer interface {
	GetNotifications(ctx context.Context, req *GetNotificationsRequest) (*GetNotificationsResponse, error)
	GetUnreadCount(ctx context.Context, req *GetUnreadCountRequest) (*GetUnreadCountResponse, error)
	MarkAsRead(ctx context.Context, req *MarkAsReadRequest) (*MarkAsReadResponse, error)
	MarkAllAsRead(ctx context.Context, req *MarkAllAsReadRequest) (*MarkAllAsReadResponse, error)
	DeleteNotification(ctx context.Context, req *DeleteNotificationRequest) (*DeleteNotificationResponse, error)
	DeleteAllRead(ctx context.Context, req *DeleteAllReadRequest) (*DeleteAllReadResponse, error)
	SendNotification(ctx context.Context, req *SendNotificationRequest) (*SendNotificationResponse, error)
}

// Request/Response types
type GetNotificationsRequest struct {
	UserID     string
	Limit      int32
	Offset     int32
	UnreadOnly bool
	Type       string
}

type GetNotificationsResponse struct {
	Notifications []*Notification
	UnreadCount   int64
	Total         int64
}

type GetUnreadCountRequest struct {
	UserID string
}

type GetUnreadCountResponse struct {
	UnreadCount int64
	ByType      map[string]int32
}

type MarkAsReadRequest struct {
	UserID         string
	NotificationID string
}

type MarkAsReadResponse struct {
	Notification *Notification
}

type MarkAllAsReadRequest struct {
	UserID string
	Type   string
	Before string
}

type MarkAllAsReadResponse struct {
	MarkedCount int64
	Message     string
}

type DeleteNotificationRequest struct {
	UserID         string
	NotificationID string
}

type DeleteNotificationResponse struct {
	Message string
}

type DeleteAllReadRequest struct {
	UserID string
}

type DeleteAllReadResponse struct {
	DeletedCount int64
	Message      string
}

type SendNotificationRequest struct {
	UserID  string
	Type    string
	Title   string
	Message string
	Data    map[string]string
}

type SendNotificationResponse struct {
	Notification *Notification
}

type Notification struct {
	ID        string
	Type      string
	Title     string
	Message   string
	Data      map[string]string
	Read      bool
	ReadAt    string
	CreatedAt string
}

// NotificationServer implements NotificationServiceServer.
type NotificationServer struct {
	notifService *service.NotificationService
}

// NewNotificationServer creates a new NotificationServer.
func NewNotificationServer(notifService *service.NotificationService) *NotificationServer {
	return &NotificationServer{notifService: notifService}
}

// GetNotifications returns a list of notifications.
func (s *NotificationServer) GetNotifications(ctx context.Context, req *GetNotificationsRequest) (*GetNotificationsResponse, error) {
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, err
	}

	var notifType *repository.NotificationType
	if req.Type != "" {
		t := repository.NotificationType(req.Type)
		notifType = &t
	}

	result, err := s.notifService.GetList(&service.GetListInput{
		UserID:     userID,
		Limit:      int(req.Limit),
		Offset:     int(req.Offset),
		UnreadOnly: req.UnreadOnly,
		Type:       notifType,
	})
	if err != nil {
		return nil, err
	}

	notifications := make([]*Notification, len(result.Notifications))
	for i, n := range result.Notifications {
		notifications[i] = notifToProto(&n)
	}

	return &GetNotificationsResponse{
		Notifications: notifications,
		UnreadCount:   result.UnreadCount,
		Total:         result.Total,
	}, nil
}

// GetUnreadCount returns unread notification count.
func (s *NotificationServer) GetUnreadCount(ctx context.Context, req *GetUnreadCountRequest) (*GetUnreadCountResponse, error) {
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, err
	}

	result, err := s.notifService.GetUnreadCount(userID)
	if err != nil {
		return nil, err
	}

	byType := make(map[string]int32)
	for k, v := range result.ByType {
		byType[k] = int32(v)
	}

	return &GetUnreadCountResponse{
		UnreadCount: result.UnreadCount,
		ByType:      byType,
	}, nil
}

// MarkAsRead marks a notification as read.
func (s *NotificationServer) MarkAsRead(ctx context.Context, req *MarkAsReadRequest) (*MarkAsReadResponse, error) {
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, err
	}

	notifID, err := uuid.Parse(req.NotificationID)
	if err != nil {
		return nil, err
	}

	notif, err := s.notifService.MarkAsRead(userID, notifID)
	if err != nil {
		return nil, err
	}

	return &MarkAsReadResponse{
		Notification: notifToProto(notif),
	}, nil
}

// MarkAllAsRead marks all notifications as read.
func (s *NotificationServer) MarkAllAsRead(ctx context.Context, req *MarkAllAsReadRequest) (*MarkAllAsReadResponse, error) {
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, err
	}

	var notifType *repository.NotificationType
	if req.Type != "" {
		t := repository.NotificationType(req.Type)
		notifType = &t
	}

	var before *time.Time
	if req.Before != "" {
		t, err := time.Parse(time.RFC3339, req.Before)
		if err == nil {
			before = &t
		}
	}

	count, err := s.notifService.MarkAllAsRead(&service.MarkAllAsReadInput{
		UserID: userID,
		Type:   notifType,
		Before: before,
	})
	if err != nil {
		return nil, err
	}

	return &MarkAllAsReadResponse{
		MarkedCount: count,
		Message:     "Notifications marked as read",
	}, nil
}

// DeleteNotification deletes a notification.
func (s *NotificationServer) DeleteNotification(ctx context.Context, req *DeleteNotificationRequest) (*DeleteNotificationResponse, error) {
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, err
	}

	notifID, err := uuid.Parse(req.NotificationID)
	if err != nil {
		return nil, err
	}

	if err := s.notifService.Delete(userID, notifID); err != nil {
		return nil, err
	}

	return &DeleteNotificationResponse{
		Message: "Notification deleted",
	}, nil
}

// DeleteAllRead deletes all read notifications.
func (s *NotificationServer) DeleteAllRead(ctx context.Context, req *DeleteAllReadRequest) (*DeleteAllReadResponse, error) {
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, err
	}

	count, err := s.notifService.DeleteAllRead(userID)
	if err != nil {
		return nil, err
	}

	return &DeleteAllReadResponse{
		DeletedCount: count,
		Message:      "Notifications deleted",
	}, nil
}

// SendNotification creates a new notification.
func (s *NotificationServer) SendNotification(ctx context.Context, req *SendNotificationRequest) (*SendNotificationResponse, error) {
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, err
	}

	data := make(repository.NotificationData)
	for k, v := range req.Data {
		data[k] = v
	}

	notif, err := s.notifService.Create(&service.CreateInput{
		UserID:  userID,
		Type:    repository.NotificationType(req.Type),
		Title:   req.Title,
		Message: req.Message,
		Data:    data,
	})
	if err != nil {
		return nil, err
	}

	return &SendNotificationResponse{
		Notification: notifToProto(notif),
	}, nil
}

// RegisterNotificationServiceServer registers the notification service server.
func RegisterNotificationServiceServer(s *grpc.Server, srv NotificationServiceServer) {
	// Will be generated from proto
}

// Helper function
func notifToProto(n *repository.Notification) *Notification {
	notif := &Notification{
		ID:        n.ID.String(),
		Type:      string(n.Type),
		Title:     n.Title,
		Message:   n.Message,
		Data:      make(map[string]string),
		Read:      n.Read,
		CreatedAt: n.CreatedAt.Format(time.RFC3339),
	}

	for k, v := range n.Data {
		notif.Data[k] = v
	}

	if n.ReadAt != nil {
		notif.ReadAt = n.ReadAt.Format(time.RFC3339)
	}

	return notif
}

