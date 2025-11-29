'use client'

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { DashboardMetrics } from '@/app/_actions/dashboard'
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from 'recharts'

interface ApprovalTimeChartProps {
  metrics: DashboardMetrics
}

export function ApprovalTimeChart({ metrics }: ApprovalTimeChartProps) {
  // Generate mock approval time trend data
  const data = [
    { day: 'Mon', avgTime: 2.5 },
    { day: 'Tue', avgTime: 3.2 },
    { day: 'Wed', avgTime: 2.8 },
    { day: 'Thu', avgTime: metrics.averageApprovalTime || 3.0 },
    { day: 'Fri', avgTime: 3.5 },
    { day: 'Sat', avgTime: 2.2 },
    { day: 'Sun', avgTime: 1.8 },
  ]

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">Approval Time Trend</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="h-64 w-full">
          <ResponsiveContainer width="100%" height="100%">
            <LineChart data={data}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis
                dataKey="day"
                stroke="var(--muted-foreground)"
                style={{ fontSize: '0.75rem' }}
              />
              <YAxis
                stroke="var(--muted-foreground)"
                style={{ fontSize: '0.75rem' }}
                label={{ value: 'Days', angle: -90, position: 'insideLeft' }}
              />
              <Tooltip
                contentStyle={{
                  backgroundColor: 'var(--background)',
                  border: '1px solid var(--border)',
                  borderRadius: '0.5rem',
                }}
                formatter={(value) => [`${value.toFixed(1)} days`, 'Avg Time']}
              />
              <Line
                type="monotone"
                dataKey="avgTime"
                stroke="var(--primary)"
                strokeWidth={2}
                dot={{ fill: 'var(--primary)', r: 4 }}
                activeDot={{ r: 6 }}
              />
            </LineChart>
          </ResponsiveContainer>
        </div>
        <div className="mt-4 text-center">
          <p className="text-sm text-muted-foreground">
            Average approval time: <span className="font-semibold text-foreground">{metrics.averageApprovalTime} days</span>
          </p>
        </div>
      </CardContent>
    </Card>
  )
}
