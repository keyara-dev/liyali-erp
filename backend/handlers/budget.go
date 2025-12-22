package handlers

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/types"
)

// GetBudgets retrieves all budgets with pagination and filtering
func GetBudgets(c fiber.Ctx) error {
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
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to count budgets",
			"error":   err.Error(),
		})
	}

	var budgets []models.Budget
	offset := (page - 1) * limit
	if err := query.
		Offset(offset).
		Limit(limit).
		Preload("Owner").
		Order("created_at DESC").
		Find(&budgets).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to fetch budgets",
			"error":   err.Error(),
		})
	}

	responses := make([]types.BudgetResponse, 0, len(budgets))
	for _, budget := range budgets {
		responses = append(responses, modelToBudgetResponse(budget))
	}

	return c.JSON(types.ListResponse{
		Success: true,
		Data:    responses,
		Total:   total,
		Page:    page,
		Limit:   limit,
	})
}

// CreateBudget creates a new budget
func CreateBudget(c fiber.Ctx) error {
	var req types.CreateBudgetRequest

	if err := c.BindJSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	if req.BudgetCode == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Budget code is required",
		})
	}
	if req.TotalBudget <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Total budget must be greater than 0",
		})
	}
	if req.AllocatedAmount < 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Allocated amount cannot be negative",
		})
	}

	userID := c.Locals("user_id").(string)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "User ID not found in token",
		})
	}

	var user models.User
	if err := config.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "User not found",
		})
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
	historyJSON, _ := json.Marshal(emptyHistory)
	budget.ApprovalHistory = historyJSON

	if err := config.DB.Create(&budget).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to create budget",
			"error":   err.Error(),
		})
	}

	config.DB.Preload("Owner").First(&budget)

	return c.Status(fiber.StatusCreated).JSON(types.DetailResponse{
		Success: true,
		Data:    modelToBudgetResponse(budget),
	})
}

// GetBudget retrieves a single budget by ID
func GetBudget(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Budget ID is required",
		})
	}

	var budget models.Budget
	if err := config.DB.
		Preload("Owner").
		Where("id = ?", id).
		First(&budget).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Budget not found",
		})
	}

	return c.JSON(types.DetailResponse{
		Success: true,
		Data:    modelToBudgetResponse(budget),
	})
}

// UpdateBudget updates an existing budget
func UpdateBudget(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Budget ID is required",
		})
	}

	var req types.UpdateBudgetRequest
	if err := c.BindJSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	var budget models.Budget
	if err := config.DB.Where("id = ?", id).First(&budget).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Budget not found",
		})
	}

	if budget.Status != "draft" && budget.Status != "pending" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": fmt.Sprintf("Cannot update budget in %s status", budget.Status),
		})
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
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to update budget",
			"error":   err.Error(),
		})
	}

	config.DB.Preload("Owner").First(&budget)

	return c.JSON(types.DetailResponse{
		Success: true,
		Data:    modelToBudgetResponse(budget),
	})
}

// DeleteBudget deletes a budget
func DeleteBudget(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Budget ID is required",
		})
	}

	var budget models.Budget
	if err := config.DB.Where("id = ?", id).First(&budget).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Budget not found",
		})
	}

	if budget.Status != "draft" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Only draft budgets can be deleted",
		})
	}

	if err := config.DB.Delete(&budget).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to delete budget",
			"error":   err.Error(),
		})
	}

	return c.JSON(types.MessageResponse{
		Success: true,
		Message: "Budget deleted successfully",
	})
}

// ApproveBudget approves a budget
func ApproveBudget(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Budget ID is required",
		})
	}

	var req types.ApproveDocumentRequest
	if err := c.BindJSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	if req.Signature == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Signature is required",
		})
	}

	var budget models.Budget
	if err := config.DB.Where("id = ?", id).First(&budget).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Budget not found",
		})
	}

	approverID := c.Locals("user_id").(string)
	var approver models.User
	if err := config.DB.Where("id = ?", approverID).First(&approver).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Approver not found",
		})
	}

	var approvalHistory []types.ApprovalRecord
	if len(budget.ApprovalHistory) > 0 {
		if err := json.Unmarshal(budget.ApprovalHistory, &approvalHistory); err != nil {
			approvalHistory = []types.ApprovalRecord{}
		}
	}

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
	historyJSON, _ := json.Marshal(approvalHistory)
	budget.ApprovalHistory = historyJSON
	budget.UpdatedAt = time.Now()

	if err := config.DB.Save(&budget).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to approve budget",
			"error":   err.Error(),
		})
	}

	config.DB.Preload("Owner").First(&budget)

	return c.JSON(types.DetailResponse{
		Success: true,
		Data:    modelToBudgetResponse(budget),
	})
}

// RejectBudget rejects a budget
func RejectBudget(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Budget ID is required",
		})
	}

	var req types.RejectDocumentRequest
	if err := c.BindJSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	if req.Remarks == "" || len(req.Remarks) < 10 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Remarks must be at least 10 characters",
		})
	}
	if req.Signature == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Signature is required",
		})
	}

	var budget models.Budget
	if err := config.DB.Where("id = ?", id).First(&budget).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Budget not found",
		})
	}

	approverID := c.Locals("user_id").(string)
	var approver models.User
	if err := config.DB.Where("id = ?", approverID).First(&approver).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Approver not found",
		})
	}

	var approvalHistory []types.ApprovalRecord
	if len(budget.ApprovalHistory) > 0 {
		if err := json.Unmarshal(budget.ApprovalHistory, &approvalHistory); err != nil {
			approvalHistory = []types.ApprovalRecord{}
		}
	}

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
	historyJSON, _ := json.Marshal(approvalHistory)
	budget.ApprovalHistory = historyJSON
	budget.UpdatedAt = time.Now()

	if err := config.DB.Save(&budget).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to reject budget",
			"error":   err.Error(),
		})
	}

	config.DB.Preload("Owner").First(&budget)

	return c.JSON(types.DetailResponse{
		Success: true,
		Data:    modelToBudgetResponse(budget),
	})
}

// Helper function to convert model to response
func modelToBudgetResponse(budget models.Budget) types.BudgetResponse {
	var approvalHistory []types.ApprovalRecord
	if len(budget.ApprovalHistory) > 0 {
		json.Unmarshal(budget.ApprovalHistory, &approvalHistory)
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
