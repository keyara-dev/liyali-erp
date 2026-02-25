"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import {
  Banknote,
  User,
  Building2,
  Save,
  Send,
  Edit,
  Trash2,
  TrendingUp,
  AlertCircle,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { StatusBadge } from "@/components/status-badge";
import { PageHeader } from "@/components/base/page-header";
import { useBudgetById } from "@/hooks/use-budget-queries";
import { Budget } from "@/types/budget";
import { ApprovalChainPanel } from "./approval-chain-panel";
import { BudgetApprovalPanel } from "./budget-approval-panel";
import { BudgetItemsManager, BudgetItem } from "./budget-items-manager";
import { BudgetEditDialog } from "./budget-edit-dialog";
import { BudgetDeleteDialog } from "./budget-delete-dialog";
import { BudgetSubmitDialog } from "./budget-submit-dialog";
import {
  updateBudget,
  deleteBudget,
  submitBudget,
} from "@/app/_actions/budgets";
import { toast } from "sonner";
import { useQueryClient } from "@tanstack/react-query";
import { QUERY_KEYS } from "@/lib/constants";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Separator } from "@/components/ui/separator";
import CustomAlert from "@/components/ui/custom-alert";

interface BudgetDetailClientEnhancedProps {
  budgetId: string;
  userId: string;
  userRole: string;
}

export function BudgetDetailClientEnhanced({
  budgetId,
  userId,
  userRole,
}: BudgetDetailClientEnhancedProps) {
  const router = useRouter();
  const queryClient = useQueryClient();
  const { data: budget, isLoading } = useBudgetById(budgetId);

  const [isSubmitting, setIsSubmitting] = useState(false);
  const [budgetItems, setBudgetItems] = useState<BudgetItem[]>([]);
  const [hasUnsavedChanges, setHasUnsavedChanges] = useState(false);

  // Dialog states
  const [isEditDialogOpen, setIsEditDialogOpen] = useState(false);
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
  const [isSubmitDialogOpen, setIsSubmitDialogOpen] = useState(false);

  // Initialize budget items when budget data loads
  useEffect(() => {
    if (budget?.items) {
      console.log("Budget items from API:", budget.items);
      console.log("Budget items type:", typeof budget.items);
      console.log("Budget items is array:", Array.isArray(budget.items));
      console.log("Budget action history:", budget.actionHistory);
      console.log(
        "Budget action history length:",
        budget.actionHistory?.length,
      );

      // Ensure items have the correct structure
      const formattedItems = Array.isArray(budget.items)
        ? budget.items.map((item: any) => ({
            id: item.id || String(Date.now() + Math.random()),
            category: item.category || "",
            description: item.description || "",
            allocatedAmount: Number(item.allocatedAmount || 0),
          }))
        : [];

      console.log("Formatted items:", formattedItems);
      setBudgetItems(formattedItems);
      setHasUnsavedChanges(false);
    }
  }, [budget]);

  const handleItemsChange = (newItems: BudgetItem[]) => {
    setBudgetItems(newItems);
    setHasUnsavedChanges(true);
  };

  const handleSaveBudgetItems = async () => {
    if (!budget) return;

    setIsSubmitting(true);
    try {
      // Calculate new allocated amount from items
      const newAllocatedAmount = budgetItems.reduce(
        (sum, item) => sum + item.allocatedAmount,
        0,
      );

      const response = await updateBudget(budgetId, {
        items: budgetItems,
        allocatedAmount: newAllocatedAmount,
      });

      if (response.success) {
        toast.success("Budget items saved successfully");
        setHasUnsavedChanges(false);

        // Invalidate and refetch budget data to get updated action history
        await queryClient.invalidateQueries({
          queryKey: [QUERY_KEYS.BUDGETS.BY_ID, budgetId],
        });
        await queryClient.refetchQueries({
          queryKey: [QUERY_KEYS.BUDGETS.BY_ID, budgetId],
        });

        // Also invalidate the budgets list
        queryClient.invalidateQueries({
          queryKey: [QUERY_KEYS.BUDGETS.ALL],
        });
      } else {
        toast.error(response.message || "Failed to save budget items");
      }
    } catch (error) {
      console.error("Error saving budget items:", error);
      toast.error("An error occurred while saving");
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleBudgetUpdate = async (data: any) => {
    setIsSubmitting(true);
    try {
      const response = await updateBudget(budgetId, data);
      if (response.success) {
        toast.success("Budget updated successfully");
        setIsEditDialogOpen(false);
        queryClient.invalidateQueries({
          queryKey: [QUERY_KEYS.BUDGETS.BY_ID, budgetId],
        });
      } else {
        toast.error(response.message || "Failed to update budget");
      }
    } catch (error) {
      console.error("Error updating budget:", error);
      toast.error("An error occurred while updating");
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleBudgetDelete = async () => {
    setIsSubmitting(true);
    try {
      const response = await deleteBudget(budgetId);
      if (response.success) {
        toast.success("Budget deleted successfully");
        router.push("/budgets");
      } else {
        toast.error(response.message || "Failed to delete budget");
      }
    } catch (error) {
      console.error("Error deleting budget:", error);
      toast.error("An error occurred while deleting");
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleBudgetSubmit = async (workflowId: string, comments?: string) => {
    setIsSubmitting(true);
    try {
      const response = await submitBudget(budgetId, workflowId, comments);
      if (response.success) {
        toast.success("Budget submitted for approval");
        setIsSubmitDialogOpen(false);
        queryClient.invalidateQueries({
          queryKey: [QUERY_KEYS.BUDGETS.BY_ID, budgetId],
        });
      } else {
        toast.error(response.message || "Failed to submit budget");
      }
    } catch (error) {
      console.error("Error submitting budget:", error);
      toast.error("An error occurred while submitting");
    } finally {
      setIsSubmitting(false);
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-4"></div>
          <p className="text-muted-foreground">Loading budget details...</p>
        </div>
      </div>
    );
  }

  if (!budget) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <div className="text-center">
          <p className="text-muted-foreground">Budget not found</p>
          <Button
            variant="outline"
            onClick={() => router.push("/budgets")}
            className="mt-4"
          >
            Back to Budgets
          </Button>
        </div>
      </div>
    );
  }

  const canEdit = budget.status === "draft" && budget.ownerId === userId;
  const canDelete = budget.status === "draft" && budget.ownerId === userId;
  const canSubmit = budget.status === "draft" && budget.ownerId === userId;
  const canApprove = budget.status === "pending" && userRole === "admin";

  // Calculate budget utilization
  const utilizationPercentage =
    (budget.allocatedAmount / budget.totalBudget) * 100;
  const isOverBudget = budget.allocatedAmount > budget.totalBudget;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between gap-4">
        <PageHeader
          title={`${budget.name} | ${budget.budgetCode}`}
          subtitle={`Budget • ${budget.department} • FY ${budget.fiscalYear} • Created ${new Date(budget.createdAt).toLocaleDateString("en-ZM", { year: "numeric", month: "long", day: "numeric" })}${budget.updatedAt && new Date(budget.updatedAt).getTime() !== new Date(budget.createdAt).getTime() ? ` • Updated ${new Date(budget.updatedAt).toLocaleDateString("en-ZM", { year: "numeric", month: "long", day: "numeric" })}` : ""}`}
          // description={budget.description}
          badges={[
            {
              status: budget.status,
              type: "document",
            },
          ]}
          onBackClick={() => router.replace("/budgets")}
          showBackButton={true}
        />

        <div className="flex gap-2 mt-2">
          {canEdit && hasUnsavedChanges && (
            <Button
              onClick={handleSaveBudgetItems}
              disabled={isSubmitting}
              className="gap-2"
              isLoading={isSubmitting}
              loadingText="Saving..."
            >
              <Save className="h-4 w-4" />
              Save Changes
            </Button>
          )}
          {canSubmit && !hasUnsavedChanges && (
            <Button
              onClick={() => setIsSubmitDialogOpen(true)}
              disabled={isSubmitting || isOverBudget}
              className="gap-2"
            >
              <Send className="h-4 w-4" />
              Submit for Approval
            </Button>
          )}
          {canEdit && (
            <Button
              variant="outline"
              onClick={() => setIsEditDialogOpen(true)}
              className="gap-2"
            >
              <Edit className="h-4 w-4" />
              Edit
            </Button>
          )}
          {canDelete && (
            <Button
              variant="outline"
              onClick={() => setIsDeleteDialogOpen(true)}
              className="gap-2 text-red-600 hover:text-red-700"
            >
              <Trash2 className="h-4 w-4" />
              Delete
            </Button>
          )}
        </div>
      </div>

      {/* Validation Alerts */}
      {hasUnsavedChanges && (
        <CustomAlert type="warning" Icon={AlertCircle}>
          You have unsaved changes. Click "Save Changes" to save your budget
          items.
        </CustomAlert>
      )}

      {isOverBudget && (
        <CustomAlert type="error" Icon={AlertCircle}>
          Budget items exceed total budget by {budget.currency || "K"}
          {Math.abs(budget.remainingAmount).toLocaleString()}. Please adjust
          before submitting.
        </CustomAlert>
      )}

      {/* Budget Overview Cards */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Total Budget
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {budget.currency || "K"}
              {budget.totalBudget.toLocaleString()}
            </div>
            <p className="text-xs text-muted-foreground mt-1">
              Original + Approved Changes
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Allocated
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-blue-600">
              {budget.currency || "K"}
              {budget.allocatedAmount.toLocaleString()}
            </div>
            <p className="text-xs text-muted-foreground mt-1">
              {utilizationPercentage.toFixed(1)}% of total
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Remaining
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div
              className={`text-2xl font-bold ${budget.remainingAmount < 0 ? "text-red-600" : "text-green-600"}`}
            >
              {budget.currency || "K"}
              {budget.remainingAmount.toLocaleString()}
            </div>
            <p className="text-xs text-muted-foreground mt-1">
              Available to allocate
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Utilization
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {utilizationPercentage.toFixed(1)}%
            </div>
            <div className="w-full bg-gray-200 rounded-full h-2 mt-2">
              <div
                className={`h-2 rounded-full transition-all ${
                  utilizationPercentage > 100
                    ? "bg-red-600"
                    : utilizationPercentage >= 90
                      ? "bg-yellow-600"
                      : "bg-green-600"
                }`}
                style={{
                  width: `${Math.min(utilizationPercentage, 100)}%`,
                }}
              />
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Budget Details */}
      <div className="grid gap-6 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Building2 className="h-5 w-5" />
              Budget Information
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <p className="text-sm font-medium text-muted-foreground">
                  Department
                </p>
                <p className="text-sm mt-1">{budget.department}</p>
              </div>
              <div>
                <p className="text-sm font-medium text-muted-foreground">
                  Fiscal Year
                </p>
                <p className="text-sm mt-1">{budget.fiscalYear}</p>
              </div>
              <div>
                <p className="text-sm font-medium text-muted-foreground">
                  Owner
                </p>
                <p className="text-sm mt-1">{budget.ownerName}</p>
              </div>
              <div>
                <p className="text-sm font-medium text-muted-foreground">
                  Currency
                </p>
                <p className="text-sm mt-1">{budget.currency || "ZMW"}</p>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <TrendingUp className="h-5 w-5" />
              Timeline
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <p className="text-sm font-medium text-muted-foreground">
                  Created
                </p>
                <p className="text-sm mt-1">
                  {new Date(budget.createdAt).toLocaleDateString()}
                </p>
              </div>
              <div>
                <p className="text-sm font-medium text-muted-foreground">
                  Last Updated
                </p>
                <p className="text-sm mt-1">
                  {new Date(budget.updatedAt).toLocaleDateString()}
                </p>
              </div>
              <div>
                <p className="text-sm font-medium text-muted-foreground">
                  Status
                </p>
                <div className="mt-1">
                  <StatusBadge status={budget.status} type="document" />
                </div>
              </div>
              <div>
                <p className="text-sm font-medium text-muted-foreground">
                  Approval Stage
                </p>
                <Badge className="text-sm mt-1">
                  {budget.status === "draft"
                    ? "Not submitted"
                    : budget.status === "approved"
                      ? "Completed"
                      : budget.status === "rejected"
                        ? "Rejected"
                        : budget.approvalStage > 0
                          ? `Stage ${budget.approvalStage}`
                          : "Pending"}
                </Badge>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Budget Items Manager */}
      <BudgetItemsManager
        items={budgetItems}
        totalBudget={budget.totalBudget}
        currency={budget.currency || "K"}
        isEditable={canEdit}
        onItemsChange={handleItemsChange}
      />

      {/* Approval Panel - Enhanced with Timeline and Workflow */}
      <BudgetApprovalPanel
        budgetId={budgetId}
        budget={budget as any}
        userRole={userRole}
        actionHistory={budget.actionHistory}
        approvalChain={budget.approvalHistory}
      />

      {/* Dialogs */}
      <BudgetEditDialog
        open={isEditDialogOpen}
        onOpenChange={setIsEditDialogOpen}
        budget={budget}
        onSave={handleBudgetUpdate}
        isSubmitting={isSubmitting}
      />

      <BudgetDeleteDialog
        open={isDeleteDialogOpen}
        onOpenChange={setIsDeleteDialogOpen}
        budgetCode={budget.budgetCode}
        onConfirm={handleBudgetDelete}
        isDeleting={isSubmitting}
      />

      <BudgetSubmitDialog
        open={isSubmitDialogOpen}
        onOpenChange={setIsSubmitDialogOpen}
        budget={budget}
        onSubmit={handleBudgetSubmit}
        isSubmitting={isSubmitting}
      />
    </div>
  );
}
