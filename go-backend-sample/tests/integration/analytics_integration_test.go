package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cozyCodr/liyali-gateway/internal/db"
	"github.com/cozyCodr/liyali-gateway/internal/handlers"
	"github.com/cozyCodr/liyali-gateway/internal/middleware"
	"github.com/cozyCodr/liyali-gateway/internal/repository"
	"github.com/cozyCodr/liyali-gateway/internal/services"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnalyticsEndpoints(t *testing.T) {
	// Setup test database and repositories
	pool := setupTestDB(t)
	defer pool.Close()

	queries := db.New(pool)

	// Initialize repositories
	userRepo := repository.NewUserRepository(queries)
	sessionRepo := repository.NewSessionRepository(queries)
	passwordResetRepo := repository.NewPasswordResetRepository(queries)
	documentRepo := repository.NewDocumentRepository(queries)
	approvalTaskRepo := repository.NewApprovalTaskRepository(queries)
	approvalHistoryRepo := repository.NewApprovalHistoryRepository(queries)
	workflowRepo := repository.NewWorkflowRepository(queries)

	// Initialize services
	authService := services.NewAuthService(userRepo, sessionRepo, passwordResetRepo, "test-secret-key")
	analyticsService := services.NewAnalyticsService(*documentRepo, *approvalTaskRepo, *approvalHistoryRepo, *workflowRepo)

	// Initialize handlers
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsService)
	authMiddleware := middleware.NewAuthMiddleware(authService)

	// Create test user
	testUser := createTestUser(t, queries, "analytics@test.com", "MANAGER")
	token := generateTestToken(t, authService, testUser)

	// Setup Fiber app
	app := fiber.New()
	api := app.Group("/api")

	analytics := api.Group("/analytics", authMiddleware.Authenticate)
	analytics.Get("/metrics", analyticsHandler.GetDashboardMetrics)
	analytics.Get("/trends", analyticsHandler.GetTrendData)
	analytics.Get("/bottlenecks", analyticsHandler.GetBottlenecks)

	t.Run("GET /api/analytics/metrics - success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/analytics/metrics", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Contains(t, result, "total_documents")
		assert.Contains(t, result, "documents_by_status")
		assert.Contains(t, result, "pending_approvals")
		assert.Contains(t, result, "active_workflows")
	})

	t.Run("GET /api/analytics/metrics - unauthorized", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/analytics/metrics", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("GET /api/analytics/trends - success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/analytics/trends?days=7", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Contains(t, result, "trends")
		assert.Contains(t, result, "days")
		assert.Equal(t, float64(7), result["days"])
	})

	t.Run("GET /api/analytics/trends - custom days", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/analytics/trends?days=30", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, float64(30), result["days"])
	})

	t.Run("GET /api/analytics/bottlenecks - success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/analytics/bottlenecks", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Contains(t, result, "bottlenecks")
		assert.Contains(t, result, "count")
	})
}

// Helper function to create test user
func createTestUser(t *testing.T, queries *db.Queries, email, role string) *db.User {
	user, err := queries.CreateUser(context.Background(), db.CreateUserParams{
		Email:        email,
		PasswordHash: "$2a$10$testhashedpassword",
		Name:         "Test User",
		Role:         role,
	})
	require.NoError(t, err)
	return &user
}

// Helper function to generate test JWT token
func generateTestToken(t *testing.T, authService *services.AuthService, user *db.User) string {
	userID := uuid.MustParse(user.ID.String())
	token, _, err := authService.GenerateTokens(userID, user.Email, user.Role)
	require.NoError(t, err)
	return token
}
