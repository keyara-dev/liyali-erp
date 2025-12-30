package unit

import (
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

type MockAuditLogRepository struct {
	mock.Mock
}

func (m *MockAuditLogRepository) ListAuditLogs(ctx interface{}, limit, offset int32) ([]db.AuditLog, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]db.AuditLog), args.Error(1)
}

func (m *MockAuditLogRepository) ListAuditLogsByUser(ctx interface{}, userID uuid.UUID, limit, offset int32) ([]db.AuditLog, error) {
	args := m.Called(ctx, userID, limit, offset)
	return args.Get(0).([]db.AuditLog), args.Error(1)
}

func (m *MockAuditLogRepository) ListAuditLogsByResourceType(ctx interface{}, resourceType string, limit, offset int32) ([]db.AuditLog, error) {
	args := m.Called(ctx, resourceType, limit, offset)
	return args.Get(0).([]db.AuditLog), args.Error(1)
}

func (m *MockAuditLogRepository) ListAuditLogsByAction(ctx interface{}, action string, limit, offset int32) ([]db.AuditLog, error) {
	args := m.Called(ctx, action, limit, offset)
	return args.Get(0).([]db.AuditLog), args.Error(1)
}

func (m *MockAuditLogRepository) ListAuditLogsByResource(ctx interface{}, resourceType string, resourceID uuid.UUID, limit, offset int32) ([]db.AuditLog, error) {
	args := m.Called(ctx, resourceType, resourceID, limit, offset)
	return args.Get(0).([]db.AuditLog), args.Error(1)
}

func (m *MockAuditLogRepository) GetAuditLogByID(ctx interface{}, id uuid.UUID) (*db.AuditLog, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*db.AuditLog), args.Error(1)
}

func (m *MockAuditLogRepository) CountAuditLogs(ctx interface{}) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockAuditLogRepository) CountAuditLogsByUser(ctx interface{}, userID uuid.UUID) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockAuditLogRepository) CountAuditLogsByResource(ctx interface{}, resourceType string, resourceID uuid.UUID) (int64, error) {
	args := m.Called(ctx, resourceType, resourceID)
	return args.Get(0).(int64), args.Error(1)
}

func setupAuditLogTestApp(mockRepo *MockAuditLogRepository, userRole string) *fiber.App {
	app := fiber.New()
	handler := handlers.NewAuditLogHandler(*mockRepo)

	// Mock authentication by setting user context
	app.Use(func(c fiber.Ctx) error {
		c.Locals("userID", uuid.MustParse("12345678-1234-1234-1234-123456789012"))
		c.Locals("userEmail", "test@example.com")
		c.Locals("userRole", userRole)
		return c.Next()
	})

	app.Get("/audit-logs", handler.GetAuditLogs)
	app.Get("/audit-logs/my", handler.GetMyAuditLogs)
	app.Get("/audit-logs/:id", handler.GetAuditLogByID)
	app.Get("/audit-logs/resource/:resource_type/:resource_id", handler.GetAuditLogsByResource)

	return app
}

func TestAuditLogHandler_GetAuditLogs(t *testing.T) {
	t.Run("success - admin user", func(t *testing.T) {
		mockRepo := new(MockAuditLogRepository)
		app := setupAuditLogTestApp(mockRepo, "ADMIN")

		auditLogs := []db.AuditLog{
			{
				ID:           utils.UUIDToPgtype(uuid.New()),
				UserID:       utils.UUIDToPgtype(uuid.New()),
				Action:       "CREATE",
				ResourceType: "DOCUMENT",
				ResourceID:   utils.UUIDToPgtype(uuid.New()),
			},
		}

		mockRepo.On("ListAuditLogs", mock.Anything, int32(50), int32(0)).Return(auditLogs, nil)
		mockRepo.On("CountAuditLogs", mock.Anything).Return(int64(1), nil)

		req := httptest.NewRequest(http.MethodGet, "/audit-logs", nil)
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		assert.Contains(t, result, "audit_logs")
		assert.Contains(t, result, "total")

		mockRepo.AssertExpectations(t)
	})

	t.Run("success - manager user", func(t *testing.T) {
		mockRepo := new(MockAuditLogRepository)
		app := setupAuditLogTestApp(mockRepo, "MANAGER")

		mockRepo.On("ListAuditLogs", mock.Anything, int32(50), int32(0)).Return([]db.AuditLog{}, nil)
		mockRepo.On("CountAuditLogs", mock.Anything).Return(int64(0), nil)

		req := httptest.NewRequest(http.MethodGet, "/audit-logs", nil)
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		mockRepo.AssertExpectations(t)
	})

	t.Run("forbidden - regular user", func(t *testing.T) {
		mockRepo := new(MockAuditLogRepository)
		app := setupAuditLogTestApp(mockRepo, "USER")

		req := httptest.NewRequest(http.MethodGet, "/audit-logs", nil)
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		assert.Contains(t, result["error"], "insufficient permissions")
	})

	t.Run("filter by resource_type", func(t *testing.T) {
		mockRepo := new(MockAuditLogRepository)
		app := setupAuditLogTestApp(mockRepo, "ADMIN")

		mockRepo.On("ListAuditLogsByResourceType", mock.Anything, "DOCUMENT", int32(50), int32(0)).Return([]db.AuditLog{}, nil)
		mockRepo.On("CountAuditLogs", mock.Anything).Return(int64(0), nil)

		req := httptest.NewRequest(http.MethodGet, "/audit-logs?resource_type=DOCUMENT", nil)
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		mockRepo.AssertExpectations(t)
	})

	t.Run("filter by action", func(t *testing.T) {
		mockRepo := new(MockAuditLogRepository)
		app := setupAuditLogTestApp(mockRepo, "ADMIN")

		mockRepo.On("ListAuditLogsByAction", mock.Anything, "CREATE", int32(50), int32(0)).Return([]db.AuditLog{}, nil)
		mockRepo.On("CountAuditLogs", mock.Anything).Return(int64(0), nil)

		req := httptest.NewRequest(http.MethodGet, "/audit-logs?action=CREATE", nil)
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		mockRepo.AssertExpectations(t)
	})

	t.Run("filter by user_id", func(t *testing.T) {
		mockRepo := new(MockAuditLogRepository)
		app := setupAuditLogTestApp(mockRepo, "ADMIN")

		filterUserID := uuid.New()

		mockRepo.On("ListAuditLogsByUser", mock.Anything, filterUserID, int32(50), int32(0)).Return([]db.AuditLog{}, nil)
		mockRepo.On("CountAuditLogsByUser", mock.Anything, filterUserID).Return(int64(0), nil)

		req := httptest.NewRequest(http.MethodGet, "/audit-logs?user_id="+filterUserID.String(), nil)
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		mockRepo.AssertExpectations(t)
	})
}

func TestAuditLogHandler_GetMyAuditLogs(t *testing.T) {
	t.Run("success - any user can view their own logs", func(t *testing.T) {
		mockRepo := new(MockAuditLogRepository)
		app := setupAuditLogTestApp(mockRepo, "USER")

		userID := uuid.MustParse("12345678-1234-1234-1234-123456789012")

		auditLogs := []db.AuditLog{
			{
				ID:           utils.UUIDToPgtype(uuid.New()),
				UserID:       utils.UUIDToPgtype(userID),
				Action:       "CREATE",
				ResourceType: "DOCUMENT",
			},
		}

		mockRepo.On("ListAuditLogsByUser", mock.Anything, userID, int32(50), int32(0)).Return(auditLogs, nil)
		mockRepo.On("CountAuditLogsByUser", mock.Anything, userID).Return(int64(1), nil)

		req := httptest.NewRequest(http.MethodGet, "/audit-logs/my", nil)
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		assert.Contains(t, result, "audit_logs")
		assert.Equal(t, float64(1), result["total"])

		mockRepo.AssertExpectations(t)
	})

	t.Run("with pagination", func(t *testing.T) {
		mockRepo := new(MockAuditLogRepository)
		app := setupAuditLogTestApp(mockRepo, "USER")

		userID := uuid.MustParse("12345678-1234-1234-1234-123456789012")

		mockRepo.On("ListAuditLogsByUser", mock.Anything, userID, int32(10), int32(5)).Return([]db.AuditLog{}, nil)
		mockRepo.On("CountAuditLogsByUser", mock.Anything, userID).Return(int64(0), nil)

		req := httptest.NewRequest(http.MethodGet, "/audit-logs/my?limit=10&offset=5", nil)
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

func TestAuditLogHandler_GetAuditLogByID(t *testing.T) {
	t.Run("success - admin", func(t *testing.T) {
		mockRepo := new(MockAuditLogRepository)
		app := setupAuditLogTestApp(mockRepo, "ADMIN")

		auditLogID := uuid.New()
		auditLog := &db.AuditLog{
			ID:           utils.UUIDToPgtype(auditLogID),
			UserID:       utils.UUIDToPgtype(uuid.New()),
			Action:       "CREATE",
			ResourceType: "DOCUMENT",
			ResourceID:   utils.UUIDToPgtype(uuid.New()),
		}

		mockRepo.On("GetAuditLogByID", mock.Anything, auditLogID).Return(auditLog, nil)

		req := httptest.NewRequest(http.MethodGet, "/audit-logs/"+auditLogID.String(), nil)
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		mockRepo.AssertExpectations(t)
	})

	t.Run("forbidden - regular user", func(t *testing.T) {
		mockRepo := new(MockAuditLogRepository)
		app := setupAuditLogTestApp(mockRepo, "USER")

		auditLogID := uuid.New()

		req := httptest.NewRequest(http.MethodGet, "/audit-logs/"+auditLogID.String(), nil)
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})
}

func TestAuditLogHandler_GetAuditLogsByResource(t *testing.T) {
	t.Run("success - admin", func(t *testing.T) {
		mockRepo := new(MockAuditLogRepository)
		app := setupAuditLogTestApp(mockRepo, "ADMIN")

		resourceID := uuid.New()
		resourceType := "DOCUMENT"

		auditLogs := []db.AuditLog{
			{
				ID:           utils.UUIDToPgtype(uuid.New()),
				UserID:       utils.UUIDToPgtype(uuid.New()),
				Action:       "UPDATE",
				ResourceType: resourceType,
				ResourceID:   utils.UUIDToPgtype(resourceID),
			},
		}

		mockRepo.On("ListAuditLogsByResource", mock.Anything, resourceType, resourceID, int32(50), int32(0)).Return(auditLogs, nil)
		mockRepo.On("CountAuditLogsByResource", mock.Anything, resourceType, resourceID).Return(int64(1), nil)

		req := httptest.NewRequest(http.MethodGet, "/audit-logs/resource/"+resourceType+"/"+resourceID.String(), nil)
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		assert.Contains(t, result, "audit_logs")
		assert.Equal(t, resourceType, result["resource_type"])
		assert.Equal(t, resourceID.String(), result["resource_id"])

		mockRepo.AssertExpectations(t)
	})

	t.Run("forbidden - regular user", func(t *testing.T) {
		mockRepo := new(MockAuditLogRepository)
		app := setupAuditLogTestApp(mockRepo, "USER")

		resourceID := uuid.New()

		req := httptest.NewRequest(http.MethodGet, "/audit-logs/resource/DOCUMENT/"+resourceID.String(), nil)
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})
}
