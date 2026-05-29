package handlers

import (
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/liyali/liyali-gateway/config"
	db "github.com/liyali/liyali-gateway/database/sqlc"
	"github.com/liyali/liyali-gateway/utils"
)

// auditLogRow is a minimal projection used by the gorm fallback path so we
// don't depend on sqlc-generated Postgres types when the sqlc Queries handle
// hasn't been initialised (e.g. inside the SQLite-backed test harness).
type auditLogRow struct {
	ID             string
	OrganizationID string
	DocumentID     string
	DocumentType   string
	UserID         string
	ActorName      string
	ActorRole      string
	Action         string
	Details        []byte
	Changes        []byte
	CreatedAt      time.Time
}

func (r auditLogRow) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":             r.ID,
		"organizationId": r.OrganizationID,
		"documentId":     r.DocumentID,
		"documentType":   r.DocumentType,
		"userId":         r.UserID,
		"actorName":      r.ActorName,
		"actorRole":      r.ActorRole,
		"action":         r.Action,
		"details":        json.RawMessage(r.Details),
		"changes":        json.RawMessage(r.Changes),
		"createdAt":      r.CreatedAt,
	}
}

func auditRowToMap(row db.ListAuditLogsRow) map[string]interface{} {
	return map[string]interface{}{
		"id":             row.ID,
		"organizationId": row.OrganizationID,
		"documentId":     row.DocumentID,
		"documentType":   row.DocumentType,
		"userId":         row.UserID,
		"actorName":      row.ActorName,
		"actorRole":      row.ActorRole,
		"action":         row.Action,
		"details":        json.RawMessage(row.Details),
		"changes":        json.RawMessage(row.Changes),
		"createdAt":      row.CreatedAt.Time,
	}
}

func auditDocRowToMap(row db.ListDocumentAuditLogsRow) map[string]interface{} {
	return map[string]interface{}{
		"id":             row.ID,
		"organizationId": row.OrganizationID,
		"documentId":     row.DocumentID,
		"documentType":   row.DocumentType,
		"userId":         row.UserID,
		"actorName":      row.ActorName,
		"actorRole":      row.ActorRole,
		"action":         row.Action,
		"details":        json.RawMessage(row.Details),
		"changes":        json.RawMessage(row.Changes),
		"createdAt":      row.CreatedAt.Time,
	}
}

func auditEventRowToMap(row db.ListAuditEventsRow) map[string]interface{} {
	return map[string]interface{}{
		"id":             row.ID,
		"organizationId": row.OrganizationID,
		"documentId":     row.DocumentID,
		"documentType":   row.DocumentType,
		"userId":         row.UserID,
		"actorName":      row.ActorName,
		"actorRole":      row.ActorRole,
		"action":         row.Action,
		"details":        json.RawMessage(row.Details),
		"changes":        json.RawMessage(row.Changes),
		"createdAt":      row.CreatedAt.Time,
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

	ctx := c.Context()

	// sqlc.Queries is wired against pgx in production but is nil in the
	// SQLite-backed unit-test harness. Fall through to a gorm-based path so
	// both code paths exercise the same SQL semantics.
	if config.Queries == nil {
		total, rows, err := listAuditLogsGorm(orgID, action, documentType, userID, page, limit)
		if err != nil {
			return utils.SendInternalError(c, "Failed to fetch audit logs", err)
		}
		responses := make([]map[string]interface{}, 0, len(rows))
		for _, row := range rows {
			responses = append(responses, row.ToMap())
		}
		return utils.SendPaginatedSuccess(c, responses, "Audit logs retrieved successfully", page, limit, total)
	}

	total, err := config.Queries.CountAuditLogs(ctx, orgID, action, documentType, userID)
	if err != nil {
		return utils.SendInternalError(c, "Failed to count audit logs", err)
	}

	offset := int32((page - 1) * limit)
	rows, err := config.Queries.ListAuditLogs(ctx, db.ListAuditLogsParams{
		OrganizationID: orgID,
		Column2:        action,
		Column3:        documentType,
		Column4:        userID,
		Limit:          int32(limit),
		Offset:         offset,
	})
	if err != nil {
		return utils.SendInternalError(c, "Failed to fetch audit logs", err)
	}

	responses := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		responses = append(responses, auditRowToMap(row))
	}

	return utils.SendPaginatedSuccess(c, responses, "Audit logs retrieved successfully", page, limit, total)
}

// listAuditLogsGorm is the sqlc-free implementation used by tests and by
// callers that haven't initialised config.Queries.
func listAuditLogsGorm(orgID, action, documentType, userID string, page, limit int) (int64, []auditLogRow, error) {
	q := config.DB.Table("audit_logs").Where("organization_id = ?", orgID)
	if action != "" {
		q = q.Where("action = ?", action)
	}
	if documentType != "" {
		q = q.Where("document_type = ?", documentType)
	}
	if userID != "" {
		q = q.Where("user_id = ?", userID)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return 0, nil, err
	}

	offset := (page - 1) * limit
	var rows []auditLogRow
	if err := q.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Select("id, organization_id, document_id, document_type, user_id, actor_name, actor_role, action, details, changes, created_at").
		Find(&rows).Error; err != nil {
		return 0, nil, err
	}
	return total, rows, nil
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

	ctx := c.Context()

	if config.Queries == nil {
		total, rows, err := listDocumentAuditLogsGorm(orgID, documentID, page, limit)
		if err != nil {
			return utils.SendInternalError(c, "Failed to fetch audit logs", err)
		}
		responses := make([]map[string]interface{}, 0, len(rows))
		for _, row := range rows {
			responses = append(responses, row.ToMap())
		}
		return utils.SendPaginatedSuccess(c, responses, "Document audit logs retrieved successfully", page, limit, total)
	}

	total, err := config.Queries.CountDocumentAuditLogs(ctx, orgID, documentID)
	if err != nil {
		return utils.SendInternalError(c, "Failed to count audit logs", err)
	}

	offset := int32((page - 1) * limit)
	rows, err := config.Queries.ListDocumentAuditLogs(ctx, orgID, documentID, int32(limit), offset)
	if err != nil {
		return utils.SendInternalError(c, "Failed to fetch audit logs", err)
	}

	responses := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		responses = append(responses, auditDocRowToMap(row))
	}

	return utils.SendPaginatedSuccess(c, responses, "Document audit logs retrieved successfully", page, limit, total)
}

// listDocumentAuditLogsGorm is the sqlc-free implementation for the per-document
// audit log list. Used by tests and as a fallback when config.Queries is nil.
func listDocumentAuditLogsGorm(orgID, documentID string, page, limit int) (int64, []auditLogRow, error) {
	q := config.DB.Table("audit_logs").
		Where("organization_id = ? AND document_id = ?", orgID, documentID)

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return 0, nil, err
	}

	offset := (page - 1) * limit
	var rows []auditLogRow
	if err := q.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Select("id, organization_id, document_id, document_type, user_id, actor_name, actor_role, action, details, changes, created_at").
		Find(&rows).Error; err != nil {
		return 0, nil, err
	}
	return total, rows, nil
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

	ctx := c.Context()

	total, err := config.Queries.CountAuditEvents(ctx, orgID, entityType, entityID)
	if err != nil {
		return utils.SendInternalError(c, "Failed to count audit events", err)
	}

	offset := int32((page - 1) * limit)
	rows, err := config.Queries.ListAuditEvents(ctx, orgID, entityType, entityID, int32(limit), offset)
	if err != nil {
		return utils.SendInternalError(c, "Failed to fetch audit events", err)
	}

	responses := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		responses = append(responses, auditEventRowToMap(row))
	}

	return utils.SendPaginatedSuccess(c, responses, "Audit events retrieved successfully", page, limit, total)
}
