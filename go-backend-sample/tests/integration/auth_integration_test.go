package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cozyCodr/liyali-gateway/internal/config"
	"github.com/cozyCodr/liyali-gateway/internal/db"
	"github.com/cozyCodr/liyali-gateway/internal/handlers"
	"github.com/cozyCodr/liyali-gateway/internal/middleware"
	"github.com/cozyCodr/liyali-gateway/internal/repository"
	"github.com/cozyCodr/liyali-gateway/internal/services"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testDB       *pgxpool.Pool
	testApp      *fiber.App
	testJWTSecret = "test-jwt-secret-for-integration-tests"
)

func setupTestApp(t *testing.T) *fiber.App {
	// Load config
	cfg, err := config.Load()
	require.NoError(t, err)

	// Connect to test database
	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	require.NoError(t, err)
	testDB = pool

	// Initialize repositories and services
	queries := db.New(pool)
	userRepo := repository.NewUserRepository(queries)
	sessionRepo := repository.NewSessionRepository(queries)
	passwordResetRepo := repository.NewPasswordResetRepository(queries)
	authService := services.NewAuthService(userRepo, sessionRepo, passwordResetRepo, testJWTSecret)

	// Initialize middleware and handlers
	authMiddleware := middleware.NewAuthMiddleware(authService)
	authHandler := handlers.NewAuthHandler(authService)

	// Create Fiber app
	app := fiber.New()

	// Auth routes
	auth := app.Group("/api/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Post("/refresh", authHandler.RefreshToken)
	auth.Post("/logout", authHandler.Logout)
	auth.Post("/password-reset/request", authHandler.RequestPasswordReset)
	auth.Post("/password-reset/confirm", authHandler.ResetPassword)
	auth.Post("/change-password", authMiddleware.Authenticate, authHandler.ChangePassword)
	auth.Get("/me", authMiddleware.Authenticate, authHandler.GetCurrentUser)

	testApp = app
	return app
}

func teardownTestApp() {
	if testDB != nil {
		testDB.Close()
	}
}

func TestRegisterEndpoint(t *testing.T) {
	app := setupTestApp(t)
	defer teardownTestApp()

	// Prepare request
	payload := map[string]interface{}{
		"email":      fmt.Sprintf("test-%d@example.com", time.Now().Unix()),
		"password":   "Test@1234",
		"name":       "Test User",
		"role":       "REQUESTER",
		"department": "Engineering",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	resp, err := app.Test(req)
	require.NoError(t, err)

	// Assert
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var response map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&response)
	assert.Equal(t, "User registered successfully", response["message"])
	assert.NotNil(t, response["user"])
}

func TestRegisterEndpoint_DuplicateEmail(t *testing.T) {
	app := setupTestApp(t)
	defer teardownTestApp()

	email := fmt.Sprintf("duplicate-%d@example.com", time.Now().Unix())

	// Register first user
	payload := map[string]interface{}{
		"email":      email,
		"password":   "Test@1234",
		"name":       "Test User 1",
		"role":       "REQUESTER",
		"department": "Engineering",
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// Try to register with same email
	req = httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)

	// Assert
	assert.Equal(t, http.StatusConflict, resp.StatusCode)

	var response map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&response)
	assert.Equal(t, "Email already exists", response["error"])
}

func TestLoginEndpoint_Success(t *testing.T) {
	app := setupTestApp(t)
	defer teardownTestApp()

	email := fmt.Sprintf("login-test-%d@example.com", time.Now().Unix())
	password := "Test@1234"

	// First, register a user
	registerPayload := map[string]interface{}{
		"email":      email,
		"password":   password,
		"name":       "Test User",
		"role":       "REQUESTER",
		"department": "Engineering",
	}
	body, _ := json.Marshal(registerPayload)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	_, _ = app.Test(req)

	// Now login
	loginPayload := map[string]interface{}{
		"email":    email,
		"password": password,
	}
	body, _ = json.Marshal(loginPayload)
	req = httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	resp, err := app.Test(req)
	require.NoError(t, err)

	// Assert
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&response)
	assert.NotEmpty(t, response["access_token"])
	assert.NotEmpty(t, response["refresh_token"])
	assert.NotNil(t, response["user"])
}

func TestLoginEndpoint_InvalidCredentials(t *testing.T) {
	app := setupTestApp(t)
	defer teardownTestApp()

	loginPayload := map[string]interface{}{
		"email":    "nonexistent@example.com",
		"password": "wrongpassword",
	}
	body, _ := json.Marshal(loginPayload)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	resp, err := app.Test(req)
	require.NoError(t, err)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	var response map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&response)
	assert.Equal(t, "Invalid email or password", response["error"])
}

func TestRefreshTokenEndpoint_Success(t *testing.T) {
	app := setupTestApp(t)
	defer teardownTestApp()

	email := fmt.Sprintf("refresh-test-%d@example.com", time.Now().Unix())
	password := "Test@1234"

	// Register and login
	registerPayload := map[string]interface{}{
		"email":      email,
		"password":   password,
		"name":       "Test User",
		"role":       "REQUESTER",
		"department": "Engineering",
	}
	body, _ := json.Marshal(registerPayload)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	_, _ = app.Test(req)

	loginPayload := map[string]interface{}{
		"email":    email,
		"password": password,
	}
	body, _ = json.Marshal(loginPayload)
	req = httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	var loginResponse map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&loginResponse)
	refreshToken := loginResponse["refresh_token"].(string)

	// Use refresh token
	refreshPayload := map[string]interface{}{
		"refresh_token": refreshToken,
	}
	body, _ = json.Marshal(refreshPayload)
	req = httptest.NewRequest(http.MethodPost, "/api/auth/refresh", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	resp, err := app.Test(req)
	require.NoError(t, err)

	// Assert
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&response)
	assert.NotEmpty(t, response["access_token"])
}

func TestGetCurrentUserEndpoint_Authenticated(t *testing.T) {
	app := setupTestApp(t)
	defer teardownTestApp()

	email := fmt.Sprintf("me-test-%d@example.com", time.Now().Unix())
	password := "Test@1234"

	// Register and login
	registerPayload := map[string]interface{}{
		"email":      email,
		"password":   password,
		"name":       "Test User",
		"role":       "ADMIN",
		"department": "Engineering",
	}
	body, _ := json.Marshal(registerPayload)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	_, _ = app.Test(req)

	loginPayload := map[string]interface{}{
		"email":    email,
		"password": password,
	}
	body, _ = json.Marshal(loginPayload)
	req = httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	var loginResponse map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&loginResponse)
	accessToken := loginResponse["access_token"].(string)

	// Get current user
	req = httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	// Execute
	resp, err := app.Test(req)
	require.NoError(t, err)

	// Assert
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&response)
	assert.Equal(t, email, response["email"])
	assert.Equal(t, "ADMIN", response["role"])
}

func TestGetCurrentUserEndpoint_Unauthenticated(t *testing.T) {
	app := setupTestApp(t)
	defer teardownTestApp()

	// Try to get current user without token
	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)

	// Execute
	resp, err := app.Test(req)
	require.NoError(t, err)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestLogoutEndpoint_Success(t *testing.T) {
	app := setupTestApp(t)
	defer teardownTestApp()

	email := fmt.Sprintf("logout-test-%d@example.com", time.Now().Unix())
	password := "Test@1234"

	// Register and login
	registerPayload := map[string]interface{}{
		"email":      email,
		"password":   password,
		"name":       "Test User",
		"role":       "REQUESTER",
		"department": "Engineering",
	}
	body, _ := json.Marshal(registerPayload)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	_, _ = app.Test(req)

	loginPayload := map[string]interface{}{
		"email":    email,
		"password": password,
	}
	body, _ = json.Marshal(loginPayload)
	req = httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	var loginResponse map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&loginResponse)
	refreshToken := loginResponse["refresh_token"].(string)

	// Logout
	logoutPayload := map[string]interface{}{
		"refresh_token": refreshToken,
	}
	body, _ = json.Marshal(logoutPayload)
	req = httptest.NewRequest(http.MethodPost, "/api/auth/logout", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	resp, err := app.Test(req)
	require.NoError(t, err)

	// Assert
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&response)
	assert.Equal(t, "Logged out successfully", response["message"])
}
