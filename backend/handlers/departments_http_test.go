package handlers

import (
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// newDepartmentsApp builds a minimal Fiber app wired to the department handlers.
func newDepartmentsApp(tenantMiddleware ...fiber.Handler) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   err.Error(),
			})
		},
	})

	depts := app.Group("/departments")
	for _, mw := range tenantMiddleware {
		depts.Use(mw)
	}

	depts.Get("/", GetOrganizationDepartments)
	depts.Post("/", CreateOrganizationDepartment)
	depts.Get("/:id", GetOrganizationDepartment)
	depts.Put("/:id", UpdateOrganizationDepartment)
	depts.Delete("/:id", DeleteOrganizationDepartment)
	depts.Post("/:id/restore", RestoreOrganizationDepartment)
	depts.Get("/:id/modules", GetDepartmentModules)
	depts.Post("/:id/modules", AssignModuleToDepartment)
	depts.Delete("/:departmentId/modules/:moduleId", RemoveModuleFromDepartment)

	return app
}

// setupDepartmentsTable creates the organization_departments and department_modules
// tables in SQLite using raw DDL (AutoMigrate cannot handle the full model due to
// SQLite type differences with PostgreSQL).
func setupDepartmentsTable(t *testing.T, db *gorm.DB) {
	t.Helper()

	deptSQL := `CREATE TABLE IF NOT EXISTS organization_departments (
		id TEXT PRIMARY KEY,
		organization_id TEXT NOT NULL DEFAULT '',
		name TEXT NOT NULL DEFAULT '',
		code TEXT,
		description TEXT,
		manager_name TEXT,
		parent_id TEXT,
		is_active NUMERIC DEFAULT 1,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`
	if err := db.Exec(deptSQL).Error; err != nil {
		t.Fatalf("setupDepartmentsTable (departments): %v", err)
	}

	modSQL := `CREATE TABLE IF NOT EXISTS department_modules (
		id TEXT PRIMARY KEY,
		department_id TEXT NOT NULL DEFAULT '',
		module_id TEXT NOT NULL DEFAULT '',
		created_at DATETIME
	)`
	if err := db.Exec(modSQL).Error; err != nil {
		t.Fatalf("setupDepartmentsTable (modules): %v", err)
	}
}

// seedDepartment inserts an OrganizationDepartment row directly via GORM.
func seedDepartment(t *testing.T, db *gorm.DB, orgID, name, code string, isActive bool) models.OrganizationDepartment {
	t.Helper()

	dept := models.OrganizationDepartment{
		ID:             uuid.New().String(),
		OrganizationID: orgID,
		Name:           name,
		Code:           code,
		IsActive:       isActive,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	if err := db.Exec(
		`INSERT INTO organization_departments (id, organization_id, name, code, is_active, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		dept.ID, dept.OrganizationID, dept.Name, dept.Code, dept.IsActive, dept.CreatedAt, dept.UpdatedAt,
	).Error; err != nil {
		t.Fatalf("seedDepartment: %v", err)
	}

	return dept
}

// ---------------------------------------------------------------------------
// GET /departments
// ---------------------------------------------------------------------------

func TestGetDepartments_NoAuth(t *testing.T) {
	app := newDepartmentsApp()

	resp := testRequest(app, http.MethodGet, "/departments/", nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

func TestGetDepartments_Empty(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupDepartmentsTable(t, db)

	app := newDepartmentsApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	resp := testRequest(app, http.MethodGet, "/departments/", nil)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])
}

func TestGetDepartments_WithData(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupDepartmentsTable(t, db)

	seedDepartment(t, db, testOrgID, "Finance", "FIN", true)
	seedDepartment(t, db, testOrgID, "Operations", "OPS", true)
	// Different org — must NOT appear in response.
	seedDepartment(t, db, "other-org-999", "Other Dept", "OTH", true)

	app := newDepartmentsApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	resp := testRequest(app, http.MethodGet, "/departments/", nil)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])

	data, ok := body["data"].([]interface{})
	assert.True(t, ok, "data should be an array")
	assert.Len(t, data, 2, "only departments belonging to testOrgID should be returned")
}

func TestGetDepartments_ActiveFilter(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupDepartmentsTable(t, db)

	seedDepartment(t, db, testOrgID, "Active Dept", "ACT", true)
	seedDepartment(t, db, testOrgID, "Inactive Dept", "INA", false)

	app := newDepartmentsApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	// active=true should only return the active department.
	resp := testRequest(app, http.MethodGet, "/departments/?active=true", nil)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])

	data, ok := body["data"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, data, 1)
}

func TestGetDepartments_InactiveFilter(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupDepartmentsTable(t, db)

	seedDepartment(t, db, testOrgID, "Active Dept B", "ACTB", true)
	seedDepartment(t, db, testOrgID, "Inactive Dept B", "INAB", false)

	app := newDepartmentsApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	// active=false should only return the inactive department.
	resp := testRequest(app, http.MethodGet, "/departments/?active=false", nil)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])

	data, ok := body["data"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, data, 1)
}

// ---------------------------------------------------------------------------
// GET /departments/:id
// ---------------------------------------------------------------------------

func TestGetDepartment_NoAuth(t *testing.T) {
	app := newDepartmentsApp()

	resp := testRequest(app, http.MethodGet, "/departments/some-id", nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

func TestGetDepartment_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupDepartmentsTable(t, db)

	app := newDepartmentsApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	resp := testRequest(app, http.MethodGet, "/departments/nonexistent-id", nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

func TestGetDepartment_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupDepartmentsTable(t, db)

	dept := seedDepartment(t, db, testOrgID, "Procurement", "PROC", true)

	app := newDepartmentsApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	resp := testRequest(app, http.MethodGet, "/departments/"+dept.ID, nil)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])

	data, ok := body["data"].(map[string]interface{})
	assert.True(t, ok, "data should be an object")
	assert.Equal(t, dept.ID, data["id"])
	assert.Equal(t, "Procurement", data["name"])
	assert.Equal(t, "PROC", data["code"])
}

func TestGetDepartment_TenantIsolation(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupDepartmentsTable(t, db)

	// A department belonging to a different org.
	otherDept := seedDepartment(t, db, "other-org-999", "Foreign Dept", "FOR", true)

	app := newDepartmentsApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	// testOrgID tries to fetch a department owned by another org → 404.
	resp := testRequest(app, http.MethodGet, "/departments/"+otherDept.ID, nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404 for cross-tenant access, got %d", resp.StatusCode)
	}
}

// ---------------------------------------------------------------------------
// POST /departments
// ---------------------------------------------------------------------------

func TestCreateDepartment_NoAuth(t *testing.T) {
	app := newDepartmentsApp()

	resp := testRequest(app, http.MethodPost, "/departments/", map[string]interface{}{
		"name": "Test Dept",
		"code": "TST",
	})
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

func TestCreateDepartment_MissingName(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupDepartmentsTable(t, db)

	app := newDepartmentsApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	// name intentionally omitted.
	resp := testRequest(app, http.MethodPost, "/departments/", map[string]interface{}{
		"code": "TST",
	})
	// Handler validates: name must be at least 2 chars → 400.
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

func TestCreateDepartment_NameTooShort(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupDepartmentsTable(t, db)

	app := newDepartmentsApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	// name is only 1 character — below the 2-char minimum.
	resp := testRequest(app, http.MethodPost, "/departments/", map[string]interface{}{
		"name": "X",
		"code": "TST",
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

func TestCreateDepartment_MissingCode(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupDepartmentsTable(t, db)

	app := newDepartmentsApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	// code intentionally omitted.
	resp := testRequest(app, http.MethodPost, "/departments/", map[string]interface{}{
		"name": "Valid Name",
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

func TestCreateDepartment_CodeTooShort(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupDepartmentsTable(t, db)

	app := newDepartmentsApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	// code is only 1 character — below the 2-char minimum.
	resp := testRequest(app, http.MethodPost, "/departments/", map[string]interface{}{
		"name": "Valid Name",
		"code": "X",
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

func TestCreateDepartment_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupDepartmentsTable(t, db)

	app := newDepartmentsApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	desc := "Handles all HR functions"
	resp := testRequest(app, http.MethodPost, "/departments/", map[string]interface{}{
		"name":        "Human Resources",
		"code":        "HR",
		"description": desc,
	})
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected 201, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])

	data, ok := body["data"].(map[string]interface{})
	assert.True(t, ok, "data should be an object")
	assert.Equal(t, "Human Resources", data["name"])
	assert.Equal(t, "HR", data["code"])
	assert.NotEmpty(t, data["id"])

	// Verify the record was persisted in DB.
	var count int64
	db.Raw(
		"SELECT COUNT(*) FROM organization_departments WHERE organization_id = ? AND code = ?",
		testOrgID, "HR",
	).Scan(&count)
	assert.Equal(t, int64(1), count)
}

func TestCreateDepartment_DuplicateCode(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupDepartmentsTable(t, db)

	// Pre-create a department with the code we will attempt to duplicate.
	seedDepartment(t, db, testOrgID, "Existing Dept", "DUP", true)

	app := newDepartmentsApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	resp := testRequest(app, http.MethodPost, "/departments/", map[string]interface{}{
		"name": "New Dept Same Code",
		"code": "DUP",
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for duplicate code, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

func TestCreateDepartment_DuplicateCode_DifferentOrg_Allowed(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupDepartmentsTable(t, db)

	// A department with the same code in a DIFFERENT org — should not conflict.
	seedDepartment(t, db, "other-org-999", "Other Dept", "SHRD", true)

	app := newDepartmentsApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	resp := testRequest(app, http.MethodPost, "/departments/", map[string]interface{}{
		"name": "Shared Code Dept",
		"code": "SHRD",
	})
	// Should succeed because the duplicate belongs to a different org.
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected 201 (cross-org same code is allowed), got %d", resp.StatusCode)
	}
}

// ---------------------------------------------------------------------------
// PUT /departments/:id
// ---------------------------------------------------------------------------

func TestUpdateDepartment_NoAuth(t *testing.T) {
	app := newDepartmentsApp()

	resp := testRequest(app, http.MethodPut, "/departments/some-id", map[string]interface{}{
		"name": "Updated Name",
	})
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

func TestUpdateDepartment_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupDepartmentsTable(t, db)

	app := newDepartmentsApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	resp := testRequest(app, http.MethodPut, "/departments/nonexistent-id", map[string]interface{}{
		"name": "Updated Name",
	})
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

func TestUpdateDepartment_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupDepartmentsTable(t, db)

	dept := seedDepartment(t, db, testOrgID, "Old Name", "OLD", true)

	app := newDepartmentsApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	isActive := false
	resp := testRequest(app, http.MethodPut, "/departments/"+dept.ID, map[string]interface{}{
		"name":      "New Name",
		"is_active": isActive,
	})
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])

	data, ok := body["data"].(map[string]interface{})
	assert.True(t, ok, "data should be an object")
	assert.Equal(t, "New Name", data["name"])
}

func TestUpdateDepartment_DuplicateCode(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupDepartmentsTable(t, db)

	dept1 := seedDepartment(t, db, testOrgID, "Dept Alpha", "ALPH", true)
	dept2 := seedDepartment(t, db, testOrgID, "Dept Beta", "BETA", true)

	app := newDepartmentsApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	// Try to update dept2's code to the code already used by dept1.
	code := dept1.Code
	resp := testRequest(app, http.MethodPut, "/departments/"+dept2.ID, map[string]interface{}{
		"code": code,
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for duplicate code on update, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

func TestUpdateDepartment_SameCode_SameDept_Allowed(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupDepartmentsTable(t, db)

	dept := seedDepartment(t, db, testOrgID, "Stable Dept", "STBL", true)

	app := newDepartmentsApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	// Updating a department with its own existing code must not be treated as a duplicate.
	resp := testRequest(app, http.MethodPut, "/departments/"+dept.ID, map[string]interface{}{
		"name": "Stable Dept Renamed",
		"code": "STBL",
	})
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 when keeping same code, got %d", resp.StatusCode)
	}
}

// ---------------------------------------------------------------------------
// DELETE /departments/:id
// ---------------------------------------------------------------------------

func TestDeleteDepartment_NoAuth(t *testing.T) {
	app := newDepartmentsApp()

	resp := testRequest(app, http.MethodDelete, "/departments/some-id", nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

func TestDeleteDepartment_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupDepartmentsTable(t, db)

	app := newDepartmentsApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	resp := testRequest(app, http.MethodDelete, "/departments/nonexistent-id", nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

func TestDeleteDepartment_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupDepartmentsTable(t, db)

	dept := seedDepartment(t, db, testOrgID, "To Be Deleted", "DEL", true)

	app := newDepartmentsApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	resp := testRequest(app, http.MethodDelete, "/departments/"+dept.ID, nil)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])
	assert.Equal(t, "Department deleted successfully", body["message"])

	// The service soft-deletes by setting is_active = false.  Verify the row
	// still exists but is now inactive.
	var isActive bool
	db.Raw(
		"SELECT is_active FROM organization_departments WHERE id = ?",
		dept.ID,
	).Scan(&isActive)
	assert.False(t, isActive, "soft-deleted department should have is_active=false")
}

// ---------------------------------------------------------------------------
// POST /departments/:id/restore
// ---------------------------------------------------------------------------

func TestRestoreDepartment_NoAuth(t *testing.T) {
	app := newDepartmentsApp()

	resp := testRequest(app, http.MethodPost, "/departments/some-id/restore", nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

func TestRestoreDepartment_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupDepartmentsTable(t, db)

	app := newDepartmentsApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	// RestoreOrganizationDepartment calls RestoreDepartment which returns an
	// error when RowsAffected == 0 (no matching row).
	resp := testRequest(app, http.MethodPost, "/departments/nonexistent-id/restore", nil)
	// Any non-200 response is acceptable; the handler maps the service error to 500.
	if resp.StatusCode == http.StatusOK {
		t.Errorf("expected non-200 for restore of nonexistent department, got %d", resp.StatusCode)
	}
}

func TestRestoreDepartment_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupDepartmentsTable(t, db)

	// Seed an inactive (soft-deleted) department.
	dept := seedDepartment(t, db, testOrgID, "Dormant Dept", "DORM", false)

	app := newDepartmentsApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	resp := testRequest(app, http.MethodPost, "/departments/"+dept.ID+"/restore", nil)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])
	assert.Equal(t, "Department restored successfully", body["message"])

	// Verify the row is now active.
	var isActive bool
	db.Raw(
		"SELECT is_active FROM organization_departments WHERE id = ?",
		dept.ID,
	).Scan(&isActive)
	assert.True(t, isActive, "restored department should have is_active=true")
}

// ---------------------------------------------------------------------------
// GET /departments/:id/modules
// ---------------------------------------------------------------------------

func TestGetDepartmentModules_NoAuth(t *testing.T) {
	app := newDepartmentsApp()

	resp := testRequest(app, http.MethodGet, "/departments/some-id/modules", nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

func TestGetDepartmentModules_DeptNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupDepartmentsTable(t, db)

	app := newDepartmentsApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	resp := testRequest(app, http.MethodGet, "/departments/nonexistent-id/modules", nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

func TestGetDepartmentModules_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupDepartmentsTable(t, db)

	dept := seedDepartment(t, db, testOrgID, "IT Dept", "IT", true)

	app := newDepartmentsApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	resp := testRequest(app, http.MethodGet, "/departments/"+dept.ID+"/modules", nil)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])
	// GetDepartmentModules is a placeholder that returns an empty slice.
	data, ok := body["data"].([]interface{})
	assert.True(t, ok, "data should be an array")
	assert.Len(t, data, 0)
}

// ---------------------------------------------------------------------------
// POST /departments/:id/modules
// ---------------------------------------------------------------------------

func TestAssignModuleToDepartment_NoAuth(t *testing.T) {
	app := newDepartmentsApp()

	resp := testRequest(app, http.MethodPost, "/departments/some-id/modules", map[string]interface{}{
		"module_id": "mod-001",
	})
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

func TestAssignModuleToDepartment_DeptNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupDepartmentsTable(t, db)

	app := newDepartmentsApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	resp := testRequest(app, http.MethodPost, "/departments/nonexistent-id/modules", map[string]interface{}{
		"module_id": "mod-001",
	})
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

func TestAssignModuleToDepartment_MissingModuleID(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupDepartmentsTable(t, db)

	dept := seedDepartment(t, db, testOrgID, "Legal Dept", "LEG", true)

	app := newDepartmentsApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	// module_id intentionally omitted.
	resp := testRequest(app, http.MethodPost, "/departments/"+dept.ID+"/modules", map[string]interface{}{})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

func TestAssignModuleToDepartment_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupDepartmentsTable(t, db)

	dept := seedDepartment(t, db, testOrgID, "Marketing Dept", "MKT", true)

	app := newDepartmentsApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	// AssignModuleToDepartment is a placeholder that always succeeds.
	resp := testRequest(app, http.MethodPost, "/departments/"+dept.ID+"/modules", map[string]interface{}{
		"module_id": "mod-requisitions",
	})
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])
	assert.Equal(t, "Module assigned to department successfully", body["message"])
}

// ---------------------------------------------------------------------------
// DELETE /departments/:id/modules/:moduleId
// ---------------------------------------------------------------------------

func TestRemoveModuleFromDepartment_NoAuth(t *testing.T) {
	app := newDepartmentsApp()

	resp := testRequest(app, http.MethodDelete, "/departments/some-id/modules/mod-001", nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

func TestRemoveModuleFromDepartment_DeptNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupDepartmentsTable(t, db)

	app := newDepartmentsApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	resp := testRequest(app, http.MethodDelete, "/departments/nonexistent-id/modules/mod-001", nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

func TestRemoveModuleFromDepartment_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupDepartmentsTable(t, db)

	dept := seedDepartment(t, db, testOrgID, "Sales Dept", "SAL", true)

	app := newDepartmentsApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	// RemoveModuleFromDepartment is a placeholder that always succeeds.
	resp := testRequest(app, http.MethodDelete, "/departments/"+dept.ID+"/modules/mod-purchase-orders", nil)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])
	assert.Equal(t, "Module removed from department successfully", body["message"])
}
