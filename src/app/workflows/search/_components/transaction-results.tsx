'use client'

import { useEffect, useState } from 'react'
import * as React from 'react'
import {
  ColumnDef,
  ColumnFiltersState,
  SortingState,
  VisibilityState,
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  useReactTable,
} from '@tanstack/react-table'
import { useRouter } from 'next/navigation'
import { ArrowUpDown, Eye, ChevronLeft, ChevronRight } from 'lucide-react'

import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { searchDocuments } from '@/app/_actions/search'
import { WorkflowDocument, SearchFilters, PaginatedResponse } from '@/types/workflow'
import { DownloadButton } from './download-button'

interface TransactionResultsProps {
  filters: SearchFilters
  refreshTrigger: number
  userRole: string
}

const STATUS_COLORS: Record<string, string> = {
  DRAFT: 'outline',
  SUBMITTED: 'secondary',
  IN_APPROVAL: 'default',
  APPROVED: 'default',
  REJECTED: 'destructive',
  REVERSED: 'secondary',
}

const STATUS_LABELS: Record<string, string> = {
  DRAFT: 'Draft',
  SUBMITTED: 'Submitted',
  IN_APPROVAL: 'In Review',
  APPROVED: 'Approved',
  REJECTED: 'Rejected',
  REVERSED: 'Reversed',
}

const DOCUMENT_TYPE_LABELS: Record<string, string> = {
  REQUISITION: 'Requisition',
  PURCHASE_ORDER: 'Purchase Order',
  PAYMENT_VOUCHER: 'Payment Voucher',
  GOODS_RECEIVED_NOTE: 'GRN',
}

export function TransactionResults({
  filters,
  refreshTrigger,
  userRole,
}: TransactionResultsProps) {
  const router = useRouter()
  const [documents, setDocuments] = useState<WorkflowDocument[]>([])
  const [pagination, setPagination] = useState({
    page: 1,
    limit: 10,
    total: 0,
    totalPages: 1,
  })
  const [isLoading, setIsLoading] = useState(false)
  const [sorting, setSorting] = React.useState<SortingState>([])
  const [columnFilters, setColumnFilters] = React.useState<ColumnFiltersState>([])
  const [columnVisibility, setColumnVisibility] = React.useState<VisibilityState>({})

  // Fetch documents when filters or page changes
  useEffect(() => {
    async function fetchDocuments() {
      setIsLoading(true)
      try {
        const result = await searchDocuments(filters, pagination.page, pagination.limit)
        if (result.success && result.data) {
          setDocuments(result.data.data)
          setPagination(result.data.pagination)
        }
      } catch (error) {
        console.error('Failed to fetch documents:', error)
      } finally {
        setIsLoading(false)
      }
    }

    fetchDocuments()
  }, [filters, pagination.page, pagination.limit, refreshTrigger])

  const columns: ColumnDef<WorkflowDocument>[] = [
    {
      accessorKey: 'documentNumber',
      header: ({ column }) => (
        <Button
          variant="ghost"
          onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
          className="-ml-3"
        >
          Document #
          <ArrowUpDown className="ml-2 h-4 w-4" />
        </Button>
      ),
      cell: ({ row }) => (
        <span className="font-medium text-primary hover:underline cursor-pointer">
          {row.getValue('documentNumber')}
        </span>
      ),
    },
    {
      accessorKey: 'type',
      header: 'Type',
      cell: ({ row }) => {
        const type = row.getValue('type') as string
        return (
          <span className="text-sm">
            {DOCUMENT_TYPE_LABELS[type] || type}
          </span>
        )
      },
    },
    {
      accessorKey: 'status',
      header: 'Status',
      cell: ({ row }) => {
        const status = row.getValue('status') as string
        return (
          <Badge variant={STATUS_COLORS[status] as any}>
            {STATUS_LABELS[status] || status}
          </Badge>
        )
      },
    },
    {
      accessorKey: 'createdAt',
      header: ({ column }) => (
        <Button
          variant="ghost"
          onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
          className="-ml-3"
        >
          Created
          <ArrowUpDown className="ml-2 h-4 w-4" />
        </Button>
      ),
      cell: ({ row }) => {
        const date = new Date(row.getValue('createdAt'))
        return (
          <span className="text-sm text-muted-foreground">
            {date.toLocaleDateString()} {date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
          </span>
        )
      },
    },
    {
      id: 'actions',
      header: 'Actions',
      cell: ({ row }) => {
        const doc = row.original
        // Map document type to URL slug
        const typeSlug = {
          REQUISITION: 'requisitions',
          PURCHASE_ORDER: 'purchase-orders',
          PAYMENT_VOUCHER: 'payment-vouchers',
          GOODS_RECEIVED_NOTE: 'grn',
        }[doc.type] || 'workflows'

        return (
          <div className="flex gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => router.push(`/workflows/${typeSlug}/${doc.id}`)}
              className="gap-1"
            >
              <Eye className="h-4 w-4" />
              View
            </Button>
            <DownloadButton
              documentId={doc.id}
              documentNumber={doc.documentNumber}
            />
          </div>
        )
      },
    },
  ]

  const table = useReactTable({
    data: documents,
    columns,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    onSortingChange: setSorting,
    onColumnFiltersChange: setColumnFilters,
    onColumnVisibilityChange: setColumnVisibility,
    state: {
      sorting,
      columnFilters,
      columnVisibility,
    },
  })

  return (
    <Card>
      <CardContent className="pt-6">
        {isLoading ? (
          <div className="flex justify-center items-center py-8">
            <div className="text-muted-foreground">Loading documents...</div>
          </div>
        ) : documents.length === 0 ? (
          <div className="flex justify-center items-center py-8">
            <div className="text-muted-foreground">No documents found matching your search criteria.</div>
          </div>
        ) : (
          <>
            {/* Table */}
            <div className="rounded-md border overflow-hidden">
              <Table>
                <TableHeader>
                  {table.getHeaderGroups().map((headerGroup) => (
                    <TableRow key={headerGroup.id}>
                      {headerGroup.headers.map((header) => (
                        <TableHead key={header.id}>
                          {header.isPlaceholder
                            ? null
                            : flexRender(header.column.columnDef.header, header.getContext())}
                        </TableHead>
                      ))}
                    </TableRow>
                  ))}
                </TableHeader>
                <TableBody>
                  {table.getRowModel().rows.map((row) => (
                    <TableRow key={row.id}>
                      {row.getVisibleCells().map((cell) => (
                        <TableCell key={cell.id}>
                          {flexRender(cell.column.columnDef.cell, cell.getContext())}
                        </TableCell>
                      ))}
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>

            {/* Pagination */}
            <div className="flex items-center justify-between gap-2 py-4">
              <div className="text-sm text-muted-foreground">
                Showing {documents.length > 0 ? (pagination.page - 1) * pagination.limit + 1 : 0} to{' '}
                {Math.min(pagination.page * pagination.limit, pagination.total)} of {pagination.total} documents
              </div>
              <div className="flex gap-1">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setPagination((p) => ({ ...p, page: Math.max(p.page - 1, 1) }))}
                  disabled={pagination.page === 1 || isLoading}
                >
                  <ChevronLeft className="h-4 w-4" />
                </Button>
                <div className="flex items-center gap-1">
                  <span className="text-sm">
                    Page {pagination.page} of {pagination.totalPages}
                  </span>
                </div>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setPagination((p) => ({ ...p, page: Math.min(p.page + 1, p.totalPages) }))}
                  disabled={pagination.page >= pagination.totalPages || isLoading}
                >
                  <ChevronRight className="h-4 w-4" />
                </Button>
              </div>
            </div>
          </>
        )}
      </CardContent>
    </Card>
  )
}
