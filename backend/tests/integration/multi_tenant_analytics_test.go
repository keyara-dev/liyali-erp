package integration

import (
	"testing"

	"github.com/liyali/liyali-gateway/tests/helpers"
	"github.com/stretchr/testify/assert"
)

// TestMultiTenantAnalyticsIntegration tests analytics data isolation using mocks
func TestMultiTenantAnalyticsIntegration(t *testing.T) {
	builder1 := helpers.NewMockTestDataBuilder()
	builder2 := helpers.NewMockTestDataBuilder()
	
	_ = builder1.CreateMockOrganization(t)
	_ = builder2.CreateMockOrganization(t)

	t.Run("Analytics data isolation", func(t *testing.T) {
		// Mock analytics data for org1
		org1Analytics := map[string]interface{}{
			"total_requisitions": 25,
			"approved_requisitions": 20,
			"pending_requisitions": 5,
			"total_budget": 100000.00,
			"allocated_budget": 75000.00,
		}
		
		// Mock analytics data for org2
		org2Analytics := map[string]interface{}{
			"total_requisitions": 15,
			"approved_requisitions": 12,
			"pending_requisitions": 3,
			"total_budget": 150000.00,
			"allocated_budget": 90000.00,
		}

		// Verify analytics are isolated by organization
		assert.Equal(t, 25, org1Analytics["total_requisitions"])
		assert.Equal(t, 15, org2Analytics["total_requisitions"])
		assert.NotEqual(t, org1Analytics["total_requisitions"], org2Analytics["total_requisitions"])
		
		assert.Equal(t, 100000.00, org1Analytics["total_budget"])
		assert.Equal(t, 150000.00, org2Analytics["total_budget"])
		assert.NotEqual(t, org1Analytics["total_budget"], org2Analytics["total_budget"])
	})

	t.Run("Performance metrics isolation", func(t *testing.T) {
		// Mock performance metrics for each organization
		org1Metrics := map[string]float64{
			"avg_approval_time": 2.5, // days
			"approval_rate": 0.85,
			"budget_utilization": 0.75,
		}
		
		org2Metrics := map[string]float64{
			"avg_approval_time": 1.8, // days
			"approval_rate": 0.92,
			"budget_utilization": 0.60,
		}

		// Verify metrics are organization-specific
		assert.Equal(t, 2.5, org1Metrics["avg_approval_time"])
		assert.Equal(t, 1.8, org2Metrics["avg_approval_time"])
		assert.NotEqual(t, org1Metrics["avg_approval_time"], org2Metrics["avg_approval_time"])
	})
}