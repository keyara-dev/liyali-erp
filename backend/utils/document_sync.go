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
func SyncDocument(db *gorm.DB, entityType, entityID string) {
	if err := syncDocument(db, strings.ToUpper(entityType), entityID); err != nil {
		log.Printf("[document_sync] failed to sync %s %s: %v", entityType, entityID, err)
	}
}

func syncDocument(db *gorm.DB, entityType, entityID string) error {
	switch entityType {
	case "REQUISITION":
		return syncRequisition(db, entityID)
	case "PURCHASE_ORDER":
		return syncPurchaseOrder(db, entityID)
	case "PAYMENT_VOUCHER":
		return syncPaymentVoucher(db, entityID)
	case "GRN":
		return syncGRN(db, entityID)
	case "BUDGET":
		return syncBudget(db, entityID)
	default:
		return nil
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

func syncRequisition(db *gorm.DB, id string) error {
	var req models.Requisition
	if err := db.Preload("Requester").Where("id = ?", id).First(&req).Error; err != nil {
		return err
	}

	data, _ := json.Marshal(map[string]interface{}{
		"id":               req.ID,
		"documentNumber":   req.DocumentNumber,
		"items":            req.Items,
		"priority":         req.Priority,
		"categoryId":       req.CategoryID,
		"preferredVendorId": req.PreferredVendorID,
		"isEstimate":       req.IsEstimate,
		"approvalStage":    req.ApprovalStage,
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
	if req.Description != "" {
		doc.Description = &req.Description
	}
	if req.Department != "" {
		doc.Department = &req.Department
	}
	return upsertDocument(db, doc)
}

func syncPurchaseOrder(db *gorm.DB, id string) error {
	var po models.PurchaseOrder
	if err := db.Where("id = ?", id).First(&po).Error; err != nil {
		return err
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
		CreatedBy:      "system",
		Data:           datatypes.JSON(data),
		CreatedAt:      po.CreatedAt,
		UpdatedAt:      time.Now(),
	}
	return upsertDocument(db, doc)
}

func syncPaymentVoucher(db *gorm.DB, id string) error {
	var pv models.PaymentVoucher
	if err := db.Where("id = ?", id).First(&pv).Error; err != nil {
		return err
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
		CreatedBy:      "system",
		Data:           datatypes.JSON(data),
		CreatedAt:      pv.CreatedAt,
		UpdatedAt:      time.Now(),
	}
	if pv.Description != "" {
		doc.Description = &pv.Description
	}
	return upsertDocument(db, doc)
}

func syncGRN(db *gorm.DB, id string) error {
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
	return upsertDocument(db, doc)
}

func syncBudget(db *gorm.DB, id string) error {
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
	if budget.Department != "" {
		doc.Department = &budget.Department
	}
	return upsertDocument(db, doc)
}
