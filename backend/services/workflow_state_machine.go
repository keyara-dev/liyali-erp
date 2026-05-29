package services

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// WorkflowState represents valid document states
type WorkflowState string

const (
	StateDraft     WorkflowState = "DRAFT"
	StatePending   WorkflowState = "PENDING"
	StateApproved  WorkflowState = "APPROVED"
	StateRejected  WorkflowState = "REJECTED"
	StateRevision  WorkflowState = "REVISION"
	StateCancelled WorkflowState = "CANCELLED"
	StateFulfilled WorkflowState = "FULFILLED" // For PO
	StatePaid      WorkflowState = "PAID"      // For PV
	StateCompleted WorkflowState = "COMPLETED" // For GRN / PO terminal
)

// WorkflowTransition defines valid state transitions
type WorkflowTransition struct {
	From    WorkflowState
	To      WorkflowState
	Action  string
	RequiredRole string // Empty string means any authenticated user
}

// WorkflowStateMachine manages document state transitions
type WorkflowStateMachine struct {
	db          *gorm.DB
	transitions map[string][]WorkflowTransition
}

// NewWorkflowStateMachine creates a new workflow state machine
func NewWorkflowStateMachine(db *gorm.DB) *WorkflowStateMachine {
	wsm := &WorkflowStateMachine{
		db:          db,
		transitions: make(map[string][]WorkflowTransition),
	}
	wsm.initializeTransitions()
	return wsm
}

// initializeTransitions sets up all valid transitions for document types
func (wsm *WorkflowStateMachine) initializeTransitions() {
	// Requisition transitions
	wsm.transitions["requisition"] = []WorkflowTransition{
		{From: StateDraft, To: StatePending, Action: "submit", RequiredRole: ""},
		{From: StateDraft, To: StateApproved, Action: "auto_approve", RequiredRole: "system"},
		{From: StatePending, To: StateApproved, Action: "approve", RequiredRole: "approver"},
		{From: StatePending, To: StateRejected, Action: "reject", RequiredRole: "approver"},
		{From: StateRejected, To: StateDraft, Action: "reopen", RequiredRole: "requester"},
		{From: StateDraft, To: "deleted", Action: "delete", RequiredRole: "requester"},
	}

	// Budget transitions
	wsm.transitions["budget"] = []WorkflowTransition{
		{From: StateDraft, To: StatePending, Action: "submit", RequiredRole: ""},
		{From: StatePending, To: StateApproved, Action: "approve", RequiredRole: "finance"},
		{From: StatePending, To: StateRejected, Action: "reject", RequiredRole: "finance"},
		{From: StateRejected, To: StateDraft, Action: "reopen", RequiredRole: ""},
		{From: StateDraft, To: "deleted", Action: "delete", RequiredRole: ""},
	}

	// Purchase Order transitions
	wsm.transitions["po"] = []WorkflowTransition{
		{From: StateDraft, To: StatePending, Action: "submit", RequiredRole: ""},
		{From: StatePending, To: StateApproved, Action: "approve", RequiredRole: "finance"},
		{From: StatePending, To: StateRejected, Action: "reject", RequiredRole: "finance"},
		{From: StateApproved, To: StateFulfilled, Action: "fulfill", RequiredRole: ""},
		{From: StateFulfilled, To: StateCompleted, Action: "complete", RequiredRole: ""},
		{From: StateDraft, To: "deleted", Action: "delete", RequiredRole: ""},
	}

	// Payment Voucher transitions
	wsm.transitions["pv"] = []WorkflowTransition{
		{From: StateDraft, To: StatePending, Action: "submit", RequiredRole: ""},
		{From: StatePending, To: StateApproved, Action: "approve", RequiredRole: "finance"},
		{From: StatePending, To: StateRejected, Action: "reject", RequiredRole: "finance"},
		{From: StateApproved, To: StatePaid, Action: "pay", RequiredRole: "finance"},
		{From: StateDraft, To: "deleted", Action: "delete", RequiredRole: ""},
	}

	// GRN transitions
	wsm.transitions["grn"] = []WorkflowTransition{
		{From: StateDraft, To: StatePending, Action: "submit", RequiredRole: ""},
		{From: StatePending, To: StateApproved, Action: "approve", RequiredRole: "approver"},
		{From: StatePending, To: StateRejected, Action: "reject", RequiredRole: "approver"},
		{From: StatePending, To: StateRevision, Action: "return_for_revision", RequiredRole: "approver"},
		// Workflow now auto-advances APPROVED → COMPLETED in the same step
		// (see workflow_execution_service GRN auto-complete), so the explicit
		// "complete" transition is kept for backfill of older GRNs.
		{From: StateApproved, To: StateCompleted, Action: "complete", RequiredRole: ""},
		// Revision cycles back to DRAFT after the user re-submits the form.
		{From: StateRevision, To: StateDraft, Action: "resubmit", RequiredRole: ""},
		// DRAFT or REVISION GRNs can be cancelled by the creator.
		{From: StateDraft, To: StateCancelled, Action: "cancel", RequiredRole: ""},
		{From: StateRevision, To: StateCancelled, Action: "cancel", RequiredRole: ""},
		{From: StateDraft, To: "deleted", Action: "delete", RequiredRole: ""},
	}

	// Cross-document cancellation transitions for REQ/PO/PV symmetry.
	for _, k := range []string{"requisition", "po", "pv"} {
		wsm.transitions[k] = append(wsm.transitions[k],
			WorkflowTransition{From: StateDraft, To: StateCancelled, Action: "cancel", RequiredRole: ""},
			WorkflowTransition{From: StateRevision, To: StateCancelled, Action: "cancel", RequiredRole: ""},
		)
	}
}

// CanTransition checks if a state transition is allowed
func (wsm *WorkflowStateMachine) CanTransition(docType string, fromState, toState, userRole string) bool {
	transitions, exists := wsm.transitions[docType]
	if !exists {
		log.Printf("Unknown document type: %s", docType)
		return false
	}

	for _, t := range transitions {
		if t.From == WorkflowState(fromState) && t.To == WorkflowState(toState) {
			// Check role requirement
			if t.RequiredRole == "" || t.RequiredRole == userRole {
				return true
			}
		}
	}

	return false
}

// TransitionDocument moves a document from one state to another
func (wsm *WorkflowStateMachine) TransitionDocument(
	docType string,
	documentID string,
	fromState, toState WorkflowState,
	userID, userRole, action, comments string,
) error {
	// Validate transition
	if !wsm.CanTransition(docType, string(fromState), string(toState), userRole) {
		return fmt.Errorf("invalid state transition from %s to %s for %s", fromState, toState, docType)
	}

	// Update document status based on type
	var result *gorm.DB
	switch docType {
	case "requisition":
		result = wsm.db.Model(&models.Requisition{}).
			Where("id = ?", documentID).
			Update("status", string(toState))
	case "budget":
		result = wsm.db.Model(&models.Budget{}).
			Where("id = ?", documentID).
			Update("status", string(toState))
	case "po":
		result = wsm.db.Model(&models.PurchaseOrder{}).
			Where("id = ?", documentID).
			Update("status", string(toState))
	case "pv":
		result = wsm.db.Model(&models.PaymentVoucher{}).
			Where("id = ?", documentID).
			Update("status", string(toState))
	case "grn":
		grnUpdates := map[string]interface{}{"status": string(toState)}
		// Keep signoff_status in sync once the workflow terminates the GRN.
		// PENDING_RECEIVER / PENDING_CERTIFIER / READY only describe the
		// pre-workflow form sign-off lifecycle.
		upper := strings.ToUpper(string(toState))
		if upper == "APPROVED" || upper == "COMPLETED" || upper == "REJECTED" || upper == "CANCELLED" {
			grnUpdates["signoff_status"] = "COMPLETED"
		}
		result = wsm.db.Model(&models.GoodsReceivedNote{}).
			Where("id = ?", documentID).
			Updates(grnUpdates)
	default:
		return fmt.Errorf("unknown document type: %s", docType)
	}

	if result.Error != nil {
		return result.Error
	}

	// Create audit log entry
	// Create changes map
	changes := map[string]interface{}{
		"from":    fromState,
		"to":      toState,
		"comment": comments,
	}
	
	auditLog := models.AuditLog{
		ID:           uuid.New().String(),
		DocumentID:   documentID,
		DocumentType: docType,
		UserID:       userID,
		Action:       action,
		Changes:      datatypes.JSONType[map[string]interface{}]{},
		CreatedAt:    time.Now(),
	}
	
	// Set the changes data
	auditLog.Changes = datatypes.JSONType[map[string]interface{}]{}
	auditLog.Changes.Scan(changes)

	if err := wsm.db.Create(&auditLog).Error; err != nil {
		log.Printf("Error creating audit log: %v", err)
		// Don't fail the transition if audit logging fails
	}

	return nil
}

// GetValidNextStates returns all valid next states from current state
func (wsm *WorkflowStateMachine) GetValidNextStates(docType, currentState, userRole string) []string {
	transitions, exists := wsm.transitions[docType]
	if !exists {
		return []string{}
	}

	var validStates []string
	for _, t := range transitions {
		if t.From == WorkflowState(currentState) {
			// Check role requirement
			if t.RequiredRole == "" || t.RequiredRole == userRole {
				validStates = append(validStates, string(t.To))
			}
		}
	}

	return validStates
}

// SubmitForApproval moves a document from draft to pending
func (wsm *WorkflowStateMachine) SubmitForApproval(docType, documentID, userID string) error {
	return wsm.TransitionDocument(
		docType,
		documentID,
		StateDraft,
		StatePending,
		userID,
		"",
		"submit",
		"Document submitted for approval",
	)
}

// ApproveDocument moves a document from pending to approved
func (wsm *WorkflowStateMachine) ApproveDocument(
	docType, documentID, userID, approverRole, comments string,
) error {
	return wsm.TransitionDocument(
		docType,
		documentID,
		StatePending,
		StateApproved,
		userID,
		approverRole,
		"approve",
		comments,
	)
}

// RejectDocument moves a document from pending to rejected
func (wsm *WorkflowStateMachine) RejectDocument(
	docType, documentID, userID, approverRole, remarks string,
) error {
	return wsm.TransitionDocument(
		docType,
		documentID,
		StatePending,
		StateRejected,
		userID,
		approverRole,
		"reject",
		remarks,
	)
}

// GetWorkflowHistory returns all state transitions for a document
func (wsm *WorkflowStateMachine) GetWorkflowHistory(documentID string) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	if err := wsm.db.Where("document_id = ?", documentID).
		Order("created_at ASC").
		Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}
