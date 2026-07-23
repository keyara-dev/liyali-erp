package handlers

import (
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/models"
)

// makeScopedBudget creates a DRAFT budget owned by ownerID.
func makeScopedBudget(t *testing.T, ownerID, docCode string) models.Budget {
	t.Helper()
	b := models.Budget{
		ID:             uuid.New().String(),
		OrganizationID: testOrgID,
		BudgetCode:     docCode,
		Department:     "IT",
		Status:         "DRAFT",
		FiscalYear:     "2026",
		TotalBudget:    100000,
		CreatedBy:      ownerID,
		OwnerID:        ownerID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	if err := config.DB.Create(&b).Error; err != nil {
		t.Fatalf("create budget: %v", err)
	}
	return b
}

func TestUpdateBudget_NonOwnerScopedOut(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	if err := db.AutoMigrate(&models.Budget{}); err != nil {
		t.Fatalf("migrate Budget: %v", err)
	}
	setupWorkflowTasksTable(t, db) // GetDocumentScope subquery targets this table

	b := makeScopedBudget(t, testUserID, "BUD-IDOR-1")

	app := fiber.New()
	auth := withTenantCtx(testOrgID, "other-user-002", "requester")
	app.Put("/budgets/:id", auth, UpdateBudget)

	body := map[string]interface{}{"department": "IT", "totalBudget": 50000}
	resp := testRequest(app, http.MethodPut, "/budgets/"+b.ID, body)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("non-owner: expected 404 (scoped out), got %d", resp.StatusCode)
	}
}

func TestUpdateBudget_OwnerPassesScope(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	if err := db.AutoMigrate(&models.Budget{}); err != nil {
		t.Fatalf("migrate Budget: %v", err)
	}
	setupWorkflowTasksTable(t, db)

	b := makeScopedBudget(t, "owner-user-003", "BUD-IDOR-2")

	app := fiber.New()
	auth := withTenantCtx(testOrgID, "owner-user-003", "requester")
	app.Put("/budgets/:id", auth, UpdateBudget)

	body := map[string]interface{}{"department": "IT", "totalBudget": 50000}
	resp := testRequest(app, http.MethodPut, "/budgets/"+b.ID, body)
	// Owner must pass the ownership scope (i.e. not be scoped out / 404).
	if resp.StatusCode == http.StatusNotFound {
		t.Fatalf("owner: expected to pass scope, got 404")
	}
}
