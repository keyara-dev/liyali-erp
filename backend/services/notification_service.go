package services

import (
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/models"
	"gorm.io/gorm"
)

// NotificationEvent represents a trigger event
type NotificationEvent struct {
	Type         string // approval_required, document_approved, document_rejected, assignment, status_change
	DocumentID   string
	DocumentType string
	Action       string
	ActorID      string  // User who triggered the event
	Details      string
	Timestamp    time.Time
}

// NotificationService handles notification creation and management
type NotificationService struct {
	db *gorm.DB
}

// NewNotificationService creates a new notification service
func NewNotificationService(db *gorm.DB) *NotificationService {
	return &NotificationService{db: db}
}

// HandleWorkflowEvent processes workflow events and creates notifications
func (ns *NotificationService) HandleWorkflowEvent(event NotificationEvent) error {
	switch event.Type {
	case "approval_required":
		return ns.notifyApprovalRequired(event)
	case "document_approved":
		return ns.notifyDocumentApproved(event)
	case "document_rejected":
		return ns.notifyDocumentRejected(event)
	case "assignment":
		return ns.notifyDocumentAssignment(event)
	case "status_change":
		return ns.notifyStatusChange(event)
	default:
		log.Printf("Unknown notification event type: %s", event.Type)
		return nil
	}
}

// notifyApprovalRequired creates notifications for approvers
func (ns *NotificationService) notifyApprovalRequired(event NotificationEvent) error {
	// Get approval tasks for this document
	var tasks []models.ApprovalTask
	if err := ns.db.Where(
		"document_id = ? AND status = ?",
		event.DocumentID, "pending",
	).Find(&tasks).Error; err != nil {
		return fmt.Errorf("failed to fetch approval tasks: %v", err)
	}

	// Create notification for each approver
	for _, task := range tasks {
		notification := models.Notification{
			ID:           uuid.New().String(),
			RecipientID:  task.ApproverID,
			Type:         "approval_required",
			DocumentID:   event.DocumentID,
			DocumentType: event.DocumentType,
			Subject:      fmt.Sprintf("Action Required: %s Needs Approval (Stage %d)", event.DocumentType, task.Stage),
			Body: fmt.Sprintf(
				"A %s (ID: %s) requires your approval at stage %d.\nPlease review and take action.",
				event.DocumentType, event.DocumentID, task.Stage,
			),
			Sent:      false,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := ns.db.Create(&notification).Error; err != nil {
			log.Printf("Error creating approval notification: %v", err)
			return err
		}
	}

	log.Printf("Created approval notifications for %d approvers", len(tasks))
	return nil
}

// notifyDocumentApproved creates notifications when document is approved
func (ns *NotificationService) notifyDocumentApproved(event NotificationEvent) error {
	// Get document requester/owner
	var recipientID string

	switch event.DocumentType {
	case "requisition":
		var req models.Requisition
		if err := ns.db.First(&req, "id = ?", event.DocumentID).Error; err != nil {
			return fmt.Errorf("failed to fetch requisition: %v", err)
		}
		recipientID = req.RequesterID
	case "budget":
		var budget models.Budget
		if err := ns.db.First(&budget, "id = ?", event.DocumentID).Error; err != nil {
			return fmt.Errorf("failed to fetch budget: %v", err)
		}
		recipientID = budget.OwnerID
	case "po":
		var po models.PurchaseOrder
		if err := ns.db.First(&po, "id = ?", event.DocumentID).Error; err != nil {
			return fmt.Errorf("failed to fetch PO: %v", err)
		}
		// For PO, notify the requester if linked to requisition
		if po.LinkedRequisition != "" {
			var req models.Requisition
			if err := ns.db.First(&req, "id = ?", po.LinkedRequisition).Error; err == nil {
				recipientID = req.RequesterID
			}
		}
	default:
		log.Printf("Notification for %s approval not configured", event.DocumentType)
		return nil
	}

	if recipientID == "" {
		return fmt.Errorf("could not determine notification recipient for %s", event.DocumentType)
	}

	notification := models.Notification{
		ID:           uuid.New().String(),
		RecipientID:  recipientID,
		Type:         "document_approved",
		DocumentID:   event.DocumentID,
		DocumentType: event.DocumentType,
		Subject:      fmt.Sprintf("%s Approved", event.DocumentType),
		Body: fmt.Sprintf(
			"Your %s (ID: %s) has been approved and is ready for the next stage.",
			event.DocumentType, event.DocumentID,
		),
		Sent:      false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := ns.db.Create(&notification).Error; err != nil {
		log.Printf("Error creating approval notification: %v", err)
		return err
	}

	log.Printf("Created approval notification for recipient %s", recipientID)
	return nil
}

// notifyDocumentRejected creates notifications when document is rejected
func (ns *NotificationService) notifyDocumentRejected(event NotificationEvent) error {
	// Get document requester/owner
	var recipientID string

	switch event.DocumentType {
	case "requisition":
		var req models.Requisition
		if err := ns.db.First(&req, "id = ?", event.DocumentID).Error; err != nil {
			return fmt.Errorf("failed to fetch requisition: %v", err)
		}
		recipientID = req.RequesterID
	case "budget":
		var budget models.Budget
		if err := ns.db.First(&budget, "id = ?", event.DocumentID).Error; err != nil {
			return fmt.Errorf("failed to fetch budget: %v", err)
		}
		recipientID = budget.OwnerID
	default:
		log.Printf("Notification for %s rejection not configured", event.DocumentType)
		return nil
	}

	if recipientID == "" {
		return fmt.Errorf("could not determine notification recipient for %s", event.DocumentType)
	}

	notification := models.Notification{
		ID:           uuid.New().String(),
		RecipientID:  recipientID,
		Type:         "document_rejected",
		DocumentID:   event.DocumentID,
		DocumentType: event.DocumentType,
		Subject:      fmt.Sprintf("%s Rejected", event.DocumentType),
		Body: fmt.Sprintf(
			"Your %s (ID: %s) has been rejected. Details: %s\nPlease review and resubmit if needed.",
			event.DocumentType, event.DocumentID, event.Details,
		),
		Sent:      false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := ns.db.Create(&notification).Error; err != nil {
		log.Printf("Error creating rejection notification: %v", err)
		return err
	}

	log.Printf("Created rejection notification for recipient %s", recipientID)
	return nil
}

// notifyDocumentAssignment creates notifications when a document is assigned
func (ns *NotificationService) notifyDocumentAssignment(event NotificationEvent) error {
	// Get the user the document was assigned to
	notification := models.Notification{
		ID:           uuid.New().String(),
		RecipientID:  event.ActorID,
		Type:         "assignment",
		DocumentID:   event.DocumentID,
		DocumentType: event.DocumentType,
		Subject:      fmt.Sprintf("%s Assigned to You", event.DocumentType),
		Body: fmt.Sprintf(
			"A %s (ID: %s) has been assigned to you for review or action.",
			event.DocumentType, event.DocumentID,
		),
		Sent:      false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := ns.db.Create(&notification).Error; err != nil {
		log.Printf("Error creating assignment notification: %v", err)
		return err
	}

	log.Printf("Created assignment notification for user %s", event.ActorID)
	return nil
}

// notifyStatusChange creates notifications for status changes
func (ns *NotificationService) notifyStatusChange(event NotificationEvent) error {
	// Get all users with finance or admin role for status change notifications
	var admins []models.User
	if err := ns.db.Where("role IN ? AND active = ?", []string{"finance", "admin"}, true).
		Find(&admins).Error; err != nil {
		log.Printf("Error fetching admin users: %v", err)
		return nil
	}

	// Create notifications for each admin
	for _, admin := range admins {
		notification := models.Notification{
			ID:           uuid.New().String(),
			RecipientID:  admin.ID,
			Type:         "status_change",
			DocumentID:   event.DocumentID,
			DocumentType: event.DocumentType,
			Subject:      fmt.Sprintf("%s Status Updated", event.DocumentType),
			Body: fmt.Sprintf(
				"A %s (ID: %s) status has changed. Details: %s",
				event.DocumentType, event.DocumentID, event.Details,
			),
			Sent:      false,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := ns.db.Create(&notification).Error; err != nil {
			log.Printf("Error creating status change notification: %v", err)
		}
	}

	log.Printf("Created status change notifications for %d admins", len(admins))
	return nil
}

// GetPendingNotifications returns undelivered notifications for a user
func (ns *NotificationService) GetPendingNotifications(userID string) ([]models.Notification, error) {
	var notifications []models.Notification
	if err := ns.db.Where(
		"recipient_id = ? AND sent = ?",
		userID, false,
	).Order("created_at DESC").
		Limit(50).
		Find(&notifications).Error; err != nil {
		return nil, err
	}
	return notifications, nil
}

// GetNotificationsSince returns notifications since a specific time
func (ns *NotificationService) GetNotificationsSince(userID string, since time.Time) ([]models.Notification, error) {
	var notifications []models.Notification
	if err := ns.db.Where(
		"recipient_id = ? AND created_at >= ?",
		userID, since,
	).Order("created_at DESC").
		Find(&notifications).Error; err != nil {
		return nil, err
	}
	return notifications, nil
}

// MarkAsRead marks a notification as read (sent)
func (ns *NotificationService) MarkAsRead(notificationID string) error {
	now := time.Now()
	if err := ns.db.Model(&models.Notification{}).
		Where("id = ?", notificationID).
		Updates(map[string]interface{}{
			"sent":     true,
			"sent_at":  &now,
			"updated_at": now,
		}).Error; err != nil {
		return fmt.Errorf("failed to mark notification as read: %v", err)
	}

	log.Printf("Marked notification %s as read", notificationID)
	return nil
}

// MarkMultipleAsRead marks multiple notifications as read
func (ns *NotificationService) MarkMultipleAsRead(notificationIDs []string) error {
	now := time.Now()
	if err := ns.db.Model(&models.Notification{}).
		Where("id IN ?", notificationIDs).
		Updates(map[string]interface{}{
			"sent":     true,
			"sent_at":  &now,
			"updated_at": now,
		}).Error; err != nil {
		return fmt.Errorf("failed to mark notifications as read: %v", err)
	}

	log.Printf("Marked %d notifications as read", len(notificationIDs))
	return nil
}

// DeleteNotification deletes a notification
func (ns *NotificationService) DeleteNotification(notificationID string) error {
	if err := ns.db.Delete(&models.Notification{}, "id = ?", notificationID).Error; err != nil {
		return fmt.Errorf("failed to delete notification: %v", err)
	}

	log.Printf("Deleted notification %s", notificationID)
	return nil
}

// GetNotificationStats returns notification statistics for a user
func (ns *NotificationService) GetNotificationStats(userID string) (map[string]interface{}, error) {
	var pendingCount int64
	var readCount int64
	var totalCount int64

	ns.db.Model(&models.Notification{}).
		Where("recipient_id = ? AND sent = ?", userID, false).
		Count(&pendingCount)

	ns.db.Model(&models.Notification{}).
		Where("recipient_id = ? AND sent = ?", userID, true).
		Count(&readCount)

	ns.db.Model(&models.Notification{}).
		Where("recipient_id = ?", userID).
		Count(&totalCount)

	return map[string]interface{}{
		"pending": pendingCount,
		"read":    readCount,
		"total":   totalCount,
	}, nil
}

// GetNotificationsByType returns notifications of a specific type
func (ns *NotificationService) GetNotificationsByType(userID, notifType string) ([]models.Notification, error) {
	var notifications []models.Notification
	if err := ns.db.Where(
		"recipient_id = ? AND type = ?",
		userID, notifType,
	).Order("created_at DESC").
		Find(&notifications).Error; err != nil {
		return nil, err
	}
	return notifications, nil
}

// ProcessPendingNotifications processes and sends all pending notifications
// In production, this would send emails/SMS
func (ns *NotificationService) ProcessPendingNotifications() error {
	var notifications []models.Notification
	if err := ns.db.Where("sent = ?", false).
		Limit(100).
		Find(&notifications).Error; err != nil {
		return fmt.Errorf("failed to fetch pending notifications: %v", err)
	}

	var notifIDs []string
	for _, notif := range notifications {
		notifIDs = append(notifIDs, notif.ID)
		// Here you would send the notification via email/SMS
		log.Printf(
			"[NOTIFICATION] To: %s | Subject: %s | Body: %s",
			notif.RecipientID, notif.Subject, notif.Body,
		)
	}

	// Mark all processed notifications as sent
	if len(notifIDs) > 0 {
		if err := ns.MarkMultipleAsRead(notifIDs); err != nil {
			log.Printf("Error marking notifications as sent: %v", err)
		}
	}

	log.Printf("Processed %d pending notifications", len(notifIDs))
	return nil
}
