'use client'

import { useState } from 'react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'
import { AlertCircle, Upload, Send, XCircle } from 'lucide-react'
import { approveDocument, rejectDocument } from '@/app/_actions/workflow'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'

interface ApprovalActionPanelProps {
  requisitionId: string
  onApprovalComplete: () => void
}

export function ApprovalActionPanel({
  requisitionId,
  onApprovalComplete,
}: ApprovalActionPanelProps) {
  const [action, setAction] = useState<'approve' | 'reject' | null>(null)
  const [comments, setComments] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const [showAttachmentDialog, setShowAttachmentDialog] = useState(false)

  const handleApprove = async () => {
    setIsLoading(true)
    try {
      const result = await approveDocument(requisitionId, comments)
      if (result.success) {
        toast.success('Requisition approved successfully')
        setComments('')
        setAction(null)
        onApprovalComplete()
      } else {
        toast.error(result.message)
      }
    } catch (error) {
      toast.error('Failed to approve requisition')
    } finally {
      setIsLoading(false)
    }
  }

  const handleReject = async () => {
    if (!comments.trim()) {
      toast.error('Please provide a reason for rejection')
      return
    }

    setIsLoading(true)
    try {
      const result = await rejectDocument(requisitionId, comments)
      if (result.success) {
        toast.success('Requisition rejected successfully')
        setComments('')
        setAction(null)
        onApprovalComplete()
      } else {
        toast.error(result.message)
      }
    } catch (error) {
      toast.error('Failed to reject requisition')
    } finally {
      setIsLoading(false)
    }
  }

  if (action === null) {
    return (
      <div className="space-y-3">
        <h3 className="font-semibold text-sm">Action Required</h3>
        <div className="grid grid-cols-2 gap-2">
          <Button
            onClick={() => setAction('approve')}
            className="bg-green-600 hover:bg-green-700 gap-2"
          >
            <Send className="h-4 w-4" />
            Approve
          </Button>
          <Button
            onClick={() => setAction('reject')}
            variant="destructive"
            className="gap-2"
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
            ? 'Approve Requisition'
            : 'Reject Requisition'}
        </h3>
        <p className="text-sm text-gray-600 mb-4">
          {action === 'approve'
            ? 'Add any approval comments or recommendations'
            : 'Please provide a reason for rejecting this requisition'}
        </p>
      </div>

      <div className="space-y-2">
        <Label htmlFor="comments">
          Comments {action === 'reject' && '*'}
        </Label>
        <Textarea
          id="comments"
          placeholder={
            action === 'approve'
              ? 'Optional approval comments...'
              : 'Required reason for rejection...'
          }
          value={comments}
          onChange={(e) => setComments(e.target.value)}
          rows={3}
          className="resize-none"
        />
      </div>

      <Button
        variant="outline"
        size="sm"
        onClick={() => setShowAttachmentDialog(true)}
        className="gap-2 w-full text-gray-700"
      >
        <Upload className="h-4 w-4" />
        Add Supporting Documents
      </Button>

      <div className="flex gap-2">
        <Button
          onClick={action === 'approve' ? handleApprove : handleReject}
          disabled={isLoading || (action === 'reject' && !comments.trim())}
          className={
            action === 'approve'
              ? 'bg-green-600 hover:bg-green-700 flex-1'
              : 'bg-red-600 hover:bg-red-700 flex-1'
          }
        >
          {isLoading
            ? 'Processing...'
            : action === 'approve'
            ? 'Confirm Approval'
            : 'Confirm Rejection'}
        </Button>
        <Button
          variant="outline"
          onClick={() => setAction(null)}
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
