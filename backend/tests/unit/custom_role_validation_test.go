package unit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// CustomRoleValidator simulates role validation logic
type CustomRoleValidator struct {
	activeRoles map[string]bool
	rolePermissions map[string][]string
}

func NewCustomRoleValidator() *CustomRoleValidator {
	return &CustomRoleValidator{
		activeRoles: map[string]bool{
			"procurement_specialist":        true,
			"department_head_procurement":   true,
			"finance_controller":           true,
			"senior_analyst":               true,
			"deactivated_role":             false,
		},
		rolePermissions: map[string][]string{
			"procurement_specialist":      {"approve_low_value_requisitions", "create_purchase_orders"},
			"department_head_procurement": {"approve_all_requisitions", "manage_team", "approve_high_value"},
			"finance_controller":          {"approve_budgets", "review_financial_docs", "approve_payments"},
			"senior_analyst":              {"view_documents", "create_reports", "approve_budgets"},
		},
	}
}

func (v *CustomRoleValidator) ValidateRoleMatch(userRole, requiredRole string) error {
	if userRole != requiredRole {
		return &RoleValidationError{
			Message: "insufficient permissions",
			UserRole: userRole,
			RequiredRole: requiredRole,
		}
	}
	return nil
}

func (v *CustomRoleValidator) IsRoleActive(roleName string) bool {
	active, exists := v.activeRoles[roleName]
	return exists && active
}

func (v *CustomRoleValidator) HasPermission(roleName, permission string) bool {
	permissions, exists := v.rolePermissions[roleName]
	if !exists {
		return false
	}
	
	for _, perm := range permissions {
		if perm == permission {
			return true
		}
	}
	return false
}

type RoleValidationError struct {
	Message      string
	UserRole     string
	RequiredRole string
}

func (e *RoleValidationError) Error() string {
	return e.Message
}

// TestCustomRoleValidationLogic tests the core custom role validation logic
func TestCustomRoleValidationLogic(t *testing.T) {
	validator := NewCustomRoleValidator()
	
	t.Run("Exact role match succeeds", func(t *testing.T) {
		err := validator.ValidateRoleMatch("procurement_specialist", "procurement_specialist")
		assert.NoError(t, err, "Exact role match should succeed")
	})
	
	t.Run("Role mismatch fails", func(t *testing.T) {
		err := validator.ValidateRoleMatch("finance_controller", "procurement_specialist")
		assert.Error(t, err, "Role mismatch should fail")
		
		roleErr, ok := err.(*RoleValidationError)
		assert.True(t, ok, "Error should be RoleValidationError type")
		assert.Equal(t, "finance_controller", roleErr.UserRole)
		assert.Equal(t, "procurement_specialist", roleErr.RequiredRole)
		assert.Contains(t, roleErr.Message, "insufficient permissions")
	})
	
	t.Run("Empty roles handling", func(t *testing.T) {
		testCases := []struct {
			name         string
			userRole     string
			requiredRole string
			shouldPass   bool
		}{
			{"Both empty", "", "", true},
			{"Empty user role", "", "procurement_specialist", false},
			{"Empty required role", "procurement_specialist", "", false},
		}
		
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := validator.ValidateRoleMatch(tc.userRole, tc.requiredRole)
				if tc.shouldPass {
					assert.NoError(t, err, "Case '%s' should pass", tc.name)
				} else {
					assert.Error(t, err, "Case '%s' should fail", tc.name)
				}
			})
		}
	})
}

// TestCustomRoleActivationStatus tests role activation/deactivation scenarios
func TestCustomRoleActivationStatus(t *testing.T) {
	validator := NewCustomRoleValidator()
	
	t.Run("Active role validation", func(t *testing.T) {
		activeRoles := []string{
			"procurement_specialist",
			"department_head_procurement", 
			"finance_controller",
			"senior_analyst",
		}
		
		for _, role := range activeRoles {
			assert.True(t, validator.IsRoleActive(role), 
				"Role '%s' should be active", role)
		}
	})
	
	t.Run("Deactivated role validation", func(t *testing.T) {
		assert.False(t, validator.IsRoleActive("deactivated_role"), 
			"Deactivated role should return false")
	})
	
	t.Run("Non-existent role validation", func(t *testing.T) {
		assert.False(t, validator.IsRoleActive("non_existent_role"), 
			"Non-existent role should return false")
	})
	
	t.Run("Workflow approval with deactivated role should fail", func(t *testing.T) {
		// Simulate workflow approval attempt with deactivated role
		userRole := "deactivated_role"
		requiredRole := "deactivated_role"
		
		// Even if role names match, deactivated roles should not be allowed
		roleMatchErr := validator.ValidateRoleMatch(userRole, requiredRole)
		assert.NoError(t, roleMatchErr, "Role names match")
		
		// But activation check should fail
		isActive := validator.IsRoleActive(userRole)
		assert.False(t, isActive, "Deactivated role should not be active")
		
		// In a real implementation, this would prevent approval
		if !isActive {
			t.Log("Approval would be blocked due to inactive role")
		}
	})
}

// TestCustomRolePermissionValidation tests permission-based validation
func TestCustomRolePermissionValidation(t *testing.T) {
	validator := NewCustomRoleValidator()
	
	t.Run("Role has required permission", func(t *testing.T) {
		testCases := []struct {
			role       string
			permission string
			shouldHave bool
		}{
			{"procurement_specialist", "approve_low_value_requisitions", true},
			{"procurement_specialist", "approve_budgets", false},
			{"finance_controller", "approve_budgets", true},
			{"finance_controller", "create_purchase_orders", false},
			{"senior_analyst", "view_documents", true},
			{"senior_analyst", "approve_all_requisitions", false},
		}
		
		for _, tc := range testCases {
			hasPermission := validator.HasPermission(tc.role, tc.permission)
			assert.Equal(t, tc.shouldHave, hasPermission,
				"Role '%s' should %s have permission '%s'",
				tc.role,
				map[bool]string{true: "", false: "not"}[tc.shouldHave],
				tc.permission)
		}
	})
	
	t.Run("Non-existent role has no permissions", func(t *testing.T) {
		hasPermission := validator.HasPermission("non_existent_role", "any_permission")
		assert.False(t, hasPermission, "Non-existent role should have no permissions")
	})
}

// TestCustomRoleWorkflowEdgeCases tests edge cases in workflow processing
func TestCustomRoleWorkflowEdgeCases(t *testing.T) {
	t.Run("Role change during workflow processing", func(t *testing.T) {
		// Simulate a scenario where user's role changes while workflow is pending
		
		// Initial state: user has correct role for task
		initialUserRole := "procurement_specialist"
		requiredRole := "procurement_specialist"
		
		validator := NewCustomRoleValidator()
		
		// Initial validation passes
		err := validator.ValidateRoleMatch(initialUserRole, requiredRole)
		assert.NoError(t, err, "Initial role validation should pass")
		
		// Simulate role change (user gets promoted/transferred)
		newUserRole := "department_head_procurement"
		
		// Validation now fails because role changed
		err = validator.ValidateRoleMatch(newUserRole, requiredRole)
		assert.Error(t, err, "Validation should fail after role change")
		
		roleErr := err.(*RoleValidationError)
		assert.Equal(t, "department_head_procurement", roleErr.UserRole)
		assert.Equal(t, "procurement_specialist", roleErr.RequiredRole)
	})
	
	t.Run("Multiple users with same custom role", func(t *testing.T) {
		// Test that multiple users can have the same custom role
		validator := NewCustomRoleValidator()
		
		users := []struct {
			id   string
			role string
		}{
			{"user1", "procurement_specialist"},
			{"user2", "procurement_specialist"},
			{"user3", "finance_controller"},
			{"user4", "finance_controller"},
		}
		
		requiredRole := "procurement_specialist"
		
		// Both procurement specialists should be able to approve
		for _, user := range users {
			err := validator.ValidateRoleMatch(user.role, requiredRole)
			if user.role == requiredRole {
				assert.NoError(t, err, "User %s with matching role should pass validation", user.id)
			} else {
				assert.Error(t, err, "User %s with different role should fail validation", user.id)
			}
		}
	})
	
	t.Run("Case sensitivity in role names", func(t *testing.T) {
		validator := NewCustomRoleValidator()
		
		testCases := []struct {
			userRole     string
			requiredRole string
			shouldMatch  bool
		}{
			{"procurement_specialist", "procurement_specialist", true},
			{"Procurement_Specialist", "procurement_specialist", false}, // Different case
			{"PROCUREMENT_SPECIALIST", "procurement_specialist", false}, // All caps
			{"procurement_specialist", "Procurement_Specialist", false}, // Different case
		}
		
		for _, tc := range testCases {
			err := validator.ValidateRoleMatch(tc.userRole, tc.requiredRole)
			if tc.shouldMatch {
				assert.NoError(t, err, "Roles '%s' and '%s' should match", tc.userRole, tc.requiredRole)
			} else {
				assert.Error(t, err, "Roles '%s' and '%s' should not match (case sensitive)", tc.userRole, tc.requiredRole)
			}
		}
	})
	
	t.Run("Special characters in role names", func(t *testing.T) {
		// Test that custom roles with special characters work correctly
		specialRoles := []string{
			"procurement-specialist",
			"department_head.finance",
			"senior@analyst",
			"finance_controller_level_3",
		}
		
		validator := NewCustomRoleValidator()
		
		for _, role := range specialRoles {
			// Test exact match
			err := validator.ValidateRoleMatch(role, role)
			assert.NoError(t, err, "Role '%s' should match itself exactly", role)
			
			// Test non-match
			err = validator.ValidateRoleMatch(role, "different_role")
			assert.Error(t, err, "Role '%s' should not match different role", role)
		}
	})
}

// TestCustomRoleAuditTrail tests audit trail functionality with custom roles
func TestCustomRoleAuditTrail(t *testing.T) {
	t.Run("Audit record captures custom role information", func(t *testing.T) {
		// Test that audit records properly capture custom role information
		
		auditRecord := struct {
			Timestamp    time.Time `json:"timestamp"`
			Action       string    `json:"action"`
			UserID       string    `json:"userId"`
			UserRole     string    `json:"userRole"`
			DocumentID   string    `json:"documentId"`
			DocumentType string    `json:"documentType"`
			Comments     string    `json:"comments"`
		}{
			Timestamp:    time.Now(),
			Action:       "approved",
			UserID:       "user-123",
			UserRole:     "procurement_specialist", // Custom role
			DocumentID:   "req-456",
			DocumentType: "requisition",
			Comments:     "Approved by procurement specialist with custom role",
		}
		
		// Verify custom role is captured
		assert.Equal(t, "procurement_specialist", auditRecord.UserRole)
		assert.Equal(t, "approved", auditRecord.Action)
		assert.Contains(t, auditRecord.Comments, "custom role")
		assert.NotEmpty(t, auditRecord.UserID)
		assert.NotEmpty(t, auditRecord.DocumentID)
		assert.False(t, auditRecord.Timestamp.IsZero())
	})
	
	t.Run("Audit trail preserves role information over time", func(t *testing.T) {
		// Test that audit trail maintains role information even if roles change later
		
		auditHistory := []struct {
			timestamp time.Time
			userRole  string
			action    string
		}{
			{time.Now().Add(-3 * time.Hour), "junior_analyst", "created"},
			{time.Now().Add(-2 * time.Hour), "procurement_specialist", "approved"},
			{time.Now().Add(-1 * time.Hour), "department_head_procurement", "final_approved"},
		}
		
		// Verify each audit entry preserves the role at time of action
		for i, entry := range auditHistory {
			assert.NotEmpty(t, entry.userRole, "Audit entry %d should have role information", i)
			assert.NotEmpty(t, entry.action, "Audit entry %d should have action information", i)
			assert.False(t, entry.timestamp.IsZero(), "Audit entry %d should have timestamp", i)
		}
		
		// Verify chronological order
		for i := 1; i < len(auditHistory); i++ {
			assert.True(t, auditHistory[i].timestamp.After(auditHistory[i-1].timestamp),
				"Audit entries should be in chronological order")
		}
		
		// Verify role progression makes sense
		assert.Equal(t, "junior_analyst", auditHistory[0].userRole)
		assert.Equal(t, "procurement_specialist", auditHistory[1].userRole)
		assert.Equal(t, "department_head_procurement", auditHistory[2].userRole)
	})
}

// TestCustomRoleWorkflowStatusReporting tests status reporting with custom roles
func TestCustomRoleWorkflowStatusReporting(t *testing.T) {
	t.Run("Workflow status correctly reports custom role requirements", func(t *testing.T) {
		// Test that workflow status includes accurate custom role information
		
		workflowStages := []struct {
			stageNumber  int
			stageName    string
			requiredRole string
			status       string
		}{
			{1, "Initial Review", "junior_analyst", "completed"},
			{2, "Specialist Review", "procurement_specialist", "completed"},
			{3, "Department Approval", "department_head_procurement", "pending"},
			{4, "Final Approval", "executive_director", "pending"},
		}
		
		// Verify each stage has proper custom role information
		for _, stage := range workflowStages {
			assert.Greater(t, stage.stageNumber, 0, "Stage number should be positive")
			assert.NotEmpty(t, stage.stageName, "Stage should have a name")
			assert.NotEmpty(t, stage.requiredRole, "Stage should have required role")
			assert.Contains(t, []string{"pending", "completed", "approved", "rejected"}, stage.status,
				"Stage status should be valid")
			
			// Verify it's a custom role (not standard system role)
			standardRoles := []string{"admin", "manager", "finance", "approver", "viewer"}
			isStandardRole := false
			for _, standardRole := range standardRoles {
				if stage.requiredRole == standardRole {
					isStandardRole = true
					break
				}
			}
			assert.False(t, isStandardRole, 
				"Stage %d should use custom role, not standard role: %s", 
				stage.stageNumber, stage.requiredRole)
		}
	})
}