/**
 * Authentication and authorization helper functions
 * Separated from server actions to avoid "use server" constraints
 */

// Helper to check if user is admin
export function isAdmin(userRole: string | undefined): boolean {
  return userRole === 'ADMIN';
}
