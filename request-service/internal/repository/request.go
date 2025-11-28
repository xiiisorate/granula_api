// Package repository handles data access for Request Service.
package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RequestRepository handles request database operations.
type RequestRepository struct {
	db *gorm.DB
}

// NewRequestRepository creates a new RequestRepository.
func NewRequestRepository(db *gorm.DB) *RequestRepository {
	return &RequestRepository{db: db}
}

// Create creates a new request.
func (r *RequestRepository) Create(request *Request) error {
	return r.db.Create(request).Error
}

// FindByID finds a request by ID.
func (r *RequestRepository) FindByID(id uuid.UUID) (*Request, error) {
	var request Request
	err := r.db.Preload("StatusHistory", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at DESC")
	}).Where("id = ?", id).First(&request).Error
	if err != nil {
		return nil, err
	}
	return &request, nil
}

// FindByWorkspaceID finds requests by workspace ID with filters.
func (r *RequestRepository) FindByWorkspaceID(workspaceID uuid.UUID, status string, page, pageSize int) ([]Request, int64, error) {
	var requests []Request
	var total int64

	query := r.db.Model(&Request{}).Where("workspace_id = ?", workspaceID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.
		Preload("StatusHistory", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC").Limit(1)
		}).
		Offset(offset).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&requests).Error

	return requests, total, err
}

// FindByUserID finds requests created by a user.
func (r *RequestRepository) FindByUserID(userID uuid.UUID, page, pageSize int) ([]Request, int64, error) {
	var requests []Request
	var total int64

	r.db.Model(&Request{}).Where("user_id = ?", userID).Count(&total)

	offset := (page - 1) * pageSize
	err := r.db.
		Preload("StatusHistory", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC").Limit(1)
		}).
		Where("user_id = ?", userID).
		Offset(offset).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&requests).Error

	return requests, total, err
}

// FindByExpertID finds requests assigned to an expert.
func (r *RequestRepository) FindByExpertID(expertID uuid.UUID, page, pageSize int) ([]Request, int64, error) {
	var requests []Request
	var total int64

	r.db.Model(&Request{}).Where("expert_id = ?", expertID).Count(&total)

	offset := (page - 1) * pageSize
	err := r.db.
		Preload("StatusHistory", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC").Limit(1)
		}).
		Where("expert_id = ?", expertID).
		Offset(offset).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&requests).Error

	return requests, total, err
}

// Update updates a request.
func (r *RequestRepository) Update(request *Request) error {
	return r.db.Save(request).Error
}

// UpdateStatus updates request status.
func (r *RequestRepository) UpdateStatus(id uuid.UUID, status string) error {
	return r.db.Model(&Request{}).Where("id = ?", id).Update("status", status).Error
}

// UpdateExpert assigns an expert to request.
func (r *RequestRepository) UpdateExpert(id uuid.UUID, expertID *uuid.UUID) error {
	return r.db.Model(&Request{}).Where("id = ?", id).Update("expert_id", expertID).Error
}

// Delete soft deletes a request.
func (r *RequestRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&Request{}, "id = ?", id).Error
}

// StatusHistoryRepository handles status history database operations.
type StatusHistoryRepository struct {
	db *gorm.DB
}

// NewStatusHistoryRepository creates a new StatusHistoryRepository.
func NewStatusHistoryRepository(db *gorm.DB) *StatusHistoryRepository {
	return &StatusHistoryRepository{db: db}
}

// Create creates a new status history entry.
func (r *StatusHistoryRepository) Create(history *StatusHistory) error {
	return r.db.Create(history).Error
}

// FindByRequestID finds status history by request ID.
func (r *StatusHistoryRepository) FindByRequestID(requestID uuid.UUID) ([]StatusHistory, error) {
	var history []StatusHistory
	err := r.db.Where("request_id = ?", requestID).Order("created_at DESC").Find(&history).Error
	return history, err
}

