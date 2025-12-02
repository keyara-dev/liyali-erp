import 'server-only'

// Simulated Authentication - Replaces NextAuth
// This module provides auth functions without NextAuth dependency
// NOTE: Server-only module - can only be imported in Server Components and Server Actions

export {
  getSession as auth,
  getCurrentUser,
  login,
  logout,
  hasRole,
  isAdmin,
  getDemoUsers,
  type AuthUser,
  type UserRole
} from '@/lib/auth'

// Re-export auth actions for server components
export { requireAuth, requireRole } from '@/app/_actions/auth-actions'
