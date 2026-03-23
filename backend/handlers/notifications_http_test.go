package handlers

import (
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/models"
)

// ─────────────────────────────────────────────────────────────────────────────
// Test DB setup with Notification table
// ─────────────────────────────────────────────────────────────────────────────

func setupNotificationTestDB(t *testing.T) {
	t.Helper()
	if config.DB == nil {
		t.Fatal("setupNotificationTestDB: config.DB is nil — call setupTestDB first")
	}
	if err := config.DB.AutoMigrate(&models.Notification{}); err != nil {
		t.Fatalf("setupNotificationTestDB AutoMigrate: %v", err)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// App factories
// ─────────────────────────────────────────────────────────────────────────────

func newNotificationApp(t *testing.T) *fiber.App {
	t.Helper()
	h := NewNotificationHandler()
	auth := withTenantCtx(testOrgID, testUserID, testUserRole)

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		},
	})
	app.Get("/notifications", auth, h.GetNotifications)
	app.Get("/notifications/recent", auth, h.GetRecentNotifications)
	app.Get("/notifications/stats", auth, h.GetNotificationStats)
	app.Post("/notifications/mark-as-read", auth, h.MarkAsRead)
	app.Post("/notifications/mark-all-as-read", auth, h.MarkAllAsRead)
	app.Delete("/notifications/:id", auth, h.DeleteNotification)
	return app
}

func newNotificationAppNoAuth(t *testing.T) *fiber.App {
	t.Helper()
	h := NewNotificationHandler()

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		},
	})
	app.Use(recover.New())
	app.Get("/notifications", h.GetNotifications)
	app.Get("/notifications/recent", h.GetRecentNotifications)
	app.Get("/notifications/stats", h.GetNotificationStats)
	app.Post("/notifications/mark-as-read", h.MarkAsRead)
	app.Post("/notifications/mark-all-as-read", h.MarkAllAsRead)
	app.Delete("/notifications/:id", h.DeleteNotification)
	return app
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

func makeNotification(t *testing.T, orgID, recipientID string) models.Notification {
	t.Helper()
	n := models.Notification{
		ID:             uuid.New().String(),
		OrganizationID: orgID,
		RecipientID:    recipientID,
		Type:           "approval_required",
		DocumentID:     uuid.New().String(),
		DocumentType:   "requisition",
		Subject:        "Approval Required",
		Body:           "Please review and approve the requisition.",
		Sent:           false,
		IsRead:         false,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	if err := config.DB.Create(&n).Error; err != nil {
		t.Fatalf("makeNotification: %v", err)
	}
	return n
}

// ─────────────────────────────────────────────────────────────────────────────
// GetNotifications
// ─────────────────────────────────────────────────────────────────────────────

// Without tenant context the handler performs safe checks and returns 400
// (not a panic), so we verify a non-200 response.
func TestGetNotifications_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupNotificationTestDB(t)

	app := newNotificationAppNoAuth(t)
	resp := testRequest(app, http.MethodGet, "/notifications", nil)
	if resp.StatusCode == http.StatusOK {
		t.Errorf("expected non-200 without auth context, got 200")
	}
}

func TestGetNotifications_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupNotificationTestDB(t)

	app := newNotificationApp(t)
	resp := testRequest(app, http.MethodGet, "/notifications", nil)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	body := decodeResponse(resp)
	if body == nil {
		t.Fatal("expected JSON body")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// GetRecentNotifications
// ─────────────────────────────────────────────────────────────────────────────

func TestGetRecentNotifications_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupNotificationTestDB(t)

	app := newNotificationAppNoAuth(t)
	resp := testRequest(app, http.MethodGet, "/notifications/recent", nil)
	if resp.StatusCode == http.StatusOK {
		t.Errorf("expected non-200 without auth context, got 200")
	}
}

func TestGetRecentNotifications_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupNotificationTestDB(t)

	app := newNotificationApp(t)
	resp := testRequest(app, http.MethodGet, "/notifications/recent", nil)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// GetNotificationStats
// ─────────────────────────────────────────────────────────────────────────────

func TestGetNotificationStats_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupNotificationTestDB(t)

	app := newNotificationAppNoAuth(t)
	resp := testRequest(app, http.MethodGet, "/notifications/stats", nil)
	if resp.StatusCode == http.StatusOK {
		t.Errorf("expected non-200 without auth context, got 200")
	}
}

func TestGetNotificationStats_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupNotificationTestDB(t)

	app := newNotificationApp(t)
	resp := testRequest(app, http.MethodGet, "/notifications/stats", nil)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// MarkAsRead
// ─────────────────────────────────────────────────────────────────────────────

// MarkAsRead directly type-asserts c.Locals → panics (500) without auth.
func TestMarkAsRead_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupNotificationTestDB(t)

	app := newNotificationAppNoAuth(t)
	resp := testRequest(app, http.MethodPost, "/notifications/mark-as-read",
		map[string]interface{}{"notificationIds": []string{"id1"}})
	if resp.StatusCode == http.StatusOK {
		t.Errorf("expected non-200 without auth, got 200")
	}
}

func TestMarkAsRead_MissingIds(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupNotificationTestDB(t)

	app := newNotificationApp(t)
	// Empty notificationIds fails the `min=1` validator.
	resp := testRequest(app, http.MethodPost, "/notifications/mark-as-read",
		map[string]interface{}{"notificationIds": []string{}})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for empty notificationIds, got %d", resp.StatusCode)
	}
}

func TestMarkAsRead_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupNotificationTestDB(t)

	// Create a notification owned by the test user.
	n := makeNotification(t, testOrgID, testUserID)
	app := newNotificationApp(t)

	resp := testRequest(app, http.MethodPost, "/notifications/mark-as-read",
		map[string]interface{}{"notificationIds": []string{n.ID}})
	if resp.StatusCode != http.StatusOK {
		body := decodeResponse(resp)
		t.Errorf("expected 200, got %d; body=%v", resp.StatusCode, body)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// MarkAllAsRead
// ─────────────────────────────────────────────────────────────────────────────

func TestMarkAllAsRead_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupNotificationTestDB(t)

	app := newNotificationAppNoAuth(t)
	resp := testRequest(app, http.MethodPost, "/notifications/mark-all-as-read", nil)
	if resp.StatusCode == http.StatusOK {
		t.Errorf("expected non-200 without auth, got 200")
	}
}

func TestMarkAllAsRead_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupNotificationTestDB(t)

	app := newNotificationApp(t)
	resp := testRequest(app, http.MethodPost, "/notifications/mark-all-as-read", nil)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// DeleteNotification
// ─────────────────────────────────────────────────────────────────────────────

func TestDeleteNotification_NoAuth(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupNotificationTestDB(t)

	app := newNotificationAppNoAuth(t)
	resp := testRequest(app, http.MethodDelete, "/notifications/"+uuid.New().String(), nil)
	if resp.StatusCode == http.StatusOK {
		t.Errorf("expected non-200 without auth, got 200")
	}
}

func TestDeleteNotification_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupNotificationTestDB(t)

	app := newNotificationApp(t)
	resp := testRequest(app, http.MethodDelete, "/notifications/"+uuid.New().String(), nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404 for non-existent notification, got %d", resp.StatusCode)
	}
}

func TestDeleteNotification_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	setupNotificationTestDB(t)

	n := makeNotification(t, testOrgID, testUserID)
	app := newNotificationApp(t)
	resp := testRequest(app, http.MethodDelete, "/notifications/"+n.ID, nil)
	if resp.StatusCode != http.StatusOK {
		body := decodeResponse(resp)
		t.Errorf("expected 200, got %d; body=%v", resp.StatusCode, body)
	}
}
