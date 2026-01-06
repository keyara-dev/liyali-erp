package services

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/logging"
	"github.com/liyali/liyali-gateway/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// RoleManagementService handles creating and managing custom organization roles
// This is used in Phase 3.5 to allow organization admins to define their own roles
type RoleManagementService struct {
	db *gorm.DB
}

// NewRoleManagementService creates a new role management service
func NewRoleManagementService(db *gorm.DB) *RoleManagementService {
	return &RoleManagementService{db: db}
}

// CreateOrganizationRole creates a new custom role for an organization
func (rms *RoleManagementService) CreateOrganizationRole(
	organizationID string,
	name string,
	description string,
) (*models.OrganizationRole, error) {
	if name == "" {
		return nil, fmt.Errorf("role name is required")
	}

	role := models.OrganizationRole{
		ID:             uuid.New(),
		OrganizationID: organizationID,
		Name:           name,
		Description:    description,
		IsSystemRole:   false,
		Active:         true,
	}

	if err := rms.db.Create(&role).Error; err != nil {
		logging.WithFields(map[string]interface{}{
			"operation":       "create_organization_role",
			"role_name":       name,
			"organization_id": organizationID,
		}).WithError(err).Error("failed_to_create_organization_role")
		return nil, fmt.Errorf("failed to create role")
	}

	return &role, nil
}

// UpdateOrganizationRole updates an existing custom role
func (rms *RoleManagementService) UpdateOrganizationRole(
	roleID string,
	name string,
	description string,
) (*models.OrganizationRole, error) {
	role := models.OrganizationRole{}
	if err := rms.db.Where("id = ?", roleID).First(&role).Error; err != nil {
		return nil, fmt.Errorf("role not found")
	}

	if role.IsSystemRole {
		return nil, fmt.Errorf("cannot modify system roles")
	}

	if name != "" {
		role.Name = name
	}
	if description != "" {
		role.Description = description
	}

	if err := rms.db.Save(&role).Error; err != nil {
		logging.WithFields(map[string]interface{}{
			"operation": "update_organization_role",
			"role_id":   roleID,
		}).WithError(err).Error("failed_to_update_organization_role")
		return nil, fmt.Errorf("failed to update role")
	}

	return &role, nil
}

// DeleteOrganizationRole deletes a custom role (only user-created roles can be deleted)
// System default roles (admin, approver, requester, finance, viewer) cannot be deleted
func (rms *RoleManagementService) DeleteOrganizationRole(roleID string) error {
	role := models.OrganizationRole{}
	if err := rms.db.Where("id = ?", roleID).First(&role).Error; err != nil {
		return fmt.Errorf("role not found")
	}

	// Check if this is a default system role
	if rms.isSystemDefaultRole(role.Name) {
		return fmt.Errorf("cannot delete system default roles (admin, approver, requester, finance, viewer)")
	}

	if role.IsSystemRole {
		return fmt.Errorf("cannot delete system roles")
	}

	// Delete all permission assignments for this role (not applicable in simplified version)
	// In the simplified version, permissions are stored directly in the role

	// Delete the role
	if err := rms.db.Delete(&role).Error; err != nil {
		logging.WithFields(map[string]interface{}{
			"operation": "delete_organization_role",
			"role_id":   roleID,
		}).WithError(err).Error("failed_to_delete_organization_role")
		return fmt.Errorf("failed to delete role")
	}

	return nil
}

// GetOrganizationRole retrieves a role by ID
func (rms *RoleManagementService) GetOrganizationRole(roleID string) (*models.OrganizationRole, error) {
	role := models.OrganizationRole{}
	if err := rms.db.Where("id = ?", roleID).First(&role).Error; err != nil {
		return nil, fmt.Errorf("role not found")
	}
	return &role, nil
}

// GetOrganizationRoles retrieves all roles for an organization
func (rms *RoleManagementService) GetOrganizationRoles(organizationID string) ([]models.OrganizationRole, error) {
	var roles []models.OrganizationRole
	if err := rms.db.Where("organization_id = ? AND active = ?", organizationID, true).
		Order("created_at DESC").
		Find(&roles).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch roles")
	}
	return roles, nil
}

// AssignPermissionToRole assigns a permission to a role (simplified version)
func (rms *RoleManagementService) AssignPermissionToRole(
	roleID string,
	permissionName string,
) error {
	// Get role
	role := models.OrganizationRole{}
	if err := rms.db.Where("id = ?", roleID).First(&role).Error; err != nil {
		return fmt.Errorf("role not found")
	}

	// For now, we'll store permissions as a JSON array in the role
	// This is a simplified implementation
	var permissions []string
	if role.Permissions != nil {
		// Convert JSON to string slice
		permissionsBytes, err := role.Permissions.MarshalJSON()
		if err == nil {
			json.Unmarshal(permissionsBytes, &permissions)
		}
	}

	// Check if permission already exists
	for _, perm := range permissions {
		if perm == permissionName {
			return nil // Already exists
		}
	}

	// Add permission
	permissions = append(permissions, permissionName)
	
	// Convert back to JSON
	permissionsJSON, _ := json.Marshal(permissions)
	role.Permissions = datatypes.JSON(permissionsJSON)

	if err := rms.db.Save(&role).Error; err != nil {
		logging.WithFields(map[string]interface{}{
			"operation":       "assign_permission_to_role",
			"role_id":         roleID,
			"permission_name": permissionName,
		}).WithError(err).Error("failed_to_assign_permission_to_role")
		return fmt.Errorf("failed to assign permission")
	}

	return nil
}

// RemovePermissionFromRole removes a permission from a role (simplified version)
func (rms *RoleManagementService) RemovePermissionFromRole(
	roleID string,
	permissionName string,
) error {
	// Get role
	role := models.OrganizationRole{}
	if err := rms.db.Where("id = ?", roleID).First(&role).Error; err != nil {
		return fmt.Errorf("role not found")
	}

	// Get current permissions
	var permissions []string
	if role.Permissions != nil {
		// Convert JSON to string slice
		permissionsBytes, err := role.Permissions.MarshalJSON()
		if err == nil {
			json.Unmarshal(permissionsBytes, &permissions)
		}
	}

	// Remove permission
	newPermissions := []string{}
	for _, perm := range permissions {
		if perm != permissionName {
			newPermissions = append(newPermissions, perm)
		}
	}

	// Convert back to JSON
	permissionsJSON, _ := json.Marshal(newPermissions)
	role.Permissions = datatypes.JSON(permissionsJSON)

	if err := rms.db.Save(&role).Error; err != nil {
		logging.WithFields(map[string]interface{}{
			"operation":       "remove_permission_from_role",
			"role_id":         roleID,
			"permission_name": permissionName,
		}).WithError(err).Error("failed_to_remove_permission_from_role")
		return fmt.Errorf("failed to remove permission")
	}

	return nil
}

// GetRolePermissions retrieves all permissions assigned to a role (simplified version)
func (rms *RoleManagementService) GetRolePermissions(roleID string) ([]string, error) {
	role := models.OrganizationRole{}
	if err := rms.db.Where("id = ?", roleID).First(&role).Error; err != nil {
		return nil, fmt.Errorf("role not found")
	}

	var permissions []string
	if role.Permissions != nil {
		// Convert JSON to string slice
		permissionsBytes, err := role.Permissions.MarshalJSON()
		if err == nil {
			json.Unmarshal(permissionsBytes, &permissions)
		}
	}

	return permissions, nil
}

// GetOrganizationPermissions retrieves all available permissions (simplified version)
func (rms *RoleManagementService) GetOrganizationPermissions(organizationID string) ([]string, error) {
	// Return standard permissions that are available in the system
	permissions := []string{
		"requisition:view", "requisition:create", "requisition:edit", "requisition:delete", "requisition:approve", "requisition:reject",
		"budget:view", "budget:create", "budget:edit", "budget:delete", "budget:approve", "budget:reject",
		"purchase_order:view", "purchase_order:create", "purchase_order:edit", "purchase_order:delete", "purchase_order:approve", "purchase_order:reject",
		"payment_voucher:view", "payment_voucher:create", "payment_voucher:edit", "payment_voucher:delete", "payment_voucher:approve", "payment_voucher:reject",
		"grn:view", "grn:create", "grn:edit", "grn:delete",
		"vendor:view", "vendor:create", "vendor:edit", "vendor:delete",
		"category:view", "category:create", "category:edit", "category:delete",
		"organization:view", "organization:edit", "organization:manage_users", "organization:manage_workflows",
		"analytics:view", "audit_log:view",
	}

	return permissions, nil
}

// isSystemDefaultRole checks if a role name is one of the system default roles
// System default roles cannot be deleted by users
func (rms *RoleManagementService) isSystemDefaultRole(roleName string) bool {
	systemDefaultRoles := map[string]bool{
		"admin":     true,
		"approver":  true,
		"requester": true,
		"finance":   true,
		"viewer":    true,
	}
	// Case-insensitive check
	return systemDefaultRoles[strings.ToLower(roleName)]
}

// InitializeDefaultRolesForOrganization creates default roles for a new organization
func (rms *RoleManagementService) InitializeDefaultRolesForOrganization(organizationID string) error {
	// Define default roles
	defaultRoles := []struct {
		name        string
		description string
		permissions []string
	}{
		{
			name:        "admin",
			description: "Full administrative access",
			permissions: []string{
				"requisition:view", "requisition:create", "requisition:edit", "requisition:delete", "requisition:approve", "requisition:reject",
				"budget:view", "budget:create", "budget:edit", "budget:delete", "budget:approve", "budget:reject",
				"purchase_order:view", "purchase_order:create", "purchase_order:edit", "purchase_order:delete", "purchase_order:approve", "purchase_order:reject",
				"payment_voucher:view", "payment_voucher:create", "payment_voucher:edit", "payment_voucher:delete", "payment_voucher:approve", "payment_voucher:reject",
				"grn:view", "grn:create", "grn:edit", "grn:delete",
				"vendor:view", "vendor:create", "vendor:edit", "vendor:delete",
				"category:view", "category:create", "category:edit", "category:delete",
				"organization:view", "organization:edit", "organization:manage_users", "organization:manage_workflows",
				"analytics:view", "audit_log:view",
			},
		},
		{
			name:        "approver",
			description: "Can approve documents",
			permissions: []string{
				"requisition:view", "requisition:approve", "requisition:reject",
				"budget:view", "budget:approve", "budget:reject",
				"purchase_order:view", "purchase_order:approve", "purchase_order:reject",
				"payment_voucher:view", "payment_voucher:approve", "payment_voucher:reject",
			},
		},
		{
			name:        "requester",
			description: "Can create and manage own requests",
			permissions: []string{
				"requisition:view", "requisition:create", "requisition:edit",
				"budget:view", "budget:create", "budget:edit",
				"vendor:view", "category:view",
			},
		},
		{
			name:        "finance",
			description: "Finance team access",
			permissions: []string{
				"requisition:view", "budget:view", "budget:create", "budget:edit", "budget:approve",
				"purchase_order:view", "payment_voucher:view", "payment_voucher:create", "payment_voucher:edit", "payment_voucher:approve",
				"analytics:view", "audit_log:view",
			},
		},
		{
			name:        "viewer",
			description: "Read-only access",
			permissions: []string{
				"requisition:view", "budget:view", "purchase_order:view", "payment_voucher:view",
				"grn:view", "vendor:view", "category:view",
			},
		},
	}

	for _, roleData := range defaultRoles {
		role := models.OrganizationRole{
			ID:             uuid.New(),
			OrganizationID: organizationID,
			Name:           roleData.name,
			Description:    roleData.description,
			IsSystemRole:   true,
			Active:         true,
		}

		// Convert permissions to JSON
		permissionsJSON, _ := json.Marshal(roleData.permissions)
		role.Permissions = datatypes.JSON(permissionsJSON)

		if err := rms.db.Create(&role).Error; err != nil {
			logging.WithFields(map[string]interface{}{
				"operation":       "create_default_role",
				"role_name":       roleData.name,
				"organization_id": organizationID,
			}).WithError(err).Warn("failed_to_create_default_role")
		}
	}

	return nil
}
