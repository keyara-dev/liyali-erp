package handlers

import (
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/models"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// newBranchesApp builds a minimal Fiber app that mirrors the real routing for
// the branches endpoints.  The tenant middleware is injected via the
// withTenantCtx helper so we can control auth state per test.
func newBranchesApp(tenantMiddleware ...fiber.Handler) *fiber.App {
	app := fiber.New(fiber.Config{
		// Surface panics as 500 responses rather than crashing the test process.
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   err.Error(),
			})
		},
	})

	branches := app.Group("/branches")
	for _, mw := range tenantMiddleware {
		branches.Use(mw)
	}

	branches.Get("/", GetBranches)
	branches.Get("/:id", GetBranch)
	branches.Post("/", CreateBranch)
	branches.Put("/:id", UpdateBranch)
	branches.Delete("/:id", DeleteBranch)

	return app
}

// ---------------------------------------------------------------------------
// GET /branches
// ---------------------------------------------------------------------------

func TestGetBranches_NoAuth(t *testing.T) {
	// No tenant middleware → handler cannot obtain tenant context → 401.
	app := newBranchesApp()

	resp := testRequest(app, http.MethodGet, "/branches/", nil)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	body := decodeResponse(resp)
	assert.NotNil(t, body)
	assert.Equal(t, false, body["success"])
}

func TestGetBranches_Empty(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newBranchesApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	resp := testRequest(app, http.MethodGet, "/branches/", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])

	// data should be an empty array (or null), but success must be true.
	if data, ok := body["data"]; ok && data != nil {
		items, ok := data.([]interface{})
		assert.True(t, ok, "data should be an array")
		assert.Len(t, items, 0)
	}
}

func TestGetBranches_WithData(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Seed one branch for our org and one for a different org (should not appear).
	branch := models.OrganizationBranch{
		ID:             uuid.New().String(),
		OrganizationID: testOrgID,
		Name:           "Kitwe Branch",
		Code:           "KIT-001",
		ProvinceID:     "province-copperbelt",
		TownID:         "town-kitwe",
		Address:        "1 Obote Avenue, Kitwe",
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	assert.NoError(t, db.Create(&branch).Error)

	otherBranch := models.OrganizationBranch{
		ID:             uuid.New().String(),
		OrganizationID: "other-org-999",
		Name:           "Other Org Branch",
		Code:           "OTH-001",
		ProvinceID:     "province-x",
		TownID:         "town-x",
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	assert.NoError(t, db.Create(&otherBranch).Error)

	app := newBranchesApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	resp := testRequest(app, http.MethodGet, "/branches/", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])

	data, ok := body["data"].([]interface{})
	assert.True(t, ok, "data should be an array")
	assert.Len(t, data, 1, "only the branch belonging to testOrgID should be returned")

	first := data[0].(map[string]interface{})
	assert.Equal(t, "Kitwe Branch", first["name"])
	assert.Equal(t, "KIT-001", first["code"])
}

// ---------------------------------------------------------------------------
// GET /branches/:id
// ---------------------------------------------------------------------------

func TestGetBranch_NoAuth(t *testing.T) {
	app := newBranchesApp()

	resp := testRequest(app, http.MethodGet, "/branches/some-id", nil)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

func TestGetBranch_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newBranchesApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	resp := testRequest(app, http.MethodGet, "/branches/nonexistent-id", nil)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

func TestGetBranch_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	branch := models.OrganizationBranch{
		ID:             uuid.New().String(),
		OrganizationID: testOrgID,
		Name:           "Ndola Branch",
		Code:           "NDL-001",
		ProvinceID:     "province-copperbelt",
		TownID:         "town-ndola",
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	assert.NoError(t, db.Create(&branch).Error)

	app := newBranchesApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	resp := testRequest(app, http.MethodGet, "/branches/"+branch.ID, nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])

	data, ok := body["data"].(map[string]interface{})
	assert.True(t, ok, "data should be an object")
	assert.Equal(t, branch.ID, data["id"])
	assert.Equal(t, "Ndola Branch", data["name"])
	assert.Equal(t, "NDL-001", data["code"])
}

// ---------------------------------------------------------------------------
// POST /branches
// ---------------------------------------------------------------------------

func TestCreateBranch_NoAuth(t *testing.T) {
	app := newBranchesApp()

	payload := map[string]interface{}{
		"name":        "Test Branch",
		"code":        "TST-001",
		"province_id": "province-lusaka",
		"town_id":     "town-lusaka",
	}

	resp := testRequest(app, http.MethodPost, "/branches/", payload)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

func TestCreateBranch_MissingName(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newBranchesApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	payload := map[string]interface{}{
		// name intentionally omitted
		"code":        "TST-001",
		"province_id": "province-lusaka",
		"town_id":     "town-lusaka",
	}

	resp := testRequest(app, http.MethodPost, "/branches/", payload)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
	assert.NotEmpty(t, body["message"])
}

func TestCreateBranch_MissingCode(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newBranchesApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	payload := map[string]interface{}{
		"name":        "Test Branch",
		// code intentionally omitted
		"province_id": "province-lusaka",
		"town_id":     "town-lusaka",
	}

	resp := testRequest(app, http.MethodPost, "/branches/", payload)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

func TestCreateBranch_MissingTownOrProvince(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newBranchesApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	// Missing both town_id and province_id.
	payload := map[string]interface{}{
		"name": "Test Branch",
		"code": "TST-001",
		// province_id and town_id intentionally omitted
	}

	resp := testRequest(app, http.MethodPost, "/branches/", payload)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])

	// Missing only province_id (town_id present) should also fail.
	payload2 := map[string]interface{}{
		"name":    "Test Branch",
		"code":    "TST-002",
		"town_id": "town-lusaka",
		// province_id intentionally omitted
	}

	resp2 := testRequest(app, http.MethodPost, "/branches/", payload2)
	assert.Equal(t, http.StatusBadRequest, resp2.StatusCode)

	// Missing only town_id (province_id present) should also fail.
	payload3 := map[string]interface{}{
		"name":        "Test Branch",
		"code":        "TST-003",
		"province_id": "province-lusaka",
		// town_id intentionally omitted
	}

	resp3 := testRequest(app, http.MethodPost, "/branches/", payload3)
	assert.Equal(t, http.StatusBadRequest, resp3.StatusCode)
}

func TestCreateBranch_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newBranchesApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	payload := map[string]interface{}{
		"name":        "Lusaka HQ",
		"code":        "LSK-HQ",
		"province_id": "province-lusaka",
		"town_id":     "town-lusaka-central",
		"address":     "Independence Avenue, Lusaka",
	}

	resp := testRequest(app, http.MethodPost, "/branches/", payload)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])

	data, ok := body["data"].(map[string]interface{})
	assert.True(t, ok, "data should be an object")
	assert.Equal(t, "Lusaka HQ", data["name"])
	assert.Equal(t, "LSK-HQ", data["code"])
	assert.NotEmpty(t, data["id"])

	// Verify the record was persisted.
	var count int64
	db.Model(&models.OrganizationBranch{}).
		Where("organization_id = ? AND code = ?", testOrgID, "LSK-HQ").
		Count(&count)
	assert.Equal(t, int64(1), count)
}

func TestCreateBranch_DuplicateCode(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Pre-create a branch with the same code.
	existing := models.OrganizationBranch{
		ID:             uuid.New().String(),
		OrganizationID: testOrgID,
		Name:           "First Branch",
		Code:           "DUP-001",
		ProvinceID:     "province-lusaka",
		TownID:         "town-lusaka",
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	assert.NoError(t, db.Create(&existing).Error)

	app := newBranchesApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	payload := map[string]interface{}{
		"name":        "Second Branch (same code)",
		"code":        "DUP-001",
		"province_id": "province-lusaka",
		"town_id":     "town-lusaka",
	}

	resp := testRequest(app, http.MethodPost, "/branches/", payload)
	assert.Equal(t, http.StatusConflict, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

// ---------------------------------------------------------------------------
// PUT /branches/:id
// ---------------------------------------------------------------------------

func TestUpdateBranch_NoAuth(t *testing.T) {
	app := newBranchesApp()

	payload := map[string]interface{}{
		"name": "Updated Name",
	}

	resp := testRequest(app, http.MethodPut, "/branches/some-id", payload)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

func TestUpdateBranch_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newBranchesApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	payload := map[string]interface{}{
		"name": "Updated Name",
	}

	resp := testRequest(app, http.MethodPut, "/branches/nonexistent-id", payload)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

func TestUpdateBranch_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Seed a branch to update.
	branch := models.OrganizationBranch{
		ID:             uuid.New().String(),
		OrganizationID: testOrgID,
		Name:           "Old Name",
		Code:           "OLD-001",
		ProvinceID:     "province-lusaka",
		TownID:         "town-lusaka",
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	assert.NoError(t, db.Create(&branch).Error)

	app := newBranchesApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	isActive := false
	payload := map[string]interface{}{
		"name":      "New Name",
		"is_active": isActive,
	}

	resp := testRequest(app, http.MethodPut, "/branches/"+branch.ID, payload)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])

	data, ok := body["data"].(map[string]interface{})
	assert.True(t, ok, "data should be an object")
	assert.Equal(t, "New Name", data["name"])
	// isActive was set to false in the update.
	assert.Equal(t, false, data["isActive"])
}

// ---------------------------------------------------------------------------
// DELETE /branches/:id
// ---------------------------------------------------------------------------

func TestDeleteBranch_NoAuth(t *testing.T) {
	app := newBranchesApp()

	resp := testRequest(app, http.MethodDelete, "/branches/some-id", nil)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

func TestDeleteBranch_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newBranchesApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	resp := testRequest(app, http.MethodDelete, "/branches/nonexistent-id", nil)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

func TestDeleteBranch_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Seed a branch to delete.
	branch := models.OrganizationBranch{
		ID:             uuid.New().String(),
		OrganizationID: testOrgID,
		Name:           "Branch To Delete",
		Code:           "DEL-001",
		ProvinceID:     "province-lusaka",
		TownID:         "town-lusaka",
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	assert.NoError(t, db.Create(&branch).Error)

	app := newBranchesApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	resp := testRequest(app, http.MethodDelete, "/branches/"+branch.ID, nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])
	assert.Equal(t, "Branch deleted successfully", body["message"])

	// Verify the record was removed from the database.
	var count int64
	db.Model(&models.OrganizationBranch{}).
		Where("id = ?", branch.ID).
		Count(&count)
	assert.Equal(t, int64(0), count)
}
