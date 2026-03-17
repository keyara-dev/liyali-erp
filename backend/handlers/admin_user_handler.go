package handlers

import (
	"fmt"
	"log"
	"strings"

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

	// Get creator's user ID for audit log
	creatorID, _ := c.Locals("userID").(string)

	// Parse request
	var req struct {
		Email        string  `json:"email" validate:"required,email"`
		Password     string  `json:"password" validate:"required,min=8"`
		Name         string  `json:"name" validate:"required"`
		FirstName    string  `json:"first_name"`
		LastName     string  `json:"last_name"`
		Role         string  `json:"role"`
		DepartmentID string  `json:"department_id"`
		BranchID     *string `json:"branch_id"`
		Position     string  `json:"position"`
		ManNumber    string  `json:"manNumber"`
		NrcNumber    string  `json:"nrcNumber"`
		Contact      string  `json:"contact"`
	}

	if err := c.BodyParser(&req); err != nil {
		return utils.SendBadRequestError(c, "Invalid request body")
	}

	// Trim whitespace from name fields before validation
	req.Name = strings.TrimSpace(req.Name)
	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)

	// Validate required fields
	if req.Email == "" {
		return utils.SendBadRequestError(c, "Email is required")
	}
	if req.Password == "" {
		return utils.SendBadRequestError(c, "Password is required")
	}
	if err := utils.ValidatePasswordStrength(req.Password); err != nil {
		return utils.SendBadRequestError(c, "Password validation failed: "+err.Error())
	}
	if req.Name == "" && req.FirstName == "" {
		return utils.SendBadRequestError(c, "Name or first name is required")
	}
	if req.Position == "" {
		return utils.SendBadRequestError(c, "Position is required")
	}
	if req.ManNumber == "" {
		return utils.SendBadRequestError(c, "Man Number is required")
	}
	if req.NrcNumber == "" {
		return utils.SendBadRequestError(c, "NRC Number is required")
	}
	if req.Contact == "" {
		return utils.SendBadRequestError(c, "Contact is required")
	}

	// Set default role if not provided
	if req.Role == "" {
		req.Role = "requester"
	}

	// Use name if first/last names not provided
	if req.FirstName == "" && req.LastName == "" {
		req.FirstName = req.Name
	}

	// Validate department belongs to this organisation (if provided)
	if req.DepartmentID != "" {
		var deptCount int64
		if err := config.DB.Table("organization_departments").
			Where("id = ? AND organization_id = ? AND is_active = true", req.DepartmentID, tenant.OrganizationID).
			Count(&deptCount).Error; err != nil || deptCount == 0 {
			return utils.SendBadRequestError(c, "Department not found in this organization")
		}
	}

	// Validate branch belongs to this organisation (if provided)
	if req.BranchID != nil && *req.BranchID != "" {
		var branchCount int64
		if err := config.DB.Table("organization_branches").
			Where("id = ? AND organization_id = ? AND is_active = true", *req.BranchID, tenant.OrganizationID).
			Count(&branchCount).Error; err != nil || branchCount == 0 {
			return utils.SendBadRequestError(c, "Branch not found in this organization")
		}
	}

	// Email lookup — run against config.DB before opening TX to avoid aborting
	// the PostgreSQL transaction on a failed SELECT.
	// Three distinct cases are surfaced to the frontend:
	//   1. No global account  → proceed to creation
	//   2. Global account, already a member → hard block
	//   3. Global account, not yet a member → invite flow (code: "email_has_global_account")
	preCheckService := services.NewUserService(config.DB)
	emailLookup, err := preCheckService.LookupUserByEmailForOrg(tenant.OrganizationID, req.Email)
	if err != nil {
		logging.LogError(c, err, "email_lookup_failed", nil)
		return utils.SendInternalError(c, "Failed to validate email", err)
	}
	if emailLookup.User != nil {
		if emailLookup.IsMember {
			return utils.SendConflictError(c, "This user is already a member of your organization")
		}
		// User has a platform account but belongs to a different org — prompt invite flow
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"success": false,
			"message": "This email belongs to an existing Liyali user",
			"code":    "email_has_global_account",
			"data": fiber.Map{
				"userId": emailLookup.User.ID,
				"name":   emailLookup.User.Name,
				"email":  emailLookup.User.Email,
			},
		})
	}

	// Man Number uniqueness within this organisation
	var manCount int64
	config.DB.Table("users").
		Joins("JOIN organization_members ON organization_members.user_id = users.id").
		Where("organization_members.organization_id = ? AND organization_members.active = true AND users.man_number = ? AND users.deleted_at IS NULL",
			tenant.OrganizationID, req.ManNumber).
		Count(&manCount)
	if manCount > 0 {
		return utils.SendConflictError(c, "A user with this Man Number already exists in this organization")
	}

	// NRC Number uniqueness within this organisation
	var nrcCount int64
	config.DB.Table("users").
		Joins("JOIN organization_members ON organization_members.user_id = users.id").
		Where("organization_members.organization_id = ? AND organization_members.active = true AND users.nrc_number = ? AND users.deleted_at IS NULL",
			tenant.OrganizationID, req.NrcNumber).
		Count(&nrcCount)
	if nrcCount > 0 {
		return utils.SendConflictError(c, "A user with this NRC Number already exists in this organization")
	}

	// Start transaction for atomic user creation
	tx := config.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			_ = utils.SendInternalError(c, "Unexpected error during user creation", fmt.Errorf("%v", r))
		}
	}()

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		tx.Rollback()
		logging.LogError(c, err, "password_hashing_failed", nil)
		return utils.SendInternalError(c, "Failed to process password", err)
	}

	// Resolve role — if it's a UUID look up the name; also capture the role ID for response
	roleName := req.Role
	roleID := ""
	if _, err := uuid.Parse(req.Role); err == nil {
		roleID = req.Role
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

	// Create user — MustChangePassword forces a password reset on first login
	user := &models.User{
		ID:                    uuid.New().String(),
		Email:                 req.Email,
		Name:                  req.Name,
		Password:              hashedPassword,
		Role:                  roleName,
		Active:                true,
		MustChangePassword:    true,
		CurrentOrganizationID: &tenant.OrganizationID,
		Position:              req.Position,
		ManNumber:             req.ManNumber,
		NrcNumber:             req.NrcNumber,
		Contact:               req.Contact,
	}

	log.Printf("[CreateUser] creating user email=%s org=%s", req.Email, tenant.OrganizationID)
	if err := tx.Create(user).Error; err != nil {
		tx.Rollback()
		logging.LogError(c, err, "user_creation_failed", map[string]interface{}{
			"email": req.Email,
		})
		return utils.SendInternalError(c, "Failed to create user", err)
	}

	// Add user to the organisation with department assignment
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
		})
		return utils.SendInternalError(c, "Failed to add user to organization", err)
	}
	// Set branch if provided
	if req.BranchID != nil && *req.BranchID != "" {
		if err := tx.Table("organization_members").
			Where("organization_id = ? AND user_id = ?", tenant.OrganizationID, user.ID).
			Update("branch_id", req.BranchID).Error; err != nil {
			tx.Rollback()
			logging.LogError(c, err, "org_member_branch_set_failed", map[string]interface{}{
				"user_id":   user.ID,
				"branch_id": *req.BranchID,
			})
			return utils.SendInternalError(c, "Failed to set branch assignment", err)
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		logging.LogError(c, err, "transaction_commit_failed", nil)
		return utils.SendInternalError(c, "Failed to complete user creation", err)
	}
	log.Printf("[CreateUser] committed user=%s email=%s org=%s", user.ID, user.Email, tenant.OrganizationID)

	// Audit log — record who created this user
	logging.AddFieldsToRequest(c, map[string]interface{}{
		"created_user_id":  user.ID,
		"created_by":       creatorID,
		"organization_id":  tenant.OrganizationID,
		"role":             roleName,
		"creation_success": true,
	})
	logger.Info("admin_user_creation_successful")

	// Return user response (without sensitive data)
	userResponse := map[string]interface{}{
		"id":                 user.ID,
		"email":              user.Email,
		"name":               user.Name,
		"role":               roleName,
		"roleId":             roleID,
		"is_active":          user.Active,
		"mustChangePassword": user.MustChangePassword,
		"createdAt":          user.CreatedAt,
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
		Name         string  `json:"name"`
		Email        string  `json:"email"`
		Role         string  `json:"role"`
		DepartmentID string  `json:"department_id"`
		BranchID     *string `json:"branch_id"`
		Position     string  `json:"position"`
		ManNumber    string  `json:"manNumber"`
		NrcNumber    string  `json:"nrcNumber"`
		Contact      string  `json:"contact"`
		Status       string  `json:"status"` // "active" | "inactive"
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

	// Man Number uniqueness check (only when the field is being changed)
	if req.ManNumber != "" {
		var manCount int64
		config.DB.Table("users").
			Joins("JOIN organization_members ON organization_members.user_id = users.id").
			Where("organization_members.organization_id = ? AND organization_members.active = true AND users.man_number = ? AND users.id != ? AND users.deleted_at IS NULL",
				tenant.OrganizationID, req.ManNumber, userID).
			Count(&manCount)
		if manCount > 0 {
			return utils.SendConflictError(c, "A user with this Man Number already exists in this organization")
		}
	}

	// NRC Number uniqueness check (only when the field is being changed)
	if req.NrcNumber != "" {
		var nrcCount int64
		config.DB.Table("users").
			Joins("JOIN organization_members ON organization_members.user_id = users.id").
			Where("organization_members.organization_id = ? AND organization_members.active = true AND users.nrc_number = ? AND users.id != ? AND users.deleted_at IS NULL",
				tenant.OrganizationID, req.NrcNumber, userID).
			Count(&nrcCount)
		if nrcCount > 0 {
			return utils.SendConflictError(c, "A user with this NRC Number already exists in this organization")
		}
	}

	// Email uniqueness check (only when the email is being changed)
	if req.Email != "" {
		emailLookup, err := services.NewUserService(config.DB).LookupUserByEmailForOrg(tenant.OrganizationID, req.Email)
		if err != nil {
			return utils.SendInternalError(c, "Failed to validate email", err)
		}
		if emailLookup.User != nil && emailLookup.User.ID != userID {
			if emailLookup.IsMember {
				return utils.SendConflictError(c, "This email belongs to another member of your organization")
			}
			return utils.SendConflictError(c, "This email is already registered on the platform")
		}
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

	// Update branch assignment if provided (nil = keep current; non-nil = set or clear)
	if req.BranchID != nil {
		config.DB.Table("organization_members").
			Where("organization_id = ? AND user_id = ?", tenant.OrganizationID, userID).
			Update("branch_id", req.BranchID)
	}

	logging.AddFieldsToRequest(c, map[string]interface{}{
		"user_id":         userID,
		"organization_id": tenant.OrganizationID,
	})
	logger.Info("org_user_update_successful")

	return utils.SendSuccess(c, fiber.StatusOK, map[string]interface{}{"id": userID}, "User updated successfully", nil)
}