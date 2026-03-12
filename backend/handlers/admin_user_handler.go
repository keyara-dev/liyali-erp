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
		Position    string `json:"position"`
		ManNumber   string `json:"manNumber"`
		NrcNumber   string `json:"nrcNumber"`
		Contact     string `json:"contact"`
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
		Position:              req.Position,
		ManNumber:             req.ManNumber,
		NrcNumber:             req.NrcNumber,
		Contact:               req.Contact,
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

	// System roles are now global — no per-org initialization needed

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

// UpdateOrganizationUser updates a user within the current organization
// Only allows updating fields the org admin is permitted to change
// PUT /api/v1/organization/users/:id
func UpdateOrganizationUser(c *fiber.Ctx) error {
	logger := logging.FromContext(c)

	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return utils.SendUnauthorizedError(c, "Organization context required")
	}

	userID := c.Params("id")
	if userID == "" {
		return utils.SendBadRequestError(c, "User ID is required")
	}

	var req struct {
		Name         string `json:"name"`
		Email        string `json:"email"`
		Role         string `json:"role"`
		DepartmentID string `json:"department_id"`
		Position     string `json:"position"`
		ManNumber    string `json:"manNumber"`
		NrcNumber    string `json:"nrcNumber"`
		Contact      string `json:"contact"`
		Status       string `json:"status"` // "active" | "inactive"
	}

	if err := c.BodyParser(&req); err != nil {
		return utils.SendBadRequestError(c, "Invalid request body")
	}

	// Verify the user belongs to this organisation
	var memberCount int64
	if err := config.DB.Table("organization_members").
		Where("organization_id = ? AND user_id = ? AND active = true", tenant.OrganizationID, userID).
		Count(&memberCount).Error; err != nil || memberCount == 0 {
		return utils.SendNotFoundError(c, "User not found in this organization")
	}

	// Resolve role name from UUID if needed
	roleName := req.Role
	if roleName != "" {
		if _, err := uuid.Parse(roleName); err == nil {
			roleService := services.NewRoleManagementService(config.DB)
			role, err := roleService.GetOrganizationRole(roleName)
			if err != nil {
				return utils.SendBadRequestError(c, "Invalid role ID")
			}
			roleName = role.Name
		}
	}

	// Build user updates
	userUpdates := map[string]interface{}{}
	if req.Name != "" {
		userUpdates["name"] = req.Name
	}
	if req.Email != "" {
		userUpdates["email"] = req.Email
	}
	if roleName != "" {
		userUpdates["role"] = roleName
	}
	if req.Position != "" {
		userUpdates["position"] = req.Position
	}
	if req.ManNumber != "" {
		userUpdates["man_number"] = req.ManNumber
	}
	if req.NrcNumber != "" {
		userUpdates["nrc_number"] = req.NrcNumber
	}
	if req.Contact != "" {
		userUpdates["contact"] = req.Contact
	}
	if req.Status == "active" {
		userUpdates["active"] = true
	} else if req.Status == "inactive" {
		userUpdates["active"] = false
	}

	if len(userUpdates) > 0 {
		if err := config.DB.Model(&models.User{}).Where("id = ?", userID).Updates(userUpdates).Error; err != nil {
			logging.LogError(c, err, "org_user_update_failed", map[string]interface{}{"user_id": userID})
			return utils.SendInternalError(c, "Failed to update user", err)
		}
	}

	// Update department membership if provided
	if req.DepartmentID != "" {
		orgService := services.NewOrganizationService(config.DB)
		deptPtr := &req.DepartmentID
		if err := orgService.AddMemberWithDepartment(tenant.OrganizationID, userID, roleName, deptPtr); err != nil {
			logging.LogError(c, err, "org_member_dept_update_failed", map[string]interface{}{"user_id": userID})
			// Non-fatal — user fields already updated
		}
	} else if roleName != "" {
		// Update role in org_members table too
		config.DB.Table("organization_members").
			Where("organization_id = ? AND user_id = ?", tenant.OrganizationID, userID).
			Update("role", roleName)
	}

	logging.AddFieldsToRequest(c, map[string]interface{}{
		"user_id":         userID,
		"organization_id": tenant.OrganizationID,
	})
	logger.Info("org_user_update_successful")

	return utils.SendSuccess(c, fiber.StatusOK, map[string]interface{}{"id": userID}, "User updated successfully", nil)
}