'use server';

import { cache } from 'react';
import {
  PurchaseOrder,
  PurchaseOrderStatus,
  CreatePurchaseOrderRequest,
  UpdatePurchaseOrderRequest,
  SubmitPurchaseOrderRequest,
  ApprovePurchaseOrderRequest,
  RejectPurchaseOrderRequest,
  PurchaseOrderStats,
  POActionHistoryEntry,
} from '@/types/purchase-order';
import { Requisition } from '@/types/requisition';
import { APIResponse } from '@/types';

/**
 * Mock purchase orders database
 * In production, replace with actual database queries
 */
let mockPurchaseOrders: PurchaseOrder[] = [];

/**
 * Generate next PO number
 * Format: PO-YYYY-001, PO-YYYY-002, etc.
 */
function generatePONumber(): string {
  const year = new Date().getFullYear();
  const count = mockPurchaseOrders.filter((po) =>
    po.poNumber.startsWith(`PO-${year}`)
  ).length;
  return `PO-${year}-${String(count + 1).padStart(3, '0')}`;
}

/**
 * Create a Purchase Order from an approved Requisition
 * This is called automatically when a requisition reaches APPROVED status
 */
export async function createPurchaseOrderFromRequisition(
  requisition: Requisition
): Promise<APIResponse<PurchaseOrder>> {
  try {
    // Validate requisition is approved
    if (requisition.status !== 'APPROVED') {
      return {
        success: false,
        message: 'Only APPROVED requisitions can be converted to purchase orders',
        data: null,
        status: 400,
        statusText: 'BAD_REQUEST',
      };
    }

    const poId = `po-${Date.now()}`;
    const poNumber = generatePONumber();
    const now = new Date();

    // Map requisition items to PO items
    const poItems = requisition.items.map((item, index) => ({
      id: `po-item-${poId}-${index + 1}`,
      poId,
      itemNumber: item.itemNumber,
      description: item.description,
      category: item.category,
      quantity: item.quantity,
      unitPrice: item.unitPrice,
      unit: item.unit,
      totalPrice: item.totalPrice,
      notes: item.notes || '',
      createdAt: now,
      updatedAt: now,
    }));

    // Create PO with 4-stage approval chain
    const purchaseOrder: PurchaseOrder = {
      id: poId,
      poNumber,
      title: requisition.title,
      description: requisition.description,
      vendorId: requisition.vendorId,
      vendorName: requisition.vendorName || 'TBD',
      department: requisition.department,
      departmentId: requisition.departmentId,
      requestedBy: requisition.requestedBy,
      requestedByName: requisition.requestedByName,
      requestedByRole: requisition.requestedByRole,
      requestedDate: now,
      requiredByDate: requisition.requiredByDate,
      priority: requisition.priority,
      status: 'DRAFT',
      items: poItems,
      totalAmount: requisition.totalAmount,
      currency: requisition.currency,
      currentApprovalStage: 0,
      totalApprovalStages: 4,
      approvalChain: [
        {
          stageNumber: 1,
          stageName: 'Procurement Manager',
          assignedTo: 'user-007',
          assignedRole: 'PROCUREMENT_MANAGER',
          status: 'PENDING',
        },
        {
          stageNumber: 2,
          stageName: 'Finance Manager',
          assignedTo: 'user-008',
          assignedRole: 'FINANCE_MANAGER',
          status: 'PENDING',
        },
        {
          stageNumber: 3,
          stageName: 'Vendor Compliance',
          assignedTo: 'user-009',
          assignedRole: 'VENDOR_COMPLIANCE',
          status: 'PENDING',
        },
        {
          stageNumber: 4,
          stageName: 'Director',
          assignedTo: 'user-005',
          assignedRole: 'DIRECTOR',
          status: 'PENDING',
        },
      ],
      sourceRequisitionId: requisition.id,
      sourceRequisitionNumber: requisition.requisitionNumber,
      createdFromRequisition: true,
      budgetCode: requisition.budgetCode,
      costCenter: requisition.costCenter,
      projectCode: requisition.projectCode,
      createdAt: now,
      updatedAt: now,
      // Initialize action history with creation entry
      actionHistory: [
        {
          id: `action-${Date.now()}-1`,
          actionType: 'CREATE',
          performedBy: requisition.requestedBy,
          performedByName: requisition.requestedByName,
          performedByRole: requisition.requestedByRole,
          performedAt: now,
          newStatus: 'DRAFT',
          comments: `Purchase Order created from approved requisition ${requisition.requisitionNumber} with ${poItems.length} item(s)`,
        },
      ],
    };

    mockPurchaseOrders.push(purchaseOrder);

    return {
      success: true,
      message: 'Purchase Order created from requisition successfully',
      data: purchaseOrder,
      status: 201,
      statusText: 'CREATED',
    };
  } catch (error) {
    return {
      success: false,
      message: error instanceof Error ? error.message : 'Failed to create purchase order from requisition',
      data: null,
      status: 500,
      statusText: 'ERROR',
    };
  }
}

/**
 * Create a new purchase order manually
 */
export async function createPurchaseOrder(
  data: CreatePurchaseOrderRequest
): Promise<APIResponse<PurchaseOrder>> {
  try {
    const poId = `po-${Date.now()}`;
    const poNumber = generatePONumber();
    const now = new Date();

    // Calculate total amount
    const totalAmount = data.items.reduce((sum, item) => sum + item.totalPrice, 0);

    // Map items
    const poItems = data.items.map((item, index) => ({
      ...item,
      id: `po-item-${poId}-${index + 1}`,
      poId,
      createdAt: now,
      updatedAt: now,
    }));

    // Create PO with 4-stage approval chain
    const purchaseOrder: PurchaseOrder = {
      id: poId,
      poNumber,
      title: data.title,
      description: data.description,
      vendorId: data.vendorId,
      vendorName: data.vendorName,
      department: data.department,
      departmentId: data.departmentId,
      requestedBy: data.createdBy,
      requestedByName: data.createdByName,
      requestedByRole: data.createdByRole,
      requestedDate: now,
      requiredByDate: new Date(data.requiredByDate),
      priority: data.priority,
      status: 'DRAFT',
      items: poItems,
      totalAmount,
      currency: 'ZMW',
      currentApprovalStage: 0,
      totalApprovalStages: 4,
      approvalChain: [
        {
          stageNumber: 1,
          stageName: 'Procurement Manager',
          assignedTo: 'user-007',
          assignedRole: 'PROCUREMENT_MANAGER',
          status: 'PENDING',
        },
        {
          stageNumber: 2,
          stageName: 'Finance Manager',
          assignedTo: 'user-008',
          assignedRole: 'FINANCE_MANAGER',
          status: 'PENDING',
        },
        {
          stageNumber: 3,
          stageName: 'Vendor Compliance',
          assignedTo: 'user-009',
          assignedRole: 'VENDOR_COMPLIANCE',
          status: 'PENDING',
        },
        {
          stageNumber: 4,
          stageName: 'Director',
          assignedTo: 'user-005',
          assignedRole: 'DIRECTOR',
          status: 'PENDING',
        },
      ],
      sourceRequisitionId: data.sourceRequisitionId,
      sourceRequisitionNumber: data.sourceRequisitionNumber,
      createdFromRequisition: true,
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
          comments: `Purchase Order created with ${data.items.length} item(s)`,
        },
      ],
    };

    mockPurchaseOrders.push(purchaseOrder);

    return {
      success: true,
      message: 'Purchase Order created successfully',
      data: purchaseOrder,
      status: 201,
      statusText: 'CREATED',
    };
  } catch (error) {
    return {
      success: false,
      message: error instanceof Error ? error.message : 'Failed to create purchase order',
      data: null,
      status: 500,
      statusText: 'ERROR',
    };
  }
}

/**
 * Get all purchase orders (cached)
 */
export const getPurchaseOrders = cache(async (): Promise<APIResponse<PurchaseOrder[]>> => {
  try {
    return {
      success: true,
      message: 'Purchase orders retrieved successfully',
      data: mockPurchaseOrders,
      status: 200,
      statusText: 'OK',
    };
  } catch (error) {
    return {
      success: false,
      message: error instanceof Error ? error.message : 'Failed to fetch purchase orders',
      data: null,
      status: 500,
      statusText: 'ERROR',
    };
  }
});

/**
 * Get purchase order by ID
 */
export async function getPurchaseOrderById(poId: string): Promise<APIResponse<PurchaseOrder>> {
  try {
    const purchaseOrder = mockPurchaseOrders.find((po) => po.id === poId);

    if (!purchaseOrder) {
      return {
        success: false,
        message: 'Purchase order not found',
        data: null,
        status: 404,
        statusText: 'NOT_FOUND',
      };
    }

    return {
      success: true,
      message: 'Purchase order retrieved successfully',
      data: purchaseOrder,
      status: 200,
      statusText: 'OK',
    };
  } catch (error) {
    return {
      success: false,
      message: error instanceof Error ? error.message : 'Failed to fetch purchase order',
      data: null,
      status: 500,
      statusText: 'ERROR',
    };
  }
}

/**
 * Update purchase order (DRAFT only)
 */
export async function updatePurchaseOrder(
  data: UpdatePurchaseOrderRequest
): Promise<APIResponse<PurchaseOrder>> {
  try {
    const purchaseOrder = mockPurchaseOrders.find((po) => po.id === data.poId);

    if (!purchaseOrder) {
      return {
        success: false,
        message: 'Purchase order not found',
        data: null,
        status: 404,
        statusText: 'NOT_FOUND',
      };
    }

    if (purchaseOrder.status !== 'DRAFT') {
      return {
        success: false,
        message: 'Cannot update purchase order that is not in DRAFT status',
        data: null,
        status: 400,
        statusText: 'BAD_REQUEST',
      };
    }

    if (data.title) purchaseOrder.title = data.title;
    if (data.description) purchaseOrder.description = data.description;
    if (data.vendorId) purchaseOrder.vendorId = data.vendorId;
    if (data.vendorName) purchaseOrder.vendorName = data.vendorName;
    if (data.requiredByDate) purchaseOrder.requiredByDate = new Date(data.requiredByDate);
    if (data.priority) purchaseOrder.priority = data.priority;
    if (data.items) {
      purchaseOrder.items = data.items.map((item) => ({
        ...item,
        poId: purchaseOrder.id,
        createdAt: new Date(),
        updatedAt: new Date(),
      }));
      purchaseOrder.totalAmount = purchaseOrder.items.reduce((sum, item) => sum + item.totalPrice, 0);
    }
    if (data.budgetCode) purchaseOrder.budgetCode = data.budgetCode;
    if (data.costCenter) purchaseOrder.costCenter = data.costCenter;
    if (data.projectCode) purchaseOrder.projectCode = data.projectCode;

    purchaseOrder.updatedAt = new Date();

    return {
      success: true,
      message: 'Purchase order updated successfully',
      data: purchaseOrder,
      status: 200,
      statusText: 'OK',
    };
  } catch (error) {
    return {
      success: false,
      message: error instanceof Error ? error.message : 'Failed to update purchase order',
      data: null,
      status: 500,
      statusText: 'ERROR',
    };
  }
}

/**
 * Submit purchase order for approval
 */
export async function submitPurchaseOrderForApproval(
  data: SubmitPurchaseOrderRequest
): Promise<APIResponse<PurchaseOrder>> {
  try {
    const purchaseOrder = mockPurchaseOrders.find((po) => po.id === data.poId);

    if (!purchaseOrder) {
      return {
        success: false,
        message: 'Purchase order not found',
        data: null,
        status: 404,
        statusText: 'NOT_FOUND',
      };
    }

    if (purchaseOrder.status !== 'DRAFT') {
      return {
        success: false,
        message: 'Only DRAFT purchase orders can be submitted',
        data: null,
        status: 400,
        statusText: 'BAD_REQUEST',
      };
    }

    if (!purchaseOrder.items || purchaseOrder.items.length === 0) {
      return {
        success: false,
        message: 'Purchase order must have at least one item',
        data: null,
        status: 400,
        statusText: 'BAD_REQUEST',
      };
    }

    purchaseOrder.status = 'SUBMITTED';
    purchaseOrder.currentApprovalStage = 1;
    purchaseOrder.submittedAt = new Date();
    purchaseOrder.updatedAt = new Date();

    // Add action to history
    if (!purchaseOrder.actionHistory) {
      purchaseOrder.actionHistory = [];
    }
    purchaseOrder.actionHistory.push({
      id: `action-${Date.now()}-${purchaseOrder.actionHistory.length + 1}`,
      actionType: 'SUBMIT',
      performedBy: data.submittedBy,
      performedByName: data.submittedByName,
      performedByRole: data.submittedByRole,
      performedAt: purchaseOrder.submittedAt,
      previousStatus: 'DRAFT',
      newStatus: 'SUBMITTED',
      comments: data.comments || 'Purchase order submitted for approval',
    });

    return {
      success: true,
      message: 'Purchase order submitted for approval',
      data: purchaseOrder,
      status: 200,
      statusText: 'OK',
    };
  } catch (error) {
    return {
      success: false,
      message: error instanceof Error ? error.message : 'Failed to submit purchase order',
      data: null,
      status: 500,
      statusText: 'ERROR',
    };
  }
}

/**
 * Approve purchase order
 */
export async function approvePurchaseOrder(
  data: ApprovePurchaseOrderRequest
): Promise<APIResponse<PurchaseOrder>> {
  try {
    const purchaseOrder = mockPurchaseOrders.find((po) => po.id === data.poId);

    if (!purchaseOrder) {
      return {
        success: false,
        message: 'Purchase order not found',
        data: null,
        status: 404,
        statusText: 'NOT_FOUND',
      };
    }

    if (!purchaseOrder.approvalChain) {
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

    const stage = purchaseOrder.approvalChain.find(
      (s) => s.stageNumber === (data.stageNumber || purchaseOrder.currentApprovalStage)
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
    const allApproved = purchaseOrder.approvalChain.every((s) => s.status === 'APPROVED');
    const previousStatus = purchaseOrder.status;

    if (allApproved) {
      purchaseOrder.status = 'APPROVED';
      purchaseOrder.approvedAt = new Date();
    } else {
      // Move to next stage
      purchaseOrder.status = 'IN_REVIEW';
      const nextStage = purchaseOrder.approvalChain.find((s) => s.status === 'PENDING');
      if (nextStage) {
        purchaseOrder.currentApprovalStage = nextStage.stageNumber;
      }
    }

    purchaseOrder.updatedAt = new Date();

    // Add action to history
    if (!purchaseOrder.actionHistory) {
      purchaseOrder.actionHistory = [];
    }
    purchaseOrder.actionHistory.push({
      id: `action-${Date.now()}-${purchaseOrder.actionHistory.length + 1}`,
      actionType: 'APPROVE',
      performedBy: data.approvingUserId,
      performedByName: data.approvingUserName,
      performedByRole: data.approvingUserRole,
      performedAt: stage.actionTakenAt,
      stageNumber: stage.stageNumber,
      stageName: stage.stageName,
      previousStatus: previousStatus as any,
      newStatus: purchaseOrder.status,
      comments: data.comments || `Approved at stage ${stage.stageNumber}: ${stage.stageName}`,
      signature: data.signature,
    });

    return {
      success: true,
      message: allApproved ? 'Purchase order approved' : 'Approval recorded, moving to next stage',
      data: purchaseOrder,
      status: 200,
      statusText: 'OK',
    };
  } catch (error) {
    return {
      success: false,
      message: error instanceof Error ? error.message : 'Failed to approve purchase order',
      data: null,
      status: 500,
      statusText: 'ERROR',
    };
  }
}

/**
 * Reject purchase order
 */
export async function rejectPurchaseOrder(
  data: RejectPurchaseOrderRequest
): Promise<APIResponse<PurchaseOrder>> {
  try {
    const purchaseOrder = mockPurchaseOrders.find((po) => po.id === data.poId);

    if (!purchaseOrder) {
      return {
        success: false,
        message: 'Purchase order not found',
        data: null,
        status: 404,
        statusText: 'NOT_FOUND',
      };
    }

    if (!purchaseOrder.approvalChain) {
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

    const stage = purchaseOrder.approvalChain.find(
      (s) => s.stageNumber === (purchaseOrder.currentApprovalStage || 1)
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

    // Set PO back to DRAFT for resubmission
    const previousStatus = purchaseOrder.status;
    purchaseOrder.status = 'REJECTED';
    purchaseOrder.rejectedAt = new Date();
    purchaseOrder.currentApprovalStage = 0;
    purchaseOrder.updatedAt = new Date();

    // Add action to history
    if (!purchaseOrder.actionHistory) {
      purchaseOrder.actionHistory = [];
    }
    purchaseOrder.actionHistory.push({
      id: `action-${Date.now()}-${purchaseOrder.actionHistory.length + 1}`,
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
      message: 'Purchase order rejected and returned to draft',
      data: purchaseOrder,
      status: 200,
      statusText: 'OK',
    };
  } catch (error) {
    return {
      success: false,
      message: error instanceof Error ? error.message : 'Failed to reject purchase order',
      data: null,
      status: 500,
      statusText: 'ERROR',
    };
  }
}

/**
 * Get purchase order statistics
 */
export async function getPurchaseOrderStats(): Promise<APIResponse<PurchaseOrderStats>> {
  try {
    const stats: PurchaseOrderStats = {
      total: mockPurchaseOrders.length,
      draft: mockPurchaseOrders.filter((po) => po.status === 'DRAFT').length,
      submitted: mockPurchaseOrders.filter((po) => po.status === 'SUBMITTED').length,
      inApproval: mockPurchaseOrders.filter((po) => po.status === 'IN_REVIEW').length,
      approved: mockPurchaseOrders.filter((po) => po.status === 'APPROVED').length,
      rejected: mockPurchaseOrders.filter((po) => po.status === 'REJECTED').length,
      totalValue: mockPurchaseOrders.reduce((sum, po) => sum + po.totalAmount, 0),
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
 * Delete purchase order (DRAFT only)
 */
export async function deletePurchaseOrder(poId: string): Promise<APIResponse> {
  try {
    const index = mockPurchaseOrders.findIndex((po) => po.id === poId);

    if (index === -1) {
      return {
        success: false,
        message: 'Purchase order not found',
        data: null,
        status: 404,
        statusText: 'NOT_FOUND',
      };
    }

    const purchaseOrder = mockPurchaseOrders[index];

    if (purchaseOrder.status !== 'DRAFT') {
      return {
        success: false,
        message: 'Only DRAFT purchase orders can be deleted',
        data: null,
        status: 400,
        statusText: 'BAD_REQUEST',
      };
    }

    mockPurchaseOrders.splice(index, 1);

    return {
      success: true,
      message: 'Purchase order deleted successfully',
      data: null,
      status: 200,
      statusText: 'OK',
    };
  } catch (error) {
    return {
      success: false,
      message: error instanceof Error ? error.message : 'Failed to delete purchase order',
      data: null,
      status: 500,
      statusText: 'ERROR',
    };
  }
}
