'use client';

import { useRouter } from 'next/navigation';
import { ColumnDef } from '@tanstack/react-table';
import Link from 'next/link';
import { useCallback } from 'react';
import { DataTable } from '@/components/ui/data-table';
import { StatusBadge } from '@/components/status-badge';
import { Download, Eye, Pencil, Trash2, MoreVertical } from 'lucide-react';
import { WorkflowDocument } from '@/types/workflow';
import type { ActionButton } from '@/components/ui/action-buttons';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';

interface GrnTableProps {
  userId: string;
  userRole: string;
  refreshTrigger: number;
  onRefresh: () => void;
}

// Mock GRN data
const mockGRNs: WorkflowDocument[] = [
  {
    id: 'grn-1',
    type: 'GOODS_RECEIVED_NOTE',
    documentNumber: 'GRN-2024-001',
    status: 'IN_REVIEW',
    currentStage: 1,
    createdBy: 'user-1',
    createdAt: new Date('2024-11-27'),
    updatedAt: new Date('2024-11-28'),
    metadata: {
      poId: 'po-1',
      poNumber: 'PO-2024-001',
      vendorName: 'Broadway Ventures',
      receivedQuantity: 5,
      totalQuantity: 5,
      amount: 7500.0,
      receivedDate: '2024-11-27',
    },
  },
  {
    id: 'grn-2',
    type: 'GOODS_RECEIVED_NOTE',
    documentNumber: 'GRN-2024-002',
    status: 'APPROVED',
    currentStage: 1,
    createdBy: 'user-2',
    createdAt: new Date('2024-11-26'),
    updatedAt: new Date('2024-11-27'),
    metadata: {
      poId: 'po-2',
      poNumber: 'PO-2024-002',
      vendorName: 'Tech Solutions Ltd',
      receivedQuantity: 10,
      totalQuantity: 10,
      amount: 12000.0,
      receivedDate: '2024-11-26',
    },
  },
  {
    id: 'grn-3',
    type: 'GOODS_RECEIVED_NOTE',
    documentNumber: 'GRN-2024-003',
    status: 'APPROVED',
    currentStage: 1,
    createdBy: 'user-3',
    createdAt: new Date('2024-11-25'),
    updatedAt: new Date('2024-11-25'),
    metadata: {
      poId: 'po-3',
      poNumber: 'PO-2024-003',
      vendorName: 'Office Supplies Co',
      receivedQuantity: 20,
      totalQuantity: 20,
      amount: 5000.0,
      receivedDate: '2024-11-25',
    },
  },
];

// Options dropdown component
function GrnOptionsMenu({ grn, router }: { grn: WorkflowDocument; router: ReturnType<typeof useRouter> }) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <button className="h-8 w-8 rounded-md border border-input bg-background px-2 py-1.5 hover:bg-accent hover:text-accent-foreground">
          <MoreVertical className="h-4 w-4" />
        </button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuItem onClick={() => console.log('Download PDF for GRN:', grn.id)}>
          <Download className="mr-2 h-4 w-4" />
          Download
        </DropdownMenuItem>
        {grn.status === 'IN_REVIEW' && (
          <>
            <DropdownMenuItem onClick={() => console.log('Approve GRN:', grn.id)}>
              <div className="mr-2 h-4 w-4 text-green-600">✓</div>
              Approve
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => console.log('Reject GRN:', grn.id)}>
              <div className="mr-2 h-4 w-4 text-red-600">✕</div>
              Reject
            </DropdownMenuItem>
          </>
        )}
        {grn.status !== 'APPROVED' && (
          <DropdownMenuItem onClick={() => console.log('Delete GRN:', grn.id)} className="text-destructive">
            <Trash2 className="mr-2 h-4 w-4" />
            Delete
          </DropdownMenuItem>
        )}
      </DropdownMenuContent>
    </DropdownMenu>
  );
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
          href={`/purchase-orders/${row.original.metadata?.poId}`}
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
    cell: ({ row }) => (
      <div className="text-sm">{row.original.metadata?.vendorName}</div>
    ),
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
      <StatusBadge status={row.original.status} type="document" />
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
];

export function GrnTable({
  userId: _userId,
  userRole: _userRole,
  refreshTrigger: _refreshTrigger,
  onRefresh: _onRefresh,
}: GrnTableProps) {
  const router = useRouter();

  const getActions = useCallback(
    (grn: WorkflowDocument): ActionButton[] => {
      return [
        {
          icon: <Eye className="h-3.5 w-3.5" />,
          label: 'View',
          tooltip: 'View Details',
          onClick: () => router.push(`/grn/${grn.id}`),
        },
        ...(grn.status !== 'APPROVED'
          ? [
              {
                icon: <Pencil className="h-3.5 w-3.5" />,
                label: 'Edit',
                tooltip: 'Edit GRN',
                onClick: () => router.push(`/grn/${grn.id}/edit`),
              },
            ]
          : []),
      ];
    },
    [router]
  );

  return (
    <DataTable
      columns={columns}
      data={mockGRNs}
      actions={getActions}
      hideSearchBar={false}
      renderRowActions={(grn: WorkflowDocument) => (
        <GrnOptionsMenu grn={grn} router={router} />
      )}
    />
  );
}
