package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/logging"
	"github.com/liyali/liyali-gateway/middleware"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/services"
	"github.com/liyali/liyali-gateway/utils"
)

// CreateOrganizationUser creates a new user directly in the current organization
// This is used by admins to create team members without personal organizations
// POST /api/v1/organization/users
func CreateOrganizationUser(c *fiber.Ctx) error {
	logger := logging.FromContext(c)
	logger.Info("admin_user_creation_attempt")

	// Get tenant context (admin's organization)
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return utils.SendUnauthorizedError(c, "Organization context required")
	}

	// Parse request
	var req struct {
		Email       string `json:"email" validate:"required,email"`
		Password    string `json:"password" validate:"required,min=8"`
		Name        string `json:"name" validate:"required"`
		FirstName   string `json:"first_name"`
		LastName    string `json:"last_name"`
		Role        string `json:"role"`
		DepartmentID string `json:"department_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		return utils.SendBadRequestError(c, "Invalid request body")
	}

	// Validate request manually
	if req.Email == "" {
		return utils.SendBadRequestError(c, "Email is required")
	}
	if req.Password == "" || len(req.Password) < 8 {
		return utils.SendBadRequestError(c, "Password is required and must be at least 8 characters")
	}
	if req.Name == "" && req.FirstName == "" {
		return utils.SendBadRequestError(c, "Name or first name is required")
	}

	// Set default role if not provided
	if req.Role == "" {
		req.Role = "requester"
	}

	// Use name if first/last names not provided
	if req.FirstName == "" && req.LastName == "" {
		req.FirstName = req.Name
	}

	// Start transaction for atomic user creation
	tx := config.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Check if user already exists
	userService := services.NewUserService(tx)
	existingUser, err := userService.GetUserByEmail(tenant.OrganizationID, req.Email)
	if err == nil && existingUser != nil {
		tx.Rollback()
		return utils.SendConflictError(c, "User with this email already exists")
	}

	// Validate password strength
	if err := utils.ValidatePasswordStrength(req.Password); err != nil {
		tx.Rollback()
		return utils.SendBadRequestError(c, "Password validation failed: "+err.Error())
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		tx.Rollback()
		logging.LogError(c, err, "password_hashing_failed", nil)
		return utils.SendInternalError(c, "Failed to process password", err)
	}

	// Resolve role name from role ID if it's a UUID
	roleName := req.Role
	if _, err := uuid.Parse(req.Role); err == nil {
		// It's a UUID, look up the role name
		roleService := services.NewRoleManagementService(tx)
		role, err := roleService.GetOrganizationRole(req.Role)
		if err != nil {
			tx.Rollback()
			logging.LogError(c, err, "role_lookup_failed", map[string]interface{}{
				"role_id": req.Role,
			})
			return utils.SendBadRequestError(c, "Invalid role ID")
		}
		roleName = role.Name
	}

	// Create user without personal organization
	user := &models.User{
		ID:                    uuid.New().String(),
		Email:                 req.Email,
		Name:                  req.Name,
		Password:              hashedPassword,
		Role:                  roleName, // Store role name, not role ID
		Active:                true,
		CurrentOrganizationID: &tenant.OrganizationID, // Set to admin's organization
	}

	// Create user in database
	if err := tx.Create(user).Error; err != nil {
		tx.Rollback()
		logging.LogError(c, err, "user_creation_failed", map[string]interface{}{
			"email": req.Email,
		})
		return utils.SendInternalError(c, "Failed to create user", err)
	}

	// Add user to the admin's organization with department assignment
	orgService := services.NewOrganizationService(tx)
	var departmentPtr *string
	if req.DepartmentID != "" {
		departmentPtr = &req.DepartmentID
	}
	
	if err := orgService.AddMemberWithDepartment(tenant.OrganizationID, user.ID, roleName, departmentPtr); err != nil {
		tx.Rollback()
		logging.LogError(c, err, "organization_member_addition_failed", map[string]interface{}{
			"user_id":         user.ID,
			"organization_id": tenant.OrganizationID,
			"role":           roleName,
			"department_id":  req.DepartmentID,
		})
		return utils.SendInternalError(c, "Failed to add user to organization", err)
	}

	// Remove the separate department assignment since it's now handled above
	// The AddMemberWithDepartment method handles both role and department assignment

	// Initialize default roles for the user in this organization
	roleService := services.NewRoleManagementService(tx)
	if err := roleService.InitializeDefaultRolesForOrganization(tenant.OrganizationID); err != nil {
		// Don't fail, just log warning - roles might already exist
		logging.WithFields(map[string]interface{}{
			"organization_id": tenant.OrganizationID,
		}).WithError(err).Warn("failed_to_initialize_default_roles")
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		logging.LogError(c, err, "transaction_commit_failed", nil)
		return utils.SendInternalError(c, "Failed to complete user creation", err)
	}

	// Log successful creation
	logging.AddFieldsToRequest(c, map[string]interface{}{
		"user_id":         user.ID,
		"organization_id": tenant.OrganizationID,
		"role":           roleName,
		"creation_success": true,
	})

	logger.Info("admin_user_creation_successful")

	// Return user response (without sensitive data)
	userResponse := map[string]interface{}{
		"id":        user.ID,
		"email":     user.Email,
		"name":      user.Name,
		"role":      roleName,
		"active":    user.Active,
		"createdAt": user.CreatedAt,
	}

	return utils.SendCreatedSuccess(c, userResponse, "User created successfully")
}

// GetOrganizationUsers returns all users in the current organization
// This is an alias for GetOrganizationMembers but with a cleaner endpoint
// GET /api/v1/organization/users
func GetOrganizationUsers(c *fiber.Ctx) error {
	// Delegate to the existing GetOrganizationMembers handler
	return GetOrganizationMembers(c)
}