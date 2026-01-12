package handlers

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/services"
	"github.com/liyali/liyali-gateway/types"
	"github.com/liyali/liyali-gateway/utils"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// ApproverInfo represents an approver
type ApproverInfo struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

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
	Signature       string `json:"signature" validate:"required"`
	Comment         string `json:"comment"`
	ExpectedVersion int    `json:"expectedVersion"`
}

type RejectTaskRequest struct {
	Signature       string `json:"signature" validate:"required"`
	Reason          string `json:"reason" validate:"required"`
	ExpectedVersion int    `json:"expectedVersion"`
}

type ReassignTaskRequest struct {
	NewUserID string `json:"newUserId" validate:"required"`
	Reason    string `json:"reason"`
}

type ClaimTaskRequest struct {
	// No additional fields needed for claiming
}

type UnclaimTaskRequest struct {
	// No additional fields needed for unclaiming
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
	organizationID := c.Locals("organizationID").(string) // Fixed: was "organization_id"
	userID := c.Locals("userID").(string)                 // Fixed: was "user_id"

	// Extract query parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	status := c.Query("status", "")
	documentType := c.Query("document_type", "")
	assignedToMe := c.Query("assigned_to_me", "false") == "true"

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Build query
	query := db.Where("organization_id = ?", organizationID)
	
	// Filter by assigned user if requested
	if assignedToMe {
		query = query.Where("assigned_to = ?", userID)
	}

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
	organizationID := c.Locals("organizationID").(string) // Fixed: was "organizationId"
	userID := c.Locals("userID").(string)

	var task models.ApprovalTask
	if err := db.Where("id = ? AND organization_id = ? AND assigned_to = ?", taskID, organizationID, userID).First(&task).Error; err != nil {
		log.Printf("Error fetching approval task %s: %v", taskID, err)
		return utils.SendNotFoundError(c, "Approval task not found or access denied")
	}

	return utils.SendSimpleSuccess(c, task, "Approval task retrieved successfully")
}

// ClaimTask claims a workflow task for exclusive access
// POST /api/v1/approvals/tasks/:id/claim
func (h *ApprovalHandler) ClaimTask(c *fiber.Ctx) error {
	taskID := c.Params("id")
	if taskID == "" {
		return utils.SendBadRequestError(c, "Task ID is required")
	}

	userID := c.Locals("userID").(string)

	workflowExecutionService := c.Locals("workflowExecutionService").(*services.WorkflowExecutionService)

	err := workflowExecutionService.ClaimWorkflowTask(c.Context(), taskID, userID)
	if err != nil {
		log.Printf("Error claiming workflow task %s: %v", taskID, err)
		return c.Status(fiber.StatusConflict).JSON(types.ErrorResponse{
			Error:   "Claim failed",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(types.SuccessResponse{
		Message: "Task claimed successfully",
		Data:    map[string]interface{}{"taskId": taskID, "claimedBy": userID},
	})
}

// UnclaimTask releases a claimed task
// POST /api/v1/approvals/tasks/:id/unclaim
func (h *ApprovalHandler) UnclaimTask(c *fiber.Ctx) error {
	taskID := c.Params("id")
	if taskID == "" {
		return utils.SendBadRequestError(c, "Task ID is required")
	}

	userID := c.Locals("userID").(string)

	workflowExecutionService := c.Locals("workflowExecutionService").(*services.WorkflowExecutionService)

	err := workflowExecutionService.UnclaimWorkflowTask(c.Context(), taskID, userID)
	if err != nil {
		log.Printf("Error unclaiming workflow task %s: %v", taskID, err)
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{
			Error:   "Unclaim failed",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(types.SuccessResponse{
		Message: "Task unclaimed successfully",
		Data:    map[string]interface{}{"taskId": taskID},
	})
}

// ApproveTask marks a task as approved and moves to next stage
func (h *ApprovalHandler) ApproveTask(c *fiber.Ctx) error {
	taskID := c.Params("id")
	userID := c.Locals("userID").(string)

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

	// Get workflow execution service
	workflowExecutionService := c.Locals("workflowExecutionService").(*services.WorkflowExecutionService)

	// Use workflow system to approve the task with version control
	var err error
	if req.ExpectedVersion > 0 {
		err = workflowExecutionService.ApproveWorkflowTaskWithVersion(c.Context(), taskID, userID, req.Signature, req.Comment, req.ExpectedVersion)
	} else {
		err = workflowExecutionService.ApproveWorkflowTask(c.Context(), taskID, userID, req.Signature, req.Comment)
	}
	
	if err != nil {
		log.Printf("Error approving workflow task: %v", err)
		
		// Handle specific error types
		if contains(err.Error(), "version") || contains(err.Error(), "modified by another user") {
			return c.Status(fiber.StatusConflict).JSON(types.ErrorResponse{
				Error:   "Concurrent modification",
				Message: err.Error(),
			})
		}
		
		if contains(err.Error(), "claimed by another user") || contains(err.Error(), "claim has expired") {
			return c.Status(fiber.StatusConflict).JSON(types.ErrorResponse{
				Error:   "Task claim issue",
				Message: err.Error(),
			})
		}
		
		return c.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse{
			Error:   "Approval failed",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(types.SuccessResponse{
		Message: "Task approved successfully",
		Data:    map[string]interface{}{"taskId": taskID},
	})
}

// RejectTask marks a task as rejected and returns document to draft
func (h *ApprovalHandler) RejectTask(c *fiber.Ctx) error {
	taskID := c.Params("id")
	userID := c.Locals("userID").(string)

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

	// Get workflow execution service
	workflowExecutionService := c.Locals("workflowExecutionService").(*services.WorkflowExecutionService)

	// Use workflow system to reject the task with version control
	var err error
	if req.ExpectedVersion > 0 {
		err = workflowExecutionService.RejectWorkflowTaskWithVersion(c.Context(), taskID, userID, req.Signature, req.Reason, req.ExpectedVersion)
	} else {
		err = workflowExecutionService.RejectWorkflowTask(c.Context(), taskID, userID, req.Signature, req.Reason)
	}
	
	if err != nil {
		log.Printf("Error rejecting workflow task: %v", err)
		
		// Handle specific error types
		if strings.Contains(err.Error(), "version") || strings.Contains(err.Error(), "modified by another user") {
			return c.Status(fiber.StatusConflict).JSON(types.ErrorResponse{
				Error:   "Concurrent modification",
				Message: err.Error(),
			})
		}
		
		if strings.Contains(err.Error(), "claimed by another user") || strings.Contains(err.Error(), "claim has expired") {
			return c.Status(fiber.StatusConflict).JSON(types.ErrorResponse{
				Error:   "Task claim issue",
				Message: err.Error(),
			})
		}
		
		return c.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse{
			Error:   "Rejection failed",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(types.SuccessResponse{
		Message: "Task rejected successfully",
		Data:    map[string]interface{}{"taskId": taskID},
	})
}

// ReassignTask reassigns task to different approver
func (h *ApprovalHandler) ReassignTask(c *fiber.Ctx) error {
	taskID := c.Params("id")
	organizationID := c.Locals("organizationID").(string) // Fixed: was "organizationId"

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
	organizationID := c.Locals("organizationID").(string) // Fixed: was "organizationId"

	if documentID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{
			Error:   "Invalid request",
			Message: "Document ID is required",
		})
	}

	db := config.DB

	// First, try to find the actual document ID if a requisition number was provided
	var actualDocumentID string
	var requisition models.Requisition
	
	// Try to find requisition by ID or req_number
	err := db.Where("id = ? OR req_number = ?", documentID, documentID).
		First(&requisition).Error
	
	if err == nil {
		// Found requisition, use its actual ID
		actualDocumentID = requisition.ID
	} else {
		// Assume it's already a valid document ID
		actualDocumentID = documentID
	}

	var history []models.ApprovalTask
	if err := db.Where("document_id = ? AND organization_id = ?", actualDocumentID, organizationID).
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
	userID := c.Locals("userID").(string)

	var req BulkApproveRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendBadRequestError(c, "Invalid request body")
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		return utils.SendBadRequestError(c, "Validation failed: "+err.Error())
	}

	// Get workflow execution service
	workflowExecutionService := c.Locals("workflowExecutionService").(*services.WorkflowExecutionService)

	var successIDs []string
	var errors []string

	// Process each task through workflow system
	for _, taskID := range req.TaskIDs {
		err := workflowExecutionService.ApproveWorkflowTask(c.Context(), taskID, userID, req.Signature, req.Comment)
		if err != nil {
			errors = append(errors, "Task "+taskID+": "+err.Error())
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
	userID := c.Locals("userID").(string)

	var req BulkRejectRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendBadRequestError(c, "Invalid request body")
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		return utils.SendBadRequestError(c, "Validation failed: "+err.Error())
	}

	// Get workflow execution service
	workflowExecutionService := c.Locals("workflowExecutionService").(*services.WorkflowExecutionService)

	var successIDs []string
	var errors []string

	// Process each task through workflow system
	for _, taskID := range req.TaskIDs {
		err := workflowExecutionService.RejectWorkflowTask(c.Context(), taskID, userID, req.Signature, req.Reason)
		if err != nil {
			errors = append(errors, "Task "+taskID+": "+err.Error())
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
	organizationID := c.Locals("organizationID").(string) // Fixed: was "organizationId"

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
	organizationID := c.Locals("organizationID").(string) // Fixed: was "organizationId"

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

// GetApprovalWorkflowStatus retrieves the current approval workflow status for a document
// GET /api/v1/documents/{documentId}/approval-status
func (h *ApprovalHandler) GetApprovalWorkflowStatus(c *fiber.Ctx) error {
	documentID := c.Params("documentId")
	organizationID := c.Locals("organizationID").(string)
	userID := c.Locals("userID").(string)

	if documentID == "" {
		return utils.SendBadRequestError(c, "Document ID is required")
	}

	db := config.DB

	// First, try to find the actual document ID if a requisition number was provided
	var actualDocumentID string
	var requisition models.Requisition
	
	// Try to find requisition by ID or req_number
	err := db.Where("id = ? OR req_number = ?", documentID, documentID).
		First(&requisition).Error
	
	if err == nil {
		// Found requisition, use its actual ID
		actualDocumentID = requisition.ID
	} else {
		// Assume it's already a valid document ID
		actualDocumentID = documentID
	}

	// Get workflow execution service from handler registry
	workflowExecutionService := c.Locals("workflowExecutionService").(*services.WorkflowExecutionService)

	// Get workflow status with detailed stage progress
	workflowStatus, err := workflowExecutionService.GetWorkflowStatus(c.Context(), organizationID, actualDocumentID)
	if err != nil {
		log.Printf("Error fetching workflow status: %v", err)
		return utils.SendInternalError(c, "Failed to fetch workflow status", err)
	}

	// If no workflow is assigned, return basic status
	if workflowStatus.Status == "no_workflow" {
		return c.JSON(types.DetailResponse{
			Success: true,
			Data: map[string]interface{}{
				"currentStage":  0,
				"totalStages":   0,
				"status":        "no_workflow",
				"canApprove":    false,
				"canReject":     false,
				"stageProgress": []interface{}{},
			},
		})
	}

	// Get pending workflow tasks to determine if user can approve
	pendingTasks, err := workflowExecutionService.GetPendingWorkflowTasks(c.Context(), organizationID, actualDocumentID)
	if err != nil {
		log.Printf("Error fetching pending tasks: %v", err)
		// Continue without failing, just set canApprove to false
		pendingTasks = []models.WorkflowTask{}
	}

	canApprove := false
	canReject := false
	nextApprover := ""

	if len(pendingTasks) > 0 {
		currentTask := pendingTasks[0]
		
		// Get user role to check if they can approve
		var user models.User
		if err := db.Where("id = ?", userID).First(&user).Error; err == nil {
			// Check if user's role matches the required role for current task
			if currentTask.AssignedRole != nil && user.Role == *currentTask.AssignedRole {
				canApprove = true
				canReject = true
			}
		}

		// Get next approver name
		if currentTask.AssignedRole != nil {
			var approver models.User
			if err := db.Where("current_organization_id = ? AND role = ? AND active = ?", 
				organizationID, *currentTask.AssignedRole, true).First(&approver).Error; err == nil {
				nextApprover = approver.Name
			} else {
				nextApprover = fmt.Sprintf("Any %s", *currentTask.AssignedRole)
			}
		}
	}

	// Update the workflow status response with user permissions
	workflowStatus.CanApprove = canApprove
	workflowStatus.CanReject = canReject
	if nextApprover != "" {
		workflowStatus.NextApprover = nextApprover
	}

	return c.JSON(types.DetailResponse{
		Success: true,
		Data:    workflowStatus,
	})
}

// GetAvailableApprovers retrieves available approvers for a document type and stage
// GET /api/v1/approvals/available-approvers?documentType=...&stage=...
func (h *ApprovalHandler) GetAvailableApprovers(c *fiber.Ctx) error {
	organizationIDInterface := c.Locals("organizationID")
	if organizationIDInterface == nil {
		return utils.SendUnauthorizedError(c, "Organization ID not found in context")
	}
	
	organizationID, ok := organizationIDInterface.(string)
	if !ok {
		return utils.SendUnauthorizedError(c, "Invalid organization ID in context")
	}
	
	documentType := c.Query("documentType")
	entityID := c.Query("entityId") // Optional: specific entity ID to get workflow-specific approvers

	if documentType == "" {
		return utils.SendBadRequestError(c, "Document type is required")
	}

	db := config.DB

	// If entityId is provided, try to get workflow-specific approvers
	if entityID != "" {
		workflowExecutionService := c.Locals("workflowExecutionService").(*services.WorkflowExecutionService)
		
		workflowApprovers, err := workflowExecutionService.GetAvailableApproversForWorkflow(c.Context(), organizationID, entityID)
		if err == nil && len(workflowApprovers) > 0 {
			return utils.SendSuccess(c, fiber.StatusOK, workflowApprovers, "Available approvers retrieved successfully", nil)
		}
		// If workflow approvers not found, fall back to role-based approach
	}

	// Fallback to role-based approach for document type
	var roleFilters []string
	switch documentType {
	case "REQUISITION", "requisition":
		roleFilters = []string{"manager", "supervisor", "department_head", "finance"}
	case "PURCHASE_ORDER", "purchase_order":
		roleFilters = []string{"procurement", "finance", "admin"}
	case "PAYMENT_VOUCHER", "payment_voucher":
		roleFilters = []string{"finance", "accountant", "admin"}
	case "BUDGET", "budget":
		roleFilters = []string{"finance", "admin", "executive"}
	default:
		roleFilters = []string{"manager", "admin"}
	}

	// Execute query
	var approvers []ApproverInfo

	queryErr := db.Table("users").
		Select("users.id, users.name, users.email, users.role").
		Where("users.current_organization_id = ? AND users.active = ?", organizationID, true).
		Where("users.role IN ?", roleFilters).
		Find(&approvers).Error
		
	if queryErr != nil {
		log.Printf("Error fetching available approvers: %v", queryErr)
		return utils.SendInternalError(c, "Failed to fetch available approvers", queryErr)
	}

	return utils.SendSuccess(c, fiber.StatusOK, approvers, "Available approvers retrieved successfully", nil)
}

// Helper function for string contains check
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}