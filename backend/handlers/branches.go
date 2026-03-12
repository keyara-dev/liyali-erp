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

// GetBranches lists all branches for the tenant organization.
// GET /api/v1/branches
func GetBranches(c *fiber.Ctx) error {
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return utils.SendUnauthorizedError(c, "Organization context required")
	}

	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("page_size", 10)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	query := config.DB.Model(&models.OrganizationBranch{}).
		Where("organization_id = ?", tenant.OrganizationID)

	if provinceID := c.Query("province_id"); provinceID != "" {
		query = query.Where("province_id = ?", provinceID)
	}
	if townID := c.Query("town_id"); townID != "" {
		query = query.Where("town_id = ?", townID)
	}
	if isActive := c.Query("is_active"); isActive == "true" {
		query = query.Where("is_active = true")
	} else if isActive == "false" {
		query = query.Where("is_active = false")
	}

	var total int64
	query.Count(&total)

	var branches []models.OrganizationBranch
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("name ASC").Find(&branches).Error; err != nil {
		return utils.SendInternalError(c, "Failed to retrieve branches", err)
	}

	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)
	return utils.SendSuccess(c, fiber.StatusOK, branches, "Branches retrieved successfully", &types.PaginationMeta{
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    int64(page) < totalPages,
		HasPrev:    page > 1,
	})
}

// GetBranch returns a single branch by ID.
// GET /api/v1/branches/:id
func GetBranch(c *fiber.Ctx) error {
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return utils.SendUnauthorizedError(c, "Organization context required")
	}

	id := c.Params("id")
	if id == "" {
		return utils.SendBadRequestError(c, "Branch ID is required")
	}

	var branch models.OrganizationBranch
	if err := config.DB.
		Where("id = ? AND organization_id = ?", id, tenant.OrganizationID).
		First(&branch).Error; err != nil {
		return utils.SendNotFoundError(c, "Branch not found")
	}

	return utils.SendSimpleSuccess(c, branch, "Branch retrieved successfully")
}

// CreateBranch creates a new branch for the tenant organization.
// POST /api/v1/branches
func CreateBranch(c *fiber.Ctx) error {
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return utils.SendUnauthorizedError(c, "Organization context required")
	}

	var req struct {
		Name       string  `json:"name"`
		Code       string  `json:"code"`
		ProvinceID string  `json:"province_id"`
		TownID     string  `json:"town_id"`
		Address    string  `json:"address"`
		ManagerID  *string `json:"manager_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		return utils.SendBadRequestError(c, "Invalid request body")
	}
	if req.Name == "" {
		return utils.SendBadRequestError(c, "Branch name is required")
	}
	if req.Code == "" {
		return utils.SendBadRequestError(c, "Branch code is required")
	}
	if req.TownID == "" || req.ProvinceID == "" {
		return utils.SendBadRequestError(c, "Town ID and Province ID are required")
	}

	// Ensure code is unique within the org
	var existing int64
	config.DB.Model(&models.OrganizationBranch{}).
		Where("organization_id = ? AND code = ?", tenant.OrganizationID, req.Code).
		Count(&existing)
	if existing > 0 {
		return utils.SendConflictError(c, "A branch with this code already exists")
	}

	branch := models.OrganizationBranch{
		ID:             uuid.New().String(),
		OrganizationID: tenant.OrganizationID,
		Name:           req.Name,
		Code:           req.Code,
		ProvinceID:     req.ProvinceID,
		TownID:         req.TownID,
		Address:        req.Address,
		ManagerID:      req.ManagerID,
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := config.DB.Create(&branch).Error; err != nil {
		return utils.SendInternalError(c, "Failed to create branch", err)
	}

	return utils.SendCreatedSuccess(c, branch, "Branch created successfully")
}

// UpdateBranch updates an existing branch.
// PUT /api/v1/branches/:id
func UpdateBranch(c *fiber.Ctx) error {
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return utils.SendUnauthorizedError(c, "Organization context required")
	}

	id := c.Params("id")
	if id == "" {
		return utils.SendBadRequestError(c, "Branch ID is required")
	}

	var req struct {
		Name       string  `json:"name"`
		Code       string  `json:"code"`
		ProvinceID string  `json:"province_id"`
		TownID     string  `json:"town_id"`
		Address    string  `json:"address"`
		ManagerID  *string `json:"manager_id"`
		IsActive   *bool   `json:"is_active"`
	}

	if err := c.BodyParser(&req); err != nil {
		return utils.SendBadRequestError(c, "Invalid request body")
	}

	var branch models.OrganizationBranch
	if err := config.DB.
		Where("id = ? AND organization_id = ?", id, tenant.OrganizationID).
		First(&branch).Error; err != nil {
		return utils.SendNotFoundError(c, "Branch not found")
	}

	updates := map[string]interface{}{"updated_at": time.Now()}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Code != "" {
		updates["code"] = req.Code
	}
	if req.ProvinceID != "" {
		updates["province_id"] = req.ProvinceID
	}
	if req.TownID != "" {
		updates["town_id"] = req.TownID
	}
	if req.Address != "" {
		updates["address"] = req.Address
	}
	if req.ManagerID != nil {
		updates["manager_id"] = req.ManagerID
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if err := config.DB.Model(&branch).Updates(updates).Error; err != nil {
		return utils.SendInternalError(c, "Failed to update branch", err)
	}

	config.DB.First(&branch, "id = ?", id)
	return utils.SendSimpleSuccess(c, branch, "Branch updated successfully")
}

// DeleteBranch deletes a branch (hard delete).
// DELETE /api/v1/branches/:id
func DeleteBranch(c *fiber.Ctx) error {
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return utils.SendUnauthorizedError(c, "Organization context required")
	}

	id := c.Params("id")
	if id == "" {
		return utils.SendBadRequestError(c, "Branch ID is required")
	}

	var branch models.OrganizationBranch
	if err := config.DB.
		Where("id = ? AND organization_id = ?", id, tenant.OrganizationID).
		First(&branch).Error; err != nil {
		return utils.SendNotFoundError(c, "Branch not found")
	}

	if err := config.DB.Delete(&branch).Error; err != nil {
		return utils.SendInternalError(c, "Failed to delete branch", err)
	}

	return utils.SendSimpleSuccess(c, nil, "Branch deleted successfully")
}
