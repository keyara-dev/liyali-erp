'use server';

import { auth } from '@/auth';
import { APIResponse } from '@/types';
import { User, UserRole } from '@/types/workflow';
import { MOCK_USERS } from '@/lib/mock-data';
import { getCustomRole, getAllCustomRoles } from '@/lib/rbac';
import { handleError, unauthorizedResponse } from '@/app/_actions/api-config';

// In-memory store for user-role assignments (in production, this would be a database)
interface UserRoleAssignment {
  userId: string;
  customRoleId: string;
  assignedAt: Date;
  assignedBy: string;
}

const userRoleAssignmentsStore = new Map<string, UserRoleAssignment>();

// Helper to check if user is admin
function isAdmin(userRole: string | undefined): boolean {
  return userRole === 'ADMIN';
}

// =============== USER MANAGEMENT ===============

export async function getAllUsers(): Promise<APIResponse<User[]>> {
  const session = await auth();
  if (!session?.user) return unauthorizedResponse();

  try {
    const allUsers = Object.values(MOCK_USERS).flat();

    return {
      success: true,
      message: 'All users retrieved successfully',
      data: allUsers,
      status: 200,
      statusText: 'OK',
    };
  } catch (error) {
    console.error('Error fetching users:', error);
    return handleError(error, 'GET', '/users');
  }
}

export async function getUsersByRole(role: UserRole): Promise<APIResponse<User[]>> {
  const session = await auth();
  if (!session?.user) return unauthorizedResponse();

  try {
    const users = MOCK_USERS[role] || [];

    return {
      success: true,
      message: `Users with role ${role} retrieved successfully`,
      data: users,
      status: 200,
      statusText: 'OK',
    };
  } catch (error) {
    console.error('Error fetching users by role:', error);
    return handleError(error, 'GET', `/users?role=${role}`);
  }
}

export async function getUserById(userId: string): Promise<APIResponse<User>> {
  const session = await auth();
  if (!session?.user) return unauthorizedResponse();

  try {
    const user = Object.values(MOCK_USERS)
      .flat()
      .find((u) => u.id === userId);

    if (!user) {
      return {
        success: false,
        message: 'User not found',
        data: null,
        status: 404,
        statusText: 'NOT FOUND',
      };
    }

    return {
      success: true,
      message: 'User retrieved successfully',
      data: user,
      status: 200,
      statusText: 'OK',
    };
  } catch (error) {
    console.error('Error fetching user:', error);
    return handleError(error, 'GET', `/users/${userId}`);
  }
}

export async function searchUsers(query: string): Promise<APIResponse<User[]>> {
  const session = await auth();
  if (!session?.user) return unauthorizedResponse();

  try {
    const allUsers = Object.values(MOCK_USERS).flat();
    const searchTerm = query.toLowerCase();

    const results = allUsers.filter(
      (user) =>
        user.name.toLowerCase().includes(searchTerm) ||
        user.email.toLowerCase().includes(searchTerm) ||
        user.department?.toLowerCase().includes(searchTerm)
    );

    return {
      success: true,
      message: `Found ${results.length} user(s)`,
      data: results,
      status: 200,
      statusText: 'OK',
    };
  } catch (error) {
    console.error('Error searching users:', error);
    return handleError(error, 'GET', `/users/search?q=${query}`);
  }
}

// =============== USER-ROLE ASSIGNMENT ===============

export async function assignCustomRoleToUser(
  userId: string,
  customRoleId: string
): Promise<APIResponse<UserRoleAssignment>> {
  const session = await auth();
  if (!session?.user) return unauthorizedResponse();

  const userRole = (session.user as any).role;
  if (!isAdmin(userRole)) {
    return {
      success: false,
      message: 'You do not have permission to assign roles',
      data: null,
      status: 403,
      statusText: 'FORBIDDEN',
    };
  }

  try {
    // Verify user exists
    const user = Object.values(MOCK_USERS)
      .flat()
      .find((u) => u.id === userId);

    if (!user) {
      return {
        success: false,
        message: 'User not found',
        data: null,
        status: 404,
        statusText: 'NOT FOUND',
      };
    }

    // Verify role exists
    const role = getCustomRole(customRoleId);
    if (!role) {
      return {
        success: false,
        message: 'Role not found',
        data: null,
        status: 404,
        statusText: 'NOT FOUND',
      };
    }

    const assignmentId = `${userId}-${customRoleId}`;
    const assignment: UserRoleAssignment = {
      userId,
      customRoleId,
      assignedAt: new Date(),
      assignedBy: session.user.id,
    };

    userRoleAssignmentsStore.set(assignmentId, assignment);

    console.log(`✅ Role assigned to user: ${user.name} -> ${role.name}`);

    return {
      success: true,
      message: `Role ${role.name} assigned to ${user.name}`,
      data: assignment,
      status: 201,
      statusText: 'CREATED',
    };
  } catch (error) {
    console.error('Error assigning role:', error);
    return handleError(error, 'POST', `/users/${userId}/roles`);
  }
}

export async function removeCustomRoleFromUser(
  userId: string,
  customRoleId: string
): Promise<APIResponse> {
  const session = await auth();
  if (!session?.user) return unauthorizedResponse();

  const userRole = (session.user as any).role;
  if (!isAdmin(userRole)) {
    return {
      success: false,
      message: 'You do not have permission to remove roles',
      data: null,
      status: 403,
      statusText: 'FORBIDDEN',
    };
  }

  try {
    const assignmentId = `${userId}-${customRoleId}`;
    const assignment = userRoleAssignmentsStore.get(assignmentId);

    if (!assignment) {
      return {
        success: false,
        message: 'User role assignment not found',
        data: null,
        status: 404,
        statusText: 'NOT FOUND',
      };
    }

    userRoleAssignmentsStore.delete(assignmentId);

    const user = Object.values(MOCK_USERS)
      .flat()
      .find((u) => u.id === userId);
    const role = getCustomRole(customRoleId);

    console.log(
      `✅ Role removed from user: ${user?.name} -> ${role?.name}`
    );

    return {
      success: true,
      message: `Role removed from user`,
      data: null,
      status: 200,
      statusText: 'OK',
    };
  } catch (error) {
    console.error('Error removing role:', error);
    return handleError(error, 'DELETE', `/users/${userId}/roles/${customRoleId}`);
  }
}

export async function getUserCustomRoles(userId: string): Promise<APIResponse> {
  const session = await auth();
  if (!session?.user) return unauthorizedResponse();

  try {
    const user = Object.values(MOCK_USERS)
      .flat()
      .find((u) => u.id === userId);

    if (!user) {
      return {
        success: false,
        message: 'User not found',
        data: null,
        status: 404,
        statusText: 'NOT FOUND',
      };
    }

    // Get all assignments for this user
    const assignments = Array.from(userRoleAssignmentsStore.values()).filter(
      (a) => a.userId === userId
    );

    // Get the custom role details
    const roles = assignments
      .map((a) => getCustomRole(a.customRoleId))
      .filter((r) => r !== null);

    return {
      success: true,
      message: 'User custom roles retrieved successfully',
      data: roles,
      status: 200,
      statusText: 'OK',
    };
  } catch (error) {
    console.error('Error fetching user custom roles:', error);
    return handleError(error, 'GET', `/users/${userId}/custom-roles`);
  }
}

export async function getUsersWithRole(customRoleId: string): Promise<APIResponse<User[]>> {
  const session = await auth();
  if (!session?.user) return unauthorizedResponse();

  try {
    // Verify role exists
    const role = getCustomRole(customRoleId);
    if (!role) {
      return {
        success: false,
        message: 'Role not found',
        data: null,
        status: 404,
        statusText: 'NOT FOUND',
      };
    }

    // Get all assignments for this role
    const assignments = Array.from(userRoleAssignmentsStore.values()).filter(
      (a) => a.customRoleId === customRoleId
    );

    // Get user details
    const users = assignments
      .map((a) =>
        Object.values(MOCK_USERS)
          .flat()
          .find((u) => u.id === a.userId)
      )
      .filter((u): u is User => u !== undefined);

    return {
      success: true,
      message: `Found ${users.length} user(s) with role ${role.name}`,
      data: users,
      status: 200,
      statusText: 'OK',
    };
  } catch (error) {
    console.error('Error fetching users with role:', error);
    return handleError(error, 'GET', `/roles/${customRoleId}/users`);
  }
}

export async function bulkAssignRolesToUser(
  userId: string,
  customRoleIds: string[]
): Promise<APIResponse> {
  const session = await auth();
  if (!session?.user) return unauthorizedResponse();

  const userRole = (session.user as any).role;
  if (!isAdmin(userRole)) {
    return {
      success: false,
      message: 'You do not have permission to assign roles',
      data: null,
      status: 403,
      statusText: 'FORBIDDEN',
    };
  }

  try {
    const user = Object.values(MOCK_USERS)
      .flat()
      .find((u) => u.id === userId);

    if (!user) {
      return {
        success: false,
        message: 'User not found',
        data: null,
        status: 404,
        statusText: 'NOT FOUND',
      };
    }

    const assignedRoles: string[] = [];
    const failedRoles: { roleId: string; reason: string }[] = [];

    for (const customRoleId of customRoleIds) {
      const role = getCustomRole(customRoleId);
      if (!role) {
        failedRoles.push({ roleId: customRoleId, reason: 'Role not found' });
        continue;
      }

      const assignmentId = `${userId}-${customRoleId}`;
      const assignment: UserRoleAssignment = {
        userId,
        customRoleId,
        assignedAt: new Date(),
        assignedBy: session.user.id,
      };

      userRoleAssignmentsStore.set(assignmentId, assignment);
      assignedRoles.push(role.name);
    }

    console.log(`✅ Bulk role assignment completed for ${user.name}`);

    return {
      success: failedRoles.length === 0,
      message:
        failedRoles.length === 0
          ? `All ${assignedRoles.length} role(s) assigned successfully`
          : `Assigned ${assignedRoles.length} role(s), ${failedRoles.length} failed`,
      data: {
        assignedRoles,
        failedRoles,
      },
      status: failedRoles.length === 0 ? 201 : 207,
      statusText: failedRoles.length === 0 ? 'CREATED' : 'MULTI_STATUS',
    };
  } catch (error) {
    console.error('Error bulk assigning roles:', error);
    return handleError(error, 'POST', `/users/${userId}/roles/bulk`);
  }
}

export async function getAvailableRolesForUser(
  userId: string
): Promise<APIResponse> {
  const session = await auth();
  if (!session?.user) return unauthorizedResponse();

  try {
    const user = Object.values(MOCK_USERS)
      .flat()
      .find((u) => u.id === userId);

    if (!user) {
      return {
        success: false,
        message: 'User not found',
        data: null,
        status: 404,
        statusText: 'NOT FOUND',
      };
    }

    // Get all custom roles
    const allRoles = getAllCustomRoles();

    // Get user's currently assigned roles
    const userAssignments = Array.from(userRoleAssignmentsStore.values()).filter(
      (a) => a.userId === userId
    );
    const assignedRoleIds = new Set(userAssignments.map((a) => a.customRoleId));

    // Return roles not yet assigned to the user
    const availableRoles = allRoles.filter((role) => !assignedRoleIds.has(role.id));

    return {
      success: true,
      message: 'Available roles retrieved successfully',
      data: {
        allRoles,
        assignedRoleIds: Array.from(assignedRoleIds),
        availableRoles,
      },
      status: 200,
      statusText: 'OK',
    };
  } catch (error) {
    console.error('Error fetching available roles:', error);
    return handleError(error, 'GET', `/users/${userId}/available-roles`);
  }
}
