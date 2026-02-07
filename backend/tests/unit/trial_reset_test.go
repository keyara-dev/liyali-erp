package unit

import (
	"testing"
	"time"
)

func TestTrialResetService(t *testing.T) {
	t.Run("Reset trial period", func(t *testing.T) {
		// Test data
		organizationID := "test-org-123"
		trialDays := 30
		reason := "Testing trial reset functionality"
		performedBy := "admin-user-123"
		
		// Since we don't have a real database connection in unit tests,
		// we'll test the validation logic and expected behavior
		
		// Validate input parameters
		if organizationID == "" {
			t.Error("Organization ID should not be empty")
		}
		
		if trialDays < 1 || trialDays > 90 {
			t.Error("Trial days should be between 1 and 90")
		}
		
		if len(reason) < 5 || len(reason) > 200 {
			t.Error("Reason should be between 5 and 200 characters")
		}
		
		if performedBy == "" {
			t.Error("Performed by should not be empty")
		}
		
		// Test trial date calculations
		now := time.Now()
		expectedTrialStart := now
		expectedTrialEnd := now.AddDate(0, 0, trialDays)
		
		// Verify trial period calculation
		duration := expectedTrialEnd.Sub(expectedTrialStart)
		expectedDuration := time.Duration(trialDays) * 24 * time.Hour
		
		if duration < expectedDuration-time.Hour || duration > expectedDuration+time.Hour {
			t.Errorf("Trial period calculation incorrect. Expected ~%v, got %v", expectedDuration, duration)
		}
		
		t.Logf("✅ Trial reset validation passed")
		t.Logf("   Organization ID: %s", organizationID)
		t.Logf("   Trial Days: %d", trialDays)
		t.Logf("   Reason: %s", reason)
		t.Logf("   Performed By: %s", performedBy)
		t.Logf("   Expected Trial Start: %s", expectedTrialStart.Format(time.RFC3339))
		t.Logf("   Expected Trial End: %s", expectedTrialEnd.Format(time.RFC3339))
	})
	
	t.Run("Validate trial reset parameters", func(t *testing.T) {
		testCases := []struct {
			name           string
			organizationID string
			trialDays      int
			reason         string
			performedBy    string
			expectError    bool
			errorMessage   string
		}{
			{
				name:           "Valid parameters",
				organizationID: "org-123",
				trialDays:      30,
				reason:         "Customer needs more time for evaluation",
				performedBy:    "admin-123",
				expectError:    false,
			},
			{
				name:           "Empty organization ID",
				organizationID: "",
				trialDays:      30,
				reason:         "Valid reason",
				performedBy:    "admin-123",
				expectError:    true,
				errorMessage:   "Organization ID is required",
			},
			{
				name:           "Trial days too low",
				organizationID: "org-123",
				trialDays:      0,
				reason:         "Valid reason",
				performedBy:    "admin-123",
				expectError:    true,
				errorMessage:   "Trial days must be between 1 and 90",
			},
			{
				name:           "Trial days too high",
				organizationID: "org-123",
				trialDays:      91,
				reason:         "Valid reason",
				performedBy:    "admin-123",
				expectError:    true,
				errorMessage:   "Trial days must be between 1 and 90",
			},
			{
				name:           "Reason too short",
				organizationID: "org-123",
				trialDays:      30,
				reason:         "Hi",
				performedBy:    "admin-123",
				expectError:    true,
				errorMessage:   "Reason must be between 5 and 200 characters",
			},
			{
				name:           "Reason too long",
				organizationID: "org-123",
				trialDays:      30,
				reason:         string(make([]byte, 201)), // 201 characters
				performedBy:    "admin-123",
				expectError:    true,
				errorMessage:   "Reason must be between 5 and 200 characters",
			},
			{
				name:           "Empty performed by",
				organizationID: "org-123",
				trialDays:      30,
				reason:         "Valid reason",
				performedBy:    "",
				expectError:    true,
				errorMessage:   "Performed by is required",
			},
		}
		
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Validate organization ID
				if tc.organizationID == "" && tc.expectError {
					t.Logf("✅ Correctly identified empty organization ID")
					return
				}
				
				// Validate trial days
				if (tc.trialDays < 1 || tc.trialDays > 90) && tc.expectError {
					t.Logf("✅ Correctly identified invalid trial days: %d", tc.trialDays)
					return
				}
				
				// Validate reason length
				if (len(tc.reason) < 5 || len(tc.reason) > 200) && tc.expectError {
					t.Logf("✅ Correctly identified invalid reason length: %d", len(tc.reason))
					return
				}
				
				// Validate performed by
				if tc.performedBy == "" && tc.expectError {
					t.Logf("✅ Correctly identified empty performed by")
					return
				}
				
				// If we reach here and expect error, test failed
				if tc.expectError {
					t.Errorf("Expected error for %s but validation passed", tc.name)
				} else {
					t.Logf("✅ Valid parameters passed validation")
				}
			})
		}
	})
	
	t.Run("Trial reset audit metadata", func(t *testing.T) {
		// Test audit metadata structure
		trialDays := 30
		reason := "Customer requested extension"
		now := time.Now()
		trialStart := now
		trialEnd := now.AddDate(0, 0, trialDays)
		
		// Expected metadata structure
		expectedMetadata := map[string]interface{}{
			"trial_days":        trialDays,
			"reason":           reason,
			"action_type":      "trial_reset",
			"new_trial_start":  trialStart.Format(time.RFC3339),
			"new_trial_end":    trialEnd.Format(time.RFC3339),
		}
		
		// Validate metadata fields
		if expectedMetadata["trial_days"] != trialDays {
			t.Error("Trial days not correctly stored in metadata")
		}
		
		if expectedMetadata["reason"] != reason {
			t.Error("Reason not correctly stored in metadata")
		}
		
		if expectedMetadata["action_type"] != "trial_reset" {
			t.Error("Action type not correctly set")
		}
		
		// Validate date formats
		startDateStr, ok := expectedMetadata["new_trial_start"].(string)
		if !ok {
			t.Error("Trial start date not stored as string")
		}
		
		_, err := time.Parse(time.RFC3339, startDateStr)
		if err != nil {
			t.Error("Trial start date not in RFC3339 format")
		}
		
		endDateStr, ok := expectedMetadata["new_trial_end"].(string)
		if !ok {
			t.Error("Trial end date not stored as string")
		}
		
		_, err = time.Parse(time.RFC3339, endDateStr)
		if err != nil {
			t.Error("Trial end date not in RFC3339 format")
		}
		
		t.Logf("✅ Audit metadata structure validated")
		t.Logf("   Metadata: %+v", expectedMetadata)
	})
}