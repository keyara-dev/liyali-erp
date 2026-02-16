package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/types"
	"github.com/liyali/liyali-gateway/utils"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// DocumentAutomationService handles automatic document generation
type DocumentAutomationService struct {
	db              *gorm.DB
	auditService    *AuditService
	notificationSvc *NotificationService
}

// AutomationConfig controls automation behavior
type AutomationConfig struct {
	AutoCreatePOFromRequisition bool
	AutoCreateGRNFromPO         bool
	AutoCreatePVFromGRN         bool
	RequireApprovalForAuto      bool
}

// AutomationResult contains the result of an automation operation
type AutomationResult struct {
	Success         bool
	CreatedDocument interface{}
	DocumentType    string
	DocumentID      string
	Error           error
}

// NewDocumentAutomationService creates a new document automation service
func NewDocumentAutomationService(
	db *gorm.DB,
	auditService *AuditService,
	notificationSvc *NotificationService,
) *DocumentAutomationService {
	return &DocumentAutomationService{
		db:              db,
		auditService:    auditService,
		notificationSvc: notificationSvc,
	}
}

// CreatePurchaseOrderFromRequisition automatically creates a PO from an approved requisition
func (s *DocumentAutomationService) CreatePurchaseOrderFromRequisition(
	ctx context.Context,
	requisition *models.Requisition,
	config AutomationConfig,
) (*AutomationResult, error) {
	if !config.AutoCreatePOFromRequisition {
		return &AutomationResult{
			Success: false,
			Error:   fmt.Errorf("automatic PO creation is disabled"),
		}, nil
	}

	if requisition.Status != "approved" {
		return &AutomationResult{
			Success: false,
			Error:   fmt.Errorf("requisition must be approved to create PO"),
		}, nil
	}

	// Handle vendor - create PO with or without vendor
	var vendorID string
	var vendorName string = "To Be Determined"

	if requisition.PreferredVendorID != nil && *requisition.PreferredVendorID != "" {
		// Verify vendor exists if provided
		var vendor models.Vendor
		if err := s.db.Where("id = ?", *requisition.PreferredVendorID).First(&vendor).Error; err != nil {
			// If vendor not found, use placeholder vendor
			vendorID = "vendor-placeholder-001"
			vendorName = "To Be Determined (Invalid Vendor)"
		} else {
			vendorID = vendor.ID
			vendorName = vendor.Name
		}
	} else {
		// No vendor specified - use placeholder vendor
		vendorID = "vendor-placeholder-001"
		vendorName = "To Be Determined"
	}

	// Generate PO number
	documentNumber := utils.GeneratePurchaseOrderNumber()

	// Convert requisition items to PO items
	var requisitionItems []types.RequisitionItem
	if len(requisition.Items.Data()) > 0 {
		requisitionItems = requisition.Items.Data()
	}

	poItems := make([]types.POItem, len(requisitionItems))
	for i, reqItem := range requisitionItems {
		poItems[i] = types.POItem{
			Description: reqItem.Description,
			Quantity:    reqItem.Quantity,
			UnitPrice:   reqItem.UnitPrice,
			Amount:      reqItem.Amount,
		}
	}

	// Create Purchase Order
	purchaseOrder := models.PurchaseOrder{
		ID:                uuid.New().String(),
		DocumentNumber:    documentNumber,
		VendorID:          vendorID, // Now can be the placeholder vendor ID
		Status:            "draft",  // Start as draft for review
		TotalAmount:       requisition.TotalAmount,
		Currency:          requisition.Currency,
		DeliveryDate:      time.Now().AddDate(0, 1, 0), // Default 1 month delivery
		ApprovalStage:     0,
		LinkedRequisition: requisition.ID,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		OrganizationID:    requisition.OrganizationID,

		// Add description to track auto-creation details
		Description: fmt.Sprintf("Auto-created from requisition %s. Vendor: %s", requisition.DocumentNumber, vendorName),

		// Copy additional fields from requisition
		Department:   requisition.Department,
		DepartmentID: requisition.DepartmentId,
		Title:        fmt.Sprintf("PO for %s", requisition.Title),
		BudgetCode:   requisition.BudgetCode,
		CostCenter:   requisition.CostCenter,
		ProjectCode:  requisition.ProjectCode,

		// Link to source requisition
		SourceRequisitionId: &requisition.ID,

		// Mark as auto-created
		AutomationUsed: true,
	}

	// Set items
	purchaseOrder.Items = datatypes.NewJSONType(poItems)
	purchaseOrder.ApprovalHistory = datatypes.NewJSONType([]types.ApprovalRecord{})

	// Save to database
	if err := s.db.Create(&purchaseOrder).Error; err != nil {
		return &AutomationResult{
			Success: false,
			Error:   fmt.Errorf("failed to create purchase order: %w", err),
		}, nil
	}

	// Log audit event with vendor info
	if s.auditService != nil {
		details := fmt.Sprintf("Auto-created PO %s from approved requisition %s (Vendor: %s)", documentNumber, requisition.DocumentNumber, vendorName)
		s.auditService.LogEvent(ctx, "system", "", "po_auto_created", "purchase_order", purchaseOrder.ID, details, "", "")
	}

	// Send notification to requisition creator
	if s.notificationSvc != nil {
		event := NotificationEvent{
			Type:         "document_created",
			DocumentID:   purchaseOrder.ID,
			DocumentType: "purchase_order",
			Action:       "auto_created",
			ActorID:      "system",
			Details:      fmt.Sprintf("Purchase Order %s was automatically created from your requisition (Vendor: %s)", documentNumber, vendorName),
			Timestamp:    time.Now(),
		}
		s.notificationSvc.HandleWorkflowEvent(event)
	}

	return &AutomationResult{
		Success:         true,
		CreatedDocument: purchaseOrder,
		DocumentType:    "purchase_order",
		DocumentID:      purchaseOrder.ID,
	}, nil
}

// CreateGRNFromPurchaseOrder automatically creates a GRN template from an approved PO
func (s *DocumentAutomationService) CreateGRNFromPurchaseOrder(
	ctx context.Context,
	purchaseOrder *models.PurchaseOrder,
	config AutomationConfig,
) (*AutomationResult, error) {
	if !config.AutoCreateGRNFromPO {
		return &AutomationResult{
			Success: false,
			Error:   fmt.Errorf("automatic GRN creation is disabled"),
		}, nil
	}

	if purchaseOrder.Status != "approved" {
		return &AutomationResult{
			Success: false,
			Error:   fmt.Errorf("purchase order must be approved to create GRN"),
		}, nil
	}

	// Generate GRN document number
	documentNumber := utils.GenerateDocumentNumber("GRN")

	// Convert PO items to GRN items
	var poItems []types.POItem
	if len(purchaseOrder.Items.Data()) > 0 {
		poItems = purchaseOrder.Items.Data()
	}

	grnItems := make([]types.GRNItem, len(poItems))
	for i, poItem := range poItems {
		grnItems[i] = types.GRNItem{
			Description:      poItem.Description,
			QuantityOrdered:  poItem.Quantity,
			QuantityReceived: 0, // To be filled when goods are received
			Variance:         0,
			Condition:        "pending", // To be updated during inspection
		}
	}

	// Create GRN
	grn := models.GoodsReceivedNote{
		ID:               uuid.New().String(),
		DocumentNumber:   documentNumber,
		PODocumentNumber: purchaseOrder.DocumentNumber,
		Status:           "draft", // Start as draft for warehouse team
		ReceivedDate:     time.Now(),
		ReceivedBy:       "", // To be filled by warehouse team
		ApprovalStage:    0,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		OrganizationID:   purchaseOrder.OrganizationID,
	}

	// Set items
	grn.Items = datatypes.NewJSONType(grnItems)
	grn.ApprovalHistory = datatypes.NewJSONType([]types.ApprovalRecord{})

	// Save to database
	if err := s.db.Create(&grn).Error; err != nil {
		return &AutomationResult{
			Success: false,
			Error:   fmt.Errorf("failed to create GRN: %w", err),
		}, nil
	}

	// Log audit event
	if s.auditService != nil {
		details := fmt.Sprintf("Auto-created GRN %s from approved PO %s", documentNumber, purchaseOrder.DocumentNumber)
		s.auditService.LogEvent(ctx, "system", "", "grn_auto_created", "grn", grn.ID, details, "", "")
	}

	// Send notification to warehouse team (assuming role-based notification)
	if s.notificationSvc != nil {
		event := NotificationEvent{
			Type:         "document_created",
			DocumentID:   grn.ID,
			DocumentType: "grn",
			Action:       "auto_created",
			ActorID:      "system",
			Details:      fmt.Sprintf("GRN %s was automatically created from PO %s and is ready for goods receipt", documentNumber, purchaseOrder.DocumentNumber),
			Timestamp:    time.Now(),
		}
		s.notificationSvc.HandleWorkflowEvent(event)
	}

	return &AutomationResult{
		Success:         true,
		CreatedDocument: grn,
		DocumentType:    "grn",
		DocumentID:      grn.ID,
	}, nil
}

// CreatePaymentVoucherFromGRN automatically creates a PV from an approved GRN
func (s *DocumentAutomationService) CreatePaymentVoucherFromGRN(
	ctx context.Context,
	grn *models.GoodsReceivedNote,
	config AutomationConfig,
) (*AutomationResult, error) {
	if !config.AutoCreatePVFromGRN {
		return &AutomationResult{
			Success: false,
			Error:   fmt.Errorf("automatic PV creation is disabled"),
		}, nil
	}

	if grn.Status != "approved" {
		return &AutomationResult{
			Success: false,
			Error:   fmt.Errorf("GRN must be approved to create payment voucher"),
		}, nil
	}

	// Get the linked PO to extract vendor and amount information
	var purchaseOrder models.PurchaseOrder
	if err := s.db.Where("document_number = ?", grn.PODocumentNumber).First(&purchaseOrder).Error; err != nil {
		return &AutomationResult{
			Success: false,
			Error:   fmt.Errorf("linked purchase order not found: %w", err),
		}, nil
	}

	// Generate PV document number
	documentNumber := utils.GenerateDocumentNumber("PV")

	// Create Payment Voucher
	paymentVoucher := models.PaymentVoucher{
		ID:             uuid.New().String(),
		DocumentNumber: documentNumber,
		VendorID:       purchaseOrder.VendorID,
		InvoiceNumber:  "",      // To be filled when invoice is received
		Status:         "draft", // Start as draft for finance team
		Amount:         purchaseOrder.TotalAmount,
		Currency:       purchaseOrder.Currency,
		PaymentMethod:  "bank_transfer", // Default payment method
		LinkedPO:       purchaseOrder.DocumentNumber,
		ApprovalStage:  0,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		OrganizationID: grn.OrganizationID,
	}

	// Initialize empty approval history
	paymentVoucher.ApprovalHistory = datatypes.NewJSONType([]types.ApprovalRecord{})

	// Save to database
	if err := s.db.Create(&paymentVoucher).Error; err != nil {
		return &AutomationResult{
			Success: false,
			Error:   fmt.Errorf("failed to create payment voucher: %w", err),
		}, nil
	}

	// Log audit event
	if s.auditService != nil {
		details := fmt.Sprintf("Auto-created PV %s from approved GRN %s", documentNumber, grn.DocumentNumber)
		s.auditService.LogEvent(ctx, "system", "", "pv_auto_created", "payment_voucher", paymentVoucher.ID, details, "", "")
	}

	// Send notification to finance team
	if s.notificationSvc != nil {
		event := NotificationEvent{
			Type:         "document_created",
			DocumentID:   paymentVoucher.ID,
			DocumentType: "payment_voucher",
			Action:       "auto_created",
			ActorID:      "system",
			Details:      fmt.Sprintf("Payment Voucher %s was automatically created from GRN %s and is ready for processing", documentNumber, grn.DocumentNumber),
			Timestamp:    time.Now(),
		}
		s.notificationSvc.HandleWorkflowEvent(event)
	}

	return &AutomationResult{
		Success:         true,
		CreatedDocument: paymentVoucher,
		DocumentType:    "payment_voucher",
		DocumentID:      paymentVoucher.ID,
	}, nil
}

// GetDefaultAutomationConfig returns the default automation configuration
func (s *DocumentAutomationService) GetDefaultAutomationConfig() AutomationConfig {
	return AutomationConfig{
		AutoCreatePOFromRequisition: false,
		AutoCreateGRNFromPO:         false,
		AutoCreatePVFromGRN:         false,
		RequireApprovalForAuto:      true,
	}
}

// ValidateAutomationPrerequisites checks if automation can proceed
func (s *DocumentAutomationService) ValidateAutomationPrerequisites(
	documentType string,
	document interface{},
) error {
	switch documentType {
	case "requisition":
		req, ok := document.(*models.Requisition)
		if !ok {
			return fmt.Errorf("invalid requisition document")
		}
		if req.Status != "approved" {
			return fmt.Errorf("requisition must be approved")
		}
		// Removed vendor requirement - PO can be created without vendor
	case "purchase_order":
		po, ok := document.(*models.PurchaseOrder)
		if !ok {
			return fmt.Errorf("invalid purchase order document")
		}
		if po.Status != "approved" {
			return fmt.Errorf("purchase order must be approved")
		}
	case "grn":
		grn, ok := document.(*models.GoodsReceivedNote)
		if !ok {
			return fmt.Errorf("invalid GRN document")
		}
		if grn.Status != "approved" {
			return fmt.Errorf("GRN must be approved")
		}
	default:
		return fmt.Errorf("unsupported document type: %s", documentType)
	}
	return nil
}
