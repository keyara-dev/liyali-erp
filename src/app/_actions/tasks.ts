'use server'

import { cache } from 'react'
import { APIResponse } from '@/types'
import {
  Task,
  TaskType,
  TaskStatus,
  TaskPriority,
  TaskStats
} from '@/types/tasks'

/**
 * Get all tasks for a user with optional filters
 */
export const getTasksForUser = cache(async (
  userId: string,
  status?: TaskStatus
): Promise<APIResponse<Task[]>> => {
  try {
    // Mock data - in production, fetch from database
    const mockTasks: Task[] = [
      {
        id: 'task-1',
        type: 'BUDGET_APPROVAL',
        title: 'Approve IT Department Budget 2024',
        description: 'Review and approve the annual budget allocation for IT department',
        assignedTo: userId,
        assignedRole: 'FINANCE_OFFICER',
        status: 'PENDING',
        priority: 'HIGH',
        documentType: 'Budget',
        documentId: 'budget-1',
        documentNumber: 'BDG-2024-00001',
        createdAt: new Date('2024-02-05'),
        dueAt: new Date('2024-02-15'),
        actionUrl: '/workflows/budgets/budget-1'
      },
      {
        id: 'task-2',
        type: 'REQUISITION_APPROVAL',
        title: 'Approve Purchase Requisition PR-2024-001',
        description: 'Review and approve the purchase requisition for office equipment',
        assignedTo: userId,
        assignedRole: 'DEPARTMENT_MANAGER',
        status: 'PENDING',
        priority: 'MEDIUM',
        documentType: 'Requisition',
        documentId: 'req-1001',
        documentNumber: 'PR-2024-001',
        createdAt: new Date('2024-02-06'),
        dueAt: new Date('2024-02-20'),
        actionUrl: '/workflows/requisitions/req-1001'
      },
      {
        id: 'task-3',
        type: 'PURCHASE_ORDER_APPROVAL',
        title: 'Approve Purchase Order PO-2024-042',
        description: 'Review and approve the purchase order for software licenses',
        assignedTo: userId,
        assignedRole: 'FINANCE_OFFICER',
        status: 'PENDING',
        priority: 'MEDIUM',
        documentType: 'PurchaseOrder',
        documentId: 'po-2024-042',
        documentNumber: 'PO-2024-042',
        createdAt: new Date('2024-02-07'),
        dueAt: new Date('2024-02-17'),
        actionUrl: '/workflows/purchase-orders/po-2024-042'
      },
      {
        id: 'task-4',
        type: 'PAYMENT_VOUCHER_APPROVAL',
        title: 'Approve Payment Voucher PV-2024-125',
        description: 'Review and approve payment for vendor invoice INV-20240205',
        assignedTo: userId,
        assignedRole: 'DIRECTOR',
        status: 'PENDING',
        priority: 'HIGH',
        documentType: 'PaymentVoucher',
        documentId: 'pv-2024-125',
        documentNumber: 'PV-2024-125',
        createdAt: new Date('2024-02-08'),
        dueAt: new Date('2024-02-12'),
        actionUrl: '/workflows/payment-vouchers/pv-2024-125'
      },
      {
        id: 'task-5',
        type: 'GOODS_RECEIVED_NOTE_CONFIRMATION',
        title: 'Confirm Goods Received Note GRN-2024-89',
        description: 'Verify receipt of goods for purchase order PO-2024-035',
        assignedTo: userId,
        assignedRole: 'WAREHOUSE_MANAGER',
        status: 'PENDING',
        priority: 'MEDIUM',
        documentType: 'GoodsReceivedNote',
        documentId: 'grn-2024-89',
        documentNumber: 'GRN-2024-89',
        createdAt: new Date('2024-02-09'),
        dueAt: new Date('2024-02-14'),
        actionUrl: '/workflows/grn/grn-2024-89'
      },
      {
        id: 'task-6',
        type: 'BUDGET_APPROVAL',
        title: 'Approve HR Department Budget 2024',
        description: 'Review and approve the annual budget allocation for HR department',
        assignedTo: userId,
        assignedRole: 'CFO',
        status: 'PENDING',
        priority: 'URGENT',
        documentType: 'Budget',
        documentId: 'budget-3',
        documentNumber: 'BDG-2024-00003',
        createdAt: new Date('2024-02-10'),
        dueAt: new Date('2024-02-11'),
        actionUrl: '/workflows/budgets/budget-3'
      },
      {
        id: 'task-7',
        type: 'REQUISITION_APPROVAL',
        title: 'Approve Purchase Requisition PR-2024-002',
        description: 'Review and approve the purchase requisition for IT infrastructure',
        assignedTo: userId,
        assignedRole: 'FINANCE_OFFICER',
        status: 'IN_PROGRESS',
        priority: 'MEDIUM',
        documentType: 'Requisition',
        documentId: 'req-1002',
        documentNumber: 'PR-2024-002',
        createdAt: new Date('2024-02-03'),
        dueAt: new Date('2024-02-16'),
        actionUrl: '/workflows/requisitions/req-1002'
      }
    ]

    // Filter by status if provided
    const filtered = status
      ? mockTasks.filter(t => t.status === status)
      : mockTasks

    return {
      success: true,
      message: 'Tasks retrieved successfully',
      data: filtered,
      status: 200,
      statusText: 'OK'
    }
  } catch (error: any) {
    return {
      success: false,
      message: error.message || 'Failed to retrieve tasks',
      data: null,
      status: 500,
      statusText: 'ERROR'
    }
  }
})

/**
 * Get task statistics for a user
 */
export async function getTaskStats(userId: string): Promise<APIResponse<TaskStats>> {
  try {
    const tasksResponse = await getTasksForUser(userId)

    if (!tasksResponse.success || !tasksResponse.data) {
      return {
        success: false,
        message: 'Failed to calculate task stats',
        data: null,
        status: 500,
        statusText: 'ERROR'
      }
    }

    const tasks = tasksResponse.data
    const now = new Date()

    const stats: TaskStats = {
      totalTasks: tasks.length,
      pendingTasks: tasks.filter(t => t.status === 'PENDING').length,
      inProgressTasks: tasks.filter(t => t.status === 'IN_PROGRESS').length,
      completedTasks: tasks.filter(t => t.status === 'COMPLETED').length,
      overdueTasks: tasks.filter(
        t => t.dueAt && t.dueAt < now && t.status !== 'COMPLETED'
      ).length,
      urgentTasks: tasks.filter(
        t => t.priority === 'URGENT' && t.status !== 'COMPLETED'
      ).length,
      tasksByType: {
        BUDGET_APPROVAL: tasks.filter(t => t.type === 'BUDGET_APPROVAL').length,
        REQUISITION_APPROVAL: tasks.filter(t => t.type === 'REQUISITION_APPROVAL').length,
        PURCHASE_ORDER_APPROVAL: tasks.filter(t => t.type === 'PURCHASE_ORDER_APPROVAL').length,
        PAYMENT_VOUCHER_APPROVAL: tasks.filter(t => t.type === 'PAYMENT_VOUCHER_APPROVAL').length,
        GOODS_RECEIVED_NOTE_CONFIRMATION: tasks.filter(
          t => t.type === 'GOODS_RECEIVED_NOTE_CONFIRMATION'
        ).length
      }
    }

    return {
      success: true,
      message: 'Task statistics calculated successfully',
      data: stats,
      status: 200,
      statusText: 'OK'
    }
  } catch (error: any) {
    return {
      success: false,
      message: error.message || 'Failed to calculate task stats',
      data: null,
      status: 500,
      statusText: 'ERROR'
    }
  }
}

/**
 * Get task by ID with full details
 */
export async function getTaskById(taskId: string): Promise<APIResponse<Task>> {
  try {
    const tasksResponse = await getTasksForUser('user-1') // dummy userId
    const task = tasksResponse.data?.find(t => t.id === taskId)

    if (!task) {
      return {
        success: false,
        message: 'Task not found',
        data: null,
        status: 404,
        statusText: 'NOT_FOUND'
      }
    }

    return {
      success: true,
      message: 'Task retrieved successfully',
      data: task,
      status: 200,
      statusText: 'OK'
    }
  } catch (error: any) {
    return {
      success: false,
      message: error.message || 'Failed to retrieve task',
      data: null,
      status: 500,
      statusText: 'ERROR'
    }
  }
}

/**
 * Complete a task (mark as completed)
 */
export async function completeTask(
  taskId: string,
  userId: string
): Promise<APIResponse<Task>> {
  try {
    const taskResponse = await getTaskById(taskId)

    if (!taskResponse.success || !taskResponse.data) {
      return {
        success: false,
        message: 'Task not found',
        data: null,
        status: 404,
        statusText: 'NOT_FOUND'
      }
    }

    const updatedTask: Task = {
      ...taskResponse.data,
      status: 'COMPLETED',
      completedAt: new Date(),
      completedBy: userId
    }

    return {
      success: true,
      message: 'Task completed successfully',
      data: updatedTask,
      status: 200,
      statusText: 'OK'
    }
  } catch (error: any) {
    return {
      success: false,
      message: error.message || 'Failed to complete task',
      data: null,
      status: 500,
      statusText: 'ERROR'
    }
  }
}

/**
 * Start a task (mark as in progress)
 */
export async function startTask(
  taskId: string,
  userId: string
): Promise<APIResponse<Task>> {
  try {
    const taskResponse = await getTaskById(taskId)

    if (!taskResponse.success || !taskResponse.data) {
      return {
        success: false,
        message: 'Task not found',
        data: null,
        status: 404,
        statusText: 'NOT_FOUND'
      }
    }

    const updatedTask: Task = {
      ...taskResponse.data,
      status: 'IN_PROGRESS',
      startedAt: new Date(),
      startedBy: userId
    }

    return {
      success: true,
      message: 'Task started successfully',
      data: updatedTask,
      status: 200,
      statusText: 'OK'
    }
  } catch (error: any) {
    return {
      success: false,
      message: error.message || 'Failed to start task',
      data: null,
      status: 500,
      statusText: 'ERROR'
    }
  }
}
