"use server";

import { revalidatePath } from "next/cache";
import type { APIResponse } from "@/types";
import authenticatedApiClient, {
  handleError,
  successResponse,
} from "./api-config";

export type InvitationStatus =
  | "pending"
  | "accepted"
  | "declined"
  | "expired"
  | "cancelled";

export interface OrganizationInvitation {
  id: string;
  organizationId: string;
  organization?: { id: string; name: string; logoUrl?: string };
  invitedUserId?: string;
  invitedUser?: { id: string; name: string; email: string };
  invitedEmail: string;
  invitedBy: string;
  invitedByUser?: { id: string; name: string; email: string };
  role: string;
  departmentId?: string;
  branchId?: string;
  status: InvitationStatus;
  expiresAt: string;
  acceptedAt?: string;
  declinedAt?: string;
  createdAt: string;
  updatedAt: string;
}

export interface EmailLookupResult {
  exists: boolean;
  isOrgMember: boolean;
  hasPendingInvitation: boolean;
  userId?: string;
  name?: string;
  email?: string;
}

export interface SendInvitationRequest {
  email: string;
  role?: string;
  department_id?: string;
  branch_id?: string;
}

// ─── Email lookup ─────────────────────────────────────────────────────────────

export async function lookupUserByEmail(
  email: string,
): Promise<APIResponse<EmailLookupResult>> {
  const url = `/api/v1/organization/users/lookup?email=${encodeURIComponent(email)}`;
  try {
    const response = await authenticatedApiClient({ url, method: "GET" });
    if (!response.data?.success) {
      return handleError(
        new Error(response.data?.message || "Lookup failed"),
        "GET",
        url,
      );
    }
    return successResponse(response.data.data);
  } catch (error: any) {
    return handleError(error, "GET", url);
  }
}

// ─── Admin-side actions ───────────────────────────────────────────────────────

export async function sendInvitation(
  data: SendInvitationRequest,
): Promise<APIResponse> {
  const url = `/api/v1/organization/invitations`;
  try {
    const response = await authenticatedApiClient({
      url,
      method: "POST",
      data,
    });
    if (!response.data?.success) {
      return handleError(
        new Error(response.data?.message || "Failed to send invitation"),
        "POST",
        url,
      );
    }
    revalidatePath("/admin/invitations");
    return successResponse(response.data.data);
  } catch (error: any) {
    return handleError(error, "POST", url);
  }
}

export async function listOrgInvitations(): Promise<
  APIResponse<OrganizationInvitation[]>
> {
  const url = `/api/v1/organization/invitations`;
  try {
    const response = await authenticatedApiClient({ url, method: "GET" });
    if (!response.data?.success) {
      return handleError(
        new Error(response.data?.message || "Failed to fetch invitations"),
        "GET",
        url,
      );
    }
    return successResponse(response.data.data ?? []);
  } catch (error: any) {
    return handleError(error, "GET", url);
  }
}

export async function cancelInvitation(id: string): Promise<APIResponse> {
  const url = `/api/v1/organization/invitations/${id}`;
  try {
    const response = await authenticatedApiClient({ url, method: "DELETE" });
    if (!response.data?.success) {
      return handleError(
        new Error(response.data?.message || "Failed to cancel invitation"),
        "DELETE",
        url,
      );
    }
    revalidatePath("/admin/invitations");
    return successResponse(response.data.data);
  } catch (error: any) {
    return handleError(error, "DELETE", url);
  }
}

export async function resendInvitation(id: string): Promise<APIResponse> {
  const url = `/api/v1/organization/invitations/${id}/resend`;
  try {
    const response = await authenticatedApiClient({ url, method: "POST" });
    if (!response.data?.success) {
      return handleError(
        new Error(response.data?.message || "Failed to resend invitation"),
        "POST",
        url,
      );
    }
    revalidatePath("/admin/invitations");
    return successResponse(response.data.data);
  } catch (error: any) {
    return handleError(error, "POST", url);
  }
}

// ─── Invitee-facing actions ───────────────────────────────────────────────────

export async function getMyPendingInvitations(): Promise<
  APIResponse<OrganizationInvitation[]>
> {
  const url = `/api/v1/invitations/pending`;
  try {
    const response = await authenticatedApiClient({ url, method: "GET" });
    if (!response.data?.success) {
      return handleError(
        new Error(response.data?.message || "Failed to fetch invitations"),
        "GET",
        url,
      );
    }
    return successResponse(response.data.data ?? []);
  } catch (error: any) {
    return handleError(error, "GET", url);
  }
}

export async function acceptInvitation(token: string): Promise<APIResponse> {
  const url = `/api/v1/invitations/${token}/accept`;
  try {
    const response = await authenticatedApiClient({ url, method: "POST" });
    if (!response.data?.success) {
      return handleError(
        new Error(response.data?.message || "Failed to accept invitation"),
        "POST",
        url,
      );
    }
    revalidatePath("/invitations");
    return successResponse(response.data.data);
  } catch (error: any) {
    return handleError(error, "POST", url);
  }
}

export async function declineInvitation(token: string): Promise<APIResponse> {
  const url = `/api/v1/invitations/${token}/decline`;
  try {
    const response = await authenticatedApiClient({ url, method: "POST" });
    if (!response.data?.success) {
      return handleError(
        new Error(response.data?.message || "Failed to decline invitation"),
        "POST",
        url,
      );
    }
    revalidatePath("/invitations");
    return successResponse(response.data.data);
  } catch (error: any) {
    return handleError(error, "POST", url);
  }
}
