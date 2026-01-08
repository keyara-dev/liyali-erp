'use client';

import { useCallback, useMemo, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { ColumnDef } from '@tanstack/react-table';
import { ArrowUpDown, Eye, Pencil, CheckCircle2, XCircle, MoreVertical } from 'lucide-react';

import { StatusBadge } from '@/components/status-badge';
import { Button } from '@/components/ui/button';
import { DataTable } from '@/components/ui/data-table';
import { Requisition } from '@/types/requisition';
import { useRequisitions } from '@/hooks/use-requisition-queries';
import type { ActionButton } from '@/components/ui/action-buttons';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';

interface RequisitionsTableProps {
  userId: string;
  userRole: string;
  refreshTrigger: number;
}

const columns: ColumnDef<Requisition>[] = [
  {
    accessorKey: 'id',
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
      <div className="font-semibold">{row.original.id}</div>
    ),
  },
  {
    accessorKey: 'requesterName',
    header: 'Requested By',
    cell: ({ row }) => (
      <div>{row.original.requesterName || '-'}</div>
    ),
  },
  {
    accessorKey: 'department',
    header: 'Department',
    cell: ({ row }) => (
      <div>{row.original.department || '-'}</div>
    ),
  },
  {
    id: 'totalAmount',
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
      const amount = row.original.totalAmount;
      return (
        <div className="font-medium">
          {amount
            ? `${row.original.currency} ${amount.toLocaleString('en-ZM', {
                minimumFractionDigits: 2,
                maximumFractionDigits: 2,
              })}`
            : '-'}
        </div>
      );
    },
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
      const date = new Date(row.original.createdAt);
      return (
        <div className="text-sm text-muted-foreground">
          {date.toLocaleDateString()}
        </div>
      );
    },
  },
];

// Options dropdown component
function ReqOptionsMenu({
  req,
  router,
}: {
  req: Requisition;
  router: ReturnType<typeof useRouter>;
}) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <button className="h-8 w-8 rounded-md border border-input bg-background px-2 py-1.5 hover:bg-accent hover:text-accent-foreground">
          <MoreVertical className="h-4 w-4" />
        </button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        {req.status === 'submitted' && (
          <>
            <DropdownMenuItem onClick={() => console.log('Approve requisition:', req.id)}>
              <CheckCircle2 className="mr-2 h-4 w-4 text-green-600" />
              Approve
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => console.log('Reject requisition:', req.id)}>
              <XCircle className="mr-2 h-4 w-4 text-red-600" />
              Reject
            </DropdownMenuItem>
          </>
        )}
        {req.status === 'draft' && (
          <DropdownMenuItem 
            onClick={() => console.log('Delete requisition:', req.id)}
            className="text-red-600 focus:text-red-600"
          >
            <XCircle className="mr-2 h-4 w-4" />
            Delete
          </DropdownMenuItem>
        )}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

export function RequisitionsTable({
  userId,
  userRole: _userRole,
  refreshTrigger,
}: RequisitionsTableProps) {
  const router = useRouter();
  const { data: requisitions = [], refetch } =
    useRequisitions(1, 50); // Get first 50 requisitions

  // Refetch when refreshTrigger changes
  useEffect(() => {
    refetch();
  }, [refreshTrigger, refetch]);

  // Memoize the data to prevent unnecessary re-renders
  // React Query returns a new array reference on each render,
  // so we memoize based on the actual content changes
  const data = useMemo(() => {
    if (requisitions && requisitions.length > 0) {
      return requisitions;
    }
    return [];
  }, [requisitions]);

  const getActions = useCallback(
    (req: Requisition): ActionButton[] => {
      const actions: ActionButton[] = [
        {
          icon: <Eye className="h-3.5 w-3.5" />,
          label: 'View',
          tooltip: 'View Details',
          onClick: () => router.push(`/requisitions/${req.id}`),
        },
      ];

      // Only allow edit and delete for draft status
      if (req.status === 'draft') {
        actions.push(
          {
            icon: <Pencil className="h-3.5 w-3.5" />,
            label: 'Edit',
            tooltip: 'Edit Requisition',
            onClick: () => router.push(`/requisitions/${req.id}/edit`),
          }
        );
      }

      return actions;
    },
    [router]
  );

  return (
    <DataTable
      columns={columns}
      data={data}
      searchKey="id"
      searchPlaceholder="Filter by document number..."
      actions={getActions}
      renderRowActions={(req: Requisition) => (
        <ReqOptionsMenu req={req} router={router} />
      )}
    />
  );
}
