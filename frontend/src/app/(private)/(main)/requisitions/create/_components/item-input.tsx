'use client'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card } from '@/components/ui/card'
import { RequisitionItem } from '@/types/requisition';
import { Trash2 } from 'lucide-react'

interface ItemInputProps {
  index: number
  item: RequisitionItem
  errors: Record<string, string>
  onChange: (index: number, field: string, value: any) => void
  onRemove: (index: number) => void
}

export function ItemInput({
  index,
  item,
  errors,
  onChange,
  onRemove,
}: ItemInputProps) {
  const itemTotal = item.quantity * (item.estimatedCost || 0)

  return (
    <Card className="p-4">
      <div className="space-y-4">
        {/* Header with Item Number and Delete */}
        <div className="flex items-center justify-between">
          <h4 className="text-sm font-semibold">Item {index + 1}</h4>
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={() => onRemove(index)}
            className="text-destructive hover:text-destructive hover:bg-destructive/10"
          >
            <Trash2 className="h-4 w-4" />
          </Button>
        </div>

        {/* Item Description */}
        <div className="space-y-2">
          <label className="text-sm font-medium">Description *</label>
          <Input
            value={item.itemDescription}
            onChange={(e) => onChange(index, 'itemDescription', e.target.value)}
            placeholder="What items are you requesting?"
            className={errors[`item-${index}-description`] ? 'border-destructive' : ''}
          />
          {errors[`item-${index}-description`] && (
            <p className="text-sm text-destructive">
              {errors[`item-${index}-description`]}
            </p>
          )}
        </div>

        {/* Quantity and Unit Cost */}
        <div className="grid grid-cols-3 gap-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Quantity *</label>
            <Input
              type="number"
              min="1"
              value={item.quantity}
              onChange={(e) => onChange(index, 'quantity', e.target.value)}
              placeholder="Qty"
              className={errors[`item-${index}-quantity`] ? 'border-destructive' : ''}
            />
            {errors[`item-${index}-quantity`] && (
              <p className="text-sm text-destructive">
                {errors[`item-${index}-quantity`]}
              </p>
            )}
          </div>

          <div className="space-y-2">
            <label className="text-sm font-medium">Unit Cost (K) *</label>
            <Input
              type="number"
              min="0"
              step="0.01"
              value={item.estimatedCost}
              onChange={(e) => onChange(index, 'estimatedCost', e.target.value)}
              placeholder="0.00"
              className={errors[`item-${index}-cost`] ? 'border-destructive' : ''}
            />
            {errors[`item-${index}-cost`] && (
              <p className="text-sm text-destructive">
                {errors[`item-${index}-cost`]}
              </p>
            )}
          </div>

          <div className="space-y-2">
            <label className="text-sm font-medium">Total</label>
            <div className="px-3 py-2 border rounded-md bg-muted text-sm font-medium">
              K{itemTotal.toFixed(2)}
            </div>
          </div>
        </div>
      </div>
    </Card>
  )
}
