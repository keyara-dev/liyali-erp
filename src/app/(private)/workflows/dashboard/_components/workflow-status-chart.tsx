'use client'

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { DashboardMetrics } from '@/types'
import {
  PieChart,
  Pie,
  Cell,
  ResponsiveContainer,
  Legend,
  Tooltip,
} from 'recharts'

interface WorkflowStatusChartProps {
  metrics: DashboardMetrics
}

export function WorkflowStatusChart({ metrics }: WorkflowStatusChartProps) {
  const data = [
    { name: 'Draft', value: metrics.draftDocuments },
    { name: 'Submitted', value: metrics.submittedDocuments },
    { name: 'In Approval', value: metrics.statusBreakdown.IN_REVIEW },
    { name: 'Approved', value: metrics.approvedDocuments },
    { name: 'Rejected', value: metrics.rejectedDocuments },
  ].filter((item) => item.value > 0)

  const COLORS = [
    'var(--chart-1)',
    'var(--chart-2)',
    'var(--chart-3)',
    'var(--chart-4)',
    'var(--chart-5)',
  ]

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">Workflow Status</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="h-64 w-full">
          <ResponsiveContainer width="100%" height="100%">
            <PieChart>
              <Pie
                data={data}
                cx="50%"
                cy="50%"
                labelLine={false}
                label={({ name, value }) => `${name}: ${value}`}
                outerRadius={80}
                fill="#8884d8"
                dataKey="value"
              >
                {data.map((entry, index) => (
                  <Cell
                    key={`cell-${index}`}
                    fill={COLORS[index % COLORS.length]}
                  />
                ))}
              </Pie>
              <Tooltip />
              <Legend />
            </PieChart>
          </ResponsiveContainer>
        </div>
      </CardContent>
    </Card>
  )
}
