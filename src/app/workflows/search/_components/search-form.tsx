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
  const [startDate, setStartDate] = useState('')
  const [endDate, setEndDate] = useState('')

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    onSearch({
      documentNumber,
      documentType: documentType as 'ALL' | WorkflowDocumentType,
      status: status as 'ALL' | DocumentStatus,
      startDate,
      endDate,
    })
  }

  const handleReset = () => {
    setDocumentNumber('')
    setDocumentType('ALL')
    setStatus('ALL')
    setStartDate('')
    setEndDate('')
    onSearch({
      documentNumber: '',
      documentType: 'ALL',
      status: 'ALL',
      startDate: '',
      endDate: '',
    })
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-lg">Search Filters</CardTitle>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-4">
          {/* First Row: Document Number and Type */}
          <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
            <div className="space-y-2">
              <label className="text-sm font-medium">Document Number</label>
              <Input
                placeholder="e.g., REQ-2024-001"
                value={documentNumber}
                onChange={(e) => setDocumentNumber(e.target.value)}
              />
            </div>

            <div className="space-y-2">
              <label className="text-sm font-medium">Document Type</label>
              <Select value={documentType} onValueChange={setDocumentType}>
                <SelectTrigger>
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

          {/* Second Row: Status and Date Range */}
          <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
            <div className="space-y-2">
              <label className="text-sm font-medium">Status</label>
              <Select value={status} onValueChange={setStatus}>
                <SelectTrigger>
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

            <div className="space-y-2">
              <label className="text-sm font-medium">Start Date</label>
              <Input
                type="date"
                value={startDate}
                onChange={(e) => setStartDate(e.target.value)}
              />
            </div>

            <div className="space-y-2">
              <label className="text-sm font-medium">End Date</label>
              <Input
                type="date"
                value={endDate}
                onChange={(e) => setEndDate(e.target.value)}
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
