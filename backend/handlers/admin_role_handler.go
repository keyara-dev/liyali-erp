package handlers

import (
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/utils"
	"gorm.io/datatypes"
)

// AdminPermission represents a system permission for the admin console
type AdminPermission struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

// AllAdminPermissions defines the static permission list
var AllAdminPermissions = []AdminPermission{
	// Users
	{ID: "users.view", Name: "users.view", DisplayName: "View Users", Description: "View user list and details", Category: "Users"},
	{ID: "users.create", Name: "users.create", DisplayName: "Create Users", Description: "Create new users", Category: "Users"},
	{ID: "users.edit", Name: "users.edit", DisplayName: "Edit Users", Description: "Edit user profiles and roles", Category: "Users"},
	{ID: "users.delete", Name: "users.delete", DisplayName: "Delete Users", Description: "Delete or deactivate users", Category: "Users"},
	// Organizations
	{ID: "organizations.view", Name: "organizations.view", DisplayName: "View Organizations", Description: "View organization list and details", Category: "Organizations"},
	{ID: "organizations.create", Name: "organizations.create", DisplayName: "Create Organizations", Description: "Create new organizations", Category: "Organizations"},
	{ID: "organizations.edit", Name: "organizations.edit", DisplayName: "Edit Organizations", Description: "Edit organization settings", Category: "Organizations"},
	{ID: "organizations.delete", Name: "organizations.delete", DisplayName: "Delete Organizations", Description: "Delete organizations", Category: "Organizations"},
	{ID: "organizations.manage", Name: "organizations.manage", DisplayName: "Manage Organizations", Description: "Manage organization members and settings", Category: "Organizations"},
	// Budgets
	{ID: "budgets.view", Name: "budgets.view", DisplayName: "View Budgets", Description: "View budget documents", Category: "Budgets"},
	{ID: "budgets.create", Name: "budgets.create", DisplayName: "Create Budgets", Description: "Create budget documents", Category: "Budgets"},
	{ID: "budgets.edit", Name: "budgets.edit", DisplayName: "Edit Budgets", Description: "Edit budget documents", Category: "Budgets"},
	{ID: "budgets.approve", Name: "budgets.approve", DisplayName: "Approve Budgets", Description: "Approve or reject budgets", Category: "Budgets"},
	// Requisitions
	{ID: "requisitions.view", Name: "requisitions.view", DisplayName: "View Requisitions", Description: "View requisition documents", Category: "Requisitions"},
	{ID: "requisitions.create", Name: "requisitions.create", DisplayName: "Create Requisitions", Description: "Create requisition documents", Category: "Requisitions"},
	{ID: "requisitions.edit", Name: "requisitions.edit", DisplayName: "Edit Requisitions", Description: "Edit requisition documents", Category: "Requisitions"},
	{ID: "requisitions.approve", Name: "requisitions.approve", DisplayName: "Approve Requisitions", Description: "Approve or reject requisitions", Category: "Requisitions"},
	// Purchase Orders
	{ID: "purchase_orders.view", Name: "purchase_orders.view", DisplayName: "View Purchase Orders", Description: "View purchase orders", Category: "Purchase Orders"},
	{ID: "purchase_orders.create", Name: "purchase_orders.create", DisplayName: "Create Purchase Orders", Description: "Create purchase orders", Category: "Purchase Orders"},
	{ID: "purchase_orders.edit", Name: "purchase_orders.edit", DisplayName: "Edit Purchase Orders", Description: "Edit purchase orders", Category: "Purchase Orders"},
	{ID: "purchase_orders.approve", Name: "purchase_orders.approve", DisplayName: "Approve Purchase Orders", Description: "Approve or reject purchase orders", Category: "Purchase Orders"},
	// Payments
	{ID: "payments.view", Name: "payments.view", DisplayName: "View Payments", Description: "View payment vouchers", Category: "Payments"},
	{ID: "payments.create", Name: "payments.create", DisplayName: "Create Payments", Description: "Create payment vouchers", Category: "Payments"},
	{ID: "payments.approve", Name: "payments.approve", DisplayName: "Approve Payments", Description: "Approve or reject payments", Category: "Payments"},
	// Reports
	{ID: "reports.view", Name: "reports.view", DisplayName: "View Reports", Description: "View analytics and reports", Category: "Reports"},
	{ID: "reports.export", Name: "reports.export", DisplayName: "Export Reports", Description: "Export reports and data", Category: "Reports"},
	// Settings
	{ID: "settings.view", Name: "settings.view", DisplayName: "View Settings", Description: "View system settings", Category: "Settings"},
	{ID: "settings.edit", Name: "settings.edit", DisplayName: "Edit Settings", Description: "Modify system settings", Category: "Settings"},
	// Workflows
	{ID: "workflows.view", Name: "workflows.view", DisplayName: "View Workflows", Description: "View workflow definitions", Category: "Workflows"},
	{ID: "workflows.create", Name: "workflows.create", DisplayName: "Create Workflows", Description: "Create workflow definitions", Category: "Workflows"},
	{ID: "workflows.edit", Name: "workflows.edit", DisplayName: "Edit Workflows", Description: "Edit workflow definitions", Category: "Workflows"},
	{ID: "workflows.delete", Name: "workflows.delete", DisplayName: "Delete Workflows", Description: "Delete workflow definitions", Category: "Workflows"},
	// Audit
	{ID: "audit.view", Name: "audit.view", DisplayName: "View Audit Logs", Description: "View audit trail", Category: "Audit"},
}

// titleCase capitalizes the first letter of a string
func titleCase(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// formatPermissionName formats a permission ID like "requisition:view" into "Requisition View"
func formatPermissionName(pid string) string {
	// Handle colon-separated (org-level: "requisition:view")
	if parts := strings.SplitN(pid, ":", 2); len(parts) == 2 {
		resource := strings.ReplaceAll(parts[0], "_", " ")
		action := strings.ReplaceAll(parts[1], "_", " ")
		return titleCase(resource) + " " + titleCase(action)
	}
	// Handle dot-separated (admin-level: "users.view")
	if parts := strings.SplitN(pid, ".", 2); len(parts) == 2 {
		resource := strings.ReplaceAll(parts[0], "_", " ")
		action := strings.ReplaceAll(parts[1], "_", " ")
		return titleCase(resource) + " " + titleCase(action)
	}
	return pid
}

// formatPermissionCategory extracts category from permission ID
func formatPermissionCategory(pid string) string {
	if parts := strings.SplitN(pid, ":", 2); len(parts) == 2 {
		return titleCase(strings.ReplaceAll(parts[0], "_", " "))
	}
	if parts := strings.SplitN(pid, ".", 2); len(parts) == 2 {
		return titleCase(strings.ReplaceAll(parts[0], "_", " "))
	}
	return "General"
}

// roleToFrontend transforms a raw DB role row into the Role shape the frontend expects
func roleToFrontend(role map[string]interface{}) map[string]interface{} {
	// Map 'active' -> 'is_active'
	if active, ok := role["active"]; ok {
		role["is_active"] = active
	}

	// Add display_name from name if not present
	if _, ok := role["display_name"]; !ok {
		if name, ok := role["name"].(string); ok {
			role["display_name"] = titleCase(name)
		}
	}

	// Parse permissions JSON into Permission objects
	if permRaw, ok := role["permissions"]; ok && permRaw != nil {
		var permIDs []string
		switch v := permRaw.(type) {
		case string:
			json.Unmarshal([]byte(v), &permIDs)
		case []byte:
			json.Unmarshal(v, &permIDs)
		}

		// Convert permission IDs to full Permission objects
		permissions := make([]map[string]interface{}, 0)
		for _, pid := range permIDs {
			// First try to match against admin-level permissions
			found := false
			for _, ap := range AllAdminPermissions {
				if ap.ID == pid {
					permissions = append(permissions, map[string]interface{}{
						"id":                   ap.ID,
						"name":                 ap.Name,
						"display_name":         ap.DisplayName,
						"description":          ap.Description,
						"resource":             ap.Category,
						"action":               ap.Name,
						"category":             ap.Category,
						"is_system_permission": true,
					})
					found = true
					break
				}
			}
			// If not found in admin permissions, handle as org-level permission
			if !found && pid != "" {
				category := formatPermissionCategory(pid)
				permissions = append(permissions, map[string]interface{}{
					"id":                   pid,
					"name":                 pid,
					"display_name":         formatPermissionName(pid),
					"description":          "",
					"resource":             category,
					"action":               pid,
					"category":             category,
					"is_system_permission": false,
				})
			}
		}
		role["permissions"] = permissions
	} else {
		role["permissions"] = []interface{}{}
	}

	return role
}

// toInt64 safely converts an interface{} numeric value to int64
func toInt64(v interface{}) int64 {
	switch n := v.(type) {
	case int64:
		return n
	case int:
		return int64(n)
	case int32:
		return int64(n)
	case float64:
		return int64(n)
	case float32:
		return int64(n)
	default:
		return 0
	}
}

// AdminGetAllRoles returns all roles with filters
// System roles are global (organization_id IS NULL), custom roles are per-org
func AdminGetAllRoles(c *fiber.Ctx) error {
	db := config.DB

	search := c.Query("search")
	isActive := c.Query("is_active")
	isSystemRole := c.Query("is_system_role")
	adminOnly := c.Query("admin_only")

	query := db.Table("organization_roles").
		Select(`organization_roles.*,
			(SELECT COUNT(*) FROM user_organization_roles WHERE user_organization_roles.role_id = organization_roles.id AND user_organization_roles.active = true) as user_count`)

	if search != "" {
		searchTerm := "%" + search + "%"
		query = query.Where("LOWER(organization_roles.name) LIKE LOWER(?) OR LOWER(organization_roles.description) LIKE LOWER(?)", searchTerm, searchTerm)
	}

	if isActive == "false" {
		query = query.Where("organization_roles.active = ?", false)
	} else {
		// Default to active roles only (unless explicitly requesting inactive)
		query = query.Where("organization_roles.active = ?", true)
	}

	if isSystemRole == "true" {
		query = query.Where("organization_roles.is_system_role = ?", true)
	} else if isSystemRole == "false" {
		query = query.Where("organization_roles.is_system_role = ?", false)
	}

	if adminOnly == "true" {
		query = query.Where("organization_roles.is_system_role = ? OR organization_roles.name IN ('admin', 'super_admin', 'compliance_officer')", true)
	}

	query = query.Order("organization_roles.is_system_role DESC, organization_roles.name ASC")

	var rawRoles []map[string]interface{}
	if err := query.Find(&rawRoles).Error; err != nil {
		log.Printf("Error getting roles: %v", err)
		return utils.SendInternalError(c, "Failed to retrieve roles", err)
	}

	roles := make([]map[string]interface{}, len(rawRoles))
	for i, r := range rawRoles {
		roles[i] = roleToFrontend(r)
	}

	return utils.SendSimpleSuccess(c, roles, "Roles retrieved successfully")
}

// AdminGetRoleStats returns role statistics
func AdminGetRoleStats(c *fiber.Ctx) error {
	db := config.DB

	var totalRoles, activeRoles, systemRoles, customRoles, usersWithRoles int64

	db.Table("organization_roles").Where("active = ?", true).Count(&activeRoles)
	db.Table("organization_roles").Count(&totalRoles)
	db.Table("organization_roles").Where("is_system_role = ? AND organization_id IS NULL", true).Count(&systemRoles)
	db.Table("organization_roles").Where("is_system_role = ?", false).Count(&customRoles)
	db.Table("user_organization_roles").Where("active = ?", true).Distinct("user_id").Count(&usersWithRoles)

	// Role distribution
	var roleDistribution []map[string]interface{}
	db.Table("organization_roles").
		Select(`organization_roles.id as role_id,
			organization_roles.name as role_name,
			(SELECT COUNT(*) FROM user_organization_roles WHERE user_organization_roles.role_id = organization_roles.id AND user_organization_roles.active = true) as user_count`).
		Where("organization_roles.active = ?", true).
		Find(&roleDistribution)

	var totalAssignments int64
	db.Table("user_organization_roles").Where("active = ?", true).Count(&totalAssignments)

	for i := range roleDistribution {
		if totalAssignments > 0 {
			count := toInt64(roleDistribution[i]["user_count"])
			roleDistribution[i]["percentage"] = float64(count) / float64(totalAssignments) * 100
		} else {
			roleDistribution[i]["percentage"] = 0
		}
	}

	stats := map[string]interface{}{
		"total_roles":           totalRoles,
		"active_roles":          activeRoles,
		"system_roles":          systemRoles,
		"custom_roles":          customRoles,
		"total_permissions":     len(AllAdminPermissions),
		"roles_with_users":      usersWithRoles,
		"most_used_permissions": []interface{}{},
		"role_distribution":     roleDistribution,
	}

	return utils.SendSimpleSuccess(c, stats, "Role statistics retrieved successfully")
}

// AdminGetRoleById returns a role by ID
func AdminGetRoleById(c *fiber.Ctx) error {
	db := config.DB
	roleID := c.Params("id")

	var role map[string]interface{}
	err := db.Table("organization_roles").
		Select(`organization_roles.*,
			(SELECT COUNT(*) FROM user_organization_roles WHERE user_organization_roles.role_id = organization_roles.id AND user_organization_roles.active = true) as user_count`).
		Where("organization_roles.id = ?", roleID).
		First(&role).Error

	if err != nil {
		return utils.SendNotFound(c, "Role not found")
	}

	return utils.SendSimpleSuccess(c, roleToFrontend(role), "Role retrieved successfully")
}

// AdminCreateRole creates a new role
func AdminCreateRole(c *fiber.Ctx) error {
	db := config.DB

	var request struct {
		Name          string   `json:"name"`
		DisplayName   string   `json:"display_name"`
		Description   string   `json:"description"`
		PermissionIDs []string `json:"permission_ids"`
		IsActive      bool     `json:"is_active"`
	}

	if err := c.BodyParser(&request); err != nil {
		return utils.SendBadRequest(c, "Invalid request body")
	}

	if request.Name == "" {
		return utils.SendBadRequest(c, "Role name is required")
	}

	// Convert permission IDs to JSON
	permissionsJSON, _ := json.Marshal(request.PermissionIDs)

	adminUserID := c.Locals("userID").(string)

	role := map[string]interface{}{
		"id":              uuid.New().String(),
		"organization_id": nil, // Admin-created roles are global (no specific org)
		"name":            request.Name,
		"description":     request.Description,
		"is_system_role":  false,
		"permissions":     datatypes.JSON(permissionsJSON),
		"active":          request.IsActive,
		"created_by":      adminUserID,
		"created_at":      time.Now(),
		"updated_at":      time.Now(),
	}

	if err := db.Table("organization_roles").Create(role).Error; err != nil {
		log.Printf("Error creating role: %v", err)
		return utils.SendInternalError(c, "Failed to create role", err)
	}

	return utils.SendCreatedSuccess(c, role, "Role created successfully")
}

// AdminUpdateRole updates an existing role
func AdminUpdateRole(c *fiber.Ctx) error {
	db := config.DB
	roleID := c.Params("id")

	// Check if role exists
	var existing map[string]interface{}
	if err := db.Table("organization_roles").Where("id = ?", roleID).First(&existing).Error; err != nil {
		return utils.SendNotFound(c, "Role not found")
	}

	// System roles can only be edited by super_admin
	if isSystem, ok := existing["is_system_role"].(bool); ok && isSystem {
		userRole, _ := c.Locals("userRole").(string)
		if userRole != "super_admin" {
			return utils.SendBadRequest(c, "Only super admins can modify system roles")
		}
	}

	var request struct {
		Name          *string  `json:"name,omitempty"`
		DisplayName   *string  `json:"display_name,omitempty"`
		Description   *string  `json:"description,omitempty"`
		PermissionIDs []string `json:"permission_ids,omitempty"`
		IsActive      *bool    `json:"is_active,omitempty"`
	}

	if err := c.BodyParser(&request); err != nil {
		return utils.SendBadRequest(c, "Invalid request body")
	}

	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}

	if request.Name != nil {
		updates["name"] = *request.Name
	}
	if request.Description != nil {
		updates["description"] = *request.Description
	}
	if request.PermissionIDs != nil {
		permissionsJSON, _ := json.Marshal(request.PermissionIDs)
		updates["permissions"] = datatypes.JSON(permissionsJSON)
	}
	if request.IsActive != nil {
		updates["active"] = *request.IsActive
	}

	if err := db.Table("organization_roles").Where("id = ?", roleID).Updates(updates).Error; err != nil {
		return utils.SendInternalError(c, "Failed to update role", err)
	}

	return utils.SendSimpleSuccess(c, map[string]interface{}{"id": roleID}, "Role updated successfully")
}

// AdminDeleteRole soft deletes a role
func AdminDeleteRole(c *fiber.Ctx) error {
	db := config.DB
	roleID := c.Params("id")

	// Check if system role
	var isSystem bool
	db.Table("organization_roles").Where("id = ?", roleID).Pluck("is_system_role", &isSystem)
	if isSystem {
		return utils.SendBadRequest(c, "Cannot delete system roles")
	}

	// Check if role has assigned users
	var assignedCount int64
	db.Table("user_organization_roles").Where("role_id = ? AND active = ?", roleID, true).Count(&assignedCount)
	if assignedCount > 0 {
		return utils.SendBadRequest(c, "Cannot delete role with assigned users. Remove users from this role first.")
	}

	// Soft delete (deactivate)
	if err := db.Table("organization_roles").Where("id = ?", roleID).Updates(map[string]interface{}{
		"active":     false,
		"updated_at": time.Now(),
	}).Error; err != nil {
		return utils.SendInternalError(c, "Failed to delete role", err)
	}

	return utils.SendSimpleSuccess(c, nil, "Role deleted successfully")
}

// AdminGetAllPermissions returns all system permissions
func AdminGetAllPermissions(c *fiber.Ctx) error {
	return utils.SendSimpleSuccess(c, AllAdminPermissions, "Permissions retrieved successfully")
}

// AdminGetPermissionsByCategory returns permissions grouped by category
func AdminGetPermissionsByCategory(c *fiber.Ctx) error {
	grouped := make(map[string][]AdminPermission)
	for _, perm := range AllAdminPermissions {
		grouped[perm.Category] = append(grouped[perm.Category], perm)
	}
	return utils.SendSimpleSuccess(c, grouped, "Permissions by category retrieved successfully")
}

// AdminGetRoleUsers returns users assigned to a role
func AdminGetRoleUsers(c *fiber.Ctx) error {
	db := config.DB
	roleID := c.Params("id")

	var users []map[string]interface{}
	err := db.Table("user_organization_roles").
		Select(`user_organization_roles.id as assignment_id,
			user_organization_roles.user_id,
			users.name as user_name,
			users.email as user_email,
			user_organization_roles.organization_id,
			organizations.name as organization_name,
			user_organization_roles.assigned_at,
			user_organization_roles.active`).
		Joins("LEFT JOIN users ON users.id = user_organization_roles.user_id").
		Joins("LEFT JOIN organizations ON organizations.id = user_organization_roles.organization_id").
		Where("user_organization_roles.role_id = ? AND user_organization_roles.active = ?", roleID, true).
		Find(&users).Error

	if err != nil {
		return utils.SendInternalError(c, "Failed to retrieve role users", err)
	}

	return utils.SendSimpleSuccess(c, users, "Role users retrieved successfully")
}

// AdminAssignRoleToUsers assigns a role to multiple users
func AdminAssignRoleToUsers(c *fiber.Ctx) error {
	db := config.DB
	roleID := c.Params("id")

	var request struct {
		UserIDs []string `json:"user_ids"`
	}

	if err := c.BodyParser(&request); err != nil {
		return utils.SendBadRequest(c, "Invalid request body")
	}

	if len(request.UserIDs) == 0 {
		return utils.SendBadRequest(c, "At least one user ID is required")
	}

	adminUserID := c.Locals("userID").(string)
	now := time.Now()

	for _, userID := range request.UserIDs {
		// Check if assignment already exists
		var existingCount int64
		db.Table("user_organization_roles").Where("user_id = ? AND role_id = ? AND active = ?", userID, roleID, true).Count(&existingCount)
		if existingCount > 0 {
			continue
		}

		assignment := map[string]interface{}{
			"id":              uuid.New().String(),
			"user_id":         userID,
			"organization_id": nil, // System-level assignment (no specific org)
			"role_id":         roleID,
			"assigned_by":     adminUserID,
			"assigned_at":     now,
			"active":          true,
		}
		db.Table("user_organization_roles").Create(assignment)
	}

	return utils.SendSimpleSuccess(c, nil, "Role assigned to users successfully")
}

// AdminRemoveRoleFromUsers removes a role from multiple users
func AdminRemoveRoleFromUsers(c *fiber.Ctx) error {
	db := config.DB
	roleID := c.Params("id")

	var request struct {
		UserIDs []string `json:"user_ids"`
	}

	if err := c.BodyParser(&request); err != nil {
		return utils.SendBadRequest(c, "Invalid request body")
	}

	if len(request.UserIDs) == 0 {
		return utils.SendBadRequest(c, "At least one user ID is required")
	}

	db.Table("user_organization_roles").
		Where("role_id = ? AND user_id IN ? AND active = ?", roleID, request.UserIDs, true).
		Updates(map[string]interface{}{
			"active": false,
		})

	return utils.SendSimpleSuccess(c, nil, "Role removed from users successfully")
}

// AdminCloneRole clones an existing role with a new name
func AdminCloneRole(c *fiber.Ctx) error {
	db := config.DB
	roleID := c.Params("id")

	var request struct {
		Name        string `json:"name"`
		DisplayName string `json:"display_name"`
	}

	if err := c.BodyParser(&request); err != nil {
		return utils.SendBadRequest(c, "Invalid request body")
	}

	if request.Name == "" {
		return utils.SendBadRequest(c, "Name is required for cloned role")
	}

	// Get original role
	var original map[string]interface{}
	if err := db.Table("organization_roles").Where("id = ?", roleID).First(&original).Error; err != nil {
		return utils.SendNotFound(c, "Original role not found")
	}

	adminUserID := c.Locals("userID").(string)

	cloned := map[string]interface{}{
		"id":              uuid.New().String(),
		"organization_id": nil, // Cloned roles are global custom roles
		"name":            request.Name,
		"description":     original["description"],
		"is_system_role":  false,
		"permissions":     original["permissions"],
		"active":          true,
		"created_by":      adminUserID,
		"created_at":      time.Now(),
		"updated_at":      time.Now(),
	}

	if err := db.Table("organization_roles").Create(cloned).Error; err != nil {
		return utils.SendInternalError(c, "Failed to clone role", err)
	}

	return utils.SendCreatedSuccess(c, cloned, "Role cloned successfully")
}

// AdminExportRoles exports roles as JSON
func AdminExportRoles(c *fiber.Ctx) error {
	db := config.DB

	query := db.Table("organization_roles")

	if search := c.Query("search"); search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("name ILIKE ? OR COALESCE(display_name, '') ILIKE ? OR COALESCE(description, '') ILIKE ?",
			searchPattern, searchPattern, searchPattern)
	}

	var roles []map[string]interface{}
	if err := query.
		Order("created_at DESC").
		Limit(10000).
		Find(&roles).Error; err != nil {
		log.Printf("Error exporting roles: %v", err)
		return utils.SendInternalError(c, "Failed to export roles", err)
	}

	// Enrich each role with permissions and user count
	enrichedRoles := make([]map[string]interface{}, 0, len(roles))
	for _, role := range roles {
		enriched := roleToFrontend(role)
		enrichedRoles = append(enrichedRoles, enriched)
	}

	exportData := map[string]interface{}{
		"roles":       enrichedRoles,
		"total_count": len(enrichedRoles),
		"exported_at": time.Now().Format(time.RFC3339),
	}

	c.Set("Content-Disposition", "attachment; filename=roles-export-"+time.Now().Format("2006-01-02")+".json")
	c.Set("Content-Type", "application/json")

	return c.JSON(exportData)
}

// AdminBulkUpdateRoles applies bulk actions to roles
func AdminBulkUpdateRoles(c *fiber.Ctx) error {
	db := config.DB

	var req struct {
		RoleIDs []string `json:"role_ids"`
		Action  string   `json:"action"` // activate, deactivate, delete
	}
	if err := c.BodyParser(&req); err != nil {
		return utils.SendBadRequest(c, "Invalid request body")
	}
	if len(req.RoleIDs) == 0 {
		return utils.SendBadRequest(c, "No role IDs provided")
	}
	if req.Action == "" {
		return utils.SendBadRequest(c, "Action is required")
	}

	var affected int64
	switch req.Action {
	case "activate":
		result := db.Table("organization_roles").
			Where("id IN ? AND COALESCE(is_system_role, false) = false", req.RoleIDs).
			Update("active", true)
		affected = result.RowsAffected
		if result.Error != nil {
			return utils.SendInternalError(c, "Failed to activate roles", result.Error)
		}
	case "deactivate":
		result := db.Table("organization_roles").
			Where("id IN ? AND COALESCE(is_system_role, false) = false", req.RoleIDs).
			Update("active", false)
		affected = result.RowsAffected
		if result.Error != nil {
			return utils.SendInternalError(c, "Failed to deactivate roles", result.Error)
		}
	case "delete":
		result := db.Table("organization_roles").
			Where("id IN ? AND COALESCE(is_system_role, false) = false", req.RoleIDs).
			Delete(nil)
		affected = result.RowsAffected
		if result.Error != nil {
			return utils.SendInternalError(c, "Failed to delete roles", result.Error)
		}
	default:
		return utils.SendBadRequest(c, "Invalid action. Supported: activate, deactivate, delete")
	}

	return utils.SendSimpleSuccess(c, map[string]interface{}{
		"action":   req.Action,
		"affected": affected,
		"total":    len(req.RoleIDs),
	}, "Bulk operation completed successfully")
}

// AdminGetRoleAuditHistory returns audit history for a role
func AdminGetRoleAuditHistory(c *fiber.Ctx) error {
	db := config.DB
	roleID := c.Params("id")

	var activities []map[string]interface{}
	db.Table("admin_audit_logs").
		Where("action LIKE '%role%' AND (new_value = ? OR old_value = ?)", roleID, roleID).
		Order("created_at DESC").
		Limit(50).
		Find(&activities)

	return utils.SendSimpleSuccess(c, activities, "Role audit history retrieved successfully")
}
