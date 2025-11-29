// =============================================================================
// Package entity defines domain entities for Request Service.
// =============================================================================
// This package contains the core business entities for expert request management.
// It models the workflow for users to request professional consultations
// from BTI (Bureau of Technical Inventory) experts.
//
// Request Workflow:
//
//	Created → Pending Review → In Review → Approved/Rejected
//	                                    ↓
//	                        Assigned → In Progress → Completed
//
// =============================================================================
package entity

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// =============================================================================
// Domain Errors
// =============================================================================

var (
	// ErrRequestNotFound is returned when a request cannot be found.
	ErrRequestNotFound = errors.New("request not found")

	// ErrInvalidRequestStatus is returned for invalid status transitions.
	ErrInvalidRequestStatus = errors.New("invalid request status")

	// ErrCannotModifyRequest is returned when request cannot be modified.
	ErrCannotModifyRequest = errors.New("cannot modify request in current status")

	// ErrInvalidRequestTitle is returned when title validation fails.
	ErrInvalidRequestTitle = errors.New("request title must be 5-200 characters")

	// ErrExpertAlreadyAssigned is returned when trying to assign to already assigned request.
	ErrExpertAlreadyAssigned = errors.New("expert already assigned to this request")
)

// =============================================================================
// Request Status Constants
// =============================================================================

// RequestStatus represents the current state of an expert request.
type RequestStatus string

const (
	// StatusDraft - Request is being created, not yet submitted.
	StatusDraft RequestStatus = "draft"

	// StatusPending - Request submitted, waiting for review.
	StatusPending RequestStatus = "pending"

	// StatusInReview - Request is being reviewed by staff.
	StatusInReview RequestStatus = "in_review"

	// StatusApproved - Request approved, waiting for expert assignment.
	StatusApproved RequestStatus = "approved"

	// StatusRejected - Request rejected with reason.
	StatusRejected RequestStatus = "rejected"

	// StatusAssigned - Expert assigned, waiting for work to begin.
	StatusAssigned RequestStatus = "assigned"

	// StatusInProgress - Expert is working on the request.
	StatusInProgress RequestStatus = "in_progress"

	// StatusCompleted - Work completed successfully.
	StatusCompleted RequestStatus = "completed"

	// StatusCancelled - Request cancelled by user.
	StatusCancelled RequestStatus = "cancelled"
)

// ValidStatuses contains all valid request statuses.
var ValidStatuses = []RequestStatus{
	StatusDraft, StatusPending, StatusInReview, StatusApproved,
	StatusRejected, StatusAssigned, StatusInProgress, StatusCompleted, StatusCancelled,
}

// IsValid checks if the status is valid.
func (s RequestStatus) IsValid() bool {
	for _, valid := range ValidStatuses {
		if s == valid {
			return true
		}
	}
	return false
}

// String returns the string representation.
func (s RequestStatus) String() string {
	return string(s)
}

// IsFinal returns true if this is a terminal status.
func (s RequestStatus) IsFinal() bool {
	return s == StatusCompleted || s == StatusRejected || s == StatusCancelled
}

// CanTransitionTo checks if transition to target status is allowed.
func (s RequestStatus) CanTransitionTo(target RequestStatus) bool {
	validTransitions := map[RequestStatus][]RequestStatus{
		StatusDraft:      {StatusPending, StatusCancelled},
		StatusPending:    {StatusInReview, StatusCancelled},
		StatusInReview:   {StatusApproved, StatusRejected},
		StatusApproved:   {StatusAssigned, StatusCancelled},
		StatusRejected:   {}, // Final state
		StatusAssigned:   {StatusInProgress, StatusCancelled},
		StatusInProgress: {StatusCompleted, StatusCancelled},
		StatusCompleted:  {}, // Final state
		StatusCancelled:  {}, // Final state
	}

	allowed, exists := validTransitions[s]
	if !exists {
		return false
	}

	for _, a := range allowed {
		if a == target {
			return true
		}
	}
	return false
}

// =============================================================================
// Request Category Constants
// =============================================================================

// RequestCategory represents the type of expert service requested.
type RequestCategory string

const (
	// CategoryConsultation - General consultation service.
	CategoryConsultation RequestCategory = "consultation"

	// CategoryDocumentation - Documentation preparation service.
	CategoryDocumentation RequestCategory = "documentation"

	// CategoryExpertVisit - On-site expert visit.
	CategoryExpertVisit RequestCategory = "expert_visit"

	// CategoryFullPackage - Complete package of all services.
	CategoryFullPackage RequestCategory = "full_package"
)

// Price returns the base price for this category in rubles.
func (c RequestCategory) Price() int {
	prices := map[RequestCategory]int{
		CategoryConsultation:  2000,
		CategoryDocumentation: 15000,
		CategoryExpertVisit:   5000,
		CategoryFullPackage:   30000,
	}
	if price, ok := prices[c]; ok {
		return price
	}
	return 0
}

// =============================================================================
// Request Priority Constants
// =============================================================================

// RequestPriority indicates urgency of the request.
type RequestPriority string

const (
	// PriorityLow - Standard processing time (5-7 business days).
	PriorityLow RequestPriority = "low"

	// PriorityNormal - Normal processing time (3-5 business days).
	PriorityNormal RequestPriority = "normal"

	// PriorityHigh - Expedited processing (1-2 business days).
	PriorityHigh RequestPriority = "high"

	// PriorityUrgent - Same-day or next-day processing.
	PriorityUrgent RequestPriority = "urgent"
)

// =============================================================================
// Request Entity
// =============================================================================

// Request represents an expert service request from a user.
// It tracks the full lifecycle from creation to completion.
type Request struct {
	// ID is the unique identifier (UUID v4).
	ID uuid.UUID `json:"id" db:"id"`

	// WorkspaceID links the request to a workspace.
	WorkspaceID uuid.UUID `json:"workspace_id" db:"workspace_id"`

	// UserID is the user who created the request.
	UserID uuid.UUID `json:"user_id" db:"user_id"`

	// Title is a brief summary of the request (5-200 chars).
	Title string `json:"title" db:"title"`

	// Description provides detailed information about the request.
	Description string `json:"description" db:"description"`

	// Category is the type of service requested.
	Category RequestCategory `json:"category" db:"category"`

	// Priority indicates urgency level.
	Priority RequestPriority `json:"priority" db:"priority"`

	// Status is the current workflow state.
	Status RequestStatus `json:"status" db:"status"`

	// ExpertID is the assigned expert's user ID (nullable).
	ExpertID *uuid.UUID `json:"expert_id,omitempty" db:"expert_id"`

	// AssignedAt is when an expert was assigned.
	AssignedAt *time.Time `json:"assigned_at,omitempty" db:"assigned_at"`

	// EstimatedCost is the quoted price in rubles.
	EstimatedCost int `json:"estimated_cost" db:"estimated_cost"`

	// FinalCost is the actual price after completion.
	FinalCost *int `json:"final_cost,omitempty" db:"final_cost"`

	// RejectionReason explains why the request was rejected.
	RejectionReason string `json:"rejection_reason,omitempty" db:"rejection_reason"`

	// Notes contains internal notes (visible to staff only).
	Notes string `json:"notes,omitempty" db:"notes"`

	// ContactPhone is the user's contact phone number.
	ContactPhone string `json:"contact_phone,omitempty" db:"contact_phone"`

	// ContactEmail is the user's contact email.
	ContactEmail string `json:"contact_email,omitempty" db:"contact_email"`

	// StatusHistory tracks all status changes.
	StatusHistory []StatusChange `json:"status_history,omitempty" db:"-"`

	// Documents attached to this request.
	Documents []Document `json:"documents,omitempty" db:"-"`

	// Timestamps
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty" db:"completed_at"`
}

// NewRequest creates a new request with default values.
//
// Parameters:
//   - workspaceID: UUID of the associated workspace
//   - userID: UUID of the creating user
//   - title: Brief summary (will be validated)
//   - category: Type of service requested
//
// Returns:
//   - *Request: New request entity in draft status
//   - error: ErrInvalidRequestTitle if title validation fails
func NewRequest(workspaceID, userID uuid.UUID, title string, category RequestCategory) (*Request, error) {
	// Validate title
	title = strings.TrimSpace(title)
	if len(title) < 5 || len(title) > 200 {
		return nil, ErrInvalidRequestTitle
	}

	now := time.Now().UTC()
	return &Request{
		ID:            uuid.New(),
		WorkspaceID:   workspaceID,
		UserID:        userID,
		Title:         title,
		Category:      category,
		Priority:      PriorityNormal,
		Status:        StatusDraft,
		EstimatedCost: category.Price(),
		StatusHistory: make([]StatusChange, 0),
		Documents:     make([]Document, 0),
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

// Submit transitions the request from draft to pending.
func (r *Request) Submit() error {
	if r.Status != StatusDraft {
		return ErrCannotModifyRequest
	}
	return r.ChangeStatus(StatusPending, "Request submitted by user", nil)
}

// ChangeStatus updates the request status with validation.
//
// Parameters:
//   - newStatus: Target status
//   - comment: Reason for the change
//   - changedBy: UUID of user making the change (nil for system)
//
// Returns:
//   - error: ErrInvalidRequestStatus if transition is not allowed
func (r *Request) ChangeStatus(newStatus RequestStatus, comment string, changedBy *uuid.UUID) error {
	if !r.Status.CanTransitionTo(newStatus) {
		return ErrInvalidRequestStatus
	}

	// Record status change
	change := StatusChange{
		ID:         uuid.New(),
		RequestID:  r.ID,
		FromStatus: r.Status,
		ToStatus:   newStatus,
		Comment:    comment,
		ChangedBy:  changedBy,
		ChangedAt:  time.Now().UTC(),
	}
	r.StatusHistory = append(r.StatusHistory, change)

	// Update status
	r.Status = newStatus
	r.UpdatedAt = time.Now().UTC()

	// Handle special cases
	if newStatus == StatusCompleted {
		now := time.Now().UTC()
		r.CompletedAt = &now
	}

	return nil
}

// AssignExpert assigns an expert to work on this request.
func (r *Request) AssignExpert(expertID uuid.UUID) error {
	if r.Status != StatusApproved {
		return ErrInvalidRequestStatus
	}
	if r.ExpertID != nil {
		return ErrExpertAlreadyAssigned
	}

	r.ExpertID = &expertID
	now := time.Now().UTC()
	r.AssignedAt = &now

	return r.ChangeStatus(StatusAssigned, "Expert assigned", nil)
}

// Reject rejects the request with a reason.
func (r *Request) Reject(reason string, rejectedBy uuid.UUID) error {
	if r.Status != StatusInReview {
		return ErrInvalidRequestStatus
	}

	r.RejectionReason = reason
	return r.ChangeStatus(StatusRejected, reason, &rejectedBy)
}

// Cancel cancels the request by the user.
func (r *Request) Cancel(userID uuid.UUID) error {
	if r.Status.IsFinal() {
		return ErrCannotModifyRequest
	}

	return r.ChangeStatus(StatusCancelled, "Cancelled by user", &userID)
}

// UpdateDescription updates the request description.
func (r *Request) UpdateDescription(description string) error {
	if r.Status != StatusDraft {
		return ErrCannotModifyRequest
	}

	r.Description = strings.TrimSpace(description)
	r.UpdatedAt = time.Now().UTC()
	return nil
}

// SetContact sets contact information.
func (r *Request) SetContact(phone, email string) {
	r.ContactPhone = strings.TrimSpace(phone)
	r.ContactEmail = strings.TrimSpace(email)
	r.UpdatedAt = time.Now().UTC()
}

// CanUserModify checks if the given user can modify this request.
func (r *Request) CanUserModify(userID uuid.UUID) bool {
	return r.UserID == userID && r.Status == StatusDraft
}

// CanUserCancel checks if the given user can cancel this request.
func (r *Request) CanUserCancel(userID uuid.UUID) bool {
	return r.UserID == userID && !r.Status.IsFinal()
}

// =============================================================================
// StatusChange Entity
// =============================================================================

// StatusChange records a status transition in the request lifecycle.
type StatusChange struct {
	// ID is the unique identifier.
	ID uuid.UUID `json:"id" db:"id"`

	// RequestID links to the request.
	RequestID uuid.UUID `json:"request_id" db:"request_id"`

	// FromStatus is the previous status.
	FromStatus RequestStatus `json:"from_status" db:"from_status"`

	// ToStatus is the new status.
	ToStatus RequestStatus `json:"to_status" db:"to_status"`

	// Comment explains the reason for change.
	Comment string `json:"comment" db:"comment"`

	// ChangedBy is the user who made the change (nil for system).
	ChangedBy *uuid.UUID `json:"changed_by,omitempty" db:"changed_by"`

	// ChangedAt is when the change occurred.
	ChangedAt time.Time `json:"changed_at" db:"changed_at"`
}

// =============================================================================
// Document Entity
// =============================================================================

// DocumentType represents the type of attached document.
type DocumentType string

const (
	// DocTypeFloorPlan - Floor plan or technical drawing.
	DocTypeFloorPlan DocumentType = "floor_plan"

	// DocTypeBTICertificate - BTI certificate or passport.
	DocTypeBTICertificate DocumentType = "bti_certificate"

	// DocTypeOwnership - Ownership documents.
	DocTypeOwnership DocumentType = "ownership"

	// DocTypeOther - Other supporting documents.
	DocTypeOther DocumentType = "other"
)

// Document represents an attached file to a request.
type Document struct {
	// ID is the unique identifier.
	ID uuid.UUID `json:"id" db:"id"`

	// RequestID links to the request.
	RequestID uuid.UUID `json:"request_id" db:"request_id"`

	// Type categorizes the document.
	Type DocumentType `json:"type" db:"type"`

	// Name is the original filename.
	Name string `json:"name" db:"name"`

	// StoragePath is the path in object storage.
	StoragePath string `json:"storage_path" db:"storage_path"`

	// MimeType is the file MIME type.
	MimeType string `json:"mime_type" db:"mime_type"`

	// Size is the file size in bytes.
	Size int64 `json:"size" db:"size"`

	// UploadedBy is the user who uploaded.
	UploadedBy uuid.UUID `json:"uploaded_by" db:"uploaded_by"`

	// UploadedAt is when the file was uploaded.
	UploadedAt time.Time `json:"uploaded_at" db:"uploaded_at"`
}

// NewDocument creates a new document record.
func NewDocument(requestID uuid.UUID, docType DocumentType, name, storagePath, mimeType string, size int64, uploadedBy uuid.UUID) *Document {
	return &Document{
		ID:          uuid.New(),
		RequestID:   requestID,
		Type:        docType,
		Name:        name,
		StoragePath: storagePath,
		MimeType:    mimeType,
		Size:        size,
		UploadedBy:  uploadedBy,
		UploadedAt:  time.Now().UTC(),
	}
}
