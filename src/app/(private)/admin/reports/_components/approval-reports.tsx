'use client'

import { useEffect, useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { getDashboardMetrics } from '@/app/_actions/dashboard'
import { DashboardMetrics } from '@/types'
import { Input } from '@/components/ui/input'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Search } from 'lucide-react'

export function ApprovalReports() {
  const [metrics, setMetrics] = useState<DashboardMetrics | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [searchTerm, setSearchTerm] = useState('')

  useEffect(() => {
    async function fetchMetrics() {
      setIsLoading(true)
      try {
        const result = await getDashboardMetrics()
        if (result.success && result.data) {
          setMetrics(result.data)
        }
      } catch (error) {
        console.error('Failed to fetch metrics:', error)
      } finally {
        setIsLoading(false)
      }
    }

    fetchMetrics()
  }, [])

  if (isLoading || !metrics) {
    return (
      <div className="text-center py-8 text-muted-foreground">
        Loading approval reports...
      </div>
    )
  }

  const filteredActivity = metrics.recentActivity.filter((item) =>
    item.documentNumber.toLowerCase().includes(searchTerm.toLowerCase()) ||
    item.user.toLowerCase().includes(searchTerm.toLowerCase())
  )

  const statusColors: Record<string, string> = {
    APPROVED: 'default',
    REJECTED: 'destructive',
    IN_APPROVAL: 'default',
    SUBMITTED: 'secondary',
    DRAFT: 'outline',
  }

  return (
    <div className="space-y-6">
      {/* Approval Metrics */}
      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardHeader>
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Approved This Period
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold text-secondary">
              {metrics.approvedDocuments}
            </div>
            <p className="text-xs text-muted-foreground mt-1">
              {metrics.totalDocuments > 0
                ? Math.round((metrics.approvedDocuments / metrics.totalDocuments) * 100)
                : 0}
              % of total
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Rejections
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold text-destructive">
              {metrics.rejectedDocuments}
            </div>
            <p className="text-xs text-muted-foreground mt-1">
              {metrics.totalDocuments > 0
                ? Math.round((metrics.rejectedDocuments / metrics.totalDocuments) * 100)
                : 0}
              % of total
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Pending Review
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold text-accent">
              {metrics.pendingApproval}
            </div>
            <p className="text-xs text-muted-foreground mt-1">
              Awaiting next approver
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Recent Approvals Table */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Recent Approvals & Actions</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* Search */}
          <div className="relative">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="Search by document or user..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="pl-10"
            />
          </div>

          {/* Table */}
          <div className="rounded-md border overflow-hidden">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Document</TableHead>
                  <TableHead>Type</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Approver</TableHead>
                  <TableHead>Time</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredActivity.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={5} className="text-center py-4 text-muted-foreground">
                      No approvals found
                    </TableCell>
                  </TableRow>
                ) : (
                  filteredActivity.map((activity) => (
                    <TableRow key={activity.id}>
                      <TableCell className="font-medium text-primary">
                        {activity.documentNumber}
                      </TableCell>
                      <TableCell className="text-sm">
                        {activity.type === 'REQUISITION'
                          ? 'Requisition'
                          : activity.type === 'PURCHASE_ORDER'
                            ? 'PO'
                            : activity.type === 'PAYMENT_VOUCHER'
                              ? 'Voucher'
                              : 'GRN'}
                      </TableCell>
                      <TableCell>
                        <Badge variant={statusColors[activity.action] as any}>
                          {activity.action}
                        </Badge>
                      </TableCell>
                      <TableCell className="text-sm text-muted-foreground">
                        {activity.user}
                      </TableCell>
                      <TableCell className="text-sm text-muted-foreground">
                        {new Date(activity.timestamp).toLocaleDateString([], {
                          month: 'short',
                          day: 'numeric',
                          hour: '2-digit',
                          minute: '2-digit',
                        })}
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
