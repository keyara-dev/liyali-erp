package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/cozyCodr/liyali-gateway/internal/db"
	"github.com/cozyCodr/liyali-gateway/internal/repository"
	"github.com/cozyCodr/liyali-gateway/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrTaskNotFound       = errors.New("approval task not found")
	ErrDocumentNotFound   = errors.New("document not found")
	ErrUnauthorized       = errors.New("user not authorized for this action")
	ErrInvalidStage       = errors.New("invalid workflow stage")
	ErrAlreadyProcessed   = errors.New("task already processed")
	ErrInvalidStatus      = errors.New("invalid task status")
)

type ApprovalService struct {
	taskRepo    repository.ApprovalTaskRepositoryInterface
	historyRepo repository.ApprovalHistoryRepositoryInterface
	docRepo     repository.DocumentRepositoryInterface
	auditRepo   repository.AuditLogRepositoryInterface
	notifRepo   repository.NotificationRepositoryInterface
}

func NewApprovalService(
	taskRepo repository.ApprovalTaskRepositoryInterface,
	historyRepo repository.ApprovalHistoryRepositoryInterface,
	docRepo repository.DocumentRepositoryInterface,
	auditRepo repository.AuditLogRepositoryInterface,
	notifRepo repository.NotificationRepositoryInterface,
) *ApprovalService {
	return &ApprovalService{
		taskRepo:    taskRepo,
		historyRepo: historyRepo,
		docRepo:     docRepo,
		auditRepo:   auditRepo,
		notifRepo:   notifRepo,
	}
}

// ApproveTask handles the approval of a task
func (s *ApprovalService) ApproveTask(ctx context.Context, taskID, userID uuid.UUID, signature, comment string) (*db.ApprovalTask, error) {
	// Get the task
	task, err := s.taskRepo.GetApprovalTaskByID(ctx, taskID)
	if err != nil {
		return nil, ErrTaskNotFound
	}

	// Verify user is assigned to this task
	if utils.PgtypeToUUID(task.AssignedTo) != userID {
		return nil, ErrUnauthorized
	}

	// Check task status
	if task.Status != "PENDING" && task.Status != "IN_REVIEW" {
		return nil, ErrAlreadyProcessed
	}

	// Get the document
	doc, err := s.docRepo.GetDocumentByID(ctx, utils.PgtypeToUUID(task.DocumentID))
	if err != nil {
		return nil, ErrDocumentNotFound
	}

	// Create approval history entry
	_, err = s.historyRepo.CreateApprovalHistoryEntry(ctx, db.CreateApprovalHistoryEntryParams{
		TaskID:    task.ID,
		UserID:    utils.UUIDToPgtype(userID),
		Action:    "APPROVED",
		Stage:     task.CurrentStage,
		Comment:   pgtype.Text{String: comment, Valid: comment != ""},
		Signature: pgtype.Text{String: signature, Valid: signature != ""},
		IpAddress: pgtype.Text{String: "", Valid: false}, // TODO: Get from request context
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create approval history: %w", err)
	}

	// Update task status and stage
	var updatedTask *db.ApprovalTask
	if task.CurrentStage >= task.TotalStages {
		// Final approval - mark as APPROVED
		updatedTask, err = s.taskRepo.UpdateApprovalTaskStatus(ctx, db.UpdateApprovalTaskStatusParams{
			ID:     utils.UUIDToPgtype(taskID),
			Status: "APPROVED",
		})
		if err != nil {
			return nil, fmt.Errorf("failed to update task status: %w", err)
		}

		// Update document status to APPROVED
		_, err = s.docRepo.UpdateDocumentStatus(ctx, db.UpdateDocumentStatusParams{
			ID:     task.DocumentID,
			Status: "APPROVED",
		})
		if err != nil {
			return nil, fmt.Errorf("failed to update document status: %w", err)
		}

		// Create notification for document creator
		_, err = s.notifRepo.CreateNotification(ctx, db.CreateNotificationParams{
			UserID:    doc.CreatedBy,
			Type:      "TASK_APPROVED",
			Title:     "Document Approved",
			Message:   fmt.Sprintf("Your document %s has been approved", doc.DocumentNumber),
			RelatedID: task.DocumentID,
		})
		if err != nil {
			// Log error but don't fail the operation
			fmt.Printf("failed to create notification: %v\n", err)
		}
	} else {
		// Move to next stage
		updatedTask, err = s.taskRepo.UpdateApprovalTaskStage(ctx, db.UpdateApprovalTaskStageParams{
			ID:           utils.UUIDToPgtype(taskID),
			CurrentStage: task.CurrentStage + 1,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to update task stage: %w", err)
		}

		// Update document status to IN_REVIEW
		_, err = s.docRepo.UpdateDocumentStatus(ctx, db.UpdateDocumentStatusParams{
			ID:     task.DocumentID,
			Status: "IN_REVIEW",
		})
		if err != nil {
			return nil, fmt.Errorf("failed to update document status: %w", err)
		}

		// TODO: Create new task for next stage approver
		// This requires workflow stage information to determine next approver
	}

	// Create audit log
	_, err = s.auditRepo.CreateAuditLog(ctx, db.CreateAuditLogParams{
		UserID:       utils.UUIDToPgtype(userID),
		Action:       "APPROVE_TASK",
		ResourceType: "approval_task",
		ResourceID:   utils.UUIDToPgtype(taskID),
		Changes:      []byte(fmt.Sprintf(`{"action":"approved","stage":%d,"comment":"%s"}`, task.CurrentStage, comment)),
	})
	if err != nil {
		// Log error but don't fail the operation
		fmt.Printf("failed to create audit log: %v\n", err)
	}

	return updatedTask, nil
}

// RejectTask handles the rejection of a task
func (s *ApprovalService) RejectTask(ctx context.Context, taskID, userID uuid.UUID, signature, reason string) (*db.ApprovalTask, error) {
	// Get the task
	task, err := s.taskRepo.GetApprovalTaskByID(ctx, taskID)
	if err != nil {
		return nil, ErrTaskNotFound
	}

	// Verify user is assigned to this task
	if utils.PgtypeToUUID(task.AssignedTo) != userID {
		return nil, ErrUnauthorized
	}

	// Check task status
	if task.Status != "PENDING" && task.Status != "IN_REVIEW" {
		return nil, ErrAlreadyProcessed
	}

	// Get the document
	doc, err := s.docRepo.GetDocumentByID(ctx, utils.PgtypeToUUID(task.DocumentID))
	if err != nil {
		return nil, ErrDocumentNotFound
	}

	// Create approval history entry
	_, err = s.historyRepo.CreateApprovalHistoryEntry(ctx, db.CreateApprovalHistoryEntryParams{
		TaskID:    task.ID,
		UserID:    utils.UUIDToPgtype(userID),
		Action:    "REJECTED",
		Stage:     task.CurrentStage,
		Comment:   pgtype.Text{String: reason, Valid: true},
		Signature: pgtype.Text{String: signature, Valid: signature != ""},
		IpAddress: pgtype.Text{String: "", Valid: false},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create approval history: %w", err)
	}

	// Update task status to REJECTED
	updatedTask, err := s.taskRepo.UpdateApprovalTaskStatus(ctx, db.UpdateApprovalTaskStatusParams{
		ID:     utils.UUIDToPgtype(taskID),
		Status: "REJECTED",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update task status: %w", err)
	}

	// Update document status to REJECTED
	_, err = s.docRepo.UpdateDocumentStatus(ctx, db.UpdateDocumentStatusParams{
		ID:     task.DocumentID,
		Status: "REJECTED",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update document status: %w", err)
	}

	// Create notification for document creator
	_, err = s.notifRepo.CreateNotification(ctx, db.CreateNotificationParams{
		UserID:    doc.CreatedBy,
		Type:      "TASK_REJECTED",
		Title:     "Document Rejected",
		Message:   fmt.Sprintf("Your document %s has been rejected: %s", doc.DocumentNumber, reason),
		RelatedID: task.DocumentID,
	})
	if err != nil {
		fmt.Printf("failed to create notification: %v\n", err)
	}

	// Create audit log
	_, err = s.auditRepo.CreateAuditLog(ctx, db.CreateAuditLogParams{
		UserID:       utils.UUIDToPgtype(userID),
		Action:       "REJECT_TASK",
		ResourceType: "approval_task",
		ResourceID:   utils.UUIDToPgtype(taskID),
		Changes:      []byte(fmt.Sprintf(`{"action":"rejected","stage":%d,"reason":"%s"}`, task.CurrentStage, reason)),
	})
	if err != nil {
		fmt.Printf("failed to create audit log: %v\n", err)
	}

	return updatedTask, nil
}

// ReassignTask handles the reassignment of a task to another user
func (s *ApprovalService) ReassignTask(ctx context.Context, taskID, currentUserID, newUserID uuid.UUID, reason string) (*db.ApprovalTask, error) {
	// Get the task
	task, err := s.taskRepo.GetApprovalTaskByID(ctx, taskID)
	if err != nil {
		return nil, ErrTaskNotFound
	}

	// Verify current user is assigned to this task
	if utils.PgtypeToUUID(task.AssignedTo) != currentUserID {
		return nil, ErrUnauthorized
	}

	// Check task status
	if task.Status != "PENDING" && task.Status != "IN_REVIEW" {
		return nil, ErrAlreadyProcessed
	}

	// Create approval history entry
	_, err = s.historyRepo.CreateApprovalHistoryEntry(ctx, db.CreateApprovalHistoryEntryParams{
		TaskID:    task.ID,
		UserID:    utils.UUIDToPgtype(currentUserID),
		Action:    "REASSIGNED",
		Stage:     task.CurrentStage,
		Comment:   pgtype.Text{String: fmt.Sprintf("Reassigned to user %s: %s", newUserID.String(), reason), Valid: true},
		IpAddress: pgtype.Text{String: "", Valid: false},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create approval history: %w", err)
	}

	// Reassign the task
	updatedTask, err := s.taskRepo.ReassignApprovalTask(ctx, db.ReassignApprovalTaskParams{
		ID:         utils.UUIDToPgtype(taskID),
		AssignedTo: utils.UUIDToPgtype(newUserID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to reassign task: %w", err)
	}

	// Create notification for new assignee
	doc, err := s.docRepo.GetDocumentByID(ctx, utils.PgtypeToUUID(task.DocumentID))
	if err == nil {
		_, err = s.notifRepo.CreateNotification(ctx, db.CreateNotificationParams{
			UserID:    utils.UUIDToPgtype(newUserID),
			Type:      "TASK_REASSIGNED",
			Title:     "Task Reassigned to You",
			Message:   fmt.Sprintf("Document %s has been reassigned to you for approval", doc.DocumentNumber),
			RelatedID: task.DocumentID,
		})
		if err != nil {
			fmt.Printf("failed to create notification: %v\n", err)
		}
	}

	// Create audit log
	_, err = s.auditRepo.CreateAuditLog(ctx, db.CreateAuditLogParams{
		UserID:       utils.UUIDToPgtype(currentUserID),
		Action:       "REASSIGN_TASK",
		ResourceType: "approval_task",
		ResourceID:   utils.UUIDToPgtype(taskID),
		Changes:      []byte(fmt.Sprintf(`{"action":"reassigned","from":"%s","to":"%s","reason":"%s"}`, currentUserID, newUserID, reason)),
	})
	if err != nil {
		fmt.Printf("failed to create audit log: %v\n", err)
	}

	return updatedTask, nil
}

// GetTasksByUser retrieves all tasks assigned to a user
func (s *ApprovalService) GetTasksByUser(ctx context.Context, userID uuid.UUID, status string, limit, offset int32) ([]db.ApprovalTask, error) {
	if status != "" {
		return s.taskRepo.ListApprovalTasksByAssigneeAndStatus(ctx, userID, status, limit, offset)
	}
	return s.taskRepo.ListApprovalTasksByAssignee(ctx, userID, limit, offset)
}

// GetTaskByID retrieves a single task by ID
func (s *ApprovalService) GetTaskByID(ctx context.Context, taskID uuid.UUID) (*db.ApprovalTask, error) {
	return s.taskRepo.GetApprovalTaskByID(ctx, taskID)
}

// GetTaskHistory retrieves approval history for a task
func (s *ApprovalService) GetTaskHistory(ctx context.Context, taskID uuid.UUID) ([]db.ApprovalHistory, error) {
	return s.historyRepo.ListApprovalHistoryByTask(ctx, taskID)
}

// GetPendingTasksCount gets count of pending tasks for a user
func (s *ApprovalService) GetPendingTasksCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	return s.taskRepo.CountPendingApprovalTasksByAssignee(ctx, userID)
}

// BulkApprove approves multiple tasks at once
func (s *ApprovalService) BulkApprove(ctx context.Context, taskIDs []uuid.UUID, userID uuid.UUID, signature, comment string) ([]uuid.UUID, []error) {
	successIDs := make([]uuid.UUID, 0)
	errors := make([]error, 0)

	for _, taskID := range taskIDs {
		_, err := s.ApproveTask(ctx, taskID, userID, signature, comment)
		if err != nil {
			errors = append(errors, fmt.Errorf("task %s: %w", taskID, err))
		} else {
			successIDs = append(successIDs, taskID)
		}
	}

	return successIDs, errors
}

// BulkReject rejects multiple tasks at once
func (s *ApprovalService) BulkReject(ctx context.Context, taskIDs []uuid.UUID, userID uuid.UUID, signature, reason string) ([]uuid.UUID, []error) {
	successIDs := make([]uuid.UUID, 0)
	errors := make([]error, 0)

	for _, taskID := range taskIDs {
		_, err := s.RejectTask(ctx, taskID, userID, signature, reason)
		if err != nil {
			errors = append(errors, fmt.Errorf("task %s: %w", taskID, err))
		} else {
			successIDs = append(successIDs, taskID)
		}
	}

	return successIDs, errors
}

// BulkReassign reassigns multiple tasks at once
func (s *ApprovalService) BulkReassign(ctx context.Context, taskIDs []uuid.UUID, currentUserID, newUserID uuid.UUID, reason string) ([]uuid.UUID, []error) {
	successIDs := make([]uuid.UUID, 0)
	errors := make([]error, 0)

	for _, taskID := range taskIDs {
		_, err := s.ReassignTask(ctx, taskID, currentUserID, newUserID, reason)
		if err != nil {
			errors = append(errors, fmt.Errorf("task %s: %w", taskID, err))
		} else {
			successIDs = append(successIDs, taskID)
		}
	}

	return successIDs, errors
}

// GetOverdueTasks retrieves tasks that are past their due date
func (s *ApprovalService) GetOverdueTasks(ctx context.Context, limit, offset int32) ([]db.ApprovalTask, error) {
	return s.taskRepo.ListOverdueApprovalTasks(ctx, limit, offset)
}

// AddComment adds a comment to an approval task
func (s *ApprovalService) AddComment(ctx context.Context, taskID, userID uuid.UUID, comment string) error {
	// Get the task to verify it exists
	task, err := s.taskRepo.GetApprovalTaskByID(ctx, taskID)
	if err != nil {
		return ErrTaskNotFound
	}

	// Create approval history entry for the comment
	_, err = s.historyRepo.CreateApprovalHistoryEntry(ctx, db.CreateApprovalHistoryEntryParams{
		TaskID:    task.ID,
		UserID:    utils.UUIDToPgtype(userID),
		Action:    "COMMENTED",
		Stage:     task.CurrentStage,
		Comment:   pgtype.Text{String: comment, Valid: true},
		IpAddress: pgtype.Text{String: "", Valid: false},
	})
	if err != nil {
		return fmt.Errorf("failed to create comment: %w", err)
	}

	// Create notification for task assignee if commenter is not the assignee
	if utils.PgtypeToUUID(task.AssignedTo) != userID {
		_, err = s.notifRepo.CreateNotification(ctx, db.CreateNotificationParams{
			UserID:    task.AssignedTo,
			Type:      "TASK_COMMENTED",
			Title:     "New Comment on Task",
			Message:   fmt.Sprintf("A new comment has been added to your approval task"),
			RelatedID: task.ID,
		})
		if err != nil {
			fmt.Printf("failed to create notification: %v\n", err)
		}
	}

	return nil
}
