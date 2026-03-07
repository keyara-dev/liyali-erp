package handlers

import (
	"log"
	"math"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/utils"
	"golang.org/x/crypto/bcrypt"
)

// platformUserEnrich adds missing fields that the frontend PlatformUser interface expects
func platformUserEnrich(user map[string]interface{}) map[string]interface{} {
	db := config.DB
	user["email_verified"] = true
	user["login_count"] = 0
	user["phone"] = nil
	user["profile"] = nil

	if uid, ok := user["id"]; ok {
		var orgs []map[string]interface{}
		db.Table("organization_members").
			Select(`organization_members.organization_id,
				organizations.name as organization_name,
				COALESCE(organizations.slug, '') as organization_domain,
				organization_members.role,
				CASE WHEN organization_members.active = true THEN 'active' ELSE 'suspended' END as status,
				organization_members.joined_at`).
			Joins("LEFT JOIN organizations ON organizations.id = organization_members.organization_id").
			Where("organization_members.user_id = ?", uid).
			Find(&orgs)

		for i := range orgs {
			orgs[i]["permissions"] = []string{}
			orgs[i]["is_primary"] = i == 0
		}
		user["organizations"] = orgs
	}

	return user
}

// AdminGetAllUsers returns all platform users with filters and pagination
func AdminGetAllUsers(c *fiber.Ctx) error {
	db := config.DB

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	search := c.Query("search")
	status := c.Query("status")
	role := c.Query("role")
	organizationID := c.Query("organization_id")
	sortBy := c.Query("sort_by", "created_at")
	sortOrder := c.Query("sort_order", "desc")

	page, limit = utils.NormalizePaginationParams(page, limit)
	offset := (page - 1) * limit

	query := db.Table("users").
		Select(`users.id, users.email, users.name, users.role,
			CASE WHEN users.active = true THEN 'active' WHEN users.active = false AND users.last_login IS NULL THEN 'pending' ELSE 'suspended' END as status,
			users.created_at, users.updated_at, users.last_login,
			(SELECT COUNT(*) FROM organization_members WHERE organization_members.user_id = users.id AND organization_members.active = true) as organization_count`).
		Where("users.deleted_at IS NULL")

	countQuery := db.Table("users").Where("deleted_at IS NULL")

	if search != "" {
		searchTerm := "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(users.name) LIKE ? OR LOWER(users.email) LIKE ?", searchTerm, searchTerm)
		countQuery = countQuery.Where("LOWER(name) LIKE ? OR LOWER(email) LIKE ?", searchTerm, searchTerm)
	}

	if status != "" && status != "all" {
		if status == "active" {
			query = query.Where("users.active = ?", true)
			countQuery = countQuery.Where("active = ?", true)
		} else if status == "suspended" || status == "inactive" {
			query = query.Where("users.active = ?", false)
			countQuery = countQuery.Where("active = ?", false)
		}
	}

	if role != "" {
		query = query.Where("users.role = ?", role)
		countQuery = countQuery.Where("role = ?", role)
	}

	if organizationID != "" {
		query = query.Where("users.id IN (SELECT user_id FROM organization_members WHERE organization_id = ? AND active = true)", organizationID)
		countQuery = countQuery.Where("id IN (SELECT user_id FROM organization_members WHERE organization_id = ? AND active = true)", organizationID)
	}

	var total int64
	countQuery.Count(&total)

	allowedSorts := map[string]string{
		"name":       "users.name",
		"email":      "users.email",
		"created_at": "users.created_at",
		"last_login": "users.last_login",
	}
	sortCol, ok := allowedSorts[sortBy]
	if !ok {
		sortCol = "users.created_at"
	}
	if sortOrder != "asc" {
		sortOrder = "desc"
	}
	query = query.Order(sortCol + " " + sortOrder)

	var rawUsers []map[string]interface{}
	if err := query.Offset(offset).Limit(limit).Find(&rawUsers).Error; err != nil {
		log.Printf("Error getting users: %v", err)
		return utils.SendInternalError(c, "Failed to retrieve users", err)
	}

	users := make([]map[string]interface{}, len(rawUsers))
	for i, u := range rawUsers {
		users[i] = platformUserEnrich(u)
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	response := map[string]interface{}{
		"users":      users,
		"total":      total,
		"page":       page,
		"limit":      limit,
		"totalPages": totalPages,
	}

	return utils.SendSimpleSuccess(c, response, "Users retrieved successfully")
}

// AdminGetUserStatistics returns user statistics for admin dashboard
func AdminGetUserStatistics(c *fiber.Ctx) error {
	db := config.DB

	var totalUsers, activeUsers, suspendedUsers, createdThisMonth, loggedInToday int64

	db.Table("users").Where("deleted_at IS NULL").Count(&totalUsers)
	db.Table("users").Where("deleted_at IS NULL AND active = ?", true).Count(&activeUsers)
	db.Table("users").Where("deleted_at IS NULL AND active = ?", false).Count(&suspendedUsers)

	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	db.Table("users").Where("deleted_at IS NULL AND created_at >= ?", thirtyDaysAgo).Count(&createdThisMonth)

	today := time.Now().Truncate(24 * time.Hour)
	db.Table("users").Where("deleted_at IS NULL AND last_login >= ?", today).Count(&loggedInToday)

	var topOrgs []map[string]interface{}
	db.Table("organization_members").
		Select("organization_id, COUNT(*) as user_count").
		Where("active = ?", true).
		Group("organization_id").
		Order("user_count DESC").
		Limit(5).
		Find(&topOrgs)

	for i := range topOrgs {
		if orgID, ok := topOrgs[i]["organization_id"].(string); ok {
			var orgName string
			db.Table("organizations").Where("id = ?", orgID).Pluck("name", &orgName)
			topOrgs[i]["organization_name"] = orgName
		}
	}

	stats := map[string]interface{}{
		"total_users":                totalUsers,
		"active_users":               activeUsers,
		"suspended_users":            suspendedUsers,
		"pending_users":              0,
		"users_created_this_month":   createdThisMonth,
		"users_logged_in_today":      loggedInToday,
		"top_organizations_by_users": topOrgs,
	}

	return utils.SendSimpleSuccess(c, stats, "User statistics retrieved successfully")
}

// AdminGetUserById returns a user by ID with organization memberships
func AdminGetUserById(c *fiber.Ctx) error {
	db := config.DB
	userID := c.Params("id")

	var user map[string]interface{}
	err := db.Table("users").
		Select(`id, email, name, role,
			CASE WHEN active = true THEN 'active' ELSE 'suspended' END as status,
			created_at, updated_at, last_login, is_super_admin`).
		Where("id = ? AND deleted_at IS NULL", userID).
		First(&user).Error

	if err != nil {
		return utils.SendNotFound(c, "User not found")
	}

	var orgs []map[string]interface{}
	db.Table("organization_members").
		Select(`organization_members.organization_id,
			organizations.name as organization_name,
			COALESCE(organizations.slug, '') as organization_domain,
			organization_members.role,
			CASE WHEN organization_members.active = true THEN 'active' ELSE 'suspended' END as status,
			organization_members.joined_at`).
		Joins("LEFT JOIN organizations ON organizations.id = organization_members.organization_id").
		Where("organization_members.user_id = ?", userID).
		Find(&orgs)

	for i := range orgs {
		orgs[i]["permissions"] = []string{}
		orgs[i]["is_primary"] = i == 0
	}

	user["organizations"] = orgs
	user["email_verified"] = true
	user["login_count"] = 0
	user["phone"] = nil
	user["profile"] = nil

	return utils.SendSimpleSuccess(c, user, "User retrieved successfully")
}

// AdminUpdateUser updates a platform user's profile
func AdminUpdateUser(c *fiber.Ctx) error {
	db := config.DB
	userID := c.Params("id")

	var existingCount int64
	db.Table("users").Where("id = ? AND deleted_at IS NULL", userID).Count(&existingCount)
	if existingCount == 0 {
		return utils.SendNotFound(c, "User not found")
	}

	var request struct {
		Name      *string `json:"name,omitempty"`
		Email     *string `json:"email,omitempty"`
		Role      *string `json:"role,omitempty"`
		Status    *string `json:"status,omitempty"`
		Position  *string `json:"position,omitempty"`
		ManNumber *string `json:"manNumber,omitempty"`
		NrcNumber *string `json:"nrcNumber,omitempty"`
		Contact   *string `json:"contact,omitempty"`
	}

	if err := c.BodyParser(&request); err != nil {
		return utils.SendBadRequest(c, "Invalid request body")
	}

	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}

	if request.Name != nil {
		updates["name"] = *request.Name
	}
	if request.Email != nil {
		var emailCount int64
		db.Table("users").Where("email = ? AND id != ? AND deleted_at IS NULL", *request.Email, userID).Count(&emailCount)
		if emailCount > 0 {
			return utils.SendConflictError(c, "Email already in use")
		}
		updates["email"] = *request.Email
	}
	if request.Role != nil {
		updates["role"] = *request.Role
	}
	if request.Status != nil {
		updates["active"] = *request.Status == "active"
	}
	if request.Position != nil {
		updates["position"] = *request.Position
	}
	if request.ManNumber != nil {
		updates["man_number"] = *request.ManNumber
	}
	if request.NrcNumber != nil {
		updates["nrc_number"] = *request.NrcNumber
	}
	if request.Contact != nil {
		updates["contact"] = *request.Contact
	}

	if err := db.Table("users").Where("id = ?", userID).Updates(updates).Error; err != nil {
		return utils.SendInternalError(c, "Failed to update user", err)
	}

	db.Table("admin_audit_logs").Create(map[string]interface{}{
		"id":            utils.GenerateID(),
		"action":        "user_updated",
		"admin_user_id": c.Locals("userID"),
		"new_value":     userID,
		"created_at":    time.Now(),
	})

	return utils.SendSimpleSuccess(c, map[string]interface{}{"id": userID}, "User updated successfully")
}

// AdminUpdateUserStatus updates a platform user's status
func AdminUpdateUserStatus(c *fiber.Ctx) error {
	db := config.DB
	userID := c.Params("id")

	var request struct {
		Status string `json:"status"`
		Reason string `json:"reason"`
	}

	if err := c.BodyParser(&request); err != nil {
		return utils.SendBadRequest(c, "Invalid request body")
	}

	if request.Status != "active" && request.Status != "suspended" && request.Status != "inactive" {
		return utils.SendBadRequest(c, "Invalid status")
	}

	active := request.Status == "active"

	if err := db.Table("users").Where("id = ? AND deleted_at IS NULL", userID).Updates(map[string]interface{}{
		"active":     active,
		"updated_at": time.Now(),
	}).Error; err != nil {
		return utils.SendInternalError(c, "Failed to update user status", err)
	}

	db.Table("admin_audit_logs").Create(map[string]interface{}{
		"id":            utils.GenerateID(),
		"action":        "user_status_changed",
		"admin_user_id": c.Locals("userID"),
		"new_value":     request.Status,
		"reason":        request.Reason,
		"created_at":    time.Now(),
	})

	return utils.SendSimpleSuccess(c, map[string]interface{}{
		"id":     userID,
		"status": request.Status,
	}, "User status updated successfully")
}

// AdminGetUserActivity returns activity logs for a user
func AdminGetUserActivity(c *fiber.Ctx) error {
	db := config.DB
	userID := c.Params("id")

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)
	page, limit = utils.NormalizePaginationParams(page, limit)
	offset := (page - 1) * limit

	var total int64
	db.Table("admin_audit_logs").Where("admin_user_id = ?", userID).Count(&total)

	var activities []map[string]interface{}
	db.Table("admin_audit_logs").
		Where("admin_user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&activities)

	for i := range activities {
		activities[i]["timestamp"] = activities[i]["created_at"]
		if _, ok := activities[i]["description"]; !ok {
			activities[i]["description"] = activities[i]["action"]
		}
	}

	response := map[string]interface{}{
		"activities": activities,
		"total":      total,
		"page":       page,
		"limit":      limit,
	}

	return utils.SendSimpleSuccess(c, response, "User activity retrieved successfully")
}

// AdminGetUserSessions returns active sessions for a user
func AdminGetUserSessions(c *fiber.Ctx) error {
	userID := c.Params("id")
	db := config.DB

	var sessions []map[string]interface{}
	db.Table("sessions").
		Select("id, user_id, ip_address, user_agent, created_at, expires_at").
		Where("user_id = ? AND expires_at > ?", userID, time.Now()).
		Order("created_at DESC").
		Find(&sessions)

	return utils.SendSimpleSuccess(c, sessions, "User sessions retrieved successfully")
}

// AdminTerminateUserSession terminates a specific user session
func AdminTerminateUserSession(c *fiber.Ctx) error {
	db := config.DB
	userID := c.Params("id")
	sessionID := c.Params("sessionId")

	db.Table("sessions").Where("id = ? AND user_id = ?", sessionID, userID).Delete(nil)

	return utils.SendSimpleSuccess(c, nil, "Session terminated successfully")
}

// AdminTerminateAllUserSessions terminates all sessions for a user
func AdminTerminateAllUserSessions(c *fiber.Ctx) error {
	db := config.DB
	userID := c.Params("id")

	db.Table("sessions").Where("user_id = ?", userID).Delete(nil)

	return utils.SendSimpleSuccess(c, nil, "All sessions terminated successfully")
}

// AdminResetUserPassword resets a user's password
func AdminResetUserPassword(c *fiber.Ctx) error {
	db := config.DB
	userID := c.Params("id")

	var request struct {
		SendEmail bool `json:"send_email"`
	}
	c.BodyParser(&request)

	var existingCount int64
	db.Table("users").Where("id = ? AND deleted_at IS NULL", userID).Count(&existingCount)
	if existingCount == 0 {
		return utils.SendNotFound(c, "User not found")
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
		"action":        "user_password_reset",
		"admin_user_id": c.Locals("userID"),
		"new_value":     userID,
		"created_at":    time.Now(),
	})

	response := map[string]interface{}{}
	if !request.SendEmail {
		response["temporary_password"] = tempPassword
	}

	return utils.SendSimpleSuccess(c, response, "Password reset successfully")
}

// AdminImpersonateUser generates a short-lived impersonation token for a platform user
func AdminImpersonateUser(c *fiber.Ctx) error {
	db := config.DB
	userID := c.Params("id")
	adminUserID, _ := c.Locals("userID").(string)

	// Look up the target user
	var user map[string]interface{}
	err := db.Table("users").Where("id = ?", userID).First(&user).Error
	if err != nil {
		return utils.SendNotFound(c, "User not found")
	}

	// Check user is active
	status, _ := user["status"].(string)
	if status != "active" {
		return utils.SendBadRequest(c, "Cannot impersonate inactive or suspended user")
	}

	email, _ := user["email"].(string)
	name, _ := user["name"].(string)
	role, _ := user["role"].(string)
	if role == "" {
		role = "user"
	}

	// Generate a short-lived token (15 minutes) for impersonation
	token, err := utils.GenerateToken(userID, email, name, role, nil)
	if err != nil {
		log.Printf("Error generating impersonation token: %v", err)
		return utils.SendInternalError(c, "Failed to generate impersonation token", err)
	}

	// Log the impersonation in audit log
	db.Table("admin_audit_logs").Create(map[string]interface{}{
		"id":            utils.GenerateID(),
		"action":        "user_impersonation",
		"admin_user_id": adminUserID,
		"new_value":     userID,
		"description":   "Admin impersonated platform user: " + email,
		"created_at":    time.Now(),
	})

	return utils.SendSimpleSuccess(c, map[string]interface{}{
		"token":            token,
		"expires_in":       900, // 15 minutes in seconds
		"impersonated_user": map[string]interface{}{
			"id":    userID,
			"email": email,
			"name":  name,
		},
		"warning": "This is a short-lived token for impersonation purposes. All actions will be logged.",
	}, "Impersonation token generated successfully")
}

// AdminGetUserOrganizations returns organizations a user belongs to
func AdminGetUserOrganizations(c *fiber.Ctx) error {
	db := config.DB
	userID := c.Params("id")

	var orgs []map[string]interface{}
	err := db.Table("organization_members").
		Select(`organization_members.organization_id,
			organizations.name as organization_name,
			COALESCE(organizations.slug, '') as organization_domain,
			organization_members.role,
			CASE WHEN organization_members.active = true THEN 'active' ELSE 'suspended' END as status,
			organization_members.joined_at`).
		Joins("LEFT JOIN organizations ON organizations.id = organization_members.organization_id").
		Where("organization_members.user_id = ?", userID).
		Find(&orgs).Error

	if err != nil {
		return utils.SendInternalError(c, "Failed to retrieve user organizations", err)
	}

	for i := range orgs {
		orgs[i]["permissions"] = []string{}
		orgs[i]["is_primary"] = i == 0
	}

	return utils.SendSimpleSuccess(c, orgs, "User organizations retrieved successfully")
}

// AdminUpdateUserOrgRole updates a user's role in an organization
func AdminUpdateUserOrgRole(c *fiber.Ctx) error {
	db := config.DB
	userID := c.Params("id")
	orgID := c.Params("orgId")

	var request struct {
		Role   string  `json:"role"`
		Status *string `json:"status,omitempty"`
	}

	if err := c.BodyParser(&request); err != nil {
		return utils.SendBadRequest(c, "Invalid request body")
	}

	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}
	if request.Role != "" {
		updates["role"] = request.Role
	}
	if request.Status != nil {
		updates["active"] = *request.Status == "active"
	}

	result := db.Table("organization_members").
		Where("user_id = ? AND organization_id = ?", userID, orgID).
		Updates(updates)

	if result.RowsAffected == 0 {
		return utils.SendNotFound(c, "User membership not found in this organization")
	}

	return utils.SendSimpleSuccess(c, map[string]interface{}{
		"user_id":         userID,
		"organization_id": orgID,
	}, "User organization role updated successfully")
}

// AdminRemoveUserFromOrg removes a user from an organization
func AdminRemoveUserFromOrg(c *fiber.Ctx) error {
	db := config.DB
	userID := c.Params("id")
	orgID := c.Params("orgId")

	result := db.Table("organization_members").
		Where("user_id = ? AND organization_id = ?", userID, orgID).
		Delete(nil)

	if result.RowsAffected == 0 {
		return utils.SendNotFound(c, "User membership not found in this organization")
	}

	db.Table("admin_audit_logs").Create(map[string]interface{}{
		"id":              utils.GenerateID(),
		"organization_id": orgID,
		"action":          "user_removed_from_org",
		"admin_user_id":   c.Locals("userID"),
		"new_value":       userID,
		"created_at":      time.Now(),
	})

	return utils.SendSimpleSuccess(c, nil, "User removed from organization successfully")
}
