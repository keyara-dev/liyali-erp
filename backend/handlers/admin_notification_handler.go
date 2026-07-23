package handlers

import (
	"log"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/utils"
	"github.com/google/uuid"
)

// GetAdminNotifications returns all notifications across the platform with filtering
func GetAdminNotifications(c *fiber.Ctx) error {
	db := config.DB

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	notificationType := c.Query("type", "")
	search := c.Query("search", "")
	status := c.Query("status", "") // read, unread, all

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	query := db.Model(&models.Notification{})

	if notificationType != "" {
		query = query.Where("type = ?", notificationType)
	}
	if search != "" {
		query = query.Where("subject ILIKE ? OR body ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	if status == "read" {
		query = query.Where("sent = ?", true)
	} else if status == "unread" {
		query = query.Where("sent = ?", false)
	}

	var total int64
	query.Count(&total)

	var notifications []models.Notification
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&notifications).Error; err != nil {
		log.Printf("Error fetching admin notifications: %v", err)
		return utils.SendInternalError(c, "Failed to fetch notifications", err)
	}

	result := make([]map[string]interface{}, 0, len(notifications))
	for _, n := range notifications {
		result = append(result, map[string]interface{}{
			"id":              n.ID,
			"organization_id": n.OrganizationID,
			"recipient_id":    n.RecipientID,
			"type":            n.Type,
			"subject":         n.Subject,
			"body":            n.Body,
			"document_id":     n.DocumentID,
			"document_type":   n.DocumentType,
			"is_read":         n.Sent,
			"read_at":         n.SentAt,
			"importance":      n.Importance,
			"created_at":      n.CreatedAt,
			"updated_at":      n.UpdatedAt,
		})
	}

	return utils.SendPaginatedSuccess(c, result, "Admin notifications retrieved successfully", page, limit, total)
}

// GetAdminNotificationStats returns platform-wide notification statistics
func GetAdminNotificationStats(c *fiber.Ctx) error {
	db := config.DB

	var totalCount int64
	var unreadCount int64
	var readCount int64
	var todayCount int64

	db.Model(&models.Notification{}).Count(&totalCount)
	db.Model(&models.Notification{}).Where("sent = ?", false).Count(&unreadCount)
	db.Model(&models.Notification{}).Where("sent = ?", true).Count(&readCount)

	today := time.Now().Truncate(24 * time.Hour)
	db.Model(&models.Notification{}).Where("created_at >= ?", today).Count(&todayCount)

	// Count by type
	type typeCount struct {
		Type  string `gorm:"column:type"`
		Count int64  `gorm:"column:count"`
	}
	var typeCounts []typeCount
	db.Model(&models.Notification{}).Select("type, COUNT(*) as count").Group("type").Scan(&typeCounts)

	byType := make(map[string]int64)
	for _, tc := range typeCounts {
		byType[tc.Type] = tc.Count
	}

	stats := map[string]interface{}{
		"total":        totalCount,
		"unread":       unreadCount,
		"read":         readCount,
		"today":        todayCount,
		"by_type":      byType,
		"collected_at": time.Now().Format(time.RFC3339),
	}

	return utils.SendSimpleSuccess(c, stats, "Notification stats retrieved successfully")
}

// CreateAdminNotification creates a broadcast notification for admin use
func CreateAdminNotification(c *fiber.Ctx) error {
	db := config.DB

	type createRequest struct {
		Subject        string   `json:"subject"`
		Body           string   `json:"body"`
		Type           string   `json:"type"`
		Importance     string   `json:"importance"`
		RecipientIDs   []string `json:"recipient_ids"`   // specific users
		OrganizationID string   `json:"organization_id"` // all users in org
		Broadcast      bool     `json:"broadcast"`       // all users
	}

	var req createRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendBadRequest(c, "Invalid request body")
	}

	if req.Subject == "" || req.Body == "" {
		return utils.SendBadRequest(c, "Subject and body are required")
	}

	if req.Type == "" {
		req.Type = "admin_announcement"
	}
	if req.Importance == "" {
		req.Importance = "MEDIUM"
	}

	// Determine recipients
	var recipientIDs []string

	if req.Broadcast {
		// Get all active users
		db.Model(&models.User{}).Where("status = ?", "active").Pluck("id", &recipientIDs)
	} else if req.OrganizationID != "" {
		// Get all users in the organization
		db.Model(&models.User{}).
			Joins("JOIN user_organizations ON users.id = user_organizations.user_id").
			Where("user_organizations.organization_id = ? AND users.status = ?", req.OrganizationID, "active").
			Pluck("users.id", &recipientIDs)
	} else if len(req.RecipientIDs) > 0 {
		recipientIDs = req.RecipientIDs
	} else {
		return utils.SendBadRequest(c, "Must specify recipients: recipient_ids, organization_id, or broadcast=true")
	}

	if len(recipientIDs) == 0 {
		return utils.SendBadRequest(c, "No recipients found matching criteria")
	}

	// Create notifications in batch
	now := time.Now()
	notifications := make([]models.Notification, 0, len(recipientIDs))
	for _, rid := range recipientIDs {
		notifications = append(notifications, models.Notification{
			ID:          uuid.New().String(),
			RecipientID: rid,
			Type:        req.Type,
			Subject:     req.Subject,
			Body:        req.Body,
			Importance:  req.Importance,
			Sent:        false,
			CreatedAt:   now,
			UpdatedAt:   now,
		})
	}

	if err := db.CreateInBatches(notifications, 100).Error; err != nil {
		log.Printf("Error creating admin notifications: %v", err)
		return utils.SendInternalError(c, "Failed to create notifications", err)
	}

	return utils.SendSimpleSuccess(c, map[string]interface{}{
		"created_count":  len(notifications),
		"recipient_count": len(recipientIDs),
	}, "Notifications created successfully")
}

// DeleteAdminNotification deletes a notification by ID (admin)
func DeleteAdminNotification(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return utils.SendBadRequest(c, "Notification ID is required")
	}

	result := config.DB.Where("id = ?", id).Delete(&models.Notification{})
	if result.Error != nil {
		log.Printf("Error deleting notification %s: %v", id, result.Error)
		return utils.SendInternalError(c, "Failed to delete notification", result.Error)
	}
	if result.RowsAffected == 0 {
		return utils.SendNotFound(c, "Notification not found")
	}

	return utils.SendSimpleSuccess(c, map[string]interface{}{
		"deleted_id": id,
	}, "Notification deleted successfully")
}

// BulkDeleteAdminNotifications deletes multiple notifications
func BulkDeleteAdminNotifications(c *fiber.Ctx) error {
	type bulkRequest struct {
		IDs []string `json:"ids"`
	}

	var req bulkRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendBadRequest(c, "Invalid request body")
	}

	if len(req.IDs) == 0 {
		return utils.SendBadRequest(c, "At least one notification ID is required")
	}

	result := config.DB.Where("id IN ?", req.IDs).Delete(&models.Notification{})
	if result.Error != nil {
		log.Printf("Error bulk deleting notifications: %v", result.Error)
		return utils.SendInternalError(c, "Failed to delete notifications", result.Error)
	}

	return utils.SendSimpleSuccess(c, map[string]interface{}{
		"deleted_count": result.RowsAffected,
	}, "Notifications deleted successfully")
}

// MarkAdminNotificationRead marks a notification as read (admin)
func MarkAdminNotificationRead(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return utils.SendBadRequest(c, "Notification ID is required")
	}

	now := time.Now()
	result := config.DB.Model(&models.Notification{}).Where("id = ?", id).Updates(map[string]interface{}{
		"sent":    true,
		"sent_at": now,
	})

	if result.Error != nil {
		return utils.SendInternalError(c, "Failed to mark notification as read", result.Error)
	}
	if result.RowsAffected == 0 {
		return utils.SendNotFound(c, "Notification not found")
	}

	return utils.SendSimpleSuccess(c, map[string]interface{}{"id": id}, "Notification marked as read")
}
