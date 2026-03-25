package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/repository"
	"github.com/liyali/liyali-gateway/services"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// ─────────────────────────────────────────────────────────────────────────────
// Helpers — all names carry a "WF" suffix to avoid clashes with the identical
// helpers already defined in workflows_http_test.go.
// ─────────────────────────────────────────────────────────────────────────────

// setupWorkflowsTableWF creates the workflows table (and its sibling tables
// used by WorkflowService) using raw DDL so that SQLite-incompatible GORM tags
// on models.Workflow are bypassed.  It is identical in structure to the DDL
// in setupWorkflowTestDB (workflows_http_test.go) but lives under a different
// function name.
func setupWorkflowsTableWF(t *testing.T, db *gorm.DB) {
	t.Helper()
	ddls := []string{
		`CREATE TABLE IF NOT EXISTS workflows (
			id TEXT PRIMARY KEY,
			organization_id TEXT NOT NULL DEFAULT '',
			name TEXT NOT NULL DEFAULT '',
			description TEXT,
			document_type TEXT NOT NULL DEFAULT '',
			entity_type TEXT NOT NULL DEFAULT '',
			version INTEGER DEFAULT 1,
			is_active NUMERIC DEFAULT 1,
			is_default NUMERIC DEFAULT 0,
			conditions JSON,
			stages JSON NOT NULL DEFAULT '[]',
			created_by TEXT NOT NULL DEFAULT '',
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS workflow_defaults (
			id TEXT PRIMARY KEY,
			organization_id TEXT NOT NULL DEFAULT '',
			entity_type TEXT NOT NULL DEFAULT '',
			default_workflow_id TEXT NOT NULL DEFAULT '',
			default_workflow_version INTEGER NOT NULL DEFAULT 1,
			set_by TEXT NOT NULL DEFAULT '',
			set_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS workflow_assignments (
			id TEXT PRIMARY KEY,
			organization_id TEXT NOT NULL DEFAULT '',
			entity_id TEXT NOT NULL DEFAULT '',
			entity_type TEXT NOT NULL DEFAULT '',
			workflow_id TEXT NOT NULL DEFAULT '',
			workflow_version INTEGER NOT NULL DEFAULT 1,
			current_stage INTEGER DEFAULT 0,
			status TEXT DEFAULT 'IN_PROGRESS',
			stage_history JSON,
			assigned_at DATETIME,
			assigned_by TEXT NOT NULL DEFAULT '',
			completed_at DATETIME,
			created_at DATETIME,
			updated_at DATETIME
		)`,
	}
	for _, ddl := range ddls {
		if err := db.Exec(ddl).Error; err != nil {
			t.Fatalf("setupWorkflowsTableWF: %v", err)
		}
	}
}

// newWorkflowServiceWF constructs a WorkflowService backed by the supplied
// test DB.  nil is safe for workflowRepo because GetWorkflows is implemented
// directly against s.db rather than through the repository.
func newWorkflowServiceWF(t *testing.T, db *gorm.DB) *services.WorkflowService {
	t.Helper()
	repo := repository.NewWorkflowRepository(nil, db)
	auditSvc := services.NewAuditService()
	return services.NewWorkflowService(repo, auditSvc, db)
}

// newWorkflowAppWF builds a Fiber app that exposes the two handler routes
// exercised by this file.
func newWorkflowAppWF(t *testing.T, db *gorm.DB) *fiber.App {
	t.Helper()
	svc := newWorkflowServiceWF(t, db)
	h := NewWorkflowHandler(svc)
	auth := withTenantCtx(testOrgID, testUserID, testUserRole)

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		},
	})
	app.Get("/workflows", auth, h.GetWorkflows)
	app.Get("/workflows/:id", auth, h.GetWorkflowByID)
	return app
}

// seedWFRow inserts a workflow row via raw SQL (no GORM model) so that
// PostgreSQL-specific column tags do not cause issues with SQLite.
func seedWFRow(t *testing.T, db *gorm.DB, id, orgID, entityType string, isActive, isDefault bool) {
	t.Helper()
	activeVal := 0
	if isActive {
		activeVal = 1
	}
	defaultVal := 0
	if isDefault {
		defaultVal = 1
	}
	sql := fmt.Sprintf(
		`INSERT INTO workflows
			(id, organization_id, name, description, document_type, entity_type,
			 version, is_active, is_default, conditions, stages, created_by,
			 created_at, updated_at)
		VALUES
			('%s', '%s', 'WF %s', '', '%s', '%s',
			 1, %d, %d, NULL, '[]', '%s',
			 '%s', '%s')`,
		id, orgID, id, entityType, entityType,
		activeVal, defaultVal, testUserID,
		time.Now().UTC().Format("2006-01-02 15:04:05"),
		time.Now().UTC().Format("2006-01-02 15:04:05"),
	)
	if err := db.Exec(sql).Error; err != nil {
		t.Fatalf("seedWFRow(%s): %v", id, err)
	}
}

// decodeSliceResponse reads the response body and decodes it as a JSON array.
// Used for endpoints (like GetWorkflows) that return bare arrays rather than
// wrapped objects.
func decodeSliceResponse(resp *http.Response) []interface{} {
	defer resp.Body.Close()
	var result []interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil
	}
	return result
}

// ─────────────────────────────────────────────────────────────────────────────
// Tests
// ─────────────────────────────────────────────────────────────────────────────

// TestGetWorkflows_WithEntityTypeFilter seeds two workflows with different
// entityTypes, then filters by entityType=requisition and asserts 200.
func TestGetWorkflows_WithEntityTypeFilter(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupWorkflowsTableWF(t, db)

	seedWFRow(t, db, uuid.New().String(), testOrgID, "requisition", true, false)
	seedWFRow(t, db, uuid.New().String(), testOrgID, "purchase_order", true, false)

	app := newWorkflowAppWF(t, db)
	resp := testRequest(app, http.MethodGet, "/workflows?entityType=requisition", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	items := decodeSliceResponse(resp)
	// Should return exactly the requisition workflow (not the purchase_order one).
	assert.NotNil(t, items)
	assert.Len(t, items, 1)
}

// TestGetWorkflows_WithIsActiveFilter seeds one active and one inactive workflow,
// filters isActive=true, and asserts 200.
func TestGetWorkflows_WithIsActiveFilter(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupWorkflowsTableWF(t, db)

	seedWFRow(t, db, uuid.New().String(), testOrgID, "requisition", true, false)
	seedWFRow(t, db, uuid.New().String(), testOrgID, "requisition", false, false)

	app := newWorkflowAppWF(t, db)
	resp := testRequest(app, http.MethodGet, "/workflows?isActive=true", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	items := decodeSliceResponse(resp)
	// Only the active workflow should be returned.
	assert.NotNil(t, items)
	assert.Len(t, items, 1)
}

// TestGetWorkflows_NoFilters fetches all workflows with no query parameters and
// asserts a 200 response.
func TestGetWorkflows_NoFilters(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupWorkflowsTableWF(t, db)

	seedWFRow(t, db, uuid.New().String(), testOrgID, "requisition", true, false)
	seedWFRow(t, db, uuid.New().String(), testOrgID, "purchase_order", true, true)

	app := newWorkflowAppWF(t, db)
	resp := testRequest(app, http.MethodGet, "/workflows", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	items := decodeSliceResponse(resp)
	// Both workflows belong to testOrgID and should be returned.
	assert.NotNil(t, items)
	assert.Len(t, items, 2)
}

// TestGetWorkflowByID_NotFound requests a workflow with a non-existent UUID
// and expects a 404 response.
func TestGetWorkflowByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupWorkflowsTableWF(t, db)

	app := newWorkflowAppWF(t, db)
	resp := testRequest(app, http.MethodGet, "/workflows/"+uuid.New().String(), nil)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// TestGetWorkflows_WithIsDefaultFilter seeds a default and a non-default
// workflow, filters isDefault=true, and asserts 200.
func TestGetWorkflows_WithIsDefaultFilter(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupWorkflowsTableWF(t, db)

	seedWFRow(t, db, uuid.New().String(), testOrgID, "requisition", true, true)  // default
	seedWFRow(t, db, uuid.New().String(), testOrgID, "requisition", true, false) // not default

	app := newWorkflowAppWF(t, db)
	resp := testRequest(app, http.MethodGet, "/workflows?isDefault=true", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	items := decodeSliceResponse(resp)
	// Only the default workflow should be returned.
	assert.NotNil(t, items)
	assert.Len(t, items, 1)
}
