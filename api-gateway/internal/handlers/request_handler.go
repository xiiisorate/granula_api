// =============================================================================
// Request Handler - HTTP handlers for expert request operations.
// =============================================================================
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/xiiisorate/granula_api/api-gateway/internal/dto"
)

// =============================================================================
// RequestHandler handles expert request-related HTTP requests.
// =============================================================================

// RequestHandler provides HTTP handlers for expert request operations.
type RequestHandler struct {
	// requestClient pb.RequestServiceClient
}

// NewRequestHandler creates a new RequestHandler.
func NewRequestHandler() *RequestHandler {
	return &RequestHandler{}
}

// CreateRequest godoc
// @Summary Создать заявку
// @Description Создает новую заявку на услуги эксперта
// @Tags requests
// @Accept json
// @Produce json
// @Param request body dto.CreateExpertRequestRequest true "Данные заявки"
// @Success 201 {object} dto.ExpertRequestResponse "Созданная заявка"
// @Failure 400 {object} dto.ErrorResponse "Ошибка валидации"
// @Failure 401 {object} dto.ErrorResponse "Не авторизован"
// @Failure 500 {object} dto.ErrorResponse "Внутренняя ошибка"
// @Security BearerAuth
// @Router /requests [post]
func (h *RequestHandler) CreateRequest(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateExpertRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request body")
		return
	}

	// TODO: Get user ID from context
	// TODO: Call request gRPC service

	response := dto.ExpertRequestResponse{
		ID:            "request-placeholder-id",
		WorkspaceID:   req.WorkspaceID,
		Title:         req.Title,
		Description:   req.Description,
		Category:      req.Category,
		Priority:      "normal",
		Status:        "draft",
		EstimatedCost: 2000, // Based on category
		ContactPhone:  req.ContactPhone,
		ContactEmail:  req.ContactEmail,
	}

	respondJSON(w, http.StatusCreated, response)
}

// GetRequest godoc
// @Summary Получить заявку
// @Description Возвращает информацию о заявке по ID
// @Tags requests
// @Accept json
// @Produce json
// @Param id path string true "ID заявки" format(uuid)
// @Success 200 {object} dto.ExpertRequestResponse "Данные заявки"
// @Failure 401 {object} dto.ErrorResponse "Не авторизован"
// @Failure 403 {object} dto.ErrorResponse "Доступ запрещен"
// @Failure 404 {object} dto.ErrorResponse "Заявка не найдена"
// @Security BearerAuth
// @Router /requests/{id} [get]
func (h *RequestHandler) GetRequest(w http.ResponseWriter, r *http.Request) {
	// TODO: Call request gRPC service

	response := dto.ExpertRequestResponse{
		ID:       "request-id",
		Title:    "Sample Request",
		Status:   "pending",
		Category: "consultation",
	}

	respondJSON(w, http.StatusOK, response)
}

// ListRequests godoc
// @Summary Список заявок
// @Description Возвращает список заявок пользователя
// @Tags requests
// @Accept json
// @Produce json
// @Param workspace_id query string false "Фильтр по воркспейсу" format(uuid)
// @Param status query string false "Фильтр по статусу" Enums(draft, pending, in_review, approved, rejected, assigned, in_progress, completed, cancelled)
// @Param page query int false "Номер страницы" default(1)
// @Param page_size query int false "Размер страницы" default(20)
// @Success 200 {object} dto.ExpertRequestListResponse "Список заявок"
// @Failure 401 {object} dto.ErrorResponse "Не авторизован"
// @Security BearerAuth
// @Router /requests [get]
func (h *RequestHandler) ListRequests(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// TODO: Call request gRPC service

	response := dto.ExpertRequestListResponse{
		Requests: []dto.ExpertRequestResponse{},
		Pagination: dto.PaginationResponse{
			Page:     page,
			PageSize: pageSize,
		},
	}

	respondJSON(w, http.StatusOK, response)
}

// UpdateRequest godoc
// @Summary Обновить заявку
// @Description Обновляет заявку (только для статуса draft)
// @Tags requests
// @Accept json
// @Produce json
// @Param id path string true "ID заявки" format(uuid)
// @Param request body dto.UpdateExpertRequestRequest true "Данные для обновления"
// @Success 200 {object} dto.ExpertRequestResponse "Обновленная заявка"
// @Failure 400 {object} dto.ErrorResponse "Ошибка валидации"
// @Failure 401 {object} dto.ErrorResponse "Не авторизован"
// @Failure 403 {object} dto.ErrorResponse "Нельзя редактировать заявку в текущем статусе"
// @Failure 404 {object} dto.ErrorResponse "Заявка не найдена"
// @Security BearerAuth
// @Router /requests/{id} [patch]
func (h *RequestHandler) UpdateRequest(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdateExpertRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request body")
		return
	}

	// TODO: Call request gRPC service

	response := dto.ExpertRequestResponse{
		ID:     "request-id",
		Title:  req.Title,
		Status: "draft",
	}

	respondJSON(w, http.StatusOK, response)
}

// SubmitRequest godoc
// @Summary Отправить заявку
// @Description Отправляет заявку на рассмотрение (переводит из draft в pending)
// @Tags requests
// @Accept json
// @Produce json
// @Param id path string true "ID заявки" format(uuid)
// @Success 200 {object} dto.ExpertRequestResponse "Отправленная заявка"
// @Failure 401 {object} dto.ErrorResponse "Не авторизован"
// @Failure 403 {object} dto.ErrorResponse "Заявка не в статусе draft"
// @Failure 404 {object} dto.ErrorResponse "Заявка не найдена"
// @Security BearerAuth
// @Router /requests/{id}/submit [post]
func (h *RequestHandler) SubmitRequest(w http.ResponseWriter, r *http.Request) {
	// TODO: Call request gRPC service

	response := dto.ExpertRequestResponse{
		ID:     "request-id",
		Status: "pending",
	}

	respondJSON(w, http.StatusOK, response)
}

// CancelRequest godoc
// @Summary Отменить заявку
// @Description Отменяет заявку (возможно до статуса completed)
// @Tags requests
// @Accept json
// @Produce json
// @Param id path string true "ID заявки" format(uuid)
// @Success 200 {object} dto.ExpertRequestResponse "Отмененная заявка"
// @Failure 401 {object} dto.ErrorResponse "Не авторизован"
// @Failure 403 {object} dto.ErrorResponse "Нельзя отменить заявку в текущем статусе"
// @Failure 404 {object} dto.ErrorResponse "Заявка не найдена"
// @Security BearerAuth
// @Router /requests/{id}/cancel [post]
func (h *RequestHandler) CancelRequest(w http.ResponseWriter, r *http.Request) {
	// TODO: Call request gRPC service

	response := dto.ExpertRequestResponse{
		ID:     "request-id",
		Status: "cancelled",
	}

	respondJSON(w, http.StatusOK, response)
}

// GetStatusHistory godoc
// @Summary История статусов
// @Description Возвращает историю изменения статуса заявки
// @Tags requests
// @Accept json
// @Produce json
// @Param id path string true "ID заявки" format(uuid)
// @Success 200 {object} dto.StatusHistoryResponse "История статусов"
// @Failure 401 {object} dto.ErrorResponse "Не авторизован"
// @Failure 404 {object} dto.ErrorResponse "Заявка не найдена"
// @Security BearerAuth
// @Router /requests/{id}/history [get]
func (h *RequestHandler) GetStatusHistory(w http.ResponseWriter, r *http.Request) {
	// TODO: Call request gRPC service

	response := dto.StatusHistoryResponse{
		History: []dto.StatusChangeResponse{},
	}

	respondJSON(w, http.StatusOK, response)
}

// =============================================================================
// Document Endpoints
// =============================================================================

// UploadDocument godoc
// @Summary Загрузить документ
// @Description Загружает документ к заявке
// @Tags requests
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "ID заявки" format(uuid)
// @Param type formData string true "Тип документа" Enums(floor_plan, bti_certificate, ownership, other)
// @Param file formData file true "Файл документа"
// @Success 201 {object} dto.DocumentResponse "Загруженный документ"
// @Failure 400 {object} dto.ErrorResponse "Ошибка валидации"
// @Failure 401 {object} dto.ErrorResponse "Не авторизован"
// @Failure 404 {object} dto.ErrorResponse "Заявка не найдена"
// @Failure 413 {object} dto.ErrorResponse "Файл слишком большой"
// @Security BearerAuth
// @Router /requests/{id}/documents [post]
func (h *RequestHandler) UploadDocument(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form
	if err := r.ParseMultipartForm(20 << 20); err != nil { // 20MB limit
		respondError(w, http.StatusBadRequest, "INVALID_FORM", "Failed to parse form data")
		return
	}

	// TODO: Upload file to MinIO/S3
	// TODO: Call request gRPC service

	response := dto.DocumentResponse{
		ID:       "doc-placeholder-id",
		Type:     r.FormValue("type"),
		Name:     "uploaded-file.pdf",
		MimeType: "application/pdf",
		Size:     1024000,
	}

	respondJSON(w, http.StatusCreated, response)
}

// ListDocuments godoc
// @Summary Список документов
// @Description Возвращает список документов заявки
// @Tags requests
// @Accept json
// @Produce json
// @Param id path string true "ID заявки" format(uuid)
// @Success 200 {array} dto.DocumentResponse "Список документов"
// @Failure 401 {object} dto.ErrorResponse "Не авторизован"
// @Failure 404 {object} dto.ErrorResponse "Заявка не найдена"
// @Security BearerAuth
// @Router /requests/{id}/documents [get]
func (h *RequestHandler) ListDocuments(w http.ResponseWriter, r *http.Request) {
	// TODO: Call request gRPC service

	documents := []dto.DocumentResponse{}
	respondJSON(w, http.StatusOK, documents)
}

// DeleteDocument godoc
// @Summary Удалить документ
// @Description Удаляет документ из заявки
// @Tags requests
// @Accept json
// @Produce json
// @Param id path string true "ID заявки" format(uuid)
// @Param doc_id path string true "ID документа" format(uuid)
// @Success 200 {object} dto.SuccessResponse "Документ удален"
// @Failure 401 {object} dto.ErrorResponse "Не авторизован"
// @Failure 404 {object} dto.ErrorResponse "Документ не найден"
// @Security BearerAuth
// @Router /requests/{id}/documents/{doc_id} [delete]
func (h *RequestHandler) DeleteDocument(w http.ResponseWriter, r *http.Request) {
	// TODO: Call request gRPC service

	respondJSON(w, http.StatusOK, dto.SuccessResponse{
		Success: true,
		Message: "Document deleted successfully",
	})
}

// =============================================================================
// Pricing Endpoints
// =============================================================================

// GetPricing godoc
// @Summary Цены на услуги
// @Description Возвращает информацию о ценах на услуги экспертов
// @Tags requests
// @Accept json
// @Produce json
// @Success 200 {object} dto.PricingResponse "Информация о ценах"
// @Security BearerAuth
// @Router /requests/pricing [get]
func (h *RequestHandler) GetPricing(w http.ResponseWriter, r *http.Request) {
	response := dto.PricingResponse{
		Categories: []dto.ServiceCategoryPriceResponse{
			{
				Code:          "consultation",
				Name:          "Консультация",
				Description:   "Консультация специалиста БТИ по вопросам перепланировки",
				BasePrice:     2000,
				EstimatedDays: "3-5 рабочих дней",
				Includes: []string{
					"Анализ документации",
					"Устная консультация (до 1 часа)",
					"Письменное заключение",
				},
			},
			{
				Code:          "documentation",
				Name:          "Подготовка документации",
				Description:   "Полный пакет документов для согласования перепланировки",
				BasePrice:     15000,
				EstimatedDays: "10-14 рабочих дней",
				Includes: []string{
					"Техническое заключение",
					"Проект перепланировки",
					"Подготовка заявления",
					"Сопровождение согласования",
				},
			},
			{
				Code:          "expert_visit",
				Name:          "Выезд эксперта",
				Description:   "Осмотр объекта экспертом БТИ на месте",
				BasePrice:     5000,
				EstimatedDays: "1-3 рабочих дня",
				Includes: []string{
					"Выезд на объект",
					"Осмотр и обмеры",
					"Устное заключение",
					"Фотофиксация",
				},
			},
			{
				Code:          "full_package",
				Name:          "Полный пакет",
				Description:   "Комплексное сопровождение перепланировки под ключ",
				BasePrice:     30000,
				EstimatedDays: "30-45 рабочих дней",
				Includes: []string{
					"Все услуги консультации",
					"Выезд эксперта",
					"Полный пакет документов",
					"Согласование в инстанциях",
					"Получение акта о завершении",
				},
			},
		},
	}

	respondJSON(w, http.StatusOK, response)
}

