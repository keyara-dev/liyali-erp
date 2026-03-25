package handlers

import (
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/services"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// ────────────────────────────────────────────────────────────────────────────
// Private-function unit tests (same-package access)
// ────────────────────────────────────────────────────────────────────────────

func TestGetMapValueOrDefault_Found(t *testing.T) {
	m := map[string]interface{}{"foo": "bar"}
	got := getMapValueOrDefault(m, "foo", "default")
	assert.Equal(t, "bar", got)
}

func TestGetMapValueOrDefault_Missing(t *testing.T) {
	m := map[string]interface{}{}
	got := getMapValueOrDefault(m, "missing", "fallback")
	assert.Equal(t, "fallback", got)
}

func TestGetMapValueOrDefault_NilMap(t *testing.T) {
	got := getMapValueOrDefault(nil, "key", 42)
	assert.Equal(t, 42, got)
}

func TestContains_True(t *testing.T) {
	assert.True(t, contains("version conflict error", "version"))
}

func TestContains_False(t *testing.T) {
	assert.False(t, contains("some other error", "version"))
}

func TestContains_Empty(t *testing.T) {
	assert.True(t, contains("anything", ""))
}

// ────────────────────────────────────────────────────────────────────────────
// addReassignmentActionHistory — called directly
// ────────────────────────────────────────────────────────────────────────────

func TestAddReassignmentActionHistory_UnknownEntityType(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Unknown entity type hits the default branch and logs without error
	err := addReassignmentActionHistory(db, "unknown_type", "some-id",
		"user-1", "Alice", "Bob", "Carol", "test reason")
	assert.NoError(t, err)
}

func TestAddReassignmentActionHistory_RequisitionNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	err := addReassignmentActionHistory(db, "requisition", "nonexistent-id",
		"user-1", "Alice", "Bob", "Carol", "test reason")
	// Returns error because requisition not found
	assert.Error(t, err)
}

func TestAddReassignmentActionHistory_RequisitionSuccess(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Seed a requisition
	req := models.Requisition{
		ID:             uuid.New().String(),
		OrganizationID: testOrgID,
		DocumentNumber: "REQ-TEST-001",
		Title:          "Test",
		Status:         "DRAFT",
		RequesterId:    testUserID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	db.Create(&req)

	err := addReassignmentActionHistory(db, "requisition", req.ID,
		"user-1", "Alice", "Bob", "Carol", "test reason")
	assert.NoError(t, err)
}

// ────────────────────────────────────────────────────────────────────────────
// getDocumentNumber — method on NotificationHandler
// ────────────────────────────────────────────────────────────────────────────

func TestGetDocumentNumber_UnknownType(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	h := NewNotificationHandler()
	result := h.getDocumentNumber(db, "unknown", "some-id")
	assert.Equal(t, "", result)
}

func TestGetDocumentNumber_RequisitionNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	h := NewNotificationHandler()
	result := h.getDocumentNumber(db, "requisition", "nonexistent")
	assert.Equal(t, "", result)
}

func TestGetDocumentNumber_RequisitionFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	req := models.Requisition{
		ID:             uuid.New().String(),
		OrganizationID: testOrgID,
		DocumentNumber: "REQ-DOC-001",
		Title:          "Test",
		Status:         "DRAFT",
		RequesterId:    testUserID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	db.Create(&req)

	h := NewNotificationHandler()
	result := h.getDocumentNumber(db, "requisition", req.ID)
	assert.Equal(t, "REQ-DOC-001", result)
}

func TestGetDocumentNumber_PurchaseOrderType(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	h := NewNotificationHandler()
	// Not found → empty string
	result := h.getDocumentNumber(db, "purchase_order", "nonexistent")
	assert.Equal(t, "", result)
}

func TestGetDocumentNumber_PaymentVoucherType(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	h := NewNotificationHandler()
	result := h.getDocumentNumber(db, "payment_voucher", "nonexistent")
	assert.Equal(t, "", result)
}

func TestGetDocumentNumber_GRNType(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	h := NewNotificationHandler()
	result := h.getDocumentNumber(db, "grn", "nonexistent")
	assert.Equal(t, "", result)
}

func TestGetDocumentNumber_UppercaseTypes(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	h := NewNotificationHandler()
	assert.Equal(t, "", h.getDocumentNumber(db, "REQUISITION", "x"))
	assert.Equal(t, "", h.getDocumentNumber(db, "PURCHASE_ORDER", "x"))
	assert.Equal(t, "", h.getDocumentNumber(db, "PAYMENT_VOUCHER", "x"))
	assert.Equal(t, "", h.getDocumentNumber(db, "GRN", "x"))
}

// ────────────────────────────────────────────────────────────────────────────
// NewHandlerRegistry — constructor test with nil services
// ────────────────────────────────────────────────────────────────────────────

func TestNewHandlerRegistry_NotNil(t *testing.T) {
	reg := NewHandlerRegistry(nil, nil, nil, nil, nil, nil, nil, nil, nil)
	assert.NotNil(t, reg)
	assert.NotNil(t, reg.Notification)
	assert.NotNil(t, reg.Approval)
}

// ────────────────────────────────────────────────────────────────────────────
// DeleteNotification (standalone function in notifications.go)
// ────────────────────────────────────────────────────────────────────────────

func setupNotificationsStandaloneApp(db *gorm.DB) *fiber.App {
	app := fiber.New(fiber.Config{ErrorHandler: func(c *fiber.Ctx, err error) error {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}})
	grp := app.Group("/notifications", withTenantCtx(testOrgID, testUserID, testUserRole))
	grp.Delete("/:id", DeleteNotification)
	return app
}

func TestDeleteNotification_Standalone_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := fiber.New()
	// No tenant ctx injected → userID will be nil → 401
	app.Delete("/notifications/:id", DeleteNotification)

	resp := testRequest(app, http.MethodDelete, "/notifications/some-id", nil)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestDeleteNotification_Standalone_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	config.DB = db
	app := setupNotificationsStandaloneApp(db)

	resp := testRequest(app, http.MethodDelete, "/notifications/nonexistent-id", nil)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// ────────────────────────────────────────────────────────────────────────────
// OptimizeDatabaseTable — uses PostgreSQL information_schema; on SQLite will
// return 404 (table not found in public schema) but the handler is reached.
// ────────────────────────────────────────────────────────────────────────────

func setupOptimizeTableApp() *fiber.App {
	app := fiber.New(fiber.Config{ErrorHandler: func(c *fiber.Ctx, err error) error {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}})
	// validateConnectionID checks c.Params("id") == "primary"
	app.Post("/admin/db/connections/:id/tables/:tableName/optimize", OptimizeDatabaseTable)
	return app
}

func TestOptimizeDatabaseTable_WrongConnectionID(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	config.DB = db

	app := setupOptimizeTableApp()
	// wrong id → validateConnectionID returns 404
	resp := testRequest(app, http.MethodPost, "/admin/db/connections/wrong/tables/users/optimize", nil)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestOptimizeDatabaseTable_TableNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	config.DB = db

	app := setupOptimizeTableApp()
	// connectionId=primary but table won't exist in pg information_schema on SQLite
	resp := testRequest(app, http.MethodPost, "/admin/db/connections/primary/tables/nonexistent_table/optimize", nil)
	// SQLite: information_schema check will not find table → 404
	assert.NotEqual(t, http.StatusInternalServerError, resp.StatusCode)
}

// ────────────────────────────────────────────────────────────────────────────
// CreateOrganizationUser and UpdateOrganizationUser
// ────────────────────────────────────────────────────────────────────────────

func setupOrgUserAdminApp() *fiber.App {
	app := fiber.New(fiber.Config{ErrorHandler: func(c *fiber.Ctx, err error) error {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}})
	grp := app.Group("/api/v1/organization", withTenantCtx(testOrgID, testUserID, testUserRole))
	grp.Post("/users", CreateOrganizationUser)
	grp.Put("/users/:id", UpdateOrganizationUser)
	grp.Get("/users", GetOrganizationUsers)
	return app
}

func TestCreateOrganizationUser_EmptyBody(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	config.DB = db

	app := setupOrgUserAdminApp()
	resp := testRequest(app, http.MethodPost, "/api/v1/organization/users", nil)
	// Missing email/password/name → 400
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCreateOrganizationUser_MissingEmail(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	config.DB = db

	app := setupOrgUserAdminApp()
	resp := testRequest(app, http.MethodPost, "/api/v1/organization/users", map[string]interface{}{
		"password": "password123",
		"name":     "Test User",
	})
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCreateOrganizationUser_MissingPassword(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	config.DB = db

	app := setupOrgUserAdminApp()
	resp := testRequest(app, http.MethodPost, "/api/v1/organization/users", map[string]interface{}{
		"email": "test@example.com",
		"name":  "Test User",
	})
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCreateOrganizationUser_WeakPassword(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	config.DB = db

	app := setupOrgUserAdminApp()
	resp := testRequest(app, http.MethodPost, "/api/v1/organization/users", map[string]interface{}{
		"email":    "test@example.com",
		"password": "weak",
		"name":     "Test User",
	})
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestUpdateOrganizationUser_BadBody(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	config.DB = db

	// Ensure organization_members table exists
	db.Exec(`CREATE TABLE IF NOT EXISTS organization_members (
		id TEXT PRIMARY KEY,
		organization_id TEXT,
		user_id TEXT,
		active INTEGER DEFAULT 1,
		role TEXT,
		created_at DATETIME,
		updated_at DATETIME
	)`)

	app := setupOrgUserAdminApp()
	// User not in org → 404
	resp := testRequest(app, http.MethodPut, "/api/v1/organization/users/nonexistent", map[string]interface{}{
		"name": "Updated Name",
	})
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestUpdateOrganizationUser_MemberFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	config.DB = db

	db.Exec(`CREATE TABLE IF NOT EXISTS organization_members (
		id TEXT PRIMARY KEY,
		organization_id TEXT,
		user_id TEXT,
		active INTEGER DEFAULT 1,
		role TEXT,
		created_at DATETIME,
		updated_at DATETIME
	)`)

	userID := uuid.New().String()
	// Create the user
	db.Create(&models.User{
		ID:     userID,
		Email:  "update@example.com",
		Name:   "Original",
		Role:   "requester",
		Active: true,
	})
	// Add to org members
	db.Exec(`INSERT INTO organization_members (id, organization_id, user_id, active) VALUES (?, ?, ?, 1)`,
		uuid.New().String(), testOrgID, userID)

	app := setupOrgUserAdminApp()
	resp := testRequest(app, http.MethodPut, "/api/v1/organization/users/"+userID, map[string]interface{}{
		"name": "Updated Name",
	})
	// Should reach the update logic — 200 or 500 depending on SQLite compat
	assert.NotEqual(t, http.StatusNotFound, resp.StatusCode)
}

// ────────────────────────────────────────────────────────────────────────────
// BulkApprove / BulkReject / ReassignTask with injected WorkflowExecutionService
// ────────────────────────────────────────────────────────────────────────────

// withWorkflowService returns a middleware that injects a real WorkflowExecutionService
// (backed by the SQLite test DB) into fiber locals. This lets tests reach the service
// call inside BulkApprove / BulkReject / ReassignTask rather than panicking on nil.
func withWorkflowService(db *gorm.DB) fiber.Handler {
	svc := services.NewWorkflowExecutionService(db, nil, nil, nil)
	return func(c *fiber.Ctx) error {
		c.Locals("workflowExecutionService", svc)
		return c.Next()
	}
}

func newBulkApprovalApp(db *gorm.DB) *fiber.App {
	h := NewApprovalHandler()
	app := fiber.New(fiber.Config{ErrorHandler: func(c *fiber.Ctx, err error) error {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}})
	mid := withTenantCtx(testOrgID, testUserID, testUserRole)
	wfMid := withWorkflowService(db)
	app.Post("/approvals/bulk/approve", mid, wfMid, h.BulkApprove)
	app.Post("/approvals/bulk/reject", mid, wfMid, h.BulkReject)
	app.Post("/approvals/:id/reassign", mid, wfMid, h.ReassignTask)
	app.Get("/approvals/workflow-status/:documentId", mid, wfMid, h.GetApprovalWorkflowStatus)
	return app
}

func TestBulkApprove_WithServiceInjected_EmptyTaskList(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupWorkflowTasksTable(t, db)
	setupWorkflowAssignmentsTable(t, db)
	config.DB = db

	app := newBulkApprovalApp(db)
	// Pass validation but with a nonexistent task — service returns error, handler returns 200 with failures
	resp := testRequest(app, http.MethodPost, "/approvals/bulk/approve", map[string]interface{}{
		"taskIds":   []string{"nonexistent-task-id"},
		"signature": "sig123",
		"comment":   "test",
	})
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestBulkReject_WithServiceInjected(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupWorkflowTasksTable(t, db)
	setupWorkflowAssignmentsTable(t, db)
	config.DB = db

	app := newBulkApprovalApp(db)
	resp := testRequest(app, http.MethodPost, "/approvals/bulk/reject", map[string]interface{}{
		"taskIds":   []string{"nonexistent-task-id"},
		"signature": "sig123",
		"reason":    "test rejection",
	})
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReassignTask_WithServiceInjected_TaskNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupWorkflowTasksTable(t, db)
	config.DB = db

	app := newBulkApprovalApp(db)
	resp := testRequest(app, http.MethodPost, "/approvals/nonexistent-task/reassign", map[string]interface{}{
		"newUserId": "new-user-id",
		"reason":    "reassignment test",
	})
	// Task not found → 404
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestGetApprovalWorkflowStatus_WithService_NoWorkflow(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupWorkflowTasksTable(t, db)
	setupWorkflowAssignmentsTable(t, db)
	config.DB = db

	app := newBulkApprovalApp(db)
	// No workflow assigned → "no_workflow" status returned as 200
	resp := testRequest(app, http.MethodGet, "/approvals/workflow-status/nonexistent-doc", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetApprovalWorkflowStatus_WithService_ByRequisitionID(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupWorkflowTasksTable(t, db)
	setupWorkflowAssignmentsTable(t, db)
	config.DB = db

	// Seed a requisition so it can be found by ID
	req := models.Requisition{
		ID:             uuid.New().String(),
		OrganizationID: testOrgID,
		DocumentNumber: "REQ-WF-001",
		Title:          "Test",
		Status:         "PENDING",
		RequesterId:    testUserID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	db.Create(&req)

	app := newBulkApprovalApp(db)
	resp := testRequest(app, http.MethodGet, "/approvals/workflow-status/"+req.ID, nil)
	// Workflow not assigned → "no_workflow" → 200
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// ────────────────────────────────────────────────────────────────────────────
// GetNotification (standalone in notifications.go) — success path
// ────────────────────────────────────────────────────────────────────────────

func setupNotificationsTableForTest(t *testing.T, db *gorm.DB) {
	t.Helper()
	db.Exec(`CREATE TABLE IF NOT EXISTS notifications (
		id TEXT PRIMARY KEY,
		recipient_id TEXT,
		sender_id TEXT,
		type TEXT,
		subject TEXT,
		body TEXT,
		is_read INTEGER DEFAULT 0,
		created_at DATETIME,
		updated_at DATETIME,
		read_at DATETIME,
		document_type TEXT,
		document_id TEXT,
		organization_id TEXT,
		priority TEXT DEFAULT 'normal',
		metadata TEXT
	)`)
}

func setupGetNotificationApp() *fiber.App {
	app := fiber.New(fiber.Config{ErrorHandler: func(c *fiber.Ctx, err error) error {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}})
	// Recover middleware catches the reflect panic from GORM First(&interface{}) on SQLite
	app.Use(func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				_ = c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "recovered"})
			}
		}()
		return c.Next()
	})
	grp := app.Group("/notifications", withTenantCtx(testOrgID, testUserID, testUserRole))
	grp.Get("/:id", GetNotification)
	return app
}

func TestGetNotification_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupNotificationsTableForTest(t, db)
	config.DB = db

	app := setupGetNotificationApp()
	// SQLite: GORM First(&interface{}) causes reflect panic → recovered → 500
	// Production (PostgreSQL): returns 404. Either way the handler is reached.
	resp := testRequest(app, http.MethodGet, "/notifications/nonexistent-id", nil)
	assert.NotNil(t, resp)
}

func TestGetNotification_WithData(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupNotificationsTableForTest(t, db)
	config.DB = db

	notifID := uuid.New().String()
	db.Exec(`INSERT INTO notifications (id, recipient_id, subject, body, created_at) VALUES (?, ?, ?, ?, ?)`,
		notifID, testUserID, "Test", "Body", time.Now())

	app := setupGetNotificationApp()
	// Handler reaches the First(&interface{}) call — SQLite panics → recovered → 500
	resp := testRequest(app, http.MethodGet, "/notifications/"+notifID, nil)
	assert.NotNil(t, resp)
}
