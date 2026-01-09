'use client'

import { Card } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Clock, CheckCircle, XCircle, Edit, Plus, Send, AlertCircle, User } from 'lucide-react'
import { ActionHistoryEntry, ApprovalRecord } from '@/types'
import { WorkflowDocument } from '@/types/workflow'
import { ApprovalActionPanel } from './approval-action-panel'
import { useApprovalPanelData } from '@/hooks/use-approval-history'

interface UnifiedHistoryPanelProps {
  requisitionId: string
  requisition: WorkflowDocument
  userRole: string
  actionHistory?: ActionHistoryEntry[]
  approvalChain?: ApprovalRecord[]
}

export function UnifiedHistoryPanel({
  requisitionId,
  requisition,
  userRole,
  actionHistory,
  approvalChain,
}: UnifiedHistoryPanelProps) {
  const {
    approvalHistory,
    availableApprovers,
    workflowStatus,
    isLoading,
    hasError,
    refetchAll,
  } = useApprovalPanelData(requisitionId, 'REQUISITION')

  const getActionIcon = (actionType: string) => {
    switch (actionType.toUpperCase()) {
      case 'APPROVE':
      case 'APPROVED':
        return <CheckCircle className="h-5 w-5 text-green-600" />
      case 'REJECT':
      case 'REJECTED':
        return <XCircle className="h-5 w-5 text-red-600" />
      case 'CREATE':
        return <Plus className="h-5 w-5 text-blue-600" />
      case 'UPDATE':
        return <Edit className="h-5 w-5 text-amber-600" />
      case 'SUBMIT':
        return <Send className="h-5 w-5 text-purple-600" />
      case 'REVERSE':
      case 'REVERSED':
        return <Edit className="h-5 w-5 text-amber-600" />
      default:
        return <Clock className="h-5 w-5 text-gray-600" />
    }
  }

  const getActionColor = (actionType: string) => {
    switch (actionType.toUpperCase()) {
      case 'APPROVE':
      case 'APPROVED':
        return 'bg-green-50 border-green-200'
      case 'REJECT':
      case 'REJECTED':
        return 'bg-red-50 border-red-200'
      case 'CREATE':
        return 'bg-blue-50 border-blue-200'
      case 'UPDATE':
        return 'bg-amber-50 border-amber-200'
      case 'SUBMIT':
        return 'bg-purple-50 border-purple-200'
      case 'REVERSE':
      case 'REVERSED':
        return 'bg-amber-50 border-amber-200'
      default:
        return 'bg-gray-50 border-gray-200'
    }
  }

  const getActionLabel = (actionType: string) => {
    switch (actionType.toUpperCase()) {
      case 'APPROVE':
      case 'APPROVED':
        return 'Approved'
      case 'REJECT':
      case 'REJECTED':
        return 'Rejected'
      case 'CREATE':
        return 'Created'
      case 'UPDATE':
        return 'Updated'
      case 'SUBMIT':
        return 'Submitted'
      case 'REVERSE':
      case 'REVERSED':
        return 'Reversed'
      case 'DELETE':
        return 'Deleted'
      case 'REVERT_TO_DRAFT':
        return 'Reverted to Draft'
      default:
        return actionType
    }
  }

  const handleApprovalComplete = () => {
    refetchAll()
  }

  // Combine and sort all history entries
  const sortedHistory = [...(actionHistory || [])].sort(
    (a, b) => new Date(b.performedAt || b.timestamp || 0).getTime() - new Date(a.performedAt || a.timestamp || 0).getTime()
  )

  // Combine approval history from both sources
  const combinedApprovalHistory = [
    ...(approvalHistory || []),
    ...(approvalChain || [])
  ].filter((item, index, self) => 
    index === self.findIndex(t => 
      (t.approverId && item.approverId && t.approverId === item.approverId) ||
      (t.stageNumber && item.stageNumber && t.stageNumber === item.stageNumber)
    )
  )

  if (hasError && !actionHistory?.length) {
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
      <Tabs defaultValue="timeline" className="w-full">
        <TabsList className="grid w-full grid-cols-3">
          <TabsTrigger value="timeline">
            Timeline
            {sortedHistory.length > 0 && (
              <Badge variant="secondary" className="ml-2 text-xs">
                {sortedHistory.length}
              </Badge>
            )}
          </TabsTrigger>
          <TabsTrigger value="chain">
            Approval Chain
            {combinedApprovalHistory.length > 0 && (
              <Badge variant="secondary" className="ml-2 text-xs">
                {combinedApprovalHistory.length}
              </Badge>
            )}
          </TabsTrigger>
          <TabsTrigger value="approvers">
            Approval Actions
            {availableApprovers.length > 0 && (
              <Badge variant="secondary" className="ml-2 text-xs">
                {availableApprovers.length}
              </Badge>
            )}
          </TabsTrigger>
        </TabsList>

        {/* Timeline Tab - All Actions Chronologically */}
        <TabsContent value="timeline" className="space-y-4 mt-4">
          {sortedHistory.length > 0 ? (
            <div className="space-y-3 max-h-96 overflow-y-auto">
              {sortedHistory.map((action) => (
                <div
                  key={action.id}
                  className={`p-4 rounded-lg border ${getActionColor(action.actionType || 'unknown')}`}
                >
                  <div className="flex items-start gap-3">
                    {getActionIcon(action.actionType || 'unknown')}
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2 flex-wrap">
                        <span className="font-semibold text-sm">
                          {action.performedByName}
                        </span>
                        <Badge variant="outline" className="text-xs">
                          {getActionLabel(action.actionType || 'unknown')}
                        </Badge>
                        {action.performedByRole && (
                          <Badge variant="secondary" className="text-xs">
                            {action.performedByRole}
                          </Badge>
                        )}
                      </div>
                      <p className="text-xs text-gray-600 mt-1">
                        {new Date(action.performedAt || action.timestamp || 0).toLocaleString()}
                      </p>

                      {/* Status transition */}
                      {action.previousStatus && action.newStatus && (
                        <div className="text-xs mt-2 text-gray-700">
                          Status: <span className="font-mono">{action.previousStatus}</span> →{' '}
                          <span className="font-mono">{action.newStatus}</span>
                        </div>
                      )}

                      {/* Stage info for approval actions */}
                      {action.stageNumber && action.stageName && (
                        <div className="text-xs mt-2 text-gray-700">
                          Stage {action.stageNumber}: <span className="font-semibold">{action.stageName}</span>
                        </div>
                      )}

                      {/* Comments */}
                      {action.comments && (
                        <p className="text-sm mt-2 text-gray-700 italic">
                          "{action.comments}"
                        </p>
                      )}

                      {/* Remarks (for rejections) */}
                      {action.remarks && (
                        <p className="text-sm mt-2 text-red-700 font-semibold">
                          Reason: "{action.remarks}"
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
              <p className="text-sm">No activity yet</p>
              <p className="text-xs text-gray-400 mt-1">
                Actions will appear here as the requisition progresses
              </p>
            </div>
          )}
        </TabsContent>

        {/* Approval Chain Tab - Required Signatories, Their Status, and Available Approvers */}
        <TabsContent value="chain" className="space-y-4 mt-4">
          {isLoading ? (
            <div className="text-center py-8">
              <div className="inline-block h-6 w-6 rounded-full border-2 border-blue-200 border-t-blue-600 animate-spin"></div>
              <p className="text-sm text-gray-500 mt-2">Loading approval chain...</p>
            </div>
          ) : (
            <>
              {/* Approval Chain Header */}
              <div className="text-xs text-gray-600 mb-4 p-3 bg-blue-50 rounded-lg border border-blue-200">
                <p className="font-semibold text-blue-900">Approval Chain</p>
                <p className="text-blue-700">Required signatories for this requisition in order of approval</p>
              </div>

              {/* Approval Chain Steps */}
              {combinedApprovalHistory.length > 0 ? (
                <div className="space-y-3 mb-6">
                  {combinedApprovalHistory.map((approval, index) => (
                    <div
                      key={approval.approverId || index}
                      className={`p-4 rounded-lg border-2 ${
                        approval.status === 'APPROVED' 
                          ? 'border-green-200 bg-green-50' 
                          : approval.status === 'REJECTED'
                            ? 'border-red-200 bg-red-50'
                            : approval.status === 'PENDING'
                              ? 'border-yellow-200 bg-yellow-50'
                              : 'border-gray-200 bg-gray-50'
                      }`}
                    >
                      <div className="flex items-start gap-3">
                        <div className="flex-shrink-0">
                          <div className={`w-8 h-8 rounded-full flex items-center justify-center text-xs font-bold ${
                            approval.status === 'APPROVED' 
                              ? 'bg-green-600 text-white' 
                              : approval.status === 'REJECTED'
                                ? 'bg-red-600 text-white'
                                : approval.status === 'PENDING'
                                  ? 'bg-yellow-600 text-white'
                                  : 'bg-gray-400 text-white'
                          }`}>
                            {approval.stageNumber || index + 1}
                          </div>
                        </div>
                        
                        <div className="flex-1 min-w-0">
                          <div className="flex items-center gap-2 flex-wrap mb-2">
                            <span className="font-semibold text-sm">
                              {approval.stageName || `Stage ${approval.stageNumber || index + 1}`}
                            </span>
                            <Badge
                              variant={
                                approval.status === 'APPROVED'
                                  ? 'default'
                                  : approval.status === 'REJECTED'
                                    ? 'destructive'
                                    : approval.status === 'PENDING'
                                      ? 'secondary'
                                      : 'outline'
                              }
                              className="text-xs"
                            >
                              {approval.status || 'PENDING'}
                            </Badge>
                          </div>

                          {approval.assignedRole && (
                            <p className="text-sm text-gray-700 mb-1">
                              <span className="font-medium">Required Role:</span> {approval.assignedRole}
                            </p>
                          )}

                          {(approval.approverName || approval.actionTakenBy) && (
                            <p className="text-sm text-gray-700 mb-1">
                              <span className="font-medium">Signatory:</span> {approval.approverName || approval.actionTakenBy}
                              {approval.actionTakenByRole && (
                                <span className="text-gray-500 ml-1">({approval.actionTakenByRole})</span>
                              )}
                            </p>
                          )}

                          {(approval.actionTakenAt || approval.approvedAt) && (
                            <p className="text-xs text-gray-600 mb-2">
                              <span className="font-medium">Date:</span> {new Date(approval.actionTakenAt || approval.approvedAt || '').toLocaleString()}
                            </p>
                          )}

                          {(approval.comments || approval.remarks) && (
                            <div className="mt-2 p-2 bg-white/50 rounded border">
                              <p className="text-sm text-gray-700">
                                <span className="font-medium">Comments:</span> "{approval.comments || approval.remarks}"
                              </p>
                            </div>
                          )}
                        </div>

                        <div className="flex-shrink-0">
                          {getActionIcon(approval.status || 'PENDING')}
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              ) : (
                <div className="text-center py-6 text-gray-500 mb-6">
                  <AlertCircle className="h-6 w-6 mx-auto mb-2 text-gray-400" />
                  <p className="text-sm">No approval chain configured</p>
                  <p className="text-xs text-gray-400 mt-1">
                    The approval workflow will appear here once configured
                  </p>
                </div>
              )}

              {/* Available Approvers Section */}
              {availableApprovers.length > 0 && (
                <div className="border-t pt-4">
                  <h4 className="font-semibold text-sm text-gray-700 mb-3 flex items-center gap-2">
                    <User className="h-4 w-4" />
                    Available Approvers ({availableApprovers.length})
                  </h4>
                  <p className="text-xs text-gray-600 mb-3">
                    People who have permission to approve this requisition at various stages
                  </p>
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-3 max-h-60 overflow-y-auto">
                    {availableApprovers.map((approver) => (
                      <div
                        key={approver.id}
                        className="p-3 border rounded-lg hover:bg-gray-50 transition-colors"
                      >
                        <div className="flex items-center gap-3">
                          <div className="w-8 h-8 rounded-full bg-blue-100 flex items-center justify-center flex-shrink-0">
                            <User className="h-4 w-4 text-blue-600" />
                          </div>
                          <div className="flex-1 min-w-0">
                            <p className="font-semibold text-sm truncate">
                              {approver.name || 'Unknown'}
                            </p>
                            <p className="text-xs text-gray-600 truncate">
                              {approver.role} {approver.department && `• ${approver.department}`}
                            </p>
                            {approver.email && (
                              <p className="text-xs text-gray-500 truncate">
                                📧 {approver.email}
                              </p>
                            )}
                          </div>
                          <Badge variant="outline" className="text-xs flex-shrink-0">
                            Can Approve
                          </Badge>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </>
          )}
        </TabsContent>

        {/* Approval Actions Tab - ONLY Interactive Approval Actions */}
        <TabsContent value="approvers" className="space-y-4 mt-4">
          {isLoading ? (
            <div className="text-center py-8">
              <div className="inline-block h-6 w-6 rounded-full border-2 border-blue-200 border-t-blue-600 animate-spin"></div>
              <p className="text-sm text-gray-500 mt-2">Loading approval data...</p>
            </div>
          ) : (
            <>
              {/* Approval Action Panel - Show if user can approve */}
              {(requisition.status === 'IN_REVIEW' || requisition.status === 'pending') && 
               workflowStatus?.canApprove ? (
                <div className="p-6 bg-blue-50 border border-blue-200 rounded-lg">
                  <h4 className="font-semibold text-lg text-blue-900 mb-2 flex items-center gap-2">
                    <CheckCircle className="h-5 w-5" />
                    Take Approval Action
                  </h4>
                  <p className="text-sm text-blue-700 mb-4">
                    You have permission to approve or reject this requisition at the current stage.
                  </p>
                  <ApprovalActionPanel
                    requisitionId={requisitionId}
                    onApprovalComplete={handleApprovalComplete}
                  />
                </div>
              ) : (
                <div className="text-center py-12 text-gray-500">
                  <AlertCircle className="h-12 w-12 mx-auto mb-4 text-gray-400" />
                  <h4 className="font-semibold text-lg mb-2">No Actions Available</h4>
                  <p className="text-sm text-gray-600 mb-2">
                    You don't have permission to approve this requisition at this stage.
                  </p>
                  <p className="text-xs text-gray-500">
                    Check the Approval Chain tab to see who can approve this requisition.
                  </p>
                </div>
              )}
            </>
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
    </Card>
  )
}