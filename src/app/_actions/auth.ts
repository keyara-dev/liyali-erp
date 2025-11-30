'use server'

import { cache } from 'react'
import { getCurrentUser as getAuthUser, logout, isAdmin as checkAdmin } from '@/auth'
import { APIResponse } from '@/types'

/**
 * Get current authenticated user session
 */
export const getCurrentUser = cache(async (): Promise<APIResponse> => {
  try {
    const user = await getAuthUser()

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
      message: 'User session retrieved successfully',
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
})

/**
 * Sign out user
 */
export async function signOutAction(): Promise<APIResponse> {
  try {
    await logout()
    return {
      success: true,
      message: 'Signed out successfully',
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
 * Verify admin role
 */
export const verifyAdminRole = cache(async (): Promise<APIResponse> => {
  try {
    const user = await getAuthUser()

    if (!user) {
      return {
        success: false,
        message: 'Authentication required',
        data: null,
        status: 401,
        statusText: 'UNAUTHORIZED'
      }
    }

    const isAdminUser = await checkAdmin()

    if (!isAdminUser) {
      return {
        success: false,
        message: 'Admin access required',
        data: null,
        status: 403,
        statusText: 'FORBIDDEN'
      }
    }

    return {
      success: true,
      message: 'Admin access verified',
      data: {
        user,
        role: user.role
      },
      status: 200,
      statusText: 'OK'
    }
  } catch (error: any) {
    return {
      success: false,
      message: error.message || 'Verification failed',
      data: null,
      status: 500,
      statusText: 'ERROR'
    }
  }
})

/**
 * Send password reset email
 */
export async function sendResetEmail(email: string): Promise<APIResponse> {
  try {
    // This is a stub implementation for password reset
    // In a production system, this would send an actual email
    return {
      success: true,
      message: 'Password reset email sent successfully',
      data: { email },
      status: 200,
      statusText: 'OK'
    }
  } catch (error: any) {
    return {
      success: false,
      message: error.message || 'Failed to send reset email',
      data: null,
      status: 500,
      statusText: 'ERROR'
    }
  }
}

/**
 * Reset password with token
 */
export async function resetPassword(token: string, newPassword: string): Promise<APIResponse> {
  try {
    // This is a stub implementation for password reset
    // In a production system, this would validate the token and update the password
    return {
      success: true,
      message: 'Password reset successfully',
      data: { token },
      status: 200,
      statusText: 'OK'
    }
  } catch (error: any) {
    return {
      success: false,
      message: error.message || 'Failed to reset password',
      data: null,
      status: 500,
      statusText: 'ERROR'
    }
  }
}