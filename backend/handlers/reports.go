package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/liyali/liyali-gateway/logging"
	"github.com/liyali/liyali-gateway/middleware"
	"github.com/liyali/liyali-gateway/services"
	"github.com/liyali/liyali-gateway/utils"
)

type ReportsHandler struct {
	reportsService *services.ReportsService
}

func NewReportsHandler(reportsService *services.ReportsService) *ReportsHandler {
	return &ReportsHandler{
		reportsService: reportsService,
	}
}

// GetSystemStatistics handles GET /api/admin/reports/system-stats
func (h *ReportsHandler) GetSystemStatistics(c *fiber.Ctx) error {
	logger := logging.FromContext(c)
	logger.Info("get_system_statistics_request")

	// Get organization context from tenant middleware
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Organization context required",
			"error":   err.Error(),
		})
	}

	// Verify admin role
	if tenant.UserRole != "admin" && tenant.UserRole != "superadmin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Admin access required",
		})
	}

	// Parse date range query parameters
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	// Add query parameters to context
	logging.AddFieldsToRequest(c, map[string]interface{}{
		"operation":       "get_system_statistics",
		"organization_id": tenant.OrganizationID,
		"start_date":      startDate,
		"end_date":        endDate,
	})

	// Get statistics from service
	stats, err := h.reportsService.GetSystemStatistics(
		c.Context(),
		tenant.OrganizationID,
		startDate,
		endDate,
	)
	if err != nil {
		logging.LogError(c, err, "failed_to_get_system_statistics")
		return utils.SendInternalError(c, "Failed to fetch system statistics", err)
	}

	logger.Info("system_statistics_retrieved")
	return c.JSON(stats)
}

// GetApprovalMetrics handles GET /api/admin/reports/approval-metrics
func (h *ReportsHandler) GetApprovalMetrics(c *fiber.Ctx) error {
	logger := logging.FromContext(c)
	logger.Info("get_approval_metrics_request")

	// Get organization context from tenant middleware
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Organization context required",
			"error":   err.Error(),
		})
	}

	// Verify admin role
	if tenant.UserRole != "admin" && tenant.UserRole != "superadmin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Admin access required",
		})
	}

	// Parse date range query parameters
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	// Add query parameters to context
	logging.AddFieldsToRequest(c, map[string]interface{}{
		"operation":       "get_approval_metrics",
		"organization_id": tenant.OrganizationID,
		"start_date":      startDate,
		"end_date":        endDate,
	})

	// Get metrics from service
	metrics, err := h.reportsService.GetApprovalMetrics(
		c.Context(),
		tenant.OrganizationID,
		startDate,
		endDate,
	)
	if err != nil {
		logging.LogError(c, err, "failed_to_get_approval_metrics")
		return utils.SendInternalError(c, "Failed to fetch approval metrics", err)
	}

	logger.Info("approval_metrics_retrieved")
	return c.JSON(metrics)
}

// GetUserActivityMetrics handles GET /api/admin/reports/user-activity
func (h *ReportsHandler) GetUserActivityMetrics(c *fiber.Ctx) error {
	logger := logging.FromContext(c)
	logger.Info("get_user_activity_metrics_request")

	// Get organization context from tenant middleware
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Organization context required",
			"error":   err.Error(),
		})
	}

	// Verify admin role
	if tenant.UserRole != "admin" && tenant.UserRole != "superadmin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Admin access required",
		})
	}

	// Add query parameters to context
	logging.AddFieldsToRequest(c, map[string]interface{}{
		"operation":       "get_user_activity_metrics",
		"organization_id": tenant.OrganizationID,
	})

	// Get metrics from service
	metrics, err := h.reportsService.GetUserActivityMetrics(
		c.Context(),
		tenant.OrganizationID,
	)
	if err != nil {
		logging.LogError(c, err, "failed_to_get_user_activity_metrics")
		return utils.SendInternalError(c, "Failed to fetch user activity metrics", err)
	}

	logger.Info("user_activity_metrics_retrieved")
	return c.JSON(metrics)
}

// GetAnalyticsDashboard handles GET /api/admin/reports/analytics
func (h *ReportsHandler) GetAnalyticsDashboard(c *fiber.Ctx) error {
	logger := logging.FromContext(c)
	logger.Info("get_analytics_dashboard_request")

	// Get organization context from tenant middleware
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Organization context required",
			"error":   err.Error(),
		})
	}

	// Verify admin role
	if tenant.UserRole != "admin" && tenant.UserRole != "superadmin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Admin access required",
		})
	}

	// Parse date range query parameters
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	// Add query parameters to context
	logging.AddFieldsToRequest(c, map[string]interface{}{
		"operation":       "get_analytics_dashboard",
		"organization_id": tenant.OrganizationID,
		"start_date":      startDate,
		"end_date":        endDate,
	})

	// Get analytics from service
	analytics, err := h.reportsService.GetAnalyticsDashboard(
		c.Context(),
		tenant.OrganizationID,
		startDate,
		endDate,
	)
	if err != nil {
		logging.LogError(c, err, "failed_to_get_analytics_dashboard")
		return utils.SendInternalError(c, "Failed to fetch analytics dashboard", err)
	}

	logger.Info("analytics_dashboard_retrieved")
	return c.JSON(analytics)
}
