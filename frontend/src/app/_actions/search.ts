"use server";

import { verifySession } from "@/lib/auth";
import {
  APIResponse,
  SearchFilters,
  WorkflowDocument,
  PaginatedResponse,
} from "@/types";
import { getDocumentsByCreator, getPendingApprovals } from "./workflow";
import { unauthorizedResponse, handleError } from "@/app/_actions/api-config";

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
    // Get all documents - use high limit to fetch all documents
    const createdResult = await getDocumentsByCreator(session.user.id, 1, 1000);
    const pendingResult = await getPendingApprovals(session.user.id);

    let allDocuments: WorkflowDocument[] = [];

    if (createdResult.success && createdResult.data?.data) {
      allDocuments = [...allDocuments, ...createdResult.data.data];
    }

    if (pendingResult.success && pendingResult.data) {
      allDocuments = [...allDocuments, ...pendingResult.data];
    }

    // Remove duplicates by ID
    const uniqueMap = new Map<string, WorkflowDocument>();
    allDocuments.forEach((doc) => uniqueMap.set(doc.id, doc));
    const uniqueDocuments = Array.from(uniqueMap.values());

    // Apply filters
    let filtered = uniqueDocuments.filter((doc) => {
      // Filter by document number (case-insensitive, partial match)
      if (
        filters.documentNumber &&
        !doc.documentNumber
          .toLowerCase()
          .includes(filters.documentNumber.toLowerCase())
      ) {
        return false;
      }

      // Filter by document type
      if (filters.documentType !== "ALL" && doc.type !== filters.documentType) {
        return false;
      }

      // Filter by status
      if (filters.status !== "ALL" && doc.status !== filters.status) {
        return false;
      }

      // Filter by start date
      if (filters.startDate) {
        const startDate = new Date(filters.startDate);
        if (doc.createdAt < startDate) {
          return false;
        }
      }

      // Filter by end date
      if (filters.endDate) {
        const endDate = new Date(filters.endDate);
        endDate.setHours(23, 59, 59, 999); // Include the entire end date
        if (doc.createdAt > endDate) {
          return false;
        }
      }

      return true;
    });

    // Sort by created date (newest first)
    filtered.sort(
      (a, b) =>
        new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
    );

    // Calculate pagination
    const total = filtered.length;
    const totalPages = Math.ceil(total / limit);
    const skip = (page - 1) * limit;
    const paginatedData = filtered.slice(skip, skip + limit);

    return {
      success: true,
      message: "Documents search completed",
      data: {
        data: paginatedData,
        pagination: {
          page,
          limit,
          total,
          totalPages,
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
