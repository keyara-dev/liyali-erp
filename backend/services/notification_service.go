package services

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/logging"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/utils"
	"gorm.io/gorm"
)

// NotificationEvent represents a trigger event
type NotificationEvent struct {
	Type           string // approval_required, document_approved, document_rejected, assignment, status_change
	DocumentID     string
	DocumentType   string
	OrganizationID string // Required for org-scoped queries (e.g. notifyStatusChange)
	Action         string
	ActorID        string // User who triggered the event
	Details        string
	Timestamp      time.Time
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
	// Always dispatched in a fire-and-forget goroutine — recover so a panic in
	// any notification path can't crash the process.
	defer utils.RecoverPanic("notification.HandleWorkflowEvent")

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
	case "document_returned_for_revision", "document_returned_to_draft":
		return ns.notifyDocumentReturnedForRevision(event)
	default:
		logging.WithFields(map[string]interface{}{
			"event_type": event.Type,
			"operation":  "handle_notification_event",
		}).Warn("unknown_notification_event_type")
		return nil
	}
}

// notifyApprovalRequired creates notifications for approvers.
// When a task is assigned to a role (by UUID or plain name), all users with that role are notified.
func (ns *NotificationService) notifyApprovalRequired(event NotificationEvent) error {
	// Get workflow tasks for this document
	var tasks []models.WorkflowTask
	if err := ns.db.Where(
		"entity_id = ? AND UPPER(status) IN ('PENDING','CLAIMED')",
		event.DocumentID,
	).Find(&tasks).Error; err != nil {
		return fmt.Errorf("failed to fetch workflow tasks: %v", err)
	}

	totalNotified := 0

	for _, task := range tasks {
		// Collect recipient IDs for this task
		var recipientIDs []string

		if task.AssignedUserID != nil && *task.AssignedUserID != "" {
			// Specific user assignment
			recipientIDs = append(recipientIDs, *task.AssignedUserID)
		} else if task.AssignedRole != nil && *task.AssignedRole != "" {
			// Role-based assignment — notify all users with this role
			assignedRole := *task.AssignedRole
			if _, err := uuid.Parse(assignedRole); err == nil {
				// It's a UUID — resolve the org role
				var orgRole models.OrganizationRole
				if ns.db.Where("id = ?", assignedRole).First(&orgRole).Error == nil {
					if orgRole.IsSystemRole {
						// Notify all users with this system role name in the org
						var users []models.User
						ns.db.Where("role = ? AND current_organization_id = ? AND active = ?",
							orgRole.Name, task.OrganizationID, true).Find(&users)
						for _, u := range users {
							recipientIDs = append(recipientIDs, u.ID)
						}
					} else {
						// Notify all users with this custom org role
						var uors []models.UserOrganizationRole
						ns.db.Where("role_id = ? AND organization_id = ? AND active = ?",
							assignedRole, task.OrganizationID, true).Find(&uors)
						for _, uor := range uors {
							recipientIDs = append(recipientIDs, uor.UserID)
						}
					}
				}
			} else {
				// Plain role name — notify all users with this system role in the org
				var users []models.User
				ns.db.Where("role = ? AND current_organization_id = ? AND active = ?",
					assignedRole, task.OrganizationID, true).Find(&users)
				for _, u := range users {
					recipientIDs = append(recipientIDs, u.ID)
				}
			}
		} else if task.ClaimedBy != nil {
			recipientIDs = append(recipientIDs, *task.ClaimedBy)
		}

		if len(recipientIDs) == 0 {
			continue // no one to notify for this task
		}

		for _, recipientID := range recipientIDs {
			notification := models.Notification{
				ID:           uuid.New().String(),
				RecipientID:  recipientID,
				Type:         "approval_required",
				DocumentID:   event.DocumentID,
				DocumentType: event.DocumentType,
				Subject:      fmt.Sprintf("Action Required: %s Needs Approval (Stage %d)", event.DocumentType, task.StageNumber),
				Body: fmt.Sprintf(
					"A %s (ID: %s) requires your approval at stage %d: %s.\nPlease review and take action.",
					event.DocumentType, event.DocumentID, task.StageNumber, task.StageName,
				),
				Sent:      false,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			if err := ns.db.Create(&notification).Error; err != nil {
				logging.WithFields(map[string]interface{}{
					"operation":     "create_approval_notification",
					"recipient_id":  recipientID,
					"document_id":   event.DocumentID,
					"document_type": event.DocumentType,
				}).WithError(err).Error("failed_to_create_approval_notification")
				return err
			}
			totalNotified++
		}
	}

	logging.WithFields(map[string]interface{}{
		"operation":        "create_approval_notifications",
		"document_id":      event.DocumentID,
		"document_type":    event.DocumentType,
		"approvers_count":  totalNotified,
	}).Info("created_approval_notifications_for_approvers")
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
		recipientID = req.RequesterId
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
				recipientID = req.RequesterId
			}
		}
	default:
		logging.WithFields(map[string]interface{}{
			"operation":     "notify_approval_required",
			"document_type": event.DocumentType,
		}).Warn("notification_for_approval_not_configured")
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
		logging.WithFields(map[string]interface{}{
			"operation":     "create_approval_notification",
			"recipient_id":  recipientID,
			"document_id":   event.DocumentID,
			"document_type": event.DocumentType,
		}).WithError(err).Error("failed_to_create_approval_notification")
		return err
	}

	logging.WithFields(map[string]interface{}{
		"operation":     "create_approval_notification",
		"recipient_id":  recipientID,
		"document_id":   event.DocumentID,
		"document_type": event.DocumentType,
	}).Info("created_approval_notification_for_recipient")
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
		recipientID = req.RequesterId
	case "budget":
		var budget models.Budget
		if err := ns.db.First(&budget, "id = ?", event.DocumentID).Error; err != nil {
			return fmt.Errorf("failed to fetch budget: %v", err)
		}
		recipientID = budget.OwnerID
	case "purchase_order":
		var po models.PurchaseOrder
		if err := ns.db.First(&po, "id = ?", event.DocumentID).Error; err != nil {
			return fmt.Errorf("failed to fetch purchase order: %v", err)
		}
		// PO has no direct requester — trace back through linked requisition
		reqID := po.LinkedRequisition
		if reqID == "" && po.SourceRequisitionId != nil {
			reqID = *po.SourceRequisitionId
		}
		if reqID != "" {
			var req models.Requisition
			if err := ns.db.First(&req, "id = ?", reqID).Error; err == nil {
				recipientID = req.RequesterId
			}
		}
	case "payment_voucher":
		var pv models.PaymentVoucher
		if err := ns.db.First(&pv, "id = ?", event.DocumentID).Error; err != nil {
			return fmt.Errorf("failed to fetch payment voucher: %v", err)
		}
		// PV traces back through linked PO → requisition
		if pv.LinkedPO != "" {
			var po models.PurchaseOrder
			if err := ns.db.First(&po, "id = ?", pv.LinkedPO).Error; err == nil {
				reqID := po.LinkedRequisition
				if reqID == "" && po.SourceRequisitionId != nil {
					reqID = *po.SourceRequisitionId
				}
				if reqID != "" {
					var req models.Requisition
					if err := ns.db.First(&req, "id = ?", reqID).Error; err == nil {
						recipientID = req.RequesterId
					}
				}
			}
		}
	case "grn":
		var grn models.GoodsReceivedNote
		if err := ns.db.First(&grn, "id = ?", event.DocumentID).Error; err != nil {
			return fmt.Errorf("failed to fetch GRN: %v", err)
		}
		if grn.ReceivedBy != "" {
			recipientID = grn.ReceivedBy
		} else {
			recipientID = grn.CreatedBy
		}
	default:
		logging.WithFields(map[string]interface{}{
			"operation":     "notify_document_rejected",
			"document_type": event.DocumentType,
		}).Warn("notification_for_rejection_not_configured")
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
		logging.WithFields(map[string]interface{}{
			"operation":     "create_rejection_notification",
			"recipient_id":  recipientID,
			"document_id":   event.DocumentID,
			"document_type": event.DocumentType,
		}).WithError(err).Error("failed_to_create_rejection_notification")
		return err
	}

	logging.WithFields(map[string]interface{}{
		"operation":     "create_rejection_notification",
		"recipient_id":  recipientID,
		"document_id":   event.DocumentID,
		"document_type": event.DocumentType,
	}).Info("created_rejection_notification_for_recipient")
	return nil
}

// notifyDocumentReturnedForRevision notifies the document owner when their submission is returned for changes
func (ns *NotificationService) notifyDocumentReturnedForRevision(event NotificationEvent) error {
	var recipientID string

	switch event.DocumentType {
	case "requisition":
		var req models.Requisition
		if err := ns.db.First(&req, "id = ?", event.DocumentID).Error; err != nil {
			return fmt.Errorf("failed to fetch requisition: %v", err)
		}
		recipientID = req.RequesterId
	case "budget":
		var budget models.Budget
		if err := ns.db.First(&budget, "id = ?", event.DocumentID).Error; err != nil {
			return fmt.Errorf("failed to fetch budget: %v", err)
		}
		recipientID = budget.OwnerID
	case "purchase_order":
		var po models.PurchaseOrder
		if err := ns.db.First(&po, "id = ?", event.DocumentID).Error; err != nil {
			return fmt.Errorf("failed to fetch purchase order: %v", err)
		}
		reqID := po.LinkedRequisition
		if reqID == "" && po.SourceRequisitionId != nil {
			reqID = *po.SourceRequisitionId
		}
		if reqID != "" {
			var req models.Requisition
			if err := ns.db.First(&req, "id = ?", reqID).Error; err == nil {
				recipientID = req.RequesterId
			}
		}
	case "payment_voucher":
		var pv models.PaymentVoucher
		if err := ns.db.First(&pv, "id = ?", event.DocumentID).Error; err != nil {
			return fmt.Errorf("failed to fetch payment voucher: %v", err)
		}
		if pv.LinkedPO != "" {
			var po models.PurchaseOrder
			if err := ns.db.First(&po, "id = ?", pv.LinkedPO).Error; err == nil {
				reqID := po.LinkedRequisition
				if reqID == "" && po.SourceRequisitionId != nil {
					reqID = *po.SourceRequisitionId
				}
				if reqID != "" {
					var req models.Requisition
					if err := ns.db.First(&req, "id = ?", reqID).Error; err == nil {
						recipientID = req.RequesterId
					}
				}
			}
		}
	case "grn":
		var grn models.GoodsReceivedNote
		if err := ns.db.First(&grn, "id = ?", event.DocumentID).Error; err != nil {
			return fmt.Errorf("failed to fetch GRN: %v", err)
		}
		if grn.ReceivedBy != "" {
			recipientID = grn.ReceivedBy
		} else {
			recipientID = grn.CreatedBy
		}
	default:
		logging.WithFields(map[string]interface{}{
			"operation":     "notify_document_returned",
			"document_type": event.DocumentType,
		}).Warn("notification_for_revision_not_configured")
		return nil
	}

	if recipientID == "" {
		return fmt.Errorf("could not determine notification recipient for %s revision", event.DocumentType)
	}

	notification := models.Notification{
		ID:           uuid.New().String(),
		RecipientID:  recipientID,
		Type:         "document_rejected",
		DocumentID:   event.DocumentID,
		DocumentType: event.DocumentType,
		Subject:      fmt.Sprintf("Revision Required — %s", event.DocumentType),
		Body: fmt.Sprintf(
			"Your %s (ID: %s) has been returned for revision. Reason: %s\nPlease review and resubmit.",
			event.DocumentType, event.DocumentID, event.Details,
		),
		Sent:      false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := ns.db.Create(&notification).Error; err != nil {
		logging.WithFields(map[string]interface{}{
			"operation":     "create_revision_notification",
			"recipient_id":  recipientID,
			"document_id":   event.DocumentID,
			"document_type": event.DocumentType,
		}).WithError(err).Error("failed_to_create_revision_notification")
		return err
	}

	logging.WithFields(map[string]interface{}{
		"operation":     "create_revision_notification",
		"recipient_id":  recipientID,
		"document_id":   event.DocumentID,
		"document_type": event.DocumentType,
	}).Info("created_revision_notification_for_recipient")
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
		logging.WithFields(map[string]interface{}{
			"operation":   "create_assignment_notification",
			"actor_id":    event.ActorID,
			"document_id": event.DocumentID,
		}).WithError(err).Error("failed_to_create_assignment_notification")
		return err
	}

	logging.WithFields(map[string]interface{}{
		"operation":   "create_assignment_notification",
		"actor_id":    event.ActorID,
		"document_id": event.DocumentID,
	}).Info("created_assignment_notification_for_user")
	return nil
}

// notifyStatusChange creates notifications for status changes
func (ns *NotificationService) notifyStatusChange(event NotificationEvent) error {
	// Get all users with finance or admin role in the same organization
	var admins []models.User
	if err := ns.db.Where("role IN ? AND active = ? AND current_organization_id = ?",
		[]string{"finance", "admin"}, true, event.OrganizationID).
		Find(&admins).Error; err != nil {
		logging.WithFields(map[string]interface{}{
			"operation": "fetch_admin_users",
		}).WithError(err).Error("failed_to_fetch_admin_users")
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
			logging.WithFields(map[string]interface{}{
				"operation":   "create_status_change_notification",
				"admin_id":    admin.ID,
				"document_id": event.DocumentID,
			}).WithError(err).Error("failed_to_create_status_change_notification")
		}
	}

	logging.WithFields(map[string]interface{}{
		"operation":    "create_status_change_notifications",
		"admins_count": len(admins),
		"document_id":  event.DocumentID,
	}).Info("created_status_change_notifications_for_admins")
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
			"sent":      true,
			"sent_at":   &now,
			"is_read":   true,
			"read_at":   &now,
			"updated_at": now,
		}).Error; err != nil {
		return fmt.Errorf("failed to mark notification as read: %v", err)
	}

	logging.WithFields(map[string]interface{}{
		"operation":       "mark_notification_as_read",
		"notification_id": notificationID,
	}).Info("marked_notification_as_read")
	return nil
}

// MarkMultipleAsRead marks multiple notifications as read
func (ns *NotificationService) MarkMultipleAsRead(notificationIDs []string) error {
	now := time.Now()
	if err := ns.db.Model(&models.Notification{}).
		Where("id IN ?", notificationIDs).
		Updates(map[string]interface{}{
			"sent":      true,
			"sent_at":   &now,
			"is_read":   true,
			"read_at":   &now,
			"updated_at": now,
		}).Error; err != nil {
		return fmt.Errorf("failed to mark notifications as read: %v", err)
	}

	logging.WithFields(map[string]interface{}{
		"operation":          "mark_multiple_notifications_as_read",
		"notifications_count": len(notificationIDs),
	}).Info("marked_multiple_notifications_as_read")
	return nil
}

// DeleteNotification deletes a notification
func (ns *NotificationService) DeleteNotification(notificationID string) error {
	if err := ns.db.Delete(&models.Notification{}, "id = ?", notificationID).Error; err != nil {
		return fmt.Errorf("failed to delete notification: %v", err)
	}

	logging.WithFields(map[string]interface{}{
		"operation":       "delete_notification",
		"notification_id": notificationID,
	}).Info("deleted_notification")
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
		logging.WithFields(map[string]interface{}{
			"operation":    "send_notification",
			"recipient_id": notif.RecipientID,
			"subject":      notif.Subject,
		}).Info("notification_sent")
	}

	// Mark all processed notifications as sent
	if len(notifIDs) > 0 {
		if err := ns.MarkMultipleAsRead(notifIDs); err != nil {
			logging.WithFields(map[string]interface{}{
				"operation":          "mark_notifications_as_sent",
				"notifications_count": len(notifIDs),
			}).WithError(err).Error("failed_to_mark_notifications_as_sent")
		}
	}

	logging.WithFields(map[string]interface{}{
		"operation":          "process_pending_notifications",
		"notifications_count": len(notifIDs),
	}).Info("processed_pending_notifications")
	return nil
}
