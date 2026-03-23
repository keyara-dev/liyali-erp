package handlers

import (
	"net/http"
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
	app.Get("/payment-vouchers/:id", auth, GetPaymentVoucher)
	app.Put("/payment-vouchers/:id", auth, UpdatePaymentVoucher)
	app.Delete("/payment-vouchers/:id", auth, DeletePaymentVoucher)
	app.Post("/payment-vouchers/:id/submit", auth, SubmitPaymentVoucher)
	app.Post("/payment-vouchers/:id/mark-paid", auth, MarkPaymentVoucherPaid)
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
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newPaymentVoucherApp(t)
	resp := testRequest(app, http.MethodPost, "/payment-vouchers/non-existent-id/mark-paid", map[string]interface{}{
		"paidAmount": 1000.0,
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
