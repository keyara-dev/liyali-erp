'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { PurchaseOrder } from '@/types/purchase-order'
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
} from 'lucide-react'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { usePurchaseOrderStorage } from '@/hooks/use-purchase-order-storage'
import { usePurchaseOrderById } from '@/hooks/use-purchase-order-queries'
import { POActionHistoryPanel } from '../../../_components/po-action-history-panel'
import { POApprovalActionPanel } from '../../../_components/po-approval-action-panel'

interface PODetailClientProps {
  poId: string
  initialPO?: PurchaseOrder
}

export function PODetailClient({ poId, initialPO }: PODetailClientProps) {
  const router = useRouter()
  const { saveToStorage } = usePurchaseOrderStorage()
  const { data: purchaseOrder, isLoading, refetch } = usePurchaseOrderById(poId, initialPO)
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

  if (!purchaseOrder) {
    return (
      <Alert variant="destructive">
        <AlertCircle className="h-4 w-4" />
        <AlertDescription>
          Purchase order not found. Please check the ID and try again.
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
              {purchaseOrder.poNumber}
            </h1>
            <p className="text-muted-foreground">
              {purchaseOrder.title}
            </p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Badge className={getStatusColor(purchaseOrder.status)}>
            {purchaseOrder.status}
          </Badge>
          <Badge className={getPriorityColor(purchaseOrder.priority)}>
            {purchaseOrder.priority}
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
            <p className="font-semibold text-sm">{purchaseOrder.vendorName || 'TBD'}</p>
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
              {purchaseOrder.currency} {purchaseOrder.totalAmount.toLocaleString()}
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
            <p className="font-semibold text-sm">{purchaseOrder.items?.length || 0} items</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground flex items-center gap-2">
              <Calendar className="h-4 w-4" />
              Required By
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="font-semibold text-sm">
              {new Date(purchaseOrder.requiredByDate).toLocaleDateString()}
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
                Purchase Order Details
              </CardTitle>
              <CardDescription>
                Information about this purchase order
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid gap-4 md:grid-cols-2">
                <div>
                  <h4 className="text-sm font-semibold text-muted-foreground mb-1">
                    PO Number
                  </h4>
                  <p className="font-mono text-sm">{purchaseOrder.poNumber}</p>
                </div>
                <div>
                  <h4 className="text-sm font-semibold text-muted-foreground mb-1">
                    Department
                  </h4>
                  <p className="text-sm">{purchaseOrder.department}</p>
                </div>
                <div>
                  <h4 className="text-sm font-semibold text-muted-foreground mb-1">
                    Requested By
                  </h4>
                  <p className="text-sm">{purchaseOrder.requestedByName}</p>
                </div>
                <div>
                  <h4 className="text-sm font-semibold text-muted-foreground mb-1">
                    Created Date
                  </h4>
                  <p className="text-sm">
                    {new Date(purchaseOrder.createdAt).toLocaleDateString()}
                  </p>
                </div>
              </div>

              {purchaseOrder.description && (
                <div>
                  <h4 className="text-sm font-semibold text-muted-foreground mb-2">
                    Description
                  </h4>
                  <p className="text-sm bg-muted p-3 rounded">
                    {purchaseOrder.description}
                  </p>
                </div>
              )}

              {purchaseOrder.createdFromRequisition && purchaseOrder.sourceRequisitionNumber && (
                <div>
                  <h4 className="text-sm font-semibold text-muted-foreground mb-2">
                    Source Requisition
                  </h4>
                  <p className="text-sm bg-blue-50 p-3 rounded border border-blue-200">
                    {purchaseOrder.sourceRequisitionNumber}
                  </p>
                </div>
              )}
            </CardContent>
          </Card>

          {/* Line Items */}
          {purchaseOrder.items && purchaseOrder.items.length > 0 && (
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Package className="h-5 w-5" />
                  Line Items ({purchaseOrder.items.length})
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
                    {purchaseOrder.items.map((item) => (
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
                          {purchaseOrder.currency} {item.unitPrice.toLocaleString()}
                        </TableCell>
                        <TableCell className="text-right text-sm font-mono font-semibold">
                          {purchaseOrder.currency} {item.totalPrice.toLocaleString()}
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
                <div className="flex justify-end mt-4 pt-4 border-t">
                  <div className="text-right">
                    <p className="text-sm text-muted-foreground mb-1">Total Amount</p>
                    <p className="text-xl font-bold font-mono">
                      {purchaseOrder.currency} {purchaseOrder.totalAmount.toLocaleString()}
                    </p>
                  </div>
                </div>
              </CardContent>
            </Card>
          )}

          {/* Approval Action */}
          {purchaseOrder.status === 'IN_REVIEW' && (
            <POApprovalActionPanel
              poId={poId}
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
                  {new Date(purchaseOrder.createdAt).toLocaleString()}
                </p>
              </div>

              {purchaseOrder.submittedAt && (
                <div>
                  <h4 className="font-semibold text-muted-foreground mb-1">
                    Submitted
                  </h4>
                  <p className="flex items-center gap-2">
                    <Calendar className="h-4 w-4 text-muted-foreground" />
                    {new Date(purchaseOrder.submittedAt).toLocaleString()}
                  </p>
                </div>
              )}

              {purchaseOrder.approvedAt && (
                <div>
                  <h4 className="font-semibold text-muted-foreground mb-1">
                    Approved
                  </h4>
                  <p className="flex items-center gap-2">
                    <Calendar className="h-4 w-4 text-muted-foreground" />
                    {new Date(purchaseOrder.approvedAt).toLocaleString()}
                  </p>
                </div>
              )}

              {purchaseOrder.rejectedAt && (
                <div>
                  <h4 className="font-semibold text-muted-foreground mb-1">
                    Rejected
                  </h4>
                  <p className="flex items-center gap-2">
                    <Calendar className="h-4 w-4 text-muted-foreground" />
                    {new Date(purchaseOrder.rejectedAt).toLocaleString()}
                  </p>
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      </div>

      {/* Action History and Approval Chain */}
      {(purchaseOrder.actionHistory || purchaseOrder.approvalChain) && (
        <POActionHistoryPanel
          actionHistory={purchaseOrder.actionHistory}
          approvalChain={purchaseOrder.approvalChain}
        />
      )}
    </div>
  )
}
