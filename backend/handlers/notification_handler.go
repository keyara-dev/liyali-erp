package handlers

import (
	"log"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/services"
	"github.com/liyali/liyali-gateway/types"
	"github.com/liyali/liyali-gateway/utils"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type NotificationHandler struct {
	validate *validator.Validate
}

func NewNotificationHandler() *NotificationHandler {
	return &NotificationHandler{
		validate: validator.New(),
	}
}

// Request/Response Types
type MarkAsReadRequest struct {
	NotificationIDs []string `json:"notificationIds" validate:"required,min=1"`
}

type NotificationStatsResponse struct {
	Pending int64 `json:"pending"`
	Read    int64 `json:"read"`
	Total   int64 `json:"total"`
}

// NotificationPreferencesRequest represents the request body for updating preferences
type NotificationPreferencesRequest struct {
	EmailEnabled           bool `json:"emailEnabled"`
	PushEnabled            bool `json:"pushEnabled"`
	InAppEnabled           bool `json:"inAppEnabled"`
	NotifyTaskAssigned     bool `json:"notifyTaskAssigned"`
	NotifyTaskReassigned   bool `json:"notifyTaskReassigned"`
	NotifyTaskApproved     bool `json:"notifyTaskApproved"`
	NotifyTaskRejected     bool `json:"notifyTaskRejected"`
	NotifyWorkflowComplete bool `json:"notifyWorkflowComplete"`
	NotifyApprovalOverdue  bool `json:"notifyApprovalOverdue"`
	NotifyCommentsAdded    bool `json:"notifyCommentsAdded"`
	QuietHoursEnabled      bool `json:"quietHoursEnabled"`
	QuietHoursStart        int  `json:"quietHoursStart"`
	QuietHoursEnd          int  `json:"quietHoursEnd"`
}

// GetNotifications retrieves notifications for the current user with pagination and filtering
func (h *NotificationHandler) GetNotifications(c *fiber.Ctx) error {
	db := config.DB
	
	// Safely get organizationID
	organizationIDRaw := c.Locals("organizationID")
	if organizationIDRaw == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Organization context required",
		})
	}
	organizationID, ok := organizationIDRaw.(string)
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid organization context",
		})
	}
	
	// Safely get userID
	userIDRaw := c.Locals("userID")
	if userIDRaw == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "User context required",
		})
	}
	userID, ok := userIDRaw.(string)
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user context",
		})
	}

	// Extract query parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	notificationType := c.Query("type", "")
	unreadOnly := c.Query("unread_only", "false") == "true"

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Build query
	query := db.Where("organization_id = ? AND recipient_id = ?", organizationID, userID)
	
	if notificationType != "" {
		query = query.Where("type = ?", notificationType)
	}
	
	if unreadOnly {
		query = query.Where("is_read = ?", false)
	}

	// Get total count
	var total int64
	query.Model(&models.Notification{}).Count(&total)

	// Get notifications with pagination
	var notifications []models.Notification
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&notifications).Error; err != nil {
		log.Printf("Error fetching notifications: %v", err)
		return utils.SendInternalError(c, "Failed to fetch notifications", err)
	}

	// Transform notifications for frontend compatibility
	var transformedNotifications []map[string]interface{}
	for _, notification := range notifications {
		transformed := map[string]interface{}{
			"id":             notification.ID,
			"type":           notification.Type,
			"subject":        notification.Subject,
			"body":           notification.Body,
			"documentId":     notification.DocumentID,
			"documentType":   notification.DocumentType,
			"entityId":       notification.DocumentID,       // Alias for backward compatibility
			"entityType":     notification.DocumentType,     // Alias for backward compatibility
			"isRead":         notification.IsRead,           // Use actual IsRead field
			"readAt":         notification.ReadAt,           // Use actual ReadAt field
			"createdAt":      notification.CreatedAt,
			"updatedAt":      notification.UpdatedAt,
			"importance":     "MEDIUM",                       // Default importance
			"message":        notification.Body,             // Alias for filtering
		}

		// Set importance based on notification type
		switch notification.Type {
		case "approval_required":
			transformed["importance"] = "HIGH"
		case "document_rejected":
			transformed["importance"] = "HIGH"
		case "document_approved":
			transformed["importance"] = "MEDIUM"
		default:
			transformed["importance"] = "LOW"
		}

		// Get document number for display
		if notification.DocumentID != "" && notification.DocumentType != "" {
			documentNumber := h.getDocumentNumber(db, notification.DocumentType, notification.DocumentID)
			if documentNumber != "" {
				transformed["entityNumber"] = documentNumber
			}
		}

		transformedNotifications = append(transformedNotifications, transformed)
	}

	return utils.SendPaginatedSuccess(c, transformedNotifications, "Notifications retrieved successfully", page, limit, total)
}

// GetNotificationStats returns notification statistics for the current user
func (h *NotificationHandler) GetNotificationStats(c *fiber.Ctx) error {
	db := config.DB
	
	// Safely get organizationID
	organizationIDRaw := c.Locals("organizationID")
	if organizationIDRaw == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Organization context required",
		})
	}
	organizationID, ok := organizationIDRaw.(string)
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid organization context",
		})
	}
	
	// Safely get userID
	userIDRaw := c.Locals("userID")
	if userIDRaw == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "User context required",
		})
	}
	userID, ok := userIDRaw.(string)
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user context",
		})
	}

	var pendingCount int64
	var readCount int64
	var totalCount int64

	// Get pending (unread) notifications count
	db.Model(&models.Notification{}).
		Where("organization_id = ? AND recipient_id = ? AND is_read = ?", organizationID, userID, false).
		Count(&pendingCount)

	// Get read notifications count
	db.Model(&models.Notification{}).
		Where("organization_id = ? AND recipient_id = ? AND is_read = ?", organizationID, userID, true).
		Count(&readCount)

	// Get total notifications count
	db.Model(&models.Notification{}).
		Where("organization_id = ? AND recipient_id = ?", organizationID, userID).
		Count(&totalCount)

	stats := NotificationStatsResponse{
		Pending: pendingCount,
		Read:    readCount,
		Total:   totalCount,
	}

	return utils.SendSimpleSuccess(c, stats, "Notification statistics retrieved successfully")
}

// MarkAsRead marks one or more notifications as read
func (h *NotificationHandler) MarkAsRead(c *fiber.Ctx) error {
	db := config.DB
	organizationID := c.Locals("organizationID").(string)
	userID := c.Locals("userID").(string)

	var req MarkAsReadRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{
			Error:   "Invalid request body",
			Message: "Failed to parse mark as read request",
		})
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{
			Error:   "Validation failed",
			Message: err.Error(),
		})
	}

	// Get notification service
	notificationService := services.NewNotificationService(db)

	// Verify notifications belong to the user and organization
	var notifications []models.Notification
	if err := db.Where("id IN ? AND organization_id = ? AND recipient_id = ?", 
		req.NotificationIDs, organizationID, userID).Find(&notifications).Error; err != nil {
		log.Printf("Error fetching notifications for mark as read: %v", err)
		return utils.SendInternalError(c, "Failed to fetch notifications", err)
	}

	if len(notifications) != len(req.NotificationIDs) {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{
			Error:   "Invalid notifications",
			Message: "Some notifications not found or access denied",
		})
	}

	// Mark notifications as read
	if err := notificationService.MarkMultipleAsRead(req.NotificationIDs); err != nil {
		log.Printf("Error marking notifications as read: %v", err)
		return utils.SendInternalError(c, "Failed to mark notifications as read", err)
	}

	return c.Status(fiber.StatusOK).JSON(types.SuccessResponse{
		Message: "Notifications marked as read successfully",
		Data:    map[string]interface{}{"markedCount": len(req.NotificationIDs)},
	})
}

// MarkAllAsRead marks all unread notifications as read for the current user
func (h *NotificationHandler) MarkAllAsRead(c *fiber.Ctx) error {
	db := config.DB
	organizationID := c.Locals("organizationID").(string)
	userID := c.Locals("userID").(string)

	// Get all unread notification IDs for the user
	var notificationIDs []string
	if err := db.Model(&models.Notification{}).
		Where("organization_id = ? AND recipient_id = ? AND is_read = ?", organizationID, userID, false).
		Pluck("id", &notificationIDs).Error; err != nil {
		log.Printf("Error fetching unread notification IDs: %v", err)
		return utils.SendInternalError(c, "Failed to fetch unread notifications", err)
	}

	if len(notificationIDs) == 0 {
		return c.Status(fiber.StatusOK).JSON(types.SuccessResponse{
			Message: "No unread notifications to mark as read",
			Data:    map[string]interface{}{"markedCount": 0},
		})
	}

	// Get notification service
	notificationService := services.NewNotificationService(db)

	// Mark all as read
	if err := notificationService.MarkMultipleAsRead(notificationIDs); err != nil {
		log.Printf("Error marking all notifications as read: %v", err)
		return utils.SendInternalError(c, "Failed to mark all notifications as read", err)
	}

	return c.Status(fiber.StatusOK).JSON(types.SuccessResponse{
		Message: "All notifications marked as read successfully",
		Data:    map[string]interface{}{"markedCount": len(notificationIDs)},
	})
}

// DeleteNotification deletes a notification
func (h *NotificationHandler) DeleteNotification(c *fiber.Ctx) error {
	notificationID := c.Params("id")
	if notificationID == "" {
		return utils.SendBadRequestError(c, "Notification ID is required")
	}

	db := config.DB
	organizationID := c.Locals("organizationID").(string)
	userID := c.Locals("userID").(string)

	// Verify notification belongs to the user
	var notification models.Notification
	if err := db.Where("id = ? AND organization_id = ? AND recipient_id = ?", 
		notificationID, organizationID, userID).First(&notification).Error; err != nil {
		return utils.SendNotFoundError(c, "Notification not found or access denied")
	}

	// Get notification service
	notificationService := services.NewNotificationService(db)

	// Delete notification
	if err := notificationService.DeleteNotification(notificationID); err != nil {
		log.Printf("Error deleting notification: %v", err)
		return utils.SendInternalError(c, "Failed to delete notification", err)
	}

	return c.Status(fiber.StatusOK).JSON(types.SuccessResponse{
		Message: "Notification deleted successfully",
		Data:    map[string]interface{}{"deletedId": notificationID},
	})
}

// GetRecentNotifications returns the most recent notifications for header display
func (h *NotificationHandler) GetRecentNotifications(c *fiber.Ctx) error {
	db := config.DB
	
	// Safely get organizationID
	organizationIDRaw := c.Locals("organizationID")
	if organizationIDRaw == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Organization context required",
		})
	}
	organizationID, ok := organizationIDRaw.(string)
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid organization context",
		})
	}
	
	// Safely get userID
	userIDRaw := c.Locals("userID")
	if userIDRaw == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "User context required",
		})
	}
	userID, ok := userIDRaw.(string)
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user context",
		})
	}

	// Get recent notifications (last 10 unread + last 5 read)
	var unreadNotifications []models.Notification
	var readNotifications []models.Notification

	// Get unread notifications
	if err := db.Where("organization_id = ? AND recipient_id = ? AND is_read = ?", 
		organizationID, userID, false).
		Order("created_at DESC").
		Limit(10).
		Find(&unreadNotifications).Error; err != nil {
		log.Printf("Error fetching unread notifications: %v", err)
		return utils.SendInternalError(c, "Failed to fetch unread notifications", err)
	}

	// Get recent read notifications
	if err := db.Where("organization_id = ? AND recipient_id = ? AND is_read = ?", 
		organizationID, userID, true).
		Order("created_at DESC").
		Limit(5).
		Find(&readNotifications).Error; err != nil {
		log.Printf("Error fetching read notifications: %v", err)
		return utils.SendInternalError(c, "Failed to fetch read notifications", err)
	}

	// Combine and transform notifications
	allNotifications := append(unreadNotifications, readNotifications...)
	var transformedNotifications []map[string]interface{}

	for _, notification := range allNotifications {
		transformed := map[string]interface{}{
			"id":           notification.ID,
			"type":         notification.Type,
			"subject":      notification.Subject,
			"body":         notification.Body,
			"message":      notification.Body, // Frontend expects 'message' field
			"documentId":   notification.DocumentID,
			"documentType": notification.DocumentType,
			"entityId":     notification.EntityID,     // Frontend expects entityId
			"entityType":   notification.EntityType,   // Frontend expects entityType
			"isRead":       notification.IsRead,       // Use actual IsRead field
			"readAt":       notification.ReadAt,       // Use actual ReadAt field
			"createdAt":    notification.CreatedAt,
			"updatedAt":    notification.UpdatedAt,
			"importance":   "MEDIUM",
		}

		// Set importance and get document number
		switch notification.Type {
		case "approval_required":
			transformed["importance"] = "HIGH"
		case "document_rejected":
			transformed["importance"] = "HIGH"
		case "document_approved":
			transformed["importance"] = "MEDIUM"
		default:
			transformed["importance"] = "LOW"
		}

		// Get document number for display
		if notification.DocumentID != "" && notification.DocumentType != "" {
			documentNumber := h.getDocumentNumber(db, notification.DocumentType, notification.DocumentID)
			if documentNumber != "" {
				transformed["entityNumber"] = documentNumber
			}
		}

		transformedNotifications = append(transformedNotifications, transformed)
	}

	return utils.SendSimpleSuccess(c, transformedNotifications, "Recent notifications retrieved successfully")
}

// Helper function to get document number for display
func (h *NotificationHandler) getDocumentNumber(db *gorm.DB, documentType, documentID string) string {
	switch documentType {
	case "requisition", "REQUISITION":
		var req models.Requisition
		if err := db.Select("document_number").Where("id = ?", documentID).First(&req).Error; err == nil {
			return req.DocumentNumber
		}
	case "purchase_order", "PURCHASE_ORDER":
		var po models.PurchaseOrder
		if err := db.Select("document_number").Where("id = ?", documentID).First(&po).Error; err == nil {
			return po.DocumentNumber
		}
	case "payment_voucher", "PAYMENT_VOUCHER":
		var pv models.PaymentVoucher
		if err := db.Select("document_number").Where("id = ?", documentID).First(&pv).Error; err == nil {
			return pv.DocumentNumber
		}
	case "grn", "GRN":
		var grn models.GoodsReceivedNote
		if err := db.Select("document_number").Where("id = ?", documentID).First(&grn).Error; err == nil {
			return grn.DocumentNumber
		}
	}
	return ""
}

// GetNotificationPreferences retrieves notification preferences for the current user
func (h *NotificationHandler) GetNotificationPreferences(c *fiber.Ctx) error {
	db := config.DB
	
	// Safely get organizationID
	organizationIDRaw := c.Locals("organizationID")
	if organizationIDRaw == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Organization context required",
		})
	}
	organizationID, ok := organizationIDRaw.(string)
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid organization context",
		})
	}
	
	// Safely get userID
	userIDRaw := c.Locals("userID")
	if userIDRaw == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "User context required",
		})
	}
	userID, ok := userIDRaw.(string)
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user context",
		})
	}

	// Get or create preferences for the user
	var preferences models.NotificationPreferences
	if err := db.Where("user_id = ? AND organization_id = ?", userID, organizationID).
		First(&preferences).Error; err != nil {
		
		// If not found, create default preferences
		if err == gorm.ErrRecordNotFound {
			preferences = models.NotificationPreferences{
				ID:                     uuid.New().String(),
				UserID:                 userID,
				OrganizationID:         organizationID,
				EmailEnabled:           false,
				PushEnabled:            true,
				InAppEnabled:           true,
				NotifyTaskAssigned:     true,
				NotifyTaskReassigned:   true,
				NotifyTaskApproved:     true,
				NotifyTaskRejected:     true,
				NotifyWorkflowComplete: true,
				NotifyApprovalOverdue:  true,
				NotifyCommentsAdded:    false,
				QuietHoursEnabled:      false,
				QuietHoursStart:        22,
				QuietHoursEnd:          8,
				CreatedAt:              time.Now(),
				UpdatedAt:              time.Now(),
			}
			
			if err := db.Create(&preferences).Error; err != nil {
				log.Printf("Error creating default notification preferences: %v", err)
				return utils.SendInternalError(c, "Failed to create notification preferences", err)
			}
		} else {
			log.Printf("Error fetching notification preferences: %v", err)
			return utils.SendInternalError(c, "Failed to fetch notification preferences", err)
		}
	}

	return utils.SendSimpleSuccess(c, preferences, "Notification preferences retrieved successfully")
}

// UpdateNotificationPreferences updates notification preferences for the current user
func (h *NotificationHandler) UpdateNotificationPreferences(c *fiber.Ctx) error {
	db := config.DB
	
	// Safely get organizationID
	organizationIDRaw := c.Locals("organizationID")
	if organizationIDRaw == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Organization context required",
		})
	}
	organizationID, ok := organizationIDRaw.(string)
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid organization context",
		})
	}
	
	// Safely get userID
	userIDRaw := c.Locals("userID")
	if userIDRaw == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "User context required",
		})
	}
	userID, ok := userIDRaw.(string)
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user context",
		})
	}

	// Parse request body
	var req NotificationPreferencesRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{
			Error:   "Invalid request body",
			Message: "Failed to parse notification preferences request",
		})
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{
			Error:   "Validation failed",
			Message: err.Error(),
		})
	}

	// Validate quiet hours
	if req.QuietHoursStart < 0 || req.QuietHoursStart > 23 {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{
			Error:   "Invalid quiet hours start",
			Message: "Quiet hours start must be between 0 and 23",
		})
	}
	if req.QuietHoursEnd < 0 || req.QuietHoursEnd > 23 {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{
			Error:   "Invalid quiet hours end",
			Message: "Quiet hours end must be between 0 and 23",
		})
	}

	// Get or create preferences
	var preferences models.NotificationPreferences
	if err := db.Where("user_id = ? AND organization_id = ?", userID, organizationID).
		First(&preferences).Error; err != nil {
		
		// If not found, create new preferences
		if err == gorm.ErrRecordNotFound {
			preferences = models.NotificationPreferences{
				ID:                     uuid.New().String(),
				UserID:                 userID,
				OrganizationID:         organizationID,
				CreatedAt:              time.Now(),
				UpdatedAt:              time.Now(),
			}
		} else {
			log.Printf("Error fetching notification preferences: %v", err)
			return utils.SendInternalError(c, "Failed to fetch notification preferences", err)
		}
	}

	// Update preferences
	preferences.EmailEnabled = req.EmailEnabled
	preferences.PushEnabled = req.PushEnabled
	preferences.InAppEnabled = req.InAppEnabled
	preferences.NotifyTaskAssigned = req.NotifyTaskAssigned
	preferences.NotifyTaskReassigned = req.NotifyTaskReassigned
	preferences.NotifyTaskApproved = req.NotifyTaskApproved
	preferences.NotifyTaskRejected = req.NotifyTaskRejected
	preferences.NotifyWorkflowComplete = req.NotifyWorkflowComplete
	preferences.NotifyApprovalOverdue = req.NotifyApprovalOverdue
	preferences.NotifyCommentsAdded = req.NotifyCommentsAdded
	preferences.QuietHoursEnabled = req.QuietHoursEnabled
	preferences.QuietHoursStart = req.QuietHoursStart
	preferences.QuietHoursEnd = req.QuietHoursEnd
	preferences.UpdatedAt = time.Now()

	// Save to database
	if err := db.Save(&preferences).Error; err != nil {
		log.Printf("Error updating notification preferences: %v", err)
		return utils.SendInternalError(c, "Failed to update notification preferences", err)
	}

	return utils.SendSimpleSuccess(c, preferences, "Notification preferences updated successfully")
}