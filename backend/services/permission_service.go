package services

import (
	"fmt"
	"log"

	"gorm.io/gorm"
)

// Permission represents a single permission (resource + action)
type Permission struct {
	Resource string
	Action   string
}

// PermissionService handles all permission-related operations
type PermissionService struct {
	db *gorm.DB
}

// NewPermissionService creates a new permission service
func NewPermissionService(db *gorm.DB) *PermissionService {
	return &PermissionService{db: db}
}

// RolePermissions maps roles to their permissions
// This is the hardcoded role-to-permission mapping for Phase 3
var RolePermissions = map[string][]Permission{
	"admin": {
		{Resource: "requisition", Action: "view"},
		{Resource: "requisition", Action: "create"},
		{Resource: "requisition", Action: "edit"},
		{Resource: "requisition", Action: "delete"},
		{Resource: "requisition", Action: "approve"},
		{Resource: "requisition", Action: "reject"},

		{Resource: "budget", Action: "view"},
		{Resource: "budget", Action: "create"},
		{Resource: "budget", Action: "edit"},
		{Resource: "budget", Action: "delete"},
		{Resource: "budget", Action: "approve"},
		{Resource: "budget", Action: "reject"},

		{Resource: "purchase_order", Action: "view"},
		{Resource: "purchase_order", Action: "create"},
		{Resource: "purchase_order", Action: "edit"},
		{Resource: "purchase_order", Action: "delete"},
		{Resource: "purchase_order", Action: "approve"},
		{Resource: "purchase_order", Action: "reject"},

		{Resource: "payment_voucher", Action: "view"},
		{Resource: "payment_voucher", Action: "create"},
		{Resource: "payment_voucher", Action: "edit"},
		{Resource: "payment_voucher", Action: "delete"},
		{Resource: "payment_voucher", Action: "approve"},
		{Resource: "payment_voucher", Action: "reject"},

		{Resource: "grn", Action: "view"},
		{Resource: "grn", Action: "create"},
		{Resource: "grn", Action: "edit"},
		{Resource: "grn", Action: "delete"},

		{Resource: "vendor", Action: "view"},
		{Resource: "vendor", Action: "create"},
		{Resource: "vendor", Action: "edit"},
		{Resource: "vendor", Action: "delete"},

		{Resource: "category", Action: "view"},
		{Resource: "category", Action: "create"},
		{Resource: "category", Action: "edit"},
		{Resource: "category", Action: "delete"},

		{Resource: "organization", Action: "view"},
		{Resource: "organization", Action: "edit"},
		{Resource: "organization", Action: "manage_users"},
		{Resource: "organization", Action: "manage_workflows"},

		{Resource: "analytics", Action: "view"},
		{Resource: "audit_log", Action: "view"},
	},
	"approver": {
		{Resource: "requisition", Action: "view"},
		{Resource: "requisition", Action: "create"},
		{Resource: "requisition", Action: "edit"},
		{Resource: "requisition", Action: "approve"},
		{Resource: "requisition", Action: "reject"},

		{Resource: "budget", Action: "view"},
		{Resource: "budget", Action: "approve"},
		{Resource: "budget", Action: "reject"},

		{Resource: "purchase_order", Action: "view"},
		{Resource: "purchase_order", Action: "approve"},
		{Resource: "purchase_order", Action: "reject"},

		{Resource: "payment_voucher", Action: "view"},
		{Resource: "payment_voucher", Action: "approve"},
		{Resource: "payment_voucher", Action: "reject"},

		{Resource: "grn", Action: "view"},

		{Resource: "vendor", Action: "view"},

		{Resource: "category", Action: "view"},

		{Resource: "analytics", Action: "view"},
	},
	"requester": {
		{Resource: "requisition", Action: "view"},
		{Resource: "requisition", Action: "create"},
		{Resource: "requisition", Action: "edit"},

		{Resource: "budget", Action: "view"},

		{Resource: "purchase_order", Action: "view"},

		{Resource: "payment_voucher", Action: "view"},

		{Resource: "grn", Action: "view"},

		{Resource: "vendor", Action: "view"},

		{Resource: "category", Action: "view"},
	},
	"finance": {
		{Resource: "requisition", Action: "view"},
		{Resource: "requisition", Action: "approve"},
		{Resource: "requisition", Action: "reject"},

		{Resource: "budget", Action: "view"},
		{Resource: "budget", Action: "create"},
		{Resource: "budget", Action: "edit"},
		{Resource: "budget", Action: "approve"},
		{Resource: "budget", Action: "reject"},

		{Resource: "purchase_order", Action: "view"},
		{Resource: "purchase_order", Action: "approve"},
		{Resource: "purchase_order", Action: "reject"},

		{Resource: "payment_voucher", Action: "view"},
		{Resource: "payment_voucher", Action: "create"},
		{Resource: "payment_voucher", Action: "edit"},
		{Resource: "payment_voucher", Action: "approve"},
		{Resource: "payment_voucher", Action: "reject"},

		{Resource: "grn", Action: "view"},

		{Resource: "vendor", Action: "view"},

		{Resource: "category", Action: "view"},

		{Resource: "analytics", Action: "view"},
		{Resource: "audit_log", Action: "view"},
	},
	"viewer": {
		{Resource: "requisition", Action: "view"},
		{Resource: "budget", Action: "view"},
		{Resource: "purchase_order", Action: "view"},
		{Resource: "payment_voucher", Action: "view"},
		{Resource: "grn", Action: "view"},
		{Resource: "vendor", Action: "view"},
		{Resource: "category", Action: "view"},
		{Resource: "analytics", Action: "view"},
	},
}

// HasPermission checks if a user with the given role has a specific permission
// This method will first check the database for custom permissions (Phase 3.5+)
// and fall back to the hardcoded role permissions (Phase 3)
func (ps *PermissionService) HasPermission(userID, organizationID, role, resource, action string) bool {
	// Check custom permissions in database first (Phase 3.5+)
	// For Phase 3, this will be empty and we'll use the hardcoded mapping
	customPermissions, err := ps.getCustomPermissions(userID, organizationID, role)
	if err == nil && len(customPermissions) > 0 {
		return ps.permissionExists(customPermissions, resource, action)
	}

	// Fall back to hardcoded role permissions
	return ps.checkRolePermission(role, resource, action)
}

// checkRolePermission checks if a role has a specific permission in the hardcoded mapping
func (ps *PermissionService) checkRolePermission(role, resource, action string) bool {
	permissions, exists := RolePermissions[role]
	if !exists {
		log.Printf("Role %s not found in permissions map", role)
		return false
	}

	for _, perm := range permissions {
		if perm.Resource == resource && perm.Action == action {
			return true
		}
	}

	return false
}

// permissionExists checks if a permission exists in a list
func (ps *PermissionService) permissionExists(permissions []Permission, resource, action string) bool {
	for _, perm := range permissions {
		if perm.Resource == resource && perm.Action == action {
			return true
		}
	}
	return false
}

// getCustomPermissions retrieves custom permissions from the database for a user
// This will be implemented in Phase 3.5 when custom roles are added
// For Phase 3, this returns empty and we fall back to hardcoded permissions
func (ps *PermissionService) getCustomPermissions(userID, organizationID, role string) ([]Permission, error) {
	// TODO: Implement in Phase 3.5
	// This would query the database for OrganizationRole and PermissionAssignment
	// to get the custom permissions defined for this role in this organization
	return nil, fmt.Errorf("custom permissions not yet implemented")
}

// GetRolePermissions returns all permissions for a given role
func (ps *PermissionService) GetRolePermissions(role string) []Permission {
	if permissions, exists := RolePermissions[role]; exists {
		return permissions
	}
	return []Permission{}
}

// GetAllRoles returns all available roles
func (ps *PermissionService) GetAllRoles() []string {
	roles := make([]string, 0)
	for role := range RolePermissions {
		roles = append(roles, role)
	}
	return roles
}

// GetResources returns all available resources
func (ps *PermissionService) GetResources() []string {
	resourceMap := make(map[string]bool)
	for _, permissions := range RolePermissions {
		for _, perm := range permissions {
			resourceMap[perm.Resource] = true
		}
	}

	resources := make([]string, 0)
	for resource := range resourceMap {
		resources = append(resources, resource)
	}
	return resources
}

// GetActionsForResource returns all available actions for a resource
func (ps *PermissionService) GetActionsForResource(resource string) []string {
	actionMap := make(map[string]bool)
	for _, permissions := range RolePermissions {
		for _, perm := range permissions {
			if perm.Resource == resource {
				actionMap[perm.Action] = true
			}
		}
	}

	actions := make([]string, 0)
	for action := range actionMap {
		actions = append(actions, action)
	}
	return actions
}
