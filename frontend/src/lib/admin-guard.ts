import { verifySession } from './auth'
import { redirect } from 'next/navigation'

/**
 * Verify that the user has admin privileges before rendering admin pages.
 * This should be called in server components for admin routes.
 *
 * Redirects to /login if not authenticated
 * Redirects to /unauthorized if not an admin
 *
 * @returns {Promise<{ userId: string; userRole: string; userName?: string }>} User session data
 */
export async function requireAdminRole() {
  const { session, isAuthenticated } = await verifySession()

  if (!isAuthenticated || !session?.user) {
    redirect('/login')
  }

  // Check for admin or superadmin role
  const isAdmin = ['admin', 'compliance_officer'].includes(
    session.user.role || ''
  )

  if (!isAdmin) {
    redirect('/unauthorized')
  }

  return {
    userId: session.user.id,
    userRole: session.user.role,
    userName: session.user.name || session.user.email,
  }
}

/**
 * Verify that the user has specific admin permission(s).
 * This is a stricter check than requireAdminRole.
 *
 * @param requiredPermission - Single permission string
 * @returns {Promise<void>} Throws redirect if not authorized
 */
export async function requireAdminPermission(requiredPermission: string) {
  const { session, isAuthenticated } = await verifySession()

  if (!isAuthenticated || !session?.user) {
    redirect('/login')
  }

  // Only super admins bypass permission checks
  if (session.user.role === 'admin') {
    return
  }

  // Check for the specific permission
  const hasPermission = session.user.permissions?.includes(requiredPermission)

  if (!hasPermission) {
    redirect('/unauthorized')
  }
}

/**
 * Verify that the user is authenticated (used for private routes).
 * Less strict than requireAdminRole - just checks authentication.
 *
 * @returns {Promise<{ userId: string; userRole: string }>} User session data
 */
export async function requireAuthentication() {
  const { session, isAuthenticated } = await verifySession()

  if (!isAuthenticated || !session?.user) {
    redirect('/login')
  }

  return {
    userId: session.user.id,
    userRole: session.user.role,
  }
}
