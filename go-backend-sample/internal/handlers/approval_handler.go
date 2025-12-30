package handlers

import (
	"fmt"
	"strconv"

	"github.com/cozyCodr/liyali-gateway/internal/db"
	"github.com/cozyCodr/liyali-gateway/internal/services"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type ApprovalHandler struct {
	approvalService *services.ApprovalService
}

func NewApprovalHandler(approvalService *services.ApprovalService) *ApprovalHandler {
	return &ApprovalHandler{
		approvalService: approvalService,
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
	NewUserID uuid.UUID `json:"newUserId" validate:"required"`
	Reason    string    `json:"reason"`
}

type CommentRequest struct {
	Comment string `json:"comment" validate:"required"`
}

type BulkApproveRequest struct {
	TaskIDs   []uuid.UUID `json:"taskIds" validate:"required,min=1"`
	Signature string      `json:"signature" validate:"required"`
	Comment   string      `json:"comment"`
}

type BulkRejectRequest struct {
	TaskIDs   []uuid.UUID `json:"taskIds" validate:"required,min=1"`
	Signature string      `json:"signature" validate:"required"`
	Reason    string      `json:"reason" validate:"required"`
}

type BulkReassignRequest struct {
	TaskIDs   []uuid.UUID `json:"taskIds" validate:"required,min=1"`
	NewUserID uuid.UUID   `json:"newUserId" validate:"required"`
	Reason    string      `json:"reason"`
}

type BulkOperationResponse struct {
	SuccessCount int         `json:"successCount"`
	FailureCount int         `json:"failureCount"`
	SuccessIDs   []uuid.UUID `json:"successIds"`
	Errors       []string    `json:"errors,omitempty"`
}

// GetTasks retrieves approval tasks for the authenticated user
// GET /api/approvals/tasks
func (h *ApprovalHandler) GetTasks(c fiber.Ctx) error {
	// Get user ID from context (set by auth middleware)
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	// Get query parameters
	status := c.Query("status", "")
	limitStr := c.Query("limit", "20")
	offsetStr := c.Query("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	// Get tasks
	tasks, err := h.approvalService.GetTasksByUser(c.Context(), userID, status, int32(limit), int32(offset))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve tasks",
		})
	}

	// Get total count
	count, err := h.approvalService.GetPendingTasksCount(c.Context(), userID)
	if err != nil {
		count = 0
	}

	return c.JSON(fiber.Map{
		"tasks": tasks,
		"total": count,
		"limit": limit,
		"offset": offset,
	})
}

// GetTaskByID retrieves a single approval task by ID
// GET /api/approvals/tasks/:id
func (h *ApprovalHandler) GetTaskByID(c fiber.Ctx) error {
	// Get task ID from params
	taskIDStr := c.Params("id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid task ID",
		})
	}

	// Get task
	task, err := h.approvalService.GetTaskByID(c.Context(), taskID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Task not found",
		})
	}

	// Get task history
	history, err := h.approvalService.GetTaskHistory(c.Context(), taskID)
	if err != nil {
		history = []db.ApprovalHistory{} // Return empty array on error
	}

	return c.JSON(fiber.Map{
		"task":    task,
		"history": history,
	})
}

// ApproveTask approves an approval task
// POST /api/approvals/tasks/:id/approve
func (h *ApprovalHandler) ApproveTask(c fiber.Ctx) error {
	// Get user ID from context
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	// Get task ID from params
	taskIDStr := c.Params("id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid task ID",
		})
	}

	// Parse request body
	var req ApproveTaskRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate request
	if req.Signature == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Signature is required",
		})
	}

	// Approve task
	task, err := h.approvalService.ApproveTask(c.Context(), taskID, userID, req.Signature, req.Comment)
	if err != nil {
		switch err {
		case services.ErrTaskNotFound:
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Task not found",
			})
		case services.ErrUnauthorized:
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "You are not authorized to approve this task",
			})
		case services.ErrAlreadyProcessed:
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "Task has already been processed",
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to approve task",
			})
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Task approved successfully",
		"task":    task,
	})
}

// RejectTask rejects an approval task
// POST /api/approvals/tasks/:id/reject
func (h *ApprovalHandler) RejectTask(c fiber.Ctx) error {
	// Get user ID from context
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	// Get task ID from params
	taskIDStr := c.Params("id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid task ID",
		})
	}

	// Parse request body
	var req RejectTaskRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate request
	if req.Signature == "" || req.Reason == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Signature and reason are required",
		})
	}

	// Reject task
	task, err := h.approvalService.RejectTask(c.Context(), taskID, userID, req.Signature, req.Reason)
	if err != nil {
		switch err {
		case services.ErrTaskNotFound:
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Task not found",
			})
		case services.ErrUnauthorized:
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "You are not authorized to reject this task",
			})
		case services.ErrAlreadyProcessed:
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "Task has already been processed",
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to reject task",
			})
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Task rejected successfully",
		"task":    task,
	})
}

// ReassignTask reassigns an approval task to another user
// POST /api/approvals/tasks/:id/reassign
func (h *ApprovalHandler) ReassignTask(c fiber.Ctx) error {
	// Get user ID from context
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	// Get task ID from params
	taskIDStr := c.Params("id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid task ID",
		})
	}

	// Parse request body
	var req ReassignTaskRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate request
	if req.NewUserID == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "New user ID is required",
		})
	}

	// Reassign task
	task, err := h.approvalService.ReassignTask(c.Context(), taskID, userID, req.NewUserID, req.Reason)
	if err != nil {
		switch err {
		case services.ErrTaskNotFound:
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Task not found",
			})
		case services.ErrUnauthorized:
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "You are not authorized to reassign this task",
			})
		case services.ErrAlreadyProcessed:
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "Task has already been processed",
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to reassign task",
			})
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Task reassigned successfully",
		"task":    task,
	})
}

// AddComment adds a comment to an approval task
// POST /api/approvals/tasks/:id/comment
func (h *ApprovalHandler) AddComment(c fiber.Ctx) error {
	// Get user ID from context
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	// Get task ID from params
	taskIDStr := c.Params("id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid task ID",
		})
	}

	// Parse request body
	var req CommentRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate request
	if req.Comment == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Comment is required",
		})
	}

	// Add comment
	err = h.approvalService.AddComment(c.Context(), taskID, userID, req.Comment)
	if err != nil {
		if err == services.ErrTaskNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Task not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to add comment",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Comment added successfully",
	})
}

// BulkApprove approves multiple tasks at once
// POST /api/approvals/bulk/approve
func (h *ApprovalHandler) BulkApprove(c fiber.Ctx) error {
	// Get user ID from context
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	// Parse request body
	var req BulkApproveRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate request
	if len(req.TaskIDs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "At least one task ID is required",
		})
	}
	if req.Signature == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Signature is required",
		})
	}

	// Bulk approve
	successIDs, errors := h.approvalService.BulkApprove(c.Context(), req.TaskIDs, userID, req.Signature, req.Comment)

	// Format errors
	errorMessages := make([]string, 0)
	for _, err := range errors {
		errorMessages = append(errorMessages, err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(BulkOperationResponse{
		SuccessCount: len(successIDs),
		FailureCount: len(errors),
		SuccessIDs:   successIDs,
		Errors:       errorMessages,
	})
}

// BulkReject rejects multiple tasks at once
// POST /api/approvals/bulk/reject
func (h *ApprovalHandler) BulkReject(c fiber.Ctx) error {
	// Get user ID from context
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	// Parse request body
	var req BulkRejectRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate request
	if len(req.TaskIDs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "At least one task ID is required",
		})
	}
	if req.Signature == "" || req.Reason == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Signature and reason are required",
		})
	}

	// Bulk reject
	successIDs, errors := h.approvalService.BulkReject(c.Context(), req.TaskIDs, userID, req.Signature, req.Reason)

	// Format errors
	errorMessages := make([]string, 0)
	for _, err := range errors {
		errorMessages = append(errorMessages, err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(BulkOperationResponse{
		SuccessCount: len(successIDs),
		FailureCount: len(errors),
		SuccessIDs:   successIDs,
		Errors:       errorMessages,
	})
}

// BulkReassign reassigns multiple tasks at once
// POST /api/approvals/bulk/reassign
func (h *ApprovalHandler) BulkReassign(c fiber.Ctx) error {
	// Get user ID from context
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	// Parse request body
	var req BulkReassignRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate request
	if len(req.TaskIDs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "At least one task ID is required",
		})
	}
	if req.NewUserID == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "New user ID is required",
		})
	}

	// Bulk reassign
	successIDs, errors := h.approvalService.BulkReassign(c.Context(), req.TaskIDs, userID, req.NewUserID, req.Reason)

	// Format errors
	errorMessages := make([]string, 0)
	for _, err := range errors {
		errorMessages = append(errorMessages, fmt.Sprint(err))
	}

	return c.Status(fiber.StatusOK).JSON(BulkOperationResponse{
		SuccessCount: len(successIDs),
		FailureCount: len(errors),
		SuccessIDs:   successIDs,
		Errors:       errorMessages,
	})
}

// GetOverdueTasks retrieves tasks that are past their due date
// GET /api/approvals/tasks/overdue
func (h *ApprovalHandler) GetOverdueTasks(c fiber.Ctx) error {
	// Get query parameters
	limitStr := c.Query("limit", "20")
	offsetStr := c.Query("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	// Get overdue tasks
	tasks, err := h.approvalService.GetOverdueTasks(c.Context(), int32(limit), int32(offset))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve overdue tasks",
		})
	}

	return c.JSON(fiber.Map{
		"tasks":  tasks,
		"count":  len(tasks),
		"limit":  limit,
		"offset": offset,
	})
}
