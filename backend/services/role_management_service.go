package services

import (
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/models"
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
		ID:             uuid.New().String(),
		OrganizationID: organizationID,
		Name:           name,
		Description:    description,
		IsDefault:      false,
		IsActive:       true,
	}

	if err := rms.db.Create(&role).Error; err != nil {
		log.Printf("Error creating organization role: %v", err)
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

	if role.IsDefault {
		return nil, fmt.Errorf("cannot modify default system roles")
	}

	if name != "" {
		role.Name = name
	}
	if description != "" {
		role.Description = description
	}

	if err := rms.db.Save(&role).Error; err != nil {
		log.Printf("Error updating organization role: %v", err)
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

	if role.IsDefault {
		return fmt.Errorf("cannot delete default system roles")
	}

	// Delete all permission assignments for this role
	if err := rms.db.Where("organization_role_id = ?", roleID).Delete(&models.PermissionAssignment{}).Error; err != nil {
		log.Printf("Error deleting permission assignments: %v", err)
		return fmt.Errorf("failed to delete role permissions")
	}

	// Delete the role
	if err := rms.db.Delete(&role).Error; err != nil {
		log.Printf("Error deleting organization role: %v", err)
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
	if err := rms.db.Where("organization_id = ? AND is_active = ?", organizationID, true).
		Order("created_at DESC").
		Find(&roles).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch roles")
	}
	return roles, nil
}

// AssignPermissionToRole assigns a permission to a role
func (rms *RoleManagementService) AssignPermissionToRole(
	roleID string,
	permissionID string,
) (*models.PermissionAssignment, error) {
	// Check if role exists
	role := models.OrganizationRole{}
	if err := rms.db.Where("id = ?", roleID).First(&role).Error; err != nil {
		return nil, fmt.Errorf("role not found")
	}

	// Check if permission exists
	perm := models.OrganizationPermission{}
	if err := rms.db.Where("id = ?", permissionID).First(&perm).Error; err != nil {
		return nil, fmt.Errorf("permission not found")
	}

	// Check if assignment already exists
	existing := models.PermissionAssignment{}
	if err := rms.db.Where(
		"organization_role_id = ? AND organization_permission_id = ?",
		roleID,
		permissionID,
	).First(&existing).Error; err == nil {
		// Assignment already exists
		return &existing, nil
	} else if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to check existing assignment")
	}

	// Create new assignment
	assignment := models.PermissionAssignment{
		ID:                      uuid.New().String(),
		OrganizationRoleID:      roleID,
		OrganizationPermissionID: permissionID,
	}

	if err := rms.db.Create(&assignment).Error; err != nil {
		log.Printf("Error assigning permission to role: %v", err)
		return nil, fmt.Errorf("failed to assign permission")
	}

	return &assignment, nil
}

// RemovePermissionFromRole removes a permission from a role
func (rms *RoleManagementService) RemovePermissionFromRole(
	roleID string,
	permissionID string,
) error {
	if err := rms.db.Where(
		"organization_role_id = ? AND organization_permission_id = ?",
		roleID,
		permissionID,
	).Delete(&models.PermissionAssignment{}).Error; err != nil {
		log.Printf("Error removing permission from role: %v", err)
		return fmt.Errorf("failed to remove permission")
	}
	return nil
}

// GetRolePermissions retrieves all permissions assigned to a role
func (rms *RoleManagementService) GetRolePermissions(roleID string) ([]models.OrganizationPermission, error) {
	var permissions []models.OrganizationPermission

	if err := rms.db.
		Joins("INNER JOIN permission_assignments ON permission_assignments.organization_permission_id = organization_permissions.id").
		Where("permission_assignments.organization_role_id = ?", roleID).
		Find(&permissions).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch role permissions")
	}

	return permissions, nil
}

// CreateOrganizationPermission creates a new available permission in the organization
func (rms *RoleManagementService) CreateOrganizationPermission(
	organizationID string,
	resource string,
	action string,
	description string,
) (*models.OrganizationPermission, error) {
	if resource == "" || action == "" {
		return nil, fmt.Errorf("resource and action are required")
	}

	// Check if permission already exists
	existing := models.OrganizationPermission{}
	if err := rms.db.Where(
		"organization_id = ? AND resource = ? AND action = ?",
		organizationID,
		resource,
		action,
	).First(&existing).Error; err == nil {
		// Permission already exists
		if !existing.IsActive {
			existing.IsActive = true
			rms.db.Save(&existing)
		}
		return &existing, nil
	}

	permission := models.OrganizationPermission{
		ID:             uuid.New().String(),
		OrganizationID: organizationID,
		Resource:       resource,
		Action:         action,
		Description:    description,
		IsActive:       true,
	}

	if err := rms.db.Create(&permission).Error; err != nil {
		log.Printf("Error creating organization permission: %v", err)
		return nil, fmt.Errorf("failed to create permission")
	}

	return &permission, nil
}

// GetOrganizationPermissions retrieves all available permissions for an organization
func (rms *RoleManagementService) GetOrganizationPermissions(organizationID string) ([]models.OrganizationPermission, error) {
	var permissions []models.OrganizationPermission

	if err := rms.db.
		Where("organization_id = ? AND is_active = ?", organizationID, true).
		Order("resource ASC, action ASC").
		Find(&permissions).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch permissions")
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

// InitializeDefaultPermissionsForOrganization creates the default permission set for a new organization
// This is called when an organization is created to ensure all standard permissions are available
func (rms *RoleManagementService) InitializeDefaultPermissionsForOrganization(organizationID string) error {
	// Define all default permissions based on Phase 3 permission mapping
	defaultPermissions := []struct {
		resource    string
		action      string
		description string
	}{
		// Requisition
		{resource: "requisition", action: "view", description: "View requisitions"},
		{resource: "requisition", action: "create", description: "Create requisitions"},
		{resource: "requisition", action: "edit", description: "Edit requisitions"},
		{resource: "requisition", action: "delete", description: "Delete requisitions"},
		{resource: "requisition", action: "approve", description: "Approve requisitions"},
		{resource: "requisition", action: "reject", description: "Reject requisitions"},

		// Budget
		{resource: "budget", action: "view", description: "View budgets"},
		{resource: "budget", action: "create", description: "Create budgets"},
		{resource: "budget", action: "edit", description: "Edit budgets"},
		{resource: "budget", action: "delete", description: "Delete budgets"},
		{resource: "budget", action: "approve", description: "Approve budgets"},
		{resource: "budget", action: "reject", description: "Reject budgets"},

		// Purchase Order
		{resource: "purchase_order", action: "view", description: "View purchase orders"},
		{resource: "purchase_order", action: "create", description: "Create purchase orders"},
		{resource: "purchase_order", action: "edit", description: "Edit purchase orders"},
		{resource: "purchase_order", action: "delete", description: "Delete purchase orders"},
		{resource: "purchase_order", action: "approve", description: "Approve purchase orders"},
		{resource: "purchase_order", action: "reject", description: "Reject purchase orders"},

		// Payment Voucher
		{resource: "payment_voucher", action: "view", description: "View payment vouchers"},
		{resource: "payment_voucher", action: "create", description: "Create payment vouchers"},
		{resource: "payment_voucher", action: "edit", description: "Edit payment vouchers"},
		{resource: "payment_voucher", action: "delete", description: "Delete payment vouchers"},
		{resource: "payment_voucher", action: "approve", description: "Approve payment vouchers"},
		{resource: "payment_voucher", action: "reject", description: "Reject payment vouchers"},

		// GRN
		{resource: "grn", action: "view", description: "View goods received notes"},
		{resource: "grn", action: "create", description: "Create goods received notes"},
		{resource: "grn", action: "edit", description: "Edit goods received notes"},
		{resource: "grn", action: "delete", description: "Delete goods received notes"},

		// Vendor
		{resource: "vendor", action: "view", description: "View vendors"},
		{resource: "vendor", action: "create", description: "Create vendors"},
		{resource: "vendor", action: "edit", description: "Edit vendors"},
		{resource: "vendor", action: "delete", description: "Delete vendors"},

		// Category
		{resource: "category", action: "view", description: "View categories"},
		{resource: "category", action: "create", description: "Create categories"},
		{resource: "category", action: "edit", description: "Edit categories"},
		{resource: "category", action: "delete", description: "Delete categories"},

		// Organization
		{resource: "organization", action: "view", description: "View organization"},
		{resource: "organization", action: "edit", description: "Edit organization settings"},
		{resource: "organization", action: "manage_users", description: "Manage organization users"},
		{resource: "organization", action: "manage_workflows", description: "Manage organization workflows"},

		// Analytics & Audit
		{resource: "analytics", action: "view", description: "View analytics"},
		{resource: "audit_log", action: "view", description: "View audit logs"},
	}

	for _, perm := range defaultPermissions {
		if _, err := rms.CreateOrganizationPermission(
			organizationID,
			perm.resource,
			perm.action,
			perm.description,
		); err != nil {
			// Log but continue - permission might already exist
			log.Printf("Warning creating permission %s:%s - %v", perm.resource, perm.action, err)
		}
	}

	return nil
}
