'use client'

import { useState } from 'react'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'

interface AddBudgetItemDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onItemAdded: (item: {
    category: string
    description: string
    allocatedAmount: number
  }) => void
}

export function AddBudgetItemDialog({
  open,
  onOpenChange,
  onItemAdded,
}: AddBudgetItemDialogProps) {
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [formData, setFormData] = useState({
    category: '',
    description: '',
    allocatedAmount: '',
  })

  const handleInputChange = (field: string, value: string | number) => {
    setFormData((prev) => ({
      ...prev,
      [field]: value,
    }))
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!formData.category || !formData.allocatedAmount) {
      alert('Please fill in all required fields')
      return
    }

    setIsSubmitting(true)
    try {
      onItemAdded({
        category: formData.category,
        description: formData.description,
        allocatedAmount: parseFloat(formData.allocatedAmount),
      })

      setFormData({
        category: '',
        description: '',
        allocatedAmount: '',
      })
      onOpenChange(false)
    } catch (error) {
      console.error('Error adding budget item:', error)
      alert('An error occurred while adding the budget item')
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>Add Budget Item</DialogTitle>
          <DialogDescription>
            Add a new line item to your budget allocation
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          {/* Category */}
          <div className="space-y-2">
            <Label htmlFor="category">Category *</Label>
            <Input
              id="category"
              placeholder="e.g., Hardware, Software, Personnel"
              value={formData.category}
              onChange={(e) => handleInputChange('category', e.target.value)}
              disabled={isSubmitting}
            />
          </div>

          {/* Description */}
          <div className="space-y-2">
            <Label htmlFor="description">Description</Label>
            <Textarea
              id="description"
              placeholder="Describe this budget item (optional)"
              value={formData.description}
              onChange={(e) => handleInputChange('description', e.target.value)}
              disabled={isSubmitting}
              rows={3}
            />
          </div>

          {/* Allocated Amount */}
          <div className="space-y-2">
            <Label htmlFor="allocatedAmount">Allocated Amount *</Label>
            <Input
              id="allocatedAmount"
              type="number"
              placeholder="0.00"
              step="0.01"
              value={formData.allocatedAmount}
              onChange={(e) => handleInputChange('allocatedAmount', e.target.value)}
              disabled={isSubmitting}
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
              {isSubmitting ? 'Adding...' : 'Add Item'}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  )
}
