package handlers

import (
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/models"
)

func setupOrgSettingsApp(db interface{}) *fiber.App {
	app := fiber.New()
	app.Put("/organization/settings",
		withTenantCtx(testOrgID, testUserID, testUserRole),
		UpdateOrganizationSettings,
	)
	return app
}

func TestUpdateOrganizationSettings_PersistsAutomationFields(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	if err := db.AutoMigrate(&models.OrganizationSettings{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	// Seed an existing settings row for testOrgID
	seed := models.OrganizationSettings{
		ID:             uuid.New().String(),
		OrganizationID: testOrgID,
		Currency:       "USD",
		FiscalYearStart: 1,
	}
	if err := db.Create(&seed).Error; err != nil {
		t.Fatalf("seed settings: %v", err)
	}

	app := fiber.New()
	app.Put("/organization/settings",
		withTenantCtx(testOrgID, testUserID, testUserRole),
		UpdateOrganizationSettings,
	)

	body := map[string]interface{}{
		"autoCreateGRNFromPO":  true,
		"grnAutomationLevel":   "auto_submit",
		"autoApproveMaxAmount": 5000.0,
	}

	resp := testRequest(app, http.MethodPut, "/organization/settings", body)
	if resp.StatusCode != http.StatusOK {
		decoded := decodeResponse(resp)
		t.Fatalf("expected 200, got %d; body=%v", resp.StatusCode, decoded)
	}

	// Reload the row and verify the 3 values persisted
	var updated models.OrganizationSettings
	if err := db.Where("organization_id = ?", testOrgID).First(&updated).Error; err != nil {
		t.Fatalf("reload settings: %v", err)
	}

	if !updated.AutoCreateGRNFromPO {
		t.Errorf("expected AutoCreateGRNFromPO=true, got false")
	}
	if updated.GRNAutomationLevel != "auto_submit" {
		t.Errorf("expected GRNAutomationLevel=%q, got %q", "auto_submit", updated.GRNAutomationLevel)
	}
	if updated.AutoApproveMaxAmount != 5000.0 {
		t.Errorf("expected AutoApproveMaxAmount=5000, got %v", updated.AutoApproveMaxAmount)
	}
}

func TestUpdateOrganizationSettings_RejectsInvalidAutomationLevel(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	if err := db.AutoMigrate(&models.OrganizationSettings{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	app := fiber.New()
	app.Put("/organization/settings",
		withTenantCtx(testOrgID, testUserID, testUserRole),
		UpdateOrganizationSettings,
	)

	body := map[string]interface{}{
		"grnAutomationLevel": "bogus",
	}

	resp := testRequest(app, http.MethodPut, "/organization/settings", body)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid grnAutomationLevel, got %d", resp.StatusCode)
	}
}
