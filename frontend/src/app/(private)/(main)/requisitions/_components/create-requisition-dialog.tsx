"use client";

import { useState } from "react";
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
import { ScrollArea } from "@/components/ui/scroll-area";
import { Plus, Trash2 } from "lucide-react";
import { RequisitionItem, RequisitionPriority } from "@/types/requisition";
import { useCreateRequisition } from "@/hooks/use-requisition-mutations";

interface CreateRequisitionDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onRequisitionCreated: () => void;
  userId: string;
}

export function CreateRequisitionDialog({
  open,
  onOpenChange,
  onRequisitionCreated,
  userId,
}: CreateRequisitionDialogProps) {
  const createMutation = useCreateRequisition(() => {
    // Reset form on success
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
    });
    onRequisitionCreated();
  });

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
  });

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
      items: prev.items.map((item) =>
        item.id === itemId ? { ...item, [field]: value } : item
      ),
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

    // Validate all items have descriptions and quantities
    const allItemsValid = formData.items.every(
      (item) => item.itemDescription?.trim() && item.quantity > 0
    );
    if (!allItemsValid) {
      toast.error("Please fill in all item details");
      return;
    }

    // Use the mutation hook
    createMutation.mutate({
      title: formData.title,
      description: formData.justification,
      department: formData.department,
      departmentId: formData.departmentId || formData.department, // Use department as fallback
      priority: formData.priority,
      items: formData.items,
      totalAmount: totalAmount,
      currency: formData.currency,
      categoryId: undefined, // Optional field
      preferredVendorId: undefined, // Optional field
      isEstimate: formData.isEstimate,
      requiredByDate: formData.requiredByDate,
      budgetCode: formData.budgetCode,
      costCenter: formData.costCenter || formData.budgetCode, // Use budgetCode as fallback
      projectCode: formData.projectCode || formData.budgetCode, // Use budgetCode as fallback
      createdBy: userId,
      createdByName: 'User', // Default name - could be enhanced with actual user data
      createdByRole: 'requester', // Default role
      requestedFor: formData.requestedFor,
      justification: formData.justification,
    });
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl! max-h-[90vh]">
        <DialogHeader>
          <DialogTitle>Create New Requisition</DialogTitle>
          <DialogDescription>
            Fill in the requisition details and add items you need
          </DialogDescription>
        </DialogHeader>

        <ScrollArea className="h-[calc(90vh-200px)]">
          <div className="space-y-6 pr-4">
            {/* Basic Information */}
            <div className="space-y-4">
              <h3 className="font-semibold text-lg">Basic Information</h3>

              <div className="space-y-2">
                <Label htmlFor="title">Title *</Label>
                <Input
                  id="title"
                  placeholder="Enter requisition title"
                  value={formData.title}
                  onChange={(e) =>
                    setFormData({ ...formData, title: e.target.value })
                  }
                />
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="department">Department *</Label>
                  <Input
                    id="department"
                    placeholder="e.g., Operations"
                    value={formData.department}
                    onChange={(e) =>
                      setFormData({ ...formData, department: e.target.value })
                    }
                  />
                </div>

                <div className="space-y-2">
                  <Label htmlFor="priority">Priority</Label>
                  <select
                    id="priority"
                    value={formData.priority}
                    onChange={(e) =>
                      setFormData({ ...formData, priority: e.target.value as RequisitionPriority })
                    }
                    className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                  >
                    <option value="low">Low</option>
                    <option value="medium">Medium</option>
                    <option value="high">High</option>
                    <option value="urgent">Urgent</option>
                  </select>
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="requestedFor">Requested For *</Label>
                  <Input
                    id="requestedFor"
                    placeholder="e.g., John Mwale"
                    value={formData.requestedFor}
                    onChange={(e) =>
                      setFormData({ ...formData, requestedFor: e.target.value })
                    }
                  />
                </div>

                <div className="space-y-2">
                  <Label htmlFor="currency">Currency</Label>
                  <select
                    id="currency"
                    value={formData.currency}
                    onChange={(e) =>
                      setFormData({ ...formData, currency: e.target.value })
                    }
                    className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                  >
                    <option value="ZMW">ZMW</option>
                    <option value="USD">USD</option>
                  </select>
                </div>
              </div>

              <div className="grid grid-cols-3 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="budgetCode">Budget Code *</Label>
                  <Input
                    id="budgetCode"
                    placeholder="e.g., CAP-2024-001"
                    value={formData.budgetCode}
                    onChange={(e) =>
                      setFormData({ ...formData, budgetCode: e.target.value })
                    }
                  />
                </div>

                <div className="space-y-2">
                  <Label htmlFor="costCenter">Cost Center</Label>
                  <Input
                    id="costCenter"
                    placeholder="Cost center"
                    value={formData.costCenter}
                    onChange={(e) =>
                      setFormData({ ...formData, costCenter: e.target.value })
                    }
                  />
                </div>

                <div className="space-y-2">
                  <Label htmlFor="projectCode">Project Code</Label>
                  <Input
                    id="projectCode"
                    placeholder="Project code"
                    value={formData.projectCode}
                    onChange={(e) =>
                      setFormData({ ...formData, projectCode: e.target.value })
                    }
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="requiredByDate">Required By Date</Label>
                  <Input
                    id="requiredByDate"
                    type="date"
                    value={formData.requiredByDate.toISOString().split('T')[0]}
                    onChange={(e) =>
                      setFormData({ ...formData, requiredByDate: new Date(e.target.value) })
                    }
                  />
                </div>

                <div className="space-y-2">
                  <Label htmlFor="isEstimate">Is Estimate</Label>
                  <div className="flex items-center space-x-2 h-10">
                    <input
                      id="isEstimate"
                      type="checkbox"
                      checked={formData.isEstimate}
                      onChange={(e) =>
                        setFormData({ ...formData, isEstimate: e.target.checked })
                      }
                      className="h-4 w-4 rounded border-gray-300"
                    />
                    <span className="text-sm text-gray-600">
                      This is an estimated cost
                    </span>
                  </div>
                </div>
              </div>

              <div className="space-y-2">
                <Label htmlFor="justification">Justification *</Label>
                <Textarea
                  id="justification"
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
                <h3 className="font-semibold text-lg">Items *</h3>
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
                        <Label>Description</Label>
                        <Input
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
                          <div className="flex items-center justify-center h-9 bg-gray-50 rounded-lg border border-gray-200">
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
        </ScrollArea>

        {/* Dialog Footer */}
        <div className="flex items-center justify-end gap-3 pt-6 border-t">
          <Button
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={createMutation.isPending}
          >
            Cancel
          </Button>
          <Button
            onClick={handleSubmit}
            disabled={createMutation.isPending}
            className="min-w-32"
          >
            {createMutation.isPending ? "Creating..." : "Create Requisition"}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
