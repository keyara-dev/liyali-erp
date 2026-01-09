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

        {/* Approval Chain Tab - Enhanced Workflow Stage Tracker */}
        <TabsContent value="chain" className="space-y-4 mt-4">
          {isLoading ? (
            <div className="text-center py-8">
              <div className="inline-block h-6 w-6 rounded-full border-2 border-blue-200 border-t-blue-600 animate-spin"></div>
              <p className="text-sm text-gray-500 mt-2">Loading approval chain...</p>
            </div>
          ) : (
            <>
              {/* Workflow Progress Header */}
              <div className="text-xs text-gray-600 mb-4 p-3 bg-blue-50 rounded-lg border border-blue-200">
                <p className="font-semibold text-blue-900">Workflow Progress Tracker</p>
                <p className="text-blue-700">Track each approval stage and see who has approved or is required to approve</p>
                {workflowStatus && (
                  <div className="mt-2 flex items-center gap-4">
                    <span className="text-blue-800 font-medium">
                      Stage {workflowStatus.currentStage} of {workflowStatus.totalStages}
                    </span>
                    <Badge 
                      variant={workflowStatus.status === 'completed' ? 'default' : 'secondary'}
                      className="text-xs"
                    >
                      {workflowStatus.status?.toUpperCase()}
                    </Badge>
                  </div>
                )}
              </div>

              {/* Enhanced Workflow Stage Progress */}
              {workflowStatus?.stageProgress && workflowStatus.stageProgress.length > 0 ? (
                <div className="space-y-3 mb-6">
                  {workflowStatus.stageProgress.map((stage, index) => (
                    <div
                      key={stage.stageNumber || index}
                      className={`p-4 rounded-lg border-2 transition-all ${
                        stage.status === 'approved' 
                          ? 'border-green-200 bg-green-50 shadow-sm' 
                          : stage.status === 'rejected'
                            ? 'border-red-200 bg-red-50 shadow-sm'
                            : stage.isCurrentStage
                              ? 'border-blue-300 bg-blue-50 shadow-md ring-2 ring-blue-100'
                              : stage.status === 'completed'
                                ? 'border-gray-300 bg-gray-50'
                                : 'border-gray-200 bg-gray-50'
                      }`}
                    >
                      <div className="flex items-start gap-3">
                        {/* Stage Number Circle */}
                        <div className="flex-shrink-0">
                          <div className={`w-10 h-10 rounded-full flex items-center justify-center text-sm font-bold ${
                            stage.status === 'approved' 
                              ? 'bg-green-600 text-white' 
                              : stage.status === 'rejected'
                                ? 'bg-red-600 text-white'
                                : stage.isCurrentStage
                                  ? 'bg-blue-600 text-white ring-2 ring-blue-300'
                                  : stage.status === 'completed'
                                    ? 'bg-gray-500 text-white'
                                    : 'bg-gray-300 text-gray-600'
                          }`}>
                            {stage.stageNumber || index + 1}
                          </div>
                        </div>
                        
                        <div className="flex-1 min-w-0">
                          {/* Stage Header */}
                          <div className="flex items-center gap-2 flex-wrap mb-2">
                            <span className="font-semibold text-base">
                              {stage.stageName || `Stage ${stage.stageNumber || index + 1}`}
                            </span>
                            <Badge
                              variant={
                                stage.status === 'approved'
                                  ? 'default'
                                  : stage.status === 'rejected'
                                    ? 'destructive'
                                    : stage.isCurrentStage
                                      ? 'secondary'
                                      : 'outline'
                              }
                              className="text-xs"
                            >
                              {stage.status === 'approved' ? 'APPROVED' : 
                               stage.status === 'rejected' ? 'REJECTED' :
                               stage.isCurrentStage ? 'CURRENT STAGE' :
                               stage.status === 'completed' ? 'COMPLETED' : 'PENDING'}
                            </Badge>
                            {stage.isCurrentStage && (
                              <Badge variant="outline" className="text-xs bg-blue-100 text-blue-800">
                                ⏳ Awaiting Action
                              </Badge>
                            )}
                          </div>

                          {/* Required Role */}
                          <div className="grid grid-cols-1 md:grid-cols-2 gap-3 mb-3">
                            <div>
                              <p className="text-sm text-gray-700 mb-1">
                                <span className="font-medium">Required Role:</span> 
                                <span className="ml-1 px-2 py-1 bg-gray-100 rounded text-xs font-mono">
                                  {stage.requiredRole}
                                </span>
                              </p>
                            </div>

                            {/* Approver Info */}
                            {(stage.approverName || stage.approverId) && (
                              <div>
                                <p className="text-sm text-gray-700 mb-1">
                                  <span className="font-medium">Approved By:</span> 
                                  <span className="ml-1 text-green-700 font-semibold">
                                    {stage.approverName || 'Unknown User'}
                                  </span>
                                  {stage.approverRole && (
                                    <span className="text-gray-500 ml-1">({stage.approverRole})</span>
                                  )}
                                </p>
                              </div>
                            )}
                          </div>

                          {/* Completion Date */}
                          {stage.completedAt && (
                            <p className="text-xs text-gray-600 mb-2">
                              <span className="font-medium">Completed:</span> 
                              <span className="ml-1">
                                {new Date(stage.completedAt).toLocaleString()}
                              </span>
                            </p>
                          )}

                          {/* Comments */}
                          {stage.comments && (
                            <div className="mt-2 p-3 bg-white/70 rounded border border-gray-200">
                              <p className="text-sm text-gray-700">
                                <span className="font-medium">Comments:</span> 
                                <span className="ml-1 italic">"{stage.comments}"</span>
                              </p>
                            </div>
                          )}

                          {/* Current Stage Instructions */}
                          {stage.isCurrentStage && stage.status === 'pending' && (
                            <div className="mt-3 p-3 bg-blue-100 rounded border border-blue-200">
                              <p className="text-sm text-blue-800">
                                <span className="font-medium">⚡ Next Action Required:</span> 
                                <span className="ml-1">
                                  This stage requires approval from a user with the <strong>{stage.requiredRole}</strong> role.
                                </span>
                              </p>
                            </div>
                          )}
                        </div>

                        {/* Status Icon */}
                        <div className="flex-shrink-0">
                          {stage.status === 'approved' ? (
                            <CheckCircle className="h-6 w-6 text-green-600" />
                          ) : stage.status === 'rejected' ? (
                            <XCircle className="h-6 w-6 text-red-600" />
                          ) : stage.isCurrentStage ? (
                            <Clock className="h-6 w-6 text-blue-600 animate-pulse" />
                          ) : (
                            <Clock className="h-6 w-6 text-gray-400" />
                          )}
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              ) : (
                // Fallback to legacy approval history if no workflow stages
                <>
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
                </>
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

      {/* Enhanced Workflow Status Summary */}
      {workflowStatus && (
        <div className="mt-6 pt-6 border-t">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4 text-sm">
            {/* Progress Indicator */}
            <div className="flex items-center gap-2">
              <span className="text-gray-600">Progress:</span>
              <div className="flex-1 bg-gray-200 rounded-full h-2">
                <div 
                  className={`h-2 rounded-full transition-all duration-300 ${
                    workflowStatus.status === 'completed' ? 'bg-green-500' :
                    workflowStatus.status === 'rejected' ? 'bg-red-500' : 'bg-blue-500'
                  }`}
                  style={{ 
                    width: `${Math.max(10, (workflowStatus.currentStage / Math.max(1, workflowStatus.totalStages)) * 100)}%` 
                  }}
                />
              </div>
              <span className="text-xs text-gray-500">
                {workflowStatus.currentStage}/{workflowStatus.totalStages}
              </span>
            </div>

            {/* Status Badge */}
            <div className="flex items-center justify-center">
              <Badge 
                variant={
                  workflowStatus.status === 'completed' ? 'default' : 
                  workflowStatus.status === 'rejected' ? 'destructive' : 'secondary'
                }
                className="text-xs px-3 py-1"
              >
                {workflowStatus.status?.toUpperCase() || 'UNKNOWN'}
              </Badge>
            </div>

            {/* Next Action */}
            <div className="flex items-center justify-end">
              {workflowStatus.nextApprover && workflowStatus.status !== 'completed' && workflowStatus.status !== 'rejected' && (
                <div className="text-right">
                  <p className="text-xs text-gray-500">Next approver:</p>
                  <p className="font-medium text-gray-700 truncate">
                    {workflowStatus.nextApprover}
                  </p>
                </div>
              )}
              {workflowStatus.status === 'completed' && (
                <div className="text-right text-green-600">
                  <CheckCircle className="h-4 w-4 inline mr-1" />
                  <span className="text-xs font-medium">Fully Approved</span>
                </div>
              )}
              {workflowStatus.status === 'rejected' && (
                <div className="text-right text-red-600">
                  <XCircle className="h-4 w-4 inline mr-1" />
                  <span className="text-xs font-medium">Rejected</span>
                </div>
              )}
            </div>
          </div>
        </div>
      )}
    </Card>
  )
}