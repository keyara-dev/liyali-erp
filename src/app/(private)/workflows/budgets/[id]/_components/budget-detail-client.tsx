'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { ArrowLeft, Send, Plus } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { PageHeader } from '@/components/base/page-header'
import { BudgetItemsTable } from './budget-items-table'
import { ApprovalChainPanel } from './approval-chain-panel'
import { AddBudgetItemDialog } from './add-budget-item-dialog'
import { getBudgetById, submitBudgetForApproval } from '@/app/_actions/budgets'
import { Budget, BudgetItem } from '@/types/budget'

interface BudgetDetailClientProps {
  budgetId: string
  userId: string
  userRole: string
}

export function BudgetDetailClient({
  budgetId,
  userId,
  userRole,
}: BudgetDetailClientProps) {
  const router = useRouter()
  const [budget, setBudget] = useState<Budget | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [isAddItemDialogOpen, setIsAddItemDialogOpen] = useState(false)

  useEffect(() => {
    async function fetchBudget() {
      setIsLoading(true)
      try {
        const result = await getBudgetById(budgetId)
        if (result.success && result.data) {
          setBudget(result.data)
        }
      } catch (error) {
        console.error('Failed to fetch budget:', error)
      } finally {
        setIsLoading(false)
      }
    }

    fetchBudget()
  }, [budgetId])

  const handleSubmitForApproval = async () => {
    if (!budget) return

    setIsSubmitting(true)
    try {
      const result = await submitBudgetForApproval({
        budgetId: budget.id,
        submittedBy: userId,
        comments: 'Budget submitted for approval',
      })

      if (result.success && result.data) {
        setBudget(result.data)
      }
    } catch (error) {
      console.error('Failed to submit budget:', error)
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleAddBudgetItem = (newItem: {
    category: string
    description: string
    allocatedAmount: number
  }) => {
    if (!budget) return

    const budgetItem: BudgetItem = {
      id: `item-${Date.now()}-${Math.random()}`,
      category: newItem.category,
      description: newItem.description,
      allocatedAmount: newItem.allocatedAmount,
      spentAmount: 0,
      remainingAmount: newItem.allocatedAmount,
      createdAt: new Date(),
      updatedAt: new Date(),
    }

    setBudget({
      ...budget,
      items: [...budget.items, budgetItem],
    })
  }

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: budget?.currency || 'USD',
    }).format(amount)
  }

  if (isLoading) {
    return (
      <div className="space-y-4">
        <div className="h-8 bg-muted rounded w-48" />
        <div className="space-y-4">
          {[...Array(3)].map((_, i) => (
            <div key={i} className="h-48 bg-muted rounded" />
          ))}
        </div>
      </div>
    )
  }

  if (!budget) {
    return (
      <div className="text-center py-12">
        <p className="text-muted-foreground">Budget not found</p>
        <Button className="mt-4" onClick={() => router.back()}>
          <ArrowLeft className="mr-2 h-4 w-4" />
          Go Back
        </Button>
      </div>
    )
  }

  const totalSpent = budget.items.reduce((sum, item) => sum + item.spentAmount, 0)
  const totalRemaining = budget.totalAmount - totalSpent
  const spentPercentage = (totalSpent / budget.totalAmount) * 100

  return (
    <div className="space-y-6">
      {/* Header */}
      <PageHeader
        title={budget.name}
        subtitle={budget.budgetNumber}
        badges={[
          {
            status: budget.status,
            type: "document",
          },
        ]}
        onBackClick={() => router.back()}
        showBackButton={true}
      />

      {/* Tabs Container */}
      <Tabs defaultValue="overview" className="w-full">
        <TabsList className="grid w-full grid-cols-4">
          <TabsTrigger value="overview">Overview</TabsTrigger>
          <TabsTrigger value="information">Information</TabsTrigger>
          <TabsTrigger value="items">Items</TabsTrigger>
          <TabsTrigger value="approvals">Approvals</TabsTrigger>
        </TabsList>

        {/* Overview Tab */}
        <TabsContent value="overview" className="space-y-4">
          <div className="grid gap-4 md:grid-cols-3">
            {/* Total Budget */}
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium">Total Budget</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{formatCurrency(budget.totalAmount)}</div>
                <p className="text-xs text-muted-foreground mt-1">
                  Fiscal Year {budget.fiscalYear}
                </p>
              </CardContent>
            </Card>

            {/* Spent Amount */}
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium">Spent Amount</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold text-orange-600">{formatCurrency(totalSpent)}</div>
                <div className="w-full bg-gray-200 rounded-full h-2 mt-2">
                  <div
                    className="bg-orange-600 h-2 rounded-full"
                    style={{ width: `${Math.min(spentPercentage, 100)}%` }}
                  />
                </div>
                <p className="text-xs text-muted-foreground mt-1">
                  {spentPercentage.toFixed(1)}% of budget
                </p>
              </CardContent>
            </Card>

            {/* Remaining Amount */}
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium">Remaining Amount</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold text-green-600">{formatCurrency(totalRemaining)}</div>
                <p className="text-xs text-muted-foreground mt-1">
                  Available for allocation
                </p>
              </CardContent>
            </Card>
          </div>

          {/* Actions Section */}
          {budget.status === 'DRAFT' && (
            <Card>
              <CardHeader>
                <CardTitle>Actions</CardTitle>
              </CardHeader>
              <CardContent>
                <Button
                  onClick={handleSubmitForApproval}
                  disabled={isSubmitting}
                  size="lg"
                >
                  <Send className="mr-2 h-4 w-4" />
                  {isSubmitting ? 'Submitting...' : 'Submit for Approval'}
                </Button>
              </CardContent>
            </Card>
          )}
        </TabsContent>

        {/* Information Tab */}
        <TabsContent value="information">
          <Card>
            <CardHeader>
              <CardTitle>Budget Information</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid gap-4 md:grid-cols-2">
                <div>
                  <p className="text-sm text-muted-foreground">Department</p>
                  <p className="font-medium">{budget.department}</p>
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">Fiscal Year</p>
                  <p className="font-medium">{budget.fiscalYear}</p>
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">Created By</p>
                  <p className="font-medium">{budget.createdBy}</p>
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">Created At</p>
                  <p className="font-medium">{budget.createdAt.toLocaleDateString()}</p>
                </div>
              </div>
              {budget.description && (
                <div>
                  <p className="text-sm text-muted-foreground">Description</p>
                  <p className="font-medium">{budget.description}</p>
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        {/* Items Tab */}
        <TabsContent value="items">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between">
              <CardTitle>Budget Items ({budget.items.length})</CardTitle>
              {budget.status === 'DRAFT' && (
                <Button
                  size="sm"
                  onClick={() => setIsAddItemDialogOpen(true)}
                >
                  <Plus className="h-4 w-4 mr-1" />
                  Add Item
                </Button>
              )}
            </CardHeader>
            <CardContent>
              <BudgetItemsTable items={budget.items} currency={budget.currency} />
            </CardContent>
          </Card>
        </TabsContent>

        {/* Approvals Tab */}
        <TabsContent value="approvals">
          <ApprovalChainPanel approvalChain={budget.approvalChain} />
        </TabsContent>
      </Tabs>

      {/* Add Budget Item Dialog */}
      <AddBudgetItemDialog
        open={isAddItemDialogOpen}
        onOpenChange={setIsAddItemDialogOpen}
        onItemAdded={handleAddBudgetItem}
      />
    </div>
  )
}
