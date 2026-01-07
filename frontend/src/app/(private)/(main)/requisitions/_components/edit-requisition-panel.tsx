'use client'

import { useState } from 'react'
import { toast } from 'sonner'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'
import { Plus, Trash2, Save } from 'lucide-react'
import { updateRequisition } from '@/app/_actions/requisitions';
import { Requisition, RequisitionItem } from '@/types/requisition';

interface EditRequisitionPanelProps {
  requisition: Requisition
  onRequisitionUpdated: () => void
}

export function EditRequisitionPanel({
  requisition,
  onRequisitionUpdated,
}: EditRequisitionPanelProps) {
  const [isEditing, setIsEditing] = useState(false)
  const [isSaving, setIsSaving] = useState(false)
  const [formData, setFormData] = useState({
    department: requisition.metadata?.department || '',
    requestedFor: requisition.metadata?.requestedFor || '',
    justification: requisition.metadata?.justification || '',
    budgetCode: requisition.metadata?.budgetCode || '',
    items: requisition.metadata?.items || ([] as RequisitionItem[]),
  })

  const totalEstimatedCost = formData.items.reduce(
    (sum, item) => sum + (item.estimatedCost || 0),
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
      items: prev.items.filter((item) => item.id !== itemId),
    }))
  }

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
    }))
  }

  const handleSave = async () => {
    // Validation
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
      (item) => item.itemDescription.trim() && item.quantity > 0
    )
    if (!allItemsValid) {
      toast.error('Please fill in all item details')
      return
    }

    setIsSaving(true)
    try {
      const result = await updateRequisition({
        id: requisition.id,
        department: formData.department,
        requestedFor: formData.requestedFor,
        items: formData.items,
        justification: formData.justification,
        budgetCode: formData.budgetCode,
      })

      if (result.success) {
        toast.success('Requisition updated successfully')
        setIsEditing(false)
        onRequisitionUpdated()
      } else {
        toast.error(result.message)
      }
    } catch (error) {
      toast.error('Failed to save requisition')
    } finally {
      setIsSaving(false)
    }
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

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="edit-department">Department</Label>
              <Input
                id="edit-department"
                value={formData.department}
                onChange={(e) =>
                  setFormData({ ...formData, department: e.target.value })
                }
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="edit-requestedFor">Requested For</Label>
              <Input
                id="edit-requestedFor"
                value={formData.requestedFor}
                onChange={(e) =>
                  setFormData({ ...formData, requestedFor: e.target.value })
                }
              />
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="edit-budgetCode">Budget Code</Label>
            <Input
              id="edit-budgetCode"
              value={formData.budgetCode}
              onChange={(e) =>
                setFormData({ ...formData, budgetCode: e.target.value })
              }
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="edit-justification">Justification</Label>
            <Textarea
              id="edit-justification"
              rows={3}
              value={formData.justification}
              onChange={(e) =>
                setFormData({ ...formData, justification: e.target.value })
              }
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
            {formData.items.map((item, index) => (
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
                    onClick={() => handleRemoveItem(item.id)}
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
                        item.id,
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
                          item.id,
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
                          item.id,
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
                        {(
                          item.quantity * item.estimatedCost
                        ).toLocaleString('en-ZM', {
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
                Total Estimated Cost:
              </span>
              <span className="text-xl font-bold text-blue-600">
                ZMW{' '}
                {totalEstimatedCost.toLocaleString('en-ZM', {
                  minimumFractionDigits: 2,
                  maximumFractionDigits: 2,
                })}
              </span>
            </div>
          </div>
        )}

        {/* Action Buttons */}
        <div className="flex gap-3 pt-4 border-t">
          <Button
            variant="outline"
            onClick={() => setIsEditing(false)}
            disabled={isSaving}
            className="flex-1"
          >
            Cancel
          </Button>
          <Button
            onClick={handleSave}
            disabled={isSaving}
            className="flex-1 gap-2"
          >
            <Save className="h-4 w-4" />
            {isSaving ? 'Saving...' : 'Save Changes'}
          </Button>
        </div>
      </div>
    </Card>
  )
}
