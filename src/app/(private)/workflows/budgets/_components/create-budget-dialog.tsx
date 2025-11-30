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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { createBudget } from '@/app/_actions/budgets'

interface CreateBudgetDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onBudgetCreated: () => void
  userId: string
}

const departments = [
  { id: 'dept-it', label: 'Information Technology' },
  { id: 'dept-hr', label: 'Human Resources' },
  { id: 'dept-ops', label: 'Operations' },
  { id: 'dept-sales', label: 'Sales' },
  { id: 'dept-marketing', label: 'Marketing' },
  { id: 'dept-finance', label: 'Finance' },
]

const currencies = [
  { code: 'ZMW', label: 'Zambian Kwacha (K)' },
  { code: 'USD', label: 'US Dollar ($)' },
]

export function CreateBudgetDialog({
  open,
  onOpenChange,
  onBudgetCreated,
  userId,
}: CreateBudgetDialogProps) {
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    department: '',
    departmentId: '',
    fiscalYear: new Date().getFullYear().toString(),
    totalAmount: '',
    currency: 'ZMW',
  })

  const handleInputChange = (field: string, value: string | number) => {
    setFormData((prev) => ({
      ...prev,
      [field]: value,
    }))
  }

  const handleDepartmentChange = (departmentId: string) => {
    const dept = departments.find((d) => d.id === departmentId)
    setFormData((prev) => ({
      ...prev,
      departmentId,
      department: dept?.label || '',
    }))
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (
      !formData.name ||
      !formData.department ||
      !formData.totalAmount ||
      !formData.fiscalYear
    ) {
      alert('Please fill in all required fields')
      return
    }

    setIsSubmitting(true)
    try {
      const result = await createBudget({
        name: formData.name,
        description: formData.description,
        department: formData.department,
        departmentId: formData.departmentId,
        fiscalYear: formData.fiscalYear,
        totalAmount: parseFloat(formData.totalAmount),
        currency: formData.currency,
        items: [],
        createdBy: userId,
      })

      if (result.success) {
        setFormData({
          name: '',
          description: '',
          department: '',
          departmentId: '',
          fiscalYear: new Date().getFullYear().toString(),
          totalAmount: '',
          currency: 'ZMW',
        })
        onOpenChange(false)
        onBudgetCreated()
      } else {
        alert('Failed to create budget: ' + result.message)
      }
    } catch (error) {
      console.error('Error creating budget:', error)
      alert('An error occurred while creating the budget')
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>Create New Budget</DialogTitle>
          <DialogDescription>
            Create a new budget for your department. You can add budget items later.
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          {/* Budget Name */}
          <div className="space-y-2">
            <Label htmlFor="name">Budget Name *</Label>
            <Input
              id="name"
              placeholder="e.g., IT Department Annual Budget 2024"
              value={formData.name}
              onChange={(e) => handleInputChange('name', e.target.value)}
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
              onChange={(e) => handleInputChange('description', e.target.value)}
              disabled={isSubmitting}
              rows={3}
            />
          </div>

          {/* Department */}
          <div className="space-y-2">
            <Label htmlFor="department">Department *</Label>
            <Select value={formData.departmentId} onValueChange={handleDepartmentChange}>
              <SelectTrigger id="department" disabled={isSubmitting}>
                <SelectValue placeholder="Select a department" />
              </SelectTrigger>
              <SelectContent>
                {departments.map((dept) => (
                  <SelectItem key={dept.id} value={dept.id}>
                    {dept.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          {/* Fiscal Year */}
          <div className="space-y-2">
            <Label htmlFor="fiscalYear">Fiscal Year *</Label>
            <Input
              id="fiscalYear"
              type="number"
              placeholder="2024"
              value={formData.fiscalYear}
              onChange={(e) => handleInputChange('fiscalYear', e.target.value)}
              disabled={isSubmitting}
            />
          </div>

          {/* Total Amount */}
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="totalAmount">Total Amount *</Label>
              <Input
                id="totalAmount"
                type="number"
                placeholder="0.00"
                step="0.01"
                value={formData.totalAmount}
                onChange={(e) => handleInputChange('totalAmount', e.target.value)}
                disabled={isSubmitting}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="currency">Currency *</Label>
              <Select value={formData.currency} onValueChange={(value) => handleInputChange('currency', value)}>
                <SelectTrigger id="currency" disabled={isSubmitting}>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {currencies.map((curr) => (
                    <SelectItem key={curr.code} value={curr.code}>
                      {curr.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
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
              {isSubmitting ? 'Creating...' : 'Create Budget'}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  )
}
