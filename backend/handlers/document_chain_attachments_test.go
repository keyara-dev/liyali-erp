package handlers

// document_chain_attachments_test.go — Task C1: chain-wide supporting-document
// aggregation endpoint (GET /document-chain/:id/attachments). Seeds a full
// REQ -> PO -> GRN -> PV(x2) chain with attachments/quotations at every level
// (including a PO attachment copied from the REQ with the same fileId, tagged
// fromRequisition) plus one PV carrying a proof-of-payment blob, and asserts:
//   - dedupe collapses the shared fileId to a single entry (REQ wins)
//   - all four document types contribute at least one attachment
//   - the proof-of-payment entry has a downloadRef and NEVER serializes the
//     base64 blob
//   - both PVs (multi-PV / partial payments) are represented
// anchored from each of the four document types.

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/models"
	"gorm.io/datatypes"
)

func documentChainAttachmentsApp() *fiber.App {
	app := fiber.New()
	auth := withTenantCtx(testOrgID, testUserID, testUserRole)
	app.Get("/document-chain/:id/attachments", auth, GetDocumentChainAttachments)
	return app
}

func mustJSON(t *testing.T, v interface{}) datatypes.JSON {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal metadata: %v", err)
	}
	return datatypes.JSON(b)
}

// popBlobMarker is a distinctive base64-ish string that must NEVER appear
// anywhere in a serialized /attachments response.
const popBlobMarker = "VEhJU19JU19TRUNSRVRfUE9QX0JBU0U2NF9CTE9C"

func TestGetDocumentChainAttachments(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	now := time.Now()

	// --- Requisition: one attachment (shared fileId, later copied to the PO)
	// plus one quotation. ---
	reqMetadata := mustJSON(t, map[string]interface{}{
		"attachments": []map[string]interface{}{
			{
				"fileId":     "file-shared-1",
				"fileName":   "req-support.pdf",
				"fileUrl":    "https://files.example.com/req-support.pdf",
				"fileSize":   float64(1024),
				"mimeType":   "application/pdf",
				"uploadedAt": "2026-07-01T10:00:00Z",
			},
		},
		"quotations": []map[string]interface{}{
			{
				"fileId":     "file-quote-1",
				"fileName":   "vendor-quote.pdf",
				"fileUrl":    "https://files.example.com/vendor-quote.pdf",
				"fileSize":   float64(2048),
				"uploadedAt": "2026-07-01T09:00:00Z",
			},
		},
	})
	req := models.Requisition{
		ID: uuid.New().String(), OrganizationID: testOrgID, DocumentNumber: "REQ-CA-1",
		Title: "Chain attachments fixture", Status: "APPROVED",
		Metadata:  reqMetadata,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := db.Create(&req).Error; err != nil {
		t.Fatalf("seed REQ: %v", err)
	}

	// --- Purchase Order: the REQ's attachment copied over (SAME fileId,
	// tagged fromRequisition) + one attachment of its own. ---
	poMetadata := mustJSON(t, map[string]interface{}{
		"attachments": []map[string]interface{}{
			{
				"fileId":          "file-shared-1",
				"fileName":        "req-support.pdf",
				"fileUrl":         "https://files.example.com/req-support.pdf",
				"fileSize":        float64(1024),
				"mimeType":        "application/pdf",
				"uploadedAt":      "2026-07-01T10:00:00Z",
				"fromRequisition": true,
			},
			{
				"fileId":     "file-po-own-1",
				"fileName":   "po-own.pdf",
				"fileUrl":    "https://files.example.com/po-own.pdf",
				"fileSize":   float64(512),
				"mimeType":   "application/pdf",
				"uploadedAt": "2026-07-02T10:00:00Z",
			},
		},
	})
	srcReq := req.ID
	po := models.PurchaseOrder{
		ID: uuid.New().String(), OrganizationID: testOrgID, DocumentNumber: "PO-CA-1",
		Status: "APPROVED", SourceRequisitionId: &srcReq, LinkedRequisition: req.DocumentNumber,
		TotalAmount: 1000, Currency: "ZMW",
		Metadata:  poMetadata,
		CreatedAt: now.Add(time.Minute), UpdatedAt: now.Add(time.Minute),
	}
	if err := db.Create(&po).Error; err != nil {
		t.Fatalf("seed PO: %v", err)
	}

	// --- GRN: one attachment of its own. ---
	grnMetadata := mustJSON(t, map[string]interface{}{
		"attachments": []map[string]interface{}{
			{
				"fileId":     "file-grn-1",
				"fileName":   "delivery-note.jpg",
				"fileUrl":    "https://files.example.com/delivery-note.jpg",
				"fileSize":   float64(4096),
				"mimeType":   "image/jpeg",
				"uploadedAt": "2026-07-03T10:00:00Z",
			},
		},
	})
	grn := models.GoodsReceivedNote{
		ID: uuid.New().String(), OrganizationID: testOrgID, DocumentNumber: "GRN-CA-1",
		PODocumentNumber: po.DocumentNumber, Status: "COMPLETED",
		Metadata:  grnMetadata,
		CreatedAt: now.Add(2 * time.Minute), UpdatedAt: now.Add(2 * time.Minute),
	}
	if err := db.Create(&grn).Error; err != nil {
		t.Fatalf("seed GRN: %v", err)
	}

	// --- PV #1 (paid, with proof of payment carrying a base64 blob that
	// must never leak into the response). ---
	popPayload := mustJSON(t, map[string]interface{}{
		"id":         "pop-1",
		"fileName":   "proof-of-payment.png",
		"mimeType":   "image/png",
		"sizeBytes":  float64(777),
		"dataBase64": popBlobMarker,
		"uploadedAt": "2026-07-04T10:00:00Z",
		"uploadedBy": "user-finance-1",
	})
	pv1 := models.PaymentVoucher{
		ID: uuid.New().String(), OrganizationID: testOrgID, DocumentNumber: "PV-CA-1",
		LinkedPO: po.DocumentNumber, Status: "PAID", Amount: 400, Currency: "ZMW",
		ProofOfPayment: popPayload,
		CreatedAt:      now.Add(3 * time.Minute), UpdatedAt: now.Add(3 * time.Minute),
	}
	if err := db.Create(&pv1).Error; err != nil {
		t.Fatalf("seed PV1: %v", err)
	}

	// --- PV #2 (second installment — partial payment, no POP yet). ---
	pv2 := models.PaymentVoucher{
		ID: uuid.New().String(), OrganizationID: testOrgID, DocumentNumber: "PV-CA-2",
		LinkedPO: po.DocumentNumber, Status: "APPROVED", Amount: 600, Currency: "ZMW",
		CreatedAt: now.Add(4 * time.Minute), UpdatedAt: now.Add(4 * time.Minute),
	}
	if err := db.Create(&pv2).Error; err != nil {
		t.Fatalf("seed PV2: %v", err)
	}

	type attachmentDTO struct {
		Kind            string `json:"kind"`
		SourceDocType   string `json:"sourceDocType"`
		SourceDocID     string `json:"sourceDocId"`
		FileID          string `json:"fileId"`
		FileName        string `json:"fileName"`
		FromRequisition bool   `json:"fromRequisition"`
		Category        string `json:"category"`
		DownloadRef     string `json:"downloadRef"`
	}
	type docDTO struct {
		ID   string `json:"id"`
		Type string `json:"type"`
	}
	type respBody struct {
		Success bool `json:"success"`
		Data    struct {
			Attachments []attachmentDTO `json:"attachments"`
			Documents   []docDTO        `json:"documents"`
		} `json:"data"`
	}

	anchors := []struct {
		name string
		id   string
		typ  string
	}{
		{"requisition", req.ID, "requisition"},
		{"purchase_order", po.ID, "purchase_order"},
		{"grn", grn.ID, "grn"},
		{"payment_voucher_1", pv1.ID, "payment_voucher"},
	}

	for _, anchor := range anchors {
		t.Run(anchor.name, func(t *testing.T) {
			resp := testRequest(documentChainAttachmentsApp(), http.MethodGet,
				"/document-chain/"+anchor.id+"/attachments?documentType="+anchor.typ, nil)
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("expected 200, got %d", resp.StatusCode)
			}

			rawBody, err := readAllAndDecode(resp)
			if err != nil {
				t.Fatalf("read body: %v", err)
			}

			// The base64 blob (or the literal key it's stored under) must
			// never appear anywhere in the serialized response.
			if strings.Contains(rawBody, popBlobMarker) {
				t.Fatalf("proof-of-payment base64 blob leaked into response body: %s", rawBody)
			}
			if strings.Contains(rawBody, "dataBase64") {
				t.Fatalf("dataBase64 key leaked into response body: %s", rawBody)
			}

			var body respBody
			if err := json.Unmarshal([]byte(rawBody), &body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if !body.Success {
				t.Fatalf("expected success=true, got body=%s", rawBody)
			}

			// --- All 4 document sources represented ---
			docTypeCounts := map[string]int{}
			for _, d := range body.Data.Documents {
				docTypeCounts[d.Type]++
			}
			if docTypeCounts["requisition"] != 1 {
				t.Errorf("expected exactly 1 requisition in documents, got %d (%+v)", docTypeCounts["requisition"], body.Data.Documents)
			}
			if docTypeCounts["purchase_order"] != 1 {
				t.Errorf("expected exactly 1 purchase_order in documents, got %d", docTypeCounts["purchase_order"])
			}
			if docTypeCounts["grn"] != 1 {
				t.Errorf("expected exactly 1 grn in documents, got %d", docTypeCounts["grn"])
			}
			// --- Multi-PV: both PVs must appear ---
			if docTypeCounts["payment_voucher"] != 2 {
				t.Errorf("expected exactly 2 payment_vouchers in documents (multi-PV), got %d", docTypeCounts["payment_voucher"])
			}

			byFileID := map[string][]attachmentDTO{}
			for _, a := range body.Data.Attachments {
				if a.FileID != "" {
					byFileID[a.FileID] = append(byFileID[a.FileID], a)
				}
			}

			// --- Dedupe: the shared fileId appears exactly once, sourced
			// from the requisition (first occurrence in REQ->PO->GRN->PV
			// order wins) even though the PO also carries a copy. ---
			shared := byFileID["file-shared-1"]
			if len(shared) != 1 {
				t.Fatalf("expected exactly 1 deduped entry for file-shared-1, got %d: %+v", len(shared), shared)
			}
			if shared[0].SourceDocType != "requisition" {
				t.Errorf("expected shared fileId to resolve to the requisition (first occurrence wins), got sourceDocType=%q", shared[0].SourceDocType)
			}

			// --- Quotation surfaced with category "quotation" ---
			quote := byFileID["file-quote-1"]
			if len(quote) != 1 {
				t.Fatalf("expected exactly 1 quotation entry, got %d", len(quote))
			}
			if quote[0].Kind != "quotation" || quote[0].Category != "quotation" {
				t.Errorf("expected quotation kind/category, got kind=%q category=%q", quote[0].Kind, quote[0].Category)
			}

			// --- PO's own attachment present ---
			if len(byFileID["file-po-own-1"]) != 1 {
				t.Errorf("expected PO's own attachment to be present, got %+v", byFileID["file-po-own-1"])
			}

			// --- GRN's own attachment present ---
			if len(byFileID["file-grn-1"]) != 1 {
				t.Errorf("expected GRN's own attachment to be present, got %+v", byFileID["file-grn-1"])
			}

			// --- Proof of payment: downloadRef set, sourced from PV1, no base64 ---
			var popEntry *attachmentDTO
			for i := range body.Data.Attachments {
				if body.Data.Attachments[i].Kind == "proof_of_payment" {
					popEntry = &body.Data.Attachments[i]
					break
				}
			}
			if popEntry == nil {
				t.Fatalf("expected a proof_of_payment attachment entry, got %+v", body.Data.Attachments)
			}
			if popEntry.SourceDocID != pv1.ID {
				t.Errorf("expected POP sourceDocId=%s, got %s", pv1.ID, popEntry.SourceDocID)
			}
			if popEntry.DownloadRef != "/payment-vouchers/"+pv1.ID {
				t.Errorf("expected downloadRef=/payment-vouchers/%s, got %q", pv1.ID, popEntry.DownloadRef)
			}
			if popEntry.FileName != "proof-of-payment.png" {
				t.Errorf("expected POP fileName, got %q", popEntry.FileName)
			}
		})
	}
}

// Fix 4 regression: an unlinked PV (linked_po=="", e.g. a direct/payment-first
// PV never tied to a PO) has no PO to resolve a chain from, so
// resolveChainDocumentSet used to return zero refs — dropping the anchor PV
// itself and hiding its own metadata.attachments/proof-of-payment. The anchor
// must always be present in the resolved set even with no chain around it.
func TestGetDocumentChainAttachments_UnlinkedPV_ReturnsOwnAttachments(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	now := time.Now()

	pvMetadata := mustJSON(t, map[string]interface{}{
		"attachments": []map[string]interface{}{
			{
				"fileId":     "file-unlinked-pv-1",
				"fileName":   "invoice.pdf",
				"fileUrl":    "https://files.example.com/invoice.pdf",
				"fileSize":   float64(2048),
				"mimeType":   "application/pdf",
				"uploadedAt": "2026-07-05T10:00:00Z",
			},
		},
	})
	popPayload := mustJSON(t, map[string]interface{}{
		"id":         "pop-unlinked-1",
		"fileName":   "unlinked-pop.png",
		"mimeType":   "image/png",
		"sizeBytes":  float64(333),
		"dataBase64": popBlobMarker,
		"uploadedAt": "2026-07-05T11:00:00Z",
		"uploadedBy": "user-finance-2",
	})

	pv := models.PaymentVoucher{
		ID: uuid.New().String(), OrganizationID: testOrgID, DocumentNumber: "PV-UNLINKED-1",
		LinkedPO: "", Status: "PAID", Amount: 250, Currency: "ZMW",
		Metadata:       pvMetadata,
		ProofOfPayment: popPayload,
		CreatedAt:      now, UpdatedAt: now,
	}
	if err := db.Create(&pv).Error; err != nil {
		t.Fatalf("seed unlinked PV: %v", err)
	}

	resp := testRequest(documentChainAttachmentsApp(), http.MethodGet,
		"/document-chain/"+pv.ID+"/attachments?documentType=payment_voucher", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	rawBody, err := readAllAndDecode(resp)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if strings.Contains(rawBody, popBlobMarker) {
		t.Fatalf("proof-of-payment base64 blob leaked into response body: %s", rawBody)
	}

	type attachmentDTO struct {
		Kind        string `json:"kind"`
		SourceDocID string `json:"sourceDocId"`
		FileID      string `json:"fileId"`
		DownloadRef string `json:"downloadRef"`
	}
	type docDTO struct {
		ID   string `json:"id"`
		Type string `json:"type"`
	}
	type respBody struct {
		Success bool `json:"success"`
		Data    struct {
			Attachments []attachmentDTO `json:"attachments"`
			Documents   []docDTO        `json:"documents"`
		} `json:"data"`
	}
	var body respBody
	if err := json.Unmarshal([]byte(rawBody), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if !body.Success {
		t.Fatalf("expected success=true, got body=%s", rawBody)
	}

	if len(body.Data.Documents) != 1 || body.Data.Documents[0].ID != pv.ID || body.Data.Documents[0].Type != "payment_voucher" {
		t.Fatalf("expected the unlinked PV itself as the sole document, got %+v", body.Data.Documents)
	}

	var attachmentEntry, popEntry *attachmentDTO
	for i := range body.Data.Attachments {
		switch body.Data.Attachments[i].Kind {
		case "attachment":
			attachmentEntry = &body.Data.Attachments[i]
		case "proof_of_payment":
			popEntry = &body.Data.Attachments[i]
		}
	}
	if attachmentEntry == nil || attachmentEntry.FileID != "file-unlinked-pv-1" {
		t.Fatalf("expected the unlinked PV's own attachment to be present, got %+v", body.Data.Attachments)
	}
	if popEntry == nil || popEntry.SourceDocID != pv.ID || popEntry.DownloadRef != "/payment-vouchers/"+pv.ID {
		t.Fatalf("expected the unlinked PV's own proof-of-payment to be present, got %+v", body.Data.Attachments)
	}
}

// readAllAndDecode reads the full response body as a string (once) so tests
// can both scan the raw text (for base64-leak assertions) and JSON-decode it.
func readAllAndDecode(resp *http.Response) (string, error) {
	defer resp.Body.Close()
	buf := make([]byte, 0, 8192)
	chunk := make([]byte, 4096)
	for {
		n, err := resp.Body.Read(chunk)
		if n > 0 {
			buf = append(buf, chunk[:n]...)
		}
		if err != nil {
			break
		}
	}
	return string(buf), nil
}
