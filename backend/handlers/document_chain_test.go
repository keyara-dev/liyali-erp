package handlers

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/models"
)

func documentChainApp() *fiber.App {
	app := fiber.New()
	auth := withTenantCtx(testOrgID, testUserID, testUserRole)
	app.Get("/document-chain/:id", auth, GetDocumentChain)
	return app
}

// Regression: the document-chain endpoint passed the requested document's own id
// straight into a requisition-rooted walker, so requesting the chain for a
// purchase order looked the PO id up as a requisition → "record not found" → 500.
// It must resolve the chain from the document's direct link fields instead, and
// return 200 with the related requisition (parent) + GRN/PV (children).
func TestGetDocumentChain_PurchaseOrder_DoesNotErrorAndResolvesChain(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	req := models.Requisition{
		ID: uuid.New().String(), OrganizationID: testOrgID, DocumentNumber: "REQ-DC-1",
		Title: "Office chairs", Status: "APPROVED",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	if err := db.Create(&req).Error; err != nil {
		t.Fatalf("seed REQ: %v", err)
	}

	srcReq := req.ID
	po := models.PurchaseOrder{
		ID: uuid.New().String(), OrganizationID: testOrgID, DocumentNumber: "PO-DC-1",
		Status: "APPROVED", SourceRequisitionId: &srcReq, LinkedRequisition: req.DocumentNumber,
		TotalAmount: 1000, Currency: "ZMW",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	if err := db.Create(&po).Error; err != nil {
		t.Fatalf("seed PO: %v", err)
	}

	grn := models.GoodsReceivedNote{
		ID: uuid.New().String(), OrganizationID: testOrgID, DocumentNumber: "GRN-DC-1",
		PODocumentNumber: po.DocumentNumber, Status: "COMPLETED",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	if err := db.Create(&grn).Error; err != nil {
		t.Fatalf("seed GRN: %v", err)
	}

	pv := models.PaymentVoucher{
		ID: uuid.New().String(), OrganizationID: testOrgID, DocumentNumber: "PV-DC-1",
		LinkedPO: po.DocumentNumber, Status: "APPROVED", Amount: 1000, Currency: "ZMW",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	if err := db.Create(&pv).Error; err != nil {
		t.Fatalf("seed PV: %v", err)
	}

	resp := testRequest(documentChainApp(), http.MethodGet,
		"/document-chain/"+po.ID+"?documentType=purchase_order", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for PO document chain, got %d", resp.StatusCode)
	}

	var body struct {
		Success bool `json:"success"`
		Data    struct {
			ParentDocuments []map[string]interface{} `json:"parentDocuments"`
			ChildDocuments  []map[string]interface{} `json:"childDocuments"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}

	hasType := func(docs []map[string]interface{}, typ string) bool {
		for _, d := range docs {
			if t2, ok := d["type"].(string); ok && t2 == typ {
				return true
			}
		}
		return false
	}

	if !hasType(body.Data.ParentDocuments, "requisition") {
		t.Errorf("expected source requisition as a parent document, got %+v", body.Data.ParentDocuments)
	}
	if !hasType(body.Data.ChildDocuments, "grn") {
		t.Errorf("expected GRN as a child document, got %+v", body.Data.ChildDocuments)
	}
	if !hasType(body.Data.ChildDocuments, "payment_voucher") {
		t.Errorf("expected PV as a child document, got %+v", body.Data.ChildDocuments)
	}
}

// Partial payments (Task B) mean a PO can have MULTIPLE payment vouchers.
// buildDocumentChain's PV child lookup uses Find (not First), so ALL PVs
// linked to the PO must appear in childDocuments, ordered created_at ASC.
// buildDocumentChain applies NO status filter at this call site, so a
// CANCELLED PV is included too — this test pins that real behavior rather
// than an aspiration (the cancelled-exclusion filter lives only in the
// /attachments resolver, resolveChainDocumentSet).
func TestGetDocumentChain_PurchaseOrder_ReturnsAllLinkedPVs(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	base := time.Now()

	po := models.PurchaseOrder{
		ID: uuid.New().String(), OrganizationID: testOrgID, DocumentNumber: "PO-MPV-1",
		Status: "APPROVED", TotalAmount: 1000, Currency: "ZMW",
		CreatedAt: base, UpdatedAt: base,
	}
	if err := db.Create(&po).Error; err != nil {
		t.Fatalf("seed PO: %v", err)
	}

	// Three PVs linked to the same PO: two live installments plus one
	// cancelled. Distinct created_at values so we can assert ordering.
	pv1 := models.PaymentVoucher{
		ID: uuid.New().String(), OrganizationID: testOrgID, DocumentNumber: "PV-MPV-1",
		LinkedPO: po.DocumentNumber, Status: "PAID", Amount: 400, Currency: "ZMW",
		CreatedAt: base.Add(1 * time.Minute), UpdatedAt: base.Add(1 * time.Minute),
	}
	pv2 := models.PaymentVoucher{
		ID: uuid.New().String(), OrganizationID: testOrgID, DocumentNumber: "PV-MPV-2",
		LinkedPO: po.DocumentNumber, Status: "APPROVED", Amount: 600, Currency: "ZMW",
		CreatedAt: base.Add(2 * time.Minute), UpdatedAt: base.Add(2 * time.Minute),
	}
	pvCancelled := models.PaymentVoucher{
		ID: uuid.New().String(), OrganizationID: testOrgID, DocumentNumber: "PV-MPV-3-CANCELLED",
		LinkedPO: po.DocumentNumber, Status: "CANCELLED", Amount: 100, Currency: "ZMW",
		CreatedAt: base.Add(3 * time.Minute), UpdatedAt: base.Add(3 * time.Minute),
	}
	for _, pv := range []*models.PaymentVoucher{&pv1, &pv2, &pvCancelled} {
		if err := db.Create(pv).Error; err != nil {
			t.Fatalf("seed PV %s: %v", pv.DocumentNumber, err)
		}
	}

	resp := testRequest(documentChainApp(), http.MethodGet,
		"/document-chain/"+po.ID+"?documentType=purchase_order", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for PO document chain, got %d", resp.StatusCode)
	}

	var body struct {
		Success bool `json:"success"`
		Data    struct {
			ChildDocuments []map[string]interface{} `json:"childDocuments"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}

	// Collect the PV document numbers in the order they appear.
	pvDocNumbers := []string{}
	for _, d := range body.Data.ChildDocuments {
		if typ, _ := d["type"].(string); typ == "payment_voucher" {
			if dn, _ := d["documentNumber"].(string); dn != "" {
				pvDocNumbers = append(pvDocNumbers, dn)
			}
		}
	}

	// Both live PVs must be present (multi-PV: Find, not First).
	has := func(dn string) bool {
		for _, got := range pvDocNumbers {
			if got == dn {
				return true
			}
		}
		return false
	}
	if !has("PV-MPV-1") {
		t.Errorf("expected first live PV PV-MPV-1 in childDocuments, got %v", pvDocNumbers)
	}
	if !has("PV-MPV-2") {
		t.Errorf("expected second live PV PV-MPV-2 in childDocuments, got %v", pvDocNumbers)
	}

	// Real behavior at this call site: no CANCELLED filter, so the cancelled
	// PV is included too. Assert that rather than a filtered aspiration.
	if !has("PV-MPV-3-CANCELLED") {
		t.Errorf("buildDocumentChain applies no status filter, so the cancelled PV should appear; got %v", pvDocNumbers)
	}

	// All three PVs, ordered created_at ASC.
	want := []string{"PV-MPV-1", "PV-MPV-2", "PV-MPV-3-CANCELLED"}
	if len(pvDocNumbers) != len(want) {
		t.Fatalf("expected %d PVs in childDocuments, got %d: %v", len(want), len(pvDocNumbers), pvDocNumbers)
	}
	for i := range want {
		if pvDocNumbers[i] != want[i] {
			t.Fatalf("PV ordering: expected %v (created_at ASC), got %v", want, pvDocNumbers)
		}
	}
}
