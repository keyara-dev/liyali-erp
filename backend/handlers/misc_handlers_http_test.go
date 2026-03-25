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

// ============================================================================
// Shared table setup helpers
// ============================================================================

func setupSubscriptionTiersTable(t *testing.T, db *gorm.DB) {
	t.Helper()
	sqls := []string{
		`CREATE TABLE IF NOT EXISTS subscription_tiers (
			id TEXT PRIMARY KEY,
			name TEXT UNIQUE NOT NULL DEFAULT '',
			display_name TEXT NOT NULL DEFAULT '',
			description TEXT NOT NULL DEFAULT '',
			price_monthly REAL NOT NULL DEFAULT 0,
			price_yearly REAL NOT NULL DEFAULT 0,
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
			is_active NUMERIC NOT NULL DEFAULT 1,
			sort_order INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME,
			updated_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS organization_limit_overrides (
			id TEXT PRIMARY KEY,
			organization_id TEXT UNIQUE NOT NULL DEFAULT '',
			max_workspaces INTEGER,
			max_team_members INTEGER,
			max_documents INTEGER,
			max_workflows INTEGER,
			max_custom_roles INTEGER,
			max_requisitions INTEGER,
			max_budgets INTEGER,
			max_purchase_orders INTEGER,
			max_payment_vouchers INTEGER,
			"max_gr_ns" INTEGER,
			max_departments INTEGER,
			max_vendors INTEGER,
			features TEXT,
			reason TEXT NOT NULL DEFAULT '',
			admin_user_id TEXT NOT NULL DEFAULT '',
			expires_at DATETIME,
			created_at DATETIME,
			updated_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS admin_audit_logs (
			id TEXT PRIMARY KEY,
			organization_id TEXT,
			action TEXT,
			old_value TEXT,
			new_value TEXT,
			reason TEXT,
			admin_user_id TEXT,
			created_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS subscription_audit_logs (
			id TEXT PRIMARY KEY,
			organization_id TEXT,
			action TEXT,
			old_status TEXT,
			new_status TEXT,
			metadata TEXT,
			performed_by TEXT,
			performed_at DATETIME
		)`,
	}
	for _, sql := range sqls {
		if err := db.Exec(sql).Error; err != nil {
			t.Fatalf("setupSubscriptionTiersTable DDL failed: %v", err)
		}
	}
}

func setupNotificationPreferencesTable(t *testing.T, db *gorm.DB) {
	t.Helper()
	sql := `CREATE TABLE IF NOT EXISTS notification_preferences (
		id TEXT PRIMARY KEY,
		user_id TEXT UNIQUE NOT NULL DEFAULT '',
		organization_id TEXT NOT NULL DEFAULT '',
		email_enabled NUMERIC DEFAULT 0,
		push_enabled NUMERIC DEFAULT 1,
		in_app_enabled NUMERIC DEFAULT 1,
		notify_task_assigned NUMERIC DEFAULT 1,
		notify_task_reassigned NUMERIC DEFAULT 1,
		notify_task_approved NUMERIC DEFAULT 1,
		notify_task_rejected NUMERIC DEFAULT 1,
		notify_workflow_complete NUMERIC DEFAULT 1,
		notify_approval_overdue NUMERIC DEFAULT 1,
		notify_comments_added NUMERIC DEFAULT 0,
		quiet_hours_enabled NUMERIC DEFAULT 0,
		quiet_hours_start INTEGER DEFAULT 22,
		quiet_hours_end INTEGER DEFAULT 8,
		created_at DATETIME,
		updated_at DATETIME
	)`
	if err := db.Exec(sql).Error; err != nil {
		t.Fatalf("setupNotificationPreferencesTable: %v", err)
	}
}

func setupNotificationsTableWithDB(t *testing.T, db *gorm.DB) {
	t.Helper()
	sql := `CREATE TABLE IF NOT EXISTS notifications (
		id TEXT PRIMARY KEY,
		organization_id TEXT NOT NULL DEFAULT '',
		recipient_id TEXT NOT NULL DEFAULT '',
		type TEXT DEFAULT '',
		document_id TEXT DEFAULT '',
		document_type TEXT DEFAULT '',
		subject TEXT DEFAULT '',
		body TEXT DEFAULT '',
		sent NUMERIC DEFAULT 0,
		sent_at DATETIME,
		entity_id TEXT DEFAULT '',
		entity_type TEXT DEFAULT '',
		entity_number TEXT DEFAULT '',
		related_user_id TEXT DEFAULT '',
		related_user_name TEXT DEFAULT '',
		is_read NUMERIC DEFAULT 0,
		read_at DATETIME,
		action_taken NUMERIC DEFAULT 0,
		action_taken_at DATETIME,
		importance TEXT DEFAULT '',
		quick_action TEXT,
		reassignment_reason TEXT DEFAULT '',
		message TEXT DEFAULT '',
		created_at DATETIME,
		updated_at DATETIME
	)`
	if err := db.Exec(sql).Error; err != nil {
		t.Fatalf("setupNotificationsTableWithDB: %v", err)
	}
}

// seedTier inserts a starter tier into the test DB and returns it.
func seedTier(t *testing.T, db *gorm.DB) models.SubscriptionTier {
	t.Helper()
	tier := models.SubscriptionTier{
		ID:             "tier-starter",
		Name:           "starter",
		DisplayName:    "Starter",
		Description:    "Starter plan",
		PriceMonthly:   0,
		PriceYearly:    0,
		MaxWorkspaces:  1,
		MaxTeamMembers: 5,
		MaxDocuments:   100,
		MaxWorkflows:   1,
		MaxCustomRoles: 0,
		Features:       []byte(`["basic_workflows"]`),
		IsActive:       true,
		SortOrder:      1,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	if err := db.Create(&tier).Error; err != nil {
		t.Fatalf("seedTier: %v", err)
	}
	return tier
}

// seedOrg inserts an organization using the starter tier.
func seedOrg(t *testing.T, db *gorm.DB, orgID string) models.Organization {
	t.Helper()
	org := models.Organization{
		ID:        orgID,
		Name:      "Test Org",
		Slug:      "test-org-" + orgID,
		Active:    true,
		Tier:      "starter",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := db.Create(&org).Error; err != nil {
		t.Fatalf("seedOrg: %v", err)
	}
	return org
}

// ============================================================================
// Tier app builder
// ============================================================================

func newTiersApp(tenantMiddleware ...fiber.Handler) *fiber.App {
	app := fiber.New(fiber.Config{ErrorHandler: func(c *fiber.Ctx, err error) error {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}})
	g := app.Group("/tiers")
	for _, mw := range tenantMiddleware {
		g.Use(mw)
	}
	g.Get("/", GetAllTiers)
	g.Get("/:id", GetTierByID)
	g.Post("/", CreateTier)
	g.Put("/:id", UpdateTier)
	return app
}

func newOrgTiersApp(tenantMiddleware ...fiber.Handler) *fiber.App {
	app := fiber.New(fiber.Config{ErrorHandler: func(c *fiber.Ctx, err error) error {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}})
	g := app.Group("/orgs")
	for _, mw := range tenantMiddleware {
		g.Use(mw)
	}
	g.Post("/:id/tier", ChangeOrganizationTier)
	g.Post("/:id/limits", OverrideOrganizationLimits)
	g.Get("/:id/subscription", GetOrganizationSubscription)
	g.Get("/:id/features", GetOrganizationFeatures)
	g.Get("/:id/limits", GetOrganizationLimits)
	g.Get("/:id/usage", GetOrganizationUsage)
	return app
}

// ============================================================================
// GetAllTiers
// ============================================================================

func TestGetAllTiers_Empty(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupSubscriptionTiersTable(t, db)

	app := newTiersApp()
	resp := testRequest(app, http.MethodGet, "/tiers/", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])
}

func TestGetAllTiers_WithData(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupSubscriptionTiersTable(t, db)
	seedTier(t, db)

	app := newTiersApp()
	resp := testRequest(app, http.MethodGet, "/tiers/", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])
	data, ok := body["data"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, data, 1)
}

// ============================================================================
// GetTierByID
// ============================================================================

func TestGetTierByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupSubscriptionTiersTable(t, db)

	app := newTiersApp()
	resp := testRequest(app, http.MethodGet, "/tiers/nonexistent", nil)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestGetTierByID_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupSubscriptionTiersTable(t, db)
	tier := seedTier(t, db)

	app := newTiersApp()
	resp := testRequest(app, http.MethodGet, "/tiers/"+tier.ID, nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])
	data, ok := body["data"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "starter", data["name"])
}

// ============================================================================
// CreateTier
// ============================================================================

func TestCreateTier_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupSubscriptionTiersTable(t, db)

	app := newTiersApp()
	payload := map[string]interface{}{
		"name":           "pro",
		"displayName":    "Pro Plan",
		"description":    "Professional plan with more features",
		"priceMonthly":   29.99,
		"priceYearly":    299.0,
		"maxWorkspaces":  5,
		"maxTeamMembers": 25,
		"maxDocuments":   1000,
		"maxWorkflows":   10,
		"maxCustomRoles": 5,
		"features":       []string{"advanced_workflows", "priority_support"},
		"isActive":       true,
		"sortOrder":      2,
	}
	resp := testRequest(app, http.MethodPost, "/tiers/", payload)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])
	data, ok := body["data"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "pro", data["name"])
}

// ============================================================================
// UpdateTier
// ============================================================================

func TestUpdateTier_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupSubscriptionTiersTable(t, db)

	app := newTiersApp()
	resp := testRequest(app, http.MethodPut, "/tiers/nonexistent", map[string]interface{}{
		"displayName": "Updated",
	})
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestUpdateTier_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupSubscriptionTiersTable(t, db)
	tier := seedTier(t, db)

	newName := "Starter Updated"
	app := newTiersApp()
	resp := testRequest(app, http.MethodPut, "/tiers/"+tier.ID, map[string]interface{}{
		"displayName": newName,
	})
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])
}

// ============================================================================
// ChangeOrganizationTier
// ============================================================================

func TestChangeTier_TierNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupSubscriptionTiersTable(t, db)
	seedOrg(t, db, testOrgID)

	app := newOrgTiersApp()
	resp := testRequest(app, http.MethodPost, "/orgs/"+testOrgID+"/tier", map[string]interface{}{
		"newTier": "pro",
		"reason":  "Upgrading for testing purposes now",
	})
	// tier "pro" not in DB → 400
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestChangeTier_OrgNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupSubscriptionTiersTable(t, db)
	// seed the pro tier so tier validation passes
	proTier := models.SubscriptionTier{
		ID:             "tier-pro",
		Name:           "pro",
		DisplayName:    "Pro",
		Description:    "Pro plan",
		Features:       []byte(`[]`),
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	assert.NoError(t, db.Create(&proTier).Error)

	app := newOrgTiersApp()
	resp := testRequest(app, http.MethodPost, "/orgs/nonexistent-org/tier", map[string]interface{}{
		"newTier": "pro",
		"reason":  "Upgrading for testing purposes now",
	})
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestChangeTier_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupSubscriptionTiersTable(t, db)
	seedTier(t, db) // starter
	// seed pro
	proTier := models.SubscriptionTier{
		ID:          "tier-pro",
		Name:        "pro",
		DisplayName: "Pro",
		Description: "Pro plan",
		Features:    []byte(`[]`),
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	assert.NoError(t, db.Create(&proTier).Error)
	seedOrg(t, db, testOrgID)

	app := newOrgTiersApp(withTenantCtx(testOrgID, testUserID, testUserRole))
	resp := testRequest(app, http.MethodPost, "/orgs/"+testOrgID+"/tier", map[string]interface{}{
		"newTier": "pro",
		"reason":  "Upgrading for testing purposes now",
	})
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])
}

// ============================================================================
// OverrideOrganizationLimits
// ============================================================================

func TestOverrideLimits_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupSubscriptionTiersTable(t, db)
	seedOrg(t, db, testOrgID)

	maxDocs := 500
	app := newOrgTiersApp(withTenantCtx(testOrgID, testUserID, testUserRole))
	resp := testRequest(app, http.MethodPost, "/orgs/"+testOrgID+"/limits", map[string]interface{}{
		"maxDocuments": maxDocs,
		"reason":       "Special override for testing this organization",
	})
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])
}

func TestOverrideLimits_InvalidExpiresAt(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupSubscriptionTiersTable(t, db)
	seedOrg(t, db, testOrgID)

	app := newOrgTiersApp(withTenantCtx(testOrgID, testUserID, testUserRole))
	resp := testRequest(app, http.MethodPost, "/orgs/"+testOrgID+"/limits", map[string]interface{}{
		"reason":    "Override for testing purposes here",
		"expiresAt": "not-a-date",
	})
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// ============================================================================
// GetOrganizationSubscription
// ============================================================================

func TestGetOrgSubscription_OrgNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupSubscriptionTiersTable(t, db)

	app := newOrgTiersApp()
	resp := testRequest(app, http.MethodGet, "/orgs/nonexistent/subscription", nil)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestGetOrgSubscription_TierNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupSubscriptionTiersTable(t, db)
	// org exists but tier "starter" not in subscription_tiers table
	seedOrg(t, db, testOrgID)

	app := newOrgTiersApp()
	resp := testRequest(app, http.MethodGet, "/orgs/"+testOrgID+"/subscription", nil)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestGetOrgSubscription_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupSubscriptionTiersTable(t, db)
	seedTier(t, db)
	seedOrg(t, db, testOrgID)

	app := newOrgTiersApp()
	resp := testRequest(app, http.MethodGet, "/orgs/"+testOrgID+"/subscription", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])
}

// ============================================================================
// GetOrganizationFeatures
// ============================================================================

func TestGetOrgFeatures_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupSubscriptionTiersTable(t, db)
	seedTier(t, db)
	seedOrg(t, db, testOrgID)

	app := newOrgTiersApp()
	resp := testRequest(app, http.MethodGet, "/orgs/"+testOrgID+"/features", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])
}

func TestGetOrgFeatures_OrgNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupSubscriptionTiersTable(t, db)

	app := newOrgTiersApp()
	resp := testRequest(app, http.MethodGet, "/orgs/nonexistent/features", nil)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

// ============================================================================
// GetOrganizationLimits
// ============================================================================

func TestGetOrgLimits_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupSubscriptionTiersTable(t, db)
	seedTier(t, db)
	seedOrg(t, db, testOrgID)

	app := newOrgTiersApp()
	resp := testRequest(app, http.MethodGet, "/orgs/"+testOrgID+"/limits", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])
}

func TestGetOrgLimits_OrgNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupSubscriptionTiersTable(t, db)

	app := newOrgTiersApp()
	resp := testRequest(app, http.MethodGet, "/orgs/nonexistent/limits", nil)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

// ============================================================================
// GetOrganizationUsage
// ============================================================================

func TestGetOrgUsage_ViaURLParam(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupSubscriptionTiersTable(t, db)
	seedTier(t, db)
	seedOrg(t, db, testOrgID)

	app := newOrgTiersApp()
	resp := testRequest(app, http.MethodGet, "/orgs/"+testOrgID+"/usage", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])
}

func TestGetOrgUsage_ViaLocals(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupSubscriptionTiersTable(t, db)
	seedTier(t, db)
	seedOrg(t, db, testOrgID)

	// Route without :id param — usage falls back to c.Locals("organizationID")
	app := fiber.New()
	app.Get("/usage", withTenantCtx(testOrgID, testUserID, testUserRole), GetOrganizationUsage)

	resp := testRequest(app, http.MethodGet, "/usage", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetOrgUsage_NoOrgID(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupSubscriptionTiersTable(t, db)

	// No URL param, no local → 400
	app := fiber.New()
	app.Get("/usage", GetOrganizationUsage)

	resp := testRequest(app, http.MethodGet, "/usage", nil)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// ============================================================================
// GenerateDocument
// ============================================================================

func newDocGenApp(h *DocumentGenerationHandler, tenantMiddleware ...fiber.Handler) *fiber.App {
	app := fiber.New(fiber.Config{ErrorHandler: func(c *fiber.Ctx, err error) error {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}})
	g := app.Group("/documents")
	for _, mw := range tenantMiddleware {
		g.Use(mw)
	}
	g.Post("/generate", h.GenerateDocument)
	return app
}

func TestGenerateDocument_NoAuth(t *testing.T) {
	h := NewDocumentGenerationHandler(nil)
	app := newDocGenApp(h) // no tenant middleware
	resp := testRequest(app, http.MethodPost, "/documents/generate", map[string]interface{}{
		"id":      "some-id",
		"docType": "requisition",
	})
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestGenerateDocument_MissingID(t *testing.T) {
	h := NewDocumentGenerationHandler(nil)
	app := newDocGenApp(h, withTenantCtx(testOrgID, testUserID, testUserRole))
	resp := testRequest(app, http.MethodPost, "/documents/generate", map[string]interface{}{
		"docType": "requisition",
	})
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGenerateDocument_MissingDocType(t *testing.T) {
	h := NewDocumentGenerationHandler(nil)
	app := newDocGenApp(h, withTenantCtx(testOrgID, testUserID, testUserRole))
	resp := testRequest(app, http.MethodPost, "/documents/generate", map[string]interface{}{
		"id": "some-id",
	})
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGenerateDocument_BothMissing(t *testing.T) {
	h := NewDocumentGenerationHandler(nil)
	app := newDocGenApp(h, withTenantCtx(testOrgID, testUserID, testUserRole))
	resp := testRequest(app, http.MethodPost, "/documents/generate", map[string]interface{}{})
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// ============================================================================
// notifications.go — GetNotifications / GetNotificationStats /
//                    MarkNotificationAsRead / MarkAllNotificationsAsRead
// ============================================================================

func newNotificationsApp(tenantMiddleware ...fiber.Handler) *fiber.App {
	app := fiber.New(fiber.Config{ErrorHandler: func(c *fiber.Ctx, err error) error {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}})
	g := app.Group("/notifications")
	for _, mw := range tenantMiddleware {
		g.Use(mw)
	}
	g.Get("/", GetNotifications)
	g.Get("/stats", GetNotificationStats)
	g.Put("/read-all", MarkAllNotificationsAsRead)
	g.Get("/:id", GetNotification)
	g.Put("/:id/read", MarkNotificationAsRead)
	return app
}

func TestNotificationGetNotifications_NoAuth(t *testing.T) {
	app := newNotificationsApp()
	resp := testRequest(app, http.MethodGet, "/notifications/", nil)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestNotificationGetNotifications_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupNotificationsTableWithDB(t, db)

	app := newNotificationsApp(withTenantCtx(testOrgID, testUserID, testUserRole))
	resp := testRequest(app, http.MethodGet, "/notifications/", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])
}

func TestNotificationGetNotificationStats_NoAuth(t *testing.T) {
	app := newNotificationsApp()
	resp := testRequest(app, http.MethodGet, "/notifications/stats", nil)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestNotificationGetNotificationStats_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupNotificationsTableWithDB(t, db)

	app := newNotificationsApp(withTenantCtx(testOrgID, testUserID, testUserRole))
	resp := testRequest(app, http.MethodGet, "/notifications/stats", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestNotificationMarkAllRead_NoAuth(t *testing.T) {
	app := newNotificationsApp()
	resp := testRequest(app, http.MethodPut, "/notifications/read-all", nil)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestNotificationMarkAllRead_NoneUnread(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupNotificationsTableWithDB(t, db)

	app := newNotificationsApp(withTenantCtx(testOrgID, testUserID, testUserRole))
	resp := testRequest(app, http.MethodPut, "/notifications/read-all", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])
}

func TestNotificationGetNotification_NoAuth(t *testing.T) {
	app := newNotificationsApp()
	resp := testRequest(app, http.MethodGet, "/notifications/some-id", nil)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestNotificationMarkAsRead_NoAuth(t *testing.T) {
	app := newNotificationsApp()
	resp := testRequest(app, http.MethodPut, "/notifications/some-id/read", nil)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestNotificationMarkAsRead_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupNotificationsTableWithDB(t, db)

	app := newNotificationsApp(withTenantCtx(testOrgID, testUserID, testUserRole))
	resp := testRequest(app, http.MethodPut, "/notifications/nonexistent/read", nil)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// ============================================================================
// notification_handler.go — GetNotificationPreferences / UpdateNotificationPreferences
// ============================================================================

func newNotifPrefsApp(h *NotificationHandler, tenantMiddleware ...fiber.Handler) *fiber.App {
	app := fiber.New(fiber.Config{ErrorHandler: func(c *fiber.Ctx, err error) error {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}})
	g := app.Group("/prefs")
	for _, mw := range tenantMiddleware {
		g.Use(mw)
	}
	g.Get("/", h.GetNotificationPreferences)
	g.Put("/", h.UpdateNotificationPreferences)
	return app
}

func TestNotificationGetPreferences_NoOrgContext(t *testing.T) {
	h := NewNotificationHandler()
	app := newNotifPrefsApp(h) // no middleware
	resp := testRequest(app, http.MethodGet, "/prefs/", nil)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestNotificationGetPreferences_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupNotificationPreferencesTable(t, db)

	h := NewNotificationHandler()
	app := newNotifPrefsApp(h, withTenantCtx(testOrgID, testUserID, testUserRole))
	// First call creates default preferences
	resp := testRequest(app, http.MethodGet, "/prefs/", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])
}

func TestNotificationUpdatePreferences_NoOrgContext(t *testing.T) {
	h := NewNotificationHandler()
	app := newNotifPrefsApp(h)
	resp := testRequest(app, http.MethodPut, "/prefs/", map[string]interface{}{
		"emailEnabled": true,
		"pushEnabled":  true,
		"inAppEnabled": true,
	})
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestNotificationUpdatePreferences_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupNotificationPreferencesTable(t, db)

	h := NewNotificationHandler()
	app := newNotifPrefsApp(h, withTenantCtx(testOrgID, testUserID, testUserRole))
	resp := testRequest(app, http.MethodPut, "/prefs/", map[string]interface{}{
		"emailEnabled":           true,
		"pushEnabled":            true,
		"inAppEnabled":           true,
		"notifyTaskAssigned":     true,
		"notifyTaskReassigned":   false,
		"notifyTaskApproved":     true,
		"notifyTaskRejected":     true,
		"notifyWorkflowComplete": true,
		"notifyApprovalOverdue":  false,
		"notifyCommentsAdded":    false,
		"quietHoursEnabled":      false,
		"quietHoursStart":        22,
		"quietHoursEnd":          8,
	})
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])
}

func TestNotificationUpdatePreferences_InvalidQuietHours(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupNotificationPreferencesTable(t, db)

	h := NewNotificationHandler()
	app := newNotifPrefsApp(h, withTenantCtx(testOrgID, testUserID, testUserRole))
	resp := testRequest(app, http.MethodPut, "/prefs/", map[string]interface{}{
		"emailEnabled":     false,
		"pushEnabled":      true,
		"inAppEnabled":     true,
		"quietHoursStart": 25, // out of range
		"quietHoursEnd":   8,
	})
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// ============================================================================
// vendor.go — DeleteVendor
// ============================================================================

func newDeleteVendorApp(tenantMiddleware ...fiber.Handler) *fiber.App {
	app := fiber.New(fiber.Config{ErrorHandler: func(c *fiber.Ctx, err error) error {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}})
	g := app.Group("/vendors")
	for _, mw := range tenantMiddleware {
		g.Use(mw)
	}
	g.Delete("/:id", DeleteVendor)
	return app
}

func TestDeleteVendor_NoAuth(t *testing.T) {
	app := newDeleteVendorApp()
	resp := testRequest(app, http.MethodDelete, "/vendors/some-id", nil)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestDeleteVendor_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	app := newDeleteVendorApp(withTenantCtx(testOrgID, testUserID, testUserRole))
	resp := testRequest(app, http.MethodDelete, "/vendors/nonexistent-id", nil)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestDeleteVendor_CrossOrgBlocked(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Vendor belongs to a different org
	vendor := models.Vendor{
		ID:             uuid.New().String(),
		OrganizationID: "other-org-999",
		VendorCode:     "VND-DEL-CROSS",
		Name:           "Cross Org Vendor",
		Email:          "cross@example.com",
		Phone:          "+260971000099",
		Country:        "Zambia",
		City:           "Lusaka",
		BankAccount:    "ACC-CROSS",
		TaxID:          "TAX-CROSS",
		Active:         true,
		CreatedBy:      "other-user",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	assert.NoError(t, db.Create(&vendor).Error)

	app := newDeleteVendorApp(withTenantCtx(testOrgID, testUserID, testUserRole))
	resp := testRequest(app, http.MethodDelete, "/vendors/"+vendor.ID, nil)
	// Different org → not found (security: no cross-org leak)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestDeleteVendor_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	vendor := models.Vendor{
		ID:             uuid.New().String(),
		OrganizationID: testOrgID,
		VendorCode:     "VND-DEL-001",
		Name:           "To Be Deleted",
		Email:          "delete@example.com",
		Phone:          "+260971000088",
		Country:        "Zambia",
		City:           "Lusaka",
		BankAccount:    "ACC-DEL",
		TaxID:          "TAX-DEL",
		Active:         true,
		CreatedBy:      testUserID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	assert.NoError(t, db.Create(&vendor).Error)

	app := newDeleteVendorApp(withTenantCtx(testOrgID, testUserID, testUserRole))
	resp := testRequest(app, http.MethodDelete, "/vendors/"+vendor.ID, nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])

	// Verify soft-deleted (active = false)
	var updated models.Vendor
	assert.NoError(t, db.First(&updated, "id = ?", vendor.ID).Error)
	assert.False(t, updated.Active)
}
