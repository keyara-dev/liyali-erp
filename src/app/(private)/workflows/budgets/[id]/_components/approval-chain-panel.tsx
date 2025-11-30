'use client'

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { ApprovalRecord } from '@/types/budget'
import { CheckCircle2, Clock, XCircle } from 'lucide-react'

interface ApprovalChainPanelProps {
  approvalChain: ApprovalRecord[]
}

export function ApprovalChainPanel({ approvalChain }: ApprovalChainPanelProps) {
  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'APPROVED':
        return <CheckCircle2 className="h-5 w-5 text-green-600" />
      case 'REJECTED':
        return <XCircle className="h-5 w-5 text-red-600" />
      case 'PENDING':
        return <Clock className="h-5 w-5 text-yellow-600" />
      default:
        return <Clock className="h-5 w-5 text-gray-400" />
    }
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'APPROVED':
        return 'bg-green-50'
      case 'REJECTED':
        return 'bg-red-50'
      case 'PENDING':
        return 'bg-yellow-50'
      default:
        return 'bg-gray-50'
    }
  }

  if (!approvalChain || approvalChain.length === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Approval Chain</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground">
            No approval records yet. Submit the budget to initiate the approval process.
          </p>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Approval Chain</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          {approvalChain.map((record, index) => (
            <div key={index} className={`rounded-lg p-4 ${getStatusColor(record.status)}`}>
              <div className="flex items-start gap-3">
                <div className="mt-1">{getStatusIcon(record.status)}</div>
                <div className="flex-1">
                  <div className="flex items-center justify-between">
                    <h4 className="font-semibold">{record.stageName}</h4>
                    <span className={`px-2 py-1 rounded text-xs font-medium ${
                      record.status === 'APPROVED'
                        ? 'bg-green-200 text-green-800'
                        : record.status === 'REJECTED'
                        ? 'bg-red-200 text-red-800'
                        : 'bg-yellow-200 text-yellow-800'
                    }`}>
                      {record.status}
                    </span>
                  </div>
                  <p className="text-sm text-muted-foreground mt-1">
                    Assigned to: <span className="font-medium">{record.assignedTo}</span>
                  </p>
                  {record.assignedRole && (
                    <p className="text-sm text-muted-foreground">
                      Role: <span className="font-medium">{record.assignedRole}</span>
                    </p>
                  )}
                  {record.actionTakenAt && (
                    <p className="text-sm text-muted-foreground mt-1">
                      Action taken on: <span className="font-medium">
                        {record.actionTakenAt.toLocaleDateString()} {record.actionTakenAt.toLocaleTimeString()}
                      </span>
                    </p>
                  )}
                  {record.actionTakenBy && (
                    <p className="text-sm text-muted-foreground">
                      By: <span className="font-medium">{record.actionTakenBy}</span>
                    </p>
                  )}
                  {record.comments && (
                    <div className="mt-2 p-2 bg-white rounded text-sm italic">
                      "{record.comments}"
                    </div>
                  )}
                </div>
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  )
}
