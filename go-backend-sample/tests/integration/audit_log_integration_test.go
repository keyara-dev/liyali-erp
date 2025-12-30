package integration

import (
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

func TestAuditLogEndpoints(t *testing.T) {
	// Setup test database and repositories
	pool := setupTestDB(t)
	defer pool.Close()

	queries := db.New(pool)
	ctx := context.Background()

	// Initialize repositories
	userRepo := repository.NewUserRepository(queries)
	sessionRepo := repository.NewSessionRepository(queries)
	passwordResetRepo := repository.NewPasswordResetRepository(queries)
	auditLogRepo := repository.NewAuditLogRepository(queries)

	// Initialize services
	authService := services.NewAuthService(userRepo, sessionRepo, passwordResetRepo, "test-secret-key")

	// Initialize handlers
	auditLogHandler := handlers.NewAuditLogHandler(*auditLogRepo)
	authMiddleware := middleware.NewAuthMiddleware(authService)

	// Create test users - regular user and admin
	regularUser := createTestUser(t, queries, "regular@test.com", "USER")
	adminUser := createTestUser(t, queries, "admin@test.com", "ADMIN")

	regularToken := generateTestToken(t, authService, regularUser)
	adminToken := generateTestToken(t, authService, adminUser)

	regularUserID := utils.PgtypeToUUID(regularUser.ID)

	// Create test audit log
	auditLog, err := queries.CreateAuditLog(ctx, db.CreateAuditLogParams{
		UserID:       utils.UUIDToPgtype(regularUserID),
		Action:       "CREATE",
		ResourceType: "DOCUMENT",
		ResourceID:   utils.UUIDToPgtype(regularUserID),
		Changes:      []byte(`{"field": "status", "old": "DRAFT", "new": "SUBMITTED"}`),
		IpAddress:    pgtype.Text{String: "127.0.0.1", Valid: true},
		UserAgent:    pgtype.Text{String: "test-agent", Valid: true},
	})
	require.NoError(t, err)

	// Setup Fiber app
	app := fiber.New()
	api := app.Group("/api")

	auditLogs := api.Group("/audit-logs", authMiddleware.Authenticate)
	auditLogs.Get("/", auditLogHandler.GetAuditLogs)
	auditLogs.Get("/my", auditLogHandler.GetMyAuditLogs)
	auditLogs.Get("/:id", auditLogHandler.GetAuditLogByID)
	auditLogs.Get("/resource/:resource_type/:resource_id", auditLogHandler.GetAuditLogsByResource)

	t.Run("GET /api/audit-logs/my - success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/audit-logs/my", nil)
		req.Header.Set("Authorization", "Bearer "+regularToken)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Contains(t, result, "audit_logs")
		assert.Contains(t, result, "total")
		assert.Contains(t, result, "limit")
		assert.Contains(t, result, "offset")
	})

	t.Run("GET /api/audit-logs - admin access", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/audit-logs", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Contains(t, result, "audit_logs")
		assert.Contains(t, result, "total")
	})

	t.Run("GET /api/audit-logs - regular user forbidden", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/audit-logs", nil)
		req.Header.Set("Authorization", "Bearer "+regularToken)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Contains(t, result, "error")
		assert.Contains(t, result["error"], "insufficient permissions")
	})

	t.Run("GET /api/audit-logs/:id - admin success", func(t *testing.T) {
		auditLogID := utils.PgtypeToUUID(auditLog.ID)
		req := httptest.NewRequest(http.MethodGet, "/api/audit-logs/"+auditLogID.String(), nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("GET /api/audit-logs/:id - regular user forbidden", func(t *testing.T) {
		auditLogID := utils.PgtypeToUUID(auditLog.ID)
		req := httptest.NewRequest(http.MethodGet, "/api/audit-logs/"+auditLogID.String(), nil)
		req.Header.Set("Authorization", "Bearer "+regularToken)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("GET /api/audit-logs with filters", func(t *testing.T) {
		// Test with resource_type filter
		req := httptest.NewRequest(http.MethodGet, "/api/audit-logs?resource_type=DOCUMENT", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Test with action filter
		req = httptest.NewRequest(http.MethodGet, "/api/audit-logs?action=CREATE", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)

		resp, err = app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("GET /api/audit-logs/resource/:resource_type/:resource_id - admin success", func(t *testing.T) {
		resourceID := utils.PgtypeToUUID(auditLog.ResourceID)
		req := httptest.NewRequest(http.MethodGet, "/api/audit-logs/resource/DOCUMENT/"+resourceID.String(), nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Contains(t, result, "audit_logs")
		assert.Contains(t, result, "resource_type")
		assert.Contains(t, result, "resource_id")
	})

	t.Run("GET /api/audit-logs - unauthorized", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/audit-logs", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}
