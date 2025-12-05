'use client'

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import {
  TrendingUp,
  CheckCircle2,
  AlertCircle,
  Clock,
  BarChart3,
  PieChart as PieChartIcon,
  LineChart as LineChartIcon,
  Target,
} from 'lucide-react'

interface MetricsData {
  totalPending: number
  totalApproved: number
  totalRejected: number
  avgApprovalTime: string
  slaCompliance: number
  bottleneckStage: string
  bottleneckDays: number
}

interface TimelineData {
  date: string
  approved: number
  rejected: number
  pending: number
}

interface StageMetricsData {
  stage: string
  avgTime: string
  count: number
  slaCompliance: number
}

// Mock metrics data
const METRICS_DATA: MetricsData = {
  totalPending: 24,
  totalApproved: 187,
  totalRejected: 12,
  avgApprovalTime: '3.2 days',
  slaCompliance: 94,
  bottleneckStage: 'Finance Officer Review',
  bottleneckDays: 4.5,
}

const TIMELINE_DATA: TimelineData[] = [
  { date: 'Nov 20', approved: 8, rejected: 1, pending: 5 },
  { date: 'Nov 21', approved: 12, rejected: 2, pending: 8 },
  { date: 'Nov 22', approved: 15, rejected: 1, pending: 12 },
  { date: 'Nov 23', approved: 18, rejected: 3, pending: 15 },
  { date: 'Nov 24', approved: 22, rejected: 2, pending: 18 },
  { date: 'Nov 25', approved: 28, rejected: 1, pending: 22 },
  { date: 'Nov 26', approved: 35, rejected: 2, pending: 24 },
]

const STAGE_METRICS: StageMetricsData[] = [
  { stage: 'Department Manager', avgTime: '1.2 days', count: 45, slaCompliance: 98 },
  { stage: 'Finance Officer', avgTime: '4.5 days', count: 38, slaCompliance: 85 },
  { stage: 'Director/CFO', avgTime: '2.1 days', count: 42, slaCompliance: 95 },
]

const APPROVAL_DISTRIBUTION = [
  { type: 'Requisition', count: 67, percentage: 28 },
  { type: 'Budget', count: 58, percentage: 24 },
  { type: 'Purchase Order', count: 54, percentage: 22 },
  { type: 'Payment Voucher', count: 42, percentage: 17 },
  { type: 'GRN', count: 20, percentage: 9 },
]

export function AnalyticsDashboard() {
  return (
    <div className="space-y-6">
      {/* Key Metrics */}
      <div className="grid gap-4 md:grid-cols-5">
        {/* Total Pending */}
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Total Pending
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">{METRICS_DATA.totalPending}</div>
            <p className="text-xs text-muted-foreground mt-2">
              Awaiting approval
            </p>
            <div className="mt-3 h-2 bg-yellow-200 rounded-full"></div>
          </CardContent>
        </Card>

        {/* Total Approved */}
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Total Approved
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold text-green-600">{METRICS_DATA.totalApproved}</div>
            <p className="text-xs text-muted-foreground mt-2">
              This period
            </p>
            <div className="mt-3 h-2 bg-green-200 rounded-full"></div>
          </CardContent>
        </Card>

        {/* Total Rejected */}
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Total Rejected
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold text-red-600">{METRICS_DATA.totalRejected}</div>
            <p className="text-xs text-muted-foreground mt-2">
              Returned to requester
            </p>
            <div className="mt-3 h-2 bg-red-200 rounded-full"></div>
          </CardContent>
        </Card>

        {/* Avg Approval Time */}
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Avg Approval Time
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold text-blue-600">
              {METRICS_DATA.avgApprovalTime}
            </div>
            <p className="text-xs text-muted-foreground mt-2">
              Average turnaround
            </p>
            <div className="mt-3 h-2 bg-blue-200 rounded-full"></div>
          </CardContent>
        </Card>

        {/* SLA Compliance */}
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              SLA Compliance
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">{METRICS_DATA.slaCompliance}%</div>
            <p className="text-xs text-muted-foreground mt-2">
              On-time delivery
            </p>
            <div className="mt-3 h-2 bg-green-200 rounded-full" style={{ width: `${METRICS_DATA.slaCompliance}%` }}></div>
          </CardContent>
        </Card>
      </div>

      {/* Trends and Distribution */}
      <div className="grid gap-6 md:grid-cols-2">
        {/* Approval Timeline */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <LineChartIcon className="h-5 w-5" />
              Approval Trends (Last 7 Days)
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {TIMELINE_DATA.map((item, index) => (
              <div key={index} className="space-y-2">
                <div className="flex justify-between text-sm">
                  <span className="text-muted-foreground">{item.date}</span>
                  <span className="font-medium">
                    ✓ {item.approved} | ✗ {item.rejected} | ⏳ {item.pending}
                  </span>
                </div>
                <div className="flex gap-1 h-2 bg-gray-100 rounded-full overflow-hidden">
                  <div
                    className="bg-green-500"
                    style={{ width: `${(item.approved / 40) * 100}%` }}
                  ></div>
                  <div
                    className="bg-red-500"
                    style={{ width: `${(item.rejected / 40) * 100}%` }}
                  ></div>
                  <div
                    className="bg-yellow-500"
                    style={{ width: `${(item.pending / 40) * 100}%` }}
                  ></div>
                </div>
              </div>
            ))}
          </CardContent>
        </Card>

        {/* Document Type Distribution */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <PieChartIcon className="h-5 w-5" />
              Approvals by Document Type
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            {APPROVAL_DISTRIBUTION.map((item, index) => (
              <div key={index} className="space-y-2">
                <div className="flex justify-between text-sm">
                  <span className="font-medium">{item.type}</span>
                  <span className="text-muted-foreground">{item.count} ({item.percentage}%)</span>
                </div>
                <div className="h-3 bg-gray-100 rounded-full overflow-hidden">
                  <div
                    className="h-full bg-gradient-to-r from-blue-500 to-purple-500"
                    style={{ width: `${item.percentage}%` }}
                  ></div>
                </div>
              </div>
            ))}
          </CardContent>
        </Card>
      </div>

      {/* Stage Performance and Bottleneck */}
      <div className="grid gap-6 md:grid-cols-2">
        {/* Stage Performance */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <BarChart3 className="h-5 w-5" />
              Stage Performance Metrics
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {STAGE_METRICS.map((metric, index) => (
              <div key={index} className="space-y-2 pb-4 border-b last:border-b-0">
                <div className="flex justify-between items-start">
                  <div>
                    <p className="font-medium">{metric.stage}</p>
                    <p className="text-sm text-muted-foreground">{metric.count} items processed</p>
                  </div>
                  <Badge
                    variant="outline"
                    className={
                      metric.slaCompliance >= 90
                        ? 'bg-green-50 text-green-700 border-green-200'
                        : metric.slaCompliance >= 80
                        ? 'bg-yellow-50 text-yellow-700 border-yellow-200'
                        : 'bg-red-50 text-red-700 border-red-200'
                    }
                  >
                    {metric.slaCompliance}% SLA
                  </Badge>
                </div>
                <div className="space-y-1">
                  <div className="flex justify-between text-xs">
                    <span className="text-muted-foreground">Avg Processing Time</span>
                    <span className="font-medium">{metric.avgTime}</span>
                  </div>
                  <div className="h-2 bg-gray-100 rounded-full overflow-hidden">
                    <div
                      className={
                        metric.slaCompliance >= 90
                          ? 'h-full bg-green-500'
                          : metric.slaCompliance >= 80
                          ? 'h-full bg-yellow-500'
                          : 'h-full bg-red-500'
                      }
                      style={{ width: `${metric.slaCompliance}%` }}
                    ></div>
                  </div>
                </div>
              </div>
            ))}
          </CardContent>
        </Card>

        {/* Bottleneck Analysis */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <AlertCircle className="h-5 w-5" />
              Bottleneck Analysis
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-6">
            {/* Current Bottleneck */}
            <div className="p-4 border-2 border-orange-200 rounded-lg bg-orange-50">
              <div className="flex items-start gap-3">
                <Target className="h-5 w-5 text-orange-600 flex-shrink-0 mt-0.5" />
                <div className="flex-1">
                  <p className="font-semibold text-orange-900">Current Bottleneck</p>
                  <p className="text-sm text-orange-800 mt-1">
                    {METRICS_DATA.bottleneckStage}
                  </p>
                  <p className="text-xs text-orange-700 mt-2">
                    ⏱️ Average {METRICS_DATA.bottleneckDays} days at this stage
                  </p>
                </div>
              </div>
            </div>

            {/* Recommendations */}
            <div className="space-y-2">
              <p className="font-medium text-sm">Recommendations:</p>
              <ul className="space-y-2 text-sm text-muted-foreground">
                <li className="flex gap-2">
                  <span className="text-green-600 font-bold">•</span>
                  <span>Consider adding additional Finance Officer capacity</span>
                </li>
                <li className="flex gap-2">
                  <span className="text-green-600 font-bold">•</span>
                  <span>Review approval criteria for faster processing</span>
                </li>
                <li className="flex gap-2">
                  <span className="text-green-600 font-bold">•</span>
                  <span>Implement parallel approvals where applicable</span>
                </li>
              </ul>
            </div>

            {/* Trend */}
            <div className="flex items-center gap-2 pt-2 border-t">
              <TrendingUp className="h-4 w-4 text-green-600" />
              <p className="text-xs text-muted-foreground">
                Bottleneck reducing: was 5.2 days last week
              </p>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Performance Summary */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <CheckCircle2 className="h-5 w-5" />
            Performance Summary
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 md:grid-cols-3">
            <div className="space-y-2">
              <p className="text-sm font-medium">✅ Strengths</p>
              <ul className="text-sm text-muted-foreground space-y-1">
                <li>• High overall SLA compliance (94%)</li>
                <li>• Fast Department Manager approvals</li>
                <li>• Consistent approval rates</li>
              </ul>
            </div>
            <div className="space-y-2">
              <p className="text-sm font-medium">⚠️ Areas to Improve</p>
              <ul className="text-sm text-muted-foreground space-y-1">
                <li>• Finance Officer stage bottleneck</li>
                <li>• Higher rejection rate needed review</li>
                <li>• GRN processing efficiency</li>
              </ul>
            </div>
            <div className="space-y-2">
              <p className="text-sm font-medium">📊 Key Actions</p>
              <ul className="text-sm text-muted-foreground space-y-1">
                <li>• Monitor Finance Officer queue</li>
                <li>• Review rejected items trends</li>
                <li>• Optimize approval workflow</li>
              </ul>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
