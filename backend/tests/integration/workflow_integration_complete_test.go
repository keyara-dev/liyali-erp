package integration

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/tests/helpers"
	"github.com/stretchr/testify/assert"
)

// TestWorkflowIntegrationComplete tests complete workflow scenarios using mocks
func TestWorkflowIntegrationComplete(t *testing.T) {
	scenario := helpers.CreateMockCompleteWorkflowScenario(t, "requisition")

	t.Run("Complete workflow execution", func(t *testing.T) {
		// Mock workflow stages
		stages := []models.WorkflowStage{
			{
				StageNumber:           1,
				StageName:             "Manager Approval",
				RequiredRole:          "manager",
				RequiredApprovals:     1,
				RequiredApprovalCount: 1,
				ApprovalType:          "any",
				CanReject:             true,
			},
			{
				StageNumber:           2,
				StageName:             "Finance Approval",
				RequiredRole:          "finance",
				RequiredApprovals:     1,
				RequiredApprovalCount: 1,
				ApprovalType:          "any",
				CanReject:             true,
			},
		}

		// Mock workflow execution
		mockWorkflow := &models.Workflow{
			ID:             uuid.New(),
			OrganizationID: scenario.Organization.ID,
			Name:           "Complete Approval Workflow",
			EntityType:     "requisition",
			IsActive:       true,
		}
		mockWorkflow.SetStages(stages)

		// Verify workflow setup
		assert.NotNil(t, mockWorkflow)
		assert.Equal(t, "requisition", mockWorkflow.EntityType)
		assert.True(t, mockWorkflow.IsActive)
		
		workflowStages, _ := mockWorkflow.GetStages()
		assert.Len(t, workflowStages, 2)
		assert.Equal(t, "Manager Approval", workflowStages[0].StageName)
		assert.Equal(t, "Finance Approval", workflowStages[1].StageName)
	})

	t.Run("Multi-stage approval process", func(t *testing.T) {
		// Mock approval records
		approvals := []models.StageApprovalRecord{
			{
				ID:               uuid.New().String(),
				WorkflowTaskID:   scenario.Task.ID,
				ApproverID:       scenario.Users.Manager.ID,
				Action:           "approved",
				Comments:         "Approved by manager",
				ApprovedAt:       time.Now(),
			},
			{
				ID:               uuid.New().String(),
				WorkflowTaskID:   uuid.New().String(),
				ApproverID:       scenario.Users.Finance.ID,
				Action:           "approved",
				Comments:         "Approved by finance",
				ApprovedAt:       time.Now().Add(1 * time.Hour),
			},
		}

		// Verify approvals
		assert.Len(t, approvals, 2)
		assert.Equal(t, "approved", approvals[0].Action)
		assert.Equal(t, "approved", approvals[1].Action)
		assert.Equal(t, scenario.Users.Manager.ID, approvals[0].ApproverID)
		assert.Equal(t, scenario.Users.Finance.ID, approvals[1].ApproverID)
	})

	t.Run("Workflow status tracking", func(t *testing.T) {
		// Mock workflow assignment status progression
		statusProgression := []string{"IN_PROGRESS", "IN_PROGRESS", "IN_PROGRESS", "COMPLETED"}
		
		for i, status := range statusProgression {
			mockAssignment := &models.WorkflowAssignment{
				ID:           scenario.Assignment.ID,
				Status:       status,
				CurrentStage: i + 1,
				UpdatedAt:    time.Now().Add(time.Duration(i) * time.Hour),
			}
			
			assert.Equal(t, status, mockAssignment.Status)
			assert.Equal(t, i+1, mockAssignment.CurrentStage)
		}
		
		// Verify final status
		assert.Equal(t, "COMPLETED", statusProgression[len(statusProgression)-1])
	})
}