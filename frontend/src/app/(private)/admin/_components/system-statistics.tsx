'use client'

import { useQuery } from '@tanstack/react-query'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { getDashboardMetrics } from '@/app/_actions/dashboard'
import { DashboardMetrics } from '@/types'
import { FileText, Clock, CheckCircle2, AlertCircle, TrendingUp } from 'lucide-react'
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts'
import { QUERY_KEYS } from '@/lib/constants'

export function SystemStatistics() {
  // Fetch metrics using React Query with caching
  const { data: metrics, isLoading } = useQuery<DashboardMetrics>({
    queryKey: [QUERY_KEYS.DASHBOARD.METRICS],
    queryFn: async () => {
      const result = await getDashboardMetrics()
      if (result.success && result.data) {
        return result.data
      }
      throw new Error('Failed to fetch metrics')
    },
    staleTime: 5 * 60 * 1000, // 5 minutes
  })

  if (isLoading || !metrics) {
    return (
      <div className="text-center py-8 text-muted-foreground">
        Loading system statistics...
      </div>
    )
  }

  // Prepare chart data for document types
  const chartData = Object.entries(metrics.documentTypeBreakdown || {}).map(([type, count]) => ({
    name: type === 'REQUISITION' ? 'Requisitions' : type === 'PURCHASE_ORDER' ? 'POs' : type === 'PAYMENT_VOUCHER' ? 'Vouchers' : 'GRNs',
    count,
  }))

  const successRate = metrics.totalDocuments > 0
    ? Math.round(((metrics.approvedDocuments || 0) / metrics.totalDocuments) * 100)
    : 0

  return (
    <div className="space-y-6">
      {/* Key Metrics Grid */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Total Documents
            </CardTitle>
            <FileText className="h-5 w-5 text-primary" />
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">{metrics.totalDocuments}</div>
            <p className="text-xs text-muted-foreground mt-1">All time</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Approval Rate
            </CardTitle>
            <TrendingUp className="h-5 w-5 text-secondary" />
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">{successRate}%</div>
            <p className="text-xs text-muted-foreground mt-1">
              {metrics.approvedDocuments} approved
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Avg Approval Time
            </CardTitle>
            <Clock className="h-5 w-5 text-accent" />
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">{metrics.averageApprovalTime}</div>
            <p className="text-xs text-muted-foreground mt-1">days</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Rejection Rate
            </CardTitle>
            <AlertCircle className="h-5 w-5 text-destructive" />
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">
              {metrics.totalDocuments > 0
                ? Math.round(((metrics.rejectedDocuments || 0) / metrics.totalDocuments) * 100)
                : 0}
              %
            </div>
            <p className="text-xs text-muted-foreground mt-1">
              {metrics.rejectedDocuments} rejected
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Status Breakdown */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Document Type Distribution</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="h-80 w-full">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={chartData}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="name" />
                <YAxis />
                <Tooltip />
                <Bar dataKey="count" fill="var(--primary)" />
              </BarChart>
            </ResponsiveContainer>
          </div>
        </CardContent>
      </Card>

      {/* Status Summary Table */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Status Summary</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            {[
              { label: 'Draft', value: metrics.draftDocuments, variant: 'outline' as const },
              { label: 'Submitted', value: metrics.submittedDocuments, variant: 'secondary' as const },
              { label: 'In Approval', value: metrics.statusBreakdown?.IN_REVIEW || 0, variant: 'default' as const },
              { label: 'Approved', value: metrics.approvedDocuments, variant: 'default' as const },
              { label: 'Rejected', value: metrics.rejectedDocuments, variant: 'destructive' as const },
            ].map((item) => (
              <div key={item.label} className="flex items-center justify-between p-3 border rounded-lg">
                <span className="font-medium">{item.label}</span>
                <Badge variant={item.variant}>{item.value}</Badge>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
