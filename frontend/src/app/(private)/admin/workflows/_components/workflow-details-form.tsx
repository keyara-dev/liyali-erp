'use client'

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Checkbox } from '@/components/ui/checkbox'
import type { WorkflowFormData } from '../create/_components/create-workflow-client'

interface WorkflowDetailsFormProps {
  data: WorkflowFormData
  onChange: (key: keyof WorkflowFormData, value: any) => void
  errors: Record<string, string>
}

const DOCUMENT_TYPES = [
  { id: 'REQUISITION', name: 'Requisition' },
  { id: 'PURCHASE_ORDER', name: 'Purchase Order' },
  { id: 'PAYMENT_VOUCHER', name: 'Payment Voucher' },
  { id: 'GOODS_RECEIVED_NOTE', name: 'Goods Received Note' },
  { id: 'BUDGET', name: 'Budget' },
]

export function WorkflowDetailsForm({
  data,
  onChange,
  errors,
}: WorkflowDetailsFormProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Workflow Details</CardTitle>
      </CardHeader>
      <CardContent className="space-y-6">
        {/* Name */}
        <div className="space-y-2">
          <label className="text-sm font-medium">
            Workflow Name <span className="text-destructive">*</span>
          </label>
          <Input
            placeholder="e.g., Standard Requisition Approval"
            value={data.name}
            onChange={(e) => onChange('name', e.target.value)}
            className={errors.name ? 'border-destructive' : ''}
          />
          {errors.name && (
            <p className="text-sm text-destructive">{errors.name}</p>
          )}
        </div>

        {/* Description */}
        <div className="space-y-2">
          <label className="text-sm font-medium">Description</label>
          <Textarea
            placeholder="Describe the purpose and use case for this workflow..."
            value={data.description}
            onChange={(e) => onChange('description', e.target.value)}
            rows={3}
          />
        </div>

        {/* Document Type */}
        <div className="space-y-2">
          <label className="text-sm font-medium">
            Applies To <span className="text-destructive">*</span>
          </label>
          <Select value={data.documentType} onValueChange={(value) => onChange('documentType', value)}>
            <SelectTrigger className={errors.documentType ? 'border-destructive' : ''}>
              <SelectValue placeholder="Select document type" />
            </SelectTrigger>
            <SelectContent>
              {DOCUMENT_TYPES.map((type) => (
                <SelectItem key={type.id} value={type.id}>
                  {type.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          {errors.documentType && (
            <p className="text-sm text-destructive">{errors.documentType}</p>
          )}
        </div>

        {/* Set as Default */}
        <div className="flex items-center gap-3 p-4 border rounded-lg bg-muted/30">
          <Checkbox
            id="isDefault"
            checked={data.isDefault}
            onCheckedChange={(checked) => onChange('isDefault', checked)}
          />
          <label htmlFor="isDefault" className="text-sm font-medium cursor-pointer">
            Set as default workflow for {data.documentType ? DOCUMENT_TYPES.find(t => t.id === data.documentType)?.name : 'selected document type'}
          </label>
        </div>
      </CardContent>
    </Card>
  )
}
