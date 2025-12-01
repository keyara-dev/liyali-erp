'use client'

import { useState, useEffect } from 'react'
import { useSearchParams } from 'next/navigation'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Button } from '@/components/ui/button'
import { TasksTable } from './tasks-table'
import { TaskStatsCards } from './task-stats-cards'
import { ApprovalsList } from './approvals-list'

interface TasksClientProps {
  userId: string
  userRole: string
}

export function TasksClient({
  userId,
  userRole,
}: TasksClientProps) {
  const searchParams = useSearchParams()
  const [refreshTrigger, setRefreshTrigger] = useState(0)
  const [selectedStatus, setSelectedStatus] = useState<'all' | 'PENDING' | 'IN_PROGRESS'>('all')
  const [activeTab, setActiveTab] = useState<'tasks' | 'approvals'>('tasks')

  // Check for tab query parameter on mount
  useEffect(() => {
    const tabParam = searchParams.get('tab')
    if (tabParam === 'approvals') {
      setActiveTab('approvals')
    }
  }, [searchParams])

  const handleTaskAction = () => {
    setRefreshTrigger((prev) => prev + 1)
  }

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Workflows</h1>
        <p className="text-muted-foreground">
          Manage your tasks and approvals
        </p>
      </div>

      {/* Tabs */}
      <Tabs value={activeTab} onValueChange={(value) => setActiveTab(value as 'tasks' | 'approvals')} className="space-y-6">
        <TabsList className="grid w-full max-w-md grid-cols-2">
          <TabsTrigger value="tasks">Tasks</TabsTrigger>
          <TabsTrigger value="approvals">Approvals</TabsTrigger>
        </TabsList>

        {/* Tasks Tab */}
        <TabsContent value="tasks" className="space-y-4">
          {/* Task Stats */}
          <TaskStatsCards userId={userId} refreshTrigger={refreshTrigger} />

          {/* Filter Buttons */}
          <div className="flex gap-2">
            <Button
              variant={selectedStatus === 'all' ? 'default' : 'outline'}
              onClick={() => setSelectedStatus('all')}
            >
              All Tasks
            </Button>
            <Button
              variant={selectedStatus === 'PENDING' ? 'default' : 'outline'}
              onClick={() => setSelectedStatus('PENDING')}
            >
              Pending
            </Button>
            <Button
              variant={selectedStatus === 'IN_PROGRESS' ? 'default' : 'outline'}
              onClick={() => setSelectedStatus('IN_PROGRESS')}
            >
              In Progress
            </Button>
          </div>

          {/* Tasks Table */}
          <TasksTable
            userId={userId}
            userRole={userRole}
            refreshTrigger={refreshTrigger}
            status={selectedStatus === 'all' ? undefined : (selectedStatus as 'PENDING' | 'IN_PROGRESS')}
            onTaskAction={handleTaskAction}
          />
        </TabsContent>

        {/* Approvals Tab */}
        <TabsContent value="approvals" className="space-y-4">
          <ApprovalsList userId={userId} />
        </TabsContent>
      </Tabs>
    </div>
  )
}
