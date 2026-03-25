package handlers

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// ─────────────────────────────────────────────────────────────────────────────
// Helpers: Workflow table setup for GetApprovalWorkflowStatus tests
// ─────────────────────────────────────────────────────────────────────────────

func setupWorkflowsTable(t *testing.T, db *gorm.DB) {
	t.Helper()
	if err := db.Exec(`CREATE TABLE IF NOT EXISTS workflows (
		id TEXT PRIMARY KEY,
		organization_id TEXT NOT NULL DEFAULT '',
		name TEXT NOT NULL DEFAULT '',
		description TEXT DEFAULT '',
		document_type TEXT NOT NULL DEFAULT '',
		entity_type TEXT NOT NULL DEFAULT '',
		version INTEGER DEFAULT 1,
		is_active NUMERIC DEFAULT 1,
		is_default NUMERIC DEFAULT 0,
		conditions JSON,
		stages JSON NOT NULL DEFAULT '[]',
		created_by TEXT NOT NULL DEFAULT '',
		created_at DATETIME, updated_at DATETIME, deleted_at DATETIME
	)`).Error; err != nil {
		t.Fatalf("setupWorkflowsTable: %v", err)
	}
}

func seedWorkflowRow(t *testing.T, db *gorm.DB, wfID, orgID string) {
	t.Helper()
	stages := `[{"stageNumber":1,"stageName":"Review","requiredRole":"admin","requiredApprovals":1,"canReject":true,"canReassign":true}]`
	err := db.Exec(`INSERT INTO workflows (id, organization_id, name, description, document_type, entity_type, version, is_active, is_default, stages, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		wfID, orgID, "Test Workflow", "", "requisition", "requisition", 1, 1, 0, stages, "system", time.Now(), time.Now()).Error
	if err != nil {
		t.Fatalf("seedWorkflowRow: %v", err)
	}
}

func seedWorkflowAssignmentForEntity(t *testing.T, db *gorm.DB, assignmentID, entityID, workflowID, orgID string) {
	t.Helper()
	err := db.Exec(`INSERT INTO workflow_assignments (id, organization_id, entity_id, entity_type, workflow_id, workflow_version, current_stage, status, stage_history, assigned_at, assigned_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		assignmentID, orgID, entityID, "requisition", workflowID, 1, 1, "IN_PROGRESS", "[]", time.Now(), "system", time.Now(), time.Now()).Error
	if err != nil {
		t.Fatalf("seedWorkflowAssignmentForEntity: %v", err)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// GetApprovalWorkflowStatus — active workflow, no pending tasks
// ─────────────────────────────────────────────────────────────────────────────

func TestGetApprovalWorkflowStatus_WithWorkflow_NoPendingTasks(t *testing.T) {
	db := setupExtraApprovalDB(t)
	defer teardownTestDB(t, db)
	setupWorkflowsTable(t, db)

	wfID := uuid.New().String()
	entityID := uuid.New().String()
	seedWorkflowRow(t, db, wfID, testOrgID)
	seedWorkflowAssignmentForEntity(t, db, uuid.New().String(), entityID, wfID, testOrgID)
	// No pending task seeded

	app := newFullApprovalApp(t, db)
	resp := testRequest(app, http.MethodGet, "/documents/"+entityID+"/approval-status", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.True(t, body["success"].(bool))
	data, ok := body["data"].(map[string]interface{})
	assert.True(t, ok)
	// Status should NOT be "no_workflow" since we have a workflow assigned
	assert.NotEqual(t, "no_workflow", data["status"])
	// canApprove should be false since no pending tasks
	assert.Equal(t, false, data["canApprove"])
}

// ─────────────────────────────────────────────────────────────────────────────
// GetApprovalWorkflowStatus — active workflow + PENDING task for admin role
// ─────────────────────────────────────────────────────────────────────────────

func TestGetApprovalWorkflowStatus_WithPendingAdminTask(t *testing.T) {
	db := setupExtraApprovalDB(t)
	defer teardownTestDB(t, db)
	setupWorkflowsTable(t, db)
	seedTestUser(t) // testUserID with role="admin"

	wfID := uuid.New().String()
	entityID := uuid.New().String()
	seedWorkflowRow(t, db, wfID, testOrgID)
	seedWorkflowAssignmentForEntity(t, db, uuid.New().String(), entityID, wfID, testOrgID)
	// Seed PENDING task assigned to role "admin"
	seedExtraTask(t, db, uuid.New().String(), testOrgID, "PENDING", entityID, "requisition")

	app := newFullApprovalApp(t, db)
	resp := testRequest(app, http.MethodGet, "/documents/"+entityID+"/approval-status", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.True(t, body["success"].(bool))
	data, ok := body["data"].(map[string]interface{})
	assert.True(t, ok)
	// testUser.Role == "admin" matches assignedRole == "admin" → canApprove = true
	assert.Equal(t, true, data["canApprove"])
	assert.Equal(t, true, data["canReject"])
}

// ─────────────────────────────────────────────────────────────────────────────
// GetProfile — success path via mock service
// ─────────────────────────────────────────────────────────────────────────────

func TestGetProfile_SuccessWithMockService(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	config.DB = db

	userRepo := &mockUserRepo{
		getByIDFn: func(_ context.Context, id string) (*models.User, error) {
			return &models.User{ID: id, Email: "test@example.com", Name: "Test User", Role: "admin", Active: true}, nil
		},
	}
	authSvc := newMockAuthService(userRepo, &mockSessionRepo{})
	// Use empty orgID so no DB lookup for org roles is triggered
	app := newAuthAppWithMockService(authSvc, nil, withTenantCtx("", testUserID, "admin"))
	resp := testRequest(app, http.MethodGet, "/auth/profile", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.True(t, body["success"].(bool))
}

// ─────────────────────────────────────────────────────────────────────────────
// UpdateBudget — forbidden when APPROVED status
// ─────────────────────────────────────────────────────────────────────────────

func TestUpdateBudget_ForbiddenWhenApproved(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupBudgetTestDB(t)
	seedTestUser(t)

	b := makeBudget(t, testOrgID, testUserID, "APPROVED")

	app := newBudgetApp(t)
	resp := testRequest(app, http.MethodPut, "/budgets/"+b.ID, map[string]interface{}{
		"department": "Finance",
	})
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// UpdateBudget — success: update description, currency, items on DRAFT budget
// ─────────────────────────────────────────────────────────────────────────────

func TestUpdateBudget_UpdateDescriptionCurrencyItems(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupBudgetTestDB(t)
	seedTestUser(t)

	b := makeBudget(t, testOrgID, testUserID, "DRAFT")

	app := newBudgetApp(t)
	resp := testRequest(app, http.MethodPut, "/budgets/"+b.ID, map[string]interface{}{
		"description": "Updated description",
		"currency":    "USD",
		"items":       []map[string]interface{}{{"description": "Item 1", "amount": 100.0}},
		"name":        "Updated Budget",
	})
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// GetSystemSettings — filters (category, is_secret, is_required, environment, type)
// Note: search filter uses ILIKE which is PostgreSQL-specific; we skip it here.
// ─────────────────────────────────────────────────────────────────────────────

func setupSystemSettingsTableForPush2(t *testing.T, db *gorm.DB) {
	t.Helper()
	// Use AutoMigrate via SystemSetting struct (same as setupSystemSettingsTable)
	if err := db.AutoMigrate(&SystemSetting{}); err != nil {
		t.Fatalf("setupSystemSettingsTableForPush2: %v", err)
	}
}

func newAdminSettingsAppForPush2(t *testing.T) *fiber.App {
	t.Helper()
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		},
	})
	app.Get("/settings", GetSystemSettings)
	app.Get("/settings/:id", GetSystemSetting)
	return app
}

func TestGetSystemSettings_Filters(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupSystemSettingsTableForPush2(t, db)
	config.DB = db

	// Seed a non-secret setting
	db.Exec(`INSERT INTO system_settings (id, key, value, type, category, environment, is_secret, is_required, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		uuid.New().String(), "app.name", "Liyali", "string", "general", "all", 0, 1, time.Now(), time.Now())

	// Seed a secret setting
	db.Exec(`INSERT INTO system_settings (id, key, value, type, category, environment, is_secret, is_required, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		uuid.New().String(), "api.secret_key", "my-secret-value", "string", "security", "production", 1, 1, time.Now(), time.Now())

	app := newAdminSettingsAppForPush2(t)

	// Test with category filter
	resp := testRequest(app, http.MethodGet, "/settings?category=general", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeResponse(resp)
	assert.True(t, body["success"].(bool))

	// Test with is_secret=true filter (value should be hidden)
	resp = testRequest(app, http.MethodGet, "/settings?is_secret=true", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body = decodeResponse(resp)
	if data, ok := body["data"].([]interface{}); ok && len(data) > 0 {
		item := data[0].(map[string]interface{})
		assert.Equal(t, "***HIDDEN***", item["value"])
	}

	// Test with is_required filter
	resp = testRequest(app, http.MethodGet, "/settings?is_required=true", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Test with type filter
	resp = testRequest(app, http.MethodGet, "/settings?type=string", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Test with environment filter
	resp = testRequest(app, http.MethodGet, "/settings?environment=production", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Test empty (no filters)
	resp = testRequest(app, http.MethodGet, "/settings", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// GetRequisitionChain — with PO link in document chain
// ─────────────────────────────────────────────────────────────────────────────

func TestGetRequisitionChain_WithPOLink(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	ensureDocumentLinksTable(t, db)
	setupWorkflowAssignmentsTable(t, db)
	config.DB = db

	req := makeRequisition(t, "REQ-CHAIN-PO-001", "APPROVED")

	// Seed PO
	poID := uuid.New().String()
	po := models.PurchaseOrder{
		ID:             poID,
		OrganizationID: testOrgID,
		DocumentNumber: "PO-CHAIN-001",
		Title:          "Test PO",
		Status:         "APPROVED",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	po.Items = emptyPOItems()
	po.ActionHistory = emptyActionHistory()
	po.ApprovalHistory = emptyApprovalHistory()
	if err := db.Create(&po).Error; err != nil {
		t.Fatalf("create PO: %v", err)
	}

	// Seed document_link: REQ → PO (creates link)
	db.Exec(`INSERT INTO document_links (id, source_doc_id, source_doc_type, target_doc_id, target_doc_type, link_type, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		uuid.New().String(), req.ID, "requisition", poID, "po", "creates", "active", time.Now(), time.Now())

	app := fiber.New()
	auth := withTenantCtx(testOrgID, testUserID, testUserRole)
	app.Get("/requisitions/:id/chain", auth, GetRequisitionChain)

	resp := testRequest(app, http.MethodGet, "/requisitions/"+req.ID+"/chain", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeResponse(resp)
	assert.True(t, body["success"].(bool))
	data := body["data"].(map[string]interface{})
	assert.Equal(t, poID, data["poId"])
	assert.Equal(t, "PO-CHAIN-001", data["poDocumentNumber"])
}

// ─────────────────────────────────────────────────────────────────────────────
// CreateOrganizationUser — success path
// ─────────────────────────────────────────────────────────────────────────────

// setupOrgMembersTableWithDB creates the organization_members table in the
// given db via AutoMigrate (same columns as models.OrganizationMember).
func setupOrgMembersTableWithDB(t *testing.T, db *gorm.DB) {
	t.Helper()
	if err := db.AutoMigrate(&models.OrganizationMember{}); err != nil {
		t.Fatalf("setupOrgMembersTableWithDB AutoMigrate: %v", err)
	}
}

func newAdminUserAppForCoverage(t *testing.T) *fiber.App {
	t.Helper()
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		},
	})
	auth := withTenantCtx(testOrgID, testUserID, testUserRole)
	app.Post("/admin/users", auth, CreateOrganizationUser)
	app.Put("/admin/users/:id", auth, UpdateOrganizationUser)
	return app
}

func TestCreateOrganizationUser_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupOrgMembersTableWithDB(t, db)
	config.DB = db

	app := newAdminUserAppForCoverage(t)
	resp := testRequest(app, http.MethodPost, "/admin/users", map[string]interface{}{
		"email":     "newuser-" + uuid.New().String()[:8] + "@example.com",
		"password":  "SecurePass123!",
		"name":      "New User",
		"position":  "Engineer",
		"manNumber": "MAN-" + uuid.New().String()[:6],
		"nrcNumber": "NRC-" + uuid.New().String()[:6],
		"contact":   "+260971234567",
		"role":      "requester",
	})
	// Expect 201 Created
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// UpdateOrganizationUser — success path
// ─────────────────────────────────────────────────────────────────────────────

func TestUpdateOrganizationUser_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupOrgMembersTableWithDB(t, db)
	config.DB = db

	// Create a user to update
	userToUpdate := models.User{
		ID:     uuid.New().String(),
		Email:  "updateme@example.com",
		Name:   "Old Name",
		Role:   "requester",
		Active: true,
	}
	if err := db.Create(&userToUpdate).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	// Add to org_members
	db.Exec(`INSERT INTO organization_members (id, organization_id, user_id, role, active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		uuid.New().String(), testOrgID, userToUpdate.ID, "requester", 1, time.Now(), time.Now())

	app := newAdminUserAppForCoverage(t)
	resp := testRequest(app, http.MethodPut, "/admin/users/"+userToUpdate.ID, map[string]interface{}{
		"name":     "New Name",
		"position": "Manager",
		"contact":  "+260971111111",
	})
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
