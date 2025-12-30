package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cozyCodr/liyali-gateway/internal/db"
	"github.com/cozyCodr/liyali-gateway/internal/handlers"
	"github.com/cozyCodr/liyali-gateway/internal/middleware"
	"github.com/cozyCodr/liyali-gateway/internal/repository"
	"github.com/cozyCodr/liyali-gateway/internal/services"
	"github.com/cozyCodr/liyali-gateway/internal/utils"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotificationEndpoints(t *testing.T) {
	// Setup test database and repositories
	pool := setupTestDB(t)
	defer pool.Close()

	queries := db.New(pool)
	ctx := context.Background()

	// Initialize repositories
	userRepo := repository.NewUserRepository(queries)
	sessionRepo := repository.NewSessionRepository(queries)
	passwordResetRepo := repository.NewPasswordResetRepository(queries)
	notificationRepo := repository.NewNotificationRepository(queries)

	// Initialize services
	authService := services.NewAuthService(userRepo, sessionRepo, passwordResetRepo, "test-secret-key")

	// Initialize handlers
	notificationHandler := handlers.NewNotificationHandler(*notificationRepo)
	authMiddleware := middleware.NewAuthMiddleware(authService)

	// Create test user
	testUser := createTestUser(t, queries, "notification@test.com", "USER")
	token := generateTestToken(t, authService, testUser)
	userID := utils.PgtypeToUUID(testUser.ID)

	// Create test notification
	notification, err := queries.CreateNotification(ctx, db.CreateNotificationParams{
		UserID:       utils.UUIDToPgtype(userID),
		Type:         "TASK_ASSIGNED",
		Title:        "Test Notification",
		Message:      "This is a test notification",
		RelatedID:    utils.UUIDToPgtype(userID),
		SentViaEmail: pgtype.Bool{Bool: false, Valid: true},
	})
	require.NoError(t, err)

	// Setup Fiber app
	app := fiber.New()
	api := app.Group("/api")

	notifications := api.Group("/notifications", authMiddleware.Authenticate)
	notifications.Get("/", notificationHandler.GetNotifications)
	notifications.Get("/unread", notificationHandler.GetUnreadNotifications)
	notifications.Get("/unread/count", notificationHandler.GetUnreadCount)
	notifications.Get("/:id", notificationHandler.GetNotificationByID)
	notifications.Post("/:id/read", notificationHandler.MarkAsRead)
	notifications.Post("/read-all", notificationHandler.MarkAllAsRead)
	notifications.Delete("/:id", notificationHandler.DeleteNotification)

	t.Run("GET /api/notifications - success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/notifications", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Contains(t, result, "notifications")
		assert.Contains(t, result, "total")
		assert.Contains(t, result, "unread_count")
	})

	t.Run("GET /api/notifications/unread - success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/notifications/unread", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Contains(t, result, "notifications")
		assert.Contains(t, result, "unread_count")
	})

	t.Run("GET /api/notifications/unread/count - success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/notifications/unread/count", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Contains(t, result, "unread_count")
	})

	t.Run("GET /api/notifications/:id - success", func(t *testing.T) {
		notificationID := utils.PgtypeToUUID(notification.ID)
		req := httptest.NewRequest(http.MethodGet, "/api/notifications/"+notificationID.String(), nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("POST /api/notifications/:id/read - success", func(t *testing.T) {
		notificationID := utils.PgtypeToUUID(notification.ID)
		req := httptest.NewRequest(http.MethodPost, "/api/notifications/"+notificationID.String()+"/read", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("POST /api/notifications/read-all - success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/notifications/read-all", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Contains(t, result, "message")
	})

	t.Run("DELETE /api/notifications/:id - success", func(t *testing.T) {
		// Create a new notification to delete
		deleteNotification, err := queries.CreateNotification(ctx, db.CreateNotificationParams{
			UserID:       utils.UUIDToPgtype(userID),
			Type:         "TASK_ASSIGNED",
			Title:        "To Delete",
			Message:      "This notification will be deleted",
			RelatedID:    utils.UUIDToPgtype(userID),
			SentViaEmail: pgtype.Bool{Bool: false, Valid: true},
		})
		require.NoError(t, err)

		notificationID := utils.PgtypeToUUID(deleteNotification.ID)
		req := httptest.NewRequest(http.MethodDelete, "/api/notifications/"+notificationID.String(), nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("GET /api/notifications - unauthorized", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/notifications", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}
