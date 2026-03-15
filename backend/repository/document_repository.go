package repository

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/liyali/liyali-gateway/models"
	"gorm.io/gorm"
)

// DocumentRepositoryInterface defines the contract for document repository
type DocumentRepositoryInterface interface {
	// Basic CRUD operations
	Create(ctx context.Context, document *models.Document) (*models.Document, error)
	GetByID(ctx context.Context, id uuid.UUID, organizationID string) (*models.Document, error)
	GetByNumber(ctx context.Context, documentNumber, organizationID string) (*models.Document, error)
	GetByNumberOnly(ctx context.Context, documentNumber string) (*models.Document, error) // Public verification (no org filter)
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
	CountSearch(ctx context.Context, organizationID, query string, filter *models.DocumentFilter) (int64, error)
	CountByType(ctx context.Context, organizationID, documentType string) (int64, error)
	CountByStatus(ctx context.Context, organizationID, status string) (int64, error)
	CountByUser(ctx context.Context, organizationID, userID string) (int64, error)

	// Status operations
	UpdateStatus(ctx context.Context, id uuid.UUID, organizationID, status string) error
	Submit(ctx context.Context, id uuid.UUID, organizationID string) error

	// Statistics
	GetStats(ctx context.Context, organizationID string) (*models.DocumentStats, error)

	// Public document retrieval for PDF generation
	GetRequisitionByNumberPublic(ctx context.Context, documentNumber string) (*models.Requisition, error)
	GetPurchaseOrderByNumberPublic(ctx context.Context, documentNumber string) (*models.PurchaseOrder, error)
	GetPaymentVoucherByNumberPublic(ctx context.Context, documentNumber string) (*models.PaymentVoucher, error)
	GetGRNByNumberPublic(ctx context.Context, documentNumber string) (*models.GoodsReceivedNote, error)

	// GetOrganizationBranding fetches minimal branding fields for an org (no auth required)
	GetOrganizationBranding(ctx context.Context, organizationID string) (*models.Organization, error)
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

// GetByNumberOnly retrieves a document by document number only (for public verification)
// This is used for public document verification without requiring authentication
func (r *DocumentRepository) GetByNumberOnly(ctx context.Context, documentNumber string) (*models.Document, error) {
	var document models.Document
	err := r.db.WithContext(ctx).
		Where("document_number = ?", documentNumber).
		Preload("Creator").
		Preload("Organization").
		First(&document).Error

	if err != nil {
		return nil, err
	}

	return &document, nil
}

// GetRequisitionByNumberPublic retrieves a requisition by document number for public PDF generation
func (r *DocumentRepository) GetRequisitionByNumberPublic(ctx context.Context, documentNumber string) (*models.Requisition, error) {
	var requisition models.Requisition
	err := r.db.WithContext(ctx).
		Where("document_number = ?", documentNumber).
		Preload("Requester").
		Preload("Organization").
		Preload("Category").
		Preload("PreferredVendor").
		First(&requisition).Error

	if err != nil {
		return nil, err
	}

	return &requisition, nil
}

// GetPurchaseOrderByNumberPublic retrieves a purchase order by document number for public PDF generation
func (r *DocumentRepository) GetPurchaseOrderByNumberPublic(ctx context.Context, documentNumber string) (*models.PurchaseOrder, error) {
	var po models.PurchaseOrder
	err := r.db.WithContext(ctx).
		Where("document_number = ?", documentNumber).
		Preload("Vendor").
		Preload("Organization").
		First(&po).Error

	if err != nil {
		return nil, err
	}

	return &po, nil
}

// GetPaymentVoucherByNumberPublic retrieves a payment voucher by document number for public PDF generation
func (r *DocumentRepository) GetPaymentVoucherByNumberPublic(ctx context.Context, documentNumber string) (*models.PaymentVoucher, error) {
	var pv models.PaymentVoucher
	err := r.db.WithContext(ctx).
		Where("document_number = ?", documentNumber).
		Preload("Vendor").
		Preload("Organization").
		First(&pv).Error

	if err != nil {
		return nil, err
	}

	return &pv, nil
}

// GetGRNByNumberPublic retrieves a GRN by document number for public PDF generation
func (r *DocumentRepository) GetGRNByNumberPublic(ctx context.Context, documentNumber string) (*models.GoodsReceivedNote, error) {
	var grn models.GoodsReceivedNote
	err := r.db.WithContext(ctx).
		Where("document_number = ?", documentNumber).
		Preload("Organization").
		First(&grn).Error

	if err != nil {
		return nil, err
	}

	return &grn, nil
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

// Search performs full-text search on documents across all entity tables
func (r *DocumentRepository) Search(ctx context.Context, organizationID, query string, filter *models.DocumentFilter, limit, offset int) ([]*models.DocumentSearchResult, error) {
	var allResults []*models.DocumentSearchResult

	// Check if we need to filter by document type
	searchTypes := filter.DocumentTypes
	if len(searchTypes) == 0 {
		// Search all types if no filter specified
		searchTypes = []string{"REQUISITION", "PURCHASE_ORDER", "PAYMENT_VOUCHER", "GRN"}
	}

	// Normalize search types to uppercase
	normalizedTypes := toUpperSlice(searchTypes)

	// Search requisitions
	if containsType(normalizedTypes, "REQUISITION") {
		results, err := r.searchRequisitions(ctx, organizationID, query, filter, limit, offset)
		if err == nil {
			allResults = append(allResults, results...)
		}
	}

	// Search purchase orders
	if containsType(normalizedTypes, "PURCHASE_ORDER") || containsType(normalizedTypes, "PO") {
		results, err := r.searchPurchaseOrders(ctx, organizationID, query, filter, limit, offset)
		if err == nil {
			allResults = append(allResults, results...)
		}
	}

	// Search payment vouchers
	if containsType(normalizedTypes, "PAYMENT_VOUCHER") || containsType(normalizedTypes, "PV") {
		results, err := r.searchPaymentVouchers(ctx, organizationID, query, filter, limit, offset)
		if err == nil {
			allResults = append(allResults, results...)
		}
	}

	// Search GRNs
	if containsType(normalizedTypes, "GRN") || containsType(normalizedTypes, "GOODS_RECEIVED_NOTE") {
		results, err := r.searchGRNs(ctx, organizationID, query, filter, limit, offset)
		if err == nil {
			allResults = append(allResults, results...)
		}
	}

	// Sort by created_at DESC
	sortResultsByDate(allResults)

	// Apply pagination to combined results
	start := offset
	if start > len(allResults) {
		start = len(allResults)
	}
	end := start + limit
	if end > len(allResults) {
		end = len(allResults)
	}

	return allResults[start:end], nil
}

// containsType checks if a slice contains a type (case-insensitive)
func containsType(types []string, target string) bool {
	targetUpper := strings.ToUpper(target)
	for _, t := range types {
		if strings.ToUpper(t) == targetUpper {
			return true
		}
	}
	return false
}

// sortResultsByDate sorts results by created_at in descending order
func sortResultsByDate(results []*models.DocumentSearchResult) {
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].CreatedAt.After(results[i].CreatedAt) {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
}

// searchRequisitions searches the requisitions table
func (r *DocumentRepository) searchRequisitions(ctx context.Context, organizationID, query string, filter *models.DocumentFilter, limit, offset int) ([]*models.DocumentSearchResult, error) {
	searchQuery := r.db.WithContext(ctx).
		Model(&models.Requisition{}).
		Where("organization_id = ?", organizationID).
		Preload("Requester")

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

	// Apply document number filter
	if filter != nil && filter.DocumentNumber != "" {
		searchQuery = searchQuery.Where("document_number ILIKE ?", "%"+filter.DocumentNumber+"%")
	}

	// Apply status filter
	if filter != nil && len(filter.Statuses) > 0 {
		searchQuery = searchQuery.Where("UPPER(status) IN ?", toUpperSlice(filter.Statuses))
	}

	// Apply date filters
	if filter != nil && filter.DateFrom != nil {
		searchQuery = searchQuery.Where("created_at >= ?", filter.DateFrom)
	}
	if filter != nil && filter.DateTo != nil {
		searchQuery = searchQuery.Where("created_at <= ?", filter.DateTo)
	}

	var requisitions []*models.Requisition
	err := searchQuery.Order("created_at DESC").Find(&requisitions).Error
	if err != nil {
		return nil, err
	}

	// Convert to DocumentSearchResult
	results := make([]*models.DocumentSearchResult, len(requisitions))
	for i, req := range requisitions {
		doc := r.requisitionToDocument(req)
		results[i] = &models.DocumentSearchResult{
			Document:  *doc,
			Relevance: 1.0,
		}
	}

	return results, nil
}

// searchPurchaseOrders searches the purchase_orders table
func (r *DocumentRepository) searchPurchaseOrders(ctx context.Context, organizationID, query string, filter *models.DocumentFilter, limit, offset int) ([]*models.DocumentSearchResult, error) {
	searchQuery := r.db.WithContext(ctx).
		Model(&models.PurchaseOrder{}).
		Where("organization_id = ?", organizationID).
		Preload("Vendor")

	// Apply text search
	if query != "" {
		searchTerms := strings.Fields(query)
		for _, term := range searchTerms {
			searchQuery = searchQuery.Where(
				"title ILIKE ? OR document_number ILIKE ?",
				"%"+term+"%", "%"+term+"%",
			)
		}
	}

	// Apply document number filter
	if filter != nil && filter.DocumentNumber != "" {
		searchQuery = searchQuery.Where("document_number ILIKE ?", "%"+filter.DocumentNumber+"%")
	}

	// Apply status filter
	if filter != nil && len(filter.Statuses) > 0 {
		searchQuery = searchQuery.Where("UPPER(status) IN ?", toUpperSlice(filter.Statuses))
	}

	// Apply date filters
	if filter != nil && filter.DateFrom != nil {
		searchQuery = searchQuery.Where("created_at >= ?", filter.DateFrom)
	}
	if filter != nil && filter.DateTo != nil {
		searchQuery = searchQuery.Where("created_at <= ?", filter.DateTo)
	}

	var pos []*models.PurchaseOrder
	err := searchQuery.Order("created_at DESC").Find(&pos).Error
	if err != nil {
		return nil, err
	}

	// Convert to DocumentSearchResult
	results := make([]*models.DocumentSearchResult, len(pos))
	for i, po := range pos {
		doc := r.purchaseOrderToDocument(po)
		results[i] = &models.DocumentSearchResult{
			Document:  *doc,
			Relevance: 1.0,
		}
	}

	return results, nil
}

// searchPaymentVouchers searches the payment_vouchers table
func (r *DocumentRepository) searchPaymentVouchers(ctx context.Context, organizationID, query string, filter *models.DocumentFilter, limit, offset int) ([]*models.DocumentSearchResult, error) {
	searchQuery := r.db.WithContext(ctx).
		Model(&models.PaymentVoucher{}).
		Where("organization_id = ?", organizationID).
		Preload("Vendor")

	// Apply text search
	if query != "" {
		searchTerms := strings.Fields(query)
		for _, term := range searchTerms {
			searchQuery = searchQuery.Where(
				"title ILIKE ? OR description ILIKE ? OR document_number ILIKE ?",
				"%"+term+"%", "%"+term+"%", "%"+term+"%",
			)
		}
	}

	// Apply document number filter
	if filter != nil && filter.DocumentNumber != "" {
		searchQuery = searchQuery.Where("document_number ILIKE ?", "%"+filter.DocumentNumber+"%")
	}

	// Apply status filter
	if filter != nil && len(filter.Statuses) > 0 {
		searchQuery = searchQuery.Where("UPPER(status) IN ?", toUpperSlice(filter.Statuses))
	}

	// Apply date filters
	if filter != nil && filter.DateFrom != nil {
		searchQuery = searchQuery.Where("created_at >= ?", filter.DateFrom)
	}
	if filter != nil && filter.DateTo != nil {
		searchQuery = searchQuery.Where("created_at <= ?", filter.DateTo)
	}

	var pvs []*models.PaymentVoucher
	err := searchQuery.Order("created_at DESC").Find(&pvs).Error
	if err != nil {
		return nil, err
	}

	// Convert to DocumentSearchResult
	results := make([]*models.DocumentSearchResult, len(pvs))
	for i, pv := range pvs {
		doc := r.paymentVoucherToDocument(pv)
		results[i] = &models.DocumentSearchResult{
			Document:  *doc,
			Relevance: 1.0,
		}
	}

	return results, nil
}

// searchGRNs searches the goods_received_notes table
func (r *DocumentRepository) searchGRNs(ctx context.Context, organizationID, query string, filter *models.DocumentFilter, limit, offset int) ([]*models.DocumentSearchResult, error) {
	searchQuery := r.db.WithContext(ctx).
		Model(&models.GoodsReceivedNote{}).
		Where("organization_id = ?", organizationID)

	// Apply text search
	if query != "" {
		searchTerms := strings.Fields(query)
		for _, term := range searchTerms {
			searchQuery = searchQuery.Where(
				"document_number ILIKE ? OR notes ILIKE ?",
				"%"+term+"%", "%"+term+"%",
			)
		}
	}

	// Apply document number filter
	if filter != nil && filter.DocumentNumber != "" {
		searchQuery = searchQuery.Where("document_number ILIKE ?", "%"+filter.DocumentNumber+"%")
	}

	// Apply status filter
	if filter != nil && len(filter.Statuses) > 0 {
		searchQuery = searchQuery.Where("UPPER(status) IN ?", toUpperSlice(filter.Statuses))
	}

	// Apply date filters
	if filter != nil && filter.DateFrom != nil {
		searchQuery = searchQuery.Where("created_at >= ?", filter.DateFrom)
	}
	if filter != nil && filter.DateTo != nil {
		searchQuery = searchQuery.Where("created_at <= ?", filter.DateTo)
	}

	var grns []*models.GoodsReceivedNote
	err := searchQuery.Order("created_at DESC").Find(&grns).Error
	if err != nil {
		return nil, err
	}

	// Convert to DocumentSearchResult
	results := make([]*models.DocumentSearchResult, len(grns))
	for i, grn := range grns {
		doc := r.grnToDocument(grn)
		results[i] = &models.DocumentSearchResult{
			Document:  *doc,
			Relevance: 1.0,
		}
	}

	return results, nil
}

// requisitionToDocument converts a Requisition to a Document
func (r *DocumentRepository) requisitionToDocument(req *models.Requisition) *models.Document {
	id, _ := uuid.Parse(req.ID)
	doc := &models.Document{
		ID:             id,
		OrganizationID: req.OrganizationID,
		DocumentType:   "REQUISITION",
		DocumentNumber: req.DocumentNumber,
		Title:          req.Title,
		Status:         req.Status,
		Amount:         &req.TotalAmount,
		Currency:       &req.Currency,
		CreatedBy:      req.RequesterId,
		CreatedAt:      req.CreatedAt,
		UpdatedAt:      req.UpdatedAt,
	}
	if req.Description != "" {
		doc.Description = &req.Description
	}
	if req.Department != "" {
		doc.Department = &req.Department
	}
	if req.Requester != nil {
		doc.Creator = req.Requester
	}
	return doc
}

// purchaseOrderToDocument converts a PurchaseOrder to a Document
func (r *DocumentRepository) purchaseOrderToDocument(po *models.PurchaseOrder) *models.Document {
	id, _ := uuid.Parse(po.ID)
	doc := &models.Document{
		ID:             id,
		OrganizationID: po.OrganizationID,
		DocumentType:   "PURCHASE_ORDER",
		DocumentNumber: po.DocumentNumber,
		Title:          po.Title,
		Status:         po.Status,
		Amount:         &po.TotalAmount,
		Currency:       &po.Currency,
		CreatedBy:      "system",
		CreatedAt:      po.CreatedAt,
		UpdatedAt:      po.UpdatedAt,
	}
	return doc
}

// paymentVoucherToDocument converts a PaymentVoucher to a Document
func (r *DocumentRepository) paymentVoucherToDocument(pv *models.PaymentVoucher) *models.Document {
	id, _ := uuid.Parse(pv.ID)
	doc := &models.Document{
		ID:             id,
		OrganizationID: pv.OrganizationID,
		DocumentType:   "PAYMENT_VOUCHER",
		DocumentNumber: pv.DocumentNumber,
		Title:          pv.Title,
		Status:         pv.Status,
		Amount:         &pv.Amount,
		Currency:       &pv.Currency,
		CreatedBy:      "system",
		CreatedAt:      pv.CreatedAt,
		UpdatedAt:      pv.UpdatedAt,
	}
	if pv.Description != "" {
		doc.Description = &pv.Description
	}
	return doc
}

// grnToDocument converts a GoodsReceivedNote to a Document
func (r *DocumentRepository) grnToDocument(grn *models.GoodsReceivedNote) *models.Document {
	id, _ := uuid.Parse(grn.ID)
	title := "GRN: " + grn.DocumentNumber
	if grn.Notes != "" {
		title = grn.Notes
	}
	doc := &models.Document{
		ID:             id,
		OrganizationID: grn.OrganizationID,
		DocumentType:   "GRN",
		DocumentNumber: grn.DocumentNumber,
		Title:          title,
		Status:         grn.Status,
		CreatedBy:      grn.ReceivedBy,
		CreatedAt:      grn.CreatedAt,
		UpdatedAt:      grn.UpdatedAt,
	}
	return doc
}

// Count counts documents with optional filtering (no text search).
func (r *DocumentRepository) Count(ctx context.Context, organizationID string, filter *models.DocumentFilter) (int64, error) {
	return r.CountSearch(ctx, organizationID, "", filter)
}

// CountSearch counts documents matching both the text query and optional filter.
func (r *DocumentRepository) CountSearch(ctx context.Context, organizationID, query string, filter *models.DocumentFilter) (int64, error) {
	var totalCount int64

	searchTypes := []string{}
	if filter != nil && len(filter.DocumentTypes) > 0 {
		searchTypes = toUpperSlice(filter.DocumentTypes)
	} else {
		searchTypes = []string{"REQUISITION", "PURCHASE_ORDER", "PAYMENT_VOUCHER", "GRN"}
	}

	if containsType(searchTypes, "REQUISITION") {
		count, _ := r.countRequisitions(ctx, organizationID, query, filter)
		totalCount += count
	}
	if containsType(searchTypes, "PURCHASE_ORDER") || containsType(searchTypes, "PO") {
		count, _ := r.countPurchaseOrders(ctx, organizationID, query, filter)
		totalCount += count
	}
	if containsType(searchTypes, "PAYMENT_VOUCHER") || containsType(searchTypes, "PV") {
		count, _ := r.countPaymentVouchers(ctx, organizationID, query, filter)
		totalCount += count
	}
	if containsType(searchTypes, "GRN") || containsType(searchTypes, "GOODS_RECEIVED_NOTE") {
		count, _ := r.countGRNs(ctx, organizationID, query, filter)
		totalCount += count
	}

	return totalCount, nil
}

func applyTextSearch(q *gorm.DB, searchText string) *gorm.DB {
	if searchText == "" {
		return q
	}
	for _, term := range strings.Fields(searchText) {
		q = q.Where(
			"title ILIKE ? OR description ILIKE ? OR document_number ILIKE ? OR department ILIKE ?",
			"%"+term+"%", "%"+term+"%", "%"+term+"%", "%"+term+"%",
		)
	}
	return q
}

func (r *DocumentRepository) countRequisitions(ctx context.Context, organizationID, searchText string, filter *models.DocumentFilter) (int64, error) {
	q := r.db.WithContext(ctx).Model(&models.Requisition{}).Where("organization_id = ?", organizationID)
	q = applyTextSearch(q, searchText)
	if filter != nil && filter.DocumentNumber != "" {
		q = q.Where("document_number ILIKE ?", "%"+filter.DocumentNumber+"%")
	}
	if filter != nil && len(filter.Statuses) > 0 {
		q = q.Where("UPPER(status) IN ?", toUpperSlice(filter.Statuses))
	}
	if filter != nil && filter.DateFrom != nil {
		q = q.Where("created_at >= ?", filter.DateFrom)
	}
	if filter != nil && filter.DateTo != nil {
		q = q.Where("created_at <= ?", filter.DateTo)
	}
	var count int64
	err := q.Count(&count).Error
	return count, err
}

func (r *DocumentRepository) countPurchaseOrders(ctx context.Context, organizationID, searchText string, filter *models.DocumentFilter) (int64, error) {
	q := r.db.WithContext(ctx).Model(&models.PurchaseOrder{}).Where("organization_id = ?", organizationID)
	q = applyTextSearch(q, searchText)
	if filter != nil && filter.DocumentNumber != "" {
		q = q.Where("document_number ILIKE ?", "%"+filter.DocumentNumber+"%")
	}
	if filter != nil && len(filter.Statuses) > 0 {
		q = q.Where("UPPER(status) IN ?", toUpperSlice(filter.Statuses))
	}
	if filter != nil && filter.DateFrom != nil {
		q = q.Where("created_at >= ?", filter.DateFrom)
	}
	if filter != nil && filter.DateTo != nil {
		q = q.Where("created_at <= ?", filter.DateTo)
	}
	var count int64
	err := q.Count(&count).Error
	return count, err
}

func (r *DocumentRepository) countPaymentVouchers(ctx context.Context, organizationID, searchText string, filter *models.DocumentFilter) (int64, error) {
	q := r.db.WithContext(ctx).Model(&models.PaymentVoucher{}).Where("organization_id = ?", organizationID)
	q = applyTextSearch(q, searchText)
	if filter != nil && filter.DocumentNumber != "" {
		q = q.Where("document_number ILIKE ?", "%"+filter.DocumentNumber+"%")
	}
	if filter != nil && len(filter.Statuses) > 0 {
		q = q.Where("UPPER(status) IN ?", toUpperSlice(filter.Statuses))
	}
	if filter != nil && filter.DateFrom != nil {
		q = q.Where("created_at >= ?", filter.DateFrom)
	}
	if filter != nil && filter.DateTo != nil {
		q = q.Where("created_at <= ?", filter.DateTo)
	}
	var count int64
	err := q.Count(&count).Error
	return count, err
}

func (r *DocumentRepository) countGRNs(ctx context.Context, organizationID, searchText string, filter *models.DocumentFilter) (int64, error) {
	q := r.db.WithContext(ctx).Model(&models.GoodsReceivedNote{}).Where("organization_id = ?", organizationID)
	q = applyTextSearch(q, searchText)
	if filter != nil && filter.DocumentNumber != "" {
		q = q.Where("document_number ILIKE ?", "%"+filter.DocumentNumber+"%")
	}
	if filter != nil && len(filter.Statuses) > 0 {
		q = q.Where("UPPER(status) IN ?", toUpperSlice(filter.Statuses))
	}
	if filter != nil && filter.DateFrom != nil {
		q = q.Where("created_at >= ?", filter.DateFrom)
	}
	if filter != nil && filter.DateTo != nil {
		q = q.Where("created_at <= ?", filter.DateTo)
	}
	var count int64
	err := q.Count(&count).Error
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

// toUpperSlice converts a slice of strings to uppercase
func toUpperSlice(slice []string) []string {
	result := make([]string, len(slice))
	for i, s := range slice {
		result[i] = strings.ToUpper(s)
	}
	return result
}

// applyFilters applies filters to a query
func (r *DocumentRepository) applyFilters(query *gorm.DB, filter *models.DocumentFilter) *gorm.DB {
	if filter == nil {
		return query
	}

	// Exact document number filter (for specific document lookup)
	if filter.DocumentNumber != "" {
		query = query.Where("document_number ILIKE ?", "%"+filter.DocumentNumber+"%")
	}

	if len(filter.DocumentTypes) > 0 {
		// Case-insensitive document type matching
		query = query.Where("UPPER(document_type) IN ?", toUpperSlice(filter.DocumentTypes))
	}

	if len(filter.Statuses) > 0 {
		// Case-insensitive status matching
		query = query.Where("UPPER(status) IN ?", toUpperSlice(filter.Statuses))
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

// GetOrganizationBranding fetches minimal branding fields for a given organization.
// This is intentionally read-only and used by public PDF generation endpoints.
func (r *DocumentRepository) GetOrganizationBranding(ctx context.Context, organizationID string) (*models.Organization, error) {
	var org models.Organization
	err := r.db.WithContext(ctx).
		Select("id, name, logo_url, tagline").
		Where("id = ?", organizationID).
		First(&org).Error
	if err != nil {
		return nil, err
	}
	return &org, nil
}

