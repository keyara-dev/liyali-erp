"use server";

import { redirect } from "next/navigation";
import {
  hasRole as checkRole,
  isAdmin as checkAdmin,
  setScreenLockCookie,
  clearScreenLockCookie,
  getScreenLockState,
  deleteSession,
  updateAuthSession,
  createAuthSession,
  getCurrentUser,
  verifySession,
} from "@/lib/auth";
import { cache } from "react";

import { APIResponse } from "@/types";
import { checkIsAdminAction } from "./session";
import {
  axios,
  handleError,
  successResponse,
  unauthorizedResponse,
  badRequestResponse,
} from "./api-config";

/**
 * Get current authenticated user
 */
export async function getCurrentUserAction(): Promise<APIResponse<any>> {
  const url = `/api/v1/auth/profile`;

  try {
    const { session } = await verifySession();

    if (!session?.access_token) {
      return unauthorizedResponse("No authenticated user found");
    }

    // Call backend to get current user profile
    const response = await axios.get(url, {
      headers: {
        Authorization: `Bearer ${session.access_token}`,
      },
    });

    return successResponse(response.data, "User retrieved successfully");
  } catch (error: any) {
    return handleError(error, "GET", url);
  }
}

/**
 * Login with email and password using backend API
 */
export async function loginAction(
  email: string,
  password: string
): Promise<APIResponse<any>> {
  const url = `/api/v1/auth/login`;

  try {
    const response = await axios.post(url, {
      email,
      password,
    });

    const data = response?.data;

    // Backend returns: { success, message, token, user }
    if (!data.success || !data.token) {
      return unauthorizedResponse(data.message || "Login failed");
    }

    // Create session with backend token
    await createAuthSession({
      access_token: data.token, // Backend uses "token" field
      role: data.user.role, // Map role to role
      user_id: data.user.id,
      organization_id: undefined, // Will be set when user switches org
    });

    return successResponse(data.user, data.message);
  } catch (error: Error | any) {
    return handleError(error, "POST", url);
  }
}

/**
 * Logout the current user
 */
export async function logoutAction(): Promise<APIResponse<null>> {
  try {
    await deleteSession();
    return successResponse(null, "Logged out successfully");
  } catch (error: any) {
    return handleError(error, "POST", "/api/v1/auth/logout");
  }
}

/**
 * Check if user has specific role
 */
export async function hasRoleAction(role: string | string[]): Promise<boolean> {
  try {
    return await checkRole(role as any);
  } catch {
    return false;
  }
}

/**
 * Check if user is admin
 */
export async function isAdminAction(): Promise<boolean> {
  try {
    return await checkAdmin();
  } catch {
    return false;
  }
}

/**
 * Require authentication - redirect to login if not authenticated
 */
export async function requireAuth() {
  const user = await getCurrentUser();
  if (!user) {
    redirect("/login");
  }
  return user;
}

/**
 * Require specific role - redirect to workflows if user doesn't have role
 */
export async function requireRole(allowedRoles: string[]) {
  const user = await getCurrentUser();
  if (!user) {
    redirect("/login");
  }
  if (!allowedRoles.includes(user.role)) {
    redirect("/home");
  }
  return user;
}

/**
 * Lock screen on user idle
 * Sets screen lock cookie when user becomes idle
 * @param isLocked - true to lock, false to unlock
 * @returns true if successful, false otherwise
 */
export async function lockScreenOnUserIdle(
  isLocked: boolean
): Promise<boolean> {
  try {
    const user = await getCurrentUser();
    if (!user) {
      return false;
    }

    if (isLocked) {
      await setScreenLockCookie(true);
    } else {
      await clearScreenLockCookie();
    }

    return true;
  } catch (error: any) {
    console.error("Error locking screen on idle:", error);
    return false;
  }
}

/**
 * Check screen lock state from cookie
 * Returns true if screen is locked, false otherwise
 */
export async function checkScreenLockState(): Promise<boolean> {
  try {
    return await getScreenLockState();
  } catch (error: any) {
    console.error("Error checking screen lock state:", error);
    return false;
  }
}

/**
 * Log user out due to session timeout or inactivity
 * Deletes all session cookies and clears auth state
 * @param reason - reason for logout (e.g., "Session expired")
 * @returns success response
 */
export async function logUserOut(
  reason: string = "User logged out"
): Promise<APIResponse<null>> {
  try {
    // Delete JWT sessions and screen lock state
    await deleteSession();
    return successResponse(null, reason);
  } catch (error: any) {
    console.error("Error logging out user:", error);
    return handleError(error, "POST", "/api/v1/auth/logout");
  }
}

/**
 * Refresh user token to extend session
 * Called when user confirms they're still active
 * @returns success response with token info
 */
export async function getRefreshToken(): Promise<APIResponse<any>> {
  const url = `/api/v1/auth/refresh`;

  try {
    const { session } = await verifySession();

    if (!session?.access_token) {
      return unauthorizedResponse("No active session");
    }

    // Call backend refresh endpoint
    const response = await axios.post(url, {
      token: session.access_token,
    });

    const newToken = response.data.token;

    // Update session with new token
    await updateAuthSession({
      access_token: newToken,
      expiresAt: new Date(Date.now() + 30 * 60 * 1000), // 30 minutes
    });

    return successResponse({ token: newToken }, "Token refreshed successfully");
  } catch (error: any) {
    console.error("Error refreshing token:", error);
    return handleError(error, "POST", url);
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
  const url = `/api/v1/auth/change-password`;

  try {
    await axios.post(url, {
      old_password: oldPassword,
      new_password: newPassword,
    });

    return successResponse(null, "Password changed successfully");
  } catch (error: any) {
    return handleError(error, "POST", url);
  }
}

/**
 * Verify admin role
 */
export const verifyAdminRole = cache(async (): Promise<APIResponse> => {
  try {
    const user = await getCurrentUser();

    if (!user) {
      return unauthorizedResponse("Authentication required");
    }

    const isAdminUser = await checkIsAdminAction();

    if (!isAdminUser) {
      return {
        success: false,
        message: "Admin access required",
        data: null,
        status: 403,
        statusText: "FORBIDDEN",
      } as APIResponse;
    }

    return successResponse(
      {
        user,
        role: user.role,
      },
      "Admin access verified"
    );
  } catch (error: any) {
    return handleError(error, "GET", "/api/v1/auth/admin/verify");
  }
});

/**
 * Send password reset email
 */
export async function sendResetEmail(email: string): Promise<APIResponse> {
  try {
    // This is a stub implementation for password reset
    // In a production system, this would send an actual email
    return successResponse({ email }, "Password reset email sent successfully");
  } catch (error: any) {
    return handleError(error, "POST", "/api/v1/auth/reset-password/send");
  }
}

/**
 * Reset password with token
 */
export async function resetPassword(
  token: string,
  newPassword: string
): Promise<APIResponse> {
  try {
    // This is a stub implementation for password reset
    // In a production system, this would validate the token and update the password
    return successResponse({ token }, "Password reset successfully");
  } catch (error: any) {
    return handleError(error, "POST", "/api/v1/auth/reset-password");
  }
}

/**
 * Create new user account
 */
export async function createNewAccount(data: {
  email: string;
  name: string;
  password: string;
  role?: string;
}): Promise<APIResponse<any>> {
  const url = `/api/v1/auth/register`;

  try {
    const response = await axios.post(url, {
      email: data.email,
      name: data.name,
      password: data.password,
      role: data.role || "requester",
    });

    const responseData = response?.data;

    if (!responseData.success || !responseData.token) {
      return unauthorizedResponse(
        responseData.message || "Registration failed"
      );
    }

    // Create session with token AND org context
    await createAuthSession({
      access_token: responseData.token,
      role: responseData.user.role,
      user_id: responseData.user.id,
      organization_id: responseData.organization?.id,
    });

    return successResponse(
      {
        user: responseData.user,
        organization: responseData.organization,
      },
      responseData.message
    );
  } catch (error: any) {
    return handleError(error, "POST", url);
  }
}

// ✅ Server action wrapper for getting server session
export const getServerSession = async () => await verifySession();
