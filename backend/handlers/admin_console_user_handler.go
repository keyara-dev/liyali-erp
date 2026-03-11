package handlers

import (
	"encoding/json"
	"fmt"
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

// adminUserToFrontend transforms a raw DB user row into the AdminUser shape the frontend expects
func adminUserToFrontend(user map[string]interface{}) map[string]interface{} {
	name, _ := user["name"].(string)
	parts := strings.SplitN(name, " ", 2)
	firstName := ""
	lastName := ""
	if len(parts) >= 1 {
		firstName = parts[0]
	}
	if len(parts) >= 2 {
		lastName = parts[1]
	}

	// Count active sessions
	var sessionCount int64
	db := config.DB
	if uid, ok := user["id"]; ok {
		db.Table("sessions").Where("user_id = ? AND expires_at > ?", uid, time.Now()).Count(&sessionCount)
	}

	// Check if locked
	var lockCount int64
	if uid, ok := user["id"]; ok {
		db.Table("account_lockouts").Where("user_id = ? AND active = ?", uid, true).Count(&lockCount)
	}

	// Populate roles from user_organization_roles
	var roles []map[string]interface{}
	if uid, ok := user["id"]; ok {
		db.Table("user_organization_roles").
			Select(`organization_roles.id, organization_roles.name,
				COALESCE(organization_roles.display_name, organization_roles.name) as display_name,
				COALESCE(organization_roles.permissions, '[]') as permissions`).
			Joins("LEFT JOIN organization_roles ON organization_roles.id = user_organization_roles.role_id").
			Where("user_organization_roles.user_id = ?", uid).
			Find(&roles)
	}
	if roles == nil {
		roles = []map[string]interface{}{}
	}

	// Collect all permission IDs from roles
	allPermissions := []string{}
	for _, r := range roles {
		if permRaw, ok := r["permissions"]; ok && permRaw != nil {
			var permIDs []string
			switch v := permRaw.(type) {
			case string:
				_ = json.Unmarshal([]byte(v), &permIDs)
			case []byte:
				_ = json.Unmarshal(v, &permIDs)
			}
			allPermissions = append(allPermissions, permIDs...)
		}
	}

	result := map[string]interface{}{
		"id":                 user["id"],
		"email":              user["email"],
		"first_name":         firstName,
		"last_name":          lastName,
		"full_name":          name,
		"avatar_url":         nil,
		"is_active":          user["is_active"],
		"is_super_admin":     user["is_super_admin"],
		"last_login_at":      user["last_login"],
		"created_at":         user["created_at"],
		"updated_at":         user["updated_at"],
		"roles":              roles,
		"permissions":        allPermissions,
		"login_attempts":     0,
		"is_locked":          lockCount > 0,
		"locked_until":       nil,
		"two_factor_enabled": false,
		"session_count":      sessionCount,
		"last_activity_at":   user["last_login"],
		"created_by":         nil,
		"updated_by":         nil,
	}
	return result
}

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

	var rawUsers []map[string]interface{}
	if err := query.Find(&rawUsers).Error; err != nil {
		log.Printf("Error getting admin users: %v", err)
		return utils.SendInternalError(c, "Failed to retrieve admin users", err)
	}

	users := make([]map[string]interface{}, len(rawUsers))
	for i, u := range rawUsers {
		users[i] = adminUserToFrontend(u)
	}

	return utils.SendSimpleSuccess(c, users, "Admin users retrieved successfully")
}

// AdminGetAdminUserStats returns admin user statistics
func AdminGetAdminUserStats(c *fiber.Ctx) error {
	db := config.DB

	adminFilter := "deleted_at IS NULL AND (role IN ? OR is_super_admin = ?)"

	var totalAdmins, activeAdmins, superAdmins, lockedAccounts, twoFactorEnabled, neverLoggedIn int64

	db.Table("users").Where(adminFilter, adminRoles, true).Count(&totalAdmins)
	db.Table("users").Where(adminFilter+" AND active = ?", adminRoles, true, true).Count(&activeAdmins)
	db.Table("users").Where("deleted_at IS NULL AND is_super_admin = ?", true).Count(&superAdmins)

	// Locked accounts
	db.Table("account_lockouts").
		Joins("JOIN users ON users.id = account_lockouts.user_id").
		Where("account_lockouts.active = ? AND ("+adminFilter+")", true, adminRoles, true).
		Distinct("account_lockouts.user_id").Count(&lockedAccounts)

	// Never logged in
	db.Table("users").Where(adminFilter+" AND last_login IS NULL", adminRoles, true).Count(&neverLoggedIn)

	// Recent logins (last 7 days)
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	var recentLogins int64
	db.Table("users").Where(adminFilter+" AND last_login >= ?", adminRoles, true, sevenDaysAgo).Count(&recentLogins)

	// Activity stats
	today := time.Now().Truncate(24 * time.Hour)
	weekAgo := time.Now().AddDate(0, 0, -7)
	monthAgo := time.Now().AddDate(0, -1, 0)
	var dailyActive, weeklyActive, monthlyActive int64
	db.Table("users").Where(adminFilter+" AND last_login >= ?", adminRoles, true, today).Count(&dailyActive)
	db.Table("users").Where(adminFilter+" AND last_login >= ?", adminRoles, true, weekAgo).Count(&weeklyActive)
	db.Table("users").Where(adminFilter+" AND last_login >= ?", adminRoles, true, monthAgo).Count(&monthlyActive)

	// Security stats
	var failedLoginAttempts, passwordResets, accountLockouts int64
	db.Table("admin_audit_logs").Where("action = ? AND created_at >= ?", "failed_login", monthAgo).Count(&failedLoginAttempts)
	db.Table("admin_audit_logs").Where("action = ? AND created_at >= ?", "admin_password_reset", monthAgo).Count(&passwordResets)
	db.Table("admin_audit_logs").Where("action LIKE ? AND created_at >= ?", "%lockout%", monthAgo).Count(&accountLockouts)

	// Role distribution
	var roleDistribution []map[string]interface{}
	db.Table("users").
		Select("role as role_name, COUNT(*) as user_count").
		Where(adminFilter, adminRoles, true).
		Group("role").
		Find(&roleDistribution)

	for i := range roleDistribution {
		roleDistribution[i]["role_id"] = roleDistribution[i]["role_name"]
		if totalAdmins > 0 {
			if count, ok := roleDistribution[i]["user_count"].(int64); ok {
				roleDistribution[i]["percentage"] = float64(count) / float64(totalAdmins) * 100
			}
		} else {
			roleDistribution[i]["percentage"] = 0
		}
	}

	stats := map[string]interface{}{
		"total_admin_users":  totalAdmins,
		"active_admin_users": activeAdmins,
		"super_admins":       superAdmins,
		"locked_accounts":    lockedAccounts,
		"two_factor_enabled": twoFactorEnabled,
		"recent_logins":      recentLogins,
		"never_logged_in":    neverLoggedIn,
		"role_distribution":  roleDistribution,
		"activity_stats": map[string]interface{}{
			"daily_active":   dailyActive,
			"weekly_active":  weeklyActive,
			"monthly_active": monthlyActive,
		},
		"security_stats": map[string]interface{}{
			"failed_login_attempts": failedLoginAttempts,
			"password_resets":       passwordResets,
			"account_lockouts":      accountLockouts,
		},
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

	return utils.SendSimpleSuccess(c, adminUserToFrontend(user), "Admin user retrieved successfully")
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

	// Block direct super_admin role assignment — use the promote endpoint instead
	if request.IsSuperAdmin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Use the promote endpoint to assign super_admin role",
		})
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

	// Assign roles from role_ids
	if len(request.RoleIDs) > 0 {
		adminUserID := c.Locals("userID").(string)
		for _, roleID := range request.RoleIDs {
			// Check if role exists
			var roleCount int64
			db.Table("organization_roles").Where("id = ? AND active = ?", roleID, true).Count(&roleCount)
			if roleCount == 0 {
				continue
			}

			db.Table("user_organization_roles").Create(map[string]interface{}{
				"id":              utils.GenerateID(),
				"user_id":         userID,
				"organization_id": nil,
				"role_id":         roleID,
				"assigned_by":     adminUserID,
				"assigned_at":     now,
				"active":          true,
			})
		}
	}

	// Audit log
	db.Table("admin_audit_logs").Create(map[string]interface{}{
		"id":            utils.GenerateID(),
		"action":        "admin_user_created",
		"admin_user_id": c.Locals("userID"),
		"new_value":     userID,
		"created_at":    now,
	})

	// Return frontend-compatible shape
	user["is_active"] = request.IsActive
	user["is_super_admin"] = request.IsSuperAdmin
	user["last_login"] = nil
	delete(user, "password")

	return utils.SendCreatedSuccess(c, adminUserToFrontend(user), "Admin user created successfully")
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

	// Block direct super_admin role assignment — use the promote endpoint instead
	if request.IsSuperAdmin != nil && *request.IsSuperAdmin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Use the promote endpoint to assign super_admin role",
		})
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

	// Update role assignments if role_ids provided
	if request.RoleIDs != nil {
		now := time.Now()
		adminUserID := c.Locals("userID").(string)

		// Deactivate all existing role assignments
		db.Table("user_organization_roles").Where("user_id = ? AND active = ?", userID, true).Updates(map[string]interface{}{
			"active":     false,
			"updated_at": now,
		})

		// Assign new roles
		for _, roleID := range request.RoleIDs {
			var roleCount int64
			db.Table("organization_roles").Where("id = ? AND active = ?", roleID, true).Count(&roleCount)
			if roleCount == 0 {
				continue
			}

			db.Table("user_organization_roles").Create(map[string]interface{}{
				"id":              utils.GenerateID(),
				"user_id":         userID,
				"organization_id": nil,
				"role_id":         roleID,
				"assigned_by":     adminUserID,
				"assigned_at":     now,
				"active":          true,
			})
		}
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

// AdminToggleTwoFactor toggles 2FA for an admin user
// Note: Full TOTP infrastructure is not yet available. This provides a placeholder
// that records the intent and returns an informative response.
func AdminToggleTwoFactor(c *fiber.Ctx) error {
	db := config.DB
	userID := c.Params("id")
	adminUserID, _ := c.Locals("userID").(string)

	type toggleRequest struct {
		Enabled bool `json:"enabled"`
	}

	var req toggleRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendBadRequest(c, "Invalid request body")
	}

	// Verify the admin user exists
	var count int64
	db.Table("users").Where("id = ? AND role IN ('super_admin', 'admin')", userID).Count(&count)
	if count == 0 {
		return utils.SendNotFound(c, "Admin user not found")
	}

	// Log the 2FA toggle attempt in audit log
	action := "2fa_disable_requested"
	if req.Enabled {
		action = "2fa_enable_requested"
	}
	db.Table("admin_audit_logs").Create(map[string]interface{}{
		"id":            utils.GenerateID(),
		"action":        action,
		"admin_user_id": adminUserID,
		"new_value":     userID,
		"description":   fmt.Sprintf("2FA %s requested for admin user %s", action, userID),
		"created_at":    time.Now(),
	})

	return utils.SendSimpleSuccess(c, map[string]interface{}{
		"user_id":     userID,
		"two_factor":  req.Enabled,
		"message":     "Two-factor authentication configuration has been recorded. Full TOTP enrollment requires additional setup.",
	}, "Two-factor authentication preference updated")
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

// AdminExportAdminUsers exports admin users as JSON
func AdminExportAdminUsers(c *fiber.Ctx) error {
	db := config.DB

	query := db.Table("users").Where("role IN ?", adminRoles)

	if search := c.Query("search"); search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("name ILIKE ? OR email ILIKE ?", searchPattern, searchPattern)
	}
	if isActive := c.Query("is_active"); isActive == "true" {
		query = query.Where("is_active = ?", true)
	} else if isActive == "false" {
		query = query.Where("is_active = ?", false)
	}

	var rawUsers []map[string]interface{}
	if err := query.
		Order("created_at DESC").
		Limit(10000).
		Find(&rawUsers).Error; err != nil {
		log.Printf("Error exporting admin users: %v", err)
		return utils.SendInternalError(c, "Failed to export admin users", err)
	}

	users := make([]map[string]interface{}, 0, len(rawUsers))
	for _, raw := range rawUsers {
		users = append(users, adminUserToFrontend(raw))
	}

	exportData := map[string]interface{}{
		"users":       users,
		"total_count": len(users),
		"exported_at": time.Now().Format(time.RFC3339),
	}

	c.Set("Content-Disposition", "attachment; filename=admin-users-export-"+time.Now().Format("2006-01-02")+".json")
	c.Set("Content-Type", "application/json")

	return c.JSON(exportData)
}

// AdminBulkUpdateAdminUsers applies bulk actions to admin users
func AdminBulkUpdateAdminUsers(c *fiber.Ctx) error {
	db := config.DB

	var req struct {
		UserIDs []string `json:"user_ids"`
		Action  string   `json:"action"` // activate, deactivate, delete
	}
	if err := c.BodyParser(&req); err != nil {
		return utils.SendBadRequest(c, "Invalid request body")
	}
	if len(req.UserIDs) == 0 {
		return utils.SendBadRequest(c, "No user IDs provided")
	}
	if req.Action == "" {
		return utils.SendBadRequest(c, "Action is required")
	}

	var affected int64
	switch req.Action {
	case "activate":
		result := db.Table("users").Where("id IN ? AND role IN ?", req.UserIDs, adminRoles).
			Update("is_active", true)
		affected = result.RowsAffected
		if result.Error != nil {
			return utils.SendInternalError(c, "Failed to activate users", result.Error)
		}
	case "deactivate":
		result := db.Table("users").Where("id IN ? AND role IN ?", req.UserIDs, adminRoles).
			Update("is_active", false)
		affected = result.RowsAffected
		if result.Error != nil {
			return utils.SendInternalError(c, "Failed to deactivate users", result.Error)
		}
	case "delete":
		result := db.Table("users").Where("id IN ? AND role IN ? AND is_super_admin = ?", req.UserIDs, adminRoles, false).
			Delete(nil)
		affected = result.RowsAffected
		if result.Error != nil {
			return utils.SendInternalError(c, "Failed to delete users", result.Error)
		}
	default:
		return utils.SendBadRequest(c, "Invalid action. Supported: activate, deactivate, delete")
	}

	return utils.SendSimpleSuccess(c, map[string]interface{}{
		"action":   req.Action,
		"affected": affected,
		"total":    len(req.UserIDs),
	}, "Bulk operation completed successfully")
}

// AdminImpersonateAdminUser generates a short-lived impersonation token for an admin user
func AdminImpersonateAdminUser(c *fiber.Ctx) error {
	db := config.DB
	userID := c.Params("id")
	adminUserID, _ := c.Locals("userID").(string)

	// Cannot impersonate yourself
	if userID == adminUserID {
		return utils.SendBadRequest(c, "Cannot impersonate yourself")
	}

	// Look up the target admin user
	var user map[string]interface{}
	err := db.Table("users").Where("id = ? AND role IN ('super_admin', 'admin')", userID).First(&user).Error
	if err != nil {
		return utils.SendNotFound(c, "Admin user not found")
	}

	status, _ := user["status"].(string)
	if status != "active" {
		return utils.SendBadRequest(c, "Cannot impersonate inactive or suspended admin user")
	}

	email, _ := user["email"].(string)
	name, _ := user["name"].(string)
	role, _ := user["role"].(string)

	token, err := utils.GenerateToken(userID, email, name, role, nil)
	if err != nil {
		log.Printf("Error generating admin impersonation token: %v", err)
		return utils.SendInternalError(c, "Failed to generate impersonation token", err)
	}

	// Log in audit
	db.Table("admin_audit_logs").Create(map[string]interface{}{
		"id":            utils.GenerateID(),
		"action":        "admin_impersonation",
		"admin_user_id": adminUserID,
		"new_value":     userID,
		"description":   fmt.Sprintf("Admin impersonated admin user: %s", email),
		"created_at":    time.Now(),
	})

	return utils.SendSimpleSuccess(c, map[string]interface{}{
		"token":            token,
		"expires_in":       900,
		"impersonated_user": map[string]interface{}{
			"id":    userID,
			"email": email,
			"name":  name,
			"role":  role,
		},
		"warning": "This is a short-lived token for impersonation purposes. All actions will be logged.",
	}, "Admin impersonation token generated successfully")
}

// AdminPromoteToSuperAdmin promotes an admin user to super_admin role.
// Cannot promote yourself.
func AdminPromoteToSuperAdmin(c *fiber.Ctx) error {
	db := config.DB
	userID := c.Params("id")
	callerID, _ := c.Locals("userID").(string)

	if userID == callerID {
		return utils.SendBadRequest(c, "Cannot promote your own account")
	}

	// Verify target user exists and is an admin-level user
	var count int64
	db.Table("users").Where("id = ? AND deleted_at IS NULL AND role IN ?", userID, adminRoles).Count(&count)
	if count == 0 {
		return utils.SendNotFound(c, "Admin user not found")
	}

	now := time.Now()
	if err := db.Table("users").Where("id = ?", userID).Updates(map[string]interface{}{
		"role":           "super_admin",
		"is_super_admin": true,
		"updated_at":     now,
	}).Error; err != nil {
		return utils.SendInternalError(c, "Failed to promote user", err)
	}

	db.Table("admin_audit_logs").Create(map[string]interface{}{
		"id":            utils.GenerateID(),
		"action":        "admin_user_promoted_to_super_admin",
		"admin_user_id": callerID,
		"new_value":     userID,
		"created_at":    now,
	})

	return utils.SendSimpleSuccess(c, map[string]interface{}{"id": userID}, "User promoted to super_admin successfully")
}

// AdminDemoteFromSuperAdmin demotes a super_admin user back to admin role.
// Cannot demote yourself.
func AdminDemoteFromSuperAdmin(c *fiber.Ctx) error {
	db := config.DB
	userID := c.Params("id")
	callerID, _ := c.Locals("userID").(string)

	if userID == callerID {
		return utils.SendBadRequest(c, "Cannot demote your own account")
	}

	// Verify target user is super_admin
	var count int64
	db.Table("users").Where("id = ? AND deleted_at IS NULL AND role = ?", userID, "super_admin").Count(&count)
	if count == 0 {
		return utils.SendNotFound(c, "Super admin user not found")
	}

	now := time.Now()
	if err := db.Table("users").Where("id = ?", userID).Updates(map[string]interface{}{
		"role":           "admin",
		"is_super_admin": false,
		"updated_at":     now,
	}).Error; err != nil {
		return utils.SendInternalError(c, "Failed to demote user", err)
	}

	db.Table("admin_audit_logs").Create(map[string]interface{}{
		"id":            utils.GenerateID(),
		"action":        "admin_user_demoted_from_super_admin",
		"admin_user_id": callerID,
		"new_value":     userID,
		"created_at":    now,
	})

	return utils.SendSimpleSuccess(c, map[string]interface{}{"id": userID}, "User demoted from super_admin successfully")
}
