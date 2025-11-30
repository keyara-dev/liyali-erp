'use client'

import { useState } from 'react'
import { PlusCircledIcon } from '@radix-ui/react-icons'
import { Button } from '@/components/ui/button'
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
    <div className="space-y-4">
      {/* Page Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold tracking-tight lg:text-2xl">Budgets</h1>
          <p className="text-sm text-muted-foreground mt-1">
            Manage and track departmental budgets
          </p>
        </div>
        <Button onClick={() => setIsCreateDialogOpen(true)}>
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
