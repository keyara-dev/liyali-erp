"use server";

import type { APIResponse } from "@/types";
import authenticatedApiClient, { handleError, successResponse } from "./api-config";

export interface OrganizationRole {
  id: string;
  name: string;
  description?: string;
  permissions: string[];
  active: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface CreateRoleRequest {
  name: string;
  description?: string;
  permissions: string[];
}

export interface UpdateRoleRequest {
  id: string;
  name?: string;
  description?: string;
  permissions?: string[];
}

export interface Permission {
  id: string;
  name: string;
  category: string;
  description?: string;
}

/**
 * List organization roles
 * Calls: GET /api/v1/organization/roles
 */
export async function fetchOrganizationRoles(active?: boolean): Promise<APIResponse<OrganizationRole[]>> {
  const params = new URLSearchParams();
  if (active !== undefined) {
    params.set('active', active.toString());
  }

  const url = `/api/v1/organization/roles?${params.toString()}`;

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "GET"
    });

    return successResponse(response.data.data || [], "Roles retrieved successfully");
  } catch (error: any) {
    return handleError(error, "GET", url);
  }
}

/**
 * Create organization role
 * Calls: POST /api/v1/organization/roles
 */
export async function createOrganizationRole(data: CreateRoleRequest): Promise<APIResponse<OrganizationRole>> {
  const url = `/api/v1/organization/roles`;

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "POST",
      data: {
        name: data.name,
        description: data.description,
        permissions: data.permissions,
      }
    });

    return successResponse(response.data.data, "Role created successfully");
  } catch (error: any) {
    return handleError(error, "POST", url);
  }
}

/**
 * Update organization role
 * Calls: PUT /api/v1/organization/roles/{id}
 */
export async function updateOrganizationRole(data: UpdateRoleRequest): Promise<APIResponse<OrganizationRole>> {
  const url = `/api/v1/organization/roles/${data.id}`;

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "PUT",
      data: {
        name: data.name,
        description: data.description,
        permissions: data.permissions,
      }
    });

    return successResponse(response.data.data, "Role updated successfully");
  } catch (error: any) {
    return handleError(error, "PUT", url);
  }
}

/**
 * Delete organization role
 * Calls: DELETE /api/v1/organization/roles/{id}
 */
export async function deleteOrganizationRole(roleId: string): Promise<APIResponse> {
  const url = `/api/v1/organization/roles/${roleId}`;

  try {
    await authenticatedApiClient({
      url: url,
      method: "DELETE"
    });

    return successResponse(null, "Role deleted successfully");
  } catch (error: any) {
    return handleError(error, "DELETE", url);
  }
}

/**
 * Get role permissions
 * Calls: GET /api/v1/organization/roles/{id}/permissions
 */
export async function getRolePermissions(roleId: string): Promise<APIResponse<string[]>> {
  const url = `/api/v1/organization/roles/${roleId}/permissions`;

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "GET"
    });

    return successResponse(response.data.data || [], "Role permissions retrieved successfully");
  } catch (error: any) {
    return handleError(error, "GET", url);
  }
}

/**
 * Assign permission to role
 * Calls: POST /api/v1/organization/roles/{roleId}/permissions/{permissionId}
 */
export async function assignPermissionToRole(roleId: string, permissionId: string): Promise<APIResponse> {
  const url = `/api/v1/organization/roles/${roleId}/permissions/${permissionId}`;

  try {
    await authenticatedApiClient({
      url: url,
      method: "POST"
    });

    return successResponse(null, "Permission assigned successfully");
  } catch (error: any) {
    return handleError(error, "POST", url);
  }
}

/**
 * Remove permission from role
 * Calls: DELETE /api/v1/organization/roles/{roleId}/permissions/{permissionId}
 */
export async function removePermissionFromRole(roleId: string, permissionId: string): Promise<APIResponse> {
  const url = `/api/v1/organization/roles/${roleId}/permissions/${permissionId}`;

  try {
    await authenticatedApiClient({
      url: url,
      method: "DELETE"
    });

    return successResponse(null, "Permission removed successfully");
  } catch (error: any) {
    return handleError(error, "DELETE", url);
  }
}

/**
 * List available permissions
 * Calls: GET /api/v1/organization/permissions
 */
export async function fetchAvailablePermissions(category?: string): Promise<APIResponse<Permission[]>> {
  const params = new URLSearchParams();
  if (category) {
    params.set('category', category);
  }

  const url = `/api/v1/organization/permissions?${params.toString()}`;

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "GET"
    });

    return successResponse(response.data.data || [], "Permissions retrieved successfully");
  } catch (error: any) {
    return handleError(error, "GET", url);
  }
}