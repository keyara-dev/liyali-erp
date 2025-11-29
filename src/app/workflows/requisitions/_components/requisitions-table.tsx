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
import { ArrowUpDown, Eye, AlertCircle } from 'lucide-react'

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
import { Input } from '@/components/ui/input'
import { getDocumentsByCreator } from '@/app/_actions/workflow'
import { WorkflowDocument } from '@/types/workflow'

interface RequisitionsTableProps {
  userId: string
  userRole: string
  refreshTrigger: number
}

const STATUS_COLORS: Record<string, string> = {
  DRAFT: 'outline',
  SUBMITTED: 'secondary',
  IN_APPROVAL: 'warning',
  APPROVED: 'success',
  REJECTED: 'destructive',
}

const STATUS_LABELS: Record<string, string> = {
  DRAFT: 'Draft',
  SUBMITTED: 'Submitted',
  IN_APPROVAL: 'In Review',
  APPROVED: 'Approved',
  REJECTED: 'Rejected',
}

export function RequisitionsTable({
  userId,
  userRole,
  refreshTrigger,
}: RequisitionsTableProps) {
  const router = useRouter()
  const [requisitions, setRequisitions] = useState<WorkflowDocument[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [sorting, setSorting] = React.useState<SortingState>([])
  const [columnFilters, setColumnFilters] = React.useState<ColumnFiltersState>([])
  const [columnVisibility, setColumnVisibility] = React.useState<VisibilityState>({})

  useEffect(() => {
    async function fetchRequisitions() {
      setIsLoading(true)
      try {
        const result = await getDocumentsByCreator(userId, 1, 100)
        if (result.success) {
          const filteredReqs = result.data?.data.filter(
            (doc) => doc.type === 'REQUISITION'
          ) || []
          setRequisitions(filteredReqs)
        }
      } catch (error) {
        console.error('Failed to fetch requisitions:', error)
      } finally {
        setIsLoading(false)
      }
    }

    fetchRequisitions()
  }, [userId, refreshTrigger])

  const columns: ColumnDef<WorkflowDocument>[] = [
    {
      accessorKey: 'documentNumber',
      header: ({ column }) => (
        <Button
          variant="ghost"
          onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
          className="-ml-3"
        >
          Document Number
          <ArrowUpDown className="ml-2 h-4 w-4" />
        </Button>
      ),
      cell: ({ row }) => (
        <div className="font-semibold">{row.getValue('documentNumber')}</div>
      ),
    },
    {
      accessorKey: 'metadata.requestedFor',
      header: 'Requested For',
      cell: ({ row }) => (
        <div>{(row.original.metadata as any)?.requestedFor || '-'}</div>
      ),
    },
    {
      accessorKey: 'metadata.department',
      header: 'Department',
      cell: ({ row }) => (
        <div>{(row.original.metadata as any)?.department || '-'}</div>
      ),
    },
    {
      accessorKey: 'totalAmount',
      header: ({ column }) => (
        <Button
          variant="ghost"
          onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
          className="-ml-3"
        >
          Total Amount
          <ArrowUpDown className="ml-2 h-4 w-4" />
        </Button>
      ),
      cell: ({ row }) => {
        const amount = row.original.totalAmount
        return (
          <div className="font-medium">
            {amount ? `${amount.toLocaleString()}` : '-'}
          </div>
        )
      },
    },
    {
      accessorKey: 'status',
      header: 'Status',
      cell: ({ row }) => {
        const status = row.original.status
        const variant = STATUS_COLORS[status] || 'outline'
        const label = STATUS_LABELS[status] || status

        return (
          <Badge variant={variant as any} className="capitalize">
            {label}
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
        const date = new Date(row.original.createdAt)
        return <div className="text-sm text-muted-foreground">{date.toLocaleDateString()}</div>
      },
    },
    {
      id: 'actions',
      enableHiding: false,
      cell: ({ row }) => (
        <div className="flex justify-end">
          <Button
            variant="ghost"
            size="sm"
            onClick={() => router.push(`/workflows/requisitions/${row.original.id}`)}
            className="gap-2"
          >
            <Eye className="h-4 w-4" />
            View
          </Button>
        </div>
      ),
    },
  ]

  const table = useReactTable({
    data: requisitions,
    columns,
    onSortingChange: setSorting,
    onColumnFiltersChange: setColumnFilters,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    onColumnVisibilityChange: setColumnVisibility,
    initialState: {
      pagination: {
        pageSize: 10,
      },
    },
    state: {
      sorting,
      columnFilters,
      columnVisibility,
    },
  })

  if (isLoading) {
    return (
      <div className="rounded-lg border bg-white p-8 text-center">
        <p className="text-muted-foreground">Loading requisitions...</p>
      </div>
    )
  }

  if (requisitions.length === 0) {
    return (
      <div className="rounded-lg border bg-white p-8 text-center">
        <AlertCircle className="mx-auto mb-2 h-8 w-8 text-muted-foreground" />
        <p className="text-muted-foreground">No requisitions found</p>
        <p className="text-sm text-muted-foreground/75">
          Create a new requisition to get started
        </p>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-2">
        <Input
          placeholder="Filter by document number..."
          value={(table.getColumn('documentNumber')?.getFilterValue() as string) ?? ''}
          onChange={(event) =>
            table.getColumn('documentNumber')?.setFilterValue(event.target.value)
          }
          className="max-w-sm"
        />
      </div>
      <div className="rounded-md border">
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
            {table.getRowModel().rows?.length ? (
              table.getRowModel().rows.map((row) => (
                <TableRow key={row.id} data-state={row.getIsSelected() && 'selected'}>
                  {row.getVisibleCells().map((cell) => (
                    <TableCell key={cell.id}>
                      {flexRender(cell.column.columnDef.cell, cell.getContext())}
                    </TableCell>
                  ))}
                </TableRow>
              ))
            ) : (
              <TableRow>
                <TableCell colSpan={columns.length} className="h-24 text-center">
                  No results.
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
      <div className="flex items-center justify-between space-x-2 pt-4">
        <div className="flex-1 text-sm text-muted-foreground">
          {table.getFilteredRowModel().rows.length} requisition(s)
        </div>
        <div className="space-x-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => table.previousPage()}
            disabled={!table.getCanPreviousPage()}
          >
            Previous
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => table.nextPage()}
            disabled={!table.getCanNextPage()}
          >
            Next
          </Button>
        </div>
      </div>
    </div>
  )
}
