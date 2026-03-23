package handlers

import (
	"net/http"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// newAuthApp builds a minimal Fiber app wired to the AuthHandler validation
// layer.  Because AuthHandler depends on concrete *services.AuthService and
// *services.RBACService types that require live database/repository wiring,
// we construct the handler struct directly with nil services.
//
// Every test in this file is written so the handler returns before it reaches
// any service call — either because validation fails (returning 400) or
// because the auth-context check fails (returning 401).
//
// Optional middlewares can be passed to simulate an authenticated context
// (e.g. setting c.Locals("userID")).
func newAuthApp(injectedMiddleware ...fiber.Handler) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   err.Error(),
			})
		},
	})

	// Construct the handler with nil services.  The validate field is
	// initialised independently so the validation layer works correctly.
	h := &AuthHandler{
		authService: nil,
		rbacService: nil,
		validate:    validator.New(),
	}

	auth := app.Group("/auth")
	for _, mw := range injectedMiddleware {
		auth.Use(mw)
	}

	auth.Post("/login", h.Login)
	auth.Get("/profile", h.GetProfile)
	auth.Post("/change-password", h.ChangePassword)
	auth.Post("/logout", h.Logout)
	auth.Post("/logout-all", h.LogoutAll)

	return app
}

// withUserID returns a Fiber middleware that injects the given userID into
// c.Locals("userID"), simulating an authenticated request.
func withUserID(userID string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Locals("userID", userID)
		return c.Next()
	}
}

// ---------------------------------------------------------------------------
// POST /auth/login — validation layer tests
// ---------------------------------------------------------------------------

// TestLogin_EmptyBody verifies that an empty body causes a 400 because both
// email and password are required fields.
func TestLogin_EmptyBody(t *testing.T) {
	app := newAuthApp()

	resp := testRequest(app, http.MethodPost, "/auth/login", nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for empty body, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	if body["success"] != false {
		t.Errorf("expected success=false")
	}
}

// TestLogin_MissingEmail verifies that omitting the email field returns 400.
func TestLogin_MissingEmail(t *testing.T) {
	app := newAuthApp()

	resp := testRequest(app, http.MethodPost, "/auth/login", map[string]interface{}{
		// email intentionally omitted
		"password": "secret123",
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for missing email, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	if body["success"] != false {
		t.Errorf("expected success=false")
	}
}

// TestLogin_MissingPassword verifies that omitting the password field returns 400.
func TestLogin_MissingPassword(t *testing.T) {
	app := newAuthApp()

	resp := testRequest(app, http.MethodPost, "/auth/login", map[string]interface{}{
		"email": "user@example.com",
		// password intentionally omitted
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for missing password, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	if body["success"] != false {
		t.Errorf("expected success=false")
	}
}

// TestLogin_InvalidEmailFormat verifies that an email failing the `email`
// validator tag returns 400.
func TestLogin_InvalidEmailFormat(t *testing.T) {
	app := newAuthApp()

	resp := testRequest(app, http.MethodPost, "/auth/login", map[string]interface{}{
		"email":    "notanemail",
		"password": "secret123",
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid email format, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	if body["success"] != false {
		t.Errorf("expected success=false")
	}
}

// ---------------------------------------------------------------------------
// GET /auth/profile — NoAuth check
// ---------------------------------------------------------------------------

// TestGetProfile_NoAuth verifies that GetProfile returns 401 when userID is
// not present in the request context (no auth middleware ran).
func TestGetProfile_NoAuth(t *testing.T) {
	app := newAuthApp() // no withUserID middleware

	resp := testRequest(app, http.MethodGet, "/auth/profile", nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 for unauthenticated profile request, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	if body["success"] != false {
		t.Errorf("expected success=false")
	}
}

// ---------------------------------------------------------------------------
// POST /auth/change-password
// ---------------------------------------------------------------------------

// TestChangePassword_NoAuth verifies that ChangePassword returns 401 when the
// request carries no authenticated user context.
func TestChangePassword_NoAuth(t *testing.T) {
	app := newAuthApp() // no withUserID middleware

	resp := testRequest(app, http.MethodPost, "/auth/change-password", map[string]interface{}{
		"currentPassword": "old-password",
		"newPassword":     "new-password-secure",
	})
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 for unauthenticated change-password, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	if body["success"] != false {
		t.Errorf("expected success=false")
	}
}

// TestChangePassword_MissingCurrentPassword verifies that the handler returns
// 400 when currentPassword is absent (userID is present).
func TestChangePassword_MissingCurrentPassword(t *testing.T) {
	app := newAuthApp(withUserID(testUserID))

	resp := testRequest(app, http.MethodPost, "/auth/change-password", map[string]interface{}{
		// currentPassword intentionally omitted
		"newPassword": "new-password-secure",
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for missing currentPassword, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	if body["success"] != false {
		t.Errorf("expected success=false")
	}
}

// TestChangePassword_MissingNewPassword verifies that the handler returns 400
// when newPassword is absent (userID is present).
func TestChangePassword_MissingNewPassword(t *testing.T) {
	app := newAuthApp(withUserID(testUserID))

	resp := testRequest(app, http.MethodPost, "/auth/change-password", map[string]interface{}{
		"currentPassword": "old-password",
		// newPassword intentionally omitted
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for missing newPassword, got %d", resp.StatusCode)
	}
}

// TestChangePassword_NewPasswordTooShort verifies that the handler returns 400
// when newPassword is shorter than 8 characters (min=8 validation tag).
func TestChangePassword_NewPasswordTooShort(t *testing.T) {
	app := newAuthApp(withUserID(testUserID))

	resp := testRequest(app, http.MethodPost, "/auth/change-password", map[string]interface{}{
		"currentPassword": "old-password",
		"newPassword":     "short",
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for short newPassword, got %d", resp.StatusCode)
	}
}

// TestChangePassword_EmptyBody verifies that sending an empty body with a
// valid userID still returns 400.
func TestChangePassword_EmptyBody(t *testing.T) {
	app := newAuthApp(withUserID(testUserID))

	resp := testRequest(app, http.MethodPost, "/auth/change-password", map[string]interface{}{})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for empty change-password body, got %d", resp.StatusCode)
	}
}

// ---------------------------------------------------------------------------
// POST /auth/logout — validation layer tests
// ---------------------------------------------------------------------------

// TestLogout_MissingRefreshToken verifies that Logout returns 400 when the
// required refreshToken field is absent.
//
// NOTE: The Logout handler does not itself check c.Locals("userID") — it
// relies on JWT middleware applied at the route level in production.  The
// handler's own guard is the refreshToken validation tag.  Sending no
// refreshToken must therefore yield 400, not 401.
func TestLogout_MissingRefreshToken(t *testing.T) {
	app := newAuthApp()

	resp := testRequest(app, http.MethodPost, "/auth/logout", map[string]interface{}{
		// refreshToken intentionally omitted
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for missing refreshToken, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	if body["success"] != false {
		t.Errorf("expected success=false")
	}
}

// TestLogout_EmptyBody verifies that sending a nil body returns 400 because
// the refreshToken validation tag fires.
func TestLogout_EmptyBody(t *testing.T) {
	app := newAuthApp()

	resp := testRequest(app, http.MethodPost, "/auth/logout", nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for empty logout body, got %d", resp.StatusCode)
	}
}

// ---------------------------------------------------------------------------
// POST /auth/logout-all — NoAuth check
// ---------------------------------------------------------------------------

// TestLogoutAll_NoAuth verifies that LogoutAll returns 401 when no userID is
// present in the request context.
func TestLogoutAll_NoAuth(t *testing.T) {
	app := newAuthApp() // no withUserID middleware

	resp := testRequest(app, http.MethodPost, "/auth/logout-all", nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 for unauthenticated logout-all, got %d", resp.StatusCode)
	}

	body := decodeResponse(resp)
	if body["success"] != false {
		t.Errorf("expected success=false")
	}
}
