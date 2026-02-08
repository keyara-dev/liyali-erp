"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import {
  ArrowLeft,
  Calendar,
  Banknote,
  User,
  Building2,
  Save,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { StatusBadge } from "@/components/status-badge";
import { useBudgetById } from "@/hooks/use-budget-queries";
import { Budget } from "@/types/budget";
import { ApprovalChainPanel } from "./approval-chain-panel";
import { BudgetItemsManager, BudgetItem } from "./budget-items-manager";
import { updateBudget } from "@/app/_actions/budgets";
import { toast } from "sonner";
import { useQueryClient } from "@tanstack/react-query";
import { QUERY_KEYS } from "@/lib/constants";

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
  const queryClient = useQueryClient();
  const { data: budget, isLoading } = useBudgetById(budgetId);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [budgetItems, setBudgetItems] = useState<BudgetItem[]>([]);
  const [hasUnsavedChanges, setHasUnsavedChanges] = useState(false);

  // Initialize budget items when budget data loads
  useEffect(() => {
    if (budget?.items) {
      setBudgetItems(budget.items);
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
        // Refresh budget data
        queryClient.invalidateQueries({
          queryKey: [QUERY_KEYS.BUDGETS.BY_ID, budgetId],
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
        </div>
      </div>
    );
  }

  const canEdit = budget.status === "draft" && budget.ownerId === userId;
  const canApprove = budget.status === "pending" && userRole === "admin";

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">
            Budget: {budget.budgetCode}
          </h1>
          <p className="text-muted-foreground">
            {budget.department} - {budget.fiscalYear}
          </p>
        </div>
        <div className="flex gap-2">
          {canEdit && hasUnsavedChanges && (
            <Button
              onClick={handleSaveBudgetItems}
              disabled={isSubmitting}
              className="gap-2"
            >
              <Save className="h-4 w-4" />
              {isSubmitting ? "Saving..." : "Save Changes"}
            </Button>
          )}
          <Button
            variant="outline"
            onClick={() => router.back()}
            className="gap-2"
          >
            <ArrowLeft className="h-4 w-4" />
            Back
          </Button>
        </div>
      </div>

      <div className="grid gap-6 md:grid-cols-2">
        {/* Budget Overview */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Banknote className="h-5 w-5" />
              Budget Overview
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <p className="text-sm font-medium text-muted-foreground">
                  Status
                </p>
                <StatusBadge status={budget.status} type="document" />
              </div>
              <div>
                <p className="text-sm font-medium text-muted-foreground">
                  Fiscal Year
                </p>
                <p className="text-sm">{budget.fiscalYear}</p>
              </div>
            </div>

            <div className="space-y-3">
              <div className="flex justify-between">
                <span className="text-sm font-medium">Total Budget:</span>
                <span className="text-sm font-mono">
                  ${budget.totalBudget.toLocaleString()}
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-sm font-medium">Allocated:</span>
                <span className="text-sm font-mono">
                  ${budget.allocatedAmount.toLocaleString()}
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-sm font-medium">Remaining:</span>
                <span className="text-sm font-mono">
                  ${budget.remainingAmount.toLocaleString()}
                </span>
              </div>
            </div>

            {/* Progress Bar */}
            <div className="space-y-2">
              <div className="flex justify-between text-xs">
                <span>Allocation Progress</span>
                <span>
                  {(
                    (budget.allocatedAmount / budget.totalBudget) *
                    100
                  ).toFixed(1)}
                  %
                </span>
              </div>
              <div className="w-full bg-gray-200 rounded-full h-2">
                <div
                  className="bg-blue-600 h-2 rounded-full"
                  style={{
                    width: `${Math.min((budget.allocatedAmount / budget.totalBudget) * 100, 100)}%`,
                  }}
                />
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Budget Details */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Building2 className="h-5 w-5" />
              Budget Details
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div>
              <p className="text-sm font-medium text-muted-foreground">
                Department
              </p>
              <p className="text-sm">{budget.department}</p>
            </div>
            <div>
              <p className="text-sm font-medium text-muted-foreground">Owner</p>
              <p className="text-sm">{budget.ownerName}</p>
            </div>
            <div>
              <p className="text-sm font-medium text-muted-foreground">
                Created
              </p>
              <p className="text-sm">
                {new Date(budget.createdAt).toLocaleDateString()}
              </p>
            </div>
            <div>
              <p className="text-sm font-medium text-muted-foreground">
                Last Updated
              </p>
              <p className="text-sm">
                {new Date(budget.updatedAt).toLocaleDateString()}
              </p>
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

      {/* Approval Chain */}
      {budget.approvalHistory && budget.approvalHistory.length > 0 && (
        <ApprovalChainPanel approvalChain={budget.approvalHistory} />
      )}

      {/* Actions */}
      {(canEdit || canApprove) && (
        <Card>
          <CardHeader>
            <CardTitle>Actions</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex gap-2">
              {canEdit && <Button variant="outline">Edit Budget</Button>}
              {canApprove && <Button>Approve Budget</Button>}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
