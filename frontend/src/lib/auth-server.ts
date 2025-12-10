import 'server-only'

// Server-only authentication utilities
// This module can ONLY be imported in Server Components and Server Actions

export {
  getSession,
  getCurrentUser,
  login,
  hasRole,
  isAdmin,
  getDemoUsers,
  encrypt,
  decrypt,
  verifySession,
  deleteSession,
  updateAuthSession,
  setScreenLockCookie,
  getScreenLockState,
  clearScreenLockCookie,
  type AuthUser,
  type UserRole
} from '@/lib/auth'
