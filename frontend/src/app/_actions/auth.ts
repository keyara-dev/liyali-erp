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
import authenticatedApiClient, {
  axios,
  handleError,
  successResponse,
  unauthorizedResponse,
  badRequestResponse,
} from "./api-config";



/**
 * Login with email and password using backend API
 */
export async function loginAction(
  email: string,
  password: string
): Promise<APIResponse<any>> {
  const url = `/api/v1/auth/login`;

  try {
    const query = await axios.post(url, {
      email,
      password,
    });

    const response = query?.data;

    // Backend returns: { success, message, data: { accessToken, refreshToken, expiresIn, user, organization } }
    if (!response.success || !response.data?.accessToken) {
      return unauthorizedResponse(response.message || "Login failed");
    }

    // Create session with backend token and expiration
    await createAuthSession({
      access_token: response.data.accessToken,
      refresh_token: response.data.refreshToken,
      role: response.data.user.role,
      user_id: response.data.user.id,
      organization_id: response.data.organization?.id,
      expiresIn: response.data.expiresIn, // Use backend's expiration time
      user: response.data.user, // Add the full user object
    });

    return successResponse(response.data.user, response.message);
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

    if (!session?.refresh_token) {
      return unauthorizedResponse("No refresh token available");
    }

    // Call backend refresh endpoint with the stored refresh token
    const response = await authenticatedApiClient( {url, method: "POST",
      data: {refreshToken: session.refresh_token}, // Use stored refresh token
    });

    // Backend returns: { success, message, data: { accessToken, expiresIn } }
    const newToken = response.data.data?.accessToken;
    const expiresIn = response.data.data?.expiresIn;

    if (!newToken) {
      return unauthorizedResponse("Failed to refresh token");
    }

    // Calculate expiration time using backend's expiresIn value
    const expirationMs = expiresIn ? expiresIn * 1000 : 30 * 60 * 1000; // fallback to 30 minutes
    
    // Update session with new access token (keep existing refresh token)
    await updateAuthSession({
      access_token: newToken,
      expiresAt: new Date(Date.now() + expirationMs), // Use backend's expiration time
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
    await authenticatedApiClient({url, method: "POST", data:{
      currentPassword: oldPassword, // Match backend parameter name
      newPassword: newPassword,
    }});

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

    // Backend returns: { success, message, data: { token, user, organization } }
    if (!responseData.success || !responseData.data?.token) {
      return unauthorizedResponse(
        responseData.message || "Registration failed"
      );
    }

    // Create session with token AND org context
    await createAuthSession({
      access_token: responseData.data.accessToken,
      refresh_token: responseData.data.refreshToken,
      role: responseData.data.user.role,
      user_id: responseData.data.user.id,
      organization_id: responseData.data.organization?.id,
      expiresIn: responseData.data.expiresIn, // Use backend's expiration time
      user: responseData.data.user, // Add the full user object
    });

    return successResponse(
      {
        user: responseData.data.user,
        organization: responseData.data.organization,
      },
      responseData.message
    );
  } catch (error: any) {
    return handleError(error, "POST", url);
  }
}

/**
 * Check if user signup is available/enabled
 * This can be used to control registration availability
 */
export async function checkSignupAvailability(): Promise<APIResponse<{ enabled: boolean }>> {
  try {
    // For now, always allow signups
    // In the future, this could check backend settings or environment variables
    return successResponse({ enabled: true }, "Signup availability checked");
  } catch (error: any) {
    return handleError(error, "GET", "/api/v1/auth/signup-availability");
  }
}

// ✅ Server action wrapper for getting server session
export const getServerSession = async () => await verifySession();
