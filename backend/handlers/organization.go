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

	logger.Info("user_organizations_retrieved_successfully", map[string]interface{}{
		"organizations_count": len(orgs),
	})

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
	}

	if err := c.BodyParser(&req); err != nil {
		return utils.SendBadRequestError(c, "Invalid request body")
	}

	orgService := services.NewOrganizationService(config.DB)
	org, err := orgService.CreateOrganization(req.Name, req.Description, userID)

	if err != nil {
		return utils.SendInternalError(c, err.Error(), err)
	}

	return utils.SendCreatedSuccess(c, org, "Organization created successfully")
}

// SwitchOrganization sets user's current organization
// POST /api/v1/organizations/:id/switch
func SwitchOrganization(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User context required",
		})
	}

	orgID := c.Params("id")
	if orgID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Organization ID is required",
		})
	}

	orgService := services.NewOrganizationService(config.DB)
	if err := orgService.SwitchOrganization(userID, orgID); err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Organization switched successfully",
	})
}

// GetOrganizationMembers returns all members of an organization
// GET /api/v1/organization/members
func GetOrganizationMembers(c *fiber.Ctx) error {
	tenant, err := middleware.GetTenantContext(*c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Organization context required",
		})
	}

	orgService := services.NewOrganizationService(config.DB)
	members, err := orgService.GetOrganizationMembers(tenant.OrganizationID)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch organization members",
		})
	}

	if len(members) == 0 {
		members = []models.OrganizationMember{}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    members,
	})
}

// AddOrganizationMember adds a user to an organization
// POST /api/v1/organization/members
func AddOrganizationMember(c *fiber.Ctx) error {
	tenant, err := middleware.GetTenantContext(*c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Organization context required",
		})
	}

	// Verify admin role
	if tenant.UserRole != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Only administrators can add members",
		})
	}

	var req struct {
		UserID string `json:"userId" validate:"required"`
		Role   string `json:"role"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Role == "" {
		req.Role = "requester"
	}

	orgService := services.NewOrganizationService(config.DB)
	if err := orgService.AddMember(tenant.OrganizationID, req.UserID, req.Role); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Member added successfully",
	})
}

// RemoveOrganizationMember removes a user from an organization
// DELETE /api/v1/organization/members/:userId
func RemoveOrganizationMember(c *fiber.Ctx) error {
	tenant, err := middleware.GetTenantContext(*c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Organization context required",
		})
	}

	// Verify admin role
	if tenant.UserRole != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Only administrators can remove members",
		})
	}

	userID := c.Params("userId")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "User ID is required",
		})
	}

	orgService := services.NewOrganizationService(config.DB)
	if err := orgService.RemoveMember(tenant.OrganizationID, userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Member removed successfully",
	})
}

// GetOrganizationSettings retrieves organization settings
// GET /api/v1/organization/settings
func GetOrganizationSettings(c *fiber.Ctx) error {
	tenant, err := middleware.GetTenantContext(*c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Organization context required",
		})
	}

	orgService := services.NewOrganizationService(config.DB)
	settings, err := orgService.GetOrganizationSettings(tenant.OrganizationID)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch organization settings",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    settings,
	})
}

// UpdateOrganizationSettings updates organization settings
// PUT /api/v1/organization/settings
func UpdateOrganizationSettings(c *fiber.Ctx) error {
	tenant, err := middleware.GetTenantContext(*c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Organization context required",
		})
	}

	// Verify admin role
	if tenant.UserRole != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Only administrators can update settings",
		})
	}

	var settings struct {
		RequireDigitalSignatures bool    `json:"requireDigitalSignatures"`
		DefaultApprovalChain     string  `json:"defaultApprovalChain"`
		Currency                 string  `json:"currency"`
		FiscalYearStart          int     `json:"fiscalYearStart"`
		EnableBudgetValidation   bool    `json:"enableBudgetValidation"`
		BudgetVarianceThreshold  float64 `json:"budgetVarianceThreshold"`
	}

	if err := c.BodyParser(&settings); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	orgService := services.NewOrganizationService(config.DB)

	orgSettings := &models.OrganizationSettings{
		RequireDigitalSignatures: settings.RequireDigitalSignatures,
		DefaultApprovalChain:     settings.DefaultApprovalChain,
		Currency:                 settings.Currency,
		FiscalYearStart:          settings.FiscalYearStart,
		EnableBudgetValidation:   settings.EnableBudgetValidation,
		BudgetVarianceThreshold:  settings.BudgetVarianceThreshold,
	}

	if err := orgService.UpdateOrganizationSettings(tenant.OrganizationID, orgSettings); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Settings updated successfully",
	})
}

// UpdateOrganization updates organization details
// PUT /api/v1/organizations/:id
func UpdateOrganization(c *fiber.Ctx) error {
	tenant, err := middleware.GetTenantContext(*c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Organization context required",
		})
	}

	// Verify admin role
	if tenant.UserRole != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Only administrators can update organization",
		})
	}

	orgID := c.Params("id")
	if orgID != tenant.OrganizationID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You can only update your own organization",
		})
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	orgService := services.NewOrganizationService(config.DB)
	if err := orgService.UpdateOrganization(orgID, req.Name, req.Description); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	org, _ := orgService.GetOrganization(orgID)

	return c.JSON(fiber.Map{
		"success": true,
		"data":    org,
	})
}
