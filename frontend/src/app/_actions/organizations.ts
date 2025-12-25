"use server";

import type { APIResponse } from "@/types";
import authenticatedApiClient, { handleError, successResponse } from "./api-config";
import { updateAuthSession } from "@/lib/auth";

export interface Organization {
  id: string;
  name: string;
  slug: string;
  logoUrl?: string;
  primaryColor?: string;
  tier: 'free' | 'pro' | 'enterprise';
  active: boolean;
  description?: string;
  createdAt: string;
  updatedAt: string;
}

/**
 * Fetch all organizations for the current user
 * Server action that can be called from client components using React Query
 *
 * @returns {Promise<Organization[]>}
 *
 * @example
 * ```typescript
 * const { data: organizations } = useQuery({
 *   queryKey: ['organizations'],
 *   queryFn: () => fetchUserOrganizations(),
 * })
 * ```
 */
export async function fetchUserOrganizations(): Promise<Organization[]> {
  const url = `/api/v1/organizations`;

  try {
    const response = await authenticatedApiClient({
      url: url,
      method: "GET"
    });

    return response.data.data || [];
  } catch (error: any) {
    console.error("Failed to fetch organizations:", error);
    throw error;
  }
}

/**
 * Switch to a different organization/workspace
 * Server action that can be called from client components using React Query mutations
 *
 * @param {string} orgId - The organization ID to switch to
 * @returns {Promise<string>} - Returns the organization ID on success
 *
 * @example
 * ```typescript
 * const switchMutation = useMutation({
 *   mutationFn: (orgId: string) => switchOrganization(orgId),
 * })
 * ```
 */
export async function switchOrganization(orgId: string): Promise<string> {
  const url = `/api/v1/organizations/${orgId}/switch`;

  try {
    await authenticatedApiClient({
      url: url,
      method: "POST"
    });

    // Update frontend session with new organization ID
    await updateAuthSession({
      organization_id: orgId,
    });

    return orgId;
  } catch (error: any) {
    console.error("Failed to switch organization:", error);
    throw error;
  }
}
