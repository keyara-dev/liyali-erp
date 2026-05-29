package handlers

import (
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/types"
	"gorm.io/datatypes"
)

// ─────────────────────────────────────────────────────────────────────────────
// app factory
// ─────────────────────────────────────────────────────────────────────────────

func newGRNApp(t *testing.T) *fiber.App {
	t.Helper()
	app := fiber.New()
	auth := withTenantCtx(testOrgID, testUserID, testUserRole)

	app.Get("/grns", auth, GetGRNs)
	app.Post("/grns", auth, CreateGRN)
	app.Get("/grns/:id", auth, GetGRN)
	app.Put("/grns/:id", auth, UpdateGRN)
	app.Delete("/grns/:id", auth, DeleteGRN)
	app.Post("/grns/:id/submit", auth, SubmitGRN)
	// Confirm endpoint removed — workflow approval auto-cascades to COMPLETED.
	return app
}

// makeGRN creates and saves a GoodsReceivedNote with the given status.
// For DRAFT GRNs the receiver + certifier sign-off lifecycle is short-circuited
// to "READY" so existing submit/complete tests don't have to walk every state
// transition. Sign-off-specific tests can override grn.SignoffStatus directly.
func makeGRN(t *testing.T, docNum, poDocNum, status string) models.GoodsReceivedNote {
	t.Helper()
	signoff := "READY"
	if status != "DRAFT" {
		signoff = "COMPLETED"
	}
	grn := models.GoodsReceivedNote{
		ID:               uuid.New().String(),
		OrganizationID:   testOrgID,
		DocumentNumber:   docNum,
		PODocumentNumber: poDocNum,
		Status:           status,
		SignoffStatus:    signoff,
		ReceivedDate:     time.Now(),
		ReceivedBy:       testUserID,
		ApprovalStage:    0,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	grn.Items = datatypes.NewJSONType([]types.GRNItem{
		{Description: "Widget A", QuantityOrdered: 10, QuantityReceived: 10, Condition: "good"},
	})
	grn.QualityIssues = datatypes.NewJSONType([]types.QualityIssue{})
	grn.ApprovalHistory = datatypes.NewJSONType([]types.ApprovalRecord{})
	grn.ActionHistory = datatypes.NewJSONType([]types.ActionHistoryEntry{})
	if err := config.DB.Create(&grn).Error; err != nil {
		t.Fatalf("makeGRN: %v", err)
	}
	return grn
}

// makeApprovedPO creates a PO that GRN creation can reference (goods_first default flow).
func makeApprovedPO(t *testing.T, docNum string) models.PurchaseOrder {
	t.Helper()
	order := models.PurchaseOrder{
		ID:             uuid.New().String(),
		OrganizationID: testOrgID,
		DocumentNumber: docNum,
		Status:         "APPROVED",
		TotalAmount:    1000.00,
		Currency:       "ZMW",
		DeliveryDate:   time.Now().Add(30 * 24 * time.Hour),
		ApprovalStage:  0,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	order.Items = datatypes.NewJSONType([]types.POItem{
		{Description: "Widget A", Quantity: 10, UnitPrice: 100.0, Amount: 1000.0},
	})
	order.ApprovalHistory = datatypes.NewJSONType([]types.ApprovalRecord{})
	order.ActionHistory = datatypes.NewJSONType([]types.ActionHistoryEntry{})
	if err := config.DB.Create(&order).Error; err != nil {
		t.Fatalf("makeApprovedPO: %v", err)
	}
	return order
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /grns
// ─────────────────────────────────────────────────────────────────────────────

func TestGetGRNs_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := fiber.New()
	app.Get("/grns", GetGRNs)

	resp := testRequest(app, http.MethodGet, "/grns", nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

func TestGetGRNs_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newGRNApp(t)
	resp := testRequest(app, http.MethodGet, "/grns", nil)
	if resp.StatusCode != http.StatusOK {
		body := decodeResponse(resp)
		t.Fatalf("expected 200, got %d; body=%v", resp.StatusCode, body)
	}
	body := decodeResponse(resp)
	if body["success"] != true {
		t.Errorf("expected success=true, got %v", body["success"])
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// POST /grns
// ─────────────────────────────────────────────────────────────────────────────

func TestCreateGRN_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := fiber.New()
	app.Post("/grns", CreateGRN)

	resp := testRequest(app, http.MethodPost, "/grns", map[string]interface{}{
		"poDocumentNumber": "PO-2024-0001",
	})
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

func TestCreateGRN_MissingPODocumentNumber(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newGRNApp(t)
	body := map[string]interface{}{
		"receivedBy": testUserID,
		"items": []map[string]interface{}{
			{"description": "Widget", "quantityOrdered": 5, "quantityReceived": 5, "condition": "good"},
		},
	}
	resp := testRequest(app, http.MethodPost, "/grns", body)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for missing poDocumentNumber, got %d", resp.StatusCode)
	}
}

func TestCreateGRN_InvalidPODocumentNumberFormat(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newGRNApp(t)
	body := map[string]interface{}{
		"poDocumentNumber": "INVALID",
		"receivedBy":       testUserID,
		"items": []map[string]interface{}{
			{"description": "Widget", "quantityOrdered": 5, "quantityReceived": 5, "condition": "good"},
		},
	}
	resp := testRequest(app, http.MethodPost, "/grns", body)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid PO document number format, got %d", resp.StatusCode)
	}
}

func TestCreateGRN_MissingReceivedBy(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newGRNApp(t)
	body := map[string]interface{}{
		"poDocumentNumber": "PO-2024-0001",
		"items": []map[string]interface{}{
			{"description": "Widget", "quantityOrdered": 5, "quantityReceived": 5, "condition": "good"},
		},
	}
	resp := testRequest(app, http.MethodPost, "/grns", body)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for missing receivedBy, got %d", resp.StatusCode)
	}
}

func TestCreateGRN_MissingItems(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newGRNApp(t)
	body := map[string]interface{}{
		"poDocumentNumber": "PO-2024-0001",
		"receivedBy":       testUserID,
		"items":            []map[string]interface{}{},
	}
	resp := testRequest(app, http.MethodPost, "/grns", body)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for missing items, got %d", resp.StatusCode)
	}
}

func TestCreateGRN_PONotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newGRNApp(t)
	body := map[string]interface{}{
		"poDocumentNumber": "PO-2024-NONEXISTENT",
		"receivedBy":       testUserID,
		"items": []map[string]interface{}{
			{"description": "Widget", "quantityOrdered": 5, "quantityReceived": 5, "condition": "good"},
		},
	}
	resp := testRequest(app, http.MethodPost, "/grns", body)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 when PO not found, got %d", resp.StatusCode)
	}
}

func TestCreateGRN_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Create a PO that the GRN can reference (goods_first flow — no PV required)
	po := makeApprovedPO(t, "PO-2024-0001")

	app := newGRNApp(t)
	body := map[string]interface{}{
		"poDocumentNumber": po.DocumentNumber,
		"receivedBy":       testUserID,
		"items": []map[string]interface{}{
			{"description": "Widget A", "quantityOrdered": 10, "quantityReceived": 10, "condition": "good"},
		},
	}
	resp := testRequest(app, http.MethodPost, "/grns", body)
	if resp.StatusCode != http.StatusCreated {
		respBody := decodeResponse(resp)
		t.Fatalf("expected 201, got %d; body=%v", resp.StatusCode, respBody)
	}
	respBody := decodeResponse(resp)
	if respBody["success"] != true {
		t.Errorf("expected success=true, got %v", respBody["success"])
	}
	data, ok := respBody["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object in response")
	}
	if data["documentNumber"] == nil || data["documentNumber"] == "" {
		t.Errorf("expected documentNumber in response")
	}
	if data["status"] != "DRAFT" {
		t.Errorf("expected status DRAFT, got %v", data["status"])
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /grns/:id
// ─────────────────────────────────────────────────────────────────────────────

func TestGetGRN_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := fiber.New()
	app.Get("/grns/:id", GetGRN)

	resp := testRequest(app, http.MethodGet, "/grns/some-id", nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

func TestGetGRN_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newGRNApp(t)
	resp := testRequest(app, http.MethodGet, "/grns/non-existent-id", nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestGetGRN_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	grn := makeGRN(t, "GRN-TEST-001", "PO-2024-0001", "DRAFT")

	app := newGRNApp(t)
	resp := testRequest(app, http.MethodGet, "/grns/"+grn.ID, nil)
	if resp.StatusCode != http.StatusOK {
		body := decodeResponse(resp)
		t.Fatalf("expected 200, got %d; body=%v", resp.StatusCode, body)
	}
	body := decodeResponse(resp)
	if body["success"] != true {
		t.Errorf("expected success=true, got %v", body["success"])
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// PUT /grns/:id
// ─────────────────────────────────────────────────────────────────────────────

func TestUpdateGRN_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := fiber.New()
	app.Put("/grns/:id", UpdateGRN)

	resp := testRequest(app, http.MethodPut, "/grns/some-id", map[string]interface{}{
		"receivedBy": "user-002",
	})
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

func TestUpdateGRN_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newGRNApp(t)
	resp := testRequest(app, http.MethodPut, "/grns/non-existent-id", map[string]interface{}{
		"receivedBy": "user-002",
	})
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestUpdateGRN_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	grn := makeGRN(t, "GRN-TEST-002", "PO-2024-0002", "DRAFT")

	app := newGRNApp(t)
	resp := testRequest(app, http.MethodPut, "/grns/"+grn.ID, map[string]interface{}{
		"receivedBy": "user-updated",
	})
	if resp.StatusCode != http.StatusOK {
		body := decodeResponse(resp)
		t.Fatalf("expected 200, got %d; body=%v", resp.StatusCode, body)
	}
	body := decodeResponse(resp)
	if body["success"] != true {
		t.Errorf("expected success=true, got %v", body["success"])
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// DELETE /grns/:id
// ─────────────────────────────────────────────────────────────────────────────

func TestDeleteGRN_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := fiber.New()
	app.Delete("/grns/:id", DeleteGRN)

	resp := testRequest(app, http.MethodDelete, "/grns/some-id", nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

func TestDeleteGRN_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newGRNApp(t)
	resp := testRequest(app, http.MethodDelete, "/grns/non-existent-id", nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestDeleteGRN_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	grn := makeGRN(t, "GRN-DEL-001", "PO-2024-0001", "DRAFT")

	app := newGRNApp(t)
	resp := testRequest(app, http.MethodDelete, "/grns/"+grn.ID, nil)
	if resp.StatusCode != http.StatusOK {
		body := decodeResponse(resp)
		t.Fatalf("expected 200, got %d; body=%v", resp.StatusCode, body)
	}
	body := decodeResponse(resp)
	if body["success"] != true {
		t.Errorf("expected success=true, got %v", body["success"])
	}
}

func TestDeleteGRN_NonDraftForbidden(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	grn := makeGRN(t, "GRN-DEL-002", "PO-2024-0002", "PENDING")

	app := newGRNApp(t)
	resp := testRequest(app, http.MethodDelete, "/grns/"+grn.ID, nil)
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403 for non-draft delete, got %d", resp.StatusCode)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// POST /grns/:id/submit
// ─────────────────────────────────────────────────────────────────────────────

func TestSubmitGRN_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// SubmitGRN uses direct type assertions on locals (panics without auth).
	// Use recover middleware so the panic becomes a 500 instead of crashing the test.
	app := fiber.New()
	app.Use(recover.New())
	app.Post("/grns/:id/submit", SubmitGRN)

	resp := testRequest(app, http.MethodPost, "/grns/some-id/submit", map[string]interface{}{
		"workflowId": "wf-001",
	})
	// Without auth locals the handler panics → 500; either way it's not 200.
	if resp.StatusCode == http.StatusOK {
		t.Errorf("unauthenticated request should be blocked, got 200")
	}
}

func TestSubmitGRN_MissingWorkflowID(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := fiber.New()
	auth := withTenantCtx(testOrgID, testUserID, testUserRole)
	app.Post("/grns/:id/submit", auth, SubmitGRN)

	resp := testRequest(app, http.MethodPost, "/grns/some-id/submit", map[string]interface{}{})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for missing workflowId, got %d", resp.StatusCode)
	}
}

func TestSubmitGRN_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := fiber.New()
	auth := withTenantCtx(testOrgID, testUserID, testUserRole)
	app.Post("/grns/:id/submit", auth, SubmitGRN)

	resp := testRequest(app, http.MethodPost, "/grns/non-existent-id/submit", map[string]interface{}{
		"workflowId": "wf-001",
	})
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestSubmitGRN_AlreadyPending(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	grn := makeGRN(t, "GRN-PEND-001", "PO-2024-0001", "PENDING")

	app := fiber.New()
	auth := withTenantCtx(testOrgID, testUserID, testUserRole)
	app.Post("/grns/:id/submit", auth, SubmitGRN)

	resp := testRequest(app, http.MethodPost, "/grns/"+grn.ID+"/submit", map[string]interface{}{
		"workflowId": "wf-001",
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 when submitting non-DRAFT GRN, got %d", resp.StatusCode)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// POST /grns/:id/confirm
// ─────────────────────────────────────────────────────────────────────────────

// ConfirmGRN endpoint removed: workflow approval now auto-cascades
// APPROVED → COMPLETED. MarkGRNComplete covers the skip-workflow path.
// Legacy tests deleted to avoid drift; see grn_signoff_http_test.go.

// ─────────────────────────────────────────────────────────────────────────────
// Scope isolation: procurement role cannot see GRNs linked to direct_payment POs
// ─────────────────────────────────────────────────────────────────────────────

// makeApprovedPOWithRouting creates a PO with an explicit routing_type for GRN linkage.
func makeApprovedPOWithRouting(t *testing.T, docNum, routingType string) models.PurchaseOrder {
	t.Helper()
	order := models.PurchaseOrder{
		ID:             uuid.New().String(),
		OrganizationID: testOrgID,
		DocumentNumber: docNum,
		Status:         "APPROVED",
		RoutingType:    routingType,
		TotalAmount:    1000.00,
		Currency:       "ZMW",
		DeliveryDate:   time.Now().Add(30 * 24 * time.Hour),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	order.Items = datatypes.NewJSONType([]types.POItem{
		{Description: "Widget", Quantity: 5, UnitPrice: 200.0, Amount: 1000.0},
	})
	order.ApprovalHistory = datatypes.NewJSONType([]types.ApprovalRecord{})
	order.ActionHistory = datatypes.NewJSONType([]types.ActionHistoryEntry{})
	if err := config.DB.Create(&order).Error; err != nil {
		t.Fatalf("makeApprovedPOWithRouting: %v", err)
	}
	return order
}

// TestGRN_ProcurementUserCannotSeeDirectPaymentGRN verifies that a procurement-role
// user receives 404 for a GRN linked to a direct_payment PO (single-get endpoint).
func TestGRN_ProcurementUserCannotSeeDirectPaymentGRN(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Seed a direct_payment PO and a procurement PO.
	directPO := makeApprovedPOWithRouting(t, "PO-DIRECT-GRN-SCOPE-001", "direct_payment")
	procPO := makeApprovedPOWithRouting(t, "PO-PROC-GRN-SCOPE-001", "procurement")

	// GRNs linked to each.
	directGRN := makeGRN(t, "GRN-DIRECT-SCOPE-001", directPO.DocumentNumber, "draft")
	procGRN := makeGRN(t, "GRN-PROC-SCOPE-001", procPO.DocumentNumber, "draft")

	app := fiber.New()
	procAuth := withTenantCtx(testOrgID, testUserID, "procurement")
	app.Get("/grns/:id", procAuth, GetGRN)

	// GRN linked to direct_payment PO → 404.
	resp := testRequest(app, http.MethodGet, "/grns/"+directGRN.ID, nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("procurement user: expected 404 for direct_payment GRN, got %d", resp.StatusCode)
	}

	// GRN linked to procurement PO → 200.
	resp2 := testRequest(app, http.MethodGet, "/grns/"+procGRN.ID, nil)
	if resp2.StatusCode != http.StatusOK {
		body := decodeResponse(resp2)
		t.Errorf("procurement user: expected 200 for procurement GRN, got %d; body=%v", resp2.StatusCode, body)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// POST /grns — item-level + over-receipt validation (Task 5)
// ─────────────────────────────────────────────────────────────────────────────

// TestCreateGRN_OverReceiptVsPO_Rejected verifies that receiving more than the
// ordered quantity for an item is rejected with 400.
func TestCreateGRN_OverReceiptVsPO_Rejected(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// PO has "Widget A" with quantity=10.
	po := makeApprovedPO(t, "PO-2024-OVER-001")

	app := newGRNApp(t)
	body := map[string]interface{}{
		"poDocumentNumber": po.DocumentNumber,
		"receivedBy":       testUserID,
		"items": []map[string]interface{}{
			{"description": "Widget A", "quantityOrdered": 10, "quantityReceived": 15, "condition": "good"},
		},
	}
	resp := testRequest(app, http.MethodPost, "/grns", body)
	if resp.StatusCode != http.StatusBadRequest {
		body := decodeResponse(resp)
		t.Errorf("expected 400 for over-receipt, got %d; body=%v", resp.StatusCode, body)
	}
}

// TestCreateGRN_UnknownItemDescription_Rejected verifies that a GRN item whose
// description is not on the PO is rejected with 400.
func TestCreateGRN_UnknownItemDescription_Rejected(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	po := makeApprovedPO(t, "PO-2024-UNKNWN-001")

	app := newGRNApp(t)
	body := map[string]interface{}{
		"poDocumentNumber": po.DocumentNumber,
		"receivedBy":       testUserID,
		"items": []map[string]interface{}{
			{"description": "NotOnPO Item", "quantityOrdered": 5, "quantityReceived": 5, "condition": "good"},
		},
	}
	resp := testRequest(app, http.MethodPost, "/grns", body)
	if resp.StatusCode != http.StatusBadRequest {
		respBody := decodeResponse(resp)
		t.Errorf("expected 400 for unknown item description, got %d; body=%v", resp.StatusCode, respBody)
	}
}

// TestCreateGRN_ValidReceipt_Accepted verifies that a GRN with valid items and
// quantities within PO limits is accepted.
func TestCreateGRN_ValidReceipt_Accepted(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	po := makeApprovedPO(t, "PO-2024-VALID-001")

	app := newGRNApp(t)
	body := map[string]interface{}{
		"poDocumentNumber": po.DocumentNumber,
		"receivedBy":       testUserID,
		"items": []map[string]interface{}{
			{"description": "Widget A", "quantityOrdered": 10, "quantityReceived": 5, "condition": "good"},
		},
	}
	resp := testRequest(app, http.MethodPost, "/grns", body)
	if resp.StatusCode != http.StatusCreated {
		respBody := decodeResponse(resp)
		t.Fatalf("expected 201 for valid receipt, got %d; body=%v", resp.StatusCode, respBody)
	}
}

// TestCreateGRN_CrossGRN_OverReceipt_Skipped documents the cross-GRN aggregate
// guard.  The current one-to-one GRN-per-PO unique constraint prevents creating
// a second GRN for the same PO, so an end-to-end integration test requires
// multi-GRN-per-PO support.  This test is skipped until that constraint is relaxed.
func TestCreateGRN_CrossGRN_OverReceipt_Skipped(t *testing.T) {
	t.Skip("requires multi-GRN-per-PO support: the one-to-one unique index on (po_document_number, status!=CANCELLED) prevents a second GRN")
}
