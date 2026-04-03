package utils

import (
	"encoding/json"
	"time"

	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/models"
	"gorm.io/datatypes"
)

// AuditAction represents the type of action being audited
type AuditAction string

const (
	// Document lifecycle actions
	AuditActionCreated   AuditAction = "created"
	AuditActionUpdated   AuditAction = "updated"
	AuditActionDeleted   AuditAction = "deleted"
	AuditActionSubmitted AuditAction = "submitted"
	AuditActionApproved  AuditAction = "approved"
	AuditActionRejected  AuditAction = "rejected"
	AuditActionWithdrawn AuditAction = "withdrawn"
	AuditActionCancelled AuditAction = "cancelled"
	
	// Attachment and document actions
	AuditActionAttachmentUploaded AuditAction = "attachment_uploaded"
	AuditActionAttachmentDeleted  AuditAction = "attachment_deleted"
	AuditActionQuotationUploaded  AuditAction = "quotation_uploaded"
	AuditActionQuotationUpdated   AuditAction = "quotation_updated"
	AuditActionQuotationDeleted   AuditAction = "quotation_deleted"
	
	// Field update actions
	AuditActionFieldUpdated       AuditAction = "field_updated"
	AuditActionMetadataUpdated    AuditAction = "metadata_updated"
	AuditActionStatusChanged      AuditAction = "status_changed"
	AuditActionPriorityChanged    AuditAction = "priority_changed"
	
	// Procurement flow actions
	AuditActionQuotationGateBypassed AuditAction = "quotation_gate_bypassed"
	AuditActionVendorSelected         AuditAction = "vendor_selected"
	AuditActionVendorChanged          AuditAction = "vendor_changed"
	
	// Payment actions
	AuditActionMarkedPaid   AuditAction = "marked_paid"
	AuditActionPaymentFailed AuditAction = "payment_failed"
	
	// GRN actions
	AuditActionGoodsReceived AuditAction = "goods_received"
	AuditActionGoodsConfirmed AuditAction = "goods_confirmed"
	AuditActionGoodsRejected  AuditAction = "goods_rejected"
)

// AuditLogParams contains parameters for creating an audit log entry
type AuditLogParams struct {
	OrganizationID string
	DocumentID     string
	DocumentType   string // requisition, purchase_order, payment_voucher, grn
	UserID         string
	ActorName      string
	ActorRole      string
	Action         AuditAction
	Changes        map[string]interface{} // Field-level changes: {"field": {"old": "value1", "new": "value2"}}
	Details        map[string]interface{} // Additional context
}

// CreateAuditLog creates a new audit log entry
func CreateAuditLog(params AuditLogParams) error {
	auditLog := models.AuditLog{
		ID:             GenerateID(),
		OrganizationID: params.OrganizationID,
		DocumentID:     params.DocumentID,
		DocumentType:   params.DocumentType,
		UserID:         params.UserID,
		ActorName:      params.ActorName,
		ActorRole:      params.ActorRole,
		Action:         string(params.Action),
		CreatedAt:      time.Now(),
	}

	// Convert changes to JSON
	if params.Changes != nil {
		auditLog.Changes = datatypes.NewJSONType(params.Changes)
	}

	// Convert details to JSON
	if params.Details != nil {
		detailsJSON, err := json.Marshal(params.Details)
		if err == nil {
			auditLog.Details = datatypes.JSON(detailsJSON)
		}
	}

	return config.DB.Create(&auditLog).Error
}

// LogDocumentUpdate logs a document update with field-level changes
func LogDocumentUpdate(organizationID, documentID, documentType, userID, actorName, actorRole string, changes map[string]interface{}) error {
	return CreateAuditLog(AuditLogParams{
		OrganizationID: organizationID,
		DocumentID:     documentID,
		DocumentType:   documentType,
		UserID:         userID,
		ActorName:      actorName,
		ActorRole:      actorRole,
		Action:         AuditActionUpdated,
		Changes:        changes,
	})
}

// LogAttachmentUpload logs when a supporting document is uploaded
func LogAttachmentUpload(organizationID, documentID, documentType, userID, actorName, actorRole, fileName string, fileSize int64) error {
	return CreateAuditLog(AuditLogParams{
		OrganizationID: organizationID,
		DocumentID:     documentID,
		DocumentType:   documentType,
		UserID:         userID,
		ActorName:      actorName,
		ActorRole:      actorRole,
		Action:         AuditActionAttachmentUploaded,
		Details: map[string]interface{}{
			"fileName": fileName,
			"fileSize": fileSize,
		},
	})
}

// LogAttachmentDelete logs when a supporting document is deleted
func LogAttachmentDelete(organizationID, documentID, documentType, userID, actorName, actorRole, fileName string) error {
	return CreateAuditLog(AuditLogParams{
		OrganizationID: organizationID,
		DocumentID:     documentID,
		DocumentType:   documentType,
		UserID:         userID,
		ActorName:      actorName,
		ActorRole:      actorRole,
		Action:         AuditActionAttachmentDeleted,
		Details: map[string]interface{}{
			"fileName": fileName,
		},
	})
}

// LogQuotationUpload logs when a quotation is uploaded
func LogQuotationUpload(organizationID, documentID, documentType, userID, actorName, actorRole, vendorName string, amount float64) error {
	return CreateAuditLog(AuditLogParams{
		OrganizationID: organizationID,
		DocumentID:     documentID,
		DocumentType:   documentType,
		UserID:         userID,
		ActorName:      actorName,
		ActorRole:      actorRole,
		Action:         AuditActionQuotationUploaded,
		Details: map[string]interface{}{
			"vendorName": vendorName,
			"amount":     amount,
		},
	})
}

// LogQuotationUpdate logs when a quotation is updated
func LogQuotationUpdate(organizationID, documentID, documentType, userID, actorName, actorRole, vendorName string, oldAmount, newAmount float64) error {
	return CreateAuditLog(AuditLogParams{
		OrganizationID: organizationID,
		DocumentID:     documentID,
		DocumentType:   documentType,
		UserID:         userID,
		ActorName:      actorName,
		ActorRole:      actorRole,
		Action:         AuditActionQuotationUpdated,
		Changes: map[string]interface{}{
			"amount": map[string]interface{}{
				"old": oldAmount,
				"new": newAmount,
			},
		},
		Details: map[string]interface{}{
			"vendorName": vendorName,
		},
	})
}

// LogQuotationDelete logs when a quotation is deleted
func LogQuotationDelete(organizationID, documentID, documentType, userID, actorName, actorRole, vendorName string) error {
	return CreateAuditLog(AuditLogParams{
		OrganizationID: organizationID,
		DocumentID:     documentID,
		DocumentType:   documentType,
		UserID:         userID,
		ActorName:      actorName,
		ActorRole:      actorRole,
		Action:         AuditActionQuotationDeleted,
		Details: map[string]interface{}{
			"vendorName": vendorName,
		},
	})
}

// LogStatusChange logs when a document status changes
func LogStatusChange(organizationID, documentID, documentType, userID, actorName, actorRole, oldStatus, newStatus string) error {
	return CreateAuditLog(AuditLogParams{
		OrganizationID: organizationID,
		DocumentID:     documentID,
		DocumentType:   documentType,
		UserID:         userID,
		ActorName:      actorName,
		ActorRole:      actorRole,
		Action:         AuditActionStatusChanged,
		Changes: map[string]interface{}{
			"status": map[string]interface{}{
				"old": oldStatus,
				"new": newStatus,
			},
		},
	})
}

// LogFieldChange logs when a specific field changes
func LogFieldChange(organizationID, documentID, documentType, userID, actorName, actorRole, fieldName string, oldValue, newValue interface{}) error {
	return CreateAuditLog(AuditLogParams{
		OrganizationID: organizationID,
		DocumentID:     documentID,
		DocumentType:   documentType,
		UserID:         userID,
		ActorName:      actorName,
		ActorRole:      actorRole,
		Action:         AuditActionFieldUpdated,
		Changes: map[string]interface{}{
			fieldName: map[string]interface{}{
				"old": oldValue,
				"new": newValue,
			},
		},
	})
}

// LogMetadataUpdate logs when document metadata is updated
func LogMetadataUpdate(organizationID, documentID, documentType, userID, actorName, actorRole string, changes map[string]interface{}) error {
	return CreateAuditLog(AuditLogParams{
		OrganizationID: organizationID,
		DocumentID:     documentID,
		DocumentType:   documentType,
		UserID:         userID,
		ActorName:      actorName,
		ActorRole:      actorRole,
		Action:         AuditActionMetadataUpdated,
		Changes:        changes,
	})
}

// CompareAndLogChanges compares old and new values and logs changes
func CompareAndLogChanges(organizationID, documentID, documentType, userID, actorName, actorRole string, oldValues, newValues map[string]interface{}) error {
	changes := make(map[string]interface{})
	
	for key, newVal := range newValues {
		oldVal, exists := oldValues[key]
		if !exists || oldVal != newVal {
			changes[key] = map[string]interface{}{
				"old": oldVal,
				"new": newVal,
			}
		}
	}
	
	if len(changes) == 0 {
		return nil // No changes to log
	}
	
	return LogDocumentUpdate(organizationID, documentID, documentType, userID, actorName, actorRole, changes)
}
