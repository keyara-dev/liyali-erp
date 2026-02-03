'use client'

import { useState } from 'react'
import { PlusCircledIcon } from '@radix-ui/react-icons'
import { Button } from '@/components/ui/button'
import { PageHeader } from '@/components/base/page-header'
import { RequisitionsTable } from './requisitions-table'
import { CreateRequisitionDialog } from './create-requisition-dialog'
import { Requisition } from '@/types/requisition'

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
  const [editingRequisition, setEditingRequisition] = useState<Requisition | null>(null)
  const [isEditing, setIsEditing] = useState(false)

  const handleRequisitionCreated = () => {
    setIsCreateDialogOpen(false)
    setIsEditing(false)
    setEditingRequisition(null)
    setRefreshTrigger((prev) => prev + 1)
  }

  const handleCreateNew = () => {
    setIsEditing(false)
    setEditingRequisition(null)
    setIsCreateDialogOpen(true)
  }

  const handleEditRequisition = (requisition: Requisition) => {
    setIsEditing(true)
    setEditingRequisition(requisition)
    setIsCreateDialogOpen(true)
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
        <Button onClick={handleCreateNew} className="mt-2 h-11">
          <PlusCircledIcon className="h-4 w-4" />
          Create Requisition
        </Button>
      </div>

      {/* Requisitions Table */}
      <RequisitionsTable
        userId={userId}
        userRole={userRole}
        refreshTrigger={refreshTrigger}
        onEditRequisition={handleEditRequisition}
        onCreateRequisition={handleCreateNew}
      />

      {/* Create/Edit Dialog */}
      <CreateRequisitionDialog
        open={isCreateDialogOpen}
        onOpenChange={setIsCreateDialogOpen}
        onRequisitionCreated={handleRequisitionCreated}
        userId={userId}
        editingRequisition={editingRequisition}
        isEditing={isEditing}
      />
    </div>
  )
}
