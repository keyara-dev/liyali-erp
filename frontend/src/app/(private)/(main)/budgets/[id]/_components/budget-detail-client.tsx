"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { ArrowLeft, Send, Plus, Pencil, Trash2 } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { PageHeader } from "@/components/base/page-header";
import { BudgetItemsTable } from "./budget-items-table";
import { ApprovalChainPanel } from "./approval-chain-panel";
import { BudgetApprovalActionPanel } from "./budget-approval-action-panel";
import { AddBudgetItemDialog } from "./add-budget-item-dialog";
import { EditBudgetItemDialog } from "./edit-budget-item-dialog";
import {
  getBudgetById,
  submitBudgetForApproval,
  updateBudget,
} from "@/app/_actions/budgets";
import { useBudget } from "@/hooks/use-budgets-queries";
import { Budget, BudgetItem } from "@/types/budget";
import {
  validateBudgetForSubmission,
  calculateTotalAllocated,
  calculateTotalSpent,
} from "@/lib/budget-validation";

interface BudgetDetailClientProps {
  budgetId: string;
  userId: string;
  userRole: string;
}

export function BudgetDetailClient({
  budgetId,
  userId,
  userRole,
}: BudgetDetailClientProps) {
  const router = useRouter();
  const { data: budget, isLoading } = useBudget(budgetId);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isAddItemDialogOpen, setIsAddItemDialogOpen] = useState(false);
  const [isEditItemDialogOpen, setIsEditItemDialogOpen] = useState(false);
  const [itemToEdit, setItemToEdit] = useState<BudgetItem | null>(null);
  const [isSubmitConfirmationOpen, setIsSubmitConfirmationOpen] = useState(false);

  // Helper to convert date strings to Date objects
  const normalizeBudgetDates = (budget: Budget): Budget => {
    return {
      ...budget,
      createdAt:
        budget.createdAt instanceof Date
          ? budget.createdAt
          : new Date(budget.createdAt),
      updatedAt:
        budget.updatedAt instanceof Date
          ? budget.updatedAt
          : new Date(budget.updatedAt),
      submittedAt:
        budget.submittedAt instanceof Date
          ? budget.submittedAt
          : budget.submittedAt
            ? new Date(budget.submittedAt)
            : undefined,
      approvedAt:
        budget.approvedAt instanceof Date
          ? budget.approvedAt
          : budget.approvedAt
            ? new Date(budget.approvedAt)
            : undefined,
      rejectedAt:
        budget.rejectedAt instanceof Date
          ? budget.rejectedAt
          : budget.rejectedAt
            ? new Date(budget.rejectedAt)
            : undefined,
      items: budget.items.map((item) => ({
        ...item,
        createdAt:
          item.createdAt instanceof Date
            ? item.createdAt
            : new Date(item.createdAt),
        updatedAt:
          item.updatedAt instanceof Date
            ? item.updatedAt
            : new Date(item.updatedAt),
      })),
    };
  };

  useEffect(() => {
    async function fetchBudget() {
      setIsLoading(true);
      try {
        const result = await getBudgetById(budgetId);
        if (result.success && result.data) {
          setBudget(normalizeBudgetDates(result.data));
        } else {
          // Fallback to localStorage if server fetch fails
          // This allows access to newly created budgets before they're in the database
          const localBudget = loadOneFromStorage(budgetId);
          if (localBudget) {
            setBudget(normalizeBudgetDates(localBudget));
          }
        }
      } catch (error) {
        console.error("Failed to fetch budget:", error);
        // Try localStorage as fallback
        const localBudget = loadOneFromStorage(budgetId);
        if (localBudget) {
          setBudget(normalizeBudgetDates(localBudget));
        }
      } finally {
        setIsLoading(false);
      }
    }

    fetchBudget();
  }, [budgetId, loadOneFromStorage]);

  const handleConfirmSubmitForApproval = async () => {
    if (!budget) return;

    // Validate budget before submission
    const validation = validateBudgetForSubmission(budget);
    if (!validation.valid) {
      toast.error(validation.error || "Cannot submit budget");
      setIsSubmitConfirmationOpen(false);
      return;
    }

    setIsSubmitting(true);
    try {
      const result = await submitBudgetForApproval({
        budgetId: budget.id,
        submittingUserId: userId,
      });

      if (result.success) {
        // Merge server response with current budget to preserve all data
        const submittedBudget: Budget = {
          ...budget,
          ...(result.data || {}),
          status: "SUBMITTED",
          submittedAt: new Date(),
          updatedAt: new Date(),
          currentApprovalStage: 1,
          approvalChain: result.data?.approvalChain || [
            {
              stageNumber: 1,
              stageName: "Department Head Review",
              assignedTo: "manager-1",
              assignedRole: "DEPARTMENT_MANAGER",
              status: "PENDING"
            }
          ]
        };

        const normalizedBudget = normalizeBudgetDates(submittedBudget);
        setBudget(normalizedBudget);
        saveToStorage(normalizedBudget);
        toast.success("Budget submitted for approval");
        setIsSubmitConfirmationOpen(false);
      } else {
        toast.error(result.message || "Failed to submit budget");
      }
    } catch (error) {
      console.error("Failed to submit budget:", error);
      toast.error("An error occurred while submitting the budget");
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleAddBudgetItem = async (newItem: {
    category: string;
    description: string;
    allocatedAmount: number;
    spentAmount: number;
  }) => {
    if (!budget) return;

    const budgetItem: BudgetItem = {
      id: `item-${Date.now()}-${Math.random()}`,
      category: newItem.category,
      description: newItem.description,
      allocatedAmount: newItem.allocatedAmount,
      spentAmount: newItem.spentAmount,
      remainingAmount: newItem.allocatedAmount - newItem.spentAmount,
      createdAt: new Date(),
      updatedAt: new Date(),
    };

    const updatedBudget: Budget = {
      ...budget,
      items: [...budget.items, budgetItem],
      updatedAt: new Date(),
    };

    // Update local state
    setBudget(updatedBudget);

    // Persist to storage
    saveToStorage(updatedBudget);

    // Also update via server action
    try {
      const result = await updateBudget(budget.id, {
        items: updatedBudget.items,
      });
      if (result.success) {
        toast.success("Budget item added and saved");
      }
    } catch (error) {
      console.error("Failed to persist budget item:", error);
      toast.error("Budget item added locally but failed to save");
    }
  };

  const handleEditBudgetItem = async (updatedItem: BudgetItem) => {
    if (!budget) return;

    const updatedBudget: Budget = {
      ...budget,
      items: budget.items.map((item) =>
        item.id === updatedItem.id ? updatedItem : item
      ),
      updatedAt: new Date(),
    };

    // Update local state
    setBudget(updatedBudget);

    // Persist to storage
    saveToStorage(updatedBudget);

    // Also update via server action
    try {
      const result = await updateBudget(budget.id, {
        items: updatedBudget.items,
      });
      if (result.success) {
        toast.success("Budget item updated and saved");
      }
    } catch (error) {
      console.error("Failed to persist budget item update:", error);
      toast.error("Budget item updated locally but failed to save");
    }

    setIsEditItemDialogOpen(false);
    setItemToEdit(null);
  };

  const handleDeleteBudgetItem = async (itemId: string) => {
    if (!budget) return;

    const updatedBudget: Budget = {
      ...budget,
      items: budget.items.filter((item) => item.id !== itemId),
      updatedAt: new Date(),
    };

    // Update local state
    setBudget(updatedBudget);

    // Persist to storage
    saveToStorage(updatedBudget);

    // Also update via server action
    try {
      const result = await updateBudget(budget.id, {
        items: updatedBudget.items,
      });
      if (result.success) {
        toast.success("Budget item deleted and saved");
      }
    } catch (error) {
      console.error("Failed to persist budget item deletion:", error);
      toast.error("Budget item deleted locally but failed to save");
    }
  };

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat("en-US", {
      style: "currency",
      currency: budget?.currency || "USD",
    }).format(amount);
  };

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
    );
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
    );
  }

  const totalSpent = budget.items.reduce(
    (sum, item) => sum + item.spentAmount,
    0
  );
  const totalRemaining = budget.totalAmount - totalSpent;
  const spentPercentage = (totalSpent / budget.totalAmount) * 100;

  return (
    <div className="space-y-6">
      {/* Header */}

      <div className="flex items-center justify-between gap-4">
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
        <div>
          {/* Actions Section */}
          {budget.status === "DRAFT" && (
            <>
              <Button
                onClick={() => setIsSubmitConfirmationOpen(true)}
                disabled={isSubmitting}
                size="lg"
                className="h-11"
              >
                <Send className="mr-2 h-4 w-4" />
                {isSubmitting ? "Submitting..." : "Submit for Approval"}
              </Button>

              {/* Submit Confirmation Dialog */}
              <AlertDialog open={isSubmitConfirmationOpen} onOpenChange={setIsSubmitConfirmationOpen}>
                <AlertDialogContent>
                  <AlertDialogHeader>
                    <AlertDialogTitle>Submit Budget for Approval?</AlertDialogTitle>
                    <AlertDialogDescription>
                      You are about to submit this budget for approval. Once submitted, it will move through the approval chain. You won't be able to edit items after submission unless it's rejected.
                    </AlertDialogDescription>
                  </AlertDialogHeader>
                  <AlertDialogFooter>
                    <AlertDialogCancel disabled={isSubmitting}>Cancel</AlertDialogCancel>
                    <AlertDialogAction
                      onClick={handleConfirmSubmitForApproval}
                      disabled={isSubmitting}
                      className="bg-blue-600 hover:bg-blue-700"
                    >
                      {isSubmitting ? "Submitting..." : "Submit"}
                    </AlertDialogAction>
                  </AlertDialogFooter>
                </AlertDialogContent>
              </AlertDialog>
            </>
          )}
        </div>
      </div>

      {/* Tabs Container */}
      <Tabs defaultValue="overview" className="w-full">
        <TabsList className="grid w-full grid-cols-3">
          <TabsTrigger value="overview">Overview</TabsTrigger>
          <TabsTrigger value="items">Items</TabsTrigger>
          <TabsTrigger value="approvals">Approvals</TabsTrigger>
        </TabsList>

        {/* Overview Tab */}
        <TabsContent value="overview" className="space-y-4">
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
                  <p className="font-medium">
                    {budget.createdAt.toLocaleDateString()}
                  </p>
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
          <div className="grid gap-4 md:grid-cols-3">
            {/* Total Budget */}
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium">
                  Total Budget
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">
                  {formatCurrency(budget.totalAmount)}
                </div>
                <p className="text-xs text-muted-foreground mt-1">
                  Fiscal Year {budget.fiscalYear}
                </p>
              </CardContent>
            </Card>

            {/* Spent Amount */}
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium">
                  Spent Amount
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold text-orange-600">
                  {formatCurrency(totalSpent)}
                </div>
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
                <CardTitle className="text-sm font-medium">
                  Remaining Amount
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold text-green-600">
                  {formatCurrency(totalRemaining)}
                </div>
                <p className="text-xs text-muted-foreground mt-1">
                  Available for allocation
                </p>
              </CardContent>
            </Card>
          </div>

          {/* Approval Action Panel */}
          {(budget.status === "IN_REVIEW" || budget.status === "SUBMITTED") && (
            <BudgetApprovalActionPanel
              budgetId={budget.id}
              budgetStatus={budget.status}
              budget={budget}
              onApprovalComplete={() => {
                // Refresh budget after approval/rejection
                setIsLoading(true);
                const updatedBudget = loadOneFromStorage(budget.id);
                if (updatedBudget) {
                  setBudget(normalizeBudgetDates(updatedBudget));
                }
                setIsLoading(false);
              }}
            />
          )}
        </TabsContent>

        {/* Items Tab */}
        <TabsContent value="items">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between">
              <CardTitle>Budget Items ({budget.items.length})</CardTitle>
              {budget.status === "DRAFT" && (
                <Button size="sm" onClick={() => setIsAddItemDialogOpen(true)}>
                  <Plus className="h-4 w-4 mr-1" />
                  Add Item
                </Button>
              )}
            </CardHeader>
            <CardContent>
              <BudgetItemsTable
                items={budget.items}
                currency={budget.currency}
                status={budget.status}
                onEditItem={(item) => {
                  setItemToEdit(item);
                  setIsEditItemDialogOpen(true);
                }}
                onDeleteItem={handleDeleteBudgetItem}
              />
            </CardContent>
          </Card>
        </TabsContent>

        {/* Approvals Tab */}
        <TabsContent value="approvals" className="space-y-6">
          {/* Approval Action Panel - Only shown if user can approve/reject */}
          {(budget.status === "IN_REVIEW" || budget.status === "SUBMITTED") && (
            <BudgetApprovalActionPanel
              budgetId={budget.id}
              budgetStatus={budget.status}
              budget={budget}
              onApprovalComplete={() => {
                // Refresh budget after approval/rejection
                setIsLoading(true);
                const updatedBudget = loadOneFromStorage(budget.id);
                if (updatedBudget) {
                  setBudget(normalizeBudgetDates(updatedBudget));
                }
                setIsLoading(false);
              }}
            />
          )}

          {/* Approval Chain History */}
          <ApprovalChainPanel approvalChain={budget?.approvalChain as any} />
        </TabsContent>
      </Tabs>

      {/* Add Budget Item Dialog */}
      <AddBudgetItemDialog
        open={isAddItemDialogOpen}
        onOpenChange={setIsAddItemDialogOpen}
        onItemAdded={handleAddBudgetItem}
        existingItems={budget.items}
        totalBudget={budget.totalAmount}
        currency={budget.currency}
      />

      {/* Edit Budget Item Dialog */}
      <EditBudgetItemDialog
        open={isEditItemDialogOpen}
        onOpenChange={setIsEditItemDialogOpen}
        onItemUpdated={handleEditBudgetItem}
        existingItems={budget.items}
        itemToEdit={itemToEdit}
        totalBudget={budget.totalAmount}
        currency={budget.currency}
      />
    </div>
  );
}
