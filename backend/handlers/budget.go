package handlers

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/logging"
	"github.com/liyali/liyali-gateway/middleware"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/services"
	"github.com/liyali/liyali-gateway/types"
	"github.com/liyali/liyali-gateway/utils"
	"gorm.io/datatypes"
)

// GetBudgets retrieves all budgets with pagination and filtering
func GetBudgets(c *fiber.Ctx) error {
	logger := logging.FromContext(c)
	logger.Info("get_budgets_request")

	// Get organization context from tenant middleware
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Organization context required",
			"error":   err.Error(),
		})
	}

	db := config.DB

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	status := c.Query("status")
	department := c.Query("department")
	fiscalYear := c.Query("fiscalYear")

	// Add query parameters to context
	logging.AddFieldsToRequest(c, map[string]interface{}{
		"page":            page,
		"limit":           limit,
		"status":          status,
		"department":      department,
		"fiscal_year":     fiscalYear,
		"operation":       "get_budgets",
		"organization_id": tenant.OrganizationID,
	})

	// Start with organization filter - CRITICAL SECURITY FIX
	query := db.Where("organization_id = ?", tenant.OrganizationID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if department != "" {
		query = query.Where("department = ?", department)
	}
	if fiscalYear != "" {
		query = query.Where("fiscal_year = ?", fiscalYear)
	}

	logger.Debug("counting_budgets")

	var total int64
	if err := query.Model(&models.Budget{}).Count(&total).Error; err != nil {
		logging.LogError(c, err, "failed_to_count_budgets")
		return utils.SendInternalError(c, "Failed to count budgets", err)
	}

	logger.Debug("fetching_budgets")

	var budgets []models.Budget
	offset := (page - 1) * limit
	if err := query.
		Offset(offset).
		Limit(limit).
		Preload("Owner").
		Order("created_at DESC").
		Find(&budgets).Error; err != nil {
		logging.LogError(c, err, "failed_to_fetch_budgets")
		return utils.SendInternalError(c, "Failed to fetch budgets", err)
	}

	responses := make([]types.BudgetResponse, 0, len(budgets))
	for _, budget := range budgets {
		responses = append(responses, modelToBudgetResponse(budget))
	}

	logger.WithFields(map[string]interface{}{
		"budget_count": len(budgets),
		"total_count":  total,
	}).Info("budgets_retrieved_successfully")

	return utils.SendPaginatedSuccess(c, responses, "Budgets retrieved successfully", page, limit, total)
}

// CreateBudget creates a new budget
func CreateBudget(c *fiber.Ctx) error {
	logger := logging.FromContext(c)
	logger.Info("create_budget_request")

	// Get organization context from tenant middleware
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Organization context required",
			"error":   err.Error(),
		})
	}

	var req types.CreateBudgetRequest

	if err := c.BodyParser(&req); err != nil {
		logging.LogError(c, err, "failed_to_parse_create_budget_request")
		return utils.SendBadRequestError(c, "Invalid request body")
	}

	// Add budget details to context
	logging.AddFieldsToRequest(c, map[string]interface{}{
		"budget_code":      req.BudgetCode,
		"department":       req.Department,
		"fiscal_year":      req.FiscalYear,
		"total_budget":     req.TotalBudget,
		"allocated_amount": req.AllocatedAmount,
		"operation":        "create_budget",
		"organization_id":  tenant.OrganizationID,
	})

	// Auto-generate budget code if not provided
	if req.BudgetCode == "" {
		year := time.Now().Year()
		randomID := uuid.New().String()[:8] // Take first 8 characters
		req.BudgetCode = fmt.Sprintf("BG-%d-%s", year, strings.ToUpper(randomID))
		logging.AddFieldToRequest(c, "generated_budget_code", req.BudgetCode)
	}
	if req.TotalBudget <= 0 {
		logging.LogWarn(c, "invalid_total_budget", map[string]interface{}{
			"total_budget": req.TotalBudget,
		})
		return utils.SendBadRequestError(c, "Total budget must be greater than 0")
	}
	if req.AllocatedAmount < 0 {
		logging.LogWarn(c, "invalid_allocated_amount", map[string]interface{}{
			"allocated_amount": req.AllocatedAmount,
		})
		return utils.SendBadRequestError(c, "Allocated amount cannot be negative")
	}

	userID := c.Locals("userID").(string)
	if userID == "" {
		logging.LogWarn(c, "user_id_missing_from_context")
		return utils.SendUnauthorizedError(c, "User ID not found in token")
	}

	// Add user context
	logging.AddFieldToRequest(c, "user_id", userID)

	logger.Debug("validating_user")

	var user models.User
	if err := config.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		logging.LogError(c, err, "user_not_found_for_budget_creation")
		return utils.SendUnauthorizedError(c, "User not found")
	}

	remainingAmount := req.TotalBudget - req.AllocatedAmount
	budgetID := uuid.New().String()

	// Add calculated values to context
	logging.AddFieldsToRequest(c, map[string]interface{}{
		"remaining_amount": remainingAmount,
		"budget_id":        budgetID,
	})

	budget := models.Budget{
		ID:              budgetID,
		OrganizationID:  tenant.OrganizationID, // SECURITY FIX: Set organization ID
		OwnerID:         userID,
		BudgetCode:      req.BudgetCode,
		Name:            req.Name,        // Add name field
		Description:     req.Description, // Add description field
		Department:      req.Department,
		DepartmentID:    req.DepartmentID, // Add department ID field
		Status:          "draft",
		FiscalYear:      req.FiscalYear,
		TotalBudget:     req.TotalBudget,
		AllocatedAmount: req.AllocatedAmount,
		RemainingAmount: remainingAmount,
		Currency:        req.Currency, // Add currency field
		ApprovalStage:   0,
		CreatedBy:       userID, // Add created by field
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	emptyHistory := []types.ApprovalRecord{}
	budget.ApprovalHistory = datatypes.NewJSONType(emptyHistory)

	logger.Debug("creating_budget_in_database")

	if err := config.DB.Create(&budget).Error; err != nil {
		logging.LogError(c, err, "failed_to_create_budget_in_database")
		return utils.SendInternalError(c, "Failed to create budget", err)
	}

	config.DB.Preload("Owner").First(&budget)

	logger.Info("budget_created_successfully")
	return utils.SendCreatedSuccess(c, modelToBudgetResponse(budget), "Budget created successfully")
}

// GetBudget retrieves a single budget by ID
func GetBudget(c *fiber.Ctx) error {
	logger := logging.FromContext(c)
	logger.Info("get_budget_request")

	// Get organization context from tenant middleware
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Organization context required",
			"error":   err.Error(),
		})
	}

	id := c.Params("id")
	if id == "" {
		logging.LogWarn(c, "budget_id_missing")
		return utils.SendBadRequestError(c, "Budget ID is required")
	}

	// Add budget ID to context
	logging.AddFieldsToRequest(c, map[string]interface{}{
		"budget_id":       id,
		"operation":       "get_budget",
		"organization_id": tenant.OrganizationID,
	})

	logger.Debug("fetching_budget_by_id")

	var budget models.Budget
	// SECURITY FIX: Filter by organization ID
	if err := config.DB.
		Preload("Owner").
		Where("id = ? AND organization_id = ?", id, tenant.OrganizationID).
		First(&budget).Error; err != nil {
		logging.LogError(c, err, "budget_not_found")
		return utils.SendNotFoundError(c, "Budget")
	}

	logger.Info("budget_retrieved_successfully")
	return utils.SendSimpleSuccess(c, modelToBudgetResponse(budget), "Budget retrieved successfully")
}

// UpdateBudget updates an existing budget
func UpdateBudget(c *fiber.Ctx) error {
	logger := logging.FromContext(c)
	logger.Info("update_budget_request")

	// Get organization context from tenant middleware
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Organization context required",
			"error":   err.Error(),
		})
	}

	id := c.Params("id")
	if id == "" {
		logging.LogWarn(c, "budget_id_missing_for_update")
		return utils.SendBadRequestError(c, "Budget ID is required")
	}

	var req types.UpdateBudgetRequest
	if err := c.BodyParser(&req); err != nil {
		logging.LogError(c, err, "failed_to_parse_update_budget_request")
		return utils.SendBadRequestError(c, "Invalid request body")
	}

	// Add context
	logging.AddFieldsToRequest(c, map[string]interface{}{
		"budget_id":        id,
		"operation":        "update_budget",
		"new_department":   req.Department,
		"new_total_budget": req.TotalBudget,
		"new_allocated":    req.AllocatedAmount,
		"organization_id":  tenant.OrganizationID,
	})

	logger.Debug("fetching_budget_for_update")

	var budget models.Budget
	// SECURITY FIX: Filter by organization ID
	if err := config.DB.Where("id = ? AND organization_id = ?", id, tenant.OrganizationID).First(&budget).Error; err != nil {
		logging.LogError(c, err, "budget_not_found_for_update")
		return utils.SendNotFoundError(c, "Budget")
	}

	// Add current budget status to context
	logging.AddFieldToRequest(c, "current_status", budget.Status)

	if budget.Status != "draft" && budget.Status != "pending" {
		logging.LogWarn(c, "budget_update_not_allowed", map[string]interface{}{
			"current_status": budget.Status,
		})
		return utils.SendForbiddenError(c, fmt.Sprintf("Cannot update budget in %s status", budget.Status))
	}

	logger.Debug("updating_budget_fields")

	if req.Department != "" {
		budget.Department = req.Department
	}
	if req.TotalBudget > 0 {
		budget.TotalBudget = req.TotalBudget
	}
	if req.AllocatedAmount >= 0 {
		budget.AllocatedAmount = req.AllocatedAmount
	}
	if req.Name != "" {
		budget.Name = req.Name
	}
	if req.Description != "" {
		budget.Description = req.Description
	}
	if req.Currency != "" {
		budget.Currency = req.Currency
	}
	// Update items if provided
	if req.Items != nil {
		itemsJSON, err := json.Marshal(req.Items)
		if err != nil {
			logging.LogError(c, err, "failed_to_marshal_budget_items")
			return utils.SendBadRequestError(c, "Invalid items format")
		}
		budget.Items = itemsJSON
	}

	budget.RemainingAmount = budget.TotalBudget - budget.AllocatedAmount
	budget.UpdatedAt = time.Now()

	// Add updated values to context
	logging.AddFieldsToRequest(c, map[string]interface{}{
		"updated_remaining_amount": budget.RemainingAmount,
	})

	if err := config.DB.Save(&budget).Error; err != nil {
		logging.LogError(c, err, "failed_to_save_updated_budget")
		return utils.SendInternalError(c, "Failed to update budget", err)
	}

	config.DB.Preload("Owner").First(&budget)

	logger.Info("budget_updated_successfully")
	return utils.SendSimpleSuccess(c, modelToBudgetResponse(budget), "Budget updated successfully")
}

// DeleteBudget deletes a budget
func DeleteBudget(c *fiber.Ctx) error {
	logger := logging.FromContext(c)
	logger.Info("delete_budget_request")

	// Get organization context from tenant middleware
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Organization context required",
			"error":   err.Error(),
		})
	}

	id := c.Params("id")
	if id == "" {
		logging.LogWarn(c, "budget_id_missing_for_deletion")
		return utils.SendBadRequestError(c, "Budget ID is required")
	}

	// Add context
	logging.AddFieldsToRequest(c, map[string]interface{}{
		"budget_id":       id,
		"operation":       "delete_budget",
		"organization_id": tenant.OrganizationID,
	})

	logger.Debug("fetching_budget_for_deletion")

	var budget models.Budget
	// SECURITY FIX: Filter by organization ID
	if err := config.DB.Where("id = ? AND organization_id = ?", id, tenant.OrganizationID).First(&budget).Error; err != nil {
		logging.LogError(c, err, "budget_not_found_for_deletion")
		return utils.SendNotFoundError(c, "Budget")
	}

	// Add budget status to context
	logging.AddFieldToRequest(c, "budget_status", budget.Status)

	if budget.Status != "draft" {
		logging.LogWarn(c, "budget_deletion_not_allowed", map[string]interface{}{
			"current_status": budget.Status,
		})
		return utils.SendForbiddenError(c, "Only draft budgets can be deleted")
	}

	logger.Debug("deleting_budget_from_database")

	if err := config.DB.Delete(&budget).Error; err != nil {
		logging.LogError(c, err, "failed_to_delete_budget_from_database")
		return utils.SendInternalError(c, "Failed to delete budget", err)
	}

	logger.Info("budget_deleted_successfully")
	return utils.SendSimpleSuccess(c, nil, "Budget deleted successfully")
}

// SubmitBudget submits a budget for approval workflow
func SubmitBudget(c *fiber.Ctx) error {
	logger := logging.FromContext(c)
	logger.Info("submit_budget_request")

	// Get organization context from tenant middleware
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Organization context required",
			"error":   err.Error(),
		})
	}

	id := c.Params("id")
	if id == "" {
		logging.LogWarn(c, "budget_id_missing_for_submission")
		return utils.SendBadRequestError(c, "Budget ID is required")
	}

	// Add context
	logging.AddFieldsToRequest(c, map[string]interface{}{
		"budget_id":       id,
		"operation":       "submit_budget",
		"organization_id": tenant.OrganizationID,
	})

	logger.Debug("fetching_budget_for_submission")

	var budget models.Budget
	// SECURITY FIX: Filter by organization ID
	if err := config.DB.Where("id = ? AND organization_id = ?", id, tenant.OrganizationID).First(&budget).Error; err != nil {
		logging.LogError(c, err, "budget_not_found_for_submission")
		return utils.SendNotFoundError(c, "Budget")
	}

	// Add budget status to context
	logging.AddFieldToRequest(c, "current_status", budget.Status)

	if budget.Status != "draft" {
		logging.LogWarn(c, "budget_submission_not_allowed", map[string]interface{}{
			"current_status": budget.Status,
		})
		return utils.SendBadRequestError(c, fmt.Sprintf("Cannot submit budget in %s status", budget.Status))
	}

	userID := c.Locals("userID").(string)
	organizationID := c.Locals("organizationID").(string)

	// Add user context
	logging.AddFieldsToRequest(c, map[string]interface{}{
		"user_id":         userID,
		"organization_id": organizationID,
	})

	// Get workflow execution service
	workflowExecutionService := c.Locals("workflowExecutionService").(*services.WorkflowExecutionService)

	logger.Debug("assigning_workflow_to_budget")

	// Assign workflow to the budget
	_, err = workflowExecutionService.AssignWorkflowToDocument(c.Context(), organizationID, budget.ID, "budget", userID)
	if err != nil {
		logging.LogError(c, err, "failed_to_assign_workflow_to_budget")
		return utils.SendInternalError(c, "Failed to assign workflow to budget", err)
	}

	// Update budget status to pending
	budget.Status = "pending"
	budget.UpdatedAt = time.Now()

	// Add action to history
	var actionHistory []types.ActionHistoryEntry
	actionHistory = budget.ActionHistory.Data()

	actionRecord := types.ActionHistoryEntry{
		Action:          "SUBMIT",
		PerformedBy:     userID,
		PerformedByName: "", // Will be populated by the database trigger or service
		Timestamp:       time.Now(),
		Comments:        "Budget submitted for approval",
	}
	actionHistory = append(actionHistory, actionRecord)
	budget.ActionHistory = datatypes.NewJSONType(actionHistory)

	// Add updated status to context
	logging.AddFieldToRequest(c, "new_status", budget.Status)

	logger.Debug("saving_submitted_budget")

	if err := config.DB.Save(&budget).Error; err != nil {
		logging.LogError(c, err, "failed_to_save_submitted_budget")
		return utils.SendInternalError(c, "Failed to submit budget", err)
	}

	config.DB.Preload("Owner").First(&budget)

	logger.Info("budget_submitted_successfully")
	return utils.SendSimpleSuccess(c, modelToBudgetResponse(budget), "Budget submitted for approval successfully")
}

// Helper function to convert model to response
func modelToBudgetResponse(budget models.Budget) types.BudgetResponse {
	var approvalHistory []types.ApprovalRecord
	approvalHistory = budget.ApprovalHistory.Data()

	var items []interface{}
	if budget.Items != nil && len(budget.Items) > 0 {
		if err := json.Unmarshal(budget.Items, &items); err != nil {
			// If unmarshal fails, return empty array
			items = []interface{}{}
		}
	}

	ownerName := ""
	if budget.Owner != nil {
		ownerName = budget.Owner.Name
	}

	return types.BudgetResponse{
		ID:              budget.ID,
		BudgetCode:      budget.BudgetCode,
		OwnerID:         budget.OwnerID,
		OwnerName:       ownerName,
		Department:      budget.Department,
		DepartmentID:    budget.DepartmentID,
		Status:          budget.Status,
		FiscalYear:      budget.FiscalYear,
		TotalBudget:     budget.TotalBudget,
		AllocatedAmount: budget.AllocatedAmount,
		RemainingAmount: budget.RemainingAmount,
		ApprovalStage:   budget.ApprovalStage,
		ApprovalHistory: approvalHistory,
		Name:            budget.Name,
		Description:     budget.Description,
		Currency:        budget.Currency,
		CreatedBy:       budget.CreatedBy,
		Items:           items,
		CreatedAt:       budget.CreatedAt,
		UpdatedAt:       budget.UpdatedAt,
	}
}
