package unit

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/handlers"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/types"
	"gorm.io/gorm"
)

// TestCreateCategory tests category creation
func TestCreateCategory(t *testing.T) {
	if config.DB == nil {
		t.Skip("Database not initialized")
	}

	app := fiber.New()
	app.Post("/categories", handlers.CreateCategory)

	tests := []struct {
		name           string
		request        types.CreateCategoryRequest
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "valid category creation",
			request: types.CreateCategoryRequest{
				Name:        "Office Supplies",
				Description: "General office supplies",
				BudgetCodes: []string{"BDG-001"},
			},
			expectedStatus: http.StatusCreated,
			expectedError:  false,
		},
		{
			name: "missing category name",
			request: types.CreateCategoryRequest{
				Name:        "",
				Description: "General office supplies",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "category name too short",
			request: types.CreateCategoryRequest{
				Name:        "ab",
				Description: "General office supplies",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", "/categories", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			resp, _ := app.Test(req)
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

// TestGetCategories tests category listing
func TestGetCategories(t *testing.T) {
	if config.DB == nil {
		t.Skip("Database not initialized")
	}

	// Cleanup
	config.DB.Exec("TRUNCATE categories CASCADE")

	// Create test data
	category := models.Category{
		ID:          uuid.New().String(),
		Name:        "Test Category",
		Description: "Test description",
		Active:      true,
	}
	config.DB.Create(&category)

	app := fiber.New()
	app.Get("/categories", handlers.GetCategories)

	req := httptest.NewRequest("GET", "/categories?page=1&limit=10", nil)
	resp, _ := app.Test(req)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var response types.ListResponse
	json.NewDecoder(resp.Body).Decode(&response)
	if !response.Success {
		t.Error("response should be successful")
	}
}

// TestUpdateCategory tests category updates
func TestUpdateCategory(t *testing.T) {
	if config.DB == nil {
		t.Skip("Database not initialized")
	}

	// Create test category
	category := models.Category{
		ID:          uuid.New().String(),
		Name:        "Original Name",
		Description: "Original description",
		Active:      true,
	}
	config.DB.Create(&category)

	updateReq := types.UpdateCategoryRequest{
		Name:        "Updated Name",
		Description: "Updated description",
	}

	body, _ := json.Marshal(updateReq)
	app := fiber.New()
	app.Put("/categories/:id", handlers.UpdateCategory)

	req := httptest.NewRequest("PUT", "/categories/"+category.ID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	// Verify database was updated
	var updated models.Category
	config.DB.Where("id = ?", category.ID).First(&updated)
	if updated.Name != "Updated Name" {
		t.Errorf("expected name 'Updated Name', got '%s'", updated.Name)
	}
}

// TestDeleteCategory tests category soft deletion
func TestDeleteCategory(t *testing.T) {
	if config.DB == nil {
		t.Skip("Database not initialized")
	}

	// Create test category
	category := models.Category{
		ID:          uuid.New().String(),
		Name:        "To Delete",
		Description: "Will be deleted",
		Active:      true,
	}
	config.DB.Create(&category)

	app := fiber.New()
	app.Delete("/categories/:id", handlers.DeleteCategory)

	req := httptest.NewRequest("DELETE", "/categories/"+category.ID, nil)
	resp, _ := app.Test(req)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	// Verify soft delete
	var deleted models.Category
	config.DB.Where("id = ?", category.ID).First(&deleted)
	if deleted.Active {
		t.Error("category should be marked as inactive")
	}
}

// TestAddBudgetCodeToCategory tests adding budget codes
func TestAddBudgetCodeToCategory(t *testing.T) {
	if config.DB == nil {
		t.Skip("Database not initialized")
	}

	// Create test category
	category := models.Category{
		ID:          uuid.New().String(),
		Name:        "Test Category",
		Description: "Test",
		Active:      true,
	}
	config.DB.Create(&category)

	addReq := struct {
		BudgetCode string `json:"budgetCode"`
	}{
		BudgetCode: "BDG-001",
	}

	body, _ := json.Marshal(addReq)
	app := fiber.New()
	app.Post("/categories/:id/budget-codes", handlers.AddBudgetCodeToCategory)

	req := httptest.NewRequest("POST", "/categories/"+category.ID+"/budget-codes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status 201, got %d", resp.StatusCode)
	}

	// Verify budget code was added
	var mapping models.CategoryBudgetCode
	err := config.DB.Where("category_id = ? AND budget_code = ?", category.ID, "BDG-001").First(&mapping).Error
	if err != nil {
		t.Errorf("budget code mapping not found: %v", err)
	}
}

// TestGetCategoryBudgetCodes tests retrieving budget codes
func TestGetCategoryBudgetCodes(t *testing.T) {
	if config.DB == nil {
		t.Skip("Database not initialized")
	}

	// Create test data
	category := models.Category{
		ID:          uuid.New().String(),
		Name:        "Test Category",
		Description: "Test",
		Active:      true,
	}
	config.DB.Create(&category)

	// Add budget codes
	for _, code := range []string{"BDG-001", "BDG-002"} {
		mapping := models.CategoryBudgetCode{
			ID:         uuid.New().String(),
			CategoryID: category.ID,
			BudgetCode: code,
			Active:     true,
		}
		config.DB.Create(&mapping)
	}

	app := fiber.New()
	app.Get("/categories/:id/budget-codes", handlers.GetCategoryBudgetCodes)

	req := httptest.NewRequest("GET", "/categories/"+category.ID+"/budget-codes", nil)
	resp, _ := app.Test(req)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

// TestRemoveBudgetCodeFromCategory tests removing budget codes
func TestRemoveBudgetCodeFromCategory(t *testing.T) {
	if config.DB == nil {
		t.Skip("Database not initialized")
	}

	// Create test data
	category := models.Category{
		ID:          uuid.New().String(),
		Name:        "Test Category",
		Description: "Test",
		Active:      true,
	}
	config.DB.Create(&category)

	mapping := models.CategoryBudgetCode{
		ID:         uuid.New().String(),
		CategoryID: category.ID,
		BudgetCode: "BDG-001",
		Active:     true,
	}
	config.DB.Create(&mapping)

	app := fiber.New()
	app.Delete("/categories/:id/budget-codes/:budgetCode", handlers.RemoveBudgetCodeFromCategory)

	req := httptest.NewRequest("DELETE", "/categories/"+category.ID+"/budget-codes/BDG-001", nil)
	resp, _ := app.Test(req)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	// Verify mapping was deleted
	var deletedMapping models.CategoryBudgetCode
	err := config.DB.Where("category_id = ? AND budget_code = ?", category.ID, "BDG-001").First(&deletedMapping).Error
	if err != gorm.ErrRecordNotFound {
		t.Error("budget code mapping should have been deleted")
	}
}
