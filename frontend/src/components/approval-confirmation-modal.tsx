'use client';

import { useState, useRef } from 'react';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';
import { SignatureCanvas } from '@/components/ui/signature-canvas';
import type { SignatureCanvasHandle } from '@/components/ui/signature-canvas';
import { AlertCircle, CheckCircle2 } from 'lucide-react';

export type ApprovalAction = 'approve' | 'reject';

interface ApprovalConfirmationModalProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: (data: ApprovalData) => Promise<void>;
  action: ApprovalAction;
  documentTitle: string;
  documentNumber?: string;
  isLoading?: boolean;
}

export interface ApprovalData {
  signature: string;
  remarks?: string; // Required for rejection
  comments?: string; // Optional for both
}

/**
 * Reusable modal for approval and rejection of documents
 * Ensures consistent UX across all approval workflows
 *
 * @example
 * const [modalOpen, setModalOpen] = useState(false)
 * const handleApprove = async (data) => {
 *   await useApproveBudget(budgetId).mutateAsync({
 *     signature: data.signature,
 *     comments: data.comments
 *   })
 * }
 *
 * <ApprovalConfirmationModal
 *   isOpen={modalOpen}
 *   onClose={() => setModalOpen(false)}
 *   onConfirm={handleApprove}
 *   action="approve"
 *   documentTitle="Q1 2024 Budget"
 *   documentNumber="BDG-2024-001"
 * />
 */
export function ApprovalConfirmationModal({
  isOpen,
  onClose,
  onConfirm,
  action,
  documentTitle,
  documentNumber,
  isLoading = false,
}: ApprovalConfirmationModalProps) {
  const [signature, setSignature] = useState<string>('');
  const [remarks, setRemarks] = useState<string>('');
  const [comments, setComments] = useState<string>('');
  const [error, setError] = useState<string>('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const signatureRef = useRef<SignatureCanvasHandle>(null);

  const isApproval = action === 'approve';
  const isRejection = action === 'reject';

  const handleClearSignature = () => {
    setSignature('');
    signatureRef.current?.clearSignature();
  };

  const handleSubmit = async () => {
    setError('');

    // Validation
    if (!signature) {
      setError('Signature is required');
      return;
    }

    if (isRejection && !remarks.trim()) {
      setError('Remarks are required for rejection');
      return;
    }

    setIsSubmitting(true);
    try {
      await onConfirm({
        signature,
        remarks: remarks || undefined,
        comments: comments || undefined,
      });

      // Reset form on success
      setSignature('');
      setRemarks('');
      setComments('');
      onClose();
    } catch (err) {
      // Error handling is done in the mutation hook
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleOpenChange = (open: boolean) => {
    if (!open) {
      // Reset when closing
      setSignature('');
      setRemarks('');
      setComments('');
      setError('');
      onClose();
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={handleOpenChange}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <div className="flex items-start gap-3">
            {isApproval ? (
              <CheckCircle2 className="h-6 w-6 text-green-600 flex-shrink-0 mt-1" />
            ) : (
              <AlertCircle className="h-6 w-6 text-red-600 flex-shrink-0 mt-1" />
            )}
            <div>
              <DialogTitle className="text-lg">
                {isApproval ? 'Approve' : 'Reject'} Document
              </DialogTitle>
              <DialogDescription>
                <span className="block font-semibold text-foreground mt-1">
                  {documentTitle}
                </span>
                {documentNumber && (
                  <span className="block text-sm text-muted-foreground mt-1">
                    Document: {documentNumber}
                  </span>
                )}
              </DialogDescription>
            </div>
          </div>
        </DialogHeader>

        <div className="space-y-6 py-4">
          {/* Rejection Remarks - Required for rejection, not shown for approval */}
          {isRejection && (
            <div className="space-y-2">
              <label className="block text-sm font-semibold text-foreground">
                Remarks <span className="text-red-600">*</span>
              </label>
              <p className="text-sm text-muted-foreground">
                Please provide detailed remarks explaining why this document is being rejected.
              </p>
              <Textarea
                placeholder="Enter rejection remarks (required)..."
                value={remarks}
                onChange={(e) => {
                  setRemarks(e.target.value);
                  if (error === 'Remarks are required for rejection') {
                    setError('');
                  }
                }}
                className="min-h-[100px]"
              />
            </div>
          )}

          {/* Optional Comments - for both approval and rejection */}
          <div className="space-y-2">
            <label className="block text-sm font-semibold text-foreground">
              {isApproval ? 'Approval' : 'Additional'} Comments{' '}
              <span className="text-muted-foreground font-normal">(Optional)</span>
            </label>
            <p className="text-sm text-muted-foreground">
              {isApproval
                ? 'Add any comments or notes to accompany your approval.'
                : 'Provide additional context or instructions.'}
            </p>
            <Textarea
              placeholder="Enter optional comments..."
              value={comments}
              onChange={(e) => setComments(e.target.value)}
              className="min-h-[80px]"
            />
          </div>

          {/* Signature - Required for both */}
          <div className="space-y-2">
            <label className="block text-sm font-semibold text-foreground">
              Digital Signature <span className="text-red-600">*</span>
            </label>
            <p className="text-sm text-muted-foreground">
              Draw your signature in the box below. This is required to {isApproval ? 'approve' : 'reject'} the
              document.
            </p>

            <div className="border-2 border-dashed border-muted-foreground/30 rounded-lg overflow-hidden">
              <SignatureCanvas
                ref={signatureRef}
                onSignatureChange={setSignature}
                disabled={isSubmitting || isLoading}
              />
            </div>

            <div className="flex gap-2">
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={handleClearSignature}
                disabled={!signature || isSubmitting || isLoading}
              >
                Clear Signature
              </Button>
              {signature && (
                <div className="text-xs text-green-600 flex items-center gap-1">
                  ✓ Signature captured
                </div>
              )}
            </div>
          </div>

          {/* Error Message */}
          {error && (
            <div className="p-3 bg-red-50 border border-red-200 rounded-md">
              <p className="text-sm text-red-800 flex items-center gap-2">
                <AlertCircle className="h-4 w-4 flex-shrink-0" />
                {error}
              </p>
            </div>
          )}

          {/* Approval Summary */}
          <div className="p-3 bg-blue-50 border border-blue-200 rounded-md">
            <p className="text-sm text-blue-900">
              <span className="font-semibold">
                {isApproval
                  ? 'By approving, you confirm:'
                  : 'By rejecting, you confirm:'}
              </span>
            </p>
            <ul className="mt-2 space-y-1 text-sm text-blue-800 ml-4">
              <li>
                • Your digital signature certifies this action
              </li>
              <li>
                • This action creates an auditable record
              </li>
              {isApproval && (
                <li>
                  • You authorize this document to proceed
                </li>
              )}
              {isRejection && (
                <li>
                  • The document will be returned to the creator
                </li>
              )}
            </ul>
          </div>
        </div>

        <DialogFooter className="gap-2 sm:gap-0">
          <Button
            variant="outline"
            onClick={() => handleOpenChange(false)}
            disabled={isSubmitting || isLoading}
          >
            Cancel
          </Button>
          <Button
            variant={isApproval ? 'default' : 'destructive'}
            onClick={handleSubmit}
            disabled={isSubmitting || isLoading || !signature}
            className="gap-2"
          >
            {isSubmitting || isLoading ? (
              <>
                <div className="h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent" />
                {isApproval ? 'Approving...' : 'Rejecting...'}
              </>
            ) : (
              `${isApproval ? 'Confirm Approval' : 'Confirm Rejection'}`
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
