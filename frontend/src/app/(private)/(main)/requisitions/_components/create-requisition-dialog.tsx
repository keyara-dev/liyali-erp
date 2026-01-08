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
import { Label } from "@/components/ui/label";
import { Checkbox } from "@/components/ui/checkbox";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { SelectField } from "@/components/ui/select-field";
import { Plus, Trash2 } from "lucide-react";
import { RequisitionItem, RequisitionPriority, Requisition } from "@/types/requisition";
import { useCreateRequisition, useUpdateRequisition } from "@/hooks/use-requisition-mutations";
import { useCategories } from "@/hooks/use-category-queries";

interface CreateRequisitionDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onRequisitionCreated: () => void;
  userId: string;
  editingRequisition?: Requisition | null; // Add editing support
  isEditing?: boolean; // Add editing mode flag
}

export function CreateRequisitionDialog({
  open,
  onOpenChange,
  onRequisitionCreated,
  userId,
  editingRequisition = null,
  isEditing = false,
}: CreateRequisitionDialogProps) {
  const createMutation = useCreateRequisition(() => {
    // Reset form on success
    resetForm();
    onRequisitionCreated();
  });

  const updateMutation = useUpdateRequisition(() => {
    // Reset form on success
    resetForm();
    onRequisitionCreated();
  });

  // Fetch categories for the dropdown
  const { data: categories = [] } = useCategories(1, 50, true);

  const [formData, setFormData] = useState({
    title: "",
    department: "",
    departmentId: "",
    priority: "medium" as RequisitionPriority,
    requestedFor: "",
    justification: "",
    budgetCode: "",
    costCenter: "",
    projectCode: "",
    currency: "ZMW",
    isEstimate: true,
    requiredByDate: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000), // 30 days from now
    items: [] as RequisitionItem[],
    categoryId: "",
    otherCategoryText: "",
  });

  const resetForm = () => {
    setFormData({
      title: "",
      department: "",
      departmentId: "",
      priority: "medium",
      requestedFor: "",
      justification: "",
      budgetCode: "",
      costCenter: "",
      projectCode: "",
      currency: "ZMW",
      isEstimate: true,
      requiredByDate: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000),
      items: [],
      categoryId: "",
      otherCategoryText: "",
    });
  };

  // Populate form when editing
  useEffect(() => {
    if (isEditing && editingRequisition && open) {
      setFormData({
        title: editingRequisition.title || "",
        department: editingRequisition.department || "",
        departmentId: editingRequisition.departmentId || "",
        priority: editingRequisition.priority || "medium",
        requestedFor: editingRequisition.requestedFor || "",
        justification: editingRequisition.description || "",
        budgetCode: editingRequisition.budgetCode || "",
        costCenter: editingRequisition.costCenter || "",
        projectCode: editingRequisition.projectCode || "",
        currency: editingRequisition.currency || "ZMW",
        isEstimate: editingRequisition.isEstimate || false,
        requiredByDate: editingRequisition.requiredByDate ? new Date(editingRequisition.requiredByDate) : new Date(Date.now() + 30 * 24 * 60 * 60 * 1000),
        items: editingRequisition.items || [],
        categoryId: editingRequisition.categoryId || "",
        otherCategoryText: editingRequisition.otherCategoryText || "",
      });
    } else if (!isEditing && open) {
      resetForm();
    }
  }, [isEditing, editingRequisition, open]);

  const totalEstimatedCost = formData.items.reduce(
    (sum, item) => sum + (item.estimatedCost || 0) * item.quantity,
    0
  );

  const totalAmount = formData.items.reduce(
    (sum, item) => sum + (item.amount || (item.estimatedCost || 0) * item.quantity),
    0
  );

  const handleAddItem = () => {
    const newItem: RequisitionItem = {
      id: Date.now().toString(),
      description: "",
      itemDescription: "",  // Alias
      quantity: 1,
      unitPrice: 0,
      amount: 0,
      estimatedCost: 0,     // Alias
    };
    setFormData((prev) => ({
      ...prev,
      items: [...prev.items, newItem],
    }));
  };

  const handleRemoveItem = (itemId: string) => {
    setFormData((prev) => ({
      ...prev,
      items: prev.items.filter((item) => item.id !== itemId),
    }));
  };

  const handleUpdateItem = (
    itemId: string,
    field: keyof RequisitionItem,
    value: any
  ) => {
    setFormData((prev) => ({
      ...prev,
      items: prev.items.map((item) => {
        if (item.id === itemId) {
          const updatedItem = { ...item, [field]: value };
          
          // Calculate amount when quantity or estimatedCost changes
          if (field === 'quantity' || field === 'estimatedCost') {
            updatedItem.amount = updatedItem.quantity * (updatedItem.estimatedCost || 0);
            updatedItem.unitPrice = updatedItem.estimatedCost || 0;
          }
          
          // Ensure description is set from itemDescription
          if (field === 'itemDescription') {
            updatedItem.description = value;
          }
          
          return updatedItem;
        }
        return item;
      }),
    }));
  };

  const handleSubmit = async () => {
    // Validation
    if (!formData.title.trim()) {
      toast.error("Please enter a title for the requisition");
      return;
    }
    if (!formData.department.trim()) {
      toast.error("Please enter department");
      return;
    }
    if (!formData.requestedFor.trim()) {
      toast.error("Please enter who this is requested for");
      return;
    }
    if (formData.items.length === 0) {
      toast.error("Please add at least one item");
      return;
    }
    if (!formData.justification.trim()) {
      toast.error("Please provide justification");
      return;
    }
    if (!formData.budgetCode.trim()) {
      toast.error("Please enter budget code");
      return;
    }
    if (formData.categoryId === "OTHER" && !formData.otherCategoryText.trim()) {
      toast.error("Please specify the custom category name");
      return;
    }

    // Validate all items have descriptions and quantities
    const allItemsValid = formData.items.every(
      (item) => (item.itemDescription?.trim() || item.description?.trim()) && item.quantity > 0
    );
    if (!allItemsValid) {
      toast.error("Please fill in all item details");
      return;
    }

    // Use the appropriate mutation hook based on mode
    if (isEditing && editingRequisition) {
      updateMutation.mutate({
        requisitionId: editingRequisition.id,
        title: formData.title,
        description: formData.justification,
        department: formData.department,
        departmentId: formData.departmentId || formData.department,
        priority: formData.priority,
        items: formData.items,
        totalAmount: totalAmount,
        currency: formData.currency,
        categoryId: formData.categoryId === "OTHER" ? undefined : formData.categoryId || undefined,
        preferredVendorId: undefined,
        isEstimate: formData.isEstimate,
        requiredByDate: formData.requiredByDate,
        budgetCode: formData.budgetCode,
        costCenter: formData.costCenter || formData.budgetCode,
        projectCode: formData.projectCode || formData.budgetCode,
        requestedFor: formData.requestedFor,
        otherCategoryText: formData.categoryId === "OTHER" ? formData.otherCategoryText : undefined,
      });
    } else {
      createMutation.mutate({
        title: formData.title,
        description: formData.justification,
        department: formData.department,
        departmentId: formData.departmentId || formData.department,
        priority: formData.priority,
        items: formData.items,
        totalAmount: totalAmount,
        currency: formData.currency,
        categoryId: formData.categoryId === "OTHER" ? undefined : formData.categoryId || undefined,
        preferredVendorId: undefined,
        isEstimate: formData.isEstimate,
        requiredByDate: formData.requiredByDate,
        budgetCode: formData.budgetCode,
        costCenter: formData.costCenter || formData.budgetCode,
        projectCode: formData.projectCode || formData.budgetCode,
        requestedFor: formData.requestedFor,
        otherCategoryText: formData.categoryId === "OTHER" ? formData.otherCategoryText : undefined,
      });
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-3xl! p-0 overflow-y-auto max-h-[90vh]">
        <DialogHeader className="p-4 pb-0">
          <DialogTitle className="font-bold">
            {isEditing ? "Edit Requisition" : "Create New Requisition"}
          </DialogTitle>
          <DialogDescription>
            {isEditing 
              ? "Update the requisition details and items" 
              : "Fill in the requisition details and add items you need"
            }
          </DialogDescription>
        </DialogHeader>

          <div className="space-y-6 p-4  ">
            {/* Basic Information */}
            <div className="space-y-4">
              <h3 className="font-semibold text-lg">Basic Information</h3>

              <Input
                label="Title"
                required
                placeholder="Enter requisition title"
                value={formData.title}
                onChange={(e) =>
                  setFormData({ ...formData, title: e.target.value })
                }
              />

              <div className="grid grid-cols-2 gap-4">
                <Input
                  label="Department"
                  required
                  placeholder="e.g., Operations"
                  value={formData.department}
                  onChange={(e) =>
                    setFormData({ ...formData, department: e.target.value })
                  }
                />

                <SelectField
                  label="Priority"
                  required
                  value={formData.priority}
                  onValueChange={(value) =>
                    setFormData({ ...formData, priority: value as RequisitionPriority })
                  }
                  options={[
                    { id: "low", name: "Low" },
                    { id: "medium", name: "Medium" },
                    { id: "high", name: "High" },
                    { id: "urgent", name: "Urgent" },
                  ]}
                  placeholder="Select priority"
                />
              </div>

              <div className="grid grid-cols-2 gap-4">
                <Input
                  label="Requested For"
                  required
                  placeholder="e.g., John Mwale"
                  value={formData.requestedFor}
                  onChange={(e) =>
                    setFormData({ ...formData, requestedFor: e.target.value })
                  }
                />

                <SelectField
                  label="Currency"
                  value={formData.currency}
                  onValueChange={(value) =>
                    setFormData({ ...formData, currency: value })
                  }
                  options={[
                    { id: "ZMW", name: "ZMW" },
                    { id: "USD", name: "USD" },
                  ]}
                  placeholder="Select currency"
                />
              </div>

              {/* Category Selection */}
              <SelectField
                label="Category"
                className="w-full" 
                value={formData.categoryId}
                onValueChange={(value) =>
                  setFormData({ ...formData, categoryId: value, otherCategoryText: "" })
                }
                options={[
                  ...categories.map((category) => ({
                    id: category.id,
                    name: category.name,
                  })),
                  { id: "OTHER", name: "Other (specify below)" },
                ]}
                placeholder="Select a category"
              />

              {/* Other Category Text Input */}
              {formData.categoryId === "OTHER" && (
                <Input
                  label="Specify Category"
                  required
                  placeholder="Enter custom category name"
                  value={formData.otherCategoryText}
                  onChange={(e) =>
                    setFormData({ ...formData, otherCategoryText: e.target.value })
                  }
                />
              )}

              <div className="grid grid-cols-3 gap-4">
                <Input
                  label="Budget Code"
                  required
                  placeholder="e.g., CAP-2024-001"
                  value={formData.budgetCode}
                  onChange={(e) =>
                    setFormData({ ...formData, budgetCode: e.target.value })
                  }
                />

                <Input
                  label="Cost Center"
                  placeholder="Cost center"
                  value={formData.costCenter}
                  onChange={(e) =>
                    setFormData({ ...formData, costCenter: e.target.value })
                  }
                />

                <Input
                  label="Project Code"
                  placeholder="Project code"
                  value={formData.projectCode}
                  onChange={(e) =>
                    setFormData({ ...formData, projectCode: e.target.value })
                  }
                />
              </div>

              <div className="grid grid-cols-2 gap-4">
                <Input
                  label="Required By Date"
                  type="date"
                  value={formData.requiredByDate.toISOString().split('T')[0]}
                  onChange={(e) =>
                    setFormData({ ...formData, requiredByDate: new Date(e.target.value) })
                  }
                />

                <div className="space-y-2">
                  <Label htmlFor="isEstimate">Is Estimate</Label>
                  <div className="flex items-center space-x-2 h-10">
                    <Checkbox
                      id="isEstimate"
                      checked={formData.isEstimate}
                      onCheckedChange={(checked) =>
                        setFormData({ ...formData, isEstimate: checked === true })
                      }
                    />
                    <span className="text-sm text-gray-600">
                      This is an estimated cost
                    </span>
                  </div>
                </div>
              </div>

              <div className="space-y-2">
                <Textarea
                  id="justification"
                  label="Justification"
                  required
                  placeholder="Explain why these items are needed..."
                  rows={3}
                  value={formData.justification}
                  onChange={(e) =>
                    setFormData({
                      ...formData,
                      justification: e.target.value,
                    })
                  }
                />
              </div>
            </div>

            {/* Items Section */}
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <h3 className="font-semibold text-base">Items *</h3>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={handleAddItem}
                  className="gap-2"
                >
                  <Plus className="h-4 w-4" />
                  Add Item
                </Button>
              </div>

              {formData.items.length > 0 ? (
                <div className="space-y-3">
                  {formData.items.map((item, index) => (
                    <div
                      key={item.id}
                      className="border rounded-lg p-4 space-y-3"
                    >
                      <div className="flex items-start justify-between">
                        <span className="text-sm font-medium text-gray-600">
                          Item {index + 1}
                        </span>
                        <Button
                          type="button"
                          variant="ghost"
                          size="sm"
                          onClick={() => handleRemoveItem(item.id || '')}
                          className="text-red-500 hover:text-red-700 hover:bg-red-50"
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </div>

                      <div className="space-y-2">
                        <Input
                        label="Description"
                        required
                          placeholder="e.g., Office Chair - Ergonomic"
                          value={item.itemDescription}
                          onChange={(e) =>
                            handleUpdateItem(
                              item.id || '',
                              "itemDescription",
                              e.target.value
                            )
                          }
                        />
                      </div>

                      <div className="grid grid-cols-3 gap-4">
                        <Input
                          type="number"
                          placeholder="1"
                          label="Quantity"
                          min="1"
                          value={item.quantity}
                          onChange={(e) =>
                            handleUpdateItem(
                              item.id || '',
                              "quantity",
                              parseInt(e.target.value) || 1
                            )
                          }
                        />

                        <Input
                          type="number"
                          placeholder="0.00"
                          label="Est. Unit Cost (ZMW)"
                          min="0"
                          value={item.estimatedCost}
                          onChange={(e) =>
                            handleUpdateItem(
                              item.id || '',
                              "estimatedCost",
                              parseFloat(e.target.value) || 0
                            )
                          }
                        />

                        <div className="space-y-2">
                          <Label>Total ({formData.currency})</Label>
                          <div className="flex items-center justify-center h-10 bg-gray-50 rounded-lg border border-gray-200">
                            <span className="font-semibold">
                              {(
                                item.quantity * (item.estimatedCost || 0)
                              ).toLocaleString("en-US", {
                                minimumFractionDigits: 2,
                                maximumFractionDigits: 2,
                              })}
                            </span>
                          </div>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              ) : (
                <div className="border-2 border-dashed rounded-lg p-6 text-center">
                  <p className="text-gray-600 text-sm">
                    No items added yet. Click "Add Item" to get started.
                  </p>
                </div>
              )}
            </div>

            {/* Summary */}
            {formData.items.length > 0 && (
              <div className="bg-blue-50 rounded-lg p-4">
                <div className="flex items-center justify-between">
                  <span className="font-semibold text-gray-700">
                    Total Amount:
                  </span>
                  <span className="text-xl font-bold text-blue-600">
                    {formData.currency}{" "}
                    {totalAmount.toLocaleString("en-US", {
                      minimumFractionDigits: 2,
                      maximumFractionDigits: 2,
                    })}
                  </span>
                </div>
                {formData.isEstimate && (
                  <p className="text-xs text-gray-500 mt-1">
                    * This is an estimated amount
                  </p>
                )}
              </div>
            )}
          </div>
  

        {/* Dialog Footer */}
        <div 
        
        className="bg-card/5 backdrop-blur-xs sticky bottom-0 flex flex-col-reverse justify-end gap-3 p-4 rounded-b-lg border-t py-6 sm:flex-row sm:py-6"
        // className="flex items-center justify-end gap-3 pt-6 border-t"
        
        >
          <Button
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={createMutation.isPending || updateMutation.isPending}
          >
            Cancel
          </Button>
          <Button
            onClick={handleSubmit}
            isLoading={createMutation.isPending || updateMutation.isPending}
            loadingText={isEditing ? "Updating..." : "Creating..."}
            className="min-w-32"
          >
            {isEditing ? "Update Requisition" : "Create Requisition"}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
