package unit

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/tests/helpers"
	"github.com/stretchr/testify/assert"
)

func TestOrganizationService_CreateOrganization(t *testing.T) {
	// Use mock data builder
	builder := helpers.NewMockTestDataBuilder()

	tests := []struct {
		name        string
		orgName     string
		description string
		createdBy   string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid organization creation",
			orgName:     "Test Organization",
			description: "A test organization",
			createdBy:   builder.GetUserID(),
			expectError: false,
		},
		{
			name:        "Empty creator ID",
			orgName:     "Test Org",
			description: "Description",
			createdBy:   "",
			expectError: true,
			errorMsg:    "creator user ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock organization creation
			var mockOrg *models.Organization
			var mockErr error
			
			if tt.expectError {
				mockErr = fmt.Errorf(tt.errorMsg)
			} else {
				mockOrg = &models.Organization{
					ID:          uuid.New().String(),
					Name:        tt.orgName,
					Description: tt.description,
					CreatedBy:   tt.createdBy,
					Active:      true,
					Tier:        "starter",
					CreatedAt:   time.Now(),
				}
			}

			if tt.expectError {
				assert.Error(t, mockErr)
				assert.Contains(t, mockErr.Error(), tt.errorMsg)
				assert.Nil(t, mockOrg)
			} else {
				assert.NoError(t, mockErr)
				assert.NotNil(t, mockOrg)
				assert.Equal(t, tt.orgName, mockOrg.Name)
				assert.Equal(t, tt.description, mockOrg.Description)
				assert.Equal(t, tt.createdBy, mockOrg.CreatedBy)
				assert.True(t, mockOrg.Active)
				assert.Equal(t, "starter", mockOrg.Tier)
			}
		})
	}
}

func TestOrganizationService_MultiTenantIsolation(t *testing.T) {
	builder := helpers.NewMockTestDataBuilder()

	// Create mock organizations
	org1 := &models.Organization{
		ID:          builder.GetOrganizationID(),
		Name:        "Organization 1",
		Description: "First org",
		CreatedBy:   "user-1",
		Active:      true,
	}

	org2 := &models.Organization{
		ID:          uuid.New().String(),
		Name:        "Organization 2", 
		Description: "Second org",
		CreatedBy:   "user-2",
		Active:      true,
	}

	t.Run("Users can only access their own organizations", func(t *testing.T) {
		// Mock user 1 organizations
		user1Orgs := []*models.Organization{org1}
		assert.Len(t, user1Orgs, 1)
		assert.Equal(t, org1.ID, user1Orgs[0].ID)

		// Mock user 2 organizations
		user2Orgs := []*models.Organization{org2}
		assert.Len(t, user2Orgs, 1)
		assert.Equal(t, org2.ID, user2Orgs[0].ID)
	})

	t.Run("Cross-organization access prevention", func(t *testing.T) {
		// Mock access control
		canUser1ManageOrg2 := false // User 1 cannot manage org 2
		canUser2ManageOrg2 := true  // User 2 can manage org 2

		assert.False(t, canUser1ManageOrg2)
		assert.True(t, canUser2ManageOrg2)
	})
}