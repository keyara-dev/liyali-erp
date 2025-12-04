'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { PaymentVoucher } from '@/types/payment-voucher'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  AlertCircle,
  ArrowLeft,
  Calendar,
  DollarSign,
  FileText,
  Package,
  User,
  CreditCard,
} from 'lucide-react'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { usePaymentVoucherStorage } from '@/hooks/use-payment-voucher-storage'
import { usePaymentVoucherById } from '@/hooks/use-payment-voucher-queries'
import { PVActionHistoryPanel } from '../../../_components/pv-action-history-panel'
import { PVApprovalActionPanel } from '../../../_components/pv-approval-action-panel'

interface PVDetailClientProps {
  pvId: string
  initialPV?: PaymentVoucher
  userId?: string
  userRole?: string
}

export function PVDetailClient({ pvId, initialPV, userId, userRole }: PVDetailClientProps) {
  const router = useRouter()
  const { saveToStorage } = usePaymentVoucherStorage()
  const { data: paymentVoucher, isLoading, refetch } = usePaymentVoucherById(pvId, initialPV)
  const [isRefreshing, setIsRefreshing] = useState(false)

  const handleRefresh = async () => {
    setIsRefreshing(true)
    try {
      await refetch()
    } finally {
      setIsRefreshing(false)
    }
  }

  if (isLoading) {
    return (
      <div className="space-y-4">
        <div className="h-12 bg-gray-200 rounded animate-pulse" />
        <div className="h-96 bg-gray-200 rounded animate-pulse" />
      </div>
    )
  }

  if (!paymentVoucher) {
    return (
      <Alert variant="destructive">
        <AlertCircle className="h-4 w-4" />
        <AlertDescription>
          Payment voucher not found. Please check the ID and try again.
        </AlertDescription>
      </Alert>
    )
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'DRAFT':
        return 'bg-gray-100 text-gray-800'
      case 'SUBMITTED':
        return 'bg-blue-100 text-blue-800'
      case 'IN_REVIEW':
        return 'bg-yellow-100 text-yellow-800'
      case 'APPROVED':
        return 'bg-green-100 text-green-800'
      case 'REJECTED':
        return 'bg-red-100 text-red-800'
      case 'PAID':
        return 'bg-emerald-100 text-emerald-800'
      default:
        return 'bg-gray-100 text-gray-800'
    }
  }

  const getPriorityColor = (priority: string) => {
    switch (priority) {
      case 'URGENT':
        return 'bg-red-100 text-red-800'
      case 'HIGH':
        return 'bg-orange-100 text-orange-800'
      case 'MEDIUM':
        return 'bg-blue-100 text-blue-800'
      case 'LOW':
        return 'bg-green-100 text-green-800'
      default:
        return 'bg-gray-100 text-gray-800'
    }
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between gap-4">
        <div className="flex items-center gap-4">
          <Button
            variant="outline"
            size="sm"
            onClick={() => router.back()}
            className="gap-2"
          >
            <ArrowLeft className="h-4 w-4" />
            Back
          </Button>
          <div>
            <h1 className="text-3xl font-bold tracking-tight">
              {paymentVoucher.pvNumber}
            </h1>
            <p className="text-muted-foreground">
              {paymentVoucher.title}
            </p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Badge className={getStatusColor(paymentVoucher.status)}>
            {paymentVoucher.status}
          </Badge>
          <Badge className={getPriorityColor(paymentVoucher.priority)}>
            {paymentVoucher.priority}
          </Badge>
        </div>
      </div>

      {/* Quick Info Cards */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground flex items-center gap-2">
              <User className="h-4 w-4" />
              Vendor
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="font-semibold text-sm">{paymentVoucher.vendorName || 'TBD'}</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground flex items-center gap-2">
              <DollarSign className="h-4 w-4" />
              Total Amount
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="font-semibold text-sm">
              {paymentVoucher.currency} {paymentVoucher.totalAmount.toLocaleString()}
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground flex items-center gap-2">
              <Package className="h-4 w-4" />
              Items
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="font-semibold text-sm">{paymentVoucher.items?.length || 0} items</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground flex items-center gap-2">
              <Calendar className="h-4 w-4" />
              Due Date
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="font-semibold text-sm">
              {new Date(paymentVoucher.paymentDueDate).toLocaleDateString()}
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Main Content */}
      <div className="grid gap-6 lg:grid-cols-3">
        {/* Left Column */}
        <div className="lg:col-span-2 space-y-6">
          {/* Details */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <FileText className="h-5 w-5" />
                Payment Voucher Details
              </CardTitle>
              <CardDescription>
                Information about this payment voucher
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid gap-4 md:grid-cols-2">
                <div>
                  <h4 className="text-sm font-semibold text-muted-foreground mb-1">
                    PV Number
                  </h4>
                  <p className="font-mono text-sm">{paymentVoucher.pvNumber}</p>
                </div>
                <div>
                  <h4 className="text-sm font-semibold text-muted-foreground mb-1">
                    Department
                  </h4>
                  <p className="text-sm">{paymentVoucher.department}</p>
                </div>
                <div>
                  <h4 className="text-sm font-semibold text-muted-foreground mb-1">
                    Requested By
                  </h4>
                  <p className="text-sm">{paymentVoucher.requestedByName}</p>
                </div>
                <div>
                  <h4 className="text-sm font-semibold text-muted-foreground mb-1">
                    Created Date
                  </h4>
                  <p className="text-sm">
                    {new Date(paymentVoucher.createdAt).toLocaleDateString()}
                  </p>
                </div>
              </div>

              {paymentVoucher.description && (
                <div>
                  <h4 className="text-sm font-semibold text-muted-foreground mb-2">
                    Description
                  </h4>
                  <p className="text-sm bg-muted p-3 rounded">
                    {paymentVoucher.description}
                  </p>
                </div>
              )}

              {paymentVoucher.sourcePurchaseOrderNumber && (
                <div>
                  <h4 className="text-sm font-semibold text-muted-foreground mb-2">
                    Source Purchase Order
                  </h4>
                  <p className="text-sm bg-blue-50 p-3 rounded border border-blue-200">
                    {paymentVoucher.sourcePurchaseOrderNumber}
                  </p>
                </div>
              )}
            </CardContent>
          </Card>

          {/* Payment Information */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <CreditCard className="h-5 w-5" />
                Payment Information
              </CardTitle>
              <CardDescription>
                Payment method and bank details
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid gap-4 md:grid-cols-2">
                <div>
                  <h4 className="text-sm font-semibold text-muted-foreground mb-1">
                    Payment Method
                  </h4>
                  <p className="text-sm">{paymentVoucher.paymentMethod}</p>
                </div>
                {paymentVoucher.paymentMethod === 'BANK_TRANSFER' && paymentVoucher.bankDetails && (
                  <>
                    <div>
                      <h4 className="text-sm font-semibold text-muted-foreground mb-1">
                        Bank Name
                      </h4>
                      <p className="text-sm">{paymentVoucher.bankDetails.bankName || '-'}</p>
                    </div>
                    <div>
                      <h4 className="text-sm font-semibold text-muted-foreground mb-1">
                        Account Name
                      </h4>
                      <p className="text-sm">{paymentVoucher.bankDetails.accountName || '-'}</p>
                    </div>
                    <div>
                      <h4 className="text-sm font-semibold text-muted-foreground mb-1">
                        Account Number
                      </h4>
                      <p className="text-sm font-mono">{paymentVoucher.bankDetails.accountNumber || '-'}</p>
                    </div>
                  </>
                )}
              </div>

              {paymentVoucher.status === 'PAID' && (
                <div className="mt-4 p-3 bg-green-50 border border-green-200 rounded">
                  <p className="text-sm font-semibold text-green-900 mb-2">
                    Payment Confirmed
                  </p>
                  <div className="space-y-1 text-sm text-green-800">
                    <p>Amount Paid: {paymentVoucher.currency} {(paymentVoucher.paidAmount || 0).toLocaleString()}</p>
                    {paymentVoucher.paidDate && (
                      <p>Date: {new Date(paymentVoucher.paidDate).toLocaleDateString()}</p>
                    )}
                    {paymentVoucher.referenceNumber && (
                      <p>Reference: {paymentVoucher.referenceNumber}</p>
                    )}
                  </div>
                </div>
              )}
            </CardContent>
          </Card>

          {/* Line Items */}
          {paymentVoucher.items && paymentVoucher.items.length > 0 && (
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Package className="h-5 w-5" />
                  Line Items ({paymentVoucher.items.length})
                </CardTitle>
              </CardHeader>
              <CardContent>
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead className="w-12">#</TableHead>
                      <TableHead>Description</TableHead>
                      <TableHead>Category</TableHead>
                      <TableHead className="text-right">Qty</TableHead>
                      <TableHead className="text-right">Unit Price</TableHead>
                      <TableHead className="text-right">Total</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {paymentVoucher.items.map((item) => (
                      <TableRow key={item.id}>
                        <TableCell className="font-mono text-sm">
                          {item.itemNumber}
                        </TableCell>
                        <TableCell>
                          <div>
                            <p className="font-medium text-sm">{item.description}</p>
                            {item.notes && (
                              <p className="text-xs text-muted-foreground">{item.notes}</p>
                            )}
                          </div>
                        </TableCell>
                        <TableCell className="text-sm">{item.category}</TableCell>
                        <TableCell className="text-right text-sm">
                          {item.quantity} {item.unit}
                        </TableCell>
                        <TableCell className="text-right text-sm font-mono">
                          {paymentVoucher.currency} {item.unitPrice.toLocaleString()}
                        </TableCell>
                        <TableCell className="text-right text-sm font-mono font-semibold">
                          {paymentVoucher.currency} {item.totalPrice.toLocaleString()}
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
                <div className="flex justify-end mt-4 pt-4 border-t">
                  <div className="text-right">
                    <p className="text-sm text-muted-foreground mb-1">Total Amount</p>
                    <p className="text-xl font-bold font-mono">
                      {paymentVoucher.currency} {paymentVoucher.totalAmount.toLocaleString()}
                    </p>
                  </div>
                </div>
              </CardContent>
            </Card>
          )}

          {/* Approval Action */}
          {paymentVoucher.status === 'IN_REVIEW' && (
            <PVApprovalActionPanel
              pvId={pvId}
              onApprovalComplete={() => {
                handleRefresh()
              }}
            />
          )}
        </div>

        {/* Right Column - Timeline and Info */}
        <div className="space-y-6">
          {/* Timeline */}
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Timeline</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3 text-sm">
              <div>
                <h4 className="font-semibold text-muted-foreground mb-1">
                  Created
                </h4>
                <p className="flex items-center gap-2">
                  <Calendar className="h-4 w-4 text-muted-foreground" />
                  {new Date(paymentVoucher.createdAt).toLocaleString()}
                </p>
              </div>

              {paymentVoucher.submittedAt && (
                <div>
                  <h4 className="font-semibold text-muted-foreground mb-1">
                    Submitted
                  </h4>
                  <p className="flex items-center gap-2">
                    <Calendar className="h-4 w-4 text-muted-foreground" />
                    {new Date(paymentVoucher.submittedAt).toLocaleString()}
                  </p>
                </div>
              )}

              {paymentVoucher.approvedAt && (
                <div>
                  <h4 className="font-semibold text-muted-foreground mb-1">
                    Approved
                  </h4>
                  <p className="flex items-center gap-2">
                    <Calendar className="h-4 w-4 text-muted-foreground" />
                    {new Date(paymentVoucher.approvedAt).toLocaleString()}
                  </p>
                </div>
              )}

              {paymentVoucher.rejectedAt && (
                <div>
                  <h4 className="font-semibold text-muted-foreground mb-1">
                    Rejected
                  </h4>
                  <p className="flex items-center gap-2">
                    <Calendar className="h-4 w-4 text-muted-foreground" />
                    {new Date(paymentVoucher.rejectedAt).toLocaleString()}
                  </p>
                </div>
              )}

              {paymentVoucher.paidAt && (
                <div>
                  <h4 className="font-semibold text-muted-foreground mb-1">
                    Paid
                  </h4>
                  <p className="flex items-center gap-2">
                    <Calendar className="h-4 w-4 text-muted-foreground" />
                    {new Date(paymentVoucher.paidAt).toLocaleString()}
                  </p>
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      </div>

      {/* Action History and Approval Chain */}
      {(paymentVoucher.actionHistory || paymentVoucher.approvalChain) && (
        <PVActionHistoryPanel
          actionHistory={paymentVoucher.actionHistory}
          approvalChain={paymentVoucher.approvalChain}
        />
      )}
    </div>
  )
}
