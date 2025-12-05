'use client'

import { useState } from 'react'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Button } from '@/components/ui/button'
import { ApprovalReports } from './approval-reports'
import { UserActivityReports } from './user-activity-reports'
import { SystemStatistics } from './system-statistics'
import { AnalyticsDashboard } from '@/components/workflows/analytics-dashboard'
import { Download, RefreshCw } from 'lucide-react'

interface AdminReportsClientProps {
  userId: string
  userRole: string
}

export function AdminReportsClient({ userId, userRole }: AdminReportsClientProps) {
  const [activeTab, setActiveTab] = useState('overview')
  const [isRefreshing, setIsRefreshing] = useState(false)

  const handleRefresh = async () => {
    setIsRefreshing(true)
    try {
      await new Promise((resolve) => setTimeout(resolve, 1000))
    } finally {
      setIsRefreshing(false)
    }
  }

  const handleExport = () => {
    const csv = `Workflow Analytics Report
Generated: ${new Date().toISOString()}

METRICS SUMMARY
Total Pending,24
Total Approved,187
Total Rejected,12
Avg Approval Time,3.2 days
SLA Compliance,94%`

    const blob = new Blob([csv], { type: 'text/csv' })
    const url = window.URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `analytics-report-${new Date().toISOString().split('T')[0]}.csv`
    a.click()
    window.URL.revokeObjectURL(url)
  }

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Admin Reports</h1>
          <p className="text-muted-foreground">
            Monitor workflow approvals, user activity, and system metrics
          </p>
        </div>
        <div className="flex gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={handleRefresh}
            disabled={isRefreshing}
          >
            <RefreshCw className={`h-4 w-4 mr-2 ${isRefreshing ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={handleExport}
          >
            <Download className="h-4 w-4 mr-2" />
            Export
          </Button>
        </div>
      </div>

      {/* Tabs */}
      <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
        <TabsList className="grid w-full grid-cols-4 lg:w-auto">
          <TabsTrigger value="overview">Overview</TabsTrigger>
          <TabsTrigger value="analytics">Analytics</TabsTrigger>
          <TabsTrigger value="approvals">Approvals</TabsTrigger>
          <TabsTrigger value="activity">Activity</TabsTrigger>
        </TabsList>

        {/* System Statistics Tab */}
        <TabsContent value="overview" className="space-y-6">
          <SystemStatistics />
        </TabsContent>

        {/* Analytics Tab */}
        <TabsContent value="analytics" className="space-y-6">
          <AnalyticsDashboard />
        </TabsContent>

        {/* Approval Reports Tab */}
        <TabsContent value="approvals" className="space-y-6">
          <ApprovalReports />
        </TabsContent>

        {/* User Activity Reports Tab */}
        <TabsContent value="activity" className="space-y-6">
          <UserActivityReports />
        </TabsContent>
      </Tabs>
    </div>
  )
}
