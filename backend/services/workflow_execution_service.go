package services

import (
	"context"
	"fmt"
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
		Priority:             "medium",
		CreatedAt:            time.Now(),
	}

	// Set due date if timeout is specified
	if firstStage.TimeoutHours != nil && *firstStage.TimeoutHours > 0 {
		dueDate := time.Now().Add(time.Duration(*firstStage.TimeoutHours) * time.Hour)
		task.DueDate = &dueDate
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
		go func() {
			if err := s.notificationService.HandleWorkflowEvent(notificationEvent); err != nil {
				fmt.Printf("Failed to send approval required notification: %v\n", err)
			}
		}()
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

// ApproveWorkflowTask approves a workflow task and progresses the workflow
func (s *WorkflowExecutionService) ApproveWorkflowTask(ctx context.Context, taskID, userID, signature, comments string) error {
	// Start transaction
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

	if task.Status != "pending" {
		tx.Rollback()
		return fmt.Errorf("task is not in pending status")
	}

	// Validate user has the required role for this task
	if task.AssignedRole != nil {
		var user models.User
		if err := tx.Where("id = ?", userID).First(&user).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("user not found: %w", err)
		}

		// Check if user's role matches the required role for this task
		if user.Role != *task.AssignedRole {
			tx.Rollback()
			return fmt.Errorf("insufficient permissions: user role '%s' does not match required role '%s'", user.Role, *task.AssignedRole)
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

	// Update task status
	now := time.Now()
	task.Status = "completed"
	task.CompletedAt = &now
	task.ClaimedBy = &userID
	task.ClaimedAt = &now

	if err := tx.Save(&task).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update task: %w", err)
	}

	// Add stage execution to history
	stageExecution := models.StageExecution{
		StageNumber:  task.StageNumber,
		StageName:    task.StageName,
		ApproverID:   userID,
		ApproverName: "", // Will be filled by caller
		ApproverRole: "", // Will be filled by caller
		Action:       "approved",
		Comments:     comments,
		Signature:    signature,
		ExecutedAt:   now,
	}

	// Update assignment stage history
	history, err := assignment.GetStageHistory()
	if err != nil {
		history = []models.StageExecution{}
	}
	history = append(history, stageExecution)

	if err := assignment.AddStageExecution(stageExecution); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update stage history: %w", err)
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
		nextStage := stages[nextStageNumber-1] // stages are 1-indexed

		assignment.CurrentStage = nextStageNumber

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
			Priority:             "medium",
			CreatedAt:            time.Now(),
		}

		// Set due date if timeout is specified
		if nextStage.TimeoutHours != nil && *nextStage.TimeoutHours > 0 {
			dueDate := time.Now().Add(time.Duration(*nextStage.TimeoutHours) * time.Hour)
			nextTask.DueDate = &dueDate
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
			
			// Send notification asynchronously
			go func() {
				if err := s.notificationService.HandleWorkflowEvent(notificationEvent); err != nil {
					fmt.Printf("Failed to send approval notification: %v\n", err)
				}
			}()
		}

		if err := s.triggerPostApprovalAutomation(ctx, assignment.EntityType, assignment.EntityID); err != nil {
			// Log error but don't fail the approval since the workflow is already completed
			fmt.Printf("Post-approval automation failed for %s %s: %v\n", assignment.EntityType, assignment.EntityID, err)
		}
	} else {
		// Send notification for next stage approval required
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
			
			// Send notification asynchronously
			go func() {
				if err := s.notificationService.HandleWorkflowEvent(notificationEvent); err != nil {
					fmt.Printf("Failed to send next stage approval notification: %v\n", err)
				}
			}()
		}
	}

	return nil
}

// RejectWorkflowTask rejects a workflow task and marks the workflow as rejected
func (s *WorkflowExecutionService) RejectWorkflowTask(ctx context.Context, taskID, userID, signature, reason string) error {
	// Start transaction
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

	if task.Status != "pending" {
		tx.Rollback()
		return fmt.Errorf("task is not in pending status")
	}

	// Validate user has the required role for this task
	if task.AssignedRole != nil {
		var user models.User
		if err := tx.Where("id = ?", userID).First(&user).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("user not found: %w", err)
		}

		// Check if user's role matches the required role for this task
		if user.Role != *task.AssignedRole {
			tx.Rollback()
			return fmt.Errorf("insufficient permissions: user role '%s' does not match required role '%s'", user.Role, *task.AssignedRole)
		}
	}

	// Get the workflow assignment
	var assignment models.WorkflowAssignment
	if err := tx.Where("id = ?", task.WorkflowAssignmentID).First(&assignment).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("workflow assignment not found: %w", err)
	}

	// Update task status
	now := time.Now()
	task.Status = "completed"
	task.CompletedAt = &now
	task.ClaimedBy = &userID
	task.ClaimedAt = &now

	if err := tx.Save(&task).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update task: %w", err)
	}

	// Add stage execution to history
	stageExecution := models.StageExecution{
		StageNumber:  task.StageNumber,
		StageName:    task.StageName,
		ApproverID:   userID,
		ApproverName: "", // Will be filled by caller
		ApproverRole: "", // Will be filled by caller
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
		
		// Send notification asynchronously
		go func() {
			if err := s.notificationService.HandleWorkflowEvent(notificationEvent); err != nil {
				fmt.Printf("Failed to send rejection notification: %v\n", err)
			}
		}()
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
		stageInfo := StageProgressInfo{
			StageNumber:   stage.StageNumber,
			StageName:     stage.StageName,
			RequiredRole:  stage.RequiredRole,
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
			response.NextApprover = fmt.Sprintf("Required Role: %s", *currentTask.AssignedRole)
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
				autoCreatedPO["poNumber"] = po.PONumber
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
				autoCreatedGRN["grnNumber"] = grn.GRNNumber
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
				autoCreatedPV["pvNumber"] = pv.PVNumber
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