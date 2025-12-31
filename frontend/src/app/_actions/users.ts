"use server";

import type { APIResponse } from "@/types";
import authenticatedApiClient, { handleError, successResponse } from "./api-config";

export interface User {
  id: string;
  email: string;
  name: string;
  roles: string[];
  department?: string;
  position?: string;
  active: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface CreateUserRequest {
  email: string;
  name: string;
  roles: string[];
  department?: string;
  position?: string;
}

export interface UpdateUserRequest {
  id: string;
  name?: string;
  department?: string;
  roles?: string[];
}

/**
 * Fetch all users with pagination and filters
 * Calls: GET /api/v1/users
 */
export async function fetchUsers(
  page: number = 1,
  limit: number = 20,
  filters?: {
    department?: string;
    role?: string;
  }
): Promise<APIResponse<User[]>> {
  const params = new URLSearchParams();
  params.set('page', page.toString());
  params.set('limit', limit.toString());

  if (filters?.department) {
    params.set('department', filters.department);
  }
  if (filters?.role) {
    params.set('role', filters.role);
  }

  const url = `/api/v1/users?${params.toString()}`;

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "GET"
    });

    return successResponse(response.data.data || [], "Users retrieved successfully");
  } catch (error: any) {
    return handleError(error, "GET", url);
  }
}

/**
 * Create a new user
 * Calls: POST /api/v1/users
 */
export async function createUser(data: CreateUserRequest): Promise<APIResponse<User>> {
  const url = `/api/v1/users`;

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "POST",
      data: {
        email: data.email,
        name: data.name,
        roles: data.roles,
        department: data.department,
        position: data.position,
      }
    });

    return successResponse(response.data.data, "User created successfully");
  } catch (error: any) {
    return handleError(error, "POST", url);
  }
}

/**
 * Get user by ID
 * Calls: GET /api/v1/users/{id}
 */
export async function getUserById(userId: string): Promise<APIResponse<User>> {
  const url = `/api/v1/users/${userId}`;

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "GET"
    });

    return successResponse(response.data.data, "User retrieved successfully");
  } catch (error: any) {
    return handleError(error, "GET", url);
  }
}

/**
 * Update user
 * Calls: PUT /api/v1/users/{id}
 */
export async function updateUser(data: UpdateUserRequest): Promise<APIResponse<User>> {
  const url = `/api/v1/users/${data.id}`;

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "PUT",
      data: {
        name: data.name,
        department: data.department,
        roles: data.roles,
      }
    });

    return successResponse(response.data.data, "User updated successfully");
  } catch (error: any) {
    return handleError(error, "PUT", url);
  }
}

/**
 * Deactivate user
 * Calls: DELETE /api/v1/users/{id}
 */
export async function deactivateUser(userId: string): Promise<APIResponse> {
  const url = `/api/v1/users/${userId}`;

  try {
    await authenticatedApiClient({
      url: url,
      method: "DELETE"
    });

    return successResponse(null, "User deactivated successfully");
  } catch (error: any) {
    return handleError(error, "DELETE", url);
  }
}