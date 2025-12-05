'use server'

import { redirect } from 'next/navigation'
import {
  getSession,
  getCurrentUser,
  login as authLogin,
  logout as authLogout,
  hasRole as checkRole,
  isAdmin as checkAdmin,
  getDemoUsers,
  DEMO_USERS,
  setScreenLockCookie,
  clearScreenLockCookie,
  getScreenLockState,
  deleteSession,
  updateAuthSession
} from '@/lib/auth'
import { APIResponse } from '@/types'

/**
 * Get current authenticated user
 */
export async function getCurrentUserAction(): Promise<APIResponse<any>> {
  try {
    const user = await getCurrentUser()
    if (!user) {
      return {
        success: false,
        message: 'No authenticated user found',
        data: null,
        status: 401,
        statusText: 'UNAUTHORIZED'
      }
    }
    return {
      success: true,
      message: 'User retrieved successfully',
      data: user,
      status: 200,
      statusText: 'OK'
    }
  } catch (error: any) {
    return {
      success: false,
      message: error.message || 'Failed to get user',
      data: null,
      status: 500,
      statusText: 'ERROR'
    }
  }
}

/**
 * Login with email and password
 */
export async function loginAction(
  email: string,
  password: string
): Promise<APIResponse<any>> {
  try {
    const result = await authLogin(email, password)
    if (!result.success) {
      return {
        success: false,
        message: result.error || 'Login failed',
        data: null,
        status: 401,
        statusText: 'UNAUTHORIZED'
      }
    }
    return {
      success: true,
      message: 'Login successful',
      data: result.user,
      status: 200,
      statusText: 'OK'
    }
  } catch (error: any) {
    return {
      success: false,
      message: error.message || 'Login error',
      data: null,
      status: 500,
      statusText: 'ERROR'
    }
  }
}

/**
 * Logout the current user
 */
export async function logoutAction(): Promise<APIResponse<null>> {
  try {
    await authLogout()
    return {
      success: true,
      message: 'Logged out successfully',
      data: null,
      status: 200,
      statusText: 'OK'
    }
  } catch (error: any) {
    return {
      success: false,
      message: error.message || 'Logout failed',
      data: null,
      status: 500,
      statusText: 'ERROR'
    }
  }
}

/**
 * Check if user has specific role
 */
export async function hasRoleAction(role: string | string[]): Promise<boolean> {
  try {
    return await checkRole(role as any)
  } catch {
    return false
  }
}

/**
 * Check if user is admin
 */
export async function isAdminAction(): Promise<boolean> {
  try {
    return await checkAdmin()
  } catch {
    return false
  }
}

/**
 * Get demo users for login page
 */
export async function getDemoUsersAction(): Promise<APIResponse<any[] | null>> {
  try {
    const users = await getDemoUsers()
    return {
      success: true,
      message: 'Demo users retrieved',
      data: users,
      status: 200,
      statusText: 'OK'
    }
  } catch (error: any) {
    return {
      success: false,
      message: error.message || 'Failed to get demo users',
      data: null,
      status: 500,
      statusText: 'ERROR'
    }
  }
}

/**
 * Require authentication - redirect to login if not authenticated
 */
export async function requireAuth() {
  const user = await getCurrentUser()
  if (!user) {
    redirect('/login')
  }
  return user
}

/**
 * Require specific role - redirect to workflows if user doesn't have role
 */
export async function requireRole(allowedRoles: string[]) {
  const user = await getCurrentUser()
  if (!user) {
    redirect('/login')
  }
  if (!allowedRoles.includes(user.role)) {
    redirect('/home')
  }
  return user
}

/**
 * Lock screen on user idle
 * Sets screen lock cookie when user becomes idle
 * @param isLocked - true to lock, false to unlock
 * @returns true if successful, false otherwise
 */
export async function lockScreenOnUserIdle(isLocked: boolean): Promise<boolean> {
  try {
    const user = await getCurrentUser()
    if (!user) {
      return false
    }

    if (isLocked) {
      await setScreenLockCookie(true)
    } else {
      await clearScreenLockCookie()
    }

    return true
  } catch (error: any) {
    console.error('Error locking screen on idle:', error)
    return false
  }
}

/**
 * Check screen lock state from cookie
 * Returns true if screen is locked, false otherwise
 */
export async function checkScreenLockState(): Promise<boolean> {
  try {
    return await getScreenLockState()
  } catch (error: any) {
    console.error('Error checking screen lock state:', error)
    return false
  }
}

/**
 * Log user out due to session timeout or inactivity
 * Deletes all session cookies and clears auth state
 * @param reason - reason for logout (e.g., "Session expired")
 * @returns success response
 */
export async function logUserOut(reason: string = 'User logged out'): Promise<APIResponse<null>> {
  try {
    // Delete JWT sessions and screen lock state
    await deleteSession()

    // Clear simulated auth (if needed)
    await authLogout()

    return {
      success: true,
      message: reason,
      data: null,
      status: 200,
      statusText: 'OK'
    }
  } catch (error: any) {
    console.error('Error logging out user:', error)
    return {
      success: false,
      message: error.message || 'Logout failed',
      data: null,
      status: 500,
      statusText: 'ERROR'
    }
  }
}

/**
 * Refresh user token to extend session
 * Called when user confirms they're still active
 * @returns success response with token info
 */
export async function getRefreshToken(): Promise<APIResponse<any>> {
  try {
    const user = await getCurrentUser()
    if (!user) {
      return {
        success: false,
        message: 'No active session',
        data: null,
        status: 401,
        statusText: 'UNAUTHORIZED'
      }
    }

    // Update auth session to extend expiration
    const updatedSession = await updateAuthSession({
      expiresAt: new Date(Date.now() + 30 * 60 * 1000) // Extend 30 minutes
    })

    if (!updatedSession) {
      return {
        success: false,
        message: 'Failed to refresh token',
        data: null,
        status: 500,
        statusText: 'ERROR'
      }
    }

    return {
      success: true,
      message: 'Token refreshed successfully',
      data: {
        expiresAt: updatedSession.expiresAt,
        user_id: updatedSession.user_id
      },
      status: 200,
      statusText: 'OK'
    }
  } catch (error: any) {
    console.error('Error refreshing token:', error)
    return {
      success: false,
      message: error.message || 'Failed to refresh token',
      data: null,
      status: 500,
      statusText: 'ERROR'
    }
  }
}

/**
 * Change user password
 * @param oldPassword - Current password
 * @param newPassword - New password
 * @returns success response
 */
export async function changePassword(
  oldPassword: string,
  newPassword: string
): Promise<APIResponse<null>> {
  try {
    const user = await getCurrentUser()
    if (!user) {
      return {
        success: false,
        message: 'User not authenticated',
        data: null,
        status: 401,
        statusText: 'UNAUTHORIZED'
      }
    }

    // Note: For demo purposes, we just validate that oldPassword matches demo password
    // In production, this would validate against hashed password in database
    const userConfig = DEMO_USERS[user.email.toLowerCase()]

    if (!userConfig || userConfig.password !== oldPassword) {
      return {
        success: false,
        message: 'Current password is incorrect',
        data: null,
        status: 400,
        statusText: 'BAD_REQUEST'
      }
    }

    // In production, this would update the database
    // For now, we just return success
    return {
      success: true,
      message: 'Password changed successfully',
      data: null,
      status: 200,
      statusText: 'OK'
    }
  } catch (error: any) {
    console.error('Error changing password:', error)
    return {
      success: false,
      message: error.message || 'Failed to change password',
      data: null,
      status: 500,
      statusText: 'ERROR'
    }
  }
}
