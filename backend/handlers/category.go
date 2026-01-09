package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/middleware"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/types"
	"github.com/liyali/liyali-gateway/utils"
)

// GetCategories retrieves all categories with pagination and filtering
func GetCategories(c *fiber.Ctx) error {
	// Get organization context from tenant middleware
	tenant, err := middleware.GetTenantContext(*c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Organization context required",
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

	active := c.Query("active")

	// SECURITY: Always filter by organization ID first
	query := db.Where("organization_id = ?", tenant.OrganizationID)
	
	if active == "true" {
		query = query.Where("active = ?", true)
	} else if active == "false" {
		query = query.Where("active = ?", false)
	}

	var total int64
	if err := query.Model(&models.Category{}).Count(&total).Error; err != nil {
		return utils.SendInternalError(c, "Failed to count categories", err)
	}

	var categories []models.Category
	offset := (page - 1) * limit
	if err := query.
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&categories).Error; err != nil {
		return utils.SendInternalError(c, "Failed to fetch categories", err)
	}

	responses := make([]types.CategoryResponse, 0, len(categories))
	for _, category := range categories {
		budgetCodes, _ := getCategoryBudgetCodes(category.ID)
		responses = append(responses, modelToCategoryResponse(category, budgetCodes))
	}

	return utils.SendPaginatedSuccess(c, responses, "Categories retrieved successfully", page, limit, total)
}

// CreateCategory creates a new category with budget code mappings
func CreateCategory(c *fiber.Ctx) error {
	// Get organization context from tenant middleware
	tenant, err := middleware.GetTenantContext(*c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Organization context required",
		})
	}

	var req types.CreateCategoryRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	if req.Name == "" || len(req.Name) < 3 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Category name is required and must be at least 3 characters",
		})
	}

	// SECURITY: Check if category with same name already exists in THIS organization
	var existingCategory models.Category
	if err := config.DB.Where("name = ? AND organization_id = ?", req.Name, tenant.OrganizationID).First(&existingCategory).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"success": false,
			"message": "Category with this name already exists in your organization",
		})
	}

	category := models.Category{
		ID:             uuid.New().String(),
		OrganizationID: tenant.OrganizationID, // SECURITY: Set organization ID
		Name:        req.Name,
		Description: req.Description,
		Active:      true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := config.DB.Create(&category).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to create category",
			"error":   err.Error(),
		})
	}

	// Create budget code mappings if provided
	if len(req.BudgetCodes) > 0 {
		for _, budgetCode := range req.BudgetCodes {
			mapping := models.CategoryBudgetCode{
				ID:         uuid.New().String(),
				CategoryID: category.ID,
				BudgetCode: budgetCode,
				Active:     true,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			}
			if err := config.DB.Create(&mapping).Error; err != nil {
				// Log error but don't fail the category creation
				continue
			}
		}
	}

	return c.Status(fiber.StatusCreated).JSON(types.DetailResponse{
		Success: true,
		Data:    modelToCategoryResponse(category, req.BudgetCodes),
	})
}

// GetCategory retrieves a single category by ID with its budget codes
func GetCategory(c *fiber.Ctx) error {
	// Get organization context from tenant middleware
	tenant, err := middleware.GetTenantContext(*c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Organization context required",
		})
	}

	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Category ID is required",
		})
	}

	var category models.Category
	// SECURITY: Filter by organization ID to prevent cross-organization access
	if err := config.DB.Where("id = ? AND organization_id = ?", id, tenant.OrganizationID).First(&category).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Category not found",
		})
	}

	budgetCodes, _ := getCategoryBudgetCodes(id)

	return c.JSON(types.DetailResponse{
		Success: true,
		Data:    modelToCategoryResponse(category, budgetCodes),
	})
}

// UpdateCategory updates an existing category and its budget code mappings
func UpdateCategory(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Category ID is required",
		})
	}

	var req types.UpdateCategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	var category models.Category
	if err := config.DB.Where("id = ?", id).First(&category).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Category not found",
		})
	}

	if req.Name != "" {
		if len(req.Name) < 3 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Category name must be at least 3 characters",
			})
		}
		// Check if name is already used by another category
		var existingCategory models.Category
		if err := config.DB.Where("name = ? AND id != ?", req.Name, id).First(&existingCategory).Error; err == nil {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"success": false,
				"message": "Category with this name already exists",
			})
		}
		category.Name = req.Name
	}

	if req.Description != "" {
		category.Description = req.Description
	}

	if req.Active != nil {
		category.Active = *req.Active
	}

	category.UpdatedAt = time.Now()

	if err := config.DB.Save(&category).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to update category",
			"error":   err.Error(),
		})
	}

	// Update budget code mappings if provided
	budgetCodes := req.BudgetCodes
	if len(req.BudgetCodes) > 0 {
		// Delete existing mappings
		config.DB.Where("category_id = ?", id).Delete(&models.CategoryBudgetCode{})

		// Create new mappings
		for _, budgetCode := range req.BudgetCodes {
			mapping := models.CategoryBudgetCode{
				ID:         uuid.New().String(),
				CategoryID: id,
				BudgetCode: budgetCode,
				Active:     true,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			}
			config.DB.Create(&mapping)
		}
	} else {
		budgetCodes, _ = getCategoryBudgetCodes(id)
	}

	return c.JSON(types.DetailResponse{
		Success: true,
		Data:    modelToCategoryResponse(category, budgetCodes),
	})
}

// DeleteCategory soft deletes a category by setting Active to false
func DeleteCategory(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Category ID is required",
		})
	}

	var category models.Category
	if err := config.DB.Where("id = ?", id).First(&category).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Category not found",
		})
	}

	// Soft delete by setting Active to false
	if err := config.DB.Model(&category).Update("active", false).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to delete category",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Category deleted successfully",
	})
}

// GetCategoryBudgetCodes retrieves all budget codes for a category
func GetCategoryBudgetCodes(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Category ID is required",
		})
	}

	budgetCodes, err := getCategoryBudgetCodes(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to fetch budget codes",
			"error":   err.Error(),
		})
	}

	return c.JSON(types.DetailResponse{
		Success: true,
		Data:    budgetCodes,
	})
}

// AddBudgetCodeToCategory adds a new budget code mapping to a category
func AddBudgetCodeToCategory(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Category ID is required",
		})
	}

	// Verify category exists
	var category models.Category
	if err := config.DB.Where("id = ?", id).First(&category).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Category not found",
		})
	}

	var req struct {
		BudgetCode string `json:"budgetCode" validate:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
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

	// Check if mapping already exists
	var existingMapping models.CategoryBudgetCode
	if err := config.DB.Where("category_id = ? AND budget_code = ?", id, req.BudgetCode).First(&existingMapping).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"success": false,
			"message": "Budget code is already mapped to this category",
		})
	}

	mapping := models.CategoryBudgetCode{
		ID:         uuid.New().String(),
		CategoryID: id,
		BudgetCode: req.BudgetCode,
		Active:     true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := config.DB.Create(&mapping).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to add budget code",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(types.DetailResponse{
		Success: true,
		Data:    modelToCategoryBudgetCodeResponse(mapping),
	})
}

// RemoveBudgetCodeFromCategory removes a budget code mapping from a category
func RemoveBudgetCodeFromCategory(c *fiber.Ctx) error {
	id := c.Params("id")
	budgetCode := c.Params("budgetCode")

	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Category ID is required",
		})
	}

	if budgetCode == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Budget code is required",
		})
	}

	var mapping models.CategoryBudgetCode
	if err := config.DB.Where("category_id = ? AND budget_code = ?", id, budgetCode).First(&mapping).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Budget code mapping not found",
		})
	}

	if err := config.DB.Delete(&mapping).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to remove budget code",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Budget code removed successfully",
	})
}

// Helper functions

// getCategoryBudgetCodes retrieves all budget codes for a given category ID
func getCategoryBudgetCodes(categoryID string) ([]string, error) {
	var mappings []models.CategoryBudgetCode
	if err := config.DB.Where("category_id = ? AND active = ?", categoryID, true).Find(&mappings).Error; err != nil {
		return []string{}, err
	}

	budgetCodes := make([]string, 0, len(mappings))
	for _, mapping := range mappings {
		budgetCodes = append(budgetCodes, mapping.BudgetCode)
	}

	return budgetCodes, nil
}

// modelToCategoryResponse converts a Category model to a CategoryResponse
func modelToCategoryResponse(category models.Category, budgetCodes []string) types.CategoryResponse {
	if budgetCodes == nil {
		budgetCodes = []string{}
	}

	return types.CategoryResponse{
		ID:          category.ID,
		Name:        category.Name,
		Description: category.Description,
		BudgetCodes: budgetCodes,
		Active:      category.Active,
		CreatedAt:   category.CreatedAt,
		UpdatedAt:   category.UpdatedAt,
	}
}

// modelToCategoryBudgetCodeResponse converts a CategoryBudgetCode model to a response
func modelToCategoryBudgetCodeResponse(mapping models.CategoryBudgetCode) types.CategoryBudgetCodeResponse {
	return types.CategoryBudgetCodeResponse{
		ID:         mapping.ID,
		CategoryID: mapping.CategoryID,
		BudgetCode: mapping.BudgetCode,
		Active:     mapping.Active,
		CreatedAt:  mapping.CreatedAt,
		UpdatedAt:  mapping.UpdatedAt,
	}
}
