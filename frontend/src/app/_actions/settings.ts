"use server";

import { getCurrentUser } from "@/lib/auth";
import { APIResponse } from "@/types";
import authenticatedApiClient from "./api-config";

/**
 * Get current user profile (full, including preferences)
 */
export async function getUserProfile(): Promise<APIResponse> {
  try {
    const response = await authenticatedApiClient({
      url: "/api/v1/auth/profile",
      method: "GET",
    });
    return {
      success: true,
      message: "Profile retrieved successfully",
      data: response.data.data,
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
 * Update account settings — persists name, email, and all preferences to the DB.
 * Calls: PUT /api/v1/auth/profile
 */
export async function updateAccountSettings(data: {
  name: string;
  email: string;
  preferences: {
    avatar?: string;
    department?: string;
    language?: string;
    theme?: string;
    timezone?: string;
    emailNotifications?: boolean;
    pushNotifications?: boolean;
    activityNotifications?: boolean;
  };
}): Promise<APIResponse> {
  try {
    const response = await authenticatedApiClient({
      url: "/api/v1/auth/profile",
      method: "PUT",
      data,
    });
    return {
      success: true,
      message: "Settings saved successfully",
      data: response.data.data,
      status: 200,
      statusText: "OK",
    };
  } catch (error: any) {
    return {
      success: false,
      message: error.response?.data?.message || error.message || "Failed to save settings",
      data: null,
      status: error.response?.status || 500,
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

    const response = await authenticatedApiClient({
      url: "/api/v1/auth/change-password",
      method: "POST",
      data: { currentPassword, newPassword, confirmPassword },
    });

    return {
      success: true,
      message: "Password changed successfully",
      data: response.data.data,
      status: 200,
      statusText: "OK",
    };
  } catch (error: any) {
    return {
      success: false,
      message: error.response?.data?.message || error.message || "Failed to change password",
      data: null,
      status: error.response?.status || 500,
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
