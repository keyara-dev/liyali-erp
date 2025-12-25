package handlers

import (
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/services"
	"github.com/liyali/liyali-gateway/utils"
)

// CreateRoleRequest is the request body for creating a role
type CreateRoleRequest struct {
	Name        string `json:"name" validate:"required,min=3"`
	Description string `json:"description" validate:"required,min=10"`
}

// UpdateRoleRequest is the request body for updating a role
type UpdateRoleRequest struct {
	Name        string `json:"name" validate:"min=3"`
	Description string `json:"description" validate:"min=10"`
}

// RoleResponse is the response format for roles
type RoleResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsDefault   bool   `json:"isDefault"`
	IsActive    bool   `json:"isActive"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

// GetOrganizationRoles retrieves all roles for the organization
func GetOrganizationRoles(c fiber.Ctx) error {
	organizationID, ok := c.Locals("organizationID").(string)
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Organization ID not found",
		})
	}

	svc := services.NewRoleManagementService(config.DB)
	roles, err := svc.GetOrganizationRoles(organizationID)
	if err != nil {
		log.Printf("Error getting organization roles: %v", err)
		return utils.SendInternalError(c, "Failed to fetch roles", err)
	}

	responses := make([]RoleResponse, 0, len(roles))
	for _, role := range roles {
		responses = append(responses, RoleResponse{
			ID:          role.ID,
			Name:        role.Name,
			Description: role.Description,
			IsDefault:   role.IsDefault,
			IsActive:    role.IsActive,
			CreatedAt:   role.CreatedAt.String(),
			UpdatedAt:   role.UpdatedAt.String(),
		})
	}

	return utils.SendSuccess(c, fiber.StatusOK, responses, "Roles retrieved successfully", nil)
}

// CreateOrganizationRole creates a new role
func CreateOrganizationRole(c fiber.Ctx) error {
	organizationID, ok := c.Locals("organizationID").(string)
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Organization ID not found",
		})
	}

	var req CreateRoleRequest
	if err := c.BindJSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	if req.Name == "" || len(req.Name) < 3 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Role name is required and must be at least 3 characters",
		})
	}

	if req.Description == "" || len(req.Description) < 10 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Description is required and must be at least 10 characters",
		})
	}

	svc := services.NewRoleManagementService(config.DB)
	role, err := svc.CreateOrganizationRole(organizationID, req.Name, req.Description)
	if err != nil {
		log.Printf("Error creating role: %v", err)
		return utils.SendInternalError(c, "Failed to create role", err)
	}

	response := RoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
		IsDefault:   role.IsDefault,
		IsActive:    role.IsActive,
		CreatedAt:   role.CreatedAt.String(),
		UpdatedAt:   role.UpdatedAt.String(),
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    response,
	})
}

// UpdateOrganizationRole updates an existing role
func UpdateOrganizationRole(c fiber.Ctx) error {
	roleID := c.Params("roleId")
	if roleID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Role ID is required",
		})
	}

	var req UpdateRoleRequest
	if err := c.BindJSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	if req.Name != "" && len(req.Name) < 3 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Role name must be at least 3 characters",
		})
	}

	if req.Description != "" && len(req.Description) < 10 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Description must be at least 10 characters",
		})
	}

	svc := services.NewRoleManagementService(config.DB)
	role, err := svc.UpdateOrganizationRole(roleID, req.Name, req.Description)
	if err != nil {
		log.Printf("Error updating role: %v", err)
		if err.Error() == "role not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Role not found",
			})
		}
		return utils.SendInternalError(c, "Failed to update role", err)
	}

	response := RoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
		IsDefault:   role.IsDefault,
		IsActive:    role.IsActive,
		CreatedAt:   role.CreatedAt.String(),
		UpdatedAt:   role.UpdatedAt.String(),
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    response,
	})
}

// DeleteOrganizationRole deletes a role
func DeleteOrganizationRole(c fiber.Ctx) error {
	roleID := c.Params("roleId")
	if roleID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Role ID is required",
		})
	}

	svc := services.NewRoleManagementService(config.DB)
	err := svc.DeleteOrganizationRole(roleID)
	if err != nil {
		log.Printf("Error deleting role: %v", err)
		if err.Error() == "role not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Role not found",
			})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Role deleted successfully",
	})
}

// GetRolePermissions retrieves all permissions assigned to a role
func GetRolePermissions(c fiber.Ctx) error {
	roleID := c.Params("roleId")
	if roleID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Role ID is required",
		})
	}

	svc := services.NewRoleManagementService(config.DB)
	permissions, err := svc.GetRolePermissions(roleID)
	if err != nil {
		log.Printf("Error getting role permissions: %v", err)
		return utils.SendInternalError(c, "Failed to fetch permissions", err)
	}

	responses := make([]PermissionResponse, 0, len(permissions))
	for _, perm := range permissions {
		responses = append(responses, PermissionResponse{
			ID:          perm.ID,
			Resource:    perm.Resource,
			Action:      perm.Action,
			Description: perm.Description,
			IsActive:    perm.IsActive,
		})
	}

	return utils.SendSuccess(c, fiber.StatusOK, responses, "Permissions retrieved successfully", nil)
}

// AssignPermissionToRole assigns a permission to a role
func AssignPermissionToRole(c fiber.Ctx) error {
	roleID := c.Params("roleId")
	permissionID := c.Params("permissionId")

	if roleID == "" || permissionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Role ID and Permission ID are required",
		})
	}

	svc := services.NewRoleManagementService(config.DB)
	assignment, err := svc.AssignPermissionToRole(roleID, permissionID)
	if err != nil {
		log.Printf("Error assigning permission: %v", err)
		if err.Error() == "role not found" || err.Error() == "permission not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		}
		return utils.SendInternalError(c, "Failed to assign permission", err)
	}

	response := PermissionAssignmentResponse{
		ID:                      assignment.ID,
		OrganizationRoleID:      assignment.OrganizationRoleID,
		OrganizationPermissionID: assignment.OrganizationPermissionID,
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    response,
	})
}

// RemovePermissionFromRole removes a permission from a role
func RemovePermissionFromRole(c fiber.Ctx) error {
	roleID := c.Params("roleId")
	permissionID := c.Params("permissionId")

	if roleID == "" || permissionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Role ID and Permission ID are required",
		})
	}

	svc := services.NewRoleManagementService(config.DB)
	err := svc.RemovePermissionFromRole(roleID, permissionID)
	if err != nil {
		log.Printf("Error removing permission: %v", err)
		return utils.SendInternalError(c, "Failed to remove permission", err)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Permission removed successfully",
	})
}

// GetOrganizationPermissions retrieves all available permissions for the organization
func GetOrganizationPermissions(c fiber.Ctx) error {
	organizationID, ok := c.Locals("organizationID").(string)
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Organization ID not found",
		})
	}

	svc := services.NewRoleManagementService(config.DB)
	permissions, err := svc.GetOrganizationPermissions(organizationID)
	if err != nil {
		log.Printf("Error getting organization permissions: %v", err)
		return utils.SendInternalError(c, "Failed to fetch permissions", err)
	}

	responses := make([]PermissionResponse, 0, len(permissions))
	for _, perm := range permissions {
		responses = append(responses, PermissionResponse{
			ID:          perm.ID,
			Resource:    perm.Resource,
			Action:      perm.Action,
			Description: perm.Description,
			IsActive:    perm.IsActive,
		})
	}

	return utils.SendSuccess(c, fiber.StatusOK, responses, "Permissions retrieved successfully", nil)
}

// PermissionResponse is the response format for permissions
type PermissionResponse struct {
	ID          string `json:"id"`
	Resource    string `json:"resource"`
	Action      string `json:"action"`
	Description string `json:"description"`
	IsActive    bool   `json:"isActive"`
}

// PermissionAssignmentResponse is the response format for permission assignments
type PermissionAssignmentResponse struct {
	ID                       string `json:"id"`
	OrganizationRoleID       string `json:"organizationRoleId"`
	OrganizationPermissionID string `json:"organizationPermissionId"`
}
