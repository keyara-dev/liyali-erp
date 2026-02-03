package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/types"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// WorkflowExecutionService handles workflow assignment and execution
type WorkflowExecutionService struct {
	db                *gorm.DB
	workflowService   *WorkflowService
	auditService      *AuditService
	automationService *DocumentAutomationService
	notificationService *NotificationService
}

// NewWorkflowExecutionService creates a new workflow execution service
func NewWorkflowExecutionService(db *gorm.DB, workflowService *WorkflowService, auditService *AuditService, automationService *DocumentAutomationService) *WorkflowExecutionService {
	return &WorkflowExecutionService{
		db:                db,
		workflowService:   workflowService,
		auditService:      auditService,
		automationService: automationService,
		notificationService: NewNotificationService(db),
	}
}

// AssignWorkflowToDocument assigns a workflow to a document and creates initial tasks
func (s *WorkflowExecutionService) AssignWorkflowToDocument(ctx context.Context, organizationID, entityID, entityType, userID string) (*models.WorkflowAssignment, error) {
	// Get the default workflow for this entity type
	workflow, err := s.workflowService.GetDefaultWorkflow(ctx, organizationID, entityType)
	if err != nil {
		return nil, fmt.Errorf("failed to get default workflow: %w", err)
	}

	// Get workflow stages
	stages, err := workflow.GetStages()
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow stages: %w", err)
	}

	if len(stages) == 0 {
		return nil, fmt.Errorf("workflow has no stages")
	}

	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create workflow assignment
	assignment := &models.WorkflowAssignment{
		ID:              uuid.New().String(),
		OrganizationID:  organizationID,
		EntityID:        entityID,
		EntityType:      entityType,
		WorkflowID:      workflow.ID,
		WorkflowVersion: workflow.Version,
		CurrentStage:    1,
		Status:          "in_progress",
		StageHistory:    datatypes.JSON{},
		AssignedAt:      time.Now(),
		AssignedBy:      userID,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := tx.Create(assignment).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create workflow assignment: %w", err)
	}

	// Create the first workflow task
	firstStage := stages[0]

	// Get priority from the document
	documentPriority := s.getDocumentPriority(tx, entityID, entityType)

	task := &models.WorkflowTask{
		ID:                   uuid.New().String(),
		OrganizationID:       organizationID,
		WorkflowAssignmentID: assignment.ID,
		EntityID:             entityID,
		EntityType:           entityType,
		StageNumber:          firstStage.StageNumber,
		StageName:            firstStage.StageName,
		AssignmentType:       "role",
		AssignedRole:         &firstStage.RequiredRole,
		Status:               "pending",
		Priority:             documentPriority,
		CreatedAt:            time.Now(),
	}

	// Set due date - calculate from stage timeout or default, then cap at document's required date
	var calculatedDueDate time.Time
	if firstStage.TimeoutHours != nil && *firstStage.TimeoutHours > 0 {
		calculatedDueDate = time.Now().Add(time.Duration(*firstStage.TimeoutHours) * time.Hour)
	} else {
		// Default due date: 7 days from creation
		calculatedDueDate = time.Now().Add(7 * 24 * time.Hour)
	}

	// Cap at document's required by date if it's earlier
	documentDueDate := s.getDocumentDueDate(tx, entityID, entityType)
	if documentDueDate != nil && documentDueDate.Before(calculatedDueDate) {
		task.DueDate = documentDueDate
	} else {
		task.DueDate = &calculatedDueDate
	}

	if err := tx.Create(task).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create workflow task: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit workflow assignment: %w", err)
	}

	// Send notification for approval required
	if s.notificationService != nil {
		notificationEvent := NotificationEvent{
			Type:         "approval_required",
			DocumentID:   entityID,
			DocumentType: entityType,
			Action:       "workflow_assigned",
			ActorID:      userID,
			Details:      fmt.Sprintf("Workflow assigned for %s approval", entityType),
			Timestamp:    time.Now(),
		}

		// Send notification asynchronously to avoid blocking
		// Use a timeout context to prevent goroutine leaks
		go func(event NotificationEvent) {
			notifyCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			select {
			case <-notifyCtx.Done():
				fmt.Printf("Notification timed out for %s\n", event.DocumentID)
				return
			default:
				if err := s.notificationService.HandleWorkflowEvent(event); err != nil {
					fmt.Printf("Failed to send approval required notification: %v\n", err)
				}
			}
		}(notificationEvent)
	}

	return assignment, nil
}

// GetWorkflowAssignment retrieves a workflow assignment for an entity
func (s *WorkflowExecutionService) GetWorkflowAssignment(ctx context.Context, organizationID, entityID string) (*models.WorkflowAssignment, error) {
	var assignment models.WorkflowAssignment
	err := s.db.Where("organization_id = ? AND entity_id = ?", organizationID, entityID).
		Preload("Workflow").
		First(&assignment).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // No workflow assigned
		}
		return nil, fmt.Errorf("failed to get workflow assignment: %w", err)
	}

	return &assignment, nil
}

// GetPendingWorkflowTasks retrieves pending workflow tasks for an entity
func (s *WorkflowExecutionService) GetPendingWorkflowTasks(ctx context.Context, organizationID, entityID string) ([]models.WorkflowTask, error) {
	var tasks []models.WorkflowTask
	err := s.db.Where("organization_id = ? AND entity_id = ? AND status = ?", organizationID, entityID, "pending").
		Order("stage_number ASC").
		Find(&tasks).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to get pending workflow tasks: %w", err)
	}

	return tasks, nil
}

// ApproveWorkflowTask approves a workflow task and progresses the workflow with optimistic locking
func (s *WorkflowExecutionService) ApproveWorkflowTask(ctx context.Context, taskID, userID, signature, comments string) error {
	return s.ApproveWorkflowTaskWithVersion(ctx, taskID, userID, signature, comments, 0)
}

// ApproveWorkflowTaskWithVersion approves a workflow task with version control for optimistic locking
func (s *WorkflowExecutionService) ApproveWorkflowTaskWithVersion(ctx context.Context, taskID, userID, signature, comments string, expectedVersion int) error {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get the task
	var task models.WorkflowTask
	if err := tx.Where("id = ?", taskID).First(&task).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("task not found: %w", err)
	}

	// Check optimistic locking if version is provided
	if expectedVersion > 0 && task.Version != expectedVersion {
		tx.Rollback()
		return fmt.Errorf("task was modified by another user (expected version %d, current version %d). Please refresh and try again", expectedVersion, task.Version)
	}

	// Check task status
	if task.Status != "pending" && task.Status != "claimed" {
		tx.Rollback()
		return fmt.Errorf("task is not in pending or claimed status (current: %s)", task.Status)
	}

	// Check task is claimed by this user (if claiming is enabled)
	if task.ClaimedBy != nil && *task.ClaimedBy != userID {
		tx.Rollback()
		return fmt.Errorf("task is claimed by another user, please wait for them to complete or unclaim it")
	}

	// Check claim hasn't expired
	if task.ClaimExpiry != nil && time.Now().After(*task.ClaimExpiry) {
		tx.Rollback()
		return fmt.Errorf("task claim has expired, please reclaim the task")
	}

	// Validate user has the required role for this task
	var user models.User
	if err := tx.Where("id = ?", userID).First(&user).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("user not found: %w", err)
	}

	// Built-in roles that have approval permissions
	approverRoles := []string{"admin", "approver", "finance", "manager", "supervisor", "department_head"}

	// Approval-related permissions to check for in organization roles
	approvalPermissions := []string{
		"requisition.approve", "approval.approve", "budget.approve",
		"purchase_order.approve", "payment_voucher.approve", "grn.approve",
	}

	// Helper function to check if user has any organization role with approval permissions
	checkOrgRoleApprovalPermissions := func() bool {
		var userOrgRoles []models.UserOrganizationRole
		if err := tx.Where("user_id = ? AND organization_id = ? AND active = ?",
			userID, task.OrganizationID, true).Find(&userOrgRoles).Error; err != nil || len(userOrgRoles) == 0 {
			return false
		}

		for _, userOrgRole := range userOrgRoles {
			var orgRole models.OrganizationRole
			if err := tx.Where("id = ? AND active = ?", userOrgRole.RoleID, true).First(&orgRole).Error; err != nil {
				continue
			}

			// Parse permissions from JSON
			var permissions []string
			if err := json.Unmarshal(orgRole.Permissions, &permissions); err != nil {
				continue
			}

			// Check if any approval permission exists
			for _, perm := range permissions {
				for _, approvalPerm := range approvalPermissions {
					if strings.EqualFold(perm, approvalPerm) {
						return true
					}
				}
			}
		}
		return false
	}

	// PRIORITY 1: If task is assigned to a specific user (after reassignment), ONLY that user can approve
	if task.AssignedUserID != nil && *task.AssignedUserID != "" {
		if *task.AssignedUserID != userID {
			tx.Rollback()
			return fmt.Errorf("insufficient permissions: this task has been assigned to a specific user and only they can approve it")
		}
		// User is the assigned user - permission granted, skip role checks
		log.Printf("[DEBUG] User %s is the specifically assigned user for this task - permission granted", userID)
	} else if task.AssignedRole != nil {
		// PRIORITY 2: Check role-based permissions (when task is assigned to a role, not a specific user)
		assignedRole := *task.AssignedRole
		hasPermission := false

		// Check if assignedRole is a UUID (custom organization role)
		if _, parseErr := uuid.Parse(assignedRole); parseErr == nil {
			// It's a UUID - check if user has this organization role
			var userOrgRole models.UserOrganizationRole
			if err := tx.Where("user_id = ? AND organization_id = ? AND role_id = ? AND active = ?",
				userID, task.OrganizationID, assignedRole, true).First(&userOrgRole).Error; err == nil {
				hasPermission = true
			} else {
				// Fallback 1: Check if user has a built-in approver role
				for _, approverRole := range approverRoles {
					if strings.EqualFold(user.Role, approverRole) {
						hasPermission = true
						break
					}
				}
				// Fallback 2: Check if user has any organization role with approval permissions
				if !hasPermission {
					hasPermission = checkOrgRoleApprovalPermissions()
				}
			}
		} else {
			// It's a built-in role name - check user.Role directly (case-insensitive)
			if strings.EqualFold(user.Role, assignedRole) {
				hasPermission = true
			} else {
				// Fallback 1: Check if user has a built-in approver role
				for _, approverRole := range approverRoles {
					if strings.EqualFold(user.Role, approverRole) {
						hasPermission = true
						break
					}
				}
				// Fallback 2: Check if user has any organization role with approval permissions
				if !hasPermission {
					hasPermission = checkOrgRoleApprovalPermissions()
				}
			}
		}

		if !hasPermission {
			tx.Rollback()
			return fmt.Errorf("insufficient permissions: user does not have the required role '%s'", assignedRole)
		}
	}

	// Get the workflow assignment
	var assignment models.WorkflowAssignment
	if err := tx.Where("id = ?", task.WorkflowAssignmentID).Preload("Workflow").First(&assignment).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("workflow assignment not found: %w", err)
	}

	// Get workflow stages
	stages, err := assignment.Workflow.GetStages()
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get workflow stages: %w", err)
	}

	currentStage := stages[task.StageNumber-1]

	// Record this approval in stage approval records
	now := time.Now()
	approvalRecord := &models.StageApprovalRecord{
		ID:               uuid.New().String(),
		OrganizationID:   assignment.OrganizationID,
		WorkflowTaskID:   taskID,
		StageNumber:      task.StageNumber,
		ApproverID:       userID,
		ApproverName:     user.Name,
		ApproverRole:     user.Role,
		Action:           "approved",
		Comments:         comments,
		Signature:        signature,
		ApprovedAt:       now,
		CreatedAt:        now,
	}

	if err := tx.Create(approvalRecord).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to record approval: %w", err)
	}

	// Check if stage completion criteria are met
	stageComplete, err := s.checkStageCompletionCriteria(tx, taskID, currentStage, assignment.OrganizationID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to check stage completion: %w", err)
	}

	if stageComplete {
		// Update task with version increment
		result := tx.Model(&task).
			Where("id = ? AND version = ?", taskID, task.Version).
			Updates(map[string]interface{}{
				"status":       "completed",
				"completed_at": now,
				"updated_by":   userID,
				"version":      task.Version + 1,
			})

		if result.Error != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update task: %w", result.Error)
		}

		if result.RowsAffected == 0 {
			tx.Rollback()
			return fmt.Errorf("task was modified by another user, please refresh and try again")
		}

		// Add stage execution to history
		stageExecution := models.StageExecution{
			StageNumber:  task.StageNumber,
			StageName:    task.StageName,
			ApproverID:   userID,
			ApproverName: user.Name,
			ApproverRole: user.Role,
			Action:       "approved",
			Comments:     comments,
			Signature:    signature,
			ExecutedAt:   now,
		}

		if err := assignment.AddStageExecution(stageExecution); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update stage history: %w", err)
		}

		// Add action history entry for this stage approval
		actionMessage := fmt.Sprintf("Stage %d (%s) approved by %s", task.StageNumber, task.StageName, user.Name)
		if err := s.addActionHistoryEntry(tx, assignment.EntityType, assignment.EntityID, userID, "STAGE_APPROVED", actionMessage); err != nil {
			// Log error but don't fail the approval
			fmt.Printf("Warning: failed to add action history entry for stage approval: %v\n", err)
		}

		// Check if this is the last stage
		workflowCompleted := task.StageNumber >= len(stages)
		
		if workflowCompleted {
			// Workflow completed
			assignment.Status = "completed"
			assignment.CompletedAt = &now
			assignment.CurrentStage = len(stages)
			
			// Update the actual document status to "approved"
			if err := s.updateDocumentStatus(tx, assignment.EntityType, assignment.EntityID, "approved"); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update document status: %w", err)
			}
			
			// Add action history entry to the document
			if err := s.addActionHistoryEntry(tx, assignment.EntityType, assignment.EntityID, userID, "WORKFLOW_COMPLETED", "Document approved through workflow system"); err != nil {
				// Log error but don't fail the approval
				fmt.Printf("Warning: failed to add action history entry: %v\n", err)
			}
		} else {
			// Move to next stage
			nextStageNumber := task.StageNumber + 1
			nextStage := stages[nextStageNumber-1]

			assignment.CurrentStage = nextStageNumber

			// Get priority from document for next task
			nextTaskPriority := s.getDocumentPriority(tx, assignment.EntityID, assignment.EntityType)

			// Create next workflow task
			nextTask := &models.WorkflowTask{
				ID:                   uuid.New().String(),
				OrganizationID:       assignment.OrganizationID,
				WorkflowAssignmentID: assignment.ID,
				EntityID:             assignment.EntityID,
				EntityType:           assignment.EntityType,
				StageNumber:          nextStage.StageNumber,
				StageName:            nextStage.StageName,
				AssignmentType:       "role",
				AssignedRole:         &nextStage.RequiredRole,
				Status:               "pending",
				Priority:             nextTaskPriority,
				Version:              1,
				CreatedAt:            time.Now(),
			}

			// Set due date - calculate from stage timeout or default, then cap at document's required date
			var nextCalculatedDueDate time.Time
			if nextStage.TimeoutHours != nil && *nextStage.TimeoutHours > 0 {
				nextCalculatedDueDate = time.Now().Add(time.Duration(*nextStage.TimeoutHours) * time.Hour)
			} else {
				// Default due date: 7 days from creation
				nextCalculatedDueDate = time.Now().Add(7 * 24 * time.Hour)
			}

			// Cap at document's required by date if it's earlier
			nextDocDueDate := s.getDocumentDueDate(tx, assignment.EntityID, assignment.EntityType)
			if nextDocDueDate != nil && nextDocDueDate.Before(nextCalculatedDueDate) {
				nextTask.DueDate = nextDocDueDate
			} else {
				nextTask.DueDate = &nextCalculatedDueDate
			}

			if err := tx.Create(nextTask).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create next workflow task: %w", err)
			}
		}

		// Update assignment
		assignment.UpdatedAt = time.Now()
		if err := tx.Save(&assignment).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update workflow assignment: %w", err)
		}

		// Commit transaction
		if err := tx.Commit().Error; err != nil {
			return fmt.Errorf("failed to commit workflow approval: %w", err)
		}

		// Trigger post-approval automation if workflow completed
		if workflowCompleted {
			s.handleWorkflowCompletion(ctx, assignment, userID)
		} else {
			s.handleStageProgression(ctx, assignment, userID)
		}
	} else {
		// Stage not complete yet, update task status to partially approved
		result := tx.Model(&task).
			Where("id = ? AND version = ?", taskID, task.Version).
			Updates(map[string]interface{}{
				"status":     "partially_approved",
				"updated_by": userID,
				"version":    task.Version + 1,
			})

		if result.Error != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update task status: %w", result.Error)
		}

		if result.RowsAffected == 0 {
			tx.Rollback()
			return fmt.Errorf("task was modified by another user, please refresh and try again")
		}

		// Commit transaction
		if err := tx.Commit().Error; err != nil {
			return fmt.Errorf("failed to commit partial approval: %w", err)
		}

		// Send notification for partial approval
		s.handlePartialApproval(ctx, assignment, userID, currentStage)
	}

	return nil
}

// Helper methods for handling workflow events
func (s *WorkflowExecutionService) handleWorkflowCompletion(ctx context.Context, assignment models.WorkflowAssignment, userID string) {
	// Send approval notification
	if s.notificationService != nil {
		notificationEvent := NotificationEvent{
			Type:         "document_approved",
			DocumentID:   assignment.EntityID,
			DocumentType: assignment.EntityType,
			Action:       "workflow_completed",
			ActorID:      userID,
			Details:      "Document has been fully approved through workflow",
			Timestamp:    time.Now(),
		}
		
		go func(event NotificationEvent) {
			notifyCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			select {
			case <-notifyCtx.Done():
				return
			default:
				if err := s.notificationService.HandleWorkflowEvent(event); err != nil {
					fmt.Printf("Failed to send approval notification: %v\n", err)
				}
			}
		}(notificationEvent)
	}

	// Trigger automation
	if err := s.triggerPostApprovalAutomation(ctx, assignment.EntityType, assignment.EntityID); err != nil {
		fmt.Printf("Post-approval automation failed for %s %s: %v\n", assignment.EntityType, assignment.EntityID, err)
	}
}

func (s *WorkflowExecutionService) handleStageProgression(ctx context.Context, assignment models.WorkflowAssignment, userID string) {
	if s.notificationService != nil {
		notificationEvent := NotificationEvent{
			Type:         "approval_required",
			DocumentID:   assignment.EntityID,
			DocumentType: assignment.EntityType,
			Action:       "next_stage_approval",
			ActorID:      userID,
			Details:      fmt.Sprintf("Document moved to next approval stage (%d)", assignment.CurrentStage),
			Timestamp:    time.Now(),
		}

		go func(event NotificationEvent) {
			notifyCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			select {
			case <-notifyCtx.Done():
				return
			default:
				if err := s.notificationService.HandleWorkflowEvent(event); err != nil {
					fmt.Printf("Failed to send next stage approval notification: %v\n", err)
				}
			}
		}(notificationEvent)
	}
}

func (s *WorkflowExecutionService) handlePartialApproval(ctx context.Context, assignment models.WorkflowAssignment, userID string, stage models.WorkflowStage) {
	if s.notificationService != nil {
		notificationEvent := NotificationEvent{
			Type:         "partial_approval",
			DocumentID:   assignment.EntityID,
			DocumentType: assignment.EntityType,
			Action:       "partial_stage_approval",
			ActorID:      userID,
			Details:      fmt.Sprintf("Partial approval received for stage %d (%s)", stage.StageNumber, stage.StageName),
			Timestamp:    time.Now(),
		}

		go func(event NotificationEvent) {
			notifyCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			select {
			case <-notifyCtx.Done():
				return
			default:
				if err := s.notificationService.HandleWorkflowEvent(event); err != nil {
					fmt.Printf("Failed to send partial approval notification: %v\n", err)
				}
			}
		}(notificationEvent)
	}
}

// RejectWorkflowTask rejects a workflow task and marks the workflow as rejected
func (s *WorkflowExecutionService) RejectWorkflowTask(ctx context.Context, taskID, userID, signature, reason string) error {
	return s.RejectWorkflowTaskWithVersion(ctx, taskID, userID, signature, reason, 0)
}

// RejectWorkflowTaskWithVersion rejects a workflow task with version control for optimistic locking
func (s *WorkflowExecutionService) RejectWorkflowTaskWithVersion(ctx context.Context, taskID, userID, signature, reason string, expectedVersion int) error {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get the task
	var task models.WorkflowTask
	if err := tx.Where("id = ?", taskID).First(&task).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("task not found: %w", err)
	}

	// Check optimistic locking if version is provided
	if expectedVersion > 0 && task.Version != expectedVersion {
		tx.Rollback()
		return fmt.Errorf("task was modified by another user (expected version %d, current version %d). Please refresh and try again", expectedVersion, task.Version)
	}

	// Check task status
	if task.Status != "pending" && task.Status != "claimed" {
		tx.Rollback()
		return fmt.Errorf("task is not in pending or claimed status (current: %s)", task.Status)
	}

	// Check task is claimed by this user (if claiming is enabled)
	if task.ClaimedBy != nil && *task.ClaimedBy != userID {
		tx.Rollback()
		return fmt.Errorf("task is claimed by another user, please wait for them to complete or unclaim it")
	}

	// Check claim hasn't expired
	if task.ClaimExpiry != nil && time.Now().After(*task.ClaimExpiry) {
		tx.Rollback()
		return fmt.Errorf("task claim has expired, please reclaim the task")
	}

	// Validate user has the required role for this task
	var user models.User
	if err := tx.Where("id = ?", userID).First(&user).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("user not found: %w", err)
	}

	// Built-in roles that have approval/rejection permissions
	approverRoles := []string{"admin", "approver", "finance", "manager", "supervisor", "department_head"}

	// Approval-related permissions to check for in organization roles
	approvalPermissions := []string{
		"requisition.approve", "approval.approve", "budget.approve",
		"purchase_order.approve", "payment_voucher.approve", "grn.approve",
	}

	// Helper function to check if user has any organization role with approval permissions
	checkOrgRoleApprovalPermissions := func() bool {
		var userOrgRoles []models.UserOrganizationRole
		if err := tx.Where("user_id = ? AND organization_id = ? AND active = ?",
			userID, task.OrganizationID, true).Find(&userOrgRoles).Error; err != nil || len(userOrgRoles) == 0 {
			return false
		}

		for _, userOrgRole := range userOrgRoles {
			var orgRole models.OrganizationRole
			if err := tx.Where("id = ? AND active = ?", userOrgRole.RoleID, true).First(&orgRole).Error; err != nil {
				continue
			}

			// Parse permissions from JSON
			var permissions []string
			if err := json.Unmarshal(orgRole.Permissions, &permissions); err != nil {
				continue
			}

			// Check if any approval permission exists
			for _, perm := range permissions {
				for _, approvalPerm := range approvalPermissions {
					if strings.EqualFold(perm, approvalPerm) {
						return true
					}
				}
			}
		}
		return false
	}

	// PRIORITY 1: If task is assigned to a specific user (after reassignment), ONLY that user can reject
	if task.AssignedUserID != nil && *task.AssignedUserID != "" {
		if *task.AssignedUserID != userID {
			tx.Rollback()
			return fmt.Errorf("insufficient permissions: this task has been assigned to a specific user and only they can reject it")
		}
		// User is the assigned user - permission granted, skip role checks
		log.Printf("[DEBUG] User %s is the specifically assigned user for this task - permission granted for rejection", userID)
	} else if task.AssignedRole != nil {
		// PRIORITY 2: Check role-based permissions (when task is assigned to a role, not a specific user)
		assignedRole := *task.AssignedRole
		hasPermission := false

		// Check if assignedRole is a UUID (custom organization role)
		if _, parseErr := uuid.Parse(assignedRole); parseErr == nil {
			// It's a UUID - check if user has this organization role
			var userOrgRole models.UserOrganizationRole
			if err := tx.Where("user_id = ? AND organization_id = ? AND role_id = ? AND active = ?",
				userID, task.OrganizationID, assignedRole, true).First(&userOrgRole).Error; err == nil {
				hasPermission = true
			} else {
				// Fallback 1: Check if user has a built-in approver role
				for _, approverRole := range approverRoles {
					if strings.EqualFold(user.Role, approverRole) {
						hasPermission = true
						break
					}
				}
				// Fallback 2: Check if user has any organization role with approval permissions
				if !hasPermission {
					hasPermission = checkOrgRoleApprovalPermissions()
				}
			}
		} else {
			// It's a built-in role name - check user.Role directly (case-insensitive)
			if strings.EqualFold(user.Role, assignedRole) {
				hasPermission = true
			} else {
				// Fallback 1: Check if user has a built-in approver role
				for _, approverRole := range approverRoles {
					if strings.EqualFold(user.Role, approverRole) {
						hasPermission = true
						break
					}
				}
				// Fallback 2: Check if user has any organization role with approval permissions
				if !hasPermission {
					hasPermission = checkOrgRoleApprovalPermissions()
				}
			}
		}

		if !hasPermission {
			tx.Rollback()
			return fmt.Errorf("insufficient permissions: user does not have the required role '%s'", assignedRole)
		}
	}

	// Get the workflow assignment
	var assignment models.WorkflowAssignment
	if err := tx.Where("id = ?", task.WorkflowAssignmentID).First(&assignment).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("workflow assignment not found: %w", err)
	}

	// Update task with version increment
	now := time.Now()
	result := tx.Model(&task).
		Where("id = ? AND version = ?", taskID, task.Version).
		Updates(map[string]interface{}{
			"status":       "completed",
			"completed_at": now,
			"updated_by":   userID,
			"version":      task.Version + 1,
		})

	if result.Error != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update task: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		return fmt.Errorf("task was modified by another user, please refresh and try again")
	}

	// Record this rejection in stage approval records
	approvalRecord := &models.StageApprovalRecord{
		ID:               uuid.New().String(),
		OrganizationID:   assignment.OrganizationID,
		WorkflowTaskID:   taskID,
		StageNumber:      task.StageNumber,
		ApproverID:       userID,
		ApproverName:     user.Name,
		ApproverRole:     user.Role,
		Action:           "rejected",
		Comments:         reason,
		Signature:        signature,
		ApprovedAt:       now,
		CreatedAt:        now,
	}

	if err := tx.Create(approvalRecord).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to record rejection: %w", err)
	}

	// Add stage execution to history
	stageExecution := models.StageExecution{
		StageNumber:  task.StageNumber,
		StageName:    task.StageName,
		ApproverID:   userID,
		ApproverName: user.Name,
		ApproverRole: user.Role,
		Action:       "rejected",
		Comments:     reason,
		Signature:    signature,
		ExecutedAt:   now,
	}

	if err := assignment.AddStageExecution(stageExecution); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update stage history: %w", err)
	}

	// Mark workflow as rejected
	assignment.Status = "rejected"
	assignment.CompletedAt = &now
	assignment.UpdatedAt = time.Now()

	// Update the actual document status to "rejected"
	if err := s.updateDocumentStatus(tx, assignment.EntityType, assignment.EntityID, "rejected"); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update document status: %w", err)
	}
	
	// Add action history entry to the document
	if err := s.addActionHistoryEntry(tx, assignment.EntityType, assignment.EntityID, userID, "WORKFLOW_REJECTED", reason); err != nil {
		// Log error but don't fail the rejection
		fmt.Printf("Warning: failed to add action history entry: %v\n", err)
	}

	if err := tx.Save(&assignment).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update workflow assignment: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit workflow rejection: %w", err)
	}

	// Send rejection notification
	if s.notificationService != nil {
		notificationEvent := NotificationEvent{
			Type:         "document_rejected",
			DocumentID:   assignment.EntityID,
			DocumentType: assignment.EntityType,
			Action:       "workflow_rejected",
			ActorID:      userID,
			Details:      reason,
			Timestamp:    time.Now(),
		}

		// Send notification asynchronously with timeout context
		go func(event NotificationEvent) {
			notifyCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			select {
			case <-notifyCtx.Done():
				return
			default:
				if err := s.notificationService.HandleWorkflowEvent(event); err != nil {
					fmt.Printf("Failed to send rejection notification: %v\n", err)
				}
			}
		}(notificationEvent)
	}

	return nil
}

// GetWorkflowStatus returns the current workflow status for an entity
func (s *WorkflowExecutionService) GetWorkflowStatus(ctx context.Context, organizationID, entityID string) (*WorkflowStatusResponse, error) {
	assignment, err := s.GetWorkflowAssignment(ctx, organizationID, entityID)
	if err != nil {
		return nil, err
	}

	if assignment == nil {
		return &WorkflowStatusResponse{
			CurrentStage: 0,
			TotalStages:  0,
			Status:       "no_workflow",
			CanApprove:   false,
			CanReject:    false,
		}, nil
	}

	// Get workflow stages
	stages, err := assignment.Workflow.GetStages()
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow stages: %w", err)
	}

	// Get pending tasks
	pendingTasks, err := s.GetPendingWorkflowTasks(ctx, organizationID, entityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending tasks: %w", err)
	}

	// Get stage history for detailed tracking
	stageHistory, err := assignment.GetStageHistory()
	if err != nil {
		stageHistory = []models.StageExecution{}
	}

	response := &WorkflowStatusResponse{
		CurrentStage:  assignment.CurrentStage,
		TotalStages:   len(stages),
		Status:        assignment.Status,
		CanApprove:    false,
		CanReject:     false,
		StageProgress: make([]StageProgressInfo, len(stages)),
	}

	// Build detailed stage progress information
	for i, stage := range stages {
		// Resolve role name if RequiredRole is a UUID
		requiredRoleDisplay := stage.RequiredRole
		if _, parseErr := uuid.Parse(stage.RequiredRole); parseErr == nil {
			var orgRole models.OrganizationRole
			if err := s.db.Where("id = ?", stage.RequiredRole).First(&orgRole).Error; err == nil {
				requiredRoleDisplay = orgRole.Name
			}
		}

		stageInfo := StageProgressInfo{
			StageNumber:   stage.StageNumber,
			StageName:     stage.StageName,
			RequiredRole:  requiredRoleDisplay,
			Status:        "pending",
			IsCurrentStage: stage.StageNumber == assignment.CurrentStage,
		}

		// Check if this stage has been completed
		for _, execution := range stageHistory {
			if execution.StageNumber == stage.StageNumber {
				stageInfo.Status = execution.Action // "approved" or "rejected"
				stageInfo.ApproverID = execution.ApproverID
				stageInfo.ApproverName = execution.ApproverName
				stageInfo.ApproverRole = execution.ApproverRole
				stageInfo.CompletedAt = &execution.ExecutedAt
				stageInfo.Comments = execution.Comments
				break
			}
		}

		// Mark stages before current as completed if not found in history
		if stage.StageNumber < assignment.CurrentStage && stageInfo.Status == "pending" {
			stageInfo.Status = "completed"
		}

		response.StageProgress[i] = stageInfo
	}

	// Check if user can approve current stage
	if len(pendingTasks) > 0 {
		currentTask := pendingTasks[0]
		if currentTask.AssignedRole != nil {
			assignedRole := *currentTask.AssignedRole
			roleDisplayName := assignedRole

			// Check if assignedRole is a UUID (custom organization role)
			if _, parseErr := uuid.Parse(assignedRole); parseErr == nil {
				// It's a UUID - look up the organization role name
				var orgRole models.OrganizationRole
				if err := s.db.Where("id = ?", assignedRole).First(&orgRole).Error; err == nil {
					roleDisplayName = orgRole.Name
				}
			}

			response.NextApprover = fmt.Sprintf("Required Role: %s", roleDisplayName)
		}
	}

	return response, nil
}

// WorkflowStatusResponse represents the workflow status
type WorkflowStatusResponse struct {
	CurrentStage  int                 `json:"currentStage"`
	TotalStages   int                 `json:"totalStages"`
	Status        string              `json:"status"`
	NextApprover  string              `json:"nextApprover,omitempty"`
	CanApprove    bool                `json:"canApprove"`
	CanReject     bool                `json:"canReject"`
	StageProgress []StageProgressInfo `json:"stageProgress"`
}

// StageProgressInfo represents detailed information about each workflow stage
type StageProgressInfo struct {
	StageNumber    int        `json:"stageNumber"`
	StageName      string     `json:"stageName"`
	RequiredRole   string     `json:"requiredRole"`
	Status         string     `json:"status"` // "pending", "approved", "rejected", "completed"
	IsCurrentStage bool       `json:"isCurrentStage"`
	ApproverID     string     `json:"approverId,omitempty"`
	ApproverName   string     `json:"approverName,omitempty"`
	ApproverRole   string     `json:"approverRole,omitempty"`
	CompletedAt    *time.Time `json:"completedAt,omitempty"`
	Comments       string     `json:"comments,omitempty"`
}

// GetAvailableApproversForWorkflow returns available approvers for the current workflow stage
func (s *WorkflowExecutionService) GetAvailableApproversForWorkflow(ctx context.Context, organizationID, entityID string) ([]ApproverInfo, error) {
	// Get pending tasks
	pendingTasks, err := s.GetPendingWorkflowTasks(ctx, organizationID, entityID)
	if err != nil {
		return nil, err
	}

	if len(pendingTasks) == 0 {
		return []ApproverInfo{}, nil
	}

	// Get the current pending task
	currentTask := pendingTasks[0]
	
	if currentTask.AssignedRole == nil {
		return []ApproverInfo{}, nil
	}

	// Query users with the required role
	var approvers []ApproverInfo
	err = s.db.Table("users").
		Select("users.id, users.name, users.email, users.role").
		Where("users.current_organization_id = ? AND users.active = ? AND users.role = ?", 
			organizationID, true, *currentTask.AssignedRole).
		Find(&approvers).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get available approvers: %w", err)
	}

	return approvers, nil
}

// ClaimWorkflowTask claims a workflow task for exclusive access
func (s *WorkflowExecutionService) ClaimWorkflowTask(ctx context.Context, taskID, userID string) error {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var task models.WorkflowTask

	// Atomic claim operation with optimistic locking
	result := tx.Model(&task).
		Where("id = ? AND status = ? AND (claimed_by IS NULL OR claim_expiry < ?)",
			taskID, "pending", time.Now()).
		Updates(map[string]interface{}{
			"claimed_by":    userID,
			"claimed_at":    time.Now(),
			"claim_expiry":  time.Now().Add(30 * time.Minute), // 30-minute claim
			"status":        "claimed",
			"version":       gorm.Expr("version + 1"),
		})

	if result.Error != nil {
		tx.Rollback()
		return fmt.Errorf("failed to claim task: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		return fmt.Errorf("task is not available for claiming (already claimed or completed)")
	}

	return tx.Commit().Error
}

// UnclaimWorkflowTask releases a claimed task
func (s *WorkflowExecutionService) UnclaimWorkflowTask(ctx context.Context, taskID, userID string) error {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var task models.WorkflowTask
	if err := tx.Where("id = ?", taskID).First(&task).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("task not found: %w", err)
	}

	// Check if task is claimed by this user
	if task.ClaimedBy == nil || *task.ClaimedBy != userID {
		tx.Rollback()
		return fmt.Errorf("task is not claimed by you or is not claimed at all")
	}

	// Release the claim
	result := tx.Model(&task).
		Where("id = ? AND claimed_by = ?", taskID, userID).
		Updates(map[string]interface{}{
			"claimed_by":    nil,
			"claimed_at":    nil,
			"claim_expiry":  nil,
			"status":        "pending",
			"version":       gorm.Expr("version + 1"),
		})

	if result.Error != nil {
		tx.Rollback()
		return fmt.Errorf("failed to unclaim task: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		return fmt.Errorf("task was not found or not claimed by you")
	}

	return tx.Commit().Error
}

// checkStageCompletionCriteria checks if a workflow stage has met its completion criteria
func (s *WorkflowExecutionService) checkStageCompletionCriteria(tx *gorm.DB, taskID string, stage models.WorkflowStage, organizationID string) (bool, error) {
	// Get all approvals for this stage
	var approvals []models.StageApprovalRecord
	if err := tx.Where("workflow_task_id = ? AND stage_number = ? AND action = ?",
		taskID, stage.StageNumber, "approved").Find(&approvals).Error; err != nil {
		return false, err
	}

	approvalCount := len(approvals)

	// If no approval type specified, default to single approval
	if stage.ApprovalType == "" {
		return approvalCount >= 1, nil
	}

	switch stage.ApprovalType {
	case "any":
		return approvalCount >= 1, nil
	case "all":
		// Get total number of qualified users for this role in the organization
		var totalQualified int64
		if err := tx.Model(&models.User{}).
			Where("current_organization_id = ? AND role = ? AND active = ?", organizationID, stage.RequiredRole, true).
			Count(&totalQualified).Error; err != nil {
			return false, err
		}
		return approvalCount >= int(totalQualified), nil
	case "majority":
		// Get total number of qualified users for this role in the organization
		var totalQualified int64
		if err := tx.Model(&models.User{}).
			Where("current_organization_id = ? AND role = ? AND active = ?", organizationID, stage.RequiredRole, true).
			Count(&totalQualified).Error; err != nil {
			return false, err
		}
		required := int(totalQualified)/2 + 1
		return approvalCount >= required, nil
	case "quorum":
		if stage.QuorumCount == nil {
			return false, fmt.Errorf("quorum count not specified for quorum-based approval")
		}
		return approvalCount >= *stage.QuorumCount, nil
	default:
		// Default: require specified count (RequiredApprovalCount)
		requiredCount := stage.RequiredApprovalCount
		if requiredCount <= 0 {
			requiredCount = 1 // Default to 1 if not specified
		}
		return approvalCount >= requiredCount, nil
	}
}

// ApproverInfo represents an approver
type ApproverInfo struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

// updateDocumentStatus updates the status of the actual document when workflow completes
func (s *WorkflowExecutionService) updateDocumentStatus(tx *gorm.DB, entityType, entityID, newStatus string) error {
	switch entityType {
	case "REQUISITION", "requisition":
		return tx.Model(&models.Requisition{}).Where("id = ?", entityID).Update("status", newStatus).Error
	case "BUDGET", "budget":
		return tx.Model(&models.Budget{}).Where("id = ?", entityID).Update("status", newStatus).Error
	case "PURCHASE_ORDER", "purchase_order":
		return tx.Model(&models.PurchaseOrder{}).Where("id = ?", entityID).Update("status", newStatus).Error
	case "PAYMENT_VOUCHER", "payment_voucher":
		return tx.Model(&models.PaymentVoucher{}).Where("id = ?", entityID).Update("status", newStatus).Error
	case "GRN", "grn":
		return tx.Model(&models.GoodsReceivedNote{}).Where("id = ?", entityID).Update("status", newStatus).Error
	default:
		return fmt.Errorf("unsupported entity type: %s", entityType)
	}
}

// addActionHistoryEntry adds an action history entry to the document
func (s *WorkflowExecutionService) addActionHistoryEntry(tx *gorm.DB, entityType, entityID, userID, action, comments string) error {
	actionEntry := types.ActionHistoryEntry{
		ID:               uuid.New().String(),
		ActionType:       action,
		PerformedBy:      userID,
		PerformedByName:  "", // Will be filled by caller if needed
		PerformedByRole:  "", // Will be filled by caller if needed
		PerformedAt:      time.Now(),
		Comments:         comments,
		PreviousStatus:   "", // Could be enhanced to track status transitions
		NewStatus:        "approved",
	}

	switch entityType {
	case "REQUISITION", "requisition":
		var requisition models.Requisition
		if err := tx.Where("id = ?", entityID).First(&requisition).Error; err != nil {
			return err
		}
		
		// Get existing history
		var history []types.ActionHistoryEntry
		history = requisition.ActionHistory.Data()
		
		// Add new entry
		history = append(history, actionEntry)
		
		// Update with new history
		requisition.ActionHistory = datatypes.NewJSONType(history)
		
		return tx.Save(&requisition).Error
		
	case "BUDGET", "budget":
		var budget models.Budget
		if err := tx.Where("id = ?", entityID).First(&budget).Error; err != nil {
			return err
		}
		
		var history []types.ActionHistoryEntry
		history = budget.ActionHistory.Data()
		history = append(history, actionEntry)
		
		budget.ActionHistory = datatypes.NewJSONType(history)
		
		return tx.Save(&budget).Error
		
	case "PURCHASE_ORDER", "purchase_order":
		var po models.PurchaseOrder
		if err := tx.Where("id = ?", entityID).First(&po).Error; err != nil {
			return err
		}
		
		var history []types.ActionHistoryEntry
		history = po.ActionHistory.Data()
		history = append(history, actionEntry)
		
		po.ActionHistory = datatypes.NewJSONType(history)
		
		return tx.Save(&po).Error
		
	case "PAYMENT_VOUCHER", "payment_voucher":
		var pv models.PaymentVoucher
		if err := tx.Where("id = ?", entityID).First(&pv).Error; err != nil {
			return err
		}
		
		var history []types.ActionHistoryEntry
		history = pv.ActionHistory.Data()
		history = append(history, actionEntry)
		
		pv.ActionHistory = datatypes.NewJSONType(history)
		
		return tx.Save(&pv).Error
		
	case "GRN", "grn":
		var grn models.GoodsReceivedNote
		if err := tx.Where("id = ?", entityID).First(&grn).Error; err != nil {
			return err
		}
		
		var history []types.ActionHistoryEntry
		history = grn.ActionHistory.Data()
		history = append(history, actionEntry)
		
		grn.ActionHistory = datatypes.NewJSONType(history)
		
		return tx.Save(&grn).Error
	}
	
	return nil
}

// triggerPostApprovalAutomation triggers automation after document approval
func (s *WorkflowExecutionService) triggerPostApprovalAutomation(ctx context.Context, entityType, entityID string) error {
	if s.automationService == nil {
		return nil // No automation service configured
	}

	config := s.automationService.GetDefaultAutomationConfig()
	
	switch entityType {
	case "REQUISITION", "requisition":
		if !config.AutoCreatePOFromRequisition {
			return nil // Automation disabled
		}
		
		// Get the approved requisition
		var requisition models.Requisition
		if err := s.db.Where("id = ?", entityID).First(&requisition).Error; err != nil {
			return fmt.Errorf("failed to get requisition: %w", err)
		}
		
		// Validate automation prerequisites
		if err := s.automationService.ValidateAutomationPrerequisites("requisition", &requisition); err != nil {
			return fmt.Errorf("automation prerequisites not met: %w", err)
		}
		
		// Create purchase order
		result, err := s.automationService.CreatePurchaseOrderFromRequisition(ctx, &requisition, config)
		if err != nil {
			return fmt.Errorf("failed to create purchase order: %w", err)
		}
		
		if !result.Success {
			return fmt.Errorf("purchase order creation failed: %s", result.Error)
		}
		
		// Update requisition with auto-created PO info
		autoCreatedPO := map[string]interface{}{
			"id":       result.DocumentID,
			"created":  true,
		}
		
		if result.CreatedDocument != nil {
			if po, ok := result.CreatedDocument.(*models.PurchaseOrder); ok {
				autoCreatedPO["documentNumber"] = po.DocumentNumber
				autoCreatedPO["amount"] = po.TotalAmount
			}
		}
		
		autoCreatedJSON, _ := datatypes.NewJSONType(autoCreatedPO).MarshalJSON()
		s.db.Model(&requisition).Updates(map[string]interface{}{
			"automation_used": true,
			"auto_created_po": datatypes.JSON(autoCreatedJSON),
		})
		
	case "PURCHASE_ORDER", "purchase_order":
		if !config.AutoCreateGRNFromPO {
			return nil // Automation disabled
		}
		
		// Get the approved purchase order
		var po models.PurchaseOrder
		if err := s.db.Where("id = ?", entityID).First(&po).Error; err != nil {
			return fmt.Errorf("failed to get purchase order: %w", err)
		}
		
		// Validate automation prerequisites
		if err := s.automationService.ValidateAutomationPrerequisites("purchase_order", &po); err != nil {
			return fmt.Errorf("automation prerequisites not met: %w", err)
		}
		
		// Create GRN
		result, err := s.automationService.CreateGRNFromPurchaseOrder(ctx, &po, config)
		if err != nil {
			return fmt.Errorf("failed to create GRN: %w", err)
		}
		
		if !result.Success {
			return fmt.Errorf("GRN creation failed: %s", result.Error)
		}
		
		// Update PO with auto-created GRN info
		autoCreatedGRN := map[string]interface{}{
			"id":      result.DocumentID,
			"created": true,
		}
		
		if result.CreatedDocument != nil {
			if grn, ok := result.CreatedDocument.(*models.GoodsReceivedNote); ok {
				autoCreatedGRN["documentNumber"] = grn.DocumentNumber
			}
		}
		
		autoCreatedJSON, _ := datatypes.NewJSONType(autoCreatedGRN).MarshalJSON()
		s.db.Model(&po).Updates(map[string]interface{}{
			"automation_used":   true,
			"auto_created_grn": datatypes.JSON(autoCreatedJSON),
		})
		
	case "GRN", "grn":
		if !config.AutoCreatePVFromGRN {
			return nil // Automation disabled
		}
		
		// Get the approved GRN
		var grn models.GoodsReceivedNote
		if err := s.db.Where("id = ?", entityID).First(&grn).Error; err != nil {
			return fmt.Errorf("failed to get GRN: %w", err)
		}
		
		// Validate automation prerequisites
		if err := s.automationService.ValidateAutomationPrerequisites("grn", &grn); err != nil {
			return fmt.Errorf("automation prerequisites not met: %w", err)
		}
		
		// Create Payment Voucher
		result, err := s.automationService.CreatePaymentVoucherFromGRN(ctx, &grn, config)
		if err != nil {
			return fmt.Errorf("failed to create payment voucher: %w", err)
		}
		
		if !result.Success {
			return fmt.Errorf("payment voucher creation failed: %s", result.Error)
		}
		
		// Update GRN with auto-created PV info
		autoCreatedPV := map[string]interface{}{
			"id":      result.DocumentID,
			"created": true,
		}
		
		if result.CreatedDocument != nil {
			if pv, ok := result.CreatedDocument.(*models.PaymentVoucher); ok {
				autoCreatedPV["documentNumber"] = pv.DocumentNumber
				autoCreatedPV["amount"] = pv.Amount
			}
		}
		
		autoCreatedJSON, _ := datatypes.NewJSONType(autoCreatedPV).MarshalJSON()
		s.db.Model(&grn).Updates(map[string]interface{}{
			"automation_used": true,
			"auto_created_pv": datatypes.JSON(autoCreatedJSON),
		})
	}
	
	return nil
}

// UpdateDocumentStatus updates the status of the actual document when workflow completes (public for testing)
func (s *WorkflowExecutionService) UpdateDocumentStatus(tx *gorm.DB, entityType, entityID, newStatus string) error {
	return s.updateDocumentStatus(tx, entityType, entityID, newStatus)
}

// AddActionHistoryEntry adds an action history entry to the document (public for testing)
func (s *WorkflowExecutionService) AddActionHistoryEntry(tx *gorm.DB, entityType, entityID, userID, action, comments string) error {
	return s.addActionHistoryEntry(tx, entityType, entityID, userID, action, comments)
}

// getDocumentPriority fetches the priority from the document based on entity type
func (s *WorkflowExecutionService) getDocumentPriority(tx *gorm.DB, entityID, entityType string) string {
	defaultPriority := "medium"

	switch strings.ToLower(entityType) {
	case "requisition":
		var req models.Requisition
		if err := tx.Where("id = ?", entityID).First(&req).Error; err == nil {
			if req.Priority != "" {
				return strings.ToLower(req.Priority)
			}
		}
	case "purchase_order":
		var po models.PurchaseOrder
		if err := tx.Where("id = ?", entityID).First(&po).Error; err == nil {
			if po.Priority != "" {
				return strings.ToLower(po.Priority)
			}
		}
	case "payment_voucher":
		var pv models.PaymentVoucher
		if err := tx.Where("id = ?", entityID).First(&pv).Error; err == nil {
			if pv.Priority != "" {
				return strings.ToLower(pv.Priority)
			}
		}
	// budget and goods_received_note don't have priority field - use default
	}

	return defaultPriority
}

// getDocumentDueDate fetches the due date from the document based on entity type
func (s *WorkflowExecutionService) getDocumentDueDate(tx *gorm.DB, entityID, entityType string) *time.Time {
	switch strings.ToLower(entityType) {
	case "requisition":
		var req models.Requisition
		if err := tx.Where("id = ?", entityID).First(&req).Error; err == nil {
			if !req.RequiredByDate.IsZero() {
				return &req.RequiredByDate
			}
		}
	case "purchase_order":
		var po models.PurchaseOrder
		if err := tx.Where("id = ?", entityID).First(&po).Error; err == nil {
			// Try RequiredByDate first (pointer), then DeliveryDate (value)
			if po.RequiredByDate != nil {
				return po.RequiredByDate
			}
			if !po.DeliveryDate.IsZero() {
				return &po.DeliveryDate
			}
		}
	case "payment_voucher":
		var pv models.PaymentVoucher
		if err := tx.Where("id = ?", entityID).First(&pv).Error; err == nil {
			if pv.PaymentDueDate != nil {
				return pv.PaymentDueDate
			}
		}
	}

	return nil
}