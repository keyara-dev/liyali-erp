package handlers

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/utils"
)

func auditLogToMap(a models.AuditLog) map[string]interface{} {
	return map[string]interface{}{
		"id":             a.ID,
		"organizationId": a.OrganizationID,
		"documentId":     a.DocumentID,
		"documentType":   a.DocumentType,
		"userId":         a.UserID,
		"actorName":      a.ActorName,
		"actorRole":      a.ActorRole,
		"action":         a.Action,
		"details":        a.Details,
		"changes":        a.Changes,
		"createdAt":      a.CreatedAt,
	}
}

// GetAuditLogs returns org-scoped audit logs with pagination and filtering
func GetAuditLogs(c *fiber.Ctx) error {
	orgID := c.Locals("organizationID").(string)

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

	query := config.DB.Model(&models.AuditLog{}).Where("organization_id = ?", orgID)

	if action != "" {
		query = query.Where("action = ?", action)
	}
	if documentType != "" {
		query = query.Where("document_type = ?", documentType)
	}
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		log.Printf("Error counting audit logs: %v", err)
		return utils.SendInternalError(c, "Failed to count audit logs", err)
	}

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

	responses := make([]map[string]interface{}, 0, len(auditLogs))
	for _, a := range auditLogs {
		responses = append(responses, auditLogToMap(a))
	}

	return utils.SendPaginatedSuccess(c, responses, "Audit logs retrieved successfully", page, limit, total)
}

// GetDocumentAuditLogs returns audit logs for a specific document (org-scoped)
func GetDocumentAuditLogs(c *fiber.Ctx) error {
	orgID := c.Locals("organizationID").(string)
	documentID := c.Params("documentId")
	if documentID == "" {
		return utils.SendBadRequestError(c, "Document ID is required")
	}

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 100)

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 100
	}

	query := config.DB.Where("organization_id = ? AND document_id = ?", orgID, documentID)

	var total int64
	if err := query.Model(&models.AuditLog{}).Count(&total).Error; err != nil {
		log.Printf("Error counting document audit logs: %v", err)
		return utils.SendInternalError(c, "Failed to count audit logs", err)
	}

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

	responses := make([]map[string]interface{}, 0, len(auditLogs))
	for _, a := range auditLogs {
		responses = append(responses, auditLogToMap(a))
	}

	return utils.SendPaginatedSuccess(c, responses, "Document audit logs retrieved successfully", page, limit, total)
}

// GetDocumentAuditEvents returns audit events for a specific document by entityType + entityId query params.
// Route: GET /api/v1/audit-events?entityType=purchase_order&entityId=<id>
func GetDocumentAuditEvents(c *fiber.Ctx) error {
	orgID := c.Locals("organizationID").(string)
	entityType := c.Query("entityType")
	entityID := c.Query("entityId")

	if entityType == "" || entityID == "" {
		return utils.SendBadRequestError(c, "entityType and entityId query parameters are required")
	}

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 100)
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 100
	}

	query := config.DB.Where(
		"organization_id = ? AND document_type = ? AND document_id = ?",
		orgID, entityType, entityID,
	)

	var total int64
	if err := query.Model(&models.AuditLog{}).Count(&total).Error; err != nil {
		return utils.SendInternalError(c, "Failed to count audit events", err)
	}

	var auditLogs []models.AuditLog
	offset := (page - 1) * limit
	if err := query.
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&auditLogs).Error; err != nil {
		return utils.SendInternalError(c, "Failed to fetch audit events", err)
	}

	responses := make([]map[string]interface{}, 0, len(auditLogs))
	for _, a := range auditLogs {
		responses = append(responses, auditLogToMap(a))
	}

	return utils.SendPaginatedSuccess(c, responses, "Audit events retrieved successfully", page, limit, total)
}
