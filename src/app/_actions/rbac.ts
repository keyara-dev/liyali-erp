'use server';

import { auth } from '@/auth';
import { APIResponse } from '@/types';
import { Permission, UserRole } from '@/types/workflow';
import {
  createCustomRole,
  updateCustomRole,
  deleteCustomRole,
  getCustomRole,
  getAllCustomRoles,
  getCustomRolesByIsBuiltIn,
  updateCustomRolePermissions,
  addPermissionToCustomRole,
  removePermissionFromCustomRole,
  ALL_PERMISSIONS,
  PERMISSION_DESCRIPTIONS,
  CustomRole,
} from '@/lib/rbac';
import { handleError, unauthorizedResponse } from '@/app/_actions/api-config';

// Helper to check if user is admin
function isAdmin(userRole: string | undefined): boolean {
  return userRole === 'ADMIN';
}

// =============== ROLE MANAGEMENT ===============

export async function getAllRoles(): Promise<APIResponse<CustomRole[]>> {
  const session = await auth();
  if (!session?.user) return unauthorizedResponse();

  try {
    const roles = getAllCustomRoles();

    return {
      success: true,
      message: 'All roles retrieved successfully',
      data: roles,
      status: 200,
      statusText: 'OK',
    };
  } catch (error) {
    console.error('Error fetching roles:', error);
    return handleError(error, 'GET', '/rbac/roles');
  }
}

export async function getBuiltInRoles(): Promise<APIResponse<CustomRole[]>> {
  const session = await auth();
  if (!session?.user) return unauthorizedResponse();

  try {
    const roles = getCustomRolesByIsBuiltIn(true);

    return {
      success: true,
      message: 'Built-in roles retrieved successfully',
      data: roles,
      status: 200,
      statusText: 'OK',
    };
  } catch (error) {
    console.error('Error fetching built-in roles:', error);
    return handleError(error, 'GET', '/rbac/roles/builtin');
  }
}

export async function getCustomRoles(): Promise<APIResponse<CustomRole[]>> {
  const session = await auth();
  if (!session?.user) return unauthorizedResponse();

  try {
    const roles = getCustomRolesByIsBuiltIn(false);

    return {
      success: true,
      message: 'Custom roles retrieved successfully',
      data: roles,
      status: 200,
      statusText: 'OK',
    };
  } catch (error) {
    console.error('Error fetching custom roles:', error);
    return handleError(error, 'GET', '/rbac/roles/custom');
  }
}

export async function getRoleById(roleId: string): Promise<APIResponse<CustomRole>> {
  const session = await auth();
  if (!session?.user) return unauthorizedResponse();

  try {
    const role = getCustomRole(roleId);

    if (!role) {
      return {
        success: false,
        message: 'Role not found',
        data: null,
        status: 404,
        statusText: 'NOT FOUND',
      };
    }

    return {
      success: true,
      message: 'Role retrieved successfully',
      data: role,
      status: 200,
      statusText: 'OK',
    };
  } catch (error) {
    console.error('Error fetching role:', error);
    return handleError(error, 'GET', `/rbac/roles/${roleId}`);
  }
}

export async function createRole(
  name: string,
  description: string,
  permissions: Permission[]
): Promise<APIResponse<CustomRole>> {
  const session = await auth();
  if (!session?.user) return unauthorizedResponse();

  // Check if user has permission to manage workflows (typically admin)
  const userRole = (session.user as any).role;
  if (!isAdmin(userRole)) {
    return {
      success: false,
      message: 'You do not have permission to create roles',
      data: null,
      status: 403,
      statusText: 'FORBIDDEN',
    };
  }

  try {
    // Validate input
    if (!name || name.trim().length === 0) {
      return {
        success: false,
        message: 'Role name is required',
        data: null,
        status: 400,
        statusText: 'BAD REQUEST',
      };
    }

    if (!Array.isArray(permissions) || permissions.length === 0) {
      return {
        success: false,
        message: 'At least one permission is required',
        data: null,
        status: 400,
        statusText: 'BAD REQUEST',
      };
    }

    // Validate permissions
    const invalidPermissions = permissions.filter((p) => !ALL_PERMISSIONS.includes(p));
    if (invalidPermissions.length > 0) {
      return {
        success: false,
        message: `Invalid permissions: ${invalidPermissions.join(', ')}`,
        data: null,
        status: 400,
        statusText: 'BAD REQUEST',
      };
    }

    const newRole = createCustomRole(name, description || '', permissions);

    console.log(`✅ Custom role created: ${name}`);

    return {
      success: true,
      message: 'Role created successfully',
      data: newRole,
      status: 201,
      statusText: 'CREATED',
    };
  } catch (error) {
    console.error('Error creating role:', error);
    return handleError(error, 'POST', '/rbac/roles');
  }
}

export async function updateRole(
  roleId: string,
  name?: string,
  description?: string
): Promise<APIResponse<CustomRole>> {
  const session = await auth();
  if (!session?.user) return unauthorizedResponse();

  const userRole = (session.user as any).role;
  if (!isAdmin(userRole)) {
    return {
      success: false,
      message: 'You do not have permission to update roles',
      data: null,
      status: 403,
      statusText: 'FORBIDDEN',
    };
  }

  try {
    const role = getCustomRole(roleId);

    if (!role) {
      return {
        success: false,
        message: 'Role not found',
        data: null,
        status: 404,
        statusText: 'NOT FOUND',
      };
    }

    if (role.isBuiltIn) {
      return {
        success: false,
        message: 'Built-in roles cannot be modified',
        data: null,
        status: 400,
        statusText: 'BAD REQUEST',
      };
    }

    const updates: any = {};
    if (name !== undefined) updates.name = name;
    if (description !== undefined) updates.description = description;

    const updatedRole = updateCustomRole(roleId, updates);

    if (!updatedRole) {
      return {
        success: false,
        message: 'Failed to update role',
        data: null,
        status: 500,
        statusText: 'INTERNAL SERVER ERROR',
      };
    }

    console.log(`✅ Role updated: ${name || roleId}`);

    return {
      success: true,
      message: 'Role updated successfully',
      data: updatedRole,
      status: 200,
      statusText: 'OK',
    };
  } catch (error) {
    console.error('Error updating role:', error);
    return handleError(error, 'PUT', `/rbac/roles/${roleId}`);
  }
}

export async function deleteRole(roleId: string): Promise<APIResponse> {
  const session = await auth();
  if (!session?.user) return unauthorizedResponse();

  const userRole = (session.user as any).role;
  if (!isAdmin(userRole)) {
    return {
      success: false,
      message: 'You do not have permission to delete roles',
      data: null,
      status: 403,
      statusText: 'FORBIDDEN',
    };
  }

  try {
    const role = getCustomRole(roleId);

    if (!role) {
      return {
        success: false,
        message: 'Role not found',
        data: null,
        status: 404,
        statusText: 'NOT FOUND',
      };
    }

    if (role.isBuiltIn) {
      return {
        success: false,
        message: 'Built-in roles cannot be deleted',
        data: null,
        status: 400,
        statusText: 'BAD REQUEST',
      };
    }

    const deleted = deleteCustomRole(roleId);

    if (!deleted) {
      return {
        success: false,
        message: 'Failed to delete role',
        data: null,
        status: 500,
        statusText: 'INTERNAL SERVER ERROR',
      };
    }

    console.log(`✅ Role deleted: ${role.name}`);

    return {
      success: true,
      message: 'Role deleted successfully',
      data: null,
      status: 200,
      statusText: 'OK',
    };
  } catch (error) {
    console.error('Error deleting role:', error);
    return handleError(error, 'DELETE', `/rbac/roles/${roleId}`);
  }
}

// =============== PERMISSION MANAGEMENT ===============

export async function getAllPermissions(): Promise<
  APIResponse<
    Array<{
      name: Permission;
      description: string;
    }>
  >
> {
  const session = await auth();
  if (!session?.user) return unauthorizedResponse();

  try {
    const permissions = ALL_PERMISSIONS.map((perm) => ({
      name: perm,
      description: PERMISSION_DESCRIPTIONS[perm],
    }));

    return {
      success: true,
      message: 'All permissions retrieved successfully',
      data: permissions,
      status: 200,
      statusText: 'OK',
    };
  } catch (error) {
    console.error('Error fetching permissions:', error);
    return handleError(error, 'GET', '/rbac/permissions');
  }
}

export async function addRolePermission(
  roleId: string,
  permission: Permission
): Promise<APIResponse<CustomRole>> {
  const session = await auth();
  if (!session?.user) return unauthorizedResponse();

  const userRole = (session.user as any).role;
  if (!isAdmin(userRole)) {
    return {
      success: false,
      message: 'You do not have permission to manage permissions',
      data: null,
      status: 403,
      statusText: 'FORBIDDEN',
    };
  }

  try {
    const role = getCustomRole(roleId);

    if (!role) {
      return {
        success: false,
        message: 'Role not found',
        data: null,
        status: 404,
        statusText: 'NOT FOUND',
      };
    }

    if (!ALL_PERMISSIONS.includes(permission)) {
      return {
        success: false,
        message: `Invalid permission: ${permission}`,
        data: null,
        status: 400,
        statusText: 'BAD REQUEST',
      };
    }

    const updatedRole = addPermissionToCustomRole(roleId, permission);

    if (!updatedRole) {
      return {
        success: false,
        message: 'Failed to add permission',
        data: null,
        status: 500,
        statusText: 'INTERNAL SERVER ERROR',
      };
    }

    console.log(`✅ Permission added to role ${role.name}: ${permission}`);

    return {
      success: true,
      message: 'Permission added successfully',
      data: updatedRole,
      status: 200,
      statusText: 'OK',
    };
  } catch (error) {
    console.error('Error adding permission:', error);
    return handleError(error, 'POST', `/rbac/roles/${roleId}/permissions`);
  }
}

export async function removeRolePermission(
  roleId: string,
  permission: Permission
): Promise<APIResponse<CustomRole>> {
  const session = await auth();
  if (!session?.user) return unauthorizedResponse();

  const userRole = (session.user as any).role;
  if (!isAdmin(userRole)) {
    return {
      success: false,
      message: 'You do not have permission to manage permissions',
      data: null,
      status: 403,
      statusText: 'FORBIDDEN',
    };
  }

  try {
    const role = getCustomRole(roleId);

    if (!role) {
      return {
        success: false,
        message: 'Role not found',
        data: null,
        status: 404,
        statusText: 'NOT FOUND',
      };
    }

    const updatedRole = removePermissionFromCustomRole(roleId, permission);

    if (!updatedRole) {
      return {
        success: false,
        message: 'Failed to remove permission',
        data: null,
        status: 500,
        statusText: 'INTERNAL SERVER ERROR',
      };
    }

    console.log(`✅ Permission removed from role ${role.name}: ${permission}`);

    return {
      success: true,
      message: 'Permission removed successfully',
      data: updatedRole,
      status: 200,
      statusText: 'OK',
    };
  } catch (error) {
    console.error('Error removing permission:', error);
    return handleError(error, 'DELETE', `/rbac/roles/${roleId}/permissions/${permission}`);
  }
}

export async function updateRolePermissions(
  roleId: string,
  permissions: Permission[]
): Promise<APIResponse<CustomRole>> {
  const session = await auth();
  if (!session?.user) return unauthorizedResponse();

  const userRole = (session.user as any).role;
  if (!isAdmin(userRole)) {
    return {
      success: false,
      message: 'You do not have permission to manage permissions',
      data: null,
      status: 403,
      statusText: 'FORBIDDEN',
    };
  }

  try {
    const role = getCustomRole(roleId);

    if (!role) {
      return {
        success: false,
        message: 'Role not found',
        data: null,
        status: 404,
        statusText: 'NOT FOUND',
      };
    }

    if (role.isBuiltIn) {
      return {
        success: false,
        message: 'Built-in role permissions cannot be modified',
        data: null,
        status: 400,
        statusText: 'BAD REQUEST',
      };
    }

    // Validate permissions
    const invalidPermissions = permissions.filter((p) => !ALL_PERMISSIONS.includes(p));
    if (invalidPermissions.length > 0) {
      return {
        success: false,
        message: `Invalid permissions: ${invalidPermissions.join(', ')}`,
        data: null,
        status: 400,
        statusText: 'BAD REQUEST',
      };
    }

    const updatedRole = updateCustomRolePermissions(roleId, permissions);

    if (!updatedRole) {
      return {
        success: false,
        message: 'Failed to update permissions',
        data: null,
        status: 500,
        statusText: 'INTERNAL SERVER ERROR',
      };
    }

    console.log(`✅ Role permissions updated: ${role.name}`);

    return {
      success: true,
      message: 'Role permissions updated successfully',
      data: updatedRole,
      status: 200,
      statusText: 'OK',
    };
  } catch (error) {
    console.error('Error updating role permissions:', error);
    return handleError(error, 'PUT', `/rbac/roles/${roleId}/permissions`);
  }
}
