package utils

import (
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// SyncDocument fetches the latest entity data and upserts it into the generic
// documents table. It is safe to call as a goroutine after any entity mutation.
// Errors are logged but not propagated — the caller's response is unaffected.
//
// Use SyncDocumentAs when the acting user is known (submit, approve, update).
func SyncDocument(db *gorm.DB, entityType, entityID string) {
	SyncDocumentAs(db, entityType, entityID, "")
}

// SyncDocumentAs is SyncDocument with an attribution for documents.updated_by.
// Pass the authenticated user ID of whoever triggered the mutation (submitter,
// approver, editor). Pass "" for system/automated mutations — UpdatedBy stays
// NULL in that case.
func SyncDocumentAs(db *gorm.DB, entityType, entityID, updaterID string) {
	// Almost always launched as `go SyncDocumentAs(...)` — recover so a panic
	// here can't crash the process.
	defer RecoverPanic("document_sync.SyncDocumentAs")

	if err := syncDocument(db, strings.ToUpper(entityType), entityID, updaterID); err != nil {
		log.Printf("[document_sync] failed to sync %s %s: %v", entityType, entityID, err)
	}
}

func syncDocument(db *gorm.DB, entityType, entityID, updaterID string) error {
	switch entityType {
	case "REQUISITION":
		return syncRequisition(db, entityID, updaterID)
	case "PURCHASE_ORDER":
		return syncPurchaseOrder(db, entityID, updaterID)
	case "PAYMENT_VOUCHER":
		return syncPaymentVoucher(db, entityID, updaterID)
	case "GRN":
		return syncGRN(db, entityID, updaterID)
	case "BUDGET":
		return syncBudget(db, entityID, updaterID)
	default:
		return nil
	}
}

// applyUpdater sets doc.UpdatedBy when an actor is known. Empty string leaves
// UpdatedBy NULL (appropriate for system/automated syncs and for initial
// creation where created_by already records the actor).
func applyUpdater(doc *models.Document, updaterID string) {
	if updaterID != "" {
		id := updaterID
		doc.UpdatedBy = &id
	}
}

// upsertDocument performs an atomic upsert keyed by document_number.
// If a record with that document_number already exists it is updated in-place;
// otherwise a new record is created.
func upsertDocument(db *gorm.DB, doc *models.Document) error {
	var existing models.Document
	err := db.Where("document_number = ?", doc.DocumentNumber).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		doc.ID = uuid.New()
		return db.Create(doc).Error
	}
	if err != nil {
		return err
	}
	doc.ID = existing.ID
	doc.CreatedAt = existing.CreatedAt // preserve original creation time
	return db.Save(doc).Error
}

func syncRequisition(db *gorm.DB, id, updaterID string) error {
	var req models.Requisition
	if err := db.Preload("Requester").Where("id = ?", id).First(&req).Error; err != nil {
		return err
	}

	data, _ := json.Marshal(map[string]interface{}{
		"id":                req.ID,
		"documentNumber":    req.DocumentNumber,
		"items":             req.Items,
		"priority":          req.Priority,
		"categoryId":        req.CategoryID,
		"preferredVendorId": req.PreferredVendorID,
		"isEstimate":        req.IsEstimate,
		"approvalStage":     req.ApprovalStage,
	})

	doc := &models.Document{
		OrganizationID: req.OrganizationID,
		DocumentType:   "REQUISITION",
		DocumentNumber: req.DocumentNumber,
		Title:          req.Title,
		Status:         req.Status,
		Amount:         &req.TotalAmount,
		Currency:       &req.Currency,
		CreatedBy:      req.RequesterId,
		Data:           datatypes.JSON(data),
		CreatedAt:      req.CreatedAt,
		UpdatedAt:      time.Now(),
	}
	applyUpdater(doc, updaterID)
	if req.Description != "" {
		doc.Description = &req.Description
	}
	if req.Department != "" {
		doc.Department = &req.Department
	}
	return upsertDocument(db, doc)
}

func syncPurchaseOrder(db *gorm.DB, id, updaterID string) error {
	var po models.PurchaseOrder
	if err := db.Where("id = ?", id).First(&po).Error; err != nil {
		return err
	}

	createdBy := resolvePOCreator(db, &po)
	if createdBy == "" {
		// documents.created_by has a NOT NULL FK to users(id); skip rather than
		// violate the constraint. The PO itself is unaffected.
		log.Printf("[document_sync] skipping PO %s — no valid creator to attribute", po.DocumentNumber)
		return nil
	}

	title := po.Title
	if title == "" {
		title = "Purchase Order " + po.DocumentNumber
	}

	data, _ := json.Marshal(map[string]interface{}{
		"id":                po.ID,
		"documentNumber":    po.DocumentNumber,
		"vendorId":          po.VendorID,
		"items":             po.Items,
		"deliveryDate":      po.DeliveryDate,
		"linkedRequisition": po.LinkedRequisition,
		"approvalStage":     po.ApprovalStage,
	})

	doc := &models.Document{
		OrganizationID: po.OrganizationID,
		DocumentType:   "PURCHASE_ORDER",
		DocumentNumber: po.DocumentNumber,
		Title:          title,
		Status:         po.Status,
		Amount:         &po.TotalAmount,
		Currency:       &po.Currency,
		CreatedBy:      createdBy,
		Data:           datatypes.JSON(data),
		CreatedAt:      po.CreatedAt,
		UpdatedAt:      time.Now(),
	}
	applyUpdater(doc, updaterID)
	return upsertDocument(db, doc)
}

// resolvePOCreator returns a user ID suitable for documents.created_by.
// Prefers po.CreatedBy, then falls back to the linked requisition's requester.
// Returns "" when no attributable user can be found.
func resolvePOCreator(db *gorm.DB, po *models.PurchaseOrder) string {
	if po.CreatedBy != "" {
		return po.CreatedBy
	}
	if po.LinkedRequisition != "" {
		var req models.Requisition
		if err := db.Select("requester_id").
			Where("document_number = ?", po.LinkedRequisition).
			First(&req).Error; err == nil && req.RequesterId != "" {
			return req.RequesterId
		}
	}
	return ""
}

func syncPaymentVoucher(db *gorm.DB, id, updaterID string) error {
	var pv models.PaymentVoucher
	if err := db.Where("id = ?", id).First(&pv).Error; err != nil {
		return err
	}

	createdBy := resolvePVCreator(db, &pv)
	if createdBy == "" {
		log.Printf("[document_sync] skipping PV %s — no valid creator to attribute", pv.DocumentNumber)
		return nil
	}

	data, _ := json.Marshal(map[string]interface{}{
		"id":             pv.ID,
		"documentNumber": pv.DocumentNumber,
		"vendorId":       pv.VendorID,
		"invoiceNumber":  pv.InvoiceNumber,
		"paymentMethod":  pv.PaymentMethod,
		"glCode":         pv.GLCode,
		"linkedPO":       pv.LinkedPO,
		"approvalStage":  pv.ApprovalStage,
	})

	doc := &models.Document{
		OrganizationID: pv.OrganizationID,
		DocumentType:   "PAYMENT_VOUCHER",
		DocumentNumber: pv.DocumentNumber,
		Title:          "Payment Voucher " + pv.DocumentNumber,
		Status:         pv.Status,
		Amount:         &pv.Amount,
		Currency:       &pv.Currency,
		CreatedBy:      createdBy,
		Data:           datatypes.JSON(data),
		CreatedAt:      pv.CreatedAt,
		UpdatedAt:      time.Now(),
	}
	applyUpdater(doc, updaterID)
	if pv.Description != "" {
		doc.Description = &pv.Description
	}
	return upsertDocument(db, doc)
}

// resolvePVCreator returns a user ID suitable for documents.created_by.
// Prefers pv.CreatedBy, then falls back through the linked PO's creator.
func resolvePVCreator(db *gorm.DB, pv *models.PaymentVoucher) string {
	if pv.CreatedBy != "" {
		return pv.CreatedBy
	}
	if pv.LinkedPO != "" {
		var po models.PurchaseOrder
		if err := db.Where("document_number = ?", pv.LinkedPO).First(&po).Error; err == nil {
			return resolvePOCreator(db, &po)
		}
	}
	return ""
}

func syncGRN(db *gorm.DB, id, updaterID string) error {
	var grn models.GoodsReceivedNote
	if err := db.Where("id = ?", id).First(&grn).Error; err != nil {
		return err
	}

	title := "Goods Received Note " + grn.DocumentNumber
	if grn.Notes != "" {
		title = grn.Notes
	}

	data, _ := json.Marshal(map[string]interface{}{
		"id":               grn.ID,
		"documentNumber":   grn.DocumentNumber,
		"poDocumentNumber": grn.PODocumentNumber,
		"items":            grn.Items,
		"receivedDate":     grn.ReceivedDate,
		"receivedBy":       grn.ReceivedBy,
		"approvalStage":    grn.ApprovalStage,
	})

	doc := &models.Document{
		OrganizationID: grn.OrganizationID,
		DocumentType:   "GRN",
		DocumentNumber: grn.DocumentNumber,
		Title:          title,
		Status:         grn.Status,
		CreatedBy:      grn.ReceivedBy,
		Data:           datatypes.JSON(data),
		CreatedAt:      grn.CreatedAt,
		UpdatedAt:      time.Now(),
	}
	applyUpdater(doc, updaterID)
	return upsertDocument(db, doc)
}

func syncBudget(db *gorm.DB, id, updaterID string) error {
	var budget models.Budget
	if err := db.Where("id = ?", id).First(&budget).Error; err != nil {
		return err
	}

	if budget.BudgetCode == "" {
		return nil // Budgets without a budget code cannot be indexed
	}

	data, _ := json.Marshal(map[string]interface{}{
		"id":              budget.ID,
		"budgetCode":      budget.BudgetCode,
		"fiscalYear":      budget.FiscalYear,
		"totalBudget":     budget.TotalBudget,
		"allocatedAmount": budget.AllocatedAmount,
		"remainingAmount": budget.RemainingAmount,
		"approvalStage":   budget.ApprovalStage,
	})

	title := budget.BudgetCode + " – " + budget.FiscalYear

	doc := &models.Document{
		OrganizationID: budget.OrganizationID,
		DocumentType:   "BUDGET",
		DocumentNumber: budget.BudgetCode,
		Title:          title,
		Status:         budget.Status,
		Amount:         &budget.TotalBudget,
		CreatedBy:      budget.OwnerID,
		Data:           datatypes.JSON(data),
		CreatedAt:      budget.CreatedAt,
		UpdatedAt:      time.Now(),
	}
	applyUpdater(doc, updaterID)
	if budget.Department != "" {
		doc.Department = &budget.Department
	}
	return upsertDocument(db, doc)
}
