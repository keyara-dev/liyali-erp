package handlers

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/services"
	"github.com/liyali/liyali-gateway/utils"
)

// GetNotifications returns notifications for the authenticated user
func GetNotifications(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(string)
	if !ok {
		return utils.SendUnauthorizedError(c, "User not authenticated")
	}

	// Parse query parameters
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)
	notifType := c.Query("type")
	unreadOnly := c.Query("unreadOnly") == "true"

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	notificationService := services.NewNotificationService(config.DB)

	var notifications []interface{}
	var total int64

	if unreadOnly {
		// Get only unread notifications
		notifs, err := notificationService.GetPendingNotifications(userID)
		if err != nil {
			log.Printf("Error getting pending notifications: %v", err)
			return utils.SendInternalError(c, "Failed to fetch notifications", err)
		}
		
		// Convert to interface slice for response
		for _, notif := range notifs {
			notifications = append(notifications, map[string]interface{}{
				"id":           notif.ID,
				"type":         notif.Type,
				"documentId":   notif.DocumentID,
				"documentType": notif.DocumentType,
				"subject":      notif.Subject,
				"body":         notif.Body,
				"sent":         notif.Sent,
				"sentAt":       notif.SentAt,
				"createdAt":    notif.CreatedAt,
				"updatedAt":    notif.UpdatedAt,
			})
		}
		total = int64(len(notifications))
	} else if notifType != "" {
		// Get notifications by type
		notifs, err := notificationService.GetNotificationsByType(userID, notifType)
		if err != nil {
			log.Printf("Error getting notifications by type: %v", err)
			return utils.SendInternalError(c, "Failed to fetch notifications", err)
		}
		
		// Convert to interface slice for response
		for _, notif := range notifs {
			notifications = append(notifications, map[string]interface{}{
				"id":           notif.ID,
				"type":         notif.Type,
				"documentId":   notif.DocumentID,
				"documentType": notif.DocumentType,
				"subject":      notif.Subject,
				"body":         notif.Body,
				"sent":         notif.Sent,
				"sentAt":       notif.SentAt,
				"createdAt":    notif.CreatedAt,
				"updatedAt":    notif.UpdatedAt,
			})
		}
		total = int64(len(notifications))
	} else {
		// Get notifications since a specific time (for pagination)
		since := time.Now().AddDate(0, -1, 0) // Last month by default
		if sinceStr := c.Query("since"); sinceStr != "" {
			if parsedSince, err := time.Parse(time.RFC3339, sinceStr); err == nil {
				since = parsedSince
			}
		}

		notifs, err := notificationService.GetNotificationsSince(userID, since)
		if err != nil {
			log.Printf("Error getting notifications since: %v", err)
			return utils.SendInternalError(c, "Failed to fetch notifications", err)
		}

		// Apply pagination
		start := (page - 1) * limit
		end := start + limit
		total = int64(len(notifs))

		if start < len(notifs) {
			if end > len(notifs) {
				end = len(notifs)
			}
			
			// Convert to interface slice for response
			for _, notif := range notifs[start:end] {
				notifications = append(notifications, map[string]interface{}{
					"id":           notif.ID,
					"type":         notif.Type,
					"documentId":   notif.DocumentID,
					"documentType": notif.DocumentType,
					"subject":      notif.Subject,
					"body":         notif.Body,
					"sent":         notif.Sent,
					"sentAt":       notif.SentAt,
					"createdAt":    notif.CreatedAt,
					"updatedAt":    notif.UpdatedAt,
				})
			}
		}
	}

	return utils.SendPaginatedSuccess(c, notifications, "Notifications retrieved successfully", page, limit, total)
}

// GetNotification returns a specific notification
func GetNotification(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(string)
	if !ok {
		return utils.SendUnauthorizedError(c, "User not authenticated")
	}

	notificationID := c.Params("id")
	if notificationID == "" {
		return utils.SendBadRequestError(c, "Notification ID is required")
	}

	// Get notification from database
	var notification interface{}
	if err := config.DB.Where("id = ? AND recipient_id = ?", notificationID, userID).
		First(&notification).Error; err != nil {
		return utils.SendNotFoundError(c, "Notification")
	}

	return utils.SendSimpleSuccess(c, notification, "Notification retrieved successfully")
}

// MarkNotificationAsRead marks a notification as read
func MarkNotificationAsRead(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(string)
	if !ok {
		return utils.SendUnauthorizedError(c, "User not authenticated")
	}

	notificationID := c.Params("id")
	if notificationID == "" {
		return utils.SendBadRequestError(c, "Notification ID is required")
	}

	// Verify notification belongs to user
	var count int64
	if err := config.DB.Model(&struct{}{}).
		Where("id = ? AND recipient_id = ?", notificationID, userID).
		Count(&count).Error; err != nil || count == 0 {
		return utils.SendNotFoundError(c, "Notification")
	}

	notificationService := services.NewNotificationService(config.DB)
	if err := notificationService.MarkAsRead(notificationID); err != nil {
		log.Printf("Error marking notification as read: %v", err)
		return utils.SendInternalError(c, "Failed to mark notification as read", err)
	}

	return utils.SendSimpleSuccess(c, nil, "Notification marked as read")
}

// MarkAllNotificationsAsRead marks all notifications as read for the user
func MarkAllNotificationsAsRead(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(string)
	if !ok {
		return utils.SendUnauthorizedError(c, "User not authenticated")
	}

	notificationService := services.NewNotificationService(config.DB)
	
	// Get all unread notification IDs for the user
	pendingNotifs, err := notificationService.GetPendingNotifications(userID)
	if err != nil {
		log.Printf("Error getting pending notifications: %v", err)
		return utils.SendInternalError(c, "Failed to fetch notifications", err)
	}

	if len(pendingNotifs) == 0 {
		return utils.SendSimpleSuccess(c, map[string]interface{}{
			"markedCount": 0,
		}, "No unread notifications found")
	}

	// Extract notification IDs
	notificationIDs := make([]string, len(pendingNotifs))
	for i, notif := range pendingNotifs {
		notificationIDs[i] = notif.ID
	}

	// Mark all as read
	if err := notificationService.MarkMultipleAsRead(notificationIDs); err != nil {
		log.Printf("Error marking notifications as read: %v", err)
		return utils.SendInternalError(c, "Failed to mark notifications as read", err)
	}

	return utils.SendSimpleSuccess(c, map[string]interface{}{
		"markedCount": len(notificationIDs),
	}, "All notifications marked as read")
}

// GetNotificationStats returns notification statistics for the user
func GetNotificationStats(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(string)
	if !ok {
		return utils.SendUnauthorizedError(c, "User not authenticated")
	}

	notificationService := services.NewNotificationService(config.DB)
	stats, err := notificationService.GetNotificationStats(userID)
	if err != nil {
		log.Printf("Error getting notification stats: %v", err)
		return utils.SendInternalError(c, "Failed to fetch notification statistics", err)
	}

	return utils.SendSimpleSuccess(c, stats, "Notification statistics retrieved successfully")
}

// DeleteNotification deletes a notification
func DeleteNotification(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(string)
	if !ok {
		return utils.SendUnauthorizedError(c, "User not authenticated")
	}

	notificationID := c.Params("id")
	if notificationID == "" {
		return utils.SendBadRequestError(c, "Notification ID is required")
	}

	// Verify notification belongs to user
	var count int64
	if err := config.DB.Model(&struct{}{}).
		Where("id = ? AND recipient_id = ?", notificationID, userID).
		Count(&count).Error; err != nil || count == 0 {
		return utils.SendNotFoundError(c, "Notification")
	}

	notificationService := services.NewNotificationService(config.DB)
	if err := notificationService.DeleteNotification(notificationID); err != nil {
		log.Printf("Error deleting notification: %v", err)
		return utils.SendInternalError(c, "Failed to delete notification", err)
	}

	return utils.SendSimpleSuccess(c, nil, "Notification deleted successfully")
}