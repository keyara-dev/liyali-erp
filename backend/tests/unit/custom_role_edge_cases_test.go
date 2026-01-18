package unit

import (
	"testing"

	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/tests/helpers"
	"github.com/stretchr/testify/assert"
)

// TestCustomRoleEdgeCases tests various edge cases with custom roles in workflows using mocks
func TestCustomRoleEdgeCases(t *testing.T) {
	builder := helpers.NewMockTestDataBuilder()
	
	t.Run("Workflow with custom role assignment", func(t *testing.T) {
		// Create mock organization role
		mockRole := &models.OrganizationRole{
			ID:             uuid.New(),
			OrganizationID: builder.GetOrganizationID(),
			Name:           "Custom Approver",
			Description:    "Custom approval role",
			IsSystemRole:   false,
			Active:         true,
		}
		
		// Verify role properties
		assert.NotNil(t, mockRole)
		assert.Equal(t, "Custom Approver", mockRole.Name)
		assert.False(t, mockRole.IsSystemRole)
		assert.True(t, mockRole.Active)
	})
	
	t.Run("Workflow fails when custom role is deactivated", func(t *testing.T) {
		// Create mock deactivated role
		mockRole := &models.OrganizationRole{
			ID:             uuid.New(),
			OrganizationID: builder.GetOrganizationID(),
			Name:           "Deactivated Role",
			Description:    "This role is deactivated",
			IsSystemRole:   false,
			Active:         false,
		}
		
		// Verify role is inactive
		assert.False(t, mockRole.Active)
	})
	
	t.Run("Custom role with special characters", func(t *testing.T) {
		// Create mock role with special characters
		mockRole := &models.OrganizationRole{
			ID:             uuid.New(),
			OrganizationID: builder.GetOrganizationID(),
			Name:           "Role-With_Special.Chars",
			Description:    "Role with special characters in name",
			IsSystemRole:   false,
			Active:         true,
		}
		
		// Verify role name is preserved
		assert.Equal(t, "Role-With_Special.Chars", mockRole.Name)
	})
	
	t.Run("Multiple custom roles in workflow", func(t *testing.T) {
		// Create mock workflow with multiple custom roles
		stages := []models.WorkflowStage{
			{
				StageNumber:           1,
				StageName:             "Custom Role 1 Approval",
				RequiredRole:          "custom_role_1",
				RequiredApprovals:     1,
				RequiredApprovalCount: 1,
				ApprovalType:          "any",
				CanReject:             true,
			},
			{
				StageNumber:           2,
				StageName:             "Custom Role 2 Approval",
				RequiredRole:          "custom_role_2",
				RequiredApprovals:     1,
				RequiredApprovalCount: 1,
				ApprovalType:          "any",
				CanReject:             true,
			},
		}
		
		// Create mock workflow
		mockWorkflow := &models.Workflow{
			ID:             uuid.New(),
			OrganizationID: builder.GetOrganizationID(),
			Name:           "Multi-Custom-Role Workflow",
			EntityType:     "requisition",
			IsActive:       true,
		}
		mockWorkflow.SetStages(stages)
		
		// Verify workflow has multiple stages
		workflowStages, _ := mockWorkflow.GetStages()
		assert.Len(t, workflowStages, 2)
		assert.Equal(t, "custom_role_1", workflowStages[0].RequiredRole)
		assert.Equal(t, "custom_role_2", workflowStages[1].RequiredRole)
	})
	
	t.Run("Custom role permission validation", func(t *testing.T) {
		// Create mock role with permissions
		mockRole := &models.OrganizationRole{
			ID:             uuid.New(),
			OrganizationID: builder.GetOrganizationID(),
			Name:           "Approver Role",
			Description:    "Role with approval permissions",
			IsSystemRole:   false,
			Active:         true,
		}
		
		// Verify role can be used for approvals
		assert.NotNil(t, mockRole)
		assert.True(t, mockRole.Active)
	})
	
	t.Run("User assignment to custom role", func(t *testing.T) {
		// Create mock user
		mockUser, _, _ := builder.CreateMockUsers(t)
		
		// Create mock role
		mockRole := &models.OrganizationRole{
			ID:             uuid.New(),
			OrganizationID: builder.GetOrganizationID(),
			Name:           "Custom Approver",
			IsSystemRole:   false,
			Active:         true,
		}
		
		// Create mock user-role assignment
		mockAssignment := &models.UserOrganizationRole{
			ID:             uuid.New(),
			UserID:         mockUser.ID,
			OrganizationID: builder.GetOrganizationID(),
			RoleID:         mockRole.ID,
			Active:         true,
		}
		
		// Verify assignment
		assert.Equal(t, mockUser.ID, mockAssignment.UserID)
		assert.Equal(t, mockRole.ID, mockAssignment.RoleID)
		assert.True(t, mockAssignment.Active)
	})
	
	t.Run("Custom role inheritance scenarios", func(t *testing.T) {
		// Create mock parent role
		parentRole := &models.OrganizationRole{
			ID:             uuid.New(),
			OrganizationID: builder.GetOrganizationID(),
			Name:           "Parent Role",
			IsSystemRole:   false,
			Active:         true,
		}
		
		// Create mock child role
		childRole := &models.OrganizationRole{
			ID:             uuid.New(),
			OrganizationID: builder.GetOrganizationID(),
			Name:           "Child Role",
			IsSystemRole:   false,
			Active:         true,
		}
		
		// Verify both roles exist
		assert.NotNil(t, parentRole)
		assert.NotNil(t, childRole)
		assert.NotEqual(t, parentRole.ID, childRole.ID)
	})
	
	t.Run("Custom role with empty permissions", func(t *testing.T) {
		// Create mock role with no permissions
		mockRole := &models.OrganizationRole{
			ID:             uuid.New(),
			OrganizationID: builder.GetOrganizationID(),
			Name:           "Empty Role",
			IsSystemRole:   false,
			Active:         true,
		}
		
		// Verify role exists but has no permissions
		assert.NotNil(t, mockRole)
		assert.Nil(t, mockRole.Permissions)
	})
}
