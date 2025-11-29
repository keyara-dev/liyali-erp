'use client'

import { useState, useMemo } from 'react'
import {
  ColumnDef,
  VisibilityState,
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  SortingState,
  useReactTable,
} from '@tanstack/react-table'
import Link from 'next/link'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { MoreHorizontal, Download } from 'lucide-react'
import { WorkflowDocument } from '@/types/workflow'

interface GrnTableProps {
  userId: string
  userRole: string
  refreshTrigger: number
  onRefresh: () => void
}

// Mock GRN data
const mockGRNs: WorkflowDocument[] = [
  {
    id: 'grn-1',
    type: 'GRN',
    documentNumber: 'GRN-2024-001',
    status: 'IN_APPROVAL',
    currentStage: 1,
    totalStages: 1,
    createdBy: 'user-1',
    createdAt: new Date('2024-11-27'),
    updatedAt: new Date('2024-11-28'),
    metadata: {
      poId: 'po-1',
      poNumber: 'PO-2024-001',
      vendorName: 'Broadway Ventures',
      receivedQuantity: 5,
      totalQuantity: 5,
      amount: 7500.00,
      receivedDate: '2024-11-27',
    },
  },
  {
    id: 'grn-2',
    type: 'GRN',
    documentNumber: 'GRN-2024-002',
    status: 'APPROVED',
    currentStage: 1,
    totalStages: 1,
    createdBy: 'user-2',
    createdAt: new Date('2024-11-26'),
    updatedAt: new Date('2024-11-27'),
    metadata: {
      poId: 'po-2',
      poNumber: 'PO-2024-002',
      vendorName: 'Tech Solutions Ltd',
      receivedQuantity: 10,
      totalQuantity: 10,
      amount: 12000.00,
      receivedDate: '2024-11-26',
    },
  },
  {
    id: 'grn-3',
    type: 'GRN',
    documentNumber: 'GRN-2024-003',
    status: 'APPROVED',
    currentStage: 1,
    totalStages: 1,
    createdBy: 'user-3',
    createdAt: new Date('2024-11-25'),
    updatedAt: new Date('2024-11-25'),
    metadata: {
      poId: 'po-3',
      poNumber: 'PO-2024-003',
      vendorName: 'Office Supplies Co',
      receivedQuantity: 20,
      totalQuantity: 20,
      amount: 5000.00,
      receivedDate: '2024-11-25',
    },
  },
]

const statusVariants: Record<string, 'default' | 'secondary' | 'destructive' | 'outline'> = {
  DRAFT: 'outline',
  SUBMITTED: 'secondary',
  IN_APPROVAL: 'secondary',
  APPROVED: 'default',
  REJECTED: 'destructive',
}

const columns: ColumnDef<WorkflowDocument>[] = [
  {
    accessorKey: 'documentNumber',
    header: 'GRN Number',
    cell: ({ row }) => (
      <div className="font-medium">{row.original.documentNumber}</div>
    ),
  },
  {
    accessorKey: 'metadata.poNumber',
    header: 'PO Reference',
    cell: ({ row }) => (
      <div className="text-sm">
        <Link
          href={`/workflows/purchase-orders/${row.original.metadata?.poId}`}
          className="text-blue-600 hover:underline"
        >
          {row.original.metadata?.poNumber}
        </Link>
      </div>
    ),
  },
  {
    accessorKey: 'metadata.vendorName',
    header: 'Vendor',
    cell: ({ row }) => <div className="text-sm">{row.original.metadata?.vendorName}</div>,
  },
  {
    accessorKey: 'metadata.amount',
    header: 'Amount',
    cell: ({ row }) => (
      <div className="text-right font-medium">
        K {(row.original.metadata?.amount || 0).toLocaleString()}
      </div>
    ),
  },
  {
    accessorKey: 'status',
    header: 'Status',
    cell: ({ row }) => (
      <Badge variant={statusVariants[row.original.status] || 'outline'}>
        {row.original.status}
      </Badge>
    ),
  },
  {
    accessorKey: 'createdAt',
    header: 'Received Date',
    cell: ({ row }) => (
      <div className="text-sm text-muted-foreground">
        {new Date(row.original.createdAt).toLocaleDateString()}
      </div>
    ),
  },
  {
    id: 'actions',
    cell: ({ row }) => (
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="ghost" className="h-8 w-8 p-0">
            <MoreHorizontal className="h-4 w-4" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <DropdownMenuLabel>Actions</DropdownMenuLabel>
          <Link href={`/workflows/grn/${row.original.id}`}>
            <DropdownMenuItem>View Details</DropdownMenuItem>
          </Link>
          <DropdownMenuSeparator />
          <DropdownMenuItem className="gap-2">
            <Download className="h-4 w-4" />
            Download PDF
          </DropdownMenuItem>
          {row.original.status === 'IN_APPROVAL' && (
            <>
              <DropdownMenuSeparator />
              <DropdownMenuItem className="text-green-600">Approve</DropdownMenuItem>
              <DropdownMenuItem className="text-red-600">Reject</DropdownMenuItem>
            </>
          )}
        </DropdownMenuContent>
      </DropdownMenu>
    ),
  },
]

export function GrnTable({
  userId,
  userRole,
  refreshTrigger,
  onRefresh,
}: GrnTableProps) {
  const [sorting, setSorting] = useState<SortingState>([])
  const [columnVisibility, setColumnVisibility] = useState<VisibilityState>({})

  const data = useMemo(() => mockGRNs, [refreshTrigger])

  const table = useReactTable({
    data,
    columns,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    onSortingChange: setSorting,
    onColumnVisibilityChange: setColumnVisibility,
    state: {
      sorting,
      columnVisibility,
    },
  })

  return (
    <div className="rounded-lg border bg-white">
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
              <TableRow key={row.id} className="hover:bg-gray-50">
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
                No GRNs found
              </TableCell>
            </TableRow>
          )}
        </TableBody>
      </Table>
      <div className="flex items-center justify-between border-t px-6 py-4">
        <div className="text-sm text-muted-foreground">
          Page {table.getState().pagination.pageIndex + 1} of{' '}
          {table.getPageCount()}
        </div>
        <div className="flex gap-2">
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
