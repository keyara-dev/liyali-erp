'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { RequisitionItem } from '@/types/workflow'
import { Plus, Trash2 } from 'lucide-react'
import { RequisitionFormData } from './create-requisition-client'
import { ItemInput } from './item-input'
import { v4 as uuidv4 } from 'uuid'

interface CreateRequisitionFormProps {
  onSubmit: (data: RequisitionFormData) => void
  initialData: RequisitionFormData
}

const DEPARTMENTS = [
  'Operations',
  'HR',
  'Finance',
  'IT',
  'Marketing',
  'Sales',
  'Legal',
]

export function CreateRequisitionForm({
  onSubmit,
  initialData,
}: CreateRequisitionFormProps) {
  const [department, setDepartment] = useState(initialData.department)
  const [requestedFor, setRequestedFor] = useState(initialData.requestedFor)
  const [justification, setJustification] = useState(initialData.justification)
  const [budgetCode, setBudgetCode] = useState(initialData.budgetCode)
  const [items, setItems] = useState<RequisitionItem[]>(initialData.items)
  const [errors, setErrors] = useState<Record<string, string>>({})

  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {}

    if (!department.trim()) newErrors.department = 'Department is required'
    if (!requestedFor.trim()) newErrors.requestedFor = 'Requested for is required'
    if (!justification.trim()) newErrors.justification = 'Justification is required'
    if (!budgetCode.trim()) newErrors.budgetCode = 'Budget code is required'
    if (items.length === 0) newErrors.items = 'At least one item is required'

    items.forEach((item, index) => {
      if (!item.itemDescription.trim())
        newErrors[`item-${index}-description`] = 'Description is required'
      if (item.quantity <= 0)
        newErrors[`item-${index}-quantity`] = 'Quantity must be greater than 0'
      if (item.estimatedCost <= 0)
        newErrors[`item-${index}-cost`] = 'Cost must be greater than 0'
    })

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const handleAddItem = () => {
    setItems([
      ...items,
      {
        id: uuidv4(),
        itemDescription: '',
        quantity: 1,
        estimatedCost: 0,
      },
    ])
  }

  const handleRemoveItem = (index: number) => {
    setItems(items.filter((_, i) => i !== index))
  }

  const handleItemChange = (index: number, field: string, value: any) => {
    const updatedItems = [...items]
    const item = updatedItems[index]

    if (field === 'itemDescription') {
      item.itemDescription = value
    } else if (field === 'quantity') {
      item.quantity = parseInt(value) || 0
    } else if (field === 'estimatedCost') {
      item.estimatedCost = parseFloat(value) || 0
    }

    updatedItems[index] = item
    setItems(updatedItems)
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()

    if (!validateForm()) {
      return
    }

    onSubmit({
      department,
      requestedFor,
      justification,
      budgetCode,
      items,
    })
  }

  const totalAmount = items.reduce((sum, item) => sum + item.estimatedCost * item.quantity, 0)

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      {/* Basic Information */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Requisition Details</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
            {/* Department */}
            <div className="space-y-2">
              <label className="text-sm font-medium">Department *</label>
              <select
                value={department}
                onChange={(e) => {
                  setDepartment(e.target.value)
                  setErrors({ ...errors, department: '' })
                }}
                className={`w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-primary ${
                  errors.department ? 'border-destructive' : 'border-border'
                }`}
              >
                <option value="">Select department</option>
                {DEPARTMENTS.map((dept) => (
                  <option key={dept} value={dept}>
                    {dept}
                  </option>
                ))}
              </select>
              {errors.department && (
                <p className="text-sm text-destructive">{errors.department}</p>
              )}
            </div>

            {/* Requested For */}
            <div className="space-y-2">
              <label className="text-sm font-medium">Requested For *</label>
              <Input
                value={requestedFor}
                onChange={(e) => {
                  setRequestedFor(e.target.value)
                  setErrors({ ...errors, requestedFor: '' })
                }}
                placeholder="Name or ID"
              />
              {errors.requestedFor && (
                <p className="text-sm text-destructive">{errors.requestedFor}</p>
              )}
            </div>
          </div>

          {/* Justification */}
          <div className="space-y-2">
            <label className="text-sm font-medium">Justification *</label>
            <Textarea
              value={justification}
              onChange={(e) => {
                setJustification(e.target.value)
                setErrors({ ...errors, justification: '' })
              }}
              placeholder="Explain why these items are needed"
              rows={3}
            />
            {errors.justification && (
              <p className="text-sm text-destructive">{errors.justification}</p>
            )}
          </div>

          {/* Budget Code */}
          <div className="space-y-2">
            <label className="text-sm font-medium">Budget Code *</label>
            <Input
              value={budgetCode}
              onChange={(e) => {
                setBudgetCode(e.target.value)
                setErrors({ ...errors, budgetCode: '' })
              }}
              placeholder="e.g., CAP-2024-001"
            />
            {errors.budgetCode && (
              <p className="text-sm text-destructive">{errors.budgetCode}</p>
            )}
          </div>
        </CardContent>
      </Card>

      {/* Items Section */}
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0">
          <CardTitle className="text-lg">Items</CardTitle>
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
        </CardHeader>
        <CardContent className="space-y-4">
          {errors.items && (
            <p className="text-sm text-destructive bg-destructive/10 p-2 rounded">
              {errors.items}
            </p>
          )}

          {items.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground border border-dashed rounded-lg">
              No items added yet. Click "Add Item" to start.
            </div>
          ) : (
            <div className="space-y-3">
              {items.map((item, index) => (
                <ItemInput
                  key={item.id}
                  index={index}
                  item={item}
                  errors={errors}
                  onChange={handleItemChange}
                  onRemove={handleRemoveItem}
                />
              ))}
            </div>
          )}

          {/* Summary */}
          {items.length > 0 && (
            <div className="border-t pt-4 mt-4">
              <div className="flex justify-between items-center">
                <span className="text-sm font-medium">
                  Total Amount (K{totalAmount.toFixed(2)})
                </span>
                <Badge variant="outline" className="text-lg font-semibold px-3 py-1">
                  K{totalAmount.toFixed(2)}
                </Badge>
              </div>
              <p className="text-xs text-muted-foreground mt-2">
                {items.length} item{items.length !== 1 ? 's' : ''} · Auto-calculated
              </p>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Action Buttons */}
      <div className="flex gap-3 justify-end">
        <Button type="button" variant="outline">
          Save as Draft
        </Button>
        <Button type="submit" className="gap-2">
          Continue to Preview
        </Button>
      </div>
    </form>
  )
}
