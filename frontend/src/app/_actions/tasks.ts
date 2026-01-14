"use server";

import { cache } from "react";
import { APIResponse } from "@/types";
import {
  Task,
  TaskType,
  TaskStatus,
  TaskPriority,
  TaskStats,
} from "@/types/tasks";
import authenticatedApiClient from "./api-config";
import { successResponse, handleError } from "./api-config";

/**
 * Get all tasks for a user with optional filters
 */
export const getTasksForUser = cache(
  async (userId: string, status?: TaskStatus): Promise<APIResponse<Task[]>> => {
    try {
      // Mock data - in production, fetch from database
      const mockTasks: Task[] = [
        {
          id: "task-1",
          taskType: "BUDGET_APPROVAL",
          title: "Approve IT Department Budget 2024",
          description:
            "Review and approve the annual budget allocation for IT department",
          assignedTo: userId,
          assignedRole: "FINANCE_OFFICER",
          status: "PENDING",
          priority: "HIGH",
          documentType: "Budget",
          documentId: "budget-1",
          documentNumber: "BDG-2024-00001",
          createdAt: new Date("2024-02-05"),
          dueDate: new Date("2024-02-15"),
          metadata: {
            currentApprovalStage: 1,
            totalApprovalStages: 4,
            approvalStageName: "Stage Name",
          },
        },
        {
          id: "task-2",
          taskType: "REQUISITION_APPROVAL",
          title: "Approve Purchase Requisition PR-2024-001",
          description:
            "Review and approve the purchase requisition for office equipment",
          assignedTo: userId,
          assignedRole: "DEPARTMENT_MANAGER",
          status: "PENDING",
          priority: "MEDIUM",
          documentType: "Requisition",
          documentId: "req-1001",
          documentNumber: "PR-2024-001",
          createdAt: new Date("2024-02-06"),
          dueDate: new Date("2024-02-20"),
          metadata: {
            currentApprovalStage: 1,
            totalApprovalStages: 4,
            approvalStageName: "Stage Name",
          },
        },
        {
          id: "task-3",
          taskType: "PURCHASE_ORDER_APPROVAL",
          title: "Approve Purchase Order PO-2024-042",
          description:
            "Review and approve the purchase order for software licenses",
          assignedTo: userId,
          assignedRole: "FINANCE_OFFICER",
          status: "PENDING",
          priority: "MEDIUM",
          documentType: "PurchaseOrder",
          documentId: "po-2024-042",
          documentNumber: "PO-2024-042",
          createdAt: new Date("2024-02-07"),
          dueDate: new Date("2024-02-17"),
          metadata: {
            currentApprovalStage: 1,
            totalApprovalStages: 4,
            approvalStageName: "Stage Name",
          },
        },
        {
          id: "task-4",
          taskType: "PAYMENT_VOUCHER_APPROVAL",
          title: "Approve Payment Voucher PV-2024-125",
          description:
            "Review and approve payment for vendor invoice INV-20240205",
          assignedTo: userId,
          assignedRole: "DIRECTOR",
          status: "PENDING",
          priority: "HIGH",
          documentType: "PaymentVoucher",
          documentId: "pv-2024-125",
          documentNumber: "PV-2024-125",
          createdAt: new Date("2024-02-08"),
          dueDate: new Date("2024-02-12"),
          metadata: {
            currentApprovalStage: 1,
            totalApprovalStages: 4,
            approvalStageName: "Stage Name",
          },
        },
        {
          id: "task-5",
          taskType: "GOODS_RECEIVED_NOTE_CONFIRMATION",
          title: "Confirm Goods Received Note GRN-2024-89",
          description: "Verify receipt of goods for purchase order PO-2024-035",
          assignedTo: userId,
          assignedRole: "WAREHOUSE_MANAGER",
          status: "PENDING",
          priority: "MEDIUM",
          documentType: "GoodsReceivedNote",
          documentId: "grn-2024-89",
          documentNumber: "GRN-2024-89",
          createdAt: new Date("2024-02-09"),
          dueDate: new Date("2024-02-14"),
          metadata: {
            currentApprovalStage: 1,
            totalApprovalStages: 4,
            approvalStageName: "Stage Name",
          },
        },
        {
          id: "task-6",
          taskType: "BUDGET_APPROVAL",
          title: "Approve HR Department Budget 2024",
          description:
            "Review and approve the annual budget allocation for HR department",
          assignedTo: userId,
          assignedRole: "CFO",
          status: "PENDING",
          priority: "URGENT",
          documentType: "Budget",
          documentId: "budget-3",
          documentNumber: "BDG-2024-00003",
          createdAt: new Date("2024-02-10"),
          dueDate: new Date("2024-02-11"),
          metadata: {
            currentApprovalStage: 1,
            totalApprovalStages: 4,
            approvalStageName: "Stage Name",
          },
        },
        {
          id: "task-7",
          taskType: "REQUISITION_APPROVAL",
          title: "Approve Purchase Requisition PR-2024-002",
          description:
            "Review and approve the purchase requisition for IT infrastructure",
          assignedTo: userId,
          assignedRole: "FINANCE_OFFICER",
          status: "IN_PROGRESS",
          priority: "MEDIUM",
          documentType: "Requisition",
          documentId: "req-1002",
          documentNumber: "PR-2024-002",
          createdAt: new Date("2024-02-03"),
          dueDate: new Date("2024-02-16"),
          metadata: {
            currentApprovalStage: 1,
            totalApprovalStages: 4,
            approvalStageName: "Stage Name",
          },
        },
      ];

      // Filter by status if provided
      const filtered = status
        ? mockTasks.filter((t) => t.status === status)
        : mockTasks;

      return {
        success: true,
        message: "Tasks retrieved successfully",
        data: filtered,
        status: 200,
        statusText: "OK",
      };
    } catch (error: any) {
      return {
        success: false,
        message: error.message || "Failed to retrieve tasks",
        data: undefined,
        status: 500,
        statusText: "ERROR",
      };
    }
  }
);

/**
 * Get task statistics for a user
 */
export async function getTaskStats(
  userId: string
): Promise<APIResponse<TaskStats>> {
  const url = `/api/v1/approvals/stats`;

  try {
    const response = await authenticatedApiClient({
      method: "GET",
      url,
    });

    return successResponse(
      response.data?.data || {
        totalTasks: 0,
        pendingTasks: 0,
        completedTasks: 0,
        overdueTasks: 0,
        highPriorityTasks: 0,
        byType: {},
        byPriority: {},
      },
      "Task statistics retrieved successfully"
    );
  } catch (error: any) {
    return handleError(error, "GET", url);
  }
}

/**
 * Get task by ID with full details
 */
export async function getTaskById(taskId: string): Promise<APIResponse<Task>> {
  try {
    const tasksResponse = await getTasksForUser("user-1"); // dummy userId
    const task = tasksResponse.data?.find((t) => t.id === taskId);

    if (!task) {
      return {
        success: false,
        message: "Task not found",
        data: undefined,
        status: 404,
        statusText: "NOT_FOUND",
      };
    }

    return {
      success: true,
      message: "Task retrieved successfully",
      data: task,
      status: 200,
      statusText: "OK",
    };
  } catch (error: any) {
    return {
      success: false,
      message: error.message || "Failed to retrieve task",
      data: undefined,
      status: 500,
      statusText: "ERROR",
    };
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
    const taskResponse = await getTaskById(taskId);

    if (!taskResponse.success || !taskResponse.data) {
      return {
        success: false,
        message: "Task not found",
        data: undefined,
        status: 404,
        statusText: "NOT_FOUND",
      };
    }

    const updatedTask: Task = {
      ...taskResponse.data,
      status: "COMPLETED",
      completedAt: new Date(),
      completedBy: userId,
    };

    return {
      success: true,
      message: "Task completed successfully",
      data: updatedTask,
      status: 200,
      statusText: "OK",
    };
  } catch (error: any) {
    return {
      success: false,
      message: error.message || "Failed to complete task",
      data: undefined,
      status: 500,
      statusText: "ERROR",
    };
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
    const taskResponse = await getTaskById(taskId);

    if (!taskResponse.success || !taskResponse.data) {
      return {
        success: false,
        message: "Task not found",
        data: undefined,
        status: 404,
        statusText: "NOT_FOUND",
      };
    }

    const updatedTask: Task = {
      ...taskResponse.data,
      status: "IN_PROGRESS",
    };

    return {
      success: true,
      message: "Task started successfully",
      data: updatedTask,
      status: 200,
      statusText: "OK",
    };
  } catch (error: any) {
    return {
      success: false,
      message: error.message || "Failed to start task",
      data: undefined,
      status: 500,
      statusText: "ERROR",
    };
  }
}
