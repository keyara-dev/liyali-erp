'use client'

import { useEffect, useState } from 'react'
import { toast } from 'sonner'
import { Card } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { AlertCircle, Clock, CheckCircle, XCircle } from 'lucide-react'
import { getApprovalLog, getDocumentApprovers } from '@/app/_actions/workflow'
import { ApprovalLogEntry, Approver, WorkflowDocument } from '@/types/workflow'
import { ApprovalActionPanel } from './approval-action-panel'

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
  const [approvalLogs, setApprovalLogs] = useState<ApprovalLogEntry[]>([])
  const [approvers, setApprovers] = useState<Approver[]>([])
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    fetchApprovalData()
  }, [requisitionId])

  const fetchApprovalData = async () => {
    setIsLoading(true)
    try {
      const [logsResult, approversResult] = await Promise.all([
        getApprovalLog(requisitionId),
        getDocumentApprovers(requisitionId),
      ])

      if (logsResult.success) {
        setApprovalLogs(logsResult.data || [])
      }
      if (approversResult.success) {
        setApprovers(approversResult.data || [])
      }
    } catch (error) {
      console.error('Failed to fetch approval data:', error)
    } finally {
      setIsLoading(false)
    }
  }

  const getActionIcon = (action: string) => {
    switch (action) {
      case 'APPROVED':
        return <CheckCircle className="h-5 w-5 text-green-600" />
      case 'REJECTED':
        return <XCircle className="h-5 w-5 text-red-600" />
      default:
        return <Clock className="h-5 w-5 text-gray-600" />
    }
  }

  const getActionColor = (action: string) => {
    switch (action) {
      case 'APPROVED':
        return 'bg-green-50'
      case 'REJECTED':
        return 'bg-red-50'
      default:
        return 'bg-gray-50'
    }
  }

  return (
    <Card className="p-6">
      <Tabs defaultValue="history" className="w-full">
        <TabsList className="grid w-full grid-cols-2">
          <TabsTrigger value="history">Approval Log</TabsTrigger>
          <TabsTrigger value="approvers">Approvers</TabsTrigger>
        </TabsList>

        <TabsContent value="history" className="space-y-4 mt-4">
          {isLoading ? (
            <div className="text-center py-8">
              <div className="inline-block h-6 w-6 rounded-full border-2 border-blue-200 border-t-blue-600 animate-spin"></div>
            </div>
          ) : approvalLogs.length > 0 ? (
            <div className="space-y-3 max-h-96 overflow-y-auto">
              {approvalLogs.map((log) => (
                <div
                  key={log.id}
                  className={`p-3 rounded-lg ${getActionColor(log.action)}`}
                >
                  <div className="flex items-start gap-3">
                    {getActionIcon(log.action)}
                    <div className="flex-1">
                      <div className="flex items-center gap-2">
                        <span className="font-semibold text-sm">
                          {log.approver.name}
                        </span>
                        <Badge variant="outline" className="text-xs">
                          {log.action}
                        </Badge>
                      </div>
                      <p className="text-xs text-gray-600 mt-1">
                        {new Date(log.timestamp).toLocaleString()}
                      </p>
                      {log.comments && (
                        <p className="text-sm mt-2 text-gray-700">
                          "{log.comments}"
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
            </div>
          )}
        </TabsContent>

        <TabsContent value="approvers" className="space-y-4 mt-4">
          {isLoading ? (
            <div className="text-center py-8">
              <div className="inline-block h-6 w-6 rounded-full border-2 border-blue-200 border-t-blue-600 animate-spin"></div>
            </div>
          ) : approvers.length > 0 ? (
            <div className="space-y-2 max-h-96 overflow-y-auto">
              {approvers.map((approver) => (
                <div
                  key={approver.id}
                  className="p-3 border rounded-lg hover:bg-gray-50"
                >
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="font-semibold text-sm">
                        {approver.user?.name || 'Unknown'}
                      </p>
                      <p className="text-xs text-gray-600">
                        Stage {approver.stepOrder}
                      </p>
                    </div>
                    <Badge
                      variant={
                        approver.status === 'APPROVED'
                          ? 'default'
                          : approver.status === 'REJECTED'
                          ? 'destructive'
                          : 'secondary'
                      }
                      className="text-xs"
                    >
                      {approver.status}
                    </Badge>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="text-center py-8 text-gray-500">
              <AlertCircle className="h-8 w-8 mx-auto mb-2 text-gray-400" />
              <p className="text-sm">No approvers assigned yet</p>
            </div>
          )}
        </TabsContent>
      </Tabs>

      {/* Approval Action Panel */}
      {requisition.status === 'IN_APPROVAL' && (
        <div className="mt-6 pt-6 border-t">
          <ApprovalActionPanel
            requisitionId={requisitionId}
            onApprovalComplete={fetchApprovalData}
          />
        </div>
      )}
    </Card>
  )
}
