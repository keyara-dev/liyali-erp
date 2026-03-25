package handlers

import (
	"context"
	"encoding/base64"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	fiberrecover "github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/repository"
	"github.com/liyali/liyali-gateway/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─────────────────────────────────────────────────────────────────────────────
// DB helpers for extras tests
// ─────────────────────────────────────────────────────────────────────────────

// setupOrganizationMembersTable creates a SQLite-compatible organization_members table.
func setupOrganizationMembersTable(t *testing.T) {
	t.Helper()
	sql := `CREATE TABLE IF NOT EXISTS organization_members (
		id TEXT PRIMARY KEY,
		organization_id TEXT NOT NULL DEFAULT '',
		user_id TEXT NOT NULL DEFAULT '',
		role TEXT NOT NULL DEFAULT '',
		department TEXT,
		department_id TEXT,
		branch_id TEXT,
		title TEXT,
		active NUMERIC DEFAULT 1,
		invited_at DATETIME,
		joined_at DATETIME,
		invited_by TEXT,
		custom_permissions JSON,
		created_at DATETIME,
		updated_at DATETIME
	)`
	if err := config.DB.Exec(sql).Error; err != nil {
		t.Fatalf("setupOrganizationMembersTable: %v", err)
	}
}

// setupUsersTable creates the users table if not already present.
func setupUsersTable(t *testing.T) {
	t.Helper()
	sql := `CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL DEFAULT '',
		email TEXT NOT NULL DEFAULT '',
		role TEXT NOT NULL DEFAULT '',
		position TEXT,
		active NUMERIC DEFAULT 1,
		password TEXT,
		organization_id TEXT,
		created_at DATETIME,
		updated_at DATETIME
	)`
	if err := config.DB.Exec(sql).Error; err != nil {
		t.Fatalf("setupUsersTable: %v", err)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// App factories for free-function handlers
// ─────────────────────────────────────────────────────────────────────────────

func newDeptHeadsApp(withAuth bool) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		},
	})
	app.Use(fiberrecover.New())

	if withAuth {
		app.Get("/dept-heads", withTenantCtx(testOrgID, testUserID, testUserRole), GetDepartmentHeadsList)
	} else {
		app.Get("/dept-heads", GetDepartmentHeadsList)
	}
	return app
}

func newValidateSignatureApp() *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		},
	})
	app.Post("/validate-signature", ValidateSignature)
	return app
}

func newApproverWorkloadApp(withAuth bool) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		},
	})
	app.Use(fiberrecover.New())

	if withAuth {
		app.Get("/approver-workload/:approverId", withTenantCtx(testOrgID, testUserID, testUserRole), GetApproverWorkload)
	} else {
		app.Get("/approver-workload/:approverId", GetApproverWorkload)
	}
	return app
}

// newPDFPublicApp returns an app wired to GetDocumentForPDFPublic (no auth required).
func newPDFPublicApp(repo repository.DocumentRepositoryInterface) *fiber.App {
	auditSvc := services.NewAuditService()
	docSvc := services.NewDocumentService(repo, auditSvc, nil)
	h := NewDocumentHandler(docSvc)

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		},
	})
	app.Get("/documents/pdf/:documentNumber", h.GetDocumentForPDFPublic)
	return app
}

// ─────────────────────────────────────────────────────────────────────────────
// TestGetDepartmentHeadsList
// ─────────────────────────────────────────────────────────────────────────────

func TestGetDepartmentHeadsList_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupOrganizationMembersTable(t)
	setupUsersTable(t)

	app := newDeptHeadsApp(false)
	resp := testRequest(app, http.MethodGet, "/dept-heads", nil)
	assert.NotEqual(t, http.StatusOK, resp.StatusCode,
		"expected non-200 when tenant context is missing")
}

func TestGetDepartmentHeadsList_Success_EmptyList(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupOrganizationMembersTable(t)
	setupUsersTable(t)

	app := newDeptHeadsApp(true)
	resp := testRequest(app, http.MethodGet, "/dept-heads", nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	require.NotNil(t, body)
	assert.True(t, body["success"].(bool))
}

func TestGetDepartmentHeadsList_WithMembers(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupOrganizationMembersTable(t)
	setupUsersTable(t)

	// Insert an eligible user
	user := models.User{
		ID:    "user-hod-001",
		Name:  "Head Of Dept",
		Email: "hod@example.com",
		Role:  "approver",
	}
	require.NoError(t, db.Exec(
		"INSERT INTO users (id, name, email, role, active) VALUES (?, ?, ?, ?, 1)",
		user.ID, user.Name, user.Email, user.Role,
	).Error)

	// Insert a member record linking the user to the org
	require.NoError(t, db.Exec(
		"INSERT INTO organization_members (id, organization_id, user_id, role, active) VALUES (?, ?, ?, ?, 1)",
		"om-001", testOrgID, user.ID, user.Role,
	).Error)

	app := newDeptHeadsApp(true)
	resp := testRequest(app, http.MethodGet, "/dept-heads", nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	require.NotNil(t, body)
	assert.True(t, body["success"].(bool))
	// data should be a slice with at least one entry
	data, ok := body["data"].([]interface{})
	require.True(t, ok, "data should be a JSON array")
	assert.GreaterOrEqual(t, len(data), 1)
}

func TestGetDepartmentHeadsList_FilterByDepartment(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupOrganizationMembersTable(t)
	setupUsersTable(t)

	require.NoError(t, db.Exec(
		"INSERT INTO users (id, name, email, role, active) VALUES (?, ?, ?, ?, 1)",
		"user-fin-001", "Finance Head", "fin@example.com", "finance",
	).Error)
	require.NoError(t, db.Exec(
		"INSERT INTO organization_members (id, organization_id, user_id, role, department_id, active) VALUES (?, ?, ?, ?, ?, 1)",
		"om-fin-001", testOrgID, "user-fin-001", "finance", "dept-finance",
	).Error)

	app := newDeptHeadsApp(true)
	// Filter by a different department — result should be empty
	resp := testRequest(app, http.MethodGet, "/dept-heads?department_id=dept-other", nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	require.NotNil(t, body)
	// When the result set is empty the data field may be null or [] — both are acceptable.
	if data, ok := body["data"].([]interface{}); ok {
		assert.Len(t, data, 0)
	}

	// Filter by the correct department — result should have 1 entry
	resp2 := testRequest(app, http.MethodGet, "/dept-heads?department_id=dept-finance", nil)
	require.Equal(t, http.StatusOK, resp2.StatusCode)
	body2 := decodeResponse(resp2)
	require.NotNil(t, body2)
	data2, ok := body2["data"].([]interface{})
	require.True(t, ok, "data should be a JSON array")
	assert.Len(t, data2, 1)
}

// ─────────────────────────────────────────────────────────────────────────────
// TestValidateSignature
// ─────────────────────────────────────────────────────────────────────────────

func TestValidateSignature_MissingBody(t *testing.T) {
	app := newValidateSignatureApp()
	// Send an empty JSON object — signature field will be empty string
	resp := testRequest(app, http.MethodPost, "/validate-signature", map[string]interface{}{})
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	body := decodeResponse(resp)
	require.NotNil(t, body)
	assert.False(t, body["success"].(bool))
}

func TestValidateSignature_EmptySignature(t *testing.T) {
	app := newValidateSignatureApp()
	resp := testRequest(app, http.MethodPost, "/validate-signature", map[string]interface{}{
		"signature": "",
		"userId":    testUserID,
	})
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestValidateSignature_InvalidBase64(t *testing.T) {
	app := newValidateSignatureApp()
	resp := testRequest(app, http.MethodPost, "/validate-signature", map[string]interface{}{
		"signature": "this is not base64!!!",
		"userId":    testUserID,
	})
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeResponse(resp)
	require.NotNil(t, body)
	// Handler returns 200 with valid=false for bad base64
	data, ok := body["data"].(map[string]interface{})
	require.True(t, ok, "data should be an object")
	assert.Equal(t, false, data["valid"])
}

func TestValidateSignature_ValidBase64(t *testing.T) {
	app := newValidateSignatureApp()
	encoded := base64.StdEncoding.EncodeToString([]byte("fake-png-bytes"))
	resp := testRequest(app, http.MethodPost, "/validate-signature", map[string]interface{}{
		"signature": encoded,
		"userId":    testUserID,
	})
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeResponse(resp)
	require.NotNil(t, body)
	data, ok := body["data"].(map[string]interface{})
	require.True(t, ok, "data should be an object")
	assert.Equal(t, true, data["valid"])
}

func TestValidateSignature_DataURIFormat(t *testing.T) {
	app := newValidateSignatureApp()
	raw := base64.StdEncoding.EncodeToString([]byte("fake-image-data"))
	dataURI := "data:image/png;base64," + raw
	resp := testRequest(app, http.MethodPost, "/validate-signature", map[string]interface{}{
		"signature": dataURI,
		"userId":    testUserID,
	})
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeResponse(resp)
	data, ok := body["data"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, true, data["valid"])
}

// ─────────────────────────────────────────────────────────────────────────────
// TestGetApproverWorkload
// ─────────────────────────────────────────────────────────────────────────────

func TestGetApproverWorkload_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupWorkflowTasksTable(t, db)
	setupWorkflowAssignmentsTable(t, db)

	app := newApproverWorkloadApp(false)
	resp := testRequest(app, http.MethodGet, "/approver-workload/some-approver", nil)
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}

func TestGetApproverWorkload_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupWorkflowTasksTable(t, db)
	setupWorkflowAssignmentsTable(t, db)

	app := newApproverWorkloadApp(true)
	resp := testRequest(app, http.MethodGet, "/approver-workload/approver-001", nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	require.NotNil(t, body)
	assert.True(t, body["success"].(bool))
	data, ok := body["data"].(map[string]interface{})
	require.True(t, ok, "data should be an object")
	assert.Contains(t, data, "pendingCount")
	assert.Contains(t, data, "completedThisMonth")
	assert.Contains(t, data, "overdueTasks")
}

// setupWorkflowTasksWithAssignedTo creates a workflow_tasks table that includes the
// assigned_to column used by GetApproverWorkload.  The shared setupWorkflowTasksTable
// helper uses assigned_user_id; this variant matches what the handler actually queries.
func setupWorkflowTasksWithAssignedTo(t *testing.T) {
	t.Helper()
	sql := `CREATE TABLE IF NOT EXISTS workflow_tasks (
		id TEXT PRIMARY KEY,
		organization_id TEXT NOT NULL DEFAULT '',
		workflow_assignment_id TEXT NOT NULL DEFAULT '',
		entity_id TEXT NOT NULL DEFAULT '',
		entity_type TEXT NOT NULL DEFAULT '',
		stage_number INTEGER NOT NULL DEFAULT 0,
		stage_name TEXT NOT NULL DEFAULT '',
		assignment_type TEXT DEFAULT 'role',
		assigned_role TEXT,
		assigned_user_id TEXT,
		assigned_to TEXT,
		status TEXT DEFAULT 'PENDING',
		priority TEXT DEFAULT 'MEDIUM',
		created_at DATETIME,
		claimed_at DATETIME,
		claimed_by TEXT,
		completed_at DATETIME,
		due_date DATETIME,
		version INTEGER DEFAULT 1,
		updated_by TEXT,
		claim_expiry DATETIME,
		updated_at DATETIME
	)`
	if err := config.DB.Exec(sql).Error; err != nil {
		t.Fatalf("setupWorkflowTasksWithAssignedTo: %v", err)
	}
}

func TestGetApproverWorkload_WithPendingTask(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupWorkflowTasksWithAssignedTo(t)
	setupWorkflowAssignmentsTable(t, db)

	// Seed a workflow_assignment
	require.NoError(t, db.Exec(
		`INSERT INTO workflow_assignments (id, organization_id, entity_id, entity_type, workflow_id, workflow_version, assigned_by)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"wa-001", testOrgID, "entity-001", "requisition", "wf-001", 1, testUserID,
	).Error)

	// Seed a pending task
	require.NoError(t, db.Exec(
		`INSERT INTO workflow_tasks (id, workflow_assignment_id, entity_id, entity_type, stage_number, stage_name, assigned_to, status)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"task-001", "wa-001", "entity-001", "requisition", 1, "Review", "approver-001", "PENDING",
	).Error)

	app := newApproverWorkloadApp(true)
	resp := testRequest(app, http.MethodGet, "/approver-workload/approver-001", nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	data := body["data"].(map[string]interface{})
	// pendingCount should be at least 1
	assert.GreaterOrEqual(t, data["pendingCount"].(float64), float64(1))
}

// ─────────────────────────────────────────────────────────────────────────────
// TestGetDocumentForPDFPublic
// ─────────────────────────────────────────────────────────────────────────────

func TestGetDocumentForPDFPublic_NotFound(t *testing.T) {
	repo := &mockDocumentRepo{
		getByNumberOnlyFn: func(ctx context.Context, number string) (*models.Document, error) {
			return nil, nil // not found, no doc
		},
	}
	app := newPDFPublicApp(repo)
	resp := testRequest(app, http.MethodGet, "/documents/pdf/UNKNOWN-999", nil)
	// Should be 404 when document cannot be resolved
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestGetDocumentForPDFPublic_RequisitionFound(t *testing.T) {
	req := &models.Requisition{
		ID:             "req-001",
		OrganizationID: testOrgID,
		DocumentNumber: "REQ-2024-001",
	}
	repo := &mockDocumentRepo{
		getByNumberOnlyFn: func(ctx context.Context, number string) (*models.Document, error) {
			// Return a generic document record that points to a requisition type
			return &models.Document{
				OrganizationID: testOrgID,
				DocumentNumber: number,
				DocumentType:   "REQUISITION",
			}, nil
		},
		getRequisitionByNumberFn: func(ctx context.Context, number string) (*models.Requisition, error) {
			return req, nil
		},
		getOrganizationBrandingFn: func(ctx context.Context, orgID string) (*models.Organization, error) {
			return &models.Organization{ID: orgID, Name: "Test Org"}, nil
		},
	}
	app := newPDFPublicApp(repo)
	resp := testRequest(app, http.MethodGet, "/documents/pdf/REQ-2024-001", nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeResponse(resp)
	require.NotNil(t, body)
	assert.True(t, body["success"].(bool))
}

func TestGetDocumentForPDFPublic_MissingDocNumber(t *testing.T) {
	repo := &mockDocumentRepo{}
	app := newPDFPublicApp(repo)
	// Route is /documents/pdf/:documentNumber — calling without param gives 404 from router
	resp := testRequest(app, http.MethodGet, "/documents/pdf/", nil)
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// TestGetWorkflowsLegacy — covered via direct call through WorkflowHandler
// ─────────────────────────────────────────────────────────────────────────────

// newWorkflowLegacyApp wires a Fiber app that calls getWorkflowsLegacy directly.
// Because the tests are in package handlers (same package), we can access the
// private method directly.
func newWorkflowLegacyApp(t *testing.T, withAuth bool) *fiber.App {
	t.Helper()
	svc := newWorkflowService(t)
	h := NewWorkflowHandler(svc)

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		},
	})
	app.Use(fiberrecover.New())

	// Thin wrapper that guards on the organizationID local (mirrors what GetWorkflows does)
	// and then delegates to the private getWorkflowsLegacy.
	handler := func(c *fiber.Ctx) error {
		orgIDRaw := c.Locals("organizationID")
		if orgIDRaw == nil {
			return fiber.NewError(fiber.StatusBadRequest, "Organization ID not found in context")
		}
		orgID, ok := orgIDRaw.(string)
		if !ok {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid organization ID format")
		}
		return h.getWorkflowsLegacy(c, orgID)
	}

	if withAuth {
		app.Get("/workflows-legacy", withTenantCtx(testOrgID, testUserID, testUserRole), handler)
	} else {
		app.Get("/workflows-legacy", handler)
	}
	return app
}

func TestGetWorkflowsLegacy_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupWorkflowTestDB(t)

	app := newWorkflowLegacyApp(t, false)
	resp := testRequest(app, http.MethodGet, "/workflows-legacy", nil)
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}

func TestGetWorkflowsLegacy_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupWorkflowTestDB(t)

	app := newWorkflowLegacyApp(t, true)
	resp := testRequest(app, http.MethodGet, "/workflows-legacy", nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	require.NotNil(t, body)
	assert.True(t, body["success"].(bool))
}

func TestGetWorkflowsLegacy_WithDocumentTypeFilter(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupWorkflowTestDB(t)

	// Seed a workflow of type "requisition"
	makeWorkflow(t, testOrgID, testUserID)

	app := newWorkflowLegacyApp(t, true)
	resp := testRequest(app, http.MethodGet, "/workflows-legacy?documentType=requisition&activeOnly=true", nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	require.NotNil(t, body)
	assert.True(t, body["success"].(bool))
}
