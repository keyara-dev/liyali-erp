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

// newVendorsApp builds a minimal Fiber app wired to the vendor handlers.
// Any tenant middlewares passed are applied to all vendor routes.
func newVendorsApp(tenantMiddleware ...fiber.Handler) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   err.Error(),
			})
		},
	})

	vendors := app.Group("/vendors")
	for _, mw := range tenantMiddleware {
		vendors.Use(mw)
	}

	vendors.Get("/", GetVendors)
	vendors.Post("/", CreateVendor)
	vendors.Get("/:id", GetVendor)
	vendors.Put("/:id", UpdateVendor)

	return app
}

// validVendorPayload returns a map with all required vendor fields.
func validVendorPayload() map[string]interface{} {
	return map[string]interface{}{
		"name":        "Acme Supplies Ltd",
		"email":       "acme@example.com",
		"phone":       "+260971000001",
		"country":     "Zambia",
		"city":        "Lusaka",
		"bankAccount": "ACC-1234567",
		"taxId":       "TAX-9999",
	}
}

// ---------------------------------------------------------------------------
// GET /vendors
// ---------------------------------------------------------------------------

func TestGetVendors_NoAuth(t *testing.T) {
	// No tenant middleware → handler returns 401.
	app := newVendorsApp()

	resp := testRequest(app, http.MethodGet, "/vendors/", nil)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestGetVendors_Empty(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newVendorsApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	resp := testRequest(app, http.MethodGet, "/vendors/", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])

	data, ok := body["data"].([]interface{})
	assert.True(t, ok, "data should be a JSON array")
	assert.Len(t, data, 0)
}

func TestGetVendors_WithData(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Seed one vendor for testOrgID.
	myVendor := models.Vendor{
		ID:             uuid.New().String(),
		OrganizationID: testOrgID,
		VendorCode:     "VND-AAA",
		Name:           "My Vendor",
		Email:          "myvendor@example.com",
		Phone:          "+260971111111",
		Country:        "Zambia",
		City:           "Lusaka",
		BankAccount:    "ACC-MY",
		TaxID:          "TAX-MY",
		Active:         true,
		CreatedBy:      testUserID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	assert.NoError(t, db.Create(&myVendor).Error)

	// Seed one vendor for a different org — must NOT appear in the response.
	otherVendor := models.Vendor{
		ID:             uuid.New().String(),
		OrganizationID: "other-org-999",
		VendorCode:     "VND-BBB",
		Name:           "Other Org Vendor",
		Email:          "other@example.com",
		Phone:          "+260972222222",
		Country:        "Zambia",
		City:           "Ndola",
		BankAccount:    "ACC-OTHER",
		TaxID:          "TAX-OTHER",
		Active:         true,
		CreatedBy:      "other-user",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	assert.NoError(t, db.Create(&otherVendor).Error)

	app := newVendorsApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	resp := testRequest(app, http.MethodGet, "/vendors/", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])

	data, ok := body["data"].([]interface{})
	assert.True(t, ok, "data should be a JSON array")
	assert.Len(t, data, 1, "only vendors belonging to testOrgID should be returned")

	first := data[0].(map[string]interface{})
	assert.Equal(t, "My Vendor", first["name"])
	assert.Equal(t, "myvendor@example.com", first["email"])
}

// ---------------------------------------------------------------------------
// GET /vendors/:id
// ---------------------------------------------------------------------------

func TestGetVendor_NoAuth(t *testing.T) {
	app := newVendorsApp()

	resp := testRequest(app, http.MethodGet, "/vendors/some-id", nil)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestGetVendor_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newVendorsApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	resp := testRequest(app, http.MethodGet, "/vendors/nonexistent-id", nil)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

func TestGetVendor_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	vendor := models.Vendor{
		ID:             uuid.New().String(),
		OrganizationID: testOrgID,
		VendorCode:     "VND-GET-001",
		Name:           "Get Vendor",
		Email:          "getvendor@example.com",
		Phone:          "+260973333333",
		Country:        "Zambia",
		City:           "Kitwe",
		BankAccount:    "ACC-GET",
		TaxID:          "TAX-GET",
		Active:         true,
		CreatedBy:      testUserID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	assert.NoError(t, db.Create(&vendor).Error)

	app := newVendorsApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	resp := testRequest(app, http.MethodGet, "/vendors/"+vendor.ID, nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])

	data, ok := body["data"].(map[string]interface{})
	assert.True(t, ok, "data should be an object")
	assert.Equal(t, vendor.ID, data["id"])
	assert.Equal(t, "Get Vendor", data["name"])
	assert.Equal(t, "getvendor@example.com", data["email"])
	assert.Equal(t, "VND-GET-001", data["vendorCode"])
}

// ---------------------------------------------------------------------------
// POST /vendors
// ---------------------------------------------------------------------------

func TestCreateVendor_NoAuth(t *testing.T) {
	app := newVendorsApp()

	resp := testRequest(app, http.MethodPost, "/vendors/", validVendorPayload())
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestCreateVendor_MissingName(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newVendorsApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	payload := validVendorPayload()
	delete(payload, "name")

	resp := testRequest(app, http.MethodPost, "/vendors/", payload)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
	assert.NotEmpty(t, body["message"])
}

func TestCreateVendor_NameTooShort(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newVendorsApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	// Name is only 2 characters — below the 3-char minimum.
	payload := validVendorPayload()
	payload["name"] = "AB"

	resp := testRequest(app, http.MethodPost, "/vendors/", payload)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

func TestCreateVendor_MissingEmail(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newVendorsApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	payload := validVendorPayload()
	delete(payload, "email")

	resp := testRequest(app, http.MethodPost, "/vendors/", payload)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

func TestCreateVendor_InvalidEmail(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newVendorsApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	// Email without "@" — fails basic validation in the handler.
	payload := validVendorPayload()
	payload["email"] = "notanemail"

	resp := testRequest(app, http.MethodPost, "/vendors/", payload)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])

	// Email too short (< 5 chars) — also invalid.
	payload2 := validVendorPayload()
	payload2["email"] = "a@b"

	resp2 := testRequest(app, http.MethodPost, "/vendors/", payload2)
	assert.Equal(t, http.StatusBadRequest, resp2.StatusCode)
}

func TestCreateVendor_MissingPhone(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newVendorsApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	payload := validVendorPayload()
	delete(payload, "phone")

	resp := testRequest(app, http.MethodPost, "/vendors/", payload)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

func TestCreateVendor_MissingCountry(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newVendorsApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	payload := validVendorPayload()
	delete(payload, "country")

	resp := testRequest(app, http.MethodPost, "/vendors/", payload)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

func TestCreateVendor_MissingCity(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newVendorsApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	payload := validVendorPayload()
	delete(payload, "city")

	resp := testRequest(app, http.MethodPost, "/vendors/", payload)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

func TestCreateVendor_MissingBankAccount(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newVendorsApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	payload := validVendorPayload()
	delete(payload, "bankAccount")

	resp := testRequest(app, http.MethodPost, "/vendors/", payload)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

func TestCreateVendor_MissingTaxID(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newVendorsApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	payload := validVendorPayload()
	delete(payload, "taxId")

	resp := testRequest(app, http.MethodPost, "/vendors/", payload)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

func TestCreateVendor_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newVendorsApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	resp := testRequest(app, http.MethodPost, "/vendors/", validVendorPayload())
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])

	data, ok := body["data"].(map[string]interface{})
	assert.True(t, ok, "data should be an object")
	assert.Equal(t, "Acme Supplies Ltd", data["name"])
	assert.Equal(t, "acme@example.com", data["email"])
	assert.Equal(t, "Zambia", data["country"])
	assert.Equal(t, "Lusaka", data["city"])
	assert.NotEmpty(t, data["id"])
	assert.NotEmpty(t, data["vendorCode"])
	assert.Equal(t, true, data["active"])

	// Verify persisted to DB.
	var count int64
	db.Model(&models.Vendor{}).
		Where("organization_id = ? AND email = ?", testOrgID, "acme@example.com").
		Count(&count)
	assert.Equal(t, int64(1), count)
}

func TestCreateVendor_DuplicateEmail(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Pre-create a vendor with the same email in testOrgID.
	existing := models.Vendor{
		ID:             uuid.New().String(),
		OrganizationID: testOrgID,
		VendorCode:     "VND-DUP-001",
		Name:           "First Vendor",
		Email:          "duplicate@example.com",
		Phone:          "+260974444444",
		Country:        "Zambia",
		City:           "Lusaka",
		BankAccount:    "ACC-DUP",
		TaxID:          "TAX-DUP",
		Active:         true,
		CreatedBy:      testUserID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	assert.NoError(t, db.Create(&existing).Error)

	app := newVendorsApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	payload := validVendorPayload()
	payload["email"] = "duplicate@example.com"

	resp := testRequest(app, http.MethodPost, "/vendors/", payload)
	assert.Equal(t, http.StatusConflict, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

// ---------------------------------------------------------------------------
// PUT /vendors/:id
// ---------------------------------------------------------------------------

func TestUpdateVendor_NoAuth(t *testing.T) {
	app := newVendorsApp()

	resp := testRequest(app, http.MethodPut, "/vendors/some-id", map[string]interface{}{"name": "New Name"})
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestUpdateVendor_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newVendorsApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	resp := testRequest(app, http.MethodPut, "/vendors/nonexistent-id", map[string]interface{}{"name": "New Name"})
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

func TestUpdateVendor_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	vendor := models.Vendor{
		ID:             uuid.New().String(),
		OrganizationID: testOrgID,
		VendorCode:     "VND-UPD-001",
		Name:           "Old Vendor Name",
		Email:          "oldvendor@example.com",
		Phone:          "+260975555555",
		Country:        "Zambia",
		City:           "Lusaka",
		BankAccount:    "ACC-OLD",
		TaxID:          "TAX-OLD",
		Active:         true,
		CreatedBy:      testUserID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	assert.NoError(t, db.Create(&vendor).Error)

	app := newVendorsApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	updatePayload := map[string]interface{}{
		"name":  "Updated Vendor Name",
		"city":  "Ndola",
		"phone": "+260976666666",
	}

	resp := testRequest(app, http.MethodPut, "/vendors/"+vendor.ID, updatePayload)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])

	data, ok := body["data"].(map[string]interface{})
	assert.True(t, ok, "data should be an object")
	assert.Equal(t, "Updated Vendor Name", data["name"])
	assert.Equal(t, "Ndola", data["city"])
	assert.Equal(t, "+260976666666", data["phone"])

	// Fields not in the update payload should be unchanged.
	assert.Equal(t, "oldvendor@example.com", data["email"])
}

func TestUpdateVendor_NameTooShort(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	vendor := models.Vendor{
		ID:             uuid.New().String(),
		OrganizationID: testOrgID,
		VendorCode:     "VND-SHORT-001",
		Name:           "Valid Vendor",
		Email:          "shorttest@example.com",
		Phone:          "+260977777777",
		Country:        "Zambia",
		City:           "Lusaka",
		BankAccount:    "ACC-SHORT",
		TaxID:          "TAX-SHORT",
		Active:         true,
		CreatedBy:      testUserID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	assert.NoError(t, db.Create(&vendor).Error)

	app := newVendorsApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	// Attempt to set name to a 2-character string.
	resp := testRequest(app, http.MethodPut, "/vendors/"+vendor.ID, map[string]interface{}{"name": "AB"})
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

func TestUpdateVendor_DuplicateEmail(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Create two vendors.
	v1 := models.Vendor{
		ID:             uuid.New().String(),
		OrganizationID: testOrgID,
		VendorCode:     "VND-CON-001",
		Name:           "Vendor One",
		Email:          "vendorone@example.com",
		Phone:          "+260978000001",
		Country:        "Zambia",
		City:           "Lusaka",
		BankAccount:    "ACC-V1",
		TaxID:          "TAX-V1",
		Active:         true,
		CreatedBy:      testUserID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	assert.NoError(t, db.Create(&v1).Error)

	v2 := models.Vendor{
		ID:             uuid.New().String(),
		OrganizationID: testOrgID,
		VendorCode:     "VND-CON-002",
		Name:           "Vendor Two",
		Email:          "vendortwo@example.com",
		Phone:          "+260978000002",
		Country:        "Zambia",
		City:           "Ndola",
		BankAccount:    "ACC-V2",
		TaxID:          "TAX-V2",
		Active:         true,
		CreatedBy:      testUserID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	assert.NoError(t, db.Create(&v2).Error)

	app := newVendorsApp(withTenantCtx(testOrgID, testUserID, testUserRole))

	// Try to update v2 to have the same email as v1.
	resp := testRequest(app, http.MethodPut, "/vendors/"+v2.ID, map[string]interface{}{
		"email": "vendorone@example.com",
	})
	assert.Equal(t, http.StatusConflict, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}
