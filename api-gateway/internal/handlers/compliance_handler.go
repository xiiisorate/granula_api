// =============================================================================
// Compliance Handler - HTTP handlers for compliance checking.
// =============================================================================
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/xiiisorate/granula_api/api-gateway/internal/dto"
)

// =============================================================================
// ComplianceHandler handles compliance-related HTTP requests.
// =============================================================================

// ComplianceHandler provides HTTP handlers for compliance operations.
type ComplianceHandler struct {
	// complianceClient pb.ComplianceServiceClient
}

// NewComplianceHandler creates a new ComplianceHandler.
func NewComplianceHandler() *ComplianceHandler {
	return &ComplianceHandler{}
}

// CheckCompliance godoc
// @Summary Проверить соответствие нормам
// @Description Запускает проверку сцены на соответствие СНиП и ЖК РФ
// @Tags compliance
// @Accept json
// @Produce json
// @Param request body dto.CheckComplianceRequest true "Параметры проверки"
// @Success 200 {object} dto.ComplianceCheckResponse "Результат проверки"
// @Failure 400 {object} dto.ErrorResponse "Ошибка валидации"
// @Failure 401 {object} dto.ErrorResponse "Не авторизован"
// @Failure 404 {object} dto.ErrorResponse "Сцена не найдена"
// @Security BearerAuth
// @Router /compliance/check [post]
func (h *ComplianceHandler) CheckCompliance(w http.ResponseWriter, r *http.Request) {
	var req dto.CheckComplianceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request body")
		return
	}

	// TODO: Call compliance gRPC service

	response := dto.ComplianceCheckResponse{
		CheckID:        "check-placeholder-id",
		SceneID:        req.SceneID,
		Status:         "warning",
		Score:          78,
		ViolationCount: 2,
		WarningCount:   3,
		Violations: []dto.ViolationResponse{
			{
				ID:          "v-1",
				RuleID:      "SNIP-MIN-AREA-01",
				RuleName:    "Минимальная площадь жилой комнаты",
				Category:    "minimum_area",
				Severity:    "error",
				Description: "Площадь спальни (7.5 кв.м) меньше минимально допустимой (8 кв.м)",
				Recommendations: []string{
					"Увеличить площадь комнаты минимум на 0.5 кв.м",
					"Рассмотреть объединение с соседней комнатой",
				},
				LegalReference: "СНиП 31-01-2003, п. 5.3",
			},
		},
		ProcessingTime: 320,
	}

	respondJSON(w, http.StatusOK, response)
}

// GetCheckResult godoc
// @Summary Получить результат проверки
// @Description Возвращает детальный результат проверки по ID
// @Tags compliance
// @Accept json
// @Produce json
// @Param check_id path string true "ID проверки" format(uuid)
// @Success 200 {object} dto.ComplianceCheckResponse "Результат проверки"
// @Failure 401 {object} dto.ErrorResponse "Не авторизован"
// @Failure 404 {object} dto.ErrorResponse "Проверка не найдена"
// @Security BearerAuth
// @Router /compliance/checks/{check_id} [get]
func (h *ComplianceHandler) GetCheckResult(w http.ResponseWriter, r *http.Request) {
	// TODO: Call compliance gRPC service

	response := dto.ComplianceCheckResponse{
		CheckID: "check-id",
		Status:  "passed",
		Score:   95,
	}

	respondJSON(w, http.StatusOK, response)
}

// GetCheckHistory godoc
// @Summary История проверок
// @Description Возвращает историю проверок для сцены
// @Tags compliance
// @Accept json
// @Produce json
// @Param scene_id query string true "ID сцены" format(uuid)
// @Param page query int false "Номер страницы" default(1)
// @Param page_size query int false "Размер страницы" default(20)
// @Success 200 {object} dto.ComplianceHistoryResponse "История проверок"
// @Failure 401 {object} dto.ErrorResponse "Не авторизован"
// @Failure 404 {object} dto.ErrorResponse "Сцена не найдена"
// @Security BearerAuth
// @Router /compliance/history [get]
func (h *ComplianceHandler) GetCheckHistory(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// TODO: Call compliance gRPC service

	response := dto.ComplianceHistoryResponse{
		Checks: []dto.ComplianceCheckSummaryResponse{},
		Pagination: dto.PaginationResponse{
			Page:     page,
			PageSize: pageSize,
		},
	}

	respondJSON(w, http.StatusOK, response)
}

// =============================================================================
// Rules Endpoints
// =============================================================================

// ListRules godoc
// @Summary Список правил
// @Description Возвращает список всех доступных правил проверки
// @Tags compliance
// @Accept json
// @Produce json
// @Param category query string false "Фильтр по категории"
// @Param operation_type query string false "Фильтр по типу операции" Enums(construction, renovation, redevelopment)
// @Success 200 {object} dto.RulesListResponse "Список правил"
// @Failure 401 {object} dto.ErrorResponse "Не авторизован"
// @Security BearerAuth
// @Router /compliance/rules [get]
func (h *ComplianceHandler) ListRules(w http.ResponseWriter, r *http.Request) {
	// TODO: Call compliance gRPC service

	response := dto.RulesListResponse{
		Rules: []dto.ComplianceRuleResponse{
			{
				ID:              "SNIP-MIN-AREA-01",
				Name:            "Минимальная площадь жилой комнаты",
				Category:        "minimum_area",
				Description:     "Площадь жилой комнаты должна быть не менее 8 кв.м",
				DefaultSeverity: "error",
				LegalReference:  "СНиП 31-01-2003",
				ApplicableOperations: []string{"construction", "redevelopment"},
				IsActive:        true,
			},
			{
				ID:              "SNIP-VENT-01",
				Name:            "Требования к вентиляции",
				Category:        "ventilation",
				Description:     "Каждое жилое помещение должно иметь систему вентиляции",
				DefaultSeverity: "error",
				LegalReference:  "СНиП 41-01-2003",
				ApplicableOperations: []string{"construction", "renovation", "redevelopment"},
				IsActive:        true,
			},
		},
		Categories: []dto.RuleCategoryResponse{
			{Code: "minimum_area", Name: "Минимальные площади", RuleCount: 8},
			{Code: "ventilation", Name: "Вентиляция", RuleCount: 5},
			{Code: "load_bearing", Name: "Несущие конструкции", RuleCount: 3},
			{Code: "fire_safety", Name: "Пожарная безопасность", RuleCount: 12},
		},
		TotalRules: 45,
	}

	respondJSON(w, http.StatusOK, response)
}

// GetRule godoc
// @Summary Получить правило
// @Description Возвращает детальную информацию о правиле
// @Tags compliance
// @Accept json
// @Produce json
// @Param rule_id path string true "ID правила"
// @Success 200 {object} dto.ComplianceRuleResponse "Данные правила"
// @Failure 401 {object} dto.ErrorResponse "Не авторизован"
// @Failure 404 {object} dto.ErrorResponse "Правило не найдено"
// @Security BearerAuth
// @Router /compliance/rules/{rule_id} [get]
func (h *ComplianceHandler) GetRule(w http.ResponseWriter, r *http.Request) {
	// TODO: Call compliance gRPC service

	response := dto.ComplianceRuleResponse{
		ID:              "SNIP-MIN-AREA-01",
		Name:            "Минимальная площадь жилой комнаты",
		Category:        "minimum_area",
		Description:     "Площадь жилой комнаты должна быть не менее 8 кв.м согласно СНиП 31-01-2003",
		DefaultSeverity: "error",
		LegalReference:  "СНиП 31-01-2003, п. 5.3",
		ApplicableOperations: []string{"construction", "redevelopment"},
		IsActive:        true,
	}

	respondJSON(w, http.StatusOK, response)
}

// GetCategories godoc
// @Summary Список категорий правил
// @Description Возвращает список всех категорий правил проверки
// @Tags compliance
// @Accept json
// @Produce json
// @Success 200 {array} dto.RuleCategoryResponse "Список категорий"
// @Failure 401 {object} dto.ErrorResponse "Не авторизован"
// @Security BearerAuth
// @Router /compliance/categories [get]
func (h *ComplianceHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	// TODO: Call compliance gRPC service

	categories := []dto.RuleCategoryResponse{
		{
			Code:        "minimum_area",
			Name:        "Минимальные площади",
			Description: "Правила минимальных площадей помещений согласно СНиП",
			RuleCount:   8,
		},
		{
			Code:        "ventilation",
			Name:        "Вентиляция",
			Description: "Требования к системам вентиляции",
			RuleCount:   5,
		},
		{
			Code:        "load_bearing",
			Name:        "Несущие конструкции",
			Description: "Правила изменения несущих конструкций",
			RuleCount:   3,
		},
		{
			Code:        "fire_safety",
			Name:        "Пожарная безопасность",
			Description: "Требования пожарной безопасности",
			RuleCount:   12,
		},
		{
			Code:        "wet_zones",
			Name:        "Мокрые зоны",
			Description: "Правила размещения санузлов и кухонь",
			RuleCount:   6,
		},
		{
			Code:        "accessibility",
			Name:        "Доступность",
			Description: "Требования доступности для маломобильных групп",
			RuleCount:   4,
		},
	}

	respondJSON(w, http.StatusOK, categories)
}

