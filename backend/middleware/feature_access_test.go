package middleware

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/gofiber/fiber/v2"
	"github.com/liyali/liyali-gateway/cache"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// makeTestTier returns a SubscriptionTier with timestamps pre-filled. glebarez
// sqlite cannot scan the string returned by the `default:CURRENT_TIMESTAMP`
// clause back into *time.Time, so we set the fields explicitly to avoid the
// RETURNING path for those columns.
func makeTestTier(features datatypes.JSON) models.SubscriptionTier {
	now := time.Now()
	return models.SubscriptionTier{
		ID:          "tier-test",
		Name:        "test",
		DisplayName: "Test",
		Features:    features,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// seedTier inserts a SubscriptionTier via raw SQL so gorm's RETURNING clause
// is not used — gorm always emits RETURNING for `default:CURRENT_TIMESTAMP`
// columns, which glebarez/sqlite can't scan back into time.Time.
func seedTier(db *gorm.DB, tier models.SubscriptionTier) error {
	if tier.CreatedAt.IsZero() {
		tier.CreatedAt = time.Now()
	}
	if tier.UpdatedAt.IsZero() {
		tier.UpdatedAt = time.Now()
	}
	if tier.Features == nil {
		tier.Features = datatypes.JSON([]byte("[]"))
	}
	return db.Exec(`INSERT INTO subscription_tiers
        (id, name, display_name, description, price_monthly, price_yearly,
         max_workspaces, max_team_members, max_documents, max_workflows,
         max_custom_roles, max_requisitions, max_budgets, max_purchase_orders,
         max_payment_vouchers, max_grns, max_departments, max_vendors,
         features, is_active, sort_order, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		tier.ID, tier.Name, tier.DisplayName, tier.Description,
		tier.PriceMonthly, tier.PriceYearly,
		tier.MaxWorkspaces, tier.MaxTeamMembers, tier.MaxDocuments, tier.MaxWorkflows,
		tier.MaxCustomRoles, tier.MaxRequisitions, tier.MaxBudgets, tier.MaxPurchaseOrders,
		tier.MaxPaymentVouchers, tier.MaxGRNs, tier.MaxDepartments, tier.MaxVendors,
		[]byte(tier.Features), tier.IsActive, tier.SortOrder,
		tier.CreatedAt, tier.UpdatedAt,
	).Error
}

// seedOverride inserts an OrganizationLimitOverride bypassing RETURNING. See
// seedTier above for the underlying glebarez sqlite limitation.
func seedOverride(db *gorm.DB, o models.OrganizationLimitOverride) error {
	if o.CreatedAt.IsZero() {
		o.CreatedAt = time.Now()
	}
	if o.UpdatedAt.IsZero() {
		o.UpdatedAt = time.Now()
	}
	var features interface{}
	if o.Features != nil {
		features = []byte(o.Features)
	}
	return db.Exec(`INSERT INTO organization_limit_overrides
        (id, organization_id, max_documents, features, reason, admin_user_id, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		o.ID, o.OrganizationID, o.MaxDocuments, features, o.Reason, o.AdminUserID,
		o.CreatedAt, o.UpdatedAt,
	).Error
}

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// glebarez/sqlite (modernc) returns CURRENT_TIMESTAMP defaults as raw
	// strings that gorm can't scan back into time.Time. The simplest cure is to
	// always set CreatedAt/UpdatedAt to time.Now() before insert, which removes
	// the columns from gorm's RETURNING clause.
	_ = db.Callback().Create().Before("gorm:create").Register("test:fill_timestamps", func(tx *gorm.DB) {
		if tx.Statement == nil || tx.Statement.Schema == nil {
			return
		}
		now := time.Now()
		if f, ok := tx.Statement.Schema.FieldsByName["CreatedAt"]; ok {
			_ = f.Set(tx.Statement.Context, tx.Statement.ReflectValue, now)
		}
		if f, ok := tx.Statement.Schema.FieldsByName["UpdatedAt"]; ok {
			_ = f.Set(tx.Statement.Context, tx.Statement.ReflectValue, now)
		}
	})

	// Auto migrate tables. The subscription_tiers and
	// organization_limit_overrides models declare
	// `type:timestamp with time zone` + `default:CURRENT_TIMESTAMP` for their
	// timestamps. Under glebarez/sqlite that combination round-trips as TEXT
	// which gorm cannot scan back into time.Time. We migrate the rest of the
	// models via gorm and create those two tables by hand with DATETIME
	// columns so the rest of the test fixtures behave normally.
	err = db.AutoMigrate(
		&models.Organization{},
		&models.SubscriptionFeature{},
		&models.OrganizationMember{},
	)
	if err == nil {
		err = db.Exec(`CREATE TABLE IF NOT EXISTS subscription_tiers (
            id TEXT PRIMARY KEY,
            name TEXT UNIQUE NOT NULL,
            display_name TEXT NOT NULL,
            description TEXT NOT NULL DEFAULT '',
            price_monthly NUMERIC NOT NULL DEFAULT 0,
            price_yearly NUMERIC NOT NULL DEFAULT 0,
            max_workspaces INTEGER NOT NULL DEFAULT 1,
            max_team_members INTEGER NOT NULL DEFAULT 1,
            max_documents INTEGER NOT NULL DEFAULT 100,
            max_workflows INTEGER NOT NULL DEFAULT 1,
            max_custom_roles INTEGER NOT NULL DEFAULT 0,
            max_requisitions INTEGER NOT NULL DEFAULT 100,
            max_budgets INTEGER NOT NULL DEFAULT 20,
            max_purchase_orders INTEGER NOT NULL DEFAULT 50,
            max_payment_vouchers INTEGER NOT NULL DEFAULT 50,
            max_grns INTEGER NOT NULL DEFAULT 50,
            max_departments INTEGER NOT NULL DEFAULT 5,
            max_vendors INTEGER NOT NULL DEFAULT 50,
            features TEXT NOT NULL DEFAULT '[]',
            is_active BOOLEAN NOT NULL DEFAULT 1,
            sort_order INTEGER NOT NULL DEFAULT 0,
            created_at DATETIME NOT NULL,
            updated_at DATETIME NOT NULL
        )`).Error
	}
	if err == nil {
		// limit_check.go counts rows in `documents` for the "document" resource.
		// We only need the table to exist for the count; no rows seeded => 0.
		err = db.Exec(`CREATE TABLE IF NOT EXISTS documents (
            id TEXT PRIMARY KEY,
            organization_id TEXT NOT NULL,
            deleted_at DATETIME
        )`).Error
	}
	if err == nil {
		err = db.Exec(`CREATE TABLE IF NOT EXISTS organization_limit_overrides (
            id TEXT PRIMARY KEY,
            organization_id TEXT UNIQUE NOT NULL,
            max_workspaces INTEGER,
            max_team_members INTEGER,
            max_documents INTEGER,
            max_workflows INTEGER,
            max_custom_roles INTEGER,
            max_requisitions INTEGER,
            max_budgets INTEGER,
            max_purchase_orders INTEGER,
            max_payment_vouchers INTEGER,
            max_grns INTEGER,
            max_departments INTEGER,
            max_vendors INTEGER,
            features TEXT,
            reason TEXT NOT NULL DEFAULT '',
            admin_user_id TEXT NOT NULL,
            expires_at DATETIME,
            created_at DATETIME NOT NULL,
            updated_at DATETIME NOT NULL
        )`).Error
	}
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
	if err := seedTier(db, tier); err != nil {
		t.Fatalf("Failed to seed tier: %v", err)
	}

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
	if err := seedTier(db, tier); err != nil {
		t.Fatalf("Failed to seed tier: %v", err)
	}

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
	if err := seedTier(db, tier); err != nil {
		t.Fatalf("Failed to seed tier: %v", err)
	}

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
	if err := seedOverride(db, override); err != nil {
		t.Fatalf("Failed to seed override: %v", err)
	}

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
	if err := seedTier(db, tier); err != nil {
		t.Fatalf("Failed to seed tier: %v", err)
	}

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
