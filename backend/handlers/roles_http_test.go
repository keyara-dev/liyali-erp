package handlers

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ---------------------------------------------------------------------------
// table setup helpers
// ---------------------------------------------------------------------------

// setupRolesTable creates a SQLite-compatible organization_roles table.
// OrganizationRole uses gorm:"type:uuid" which breaks AutoMigrate on SQLite,
// so we create the table manually via raw DDL.
func setupRolesTable(t *testing.T, db *gorm.DB) {
	t.Helper()
	ddl := `CREATE TABLE IF NOT EXISTS organization_roles (
		id TEXT PRIMARY KEY,
		organization_id TEXT,
		name TEXT NOT NULL DEFAULT '',
		description TEXT,
		is_system_role NUMERIC DEFAULT 0,
		active NUMERIC DEFAULT 1,
		permissions JSON,
		created_by TEXT,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`
	if err := db.Exec(ddl).Error; err != nil {
		t.Fatalf("setupRolesTable: %v", err)
	}
}

// setupPermissionsTable creates a SQLite-compatible permissions table (used by
// GetOrganizationPermissions if the service ever queries it).  The current
// implementation returns a hard-coded list so this is mainly defensive.
func setupPermissionsTable(t *testing.T, db *gorm.DB) {
	t.Helper()
	ddl := `CREATE TABLE IF NOT EXISTS permissions (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL DEFAULT '',
		description TEXT,
		resource TEXT,
		action TEXT,
		created_at DATETIME,
		updated_at DATETIME
	)`
	if err := db.Exec(ddl).Error; err != nil {
		t.Fatalf("setupPermissionsTable: %v", err)
	}
}

// seedRole inserts a row directly into organization_roles via raw SQL so the
// tests never depend on GORM's uuid-type column handling.
func seedRole(t *testing.T, db *gorm.DB, id, orgID, name, description string, isSystem bool) {
	t.Helper()
	isSystemInt := 0
	if isSystem {
		isSystemInt = 1
	}
	sql := `INSERT INTO organization_roles (id, organization_id, name, description, is_system_role, active, permissions, created_at, updated_at)
	        VALUES (?, ?, ?, ?, ?, 1, '[]', ?, ?)`
	now := time.Now().Format("2006-01-02 15:04:05")
	if err := db.Exec(sql, id, orgID, name, description, isSystemInt, now, now).Error; err != nil {
		t.Fatalf("seedRole: %v", err)
	}
}

// ---------------------------------------------------------------------------
// app factory
// ---------------------------------------------------------------------------

// newRolesApp builds a minimal Fiber app wired to the roles handlers,
// mirroring the real routing structure.  The optional tenantMiddleware slice
// controls whether a tenant context is injected (auth simulation).
func newRolesApp(tenantMiddleware ...fiber.Handler) *fiber.App {
	app := fiber.New(fiber.Config{
		// Surface panics as 500 so tests can assert non-200 without crashing.
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   err.Error(),
			})
		},
	})

	roles := app.Group("/roles")
	for _, mw := range tenantMiddleware {
		roles.Use(mw)
	}

	roles.Get("/", GetOrganizationRoles)
	roles.Post("/", CreateOrganizationRole)
	roles.Get("/permissions", GetOrganizationPermissions)
	roles.Post("/initialize", InitializeDefaultRoles)
	roles.Put("/:roleId", UpdateOrganizationRole)
	roles.Delete("/:roleId", DeleteOrganizationRole)
	roles.Get("/:roleId/permissions", GetRolePermissions)
	roles.Post("/:roleId/permissions/:permissionId", AssignPermissionToRole)
	roles.Delete("/:roleId/permissions/:permissionId", RemovePermissionFromRole)

	return app
}

// ---------------------------------------------------------------------------
// GET /roles
// ---------------------------------------------------------------------------

func TestGetOrganizationRoles_NoAuth(t *testing.T) {
	// No tenant middleware → organizationID local is missing → 400.
	app := newRolesApp()

	resp := testRequest(app, http.MethodGet, "/roles/", nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	if body == nil {
		t.Fatal("expected JSON response body")
	}
	if body["success"] != false {
		t.Errorf("expected success=false, got %v", body["success"])
	}
}

func TestGetOrganizationRoles_Empty(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupRolesTable(t, db)

	app := newRolesApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	resp := testRequest(app, http.MethodGet, "/roles/", nil)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	if body["success"] != true {
		t.Errorf("expected success=true, got %v", body["success"])
	}

	// data should be an empty array (service returns [] when no rows).
	data, ok := body["data"].([]interface{})
	if !ok {
		t.Fatalf("expected data to be an array, got %T", body["data"])
	}
	if len(data) != 0 {
		t.Errorf("expected 0 roles, got %d", len(data))
	}
}

func TestGetOrganizationRoles_WithData(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupRolesTable(t, db)

	roleID := uuid.New().String()
	seedRole(t, db, roleID, testOrgID, "Procurement Manager", "Handles procurement activities", false)

	// Seed a role for a different org — should not appear (service filters by org + active).
	otherID := uuid.New().String()
	seedRole(t, db, otherID, "other-org-999", "Other Role", "Belongs to another org", false)

	app := newRolesApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	resp := testRequest(app, http.MethodGet, "/roles/", nil)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	if body["success"] != true {
		t.Errorf("expected success=true, got %v", body["success"])
	}

	data, ok := body["data"].([]interface{})
	if !ok {
		t.Fatalf("expected data to be an array, got %T", body["data"])
	}
	if len(data) != 1 {
		t.Errorf("expected 1 role for testOrgID, got %d", len(data))
	}

	first := data[0].(map[string]interface{})
	if first["name"] != "Procurement Manager" {
		t.Errorf("expected name 'Procurement Manager', got %v", first["name"])
	}
}

// ---------------------------------------------------------------------------
// POST /roles
// ---------------------------------------------------------------------------

func TestCreateOrganizationRole_NoAuth(t *testing.T) {
	app := newRolesApp()

	resp := testRequest(app, http.MethodPost, "/roles/", map[string]interface{}{
		"name":        "Test Role",
		"description": "A test role description",
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	if body["success"] != false {
		t.Errorf("expected success=false, got %v", body["success"])
	}
}

func TestCreateOrganizationRole_MissingName(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupRolesTable(t, db)

	app := newRolesApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	resp := testRequest(app, http.MethodPost, "/roles/", map[string]interface{}{
		// name intentionally omitted
		"description": "A description with enough characters",
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	if body["success"] != false {
		t.Errorf("expected success=false, got %v", body["success"])
	}
}

func TestCreateOrganizationRole_NameTooShort(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupRolesTable(t, db)

	app := newRolesApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	// Name is only 2 characters — below the 3-character minimum.
	resp := testRequest(app, http.MethodPost, "/roles/", map[string]interface{}{
		"name":        "AB",
		"description": "A description with enough characters",
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	if body["success"] != false {
		t.Errorf("expected success=false, got %v", body["success"])
	}
}

func TestCreateOrganizationRole_MissingDescription(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupRolesTable(t, db)

	app := newRolesApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	resp := testRequest(app, http.MethodPost, "/roles/", map[string]interface{}{
		"name": "ValidName",
		// description intentionally omitted
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	if body["success"] != false {
		t.Errorf("expected success=false, got %v", body["success"])
	}
}

func TestCreateOrganizationRole_DescriptionTooShort(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupRolesTable(t, db)

	app := newRolesApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	// Description is only 5 characters — below the 10-character minimum.
	resp := testRequest(app, http.MethodPost, "/roles/", map[string]interface{}{
		"name":        "ValidName",
		"description": "Short",
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	if body["success"] != false {
		t.Errorf("expected success=false, got %v", body["success"])
	}
}

func TestCreateOrganizationRole_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupRolesTable(t, db)

	app := newRolesApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	resp := testRequest(app, http.MethodPost, "/roles/", map[string]interface{}{
		"name":        "Custom Viewer",
		"description": "Read-only access to all documents",
	})
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected 201, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	if body["success"] != true {
		t.Errorf("expected success=true, got %v", body["success"])
	}

	data, ok := body["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data to be an object, got %T", body["data"])
	}
	if data["name"] != "Custom Viewer" {
		t.Errorf("expected name 'Custom Viewer', got %v", data["name"])
	}
	if data["description"] != "Read-only access to all documents" {
		t.Errorf("expected description to match, got %v", data["description"])
	}
	if data["id"] == nil || data["id"] == "" {
		t.Errorf("expected non-empty id in response")
	}
	if data["isActive"] != true {
		t.Errorf("expected isActive=true for new role, got %v", data["isActive"])
	}

	// Verify persisted in the DB.
	var count int64
	if err := db.Raw("SELECT COUNT(*) FROM organization_roles WHERE organization_id = ? AND name = ?", testOrgID, "Custom Viewer").Scan(&count).Error; err != nil {
		t.Fatalf("count query failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 persisted role, got %d", count)
	}
}

// ---------------------------------------------------------------------------
// PUT /roles/:roleId
// ---------------------------------------------------------------------------

func TestUpdateOrganizationRole_NoAuth(t *testing.T) {
	app := newRolesApp()

	resp := testRequest(app, http.MethodPut, "/roles/some-id", map[string]interface{}{
		"name": "Updated Name",
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	if body["success"] != false {
		t.Errorf("expected success=false, got %v", body["success"])
	}
}

func TestUpdateOrganizationRole_NameTooShort(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupRolesTable(t, db)

	roleID := uuid.New().String()
	seedRole(t, db, roleID, testOrgID, "Finance Analyst", "Manages financial documents and budgets", false)

	app := newRolesApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	resp := testRequest(app, http.MethodPut, "/roles/"+roleID, map[string]interface{}{
		"name": "AB", // too short
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	if body["success"] != false {
		t.Errorf("expected success=false, got %v", body["success"])
	}
}

func TestUpdateOrganizationRole_DescriptionTooShort(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupRolesTable(t, db)

	roleID := uuid.New().String()
	seedRole(t, db, roleID, testOrgID, "Finance Analyst", "Manages financial documents and budgets", false)

	app := newRolesApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	resp := testRequest(app, http.MethodPut, "/roles/"+roleID, map[string]interface{}{
		"description": "Too short", // 9 chars, below minimum of 10
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	if body["success"] != false {
		t.Errorf("expected success=false, got %v", body["success"])
	}
}

func TestUpdateOrganizationRole_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupRolesTable(t, db)

	app := newRolesApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	resp := testRequest(app, http.MethodPut, "/roles/nonexistent-role-id", map[string]interface{}{
		"name": "Updated Name",
	})
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	if body["success"] != false {
		t.Errorf("expected success=false, got %v", body["success"])
	}
}

func TestUpdateOrganizationRole_WrongOrg(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupRolesTable(t, db)

	// Role belongs to a different org.
	roleID := uuid.New().String()
	seedRole(t, db, roleID, "other-org-999", "Foreign Role", "Belongs to another organisation entirely", false)

	app := newRolesApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	// testOrgID tries to update a role owned by other-org-999.
	resp := testRequest(app, http.MethodPut, "/roles/"+roleID, map[string]interface{}{
		"name": "Hijacked Name",
	})
	// The handler checks organization_id in the WHERE clause → 404.
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestUpdateOrganizationRole_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupRolesTable(t, db)

	roleID := uuid.New().String()
	seedRole(t, db, roleID, testOrgID, "Old Role Name", "Original description of the role", false)

	app := newRolesApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	resp := testRequest(app, http.MethodPut, "/roles/"+roleID, map[string]interface{}{
		"name":        "New Role Name",
		"description": "Updated description of the role",
	})
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	if body["success"] != true {
		t.Errorf("expected success=true, got %v", body["success"])
	}

	data, ok := body["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data to be an object, got %T", body["data"])
	}
	if data["name"] != "New Role Name" {
		t.Errorf("expected updated name 'New Role Name', got %v", data["name"])
	}
	if data["description"] != "Updated description of the role" {
		t.Errorf("expected updated description, got %v", data["description"])
	}
}

func TestUpdateOrganizationRole_PartialUpdate_NameOnly(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupRolesTable(t, db)

	roleID := uuid.New().String()
	seedRole(t, db, roleID, testOrgID, "Old Name", "Existing description kept intact here", false)

	app := newRolesApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	// Only update name — description should be preserved.
	resp := testRequest(app, http.MethodPut, "/roles/"+roleID, map[string]interface{}{
		"name": "Brand New Name",
	})
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	if body["success"] != true {
		t.Errorf("expected success=true, got %v", body["success"])
	}

	data, ok := body["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data to be an object, got %T", body["data"])
	}
	if data["name"] != "Brand New Name" {
		t.Errorf("expected 'Brand New Name', got %v", data["name"])
	}
}

// ---------------------------------------------------------------------------
// DELETE /roles/:roleId
// ---------------------------------------------------------------------------

func TestDeleteOrganizationRole_NoAuth(t *testing.T) {
	// DeleteOrganizationRole does not check organizationID locals — it reads
	// only the roleId URL param and calls the service directly.  To avoid a
	// nil-DB panic we set up the DB but skip injecting tenant context, then
	// request a non-existent role ID.  The service returns "role not found"
	// which the handler maps to 404.
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupRolesTable(t, db)

	// No tenant middleware — handler does not gate on organizationID.
	app := newRolesApp()

	resp := testRequest(app, http.MethodDelete, "/roles/does-not-exist-at-all", nil)
	// Should be 404 (role not found) — definitely not 200.
	if resp.StatusCode == http.StatusOK {
		t.Errorf("expected non-200 for missing role without auth, got 200")
	}
}

func TestDeleteOrganizationRole_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupRolesTable(t, db)

	app := newRolesApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	resp := testRequest(app, http.MethodDelete, "/roles/nonexistent-role-id", nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	if body["success"] != false {
		t.Errorf("expected success=false, got %v", body["success"])
	}
}

func TestDeleteOrganizationRole_SystemRoleBlocked(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupRolesTable(t, db)

	// Seed a system default role (is_system_role=1, name="admin").
	// The service blocks deletion of system default roles by name.
	roleID := uuid.New().String()
	seedRole(t, db, roleID, testOrgID, "admin", "Full administrative access to the platform", true)

	app := newRolesApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	resp := testRequest(app, http.MethodDelete, "/roles/"+roleID, nil)
	// Service returns an error, handler responds with 400.
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 when deleting system role, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	if body["success"] != false {
		t.Errorf("expected success=false, got %v", body["success"])
	}
}

func TestDeleteOrganizationRole_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupRolesTable(t, db)

	roleID := uuid.New().String()
	seedRole(t, db, roleID, testOrgID, "Custom Reviewer", "Custom role that can be deleted freely", false)

	app := newRolesApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	resp := testRequest(app, http.MethodDelete, "/roles/"+roleID, nil)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	if body["success"] != true {
		t.Errorf("expected success=true, got %v", body["success"])
	}

	// Verify the record was soft-deleted (deleted_at set) or hard-deleted.
	var count int64
	db.Raw("SELECT COUNT(*) FROM organization_roles WHERE id = ? AND deleted_at IS NULL", roleID).Scan(&count)
	if count != 0 {
		t.Errorf("expected role to be deleted, but found %d rows still active", count)
	}
}

// ---------------------------------------------------------------------------
// GET /roles/:roleId/permissions
// ---------------------------------------------------------------------------

func TestGetRolePermissions_NoAuth(t *testing.T) {
	// GetRolePermissions reads only roleId from params — it does not gate on
	// organizationID.  We provide a DB to avoid a nil-DB panic and request a
	// non-existent role so the service returns an error.
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupRolesTable(t, db)

	app := newRolesApp() // no tenant middleware

	resp := testRequest(app, http.MethodGet, "/roles/no-such-role/permissions", nil)
	if resp.StatusCode == http.StatusOK {
		t.Errorf("expected non-200 for missing role, got 200")
	}
}

func TestGetRolePermissions_RoleNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupRolesTable(t, db)

	app := newRolesApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	resp := testRequest(app, http.MethodGet, "/roles/nonexistent-role-id/permissions", nil)
	// Service returns "role not found" error → 500 from SendInternalError.
	if resp.StatusCode == http.StatusOK {
		t.Errorf("expected non-200 for non-existent role, got 200")
	}
}

func TestGetRolePermissions_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupRolesTable(t, db)

	roleID := uuid.New().String()
	// Pre-seed permissions JSON in the role row.
	now := time.Now().Format("2006-01-02 15:04:05")
	insertSQL := `INSERT INTO organization_roles (id, organization_id, name, description, is_system_role, active, permissions, created_at, updated_at)
	              VALUES (?, ?, ?, ?, 0, 1, ?, ?, ?)`
	if err := db.Exec(insertSQL, roleID, testOrgID, "Permission Rich Role", "A role with several assigned permissions", `["requisition:view","budget:view"]`, now, now).Error; err != nil {
		t.Fatalf("seedRole with permissions: %v", err)
	}

	app := newRolesApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	resp := testRequest(app, http.MethodGet, fmt.Sprintf("/roles/%s/permissions", roleID), nil)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	if body["success"] != true {
		t.Errorf("expected success=true, got %v", body["success"])
	}

	data, ok := body["data"].([]interface{})
	if !ok {
		t.Fatalf("expected data to be an array, got %T", body["data"])
	}
	if len(data) != 2 {
		t.Errorf("expected 2 permissions, got %d", len(data))
	}
}

func TestGetRolePermissions_EmptyPermissions(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupRolesTable(t, db)

	roleID := uuid.New().String()
	seedRole(t, db, roleID, testOrgID, "Empty Perms Role", "A role with no permissions assigned yet", false)

	app := newRolesApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	resp := testRequest(app, http.MethodGet, fmt.Sprintf("/roles/%s/permissions", roleID), nil)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	if body["success"] != true {
		t.Errorf("expected success=true, got %v", body["success"])
	}
}

// ---------------------------------------------------------------------------
// POST /roles/:roleId/permissions/:permissionId
// ---------------------------------------------------------------------------

func TestAssignPermissionToRole_NoAuth(t *testing.T) {
	// AssignPermissionToRole does not check organizationID.  Provide a DB to
	// avoid a nil-DB panic and use a non-existent role so we get 404.
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupRolesTable(t, db)

	app := newRolesApp() // no tenant middleware

	resp := testRequest(app, http.MethodPost, "/roles/no-such-role/permissions/requisition:view", nil)
	if resp.StatusCode == http.StatusOK {
		t.Errorf("expected non-200 for missing role, got 200")
	}
}

func TestAssignPermissionToRole_RoleNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupRolesTable(t, db)

	app := newRolesApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	resp := testRequest(app, http.MethodPost, "/roles/nonexistent-role-id/permissions/requisition:view", nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	if body["success"] != false {
		t.Errorf("expected success=false, got %v", body["success"])
	}
}

func TestAssignPermissionToRole_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupRolesTable(t, db)

	roleID := uuid.New().String()
	seedRole(t, db, roleID, testOrgID, "Assignable Role", "Role to which permissions will be assigned", false)

	app := newRolesApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	resp := testRequest(app, http.MethodPost, fmt.Sprintf("/roles/%s/permissions/requisition:view", roleID), nil)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	if body["success"] != true {
		t.Errorf("expected success=true, got %v", body["success"])
	}

	// Verify the permission was persisted in the role's permissions JSON.
	var permsJSON string
	db.Raw("SELECT permissions FROM organization_roles WHERE id = ?", roleID).Scan(&permsJSON)
	if permsJSON == "" || permsJSON == "[]" || permsJSON == "null" {
		t.Errorf("expected permissions JSON to contain the assigned permission, got: %s", permsJSON)
	}
}

func TestAssignPermissionToRole_Idempotent(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupRolesTable(t, db)

	// Seed a role that already has the permission.
	roleID := uuid.New().String()
	now := time.Now().Format("2006-01-02 15:04:05")
	insertSQL := `INSERT INTO organization_roles (id, organization_id, name, description, is_system_role, active, permissions, created_at, updated_at)
	              VALUES (?, ?, ?, ?, 0, 1, ?, ?, ?)`
	if err := db.Exec(insertSQL, roleID, testOrgID, "Already Has Perm", "Already has the requisition:view permission", `["requisition:view"]`, now, now).Error; err != nil {
		t.Fatalf("seedRole: %v", err)
	}

	app := newRolesApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	// Assigning the same permission again should be a no-op and succeed.
	resp := testRequest(app, http.MethodPost, fmt.Sprintf("/roles/%s/permissions/requisition:view", roleID), nil)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for idempotent assignment, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	if body["success"] != true {
		t.Errorf("expected success=true, got %v", body["success"])
	}
}

// ---------------------------------------------------------------------------
// DELETE /roles/:roleId/permissions/:permissionId
// ---------------------------------------------------------------------------

func TestRemovePermissionFromRole_NoAuth(t *testing.T) {
	// RemovePermissionFromRole does not check organizationID.  Provide a DB to
	// avoid a nil-DB panic and use a non-existent role so the service errors.
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupRolesTable(t, db)

	app := newRolesApp() // no tenant middleware

	resp := testRequest(app, http.MethodDelete, "/roles/no-such-role/permissions/budget:view", nil)
	if resp.StatusCode == http.StatusOK {
		t.Errorf("expected non-200 for missing role, got 200")
	}
}

func TestRemovePermissionFromRole_RoleNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupRolesTable(t, db)

	app := newRolesApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	resp := testRequest(app, http.MethodDelete, "/roles/nonexistent-role-id/permissions/requisition:view", nil)
	// Service returns "role not found" → handler calls SendInternalError → 500.
	if resp.StatusCode == http.StatusOK {
		t.Errorf("expected non-200 for non-existent role, got 200")
	}

	body := decodeResponse(resp)
	if body["success"] != false {
		t.Errorf("expected success=false, got %v", body["success"])
	}
}

func TestRemovePermissionFromRole_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupRolesTable(t, db)

	roleID := uuid.New().String()
	now := time.Now().Format("2006-01-02 15:04:05")
	insertSQL := `INSERT INTO organization_roles (id, organization_id, name, description, is_system_role, active, permissions, created_at, updated_at)
	              VALUES (?, ?, ?, ?, 0, 1, ?, ?, ?)`
	if err := db.Exec(insertSQL, roleID, testOrgID, "Role With Permissions", "Has two permissions, one will be removed", `["requisition:view","budget:view"]`, now, now).Error; err != nil {
		t.Fatalf("seedRole: %v", err)
	}

	app := newRolesApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	resp := testRequest(app, http.MethodDelete, fmt.Sprintf("/roles/%s/permissions/budget:view", roleID), nil)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	if body["success"] != true {
		t.Errorf("expected success=true, got %v", body["success"])
	}
}

// ---------------------------------------------------------------------------
// GET /roles/permissions (GetOrganizationPermissions)
// ---------------------------------------------------------------------------

func TestGetOrganizationPermissions_NoAuth(t *testing.T) {
	app := newRolesApp()

	resp := testRequest(app, http.MethodGet, "/roles/permissions", nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 without auth, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	if body["success"] != false {
		t.Errorf("expected success=false, got %v", body["success"])
	}
}

func TestGetOrganizationPermissions_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupRolesTable(t, db)
	setupPermissionsTable(t, db)

	app := newRolesApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	resp := testRequest(app, http.MethodGet, "/roles/permissions", nil)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	if body["success"] != true {
		t.Errorf("expected success=true, got %v", body["success"])
	}

	// The service returns a hard-coded list of permissions.
	data, ok := body["data"].([]interface{})
	if !ok {
		t.Fatalf("expected data to be an array, got %T", body["data"])
	}
	if len(data) == 0 {
		t.Errorf("expected at least one permission in the list, got 0")
	}
}

// ---------------------------------------------------------------------------
// POST /roles/initialize (InitializeDefaultRoles)
// ---------------------------------------------------------------------------

func TestInitializeDefaultRoles_NoAuth(t *testing.T) {
	app := newRolesApp()

	resp := testRequest(app, http.MethodPost, "/roles/initialize", nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 without auth, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	if body["success"] != false {
		t.Errorf("expected success=false, got %v", body["success"])
	}
}

func TestInitializeDefaultRoles_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupRolesTable(t, db)

	app := newRolesApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	resp := testRequest(app, http.MethodPost, "/roles/initialize", nil)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	if body["success"] != true {
		t.Errorf("expected success=true, got %v", body["success"])
	}
	if body["message"] != "Default roles initialized successfully" {
		t.Errorf("unexpected message: %v", body["message"])
	}
}
