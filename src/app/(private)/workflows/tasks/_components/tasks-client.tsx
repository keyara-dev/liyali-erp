'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { TasksTable } from './tasks-table'
import { TaskStatsCards } from './task-stats-cards'

interface TasksClientProps {
  userId: string
  userRole: string
}

export function TasksClient({
  userId,
  userRole,
}: TasksClientProps) {
  const [refreshTrigger, setRefreshTrigger] = useState(0)
  const [selectedStatus, setSelectedStatus] = useState<'all' | 'PENDING' | 'IN_PROGRESS'>('all')

  const handleTaskAction = () => {
    setRefreshTrigger((prev) => prev + 1)
  }

  return (
    <div className="space-y-4">
      {/* Page Header */}
      <div>
        <h1 className="text-xl font-bold tracking-tight lg:text-2xl">Tasks</h1>
        <p className="text-sm text-muted-foreground mt-1">
          Review and act on workflow tasks assigned to you
        </p>
      </div>

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
    </div>
  )
}
