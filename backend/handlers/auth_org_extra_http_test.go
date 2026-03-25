package handlers

// auth_org_extra_http_test.go – additional coverage for:
//   • auth_handler.go:  AdminLogin, AdminRefreshToken, LogoutAll (admin app wiring)
//   • organization.go:  GetUserOrganizations, CreateOrganization, GetOrganizationByID,
//                       SwitchOrganization (success via top-level route),
//                       GetOrganizationMembers (paginated / filtered paths),
//                       DeleteOrganization (via top-level route)
//
// These tests are designed to complement (not duplicate) what already exists in:
//   handlers/auth_http_test.go
//   handlers/org_invitation_perms_http_test.go
//   handlers/subscription_http_test.go

import (
	"net/http"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/liyali/liyali-gateway/config"
)

// ─────────────────────────────────────────────────────────────────────────────
// Fiber app helpers
// ─────────────────────────────────────────────────────────────────────────────

// newAdminAuthApp wires AdminLogin, AdminRefreshToken, and LogoutAll to a
// Fiber app using a nil authService.  All tests exercising this app only hit
// paths that return before reaching the service (validation failures, auth
// guard failures).
func newAdminAuthApp(mw ...fiber.Handler) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   err.Error(),
			})
		},
	})

	h := &AuthHandler{
		authService: nil,
		rbacService: nil,
		validate:    validator.New(),
	}

	g := app.Group("/admin")
	for _, m := range mw {
		g.Use(m)
	}
	g.Post("/login", h.AdminLogin)
	g.Post("/refresh", h.AdminRefreshToken)
	g.Post("/logout-all", h.LogoutAll)
	return app
}

// newOrgTopApp wires the top-level organization handlers (those that read
// c.Locals("userID") directly, not via tenant context).  These are distinct
// from the routes registered inside newOrgApp (which uses tenant middleware).
func newOrgTopApp(mw ...fiber.Handler) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   err.Error(),
			})
		},
	})
	g := app.Group("/orgs")
	for _, m := range mw {
		g.Use(m)
	}
	g.Get("/", GetUserOrganizations)
	g.Post("/", CreateOrganization)
	g.Get("/:id", GetOrganizationByID)
	g.Post("/:id/switch", SwitchOrganization)
	g.Delete("/:id", DeleteOrganization)
	return app
}

// ─────────────────────────────────────────────────────────────────────────────
// AdminLogin — validation tests
// ─────────────────────────────────────────────────────────────────────────────

// TestAdminLogin_EmptyBody expects 400 — both email and password are required.
func TestAdminLogin_EmptyBody(t *testing.T) {
	app := newAdminAuthApp()
	resp := testRequest(app, http.MethodPost, "/admin/login", nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for empty body, got %d", resp.StatusCode)
	}
	body := decodeResponse(resp)
	if body["success"] != false {
		t.Errorf("expected success=false")
	}
}

// TestAdminLogin_MissingEmail expects 400 — email absent.
func TestAdminLogin_MissingEmail(t *testing.T) {
	app := newAdminAuthApp()
	resp := testRequest(app, http.MethodPost, "/admin/login", map[string]interface{}{
		"password": "secret123",
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for missing email, got %d", resp.StatusCode)
	}
}

// TestAdminLogin_MissingPassword expects 400 — password absent.
func TestAdminLogin_MissingPassword(t *testing.T) {
	app := newAdminAuthApp()
	resp := testRequest(app, http.MethodPost, "/admin/login", map[string]interface{}{
		"email": "admin@example.com",
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for missing password, got %d", resp.StatusCode)
	}
}

// TestAdminLogin_InvalidEmailFormat expects 400 — bad email format triggers
// the go-playground/validator `email` tag.
func TestAdminLogin_InvalidEmailFormat(t *testing.T) {
	app := newAdminAuthApp()
	resp := testRequest(app, http.MethodPost, "/admin/login", map[string]interface{}{
		"email":    "notanemail",
		"password": "secret123",
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid email format, got %d", resp.StatusCode)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// AdminRefreshToken — validation tests
// ─────────────────────────────────────────────────────────────────────────────

// TestAdminRefreshToken_EmptyBody expects 400 — refreshToken is required.
func TestAdminRefreshToken_EmptyBody(t *testing.T) {
	app := newAdminAuthApp()
	resp := testRequest(app, http.MethodPost, "/admin/refresh", nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for empty body, got %d", resp.StatusCode)
	}
	body := decodeResponse(resp)
	if body["success"] != false {
		t.Errorf("expected success=false")
	}
}

// TestAdminRefreshToken_MissingRefreshToken expects 400 — empty JSON object.
func TestAdminRefreshToken_MissingRefreshToken(t *testing.T) {
	app := newAdminAuthApp()
	resp := testRequest(app, http.MethodPost, "/admin/refresh", map[string]interface{}{})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for missing refreshToken, got %d", resp.StatusCode)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// LogoutAll via admin app — NoAuth check (mirrors auth_http_test.go but via
// the admin-wired app to cover AdminRefreshToken/LogoutAll code paths together)
// ─────────────────────────────────────────────────────────────────────────────

// TestLogoutAll_AdminApp_NoAuth verifies 401 when no userID is in context on
// the admin-wired app.
func TestLogoutAll_AdminApp_NoAuth(t *testing.T) {
	app := newAdminAuthApp() // no userID middleware
	resp := testRequest(app, http.MethodPost, "/admin/logout-all", nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 for unauthenticated logout-all (admin app), got %d", resp.StatusCode)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// GetUserOrganizations
// ─────────────────────────────────────────────────────────────────────────────

// TestGetUserOrganizations_NoAuth expects 401 when userID is absent.
func TestGetUserOrganizations_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newOrgTopApp() // no userID middleware
	resp := testRequest(app, http.MethodGet, "/orgs/", nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

// TestGetUserOrganizations_WithAuth_EmptyMemberships expects 200 with an empty
// array when the user has no org memberships.
func TestGetUserOrganizations_WithAuth_EmptyMemberships(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupOrgMembersTable(t)

	app := newOrgTopApp(withTenantCtx(testOrgID, testUserID, testUserRole))
	resp := testRequest(app, http.MethodGet, "/orgs/", nil)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for authenticated get-user-organizations, got %d", resp.StatusCode)
	}
	body := decodeResponse(resp)
	if body["success"] != true {
		t.Errorf("expected success=true, got %v", body)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// CreateOrganization — top-level route tests
// ─────────────────────────────────────────────────────────────────────────────

// TestCreateOrganization_NoAuth_TopRoute expects 401 when userID is absent.
func TestCreateOrganization_NoAuth_TopRoute(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newOrgTopApp()
	resp := testRequest(app, http.MethodPost, "/orgs/", map[string]interface{}{
		"name": "My Org",
	})
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 for unauthenticated create-organization, got %d", resp.StatusCode)
	}
}

// TestCreateOrganization_WithName_Success_TopRoute verifies a 201 when name
// is present and user is authenticated (in-memory SQLite org is created).
func TestCreateOrganization_WithName_Success_TopRoute(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupOrgMembersTable(t)

	app := newOrgTopApp(withTenantCtx(testOrgID, testUserID, testUserRole))
	resp := testRequest(app, http.MethodPost, "/orgs/", map[string]interface{}{
		"name": "Test Organisation Beta",
	})
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected 201 for create-organization with name, got %d", resp.StatusCode)
	}
	body := decodeResponse(resp)
	if body["success"] != true {
		t.Errorf("expected success=true, got %v", body)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// GetOrganizationByID — top-level route tests
// ─────────────────────────────────────────────────────────────────────────────

// TestGetOrganizationByID_NoAuth_TopRoute expects 401 when userID is absent.
func TestGetOrganizationByID_NoAuth_TopRoute(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newOrgTopApp()
	resp := testRequest(app, http.MethodGet, "/orgs/some-id", nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

// TestGetOrganizationByID_NotManager_TopRoute expects 403 when user has no
// admin membership for the requested org.
func TestGetOrganizationByID_NotManager_TopRoute(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupOrgMembersTable(t)

	app := newOrgTopApp(withTenantCtx(testOrgID, testUserID, testUserRole))
	// No member seeded → CanUserManageOrganization returns false → 403
	resp := testRequest(app, http.MethodGet, "/orgs/nonexistent-org", nil)
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403 for non-manager, got %d", resp.StatusCode)
	}
}

// TestGetOrganizationByID_AdminMember_TopRoute expects 200 when user IS an
// admin member (CanUserManageOrganization = true).
func TestGetOrganizationByID_AdminMember_TopRoute(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupOrgMembersTable(t)
	seedOrgWithAdminMember(t, testOrgID, testUserID)

	app := newOrgTopApp(withTenantCtx(testOrgID, testUserID, testUserRole))
	resp := testRequest(app, http.MethodGet, "/orgs/"+testOrgID, nil)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for admin member, got %d", resp.StatusCode)
	}
	body := decodeResponse(resp)
	if body["success"] != true {
		t.Errorf("expected success=true, got %v", body)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// SwitchOrganization — additional coverage via top-level route
// ─────────────────────────────────────────────────────────────────────────────

// TestSwitchOrganization_NoAuth_TopRoute expects 401 on top-level route.
func TestSwitchOrganization_NoAuth_TopRoute(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newOrgTopApp()
	resp := testRequest(app, http.MethodPost, "/orgs/some-id/switch", nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

// TestSwitchOrganization_ValidMember_TopRoute expects 200 when user is a
// member (via the top-level route wiring).
func TestSwitchOrganization_ValidMember_TopRoute(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupOrgMembersTable(t)
	seedOrgWithAdminMember(t, testOrgID, testUserID)

	// Seed the user row so SwitchOrganization's DB UPDATE on users succeeds
	db.Exec(`INSERT OR IGNORE INTO users (id, name, email, created_at, updated_at) VALUES (?, 'Test User', 'test@example.com', ?, ?)`,
		testUserID, time.Now(), time.Now())

	app := newOrgTopApp(withTenantCtx(testOrgID, testUserID, testUserRole))
	resp := testRequest(app, http.MethodPost, "/orgs/"+testOrgID+"/switch", nil)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for valid member switch (top route), got %d", resp.StatusCode)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// DeleteOrganization — additional coverage via top-level route
// ─────────────────────────────────────────────────────────────────────────────

// TestDeleteOrganization_NoAuth_TopRoute expects 401 on top-level route.
func TestDeleteOrganization_NoAuth_TopRoute(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newOrgTopApp()
	resp := testRequest(app, http.MethodDelete, "/orgs/some-id", nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

// TestDeleteOrganization_NotManager_TopRoute expects 403 for non-manager user.
func TestDeleteOrganization_NotManager_TopRoute(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupOrgMembersTable(t)

	app := newOrgTopApp(withTenantCtx(testOrgID, testUserID, testUserRole))
	resp := testRequest(app, http.MethodDelete, "/orgs/nonexistent-org", nil)
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403 for non-manager delete (top route), got %d", resp.StatusCode)
	}
}

// TestDeleteOrganization_AdminMember_TopRoute expects 200 when user IS admin.
func TestDeleteOrganization_AdminMember_TopRoute(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupOrgMembersTable(t)
	seedOrgWithAdminMember(t, testOrgID, testUserID)

	app := newOrgTopApp(withTenantCtx(testOrgID, testUserID, testUserRole))
	resp := testRequest(app, http.MethodDelete, "/orgs/"+testOrgID, nil)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for admin-member delete (top route), got %d", resp.StatusCode)
	}
	body := decodeResponse(resp)
	if body["success"] != true {
		t.Errorf("expected success=true, got %v", body)
	}
}

// setupOrgDepartmentsTable creates the organization_departments table needed
// by the paginated GetOrganizationMembers LEFT JOIN.
func setupOrgDepartmentsTable(t *testing.T) {
	t.Helper()
	sql := `CREATE TABLE IF NOT EXISTS organization_departments (
		id TEXT PRIMARY KEY,
		organization_id TEXT,
		name TEXT,
		created_at DATETIME,
		updated_at DATETIME
	)`
	if err := config.DB.Exec(sql).Error; err != nil {
		t.Fatalf("setupOrgDepartmentsTable: %v", err)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// GetOrganizationMembers — paginated / filtered paths
// (non-auth guard paths not covered by existing TestGetOrganizationMembers_*)
// ─────────────────────────────────────────────────────────────────────────────

// TestGetOrganizationMembers_PaginatedEmpty returns paginated wrapper with
// total=0 when no members belong to the org.
func TestGetOrganizationMembers_PaginatedEmpty(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupOrgMembersTable(t)
	setupOrgDepartmentsTable(t)

	app := newOrgApp(withTenantCtx(testOrgID, testUserID, testUserRole))
	resp := testRequest(app, http.MethodGet, "/org/members?page=1&page_size=10", nil)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for paginated empty members, got %d", resp.StatusCode)
	}
	body := decodeResponse(resp)
	if body["success"] != true {
		t.Errorf("expected success=true, got %v", body)
	}
	data, ok := body["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data to be a map, got %T", body["data"])
	}
	if data["total"] != float64(0) {
		t.Errorf("expected total=0, got %v", data["total"])
	}
}

// TestGetOrganizationMembers_SearchParam passes a search query and expects 200.
func TestGetOrganizationMembers_SearchParam(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupOrgMembersTable(t)
	setupOrgDepartmentsTable(t)

	app := newOrgApp(withTenantCtx(testOrgID, testUserID, testUserRole))
	resp := testRequest(app, http.MethodGet, "/org/members?search=alice", nil)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for search-param members, got %d", resp.StatusCode)
	}
}

// TestGetOrganizationMembers_RoleFilter passes role=admin and expects 200.
func TestGetOrganizationMembers_RoleFilter(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupOrgMembersTable(t)
	setupOrgDepartmentsTable(t)

	app := newOrgApp(withTenantCtx(testOrgID, testUserID, testUserRole))
	resp := testRequest(app, http.MethodGet, "/org/members?role=admin", nil)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for role-filter members, got %d", resp.StatusCode)
	}
}

// TestGetOrganizationMembers_ActiveFilter passes active=true and expects 200.
func TestGetOrganizationMembers_ActiveFilter(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupOrgMembersTable(t)
	setupOrgDepartmentsTable(t)

	app := newOrgApp(withTenantCtx(testOrgID, testUserID, testUserRole))
	resp := testRequest(app, http.MethodGet, "/org/members?active=true", nil)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for active-filter members, got %d", resp.StatusCode)
	}
}

// TestGetOrganizationMembers_InactiveFilter passes active=false and expects 200.
func TestGetOrganizationMembers_InactiveFilter(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupOrgMembersTable(t)
	setupOrgDepartmentsTable(t)

	app := newOrgApp(withTenantCtx(testOrgID, testUserID, testUserRole))
	resp := testRequest(app, http.MethodGet, "/org/members?active=false", nil)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for inactive-filter members, got %d", resp.StatusCode)
	}
}

// TestGetOrganizationMembers_AllFilters combines all filter params and expects 200.
func TestGetOrganizationMembers_AllFilters(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupOrgMembersTable(t)
	setupOrgDepartmentsTable(t)

	app := newOrgApp(withTenantCtx(testOrgID, testUserID, testUserRole))
	resp := testRequest(app, http.MethodGet, "/org/members?page=1&page_size=5&search=bob&role=requester&active=false", nil)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for all-filter members, got %d", resp.StatusCode)
	}
}

// TestGetOrganizationMembers_PaginatedWithData adds a member row and
// verifies total > 0 is returned correctly in paginated mode.
func TestGetOrganizationMembers_PaginatedWithData(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupOrgMembersTable(t)
	setupOrgDepartmentsTable(t)
	seedOrgWithAdminMember(t, testOrgID, testUserID)

	// Seed corresponding user row so the INNER JOIN on users doesn't drop the row
	db.Exec(`INSERT OR IGNORE INTO users (id, name, email, created_at, updated_at) VALUES (?, 'Paged User', 'paged@example.com', ?, ?)`,
		testUserID, time.Now(), time.Now())

	app := newOrgApp(withTenantCtx(testOrgID, testUserID, testUserRole))
	resp := testRequest(app, http.MethodGet, "/org/members?page=1&page_size=10", nil)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for paginated members with data, got %d", resp.StatusCode)
	}
	body := decodeResponse(resp)
	if body["success"] != true {
		t.Errorf("expected success=true, got %v", body)
	}
	data, ok := body["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data to be a map, got %T", body["data"])
	}
	if data["total"] == float64(0) {
		t.Errorf("expected total > 0 for seeded member, got 0")
	}
}
