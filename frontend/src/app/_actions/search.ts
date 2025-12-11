"use server";

import { verifySession } from "@/lib/auth";
import {
  APIResponse,
  SearchFilters,
  WorkflowDocument,
  PaginatedResponse,
} from "@/types";
import { unauthorizedResponse, handleError } from "@/app/_actions/api-config";

// Note: This is a temporary placeholder. Server Actions cannot directly access localStorage.
// In production, implement proper API endpoints that query from a backend database.
// The search functionality is currently handled on the client side using React Query hooks
// and localStorage access in the transaction-results component.

export async function searchDocuments(
  filters: SearchFilters,
  page: number = 1,
  limit: number = 10
): Promise<APIResponse<PaginatedResponse<WorkflowDocument>>> {
  const { session } = await verifySession();

  if (!session?.user) {
    return unauthorizedResponse();
  }

  try {
    // Server Actions run on the server and cannot access browser localStorage directly.
    // This endpoint is kept for compatibility but returns empty results.
    // Client-side search using React Query hooks is the recommended approach until
    // proper backend APIs are implemented.

    return {
      success: true,
      message: "Use client-side search for now",
      data: {
        data: [],
        pagination: {
          page,
          limit,
          total: 0,
          totalPages: 0,
        },
      },
      status: 200,
    };
  } catch (error) {
    return handleError(error, "GET", "/search") as any;
  }
}

export async function downloadDocumentPDF(
  documentId: string
): Promise<APIResponse<{ downloadUrl: string }>> {
  const { session } = await verifySession();

  if (!session?.user) {
    return unauthorizedResponse() as any;
  }

  try {
    // Try to fetch document to verify it exists
    const { getDocument } = await import("./workflow");
    const result = await getDocument(documentId);

    if (!result.success || !result.data) {
      return {
        success: false,
        message: "Document not found",
      } as any;
    }

    // Generate a mock download URL (in real app, would generate actual PDF)
    const downloadUrl = `/api/documents/${documentId}/download`;

    return {
      success: true,
      message: "Download URL generated",
      data: {
        downloadUrl,
      },
    } as any;
  } catch (error) {
    return handleError(
      error,
      "GET",
      `/documents/${documentId}/download`
    ) as any;
  }
}
