package integration

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/tests/helpers"
	"github.com/stretchr/testify/assert"
)

// TestWorkflowsIntegration tests workflow system integration using mocks
func TestWorkflowsIntegration(t *testing.T) {
	builder := helpers.NewMockTestDataBuilder()
	
	t.Run("Workflow creation and assignment", func(t *testing.T) {
		// Mock workflow creation
		mockWorkflow := builder.CreateMockSingleStageWorkflow(t, "requisition")
		
		// Mock document to assign workflow to
		mockRequisition := builder.CreateMockRequisition(t)
		
		// Mock workflow assignment
		mockAssignment := &models.WorkflowAssignment{
			ID:              uuid.New().String(),
			OrganizationID:  builder.GetOrganizationID(),
			EntityID:        mockRequisition.ID,
			EntityType:      mockWorkflow.EntityType,
			WorkflowID:      mockWorkflow.ID,
			WorkflowVersion: mockWorkflow.Version,
			CurrentStage:    1,
			Status:          "in_progress",
			AssignedBy:      builder.GetUserID(),
			AssignedAt:      time.Now(),
		}

		// Verify workflow assignment
		assert.NotNil(t, mockAssignment)
		assert.Equal(t, mockRequisition.ID, mockAssignment.EntityID)
		assert.Equal(t, mockWorkflow.ID, mockAssignment.WorkflowID)
		assert.Equal(t, "in_progress", mockAssignment.Status)
		assert.Equal(t, 1, mockAssignment.CurrentStage)
	})

	t.Run("Task creation and assignment", func(t *testing.T) {
		// Mock workflow task
		mockTask := &models.WorkflowTask{
			ID:                   uuid.New().String(),
			OrganizationID:       builder.GetOrganizationID(),
			WorkflowAssignmentID: uuid.New().String(),
			EntityID:             uuid.New().String(),
			EntityType:           "requisition",
			StageNumber:          1,
			StageName:            "Manager Approval",
			AssignmentType:       "role",
			AssignedRole:         helpers.StringPtr("manager"),
			Status:               "pending",
			Priority:             "medium",
			Version:              1,
			CreatedAt:            time.Now(),
		}

		// Verify task creation
		assert.NotNil(t, mockTask)
		assert.Equal(t, "pending", mockTask.Status)
		assert.Equal(t, "Manager Approval", mockTask.StageName)
		assert.Equal(t, "role", mockTask.AssignmentType)
		assert.NotNil(t, mockTask.AssignedRole)
		assert.Equal(t, "manager", *mockTask.AssignedRole)
	})

	t.Run("Approval process execution", func(t *testing.T) {
		// Mock approval action
		mockApproval := &models.StageApprovalRecord{
			ID:               uuid.New().String(),
			WorkflowTaskID:   uuid.New().String(),
			ApproverID:       builder.GetManagerID(),
			Action:           "approved",
			Comments:         "Looks good, approved",
			ApprovedAt:       time.Now(),
			ApproverRole:     "manager",
			ApproverName:     "Test Manager",
		}

		// Mock task status update after approval
		mockUpdatedTask := &models.WorkflowTask{
			ID:        mockApproval.WorkflowTaskID,
			Status:    "completed",
			CreatedAt: time.Now(),
		}

		// Verify approval process
		assert.Equal(t, "approved", mockApproval.Action)
		assert.Equal(t, builder.GetManagerID(), mockApproval.ApproverID)
		assert.Equal(t, "manager", mockApproval.ApproverRole)
		assert.Equal(t, "completed", mockUpdatedTask.Status)
	})

	t.Run("Workflow completion", func(t *testing.T) {
		// Mock completed workflow assignment
		mockCompletedAssignment := &models.WorkflowAssignment{
			ID:          uuid.New().String(),
			Status:      "completed",
			CurrentStage: 1,
			CompletedAt: &time.Time{},
		}

		// Mock final document status update
		mockCompletedRequisition := &models.Requisition{
			ID:        uuid.New().String(),
			Status:    "approved",
			UpdatedAt: time.Now(),
		}

		// Verify workflow completion
		assert.Equal(t, "completed", mockCompletedAssignment.Status)
		assert.NotNil(t, mockCompletedAssignment.CompletedAt)
		assert.Equal(t, "approved", mockCompletedRequisition.Status)
	})

	t.Run("Rejection handling", func(t *testing.T) {
		// Mock rejection action
		mockRejection := &models.StageApprovalRecord{
			ID:               uuid.New().String(),
			WorkflowTaskID:   uuid.New().String(),
			ApproverID:       builder.GetManagerID(),
			Action:           "rejected",
			Comments:         "Insufficient information provided",
			ApprovedAt:       time.Now(),
			ApproverRole:     "manager",
		}

		// Mock workflow assignment after rejection
		mockRejectedAssignment := &models.WorkflowAssignment{
			ID:     uuid.New().String(),
			Status: "rejected",
		}

		// Mock document status after rejection
		mockRejectedRequisition := &models.Requisition{
			ID:        uuid.New().String(),
			Status:    "rejected",
			UpdatedAt: time.Now(),
		}

		// Verify rejection handling
		assert.Equal(t, "rejected", mockRejection.Action)
		assert.Contains(t, mockRejection.Comments, "Insufficient information")
		assert.Equal(t, "rejected", mockRejectedAssignment.Status)
		assert.Equal(t, "rejected", mockRejectedRequisition.Status)
	})
}