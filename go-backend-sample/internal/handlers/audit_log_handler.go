package handlers

import (
	"strconv"

	"github.com/cozyCodr/liyali-gateway/internal/middleware"
	"github.com/cozyCodr/liyali-gateway/internal/repository"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type AuditLogHandler struct {
	auditLogRepo repository.AuditLogRepository
}

func NewAuditLogHandler(auditLogRepo repository.AuditLogRepository) *AuditLogHandler {
	return &AuditLogHandler{
		auditLogRepo: auditLogRepo,
	}
}

// GetAuditLogs retrieves audit logs with optional filtering
// GET /api/audit-logs
func (h *AuditLogHandler) GetAuditLogs(c fiber.Ctx) error {
	_, ok := middleware.GetUserID(c)
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

	// Only admins and managers can view all audit logs
	if userRole != "ADMIN" && userRole != "MANAGER" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "insufficient permissions to view audit logs",
		})
	}

	// Parse query parameters
	limitStr := c.Query("limit", "50")
	offsetStr := c.Query("offset", "0")
	resourceType := c.Query("resource_type")
	action := c.Query("action")
	userIDFilter := c.Query("user_id")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	var auditLogs interface{}
	var totalCount int64

	// Apply filters based on query parameters
	if userIDFilter != "" {
		// Filter by user
		filterUserID, err := uuid.Parse(userIDFilter)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid user ID format",
			})
		}

		auditLogs, err = h.auditLogRepo.ListAuditLogsByUser(c.Context(), filterUserID, int32(limit), int32(offset))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to retrieve audit logs",
			})
		}

		totalCount, _ = h.auditLogRepo.CountAuditLogsByUser(c.Context(), filterUserID)

	} else if resourceType != "" {
		// Filter by resource type
		auditLogs, err = h.auditLogRepo.ListAuditLogsByResourceType(c.Context(), resourceType, int32(limit), int32(offset))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to retrieve audit logs",
			})
		}

		totalCount, _ = h.auditLogRepo.CountAuditLogs(c.Context())

	} else if action != "" {
		// Filter by action
		auditLogs, err = h.auditLogRepo.ListAuditLogsByAction(c.Context(), action, int32(limit), int32(offset))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to retrieve audit logs",
			})
		}

		totalCount, _ = h.auditLogRepo.CountAuditLogs(c.Context())

	} else {
		// No filter, get all audit logs
		auditLogs, err = h.auditLogRepo.ListAuditLogs(c.Context(), int32(limit), int32(offset))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to retrieve audit logs",
			})
		}

		totalCount, _ = h.auditLogRepo.CountAuditLogs(c.Context())
	}

	return c.JSON(fiber.Map{
		"audit_logs": auditLogs,
		"total":      totalCount,
		"limit":      limit,
		"offset":     offset,
	})
}

// GetAuditLogByID retrieves a specific audit log entry
// GET /api/audit-logs/:id
func (h *AuditLogHandler) GetAuditLogByID(c fiber.Ctx) error {
	_, ok := middleware.GetUserID(c)
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

	// Only admins and managers can view audit logs
	if userRole != "ADMIN" && userRole != "MANAGER" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "insufficient permissions to view audit logs",
		})
	}

	// Parse audit log ID
	auditLogID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid audit log ID",
		})
	}

	auditLog, err := h.auditLogRepo.GetAuditLogByID(c.Context(), auditLogID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "audit log not found",
		})
	}

	return c.JSON(auditLog)
}

// GetAuditLogsByResource retrieves audit logs for a specific resource
// GET /api/audit-logs/resource/:resource_type/:resource_id
func (h *AuditLogHandler) GetAuditLogsByResource(c fiber.Ctx) error {
	_, ok := middleware.GetUserID(c)
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

	// Only admins and managers can view audit logs
	if userRole != "ADMIN" && userRole != "MANAGER" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "insufficient permissions to view audit logs",
		})
	}

	// Parse parameters
	resourceType := c.Params("resource_type")
	resourceID, err := uuid.Parse(c.Params("resource_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid resource ID",
		})
	}

	// Parse query parameters
	limitStr := c.Query("limit", "50")
	offsetStr := c.Query("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	auditLogs, err := h.auditLogRepo.ListAuditLogsByResource(c.Context(), resourceType, resourceID, int32(limit), int32(offset))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to retrieve audit logs",
		})
	}

	totalCount, _ := h.auditLogRepo.CountAuditLogsByResource(c.Context(), resourceType, resourceID)

	return c.JSON(fiber.Map{
		"audit_logs":    auditLogs,
		"resource_type": resourceType,
		"resource_id":   resourceID,
		"total":         totalCount,
		"limit":         limit,
		"offset":        offset,
	})
}

// GetMyAuditLogs retrieves audit logs for the current user's actions
// GET /api/audit-logs/my
func (h *AuditLogHandler) GetMyAuditLogs(c fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "user not authenticated",
		})
	}

	// Parse query parameters
	limitStr := c.Query("limit", "50")
	offsetStr := c.Query("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	auditLogs, err := h.auditLogRepo.ListAuditLogsByUser(c.Context(), userID, int32(limit), int32(offset))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to retrieve audit logs",
		})
	}

	totalCount, _ := h.auditLogRepo.CountAuditLogsByUser(c.Context(), userID)

	return c.JSON(fiber.Map{
		"audit_logs": auditLogs,
		"total":      totalCount,
		"limit":      limit,
		"offset":     offset,
	})
}
