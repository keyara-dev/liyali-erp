package services

import (
	"testing"

	"gorm.io/gorm"
)

// TestHasPermission tests the HasPermission function for different roles
func TestHasPermission(t *testing.T) {
	ps := NewPermissionService((*gorm.DB)(nil))

	tests := []struct {
		name           string
		userID         string
		organizationID string
		role           string
		resource       string
		action         string
		expected       bool
	}{
		// Admin permissions - should have all
		{
			name:           "Admin - view requisition",
			userID:         "user1",
			organizationID: "org1",
			role:           "admin",
			resource:       "requisition",
			action:         "view",
			expected:       true,
		},
		{
			name:           "Admin - approve requisition",
			userID:         "user1",
			organizationID: "org1",
			role:           "admin",
			resource:       "requisition",
			action:         "approve",
			expected:       true,
		},
		{
			name:           "Admin - manage organization",
			userID:         "user1",
			organizationID: "org1",
			role:           "admin",
			resource:       "organization",
			action:         "manage_users",
			expected:       true,
		},
		// Approver permissions
		{
			name:           "Approver - view requisition",
			userID:         "user2",
			organizationID: "org1",
			role:           "approver",
			resource:       "requisition",
			action:         "view",
			expected:       true,
		},
		{
			name:           "Approver - approve requisition",
			userID:         "user2",
			organizationID: "org1",
			role:           "approver",
			resource:       "requisition",
			action:         "approve",
			expected:       true,
		},
		{
			name:           "Approver - cannot delete requisition",
			userID:         "user2",
			organizationID: "org1",
			role:           "approver",
			resource:       "requisition",
			action:         "delete",
			expected:       false,
		},
		{
			name:           "Approver - cannot manage organization",
			userID:         "user2",
			organizationID: "org1",
			role:           "approver",
			resource:       "organization",
			action:         "manage_users",
			expected:       false,
		},
		// Requester permissions
		{
			name:           "Requester - view requisition",
			userID:         "user3",
			organizationID: "org1",
			role:           "requester",
			resource:       "requisition",
			action:         "view",
			expected:       true,
		},
		{
			name:           "Requester - create requisition",
			userID:         "user3",
			organizationID: "org1",
			role:           "requester",
			resource:       "requisition",
			action:         "create",
			expected:       true,
		},
		{
			name:           "Requester - cannot approve requisition",
			userID:         "user3",
			organizationID: "org1",
			role:           "requester",
			resource:       "requisition",
			action:         "approve",
			expected:       false,
		},
		{
			name:           "Requester - cannot edit budget",
			userID:         "user3",
			organizationID: "org1",
			role:           "requester",
			resource:       "budget",
			action:         "edit",
			expected:       false,
		},
		// Finance permissions
		{
			name:           "Finance - view requisition",
			userID:         "user4",
			organizationID: "org1",
			role:           "finance",
			resource:       "requisition",
			action:         "view",
			expected:       true,
		},
		{
			name:           "Finance - approve budget",
			userID:         "user4",
			organizationID: "org1",
			role:           "finance",
			resource:       "budget",
			action:         "approve",
			expected:       true,
		},
		{
			name:           "Finance - cannot create requisition",
			userID:         "user4",
			organizationID: "org1",
			role:           "finance",
			resource:       "requisition",
			action:         "create",
			expected:       false,
		},
		// Viewer permissions
		{
			name:           "Viewer - view requisition",
			userID:         "user5",
			organizationID: "org1",
			role:           "viewer",
			resource:       "requisition",
			action:         "view",
			expected:       true,
		},
		{
			name:           "Viewer - cannot create requisition",
			userID:         "user5",
			organizationID: "org1",
			role:           "viewer",
			resource:       "requisition",
			action:         "create",
			expected:       false,
		},
		{
			name:           "Viewer - cannot approve",
			userID:         "user5",
			organizationID: "org1",
			role:           "viewer",
			resource:       "budget",
			action:         "approve",
			expected:       false,
		},
		// Invalid role
		{
			name:           "Invalid role - no permissions",
			userID:         "user6",
			organizationID: "org1",
			role:           "invalid_role",
			resource:       "requisition",
			action:         "view",
			expected:       false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := ps.HasPermission(test.userID, test.organizationID, test.role, test.resource, test.action)
			if result != test.expected {
				t.Errorf("HasPermission(%s, %s, %s, %s, %s) = %v, want %v",
					test.userID, test.organizationID, test.role, test.resource, test.action,
					result, test.expected)
			}
		})
	}
}

// TestGetRolePermissions tests getting all permissions for a role
func TestGetRolePermissions(t *testing.T) {
	ps := NewPermissionService((*gorm.DB)(nil))

	tests := []struct {
		name        string
		role        string
		expectedLen int
		shouldExist bool
	}{
		{
			name:        "Admin role permissions",
			role:        "admin",
			expectedLen: 43, // Admin has 43 permissions
			shouldExist: true,
		},
		{
			name:        "Approver role permissions",
			role:        "approver",
			expectedLen: 21, // Approver has 21 permissions
			shouldExist: true,
		},
		{
			name:        "Requester role permissions",
			role:        "requester",
			expectedLen: 8, // Requester has 8 permissions
			shouldExist: true,
		},
		{
			name:        "Finance role permissions",
			role:        "finance",
			expectedLen: 21, // Finance has 21 permissions
			shouldExist: true,
		},
		{
			name:        "Viewer role permissions",
			role:        "viewer",
			expectedLen: 7, // Viewer has 7 permissions
			shouldExist: true,
		},
		{
			name:        "Invalid role",
			role:        "invalid",
			expectedLen: 0,
			shouldExist: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			perms := ps.GetRolePermissions(test.role)
			if test.shouldExist {
				if len(perms) != test.expectedLen {
					t.Errorf("GetRolePermissions(%s) returned %d permissions, want %d",
						test.role, len(perms), test.expectedLen)
				}
			} else {
				if len(perms) != 0 {
					t.Errorf("GetRolePermissions(%s) returned %d permissions, want 0",
						test.role, len(perms))
				}
			}
		})
	}
}

// TestGetAllRoles tests getting all available roles
func TestGetAllRoles(t *testing.T) {
	ps := NewPermissionService((*gorm.DB)(nil))

	roles := ps.GetAllRoles()
	expectedRoles := map[string]bool{
		"admin":     true,
		"approver":  true,
		"requester": true,
		"finance":   true,
		"viewer":    true,
	}

	if len(roles) != len(expectedRoles) {
		t.Errorf("GetAllRoles() returned %d roles, want %d", len(roles), len(expectedRoles))
	}

	for _, role := range roles {
		if !expectedRoles[role] {
			t.Errorf("GetAllRoles() returned unexpected role: %s", role)
		}
	}
}

// TestGetResources tests getting all available resources
func TestGetResources(t *testing.T) {
	ps := NewPermissionService((*gorm.DB)(nil))

	resources := ps.GetResources()
	expectedResources := map[string]bool{
		"requisition":     true,
		"budget":          true,
		"purchase_order":  true,
		"payment_voucher": true,
		"grn":             true,
		"vendor":          true,
		"category":        true,
		"organization":    true,
		"analytics":       true,
		"audit_log":       true,
	}

	if len(resources) != len(expectedResources) {
		t.Errorf("GetResources() returned %d resources, want %d", len(resources), len(expectedResources))
	}

	for _, resource := range resources {
		if !expectedResources[resource] {
			t.Errorf("GetResources() returned unexpected resource: %s", resource)
		}
	}
}

// TestGetActionsForResource tests getting all actions for a resource
func TestGetActionsForResource(t *testing.T) {
	ps := NewPermissionService((*gorm.DB)(nil))

	tests := []struct {
		name             string
		resource         string
		expectedActions  map[string]bool
		shouldHaveActions bool
	}{
		{
			name:     "Requisition actions",
			resource: "requisition",
			expectedActions: map[string]bool{
				"view":    true,
				"create":  true,
				"edit":    true,
				"delete":  true,
				"approve": true,
				"reject":  true,
			},
			shouldHaveActions: true,
		},
		{
			name:     "Budget actions",
			resource: "budget",
			expectedActions: map[string]bool{
				"view":    true,
				"create":  true,
				"edit":    true,
				"delete":  true,
				"approve": true,
				"reject":  true,
			},
			shouldHaveActions: true,
		},
		{
			name:             "Invalid resource",
			resource:         "invalid",
			expectedActions:  map[string]bool{},
			shouldHaveActions: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actions := ps.GetActionsForResource(test.resource)
			if test.shouldHaveActions {
				if len(actions) == 0 {
					t.Errorf("GetActionsForResource(%s) returned 0 actions", test.resource)
				}
				for _, action := range actions {
					if !test.expectedActions[action] {
						t.Errorf("GetActionsForResource(%s) returned unexpected action: %s",
							test.resource, action)
					}
				}
			} else {
				if len(actions) != 0 {
					t.Errorf("GetActionsForResource(%s) returned %d actions, want 0",
						test.resource, len(actions))
				}
			}
		})
	}
}
