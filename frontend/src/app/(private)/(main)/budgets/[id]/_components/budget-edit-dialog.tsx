"use client";

import { useState, useEffect } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { SelectField } from "@/components/ui/select-field";
import { Label } from "@/components/ui/label";
import { Budget } from "@/types/budget";
import { useActiveDepartments } from "@/hooks/use-department-queries";
import { Loader2 } from "lucide-react";

interface BudgetEditDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  budget: Budget;
  onSave: (data: any) => Promise<void>;
  isSubmitting: boolean;
}

const currencies = [
  { code: "ZMW", label: "Zambian Kwacha (K)" },
  { code: "USD", label: "US Dollar ($)" },
  { code: "EUR", label: "Euro (€)" },
  { code: "GBP", label: "British Pound (£)" },
];

export function BudgetEditDialog({
  open,
  onOpenChange,
  budget,
  onSave,
  isSubmitting,
}: BudgetEditDialogProps) {
  const { data: departments = [], isLoading: isLoadingDepartments } =
    useActiveDepartments();

  const [formData, setFormData] = useState({
    name: budget.name || "",
    description: budget.description || "",
    department: budget.department || "",
    departmentId: budget.departmentId || "",
    fiscalYear: budget.fiscalYear || "",
    totalBudget: budget.totalBudget.toString(),
    currency: budget.currency || "ZMW",
  });

  // Reset form when budget changes
  useEffect(() => {
    setFormData({
      name: budget.name || "",
      description: budget.description || "",
      department: budget.department || "",
      departmentId: budget.departmentId || "",
      fiscalYear: budget.fiscalYear || "",
      totalBudget: budget.totalBudget.toString(),
      currency: budget.currency || "ZMW",
    });
  }, [budget]);

  const handleInputChange = (field: string, value: string | number) => {
    setFormData((prev) => ({
      ...prev,
      [field]: value,
    }));
  };

  const handleDepartmentChange = (departmentId: string) => {
    const dept = departments.find((d) => d.id === departmentId);
    setFormData((prev) => ({
      ...prev,
      departmentId,
      department: dept?.name || "",
    }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    await onSave({
      name: formData.name,
      description: formData.description,
      department: formData.department,
      departmentId: formData.departmentId,
      fiscalYear: formData.fiscalYear,
      totalBudget: parseFloat(formData.totalBudget),
      currency: formData.currency,
    });
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Edit Budget</DialogTitle>
          <DialogDescription>
            Update budget information. Budget items are managed separately.
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          {/* Budget Name */}
          <div className="space-y-2">
            <Label htmlFor="name">Budget Name *</Label>
            <Input
              id="name"
              required
              placeholder="e.g., IT Department Annual Budget 2024"
              value={formData.name}
              onChange={(e) => handleInputChange("name", e.target.value)}
              disabled={isSubmitting}
            />
          </div>

          {/* Description */}
          <div className="space-y-2">
            <Label htmlFor="description">Description</Label>
            <Textarea
              id="description"
              placeholder="Add a description for this budget (optional)"
              value={formData.description}
              onChange={(e) => handleInputChange("description", e.target.value)}
              disabled={isSubmitting}
              rows={3}
            />
          </div>

          {/* Department */}
          <div className="space-y-2">
            <Label htmlFor="department">Department *</Label>
            <SelectField
              placeholder={
                isLoadingDepartments
                  ? "Loading departments..."
                  : "Select a department"
              }
              value={formData.departmentId}
              onValueChange={handleDepartmentChange}
              disabled={isSubmitting || isLoadingDepartments}
              options={departments.map((dept) => ({
                value: dept.id,
                label: dept.name,
              }))}
            />
            {isLoadingDepartments && (
              <div className="flex items-center gap-2 text-sm text-muted-foreground">
                <Loader2 className="h-3 w-3 animate-spin" />
                <span>Loading departments...</span>
              </div>
            )}
          </div>

          {/* Fiscal Year */}
          <div className="space-y-2">
            <Label htmlFor="fiscalYear">Fiscal Year *</Label>
            <Input
              id="fiscalYear"
              required
              type="number"
              placeholder="2024"
              value={formData.fiscalYear}
              onChange={(e) => handleInputChange("fiscalYear", e.target.value)}
              disabled={isSubmitting}
            />
          </div>

          {/* Total Budget and Currency */}
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="totalBudget">Total Budget *</Label>
              <Input
                id="totalBudget"
                required
                type="number"
                placeholder="0.00"
                step="0.01"
                value={formData.totalBudget}
                onChange={(e) =>
                  handleInputChange("totalBudget", e.target.value)
                }
                disabled={isSubmitting}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="currency">Currency *</Label>
              <SelectField
                value={formData.currency}
                onValueChange={(value) => handleInputChange("currency", value)}
                disabled={isSubmitting}
                options={currencies.map((curr) => ({
                  value: curr.code,
                  label: curr.label,
                }))}
              />
            </div>
          </div>

          {/* Actions */}
          <DialogFooter className="gap-2 sm:gap-0">
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={isSubmitting}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={isSubmitting} isLoading={isSubmitting} loadingText="Saving...">
              Save Changes
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
