package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAdminOrganizationService is a mock implementation
type MockAdminOrganizationService struct {
	mock.Mock
}

func (m *MockAdminOrganizationService) ChangeSubscriptionTier(
	organizationID, oldTier, newTier, reason, adminUserID, ipAddress string,
) (*ChangeTierResponse, error) {
	args := m.Called(organizationID, oldTier, newTier, reason, adminUserID, ipAddress)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ChangeTierResponse), args.Error(1)
}

func (m *MockAdminOrganizationService) GetOrganizationByID(organizationID string) (*Organization, error) {
	args := m.Called(organizationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Organization), args.Error(1)
}

func (m *MockAdminOrganizationService) IsSuperAdmin(userID string) (bool, error) {
	args := m.Called(userID)
	return args.Bool(0), args.Error(1)
}

func TestChangeSubscriptionTier_Success(t *testing.T) {
	// Setup
	app := fiber.New()
	mockService := new(MockAdminOrganizationService)
	handler := &AdminOrganizationHandler{
		service: mockService,
	}

	// Mock data
	organizationID := "org-test-001"
	userID := "user-admin-001"
	oldTier := "basic"
	newTier := "professional"

	// Mock expectations
	mockService.On("IsSuperAdmin", userID).Return(true, nil)
	mockService.On("GetOrganizationByID", organizationID).Return(&Organization{
		ID:               organizationID,
		Name:             "Test Org",
		SubscriptionTier: oldTier,
	}, nil)
	mockService.On("ChangeSubscriptionTier",
		organizationID,
		oldTier,
		newTier,
		"Customer upgrade request",
		userID,
		mock.Anything,
	).Return(&ChangeTierResponse{
		OrganizationID: organizationID,
		OldTier:        oldTier,
		NewTier:        newTier,
		ChangedBy:      userID,
		Reason:         "Customer upgrade request",
	}, nil)

	// Setup route
	app.Post("/api/v1/admin/organizations/:id/subscription-tier", func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		return handler.ChangeSubscriptionTier(c)
	})

	// Request body
	reqBody := map[string]string{
		"tier":   newTier,
		"reason": "Customer upgrade request",
	}
	body, _ := json.Marshal(reqBody)

	// Make request
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/organizations/"+organizationID+"/subscription-tier", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Parse response
	var response map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&response)

	assert.True(t, response["success"].(bool))
	assert.Equal(t, "Subscription tier changed successfully", response["message"])

	mockService.AssertExpectations(t)
}

func TestChangeSubscriptionTier_Unauthorized(t *testing.T) {
	// Setup
	app := fiber.New()
	mockService := new(MockAdminOrganizationService)
	handler := &AdminOrganizationHandler{
		service: mockService,
	}

	// Mock data
	organizationID := "org-test-001"
	userID := "user-regular-001"

	// Mock expectations - not super admin
	mockService.On("IsSuperAdmin", userID).Return(false, nil)

	// Setup route
	app.Post("/api/v1/admin/organizations/:id/subscription-tier", func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		return handler.ChangeSubscriptionTier(c)
	})

	// Request body
	reqBody := map[string]string{
		"tier":   "professional",
		"reason": "Test",
	}
	body, _ := json.Marshal(reqBody)

	// Make request
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/organizations/"+organizationID+"/subscription-tier", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)

	// Parse response
	var response map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&response)

	assert.False(t, response["success"].(bool))
	assert.Contains(t, response["message"], "Insufficient permissions")

	mockService.AssertExpectations(t)
}

func TestChangeSubscriptionTier_InvalidTier(t *testing.T) {
	// Setup
	app := fiber.New()
	mockService := new(MockAdminOrganizationService)
	handler := &AdminOrganizationHandler{
		service: mockService,
	}

	// Mock data
	organizationID := "org-test-001"
	userID := "user-admin-001"

	// Mock expectations
	mockService.On("IsSuperAdmin", userID).Return(true, nil)

	// Setup route
	app.Post("/api/v1/admin/organizations/:id/subscription-tier", func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		return handler.ChangeSubscriptionTier(c)
	})

	// Request body with invalid tier
	reqBody := map[string]string{
		"tier":   "invalid_tier",
		"reason": "Test invalid tier",
	}
	body, _ := json.Marshal(reqBody)

	// Make request
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/organizations/"+organizationID+"/subscription-tier", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// Parse response
	var response map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&response)

	assert.False(t, response["success"].(bool))
	assert.Contains(t, response["error"], "tier")

	mockService.AssertExpectations(t)
}

func TestChangeSubscriptionTier_ShortReason(t *testing.T) {
	// Setup
	app := fiber.New()
	mockService := new(MockAdminOrganizationService)
	handler := &AdminOrganizationHandler{
		service: mockService,
	}

	// Mock data
	organizationID := "org-test-001"
	userID := "user-admin-001"

	// Mock expectations
	mockService.On("IsSuperAdmin", userID).Return(true, nil)

	// Setup route
	app.Post("/api/v1/admin/organizations/:id/subscription-tier", func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		return handler.ChangeSubscriptionTier(c)
	})

	// Request body with short reason
	reqBody := map[string]string{
		"tier":   "professional",
		"reason": "Short",
	}
	body, _ := json.Marshal(reqBody)

	// Make request
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/organizations/"+organizationID+"/subscription-tier", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// Parse response
	var response map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&response)

	assert.False(t, response["success"].(bool))
	assert.Contains(t, response["error"], "reason")

	mockService.AssertExpectations(t)
}

func TestChangeSubscriptionTier_OrganizationNotFound(t *testing.T) {
	// Setup
	app := fiber.New()
	mockService := new(MockAdminOrganizationService)
	handler := &AdminOrganizationHandler{
		service: mockService,
	}

	// Mock data
	organizationID := "org-nonexistent"
	userID := "user-admin-001"

	// Mock expectations
	mockService.On("IsSuperAdmin", userID).Return(true, nil)
	mockService.On("GetOrganizationByID", organizationID).Return(nil, fiber.NewError(fiber.StatusNotFound, "Organization not found"))

	// Setup route
	app.Post("/api/v1/admin/organizations/:id/subscription-tier", func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		return handler.ChangeSubscriptionTier(c)
	})

	// Request body
	reqBody := map[string]string{
		"tier":   "professional",
		"reason": "Test organization not found",
	}
	body, _ := json.Marshal(reqBody)

	// Make request
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/organizations/"+organizationID+"/subscription-tier", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	mockService.AssertExpectations(t)
}

func TestChangeSubscriptionTier_SameTier(t *testing.T) {
	// Setup
	app := fiber.New()
	mockService := new(MockAdminOrganizationService)
	handler := &AdminOrganizationHandler{
		service: mockService,
	}

	// Mock data
	organizationID := "org-test-001"
	userID := "user-admin-001"
	currentTier := "professional"

	// Mock expectations
	mockService.On("IsSuperAdmin", userID).Return(true, nil)
	mockService.On("GetOrganizationByID", organizationID).Return(&Organization{
		ID:               organizationID,
		Name:             "Test Org",
		SubscriptionTier: currentTier,
	}, nil)

	// Setup route
	app.Post("/api/v1/admin/organizations/:id/subscription-tier", func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		return handler.ChangeSubscriptionTier(c)
	})

	// Request body with same tier
	reqBody := map[string]string{
		"tier":   currentTier,
		"reason": "Test same tier change",
	}
	body, _ := json.Marshal(reqBody)

	// Make request
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/organizations/"+organizationID+"/subscription-tier", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// Parse response
	var response map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&response)

	assert.False(t, response["success"].(bool))
	assert.Contains(t, response["error"], "already on this tier")

	mockService.AssertExpectations(t)
}

func TestChangeSubscriptionTier_MissingRequestBody(t *testing.T) {
	// Setup
	app := fiber.New()
	mockService := new(MockAdminOrganizationService)
	handler := &AdminOrganizationHandler{
		service: mockService,
	}

	// Mock data
	organizationID := "org-test-001"
	userID := "user-admin-001"

	// Mock expectations
	mockService.On("IsSuperAdmin", userID).Return(true, nil)

	// Setup route
	app.Post("/api/v1/admin/organizations/:id/subscription-tier", func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		return handler.ChangeSubscriptionTier(c)
	})

	// Make request with empty body
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/organizations/"+organizationID+"/subscription-tier", nil)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	mockService.AssertExpectations(t)
}
