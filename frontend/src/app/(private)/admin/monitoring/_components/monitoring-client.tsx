'use client'

import { useState, useEffect } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { AlertCircle, Activity, Zap, Database, Clock } from 'lucide-react'
import {
  LineChart,
  Line,
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend,
} from 'recharts'

interface MonitoringClientProps {
  userId: string
  userRole: string
}

// Mock real-time metrics data
const generateMetricsData = () => {
  const data = []
  const now = new Date()
  for (let i = 23; i >= 0; i--) {
    data.unshift({
      time: `${String(now.getHours() - i).padStart(2, '0')}:00`,
      approvals: Math.floor(Math.random() * 20) + 5,
      submissions: Math.floor(Math.random() * 30) + 10,
      rejections: Math.floor(Math.random() * 5) + 1,
    })
  }
  return data
}

const generateSystemMetrics = () => {
  const data = []
  for (let i = 30; i >= 0; i--) {
    data.unshift({
      time: `${30 - i}m`,
      cpu: Math.floor(Math.random() * 40) + 30,
      memory: Math.floor(Math.random() * 50) + 40,
      disk: Math.floor(Math.random() * 30) + 50,
    })
  }
  return data
}

export function MonitoringClient({ userId, userRole }: MonitoringClientProps) {
  const [metricsData] = useState(generateMetricsData())
  const [systemData] = useState(generateSystemMetrics())
  const [activeTab, setActiveTab] = useState('overview')

  // Mock system health status
  const systemStatus = {
    database: 'healthy',
    apiServer: 'healthy',
    cache: 'healthy',
    storage: 'healthy',
    avgResponseTime: 145,
    errorRate: 0.2,
    uptime: 99.98,
  }

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div>
        <h1 className="text-xl font-bold tracking-tight lg:text-2xl">System Monitoring</h1>
        <p className="text-sm text-muted-foreground">
          Real-time system performance and workflow metrics
        </p>
      </div>

      {/* System Health Cards */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Avg Response Time
            </CardTitle>
            <Zap className="h-5 w-5 text-primary" />
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">{systemStatus.avgResponseTime}</div>
            <p className="text-xs text-muted-foreground mt-1">milliseconds</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Error Rate
            </CardTitle>
            <AlertCircle className="h-5 w-5 text-secondary" />
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">{systemStatus.errorRate}</div>
            <p className="text-xs text-muted-foreground mt-1">percent</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Uptime
            </CardTitle>
            <Activity className="h-5 w-5 text-accent" />
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">{systemStatus.uptime}</div>
            <p className="text-xs text-muted-foreground mt-1">percent</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Services
            </CardTitle>
            <Database className="h-5 w-5 text-green-500" />
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">4/4</div>
            <p className="text-xs text-muted-foreground mt-1">operational</p>
          </CardContent>
        </Card>
      </div>

      {/* Tabs */}
      <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
        <TabsList className="grid w-full grid-cols-3 lg:w-auto">
          <TabsTrigger value="overview">Workflow Activity</TabsTrigger>
          <TabsTrigger value="performance">Performance</TabsTrigger>
          <TabsTrigger value="health">System Health</TabsTrigger>
        </TabsList>

        {/* Workflow Activity Tab */}
        <TabsContent value="overview" className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">Workflow Activity (24h)</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="h-80 w-full">
                <ResponsiveContainer width="100%" height="100%">
                  <LineChart data={metricsData}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="time" />
                    <YAxis />
                    <Tooltip />
                    <Legend />
                    <Line
                      type="monotone"
                      dataKey="approvals"
                      stroke="var(--secondary)"
                      strokeWidth={2}
                      dot={false}
                      name="Approvals"
                    />
                    <Line
                      type="monotone"
                      dataKey="submissions"
                      stroke="var(--primary)"
                      strokeWidth={2}
                      dot={false}
                      name="Submissions"
                    />
                    <Line
                      type="monotone"
                      dataKey="rejections"
                      stroke="var(--destructive)"
                      strokeWidth={2}
                      dot={false}
                      name="Rejections"
                    />
                  </LineChart>
                </ResponsiveContainer>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Performance Tab */}
        <TabsContent value="performance" className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">System Resources (30min)</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="h-80 w-full">
                <ResponsiveContainer width="100%" height="100%">
                  <AreaChart data={systemData}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="time" />
                    <YAxis domain={[0, 100]} />
                    <Tooltip formatter={(value) => `${value}%`} />
                    <Legend />
                    <Area
                      type="monotone"
                      dataKey="cpu"
                      stroke="var(--primary)"
                      fill="var(--primary)"
                      fillOpacity={0.2}
                      name="CPU Usage"
                    />
                    <Area
                      type="monotone"
                      dataKey="memory"
                      stroke="var(--secondary)"
                      fill="var(--secondary)"
                      fillOpacity={0.2}
                      name="Memory Usage"
                    />
                  </AreaChart>
                </ResponsiveContainer>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* System Health Tab */}
        <TabsContent value="health" className="space-y-4">
          <div className="grid gap-4 md:grid-cols-2">
            {[
              { name: 'Database', status: systemStatus.database },
              { name: 'API Server', status: systemStatus.apiServer },
              { name: 'Cache Layer', status: systemStatus.cache },
              { name: 'Storage', status: systemStatus.storage },
            ].map((service) => (
              <Card key={service.name}>
                <CardContent className="pt-6">
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="font-semibold">{service.name}</p>
                      <p className="text-sm text-muted-foreground mt-1">
                        Status: {service.status}
                      </p>
                    </div>
                    <Badge
                      variant={service.status === 'healthy' ? 'default' : 'destructive'}
                      className={service.status === 'healthy' ? 'bg-secondary' : ''}
                    >
                      {service.status === 'healthy' ? '✓ Healthy' : '✗ Issues'}
                    </Badge>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>

          {/* Detailed Health Info */}
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">Health Metrics</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-3">
                {[
                  { label: 'Average Response Time', value: `${systemStatus.avgResponseTime}ms`, status: 'healthy' },
                  { label: 'Error Rate', value: `${systemStatus.errorRate}%`, status: 'healthy' },
                  { label: 'System Uptime', value: `${systemStatus.uptime}%`, status: 'healthy' },
                  { label: 'Database Connections', value: '24/50', status: 'healthy' },
                  { label: 'Active Workflows', value: '18', status: 'healthy' },
                  { label: 'Pending Approvals', value: '12', status: 'warning' },
                ].map((metric, index) => (
                  <div key={index} className="flex items-center justify-between p-3 border rounded-lg">
                    <span className="font-medium">{metric.label}</span>
                    <div className="flex items-center gap-2">
                      <span className="text-right">{metric.value}</span>
                      <Badge
                        variant={metric.status === 'healthy' ? 'default' : 'outline'}
                        className={metric.status === 'healthy' ? 'bg-secondary' : ''}
                      >
                        {metric.status}
                      </Badge>
                    </div>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      {/* Live Feed */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg flex items-center gap-2">
            <Clock className="h-5 w-5" />
            Live Event Feed
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-3 max-h-96 overflow-y-auto">
            {[
              { time: '14:32:15', event: 'Requisition REQ-2024-025 submitted', type: 'info' },
              { time: '14:31:42', event: 'Purchase Order PO-2024-089 approved', type: 'success' },
              { time: '14:30:18', event: 'Payment Voucher PV-2024-031 rejected', type: 'error' },
              { time: '14:29:05', event: 'System backup completed successfully', type: 'success' },
              { time: '14:27:33', event: 'Database optimization running', type: 'warning' },
              { time: '14:26:12', event: 'User Sarah Banda logged in', type: 'info' },
              { time: '14:24:47', event: 'Requisition REQ-2024-024 approved by Director', type: 'success' },
              { time: '14:23:19', event: 'API response time elevated to 245ms', type: 'warning' },
            ].map((entry, index) => (
              <div key={index} className="flex items-start gap-3 p-2 rounded border">
                <span className="text-xs text-muted-foreground whitespace-nowrap">
                  {entry.time}
                </span>
                <span className="text-sm flex-1">{entry.event}</span>
                <Badge
                  variant={
                    entry.type === 'success'
                      ? 'default'
                      : entry.type === 'error'
                        ? 'destructive'
                        : entry.type === 'warning'
                          ? 'outline'
                          : 'outline'
                  }
                  className={entry.type === 'success' ? 'bg-secondary' : ''}
                >
                  {entry.type}
                </Badge>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
