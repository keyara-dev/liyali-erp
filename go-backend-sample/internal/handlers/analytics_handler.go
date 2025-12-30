package handlers

import (
	"strconv"

	"github.com/cozyCodr/liyali-gateway/internal/middleware"
	"github.com/cozyCodr/liyali-gateway/internal/services"
	"github.com/gofiber/fiber/v3"
)

type AnalyticsHandler struct {
	analyticsService *services.AnalyticsService
}

func NewAnalyticsHandler(analyticsService *services.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsService: analyticsService,
	}
}

// GetDashboardMetrics retrieves comprehensive dashboard metrics
// GET /api/analytics/metrics
func (h *AnalyticsHandler) GetDashboardMetrics(c fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "user not authenticated",
		})
	}

	userRole, ok := middleware.GetUserRole(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "user role not found",
		})
	}

	metrics, err := h.analyticsService.GetDashboardMetrics(c.Context(), userID, userRole)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to retrieve dashboard metrics",
		})
	}

	return c.JSON(metrics)
}

// GetTrendData retrieves trend data for the specified number of days
// GET /api/analytics/trends?days=7
func (h *AnalyticsHandler) GetTrendData(c fiber.Ctx) error {
	_, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "user not authenticated",
		})
	}

	// Parse days parameter
	daysStr := c.Query("days", "7")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 {
		days = 7
	}
	if days > 90 {
		days = 90
	}

	trends, err := h.analyticsService.GetTrendData(c.Context(), int32(days))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to retrieve trend data",
		})
	}

	return c.JSON(fiber.Map{
		"trends": trends,
		"days":   days,
	})
}

// GetBottlenecks retrieves workflow bottleneck analysis
// GET /api/analytics/bottlenecks
func (h *AnalyticsHandler) GetBottlenecks(c fiber.Ctx) error {
	_, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "user not authenticated",
		})
	}

	bottlenecks, err := h.analyticsService.GetBottlenecks(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to retrieve bottleneck analysis",
		})
	}

	return c.JSON(fiber.Map{
		"bottlenecks": bottlenecks,
		"count":       len(bottlenecks),
	})
}
