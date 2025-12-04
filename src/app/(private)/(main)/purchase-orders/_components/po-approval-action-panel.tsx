'use client'

import { useState } from 'react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'
import { SignatureCanvas } from '@/components/ui/signature-canvas'
import { AlertCircle, Upload, Send, XCircle, Loader2 } from 'lucide-react'
import { useApprovePurchaseOrder, useRejectPurchaseOrder } from '@/hooks/use-purchase-order-queries'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'

interface POApprovalActionPanelProps {
  poId: string
  onApprovalComplete: () => void
  userId?: string
  userName?: string
  userRole?: string
}

export function POApprovalActionPanel({
  poId,
  onApprovalComplete,
  userId = 'user-' + Math.random().toString(36).substr(2, 9),
  userName = 'Approver',
  userRole = 'APPROVER',
}: POApprovalActionPanelProps) {
  const [action, setAction] = useState<'approve' | 'reject' | null>(null)
  const [comments, setComments] = useState('')
  const [remarks, setRemarks] = useState('')
  const [signature, setSignature] = useState('')
  const [showAttachmentDialog, setShowAttachmentDialog] = useState(false)

  const approveMutation = useApprovePurchaseOrder(poId, () => {
    setComments('')
    setRemarks('')
    setSignature('')
    setAction(null)
    onApprovalComplete()
  })

  const rejectMutation = useRejectPurchaseOrder(poId, () => {
    setComments('')
    setRemarks('')
    setSignature('')
    setAction(null)
    onApprovalComplete()
  })

  const handleApprove = async () => {
    if (!signature) {
      toast.error('Signature is required to approve')
      return
    }

    try {
      await approveMutation.mutateAsync({
        approvingUserId: userId,
        approvingUserName: userName,
        approvingUserRole: userRole,
        signature,
        comments: comments || undefined,
      })
    } catch (error) {
      console.error('Approval error:', error)
    }
  }

  const handleReject = async () => {
    if (!remarks.trim()) {
      toast.error('Remarks are required for rejection')
      return
    }

    try {
      await rejectMutation.mutateAsync({
        rejectingUserId: userId,
        rejectingUserName: userName,
        rejectingUserRole: userRole,
        remarks,
        signature: signature || '',
        comments: comments || undefined,
      })
    } catch (error) {
      console.error('Rejection error:', error)
    }
  }

  const isLoading = approveMutation.isPending || rejectMutation.isPending

  if (action === null) {
    return (
      <div className="space-y-3">
        <h3 className="font-semibold text-sm">Action Required</h3>
        <div className="grid grid-cols-2 gap-2">
          <Button
            onClick={() => setAction('approve')}
            className="bg-green-600 hover:bg-green-700 gap-2"
            disabled={isLoading}
          >
            <Send className="h-4 w-4" />
            Approve
          </Button>
          <Button
            onClick={() => setAction('reject')}
            variant="destructive"
            className="gap-2"
            disabled={isLoading}
          >
            <XCircle className="h-4 w-4" />
            Reject
          </Button>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-4 p-4 bg-blue-50 rounded-lg border border-blue-200">
      <div>
        <h3 className="font-semibold mb-2">
          {action === 'approve'
            ? 'Approve Purchase Order'
            : 'Reject Purchase Order'}
        </h3>
        <p className="text-sm text-gray-600 mb-4">
          {action === 'approve'
            ? 'Add a signature and optional comments to approve this purchase order'
            : 'Provide remarks explaining the rejection reason for this purchase order'}
        </p>
      </div>

      {action === 'approve' ? (
        <>
          <div className="space-y-2">
            <Label htmlFor="comments">Comments (Optional)</Label>
            <Textarea
              id="comments"
              placeholder="Add any approval comments or recommendations..."
              value={comments}
              onChange={(e) => setComments(e.target.value)}
              rows={3}
              className="resize-none"
              disabled={isLoading}
            />
          </div>

          <SignatureCanvas
            onSignatureChange={setSignature}
            disabled={isLoading}
          />
        </>
      ) : (
        <div className="space-y-2">
          <Label htmlFor="remarks">
            Remarks *
          </Label>
          <Textarea
            id="remarks"
            placeholder="Required: Please explain why this purchase order is being rejected..."
            value={remarks}
            onChange={(e) => setRemarks(e.target.value)}
            rows={4}
            className="resize-none"
            disabled={isLoading}
          />
          <p className="text-xs text-muted-foreground">
            Detailed remarks are required for rejection to help the requester understand the issues
          </p>
        </div>
      )}

      <Button
        variant="outline"
        size="sm"
        onClick={() => setShowAttachmentDialog(true)}
        className="gap-2 w-full text-gray-700"
        disabled={isLoading}
      >
        <Upload className="h-4 w-4" />
        Add Supporting Documents
      </Button>

      <div className="flex gap-2">
        <Button
          onClick={action === 'approve' ? handleApprove : handleReject}
          disabled={
            isLoading ||
            (action === 'reject' && !remarks.trim()) ||
            (action === 'approve' && !signature)
          }
          className={
            action === 'approve'
              ? 'bg-green-600 hover:bg-green-700 flex-1'
              : 'bg-red-600 hover:bg-red-700 flex-1'
          }
        >
          {isLoading && <Loader2 className="h-4 w-4 mr-2 animate-spin" />}
          {isLoading
            ? 'Processing...'
            : action === 'approve'
            ? 'Confirm Approval'
            : 'Confirm Rejection'}
        </Button>
        <Button
          variant="outline"
          onClick={() => {
            setAction(null)
            setComments('')
            setRemarks('')
            setSignature('')
          }}
          disabled={isLoading}
        >
          Cancel
        </Button>
      </div>

      {/* Attachment Dialog */}
      <Dialog open={showAttachmentDialog} onOpenChange={setShowAttachmentDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Add Supporting Documents</DialogTitle>
            <DialogDescription>
              Upload documents to support your approval or rejection decision
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="border-2 border-dashed rounded-lg p-6 text-center">
              <Upload className="h-8 w-8 text-gray-400 mx-auto mb-2" />
              <p className="text-sm text-gray-600">
                Click or drag files here to upload
              </p>
              <p className="text-xs text-gray-500 mt-1">
                PDF, DOC, XLS up to 10MB
              </p>
            </div>
            <Button onClick={() => setShowAttachmentDialog(false)} className="w-full">
              Continue
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  )
}
