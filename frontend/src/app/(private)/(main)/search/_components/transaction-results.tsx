'use client'

import { useState } from 'react'
import * as React from 'react'
import { ColumnDef } from '@tanstack/react-table'
import { useRouter } from 'next/navigation'
import { ArrowUpDown, Eye } from 'lucide-react'

import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { DataTable } from '@/components/ui/data-table'
import { Skeleton } from '@/components/ui/skeleton'
import {
  WorkflowDocument,
  SearchFilters,
} from '@/types/workflow'
import { DownloadButton } from './download-button'
import { useSearchDocuments } from '@/hooks/use-search-queries'

// Table skeleton loader
function TransactionTableSkeleton() {
  return (
    <div className="rounded-md border overflow-hidden">
      <div className="space-y-2 p-4">
        {/* Header row */}
        <div className="flex gap-4 pb-4 border-b">
          {Array.from({ length: 5 }).map((_, i) => (
            <Skeleton key={i} className="h-4 flex-1" />
          ))}
        </div>
        {/* Data rows */}
        {Array.from({ length: 5 }).map((_, rowIdx) => (
          <div key={rowIdx} className="flex gap-4 py-3 border-b last:border-0">
            {Array.from({ length: 5 }).map((_, colIdx) => (
              <Skeleton key={colIdx} className="h-4 flex-1" />
            ))}
          </div>
        ))}
      </div>
    </div>
  )
}

interface TransactionResultsProps {
  filters: SearchFilters;
  refreshTrigger: number;
  userRole: string;
  onSearchComplete?: () => void;
}

const STATUS_COLORS: Record<string, string> = {
  DRAFT: 'outline',
  SUBMITTED: 'secondary',
  IN_REVIEW: 'default',
  APPROVED: 'default',
  REJECTED: 'destructive',
  REVERSED: 'secondary',
}

const STATUS_LABELS: Record<string, string> = {
  DRAFT: 'Draft',
  SUBMITTED: 'Submitted',
  IN_REVIEW: 'In Approval',
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
  onSearchComplete,
}: TransactionResultsProps) {
  const router = useRouter()
  const [currentPage, setCurrentPage] = useState(1)
  const pageSize = 10

  // Use React Query hook for searching
  const {
    data,
    isLoading,
    isError,
    error,
  } = useSearchDocuments(filters, currentPage, pageSize)

  // Handle search completion callback
  React.useEffect(() => {
    if (!isLoading) {
      onSearchComplete?.()
    }
  }, [isLoading, onSearchComplete])

  // Get documents and pagination info
  const documents = data?.documents || []
  const totalDocuments = data?.total || 0
  const totalPages = data?.totalPages || 0

  const columns: ColumnDef<WorkflowDocument>[] = [
    {
      accessorKey: "documentNumber",
      header: ({ column }) => (
        <Button
          variant="ghost"
          onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
          className="-ml-3"
        >
          Document #
          <ArrowUpDown className="ml-2 h-4 w-4" />
        </Button>
      ),
      cell: ({ row }) => (
        <span className="font-medium text-primary">
          {row.getValue("documentNumber")}
        </span>
      ),
    },
    {
      accessorKey: "type",
      header: "Type",
      cell: ({ row }) => {
        const type = row.getValue("type") as string;
        return (
          <span className="text-sm">{DOCUMENT_TYPE_LABELS[type] || type}</span>
        );
      },
    },
    {
      accessorKey: "status",
      header: "Status",
      cell: ({ row }) => {
        const status = row.getValue("status") as string;
        return (
          <Badge variant={STATUS_COLORS[status] as any}>
            {STATUS_LABELS[status] || status}
          </Badge>
        );
      },
    },
    {
      accessorKey: "createdAt",
      header: ({ column }) => (
        <Button
          variant="ghost"
          onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
          className="-ml-3"
        >
          Created
          <ArrowUpDown className="ml-2 h-4 w-4" />
        </Button>
      ),
      cell: ({ row }) => {
        const date = new Date(row.getValue("createdAt"));
        return (
          <span className="text-sm text-muted-foreground">
            {date.toLocaleDateString()}
          </span>
        );
      },
    },
  ];

  // Show loading state
  if (isLoading) {
    return (
      <Card>
        <CardContent className="pt-6">
          <div className="space-y-4">
            <div className="text-sm text-muted-foreground mb-4">
              Searching documents...
            </div>
            <TransactionTableSkeleton />
          </div>
        </CardContent>
      </Card>
    )
  }

  // Show error state
  if (isError) {
    return (
      <Card>
        <CardContent className="pt-6">
          <div className="text-center py-12">
            <p className="text-red-600 font-medium mb-2">Search Failed</p>
            <p className="text-sm text-muted-foreground">
              {error?.message || 'Unable to search documents. Please try again.'}
            </p>
          </div>
        </CardContent>
      </Card>
    )
  }

  // Show empty state
  if (documents.length === 0 && !isLoading) {
    return (
      <Card>
        <CardContent className="pt-6">
          <div className="text-center py-12">
            <p className="text-muted-foreground">No documents found</p>
            <p className="text-sm text-muted-foreground mt-1">
              Try adjusting your search filters
            </p>
          </div>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardContent className="pt-6">
        <div className="space-y-4">
          <DataTable
            columns={columns}
            data={documents}
            hideSearchBar={true}
            totalCount={totalDocuments}
            currentPage={currentPage}
            pageSize={pageSize}
            totalPages={totalPages}
            onPaginationChange={(page) => {
              setCurrentPage(page)
            }}
            renderRowActions={(doc: WorkflowDocument) => {
              // Map document type to URL slug
              const typeSlug =
                {
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
                    onClick={() => router.push(`/${typeSlug}/${doc.id}`)}
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
            }}
          />
        </div>
      </CardContent>
    </Card>
  )
}
