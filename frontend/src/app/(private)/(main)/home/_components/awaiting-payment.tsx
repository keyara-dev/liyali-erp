'use client';

import { useQuery } from '@tanstack/react-query';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import Link from 'next/link';
import { getPaymentVouchers } from '@/app/_actions/payment-vouchers';
import { PaymentVoucher } from '@/types/payment-voucher';

export function AwaitingPaymentWidget() {
  const { data, isLoading } = useQuery<PaymentVoucher[]>({
    queryKey: ['pv', 'awaiting-payment'],
    queryFn: async () => {
      const response = await getPaymentVouchers(1, 10, {
        status: 'APPROVED',
        hasProofOfPayment: false,
      });
      return response.success ? (response.data ?? []) : [];
    },
    staleTime: 2 * 60 * 1000,
  });

  const items = data ?? [];

  return (
    <Card className="border-0 shadow-sm">
      <CardHeader className="pb-4">
        <div className="flex items-center gap-2">
          <CardTitle className="text-base">Awaiting Payment</CardTitle>
          {!isLoading && items.length > 0 && (
            <Badge variant="secondary" className="text-xs">
              {items.length}
            </Badge>
          )}
        </div>
      </CardHeader>

      <CardContent className="pt-0">
        {isLoading ? (
          <div className="space-y-2">
            {[1, 2, 3].map((i) => (
              <div key={i} className="flex items-center justify-between py-2 border-b last:border-0">
                <Skeleton className="h-4 w-24" />
                <Skeleton className="h-4 w-32" />
                <Skeleton className="h-4 w-16" />
              </div>
            ))}
          </div>
        ) : items.length === 0 ? (
          <p className="text-sm text-muted-foreground py-2">All caught up — no approved PVs awaiting payment.</p>
        ) : (
          <div className="space-y-0">
            {items.map((pv) => (
              <Link
                key={pv.id}
                href={`/payment-vouchers/${pv.id}`}
                className="flex items-center justify-between border-b py-2 last:border-0 hover:bg-muted/40 px-2 -mx-2 rounded transition-colors"
              >
                <span className="font-mono text-sm shrink-0 text-muted-foreground">
                  {pv.documentNumber ?? pv.id.slice(0, 8)}
                </span>
                <span className="text-sm truncate max-w-[140px] mx-2 flex-1">
                  {pv.vendorName}
                </span>
                <span className="text-sm font-medium shrink-0">
                  {pv.currency} {Number(pv.amount).toLocaleString()}
                </span>
                {pv.routingType === 'direct_payment' && (
                  <Badge
                    variant="outline"
                    className="border-purple-500 text-purple-700 dark:text-purple-300 text-xs ml-2 shrink-0"
                  >
                    Direct
                  </Badge>
                )}
              </Link>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
