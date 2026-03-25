package handlers

// auth_service_mock_test.go — service-error path tests for auth_handler.go.
//
// Tests in this file create real *services.AuthService / *services.SessionService
// instances backed by lightweight mock repositories so we can exercise the
// handler branches that fire after the service call returns an error.

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/google/uuid"
	sqlc "github.com/liyali/liyali-gateway/database/sqlc"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/repository"
	"github.com/liyali/liyali-gateway/services"
	"github.com/stretchr/testify/assert"
)

// ─────────────────────────────────────────────────────────────────────────────
// Mock repository implementations
// ─────────────────────────────────────────────────────────────────────────────

// mockUserRepo — UserRepositoryInterface stub.
// GetByEmail / GetByID return errors by default (overridable via function fields).
type mockUserRepo struct {
	getByEmailFn func(ctx context.Context, email string) (*models.User, error)
	getByIDFn    func(ctx context.Context, id string) (*models.User, error)
}

func (m *mockUserRepo) Create(ctx context.Context, user *models.User) (*models.User, error) {
	return nil, fmt.Errorf("mock: create not implemented")
}
func (m *mockUserRepo) GetByID(ctx context.Context, id string) (*models.User, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, fmt.Errorf("not found")
}
func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	if m.getByEmailFn != nil {
		return m.getByEmailFn(ctx, email)
	}
	return nil, fmt.Errorf("user not found")
}
func (m *mockUserRepo) Update(ctx context.Context, user *models.User) (*models.User, error) {
	return nil, fmt.Errorf("mock: update not implemented")
}
func (m *mockUserRepo) UpdatePassword(ctx context.Context, id string, hashedPassword string) error {
	return nil
}
func (m *mockUserRepo) UpdateLastLogin(ctx context.Context, id string) error { return nil }
func (m *mockUserRepo) Delete(ctx context.Context, id string) error          { return nil }
func (m *mockUserRepo) List(ctx context.Context, limit, offset int) ([]*models.User, error) {
	return nil, nil
}
func (m *mockUserRepo) ListByOrganization(ctx context.Context, organizationID string, limit, offset int) ([]*models.User, error) {
	return nil, nil
}
func (m *mockUserRepo) Count(ctx context.Context) (int64, error)       { return 0, nil }
func (m *mockUserRepo) CountActive(ctx context.Context) (int64, error) { return 0, nil }
func (m *mockUserRepo) Activate(ctx context.Context, id string) error  { return nil }
func (m *mockUserRepo) Deactivate(ctx context.Context, id string) error { return nil }

// Ensure interface satisfaction at compile time.
var _ repository.UserRepositoryInterface = (*mockUserRepo)(nil)

// ─────────────────────────────────────────────────────────────────────────────

// mockSessionRepo — SessionRepositoryInterface stub.
type mockSessionRepo struct {
	deleteByUserIDFn func(ctx context.Context, userID string) error
	getByUserIDFn    func(ctx context.Context, userID string) ([]*sqlc.Session, error)
}

func (m *mockSessionRepo) Create(ctx context.Context, userID, refreshToken, ipAddress, userAgent string, expiresAt time.Time) (*sqlc.Session, error) {
	return nil, fmt.Errorf("mock: session create not implemented")
}
func (m *mockSessionRepo) GetByRefreshToken(ctx context.Context, refreshToken string) (*sqlc.Session, error) {
	return nil, fmt.Errorf("mock: not found")
}
func (m *mockSessionRepo) GetByUserID(ctx context.Context, userID string) ([]*sqlc.Session, error) {
	if m.getByUserIDFn != nil {
		return m.getByUserIDFn(ctx, userID)
	}
	return nil, fmt.Errorf("mock: db error")
}
func (m *mockSessionRepo) UpdateRefreshToken(ctx context.Context, id uuid.UUID, oldRefreshToken, newRefreshToken string, expiresAt time.Time) (int64, error) {
	return 0, fmt.Errorf("mock: not implemented")
}
func (m *mockSessionRepo) Delete(ctx context.Context, id uuid.UUID) error { return nil }
func (m *mockSessionRepo) DeleteByRefreshToken(ctx context.Context, refreshToken string) error {
	return nil
}
func (m *mockSessionRepo) DeleteByUserID(ctx context.Context, userID string) error {
	if m.deleteByUserIDFn != nil {
		return m.deleteByUserIDFn(ctx, userID)
	}
	return nil
}
func (m *mockSessionRepo) DeleteExpired(ctx context.Context) error           { return nil }
func (m *mockSessionRepo) CountActive(ctx context.Context) (int64, error)   { return 0, nil }
func (m *mockSessionRepo) CountUserActive(ctx context.Context, userID string) (int64, error) {
	return 0, nil
}

var _ repository.SessionRepositoryInterface = (*mockSessionRepo)(nil)

// ─────────────────────────────────────────────────────────────────────────────

// mockPasswordResetRepo — PasswordResetRepositoryInterface stub.
type mockPasswordResetRepo struct{}

func (m *mockPasswordResetRepo) Create(ctx context.Context, userID, token string, expiresAt time.Time) (*sqlc.PasswordReset, error) {
	return nil, fmt.Errorf("mock: not implemented")
}
func (m *mockPasswordResetRepo) GetByToken(ctx context.Context, token string) (*sqlc.PasswordReset, error) {
	return nil, fmt.Errorf("mock: not found")
}
func (m *mockPasswordResetRepo) MarkAsUsed(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (m *mockPasswordResetRepo) DeleteByUserID(ctx context.Context, userID string) error {
	return nil
}
func (m *mockPasswordResetRepo) DeleteExpired(ctx context.Context) error { return nil }
func (m *mockPasswordResetRepo) DeleteUsed(ctx context.Context) error    { return nil }

var _ repository.PasswordResetRepositoryInterface = (*mockPasswordResetRepo)(nil)

// ─────────────────────────────────────────────────────────────────────────────

// mockLoginAttemptRepo — LoginAttemptRepositoryInterface stub.
// GetRecentFailedAttempts returns 0, nil by default (no recent failures).
type mockLoginAttemptRepo struct{}

func (m *mockLoginAttemptRepo) Create(ctx context.Context, userID, email, ipAddress, userAgent string, success bool, failureReason string) (*sqlc.LoginAttempt, error) {
	return nil, nil
}
func (m *mockLoginAttemptRepo) GetRecentFailedAttempts(ctx context.Context, email string, since time.Time) (int64, error) {
	return 0, nil
}
func (m *mockLoginAttemptRepo) GetRecentFailedAttemptsByIP(ctx context.Context, ipAddress string, since time.Time) (int64, error) {
	return 0, nil
}
func (m *mockLoginAttemptRepo) GetByUser(ctx context.Context, userID string, limit, offset int) ([]*sqlc.LoginAttempt, error) {
	return nil, nil
}
func (m *mockLoginAttemptRepo) GetByEmail(ctx context.Context, email string, limit, offset int) ([]*sqlc.LoginAttempt, error) {
	return nil, nil
}
func (m *mockLoginAttemptRepo) DeleteOld(ctx context.Context, before time.Time) error { return nil }

var _ repository.LoginAttemptRepositoryInterface = (*mockLoginAttemptRepo)(nil)

// ─────────────────────────────────────────────────────────────────────────────

// mockLockoutRepo — AccountLockoutRepositoryInterface stub.
// GetActiveByEmail / GetActiveByUserID return nil, nil (no lockout).
type mockLockoutRepo struct{}

func (m *mockLockoutRepo) Create(ctx context.Context, userID, email, ipAddress, reason string, unlocksAt time.Time) (*sqlc.AccountLockout, error) {
	return nil, nil
}
func (m *mockLockoutRepo) GetActiveByUserID(ctx context.Context, userID string) (*sqlc.AccountLockout, error) {
	return nil, nil
}
func (m *mockLockoutRepo) GetActiveByEmail(ctx context.Context, email string) (*sqlc.AccountLockout, error) {
	return nil, nil
}
func (m *mockLockoutRepo) Unlock(ctx context.Context, userID string) error          { return nil }
func (m *mockLockoutRepo) UnlockByEmail(ctx context.Context, email string) error    { return nil }
func (m *mockLockoutRepo) GetHistory(ctx context.Context, userID string, limit, offset int) ([]*sqlc.AccountLockout, error) {
	return nil, nil
}
func (m *mockLockoutRepo) CleanupExpired(ctx context.Context) error { return nil }

var _ repository.AccountLockoutRepositoryInterface = (*mockLockoutRepo)(nil)

// ─────────────────────────────────────────────────────────────────────────────
// Helper: build a real AuthService wired to mock repos
// ─────────────────────────────────────────────────────────────────────────────

// newMockAuthService returns an *services.AuthService backed entirely by mock
// repositories.  The userRepo and sessionRepo can be customised per test.
func newMockAuthService(userRepo repository.UserRepositoryInterface, sessionRepo repository.SessionRepositoryInterface) *services.AuthService {
	return services.NewAuthService(
		userRepo,
		sessionRepo,
		&mockPasswordResetRepo{},
		&mockLoginAttemptRepo{},
		&mockLockoutRepo{},
		nil,   // auditService — nil is safe
		"test-jwt-secret",
		nil,   // gorm.DB — nil; service path under test doesn't reach db
	)
}

// ─────────────────────────────────────────────────────────────────────────────
// Helper: Fiber app with a real AuthService (all routes)
// ─────────────────────────────────────────────────────────────────────────────

func newAuthAppWithMockService(authSvc *services.AuthService, sessionSvc *services.SessionService, mids ...fiber.Handler) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   err.Error(),
			})
		},
	})
	app.Use(recover.New())

	h := &AuthHandler{
		authService:     authSvc,
		sessionService:  sessionSvc,
		activityService: nil,
		rbacService:     nil,
		validate:        validator.New(),
	}

	auth := app.Group("/auth")
	for _, mw := range mids {
		auth.Use(mw)
	}

	auth.Post("/login", h.Login)
	auth.Post("/admin-login", h.AdminLogin)
	auth.Get("/profile", h.GetProfile)
	auth.Put("/profile", h.UpdateProfile)
	auth.Post("/logout-all", h.LogoutAll)
	auth.Post("/register", h.Register)
	auth.Get("/activity", h.GetUserActivity)
	auth.Get("/sessions", h.GetUserSessions)
	auth.Delete("/sessions/:id", h.TerminateSession)

	return app
}

// ─────────────────────────────────────────────────────────────────────────────
// Login — service error path
// ─────────────────────────────────────────────────────────────────────────────

// TestLogin_AuthServiceFails sends a valid login request; the mock userRepo
// returns an error from GetByEmail → AuthService.Login returns
// ErrInvalidCredentials → handler responds 401.
func TestLogin_AuthServiceFails(t *testing.T) {
	userRepo := &mockUserRepo{
		getByEmailFn: func(_ context.Context, _ string) (*models.User, error) {
			return nil, fmt.Errorf("user not found")
		},
	}
	svc := newMockAuthService(userRepo, &mockSessionRepo{})
	app := newAuthAppWithMockService(svc, nil)

	resp := testRequest(app, http.MethodPost, "/auth/login", map[string]interface{}{
		"email":    "user@example.com",
		"password": "password123",
	})

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

// TestLogin_WithActivityServiceNil ensures the activityService nil branch is
// hit (false path of `if h.activityService != nil`) when login fails.
func TestLogin_WithActivityServiceNil(t *testing.T) {
	userRepo := &mockUserRepo{
		getByEmailFn: func(_ context.Context, _ string) (*models.User, error) {
			return nil, fmt.Errorf("user not found")
		},
	}
	svc := newMockAuthService(userRepo, &mockSessionRepo{})
	// activityService is nil in newAuthAppWithMockService by default
	app := newAuthAppWithMockService(svc, nil)

	resp := testRequest(app, http.MethodPost, "/auth/login", map[string]interface{}{
		"email":    "nouser@example.com",
		"password": "somepassword",
	})

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// AdminLogin — service error path
// ─────────────────────────────────────────────────────────────────────────────

// TestAdminLogin_ServiceFails sends a valid body; mock returns error → 401.
func TestAdminLogin_ServiceFails(t *testing.T) {
	userRepo := &mockUserRepo{
		getByEmailFn: func(_ context.Context, _ string) (*models.User, error) {
			return nil, fmt.Errorf("user not found")
		},
	}
	svc := newMockAuthService(userRepo, &mockSessionRepo{})
	app := newAuthAppWithMockService(svc, nil)

	resp := testRequest(app, http.MethodPost, "/auth/admin-login", map[string]interface{}{
		"email":    "admin@example.com",
		"password": "adminpass123",
	})

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

// ─────────────────────────────────────────────────────────────────────────────
// GetProfile — service error path
// ─────────────────────────────────────────────────────────────────────────────

// TestGetProfile_ServiceFails — authenticated request but GetProfileByID fails
// (GetByID returns error) → 404.
func TestGetProfile_ServiceFails(t *testing.T) {
	userRepo := &mockUserRepo{
		getByIDFn: func(_ context.Context, _ string) (*models.User, error) {
			return nil, fmt.Errorf("not found")
		},
	}
	svc := newMockAuthService(userRepo, &mockSessionRepo{})
	app := newAuthAppWithMockService(svc, nil, withUserID(testUserID))

	resp := testRequest(app, http.MethodGet, "/auth/profile", nil)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

// TestGetProfile_ServiceFailsWithOrgID — with organizationID also set in locals
// to exercise the org-role-loading branch (which requires config.DB and will
// also hit the not-found path from GetByID returning error → 404).
func TestGetProfile_ServiceFailsWithOrgID(t *testing.T) {
	userRepo := &mockUserRepo{
		getByIDFn: func(_ context.Context, _ string) (*models.User, error) {
			return nil, fmt.Errorf("not found")
		},
	}
	svc := newMockAuthService(userRepo, &mockSessionRepo{})

	withBothLocals := func(c *fiber.Ctx) error {
		c.Locals("userID", testUserID)
		c.Locals("organizationID", testOrgID)
		return c.Next()
	}
	app := newAuthAppWithMockService(svc, nil, withBothLocals)

	resp := testRequest(app, http.MethodGet, "/auth/profile", nil)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// LogoutAll — service success and failure paths
// ─────────────────────────────────────────────────────────────────────────────

// TestLogoutAll_WithService — sessionRepo.DeleteByUserID returns nil → 200.
func TestLogoutAll_WithService(t *testing.T) {
	sessionRepo := &mockSessionRepo{
		deleteByUserIDFn: func(_ context.Context, _ string) error {
			return nil // success
		},
	}
	svc := newMockAuthService(&mockUserRepo{}, sessionRepo)
	app := newAuthAppWithMockService(svc, nil, withUserID(testUserID))

	resp := testRequest(app, http.MethodPost, "/auth/logout-all", nil)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeResponse(resp)
	assert.Equal(t, true, body["success"])
}

// TestLogoutAll_ServiceFails — sessionRepo.DeleteByUserID returns error → 500.
func TestLogoutAll_ServiceFails(t *testing.T) {
	sessionRepo := &mockSessionRepo{
		deleteByUserIDFn: func(_ context.Context, _ string) error {
			return fmt.Errorf("db error")
		},
	}
	svc := newMockAuthService(&mockUserRepo{}, sessionRepo)
	app := newAuthAppWithMockService(svc, nil, withUserID(testUserID))

	resp := testRequest(app, http.MethodPost, "/auth/logout-all", nil)

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// Register — service error paths
// ─────────────────────────────────────────────────────────────────────────────

// validRegisterBody returns a minimal valid register request body.
func validRegisterBody() map[string]interface{} {
	return map[string]interface{}{
		"email":    "newuser@example.com",
		"password": "StrongPass1",
		"name":     "New User",
		"role":     "requester",
	}
}

// TestRegister_EmailAlreadyExists — userRepo.GetByEmail returns an existing user
// (no error) → AuthService.Register returns ErrEmailAlreadyExists → handler
// responds 409.
func TestRegister_EmailAlreadyExists(t *testing.T) {
	userRepo := &mockUserRepo{
		getByEmailFn: func(_ context.Context, _ string) (*models.User, error) {
			// Return a non-nil user with no error → "email already exists"
			return &models.User{ID: "existing-id", Email: "newuser@example.com"}, nil
		},
	}
	svc := newMockAuthService(userRepo, &mockSessionRepo{})
	app := newAuthAppWithMockService(svc, nil)

	resp := testRequest(app, http.MethodPost, "/auth/register", validRegisterBody())

	assert.Equal(t, http.StatusConflict, resp.StatusCode)
	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

// TestRegister_ServiceFails — userRepo.GetByEmail returns error (email not
// found) but Create also fails → the create step returns an error → 500.
func TestRegister_ServiceFails(t *testing.T) {
	userRepo := &mockUserRepo{
		getByEmailFn: func(_ context.Context, _ string) (*models.User, error) {
			// Email not found → proceed with registration
			return nil, fmt.Errorf("not found")
		},
		// Create is not overridden — default returns error from mockUserRepo.Create
	}
	svc := newMockAuthService(userRepo, &mockSessionRepo{})
	app := newAuthAppWithMockService(svc, nil)

	resp := testRequest(app, http.MethodPost, "/auth/register", validRegisterBody())

	// Create fails → Register returns error → handler sends 500
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// UpdateProfile — service error path
// ─────────────────────────────────────────────────────────────────────────────

// TestUpdateProfile_ServiceFails — GetByID returns error →
// UpdateProfile returns error → handler responds 500.
func TestUpdateProfile_ServiceFails(t *testing.T) {
	userRepo := &mockUserRepo{
		getByIDFn: func(_ context.Context, _ string) (*models.User, error) {
			return nil, fmt.Errorf("not found")
		},
	}
	svc := newMockAuthService(userRepo, &mockSessionRepo{})
	app := newAuthAppWithMockService(svc, nil, withUserID(testUserID))

	resp := testRequest(app, http.MethodPut, "/auth/profile", map[string]interface{}{
		"name":  "Updated Name",
		"email": "updated@example.com",
	})

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// GetUserSessions — session service error path
// ─────────────────────────────────────────────────────────────────────────────

// TestGetUserSessions_WithService — sessionRepo.GetByUserID returns error
// → SessionService.GetUserSessions fails → handler responds 500.
func TestGetUserSessions_WithService(t *testing.T) {
	sessionRepo := &mockSessionRepo{
		getByUserIDFn: func(_ context.Context, _ string) ([]*sqlc.Session, error) {
			return nil, fmt.Errorf("db error")
		},
	}
	sessionSvc := services.NewSessionService(sessionRepo)
	svc := newMockAuthService(&mockUserRepo{}, sessionRepo)
	app := newAuthAppWithMockService(svc, sessionSvc, withUserID(testUserID))

	resp := testRequest(app, http.MethodGet, "/auth/sessions", nil)

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// TerminateSession — session service error path
// ─────────────────────────────────────────────────────────────────────────────

// TestTerminateSession_WithService_Error — sessionRepo.GetByUserID returns an
// error → SessionService.TerminateSession fails → handler responds 500.
func TestTerminateSession_WithService_Error(t *testing.T) {
	sessionRepo := &mockSessionRepo{
		getByUserIDFn: func(_ context.Context, _ string) ([]*sqlc.Session, error) {
			return nil, fmt.Errorf("db error")
		},
	}
	sessionSvc := services.NewSessionService(sessionRepo)
	svc := newMockAuthService(&mockUserRepo{}, sessionRepo)
	app := newAuthAppWithMockService(svc, sessionSvc, withUserID(testUserID))

	// Use a valid UUID as the session ID so uuid.Parse doesn't fail first
	sessionID := uuid.New().String()
	resp := testRequest(app, http.MethodDelete, "/auth/sessions/"+sessionID, nil)

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// GetUserActivity — activity service error path via SQLite-backed repo
// ─────────────────────────────────────────────────────────────────────────────

// TestGetUserActivity_ServiceError — an ActivityService backed by a real
// repository whose underlying table does not exist returns an error →
// handler responds 500.
func TestGetUserActivity_ServiceError(t *testing.T) {
	// Use a fresh in-memory SQLite DB without any migrations so queries fail.
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Drop the user_activity_logs table if AutoMigrate created it, so queries fail.
	db.Exec("DROP TABLE IF EXISTS user_activity_logs")

	actRepo := repository.NewActivityRepository(db)
	actSvc := services.NewActivityService(actRepo)

	svc := newMockAuthService(&mockUserRepo{}, &mockSessionRepo{})

	// Build a handler with both authService and activityService set
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   err.Error(),
			})
		},
	})
	app.Use(recover.New())

	h := &AuthHandler{
		authService:     svc,
		sessionService:  nil,
		activityService: actSvc,
		rbacService:     nil,
		validate:        validator.New(),
	}

	auth := app.Group("/auth")
	auth.Use(withUserID(testUserID))
	auth.Get("/activity", h.GetUserActivity)

	resp := testRequest(app, http.MethodGet, "/auth/activity", nil)

	// The table doesn't exist → GetUserActivity returns error → 500
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}
