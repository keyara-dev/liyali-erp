'use server'

import { APIResponse } from '@/types'
import {
  Budget,
  BudgetStatus,
  CreateBudgetRequest,
  ApproveBudgetRequest,
  RejectBudgetRequest,
  SubmitBudgetRequest,
  BudgetFilters
} from '@/types/budget'
import authenticatedApiClient from './api-config'
import { handleError, successResponse, badRequestResponse } from './api-config'

/**
 * Create a new budget draft
 */
export async function createBudget(
  request: CreateBudgetRequest
): Promise<APIResponse<Budget | null>> {
  try {
    const response = await authenticatedApiClient({
      method: 'POST',
      url: '/api/v1/budgets',
      data: request
    })

    return successResponse(
      response.data,
      'Budget created successfully'
    )
  } catch (error: any) {
    return handleError(error, 'POST', '/api/v1/budgets')
  }
}

/**
 * Update an existing budget (items, metadata, etc.)
 */
export async function updateBudget(
  budgetId: string,
  updates: Partial<Budget>
): Promise<APIResponse<Budget | null>> {
  try {
    const response = await authenticatedApiClient({
      method: 'PUT',
      url: `/api/v1/budgets/${budgetId}`,
      data: updates
    })

    return successResponse(
      response.data,
      'Budget updated successfully'
    )
  } catch (error: any) {
    return handleError(error, 'PUT', `/api/v1/budgets/${budgetId}`)
  }
}

/**
 * Get all budgets with optional filters
 */
export async function getBudgets(
  filters?: BudgetFilters,
  page: number = 1,
  limit: number = 10
): Promise<APIResponse<Budget[] | null>> {
  try {
    const params: any = {
      page,
      limit
    }

    // Add filter parameters if provided
    if (filters) {
      if (filters.status) params.status = filters.status
      if (filters.fiscalYear) params.fiscalYear = filters.fiscalYear
      if (filters.departmentId) params.departmentId = filters.departmentId
      if (filters.searchTerm) params.search = filters.searchTerm
    }

    const response = await authenticatedApiClient({
      method: 'GET',
      url: '/api/v1/budgets',
      params
    })

    return successResponse(
      response.data,
      'Budgets retrieved successfully'
    )
  } catch (error: any) {
    return handleError(error, 'GET', '/api/v1/budgets')
  }
}

/**
 * Get budget by ID
 */
export async function getBudgetById(budgetId: string): Promise<APIResponse<Budget | null>> {
  try {
    const response = await authenticatedApiClient({
      method: 'GET',
      url: `/api/v1/budgets/${budgetId}`
    })

    return successResponse(
      response.data,
      'Budget retrieved successfully'
    )
  } catch (error: any) {
    return handleError(error, 'GET', `/api/v1/budgets/${budgetId}`)
  }
}

/**
 * Submit budget for approval
 */
export async function submitBudgetForApproval(
  request: SubmitBudgetRequest
): Promise<APIResponse<Budget | null>> {
  try {
    const response = await authenticatedApiClient({
      method: 'POST',
      url: `/api/v1/budgets/${request.budgetId}/submit`,
      data: {
        submittingUserId: request.submittingUserId
      }
    })

    return successResponse(
      response.data,
      'Budget submitted for approval'
    )
  } catch (error: any) {
    return handleError(error, 'POST', `/api/v1/budgets/${request.budgetId}/submit`)
  }
}

/**
 * Approve budget
 */
export async function approveBudget(
  request: ApproveBudgetRequest
): Promise<APIResponse<Budget | null>> {
  try {
    const response = await authenticatedApiClient({
      method: 'POST',
      url: `/api/v1/budgets/${request.budgetId}/approve`,
      data: {
        approvingUserId: request.approvingUserId,
        approvingUserRole: request.approvingUserRole,
        stageNumber: request.stageNumber,
        comments: request.comments,
        signature: request.signature
      }
    })

    return successResponse(
      response.data,
      'Budget approved'
    )
  } catch (error: any) {
    return handleError(error, 'POST', `/api/v1/budgets/${request.budgetId}/approve`)
  }
}

/**
 * Reject budget
 */
export async function rejectBudget(
  request: RejectBudgetRequest
): Promise<APIResponse<Budget | null>> {
  try {
    const response = await authenticatedApiClient({
      method: 'POST',
      url: `/api/v1/budgets/${request.budgetId}/reject`,
      data: {
        rejectingUserId: request.rejectingUserId,
        rejectingUserRole: request.rejectingUserRole,
        rejectionReason: request.rejectionReason,
        comments: request.comments
      }
    })

    return successResponse(
      response.data,
      'Budget rejected'
    )
  } catch (error: any) {
    return handleError(error, 'POST', `/api/v1/budgets/${request.budgetId}/reject`)
  }
}
