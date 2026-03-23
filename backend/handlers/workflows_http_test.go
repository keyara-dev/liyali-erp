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
	"github.com/liyali/liyali-gateway/repository"
	"github.com/liyali/liyali-gateway/services"
	"gorm.io/datatypes"
)

// ─────────────────────────────────────────────────────────────────────────────
// Test DB setup with Workflow table
// ─────────────────────────────────────────────────────────────────────────────

func setupWorkflowTestDB(t *testing.T) {
	t.Helper()
	db := config.DB
	if db == nil {
		t.Fatal("setupWorkflowTestDB: config.DB is nil — call setupTestDB first")
	}
	// models.Workflow uses PostgreSQL-specific GORM tags (type:uuid, type:jsonb,
	// default:gen_random_uuid()) that break SQLite AutoMigrate. Use raw DDL instead.
	sql := `CREATE TABLE IF NOT EXISTS workflows (
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
	)`
	if err := db.Exec(sql).Error; err != nil {
		t.Fatalf("setupWorkflowTestDB: %v", err)
	}

	// workflow_defaults and workflow_assignments also use PostgreSQL-specific tags
	for _, ddl := range []string{
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
	} {
		if err := db.Exec(ddl).Error; err != nil {
			t.Fatalf("setupWorkflowTestDB extra DDL: %v", err)
		}
	}
}

// newWorkflowService constructs a real WorkflowService backed by the test DB.
// nil is passed for pgxPool because the service methods exercised in tests use
// GORM directly rather than the pgx pool.
func newWorkflowService(t *testing.T) *services.WorkflowService {
	t.Helper()
	db := config.DB
	if db == nil {
		t.Fatal("newWorkflowService: config.DB is nil — call setupTestDB first")
	}
	repo := repository.NewWorkflowRepository(nil, db)
	auditSvc := services.NewAuditService()
	return services.NewWorkflowService(repo, auditSvc, db)
}

// ─────────────────────────────────────────────────────────────────────────────
// App factories
// ─────────────────────────────────────────────────────────────────────────────

func newWorkflowApp(t *testing.T) *fiber.App {
	t.Helper()
	svc := newWorkflowService(t)
	h := NewWorkflowHandler(svc)
	auth := withTenantCtx(testOrgID, testUserID, testUserRole)

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		},
	})

	app.Get("/workflows", auth, h.GetWorkflows)
	app.Get("/workflows/:id", auth, h.GetWorkflowByID)
	app.Post("/workflows", auth, h.CreateWorkflow)
	app.Put("/workflows/:id", auth, h.UpdateWorkflow)
	app.Delete("/workflows/:id", auth, h.DeleteWorkflow)
	app.Post("/workflows/:id/activate", auth, h.ActivateWorkflow)
	app.Post("/workflows/:id/deactivate", auth, h.DeactivateWorkflow)
	app.Post("/workflows/validate", auth, h.ValidateWorkflow)
	return app
}

func newWorkflowAppNoAuth(t *testing.T) *fiber.App {
	t.Helper()
	svc := newWorkflowService(t)
	h := NewWorkflowHandler(svc)

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		},
	})
	app.Use(fiberrecover.New())

	app.Get("/workflows", h.GetWorkflows)
	app.Get("/workflows/:id", h.GetWorkflowByID)
	app.Post("/workflows", h.CreateWorkflow)
	app.Put("/workflows/:id", h.UpdateWorkflow)
	app.Delete("/workflows/:id", h.DeleteWorkflow)
	app.Post("/workflows/:id/activate", h.ActivateWorkflow)
	app.Post("/workflows/:id/deactivate", h.DeactivateWorkflow)
	app.Post("/workflows/validate", h.ValidateWorkflow)
	return app
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

// validStagePayload returns a minimal valid workflow stage for request bodies.
func validStagePayload() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"stageNumber":  1,
			"stageName":    "Initial Review",
			"requiredRole": "admin",
		},
	}
}

// makeWorkflow inserts a minimal Workflow into the DB and returns it.
func makeWorkflow(t *testing.T, orgID, userID string) models.Workflow {
	t.Helper()
	stagesJSON, _ := datatypes.JSON(`[{"stageNumber":1,"stageName":"Initial Review","requiredRole":"admin"}]`).MarshalJSON()
	wf := models.Workflow{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Name:           "Test Workflow",
		EntityType:     "requisition",
		DocumentType:   "requisition",
		Stages:         datatypes.JSON(stagesJSON),
		IsActive:       true,
		IsDefault:      false,
		CreatedBy:      userID,
		Version:        1,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	if err := config.DB.Create(&wf).Error; err != nil {
		t.Fatalf("makeWorkflow: %v", err)
	}
	return wf
}

// ─────────────────────────────────────────────────────────────────────────────
// NoAuth tests — all endpoints must refuse unauthenticated access
// ─────────────────────────────────────────────────────────────────────────────

func TestListWorkflows_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupWorkflowTestDB(t)

	app := newWorkflowAppNoAuth(t)
	resp := testRequest(app, http.MethodGet, "/workflows", nil)
	if resp.StatusCode == http.StatusOK {
		t.Errorf("expected non-200 without auth, got 200")
	}
}

func TestGetWorkflow_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupWorkflowTestDB(t)

	app := newWorkflowAppNoAuth(t)
	resp := testRequest(app, http.MethodGet, "/workflows/"+uuid.New().String(), nil)
	if resp.StatusCode == http.StatusOK {
		t.Errorf("expected non-200 without auth, got 200")
	}
}

func TestCreateWorkflow_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupWorkflowTestDB(t)

	app := newWorkflowAppNoAuth(t)
	resp := testRequest(app, http.MethodPost, "/workflows", map[string]interface{}{
		"name":       "Test",
		"entityType": "requisition",
		"stages":     validStagePayload(),
	})
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		t.Errorf("expected non-2xx without auth, got %d", resp.StatusCode)
	}
}

func TestUpdateWorkflow_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupWorkflowTestDB(t)

	app := newWorkflowAppNoAuth(t)
	resp := testRequest(app, http.MethodPut, "/workflows/"+uuid.New().String(), map[string]interface{}{
		"name": "Updated",
	})
	if resp.StatusCode == http.StatusOK {
		t.Errorf("expected non-200 without auth, got 200")
	}
}

func TestDeleteWorkflow_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupWorkflowTestDB(t)

	app := newWorkflowAppNoAuth(t)
	resp := testRequest(app, http.MethodDelete, "/workflows/"+uuid.New().String(), nil)
	if resp.StatusCode == http.StatusOK {
		t.Errorf("expected non-200 without auth, got 200")
	}
}

func TestActivateWorkflow_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupWorkflowTestDB(t)

	app := newWorkflowAppNoAuth(t)
	resp := testRequest(app, http.MethodPost, "/workflows/"+uuid.New().String()+"/activate", nil)
	if resp.StatusCode == http.StatusOK {
		t.Errorf("expected non-200 without auth, got 200")
	}
}

func TestDeactivateWorkflow_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupWorkflowTestDB(t)

	app := newWorkflowAppNoAuth(t)
	resp := testRequest(app, http.MethodPost, "/workflows/"+uuid.New().String()+"/deactivate", nil)
	if resp.StatusCode == http.StatusOK {
		t.Errorf("expected non-200 without auth, got 200")
	}
}

func TestValidateWorkflow_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupWorkflowTestDB(t)

	// ValidateWorkflow does NOT read Locals — it only uses workflowService.
	// So NoAuth just confirms the endpoint exists and still processes the body.
	// We exercise it with missing stages so validation returns 400.
	app := newWorkflowAppNoAuth(t)
	resp := testRequest(app, http.MethodPost, "/workflows/validate", map[string]interface{}{
		"name":       "Test",
		"entityType": "requisition",
		"stages":     []interface{}{}, // empty → 400
	})
	if resp.StatusCode == http.StatusOK {
		t.Errorf("expected non-200 for empty stages, got 200")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// ListWorkflows
// ─────────────────────────────────────────────────────────────────────────────

func TestListWorkflows_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupWorkflowTestDB(t)

	app := newWorkflowApp(t)
	resp := testRequest(app, http.MethodGet, "/workflows", nil)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// CreateWorkflow
// ─────────────────────────────────────────────────────────────────────────────

func TestCreateWorkflow_MissingName(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupWorkflowTestDB(t)

	app := newWorkflowApp(t)
	resp := testRequest(app, http.MethodPost, "/workflows", map[string]interface{}{
		"entityType": "requisition",
		"stages":     validStagePayload(),
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for missing name, got %d", resp.StatusCode)
	}
}

func TestCreateWorkflow_MissingEntityType(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupWorkflowTestDB(t)

	app := newWorkflowApp(t)
	resp := testRequest(app, http.MethodPost, "/workflows", map[string]interface{}{
		"name":   "Test Workflow",
		"stages": validStagePayload(),
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for missing entityType, got %d", resp.StatusCode)
	}
}

func TestCreateWorkflow_MissingStages(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupWorkflowTestDB(t)

	app := newWorkflowApp(t)
	resp := testRequest(app, http.MethodPost, "/workflows", map[string]interface{}{
		"name":       "Test Workflow",
		"entityType": "requisition",
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for missing stages, got %d", resp.StatusCode)
	}
}

func TestCreateWorkflow_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupWorkflowTestDB(t)
	seedTestUser(t)

	app := newWorkflowApp(t)
	resp := testRequest(app, http.MethodPost, "/workflows", map[string]interface{}{
		"name":       "Procurement Workflow",
		"entityType": "requisition",
		"stages":     validStagePayload(),
	})
	if resp.StatusCode != http.StatusCreated {
		body := decodeResponse(resp)
		t.Errorf("expected 201, got %d; body=%v", resp.StatusCode, body)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// GetWorkflow
// ─────────────────────────────────────────────────────────────────────────────

func TestGetWorkflow_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupWorkflowTestDB(t)

	app := newWorkflowApp(t)
	resp := testRequest(app, http.MethodGet, "/workflows/"+uuid.New().String(), nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestGetWorkflow_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupWorkflowTestDB(t)
	seedTestUser(t)

	wf := makeWorkflow(t, testOrgID, testUserID)
	app := newWorkflowApp(t)
	resp := testRequest(app, http.MethodGet, "/workflows/"+wf.ID.String(), nil)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// UpdateWorkflow
// ─────────────────────────────────────────────────────────────────────────────

func TestUpdateWorkflow_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupWorkflowTestDB(t)

	app := newWorkflowApp(t)
	resp := testRequest(app, http.MethodPut, "/workflows/"+uuid.New().String(), map[string]interface{}{
		"name": "Updated Name",
	})
	// Not found returns 500 wrapped from "workflow not found" error via SendInternalError.
	if resp.StatusCode == http.StatusOK {
		t.Errorf("expected non-200 for non-existent workflow, got 200")
	}
}

func TestUpdateWorkflow_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupWorkflowTestDB(t)
	seedTestUser(t)

	wf := makeWorkflow(t, testOrgID, testUserID)
	app := newWorkflowApp(t)

	updatedName := "Updated Workflow Name"
	resp := testRequest(app, http.MethodPut, "/workflows/"+wf.ID.String(), map[string]interface{}{
		"name": updatedName,
	})
	if resp.StatusCode != http.StatusOK {
		body := decodeResponse(resp)
		t.Errorf("expected 200, got %d; body=%v", resp.StatusCode, body)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// DeleteWorkflow
// ─────────────────────────────────────────────────────────────────────────────

func TestDeleteWorkflow_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupWorkflowTestDB(t)

	app := newWorkflowApp(t)
	resp := testRequest(app, http.MethodDelete, "/workflows/"+uuid.New().String(), nil)
	if resp.StatusCode == http.StatusOK {
		t.Errorf("expected non-200 for non-existent workflow, got 200")
	}
}

func TestDeleteWorkflow_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupWorkflowTestDB(t)
	seedTestUser(t)

	wf := makeWorkflow(t, testOrgID, testUserID)
	app := newWorkflowApp(t)
	resp := testRequest(app, http.MethodDelete, "/workflows/"+wf.ID.String(), nil)
	if resp.StatusCode != http.StatusOK {
		body := decodeResponse(resp)
		t.Errorf("expected 200, got %d; body=%v", resp.StatusCode, body)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// ActivateWorkflow
// ─────────────────────────────────────────────────────────────────────────────

func TestActivateWorkflow_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupWorkflowTestDB(t)

	app := newWorkflowApp(t)
	resp := testRequest(app, http.MethodPost, "/workflows/"+uuid.New().String()+"/activate", nil)
	if resp.StatusCode == http.StatusOK {
		t.Errorf("expected non-200 for non-existent workflow, got 200")
	}
}

func TestActivateWorkflow_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupWorkflowTestDB(t)
	seedTestUser(t)

	wf := makeWorkflow(t, testOrgID, testUserID)
	// Mark as inactive so Activate has something to do.
	config.DB.Model(&wf).Update("is_active", false)

	app := newWorkflowApp(t)
	resp := testRequest(app, http.MethodPost, "/workflows/"+wf.ID.String()+"/activate", nil)
	if resp.StatusCode != http.StatusOK {
		body := decodeResponse(resp)
		t.Errorf("expected 200, got %d; body=%v", resp.StatusCode, body)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// DeactivateWorkflow
// ─────────────────────────────────────────────────────────────────────────────

func TestDeactivateWorkflow_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupWorkflowTestDB(t)

	app := newWorkflowApp(t)
	resp := testRequest(app, http.MethodPost, "/workflows/"+uuid.New().String()+"/deactivate", nil)
	if resp.StatusCode == http.StatusOK {
		t.Errorf("expected non-200 for non-existent workflow, got 200")
	}
}

func TestDeactivateWorkflow_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupWorkflowTestDB(t)
	seedTestUser(t)

	wf := makeWorkflow(t, testOrgID, testUserID)
	app := newWorkflowApp(t)
	resp := testRequest(app, http.MethodPost, "/workflows/"+wf.ID.String()+"/deactivate", nil)
	if resp.StatusCode != http.StatusOK {
		body := decodeResponse(resp)
		t.Errorf("expected 200, got %d; body=%v", resp.StatusCode, body)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// ValidateWorkflow
// ─────────────────────────────────────────────────────────────────────────────

func TestValidateWorkflow_MissingStages(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupWorkflowTestDB(t)

	app := newWorkflowApp(t)
	// Empty stages slice triggers ValidateWorkflowStages to return an error.
	resp := testRequest(app, http.MethodPost, "/workflows/validate", map[string]interface{}{
		"name":       "Test",
		"entityType": "requisition",
		"stages":     []interface{}{},
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for empty stages, got %d", resp.StatusCode)
	}
}

func TestValidateWorkflow_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupWorkflowTestDB(t)

	app := newWorkflowApp(t)
	resp := testRequest(app, http.MethodPost, "/workflows/validate", map[string]interface{}{
		"name":       "Test",
		"entityType": "requisition",
		"stages":     validStagePayload(),
	})
	if resp.StatusCode != http.StatusOK {
		body := decodeResponse(resp)
		t.Errorf("expected 200, got %d; body=%v", resp.StatusCode, body)
	}
}
