'use client';

import { useQuery } from '@tanstack/react-query';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import Link from 'next/link';
import { getRequisitions } from '@/app/_actions/requisitions';
import { Requisition } from '@/types/requisition';

const STEPS = ['Req submitted', 'PO created', 'PV approved', 'Paid'] as const;

function currentStepIndex(req: Requisition): number {
  // Access linked PV data if available via metadata
  const linkedPV = (req as any).linkedPV as { status?: string } | undefined;
  const linkedPO = (req as any).linkedPO as { id?: string } | undefined;

  if (linkedPV?.status === 'PAID') return 3;
  if (linkedPV?.status === 'APPROVED') return 2;
  if (linkedPO?.id) return 1;
  return 0;
}

interface DirectPaymentPipelineProps {
  userId: string;
}

export function DirectPaymentPipeline({ userId: _userId }: DirectPaymentPipelineProps) {
  const { data, isLoading } = useQuery<Requisition[]>({
    queryKey: ['requisitions', 'direct-payment', _userId],
    queryFn: async () => {
      const response = await getRequisitions(1, 20);
      return response.success ? (response.data ?? []) : [];
    },
    staleTime: 2 * 60 * 1000,
  });

  // Filter client-side for direct_payment routing type
  const items = (data ?? []).filter(
    (r) => r.routingType === 'direct_payment',
  );

  if (isLoading) {
    return (
      <Card className="border-0 shadow-sm">
        <CardHeader className="pb-4">
          <CardTitle className="text-base">Direct Payments</CardTitle>
        </CardHeader>
        <CardContent className="pt-0 space-y-3">
          {[1, 2].map((i) => (
            <div key={i} className="flex items-center gap-3">
              <Skeleton className="h-4 w-20" />
              <div className="flex-1 grid grid-cols-4 gap-1">
                {[1, 2, 3, 4].map((j) => (
                  <Skeleton key={j} className="h-2 rounded" />
                ))}
              </div>
              <Skeleton className="h-4 w-24" />
            </div>
          ))}
        </CardContent>
      </Card>
    );
  }

  if (items.length === 0) return null;

  return (
    <Card className="border-0 shadow-sm">
      <CardHeader className="pb-4">
        <CardTitle className="text-base">Direct Payments</CardTitle>
      </CardHeader>
      <CardContent className="pt-0 space-y-2">
        {items.map((req) => {
          const idx = currentStepIndex(req);
          return (
            <Link
              key={req.id}
              href={`/requisitions/${req.id}`}
              className="flex items-center gap-3 py-2 px-2 -mx-2 rounded hover:bg-muted/40 transition-colors"
            >
              <span className="font-mono text-sm shrink-0 text-muted-foreground min-w-[80px]">
                {req.documentNumber}
              </span>
              <div
                className="flex-1 grid gap-1"
                style={{ gridTemplateColumns: `repeat(${STEPS.length}, 1fr)` }}
              >
                {STEPS.map((s, i) => (
                  <div
                    key={s}
                    className={`h-2 rounded transition-colors ${i <= idx ? 'bg-emerald-500' : 'bg-muted'}`}
                    title={s}
                  />
                ))}
              </div>
              <span className="text-xs text-muted-foreground shrink-0 min-w-[90px] text-right">
                {STEPS[idx]}
              </span>
            </Link>
          );
        })}
      </CardContent>
    </Card>
  );
}
