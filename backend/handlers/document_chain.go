package handlers

import (
	"encoding/json"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/middleware"
	"github.com/liyali/liyali-gateway/models"
	"gorm.io/datatypes"
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

	// Resolve the chain from the document's own link fields, anchored on the
	// requested document. The legacy DocumentLink-based walker is
	// requisition-rooted and reads a table nothing in the live flow populates,
	// so it 500s when handed a non-requisition id and otherwise returns an
	// empty chain. The direct fields (source_requisition_id, linked_po,
	// po_document_number, linked_grn) are the real source of truth.
	rawChain := resolveDocumentChain(documentID, documentType, orgID)

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

// resolveDocumentChain builds the raw chain (requisitionId / poId /
// poDocumentNumber / grnId) for any document type using the documents' direct
// link fields, anchored on the requested document. All lookups are scoped to
// the org and tolerate missing related documents (partial chains are fine).
func resolveDocumentChain(documentID, documentType, orgID string) fiber.Map {
	raw := fiber.Map{}
	db := config.DB

	// addPOAndAncestors records a PO (by model) plus its source requisition.
	addPOAndAncestors := func(po *models.PurchaseOrder) {
		raw["poId"] = po.ID
		raw["poDocumentNumber"] = po.DocumentNumber
		if po.SourceRequisitionId != nil && *po.SourceRequisitionId != "" {
			raw["requisitionId"] = *po.SourceRequisitionId
		}
		// GRN linked to this PO (goods receipt), if any.
		var grn models.GoodsReceivedNote
		if err := db.Where("po_document_number = ? AND organization_id = ?",
			po.DocumentNumber, orgID).First(&grn).Error; err == nil {
			raw["grnId"] = grn.ID
		}
	}

	findPOByDocNumber := func(docNum string) {
		if docNum == "" {
			return
		}
		var po models.PurchaseOrder
		if err := db.Where("document_number = ? AND organization_id = ?",
			docNum, orgID).First(&po).Error; err == nil {
			addPOAndAncestors(&po)
		} else {
			raw["poDocumentNumber"] = docNum
		}
	}

	switch strings.ToLower(documentType) {
	case "requisition":
		raw["requisitionId"] = documentID
		var po models.PurchaseOrder
		if err := db.Where("source_requisition_id = ? AND organization_id = ?",
			documentID, orgID).First(&po).Error; err == nil {
			addPOAndAncestors(&po)
		}

	case "purchase_order", "purchase-order":
		var po models.PurchaseOrder
		if err := db.Where("id = ? AND organization_id = ?",
			documentID, orgID).First(&po).Error; err == nil {
			addPOAndAncestors(&po)
		}

	case "payment_voucher", "payment-voucher":
		var pv models.PaymentVoucher
		if err := db.Where("id = ? AND organization_id = ?",
			documentID, orgID).First(&pv).Error; err == nil {
			findPOByDocNumber(pv.LinkedPO)
			if pv.LinkedGRN != "" {
				var grn models.GoodsReceivedNote
				if err := db.Where("document_number = ? AND organization_id = ?",
					pv.LinkedGRN, orgID).First(&grn).Error; err == nil {
					raw["grnId"] = grn.ID
				}
			}
		}

	case "grn", "goods_received_note":
		var grn models.GoodsReceivedNote
		if err := db.Where("id = ? AND organization_id = ?",
			documentID, orgID).First(&grn).Error; err == nil {
			raw["grnId"] = grn.ID
			findPOByDocNumber(grn.PODocumentNumber)
		}
	}

	return raw
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

	// For requisitions, POs, and GRNs, add PV(s) if any exist. Partial
	// payments (Task B) mean a PO can have MULTIPLE live PVs, so this looks
	// up ALL of them (not just the first) ordered oldest-first. childDocuments
	// is already an array, so this is purely additive — a consumer that only
	// reads the first entry still gets a valid single PV, unchanged from
	// before this patch.
	if strings.ToLower(documentType) != "payment_voucher" && strings.ToLower(documentType) != "payment-voucher" {
		// Look up PVs linked to the PO
		if poDocNum, ok := rawChain["poDocumentNumber"].(string); ok && poDocNum != "" {
			var pvs []models.PaymentVoucher
			config.DB.Where("linked_po = ? AND organization_id = ?", poDocNum, orgID).
				Order("created_at ASC").Find(&pvs)
			for _, pv := range pvs {
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

// ============================================================================
// CHAIN-WIDE SUPPORTING-DOCUMENT ATTACHMENTS
// GET /api/v1/document-chain/:id/attachments
// ============================================================================

// chainDocRef is a compact reference to one document in the resolved chain.
type chainDocRef struct {
	ID             string `json:"id"`
	Type           string `json:"type"`
	DocumentNumber string `json:"documentNumber"`
}

// ChainAttachment is one supporting-document attachment aggregated from
// anywhere in a procurement chain (requisition, PO, GRN, or PV), including
// proof-of-payment. See GetDocumentChainAttachments.
type ChainAttachment struct {
	Kind            string `json:"kind"` // attachment | quotation | proof_of_payment
	SourceDocType   string `json:"sourceDocType"`
	SourceDocID     string `json:"sourceDocId"`
	SourceDocNumber string `json:"sourceDocNumber"`
	FileID          string `json:"fileId,omitempty"`
	FileName        string `json:"fileName"`
	FileURL         string `json:"fileUrl,omitempty"` // empty for POP — never a downloadable URL
	FileSize        int64  `json:"fileSize,omitempty"`
	MimeType        string `json:"mimeType,omitempty"`
	UploadedAt      string `json:"uploadedAt,omitempty"`
	UploadedBy      string `json:"uploadedBy,omitempty"`
	FromRequisition bool   `json:"fromRequisition,omitempty"`
	Category        string `json:"category,omitempty"`
	DownloadRef     string `json:"downloadRef,omitempty"` // POP: "/payment-vouchers/{id}"
}

// resolveChainDocumentSet returns EVERY live document in the procurement
// chain anchored on the requested document: the source requisition (0..1),
// the purchase order (0..1), every non-cancelled GRN raised against that PO
// (0..n), and every non-cancelled PV raised against that PO (0..n), ordered
// oldest-first within each group.
//
// This intentionally does NOT reuse/overload resolveDocumentChain, which is
// single-slot (one grnId/pvId) and used by the existing GetDocumentChain
// endpoint — changing its shape would ripple into every consumer of that
// response. Since partial payments (Task B) mean a PO can now have multiple
// live PVs, and payment-first flows allow one GRN per PV (so GRNs can be
// plural too), this resolver anchors to the PO the same way
// resolveDocumentChain does and then Find()s the full set instead of
// First()ing a single row.
func resolveChainDocumentSet(documentID, documentType, orgID string) []chainDocRef {
	db := config.DB
	refs := []chainDocRef{}

	var po models.PurchaseOrder
	havePO := false

	// anchor is the requested document's own ref. It must ALWAYS end up in
	// the returned set — even when it has no PO to resolve a chain from
	// (e.g. a payment-first PV with linked_po=="", or a GRN with an empty
	// po_document_number) — otherwise the endpoint returns zero attachments
	// and hides the anchor's own metadata.attachments/POP. Appended via
	// appendAnchor() below, which no-ops if the anchor already made it into
	// refs through the normal PO -> GRNs -> PVs resolution.
	var anchor *chainDocRef

	switch strings.ToLower(documentType) {
	case "requisition":
		var req models.Requisition
		if err := db.Where("id = ? AND organization_id = ?", documentID, orgID).First(&req).Error; err == nil {
			ref := chainDocRef{ID: req.ID, Type: "requisition", DocumentNumber: req.DocumentNumber}
			refs = append(refs, ref)
			anchor = &ref
		}
		if err := db.Where("source_requisition_id = ? AND organization_id = ?", documentID, orgID).First(&po).Error; err == nil {
			havePO = true
		}

	case "purchase_order", "purchase-order":
		if err := db.Where("id = ? AND organization_id = ?", documentID, orgID).First(&po).Error; err == nil {
			havePO = true
			ref := chainDocRef{ID: po.ID, Type: "purchase_order", DocumentNumber: po.DocumentNumber}
			anchor = &ref
		}

	case "payment_voucher", "payment-voucher":
		var pv models.PaymentVoucher
		if err := db.Where("id = ? AND organization_id = ?", documentID, orgID).First(&pv).Error; err == nil {
			ref := chainDocRef{ID: pv.ID, Type: "payment_voucher", DocumentNumber: pv.DocumentNumber}
			anchor = &ref
			if pv.LinkedPO != "" {
				if err := db.Where("document_number = ? AND organization_id = ?", pv.LinkedPO, orgID).First(&po).Error; err == nil {
					havePO = true
				}
			}
		}

	case "grn", "goods_received_note":
		var grn models.GoodsReceivedNote
		if err := db.Where("id = ? AND organization_id = ?", documentID, orgID).First(&grn).Error; err == nil {
			ref := chainDocRef{ID: grn.ID, Type: "grn", DocumentNumber: grn.DocumentNumber}
			anchor = &ref
			if grn.PODocumentNumber != "" {
				if err := db.Where("document_number = ? AND organization_id = ?", grn.PODocumentNumber, orgID).First(&po).Error; err == nil {
					havePO = true
				}
			}
		}
	}

	appendAnchor := func(rs []chainDocRef) []chainDocRef {
		if anchor == nil {
			return rs
		}
		for _, r := range rs {
			if r.ID == anchor.ID && r.Type == anchor.Type {
				return rs
			}
		}
		return append(rs, *anchor)
	}

	if !havePO {
		return appendAnchor(refs)
	}

	// Source requisition — skip when the anchor IS the requisition (already
	// added above) to avoid a duplicate entry.
	if strings.ToLower(documentType) != "requisition" && po.SourceRequisitionId != nil && *po.SourceRequisitionId != "" {
		var req models.Requisition
		if err := db.Where("id = ? AND organization_id = ?", *po.SourceRequisitionId, orgID).First(&req).Error; err == nil {
			refs = append(refs, chainDocRef{ID: req.ID, Type: "requisition", DocumentNumber: req.DocumentNumber})
		}
	}

	refs = append(refs, chainDocRef{ID: po.ID, Type: "purchase_order", DocumentNumber: po.DocumentNumber})

	var grns []models.GoodsReceivedNote
	db.Where("po_document_number = ? AND organization_id = ? AND UPPER(status) != 'CANCELLED'",
		po.DocumentNumber, orgID).Order("created_at ASC").Find(&grns)
	for _, g := range grns {
		refs = append(refs, chainDocRef{ID: g.ID, Type: "grn", DocumentNumber: g.DocumentNumber})
	}

	var pvs []models.PaymentVoucher
	db.Where("linked_po = ? AND organization_id = ? AND UPPER(status) != 'CANCELLED'",
		po.DocumentNumber, orgID).Order("created_at ASC").Find(&pvs)
	for _, p := range pvs {
		refs = append(refs, chainDocRef{ID: p.ID, Type: "payment_voucher", DocumentNumber: p.DocumentNumber})
	}

	return appendAnchor(refs)
}

// chainAttachmentFromMap tolerantly extracts one attachment/quotation entry
// from the loosely-typed JSON blobs stored in metadata.attachments /
// metadata.quotations. The live upload paths (requisition, PO, PV detail
// clients — see frontend/src/types/{requisition,purchase-order}.ts) all
// write fileId/fileName/fileUrl/fileSize/mimeType/uploadedAt, but this also
// tolerates the shorter id/name/url/size aliases defensively so older or
// hand-crafted metadata blobs still surface instead of being silently
// dropped.
func chainAttachmentFromMap(m map[string]interface{}, sourceDocType, sourceDocID, sourceDocNumber, kind, category string) ChainAttachment {
	str := func(keys ...string) string {
		for _, k := range keys {
			if v, ok := m[k]; ok {
				if s, ok2 := v.(string); ok2 && s != "" {
					return s
				}
			}
		}
		return ""
	}

	var size int64
	for _, k := range []string{"fileSize", "size", "sizeBytes"} {
		if v, ok := m[k]; ok {
			switch n := v.(type) {
			case float64:
				size = int64(n)
			case int64:
				size = n
			case int:
				size = int64(n)
			}
			if size != 0 {
				break
			}
		}
	}

	fromReq, _ := m["fromRequisition"].(bool)

	cat := category
	if cat == "" {
		cat = str("category")
	}

	return ChainAttachment{
		Kind:            kind,
		SourceDocType:   sourceDocType,
		SourceDocID:     sourceDocID,
		SourceDocNumber: sourceDocNumber,
		FileID:          str("fileId", "id"),
		FileName:        str("fileName", "name"),
		FileURL:         str("fileUrl", "url"),
		FileSize:        size,
		MimeType:        str("mimeType", "mimetype", "contentType"),
		UploadedAt:      str("uploadedAt"),
		UploadedBy:      str("uploadedBy"),
		FromRequisition: fromReq,
		Category:        cat,
	}
}

// parseDocMetadataAttachments extracts every attachment (kind="attachment")
// and quotation (kind="quotation", category="quotation") entry from a
// document's metadata JSONB blob. Applies to all four document types —
// requisition attachments live ONLY in metadata (there is no top-level
// column), and PO/GRN/PV follow the same convention.
func parseDocMetadataAttachments(metadata datatypes.JSON, sourceDocType, sourceDocID, sourceDocNumber string) []ChainAttachment {
	out := []ChainAttachment{}
	if len(metadata) == 0 {
		return out
	}

	var meta map[string]interface{}
	if err := json.Unmarshal(metadata, &meta); err != nil {
		return out
	}

	appendEntries := func(raw interface{}, kind, category string) {
		slice, ok := raw.([]interface{})
		if !ok {
			return
		}
		for _, item := range slice {
			m, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			att := chainAttachmentFromMap(m, sourceDocType, sourceDocID, sourceDocNumber, kind, category)
			if att.FileID == "" && att.FileName == "" && att.FileURL == "" {
				continue // skip empty/malformed entries
			}
			out = append(out, att)
		}
	}

	if raw, ok := meta["attachments"]; ok {
		appendEntries(raw, "attachment", "")
	}
	if raw, ok := meta["quotations"]; ok {
		appendEntries(raw, "quotation", "quotation")
	}

	return out
}

// chainPOPMeta mirrors the JSON keys MarkPaidWithPOP writes into
// PaymentVoucher.ProofOfPayment (document_extras_handler.go, ~1095-1103):
// id/fileName/mimeType/sizeBytes/uploadedAt/uploadedBy. The base64 file
// content is stored under "dataBase64" — deliberately NOT a field on this
// struct, so json.Unmarshal silently discards it and it can never reach an
// API response.
type chainPOPMeta struct {
	ID         string `json:"id"`
	FileName   string `json:"fileName"`
	MimeType   string `json:"mimeType"`
	SizeBytes  int64  `json:"sizeBytes"`
	UploadedAt string `json:"uploadedAt"`
	UploadedBy string `json:"uploadedBy"`
}

// GetDocumentChainAttachments aggregates every supporting-document
// attachment across the FULL procurement chain anchored on the requested
// document: the source requisition, the PO, every GRN, and every PV
// (partial payments mean a PO can have more than one live PV — see
// resolveChainDocumentSet). Each PV's proof-of-payment (if any) is also
// surfaced as a "proof_of_payment" entry with a downloadRef instead of the
// raw base64 blob.
//
// Dedupe: entries are deduped by FileID, first occurrence wins, in
// REQ -> PO -> GRN(s) -> PV(s) order — so a PO attachment copied from its
// source requisition (fromRequisition=true, same fileId) collapses to the
// requisition's entry rather than appearing twice.
func GetDocumentChainAttachments(c *fiber.Ctx) error {
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

	if err := verifyDocumentOwnership(documentID, documentType, orgID); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Document not found",
		})
	}

	docs := resolveChainDocumentSet(documentID, documentType, orgID)

	attachments := []ChainAttachment{}
	seenFileIDs := map[string]bool{}
	addUnique := func(att ChainAttachment) {
		if att.FileID != "" {
			if seenFileIDs[att.FileID] {
				return
			}
			seenFileIDs[att.FileID] = true
		}
		attachments = append(attachments, att)
	}

	db := config.DB

	for _, d := range docs {
		switch d.Type {
		case "requisition":
			var req models.Requisition
			if err := db.Where("id = ? AND organization_id = ?", d.ID, orgID).First(&req).Error; err == nil {
				for _, att := range parseDocMetadataAttachments(req.Metadata, "requisition", req.ID, req.DocumentNumber) {
					addUnique(att)
				}
			}

		case "purchase_order":
			var po models.PurchaseOrder
			if err := db.Where("id = ? AND organization_id = ?", d.ID, orgID).First(&po).Error; err == nil {
				for _, att := range parseDocMetadataAttachments(po.Metadata, "purchase_order", po.ID, po.DocumentNumber) {
					addUnique(att)
				}
			}

		case "grn":
			var grn models.GoodsReceivedNote
			if err := db.Where("id = ? AND organization_id = ?", d.ID, orgID).First(&grn).Error; err == nil {
				for _, att := range parseDocMetadataAttachments(grn.Metadata, "grn", grn.ID, grn.DocumentNumber) {
					addUnique(att)
				}
			}

		case "payment_voucher":
			var pv models.PaymentVoucher
			if err := db.Where("id = ? AND organization_id = ?", d.ID, orgID).First(&pv).Error; err == nil {
				// PV's own uploaded attachments (metadata.attachments) — the
				// live PV detail page supports these independently of proof
				// of payment. See pv-detail-client.tsx.
				for _, att := range parseDocMetadataAttachments(pv.Metadata, "payment_voucher", pv.ID, pv.DocumentNumber) {
					addUnique(att)
				}

				// Proof of payment — NEVER unmarshal the base64 "dataBase64"
				// field into a response-bound struct (see chainPOPMeta).
				if len(pv.ProofOfPayment) > 0 {
					var pop chainPOPMeta
					if err := json.Unmarshal(pv.ProofOfPayment, &pop); err == nil && (pop.ID != "" || pop.FileName != "") {
						addUnique(ChainAttachment{
							Kind:            "proof_of_payment",
							SourceDocType:   "payment_voucher",
							SourceDocID:     pv.ID,
							SourceDocNumber: pv.DocumentNumber,
							FileID:          pop.ID,
							FileName:        pop.FileName,
							FileSize:        pop.SizeBytes,
							MimeType:        pop.MimeType,
							UploadedAt:      pop.UploadedAt,
							UploadedBy:      pop.UploadedBy,
							DownloadRef:     "/payment-vouchers/" + pv.ID,
						})
					}
				}
			}
		}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"attachments": attachments,
			"documents":   docs,
		},
	})
}
