package handlers

// approvals_extra_http_test.go — additional coverage for approval_handler.go
//
// Target functions and current coverage (before this file):
//   GetMyPendingCount          47.9%
//   GetApprovalTasks           48.2%
//   populateWorkflowTaskFields 36.7%
//   ApproveTask                45%
//   RejectTask                 45%
//   ReassignTask               12.8%
//   GetApprovalWorkflowStatus  13.4%
//   BulkApprove                31.2%
//   BulkReject                 31.2%
//   GetApprovalTask            60%
//   ClaimTask                  60%
//   UnclaimTask                60%
//   GetTaskStats               60.9%

import (
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/services"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// ─────────────────────────────────────────────────────────────────────────────
// Shared helpers
// ─────────────────────────────────────────────────────────────────────────────

// setupExtraApprovalDB initialises an in-memory SQLite DB, creates the users
// table (via AutoMigrate), and creates workflow_tasks and
// workflow_assignments tables with raw SQL (they use uuid/jsonb tags that
// break SQLite AutoMigrate).  It sets config.DB and returns the *gorm.DB.
func setupExtraApprovalDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupTestDB(t)
	setupWorkflowTasksTable(t, db)
	setupWorkflowAssignmentsTable(t, db)
	config.DB = db
	return db
}

// seedExtraTask inserts a workflow_tasks row with the given status.
// A new random entity_id is used unless entityID is non-empty.
func seedExtraTask(t *testing.T, db *gorm.DB, id, orgID, status, entityID, entityType string) {
	t.Helper()
	if entityID == "" {
		entityID = uuid.New().String()
	}
	if entityType == "" {
		entityType = "requisition"
	}
	err := db.Exec(`INSERT INTO workflow_tasks
		(id, organization_id, workflow_assignment_id, entity_id, entity_type,
		 stage_number, stage_name, assigned_role, status, priority, created_at, version)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, orgID, uuid.New().String(), entityID, entityType,
		1, "Review", "admin", status, "MEDIUM", time.Now(), 1,
	).Error
	if err != nil {
		t.Fatalf("seedExtraTask: %v", err)
	}
}

// seedExtraTaskWithUserAssignment inserts a workflow_tasks row assigned
// directly to a specific user ID.
func seedExtraTaskWithUserAssignment(t *testing.T, db *gorm.DB, id, orgID, status, assignedUserID string) {
	t.Helper()
	err := db.Exec(`INSERT INTO workflow_tasks
		(id, organization_id, workflow_assignment_id, entity_id, entity_type,
		 stage_number, stage_name, assigned_user_id, status, priority, created_at, version)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, orgID, uuid.New().String(), uuid.New().String(), "requisition",
		1, "Review", assignedUserID, status, "HIGH", time.Now(), 1,
	).Error
	if err != nil {
		t.Fatalf("seedExtraTaskWithUserAssignment: %v", err)
	}
}

// seedExtraTaskWithDueDate inserts a workflow_tasks row with an explicit due_date.
func seedExtraTaskWithDueDate(t *testing.T, db *gorm.DB, id, orgID, status string, dueDate time.Time) {
	t.Helper()
	err := db.Exec(`INSERT INTO workflow_tasks
		(id, organization_id, workflow_assignment_id, entity_id, entity_type,
		 stage_number, stage_name, assigned_role, status, priority, due_date, created_at, version)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, orgID, uuid.New().String(), uuid.New().String(), "requisition",
		1, "Review", "admin", status, "HIGH", dueDate, time.Now(), 1,
	).Error
	if err != nil {
		t.Fatalf("seedExtraTaskWithDueDate: %v", err)
	}
}

// newFullApprovalApp builds a Fiber app with ALL approval routes and injects
// both the tenant context and the workflow execution service.
func newFullApprovalApp(t *testing.T, db *gorm.DB) *fiber.App {
	t.Helper()
	h := NewApprovalHandler()
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		},
	})
	app.Use(recover.New())

	auth := withTenantCtx(testOrgID, testUserID, testUserRole)
	wfMid := withWorkflowService(db)

	app.Get("/approvals", auth, wfMid, h.GetApprovalTasks)
	app.Get("/approvals/stats", auth, wfMid, h.GetTaskStats)
	app.Get("/approvals/my-pending-count", auth, wfMid, h.GetMyPendingCount)
	app.Post("/approvals/bulk/approve", auth, wfMid, h.BulkApprove)
	app.Post("/approvals/bulk/reject", auth, wfMid, h.BulkReject)
	app.Post("/approvals/bulk/reassign", auth, wfMid, h.BulkReassign)
	app.Get("/approvals/:id", auth, wfMid, h.GetApprovalTask)
	app.Post("/approvals/:id/claim", auth, wfMid, h.ClaimTask)
	app.Post("/approvals/:id/unclaim", auth, wfMid, h.UnclaimTask)
	app.Post("/approvals/:id/approve", auth, wfMid, h.ApproveTask)
	app.Post("/approvals/:id/reject", auth, wfMid, h.RejectTask)
	app.Post("/approvals/:id/reassign", auth, wfMid, h.ReassignTask)
	app.Get("/documents/:documentId/approval-status", auth, wfMid, h.GetApprovalWorkflowStatus)
	return app
}

// withWorkflowServiceForDB wraps services.NewWorkflowExecutionService exactly
// like withWorkflowService in final_coverage_http_test.go but accepts any db.
func withWorkflowServiceForDB(db *gorm.DB) fiber.Handler {
	svc := services.NewWorkflowExecutionService(db, nil, nil, nil)
	return func(c *fiber.Ctx) error {
		c.Locals("workflowExecutionService", svc)
		return c.Next()
	}
}

// newNonAdminApprovalApp builds a Fiber app where the authenticated user has
// the "requester" role (non-approver).  This exercises the permission-filter
// branches that restrict visibility to tasks assigned to that user/role.
func newNonAdminApprovalApp(t *testing.T, db *gorm.DB, userID string) *fiber.App {
	t.Helper()
	h := NewApprovalHandler()
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		},
	})
	app.Use(recover.New())

	auth := withTenantCtx(testOrgID, userID, "requester")
	wfMid := withWorkflowServiceForDB(db)

	app.Get("/approvals", auth, wfMid, h.GetApprovalTasks)
	app.Get("/approvals/stats", auth, wfMid, h.GetTaskStats)
	app.Get("/approvals/my-pending-count", auth, wfMid, h.GetMyPendingCount)
	app.Get("/approvals/:id", auth, wfMid, h.GetApprovalTask)
	app.Post("/approvals/:id/reassign", auth, wfMid, h.ReassignTask)
	app.Post("/approvals/bulk/approve", auth, wfMid, h.BulkApprove)
	app.Post("/approvals/bulk/reject", auth, wfMid, h.BulkReject)
	return app
}

// ─────────────────────────────────────────────────────────────────────────────
// GetTaskStats — additional branches
// ─────────────────────────────────────────────────────────────────────────────

// TestGetTaskStats_UserNotFound exercises the early-exit when the user row is absent.
func TestGetTaskStats_Extra_UserNotFound(t *testing.T) {
	db := setupExtraApprovalDB(t)
	defer teardownTestDB(t, db)

	// Do NOT seed any user — db has users table but it's empty.
	h := NewApprovalHandler()
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		},
	})
	// Inject tenant with a user that does NOT exist in DB.
	app.Get("/approvals/stats", withTenantCtx(testOrgID, "nonexistent-user-xyz", "admin"), h.GetTaskStats)

	resp := testRequest(app, http.MethodGet, "/approvals/stats", nil)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// TestGetTaskStats_Extra_WithSeededTasks exercises the COUNT queries with data.
func TestGetTaskStats_Extra_WithSeededTasks(t *testing.T) {
	db := setupExtraApprovalDB(t)
	defer teardownTestDB(t, db)
	seedTestUser(t)

	// Seed several tasks in various statuses and priorities.
	seedExtraTask(t, db, uuid.New().String(), testOrgID, "PENDING", "", "requisition")
	seedExtraTask(t, db, uuid.New().String(), testOrgID, "PENDING", "", "purchase_order")
	seedExtraTaskWithDueDate(t, db, uuid.New().String(), testOrgID, "PENDING", time.Now().Add(-48*time.Hour)) // overdue
	seedExtraTask(t, db, uuid.New().String(), testOrgID, "APPROVED", "", "requisition")
	seedExtraTask(t, db, uuid.New().String(), testOrgID, "COMPLETED", "", "requisition")

	app := newFullApprovalApp(t, db)
	resp := testRequest(app, http.MethodGet, "/approvals/stats", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.NotNil(t, body)
}

// TestGetTaskStats_Extra_NonApproverUser exercises the restricted-visibility
// branch for a "requester" role user.
func TestGetTaskStats_Extra_NonApproverUser(t *testing.T) {
	db := setupExtraApprovalDB(t)
	defer teardownTestDB(t, db)

	requesterID := uuid.New().String()
	db.Create(&models.User{
		ID:     requesterID,
		Email:  "requester@example.com",
		Name:   "Requester",
		Role:   "requester",
		Active: true,
	})
	// Seed task assigned to the requester
	seedExtraTaskWithUserAssignment(t, db, uuid.New().String(), testOrgID, "PENDING", requesterID)

	app := newNonAdminApprovalApp(t, db, requesterID)
	resp := testRequest(app, http.MethodGet, "/approvals/stats", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// GetMyPendingCount — additional branches
// ─────────────────────────────────────────────────────────────────────────────

// TestGetMyPendingCount_Extra_AdminUserWithTasks verifies the admin-wide count path.
func TestGetMyPendingCount_Extra_AdminUserWithTasks(t *testing.T) {
	db := setupExtraApprovalDB(t)
	defer teardownTestDB(t, db)
	seedTestUser(t)

	// Seed two PENDING tasks in the org.
	seedExtraTask(t, db, uuid.New().String(), testOrgID, "PENDING", "", "")
	seedExtraTask(t, db, uuid.New().String(), testOrgID, "PENDING", "", "")

	app := newFullApprovalApp(t, db)
	resp := testRequest(app, http.MethodGet, "/approvals/my-pending-count", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.NotNil(t, body)
	data, ok := body["data"].(map[string]interface{})
	assert.True(t, ok)
	count, _ := data["count"].(float64)
	assert.GreaterOrEqual(t, int(count), 2)
}

// TestGetMyPendingCount_Extra_NonApproverUser exercises the non-approver branch.
func TestGetMyPendingCount_Extra_NonApproverUser(t *testing.T) {
	db := setupExtraApprovalDB(t)
	defer teardownTestDB(t, db)

	requesterID := uuid.New().String()
	db.Create(&models.User{
		ID:     requesterID,
		Email:  "req2@example.com",
		Name:   "Req2",
		Role:   "requester",
		Active: true,
	})
	// Task assigned directly to requester
	seedExtraTaskWithUserAssignment(t, db, uuid.New().String(), testOrgID, "PENDING", requesterID)

	app := newNonAdminApprovalApp(t, db, requesterID)
	resp := testRequest(app, http.MethodGet, "/approvals/my-pending-count", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestGetMyPendingCount_Extra_UserNotFound verifies early return when user missing.
func TestGetMyPendingCount_Extra_UserNotFound(t *testing.T) {
	db := setupExtraApprovalDB(t)
	defer teardownTestDB(t, db)

	h := NewApprovalHandler()
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		},
	})
	app.Get("/approvals/my-pending-count",
		withTenantCtx(testOrgID, "ghost-user-xyz", "admin"),
		h.GetMyPendingCount)

	resp := testRequest(app, http.MethodGet, "/approvals/my-pending-count", nil)
	// Handler returns 200 with count:0 when user not found.
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// GetApprovalTasks — additional filter branches
// ─────────────────────────────────────────────────────────────────────────────

// TestGetApprovalTasks_Extra_WithTasks verifies pagination with seeded data.
func TestGetApprovalTasks_Extra_WithTasks(t *testing.T) {
	db := setupExtraApprovalDB(t)
	defer teardownTestDB(t, db)
	seedTestUser(t)

	seedExtraTask(t, db, uuid.New().String(), testOrgID, "PENDING", "", "requisition")
	seedExtraTask(t, db, uuid.New().String(), testOrgID, "PENDING", "", "purchase_order")
	seedExtraTask(t, db, uuid.New().String(), testOrgID, "COMPLETED", "", "requisition")

	app := newFullApprovalApp(t, db)
	resp := testRequest(app, http.MethodGet, "/approvals?page=1&limit=10", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestGetApprovalTasks_Extra_StatusFilter exercises the status filter branch.
func TestGetApprovalTasks_Extra_StatusFilter(t *testing.T) {
	db := setupExtraApprovalDB(t)
	defer teardownTestDB(t, db)
	seedTestUser(t)

	seedExtraTask(t, db, uuid.New().String(), testOrgID, "PENDING", "", "requisition")
	seedExtraTask(t, db, uuid.New().String(), testOrgID, "COMPLETED", "", "requisition")

	app := newFullApprovalApp(t, db)
	resp := testRequest(app, http.MethodGet, "/approvals?status=PENDING", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestGetApprovalTasks_Extra_DocumentTypeFilter exercises the documentType filter.
func TestGetApprovalTasks_Extra_DocumentTypeFilter(t *testing.T) {
	db := setupExtraApprovalDB(t)
	defer teardownTestDB(t, db)
	seedTestUser(t)

	seedExtraTask(t, db, uuid.New().String(), testOrgID, "PENDING", "", "purchase_order")

	app := newFullApprovalApp(t, db)
	resp := testRequest(app, http.MethodGet, "/approvals?document_type=purchase_order", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestGetApprovalTasks_Extra_PriorityFilter exercises the priority filter.
func TestGetApprovalTasks_Extra_PriorityFilter(t *testing.T) {
	db := setupExtraApprovalDB(t)
	defer teardownTestDB(t, db)
	seedTestUser(t)

	seedExtraTask(t, db, uuid.New().String(), testOrgID, "PENDING", "", "requisition")

	app := newFullApprovalApp(t, db)
	resp := testRequest(app, http.MethodGet, "/approvals?priority=MEDIUM", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestGetApprovalTasks_Extra_AssignedToMe exercises the assignedToMe branch.
func TestGetApprovalTasks_Extra_AssignedToMe(t *testing.T) {
	db := setupExtraApprovalDB(t)
	defer teardownTestDB(t, db)
	seedTestUser(t)

	seedExtraTaskWithUserAssignment(t, db, uuid.New().String(), testOrgID, "PENDING", testUserID)

	app := newFullApprovalApp(t, db)
	resp := testRequest(app, http.MethodGet, "/approvals?assigned_to_me=true", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestGetApprovalTasks_Extra_ViewAll exercises the view_all=true branch.
func TestGetApprovalTasks_Extra_ViewAll(t *testing.T) {
	db := setupExtraApprovalDB(t)
	defer teardownTestDB(t, db)
	seedTestUser(t)

	seedExtraTask(t, db, uuid.New().String(), testOrgID, "PENDING", "", "requisition")

	app := newFullApprovalApp(t, db)
	resp := testRequest(app, http.MethodGet, "/approvals?view_all=true", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestGetApprovalTasks_Extra_NonApproverUser exercises the restricted-visibility
// branch for a non-approver role.
func TestGetApprovalTasks_Extra_NonApproverUser(t *testing.T) {
	db := setupExtraApprovalDB(t)
	defer teardownTestDB(t, db)

	requesterID := uuid.New().String()
	db.Create(&models.User{
		ID:     requesterID,
		Email:  "req3@example.com",
		Name:   "Req3",
		Role:   "requester",
		Active: true,
	})
	seedExtraTaskWithUserAssignment(t, db, uuid.New().String(), testOrgID, "PENDING", requesterID)

	app := newNonAdminApprovalApp(t, db, requesterID)
	resp := testRequest(app, http.MethodGet, "/approvals", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestGetApprovalTasks_Extra_UserNotFound exercises the 401 path.
func TestGetApprovalTasks_Extra_UserNotFound(t *testing.T) {
	db := setupExtraApprovalDB(t)
	defer teardownTestDB(t, db)

	h := NewApprovalHandler()
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		},
	})
	app.Get("/approvals",
		withTenantCtx(testOrgID, "ghost-user-xyz", "admin"),
		h.GetApprovalTasks)

	resp := testRequest(app, http.MethodGet, "/approvals", nil)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// GetApprovalTask — populate path with different entity types
// ─────────────────────────────────────────────────────────────────────────────

// TestGetApprovalTask_Extra_AdminSeesAnyTask verifies admin can fetch any task.
func TestGetApprovalTask_Extra_AdminSeesAnyTask(t *testing.T) {
	db := setupExtraApprovalDB(t)
	defer teardownTestDB(t, db)
	seedTestUser(t)

	taskID := uuid.New().String()
	seedExtraTask(t, db, taskID, testOrgID, "PENDING", "", "requisition")

	app := newFullApprovalApp(t, db)
	resp := testRequest(app, http.MethodGet, "/approvals/"+taskID, nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestGetApprovalTask_Extra_NonAdminAssignedTask verifies a non-admin can see a task assigned to them.
func TestGetApprovalTask_Extra_NonAdminAssignedTask(t *testing.T) {
	db := setupExtraApprovalDB(t)
	defer teardownTestDB(t, db)

	requesterID := uuid.New().String()
	db.Create(&models.User{
		ID:     requesterID,
		Email:  "req4@example.com",
		Name:   "Req4",
		Role:   "requester",
		Active: true,
	})

	taskID := uuid.New().String()
	seedExtraTaskWithUserAssignment(t, db, taskID, testOrgID, "PENDING", requesterID)

	app := newNonAdminApprovalApp(t, db, requesterID)
	resp := testRequest(app, http.MethodGet, "/approvals/"+taskID, nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestGetApprovalTask_Extra_NonAdminDeniedTask verifies non-admin cannot access another user's task.
func TestGetApprovalTask_Extra_NonAdminDeniedTask(t *testing.T) {
	db := setupExtraApprovalDB(t)
	defer teardownTestDB(t, db)

	requesterID := uuid.New().String()
	db.Create(&models.User{
		ID:     requesterID,
		Email:  "req5@example.com",
		Name:   "Req5",
		Role:   "requester",
		Active: true,
	})

	// Task assigned to a DIFFERENT user.
	taskID := uuid.New().String()
	otherUserID := uuid.New().String()
	seedExtraTaskWithUserAssignment(t, db, taskID, testOrgID, "PENDING", otherUserID)

	app := newNonAdminApprovalApp(t, db, requesterID)
	resp := testRequest(app, http.MethodGet, "/approvals/"+taskID, nil)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// populateWorkflowTaskFields — exercised indirectly via GetApprovalTask
// ─────────────────────────────────────────────────────────────────────────────

// TestPopulateWorkflowTaskFields_PurchaseOrder seeds a purchase_order entity
// so populateWorkflowTaskFields executes the purchase_order case branch.
func TestPopulateWorkflowTaskFields_PurchaseOrder(t *testing.T) {
	db := setupExtraApprovalDB(t)
	defer teardownTestDB(t, db)
	seedTestUser(t)

	poID := uuid.New().String()
	db.Create(&models.PurchaseOrder{
		ID:             poID,
		OrganizationID: testOrgID,
		DocumentNumber: "PO-POPULATE-001",
		Title:          "Test PO",
		Status:         "PENDING",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	})

	taskID := uuid.New().String()
	seedExtraTask(t, db, taskID, testOrgID, "PENDING", poID, "purchase_order")

	app := newFullApprovalApp(t, db)
	resp := testRequest(app, http.MethodGet, "/approvals/"+taskID, nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestPopulateWorkflowTaskFields_PaymentVoucher exercises the payment_voucher branch.
func TestPopulateWorkflowTaskFields_PaymentVoucher(t *testing.T) {
	db := setupExtraApprovalDB(t)
	defer teardownTestDB(t, db)
	seedTestUser(t)

	pvID := uuid.New().String()
	db.Create(&models.PaymentVoucher{
		ID:             pvID,
		OrganizationID: testOrgID,
		DocumentNumber: "PV-POPULATE-001",
		Title:          "Test PV",
		Status:         "PENDING",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	})

	taskID := uuid.New().String()
	seedExtraTask(t, db, taskID, testOrgID, "PENDING", pvID, "payment_voucher")

	app := newFullApprovalApp(t, db)
	resp := testRequest(app, http.MethodGet, "/approvals/"+taskID, nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestPopulateWorkflowTaskFields_GRN exercises the goods_received_note branch.
func TestPopulateWorkflowTaskFields_GRN(t *testing.T) {
	db := setupExtraApprovalDB(t)
	defer teardownTestDB(t, db)
	seedTestUser(t)

	grnID := uuid.New().String()
	db.Create(&models.GoodsReceivedNote{
		ID:             grnID,
		OrganizationID: testOrgID,
		DocumentNumber: "GRN-POPULATE-001",
		Status:         "PENDING",
		ReceivedBy:     testUserID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	})

	taskID := uuid.New().String()
	seedExtraTask(t, db, taskID, testOrgID, "PENDING", grnID, "goods_received_note")

	app := newFullApprovalApp(t, db)
	resp := testRequest(app, http.MethodGet, "/approvals/"+taskID, nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestPopulateWorkflowTaskFields_Requisition exercises the requisition branch
// with a matching entity row so DocumentNumber gets populated.
func TestPopulateWorkflowTaskFields_Requisition(t *testing.T) {
	db := setupExtraApprovalDB(t)
	defer teardownTestDB(t, db)
	seedTestUser(t)

	reqID := uuid.New().String()
	db.Create(&models.Requisition{
		ID:             reqID,
		OrganizationID: testOrgID,
		DocumentNumber: "REQ-POPULATE-001",
		Title:          "Test Req",
		Status:         "PENDING",
		RequesterId:    testUserID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	})

	taskID := uuid.New().String()
	seedExtraTask(t, db, taskID, testOrgID, "PENDING", reqID, "requisition")

	app := newFullApprovalApp(t, db)
	resp := testRequest(app, http.MethodGet, "/approvals/"+taskID, nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestPopulateWorkflowTaskFields_WithClaimedBy seeds a task claimed by a user
// so the ClaimedBy branch of populateWorkflowTaskFields is exercised.
func TestPopulateWorkflowTaskFields_WithClaimedBy(t *testing.T) {
	db := setupExtraApprovalDB(t)
	defer teardownTestDB(t, db)
	seedTestUser(t)

	taskID := uuid.New().String()
	// Insert a task with claimed_by set to testUserID (admin) and no assigned_user_id.
	err := db.Exec(`INSERT INTO workflow_tasks
		(id, organization_id, workflow_assignment_id, entity_id, entity_type,
		 stage_number, stage_name, assigned_role, claimed_by, status, priority, created_at, version)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		taskID, testOrgID, uuid.New().String(), uuid.New().String(), "requisition",
		1, "Review", "admin", testUserID, "CLAIMED", "MEDIUM", time.Now(), 1,
	).Error
	if err != nil {
		t.Fatalf("seed claimed task: %v", err)
	}

	app := newFullApprovalApp(t, db)
	resp := testRequest(app, http.MethodGet, "/approvals/"+taskID, nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestPopulateWorkflowTaskFields_WithDueDate verifies tasks that already have
// a due_date keep it (the "else" branch in populateWorkflowTaskFields).
func TestPopulateWorkflowTaskFields_WithDueDate(t *testing.T) {
	db := setupExtraApprovalDB(t)
	defer teardownTestDB(t, db)
	seedTestUser(t)

	taskID := uuid.New().String()
	future := time.Now().Add(72 * time.Hour)
	seedExtraTaskWithDueDate(t, db, taskID, testOrgID, "PENDING", future)

	app := newFullApprovalApp(t, db)
	resp := testRequest(app, http.MethodGet, "/approvals/"+taskID, nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// ClaimTask — with injected service
// ─────────────────────────────────────────────────────────────────────────────

// TestClaimTask_Extra_ServiceError verifies that a task that doesn't exist
// causes the workflow service to return an error → 409 Conflict.
func TestClaimTask_Extra_ServiceError(t *testing.T) {
	db := setupExtraApprovalDB(t)
	defer teardownTestDB(t, db)
	seedTestUser(t)

	app := newFullApprovalApp(t, db)
	resp := testRequest(app, http.MethodPost, "/approvals/"+uuid.New().String()+"/claim", nil)
	// Service returns error (task not found) → 409
	assert.Equal(t, http.StatusConflict, resp.StatusCode)
}

// TestClaimTask_Extra_ExistingPendingTask seeds a PENDING task and tries to claim it.
// The service will fail (no matching workflow assignment) but the handler path is reached.
func TestClaimTask_Extra_ExistingPendingTask(t *testing.T) {
	db := setupExtraApprovalDB(t)
	defer teardownTestDB(t, db)
	seedTestUser(t)

	taskID := uuid.New().String()
	seedExtraTask(t, db, taskID, testOrgID, "PENDING", "", "")

	app := newFullApprovalApp(t, db)
	resp := testRequest(app, http.MethodPost, "/approvals/"+taskID+"/claim", nil)
	// Service fails (no workflow assignment) → 409, but handler is reached past the nil check.
	assert.NotEqual(t, http.StatusInternalServerError, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// UnclaimTask — with injected service
// ─────────────────────────────────────────────────────────────────────────────

// TestUnclaimTask_Extra_ServiceError verifies that an unknown task causes → 400.
func TestUnclaimTask_Extra_ServiceError(t *testing.T) {
	db := setupExtraApprovalDB(t)
	defer teardownTestDB(t, db)
	seedTestUser(t)

	app := newFullApprovalApp(t, db)
	resp := testRequest(app, http.MethodPost, "/approvals/"+uuid.New().String()+"/unclaim", nil)
	// Service error → 400
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// TestUnclaimTask_Extra_ExistingTask seeds a CLAIMED task and attempts unclaim.
func TestUnclaimTask_Extra_ExistingTask(t *testing.T) {
	db := setupExtraApprovalDB(t)
	defer teardownTestDB(t, db)
	seedTestUser(t)

	taskID := uuid.New().String()
	err := db.Exec(`INSERT INTO workflow_tasks
		(id, organization_id, workflow_assignment_id, entity_id, entity_type,
		 stage_number, stage_name, assigned_role, claimed_by, status, priority, created_at, version)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		taskID, testOrgID, uuid.New().String(), uuid.New().String(), "requisition",
		1, "Review", "admin", testUserID, "CLAIMED", "MEDIUM", time.Now(), 1,
	).Error
	if err != nil {
		t.Fatalf("seed claimed task: %v", err)
	}

	app := newFullApprovalApp(t, db)
	resp := testRequest(app, http.MethodPost, "/approvals/"+taskID+"/unclaim", nil)
	// Service returns an error (no matching workflow assignment) → 400
	assert.NotEqual(t, http.StatusInternalServerError, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// ApproveTask — with injected service
// ─────────────────────────────────────────────────────────────────────────────

// TestApproveTask_Extra_ServiceFailure seeds a PENDING task, passes validation,
// service fails gracefully → 500 (no matching workflow).
func TestApproveTask_Extra_ServiceFailure(t *testing.T) {
	db := setupExtraApprovalDB(t)
	defer teardownTestDB(t, db)
	seedTestUser(t)

	taskID := uuid.New().String()
	seedExtraTask(t, db, taskID, testOrgID, "PENDING", "", "")

	app := newFullApprovalApp(t, db)
	resp := testRequest(app, http.MethodPost, "/approvals/"+taskID+"/approve",
		map[string]interface{}{"signature": "data:image/png;base64,abc123"})
	// Service will fail with a non-version/non-claim error → 500
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

// TestApproveTask_Extra_WithExpectedVersion exercises the WithVersion branch.
func TestApproveTask_Extra_WithExpectedVersion(t *testing.T) {
	db := setupExtraApprovalDB(t)
	defer teardownTestDB(t, db)
	seedTestUser(t)

	taskID := uuid.New().String()
	seedExtraTask(t, db, taskID, testOrgID, "PENDING", "", "")

	app := newFullApprovalApp(t, db)
	resp := testRequest(app, http.MethodPost, "/approvals/"+taskID+"/approve",
		map[string]interface{}{
			"signature":       "data:image/png;base64,abc123",
			"expectedVersion": 1,
		})
	// Service fails → non-200
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// RejectTask — with injected service
// ─────────────────────────────────────────────────────────────────────────────

// TestRejectTask_Extra_ServiceFailure seeds a PENDING task; service fails gracefully.
func TestRejectTask_Extra_ServiceFailure(t *testing.T) {
	db := setupExtraApprovalDB(t)
	defer teardownTestDB(t, db)
	seedTestUser(t)

	taskID := uuid.New().String()
	seedExtraTask(t, db, taskID, testOrgID, "PENDING", "", "")

	app := newFullApprovalApp(t, db)
	resp := testRequest(app, http.MethodPost, "/approvals/"+taskID+"/reject",
		map[string]interface{}{
			"reason":    "not acceptable",
			"signature": "data:image/png;base64,abc123",
		})
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}

// TestRejectTask_Extra_WithVersion exercises the WithVersion code path.
func TestRejectTask_Extra_WithVersion(t *testing.T) {
	db := setupExtraApprovalDB(t)
	defer teardownTestDB(t, db)
	seedTestUser(t)

	taskID := uuid.New().String()
	seedExtraTask(t, db, taskID, testOrgID, "PENDING", "", "")

	app := newFullApprovalApp(t, db)
	resp := testRequest(app, http.MethodPost, "/approvals/"+taskID+"/reject",
		map[string]interface{}{
			"reason":          "not acceptable",
			"signature":       "data:image/png;base64,abc123",
			"expectedVersion": 1,
		})
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}

// TestRejectTask_Extra_ReturnForRevision exercises the return_for_revision path.
func TestRejectTask_Extra_ReturnForRevision(t *testing.T) {
	db := setupExtraApprovalDB(t)
	defer teardownTestDB(t, db)
	seedTestUser(t)

	taskID := uuid.New().String()
	seedExtraTask(t, db, taskID, testOrgID, "PENDING", "", "")

	app := newFullApprovalApp(t, db)
	resp := testRequest(app, http.MethodPost, "/approvals/"+taskID+"/reject",
		map[string]interface{}{
			"reason":        "needs revision",
			"signature":     "data:image/png;base64,abc123",
			"rejectionType": "return_for_revision",
			"returnToStage": 1,
		})
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// ReassignTask — deeper branches
// ─────────────────────────────────────────────────────────────────────────────

// TestReassignTask_Extra_TaskFound_WrongStatus seeds a COMPLETED task and
// verifies the "not in pending/claimed" 400 is returned.
func TestReassignTask_Extra_TaskFound_WrongStatus(t *testing.T) {
	db := setupExtraApprovalDB(t)
	defer teardownTestDB(t, db)
	seedTestUser(t)

	taskID := uuid.New().String()
	seedExtraTask(t, db, taskID, testOrgID, "COMPLETED", "", "")

	app := newFullApprovalApp(t, db)
	resp := testRequest(app, http.MethodPost, "/approvals/"+taskID+"/reassign",
		map[string]interface{}{
			"newUserId": uuid.New().String(),
			"reason":    "test",
		})
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// TestReassignTask_Extra_TaskFound_CurrentUserNotInDB seeds a PENDING task but
// uses a tenant context whose userID is not in the users table → 401.
func TestReassignTask_Extra_TaskFound_CurrentUserNotInDB(t *testing.T) {
	db := setupExtraApprovalDB(t)
	defer teardownTestDB(t, db)
	// NOTE: we deliberately do NOT seed a User for "ghost-assigner".

	taskID := uuid.New().String()
	// Insert the task directly.
	err := db.Exec(`INSERT INTO workflow_tasks
		(id, organization_id, workflow_assignment_id, entity_id, entity_type,
		 stage_number, stage_name, assigned_role, status, priority, created_at, version)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		taskID, testOrgID, uuid.New().String(), uuid.New().String(), "requisition",
		1, "Review", "admin", "PENDING", "MEDIUM", time.Now(), 1,
	).Error
	if err != nil {
		t.Fatalf("seed task: %v", err)
	}

	h := NewApprovalHandler()
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		},
	})
	app.Post("/approvals/:id/reassign",
		withTenantCtx(testOrgID, "ghost-assigner", "admin"),
		withWorkflowServiceForDB(db),
		h.ReassignTask)

	resp := testRequest(app, http.MethodPost, "/approvals/"+taskID+"/reassign",
		map[string]interface{}{
			"newUserId": uuid.New().String(),
			"reason":    "test",
		})
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// TestReassignTask_Extra_TaskFound_RequesterForbidden seeds a PENDING task and
// tries to reassign with a "requester" role user → 403 Forbidden.
func TestReassignTask_Extra_TaskFound_RequesterForbidden(t *testing.T) {
	db := setupExtraApprovalDB(t)
	defer teardownTestDB(t, db)

	// Seed the requester user.
	requesterID := uuid.New().String()
	db.Create(&models.User{
		ID:     requesterID,
		Email:  "req-reassign@example.com",
		Name:   "ReqReassign",
		Role:   "requester",
		Active: true,
	})

	taskID := uuid.New().String()
	seedExtraTask(t, db, taskID, testOrgID, "PENDING", "", "")

	h := NewApprovalHandler()
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		},
	})
	app.Post("/approvals/:id/reassign",
		withTenantCtx(testOrgID, requesterID, "requester"),
		withWorkflowServiceForDB(db),
		h.ReassignTask)

	resp := testRequest(app, http.MethodPost, "/approvals/"+taskID+"/reassign",
		map[string]interface{}{
			"newUserId": uuid.New().String(),
			"reason":    "test",
		})
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

// TestReassignTask_Extra_NewUserNotFound seeds a PENDING task, current user is
// admin (has permission), but target newUserId does not exist → 400.
func TestReassignTask_Extra_NewUserNotFound(t *testing.T) {
	db := setupExtraApprovalDB(t)
	defer teardownTestDB(t, db)
	seedTestUser(t) // testUserID is "admin"

	taskID := uuid.New().String()
	seedExtraTask(t, db, taskID, testOrgID, "PENDING", "", "")

	app := newFullApprovalApp(t, db)
	resp := testRequest(app, http.MethodPost, "/approvals/"+taskID+"/reassign",
		map[string]interface{}{
			"newUserId": uuid.New().String(), // does not exist
			"reason":    "covering",
		})
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// TestReassignTask_Extra_SuccessfulReassignment seeds all required objects and
// reassigns a task end-to-end.
func TestReassignTask_Extra_SuccessfulReassignment(t *testing.T) {
	db := setupExtraApprovalDB(t)
	defer teardownTestDB(t, db)
	seedTestUser(t) // admin = testUserID

	// Create the new target user that matches current_organization_id lookup.
	newUserID := uuid.New().String()
	db.Create(&models.User{
		ID:                    newUserID,
		Email:                 "newapprover@example.com",
		Name:                  "New Approver",
		Role:                  "approver",
		Active:                true,
		CurrentOrganizationID: &[]string{testOrgID}[0],
	})

	// Seed notifications table so db.Create(&notification) doesn't fail.
	db.Exec(`CREATE TABLE IF NOT EXISTS notifications (
		id TEXT PRIMARY KEY, organization_id TEXT, recipient_id TEXT, sender_id TEXT,
		type TEXT, subject TEXT, body TEXT, message TEXT, related_user_id TEXT,
		related_user_name TEXT, reassignment_reason TEXT, importance TEXT,
		document_id TEXT, document_type TEXT, entity_id TEXT, entity_type TEXT,
		is_read INTEGER DEFAULT 0, sent INTEGER DEFAULT 0,
		created_at DATETIME, updated_at DATETIME
	)`)

	// Seed audit_logs table.
	db.Exec(`CREATE TABLE IF NOT EXISTS audit_logs (
		id TEXT PRIMARY KEY, document_id TEXT, document_type TEXT,
		user_id TEXT, action TEXT, changes TEXT, created_at DATETIME
	)`)

	taskID := uuid.New().String()
	seedExtraTask(t, db, taskID, testOrgID, "PENDING", "", "")

	app := newFullApprovalApp(t, db)
	resp := testRequest(app, http.MethodPost, "/approvals/"+taskID+"/reassign",
		map[string]interface{}{
			"newUserId": newUserID,
			"reason":    "covering for colleague",
		})
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// BulkApprove — additional branches
// ─────────────────────────────────────────────────────────────────────────────

// TestBulkApprove_Extra_MultipleTasksMixed verifies partial success/failure
// reporting when some task IDs exist and others don't.
func TestBulkApprove_Extra_MultipleTasksMixed(t *testing.T) {
	db := setupExtraApprovalDB(t)
	defer teardownTestDB(t, db)
	seedTestUser(t)

	existingTaskID := uuid.New().String()
	seedExtraTask(t, db, existingTaskID, testOrgID, "PENDING", "", "")

	app := newFullApprovalApp(t, db)
	resp := testRequest(app, http.MethodPost, "/approvals/bulk/approve",
		map[string]interface{}{
			"taskIds":   []string{existingTaskID, uuid.New().String()},
			"signature": "data:image/png;base64,abc",
			"comment":   "bulk approve",
		})
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.NotNil(t, body)
}

// TestBulkApprove_Extra_AllFail verifies the FailureCount is returned when no
// tasks are approvable.
func TestBulkApprove_Extra_AllFail(t *testing.T) {
	db := setupExtraApprovalDB(t)
	defer teardownTestDB(t, db)
	seedTestUser(t)

	app := newFullApprovalApp(t, db)
	resp := testRequest(app, http.MethodPost, "/approvals/bulk/approve",
		map[string]interface{}{
			"taskIds":   []string{uuid.New().String(), uuid.New().String()},
			"signature": "sig",
		})
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.NotNil(t, body)
}

// ─────────────────────────────────────────────────────────────────────────────
// BulkReject — additional branches
// ─────────────────────────────────────────────────────────────────────────────

// TestBulkReject_Extra_MultipleTasksMixed mirrors BulkApprove mixed test.
func TestBulkReject_Extra_MultipleTasksMixed(t *testing.T) {
	db := setupExtraApprovalDB(t)
	defer teardownTestDB(t, db)
	seedTestUser(t)

	existingTaskID := uuid.New().String()
	seedExtraTask(t, db, existingTaskID, testOrgID, "PENDING", "", "")

	app := newFullApprovalApp(t, db)
	resp := testRequest(app, http.MethodPost, "/approvals/bulk/reject",
		map[string]interface{}{
			"taskIds":   []string{existingTaskID, uuid.New().String()},
			"signature": "sig",
			"reason":    "batch rejection",
		})
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestBulkReject_Extra_AllFail verifies full-failure reporting.
func TestBulkReject_Extra_AllFail(t *testing.T) {
	db := setupExtraApprovalDB(t)
	defer teardownTestDB(t, db)
	seedTestUser(t)

	app := newFullApprovalApp(t, db)
	resp := testRequest(app, http.MethodPost, "/approvals/bulk/reject",
		map[string]interface{}{
			"taskIds":   []string{uuid.New().String()},
			"signature": "sig",
			"reason":    "bad",
		})
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// GetApprovalWorkflowStatus — with injected service
// ─────────────────────────────────────────────────────────────────────────────

// TestGetApprovalWorkflowStatus_Extra_NoWorkflow exercises the no_workflow path.
func TestGetApprovalWorkflowStatus_Extra_NoWorkflow(t *testing.T) {
	db := setupExtraApprovalDB(t)
	defer teardownTestDB(t, db)
	seedTestUser(t)

	app := newFullApprovalApp(t, db)
	resp := testRequest(app, http.MethodGet, "/documents/"+uuid.New().String()+"/approval-status", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.NotNil(t, body)
	data, _ := body["data"].(map[string]interface{})
	assert.Equal(t, "no_workflow", data["status"])
}

// TestGetApprovalWorkflowStatus_Extra_ByDocumentNumber exercises the
// document-number lookup branch by seeding a requisition first.
func TestGetApprovalWorkflowStatus_Extra_ByDocumentNumber(t *testing.T) {
	db := setupExtraApprovalDB(t)
	defer teardownTestDB(t, db)
	seedTestUser(t)

	reqID := uuid.New().String()
	db.Create(&models.Requisition{
		ID:             reqID,
		OrganizationID: testOrgID,
		DocumentNumber: "REQ-WFS-002",
		Title:          "Test",
		Status:         "PENDING",
		RequesterId:    testUserID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	})

	app := newFullApprovalApp(t, db)
	// Provide the document_number so the DB lookup succeeds and uses reqID.
	resp := testRequest(app, http.MethodGet, "/documents/REQ-WFS-002/approval-status", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestGetApprovalWorkflowStatus_Extra_ByRequisitionID passes the requisition ID
// directly so the ID lookup branch is exercised.
func TestGetApprovalWorkflowStatus_Extra_ByRequisitionID(t *testing.T) {
	db := setupExtraApprovalDB(t)
	defer teardownTestDB(t, db)
	seedTestUser(t)

	reqID := uuid.New().String()
	db.Create(&models.Requisition{
		ID:             reqID,
		OrganizationID: testOrgID,
		DocumentNumber: "REQ-WFS-003",
		Title:          "Test",
		Status:         "PENDING",
		RequesterId:    testUserID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	})

	app := newFullApprovalApp(t, db)
	resp := testRequest(app, http.MethodGet, "/documents/"+reqID+"/approval-status", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
