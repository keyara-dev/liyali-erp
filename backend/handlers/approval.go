package handlers

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/types"
	"github.com/liyali/liyali-gateway/utils"
	"gorm.io/gorm"
	"gorm.io/datatypes"
)

// GetApprovalTasks retrieves approval tasks with pagination and filtering
func GetApprovalTasks(c fiber.Ctx) error {
	db := config.DB
	organizationID := c.Locals("organization_id").(string)
	userID := c.Locals("user_id").(string)

	// Extract query parameters
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	status := c.Query("status")
	documentType := c.Query("document_type")
	assignedToMe := c.QueryBool("assigned_to_me", false)

	// Build query
	query := db.Where("organization_id = ?", organizationID)

	if assignedToMe {
		query = query.Where("approver_id = ?", userID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if documentType != "" {
		query = query.Where("document_type = ?", documentType)
	}

	// Get total count
	var total int64
	if err := query.Model(&models.ApprovalTask{}).Count(&total).Error; err != nil {
		return utils.SendInternalError(c, "Failed to count tasks", err)
	}

	// Fetch paginated results
	var tasks []models.ApprovalTask
	offset := (page - 1) * limit
	if err := query.
		Offset(offset).
		Limit(limit).
		Preload("Approver").
		Order("created_at DESC").
		Find(&tasks).Error; err != nil {
		return utils.SendInternalError(c, "Failed to fetch tasks", err)
	}

	// Convert to response format
	responses := make([]types.ApprovalTaskResponse, 0, len(tasks))
	for _, task := range tasks {
		responses = append(responses, modelToApprovalTaskResponse(task))
	}

	pagination := utils.CalculatePagination(page, limit, total)
	return utils.SendSuccess(c, fiber.StatusOK, responses, "Approval tasks retrieved", pagination)
}

// GetApprovalTask retrieves a single approval task with full details
func GetApprovalTask(c fiber.Ctx) error {
	taskID := c.Params("id")
	if taskID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Task ID is required",
		})
	}

	organizationID := c.Locals("organization_id").(string)

	var task models.ApprovalTask
	if err := config.DB.
		Preload("Approver").
		Where("id = ? AND organization_id = ?", taskID, organizationID).
		First(&task).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Task not found",
			})
		}
		return utils.SendInternalError(c, "Failed to fetch task", err)
	}

	// Fetch related document based on document_type
	var documentDetail interface{}
	switch task.DocumentType {
	case "requisition":
		var req models.Requisition
		if err := config.DB.
			Preload("Requester").
			Where("id = ?", task.DocumentID).
			First(&req).Error; err == nil {
			documentDetail = modelToRequisitionResponse(req)
		}
	case "purchase_order":
		var po models.PurchaseOrder
		if err := config.DB.
			Where("id = ?", task.DocumentID).
			First(&po).Error; err == nil {
			documentDetail = modelToPurchaseOrderResponse(po)
		}
	case "payment_voucher":
		var pv models.PaymentVoucher
		if err := config.DB.
			Where("id = ?", task.DocumentID).
			First(&pv).Error; err == nil {
			documentDetail = modelToPaymentVoucherResponse(pv)
		}
	case "grn":
		var grn models.GoodsReceivedNote
		if err := config.DB.
			Where("id = ?", task.DocumentID).
			First(&grn).Error; err == nil {
			documentDetail = modelToGRNResponse(grn)
		}
	}

	return utils.SendSuccess(c, fiber.StatusOK, fiber.Map{
		"task":     modelToApprovalTaskResponse(task),
		"document": documentDetail,
	}, "Approval task retrieved")
}

// ApproveTask marks a task as approved and moves to next stage
func ApproveTask(c fiber.Ctx) error {
	taskID := c.Params("id")
	userID := c.Locals("user_id").(string)
	organizationID := c.Locals("organization_id").(string)

	var req types.ApproveTaskRequest
	if err := c.BindJSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	// Validate signature
	if req.Signature == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Digital signature is required",
		})
	}

	// Fetch approval task
	var task models.ApprovalTask
	if err := config.DB.
		Where("id = ? AND organization_id = ?", taskID, organizationID).
		First(&task).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Task not found",
			})
		}
		return utils.SendInternalError(c, "Failed to fetch task", err)
	}

	// Verify user is the assigned approver
	if task.ApproverID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "You are not the assigned approver for this task",
		})
	}

	// Verify task is still pending
	if task.Status != "pending" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": fmt.Sprintf("Task is already %s", task.Status),
		})
	}

	// Get approver details
	var approver models.User
	if err := config.DB.Where("id = ?", userID).First(&approver).Error; err != nil {
		return utils.SendInternalError(c, "Failed to fetch approver details", err)
	}

	// Update task status
	task.Status = "approved"
	task.Comments = req.Comments
	task.Signature = req.Signature
	task.UpdatedAt = time.Now()

	if err := config.DB.Save(&task).Error; err != nil {
		return utils.SendInternalError(c, "Failed to approve task", err)
	}

	// Create approval record and update document
	approvalRecord := types.ApprovalRecord{
		ApproverID:   approver.ID,
		ApproverName: approver.Name,
		Status:       "approved",
		Comments:     req.Comments,
		Signature:    req.Signature,
		ApprovedAt:   time.Now(),
	}

	if err := updateDocumentApprovalHistory(task.DocumentID, task.DocumentType, approvalRecord); err != nil {
		return utils.SendInternalError(c, "Failed to update document approval history", err)
	}

	// Create audit log
	if err := createAuditLog(organizationID, task.DocumentID, task.DocumentType, userID, "approve", fiber.Map{
		"stage":    task.Stage,
		"comments": req.Comments,
	}); err != nil {
		// Log error but don't fail the approval
		fmt.Printf("Failed to create audit log: %v\n", err)
	}

	// Get the document to determine if there's a next stage
	var nextApproverTask *models.ApprovalTask
	if err := config.DB.
		Where("document_id = ? AND document_type = ? AND stage > ? AND status = ?",
			task.DocumentID, task.DocumentType, task.Stage, "pending").
		Order("stage ASC").
		First(&nextApproverTask).Error; err != nil && err != gorm.ErrRecordNotFound {
		return utils.SendInternalError(c, "Failed to check for next approver", err)
	}

	// Send notification to next approver if exists
	if nextApproverTask != nil {
		// Notify next approver (implementation depends on notification service)
		_ = createNotification(organizationID, nextApproverTask.ApproverID, "approval_required",
			task.DocumentID, task.DocumentType, "New approval required")
	} else {
		// Final approval - update document status
		if err := updateDocumentStatus(task.DocumentID, task.DocumentType, "approved"); err != nil {
			return utils.SendInternalError(c, "Failed to update document status", err)
		}

		// Notify document creator that it's approved
		doc, _ := getDocumentCreator(task.DocumentID, task.DocumentType)
		if doc != nil {
			_ = createNotification(organizationID, doc.(map[string]interface{})["creatorId"].(string), "approved",
				task.DocumentID, task.DocumentType, "Your document has been approved")
		}
	}

	return utils.SendSuccess(c, fiber.StatusOK, fiber.Map{
		"taskId":           task.ID,
		"documentId":       task.DocumentID,
		"status":           "approved",
		"stage":            task.Stage,
		"nextApproverTask": nextApproverTask,
	}, "Task approved successfully")
}

// RejectTask marks a task as rejected and returns document to draft
func RejectTask(c fiber.Ctx) error {
	taskID := c.Params("id")
	userID := c.Locals("user_id").(string)
	organizationID := c.Locals("organization_id").(string)

	var req types.RejectTaskRequest
	if err := c.BindJSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
		})
	}

	if req.Signature == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Digital signature is required",
		})
	}

	// Fetch approval task
	var task models.ApprovalTask
	if err := config.DB.
		Where("id = ? AND organization_id = ?", taskID, organizationID).
		First(&task).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Task not found",
			})
		}
		return utils.SendInternalError(c, "Failed to fetch task", err)
	}

	if task.ApproverID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "You are not the assigned approver for this task",
		})
	}

	if task.Status != "pending" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": fmt.Sprintf("Task is already %s", task.Status),
		})
	}

	// Get approver details
	var approver models.User
	if err := config.DB.Where("id = ?", userID).First(&approver).Error; err != nil {
		return utils.SendInternalError(c, "Failed to fetch approver details", err)
	}

	// Update task
	task.Status = "rejected"
	task.Comments = req.Comments
	task.Signature = req.Signature
	task.UpdatedAt = time.Now()

	if err := config.DB.Save(&task).Error; err != nil {
		return utils.SendInternalError(c, "Failed to reject task", err)
	}

	// Create rejection record
	rejectionRecord := types.ApprovalRecord{
		ApproverID:   approver.ID,
		ApproverName: approver.Name,
		Status:       "rejected",
		Comments:     req.Comments,
		Signature:    req.Signature,
		ApprovedAt:   time.Now(),
	}

	if err := updateDocumentApprovalHistory(task.DocumentID, task.DocumentType, rejectionRecord); err != nil {
		return utils.SendInternalError(c, "Failed to update document approval history", err)
	}

	// Return document to DRAFT and mark all pending tasks as cancelled
	if err := updateDocumentStatus(task.DocumentID, task.DocumentType, "draft"); err != nil {
		return utils.SendInternalError(c, "Failed to update document status", err)
	}

	// Cancel remaining approval tasks for this document
	if err := config.DB.
		Where("document_id = ? AND document_type = ? AND status = ?",
			task.DocumentID, task.DocumentType, "pending").
		Update("status", "cancelled").Error; err != nil {
		fmt.Printf("Failed to cancel remaining tasks: %v\n", err)
	}

	// Create audit log
	if err := createAuditLog(organizationID, task.DocumentID, task.DocumentType, userID, "reject", fiber.Map{
		"stage":   task.Stage,
		"remarks": req.Remarks,
	}); err != nil {
		fmt.Printf("Failed to create audit log: %v\n", err)
	}

	// Notify document creator about rejection
	doc, _ := getDocumentCreator(task.DocumentID, task.DocumentType)
	if doc != nil {
		creatorID := doc.(map[string]interface{})["creatorId"].(string)
		_ = createNotification(organizationID, creatorID, "rejected",
			task.DocumentID, task.DocumentType, req.Remarks)
	}

	return utils.SendSuccess(c, fiber.StatusOK, fiber.Map{
		"taskId":     task.ID,
		"documentId": task.DocumentID,
		"status":     "rejected",
		"remarks":    req.Remarks,
	}, "Task rejected successfully")
}

// ReassignTask reassigns task to different approver
func ReassignTask(c fiber.Ctx) error {
	taskID := c.Params("id")
	organizationID := c.Locals("organization_id").(string)
	userID := c.Locals("user_id").(string)

	var req types.ReassignTaskRequest
	if err := c.BindJSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
		})
	}

	// Fetch task
	var task models.ApprovalTask
	if err := config.DB.
		Where("id = ? AND organization_id = ?", taskID, organizationID).
		First(&task).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Task not found",
			})
		}
		return utils.SendInternalError(c, "Failed to fetch task", err)
	}

	// Verify new approver exists and is in the organization
	var newApprover models.User
	if err := config.DB.
		Where("id = ?", req.NewApproverId).
		First(&newApprover).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "New approver not found",
			})
		}
		return utils.SendInternalError(c, "Failed to verify approver", err)
	}

	// Store old approver for audit
	oldApproverID := task.ApproverID

	// Update task
	task.ApproverID = req.NewApproverId
	task.UpdatedAt = time.Now()

	if err := config.DB.Save(&task).Error; err != nil {
		return utils.SendInternalError(c, "Failed to reassign task", err)
	}

	// Create audit log
	if err := createAuditLog(organizationID, task.DocumentID, task.DocumentType, userID, "reassign", fiber.Map{
		"from":   oldApproverID,
		"to":     req.NewApproverId,
		"reason": req.Reason,
	}); err != nil {
		fmt.Printf("Failed to create audit log: %v\n", err)
	}

	// Notify new approver
	_ = createNotification(organizationID, req.NewApproverId, "approval_required",
		task.DocumentID, task.DocumentType, "Task reassigned to you - "+req.Reason)

	return utils.SendSuccess(c, fiber.StatusOK, fiber.Map{
		"taskId":              task.ID,
		"previousApproverId":  oldApproverID,
		"newApproverId":       req.NewApproverId,
		"reason":              req.Reason,
		"reassignedAt":        time.Now(),
		"reassignedBy":        userID,
	}, "Task reassigned successfully")
}

// GetApprovalHistory retrieves approval history for a document
func GetApprovalHistory(c fiber.Ctx) error {
	documentID := c.Params("documentId")
	organizationID := c.Locals("organization_id").(string)

	if documentID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Document ID is required",
		})
	}

	// Try to fetch from requisition first (most common)
	var requisition models.Requisition
	if err := config.DB.
		Where("id = ? AND organization_id = ?", documentID, organizationID).
		First(&requisition).Error; err == nil {
		// Parse approval history JSON
		var history []types.ApprovalRecord
		if err := json.Unmarshal(requisition.ApprovalHistory, &history); err == nil {
			return utils.SendSuccess(c, fiber.StatusOK, history, "Approval history retrieved")
		}
	}

	// If not requisition, try other document types
	var history []types.ApprovalRecord
	return utils.SendSuccess(c, fiber.StatusOK, history, "Approval history retrieved")
}

// Helper functions

func updateDocumentApprovalHistory(docID string, docType string, approvalRecord types.ApprovalRecord) error {
	switch docType {
	case "requisition":
		var req models.Requisition
		if err := config.DB.Where("id = ?", docID).First(&req).Error; err != nil {
			return err
		}

		var history []types.ApprovalRecord
		if err := json.Unmarshal(req.ApprovalHistory, &history); err != nil {
			history = []types.ApprovalRecord{}
		}

		history = append(history, approvalRecord)
		historyJSON, _ := json.Marshal(history)

		return config.DB.Model(&req).Update("approval_history", historyJSON).Error

	case "purchase_order":
		var po models.PurchaseOrder
		if err := config.DB.Where("id = ?", docID).First(&po).Error; err != nil {
			return err
		}

		var history []types.ApprovalRecord
		if err := json.Unmarshal(po.ApprovalHistory, &history); err != nil {
			history = []types.ApprovalRecord{}
		}

		history = append(history, approvalRecord)
		historyJSON, _ := json.Marshal(history)

		return config.DB.Model(&po).Update("approval_history", historyJSON).Error

	case "payment_voucher":
		var pv models.PaymentVoucher
		if err := config.DB.Where("id = ?", docID).First(&pv).Error; err != nil {
			return err
		}

		var history []types.ApprovalRecord
		if err := json.Unmarshal(pv.ApprovalHistory, &history); err != nil {
			history = []types.ApprovalRecord{}
		}

		history = append(history, approvalRecord)
		historyJSON, _ := json.Marshal(history)

		return config.DB.Model(&pv).Update("approval_history", historyJSON).Error

	case "grn":
		var grn models.GoodsReceivedNote
		if err := config.DB.Where("id = ?", docID).First(&grn).Error; err != nil {
			return err
		}

		var history []types.ApprovalRecord
		if err := json.Unmarshal(grn.ApprovalHistory, &history); err != nil {
			history = []types.ApprovalRecord{}
		}

		history = append(history, approvalRecord)
		historyJSON, _ := json.Marshal(history)

		return config.DB.Model(&grn).Update("approval_history", historyJSON).Error
	}

	return fmt.Errorf("unknown document type: %s", docType)
}

func updateDocumentStatus(docID string, docType string, newStatus string) error {
	switch docType {
	case "requisition":
		return config.DB.Model(&models.Requisition{}).
			Where("id = ?", docID).
			Update("status", newStatus).Error

	case "purchase_order":
		return config.DB.Model(&models.PurchaseOrder{}).
			Where("id = ?", docID).
			Update("status", newStatus).Error

	case "payment_voucher":
		return config.DB.Model(&models.PaymentVoucher{}).
			Where("id = ?", docID).
			Update("status", newStatus).Error

	case "grn":
		return config.DB.Model(&models.GoodsReceivedNote{}).
			Where("id = ?", docID).
			Update("status", newStatus).Error
	}

	return fmt.Errorf("unknown document type: %s", docType)
}

func createAuditLog(orgID, docID, docType, userID, action string, changes interface{}) error {
	changesJSON, _ := json.Marshal(changes)
	log := models.AuditLog{
		ID:           uuid.New().String(),
		DocumentID:   docID,
		DocumentType: docType,
		UserID:       userID,
		Action:       action,
		Changes:      changesJSON,
		CreatedAt:    time.Now(),
	}
	return config.DB.Create(&log).Error
}

func createNotification(orgID, recipientID, notificationType, docID, docType, message string) error {
	notification := models.Notification{
		ID:             uuid.New().String(),
		OrganizationID: orgID,
		RecipientID:    recipientID,
		Type:           notificationType,
		DocumentID:     docID,
		DocumentType:   docType,
		Subject:        fmt.Sprintf("%s %s", docType, notificationType),
		Body:           message,
		Sent:           false,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	return config.DB.Create(&notification).Error
}

func getDocumentCreator(docID string, docType string) (interface{}, error) {
	switch docType {
	case "requisition":
		var req models.Requisition
		if err := config.DB.Where("id = ?", docID).First(&req).Error; err != nil {
			return nil, err
		}
		return fiber.Map{
			"creatorId": req.RequesterID,
		}, nil
	}
	return nil, nil
}

func modelToApprovalTaskResponse(task models.ApprovalTask) types.ApprovalTaskResponse {
	return types.ApprovalTaskResponse{
		ID:             task.ID,
		OrganizationID: task.OrganizationID,
		DocumentID:     task.DocumentID,
		DocumentType:   task.DocumentType,
		ApproverID:     task.ApproverID,
		Status:         task.Status,
		Stage:          task.Stage,
		Comments:       task.Comments,
		CreatedAt:      task.CreatedAt,
		UpdatedAt:      task.UpdatedAt,
	}
}
