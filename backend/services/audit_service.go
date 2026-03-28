package services

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/logging"
	"github.com/liyali/liyali-gateway/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// AuditService handles audit logging and compliance features
type AuditService struct {
	// TODO: Add repository dependencies
}

// NewAuditService creates a new audit service
func NewAuditService() *AuditService {
	return &AuditService{}
}

// DocumentEvent holds all fields needed to write one audit_log row.
type DocumentEvent struct {
	OrganizationID string
	DocumentID     string
	DocumentType   string // "requisition", "purchase_order", "payment_voucher", "grn"
	UserID         string
	ActorName      string
	ActorRole      string
	Action         string                 // "created", "updated", "submitted", "approved", "rejected", "attachment_uploaded", ...
	Details        map[string]interface{} // arbitrary context; stored as JSONB
}

// LogDocumentEvent persists a single audit event for a document.
// It is safe to call inside or outside a transaction.
func LogDocumentEvent(db *gorm.DB, evt DocumentEvent) {
	var detailsJSON datatypes.JSON
	if len(evt.Details) > 0 {
		if b, err := json.Marshal(evt.Details); err == nil {
			detailsJSON = datatypes.JSON(b)
		}
	}

	record := &models.AuditLog{
		ID:             uuid.New().String(),
		OrganizationID: evt.OrganizationID,
		DocumentID:     evt.DocumentID,
		DocumentType:   evt.DocumentType,
		UserID:         evt.UserID,
		ActorName:      evt.ActorName,
		ActorRole:      evt.ActorRole,
		Action:         evt.Action,
		Details:        detailsJSON,
		CreatedAt:      time.Now(),
	}

	if err := db.Create(record).Error; err != nil {
		logging.WithFields(map[string]interface{}{
			"operation":     "log_document_event",
			"document_id":   evt.DocumentID,
			"document_type": evt.DocumentType,
			"action":        evt.Action,
			"error":         err.Error(),
		}).Error("audit_log_write_failed")
	}
}

// LogAuthEvent logs an authentication-related event
func (s *AuditService) LogAuthEvent(ctx context.Context, userID, email string, organizationID *string, action string, success bool, details, ipAddress, userAgent string) error {
	// TODO: Implement audit logging
	logging.WithFields(map[string]interface{}{
		"operation":       "audit_auth_event",
		"user_id":         userID,
		"action":          action,
		"success":         success,
		"details":         details,
		"ip_address":      ipAddress,
		"user_agent":      userAgent,
		"organization_id": organizationID,
	}).Info("audit_auth_event_logged")
	return nil
}

// LogEvent logs a general audit event
func (s *AuditService) LogEvent(ctx context.Context, userID, organizationID, action, resourceType, resourceID, details, ipAddress, userAgent string) error {
	// TODO: Implement audit logging
	logging.WithFields(map[string]interface{}{
		"operation":       "audit_event",
		"user_id":         userID,
		"organization_id": organizationID,
		"action":          action,
		"resource_type":   resourceType,
		"resource_id":     resourceID,
		"details":         details,
		"ip_address":      ipAddress,
		"user_agent":      userAgent,
	}).Info("audit_event_logged")
	return nil
}
