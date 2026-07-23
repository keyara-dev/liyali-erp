package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/utils"
)

// GetImpersonationLogs returns a paginated list of impersonation events.
// Accessible only to super_admin users (enforced by SuperAdminMiddleware on the route group).
//
// Query params:
//   impersonator_id  — filter by the user who performed the impersonation
//   target_id        — filter by the user who was impersonated
//   impersonation_type — "platform_user" | "admin_user"
//   revoked          — "true" | "false"
//   page, limit
func GetImpersonationLogs(c *fiber.Ctx) error {
	db := config.DB

	impersonatorID := c.Query("impersonator_id")
	targetID := c.Query("target_id")
	impersonationType := c.Query("impersonation_type")
	revokedParam := c.Query("revoked")
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	query := db.Table("impersonation_logs")

	if impersonatorID != "" {
		query = query.Where("impersonator_id = ?", impersonatorID)
	}
	if targetID != "" {
		query = query.Where("target_id = ?", targetID)
	}
	if impersonationType != "" {
		query = query.Where("impersonation_type = ?", impersonationType)
	}
	if revokedParam == "true" {
		query = query.Where("revoked = ?", true)
	} else if revokedParam == "false" {
		query = query.Where("revoked = ?", false)
	}

	var total int64
	query.Count(&total)

	var logs []map[string]interface{}
	if err := query.
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&logs).Error; err != nil {
		return utils.SendInternalError(c, "Failed to retrieve impersonation logs", err)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    logs,
		"meta": fiber.Map{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
		"message": "Impersonation logs retrieved successfully",
	})
}

// GetImpersonationLog returns a single impersonation log entry.
func GetImpersonationLog(c *fiber.Ctx) error {
	db := config.DB
	id := c.Params("id")

	var log map[string]interface{}
	if err := db.Table("impersonation_logs").Where("id = ?", id).First(&log).Error; err != nil {
		return utils.SendNotFound(c, "Impersonation log not found")
	}

	return utils.SendSimpleSuccess(c, log, "Impersonation log retrieved successfully")
}

// RevokeImpersonationLog marks an impersonation log entry as revoked.
// Note: this is a DB-level flag only — it cannot invalidate the issued JWT.
// The flag serves as an audit marker and can be used by security tooling.
func RevokeImpersonationLog(c *fiber.Ctx) error {
	db := config.DB
	id := c.Params("id")
	revokerID, _ := c.Locals("userID").(string)

	var existing map[string]interface{}
	if err := db.Table("impersonation_logs").Where("id = ?", id).First(&existing).Error; err != nil {
		return utils.SendNotFound(c, "Impersonation log not found")
	}

	if revoked, _ := existing["revoked"].(bool); revoked {
		return utils.SendBadRequest(c, "Impersonation log is already revoked")
	}

	now := time.Now()
	if err := db.Table("impersonation_logs").Where("id = ?", id).Updates(map[string]interface{}{
		"revoked":    true,
		"revoked_at": now,
		"revoked_by": revokerID,
	}).Error; err != nil {
		return utils.SendInternalError(c, "Failed to revoke impersonation log", err)
	}

	return utils.SendSimpleSuccess(c, map[string]interface{}{"id": id}, "Impersonation log revoked successfully")
}

// GetImpersonationStats returns aggregate statistics for impersonation events.
func GetImpersonationStats(c *fiber.Ctx) error {
	db := config.DB
	now := time.Now()

	var total, revoked, platformUser, adminUser int64
	db.Table("impersonation_logs").Count(&total)
	db.Table("impersonation_logs").Where("revoked = ?", true).Count(&revoked)
	db.Table("impersonation_logs").Where("impersonation_type = ?", "platform_user").Count(&platformUser)
	db.Table("impersonation_logs").Where("impersonation_type = ?", "admin_user").Count(&adminUser)

	// Active = not expired AND not revoked
	var active int64
	db.Table("impersonation_logs").
		Where("expires_at > ? AND revoked = ?", now, false).
		Count(&active)

	// Top impersonators (last 30 days)
	thirtyDaysAgo := now.AddDate(0, 0, -30)
	var topImpersonators []map[string]interface{}
	db.Table("impersonation_logs").
		Select("impersonator_id, impersonator_email, COUNT(*) as count").
		Where("created_at >= ?", thirtyDaysAgo).
		Group("impersonator_id, impersonator_email").
		Order("count DESC").
		Limit(5).
		Find(&topImpersonators)

	stats := map[string]interface{}{
		"total":         total,
		"active":        active,
		"revoked":       revoked,
		"platform_user": platformUser,
		"admin_user":    adminUser,
		"top_impersonators_30d": topImpersonators,
	}

	return utils.SendSimpleSuccess(c, stats, "Impersonation stats retrieved successfully")
}
