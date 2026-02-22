"use server";

import { getCurrentUser } from "@/lib/auth";
import { APIResponse } from "@/types";

/**
 * Get current user profile
 */
export async function getUserProfile(): Promise<APIResponse> {
  try {
    const user = await getCurrentUser();

    if (!user) {
      return {
        success: false,
        message: "User not authenticated",
        data: null,
        status: 401,
        statusText: "UNAUTHORIZED",
      };
    }

    return {
      success: true,
      message: "Profile retrieved successfully",
      data: user,
      status: 200,
      statusText: "OK",
    };
  } catch (error: any) {
    return {
      success: false,
      message: error.message || "Failed to retrieve profile",
      data: null,
      status: 500,
      statusText: "ERROR",
    };
  }
}

/**
 * Update user profile
 *
 * NOTE: Currently, the backend User model does not have an 'avatar' field.
 * This function will need to be updated once the backend supports avatar storage.
 * For now, avatar changes are not persisted to the database.
 */
export async function updateUserProfile(profileData: {
  name?: string;
  email?: string;
  department?: string;
  avatar?: string;
}): Promise<APIResponse> {
  try {
    const user = await getCurrentUser();

    if (!user) {
      return {
        success: false,
        message: "User not authenticated",
        data: null,
        status: 401,
        statusText: "UNAUTHORIZED",
      };
    }

    // TODO: Call backend API to update user profile
    // The backend needs to:
    // 1. Add 'avatar' field to User model (backend/models/models.go)
    // 2. Add 'department' field to User model if not exists
    // 3. Create PUT /api/v1/users/:id endpoint
    // 4. Implement UpdateUser handler in backend/handlers/

    // For now, this is a mock implementation
    // Avatar changes will not persist across sessions
    const updatedUser = {
      ...user,
      ...profileData,
      updatedAt: new Date().toISOString(),
    };

    return {
      success: true,
      message:
        "Profile updated successfully (Note: Avatar not persisted - backend support needed)",
      data: updatedUser,
      status: 200,
      statusText: "OK",
    };
  } catch (error: any) {
    return {
      success: false,
      message: error.message || "Failed to update profile",
      data: null,
      status: 500,
      statusText: "ERROR",
    };
  }
}

/**
 * Change user password
 */
export async function changePassword(
  currentPassword: string,
  newPassword: string,
  confirmPassword: string,
): Promise<APIResponse> {
  try {
    const user = await getCurrentUser();

    if (!user) {
      return {
        success: false,
        message: "User not authenticated",
        data: null,
        status: 401,
        statusText: "UNAUTHORIZED",
      };
    }

    // Validate passwords
    if (newPassword !== confirmPassword) {
      return {
        success: false,
        message: "Passwords do not match",
        data: null,
        status: 400,
        statusText: "BAD_REQUEST",
      };
    }

    if (newPassword.length < 8) {
      return {
        success: false,
        message: "Password must be at least 8 characters long",
        data: null,
        status: 400,
        statusText: "BAD_REQUEST",
      };
    }

    if (currentPassword === newPassword) {
      return {
        success: false,
        message: "New password must be different from current password",
        data: null,
        status: 400,
        statusText: "BAD_REQUEST",
      };
    }

    // Mock implementation - in production, this would:
    // 1. Verify current password against stored hash
    // 2. Hash new password
    // 3. Update in database
    // 4. Invalidate all sessions

    return {
      success: true,
      message: "Password changed successfully",
      data: {
        changedAt: new Date().toISOString(),
        nextChangeDate: new Date(
          Date.now() + 90 * 24 * 60 * 60 * 1000,
        ).toISOString(),
      },
      status: 200,
      statusText: "OK",
    };
  } catch (error: any) {
    return {
      success: false,
      message: error.message || "Failed to change password",
      data: null,
      status: 500,
      statusText: "ERROR",
    };
  }
}

/**
 * Update general settings
 */
export async function updateGeneralSettings(settings: {
  language?: string;
  theme?: "light" | "dark" | "system";
  timezone?: string;
  emailNotifications?: boolean;
  pushNotifications?: boolean;
  activityNotifications?: boolean;
}): Promise<APIResponse> {
  try {
    const user = await getCurrentUser();

    if (!user) {
      return {
        success: false,
        message: "User not authenticated",
        data: null,
        status: 401,
        statusText: "UNAUTHORIZED",
      };
    }

    // Mock implementation - in production, this would update user preferences in database
    return {
      success: true,
      message: "General settings updated successfully",
      data: {
        settings,
        updatedAt: new Date().toISOString(),
      },
      status: 200,
      statusText: "OK",
    };
  } catch (error: any) {
    return {
      success: false,
      message: error.message || "Failed to update settings",
      data: null,
      status: 500,
      statusText: "ERROR",
    };
  }
}

/**
 * Get user sessions (active login sessions)
 */
export async function getUserSessions(): Promise<APIResponse> {
  try {
    const user = await getCurrentUser();

    if (!user) {
      return {
        success: false,
        message: "User not authenticated",
        data: null,
        status: 401,
        statusText: "UNAUTHORIZED",
      };
    }

    // Mock implementation - in production, this would fetch active sessions from database
    const sessions = [
      {
        id: "1",
        device: "Chrome on Windows",
        location: "Lusaka, ZM",
        ipAddress: "192.168.1.100",
        lastActive: new Date(Date.now() - 5 * 60 * 1000).toISOString(),
        createdAt: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString(),
        isCurrent: true,
      },
    ];

    return {
      success: true,
      message: "Sessions retrieved successfully",
      data: sessions,
      status: 200,
      statusText: "OK",
    };
  } catch (error: any) {
    return {
      success: false,
      message: error.message || "Failed to retrieve sessions",
      data: null,
      status: 500,
      statusText: "ERROR",
    };
  }
}

/**
 * Revoke a specific session
 */
export async function revokeSession(sessionId: string): Promise<APIResponse> {
  try {
    const user = await getCurrentUser();

    if (!user) {
      return {
        success: false,
        message: "User not authenticated",
        data: null,
        status: 401,
        statusText: "UNAUTHORIZED",
      };
    }

    // Mock implementation - in production, this would delete session from database
    return {
      success: true,
      message: "Session revoked successfully",
      data: { sessionId },
      status: 200,
      statusText: "OK",
    };
  } catch (error: any) {
    return {
      success: false,
      message: error.message || "Failed to revoke session",
      data: null,
      status: 500,
      statusText: "ERROR",
    };
  }
}
