/**
 * User role assignment store
 * Separated from server actions to avoid "use server" constraints
 */

import { UserRoleAssignment } from '@/types';

// In-memory store for user-role assignments (in production, this would be a database)
export const userRoleAssignmentsStore = new Map<string, UserRoleAssignment>();
