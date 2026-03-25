package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/gofiber/fiber/v2"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ---------------------------------------------------------------------------
// DB setup
// ---------------------------------------------------------------------------

// setupAdminAnalyticsDB creates an isolated in-memory SQLite DB for admin
// analytics tests.  It auto-migrates the GORM models needed by the handlers
// and creates additional tables via raw DDL for tables the handlers reference
// with raw SQL but that don't have a corresponding model in the project.
func setupAdminAnalyticsDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}

	if err := db.AutoMigrate(
		&models.Organization{},
		&models.User{},
		&models.Requisition{},
		&models.PurchaseOrder{},
		&models.PaymentVoucher{},
		&models.GoodsReceivedNote{},
	); err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}

	// Extra columns that the handlers' raw SQL references but the GORM model
	// may not expose.
	extraAlters := []string{
		`ALTER TABLE organizations ADD COLUMN subscription_status TEXT DEFAULT 'trial'`,
		`ALTER TABLE organizations ADD COLUMN subscription_tier TEXT DEFAULT 'starter'`,
		`ALTER TABLE organizations ADD COLUMN trial_end_date DATETIME`,
		`ALTER TABLE organizations ADD COLUMN trial_ends_at DATETIME`,
		`ALTER TABLE organizations ADD COLUMN deleted_at DATETIME`,
		`ALTER TABLE users ADD COLUMN last_login DATETIME`,
		`ALTER TABLE users ADD COLUMN role TEXT DEFAULT 'requester'`,
		`ALTER TABLE users ADD COLUMN status TEXT DEFAULT 'active'`,
		`ALTER TABLE users ADD COLUMN organization_id TEXT DEFAULT ''`,
		`ALTER TABLE users ADD COLUMN deleted_at DATETIME`,
	}
	for _, stmt := range extraAlters {
		_ = db.Exec(stmt).Error // ignore duplicate column errors
	}

	// Tables that are only accessed via raw SQL (no GORM model).
	rawTables := []string{
		`CREATE TABLE IF NOT EXISTS organization_subscriptions (
			id TEXT PRIMARY KEY,
			organization_id TEXT,
			status TEXT,
			tier TEXT,
			created_at DATETIME,
			updated_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS system_metrics (
			id TEXT PRIMARY KEY,
			metric_type TEXT,
			value REAL,
			recorded_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS system_services (
			id TEXT PRIMARY KEY,
			service_name TEXT,
			status TEXT,
			response_time_ms REAL,
			last_check_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS system_alerts (
			id TEXT PRIMARY KEY,
			severity TEXT,
			status TEXT,
			message TEXT,
			created_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS system_logs (
			id TEXT PRIMARY KEY,
			level TEXT,
			service TEXT,
			message TEXT,
			created_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS documents (
			id TEXT PRIMARY KEY,
			document_type TEXT,
			workflow_status TEXT,
			created_by TEXT,
			created_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS payments (
			id TEXT PRIMARY KEY,
			organization_id TEXT,
			subscription_tier TEXT,
			amount REAL,
			payment_status TEXT,
			payment_type TEXT,
			paid_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS subscription_tiers (
			id TEXT PRIMARY KEY,
			name TEXT,
			is_active INTEGER DEFAULT 1
		)`,
		`CREATE TABLE IF NOT EXISTS subscription_events (
			id TEXT PRIMARY KEY,
			organization_id TEXT,
			event_type TEXT,
			created_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS api_request_logs (
			id TEXT PRIMARY KEY,
			user_id TEXT,
			status_code INTEGER,
			response_time_ms REAL,
			created_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS system_settings (
			id TEXT PRIMARY KEY,
			key TEXT UNIQUE,
			value TEXT,
			created_at DATETIME,
			updated_at DATETIME
		)`,
	}
	for _, stmt := range rawTables {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("failed to create table: %v\nSQL: %s", err, stmt)
		}
	}

	// Use a single connection so seeded data is visible to handler queries.
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get sql.DB: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)

	config.DB = db
	return db
}

func teardownAdminAnalyticsDB(t *testing.T, db *gorm.DB) {
	t.Helper()
	sqlDB, _ := db.DB()
	_ = sqlDB.Close()
	config.DB = nil
}

// ---------------------------------------------------------------------------
// App factory
// ---------------------------------------------------------------------------

func newAdminAnalyticsApp() *fiber.App {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/admin/dashboard", GetAdminDashboard)
	app.Get("/admin/health", GetSystemHealth)
	app.Get("/admin/analytics", GetAdminAnalytics)
	app.Get("/admin/analytics/users", GetAdminUserAnalytics)
	app.Get("/admin/analytics/organizations", GetAdminOrganizationAnalytics)
	app.Get("/admin/analytics/revenue", GetAdminRevenueAnalytics)
	app.Get("/admin/analytics/usage", GetAdminUsageAnalytics)
	app.Get("/admin/analytics/subscriptions", GetSubscriptionStatistics)
	app.Get("/admin/alerts", GetSystemAlerts)
	app.Get("/admin/logs", GetSystemLogs)
	app.Get("/admin/metrics", GetSystemMetrics)
	app.Post("/admin/analytics/export", ExportAdminAnalytics)
	app.Post("/admin/analytics/custom", RunCustomAdminAnalytics)
	app.Get("/admin/analytics/config", GetAdminAnalyticsDashboardConfig)
	app.Put("/admin/analytics/config", UpdateAdminAnalyticsDashboardConfig)
	return app
}

// ---------------------------------------------------------------------------
// GetAdminDashboard
// ---------------------------------------------------------------------------

func TestGetAdminDashboard_EmptyDB(t *testing.T) {
	db := setupAdminAnalyticsDB(t)
	defer teardownAdminAnalyticsDB(t, db)

	app := newAdminAnalyticsApp()
	req := httptest.NewRequest(http.MethodGet, "/admin/dashboard", nil)
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.True(t, body["success"].(bool))
}

func TestGetAdminDashboard_WithData(t *testing.T) {
	db := setupAdminAnalyticsDB(t)
	defer teardownAdminAnalyticsDB(t, db)

	// Seed some organizations and users.
	db.Exec(`INSERT INTO organizations (id, name, slug, created_at, updated_at) VALUES ('org-1', 'Org One', 'org-one', datetime('now'), datetime('now'))`)
	db.Exec(`INSERT INTO users (id, email, created_at, updated_at) VALUES ('usr-1', 'a@b.com', datetime('now'), datetime('now'))`)

	app := newAdminAnalyticsApp()
	req := httptest.NewRequest(http.MethodGet, "/admin/dashboard", nil)
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.True(t, body["success"].(bool))
	data := body["data"].(map[string]interface{})
	assert.EqualValues(t, 1, data["total_organizations"])
	assert.EqualValues(t, 1, data["total_users"])
}

// ---------------------------------------------------------------------------
// GetSystemHealth
// ---------------------------------------------------------------------------

func TestGetSystemHealth_EmptyDB(t *testing.T) {
	db := setupAdminAnalyticsDB(t)
	defer teardownAdminAnalyticsDB(t, db)

	app := newAdminAnalyticsApp()
	req := httptest.NewRequest(http.MethodGet, "/admin/health", nil)
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.True(t, body["success"].(bool))
}

func TestGetSystemHealth_WithMetrics(t *testing.T) {
	db := setupAdminAnalyticsDB(t)
	defer teardownAdminAnalyticsDB(t, db)

	db.Exec(`INSERT INTO system_metrics (id, metric_type, value, recorded_at) VALUES ('m-1', 'cpu', 45.5, datetime('now'))`)
	db.Exec(`INSERT INTO system_metrics (id, metric_type, value, recorded_at) VALUES ('m-2', 'memory', 72.0, datetime('now'))`)
	db.Exec(`INSERT INTO system_metrics (id, metric_type, value, recorded_at) VALUES ('m-3', 'disk', 60.0, datetime('now'))`)
	db.Exec(`INSERT INTO system_services (id, service_name, status, response_time_ms, last_check_at) VALUES ('svc-1', 'api', 'healthy', 12.5, datetime('now'))`)

	app := newAdminAnalyticsApp()
	req := httptest.NewRequest(http.MethodGet, "/admin/health", nil)
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.True(t, body["success"].(bool))
	data := body["data"].(map[string]interface{})
	assert.Equal(t, "healthy", data["status"])
}

// ---------------------------------------------------------------------------
// GetAdminAnalytics
// ---------------------------------------------------------------------------

func TestGetAdminAnalytics_EmptyDB(t *testing.T) {
	db := setupAdminAnalyticsDB(t)
	defer teardownAdminAnalyticsDB(t, db)

	app := newAdminAnalyticsApp()
	req := httptest.NewRequest(http.MethodGet, "/admin/analytics", nil)
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.True(t, body["success"].(bool))
}

func TestGetAdminAnalytics_WithData(t *testing.T) {
	db := setupAdminAnalyticsDB(t)
	defer teardownAdminAnalyticsDB(t, db)

	db.Exec(`INSERT INTO documents (id, document_type, created_at) VALUES ('doc-1', 'requisition', datetime('now'))`)
	db.Exec(`INSERT INTO documents (id, document_type, workflow_status, created_at) VALUES ('doc-2', 'purchase_order', 'approved', datetime('now'))`)
	db.Exec(`INSERT INTO users (id, email, created_at, updated_at) VALUES ('usr-aa', 'x@y.com', datetime('now'), datetime('now'))`)
	db.Exec(`INSERT INTO organizations (id, name, slug, created_at, updated_at) VALUES ('org-aa', 'Org AA', 'org-aa', datetime('now'), datetime('now'))`)

	app := newAdminAnalyticsApp()
	req := httptest.NewRequest(http.MethodGet, "/admin/analytics", nil)
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.True(t, body["success"].(bool))
	data := body["data"].(map[string]interface{})
	assert.NotNil(t, data["monthly_growth"])
}

// ---------------------------------------------------------------------------
// GetAdminUserAnalytics
// ---------------------------------------------------------------------------

func TestGetAdminUserAnalytics_EmptyDB(t *testing.T) {
	db := setupAdminAnalyticsDB(t)
	defer teardownAdminAnalyticsDB(t, db)

	app := newAdminAnalyticsApp()
	req := httptest.NewRequest(http.MethodGet, "/admin/analytics/users", nil)
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.True(t, body["success"].(bool))
}

func TestGetAdminUserAnalytics_WithData(t *testing.T) {
	db := setupAdminAnalyticsDB(t)
	defer teardownAdminAnalyticsDB(t, db)

	db.Exec(`INSERT INTO users (id, email, role, status, organization_id, last_login, created_at, updated_at) VALUES ('u1', 'a@a.com', 'admin', 'active', 'org-1', datetime('now'), datetime('now'), datetime('now'))`)
	db.Exec(`INSERT INTO users (id, email, role, status, organization_id, last_login, created_at, updated_at) VALUES ('u2', 'b@b.com', 'requester', 'active', 'org-1', datetime('now','-10 days'), datetime('now','-20 days'), datetime('now'))`)
	db.Exec(`INSERT INTO organizations (id, name, slug, created_at, updated_at) VALUES ('org-1', 'Org', 'org', datetime('now'), datetime('now'))`)

	app := newAdminAnalyticsApp()
	req := httptest.NewRequest(http.MethodGet, "/admin/analytics/users", nil)
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.True(t, body["success"].(bool))
	data := body["data"].(map[string]interface{})
	assert.EqualValues(t, 2, data["total_users"])
}

// ---------------------------------------------------------------------------
// GetAdminOrganizationAnalytics
// ---------------------------------------------------------------------------

func TestGetAdminOrganizationAnalytics_EmptyDB(t *testing.T) {
	db := setupAdminAnalyticsDB(t)
	defer teardownAdminAnalyticsDB(t, db)

	app := newAdminAnalyticsApp()
	req := httptest.NewRequest(http.MethodGet, "/admin/analytics/organizations", nil)
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.True(t, body["success"].(bool))
}

func TestGetAdminOrganizationAnalytics_WithData(t *testing.T) {
	db := setupAdminAnalyticsDB(t)
	defer teardownAdminAnalyticsDB(t, db)

	db.Exec(`INSERT INTO organizations (id, name, slug, subscription_status, subscription_tier, created_at, updated_at) VALUES ('o1', 'Alpha', 'alpha', 'active', 'pro', datetime('now'), datetime('now'))`)
	db.Exec(`INSERT INTO organizations (id, name, slug, subscription_status, subscription_tier, created_at, updated_at) VALUES ('o2', 'Beta', 'beta', 'trial', 'starter', datetime('now'), datetime('now'))`)
	db.Exec(`INSERT INTO users (id, email, organization_id, created_at, updated_at) VALUES ('u1', 'x@x.com', 'o1', datetime('now'), datetime('now'))`)

	app := newAdminAnalyticsApp()
	req := httptest.NewRequest(http.MethodGet, "/admin/analytics/organizations", nil)
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.True(t, body["success"].(bool))
	data := body["data"].(map[string]interface{})
	assert.EqualValues(t, 2, data["total_organizations"])
}

// ---------------------------------------------------------------------------
// GetAdminRevenueAnalytics
// ---------------------------------------------------------------------------

func TestGetAdminRevenueAnalytics_EmptyDB(t *testing.T) {
	db := setupAdminAnalyticsDB(t)
	defer teardownAdminAnalyticsDB(t, db)

	app := newAdminAnalyticsApp()
	req := httptest.NewRequest(http.MethodGet, "/admin/analytics/revenue", nil)
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.True(t, body["success"].(bool))
}

func TestGetAdminRevenueAnalytics_WithPayments(t *testing.T) {
	db := setupAdminAnalyticsDB(t)
	defer teardownAdminAnalyticsDB(t, db)

	db.Exec(`INSERT INTO payments (id, organization_id, subscription_tier, amount, payment_status, payment_type, paid_at) VALUES ('pay-1', 'o1', 'pro', 500.0, 'completed', 'new', datetime('now'))`)
	db.Exec(`INSERT INTO payments (id, organization_id, subscription_tier, amount, payment_status, payment_type, paid_at) VALUES ('pay-2', 'o2', 'starter', 100.0, 'completed', 'renewal', datetime('now'))`)
	db.Exec(`INSERT INTO users (id, email, created_at, updated_at) VALUES ('u1', 'a@a.com', datetime('now'), datetime('now'))`)
	db.Exec(`INSERT INTO organizations (id, name, slug, subscription_status, created_at, updated_at) VALUES ('o1', 'Org', 'org', 'active', datetime('now'), datetime('now'))`)

	app := newAdminAnalyticsApp()
	req := httptest.NewRequest(http.MethodGet, "/admin/analytics/revenue", nil)
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.True(t, body["success"].(bool))
	data := body["data"].(map[string]interface{})
	assert.NotNil(t, data["monthly_recurring_revenue"])
}

// ---------------------------------------------------------------------------
// GetAdminUsageAnalytics
// ---------------------------------------------------------------------------

func TestGetAdminUsageAnalytics_EmptyDB(t *testing.T) {
	db := setupAdminAnalyticsDB(t)
	defer teardownAdminAnalyticsDB(t, db)

	app := newAdminAnalyticsApp()
	req := httptest.NewRequest(http.MethodGet, "/admin/analytics/usage", nil)
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.True(t, body["success"].(bool))
}

func TestGetAdminUsageAnalytics_WithData(t *testing.T) {
	db := setupAdminAnalyticsDB(t)
	defer teardownAdminAnalyticsDB(t, db)

	db.Exec(`INSERT INTO documents (id, document_type, created_by, created_at) VALUES ('d1', 'requisition', 'u1', datetime('now'))`)
	db.Exec(`INSERT INTO documents (id, document_type, created_by, created_at) VALUES ('d2', 'purchase_order', 'u2', datetime('now'))`)
	db.Exec(`INSERT INTO users (id, email, last_login, created_at, updated_at) VALUES ('u1', 'a@a.com', datetime('now'), datetime('now'), datetime('now'))`)

	app := newAdminAnalyticsApp()
	req := httptest.NewRequest(http.MethodGet, "/admin/analytics/usage", nil)
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.True(t, body["success"].(bool))
	data := body["data"].(map[string]interface{})
	// total_api_requests is documents * 10
	assert.EqualValues(t, 20, data["total_api_requests"])
}

// ---------------------------------------------------------------------------
// GetSubscriptionStatistics
// ---------------------------------------------------------------------------

func TestGetSubscriptionStatistics_EmptyDB(t *testing.T) {
	db := setupAdminAnalyticsDB(t)
	defer teardownAdminAnalyticsDB(t, db)

	app := newAdminAnalyticsApp()
	req := httptest.NewRequest(http.MethodGet, "/admin/analytics/subscriptions", nil)
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.True(t, body["success"].(bool))
}

func TestGetSubscriptionStatistics_WithData(t *testing.T) {
	db := setupAdminAnalyticsDB(t)
	defer teardownAdminAnalyticsDB(t, db)

	db.Exec(`INSERT INTO subscription_tiers (id, name, is_active) VALUES ('t1', 'starter', 1)`)
	db.Exec(`INSERT INTO subscription_tiers (id, name, is_active) VALUES ('t2', 'pro', 1)`)
	db.Exec(`INSERT INTO organizations (id, name, slug, subscription_status, created_at, updated_at) VALUES ('o1', 'A', 'a', 'active', datetime('now'), datetime('now'))`)
	db.Exec(`INSERT INTO organizations (id, name, slug, subscription_status, created_at, updated_at) VALUES ('o2', 'B', 'b', 'trial', datetime('now'), datetime('now'))`)
	db.Exec(`INSERT INTO payments (id, organization_id, amount, payment_status, paid_at) VALUES ('p1', 'o1', 200.0, 'completed', datetime('now'))`)

	app := newAdminAnalyticsApp()
	req := httptest.NewRequest(http.MethodGet, "/admin/analytics/subscriptions", nil)
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.True(t, body["success"].(bool))
	data := body["data"].(map[string]interface{})
	assert.EqualValues(t, 2, data["total_tiers"])
	assert.EqualValues(t, 1, data["active_subscriptions"])
	assert.EqualValues(t, 1, data["trial_organizations"])
}

// ---------------------------------------------------------------------------
// GetSystemAlerts
// ---------------------------------------------------------------------------

func TestGetSystemAlerts_EmptyDB(t *testing.T) {
	db := setupAdminAnalyticsDB(t)
	defer teardownAdminAnalyticsDB(t, db)

	app := newAdminAnalyticsApp()
	req := httptest.NewRequest(http.MethodGet, "/admin/alerts", nil)
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.True(t, body["success"].(bool))
}

func TestGetSystemAlerts_FilterBySeverity(t *testing.T) {
	db := setupAdminAnalyticsDB(t)
	defer teardownAdminAnalyticsDB(t, db)

	db.Exec(`INSERT INTO system_alerts (id, severity, status, message, created_at) VALUES ('a1', 'critical', 'open', 'DB down', datetime('now'))`)
	db.Exec(`INSERT INTO system_alerts (id, severity, status, message, created_at) VALUES ('a2', 'warning', 'open', 'High CPU', datetime('now'))`)

	app := newAdminAnalyticsApp()
	req := httptest.NewRequest(http.MethodGet, "/admin/alerts?severity=critical", nil)
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.True(t, body["success"].(bool))
	data := body["data"].(map[string]interface{})
	// Only the critical alert matches — total_count should be 1.
	assert.EqualValues(t, 1, data["total_count"])
}

func TestGetSystemAlerts_FilterByResolved(t *testing.T) {
	db := setupAdminAnalyticsDB(t)
	defer teardownAdminAnalyticsDB(t, db)

	db.Exec(`INSERT INTO system_alerts (id, severity, status, message, created_at) VALUES ('a1', 'info', 'resolved', 'OK', datetime('now'))`)
	db.Exec(`INSERT INTO system_alerts (id, severity, status, message, created_at) VALUES ('a2', 'warning', 'open', 'Warn', datetime('now'))`)

	app := newAdminAnalyticsApp()

	// resolved=true → only resolved alerts
	req := httptest.NewRequest(http.MethodGet, "/admin/alerts?resolved=true", nil)
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeResponse(resp)
	data := body["data"].(map[string]interface{})
	assert.EqualValues(t, 1, data["total_count"])

	// resolved=false → only unresolved alerts
	req2 := httptest.NewRequest(http.MethodGet, "/admin/alerts?resolved=false", nil)
	resp2, err2 := app.Test(req2, -1)
	assert.NoError(t, err2)
	assert.Equal(t, http.StatusOK, resp2.StatusCode)
	body2 := decodeResponse(resp2)
	data2 := body2["data"].(map[string]interface{})
	assert.EqualValues(t, 1, data2["total_count"])
}

// ---------------------------------------------------------------------------
// GetSystemLogs
// ---------------------------------------------------------------------------

func TestGetSystemLogs_EmptyDB(t *testing.T) {
	db := setupAdminAnalyticsDB(t)
	defer teardownAdminAnalyticsDB(t, db)

	app := newAdminAnalyticsApp()
	req := httptest.NewRequest(http.MethodGet, "/admin/logs", nil)
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.True(t, body["success"].(bool))
}

func TestGetSystemLogs_FilterByLevel(t *testing.T) {
	db := setupAdminAnalyticsDB(t)
	defer teardownAdminAnalyticsDB(t, db)

	db.Exec(`INSERT INTO system_logs (id, level, service, message, created_at) VALUES ('l1', 'error', 'api', 'Bad thing', datetime('now'))`)
	db.Exec(`INSERT INTO system_logs (id, level, service, message, created_at) VALUES ('l2', 'info', 'api', 'Good thing', datetime('now'))`)

	app := newAdminAnalyticsApp()
	req := httptest.NewRequest(http.MethodGet, "/admin/logs?level=error", nil)
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	data := body["data"].(map[string]interface{})
	assert.EqualValues(t, 1, data["total_count"])
}

func TestGetSystemLogs_FilterByService(t *testing.T) {
	db := setupAdminAnalyticsDB(t)
	defer teardownAdminAnalyticsDB(t, db)

	db.Exec(`INSERT INTO system_logs (id, level, service, message, created_at) VALUES ('l1', 'info', 'worker', 'Job done', datetime('now'))`)
	db.Exec(`INSERT INTO system_logs (id, level, service, message, created_at) VALUES ('l2', 'info', 'api', 'Req ok', datetime('now'))`)

	app := newAdminAnalyticsApp()
	req := httptest.NewRequest(http.MethodGet, "/admin/logs?service=worker", nil)
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	data := body["data"].(map[string]interface{})
	assert.EqualValues(t, 1, data["total_count"])
}

// ---------------------------------------------------------------------------
// GetSystemMetrics
// ---------------------------------------------------------------------------

func TestGetSystemMetrics_EmptyDB(t *testing.T) {
	db := setupAdminAnalyticsDB(t)
	defer teardownAdminAnalyticsDB(t, db)

	app := newAdminAnalyticsApp()
	req := httptest.NewRequest(http.MethodGet, "/admin/metrics", nil)
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.True(t, body["success"].(bool))
}

func TestGetSystemMetrics_WithData(t *testing.T) {
	db := setupAdminAnalyticsDB(t)
	defer teardownAdminAnalyticsDB(t, db)

	db.Exec(`INSERT INTO system_metrics (id, metric_type, value, recorded_at) VALUES ('m1', 'cpu', 30.0, datetime('now'))`)
	db.Exec(`INSERT INTO system_metrics (id, metric_type, value, recorded_at) VALUES ('m2', 'memory', 55.0, datetime('now'))`)
	db.Exec(`INSERT INTO system_metrics (id, metric_type, value, recorded_at) VALUES ('m3', 'disk', 40.0, datetime('now'))`)
	db.Exec(`INSERT INTO api_request_logs (id, status_code, response_time_ms, created_at) VALUES ('req-1', 200, 120.0, datetime('now'))`)
	db.Exec(`INSERT INTO api_request_logs (id, status_code, response_time_ms, created_at) VALUES ('req-2', 500, 300.0, datetime('now'))`)

	app := newAdminAnalyticsApp()
	req := httptest.NewRequest(http.MethodGet, "/admin/metrics", nil)
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.True(t, body["success"].(bool))
	data := body["data"].(map[string]interface{})
	assert.NotNil(t, data["server"])
	assert.NotNil(t, data["api"])
}

// ---------------------------------------------------------------------------
// ExportAdminAnalytics
// ---------------------------------------------------------------------------

func TestExportAdminAnalytics_EmptyBody(t *testing.T) {
	db := setupAdminAnalyticsDB(t)
	defer teardownAdminAnalyticsDB(t, db)

	app := newAdminAnalyticsApp()
	req := httptest.NewRequest(http.MethodPost, "/admin/analytics/export", nil)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	// Empty body → BodyParser returns error → 400
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestExportAdminAnalytics_WithType(t *testing.T) {
	db := setupAdminAnalyticsDB(t)
	defer teardownAdminAnalyticsDB(t, db)

	app := newAdminAnalyticsApp()
	req := httptest.NewRequest(http.MethodPost, "/admin/analytics/export",
		jsonBody(map[string]interface{}{
			"type":   "users",
			"format": "json",
		}))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.True(t, body["success"].(bool))
	data := body["data"].(map[string]interface{})
	assert.Equal(t, "users", data["type"])
	assert.Equal(t, "ready", data["status"])
}

func TestExportAdminAnalytics_DefaultsApplied(t *testing.T) {
	db := setupAdminAnalyticsDB(t)
	defer teardownAdminAnalyticsDB(t, db)

	app := newAdminAnalyticsApp()
	// Minimal payload — type and format should default to "overview" and "json"
	req := httptest.NewRequest(http.MethodPost, "/admin/analytics/export",
		jsonBody(map[string]interface{}{}))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	data := body["data"].(map[string]interface{})
	assert.Equal(t, "overview", data["type"])
	assert.Equal(t, "json", data["format"])
}

// ---------------------------------------------------------------------------
// RunCustomAdminAnalytics
// ---------------------------------------------------------------------------

func TestRunCustomAdminAnalytics_EmptyBody(t *testing.T) {
	db := setupAdminAnalyticsDB(t)
	defer teardownAdminAnalyticsDB(t, db)

	app := newAdminAnalyticsApp()
	req := httptest.NewRequest(http.MethodPost, "/admin/analytics/custom", nil)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	// Empty body → BodyParser error → 400
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestRunCustomAdminAnalytics_WithMetric(t *testing.T) {
	db := setupAdminAnalyticsDB(t)
	defer teardownAdminAnalyticsDB(t, db)

	db.Exec(`INSERT INTO users (id, email, created_at, updated_at) VALUES ('u1', 'a@b.com', datetime('now'), datetime('now'))`)
	db.Exec(`INSERT INTO organizations (id, name, slug, created_at, updated_at) VALUES ('o1', 'Org', 'org', datetime('now'), datetime('now'))`)

	app := newAdminAnalyticsApp()
	req := httptest.NewRequest(http.MethodPost, "/admin/analytics/custom",
		jsonBody(map[string]interface{}{
			"metric":     "user_growth",
			"start_date": "2026-01-01",
			"end_date":   "2026-03-23",
		}))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.True(t, body["success"].(bool))
	data := body["data"].(map[string]interface{})
	assert.Equal(t, "user_growth", data["metric"])
	assert.EqualValues(t, 1, data["total_users"])
	assert.EqualValues(t, 1, data["total_orgs"])
}

// ---------------------------------------------------------------------------
// GetAdminAnalyticsDashboardConfig
// ---------------------------------------------------------------------------

func TestGetAdminAnalyticsDashboardConfig_EmptyDB(t *testing.T) {
	db := setupAdminAnalyticsDB(t)
	defer teardownAdminAnalyticsDB(t, db)

	app := newAdminAnalyticsApp()
	req := httptest.NewRequest(http.MethodGet, "/admin/analytics/config", nil)
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.True(t, body["success"].(bool))
	// Should return the default config.
	data := body["data"].(map[string]interface{})
	assert.Equal(t, "grid", data["layout"])
}

// ---------------------------------------------------------------------------
// UpdateAdminAnalyticsDashboardConfig
// ---------------------------------------------------------------------------

func TestUpdateAdminAnalyticsDashboardConfig_EmptyBody(t *testing.T) {
	db := setupAdminAnalyticsDB(t)
	defer teardownAdminAnalyticsDB(t, db)

	app := newAdminAnalyticsApp()
	req := httptest.NewRequest(http.MethodPut, "/admin/analytics/config", nil)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	// Empty body → BodyParser error → 400
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestUpdateAdminAnalyticsDashboardConfig_Create(t *testing.T) {
	db := setupAdminAnalyticsDB(t)
	defer teardownAdminAnalyticsDB(t, db)

	app := newAdminAnalyticsApp()
	req := httptest.NewRequest(http.MethodPut, "/admin/analytics/config",
		jsonBody(map[string]interface{}{
			"layout":     "list",
			"time_range": "7d",
			"widgets":    []string{"users", "revenue"},
		}))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.True(t, body["success"].(bool))
}

func TestUpdateAdminAnalyticsDashboardConfig_Update(t *testing.T) {
	db := setupAdminAnalyticsDB(t)
	defer teardownAdminAnalyticsDB(t, db)

	// Pre-seed an existing setting so the UPDATE branch is exercised.
	db.Exec(`INSERT INTO system_settings (id, key, value, created_at, updated_at) VALUES ('cfg-1', 'admin_analytics_dashboard', '{"layout":"grid"}', datetime('now'), datetime('now'))`)

	app := newAdminAnalyticsApp()
	req := httptest.NewRequest(http.MethodPut, "/admin/analytics/config",
		jsonBody(map[string]interface{}{
			"layout":     "compact",
			"time_range": "90d",
		}))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.True(t, body["success"].(bool))
	data := body["data"].(map[string]interface{})
	assert.Equal(t, "compact", data["layout"])
}
