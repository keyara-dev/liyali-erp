"use client";

import { useState, useEffect } from "react";
import { toast } from "sonner";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { SelectField } from "@/components/ui/select-field";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { InfoIcon } from "lucide-react";
import { createBudget } from "@/app/_actions/budgets";
import { useBudgetStorage } from "@/hooks/use-budget-storage";

interface CreateBudgetDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onBudgetCreated: () => void;
}

const departments = [
  { id: "dept-it", label: "Information Technology" },
  { id: "dept-hr", label: "Human Resources" },
  { id: "dept-ops", label: "Operations" },
  { id: "dept-sales", label: "Sales" },
  { id: "dept-marketing", label: "Marketing" },
  { id: "dept-finance", label: "Finance" },
];

const currencies = [
  { code: "ZMW", label: "Zambian Kwacha (K)" },
  { code: "USD", label: "US Dollar ($)" },
];

// Generate automatic budget code
const generateBudgetCode = () => {
  const year = new Date().getFullYear();
  const randomChars = Math.random().toString(36).substring(2, 10).toUpperCase();
  return `BG-${year}-${randomChars}`;
};

export function CreateBudgetDialog({
  open,
  onOpenChange,
  onBudgetCreated,
}: CreateBudgetDialogProps) {
  const [isSubmitting, setIsSubmitting] = useState(false);
  const { saveToStorage } = useBudgetStorage();
  const [formData, setFormData] = useState({
    name: "",
    description: "",
    department: "",
    departmentId: "",
    fiscalYear: new Date().getFullYear().toString(),
    totalAmount: "",
    currency: "ZMW",
    budgetCode: "", // Will be auto-generated
  });

  // Generate budget code when dialog opens
  useEffect(() => {
    if (open && !formData.budgetCode) {
      setFormData((prev) => ({
        ...prev,
        budgetCode: generateBudgetCode(),
      }));
    }
  }, [open, formData.budgetCode]);

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
      department: dept?.label || "",
    }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (
      !formData.name ||
      !formData.department ||
      !formData.totalAmount ||
      !formData.fiscalYear
    ) {
      toast.error("Please fill in all required fields");
      return;
    }

    setIsSubmitting(true);
    try {
      const result = await createBudget({
        budgetCode: formData.budgetCode,
        name: formData.name,
        description: formData.description,
        department: formData.department,
        departmentId: formData.departmentId,
        fiscalYear: formData.fiscalYear,
        totalBudget: parseFloat(formData.totalAmount),
        allocatedAmount: parseFloat(formData.totalAmount),
        currency: formData.currency,
        // createdBy is automatically handled by the backend from the authenticated user
      });

      if (result.success && result.data) {
        // Store the new budget in localStorage (single source of truth)
        saveToStorage(result.data);

        // Show success toast
        toast.success(`Budget "${formData.name}" created successfully`);

        // Reset form
        setFormData({
          name: "",
          description: "",
          department: "",
          departmentId: "",
          fiscalYear: new Date().getFullYear().toString(),
          totalAmount: "",
          currency: "ZMW",
          budgetCode: "", // Will be regenerated when dialog opens again
        });

        // Close modal and notify parent after a brief delay to ensure toast is visible
        setTimeout(() => {
          onOpenChange(false);
          onBudgetCreated();
        }, 300);
      } else {
        // Show error toast but keep modal open
        toast.error(result.message || "Failed to create budget");
      }
    } catch (error) {
      console.error("Error creating budget:", error);
      // Show error toast but keep modal open
      toast.error("An error occurred while creating the budget");
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>Create New Budget</DialogTitle>
          <DialogDescription>
            Create a new budget for your department. You can add budget items
            later.
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          {/* Alert about budget items */}
          <Alert>
            <InfoIcon className="h-4 w-4" />
            <AlertDescription>
              Budget items will be added after the budget is successfully
              created.
            </AlertDescription>
          </Alert>

          {/* Budget Code (Auto-generated) */}
          <Input
            label="Budget Code"
            value={formData.budgetCode}
            disabled={true}
            placeholder="Auto-generated"
            descriptionText="This code is automatically generated"
          />

          {/* Budget Name */}
          <Input
            label="Budget Name"
            required
            placeholder="e.g., IT Department Annual Budget 2024"
            value={formData.name}
            onChange={(e) => handleInputChange("name", e.target.value)}
            disabled={isSubmitting}
          />

          {/* Description */}
          <Textarea
            label="Description"
            placeholder="Add a description for this budget (optional)"
            value={formData.description}
            onChange={(e) => handleInputChange("description", e.target.value)}
            disabled={isSubmitting}
            rows={3}
          />

          {/* Department */}
          <SelectField
            label="Department"
            required
            placeholder="Select a department"
            value={formData.departmentId}
            onValueChange={handleDepartmentChange}
            disabled={isSubmitting}
            options={departments.map((dept) => ({
              value: dept.id,
              label: dept.label,
            }))}
          />

          {/* Fiscal Year */}
          <Input
            label="Fiscal Year"
            required
            type="number"
            placeholder="2024"
            value={formData.fiscalYear}
            onChange={(e) => handleInputChange("fiscalYear", e.target.value)}
            disabled={isSubmitting}
          />

          {/* Total Amount and Currency */}
          <div className="grid grid-cols-2 gap-4">
            <Input
              label="Total Amount"
              required
              type="number"
              placeholder="0.00"
              step="0.01"
              value={formData.totalAmount}
              onChange={(e) => handleInputChange("totalAmount", e.target.value)}
              disabled={isSubmitting}
            />

            <SelectField
              label="Currency"
              required
              value={formData.currency}
              onValueChange={(value) => handleInputChange("currency", value)}
              disabled={isSubmitting}
              options={currencies.map((curr) => ({
                value: curr.code,
                label: curr.label,
              }))}
            />
          </div>

          {/* Actions */}
          <div className="flex gap-2 pt-4">
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={isSubmitting}
              className="flex-1"
            >
              Cancel
            </Button>
            <Button type="submit" disabled={isSubmitting} className="flex-1">
              {isSubmitting ? "Creating..." : "Create Budget"}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}
