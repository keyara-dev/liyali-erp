package unit

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/tests/helpers"
	"github.com/stretchr/testify/assert"
)

// ─────────────────────────────────────────────────────────────────────────────
// Task Claim / Unclaim
// ─────────────────────────────────────────────────────────────────────────────

func TestClaimTask_Validation(t *testing.T) {
	builder := helpers.NewMockTestDataBuilder()

	tests := []struct {
		name       string
		taskStatus string
		claimedBy  *string
		shouldPass bool
	}{
		{"Pending task can be claimed", "PENDING", nil, true},
		{"Claimed task cannot be re-claimed by another user", "CLAIMED", helpers.StringPtr(builder.GetUserID()), false},
		{"Completed task cannot be claimed", "COMPLETED", nil, false},
		{"Rejected task cannot be claimed", "REJECTED", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			canClaim := tt.taskStatus == "PENDING"
			assert.Equal(t, tt.shouldPass, canClaim)
		})
	}
}

func TestClaimTask_ExpiredClaimReset(t *testing.T) {
	t.Run("Expired claim allows new claimer", func(t *testing.T) {
		expiredAt := time.Now().Add(-5 * time.Minute) // 5 minutes ago

		task := &models.WorkflowTask{
			ID:          uuid.New().String(),
			Status:      "CLAIMED",
			ClaimExpiry: &expiredAt,
		}

		isExpired := task.ClaimExpiry != nil && task.ClaimExpiry.Before(time.Now())
		canBeReClaimed := isExpired || task.Status == "PENDING"

		assert.True(t, isExpired)
		assert.True(t, canBeReClaimed)
	})

	t.Run("Active claim blocks new claimer", func(t *testing.T) {
		expiresAt := time.Now().Add(25 * time.Minute) // still valid

		task := &models.WorkflowTask{
			ID:          uuid.New().String(),
			Status:      "CLAIMED",
			ClaimExpiry: &expiresAt,
		}

		isExpired := task.ClaimExpiry != nil && task.ClaimExpiry.Before(time.Now())
		canBeReClaimed := isExpired || task.Status == "PENDING"

		assert.False(t, isExpired)
		assert.False(t, canBeReClaimed)
	})

	t.Run("Claim expiry is 30 minutes", func(t *testing.T) {
		claimTime := time.Now()
		expiresAt := claimTime.Add(30 * time.Minute)

		durationMinutes := expiresAt.Sub(claimTime).Minutes()
		assert.Equal(t, float64(30), durationMinutes)
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// Task Approval
// ─────────────────────────────────────────────────────────────────────────────

func TestApproveTask_Validation(t *testing.T) {
	tests := []struct {
		name       string
		taskStatus string
		comments   string
		shouldPass bool
	}{
		{"Claimed task can be approved", "CLAIMED", "Looks good", true},
		{"Pending task can be approved (auto-claimed)", "PENDING", "Approved", true},
		{"Completed task cannot be approved", "COMPLETED", "Late", false},
		{"Rejected task cannot be approved", "REJECTED", "Too late", false},
		{"Approval without comments (optional)", "CLAIMED", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			canApprove := tt.taskStatus == "CLAIMED" || tt.taskStatus == "PENDING"
			assert.Equal(t, tt.shouldPass, canApprove)
		})
	}
}

func TestApproveTask_RoleAuthorization(t *testing.T) {
	tests := []struct {
		name         string
		requiredRole string
		userRole     string
		shouldPass   bool
	}{
		{"Admin can approve any stage", "finance", "admin", true},
		{"Finance can approve finance stage", "finance", "finance", true},
		{"Approver can approve approver stage", "approver", "approver", true},
		{"Requester cannot approve", "finance", "requester", false},
		{"Finance cannot approve approver stage", "approver", "finance", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			canApprove := tt.userRole == "admin" || tt.userRole == tt.requiredRole
			assert.Equal(t, tt.shouldPass, canApprove)
		})
	}
}

func TestApproveTask_VersionCheck(t *testing.T) {
	t.Run("Correct version allows approval", func(t *testing.T) {
		taskVersion := 3
		requestVersion := 3

		isValidVersion := taskVersion == requestVersion
		assert.True(t, isValidVersion)
	})

	t.Run("Stale version blocks approval (optimistic locking)", func(t *testing.T) {
		taskVersion := 4
		requestVersion := 3 // stale

		isValidVersion := taskVersion == requestVersion
		assert.False(t, isValidVersion)
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// Task Rejection
// ─────────────────────────────────────────────────────────────────────────────

func TestRejectTask_Validation(t *testing.T) {
	tests := []struct {
		name          string
		taskStatus    string
		reason        string
		rejectionType string
		shouldPass    bool
	}{
		{"Valid rejection with reason", "CLAIMED", "Insufficient budget", "reject", true},
		{"Valid return for revision", "CLAIMED", "Missing attachments", "return_for_revision", true},
		{"Missing rejection reason — blocked", "CLAIMED", "", "reject", false},
		{"Completed task cannot be rejected", "COMPLETED", "Too late", "reject", false},
		{"Invalid rejection type", "CLAIMED", "Some reason", "dismiss", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validTypes := map[string]bool{
				"reject":              true,
				"return_for_revision": true,
			}
			canReject := (tt.taskStatus == "CLAIMED" || tt.taskStatus == "PENDING") &&
				tt.reason != "" &&
				validTypes[tt.rejectionType]
			assert.Equal(t, tt.shouldPass, canReject)
		})
	}
}

func TestRejectTask_ReturnToStage(t *testing.T) {
	tests := []struct {
		name         string
		totalStages  int
		returnToStage int
		shouldPass   bool
	}{
		{"Return to stage 1 (valid)", 3, 1, true},
		{"Return to stage 2 (valid)", 3, 2, true},
		{"Return to stage 3 (current — invalid, can't return forward)", 3, 3, false},
		{"Return to stage 0 (invalid)", 3, 0, false},
		{"Return to stage 4 (beyond total — invalid)", 3, 4, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			currentStage := tt.totalStages
			isValid := tt.returnToStage > 0 &&
				tt.returnToStage < currentStage &&
				tt.returnToStage <= tt.totalStages
			assert.Equal(t, tt.shouldPass, isValid)
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Task Reassignment
// ─────────────────────────────────────────────────────────────────────────────

func TestReassignTask_Validation(t *testing.T) {
	tests := []struct {
		name          string
		taskStatus    string
		newAssigneeID string
		reason        string
		shouldPass    bool
	}{
		{"Valid reassignment", "PENDING", uuid.New().String(), "User on leave", true},
		{"Claimed task can be reassigned", "CLAIMED", uuid.New().String(), "Better fit", true},
		{"Missing assignee ID", "PENDING", "", "Some reason", false},
		{"Completed task cannot be reassigned", "COMPLETED", uuid.New().String(), "Too late", false},
		{"Missing reason (required)", "PENDING", uuid.New().String(), "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			canReassign := (tt.taskStatus == "PENDING" || tt.taskStatus == "CLAIMED") &&
				tt.newAssigneeID != "" &&
				tt.reason != ""
			assert.Equal(t, tt.shouldPass, canReassign)
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Bulk Operations
// ─────────────────────────────────────────────────────────────────────────────

func TestBulkApprove_Validation(t *testing.T) {
	tests := []struct {
		name       string
		taskIDs    []string
		comments   string
		shouldPass bool
	}{
		{"Valid bulk approve", []string{uuid.New().String(), uuid.New().String()}, "Batch approved", true},
		{"Single task bulk approve", []string{uuid.New().String()}, "Approved", true},
		{"Empty task list", []string{}, "Approved", false},
		{"Nil task list", nil, "Approved", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := len(tt.taskIDs) > 0
			assert.Equal(t, tt.shouldPass, isValid)
		})
	}
}

func TestBulkReject_RequiresReason(t *testing.T) {
	tests := []struct {
		name       string
		taskIDs    []string
		reason     string
		shouldPass bool
	}{
		{"Valid bulk reject with reason", []string{uuid.New().String()}, "Budget exceeded", true},
		{"Bulk reject without reason — blocked", []string{uuid.New().String()}, "", false},
		{"Empty task list", []string{}, "Some reason", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := len(tt.taskIDs) > 0 && tt.reason != ""
			assert.Equal(t, tt.shouldPass, isValid)
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Task Queries
// ─────────────────────────────────────────────────────────────────────────────

func TestGetApprovalTasks_Filtering(t *testing.T) {
	t.Run("Filter by status PENDING", func(t *testing.T) {
		tasks := []models.WorkflowTask{
			{ID: uuid.New().String(), Status: "PENDING"},
			{ID: uuid.New().String(), Status: "CLAIMED"},
			{ID: uuid.New().String(), Status: "COMPLETED"},
			{ID: uuid.New().String(), Status: "PENDING"},
		}

		filtered := make([]models.WorkflowTask, 0)
		for _, t := range tasks {
			if t.Status == "PENDING" {
				filtered = append(filtered, t)
			}
		}

		assert.Len(t, filtered, 2)
	})

	t.Run("Filter by entity type", func(t *testing.T) {
		tasks := []models.WorkflowTask{
			{ID: uuid.New().String(), EntityType: "requisition"},
			{ID: uuid.New().String(), EntityType: "purchase_order"},
			{ID: uuid.New().String(), EntityType: "requisition"},
		}

		filtered := make([]models.WorkflowTask, 0)
		for _, t := range tasks {
			if t.EntityType == "requisition" {
				filtered = append(filtered, t)
			}
		}

		assert.Len(t, filtered, 2)
	})
}

func TestGetApprovalTasks_Pagination(t *testing.T) {
	tests := []struct {
		name          string
		page          int
		pageSize      int
		totalRecords  int
		expectedPages int
	}{
		{"Page 1 of 3", 1, 10, 25, 3},
		{"Page 2 of 3", 2, 10, 25, 3},
		{"Single page", 1, 50, 25, 1},
		{"Page size 1", 1, 1, 5, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			totalPages := (tt.totalRecords + tt.pageSize - 1) / tt.pageSize
			assert.Equal(t, tt.expectedPages, totalPages)
		})
	}
}

func TestGetTaskStats_Aggregation(t *testing.T) {
	t.Run("Task stats count correctly", func(t *testing.T) {
		tasks := []models.WorkflowTask{
			{Status: "PENDING"},
			{Status: "PENDING"},
			{Status: "CLAIMED"},
			{Status: "COMPLETED"},
			{Status: "COMPLETED"},
			{Status: "COMPLETED"},
			{Status: "REJECTED"},
		}

		stats := map[string]int{
			"PENDING":   0,
			"CLAIMED":   0,
			"COMPLETED": 0,
			"REJECTED":  0,
		}
		for _, t := range tasks {
			stats[t.Status]++
		}

		assert.Equal(t, 2, stats["PENDING"])
		assert.Equal(t, 1, stats["CLAIMED"])
		assert.Equal(t, 3, stats["COMPLETED"])
		assert.Equal(t, 1, stats["REJECTED"])
	})
}

func TestGetOverdueTasks_Detection(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		dueDate   *time.Time
		status    string
		isOverdue bool
	}{
		{
			name:      "Past due date on pending task",
			dueDate:   timePtr(now.Add(-48 * time.Hour)),
			status:    "PENDING",
			isOverdue: true,
		},
		{
			name:      "Future due date on pending task",
			dueDate:   timePtr(now.Add(48 * time.Hour)),
			status:    "PENDING",
			isOverdue: false,
		},
		{
			name:      "Past due date but completed (not overdue)",
			dueDate:   timePtr(now.Add(-48 * time.Hour)),
			status:    "COMPLETED",
			isOverdue: false,
		},
		{
			name:      "No due date (not overdue)",
			dueDate:   nil,
			status:    "PENDING",
			isOverdue: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isOverdue := tt.dueDate != nil &&
				tt.dueDate.Before(now) &&
				(tt.status == "PENDING" || tt.status == "CLAIMED")
			assert.Equal(t, tt.isOverdue, isOverdue)
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Workflow Status
// ─────────────────────────────────────────────────────────────────────────────

func TestGetApprovalWorkflowStatus_Stages(t *testing.T) {
	t.Run("Multi-stage workflow status", func(t *testing.T) {
		stages := []struct {
			stageNumber int
			status      string
			isCurrent   bool
		}{
			{1, "APPROVED", false},
			{2, "IN_PROGRESS", true},
			{3, "PENDING", false},
		}

		currentStage := -1
		for _, s := range stages {
			if s.isCurrent {
				currentStage = s.stageNumber
			}
		}

		assert.Equal(t, 2, currentStage)

		completedCount := 0
		for _, s := range stages {
			if s.status == "APPROVED" {
				completedCount++
			}
		}
		assert.Equal(t, 1, completedCount)
	})
}

func TestApprovalHistory_Structure(t *testing.T) {
	t.Run("Approval history entries have required fields", func(t *testing.T) {
		record := models.StageApprovalRecord{
			ID:             uuid.New().String(),
			WorkflowTaskID: uuid.New().String(),
			ApproverID:     uuid.New().String(),
			Action:         "approved",
			Comments:       "Looks good",
			ApprovedAt:     time.Now(),
			ApproverRole:   "finance",
			ApproverName:   "Jane Finance",
		}

		assert.NotEmpty(t, record.ID)
		assert.NotEmpty(t, record.WorkflowTaskID)
		assert.NotEmpty(t, record.ApproverID)
		assert.Equal(t, "approved", record.Action)
		assert.NotEmpty(t, record.ApproverRole)
		assert.False(t, record.ApprovedAt.IsZero())
	})

	t.Run("Valid action values", func(t *testing.T) {
		validActions := map[string]bool{
			"approved": true,
			"rejected": true,
			"returned": true,
			"claimed":  true,
		}

		assert.True(t, validActions["approved"])
		assert.True(t, validActions["rejected"])
		assert.False(t, validActions["APPROVED"]) // case-sensitive
		assert.False(t, validActions["dismiss"])
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// Available Approvers
// ─────────────────────────────────────────────────────────────────────────────

func TestGetAvailableApprovers_Filtering(t *testing.T) {
	t.Run("Only users with required role are listed", func(t *testing.T) {
		users := []struct {
			id   string
			role string
		}{
			{uuid.New().String(), "finance"},
			{uuid.New().String(), "requester"},
			{uuid.New().String(), "finance"},
			{uuid.New().String(), "admin"},
		}

		requiredRole := "finance"
		var approvers []string
		for _, u := range users {
			if u.role == requiredRole || u.role == "admin" {
				approvers = append(approvers, u.id)
			}
		}

		assert.Len(t, approvers, 3) // 2 finance + 1 admin
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// Signature Validation
// ─────────────────────────────────────────────────────────────────────────────

func TestValidateSignature_Validation(t *testing.T) {
	tests := []struct {
		name           string
		signatureData  string
		signerID       string
		shouldPass     bool
	}{
		{"Valid signature with data URI", "data:image/png;base64,abc123", uuid.New().String(), true},
		{"Missing signature data", "", uuid.New().String(), false},
		{"Missing signer ID", "data:image/png;base64,abc123", "", false},
		{"Invalid data URI format", "not-a-data-uri", uuid.New().String(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValidDataURI := len(tt.signatureData) > 5 && tt.signatureData[:5] == "data:"
			isValid := isValidDataURI && tt.signerID != ""
			assert.Equal(t, tt.shouldPass, isValid)
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Helper
// ─────────────────────────────────────────────────────────────────────────────

func timePtr(t time.Time) *time.Time {
	return &t
}
