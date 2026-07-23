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

// newPayeesApp builds a minimal Fiber app wired to the payee handlers.
// Any tenant middlewares passed are applied to all payee routes.
func newPayeesApp(tenantMiddleware ...fiber.Handler) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   err.Error(),
			})
		},
	})

	payees := app.Group("/payees")
	for _, mw := range tenantMiddleware {
		payees.Use(mw)
	}

	payees.Get("/", GetPayees)
	payees.Post("/", CreatePayee)
	payees.Get("/:id", GetPayee)
	payees.Put("/:id", UpdatePayee)
	payees.Delete("/:id", DeletePayee)

	return app
}

// validPayeePayload returns a map with required payee fields.
func validPayeePayload() map[string]interface{} {
	return map[string]interface{}{
		"payeeType":   "other",
		"name":        "John Doe",
		"email":       "jd@example.com",
		"bankName":    "FNB",
		"bankAccount": "1234567890",
	}
}

// seedPayee creates and inserts a Payee into the test DB.
func seedPayee(t *testing.T, db interface{ Create(value interface{}) *gorm.DB }, orgID, name, payeeType string) models.Payee {
	t.Helper()
	createdBy := testUserID
	p := models.Payee{
		ID:             uuid.New().String(),
		OrganizationID: orgID,
		PayeeType:      payeeType,
		Name:           name,
		Email:          name + "@example.com",
		CreatedBy:      &createdBy,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	if err := db.Create(&p).Error; err != nil {
		t.Fatalf("seedPayee: %v", err)
	}
	return p
}

// ---------------------------------------------------------------------------
// GET /payees  — no auth
// ---------------------------------------------------------------------------

func TestGetPayees_NoAuth(t *testing.T) {
	app := newPayeesApp()

	resp := testRequest(app, http.MethodGet, "/payees/", nil)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// POST + GET /payees — happy path
// ---------------------------------------------------------------------------

func TestPayees_CreateAndList(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	// Ensure payees table exists.
	assert.NoError(t, db.AutoMigrate(&models.Payee{}))

	app := newPayeesApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	// Create a payee.
	resp := testRequest(app, http.MethodPost, "/payees/", validPayeePayload())
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])
	data, ok := body["data"].(map[string]interface{})
	assert.True(t, ok, "data should be an object")
	assert.NotEmpty(t, data["id"])
	assert.Equal(t, "John Doe", data["name"])
	assert.Equal(t, "other", data["payeeType"])

	// List with ?type=other — the new payee must appear.
	resp2 := testRequest(app, http.MethodGet, "/payees/?type=other", nil)
	assert.Equal(t, http.StatusOK, resp2.StatusCode)

	body2 := decodeResponse(resp2)
	assert.Equal(t, true, body2["success"])
	items, ok2 := body2["data"].([]interface{})
	assert.True(t, ok2, "data should be a JSON array")
	assert.Len(t, items, 1)
	first := items[0].(map[string]interface{})
	assert.Equal(t, "John Doe", first["name"])
}

// ---------------------------------------------------------------------------
// GET /payees — org isolation
// ---------------------------------------------------------------------------

func TestGetPayees_OrgIsolation(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	assert.NoError(t, db.AutoMigrate(&models.Payee{}))

	seedPayee(t, db, testOrgID, "My Payee", "other")
	seedPayee(t, db, "other-org-999", "Other Org Payee", "other")

	app := newPayeesApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	resp := testRequest(app, http.MethodGet, "/payees/", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	items, ok := body["data"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, items, 1, "only payees for testOrgID should appear")
	first := items[0].(map[string]interface{})
	assert.Equal(t, "My Payee", first["name"])
}

// ---------------------------------------------------------------------------
// GET /payees — search by name (?q=)
// ---------------------------------------------------------------------------

func TestGetPayees_SearchByName(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	assert.NoError(t, db.AutoMigrate(&models.Payee{}))

	seedPayee(t, db, testOrgID, "Alice Smith", "other")
	seedPayee(t, db, testOrgID, "Bob Jones", "vendor")

	app := newPayeesApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	resp := testRequest(app, http.MethodGet, "/payees/?q=Alice", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	items, ok := body["data"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, items, 1, "only Alice should match")
	first := items[0].(map[string]interface{})
	assert.Equal(t, "Alice Smith", first["name"])
}

// ---------------------------------------------------------------------------
// DELETE /payees/:id — soft delete
// ---------------------------------------------------------------------------

func TestDeletePayee_SoftDelete(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	assert.NoError(t, db.AutoMigrate(&models.Payee{}))

	p := seedPayee(t, db, testOrgID, "To Delete", "other")

	app := newPayeesApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	// Delete the payee.
	resp := testRequest(app, http.MethodDelete, "/payees/"+p.ID, nil)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// List — must be absent.
	resp2 := testRequest(app, http.MethodGet, "/payees/", nil)
	assert.Equal(t, http.StatusOK, resp2.StatusCode)
	body2 := decodeResponse(resp2)
	items, _ := body2["data"].([]interface{})
	assert.Len(t, items, 0, "deleted payee should not appear in list")
}

// ---------------------------------------------------------------------------
// POST /payees — validation
// ---------------------------------------------------------------------------

func TestCreatePayee_NoAuth(t *testing.T) {
	app := newPayeesApp()

	resp := testRequest(app, http.MethodPost, "/payees/", validPayeePayload())
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestCreatePayee_MissingName(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	assert.NoError(t, db.AutoMigrate(&models.Payee{}))

	app := newPayeesApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	payload := validPayeePayload()
	delete(payload, "name")

	resp := testRequest(app, http.MethodPost, "/payees/", payload)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

func TestCreatePayee_InvalidPayeeType(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	assert.NoError(t, db.AutoMigrate(&models.Payee{}))

	app := newPayeesApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	payload := validPayeePayload()
	payload["payeeType"] = "alien"

	resp := testRequest(app, http.MethodPost, "/payees/", payload)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

// ---------------------------------------------------------------------------
// GET /payees/:id
// ---------------------------------------------------------------------------

func TestGetPayee_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	assert.NoError(t, db.AutoMigrate(&models.Payee{}))

	app := newPayeesApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	resp := testRequest(app, http.MethodGet, "/payees/nonexistent-id", nil)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

func TestGetPayee_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	assert.NoError(t, db.AutoMigrate(&models.Payee{}))

	p := seedPayee(t, db, testOrgID, "Single Payee", "employee")

	app := newPayeesApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	resp := testRequest(app, http.MethodGet, "/payees/"+p.ID, nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])
	data, ok := body["data"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, p.ID, data["id"])
	assert.Equal(t, "Single Payee", data["name"])
	assert.Equal(t, "employee", data["payeeType"])
}

// ---------------------------------------------------------------------------
// PUT /payees/:id
// ---------------------------------------------------------------------------

func TestUpdatePayee_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	assert.NoError(t, db.AutoMigrate(&models.Payee{}))

	app := newPayeesApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	resp := testRequest(app, http.MethodPut, "/payees/nonexistent-id", map[string]interface{}{"name": "New Name"})
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

func TestUpdatePayee_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	assert.NoError(t, db.AutoMigrate(&models.Payee{}))

	p := seedPayee(t, db, testOrgID, "Old Name", "other")

	app := newPayeesApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	resp := testRequest(app, http.MethodPut, "/payees/"+p.ID, map[string]interface{}{
		"name":     "Updated Name",
		"bankName": "Standard Bank",
	})
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])
	data, ok := body["data"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "Updated Name", data["name"])
	assert.Equal(t, "Standard Bank", data["bankName"])
	// payeeType unchanged
	assert.Equal(t, "other", data["payeeType"])
}

// ---------------------------------------------------------------------------
// Cross-org isolation — single-resource endpoints
// ---------------------------------------------------------------------------

func TestGetPayee_CrossOrgReturns404(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	assert.NoError(t, db.AutoMigrate(&models.Payee{}))

	// Payee belongs to org A.
	p := seedPayee(t, db, testOrgID, "Org A Payee", "vendor")

	// App authenticated as org B.
	app := newPayeesApp(withTenantCtx("other-org-999", testUserID, testUserRole))

	resp := testRequest(app, http.MethodGet, "/payees/"+p.ID, nil)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

func TestUpdatePayee_CrossOrgReturns404(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	assert.NoError(t, db.AutoMigrate(&models.Payee{}))

	// Payee belongs to org A.
	p := seedPayee(t, db, testOrgID, "Org A Payee", "vendor")

	// App authenticated as org B.
	app := newPayeesApp(withTenantCtx("other-org-999", testUserID, testUserRole))

	resp := testRequest(app, http.MethodPut, "/payees/"+p.ID, map[string]interface{}{"name": "Hijacked"})
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

func TestDeletePayee_CrossOrgReturns404(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	assert.NoError(t, db.AutoMigrate(&models.Payee{}))

	// Payee belongs to org A.
	p := seedPayee(t, db, testOrgID, "Org A Payee", "vendor")

	// App authenticated as org B.
	app := newPayeesApp(withTenantCtx("other-org-999", testUserID, testUserRole))

	resp := testRequest(app, http.MethodDelete, "/payees/"+p.ID, nil)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}
