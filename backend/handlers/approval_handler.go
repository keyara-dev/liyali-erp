package handlers

import (
	"log"
	"strconv"
	"time"

	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/types"
	"github.com/liyali/liyali-gateway/utils"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type ApprovalHandler struct {
	validate *validator.Validate
}

func NewApprovalHandler() *ApprovalHandler {
	return &ApprovalHandler{
		validate: validator.New(),
	}
}

// Request/Response Types
type ApproveTaskRequest struct {
	Signature string `json:"signature" validate:"required"`
	Comment   string `json:"comment"`
}

type RejectTaskRequest struct {
	Signature string `json:"signature" validate:"required"`
	Reason    string `json:"reason" validate:"required"`
}

type ReassignTaskRequest struct {
	NewUserID string `json:"newUserId" validate:"required"`
	Reason    string `json:"reason"`
}

type BulkApproveRequest struct {
	TaskIDs   []string `json:"taskIds" validate:"required,min=1"`
	Signature string   `json:"signature" validate:"required"`
	Comment   string   `json:"comment"`
}

type BulkRejectRequest struct {
	TaskIDs   []string `json:"taskIds" validate:"required,min=1"`
	Signature string   `json:"signature" validate:"required"`
	Reason    string   `json:"reason" validate:"required"`
}

type BulkReassignRequest struct {
	TaskIDs   []string `json:"taskIds" validate:"required,min=1"`
	NewUserID string   `json:"newUserId" validate:"required"`
	Reason    string   `json:"reason"`
}

type BulkOperationResponse struct {
	SuccessCount int      `json:"successCount"`
	FailureCount int      `json:"failureCount"`
	SuccessIDs   []string `json:"successIds"`
	Errors       []string `json:"errors,omitempty"`
}

// GetApprovalTasks retrieves approval tasks with pagination and filtering
func (h *ApprovalHandler) GetApprovalTasks(c *fiber.Ctx) error {
	db := config.DB
	organizationID := c.Locals("organization_id").(string)
	userID := c.Locals("user_id").(string)

	// Extract query parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	status := c.Query("status", "")
	documentType := c.Query("document_type", "")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Build query
	query := db.Where("organization_id = ? AND assigned_to = ?", organizationID, userID)

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if documentType != "" {
		query = query.Where("document_type = ?", documentType)
	}

	// Get total count
	var total int64
	query.Model(&models.ApprovalTask{}).Count(&total)

	// Get tasks with pagination
	var tasks []models.ApprovalTask
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&tasks).Error; err != nil {
		log.Printf("Error fetching approval tasks: %v", err)
		return utils.SendInternalError(c, "Failed to fetch approval tasks", err)
	}

	return utils.SendPaginatedSuccess(c, tasks, "Approval tasks retrieved successfully", page, limit, total)
}

// GetApprovalTask retrieves a single approval task with full details
func (h *ApprovalHandler) GetApprovalTask(c *fiber.Ctx) error {
	taskID := c.Params("id")
	if taskID == "" {
		return utils.SendBadRequestError(c, "Task ID is required")
	}

	db := config.DB
	organizationID := c.Locals("organization_id").(string)
	userID := c.Locals("user_id").(string)

	var task models.ApprovalTask
	if err := db.Where("id = ? AND organization_id = ? AND assigned_to = ?", taskID, organizationID, userID).First(&task).Error; err != nil {
		log.Printf("Error fetching approval task %s: %v", taskID, err)
		return utils.SendNotFoundError(c, "Approval task not found or access denied")
	}

	return utils.SendSimpleSuccess(c, task, "Approval task retrieved successfully")
}

// ApproveTask marks a task as approved and moves to next stage
func (h *ApprovalHandler) ApproveTask(c *fiber.Ctx) error {
	taskID := c.Params("id")
	userID := c.Locals("user_id").(string)
	organizationID := c.Locals("organization_id").(string)

	var req ApproveTaskRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{
			Error:   "Invalid request body",
			Message: "Failed to parse approval request",
		})
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{
			Error:   "Validation failed",
			Message: err.Error(),
		})
	}

	db := config.DB

	// Get the task
	var task models.ApprovalTask
	if err := db.Where("id = ? AND organization_id = ? AND assigned_to = ?", taskID, organizationID, userID).First(&task).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(types.ErrorResponse{
			Error:   "Task not found",
			Message: "Approval task not found or access denied",
		})
	}

	// Check if task is in pending status
	if task.Status != "pending" {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{
			Error:   "Invalid task status",
			Message: "Task is not in pending status",
		})
	}

	// Start transaction
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Update task status
	now := time.Now()
	task.Status = "approved"
	task.ApprovedBy = &userID
	task.ApprovedAt = &now
	task.Signature = &req.Signature
	if req.Comment != "" {
		task.Comments = &req.Comment
	}

	if err := tx.Save(&task).Error; err != nil {
		tx.Rollback()
		log.Printf("Error updating approval task: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to approve task",
		})
	}

	// Update the document status based on document type
	switch task.DocumentType {
	case "requisition":
		if err := h.updateRequisitionStatus(tx, task.DocumentID, "approved"); err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse{
				Error:   "Update failed",
				Message: "Failed to update requisition status",
			})
		}
	case "purchase_order":
		if err := h.updatePurchaseOrderStatus(tx, task.DocumentID, "approved"); err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse{
				Error:   "Update failed",
				Message: "Failed to update purchase order status",
			})
		}
	case "payment_voucher":
		if err := h.updatePaymentVoucherStatus(tx, task.DocumentID, "approved"); err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse{
				Error:   "Update failed",
				Message: "Failed to update payment voucher status",
			})
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		log.Printf("Error committing approval transaction: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse{
			Error:   "Transaction failed",
			Message: "Failed to complete approval",
		})
	}

	return c.Status(fiber.StatusOK).JSON(types.SuccessResponse{
		Message: "Task approved successfully",
		Data:    task,
	})
}

// RejectTask marks a task as rejected and returns document to draft
func (h *ApprovalHandler) RejectTask(c *fiber.Ctx) error {
	taskID := c.Params("id")
	userID := c.Locals("user_id").(string)
	organizationID := c.Locals("organization_id").(string)

	var req RejectTaskRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{
			Error:   "Invalid request body",
			Message: "Failed to parse rejection request",
		})
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{
			Error:   "Validation failed",
			Message: err.Error(),
		})
	}

	db := config.DB

	// Get the task
	var task models.ApprovalTask
	if err := db.Where("id = ? AND organization_id = ? AND assigned_to = ?", taskID, organizationID, userID).First(&task).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(types.ErrorResponse{
			Error:   "Task not found",
			Message: "Approval task not found or access denied",
		})
	}

	// Check if task is in pending status
	if task.Status != "pending" {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{
			Error:   "Invalid task status",
			Message: "Task is not in pending status",
		})
	}

	// Start transaction
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Update task status
	now := time.Now()
	task.Status = "rejected"
	task.RejectedBy = &userID
	task.RejectedAt = &now
	task.Signature = &req.Signature
	task.RejectionReason = &req.Reason

	if err := tx.Save(&task).Error; err != nil {
		tx.Rollback()
		log.Printf("Error updating approval task: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to reject task",
		})
	}

	// Update the document status to rejected
	switch task.DocumentType {
	case "requisition":
		if err := h.updateRequisitionStatus(tx, task.DocumentID, "rejected"); err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse{
				Error:   "Update failed",
				Message: "Failed to update requisition status",
			})
		}
	case "purchase_order":
		if err := h.updatePurchaseOrderStatus(tx, task.DocumentID, "rejected"); err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse{
				Error:   "Update failed",
				Message: "Failed to update purchase order status",
			})
		}
	case "payment_voucher":
		if err := h.updatePaymentVoucherStatus(tx, task.DocumentID, "rejected"); err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse{
				Error:   "Update failed",
				Message: "Failed to update payment voucher status",
			})
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		log.Printf("Error committing rejection transaction: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse{
			Error:   "Transaction failed",
			Message: "Failed to complete rejection",
		})
	}

	return c.Status(fiber.StatusOK).JSON(types.SuccessResponse{
		Message: "Task rejected successfully",
		Data:    task,
	})
}

// ReassignTask reassigns task to different approver
func (h *ApprovalHandler) ReassignTask(c *fiber.Ctx) error {
	taskID := c.Params("id")
	organizationID := c.Locals("organization_id").(string)

	var req ReassignTaskRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{
			Error:   "Invalid request body",
			Message: "Failed to parse reassignment request",
		})
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{
			Error:   "Validation failed",
			Message: err.Error(),
		})
	}

	db := config.DB

	// Get the task
	var task models.ApprovalTask
	if err := db.Where("id = ? AND organization_id = ?", taskID, organizationID).First(&task).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(types.ErrorResponse{
			Error:   "Task not found",
			Message: "Approval task not found",
		})
	}

	// Check if task is in pending status
	if task.Status != "pending" {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{
			Error:   "Invalid task status",
			Message: "Task is not in pending status",
		})
	}

	// Update task assignment
	task.AssignedTo = req.NewUserID
	if req.Reason != "" {
		task.Comments = &req.Reason
	}

	if err := db.Save(&task).Error; err != nil {
		log.Printf("Error reassigning approval task: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to reassign task",
		})
	}

	return c.Status(fiber.StatusOK).JSON(types.SuccessResponse{
		Message: "Task reassigned successfully",
		Data:    task,
	})
}

// GetApprovalHistory retrieves approval history for a document
func (h *ApprovalHandler) GetApprovalHistory(c *fiber.Ctx) error {
	documentID := c.Params("documentId")
	organizationID := c.Locals("organization_id").(string)

	if documentID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{
			Error:   "Invalid request",
			Message: "Document ID is required",
		})
	}

	db := config.DB

	var history []models.ApprovalTask
	if err := db.Where("document_id = ? AND organization_id = ?", documentID, organizationID).
		Order("created_at ASC").Find(&history).Error; err != nil {
		log.Printf("Error fetching approval history: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to fetch approval history",
		})
	}

	return c.Status(fiber.StatusOK).JSON(history)
}

// BulkApprove approves multiple tasks at once
// POST /api/v1/approvals/bulk/approve
func (h *ApprovalHandler) BulkApprove(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	organizationID := c.Locals("organization_id").(string)

	var req BulkApproveRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendBadRequestError(c, "Invalid request body")
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		return utils.SendBadRequestError(c, "Validation failed: "+err.Error())
	}

	db := config.DB
	var successIDs []string
	var errors []string

	// Process each task
	for _, taskID := range req.TaskIDs {
		// Get the task
		var task models.ApprovalTask
		if err := db.Where("id = ? AND organization_id = ? AND assigned_to = ?", taskID, organizationID, userID).First(&task).Error; err != nil {
			errors = append(errors, "Task "+taskID+": not found or access denied")
			continue
		}

		// Check if task is in pending status
		if task.Status != "pending" {
			errors = append(errors, "Task "+taskID+": not in pending status")
			continue
		}

		// Start transaction for this task
		tx := db.Begin()

		// Update task status
		now := time.Now()
		task.Status = "approved"
		task.ApprovedBy = &userID
		task.ApprovedAt = &now
		task.Signature = &req.Signature
		if req.Comment != "" {
			task.Comments = &req.Comment
		}

		if err := tx.Save(&task).Error; err != nil {
			tx.Rollback()
			errors = append(errors, "Task "+taskID+": failed to update task")
			continue
		}

		// Update the document status based on document type
		var updateErr error
		switch task.DocumentType {
		case "requisition":
			updateErr = h.updateRequisitionStatus(tx, task.DocumentID, "approved")
		case "purchase_order":
			updateErr = h.updatePurchaseOrderStatus(tx, task.DocumentID, "approved")
		case "payment_voucher":
			updateErr = h.updatePaymentVoucherStatus(tx, task.DocumentID, "approved")
		}

		if updateErr != nil {
			tx.Rollback()
			errors = append(errors, "Task "+taskID+": failed to update document status")
			continue
		}

		// Commit transaction
		if err := tx.Commit().Error; err != nil {
			errors = append(errors, "Task "+taskID+": failed to commit transaction")
			continue
		}

		successIDs = append(successIDs, taskID)
	}

	return utils.SendSimpleSuccess(c, BulkOperationResponse{
		SuccessCount: len(successIDs),
		FailureCount: len(errors),
		SuccessIDs:   successIDs,
		Errors:       errors,
	}, "Bulk approval completed")
}

// BulkReject rejects multiple tasks at once
// POST /api/v1/approvals/bulk/reject
func (h *ApprovalHandler) BulkReject(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	organizationID := c.Locals("organization_id").(string)

	var req BulkRejectRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendBadRequestError(c, "Invalid request body")
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		return utils.SendBadRequestError(c, "Validation failed: "+err.Error())
	}

	db := config.DB
	var successIDs []string
	var errors []string

	// Process each task
	for _, taskID := range req.TaskIDs {
		// Get the task
		var task models.ApprovalTask
		if err := db.Where("id = ? AND organization_id = ? AND assigned_to = ?", taskID, organizationID, userID).First(&task).Error; err != nil {
			errors = append(errors, "Task "+taskID+": not found or access denied")
			continue
		}

		// Check if task is in pending status
		if task.Status != "pending" {
			errors = append(errors, "Task "+taskID+": not in pending status")
			continue
		}

		// Start transaction for this task
		tx := db.Begin()

		// Update task status
		now := time.Now()
		task.Status = "rejected"
		task.RejectedBy = &userID
		task.RejectedAt = &now
		task.Signature = &req.Signature
		task.RejectionReason = &req.Reason

		if err := tx.Save(&task).Error; err != nil {
			tx.Rollback()
			errors = append(errors, "Task "+taskID+": failed to update task")
			continue
		}

		// Update the document status to rejected
		var updateErr error
		switch task.DocumentType {
		case "requisition":
			updateErr = h.updateRequisitionStatus(tx, task.DocumentID, "rejected")
		case "purchase_order":
			updateErr = h.updatePurchaseOrderStatus(tx, task.DocumentID, "rejected")
		case "payment_voucher":
			updateErr = h.updatePaymentVoucherStatus(tx, task.DocumentID, "rejected")
		}

		if updateErr != nil {
			tx.Rollback()
			errors = append(errors, "Task "+taskID+": failed to update document status")
			continue
		}

		// Commit transaction
		if err := tx.Commit().Error; err != nil {
			errors = append(errors, "Task "+taskID+": failed to commit transaction")
			continue
		}

		successIDs = append(successIDs, taskID)
	}

	return utils.SendSimpleSuccess(c, BulkOperationResponse{
		SuccessCount: len(successIDs),
		FailureCount: len(errors),
		SuccessIDs:   successIDs,
		Errors:       errors,
	}, "Bulk rejection completed")
}

// BulkReassign reassigns multiple tasks at once
// POST /api/v1/approvals/bulk/reassign
func (h *ApprovalHandler) BulkReassign(c *fiber.Ctx) error {
	organizationID := c.Locals("organization_id").(string)

	var req BulkReassignRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendBadRequestError(c, "Invalid request body")
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		return utils.SendBadRequestError(c, "Validation failed: "+err.Error())
	}

	db := config.DB
	var successIDs []string
	var errors []string

	// Process each task
	for _, taskID := range req.TaskIDs {
		// Get the task
		var task models.ApprovalTask
		if err := db.Where("id = ? AND organization_id = ?", taskID, organizationID).First(&task).Error; err != nil {
			errors = append(errors, "Task "+taskID+": not found")
			continue
		}

		// Check if task is in pending status
		if task.Status != "pending" {
			errors = append(errors, "Task "+taskID+": not in pending status")
			continue
		}

		// Update task assignment
		task.AssignedTo = req.NewUserID
		if req.Reason != "" {
			task.Comments = &req.Reason
		}

		if err := db.Save(&task).Error; err != nil {
			errors = append(errors, "Task "+taskID+": failed to reassign")
			continue
		}

		successIDs = append(successIDs, taskID)
	}

	return utils.SendSimpleSuccess(c, BulkOperationResponse{
		SuccessCount: len(successIDs),
		FailureCount: len(errors),
		SuccessIDs:   successIDs,
		Errors:       errors,
	}, "Bulk reassignment completed")
}

// GetOverdueTasks retrieves tasks that are past their due date
// GET /api/v1/approvals/tasks/overdue
func (h *ApprovalHandler) GetOverdueTasks(c *fiber.Ctx) error {
	organizationID := c.Locals("organization_id").(string)

	// Get query parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	db := config.DB

	// Get overdue tasks (tasks created more than 7 days ago and still pending)
	var tasks []models.ApprovalTask
	if err := db.Where("organization_id = ? AND status = ? AND created_at < ?", 
		organizationID, "pending", time.Now().AddDate(0, 0, -7)).
		Offset(offset).Limit(limit).Order("created_at ASC").Find(&tasks).Error; err != nil {
		log.Printf("Error fetching overdue tasks: %v", err)
		return utils.SendInternalError(c, "Failed to retrieve overdue tasks", err)
	}

	// Get total count
	var total int64
	db.Model(&models.ApprovalTask{}).Where("organization_id = ? AND status = ? AND created_at < ?", 
		organizationID, "pending", time.Now().AddDate(0, 0, -7)).Count(&total)

	return utils.SendPaginatedSuccess(c, tasks, "Overdue tasks retrieved successfully", page, limit, total)
}

// Helper methods for updating document statuses
func (h *ApprovalHandler) updateRequisitionStatus(tx *gorm.DB, documentID, status string) error {
	return tx.Model(&models.Requisition{}).Where("id = ?", documentID).Update("status", status).Error
}

func (h *ApprovalHandler) updatePurchaseOrderStatus(tx *gorm.DB, documentID, status string) error {
	return tx.Model(&models.PurchaseOrder{}).Where("id = ?", documentID).Update("status", status).Error
}

func (h *ApprovalHandler) updatePaymentVoucherStatus(tx *gorm.DB, documentID, status string) error {
	return tx.Model(&models.PaymentVoucher{}).Where("id = ?", documentID).Update("status", status).Error
}