'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { DatePicker } from '@/components/ui/date-picker'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { SearchFilters, WorkflowDocumentType, DocumentStatus } from '@/types/workflow'
import { Search } from 'lucide-react'

interface SearchFormProps {
  onSearch: (filters: SearchFilters) => void
  isSearching: boolean
}

const DOCUMENT_TYPES: { value: string; label: string }[] = [
  { value: 'ALL', label: 'All Document Types' },
  { value: 'REQUISITION', label: 'Requisitions' },
  { value: 'PURCHASE_ORDER', label: 'Purchase Orders' },
  { value: 'PAYMENT_VOUCHER', label: 'Payment Vouchers' },
  { value: 'GOODS_RECEIVED_NOTE', label: 'Goods Received Notes' },
]

const STATUSES: { value: string; label: string }[] = [
  { value: 'ALL', label: 'All Statuses' },
  { value: 'DRAFT', label: 'Draft' },
  { value: 'SUBMITTED', label: 'Submitted' },
  { value: 'IN_APPROVAL', label: 'In Approval' },
  { value: 'APPROVED', label: 'Approved' },
  { value: 'REJECTED', label: 'Rejected' },
  { value: 'REVERSED', label: 'Reversed' },
]

export function SearchForm({ onSearch, isSearching }: SearchFormProps) {
  const [documentNumber, setDocumentNumber] = useState('')
  const [documentType, setDocumentType] = useState('ALL')
  const [status, setStatus] = useState('ALL')
  const [startDate, setStartDate] = useState<Date | undefined>(undefined)
  const [endDate, setEndDate] = useState<Date | undefined>(undefined)

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    onSearch({
      documentNumber,
      documentType: documentType as 'ALL' | WorkflowDocumentType,
      status: status as 'ALL' | DocumentStatus,
      startDate: startDate ? startDate.toISOString().split('T')[0] : '',
      endDate: endDate ? endDate.toISOString().split('T')[0] : '',
    })
  }

  const handleReset = () => {
    setDocumentNumber('')
    setDocumentType('ALL')
    setStatus('ALL')
    setStartDate(undefined)
    setEndDate(undefined)
    onSearch({
      documentNumber: '',
      documentType: 'ALL',
      status: 'ALL',
      startDate: '',
      endDate: '',
    })
  }

  return (
    <Card className="bg-linear-to-br from-primary via-primary/80 to-primary/60 border-0 shadow-lg">
      <CardHeader>
        <CardTitle className="text-lg text-primary-foreground">Search Filters</CardTitle>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-4">
          {/* First Row: Document Number and Type */}
          <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
            <div className="space-y-2">
              <label className="text-sm font-medium text-primary-foreground">Document Number</label>
              <div className="backdrop-blur-md bg-white/10 rounded-lg border border-white/20">
                <Input
                  placeholder="e.g., REQ-2024-001"
                  value={documentNumber}
                  onChange={(e) => setDocumentNumber(e.target.value)}
                  className="bg-transparent border-0 text-white placeholder:text-white/60"
                />
              </div>
            </div>

            <div className="space-y-2">
              <label className="text-sm font-medium text-primary-foreground">Document Type</label>
              <div className="backdrop-blur-md bg-white/10 rounded-lg border border-white/20">
                <Select value={documentType} onValueChange={setDocumentType}>
                  <SelectTrigger className="bg-transparent border-0 text-white">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {DOCUMENT_TYPES.map((type) => (
                      <SelectItem key={type.value} value={type.value}>
                        {type.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>
          </div>

          {/* Second Row: Status and Date Range */}
          <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
            <div className="space-y-2">
              <label className="text-sm font-medium text-primary-foreground">Status</label>
              <div className="backdrop-blur-md bg-white/10 rounded-lg border border-white/20">
                <Select value={status} onValueChange={setStatus}>
                  <SelectTrigger className="bg-transparent border-0 text-white">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {STATUSES.map((s) => (
                      <SelectItem key={s.value} value={s.value}>
                        {s.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>

            <div className="backdrop-blur-md bg-white/10 rounded-lg border border-white/20 p-3">
              <DatePicker
                label="Start Date"
                value={startDate}
                onValueChange={setStartDate}
                classNames={{
                  label: "text-primary-foreground",
                  input: "bg-transparent border-0 text-white placeholder:text-white/60",
                }}
              />
            </div>

            <div className="backdrop-blur-md bg-white/10 rounded-lg border border-white/20 p-3">
              <DatePicker
                label="End Date"
                value={endDate}
                onValueChange={setEndDate}
                classNames={{
                  label: "text-primary-foreground",
                  input: "bg-transparent border-0 text-white placeholder:text-white/60",
                }}
              />
            </div>
          </div>

          {/* Action Buttons */}
          <div className="flex gap-3 pt-2">
            <Button
              type="submit"
              className="gap-2"
              disabled={isSearching}
            >
              <Search className="h-4 w-4" />
              {isSearching ? 'Searching...' : 'Search'}
            </Button>
            <Button
              type="button"
              variant="outline"
              onClick={handleReset}
              disabled={isSearching}
            >
              Reset
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  )
}
