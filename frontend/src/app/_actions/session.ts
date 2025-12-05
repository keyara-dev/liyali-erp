'use server';

import { getCurrentUser, hasRole, isAdmin } from '@/lib/auth';
import type { User } from '@/types/auth';

/**
 * Get the current authenticated user session
 * Server action that can be called from client components using React Query
 *
 * @returns {Promise<{user: User | null, isAuthenticated: boolean}>}
 *
 * @example
 * ```typescript
 * const { data: session } = useQuery({
 *   queryKey: ['session'],
 *   queryFn: () => getSessionAction(),
 * })
 * ```
 */
export async function getSessionAction(): Promise<{
  user: User | null;
  isAuthenticated: boolean;
}> {
  try {
    const user = await getCurrentUser();

    return {
      user,
      isAuthenticated: !!user,
    };
  } catch (error) {
    console.error('Failed to get session:', error);
    return {
      user: null,
      isAuthenticated: false,
    };
  }
}

/**
 * Check if current user has a specific role
 * Server action that can be called from client components
 *
 * @param {string | string[]} requiredRole - Role(s) to check for
 * @returns {Promise<boolean>}
 *
 * @example
 * ```typescript
 * const { data: isAdmin } = useQuery({
 *   queryKey: ['user-role', 'ADMIN'],
 *   queryFn: () => checkUserRoleAction('ADMIN'),
 * })
 * ```
 */
export async function checkUserRoleAction(
  requiredRole: string | string[]
): Promise<boolean> {
  try {
    return await hasRole(requiredRole as any);
  } catch (error) {
    console.error('Failed to check user role:', error);
    return false;
  }
}

/**
 * Check if current user is admin
 * Server action that can be called from client components
 *
 * @returns {Promise<boolean>}
 *
 * @example
 * ```typescript
 * const { data: isUserAdmin } = useQuery({
 *   queryKey: ['user-is-admin'],
 *   queryFn: () => checkIsAdminAction(),
 * })
 * ```
 */
export async function checkIsAdminAction(): Promise<boolean> {
  try {
    return await isAdmin();
  } catch (error) {
    console.error('Failed to check admin status:', error);
    return false;
  }
}
