package unit

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/services"
	"github.com/liyali/liyali-gateway/tests/helpers"
	"github.com/stretchr/testify/assert"
)

// TestConcurrentApprovalIssues demonstrates critical issues with the current workflow system
// when multiple users with the same role attempt to approve/reject the same task
func TestConcurrentApprovalIssues(t *testing.T) {
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

	// Create workflow
	workflowID := uuid.New()
	workflow := &models.Workflow{
		ID:             workflowID,
		OrganizationID: orgID,
		Name:           "Test Workflow",
		EntityType:     "requisition",
		IsActive:       true,
		IsDefault:      true,
		Version:        1,
	}

	// Set workflow stages
	stages := []models.WorkflowStage{
		{
			StageNumber:  1,
			StageName:    "Manager Approval",
			RequiredRole: "manager",
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

	t.Run("ISSUE 1: Multiple users receive the same workflow task", func(t *testing.T) {
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

		// Get pending tasks - this shows the fundamental issue
		pendingTasks, err := workflowExecutionService.GetPendingWorkflowTasks(
			context.Background(),
			orgID,
			documentID,
		)
		assert.NoError(t, err)
		assert.Len(t, pendingTasks, 1, "Should have exactly one pending task")

		task := pendingTasks[0]
		assert.Equal(t, "manager", *task.AssignedRole)
		
		// ISSUE: The task is assigned to a ROLE, not a specific user
		// This means ALL users with "manager" role can see and act on this task
		assert.Nil(t, task.AssignedUserID, "Task is assigned to role, not specific user")
		assert.Equal(t, "role", task.AssignmentType)

		// All three managers can potentially see this task in their approval queue
		// This is the root cause of all subsequent issues
		t.Logf("ISSUE IDENTIFIED: Task %s is assigned to role 'manager', not a specific user", task.ID)
		t.Logf("This means all users with 'manager' role (%s, %s, %s) can act on it", user1ID, user2ID, user3ID)
	})

	t.Run("ISSUE 2: Race condition when multiple users approve simultaneously", func(t *testing.T) {
		// Create a new document for this test
		documentID := uuid.New().String()
		
		_, err := workflowExecutionService.AssignWorkflowToDocument(
			context.Background(),
			orgID,
			documentID,
			"requisition",
			user1ID,
		)
		assert.NoError(t, err)

		pendingTasks, err := workflowExecutionService.GetPendingWorkflowTasks(
			context.Background(),
			orgID,
			documentID,
		)
		assert.NoError(t, err)
		assert.Len(t, pendingTasks, 1)

		taskID := pendingTasks[0].ID

		// Simulate concurrent approvals using goroutines
		var wg sync.WaitGroup
		var results []error
		var mu sync.Mutex

		// Three managers try to approve the same task simultaneously
		for i, userID := range []string{user1ID, user2ID, user3ID} {
			wg.Add(1)
			go func(uid string, index int) {
				defer wg.Done()
				
				// Add small delay to increase chance of race condition
				time.Sleep(time.Duration(index) * 10 * time.Millisecond)
				
				err := workflowExecutionService.ApproveWorkflowTask(
					context.Background(),
					taskID,
					uid,
					"signature-"+uid,
					"Approved by "+uid,
				)
				
				mu.Lock()
				results = append(results, err)
				mu.Unlock()
			}(userID, i)
		}

		wg.Wait()

		// Analyze results
		successCount := 0
		errorCount := 0
		for i, err := range results {
			if err == nil {
				successCount++
				t.Logf("User %d approval: SUCCESS", i+1)
			} else {
				errorCount++
				t.Logf("User %d approval: ERROR - %v", i+1, err)
			}
		}

		// ISSUE: Only one should succeed, but the system doesn't handle this properly
		// The first one to complete the transaction wins, others get "not in pending status" error
		assert.Equal(t, 1, successCount, "Only one approval should succeed")
		assert.Equal(t, 2, errorCount, "Two approvals should fail")

		t.Logf("ISSUE IDENTIFIED: Race condition exists - %d succeeded, %d failed", successCount, errorCount)
		t.Logf("The system relies on database transaction timing rather than proper concurrency control")
	})

	t.Run("ISSUE 3: Conflicting actions (approve vs reject) on same task", func(t *testing.T) {
		// Create a new document for this test
		documentID := uuid.New().String()
		
		_, err := workflowExecutionService.AssignWorkflowToDocument(
			context.Background(),
			orgID,
			documentID,
			"requisition",
			user1ID,
		)
		assert.NoError(t, err)

		pendingTasks, err := workflowExecutionService.GetPendingWorkflowTasks(
			context.Background(),
			orgID,
			documentID,
		)
		assert.NoError(t, err)
		assert.Len(t, pendingTasks, 1)

		taskID := pendingTasks[0].ID

		// Simulate conflicting actions
		var wg sync.WaitGroup
		var approveErr, rejectErr error

		// User 1 tries to approve
		wg.Add(1)
		go func() {
			defer wg.Done()
			approveErr = workflowExecutionService.ApproveWorkflowTask(
				context.Background(),
				taskID,
				user1ID,
				"approve-signature",
				"Approved",
			)
		}()

		// User 2 tries to reject (almost simultaneously)
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(5 * time.Millisecond) // Slight delay
			rejectErr = workflowExecutionService.RejectWorkflowTask(
				context.Background(),
				taskID,
				user2ID,
				"reject-signature",
				"Rejected",
			)
		}()

		wg.Wait()

		// ISSUE: One action will succeed, the other will fail with "not in pending status"
		// But which one succeeds is non-deterministic and depends on timing
		if approveErr == nil && rejectErr != nil {
			t.Logf("Approval succeeded, rejection failed: %v", rejectErr)
		} else if rejectErr == nil && approveErr != nil {
			t.Logf("Rejection succeeded, approval failed: %v", approveErr)
		} else {
			t.Logf("Unexpected result - approve err: %v, reject err: %v", approveErr, rejectErr)
		}

		t.Logf("ISSUE IDENTIFIED: Conflicting actions create non-deterministic outcomes")
		t.Logf("The system doesn't prevent conflicting actions, just fails the second one")
	})

	t.Run("ISSUE 4: No concept of 'required approvals count'", func(t *testing.T) {
		// The current system doesn't support requiring multiple approvals from the same role
		// For example: "Requires 2 out of 3 managers to approve"
		
		// Create workflow with multiple approval requirement (this is NOT currently supported)
		multiApprovalWorkflowID := uuid.New()
		multiApprovalWorkflow := &models.Workflow{
			ID:             multiApprovalWorkflowID,
			OrganizationID: orgID,
			Name:           "Multi-Approval Workflow",
			EntityType:     "budget",
			IsActive:       true,
			Version:        1,
		}

		// Current system doesn't support this concept
		stagesWithMultipleApprovals := []models.WorkflowStage{
			{
				StageNumber:       1,
				StageName:         "Manager Consensus",
				RequiredRole:      "manager",
				// MISSING: RequiredApprovalCount field
				// MISSING: Logic to handle multiple approvals from same role
			},
		}
		
		assert.NoError(t, multiApprovalWorkflow.SetStages(stagesWithMultipleApprovals))
		assert.NoError(t, db.Create(multiApprovalWorkflow).Error)

		t.Logf("ISSUE IDENTIFIED: System doesn't support requiring multiple approvals from same role")
		t.Logf("Current design: 1 approval from any user with required role = stage complete")
		t.Logf("Missing feature: Require N approvals from users with same role")
		t.Logf("Missing feature: Consensus-based approval (e.g., majority vote)")
	})

	t.Run("ISSUE 5: Task visibility and assignment ambiguity", func(t *testing.T) {
		// Create a document
		documentID := uuid.New().String()
		
		_, err := workflowExecutionService.AssignWorkflowToDocument(
			context.Background(),
			orgID,
			documentID,
			"requisition",
			user1ID,
		)
		assert.NoError(t, err)

		// All managers can see this task in their queue
		// But there's no way to "claim" or "assign" it to a specific user
		
		pendingTasks, err := workflowExecutionService.GetPendingWorkflowTasks(
			context.Background(),
			orgID,
			documentID,
		)
		assert.NoError(t, err)
		assert.Len(t, pendingTasks, 1)

		task := pendingTasks[0]
		
		// ISSUE: No mechanism to assign task to specific user
		assert.Nil(t, task.AssignedUserID, "Task not assigned to specific user")
		assert.Nil(t, task.ClaimedBy, "Task not claimed by any user")
		assert.Nil(t, task.ClaimedAt, "No claim timestamp")

		t.Logf("ISSUE IDENTIFIED: No task claiming/assignment mechanism")
		t.Logf("All users with required role see the task, but no way to prevent conflicts")
		t.Logf("Missing feature: Task claiming (first-come-first-served)")
		t.Logf("Missing feature: Explicit task assignment to specific users")
	})
}

// TestProposedSolutions demonstrates how the issues could be resolved
func TestProposedSolutions(t *testing.T) {
	t.Run("SOLUTION 1: Task claiming mechanism", func(t *testing.T) {
		// Proposed: Add task claiming before approval
		// 1. User sees task in their queue
		// 2. User "claims" the task (sets ClaimedBy and ClaimedAt)
		// 3. Only claimed user can approve/reject
		// 4. Other users see task as "claimed by X"
		
		t.Log("PROPOSED: Add ClaimWorkflowTask() method")
		t.Log("PROPOSED: Only claimed user can approve/reject")
		t.Log("PROPOSED: Unclaimed tasks can be claimed by any qualified user")
	})

	t.Run("SOLUTION 2: Multiple approval requirements", func(t *testing.T) {
		// Proposed: Add RequiredApprovalCount to WorkflowStage
		// 1. Stage requires N approvals from users with required role
		// 2. Track individual approvals in stage execution history
		// 3. Progress to next stage only when N approvals received
		
		t.Log("PROPOSED: Add RequiredApprovalCount field to WorkflowStage")
		t.Log("PROPOSED: Track multiple approvals per stage")
		t.Log("PROPOSED: Support consensus-based approval (majority, unanimous)")
	})

	t.Run("SOLUTION 3: Explicit user assignment", func(t *testing.T) {
		// Proposed: Support both role-based and user-specific assignment
		// 1. AssignmentType: "role" | "specific_user" | "user_group"
		// 2. For specific_user: AssignedUserID is set
		// 3. For role: AssignedRole is set, but can be claimed
		
		t.Log("PROPOSED: Enhanced assignment types")
		t.Log("PROPOSED: Support for user groups")
		t.Log("PROPOSED: Round-robin assignment within role")
	})

	t.Run("SOLUTION 4: Concurrency control", func(t *testing.T) {
		// Proposed: Add proper concurrency control
		// 1. Optimistic locking with version numbers
		// 2. Database-level constraints
		// 3. Proper error handling for concurrent access
		
		t.Log("PROPOSED: Add version field to WorkflowTask")
		t.Log("PROPOSED: Optimistic locking for task updates")
		t.Log("PROPOSED: Better error messages for concurrent access")
	})
}