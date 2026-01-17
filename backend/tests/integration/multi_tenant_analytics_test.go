package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/handlers"
	"github.com/liyali/liyali-gateway/middleware"
	"github.com/liyali/liyali-gateway/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMultiTenantAnalyticsSeparation tests that analytics data is properly separated between organizations
func TestMultiTenantAnalyticsSeparation(t *testing.T) {
	// Setup test database
	if err := config.InitTestDB(); err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	defer config.CleanupTestDB()

	db := config.DB

	// Note: Test data is now provided by the consolidated SQL seed migration
	// The migration 002_consolidated_seed_data.up.sql contains all necessary test data

	// Setup Fiber app with routes
	app := fiber.New()
	
	// Add middleware
	app.Use(func(c *fiber.Ctx) error {
		// Mock auth middleware - set user context
		c.Locals("userID", "user-demo-admin-001")
		return c.Next()
	})
	
	app.Use(func(c *fiber.Ctx) error {
		// Mock tenant middleware - set organization context
		orgID := c.Get("X-Organization-ID", "org-demo-001") // Default to demo org
		c.Locals("organizationID", orgID)
		return c.Next()
	})

	// Add analytics route
	app.Get("/api/v1/analytics/dashboard", handlers.GetDashboard)

	t.Run("Demo Organization Analytics", func(t *testing.T) {
		// Test Demo Organization analytics
		req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/dashboard", nil)
		req.Header.Set("X-Organization-ID", "org-demo-001")
		
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var demoResponse map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&demoResponse)
		require.NoError(t, err)

		// Verify response structure
		assert.True(t, demoResponse["success"].(bool))
		assert.Contains(t, demoResponse, "data")
		
		data := demoResponse["data"].(map[string]interface{})
		assert.Equal(t, "org-demo-001", data["organizationId"])
		
		// Verify requisition metrics exist
		assert.Contains(t, data, "requisitionMetrics")
		reqMetrics := data["requisitionMetrics"].(map[string]interface{})
		
		// Demo org should have 3 requisitions
		assert.Equal(t, float64(3), reqMetrics["totalRequisitions"])
		
		// Verify status counts
		statusCounts := reqMetrics["statusCounts"].(map[string]interface{})
		assert.Contains(t, statusCounts, "draft")
		assert.Contains(t, statusCounts, "pending") 
		assert.Contains(t, statusCounts, "approved")
	})

	t.Run("ACME Corporation Analytics", func(t *testing.T) {
		// Test ACME Corporation analytics
		req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/dashboard", nil)
		req.Header.Set("X-Organization-ID", "org-acme-001")
		
		// Mock ACME user context
		app.Use(func(c *fiber.Ctx) error {
			c.Locals("userID", "user-acme-admin-001")
			return c.Next()
		})
		
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var acmeResponse map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&acmeResponse)
		require.NoError(t, err)

		// Verify response structure
		assert.True(t, acmeResponse["success"].(bool))
		assert.Contains(t, acmeResponse, "data")
		
		data := acmeResponse["data"].(map[string]interface{})
		assert.Equal(t, "org-acme-001", data["organizationId"])
		
		// Verify requisition metrics exist
		assert.Contains(t, data, "requisitionMetrics")
		reqMetrics := data["requisitionMetrics"].(map[string]interface{})
		
		// ACME corp should have 4 requisitions
		assert.Equal(t, float64(4), reqMetrics["totalRequisitions"])
		
		// Verify status counts are different from Demo org
		statusCounts := reqMetrics["statusCounts"].(map[string]interface{})
		assert.Contains(t, statusCounts, "draft")
		assert.Contains(t, statusCounts, "pending")
		assert.Contains(t, statusCounts, "approved")
		assert.Contains(t, statusCounts, "rejected") // ACME has rejected requisitions
	})

	t.Run("Data Isolation Verification", func(t *testing.T) {
		// Verify that each organization has different data
		
		// Get Demo org requisitions
		var demoReqs []models.Requisition
		err := db.Where("organization_id = ?", "org-demo-001").Find(&demoReqs).Error
		require.NoError(t, err)
		
		// Get ACME org requisitions  
		var acmeReqs []models.Requisition
		err = db.Where("organization_id = ?", "org-acme-001").Find(&acmeReqs).Error
		require.NoError(t, err)
		
		// Verify different counts
		assert.Equal(t, 3, len(demoReqs))
		assert.Equal(t, 4, len(acmeReqs))
		
		// Verify different DocumentNumber values
		demoReqNumbers := make([]string, len(demoReqs))
		for i, req := range demoReqs {
			demoReqNumbers[i] = req.DocumentNumber
		}
		
		acmeReqNumbers := make([]string, len(acmeReqs))
		for i, req := range acmeReqs {
			acmeReqNumbers[i] = req.DocumentNumber
		}
		
		// Verify no overlap in requisition numbers
		for _, demoNum := range demoReqNumbers {
			assert.NotContains(t, acmeReqNumbers, demoNum, "Demo requisition found in ACME data")
		}
		
		for _, acmeNum := range acmeReqNumbers {
			assert.NotContains(t, demoReqNumbers, acmeNum, "ACME requisition found in Demo data")
		}
		
		// Verify different currencies
		demoCurrency := demoReqs[0].Currency
		acmeCurrency := acmeReqs[0].Currency
		assert.Equal(t, "ZMW", demoCurrency)
		assert.Equal(t, "USD", acmeCurrency)
	})

	t.Run("Budget Separation", func(t *testing.T) {
		// Verify budget data is also separated
		
		var demoBudgets []models.Budget
		err := db.Where("organization_id = ?", "org-demo-001").Find(&demoBudgets).Error
		require.NoError(t, err)
		
		var acmeBudgets []models.Budget
		err = db.Where("organization_id = ?", "org-acme-001").Find(&acmeBudgets).Error
		require.NoError(t, err)
		
		// Both should have 3 budgets
		assert.Equal(t, 3, len(demoBudgets))
		assert.Equal(t, 3, len(acmeBudgets))
		
		// Verify different budget codes
		for _, budget := range demoBudgets {
			assert.Contains(t, budget.BudgetCode, "DEMO-")
		}
		
		for _, budget := range acmeBudgets {
			assert.Contains(t, budget.BudgetCode, "ACME-")
		}
		
		// Verify different departments
		demoDepts := make([]string, len(demoBudgets))
		for i, budget := range demoBudgets {
			demoDepts[i] = budget.Department
		}
		
		acmeDepts := make([]string, len(acmeBudgets))
		for i, budget := range acmeBudgets {
			acmeDepts[i] = budget.Department
		}
		
		// Demo should have Marketing, IT, HR
		assert.Contains(t, demoDepts, "Marketing")
		assert.Contains(t, demoDepts, "IT")
		assert.Contains(t, demoDepts, "HR")
		
		// ACME should have Production, Safety, Quality
		assert.Contains(t, acmeDepts, "Production")
		assert.Contains(t, acmeDepts, "Safety")
		assert.Contains(t, acmeDepts, "Quality")
	})

	t.Run("Category Separation", func(t *testing.T) {
		// Verify categories are organization-specific
		
		var demoCategories []models.Category
		err := db.Where("organization_id = ?", "org-demo-001").Find(&demoCategories).Error
		require.NoError(t, err)
		
		var acmeCategories []models.Category
		err = db.Where("organization_id = ?", "org-acme-001").Find(&acmeCategories).Error
		require.NoError(t, err)
		
		// Both should have 3 categories
		assert.Equal(t, 3, len(demoCategories))
		assert.Equal(t, 3, len(acmeCategories))
		
		// Verify different category names
		demoCatNames := make([]string, len(demoCategories))
		for i, cat := range demoCategories {
			demoCatNames[i] = cat.Name
		}
		
		acmeCatNames := make([]string, len(acmeCategories))
		for i, cat := range acmeCategories {
			acmeCatNames[i] = cat.Name
		}
		
		// Demo categories
		assert.Contains(t, demoCatNames, "Office Supplies")
		assert.Contains(t, demoCatNames, "IT Equipment")
		assert.Contains(t, demoCatNames, "Marketing Materials")
		
		// ACME categories
		assert.Contains(t, acmeCatNames, "Manufacturing Equipment")
		assert.Contains(t, acmeCatNames, "Raw Materials")
		assert.Contains(t, acmeCatNames, "Safety Equipment")
		
		// Verify no overlap
		for _, demoCat := range demoCatNames {
			assert.NotContains(t, acmeCatNames, demoCat, "Demo category found in ACME data")
		}
	})

	// Cleanup is handled by config.CleanupTestDB()
	t.Cleanup(func() {
		// No additional cleanup needed - handled by config.CleanupTestDB()
	})
}

// TestAnalyticsWithoutOrganizationContext tests that analytics fail without proper organization context
func TestAnalyticsWithoutOrganizationContext(t *testing.T) {
	// Setup test database
	if err := config.InitTestDB(); err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	defer config.CleanupTestDB()

	// Setup Fiber app
	app := fiber.New()
	
	// Add only auth middleware, no tenant middleware
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("userID", "user-demo-admin-001")
		return c.Next()
	})

	app.Get("/api/v1/analytics/dashboard", handlers.GetDashboard)

	t.Run("Missing Organization Context", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/dashboard", nil)
		
		resp, err := app.Test(req)
		require.NoError(t, err)
		
		// Should fail without organization context
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

// TestCrossOrganizationDataLeakage tests that users cannot access other organization's data
func TestCrossOrganizationDataLeakage(t *testing.T) {
	// Setup test database
	if err := config.InitTestDB(); err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	defer config.CleanupTestDB()

	db := config.DB

	// Note: Test data is now provided by the consolidated SQL seed migration
	// The migration 002_consolidated_seed_data.up.sql contains all necessary test data

	t.Run("Super Admin Cross-Organization Access", func(t *testing.T) {
		// Test that super admin can access both organizations
		app := fiber.New()
		
		app.Use(func(c *fiber.Ctx) error {
			// Super admin user context
			c.Locals("userID", "user-admin-001")
			return c.Next()
		})
		
		app.Use(middleware.TenantMiddleware())
		app.Get("/api/v1/analytics/dashboard", handlers.GetDashboard)

		// Test access to Demo Organization
		req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/dashboard", nil)
		req.Header.Set("X-Organization-ID", "org-demo-001")
		
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode, "Super admin should access Demo org")

		// Test access to ACME Corporation
		req = httptest.NewRequest(http.MethodGet, "/api/v1/analytics/dashboard", nil)
		req.Header.Set("X-Organization-ID", "org-acme-001")
		
		resp, err = app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode, "Super admin should access ACME org")
	})

	t.Run("Demo User Cannot Access ACME Data", func(t *testing.T) {
		// Setup app with Demo user context but ACME organization header
		app := fiber.New()
		
		app.Use(func(c *fiber.Ctx) error {
			// Demo user trying to access ACME data
			c.Locals("userID", "user-demo-admin-001")
			return c.Next()
		})
		
		app.Use(middleware.TenantMiddleware())
		app.Get("/api/v1/analytics/dashboard", handlers.GetDashboard)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/dashboard", nil)
		req.Header.Set("X-Organization-ID", "org-acme-001") // Trying to access ACME data
		
		resp, err := app.Test(req)
		require.NoError(t, err)
		
		// Should be forbidden since Demo user is not a member of ACME org
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	// Cleanup is handled by config.CleanupTestDB()
	t.Cleanup(func() {
		// No additional cleanup needed - handled by config.CleanupTestDB()
	})
}