'use client';

import { useState, useEffect, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import { ColumnDef } from '@tanstack/react-table';
import {
  ArrowUpDown,
  Download,
  Eye,
  Pencil,
  CheckCircle2,
  XCircle,
  MoreVertical,
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { DataTable } from '@/components/ui/data-table';
import { StatusBadge as CentralizedStatusBadge } from '@/components/status-badge';
import { usePaymentVouchersAsWorkflowDocuments } from '@/hooks/use-payment-voucher-storage';
import { WorkflowDocument } from '@/types/workflow';
import type { ActionButton } from '@/components/ui/action-buttons';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';

interface PaymentVouchersTableProps {
  userId: string;
  userRole: string;
  refreshTrigger: number;
  onRefresh: () => void;
}

// Stage indicator
function StageIndicator({
  currentStage,
  totalStages,
}: {
  currentStage: number;
  totalStages: number;
}) {
  return (
    <div className="flex items-center gap-1">
      <span className="text-sm font-medium">{currentStage}</span>
      <span className="text-xs text-muted-foreground">of {totalStages}</span>
    </div>
  );
}

// Columns definition
const columns: ColumnDef<WorkflowDocument>[] = [
  {
    id: 'voucherNumber',
    accessorKey: 'documentNumber',
    header: ({ column }) => (
      <Button
        variant="ghost"
        onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
        className="h-8 p-0"
      >
        Voucher No.
        <ArrowUpDown className="ml-2 h-4 w-4" />
      </Button>
    ),
    cell: ({ row }) => (
      <div className="font-medium">{row.getValue('voucherNumber')}</div>
    ),
  },
  {
    id: 'vendor',
    accessorKey: 'metadata.payeeName',
    header: 'Payee',
    cell: ({ row }) => <div>{row.original.metadata?.payeeName || '-'}</div>,
  },
  {
    id: 'amount',
    accessorKey: 'metadata.amount',
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
        {row.original.metadata?.currency || 'ZMW'}{' '}
        {(row.original.metadata?.amount || 0).toLocaleString()}
      </div>
    ),
  },
  {
    id: 'status',
    accessorKey: 'status',
    header: 'Status',
    cell: ({ row }) => (
      <CentralizedStatusBadge
        status={row.getValue('status')}
        type="document"
      />
    ),
  },
  {
    id: 'stage',
    accessorKey: 'currentStage',
    header: 'Stage',
    cell: ({ row }) => (
      <StageIndicator
        currentStage={row.original.currentStage || 1}
        totalStages={3}
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
];

// Options dropdown component
function PvOptionsMenu({
  pv,
  router,
}: {
  pv: WorkflowDocument;
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
        <DropdownMenuItem onClick={() => console.log('Download PDF for PV:', pv.id)}>
          <Download className="mr-2 h-4 w-4" />
          Download
        </DropdownMenuItem>
        {pv.status === 'IN_REVIEW' && (
          <>
            <DropdownMenuItem onClick={() => console.log('Approve voucher:', pv.id)}>
              <CheckCircle2 className="mr-2 h-4 w-4 text-green-600" />
              Approve
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => console.log('Reject voucher:', pv.id)}>
              <XCircle className="mr-2 h-4 w-4 text-red-600" />
              Reject
            </DropdownMenuItem>
          </>
        )}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

export function PaymentVouchersTable({
  userId: _userId,
  userRole: _userRole,
  refreshTrigger,
  onRefresh: _onRefresh,
}: PaymentVouchersTableProps) {
  const router = useRouter();
  const { data: paymentVouchers } =
    usePaymentVouchersAsWorkflowDocuments(true);
  const [data, setData] = useState<WorkflowDocument[]>([]);

  useEffect(() => {
    if (paymentVouchers && paymentVouchers.length > 0) {
      // Filter by current user's payment vouchers
      const userPVs = paymentVouchers.filter((pv) => pv.createdBy === _userId);
      setData(userPVs);
    } else {
      setData([]);
    }
  }, [paymentVouchers, _userId, refreshTrigger]);

  const getActions = useCallback(
    (pv: WorkflowDocument): ActionButton[] => {
      return [
        {
          icon: <Eye className="h-3.5 w-3.5" />,
          label: 'View',
          tooltip: 'View Details',
          onClick: () => router.push(`/payment-vouchers/${pv.id}`),
        },
        {
          icon: <Pencil className="h-3.5 w-3.5" />,
          label: 'Edit',
          tooltip: 'Edit Voucher',
          onClick: () => router.push(`/payment-vouchers/${pv.id}/edit`),
        },
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
      renderRowActions={(pv: WorkflowDocument) => (
        <PvOptionsMenu pv={pv} router={router} />
      )}
    />
  );
}
