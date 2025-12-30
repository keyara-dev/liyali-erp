package unit

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cozyCodr/liyali-gateway/internal/db"
	"github.com/cozyCodr/liyali-gateway/internal/handlers"
	"github.com/cozyCodr/liyali-gateway/internal/utils"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockNotificationRepository struct {
	mock.Mock
}

func (m *MockNotificationRepository) ListNotificationsByUser(ctx interface{}, userID uuid.UUID, limit, offset int32) ([]db.Notification, error) {
	args := m.Called(ctx, userID, limit, offset)
	return args.Get(0).([]db.Notification), args.Error(1)
}

func (m *MockNotificationRepository) ListUnreadNotificationsByUser(ctx interface{}, userID uuid.UUID, limit, offset int32) ([]db.Notification, error) {
	args := m.Called(ctx, userID, limit, offset)
	return args.Get(0).([]db.Notification), args.Error(1)
}

func (m *MockNotificationRepository) CountUnreadNotificationsByUser(ctx interface{}, userID uuid.UUID) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockNotificationRepository) GetNotificationByID(ctx interface{}, id uuid.UUID) (*db.Notification, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*db.Notification), args.Error(1)
}

func (m *MockNotificationRepository) MarkNotificationAsRead(ctx interface{}, id uuid.UUID) (*db.Notification, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*db.Notification), args.Error(1)
}

func (m *MockNotificationRepository) MarkAllNotificationsAsRead(ctx interface{}, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockNotificationRepository) DeleteNotification(ctx interface{}, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func setupNotificationTestApp(mockRepo *MockNotificationRepository) *fiber.App {
	app := fiber.New()
	handler := handlers.NewNotificationHandler(*mockRepo)

	// Mock authentication by setting user context
	app.Use(func(c fiber.Ctx) error {
		c.Locals("userID", uuid.MustParse("12345678-1234-1234-1234-123456789012"))
		c.Locals("userEmail", "test@example.com")
		c.Locals("userRole", "USER")
		return c.Next()
	})

	app.Get("/notifications", handler.GetNotifications)
	app.Get("/notifications/unread", handler.GetUnreadNotifications)
	app.Get("/notifications/unread/count", handler.GetUnreadCount)
	app.Get("/notifications/:id", handler.GetNotificationByID)
	app.Post("/notifications/:id/read", handler.MarkAsRead)
	app.Post("/notifications/read-all", handler.MarkAllAsRead)
	app.Delete("/notifications/:id", handler.DeleteNotification)

	return app
}

func TestNotificationHandler_GetNotifications(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockRepo := new(MockNotificationRepository)
		app := setupNotificationTestApp(mockRepo)

		userID := uuid.MustParse("12345678-1234-1234-1234-123456789012")
		notifications := []db.Notification{
			{
				ID:      utils.UUIDToPgtype(uuid.New()),
				UserID:  utils.UUIDToPgtype(userID),
				Type:    "TASK_ASSIGNED",
				Title:   "Test Notification",
				Message: "Test message",
			},
		}

		mockRepo.On("ListNotificationsByUser", mock.Anything, userID, int32(20), int32(0)).Return(notifications, nil)
		mockRepo.On("CountUnreadNotificationsByUser", mock.Anything, userID).Return(int64(1), nil)

		req := httptest.NewRequest(http.MethodGet, "/notifications", nil)
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		assert.Contains(t, result, "notifications")
		assert.Contains(t, result, "unread_count")
		assert.Equal(t, float64(1), result["unread_count"])

		mockRepo.AssertExpectations(t)
	})

	t.Run("with pagination", func(t *testing.T) {
		mockRepo := new(MockNotificationRepository)
		app := setupNotificationTestApp(mockRepo)

		userID := uuid.MustParse("12345678-1234-1234-1234-123456789012")

		mockRepo.On("ListNotificationsByUser", mock.Anything, userID, int32(10), int32(5)).Return([]db.Notification{}, nil)
		mockRepo.On("CountUnreadNotificationsByUser", mock.Anything, userID).Return(int64(0), nil)

		req := httptest.NewRequest(http.MethodGet, "/notifications?limit=10&offset=5", nil)
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		assert.Equal(t, float64(10), result["limit"])
		assert.Equal(t, float64(5), result["offset"])

		mockRepo.AssertExpectations(t)
	})
}

func TestNotificationHandler_GetUnreadNotifications(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockRepo := new(MockNotificationRepository)
		app := setupNotificationTestApp(mockRepo)

		userID := uuid.MustParse("12345678-1234-1234-1234-123456789012")
		unreadNotifications := []db.Notification{
			{
				ID:      utils.UUIDToPgtype(uuid.New()),
				UserID:  utils.UUIDToPgtype(userID),
				Type:    "TASK_ASSIGNED",
				Title:   "Unread Notification",
				Message: "Test message",
				IsRead:  pgtype.Bool{Bool: false, Valid: true},
			},
		}

		mockRepo.On("ListUnreadNotificationsByUser", mock.Anything, userID, int32(20), int32(0)).Return(unreadNotifications, nil)
		mockRepo.On("CountUnreadNotificationsByUser", mock.Anything, userID).Return(int64(1), nil)

		req := httptest.NewRequest(http.MethodGet, "/notifications/unread", nil)
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		assert.Contains(t, result, "notifications")
		assert.Equal(t, float64(1), result["unread_count"])

		mockRepo.AssertExpectations(t)
	})
}

func TestNotificationHandler_GetUnreadCount(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockRepo := new(MockNotificationRepository)
		app := setupNotificationTestApp(mockRepo)

		userID := uuid.MustParse("12345678-1234-1234-1234-123456789012")

		mockRepo.On("CountUnreadNotificationsByUser", mock.Anything, userID).Return(int64(5), nil)

		req := httptest.NewRequest(http.MethodGet, "/notifications/unread/count", nil)
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		assert.Equal(t, float64(5), result["unread_count"])

		mockRepo.AssertExpectations(t)
	})
}

func TestNotificationHandler_GetNotificationByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockRepo := new(MockNotificationRepository)
		app := setupNotificationTestApp(mockRepo)

		userID := uuid.MustParse("12345678-1234-1234-1234-123456789012")
		notificationID := uuid.New()

		notification := &db.Notification{
			ID:      utils.UUIDToPgtype(notificationID),
			UserID:  utils.UUIDToPgtype(userID),
			Type:    "TASK_ASSIGNED",
			Title:   "Test",
			Message: "Test message",
		}

		mockRepo.On("GetNotificationByID", mock.Anything, notificationID).Return(notification, nil)

		req := httptest.NewRequest(http.MethodGet, "/notifications/"+notificationID.String(), nil)
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		mockRepo.AssertExpectations(t)
	})

	t.Run("forbidden - different user", func(t *testing.T) {
		mockRepo := new(MockNotificationRepository)
		app := setupNotificationTestApp(mockRepo)

		differentUserID := uuid.New()
		notificationID := uuid.New()

		notification := &db.Notification{
			ID:      utils.UUIDToPgtype(notificationID),
			UserID:  utils.UUIDToPgtype(differentUserID), // Different user
			Type:    "TASK_ASSIGNED",
			Title:   "Test",
			Message: "Test message",
		}

		mockRepo.On("GetNotificationByID", mock.Anything, notificationID).Return(notification, nil)

		req := httptest.NewRequest(http.MethodGet, "/notifications/"+notificationID.String(), nil)
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)

		mockRepo.AssertExpectations(t)
	})
}

func TestNotificationHandler_MarkAsRead(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockRepo := new(MockNotificationRepository)
		app := setupNotificationTestApp(mockRepo)

		userID := uuid.MustParse("12345678-1234-1234-1234-123456789012")
		notificationID := uuid.New()

		notification := &db.Notification{
			ID:      utils.UUIDToPgtype(notificationID),
			UserID:  utils.UUIDToPgtype(userID),
			Type:    "TASK_ASSIGNED",
			Title:   "Test",
			Message: "Test message",
			IsRead:  pgtype.Bool{Bool: false, Valid: true},
		}

		updatedNotification := &db.Notification{
			ID:      utils.UUIDToPgtype(notificationID),
			UserID:  utils.UUIDToPgtype(userID),
			Type:    "TASK_ASSIGNED",
			Title:   "Test",
			Message: "Test message",
			IsRead:  pgtype.Bool{Bool: true, Valid: true},
		}

		mockRepo.On("GetNotificationByID", mock.Anything, notificationID).Return(notification, nil)
		mockRepo.On("MarkNotificationAsRead", mock.Anything, notificationID).Return(updatedNotification, nil)

		req := httptest.NewRequest(http.MethodPost, "/notifications/"+notificationID.String()+"/read", nil)
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		mockRepo.AssertExpectations(t)
	})
}

func TestNotificationHandler_MarkAllAsRead(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockRepo := new(MockNotificationRepository)
		app := setupNotificationTestApp(mockRepo)

		userID := uuid.MustParse("12345678-1234-1234-1234-123456789012")

		mockRepo.On("MarkAllNotificationsAsRead", mock.Anything, userID).Return(nil)

		req := httptest.NewRequest(http.MethodPost, "/notifications/read-all", nil)
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		assert.Equal(t, "all notifications marked as read", result["message"])

		mockRepo.AssertExpectations(t)
	})
}

func TestNotificationHandler_DeleteNotification(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockRepo := new(MockNotificationRepository)
		app := setupNotificationTestApp(mockRepo)

		userID := uuid.MustParse("12345678-1234-1234-1234-123456789012")
		notificationID := uuid.New()

		notification := &db.Notification{
			ID:      utils.UUIDToPgtype(notificationID),
			UserID:  utils.UUIDToPgtype(userID),
			Type:    "TASK_ASSIGNED",
			Title:   "Test",
			Message: "Test message",
		}

		mockRepo.On("GetNotificationByID", mock.Anything, notificationID).Return(notification, nil)
		mockRepo.On("DeleteNotification", mock.Anything, notificationID).Return(nil)

		req := httptest.NewRequest(http.MethodDelete, "/notifications/"+notificationID.String(), nil)
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		assert.Equal(t, "notification deleted successfully", result["message"])

		mockRepo.AssertExpectations(t)
	})
}
