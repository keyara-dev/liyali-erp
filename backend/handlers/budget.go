package handlers

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/types"
	"github.com/liyali/liyali-gateway/utils"
	"gorm.io/datatypes"
)

// GetBudgets retrieves all budgets with pagination and filtering
func GetBudgets(c *fiber.Ctx) error {
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

	query := db
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if department != "" {
		query = query.Where("department = ?", department)
	}
	if fiscalYear != "" {
		query = query.Where("fiscal_year = ?", fiscalYear)
	}

	var total int64
	if err := query.Model(&models.Budget{}).Count(&total).Error; err != nil {
		return utils.SendInternalError(c, "Failed to count budgets", err)
	}

	var budgets []models.Budget
	offset := (page - 1) * limit
	if err := query.
		Offset(offset).
		Limit(limit).
		Preload("Owner").
		Order("created_at DESC").
		Find(&budgets).Error; err != nil {
		return utils.SendInternalError(c, "Failed to fetch budgets", err)
	}

	responses := make([]types.BudgetResponse, 0, len(budgets))
	for _, budget := range budgets {
		responses = append(responses, modelToBudgetResponse(budget))
	}

	return utils.SendPaginatedSuccess(c, responses, "Budgets retrieved successfully", page, limit, total)
}

// CreateBudget creates a new budget
func CreateBudget(c *fiber.Ctx) error {
	var req types.CreateBudgetRequest

	if err := c.BodyParser(&req); err != nil {
		return utils.SendBadRequestError(c, "Invalid request body")
	}

	if req.BudgetCode == "" {
		return utils.SendBadRequestError(c, "Budget code is required")
	}
	if req.TotalBudget <= 0 {
		return utils.SendBadRequestError(c, "Total budget must be greater than 0")
	}
	if req.AllocatedAmount < 0 {
		return utils.SendBadRequestError(c, "Allocated amount cannot be negative")
	}

	userID := c.Locals("user_id").(string)
	if userID == "" {
		return utils.SendUnauthorizedError(c, "User ID not found in token")
	}

	var user models.User
	if err := config.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		return utils.SendUnauthorizedError(c, "User not found")
	}

	remainingAmount := req.TotalBudget - req.AllocatedAmount

	budget := models.Budget{
		ID:              uuid.New().String(),
		OwnerID:         userID,
		BudgetCode:      req.BudgetCode,
		Department:      req.Department,
		Status:          "draft",
		FiscalYear:      req.FiscalYear,
		TotalBudget:     req.TotalBudget,
		AllocatedAmount: req.AllocatedAmount,
		RemainingAmount: remainingAmount,
		ApprovalStage:   0,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	emptyHistory := []types.ApprovalRecord{}
	budget.ApprovalHistory = datatypes.NewJSONType(emptyHistory)

	if err := config.DB.Create(&budget).Error; err != nil {
		return utils.SendInternalError(c, "Failed to create budget", err)
	}

	config.DB.Preload("Owner").First(&budget)

	return utils.SendCreatedSuccess(c, modelToBudgetResponse(budget), "Budget created successfully")
}

// GetBudget retrieves a single budget by ID
func GetBudget(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return utils.SendBadRequestError(c, "Budget ID is required")
	}

	var budget models.Budget
	if err := config.DB.
		Preload("Owner").
		Where("id = ?", id).
		First(&budget).Error; err != nil {
		return utils.SendNotFoundError(c, "Budget")
	}

	return utils.SendSimpleSuccess(c, modelToBudgetResponse(budget), "Budget retrieved successfully")
}

// UpdateBudget updates an existing budget
func UpdateBudget(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return utils.SendBadRequestError(c, "Budget ID is required")
	}

	var req types.UpdateBudgetRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendBadRequestError(c, "Invalid request body")
	}

	var budget models.Budget
	if err := config.DB.Where("id = ?", id).First(&budget).Error; err != nil {
		return utils.SendNotFoundError(c, "Budget")
	}

	if budget.Status != "draft" && budget.Status != "pending" {
		return utils.SendForbiddenError(c, fmt.Sprintf("Cannot update budget in %s status", budget.Status))
	}

	if req.Department != "" {
		budget.Department = req.Department
	}
	if req.TotalBudget > 0 {
		budget.TotalBudget = req.TotalBudget
	}
	if req.AllocatedAmount >= 0 {
		budget.AllocatedAmount = req.AllocatedAmount
	}

	budget.RemainingAmount = budget.TotalBudget - budget.AllocatedAmount
	budget.UpdatedAt = time.Now()

	if err := config.DB.Save(&budget).Error; err != nil {
		return utils.SendInternalError(c, "Failed to update budget", err)
	}

	config.DB.Preload("Owner").First(&budget)

	return utils.SendSimpleSuccess(c, modelToBudgetResponse(budget), "Budget updated successfully")
}

// DeleteBudget deletes a budget
func DeleteBudget(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return utils.SendBadRequestError(c, "Budget ID is required")
	}

	var budget models.Budget
	if err := config.DB.Where("id = ?", id).First(&budget).Error; err != nil {
		return utils.SendNotFoundError(c, "Budget")
	}

	if budget.Status != "draft" {
		return utils.SendForbiddenError(c, "Only draft budgets can be deleted")
	}

	if err := config.DB.Delete(&budget).Error; err != nil {
		return utils.SendInternalError(c, "Failed to delete budget", err)
	}

	return utils.SendSimpleSuccess(c, nil, "Budget deleted successfully")
}

// ApproveBudget approves a budget
func ApproveBudget(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return utils.SendBadRequestError(c, "Budget ID is required")
	}

	var req types.ApproveDocumentRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendBadRequestError(c, "Invalid request body")
	}

	if req.Signature == "" {
		return utils.SendBadRequestError(c, "Signature is required")
	}

	var budget models.Budget
	if err := config.DB.Where("id = ?", id).First(&budget).Error; err != nil {
		return utils.SendNotFoundError(c, "Budget")
	}

	approverID := c.Locals("user_id").(string)
	var approver models.User
	if err := config.DB.Where("id = ?", approverID).First(&approver).Error; err != nil {
		return utils.SendUnauthorizedError(c, "Approver not found")
	}

	var approvalHistory []types.ApprovalRecord
	approvalHistory = budget.ApprovalHistory.Data()

	approvalRecord := types.ApprovalRecord{
		ApproverID:   approverID,
		ApproverName: approver.Name,
		Status:       "approved",
		Comments:     req.Comments,
		Signature:    req.Signature,
		ApprovedAt:   time.Now(),
	}
	approvalHistory = append(approvalHistory, approvalRecord)

	budget.Status = "approved"
	budget.ApprovalStage++
	budget.ApprovalHistory = datatypes.NewJSONType(approvalHistory)
	budget.UpdatedAt = time.Now()

	if err := config.DB.Save(&budget).Error; err != nil {
		return utils.SendInternalError(c, "Failed to approve budget", err)
	}

	config.DB.Preload("Owner").First(&budget)

	return utils.SendSimpleSuccess(c, modelToBudgetResponse(budget), "Budget approved successfully")
}

// RejectBudget rejects a budget
func RejectBudget(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return utils.SendBadRequestError(c, "Budget ID is required")
	}

	var req types.RejectDocumentRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendBadRequestError(c, "Invalid request body")
	}

	if req.Remarks == "" || len(req.Remarks) < 10 {
		return utils.SendBadRequestError(c, "Remarks must be at least 10 characters")
	}
	if req.Signature == "" {
		return utils.SendBadRequestError(c, "Signature is required")
	}

	var budget models.Budget
	if err := config.DB.Where("id = ?", id).First(&budget).Error; err != nil {
		return utils.SendNotFoundError(c, "Budget")
	}

	approverID := c.Locals("user_id").(string)
	var approver models.User
	if err := config.DB.Where("id = ?", approverID).First(&approver).Error; err != nil {
		return utils.SendUnauthorizedError(c, "Approver not found")
	}

	var approvalHistory []types.ApprovalRecord
	approvalHistory = budget.ApprovalHistory.Data()

	rejectionRecord := types.ApprovalRecord{
		ApproverID:   approverID,
		ApproverName: approver.Name,
		Status:       "rejected",
		Comments:     req.Remarks,
		Signature:    req.Signature,
		ApprovedAt:   time.Now(),
	}
	approvalHistory = append(approvalHistory, rejectionRecord)

	budget.Status = "rejected"
	budget.ApprovalHistory = datatypes.NewJSONType(approvalHistory)
	budget.UpdatedAt = time.Now()

	if err := config.DB.Save(&budget).Error; err != nil {
		return utils.SendInternalError(c, "Failed to reject budget", err)
	}

	config.DB.Preload("Owner").First(&budget)

	return utils.SendSimpleSuccess(c, modelToBudgetResponse(budget), "Budget rejected successfully")
}

// Helper function to convert model to response
func modelToBudgetResponse(budget models.Budget) types.BudgetResponse {
	var approvalHistory []types.ApprovalRecord
	approvalHistory = budget.ApprovalHistory.Data()

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
		Status:          budget.Status,
		FiscalYear:      budget.FiscalYear,
		TotalBudget:     budget.TotalBudget,
		AllocatedAmount: budget.AllocatedAmount,
		RemainingAmount: budget.RemainingAmount,
		ApprovalStage:   budget.ApprovalStage,
		ApprovalHistory: approvalHistory,
		CreatedAt:       budget.CreatedAt,
		UpdatedAt:       budget.UpdatedAt,
	}
}
