package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/repository"
	"gorm.io/datatypes"
)

// DocumentService handles generic document business logic
type DocumentService struct {
	documentRepo repository.DocumentRepositoryInterface
	auditService *AuditService
}

// CreateDocumentRequest represents a document creation request
type CreateDocumentRequest struct {
	DocumentType string                 `json:"documentType" validate:"required"`
	Title        string                 `json:"title" validate:"required"`
	Description  string                 `json:"description"`
	Amount       float64                `json:"amount"`
	Currency     string                 `json:"currency"`
	Department   string                 `json:"department"`
	WorkflowID   *uuid.UUID             `json:"workflowId"`
	Data         map[string]interface{} `json:"data" validate:"required"` // Type-specific fields
	Metadata     map[string]interface{} `json:"metadata"`                 // Additional metadata
}

// UpdateDocumentRequest represents a document update request
type UpdateDocumentRequest struct {
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Amount      float64                `json:"amount"`
	Currency    string                 `json:"currency"`
	Department  string                 `json:"department"`
	Data        map[string]interface{} `json:"data"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// NewDocumentService creates a new document service
func NewDocumentService(documentRepo repository.DocumentRepositoryInterface, auditService *AuditService) *DocumentService {
	return &DocumentService{
		documentRepo: documentRepo,
		auditService: auditService,
	}
}

// CreateDocument creates a new generic document
func (s *DocumentService) CreateDocument(ctx context.Context, organizationID, userID string, req CreateDocumentRequest) (*models.Document, error) {
	// Validate document type
	validTypes := map[string]bool{
		"REQUISITION":     true,
		"BUDGET":          true,
		"PURCHASE_ORDER":  true,
		"PAYMENT_VOUCHER": true,
		"GRN":             true,
		"CATEGORY":        true,
		"VENDOR":          true,
	}
	
	if !validTypes[req.DocumentType] {
		return nil, fmt.Errorf("invalid document type: %s", req.DocumentType)
	}

	// Convert data to JSON
	dataJSON, err := json.Marshal(req.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal document data: %w", err)
	}

	var metadataJSON datatypes.JSON
	if req.Metadata != nil {
		metadataBytes, err := json.Marshal(req.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal document metadata: %w", err)
		}
		metadataJSON = datatypes.JSON(metadataBytes)
	}

	// Create document
	document := &models.Document{
		OrganizationID: organizationID,
		DocumentType:   req.DocumentType,
		Title:          req.Title,
		Status:         "draft",
		CreatedBy:      userID,
		WorkflowID:     req.WorkflowID,
		Data:           datatypes.JSON(dataJSON),
		Metadata:       metadataJSON,
	}

	// Set optional fields
	if req.Description != "" {
		document.Description = &req.Description
	}
	if req.Amount > 0 {
		document.Amount = &req.Amount
	}
	if req.Currency != "" {
		document.Currency = &req.Currency
	} else {
		defaultCurrency := "USD"
		document.Currency = &defaultCurrency
	}
	if req.Department != "" {
		document.Department = &req.Department
	}

	// Create document
	document, err = s.documentRepo.Create(ctx, document)
	if err != nil {
		return nil, fmt.Errorf("failed to create document: %w", err)
	}

	// Log audit event
	if s.auditService != nil {
		details := fmt.Sprintf("Created %s document '%s'", req.DocumentType, req.Title)
		s.auditService.LogEvent(ctx, userID, organizationID, "document_created", "document", document.ID.String(), details, "", "")
	}

	return document, nil
}

// GetDocument retrieves a document by ID
func (s *DocumentService) GetDocument(ctx context.Context, id uuid.UUID, organizationID string) (*models.Document, error) {
	return s.documentRepo.GetByID(ctx, id, organizationID)
}

// GetDocumentByNumber retrieves a document by document number
func (s *DocumentService) GetDocumentByNumber(ctx context.Context, documentNumber, organizationID string) (*models.Document, error) {
	return s.documentRepo.GetByNumber(ctx, documentNumber, organizationID)
}

// UpdateDocument updates a document
func (s *DocumentService) UpdateDocument(ctx context.Context, id uuid.UUID, organizationID, userID string, req UpdateDocumentRequest) (*models.Document, error) {
	// Get existing document
	document, err := s.documentRepo.GetByID(ctx, id, organizationID)
	if err != nil {
		return nil, fmt.Errorf("document not found: %w", err)
	}

	// Check if document is editable
	if !document.IsEditable() {
		return nil, fmt.Errorf("document cannot be edited in %s status", document.Status)
	}

	// Update fields
	if req.Title != "" {
		document.Title = req.Title
	}
	if req.Description != "" {
		document.Description = &req.Description
	}
	if req.Amount > 0 {
		document.Amount = &req.Amount
	}
	if req.Currency != "" {
		document.Currency = &req.Currency
	}
	if req.Department != "" {
		document.Department = &req.Department
	}

	// Update data if provided
	if req.Data != nil {
		dataJSON, err := json.Marshal(req.Data)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal document data: %w", err)
		}
		document.Data = datatypes.JSON(dataJSON)
	}

	// Update metadata if provided
	if req.Metadata != nil {
		metadataJSON, err := json.Marshal(req.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal document metadata: %w", err)
		}
		document.Metadata = datatypes.JSON(metadataJSON)
	}

	document.UpdatedBy = &userID

	// Update document
	document, err = s.documentRepo.Update(ctx, document)
	if err != nil {
		return nil, fmt.Errorf("failed to update document: %w", err)
	}

	// Log audit event
	if s.auditService != nil {
		details := fmt.Sprintf("Updated %s document '%s'", document.DocumentType, document.Title)
		s.auditService.LogEvent(ctx, userID, organizationID, "document_updated", "document", document.ID.String(), details, "", "")
	}

	return document, nil
}

// DeleteDocument deletes a document
func (s *DocumentService) DeleteDocument(ctx context.Context, id uuid.UUID, organizationID, userID string) error {
	// Get existing document for audit logging
	document, err := s.documentRepo.GetByID(ctx, id, organizationID)
	if err != nil {
		return fmt.Errorf("document not found: %w", err)
	}

	// Check if document can be deleted (only draft or rejected documents)
	if !document.IsEditable() {
		return fmt.Errorf("document cannot be deleted in %s status", document.Status)
	}

	// Delete document
	if err := s.documentRepo.Delete(ctx, id, organizationID); err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	// Log audit event
	if s.auditService != nil {
		details := fmt.Sprintf("Deleted %s document '%s'", document.DocumentType, document.Title)
		s.auditService.LogEvent(ctx, userID, organizationID, "document_deleted", "document", document.ID.String(), details, "", "")
	}

	return nil
}

// ListDocuments retrieves documents with filtering and pagination
func (s *DocumentService) ListDocuments(ctx context.Context, organizationID string, filter *models.DocumentFilter, limit, offset int) ([]*models.Document, int64, error) {
	documents, err := s.documentRepo.List(ctx, organizationID, filter, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list documents: %w", err)
	}

	total, err := s.documentRepo.Count(ctx, organizationID, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count documents: %w", err)
	}

	return documents, total, nil
}

// ListUserDocuments retrieves documents created by a specific user
func (s *DocumentService) ListUserDocuments(ctx context.Context, organizationID, userID string, limit, offset int) ([]*models.Document, int64, error) {
	documents, err := s.documentRepo.ListByUser(ctx, organizationID, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list user documents: %w", err)
	}

	total, err := s.documentRepo.CountByUser(ctx, organizationID, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count user documents: %w", err)
	}

	return documents, total, nil
}

// SearchDocuments performs full-text search on documents
func (s *DocumentService) SearchDocuments(ctx context.Context, organizationID, query string, filter *models.DocumentFilter, limit, offset int) ([]*models.DocumentSearchResult, int64, error) {
	results, err := s.documentRepo.Search(ctx, organizationID, query, filter, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search documents: %w", err)
	}

	total, err := s.documentRepo.Count(ctx, organizationID, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count search results: %w", err)
	}

	return results, total, nil
}

// SubmitDocument submits a document for approval
func (s *DocumentService) SubmitDocument(ctx context.Context, id uuid.UUID, organizationID, userID string) (*models.Document, error) {
	// Get document
	document, err := s.documentRepo.GetByID(ctx, id, organizationID)
	if err != nil {
		return nil, fmt.Errorf("document not found: %w", err)
	}

	// Check if document can be submitted
	if !document.CanBeSubmitted() {
		return nil, fmt.Errorf("document cannot be submitted in %s status", document.Status)
	}

	// Submit document
	if err := s.documentRepo.Submit(ctx, id, organizationID); err != nil {
		return nil, fmt.Errorf("failed to submit document: %w", err)
	}

	// Get updated document
	document, err = s.documentRepo.GetByID(ctx, id, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated document: %w", err)
	}

	// Log audit event
	if s.auditService != nil {
		details := fmt.Sprintf("Submitted %s document '%s' for approval", document.DocumentType, document.Title)
		s.auditService.LogEvent(ctx, userID, organizationID, "document_submitted", "document", document.ID.String(), details, "", "")
	}

	return document, nil
}

// GetDocumentStats retrieves document statistics
func (s *DocumentService) GetDocumentStats(ctx context.Context, organizationID string) (*models.DocumentStats, error) {
	return s.documentRepo.GetStats(ctx, organizationID)
}

// SyncFromSpecificModel syncs a specific model to the generic document table
func (s *DocumentService) SyncFromSpecificModel(ctx context.Context, modelType string, model interface{}) error {
	switch modelType {
	case "REQUISITION":
		if req, ok := model.(*models.Requisition); ok {
			return s.documentRepo.SyncFromRequisition(ctx, req)
		}
	case "BUDGET":
		if budget, ok := model.(*models.Budget); ok {
			return s.documentRepo.SyncFromBudget(ctx, budget)
		}
	case "PURCHASE_ORDER":
		if po, ok := model.(*models.PurchaseOrder); ok {
			return s.documentRepo.SyncFromPurchaseOrder(ctx, po)
		}
	case "PAYMENT_VOUCHER":
		if pv, ok := model.(*models.PaymentVoucher); ok {
			return s.documentRepo.SyncFromPaymentVoucher(ctx, pv)
		}
	case "GRN":
		if grn, ok := model.(*models.GoodsReceivedNote); ok {
			return s.documentRepo.SyncFromGRN(ctx, grn)
		}
	}
	
	return fmt.Errorf("unsupported model type: %s", modelType)
}