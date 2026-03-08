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
  email: string;
  phone?: string;
  password: string;
  first_name: string;
  last_name: string;
  name?: string; // Computed full name
  department_id?: string;
  role: UserType;
  // Profile fields
  position?: string;
  manNumber?: string;
  nrcNumber?: string;
  contact?: string;
}

export interface UpdateUserRequest {
  id: string;
  username?: string;
  email?: string;
  phone?: string;
  first_name?: string;
  last_name?: string;
  department_id?: string;
  role_id?: string;
  is_active?: boolean;
}

export async function createNewUser(
  data: CreateUserRequest,
): Promise<APIResponse> {
  // Use the dedicated admin user creation endpoint that doesn't create personal organizations
  const url = `/api/v1/organization/users`;

  try {
    // Compute full name if not provided
    const fullName = data.name || `${data.first_name} ${data.last_name}`.trim();

    // Create the user directly in the current organization
    const response = await authenticatedApiClient({
      url: url,
      data: {
        name: fullName,
        first_name: data.first_name,
        last_name: data.last_name,
        email: data.email,
        password: data.password,
        role: data.role || "requester",
        department_id: data.department_id,
        // Profile fields
        phone: data.phone,
        position: data.position,
        manNumber: data.manNumber,
        nrcNumber: data.nrcNumber,
        contact: data.contact,
      },
      method: "POST",
    });

    if (!response.data?.success) {
      return handleError(
        new Error(response.data?.message || "Failed to create user"),
        "POST",
        url,
      );
    }

    console.log("User created successfully:", response.data);

    revalidatePath("/admin/users");

    return successResponse(response.data?.data, "User created successfully");
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

  if (params?.departmentId)
    queryParams.append("department_id", params.departmentId);
  if (params?.isActive !== undefined)
    queryParams.append("active", String(params.isActive));
  if (params?.search) queryParams.append("search", params.search);
  if (params?.role) queryParams.append("role", params.role);
  if (params?.page) queryParams.append("page", String(params.page));
  if (params?.page_size)
    queryParams.append("page_size", String(params.page_size));

  const url = `/api/v1/organization/members${queryParams.toString() ? `?${queryParams.toString()}` : ""}`;

  try {
    const response = await authenticatedApiClient({ url: url, method: "GET" });

    // Transform organization members data to match expected user format
    const members = response.data?.data || response.data || [];
    const transformedUsers = members.map((member: any) => {
      // Handle both JSON field names (userId) and database field names (user_id)
      const userId = member.userId || member.user_id || member.id;

      // Get user data from nested User object or member object itself
      const userData = member.user || member.User || member;
      const userName =
        userData?.name ||
        `${userData?.first_name || ""} ${userData?.last_name || ""}`.trim() ||
        "Unknown User";
      const nameParts = userName.split(" ");
      const firstName = nameParts[0] || "";
      const lastName = nameParts.slice(1).join(" ") || "";

      return {
        id: userId,
        name: userName,
        first_name: firstName,
        last_name: lastName,
        email: userData?.email || "",
        role: member.role || member.roleName || "requester",
        role_id: member.roleId || member.role_id || "",
        role_name: member.roleName || member.role_name || member.role || "",
        department: member.department || "",
        department_id: member.departmentId || member.department_id || "",
        active: member.active !== undefined ? member.active : true,
        is_active: member.active !== undefined ? member.active : true,
        // Include original member data for reference
        member_id: member.id,
        title: member.title || "",
        joined_at: member.joinedAt || member.joined_at,
        created_at: member.createdAt || member.created_at,
        updated_at: member.updatedAt || member.updated_at,
      };
    });

    return successResponse(
      transformedUsers,
      "Organization members fetched successfully",
    );
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
  // For organization members, we'll get all members and filter by user ID
  // This is because the backend doesn't have a specific endpoint for single member by user ID
  try {
    const response = await getUsers();
    if (!response.success || !response.data) {
      return {
        success: false,
        message: "Failed to fetch organization members",
        data: null,
        status: 400,
      };
    }

    const members = Array.isArray(response.data)
      ? response.data
      : response.data.data || [];
    const member = members.find((m: any) => m.user_id === id || m.id === id);

    if (!member) {
      return {
        success: false,
        message: "User not found in organization",
        data: null,
        status: 404,
      };
    }

    return successResponse(member, "User fetched successfully");
  } catch (error) {
    return handleError(error, "GET", `/api/v1/organization/members`);
  }
}

export async function updateUser(
  id: string,
  data: Partial<User>,
): Promise<APIResponse> {
  // TODO: Backend needs to implement PUT /api/v1/organization/members/:id endpoint
  // For now, this will return an error indicating the feature is not implemented
  return {
    success: false,
    message:
      "User update functionality requires backend implementation of organization member update endpoint",
    data: null,
    status: 501,
  };
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
  isActive: boolean,
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
  enabled: boolean,
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
  password: string,
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
