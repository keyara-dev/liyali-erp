package handlers

// coverage_boost_http_test.go — Additional tests targeting specific low-coverage
// branches identified by go tool cover analysis:
//
//  • notifications.go::GetNotifications      (48.9%) — unreadOnly, type filter, since param
//  • notifications.go::MarkAllNotificationsAsRead (47.1%) — pending-notifications path
//  • notification_handler.go::GetRecentNotifications (50%) — seeded-notifications loop
//  • notification_handler.go::MarkAllAsRead (50%) — seeded unread path
//  • admin_subscription_handler.go::GetAllSubscriptionTiers (50%) — loop body with seeded tier
//  • subscription_handler.go::CheckFeatureAccess (35.7%) — missing feature param branch
//  • Various 50% functions that need a second branch covered

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

// setupCoverageBoostDB sets up a test DB, migrates the notifications table,
// and seeds user / organization rows that service queries rely on.
func setupCoverageBoostDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupTestDB(t)
	setupNotificationsTableWithDB(t, db)
	return db
}

// seedPendingNotification inserts a notification with sent=false for the given user.
func seedPendingNotification(t *testing.T, orgID, userID, notifType string) string {
	t.Helper()
	id := uuid.New().String()
	sql := `INSERT INTO notifications (id, organization_id, recipient_id, type, document_id, document_type, subject, body, sent, is_read, created_at, updated_at)
	        VALUES (?, ?, ?, ?, ?, ?, ?, ?, 0, 0, ?, ?)`
	if err := config.DB.Exec(sql,
		id, orgID, userID, notifType,
		uuid.New().String(), "requisition",
		"Test notification", "Notification body",
		time.Now(), time.Now(),
	).Error; err != nil {
		t.Fatalf("seedPendingNotification: %v", err)
	}
	return id
}

// seedUnreadNotification inserts a notification with is_read=false.
func seedUnreadNotification(t *testing.T, orgID, userID, notifType string) string {
	t.Helper()
	id := uuid.New().String()
	sql := `INSERT INTO notifications (id, organization_id, recipient_id, type, document_id, document_type, subject, body, sent, is_read, created_at, updated_at)
	        VALUES (?, ?, ?, ?, ?, ?, ?, ?, 0, 0, ?, ?)`
	if err := config.DB.Exec(sql,
		id, orgID, userID, notifType,
		uuid.New().String(), "requisition",
		"Unread notification", "Some body",
		time.Now(), time.Now(),
	).Error; err != nil {
		t.Fatalf("seedUnreadNotification: %v", err)
	}
	return id
}

// ─────────────────────────────────────────────────────────────────────────────
// notifications.go — GetNotifications additional paths
// ─────────────────────────────────────────────────────────────────────────────

// TestGetNotifications_UnreadOnly exercises the `unreadOnly=true` branch in
// notifications.go::GetNotifications, which calls GetPendingNotifications.
func TestGetNotifications_UnreadOnly(t *testing.T) {
	db := setupCoverageBoostDB(t)
	defer teardownTestDB(t, db)

	// Seed one pending notification for the test user.
	seedPendingNotification(t, testOrgID, testUserID, "approval_required")

	app := newNotificationsApp(withTenantCtx(testOrgID, testUserID, testUserRole))
	resp := testRequest(app, http.MethodGet, "/notifications/?unreadOnly=true", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])
}

// TestGetNotifications_UnreadOnly_Empty exercises the unreadOnly path when
// there are no pending notifications (slice conversion still runs).
func TestGetNotifications_UnreadOnly_Empty(t *testing.T) {
	db := setupCoverageBoostDB(t)
	defer teardownTestDB(t, db)

	app := newNotificationsApp(withTenantCtx(testOrgID, testUserID, testUserRole))
	resp := testRequest(app, http.MethodGet, "/notifications/?unreadOnly=true", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestGetNotifications_TypeFilter exercises the `type=<type>` filter branch.
func TestGetNotifications_TypeFilter(t *testing.T) {
	db := setupCoverageBoostDB(t)
	defer teardownTestDB(t, db)

	// Seed one notification with matching type.
	seedPendingNotification(t, testOrgID, testUserID, "document_approved")

	app := newNotificationsApp(withTenantCtx(testOrgID, testUserID, testUserRole))
	resp := testRequest(app, http.MethodGet, "/notifications/?type=document_approved", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestGetNotifications_TypeFilter_Empty exercises the type filter with no matches.
func TestGetNotifications_TypeFilter_Empty(t *testing.T) {
	db := setupCoverageBoostDB(t)
	defer teardownTestDB(t, db)

	app := newNotificationsApp(withTenantCtx(testOrgID, testUserID, testUserRole))
	resp := testRequest(app, http.MethodGet, "/notifications/?type=nonexistent_type", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestGetNotifications_SinceParam exercises the since= query parameter parsing.
func TestGetNotifications_SinceParam(t *testing.T) {
	db := setupCoverageBoostDB(t)
	defer teardownTestDB(t, db)

	since := time.Now().AddDate(0, -2, 0).Format(time.RFC3339)
	app := newNotificationsApp(withTenantCtx(testOrgID, testUserID, testUserRole))
	resp := testRequest(app, http.MethodGet, fmt.Sprintf("/notifications/?since=%s", since), nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestGetNotifications_Pagination exercises the pagination logic inside the
// default branch (start/end slice computation) by seeding multiple notifications.
func TestGetNotifications_Pagination(t *testing.T) {
	db := setupCoverageBoostDB(t)
	defer teardownTestDB(t, db)

	// Seed several notifications so the pagination slice is exercised.
	for i := 0; i < 5; i++ {
		seedPendingNotification(t, testOrgID, testUserID, "approval_required")
	}

	app := newNotificationsApp(withTenantCtx(testOrgID, testUserID, testUserRole))
	resp := testRequest(app, http.MethodGet, "/notifications/?page=1&limit=3", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// notifications.go — MarkAllNotificationsAsRead additional paths
// ─────────────────────────────────────────────────────────────────────────────

// TestMarkAllNotificationsAsRead_WithPending exercises the path where pending
// notifications exist, triggering MarkMultipleAsRead.
func TestMarkAllNotificationsAsRead_WithPending(t *testing.T) {
	db := setupCoverageBoostDB(t)
	defer teardownTestDB(t, db)

	// Seed a pending notification.
	seedPendingNotification(t, testOrgID, testUserID, "approval_required")

	app := newNotificationsApp(withTenantCtx(testOrgID, testUserID, testUserRole))
	resp := testRequest(app, http.MethodPut, "/notifications/read-all", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])
}

// TestMarkAllNotificationsAsRead_NoneUnread exercises the empty-pending path.
func TestMarkAllNotificationsAsRead_NoneUnread(t *testing.T) {
	db := setupCoverageBoostDB(t)
	defer teardownTestDB(t, db)

	app := newNotificationsApp(withTenantCtx(testOrgID, testUserID, testUserRole))
	resp := testRequest(app, http.MethodPut, "/notifications/read-all", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// notification_handler.go — GetRecentNotifications with seeded data
// ─────────────────────────────────────────────────────────────────────────────

// TestGetRecentNotifications_WithSeededData exercises the notification loop
// body including the importance/type switch and documentID branch.
func TestGetRecentNotifications_WithSeededData(t *testing.T) {
	db := setupCoverageBoostDB(t)
	defer teardownTestDB(t, db)

	// Seed notifications of different types.
	for _, typ := range []string{"approval_required", "document_rejected", "document_approved", "general"} {
		seedUnreadNotification(t, testOrgID, testUserID, typ)
	}

	app := newNotificationApp(t)
	resp := testRequest(app, http.MethodGet, "/notifications/recent", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestGetRecentNotifications_WithDocumentID exercises the documentID lookup
// branch inside the notification loop (getDocumentNumber helper).
func TestGetRecentNotifications_WithDocumentID(t *testing.T) {
	db := setupCoverageBoostDB(t)
	defer teardownTestDB(t, db)

	// Seed a requisition so the document number lookup succeeds.
	reqID := uuid.New().String()
	if err := db.Create(&models.Requisition{
		ID:             reqID,
		OrganizationID: testOrgID,
		DocumentNumber: "REQ-NOTIF-001",
		Title:          "Notification Test Req",
		Status:         "PENDING",
		RequesterId:    testUserID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}).Error; err != nil {
		t.Fatalf("create requisition: %v", err)
	}

	// Seed notification with the requisition's document ID.
	id := uuid.New().String()
	sql := `INSERT INTO notifications (id, organization_id, recipient_id, type, document_id, document_type, subject, body, sent, is_read, created_at, updated_at)
	        VALUES (?, ?, ?, ?, ?, ?, ?, ?, 0, 0, ?, ?)`
	if err := config.DB.Exec(sql,
		id, testOrgID, testUserID, "approval_required",
		reqID, "requisition",
		"Approval needed", "Please approve REQ-NOTIF-001",
		time.Now(), time.Now(),
	).Error; err != nil {
		t.Fatalf("seedNotificationWithDocID: %v", err)
	}

	app := newNotificationApp(t)
	resp := testRequest(app, http.MethodGet, "/notifications/recent", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// notification_handler.go — MarkAllAsRead with seeded unread notifications
// ─────────────────────────────────────────────────────────────────────────────

// TestMarkAllAsRead_WithUnread exercises the `len(notificationIDs) > 0` path
// in MarkAllAsRead, calling MarkMultipleAsRead on real seeded data.
func TestMarkAllAsRead_WithUnread(t *testing.T) {
	db := setupCoverageBoostDB(t)
	defer teardownTestDB(t, db)

	// Seed an unread notification.
	seedUnreadNotification(t, testOrgID, testUserID, "approval_required")

	app := newNotificationApp(t)
	resp := testRequest(app, http.MethodPost, "/notifications/mark-all-as-read", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// admin_subscription_handler.go — GetAllSubscriptionTiers loop body
// ─────────────────────────────────────────────────────────────────────────────

// newAdminSubscriptionBoostApp registers GetAllSubscriptionTiers.
func newAdminSubscriptionBoostApp(t *testing.T) *fiber.App {
	t.Helper()
	app := fiber.New(fiber.Config{ErrorHandler: func(c *fiber.Ctx, err error) error {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}})
	app.Use(recover.New())
	app.Use(withTenantCtx(testOrgID, testUserID, testUserRole))
	app.Get("/admin/subscription-tiers", GetAllSubscriptionTiers)
	app.Get("/admin/subscription-tiers/:id", GetSubscriptionTierByID)
	return app
}

// insertTierRaw inserts a subscription tier row via raw SQL.
func insertTierRaw(t *testing.T, db *gorm.DB, id, name string) {
	t.Helper()
	sql := `INSERT INTO subscription_tiers (id, name, display_name, description, price_monthly, price_yearly, max_workspaces, max_team_members, max_documents, max_workflows, max_custom_roles, features, is_active, sort_order, created_at, updated_at)
	        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))`
	if err := db.Exec(sql, id, name, name+" Plan", "Test tier", 0, 0, 1, 5, 100, 1, 0, `["basic_workflows"]`, 1, 1).Error; err != nil {
		t.Fatalf("insertTierRaw: %v", err)
	}
}

// TestGetAllSubscriptionTiers_WithSeededTier exercises the tier-response loop
// body (GetFeatureList, orgCount query, TierResponse construction).
func TestGetAllSubscriptionTiers_WithSeededTier(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupSubscriptionTiersTable(t, db)
	insertTierRaw(t, db, "tier-starter", "starter")

	app := newAdminSubscriptionBoostApp(t)
	resp := testRequest(app, http.MethodGet, "/admin/subscription-tiers", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])
}

// TestGetSubscriptionTierByID_Found exercises the success path with a seeded tier.
func TestGetSubscriptionTierByID_Found(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupSubscriptionTiersTable(t, db)
	insertTierRaw(t, db, "tier-pro", "pro")

	app := newAdminSubscriptionBoostApp(t)
	resp := testRequest(app, http.MethodGet, "/admin/subscription-tiers/tier-pro", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// subscription_handler.go — CheckFeatureAccess missing-feature branch
// ─────────────────────────────────────────────────────────────────────────────

// newSubscriptionCheckApp registers CheckFeatureAccess with a /:id param so
// the organizationID check passes and we can reach the featureName check.
func newSubscriptionCheckApp() *fiber.App {
	app := fiber.New(fiber.Config{ErrorHandler: func(c *fiber.Ctx, err error) error {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}})
	app.Use(recover.New())

	h := &SubscriptionHandler{subscriptionService: nil}
	app.Get("/subscriptions/:id/check-feature", h.CheckFeatureAccess)
	app.Get("/subscriptions/:id/trial-status", h.GetOrganizationTrialStatus)
	app.Get("/subscriptions/:id", h.GetOrganizationSubscription)

	return app
}

// TestCheckFeatureAccess_MissingFeatureName verifies the featureName==""
// validation branch (orgID is present, featureName is absent → 400).
func TestCheckFeatureAccess_MissingFeatureName(t *testing.T) {
	app := newSubscriptionCheckApp()

	// No ?feature= query param → featureName is "" → 400.
	resp := testRequest(app, http.MethodGet, "/subscriptions/org-001/check-feature", nil)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

// TestCheckFeatureAccess_WithFeature_NilService verifies that the service call
// is reached (panics → 500 via recover) when both params are present.
func TestCheckFeatureAccess_WithFeature_NilService(t *testing.T) {
	app := newSubscriptionCheckApp()

	resp := testRequest(app, http.MethodGet, "/subscriptions/org-001/check-feature?feature=analytics", nil)
	// Nil service panics → recover catches → non-200.
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}

// TestGetOrganizationTrialStatus_FromCheckApp exercises the /:id path
// (organizationID present) → proceeds to service call → nil service → non-200.
func TestGetOrganizationTrialStatus_FromCheckApp(t *testing.T) {
	app := newSubscriptionCheckApp()

	resp := testRequest(app, http.MethodGet, "/subscriptions/org-001/trial-status", nil)
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}
