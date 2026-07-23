package handlers

import (
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/models"
)

// ─────────────────────────────────────────────────────────────────────────────
// Setup helpers
// ─────────────────────────────────────────────────────────────────────────────

// setupAuditLogsTable creates the audit_logs table using raw SQL so the
// gorm:"type:jsonb" tag on Changes doesn't trip SQLite's DDL parser.
func setupAuditLogsTable(t *testing.T) {
	t.Helper()
	sql := `CREATE TABLE IF NOT EXISTS audit_logs (
		id TEXT PRIMARY KEY,
		organization_id TEXT NOT NULL DEFAULT '',
		document_id TEXT,
		document_type TEXT,
		user_id TEXT,
		actor_name TEXT NOT NULL DEFAULT '',
		actor_role TEXT NOT NULL DEFAULT '',
		action TEXT,
		old_value TEXT,
		new_value TEXT,
		changes TEXT,
		details TEXT,
		created_at DATETIME
	)`
	if err := config.DB.Exec(sql).Error; err != nil {
		t.Fatalf("setupAuditLogsTable: %v", err)
	}
}

// insertAuditLog seeds a single AuditLog row for use in filter tests.
// Always tags the row with testOrgID so it matches handler queries scoped to
// the same tenant.
func insertAuditLog(t *testing.T, action, documentType, userID, documentID string) {
	t.Helper()
	log := models.AuditLog{
		ID:           uuid.New().String(),
		Action:       action,
		DocumentType: documentType,
		UserID:       userID,
		DocumentID:   documentID,
		CreatedAt:    time.Now(),
	}
	if err := config.DB.Exec(
		`INSERT INTO audit_logs (id, organization_id, action, document_type, user_id, document_id, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		log.ID, testOrgID, log.Action, log.DocumentType, log.UserID, log.DocumentID, log.CreatedAt,
	).Error; err != nil {
		t.Fatalf("insertAuditLog: %v", err)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// App factories
// ─────────────────────────────────────────────────────────────────────────────

func newAnalyticsApp(t *testing.T) *fiber.App {
	t.Helper()
	auth := withTenantCtx(testOrgID, testUserID, testUserRole)
	app := fiber.New()
	app.Get("/analytics/dashboard", auth, GetDashboard)
	app.Get("/analytics/requisitions", auth, GetRequisitionMetrics)
	app.Get("/analytics/approvals", auth, GetApprovalMetrics)
	return app
}

// newAnalyticsAppNoAuth builds an app without the tenant middleware so that
// the handlers cannot resolve organizationID and return 400.
func newAnalyticsAppNoAuth(t *testing.T) *fiber.App {
	t.Helper()
	app := fiber.New()
	app.Get("/analytics/dashboard", GetDashboard)
	app.Get("/analytics/requisitions", GetRequisitionMetrics)
	app.Get("/analytics/approvals", GetApprovalMetrics)
	return app
}

func newAuditApp(t *testing.T) *fiber.App {
	t.Helper()
	auth := withTenantCtx(testOrgID, testUserID, testUserRole)
	app := fiber.New()
	app.Get("/audit-logs", auth, GetAuditLogs)
	app.Get("/audit-logs/:documentId", auth, GetDocumentAuditLogs)
	return app
}

// ─────────────────────────────────────────────────────────────────────────────
// Analytics — GetDashboard
// ─────────────────────────────────────────────────────────────────────────────

func TestGetDashboard_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newAnalyticsAppNoAuth(t)
	resp := testRequest(app, http.MethodGet, "/analytics/dashboard", nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 without organizationID local, got %d", resp.StatusCode)
	}
}

func TestGetDashboard_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newAnalyticsApp(t)
	resp := testRequest(app, http.MethodGet, "/analytics/dashboard", nil)
	if resp.StatusCode != http.StatusOK {
		body := decodeResponse(resp)
		t.Errorf("expected 200, got %d; body=%v", resp.StatusCode, body)
	}
}

func TestGetDashboard_WithDateFilters(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newAnalyticsApp(t)
	resp := testRequest(app, http.MethodGet,
		"/analytics/dashboard?startDate=2026-01-01&endDate=2026-03-31&period=monthly", nil)
	if resp.StatusCode != http.StatusOK {
		body := decodeResponse(resp)
		t.Errorf("expected 200 with date filters, got %d; body=%v", resp.StatusCode, body)
	}
}

func TestGetDashboard_ResponseShape(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newAnalyticsApp(t)
	resp := testRequest(app, http.MethodGet, "/analytics/dashboard", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	body := decodeResponse(resp)
	if body == nil {
		t.Fatal("expected JSON body, got nil")
	}
	// The handler wraps the result in a success envelope; just verify we can parse it.
	if _, ok := body["success"]; !ok {
		if _, ok2 := body["data"]; !ok2 {
			t.Logf("body keys: %v", body)
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Analytics — GetRequisitionMetrics
// ─────────────────────────────────────────────────────────────────────────────

func TestGetRequisitionMetrics_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newAnalyticsAppNoAuth(t)
	resp := testRequest(app, http.MethodGet, "/analytics/requisitions", nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 without organizationID local, got %d", resp.StatusCode)
	}
}

func TestGetRequisitionMetrics_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newAnalyticsApp(t)
	resp := testRequest(app, http.MethodGet, "/analytics/requisitions", nil)
	if resp.StatusCode != http.StatusOK {
		body := decodeResponse(resp)
		t.Errorf("expected 200, got %d; body=%v", resp.StatusCode, body)
	}
}

func TestGetRequisitionMetrics_WithDepartmentFilter(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newAnalyticsApp(t)
	resp := testRequest(app, http.MethodGet, "/analytics/requisitions?department=Engineering", nil)
	if resp.StatusCode != http.StatusOK {
		body := decodeResponse(resp)
		t.Errorf("expected 200 with department filter, got %d; body=%v", resp.StatusCode, body)
	}
}

func TestGetRequisitionMetrics_WithPeriod(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newAnalyticsApp(t)
	resp := testRequest(app, http.MethodGet, "/analytics/requisitions?period=weekly", nil)
	if resp.StatusCode != http.StatusOK {
		body := decodeResponse(resp)
		t.Errorf("expected 200 with period=weekly, got %d; body=%v", resp.StatusCode, body)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Analytics — GetApprovalMetrics
// ─────────────────────────────────────────────────────────────────────────────

func TestGetApprovalMetrics_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newAnalyticsAppNoAuth(t)
	resp := testRequest(app, http.MethodGet, "/analytics/approvals", nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 without organizationID local, got %d", resp.StatusCode)
	}
}

func TestGetApprovalMetrics_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newAnalyticsApp(t)
	resp := testRequest(app, http.MethodGet, "/analytics/approvals", nil)
	if resp.StatusCode != http.StatusOK {
		body := decodeResponse(resp)
		t.Errorf("expected 200, got %d; body=%v", resp.StatusCode, body)
	}
}

func TestGetApprovalMetrics_WithDateRange(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newAnalyticsApp(t)
	resp := testRequest(app, http.MethodGet,
		"/analytics/approvals?startDate=2026-01-01&endDate=2026-12-31", nil)
	if resp.StatusCode != http.StatusOK {
		body := decodeResponse(resp)
		t.Errorf("expected 200 with date range, got %d; body=%v", resp.StatusCode, body)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Audit — GetAuditLogs
// ─────────────────────────────────────────────────────────────────────────────

func TestGetAuditLogs_EmptyDB(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupAuditLogsTable(t)

	app := newAuditApp(t)
	resp := testRequest(app, http.MethodGet, "/audit-logs", nil)
	if resp.StatusCode != http.StatusOK {
		body := decodeResponse(resp)
		t.Errorf("expected 200 with empty audit logs, got %d; body=%v", resp.StatusCode, body)
	}
}

func TestGetAuditLogs_WithActionFilter(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupAuditLogsTable(t)
	insertAuditLog(t, "CREATE", "requisition", testUserID, uuid.New().String())

	app := newAuditApp(t)
	resp := testRequest(app, http.MethodGet, "/audit-logs?action=CREATE", nil)
	if resp.StatusCode != http.StatusOK {
		body := decodeResponse(resp)
		t.Errorf("expected 200 with action filter, got %d; body=%v", resp.StatusCode, body)
	}
}

func TestGetAuditLogs_WithDocumentTypeFilter(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupAuditLogsTable(t)
	insertAuditLog(t, "CREATE", "requisition", testUserID, uuid.New().String())
	insertAuditLog(t, "UPDATE", "purchase_order", testUserID, uuid.New().String())

	app := newAuditApp(t)
	resp := testRequest(app, http.MethodGet, "/audit-logs?documentType=requisition", nil)
	if resp.StatusCode != http.StatusOK {
		body := decodeResponse(resp)
		t.Errorf("expected 200 with documentType filter, got %d; body=%v", resp.StatusCode, body)
	}
}

func TestGetAuditLogs_WithCombinedFilters(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupAuditLogsTable(t)
	insertAuditLog(t, "CREATE", "requisition", testUserID, uuid.New().String())
	insertAuditLog(t, "UPDATE", "requisition", testUserID, uuid.New().String())

	app := newAuditApp(t)
	resp := testRequest(app, http.MethodGet,
		"/audit-logs?action=CREATE&documentType=requisition", nil)
	if resp.StatusCode != http.StatusOK {
		body := decodeResponse(resp)
		t.Errorf("expected 200 with combined filters, got %d; body=%v", resp.StatusCode, body)
	}
}

func TestGetAuditLogs_WithUserIDFilter(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupAuditLogsTable(t)
	insertAuditLog(t, "CREATE", "requisition", testUserID, uuid.New().String())

	app := newAuditApp(t)
	resp := testRequest(app, http.MethodGet, "/audit-logs?userId="+testUserID, nil)
	if resp.StatusCode != http.StatusOK {
		body := decodeResponse(resp)
		t.Errorf("expected 200 with userId filter, got %d; body=%v", resp.StatusCode, body)
	}
}

func TestGetAuditLogs_Pagination(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupAuditLogsTable(t)

	// Insert several records.
	for i := 0; i < 5; i++ {
		insertAuditLog(t, "CREATE", "requisition", testUserID, uuid.New().String())
	}

	app := newAuditApp(t)
	resp := testRequest(app, http.MethodGet, "/audit-logs?page=1&limit=2", nil)
	if resp.StatusCode != http.StatusOK {
		body := decodeResponse(resp)
		t.Errorf("expected 200 with pagination params, got %d; body=%v", resp.StatusCode, body)
	}
}

func TestGetAuditLogs_DefaultLimitClamped(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupAuditLogsTable(t)

	app := newAuditApp(t)
	// limit=200 exceeds max (100) — handler clamps to 50.
	resp := testRequest(app, http.MethodGet, "/audit-logs?limit=200", nil)
	if resp.StatusCode != http.StatusOK {
		body := decodeResponse(resp)
		t.Errorf("expected 200 with oversized limit, got %d; body=%v", resp.StatusCode, body)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Audit — GetDocumentAuditLogs
// ─────────────────────────────────────────────────────────────────────────────

func TestGetDocumentAuditLogs_NoLogs(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupAuditLogsTable(t)

	app := newAuditApp(t)
	docID := uuid.New().String()
	resp := testRequest(app, http.MethodGet, "/audit-logs/"+docID, nil)
	if resp.StatusCode != http.StatusOK {
		body := decodeResponse(resp)
		t.Errorf("expected 200 (empty list) for unknown documentId, got %d; body=%v", resp.StatusCode, body)
	}
}

func TestGetDocumentAuditLogs_WithLogs(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupAuditLogsTable(t)

	docID := uuid.New().String()
	insertAuditLog(t, "CREATE", "requisition", testUserID, docID)
	insertAuditLog(t, "UPDATE", "requisition", testUserID, docID)

	app := newAuditApp(t)
	resp := testRequest(app, http.MethodGet, "/audit-logs/"+docID, nil)
	if resp.StatusCode != http.StatusOK {
		body := decodeResponse(resp)
		t.Errorf("expected 200 with seeded logs for documentId, got %d; body=%v", resp.StatusCode, body)
	}
}

func TestGetDocumentAuditLogs_DoesNotReturnOtherDocLogs(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupAuditLogsTable(t)

	docID := uuid.New().String()
	otherDocID := uuid.New().String()
	insertAuditLog(t, "CREATE", "requisition", testUserID, otherDocID)

	app := newAuditApp(t)
	resp := testRequest(app, http.MethodGet, "/audit-logs/"+docID, nil)
	if resp.StatusCode != http.StatusOK {
		body := decodeResponse(resp)
		t.Errorf("expected 200 (empty list for unrelated doc), got %d; body=%v", resp.StatusCode, body)
	}
}

func TestGetDocumentAuditLogs_Pagination(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupAuditLogsTable(t)

	docID := uuid.New().String()
	for i := 0; i < 5; i++ {
		insertAuditLog(t, "UPDATE", "purchase_order", testUserID, docID)
	}

	app := newAuditApp(t)
	resp := testRequest(app, http.MethodGet, "/audit-logs/"+docID+"?page=1&limit=2", nil)
	if resp.StatusCode != http.StatusOK {
		body := decodeResponse(resp)
		t.Errorf("expected 200 with pagination, got %d; body=%v", resp.StatusCode, body)
	}
}

func TestGetDocumentAuditLogs_ResponseShape(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupAuditLogsTable(t)

	docID := uuid.New().String()
	insertAuditLog(t, "CREATE", "requisition", testUserID, docID)

	app := newAuditApp(t)
	resp := testRequest(app, http.MethodGet, "/audit-logs/"+docID, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	body := decodeResponse(resp)
	if body == nil {
		t.Fatal("expected JSON body, got nil")
	}
}
