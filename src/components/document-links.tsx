'use client'

import { Link as LinkIcon } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { WorkflowDocument } from '@/types/workflow'
import Link from 'next/link'

interface DocumentLinksProps {
  currentDocument: WorkflowDocument
  linkedDocuments?: {
    requisition?: { id: string; number: string }
    purchaseOrder?: { id: string; number: string }
    grn?: { id: string; number: string }
    paymentVoucher?: { id: string; number: string }
  }
}

/**
 * Display linked documents in a workflow chain
 * Shows the path: Requisition → Purchase Order → GRN → Payment Voucher
 */
export function DocumentLinks({
  currentDocument,
  linkedDocuments,
}: DocumentLinksProps) {
  if (!linkedDocuments) {
    return null
  }

  const { requisition, purchaseOrder, grn, paymentVoucher } = linkedDocuments

  // Only show if there are linked documents
  if (!requisition && !purchaseOrder && !grn && !paymentVoucher) {
    return null
  }

  return (
    <Card className="border-blue-200 bg-blue-50 dark:bg-blue-950/20">
      <CardHeader>
        <CardTitle className="flex items-center gap-2 text-blue-900 dark:text-blue-100">
          <LinkIcon className="h-5 w-5" />
          Workflow Chain
        </CardTitle>
        <CardDescription className="text-blue-800 dark:text-blue-200">
          Related documents in this requisition-to-payment process
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-2">
          {/* Requisition */}
          {requisition && (
            <div className="flex items-center justify-between bg-white dark:bg-slate-900 p-3 rounded border">
              <div>
                <p className="text-sm text-muted-foreground">Requisition</p>
                <p className="font-medium">{requisition.number}</p>
              </div>
              <Link href={`/workflows/requisitions/${requisition.id}`}>
                <Button variant="outline" size="sm">
                  View
                </Button>
              </Link>
            </div>
          )}

          {/* Purchase Order */}
          {purchaseOrder && (
            <div className="flex items-center justify-between bg-white dark:bg-slate-900 p-3 rounded border">
              <div>
                <p className="text-sm text-muted-foreground">Purchase Order</p>
                <p className="font-medium">{purchaseOrder.number}</p>
              </div>
              <Link href={`/workflows/purchase-orders/${purchaseOrder.id}`}>
                <Button variant="outline" size="sm">
                  View
                </Button>
              </Link>
            </div>
          )}

          {/* GRN */}
          {grn && (
            <div className="flex items-center justify-between bg-white dark:bg-slate-900 p-3 rounded border">
              <div>
                <p className="text-sm text-muted-foreground">Goods Received Note</p>
                <p className="font-medium">{grn.number}</p>
              </div>
              <Link href={`/workflows/grn/${grn.id}`}>
                <Button variant="outline" size="sm">
                  View
                </Button>
              </Link>
            </div>
          )}

          {/* Payment Voucher */}
          {paymentVoucher && (
            <div className="flex items-center justify-between bg-white dark:bg-slate-900 p-3 rounded border">
              <div>
                <p className="text-sm text-muted-foreground">Payment Voucher</p>
                <p className="font-medium">{paymentVoucher.number}</p>
              </div>
              <Link href={`/workflows/payment-vouchers/${paymentVoucher.id}`}>
                <Button variant="outline" size="sm">
                  View
                </Button>
              </Link>
            </div>
          )}
        </div>

        {/* Legend */}
        <div className="mt-4 p-3 bg-blue-100 dark:bg-blue-900/30 rounded text-sm text-blue-900 dark:text-blue-100">
          <p className="font-medium mb-2">Workflow Process:</p>
          <p>
            Requisition → Purchase Order → Goods Receipt → Payment Voucher
          </p>
        </div>
      </CardContent>
    </Card>
  )
}
