package unit

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/services"
	"github.com/liyali/liyali-gateway/tests/helpers"
	"github.com/stretchr/testify/assert"
)

// TestWorkflowConcurrencyFixes tests that our fixes resolve the concurrency issues
func TestWorkflowConcurrencyFixes(t *testing.T) {
	// Setup test database
	testDB := helpers.SetupTestDB(t)
	defer testDB.Cleanup()
	db := testDB.DB

	// Setup test data
	orgID := uuid.New().String()
	org := &models.Organization{
		ID:   orgID,
		Name: "Test Organization",
	}
	assert.NoError(t, db.Create(org).Error)

	// Create workflow with enhanced features
	workflowID := uuid.New()
	workflow := &models.Workflow{
		ID:             workflowID,
		OrganizationID: orgID,
		Name:           "Enhanced Test Workflow",
		EntityType:     "requisition",
		IsActive:       true,
		IsDefault:      true,
		Version:        1,
	}

	// Set workflow stages with multiple approval support
	stages := []models.WorkflowStage{
		{
			StageNumber:           1,
			StageName:             "Manager Approval",
			RequiredRole:          "manager",
			RequiredApprovalCount: 2, // Require 2 approvals
			ApprovalType:          "quorum",
			QuorumCount:           func(i int) *int { return &i }(2),
		},
	}
	assert.NoError(t, workflow.SetStages(stages))
	assert.NoError(t, db.Create(workflow).Error)

	// Create multiple users with the same role
	user1ID := uuid.New().String()
	user2ID := uuid.New().String()
	user3ID := uuid.New().String()

	users := []*models.User{
		{
			ID:                    user1ID,
			CurrentOrganizationID: &orgID,
			Email:                 "manager1@test.com",
			Name:                  "Manager 1",
			Role:                  "manager",
			Active:                true,
		},
		{
			ID:                    user2ID,
			CurrentOrganizationID: &orgID,
			Email:                 "manager2@test.com",
			Name:                  "Manager 2",
			Role:                  "manager",
			Active:                true,
		},
		{
			ID:                    user3ID,
			CurrentOrganizationID: &orgID,
			Email:                 "manager3@test.com",
			Name:                  "Manager 3",
			Role:                  "manager",
			Active:                true,
		},
	}

	for _, user := range users {
		assert.NoError(t, db.Create(user).Error)
	}

	workflowService := services.NewWorkflowService(nil, nil, db)
	workflowExecutionService := services.NewWorkflowExecutionService(db, workflowService, nil, nil)

	t.Run("FIXED: Task claiming prevents concurrent actions", func(t *testing.T) {
		// Create a document and assign workflow
		documentID := uuid.New().String()
		
		assignment, err := workflowExecutionService.AssignWorkflowToDocument(
			context.Background(),
			orgID,
			documentID,
			"requisition",
			user1ID,
		)
		assert.NoError(t, err)
		assert.NotNil(t, assignment)

		// Get pending tasks
		pendingTasks, err := workflowExecutionService.GetPendingWorkflowTasks(
			context.Background(),
			orgID,
			documentID,
		)
		assert.NoError(t, err)
		assert.Len(t, pendingTasks, 1)

		taskID := pendingTasks[0].ID

		// User 1 claims the task
		err = workflowExecutionService.ClaimWorkflowTask(context.Background(), taskID, user1ID)
		assert.NoError(t, err, "User 1 should be able to claim the task")

		// User 2 tries to claim the same task - should fail
		err = workflowExecutionService.ClaimWorkflowTask(context.Background(), taskID, user2ID)
		assert.Error(t, err, "User 2 should not be able to claim already claimed task")
		assert.Contains(t, err.Error(), "not available for claiming")

		// User 2 tries to approve without claiming - should fail
		err = workflowExecutionService.ApproveWorkflowTask(
			context.Background(),
			taskID,
			user2ID,
			"signature-user2",
			"Trying to approve without claim",
		)
		assert.Error(t, err, "User 2 should not be able to approve task claimed by User 1")
		assert.Contains(t, err.Error(), "claimed by another user")

		// User 1 can approve since they claimed it
		err = workflowExecutionService.ApproveWorkflowTask(
			context.Background(),
			taskID,
			user1ID,
			"signature-user1",
			"Approved by User 1",
		)
		assert.NoError(t, err, "User 1 should be able to approve their claimed task")

		t.Logf("SUCCESS: Task claiming prevents concurrent actions")
	})

	t.Run("FIXED: Multiple approvals required for stage completion", func(t *testing.T) {
		// Create a new document for this test
		documentID := uuid.New().String()
		
		assignment, err := workflowExecutionService.AssignWorkflowToDocument(
			context.Background(),
			orgID,
			documentID,
			"requisition",
			user1ID,
		)
		assert.NoError(t, err)
		assert.NotNil(t, assignment)

		pendingTasks, err := workflowExecutionService.GetPendingWorkflowTasks(
			context.Background(),
			orgID,
			documentID,
		)
		assert.NoError(t, err)
		assert.Len(t, pendingTasks, 1)

		taskID := pendingTasks[0].ID

		// First approval - should not complete the stage (requires 2)
		err = workflowExecutionService.ClaimWorkflowTask(context.Background(), taskID, user1ID)
		assert.NoError(t, err)

		err = workflowExecutionService.ApproveWorkflowTask(
			context.Background(),
			taskID,
			user1ID,
			"signature-user1",
			"First approval",
		)
		assert.NoError(t, err)

		// Check task status - should be partially approved
		var task models.WorkflowTask
		assert.NoError(t, db.Where("id = ?", taskID).First(&task).Error)
		assert.Equal(t, "partially_approved", task.Status, "Task should be partially approved after first approval")

		// Check stage approval records
		var approvals []models.StageApprovalRecord
		assert.NoError(t, db.Where("workflow_task_id = ?", taskID).Find(&approvals).Error)
		assert.Len(t, approvals, 1, "Should have one approval record")
		assert.Equal(t, "approved", approvals[0].Action)
		assert.Equal(t, user1ID, approvals[0].ApproverID)

		// Unclaim the task so another user can claim it
		err = workflowExecutionService.UnclaimWorkflowTask(context.Background(), taskID, user1ID)
		assert.NoError(t, err)

		// Second approval - should complete the stage
		err = workflowExecutionService.ClaimWorkflowTask(context.Background(), taskID, user2ID)
		assert.NoError(t, err)

		err = workflowExecutionService.ApproveWorkflowTask(
			context.Background(),
			taskID,
			user2ID,
			"signature-user2",
			"Second approval",
		)
		assert.NoError(t, err)

		// Check task status - should be completed now
		assert.NoError(t, db.Where("id = ?", taskID).First(&task).Error)
		assert.Equal(t, "completed", task.Status, "Task should be completed after second approval")

		// Check stage approval records
		assert.NoError(t, db.Where("workflow_task_id = ?", taskID).Find(&approvals).Error)
		assert.Len(t, approvals, 2, "Should have two approval records")

		t.Logf("SUCCESS: Multiple approvals required for stage completion")
	})

	t.Run("FIXED: Optimistic locking prevents race conditions", func(t *testing.T) {
		// Create a new document for this test
		documentID := uuid.New().String()
		
		assignment, err := workflowExecutionService.AssignWorkflowToDocument(
			context.Background(),
			orgID,
			documentID,
			"requisition",
			user1ID,
		)
		assert.NoError(t, err)
		assert.NotNil(t, assignment)

		pendingTasks, err := workflowExecutionService.GetPendingWorkflowTasks(
			context.Background(),
			orgID,
			documentID,
		)
		assert.NoError(t, err)
		assert.Len(t, pendingTasks, 1)

		taskID := pendingTasks[0].ID
		initialVersion := pendingTasks[0].Version

		// User 1 claims the task
		err = workflowExecutionService.ClaimWorkflowTask(context.Background(), taskID, user1ID)
		assert.NoError(t, err)

		// Get updated task version after claiming
		var task models.WorkflowTask
		assert.NoError(t, db.Where("id = ?", taskID).First(&task).Error)
		claimedVersion := task.Version

		assert.Greater(t, claimedVersion, initialVersion, "Version should increment after claiming")

		// Try to approve with old version - should fail
		err = workflowExecutionService.ApproveWorkflowTaskWithVersion(
			context.Background(),
			taskID,
			user1ID,
			"signature-user1",
			"Approval with old version",
			initialVersion, // Using old version
		)
		assert.Error(t, err, "Approval with old version should fail")
		assert.Contains(t, err.Error(), "expected version")

		// Approve with correct version - should succeed
		err = workflowExecutionService.ApproveWorkflowTaskWithVersion(
			context.Background(),
			taskID,
			user1ID,
			"signature-user1",
			"Approval with correct version",
			claimedVersion, // Using correct version
		)
		assert.NoError(t, err, "Approval with correct version should succeed")

		t.Logf("SUCCESS: Optimistic locking prevents race conditions")
	})

	t.Run("FIXED: Clear error messages for concurrent access", func(t *testing.T) {
		// Create a new document for this test
		documentID := uuid.New().String()
		
		assignment, err := workflowExecutionService.AssignWorkflowToDocument(
			context.Background(),
			orgID,
			documentID,
			"requisition",
			user1ID,
		)
		assert.NoError(t, err)
		assert.NotNil(t, assignment)

		pendingTasks, err := workflowExecutionService.GetPendingWorkflowTasks(
			context.Background(),
			orgID,
			documentID,
		)
		assert.NoError(t, err)
		assert.Len(t, pendingTasks, 1)

		taskID := pendingTasks[0].ID

		// User 1 claims the task
		err = workflowExecutionService.ClaimWorkflowTask(context.Background(), taskID, user1ID)
		assert.NoError(t, err)

		// User 2 tries to claim - should get clear error message
		err = workflowExecutionService.ClaimWorkflowTask(context.Background(), taskID, user2ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not available for claiming")

		// User 2 tries to approve - should get clear error message
		err = workflowExecutionService.ApproveWorkflowTask(
			context.Background(),
			taskID,
			user2ID,
			"signature-user2",
			"Trying to approve claimed task",
		)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "claimed by another user")

		// Test claim expiry
		// Manually set claim expiry to past time
		db.Model(&models.WorkflowTask{}).Where("id = ?", taskID).Update("claim_expiry", time.Now().Add(-1*time.Hour))

		// User 1 tries to approve with expired claim - should get clear error message
		err = workflowExecutionService.ApproveWorkflowTask(
			context.Background(),
			taskID,
			user1ID,
			"signature-user1",
			"Trying to approve with expired claim",
		)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "claim has expired")

		t.Logf("SUCCESS: Clear error messages for concurrent access")
	})

	t.Run("FIXED: Concurrent approval attempts are handled gracefully", func(t *testing.T) {
		// Create a new document for this test
		documentID := uuid.New().String()
		
		assignment, err := workflowExecutionService.AssignWorkflowToDocument(
			context.Background(),
			orgID,
			documentID,
			"requisition",
			user1ID,
		)
		assert.NoError(t, err)
		assert.NotNil(t, assignment)

		pendingTasks, err := workflowExecutionService.GetPendingWorkflowTasks(
			context.Background(),
			orgID,
			documentID,
		)
		assert.NoError(t, err)
		assert.Len(t, pendingTasks, 1)

		taskID := pendingTasks[0].ID

		// Simulate concurrent claim attempts
		var wg sync.WaitGroup
		var claimResults []error
		var mu sync.Mutex

		// Three users try to claim simultaneously
		for i, userID := range []string{user1ID, user2ID, user3ID} {
			wg.Add(1)
			go func(uid string, index int) {
				defer wg.Done()
				
				// Add small delay to increase chance of race condition
				time.Sleep(time.Duration(index) * 5 * time.Millisecond)
				
				err := workflowExecutionService.ClaimWorkflowTask(context.Background(), taskID, uid)
				
				mu.Lock()
				claimResults = append(claimResults, err)
				mu.Unlock()
			}(userID, i)
		}

		wg.Wait()

		// Analyze results - exactly one should succeed
		successCount := 0
		errorCount := 0
		for i, err := range claimResults {
			if err == nil {
				successCount++
				t.Logf("User %d claim: SUCCESS", i+1)
			} else {
				errorCount++
				t.Logf("User %d claim: ERROR - %v", i+1, err)
			}
		}

		assert.Equal(t, 1, successCount, "Exactly one claim should succeed")
		assert.Equal(t, 2, errorCount, "Two claims should fail")

		t.Logf("SUCCESS: Concurrent claim attempts handled gracefully - %d succeeded, %d failed", successCount, errorCount)
	})
}

// TestMultipleApprovalTypes tests different approval type configurations
func TestMultipleApprovalTypes(t *testing.T) {
	// Setup test database
	testDB := helpers.SetupTestDB(t)
	defer testDB.Cleanup()
	db := testDB.DB

	// Setup test data
	orgID := uuid.New().String()
	org := &models.Organization{
		ID:   orgID,
		Name: "Test Organization",
	}
	assert.NoError(t, db.Create(org).Error)

	// Create 5 users with manager role
	userIDs := make([]string, 5)
	for i := 0; i < 5; i++ {
		userID := uuid.New().String()
		userIDs[i] = userID
		user := &models.User{
			ID:                    userID,
			CurrentOrganizationID: &orgID,
			Email:                 fmt.Sprintf("manager%d@test.com", i+1),
			Name:                  fmt.Sprintf("Manager %d", i+1),
			Role:                  "manager",
			Active:                true,
		}
		assert.NoError(t, db.Create(user).Error)
	}

	workflowService := services.NewWorkflowService(nil, nil, db)
	workflowExecutionService := services.NewWorkflowExecutionService(db, workflowService, nil, nil)

	testCases := []struct {
		name          string
		approvalType  string
		requiredCount int
		quorumCount   *int
		expectSuccess []int // Which approval attempts should succeed (1-based)
	}{
		{
			name:          "Any approval (first one completes)",
			approvalType:  "any",
			requiredCount: 1,
			expectSuccess: []int{1}, // Only first approval needed
		},
		{
			name:          "Majority approval (3 out of 5)",
			approvalType:  "majority",
			requiredCount: 3,
			expectSuccess: []int{1, 2, 3}, // Need 3 approvals
		},
		{
			name:          "Quorum approval (2 required)",
			approvalType:  "quorum",
			requiredCount: 2,
			quorumCount:   func(i int) *int { return &i }(2),
			expectSuccess: []int{1, 2}, // Need 2 approvals
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create workflow with specific approval type
			workflowID := uuid.New()
			workflow := &models.Workflow{
				ID:             workflowID,
				OrganizationID: orgID,
				Name:           fmt.Sprintf("Test Workflow - %s", tc.name),
				EntityType:     "requisition",
				IsActive:       true,
				Version:        1,
			}

			stages := []models.WorkflowStage{
				{
					StageNumber:           1,
					StageName:             "Manager Approval",
					RequiredRole:          "manager",
					RequiredApprovalCount: tc.requiredCount,
					ApprovalType:          tc.approvalType,
					QuorumCount:           tc.quorumCount,
				},
			}
			assert.NoError(t, workflow.SetStages(stages))
			assert.NoError(t, db.Create(workflow).Error)

			// Create document and assign workflow
			documentID := uuid.New().String()
			assignment, err := workflowExecutionService.AssignWorkflowToDocument(
				context.Background(),
				orgID,
				documentID,
				"requisition",
				userIDs[0],
			)
			assert.NoError(t, err)
			assert.NotNil(t, assignment)

			pendingTasks, err := workflowExecutionService.GetPendingWorkflowTasks(
				context.Background(),
				orgID,
				documentID,
			)
			assert.NoError(t, err)
			assert.Len(t, pendingTasks, 1)

			taskID := pendingTasks[0].ID

			// Process approvals according to test case
			for i, userID := range userIDs {
				approvalIndex := i + 1
				
				// Claim and approve
				err = workflowExecutionService.ClaimWorkflowTask(context.Background(), taskID, userID)
				if err != nil {
					// Task might be completed already
					break
				}

				err = workflowExecutionService.ApproveWorkflowTask(
					context.Background(),
					taskID,
					userID,
					fmt.Sprintf("signature-user%d", approvalIndex),
					fmt.Sprintf("Approval %d", approvalIndex),
				)

				if contains(tc.expectSuccess, approvalIndex) {
					assert.NoError(t, err, "Approval %d should succeed", approvalIndex)
				}

				// Check if task is completed
				var task models.WorkflowTask
				assert.NoError(t, db.Where("id = ?", taskID).First(&task).Error)

				if approvalIndex == len(tc.expectSuccess) {
					assert.Equal(t, "completed", task.Status, "Task should be completed after required approvals")
					break
				} else if contains(tc.expectSuccess, approvalIndex) {
					assert.Equal(t, "partially_approved", task.Status, "Task should be partially approved")
				}

				// Unclaim for next user
				workflowExecutionService.UnclaimWorkflowTask(context.Background(), taskID, userID)
			}

			t.Logf("SUCCESS: %s works correctly", tc.name)
		})
	}
}

// Helper function to check if slice contains value
func contains(slice []int, value int) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}