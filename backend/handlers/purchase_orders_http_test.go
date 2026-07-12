package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
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

// decodeJSON reads the response body once and unmarshals it into the typed
// value pointed to by out. Fatals the test on any error.
func decodeJSON(t *testing.T, resp *http.Response, out interface{}) {
	t.Helper()
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		t.Fatalf("decodeJSON: %v", err)
	}
}

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
// PUT /purchase-orders/:id/items — ownership scope (IDOR)
// ─────────────────────────────────────────────────────────────────────────────

func TestUpdatePurchaseOrderItems_NonOwnerScopedOut(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupWorkflowTasksTable(t, db) // GetDocumentScope subquery targets this table

	// DRAFT PO owned by testUserID.
	order := models.PurchaseOrder{
		ID:             uuid.New().String(),
		OrganizationID: testOrgID,
		DocumentNumber: "PO-IDOR-1",
		Status:         "DRAFT",
		TotalAmount:    1000.00,
		Currency:       "ZMW",
		CreatedBy:      testUserID,
		DeliveryDate:   time.Now().Add(30 * 24 * time.Hour),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	order.Items = datatypes.NewJSONType([]types.POItem{{Description: "Widget A", Quantity: 10, UnitPrice: 100.0, Amount: 1000.0}})
	order.ApprovalHistory = datatypes.NewJSONType([]types.ApprovalRecord{})
	order.ActionHistory = datatypes.NewJSONType([]types.ActionHistoryEntry{})
	if err := db.Create(&order).Error; err != nil {
		t.Fatalf("create PO: %v", err)
	}

	// Caller is a different, non-privileged, uninvolved user.
	app := fiber.New()
	auth := withTenantCtx(testOrgID, "other-user-002", "requester")
	app.Put("/purchase-orders/:id/items", auth, UpdatePurchaseOrderItems)

	body := map[string]interface{}{
		"items":       []types.POItem{{Description: "Widget A", Quantity: 5, UnitPrice: 100.0, Amount: 500.0}},
		"totalAmount": 500.0,
	}
	resp := testRequest(app, http.MethodPut, "/purchase-orders/"+order.ID+"/items", body)
	if resp.StatusCode == http.StatusOK {
		t.Fatalf("expected non-200 (scoped out), got 200")
	}

	// Items must be unchanged in the DB.
	var reloaded models.PurchaseOrder
	if err := db.Where("id = ?", order.ID).First(&reloaded).Error; err != nil {
		t.Fatalf("reload PO: %v", err)
	}
	if reloaded.TotalAmount != 1000.00 {
		t.Fatalf("expected total unchanged at 1000, got %.2f", reloaded.TotalAmount)
	}
}

func TestUpdatePurchaseOrderItems_RecomputesTotalServerSide(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupWorkflowTasksTable(t, db)

	order := models.PurchaseOrder{
		ID:             uuid.New().String(),
		OrganizationID: testOrgID,
		DocumentNumber: "PO-RECOMPUTE-1",
		Status:         "DRAFT",
		TotalAmount:    1000.00,
		Currency:       "ZMW",
		CreatedBy:      testUserID,
		DeliveryDate:   time.Now().Add(30 * 24 * time.Hour),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	order.Items = datatypes.NewJSONType([]types.POItem{{Description: "Widget A", Quantity: 10, UnitPrice: 100.0, Amount: 1000.0}})
	order.ApprovalHistory = datatypes.NewJSONType([]types.ApprovalRecord{})
	order.ActionHistory = datatypes.NewJSONType([]types.ActionHistoryEntry{})
	if err := db.Create(&order).Error; err != nil {
		t.Fatalf("create PO: %v", err)
	}

	app := fiber.New()
	auth := withTenantCtx(testOrgID, testUserID, testUserRole)
	app.Put("/purchase-orders/:id/items", auth, UpdatePurchaseOrderItems)

	// Client sends a lying totalAmount (999) for items that sum to 500.
	body := map[string]interface{}{
		"items":       []types.POItem{{Description: "Widget A", Quantity: 5, UnitPrice: 100.0, Amount: 500.0}},
		"totalAmount": 999.0,
	}
	resp := testRequest(app, http.MethodPut, "/purchase-orders/"+order.ID+"/items", body)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	data, _ := decodeResponse(resp)["data"].(map[string]interface{})
	if data == nil {
		t.Fatalf("no data in response")
	}
	total, _ := data["totalAmount"].(float64)
	if total != 500.0 {
		t.Fatalf("expected server-recomputed total 500, got %v", total)
	}
}

func TestUpdatePurchaseOrderItems_OwnerCanEdit(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupWorkflowTasksTable(t, db)

	order := models.PurchaseOrder{
		ID:             uuid.New().String(),
		OrganizationID: testOrgID,
		DocumentNumber: "PO-IDOR-2",
		Status:         "DRAFT",
		TotalAmount:    1000.00,
		Currency:       "ZMW",
		CreatedBy:      "owner-user-003",
		DeliveryDate:   time.Now().Add(30 * 24 * time.Hour),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	order.Items = datatypes.NewJSONType([]types.POItem{{Description: "Widget A", Quantity: 10, UnitPrice: 100.0, Amount: 1000.0}})
	order.ApprovalHistory = datatypes.NewJSONType([]types.ApprovalRecord{})
	order.ActionHistory = datatypes.NewJSONType([]types.ActionHistoryEntry{})
	if err := db.Create(&order).Error; err != nil {
		t.Fatalf("create PO: %v", err)
	}

	app := fiber.New()
	auth := withTenantCtx(testOrgID, "owner-user-003", "requester")
	app.Put("/purchase-orders/:id/items", auth, UpdatePurchaseOrderItems)

	body := map[string]interface{}{
		"items":       []types.POItem{{Description: "Widget A", Quantity: 5, UnitPrice: 100.0, Amount: 500.0}},
		"totalAmount": 500.0,
	}
	resp := testRequest(app, http.MethodPut, "/purchase-orders/"+order.ID+"/items", body)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("owner edit: expected 200, got %d", resp.StatusCode)
	}
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
// PUT /purchase-orders/:id — metadata-only guard enumeration (security fix)
// ─────────────────────────────────────────────────────────────────────────────

// TestUpdatePurchaseOrder_MetadataOnly_ApprovedStatus_Succeeds verifies that a
// request containing ONLY metadata bypasses the DRAFT/PENDING status guard, so
// supporting documents can be attached to an already-APPROVED PO, and that no
// other field is mutated by the request.
func TestUpdatePurchaseOrder_MetadataOnly_ApprovedStatus_Succeeds(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupWorkflowTasksTable(t, db)

	order := makePurchaseOrder(t, "PO-META-001", "APPROVED")

	app := newPurchaseOrderApp(t)
	resp := testRequest(app, http.MethodPut, "/purchase-orders/"+order.ID, map[string]interface{}{
		"metadata": map[string]interface{}{
			"attachments": []map[string]interface{}{
				{"name": "invoice.pdf", "url": "https://example.com/invoice.pdf"},
			},
		},
	})
	if resp.StatusCode != http.StatusOK {
		body := decodeResponse(resp)
		t.Fatalf("expected 200 for metadata-only update on APPROVED PO, got %d; body=%v", resp.StatusCode, body)
	}

	var reloaded models.PurchaseOrder
	if err := db.Where("id = ?", order.ID).First(&reloaded).Error; err != nil {
		t.Fatalf("reload PO: %v", err)
	}
	if reloaded.Status != "APPROVED" {
		t.Fatalf("status must be untouched, want APPROVED, got %s", reloaded.Status)
	}
	if reloaded.TotalAmount != order.TotalAmount {
		t.Fatalf("totalAmount must be untouched, want %v, got %v", order.TotalAmount, reloaded.TotalAmount)
	}
	if reloaded.Title != order.Title {
		t.Fatalf("title must be untouched, want %q, got %q", order.Title, reloaded.Title)
	}

	var meta map[string]interface{}
	if err := json.Unmarshal(reloaded.Metadata, &meta); err != nil {
		t.Fatalf("unmarshal metadata: %v", err)
	}
	attachments, ok := meta["attachments"].([]interface{})
	if !ok || len(attachments) != 1 {
		t.Fatalf("expected 1 attachment persisted, got %v", meta["attachments"])
	}
}

// TestUpdatePurchaseOrder_NonMetadataField_ApprovedStatus_StillBlocked is a
// regression test for a bypass where isMetadataOnly only enumerated ~5 of the
// ~17 request fields: a request carrying metadata PLUS an unchecked field
// (title) was classified as "metadata-only" and slipped past the status
// guard, letting the handler mutate title on an APPROVED purchase order. The
// carve-out must be strictly metadata-only across EVERY field.
func TestUpdatePurchaseOrder_NonMetadataField_ApprovedStatus_StillBlocked(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupWorkflowTasksTable(t, db)

	order := makePurchaseOrder(t, "PO-META-002", "APPROVED")

	app := newPurchaseOrderApp(t)
	resp := testRequest(app, http.MethodPut, "/purchase-orders/"+order.ID, map[string]interface{}{
		"title":    "sneaking a field change past the guard",
		"metadata": map[string]interface{}{"attachments": []map[string]interface{}{{"name": "x.pdf"}}},
	})
	if resp.StatusCode != http.StatusForbidden {
		body := decodeResponse(resp)
		t.Fatalf("expected 403 when a non-metadata field is present on an APPROVED PO, got %d; body=%v", resp.StatusCode, body)
	}
	msg, _ := decodeResponse(resp)["message"].(string)
	if !strings.Contains(msg, "Cannot update purchase order in APPROVED status") {
		t.Fatalf("expected status-guard message, got %q", msg)
	}

	var reloaded models.PurchaseOrder
	if err := db.Where("id = ?", order.ID).First(&reloaded).Error; err != nil {
		t.Fatalf("reload PO: %v", err)
	}
	if reloaded.Title != order.Title {
		t.Fatalf("title must be untouched, want %q, got %q", order.Title, reloaded.Title)
	}
	if len(reloaded.Metadata) > 0 {
		t.Fatalf("metadata must not be persisted when the request also carries a blocked field, got %s", string(reloaded.Metadata))
	}
}

// TestUpdatePurchaseOrder_FullUpdate_DraftStatus_StillWorks is a regression
// guard: the enumerated carve-out must not break the normal DRAFT/PENDING
// full-update path.
func TestUpdatePurchaseOrder_FullUpdate_DraftStatus_StillWorks(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupWorkflowTasksTable(t, db)

	order := makePurchaseOrder(t, "PO-META-003", "DRAFT")

	app := newPurchaseOrderApp(t)
	resp := testRequest(app, http.MethodPut, "/purchase-orders/"+order.ID, map[string]interface{}{
		"title":       "Updated Title",
		"totalAmount": 2500.0,
		"currency":    "USD",
	})
	if resp.StatusCode != http.StatusOK {
		body := decodeResponse(resp)
		t.Fatalf("expected 200 for full update on DRAFT PO, got %d; body=%v", resp.StatusCode, body)
	}

	var reloaded models.PurchaseOrder
	if err := db.Where("id = ?", order.ID).First(&reloaded).Error; err != nil {
		t.Fatalf("reload PO: %v", err)
	}
	if reloaded.Title != "Updated Title" {
		t.Fatalf("expected title updated, got %q", reloaded.Title)
	}
	if reloaded.TotalAmount != 2500.0 {
		t.Fatalf("expected totalAmount updated to 2500, got %v", reloaded.TotalAmount)
	}
	if reloaded.Currency != "USD" {
		t.Fatalf("expected currency updated to USD, got %q", reloaded.Currency)
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

// ─────────────────────────────────────────────────────────────────────────────
// Manual vendor name persistence (bug regression)
// ─────────────────────────────────────────────────────────────────────────────

func TestPurchaseOrder_PersistManualVendorName(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newPurchaseOrderApp(t)

	// CREATE: manual vendor (no vendor_id, vendorName only)
	createBody := map[string]interface{}{
		"vendorId":     "",
		"vendorName":   "LIKS BUSINESS SOLUTIONS",
		"items":        []map[string]interface{}{{"description": "Laptops", "quantity": 25, "unitPrice": 150000, "amount": 3750000}},
		"totalAmount":  3750000.0,
		"currency":     "ZMW",
		"deliveryDate": time.Now().Add(7 * 24 * time.Hour).Format(time.RFC3339),
		"title":        "Purchase of 25 Laptops",
		"department":   "Information Technology",
		"priority":     "high",
	}
	resp := testRequest(app, http.MethodPost, "/purchase-orders", createBody)
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		t.Fatalf("create PO: expected 200/201, got %d", resp.StatusCode)
	}

	var createResp struct {
		Data types.PurchaseOrderResponse `json:"data"`
	}
	decodeJSON(t, resp, &createResp)

	if createResp.Data.VendorName != "LIKS BUSINESS SOLUTIONS" {
		t.Errorf("create response: expected vendorName=LIKS BUSINESS SOLUTIONS, got %q", createResp.Data.VendorName)
	}
	if createResp.Data.VendorID != "" {
		t.Errorf("create response: expected empty vendorId, got %q", createResp.Data.VendorID)
	}

	// GET: fetch the same PO and confirm vendor_name persists
	poID := createResp.Data.ID
	getResp := testRequest(app, http.MethodGet, "/purchase-orders/"+poID, nil)
	if getResp.StatusCode != http.StatusOK {
		t.Fatalf("get PO: expected 200, got %d", getResp.StatusCode)
	}

	var getBody struct {
		Data types.PurchaseOrderResponse `json:"data"`
	}
	decodeJSON(t, getResp, &getBody)

	if getBody.Data.VendorName != "LIKS BUSINESS SOLUTIONS" {
		t.Errorf("get response: expected vendorName=LIKS BUSINESS SOLUTIONS, got %q", getBody.Data.VendorName)
	}

	// UPDATE: change the manual vendor name
	updateBody := map[string]interface{}{
		"vendorId":   "",
		"vendorName": "MICOP BUSINESS VENTURES",
	}
	updResp := testRequest(app, http.MethodPut, "/purchase-orders/"+poID, updateBody)
	if updResp.StatusCode != http.StatusOK {
		t.Fatalf("update PO: expected 200, got %d", updResp.StatusCode)
	}

	var updRespBody struct {
		Data types.PurchaseOrderResponse `json:"data"`
	}
	decodeJSON(t, updResp, &updRespBody)

	if updRespBody.Data.VendorName != "MICOP BUSINESS VENTURES" {
		t.Errorf("update response: expected vendorName=MICOP BUSINESS VENTURES, got %q", updRespBody.Data.VendorName)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Scope isolation: procurement role cannot see direct_payment POs
// ─────────────────────────────────────────────────────────────────────────────

// makePurchaseOrderWithRouting creates a PO with an explicit routing_type.
func makePurchaseOrderWithRouting(t *testing.T, docNum, status, routingType string) models.PurchaseOrder {
	t.Helper()
	order := models.PurchaseOrder{
		ID:             uuid.New().String(),
		OrganizationID: testOrgID,
		DocumentNumber: docNum,
		Status:         status,
		RoutingType:    routingType,
		TotalAmount:    500.00,
		Currency:       "ZMW",
		DeliveryDate:   time.Now().Add(30 * 24 * time.Hour),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	order.Items = datatypes.NewJSONType([]types.POItem{
		{Description: "Item", Quantity: 1, UnitPrice: 500.0, Amount: 500.0},
	})
	order.ApprovalHistory = datatypes.NewJSONType([]types.ApprovalRecord{})
	order.ActionHistory = datatypes.NewJSONType([]types.ActionHistoryEntry{})
	if err := config.DB.Create(&order).Error; err != nil {
		t.Fatalf("makePurchaseOrderWithRouting: %v", err)
	}
	return order
}

// ─────────────────────────────────────────────────────────────────────────────
// Derived payment status (amountPaid/amountCommitted/balance/linkedPVs) — B5
// ─────────────────────────────────────────────────────────────────────────────

// makePaymentVoucherForPO creates and saves a PaymentVoucher linked to poDocNum.
func makePaymentVoucherForPO(t *testing.T, poDocNum, docNum, status string, amount float64) models.PaymentVoucher {
	t.Helper()
	pv := models.PaymentVoucher{
		ID:             uuid.New().String(),
		OrganizationID: testOrgID,
		DocumentNumber: docNum,
		LinkedPO:       poDocNum,
		Status:         status,
		Amount:         amount,
		Currency:       "ZMW",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	pv.ApprovalHistory = datatypes.NewJSONType([]types.ApprovalRecord{})
	if err := config.DB.Create(&pv).Error; err != nil {
		t.Fatalf("makePaymentVoucherForPO: %v", err)
	}
	return pv
}

// TestGetPurchaseOrder_DerivedPaymentStatus_PartiallyPaid seeds a PO with
// total=1000 and three PVs (PAID 400, DRAFT 600, REJECTED 100). REJECTED PVs
// are excluded from both the committed total and linkedPVs (they never
// consumed PO budget), so the response should reflect only the PAID + DRAFT
// pair: amountPaid=400, amountCommitted=1000, balance=0,
// paymentStatus="partially_paid", linkedPVs len 2, linkedPV = latest (DRAFT).
func TestGetPurchaseOrder_DerivedPaymentStatus_PartiallyPaid(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	order := makePurchaseOrder(t, "PO-PAY-STATUS-001", "PENDING")
	makePaymentVoucherForPO(t, order.DocumentNumber, "PV-PAY-STATUS-001", "PAID", 400.0)
	makePaymentVoucherForPO(t, order.DocumentNumber, "PV-PAY-STATUS-002", "DRAFT", 600.0)
	makePaymentVoucherForPO(t, order.DocumentNumber, "PV-PAY-STATUS-003", "REJECTED", 100.0)

	app := newPurchaseOrderApp(t)

	// Detail endpoint.
	resp := testRequest(app, http.MethodGet, "/purchase-orders/"+order.ID, nil)
	if resp.StatusCode != http.StatusOK {
		body := decodeResponse(resp)
		t.Fatalf("expected 200, got %d; body=%v", resp.StatusCode, body)
	}
	var detail struct {
		Data types.PurchaseOrderResponse `json:"data"`
	}
	decodeJSON(t, resp, &detail)

	if detail.Data.AmountPaid != 400.0 {
		t.Errorf("detail: expected amountPaid=400, got %v", detail.Data.AmountPaid)
	}
	if detail.Data.AmountCommitted != 1000.0 {
		t.Errorf("detail: expected amountCommitted=1000, got %v", detail.Data.AmountCommitted)
	}
	if detail.Data.Balance != 0.0 {
		t.Errorf("detail: expected balance=0, got %v", detail.Data.Balance)
	}
	if detail.Data.PaymentStatus != "partially_paid" {
		t.Errorf("detail: expected paymentStatus=partially_paid, got %q", detail.Data.PaymentStatus)
	}
	if len(detail.Data.LinkedPVs) != 2 {
		t.Fatalf("detail: expected linkedPVs len 2, got %d (%+v)", len(detail.Data.LinkedPVs), detail.Data.LinkedPVs)
	}
	if detail.Data.LinkedPV == nil || detail.Data.LinkedPV.Status != "DRAFT" {
		t.Errorf("detail: expected linkedPV=latest (DRAFT), got %+v", detail.Data.LinkedPV)
	}

	// List endpoint returns the same derived fields.
	listResp := testRequest(app, http.MethodGet, "/purchase-orders", nil)
	if listResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", listResp.StatusCode)
	}
	var list struct {
		Data []types.PurchaseOrderResponse `json:"data"`
	}
	decodeJSON(t, listResp, &list)

	var found *types.PurchaseOrderResponse
	for i := range list.Data {
		if list.Data[i].ID == order.ID {
			found = &list.Data[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("list: PO %s not found in response", order.ID)
	}
	if found.AmountPaid != 400.0 {
		t.Errorf("list: expected amountPaid=400, got %v", found.AmountPaid)
	}
	if found.AmountCommitted != 1000.0 {
		t.Errorf("list: expected amountCommitted=1000, got %v", found.AmountCommitted)
	}
	if found.Balance != 0.0 {
		t.Errorf("list: expected balance=0, got %v", found.Balance)
	}
	if found.PaymentStatus != "partially_paid" {
		t.Errorf("list: expected paymentStatus=partially_paid, got %q", found.PaymentStatus)
	}
	if len(found.LinkedPVs) != 2 {
		t.Errorf("list: expected linkedPVs len 2, got %d", len(found.LinkedPVs))
	}
}

// TestCreatePurchaseOrder_DerivedPaymentStatus_Unpaid verifies the cheap
// create-path enrichment: a brand-new PO with no PVs is "unpaid" with
// balance == totalAmount.
func TestCreatePurchaseOrder_DerivedPaymentStatus_Unpaid(t *testing.T) {
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
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
	var created struct {
		Data types.PurchaseOrderResponse `json:"data"`
	}
	decodeJSON(t, resp, &created)

	if created.Data.PaymentStatus != "unpaid" {
		t.Errorf("expected paymentStatus=unpaid, got %q", created.Data.PaymentStatus)
	}
	if created.Data.Balance != 500.0 {
		t.Errorf("expected balance=500, got %v", created.Data.Balance)
	}
	if created.Data.AmountPaid != 0.0 {
		t.Errorf("expected amountPaid=0, got %v", created.Data.AmountPaid)
	}
}

// TestPO_ProcurementUserCannotSeeDirectPayment verifies that a procurement-role
// user receives 404 when fetching a direct_payment PO by ID (single-get endpoint).
// The list endpoint is sqlc-backed and cannot be exercised in SQLite tests.
func TestPO_ProcurementUserCannotSeeDirectPayment(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Seed a direct_payment PO and a procurement PO.
	directPO := makePurchaseOrderWithRouting(t, "PO-DIRECT-SCOPE-001", "draft", "direct_payment")
	procPO := makePurchaseOrderWithRouting(t, "PO-PROC-SCOPE-001", "draft", "procurement")

	// Build app authenticating as procurement role.
	app := fiber.New()
	procAuth := withTenantCtx(testOrgID, testUserID, "procurement")
	app.Get("/purchase-orders/:id", procAuth, GetPurchaseOrder)

	// direct_payment PO → should be invisible (404, not 403, to avoid info leak).
	resp := testRequest(app, http.MethodGet, "/purchase-orders/"+directPO.ID, nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("procurement user: expected 404 for direct_payment PO, got %d", resp.StatusCode)
	}

	// procurement PO → should be visible (200).
	resp2 := testRequest(app, http.MethodGet, "/purchase-orders/"+procPO.ID, nil)
	if resp2.StatusCode != http.StatusOK {
		body := decodeResponse(resp2)
		t.Errorf("procurement user: expected 200 for procurement PO, got %d; body=%v", resp2.StatusCode, body)
	}
}
