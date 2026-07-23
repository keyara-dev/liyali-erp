"use server";

import { APIResponse } from "@/types";
import {
  handleError,
  successResponse,
  NO_CACHE_HEADERS,
} from "./api-config";
import authenticatedApiClient from "./api-config";
import type { Payee, CreatePayeeInput, UpdatePayeeInput, PayeeType } from "@/types/payee";

/**
 * List payees with optional filters
 * Calls: GET /api/payees
 */
export async function getPayees(params?: {
  type?: PayeeType;
  q?: string;
}): Promise<APIResponse<Payee[]>> {
  const search = new URLSearchParams();
  if (params?.type) search.set("type", params.type);
  if (params?.q) search.set("q", params.q);
  const url = `/api/payees${search.toString() ? `?${search}` : ""}`;

  try {
    const response = await authenticatedApiClient({
      method: "GET",
      url,
      headers: NO_CACHE_HEADERS,
    });
    return successResponse(
      response.data?.items ?? response.data?.data ?? [],
      "Payees retrieved successfully",
    );
  } catch (error: any) {
    return handleError(error, "GET", url);
  }
}

/**
 * Get a single payee by ID
 * Calls: GET /api/payees/:id
 */
export async function getPayeeById(id: string): Promise<APIResponse<Payee>> {
  const url = `/api/payees/${id}`;

  try {
    const response = await authenticatedApiClient({
      method: "GET",
      url,
      headers: NO_CACHE_HEADERS,
    });
    return successResponse(
      response.data?.data ?? response.data,
      "Payee retrieved successfully",
    );
  } catch (error: any) {
    return handleError(error, "GET", url);
  }
}

/**
 * Create a new payee
 * Calls: POST /api/payees
 */
export async function createPayee(
  data: CreatePayeeInput,
): Promise<APIResponse<Payee>> {
  const url = `/api/payees`;

  try {
    const response = await authenticatedApiClient({
      method: "POST",
      url,
      data,
    });
    return successResponse(
      response.data?.data ?? response.data,
      "Payee created successfully",
    );
  } catch (error: any) {
    return handleError(error, "POST", url);
  }
}

/**
 * Update an existing payee
 * Calls: PUT /api/payees/:id
 */
export async function updatePayee(
  id: string,
  data: UpdatePayeeInput,
): Promise<APIResponse<Payee>> {
  const url = `/api/payees/${id}`;

  try {
    const response = await authenticatedApiClient({
      method: "PUT",
      url,
      data,
    });
    return successResponse(
      response.data?.data ?? response.data,
      "Payee updated successfully",
    );
  } catch (error: any) {
    return handleError(error, "PUT", url);
  }
}

/**
 * Delete a payee
 * Calls: DELETE /api/payees/:id
 */
export async function deletePayee(id: string): Promise<APIResponse<void>> {
  const url = `/api/payees/${id}`;

  try {
    await authenticatedApiClient({ method: "DELETE", url });
    return successResponse(undefined, "Payee deleted successfully");
  } catch (error: any) {
    return handleError(error, "DELETE", url);
  }
}
