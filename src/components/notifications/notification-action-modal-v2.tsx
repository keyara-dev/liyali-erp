'use client'

import { useState, useRef } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Loader2, AlertCircle, CheckCircle2 } from 'lucide-react'
import { Notification } from '@/types'
import { approveTaskSchema, rejectTaskSchema } from '@/lib/validation-schemas'
import { notify } from '@/lib/utils'

interface SignatureCanvasProps {
  onSignatureChange: (signature: string) => void
  isRequired?: boolean
  error?: string
}

const SignatureCanvas = ({
  onSignatureChange,
  isRequired = true,
  error,
}: SignatureCanvasProps) => {
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const [isDrawing, setIsDrawing] = useState(false)
  const [hasSignature, setHasSignature] = useState(false)

  const startDrawing = (e: React.MouseEvent<HTMLCanvasElement>) => {
    const canvas = canvasRef.current
    if (!canvas) return

    const rect = canvas.getBoundingClientRect()
    const ctx = canvas.getContext('2d')
    if (!ctx) return

    ctx.beginPath()
    ctx.moveTo(e.clientX - rect.left, e.clientY - rect.top)
    setIsDrawing(true)
  }

  const draw = (e: React.MouseEvent<HTMLCanvasElement>) => {
    if (!isDrawing) return

    const canvas = canvasRef.current
    if (!canvas) return

    const rect = canvas.getBoundingClientRect()
    const ctx = canvas.getContext('2d')
    if (!ctx) return

    ctx.lineTo(e.clientX - rect.left, e.clientY - rect.top)
    ctx.stroke()
  }

  const stopDrawing = () => {
    if (!isDrawing) return

    const canvas = canvasRef.current
    if (!canvas) return

    const ctx = canvas.getContext('2d')
    if (ctx) {
      ctx.closePath()
    }

    setIsDrawing(false)
    setHasSignature(true)

    const signature = canvas.toDataURL('image/png')
    onSignatureChange(signature)
  }

  const clearSignature = () => {
    const canvas = canvasRef.current
    if (!canvas) return

    const ctx = canvas.getContext('2d')
    if (ctx) {
      ctx.clearRect(0, 0, canvas.width, canvas.height)
    }

    setHasSignature(false)
    onSignatureChange('')
  }

  return (
    <div className="space-y-3">
      <Label>
        Digital Signature {isRequired && <span className="text-destructive">*</span>}
      </Label>
      <div
        className={`border rounded-lg bg-white dark:bg-slate-900 overflow-hidden ${
          error ? 'border-destructive' : ''
        }`}
      >
        <canvas
          ref={canvasRef}
          width={400}
          height={150}
          onMouseDown={startDrawing}
          onMouseMove={draw}
          onMouseUp={stopDrawing}
          onMouseLeave={stopDrawing}
          className="w-full cursor-crosshair bg-white dark:bg-slate-900"
        />
      </div>
      <div className="flex gap-2 items-center justify-between">
        <div className="flex gap-2">
          <Button type="button" variant="outline" size="sm" onClick={clearSignature}>
            Clear
          </Button>
          <span
            className={`text-xs self-center ${
              hasSignature ? 'text-green-600 dark:text-green-400' : 'text-muted-foreground'
            }`}
          >
            {hasSignature ? '✓ Signature captured' : 'Draw your signature above'}
          </span>
        </div>
      </div>
      {error && <p className="text-xs text-destructive">{error}</p>}
    </div>
  )
}

interface NotificationActionModalProps {
  notification: Notification | null
  isOpen: boolean
  onOpenChange: (open: boolean) => void
  onApprove?: (signature: string, remarks: string) => Promise<void>
  onReject?: (remarks: string) => Promise<void>
  actionType?: 'approve' | 'reject' | 'both'
}

export function NotificationActionModal({
  notification,
  isOpen,
  onOpenChange,
  onApprove,
  onReject,
  actionType = 'both',
}: NotificationActionModalProps) {
  const [action, setAction] = useState<'approve' | 'reject' | null>(null)
  const [serverError, setServerError] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [signature, setSignature] = useState('')

  // Approval form with validation
  const approveForm = useForm({
    resolver: zodResolver(approveTaskSchema),
    mode: 'onBlur',
    defaultValues: {
      signature: '',
      comments: '',
    },
  })

  // Rejection form with validation
  const rejectForm = useForm({
    resolver: zodResolver(rejectTaskSchema),
    mode: 'onBlur',
    defaultValues: {
      signature: '',
      remarks: '',
    },
  })

  if (!notification) return null

  const handleApprove = async () => {
    setServerError(null)

    // Update form signature value
    approveForm.setValue('signature', signature)

    const isValid = await approveForm.trigger()
    if (!isValid) return

    if (!onApprove) {
      setServerError('Approve action not available')
      return
    }

    setIsSubmitting(true)
    try {
      const data = approveForm.getValues()
      await onApprove(data.signature, data.comments || '')
      approveForm.reset()
      setAction(null)
      setSignature('')
      onOpenChange(false)
      notify({ title: 'Task approved successfully!', type: 'success' })
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Failed to approve task'
      setServerError(errorMsg)
      notify({ title: errorMsg, type: 'error' })
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleReject = async () => {
    setServerError(null)

    // Update form signature value
    rejectForm.setValue('signature', signature)

    const isValid = await rejectForm.trigger()
    if (!isValid) return

    if (!onReject) {
      setServerError('Reject action not available')
      return
    }

    setIsSubmitting(true)
    try {
      const data = rejectForm.getValues()
      await onReject(data.remarks)
      rejectForm.reset()
      setAction(null)
      setSignature('')
      onOpenChange(false)
      notify({ title: 'Task rejected successfully!', type: 'success' })
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Failed to reject task'
      setServerError(errorMsg)
      notify({ title: errorMsg, type: 'error' })
    } finally {
      setIsSubmitting(false)
    }
  }

  const getTitle = () => {
    if (action === 'approve') return 'Approve Submission'
    if (action === 'reject') return 'Reject Submission'
    return `Review ${notification.entityType} #${notification.entityNumber}`
  }

  return (
    <Dialog open={isOpen} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>{getTitle()}</DialogTitle>
          <DialogDescription>{notification.message}</DialogDescription>
        </DialogHeader>

        {!action ? (
          // Preview Mode
          <div className="space-y-4 py-4">
            <div className="rounded-lg border bg-muted/50 p-4">
              <div className="grid grid-cols-2 gap-4 text-sm">
                <div>
                  <div className="font-semibold text-muted-foreground">Type</div>
                  <div>{notification.entityType}</div>
                </div>
                <div>
                  <div className="font-semibold text-muted-foreground">Number</div>
                  <div className="font-mono">{notification.entityNumber}</div>
                </div>
              </div>
            </div>

            {serverError && (
              <Alert variant="destructive">
                <AlertCircle className="h-4 w-4" />
                <AlertDescription>{serverError}</AlertDescription>
              </Alert>
            )}

            <div className="flex gap-3">
              {(actionType === 'approve' || actionType === 'both') && onApprove && (
                <Button onClick={() => setAction('approve')} className="flex-1">
                  <CheckCircle2 className="mr-2 h-4 w-4" />
                  Approve
                </Button>
              )}
              {(actionType === 'reject' || actionType === 'both') && onReject && (
                <Button onClick={() => setAction('reject')} variant="outline" className="flex-1">
                  Reject
                </Button>
              )}
            </div>
          </div>
        ) : action === 'approve' ? (
          // Approval Form
          <div className="space-y-4 py-4">
            {serverError && (
              <Alert variant="destructive">
                <AlertCircle className="h-4 w-4" />
                <AlertDescription>{serverError}</AlertDescription>
              </Alert>
            )}

            <SignatureCanvas
              onSignatureChange={setSignature}
              error={approveForm.formState.errors.signature?.message}
            />

            <div className="space-y-2">
              <Label htmlFor="comments">Comments (Optional)</Label>
              <Textarea
                id="comments"
                placeholder="Add any remarks or comments..."
                {...approveForm.register('comments')}
                className="min-h-24"
                disabled={isSubmitting}
              />
              {approveForm.formState.errors.comments && (
                <p className="text-xs text-destructive">
                  {approveForm.formState.errors.comments.message}
                </p>
              )}
            </div>

            <DialogFooter className="gap-2">
              <Button
                variant="outline"
                onClick={() => setAction(null)}
                disabled={isSubmitting}
              >
                Back
              </Button>
              <Button onClick={handleApprove} disabled={isSubmitting}>
                {isSubmitting ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    Submitting...
                  </>
                ) : (
                  'Submit Approval'
                )}
              </Button>
            </DialogFooter>
          </div>
        ) : (
          // Rejection Form
          <div className="space-y-4 py-4">
            {serverError && (
              <Alert variant="destructive">
                <AlertCircle className="h-4 w-4" />
                <AlertDescription>{serverError}</AlertDescription>
              </Alert>
            )}

            <SignatureCanvas
              onSignatureChange={setSignature}
              error={rejectForm.formState.errors.signature?.message}
            />

            <div className="space-y-2">
              <Label htmlFor="remarks">Rejection Reason *</Label>
              <Textarea
                id="remarks"
                placeholder="Please explain why you are rejecting this request..."
                {...rejectForm.register('remarks')}
                className="min-h-24"
                disabled={isSubmitting}
              />
              {rejectForm.formState.errors.remarks && (
                <p className="text-xs text-destructive">
                  {rejectForm.formState.errors.remarks.message}
                </p>
              )}
            </div>

            <DialogFooter className="gap-2">
              <Button
                variant="outline"
                onClick={() => setAction(null)}
                disabled={isSubmitting}
              >
                Back
              </Button>
              <Button onClick={handleReject} disabled={isSubmitting} variant="destructive">
                {isSubmitting ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    Submitting...
                  </>
                ) : (
                  'Submit Rejection'
                )}
              </Button>
            </DialogFooter>
          </div>
        )}
      </DialogContent>
    </Dialog>
  )
}
