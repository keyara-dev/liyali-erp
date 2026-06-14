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
