package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// newSubscriptionApp builds a minimal Fiber app wired to the SubscriptionHandler
// endpoints.  The handler is constructed with a nil service so that any code
// path that reaches the service layer returns a non-200 response (the handler
// converts service panics / nil-dereference errors to 500 via the error
// handler, or testRequest returns a synthetic 500 on panic).
func newSubscriptionApp() *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   err.Error(),
			})
		},
	})
	app.Use(recover.New())

	h := &SubscriptionHandler{subscriptionService: nil}

	sub := app.Group("/subscription")
	sub.Get("/plans", h.GetSubscriptionPlans)
	sub.Get("/trial/:id", h.GetOrganizationTrialStatus)
	sub.Post("/check-feature", h.CheckFeatureAccess)
	sub.Post("/upgrade/:id", h.UpgradeOrganization)
	sub.Post("/extend-trial/:id", h.ExtendOrganizationTrial)
	sub.Post("/reset-trial/:id", h.ResetOrganizationTrial)
	// /:id must come last so it doesn't swallow the named sub-routes above.
	sub.Get("/:id", h.GetOrganizationSubscription)

	return app
}

// withUserIDLocal returns a Fiber middleware that sets the "user_id" local
// required by ExtendOrganizationTrial and ResetOrganizationTrial.
func withUserIDLocal(userID string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		return c.Next()
	}
}

// newSubscriptionAppWithUserID is like newSubscriptionApp but additionally
// injects the "user_id" local so handlers that type-assert it don't panic
// before even reaching the service.
func newSubscriptionAppWithUserID(uid string) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   err.Error(),
			})
		},
	})
	app.Use(recover.New())

	h := &SubscriptionHandler{subscriptionService: nil}

	sub := app.Group("/subscription", withUserIDLocal(uid))
	sub.Get("/plans", h.GetSubscriptionPlans)
	sub.Get("/trial/:id", h.GetOrganizationTrialStatus)
	sub.Post("/check-feature", h.CheckFeatureAccess)
	sub.Post("/upgrade/:id", h.UpgradeOrganization)
	sub.Post("/extend-trial/:id", h.ExtendOrganizationTrial)
	sub.Post("/reset-trial/:id", h.ResetOrganizationTrial)
	sub.Get("/:id", h.GetOrganizationSubscription)

	return app
}

// testRawRequest fires a request with an arbitrary (possibly invalid-JSON) body
// so we can exercise body-parser failure paths.
func testRawRequest(app *fiber.App, method, path, contentType, rawBody string) *http.Response {
	req := httptest.NewRequest(method, path, bytes.NewBufferString(rawBody))
	req.Header.Set("Content-Type", contentType)

	resp, err := app.Test(req, -1)
	if err != nil {
		return &http.Response{StatusCode: http.StatusInternalServerError}
	}
	return resp
}

// ---------------------------------------------------------------------------
// GetSubscriptionPlans  GET /subscription/plans
// ---------------------------------------------------------------------------

// TestGetSubscriptionPlans_NilService checks that a nil service results in a
// non-200 (panic → 500 via Fiber error handler or testRequest fallback).
func TestGetSubscriptionPlans_NilService(t *testing.T) {
	app := newSubscriptionApp()

	resp := testRequest(app, http.MethodGet, "/subscription/plans", nil)
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// GetOrganizationTrialStatus  GET /subscription/trial/:id
// ---------------------------------------------------------------------------

func TestGetOrganizationTrialStatus_NilService(t *testing.T) {
	app := newSubscriptionApp()

	resp := testRequest(app, http.MethodGet, "/subscription/trial/org-123", nil)
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// CheckFeatureAccess  POST /subscription/check-feature
// ---------------------------------------------------------------------------

// The handler reads organizationID from c.Params("id") which will be empty on
// the /check-feature route → 400 before the service is called.
func TestCheckFeatureAccess_MissingOrgID(t *testing.T) {
	app := newSubscriptionApp()

	resp := testRequest(app, http.MethodPost, "/subscription/check-feature", nil)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

// ---------------------------------------------------------------------------
// UpgradeOrganization  POST /subscription/upgrade/:id
// ---------------------------------------------------------------------------

func TestUpgradeOrganization_InvalidBody(t *testing.T) {
	app := newSubscriptionApp()

	// Malformed JSON → BodyParser returns an error → 400.
	resp := testRawRequest(app, http.MethodPost, "/subscription/upgrade/org-123",
		"application/json", "not-valid-json{{{")
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

func TestUpgradeOrganization_ValidBody_NilService(t *testing.T) {
	app := newSubscriptionApp()

	resp := testRequest(app, http.MethodPost, "/subscription/upgrade/org-123", map[string]interface{}{
		"plan": "premium",
	})
	// Service is nil → non-200.
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// GetOrganizationSubscription  GET /subscription/:id
// ---------------------------------------------------------------------------

func TestGetOrganizationSubscription_NilService(t *testing.T) {
	app := newSubscriptionApp()

	resp := testRequest(app, http.MethodGet, "/subscription/org-123", nil)
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// ExtendOrganizationTrial  POST /subscription/extend-trial/:id
// ---------------------------------------------------------------------------

func TestExtendOrganizationTrial_NilService(t *testing.T) {
	app := newSubscriptionAppWithUserID(testUserID)

	resp := testRequest(app, http.MethodPost, "/subscription/extend-trial/org-123", map[string]interface{}{
		"daysToAdd": 5,
		"reason":    "extended for testing period",
	})
	// Service is nil → non-200.
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}

func TestExtendOrganizationTrial_InvalidBody(t *testing.T) {
	app := newSubscriptionAppWithUserID(testUserID)

	// Malformed JSON → BodyParser error → 400.
	resp := testRawRequest(app, http.MethodPost, "/subscription/extend-trial/org-123",
		"application/json", "bad-json{{}")
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

// ---------------------------------------------------------------------------
// ResetOrganizationTrial  POST /subscription/reset-trial/:id
// ---------------------------------------------------------------------------

func TestResetOrganizationTrial_NilService(t *testing.T) {
	app := newSubscriptionAppWithUserID(testUserID)

	resp := testRequest(app, http.MethodPost, "/subscription/reset-trial/org-123", map[string]interface{}{
		"trialDays": 14,
		"reason":    "reset for new billing cycle",
	})
	// Service is nil → non-200.
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}

func TestResetOrganizationTrial_InvalidBody(t *testing.T) {
	app := newSubscriptionAppWithUserID(testUserID)

	// Malformed JSON → BodyParser error → 400.
	resp := testRawRequest(app, http.MethodPost, "/subscription/reset-trial/org-123",
		"application/json", "bad{{json")
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body := decodeResponse(resp)
	assert.Equal(t, false, body["success"])
}

// ---------------------------------------------------------------------------
// Response shape smoke test
// ---------------------------------------------------------------------------

// TestSubscriptionErrorResponse_Shape verifies that 400 error responses from
// subscription handlers include a "success: false" field.
func TestSubscriptionErrorResponse_Shape(t *testing.T) {
	app := newSubscriptionApp()

	// CheckFeatureAccess with no org ID → guaranteed 400.
	resp := testRequest(app, http.MethodPost, "/subscription/check-feature", nil)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body := decodeResponse(resp)
	assert.NotNil(t, body)
	assert.Equal(t, false, body["success"])
}
