'use client';

import { useRouter } from 'next/navigation';
import { ColumnDef } from '@tanstack/react-table';
import Link from 'next/link';
import { useCallback, useMemo, useEffect } from 'react';
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
import { useGRNs } from '@/hooks/use-grns-queries';

interface GrnTableProps {
  userId: string;
  userRole: string;
  refreshTrigger: number;
  onRefresh: () => void;
}

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
  refreshTrigger,
  onRefresh: _onRefresh,
}: GrnTableProps) {
  const router = useRouter();
  const { data: grns = [], refetch } = useGRNs(1, 50); // Get first 50 GRNs

  // Refetch when refreshTrigger changes
  useEffect(() => {
    refetch();
  }, [refreshTrigger, refetch]);

  // Memoize the data to prevent unnecessary re-renders
  // React Query returns a new array reference on each render,
  // so we memoize based on the actual content changes
  const data = useMemo(() => {
    if (grns && grns.length > 0) {
      return grns;
    }
    return [];
  }, [grns]);

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
      data={data}
      actions={getActions}
      hideSearchBar={false}
      renderRowActions={(grn: WorkflowDocument) => (
        <GrnOptionsMenu grn={grn} router={router} />
      )}
    />
  );
}
