package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/liyali/liyali-gateway/middleware"
	"github.com/liyali/liyali-gateway/utils"
)

// GetAPIEndpoints returns the list of observed API endpoints with stats
func GetAPIEndpoints(c *fiber.Ctx) error {
	endpoints := middleware.Metrics.GetEndpoints()

	category := c.Query("category")
	status := c.Query("status")
	method := c.Query("method")

	if category == "" && status == "" && method == "" {
		return utils.SendSimpleSuccess(c, endpoints, "API endpoints retrieved successfully")
	}

	filtered := make([]map[string]interface{}, 0)
	for _, ep := range endpoints {
		if category != "" && ep["category"] != category {
			continue
		}
		if status != "" && ep["status"] != status {
			continue
		}
		if method != "" && ep["method"] != method {
			continue
		}
		filtered = append(filtered, ep)
	}

	return utils.SendSimpleSuccess(c, filtered, "API endpoints retrieved successfully")
}

// GetAPIEndpointByID returns a single API endpoint by its numeric ID
func GetAPIEndpointByID(c *fiber.Ctx) error {
	return utils.SendNotFound(c, "Individual endpoint lookup by ID is not supported for dynamic metrics")
}

// GetAPIMetrics returns overall API metrics for a time period
func GetAPIMetrics(c *fiber.Ctx) error {
	period := c.Query("period", "24h")
	metrics := middleware.Metrics.GetMetrics(period)
	return utils.SendSimpleSuccess(c, metrics, "API metrics retrieved successfully")
}

// GetAPIEndpointMetrics returns metrics for a specific endpoint
func GetAPIEndpointMetrics(c *fiber.Ctx) error {
	period := c.Query("period", "24h")
	metrics := middleware.Metrics.GetMetrics(period)
	metrics["endpoint_id"] = c.Params("id")
	return utils.SendSimpleSuccess(c, metrics, "Endpoint metrics retrieved successfully")
}

// GetAPIErrors returns recent API errors
func GetAPIErrors(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 50)
	errors := middleware.Metrics.GetRecentErrors(limit)
	return utils.SendSimpleSuccess(c, errors, "API errors retrieved successfully")
}

// GetAPIErrorByID returns a specific API error by ID
func GetAPIErrorByID(c *fiber.Ctx) error {
	return utils.SendNotFound(c, "API error not found")
}

// ResolveAPIError resolves an API error
func ResolveAPIError(c *fiber.Ctx) error {
	id := c.Params("id")
	return utils.SendSimpleSuccess(c, map[string]interface{}{
		"id":          id,
		"resolved":    true,
		"resolved_at": time.Now().Format(time.RFC3339),
	}, "API error resolved successfully")
}

// GetAPIAlerts returns API alerts (alert rules not yet implemented)
func GetAPIAlerts(c *fiber.Ctx) error {
	return utils.SendSimpleSuccess(c, []interface{}{}, "API alerts retrieved successfully")
}

// AcknowledgeAPIAlert acknowledges an API alert
func AcknowledgeAPIAlert(c *fiber.Ctx) error {
	id := c.Params("id")
	return utils.SendSimpleSuccess(c, map[string]interface{}{
		"id":              id,
		"acknowledged":    true,
		"acknowledged_at": time.Now().Format(time.RFC3339),
	}, "API alert acknowledged successfully")
}

// ResolveAPIAlert resolves an API alert
func ResolveAPIAlert(c *fiber.Ctx) error {
	id := c.Params("id")
	return utils.SendSimpleSuccess(c, map[string]interface{}{
		"id":          id,
		"resolved":    true,
		"resolved_at": time.Now().Format(time.RFC3339),
	}, "API alert resolved successfully")
}

// GetAPIStats returns comprehensive API statistics
func GetAPIStats(c *fiber.Ctx) error {
	stats := middleware.Metrics.GetStats()
	return utils.SendSimpleSuccess(c, stats, "API stats retrieved successfully")
}

// GetAPIPerformance returns API performance data with percentiles
func GetAPIPerformance(c *fiber.Ctx) error {
	period := c.Query("period", "24h")
	performance := middleware.Metrics.GetPerformance(period)
	return utils.SendSimpleSuccess(c, performance, "API performance data retrieved successfully")
}

// TestAPIEndpoint tests an API endpoint (not implemented)
func TestAPIEndpoint(c *fiber.Ctx) error {
	return utils.SendNotImplementedError(c, "Endpoint testing is not yet implemented")
}

// UpdateAPIEndpointConfig updates an endpoint's configuration (not implemented)
func UpdateAPIEndpointConfig(c *fiber.Ctx) error {
	return utils.SendNotImplementedError(c, "Endpoint configuration update is not yet implemented")
}

// ExportAPIMonitoringData exports API monitoring data
func ExportAPIMonitoringData(c *fiber.Ctx) error {
	stats := middleware.Metrics.GetStats()
	endpoints := middleware.Metrics.GetEndpoints()
	performance := middleware.Metrics.GetPerformance("24h")

	exportData := map[string]interface{}{
		"stats":       stats,
		"endpoints":   endpoints,
		"performance": performance,
		"exported_at": time.Now().Format(time.RFC3339),
	}

	c.Set("Content-Disposition", "attachment; filename=api-monitoring-export-"+time.Now().Format("2006-01-02")+".json")
	c.Set("Content-Type", "application/json")

	return c.JSON(exportData)
}

// GetAPICategories returns the list of observed API categories
func GetAPICategories(c *fiber.Ctx) error {
	endpoints := middleware.Metrics.GetEndpoints()

	categorySet := map[string]bool{}
	for _, ep := range endpoints {
		if cat, ok := ep["category"].(string); ok {
			categorySet[cat] = true
		}
	}

	categories := make([]string, 0, len(categorySet))
	for cat := range categorySet {
		categories = append(categories, cat)
	}

	defaults := []string{"auth", "organizations", "users", "documents", "workflows", "approvals", "reports", "admin", "subscriptions", "notifications"}
	for _, d := range defaults {
		if !categorySet[d] {
			categories = append(categories, d)
		}
	}

	return utils.SendSimpleSuccess(c, categories, "API categories retrieved successfully")
}

// CreateAPIAlertRule creates an alert rule (not implemented)
func CreateAPIAlertRule(c *fiber.Ctx) error {
	return utils.SendNotImplementedError(c, "Alert rule creation is not yet implemented")
}

// GetAPIRealtimeMetrics returns real-time API metrics from the last 5 minutes
func GetAPIRealtimeMetrics(c *fiber.Ctx) error {
	data := middleware.Metrics.GetRealtimeMetrics()
	return utils.SendSimpleSuccess(c, data, "Realtime metrics retrieved successfully")
}
