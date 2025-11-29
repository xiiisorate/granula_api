// =============================================================================
// Package server implements gRPC server for Notification Service.
// =============================================================================
package server

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/xiiisorate/granula_api/notification-service/internal/repository"
	"github.com/xiiisorate/granula_api/notification-service/internal/service"
	commonpb "github.com/xiiisorate/granula_api/shared/gen/common/v1"
	notifpb "github.com/xiiisorate/granula_api/shared/gen/notification/v1"
)

// NotificationServer implements notifpb.NotificationServiceServer.
type NotificationServer struct {
	notifpb.UnimplementedNotificationServiceServer
	notifService *service.NotificationService
}

// NewNotificationServer creates a new NotificationServer.
func NewNotificationServer(notifService *service.NotificationService) *NotificationServer {
	return &NotificationServer{notifService: notifService}
}

// SendNotification creates a new notification.
func (s *NotificationServer) SendNotification(ctx context.Context, req *notifpb.SendNotificationRequest) (*notifpb.SendNotificationResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
	}

	data := make(repository.NotificationData)
	for k, v := range req.Data {
		data[k] = v
	}

	notif, err := s.notifService.Create(&service.CreateInput{
		UserID:  userID,
		Type:    mapNotificationType(req.Type),
		Title:   req.Title,
		Message: req.Message,
		Data:    data,
	})
	if err != nil {
		return nil, convertError(err)
	}

	return &notifpb.SendNotificationResponse{
		NotificationId: notif.ID.String(),
		EmailSent:      false,
		PushSent:       false,
	}, nil
}

// GetNotifications returns a list of notifications.
func (s *NotificationServer) GetNotifications(ctx context.Context, req *notifpb.GetNotificationsRequest) (*notifpb.GetNotificationsResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
	}

	limit := 20
	offset := 0
	if req.Pagination != nil {
		if req.Pagination.PageSize > 0 {
			limit = int(req.Pagination.PageSize)
		}
		if req.Pagination.Page > 0 {
			offset = (int(req.Pagination.Page) - 1) * limit
		}
	}

	var notifType *repository.NotificationType
	if req.Type != notifpb.NotificationType_NOTIFICATION_TYPE_UNSPECIFIED {
		t := mapNotificationType(req.Type)
		notifType = &t
	}

	result, err := s.notifService.GetList(&service.GetListInput{
		UserID:     userID,
		Limit:      limit,
		Offset:     offset,
		UnreadOnly: req.UnreadOnly,
		Type:       notifType,
	})
	if err != nil {
		return nil, convertError(err)
	}

	notifications := make([]*notifpb.Notification, len(result.Notifications))
	for i, n := range result.Notifications {
		notifications[i] = notifToProto(&n)
	}

	return &notifpb.GetNotificationsResponse{
		Notifications: notifications,
		Pagination: &commonpb.PaginationResponse{
			Total:      int32(result.Total),
			TotalPages: int32((result.Total + int64(limit) - 1) / int64(limit)),
		},
	}, nil
}

// MarkAsRead marks a notification as read.
func (s *NotificationServer) MarkAsRead(ctx context.Context, req *notifpb.MarkAsReadRequest) (*notifpb.MarkAsReadResponse, error) {
	if req.NotificationId == "" {
		return nil, status.Error(codes.InvalidArgument, "notification_id is required")
	}

	notifID, err := uuid.Parse(req.NotificationId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid notification_id format")
	}

	// Use a placeholder userID for now - in production this should come from context
	// The service will validate ownership
	placeholderUserID := uuid.Nil

	_, err = s.notifService.MarkAsRead(placeholderUserID, notifID)
	if err != nil {
		return nil, convertError(err)
	}

	return &notifpb.MarkAsReadResponse{
		Success: true,
	}, nil
}

// MarkAllAsRead marks all notifications as read.
func (s *NotificationServer) MarkAllAsRead(ctx context.Context, req *notifpb.MarkAllAsReadRequest) (*notifpb.MarkAllAsReadResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
	}

	var notifType *repository.NotificationType
	if req.Type != notifpb.NotificationType_NOTIFICATION_TYPE_UNSPECIFIED {
		t := mapNotificationType(req.Type)
		notifType = &t
	}

	count, err := s.notifService.MarkAllAsRead(&service.MarkAllAsReadInput{
		UserID: userID,
		Type:   notifType,
	})
	if err != nil {
		return nil, convertError(err)
	}

	return &notifpb.MarkAllAsReadResponse{
		MarkedCount: int32(count),
	}, nil
}

// GetUnreadCount returns unread notification count.
func (s *NotificationServer) GetUnreadCount(ctx context.Context, req *notifpb.GetUnreadCountRequest) (*notifpb.GetUnreadCountResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
	}

	result, err := s.notifService.GetUnreadCount(userID)
	if err != nil {
		return nil, convertError(err)
	}

	byType := make(map[string]int32)
	for k, v := range result.ByType {
		byType[k] = int32(v)
	}

	return &notifpb.GetUnreadCountResponse{
		Count:  int32(result.UnreadCount),
		ByType: byType,
	}, nil
}

// SendEmail sends an email.
func (s *NotificationServer) SendEmail(ctx context.Context, req *notifpb.SendEmailRequest) (*notifpb.SendEmailResponse, error) {
	// TODO: Implement email sending
	return &notifpb.SendEmailResponse{
		Success:   true,
		MessageId: uuid.New().String(),
	}, nil
}

// DeleteNotification deletes a notification.
func (s *NotificationServer) DeleteNotification(ctx context.Context, req *notifpb.DeleteNotificationRequest) (*notifpb.DeleteNotificationResponse, error) {
	if req.NotificationId == "" {
		return nil, status.Error(codes.InvalidArgument, "notification_id is required")
	}

	notifID, err := uuid.Parse(req.NotificationId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid notification_id format")
	}

	// Use a placeholder userID for now
	placeholderUserID := uuid.Nil

	if err := s.notifService.Delete(placeholderUserID, notifID); err != nil {
		return nil, convertError(err)
	}

	return &notifpb.DeleteNotificationResponse{
		Success: true,
	}, nil
}

// DeleteAllRead deletes all read notifications.
func (s *NotificationServer) DeleteAllRead(ctx context.Context, req *notifpb.DeleteAllReadRequest) (*notifpb.DeleteAllReadResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
	}

	count, err := s.notifService.DeleteAllRead(userID)
	if err != nil {
		return nil, convertError(err)
	}

	return &notifpb.DeleteAllReadResponse{
		DeletedCount: int32(count),
	}, nil
}

// GetSettings returns notification settings.
func (s *NotificationServer) GetSettings(ctx context.Context, req *notifpb.GetSettingsRequest) (*notifpb.GetSettingsResponse, error) {
	// TODO: Implement settings
	return &notifpb.GetSettingsResponse{
		Settings: &notifpb.NotificationSettings{
			UserId:       req.UserId,
			EmailEnabled: true,
			PushEnabled:  true,
			InAppEnabled: true,
		},
	}, nil
}

// UpdateSettings updates notification settings.
func (s *NotificationServer) UpdateSettings(ctx context.Context, req *notifpb.UpdateSettingsRequest) (*notifpb.UpdateSettingsResponse, error) {
	// TODO: Implement settings
	return &notifpb.UpdateSettingsResponse{
		Settings: req.Settings,
	}, nil
}

// =============================================================================
// Helper Functions
// =============================================================================

func convertError(err error) error {
	if err == nil {
		return nil
	}

	switch err.Error() {
	case "notification not found":
		return status.Error(codes.NotFound, err.Error())
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}

func mapNotificationType(t notifpb.NotificationType) repository.NotificationType {
	switch t {
	case notifpb.NotificationType_NOTIFICATION_TYPE_SYSTEM:
		return repository.NotificationTypeSystem
	case notifpb.NotificationType_NOTIFICATION_TYPE_WORKSPACE:
		return repository.NotificationTypeWorkspaceInvite
	case notifpb.NotificationType_NOTIFICATION_TYPE_REQUEST:
		return repository.NotificationTypeRequestStatus
	case notifpb.NotificationType_NOTIFICATION_TYPE_COMPLIANCE:
		return repository.NotificationTypeComplianceWarning
	case notifpb.NotificationType_NOTIFICATION_TYPE_AI:
		return repository.NotificationTypeAIComplete
	case notifpb.NotificationType_NOTIFICATION_TYPE_COLLABORATION:
		return repository.NotificationTypeWorkspaceInvite
	default:
		return repository.NotificationTypeSystem
	}
}

func notifToProto(n *repository.Notification) *notifpb.Notification {
	notif := &notifpb.Notification{
		Id:        n.ID.String(),
		UserId:    n.UserID.String(),
		Type:      notifpb.NotificationType_NOTIFICATION_TYPE_SYSTEM,
		Priority:  notifpb.NotificationPriority_NOTIFICATION_PRIORITY_NORMAL,
		Title:     n.Title,
		Message:   n.Message,
		Data:      make(map[string]string),
		Read:      n.Read,
		CreatedAt: timestamppb.New(n.CreatedAt),
	}

	for k, v := range n.Data {
		notif.Data[k] = v
	}

	if n.ReadAt != nil {
		notif.ReadAt = timestamppb.New(*n.ReadAt)
	}

	return notif
}
