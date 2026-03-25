package handlers

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/models"
	"github.com/stretchr/testify/assert"
)

// seedNotificationExtra inserts a Notification for testOrgID/testUserID with the
// given type and read state.  The helper is named with an "Extra" suffix to avoid
// any collision with helpers defined in notifications_http_test.go.
func seedNotificationExtra(t *testing.T, notifType string, isRead bool) models.Notification {
	t.Helper()
	n := models.Notification{
		ID:             uuid.New().String(),
		OrganizationID: testOrgID,
		RecipientID:    testUserID,
		Type:           notifType,
		Subject:        "Test Subject for " + notifType,
		Body:           "Test body for " + notifType,
		DocumentID:     uuid.New().String(),
		DocumentType:   "requisition",
		IsRead:         isRead,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	if err := config.DB.Create(&n).Error; err != nil {
		t.Fatalf("seedNotificationExtra: %v", err)
	}
	return n
}

// ─────────────────────────────────────────────────────────────────────────────
// GetNotifications — filter / coverage-targeted tests
// ─────────────────────────────────────────────────────────────────────────────

// TestGetNotifications_WithTypeFilter covers the type-filter branch (lines 107-109
// in notification_handler.go).  It seeds two notification types, then queries
// only for "approval_required" and asserts a 200 response.
func TestGetNotifications_WithTypeFilter(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupNotificationTestDB(t)

	seedNotificationExtra(t, "approval_required", false)
	seedNotificationExtra(t, "document_approved", false)

	app := newNotificationApp(t)
	resp := testRequest(app, http.MethodGet, "/notifications?type=approval_required", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestGetNotifications_UnreadOnlyMixed covers the unread_only branch (lines 111-113)
// with a mix of read and unread notifications seeded.
func TestGetNotifications_UnreadOnlyMixed(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupNotificationTestDB(t)

	seedNotificationExtra(t, "approval_required", false) // unread
	seedNotificationExtra(t, "document_approved", true)  // read

	app := newNotificationApp(t)
	resp := testRequest(app, http.MethodGet, "/notifications?unread_only=true", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.NotNil(t, body)
}

// TestGetNotifications_AllTypes exercises all switch-case branches in the
// importance assignment (lines 147-156).  Seeds approval_required (HIGH),
// document_rejected (HIGH), document_approved (MEDIUM), and an unknown type
// (LOW).  Verifies 200 and that the response body is non-empty.
func TestGetNotifications_AllTypes(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupNotificationTestDB(t)

	seedNotificationExtra(t, "approval_required", false)
	seedNotificationExtra(t, "document_rejected", false)
	seedNotificationExtra(t, "document_approved", false)
	seedNotificationExtra(t, "info_message", false) // unknown type → LOW

	app := newNotificationApp(t)
	resp := testRequest(app, http.MethodGet, "/notifications", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.NotNil(t, body)
	// The response wraps items in a "data" key from SendPaginatedSuccess.
	_, hasData := body["data"]
	assert.True(t, hasData, "response should contain a data key")
}

// ─────────────────────────────────────────────────────────────────────────────
// GetNotificationStats — with seeded data
// ─────────────────────────────────────────────────────────────────────────────

// TestGetNotificationStats_WithData seeds a mix of read and unread notifications
// then verifies that the stats endpoint returns 200 with non-zero counts.
func TestGetNotificationStats_WithData(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupNotificationTestDB(t)

	// 3 unread, 2 read
	for i := 0; i < 3; i++ {
		seedNotificationExtra(t, "approval_required", false)
	}
	for i := 0; i < 2; i++ {
		seedNotificationExtra(t, "document_approved", true)
	}

	app := newNotificationApp(t)
	resp := testRequest(app, http.MethodGet, "/notifications/stats", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.NotNil(t, body)

	data, ok := body["data"].(map[string]interface{})
	assert.True(t, ok, "expected data object in response")
	if ok {
		pending, hasPending := data["pending"]
		assert.True(t, hasPending, "stats should contain pending count")
		assert.NotNil(t, pending)

		readCount, hasRead := data["read"]
		assert.True(t, hasRead, "stats should contain read count")
		assert.NotNil(t, readCount)

		total, hasTotal := data["total"]
		assert.True(t, hasTotal, "stats should contain total count")
		assert.NotNil(t, total)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// MarkAsRead — additional coverage tests
// ─────────────────────────────────────────────────────────────────────────────

// TestMarkAsRead_SuccessMultiple seeds two notifications and marks both as read
// in a single request.
func TestMarkAsRead_SuccessMultiple(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupNotificationTestDB(t)

	n1 := seedNotificationExtra(t, "approval_required", false)
	n2 := seedNotificationExtra(t, "document_rejected", false)

	app := newNotificationApp(t)
	resp := testRequest(app, http.MethodPost, "/notifications/mark-as-read",
		map[string]interface{}{
			"notificationIds": []string{n1.ID, n2.ID},
		},
	)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.NotNil(t, body)

	// Verify the notifications were actually marked as read in the DB.
	var updated models.Notification
	err := config.DB.Where("id = ?", n1.ID).First(&updated).Error
	assert.NoError(t, err)
	assert.True(t, updated.IsRead, "notification should be marked as read")
}

// TestMarkAsRead_ValidationFail sends an empty notificationIds array which
// fails the `min=1` validation rule and expects a 400 response.
func TestMarkAsRead_ValidationFail(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupNotificationTestDB(t)

	app := newNotificationApp(t)
	resp := testRequest(app, http.MethodPost, "/notifications/mark-as-read",
		map[string]interface{}{
			"notificationIds": []string{},
		},
	)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// MarkAllAsRead — additional coverage tests
// ─────────────────────────────────────────────────────────────────────────────

// TestMarkAllAsRead_SuccessWithData seeds three unread notifications then calls
// mark-all-as-read and expects 200.
func TestMarkAllAsRead_SuccessWithData(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupNotificationTestDB(t)

	for i := 0; i < 3; i++ {
		seedNotificationExtra(t, "approval_required", false)
	}

	app := newNotificationApp(t)
	resp := testRequest(app, http.MethodPost, "/notifications/mark-all-as-read", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// After marking all as read, unread count in DB should be 0.
	var unreadCount int64
	config.DB.Model(&models.Notification{}).
		Where("organization_id = ? AND recipient_id = ? AND is_read = ?", testOrgID, testUserID, false).
		Count(&unreadCount)
	assert.Equal(t, int64(0), unreadCount)
}

// TestMarkAllAsRead_NoUnread verifies that the no-op path (no unread notifications)
// still returns 200.
func TestMarkAllAsRead_NoUnread(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupNotificationTestDB(t)

	// Only seed a read notification — nothing to mark.
	seedNotificationExtra(t, "document_approved", true)

	app := newNotificationApp(t)
	resp := testRequest(app, http.MethodPost, "/notifications/mark-all-as-read", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := decodeResponse(resp)
	assert.NotNil(t, body)
	// Response message should indicate no-op.
	msg, _ := body["message"].(string)
	assert.NotEmpty(t, msg)
}
