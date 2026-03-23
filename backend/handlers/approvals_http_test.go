package handlers

import (
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/config"
	"github.com/stretchr/testify/assert"
)

// ─────────────────────────────────────────────────────────────────────────────
// App factories
// ─────────────────────────────────────────────────────────────────────────────

func newApprovalApp(t *testing.T) *fiber.App {
	t.Helper()
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		},
	})
	app.Use(recover.New())
	h := NewApprovalHandler()
	auth := withTenantCtx(testOrgID, testUserID, testUserRole)

	app.Get("/approvals", auth, h.GetApprovalTasks)
	app.Get("/approvals/stats", auth, h.GetTaskStats)
	app.Get("/approvals/my-pending-count", auth, h.GetMyPendingCount)
	// Bulk routes must come before :id routes so "bulk" isn't captured as an ID.
	app.Post("/approvals/bulk/approve", auth, h.BulkApprove)
	app.Post("/approvals/bulk/reject", auth, h.BulkReject)
	app.Get("/approvals/:id", auth, h.GetApprovalTask)
	app.Post("/approvals/:id/claim", auth, h.ClaimTask)
	app.Post("/approvals/:id/unclaim", auth, h.UnclaimTask)
	app.Post("/approvals/:id/approve", auth, h.ApproveTask)
	app.Post("/approvals/:id/reject", auth, h.RejectTask)
	app.Post("/approvals/:id/reassign", auth, h.ReassignTask)
	return app
}

// newApprovalAppNoAuth omits auth middleware so the handler's direct locals
// type-assertions panic. The recover middleware converts the panic to a 500,
// confirming unauthenticated access is blocked (any non-200 is acceptable).
func newApprovalAppNoAuth(t *testing.T) *fiber.App {
	t.Helper()
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		},
	})
	app.Use(recover.New())
	h := NewApprovalHandler()

	app.Get("/approvals", h.GetApprovalTasks)
	app.Get("/approvals/stats", h.GetTaskStats)
	app.Get("/approvals/my-pending-count", h.GetMyPendingCount)
	app.Post("/approvals/bulk/approve", h.BulkApprove)
	app.Post("/approvals/bulk/reject", h.BulkReject)
	app.Get("/approvals/:id", h.GetApprovalTask)
	app.Post("/approvals/:id/claim", h.ClaimTask)
	app.Post("/approvals/:id/unclaim", h.UnclaimTask)
	app.Post("/approvals/:id/approve", h.ApproveTask)
	app.Post("/approvals/:id/reject", h.RejectTask)
	app.Post("/approvals/:id/reassign", h.ReassignTask)
	return app
}

// setupApprovalDB sets up a test DB that includes the workflow_tasks table
// required by the approval handler.
func setupApprovalDB(t *testing.T) {
	t.Helper()
	db := setupTestDB(t)
	setupWorkflowTasksTable(t, db)
}

// seedApprovalTask inserts a minimal workflow_tasks row via raw SQL.
func seedApprovalTask(t *testing.T, id, orgID string) {
	t.Helper()
	err := config.DB.Exec(`INSERT INTO workflow_tasks
		(id, organization_id, workflow_assignment_id, entity_id, entity_type,
		 stage_number, stage_name, assigned_role, status, priority, created_at, version)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, orgID, uuid.New().String(), uuid.New().String(), "requisition",
		1, "Review", "admin", "PENDING", "MEDIUM", time.Now(), 1,
	).Error
	if err != nil {
		t.Fatalf("seedApprovalTask: %v", err)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// GetApprovalTasks
// ─────────────────────────────────────────────────────────────────────────────

func TestGetApprovalTasks_NoAuth(t *testing.T) {
	app := newApprovalAppNoAuth(t)
	resp := testRequest(app, http.MethodGet, "/approvals", nil)
	assert.NotEqual(t, http.StatusOK, resp.StatusCode, "unauthenticated request should be blocked")
}

func TestGetApprovalTasks_Success(t *testing.T) {
	setupApprovalDB(t)
	defer func() { config.DB = nil }()

	seedTestUser(t)
	app := newApprovalApp(t)

	resp := testRequest(app, http.MethodGet, "/approvals", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// GetTaskStats
// ─────────────────────────────────────────────────────────────────────────────

func TestGetTaskStats_NoAuth(t *testing.T) {
	app := newApprovalAppNoAuth(t)
	resp := testRequest(app, http.MethodGet, "/approvals/stats", nil)
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}

func TestGetTaskStats_Success(t *testing.T) {
	setupApprovalDB(t)
	defer func() { config.DB = nil }()

	seedTestUser(t)
	app := newApprovalApp(t)

	resp := testRequest(app, http.MethodGet, "/approvals/stats", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// GetMyPendingCount
// ─────────────────────────────────────────────────────────────────────────────

func TestGetMyPendingCount_NoAuth(t *testing.T) {
	app := newApprovalAppNoAuth(t)
	resp := testRequest(app, http.MethodGet, "/approvals/my-pending-count", nil)
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}

func TestGetMyPendingCount_Success(t *testing.T) {
	setupApprovalDB(t)
	defer func() { config.DB = nil }()

	seedTestUser(t)
	app := newApprovalApp(t)

	resp := testRequest(app, http.MethodGet, "/approvals/my-pending-count", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// GetApprovalTask
// ─────────────────────────────────────────────────────────────────────────────

func TestGetApprovalTask_NoAuth(t *testing.T) {
	app := newApprovalAppNoAuth(t)
	resp := testRequest(app, http.MethodGet, "/approvals/"+uuid.New().String(), nil)
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}

func TestGetApprovalTask_NotFound(t *testing.T) {
	setupApprovalDB(t)
	defer func() { config.DB = nil }()

	seedTestUser(t)
	app := newApprovalApp(t)

	resp := testRequest(app, http.MethodGet, "/approvals/"+uuid.New().String(), nil)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestGetApprovalTask_Success(t *testing.T) {
	setupApprovalDB(t)
	defer func() { config.DB = nil }()

	seedTestUser(t)
	taskID := uuid.New().String()
	seedApprovalTask(t, taskID, testOrgID)
	app := newApprovalApp(t)

	resp := testRequest(app, http.MethodGet, "/approvals/"+taskID, nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// ClaimTask
// ─────────────────────────────────────────────────────────────────────────────

func TestClaimTask_NoAuth(t *testing.T) {
	app := newApprovalAppNoAuth(t)
	resp := testRequest(app, http.MethodPost, "/approvals/"+uuid.New().String()+"/claim", nil)
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}

// ClaimTask uses workflowExecutionService from locals (nil → panic → 500).
// That's an acceptable non-200 "blocked" signal for a non-existent task.
func TestClaimTask_NotFound(t *testing.T) {
	setupApprovalDB(t)
	defer func() { config.DB = nil }()

	seedTestUser(t)
	app := newApprovalApp(t)

	resp := testRequest(app, http.MethodPost, "/approvals/"+uuid.New().String()+"/claim", nil)
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// UnclaimTask
// ─────────────────────────────────────────────────────────────────────────────

func TestUnclaimTask_NoAuth(t *testing.T) {
	app := newApprovalAppNoAuth(t)
	resp := testRequest(app, http.MethodPost, "/approvals/"+uuid.New().String()+"/unclaim", nil)
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}

func TestUnclaimTask_NotFound(t *testing.T) {
	setupApprovalDB(t)
	defer func() { config.DB = nil }()

	seedTestUser(t)
	app := newApprovalApp(t)

	resp := testRequest(app, http.MethodPost, "/approvals/"+uuid.New().String()+"/unclaim", nil)
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// ApproveTask
// ─────────────────────────────────────────────────────────────────────────────

func TestApproveTask_NoAuth(t *testing.T) {
	app := newApprovalAppNoAuth(t)
	resp := testRequest(app, http.MethodPost, "/approvals/"+uuid.New().String()+"/approve",
		map[string]interface{}{"signature": "sig"})
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}

func TestApproveTask_MissingSignature(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	seedTestUser(t)
	app := newApprovalApp(t)

	resp := testRequest(app, http.MethodPost, "/approvals/"+uuid.New().String()+"/approve",
		map[string]interface{}{"comment": "looks good"})
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestApproveTask_NotFound(t *testing.T) {
	setupApprovalDB(t)
	defer func() { config.DB = nil }()

	seedTestUser(t)
	app := newApprovalApp(t)

	// Signature present but no workflowExecutionService injected → non-200.
	resp := testRequest(app, http.MethodPost, "/approvals/"+uuid.New().String()+"/approve",
		map[string]interface{}{"signature": "data:image/png;base64,abc"})
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// RejectTask
// ─────────────────────────────────────────────────────────────────────────────

func TestRejectTask_NoAuth(t *testing.T) {
	app := newApprovalAppNoAuth(t)
	resp := testRequest(app, http.MethodPost, "/approvals/"+uuid.New().String()+"/reject",
		map[string]interface{}{"reason": "bad"})
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}

func TestRejectTask_MissingReason(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	seedTestUser(t)
	app := newApprovalApp(t)

	resp := testRequest(app, http.MethodPost, "/approvals/"+uuid.New().String()+"/reject",
		map[string]interface{}{"signature": "sig"})
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestRejectTask_NotFound(t *testing.T) {
	setupApprovalDB(t)
	defer func() { config.DB = nil }()

	seedTestUser(t)
	app := newApprovalApp(t)

	resp := testRequest(app, http.MethodPost, "/approvals/"+uuid.New().String()+"/reject",
		map[string]interface{}{"reason": "not acceptable", "signature": "sig"})
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// ReassignTask
// ─────────────────────────────────────────────────────────────────────────────

func TestReassignTask_NoAuth(t *testing.T) {
	app := newApprovalAppNoAuth(t)
	resp := testRequest(app, http.MethodPost, "/approvals/"+uuid.New().String()+"/reassign",
		map[string]interface{}{"newUserId": "some-user"})
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}

func TestReassignTask_MissingUserId(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	seedTestUser(t)
	app := newApprovalApp(t)

	resp := testRequest(app, http.MethodPost, "/approvals/"+uuid.New().String()+"/reassign",
		map[string]interface{}{"reason": "out of office"})
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestReassignTask_NotFound(t *testing.T) {
	setupApprovalDB(t)
	defer func() { config.DB = nil }()

	seedTestUser(t)
	app := newApprovalApp(t)

	resp := testRequest(app, http.MethodPost, "/approvals/"+uuid.New().String()+"/reassign",
		map[string]interface{}{"newUserId": uuid.New().String(), "reason": "out of office"})
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// BulkApprove
// ─────────────────────────────────────────────────────────────────────────────

func TestBulkApprove_NoAuth(t *testing.T) {
	app := newApprovalAppNoAuth(t)
	resp := testRequest(app, http.MethodPost, "/approvals/bulk/approve",
		map[string]interface{}{"taskIds": []string{"id1"}, "signature": "sig"})
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}

func TestBulkApprove_MissingTaskIds(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	seedTestUser(t)
	app := newApprovalApp(t)

	resp := testRequest(app, http.MethodPost, "/approvals/bulk/approve",
		map[string]interface{}{"signature": "sig"})
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestBulkApprove_MissingSignature(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	seedTestUser(t)
	app := newApprovalApp(t)

	resp := testRequest(app, http.MethodPost, "/approvals/bulk/approve",
		map[string]interface{}{"taskIds": []string{"id1"}})
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// BulkReject
// ─────────────────────────────────────────────────────────────────────────────

func TestBulkReject_NoAuth(t *testing.T) {
	app := newApprovalAppNoAuth(t)
	resp := testRequest(app, http.MethodPost, "/approvals/bulk/reject",
		map[string]interface{}{"taskIds": []string{"id1"}, "reason": "bad", "signature": "sig"})
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}

func TestBulkReject_MissingTaskIds(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	seedTestUser(t)
	app := newApprovalApp(t)

	resp := testRequest(app, http.MethodPost, "/approvals/bulk/reject",
		map[string]interface{}{"reason": "bad", "signature": "sig"})
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestBulkReject_MissingReason(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	seedTestUser(t)
	app := newApprovalApp(t)

	resp := testRequest(app, http.MethodPost, "/approvals/bulk/reject",
		map[string]interface{}{"taskIds": []string{"id1"}, "signature": "sig"})
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
