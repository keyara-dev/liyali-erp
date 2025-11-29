'use client'

import { useRouter } from 'next/navigation'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { DashboardMetrics } from '@/app/_actions/dashboard'
import { Eye } from 'lucide-react'

interface RecentActivityProps {
  metrics: DashboardMetrics
}

const STATUS_COLORS: Record<string, string> = {
  DRAFT: 'outline',
  SUBMITTED: 'secondary',
  IN_APPROVAL: 'default',
  APPROVED: 'default',
  REJECTED: 'destructive',
  REVERSED: 'secondary',
}

const STATUS_LABELS: Record<string, string> = {
  DRAFT: 'Draft',
  SUBMITTED: 'Submitted',
  IN_APPROVAL: 'In Review',
  APPROVED: 'Approved',
  REJECTED: 'Rejected',
  REVERSED: 'Reversed',
}

const DOCUMENT_TYPE_LABELS: Record<string, string> = {
  REQUISITION: 'Requisition',
  PURCHASE_ORDER: 'Purchase Order',
  PAYMENT_VOUCHER: 'Payment Voucher',
  GOODS_RECEIVED_NOTE: 'GRN',
}

export function RecentActivity({ metrics }: RecentActivityProps) {
  const router = useRouter()

  const typeSlugMap = {
    REQUISITION: 'requisitions',
    PURCHASE_ORDER: 'purchase-orders',
    PAYMENT_VOUCHER: 'payment-vouchers',
    GOODS_RECEIVED_NOTE: 'grn',
  }

  if (metrics.recentActivity.length === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="text-base">Recent Activity</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="text-center py-8 text-muted-foreground">
            No recent activity
          </div>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">Recent Activity</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="rounded-md border overflow-hidden">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Document</TableHead>
                <TableHead>Type</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>User</TableHead>
                <TableHead>Time</TableHead>
                <TableHead>Action</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {metrics.recentActivity.map((activity) => {
                const typeSlug = typeSlugMap[activity.type as keyof typeof typeSlugMap] || 'workflows'
                return (
                  <TableRow key={activity.id}>
                    <TableCell className="font-medium text-primary hover:underline cursor-pointer">
                      {activity.documentNumber}
                    </TableCell>
                    <TableCell className="text-sm">
                      {DOCUMENT_TYPE_LABELS[activity.type] || activity.type}
                    </TableCell>
                    <TableCell>
                      <Badge variant={STATUS_COLORS[activity.action] as any}>
                        {STATUS_LABELS[activity.action] || activity.action}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {activity.user}
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {new Date(activity.timestamp).toLocaleString([], {
                        month: 'short',
                        day: 'numeric',
                        hour: '2-digit',
                        minute: '2-digit',
                      })}
                    </TableCell>
                    <TableCell>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => router.push(`/workflows/${typeSlug}/${activity.id}`)}
                        className="gap-1"
                      >
                        <Eye className="h-4 w-4" />
                        View
                      </Button>
                    </TableCell>
                  </TableRow>
                )
              })}
            </TableBody>
          </Table>
        </div>
      </CardContent>
    </Card>
  )
}
