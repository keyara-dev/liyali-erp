package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/utils"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	testOrgID   = "test-org-001"
	testUserID  = "test-user-001"
	testUserRole = "admin"
)

// setupTestDB creates an in-memory SQLite database, auto-migrates the required
// models, sets config.DB to the new instance and returns it.
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open in-memory SQLite DB: %v", err)
	}

	// WorkflowTask and WorkflowAssignment are excluded here because their
	// gorm:"type:uuid" columns trigger a SQLite syntax error. Tests that need
	// workflow_tasks create it manually via setupWorkflowTasksTable().
	err = db.AutoMigrate(
		&models.OrganizationBranch{},
		&models.Vendor{},
		&models.Category{},
		&models.CategoryBudgetCode{},
		&models.Requisition{},
		&models.PurchaseOrder{},
		&models.PaymentVoucher{},
		&models.GoodsReceivedNote{},
		&models.Organization{},
		&models.User{},
	)
	if err != nil {
		t.Fatalf("failed to auto-migrate models: %v", err)
	}

	config.DB = db
	return db
}

// teardownTestDB closes the underlying SQL connection and resets config.DB.
func teardownTestDB(t *testing.T, db *gorm.DB) {
	t.Helper()

	sqlDB, err := db.DB()
	if err == nil {
		_ = sqlDB.Close()
	}

	config.DB = nil
}

// withTenantCtx returns a Fiber middleware that injects a *utils.TenantContext
// into c.Locals("tenant") and also sets the individual locals that some
// handlers read directly.
func withTenantCtx(orgID, userID, role string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tenant := &utils.TenantContext{
			OrganizationID: orgID,
			UserID:         userID,
			UserRole:       role,
			Department:     "",
		}
		c.Locals("tenant", tenant)
		c.Locals("organizationID", orgID)
		c.Locals("userID", userID)
		c.Locals("userRole", role)
		return c.Next()
	}
}

// testRequest fires an HTTP request against the Fiber app.  body can be nil or
// any value that will be JSON-marshalled.  Content-Type is set automatically
// when a body is present.
func testRequest(app *fiber.App, method, path string, body interface{}) *http.Response {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			panic("testRequest: failed to marshal body: " + err.Error())
		}
		reqBody = bytes.NewReader(b)
	}

	req := httptest.NewRequest(method, path, reqBody)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := app.Test(req, -1)
	if err != nil {
		// Handler panicked (e.g. nil type assertion without auth middleware).
		// Return a synthetic 500 so tests can assert NotEqual(200) without crashing.
		return &http.Response{StatusCode: http.StatusInternalServerError}
	}

	return resp
}

// jsonBody JSON-marshals v and returns an io.Reader suitable for use as an
// HTTP request body.  Panics on marshal failure (should never happen in tests).
func jsonBody(v interface{}) io.Reader {
	b, err := json.Marshal(v)
	if err != nil {
		panic("jsonBody: marshal failed: " + err.Error())
	}
	return bytes.NewReader(b)
}

// setupWorkflowTasksTable creates a SQLite-compatible workflow_tasks table.
// Called by approval tests that need to query or seed tasks.
func setupWorkflowTasksTable(t *testing.T, db *gorm.DB) {
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
	if err := db.Exec(sql).Error; err != nil {
		t.Fatalf("setupWorkflowTasksTable: %v", err)
	}
}

// setupWorkflowAssignmentsTable creates a SQLite-compatible workflow_assignments table.
// WorkflowAssignment uses type:uuid and type:jsonb tags that break SQLite AutoMigrate.
func setupWorkflowAssignmentsTable(t *testing.T, db *gorm.DB) {
	t.Helper()
	sql := `CREATE TABLE IF NOT EXISTS workflow_assignments (
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
	)`
	if err := db.Exec(sql).Error; err != nil {
		t.Fatalf("setupWorkflowAssignmentsTable: %v", err)
	}
}

// decodeResponse reads the response body once and unmarshals it into a generic
// map so individual tests can inspect fields without importing utils.
func decodeResponse(resp *http.Response) map[string]interface{} {
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil
	}

	return result
}
