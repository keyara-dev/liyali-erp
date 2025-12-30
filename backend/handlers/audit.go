package handlers

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/utils"
)

// GetAuditLogs returns audit logs with pagination and filtering
func GetAuditLogs(c *fiber.Ctx) error {
	// Parse query parameters
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 50)
	action := c.Query("action")
	documentType := c.Query("documentType")
	userID := c.Query("userId")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	// Build query - Note: AuditLog doesn't have organization_id in current model
	query := config.DB.Model(&models.AuditLog{})
	
	if action != "" {
		query = query.Where("action = ?", action)
	}
	if documentType != "" {
		query = query.Where("document_type = ?", documentType)
	}
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		log.Printf("Error counting audit logs: %v", err)
		return utils.SendInternalError(c, "Failed to count audit logs", err)
	}

	// Get paginated results
	var auditLogs []models.AuditLog
	offset := (page - 1) * limit
	if err := query.
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&auditLogs).Error; err != nil {
		log.Printf("Error fetching audit logs: %v", err)
		return utils.SendInternalError(c, "Failed to fetch audit logs", err)
	}

	// Convert to response format
	responses := make([]map[string]interface{}, 0, len(auditLogs))
	for _, auditLog := range auditLogs {
		responses = append(responses, map[string]interface{}{
			"id":           auditLog.ID,
			"documentId":   auditLog.DocumentID,
			"documentType": auditLog.DocumentType,
			"userId":       auditLog.UserID,
			"action":       auditLog.Action,
			"changes":      auditLog.Changes,
			"createdAt":    auditLog.CreatedAt,
		})
	}

	return utils.SendPaginatedSuccess(c, responses, "Audit logs retrieved successfully", page, limit, total)
}

// GetDocumentAuditLogs returns audit logs for a specific document
func GetDocumentAuditLogs(c *fiber.Ctx) error {
	documentID := c.Params("documentId")
	if documentID == "" {
		return utils.SendBadRequestError(c, "Document ID is required")
	}

	// Parse query parameters
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 50)

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	// Build query for document-specific logs
	query := config.DB.Where("document_id = ?", documentID)

	// Get total count
	var total int64
	if err := query.Model(&models.AuditLog{}).Count(&total).Error; err != nil {
		log.Printf("Error counting document audit logs: %v", err)
		return utils.SendInternalError(c, "Failed to count audit logs", err)
	}

	// Get paginated results
	var auditLogs []models.AuditLog
	offset := (page - 1) * limit
	if err := query.
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&auditLogs).Error; err != nil {
		log.Printf("Error fetching document audit logs: %v", err)
		return utils.SendInternalError(c, "Failed to fetch audit logs", err)
	}

	// Convert to response format
	responses := make([]map[string]interface{}, 0, len(auditLogs))
	for _, auditLog := range auditLogs {
		responses = append(responses, map[string]interface{}{
			"id":           auditLog.ID,
			"documentId":   auditLog.DocumentID,
			"documentType": auditLog.DocumentType,
			"userId":       auditLog.UserID,
			"action":       auditLog.Action,
			"changes":      auditLog.Changes,
			"createdAt":    auditLog.CreatedAt,
		})
	}

	return utils.SendPaginatedSuccess(c, responses, "Document audit logs retrieved successfully", page, limit, total)
}