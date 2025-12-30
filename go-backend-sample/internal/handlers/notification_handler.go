package handlers

import (
	"strconv"

	"github.com/cozyCodr/liyali-gateway/internal/middleware"
	"github.com/cozyCodr/liyali-gateway/internal/repository"
	"github.com/cozyCodr/liyali-gateway/internal/utils"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type NotificationHandler struct {
	notificationRepo repository.NotificationRepository
}

func NewNotificationHandler(notificationRepo repository.NotificationRepository) *NotificationHandler {
	return &NotificationHandler{
		notificationRepo: notificationRepo,
	}
}

// GetNotifications retrieves all notifications for the current user
// GET /api/notifications
func (h *NotificationHandler) GetNotifications(c fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "user not authenticated",
		})
	}

	// Parse query parameters
	limitStr := c.Query("limit", "20")
	offsetStr := c.Query("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	notifications, err := h.notificationRepo.ListNotificationsByUser(c.Context(), userID, int32(limit), int32(offset))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to retrieve notifications",
		})
	}

	// Get unread count
	unreadCount, err := h.notificationRepo.CountUnreadNotificationsByUser(c.Context(), userID)
	if err != nil {
		unreadCount = 0
	}

	return c.JSON(fiber.Map{
		"notifications": notifications,
		"total":         len(notifications),
		"unread_count":  unreadCount,
		"limit":         limit,
		"offset":        offset,
	})
}

// GetUnreadNotifications retrieves unread notifications for the current user
// GET /api/notifications/unread
func (h *NotificationHandler) GetUnreadNotifications(c fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "user not authenticated",
		})
	}

	// Parse query parameters
	limitStr := c.Query("limit", "20")
	offsetStr := c.Query("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	notifications, err := h.notificationRepo.ListUnreadNotificationsByUser(c.Context(), userID, int32(limit), int32(offset))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to retrieve unread notifications",
		})
	}

	// Get unread count
	unreadCount, err := h.notificationRepo.CountUnreadNotificationsByUser(c.Context(), userID)
	if err != nil {
		unreadCount = 0
	}

	return c.JSON(fiber.Map{
		"notifications": notifications,
		"total":         len(notifications),
		"unread_count":  unreadCount,
		"limit":         limit,
		"offset":        offset,
	})
}

// GetNotificationByID retrieves a specific notification
// GET /api/notifications/:id
func (h *NotificationHandler) GetNotificationByID(c fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "user not authenticated",
		})
	}

	// Parse notification ID
	notificationID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid notification ID",
		})
	}

	notification, err := h.notificationRepo.GetNotificationByID(c.Context(), notificationID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "notification not found",
		})
	}

	// Verify ownership
	if utils.PgtypeToUUID(notification.UserID) != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "you are not authorized to view this notification",
		})
	}

	return c.JSON(notification)
}

// MarkAsRead marks a notification as read
// POST /api/notifications/:id/read
func (h *NotificationHandler) MarkAsRead(c fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "user not authenticated",
		})
	}

	// Parse notification ID
	notificationID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid notification ID",
		})
	}

	// Get the notification first to verify ownership
	notification, err := h.notificationRepo.GetNotificationByID(c.Context(), notificationID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "notification not found",
		})
	}

	// Verify ownership
	if utils.PgtypeToUUID(notification.UserID) != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "you are not authorized to modify this notification",
		})
	}

	// Mark as read
	updatedNotification, err := h.notificationRepo.MarkNotificationAsRead(c.Context(), notificationID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to mark notification as read",
		})
	}

	return c.JSON(updatedNotification)
}

// MarkAllAsRead marks all notifications as read for the current user
// POST /api/notifications/read-all
func (h *NotificationHandler) MarkAllAsRead(c fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "user not authenticated",
		})
	}

	err := h.notificationRepo.MarkAllNotificationsAsRead(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to mark all notifications as read",
		})
	}

	return c.JSON(fiber.Map{
		"message": "all notifications marked as read",
	})
}

// DeleteNotification deletes a notification
// DELETE /api/notifications/:id
func (h *NotificationHandler) DeleteNotification(c fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "user not authenticated",
		})
	}

	// Parse notification ID
	notificationID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid notification ID",
		})
	}

	// Get the notification first to verify ownership
	notification, err := h.notificationRepo.GetNotificationByID(c.Context(), notificationID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "notification not found",
		})
	}

	// Verify ownership
	if utils.PgtypeToUUID(notification.UserID) != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "you are not authorized to delete this notification",
		})
	}

	// Delete the notification
	err = h.notificationRepo.DeleteNotification(c.Context(), notificationID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to delete notification",
		})
	}

	return c.JSON(fiber.Map{
		"message": "notification deleted successfully",
	})
}

// GetUnreadCount retrieves the unread notification count
// GET /api/notifications/unread/count
func (h *NotificationHandler) GetUnreadCount(c fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "user not authenticated",
		})
	}

	count, err := h.notificationRepo.CountUnreadNotificationsByUser(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get unread count",
		})
	}

	return c.JSON(fiber.Map{
		"unread_count": count,
	})
}
