'use client'

import { useState } from 'react'
import { toast } from 'sonner'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Plus, Trash2 } from 'lucide-react'
import { createWorkflowDocument } from '@/app/_actions/workflow'
import { RequisitionItem } from '@/types/workflow'

interface CreateRequisitionDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onRequisitionCreated: () => void
  userId: string
}

export function CreateRequisitionDialog({
  open,
  onOpenChange,
  onRequisitionCreated,
  userId,
}: CreateRequisitionDialogProps) {
  const [isLoading, setIsLoading] = useState(false)
  const [formData, setFormData] = useState({
    department: '',
    requestedFor: '',
    justification: '',
    budgetCode: '',
    items: [] as RequisitionItem[],
  })

  const totalEstimatedCost = formData.items.reduce(
    (sum, item) => sum + (item.estimatedCost || 0),
    0
  )

  const handleAddItem = () => {
    const newItem: RequisitionItem = {
      id: Date.now().toString(),
      itemDescription: '',
      quantity: 1,
      estimatedCost: 0,
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

  const handleSubmit = async () => {
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
    if (!formData.budgetCode.trim()) {
      toast.error('Please enter budget code')
      return
    }

    // Validate all items have descriptions and quantities
    const allItemsValid = formData.items.every(
      (item) => item.itemDescription.trim() && item.quantity > 0
    )
    if (!allItemsValid) {
      toast.error('Please fill in all item details')
      return
    }

    setIsLoading(true)
    try {
      const result = await createWorkflowDocument('REQUISITION', {
        department: formData.department,
        requestedFor: formData.requestedFor,
        items: formData.items,
        justification: formData.justification,
        budgetCode: formData.budgetCode,
      })

      if (result.success) {
        toast.success(
          `Requisition ${result.data.documentNumber} created successfully`
        )
        // Reset form
        setFormData({
          department: '',
          requestedFor: '',
          justification: '',
          budgetCode: '',
          items: [],
        })
        onRequisitionCreated()
      } else {
        toast.error(result.message)
      }
    } catch (error) {
      toast.error('Failed to create requisition')
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl max-h-[90vh]">
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
              </div>

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
                            placeholder="1"
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
                          <Label>Est. Unit Cost (ZMW)</Label>
                          <Input
                            type="number"
                            placeholder="0.00"
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
                          <Label>Total (ZMW)</Label>
                          <div className="flex items-center justify-center h-10 bg-gray-50 rounded border border-gray-200">
                            <span className="font-semibold">
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
          </div>
        </ScrollArea>

        {/* Dialog Footer */}
        <div className="flex items-center justify-end gap-3 pt-6 border-t">
          <Button
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={isLoading}
          >
            Cancel
          </Button>
          <Button
            onClick={handleSubmit}
            disabled={isLoading}
            className="min-w-32"
          >
            {isLoading ? 'Creating...' : 'Create Requisition'}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  )
}
