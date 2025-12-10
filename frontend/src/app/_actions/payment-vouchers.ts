'use server'

import { PaymentVoucher, CreatePaymentVoucherRequest, UpdatePaymentVoucherRequest, SubmitPaymentVoucherRequest, ApprovePaymentVoucherRequest, RejectPaymentVoucherRequest, MarkPaymentVoucherPaidRequest, PaymentVoucherStats, PVApprovalRecord, PVActionHistoryEntry } from '@/types/payment-voucher'
import { PurchaseOrder } from '@/types/purchase-order'

/**
 * Mock storage for Payment Vouchers
 * In production, this would be a database
 */
let mockPaymentVouchers: PaymentVoucher[] = []
let pvCounter = 1000

/**
 * Generate unique PV Number
 */
function generatePVNumber(): string {
  const year = new Date().getFullYear()
  pvCounter++
  return `PV-${year}-${pvCounter.toString().padStart(4, '0')}`
}

/**
 * Initialize 3-stage approval chain for PV
 */
function initializePVApprovalChain(): PVApprovalRecord[] {
  return [
    {
      stageNumber: 1,
      stageName: 'Finance Manager Review',
      assignedTo: 'finance-manager-1',
      assignedRole: 'FINANCE_MANAGER',
      status: 'PENDING',
    },
    {
      stageNumber: 2,
      stageName: 'Approval Authority Review',
      assignedTo: 'approval-authority-1',
      assignedRole: 'APPROVAL_AUTHORITY',
      status: 'PENDING',
    },
    {
      stageNumber: 3,
      stageName: 'Director Approval',
      assignedTo: 'director-1',
      assignedRole: 'DIRECTOR',
      status: 'PENDING',
    },
  ]
}

/**
 * Create Payment Voucher from approved Purchase Order
 * Automatically triggered when PO is approved
 */
export async function createPaymentVoucherFromPurchaseOrder(
  po: PurchaseOrder
): Promise<{ success: boolean; data?: PaymentVoucher; message: string }> {
  try {
    if (po.status !== 'APPROVED') {
      return { success: false, message: 'Purchase order must be approved to create payment voucher' }
    }

    const pvId = `pv-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`
    const now = new Date()

    const paymentVoucher: PaymentVoucher = {
      id: pvId,
      pvNumber: generatePVNumber(),
      title: `Payment for ${po.poNumber}`,
      description: po.description,
      vendorId: po.vendorId,
      vendorName: po.vendorName,
      department: po.department,
      departmentId: po.departmentId,
      requestedBy: po.requestedBy,
      requestedByName: po.requestedByName,
      requestedByRole: po.requestedByRole,
      requestedDate: now,
      paymentDueDate: new Date(now.getTime() + 30 * 24 * 60 * 60 * 1000), // 30 days from now
      priority: 'MEDIUM',
      paymentMethod: 'BANK_TRANSFER',
      status: 'DRAFT',

      // Line items - mapped from PO items
      items: po.items.map((poItem, index) => ({
        id: `pvi-${pvId}-${index}`,
        pvId: pvId,
        poItemId: poItem.id,
        itemNumber: poItem.itemNumber,
        description: poItem.description,
        category: poItem.category,
        quantity: poItem.quantity,
        unitPrice: poItem.unitPrice,
        unit: poItem.unit,
        totalPrice: poItem.totalPrice,
        createdAt: now,
        updatedAt: now,
      })),

      totalAmount: po.totalAmount,
      currency: po.currency,

      // Approval tracking - initialize 3-stage chain
      approvalChain: initializePVApprovalChain(),
      currentApprovalStage: 1,
      totalApprovalStages: 3,

      // Action history - record creation
      actionHistory: [
        {
          id: `pva-${pvId}-create`,
          actionType: 'CREATE',
          performedBy: po.requestedBy,
          performedByName: po.requestedByName,
          performedByRole: po.requestedByRole,
          performedAt: now,
          newStatus: 'DRAFT',
          metadata: {
            sourcePurchaseOrderId: po.id,
            sourcePurchaseOrderNumber: po.poNumber,
            autoCreated: true,
          },
        } as PVActionHistoryEntry,
      ],

      // PO linking
      sourcePurchaseOrderId: po.id,
      sourcePurchaseOrderNumber: po.poNumber,
      createdFromPurchaseOrder: true,

      // Requisition traceability
      sourceRequisitionId: po.sourceRequisitionId,
      sourceRequisitionNumber: po.sourceRequisitionNumber,

      // Financial metadata
      budgetCode: po.budgetCode,
      costCenter: po.costCenter,
      projectCode: po.projectCode,

      // Timestamps
      createdAt: now,
      updatedAt: now,
    }

    mockPaymentVouchers.push(paymentVoucher)

    return {
      success: true,
      data: paymentVoucher,
      message: 'Payment voucher created from purchase order',
    }
  } catch (error) {
    return {
      success: false,
      message: `Failed to create payment voucher: ${error instanceof Error ? error.message : 'Unknown error'}`,
    }
  }
}

/**
 * Create Payment Voucher manually
 */
export async function createPaymentVoucher(
  data: CreatePaymentVoucherRequest
): Promise<{ success: boolean; data?: PaymentVoucher; message: string }> {
  try {
    const pvId = `pv-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`
    const now = new Date()

    const paymentVoucher: PaymentVoucher = {
      id: pvId,
      pvNumber: generatePVNumber(),
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
      paymentDueDate: new Date(data.paymentDueDate),
      priority: data.priority,
      paymentMethod: data.paymentMethod,
      bankDetails: data.bankDetails,
      status: 'DRAFT',

      // Line items
      items: data.items.map((item, index) => ({
        id: `pvi-${pvId}-${index}`,
        pvId: pvId,
        poItemId: item.poItemId,
        itemNumber: item.itemNumber,
        description: item.description,
        category: item.category,
        quantity: item.quantity,
        unitPrice: item.unitPrice,
        unit: item.unit,
        totalPrice: item.totalPrice,
        notes: item.notes,
        createdAt: now,
        updatedAt: now,
      })),

      totalAmount: data.items.reduce((sum, item) => sum + item.totalPrice, 0),
      currency: 'ZMW',

      // Approval tracking
      approvalChain: initializePVApprovalChain(),
      currentApprovalStage: 1,
      totalApprovalStages: 3,

      // Action history
      actionHistory: [
        {
          id: `pva-${pvId}-create`,
          actionType: 'CREATE',
          performedBy: data.createdBy,
          performedByName: data.createdByName,
          performedByRole: data.createdByRole,
          performedAt: now,
          newStatus: 'DRAFT',
        } as PVActionHistoryEntry,
      ],

      // PO linking
      sourcePurchaseOrderId: data.sourcePurchaseOrderId,
      sourcePurchaseOrderNumber: data.sourcePurchaseOrderNumber,
      createdFromPurchaseOrder: false,

      // Requisition traceability
      sourceRequisitionId: data.sourceRequisitionId,
      sourceRequisitionNumber: data.sourceRequisitionNumber,

      // Financial metadata
      budgetCode: data.budgetCode,
      costCenter: data.costCenter,
      projectCode: data.projectCode,
      taxAmount: data.taxAmount,
      withholdingTaxAmount: data.withholdingTaxAmount,

      // Timestamps
      createdAt: now,
      updatedAt: now,
    }

    mockPaymentVouchers.push(paymentVoucher)

    return {
      success: true,
      data: paymentVoucher,
      message: 'Payment voucher created successfully',
    }
  } catch (error) {
    return {
      success: false,
      message: `Failed to create payment voucher: ${error instanceof Error ? error.message : 'Unknown error'}`,
    }
  }
}

/**
 * Get all payment vouchers
 */
export async function getPaymentVouchers(): Promise<{
  success: boolean
  data?: PaymentVoucher[]
  message: string
  status: number
}> {
  try {
    // Simulated cache behavior
    return {
      success: true,
      data: mockPaymentVouchers,
      message: 'Payment vouchers retrieved successfully',
      status: 200,
    }
  } catch (error) {
    return {
      success: false,
      message: `Failed to retrieve payment vouchers: ${error instanceof Error ? error.message : 'Unknown error'}`,
      status: 500,
    }
  }
}

/**
 * Get payment voucher by ID
 */
export async function getPaymentVoucherById(pvId: string): Promise<{
  success: boolean
  data?: PaymentVoucher
  message: string
  status: number
}> {
  try {
    const pv = mockPaymentVouchers.find((p) => p.id === pvId)

    if (!pv) {
      return {
        success: false,
        message: 'Payment voucher not found',
        status: 404,
      }
    }

    return {
      success: true,
      data: pv,
      message: 'Payment voucher retrieved successfully',
      status: 200,
    }
  } catch (error) {
    return {
      success: false,
      message: `Failed to retrieve payment voucher: ${error instanceof Error ? error.message : 'Unknown error'}`,
      status: 500,
    }
  }
}

/**
 * Update Payment Voucher (DRAFT only)
 */
export async function updatePaymentVoucher(
  data: UpdatePaymentVoucherRequest
): Promise<{ success: boolean; data?: PaymentVoucher; message: string }> {
  try {
    const pv = mockPaymentVouchers.find((p) => p.id === data.pvId)

    if (!pv) {
      return {
        success: false,
        message: 'Payment voucher not found',
      }
    }

    if (pv.status !== 'DRAFT') {
      return {
        success: false,
        message: 'Only DRAFT payment vouchers can be updated',
      }
    }

    const now = new Date()
    const changes: Record<string, { oldValue: any; newValue: any }> = {}

    // Track changes
    if (data.title && data.title !== pv.title) {
      changes['title'] = { oldValue: pv.title, newValue: data.title }
      pv.title = data.title
    }
    if (data.description && data.description !== pv.description) {
      changes['description'] = { oldValue: pv.description, newValue: data.description }
      pv.description = data.description
    }
    if (data.vendorName && data.vendorName !== pv.vendorName) {
      changes['vendorName'] = { oldValue: pv.vendorName, newValue: data.vendorName }
      pv.vendorName = data.vendorName
    }
    if (data.paymentDueDate) {
      const newDate = new Date(data.paymentDueDate)
      if (newDate.getTime() !== pv.paymentDueDate.getTime()) {
        changes['paymentDueDate'] = { oldValue: pv.paymentDueDate, newValue: newDate }
        pv.paymentDueDate = newDate
      }
    }
    if (data.priority && data.priority !== pv.priority) {
      changes['priority'] = { oldValue: pv.priority, newValue: data.priority }
      pv.priority = data.priority
    }
    if (data.paymentMethod && data.paymentMethod !== pv.paymentMethod) {
      changes['paymentMethod'] = { oldValue: pv.paymentMethod, newValue: data.paymentMethod }
      pv.paymentMethod = data.paymentMethod
    }
    if (data.bankDetails) {
      changes['bankDetails'] = { oldValue: pv.bankDetails, newValue: data.bankDetails }
      pv.bankDetails = data.bankDetails
    }
    if (data.items) {
      changes['items'] = { oldValue: pv.items.length, newValue: data.items.length }
      pv.items = data.items.map((item, index) => {
        // Get existing item or create new one
        const existingItem = pv.items.find(pi => pi.id === item.id);
        return {
          id: item.id || `pvi-${pv.id}-${index}`,
          pvId: pv.id,
          poItemId: item.poItemId,
          itemNumber: item.itemNumber,
          description: item.description,
          category: item.category,
          quantity: item.quantity,
          unitPrice: item.unitPrice,
          unit: item.unit,
          totalPrice: item.totalPrice,
          notes: item.notes,
          createdAt: existingItem?.createdAt || now,
          updatedAt: now,
        };
      })
      pv.totalAmount = pv.items.reduce((sum, item) => sum + item.totalPrice, 0)
    }

    pv.updatedAt = now

    // Add to action history
    if (!pv.actionHistory) pv.actionHistory = []
    pv.actionHistory.push({
      id: `pva-${pv.id}-${Date.now()}`,
      actionType: 'UPDATE',
      performedBy: data.updatedBy,
      performedByName: 'System',
      performedByRole: 'SYSTEM',
      performedAt: now,
      changedFields: Object.keys(changes).length > 0 ? changes : undefined,
    } as PVActionHistoryEntry)

    return {
      success: true,
      data: pv,
      message: 'Payment voucher updated successfully',
    }
  } catch (error) {
    return {
      success: false,
      message: `Failed to update payment voucher: ${error instanceof Error ? error.message : 'Unknown error'}`,
    }
  }
}

/**
 * Submit Payment Voucher for Approval
 */
export async function submitPaymentVoucherForApproval(
  data: SubmitPaymentVoucherRequest
): Promise<{ success: boolean; data?: PaymentVoucher; message: string }> {
  try {
    const pv = mockPaymentVouchers.find((p) => p.id === data.pvId)

    if (!pv) {
      return {
        success: false,
        message: 'Payment voucher not found',
      }
    }

    if (pv.status !== 'DRAFT') {
      return {
        success: false,
        message: 'Only DRAFT payment vouchers can be submitted',
      }
    }

    const now = new Date()

    pv.status = 'SUBMITTED'
    pv.submittedAt = now
    pv.updatedAt = now
    pv.currentApprovalStage = 1

    // Add to action history
    if (!pv.actionHistory) pv.actionHistory = []
    pv.actionHistory.push({
      id: `pva-${pv.id}-${Date.now()}`,
      actionType: 'SUBMIT',
      performedBy: data.submittedBy,
      performedByName: data.submittedByName,
      performedByRole: data.submittedByRole,
      performedAt: now,
      previousStatus: 'DRAFT',
      newStatus: 'SUBMITTED',
      comments: data.comments,
    } as PVActionHistoryEntry)

    return {
      success: true,
      data: pv,
      message: 'Payment voucher submitted for approval',
    }
  } catch (error) {
    return {
      success: false,
      message: `Failed to submit payment voucher: ${error instanceof Error ? error.message : 'Unknown error'}`,
    }
  }
}

/**
 * Approve Payment Voucher
 */
export async function approvePaymentVoucher(
  data: ApprovePaymentVoucherRequest
): Promise<{ success: boolean; data?: PaymentVoucher; message: string }> {
  try {
    const pv = mockPaymentVouchers.find((p) => p.id === data.pvId)

    if (!pv) {
      return {
        success: false,
        message: 'Payment voucher not found',
      }
    }

    if (pv.status !== 'SUBMITTED' && pv.status !== 'IN_REVIEW') {
      return {
        success: false,
        message: 'Payment voucher must be submitted or in review to approve',
      }
    }

    const now = new Date()
    const currentStage = pv.currentApprovalStage || 1

    // Update approval chain
    if (pv.approvalChain && pv.approvalChain[currentStage - 1]) {
      pv.approvalChain[currentStage - 1].status = 'APPROVED'
      pv.approvalChain[currentStage - 1].actionTakenAt = now
      pv.approvalChain[currentStage - 1].actionTakenBy = data.approvingUserId
      pv.approvalChain[currentStage - 1].actionTakenByRole = data.approvingUserRole
      pv.approvalChain[currentStage - 1].comments = data.comments
      pv.approvalChain[currentStage - 1].signature = data.signature
    }

    // Determine next status
    const allApproved = pv.approvalChain?.every((stage) => stage.status === 'APPROVED')

    if (allApproved) {
      pv.status = 'APPROVED'
      pv.approvedAt = now
      pv.currentApprovalStage = pv.totalApprovalStages || 3
    } else {
      pv.status = 'IN_REVIEW'
      pv.currentApprovalStage = currentStage + 1
    }

    pv.updatedAt = now

    // Add to action history
    if (!pv.actionHistory) pv.actionHistory = []
    pv.actionHistory.push({
      id: `pva-${pv.id}-${Date.now()}`,
      actionType: 'APPROVE',
      performedBy: data.approvingUserId,
      performedByName: data.approvingUserName,
      performedByRole: data.approvingUserRole,
      performedAt: now,
      stageNumber: currentStage,
      stageName: pv.approvalChain?.[currentStage - 1]?.stageName,
      comments: data.comments,
      signature: data.signature,
      previousStatus: pv.status === 'APPROVED' ? 'IN_REVIEW' : 'SUBMITTED',
      newStatus: pv.status,
    } as PVActionHistoryEntry)

    return {
      success: true,
      data: pv,
      message: allApproved
        ? 'Payment voucher fully approved'
        : `Stage ${currentStage} approved, moving to next stage`,
    }
  } catch (error) {
    return {
      success: false,
      message: `Failed to approve payment voucher: ${error instanceof Error ? error.message : 'Unknown error'}`,
    }
  }
}

/**
 * Reject Payment Voucher
 */
export async function rejectPaymentVoucher(
  data: RejectPaymentVoucherRequest
): Promise<{ success: boolean; data?: PaymentVoucher; message: string }> {
  try {
    const pv = mockPaymentVouchers.find((p) => p.id === data.pvId)

    if (!pv) {
      return {
        success: false,
        message: 'Payment voucher not found',
      }
    }

    if (pv.status !== 'SUBMITTED' && pv.status !== 'IN_REVIEW') {
      return {
        success: false,
        message: 'Payment voucher must be submitted or in review to reject',
      }
    }

    const now = new Date()
    const currentStage = pv.currentApprovalStage || 1

    // Update approval chain
    if (pv.approvalChain && pv.approvalChain[currentStage - 1]) {
      pv.approvalChain[currentStage - 1].status = 'REJECTED'
      pv.approvalChain[currentStage - 1].actionTakenAt = now
      pv.approvalChain[currentStage - 1].actionTakenBy = data.rejectingUserId
      pv.approvalChain[currentStage - 1].actionTakenByRole = data.rejectingUserRole
      pv.approvalChain[currentStage - 1].remarks = data.remarks
      pv.approvalChain[currentStage - 1].signature = data.signature
    }

    pv.status = 'REJECTED'
    pv.rejectedAt = now
    pv.updatedAt = now

    // Add to action history
    if (!pv.actionHistory) pv.actionHistory = []
    pv.actionHistory.push({
      id: `pva-${pv.id}-${Date.now()}`,
      actionType: 'REJECT',
      performedBy: data.rejectingUserId,
      performedByName: data.rejectingUserName,
      performedByRole: data.rejectingUserRole,
      performedAt: now,
      stageNumber: currentStage,
      stageName: pv.approvalChain?.[currentStage - 1]?.stageName,
      remarks: data.remarks,
      signature: data.signature,
      comments: data.comments,
      previousStatus: pv.status,
      newStatus: 'REJECTED',
    } as PVActionHistoryEntry)

    return {
      success: true,
      data: pv,
      message: 'Payment voucher rejected and returned to draft',
    }
  } catch (error) {
    return {
      success: false,
      message: `Failed to reject payment voucher: ${error instanceof Error ? error.message : 'Unknown error'}`,
    }
  }
}

/**
 * Mark Payment Voucher as Paid
 */
export async function markPaymentVoucherAsPaid(
  data: MarkPaymentVoucherPaidRequest
): Promise<{ success: boolean; data?: PaymentVoucher; message: string }> {
  try {
    const pv = mockPaymentVouchers.find((p) => p.id === data.pvId)

    if (!pv) {
      return {
        success: false,
        message: 'Payment voucher not found',
      }
    }

    if (pv.status !== 'APPROVED') {
      return {
        success: false,
        message: 'Only approved payment vouchers can be marked as paid',
      }
    }

    const now = new Date()

    pv.status = 'PAID'
    pv.paidAmount = data.paidAmount
    pv.paidDate = new Date(data.paidDate)
    pv.referenceNumber = data.referenceNumber
    pv.paidAt = now
    pv.updatedAt = now

    // Add to action history
    if (!pv.actionHistory) pv.actionHistory = []
    pv.actionHistory.push({
      id: `pva-${pv.id}-${Date.now()}`,
      actionType: 'MARK_PAID',
      performedBy: data.markedBy,
      performedByName: data.markedByName,
      performedByRole: data.markedByRole,
      performedAt: now,
      comments: data.comments,
      previousStatus: 'APPROVED',
      newStatus: 'PAID',
      metadata: {
        paidAmount: data.paidAmount,
        referenceNumber: data.referenceNumber,
      },
    } as PVActionHistoryEntry)

    return {
      success: true,
      data: pv,
      message: 'Payment voucher marked as paid',
    }
  } catch (error) {
    return {
      success: false,
      message: `Failed to mark payment voucher as paid: ${error instanceof Error ? error.message : 'Unknown error'}`,
    }
  }
}

/**
 * Delete Payment Voucher (DRAFT only)
 */
export async function deletePaymentVoucher(pvId: string): Promise<{
  success: boolean
  message: string
}> {
  try {
    const pv = mockPaymentVouchers.find((p) => p.id === pvId)

    if (!pv) {
      return {
        success: false,
        message: 'Payment voucher not found',
      }
    }

    if (pv.status !== 'DRAFT') {
      return {
        success: false,
        message: 'Only DRAFT payment vouchers can be deleted',
      }
    }

    mockPaymentVouchers = mockPaymentVouchers.filter((p) => p.id !== pvId)

    return {
      success: true,
      message: 'Payment voucher deleted successfully',
    }
  } catch (error) {
    return {
      success: false,
      message: `Failed to delete payment voucher: ${error instanceof Error ? error.message : 'Unknown error'}`,
    }
  }
}

/**
 * Get Payment Voucher Statistics
 */
export async function getPaymentVoucherStats(): Promise<{
  success: boolean
  data?: PaymentVoucherStats
  message: string
  status: number
}> {
  try {
    const stats: PaymentVoucherStats = {
      total: mockPaymentVouchers.length,
      draft: mockPaymentVouchers.filter((pv) => pv.status === 'DRAFT').length,
      submitted: mockPaymentVouchers.filter((pv) => pv.status === 'SUBMITTED').length,
      inApproval: mockPaymentVouchers.filter((pv) => pv.status === 'IN_REVIEW').length,
      approved: mockPaymentVouchers.filter((pv) => pv.status === 'APPROVED').length,
      rejected: mockPaymentVouchers.filter((pv) => pv.status === 'REJECTED').length,
      paid: mockPaymentVouchers.filter((pv) => pv.status === 'PAID').length,
      totalValue: mockPaymentVouchers.reduce((sum, pv) => sum + pv.totalAmount, 0),
      totalPaid: mockPaymentVouchers.reduce((sum, pv) => sum + (pv.paidAmount || 0), 0),
      pendingPayment: mockPaymentVouchers
        .filter((pv) => pv.status === 'APPROVED')
        .reduce((sum, pv) => sum + pv.totalAmount, 0),
      averageApprovalTime: 5, // Placeholder - would calculate from action history timestamps
    }

    return {
      success: true,
      data: stats,
      message: 'Payment voucher statistics retrieved successfully',
      status: 200,
    }
  } catch (error) {
    return {
      success: false,
      message: `Failed to retrieve payment voucher statistics: ${error instanceof Error ? error.message : 'Unknown error'}`,
      status: 500,
    }
  }
}
