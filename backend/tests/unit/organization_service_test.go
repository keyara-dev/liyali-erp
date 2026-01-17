package unit

import (
	"fmt"
	"testing"
	"time"

	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/services"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupOrgTestDB creates an in-memory SQLite database for organization testing
func setupOrgTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Auto-migrate the schema
	db.AutoMigrate(
		&models.Organization{},
		&models.OrganizationMember{},
		&models.OrganizationSettings{},
		&models.User{},
	)

	return db
}

func TestOrganizationService_CreateOrganization(t *testing.T) {
	db := setupOrgTestDB()
	orgService := services.NewOrganizationService(db)

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
			createdBy:   "user-123",
			expectError: false,
		},
		{
			name:        "Empty organization name",
			orgName:     "",
			description: "Description",
			createdBy:   "user-123",
			expectError: true,
			errorMsg:    "organization name is required",
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
			// Create a user first for valid tests
			if tt.createdBy != "" && !tt.expectError {
				user := &models.User{
					ID:    tt.createdBy,
					Email: "test@example.com",
					Name:  "Test User",
				}
				db.Create(user)
			}

			org, err := orgService.CreateOrganization(tt.orgName, tt.description, tt.createdBy)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, org)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, org)
				assert.Equal(t, tt.orgName, org.Name)
				assert.Equal(t, tt.description, org.Description)
				assert.Equal(t, tt.createdBy, org.CreatedBy)
				assert.True(t, org.Active)
				assert.Equal(t, "starter", org.Tier)

				// Verify organization member was created
				var member models.OrganizationMember
				err := db.Where("organization_id = ? AND user_id = ?", org.ID, tt.createdBy).First(&member).Error
				assert.NoError(t, err)
				assert.Equal(t, "admin", member.Role)
				assert.True(t, member.Active)

				// Verify settings were created
				var settings models.OrganizationSettings
				err = db.Where("organization_id = ?", org.ID).First(&settings).Error
				assert.NoError(t, err)
			}
		})
	}
}

func TestOrganizationService_MultiTenantIsolation(t *testing.T) {
	db := setupOrgTestDB()
	orgService := services.NewOrganizationService(db)

	// Create two organizations
	user1 := &models.User{ID: "user-1", Email: "user1@example.com", Name: "User 1"}
	user2 := &models.User{ID: "user-2", Email: "user2@example.com", Name: "User 2"}
	db.Create(user1)
	db.Create(user2)

	org1, err := orgService.CreateOrganization("Organization 1", "First org", "user-1")
	assert.NoError(t, err)

	org2, err := orgService.CreateOrganization("Organization 2", "Second org", "user-2")
	assert.NoError(t, err)

	t.Run("Users can only access their own organizations", func(t *testing.T) {
		// User 1 should only see org 1
		user1Orgs, err := orgService.GetUserOrganizations("user-1")
		assert.NoError(t, err)
		assert.Len(t, user1Orgs, 1)
		assert.Equal(t, org1.ID, user1Orgs[0].ID)

		// User 2 should only see org 2
		user2Orgs, err := orgService.GetUserOrganizations("user-2")
		assert.NoError(t, err)
		assert.Len(t, user2Orgs, 1)
		assert.Equal(t, org2.ID, user2Orgs[0].ID)
	})

	t.Run("Cross-organization access prevention", func(t *testing.T) {
		// User 1 should not be able to get org 2 details
		_, err := orgService.GetOrganization(org2.ID)
		assert.NoError(t, err) // GetOrganization doesn't check user access, that's handled at handler level

		// But user 1 should not be able to manage org 2
		canManage, err := orgService.CanUserManageOrganization("user-1", org2.ID)
		assert.NoError(t, err)
		assert.False(t, canManage)

		// User 2 should be able to manage org 2
		canManage, err = orgService.CanUserManageOrganization("user-2", org2.ID)
		assert.NoError(t, err)
		assert.True(t, canManage)
	})
}

func TestOrganizationService_MembershipManagement(t *testing.T) {
	db := setupOrgTestDB()
	orgService := services.NewOrganizationService(db)

	// Create organization and users
	admin := &models.User{ID: "admin-1", Email: "admin@example.com", Name: "Admin User"}
	member := &models.User{ID: "member-1", Email: "member@example.com", Name: "Member User"}
	db.Create(admin)
	db.Create(member)

	org, err := orgService.CreateOrganization("Test Org", "Test", "admin-1")
	assert.NoError(t, err)

	t.Run("Add member to organization", func(t *testing.T) {
		err := orgService.AddMember(org.ID, "member-1", "requester")
		assert.NoError(t, err)

		// Verify member was added
		members, err := orgService.GetOrganizationMembers(org.ID)
		assert.NoError(t, err)
		assert.Len(t, members, 2) // Admin + new member

		// Find the new member
		var newMember *models.OrganizationMember
		for _, m := range members {
			if m.UserID == "member-1" {
				newMember = &m
				break
			}
		}
		assert.NotNil(t, newMember)
		assert.Equal(t, "requester", newMember.Role)
		assert.True(t, newMember.Active)
	})

	t.Run("Prevent duplicate membership", func(t *testing.T) {
		err := orgService.AddMember(org.ID, "member-1", "requester")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already a member")
	})

	t.Run("Remove member from organization", func(t *testing.T) {
		err := orgService.RemoveMember(org.ID, "member-1")
		assert.NoError(t, err)

		// Verify member was deactivated
		var member models.OrganizationMember
		err = db.Where("organization_id = ? AND user_id = ?", org.ID, "member-1").First(&member).Error
		assert.NoError(t, err)
		assert.False(t, member.Active)
	})

	t.Run("Prevent removing last admin", func(t *testing.T) {
		err := orgService.RemoveMember(org.ID, "admin-1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot remove the last admin")
	})
}

func TestOrganizationService_OrganizationSwitching(t *testing.T) {
	db := setupOrgTestDB()
	orgService := services.NewOrganizationService(db)

	// Create user and organizations
	user := &models.User{ID: "user-1", Email: "user@example.com", Name: "Test User"}
	db.Create(user)

	org1, err := orgService.CreateOrganization("Org 1", "First", "user-1")
	assert.NoError(t, err)

	org2, err := orgService.CreateOrganization("Org 2", "Second", "user-1")
	assert.NoError(t, err)

	t.Run("Switch to valid organization", func(t *testing.T) {
		err := orgService.SwitchOrganization("user-1", org2.ID)
		assert.NoError(t, err)

		// Verify user's current organization was updated
		var updatedUser models.User
		err = db.Where("id = ?", "user-1").First(&updatedUser).Error
		assert.NoError(t, err)
		assert.NotNil(t, updatedUser.CurrentOrganizationID)
		assert.Equal(t, org2.ID, *updatedUser.CurrentOrganizationID)
	})

	t.Run("Prevent switching to non-member organization", func(t *testing.T) {
		// Create another user and organization
		otherUser := &models.User{ID: "other-user", Email: "other@example.com", Name: "Other User"}
		db.Create(otherUser)

		otherOrg, err := orgService.CreateOrganization("Other Org", "Other", "other-user")
		assert.NoError(t, err)

		// User-1 should not be able to switch to other-org
		err = orgService.SwitchOrganization("user-1", otherOrg.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not a member")
	})

	t.Run("Prevent switching to inactive organization", func(t *testing.T) {
		// Deactivate org1
		db.Model(&models.Organization{}).Where("id = ?", org1.ID).Update("active", false)

		err := orgService.SwitchOrganization("user-1", org1.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found or is inactive")
	})
}

func TestOrganizationService_SettingsManagement(t *testing.T) {
	db := setupOrgTestDB()
	orgService := services.NewOrganizationService(db)

	// Create organization
	user := &models.User{ID: "user-1", Email: "user@example.com", Name: "Test User"}
	db.Create(user)

	org, err := orgService.CreateOrganization("Test Org", "Test", "user-1")
	assert.NoError(t, err)

	t.Run("Get default settings", func(t *testing.T) {
		settings, err := orgService.GetOrganizationSettings(org.ID)
		assert.NoError(t, err)
		assert.NotNil(t, settings)
		assert.Equal(t, "USD", settings.Currency)
		assert.Equal(t, 1, settings.FiscalYearStart)
	})

	t.Run("Update organization settings", func(t *testing.T) {
		newSettings := &models.OrganizationSettings{
			RequireDigitalSignatures: true,
			Currency:                "EUR",
			FiscalYearStart:         4, // April
			EnableBudgetValidation:  true,
			BudgetVarianceThreshold: 10.0,
		}

		err := orgService.UpdateOrganizationSettings(org.ID, newSettings)
		assert.NoError(t, err)

		// Verify settings were updated
		updatedSettings, err := orgService.GetOrganizationSettings(org.ID)
		assert.NoError(t, err)
		assert.True(t, updatedSettings.RequireDigitalSignatures)
		assert.Equal(t, "EUR", updatedSettings.Currency)
		assert.Equal(t, 4, updatedSettings.FiscalYearStart)
		assert.True(t, updatedSettings.EnableBudgetValidation)
		assert.Equal(t, 10.0, updatedSettings.BudgetVarianceThreshold)
	})
}

func TestOrganizationService_OrganizationDeletion(t *testing.T) {
	db := setupOrgTestDB()
	orgService := services.NewOrganizationService(db)

	// Create organization
	admin := &models.User{ID: "admin-1", Email: "admin@example.com", Name: "Admin User"}
	member := &models.User{ID: "member-1", Email: "member@example.com", Name: "Member User"}
	db.Create(admin)
	db.Create(member)

	org, err := orgService.CreateOrganization("Test Org", "Test", "admin-1")
	assert.NoError(t, err)

	// Add a member
	err = orgService.AddMember(org.ID, "member-1", "requester")
	assert.NoError(t, err)

	// Set as current organization for both users
	err = orgService.SwitchOrganization("admin-1", org.ID)
	assert.NoError(t, err)
	err = orgService.SwitchOrganization("member-1", org.ID)
	assert.NoError(t, err)

	t.Run("Non-admin cannot delete organization", func(t *testing.T) {
		err := orgService.DeleteOrganization(org.ID, "member-1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not an admin")
	})

	t.Run("Admin can delete organization", func(t *testing.T) {
		err := orgService.DeleteOrganization(org.ID, "admin-1")
		assert.NoError(t, err)

		// Verify organization is deactivated
		var deletedOrg models.Organization
		err = db.Where("id = ?", org.ID).First(&deletedOrg).Error
		assert.NoError(t, err)
		assert.False(t, deletedOrg.Active)

		// Verify all members are deactivated
		var members []models.OrganizationMember
		err = db.Where("organization_id = ?", org.ID).Find(&members).Error
		assert.NoError(t, err)
		for _, member := range members {
			assert.False(t, member.Active)
		}

		// Verify users' current organization is cleared
		var users []models.User
		err = db.Where("id IN ?", []string{"admin-1", "member-1"}).Find(&users).Error
		assert.NoError(t, err)
		for _, user := range users {
			assert.Nil(t, user.CurrentOrganizationID)
		}
	})

	t.Run("Cannot delete already deleted organization", func(t *testing.T) {
		err := orgService.DeleteOrganization(org.ID, "admin-1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found or already deleted")
	})
}

func TestOrganizationService_ConcurrentOperations(t *testing.T) {
	db := setupOrgTestDB()
	orgService := services.NewOrganizationService(db)

	// Create organization and users
	admin := &models.User{ID: "admin-1", Email: "admin@example.com", Name: "Admin User"}
	db.Create(admin)

	org, err := orgService.CreateOrganization("Test Org", "Test", "admin-1")
	assert.NoError(t, err)

	t.Run("Concurrent member additions", func(t *testing.T) {
		// Create multiple users
		for i := 1; i <= 5; i++ {
			user := &models.User{
				ID:    fmt.Sprintf("user-%d", i),
				Email: fmt.Sprintf("user%d@example.com", i),
				Name:  fmt.Sprintf("User %d", i),
			}
			db.Create(user)
		}

		// Add members concurrently
		done := make(chan bool, 5)
		errors := make(chan error, 5)

		for i := 1; i <= 5; i++ {
			go func(userID string) {
				err := orgService.AddMember(org.ID, userID, "requester")
				if err != nil {
					errors <- err
				}
				done <- true
			}(fmt.Sprintf("user-%d", i))
		}

		// Wait for all operations to complete
		for i := 0; i < 5; i++ {
			<-done
		}

		// Check for errors
		select {
		case err := <-errors:
			t.Errorf("Concurrent member addition failed: %v", err)
		default:
			// No errors, verify all members were added
			members, err := orgService.GetOrganizationMembers(org.ID)
			assert.NoError(t, err)
			assert.Len(t, members, 6) // Admin + 5 new members
		}
	})
}

func TestOrganizationService_ValidationEdgeCases(t *testing.T) {
	db := setupOrgTestDB()
	orgService := services.NewOrganizationService(db)

	t.Run("Empty parameters validation", func(t *testing.T) {
		_, err := orgService.GetOrganization("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "organization ID is required")

		_, err = orgService.GetUserOrganizations("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user ID is required")

		err = orgService.AddMember("", "user-1", "role")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "organization ID and user ID are required")

		err = orgService.RemoveMember("org-1", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "organization ID and user ID are required")
	})

	t.Run("Non-existent resource handling", func(t *testing.T) {
		_, err := orgService.GetOrganization("non-existent-org")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "organization not found")

		_, err = orgService.GetUserOrganizations("non-existent-user")
		assert.NoError(t, err) // Should return empty slice, not error

		err = orgService.AddMember("non-existent-org", "user-1", "role")
		assert.Error(t, err) // Should fail due to foreign key constraint
	})
}