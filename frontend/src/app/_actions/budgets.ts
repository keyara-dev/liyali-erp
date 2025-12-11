'use server'

import { cache } from 'react'
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

/**
 * Create a new budget draft
 */
export async function createBudget(
  request: CreateBudgetRequest
): Promise<APIResponse<Budget | null>> {
  try {
    const budgetNumber = `BDG-${new Date().getFullYear()}-${Math.floor(Math.random() * 10000).toString().padStart(5, '0')}`

    const newBudget: Budget = {
      id: `budget-${Date.now()}`,
      budgetNumber,
      name: request.name,
      description: request.description,
      department: request.department,
      departmentId: request.departmentId,
      fiscalYear: request.fiscalYear,
      totalAmount: request.totalAmount,
      currency: request.currency,
      items: request.items.map((item) => ({
        ...item,
        id: `item-${Date.now()}-${Math.random()}`,
        spentAmount: 0,
        remainingAmount: item.allocatedAmount,
        createdAt: new Date(),
        updatedAt: new Date()
      })),
      status: 'DRAFT',
      createdBy: request.createdBy,
      createdAt: new Date(),
      updatedAt: new Date(),
      currentApprovalStage: 0,
      totalApprovalStages: 4, // Department Head, Finance Officer, Director, CFO
      approvalChain: []
    }

    // Note: This is a server action but we can't access localStorage directly
    // The client component will handle storing in localStorage after receiving the response
    return {
      success: true,
      message: 'Budget created successfully',
      data: newBudget,
      status: 201,
      statusText: 'CREATED'
    }
  } catch (error: any) {
    return {
      success: false,
      message: error.message || 'Failed to create budget',
      data: null,
      status: 500,
      statusText: 'ERROR'
    }
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
    // In production, fetch from database
    const budgets = await getBudgets('user-1') // dummy userId
    const budgetIndex = budgets.data?.findIndex((b) => b.id === budgetId) ?? -1

    if (budgetIndex === -1) {
      return {
        success: false,
        message: 'Budget not found',
        data: null,
        status: 404,
        statusText: 'NOT_FOUND'
      }
    }

    // Merge updates with existing budget
    const updatedBudget: Budget = {
      ...(budgets.data![budgetIndex]),
      ...updates,
      updatedAt: new Date(),
    }

    return {
      success: true,
      message: 'Budget updated successfully',
      data: updatedBudget,
      status: 200,
      statusText: 'OK'
    }
  } catch (error: any) {
    return {
      success: false,
      message: error.message || 'Failed to update budget',
      data: null,
      status: 500,
      statusText: 'ERROR'
    }
  }
}

/**
 * Get all budgets with optional filters
 */
export const getBudgets = cache(async (
  userId: string,
  filters?: BudgetFilters
): Promise<APIResponse<Budget[] | null>> => {
  try {
    // Mock data - in production, fetch from database
    const mockBudgets: Budget[] = [
      {
        id: 'budget-1',
        budgetNumber: 'BDG-2024-00001',
        name: 'IT Department Annual Budget 2024',
        description: 'Annual budget allocation for IT department',
        department: 'Information Technology',
        departmentId: 'dept-it',
        fiscalYear: '2024',
        totalAmount: 150000,
        currency: 'USD',
        items: [
          {
            id: 'item-1',
            category: 'Hardware',
            description: 'Computers and servers',
            allocatedAmount: 50000,
            spentAmount: 35000,
            remainingAmount: 15000,
            createdAt: new Date(),
            updatedAt: new Date()
          },
          {
            id: 'item-2',
            category: 'Software',
            description: 'Software licenses',
            allocatedAmount: 40000,
            spentAmount: 20000,
            remainingAmount: 20000,
            createdAt: new Date(),
            updatedAt: new Date()
          },
          {
            id: 'item-3',
            category: 'Infrastructure',
            description: 'Network and infrastructure',
            allocatedAmount: 60000,
            spentAmount: 15000,
            remainingAmount: 45000,
            createdAt: new Date(),
            updatedAt: new Date()
          }
        ],
        status: 'APPROVED',
        createdBy: userId,
        createdAt: new Date('2024-01-15'),
        updatedAt: new Date('2024-02-01'),
        submittedAt: new Date('2024-01-20'),
        approvedAt: new Date('2024-02-01'),
        currentApprovalStage: 4,
        totalApprovalStages: 4,
        approvalChain: [
          {
            stageNumber: 1,
            stageName: 'Department Head Review',
            assignedTo: 'manager-1',
            assignedRole: 'DEPARTMENT_MANAGER',
            status: 'APPROVED',
            actionTakenAt: new Date('2024-01-22'),
            actionTakenBy: 'manager-1',
            comments: 'Approved'
          },
          {
            stageNumber: 2,
            stageName: 'Finance Officer Review',
            assignedTo: 'finance-1',
            assignedRole: 'FINANCE_OFFICER',
            status: 'APPROVED',
            actionTakenAt: new Date('2024-01-25'),
            actionTakenBy: 'finance-1',
            comments: 'Budget verified'
          },
          {
            stageNumber: 3,
            stageName: 'Director Finance Review',
            assignedTo: 'director-1',
            assignedRole: 'DIRECTOR',
            status: 'APPROVED',
            actionTakenAt: new Date('2024-01-28'),
            actionTakenBy: 'director-1',
            comments: 'Approved'
          },
          {
            stageNumber: 4,
            stageName: 'CFO Final Approval',
            assignedTo: 'cfo-1',
            assignedRole: 'CFO',
            status: 'APPROVED',
            actionTakenAt: new Date('2024-02-01'),
            actionTakenBy: 'cfo-1',
            comments: 'Final approval granted'
          }
        ]
      },
      {
        id: 'budget-2',
        budgetNumber: 'BDG-2024-00002',
        name: 'Operations Department Budget 2024',
        description: 'Annual budget for operations',
        department: 'Operations',
        departmentId: 'dept-ops',
        fiscalYear: '2024',
        totalAmount: 200000,
        currency: 'USD',
        items: [
          {
            id: 'item-4',
            category: 'Personnel',
            description: 'Staff salaries and benefits',
            allocatedAmount: 120000,
            spentAmount: 0,
            remainingAmount: 120000,
            createdAt: new Date(),
            updatedAt: new Date()
          },
          {
            id: 'item-5',
            category: 'Operations',
            description: 'Operational expenses',
            allocatedAmount: 80000,
            spentAmount: 0,
            remainingAmount: 80000,
            createdAt: new Date(),
            updatedAt: new Date()
          }
        ],
        status: 'IN_REVIEW',
        createdBy: userId,
        createdAt: new Date('2024-02-01'),
        updatedAt: new Date('2024-02-05'),
        submittedAt: new Date('2024-02-02'),
        currentApprovalStage: 2,
        totalApprovalStages: 4,
        approvalChain: [
          {
            stageNumber: 1,
            stageName: 'Department Head Review',
            assignedTo: 'manager-2',
            assignedRole: 'DEPARTMENT_MANAGER',
            status: 'APPROVED',
            actionTakenAt: new Date('2024-02-03'),
            actionTakenBy: 'manager-2'
          },
          {
            stageNumber: 2,
            stageName: 'Finance Officer Review',
            assignedTo: 'finance-1',
            assignedRole: 'FINANCE_OFFICER',
            status: 'PENDING'
          }
        ]
      },
      {
        id: 'budget-3',
        budgetNumber: 'BDG-2024-00003',
        name: 'HR Department Budget 2024',
        description: 'Human resources department budget',
        department: 'Human Resources',
        departmentId: 'dept-hr',
        fiscalYear: '2024',
        totalAmount: 100000,
        currency: 'USD',
        items: [
          {
            id: 'item-6',
            category: 'Training',
            description: 'Employee training and development',
            allocatedAmount: 50000,
            spentAmount: 0,
            remainingAmount: 50000,
            createdAt: new Date(),
            updatedAt: new Date()
          },
          {
            id: 'item-7',
            category: 'Recruitment',
            description: 'Recruitment and hiring',
            allocatedAmount: 50000,
            spentAmount: 0,
            remainingAmount: 50000,
            createdAt: new Date(),
            updatedAt: new Date()
          }
        ],
        status: 'DRAFT',
        createdBy: userId,
        createdAt: new Date('2024-02-05'),
        updatedAt: new Date('2024-02-05'),
        currentApprovalStage: 0,
        totalApprovalStages: 4
      }
    ]

    return {
      success: true,
      message: 'Budgets retrieved successfully',
      data: mockBudgets,
      status: 200,
      statusText: 'OK'
    }
  } catch (error: any) {
    return {
      success: false,
      message: error.message || 'Failed to retrieve budgets',
      data: null,
      status: 500,
      statusText: 'ERROR'
    }
  }
})

/**
 * Get budget by ID
 * Note: This is a temporary implementation that checks mock data.
 * In production, this would fetch from the database.
 * For now, newly created budgets are stored in localStorage on the client.
 */
export async function getBudgetById(budgetId: string): Promise<APIResponse<Budget | null>> {
  try {
    // First try to get all budgets (both mock and localStorage-based)
    // In production, fetch from database with the actual userId
    const budgets = await getBudgets('user-1') // dummy userId for mock data
    let budget = budgets.data?.find((b) => b.id === budgetId)

    if (budget) {
      return {
        success: true,
        message: 'Budget retrieved successfully',
        data: budget,
        status: 200,
        statusText: 'OK'
      }
    }

    // If not found in mock data, return not found
    // Client components can fall back to localStorage if needed
    return {
      success: false,
      message: 'Budget not found',
      data: null,
      status: 404,
      statusText: 'NOT_FOUND'
    }
  } catch (error: any) {
    return {
      success: false,
      message: error.message || 'Failed to retrieve budget',
      data: null,
      status: 500,
      statusText: 'ERROR'
    }
  }
}

/**
 * Submit budget for approval
 *
 * BACKEND API INTEGRATION:
 * To integrate with a backend API, replace the implementation below with:
 *
 * const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/budgets/${request.budgetId}/submit`, {
 *   method: 'POST',
 *   headers: {
 *     'Content-Type': 'application/json',
 *     'Authorization': `Bearer ${token}`
 *   },
 *   body: JSON.stringify({ submittingUserId: request.submittingUserId })
 * })
 *
 * The server should handle updating the budget status to SUBMITTED and initializing the approval chain.
 * The response should contain the updated budget with approval chain data.
 */
export async function submitBudgetForApproval(
  request: SubmitBudgetRequest
): Promise<APIResponse<Budget | null>> {
  try {
    // TODO: BACKEND INTEGRATION - Replace with API call above
    // Currently uses localStorage as single source of truth
    const budget = await getBudgetById(request.budgetId)

    let budgetData = budget.data

    // If not found in mock data, we'll create a minimal response
    // In production, this would fetch from database
    if (!budgetData) {
      // Return success but without full data - client will use localStorage version
      return {
        success: true,
        message: 'Budget submitted for approval',
        data: {
          id: request.budgetId,
          status: 'SUBMITTED',
          submittedAt: new Date(),
          updatedAt: new Date(),
          currentApprovalStage: 1,
          approvalChain: [
            {
              stageNumber: 1,
              stageName: 'Department Head Review',
              assignedTo: 'manager-1',
              assignedRole: 'DEPARTMENT_MANAGER',
              status: 'PENDING'
            }
          ]
        } as any,
        status: 200,
        statusText: 'OK'
      }
    }

    const updatedBudget: Budget = {
      ...budgetData,
      status: 'SUBMITTED',
      submittedAt: new Date(),
      updatedAt: new Date(),
      currentApprovalStage: 1,
      approvalChain: [
        {
          stageNumber: 1,
          stageName: 'Department Head Review',
          assignedTo: 'manager-1',
          assignedRole: 'DEPARTMENT_MANAGER',
          status: 'PENDING'
        }
      ]
    }

    return {
      success: true,
      message: 'Budget submitted for approval',
      data: updatedBudget,
      status: 200,
      statusText: 'OK'
    }
  } catch (error: any) {
    return {
      success: false,
      message: error.message || 'Failed to submit budget',
      data: null,
      status: 500,
      statusText: 'ERROR'
    }
  }
}

/**
 * Approve budget
 *
 * BACKEND API INTEGRATION:
 * To integrate with a backend API, replace with:
 *
 * const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/budgets/${request.budgetId}/approve`, {
 *   method: 'POST',
 *   headers: {
 *     'Content-Type': 'application/json',
 *     'Authorization': `Bearer ${token}`
 *   },
 *   body: JSON.stringify({
 *     approvingUserId: request.approvingUserId,
 *     approvingUserRole: request.approvingUserRole,
 *     stageNumber: request.stageNumber,
 *     comments: request.comments,
 *     signature: request.signature
 *   })
 * })
 *
 * The server should update the approval chain and move to next stage or mark as APPROVED if final stage.
 */
export async function approveBudget(
  request: ApproveBudgetRequest
): Promise<APIResponse<Budget | null>> {
  try {
    // TODO: BACKEND INTEGRATION - Replace with API call above
    // Currently uses localStorage as single source of truth
    const budget = await getBudgetById(request.budgetId)
    let budgetData = budget.data

    if (!budgetData) {
      // Budget not found in mock data - return success with minimal data
      // Client will use localStorage version
      return {
        success: true,
        message: 'Budget approved',
        data: {
          id: request.budgetId,
          status: 'IN_REVIEW',
          currentApprovalStage: 2
        } as any,
        status: 200,
        statusText: 'OK'
      }
    }

    const currentStage = request.stageNumber || budgetData.currentApprovalStage || 1
    const isLastStage = currentStage >= (budgetData.totalApprovalStages || 4)

    const updatedBudget: Budget = {
      ...budgetData,
      status: isLastStage ? 'APPROVED' : 'IN_REVIEW',
      approvedAt: isLastStage ? new Date() : budgetData.approvedAt,
      currentApprovalStage: isLastStage ? currentStage : currentStage + 1,
      updatedAt: new Date(),
      approvalChain: [
        ...(budgetData.approvalChain || []),
        {
          stageNumber: currentStage,
          stageName: `Stage ${currentStage} Approval`,
          assignedTo: request.approvingUserId,
          assignedRole: request.approvingUserRole,
          status: 'APPROVED',
          actionTakenAt: new Date(),
          actionTakenBy: request.approvingUserId,
          comments: request.comments
        }
      ]
    }

    return {
      success: true,
      message: isLastStage ? 'Budget approved' : 'Budget moved to next stage',
      data: updatedBudget,
      status: 200,
      statusText: 'OK'
    }
  } catch (error: any) {
    return {
      success: false,
      message: error.message || 'Failed to approve budget',
      data: null,
      status: 500,
      statusText: 'ERROR'
    }
  }
}

/**
 * Reject budget
 *
 * BACKEND API INTEGRATION:
 * To integrate with a backend API, replace with:
 *
 * const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/budgets/${request.budgetId}/reject`, {
 *   method: 'POST',
 *   headers: {
 *     'Content-Type': 'application/json',
 *     'Authorization': `Bearer ${token}`
 *   },
 *   body: JSON.stringify({
 *     rejectingUserId: request.rejectingUserId,
 *     rejectingUserRole: request.rejectingUserRole,
 *     rejectionReason: request.rejectionReason,
 *     comments: request.comments
 *   })
 * })
 *
 * The server should update the budget status to REJECTED, record the rejection reason,
 * and add a rejection record to the approval chain.
 */
export async function rejectBudget(
  request: RejectBudgetRequest
): Promise<APIResponse<Budget | null>> {
  try {
    // TODO: BACKEND INTEGRATION - Replace with API call above
    // Currently uses localStorage as single source of truth
    const budget = await getBudgetById(request.budgetId)
    let budgetData = budget.data

    if (!budgetData) {
      // Budget not found in mock data - return success with minimal data
      // Client will use localStorage version
      return {
        success: true,
        message: 'Budget rejected',
        data: {
          id: request.budgetId,
          status: 'REJECTED',
          currentApprovalStage: 0
        } as any,
        status: 200,
        statusText: 'OK'
      }
    }

    const updatedBudget: Budget = {
      ...budgetData,
      status: 'REJECTED',
      rejectedAt: new Date(),
      rejectionReason: request.rejectionReason,
      currentApprovalStage: 0,
      updatedAt: new Date(),
      approvalChain: [
        ...(budgetData.approvalChain || []),
        {
          stageNumber: budgetData.currentApprovalStage || 1,
          stageName: `Stage ${budgetData.currentApprovalStage || 1} Review`,
          assignedTo: request.rejectingUserId,
          assignedRole: 'REVIEWER',
          status: 'REJECTED',
          actionTakenAt: new Date(),
          actionTakenBy: request.rejectingUserId,
          comments: request.comments
        }
      ]
    }

    return {
      success: true,
      message: 'Budget rejected',
      data: updatedBudget,
      status: 200,
      statusText: 'OK'
    }
  } catch (error: any) {
    return {
      success: false,
      message: error.message || 'Failed to reject budget',
      data: null,
      status: 500,
      statusText: 'ERROR'
    }
  }
}
