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
)

// adminAuditLogRow represents a row from the admin_audit_logs table
type adminAuditLogRow struct {
	ID             string          `json:"id"`
	OrganizationID *string         `json:"organization_id"`
	Action         string          `json:"action"`
	AdminUserID    string          `json:"admin_user_id"`
	Details        json.RawMessage `json:"details"`
	CreatedAt      time.Time       `json:"created_at"`
}

// deriveActionType extracts a high-level action type from the action string.
// For example "user.create" -> "create", "login_failed" -> "login".
func deriveActionType(action string) string {
	parts := strings.Split(action, ".")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	parts = strings.Split(action, "_")
	if len(parts) > 0 {
		return parts[0]
	}
	return action
}

// deriveResourceType extracts the resource type from the action string.
// For example "user.create" -> "user", "budget_approve" -> "budget".
func deriveResourceType(action string) string {
	parts := strings.Split(action, ".")
	if len(parts) > 1 {
		return parts[0]
	}
	parts = strings.Split(action, "_")
	if len(parts) > 0 {
		return parts[0]
	}
	return "system"
}

// deriveSeverity returns a severity level based on the action string.
func deriveSeverity(action string) string {
	lower := strings.ToLower(action)
	switch {
	case strings.Contains(lower, "delete") || strings.Contains(lower, "remove"):
		return "high"
	case strings.Contains(lower, "fail") || strings.Contains(lower, "error") || strings.Contains(lower, "unauthorized"):
		return "critical"
	case strings.Contains(lower, "update") || strings.Contains(lower, "edit") || strings.Contains(lower, "modify"):
		return "medium"
	case strings.Contains(lower, "login") || strings.Contains(lower, "logout") || strings.Contains(lower, "password"):
		return "medium"
	default:
		return "low"
	}
}

// deriveStatus returns a status based on the action string.
func deriveStatus(action string) string {
	lower := strings.ToLower(action)
	if strings.Contains(lower, "fail") || strings.Contains(lower, "error") {
		return "failure"
	}
	return "success"
}

// mapAuditLogRow transforms a database row into the frontend-expected shape.
func mapAuditLogRow(row adminAuditLogRow) map[string]interface{} {
	// Parse details to extract optional metadata fields
	var detailsMap map[string]interface{}
	if row.Details != nil {
		_ = json.Unmarshal(row.Details, &detailsMap)
	}
	if detailsMap == nil {
		detailsMap = map[string]interface{}{}
	}

	// Try to extract user info from details if available
	userName, _ := detailsMap["user_name"].(string)
	userEmail, _ := detailsMap["user_email"].(string)
	orgName, _ := detailsMap["organization_name"].(string)
	resourceID, _ := detailsMap["resource_id"].(string)
	ipAddress, _ := detailsMap["ip_address"].(string)
	userAgent, _ := detailsMap["user_agent"].(string)

	metadata := map[string]interface{}{
		"ip_address": ipAddress,
		"user_agent": userAgent,
	}

	result := map[string]interface{}{
		"id":                row.ID,
		"action":            row.Action,
		"action_type":       deriveActionType(row.Action),
		"user_id":           row.AdminUserID,
		"user_name":         userName,
		"user_email":        userEmail,
		"organization_id":   row.OrganizationID,
		"organization_name": orgName,
		"resource_type":     deriveResourceType(row.Action),
		"resource_id":       resourceID,
		"details":           detailsMap,
		"metadata":          metadata,
		"timestamp":         row.CreatedAt,
		"severity":          deriveSeverity(row.Action),
		"status":            deriveStatus(row.Action),
		"duration_ms":       nil,
	}

	return result
}

// GetAdminAuditLogs returns a paginated and filtered list of admin audit logs.
// GET /api/v1/admin/audit-logs
func GetAdminAuditLogs(c *fiber.Ctx) error {
	db := config.DB

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)
	page, limit = utils.NormalizePaginationParams(page, limit)

	query := db.Table("admin_audit_logs")

	// Apply filters
	if userID := c.Query("user_id"); userID != "" {
		query = query.Where("admin_user_id = ?", userID)
	}
	if organizationID := c.Query("organization_id"); organizationID != "" {
		query = query.Where("organization_id = ?", organizationID)
	}
	if actionType := c.Query("action_type"); actionType != "" {
		query = query.Where("action ILIKE ?", "%"+actionType+"%")
	}
	if resourceType := c.Query("resource_type"); resourceType != "" {
		query = query.Where("action ILIKE ?", resourceType+".%")
	}
	if search := c.Query("search"); search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("action ILIKE ? OR details::text ILIKE ?", searchPattern, searchPattern)
	}
	if ipAddress := c.Query("ip_address"); ipAddress != "" {
		query = query.Where("details->>'ip_address' = ?", ipAddress)
	}
	if severity := c.Query("severity"); severity != "" {
		// Filter by derived severity: match actions that map to this severity
		switch severity {
		case "critical":
			query = query.Where("action ILIKE '%fail%' OR action ILIKE '%error%' OR action ILIKE '%unauthorized%'")
		case "high":
			query = query.Where("action ILIKE '%delete%' OR action ILIKE '%remove%'")
		case "medium":
			query = query.Where("action ILIKE '%update%' OR action ILIKE '%edit%' OR action ILIKE '%login%' OR action ILIKE '%logout%' OR action ILIKE '%password%'")
		case "low":
			query = query.Where(
				"action NOT ILIKE '%fail%' AND action NOT ILIKE '%error%' AND action NOT ILIKE '%unauthorized%' "+
					"AND action NOT ILIKE '%delete%' AND action NOT ILIKE '%remove%' "+
					"AND action NOT ILIKE '%update%' AND action NOT ILIKE '%edit%' "+
					"AND action NOT ILIKE '%login%' AND action NOT ILIKE '%logout%' AND action NOT ILIKE '%password%'",
			)
		}
	}
	if status := c.Query("status"); status != "" {
		if status == "failure" {
			query = query.Where("action ILIKE '%fail%' OR action ILIKE '%error%'")
		} else if status == "success" {
			query = query.Where("action NOT ILIKE '%fail%' AND action NOT ILIKE '%error%'")
		}
	}

	// Date range filters
	if dateRange := c.Query("date_range"); dateRange != "" {
		now := time.Now()
		switch dateRange {
		case "today":
			query = query.Where("created_at >= ?", now.Truncate(24*time.Hour))
		case "yesterday":
			yesterday := now.AddDate(0, 0, -1).Truncate(24 * time.Hour)
			query = query.Where("created_at >= ? AND created_at < ?", yesterday, now.Truncate(24*time.Hour))
		case "last_7_days":
			query = query.Where("created_at >= ?", now.AddDate(0, 0, -7))
		case "last_30_days":
			query = query.Where("created_at >= ?", now.AddDate(0, 0, -30))
		case "last_90_days":
			query = query.Where("created_at >= ?", now.AddDate(0, 0, -90))
		}
	}
	if startDate := c.Query("start_date"); startDate != "" {
		if t, err := time.Parse("2006-01-02", startDate); err == nil {
			query = query.Where("created_at >= ?", t)
		}
	}
	if endDate := c.Query("end_date"); endDate != "" {
		if t, err := time.Parse("2006-01-02", endDate); err == nil {
			query = query.Where("created_at <= ?", t.Add(24*time.Hour))
		}
	}

	// Count total matching records
	var total int64
	if err := query.Count(&total).Error; err != nil {
		log.Printf("Error counting admin audit logs: %v", err)
		return utils.SendInternalError(c, "Failed to count audit logs", err)
	}

	// Fetch paginated rows
	var rows []adminAuditLogRow
	offset := (page - 1) * limit
	if err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&rows).Error; err != nil {
		log.Printf("Error fetching admin audit logs: %v", err)
		return utils.SendInternalError(c, "Failed to fetch audit logs", err)
	}

	// Map to frontend shape
	results := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		results = append(results, mapAuditLogRow(row))
	}

	return utils.SendPaginatedSuccess(c, results, "Audit logs retrieved successfully", page, limit, total)
}

// GetAdminAuditLogStats returns aggregated audit log statistics.
// GET /api/v1/admin/audit-logs/stats
func GetAdminAuditLogStats(c *fiber.Ctx) error {
	db := config.DB

	// Total logs
	var totalLogs int64
	if err := db.Table("admin_audit_logs").Count(&totalLogs).Error; err != nil {
		log.Printf("Error counting total audit logs: %v", err)
		return utils.SendInternalError(c, "Failed to fetch audit log stats", err)
	}

	// Logs today
	today := time.Now().Truncate(24 * time.Hour)
	var logsToday int64
	db.Table("admin_audit_logs").Where("created_at >= ?", today).Count(&logsToday)

	// Failed actions (actions containing fail/error in their name)
	var failedActions int64
	db.Table("admin_audit_logs").
		Where("action ILIKE '%fail%' OR action ILIKE '%error%'").
		Count(&failedActions)

	// Critical events (unauthorized, delete, remove)
	var criticalEvents int64
	db.Table("admin_audit_logs").
		Where("action ILIKE '%unauthorized%' OR action ILIKE '%delete%' OR action ILIKE '%remove%'").
		Count(&criticalEvents)

	// Unique users
	var uniqueUsers int64
	db.Table("admin_audit_logs").
		Distinct("admin_user_id").
		Count(&uniqueUsers)

	// Top actions
	type actionCount struct {
		Action string `json:"action"`
		Count  int64  `json:"count"`
	}
	var topActions []actionCount
	db.Table("admin_audit_logs").
		Select("action, COUNT(*) as count").
		Group("action").
		Order("count DESC").
		Limit(10).
		Scan(&topActions)

	topActionsFormatted := make([]map[string]interface{}, 0, len(topActions))
	for _, a := range topActions {
		percentage := float64(0)
		if totalLogs > 0 {
			percentage = float64(a.Count) / float64(totalLogs) * 100
		}
		topActionsFormatted = append(topActionsFormatted, map[string]interface{}{
			"action":     a.Action,
			"count":      a.Count,
			"percentage": percentage,
		})
	}

	// Activity by hour (last 24 hours)
	activityByHour := make([]map[string]interface{}, 0, 24)
	for h := 0; h < 24; h++ {
		hourStart := today.Add(time.Duration(h) * time.Hour)
		hourEnd := hourStart.Add(time.Hour)

		var hourCount, hourFailedCount int64
		db.Table("admin_audit_logs").
			Where("created_at >= ? AND created_at < ?", hourStart, hourEnd).
			Count(&hourCount)
		db.Table("admin_audit_logs").
			Where("created_at >= ? AND created_at < ? AND (action ILIKE '%fail%' OR action ILIKE '%error%')", hourStart, hourEnd).
			Count(&hourFailedCount)

		activityByHour = append(activityByHour, map[string]interface{}{
			"hour":         h,
			"count":        hourCount,
			"failed_count": hourFailedCount,
		})
	}

	// Security events
	var failedLogins, suspiciousActivities, policyViolations, unauthorizedAttempts int64
	db.Table("admin_audit_logs").
		Where("action ILIKE '%login%fail%' OR action ILIKE '%login_fail%'").
		Count(&failedLogins)
	db.Table("admin_audit_logs").
		Where("action ILIKE '%suspicious%'").
		Count(&suspiciousActivities)
	db.Table("admin_audit_logs").
		Where("action ILIKE '%policy%violation%'").
		Count(&policyViolations)
	db.Table("admin_audit_logs").
		Where("action ILIKE '%unauthorized%'").
		Count(&unauthorizedAttempts)

	stats := map[string]interface{}{
		"total_logs":       totalLogs,
		"logs_today":       logsToday,
		"failed_actions":   failedActions,
		"critical_events":  criticalEvents,
		"unique_users":     uniqueUsers,
		"top_actions":      topActionsFormatted,
		"activity_by_hour": activityByHour,
		"security_events": map[string]interface{}{
			"failed_logins":                failedLogins,
			"suspicious_activities":        suspiciousActivities,
			"policy_violations":            policyViolations,
			"unauthorized_access_attempts": unauthorizedAttempts,
		},
	}

	return utils.SendSimpleSuccess(c, stats, "Audit log stats retrieved successfully")
}

// GetAdminAuditLogAnalytics returns audit log analytics data.
// GET /api/v1/admin/audit-logs/analytics
func GetAdminAuditLogAnalytics(c *fiber.Ctx) error {
	db := config.DB

	// Total logs
	var totalLogs int64
	db.Table("admin_audit_logs").Count(&totalLogs)

	// Daily trend (last 30 days)
	dailyTrend := make([]map[string]interface{}, 0, 30)
	for i := 29; i >= 0; i-- {
		day := time.Now().AddDate(0, 0, -i).Truncate(24 * time.Hour)
		dayEnd := day.Add(24 * time.Hour)

		var dayCount int64
		db.Table("admin_audit_logs").
			Where("created_at >= ? AND created_at < ?", day, dayEnd).
			Count(&dayCount)

		dailyTrend = append(dailyTrend, map[string]interface{}{
			"date":  day.Format("2006-01-02"),
			"count": dayCount,
		})
	}

	// Action distribution
	type actionDist struct {
		Action string `json:"action"`
		Count  int64  `json:"count"`
	}
	var distribution []actionDist
	db.Table("admin_audit_logs").
		Select("action, COUNT(*) as count").
		Group("action").
		Order("count DESC").
		Limit(15).
		Scan(&distribution)

	actionDistribution := make([]map[string]interface{}, 0, len(distribution))
	for _, d := range distribution {
		percentage := float64(0)
		if totalLogs > 0 {
			percentage = float64(d.Count) / float64(totalLogs) * 100
		}
		actionDistribution = append(actionDistribution, map[string]interface{}{
			"action":     d.Action,
			"count":      d.Count,
			"percentage": percentage,
		})
	}

	// User activity ranking
	type userActivity struct {
		AdminUserID string `json:"admin_user_id"`
		Count       int64  `json:"count"`
	}
	var topUsers []userActivity
	db.Table("admin_audit_logs").
		Select("admin_user_id, COUNT(*) as count").
		Group("admin_user_id").
		Order("count DESC").
		Limit(10).
		Scan(&topUsers)

	topUsersFormatted := make([]map[string]interface{}, 0, len(topUsers))
	for _, u := range topUsers {
		topUsersFormatted = append(topUsersFormatted, map[string]interface{}{
			"user_id": u.AdminUserID,
			"count":   u.Count,
		})
	}

	// Peak hours analysis
	type hourStat struct {
		Hour  int   `json:"hour"`
		Count int64 `json:"count"`
	}
	var peakHours []hourStat
	db.Table("admin_audit_logs").
		Select("EXTRACT(HOUR FROM created_at)::int as hour, COUNT(*) as count").
		Group("hour").
		Order("count DESC").
		Scan(&peakHours)

	peakHoursFormatted := make([]map[string]interface{}, 0, len(peakHours))
	for _, h := range peakHours {
		peakHoursFormatted = append(peakHoursFormatted, map[string]interface{}{
			"hour":  h.Hour,
			"count": h.Count,
		})
	}

	analytics := map[string]interface{}{
		"total_logs":          totalLogs,
		"daily_trend":         dailyTrend,
		"action_distribution": actionDistribution,
		"top_users":           topUsersFormatted,
		"peak_hours":          peakHoursFormatted,
		"generated_at":        time.Now(),
	}

	return utils.SendSimpleSuccess(c, analytics, "Audit log analytics retrieved successfully")
}

// GetAdminAuditLogByID returns a single audit log entry by its ID.
// GET /api/v1/admin/audit-logs/:id
func GetAdminAuditLogByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return utils.SendBadRequest(c, "Audit log ID is required")
	}

	var row adminAuditLogRow
	if err := config.DB.Table("admin_audit_logs").
		Where("id = ?", id).
		First(&row).Error; err != nil {
		return utils.SendNotFound(c, "Audit log not found")
	}

	return utils.SendSimpleSuccess(c, mapAuditLogRow(row), "Audit log retrieved successfully")
}

// ExportAdminAuditLogs exports audit logs as JSON matching the applied filters.
// POST /api/v1/admin/audit-logs/export
func ExportAdminAuditLogs(c *fiber.Ctx) error {
	db := config.DB

	query := db.Table("admin_audit_logs")

	// Apply the same filters as GetAdminAuditLogs
	if userID := c.Query("user_id"); userID != "" {
		query = query.Where("admin_user_id = ?", userID)
	}
	if organizationID := c.Query("organization_id"); organizationID != "" {
		query = query.Where("organization_id = ?", organizationID)
	}
	if actionType := c.Query("action_type"); actionType != "" {
		query = query.Where("action ILIKE ?", "%"+actionType+"%")
	}
	if search := c.Query("search"); search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("action ILIKE ? OR details::text ILIKE ?", searchPattern, searchPattern)
	}
	if dateRange := c.Query("date_range"); dateRange != "" {
		now := time.Now()
		switch dateRange {
		case "1h":
			query = query.Where("created_at >= ?", now.Add(-1*time.Hour))
		case "24h":
			query = query.Where("created_at >= ?", now.Add(-24*time.Hour))
		case "7d":
			query = query.Where("created_at >= ?", now.AddDate(0, 0, -7))
		case "30d":
			query = query.Where("created_at >= ?", now.AddDate(0, 0, -30))
		case "90d":
			query = query.Where("created_at >= ?", now.AddDate(0, 0, -90))
		}
	}
	if startDate := c.Query("start_date"); startDate != "" {
		if t, err := time.Parse("2006-01-02", startDate); err == nil {
			query = query.Where("created_at >= ?", t)
		}
	}
	if endDate := c.Query("end_date"); endDate != "" {
		if t, err := time.Parse("2006-01-02", endDate); err == nil {
			query = query.Where("created_at <= ?", t.Add(24*time.Hour))
		}
	}

	// Limit export to 10000 rows for safety
	var rows []adminAuditLogRow
	if err := query.
		Order("created_at DESC").
		Limit(10000).
		Find(&rows).Error; err != nil {
		log.Printf("Error exporting audit logs: %v", err)
		return utils.SendInternalError(c, "Failed to export audit logs", err)
	}

	results := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		results = append(results, mapAuditLogRow(row))
	}

	exportData := map[string]interface{}{
		"logs":        results,
		"total_count": len(results),
		"exported_at": time.Now().Format(time.RFC3339),
		"filters_applied": map[string]string{
			"user_id":    c.Query("user_id"),
			"action_type": c.Query("action_type"),
			"date_range": c.Query("date_range"),
			"search":     c.Query("search"),
		},
	}

	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=audit-logs-export-%s.json", time.Now().Format("2006-01-02")))
	c.Set("Content-Type", "application/json")

	return c.JSON(exportData)
}

// GetAdminAuditLogSecurityEvents returns audit logs filtered for security-related actions.
// GET /api/v1/admin/audit-logs/security-events
func GetAdminAuditLogSecurityEvents(c *fiber.Ctx) error {
	db := config.DB

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)
	page, limit = utils.NormalizePaginationParams(page, limit)

	securityActions := []string{
		"%login%", "%logout%", "%password%", "%reset%",
		"%unauthorized%", "%permission%", "%role%",
		"%token%", "%session%", "%mfa%", "%2fa%",
		"%lock%", "%ban%", "%suspend%",
	}

	// Build OR conditions for security-related actions
	conditions := make([]string, 0, len(securityActions))
	args := make([]interface{}, 0, len(securityActions))
	for _, pattern := range securityActions {
		conditions = append(conditions, "action ILIKE ?")
		args = append(args, pattern)
	}
	whereClause := strings.Join(conditions, " OR ")

	query := db.Table("admin_audit_logs").Where(whereClause, args...)

	// Count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		log.Printf("Error counting security events: %v", err)
		return utils.SendInternalError(c, "Failed to count security events", err)
	}

	// Fetch
	var rows []adminAuditLogRow
	offset := (page - 1) * limit
	if err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&rows).Error; err != nil {
		log.Printf("Error fetching security events: %v", err)
		return utils.SendInternalError(c, "Failed to fetch security events", err)
	}

	results := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		results = append(results, mapAuditLogRow(row))
	}

	return utils.SendPaginatedSuccess(c, results, "Security events retrieved successfully", page, limit, total)
}

// CreateAdminAuditLog creates a new manual audit log entry.
// POST /api/v1/admin/audit-logs
func CreateAdminAuditLog(c *fiber.Ctx) error {
	userID, _ := c.Locals("userID").(string)

	var req struct {
		Action         string                 `json:"action"`
		OrganizationID *string                `json:"organization_id"`
		Details        map[string]interface{} `json:"details"`
	}

	if err := c.BodyParser(&req); err != nil {
		return utils.SendBadRequest(c, "Invalid request body")
	}

	if req.Action == "" {
		return utils.SendBadRequest(c, "Action is required")
	}

	// Marshal details to JSON
	var detailsJSON []byte
	var err error
	if req.Details != nil {
		detailsJSON, err = json.Marshal(req.Details)
		if err != nil {
			return utils.SendBadRequest(c, "Invalid details format")
		}
	} else {
		detailsJSON = []byte("{}")
	}

	id := utils.GenerateID()
	now := time.Now()

	// Insert using raw map so we work directly against the table columns
	entry := map[string]interface{}{
		"id":              id,
		"organization_id": req.OrganizationID,
		"action":          req.Action,
		"admin_user_id":   userID,
		"details":         string(detailsJSON),
		"created_at":      now,
	}

	if err := config.DB.Table("admin_audit_logs").Create(&entry).Error; err != nil {
		log.Printf("Error creating admin audit log: %v", err)
		return utils.SendInternalError(c, "Failed to create audit log entry", err)
	}

	// Return the created entry in the expected frontend shape
	row := adminAuditLogRow{
		ID:             id,
		OrganizationID: req.OrganizationID,
		Action:         req.Action,
		AdminUserID:    userID,
		Details:        detailsJSON,
		CreatedAt:      now,
	}

	return utils.SendCreatedSuccess(c, mapAuditLogRow(row), "Audit log entry created successfully")
}

// retentionSettingsDefaults returns sensible default retention settings
func retentionSettingsDefaults() map[string]interface{} {
	return map[string]interface{}{
		"retention_days":         90,
		"auto_archive_enabled":  false,
		"archive_after_days":    60,
		"auto_delete_enabled":   false,
		"delete_after_days":     365,
		"compress_after_days":   30,
		"export_before_delete":  true,
		"excluded_action_types": []string{},
	}
}

// GetAdminAuditLogRetentionSettings returns the current retention settings.
// Reads from system_settings table, falls back to defaults.
// GET /api/v1/admin/audit-logs/retention-settings
func GetAdminAuditLogRetentionSettings(c *fiber.Ctx) error {
	db := config.DB
	defaults := retentionSettingsDefaults()

	// Try to load persisted settings from system_settings
	var settingRow map[string]interface{}
	err := db.Table("system_settings").Where("key = ?", "audit_log_retention").First(&settingRow).Error

	if err == nil {
		// Parse the stored JSON value
		valueStr, ok := settingRow["value"].(string)
		if ok {
			var persisted map[string]interface{}
			if jsonErr := json.Unmarshal([]byte(valueStr), &persisted); jsonErr == nil {
				// Merge persisted with defaults (persisted values take priority)
				for k, v := range persisted {
					defaults[k] = v
				}
				defaults["updated_at"] = settingRow["updated_at"]
			}
		}
	} else {
		defaults["updated_at"] = nil
		defaults["updated_by"] = nil
	}

	return utils.SendSimpleSuccess(c, defaults, "Retention settings retrieved successfully")
}

// UpdateAdminAuditLogRetentionSettings persists retention settings to system_settings.
// PUT /api/v1/admin/audit-logs/retention-settings
func UpdateAdminAuditLogRetentionSettings(c *fiber.Ctx) error {
	var req map[string]interface{}
	if err := c.BodyParser(&req); err != nil {
		return utils.SendBadRequest(c, "Invalid request body")
	}

	userID, _ := c.Locals("userID").(string)

	// Build the settings from request merged with defaults
	defaults := retentionSettingsDefaults()
	for k, v := range req {
		defaults[k] = v
	}

	// Serialize to JSON for storage
	settingsJSON, err := json.Marshal(defaults)
	if err != nil {
		return utils.SendInternalError(c, "Failed to serialize settings", err)
	}

	now := time.Now()
	valueStr := string(settingsJSON)

	// Upsert into system_settings
	var existing map[string]interface{}
	dbErr := config.DB.Table("system_settings").Where("key = ?", "audit_log_retention").First(&existing).Error

	if dbErr != nil {
		// Insert new
		config.DB.Table("system_settings").Create(map[string]interface{}{
			"id":          utils.GenerateID(),
			"key":         "audit_log_retention",
			"value":       valueStr,
			"category":    "audit",
			"description": "Audit log retention configuration",
			"created_at":  now,
			"updated_at":  now,
		})
	} else {
		// Update existing
		config.DB.Table("system_settings").Where("key = ?", "audit_log_retention").Updates(map[string]interface{}{
			"value":      valueStr,
			"updated_at": now,
		})
	}

	// Log the change
	config.DB.Table("admin_audit_logs").Create(map[string]interface{}{
		"id":            utils.GenerateID(),
		"action":        "retention_settings_updated",
		"admin_user_id": userID,
		"new_value":     valueStr,
		"description":   fmt.Sprintf("Audit log retention settings updated by admin %s", userID),
		"created_at":    now,
	})

	defaults["updated_at"] = now
	defaults["updated_by"] = userID

	return utils.SendSimpleSuccess(c, defaults, "Retention settings updated successfully")
}

// getMapValueOrDefault returns the value from a map or the provided default.
func getMapValueOrDefault(m map[string]interface{}, key string, defaultVal interface{}) interface{} {
	if val, ok := m[key]; ok {
		return val
	}
	return defaultVal
}
