'use client';

import { useEffect, useState, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import { ColumnDef } from '@tanstack/react-table';
import { ArrowUpDown, Eye, Pencil, CheckCircle2, XCircle, MoreVertical } from 'lucide-react';

import { StatusBadge } from '@/components/status-badge';
import { Button } from '@/components/ui/button';
import { DataTable } from '@/components/ui/data-table';
import { WorkflowDocument } from '@/types/workflow';
import {
  useRequisitionsWithStorage,
  convertRequisitionToWorkflowDocument,
} from '@/hooks/use-requisition-storage';
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
      const amount =
        row.original.metadata?.totalAmount || row.original.metadata?.amount;
      return (
        <div className="font-medium">
          {amount
            ? `ZMW ${amount.toLocaleString('en-ZM', {
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
  req: WorkflowDocument;
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
        {req.status === 'SUBMITTED' && (
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
  const { data: apiRequisitions } = useRequisitionsWithStorage(true);
  const [requisitions, setRequisitions] = useState<WorkflowDocument[]>([]);

  useEffect(() => {
    if (apiRequisitions && apiRequisitions.length > 0) {
      // Convert requisitions to workflow documents and filter by current user
      const workflowDocs = apiRequisitions
        .map((req) => convertRequisitionToWorkflowDocument(req))
        .filter((doc) => doc.createdBy === userId);
      setRequisitions(workflowDocs);
    } else {
      setRequisitions([]);
    }
  }, [apiRequisitions, userId, refreshTrigger]);

  const getActions = useCallback(
    (req: WorkflowDocument): ActionButton[] => {
      return [
        {
          icon: <Eye className="h-3.5 w-3.5" />,
          label: 'View',
          tooltip: 'View Details',
          onClick: () => router.push(`/requisitions/${req.id}`),
        },
        {
          icon: <Pencil className="h-3.5 w-3.5" />,
          label: 'Edit',
          tooltip: 'Edit Requisition',
          onClick: () => router.push(`/requisitions/${req.id}/edit`),
        },
      ];
    },
    [router]
  );

  return (
    <DataTable
      columns={columns}
      data={requisitions}
      searchKey="documentNumber"
      searchPlaceholder="Filter by document number..."
      actions={getActions}
      renderRowActions={(req: WorkflowDocument) => (
        <ReqOptionsMenu req={req} router={router} />
      )}
    />
  );
}
