package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/logging"
	"github.com/liyali/liyali-gateway/middleware"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/services"
	"github.com/liyali/liyali-gateway/utils"
)

// GetUserOrganizations returns all organizations a user belongs to
// GET /api/v1/organizations
func GetUserOrganizations(c *fiber.Ctx) error {
	logger := logging.FromContext(c)
	logger.Info("get_user_organizations_request")

	userID, ok := c.Locals("userID").(string)
	if !ok {
		logging.LogWarn(c, "user_context_missing")
		return utils.SendUnauthorizedError(c, "User context required")
	}

	logging.AddFieldsToRequest(c, map[string]interface{}{
		"operation": "get_user_organizations",
		"user_id":   userID,
	})

	orgService := services.NewOrganizationService(config.DB)
	orgs, err := orgService.GetUserOrganizations(userID)

	if err != nil {
		logging.LogError(c, err, "failed_to_fetch_organizations", map[string]interface{}{
			"error_type": "service_error",
		})
		return utils.SendInternalError(c, "Failed to fetch organizations", err)
	}

	if len(orgs) == 0 {
		orgs = []models.Organization{}
	}

	logger.Info("user_organizations_retrieved_successfully")

	return utils.SendSimpleSuccess(c, orgs, "Organizations retrieved successfully")
}

// CreateOrganization creates a new organization
// POST /api/v1/organizations
func CreateOrganization(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(string)
	if !ok {
		return utils.SendUnauthorizedError(c, "User context required")
	}

	var req struct {
		Name        string `json:"name" validate:"required"`
		Description string `json:"description"`
		LogoURL     string `json:"logoUrl"`
	}

	if err := c.BodyParser(&req); err != nil {
		return utils.SendBadRequestError(c, "Invalid request body")
	}

	orgService := services.NewOrganizationService(config.DB)
	org, err := orgService.CreateOrganization(req.Name, req.Description, req.LogoURL, userID)

	if err != nil {
		return utils.SendInternalError(c, err.Error(), err)
	}

	return utils.SendCreatedSuccess(c, org, "Organization created successfully")
}

// GetOrganizationByID returns organization details by ID
// GET /api/v1/organizations/:id
func GetOrganizationByID(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(string)
	if !ok {
		return utils.SendUnauthorizedError(c, "User context required")
	}

	orgID := c.Params("id")
	if orgID == "" {
		return utils.SendBadRequestError(c, "Organization ID is required")
	}

	orgService := services.NewOrganizationService(config.DB)

	canManage, err := orgService.CanUserManageOrganization(userID, orgID)
	if err != nil {
		return utils.SendInternalError(c, "Failed to verify permissions", err)
	}
	if !canManage {
		return utils.SendForbiddenError(c, "You don't have permission to view this organization")
	}

	org, err := orgService.GetOrganization(orgID)
	if err != nil {
		return utils.SendNotFoundError(c, "Organization not found")
	}

	return utils.SendSimpleSuccess(c, org, "Organization retrieved successfully")
}

// SwitchOrganization sets user's current organization
// POST /api/v1/organizations/:id/switch
func SwitchOrganization(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(string)
	if !ok {
		return utils.SendUnauthorizedError(c, "User context required")
	}

	orgID := c.Params("id")
	if orgID == "" {
		return utils.SendBadRequestError(c, "Organization ID is required")
	}

	orgService := services.NewOrganizationService(config.DB)
	if err := orgService.SwitchOrganization(userID, orgID); err != nil {
		return utils.SendForbiddenError(c, "You do not have access to this organization")
	}

	return utils.SendSimpleSuccess(c, nil, "Organization switched successfully")
}

// GetOrganizationMembers returns all members of an organization
// Supports optional query params: search, role, active, page, page_size
// GET /api/v1/organization/members
func GetOrganizationMembers(c *fiber.Ctx) error {
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return utils.SendUnauthorizedError(c, "Organization context required")
	}

	db := config.DB

	search := c.Query("search")
	role := c.Query("role")
	activeStr := c.Query("active")
	page := c.QueryInt("page", 0)
	pageSize := c.QueryInt("page_size", 0)

	// If no pagination/filter params, fall back to existing service (preserves behavior for other callers)
	if page == 0 && pageSize == 0 && search == "" && role == "" && activeStr == "" {
		orgService := services.NewOrganizationService(db)
		members, err := orgService.GetOrganizationMembers(tenant.OrganizationID)
		if err != nil {
			return utils.SendInternalError(c, "Failed to fetch organization members", err)
		}
		if len(members) == 0 {
			members = []models.OrganizationMember{}
		}
		return utils.SendSimpleSuccess(c, members, "Members retrieved successfully")
	}

	// Paginated + filtered query
	page, pageSize = utils.NormalizePaginationParams(page, pageSize)
	offset := (page - 1) * pageSize

	query := db.Table("organization_members").
		Select(`organization_members.id,
			organization_members.user_id,
			organization_members.organization_id,
			organization_members.role,
			COALESCE(organization_departments.name, organization_members.department, '') as department,
			organization_members.department_id,
			organization_members.active as is_active,
			organization_members.joined_at,
			organization_members.created_at,
			organization_members.updated_at,
			users.name,
			users.email,
			users.last_login,
			COALESCE(users.position, '') as position,
			COALESCE(users.man_number, '') as man_number,
			COALESCE(users.nrc_number, '') as nrc_number,
			COALESCE(users.contact, '') as contact,
			users.preferences`).
		Joins("INNER JOIN users ON users.id = organization_members.user_id").
		Joins("LEFT JOIN organization_departments ON organization_departments.id = organization_members.department_id").
		Where("organization_members.organization_id = ? AND users.deleted_at IS NULL", tenant.OrganizationID)

	countQuery := db.Table("organization_members").
		Joins("INNER JOIN users ON users.id = organization_members.user_id").
		Where("organization_members.organization_id = ? AND users.deleted_at IS NULL", tenant.OrganizationID)

	if search != "" {
		pattern := "%" + search + "%"
		query = query.Where("LOWER(users.name) LIKE LOWER(?) OR LOWER(users.email) LIKE LOWER(?)", pattern, pattern)
		countQuery = countQuery.Where("LOWER(users.name) LIKE LOWER(?) OR LOWER(users.email) LIKE LOWER(?)", pattern, pattern)
	}
	if role != "" {
		query = query.Where("organization_members.role = ?", role)
		countQuery = countQuery.Where("organization_members.role = ?", role)
	}
	if activeStr != "" {
		isActive := activeStr == "true" || activeStr == "1"
		query = query.Where("organization_members.active = ?", isActive)
		countQuery = countQuery.Where("organization_members.active = ?", isActive)
	}

	var total int64
	countQuery.Count(&total)

	var rows []map[string]interface{}
	if err := query.Order("organization_members.created_at DESC").
		Offset(offset).Limit(pageSize).Find(&rows).Error; err != nil {
		return utils.SendInternalError(c, "Failed to fetch organization members", err)
	}

	totalPages := int64(1)
	if pageSize > 0 {
		totalPages = (total + int64(pageSize) - 1) / int64(pageSize)
	}

	return c.Status(fiber.StatusOK).JSON(map[string]interface{}{
		"success": true,
		"message": "Members retrieved successfully",
		"data": map[string]interface{}{
			"members":     rows,
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": totalPages,
		},
	})
}

// AddOrganizationMember adds a user to an organization
// POST /api/v1/organization/members
func AddOrganizationMember(c *fiber.Ctx) error {
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return utils.SendUnauthorizedError(c, "Organization context required")
	}

	var req struct {
		UserID string `json:"userId" validate:"required"`
		Role   string `json:"role"`
	}

	if err := c.BodyParser(&req); err != nil {
		return utils.SendBadRequestError(c, "Invalid request body")
	}

	if req.Role == "" {
		req.Role = "requester"
	}

	orgService := services.NewOrganizationService(config.DB)
	if err := orgService.AddMember(tenant.OrganizationID, req.UserID, req.Role); err != nil {
		return utils.SendInternalError(c, "Failed to add member", err)
	}

	return utils.SendSimpleSuccess(c, nil, "Member added successfully")
}

// RemoveOrganizationMember removes a user from an organization
// DELETE /api/v1/organization/members/:userId
func RemoveOrganizationMember(c *fiber.Ctx) error {
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return utils.SendUnauthorizedError(c, "Organization context required")
	}

	userID := c.Params("userId")
	if userID == "" {
		return utils.SendBadRequestError(c, "User ID is required")
	}

	orgService := services.NewOrganizationService(config.DB)
	if err := orgService.RemoveMember(tenant.OrganizationID, userID); err != nil {
		return utils.SendInternalError(c, "Failed to remove member", err)
	}

	return utils.SendSimpleSuccess(c, nil, "Member removed successfully")
}

// GetOrganizationSettings retrieves organization settings
// GET /api/v1/organization/settings
func GetOrganizationSettings(c *fiber.Ctx) error {
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return utils.SendUnauthorizedError(c, "Organization context required")
	}

	orgService := services.NewOrganizationService(config.DB)
	settings, err := orgService.GetOrganizationSettings(tenant.OrganizationID)

	if err != nil {
		return utils.SendInternalError(c, "Failed to fetch organization settings", err)
	}

	return utils.SendSimpleSuccess(c, settings, "Settings retrieved successfully")
}

// UpdateOrganizationSettings updates organization settings
// PUT /api/v1/organization/settings
func UpdateOrganizationSettings(c *fiber.Ctx) error {
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return utils.SendUnauthorizedError(c, "Organization context required")
	}

	var settings struct {
		RequireDigitalSignatures bool    `json:"requireDigitalSignatures"`
		DefaultApprovalChain     string  `json:"defaultApprovalChain"`
		Currency                 string  `json:"currency"`
		FiscalYearStart          int     `json:"fiscalYearStart"`
		EnableBudgetValidation   bool    `json:"enableBudgetValidation"`
		BudgetVarianceThreshold  float64 `json:"budgetVarianceThreshold"`
		ProcurementFlow          string  `json:"procurementFlow"`
		StampImageURL            string  `json:"stampImageUrl"`
	}

	if err := c.BodyParser(&settings); err != nil {
		return utils.SendBadRequestError(c, "Invalid request body")
	}

	orgService := services.NewOrganizationService(config.DB)

	orgSettings := &models.OrganizationSettings{
		RequireDigitalSignatures: settings.RequireDigitalSignatures,
		DefaultApprovalChain:     settings.DefaultApprovalChain,
		Currency:                 settings.Currency,
		FiscalYearStart:          settings.FiscalYearStart,
		EnableBudgetValidation:   settings.EnableBudgetValidation,
		BudgetVarianceThreshold:  settings.BudgetVarianceThreshold,
		ProcurementFlow:          settings.ProcurementFlow,
		StampImageURL:            settings.StampImageURL,
	}

	if err := orgService.UpdateOrganizationSettings(tenant.OrganizationID, orgSettings); err != nil {
		return utils.SendInternalError(c, "Failed to update settings", err)
	}

	return utils.SendSimpleSuccess(c, nil, "Settings updated successfully")
}

// UpdateOrganization updates organization details
// PUT /api/v1/organizations/:id
func UpdateOrganization(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(string)
	if !ok {
		return utils.SendUnauthorizedError(c, "User context required")
	}

	orgID := c.Params("id")
	if orgID == "" {
		return utils.SendBadRequestError(c, "Organization ID is required")
	}

	var req struct {
		Name        string  `json:"name" validate:"required"`
		Description string  `json:"description"`
		LogoURL     *string `json:"logoUrl"`
		Tagline     *string `json:"tagline"`
	}

	if err := c.BodyParser(&req); err != nil {
		return utils.SendBadRequestError(c, "Invalid request body")
	}

	orgService := services.NewOrganizationService(config.DB)

	// Check if user can manage this organization
	canManage, err := orgService.CanUserManageOrganization(userID, orgID)
	if err != nil {
		return utils.SendInternalError(c, "Failed to verify permissions", err)
	}
	if !canManage {
		return utils.SendForbiddenError(c, "You don't have permission to update this organization")
	}

	if err := orgService.UpdateOrganization(orgID, req.Name, req.Description, req.LogoURL, req.Tagline); err != nil {
		return utils.SendInternalError(c, err.Error(), err)
	}

	org, _ := orgService.GetOrganization(orgID)

	return utils.SendSimpleSuccess(c, org, "Organization updated successfully")
}

// DeleteOrganization soft deletes an organization
// DELETE /api/v1/organizations/:id
func DeleteOrganization(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(string)
	if !ok {
		return utils.SendUnauthorizedError(c, "User context required")
	}

	orgID := c.Params("id")
	if orgID == "" {
		return utils.SendBadRequestError(c, "Organization ID is required")
	}

	orgService := services.NewOrganizationService(config.DB)

	// Check if user can manage this organization
	canManage, err := orgService.CanUserManageOrganization(userID, orgID)
	if err != nil {
		return utils.SendInternalError(c, "Failed to verify permissions", err)
	}
	if !canManage {
		return utils.SendForbiddenError(c, "You don't have permission to delete this organization")
	}

	if err := orgService.DeleteOrganization(orgID, userID); err != nil {
		return utils.SendInternalError(c, err.Error(), err)
	}

	return utils.SendSimpleSuccess(c, nil, "Organization deleted successfully")
}
