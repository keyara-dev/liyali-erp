'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'
import { SignatureCanvas } from '@/components/ui/signature-canvas'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { approveBudget, rejectBudget } from '@/app/_actions/budgets'
import { AlertCircle, CheckCircle2 } from 'lucide-react'

interface BudgetApprovalActionPanelProps {
  budgetId: string
  budgetStatus: string
  onApprovalComplete: () => void
}

export function BudgetApprovalActionPanel({
  budgetId,
  budgetStatus,
  onApprovalComplete,
}: BudgetApprovalActionPanelProps) {
  const [action, setAction] = useState<'approve' | 'reject' | null>(null)
  const [comments, setComments] = useState('')
  const [remarks, setRemarks] = useState('')
  const [signature, setSignature] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState<string | null>(null)

  // Only show for budgets that are in approval
  if (budgetStatus !== 'IN_APPROVAL' && budgetStatus !== 'SUBMITTED') {
    return null
  }

  const handleApprove = async () => {
    if (!signature) {
      setError('Signature is required to approve the budget')
      return
    }

    setError(null)
    setSuccess(null)
    setIsLoading(true)

    try {
      const result = await approveBudget({
        budgetId,
        approvingUserId: 'current-user-id', // In production, get from session
        approvingUserRole: 'FINANCE_OFFICER', // In production, get from session
        comments,
      })

      if (result.success) {
        setSuccess('Budget approved successfully')
        setComments('')
        setRemarks('')
        setSignature('')
        setAction(null)
        setTimeout(onApprovalComplete, 1500)
      } else {
        setError(result.message || 'Failed to approve budget')
      }
    } catch (err) {
      setError('An error occurred while approving the budget')
      console.error(err)
    } finally {
      setIsLoading(false)
    }
  }

  const handleReject = async () => {
    if (!remarks.trim()) {
      setError('Remarks are required for rejection')
      return
    }

    setError(null)
    setSuccess(null)
    setIsLoading(true)

    try {
      const result = await rejectBudget({
        budgetId,
        rejectingUserId: 'current-user-id', // In production, get from session
        rejectionReason: remarks,
        comments,
      })

      if (result.success) {
        setSuccess('Budget rejected successfully')
        setComments('')
        setRemarks('')
        setSignature('')
        setAction(null)
        setTimeout(onApprovalComplete, 1500)
      } else {
        setError(result.message || 'Failed to reject budget')
      }
    } catch (err) {
      setError('An error occurred while rejecting the budget')
      console.error(err)
    } finally {
      setIsLoading(false)
    }
  }

  if (action === null) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Approval Action</CardTitle>
          <CardDescription>
            Approve or reject this budget submission
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 gap-3">
            <Button
              onClick={() => {
                setAction('approve')
                setError(null)
                setSuccess(null)
              }}
              className="bg-green-600 hover:bg-green-700"
            >
              <CheckCircle2 className="h-4 w-4 mr-2" />
              Approve Budget
            </Button>
            <Button
              onClick={() => {
                setAction('reject')
                setError(null)
                setSuccess(null)
              }}
              variant="destructive"
            >
              <AlertCircle className="h-4 w-4 mr-2" />
              Reject Budget
            </Button>
          </div>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>
          {action === 'approve' ? 'Approve Budget' : 'Reject Budget'}
        </CardTitle>
        <CardDescription>
          {action === 'approve'
            ? 'Add a digital signature and optional comments to approve this budget'
            : 'Provide detailed remarks explaining why this budget is being rejected'}
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-6">
        {error && (
          <div className="flex items-start gap-3 p-4 rounded-lg bg-red-50 border border-red-200">
            <AlertCircle className="h-5 w-5 text-red-600 flex-shrink-0 mt-0.5" />
            <p className="text-sm text-red-800">{error}</p>
          </div>
        )}

        {success && (
          <div className="flex items-start gap-3 p-4 rounded-lg bg-green-50 border border-green-200">
            <CheckCircle2 className="h-5 w-5 text-green-600 flex-shrink-0 mt-0.5" />
            <p className="text-sm text-green-800">{success}</p>
          </div>
        )}

        {action === 'approve' ? (
          <>
            <div className="space-y-3">
              <Label htmlFor="comments">Comments (Optional)</Label>
              <Textarea
                id="comments"
                placeholder="Add any approval comments, conditions, or recommendations..."
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
          <div className="space-y-3">
            <Label htmlFor="remarks">
              Rejection Remarks *
            </Label>
            <Textarea
              id="remarks"
              placeholder="Required: Explain in detail why this budget is being rejected. This helps the requester understand the issues and resubmit appropriately."
              value={remarks}
              onChange={(e) => setRemarks(e.target.value)}
              rows={4}
              className="resize-none"
              disabled={isLoading}
            />
            <p className="text-xs text-muted-foreground">
              Detailed remarks are required for rejection to provide clear feedback
            </p>

            <div className="space-y-3 pt-2">
              <Label htmlFor="comments">Additional Comments (Optional)</Label>
              <Textarea
                id="comments"
                placeholder="Any additional context or suggestions for improvement..."
                value={comments}
                onChange={(e) => setComments(e.target.value)}
                rows={2}
                className="resize-none"
                disabled={isLoading}
              />
            </div>
          </div>
        )}

        <div className="flex gap-3 pt-4">
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
              setError(null)
              setSuccess(null)
            }}
            disabled={isLoading}
            className="flex-1"
          >
            Cancel
          </Button>
        </div>
      </CardContent>
    </Card>
  )
}
