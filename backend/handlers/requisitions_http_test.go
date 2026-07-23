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
	"gorm.io/gorm"
)

// ─────────────────────────────────────────────────────────────────────────────
// app factory
// ─────────────────────────────────────────────────────────────────────────────

func newRequisitionApp(t *testing.T) *fiber.App {
	t.Helper()
	app := fiber.New()
	auth := withTenantCtx(testOrgID, testUserID, testUserRole)

	app.Get("/requisitions", auth, GetRequisitions)
	app.Get("/requisitions/stats", auth, GetRequisitionStats)
	app.Post("/requisitions", auth, CreateRequisition)
	app.Get("/requisitions/:id", auth, GetRequisition)
	app.Put("/requisitions/:id", auth, UpdateRequisition)
	app.Delete("/requisitions/:id", auth, DeleteRequisition)
	app.Post("/requisitions/:id/submit", auth, SubmitRequisition)
	app.Post("/requisitions/:id/withdraw", auth, WithdrawRequisition)
	return app
}

// seedTestUser inserts testUserID into the users table when not already present.
func seedTestUser(t *testing.T) {
	t.Helper()
	var count int64
	config.DB.Model(&models.User{}).Where("id = ?", testUserID).Count(&count)
	if count == 0 {
		u := models.User{
			ID:     testUserID,
			Email:  "test@example.com",
			Name:   "Test User",
			Role:   testUserRole,
			Active: true,
		}
		if err := config.DB.Create(&u).Error; err != nil {
			t.Fatalf("seedTestUser: %v", err)
		}
	}
}

// makeRequisition builds and saves a Requisition with the given status.
func makeRequisition(t *testing.T, docNum, status string) models.Requisition {
	t.Helper()
	req := models.Requisition{
		ID:             uuid.New().String(),
		OrganizationID: testOrgID,
		DocumentNumber: docNum,
		RequesterId:    testUserID,
		RequesterName:  "Test User",
		Title:          "Test Requisition",
		Description:    "Test description value",
		Department:     "Engineering",
		Status:         status,
		Priority:       "medium",
		TotalAmount:    500.00,
		Currency:       "ZMW",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	req.Items = datatypes.NewJSONType([]types.RequisitionItem{
		{Description: "Item A", Quantity: 2, UnitPrice: 250.00, Amount: 500.00},
	})
	req.ApprovalHistory = datatypes.NewJSONType([]types.ApprovalRecord{})
	req.ActionHistory = datatypes.NewJSONType([]types.ActionHistoryEntry{})
	if err := config.DB.Create(&req).Error; err != nil {
		t.Fatalf("makeRequisition: %v", err)
	}
	return req
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /requisitions
// ─────────────────────────────────────────────────────────────────────────────

func TestGetRequisitions_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := fiber.New()
	app.Get("/requisitions", GetRequisitions)

	resp := testRequest(app, http.MethodGet, "/requisitions", nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

func TestGetRequisitions_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newRequisitionApp(t)
	resp := testRequest(app, http.MethodGet, "/requisitions", nil)
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
// GET /requisitions/stats
// ─────────────────────────────────────────────────────────────────────────────

func TestGetRequisitionStats_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := fiber.New()
	app.Get("/requisitions/stats", GetRequisitionStats)

	resp := testRequest(app, http.MethodGet, "/requisitions/stats", nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

func TestGetRequisitionStats_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newRequisitionApp(t)
	resp := testRequest(app, http.MethodGet, "/requisitions/stats", nil)
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
// POST /requisitions
// ─────────────────────────────────────────────────────────────────────────────

func TestCreateRequisition_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := fiber.New()
	app.Use(fiberrecover.New())
	app.Post("/requisitions", CreateRequisition)

	resp := testRequest(app, http.MethodPost, "/requisitions", map[string]interface{}{
		"title": "Test",
	})
	if resp.StatusCode == http.StatusOK {
		t.Errorf("unauthenticated request should be blocked, got 200")
	}
}

func TestCreateRequisition_MissingTitle(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newRequisitionApp(t)
	body := map[string]interface{}{
		"description": "A sufficiently long description for the requisition",
		"department":  "Engineering",
		"priority":    "medium",
		"totalAmount": 100.0,
		"currency":    "ZMW",
		"items": []map[string]interface{}{
			{"description": "Widget", "quantity": 1, "unitPrice": 100.0, "amount": 100.0},
		},
	}
	resp := testRequest(app, http.MethodPost, "/requisitions", body)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for missing title, got %d", resp.StatusCode)
	}
}

func TestCreateRequisition_TitleTooShort(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newRequisitionApp(t)
	body := map[string]interface{}{
		"title":       "Ab",
		"description": "A sufficiently long description for the requisition",
		"department":  "Engineering",
		"priority":    "medium",
		"totalAmount": 100.0,
		"currency":    "ZMW",
		"items": []map[string]interface{}{
			{"description": "Widget", "quantity": 1, "unitPrice": 100.0, "amount": 100.0},
		},
	}
	resp := testRequest(app, http.MethodPost, "/requisitions", body)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for title too short, got %d", resp.StatusCode)
	}
}

func TestCreateRequisition_MissingDescription(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newRequisitionApp(t)
	body := map[string]interface{}{
		"title":       "Test Requisition",
		"department":  "Engineering",
		"priority":    "medium",
		"totalAmount": 100.0,
		"currency":    "ZMW",
		"items": []map[string]interface{}{
			{"description": "Widget", "quantity": 1, "unitPrice": 100.0, "amount": 100.0},
		},
	}
	resp := testRequest(app, http.MethodPost, "/requisitions", body)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for missing description, got %d", resp.StatusCode)
	}
}

func TestCreateRequisition_MissingItems(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newRequisitionApp(t)
	body := map[string]interface{}{
		"title":       "Test Requisition",
		"description": "A sufficiently long description right here",
		"department":  "Engineering",
		"priority":    "medium",
		"totalAmount": 100.0,
		"currency":    "ZMW",
		"items":       []map[string]interface{}{},
	}
	resp := testRequest(app, http.MethodPost, "/requisitions", body)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for empty items, got %d", resp.StatusCode)
	}
}

func TestCreateRequisition_ZeroAmount(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newRequisitionApp(t)
	body := map[string]interface{}{
		"title":       "Test Requisition",
		"description": "A sufficiently long description right here",
		"department":  "Engineering",
		"priority":    "medium",
		"totalAmount": 0,
		"currency":    "ZMW",
		"items": []map[string]interface{}{
			{"description": "Widget", "quantity": 1, "unitPrice": 100.0, "amount": 100.0},
		},
	}
	resp := testRequest(app, http.MethodPost, "/requisitions", body)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for zero totalAmount, got %d", resp.StatusCode)
	}
}

func TestCreateRequisition_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	seedTestUser(t)

	app := newRequisitionApp(t)
	body := map[string]interface{}{
		"title":       "Office Supplies",
		"description": "Monthly office supplies requisition for the team",
		"department":  "Engineering",
		"priority":    "medium",
		"totalAmount": 500.0,
		"currency":    "ZMW",
		"items": []map[string]interface{}{
			{"description": "Pens", "quantity": 10, "unitPrice": 50.0, "amount": 500.0},
		},
	}
	resp := testRequest(app, http.MethodPost, "/requisitions", body)
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
		t.Errorf("expected documentNumber in response, got %v", data["documentNumber"])
	}
	if data["status"] != "DRAFT" {
		t.Errorf("expected status DRAFT, got %v", data["status"])
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /requisitions/:id
// ─────────────────────────────────────────────────────────────────────────────

func TestGetRequisition_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := fiber.New()
	app.Use(fiberrecover.New())
	app.Get("/requisitions/:id", GetRequisition)

	resp := testRequest(app, http.MethodGet, "/requisitions/some-id", nil)
	if resp.StatusCode == http.StatusOK {
		t.Errorf("unauthenticated request should be blocked, got 200")
	}
}

func TestGetRequisition_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newRequisitionApp(t)
	resp := testRequest(app, http.MethodGet, "/requisitions/non-existent-id", nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestGetRequisition_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	req := makeRequisition(t, "REQ-TEST-001", "DRAFT")

	app := newRequisitionApp(t)
	resp := testRequest(app, http.MethodGet, "/requisitions/"+req.ID, nil)
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
// PUT /requisitions/:id
// ─────────────────────────────────────────────────────────────────────────────

func TestUpdateRequisition_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := fiber.New()
	app.Use(fiberrecover.New())
	app.Put("/requisitions/:id", UpdateRequisition)

	resp := testRequest(app, http.MethodPut, "/requisitions/some-id", map[string]interface{}{
		"title": "Updated",
	})
	if resp.StatusCode == http.StatusOK {
		t.Errorf("unauthenticated request should be blocked, got 200")
	}
}

func TestUpdateRequisition_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newRequisitionApp(t)
	resp := testRequest(app, http.MethodPut, "/requisitions/non-existent-id", map[string]interface{}{
		"title": "Updated Title",
	})
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestUpdateRequisition_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	seedTestUser(t)
	req := makeRequisition(t, "REQ-TEST-002", "DRAFT")

	app := newRequisitionApp(t)
	resp := testRequest(app, http.MethodPut, "/requisitions/"+req.ID, map[string]interface{}{
		"title": "Updated Title",
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
// DELETE /requisitions/:id
// ─────────────────────────────────────────────────────────────────────────────

func TestDeleteRequisition_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := fiber.New()
	app.Use(fiberrecover.New())
	app.Delete("/requisitions/:id", DeleteRequisition)

	resp := testRequest(app, http.MethodDelete, "/requisitions/some-id", nil)
	if resp.StatusCode == http.StatusOK {
		t.Errorf("unauthenticated request should be blocked, got 200")
	}
}

func TestDeleteRequisition_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newRequisitionApp(t)
	resp := testRequest(app, http.MethodDelete, "/requisitions/non-existent-id", nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestDeleteRequisition_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	req := makeRequisition(t, "REQ-DEL-001", "DRAFT")

	app := newRequisitionApp(t)
	resp := testRequest(app, http.MethodDelete, "/requisitions/"+req.ID, nil)
	if resp.StatusCode != http.StatusOK {
		body := decodeResponse(resp)
		t.Fatalf("expected 200, got %d; body=%v", resp.StatusCode, body)
	}
	body := decodeResponse(resp)
	if body["success"] != true {
		t.Errorf("expected success=true, got %v", body["success"])
	}
}

func TestDeleteRequisition_NonDraftForbidden(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	req := makeRequisition(t, "REQ-DEL-002", "PENDING")

	app := newRequisitionApp(t)
	resp := testRequest(app, http.MethodDelete, "/requisitions/"+req.ID, nil)
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403 for non-draft delete, got %d", resp.StatusCode)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// POST /requisitions/:id/submit
// ─────────────────────────────────────────────────────────────────────────────

func TestSubmitRequisition_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := fiber.New()
	app.Use(fiberrecover.New())
	app.Post("/requisitions/:id/submit", SubmitRequisition)

	resp := testRequest(app, http.MethodPost, "/requisitions/some-id/submit", map[string]interface{}{
		"workflowId": "wf-001",
	})
	if resp.StatusCode == http.StatusOK {
		t.Errorf("unauthenticated request should be blocked, got 200")
	}
}

func TestSubmitRequisition_MissingWorkflowID(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := fiber.New()
	auth := withTenantCtx(testOrgID, testUserID, testUserRole)
	app.Post("/requisitions/:id/submit", auth, SubmitRequisition)

	resp := testRequest(app, http.MethodPost, "/requisitions/some-id/submit", map[string]interface{}{})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for missing workflowId, got %d", resp.StatusCode)
	}
}

func TestSubmitRequisition_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := fiber.New()
	auth := withTenantCtx(testOrgID, testUserID, testUserRole)
	app.Post("/requisitions/:id/submit", auth, SubmitRequisition)

	resp := testRequest(app, http.MethodPost, "/requisitions/non-existent-id/submit", map[string]interface{}{
		"workflowId": "wf-001",
	})
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestSubmitRequisition_AlreadyPending(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	req := makeRequisition(t, "REQ-PEND-001", "PENDING")

	app := fiber.New()
	auth := withTenantCtx(testOrgID, testUserID, testUserRole)
	app.Post("/requisitions/:id/submit", auth, SubmitRequisition)

	resp := testRequest(app, http.MethodPost, "/requisitions/"+req.ID+"/submit", map[string]interface{}{
		"workflowId": "wf-001",
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 when submitting non-DRAFT requisition, got %d", resp.StatusCode)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// POST /requisitions/:id/withdraw
// ─────────────────────────────────────────────────────────────────────────────

func TestWithdrawRequisition_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := fiber.New()
	app.Use(fiberrecover.New())
	app.Post("/requisitions/:id/withdraw", WithdrawRequisition)

	resp := testRequest(app, http.MethodPost, "/requisitions/some-id/withdraw", nil)
	if resp.StatusCode == http.StatusOK {
		t.Errorf("unauthenticated request should be blocked, got 200")
	}
}

func TestWithdrawRequisition_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := fiber.New()
	auth := withTenantCtx(testOrgID, testUserID, testUserRole)
	app.Post("/requisitions/:id/withdraw", auth, WithdrawRequisition)

	resp := testRequest(app, http.MethodPost, "/requisitions/non-existent-id/withdraw", nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestWithdrawRequisition_DraftStatusBadRequest(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	req := makeRequisition(t, "REQ-WD-001", "DRAFT")

	app := fiber.New()
	auth := withTenantCtx(testOrgID, testUserID, testUserRole)
	app.Post("/requisitions/:id/withdraw", auth, WithdrawRequisition)

	// DRAFT requisitions cannot be withdrawn — handler requires PENDING
	resp := testRequest(app, http.MethodPost, "/requisitions/"+req.ID+"/withdraw", nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 when withdrawing a DRAFT requisition, got %d", resp.StatusCode)
	}
}

func TestWithdrawRequisition_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupWorkflowTasksTable(t, db)        // WithdrawRequisition deletes workflow_tasks
	setupWorkflowAssignmentsTable(t, db)  // WithdrawRequisition deletes workflow_assignments

	req := makeRequisition(t, "REQ-WD-002", "PENDING")

	app := fiber.New()
	auth := withTenantCtx(testOrgID, testUserID, testUserRole)
	app.Post("/requisitions/:id/withdraw", auth, WithdrawRequisition)

	resp := testRequest(app, http.MethodPost, "/requisitions/"+req.ID+"/withdraw", nil)
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
// helpers for chain / audit-trail tests
// ─────────────────────────────────────────────────────────────────────────────

// newRequisitionAppExtended returns a Fiber app with all requisition routes
// including the three endpoints added in this file.
func newRequisitionAppExtended(t *testing.T) *fiber.App {
	t.Helper()
	app := fiber.New()
	auth := withTenantCtx(testOrgID, testUserID, testUserRole)

	app.Get("/requisitions", auth, GetRequisitions)
	app.Get("/requisitions/stats", auth, GetRequisitionStats)
	app.Post("/requisitions", auth, CreateRequisition)
	app.Get("/requisitions/:id", auth, GetRequisition)
	app.Put("/requisitions/:id", auth, UpdateRequisition)
	app.Delete("/requisitions/:id", auth, DeleteRequisition)
	app.Post("/requisitions/:id/submit", auth, SubmitRequisition)
	app.Post("/requisitions/:id/withdraw", auth, WithdrawRequisition)
	app.Post("/requisitions/:id/reassign", auth, ReassignRequisition)
	app.Get("/requisitions/:id/chain", auth, GetRequisitionChain)
	app.Get("/requisitions/:id/audit-trail", auth, GetRequisitionAuditTrail)
	return app
}

// newRequisitionAppExtendedWithRole is like newRequisitionAppExtended but
// allows callers to override the user role (used by role-restricted endpoints).
func newRequisitionAppExtendedWithRole(t *testing.T, role string) *fiber.App {
	t.Helper()
	app := fiber.New()
	auth := withTenantCtx(testOrgID, testUserID, role)

	app.Post("/requisitions/:id/reassign", auth, ReassignRequisition)
	app.Get("/requisitions/:id/chain", auth, GetRequisitionChain)
	app.Get("/requisitions/:id/audit-trail", auth, GetRequisitionAuditTrail)
	return app
}

// ensureDocumentLinksTable ensures the document_links table exists in the test DB.
func ensureDocumentLinksTable(t *testing.T, db *gorm.DB) {
	t.Helper()
	sql := `CREATE TABLE IF NOT EXISTS document_links (
		id TEXT PRIMARY KEY,
		source_doc_id TEXT NOT NULL DEFAULT '',
		source_doc_type TEXT NOT NULL DEFAULT '',
		target_doc_id TEXT NOT NULL DEFAULT '',
		target_doc_type TEXT NOT NULL DEFAULT '',
		link_type TEXT NOT NULL DEFAULT '',
		amount REAL DEFAULT 0,
		proportion REAL DEFAULT 0,
		status TEXT DEFAULT 'active',
		created_at DATETIME,
		updated_at DATETIME
	)`
	if err := db.Exec(sql).Error; err != nil {
		t.Fatalf("ensureDocumentLinksTable: %v", err)
	}
}

// ensureAuditLogsTable ensures the audit_logs table exists in the test DB.
func ensureAuditLogsTable(t *testing.T, db *gorm.DB) {
	t.Helper()
	sql := `CREATE TABLE IF NOT EXISTS audit_logs (
		id TEXT PRIMARY KEY,
		document_id TEXT NOT NULL DEFAULT '',
		document_type TEXT NOT NULL DEFAULT '',
		user_id TEXT NOT NULL DEFAULT '',
		action TEXT NOT NULL DEFAULT '',
		changes JSON,
		created_at DATETIME
	)`
	if err := db.Exec(sql).Error; err != nil {
		t.Fatalf("ensureAuditLogsTable: %v", err)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// POST /requisitions/:id/reassign
// ─────────────────────────────────────────────────────────────────────────────

func TestReassignRequisition_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := fiber.New()
	app.Use(fiberrecover.New())
	app.Post("/requisitions/:id/reassign", ReassignRequisition)

	resp := testRequest(app, http.MethodPost, "/requisitions/some-id/reassign", map[string]interface{}{
		"newApproverId": "approver-001",
	})
	if resp.StatusCode == http.StatusOK {
		t.Errorf("unauthenticated request should be blocked, got 200")
	}
}

func TestReassignRequisition_MissingApproverID(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := fiber.New()
	auth := withTenantCtx(testOrgID, testUserID, testUserRole)
	app.Post("/requisitions/:id/reassign", auth, ReassignRequisition)

	resp := testRequest(app, http.MethodPost, "/requisitions/some-id/reassign", map[string]interface{}{})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for missing newApproverId, got %d", resp.StatusCode)
	}
}

func TestReassignRequisition_ApproverNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := fiber.New()
	auth := withTenantCtx(testOrgID, testUserID, testUserRole)
	app.Post("/requisitions/:id/reassign", auth, ReassignRequisition)

	resp := testRequest(app, http.MethodPost, "/requisitions/some-id/reassign", map[string]interface{}{
		"newApproverId": "non-existent-approver",
	})
	// Handler returns 400 when approver user record is not found
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 when approver not found, got %d", resp.StatusCode)
	}
}

func TestReassignRequisition_RequisitionNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	seedTestUser(t) // testUserID acts as the new approver

	app := fiber.New()
	auth := withTenantCtx(testOrgID, testUserID, testUserRole)
	app.Post("/requisitions/:id/reassign", auth, ReassignRequisition)

	resp := testRequest(app, http.MethodPost, "/requisitions/non-existent-req/reassign", map[string]interface{}{
		"newApproverId": testUserID,
	})
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404 when requisition not found, got %d", resp.StatusCode)
	}
}

func TestReassignRequisition_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	seedTestUser(t)
	req := makeRequisition(t, "REQ-REASSIGN-001", "PENDING")

	app := fiber.New()
	auth := withTenantCtx(testOrgID, testUserID, testUserRole)
	app.Post("/requisitions/:id/reassign", auth, ReassignRequisition)

	resp := testRequest(app, http.MethodPost, "/requisitions/"+req.ID+"/reassign", map[string]interface{}{
		"newApproverId": testUserID,
		"reason":        "Reassigning due to availability",
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
// GET /requisitions/:id/chain
// ─────────────────────────────────────────────────────────────────────────────

func TestGetRequisitionChain_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := fiber.New()
	app.Use(fiberrecover.New())
	app.Get("/requisitions/:id/chain", GetRequisitionChain)

	resp := testRequest(app, http.MethodGet, "/requisitions/some-id/chain", nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

func TestGetRequisitionChain_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	ensureDocumentLinksTable(t, db)

	app := fiber.New()
	auth := withTenantCtx(testOrgID, testUserID, testUserRole)
	app.Get("/requisitions/:id/chain", auth, GetRequisitionChain)

	resp := testRequest(app, http.MethodGet, "/requisitions/non-existent-id/chain", nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestGetRequisitionChain_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	ensureDocumentLinksTable(t, db)
	setupWorkflowAssignmentsTable(t, db)

	req := makeRequisition(t, "REQ-CHAIN-001", "APPROVED")

	app := fiber.New()
	auth := withTenantCtx(testOrgID, testUserID, testUserRole)
	app.Get("/requisitions/:id/chain", auth, GetRequisitionChain)

	resp := testRequest(app, http.MethodGet, "/requisitions/"+req.ID+"/chain", nil)
	if resp.StatusCode != http.StatusOK {
		body := decodeResponse(resp)
		t.Fatalf("expected 200, got %d; body=%v", resp.StatusCode, body)
	}
	body := decodeResponse(resp)
	if body["success"] != true {
		t.Errorf("expected success=true, got %v", body["success"])
	}
	data, ok := body["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object in response")
	}
	if data["requisitionId"] != req.ID {
		t.Errorf("expected requisitionId=%s, got %v", req.ID, data["requisitionId"])
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /requisitions/:id/audit-trail
// ─────────────────────────────────────────────────────────────────────────────

func TestGetRequisitionAuditTrail_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := fiber.New()
	app.Use(fiberrecover.New())
	app.Get("/requisitions/:id/audit-trail", GetRequisitionAuditTrail)

	resp := testRequest(app, http.MethodGet, "/requisitions/some-id/audit-trail", nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

func TestGetRequisitionAuditTrail_ForbiddenRole(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := fiber.New()
	// "requester" is not in the allowed list (admin, super_admin, manager, finance)
	auth := withTenantCtx(testOrgID, testUserID, "requester")
	app.Get("/requisitions/:id/audit-trail", auth, GetRequisitionAuditTrail)

	resp := testRequest(app, http.MethodGet, "/requisitions/some-id/audit-trail", nil)
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403 for requester role, got %d", resp.StatusCode)
	}
}

func TestGetRequisitionAuditTrail_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	ensureDocumentLinksTable(t, db)
	ensureAuditLogsTable(t, db)

	app := fiber.New()
	// testUserRole is "admin" which is allowed
	auth := withTenantCtx(testOrgID, testUserID, testUserRole)
	app.Get("/requisitions/:id/audit-trail", auth, GetRequisitionAuditTrail)

	resp := testRequest(app, http.MethodGet, "/requisitions/non-existent-id/audit-trail", nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestGetRequisitionAuditTrail_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	ensureDocumentLinksTable(t, db)
	ensureAuditLogsTable(t, db)

	req := makeRequisition(t, "REQ-AUDIT-001", "APPROVED")

	app := fiber.New()
	// testUserRole is "admin" which is allowed
	auth := withTenantCtx(testOrgID, testUserID, testUserRole)
	app.Get("/requisitions/:id/audit-trail", auth, GetRequisitionAuditTrail)

	resp := testRequest(app, http.MethodGet, "/requisitions/"+req.ID+"/audit-trail", nil)
	if resp.StatusCode != http.StatusOK {
		body := decodeResponse(resp)
		t.Fatalf("expected 200, got %d; body=%v", resp.StatusCode, body)
	}
	body := decodeResponse(resp)
	if body["success"] != true {
		t.Errorf("expected success=true, got %v", body["success"])
	}
	// data should be a list (possibly empty when no audit logs exist)
	if _, ok := body["data"]; !ok {
		t.Errorf("expected data field in response")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// manual preferred vendor name (ad-hoc, no vendor record)
// ─────────────────────────────────────────────────────────────────────────────

func TestRequisition_PersistManualPreferredVendorName(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	seedTestUser(t)

	app := newRequisitionApp(t)

	createBody := map[string]interface{}{
		"title":               "Test requisition with manual preferred vendor",
		"description":         "Need supplies from a vendor not in the system",
		"department":          "Information Technology",
		"priority":            "medium",
		"items":               []map[string]interface{}{{"description": "Toner", "quantity": 5, "unitPrice": 200, "amount": 1000}},
		"totalAmount":         1000.0,
		"currency":            "ZMW",
		"preferredVendorName": "AD-HOC SUPPLIER LTD",
	}
	resp := testRequest(app, http.MethodPost, "/requisitions", createBody)
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		t.Fatalf("create requisition: expected 200/201, got %d", resp.StatusCode)
	}

	var createResp struct {
		Data types.RequisitionResponse `json:"data"`
	}
	decodeJSON(t, resp, &createResp)

	if createResp.Data.PreferredVendorName != "AD-HOC SUPPLIER LTD" {
		t.Errorf("create response: expected preferredVendorName=AD-HOC SUPPLIER LTD, got %q", createResp.Data.PreferredVendorName)
	}

	reqID := createResp.Data.ID
	getResp := testRequest(app, http.MethodGet, "/requisitions/"+reqID, nil)
	if getResp.StatusCode != http.StatusOK {
		t.Fatalf("get requisition: expected 200, got %d", getResp.StatusCode)
	}
	var getBody struct {
		Data types.RequisitionResponse `json:"data"`
	}
	decodeJSON(t, getResp, &getBody)
	if getBody.Data.PreferredVendorName != "AD-HOC SUPPLIER LTD" {
		t.Errorf("get response: expected preferredVendorName=AD-HOC SUPPLIER LTD, got %q", getBody.Data.PreferredVendorName)
	}

	updateBody := map[string]interface{}{
		"preferredVendorName": "ANOTHER AD-HOC VENDOR",
	}
	updResp := testRequest(app, http.MethodPut, "/requisitions/"+reqID, updateBody)
	if updResp.StatusCode != http.StatusOK {
		t.Fatalf("update requisition: expected 200, got %d", updResp.StatusCode)
	}
	var updRespBody struct {
		Data types.RequisitionResponse `json:"data"`
	}
	decodeJSON(t, updResp, &updRespBody)
	if updRespBody.Data.PreferredVendorName != "ANOTHER AD-HOC VENDOR" {
		t.Errorf("update response: expected preferredVendorName=ANOTHER AD-HOC VENDOR, got %q", updRespBody.Data.PreferredVendorName)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Scope isolation: procurement role cannot see direct_payment requisitions
// ─────────────────────────────────────────────────────────────────────────────

// makeRequisitionWithRouting creates a Requisition with an explicit routing_type.
func makeRequisitionWithRouting(t *testing.T, docNum, status, routingType string) models.Requisition {
	t.Helper()
	req := models.Requisition{
		ID:             uuid.New().String(),
		OrganizationID: testOrgID,
		DocumentNumber: docNum,
		RequesterId:    testUserID,
		RequesterName:  "Test User",
		Title:          "Test " + routingType + " Requisition",
		Department:     "Engineering",
		Status:         status,
		Priority:       "medium",
		RoutingType:    routingType,
		TotalAmount:    250.00,
		Currency:       "ZMW",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	req.Items = datatypes.NewJSONType([]types.RequisitionItem{
		{Description: "Item", Quantity: 1, UnitPrice: 250.00, Amount: 250.00},
	})
	req.ApprovalHistory = datatypes.NewJSONType([]types.ApprovalRecord{})
	req.ActionHistory = datatypes.NewJSONType([]types.ActionHistoryEntry{})
	if err := config.DB.Create(&req).Error; err != nil {
		t.Fatalf("makeRequisitionWithRouting: %v", err)
	}
	return req
}

// TestRequisition_ProcurementUserCannotSeeDirectPayment verifies that a
// procurement-role user receives 404 for a direct_payment requisition (single-get).
func TestRequisition_ProcurementUserCannotSeeDirectPayment(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	directReq := makeRequisitionWithRouting(t, "REQ-DIRECT-SCOPE-001", "draft", "direct_payment")
	procReq := makeRequisitionWithRouting(t, "REQ-PROC-SCOPE-001", "draft", "procurement")

	app := fiber.New()
	procAuth := withTenantCtx(testOrgID, testUserID, "procurement")
	app.Get("/requisitions/:id", procAuth, GetRequisition)

	// direct_payment requisition → 404 (invisible, no info leak).
	resp := testRequest(app, http.MethodGet, "/requisitions/"+directReq.ID, nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("procurement user: expected 404 for direct_payment requisition, got %d", resp.StatusCode)
	}

	// procurement requisition → 200.
	resp2 := testRequest(app, http.MethodGet, "/requisitions/"+procReq.ID, nil)
	if resp2.StatusCode != http.StatusOK {
		body := decodeResponse(resp2)
		t.Errorf("procurement user: expected 200 for procurement requisition, got %d; body=%v", resp2.StatusCode, body)
	}
}
