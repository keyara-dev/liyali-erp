package handlers

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/middleware"
	"github.com/liyali/liyali-gateway/utils"
	"golang.org/x/crypto/bcrypt"
)

// orgMemberGuard verifies that userID is an active member of the caller's organization.
// Returns a non-nil error response if the check fails; callers should return that value immediately.
func orgMemberGuard(c *fiber.Ctx, orgID, userID string) error {
	var count int64
	config.DB.Table("organization_members").
		Where("organization_id = ? AND user_id = ? AND active = true", orgID, userID).
		Count(&count)
	if count == 0 {
		return utils.SendNotFoundError(c, "User not found in this organization")
	}
	return nil
}

// OrgGetUserById returns a single user by ID, scoped to the caller's organization.
// GET /api/v1/organization/users/:id
func OrgGetUserById(c *fiber.Ctx) error {
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return utils.SendUnauthorizedError(c, "Organization context required")
	}

	userID := c.Params("id")
	if err := orgMemberGuard(c, tenant.OrganizationID, userID); err != nil {
		return err
	}

	db := config.DB
	var user map[string]interface{}
	err = db.Table("users").
		Select(`id, email, name, role,
			CASE WHEN active = true THEN 'active' ELSE 'suspended' END as status,
			active as is_active,
			created_at, updated_at, last_login,
			COALESCE(position, '') as position,
			COALESCE(man_number, '') as man_number,
			COALESCE(nrc_number, '') as nrc_number,
			COALESCE(contact, '') as contact,
			COALESCE(mfa_enabled, false) as mfa_enabled,
			preferences`).
		Where("id = ? AND deleted_at IS NULL", userID).
		Limit(1).Scan(&user).Error

	if err != nil || len(user) == 0 {
		return utils.SendNotFoundError(c, "User not found")
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
		Where("organization_members.user_id = ? AND organization_members.organization_id = ?", userID, tenant.OrganizationID).
		Find(&orgs)

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
	if _, ok := user["mfa_enabled"]; !ok {
		user["mfa_enabled"] = false
	}
	user["email_verified"] = true
	user["login_count"] = 0
	if phone, ok := user["contact"].(string); ok {
		user["phone"] = phone
	}
	// Normalise snake_case profile fields to camelCase for the frontend
	user["manNumber"] = user["man_number"]
	user["nrcNumber"] = user["nrc_number"]

	return utils.SendSimpleSuccess(c, user, "User retrieved successfully")
}

// OrgUpdateUserStatus activates or suspends a user, scoped to the caller's organization.
// PUT /api/v1/organization/users/:id/status
func OrgUpdateUserStatus(c *fiber.Ctx) error {
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return utils.SendUnauthorizedError(c, "Organization context required")
	}

	userID := c.Params("id")
	if err := orgMemberGuard(c, tenant.OrganizationID, userID); err != nil {
		return err
	}

	var request struct {
		Status string `json:"status"`
		Reason string `json:"reason"`
	}
	if err := c.BodyParser(&request); err != nil {
		return utils.SendBadRequestError(c, "Invalid request body")
	}

	if request.Status != "active" && request.Status != "suspended" && request.Status != "inactive" {
		return utils.SendBadRequestError(c, "Invalid status. Must be 'active', 'suspended', or 'inactive'")
	}

	active := request.Status == "active"

	if err := config.DB.Table("users").Where("id = ? AND deleted_at IS NULL", userID).Updates(map[string]interface{}{
		"active":     active,
		"updated_at": time.Now(),
	}).Error; err != nil {
		return utils.SendInternalError(c, "Failed to update user status", err)
	}

	config.DB.Table("admin_audit_logs").Create(map[string]interface{}{
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

// OrgResetUserPassword resets a user's password, scoped to the caller's organization.
// POST /api/v1/organization/users/:id/reset-password
func OrgResetUserPassword(c *fiber.Ctx) error {
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return utils.SendUnauthorizedError(c, "Organization context required")
	}

	userID := c.Params("id")
	if err := orgMemberGuard(c, tenant.OrganizationID, userID); err != nil {
		return err
	}

	var request struct {
		SendEmail bool `json:"send_email"`
	}
	c.BodyParser(&request)

	var existingCount int64
	config.DB.Table("users").Where("id = ? AND deleted_at IS NULL", userID).Count(&existingCount)
	if existingCount == 0 {
		return utils.SendNotFoundError(c, "User not found")
	}

	tempPassword := utils.GenerateID()[:12]
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(tempPassword), bcrypt.DefaultCost)
	if err != nil {
		return utils.SendInternalError(c, "Failed to generate password", err)
	}

	config.DB.Table("users").Where("id = ?", userID).Updates(map[string]interface{}{
		"password":   string(hashedPassword),
		"updated_at": time.Now(),
	})

	config.DB.Table("admin_audit_logs").Create(map[string]interface{}{
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

// OrgGetUserActivity returns paginated activity logs for a user, scoped to the caller's organization.
// GET /api/v1/organization/users/:id/activity
func OrgGetUserActivity(c *fiber.Ctx) error {
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return utils.SendUnauthorizedError(c, "Organization context required")
	}

	userID := c.Params("id")
	if err := orgMemberGuard(c, tenant.OrganizationID, userID); err != nil {
		return err
	}

	db := config.DB
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)
	page, limit = utils.NormalizePaginationParams(page, limit)
	offset := (page - 1) * limit

	actionType := c.Query("action_type")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

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

	var ualTotal int64
	db.Table("user_activity_logs").Where(ualConds, ualArgs...).Count(&ualTotal)

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

	return utils.SendSimpleSuccess(c, map[string]interface{}{
		"activities": activities,
		"pagination": map[string]interface{}{
			"total_records": ualTotal,
			"total_pages":   totalPages,
			"current_page":  page,
			"has_next":      page < totalPages,
			"has_prev":      page > 1,
		},
	}, "User activity retrieved successfully")
}

// OrgExportUserActivity exports a user's activity log as CSV or JSON.
// GET /api/v1/organization/users/:id/activity/export
func OrgExportUserActivity(c *fiber.Ctx) error {
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return utils.SendUnauthorizedError(c, "Organization context required")
	}

	userID := c.Params("id")
	if err := orgMemberGuard(c, tenant.OrganizationID, userID); err != nil {
		return err
	}

	db := config.DB
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

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{"id", "action_type", "resource_type", "resource_id", "ip_address", "user_agent", "created_at"})
	for _, r := range rows {
		_ = w.Write([]string{
			r.ID, r.ActionType, r.ResourceType, r.ResourceID, r.IPAddress, r.UserAgent,
			r.CreatedAt.Format(time.RFC3339),
		})
	}
	w.Flush()

	filename := fmt.Sprintf("activity_%s_%s.csv", userID, time.Now().Format("20060102"))
	c.Set("Content-Disposition", "attachment; filename="+filename)
	c.Set("Content-Type", "text/csv")
	return c.Send(buf.Bytes())
}

// OrgGetUserSecurityEvents returns security-relevant activity events for a user, scoped to the caller's organization.
// GET /api/v1/organization/users/:id/security-events
func OrgGetUserSecurityEvents(c *fiber.Ctx) error {
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return utils.SendUnauthorizedError(c, "Organization context required")
	}

	userID := c.Params("id")
	if err := orgMemberGuard(c, tenant.OrganizationID, userID); err != nil {
		return err
	}

	db := config.DB
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

// OrgGetUserLoginHistory returns login and failed-login events for a user, scoped to the caller's organization.
// GET /api/v1/organization/users/:id/login-history
func OrgGetUserLoginHistory(c *fiber.Ctx) error {
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return utils.SendUnauthorizedError(c, "Organization context required")
	}

	userID := c.Params("id")
	if err := orgMemberGuard(c, tenant.OrganizationID, userID); err != nil {
		return err
	}

	db := config.DB
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

// OrgGetUserWorkStats returns work statistics for a user, scoped to the caller's organization.
// GET /api/v1/organization/users/:id/work-stats
func OrgGetUserWorkStats(c *fiber.Ctx) error {
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return utils.SendUnauthorizedError(c, "Organization context required")
	}

	userID := c.Params("id")
	if err := orgMemberGuard(c, tenant.OrganizationID, userID); err != nil {
		return err
	}

	db := config.DB
	orgID := tenant.OrganizationID

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
		if err := db.Table(dt.table).
			Where("created_by = ? AND organization_id = ? AND deleted_at IS NULL", userID, orgID).
			Count(&cnt).Error; err == nil {
			docCounts[dt.key] = cnt
			totalDocs += cnt
		}
	}

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

	var pendingTasks int64
	db.Table("workflow_assignments").
		Where("approver_id = ? AND UPPER(status) IN ('PENDING','CLAIMED')", userID).
		Count(&pendingTasks)

	var recentActivity int64
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	db.Table("user_activity_logs").
		Where("user_id = ? AND created_at >= ?", userID, thirtyDaysAgo).
		Count(&recentActivity)

	return utils.SendSimpleSuccess(c, map[string]interface{}{
		"documents_created": map[string]interface{}{
			"total":     totalDocs,
			"breakdown": docCounts,
		},
		"approvals": map[string]interface{}{
			"total":    totalApprovals,
			"approved": approvedCount,
			"rejected": rejectedCount,
		},
		"pending_tasks":         pendingTasks,
		"activity_last_30_days": recentActivity,
	}, "User statistics retrieved successfully")
}

// OrgGetUserSessions returns active sessions for a user, scoped to the caller's organization.
// GET /api/v1/organization/users/:id/sessions
func OrgGetUserSessions(c *fiber.Ctx) error {
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return utils.SendUnauthorizedError(c, "Organization context required")
	}

	userID := c.Params("id")
	if err := orgMemberGuard(c, tenant.OrganizationID, userID); err != nil {
		return err
	}

	var sessions []map[string]interface{}
	config.DB.Table("sessions").
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

// OrgTerminateUserSession terminates a specific session for a user, scoped to the caller's organization.
// DELETE /api/v1/organization/users/:id/sessions/:sessionId
func OrgTerminateUserSession(c *fiber.Ctx) error {
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return utils.SendUnauthorizedError(c, "Organization context required")
	}

	userID := c.Params("id")
	if err := orgMemberGuard(c, tenant.OrganizationID, userID); err != nil {
		return err
	}

	sessionID := c.Params("sessionId")
	config.DB.Table("sessions").Where("id = ? AND user_id = ?", sessionID, userID).Delete(nil)

	return utils.SendSimpleSuccess(c, nil, "Session terminated successfully")
}

// OrgTerminateAllUserSessions terminates all sessions for a user, scoped to the caller's organization.
// DELETE /api/v1/organization/users/:id/sessions
func OrgTerminateAllUserSessions(c *fiber.Ctx) error {
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return utils.SendUnauthorizedError(c, "Organization context required")
	}

	userID := c.Params("id")
	if err := orgMemberGuard(c, tenant.OrganizationID, userID); err != nil {
		return err
	}

	config.DB.Table("sessions").Where("user_id = ?", userID).Delete(nil)

	return utils.SendSimpleSuccess(c, nil, "All sessions terminated successfully")
}

// OrgImpersonateUser generates a short-lived impersonation token for a user in the caller's organization.
// Only the org admin may call this. The token is valid for 15 minutes and all usage is audit-logged.
// POST /api/v1/organization/users/:id/impersonate
func OrgImpersonateUser(c *fiber.Ctx) error {
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return utils.SendUnauthorizedError(c, "Organization context required")
	}

	// Only org admins may impersonate
	if tenant.UserRole != "admin" {
		return utils.SendForbiddenError(c, "Admin access required to impersonate users")
	}

	callerID, _ := c.Locals("userID").(string)
	userID := c.Params("id")

	// Cannot impersonate yourself
	if userID == callerID {
		return utils.SendBadRequestError(c, "You cannot impersonate yourself")
	}

	if err := orgMemberGuard(c, tenant.OrganizationID, userID); err != nil {
		return err
	}

	// Fetch target user details
	var user map[string]interface{}
	if err := config.DB.Table("users").
		Select("id, email, name, role, active").
		Where("id = ? AND deleted_at IS NULL", userID).
		Limit(1).Scan(&user).Error; err != nil || len(user) == 0 {
		return utils.SendNotFoundError(c, "User not found")
	}

	// Target user must be active
	if active, ok := user["active"].(bool); !ok || !active {
		return utils.SendBadRequestError(c, "Cannot impersonate an inactive or suspended user")
	}

	email, _ := user["email"].(string)
	name, _ := user["name"].(string)
	role, _ := user["role"].(string)
	if role == "" {
		role = "requester"
	}

	// Generate a 15-minute token scoped to the target user
	const impersonationDuration = 15 * time.Minute
	orgID := tenant.OrganizationID
	tokenInfo, err := utils.GenerateTokenWithInfo(userID, email, name, role, &orgID)
	if err != nil {
		return utils.SendInternalError(c, "Failed to generate impersonation token", err)
	}

	now := time.Now()
	expiresAt := now.Add(impersonationDuration)

	// Lookup caller's email for the audit log
	var callerEmail string
	var callerRow map[string]interface{}
	if config.DB.Table("users").Select("email").Where("id = ?", callerID).Limit(1).Scan(&callerRow).Error == nil {
		callerEmail, _ = callerRow["email"].(string)
	}

	// Write to impersonation_logs (dedicated audit table)
	config.DB.Table("impersonation_logs").Create(map[string]interface{}{
		"id":                 utils.GenerateID(),
		"impersonator_id":    callerID,
		"impersonator_email": callerEmail,
		"target_id":          userID,
		"target_email":       email,
		"impersonation_type": "platform_user",
		"token_jti":          tokenInfo.JTI,
		"expires_at":         expiresAt,
		"created_at":         now,
	})

	// Also write to admin_audit_logs for general admin activity tracking
	config.DB.Table("admin_audit_logs").Create(map[string]interface{}{
		"id":            utils.GenerateID(),
		"action":        "user_impersonation",
		"admin_user_id": callerID,
		"new_value":     userID,
		"description":   "Org admin impersonated user: " + email,
		"created_at":    now,
	})

	return utils.SendSimpleSuccess(c, map[string]interface{}{
		"token":      tokenInfo.Token,
		"expires_in": int(impersonationDuration.Seconds()),
		"impersonated_user": map[string]interface{}{
			"id":    userID,
			"email": email,
			"name":  name,
		},
		"warning": "This is a short-lived token for impersonation purposes. All actions will be logged.",
	}, "Impersonation token generated successfully")
}
