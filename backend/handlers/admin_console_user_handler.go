package handlers

import (
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/utils"
	"golang.org/x/crypto/bcrypt"
)

// Admin console user management - manages users with admin roles
// Admin users are regular users filtered by role IN ('admin', 'super_admin', 'compliance_officer')

var adminRoles = []string{"admin", "super_admin", "compliance_officer"}

// AdminGetAdminUsers returns all admin-level users
func AdminGetAdminUsers(c *fiber.Ctx) error {
	db := config.DB

	search := c.Query("search")
	isActive := c.Query("is_active")
	isSuperAdmin := c.Query("is_super_admin")

	query := db.Table("users").
		Select(`id, email, name, role,
			CASE WHEN active = true THEN 'active' ELSE 'suspended' END as status,
			active as is_active, is_super_admin,
			created_at, updated_at, last_login`).
		Where("deleted_at IS NULL AND (role IN ? OR is_super_admin = ?)", adminRoles, true)

	if search != "" {
		searchTerm := "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(email) LIKE ?", searchTerm, searchTerm)
	}

	if isActive == "true" {
		query = query.Where("active = ?", true)
	} else if isActive == "false" {
		query = query.Where("active = ?", false)
	}

	if isSuperAdmin == "true" {
		query = query.Where("is_super_admin = ?", true)
	} else if isSuperAdmin == "false" {
		query = query.Where("is_super_admin = ?", false)
	}

	query = query.Order("created_at DESC")

	var users []map[string]interface{}
	if err := query.Find(&users).Error; err != nil {
		log.Printf("Error getting admin users: %v", err)
		return utils.SendInternalError(c, "Failed to retrieve admin users", err)
	}

	return utils.SendSimpleSuccess(c, users, "Admin users retrieved successfully")
}

// AdminGetAdminUserStats returns admin user statistics
func AdminGetAdminUserStats(c *fiber.Ctx) error {
	db := config.DB

	baseQuery := db.Table("users").Where("deleted_at IS NULL AND (role IN ? OR is_super_admin = ?)", adminRoles, true)

	var totalAdmins, activeAdmins, superAdmins int64

	baseQuery.Count(&totalAdmins)

	db.Table("users").Where("deleted_at IS NULL AND (role IN ? OR is_super_admin = ?) AND active = ?", adminRoles, true, true).Count(&activeAdmins)
	db.Table("users").Where("deleted_at IS NULL AND is_super_admin = ?", true).Count(&superAdmins)

	stats := map[string]interface{}{
		"total_admins":  totalAdmins,
		"active_admins": activeAdmins,
		"super_admins":  superAdmins,
		"locked_admins": 0, // No lock mechanism currently
	}

	return utils.SendSimpleSuccess(c, stats, "Admin user statistics retrieved successfully")
}

// AdminGetAdminUser returns a specific admin user by ID
func AdminGetAdminUser(c *fiber.Ctx) error {
	db := config.DB
	userID := c.Params("id")

	var user map[string]interface{}
	err := db.Table("users").
		Select(`id, email, name, role,
			CASE WHEN active = true THEN 'active' ELSE 'suspended' END as status,
			active as is_active, is_super_admin,
			created_at, updated_at, last_login`).
		Where("id = ? AND deleted_at IS NULL AND (role IN ? OR is_super_admin = ?)", userID, adminRoles, true).
		First(&user).Error

	if err != nil {
		return utils.SendNotFound(c, "Admin user not found")
	}

	return utils.SendSimpleSuccess(c, user, "Admin user retrieved successfully")
}

// AdminCreateAdminUser creates a new admin user
func AdminCreateAdminUser(c *fiber.Ctx) error {
	db := config.DB

	var request struct {
		Email                 string   `json:"email"`
		FirstName             string   `json:"first_name"`
		LastName              string   `json:"last_name"`
		Password              string   `json:"password"`
		IsActive              bool     `json:"is_active"`
		IsSuperAdmin          bool     `json:"is_super_admin"`
		RoleIDs               []string `json:"role_ids"`
		SendWelcomeEmail      bool     `json:"send_welcome_email"`
		RequirePasswordChange bool     `json:"require_password_change"`
	}

	if err := c.BodyParser(&request); err != nil {
		return utils.SendBadRequest(c, "Invalid request body")
	}

	if request.Email == "" || request.Password == "" {
		return utils.SendBadRequest(c, "Email and password are required")
	}

	if len(request.Password) < 8 {
		return utils.SendBadRequest(c, "Password must be at least 8 characters")
	}

	// Check email uniqueness
	var emailCount int64
	db.Table("users").Where("email = ? AND deleted_at IS NULL", request.Email).Count(&emailCount)
	if emailCount > 0 {
		return utils.SendConflictError(c, "Email already in use")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		return utils.SendInternalError(c, "Failed to hash password", err)
	}

	name := strings.TrimSpace(request.FirstName + " " + request.LastName)
	if name == "" {
		name = "Admin User"
	}

	role := "admin"
	if request.IsSuperAdmin {
		role = "super_admin"
	}

	now := time.Now()
	userID := utils.GenerateID()

	user := map[string]interface{}{
		"id":             userID,
		"email":          request.Email,
		"name":           name,
		"password":       string(hashedPassword),
		"role":           role,
		"active":         request.IsActive,
		"is_super_admin": request.IsSuperAdmin,
		"created_at":     now,
		"updated_at":     now,
	}

	if err := db.Table("users").Create(user).Error; err != nil {
		log.Printf("Error creating admin user: %v", err)
		return utils.SendInternalError(c, "Failed to create admin user", err)
	}

	// Audit log
	db.Table("admin_audit_logs").Create(map[string]interface{}{
		"id":            utils.GenerateID(),
		"action":        "admin_user_created",
		"admin_user_id": c.Locals("userID"),
		"new_value":     userID,
		"created_at":    now,
	})

	// Remove password from response
	delete(user, "password")

	return utils.SendCreatedSuccess(c, user, "Admin user created successfully")
}

// AdminUpdateAdminUser updates an admin user
func AdminUpdateAdminUser(c *fiber.Ctx) error {
	db := config.DB
	userID := c.Params("id")

	var existingCount int64
	db.Table("users").Where("id = ? AND deleted_at IS NULL AND (role IN ? OR is_super_admin = ?)", userID, adminRoles, true).Count(&existingCount)
	if existingCount == 0 {
		return utils.SendNotFound(c, "Admin user not found")
	}

	var request struct {
		Email        *string  `json:"email,omitempty"`
		FirstName    *string  `json:"first_name,omitempty"`
		LastName     *string  `json:"last_name,omitempty"`
		IsActive     *bool    `json:"is_active,omitempty"`
		IsSuperAdmin *bool    `json:"is_super_admin,omitempty"`
		RoleIDs      []string `json:"role_ids,omitempty"`
	}

	if err := c.BodyParser(&request); err != nil {
		return utils.SendBadRequest(c, "Invalid request body")
	}

	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}

	if request.Email != nil {
		var emailCount int64
		db.Table("users").Where("email = ? AND id != ? AND deleted_at IS NULL", *request.Email, userID).Count(&emailCount)
		if emailCount > 0 {
			return utils.SendConflictError(c, "Email already in use")
		}
		updates["email"] = *request.Email
	}

	if request.FirstName != nil || request.LastName != nil {
		firstName := ""
		lastName := ""
		if request.FirstName != nil {
			firstName = *request.FirstName
		}
		if request.LastName != nil {
			lastName = *request.LastName
		}
		name := strings.TrimSpace(firstName + " " + lastName)
		if name != "" {
			updates["name"] = name
		}
	}

	if request.IsActive != nil {
		updates["active"] = *request.IsActive
	}

	if request.IsSuperAdmin != nil {
		updates["is_super_admin"] = *request.IsSuperAdmin
		if *request.IsSuperAdmin {
			updates["role"] = "super_admin"
		}
	}

	if err := db.Table("users").Where("id = ?", userID).Updates(updates).Error; err != nil {
		return utils.SendInternalError(c, "Failed to update admin user", err)
	}

	return utils.SendSimpleSuccess(c, map[string]interface{}{"id": userID}, "Admin user updated successfully")
}

// AdminDeleteAdminUser soft deletes an admin user
func AdminDeleteAdminUser(c *fiber.Ctx) error {
	db := config.DB
	userID := c.Params("id")
	currentUserID := c.Locals("userID").(string)

	if userID == currentUserID {
		return utils.SendBadRequest(c, "Cannot delete your own account")
	}

	now := time.Now()
	if err := db.Table("users").Where("id = ? AND (role IN ? OR is_super_admin = ?)", userID, adminRoles, true).Updates(map[string]interface{}{
		"active":     false,
		"deleted_at": now,
		"updated_at": now,
	}).Error; err != nil {
		return utils.SendInternalError(c, "Failed to delete admin user", err)
	}

	db.Table("admin_audit_logs").Create(map[string]interface{}{
		"id":            utils.GenerateID(),
		"action":        "admin_user_deleted",
		"admin_user_id": currentUserID,
		"new_value":     userID,
		"created_at":    now,
	})

	return utils.SendSimpleSuccess(c, nil, "Admin user deleted successfully")
}

// AdminActivateAdminUser activates an admin user
func AdminActivateAdminUser(c *fiber.Ctx) error {
	db := config.DB
	userID := c.Params("id")

	if err := db.Table("users").Where("id = ? AND (role IN ? OR is_super_admin = ?)", userID, adminRoles, true).Updates(map[string]interface{}{
		"active":     true,
		"updated_at": time.Now(),
	}).Error; err != nil {
		return utils.SendInternalError(c, "Failed to activate admin user", err)
	}

	return utils.SendSimpleSuccess(c, map[string]interface{}{"id": userID}, "Admin user activated successfully")
}

// AdminDeactivateAdminUser deactivates an admin user
func AdminDeactivateAdminUser(c *fiber.Ctx) error {
	db := config.DB
	userID := c.Params("id")
	currentUserID := c.Locals("userID").(string)

	if userID == currentUserID {
		return utils.SendBadRequest(c, "Cannot deactivate your own account")
	}

	if err := db.Table("users").Where("id = ? AND (role IN ? OR is_super_admin = ?)", userID, adminRoles, true).Updates(map[string]interface{}{
		"active":     false,
		"updated_at": time.Now(),
	}).Error; err != nil {
		return utils.SendInternalError(c, "Failed to deactivate admin user", err)
	}

	return utils.SendSimpleSuccess(c, map[string]interface{}{"id": userID}, "Admin user deactivated successfully")
}

// AdminUnlockAdminUser unlocks an admin user account
func AdminUnlockAdminUser(c *fiber.Ctx) error {
	db := config.DB
	userID := c.Params("id")

	// Clear any account lockouts
	db.Table("account_lockouts").Where("user_id = ? AND active = ?", userID, true).Updates(map[string]interface{}{
		"active": false,
	})

	// Re-activate user
	db.Table("users").Where("id = ?", userID).Updates(map[string]interface{}{
		"active":     true,
		"updated_at": time.Now(),
	})

	return utils.SendSimpleSuccess(c, map[string]interface{}{"id": userID}, "Admin user unlocked successfully")
}

// AdminResetAdminPassword resets an admin user's password
func AdminResetAdminPassword(c *fiber.Ctx) error {
	db := config.DB
	userID := c.Params("id")

	var request struct {
		SendEmail bool `json:"send_email"`
	}
	c.BodyParser(&request)

	var existingCount int64
	db.Table("users").Where("id = ? AND deleted_at IS NULL AND (role IN ? OR is_super_admin = ?)", userID, adminRoles, true).Count(&existingCount)
	if existingCount == 0 {
		return utils.SendNotFound(c, "Admin user not found")
	}

	tempPassword := utils.GenerateID()[:12]
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(tempPassword), bcrypt.DefaultCost)
	if err != nil {
		return utils.SendInternalError(c, "Failed to generate password", err)
	}

	db.Table("users").Where("id = ?", userID).Updates(map[string]interface{}{
		"password":   string(hashedPassword),
		"updated_at": time.Now(),
	})

	db.Table("admin_audit_logs").Create(map[string]interface{}{
		"id":            utils.GenerateID(),
		"action":        "admin_password_reset",
		"admin_user_id": c.Locals("userID"),
		"new_value":     userID,
		"created_at":    time.Now(),
	})

	response := map[string]interface{}{}
	if !request.SendEmail {
		response["temporary_password"] = tempPassword
	}

	return utils.SendSimpleSuccess(c, response, "Admin password reset successfully")
}

// AdminToggleTwoFactor toggles 2FA for an admin user (MVP stub)
func AdminToggleTwoFactor(c *fiber.Ctx) error {
	return utils.SendSimpleSuccess(c, nil, "Two-factor authentication toggle acknowledged")
}

// AdminGetAdminUserActivity returns activity for an admin user
func AdminGetAdminUserActivity(c *fiber.Ctx) error {
	db := config.DB
	userID := c.Params("id")
	limit := c.QueryInt("limit", 50)

	var activities []map[string]interface{}
	db.Table("admin_audit_logs").
		Where("admin_user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&activities)

	return utils.SendSimpleSuccess(c, activities, "Admin user activity retrieved successfully")
}

// AdminGetAdminUserSessions returns sessions for an admin user (MVP stub)
func AdminGetAdminUserSessions(c *fiber.Ctx) error {
	userID := c.Params("id")
	db := config.DB

	var sessions []map[string]interface{}
	db.Table("sessions").
		Select("id, user_id, ip_address, user_agent, created_at, expires_at").
		Where("user_id = ? AND expires_at > ?", userID, time.Now()).
		Order("created_at DESC").
		Find(&sessions)

	return utils.SendSimpleSuccess(c, sessions, "Admin user sessions retrieved successfully")
}

// AdminTerminateAdminSession terminates a specific admin session
func AdminTerminateAdminSession(c *fiber.Ctx) error {
	db := config.DB
	userID := c.Params("id")
	sessionID := c.Params("sessionId")

	db.Table("sessions").Where("id = ? AND user_id = ?", sessionID, userID).Delete(nil)

	return utils.SendSimpleSuccess(c, nil, "Admin session terminated successfully")
}

// AdminTerminateAllAdminSessions terminates all admin sessions
func AdminTerminateAllAdminSessions(c *fiber.Ctx) error {
	db := config.DB
	userID := c.Params("id")

	db.Table("sessions").Where("user_id = ?", userID).Delete(nil)

	return utils.SendSimpleSuccess(c, nil, "All admin sessions terminated successfully")
}

// AdminExportAdminUsers exports admin users (post-MVP stub)
func AdminExportAdminUsers(c *fiber.Ctx) error {
	return utils.SendNotImplementedError(c, "Admin user export is not yet implemented")
}

// AdminBulkUpdateAdminUsers bulk updates admin users (post-MVP stub)
func AdminBulkUpdateAdminUsers(c *fiber.Ctx) error {
	return utils.SendNotImplementedError(c, "Bulk admin user update is not yet implemented")
}

// AdminImpersonateAdminUser impersonates an admin user (post-MVP stub)
func AdminImpersonateAdminUser(c *fiber.Ctx) error {
	return utils.SendNotImplementedError(c, "Admin user impersonation is not yet implemented")
}
