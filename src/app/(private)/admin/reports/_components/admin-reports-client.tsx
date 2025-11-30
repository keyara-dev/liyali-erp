'use client'

import { useState } from 'react'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { ApprovalReports } from './approval-reports'
import { UserActivityReports } from './user-activity-reports'
import { SystemStatistics } from './system-statistics'

interface AdminReportsClientProps {
  userId: string
  userRole: string
}

export function AdminReportsClient({ userId, userRole }: AdminReportsClientProps) {
  const [activeTab, setActiveTab] = useState('overview')

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div>
        <h1 className="text-xl font-bold tracking-tight lg:text-2xl">Admin Reports</h1>
        <p className="text-sm text-muted-foreground">
          Monitor workflow approvals, user activity, and system metrics
        </p>
      </div>

      {/* Tabs */}
      <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
        <TabsList className="grid w-full grid-cols-3 lg:w-auto">
          <TabsTrigger value="overview">Overview</TabsTrigger>
          <TabsTrigger value="approvals">Approvals</TabsTrigger>
          <TabsTrigger value="activity">Activity</TabsTrigger>
        </TabsList>

        {/* System Statistics Tab */}
        <TabsContent value="overview" className="space-y-6">
          <SystemStatistics />
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
