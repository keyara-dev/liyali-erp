package handlers

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	fiberrecover "github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/types"
	"gorm.io/datatypes"
)

// ─────────────────────────────────────────────────────────────────────────────
// app factory
// ─────────────────────────────────────────────────────────────────────────────

func newPaymentVoucherApp(t *testing.T) *fiber.App {
	t.Helper()
	app := fiber.New()
	auth := withTenantCtx(testOrgID, testUserID, testUserRole)

	app.Get("/payment-vouchers", auth, GetPaymentVouchers)
	app.Get("/payment-vouchers/stats", auth, GetPaymentVoucherStats)
	app.Post("/payment-vouchers", auth, CreatePaymentVoucher)
	app.Post("/payment-vouchers/from-po", auth, CreatePaymentVoucherFromPO)
	app.Post("/payment-vouchers/recover-from-po/:poId", auth, RecoverPVFromPO)
	app.Get("/payment-vouchers/:id", auth, GetPaymentVoucher)
	app.Put("/payment-vouchers/:id", auth, UpdatePaymentVoucher)
	app.Delete("/payment-vouchers/:id", auth, DeletePaymentVoucher)
	app.Post("/payment-vouchers/:id/submit", auth, SubmitPaymentVoucher)
	app.Post("/payment-vouchers/:id/mark-paid", auth, MarkPaymentVoucherPaid)
	app.Post("/payment-vouchers/:id/mark-paid-with-pop", auth, MarkPaidWithPOP)
	return app
}

// makePaymentVoucher creates and saves a PaymentVoucher with the given status.
func makePaymentVoucher(t *testing.T, docNum, status string) models.PaymentVoucher {
	t.Helper()
	voucher := models.PaymentVoucher{
		ID:             uuid.New().String(),
		OrganizationID: testOrgID,
		DocumentNumber: docNum,
		InvoiceNumber:  "INV-" + uuid.New().String()[:8],
		Status:         status,
		Amount:         1000.00,
		Currency:       "ZMW",
		PaymentMethod:  "bank_transfer",
		GLCode:         "GL-001",
		Description:    "Test payment voucher description",
		ApprovalStage:  0,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	voucher.ApprovalHistory = datatypes.NewJSONType([]types.ApprovalRecord{})
	voucher.ActionHistory = datatypes.NewJSONType([]types.ActionHistoryEntry{})
	if err := config.DB.Create(&voucher).Error; err != nil {
		t.Fatalf("makePaymentVoucher: %v", err)
	}
	return voucher
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /payment-vouchers
// ─────────────────────────────────────────────────────────────────────────────

func TestGetPaymentVouchers_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := fiber.New()
	app.Get("/payment-vouchers", GetPaymentVouchers)

	resp := testRequest(app, http.MethodGet, "/payment-vouchers", nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

func TestGetPaymentVouchers_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newPaymentVoucherApp(t)
	resp := testRequest(app, http.MethodGet, "/payment-vouchers", nil)
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
// GET /payment-vouchers/stats
// ─────────────────────────────────────────────────────────────────────────────

func TestGetPaymentVoucherStats_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := fiber.New()
	app.Get("/payment-vouchers/stats", GetPaymentVoucherStats)

	resp := testRequest(app, http.MethodGet, "/payment-vouchers/stats", nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

func TestGetPaymentVoucherStats_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newPaymentVoucherApp(t)
	resp := testRequest(app, http.MethodGet, "/payment-vouchers/stats", nil)
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
// POST /payment-vouchers
// ─────────────────────────────────────────────────────────────────────────────

func TestCreatePaymentVoucher_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := fiber.New()
	app.Post("/payment-vouchers", CreatePaymentVoucher)

	resp := testRequest(app, http.MethodPost, "/payment-vouchers", map[string]interface{}{
		"invoiceNumber": "INV-001",
	})
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

func TestCreatePaymentVoucher_MissingInvoiceNumber(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newPaymentVoucherApp(t)
	body := map[string]interface{}{
		"amount":        500.0,
		"currency":      "ZMW",
		"paymentMethod": "bank_transfer",
		"glCode":        "GL-001",
		"description":   "Payment for office supplies order",
	}
	resp := testRequest(app, http.MethodPost, "/payment-vouchers", body)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for missing invoiceNumber, got %d", resp.StatusCode)
	}
}

func TestCreatePaymentVoucher_ZeroAmount(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newPaymentVoucherApp(t)
	body := map[string]interface{}{
		"invoiceNumber": "INV-001",
		"amount":        0,
		"currency":      "ZMW",
		"paymentMethod": "bank_transfer",
		"glCode":        "GL-001",
		"description":   "Payment for office supplies order",
	}
	resp := testRequest(app, http.MethodPost, "/payment-vouchers", body)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for zero amount, got %d", resp.StatusCode)
	}
}

func TestCreatePaymentVoucher_MissingDescription(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newPaymentVoucherApp(t)
	body := map[string]interface{}{
		"invoiceNumber": "INV-001",
		"amount":        500.0,
		"currency":      "ZMW",
		"paymentMethod": "bank_transfer",
		"glCode":        "GL-001",
	}
	resp := testRequest(app, http.MethodPost, "/payment-vouchers", body)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for missing description, got %d", resp.StatusCode)
	}
}

func TestCreatePaymentVoucher_DescriptionTooShort(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newPaymentVoucherApp(t)
	body := map[string]interface{}{
		"invoiceNumber": "INV-001",
		"amount":        500.0,
		"currency":      "ZMW",
		"paymentMethod": "bank_transfer",
		"glCode":        "GL-001",
		"description":   "Short",
	}
	resp := testRequest(app, http.MethodPost, "/payment-vouchers", body)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for description too short, got %d", resp.StatusCode)
	}
}

func TestCreatePaymentVoucher_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newPaymentVoucherApp(t)
	body := map[string]interface{}{
		"invoiceNumber": "INV-2024-001",
		"amount":        1500.0,
		"currency":      "ZMW",
		"paymentMethod": "bank_transfer",
		"glCode":        "GL-5001",
		"description":   "Payment for January office supplies purchase",
	}
	resp := testRequest(app, http.MethodPost, "/payment-vouchers", body)
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
// POST /payment-vouchers/from-po
// ─────────────────────────────────────────────────────────────────────────────

func TestCreatePaymentVoucherFromPO_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := fiber.New()
	app.Post("/payment-vouchers/from-po", CreatePaymentVoucherFromPO)

	resp := testRequest(app, http.MethodPost, "/payment-vouchers/from-po", map[string]interface{}{
		"poDocumentNumber": "PO-001",
	})
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /payment-vouchers/:id
// ─────────────────────────────────────────────────────────────────────────────

func TestGetPaymentVoucher_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := fiber.New()
	app.Get("/payment-vouchers/:id", GetPaymentVoucher)

	resp := testRequest(app, http.MethodGet, "/payment-vouchers/some-id", nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

func TestGetPaymentVoucher_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newPaymentVoucherApp(t)
	resp := testRequest(app, http.MethodGet, "/payment-vouchers/non-existent-id", nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestGetPaymentVoucher_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	voucher := makePaymentVoucher(t, "PV-TEST-001", "DRAFT")

	app := newPaymentVoucherApp(t)
	resp := testRequest(app, http.MethodGet, "/payment-vouchers/"+voucher.ID, nil)
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
// PUT /payment-vouchers/:id
// ─────────────────────────────────────────────────────────────────────────────

func TestUpdatePaymentVoucher_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := fiber.New()
	app.Put("/payment-vouchers/:id", UpdatePaymentVoucher)

	resp := testRequest(app, http.MethodPut, "/payment-vouchers/some-id", map[string]interface{}{
		"amount": 200.0,
	})
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

func TestUpdatePaymentVoucher_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newPaymentVoucherApp(t)
	resp := testRequest(app, http.MethodPut, "/payment-vouchers/non-existent-id", map[string]interface{}{
		"amount": 200.0,
	})
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestUpdatePaymentVoucher_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	voucher := makePaymentVoucher(t, "PV-TEST-002", "DRAFT")

	app := newPaymentVoucherApp(t)
	resp := testRequest(app, http.MethodPut, "/payment-vouchers/"+voucher.ID, map[string]interface{}{
		"amount":   2000.0,
		"currency": "USD",
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
// DELETE /payment-vouchers/:id
// ─────────────────────────────────────────────────────────────────────────────

func TestDeletePaymentVoucher_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := fiber.New()
	app.Delete("/payment-vouchers/:id", DeletePaymentVoucher)

	resp := testRequest(app, http.MethodDelete, "/payment-vouchers/some-id", nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

func TestDeletePaymentVoucher_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newPaymentVoucherApp(t)
	resp := testRequest(app, http.MethodDelete, "/payment-vouchers/non-existent-id", nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestDeletePaymentVoucher_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	voucher := makePaymentVoucher(t, "PV-DEL-001", "DRAFT")

	app := newPaymentVoucherApp(t)
	resp := testRequest(app, http.MethodDelete, "/payment-vouchers/"+voucher.ID, nil)
	if resp.StatusCode != http.StatusOK {
		body := decodeResponse(resp)
		t.Fatalf("expected 200, got %d; body=%v", resp.StatusCode, body)
	}
	body := decodeResponse(resp)
	if body["success"] != true {
		t.Errorf("expected success=true, got %v", body["success"])
	}
}

func TestDeletePaymentVoucher_NonDraftForbidden(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	voucher := makePaymentVoucher(t, "PV-DEL-002", "PENDING")

	app := newPaymentVoucherApp(t)
	resp := testRequest(app, http.MethodDelete, "/payment-vouchers/"+voucher.ID, nil)
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403 for non-draft delete, got %d", resp.StatusCode)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// POST /payment-vouchers/:id/submit
// ─────────────────────────────────────────────────────────────────────────────

func TestSubmitPaymentVoucher_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// SubmitPaymentVoucher uses direct type assertions on locals; panics without auth.
	// recover.New() converts the panic to a 500 so the goroutine doesn't crash.
	app := fiber.New()
	app.Use(fiberrecover.New())
	app.Post("/payment-vouchers/:id/submit", SubmitPaymentVoucher)

	resp := testRequest(app, http.MethodPost, "/payment-vouchers/some-id/submit", map[string]interface{}{
		"workflowId": "wf-001",
	})
	if resp.StatusCode == http.StatusOK {
		t.Errorf("unauthenticated request should be blocked, got 200")
	}
}

func TestSubmitPaymentVoucher_MissingWorkflowID(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := fiber.New()
	auth := withTenantCtx(testOrgID, testUserID, testUserRole)
	app.Post("/payment-vouchers/:id/submit", auth, SubmitPaymentVoucher)

	resp := testRequest(app, http.MethodPost, "/payment-vouchers/some-id/submit", map[string]interface{}{})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for missing workflowId, got %d", resp.StatusCode)
	}
}

func TestSubmitPaymentVoucher_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := fiber.New()
	auth := withTenantCtx(testOrgID, testUserID, testUserRole)
	app.Post("/payment-vouchers/:id/submit", auth, SubmitPaymentVoucher)

	resp := testRequest(app, http.MethodPost, "/payment-vouchers/non-existent-id/submit", map[string]interface{}{
		"workflowId": "wf-001",
	})
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestSubmitPaymentVoucher_InvalidStatus(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// A PENDING voucher is not DRAFT and cannot be submitted again
	voucher := makePaymentVoucher(t, "PV-PEND-001", "PENDING")

	app := fiber.New()
	auth := withTenantCtx(testOrgID, testUserID, testUserRole)
	app.Post("/payment-vouchers/:id/submit", auth, SubmitPaymentVoucher)

	resp := testRequest(app, http.MethodPost, "/payment-vouchers/"+voucher.ID+"/submit", map[string]interface{}{
		"workflowId": "wf-001",
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 when submitting non-DRAFT voucher, got %d", resp.StatusCode)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// POST /payment-vouchers/:id/mark-paid
// ─────────────────────────────────────────────────────────────────────────────

func TestMarkPaymentVoucherPaid_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := fiber.New()
	app.Post("/payment-vouchers/:id/mark-paid", MarkPaymentVoucherPaid)

	resp := testRequest(app, http.MethodPost, "/payment-vouchers/some-id/mark-paid", map[string]interface{}{
		"paidAmount": 100.0,
	})
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

func TestMarkPaymentVoucherPaid_NotFound(t *testing.T) {
	// mark-paid now requires a pending workflow execution task; seeding that
	// in the SQLite harness is impractical. Run against Postgres for coverage.
	t.Skip("requires Postgres test DB + seeded workflow execution task")
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newPaymentVoucherApp(t)
	resp := testRequest(app, http.MethodPost, "/payment-vouchers/non-existent-id/mark-paid", map[string]interface{}{
		"paidAmount": 1000.0,
		"signature":  "data:image/png;base64,sig",
	})
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestMarkPaymentVoucherPaid_ZeroPaidAmount(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	voucher := makePaymentVoucher(t, "PV-PAID-001", "APPROVED")

	app := newPaymentVoucherApp(t)
	resp := testRequest(app, http.MethodPost, "/payment-vouchers/"+voucher.ID+"/mark-paid", map[string]interface{}{
		"paidAmount": 0,
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for zero paidAmount, got %d", resp.StatusCode)
	}
}

func TestMarkPaymentVoucherPaid_NotApproved(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// DRAFT voucher cannot be marked paid — must be APPROVED
	voucher := makePaymentVoucher(t, "PV-PAID-002", "DRAFT")

	app := newPaymentVoucherApp(t)
	resp := testRequest(app, http.MethodPost, "/payment-vouchers/"+voucher.ID+"/mark-paid", map[string]interface{}{
		"paidAmount": 1000.0,
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 when marking non-APPROVED voucher as paid, got %d", resp.StatusCode)
	}
}

func TestMarkPaymentVoucherPaid_Success(t *testing.T) {
	t.Skip("requires Postgres test DB + seeded workflow execution task")
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	voucher := makePaymentVoucher(t, "PV-PAID-003", "APPROVED")

	// Seed the test user so action history lookup works
	seedTestUser(t)

	app := newPaymentVoucherApp(t)
	resp := testRequest(app, http.MethodPost, "/payment-vouchers/"+voucher.ID+"/mark-paid", map[string]interface{}{
		"paidAmount":      1000.0,
		"referenceNumber": "REF-001",
		"comments":        "Payment processed successfully",
		"signature":       "data:image/png;base64,sig",
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
// Manual vendor name persistence (bug regression)
// ─────────────────────────────────────────────────────────────────────────────

func TestPaymentVoucher_PersistManualVendorName(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newPaymentVoucherApp(t)

	createBody := map[string]interface{}{
		"vendorId":      "",
		"vendorName":    "BLUE FOX FARMS LIMITED",
		"invoiceNumber": "INV-2026-001",
		"amount":        3547200.0,
		"currency":      "ZMW",
		"paymentMethod": "bank_transfer",
		"glCode":        "5100",
		"description":   "Payment for laptop procurement quotation",
	}
	resp := testRequest(app, http.MethodPost, "/payment-vouchers", createBody)
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		t.Fatalf("create PV: expected 200/201, got %d", resp.StatusCode)
	}

	var createResp struct {
		Data types.PaymentVoucherResponse `json:"data"`
	}
	decodeJSON(t, resp, &createResp)

	if createResp.Data.VendorName != "BLUE FOX FARMS LIMITED" {
		t.Errorf("create response: expected vendorName=BLUE FOX FARMS LIMITED, got %q", createResp.Data.VendorName)
	}

	pvID := createResp.Data.ID
	getResp := testRequest(app, http.MethodGet, "/payment-vouchers/"+pvID, nil)
	if getResp.StatusCode != http.StatusOK {
		t.Fatalf("get PV: expected 200, got %d", getResp.StatusCode)
	}

	var getBody struct {
		Data types.PaymentVoucherResponse `json:"data"`
	}
	decodeJSON(t, getResp, &getBody)

	if getBody.Data.VendorName != "BLUE FOX FARMS LIMITED" {
		t.Errorf("get response: expected vendorName=BLUE FOX FARMS LIMITED, got %q", getBody.Data.VendorName)
	}

	updateBody := map[string]interface{}{
		"vendorId":   "",
		"vendorName": "MICOP BUSINESS VENTURES",
	}
	updResp := testRequest(app, http.MethodPut, "/payment-vouchers/"+pvID, updateBody)
	if updResp.StatusCode != http.StatusOK {
		t.Fatalf("update PV: expected 200, got %d", updResp.StatusCode)
	}
	var updRespBody struct {
		Data types.PaymentVoucherResponse `json:"data"`
	}
	decodeJSON(t, updResp, &updRespBody)
	if updRespBody.Data.VendorName != "MICOP BUSINESS VENTURES" {
		t.Errorf("update response: expected vendorName=MICOP BUSINESS VENTURES, got %q", updRespBody.Data.VendorName)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// ─────────────────────────────────────────────────────────────────────────────
// Multipart helpers for mark-paid-with-pop tests
// ─────────────────────────────────────────────────────────────────────────────

// multipartFile builds a multipart/form-data body with a single file field.
// Returns (body *bytes.Buffer, contentType string).
func multipartFile(t *testing.T, fieldName, fileName, mimeType string, content []byte) (*bytes.Buffer, string) {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile(fieldName, fileName)
	if err != nil {
		t.Fatalf("multipartFile CreateFormFile: %v", err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatalf("multipartFile Write: %v", err)
	}
	// Override Content-Type header for the part to the desired mimeType.
	// Note: CreateFormFile sets application/octet-stream; the handler reads
	// the part header, but for tests we just trust the file extension / content.
	_ = mimeType // handler sniffs via extension
	if err := w.Close(); err != nil {
		t.Fatalf("multipartFile Close: %v", err)
	}
	return &buf, w.FormDataContentType()
}

// appendField adds a plain text field to an existing multipart body.
// Returns new *bytes.Buffer — the original is consumed.
func appendField(existing *bytes.Buffer, fieldName, value string) *bytes.Buffer {
	// We need to re-open the existing boundary; easiest is a fresh writer
	// that appends fields then re-closes. Since the caller chains:
	//   body, ct := multipartFile(...)
	//   body = appendField(body, ...)
	// we parse the boundary from the existing data and re-write everything.
	// Simpler: just build a brand-new form with both file + field.
	// This helper is only called ONCE, so re-build is fine.
	_ = existing // unused; re-build handled by the test via testMultipartWithField
	return existing // no-op placeholder — tests use testMultipartWithField instead
}

// testMultipartWithField builds a multipart form with a file + extra string fields.
// fields is a map[fieldName]value for text fields.
func testMultipartWithField(t *testing.T, fileField, fileName, mimeType string, content []byte, fields map[string]string) (*bytes.Buffer, string) {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	// text fields first
	for k, v := range fields {
		if err := w.WriteField(k, v); err != nil {
			t.Fatalf("testMultipartWithField WriteField %s: %v", k, err)
		}
	}
	// file field
	_ = mimeType
	part, err := w.CreateFormFile(fileField, fileName)
	if err != nil {
		t.Fatalf("testMultipartWithField CreateFormFile: %v", err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatalf("testMultipartWithField Write: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("testMultipartWithField Close: %v", err)
	}
	return &buf, w.FormDataContentType()
}

// testMultipartNoFile builds a multipart form with only text fields (no file).
func testMultipartNoFile(t *testing.T, fields map[string]string) (*bytes.Buffer, string) {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	for k, v := range fields {
		if err := w.WriteField(k, v); err != nil {
			t.Fatalf("testMultipartNoFile WriteField %s: %v", k, err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("testMultipartNoFile Close: %v", err)
	}
	return &buf, w.FormDataContentType()
}

// multipartRequest builds a *http.Request with a multipart body.
func multipartRequest(method, path string, body *bytes.Buffer, contentType string) *http.Request {
	req := httptest.NewRequest(method, path, body)
	req.Header.Set("Content-Type", contentType)
	return req
}

// makePVWithRoutingAndLinkedPO creates a PV linked to a PO with a specific routing type.
func makePVWithRoutingAndLinkedPO(t *testing.T, docNum, status, routingType, linkedPO string) models.PaymentVoucher {
	t.Helper()
	voucher := models.PaymentVoucher{
		ID:             uuid.New().String(),
		OrganizationID: testOrgID,
		DocumentNumber: docNum,
		InvoiceNumber:  "INV-" + uuid.New().String()[:8],
		Status:         status,
		RoutingType:    routingType,
		LinkedPO:       linkedPO,
		Amount:         1000.00,
		Currency:       "ZMW",
		PaymentMethod:  "bank_transfer",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	voucher.ApprovalHistory = datatypes.NewJSONType([]types.ApprovalRecord{})
	voucher.ActionHistory = datatypes.NewJSONType([]types.ActionHistoryEntry{})
	if err := config.DB.Create(&voucher).Error; err != nil {
		t.Fatalf("makePVWithRoutingAndLinkedPO: %v", err)
	}
	return voucher
}

// makePOWithRouting creates a PurchaseOrder with a specific routing_type.
func makePOWithRouting(t *testing.T, docNum, status, routingType string) models.PurchaseOrder {
	t.Helper()
	order := models.PurchaseOrder{
		ID:             uuid.New().String(),
		OrganizationID: testOrgID,
		DocumentNumber: docNum,
		Status:         status,
		RoutingType:    routingType,
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
		t.Fatalf("makePOWithRouting: %v", err)
	}
	return order
}

// ─────────────────────────────────────────────────────────────────────────────
// POST /payment-vouchers/:id/mark-paid-with-pop (Phase 5)
// ─────────────────────────────────────────────────────────────────────────────

func TestMarkPaid_RequiresApprovedStatus(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	pv := makePaymentVoucherWithRouting(t, "PV-MKP-DRAFT-001", models.StatusDraft, models.RoutingTypeDirectPayment)

	body, ct := testMultipartWithField(t, "popFile", "slip.pdf", "application/pdf", []byte("%PDF-1.4"), map[string]string{
		"paidDate": "2026-05-11",
	})
	req := multipartRequest(http.MethodPost, "/payment-vouchers/"+pv.ID+"/mark-paid-with-pop", body, ct)

	app := newPaymentVoucherApp(t)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status=%d want 409 (must be approved first)", resp.StatusCode)
	}
}

func TestMarkPaid_RequiresPOPFile(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	pv := makePaymentVoucherWithRouting(t, "PV-MKP-NOPOP-001", models.StatusApproved, models.RoutingTypeDirectPayment)

	body, ct := testMultipartNoFile(t, map[string]string{"paidDate": "2026-05-11"})
	req := multipartRequest(http.MethodPost, "/payment-vouchers/"+pv.ID+"/mark-paid-with-pop", body, ct)

	app := newPaymentVoucherApp(t)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status=%d want 400 (no POP file)", resp.StatusCode)
	}
}

func TestMarkPaid_HappyPath(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	pv := makePaymentVoucherWithRouting(t, "PV-MKP-OK-001", models.StatusApproved, models.RoutingTypeDirectPayment)

	body, ct := testMultipartWithField(t, "popFile", "slip.pdf", "application/pdf", []byte("%PDF-1.4"), map[string]string{
		"paidDate": "2026-05-11",
	})
	req := multipartRequest(http.MethodPost, "/payment-vouchers/"+pv.ID+"/mark-paid-with-pop", body, ct)

	app := newPaymentVoucherApp(t)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		body2 := decodeResponse(resp)
		t.Fatalf("status=%d want 200; body=%v", resp.StatusCode, body2)
	}
	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)
	data, _ := result["data"].(map[string]any)
	if data == nil {
		t.Fatal("response has no data field")
	}
	st, _ := data["status"].(string)
	if st != models.StatusPaid {
		t.Fatalf("status=%q want %q", st, models.StatusPaid)
	}
	if data["proofOfPayment"] == nil {
		t.Fatal("proofOfPayment missing from response")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// POST /payment-vouchers/recover-from-po/:poId (Phase 5)
// ─────────────────────────────────────────────────────────────────────────────

func TestRecoverFromPO_CreatesDraftPV(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	po := makePOWithRouting(t, "PO-REC-001", models.StatusApproved, models.RoutingTypeDirectPayment)

	req := httptest.NewRequest(http.MethodPost, "/payment-vouchers/recover-from-po/"+po.ID, nil)
	app := newPaymentVoucherApp(t)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		body := decodeResponse(resp)
		t.Fatalf("status=%d want 201; body=%v", resp.StatusCode, body)
	}
}

func TestRecoverFromPO_Idempotent(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	po := makePOWithRouting(t, "PO-REC-IDEM-001", models.StatusApproved, models.RoutingTypeDirectPayment)

	app := newPaymentVoucherApp(t)

	req1 := httptest.NewRequest(http.MethodPost, "/payment-vouchers/recover-from-po/"+po.ID, nil)
	resp1, err := app.Test(req1, -1)
	if err != nil {
		t.Fatalf("first call app.Test: %v", err)
	}
	resp1.Body.Close()
	if resp1.StatusCode != http.StatusCreated {
		t.Fatalf("first call status=%d want 201", resp1.StatusCode)
	}

	req2 := httptest.NewRequest(http.MethodPost, "/payment-vouchers/recover-from-po/"+po.ID, nil)
	resp2, err := app.Test(req2, -1)
	if err != nil {
		t.Fatalf("second call app.Test: %v", err)
	}
	if resp2.StatusCode != http.StatusOK {
		body := decodeResponse(resp2)
		t.Fatalf("idempotent call status=%d want 200; body=%v", resp2.StatusCode, body)
	}
}

func TestRecoverFromPO_RejectsNonDirectPayment(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	po := makePOWithRouting(t, "PO-REC-PROC-001", models.StatusApproved, models.RoutingTypeProcurement)

	req := httptest.NewRequest(http.MethodPost, "/payment-vouchers/recover-from-po/"+po.ID, nil)
	app := newPaymentVoucherApp(t)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		body := decodeResponse(resp)
		t.Fatalf("status=%d want 400 for non-direct_payment PO; body=%v", resp.StatusCode, body)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Scope isolation: procurement role cannot see direct_payment PVs
// ─────────────────────────────────────────────────────────────────────────────

// makePaymentVoucherWithRouting creates a PaymentVoucher with an explicit routing_type.
func makePaymentVoucherWithRouting(t *testing.T, docNum, status, routingType string) models.PaymentVoucher {
	t.Helper()
	voucher := models.PaymentVoucher{
		ID:             uuid.New().String(),
		OrganizationID: testOrgID,
		DocumentNumber: docNum,
		InvoiceNumber:  "INV-" + uuid.New().String()[:8],
		Status:         status,
		RoutingType:    routingType,
		Amount:         500.00,
		Currency:       "ZMW",
		PaymentMethod:  "bank_transfer",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	voucher.ApprovalHistory = datatypes.NewJSONType([]types.ApprovalRecord{})
	voucher.ActionHistory = datatypes.NewJSONType([]types.ActionHistoryEntry{})
	if err := config.DB.Create(&voucher).Error; err != nil {
		t.Fatalf("makePaymentVoucherWithRouting: %v", err)
	}
	return voucher
}

// TestPV_ProcurementUserCannotSeeDirectPayment verifies that a procurement-role
// user receives 404 when fetching a direct_payment PV by ID (single-get endpoint).
func TestPV_ProcurementUserCannotSeeDirectPayment(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	directPV := makePaymentVoucherWithRouting(t, "PV-DIRECT-SCOPE-001", "draft", "direct_payment")
	procPV := makePaymentVoucherWithRouting(t, "PV-PROC-SCOPE-001", "draft", "procurement")

	app := fiber.New()
	procAuth := withTenantCtx(testOrgID, testUserID, "procurement")
	app.Get("/payment-vouchers/:id", procAuth, GetPaymentVoucher)

	// direct_payment PV → 404 (invisible, no info leak).
	resp := testRequest(app, http.MethodGet, "/payment-vouchers/"+directPV.ID, nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("procurement user: expected 404 for direct_payment PV, got %d", resp.StatusCode)
	}

	// procurement PV → 200.
	resp2 := testRequest(app, http.MethodGet, "/payment-vouchers/"+procPV.ID, nil)
	if resp2.StatusCode != http.StatusOK {
		body := decodeResponse(resp2)
		t.Errorf("procurement user: expected 200 for procurement PV, got %d; body=%v", resp2.StatusCode, body)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// POST /payment-vouchers/:id/submit — scope gate (Task 4)
// ─────────────────────────────────────────────────────────────────────────────

// makePaymentVoucherWithCreator creates a draft PV owned by the given userID.
func makePaymentVoucherWithCreator(t *testing.T, docNum, createdBy string) models.PaymentVoucher {
	t.Helper()
	voucher := models.PaymentVoucher{
		ID:             uuid.New().String(),
		OrganizationID: testOrgID,
		DocumentNumber: docNum,
		InvoiceNumber:  "INV-" + uuid.New().String()[:8],
		Status:         "DRAFT",
		Amount:         500.00,
		Currency:       "ZMW",
		PaymentMethod:  "bank_transfer",
		GLCode:         "GL-001",
		Description:    "Scope gate test PV",
		CreatedBy:      createdBy,
		ApprovalStage:  0,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	voucher.ApprovalHistory = datatypes.NewJSONType([]types.ApprovalRecord{})
	voucher.ActionHistory = datatypes.NewJSONType([]types.ActionHistoryEntry{})
	if err := config.DB.Create(&voucher).Error; err != nil {
		t.Fatalf("makePaymentVoucherWithCreator: %v", err)
	}
	return voucher
}

// TestSubmitPaymentVoucher_NonOwnerNonPrivileged_Forbidden verifies that a
// requester-role user who is not the PV owner and has no workflow task on it
// receives 403 when attempting to submit.
func TestSubmitPaymentVoucher_NonOwnerNonPrivileged_Forbidden(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupWorkflowTasksTable(t, db)

	ownerID := "owner-user-001"
	otherID := "other-user-002"

	// PV created by ownerID; we will authenticate as otherID (requester role).
	voucher := makePaymentVoucherWithCreator(t, "PV-SCOPE-NOTOWN-001", ownerID)

	// Use fiberrecover so panics (e.g. missing workflowExecutionService) become 500,
	// not a test crash.
	app := fiber.New()
	app.Use(fiberrecover.New())
	otherAuth := withTenantCtx(testOrgID, otherID, "requester")
	app.Post("/payment-vouchers/:id/submit", otherAuth, SubmitPaymentVoucher)

	resp := testRequest(app, http.MethodPost, "/payment-vouchers/"+voucher.ID+"/submit", map[string]interface{}{
		"workflowId": uuid.New().String(),
	})
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403 for non-owner non-privileged requester, got %d", resp.StatusCode)
	}
}

// TestSubmitPaymentVoucher_Owner_Allowed verifies that the PV creator can submit
// their own PV (regression guard: direct-payment auto-PV flow must still work).
func TestSubmitPaymentVoucher_Owner_Allowed(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupWorkflowTasksTable(t, db)

	ownerID := "owner-user-001"
	voucher := makePaymentVoucherWithCreator(t, "PV-SCOPE-OWNER-001", ownerID)

	// Use fiberrecover so the handler can panic at workflowExecutionService (no service
	// injected in this unit test) — the important thing is it must NOT return 403.
	app := fiber.New()
	app.Use(fiberrecover.New())
	ownerAuth := withTenantCtx(testOrgID, ownerID, "requester")
	app.Post("/payment-vouchers/:id/submit", ownerAuth, SubmitPaymentVoucher)

	resp := testRequest(app, http.MethodPost, "/payment-vouchers/"+voucher.ID+"/submit", map[string]interface{}{
		"workflowId": uuid.New().String(),
	})
	if resp.StatusCode == http.StatusForbidden {
		t.Errorf("owner must not be blocked by scope gate, got 403")
	}
}
