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

func newPurchaseOrderApp(t *testing.T) *fiber.App {
	t.Helper()
	app := fiber.New()
	auth := withTenantCtx(testOrgID, testUserID, testUserRole)

	app.Get("/purchase-orders", auth, GetPurchaseOrders)
	app.Get("/purchase-orders/stats", auth, GetPurchaseOrderStats)
	app.Post("/purchase-orders", auth, CreatePurchaseOrder)
	app.Post("/purchase-orders/from-requisition", auth, CreatePurchaseOrderFromRequisition)
	app.Get("/purchase-orders/:id", auth, GetPurchaseOrder)
	app.Put("/purchase-orders/:id", auth, UpdatePurchaseOrder)
	app.Delete("/purchase-orders/:id", auth, DeletePurchaseOrder)
	app.Post("/purchase-orders/:id/submit", auth, SubmitPurchaseOrder)
	return app
}

// makePurchaseOrder creates and saves a PurchaseOrder with the given status.
func makePurchaseOrder(t *testing.T, docNum, status string) models.PurchaseOrder {
	t.Helper()
	order := models.PurchaseOrder{
		ID:             uuid.New().String(),
		OrganizationID: testOrgID,
		DocumentNumber: docNum,
		Status:         status,
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
		t.Fatalf("makePurchaseOrder: %v", err)
	}
	return order
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /purchase-orders
// ─────────────────────────────────────────────────────────────────────────────

func TestGetPurchaseOrders_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := fiber.New()
	app.Get("/purchase-orders", GetPurchaseOrders)

	resp := testRequest(app, http.MethodGet, "/purchase-orders", nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

func TestGetPurchaseOrders_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newPurchaseOrderApp(t)
	resp := testRequest(app, http.MethodGet, "/purchase-orders", nil)
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
// GET /purchase-orders/stats
// ─────────────────────────────────────────────────────────────────────────────

func TestGetPurchaseOrderStats_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := fiber.New()
	app.Get("/purchase-orders/stats", GetPurchaseOrderStats)

	resp := testRequest(app, http.MethodGet, "/purchase-orders/stats", nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

func TestGetPurchaseOrderStats_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newPurchaseOrderApp(t)
	resp := testRequest(app, http.MethodGet, "/purchase-orders/stats", nil)
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
// POST /purchase-orders
// ─────────────────────────────────────────────────────────────────────────────

func TestCreatePurchaseOrder_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := fiber.New()
	app.Post("/purchase-orders", CreatePurchaseOrder)

	resp := testRequest(app, http.MethodPost, "/purchase-orders", map[string]interface{}{
		"totalAmount": 100.0,
	})
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

func TestCreatePurchaseOrder_MissingItems(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newPurchaseOrderApp(t)
	body := map[string]interface{}{
		"totalAmount":  100.0,
		"currency":     "ZMW",
		"deliveryDate": time.Now().Add(30 * 24 * time.Hour).Format(time.RFC3339),
		"items":        []map[string]interface{}{},
	}
	resp := testRequest(app, http.MethodPost, "/purchase-orders", body)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for missing items, got %d", resp.StatusCode)
	}
}

func TestCreatePurchaseOrder_ZeroAmount(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newPurchaseOrderApp(t)
	body := map[string]interface{}{
		"totalAmount":  0,
		"currency":     "ZMW",
		"deliveryDate": time.Now().Add(30 * 24 * time.Hour).Format(time.RFC3339),
		"items": []map[string]interface{}{
			{"description": "Widget", "quantity": 1, "unitPrice": 0, "amount": 0},
		},
	}
	resp := testRequest(app, http.MethodPost, "/purchase-orders", body)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for zero amount, got %d", resp.StatusCode)
	}
}

func TestCreatePurchaseOrder_ItemZeroQuantity(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newPurchaseOrderApp(t)
	body := map[string]interface{}{
		"totalAmount":  100.0,
		"currency":     "ZMW",
		"deliveryDate": time.Now().Add(30 * 24 * time.Hour).Format(time.RFC3339),
		"items": []map[string]interface{}{
			{"description": "Widget", "quantity": 0, "unitPrice": 100.0, "amount": 0},
		},
	}
	resp := testRequest(app, http.MethodPost, "/purchase-orders", body)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for zero item quantity, got %d", resp.StatusCode)
	}
}

func TestCreatePurchaseOrder_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newPurchaseOrderApp(t)
	body := map[string]interface{}{
		"totalAmount":  500.0,
		"currency":     "ZMW",
		"deliveryDate": time.Now().Add(30 * 24 * time.Hour).Format(time.RFC3339),
		"items": []map[string]interface{}{
			{"description": "Widget A", "quantity": 5, "unitPrice": 100.0, "amount": 500.0},
		},
	}
	resp := testRequest(app, http.MethodPost, "/purchase-orders", body)
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
// POST /purchase-orders/from-requisition
// ─────────────────────────────────────────────────────────────────────────────

func TestCreatePurchaseOrderFromRequisition_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := fiber.New()
	app.Post("/purchase-orders/from-requisition", CreatePurchaseOrderFromRequisition)

	resp := testRequest(app, http.MethodPost, "/purchase-orders/from-requisition", map[string]interface{}{
		"requisitionId": "req-001",
	})
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

func TestCreatePurchaseOrderFromRequisition_MissingRequisitionID(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newPurchaseOrderApp(t)
	body := map[string]interface{}{
		"totalAmount": 500.0,
		"currency":    "ZMW",
		"items": []map[string]interface{}{
			{"description": "Widget", "quantity": 1, "unitPrice": 500.0, "amount": 500.0},
		},
	}
	resp := testRequest(app, http.MethodPost, "/purchase-orders/from-requisition", body)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for missing requisitionId, got %d", resp.StatusCode)
	}
}

func TestCreatePurchaseOrderFromRequisition_MissingItems(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newPurchaseOrderApp(t)
	body := map[string]interface{}{
		"requisitionId": "req-001",
		"totalAmount":   500.0,
		"currency":      "ZMW",
		"items":         []map[string]interface{}{},
	}
	resp := testRequest(app, http.MethodPost, "/purchase-orders/from-requisition", body)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for missing items, got %d", resp.StatusCode)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /purchase-orders/:id
// ─────────────────────────────────────────────────────────────────────────────

func TestGetPurchaseOrder_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := fiber.New()
	app.Get("/purchase-orders/:id", GetPurchaseOrder)

	resp := testRequest(app, http.MethodGet, "/purchase-orders/some-id", nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

func TestGetPurchaseOrder_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newPurchaseOrderApp(t)
	resp := testRequest(app, http.MethodGet, "/purchase-orders/non-existent-id", nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestGetPurchaseOrder_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	order := makePurchaseOrder(t, "PO-TEST-001", "DRAFT")

	app := newPurchaseOrderApp(t)
	resp := testRequest(app, http.MethodGet, "/purchase-orders/"+order.ID, nil)
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
// PUT /purchase-orders/:id
// ─────────────────────────────────────────────────────────────────────────────

func TestUpdatePurchaseOrder_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := fiber.New()
	app.Use(fiberrecover.New())
	app.Put("/purchase-orders/:id", UpdatePurchaseOrder)

	resp := testRequest(app, http.MethodPut, "/purchase-orders/some-id", map[string]interface{}{
		"totalAmount": 200.0,
	})
	if resp.StatusCode == http.StatusOK {
		t.Errorf("unauthenticated request should be blocked, got 200")
	}
}

func TestUpdatePurchaseOrder_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newPurchaseOrderApp(t)
	resp := testRequest(app, http.MethodPut, "/purchase-orders/non-existent-id", map[string]interface{}{
		"totalAmount": 200.0,
	})
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestUpdatePurchaseOrder_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	order := makePurchaseOrder(t, "PO-TEST-002", "DRAFT")

	app := newPurchaseOrderApp(t)
	resp := testRequest(app, http.MethodPut, "/purchase-orders/"+order.ID, map[string]interface{}{
		"totalAmount": 2000.0,
		"currency":    "USD",
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
// DELETE /purchase-orders/:id
// ─────────────────────────────────────────────────────────────────────────────

func TestDeletePurchaseOrder_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := fiber.New()
	app.Use(fiberrecover.New())
	app.Delete("/purchase-orders/:id", DeletePurchaseOrder)

	resp := testRequest(app, http.MethodDelete, "/purchase-orders/some-id", nil)
	if resp.StatusCode == http.StatusOK {
		t.Errorf("unauthenticated request should be blocked, got 200")
	}
}

func TestDeletePurchaseOrder_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newPurchaseOrderApp(t)
	resp := testRequest(app, http.MethodDelete, "/purchase-orders/non-existent-id", nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestDeletePurchaseOrder_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	order := makePurchaseOrder(t, "PO-DEL-001", "DRAFT")

	app := newPurchaseOrderApp(t)
	resp := testRequest(app, http.MethodDelete, "/purchase-orders/"+order.ID, nil)
	if resp.StatusCode != http.StatusOK {
		body := decodeResponse(resp)
		t.Fatalf("expected 200, got %d; body=%v", resp.StatusCode, body)
	}
	body := decodeResponse(resp)
	if body["success"] != true {
		t.Errorf("expected success=true, got %v", body["success"])
	}
}

func TestDeletePurchaseOrder_NonDraftForbidden(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	order := makePurchaseOrder(t, "PO-DEL-002", "PENDING")

	app := newPurchaseOrderApp(t)
	resp := testRequest(app, http.MethodDelete, "/purchase-orders/"+order.ID, nil)
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403 for non-draft delete, got %d", resp.StatusCode)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// POST /purchase-orders/:id/submit
// ─────────────────────────────────────────────────────────────────────────────

func TestSubmitPurchaseOrder_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := fiber.New()
	app.Use(fiberrecover.New())
	app.Post("/purchase-orders/:id/submit", SubmitPurchaseOrder)

	resp := testRequest(app, http.MethodPost, "/purchase-orders/some-id/submit", map[string]interface{}{
		"workflowId": "wf-001",
	})
	if resp.StatusCode == http.StatusOK {
		t.Errorf("unauthenticated request should be blocked, got 200")
	}
}

func TestSubmitPurchaseOrder_MissingWorkflowID(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := fiber.New()
	auth := withTenantCtx(testOrgID, testUserID, testUserRole)
	app.Post("/purchase-orders/:id/submit", auth, SubmitPurchaseOrder)

	resp := testRequest(app, http.MethodPost, "/purchase-orders/some-id/submit", map[string]interface{}{})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for missing workflowId, got %d", resp.StatusCode)
	}
}

func TestSubmitPurchaseOrder_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := fiber.New()
	auth := withTenantCtx(testOrgID, testUserID, testUserRole)
	app.Post("/purchase-orders/:id/submit", auth, SubmitPurchaseOrder)

	resp := testRequest(app, http.MethodPost, "/purchase-orders/non-existent-id/submit", map[string]interface{}{
		"workflowId": "wf-001",
	})
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestSubmitPurchaseOrder_AlreadyPending(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	order := makePurchaseOrder(t, "PO-PEND-001", "PENDING")

	app := fiber.New()
	auth := withTenantCtx(testOrgID, testUserID, testUserRole)
	app.Post("/purchase-orders/:id/submit", auth, SubmitPurchaseOrder)

	resp := testRequest(app, http.MethodPost, "/purchase-orders/"+order.ID+"/submit", map[string]interface{}{
		"workflowId": "wf-001",
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 when submitting non-DRAFT PO, got %d", resp.StatusCode)
	}
}
