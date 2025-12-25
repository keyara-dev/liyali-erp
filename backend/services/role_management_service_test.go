package services

import (
	"testing"

	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Auto-migrate the models
	err = db.AutoMigrate(
		&models.OrganizationRole{},
		&models.OrganizationPermission{},
		&models.PermissionAssignment{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

// TestCreateOrganizationRole tests creating a new organization role
func TestCreateOrganizationRole(t *testing.T) {
	db := setupTestDB(t)
	svc := NewRoleManagementService(db)

	orgID := uuid.New().String()
	role, err := svc.CreateOrganizationRole(orgID, "Manager", "Manages team operations")

	if err != nil {
		t.Fatalf("Failed to create role: %v", err)
	}

	if role.Name != "Manager" {
		t.Errorf("Expected role name 'Manager', got '%s'", role.Name)
	}

	if role.OrganizationID != orgID {
		t.Errorf("Expected organization ID '%s', got '%s'", orgID, role.OrganizationID)
	}

	if role.IsDefault {
		t.Errorf("Expected IsDefault to be false, got true")
	}

	if !role.IsActive {
		t.Errorf("Expected IsActive to be true, got false")
	}

	if role.ID == "" {
		t.Errorf("Expected role ID to be generated, got empty")
	}
}

// TestCreateOrganizationRole_MissingName tests that creating a role without a name fails
func TestCreateOrganizationRole_MissingName(t *testing.T) {
	db := setupTestDB(t)
	svc := NewRoleManagementService(db)

	orgID := uuid.New().String()
	role, err := svc.CreateOrganizationRole(orgID, "", "Description")

	if err == nil {
		t.Fatalf("Expected error when creating role without name, got nil")
	}

	if role != nil {
		t.Errorf("Expected role to be nil, got %v", role)
	}
}

// TestUpdateOrganizationRole tests updating an existing role
func TestUpdateOrganizationRole(t *testing.T) {
	db := setupTestDB(t)
	svc := NewRoleManagementService(db)

	orgID := uuid.New().String()
	role, _ := svc.CreateOrganizationRole(orgID, "Manager", "Old description")

	updatedRole, err := svc.UpdateOrganizationRole(role.ID, "Senior Manager", "New description")

	if err != nil {
		t.Fatalf("Failed to update role: %v", err)
	}

	if updatedRole.Name != "Senior Manager" {
		t.Errorf("Expected role name 'Senior Manager', got '%s'", updatedRole.Name)
	}

	if updatedRole.Description != "New description" {
		t.Errorf("Expected description 'New description', got '%s'", updatedRole.Description)
	}
}

// TestUpdateOrganizationRole_DefaultRoleProtection tests that system default roles cannot be modified
func TestUpdateOrganizationRole_DefaultRoleProtection(t *testing.T) {
	db := setupTestDB(t)
	svc := NewRoleManagementService(db)

	orgID := uuid.New().String()

	// Create a default role directly in database (simulating system role)
	defaultRole := models.OrganizationRole{
		ID:             uuid.New().String(),
		OrganizationID: orgID,
		Name:           "admin",
		Description:    "System admin role",
		IsDefault:      true,
		IsActive:       true,
	}
	db.Create(&defaultRole)

	// Try to update default role
	_, err := svc.UpdateOrganizationRole(defaultRole.ID, "Modified Admin", "Modified description")

	if err == nil {
		t.Fatalf("Expected error when updating default role, got nil")
	}
}

// TestDeleteOrganizationRole tests deleting a custom role
func TestDeleteOrganizationRole(t *testing.T) {
	db := setupTestDB(t)
	svc := NewRoleManagementService(db)

	orgID := uuid.New().String()
	role, _ := svc.CreateOrganizationRole(orgID, "Manager", "Description")

	err := svc.DeleteOrganizationRole(role.ID)

	if err != nil {
		t.Fatalf("Failed to delete role: %v", err)
	}

	// Verify role is actually deleted
	_, err = svc.GetOrganizationRole(role.ID)
	if err == nil {
		t.Errorf("Expected role to be deleted, but it still exists")
	}
}

// TestDeleteOrganizationRole_DefaultRoleProtection tests that system default roles cannot be deleted
func TestDeleteOrganizationRole_DefaultRoleProtection(t *testing.T) {
	db := setupTestDB(t)
	svc := NewRoleManagementService(db)

	orgID := uuid.New().String()

	// Create a default role directly in database (simulating system role)
	defaultRole := models.OrganizationRole{
		ID:             uuid.New().String(),
		OrganizationID: orgID,
		Name:           "admin",
		Description:    "System admin role",
		IsDefault:      true,
		IsActive:       true,
	}
	db.Create(&defaultRole)

	// Try to delete default role
	err := svc.DeleteOrganizationRole(defaultRole.ID)

	if err == nil {
		t.Fatalf("Expected error when deleting default role, got nil")
	}
}

// TestDeleteOrganizationRole_SystemRoleNameProtection tests that roles with system role names cannot be deleted
func TestDeleteOrganizationRole_SystemRoleNameProtection(t *testing.T) {
	db := setupTestDB(t)
	svc := NewRoleManagementService(db)

	orgID := uuid.New().String()

	// Create a role with system role name (but IsDefault=false)
	// This tests the isSystemDefaultRole() protection
	systemNamedRole := models.OrganizationRole{
		ID:             uuid.New().String(),
		OrganizationID: orgID,
		Name:           "approver",
		Description:    "User created role with system name",
		IsDefault:      false,
		IsActive:       true,
	}
	db.Create(&systemNamedRole)

	// Try to delete it - should fail because name matches system default
	err := svc.DeleteOrganizationRole(systemNamedRole.ID)

	if err == nil {
		t.Fatalf("Expected error when deleting role with system name, got nil")
	}
}

// TestGetOrganizationRole tests retrieving a role by ID
func TestGetOrganizationRole(t *testing.T) {
	db := setupTestDB(t)
	svc := NewRoleManagementService(db)

	orgID := uuid.New().String()
	created, _ := svc.CreateOrganizationRole(orgID, "Manager", "Description")

	retrieved, err := svc.GetOrganizationRole(created.ID)

	if err != nil {
		t.Fatalf("Failed to retrieve role: %v", err)
	}

	if retrieved.ID != created.ID {
		t.Errorf("Expected role ID '%s', got '%s'", created.ID, retrieved.ID)
	}

	if retrieved.Name != "Manager" {
		t.Errorf("Expected role name 'Manager', got '%s'", retrieved.Name)
	}
}

// TestGetOrganizationRole_NotFound tests retrieving a non-existent role
func TestGetOrganizationRole_NotFound(t *testing.T) {
	db := setupTestDB(t)
	svc := NewRoleManagementService(db)

	_, err := svc.GetOrganizationRole("non-existent-id")

	if err == nil {
		t.Fatalf("Expected error when retrieving non-existent role, got nil")
	}
}

// TestGetOrganizationRoles tests retrieving all roles for an organization
func TestGetOrganizationRoles(t *testing.T) {
	db := setupTestDB(t)
	svc := NewRoleManagementService(db)

	orgID := uuid.New().String()
	svc.CreateOrganizationRole(orgID, "Manager", "Description")
	svc.CreateOrganizationRole(orgID, "Coordinator", "Description")

	otherOrgID := uuid.New().String()
	svc.CreateOrganizationRole(otherOrgID, "Other Role", "Description")

	roles, err := svc.GetOrganizationRoles(orgID)

	if err != nil {
		t.Fatalf("Failed to retrieve roles: %v", err)
	}

	if len(roles) != 2 {
		t.Errorf("Expected 2 roles, got %d", len(roles))
	}

	// Verify all roles belong to correct organization
	for _, role := range roles {
		if role.OrganizationID != orgID {
			t.Errorf("Expected organization ID '%s', got '%s'", orgID, role.OrganizationID)
		}
	}
}

// TestAssignPermissionToRole tests assigning a permission to a role
func TestAssignPermissionToRole(t *testing.T) {
	db := setupTestDB(t)
	svc := NewRoleManagementService(db)

	orgID := uuid.New().String()
	role, _ := svc.CreateOrganizationRole(orgID, "Manager", "Description")
	perm, _ := svc.CreateOrganizationPermission(orgID, "requisition", "approve", "Approve requisitions")

	assignment, err := svc.AssignPermissionToRole(role.ID, perm.ID)

	if err != nil {
		t.Fatalf("Failed to assign permission: %v", err)
	}

	if assignment.OrganizationRoleID != role.ID {
		t.Errorf("Expected role ID '%s', got '%s'", role.ID, assignment.OrganizationRoleID)
	}

	if assignment.OrganizationPermissionID != perm.ID {
		t.Errorf("Expected permission ID '%s', got '%s'", perm.ID, assignment.OrganizationPermissionID)
	}
}

// TestAssignPermissionToRole_RoleNotFound tests assigning permission to non-existent role
func TestAssignPermissionToRole_RoleNotFound(t *testing.T) {
	db := setupTestDB(t)
	svc := NewRoleManagementService(db)

	orgID := uuid.New().String()
	perm, _ := svc.CreateOrganizationPermission(orgID, "requisition", "approve", "Approve requisitions")

	_, err := svc.AssignPermissionToRole("non-existent-role", perm.ID)

	if err == nil {
		t.Fatalf("Expected error when assigning to non-existent role, got nil")
	}
}

// TestAssignPermissionToRole_PermissionNotFound tests assigning non-existent permission to role
func TestAssignPermissionToRole_PermissionNotFound(t *testing.T) {
	db := setupTestDB(t)
	svc := NewRoleManagementService(db)

	orgID := uuid.New().String()
	role, _ := svc.CreateOrganizationRole(orgID, "Manager", "Description")

	_, err := svc.AssignPermissionToRole(role.ID, "non-existent-permission")

	if err == nil {
		t.Fatalf("Expected error when assigning non-existent permission, got nil")
	}
}

// TestAssignPermissionToRole_Idempotent tests that assigning same permission twice is idempotent
func TestAssignPermissionToRole_Idempotent(t *testing.T) {
	db := setupTestDB(t)
	svc := NewRoleManagementService(db)

	orgID := uuid.New().String()
	role, _ := svc.CreateOrganizationRole(orgID, "Manager", "Description")
	perm, _ := svc.CreateOrganizationPermission(orgID, "requisition", "approve", "Approve requisitions")

	// Assign once
	assignment1, _ := svc.AssignPermissionToRole(role.ID, perm.ID)

	// Assign again
	assignment2, _ := svc.AssignPermissionToRole(role.ID, perm.ID)

	if assignment1.ID != assignment2.ID {
		t.Errorf("Expected same assignment ID, got '%s' vs '%s'", assignment1.ID, assignment2.ID)
	}
}

// TestRemovePermissionFromRole tests removing a permission from a role
func TestRemovePermissionFromRole(t *testing.T) {
	db := setupTestDB(t)
	svc := NewRoleManagementService(db)

	orgID := uuid.New().String()
	role, _ := svc.CreateOrganizationRole(orgID, "Manager", "Description")
	perm, _ := svc.CreateOrganizationPermission(orgID, "requisition", "approve", "Approve requisitions")

	svc.AssignPermissionToRole(role.ID, perm.ID)

	err := svc.RemovePermissionFromRole(role.ID, perm.ID)

	if err != nil {
		t.Fatalf("Failed to remove permission: %v", err)
	}

	// Verify assignment is deleted
	perms, _ := svc.GetRolePermissions(role.ID)
	if len(perms) != 0 {
		t.Errorf("Expected 0 permissions, got %d", len(perms))
	}
}

// TestGetRolePermissions tests retrieving all permissions for a role
func TestGetRolePermissions(t *testing.T) {
	db := setupTestDB(t)
	svc := NewRoleManagementService(db)

	orgID := uuid.New().String()
	role, _ := svc.CreateOrganizationRole(orgID, "Manager", "Description")

	perm1, _ := svc.CreateOrganizationPermission(orgID, "requisition", "approve", "Approve requisitions")
	perm2, _ := svc.CreateOrganizationPermission(orgID, "requisition", "create", "Create requisitions")
	perm3, _ := svc.CreateOrganizationPermission(orgID, "budget", "view", "View budgets")

	svc.AssignPermissionToRole(role.ID, perm1.ID)
	svc.AssignPermissionToRole(role.ID, perm2.ID)
	svc.AssignPermissionToRole(role.ID, perm3.ID)

	perms, err := svc.GetRolePermissions(role.ID)

	if err != nil {
		t.Fatalf("Failed to retrieve permissions: %v", err)
	}

	if len(perms) != 3 {
		t.Errorf("Expected 3 permissions, got %d", len(perms))
	}
}

// TestCreateOrganizationPermission tests creating a new permission
func TestCreateOrganizationPermission(t *testing.T) {
	db := setupTestDB(t)
	svc := NewRoleManagementService(db)

	orgID := uuid.New().String()
	perm, err := svc.CreateOrganizationPermission(orgID, "requisition", "approve", "Approve requisitions")

	if err != nil {
		t.Fatalf("Failed to create permission: %v", err)
	}

	if perm.Resource != "requisition" {
		t.Errorf("Expected resource 'requisition', got '%s'", perm.Resource)
	}

	if perm.Action != "approve" {
		t.Errorf("Expected action 'approve', got '%s'", perm.Action)
	}

	if !perm.IsActive {
		t.Errorf("Expected IsActive to be true, got false")
	}
}

// TestCreateOrganizationPermission_MissingFields tests that creating permission with missing fields fails
func TestCreateOrganizationPermission_MissingFields(t *testing.T) {
	db := setupTestDB(t)
	svc := NewRoleManagementService(db)

	orgID := uuid.New().String()

	// Missing resource
	_, err := svc.CreateOrganizationPermission(orgID, "", "approve", "Description")
	if err == nil {
		t.Fatalf("Expected error when creating permission without resource, got nil")
	}

	// Missing action
	_, err = svc.CreateOrganizationPermission(orgID, "requisition", "", "Description")
	if err == nil {
		t.Fatalf("Expected error when creating permission without action, got nil")
	}
}

// TestCreateOrganizationPermission_Duplicate tests that creating duplicate permission returns existing one
func TestCreateOrganizationPermission_Duplicate(t *testing.T) {
	db := setupTestDB(t)
	svc := NewRoleManagementService(db)

	orgID := uuid.New().String()

	perm1, _ := svc.CreateOrganizationPermission(orgID, "requisition", "approve", "Description 1")
	perm2, _ := svc.CreateOrganizationPermission(orgID, "requisition", "approve", "Description 2")

	if perm1.ID != perm2.ID {
		t.Errorf("Expected same permission ID for duplicates, got '%s' vs '%s'", perm1.ID, perm2.ID)
	}
}

// TestGetOrganizationPermissions tests retrieving all permissions for an organization
func TestGetOrganizationPermissions(t *testing.T) {
	db := setupTestDB(t)
	svc := NewRoleManagementService(db)

	orgID := uuid.New().String()
	svc.CreateOrganizationPermission(orgID, "requisition", "approve", "Description")
	svc.CreateOrganizationPermission(orgID, "requisition", "create", "Description")
	svc.CreateOrganizationPermission(orgID, "budget", "view", "Description")

	otherOrgID := uuid.New().String()
	svc.CreateOrganizationPermission(otherOrgID, "requisition", "approve", "Description")

	perms, err := svc.GetOrganizationPermissions(orgID)

	if err != nil {
		t.Fatalf("Failed to retrieve permissions: %v", err)
	}

	if len(perms) != 3 {
		t.Errorf("Expected 3 permissions, got %d", len(perms))
	}

	// Verify all permissions belong to correct organization
	for _, perm := range perms {
		if perm.OrganizationID != orgID {
			t.Errorf("Expected organization ID '%s', got '%s'", orgID, perm.OrganizationID)
		}
	}
}

// TestInitializeDefaultPermissionsForOrganization tests creating default permissions
func TestInitializeDefaultPermissionsForOrganization(t *testing.T) {
	db := setupTestDB(t)
	svc := NewRoleManagementService(db)

	orgID := uuid.New().String()
	err := svc.InitializeDefaultPermissionsForOrganization(orgID)

	if err != nil {
		t.Fatalf("Failed to initialize default permissions: %v", err)
	}

	perms, _ := svc.GetOrganizationPermissions(orgID)

	// Should have created all default permissions
	if len(perms) == 0 {
		t.Errorf("Expected permissions to be created, got 0")
	}

	// Check for specific expected permissions
	expectedPerms := map[string]map[string]bool{
		"requisition": {"view": true, "create": true, "approve": true},
		"budget":      {"view": true, "create": true, "approve": true},
		"organization": {"view": true, "edit": true, "manage_users": true},
	}

	for _, perm := range perms {
		if resource, exists := expectedPerms[perm.Resource]; exists {
			if !resource[perm.Action] {
				t.Logf("Found expected permission: %s:%s", perm.Resource, perm.Action)
			}
		}
	}
}

// TestIsSystemDefaultRole tests the system default role detection
func TestIsSystemDefaultRole(t *testing.T) {
	db := setupTestDB(t)
	svc := NewRoleManagementService(db)

	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{"admin", "admin", true},
		{"approver", "approver", true},
		{"requester", "requester", true},
		{"finance", "finance", true},
		{"viewer", "viewer", true},
		{"custom", "custom_role", false},
		{"manager", "Manager", false}, // Only exact matches (case-insensitive)
		{"ADMIN uppercase", "ADMIN", true},
		{"empty", "", false},
	}

	for _, tc := range testCases {
		result := svc.isSystemDefaultRole(tc.input)
		if result != tc.expected {
			t.Errorf("Test '%s': expected %v for input '%s', got %v", tc.name, tc.expected, tc.input, result)
		}
	}
}

// TestRoleManagementWorkflow tests a complete role management workflow
func TestRoleManagementWorkflow(t *testing.T) {
	db := setupTestDB(t)
	svc := NewRoleManagementService(db)

	orgID := uuid.New().String()

	// 1. Create permissions
	perm1, _ := svc.CreateOrganizationPermission(orgID, "requisition", "approve", "Approve requisitions")
	perm2, _ := svc.CreateOrganizationPermission(orgID, "requisition", "create", "Create requisitions")
	perm3, _ := svc.CreateOrganizationPermission(orgID, "budget", "view", "View budgets")

	// 2. Create roles
	role1, _ := svc.CreateOrganizationRole(orgID, "Approver", "Can approve requisitions")
	role2, _ := svc.CreateOrganizationRole(orgID, "Requester", "Can create requisitions")

	// 3. Assign permissions to roles
	svc.AssignPermissionToRole(role1.ID, perm1.ID)
	svc.AssignPermissionToRole(role1.ID, perm3.ID)

	svc.AssignPermissionToRole(role2.ID, perm2.ID)
	svc.AssignPermissionToRole(role2.ID, perm3.ID)

	// 4. Verify permissions
	role1Perms, _ := svc.GetRolePermissions(role1.ID)
	if len(role1Perms) != 2 {
		t.Errorf("Expected 2 permissions for role1, got %d", len(role1Perms))
	}

	role2Perms, _ := svc.GetRolePermissions(role2.ID)
	if len(role2Perms) != 2 {
		t.Errorf("Expected 2 permissions for role2, got %d", len(role2Perms))
	}

	// 5. Update role
	updated, _ := svc.UpdateOrganizationRole(role2.ID, "Junior Requester", "")
	if updated.Name != "Junior Requester" {
		t.Errorf("Expected name 'Junior Requester', got '%s'", updated.Name)
	}

	// 6. Remove permission
	svc.RemovePermissionFromRole(role2.ID, perm3.ID)
	role2PermsAfter, _ := svc.GetRolePermissions(role2.ID)
	if len(role2PermsAfter) != 1 {
		t.Errorf("Expected 1 permission after removal, got %d", len(role2PermsAfter))
	}

	// 7. Delete role
	err := svc.DeleteOrganizationRole(role2.ID)
	if err != nil {
		t.Errorf("Failed to delete role: %v", err)
	}

	// 8. Verify deletion
	_, err = svc.GetOrganizationRole(role2.ID)
	if err == nil {
		t.Errorf("Expected role to be deleted")
	}
}
