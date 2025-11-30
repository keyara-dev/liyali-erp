'use client'

import { useState, useEffect } from 'react'
import { ColumnDef } from '@tanstack/react-table'
import {
  ArrowUpDown,
  MoreHorizontal,
  Download,
  Eye,
  CheckCircle2,
  XCircle,
  Clock,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { DataTable } from '@/components/ui/data-table'
import { StatusBadge as CentralizedStatusBadge } from '@/components/status-badge'
import { WorkflowDocument } from '@/types/workflow'

interface PurchaseOrdersTableProps {
  userId: string
  userRole: string
  refreshTrigger: number
  onRefresh: () => void
}

// Stage indicator
function StageIndicator({
  currentStage,
  totalStages,
}: {
  currentStage: number
  totalStages: number
}) {
  return (
    <div className="flex items-center gap-1">
      <span className="text-sm font-medium">{currentStage}</span>
      <span className="text-xs text-muted-foreground">of {totalStages}</span>
    </div>
  )
}

// Columns definition
function getColumns(onViewClick: (id: string) => void): ColumnDef<WorkflowDocument>[] {
  return [
    {
      id: 'documentNumber',
      accessorKey: 'documentNumber',
      header: ({ column }) => (
        <Button
          variant="ghost"
          onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
          className="h-8 p-0"
        >
          PO Number
          <ArrowUpDown className="ml-2 h-4 w-4" />
        </Button>
      ),
      cell: ({ row }) => (
        <div className="font-medium">{row.getValue('documentNumber')}</div>
      ),
    },
    {
      id: 'vendor',
      accessorKey: 'metadata.vendorName',
      header: 'Vendor',
      cell: ({ row }) => (
        <div>{row.original.metadata?.vendorName || '-'}</div>
      ),
    },
    {
      id: 'amount',
      accessorKey: 'metadata.totalAmount',
      header: ({ column }) => (
        <Button
          variant="ghost"
          onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
          className="h-8 p-0"
        >
          Amount
          <ArrowUpDown className="ml-2 h-4 w-4" />
        </Button>
      ),
      cell: ({ row }) => (
        <div className="font-medium">
          K {(row.original.metadata?.totalAmount || 0).toLocaleString()}
        </div>
      ),
    },
    {
      id: 'status',
      accessorKey: 'status',
      header: 'Status',
      cell: ({ row }) => <CentralizedStatusBadge status={row.getValue('status')} type="document" />,
    },
    {
      id: 'stage',
      accessorKey: 'currentStage',
      header: 'Stage',
      cell: ({ row }) => (
        <StageIndicator
          currentStage={row.original.currentStage || 1}
          totalStages={4}
        />
      ),
    },
    {
      id: 'createdDate',
      accessorKey: 'createdAt',
      header: ({ column }) => (
        <Button
          variant="ghost"
          onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
          className="h-8 p-0"
        >
          Created
          <ArrowUpDown className="ml-2 h-4 w-4" />
        </Button>
      ),
      cell: ({ row }) => (
        <div className="text-sm">
          {new Date(row.original.createdAt).toLocaleDateString()}
        </div>
      ),
    },
    {
      id: 'actions',
      header: 'Actions',
      cell: ({ row }) => (
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" className="h-8 w-8 p-0">
              <span className="sr-only">Open menu</span>
              <MoreHorizontal className="h-4 w-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuLabel>Actions</DropdownMenuLabel>
            <DropdownMenuItem
              onClick={() => onViewClick(row.original.id)}
              className="flex items-center gap-2"
            >
              <Eye className="h-4 w-4" />
              View Details
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem className="flex items-center gap-2">
              <Download className="h-4 w-4" />
              Download PDF
            </DropdownMenuItem>
            {row.original.status === 'IN_APPROVAL' && (
              <>
                <DropdownMenuSeparator />
                <DropdownMenuItem className="flex items-center gap-2 text-green-600">
                  <CheckCircle2 className="h-4 w-4" />
                  Approve
                </DropdownMenuItem>
                <DropdownMenuItem className="flex items-center gap-2 text-red-600">
                  <XCircle className="h-4 w-4" />
                  Reject
                </DropdownMenuItem>
              </>
            )}
          </DropdownMenuContent>
        </DropdownMenu>
      ),
    },
  ]
}

export function PurchaseOrdersTable({
  userId,
  userRole,
  refreshTrigger,
  onRefresh,
}: PurchaseOrdersTableProps) {
  const [data, setData] = useState<WorkflowDocument[]>([])
  const [isLoading, setIsLoading] = useState(false)

  useEffect(() => {
    // Load purchase orders from mock data
    loadPurchaseOrders()
  }, [refreshTrigger])

  const loadPurchaseOrders = async () => {
    setIsLoading(true)
    try {
      // Mock data - will be replaced with API call
      const mockPOs: WorkflowDocument[] = [
        {
          id: 'po-1',
          type: 'PURCHASE_ORDER',
          documentNumber: 'PO-2024-001',
          status: 'IN_APPROVAL',
          currentStage: 2,
          createdBy: 'user-1',
          createdAt: new Date('2024-11-25'),
          updatedAt: new Date('2024-11-29'),
          metadata: {
            vendorName: 'Broadway Ventures',
            vendorId: 'vendor-1',
            totalAmount: 7500.00,
            deliveryType: 'Standard',
          },
        },
        {
          id: 'po-2',
          type: 'PURCHASE_ORDER',
          documentNumber: 'PO-2024-002',
          status: 'APPROVED',
          currentStage: 4,
          createdBy: 'user-2',
          createdAt: new Date('2024-11-20'),
          updatedAt: new Date('2024-11-28'),
          metadata: {
            vendorName: 'Tech Solutions Ltd',
            vendorId: 'vendor-2',
            totalAmount: 15000.00,
            deliveryType: 'Express',
          },
        },
        {
          id: 'po-3',
          type: 'PURCHASE_ORDER',
          documentNumber: 'PO-2024-003',
          status: 'IN_APPROVAL',
          currentStage: 1,
          createdBy: 'user-1',
          createdAt: new Date('2024-11-29'),
          updatedAt: new Date('2024-11-29'),
          metadata: {
            vendorName: 'Office Supplies Co',
            vendorId: 'vendor-3',
            totalAmount: 2500.00,
            deliveryType: 'Standard',
          },
        },
      ]
      setData(mockPOs)
    } catch (error) {
      console.error('Error loading purchase orders:', error)
    } finally {
      setIsLoading(false)
    }
  }

  const handleViewClick = (id: string) => {
    // Navigate to detail page
    window.location.href = `/workflows/purchase-orders/${id}`
  }

  const columns = getColumns(handleViewClick)

  return (
    <div className="space-y-4">
      <DataTable columns={columns} data={data} />
    </div>
  )
}
