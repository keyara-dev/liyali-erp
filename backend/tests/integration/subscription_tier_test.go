package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSubscriptionTierChangeIntegration tests the complete flow
func TestSubscriptionTierChangeIntegration(t *testing.T) {
	// Skip if not in integration test mode
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test database and app
	app, db := setupTestApp(t)
	defer db.Close()

	// Create test organization
	orgID := createTestOrganization(t, db, "Test Org", "basic")
	defer cleanupTestOrganization(t, db, orgID)

	// Create super admin user
	adminID, adminToken := createTestSuperAdmin(t, db)
	defer cleanupTestUser(t, db, adminID)

	// Test 1: Successful tier upgrade
	t.Run("Upgrade_Starter_To_Pro", func(t *testing.T) {
		reqBody := map[string]string{
			"newTier": "pro",
			"reason":  "Customer requested upgrade to pro plan",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(
			http.MethodPost,
			fmt.Sprintf("/api/v1/admin/organizations/%s/change-tier", orgID),
			bytes.NewReader(body),
		)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Parse response
		var response map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&response)

		assert.True(t, response["success"].(bool))
		data := response["data"].(map[string]interface{})
		assert.Equal(t, "starter", data["old_tier"])
		assert.Equal(t, "pro", data["new_tier"])

		// Verify database was updated
		var currentTier string
		err = db.QueryRow("SELECT subscription_tier FROM organizations WHERE id = $1", orgID).Scan(&currentTier)
		require.NoError(t, err)
		assert.Equal(t, "pro", currentTier)

		// Verify subscription event was created
		var eventCount int
		err = db.QueryRow(`
			SELECT COUNT(*) FROM subscription_events 
			WHERE organization_id = $1 
			AND event_type = 'subscription_upgraded'
			AND from_tier = 'starter'
			AND to_tier = 'pro'
		`, orgID).Scan(&eventCount)
		require.NoError(t, err)
		assert.Equal(t, 1, eventCount)

		// Verify audit log was created
		var auditCount int
		err = db.QueryRow(`
			SELECT COUNT(*) FROM admin_audit_logs 
			WHERE organization_id = $1 
			AND action = 'tier_change'
			AND old_value = 'starter'
			AND new_value = 'pro'
		`, orgID).Scan(&auditCount)
		require.NoError(t, err)
		assert.Equal(t, 1, auditCount)
	})

	// Test 2: Downgrade tier
	t.Run("Downgrade_Pro_To_Starter", func(t *testing.T) {
		reqBody := map[string]string{
			"newTier": "starter",
			"reason":  "Customer requested downgrade to starter plan",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(
			http.MethodPost,
			fmt.Sprintf("/api/v1/admin/organizations/%s/change-tier", orgID),
			bytes.NewReader(body),
		)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Verify subscription event type is downgrade
		var eventType string
		err = db.QueryRow(`
			SELECT event_type FROM subscription_events 
			WHERE organization_id = $1 
			AND from_tier = 'pro'
			AND to_tier = 'starter'
			ORDER BY created_at DESC LIMIT 1
		`, orgID).Scan(&eventType)
		require.NoError(t, err)
		assert.Equal(t, "subscription_downgraded", eventType)
	})

	// Test 3: Upgrade to Custom (unlimited)
	t.Run("Upgrade_Pro_To_Custom", func(t *testing.T) {
		reqBody := map[string]string{
			"newTier": "custom",
			"reason":  "Customer needs unlimited resources",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(
			http.MethodPost,
			fmt.Sprintf("/api/v1/admin/organizations/%s/change-tier", orgID),
			bytes.NewReader(body),
		)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Verify tier is custom
		var currentTier string
		err = db.QueryRow("SELECT subscription_tier FROM organizations WHERE id = $1", orgID).Scan(&currentTier)
		require.NoError(t, err)
		assert.Equal(t, "custom", currentTier)
	})

	// Test 4: Unauthorized access (non-super admin)
	t.Run("Unauthorized_Regular_User", func(t *testing.T) {
		// Create regular user
		regularUserID, regularToken := createTestRegularUser(t, db)
		defer cleanupTestUser(t, db, regularUserID)

		reqBody := map[string]string{
			"newTier": "custom",
			"reason":  "Unauthorized attempt",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(
			http.MethodPost,
			fmt.Sprintf("/api/v1/admin/organizations/%s/change-tier", orgID),
			bytes.NewReader(body),
		)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+regularToken)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	// Test 5: Invalid tier
	t.Run("Invalid_Tier", func(t *testing.T) {
		reqBody := map[string]string{
			"newTier": "invalid_tier",
			"reason":  "Testing invalid tier",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(
			http.MethodPost,
			fmt.Sprintf("/api/v1/admin/organizations/%s/change-tier", orgID),
			bytes.NewReader(body),
		)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	// Test 6: Short reason
	t.Run("Short_Reason", func(t *testing.T) {
		reqBody := map[string]string{
			"newTier": "custom",
			"reason":  "Short",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(
			http.MethodPost,
			fmt.Sprintf("/api/v1/admin/organizations/%s/change-tier", orgID),
			bytes.NewReader(body),
		)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	// Test 7: Organization not found
	t.Run("Organization_Not_Found", func(t *testing.T) {
		reqBody := map[string]string{
			"newTier": "pro",
			"reason":  "Testing non-existent organization",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(
			http.MethodPost,
			"/api/v1/admin/organizations/org-nonexistent/change-tier",
			bytes.NewReader(body),
		)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	// Test 8: Same tier (no change)
	t.Run("Same_Tier_No_Change", func(t *testing.T) {
		// Get current tier
		var currentTier string
		db.QueryRow("SELECT subscription_tier FROM organizations WHERE id = $1", orgID).Scan(&currentTier)

		reqBody := map[string]string{
			"newTier": currentTier,
			"reason":  "Testing same tier change",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(
			http.MethodPost,
			fmt.Sprintf("/api/v1/admin/organizations/%s/change-tier", orgID),
			bytes.NewReader(body),
		)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	// Test 9: All tier combinations
	t.Run("All_Tier_Combinations", func(t *testing.T) {
		tiers := []string{"starter", "pro", "custom"}

		for _, targetTier := range tiers {
			reqBody := map[string]string{
				"newTier": targetTier,
				"reason":  fmt.Sprintf("Testing change to %s tier", targetTier),
			}
			body, _ := json.Marshal(reqBody)

			req := httptest.NewRequest(
				http.MethodPost,
				fmt.Sprintf("/api/v1/admin/organizations/%s/change-tier", orgID),
				bytes.NewReader(body),
			)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+adminToken)

			resp, err := app.Test(req)
			require.NoError(t, err)

			// Should succeed if different tier, fail if same
			var currentTier string
			db.QueryRow("SELECT subscription_tier FROM organizations WHERE id = $1", orgID).Scan(&currentTier)

			if currentTier == targetTier {
				assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
			} else {
				assert.Equal(t, http.StatusOK, resp.StatusCode)
			}
		}
	})

	// Test 10: Audit trail completeness
	t.Run("Audit_Trail_Completeness", func(t *testing.T) {
		// Change tier
		reqBody := map[string]string{
			"newTier": "custom",
			"reason":  "Testing audit trail completeness",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(
			http.MethodPost,
			fmt.Sprintf("/api/v1/admin/organizations/%s/change-tier", orgID),
			bytes.NewReader(body),
		)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Verify audit log has all required fields
		var auditLog struct {
			ID             string
			OrganizationID string
			Action         string
			OldValue       string
			NewValue       string
			Reason         string
			AdminUserID    string
			CreatedAt      string
		}

		err = db.QueryRow(`
			SELECT id, organization_id, action, old_value, new_value, reason, admin_user_id, created_at
			FROM admin_audit_logs 
			WHERE organization_id = $1 
			AND new_value = 'custom'
			ORDER BY created_at DESC LIMIT 1
		`, orgID).Scan(
			&auditLog.ID,
			&auditLog.OrganizationID,
			&auditLog.Action,
			&auditLog.OldValue,
			&auditLog.NewValue,
			&auditLog.Reason,
			&auditLog.AdminUserID,
			&auditLog.CreatedAt,
		)

		require.NoError(t, err)
		assert.NotEmpty(t, auditLog.ID)
		assert.Equal(t, orgID, auditLog.OrganizationID)
		assert.Equal(t, "tier_change", auditLog.Action)
		assert.Equal(t, "custom", auditLog.NewValue)
		assert.Equal(t, "Testing audit trail completeness", auditLog.Reason)
		assert.Equal(t, adminID, auditLog.AdminUserID)
		assert.NotEmpty(t, auditLog.CreatedAt)
	})
}

// Helper functions

func setupTestApp(t *testing.T) (*fiber.App, *sql.DB) {
	// Setup test database connection
	// Setup test app with routes
	// Return both for testing
	// Implementation depends on your app structure
	return nil, nil
}

func createTestOrganization(t *testing.T, db *sql.DB, name, tier string) string {
	// Create test organization in database with starter tier by default
	// Return organization ID
	return "org-test-001"
}

func cleanupTestOrganization(t *testing.T, db *sql.DB, orgID string) {
	// Delete test organization and related data
}

func createTestSuperAdmin(t *testing.T, db *sql.DB) (string, string) {
	// Create super admin user
	// Generate JWT token
	// Return user ID and token
	return "user-admin-test", "test-token"
}

func createTestRegularUser(t *testing.T, db *sql.DB) (string, string) {
	// Create regular user
	// Generate JWT token
	// Return user ID and token
	return "user-regular-test", "test-token"
}

func cleanupTestUser(t *testing.T, db *sql.DB, userID string) {
	// Delete test user
}
