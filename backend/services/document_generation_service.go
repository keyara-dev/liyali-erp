package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/liyali/liyali-gateway/models"
	"gorm.io/gorm"
)

// DocumentGenerationService handles explicit document generation requests.
type DocumentGenerationService struct {
	db                *gorm.DB
	automationService *DocumentAutomationService
}

// GenerateDocumentResult contains metadata about a generation operation.
type GenerateDocumentResult struct {
	SourceID         string `json:"sourceId"`
	SourceDocType    string `json:"sourceDocType"`
	GeneratedID      string `json:"generatedId"`
	GeneratedDocType string `json:"generatedDocType"`
	DocumentNumber   string `json:"documentNumber,omitempty"`
}

func NewDocumentGenerationService(db *gorm.DB, automationService *DocumentAutomationService) *DocumentGenerationService {
	return &DocumentGenerationService{
		db:                db,
		automationService: automationService,
	}
}

func (s *DocumentGenerationService) GenerateFromSource(
	ctx context.Context,
	organizationID, sourceID, docType, targetDocType string,
) (*GenerateDocumentResult, error) {
	if s.automationService == nil {
		return nil, fmt.Errorf("document generation service unavailable")
	}
	if sourceID == "" {
		return nil, fmt.Errorf("source ID is required")
	}

	sourceType := normalizeDocType(docType)
	if sourceType == "" {
		return nil, fmt.Errorf("docType is required")
	}

	targetType := normalizeDocType(targetDocType)
	if targetType != "" {
		expectedTarget, err := expectedTargetForSource(sourceType)
		if err != nil {
			return nil, err
		}
		if targetType != expectedTarget {
			return nil, fmt.Errorf("invalid targetDocType for %s", sourceType)
		}
	}

	config := AutomationConfig{
		AutoCreatePOFromRequisition: true,
		AutoCreateGRNFromPO:         true,
		AutoCreatePVFromGRN:         true,
		RequireApprovalForAuto:      true,
	}

	switch sourceType {
	case "REQUISITION":
		var req models.Requisition
		if err := s.db.Where("id = ? AND organization_id = ?", sourceID, organizationID).First(&req).Error; err != nil {
			return nil, fmt.Errorf("requisition not found")
		}
		if strings.ToUpper(req.Status) != "APPROVED" {
			return nil, fmt.Errorf("requisition must be approved")
		}

		var existing int64
		if err := s.db.Model(&models.PurchaseOrder{}).
			Where("organization_id = ? AND (source_requisition_id = ? OR linked_requisition = ?)", organizationID, req.ID, req.ID).
			Count(&existing).Error; err != nil {
			return nil, fmt.Errorf("failed to validate existing purchase order")
		}
		if existing > 0 {
			return nil, fmt.Errorf("purchase order already generated for this requisition")
		}

		result, err := s.automationService.CreatePurchaseOrderFromRequisition(ctx, &req, config)
		if err != nil {
			return nil, err
		}
		if !result.Success {
			if result.Error != nil {
				return nil, result.Error
			}
			return nil, fmt.Errorf("failed to generate purchase order")
		}
		return &GenerateDocumentResult{
			SourceID:         req.ID,
			SourceDocType:    sourceType,
			GeneratedID:      result.DocumentID,
			GeneratedDocType: "PURCHASE_ORDER",
			DocumentNumber:   extractGeneratedDocumentNumber(result.CreatedDocument),
		}, nil

	case "PURCHASE_ORDER":
		var po models.PurchaseOrder
		if err := s.db.Where("id = ? AND organization_id = ?", sourceID, organizationID).First(&po).Error; err != nil {
			return nil, fmt.Errorf("purchase order not found")
		}
		if strings.ToUpper(po.Status) != "APPROVED" {
			return nil, fmt.Errorf("purchase order must be approved")
		}

		var existing int64
		if err := s.db.Model(&models.GoodsReceivedNote{}).
			Where("organization_id = ? AND po_document_number = ?", organizationID, po.DocumentNumber).
			Count(&existing).Error; err != nil {
			return nil, fmt.Errorf("failed to validate existing GRN")
		}
		if existing > 0 {
			return nil, fmt.Errorf("GRN already generated for this purchase order")
		}

		result, err := s.automationService.CreateGRNFromPurchaseOrder(ctx, &po, config)
		if err != nil {
			return nil, err
		}
		if !result.Success {
			if result.Error != nil {
				return nil, result.Error
			}
			return nil, fmt.Errorf("failed to generate GRN")
		}
		return &GenerateDocumentResult{
			SourceID:         po.ID,
			SourceDocType:    sourceType,
			GeneratedID:      result.DocumentID,
			GeneratedDocType: "GRN",
			DocumentNumber:   extractGeneratedDocumentNumber(result.CreatedDocument),
		}, nil

	case "GRN":
		var grn models.GoodsReceivedNote
		if err := s.db.Where("id = ? AND organization_id = ?", sourceID, organizationID).First(&grn).Error; err != nil {
			return nil, fmt.Errorf("GRN not found")
		}
		if strings.ToUpper(grn.Status) != "APPROVED" {
			return nil, fmt.Errorf("GRN must be approved")
		}

		var existing int64
		if err := s.db.Model(&models.PaymentVoucher{}).
			Where("organization_id = ? AND linked_po = ?", organizationID, grn.PODocumentNumber).
			Count(&existing).Error; err != nil {
			return nil, fmt.Errorf("failed to validate existing payment voucher")
		}
		if existing > 0 {
			return nil, fmt.Errorf("payment voucher already generated for this GRN")
		}

		result, err := s.automationService.CreatePaymentVoucherFromGRN(ctx, &grn, config)
		if err != nil {
			return nil, err
		}
		if !result.Success {
			if result.Error != nil {
				return nil, result.Error
			}
			return nil, fmt.Errorf("failed to generate payment voucher")
		}
		return &GenerateDocumentResult{
			SourceID:         grn.ID,
			SourceDocType:    sourceType,
			GeneratedID:      result.DocumentID,
			GeneratedDocType: "PAYMENT_VOUCHER",
			DocumentNumber:   extractGeneratedDocumentNumber(result.CreatedDocument),
		}, nil
	default:
		return nil, fmt.Errorf("unsupported docType: %s", sourceType)
	}
}

func normalizeDocType(docType string) string {
	return strings.ToUpper(strings.TrimSpace(docType))
}

func expectedTargetForSource(sourceDocType string) (string, error) {
	switch sourceDocType {
	case "REQUISITION":
		return "PURCHASE_ORDER", nil
	case "PURCHASE_ORDER":
		return "GRN", nil
	case "GRN":
		return "PAYMENT_VOUCHER", nil
	default:
		return "", fmt.Errorf("unsupported docType: %s", sourceDocType)
	}
}

func extractGeneratedDocumentNumber(createdDocument interface{}) string {
	switch doc := createdDocument.(type) {
	case models.PurchaseOrder:
		return doc.DocumentNumber
	case *models.PurchaseOrder:
		return doc.DocumentNumber
	case models.GoodsReceivedNote:
		return doc.DocumentNumber
	case *models.GoodsReceivedNote:
		return doc.DocumentNumber
	case models.PaymentVoucher:
		return doc.DocumentNumber
	case *models.PaymentVoucher:
		return doc.DocumentNumber
	default:
		return ""
	}
}
