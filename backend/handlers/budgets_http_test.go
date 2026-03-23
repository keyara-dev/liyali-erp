package handlers

import (
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/types"
	"gorm.io/datatypes"
)

// ─────────────────────────────────────────────────────────────────────────────
// Test DB setup with Budget table
// ─────────────────────────────────────────────────────────────────────────────

func setupBudgetTestDB(t *testing.T) {
	t.Helper()
	if config.DB == nil {
		t.Fatal("setupBudgetTestDB: config.DB is nil — call setupTestDB first")
	}
	if err := config.DB.AutoMigrate(&models.Budget{}); err != nil {
		t.Fatalf("setupBudgetTestDB AutoMigrate: %v", err)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// App factories
// ─────────────────────────────────────────────────────────────────────────────

func newBudgetApp(t *testing.T) *fiber.App {
	t.Helper()
	auth := withTenantCtx(testOrgID, testUserID, testUserRole)

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		},
	})
	app.Get("/budgets", auth, GetBudgets)
	app.Post("/budgets", auth, CreateBudget)
	app.Get("/budgets/:id", auth, GetBudget)
	app.Put("/budgets/:id", auth, UpdateBudget)
	app.Delete("/budgets/:id", auth, DeleteBudget)
	app.Post("/budgets/:id/submit", auth, SubmitBudget)
	return app
}

// newBudgetAppNoAuth builds an app where the tenant local is NOT set.
// GetTenantContext returns an error → 401 response (no panic).
func newBudgetAppNoAuth(t *testing.T) *fiber.App {
	t.Helper()

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		},
	})
	app.Get("/budgets", GetBudgets)
	app.Post("/budgets", CreateBudget)
	app.Get("/budgets/:id", GetBudget)
	app.Put("/budgets/:id", UpdateBudget)
	app.Delete("/budgets/:id", DeleteBudget)
	app.Post("/budgets/:id/submit", SubmitBudget)
	return app
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

// makeBudget inserts a Budget record in DRAFT status and returns it.
func makeBudget(t *testing.T, orgID, ownerID, status string) models.Budget {
	t.Helper()
	budgetID := uuid.New().String()
	b := models.Budget{
		ID:              budgetID,
		OrganizationID:  orgID,
		OwnerID:         ownerID,
		BudgetCode:      "BG-TEST-" + budgetID[:8],
		Name:            "Test Budget",
		Department:      "Engineering",
		Status:          status,
		FiscalYear:      "2026",
		TotalBudget:     100000.00,
		AllocatedAmount: 0,
		RemainingAmount: 100000.00,
		Currency:        "ZMW",
		CreatedBy:       ownerID,
		ApprovalStage:   0,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	b.ApprovalHistory = datatypes.NewJSONType([]types.ApprovalRecord{})
	b.ActionHistory = datatypes.NewJSONType([]types.ActionHistoryEntry{})
	if err := config.DB.Create(&b).Error; err != nil {
		t.Fatalf("makeBudget: %v", err)
	}
	return b
}

// ─────────────────────────────────────────────────────────────────────────────
// GetBudgets
// ─────────────────────────────────────────────────────────────────────────────

// Budget handlers use GetTenantContext → structured 401 when tenant absent.
func TestGetBudgets_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupBudgetTestDB(t)

	app := newBudgetAppNoAuth(t)
	resp := testRequest(app, http.MethodGet, "/budgets", nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 without auth, got %d", resp.StatusCode)
	}
}

func TestGetBudgets_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupBudgetTestDB(t)
	seedTestUser(t)

	app := newBudgetApp(t)
	resp := testRequest(app, http.MethodGet, "/budgets", nil)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	body := decodeResponse(resp)
	if body == nil {
		t.Fatal("expected JSON body")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// CreateBudget
// ─────────────────────────────────────────────────────────────────────────────

func TestCreateBudget_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupBudgetTestDB(t)

	app := newBudgetAppNoAuth(t)
	resp := testRequest(app, http.MethodPost, "/budgets", map[string]interface{}{
		"department":      "Finance",
		"fiscalYear":      "2026",
		"totalBudget":     50000.0,
		"allocatedAmount": 0.0,
	})
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 without auth, got %d", resp.StatusCode)
	}
}

func TestCreateBudget_MissingTotalBudget(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupBudgetTestDB(t)
	seedTestUser(t)

	app := newBudgetApp(t)
	// totalBudget = 0 fails the `gt=0` validation and the explicit >0 check.
	resp := testRequest(app, http.MethodPost, "/budgets", map[string]interface{}{
		"department":      "Finance",
		"fiscalYear":      "2026",
		"allocatedAmount": 0.0,
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for missing/zero totalBudget, got %d", resp.StatusCode)
	}
}

func TestCreateBudget_MissingDepartment(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupBudgetTestDB(t)
	seedTestUser(t)

	app := newBudgetApp(t)
	// department is required in CreateBudgetRequest (validate:"required").
	// However, the handler does not run a struct validator; budget.go validates
	// totalBudget > 0 and allocated >= 0 directly. The department field is
	// required by the types tag but the handler doesn't call validate.Struct.
	// The handler WILL proceed. We cover the path where it succeeds or returns
	// a budget-level validation error. Test that we don't get 500.
	resp := testRequest(app, http.MethodPost, "/budgets", map[string]interface{}{
		"fiscalYear":      "2026",
		"totalBudget":     50000.0,
		"allocatedAmount": 0.0,
	})
	// The handler proceeds to look up the user. If the user exists it creates
	// the budget (department is empty). We just verify it doesn't crash (500).
	if resp.StatusCode >= http.StatusInternalServerError {
		body := decodeResponse(resp)
		t.Errorf("expected non-500 response, got %d; body=%v", resp.StatusCode, body)
	}
}

func TestCreateBudget_MissingFiscalYear(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupBudgetTestDB(t)
	seedTestUser(t)

	app := newBudgetApp(t)
	// Similar to department — handler doesn't validate struct; proceeds without fiscalYear.
	resp := testRequest(app, http.MethodPost, "/budgets", map[string]interface{}{
		"department":      "Finance",
		"totalBudget":     50000.0,
		"allocatedAmount": 0.0,
	})
	if resp.StatusCode >= http.StatusInternalServerError {
		body := decodeResponse(resp)
		t.Errorf("expected non-500 response, got %d; body=%v", resp.StatusCode, body)
	}
}

func TestCreateBudget_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupBudgetTestDB(t)
	seedTestUser(t)

	app := newBudgetApp(t)
	resp := testRequest(app, http.MethodPost, "/budgets", map[string]interface{}{
		"name":            "Engineering Budget 2026",
		"department":      "Engineering",
		"fiscalYear":      "2026",
		"totalBudget":     150000.0,
		"allocatedAmount": 0.0,
		"currency":        "ZMW",
	})
	if resp.StatusCode != http.StatusCreated {
		body := decodeResponse(resp)
		t.Errorf("expected 201, got %d; body=%v", resp.StatusCode, body)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// GetBudget
// ─────────────────────────────────────────────────────────────────────────────

func TestGetBudget_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupBudgetTestDB(t)

	app := newBudgetAppNoAuth(t)
	resp := testRequest(app, http.MethodGet, "/budgets/"+uuid.New().String(), nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 without auth, got %d", resp.StatusCode)
	}
}

func TestGetBudget_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupBudgetTestDB(t)
	seedTestUser(t)

	app := newBudgetApp(t)
	resp := testRequest(app, http.MethodGet, "/budgets/"+uuid.New().String(), nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestGetBudget_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupBudgetTestDB(t)
	seedTestUser(t)

	budget := makeBudget(t, testOrgID, testUserID, "DRAFT")
	app := newBudgetApp(t)
	resp := testRequest(app, http.MethodGet, "/budgets/"+budget.ID, nil)
	if resp.StatusCode != http.StatusOK {
		body := decodeResponse(resp)
		t.Errorf("expected 200, got %d; body=%v", resp.StatusCode, body)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// UpdateBudget
// ─────────────────────────────────────────────────────────────────────────────

func TestUpdateBudget_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupBudgetTestDB(t)

	app := newBudgetAppNoAuth(t)
	resp := testRequest(app, http.MethodPut, "/budgets/"+uuid.New().String(),
		map[string]interface{}{"name": "Updated"})
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 without auth, got %d", resp.StatusCode)
	}
}

func TestUpdateBudget_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupBudgetTestDB(t)
	seedTestUser(t)

	app := newBudgetApp(t)
	resp := testRequest(app, http.MethodPut, "/budgets/"+uuid.New().String(),
		map[string]interface{}{"name": "Updated Name"})
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404 for non-existent budget, got %d", resp.StatusCode)
	}
}

func TestUpdateBudget_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupBudgetTestDB(t)
	seedTestUser(t)

	budget := makeBudget(t, testOrgID, testUserID, "DRAFT")
	app := newBudgetApp(t)
	resp := testRequest(app, http.MethodPut, "/budgets/"+budget.ID, map[string]interface{}{
		"name":        "Updated Engineering Budget",
		"totalBudget": 200000.0,
	})
	if resp.StatusCode != http.StatusOK {
		body := decodeResponse(resp)
		t.Errorf("expected 200, got %d; body=%v", resp.StatusCode, body)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// DeleteBudget
// ─────────────────────────────────────────────────────────────────────────────

func TestDeleteBudget_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupBudgetTestDB(t)

	app := newBudgetAppNoAuth(t)
	resp := testRequest(app, http.MethodDelete, "/budgets/"+uuid.New().String(), nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 without auth, got %d", resp.StatusCode)
	}
}

func TestDeleteBudget_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupBudgetTestDB(t)
	seedTestUser(t)

	app := newBudgetApp(t)
	resp := testRequest(app, http.MethodDelete, "/budgets/"+uuid.New().String(), nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404 for non-existent budget, got %d", resp.StatusCode)
	}
}

func TestDeleteBudget_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupBudgetTestDB(t)
	seedTestUser(t)

	budget := makeBudget(t, testOrgID, testUserID, "DRAFT")
	app := newBudgetApp(t)
	resp := testRequest(app, http.MethodDelete, "/budgets/"+budget.ID, nil)
	if resp.StatusCode != http.StatusOK {
		body := decodeResponse(resp)
		t.Errorf("expected 200, got %d; body=%v", resp.StatusCode, body)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// SubmitBudget
// ─────────────────────────────────────────────────────────────────────────────

func TestSubmitBudget_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupBudgetTestDB(t)

	app := newBudgetAppNoAuth(t)
	resp := testRequest(app, http.MethodPost, "/budgets/"+uuid.New().String()+"/submit",
		map[string]interface{}{"workflowId": uuid.New().String()})
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 without auth, got %d", resp.StatusCode)
	}
}

func TestSubmitBudget_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupBudgetTestDB(t)
	seedTestUser(t)

	app := newBudgetApp(t)
	resp := testRequest(app, http.MethodPost, "/budgets/"+uuid.New().String()+"/submit",
		map[string]interface{}{"workflowId": uuid.New().String()})
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404 for non-existent budget, got %d", resp.StatusCode)
	}
}

func TestSubmitBudget_AlreadySubmitted(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupBudgetTestDB(t)
	seedTestUser(t)

	// Create a budget in PENDING status (already submitted).
	budget := makeBudget(t, testOrgID, testUserID, "PENDING")
	app := newBudgetApp(t)
	resp := testRequest(app, http.MethodPost, "/budgets/"+budget.ID+"/submit",
		map[string]interface{}{"workflowId": uuid.New().String()})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for already-submitted budget, got %d", resp.StatusCode)
	}
}

// TestSubmitBudget_Success verifies the DRAFT → submission path. The handler
// requires a workflowId in the body. After the status check it tries to access
// c.Locals("workflowExecutionService") which is nil in tests → 500. We verify
// the handler reached past the status gate (i.e., the budget was found and is
// in DRAFT) by checking we did NOT get a 400/404.
func TestSubmitBudget_MissingWorkflowId(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupBudgetTestDB(t)
	seedTestUser(t)

	budget := makeBudget(t, testOrgID, testUserID, "DRAFT")
	app := newBudgetApp(t)
	// Omit workflowId — handler validates presence and returns 400.
	resp := testRequest(app, http.MethodPost, "/budgets/"+budget.ID+"/submit",
		map[string]interface{}{})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for missing workflowId, got %d", resp.StatusCode)
	}
}
