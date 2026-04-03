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
	Changes        map[string]interface{} // field-level changes: {"field": {"old": "value1", "new": "value2"}}
	Details        map[string]interface{} // arbitrary context; stored as JSONB
	Snapshot       map[string]interface{} // complete snapshot of document state after change
}

// FieldChange represents a change to a specific field
type FieldChange struct {
	Field    string      `json:"field"`
	OldValue interface{} `json:"oldValue"`
	NewValue interface{} `json:"newValue"`
	Changed  bool        `json:"changed"`
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

	var changesJSON datatypes.JSONType[map[string]interface{}]
	if len(evt.Changes) > 0 {
		changesJSON = datatypes.NewJSONType(evt.Changes)
	}

	// Include snapshot in details if provided
	if len(evt.Snapshot) > 0 {
		if len(evt.Details) == 0 {
			evt.Details = make(map[string]interface{})
		}
		evt.Details["snapshot"] = evt.Snapshot
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
		Changes:        changesJSON,
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

// CompareAndBuildChanges compares old and new values and builds a changes map
func CompareAndBuildChanges(oldValues, newValues map[string]interface{}) map[string]interface{} {
	changes := make(map[string]interface{})
	
	// Check for changed and new fields
	for key, newVal := range newValues {
		oldVal, exists := oldValues[key]
		
		// Skip if values are equal
		if exists {
			// Convert to JSON for comparison to handle complex types
			oldJSON, _ := json.Marshal(oldVal)
			newJSON, _ := json.Marshal(newVal)
			if string(oldJSON) == string(newJSON) {
				continue
			}
		}
		
		changes[key] = map[string]interface{}{
			"old": oldVal,
			"new": newVal,
		}
	}
	
	// Check for deleted fields
	for key, oldVal := range oldValues {
		if _, exists := newValues[key]; !exists {
			changes[key] = map[string]interface{}{
				"old": oldVal,
				"new": nil,
			}
		}
	}
	
	return changes
}

// CreateDocumentSnapshot creates a snapshot of the current document state
func CreateDocumentSnapshot(doc interface{}) map[string]interface{} {
	snapshot := make(map[string]interface{})
	
	// Convert document to JSON and back to map for snapshot
	docJSON, err := json.Marshal(doc)
	if err != nil {
		return snapshot
	}
	
	json.Unmarshal(docJSON, &snapshot)
	
	// Add timestamp
	snapshot["snapshotTimestamp"] = time.Now().Format(time.RFC3339)
	
	return snapshot
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

// LogAttachmentUpload logs when a supporting document is uploaded
func LogAttachmentUpload(db *gorm.DB, organizationID, documentID, documentType, userID, actorName, actorRole, fileName string, fileSize int64) {
	LogDocumentEvent(db, DocumentEvent{
		OrganizationID: organizationID,
		DocumentID:     documentID,
		DocumentType:   documentType,
		UserID:         userID,
		ActorName:      actorName,
		ActorRole:      actorRole,
		Action:         "attachment_uploaded",
		Details: map[string]interface{}{
			"fileName": fileName,
			"fileSize": fileSize,
		},
	})
}

// LogAttachmentDelete logs when a supporting document is deleted
func LogAttachmentDelete(db *gorm.DB, organizationID, documentID, documentType, userID, actorName, actorRole, fileName string) {
	LogDocumentEvent(db, DocumentEvent{
		OrganizationID: organizationID,
		DocumentID:     documentID,
		DocumentType:   documentType,
		UserID:         userID,
		ActorName:      actorName,
		ActorRole:      actorRole,
		Action:         "attachment_deleted",
		Details: map[string]interface{}{
			"fileName": fileName,
		},
	})
}

// LogQuotationUpload logs when a quotation is uploaded
func LogQuotationUpload(db *gorm.DB, organizationID, documentID, documentType, userID, actorName, actorRole, vendorName string, amount float64, currency string) {
	LogDocumentEvent(db, DocumentEvent{
		OrganizationID: organizationID,
		DocumentID:     documentID,
		DocumentType:   documentType,
		UserID:         userID,
		ActorName:      actorName,
		ActorRole:      actorRole,
		Action:         "quotation_uploaded",
		Details: map[string]interface{}{
			"vendorName": vendorName,
			"amount":     amount,
			"currency":   currency,
		},
	})
}

// LogQuotationUpdate logs when a quotation is updated
func LogQuotationUpdate(db *gorm.DB, organizationID, documentID, documentType, userID, actorName, actorRole, vendorName string, oldAmount, newAmount float64, currency string) {
	LogDocumentEvent(db, DocumentEvent{
		OrganizationID: organizationID,
		DocumentID:     documentID,
		DocumentType:   documentType,
		UserID:         userID,
		ActorName:      actorName,
		ActorRole:      actorRole,
		Action:         "quotation_updated",
		Changes: map[string]interface{}{
			"amount": map[string]interface{}{
				"old": oldAmount,
				"new": newAmount,
			},
		},
		Details: map[string]interface{}{
			"vendorName": vendorName,
			"currency":   currency,
		},
	})
}

// LogQuotationDelete logs when a quotation is deleted
func LogQuotationDelete(db *gorm.DB, organizationID, documentID, documentType, userID, actorName, actorRole, vendorName string) {
	LogDocumentEvent(db, DocumentEvent{
		OrganizationID: organizationID,
		DocumentID:     documentID,
		DocumentType:   documentType,
		UserID:         userID,
		ActorName:      actorName,
		ActorRole:      actorRole,
		Action:         "quotation_deleted",
		Details: map[string]interface{}{
			"vendorName": vendorName,
		},
	})
}

// LogStatusChange logs when a document status changes
func LogStatusChange(db *gorm.DB, organizationID, documentID, documentType, userID, actorName, actorRole, oldStatus, newStatus string) {
	LogDocumentEvent(db, DocumentEvent{
		OrganizationID: organizationID,
		DocumentID:     documentID,
		DocumentType:   documentType,
		UserID:         userID,
		ActorName:      actorName,
		ActorRole:      actorRole,
		Action:         "status_changed",
		Changes: map[string]interface{}{
			"status": map[string]interface{}{
				"old": oldStatus,
				"new": newStatus,
			},
		},
	})
}

// LogFieldChange logs when a specific field changes
func LogFieldChange(db *gorm.DB, organizationID, documentID, documentType, userID, actorName, actorRole, fieldName string, oldValue, newValue interface{}) {
	LogDocumentEvent(db, DocumentEvent{
		OrganizationID: organizationID,
		DocumentID:     documentID,
		DocumentType:   documentType,
		UserID:         userID,
		ActorName:      actorName,
		ActorRole:      actorRole,
		Action:         "field_updated",
		Changes: map[string]interface{}{
			fieldName: map[string]interface{}{
				"old": oldValue,
				"new": newValue,
			},
		},
	})
}

// LogMetadataUpdate logs when document metadata is updated
func LogMetadataUpdate(db *gorm.DB, organizationID, documentID, documentType, userID, actorName, actorRole string, changes map[string]interface{}) {
	LogDocumentEvent(db, DocumentEvent{
		OrganizationID: organizationID,
		DocumentID:     documentID,
		DocumentType:   documentType,
		UserID:         userID,
		ActorName:      actorName,
		ActorRole:      actorRole,
		Action:         "metadata_updated",
		Changes:        changes,
	})
}
