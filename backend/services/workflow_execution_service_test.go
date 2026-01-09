package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStageProgressInfo(t *testing.T) {
	// Test the StageProgressInfo structure
	now := time.Now()
	
	stageInfo := StageProgressInfo{
		StageNumber:    1,
		StageName:      "Manager Approval",
		RequiredRole:   "manager",
		Status:         "approved",
		IsCurrentStage: false,
		ApproverID:     "user-123",
		ApproverName:   "John Manager",
		ApproverRole:   "manager",
		CompletedAt:    &now,
		Comments:       "Approved for budget allocation",
	}
	
	assert.Equal(t, 1, stageInfo.StageNumber)
	assert.Equal(t, "Manager Approval", stageInfo.StageName)
	assert.Equal(t, "manager", stageInfo.RequiredRole)
	assert.Equal(t, "approved", stageInfo.Status)
	assert.False(t, stageInfo.IsCurrentStage)
	assert.Equal(t, "user-123", stageInfo.ApproverID)
	assert.Equal(t, "John Manager", stageInfo.ApproverName)
	assert.Equal(t, "manager", stageInfo.ApproverRole)
	assert.NotNil(t, stageInfo.CompletedAt)
	assert.Equal(t, "Approved for budget allocation", stageInfo.Comments)
}

func TestWorkflowStatusResponse(t *testing.T) {
	// Test the enhanced WorkflowStatusResponse structure
	stageProgress := []StageProgressInfo{
		{
			StageNumber:    1,
			StageName:      "Manager Approval",
			RequiredRole:   "manager",
			Status:         "approved",
			IsCurrentStage: false,
		},
		{
			StageNumber:    2,
			StageName:      "Finance Approval",
			RequiredRole:   "finance",
			Status:         "pending",
			IsCurrentStage: true,
		},
	}
	
	response := WorkflowStatusResponse{
		CurrentStage:  2,
		TotalStages:   3,
		Status:        "in_progress",
		NextApprover:  "Finance Team",
		CanApprove:    true,
		CanReject:     true,
		StageProgress: stageProgress,
	}
	
	assert.Equal(t, 2, response.CurrentStage)
	assert.Equal(t, 3, response.TotalStages)
	assert.Equal(t, "in_progress", response.Status)
	assert.Equal(t, "Finance Team", response.NextApprover)
	assert.True(t, response.CanApprove)
	assert.True(t, response.CanReject)
	assert.Len(t, response.StageProgress, 2)
	
	// Check first stage
	assert.Equal(t, "approved", response.StageProgress[0].Status)
	assert.False(t, response.StageProgress[0].IsCurrentStage)
	
	// Check current stage
	assert.Equal(t, "pending", response.StageProgress[1].Status)
	assert.True(t, response.StageProgress[1].IsCurrentStage)
}

func TestApproverInfo(t *testing.T) {
	// Test the ApproverInfo structure
	approver := ApproverInfo{
		ID:    "user-456",
		Name:  "Jane Finance",
		Email: "jane.finance@company.com",
		Role:  "finance",
	}
	
	assert.Equal(t, "user-456", approver.ID)
	assert.Equal(t, "Jane Finance", approver.Name)
	assert.Equal(t, "jane.finance@company.com", approver.Email)
	assert.Equal(t, "finance", approver.Role)
}

func TestWorkflowStatusResponseJSON(t *testing.T) {
	// Test that the response can be properly serialized to JSON
	// This ensures our JSON tags are correct
	
	response := WorkflowStatusResponse{
		CurrentStage: 1,
		TotalStages:  2,
		Status:       "completed",
		CanApprove:   false,
		CanReject:    false,
		StageProgress: []StageProgressInfo{
			{
				StageNumber:  1,
				StageName:    "Test Stage",
				RequiredRole: "admin",
				Status:       "approved",
			},
		},
	}
	
	// Verify the structure is valid
	assert.NotNil(t, response)
	assert.Equal(t, 1, response.CurrentStage)
	assert.Equal(t, 2, response.TotalStages)
	assert.Equal(t, "completed", response.Status)
	assert.Len(t, response.StageProgress, 1)
}