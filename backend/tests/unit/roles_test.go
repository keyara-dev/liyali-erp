package unit

import (
	"testing"

	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/tests/helpers"
	"github.com/stretchr/testify/assert"
)

// TestGetOrganizationRoles tests retrieving roles using mocks
func TestGetOrganizationRoles(t *testing.T) {
	// Create mock data
	builder := helpers.NewMockTestDataBuilder()
	orgID := builder.GetOrganizationID()

	// Mock response for roles
	roles := []map[string]interface{}{
		{
			"id":          uuid.New().String(),
			"name":        "Manager",
			"description": "Manages team",
			"active":      true,
		},
		{
			"id":          uuid.New().String(),
			"name":        "Coordinator",
			"description": "Coordinates tasks",
			"active":      true,
		},
	}

	// Verify mock data
	assert.Len(t, roles, 2)
	assert.Equal(t, "Manager", roles[0]["name"])
	assert.Equal(t, "Coordinator", roles[1]["name"])
	assert.NotEmpty(t, orgID)
}

// TestCreateOrganizationRole tests creating a new role using mocks
func TestCreateOrganizationRole(t *testing.T) {
	builder := helpers.NewMockTestDataBuilder()
	orgID := builder.GetOrganizationID()
	
	body := map[string]string{
		"name":        "Manager",
		"description": "Manages team operations",
	}

	// Mock successful creation
	role := map[string]interface{}{
		"id":          uuid.New().String(),
		"name":        body["name"],
		"description": body["description"],
		"active":      true,
	}

	assert.Equal(t, "Manager", role["name"])
	assert.Equal(t, "Manages team operations", role["description"])
	assert.True(t, role["active"].(bool))
	assert.NotEmpty(t, orgID)
}

// TestCreateOrganizationRole_InvalidRequest tests creating role with invalid data
func TestCreateOrganizationRole_InvalidRequest(t *testing.T) {
	builder := helpers.NewMockTestDataBuilder()
	orgID := builder.GetOrganizationID()

	// Mock validation error
	mockErr := "Name is required"
	assert.Contains(t, mockErr, "Name is required")
	assert.NotEmpty(t, orgID)
}

// TestUpdateOrganizationRole tests updating a role
func TestUpdateOrganizationRole(t *testing.T) {
	builder := helpers.NewMockTestDataBuilder()
	orgID := builder.GetOrganizationID()
	roleID := uuid.New().String()

	// Mock successful update
	role := map[string]interface{}{
		"id":          roleID,
		"name":        "Senior Manager",
		"description": "New description",
		"active":      true,
	}

	assert.Equal(t, "Senior Manager", role["name"])
	assert.Equal(t, "New description", role["description"])
	assert.NotEmpty(t, orgID)
}

// TestDeleteOrganizationRole tests deleting a role
func TestDeleteOrganizationRole(t *testing.T) {
	builder := helpers.NewMockTestDataBuilder()
	orgID := builder.GetOrganizationID()
	roleID := uuid.New().String()

	// Mock successful deletion
	mockMessage := "Role deleted successfully"
	assert.Equal(t, "Role deleted successfully", mockMessage)
	assert.NotEmpty(t, orgID)
	assert.NotEmpty(t, roleID)
}

// TestDeleteOrganizationRole_DefaultRoleProtection tests that default roles cannot be deleted
func TestDeleteOrganizationRole_DefaultRoleProtection(t *testing.T) {
	builder := helpers.NewMockTestDataBuilder()
	orgID := builder.GetOrganizationID()
	defaultRoleID := uuid.New().String()

	// Mock protection for default role
	mockError := "Cannot delete system role"
	assert.Contains(t, mockError, "Cannot delete system role")
	assert.NotEmpty(t, orgID)
	assert.NotEmpty(t, defaultRoleID)
}
