"use server";

import { APIResponse } from "@/types";
import authenticatedApiClient, {
  axios,
  badRequestResponse,
  handleError,
  successResponse,
} from "./api-config";

// ============================================================================
// ROLE PERMISSIONS MANAGEMENT
// ============================================================================

/**
 * Get all permissions for a specific role
 * Endpoint: GET /api/v1/roles/{id}/permissions
 * Status: ✅ Documented in API
 *
 * IMPORTANT: Only returns permissions for modules assigned to the role's department
 */
export async function getRolePermissions(roleId: string): Promise<APIResponse> {
  const url = `/api/v1/roles/${roleId}/permissions`;

  if (!roleId) {
    return badRequestResponse("Role ID is required");
  }

  try {
    const response = await authenticatedApiClient({ url });
    const permissions = response?.data?.data || [];

    return successResponse(
      permissions,
      "Role permissions fetched successfully"
    );
  } catch (error: Error | any) {
    return handleError(error, "GET", url);
  }
}

/**
 * Grant or update permission for a role on a specific module
 * Endpoint: POST /api/v1/roles/{id}/permissions
 * Status: ✅ Documented in API
 *
 * IMPORTANT: The module must be assigned to the role's department
 *
 * Permission Types:
 * - can_view: View access
 * - can_create: Create new items
 * - can_edit: Edit existing items
 * - can_delete: Delete items
 * - can_approve: Approve actions
 * - can_export: Export data
 * - can_assign: Assign tasks/items to others
 * - can_configure: Configure settings
 * - custom_permissions: JSONB field for module-specific permissions
 */
export async function grantOrUpdateRolePermission({
  roleId,
  moduleId,
  canView = false,
  canCreate = false,
  canEdit = false,
  canDelete = false,
  canApprove = false,
  canExport = false,
  canAssign = false,
  canConfigure = false,
  customPermissions,
  parentModuleId,
}: {
  roleId: string;
  moduleId: string;
  canView?: boolean;
  canCreate?: boolean;
  canEdit?: boolean;
  canDelete?: boolean;
  canApprove?: boolean;
  canExport?: boolean;
  canAssign?: boolean;
  canConfigure?: boolean;
  customPermissions?: Record<string, any>;
  parentModuleId?: string | null;
}): Promise<APIResponse> {
  const url = `/api/v1/roles/${roleId}/permissions`;

  if (!roleId || !moduleId) {
    return badRequestResponse("Role ID and Module ID are required");
  }

  try {
    const response = await authenticatedApiClient({
      url,
      method: "POST",
      data: {
        module_id: moduleId,
        parent_module_id: parentModuleId || null,
        can_view: canView,
        can_create: canCreate,
        can_edit: canEdit,
        can_delete: canDelete,
        can_approve: canApprove,
        can_export: canExport,
        can_assign: canAssign,
        can_configure: canConfigure,
        custom_permissions: customPermissions || {},
      },
    });

    return successResponse(response?.data, "Permission granted successfully");
  } catch (error: Error | any) {
    return handleError(error, "POST", url);
  }
}

/**
 * Revoke all permissions for a role on a specific module
 * Endpoint: DELETE /api/v1/roles/{role_id}/permissions/{module_id}
 * Status: ✅ Documented in API
 */
export async function revokeRolePermission({
  roleId,
  moduleId,
}: {
  roleId: string;
  moduleId: string;
}): Promise<APIResponse> {
  const url = `/api/v1/roles/${roleId}/permissions/${moduleId}`;

  if (!roleId || !moduleId) {
    return badRequestResponse("Role ID and Module ID are required");
  }

  try {
    // Fixed: Use authenticatedApiClient instead of axios.delete() to ensure proper authentication
    await authenticatedApiClient({ url, method: "DELETE" });
    return successResponse(null, "Permission revoked successfully");
  } catch (error: Error | any) {
    return handleError(error, "DELETE", url);
  }
}

/**
 * Get available modules that can be assigned permissions for a role
 * Endpoint: GET /api/v1/roles/{id}/available-modules
 * Status: ✅ Documented in API
 *
 * IMPORTANT: Only returns modules assigned to the role's department
 * This implements the department-constrained RBAC system
 */
export async function getAvailableModulesForRole(
  roleId: string
): Promise<APIResponse> {
  const url = `/api/v1/roles/${roleId}/available-modules`;

  if (!roleId) {
    return badRequestResponse("Role ID is required");
  }

  try {
    const response = await authenticatedApiClient({ url });
    return successResponse(
      response?.data,
      "Available modules fetched successfully"
    );
  } catch (error: Error | any) {
    return handleError(error, "GET", url);
  }
}

/**
 * Bulk update multiple role permissions at once
 * Helper function that wraps multiple grantOrUpdateRolePermission calls
 *
 * @param roleId - The role ID to update permissions for
 * @param permissions - Array of permission objects to update
 */
export async function bulkUpdateRolePermissions({
  roleId,
  permissions,
}: {
  roleId: string;
  permissions: Array<{
    moduleId: string;
    canView?: boolean;
    canCreate?: boolean;
    canEdit?: boolean;
    canDelete?: boolean;
    canApprove?: boolean;
    canExport?: boolean;
    canAssign?: boolean;
    canConfigure?: boolean;
    customPermissions?: Record<string, any>;
    parentModuleId?: string | null;
  }>;
}): Promise<APIResponse> {
  if (!roleId || !permissions || permissions.length === 0) {
    return badRequestResponse("Role ID and permissions array are required");
  }

  const results: Array<{ moduleId: string; success: boolean; error?: string }> =
    [];

  try {
    // Process all permissions sequentially to avoid race conditions
    for (const perm of permissions) {
      const result = await grantOrUpdateRolePermission({
        roleId,
        ...perm,
      });

      results.push({
        moduleId: perm.moduleId,
        success: result.success,
        error: result.success ? undefined : result.message,
      });
    }

    const successCount = results.filter((r) => r.success).length;
    const failureCount = results.filter((r) => !r.success).length;

    if (failureCount === 0) {
      return successResponse(
        { results, successCount, failureCount },
        `Successfully updated ${successCount} permissions`
      );
    } else {
      return {
        success: false,
        message: `Updated ${successCount} permissions, ${failureCount} failed`,
        data: { results, successCount, failureCount },
        status: 207, // Multi-Status
        statusText: "PARTIAL_SUCCESS",
      };
    }
  } catch (error: Error | any) {
    return handleError(
      error,
      "BULK_UPDATE",
      `/api/v1/roles/${roleId}/permissions`
    );
  }
}

// ============================================================================
// HELPER FUNCTIONS FOR PERMISSION MANAGEMENT
// ============================================================================

/**
 * Helper function to set all standard permissions at once
 *
 * @param roleId - The role ID
 * @param moduleId - The module ID
 * @param level - Permission level: 'none', 'view', 'edit', 'full'
 */
export async function setPermissionLevel({
  roleId,
  moduleId,
  level,
}: {
  roleId: string;
  moduleId: string;
  level: "none" | "view" | "edit" | "full";
}): Promise<APIResponse> {
  const permissionLevels = {
    none: {
      canView: false,
      canCreate: false,
      canEdit: false,
      canDelete: false,
      canApprove: false,
      canExport: false,
      canAssign: false,
      canConfigure: false,
    },
    view: {
      canView: true,
      canCreate: false,
      canEdit: false,
      canDelete: false,
      canApprove: false,
      canExport: true, // Allow export with view
      canAssign: false,
      canConfigure: false,
    },
    edit: {
      canView: true,
      canCreate: true,
      canEdit: true,
      canDelete: false,
      canApprove: false,
      canExport: true,
      canAssign: true, // Allow assignment with edit
      canConfigure: false,
    },
    full: {
      canView: true,
      canCreate: true,
      canEdit: true,
      canDelete: true,
      canApprove: true,
      canExport: true,
      canAssign: true,
      canConfigure: true,
    },
  };

  return grantOrUpdateRolePermission({
    roleId,
    moduleId,
    ...permissionLevels[level],
  });
}

/**
 * Copy permissions from one role to another
 * Useful for creating new roles based on existing ones
 *
 * @param sourceRoleId - Role to copy from
 * @param targetRoleId - Role to copy to
 */
export async function copyRolePermissions({
  sourceRoleId,
  targetRoleId,
}: {
  sourceRoleId: string;
  targetRoleId: string;
}): Promise<APIResponse> {
  if (!sourceRoleId || !targetRoleId) {
    return badRequestResponse("Source and target role IDs are required");
  }

  try {
    // Get permissions from source role
    const sourcePermissionsResult = await getRolePermissions(sourceRoleId);

    if (!sourcePermissionsResult.success || !sourcePermissionsResult.data) {
      return {
        success: false,
        message: "Failed to fetch source role permissions",
        data: null,
        status: 400,
        statusText: "BAD_REQUEST",
      };
    }

    const sourcePermissions = sourcePermissionsResult.data;

    // Map to bulk update format
    const permissionsToApply = sourcePermissions.map((perm: any) => ({
      moduleId: perm.module_id,
      canView: perm.can_view,
      canCreate: perm.can_create,
      canEdit: perm.can_edit,
      canDelete: perm.can_delete,
      canApprove: perm.can_approve,
      canExport: perm.can_export,
      canAssign: perm.can_assign,
      canConfigure: perm.can_configure,
      customPermissions: perm.custom_permissions,
    }));

    // Apply to target role
    return await bulkUpdateRolePermissions({
      roleId: targetRoleId,
      permissions: permissionsToApply,
    });
  } catch (error: Error | any) {
    return handleError(
      error,
      "COPY_PERMISSIONS",
      "/api/v1/roles/copy-permissions"
    );
  }
}

/**
 * Check if a role has a specific permission on a module
 *
 * @param roleId - The role ID
 * @param moduleId - The module ID
 * @param permission - The permission type to check
 */
export async function checkRolePermission({
  roleId,
  moduleId,
  permission,
}: {
  roleId: string;
  moduleId: string;
  permission:
    | "view"
    | "create"
    | "edit"
    | "delete"
    | "approve"
    | "export"
    | "assign"
    | "configure";
}): Promise<APIResponse> {
  try {
    const permissionsResult = await getRolePermissions(roleId);

    if (!permissionsResult.success || !permissionsResult.data) {
      return {
        success: false,
        message: "Failed to fetch role permissions",
        data: { hasPermission: false },
        status: 400,
        statusText: "BAD_REQUEST",
      };
    }

    const modulePermission = permissionsResult.data.find(
      (perm: any) => perm.module_id === moduleId
    );

    if (!modulePermission) {
      return successResponse(
        { hasPermission: false },
        "No permission found for module"
      );
    }

    const permissionMap: Record<string, string> = {
      view: "can_view",
      create: "can_create",
      edit: "can_edit",
      delete: "can_delete",
      approve: "can_approve",
      export: "can_export",
      assign: "can_assign",
      configure: "can_configure",
    };

    const hasPermission = modulePermission[permissionMap[permission]] === true;

    return successResponse(
      { hasPermission, modulePermission },
      hasPermission ? "Permission granted" : "Permission denied"
    );
  } catch (error: Error | any) {
    return handleError(
      error,
      "CHECK_PERMISSION",
      `/api/v1/roles/${roleId}/permissions`
    );
  }
}
