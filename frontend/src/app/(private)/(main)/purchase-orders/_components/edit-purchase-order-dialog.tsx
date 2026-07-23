"use client";

import { useState, useEffect } from "react";
import { toast } from "sonner";
import { ResponsiveSheet } from "@/components/ui/responsive-sheet";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { SelectField } from "@/components/ui/select-field";
import { Loader2, Save } from "lucide-react";
import { PurchaseOrder } from "@/types/purchase-order";
import { useUpdatePurchaseOrder } from "@/hooks/use-purchase-order-mutations";
import { useActiveDepartments } from "@/hooks/use-department-queries";
import { useAllBudgets } from "@/hooks/use-budget-queries";
import { useVendors } from "@/hooks/use-vendor-queries";
import { DatePicker } from "@/components/ui/date-picker";

interface EditPurchaseOrderDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  purchaseOrder: PurchaseOrder;
  onSuccess: () => void;
}

export function EditPurchaseOrderDialog({
  open,
  onOpenChange,
  purchaseOrder,
  onSuccess,
}: EditPurchaseOrderDialogProps) {
  const updateMutation = useUpdatePurchaseOrder(() => {
    toast.success("Purchase order updated successfully");
    onSuccess();
    onOpenChange(false);
  });

  // Fetch departments for the dropdown
  const { data: departments = [], isLoading: departmentsLoading } =
    useActiveDepartments();

  // Fetch budgets for the dropdown
  const { data: budgets = [], isLoading: budgetsLoading } = useAllBudgets();

  // Fetch vendors for the vendor selector (DRAFT only)
  const { data: vendors = [], isLoading: vendorsLoading } = useVendors({
    active: true,
  });

  const [formData, setFormData] = useState({
    title: "",
    description: "",
    department: "",
    departmentId: "",
    priority: "MEDIUM",
    budgetCode: "",
    costCenter: "",
    projectCode: "",
    deliveryDate: null as Date | null,
    vendorId: "",
    vendorName: "",
  });

  // Populate form when dialog opens
  useEffect(() => {
    if (open && purchaseOrder) {
      setFormData({
        title: purchaseOrder.title || "",
        description: purchaseOrder.description || "",
        department: purchaseOrder.department || "",
        departmentId: purchaseOrder.departmentId || "",
        priority: purchaseOrder.priority || "MEDIUM",
        budgetCode: purchaseOrder.budgetCode || "",
        costCenter: purchaseOrder.costCenter || "",
        projectCode: purchaseOrder.projectCode || "",
        deliveryDate: purchaseOrder.deliveryDate
          ? new Date(purchaseOrder.deliveryDate)
          : null,
        vendorId: purchaseOrder.vendorId || "",
        vendorName: purchaseOrder.vendorName || "",
      });
    }
  }, [open, purchaseOrder]);

  const handleSubmit = async () => {
    // Validation
    if (!formData.title.trim()) {
      toast.error("Please enter a title");
      return;
    }
    if (!formData.department.trim()) {
      toast.error("Please select a department");
      return;
    }

    updateMutation.mutate({
      purchaseOrderId: purchaseOrder.id,
      poId: purchaseOrder.id,
      title: formData.title,
      description: formData.description,
      department: formData.department,
      departmentId: formData.departmentId || formData.department,
      priority: formData.priority,
      budgetCode: formData.budgetCode,
      costCenter: formData.costCenter,
      projectCode: formData.projectCode,
      deliveryDate: formData.deliveryDate ?? undefined,
      // Only send vendor fields if they changed and PO is DRAFT
      ...(isDraft && formData.vendorId !== (purchaseOrder.vendorId || "")
        ? { vendorId: formData.vendorId, vendorName: formData.vendorName }
        : {}),
    });
  };

  // DRAFT and REJECTED can edit metadata fields; only DRAFT can change vendor
  const canEdit =
    purchaseOrder.status?.toUpperCase() === "DRAFT" ||
    purchaseOrder.status?.toUpperCase() === "REJECTED";

  const isDraft = purchaseOrder.status?.toUpperCase() === "DRAFT";

  return (
    <ResponsiveSheet
      open={open}
      onOpenChange={onOpenChange}
      title="Edit Purchase Order"
      description={purchaseOrder.documentNumber}
      desktopMaxWidth="sm:max-w-lg"
      dismissibleOnOutsideClick={false}
      footer={
        <div className="flex flex-wrap items-center justify-end gap-2">
          <Button
            type="button"
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={updateMutation.isPending}
          >
            Cancel
          </Button>
          <Button
            onClick={handleSubmit}
            disabled={updateMutation.isPending || !canEdit}
            isLoading={updateMutation.isPending}
            loadingText="Saving..."
          >
            <Save className="mr-2 h-4 w-4" />
            Save Changes
          </Button>
        </div>
      }
    >
      <div className="space-y-4 py-4">
        {/* Status Warning */}
        {!canEdit && (
          <div className="rounded-lg border border-amber-200 bg-amber-50 p-3">
            <p className="text-sm text-amber-800">
              This purchase order cannot be edited because it is{" "}
              <span className="font-semibold">{purchaseOrder.status}</span>.
              Only DRAFT or REJECTED purchase orders can be edited.
            </p>
          </div>
        )}

        {/* Document Number (Read-only) */}
        <div className="space-y-2">
          <Label>Document Number</Label>
          <Input
            value={purchaseOrder.documentNumber}
            disabled
            className="bg-muted"
          />
          <p className="text-xs text-muted-foreground">
            Document number cannot be changed
          </p>
        </div>

        {/* Title */}
        <Input
          label="Title"
          required
          placeholder="Enter purchase order title"
          value={formData.title}
          onChange={(e) => setFormData({ ...formData, title: e.target.value })}
          disabled={!canEdit}
        />

        {/* Description */}
        <div className="space-y-2">
          <Textarea
            id="description"
            label="Description"
            placeholder="Enter purchase order description..."
            rows={3}
            value={formData.description}
            onChange={(e) =>
              setFormData({ ...formData, description: e.target.value })
            }
            disabled={!canEdit}
          />
        </div>

        {/* Vendor — editable in DRAFT, read-only otherwise */}
        {isDraft ? (
          <SelectField
            label="Vendor / Supplier"
            isLoading={vendorsLoading}
            placeholder="Select vendor"
            value={formData.vendorId}
            onValueChange={(value) => {
              const vendor = vendors.find((v) => v.id === value);
              setFormData({
                ...formData,
                vendorId: value,
                vendorName: vendor?.name || "",
              });
            }}
            options={[
              { value: "", label: "No vendor selected" },
              ...vendors.map((v) => ({ value: v.id, label: v.name })),
            ]}
            disabled={!canEdit}
          />
        ) : (
          <div className="space-y-2">
            <Label>Vendor / Supplier</Label>
            <Input
              value={purchaseOrder.vendorName || "—"}
              disabled
              className="bg-muted"
            />
            <p className="text-xs text-muted-foreground">
              Vendor can only be changed while the PO is in Draft status
            </p>
          </div>
        )}

        {/* Department and Priority */}
        <div className="grid grid-cols-2 gap-4">
          <SelectField
            label="Department"
            required
            isLoading={departmentsLoading}
            placeholder="Select department"
            value={formData.department}
            onValueChange={(value) =>
              setFormData({
                ...formData,
                department: value,
                departmentId: Array.isArray(departments)
                  ? departments.find((d) => d.name === value)?.id || value
                  : value,
              })
            }
            options={
              Array.isArray(departments)
                ? departments.map((department) => ({
                    value: department.name,
                    label: department.name,
                  }))
                : []
            }
            disabled={!canEdit}
          />

          <SelectField
            label="Priority"
            required
            value={formData.priority}
            onValueChange={(value) =>
              setFormData({ ...formData, priority: value })
            }
            options={[
              { value: "LOW", label: "Low" },
              { value: "MEDIUM", label: "Medium" },
              { value: "HIGH", label: "High" },
              { value: "URGENT", label: "Urgent" },
            ]}
            placeholder="Select priority"
            disabled={!canEdit}
          />
        </div>

        {/* Budget Code */}
        <SelectField
          label="Budget Code"
          placeholder="Select budget code"
          isLoading={budgetsLoading}
          value={formData.budgetCode}
          onValueChange={(value) =>
            setFormData({ ...formData, budgetCode: value })
          }
          options={[
            { value: "", label: "None" },
            ...(Array.isArray(budgets)
              ? budgets.map((budget) => ({
                  value: budget.budgetCode,
                  label: `${budget.budgetCode} - ${budget.name}`,
                }))
              : []),
          ]}
          disabled={!canEdit}
        />

        {/* Cost Center and Project Code */}
        <div className="grid grid-cols-2 gap-4">
          <Input
            label="Cost Center"
            placeholder="Cost center code"
            value={formData.costCenter}
            onChange={(e) =>
              setFormData({ ...formData, costCenter: e.target.value })
            }
            disabled={!canEdit}
          />

          <Input
            label="Project Code"
            placeholder="Project code"
            value={formData.projectCode}
            onChange={(e) =>
              setFormData({ ...formData, projectCode: e.target.value })
            }
            disabled={!canEdit}
          />
        </div>

        {/* Delivery Date */}
        <DatePicker
          label="Delivery Date"
          value={formData.deliveryDate ?? undefined}
          onValueChange={(date) =>
            setFormData({ ...formData, deliveryDate: date as Date })
          }
          disabled={!canEdit}
        />

        {/* Info: items and amounts are managed via the Items tab */}
        {isDraft && (
          <div className="rounded-lg border border-blue-200 bg-blue-50 p-3">
            <p className="text-sm text-blue-800">
              To edit line items, quantities, or unit prices, use the{" "}
              <span className="font-semibold">Edit Items</span> button on the{" "}
              <span className="font-semibold">PO Items</span> tab.
            </p>
          </div>
        )}

        {/* Read-only fields */}
        <div className="space-y-4 pt-4 border-t">
          <h3 className="text-sm font-semibold text-muted-foreground">
            Read-Only Fields
          </h3>
          <div className="space-y-2">
            <Label>Status</Label>
            <Input
              value={purchaseOrder.status || "—"}
              disabled
              className="bg-muted"
            />
            <p className="text-xs text-muted-foreground">
              Status is managed by the approval workflow
            </p>
          </div>
        </div>
      </div>
    </ResponsiveSheet>
  );
}
