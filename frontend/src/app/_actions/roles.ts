"use server";

import { apiCall } from "@/utils/api";

/**
 * Get all roles for the current organization
 */
export async function getRolesAction() {
  try {
    const response = await apiCall("/organization/roles", { method: "GET" });
    return response;
  } catch (error) {
    console.error("Error fetching roles:", error);
    throw error;
  }
}

/**
 * Create a new organization role
 */
export async function createRoleAction(name: string, description: string) {
  try {
    const response = await apiCall("/organization/roles", {
      method: "POST",
      body: { name, description },
    });
    return response;
  } catch (error) {
    console.error("Error creating role:", error);
    throw error;
  }
}

/**
 * Update an existing organization role
 */
export async function updateRoleAction(
  roleId: string,
  name: string,
  description: string
) {
  try {
    const response = await apiCall(`/organization/roles/${roleId}`, {
      method: "PUT",
      body: { name, description },
    });
    return response;
  } catch (error) {
    console.error("Error updating role:", error);
    throw error;
  }
}

/**
 * Delete an organization role
 * Note: System default roles cannot be deleted
 */
export async function deleteRoleAction(roleId: string) {
  try {
    const response = await apiCall(`/organization/roles/${roleId}`, {
      method: "DELETE",
    });
    return response;
  } catch (error) {
    console.error("Error deleting role:", error);
    throw error;
  }
}

/**
 * Get all permissions assigned to a specific role
 */
export async function getRolePermissionsAction(roleId: string) {
  try {
    const response = await apiCall(
      `/organization/roles/${roleId}/permissions`,
      {
        method: "GET",
      }
    );
    return response;
  } catch (error) {
    console.error("Error fetching role permissions:", error);
    throw error;
  }
}

/**
 * Get all available permissions for the organization
 */
export async function getAvailablePermissionsAction() {
  try {
    const response = await apiCall("/organization/permissions", {
      method: "GET",
    });
    return response;
  } catch (error) {
    console.error("Error fetching available permissions:", error);
    throw error;
  }
}

/**
 * Assign a permission to a role
 */
export async function assignPermissionAction(
  roleId: string,
  permissionId: string
) {
  try {
    const response = await apiCall(
      `/organization/roles/${roleId}/permissions/${permissionId}`,
      {
        method: "POST",
      }
    );
    return response;
  } catch (error) {
    console.error("Error assigning permission:", error);
    throw error;
  }
}

/**
 * Remove a permission from a role
 */
export async function removePermissionAction(
  roleId: string,
  permissionId: string
) {
  try {
    const response = await apiCall(
      `/organization/roles/${roleId}/permissions/${permissionId}`,
      {
        method: "DELETE",
      }
    );
    return response;
  } catch (error) {
    console.error("Error removing permission:", error);
    throw error;
  }
}
