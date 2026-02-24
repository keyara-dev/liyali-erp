package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/liyali/liyali-gateway/cache"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/models"
)

func TestCheckLimit_UnderLimit(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	config.DB = db
	defer func() { config.DB = nil }()

	// Clear cache
	cache.GetCache().ClearAll()

	// Create test tier with limits
	tier := models.SubscriptionTier{
		ID:             "tier-test",
		Name:           "test",
		DisplayName:    "Test",
		MaxDocuments:   100,
		IsActive:       true,
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
	app.Post("/documents", CheckLimit("document"), func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Test
	req := httptest.NewRequest("POST", "/documents", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestCheckLimit_AtLimit(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	config.DB = db
	defer func() { config.DB = nil }()

	// Clear cache
	cache.GetCache().ClearAll()

	// Create test tier with low limit
	tier := models.SubscriptionTier{
		ID:             "tier-test",
		Name:           "test",
		DisplayName:    "Test",
		MaxTeamMembers: 2,
		IsActive:       true,
	}
	db.Create(&tier)

	// Create test organization
	org := models.Organization{
		ID:   "org-test",
		Name: "Test Org",
		Tier: "test",
	}
	db.Table("organizations").Create(&org)

	// Create 2 team members (at limit)
	db.Table("organization_members").Create(&models.OrganizationMember{
		ID:             "member-1",
		OrganizationID: "org-test",
		UserID:         "user-1",
		Active:         true,
	})
	db.Table("organization_members").Create(&models.OrganizationMember{
		ID:             "member-2",
		OrganizationID: "org-test",
		UserID:         "user-2",
		Active:         true,
	})

	// Create Fiber app
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("organizationID", "org-test")
		return c.Next()
	})
	app.Post("/members", CheckLimit("team_member"), func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Test
	req := httptest.NewRequest("POST", "/members", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != 400 {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

func TestCheckLimit_Unlimited(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	config.DB = db
	defer func() { config.DB = nil }()

	// Clear cache
	cache.GetCache().ClearAll()

	// Create test tier with unlimited documents
	tier := models.SubscriptionTier{
		ID:             "tier-test",
		Name:           "test",
		DisplayName:    "Test",
		MaxDocuments:   -1, // Unlimited
		IsActive:       true,
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
	app.Post("/documents", CheckLimit("document"), func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Test
	req := httptest.NewRequest("POST", "/documents", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200 (unlimited), got %d", resp.StatusCode)
	}
}

func TestCheckLimit_WithOverride(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	config.DB = db
	defer func() { config.DB = nil }()

	// Clear cache
	cache.GetCache().ClearAll()

	// Create test tier with low limit
	tier := models.SubscriptionTier{
		ID:             "tier-test",
		Name:           "test",
		DisplayName:    "Test",
		MaxDocuments:   10,
		IsActive:       true,
	}
	db.Create(&tier)

	// Create test organization
	org := models.Organization{
		ID:   "org-test",
		Name: "Test Org",
		Tier: "test",
	}
	db.Table("organizations").Create(&org)

	// Create override with higher limit
	higherLimit := 100
	override := models.OrganizationLimitOverride{
		ID:             "override-test",
		OrganizationID: "org-test",
		MaxDocuments:   &higherLimit,
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
	app.Post("/documents", CheckLimit("document"), func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Test
	req := httptest.NewRequest("POST", "/documents", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200 (override applied), got %d", resp.StatusCode)
	}
}

func TestGetEffectiveLimits_Cache(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	config.DB = db
	defer func() { config.DB = nil }()

	// Clear cache
	appCache := cache.GetCache()
	appCache.ClearAll()

	// Create test tier
	tier := models.SubscriptionTier{
		ID:             "tier-test",
		Name:           "test",
		DisplayName:    "Test",
		MaxWorkspaces:  5,
		MaxTeamMembers: 50,
		MaxDocuments:   500,
		MaxWorkflows:   20,
		MaxCustomRoles: 10,
		IsActive:       true,
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
	limits, err := getEffectiveLimits("org-test")
	if err != nil {
		t.Fatalf("getEffectiveLimits failed: %v", err)
	}
	if limits.MaxDocuments != 500 {
		t.Errorf("Expected MaxDocuments 500, got %d", limits.MaxDocuments)
	}

	// Second call - should use cache
	limits, err = getEffectiveLimits("org-test")
	if err != nil {
		t.Fatalf("getEffectiveLimits failed: %v", err)
	}
	if limits.MaxDocuments != 500 {
		t.Errorf("Expected MaxDocuments 500 from cache, got %d", limits.MaxDocuments)
	}

	// Verify cache was used
	cacheKey := "limits:org-test"
	if _, found := appCache.Get(cacheKey); !found {
		t.Error("Expected result to be cached")
	}
}

func TestGetCurrentUsage_TeamMembers(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	config.DB = db
	defer func() { config.DB = nil }()

	// Create test organization
	org := models.Organization{
		ID:   "org-test",
		Name: "Test Org",
	}
	db.Table("organizations").Create(&org)

	// Create 3 active team members
	for i := 1; i <= 3; i++ {
		db.Table("organization_members").Create(&models.OrganizationMember{
			ID:             fiber.NewError(i).Error(),
			OrganizationID: "org-test",
			UserID:         fiber.NewError(i).Error(),
			Active:         true,
		})
	}

	// Create 1 inactive member (should not be counted)
	db.Table("organization_members").Create(&models.OrganizationMember{
		ID:             "member-inactive",
		OrganizationID: "org-test",
		UserID:         "user-inactive",
		Active:         false,
	})

	// Test
	usage, err := getCurrentUsage("org-test", "team_member")
	if err != nil {
		t.Fatalf("getCurrentUsage failed: %v", err)
	}

	if usage != 3 {
		t.Errorf("Expected usage 3, got %d", usage)
	}
}
