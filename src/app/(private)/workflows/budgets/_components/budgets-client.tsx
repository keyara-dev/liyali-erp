'use client'

import { useState } from 'react'
import { PlusCircledIcon } from '@radix-ui/react-icons'
import { Button } from '@/components/ui/button'
import { PageHeader } from '@/components/base/page-header'
import { BudgetsTable } from './budgets-table'
import { CreateBudgetDialog } from './create-budget-dialog'

interface BudgetsClientProps {
  userId: string
  userRole: string
}

export function BudgetsClient({
  userId,
  userRole,
}: BudgetsClientProps) {
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false)
  const [refreshTrigger, setRefreshTrigger] = useState(0)

  const handleBudgetCreated = () => {
    setIsCreateDialogOpen(false)
    setRefreshTrigger((prev) => prev + 1)
  }

  const handleBudgetAction = () => {
    setRefreshTrigger((prev) => prev + 1)
  }

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex items-start justify-between gap-4">
        <PageHeader
          title="Budgets"
          subtitle="Manage and track departmental budgets"
          showBackButton={false}
        />
        <Button onClick={() => setIsCreateDialogOpen(true)} className="mt-2 h-11">
          <PlusCircledIcon className="h-4 w-4" />
          Create Budget
        </Button>
      </div>

      {/* Budgets Table */}
      <BudgetsTable
        userId={userId}
        userRole={userRole}
        refreshTrigger={refreshTrigger}
        onBudgetAction={handleBudgetAction}
      />

      {/* Create Budget Dialog */}
      <CreateBudgetDialog
        open={isCreateDialogOpen}
        onOpenChange={setIsCreateDialogOpen}
        onBudgetCreated={handleBudgetCreated}
        userId={userId}
      />
    </div>
  )
}
