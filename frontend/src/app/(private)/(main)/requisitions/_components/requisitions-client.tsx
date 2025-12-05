'use client'

import { useState } from 'react'
import { PlusCircledIcon } from '@radix-ui/react-icons'
import { Button } from '@/components/ui/button'
import { PageHeader } from '@/components/base/page-header'
import { RequisitionsTable } from './requisitions-table'
import { CreateRequisitionDialog } from './create-requisition-dialog'

interface RequisitionsClientProps {
  userId: string
  userRole: string
}

export function RequisitionsClient({
  userId,
  userRole,
}: RequisitionsClientProps) {
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false)
  const [refreshTrigger, setRefreshTrigger] = useState(0)

  const handleRequisitionCreated = () => {
    setIsCreateDialogOpen(false)
    setRefreshTrigger((prev) => prev + 1)
  }

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex items-start justify-between gap-4">
        <PageHeader
          title="Requisitions"
          subtitle="Request and track requisition forms through the approval workflow"
          showBackButton={false}
        />
        <Button onClick={() => setIsCreateDialogOpen(true)} className="mt-2 h-11">
          <PlusCircledIcon className="h-4 w-4" />
          Create Requisition
        </Button>
      </div>

      {/* Requisitions Table */}
      <RequisitionsTable
        userId={userId}
        userRole={userRole}
        refreshTrigger={refreshTrigger}
      />

      {/* Create Dialog */}
      <CreateRequisitionDialog
        open={isCreateDialogOpen}
        onOpenChange={setIsCreateDialogOpen}
        onRequisitionCreated={handleRequisitionCreated}
        userId={userId}
      />
    </div>
  )
}
