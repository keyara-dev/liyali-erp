package handlers

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
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
			users.active as is_active,
			users.created_at, users.updated_at, users.last_login,
			COALESCE(users.position, '') as position,
			COALESCE(users.man_number, '') as man_number,
			COALESCE(users.nrc_number, '') as nrc_number,
			COALESCE(users.contact, '') as contact,
			COALESCE(users.contact, '') as phone,
			COALESCE(users.mfa_enabled, false) as mfa_enabled,
			users.preferences,
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
			active as is_active,
			created_at, updated_at, last_login, is_super_admin,
			COALESCE(position, '') as position,
			COALESCE(man_number, '') as man_number,
			COALESCE(nrc_number, '') as nrc_number,
			COALESCE(contact, '') as contact,
			COALESCE(mfa_enabled, false) as mfa_enabled,
			preferences`).
		Where("id = ? AND deleted_at IS NULL", userID).
		Limit(1).Scan(&user).Error

	if err != nil || len(user) == 0 {
		return utils.SendNotFound(c, "User not found")
	}

	var orgs []map[string]interface{}
	db.Table("organization_members").
		Select(`organization_members.organization_id,
			organizations.name as organization_name,
			COALESCE(organizations.slug, '') as organization_domain,
			organization_members.role,
			COALESCE(organization_members.department, '') as department,
			CASE WHEN organization_members.active = true THEN 'active' ELSE 'suspended' END as status,
			organization_members.joined_at`).
		Joins("LEFT JOIN organizations ON organizations.id = organization_members.organization_id").
		Where("organization_members.user_id = ?", userID).
		Find(&orgs)

	// Pull department from primary org membership
	department := ""
	for i := range orgs {
		orgs[i]["permissions"] = []string{}
		orgs[i]["is_primary"] = i == 0
		if i == 0 {
			if d, ok := orgs[i]["department"].(string); ok {
				department = d
			}
		}
	}

	user["organizations"] = orgs
	user["department"] = department
	// mfa_enabled is now queried directly from the users table (column added in migration 024)
	// Set defaults only if not already populated by the SELECT
	if _, ok := user["mfa_enabled"]; !ok {
		user["mfa_enabled"] = false
	}
	user["email_verified"] = true
	user["login_count"] = 0
	// phone comes from contact field (same column)
	if phone, ok := user["contact"].(string); ok {
		user["phone"] = phone
	}

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

// AdminGetUserActivity returns activity logs for a user (merges user_activity_logs + admin_audit_logs)
func AdminGetUserActivity(c *fiber.Ctx) error {
	db := config.DB
	userID := c.Params("id")

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)
	page, limit = utils.NormalizePaginationParams(page, limit)
	offset := (page - 1) * limit

	actionType := c.Query("action_type")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	// Build conditions for user_activity_logs
	ualConds := "user_id = ?"
	ualArgs := []interface{}{userID}
	if actionType != "" {
		ualConds += " AND action_type = ?"
		ualArgs = append(ualArgs, actionType)
	}
	if startDateStr != "" {
		if t, err := time.Parse("2006-01-02", startDateStr); err == nil {
			ualConds += " AND created_at >= ?"
			ualArgs = append(ualArgs, t)
		}
	}
	if endDateStr != "" {
		if t, err := time.Parse("2006-01-02", endDateStr); err == nil {
			ualConds += " AND created_at <= ?"
			ualArgs = append(ualArgs, t.Add(24*time.Hour))
		}
	}

	// Count from user_activity_logs
	var ualTotal int64
	db.Table("user_activity_logs").Where(ualConds, ualArgs...).Count(&ualTotal)

	// Fetch from user_activity_logs
	type activityRow struct {
		ID           string    `json:"id"`
		ActionType   string    `json:"action_type"`
		ResourceType string    `json:"resource_type"`
		ResourceID   string    `json:"resource_id"`
		IPAddress    string    `json:"ip_address"`
		UserAgent    string    `json:"user_agent"`
		CreatedAt    time.Time `json:"created_at"`
		Source       string    `json:"source"`
	}

	var ualRows []activityRow
	db.Table("user_activity_logs").
		Select("id::text, action_type, COALESCE(resource_type,'') as resource_type, COALESCE(resource_id,'') as resource_id, COALESCE(ip_address,'') as ip_address, COALESCE(user_agent,'') as user_agent, created_at, 'activity' as source").
		Where(ualConds, ualArgs...).
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Scan(&ualRows)

	// Also fetch admin_audit_logs for backward compat (only when no action_type filter or it could match)
	var auditRows []map[string]interface{}
	if actionType == "" {
		auditQ := db.Table("admin_audit_logs").Where("admin_user_id = ?", userID)
		if startDateStr != "" {
			if t, err := time.Parse("2006-01-02", startDateStr); err == nil {
				auditQ = auditQ.Where("created_at >= ?", t)
			}
		}
		if endDateStr != "" {
			if t, err := time.Parse("2006-01-02", endDateStr); err == nil {
				auditQ = auditQ.Where("created_at <= ?", t.Add(24*time.Hour))
			}
		}
		auditQ.Select("id::text, action, COALESCE(description,'') as description, created_at").
			Order("created_at DESC").Limit(10).Scan(&auditRows)
	}

	// Normalize ualRows to maps
	activities := make([]map[string]interface{}, 0, len(ualRows))
	for _, r := range ualRows {
		activities = append(activities, map[string]interface{}{
			"id":            r.ID,
			"action_type":   r.ActionType,
			"resource_type": r.ResourceType,
			"resource_id":   r.ResourceID,
			"ip_address":    r.IPAddress,
			"user_agent":    r.UserAgent,
			"created_at":    r.CreatedAt,
			"source":        "activity",
		})
	}

	// Normalize audit rows
	for _, r := range auditRows {
		action, _ := r["action"].(string)
		desc, _ := r["description"].(string)
		if desc == "" {
			desc = action
		}
		activities = append(activities, map[string]interface{}{
			"id":            r["id"],
			"action_type":   action,
			"resource_type": "admin_action",
			"description":   desc,
			"created_at":    r["created_at"],
			"source":        "admin_audit",
		})
	}

	totalPages := int(math.Ceil(float64(ualTotal) / float64(limit)))

	response := map[string]interface{}{
		"activities": activities,
		"pagination": map[string]interface{}{
			"total_records": ualTotal,
			"total_pages":   totalPages,
			"current_page":  page,
			"has_next":      page < totalPages,
			"has_prev":      page > 1,
		},
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
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&sessions)

	now := time.Now()
	for i := range sessions {
		ua, _ := sessions[i]["user_agent"].(string)
		sessions[i]["browser"] = parseBrowserHint(ua)
		sessions[i]["os"] = parseOSHint(ua)
		sessions[i]["device_type"] = parseDeviceHint(ua)
		if expiresAt, ok := sessions[i]["expires_at"].(time.Time); ok {
			sessions[i]["is_expired"] = expiresAt.Before(now)
		} else {
			sessions[i]["is_expired"] = false
		}
	}

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
	const impersonationDuration = 15 * time.Minute
	tokenInfo, err := utils.GenerateTokenWithInfo(userID, email, name, role, nil)
	if err != nil {
		log.Printf("Error generating impersonation token: %v", err)
		return utils.SendInternalError(c, "Failed to generate impersonation token", err)
	}

	now := time.Now()
	expiresAt := now.Add(impersonationDuration)

	// Lookup impersonator email for the log
	var impersonatorEmail string
	var impersonatorRow map[string]interface{}
	if db.Table("users").Select("email").Where("id = ?", adminUserID).First(&impersonatorRow).Error == nil {
		impersonatorEmail, _ = impersonatorRow["email"].(string)
	}

	// Write to impersonation_logs
	db.Table("impersonation_logs").Create(map[string]interface{}{
		"id":                 utils.GenerateID(),
		"impersonator_id":    adminUserID,
		"impersonator_email": impersonatorEmail,
		"target_id":          userID,
		"target_email":       email,
		"impersonation_type": "platform_user",
		"token_jti":          tokenInfo.JTI,
		"expires_at":         expiresAt,
		"created_at":         now,
	})

	// Also log in admin_audit_logs
	db.Table("admin_audit_logs").Create(map[string]interface{}{
		"id":            utils.GenerateID(),
		"action":        "user_impersonation",
		"admin_user_id": adminUserID,
		"new_value":     userID,
		"description":   "Admin impersonated platform user: " + email,
		"created_at":    now,
	})

	return utils.SendSimpleSuccess(c, map[string]interface{}{
		"impersonation_token": tokenInfo.Token,
		"expires_in":          int(impersonationDuration.Seconds()),
		"impersonated_user": map[string]interface{}{
			"id":    userID,
			"email": email,
			"name":  name,
		},
		"warning": "This is a short-lived token for impersonation purposes. All actions will be logged.",
	}, "Impersonation token generated successfully")
}

// AdminGetUserWorkStats returns work statistics for a specific user (documents created, approvals, pending)
func AdminGetUserWorkStats(c *fiber.Ctx) error {
	db := config.DB
	userID := c.Params("id")

	// Documents created by type
	docTypes := []struct {
		table string
		key   string
	}{
		{"requisitions", "requisitions"},
		{"purchase_orders", "purchase_orders"},
		{"payment_vouchers", "payment_vouchers"},
		{"goods_received_notes", "grns"},
		{"budgets", "budgets"},
	}

	docCounts := map[string]int64{}
	var totalDocs int64
	for _, dt := range docTypes {
		var cnt int64
		// Attempt query; silently skip if table doesn't exist
		if err := db.Table(dt.table).Where("created_by = ? AND deleted_at IS NULL", userID).Count(&cnt).Error; err == nil {
			docCounts[dt.key] = cnt
			totalDocs += cnt
		}
	}

	// Approvals made (workflow_assignments where user was approver)
	var totalApprovals, approvedCount, rejectedCount int64
	db.Table("workflow_assignments").
		Where("approver_id = ? AND UPPER(status) IN ('APPROVED','REJECTED')", userID).
		Count(&totalApprovals)
	db.Table("workflow_assignments").
		Where("approver_id = ? AND UPPER(status) = 'APPROVED'", userID).
		Count(&approvedCount)
	db.Table("workflow_assignments").
		Where("approver_id = ? AND UPPER(status) = 'REJECTED'", userID).
		Count(&rejectedCount)

	// Pending tasks
	var pendingTasks int64
	db.Table("workflow_assignments").
		Where("approver_id = ? AND UPPER(status) IN ('PENDING','CLAIMED')", userID).
		Count(&pendingTasks)

	// Activity in last 30 days
	var recentActivity int64
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	db.Table("user_activity_logs").
		Where("user_id = ? AND created_at >= ?", userID, thirtyDaysAgo).
		Count(&recentActivity)

	stats := map[string]interface{}{
		"documents_created": map[string]interface{}{
			"total":           totalDocs,
			"breakdown":       docCounts,
		},
		"approvals": map[string]interface{}{
			"total":    totalApprovals,
			"approved": approvedCount,
			"rejected": rejectedCount,
		},
		"pending_tasks":          pendingTasks,
		"activity_last_30_days":  recentActivity,
	}

	return utils.SendSimpleSuccess(c, stats, "User statistics retrieved successfully")
}

// AdminGetUserSecurityEvents returns security-relevant activity events for a user
func AdminGetUserSecurityEvents(c *fiber.Ctx) error {
	db := config.DB
	userID := c.Params("id")

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)
	page, limit = utils.NormalizePaginationParams(page, limit)
	offset := (page - 1) * limit

	securityTypes := []string{
		"login", "failed_login", "logout",
		"password_change", "password_reset_request",
		"session_terminate", "account_lockout",
	}

	var total int64
	db.Table("user_activity_logs").
		Where("user_id = ? AND action_type IN ?", userID, securityTypes).
		Count(&total)

	var events []map[string]interface{}
	db.Table("user_activity_logs").
		Select("id::text, action_type, COALESCE(ip_address,'') as ip_address, COALESCE(user_agent,'') as user_agent, metadata, created_at").
		Where("user_id = ? AND action_type IN ?", userID, securityTypes).
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Scan(&events)

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return utils.SendSimpleSuccess(c, map[string]interface{}{
		"events": events,
		"pagination": map[string]interface{}{
			"total_records": total,
			"total_pages":   totalPages,
			"current_page":  page,
			"has_next":      page < totalPages,
			"has_prev":      page > 1,
		},
	}, "Security events retrieved successfully")
}

// AdminGetUserLoginHistory returns login and failed login events for a user
func AdminGetUserLoginHistory(c *fiber.Ctx) error {
	db := config.DB
	userID := c.Params("id")

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)
	page, limit = utils.NormalizePaginationParams(page, limit)
	offset := (page - 1) * limit

	var total int64
	db.Table("user_activity_logs").
		Where("user_id = ? AND action_type IN ('login','failed_login')", userID).
		Count(&total)

	var logins []map[string]interface{}
	db.Table("user_activity_logs").
		Select("id::text, action_type, COALESCE(ip_address,'') as ip_address, COALESCE(user_agent,'') as user_agent, metadata, created_at").
		Where("user_id = ? AND action_type IN ('login','failed_login')", userID).
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Scan(&logins)

	// Annotate each with success flag and simple device info
	for i := range logins {
		logins[i]["success"] = logins[i]["action_type"] == "login"
		ua, _ := logins[i]["user_agent"].(string)
		logins[i]["device"] = parseDeviceHint(ua)
		logins[i]["browser"] = parseBrowserHint(ua)
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return utils.SendSimpleSuccess(c, map[string]interface{}{
		"logins": logins,
		"pagination": map[string]interface{}{
			"total_records": total,
			"total_pages":   totalPages,
			"current_page":  page,
			"has_next":      page < totalPages,
			"has_prev":      page > 1,
		},
	}, "Login history retrieved successfully")
}

// parseDeviceHint returns a simple device label from the user-agent string
func parseDeviceHint(ua string) string {
	ua = strings.ToLower(ua)
	if strings.Contains(ua, "mobile") || strings.Contains(ua, "android") || strings.Contains(ua, "iphone") {
		return "Mobile"
	}
	if strings.Contains(ua, "tablet") || strings.Contains(ua, "ipad") {
		return "Tablet"
	}
	return "Desktop"
}

// parseOSHint returns a simple OS label from the user-agent string
func parseOSHint(ua string) string {
	ual := strings.ToLower(ua)
	switch {
	case strings.Contains(ual, "windows"):
		return "Windows"
	case strings.Contains(ual, "mac os") || strings.Contains(ual, "macos") || strings.Contains(ual, "darwin"):
		return "macOS"
	case strings.Contains(ual, "android"):
		return "Android"
	case strings.Contains(ual, "iphone") || strings.Contains(ual, "ipad") || strings.Contains(ual, "ios"):
		return "iOS"
	case strings.Contains(ual, "linux"):
		return "Linux"
	default:
		return ""
	}
}

// parseBrowserHint returns a simple browser label from the user-agent string
func parseBrowserHint(ua string) string {
	ua = strings.ToLower(ua)
	switch {
	case strings.Contains(ua, "edg/"):
		return "Edge"
	case strings.Contains(ua, "chrome"):
		return "Chrome"
	case strings.Contains(ua, "firefox"):
		return "Firefox"
	case strings.Contains(ua, "safari"):
		return "Safari"
	case strings.Contains(ua, "opera") || strings.Contains(ua, "opr/"):
		return "Opera"
	case strings.Contains(ua, "axios") || strings.Contains(ua, "curl") || strings.Contains(ua, "python") || strings.Contains(ua, "go-http"):
		return "API Client"
	default:
		return ""
	}
}

// AdminExportUserActivity exports a user's activity log as CSV or JSON
func AdminExportUserActivity(c *fiber.Ctx) error {
	db := config.DB
	userID := c.Params("id")
	format := strings.ToLower(c.Query("format", "csv"))

	actionType := c.Query("action_type")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	query := db.Table("user_activity_logs").
		Select("id::text, action_type, COALESCE(resource_type,'') as resource_type, COALESCE(resource_id,'') as resource_id, COALESCE(ip_address,'') as ip_address, COALESCE(user_agent,'') as user_agent, created_at").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(10000)

	if actionType != "" {
		query = query.Where("action_type = ?", actionType)
	}
	if startDateStr != "" {
		if t, err := time.Parse("2006-01-02", startDateStr); err == nil {
			query = query.Where("created_at >= ?", t)
		}
	}
	if endDateStr != "" {
		if t, err := time.Parse("2006-01-02", endDateStr); err == nil {
			query = query.Where("created_at <= ?", t.Add(24*time.Hour))
		}
	}

	type exportRow struct {
		ID           string    `json:"id"`
		ActionType   string    `json:"action_type"`
		ResourceType string    `json:"resource_type"`
		ResourceID   string    `json:"resource_id"`
		IPAddress    string    `json:"ip_address"`
		UserAgent    string    `json:"user_agent"`
		CreatedAt    time.Time `json:"created_at"`
	}

	var rows []exportRow
	query.Scan(&rows)

	if format == "json" {
		data, err := json.Marshal(rows)
		if err != nil {
			return utils.SendInternalError(c, "Failed to serialize activity data", err)
		}
		filename := fmt.Sprintf("activity_%s_%s.json", userID, time.Now().Format("20060102"))
		c.Set("Content-Disposition", "attachment; filename="+filename)
		c.Set("Content-Type", "application/json")
		return c.Send(data)
	}

	// Default: CSV
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{"id", "action_type", "resource_type", "resource_id", "ip_address", "user_agent", "created_at"})
	for _, r := range rows {
		_ = w.Write([]string{
			r.ID,
			r.ActionType,
			r.ResourceType,
			r.ResourceID,
			r.IPAddress,
			r.UserAgent,
			r.CreatedAt.Format(time.RFC3339),
		})
	}
	w.Flush()

	filename := fmt.Sprintf("activity_%s_%s.csv", userID, time.Now().Format("20060102"))
	c.Set("Content-Disposition", "attachment; filename="+filename)
	c.Set("Content-Type", "text/csv")
	return c.Send(buf.Bytes())
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
