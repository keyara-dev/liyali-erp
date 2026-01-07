'use client'

import { useState } from 'react'
import { toast } from 'sonner'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'
import { Plus, Trash2, Save } from 'lucide-react'
import { Requisition, RequisitionItem, RequisitionPriority } from '@/types/requisition';
import { useUpdateRequisition } from '@/hooks/use-requisition-mutations';

interface EditRequisitionPanelProps {
  requisition: Requisition
  onRequisitionUpdated: () => void
}

export function EditRequisitionPanel({
  requisition,
  onRequisitionUpdated,
}: EditRequisitionPanelProps) {
  const [isEditing, setIsEditing] = useState(false)
  
  const updateMutation = useUpdateRequisition(() => {
    setIsEditing(false);
    onRequisitionUpdated();
  });

  const [formData, setFormData] = useState({
    title: requisition.title || '',
    department: requisition.department || '',
    departmentId: requisition.departmentId || '',
    priority: requisition.priority || 'medium',
    requestedFor: requisition.metadata?.requestedFor || '',
    justification: requisition.description || '',
    budgetCode: requisition.budgetCode || '',
    costCenter: requisition.costCenter || '',
    projectCode: requisition.projectCode || '',
    currency: requisition.currency || 'USD',
    isEstimate: requisition.isEstimate ?? true,
    requiredByDate: requisition.requiredByDate ? new Date(requisition.requiredByDate) : new Date(Date.now() + 30 * 24 * 60 * 60 * 1000),
    items: requisition.items || ([] as RequisitionItem[]),
  })

  const totalEstimatedCost = formData.items.reduce(
    (sum: number, item: RequisitionItem) => sum + (item.estimatedCost || 0) * item.quantity,
    0
  )

  const totalAmount = formData.items.reduce(
    (sum: number, item: RequisitionItem) => sum + (item.amount || (item.estimatedCost || 0) * item.quantity),
    0
  )

  const handleAddItem = () => {
    const newItem: RequisitionItem = {
      id: Date.now().toString(),
      description: '',
      itemDescription: '',  // Alias
      quantity: 1,
      unitPrice: 0,
      amount: 0,
      estimatedCost: 0,     // Alias
    }
    setFormData((prev) => ({
      ...prev,
      items: [...prev.items, newItem],
    }))
  }

  const handleRemoveItem = (itemId: string) => {
    setFormData((prev) => ({
      ...prev,
      items: prev.items.filter((item: RequisitionItem) => item.id !== itemId),
    }))
  }

  const handleUpdateItem = (
    itemId: string,
    field: keyof RequisitionItem,
    value: any
  ) => {
    setFormData((prev) => ({
      ...prev,
      items: prev.items.map((item: RequisitionItem) =>
        item.id === itemId ? { ...item, [field]: value } : item
      ),
    }))
  }

  const handleSave = async () => {
    // Validation
    if (!formData.title.trim()) {
      toast.error('Please enter a title for the requisition')
      return
    }
    if (!formData.department.trim()) {
      toast.error('Please enter department')
      return
    }
    if (!formData.requestedFor.trim()) {
      toast.error('Please enter who this is requested for')
      return
    }
    if (formData.items.length === 0) {
      toast.error('Please add at least one item')
      return
    }
    if (!formData.justification.trim()) {
      toast.error('Please provide justification')
      return
    }

    const allItemsValid = formData.items.every(
      (item: RequisitionItem) => item.itemDescription?.trim() && item.quantity > 0
    )
    if (!allItemsValid) {
      toast.error('Please fill in all item details')
      return
    }

    // Use the mutation hook
    updateMutation.mutate({
      requisitionId: requisition.id,
      title: formData.title,
      description: formData.justification,
      department: formData.department,
      departmentId: formData.departmentId || formData.department, // Use department as fallback
      priority: formData.priority,
      items: formData.items,
      totalAmount: totalAmount,
      currency: formData.currency,
      isEstimate: formData.isEstimate,
      requiredByDate: formData.requiredByDate,
      budgetCode: formData.budgetCode,
      costCenter: formData.costCenter || formData.budgetCode, // Use budgetCode as fallback
      projectCode: formData.projectCode || formData.budgetCode, // Use budgetCode as fallback
    });
  }

  if (!isEditing) {
    return (
      <Card className="p-6 bg-blue-50 border-blue-200">
        <p className="text-sm text-gray-700 mb-3">
          This requisition is in draft status and can be edited.
        </p>
        <Button
          variant="outline"
          onClick={() => setIsEditing(true)}
          className="gap-2"
        >
          <Save className="h-4 w-4" />
          Edit Requisition
        </Button>
      </Card>
    )
  }

  return (
    <Card className="p-6 border-2 border-blue-300 bg-blue-50">
      <h2 className="text-xl font-semibold mb-4">Edit Requisition</h2>

      <div className="space-y-6">
        {/* Basic Information */}
        <div className="space-y-4">
          <h3 className="font-semibold text-lg">Basic Information</h3>

          <div className="space-y-2">
            <Label htmlFor="edit-title">Title *</Label>
            <Input
              id="edit-title"
              value={formData.title}
              onChange={(e) =>
                setFormData({ ...formData, title: e.target.value })
              }
              placeholder="Enter requisition title"
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="edit-department">Department *</Label>
              <Input
                id="edit-department"
                value={formData.department}
                onChange={(e) =>
                  setFormData({ ...formData, department: e.target.value })
                }
                placeholder="Enter department"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="edit-priority">Priority</Label>
              <select
                id="edit-priority"
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
              <Label htmlFor="edit-requestedFor">Requested For</Label>
              <Input
                id="edit-requestedFor"
                value={formData.requestedFor}
                onChange={(e) =>
                  setFormData({ ...formData, requestedFor: e.target.value })
                }
                placeholder="Who is this for?"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="edit-currency">Currency</Label>
              <select
                id="edit-currency"
                value={formData.currency}
                onChange={(e) =>
                  setFormData({ ...formData, currency: e.target.value })
                }
                className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
              >
                <option value="USD">USD</option>
                <option value="ZMW">ZMW</option>
                <option value="EUR">EUR</option>
                <option value="GBP">GBP</option>
              </select>
            </div>
          </div>

          <div className="grid grid-cols-3 gap-4">
            <div className="space-y-2">
              <Label htmlFor="edit-budgetCode">Budget Code</Label>
              <Input
                id="edit-budgetCode"
                value={formData.budgetCode}
                onChange={(e) =>
                  setFormData({ ...formData, budgetCode: e.target.value })
                }
                placeholder="Budget code"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="edit-costCenter">Cost Center</Label>
              <Input
                id="edit-costCenter"
                value={formData.costCenter}
                onChange={(e) =>
                  setFormData({ ...formData, costCenter: e.target.value })
                }
                placeholder="Cost center"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="edit-projectCode">Project Code</Label>
              <Input
                id="edit-projectCode"
                value={formData.projectCode}
                onChange={(e) =>
                  setFormData({ ...formData, projectCode: e.target.value })
                }
                placeholder="Project code"
              />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="edit-requiredByDate">Required By Date</Label>
              <Input
                id="edit-requiredByDate"
                type="date"
                value={formData.requiredByDate.toISOString().split('T')[0]}
                onChange={(e) =>
                  setFormData({ ...formData, requiredByDate: new Date(e.target.value) })
                }
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="edit-isEstimate">Is Estimate</Label>
              <div className="flex items-center space-x-2 h-10">
                <input
                  id="edit-isEstimate"
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
            <Label htmlFor="edit-justification">Justification *</Label>
            <Textarea
              id="edit-justification"
              rows={3}
              value={formData.justification}
              onChange={(e) =>
                setFormData({ ...formData, justification: e.target.value })
              }
              placeholder="Provide detailed justification for this requisition"
            />
          </div>
        </div>

        {/* Items Section */}
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <h3 className="font-semibold text-lg">Items</h3>
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

          <div className="space-y-3">
            {formData.items.map((item: RequisitionItem, index: number) => (
              <div
                key={item.id}
                className="border rounded-lg p-4 space-y-3 bg-white"
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
                    value={item.itemDescription || ''}
                    onChange={(e) =>
                      handleUpdateItem(
                        item.id || '',
                        'itemDescription',
                        e.target.value
                      )
                    }
                  />
                </div>

                <div className="grid grid-cols-3 gap-4">
                  <div className="space-y-2">
                    <Label>Quantity</Label>
                    <Input
                      type="number"
                      min="1"
                      value={item.quantity}
                      onChange={(e) =>
                        handleUpdateItem(
                          item.id || '',
                          'quantity',
                          parseInt(e.target.value) || 1
                        )
                      }
                    />
                  </div>

                  <div className="space-y-2">
                    <Label>Est. Unit Cost</Label>
                    <Input
                      type="number"
                      min="0"
                      value={item.estimatedCost}
                      onChange={(e) =>
                        handleUpdateItem(
                          item.id || '',
                          'estimatedCost',
                          parseFloat(e.target.value) || 0
                        )
                      }
                    />
                  </div>

                  <div className="space-y-2">
                    <Label>Total</Label>
                    <div className="flex items-center justify-center h-10 bg-gray-100 rounded border border-gray-300">
                      <span className="font-semibold text-sm">
                        {formData.currency}{' '}
                        {(
                          item.quantity * (item.estimatedCost || 0)
                        ).toLocaleString('en-US', {
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
        </div>

        {/* Total */}
        {formData.items.length > 0 && (
          <div className="bg-white rounded-lg p-4 border">
            <div className="flex items-center justify-between">
              <span className="font-semibold text-gray-700">
                Total Amount:
              </span>
              <span className="text-xl font-bold text-blue-600">
                {formData.currency}{' '}
                {totalAmount.toLocaleString('en-US', {
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

        {/* Action Buttons */}
        <div className="flex gap-3 pt-4 border-t">
          <Button
            variant="outline"
            onClick={() => setIsEditing(false)}
            disabled={updateMutation.isPending}
            className="flex-1"
          >
            Cancel
          </Button>
          <Button
            onClick={handleSave}
            disabled={updateMutation.isPending}
            isLoading={updateMutation.isPending}
            loadingText='Saving...'
            className="flex-1 gap-2"
          >
            <Save className="h-4 w-4" />
          Save Changes
          </Button>
        </div>
      </div>
    </Card>
  )
}
