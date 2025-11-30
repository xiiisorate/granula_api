// =============================================================================
// Package handlers provides HTTP handlers for API Gateway.
// =============================================================================
// RequestHandler handles expert request HTTP requests including:
// - Create: Create a new expert request
// - List: List user's requests
// - Get: Get request details
// - Update: Update request info
// - Cancel: Cancel a request
// - Submit: Submit request for review
// - AddDocument: Upload document to request
// - GetDocuments: List request documents
//
// Documentation: docs/api/requests.md
//
// Request types:
// - consultation: Online consultation
// - documentation: Document preparation
// - expert_visit: Expert site visit
// - full_service: Full service package
// =============================================================================
package handlers

import (
	"context"
	"io"
	"time"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc"

	commonpb "github.com/xiiisorate/granula_api/shared/gen/common/v1"
	requestpb "github.com/xiiisorate/granula_api/shared/gen/request/v1"
)

// =============================================================================
// RequestHandler
// =============================================================================

// RequestHandler handles expert request HTTP requests.
// It communicates with the Request microservice via gRPC.
type RequestHandler struct {
	// client is the gRPC client for RequestService.
	client requestpb.RequestServiceClient
}

// NewRequestHandler creates a new RequestHandler.
//
// Parameters:
//   - conn: gRPC connection to the Request service
//
// Returns:
//   - *RequestHandler: New handler instance
func NewRequestHandler(conn *grpc.ClientConn) *RequestHandler {
	return &RequestHandler{
		client: requestpb.NewRequestServiceClient(conn),
	}
}

// =============================================================================
// Create - POST /requests
// =============================================================================

// CreateRequestInput represents input for creating a new request.
type CreateRequestInput struct {
	// WorkspaceID - ID воркспейса (обязательно)
	WorkspaceID string `json:"workspace_id" validate:"required,uuid"`
	// SceneID - ID сцены (опционально)
	SceneID string `json:"scene_id,omitempty"`
	// BranchID - ID ветки (опционально)
	BranchID string `json:"branch_id,omitempty"`
	// Title - заголовок заявки (обязательно)
	Title string `json:"title" validate:"required,min=5,max=255"`
	// Description - описание
	Description string `json:"description,omitempty" validate:"max=2000"`
	// Category - категория (consultation, verification, project, approval)
	Category string `json:"category" validate:"required,oneof=consultation verification project approval"`
	// Priority - приоритет (low, normal, high, urgent)
	Priority string `json:"priority,omitempty" default:"normal"`
	// Contact - контактная информация
	Contact ContactInput `json:"contact" validate:"required"`
	// PreferredTime - предпочтительное время связи
	PreferredTime string `json:"preferred_time,omitempty" validate:"max=255"`
	// Comment - дополнительный комментарий
	Comment string `json:"comment,omitempty" validate:"max=2000"`
}

// ContactInput represents contact information.
type ContactInput struct {
	// Name - имя контакта (обязательно)
	Name string `json:"name" validate:"required,min=2,max=255"`
	// Phone - телефон (обязательно)
	Phone string `json:"phone" validate:"required"`
	// Email - email (обязательно)
	Email string `json:"email" validate:"required,email"`
}

// Create создаёт новую заявку на экспертную проверку.
//
// @Summary Создать заявку
// @Description Создать заявку на консультацию, проверку планировки или согласование.
// @Tags requests
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body CreateRequestInput true "Данные заявки"
// @Success 201 {object} RequestResponse "Заявка создана"
// @Failure 400 {object} ErrorResponse "Некорректный запрос"
// @Failure 401 {object} ErrorResponse "Не авторизован"
// @Router /requests [post]
func (h *RequestHandler) Create(c *fiber.Ctx) error {
	// Extract user ID from context (set by auth middleware)
	userIDStr, ok := c.Locals("user_id").(string)
	if !ok || userIDStr == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "user not authenticated")
	}
	userID := userIDStr

	// Parse request body
	var input CreateRequestInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	// Validate required fields
	if input.WorkspaceID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "workspace_id is required")
	}
	if input.Title == "" {
		return fiber.NewError(fiber.StatusBadRequest, "title is required")
	}
	if input.Category == "" {
		return fiber.NewError(fiber.StatusBadRequest, "category is required")
	}
	if input.Contact.Name == "" || input.Contact.Phone == "" || input.Contact.Email == "" {
		return fiber.NewError(fiber.StatusBadRequest, "contact.name, contact.phone and contact.email are required")
	}

	// Default priority
	if input.Priority == "" {
		input.Priority = "normal"
	}

	// Create gRPC request
	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	req := &requestpb.CreateRequestRequest{
		WorkspaceId: input.WorkspaceID,
		SceneId:     input.SceneID,
		BranchId:    input.BranchID,
		Title:       input.Title,
		Description: input.Description,
		Category:    stringToRequestCategory(input.Category),
		Priority:    stringToRequestPriority(input.Priority),
		Contact: &requestpb.Contact{
			Name:  input.Contact.Name,
			Phone: input.Contact.Phone,
			Email: input.Contact.Email,
		},
		PreferredTime: input.PreferredTime,
		Comment:       input.Comment,
		UserId:        userID,
	}

	// Call gRPC service
	resp, err := h.client.CreateRequest(ctx, req)
	if err != nil {
		return HandleGRPCError(err)
	}

	// Return success response
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data":       requestToResponse(resp.Request, userID),
		"message":    "Request created successfully",
		"request_id": c.GetRespHeader("X-Request-ID"),
	})
}

// =============================================================================
// List - GET /requests
// =============================================================================

// List возвращает список заявок пользователя.
//
// @Summary Список заявок
// @Description Получить список заявок текущего пользователя с фильтрацией и пагинацией.
// @Tags requests
// @Produce json
// @Security BearerAuth
// @Param workspace_id query string false "Фильтр по воркспейсу"
// @Param status query string false "Фильтр по статусу"
// @Param category query string false "Фильтр по категории"
// @Param limit query int false "Количество записей" default(20)
// @Param page query int false "Номер страницы" default(1)
// @Success 200 {object} RequestsListResponse "Список заявок"
// @Failure 401 {object} ErrorResponse "Не авторизован"
// @Router /requests [get]
func (h *RequestHandler) List(c *fiber.Ctx) error {
	// Extract user ID from context
	userID := GetUserIDFromContext(c)

	// Get filter parameters
	workspaceID := c.Query("workspace_id", "")
	statusStr := c.Query("status", "")
	categoryStr := c.Query("category", "")

	// Get pagination
	pagination := GetPaginationParams(c)

	// Create gRPC request
	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	req := &requestpb.ListRequestsRequest{
		UserId:      userID,
		WorkspaceId: workspaceID,
		Status:      stringToRequestStatus(statusStr),
		Category:    stringToRequestCategory(categoryStr),
		Pagination: &commonpb.PaginationRequest{
			Page:     int32(pagination.Page),
			PageSize: int32(pagination.Limit),
		},
	}

	// Call gRPC service
	resp, err := h.client.ListRequests(ctx, req)
	if err != nil {
		return HandleGRPCError(err)
	}

	// Convert requests to response format
	requests := make([]fiber.Map, 0, len(resp.Requests))
	for _, r := range resp.Requests {
		requests = append(requests, requestBriefToResponse(r))
	}

	// Get total from pagination response
	total := 0
	if resp.Pagination != nil {
		total = int(resp.Pagination.Total)
	}

	// Return response
	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"requests": requests,
			"total":    total,
		},
		"pagination": PaginationResponse(pagination.Page, pagination.Limit, total),
		"request_id": c.GetRespHeader("X-Request-ID"),
	})
}

// =============================================================================
// Get - GET /requests/:id
// =============================================================================

// Get возвращает заявку по ID.
//
// @Summary Получить заявку
// @Description Получить детальную информацию о заявке, включая историю статусов и документы.
// @Tags requests
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID заявки"
// @Success 200 {object} RequestResponse "Данные заявки"
// @Failure 401 {object} ErrorResponse "Не авторизован"
// @Failure 403 {object} ErrorResponse "Доступ запрещён"
// @Failure 404 {object} ErrorResponse "Заявка не найдена"
// @Router /requests/{id} [get]
func (h *RequestHandler) Get(c *fiber.Ctx) error {
	// Extract user ID from context
	userID := GetUserIDFromContext(c)

	// Get request ID from path
	id := c.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "request ID is required")
	}

	// Create gRPC request
	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	req := &requestpb.GetRequestRequest{
		RequestId: id,
	}

	// Call gRPC service
	resp, err := h.client.GetRequest(ctx, req)
	if err != nil {
		return HandleGRPCError(err)
	}

	// Return response
	return SuccessResponseData(c, requestToResponse(resp.Request, userID))
}

// =============================================================================
// Update - PATCH /requests/:id
// =============================================================================

// UpdateRequestInput represents input for updating a request.
type UpdateRequestInput struct {
	// Title - новый заголовок
	Title string `json:"title,omitempty" validate:"omitempty,min=5,max=255"`
	// Description - новое описание
	Description string `json:"description,omitempty" validate:"omitempty,max=2000"`
	// Contact - новая контактная информация
	Contact *ContactInput `json:"contact,omitempty"`
	// PreferredTime - новое предпочтительное время
	PreferredTime string `json:"preferred_time,omitempty" validate:"omitempty,max=255"`
	// Comment - новый комментарий
	Comment string `json:"comment,omitempty" validate:"omitempty,max=2000"`
}

// Update обновляет данные заявки.
//
// @Summary Обновить заявку
// @Description Обновить данные заявки. Доступно только в статусах draft и pending.
// @Tags requests
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID заявки"
// @Param body body UpdateRequestInput true "Данные для обновления"
// @Success 200 {object} RequestResponse "Заявка обновлена"
// @Failure 400 {object} ErrorResponse "Некорректный запрос или недопустимый статус"
// @Failure 401 {object} ErrorResponse "Не авторизован"
// @Failure 404 {object} ErrorResponse "Заявка не найдена"
// @Router /requests/{id} [patch]
func (h *RequestHandler) Update(c *fiber.Ctx) error {
	// Extract user ID from context
	userID := GetUserIDFromContext(c)

	// Get request ID from path
	id := c.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "request ID is required")
	}

	// Parse request body
	var input UpdateRequestInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	// Create gRPC request
	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	req := &requestpb.UpdateRequestRequest{
		RequestId:     id,
		Title:         input.Title,
		Description:   input.Description,
		PreferredTime: input.PreferredTime,
		Comment:       input.Comment,
	}

	// Add contact if provided
	if input.Contact != nil {
		req.Contact = &requestpb.Contact{
			Name:  input.Contact.Name,
			Phone: input.Contact.Phone,
			Email: input.Contact.Email,
		}
	}

	// Call gRPC service
	resp, err := h.client.UpdateRequest(ctx, req)
	if err != nil {
		return HandleGRPCError(err)
	}

	// Return response
	return SuccessResponseDataMessage(c, requestToResponse(resp.Request, userID), "Request updated successfully")
}

// =============================================================================
// Cancel - POST /requests/:id/cancel
// =============================================================================

// CancelRequestInput represents input for cancelling a request.
type CancelRequestInput struct {
	// Reason - причина отмены
	Reason string `json:"reason,omitempty" validate:"max=1000"`
}

// Cancel отменяет заявку.
//
// @Summary Отменить заявку
// @Description Отменить заявку. Доступно в любом статусе кроме completed.
// @Tags requests
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID заявки"
// @Param body body CancelRequestInput false "Причина отмены"
// @Success 200 {object} SuccessResponse "Заявка отменена"
// @Failure 400 {object} ErrorResponse "Недопустимый статус"
// @Failure 401 {object} ErrorResponse "Не авторизован"
// @Failure 404 {object} ErrorResponse "Заявка не найдена"
// @Router /requests/{id}/cancel [post]
func (h *RequestHandler) Cancel(c *fiber.Ctx) error {
	// Get request ID from path
	id := c.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "request ID is required")
	}

	// Parse optional request body
	var input CancelRequestInput
	c.BodyParser(&input)

	// Create gRPC request
	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	req := &requestpb.CancelRequestRequest{
		RequestId: id,
		Reason:    input.Reason,
	}

	// Call gRPC service
	_, err := h.client.CancelRequest(ctx, req)
	if err != nil {
		return HandleGRPCError(err)
	}

	// Return success response
	return SuccessResponseMessage(c, "Request cancelled successfully")
}

// =============================================================================
// Submit - POST /requests/:id/submit
// =============================================================================

// Submit отправляет заявку на рассмотрение.
//
// @Summary Отправить заявку
// @Description Отправить черновик заявки на рассмотрение экспертам.
// @Tags requests
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID заявки"
// @Success 200 {object} RequestResponse "Заявка отправлена"
// @Failure 400 {object} ErrorResponse "Заявка уже отправлена или некорректные данные"
// @Failure 401 {object} ErrorResponse "Не авторизован"
// @Failure 404 {object} ErrorResponse "Заявка не найдена"
// @Router /requests/{id}/submit [post]
func (h *RequestHandler) Submit(c *fiber.Ctx) error {
	// Extract user ID from context
	userID := GetUserIDFromContext(c)

	// Get request ID from path
	id := c.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "request ID is required")
	}

	// Create gRPC request to update status
	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	req := &requestpb.UpdateStatusRequest{
		RequestId: id,
		Status:    requestpb.RequestStatus_REQUEST_STATUS_PENDING,
		Comment:   "Submitted by user",
	}

	// Call gRPC service
	resp, err := h.client.UpdateStatus(ctx, req)
	if err != nil {
		return HandleGRPCError(err)
	}

	// Return response
	return c.JSON(fiber.Map{
		"data":       requestToResponse(resp.Request, userID),
		"message":    "Request submitted successfully",
		"request_id": c.GetRespHeader("X-Request-ID"),
	})
}

// =============================================================================
// AddDocument - POST /requests/:id/documents
// =============================================================================

// AddDocument загружает документ к заявке.
//
// @Summary Загрузить документ
// @Description Прикрепить документ к заявке (проект, заключение и т.д.).
// @Tags requests
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID заявки"
// @Param file formData file true "Файл документа"
// @Success 201 {object} DocumentResponse "Документ загружен"
// @Failure 400 {object} ErrorResponse "Некорректный запрос"
// @Failure 401 {object} ErrorResponse "Не авторизован"
// @Failure 404 {object} ErrorResponse "Заявка не найдена"
// @Failure 413 {object} ErrorResponse "Файл слишком большой"
// @Router /requests/{id}/documents [post]
func (h *RequestHandler) AddDocument(c *fiber.Ctx) error {
	// Get request ID from path
	id := c.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "request ID is required")
	}

	// Get file from form
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "file is required")
	}

	// Open and read file
	file, err := fileHeader.Open()
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "failed to open uploaded file")
	}
	defer file.Close()

	fileData, err := io.ReadAll(file)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "failed to read uploaded file")
	}

	// Get content type
	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Create gRPC request
	ctx, cancel := context.WithTimeout(c.Context(), 60*time.Second)
	defer cancel()

	req := &requestpb.UploadDocumentRequest{
		RequestId:   id,
		Filename:    fileHeader.Filename,
		ContentType: contentType,
		Data:        fileData,
	}

	// Call gRPC service
	resp, err := h.client.UploadDocument(ctx, req)
	if err != nil {
		return HandleGRPCError(err)
	}

	// Return success response
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data":       documentToResponse(resp.Document),
		"message":    "Document uploaded successfully",
		"request_id": c.GetRespHeader("X-Request-ID"),
	})
}

// =============================================================================
// GetDocuments - GET /requests/:id/documents
// =============================================================================

// GetDocuments возвращает список документов заявки.
//
// @Summary Список документов
// @Description Получить список документов, прикреплённых к заявке.
// @Tags requests
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID заявки"
// @Success 200 {object} DocumentsListResponse "Список документов"
// @Failure 401 {object} ErrorResponse "Не авторизован"
// @Failure 404 {object} ErrorResponse "Заявка не найдена"
// @Router /requests/{id}/documents [get]
func (h *RequestHandler) GetDocuments(c *fiber.Ctx) error {
	// Get request ID from path
	id := c.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "request ID is required")
	}

	// Create gRPC request
	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	req := &requestpb.GetDocumentsRequest{
		RequestId: id,
	}

	// Call gRPC service
	resp, err := h.client.GetDocuments(ctx, req)
	if err != nil {
		return HandleGRPCError(err)
	}

	// Convert documents to response format
	documents := make([]fiber.Map, 0, len(resp.Documents))
	for _, d := range resp.Documents {
		documents = append(documents, documentToResponse(d))
	}

	// Return response
	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"documents": documents,
			"total":     len(documents),
		},
		"request_id": c.GetRespHeader("X-Request-ID"),
	})
}

// =============================================================================
// Delete - DELETE /requests/:id
// =============================================================================

// Delete удаляет заявку (отменяет).
//
// @Summary Удалить заявку
// @Description Удалить (отменить) заявку. Эквивалентно Cancel.
// @Tags requests
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID заявки"
// @Success 200 {object} SuccessResponse "Заявка удалена"
// @Failure 400 {object} ErrorResponse "Недопустимый статус"
// @Failure 401 {object} ErrorResponse "Не авторизован"
// @Failure 404 {object} ErrorResponse "Заявка не найдена"
// @Router /requests/{id} [delete]
func (h *RequestHandler) Delete(c *fiber.Ctx) error {
	// Delegate to Cancel
	return h.Cancel(c)
}

// =============================================================================
// Helper Functions
// =============================================================================

// requestToResponse converts proto Request to API response format (full).
func requestToResponse(r *requestpb.Request, currentUserID string) fiber.Map {
	if r == nil {
		return nil
	}

	result := fiber.Map{
		"id":           r.Id,
		"workspace_id": r.WorkspaceId,
		"user_id":      r.UserId,
		"title":        r.Title,
		"description":  r.Description,
		"category":     requestCategoryToString(r.Category),
		"priority":     requestPriorityToString(r.Priority),
		"status":       requestStatusToString(r.Status),
		"comment":      r.Comment,
	}

	// Add optional scene/branch
	if r.SceneId != "" {
		result["scene_id"] = r.SceneId
	}
	if r.BranchId != "" {
		result["branch_id"] = r.BranchId
	}

	// Add contact
	if r.Contact != nil {
		result["contact"] = fiber.Map{
			"name":  r.Contact.Name,
			"phone": r.Contact.Phone,
			"email": r.Contact.Email,
		}
	}

	// Add preferred time
	if r.PreferredTime != "" {
		result["preferred_time"] = r.PreferredTime
	}

	// Add expert info
	if r.Expert != nil {
		result["expert"] = fiber.Map{
			"id":             r.Expert.Id,
			"name":           r.Expert.Name,
			"specialization": r.Expert.Specialization,
			"rating":         r.Expert.Rating,
		}
	}

	// Add estimated info
	if r.EstimatedDate != nil {
		result["estimated_date"] = r.EstimatedDate.AsTime()
	}
	if r.EstimatedPrice > 0 {
		result["estimated_price"] = r.EstimatedPrice
	}

	// Add rejection reason
	if r.RejectionReason != "" {
		result["rejection_reason"] = r.RejectionReason
	}

	// Add compliance snapshot
	if r.ComplianceSnapshot != nil {
		result["compliance_snapshot"] = fiber.Map{
			"check_id":         r.ComplianceSnapshot.CheckId,
			"is_compliant":     r.ComplianceSnapshot.IsCompliant,
			"violations_count": r.ComplianceSnapshot.ViolationsCount,
			"warnings_count":   r.ComplianceSnapshot.WarningsCount,
		}
		if r.ComplianceSnapshot.CheckedAt != nil {
			result["compliance_snapshot"].(fiber.Map)["checked_at"] = r.ComplianceSnapshot.CheckedAt.AsTime()
		}
	}

	// Add documents
	if len(r.Documents) > 0 {
		docs := make([]fiber.Map, 0, len(r.Documents))
		for _, d := range r.Documents {
			docs = append(docs, documentToResponse(d))
		}
		result["documents"] = docs
	}

	// Add timestamps
	if r.CreatedAt != nil {
		result["created_at"] = r.CreatedAt.AsTime()
	}
	if r.UpdatedAt != nil {
		result["updated_at"] = r.UpdatedAt.AsTime()
	}

	return result
}

// requestBriefToResponse converts proto Request to API response format (brief).
func requestBriefToResponse(r *requestpb.Request) fiber.Map {
	if r == nil {
		return nil
	}

	result := fiber.Map{
		"id":           r.Id,
		"workspace_id": r.WorkspaceId,
		"title":        r.Title,
		"category":     requestCategoryToString(r.Category),
		"priority":     requestPriorityToString(r.Priority),
		"status":       requestStatusToString(r.Status),
	}

	// Add expert name if assigned
	if r.Expert != nil {
		result["expert_name"] = r.Expert.Name
	}

	// Add timestamps
	if r.CreatedAt != nil {
		result["created_at"] = r.CreatedAt.AsTime()
	}
	if r.UpdatedAt != nil {
		result["updated_at"] = r.UpdatedAt.AsTime()
	}

	return result
}

// documentToResponse converts proto Document to API response format.
func documentToResponse(d *requestpb.Document) fiber.Map {
	if d == nil {
		return nil
	}

	result := fiber.Map{
		"id":           d.Id,
		"filename":     d.Filename,
		"content_type": d.ContentType,
		"size":         d.Size,
		"url":          d.Url,
		"uploaded_by":  d.UploadedBy,
	}

	if d.UploadedAt != nil {
		result["uploaded_at"] = d.UploadedAt.AsTime()
	}

	return result
}

// requestStatusToString converts proto RequestStatus to string.
func requestStatusToString(s requestpb.RequestStatus) string {
	switch s {
	case requestpb.RequestStatus_REQUEST_STATUS_DRAFT:
		return "draft"
	case requestpb.RequestStatus_REQUEST_STATUS_PENDING:
		return "pending"
	case requestpb.RequestStatus_REQUEST_STATUS_IN_REVIEW:
		return "in_review"
	case requestpb.RequestStatus_REQUEST_STATUS_APPROVED:
		return "approved"
	case requestpb.RequestStatus_REQUEST_STATUS_REJECTED:
		return "rejected"
	case requestpb.RequestStatus_REQUEST_STATUS_COMPLETED:
		return "completed"
	case requestpb.RequestStatus_REQUEST_STATUS_CANCELLED:
		return "cancelled"
	default:
		return "unknown"
	}
}

// stringToRequestStatus converts string to proto RequestStatus.
func stringToRequestStatus(s string) requestpb.RequestStatus {
	switch s {
	case "draft":
		return requestpb.RequestStatus_REQUEST_STATUS_DRAFT
	case "pending":
		return requestpb.RequestStatus_REQUEST_STATUS_PENDING
	case "in_review":
		return requestpb.RequestStatus_REQUEST_STATUS_IN_REVIEW
	case "approved":
		return requestpb.RequestStatus_REQUEST_STATUS_APPROVED
	case "rejected":
		return requestpb.RequestStatus_REQUEST_STATUS_REJECTED
	case "completed":
		return requestpb.RequestStatus_REQUEST_STATUS_COMPLETED
	case "cancelled":
		return requestpb.RequestStatus_REQUEST_STATUS_CANCELLED
	default:
		return requestpb.RequestStatus_REQUEST_STATUS_UNSPECIFIED
	}
}

// requestCategoryToString converts proto RequestCategory to string.
func requestCategoryToString(c requestpb.RequestCategory) string {
	switch c {
	case requestpb.RequestCategory_REQUEST_CATEGORY_CONSULTATION:
		return "consultation"
	case requestpb.RequestCategory_REQUEST_CATEGORY_VERIFICATION:
		return "verification"
	case requestpb.RequestCategory_REQUEST_CATEGORY_PROJECT:
		return "project"
	case requestpb.RequestCategory_REQUEST_CATEGORY_APPROVAL:
		return "approval"
	default:
		return "unknown"
	}
}

// stringToRequestCategory converts string to proto RequestCategory.
func stringToRequestCategory(s string) requestpb.RequestCategory {
	switch s {
	case "consultation":
		return requestpb.RequestCategory_REQUEST_CATEGORY_CONSULTATION
	case "verification":
		return requestpb.RequestCategory_REQUEST_CATEGORY_VERIFICATION
	case "project":
		return requestpb.RequestCategory_REQUEST_CATEGORY_PROJECT
	case "approval":
		return requestpb.RequestCategory_REQUEST_CATEGORY_APPROVAL
	default:
		return requestpb.RequestCategory_REQUEST_CATEGORY_UNSPECIFIED
	}
}

// requestPriorityToString converts proto RequestPriority to string.
func requestPriorityToString(p requestpb.RequestPriority) string {
	switch p {
	case requestpb.RequestPriority_REQUEST_PRIORITY_LOW:
		return "low"
	case requestpb.RequestPriority_REQUEST_PRIORITY_NORMAL:
		return "normal"
	case requestpb.RequestPriority_REQUEST_PRIORITY_HIGH:
		return "high"
	case requestpb.RequestPriority_REQUEST_PRIORITY_URGENT:
		return "urgent"
	default:
		return "normal"
	}
}

// stringToRequestPriority converts string to proto RequestPriority.
func stringToRequestPriority(s string) requestpb.RequestPriority {
	switch s {
	case "low":
		return requestpb.RequestPriority_REQUEST_PRIORITY_LOW
	case "normal":
		return requestpb.RequestPriority_REQUEST_PRIORITY_NORMAL
	case "high":
		return requestpb.RequestPriority_REQUEST_PRIORITY_HIGH
	case "urgent":
		return requestpb.RequestPriority_REQUEST_PRIORITY_URGENT
	default:
		return requestpb.RequestPriority_REQUEST_PRIORITY_NORMAL
	}
}

// =============================================================================
// Response Types (for Swagger documentation)
// =============================================================================

// RequestResponse represents an expert request in API response.
// swagger:model
type RequestResponse struct {
	Data      interface{} `json:"data"`
	Message   string      `json:"message,omitempty"`
	RequestID string      `json:"request_id"`
}

// RequestsListResponse represents a list of requests.
// swagger:model
type RequestsListResponse struct {
	Data       RequestsListData `json:"data"`
	Pagination fiber.Map        `json:"pagination"`
	RequestID  string           `json:"request_id"`
}

// RequestsListData contains requests list data.
type RequestsListData struct {
	Requests []interface{} `json:"requests"`
	Total    int           `json:"total"`
}

// DocumentResponse represents a document in API response.
// swagger:model
type DocumentResponse struct {
	Data      interface{} `json:"data"`
	Message   string      `json:"message,omitempty"`
	RequestID string      `json:"request_id"`
}

// DocumentsListResponse represents a list of documents.
// swagger:model
type DocumentsListResponse struct {
	Data      DocumentsListData `json:"data"`
	RequestID string            `json:"request_id"`
}

// DocumentsListData contains documents list data.
type DocumentsListData struct {
	Documents []interface{} `json:"documents"`
	Total     int           `json:"total"`
}
