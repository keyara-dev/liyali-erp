package repository

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/liyali/liyali-gateway/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// DocumentRepositoryInterface defines the contract for document repository
type DocumentRepositoryInterface interface {
	// Basic CRUD operations
	Create(ctx context.Context, document *models.Document) (*models.Document, error)
	GetByID(ctx context.Context, id uuid.UUID, organizationID string) (*models.Document, error)
	GetByNumber(ctx context.Context, documentNumber, organizationID string) (*models.Document, error)
	Update(ctx context.Context, document *models.Document) (*models.Document, error)
	Delete(ctx context.Context, id uuid.UUID, organizationID string) error
	
	// List operations
	List(ctx context.Context, organizationID string, filter *models.DocumentFilter, limit, offset int) ([]*models.Document, error)
	ListByUser(ctx context.Context, organizationID, userID string, limit, offset int) ([]*models.Document, error)
	ListByType(ctx context.Context, organizationID, documentType string, limit, offset int) ([]*models.Document, error)
	ListByStatus(ctx context.Context, organizationID, status string, limit, offset int) ([]*models.Document, error)
	ListByDepartment(ctx context.Context, organizationID, department string, limit, offset int) ([]*models.Document, error)
	
	// Search operations
	Search(ctx context.Context, organizationID, query string, filter *models.DocumentFilter, limit, offset int) ([]*models.DocumentSearchResult, error)
	
	// Count operations
	Count(ctx context.Context, organizationID string, filter *models.DocumentFilter) (int64, error)
	CountByType(ctx context.Context, organizationID, documentType string) (int64, error)
	CountByStatus(ctx context.Context, organizationID, status string) (int64, error)
	CountByUser(ctx context.Context, organizationID, userID string) (int64, error)
	
	// Status operations
	UpdateStatus(ctx context.Context, id uuid.UUID, organizationID, status string) error
	Submit(ctx context.Context, id uuid.UUID, organizationID string) error
	
	// Statistics
	GetStats(ctx context.Context, organizationID string) (*models.DocumentStats, error)
	
	// Sync operations (to sync with existing specific models)
	SyncFromRequisition(ctx context.Context, requisition *models.Requisition) error
	SyncFromBudget(ctx context.Context, budget *models.Budget) error
	SyncFromPurchaseOrder(ctx context.Context, po *models.PurchaseOrder) error
	SyncFromPaymentVoucher(ctx context.Context, pv *models.PaymentVoucher) error
	SyncFromGRN(ctx context.Context, grn *models.GoodsReceivedNote) error
}

// DocumentRepository implements DocumentRepositoryInterface
type DocumentRepository struct {
	db    *gorm.DB
	pgxDB *pgxpool.Pool
}

// NewDocumentRepository creates a new document repository
func NewDocumentRepository(pgxDB *pgxpool.Pool, db *gorm.DB) DocumentRepositoryInterface {
	return &DocumentRepository{
		db:    db,
		pgxDB: pgxDB,
	}
}

// Create creates a new document
func (r *DocumentRepository) Create(ctx context.Context, document *models.Document) (*models.Document, error) {
	if err := r.db.WithContext(ctx).Create(document).Error; err != nil {
		return nil, err
	}
	
	// Load relationships
	if err := r.db.WithContext(ctx).Preload("Creator").Preload("Workflow").First(document, document.ID).Error; err != nil {
		return nil, err
	}
	
	return document, nil
}

// GetByID retrieves a document by ID
func (r *DocumentRepository) GetByID(ctx context.Context, id uuid.UUID, organizationID string) (*models.Document, error) {
	var document models.Document
	err := r.db.WithContext(ctx).
		Where("id = ? AND organization_id = ?", id, organizationID).
		Preload("Creator").
		Preload("Updater").
		Preload("Workflow").
		First(&document).Error
	
	if err != nil {
		return nil, err
	}
	
	return &document, nil
}

// GetByNumber retrieves a document by document number
func (r *DocumentRepository) GetByNumber(ctx context.Context, documentNumber, organizationID string) (*models.Document, error) {
	var document models.Document
	err := r.db.WithContext(ctx).
		Where("document_number = ? AND organization_id = ?", documentNumber, organizationID).
		Preload("Creator").
		Preload("Updater").
		Preload("Workflow").
		First(&document).Error
	
	if err != nil {
		return nil, err
	}
	
	return &document, nil
}

// Update updates a document
func (r *DocumentRepository) Update(ctx context.Context, document *models.Document) (*models.Document, error) {
	if err := r.db.WithContext(ctx).Save(document).Error; err != nil {
		return nil, err
	}
	
	// Reload with relationships
	if err := r.db.WithContext(ctx).
		Preload("Creator").
		Preload("Updater").
		Preload("Workflow").
		First(document, document.ID).Error; err != nil {
		return nil, err
	}
	
	return document, nil
}

// Delete deletes a document
func (r *DocumentRepository) Delete(ctx context.Context, id uuid.UUID, organizationID string) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND organization_id = ?", id, organizationID).
		Delete(&models.Document{}).Error
}

// List retrieves documents with filtering and pagination
func (r *DocumentRepository) List(ctx context.Context, organizationID string, filter *models.DocumentFilter, limit, offset int) ([]*models.Document, error) {
	query := r.db.WithContext(ctx).
		Where("organization_id = ?", organizationID).
		Preload("Creator").
		Preload("Workflow")
	
	// Apply filters
	query = r.applyFilters(query, filter)
	
	var documents []*models.Document
	err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&documents).Error
	
	return documents, err
}

// ListByUser retrieves documents created by a specific user
func (r *DocumentRepository) ListByUser(ctx context.Context, organizationID, userID string, limit, offset int) ([]*models.Document, error) {
	var documents []*models.Document
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND created_by = ?", organizationID, userID).
		Preload("Creator").
		Preload("Workflow").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&documents).Error
	
	return documents, err
}

// ListByType retrieves documents by type
func (r *DocumentRepository) ListByType(ctx context.Context, organizationID, documentType string, limit, offset int) ([]*models.Document, error) {
	var documents []*models.Document
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND document_type = ?", organizationID, documentType).
		Preload("Creator").
		Preload("Workflow").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&documents).Error
	
	return documents, err
}

// ListByStatus retrieves documents by status
func (r *DocumentRepository) ListByStatus(ctx context.Context, organizationID, status string, limit, offset int) ([]*models.Document, error) {
	var documents []*models.Document
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND status = ?", organizationID, status).
		Preload("Creator").
		Preload("Workflow").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&documents).Error
	
	return documents, err
}

// ListByDepartment retrieves documents by department
func (r *DocumentRepository) ListByDepartment(ctx context.Context, organizationID, department string, limit, offset int) ([]*models.Document, error) {
	var documents []*models.Document
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND department = ?", organizationID, department).
		Preload("Creator").
		Preload("Workflow").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&documents).Error
	
	return documents, err
}

// Search performs full-text search on documents
func (r *DocumentRepository) Search(ctx context.Context, organizationID, query string, filter *models.DocumentFilter, limit, offset int) ([]*models.DocumentSearchResult, error) {
	// Build search query with PostgreSQL full-text search
	searchQuery := r.db.WithContext(ctx).
		Where("organization_id = ?", organizationID).
		Preload("Creator").
		Preload("Workflow")
	
	// Apply text search
	if query != "" {
		searchTerms := strings.Fields(query)
		for _, term := range searchTerms {
			searchQuery = searchQuery.Where(
				"title ILIKE ? OR description ILIKE ? OR document_number ILIKE ? OR department ILIKE ?",
				"%"+term+"%", "%"+term+"%", "%"+term+"%", "%"+term+"%",
			)
		}
	}
	
	// Apply filters
	searchQuery = r.applyFilters(searchQuery, filter)
	
	var documents []*models.Document
	err := searchQuery.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&documents).Error
	
	if err != nil {
		return nil, err
	}
	
	// Convert to search results with relevance scoring
	results := make([]*models.DocumentSearchResult, len(documents))
	for i, doc := range documents {
		results[i] = &models.DocumentSearchResult{
			Document:  *doc,
			Relevance: r.calculateRelevance(doc, query),
			Matches:   r.findMatches(doc, query),
		}
	}
	
	return results, nil
}

// Count counts documents with optional filtering
func (r *DocumentRepository) Count(ctx context.Context, organizationID string, filter *models.DocumentFilter) (int64, error) {
	query := r.db.WithContext(ctx).
		Model(&models.Document{}).
		Where("organization_id = ?", organizationID)
	
	query = r.applyFilters(query, filter)
	
	var count int64
	err := query.Count(&count).Error
	return count, err
}

// CountByType counts documents by type
func (r *DocumentRepository) CountByType(ctx context.Context, organizationID, documentType string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Document{}).
		Where("organization_id = ? AND document_type = ?", organizationID, documentType).
		Count(&count).Error
	return count, err
}

// CountByStatus counts documents by status
func (r *DocumentRepository) CountByStatus(ctx context.Context, organizationID, status string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Document{}).
		Where("organization_id = ? AND status = ?", organizationID, status).
		Count(&count).Error
	return count, err
}

// CountByUser counts documents by user
func (r *DocumentRepository) CountByUser(ctx context.Context, organizationID, userID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Document{}).
		Where("organization_id = ? AND created_by = ?", organizationID, userID).
		Count(&count).Error
	return count, err
}

// UpdateStatus updates document status
func (r *DocumentRepository) UpdateStatus(ctx context.Context, id uuid.UUID, organizationID, status string) error {
	return r.db.WithContext(ctx).
		Model(&models.Document{}).
		Where("id = ? AND organization_id = ?", id, organizationID).
		Update("status", status).Error
}

// Submit submits a document for approval
func (r *DocumentRepository) Submit(ctx context.Context, id uuid.UUID, organizationID string) error {
	return r.db.WithContext(ctx).
		Model(&models.Document{}).
		Where("id = ? AND organization_id = ? AND status IN ('draft', 'rejected')", id, organizationID).
		Update("status", "submitted").Error
}

// GetStats retrieves document statistics
func (r *DocumentRepository) GetStats(ctx context.Context, organizationID string) (*models.DocumentStats, error) {
	stats := &models.DocumentStats{
		DocumentsByType:   make(map[string]int64),
		DocumentsByStatus: make(map[string]int64),
		DocumentsByDept:   make(map[string]int64),
	}
	
	// Total documents
	r.db.WithContext(ctx).
		Model(&models.Document{}).
		Where("organization_id = ?", organizationID).
		Count(&stats.TotalDocuments)
	
	// Documents by type
	var typeStats []struct {
		DocumentType string
		Count        int64
	}
	r.db.WithContext(ctx).
		Model(&models.Document{}).
		Select("document_type, COUNT(*) as count").
		Where("organization_id = ?", organizationID).
		Group("document_type").
		Scan(&typeStats)
	
	for _, stat := range typeStats {
		stats.DocumentsByType[stat.DocumentType] = stat.Count
	}
	
	// Documents by status
	var statusStats []struct {
		Status string
		Count  int64
	}
	r.db.WithContext(ctx).
		Model(&models.Document{}).
		Select("status, COUNT(*) as count").
		Where("organization_id = ?", organizationID).
		Group("status").
		Scan(&statusStats)
	
	for _, stat := range statusStats {
		stats.DocumentsByStatus[stat.Status] = stat.Count
	}
	
	// Recent documents (last 7 days)
	weekAgo := time.Now().AddDate(0, 0, -7)
	r.db.WithContext(ctx).
		Model(&models.Document{}).
		Where("organization_id = ? AND created_at >= ?", organizationID, weekAgo).
		Count(&stats.RecentDocuments)
	
	// Pending approvals
	stats.PendingApprovals = stats.DocumentsByStatus["submitted"]
	
	// Total and average value
	var valueStats struct {
		TotalValue   float64
		AverageValue float64
	}
	r.db.WithContext(ctx).
		Model(&models.Document{}).
		Select("COALESCE(SUM(amount), 0) as total_value, COALESCE(AVG(amount), 0) as average_value").
		Where("organization_id = ? AND amount IS NOT NULL", organizationID).
		Scan(&valueStats)
	
	stats.TotalValue = valueStats.TotalValue
	stats.AverageValue = valueStats.AverageValue
	
	return stats, nil
}

// Helper methods

// applyFilters applies filters to a query
func (r *DocumentRepository) applyFilters(query *gorm.DB, filter *models.DocumentFilter) *gorm.DB {
	if filter == nil {
		return query
	}
	
	if len(filter.DocumentTypes) > 0 {
		query = query.Where("document_type IN ?", filter.DocumentTypes)
	}
	
	if len(filter.Statuses) > 0 {
		query = query.Where("status IN ?", filter.Statuses)
	}
	
	if len(filter.Departments) > 0 {
		query = query.Where("department IN ?", filter.Departments)
	}
	
	if len(filter.CreatedBy) > 0 {
		query = query.Where("created_by IN ?", filter.CreatedBy)
	}
	
	if filter.DateFrom != nil {
		query = query.Where("created_at >= ?", filter.DateFrom)
	}
	
	if filter.DateTo != nil {
		query = query.Where("created_at <= ?", filter.DateTo)
	}
	
	if filter.AmountMin != nil {
		query = query.Where("amount >= ?", filter.AmountMin)
	}
	
	if filter.AmountMax != nil {
		query = query.Where("amount <= ?", filter.AmountMax)
	}
	
	return query
}

// calculateRelevance calculates search relevance score
func (r *DocumentRepository) calculateRelevance(doc *models.Document, query string) float64 {
	if query == "" {
		return 1.0
	}
	
	score := 0.0
	queryLower := strings.ToLower(query)
	
	// Title match (highest weight)
	if strings.Contains(strings.ToLower(doc.Title), queryLower) {
		score += 3.0
	}
	
	// Document number match
	if strings.Contains(strings.ToLower(doc.DocumentNumber), queryLower) {
		score += 2.0
	}
	
	// Description match
	if doc.Description != nil && strings.Contains(strings.ToLower(*doc.Description), queryLower) {
		score += 1.0
	}
	
	// Department match
	if doc.Department != nil && strings.Contains(strings.ToLower(*doc.Department), queryLower) {
		score += 0.5
	}
	
	return score
}

// findMatches finds which fields matched the search query
func (r *DocumentRepository) findMatches(doc *models.Document, query string) []string {
	if query == "" {
		return nil
	}
	
	var matches []string
	queryLower := strings.ToLower(query)
	
	if strings.Contains(strings.ToLower(doc.Title), queryLower) {
		matches = append(matches, "title")
	}
	
	if strings.Contains(strings.ToLower(doc.DocumentNumber), queryLower) {
		matches = append(matches, "documentNumber")
	}
	
	if doc.Description != nil && strings.Contains(strings.ToLower(*doc.Description), queryLower) {
		matches = append(matches, "description")
	}
	
	if doc.Department != nil && strings.Contains(strings.ToLower(*doc.Department), queryLower) {
		matches = append(matches, "department")
	}
	
	return matches
}

// Sync operations to keep generic documents in sync with specific models

// SyncFromRequisition syncs a requisition to the generic document table
func (r *DocumentRepository) SyncFromRequisition(ctx context.Context, requisition *models.Requisition) error {
	// Check if document already exists
	var existingDoc models.Document
	err := r.db.WithContext(ctx).
		Where("document_type = ? AND data->>'id' = ?", "REQUISITION", requisition.ID).
		First(&existingDoc).Error
	
	// Prepare document data
	dataMap := map[string]interface{}{
		"id":          requisition.ID,
		"reqNumber":   requisition.REQNumber,
		"items":       requisition.Items,
		"priority":    requisition.Priority,
		"categoryId":  requisition.CategoryID,
		"preferredVendorId": requisition.PreferredVendorID,
		"isEstimate":  requisition.IsEstimate,
		"approvalStage": requisition.ApprovalStage,
		"approvalHistory": requisition.ApprovalHistory,
	}
	
	data, _ := json.Marshal(dataMap)
	
	doc := &models.Document{
		OrganizationID: requisition.OrganizationID,
		DocumentType:   "REQUISITION",
		Title:          requisition.Title,
		Status:         requisition.Status,
		Amount:         &requisition.TotalAmount,
		Currency:       &requisition.Currency,
		Data:           datatypes.JSON(data),
		CreatedAt:      requisition.CreatedAt,
		UpdatedAt:      requisition.UpdatedAt,
	}
	
	// Set optional fields
	if requisition.Description != "" {
		doc.Description = &requisition.Description
	}
	if requisition.Department != "" {
		doc.Department = &requisition.Department
	}
	if requisition.RequesterId != "" {
		doc.CreatedBy = requisition.RequesterId
	}
	
	if err == gorm.ErrRecordNotFound {
		// Create new document
		doc.ID = uuid.New()
		return r.db.WithContext(ctx).Create(doc).Error
	} else if err != nil {
		return err
	} else {
		// Update existing document
		doc.ID = existingDoc.ID
		return r.db.WithContext(ctx).Save(doc).Error
	}
}

// SyncFromBudget syncs a budget to the generic document table
func (r *DocumentRepository) SyncFromBudget(ctx context.Context, budget *models.Budget) error {
	var existingDoc models.Document
	err := r.db.WithContext(ctx).
		Where("document_type = ? AND data->>'id' = ?", "BUDGET", budget.ID).
		First(&existingDoc).Error
	
	dataMap := map[string]interface{}{
		"id":           budget.ID,
		"budgetCode":   budget.BudgetCode,
		"fiscalYear":   budget.FiscalYear,
		"totalBudget":  budget.TotalBudget,
		"allocatedAmount": budget.AllocatedAmount,
		"remainingAmount": budget.RemainingAmount,
		"approvalStage": budget.ApprovalStage,
		"approvalHistory": budget.ApprovalHistory,
	}
	
	data, _ := json.Marshal(dataMap)
	
	doc := &models.Document{
		OrganizationID: budget.OrganizationID,
		DocumentType:   "BUDGET",
		Title:          budget.BudgetCode + " - " + budget.FiscalYear,
		Status:         budget.Status,
		Amount:         &budget.TotalBudget,
		Data:           datatypes.JSON(data),
		CreatedAt:      budget.CreatedAt,
		UpdatedAt:      budget.UpdatedAt,
	}
	
	if budget.Department != "" {
		doc.Department = &budget.Department
	}
	if budget.OwnerID != "" {
		doc.CreatedBy = budget.OwnerID
	}
	
	if err == gorm.ErrRecordNotFound {
		doc.ID = uuid.New()
		return r.db.WithContext(ctx).Create(doc).Error
	} else if err != nil {
		return err
	} else {
		doc.ID = existingDoc.ID
		return r.db.WithContext(ctx).Save(doc).Error
	}
}

// SyncFromPurchaseOrder syncs a purchase order to the generic document table
func (r *DocumentRepository) SyncFromPurchaseOrder(ctx context.Context, po *models.PurchaseOrder) error {
	var existingDoc models.Document
	err := r.db.WithContext(ctx).
		Where("document_type = ? AND data->>'id' = ?", "PURCHASE_ORDER", po.ID).
		First(&existingDoc).Error
	
	dataMap := map[string]interface{}{
		"id":           po.ID,
		"poNumber":     po.PONumber,
		"vendorId":     po.VendorID,
		"items":        po.Items,
		"deliveryDate": po.DeliveryDate,
		"linkedRequisition": po.LinkedRequisition,
		"approvalStage": po.ApprovalStage,
		"approvalHistory": po.ApprovalHistory,
	}
	
	data, _ := json.Marshal(dataMap)
	
	doc := &models.Document{
		OrganizationID: po.OrganizationID,
		DocumentType:   "PURCHASE_ORDER",
		Title:          "PO: " + po.PONumber,
		Status:         po.Status,
		Amount:         &po.TotalAmount,
		Currency:       &po.Currency,
		Data:           datatypes.JSON(data),
		CreatedAt:      po.CreatedAt,
		UpdatedAt:      po.UpdatedAt,
	}
	
	// PO doesn't have a direct creator field, so we'll leave it empty for now
	doc.CreatedBy = "system" // Default value
	
	if err == gorm.ErrRecordNotFound {
		doc.ID = uuid.New()
		return r.db.WithContext(ctx).Create(doc).Error
	} else if err != nil {
		return err
	} else {
		doc.ID = existingDoc.ID
		return r.db.WithContext(ctx).Save(doc).Error
	}
}

// SyncFromPaymentVoucher syncs a payment voucher to the generic document table
func (r *DocumentRepository) SyncFromPaymentVoucher(ctx context.Context, pv *models.PaymentVoucher) error {
	var existingDoc models.Document
	err := r.db.WithContext(ctx).
		Where("document_type = ? AND data->>'id' = ?", "PAYMENT_VOUCHER", pv.ID).
		First(&existingDoc).Error
	
	dataMap := map[string]interface{}{
		"id":             pv.ID,
		"voucherNumber":  pv.VoucherNumber,
		"vendorId":       pv.VendorID,
		"invoiceNumber":  pv.InvoiceNumber,
		"paymentMethod":  pv.PaymentMethod,
		"glCode":         pv.GLCode,
		"linkedPO":       pv.LinkedPO,
		"approvalStage":  pv.ApprovalStage,
		"approvalHistory": pv.ApprovalHistory,
	}
	
	data, _ := json.Marshal(dataMap)
	
	doc := &models.Document{
		OrganizationID: pv.OrganizationID,
		DocumentType:   "PAYMENT_VOUCHER",
		Title:          "Payment Voucher: " + pv.VoucherNumber,
		Status:         pv.Status,
		Amount:         &pv.Amount,
		Currency:       &pv.Currency,
		Data:           datatypes.JSON(data),
		CreatedAt:      pv.CreatedAt,
		UpdatedAt:      pv.UpdatedAt,
	}
	
	if pv.Description != "" {
		doc.Description = &pv.Description
	}
	
	// PV doesn't have a direct creator field, so we'll leave it empty for now
	doc.CreatedBy = "system" // Default value
	
	if err == gorm.ErrRecordNotFound {
		doc.ID = uuid.New()
		return r.db.WithContext(ctx).Create(doc).Error
	} else if err != nil {
		return err
	} else {
		doc.ID = existingDoc.ID
		return r.db.WithContext(ctx).Save(doc).Error
	}
}

// SyncFromGRN syncs a GRN to the generic document table
func (r *DocumentRepository) SyncFromGRN(ctx context.Context, grn *models.GoodsReceivedNote) error {
	var existingDoc models.Document
	err := r.db.WithContext(ctx).
		Where("document_type = ? AND data->>'id' = ?", "GRN", grn.ID).
		First(&existingDoc).Error
	
	dataMap := map[string]interface{}{
		"id":             grn.ID,
		"grnNumber":      grn.GRNNumber,
		"poNumber":       grn.PONumber,
		"items":          grn.Items,
		"receivedDate":   grn.ReceivedDate,
		"receivedBy":     grn.ReceivedBy,
		"qualityIssues":  grn.QualityIssues,
		"approvalStage":  grn.ApprovalStage,
		"approvalHistory": grn.ApprovalHistory,
	}
	
	data, _ := json.Marshal(dataMap)
	
	// Calculate total amount from items (if available)
	totalAmount := 0.0
	// Note: We'd need to calculate this from the items if needed
	
	doc := &models.Document{
		OrganizationID: grn.OrganizationID,
		DocumentType:   "GRN",
		Title:          "GRN: " + grn.GRNNumber,
		Status:         grn.Status,
		Amount:         &totalAmount,
		Currency:       func() *string { s := "USD"; return &s }(), // Default currency
		Data:           datatypes.JSON(data),
		CreatedAt:      grn.CreatedAt,
		UpdatedAt:      grn.UpdatedAt,
	}
	
	if grn.ReceivedBy != "" {
		doc.CreatedBy = grn.ReceivedBy
	} else {
		doc.CreatedBy = "system" // Default value
	}
	
	if err == gorm.ErrRecordNotFound {
		doc.ID = uuid.New()
		return r.db.WithContext(ctx).Create(doc).Error
	} else if err != nil {
		return err
	} else {
		doc.ID = existingDoc.ID
		return r.db.WithContext(ctx).Save(doc).Error
	}
}