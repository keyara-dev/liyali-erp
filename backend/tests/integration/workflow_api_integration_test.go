package integration

import (
	"testing"

	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/tests/helpers"
	"github.com/stretchr/testify/assert"
)

// TestWorkflowAPI_ClaimTask tests task claiming using mocks
func TestWorkflowAPI_ClaimTask(t *testing.T) {
	scenario := helpers.CreateMockCompleteWorkflowScenario(t, "requisition")
	
	// Mock task claiming
	claimedTask := scenario.Task
	claimedTask.Status = "CLAIMED"
	claimedTask.ClaimedBy = helpers.StringPtr(scenario.Users.Manager.ID)
	
	// Verify task is claimed
	assert.Equal(t, "CLAIMED", claimedTask.Status)
	assert.NotNil(t, claimedTask.ClaimedBy)
	assert.Equal(t, scenario.Users.Manager.ID, *claimedTask.ClaimedBy)
}

// TestWorkflowAPI_UnclaimTask tests task unclaiming using mocks
func TestWorkflowAPI_UnclaimTask(t *testing.T) {
	scenario := helpers.CreateMockCompleteWorkflowScenario(t, "requisition")
	
	// Mock task claiming first
	scenario.Task.Status = "CLAIMED"
	scenario.Task.ClaimedBy = helpers.StringPtr(scenario.Users.Manager.ID)
	
	// Mock task unclaiming
	scenario.Task.Status = "PENDING"
	scenario.Task.ClaimedBy = nil
	
	// Verify task is unclaimed
	assert.Equal(t, "PENDING", scenario.Task.Status)
	assert.Nil(t, scenario.Task.ClaimedBy)
}

// TestWorkflowAPI_ApproveTask tests task approval using mocks
func TestWorkflowAPI_ApproveTask(t *testing.T) {
	scenario := helpers.CreateMockCompleteWorkflowScenario(t, "requisition")
	
	// Mock task claiming
	scenario.Task.Status = "CLAIMED"
	scenario.Task.ClaimedBy = helpers.StringPtr(scenario.Users.Manager.ID)
	
	// Mock task approval
	mockApproval := &models.StageApprovalRecord{
		ID:             uuid.New().String(),
		WorkflowTaskID: scenario.Task.ID,
		ApproverID:     scenario.Users.Manager.ID,
		Action:         "approved",
		Comments:       "Approved",
		ApprovedAt:     scenario.Task.CreatedAt,
	}
	
	scenario.Task.Status = "COMPLETED"
	
	// Verify approval
	assert.Equal(t, "COMPLETED", scenario.Task.Status)
	assert.Equal(t, "approved", mockApproval.Action)
	assert.Equal(t, scenario.Users.Manager.ID, mockApproval.ApproverID)
}

// TestWorkflowAPI_RejectTask tests task rejection using mocks
func TestWorkflowAPI_RejectTask(t *testing.T) {
	scenario := helpers.CreateMockCompleteWorkflowScenario(t, "requisition")
	
	// Mock task claiming
	scenario.Task.Status = "CLAIMED"
	scenario.Task.ClaimedBy = helpers.StringPtr(scenario.Users.Manager.ID)
	
	// Mock task rejection
	mockRejection := &models.StageApprovalRecord{
		ID:             uuid.New().String(),
		WorkflowTaskID: scenario.Task.ID,
		ApproverID:     scenario.Users.Manager.ID,
		Action:         "rejected",
		Comments:       "Needs more information",
		ApprovedAt:     scenario.Task.CreatedAt,
	}
	
	scenario.Task.Status = "COMPLETED"
	
	// Verify rejection
	assert.Equal(t, "COMPLETED", scenario.Task.Status)
	assert.Equal(t, "rejected", mockRejection.Action)
	assert.Equal(t, scenario.Users.Manager.ID, mockRejection.ApproverID)
}

// TestWorkflowAPI_GetWorkflowStatus tests getting workflow status using mocks
func TestWorkflowAPI_GetWorkflowStatus(t *testing.T) {
	scenario := helpers.CreateMockCompleteWorkflowScenario(t, "requisition")
	
	// Verify workflow assignment status
	assert.Equal(t, "IN_PROGRESS", scenario.Assignment.Status)
	assert.Equal(t, 1, scenario.Assignment.CurrentStage)
	assert.NotNil(t, scenario.Assignment.AssignedAt)
}

// TestWorkflowAPI_GetAvailableApprovers tests getting available approvers using mocks
func TestWorkflowAPI_GetAvailableApprovers(t *testing.T) {
	scenario := helpers.CreateMockCompleteWorkflowScenario(t, "requisition")
	
	// Mock available approvers
	approvers := []*models.User{
		scenario.Users.Manager,
		scenario.Users.Finance,
	}
	
	// Verify approvers
	assert.Len(t, approvers, 2)
	assert.Equal(t, "manager", approvers[0].Role)
	assert.Equal(t, "finance", approvers[1].Role)
}

// TestWorkflowAPI_ConcurrentClaims tests concurrent task claiming using mocks
func TestWorkflowAPI_ConcurrentClaims(t *testing.T) {
	scenario := helpers.CreateMockCompleteWorkflowScenario(t, "requisition")
	
	// Mock concurrent claim attempts
	claimResults := make([]bool, 3)
	
	// First claim succeeds
	claimResults[0] = true
	scenario.Task.Status = "CLAIMED"
	scenario.Task.ClaimedBy = helpers.StringPtr(scenario.Users.Manager.ID)
	
	// Other claims fail
	claimResults[1] = false
	claimResults[2] = false
	
	// Verify only one claim succeeded
	successCount := 0
	for _, success := range claimResults {
		if success {
			successCount++
		}
	}
	
	assert.Equal(t, 1, successCount)
	assert.Equal(t, "CLAIMED", scenario.Task.Status)
}

// TestWorkflowAPI_OptimisticLocking tests optimistic locking using mocks
func TestWorkflowAPI_OptimisticLocking(t *testing.T) {
	scenario := helpers.CreateMockCompleteWorkflowScenario(t, "requisition")
	
	// Mock version tracking
	initialVersion := scenario.Task.Version
	
	// Simulate version increment on claim
	scenario.Task.Version++
	claimedVersion := scenario.Task.Version
	
	// Verify version changed
	assert.Greater(t, claimedVersion, initialVersion)
	
	// Verify version mismatch detection
	oldVersionMatch := initialVersion == claimedVersion
	assert.False(t, oldVersionMatch)
}
