'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { toast } from 'sonner'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { ArrowLeft, Send, AlertCircle } from 'lucide-react'
import { getDocument, submitDocument } from '@/app/_actions/workflow'
import { WorkflowDocument, RequisitionForm } from '@/types/workflow'
import { ApprovalHistoryPanel } from './approval-history-panel'
import { EditRequisitionPanel } from './edit-requisition-panel'
import { DocumentLinks } from '@/components/document-links'

interface RequisitionDetailClientProps {
  requisitionId: string
  userId: string
  userRole: string
}

const STATUS_COLORS: Record<string, { bg: string; text: string }> = {
  DRAFT: { bg: 'bg-gray-100', text: 'text-gray-800' },
  SUBMITTED: { bg: 'bg-blue-100', text: 'text-blue-800' },
  IN_APPROVAL: { bg: 'bg-yellow-100', text: 'text-yellow-800' },
  APPROVED: { bg: 'bg-green-100', text: 'text-green-800' },
  REJECTED: { bg: 'bg-red-100', text: 'text-red-800' },
}

export function RequisitionDetailClient({
  requisitionId,
  userId,
  userRole,
}: RequisitionDetailClientProps) {
  const router = useRouter()
  const [requisition, setRequisition] = useState<RequisitionForm | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [isSubmitting, setIsSubmitting] = useState(false)

  useEffect(() => {
    fetchRequisition()
  }, [requisitionId])

  const fetchRequisition = async () => {
    setIsLoading(true)
    try {
      const result = await getDocument(requisitionId)
      if (result.success) {
        setRequisition(result.data as RequisitionForm)
      } else {
        toast.error('Failed to load requisition')
      }
    } catch (error) {
      toast.error('Error loading requisition')
    } finally {
      setIsLoading(false)
    }
  }

  const handleSubmitForApproval = async () => {
    if (!requisition) return

    setIsSubmitting(true)
    try {
      const result = await submitDocument(requisitionId)
      if (result.success) {
        toast.success('Requisition submitted for approval')
        await fetchRequisition()
      } else {
        toast.error(result.message)
      }
    } catch (error) {
      toast.error('Failed to submit requisition')
    } finally {
      setIsSubmitting(false)
    }
  }

  const isCreator = requisition?.createdBy === userId
  const canEdit =
    isCreator && (requisition?.status === 'DRAFT' || requisition?.status === 'REJECTED')
  const canSubmit = canEdit

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="text-center">
          <div className="h-8 w-8 rounded-full border-4 border-blue-200 border-t-blue-600 animate-spin mx-auto mb-4"></div>
          <p className="text-gray-600">Loading requisition...</p>
        </div>
      </div>
    )
  }

  if (!requisition) {
    return (
      <div className="flex items-center justify-center py-12">
        <Card className="p-8 max-w-md text-center">
          <AlertCircle className="h-12 w-12 text-gray-400 mx-auto mb-4" />
          <h3 className="font-semibold text-lg mb-2">Requisition Not Found</h3>
          <p className="text-gray-600 mb-6">
            The requisition you're looking for doesn't exist.
          </p>
          <Button variant="outline" onClick={() => router.back()}>
            Go Back
          </Button>
        </Card>
      </div>
    )
  }

  const totalEstimatedCost = requisition.metadata?.items?.reduce(
    (sum, item) => sum + (item.estimatedCost || 0),
    0
  ) || 0

  const colors = STATUS_COLORS[requisition.status] || STATUS_COLORS['DRAFT']

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center gap-4">
        <Button
          variant="ghost"
          size="sm"
          onClick={() => router.back()}
          className="gap-2"
        >
          <ArrowLeft className="h-4 w-4" />
          Back
        </Button>
        <div className="flex-1">
          <div className="flex items-center gap-3 mb-2">
            <h1 className="text-3xl font-bold">{requisition.documentNumber}</h1>
            <Badge className={`${colors.bg} ${colors.text} border-0`}>
              {requisition.status}
            </Badge>
          </div>
          <p className="text-gray-600">
            Created on {new Date(requisition.createdAt).toLocaleString()}
          </p>
        </div>
        {canSubmit && (
          <Button
            onClick={handleSubmitForApproval}
            disabled={isSubmitting}
            className="gap-2"
          >
            <Send className="h-4 w-4" />
            {isSubmitting ? 'Submitting...' : 'Submit for Approval'}
          </Button>
        )}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Main Content */}
        <div className="lg:col-span-2 space-y-6">
          {/* Requisition Details */}
          <Card className="p-6">
            <h2 className="text-xl font-semibold mb-4">Requisition Details</h2>

            <div className="grid grid-cols-2 gap-6">
              <div>
                <label className="text-sm font-medium text-gray-600">
                  Department
                </label>
                <p className="text-lg font-semibold mt-1">
                  {requisition.metadata?.department}
                </p>
              </div>

              <div>
                <label className="text-sm font-medium text-gray-600">
                  Requested For
                </label>
                <p className="text-lg font-semibold mt-1">
                  {requisition.metadata?.requestedFor}
                </p>
              </div>

              <div className="col-span-2">
                <label className="text-sm font-medium text-gray-600">
                  Justification
                </label>
                <p className="text-base mt-1 whitespace-pre-wrap">
                  {requisition.metadata?.justification}
                </p>
              </div>

              <div>
                <label className="text-sm font-medium text-gray-600">
                  Budget Code
                </label>
                <p className="text-lg font-semibold mt-1">
                  {requisition.metadata?.budgetCode}
                </p>
              </div>

              <div>
                <label className="text-sm font-medium text-gray-600">
                  Current Approval Stage
                </label>
                <p className="text-lg font-semibold mt-1">
                  Stage {requisition.currentStage}
                </p>
              </div>
            </div>
          </Card>

          {/* Document Links */}
          {requisition.status === 'APPROVED' && (
            <DocumentLinks
              currentDocument={requisition as unknown as WorkflowDocument}
              linkedDocuments={{
                purchaseOrder: requisition.metadata?.purchaseOrderId
                  ? { id: requisition.metadata.purchaseOrderId, number: 'PO-2024-001' }
                  : undefined,
              }}
            />
          )}

          {/* Items Section */}
          <Card className="p-6">
            <h2 className="text-xl font-semibold mb-4">Requisition Items</h2>

            <div className="space-y-3">
              {requisition.metadata?.items?.map((item, index) => (
                <div
                  key={item.id}
                  className="border rounded-lg p-4 hover:bg-gray-50 transition"
                >
                  <div className="flex items-start justify-between mb-2">
                    <span className="font-semibold">Item {index + 1}</span>
                    <span className="text-sm text-gray-600">
                      Qty: {item.quantity}
                    </span>
                  </div>
                  <p className="text-gray-700 mb-2">{item.itemDescription}</p>
                  <div className="flex items-center justify-between text-sm">
                    <span className="text-gray-600">
                      Unit Cost: ZMW{' '}
                      {item.estimatedCost.toLocaleString('en-ZM', {
                        minimumFractionDigits: 2,
                        maximumFractionDigits: 2,
                      })}
                    </span>
                    <span className="font-semibold text-blue-600">
                      ZMW{' '}
                      {(item.quantity * item.estimatedCost).toLocaleString(
                        'en-ZM',
                        {
                          minimumFractionDigits: 2,
                          maximumFractionDigits: 2,
                        }
                      )}
                    </span>
                  </div>
                </div>
              ))}
            </div>

            {/* Total */}
            <div className="mt-4 pt-4 border-t flex items-center justify-between">
              <span className="font-semibold text-gray-700">
                Total Estimated Cost:
              </span>
              <span className="text-2xl font-bold text-blue-600">
                ZMW{' '}
                {totalEstimatedCost.toLocaleString('en-ZM', {
                  minimumFractionDigits: 2,
                  maximumFractionDigits: 2,
                })}
              </span>
            </div>
          </Card>

          {/* Edit Panel - Only for Creator in DRAFT/REJECTED status */}
          {canEdit && (
            <EditRequisitionPanel
              requisition={requisition}
              onRequisitionUpdated={fetchRequisition}
            />
          )}
        </div>

        {/* Sidebar - Approval History */}
        <div className="lg:col-span-1">
          <ApprovalHistoryPanel
            requisitionId={requisitionId}
            requisition={requisition}
            userRole={userRole}
          />
        </div>
      </div>
    </div>
  )
}
