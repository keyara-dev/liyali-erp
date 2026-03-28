package handlers

import (
	"encoding/json"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/middleware"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/services"
)

// GetDocumentChain retrieves the complete document chain for any workflow document
// Supports: requisition, purchase_order, payment_voucher, grn
// Returns parent and child documents in the procurement flow
func GetDocumentChain(c *fiber.Ctx) error {
	documentID := c.Params("id")
	if documentID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Document ID is required",
		})
	}

	documentType := c.Query("documentType", "")
	if documentType == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "documentType query parameter is required",
		})
	}

	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Organization context required",
		})
	}
	orgID := tenant.OrganizationID

	// Verify document exists and belongs to org
	if err := verifyDocumentOwnership(documentID, documentType, orgID); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Document not found",
		})
	}

	// Build chain using document linking service
	dls := services.NewDocumentLinkingService(config.DB)
	rawChain, err := dls.GetDocumentRelationshipChain(documentID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to retrieve document chain",
			"error":   err.Error(),
		})
	}

	// Build response based on document type
	chain := buildDocumentChain(documentID, documentType, rawChain, orgID)

	return c.JSON(fiber.Map{
		"success": true,
		"data":    chain,
	})
}

// verifyDocumentOwnership checks if the document exists and belongs to the organization
func verifyDocumentOwnership(documentID, documentType, orgID string) error {
	switch strings.ToLower(documentType) {
	case "requisition":
		var req models.Requisition
		return config.DB.Where("id = ? AND organization_id = ?", documentID, orgID).First(&req).Error
	case "purchase_order", "purchase-order":
		var po models.PurchaseOrder
		return config.DB.Where("id = ? AND organization_id = ?", documentID, orgID).First(&po).Error
	case "payment_voucher", "payment-voucher":
		var pv models.PaymentVoucher
		return config.DB.Where("id = ? AND organization_id = ?", documentID, orgID).First(&pv).Error
	case "grn", "goods_received_note":
		var grn models.GoodsReceivedNote
		return config.DB.Where("id = ? AND organization_id = ?", documentID, orgID).First(&grn).Error
	default:
		return fiber.NewError(fiber.StatusBadRequest, "Invalid document type")
	}
}

// buildDocumentChain constructs the document chain response
func buildDocumentChain(documentID, documentType string, rawChain fiber.Map, orgID string) fiber.Map {
	chain := fiber.Map{
		"documentId":   documentID,
		"documentType": documentType,
	}

	// Add parent documents (documents that came before this one)
	parentDocs := []fiber.Map{}

	// Add requisition if exists
	if reqID, ok := rawChain["requisitionId"].(string); ok && reqID != "" {
		var req models.Requisition
		if err := config.DB.Where("id = ? AND organization_id = ?", reqID, orgID).First(&req).Error; err == nil {
			parentDocs = append(parentDocs, fiber.Map{
				"id":             req.ID,
				"type":           "requisition",
				"documentNumber": req.DocumentNumber,
				"status":         req.Status,
				"title":          req.Title,
			})
		}
	}

	// Add PO if exists (and not the current document)
	if poID, ok := rawChain["poId"].(string); ok && poID != "" && poID != documentID {
		var po models.PurchaseOrder
		if err := config.DB.Where("id = ? AND organization_id = ?", poID, orgID).First(&po).Error; err == nil {
			parentDocs = append(parentDocs, fiber.Map{
				"id":             po.ID,
				"type":           "purchase_order",
				"documentNumber": po.DocumentNumber,
				"status":         po.Status,
				"vendorName":     po.VendorName,
			})
		}
	}

	// Add GRN if exists (and not the current document) - for goods-first flow
	if grnID, ok := rawChain["grnId"].(string); ok && grnID != "" && grnID != documentID {
		var grn models.GoodsReceivedNote
		if err := config.DB.Where("id = ? AND organization_id = ?", grnID, orgID).First(&grn).Error; err == nil {
			// For payment vouchers, GRN is a parent document in goods-first flow
			if strings.ToLower(documentType) == "payment_voucher" || strings.ToLower(documentType) == "payment-voucher" {
				parentDocs = append(parentDocs, fiber.Map{
					"id":             grn.ID,
					"type":           "grn",
					"documentNumber": grn.DocumentNumber,
					"status":         grn.Status,
				})
			}
		}
	}

	chain["parentDocuments"] = parentDocs

	// Add child documents (documents that came after this one)
	childDocs := []fiber.Map{}

	// For requisitions and POs, add GRN if it exists
	if strings.ToLower(documentType) == "requisition" || strings.ToLower(documentType) == "purchase_order" || strings.ToLower(documentType) == "purchase-order" {
		if grnID, ok := rawChain["grnId"].(string); ok && grnID != "" {
			var grn models.GoodsReceivedNote
			if err := config.DB.Where("id = ? AND organization_id = ?", grnID, orgID).First(&grn).Error; err == nil {
				childDocs = append(childDocs, fiber.Map{
					"id":             grn.ID,
					"type":           "grn",
					"documentNumber": grn.DocumentNumber,
					"status":         grn.Status,
				})
			}
		}
	}

	// For requisitions, POs, and GRNs, add PV if it exists
	if strings.ToLower(documentType) != "payment_voucher" && strings.ToLower(documentType) != "payment-voucher" {
		// Look up PV linked to the PO
		if poDocNum, ok := rawChain["poDocumentNumber"].(string); ok && poDocNum != "" {
			var pv models.PaymentVoucher
			if err := config.DB.Where("linked_po = ? AND organization_id = ?", poDocNum, orgID).First(&pv).Error; err == nil {
				childDocs = append(childDocs, fiber.Map{
					"id":             pv.ID,
					"type":           "payment_voucher",
					"documentNumber": pv.DocumentNumber,
					"status":         pv.Status,
					"vendorName":     pv.VendorName,
				})
			}
		}
	}

	chain["childDocuments"] = childDocs

	// Detect procurement flow type
	procurementFlow := "payment_first" // Default
	if len(parentDocs) > 0 {
		// Check if GRN exists in parent documents (goods-first flow)
		for _, doc := range parentDocs {
			if docType, ok := doc["type"].(string); ok && docType == "grn" {
				procurementFlow = "goods_first"
				break
			}
		}
	}
	chain["procurementFlow"] = procurementFlow

	// Detect routing type from workflow assignment (for requisitions)
	if strings.ToLower(documentType) == "requisition" {
		routingType := "procurement"
		var wa models.WorkflowAssignment
		if err := config.DB.Preload("Workflow").
			Where("entity_id = ? AND entity_type = ? AND organization_id = ?", documentID, "requisition", orgID).
			First(&wa).Error; err == nil && wa.Workflow != nil {
			var wfConditions models.WorkflowConditions
			if jsonErr := json.Unmarshal(wa.Workflow.Conditions, &wfConditions); jsonErr == nil {
				if strings.EqualFold(wfConditions.RoutingType, "accounting") {
					routingType = "accounting"
				}
			}
		}
		chain["routingType"] = routingType
	}

	return chain
}
