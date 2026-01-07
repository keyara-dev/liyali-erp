'use server'

import { verifySession } from '@/lib/auth'
import {
  APIResponse,
  SearchFilters,
  WorkflowDocument,
  PaginatedResponse,
} from '@/types'
import { unauthorizedResponse, handleError } from '@/app/_actions/api-config'

/**
 * Server action to search documents from the backend API
 * Connects to backend endpoint: GET /api/v1/documents/search
 */
export async function searchDocuments(
  filters: SearchFilters,
  page: number = 1,
  limit: number = 10
): Promise<APIResponse<PaginatedResponse<WorkflowDocument>>> {
  const { session } = await verifySession()

  if (!session?.user) {
    return unauthorizedResponse()
  }

  try {
    const backendUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'
    const token = session.user.token

    if (!token) {
      return unauthorizedResponse()
    }

    // Build query parameters
    const queryParams = new URLSearchParams()

    if (filters.documentNumber) {
      queryParams.append('documentNumber', filters.documentNumber)
    }
    if (filters.documentType && filters.documentType !== 'ALL') {
      queryParams.append('documentType', filters.documentType)
    }
    if (filters.status && filters.status !== 'ALL') {
      queryParams.append('status', filters.status)
    }
    if (filters.startDate) {
      queryParams.append('startDate', filters.startDate)
    }
    if (filters.endDate) {
      queryParams.append('endDate', filters.endDate)
    }

    queryParams.append('page', page.toString())
    queryParams.append('pageSize', limit.toString())

    const response = await fetch(
      `${backendUrl}/api/v1/documents/search?${queryParams.toString()}`,
      {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        cache: 'no-store',
      }
    )

    if (!response.ok) {
      if (response.status === 401) {
        return unauthorizedResponse()
      }

      const error = await response.json().catch(() => ({}))
      return {
        success: false,
        message: error.message || `Search failed: ${response.statusText}`,
        status: response.status,
      } as any
    }

    const data = await response.json()

    return {
      success: true,
      message: 'Search completed',
      data: {
        data: data.documents || [],
        pagination: {
          page: data.page || page,
          limit: data.pageSize || limit,
          total: data.total || 0,
          totalPages: data.totalPages || Math.ceil((data.total || 0) / limit),
          hasNext: (data.page || page) < (data.totalPages || Math.ceil((data.total || 0) / limit)),
          hasPrev: (data.page || page) > 1,
        },
      },
      status: 200,
    }
  } catch (error) {
    return handleError(error, 'GET', '/documents/search') as any
  }
}

/**
 * Server action to download a document as PDF
 * Connects to backend endpoint: GET /api/v1/documents/{documentId}/download
 */
export async function downloadDocumentPDF(
  documentId: string
): Promise<APIResponse<{ downloadUrl: string }>> {
  const { session } = await verifySession()

  if (!session?.user) {
    return unauthorizedResponse() as any
  }

  try {
    const backendUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'
    const token = session.user.token

    if (!token) {
      return unauthorizedResponse() as any
    }

    const response = await fetch(
      `${backendUrl}/api/v1/documents/${documentId}/download`,
      {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Accept': 'application/pdf',
        },
        cache: 'no-store',
      }
    )

    if (!response.ok) {
      if (response.status === 404) {
        return {
          success: false,
          message: 'Document not found',
          status: 404,
        } as any
      }

      if (response.status === 403) {
        return {
          success: false,
          message: 'You do not have permission to download this document',
          status: 403,
        } as any
      }

      return {
        success: false,
        message: `Failed to download document: ${response.statusText}`,
        status: response.status,
      } as any
    }

    // Create a blob URL for the PDF
    const blob = await response.blob()
    const downloadUrl = URL.createObjectURL(blob)

    return {
      success: true,
      message: 'Download URL generated',
      data: {
        downloadUrl,
      },
      status: 200,
    } as any
  } catch (error) {
    return handleError(error, 'GET', `/documents/${documentId}/download`) as any
  }
}
