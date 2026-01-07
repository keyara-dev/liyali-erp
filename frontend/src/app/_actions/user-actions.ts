"use server";

import { revalidatePath } from "next/cache";
import type { APIResponse } from "@/types";
import authenticatedApiClient, {
  handleError,
  successResponse,
} from "./api-config";
import { User, UserType } from "@/types";

// Types for user operations
export interface CreateUserRequest {
  username: string;
  email: string;
  phone?: string;
  password: string;
  first_name: string;
  last_name: string;
  branch_id: string;
  department_id: string;
  role_id: string;
  role?: UserType;
}

export interface UpdateUserRequest {
  id: string;
  username?: string;
  email?: string;
  phone?: string;
  first_name?: string;
  last_name?: string;
  branch_id?: string;
  department_id?: string;
  role_id?: string;
  is_active?: boolean;
}

export async function createNewUser(data: CreateUserRequest): Promise<APIResponse> {
  const url = `/api/v1/users`;

  try {
    const response = await authenticatedApiClient({
      url: url,
      data: {
        username: data.username,
        email: data.email,
        password: data.password,
        first_name: data.first_name,
        last_name: data.last_name,
        branch_id: data.branch_id,
        department_id: data.department_id,
        role_id: data.role_id,
      },
      method: "POST",
    });
    revalidatePath("/dashboard/system-configs/users");

    return successResponse(response?.data, "User registered successfully");
  } catch (error: Error | any) {
    return handleError(error, "POST", url);
  }
}

export async function getUsers(params?: {
  branchId?: string;
  departmentId?: string;
  roleId?: string;
  isActive?: boolean;
  isLdapUser?: boolean;
  search?: string;
  role?: string;
  page?: number;
  page_size?: number;
}): Promise<APIResponse> {
  const queryParams = new URLSearchParams();

  if (params?.branchId) queryParams.append("branch_id", params.branchId);
  if (params?.departmentId)
    queryParams.append("department_id", params.departmentId);
  if (params?.roleId) queryParams.append("role_id", params.roleId);
  if (params?.isActive !== undefined)
    queryParams.append("is_active", String(params.isActive));
  if (params?.isLdapUser !== undefined)
    queryParams.append("is_ldap_user", String(params.isLdapUser));
  if (params?.search) queryParams.append("search", params.search);
  if (params?.role) queryParams.append("role", params.role);
  if (params?.page) queryParams.append("page", String(params.page));
  if (params?.page_size)
    queryParams.append("page_size", String(params.page_size));

  const url = `/api/v1/users${queryParams.toString() ? `?${queryParams.toString()}` : ""}`;

  try {
    const response = await authenticatedApiClient({ url: url, method: "GET" });
    return successResponse(response.data, "Users fetched successfully");
  } catch (error) {
    return handleError(error, "GET", url);
  }
}
export async function getHeadsOfDepartments(params?: {
  department_id?: string;
  role_id?: string;
  is_active?: boolean;
  page?: number;
  page_size?: number;
}): Promise<APIResponse> {
  const queryParams = new URLSearchParams();

  if (params?.department_id)
    queryParams.append("department_id", params.department_id);
  if (params?.role_id) queryParams.append("role_id", params.role_id);
  if (params?.is_active !== undefined)
    queryParams.append("is_active", String(params.is_active));
  if (params?.page) queryParams.append("page", String(params.page));
  if (params?.page_size)
    queryParams.append("page_size", String(params.page_size));

  const url = `/api/v1/users/department-heads/list${queryParams.toString() ? `?${queryParams.toString()}` : ""}`;

  try {
    const response = await authenticatedApiClient({ url: url, method: "GET" });
    return successResponse(response.data.data, "HODs fetched successfully");
  } catch (error) {
    return handleError(error, "GET", url);
  }
}

export async function getDepartmentHeads(params?: {
  departmentId?: string;
}): Promise<APIResponse> {
  const queryParams = new URLSearchParams();
  if (params?.departmentId)
    queryParams.append("department_id", params.departmentId);
  const url = `/api/v1/users/department-heads/list${queryParams.toString() ? `?${queryParams.toString()}` : ""}`;
  try {
    const response = await authenticatedApiClient({ url: url, method: "GET" });
    return successResponse(response.data.data, "Users fetched successfully");
  } catch (error) {
    return handleError(error, "GET", url);
  }
}

export async function getUserById(id: string): Promise<APIResponse> {
  const url = `/api/v1/users/${id}`;

  try {
    const response = await authenticatedApiClient({ url: url, method: "GET" });
    return successResponse(
      response.data.data?.data || response.data,
      "User fetched successfully"
    );
  } catch (error) {
    return handleError(error, "GET", url);
  }
}

export async function updateUser(
  id: string,
  data: Partial<User>
): Promise<APIResponse> {
  const url = `/api/v1/users/${id}`;

  try {
    const response = await authenticatedApiClient({
      url: url,
      data: data,
      method: "PUT",
    });
    revalidatePath("/dashboard/system-configs/users");
    return successResponse(
      response.data.data || response.data,
      "User updated successfully"
    );
  } catch (error) {
    return handleError(error, "PUT", url);
  }
}

export async function deleteUser(id: string): Promise<APIResponse> {
  const url = `/api/v1/users/${id}`;

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "DELETE",
    });
    revalidatePath("/dashboard/system-configs/users");
    return successResponse(response.data.data, "User deleted successfully");
  } catch (error) {
    return handleError(error, "DELETE", url);
  }
}

/**
 * Toggle user active status
 */
export async function toggleUserStatus(
  id: string,
  isActive: boolean
): Promise<APIResponse> {
  try {
    // Fetch current user data first
    const userResponse = await getUserById(id);

    if (!userResponse.success || !userResponse.data) {
      return {
        success: false,
        message: "Failed to fetch user data",
        data: null,
        status: 400,
      };
    }

    const user = userResponse.data;

    // Update with complete user data plus the status change
    return updateUser(id, {
      username: user.username,
      email: user.email,
      first_name: user.first_name,
      last_name: user.last_name,
      department_id: user.department_id,
      is_active: isActive,
    });
  } catch (error) {
    return {
      success: false,
      message: "Failed to toggle user status",
      data: null,
      status: 500,
    };
  }
}

export async function deactivateUser(id: string): Promise<APIResponse> {
  const url = `/api/v1/users/${id}/deactivate`;
  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "PATCH",
    });
    revalidatePath("/dashboard/system-configs/users");
    return successResponse(response.data.data);
  } catch (error) {
    return handleError(error, "PATCH", url);
  }
}

export async function activateUser(id: string): Promise<APIResponse> {
  const url = `/api/v1/users/${id}/activate`;
  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "PATCH",
    });
    revalidatePath("/dashboard/system-configs/users");
    return successResponse(response.data.data);
  } catch (error) {
    return handleError(error, "PATCH", url);
  }
}

/**
 * Toggle user MFA
 */
export async function toggleUserMFA(
  id: string,
  enabled: boolean
): Promise<APIResponse> {
  try {
    // Fetch current user data first
    const userResponse = await getUserById(id);

    if (!userResponse.success || !userResponse.data) {
      return {
        success: false,
        message: "Failed to fetch user data",
        data: null,
        status: 400,
      };
    }

    const user = userResponse.data;

    // Update with complete user data plus the MFA change
    return updateUser(id, {
      username: user.username,
      email: user.email,
      first_name: user.first_name,
      last_name: user.last_name,
      department_id: user.department_id,
      mfa_enabled: enabled,
    });
  } catch (error) {
    return {
      success: false,
      message: "Failed to toggle user MFA",
      data: null,
      status: 500,
    };
  }
}

/**
 * Reset user password
 */
export async function resetUserPassword(
  id: string,
  password: string
): Promise<APIResponse> {
  const url = `/api/v1/users/${id}/reset-password`;
  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "POST",
      data: {
        new_password: password,
      },
    });
    return successResponse(response.data, "Password reset successfully");
  } catch (error) {
    return handleError(error, "POST", url);
  }
}
/**
 * Convenience function to get all users (wrapper around getUsers)
 */
export async function getAllUsers(): Promise<APIResponse> {
  return getUsers();
}

/**
 * Convenience function to get users by role (wrapper around getUsers)
 */
export async function getUsersByRole(role: string): Promise<APIResponse> {
  return getUsers({ role });
}

/**
 * Convenience function to search users (wrapper around getUsers)
 */
export async function searchUsers(query: string): Promise<APIResponse> {
  return getUsers({ search: query });
}