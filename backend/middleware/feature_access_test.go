package middleware

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/liyali/liyali-gateway/cache"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/models"
	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Auto migrate tables
	err = db.AutoMigrate(
		&models.Organization{},
		&models.SubscriptionTier{},
		&models.SubscriptionFeature{},
		&models.OrganizationLimitOverride{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func TestRequireFeature_Success(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	config.DB = db
	defer func() { config.DB = nil }()

	// Clear cache
	cache.GetCache().ClearAll()

	// Create test tier with features
	features, _ := json.Marshal([]string{"advanced_workflows", "custom_roles"})
	tier := models.SubscriptionTier{
		ID:          "tier-test",
		Name:        "test",
		DisplayName: "Test",
		Features:    datatypes.JSON(features),
		IsActive:    true,
	}
	db.Create(&tier)

	// Create test organization
	org := models.Organization{
		ID:   "org-test",
		Name: "Test Org",
		Tier: "test",
	}
	db.Table("organizations").Create(&org)

	// Create Fiber app
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("organizationID", "org-test")
		return c.Next()
	})
	app.Get("/test", RequireFeature("advanced_workflows"), func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Test
	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestRequireFeature_Denied(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	config.DB = db
	defer func() { config.DB = nil }()

	// Clear cache
	cache.GetCache().ClearAll()

	// Create test tier WITHOUT the feature
	features, _ := json.Marshal([]string{"basic_workflows"})
	tier := models.SubscriptionTier{
		ID:          "tier-test",
		Name:        "test",
		DisplayName: "Test",
		Features:    datatypes.JSON(features),
		IsActive:    true,
	}
	db.Create(&tier)

	// Create test organization
	org := models.Organization{
		ID:   "org-test",
		Name: "Test Org",
		Tier: "test",
	}
	db.Table("organizations").Create(&org)

	// Create Fiber app
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("organizationID", "org-test")
		return c.Next()
	})
	app.Get("/test", RequireFeature("advanced_workflows"), func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Test
	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != 403 {
		t.Errorf("Expected status 403, got %d", resp.StatusCode)
	}
}

func TestRequireFeature_WithOverride(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	config.DB = db
	defer func() { config.DB = nil }()

	// Clear cache
	cache.GetCache().ClearAll()

	// Create test tier WITHOUT the feature
	features, _ := json.Marshal([]string{"basic_workflows"})
	tier := models.SubscriptionTier{
		ID:          "tier-test",
		Name:        "test",
		DisplayName: "Test",
		Features:    datatypes.JSON(features),
		IsActive:    true,
	}
	db.Create(&tier)

	// Create test organization
	org := models.Organization{
		ID:   "org-test",
		Name: "Test Org",
		Tier: "test",
	}
	db.Table("organizations").Create(&org)

	// Create override with the feature
	overrideFeatures, _ := json.Marshal([]string{"advanced_workflows"})
	override := models.OrganizationLimitOverride{
		ID:             "override-test",
		OrganizationID: "org-test",
		Features:       datatypes.JSON(overrideFeatures),
		Reason:         "Test override",
		AdminUserID:    "admin-test",
	}
	db.Create(&override)

	// Create Fiber app
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("organizationID", "org-test")
		return c.Next()
	})
	app.Get("/test", RequireFeature("advanced_workflows"), func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Test
	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestRequireFeature_NoOrgContext(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	config.DB = db
	defer func() { config.DB = nil }()

	// Create Fiber app WITHOUT setting organizationID
	app := fiber.New()
	app.Get("/test", RequireFeature("advanced_workflows"), func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Test
	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != 400 {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

func TestCheckFeatureAccess_Cache(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	config.DB = db
	defer func() { config.DB = nil }()

	// Clear cache
	appCache := cache.GetCache()
	appCache.ClearAll()

	// Create test tier with features
	features, _ := json.Marshal([]string{"advanced_workflows"})
	tier := models.SubscriptionTier{
		ID:          "tier-test",
		Name:        "test",
		DisplayName: "Test",
		Features:    datatypes.JSON(features),
		IsActive:    true,
	}
	db.Create(&tier)

	// Create test organization
	org := models.Organization{
		ID:   "org-test",
		Name: "Test Org",
		Tier: "test",
	}
	db.Table("organizations").Create(&org)

	// First call - should query database
	hasAccess, err := checkFeatureAccess("org-test", "advanced_workflows")
	if err != nil {
		t.Fatalf("checkFeatureAccess failed: %v", err)
	}
	if !hasAccess {
		t.Error("Expected access to be granted")
	}

	// Second call - should use cache
	hasAccess, err = checkFeatureAccess("org-test", "advanced_workflows")
	if err != nil {
		t.Fatalf("checkFeatureAccess failed: %v", err)
	}
	if !hasAccess {
		t.Error("Expected access to be granted from cache")
	}

	// Verify cache was used
	cacheKey := "feature:org-test:advanced_workflows"
	if _, found := appCache.Get(cacheKey); !found {
		t.Error("Expected result to be cached")
	}
}
