"use server";

import authenticatedApiClient, {
  badRequestResponse,
  handleError,
  successResponse,
} from "./api-config";
import { revalidatePath } from "next/cache";
import { is } from "date-fns/locale";
import { APIResponse, Pagination } from "@/types";
import { Department } from "@/lib/mock-departments";

// ============================================================================
// BRANCH MANAGEMENT
// ============================================================================

/**
 * Get all branches with optional filtering
 * Endpoint: GET /api/v1/branches
 * Status: ✅ Documented in API
 * Query Parameters: province_id, town_id, is_active, limit, offset
 */
export async function getBranches(params?: {
  provinceId?: string;
  townId?: string;
  isActive?: boolean;
  page?: number;
  page_size?: number;
}): Promise<APIResponse> {
  const queryParams = new URLSearchParams();

  queryParams.append("page_size", String(params?.page_size || 10));
  queryParams.append("page", String(params?.page || 1));

  if (params?.provinceId) queryParams.append("province_id", params.provinceId);
  if (params?.townId) queryParams.append("town_id", params.townId);
  if (params?.isActive !== undefined)
    queryParams.append("is_active", String(params.isActive));

  const url = `/api/v1/branches${queryParams.toString() ? `?${queryParams.toString()}` : ""}`;

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "GET",
    });
    return successResponse(
      response?.data?.data,
      "Branches fetched successfully"
    );
  } catch (error: Error | any) {
    return handleError(error, "GET", url);
  }
}
/**
 * Get single branch by ID
 * Endpoint: GET /api/v1/branches/{id}
 * Status: ✅ Documented in API
 */
export async function getBranchById(id: string): Promise<APIResponse> {
  const url = `/api/v1/branches/${id}`;

  try {
    const response = await authenticatedApiClient({ url });
    return successResponse(response?.data, "Branch fetched successfully");
  } catch (error: Error | any) {
    return handleError(error, "GET", url);
  }
}

/**
 * Create new branch
 * Endpoint: POST /api/v1/branches
 * Status: ✅ Documented in API
 *
 * NOTE: API expects town_id and province_id (UUIDs), not string names.
 * UI should use dropdowns populated from /api/v1/provinces/with-towns
 */
export async function createBranch({
  name,
  code,
  townId,
  provinceId,
  address,
  // isActive = true
}: {
  name: string;
  code: string;
  townId: string;
  provinceId: string;
  address?: string;
  // isActive?: boolean;
}): Promise<APIResponse> {
  const url = `/api/v1/branches`;

  if (!name || !code || !townId || !provinceId) {
    return badRequestResponse(
      "Name, code, town ID, and province ID are required"
    );
  }

  try {
    const response = await authenticatedApiClient({
      url,
      method: "POST",
      data: {
        name,
        code,
        town_id: townId,
        province_id: provinceId,
        address,
        // is_active: isActive
      },
    });
    revalidatePath("/dashboard/system-configs/locations");
    return successResponse(response?.data, "Branch created successfully");
  } catch (error: Error | any) {
    return handleError(error, "POST", url);
  }
}

/**
 * Update existing branch
 * Endpoint: PUT /api/v1/branches/{id}
 * Status: ✅ Documented in API
 */
export async function updateBranch({
  id,
  name,
  code,
  townId,
  provinceId,
  address,
  // isActive
}: {
  id: string;
  name: string;
  code: string;
  townId: string;
  provinceId: string;
  address?: string;
  // isActive?: boolean;
}): Promise<APIResponse> {
  const url = `/api/v1/branches/${id}`;

  if (!id || !name || !code || !townId || !provinceId) {
    return badRequestResponse(
      "ID, name, code, town ID, and province ID are required"
    );
  }

  try {
    const response = await authenticatedApiClient({
      url,
      method: "PUT",
      data: {
        name,
        code,
        town_id: townId,
        province_id: provinceId,
        address,
        // is_active: isActive,
        manager_id: null, // Optional field from API docs
      },
    });
    revalidatePath("/dashboard/system-configs/locations");
    return successResponse(response?.data, "Branch updated successfully");
  } catch (error: Error | any) {
    return handleError(error, "PUT", url);
  }
}

/**
 * Delete branch
 * Endpoint: DELETE /api/v1/branches/{id}
 * Status: ✅ Documented in API
 */
export async function deleteBranch(id: string): Promise<APIResponse> {
  const url = `/api/v1/branches/${id}`;

  if (!id) {
    return badRequestResponse("Branch ID is required");
  }

  try {
    await authenticatedApiClient({ url, method: "DELETE" });
    revalidatePath("/dashboard/system-configs/locations");
    return successResponse(null, "Branch deleted successfully");
  } catch (error: Error | any) {
    return handleError(error, "DELETE", url);
  }
}

// ============================================================================
// DEPARTMENT MANAGEMENT
// ============================================================================

/**
 * Get all departments with optional filtering
 * Endpoint: GET /api/v1/departments
 * Status: ✅ Documented in API
 * Query Parameters: parent_id, is_active, limit, offset
 */
export async function getDepartments(
  params?: Partial<Pagination> & {
    parent_id?: string;
    is_active?: boolean;
  }
): Promise<APIResponse> {
  const queryParams = new URLSearchParams();
  queryParams.append("page_size", String(params?.page_size || 10));
  queryParams.append("page", String(params?.page || 1));

  if (params?.parent_id) queryParams.append("parent_id", params.parent_id);
  if (params?.is_active !== undefined)
    queryParams.append("is_active", String(params.is_active));

  const url = `/api/v1/departments${queryParams.toString() ? `?${queryParams.toString()}` : ""}`;

  try {
    const response = await authenticatedApiClient({ url });
    return successResponse(
      response?.data?.data,
      "Departments fetched successfully"
    );
  } catch (error: Error | any) {
    return handleError(error, "GET", url);
  }
}

/**
 * Get single department by ID
 * Endpoint: GET /api/v1/departments/{id}
 * Status: ✅ Documented in API
 */
export async function getDepartmentById(id: string): Promise<APIResponse> {
  const url = `/api/v1/departments/${id}`;

  try {
    const response = await authenticatedApiClient({ url });
    return successResponse(response?.data, "Department fetched successfully");
  } catch (error: Error | any) {
    return handleError(error, "GET", url);
  }
}
export async function getDepartmentRiskCategories(
  id: string
): Promise<APIResponse> {
  const url = `/api/v1/departments/${id}/risk-categories`;
  try {
    const response = await authenticatedApiClient({ url });
    return successResponse(response?.data.data);
  } catch (error: Error | any) {
    return handleError(error, "GET", url);
  }
}

/**
 * Create new department
 * Endpoint: POST /api/v1/departments
 * Status: ✅ Documented in API
 *
 * NOTE: Supports hierarchical departments via parent_id
 */
export async function createDepartment(data: Department): Promise<APIResponse> {
  const url = `/api/v1/departments`;

  if (!data?.name) {
    return badRequestResponse("Name and code are required");
  }

  try {
    const response = await authenticatedApiClient({
      url,
      method: "POST",
      data,
    });
    revalidatePath("/dashboard/system-configs/departments");
    return successResponse(response?.data, "Department created successfully");
  } catch (error: Error | any) {
    return handleError(error, "POST", url);
  }
}

/**
 * Update existing department
 * Endpoint: PUT /api/v1/departments/{id}
 * Status: ✅ Documented in API
 */
export async function updateDepartment(data: Department): Promise<APIResponse> {
  if (!data?.id) {
    return badRequestResponse("Department ID is required");
  }
  const url = `/api/v1/departments/${data?.id}`;

  try {
    const response = await authenticatedApiClient({ url, method: "PUT", data });
    revalidatePath("/dashboard/system-configs/departments");
    return successResponse(response?.data, "Department updated successfully");
  } catch (error: Error | any) {
    return handleError(error, "PUT", url);
  }
}

/**
 * Delete department
 * Endpoint: DELETE /api/v1/departments/{id}
 * Status: ✅ Documented in API
 */
export async function deleteDepartment(id: string): Promise<APIResponse> {
  const url = `/api/v1/departments/${id}`;

  if (!id) {
    return badRequestResponse("Department ID is required");
  }

  try {
    await authenticatedApiClient({
      url,
      method: "DELETE",
    });
    revalidatePath("/dashboard/system-configs/departments");
    return successResponse(null, "Department deleted successfully");
  } catch (error: Error | any) {
    return handleError(error, "DELETE", url);
  }
}

/**
 * Get modules assigned to a department
 * Endpoint: GET /api/v1/departments/{id}/modules
 * Status: ✅ Documented in API
 */
export async function getDepartmentModules(
  departmentId: string
): Promise<APIResponse> {
  const url = `/api/v1/departments/${departmentId}/modules`;

  if (!departmentId) {
    return badRequestResponse("Department ID is required");
  }

  try {
    const response = await authenticatedApiClient({ url });
    return successResponse(
      response?.data,
      "Department modules fetched successfully"
    );
  } catch (error: Error | any) {
    return handleError(error, "GET", url);
  }
}

/**
 * Assign module to department
 * Endpoint: POST /api/v1/departments/{id}/modules
 * Status: ✅ Documented in API
 */
export async function assignModuleToDepartment({
  departmentId,
  moduleId,
}: {
  departmentId: string;
  moduleId: string;
}): Promise<APIResponse> {
  const url = `/api/v1/departments/${departmentId}/modules`;

  if (!departmentId || !moduleId) {
    return badRequestResponse("Department ID and Module ID are required");
  }

  try {
    const response = await authenticatedApiClient({
      url,
      method: "POST",
      data: {
        module_id: moduleId,
      },
    });
    revalidatePath("/dashboard/system-configs/departments");
    return successResponse(
      response?.data,
      "Module assigned to department successfully"
    );
  } catch (error: Error | any) {
    return handleError(error, "POST", url);
  }
}

/**
 * Remove module from department
 * Endpoint: DELETE /api/v1/departments/{dept_id}/modules/{module_id}
 * Status: ✅ Documented in API
 */
export async function removeModuleFromDepartment({
  departmentId,
  moduleId,
}: {
  departmentId: string;
  moduleId: string;
}): Promise<APIResponse> {
  const url = `/api/v1/departments/${departmentId}/modules/${moduleId}`;

  if (!departmentId || !moduleId) {
    return badRequestResponse("Department ID and Module ID are required");
  }

  try {
    await authenticatedApiClient({ url, method: "DELETE" });
    revalidatePath("/dashboard/system-configs/departments");
    return successResponse(null, "Module removed from department successfully");
  } catch (error: Error | any) {
    return handleError(error, "DELETE", url);
  }
}

// ============================================================================
// MODULE MANAGEMENT
// ============================================================================

/**
 * Get all modules
 * Endpoint: GET /api/v1/modules
 * Status: ✅ Documented in API
 * Query Parameter: hierarchy=true for parent-child tree structure
 */
export async function getModules(
  hierarchy: boolean = false
): Promise<APIResponse> {
  const url = `/api/v1/modules${hierarchy ? "?hierarchy=true" : ""}`;

  try {
    const response = await authenticatedApiClient({ url });
    return successResponse(response?.data, "Modules fetched successfully");
  } catch (error: Error | any) {
    return handleError(error, "GET", url);
  }
}

/**
 * Get single module by ID
 * Endpoint: GET /api/v1/modules/{id}
 * Status: ✅ Documented in API
 */
export async function getModuleById(id: string): Promise<APIResponse> {
  const url = `/api/v1/modules/${id}`;

  try {
    const response = await authenticatedApiClient({ url });
    return successResponse(response?.data, "Module fetched successfully");
  } catch (error: Error | any) {
    return handleError(error, "GET", url);
  }
}

/**
 * Create new module
 * Endpoint: POST /api/v1/modules
 * Status: ✅ Documented in API
 */
export async function createModule({
  module_code,
  name,
  description,
  parent_module_id,
  href,
  icon,
  sortOrder,
}: {
  module_code: string;
  name: string;
  description?: string;
  parent_module_id?: string | null;
  href?: string | null;
  icon?: string;
  sortOrder?: number;
}): Promise<APIResponse> {
  const url = `/api/v1/modules`;

  if (!module_code || !name) {
    return badRequestResponse("Module code and name are required");
  }

  try {
    const response = await authenticatedApiClient({
      url,
      method: "POST",
      data: {
        module_code: module_code,
        name,
        description,
        parent_module_id: parent_module_id || null,
        href: href || null,
        icon,
        sort_order: sortOrder,
      },
    });

    return successResponse(response?.data, "Module created successfully");
  } catch (error: Error | any) {
    return handleError(error, "POST", url);
  }
}

/**
 * Update existing module
 * Endpoint: PUT /api/v1/modules/{id}
 * Status: ✅ Documented in API
 */
export async function updateModule({
  id,
  module_code,
  name,
  description,
  parent_module_id,
  href,
  icon,
  sortOrder,
  isActive,
}: {
  id: string;
  module_code: string;
  name: string;
  description?: string;
  parent_module_id?: string | null;
  href?: string | null;
  icon?: string;
  sortOrder?: number;
  isActive?: boolean;
}): Promise<APIResponse> {
  const url = `/api/v1/modules/${id}`;

  if (!id || !module_code || !name) {
    return badRequestResponse("ID, module code, and name are required");
  }

  try {
    const response = await authenticatedApiClient({
      url,
      method: "PUT",
      data: {
        module_code: module_code,
        name,
        description,
        parent_module_id: parent_module_id || null,
        href: href || null,
        icon,
        sort_order: sortOrder,
        is_active: isActive,
      },
    });

    return successResponse(response?.data, "Module updated successfully");
  } catch (error: Error | any) {
    return handleError(error, "PUT", url);
  }
}

/**
 * Delete module
 * Endpoint: DELETE /api/v1/modules/{id}
 * Status: ✅ Documented in API
 */
export async function deleteModule(id: string): Promise<APIResponse> {
  const url = `/api/v1/modules/${id}`;

  if (!id) {
    return badRequestResponse("Module ID is required");
  }

  try {
    await authenticatedApiClient({ url });
    return successResponse(null, "Module deleted successfully");
  } catch (error: Error | any) {
    return handleError(error, "DELETE", url);
  }
}

/**
 * Get sub-modules for a parent module
 * Endpoint: GET /api/v1/modules/{id}/submodules
 * Status: ✅ Documented in API
 */
export async function getSubModules(
  parent_module_id: string
): Promise<APIResponse> {
  const url = `/api/v1/modules/${parent_module_id}/submodules`;

  if (!parent_module_id) {
    return badRequestResponse("Parent module ID is required");
  }

  try {
    const response = await authenticatedApiClient({ url });
    return successResponse(response?.data, "Sub-modules fetched successfully");
  } catch (error: Error | any) {
    return handleError(error, "GET", url);
  }
}

// ============================================================================
// ROLE MANAGEMENT
// ============================================================================

/**
 * Get all roles with optional filtering
 * Endpoint: GET /api/v1/roles
 * Status: ✅ Documented in API
 * Query Parameters: department_id, is_active, limit, offset
 */
export async function getRoles(params?: {
  departmentId?: string;
  isActive?: boolean;
  limit?: number;
  offset?: number;
}): Promise<APIResponse> {
  const queryParams = new URLSearchParams();
  if (params?.departmentId)
    queryParams.append("department_id", params.departmentId);
  if (params?.isActive !== undefined)
    queryParams.append("is_active", String(params.isActive));
  if (params?.limit) queryParams.append("limit", String(params.limit));
  if (params?.offset) queryParams.append("offset", String(params.offset));

  const url = `/api/v1/roles${queryParams.toString() ? `?${queryParams.toString()}` : ""}`;

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "GET",
    });
    return successResponse(response?.data?.data, "Roles fetched successfully");
  } catch (error: Error | any) {
    return handleError(error, "GET", url);
  }
}

/**
 * Get single role by ID
 * Endpoint: GET /api/v1/roles/{id}
 * Status: ✅ Documented in API
 */
export async function getRoleById(id: string): Promise<APIResponse> {
  const url = `/api/v1/roles/${id}`;

  try {
    const response = await authenticatedApiClient({ url });
    return successResponse(response?.data, "Role fetched successfully");
  } catch (error: Error | any) {
    return handleError(error, "GET", url);
  }
}

/**
 * Create new role
 * Endpoint: POST /api/v1/roles
 * Status: ✅ Documented in API
 */
export async function createRole(role: {
  id: string;
  name: string;
  code: string;
  description?: string;
  is_active?: boolean;
  department_id?: string;
  is_department_head?: boolean;
}): Promise<APIResponse> {
  const url = `/api/v1/roles`;

  if (!role.department_id || !role.name || !role.code) {
    return badRequestResponse("Department ID, name, and code are required");
  }

  try {
    const response = await authenticatedApiClient({
      url,
      method: "POST",
      data: {
        ...role,
        is_active: undefined, // TODO: Remove when backend is fixed to accept is_active on create
      },
    });
    revalidatePath("/dashboard/system-configs/departments");
    return successResponse(response?.data, "Role created successfully");
  } catch (error: Error | any) {
    return handleError(error, "POST", url);
  }
}

/**
 * Update existing role
 * Endpoint: PUT /api/v1/roles/{id}
 * Status: ✅ Documented in API
 * NOTE: department_id cannot be changed per API docs
 */
export async function updateRole(role: {
  id: string;
  name: string;
  code: string;
  description?: string;
  is_active?: boolean;
  department_id?: string;
  is_department_head?: boolean;
}): Promise<APIResponse> {
  const url = `/api/v1/roles/${role.id}`;

  if (!role.id || !role.name || !role.code) {
    return badRequestResponse("ID, name, and code are required");
  }

  try {
    const response = await authenticatedApiClient({
      url,
      method: "PUT",
      data: {
        name: role.name,
        code: role.code,
        description: role.description,
      },
    });
    revalidatePath("/dashboard/system-configs/departments");
    return successResponse(response?.data, "Role updated successfully");
  } catch (error: Error | any) {
    return handleError(error, "PUT", url);
  }
}

/**
 * Delete role
 * Endpoint: DELETE /api/v1/roles/{id}
 * Status: ✅ Documented in API
 */
export async function deleteRole(id: string): Promise<APIResponse> {
  const url = `/api/v1/roles/${id}`;

  if (!id) {
    return badRequestResponse("Role ID is required");
  }

  try {
    await authenticatedApiClient({ url, method: "DELETE" });
    revalidatePath("/dashboard/system-configs/departments");
    return successResponse(null, "Role deleted successfully");
  } catch (error: Error | any) {
    return handleError(error, "DELETE", url);
  }
}

// ============================================================================
// PROVINCE & TOWN MANAGEMENT
// ============================================================================

/**
 * Get all provinces with their towns nested
 * Endpoint: GET /api/v1/provinces/with-towns
 * Status: ✅ Documented in API
 * IMPORTANT: Use this endpoint to populate branch form dropdowns
 */
export async function getProvincesWithTowns(): Promise<APIResponse> {
  const url = `/api/v1/provinces/with-towns`;

  try {
    const response = await authenticatedApiClient({ url });
    return successResponse(
      response?.data?.data,
      "Provinces with towns fetched successfully"
    );
  } catch (error: Error | any) {
    return handleError(error, "GET", url);
  }
}

/**
 * Get all provinces
 * Endpoint: GET /api/v1/provinces
 * Status: ✅ Documented in API
 */
export async function getProvinces(params?: {
  isActive?: boolean;
  page?: number;
  page_size?: number;
}): Promise<APIResponse> {
  const queryParams = new URLSearchParams();

  if (params?.isActive !== undefined)
    queryParams.append("is_active", String(params.isActive));
  if (params?.page) queryParams.append("page", String(params.page));
  if (params?.page_size)
    queryParams.append("page_size", String(params.page_size));

  const url = `/api/v1/provinces${queryParams.toString() ? `?${queryParams.toString()}` : ""}`;

  try {
    const response = await authenticatedApiClient({ url });
    return successResponse(
      response?.data?.data,
      "Provinces fetched successfully"
    );
  } catch (error: Error | any) {
    return handleError(error, "GET", url);
  }
}

/**
 * Get single province by ID
 * Endpoint: GET /api/v1/provinces/{id}
 * Status: ✅ Documented in API
 */
export async function getProvinceById(id: string): Promise<APIResponse> {
  const url = `/api/v1/provinces/${id}`;

  try {
    const response = await authenticatedApiClient({ url });
    return successResponse(response?.data, "Province fetched successfully");
  } catch (error: Error | any) {
    return handleError(error, "GET", url);
  }
}

/**
 * Create new province
 * Endpoint: POST /api/v1/provinces
 * Status: ✅ Documented in API
 */
export async function createProvince({
  name,
  code,
  isActive = true,
}: {
  name: string;
  code: string;
  isActive?: boolean;
}): Promise<APIResponse> {
  const url = `/api/v1/provinces`;

  if (!name || !code) {
    return badRequestResponse("Name and code are required");
  }

  try {
    const response = await authenticatedApiClient({
      url,
      method: "POST",
      data: {
        name,
        code,
        is_active: isActive,
      },
    });
    revalidatePath("/dashboard/system-configs/locations");
    return successResponse(response?.data, "Province created successfully");
  } catch (error: Error | any) {
    return handleError(error, "POST", url);
  }
}

/**
 * Update province
 * Endpoint: PUT /api/v1/provinces/:id
 * Status: ⚠️ NOT DOCUMENTED - Mock implementation
 */
export async function updateProvince({
  id,
  name,
  code,
  isActive,
}: {
  id: string;
  name: string;
  code: string;
  isActive?: boolean;
}): Promise<APIResponse> {
  // TODO: Replace with actual API call when endpoint is available
  await new Promise((resolve) => setTimeout(resolve, 300));

  try {
    // Mock successful response
    return successResponse(
      {
        id,
        name,
        code,
        is_active: isActive ?? true,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      },
      "Province updated successfully (mock)"
    );
  } catch (error: Error | any) {
    return handleError(error, "PUT", `/api/v1/provinces/${id}`);
  }
}

/**
 * Delete province
 * Endpoint: DELETE /api/v1/provinces/:id
 * Status: ⚠️ NOT DOCUMENTED - Mock implementation
 */
export async function deleteProvince(id: string): Promise<APIResponse> {
  // TODO: Replace with actual API call when endpoint is available
  await new Promise((resolve) => setTimeout(resolve, 300));

  try {
    // Mock successful response
    return successResponse(null, "Province deleted successfully (mock)");
  } catch (error: Error | any) {
    return handleError(error, "DELETE", `/api/v1/provinces/${id}`);
  }
}

/**
 * Get all towns
 * Endpoint: GET /api/v1/towns
 * Status: ✅ Documented in API
 */
export async function getTowns(params?: {
  provinceId?: string;
  isActive?: boolean;
  page_size?: number;
  page?: number;
}): Promise<APIResponse> {
  const queryParams = new URLSearchParams();

  if (params?.provinceId) queryParams.append("province_id", params.provinceId);
  if (params?.isActive !== undefined)
    queryParams.append("is_active", String(params.isActive));
  if (params?.page_size !== undefined)
    queryParams.append("page_size", String(params.page_size));
  if (params?.page !== undefined)
    queryParams.append("page", String(params.page));

  const url = `/api/v1/towns${queryParams.toString() ? `?${queryParams.toString()}` : ""}`;

  try {
    const response = await authenticatedApiClient({ url });
    return successResponse(response?.data?.data, "Towns fetched successfully");
  } catch (error: Error | any) {
    return handleError(error, "GET", url);
  }
}

/**
 * Create new town
 * Endpoint: POST /api/v1/towns
 * Status: ✅ Documented in API
 */
export async function createTown({
  name,
  provinceId,
  isActive = true,
}: {
  name: string;
  provinceId: string;
  isActive?: boolean;
}): Promise<APIResponse> {
  const url = `/api/v1/towns`;

  if (!name || !provinceId) {
    return badRequestResponse("Name and province ID are required");
  }

  try {
    const response = await authenticatedApiClient({
      url,
      method: "POST",
      data: {
        name,
        province_id: provinceId,
        is_active: isActive,
      },
    });
    return successResponse(response?.data, "Town created successfully");
  } catch (error: Error | any) {
    return handleError(error, "POST", url);
  }
}

/**
 * Update town
 * Endpoint: PUT /api/v1/towns/:id
 * Status: ⚠️ NOT DOCUMENTED - Mock implementation
 */
export async function updateTown({
  id,
  name,
  provinceId,
  isActive,
}: {
  id: string;
  name: string;
  provinceId: string;
  isActive?: boolean;
}): Promise<APIResponse> {
  // TODO: Replace with actual API call when endpoint is available
  await new Promise((resolve) => setTimeout(resolve, 300));

  try {
    // Mock successful response
    return successResponse(
      {
        id,
        name,
        province_id: provinceId,
        is_active: isActive ?? true,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      },
      "Town updated successfully (mock)"
    );
  } catch (error: Error | any) {
    return handleError(error, "PUT", `/api/v1/towns/${id}`);
  }
}

/**
 * Delete town
 * Endpoint: DELETE /api/v1/towns/:id
 * Status: ⚠️ NOT DOCUMENTED - Mock implementation
 */
export async function deleteTown(id: string): Promise<APIResponse> {
  // TODO: Replace with actual API call when endpoint is available
  await new Promise((resolve) => setTimeout(resolve, 300));

  try {
    // Mock successful response
    return successResponse(null, "Town deleted successfully (mock)");
  } catch (error: Error | any) {
    return handleError(error, "DELETE", `/api/v1/towns/${id}`);
  }
}
