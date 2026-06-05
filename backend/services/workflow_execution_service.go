package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/types"
	"github.com/liyali/liyali-gateway/utils"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// WorkflowExecutionService handles workflow assignment and execution
type WorkflowExecutionService struct {
	db                  *gorm.DB
	workflowService     *WorkflowService
	auditService        *AuditService
	automationService   *DocumentAutomationService
	notificationService *NotificationService
}

// NewWorkflowExecutionService creates a new workflow execution service
func NewWorkflowExecutionService(db *gorm.DB, workflowService *WorkflowService, auditService *AuditService, automationService *DocumentAutomationService) *WorkflowExecutionService {
	return &WorkflowExecutionService{
		db:                  db,
		workflowService:     workflowService,
		auditService:        auditService,
		automationService:   automationService,
		notificationService: NewNotificationService(db),
	}
}

// StartClaimExpiryWorker runs a background goroutine that periodically resets
// expired claimed tasks back to pending status so other users can claim them.
func (s *WorkflowExecutionService) StartClaimExpiryWorker(ctx context.Context) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	log.Println("[ClaimExpiry] Background claim expiry worker started (interval: 60s)")
	for {
		select {
		case <-ctx.Done():
			log.Println("[ClaimExpiry] Background claim expiry worker stopped")
			return
		case <-ticker.C:
			result := s.db.Table("workflow_tasks").
				Where("UPPER(status) = ? AND claim_expiry < ?", "CLAIMED", time.Now()).
				Updates(map[string]interface{}{
					"claimed_by":   nil,
					"claimed_at":   nil,
					"claim_expiry": nil,
					"status":       "PENDING",
				})
			if result.Error != nil {
				log.Printf("[ClaimExpiry] Error expiring stale claims: %v", result.Error)
			} else if result.RowsAffected > 0 {
				log.Printf("[ClaimExpiry] Auto-expired %d stale claim(s)", result.RowsAffected)
			}
		}
	}
}

// AssignWorkflowToDocument assigns a workflow to a document and creates initial tasks
func (s *WorkflowExecutionService) AssignWorkflowToDocument(ctx context.Context, organizationID, entityID, entityType, userID string) (*models.WorkflowAssignment, error) {
	// Get the default workflow for this entity type
	workflow, err := s.workflowService.GetDefaultWorkflow(ctx, organizationID, entityType)
	if err != nil {
		return nil, fmt.Errorf("failed to get default workflow: %w", err)
	}

	return s.assignWorkflow(ctx, organizationID, entityID, entityType, userID, workflow)
}

// AssignWorkflowToDocumentWithID assigns a user-selected workflow to a document.
func (s *WorkflowExecutionService) AssignWorkflowToDocumentWithID(
	ctx context.Context,
	organizationID, entityID, entityType, workflowID, userID string,
) (*models.WorkflowAssignment, error) {
	workflowUUID, err := uuid.Parse(workflowID)
	if err != nil {
		return nil, fmt.Errorf("invalid workflow ID format")
	}

	workflow, err := s.workflowService.GetWorkflow(ctx, workflowUUID, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get selected workflow: %w", err)
	}

	if !workflow.IsActive {
		return nil, fmt.Errorf("selected workflow is inactive")
	}

	if !strings.EqualFold(workflow.EntityType, entityType) {
		return nil, fmt.Errorf("workflow entity type mismatch")
	}

	return s.assignWorkflow(ctx, organizationID, entityID, entityType, userID, workflow)
}

// AssignWorkflowToDocumentWithIDTx is the transactional variant of
// AssignWorkflowToDocumentWithID. It runs all writes inside the caller-provided
// transaction so that failure in subsequent handler steps (e.g. persisting the
// document's new PENDING status) rolls the assignment back together with the
// document change — eliminating orphaned "PENDING doc without workflow" or
// "workflow with no matching doc status" inconsistencies.
//
// The caller owns Commit/Rollback and must run any async side effects
// (notifications, etc.) after commit.
func (s *WorkflowExecutionService) AssignWorkflowToDocumentWithIDTx(
	ctx context.Context,
	tx *gorm.DB,
	organizationID, entityID, entityType, workflowID, userID string,
) (*models.WorkflowAssignment, error) {
	workflowUUID, err := uuid.Parse(workflowID)
	if err != nil {
		return nil, fmt.Errorf("invalid workflow ID format")
	}

	workflow, err := s.workflowService.GetWorkflow(ctx, workflowUUID, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get selected workflow: %w", err)
	}

	if !workflow.IsActive {
		return nil, fmt.Errorf("selected workflow is inactive")
	}

	if !strings.EqualFold(workflow.EntityType, entityType) {
		return nil, fmt.Errorf("workflow entity type mismatch")
	}

	stages, err := workflow.GetStages()
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow stages: %w", err)
	}
	if len(stages) == 0 {
		return nil, fmt.Errorf("workflow has no stages")
	}

	now := time.Now()
	assignment := &models.WorkflowAssignment{
		ID:              uuid.New().String(),
		OrganizationID:  organizationID,
		EntityID:        entityID,
		EntityType:      entityType,
		WorkflowID:      workflow.ID,
		WorkflowVersion: workflow.Version,
		CurrentStage:    1,
		Status:          "IN_PROGRESS",
		StageHistory:    datatypes.JSON{},
		AssignedAt:      now,
		AssignedBy:      userID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := tx.Create(assignment).Error; err != nil {
		return nil, fmt.Errorf("failed to create workflow assignment: %w", err)
	}

	firstStage := stages[0]
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
		Status:               "PENDING",
		Priority:             documentPriority,
		CreatedAt:            now,
	}

	var calculatedDueDate time.Time
	if firstStage.TimeoutHours != nil && *firstStage.TimeoutHours > 0 {
		calculatedDueDate = now.Add(time.Duration(*firstStage.TimeoutHours) * time.Hour)
	} else {
		calculatedDueDate = now.Add(7 * 24 * time.Hour)
	}
	documentDueDate := s.getDocumentDueDate(tx, entityID, entityType)
	if documentDueDate != nil && documentDueDate.Before(calculatedDueDate) {
		task.DueDate = documentDueDate
	} else {
		task.DueDate = &calculatedDueDate
	}

	if err := tx.Create(task).Error; err != nil {
		return nil, fmt.Errorf("failed to create workflow task: %w", err)
	}

	return assignment, nil
}

// SubmitRoutingResult contains the result of a routing-aware requisition submission.
type SubmitRoutingResult struct {
	RoutingPath     string                      `json:"routingPath"`              // "accounting" or "procurement"
	AutoApproved    bool                        `json:"autoApproved"`
	Assignment      *models.WorkflowAssignment  `json:"assignment,omitempty"`
	AutoCreatedPO   *AutomationResult           `json:"autoCreatedPO,omitempty"`
	AutoCreatedPOID string                      `json:"autoCreatedPoId,omitempty"`
	AutoCreatedPVID string                      `json:"autoCreatedPvId,omitempty"`
	RoutingType     string                      `json:"routingType"`
}

// SubmitRequisitionWithRouting handles requisition submission with conditional routing.
// It checks the selected workflow's conditions and either:
//   - Auto-approves + auto-generates PO (accounting/direct_payment path with 0 stages + criteria met), or
//   - Assigns the normal workflow stages (procurement or accounting with stages)
//
// For direct_payment routing a draft PaymentVoucher is also created after the PO.
func (s *WorkflowExecutionService) SubmitRequisitionWithRouting(
	ctx context.Context,
	organizationID, entityID, workflowID, userID string,
	requisition *models.Requisition,
) (*SubmitRoutingResult, error) {
	// 1. Fetch and validate the workflow
	workflowUUID, err := uuid.Parse(workflowID)
	if err != nil {
		return nil, fmt.Errorf("invalid workflow ID format")
	}

	workflow, err := s.workflowService.GetWorkflow(ctx, workflowUUID, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow: %w", err)
	}

	if !workflow.IsActive {
		return nil, fmt.Errorf("selected workflow is inactive")
	}

	if !strings.EqualFold(workflow.EntityType, "requisition") {
		return nil, fmt.Errorf("workflow entity type mismatch: expected requisition")
	}

	// 2. Parse workflow conditions
	conditions, _ := workflow.GetConditions()

	// 3. Derive routing type from conditions (default: procurement).
	routingType := models.RoutingTypeProcurement
	if conditions != nil && conditions.RoutingType != "" {
		routingType = strings.ToLower(conditions.RoutingType)
	}

	// 4. direct_payment-specific validation: payee must be present, and workflow must have 0 stages.
	if routingType == models.RoutingTypeDirectPayment {
		hasPayee := (requisition.PayeeID != nil && *requisition.PayeeID != "") ||
			len(requisition.PayeeSnapshot) > 0
		if !hasPayee {
			return nil, fmt.Errorf("direct_payment requisition requires a payee: set payeeId or payeeSnapshot")
		}
		stages, _ := workflow.GetStages()
		if len(stages) > 0 {
			return nil, fmt.Errorf("direct_payment workflow %s must have 0 approval stages, has %d", workflow.ID, len(stages))
		}
	}

	// 5. Denormalize routing_type onto the requisition row.
	if err := s.db.WithContext(ctx).Model(&models.Requisition{}).
		Where("id = ?", entityID).
		Update("routing_type", routingType).Error; err != nil {
		log.Printf("Warning: failed to denormalize routing_type on requisition %s: %v", entityID, err)
	}
	requisition.RoutingType = routingType

	// 6. Determine whether this is an auto-approve path (accounting OR direct_payment).
	isAutoPath := routingType == models.RoutingTypeAccounting ||
		routingType == models.RoutingTypeDirectPayment

	if !isAutoPath {
		// PROCUREMENT PATH: Standard workflow assignment
		assignment, err := s.assignWorkflow(ctx, organizationID, entityID, "requisition", userID, workflow)
		if err != nil {
			return nil, err
		}
		return &SubmitRoutingResult{
			RoutingPath:  routingType,
			RoutingType:  routingType,
			AutoApproved: false,
			Assignment:   assignment,
		}, nil
	}

	// AUTO PATH (accounting or direct_payment): check auto-approval eligibility.
	stages, _ := workflow.GetStages()
	categoryID := ""
	if requisition.CategoryID != nil {
		categoryID = *requisition.CategoryID
	}

	shouldAutoApprove := conditions.MeetsAutoApprovalCriteria(requisition.TotalAmount, categoryID) && len(stages) == 0

	if !shouldAutoApprove {
		// Has stages — run through the normal stage workflow.
		assignment, err := s.assignWorkflow(ctx, organizationID, entityID, "requisition", userID, workflow)
		if err != nil {
			return nil, err
		}
		return &SubmitRoutingResult{
			RoutingPath:  routingType,
			RoutingType:  routingType,
			AutoApproved: false,
			Assignment:   assignment,
		}, nil
	}

	// AUTO-APPROVE PATH: 0 stages + criteria met
	result, err := s.autoApproveAndGeneratePO(ctx, organizationID, entityID, userID, requisition, workflow, conditions)
	if err != nil {
		return nil, err
	}
	// Stamp routing type on result (autoApproveAndGeneratePO doesn't know routingType).
	result.RoutingType = routingType

	// 7. For direct_payment: chain auto-creation of a draft PV.
	if routingType == models.RoutingTypeDirectPayment && result.AutoCreatedPOID != "" {
		pvID, pvErr := s.autoCreateDraftPV(ctx, requisition, result.AutoCreatedPOID)
		if pvErr != nil {
			log.Printf("Warning: auto-create draft PV failed for req=%s po=%s: %v (PO preserved)", entityID, result.AutoCreatedPOID, pvErr)
		} else {
			result.AutoCreatedPVID = pvID
		}
	}

	return result, nil
}

// autoCreateDraftPV creates a PaymentVoucher in DRAFT status linked to the auto-generated PO.
// It also propagates routing_type=direct_payment to the PO row.
// Returns the new PV's ID. Failure here does NOT roll back the PO (audit trail preserved).
func (s *WorkflowExecutionService) autoCreateDraftPV(
	ctx context.Context,
	req *models.Requisition,
	poID string,
) (string, error) {
	// Load PO so we can use its DocumentNumber as LinkedPO.
	var po models.PurchaseOrder
	if err := s.db.WithContext(ctx).First(&po, "id = ?", poID).Error; err != nil {
		return "", fmt.Errorf("load po: %w", err)
	}

	// Propagate routing_type and procurement_flow to the PO row.
	if err := s.db.WithContext(ctx).Model(&po).Updates(map[string]any{
		"routing_type":     models.RoutingTypeDirectPayment,
		"procurement_flow": "payment_first",
	}).Error; err != nil {
		return "", fmt.Errorf("set po routing/procurement flow: %w", err)
	}

	// Derive payee display name from PayeeSnapshot.
	name := "Direct Payment"
	if len(req.PayeeSnapshot) > 0 {
		var snap map[string]interface{}
		if err := json.Unmarshal(req.PayeeSnapshot, &snap); err == nil {
			if v, ok := snap["name"].(string); ok && v != "" {
				name = v
			}
		}
	}

	// Build metadata via json.Marshal to avoid injection from raw payeeSnapshot.
	metaPayload := map[string]interface{}{
		"autoCreated": true,
		"sourceReqID": req.ID,
	}
	if len(req.PayeeSnapshot) > 0 {
		metaPayload["payeeSnapshot"] = json.RawMessage(req.PayeeSnapshot)
	} else {
		metaPayload["payeeSnapshot"] = nil
	}
	metaBytes, err := json.Marshal(metaPayload)
	if err != nil {
		return "", fmt.Errorf("marshal pv metadata: %w", err)
	}

	pvDocNum := utils.GenerateDocumentNumber("PV")
	pv := models.PaymentVoucher{
		ID:             uuid.New().String(),
		DocumentNumber: pvDocNum,
		OrganizationID: req.OrganizationID,
		Status:         models.StatusDraft,
		CreatedBy:      req.RequesterId,
		LinkedPO:       po.DocumentNumber,
		RoutingType:    models.RoutingTypeDirectPayment,
		VendorName:     name,
		Amount:         req.TotalAmount,
		Currency:       req.Currency,
		Metadata:       datatypes.JSON(metaBytes),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	pv.ApprovalHistory = datatypes.NewJSONType([]types.ApprovalRecord{})

	if err := s.db.WithContext(ctx).Create(&pv).Error; err != nil {
		return "", fmt.Errorf("create draft pv: %w", err)
	}
	return pv.ID, nil
}

// autoApproveAndGeneratePO handles instant auto-approval of a requisition and optional PO generation.
func (s *WorkflowExecutionService) autoApproveAndGeneratePO(
	ctx context.Context,
	organizationID, entityID, userID string,
	requisition *models.Requisition,
	workflow *models.Workflow,
	conditions *models.WorkflowConditions,
) (*SubmitRoutingResult, error) {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	now := time.Now()

	// 1. Create a workflow assignment record for audit trail (marked as auto-completed)
	assignment := &models.WorkflowAssignment{
		ID:              uuid.New().String(),
		OrganizationID:  organizationID,
		EntityID:        entityID,
		EntityType:      "requisition",
		WorkflowID:      workflow.ID,
		WorkflowVersion: workflow.Version,
		CurrentStage:    0,
		Status: "COMPLETED",
		StageHistory:    datatypes.JSON("[]"),
		AssignedAt:      now,
		AssignedBy:      userID,
		CompletedAt:     &now,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	autoExecution := models.StageExecution{
		StageNumber:  0,
		StageName:    "Auto-Approval",
		ApproverID:   "system",
		ApproverName: "System Auto-Approval",
		ApproverRole: "system",
		Action:       "auto_approved",
		Comments:     "Automatically approved based on workflow conditions",
		ExecutedAt:   now,
	}
	if err := assignment.AddStageExecution(autoExecution); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to add auto-approval to history: %w", err)
	}

	if err := tx.Create(assignment).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create auto-approval assignment: %w", err)
	}

	// 2. Update requisition status to "approved"
	if err := s.updateDocumentStatusScoped(tx, "requisition", entityID, organizationID, models.StatusApproved); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update requisition status: %w", err)
	}

	// 3. Add action history entry
	if err := s.addActionHistoryEntry(tx, "requisition", entityID, "system", "AUTO_APPROVED",
		"Requisition auto-approved via accounting workflow"); err != nil {
		fmt.Printf("Warning: failed to add action history: %v\n", err)
	}

	// Commit the requisition approval
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit auto-approval: %w", err)
	}

	result := &SubmitRoutingResult{
		RoutingPath:  "accounting",
		AutoApproved: true,
		Assignment:   assignment,
	}

	// 4. Auto-generate PO if configured
	if conditions.AutoGeneratePO && s.automationService != nil {
		// Refresh requisition status for the automation service check
		requisition.Status = "APPROVED"

		targetStatus := "DRAFT"
		if conditions.AutoApprovePO {
			targetStatus = "APPROVED"
		}

		poResult, err := s.automationService.CreatePurchaseOrderFromRequisitionWithStatus(
			ctx, requisition, targetStatus,
		)
		if err != nil {
			fmt.Printf("Warning: auto-PO generation failed: %v\n", err)
		} else if poResult != nil && poResult.Success {
			result.AutoCreatedPO = poResult
			result.AutoCreatedPOID = poResult.DocumentID

			// Update requisition with auto-created PO info
			autoCreatedPO := map[string]interface{}{
				"id":      poResult.DocumentID,
				"created": true,
			}
			if po, ok := poResult.CreatedDocument.(models.PurchaseOrder); ok {
				autoCreatedPO["documentNumber"] = po.DocumentNumber
				autoCreatedPO["amount"] = po.TotalAmount
			}
			autoCreatedJSON, _ := datatypes.NewJSONType(autoCreatedPO).MarshalJSON()
			s.db.Model(&models.Requisition{}).Where("id = ?", entityID).Updates(map[string]interface{}{
				"automation_used": true,
				"auto_created_po": datatypes.JSON(autoCreatedJSON),
			})
		}
	}

	// 5. Send notification
	if s.notificationService != nil {
		event := NotificationEvent{
			Type:           "document_auto_approved",
			DocumentID:     entityID,
			DocumentType:   "requisition",
			OrganizationID: organizationID,
			Action:         "auto_approved",
			ActorID:        "system",
			Details:        "Requisition auto-approved via accounting workflow",
			Timestamp:      now,
		}
		go func(e NotificationEvent) {
			notifyCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			select {
			case <-notifyCtx.Done():
				return
			default:
				s.notificationService.HandleWorkflowEvent(e)
			}
		}(event)
	}

	return result, nil
}

func (s *WorkflowExecutionService) assignWorkflow(
	ctx context.Context,
	organizationID, entityID, entityType, userID string,
	workflow *models.Workflow,
) (*models.WorkflowAssignment, error) {
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
		Status:          "IN_PROGRESS",
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
		Status: "PENDING",
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
			Type:           "approval_required",
			DocumentID:     entityID,
			DocumentType:   entityType,
			OrganizationID: organizationID,
			Action:         "workflow_assigned",
			ActorID:        userID,
			Details:        fmt.Sprintf("Workflow assigned for %s approval", entityType),
			Timestamp:      time.Now(),
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
	err := s.db.Where("organization_id = ? AND entity_id = ? AND UPPER(status) = ?", organizationID, entityID, "PENDING").
		Order("stage_number ASC").
		Find(&tasks).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get pending workflow tasks: %w", err)
	}

	return tasks, nil
}

// ApproveWorkflowTask approves a workflow task and progresses the workflow with optimistic locking
// canUserActOnTask checks whether the given user is authorised to act on a workflow task.
// It handles three cases for AssignedRole:
//   - UUID for a system role  → compare role name against user.Role
//   - UUID for a custom role  → check user_organization_roles membership
//   - Plain name string       → compare directly against user.Role
//
// Fallbacks (in order): built-in approver role list, org role with approval permission.
func (s *WorkflowExecutionService) canUserActOnTask(tx *gorm.DB, task *models.WorkflowTask, user *models.User) error {
	approverRoles := []string{"admin", "approver", "finance", "manager", "supervisor", "department_head"}
	approvalPermissions := []string{
		"requisition.approve", "approval.approve", "budget.approve",
		"purchase_order.approve", "payment_voucher.approve", "grn.approve",
	}

	// PRIORITY 1: specific user assignment
	if task.AssignedUserID != nil && *task.AssignedUserID != "" {
		if *task.AssignedUserID != user.ID {
			return fmt.Errorf("insufficient permissions: this task has been assigned to a specific user and only they can act on it")
		}
		return nil
	}

	// PRIORITY 2: role-based assignment
	if task.AssignedRole == nil || *task.AssignedRole == "" {
		// No restriction — any built-in approver may act
		for _, r := range approverRoles {
			if strings.EqualFold(user.Role, r) {
				return nil
			}
		}
		return fmt.Errorf("insufficient permissions: no approver role")
	}

	assignedRole := *task.AssignedRole
	hasPermission := false

	if _, parseErr := uuid.Parse(assignedRole); parseErr == nil {
		// It's a UUID — resolve the org role record
		var orgRole models.OrganizationRole
		if tx.Where("id = ?", assignedRole).First(&orgRole).Error == nil {
			if orgRole.IsSystemRole {
				// System role UUID: compare role name against user.Role (names are stable, UUIDs are not)
				hasPermission = strings.EqualFold(user.Role, orgRole.Name)
			} else {
				// Custom org role UUID: check user_organization_roles membership
				var uor models.UserOrganizationRole
				hasPermission = tx.Where(
					"user_id = ? AND organization_id = ? AND role_id = ? AND active = ?",
					user.ID, task.OrganizationID, assignedRole, true,
				).First(&uor).Error == nil
			}
		}
		// Fallback: built-in approver role
		if !hasPermission {
			for _, r := range approverRoles {
				if strings.EqualFold(user.Role, r) {
					hasPermission = true
					break
				}
			}
		}
	} else {
		// Plain role name
		hasPermission = strings.EqualFold(user.Role, assignedRole)
		if !hasPermission {
			for _, r := range approverRoles {
				if strings.EqualFold(user.Role, r) {
					hasPermission = true
					break
				}
			}
		}
	}

	// Final fallback: custom org role with any approval permission
	if !hasPermission {
		var userOrgRoles []models.UserOrganizationRole
		if tx.Where("user_id = ? AND organization_id = ? AND active = ?",
			user.ID, task.OrganizationID, true).Find(&userOrgRoles).Error == nil {
			for _, uor := range userOrgRoles {
				var orgRole models.OrganizationRole
				if tx.Where("id = ? AND active = ?", uor.RoleID, true).First(&orgRole).Error != nil {
					continue
				}
				var permissions []string
				if json.Unmarshal(orgRole.Permissions, &permissions) != nil {
					continue
				}
				for _, perm := range permissions {
					for _, ap := range approvalPermissions {
						if strings.EqualFold(perm, ap) {
							hasPermission = true
							break
						}
					}
					if hasPermission {
						break
					}
				}
				if hasPermission {
					break
				}
			}
		}
	}

	if !hasPermission {
		return fmt.Errorf("insufficient permissions: user does not have the required role '%s'", assignedRole)
	}
	return nil
}

func (s *WorkflowExecutionService) ApproveWorkflowTask(ctx context.Context, taskID, userID, signature, comments string) error {
	return s.ApproveWorkflowTaskWithVersion(ctx, taskID, userID, signature, comments, 0)
}

// createPaymentExecutionTask queues the post-approval PAID step as a workflow
// task assigned to the finance role. The task is tied to the same workflow
// assignment so the PV detail page's stage history still displays it as the
// trailing step after approval. A 7-day default due date matches ordinary
// approval tasks; explicit timeout support can be added later if finance
// teams ask for tighter SLAs.
func (s *WorkflowExecutionService) createPaymentExecutionTask(
	tx *gorm.DB,
	assignment *models.WorkflowAssignment,
	now time.Time,
) error {
	// Don't double-create if a payment task already exists for this PV
	// (defensive — this path runs at most once per PV, but approve handlers
	// can be retried on network errors and the task-creation step must be
	// idempotent if the final stage is ever re-approved).
	var existing models.WorkflowTask
	if err := tx.Where("entity_id = ? AND kind = ? AND UPPER(status) IN ('PENDING','CLAIMED')",
		assignment.EntityID, models.TaskKindPaymentExecution).
		First(&existing).Error; err == nil {
		return nil
	}

	financeRole := "finance"
	dueDate := now.Add(7 * 24 * time.Hour)

	task := &models.WorkflowTask{
		ID:                   uuid.New().String(),
		OrganizationID:       assignment.OrganizationID,
		WorkflowAssignmentID: assignment.ID,
		EntityID:             assignment.EntityID,
		EntityType:           assignment.EntityType,
		StageNumber:          0,
		StageName:            "Payment Execution",
		Kind:                 models.TaskKindPaymentExecution,
		AssignmentType:       "role",
		AssignedRole:         &financeRole,
		Status:               "PENDING",
		Priority:             s.getDocumentPriority(tx, assignment.EntityID, assignment.EntityType),
		DueDate:              &dueDate,
		Version:              1,
		CreatedAt:            now,
	}

	if err := tx.Create(task).Error; err != nil {
		return fmt.Errorf("failed to create payment execution task: %w", err)
	}
	return nil
}

// completePaymentExecutionTask handles the PAID transition for a
// payment_execution task. Runs inside the caller's transaction; caller owns
// commit/rollback. Flips PV APPROVED -> PAID, sets PaidAmount/PaidDate to the
// approved PV amount, and appends a signed ActionHistory entry attributing
// the payment to the claiming user.
//
// We intentionally don't accept a different paidAmount: partial payments
// would let a claimant shave the invoice silently, and mismatched amounts
// should be handled by rejecting the payment task (which returns PV to
// APPROVED) and creating a corrected PV instead.
func (s *WorkflowExecutionService) completePaymentExecutionTask(
	tx *gorm.DB,
	task *models.WorkflowTask,
	userID string,
	user *models.User,
	signature, comments string,
	now time.Time,
) error {
	if signature == "" {
		return fmt.Errorf("signature is required to execute payment")
	}
	if !strings.EqualFold(task.EntityType, "payment_voucher") {
		return fmt.Errorf("payment_execution task is only valid for payment_voucher entities (got %s)", task.EntityType)
	}

	var voucher models.PaymentVoucher
	if err := tx.Where("id = ? AND organization_id = ?", task.EntityID, task.OrganizationID).First(&voucher).Error; err != nil {
		return fmt.Errorf("payment voucher not found: %w", err)
	}
	if strings.ToUpper(voucher.Status) != models.StatusApproved {
		return fmt.Errorf("payment voucher is in %s status; must be APPROVED to execute payment", voucher.Status)
	}

	// Mark task completed (version-locked to guard against concurrent completion).
	result := tx.Model(task).
		Where("id = ? AND version = ?", task.ID, task.Version).
		Updates(map[string]interface{}{
			"status":       "COMPLETED",
			"completed_at": now,
			"updated_by":   userID,
			"version":      task.Version + 1,
		})
	if result.Error != nil {
		return fmt.Errorf("failed to complete payment execution task: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("payment task was modified by another user, please refresh and try again")
	}

	paidAmount := voucher.Amount
	voucher.Status = models.StatusPaid
	voucher.PaidAmount = &paidAmount
	voucher.PaidDate = &now
	voucher.UpdatedAt = now

	actionHistory := voucher.ActionHistory.Data()
	actionHistory = append(actionHistory, types.ActionHistoryEntry{
		ID:              uuid.New().String(),
		Action:          "PAYMENT_EXECUTED",
		ActionType:      "MARK_PAID",
		PerformedBy:     userID,
		PerformedByName: user.Name,
		PerformedByRole: user.Role,
		Timestamp:       now,
		PerformedAt:     now,
		Comments:        comments,
		PreviousStatus:  models.StatusApproved,
		NewStatus:       models.StatusPaid,
		Metadata: map[string]interface{}{
			"paidAmount": paidAmount,
			"signature":  signature,
			"taskId":     task.ID,
		},
	})
	voucher.ActionHistory = datatypes.NewJSONType(actionHistory)

	if err := tx.Save(&voucher).Error; err != nil {
		return fmt.Errorf("failed to mark payment voucher as paid: %w", err)
	}

	// Symmetric to the GRN→PO cascade — when a PV becomes PAID, check whether
	// the parent PO is now fully delivered + all PVs paid and advance it to
	// COMPLETED. Closes both procurement chains.
	if err := s.CascadePVPaidToPO(tx, voucher.ID); err != nil {
		return fmt.Errorf("cascade PV paid → PO: %w", err)
	}

	go utils.SyncDocumentAs(s.db, "PAYMENT_VOUCHER", voucher.ID, userID)
	go LogDocumentEvent(s.db, DocumentEvent{
		OrganizationID: voucher.OrganizationID,
		DocumentID:     voucher.ID,
		DocumentType:   "payment_voucher",
		UserID:         userID,
		ActorName:      user.Name,
		ActorRole:      user.Role,
		Action:         "paid",
		Details: map[string]interface{}{
			"documentNumber": voucher.DocumentNumber,
			"paidAmount":     paidAmount,
		},
	})
	return nil
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
	if strings.ToUpper(task.Status) != "PENDING" && strings.ToUpper(task.Status) != "CLAIMED" {
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

	log.Printf("[DEBUG] Checking approval permission - User: %s, UserRole: %s, AssignedRole: %v",
		userID, user.Role, task.AssignedRole)

	if err := s.canUserActOnTask(tx, &task, &user); err != nil {
		tx.Rollback()
		return err
	}

	// Payment-execution tasks follow a different completion path: no stage
	// advancement, no workflow assignment updates — just flip the PV to PAID
	// and record a signed ActionHistory entry. Branch before touching the
	// workflow assignment so we don't require an assignment row for
	// standalone execution tasks.
	if strings.EqualFold(task.Kind, models.TaskKindPaymentExecution) {
		if err := s.completePaymentExecutionTask(tx, &task, userID, &user, signature, comments, time.Now()); err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Commit().Error; err != nil {
			return fmt.Errorf("failed to commit payment execution: %w", err)
		}
		return nil
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
		ID:             uuid.New().String(),
		OrganizationID: assignment.OrganizationID,
		WorkflowTaskID: taskID,
		StageNumber:    task.StageNumber,
		ApproverID:     userID,
		ApproverName:   user.Name,
		ApproverRole:   user.Role,
		ManNumber:      user.ManNumber,
		Position:       user.Position,
		Action:         "approved",
		Comments:       comments,
		Signature:      signature,
		ApprovedAt:     now,
		CreatedAt:      now,
	}

	if err := tx.Create(approvalRecord).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to record approval: %w", err)
	}

	go LogDocumentEvent(s.db, DocumentEvent{
		OrganizationID: assignment.OrganizationID,
		DocumentID:     assignment.EntityID,
		DocumentType:   strings.ToLower(assignment.EntityType),
		UserID:         userID,
		ActorName:      user.Name,
		ActorRole:      user.Role,
		Action:         "approved",
		Details:        map[string]interface{}{"stageNumber": task.StageNumber, "stageName": task.StageName},
	})

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
				"status": "COMPLETED",
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
			assignment.Status = "COMPLETED"
			assignment.CompletedAt = &now
			assignment.CurrentStage = len(stages)

			// Update the actual document status to "approved"
			if err := s.updateDocumentStatusScopedAs(tx, assignment.EntityType, assignment.EntityID, assignment.OrganizationID, models.StatusApproved, userID); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update document status: %w", err)
			}

			// Add action history entry to the document
			if err := s.addActionHistoryEntry(tx, assignment.EntityType, assignment.EntityID, userID, "WORKFLOW_COMPLETED", "Document approved through workflow system"); err != nil {
				// Log error but don't fail the approval
				fmt.Printf("Warning: failed to add action history entry: %v\n", err)
			}

			// Payment vouchers: after final approval, create a payment_execution
			// task so the PAID transition has an audit trail (claim + signature
			// + attributed actor) rather than being a blind endpoint flip.
			if strings.EqualFold(assignment.EntityType, "payment_voucher") {
				if err := s.createPaymentExecutionTask(tx, &assignment, now); err != nil {
					tx.Rollback()
					return fmt.Errorf("failed to create payment execution task: %w", err)
				}
			}

			// GRNs: cascade final approval to parent PO delivery tracking, then
			// auto-advance to COMPLETED — the old ConfirmGRN step was a vestige
			// of the pre-sign-off design and is no longer needed (the
			// receiver + certifier signatures captured before submit already
			// stand in for the warehouse-clerk confirmation).
			if strings.EqualFold(assignment.EntityType, "grn") {
				if err := s.cascadeGRNApprovalToPO(tx, assignment.EntityID); err != nil {
					tx.Rollback()
					return fmt.Errorf("post-approval GRN cascade: %w", err)
				}
				if err := tx.Model(&models.GoodsReceivedNote{}).
					Where("id = ?", assignment.EntityID).
					Updates(map[string]interface{}{
						"status":         models.StatusCompleted,
						"signoff_status": "COMPLETED",
						"updated_at":     now,
					}).Error; err != nil {
					tx.Rollback()
					return fmt.Errorf("auto-complete GRN: %w", err)
				}
				// Fire auto-PV-create if the org opted in. Logged but
				// non-fatal — primary GRN approval is the source of truth.
				if err := s.AutoCreatePVFromCompletedGRN(tx, assignment.EntityID); err != nil {
					fmt.Printf("Warning: AutoCreatePVFromCompletedGRN failed: %v\n", err)
				}
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
				Status: "PENDING",
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
			Type:           "document_approved",
			DocumentID:     assignment.EntityID,
			DocumentType:   assignment.EntityType,
			OrganizationID: assignment.OrganizationID,
			Action:         "workflow_completed",
			ActorID:        userID,
			Details:        "Document has been fully approved through workflow",
			Timestamp:      time.Now(),
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
			Type:           "approval_required",
			DocumentID:     assignment.EntityID,
			DocumentType:   assignment.EntityType,
			OrganizationID: assignment.OrganizationID,
			Action:         "next_stage_approval",
			ActorID:        userID,
			Details:        fmt.Sprintf("Document moved to next approval stage (%d)", assignment.CurrentStage),
			Timestamp:      time.Now(),
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
			Type:           "partial_approval",
			DocumentID:     assignment.EntityID,
			DocumentType:   assignment.EntityType,
			OrganizationID: assignment.OrganizationID,
			Action:         "partial_stage_approval",
			ActorID:        userID,
			Details:        fmt.Sprintf("Partial approval received for stage %d (%s)", stage.StageNumber, stage.StageName),
			Timestamp:      time.Now(),
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
func (s *WorkflowExecutionService) RejectWorkflowTask(ctx context.Context, taskID, userID, signature, reason, rejectionType string, returnToStage int) error {
	return s.RejectWorkflowTaskWithVersion(ctx, taskID, userID, signature, reason, 0, rejectionType, returnToStage)
}

// RejectWorkflowTaskWithVersion rejects a workflow task with version control for optimistic locking
func (s *WorkflowExecutionService) RejectWorkflowTaskWithVersion(ctx context.Context, taskID, userID, signature, reason string, expectedVersion int, rejectionType string, returnToStage int) error {
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
	if strings.ToUpper(task.Status) != "PENDING" && strings.ToUpper(task.Status) != "CLAIMED" {
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

	log.Printf("[DEBUG] Checking rejection permission - User: %s, UserRole: %s, AssignedRole: %v",
		userID, user.Role, task.AssignedRole)

	if err := s.canUserActOnTask(tx, &task, &user); err != nil {
		tx.Rollback()
		return err
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
			"status": "COMPLETED",
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

	// Determine action label for audit records
	isReturnToDraft := rejectionType == "return_to_draft"
	isReturnToPrevStage := rejectionType == "return_to_previous_stage"
	actionLabel := "rejected"
	if isReturnToDraft {
		actionLabel = "returned_to_draft"
	} else if isReturnToPrevStage {
		actionLabel = "returned_for_revision"
	}

	// Record this rejection/return in stage approval records
	approvalRecord := &models.StageApprovalRecord{
		ID:             uuid.New().String(),
		OrganizationID: assignment.OrganizationID,
		WorkflowTaskID: taskID,
		StageNumber:    task.StageNumber,
		ApproverID:     userID,
		ApproverName:   user.Name,
		ApproverRole:   user.Role,
		ManNumber:      user.ManNumber,
		Position:       user.Position,
		Action:         actionLabel,
		Comments:       reason,
		Signature:      signature,
		ApprovedAt:     now,
		CreatedAt:      now,
	}

	if err := tx.Create(approvalRecord).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to record rejection: %w", err)
	}

	go LogDocumentEvent(s.db, DocumentEvent{
		OrganizationID: assignment.OrganizationID,
		DocumentID:     assignment.EntityID,
		DocumentType:   strings.ToLower(assignment.EntityType),
		UserID:         userID,
		ActorName:      user.Name,
		ActorRole:      user.Role,
		Action:         actionLabel,
		Details:        map[string]interface{}{"stageNumber": task.StageNumber, "stageName": task.StageName, "reason": reason},
	})

	// Add stage execution to history
	stageExecution := models.StageExecution{
		StageNumber:  task.StageNumber,
		StageName:    task.StageName,
		ApproverID:   userID,
		ApproverName: user.Name,
		ApproverRole: user.Role,
		Action:       actionLabel,
		Comments:     reason,
		Signature:    signature,
		ExecutedAt:   now,
	}

	if err := assignment.AddStageExecution(stageExecution); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update stage history: %w", err)
	}

	// Determine notification type
	notificationType := "document_rejected"
	notificationAction := "workflow_rejected"

	if isReturnToPrevStage && task.StageNumber <= 1 {
		tx.Rollback()
		return fmt.Errorf("cannot return to previous stage: task is already at stage 1")
	}

	if isReturnToPrevStage {
		// RETURN TO PREVIOUS STAGE: keep workflow active, create task at previous stage
		prevStageNumber := task.StageNumber - 1

		// Load the workflow to get stage definitions
		var workflow models.Workflow
		if err := tx.Where("id = ?", assignment.WorkflowID).First(&workflow).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("workflow not found: %w", err)
		}
		stages, err := workflow.GetStages()
		if err != nil || prevStageNumber < 1 || prevStageNumber > len(stages) {
			tx.Rollback()
			return fmt.Errorf("failed to get previous stage definition")
		}
		prevStage := stages[prevStageNumber-1]

		// Move assignment back to previous stage, keep workflow active
		assignment.CurrentStage = prevStageNumber
		assignment.Status = "IN_PROGRESS"
		assignment.UpdatedAt = time.Now()

		// Update document status to "revision"
		if err := s.updateDocumentStatusScopedAs(tx, assignment.EntityType, assignment.EntityID, assignment.OrganizationID, models.StatusRevision, userID); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update document status: %w", err)
		}

		// GRN-specific: revising the document invalidates the captured
		// receiver + certifier signatures because line items can change.
		// Reset the sign-off lifecycle so the new revision walks the same
		// PENDING_RECEIVER → PENDING_CERTIFIER → READY path again.
		if strings.EqualFold(assignment.EntityType, "grn") {
			if err := tx.Model(&models.GoodsReceivedNote{}).
				Where("id = ?", assignment.EntityID).
				Updates(map[string]interface{}{
					"signoff_status":         "PENDING_RECEIVER",
					"received_by_name":       "",
					"received_by_signature":  "",
					"received_at":            nil,
					"certified_by_id":        "",
					"certified_by_name":      "",
					"certified_by_signature": "",
					"certified_at":           nil,
					"updated_at":             time.Now(),
				}).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("reset GRN signoff on revision: %w", err)
			}
		}

		// Create a new task at the previous stage
		nextTaskPriority := s.getDocumentPriority(tx, assignment.EntityID, assignment.EntityType)
		prevTask := &models.WorkflowTask{
			ID:                   uuid.New().String(),
			OrganizationID:       assignment.OrganizationID,
			WorkflowAssignmentID: assignment.ID,
			EntityID:             assignment.EntityID,
			EntityType:           assignment.EntityType,
			StageNumber:          prevStage.StageNumber,
			StageName:            prevStage.StageName,
			AssignmentType:       "role",
			AssignedRole:         &prevStage.RequiredRole,
			Status: "PENDING",
			Priority:             nextTaskPriority,
			Version:              1,
			CreatedAt:            time.Now(),
		}

		// Set due date
		var prevCalculatedDueDate time.Time
		if prevStage.TimeoutHours != nil && *prevStage.TimeoutHours > 0 {
			prevCalculatedDueDate = time.Now().Add(time.Duration(*prevStage.TimeoutHours) * time.Hour)
		} else {
			prevCalculatedDueDate = time.Now().Add(7 * 24 * time.Hour)
		}
		prevDocDueDate := s.getDocumentDueDate(tx, assignment.EntityID, assignment.EntityType)
		if prevDocDueDate != nil && prevDocDueDate.Before(prevCalculatedDueDate) {
			prevTask.DueDate = prevDocDueDate
		} else {
			prevTask.DueDate = &prevCalculatedDueDate
		}

		if err := tx.Create(prevTask).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create task for previous stage: %w", err)
		}

		// Add action history
		actionMessage := fmt.Sprintf("Returned to %s (Stage %d) for revision by %s: %s", prevStage.StageName, prevStageNumber, user.Name, reason)
		if err := s.addActionHistoryEntry(tx, assignment.EntityType, assignment.EntityID, userID, "RETURNED_FOR_REVISION", actionMessage); err != nil {
			fmt.Printf("Warning: failed to add action history entry: %v\n", err)
		}

		log.Printf("[Workflow] Task %s returned to stage %d (%s) by user %s: %s", taskID, prevStageNumber, prevStage.StageName, userID, reason)
		notificationType = "document_returned_for_revision"
		notificationAction = "workflow_returned_for_revision"
	} else if isReturnToDraft {
		// RETURN TO DRAFT: send document back to draft, cancel the workflow
		// The requester can edit and resubmit, which will start a new workflow
		assignment.Status = "RETURNED"
		assignment.CompletedAt = &now
		assignment.UpdatedAt = time.Now()

		// Update document status to "draft"
		if err := s.updateDocumentStatusScopedAs(tx, assignment.EntityType, assignment.EntityID, assignment.OrganizationID, models.StatusDraft, userID); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update document status: %w", err)
		}

		// Add action history
		actionMessage := fmt.Sprintf("Returned to draft by %s: %s", user.Name, reason)
		if err := s.addActionHistoryEntry(tx, assignment.EntityType, assignment.EntityID, userID, "RETURNED_TO_DRAFT", actionMessage); err != nil {
			fmt.Printf("Warning: failed to add action history entry: %v\n", err)
		}

		log.Printf("[Workflow] Task %s returned to draft by user %s: %s", taskID, userID, reason)
		notificationType = "document_returned_to_draft"
		notificationAction = "workflow_returned_to_draft"
	} else {
		// FULL REJECTION: terminate the workflow
		assignment.Status = "REJECTED"
		assignment.CompletedAt = &now
		assignment.UpdatedAt = time.Now()

		if strings.EqualFold(assignment.EntityType, "purchase_order") {
			// PO: revert to DRAFT (not permanently rejected); linked REQ also reverts
			var po models.PurchaseOrder
			if err := tx.First(&po, "id = ?", assignment.EntityID).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to load purchase order: %w", err)
			}
			prevStatus := po.Status
			if err := s.updateDocumentStatusScopedAs(tx, "purchase_order", assignment.EntityID, assignment.OrganizationID, models.StatusDraft, userID); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update document status: %w", err)
			}
			if err := s.addActionHistoryEntryWithMeta(tx, "purchase_order", assignment.EntityID, userID,
				"WORKFLOW_REJECTED_REVERTED_TO_DRAFT", reason,
				prevStatus, "DRAFT",
				map[string]interface{}{"approvalStage": map[string]interface{}{"from": task.StageNumber, "to": 0}},
			); err != nil {
				fmt.Printf("Warning: failed to add PO action history entry: %v\n", err)
			}

			// Revert the linked REQ to DRAFT as well
			if po.SourceRequisitionId != nil {
				reqID := *po.SourceRequisitionId
				var req models.Requisition
				if err := tx.First(&req, "id = ?", reqID).Error; err == nil {
					prevReqStatus := req.Status
					if err := s.updateDocumentStatusScopedAs(tx, "requisition", reqID, assignment.OrganizationID, models.StatusDraft, userID); err != nil {
						tx.Rollback()
						return fmt.Errorf("failed to revert linked requisition: %w", err)
					}
					tx.Model(&models.WorkflowAssignment{}).
						Where("entity_id = ? AND UPPER(status) = 'IN_PROGRESS'", reqID).
						Updates(map[string]interface{}{"status": "RETURNED", "completed_at": now, "updated_at": time.Now()})
					tx.Model(&models.WorkflowTask{}).
						Where("entity_id = ? AND UPPER(status) IN ('PENDING', 'CLAIMED')", reqID).
						Updates(map[string]interface{}{"status": "CANCELLED", "updated_at": time.Now()})
					if err := s.addActionHistoryEntryWithMeta(tx, "requisition", reqID, userID,
						"REVERTED_TO_DRAFT_BY_PO_REJECTION",
						fmt.Sprintf("Linked PO %s was rejected. Requisition returned to DRAFT for revision.", po.DocumentNumber),
						prevReqStatus, "DRAFT",
						map[string]interface{}{"triggeredBy": map[string]interface{}{"type": "purchase_order", "id": po.ID, "documentNumber": po.DocumentNumber}},
					); err != nil {
						fmt.Printf("Warning: failed to add REQ action history entry: %v\n", err)
					}
				}
			}
		} else if strings.EqualFold(assignment.EntityType, "payment_voucher") {
			// PV rejection:
			//   - Goods-first (LinkedGRN set): PV + GRN both → DRAFT; PO stays APPROVED
			//   - Payment-first (no LinkedGRN): PV → DRAFT only; PO stays APPROVED
			var pv models.PaymentVoucher
			if err := tx.First(&pv, "id = ?", assignment.EntityID).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to load payment voucher: %w", err)
			}
			prevStatus := pv.Status
			if err := s.updateDocumentStatusScopedAs(tx, "payment_voucher", assignment.EntityID, assignment.OrganizationID, models.StatusDraft, userID); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update document status: %w", err)
			}
			if err := s.addActionHistoryEntryWithMeta(tx, "payment_voucher", assignment.EntityID, userID,
				"WORKFLOW_REJECTED_REVERTED_TO_DRAFT", reason,
				prevStatus, "DRAFT",
				map[string]interface{}{"approvalStage": map[string]interface{}{"from": task.StageNumber, "to": 0}},
			); err != nil {
				fmt.Printf("Warning: failed to add PV action history entry: %v\n", err)
			}

			// Goods-first: also revert the linked GRN to DRAFT (receiving must be redone)
			if pv.LinkedGRN != "" {
				var grn models.GoodsReceivedNote
				if err := tx.Where("document_number = ? AND organization_id = ?", pv.LinkedGRN, assignment.OrganizationID).
					First(&grn).Error; err == nil {
					prevGRNStatus := grn.Status
					if err := s.updateDocumentStatusScopedAs(tx, "grn", grn.ID, assignment.OrganizationID, models.StatusDraft, userID); err != nil {
						tx.Rollback()
						return fmt.Errorf("failed to revert linked GRN: %w", err)
					}
					// Cancel any in-progress GRN workflow assignments
					tx.Model(&models.WorkflowAssignment{}).
						Where("entity_id = ? AND UPPER(status) = 'IN_PROGRESS'", grn.ID).
						Updates(map[string]interface{}{"status": "RETURNED", "completed_at": now, "updated_at": time.Now()})
					// Cancel pending/claimed GRN tasks
					tx.Model(&models.WorkflowTask{}).
						Where("entity_id = ? AND UPPER(status) IN ('PENDING', 'CLAIMED')", grn.ID).
						Updates(map[string]interface{}{"status": "CANCELLED", "updated_at": time.Now()})
					if err := s.addActionHistoryEntryWithMeta(tx, "grn", grn.ID, userID,
						"REVERTED_TO_DRAFT_BY_PV_REJECTION",
						fmt.Sprintf("Linked PV %s was rejected. GRN returned to DRAFT for correction.", pv.DocumentNumber),
						prevGRNStatus, "DRAFT",
						map[string]interface{}{"triggeredBy": map[string]interface{}{"type": "payment_voucher", "id": pv.ID, "documentNumber": pv.DocumentNumber}},
					); err != nil {
						fmt.Printf("Warning: failed to add GRN action history entry: %v\n", err)
					}
				}
			}
		} else if strings.EqualFold(assignment.EntityType, "grn") {
			// GRN: revert to DRAFT — receiving issue, upstream docs (PO/PV) are unaffected
			var grn models.GoodsReceivedNote
			if err := tx.First(&grn, "id = ?", assignment.EntityID).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to load GRN: %w", err)
			}
			prevStatus := grn.Status
			if err := s.updateDocumentStatusScopedAs(tx, "grn", assignment.EntityID, assignment.OrganizationID, models.StatusDraft, userID); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update GRN status: %w", err)
			}
			if err := s.addActionHistoryEntryWithMeta(tx, "grn", assignment.EntityID, userID,
				"WORKFLOW_REJECTED_REVERTED_TO_DRAFT", reason,
				prevStatus, "DRAFT",
				map[string]interface{}{"approvalStage": map[string]interface{}{"from": task.StageNumber, "to": 0}},
			); err != nil {
				fmt.Printf("Warning: failed to add GRN action history entry: %v\n", err)
			}
		} else if strings.EqualFold(assignment.EntityType, "requisition") {
			// REQ: revert to DRAFT — user corrects and re-submits; no cascade up
			var req models.Requisition
			if err := tx.First(&req, "id = ?", assignment.EntityID).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to load requisition: %w", err)
			}
			prevStatus := req.Status
			if err := s.updateDocumentStatusScopedAs(tx, "requisition", assignment.EntityID, assignment.OrganizationID, models.StatusDraft, userID); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update requisition status: %w", err)
			}
			if err := s.addActionHistoryEntryWithMeta(tx, "requisition", assignment.EntityID, userID,
				"WORKFLOW_REJECTED_REVERTED_TO_DRAFT", reason,
				prevStatus, "DRAFT",
				map[string]interface{}{"approvalStage": map[string]interface{}{"from": task.StageNumber, "to": 0}},
			); err != nil {
				fmt.Printf("Warning: failed to add REQ action history entry: %v\n", err)
			}
		} else {
			// Standard: all other document types → REJECTED permanently
			if err := s.updateDocumentStatusScopedAs(tx, assignment.EntityType, assignment.EntityID, assignment.OrganizationID, models.StatusRejected, userID); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update document status: %w", err)
			}
			if err := s.addActionHistoryEntry(tx, assignment.EntityType, assignment.EntityID, userID, "WORKFLOW_REJECTED", reason); err != nil {
				fmt.Printf("Warning: failed to add action history entry: %v\n", err)
			}
		}
	}

	if err := tx.Save(&assignment).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update workflow assignment: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit workflow rejection: %w", err)
	}

	// Send notification
	if s.notificationService != nil {
		notificationEvent := NotificationEvent{
			Type:           notificationType,
			DocumentID:     assignment.EntityID,
			DocumentType:   assignment.EntityType,
			OrganizationID: assignment.OrganizationID,
			Action:         notificationAction,
			ActorID:        userID,
			Details:        reason,
			Timestamp:      time.Now(),
		}

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
			StageNumber:    stage.StageNumber,
			StageName:      stage.StageName,
			RequiredRole:   requiredRoleDisplay,
			Status: "PENDING",
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
		if stage.StageNumber < assignment.CurrentStage && strings.ToUpper(stageInfo.Status) == "PENDING" {
			stageInfo.Status = "COMPLETED"
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

	// Query users with the required role — handles both plain names and UUIDs
	var approvers []ApproverInfo
	assignedRole := *currentTask.AssignedRole

	if _, parseErr := uuid.Parse(assignedRole); parseErr == nil {
		// UUID — look up the org role record
		var orgRole models.OrganizationRole
		if err = s.db.Where("id = ?", assignedRole).First(&orgRole).Error; err != nil {
			return []ApproverInfo{}, nil // role not found, no approvers
		}
		if orgRole.IsSystemRole {
			// System role UUID: find users whose role name matches
			err = s.db.Table("users").
				Select("users.id, users.name, users.email, users.role").
				Where("users.current_organization_id = ? AND users.active = ? AND LOWER(users.role) = LOWER(?)",
					organizationID, true, orgRole.Name).
				Find(&approvers).Error
		} else {
			// Custom org role UUID: find via user_organization_roles
			err = s.db.Table("users").
				Select("users.id, users.name, users.email, users.role").
				Joins("INNER JOIN user_organization_roles uor ON uor.user_id = users.id").
				Where("users.current_organization_id = ? AND users.active = ? AND uor.role_id = ? AND uor.organization_id = ? AND uor.active = ?",
					organizationID, true, assignedRole, organizationID, true).
				Find(&approvers).Error
		}
	} else {
		// Plain role name
		err = s.db.Table("users").
			Select("users.id, users.name, users.email, users.role").
			Where("users.current_organization_id = ? AND users.active = ? AND LOWER(users.role) = LOWER(?)",
				organizationID, true, assignedRole).
			Find(&approvers).Error
	}

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

	// Read task and user first to perform role-based auth check
	var task models.WorkflowTask
	if err := tx.Where("id = ? AND UPPER(status) = ?", taskID, "PENDING").First(&task).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("task not found or not available: %w", err)
	}

	var user models.User
	if err := tx.Where("id = ?", userID).First(&user).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("user not found: %w", err)
	}

	if err := s.canUserActOnTask(tx, &task, &user); err != nil {
		tx.Rollback()
		return err
	}

	// Atomic claim operation with optimistic locking
	result := tx.Model(&task).
		Where("id = ? AND UPPER(status) = ? AND (claimed_by IS NULL OR claim_expiry < ?)",
			taskID, "PENDING", time.Now()).
		Updates(map[string]interface{}{
			"claimed_by":   userID,
			"claimed_at":   time.Now(),
			"claim_expiry": time.Now().Add(30 * time.Minute), // 30-minute claim
			"status": "CLAIMED",
			"version":      gorm.Expr("version + 1"),
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
			"claimed_by":   nil,
			"claimed_at":   nil,
			"claim_expiry": nil,
			"status": "PENDING",
			"version":      gorm.Expr("version + 1"),
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

// updateDocumentStatus updates the status of the actual document when a
// workflow step completes or is rolled back.
//
// Callers pass the assignment's organization_id so the UPDATE is scoped to that
// tenant — defense in depth against a task somehow referencing an entity ID
// that belongs to another org. An empty orgID falls back to ID-only matching
// for backward compatibility (e.g. utility callers that do not yet have org
// context), so do not remove the empty check without auditing every call site.
// cascadeGRNApprovalToPO recomputes the parent PO's delivery_status and per-item
// received quantities from the set of non-cancelled GRNs for that PO. Called
// from the terminal-approve path when entity_type == "grn".
// CascadeGRNApprovalToPO is the exported wrapper used by handlers that
// complete a GRN outside the workflow (e.g. MarkGRNComplete).
func (s *WorkflowExecutionService) CascadeGRNApprovalToPO(tx *gorm.DB, grnID string) error {
	return s.cascadeGRNApprovalToPO(tx, grnID)
}

// resolveDeliveryFromGRNs sets ReceivedQuantity on each PO line by aggregating
// quantities received across all GRN lines. Lines are matched by ItemCode when
// present (a stable SKU snapshotted from the PO line, so it survives description
// edits and disambiguates duplicate descriptions) and fall back to a normalized
// Description otherwise. Returns the updated items plus whether all / any lines
// are received.
func resolveDeliveryFromGRNs(items []types.POItem, grnItemLists [][]types.GRNItem) ([]types.POItem, bool, bool) {
	receivedByCode := make(map[string]int)
	receivedByDesc := make(map[string]int)
	for _, list := range grnItemLists {
		for _, it := range list {
			if code := strings.TrimSpace(strings.ToLower(it.ItemCode)); code != "" {
				receivedByCode[code] += it.QuantityReceived
			}
			receivedByDesc[strings.TrimSpace(strings.ToLower(it.Description))] += it.QuantityReceived
		}
	}

	allFull, anyReceived := true, false
	for i := range items {
		code := strings.TrimSpace(strings.ToLower(items[i].ItemCode))
		desc := strings.TrimSpace(strings.ToLower(items[i].Description))
		received := 0
		if code != "" {
			if v, ok := receivedByCode[code]; ok {
				received = v
			} else {
				received = receivedByDesc[desc]
			}
		} else {
			received = receivedByDesc[desc]
		}
		items[i].ReceivedQuantity = received
		if received > 0 {
			anyReceived = true
		}
		if received < items[i].Quantity {
			allFull = false
		}
	}
	return items, allFull, anyReceived
}

func (s *WorkflowExecutionService) cascadeGRNApprovalToPO(tx *gorm.DB, grnID string) error {
	var grn models.GoodsReceivedNote
	if err := tx.Where("id = ?", grnID).First(&grn).Error; err != nil {
		return fmt.Errorf("cascade: load GRN: %w", err)
	}
	if grn.PODocumentNumber == "" {
		return nil // payment-first GRN with no PO link
	}

	var po models.PurchaseOrder
	if err := tx.Where("document_number = ? AND organization_id = ?",
		grn.PODocumentNumber, grn.OrganizationID).First(&po).Error; err != nil {
		return fmt.Errorf("cascade: load PO: %w", err)
	}

	var grns []models.GoodsReceivedNote
	if err := tx.Where("po_document_number = ? AND organization_id = ? AND UPPER(status) != ?",
		po.DocumentNumber, po.OrganizationID, "CANCELLED").Find(&grns).Error; err != nil {
		return fmt.Errorf("cascade: list GRNs: %w", err)
	}

	grnItemLists := make([][]types.GRNItem, 0, len(grns))
	for _, g := range grns {
		grnItemLists = append(grnItemLists, g.Items.Data())
	}
	items, allFull, anyReceived := resolveDeliveryFromGRNs(po.Items.Data(), grnItemLists)

	newDeliveryStatus := models.DeliveryStatusNotDelivered
	switch {
	case allFull && anyReceived:
		newDeliveryStatus = models.DeliveryStatusFullyDelivered
	case anyReceived:
		newDeliveryStatus = models.DeliveryStatusPartiallyDelivered
	}

	updates := map[string]interface{}{
		"items":           datatypes.NewJSONType(items),
		"delivery_status": newDeliveryStatus,
		"updated_at":      time.Now(),
	}

	// Terminal cascade: once goods are fully received, advance the PO status
	// based on payment side. Three branches, all gated on APPROVED today (the
	// only state from which delivery can land):
	//   - zero linked PVs           → leave APPROVED (PV doesn't exist yet)
	//   - some PVs not PAID         → FULFILLED (goods in, money pending)
	//   - all PVs PAID              → COMPLETED (procurement chain closed)
	// Applies to both procurement flows (goods_first and payment_first).
	if newDeliveryStatus == models.DeliveryStatusFullyDelivered &&
		strings.ToUpper(po.Status) == "APPROVED" {
		var pvs []models.PaymentVoucher
		if err := tx.Where("linked_po = ? AND organization_id = ? AND UPPER(status) != ?",
			po.DocumentNumber, po.OrganizationID, "CANCELLED").Find(&pvs).Error; err != nil {
			return fmt.Errorf("cascade: list linked PVs: %w", err)
		}
		if len(pvs) > 0 {
			allPaid := true
			for _, pv := range pvs {
				if strings.ToUpper(pv.Status) != "PAID" {
					allPaid = false
					break
				}
			}
			if allPaid {
				updates["status"] = models.StatusCompleted
			} else {
				updates["status"] = models.StatusFulfilled
			}
		}
	}

	return tx.Model(&models.PurchaseOrder{}).
		Where("id = ?", po.ID).
		Updates(updates).Error
}

// AutoCreatePVFromCompletedGRN inspects the org settings + procurement flow
// and creates a draft PV from the GRN's parent PO when AutoCreatePVFromPO is
// on. Safe to call after a GRN reaches COMPLETED — it no-ops when:
//   - the org flag is off
//   - flow is payment_first (PV already exists upstream by definition)
//   - a non-CANCELLED/non-REJECTED PV already exists for this PO
//   - no PO is linked
func (s *WorkflowExecutionService) AutoCreatePVFromCompletedGRN(tx *gorm.DB, grnID string) error {
	var grn models.GoodsReceivedNote
	if err := tx.Where("id = ?", grnID).First(&grn).Error; err != nil {
		return nil
	}
	if grn.PODocumentNumber == "" {
		return nil
	}
	var po models.PurchaseOrder
	if err := tx.Where("document_number = ? AND organization_id = ?",
		grn.PODocumentNumber, grn.OrganizationID).First(&po).Error; err != nil {
		return nil
	}

	// Resolve flow + automation flags.
	var settings models.OrganizationSettings
	if err := tx.Where("organization_id = ?", grn.OrganizationID).First(&settings).Error; err != nil {
		return nil
	}
	flow := utils.ResolveProcurementFlow(po.ProcurementFlow, settings.ProcurementFlow)
	if flow == "payment_first" {
		return nil
	}
	if !settings.AutoCreatePVFromPO {
		return nil
	}

	// Defense-in-depth: only auto-create once goods are actually received.
	// (Mirrors the GRN-status gate the manual/from-po validators enforce.)
	grnStatus := strings.ToUpper(grn.Status)
	if grnStatus != "APPROVED" && grnStatus != "COMPLETED" {
		return nil
	}

	// Skip if a usable PV already exists.
	var existing int64
	tx.Model(&models.PaymentVoucher{}).
		Where("linked_po = ? AND organization_id = ? AND UPPER(status) NOT IN ('CANCELLED','REJECTED')",
			po.DocumentNumber, po.OrganizationID).
		Count(&existing)
	if existing > 0 {
		return nil
	}

	// Build PV line items from the GRN: for each GRN line, look up the
	// matching PO line by trimmed-lowercase description and compute
	// amount = qtyReceived * unitPrice. Skip zero-quantity lines.
	poItemByDesc := make(map[string]types.POItem)
	for _, it := range po.Items.Data() {
		key := strings.TrimSpace(strings.ToLower(it.Description))
		poItemByDesc[key] = it
	}
	pvItems := make([]types.PaymentItem, 0)
	var pvTotal float64
	for _, gi := range grn.Items.Data() {
		if gi.QuantityReceived <= 0 {
			continue
		}
		key := strings.TrimSpace(strings.ToLower(gi.Description))
		poItem, ok := poItemByDesc[key]
		var unitPrice float64
		if ok {
			unitPrice = poItem.UnitPrice
		}
		amount := float64(gi.QuantityReceived) * unitPrice
		pvItems = append(pvItems, types.PaymentItem{
			Description: gi.Description,
			Amount:      amount,
			GLCode:      po.GLCode,
			TaxAmount:   0,
		})
		pvTotal += amount
	}

	// Backstop: an auto-created PV may never exceed the PO total. Scale the line
	// items proportionally so the PV's Amount always equals Σ(item amounts) — never
	// leave the header total and line items inconsistent.
	if po.TotalAmount > 0 && pvTotal > po.TotalAmount+0.01 {
		scale := po.TotalAmount / pvTotal
		var scaledTotal float64
		for i := range pvItems {
			pvItems[i].Amount = math.Round(pvItems[i].Amount*scale*100) / 100
			scaledTotal += pvItems[i].Amount
		}
		pvTotal = scaledTotal
	}

	now := time.Now()
	docNum := utils.GenerateDocumentNumber("PV")
	pv := models.PaymentVoucher{
		ID:             uuid.New().String(),
		OrganizationID: po.OrganizationID,
		DocumentNumber: docNum,
		LinkedPO:       po.DocumentNumber,
		LinkedGRN:      grn.DocumentNumber,
		VendorID:       po.VendorID,
		VendorName:     po.VendorName,
		Amount:         pvTotal,
		Currency:       po.Currency,
		Status:         models.StatusDraft,
		CreatedBy:      grn.CreatedBy,
		Description:    fmt.Sprintf("Auto-created from PO %s after GRN %s completion", po.DocumentNumber, grn.DocumentNumber),
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	pv.Items = datatypes.NewJSONType(pvItems)
	pv.ApprovalHistory = datatypes.NewJSONType([]types.ApprovalRecord{})
	pv.ActionHistory = datatypes.NewJSONType([]types.ActionHistoryEntry{{
		ID:          uuid.New().String(),
		Action:      "AUTO_CREATED",
		ActionType:  "CREATE",
		Timestamp:   now,
		PerformedAt: now,
		Comments:    "Auto-created via AutoCreatePVFromPO setting",
		NewStatus:   models.StatusDraft,
		Metadata: map[string]interface{}{
			"itemsCopiedFromGRN": grn.DocumentNumber,
			"lineCount":          len(pvItems),
		},
	}})
	return tx.Create(&pv).Error
}

// CascadePVPaidToPO is the symmetric trigger called after a PV transitions
// to PAID. It advances the parent PO to COMPLETED when delivery is already
// FULLY_DELIVERED and every other linked PV is also PAID. Accepts both
// APPROVED (payment-first: PV paid before/at the same time as final delivery)
// and FULFILLED (goods-first: PO was parked at FULFILLED awaiting payment)
// as the source state. Without this, the PO would stay stuck indefinitely.
func (s *WorkflowExecutionService) CascadePVPaidToPO(tx *gorm.DB, pvID string) error {
	var pv models.PaymentVoucher
	if err := tx.Where("id = ?", pvID).First(&pv).Error; err != nil {
		return fmt.Errorf("cascade: load PV: %w", err)
	}
	if pv.LinkedPO == "" {
		return nil
	}

	var po models.PurchaseOrder
	if err := tx.Where("document_number = ? AND organization_id = ?",
		pv.LinkedPO, pv.OrganizationID).First(&po).Error; err != nil {
		return fmt.Errorf("cascade: load PO: %w", err)
	}
	poStatus := strings.ToUpper(po.Status)
	if poStatus != "APPROVED" && poStatus != models.StatusFulfilled {
		return nil
	}
	if po.DeliveryStatus != models.DeliveryStatusFullyDelivered {
		return nil
	}

	var pvs []models.PaymentVoucher
	if err := tx.Where("linked_po = ? AND organization_id = ? AND UPPER(status) != ?",
		po.DocumentNumber, po.OrganizationID, "CANCELLED").Find(&pvs).Error; err != nil {
		return fmt.Errorf("cascade: list linked PVs: %w", err)
	}
	for _, p := range pvs {
		if strings.ToUpper(p.Status) != "PAID" {
			return nil
		}
	}

	return tx.Model(&models.PurchaseOrder{}).
		Where("id = ?", po.ID).
		Updates(map[string]interface{}{
			"status":     models.StatusCompleted,
			"updated_at": time.Now(),
		}).Error
}

func (s *WorkflowExecutionService) updateDocumentStatus(tx *gorm.DB, entityType, entityID, newStatus string) error {
	return s.updateDocumentStatusScopedAs(tx, entityType, entityID, "", newStatus, "")
}

func (s *WorkflowExecutionService) updateDocumentStatusScoped(tx *gorm.DB, entityType, entityID, organizationID, newStatus string) error {
	return s.updateDocumentStatusScopedAs(tx, entityType, entityID, organizationID, newStatus, "")
}

// updateDocumentStatusScopedAs is the actor-attributed form. actorID is the
// authenticated user who triggered the status change (approver, rejector,
// submitter) — it populates documents.updated_by via the sync. Pass "" for
// system/automated transitions (e.g. auto-approval, worker expiry).
func (s *WorkflowExecutionService) updateDocumentStatusScopedAs(tx *gorm.DB, entityType, entityID, organizationID, newStatus, actorID string) error {
	addOrgScope := func(q *gorm.DB) *gorm.DB {
		if organizationID != "" {
			return q.Where("organization_id = ?", organizationID)
		}
		return q
	}
	var err error
	switch entityType {
	case "REQUISITION", "requisition":
		err = addOrgScope(tx.Model(&models.Requisition{}).Where("id = ?", entityID)).Update("status", newStatus).Error
	case "BUDGET", "budget":
		err = addOrgScope(tx.Model(&models.Budget{}).Where("id = ?", entityID)).Update("status", newStatus).Error
	case "PURCHASE_ORDER", "purchase_order":
		err = addOrgScope(tx.Model(&models.PurchaseOrder{}).Where("id = ?", entityID)).Update("status", newStatus).Error
	case "PAYMENT_VOUCHER", "payment_voucher":
		err = addOrgScope(tx.Model(&models.PaymentVoucher{}).Where("id = ?", entityID)).Update("status", newStatus).Error
	case "GRN", "grn":
		err = addOrgScope(tx.Model(&models.GoodsReceivedNote{}).Where("id = ?", entityID)).Update("status", newStatus).Error
	default:
		return fmt.Errorf("unsupported entity type: %s", entityType)
	}
	if err != nil {
		return err
	}
	// Keep the generic documents index in sync after every status change.
	go utils.SyncDocumentAs(s.db, entityType, entityID, actorID)
	return nil
}

// addActionHistoryEntry adds an action history entry to the document
func (s *WorkflowExecutionService) addActionHistoryEntry(tx *gorm.DB, entityType, entityID, userID, action, comments string) error {
	now := time.Now()
	actionEntry := types.ActionHistoryEntry{
		ID:              uuid.New().String(),
		Action:          action,
		ActionType:      action,
		PerformedBy:     userID,
		PerformedByName: "", // Will be filled by caller if needed
		PerformedByRole: "", // Will be filled by caller if needed
		Timestamp:       now,
		PerformedAt:     now,
		Comments:        comments,
		PreviousStatus:  "", // Could be enhanced to track status transitions
		NewStatus:       "APPROVED",
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

// addActionHistoryEntryWithMeta is like addActionHistoryEntry but also records
// previousStatus, newStatus, and changedFields for full audit snapshots.
func (s *WorkflowExecutionService) addActionHistoryEntryWithMeta(
	tx *gorm.DB, entityType, entityID, userID, action, comments,
	previousStatus, newStatus string, changedFields map[string]interface{},
) error {
	now2 := time.Now()
	actionEntry := types.ActionHistoryEntry{
		ID:             uuid.New().String(),
		Action:         action,
		ActionType:     action,
		PerformedBy:    userID,
		Timestamp:      now2,
		PerformedAt:    now2,
		Comments:       comments,
		PreviousStatus: previousStatus,
		NewStatus:      newStatus,
		ChangedFields:  changedFields,
	}

	switch entityType {
	case "REQUISITION", "requisition":
		var requisition models.Requisition
		if err := tx.Where("id = ?", entityID).First(&requisition).Error; err != nil {
			return err
		}
		history := requisition.ActionHistory.Data()
		history = append(history, actionEntry)
		requisition.ActionHistory = datatypes.NewJSONType(history)
		return tx.Save(&requisition).Error

	case "BUDGET", "budget":
		var budget models.Budget
		if err := tx.Where("id = ?", entityID).First(&budget).Error; err != nil {
			return err
		}
		history := budget.ActionHistory.Data()
		history = append(history, actionEntry)
		budget.ActionHistory = datatypes.NewJSONType(history)
		return tx.Save(&budget).Error

	case "PURCHASE_ORDER", "purchase_order":
		var po models.PurchaseOrder
		if err := tx.Where("id = ?", entityID).First(&po).Error; err != nil {
			return err
		}
		history := po.ActionHistory.Data()
		history = append(history, actionEntry)
		po.ActionHistory = datatypes.NewJSONType(history)
		return tx.Save(&po).Error

	case "PAYMENT_VOUCHER", "payment_voucher":
		var pv models.PaymentVoucher
		if err := tx.Where("id = ?", entityID).First(&pv).Error; err != nil {
			return err
		}
		history := pv.ActionHistory.Data()
		history = append(history, actionEntry)
		pv.ActionHistory = datatypes.NewJSONType(history)
		return tx.Save(&pv).Error

	case "GRN", "grn":
		var grn models.GoodsReceivedNote
		if err := tx.Where("id = ?", entityID).First(&grn).Error; err != nil {
			return err
		}
		history := grn.ActionHistory.Data()
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
		// First, check the workflow's own conditions for auto-PO generation.
		// This handles the case where a requisition goes through workflow stages
		// and upon final approval, should auto-create a PO.
		var assignment models.WorkflowAssignment
		if err := s.db.Where("entity_id = ?", entityID).
			Preload("Workflow").
			First(&assignment).Error; err == nil && assignment.Workflow != nil {

			conditions, _ := assignment.Workflow.GetConditions()
			if conditions != nil && conditions.AutoGeneratePO {
				// Workflow has auto-PO generation enabled
				var requisition models.Requisition
				if err := s.db.Where("id = ?", entityID).First(&requisition).Error; err != nil {
					return fmt.Errorf("failed to get requisition for auto-PO: %w", err)
				}

				targetStatus := "DRAFT"
				if conditions.AutoApprovePO {
					targetStatus = "APPROVED"
				}

				result, err := s.automationService.CreatePurchaseOrderFromRequisitionWithStatus(
					ctx, &requisition, targetStatus,
				)
				if err != nil {
					return fmt.Errorf("failed to auto-create PO from workflow config: %w", err)
				}

				if result != nil && result.Success {
					// Update requisition with auto-created PO info
					autoCreatedPO := map[string]interface{}{
						"id":      result.DocumentID,
						"created": true,
					}
					if result.CreatedDocument != nil {
						if po, ok := result.CreatedDocument.(models.PurchaseOrder); ok {
							autoCreatedPO["documentNumber"] = po.DocumentNumber
							autoCreatedPO["amount"] = po.TotalAmount
						}
					}

					autoCreatedJSON, _ := datatypes.NewJSONType(autoCreatedPO).MarshalJSON()
					s.db.Model(&requisition).Updates(map[string]interface{}{
						"automation_used": true,
						"auto_created_po": datatypes.JSON(autoCreatedJSON),
					})
				}

				return nil // PO generated via workflow config, done
			}
		}

		// Fall through to legacy hardcoded config check (backward compat)
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
			"id":      result.DocumentID,
			"created": true,
		}

		if result.CreatedDocument != nil {
			if po, ok := result.CreatedDocument.(models.PurchaseOrder); ok {
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
			"automation_used":  true,
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
	case "grn", "goods_received_note":
		var grn models.GoodsReceivedNote
		if err := tx.Where("id = ?", entityID).First(&grn).Error; err == nil {
			if grn.PODocumentNumber != "" {
				var po models.PurchaseOrder
				if err := tx.Where("document_number = ?", grn.PODocumentNumber).First(&po).Error; err == nil {
					if po.Priority != "" {
						return strings.ToLower(po.Priority)
					}
				}
			}
		}
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
	case "grn", "goods_received_note":
		var grn models.GoodsReceivedNote
		if err := tx.Where("id = ?", entityID).First(&grn).Error; err == nil {
			if grn.PODocumentNumber != "" {
				var po models.PurchaseOrder
				if err := tx.Where("document_number = ?", grn.PODocumentNumber).First(&po).Error; err == nil {
					if po.RequiredByDate != nil {
						return po.RequiredByDate
					}
					if !po.DeliveryDate.IsZero() {
						return &po.DeliveryDate
					}
				}
			}
		}
	}

	return nil
}
