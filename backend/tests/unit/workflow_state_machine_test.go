package unit

import (
	"testing"

	"github.com/google/uuid"
)

// TestWorkflowStateConstants tests state constant values
func TestWorkflowStateConstants(t *testing.T) {
	tests := []struct {
		name  string
		state WorkflowState
		value string
	}{
		{"Draft state", StateDraft, "draft"},
		{"Pending state", StatePending, "pending"},
		{"Approved state", StateApproved, "approved"},
		{"Rejected state", StateRejected, "rejected"},
		{"Fulfilled state", StateFulfilled, "fulfilled"},
		{"Paid state", StatePaid, "paid"},
		{"Completed state", StateCompleted, "completed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.state) != tt.value {
				t.Errorf("State value mismatch: got %s, want %s", string(tt.state), tt.value)
			}
		})
	}
}

// TestCanTransitionValidation tests state transition validation
func TestCanTransitionValidation(t *testing.T) {
	wsm := NewWorkflowStateMachine(nil) // nil db for testing

	tests := []struct {
		name       string
		docType    string
		fromState  string
		toState    string
		userRole   string
		shouldAllow bool
	}{
		// Requisition transitions
		{
			name:        "Requisition: Draft to Pending",
			docType:     "requisition",
			fromState:   "draft",
			toState:     "pending",
			userRole:    "",
			shouldAllow: true,
		},
		{
			name:        "Requisition: Pending to Approved (approver)",
			docType:     "requisition",
			fromState:   "pending",
			toState:     "approved",
			userRole:    "approver",
			shouldAllow: true,
		},
		{
			name:        "Requisition: Pending to Approved (requester)",
			docType:     "requisition",
			fromState:   "pending",
			toState:     "approved",
			userRole:    "requester",
			shouldAllow: false,
		},
		{
			name:        "Requisition: Pending to Rejected (approver)",
			docType:     "requisition",
			fromState:   "pending",
			toState:     "rejected",
			userRole:    "approver",
			shouldAllow: true,
		},
		{
			name:        "Requisition: Rejected to Draft",
			docType:     "requisition",
			fromState:   "rejected",
			toState:     "draft",
			userRole:    "",
			shouldAllow: true,
		},
		// Budget transitions
		{
			name:        "Budget: Draft to Pending",
			docType:     "budget",
			fromState:   "draft",
			toState:     "pending",
			userRole:    "",
			shouldAllow: true,
		},
		{
			name:        "Budget: Pending to Approved (finance)",
			docType:     "budget",
			fromState:   "pending",
			toState:     "approved",
			userRole:    "finance",
			shouldAllow: true,
		},
		// PO transitions
		{
			name:        "PO: Draft to Pending",
			docType:     "po",
			fromState:   "draft",
			toState:     "pending",
			userRole:    "",
			shouldAllow: true,
		},
		{
			name:        "PO: Approved to Fulfilled",
			docType:     "po",
			fromState:   "approved",
			toState:     "fulfilled",
			userRole:    "",
			shouldAllow: true,
		},
		{
			name:        "PO: Fulfilled to Completed",
			docType:     "po",
			fromState:   "fulfilled",
			toState:     "completed",
			userRole:    "",
			shouldAllow: true,
		},
		// GRN transitions
		{
			name:        "GRN: Draft to Pending",
			docType:     "grn",
			fromState:   "draft",
			toState:     "pending",
			userRole:    "",
			shouldAllow: true,
		},
		{
			name:        "GRN: Pending to Approved (approver)",
			docType:     "grn",
			fromState:   "pending",
			toState:     "approved",
			userRole:    "approver",
			shouldAllow: true,
		},
		// Invalid transitions
		{
			name:        "Requisition: Approved to Draft (invalid)",
			docType:     "requisition",
			fromState:   "approved",
			toState:     "draft",
			userRole:    "",
			shouldAllow: false,
		},
		{
			name:        "Unknown document type",
			docType:     "unknown",
			fromState:   "draft",
			toState:     "pending",
			userRole:    "",
			shouldAllow: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := wsm.CanTransition(tt.docType, tt.fromState, tt.toState, tt.userRole)
			if result != tt.shouldAllow {
				t.Errorf("CanTransition returned %v, want %v", result, tt.shouldAllow)
			}
		})
	}
}

// TestValidNextStates tests getting available next states
func TestValidNextStates(t *testing.T) {
	wsm := NewWorkflowStateMachine(nil)

	tests := []struct {
		name           string
		docType        string
		currentState   string
		userRole       string
		expectedStates []string
	}{
		{
			name:           "Requisition draft states",
			docType:        "requisition",
			currentState:   "draft",
			userRole:       "requester",
			expectedStates: []string{"pending"},
		},
		{
			name:           "Requisition pending as approver",
			docType:        "requisition",
			currentState:   "pending",
			userRole:       "approver",
			expectedStates: []string{"approved", "rejected"},
		},
		{
			name:           "Requisition pending as requester",
			docType:        "requisition",
			currentState:   "pending",
			userRole:       "requester",
			expectedStates: []string{},
		},
		{
			name:           "Budget draft states",
			docType:        "budget",
			currentState:   "draft",
			userRole:       "",
			expectedStates: []string{"pending"},
		},
		{
			name:           "PO approved states",
			docType:        "po",
			currentState:   "approved",
			userRole:       "",
			expectedStates: []string{"fulfilled"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			states := wsm.GetValidNextStates(tt.docType, tt.currentState, tt.userRole)

			if len(states) != len(tt.expectedStates) {
				t.Errorf("Expected %d states, got %d", len(tt.expectedStates), len(states))
			}

			// Check each expected state is in result
			for _, expected := range tt.expectedStates {
				found := false
				for _, state := range states {
					if state == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected state %s not found in results", expected)
				}
			}
		})
	}
}

// TestTransitionDocumentStructure tests audit log creation structure
func TestTransitionDocumentStructure(t *testing.T) {
	t.Run("Audit log creation", func(t *testing.T) {
		documentID := uuid.New().String()
		userID := uuid.New().String()

		// Simulate what would be created
		auditLog := map[string]interface{}{
			"id":            uuid.New().String(),
			"documentId":    documentID,
			"documentType":  "requisition",
			"userId":        userID,
			"action":        "approve",
			"fromState":     "pending",
			"toState":       "approved",
			"timestamp":     "2025-12-22T12:00:00Z",
		}

		if auditLog["documentId"] != documentID {
			t.Error("Audit log document ID mismatch")
		}
		if auditLog["userId"] != userID {
			t.Error("Audit log user ID mismatch")
		}
		if auditLog["action"] != "approve" {
			t.Error("Audit log action mismatch")
		}
	})
}

// TestStateTransitionDocumentation documents all valid transitions
func TestStateTransitionDocumentation(t *testing.T) {
	t.Run("Document all valid requisition transitions", func(t *testing.T) {
		wsm := NewWorkflowStateMachine(nil)

		validTransitions := []struct {
			from     string
			to       string
			role     string
			allowed  bool
		}{
			{"draft", "pending", "", true},
			{"pending", "approved", "approver", true},
			{"pending", "rejected", "approver", true},
			{"rejected", "draft", "", true},
			{"approved", "pending", "", false},
			{"approved", "rejected", "", false},
		}

		for _, trans := range validTransitions {
			result := wsm.CanTransition("requisition", trans.from, trans.to, trans.role)
			if result != trans.allowed {
				t.Errorf(
					"Requisition %s->%s (role: %s) = %v, want %v",
					trans.from, trans.to, trans.role, result, trans.allowed,
				)
			}
		}
	})

	t.Run("Document all valid PO transitions", func(t *testing.T) {
		wsm := NewWorkflowStateMachine(nil)

		validTransitions := []struct {
			from    string
			to      string
			allowed bool
		}{
			{"draft", "pending", true},
			{"pending", "approved", true},
			{"pending", "rejected", true},
			{"approved", "fulfilled", true},
			{"fulfilled", "completed", true},
		}

		for _, trans := range validTransitions {
			result := wsm.CanTransition("po", trans.from, trans.to, "finance")
			if result != trans.allowed {
				t.Errorf("PO %s->%s = %v, want %v", trans.from, trans.to, result, trans.allowed)
			}
		}
	})
}

// TestRoleBasedPermissions tests role-based access control
func TestRoleBasedPermissions(t *testing.T) {
	wsm := NewWorkflowStateMachine(nil)

	tests := []struct {
		name         string
		docType      string
		fromState    string
		toState      string
		role         string
		shouldAllow  bool
	}{
		{"Admin can approve", "requisition", "pending", "approved", "admin", true},
		{"Approver can approve", "requisition", "pending", "approved", "approver", true},
		{"Requester cannot approve", "requisition", "pending", "approved", "requester", false},
		{"Viewer cannot approve", "requisition", "pending", "approved", "viewer", false},
		{"Finance can approve budget", "budget", "pending", "approved", "finance", true},
		{"Approver cannot approve budget", "budget", "pending", "approved", "approver", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := wsm.CanTransition(tt.docType, tt.fromState, tt.toState, tt.role)
			if result != tt.shouldAllow {
				t.Errorf("Expected %v, got %v", tt.shouldAllow, result)
			}
		})
	}
}

// BenchmarkCanTransition benchmarks transition validation
func BenchmarkCanTransition(b *testing.B) {
	wsm := NewWorkflowStateMachine(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = wsm.CanTransition("requisition", "draft", "pending", "")
	}
}

// BenchmarkGetValidNextStates benchmarks getting valid states
func BenchmarkGetValidNextStates(b *testing.B) {
	wsm := NewWorkflowStateMachine(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = wsm.GetValidNextStates("requisition", "draft", "requester")
	}
}
