'use client'

import { Card } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { AlertCircle, Clock, CheckCircle, XCircle } from 'lucide-react'
import { WorkflowDocument } from '@/types/workflow'
import { ApprovalActionPanel } from './approval-action-panel'
import { useApprovalPanelData } from '@/hooks/use-approval-history'

interface ApprovalHistoryPanelProps {
  requisitionId: string
  requisition: WorkflowDocument
  userRole: string
}

export function ApprovalHistoryPanel({
  requisitionId,
  requisition,
  userRole,
}: ApprovalHistoryPanelProps) {
  const {
    approvalHistory,
    availableApprovers,
    workflowStatus,
    isLoading,
    hasError,
    refetchAll,
  } = useApprovalPanelData(requisitionId, 'REQUISITION')

  const getActionIcon = (action: string) => {
    switch (action.toUpperCase()) {
      case 'APPROVED':
        return <CheckCircle className="h-5 w-5 text-green-600" />
      case 'REJECTED':
        return <XCircle className="h-5 w-5 text-red-600" />
      default:
        return <Clock className="h-5 w-5 text-gray-600" />
    }
  }

  const getActionColor = (action: string) => {
    switch (action.toUpperCase()) {
      case 'APPROVED':
        return 'bg-green-50'
      case 'REJECTED':
        return 'bg-red-50'
      default:
        return 'bg-gray-50'
    }
  }

  const handleApprovalComplete = () => {
    // Refetch all data after approval action
    refetchAll()
  }

  if (hasError) {
    return (
      <Card className="p-6">
        <div className="text-center py-8 text-red-500">
          <AlertCircle className="h-8 w-8 mx-auto mb-2" />
          <p className="text-sm">Failed to load approval data</p>
          <button
            onClick={refetchAll}
            className="mt-2 text-xs text-blue-600 hover:underline"
          >
            Try again
          </button>
        </div>
      </Card>
    )
  }

  return (
    <Card className="p-6">
      <Tabs defaultValue="history" className="w-full">
        <TabsList className="grid w-full grid-cols-2">
          <TabsTrigger value="history">
            Approval Log
            {approvalHistory.length > 0 && (
              <Badge variant="secondary" className="ml-2 text-xs">
                {approvalHistory.length}
              </Badge>
            )}
          </TabsTrigger>
          <TabsTrigger value="approvers">
            Approvers
            {availableApprovers.length > 0 && (
              <Badge variant="secondary" className="ml-2 text-xs">
                {availableApprovers.length}
              </Badge>
            )}
          </TabsTrigger>
        </TabsList>

        <TabsContent value="history" className="space-y-4 mt-4">
          {isLoading ? (
            <div className="text-center py-8">
              <div className="inline-block h-6 w-6 rounded-full border-2 border-blue-200 border-t-blue-600 animate-spin"></div>
              <p className="text-sm text-gray-500 mt-2">Loading approval history...</p>
            </div>
          ) : approvalHistory.length > 0 ? (
            <div className="space-y-3 max-h-96 overflow-y-auto">
              {approvalHistory.map((log, index) => (
                <div
                  key={log.id || index}
                  className={`p-3 rounded-lg ${getActionColor(log.action || '')}`}
                >
                  <div className="flex items-start gap-3">
                    {getActionIcon(log.action || '')}
                    <div className="flex-1">
                      <div className="flex items-center gap-2">
                        <span className="font-semibold text-sm">
                          {(log as any).approverName || (log as any).performedByName || 'Unknown'}
                        </span>
                        <Badge variant="outline" className="text-xs">
                          {log.action || 'PENDING'}
                        </Badge>
                      </div>
                      <p className="text-xs text-gray-600 mt-1">
                        {log.timestamp 
                          ? new Date(log.timestamp).toLocaleString()
                          : (log as any).performedAt 
                            ? new Date((log as any).performedAt).toLocaleString()
                            : 'No date'
                        }
                      </p>
                      {(log.comments || (log as any).remarks) && (
                        <p className="text-sm mt-2 text-gray-700">
                          "{log.comments || (log as any).remarks}"
                        </p>
                      )}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="text-center py-8 text-gray-500">
              <Clock className="h-8 w-8 mx-auto mb-2 text-gray-400" />
              <p className="text-sm">No approval history yet</p>
              <p className="text-xs text-gray-400 mt-1">
                Actions will appear here once the approval process begins
              </p>
            </div>
          )}
        </TabsContent>

        <TabsContent value="approvers" className="space-y-4 mt-4">
          {isLoading ? (
            <div className="text-center py-8">
              <div className="inline-block h-6 w-6 rounded-full border-2 border-blue-200 border-t-blue-600 animate-spin"></div>
              <p className="text-sm text-gray-500 mt-2">Loading approvers...</p>
            </div>
          ) : availableApprovers.length > 0 ? (
            <div className="space-y-2 max-h-96 overflow-y-auto">
              {availableApprovers.map((approver) => (
                <div
                  key={approver.id}
                  className="p-3 border rounded-lg hover:bg-gray-50 transition-colors"
                >
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="font-semibold text-sm">
                        {approver.name || 'Unknown'}
                      </p>
                      <p className="text-xs text-gray-600">
                        {approver.role} {approver.department && `• ${approver.department}`}
                      </p>
                      {approver.email && (
                        <p className="text-xs text-gray-500 mt-1">
                          {approver.email}
                        </p>
                      )}
                    </div>
                    <Badge variant="outline" className="text-xs">
                      Available
                    </Badge>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="text-center py-8 text-gray-500">
              <AlertCircle className="h-8 w-8 mx-auto mb-2 text-gray-400" />
              <p className="text-sm">No approvers available</p>
              <p className="text-xs text-gray-400 mt-1">
                Check workflow configuration or contact administrator
              </p>
            </div>
          )}
        </TabsContent>
      </Tabs>

      {/* Workflow Status Summary */}
      {workflowStatus && (
        <div className="mt-6 pt-6 border-t">
          <div className="flex items-center justify-between text-sm">
            <span className="text-gray-600">
              Stage {workflowStatus.currentStage} of {workflowStatus.totalStages}
            </span>
            <Badge 
              variant={workflowStatus.status === 'APPROVED' ? 'default' : 'secondary'}
              className="text-xs"
            >
              {workflowStatus.status}
            </Badge>
          </div>
          {workflowStatus.nextApprover && (
            <p className="text-xs text-gray-500 mt-1">
              Next approver: {workflowStatus.nextApprover}
            </p>
          )}
        </div>
      )}

      {/* Approval Action Panel */}
      {(requisition.status === 'IN_REVIEW' || requisition.status === 'pending') && 
       workflowStatus?.canApprove && (
        <div className="mt-6 pt-6 border-t">
          <ApprovalActionPanel
            requisitionId={requisitionId}
            onApprovalComplete={handleApprovalComplete}
          />
        </div>
      )}
    </Card>
  )
}
