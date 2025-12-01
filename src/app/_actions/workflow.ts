'use server'

import { auth } from '@/auth';
import { APIResponse } from '@/types';
import {
  WorkflowDocument,
  WorkflowDocumentType,
  DocumentStatus,
  ApprovalLogEntry,
  Attachment,
  Approver,
  WorkflowStep,
  PaginatedResponse,
  User,
} from '@/types/workflow';
import {
  createMockPurchaseOrder,
  createMockPaymentVoucher,
  createMockRequisitionForm,
  createMockApprover,
  createMockApprovalLogEntry,
  createMockAttachment,
  getRandomUserByRole,
  MOCK_USERS,
  generateDocumentNumber,
} from '@/lib/mock-data';
import { getWorkflowStepsForType, getApprovalChain, getNextApproverRole, isLastApprovalStep } from '@/lib/rbac';
import { unauthorizedResponse, handleError } from '@/app/_actions/api-config';
import {
  documentStore,
  approversStore,
  approvalLogsStore,
  attachmentsStore,
} from '@/lib/workflow-stores';

// Note: documentStore is not exported from this server action file
// It's available directly from @/lib/workflow-stores for non-server code
// Initialization happens in @/lib/workflow-initialization

// =============== DOCUMENT OPERATIONS ===============

export async function createWorkflowDocument(
  documentType: WorkflowDocumentType,
  formData: Record<string, any>
): Promise<APIResponse<WorkflowDocument | null>> {
  const session = await auth();
  if (!session?.user) return unauthorizedResponse();

  try {
    let document: WorkflowDocument;

    // Create appropriate document type
    switch (documentType) {
      case 'PURCHASE_ORDER':
        document = createMockPurchaseOrder({
          createdBy: session.user.id,
        });
        break;
      case 'PAYMENT_VOUCHER':
        document = createMockPaymentVoucher({
          createdBy: session.user.id,
        });
        break;
      case 'REQUISITION':
        document = createMockRequisitionForm({
          createdBy: session.user.id,
        });
        break;
      default:
        return {
          success: false,
          message: 'Invalid document type',
          data: null,
          status: 400,
          statusText: 'BAD REQUEST',
        };
    }

    // Store document
    documentStore.set(document.id, document);
    approversStore.set(document.id, []);
    approvalLogsStore.set(document.id, []);
    attachmentsStore.set(document.id, []);

    console.log(`✅ Created ${documentType}: ${document.documentNumber}`);

    return {
      success: true,
      message: `${documentType} created successfully`,
      data: document,
      status: 201,
      statusText: 'CREATED',
    };
  } catch (error) {
    console.error('Error creating document:', error);
    return handleError(error, 'POST', '/workflows/documents');
  }
}

export async function submitDocument(
  documentId: string
): Promise<APIResponse<WorkflowDocument | null>> {
  const session = await auth();
  if (!session?.user) return unauthorizedResponse();

  try {
    const document = documentStore.get(documentId);
    if (!document) {
      return {
        success: false,
        message: 'Document not found',
        data: null,
        status: 404,
        statusText: 'NOT FOUND',
      };
    }

    // Update document status
    document.status = 'SUBMITTED';
    document.currentStage = 1;
    document.updatedAt = new Date();

    // Auto-assign first approvers based on workflow steps
    const steps = getWorkflowStepsForType(document.type);
    const firstStep = steps[0];

    if (firstStep) {
      const approver = getRandomUserByRole(firstStep.roleName);
      const approverAssignment = createMockApprover(documentId, 1, approver);
      const currentApprovers = approversStore.get(documentId) || [];
      currentApprovers.push(approverAssignment);
      approversStore.set(documentId, currentApprovers);
    }

    documentStore.set(documentId, document);

    console.log(`✅ Document submitted: ${document.documentNumber}`);

    return {
      success: true,
      message: 'Document submitted for approval',
      data: document,
      status: 200,
      statusText: 'OK',
    };
  } catch (error) {
    console.error('Error submitting document:', error);
    return handleError(error, 'PUT', `/workflows/documents/${documentId}/submit`);
  }
}

export async function getDocument(
  documentId: string
): Promise<APIResponse<WorkflowDocument | null>> {
  const session = await auth();
  if (!session?.user) return unauthorizedResponse();

  try {
    const document = documentStore.get(documentId);
    if (!document) {
      return {
        success: false,
        message: 'Document not found',
        data: null,
        status: 404,
        statusText: 'NOT FOUND',
      };
    }

    return {
      success: true,
      message: 'Document retrieved successfully',
      data: document,
      status: 200,
      statusText: 'OK',
    };
  } catch (error) {
    console.error('Error fetching document:', error);
    return handleError(error, 'GET', `/workflows/documents/${documentId}`);
  }
}

export async function updateDocumentDraft(
  documentId: string,
  formData: Record<string, any>
): Promise<APIResponse<WorkflowDocument | null>> {
  const session = await auth();
  if (!session?.user) return unauthorizedResponse();

  try {
    const document = documentStore.get(documentId);
    if (!document) {
      return {
        success: false,
        message: 'Document not found',
        data: null,
        status: 404,
        statusText: 'NOT FOUND',
      };
    }

    if (document.status !== 'DRAFT' && document.status !== 'REJECTED') {
      return {
        success: false,
        message: 'Cannot edit document in current status',
        data: null,
        status: 400,
        statusText: 'BAD REQUEST',
      };
    }

    // Update metadata
    document.metadata = { ...document.metadata, ...formData };
    document.updatedAt = new Date();

    documentStore.set(documentId, document);

    console.log(`✅ Document updated: ${document.documentNumber}`);

    return {
      success: true,
      message: 'Document updated successfully',
      data: document,
      status: 200,
      statusText: 'OK',
    };
  } catch (error) {
    console.error('Error updating document:', error);
    return handleError(error, 'PATCH', `/workflows/documents/${documentId}`);
  }
}

export async function getDocumentsByCreator(
  userId: string,
  page: number = 1,
  limit: number = 10
): Promise<APIResponse<PaginatedResponse<WorkflowDocument>>> {
  const session = await auth();
  if (!session?.user) return unauthorizedResponse();

  try {
    const documents = Array.from(documentStore.values()).filter(
      (doc) => doc.createdBy === userId
    );

    const total = documents.length;
    const totalPages = Math.ceil(total / limit);
    const start = (page - 1) * limit;
    const paginatedDocs = documents.slice(start, start + limit);

    return {
      success: true,
      message: 'Documents retrieved successfully',
      data: {
        data: paginatedDocs,
        pagination: {
          page,
          limit,
          total,
          totalPages,
        },
      },
      status: 200,
      statusText: 'OK',
    };
  } catch (error) {
    console.error('Error fetching documents:', error);
    return handleError(error, 'GET', `/workflows/documents`);
  }
}

// =============== APPROVAL OPERATIONS ===============

export async function approveDocument(
  documentId: string,
  comments?: string
): Promise<APIResponse<ApprovalLogEntry | null>> {
  const session = await auth();
  if (!session?.user) return unauthorizedResponse();

  try {
    const document = documentStore.get(documentId);
    if (!document) {
      return {
        success: false,
        message: 'Document not found',
        data: null,
        status: 404,
        statusText: 'NOT FOUND',
      };
    }

    if (document.status !== 'IN_REVIEW') {
      return {
        success: false,
        message: 'Document is not pending approval',
        data: null,
        status: 400,
        statusText: 'BAD REQUEST',
      };
    }

    // Create approval log entry
    const approver = MOCK_USERS.ADMIN[0]; // Mock approver
    const logEntry = createMockApprovalLogEntry(documentId, approver, {
      action: 'APPROVED',
      comments,
    });

    const logs = approvalLogsStore.get(documentId) || [];
    logs.push(logEntry);
    approvalLogsStore.set(documentId, logs);

    // Update approver status
    const approvers = approversStore.get(documentId) || [];
    const currentApprover = approvers.find(
      (a) => a.stepOrder === document.currentStage && a.status === 'PENDING'
    );
    if (currentApprover) {
      currentApprover.status = 'APPROVED';
    }

    // Check if this is the last approval step
    if (isLastApprovalStep(document.type, document.currentStage)) {
      document.status = 'APPROVED';
      document.currentStage = document.currentStage;
    } else {
      // Move to next step
      const nextRole = getNextApproverRole(document.type, document.currentStage);
      if (nextRole) {
        document.currentStage += 1;
        document.status = 'IN_REVIEW';

        // Assign next approver
        const nextApprover = getRandomUserByRole(nextRole);
        const nextApproverAssignment = createMockApprover(
          documentId,
          document.currentStage,
          nextApprover
        );
        approvers.push(nextApproverAssignment);
        approversStore.set(documentId, approvers);
      }
    }

    document.updatedAt = new Date();
    documentStore.set(documentId, document);

    console.log(
      `✅ Document approved: ${document.documentNumber} (Stage: ${document.currentStage})`
    );

    return {
      success: true,
      message: 'Document approved successfully',
      data: logEntry,
      status: 200,
      statusText: 'OK',
    };
  } catch (error) {
    console.error('Error approving document:', error);
    return handleError(error, 'POST', `/workflows/documents/${documentId}/approve`);
  }
}

export async function rejectDocument(
  documentId: string,
  reason: string
): Promise<APIResponse<ApprovalLogEntry | null>> {
  const session = await auth();
  if (!session?.user) return unauthorizedResponse();

  try {
    const document = documentStore.get(documentId);
    if (!document) {
      return {
        success: false,
        message: 'Document not found',
        data: null,
        status: 404,
        statusText: 'NOT FOUND',
      };
    }

    if (document.status !== 'IN_REVIEW') {
      return {
        success: false,
        message: 'Document is not pending approval',
        data: null,
        status: 400,
        statusText: 'BAD REQUEST',
      };
    }

    // Create rejection log entry
    const approver = MOCK_USERS.ADMIN[0]; // Mock approver
    const logEntry = createMockApprovalLogEntry(documentId, approver, {
      action: 'REJECTED',
      comments: reason,
    });

    const logs = approvalLogsStore.get(documentId) || [];
    logs.push(logEntry);
    approvalLogsStore.set(documentId, logs);

    // Update approver status
    const approvers = approversStore.get(documentId) || [];
    const currentApprover = approvers.find(
      (a) => a.stepOrder === document.currentStage && a.status === 'PENDING'
    );
    if (currentApprover) {
      currentApprover.status = 'REJECTED';
    }

    // Reset document to draft
    document.status = 'REJECTED';
    document.currentStage = 0;
    document.updatedAt = new Date();
    documentStore.set(documentId, document);

    console.log(`❌ Document rejected: ${document.documentNumber}`);

    return {
      success: true,
      message: 'Document rejected successfully',
      data: logEntry,
      status: 200,
      statusText: 'OK',
    };
  } catch (error) {
    console.error('Error rejecting document:', error);
    return handleError(error, 'POST', `/workflows/documents/${documentId}/reject`);
  }
}

export async function getApprovalLog(
  documentId: string
): Promise<APIResponse<ApprovalLogEntry[]>> {
  const session = await auth();
  if (!session?.user) return unauthorizedResponse();

  try {
    const logs = approvalLogsStore.get(documentId) || [];

    return {
      success: true,
      message: 'Approval logs retrieved successfully',
      data: logs,
      status: 200,
      statusText: 'OK',
    };
  } catch (error) {
    console.error('Error fetching approval logs:', error);
    return handleError(error, 'GET', `/workflows/documents/${documentId}/approval-log`);
  }
}

export async function getPendingApprovals(
  userRole: string
): Promise<APIResponse<WorkflowDocument[]>> {
  const session = await auth();
  if (!session?.user) return unauthorizedResponse();

  try {
    const pendingDocs = Array.from(documentStore.values())
      .filter((doc) => doc.status === 'IN_REVIEW')
      .filter((doc) => {
        const approvers = approversStore.get(doc.id) || [];
        return approvers.some(
          (a) => a.stepOrder === doc.currentStage && a.role === userRole
        );
      });

    console.log(`✅ Found ${pendingDocs.length} pending approvals for ${userRole}`);

    return {
      success: true,
      message: 'Pending approvals retrieved successfully',
      data: pendingDocs,
      status: 200,
      statusText: 'OK',
    };
  } catch (error) {
    console.error('Error fetching pending approvals:', error);
    return handleError(error, 'GET', `/workflows/pending-approvals`);
  }
}

// =============== APPROVER MANAGEMENT ===============

export async function assignApprover(
  documentId: string,
  stepOrder: number,
  userId: string,
  role: string
): Promise<APIResponse<Approver | null>> {
  const session = await auth();
  if (!session?.user) return unauthorizedResponse();

  try {
    const document = documentStore.get(documentId);
    if (!document) {
      return {
        success: false,
        message: 'Document not found',
        data: null,
        status: 404,
        statusText: 'NOT FOUND',
      };
    }

    const user = Object.values(MOCK_USERS)
      .flat()
      .find((u) => u.id === userId);

    if (!user) {
      return {
        success: false,
        message: 'User not found',
        data: null,
        status: 404,
        statusText: 'NOT FOUND',
      };
    }

    const approverAssignment = createMockApprover(documentId, stepOrder, user, {
      role: role as any,
    });

    const approvers = approversStore.get(documentId) || [];
    approvers.push(approverAssignment);
    approversStore.set(documentId, approvers);

    console.log(`✅ Approver assigned: ${user.name} to step ${stepOrder}`);

    return {
      success: true,
      message: 'Approver assigned successfully',
      data: approverAssignment,
      status: 201,
      statusText: 'CREATED',
    };
  } catch (error) {
    console.error('Error assigning approver:', error);
    return handleError(error, 'POST', `/workflows/documents/${documentId}/approvers`);
  }
}

export async function reassignApprover(
  documentId: string,
  approverId: string,
  newUserId: string
): Promise<APIResponse<Approver | null>> {
  const session = await auth();
  if (!session?.user) return unauthorizedResponse();

  try {
    const approvers = approversStore.get(documentId) || [];
    const approver = approvers.find((a) => a.id === approverId);

    if (!approver) {
      return {
        success: false,
        message: 'Approver assignment not found',
        data: null,
        status: 404,
        statusText: 'NOT FOUND',
      };
    }

    if (!approver.canReassign) {
      return {
        success: false,
        message: 'This approver cannot be reassigned',
        data: null,
        status: 400,
        statusText: 'BAD REQUEST',
      };
    }

    const newUser = Object.values(MOCK_USERS)
      .flat()
      .find((u) => u.id === newUserId);

    if (!newUser) {
      return {
        success: false,
        message: 'New user not found',
        data: null,
        status: 404,
        statusText: 'NOT FOUND',
      };
    }

    approver.userId = newUserId;
    approver.user = newUser;
    approver.assignedAt = new Date();

    // Log the reassignment
    const logEntry = createMockApprovalLogEntry(documentId, newUser, {
      action: 'REASSIGNED',
      comments: `Reassigned from approver to ${newUser.name}`,
    });

    const logs = approvalLogsStore.get(documentId) || [];
    logs.push(logEntry);
    approvalLogsStore.set(documentId, logs);

    approversStore.set(documentId, approvers);

    console.log(`✅ Approver reassigned to: ${newUser.name}`);

    return {
      success: true,
      message: 'Approver reassigned successfully',
      data: approver,
      status: 200,
      statusText: 'OK',
    };
  } catch (error) {
    console.error('Error reassigning approver:', error);
    return handleError(
      error,
      'PATCH',
      `/workflows/documents/${documentId}/approvers/${approverId}`
    );
  }
}

export async function getDocumentApprovers(
  documentId: string
): Promise<APIResponse<Approver[]>> {
  const session = await auth();
  if (!session?.user) return unauthorizedResponse();

  try {
    const approvers = approversStore.get(documentId) || [];

    return {
      success: true,
      message: 'Approvers retrieved successfully',
      data: approvers,
      status: 200,
      statusText: 'OK',
    };
  } catch (error) {
    console.error('Error fetching approvers:', error);
    return handleError(error, 'GET', `/workflows/documents/${documentId}/approvers`);
  }
}

// =============== ATTACHMENT OPERATIONS ===============

export async function uploadAttachment(
  documentId: string,
  fileName: string,
  fileSize: number,
  fileType: string,
  visibleToRoles: string[]
): Promise<APIResponse<Attachment | null>> {
  const session = await auth();
  if (!session?.user) return unauthorizedResponse();

  try {
    const document = documentStore.get(documentId);
    if (!document) {
      return {
        success: false,
        message: 'Document not found',
        data: null,
        status: 404,
        statusText: 'NOT FOUND',
      };
    }

    const currentUser = Object.values(MOCK_USERS)
      .flat()
      .find((u) => u.id === session.user.id) || MOCK_USERS.ADMIN[0];

    const attachment = createMockAttachment(documentId, currentUser, {
      fileName,
      fileSize,
      fileType,
      visibleToRoles: visibleToRoles as any,
    });

    const attachments = attachmentsStore.get(documentId) || [];
    attachments.push(attachment);
    attachmentsStore.set(documentId, attachments);

    console.log(`✅ Attachment uploaded: ${fileName}`);

    return {
      success: true,
      message: 'Attachment uploaded successfully',
      data: attachment,
      status: 201,
      statusText: 'CREATED',
    };
  } catch (error) {
    console.error('Error uploading attachment:', error);
    return handleError(error, 'POST', `/workflows/documents/${documentId}/attachments`);
  }
}

export async function getAttachments(
  documentId: string
): Promise<APIResponse<Attachment[]>> {
  const session = await auth();
  if (!session?.user) return unauthorizedResponse();

  try {
    const attachments = attachmentsStore.get(documentId) || [];

    return {
      success: true,
      message: 'Attachments retrieved successfully',
      data: attachments,
      status: 200,
      statusText: 'OK',
    };
  } catch (error) {
    console.error('Error fetching attachments:', error);
    return handleError(error, 'GET', `/workflows/documents/${documentId}/attachments`);
  }
}

export async function deleteAttachment(
  documentId: string,
  attachmentId: string
): Promise<APIResponse> {
  const session = await auth();
  if (!session?.user) return unauthorizedResponse();

  try {
    const attachments = attachmentsStore.get(documentId) || [];
    const filtered = attachments.filter((a) => a.id !== attachmentId);
    attachmentsStore.set(documentId, filtered);

    console.log(`✅ Attachment deleted: ${attachmentId}`);

    return {
      success: true,
      message: 'Attachment deleted successfully',
      data: null,
      status: 200,
      statusText: 'OK',
    };
  } catch (error) {
    console.error('Error deleting attachment:', error);
    return handleError(
      error,
      'DELETE',
      `/workflows/documents/${documentId}/attachments/${attachmentId}`
    );
  }
}

// =============== WORKFLOW CONFIGURATION ===============

export async function getWorkflowSteps(
  documentType: WorkflowDocumentType
): Promise<APIResponse<WorkflowStep[]>> {
  try {
    const steps = getWorkflowStepsForType(documentType);

    return {
      success: true,
      message: 'Workflow steps retrieved successfully',
      data: steps,
      status: 200,
      statusText: 'OK',
    };
  } catch (error) {
    console.error('Error fetching workflow steps:', error);
    return handleError(error, 'GET', `/workflows/steps/${documentType}`);
  }
}

// =============== DASHBOARD & REPORTING ===============

export async function getDashboardStats(
  userId: string
): Promise<
  APIResponse<{
    createdDocuments: number;
    pendingApprovals: number;
    approvedDocuments: number;
    rejectedDocuments: number;
  }>
> {
  const session = await auth();
  if (!session?.user) return unauthorizedResponse();

  try {
    const allDocs = Array.from(documentStore.values());
    const createdDocuments = allDocs.filter((d) => d.createdBy === userId).length;
    const approvedDocuments = allDocs.filter((d) => d.status === 'APPROVED').length;
    const rejectedDocuments = allDocs.filter((d) => d.status === 'REJECTED').length;
    const pendingApprovals = allDocs.filter((d) => d.status === 'IN_REVIEW').length;

    return {
      success: true,
      message: 'Dashboard stats retrieved successfully',
      data: {
        createdDocuments,
        pendingApprovals,
        approvedDocuments,
        rejectedDocuments,
      },
      status: 200,
      statusText: 'OK',
    };
  } catch (error) {
    console.error('Error fetching dashboard stats:', error);
    return handleError(error, 'GET', `/workflows/dashboard/stats`);
  }
}

export async function getAuditLog(
  documentId: string
): Promise<APIResponse<ApprovalLogEntry[]>> {
  const session = await auth();
  if (!session?.user) return unauthorizedResponse();

  try {
    const logs = approvalLogsStore.get(documentId) || [];
    const sortedLogs = logs.sort(
      (a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime()
    );

    return {
      success: true,
      message: 'Audit log retrieved successfully',
      data: sortedLogs,
      status: 200,
      statusText: 'OK',
    };
  } catch (error) {
    console.error('Error fetching audit log:', error);
    return handleError(error, 'GET', `/workflows/documents/${documentId}/audit-log`);
  }
}
