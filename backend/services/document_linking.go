package services

import (
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/models"
	"gorm.io/gorm"
)

// DocumentLink represents a relationship between two documents
type DocumentLink struct {
	ID              string    `gorm:"primaryKey" json:"id"`
	SourceDocID     string    `json:"sourceDocId"`     // Parent document
	SourceDocType   string    `json:"sourceDocType"`   // requisition, budget, po, etc.
	TargetDocID     string    `json:"targetDocId"`     // Child document
	TargetDocType   string    `json:"targetDocType"`
	LinkType        string    `json:"linkType"`        // creates, links_to, inherits_from
	Amount          float64   `json:"amount,omitempty"` // For partial allocations
	Proportion      float64   `json:"proportion,omitempty"` // Percentage of parent
	Status          string    `json:"status"`          // active, inactive
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// DocumentLinkingService manages document relationships
type DocumentLinkingService struct {
	db *gorm.DB
}

// NewDocumentLinkingService creates a new document linking service
func NewDocumentLinkingService(db *gorm.DB) *DocumentLinkingService {
	return &DocumentLinkingService{db: db}
}

// LinkRequisitionToBudget links a requisition to a budget allocation
func (dls *DocumentLinkingService) LinkRequisitionToBudget(
	requisitionID, budgetID string,
	amount float64,
) error {
	// Verify both documents exist
	var req models.Requisition
	if err := dls.db.First(&req, "id = ?", requisitionID).Error; err != nil {
		return fmt.Errorf("requisition not found: %v", err)
	}

	var budget models.Budget
	if err := dls.db.First(&budget, "id = ?", budgetID).Error; err != nil {
		return fmt.Errorf("budget not found: %v", err)
	}

	// Check for existing link
	var existingLink DocumentLink
	result := dls.db.Where(
		"source_doc_id = ? AND target_doc_id = ? AND link_type = ?",
		budgetID, requisitionID, "allocates_to",
	).First(&existingLink)

	if result.Error == nil {
		return fmt.Errorf("link already exists between budget and requisition")
	}

	// Create link
	link := DocumentLink{
		ID:            uuid.New().String(),
		SourceDocID:   budgetID,
		SourceDocType: "budget",
		TargetDocID:   requisitionID,
		TargetDocType: "requisition",
		LinkType:      "allocates_to",
		Amount:        amount,
		Proportion:    (amount / budget.TotalBudget) * 100,
		Status:        "active",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := dls.db.Create(&link).Error; err != nil {
		return fmt.Errorf("failed to create budget-requisition link: %v", err)
	}

	log.Printf("Linked requisition %s to budget %s with amount %.2f", requisitionID, budgetID, amount)
	return nil
}

// LinkRequisitionToPurchaseOrder links a requisition to a PO
func (dls *DocumentLinkingService) LinkRequisitionToPurchaseOrder(
	requisitionID, poID string,
) error {
	// Verify both documents exist
	var req models.Requisition
	if err := dls.db.First(&req, "id = ?", requisitionID).Error; err != nil {
		return fmt.Errorf("requisition not found: %v", err)
	}

	var po models.PurchaseOrder
	if err := dls.db.First(&po, "id = ?", poID).Error; err != nil {
		return fmt.Errorf("purchase order not found: %v", err)
	}

	// Check for existing link
	var existingLink DocumentLink
	result := dls.db.Where(
		"source_doc_id = ? AND target_doc_id = ? AND link_type = ?",
		requisitionID, poID, "creates",
	).First(&existingLink)

	if result.Error == nil {
		return fmt.Errorf("requisition already linked to this PO")
	}

	// Create link
	link := DocumentLink{
		ID:            uuid.New().String(),
		SourceDocID:   requisitionID,
		SourceDocType: "requisition",
		TargetDocID:   poID,
		TargetDocType: "po",
		LinkType:      "creates",
		Amount:        po.TotalAmount,
		Status:        "active",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := dls.db.Create(&link).Error; err != nil {
		return fmt.Errorf("failed to link requisition to PO: %v", err)
	}

	// Update PO with requisition link
	if err := dls.db.Model(&po).Update("linked_requisition", requisitionID).Error; err != nil {
		log.Printf("Warning: failed to update PO with requisition link: %v", err)
	}

	log.Printf("Linked requisition %s to purchase order %s", requisitionID, poID)
	return nil
}

// LinkPurchaseOrderToPaymentVoucher links a PO to a payment voucher
func (dls *DocumentLinkingService) LinkPurchaseOrderToPaymentVoucher(
	poID, pvID string,
	amount float64,
) error {
	// Verify both documents exist
	var po models.PurchaseOrder
	if err := dls.db.First(&po, "id = ?", poID).Error; err != nil {
		return fmt.Errorf("purchase order not found: %v", err)
	}

	var pv models.PaymentVoucher
	if err := dls.db.First(&pv, "id = ?", pvID).Error; err != nil {
		return fmt.Errorf("payment voucher not found: %v", err)
	}

	// Create link
	link := DocumentLink{
		ID:            uuid.New().String(),
		SourceDocID:   poID,
		SourceDocType: "po",
		TargetDocID:   pvID,
		TargetDocType: "pv",
		LinkType:      "creates_payment_for",
		Amount:        amount,
		Proportion:    (amount / po.TotalAmount) * 100,
		Status:        "active",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := dls.db.Create(&link).Error; err != nil {
		return fmt.Errorf("failed to link PO to payment voucher: %v", err)
	}

	// Update payment voucher with PO link
	if err := dls.db.Model(&pv).Update("linked_po", poID).Error; err != nil {
		log.Printf("Warning: failed to update PV with PO link: %v", err)
	}

	log.Printf("Linked purchase order %s to payment voucher %s with amount %.2f", poID, pvID, amount)
	return nil
}

// LinkPurchaseOrderToGRN links a PO to a GRN
func (dls *DocumentLinkingService) LinkPurchaseOrderToGRN(
	poNumber, grnID string,
) error {
	// Get PO by number
	var po models.PurchaseOrder
	if err := dls.db.Where("po_number = ?", poNumber).First(&po).Error; err != nil {
		return fmt.Errorf("purchase order not found: %v", err)
	}

	// Verify GRN exists
	var grn models.GoodsReceivedNote
	if err := dls.db.First(&grn, "id = ?", grnID).Error; err != nil {
		return fmt.Errorf("GRN not found: %v", err)
	}

	// Create link
	link := DocumentLink{
		ID:            uuid.New().String(),
		SourceDocID:   po.ID,
		SourceDocType: "po",
		TargetDocID:   grnID,
		TargetDocType: "grn",
		LinkType:      "fulfilled_by",
		Amount:        po.TotalAmount,
		Status:        "active",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := dls.db.Create(&link).Error; err != nil {
		return fmt.Errorf("failed to link PO to GRN: %v", err)
	}

	// Update GRN with PO number
	if err := dls.db.Model(&grn).Update("po_number", poNumber).Error; err != nil {
		log.Printf("Warning: failed to update GRN with PO number: %v", err)
	}

	log.Printf("Linked purchase order %s to GRN %s", po.ID, grnID)
	return nil
}

// GetLinkedDocuments returns all linked documents for a given document
func (dls *DocumentLinkingService) GetLinkedDocuments(
	documentID, docType string,
) ([]DocumentLink, error) {
	var links []DocumentLink

	// Get outgoing links (where this doc is the source)
	if err := dls.db.Where(
		"source_doc_id = ? AND source_doc_type = ? AND status = ?",
		documentID, docType, "active",
	).Find(&links).Error; err != nil {
		return nil, err
	}

	// Get incoming links (where this doc is the target)
	var incomingLinks []DocumentLink
	if err := dls.db.Where(
		"target_doc_id = ? AND target_doc_type = ? AND status = ?",
		documentID, docType, "active",
	).Find(&incomingLinks).Error; err != nil {
		return nil, err
	}

	links = append(links, incomingLinks...)
	return links, nil
}

// GetDocumentRelationshipChain returns the full chain from requisition to payment
func (dls *DocumentLinkingService) GetDocumentRelationshipChain(
	requisitionID string,
) (map[string]interface{}, error) {
	chain := map[string]interface{}{
		"requisitionId": requisitionID,
		"documents":     []map[string]interface{}{},
	}

	// Get requisition
	var req models.Requisition
	if err := dls.db.First(&req, "id = ?", requisitionID).Error; err != nil {
		return nil, fmt.Errorf("requisition not found: %v", err)
	}

	// Find linked budget
	var budgetLink DocumentLink
	if err := dls.db.Where(
		"target_doc_id = ? AND target_doc_type = ? AND link_type = ?",
		requisitionID, "requisition", "allocates_to",
	).First(&budgetLink).Error; err == nil {
		var budget models.Budget
		if err := dls.db.First(&budget, "id = ?", budgetLink.SourceDocID).Error; err == nil {
			chain["budgetId"] = budget.ID
			chain["budgetCode"] = budget.BudgetCode
		}
	}

	// Find linked PO
	var poLink DocumentLink
	if err := dls.db.Where(
		"source_doc_id = ? AND source_doc_type = ? AND link_type = ?",
		requisitionID, "requisition", "creates",
	).First(&poLink).Error; err == nil {
		var po models.PurchaseOrder
		if err := dls.db.First(&po, "id = ?", poLink.TargetDocID).Error; err == nil {
			chain["poId"] = po.ID
			chain["poNumber"] = po.PONumber
		}
	}

	// Find linked GRN if PO exists
	if poID, ok := chain["poId"]; ok {
		var grnLink DocumentLink
		if err := dls.db.Where(
			"source_doc_id = ? AND source_doc_type = ? AND link_type = ?",
			poID, "po", "fulfilled_by",
		).First(&grnLink).Error; err == nil {
			var grn models.GoodsReceivedNote
			if err := dls.db.First(&grn, "id = ?", grnLink.TargetDocID).Error; err == nil {
				chain["grnId"] = grn.ID
				chain["grnNumber"] = grn.GRNNumber
			}
		}
	}

	return chain, nil
}

// UnlinkDocuments removes a link between two documents
func (dls *DocumentLinkingService) UnlinkDocuments(
	sourceDocID, targetDocID string,
) error {
	if err := dls.db.Where(
		"source_doc_id = ? AND target_doc_id = ?",
		sourceDocID, targetDocID,
	).Delete(&DocumentLink{}).Error; err != nil {
		return fmt.Errorf("failed to unlink documents: %v", err)
	}

	log.Printf("Unlinked document %s from document %s", targetDocID, sourceDocID)
	return nil
}

// GetLinkStatistics returns statistics about document links
func (dls *DocumentLinkingService) GetLinkStatistics() (map[string]interface{}, error) {
	var stats map[string]interface{}

	// Count total links
	var totalLinks int64
	dls.db.Model(&DocumentLink{}).Where("status = ?", "active").Count(&totalLinks)

	// Count by link type
	var linkTypeCounts []map[string]interface{}
	dls.db.Model(&DocumentLink{}).
		Where("status = ?", "active").
		Select("link_type, COUNT(*) as count").
		Group("link_type").
		Scan(&linkTypeCounts)

	stats = map[string]interface{}{
		"totalLinks": totalLinks,
		"byType":     linkTypeCounts,
	}

	return stats, nil
}
