/**
 * Approval Flow React Query Hooks
 *
 * Combined hooks for managing complete approval workflows.
 * Handles approval, rejection, reassignment, and related operations.
 */

'use client';

import { useCallback, useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import {
  WorkflowAssignment,
  ApproveStageRequest,
  RejectStageRequest,
  ReassignStageRequest,
} from '@/types';

import {
  approveStageAction,
  rejectStageAction,
  reassignStageAction,
} from '@/app/_actions/workflows';

import {
  markNotificationAsRead,
  markActionTaken,
} from '@/app/_actions/notifications';

import {
  useAssignmentWithWorkflow,
  useInvalidatePendingApprovals,
} from './use-workflows';

/**
 * Hook: Approve current stage
 * Handles approval with signature and comments
 * @returns Mutation for approval
 */
export function useApproveStage() {
  const queryClient = useQueryClient();
  const invalidatePending = useInvalidatePendingApprovals();

  return useMutation({
    mutationFn: async (request: ApproveStageRequest) =>
      approveStageAction(request),
    onSuccess: (data, variables) => {
      // Invalidate assignment cache
      queryClient.invalidateQueries({
        queryKey: ['assignments', variables.assignmentId],
      });

      // Invalidate pending approvals for current user
      invalidatePending(variables.approverId);

      // Invalidate pending for next approver if exists
      if (data.nextApprover) {
        invalidatePending(data.nextApprover.userId);
      }
    },
    onError: (error) => {
      console.error('Failed to approve stage:', error);
    },
  });
}

/**
 * Hook: Reject current stage
 * Handles rejection with remarks
 * @returns Mutation for rejection
 */
export function useRejectStage() {
  const queryClient = useQueryClient();
  const invalidatePending = useInvalidatePendingApprovals();

  return useMutation({
    mutationFn: async (request: RejectStageRequest) =>
      rejectStageAction(request),
    onSuccess: (data, variables) => {
      // Invalidate assignment
      queryClient.invalidateQueries({
        queryKey: ['assignments', variables.assignmentId],
      });

      // Invalidate pending for rejector
      invalidatePending(variables.rejectorId);

      // Invalidate for creator (might be returned to them)
      queryClient.invalidateQueries({
        queryKey: ['pending-approvals'],
      });
    },
    onError: (error) => {
      console.error('Failed to reject stage:', error);
    },
  });
}

/**
 * Hook: Reassign stage to different approver
 * @returns Mutation for reassignment
 */
export function useReassignStage() {
  const queryClient = useQueryClient();
  const invalidatePending = useInvalidatePendingApprovals();

  return useMutation({
    mutationFn: async (request: ReassignStageRequest) =>
      reassignStageAction(request),
    onSuccess: (data, variables) => {
      // Invalidate assignment
      queryClient.invalidateQueries({
        queryKey: ['assignments', variables.assignmentId],
      });

      // Invalidate for old approver (task gone)
      if (data.oldApprover) {
        invalidatePending(data.oldApprover.userId);
      }

      // Invalidate for new approver (new task)
      invalidatePending(variables.newApproverId);
    },
    onError: (error) => {
      console.error('Failed to reassign stage:', error);
    },
  });
}

/**
 * Hook: Complete approval flow (approve + mark notification)
 * Combined operation for quick approval
 * @returns Mutation for complete approval
 */
export function useQuickApprove() {
  const queryClient = useQueryClient();
  const approveStage = useApproveStage();
  const invalidatePending = useInvalidatePendingApprovals();

  return useMutation({
    mutationFn: async (params: {
      approveRequest: ApproveStageRequest;
      notificationId?: string;
    }) => {
      const approvalResult = await approveStage.mutateAsync(
        params.approveRequest
      );

      // Mark notification as read and action taken
      if (params.notificationId) {
        await Promise.all([
          markNotificationAsRead({ notificationId: params.notificationId }),
          markActionTaken(params.notificationId),
        ]);
      }

      return approvalResult;
    },
    onSuccess: (data, variables) => {
      // Invalidate all related caches
      queryClient.invalidateQueries({
        queryKey: ['notifications'],
      });
      invalidatePending(variables.approveRequest.approverId);
    },
  });
}

/**
 * Hook: Approval modal state and submission
 * Manages form state, validation, and submission
 * @returns Modal state and handlers
 */
export function useApprovalModal() {
  const [isOpen, setIsOpen] = useState(false);
  const [remarks, setRemarks] = useState('');
  const [signature, setSignature] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const approveMutation = useApproveStage();

  const handleSubmit = useCallback(
    async (request: Omit<ApproveStageRequest, 'comments' | 'signature'>) => {
      if (!signature) {
        throw new Error('Signature is required');
      }

      setIsSubmitting(true);
      try {
        await approveMutation.mutateAsync({
          ...request,
          comments: remarks,
          signature,
        });

        // Reset form
        setRemarks('');
        setSignature(null);
        setIsOpen(false);

        return true;
      } catch (error) {
        throw error;
      } finally {
        setIsSubmitting(false);
      }
    },
    [signature, remarks, approveMutation]
  );

  return {
    isOpen,
    setIsOpen,
    remarks,
    setRemarks,
    signature,
    setSignature,
    isSubmitting,
    handleSubmit,
    isValid: !!signature && remarks.length > 0,
  };
}

/**
 * Hook: Rejection modal state and submission
 * Manages rejection form state
 * @returns Modal state and handlers
 */
export function useRejectionModal() {
  const [isOpen, setIsOpen] = useState(false);
  const [remarks, setRemarks] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  const rejectMutation = useRejectStage();

  const handleSubmit = useCallback(
    async (request: Omit<RejectStageRequest, 'rejectionRemarks'>) => {
      if (!remarks.trim()) {
        throw new Error('Rejection reason is required');
      }

      setIsSubmitting(true);
      try {
        await rejectMutation.mutateAsync({
          ...request,
          rejectionRemarks: remarks,
        });

        // Reset form
        setRemarks('');
        setIsOpen(false);

        return true;
      } catch (error) {
        throw error;
      } finally {
        setIsSubmitting(false);
      }
    },
    [remarks, rejectMutation]
  );

  return {
    isOpen,
    setIsOpen,
    remarks,
    setRemarks,
    isSubmitting,
    handleSubmit,
    isValid: remarks.trim().length > 0,
  };
}

/**
 * Hook: Reassignment modal state and submission
 * Manages reassignment form state and permissions
 * @returns Modal state and handlers
 */
export function useReassignmentModal() {
  const [isOpen, setIsOpen] = useState(false);
  const [newApproverId, setNewApproverId] = useState('');
  const [reason, setReason] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  const reassignMutation = useReassignStage();

  const handleSubmit = useCallback(
    async (
      request: Omit<
        ReassignStageRequest,
        'newApproverId' | 'reassignmentReason'
      >
    ) => {
      if (!newApproverId) {
        throw new Error('New approver is required');
      }

      setIsSubmitting(true);
      try {
        await reassignMutation.mutateAsync({
          ...request,
          newApproverId,
          reassignmentReason: reason,
        });

        // Reset form
        setNewApproverId('');
        setReason('');
        setIsOpen(false);

        return true;
      } catch (error) {
        throw error;
      } finally {
        setIsSubmitting(false);
      }
    },
    [newApproverId, reason, reassignMutation]
  );

  return {
    isOpen,
    setIsOpen,
    newApproverId,
    setNewApproverId,
    reason,
    setReason,
    isSubmitting,
    handleSubmit,
    isValid: !!newApproverId,
  };
}

/**
 * Hook: Check permissions for approval actions
 * @param entityId Entity ID
 * @param userId Current user ID
 * @param userRole Current user role
 * @returns Permissions object
 */
export function useApprovalPermissions(
  entityId: string,
  userId: string,
  userRole: string
) {
  const { assignment, workflow, currentStage } =
    useAssignmentWithWorkflow(entityId);

  const canApprove =
    assignment &&
    currentStage &&
    (assignment.currentStageNumber === currentStage.stageNumber);

  const canReject =
    canApprove && (currentStage?.canBeRejected ?? true);

  const canReassign =
    canApprove &&
    (currentStage?.canBeReassigned ?? true) &&
    (userId === assignment.currentStageNumber ||
      userRole === 'ADMIN');

  return {
    canApprove,
    canReject,
    canReassign,
    assignment,
    currentStage,
    isLoading: !assignment || !currentStage,
  };
}

/**
 * Hook: Get approval stage history
 * Gets all executions for current assignment
 * @param assignment Workflow assignment
 * @returns Formatted history
 */
export function useApprovalHistory(assignment: WorkflowAssignment | undefined) {
  return {
    history: assignment?.stageHistory || [],
    totalApprovals: assignment?.stageHistory.filter(
      (h) => h.status === 'APPROVED'
    ).length || 0,
    totalRejections: assignment?.stageHistory.filter(
      (h) => h.status === 'REJECTED'
    ).length || 0,
    currentStageNumber: assignment?.currentStageNumber || 0,
  };
}

/**
 * Hook: Approval action buttons state
 * Determines which buttons to show based on permissions
 * @param entityId Entity ID
 * @param userId Current user ID
 * @param userRole Current user role
 * @returns Button states
 */
export function useApprovalActions(
  entityId: string,
  userId: string,
  userRole: string
) {
  const permissions = useApprovalPermissions(entityId, userId, userRole);
  const approveMutation = useApproveStage();
  const rejectMutation = useRejectStage();
  const reassignMutation = useReassignStage();

  return {
    ...permissions,
    approveMutation,
    rejectMutation,
    reassignMutation,
    isLoading:
      approveMutation.isPending ||
      rejectMutation.isPending ||
      reassignMutation.isPending,
    error:
      approveMutation.error ||
      rejectMutation.error ||
      reassignMutation.error,
  };
}

/**
 * Hook: Integration with quick action from notification
 * Combines notification handling with approval
 * @returns Handlers for quick actions
 */
export function useQuickAction() {
  const approveModal = useApprovalModal();
  const rejectModal = useRejectionModal();
  const reassignModal = useReassignmentModal();

  return {
    approval: approveModal,
    rejection: rejectModal,
    reassignment: reassignModal,
    openApprovalModal: () => approveModal.setIsOpen(true),
    openRejectionModal: () => rejectModal.setIsOpen(true),
    openReassignmentModal: () => reassignModal.setIsOpen(true),
  };
}

/**
 * Hook: Stage completion status
 * Check if stage is completed and provides info
 * @param assignment Workflow assignment
 * @param stageNumber Stage number
 * @returns Completion status
 */
export function useStageCompletion(
  assignment: WorkflowAssignment | undefined,
  stageNumber: number
) {
  const execution = assignment?.stageHistory.find(
    (h) => h.stageNumber === stageNumber
  );

  return {
    isCompleted: !!execution,
    status: execution?.status || 'PENDING',
    completedAt: execution?.createdAt,
    approvedBy: execution?.approvedBy,
    remarks: execution?.remarks,
    signature: execution?.signature,
  };
}

/**
 * Hook: Next stage preview
 * Shows what will happen on approval
 * @param assignment Workflow assignment
 * @param workflow Workflow definition
 * @returns Next stage info
 */
export function useNextStagePreview(
  assignment: WorkflowAssignment | undefined,
  workflow: any
) {
  if (!assignment || !workflow) {
    return null;
  }

  const currentStage = workflow.stages.find(
    (s: any) => s.stageNumber === assignment.currentStageNumber
  );

  if (!currentStage) {
    return null;
  }

  const nextStageNumber = currentStage.onApprove?.nextStage;

  if (nextStageNumber === 'FINAL') {
    return {
      isFinal: true,
      stageName: 'Workflow Complete',
      approverName: null,
    };
  }

  const nextStage = workflow.stages.find(
    (s: any) => s.stageNumber === nextStageNumber
  );

  return {
    isFinal: false,
    stageNumber: nextStageNumber,
    stageName: nextStage?.stageName || 'Unknown',
    approverType: nextStage?.approverAssignmentType,
    requiredRole: nextStage?.requiredRole,
  };
}
