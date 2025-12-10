'use server';

import { cache } from 'react';
import {
  Requisition,
  RequisitionStatus,
  CreateRequisitionRequest,
  UpdateRequisitionRequest,
  SubmitRequisitionRequest,
  ApproveRequisitionRequest,
  RejectRequisitionRequest,
  RequisitionStats,
} from '@/types/requisition';
import { APIResponse } from '@/types';
import { createPurchaseOrderFromRequisition } from './purchase-orders';
import { documentStore } from '@/lib/workflow-stores';
import { WorkflowDocument } from '@/types/workflow';

/**
 * Mock requisitions database
 * In production, replace with actual database queries
 */
let mockRequisitions: Requisition[] = [
  {
    id: 'req-1001',
    requisitionNumber: 'REQ-2024-001',
    title: 'Office Supplies Purchase',
    description: 'Monthly office supplies including paper, pens, and stationery',
    department: 'Administrative',
    departmentId: 'dept-admin',
    requestedBy: 'user-001',
    requestedByName: 'John Smith',
    requestedByRole: 'REQUESTER',
    requestedDate: new Date('2024-11-20'),
    requiredByDate: new Date('2024-12-15'),
    priority: 'MEDIUM',
    status: 'IN_REVIEW',
    items: [
      {
        id: 'item-1',
        requisitionId: 'req-1001',
        itemNumber: 1,
        description: 'A4 Paper Reams (500 sheets)',
        category: 'Office Supplies',
        quantity: 50,
        unitPrice: 5.50,
        unit: 'reams',
        totalPrice: 275,
        createdAt: new Date(),
        updatedAt: new Date(),
      },
      {
        id: 'item-2',
        requisitionId: 'req-1001',
        itemNumber: 2,
        description: 'Blue Ballpoint Pens (Box of 50)',
        category: 'Office Supplies',
        quantity: 10,
        unitPrice: 12.00,
        unit: 'boxes',
        totalPrice: 120,
        createdAt: new Date(),
        updatedAt: new Date(),
      },
      {
        id: 'item-3',
        requisitionId: 'req-1001',
        itemNumber: 3,
        description: 'Sticky Notes (Pack of 12)',
        category: 'Office Supplies',
        quantity: 20,
        unitPrice: 8.50,
        unit: 'packs',
        totalPrice: 170,
        createdAt: new Date(),
        updatedAt: new Date(),
      },
    ],
    totalAmount: 565,
    currency: 'ZMW',
    currentApprovalStage: 1,
    totalApprovalStages: 2,
    approvalChain: [
      {
        stageNumber: 1,
        stageName: 'Department Manager',
        assignedTo: 'user-002',
        assignedRole: 'DEPARTMENT_MANAGER',
        status: 'PENDING',
      },
      {
        stageNumber: 2,
        stageName: 'Finance Officer',
        assignedTo: 'user-003',
        assignedRole: 'FINANCE_OFFICER',
        status: 'PENDING',
      },
    ],
    createdAt: new Date('2024-11-20'),
    updatedAt: new Date('2024-11-22'),
    submittedAt: new Date('2024-11-21'),
  },
  {
    id: 'req-1002',
    requisitionNumber: 'REQ-2024-002',
    title: 'IT Equipment - Laptops',
    description: 'Purchase of laptops for new team members',
    department: 'Information Technology',
    departmentId: 'dept-it',
    requestedBy: 'user-004',
    requestedByName: 'Sarah Johnson',
    requestedByRole: 'REQUESTER',
    requestedDate: new Date('2024-11-15'),
    requiredByDate: new Date('2024-12-10'),
    priority: 'URGENT',
    status: 'APPROVED',
    items: [
      {
        id: 'item-4',
        requisitionId: 'req-1002',
        itemNumber: 1,
        description: 'Dell XPS 15 Laptop (Intel i7, 16GB RAM)',
        category: 'Equipment',
        quantity: 3,
        unitPrice: 2500.00,
        unit: 'units',
        totalPrice: 7500,
        createdAt: new Date(),
        updatedAt: new Date(),
      },
    ],
    totalAmount: 7500,
    currency: 'ZMW',
    currentApprovalStage: 3,
    totalApprovalStages: 3,
    approvalChain: [
      {
        stageNumber: 1,
        stageName: 'Department Manager',
        assignedTo: 'user-002',
        assignedRole: 'DEPARTMENT_MANAGER',
        status: 'APPROVED',
        actionTakenBy: 'user-002',
        actionTakenByRole: 'DEPARTMENT_MANAGER',
        actionTakenAt: new Date('2024-11-16'),
        comments: 'Approved - urgent need for team expansion',
        signature: 'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==',
      },
      {
        stageNumber: 2,
        stageName: 'Finance Officer',
        assignedTo: 'user-003',
        assignedRole: 'FINANCE_OFFICER',
        status: 'APPROVED',
        actionTakenBy: 'user-003',
        actionTakenByRole: 'FINANCE_OFFICER',
        actionTakenAt: new Date('2024-11-17'),
        comments: 'Budget allocation confirmed',
        signature: 'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==',
      },
      {
        stageNumber: 3,
        stageName: 'Director',
        assignedTo: 'user-005',
        assignedRole: 'DIRECTOR',
        status: 'APPROVED',
        actionTakenBy: 'user-005',
        actionTakenByRole: 'DIRECTOR',
        actionTakenAt: new Date('2024-11-18'),
        comments: 'Final approval granted',
        signature: 'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==',
      },
    ],
    createdAt: new Date('2024-11-15'),
    updatedAt: new Date('2024-11-18'),
    submittedAt: new Date('2024-11-16'),
    approvedAt: new Date('2024-11-18'),
  },
  {
    id: 'req-1003',
    requisitionNumber: 'REQ-2024-003',
    title: 'Marketing Materials',
    description: 'Printing and design materials for marketing campaign',
    department: 'Marketing',
    departmentId: 'dept-marketing',
    requestedBy: 'user-006',
    requestedByName: 'Michael Chen',
    requestedByRole: 'REQUESTER',
    requestedDate: new Date('2024-11-19'),
    requiredByDate: new Date('2024-12-05'),
    priority: 'HIGH',
    status: 'REJECTED',
    items: [
      {
        id: 'item-5',
        requisitionId: 'req-1003',
        itemNumber: 1,
        description: 'Brochure Printing (5000 units)',
        category: 'Marketing Materials',
        quantity: 1,
        unitPrice: 800.00,
        unit: 'job',
        totalPrice: 800,
        createdAt: new Date(),
        updatedAt: new Date(),
      },
    ],
    totalAmount: 800,
    currency: 'ZMW',
    currentApprovalStage: 1,
    totalApprovalStages: 2,
    approvalChain: [
      {
        stageNumber: 1,
        stageName: 'Department Manager',
        assignedTo: 'user-002',
        assignedRole: 'DEPARTMENT_MANAGER',
        status: 'REJECTED',
        actionTakenBy: 'user-002',
        actionTakenByRole: 'DEPARTMENT_MANAGER',
        actionTakenAt: new Date('2024-11-20'),
        remarks: 'Budget allocation exceeded for this quarter. Please resubmit in Q1.',
        signature: 'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==',
      },
      {
        stageNumber: 2,
        stageName: 'Finance Officer',
        assignedTo: 'user-003',
        assignedRole: 'FINANCE_OFFICER',
        status: 'PENDING',
      },
    ],
    createdAt: new Date('2024-11-19'),
    updatedAt: new Date('2024-11-20'),
    submittedAt: new Date('2024-11-19'),
    rejectedAt: new Date('2024-11-20'),
  },
];

/**
 * Create a new requisition
 */
export async function createRequisition(
  data: CreateRequisitionRequest
): Promise<APIResponse<Requisition>> {
  try {
    const requisitionId = `req-${Date.now()}`;
    const requisitionNumber = `REQ-${new Date().getFullYear()}-${String(mockRequisitions.length + 1).padStart(3, '0')}`;

    // Calculate total amount
    const totalAmount = data.items.reduce((sum, item) => sum + item.totalPrice, 0);

    // Create requisition with approval chain (3 stages by default)
    const now = new Date();
    const requisition: Requisition = {
      id: requisitionId,
      requisitionNumber,
      title: data.title,
      description: data.description,
      department: data.department,
      departmentId: data.departmentId,
      requestedBy: data.createdBy,
      requestedByName: data.createdByName,
      requestedByRole: data.createdByRole,
      requestedDate: now,
      requiredByDate: new Date(data.requiredByDate),
      priority: data.priority,
      status: 'DRAFT',
      items: data.items.map((item, index) => ({
        ...item,
        id: `item-${requisitionId}-${index + 1}`,
        requisitionId,
        createdAt: now,
        updatedAt: now,
      })),
      totalAmount,
      currency: 'ZMW',
      currentApprovalStage: 0,
      totalApprovalStages: 3,
      approvalChain: [
        {
          stageNumber: 1,
          stageName: 'Department Manager',
          assignedTo: 'user-002',
          assignedRole: 'DEPARTMENT_MANAGER',
          status: 'PENDING',
        },
        {
          stageNumber: 2,
          stageName: 'Finance Officer',
          assignedTo: 'user-003',
          assignedRole: 'FINANCE_OFFICER',
          status: 'PENDING',
        },
        {
          stageNumber: 3,
          stageName: 'Director',
          assignedTo: 'user-005',
          assignedRole: 'DIRECTOR',
          status: 'PENDING',
        },
      ],
      budgetCode: data.budgetCode,
      costCenter: data.costCenter,
      projectCode: data.projectCode,
      createdAt: now,
      updatedAt: now,
      // Initialize action history with creation entry
      actionHistory: [
        {
          id: `action-${Date.now()}-1`,
          actionType: 'CREATE',
          performedBy: data.createdBy,
          performedByName: data.createdByName,
          performedByRole: data.createdByRole,
          performedAt: now,
          newStatus: 'DRAFT',
          comments: `Requisition created with ${data.items.length} item(s)`,
        },
      ],
    };

    mockRequisitions.push(requisition);

    return {
      success: true,
      message: 'Requisition created successfully',
      data: requisition,
      status: 201,
      statusText: 'CREATED',
    };
  } catch (error) {
    return {
      success: false,
      message: error instanceof Error ? error.message : 'Failed to create requisition',
      data: null,
      status: 500,
      statusText: 'ERROR',
    };
  }
}

/**
 * Get all requisitions (cached)
 * Includes both mock data and workflow document store data
 */
export const getRequisitions = cache(async (): Promise<APIResponse<Requisition[]>> => {
  try {
    let allRequisitions: Requisition[] = [...mockRequisitions];

    // Also fetch from workflow document store
    try {
      const workflowDocs = Array.from(documentStore.values()).filter(
        (doc: WorkflowDocument) => doc.type === 'REQUISITION'
      );

      // Convert workflow documents to requisitions
      const workflowRequisitions = workflowDocs.map((doc: WorkflowDocument): Requisition => ({
        id: doc.id,
        requisitionNumber: doc.documentNumber,
        title: doc.metadata?.title || doc.metadata?.requestedFor || 'Untitled Requisition',
        description: doc.metadata?.description || '',
        department: doc.metadata?.department || 'Unknown',
        departmentId: 'dept-unknown',
        requestedBy: doc.createdBy,
        requestedByName: doc.createdByUser?.name || 'Unknown User',
        requestedByRole: doc.createdByUser?.role || 'REQUESTER',
        requestedDate: doc.createdAt,
        requiredByDate: new Date(),
        priority: (doc.metadata?.priority as any) || 'MEDIUM',
        status: (doc.status as any) || 'DRAFT',
        items: (doc.metadata?.items || []).map((item: any, idx: number) => ({
          id: item.id || `item-${doc.id}-${idx}`,
          requisitionId: doc.id,
          itemNumber: idx + 1,
          description: item.itemDescription || item.description || '',
          category: 'General',
          quantity: item.quantity || 1,
          unitPrice: item.estimatedCost || 0,
          unit: 'pcs',
          totalPrice: (item.quantity || 1) * (item.estimatedCost || 0),
          createdAt: doc.createdAt,
          updatedAt: doc.updatedAt,
        })),
        totalAmount: doc.metadata?.amount || 0,
        currency: 'ZMW',
        currentApprovalStage: doc.currentStage || 0,
        totalApprovalStages: 3,
        budgetCode: doc.metadata?.budgetCode || '',
        createdAt: doc.createdAt,
        updatedAt: doc.updatedAt,
      }));

      // Merge requisitions, avoiding duplicates by ID
      const ids = new Set(allRequisitions.map(r => r.id));
      const newRequisitions = workflowRequisitions.filter(r => !ids.has(r.id));
      allRequisitions = [...allRequisitions, ...newRequisitions];
    } catch (storeError) {
      // If document store is not available, just use mock data
      console.log('Document store not available, using mock data only');
    }

    return {
      success: true,
      message: 'Requisitions retrieved successfully',
      data: allRequisitions,
      status: 200,
      statusText: 'OK',
    };
  } catch (error) {
    return {
      success: false,
      message: error instanceof Error ? error.message : 'Failed to fetch requisitions',
      data: null,
      status: 500,
      statusText: 'ERROR',
    };
  }
});

/**
 * Get requisition by ID
 */
export async function getRequisitionById(requisitionId: string): Promise<APIResponse<Requisition>> {
  try {
    const requisition = mockRequisitions.find((r) => r.id === requisitionId);

    if (!requisition) {
      return {
        success: false,
        message: 'Requisition not found',
        data: null,
        status: 404,
        statusText: 'NOT_FOUND',
      };
    }

    return {
      success: true,
      message: 'Requisition retrieved successfully',
      data: requisition,
      status: 200,
      statusText: 'OK',
    };
  } catch (error) {
    return {
      success: false,
      message: error instanceof Error ? error.message : 'Failed to fetch requisition',
      data: null,
      status: 500,
      statusText: 'ERROR',
    };
  }
}

/**
 * Update requisition
 */
export async function updateRequisition(
  data: UpdateRequisitionRequest
): Promise<APIResponse<Requisition>> {
  try {
    const requisition = mockRequisitions.find((r) => r.id === data.requisitionId);

    if (!requisition) {
      return {
        success: false,
        message: 'Requisition not found',
        data: null,
        status: 404,
        statusText: 'NOT_FOUND',
      };
    }

    if (requisition.status !== 'DRAFT') {
      return {
        success: false,
        message: 'Cannot update requisition that is not in DRAFT status',
        data: null,
        status: 400,
        statusText: 'BAD_REQUEST',
      };
    }

    if (data.title) requisition.title = data.title;
    if (data.description) requisition.description = data.description;
    if (data.requiredByDate) requisition.requiredByDate = new Date(data.requiredByDate);
    if (data.priority) requisition.priority = data.priority;
    if (data.items) {
      requisition.items = data.items.map((item) => ({
        ...item,
        requisitionId: requisition.id,
        createdAt: new Date(),
        updatedAt: new Date(),
      }));
      requisition.totalAmount = requisition.items.reduce((sum, item) => sum + item.totalPrice, 0);
    }
    if (data.budgetCode) requisition.budgetCode = data.budgetCode;
    if (data.costCenter) requisition.costCenter = data.costCenter;
    if (data.projectCode) requisition.projectCode = data.projectCode;
    requisition.updatedAt = new Date();

    return {
      success: true,
      message: 'Requisition updated successfully',
      data: requisition,
      status: 200,
      statusText: 'OK',
    };
  } catch (error) {
    return {
      success: false,
      message: error instanceof Error ? error.message : 'Failed to update requisition',
      data: null,
      status: 500,
      statusText: 'ERROR',
    };
  }
}

/**
 * Submit requisition for approval
 */
export async function submitRequisitionForApproval(
  data: SubmitRequisitionRequest
): Promise<APIResponse<Requisition>> {
  try {
    const requisition = mockRequisitions.find((r) => r.id === data.requisitionId);

    if (!requisition) {
      return {
        success: false,
        message: 'Requisition not found',
        data: null,
        status: 404,
        statusText: 'NOT_FOUND',
      };
    }

    if (requisition.status !== 'DRAFT') {
      return {
        success: false,
        message: 'Only DRAFT requisitions can be submitted',
        data: null,
        status: 400,
        statusText: 'BAD_REQUEST',
      };
    }

    if (!requisition.items || requisition.items.length === 0) {
      return {
        success: false,
        message: 'Requisition must have at least one item',
        data: null,
        status: 400,
        statusText: 'BAD_REQUEST',
      };
    }

    requisition.status = 'SUBMITTED';
    requisition.currentApprovalStage = 1;
    requisition.submittedAt = new Date();
    requisition.updatedAt = new Date();

    // Add action to history
    if (!requisition.actionHistory) {
      requisition.actionHistory = [];
    }
    requisition.actionHistory.push({
      id: `action-${Date.now()}-${requisition.actionHistory.length + 1}`,
      actionType: 'SUBMIT',
      performedBy: data.submittedBy,
      performedByName: data.submittedByName,
      performedByRole: data.submittedByRole,
      performedAt: requisition.submittedAt,
      previousStatus: 'DRAFT',
      newStatus: 'SUBMITTED',
      comments: data.comments || 'Requisition submitted for approval',
    });

    return {
      success: true,
      message: 'Requisition submitted for approval',
      data: requisition,
      status: 200,
      statusText: 'OK',
    };
  } catch (error) {
    return {
      success: false,
      message: error instanceof Error ? error.message : 'Failed to submit requisition',
      data: null,
      status: 500,
      statusText: 'ERROR',
    };
  }
}

/**
 * Approve requisition
 */
export async function approveRequisition(
  data: ApproveRequisitionRequest
): Promise<APIResponse<Requisition>> {
  try {
    const requisition = mockRequisitions.find((r) => r.id === data.requisitionId);

    if (!requisition) {
      return {
        success: false,
        message: 'Requisition not found',
        data: null,
        status: 404,
        statusText: 'NOT_FOUND',
      };
    }

    if (!requisition.approvalChain) {
      return {
        success: false,
        message: 'Approval chain not configured',
        data: null,
        status: 400,
        statusText: 'BAD_REQUEST',
      };
    }

    if (!data.signature) {
      return {
        success: false,
        message: 'Signature is required for approval',
        data: null,
        status: 400,
        statusText: 'BAD_REQUEST',
      };
    }

    const stage = requisition.approvalChain.find(
      (s) => s.stageNumber === (data.stageNumber || requisition.currentApprovalStage)
    );

    if (!stage) {
      return {
        success: false,
        message: 'Approval stage not found',
        data: null,
        status: 404,
        statusText: 'NOT_FOUND',
      };
    }

    if (stage.status !== 'PENDING') {
      return {
        success: false,
        message: `Stage already ${stage.status.toLowerCase()}`,
        data: null,
        status: 400,
        statusText: 'BAD_REQUEST',
      };
    }

    // Update stage
    stage.status = 'APPROVED';
    stage.actionTakenAt = new Date();
    stage.actionTakenBy = data.approvingUserId;
    stage.actionTakenByRole = data.approvingUserRole;
    stage.comments = data.comments;
    stage.signature = data.signature;

    // Check if all stages approved
    const allApproved = requisition.approvalChain.every((s) => s.status === 'APPROVED');
    const previousStatus = requisition.status;

    if (allApproved) {
      requisition.status = 'APPROVED';
      requisition.approvedAt = new Date();
    } else {
      // Move to next stage
      requisition.status = 'IN_REVIEW';
      const nextStage = requisition.approvalChain.find((s) => s.status === 'PENDING');
      if (nextStage) {
        requisition.currentApprovalStage = nextStage.stageNumber;
      }
    }

    requisition.updatedAt = new Date();

    // Add action to history
    if (!requisition.actionHistory) {
      requisition.actionHistory = [];
    }
    requisition.actionHistory.push({
      id: `action-${Date.now()}-${requisition.actionHistory.length + 1}`,
      actionType: 'APPROVE',
      performedBy: data.approvingUserId,
      performedByName: data.approvingUserName,
      performedByRole: data.approvingUserRole,
      performedAt: stage.actionTakenAt,
      stageNumber: stage.stageNumber,
      stageName: stage.stageName,
      previousStatus: previousStatus as any,
      newStatus: requisition.status,
      comments: data.comments || `Approved at stage ${stage.stageNumber}: ${stage.stageName}`,
      signature: data.signature,
    });

    // CRITICAL: Create Purchase Order from fully approved requisition
    if (allApproved) {
      const poResult = await createPurchaseOrderFromRequisition(requisition);
      if (poResult.success && poResult.data) {
        // Initialize relatedPurchaseOrders array if not present
        if (!requisition.relatedPurchaseOrders) {
          requisition.relatedPurchaseOrders = [];
        }
        // Link the created PO back to this requisition
        requisition.relatedPurchaseOrders.push(poResult.data.id);
      }
    }

    return {
      success: true,
      message: allApproved ? 'Requisition approved and Purchase Order created' : 'Approval recorded, moving to next stage',
      data: requisition,
      status: 200,
      statusText: 'OK',
    };
  } catch (error) {
    return {
      success: false,
      message: error instanceof Error ? error.message : 'Failed to approve requisition',
      data: null,
      status: 500,
      statusText: 'ERROR',
    };
  }
}

/**
 * Reject requisition
 */
export async function rejectRequisition(
  data: RejectRequisitionRequest
): Promise<APIResponse<Requisition>> {
  try {
    const requisition = mockRequisitions.find((r) => r.id === data.requisitionId);

    if (!requisition) {
      return {
        success: false,
        message: 'Requisition not found',
        data: null,
        status: 404,
        statusText: 'NOT_FOUND',
      };
    }

    if (!requisition.approvalChain) {
      return {
        success: false,
        message: 'Approval chain not configured',
        data: null,
        status: 400,
        statusText: 'BAD_REQUEST',
      };
    }

    if (!data.remarks || data.remarks.trim().length === 0) {
      return {
        success: false,
        message: 'Rejection remarks are required',
        data: null,
        status: 400,
        statusText: 'BAD_REQUEST',
      };
    }

    if (!data.signature) {
      return {
        success: false,
        message: 'Signature is required for rejection',
        data: null,
        status: 400,
        statusText: 'BAD_REQUEST',
      };
    }

    const stage = requisition.approvalChain.find(
      (s) => s.stageNumber === (requisition.currentApprovalStage || 1)
    );

    if (!stage) {
      return {
        success: false,
        message: 'Approval stage not found',
        data: null,
        status: 404,
        statusText: 'NOT_FOUND',
      };
    }

    // Update stage with rejection
    stage.status = 'REJECTED';
    stage.actionTakenAt = new Date();
    stage.actionTakenBy = data.rejectingUserId;
    stage.actionTakenByRole = data.rejectingUserRole;
    stage.remarks = data.remarks;
    stage.comments = data.comments;
    stage.signature = data.signature;

    // Reset requisition to DRAFT for resubmission
    const previousStatus = requisition.status;
    requisition.status = 'REJECTED';
    requisition.rejectedAt = new Date();
    requisition.currentApprovalStage = 0;
    requisition.updatedAt = new Date();

    // Add action to history
    if (!requisition.actionHistory) {
      requisition.actionHistory = [];
    }
    requisition.actionHistory.push({
      id: `action-${Date.now()}-${requisition.actionHistory.length + 1}`,
      actionType: 'REJECT',
      performedBy: data.rejectingUserId,
      performedByName: data.rejectingUserName,
      performedByRole: data.rejectingUserRole,
      performedAt: stage.actionTakenAt,
      stageNumber: stage.stageNumber,
      stageName: stage.stageName,
      previousStatus: previousStatus as any,
      newStatus: 'REJECTED',
      remarks: data.remarks,
      comments: data.comments,
      signature: data.signature,
    });

    return {
      success: true,
      message: 'Requisition rejected and returned to draft',
      data: requisition,
      status: 200,
      statusText: 'OK',
    };
  } catch (error) {
    return {
      success: false,
      message: error instanceof Error ? error.message : 'Failed to reject requisition',
      data: null,
      status: 500,
      statusText: 'ERROR',
    };
  }
}

/**
 * Get requisition statistics
 */
export async function getRequisitionStats(): Promise<APIResponse<RequisitionStats>> {
  try {
    const stats: RequisitionStats = {
      total: mockRequisitions.length,
      draft: mockRequisitions.filter((r) => r.status === 'DRAFT').length,
      submitted: mockRequisitions.filter((r) => r.status === 'SUBMITTED').length,
      inApproval: mockRequisitions.filter((r) => r.status === 'IN_REVIEW').length,
      approved: mockRequisitions.filter((r) => r.status === 'APPROVED').length,
      rejected: mockRequisitions.filter((r) => r.status === 'REJECTED').length,
      totalValue: mockRequisitions.reduce((sum, r) => sum + r.totalAmount, 0),
      averageApprovalTime: 3, // Mock value
    };

    return {
      success: true,
      message: 'Statistics retrieved successfully',
      data: stats,
      status: 200,
      statusText: 'OK',
    };
  } catch (error) {
    return {
      success: false,
      message: error instanceof Error ? error.message : 'Failed to fetch statistics',
      data: null,
      status: 500,
      statusText: 'ERROR',
    };
  }
}

/**
 * Delete requisition (DRAFT only)
 */
export async function deleteRequisition(requisitionId: string): Promise<APIResponse> {
  try {
    const index = mockRequisitions.findIndex((r) => r.id === requisitionId);

    if (index === -1) {
      return {
        success: false,
        message: 'Requisition not found',
        data: null,
        status: 404,
        statusText: 'NOT_FOUND',
      };
    }

    const requisition = mockRequisitions[index];

    if (requisition.status !== 'DRAFT') {
      return {
        success: false,
        message: 'Only DRAFT requisitions can be deleted',
        data: null,
        status: 400,
        statusText: 'BAD_REQUEST',
      };
    }

    mockRequisitions.splice(index, 1);

    return {
      success: true,
      message: 'Requisition deleted successfully',
      data: null,
      status: 200,
      statusText: 'OK',
    };
  } catch (error) {
    return {
      success: false,
      message: error instanceof Error ? error.message : 'Failed to delete requisition',
      data: null,
      status: 500,
      statusText: 'ERROR',
    };
  }
}
