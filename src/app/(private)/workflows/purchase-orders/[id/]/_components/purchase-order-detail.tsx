'use client'

import { useState, useEffect } from 'react'
import { Download, ChevronLeft, CheckCircle2, XCircle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { WorkflowDocument } from '@/types/workflow'
import { ApprovalConfirmationDialog } from '@/components/approval-confirmation-dialog'
import { ApprovalHistory } from '@/components/approval-history'
import { generatePurchaseOrderPDF } from '@/lib/pdf-generators/purchase-order-pdf'
import { DocumentLinks } from '@/components/document-links'
import Link from 'next/link'

interface PurchaseOrderDetailProps {
  poId: string
  userId: string
  userRole: string
}

export function PurchaseOrderDetail({
  poId,
  userId,
  userRole,
}: PurchaseOrderDetailProps) {
  const [po, setPo] = useState<WorkflowDocument | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [showApprovalDialog, setShowApprovalDialog] = useState(false)
  const [isDownloadingPDF, setIsDownloadingPDF] = useState(false)

  useEffect(() => {
    loadPurchaseOrder()
  }, [poId])

  const loadPurchaseOrder = async () => {
    setIsLoading(true)
    try {
      // Mock data - will be replaced with API call
      const mockPO: WorkflowDocument = {
        id: poId,
        type: 'PURCHASE_ORDER',
        documentNumber: 'PO-2024-001',
        status: 'IN_APPROVAL',
        currentStage: 2,
        createdBy: 'user-1',
        createdAt: new Date('2024-11-25'),
        updatedAt: new Date('2024-11-29'),
        metadata: {
          vendorName: 'Broadway Ventures',
          vendorId: 'vendor-1',
          totalAmount: 7500.00,
          deliveryType: 'Standard',
          items: [
            {
              id: 'item-1',
              description: 'Office Furniture',
              quantity: 5,
              unitCost: 1000.00,
              totalCost: 5000.00,
            },
            {
              id: 'item-2',
              description: 'Installation Service',
              quantity: 1,
              unitCost: 2500.00,
              totalCost: 2500.00,
            },
          ],
        },
      }
      setPo(mockPO)
    } catch (error) {
      console.error('Error loading purchase order:', error)
    } finally {
      setIsLoading(false)
    }
  }

  const handleDownloadPDF = async () => {
    setIsDownloadingPDF(true)
    try {
      if (po) {
        // Call PDF generator with @react-pdf/renderer
        await generatePurchaseOrderPDF(po)
      }
    } catch (error) {
      console.error('Error downloading PDF:', error)
    } finally {
      setIsDownloadingPDF(false)
    }
  }

  const handleApproveSubmit = async (data: any) => {
    try {
      // Call approve document action with signature
      console.log('Approving PO with data:', data)
      // Refresh the page after approval
      loadPurchaseOrder()
      setShowApprovalDialog(false)
    } catch (error) {
      console.error('Error approving PO:', error)
    }
  }

  if (isLoading) {
    return <div className="text-center py-8">Loading...</div>
  }

  if (!po) {
    return <div className="text-center py-8 text-red-600">Purchase order not found</div>
  }

  const canApprove = po.status === 'IN_APPROVAL'
  const statusVariant =
    po.status === 'APPROVED'
      ? 'default'
      : po.status === 'REJECTED'
        ? 'destructive'
        : po.status === 'IN_APPROVAL'
          ? 'secondary'
          : 'outline'

  return (
    <div className="space-y-6">
      {/* Header with back button */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Link href="/workflows/purchase-orders">
            <Button variant="ghost" size="sm" className="gap-2">
              <ChevronLeft className="h-4 w-4" />
              Back to Purchase Orders
            </Button>
          </Link>
          <div>
            <h1 className="text-2xl font-bold">{po.documentNumber}</h1>
            <p className="text-sm text-muted-foreground">
              Created on {new Date(po.createdAt).toLocaleDateString()}
            </p>
          </div>
        </div>
        <div className="flex items-center gap-3">
          <Badge variant={statusVariant}>{po.status}</Badge>
          <Button
            variant="outline"
            size="sm"
            onClick={handleDownloadPDF}
            disabled={isDownloadingPDF}
            className="gap-2"
          >
            <Download className="h-4 w-4" />
            {isDownloadingPDF ? 'Generating...' : 'Download PDF'}
          </Button>
        </div>
      </div>

      {/* Main content grid */}
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
        {/* Left column - PO Details (2/3 width) */}
        <div className="lg:col-span-2 space-y-6">
          {/* Vendor Information */}
          <Card>
            <CardHeader>
              <CardTitle>Vendor Information</CardTitle>
            </CardHeader>
            <CardContent className="grid grid-cols-2 gap-6">
              <div>
                <p className="text-sm text-muted-foreground">Vendor Name</p>
                <p className="font-medium">{po.metadata?.vendorName}</p>
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Vendor ID</p>
                <p className="font-medium">{po.metadata?.vendorId}</p>
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Delivery Type</p>
                <p className="font-medium">{po.metadata?.deliveryType}</p>
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Total Amount</p>
                <p className="font-bold text-lg">
                  K {(po.metadata?.totalAmount || 0).toLocaleString()}
                </p>
              </div>
            </CardContent>
          </Card>

          {/* Document Links */}
          <DocumentLinks
            currentDocument={po}
            linkedDocuments={{
              requisition: { id: 'req-1', number: 'REQ-2024-001' },
              grn: undefined,
            }}
          />

          {/* Items Table */}
          <Card>
            <CardHeader>
              <CardTitle>Order Items</CardTitle>
              <CardDescription>
                {(po.metadata?.items || []).length} items in this order
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="overflow-x-auto">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="border-b">
                      <th className="text-left py-2 px-2">Description</th>
                      <th className="text-right py-2 px-2">Quantity</th>
                      <th className="text-right py-2 px-2">Unit Cost</th>
                      <th className="text-right py-2 px-2">Total Cost</th>
                    </tr>
                  </thead>
                  <tbody>
                    {(po.metadata?.items || []).map((item: any) => (
                      <tr key={item.id} className="border-b hover:bg-muted/50">
                        <td className="py-3 px-2">{item.description}</td>
                        <td className="text-right py-3 px-2">{item.quantity}</td>
                        <td className="text-right py-3 px-2">
                          K {item.unitCost.toLocaleString()}
                        </td>
                        <td className="text-right py-3 px-2 font-medium">
                          K {item.totalCost.toLocaleString()}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                  <tfoot>
                    <tr className="font-bold bg-muted/50">
                      <td colSpan={3} className="py-3 px-2 text-right">
                        Total:
                      </td>
                      <td className="text-right py-3 px-2">
                        K {(po.metadata?.totalAmount || 0).toLocaleString()}
                      </td>
                    </tr>
                  </tfoot>
                </table>
              </div>
            </CardContent>
          </Card>

          {/* Approval History */}
          <Card>
            <CardHeader>
              <CardTitle>Approval History</CardTitle>
              <CardDescription>
                Stage {po.currentStage} of 4 - Auditor Review
              </CardDescription>
            </CardHeader>
            <CardContent>
              <ApprovalHistory state={{} as any} />
            </CardContent>
          </Card>
        </div>

        {/* Right column - Status & Actions (1/3 width) */}
        <div className="space-y-6">
          {/* Status Card */}
          <Card>
            <CardHeader>
              <CardTitle>Status</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div>
                <p className="text-sm text-muted-foreground mb-2">Current Stage</p>
                <div className="flex items-center justify-between bg-muted p-3 rounded">
                  <span className="font-medium">
                    Stage {po.currentStage} of 4
                  </span>
                  <span className="text-sm">Auditor Review</span>
                </div>
              </div>

              <div className="pt-4 border-t space-y-2">
                <p className="text-sm text-muted-foreground">Document Status</p>
                <Badge variant={statusVariant} className="w-full justify-center py-2">
                  {po.status}
                </Badge>
              </div>
            </CardContent>
          </Card>

          {/* Approval Actions */}
          {canApprove && (
            <Card className="border-green-200 bg-green-50 dark:bg-green-950/20">
              <CardHeader>
                <CardTitle className="text-green-900 dark:text-green-100">
                  Your Action Required
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-3">
                <p className="text-sm text-green-800 dark:text-green-200">
                  This PO is waiting for your approval at Stage {po.currentStage}.
                </p>
                <div className="grid grid-cols-1 gap-2">
                  <Button
                    onClick={() => setShowApprovalDialog(true)}
                    className="gap-2 bg-green-600 hover:bg-green-700"
                  >
                    <CheckCircle2 className="h-4 w-4" />
                    Approve with Signature
                  </Button>
                  <Button variant="outline" className="gap-2">
                    <XCircle className="h-4 w-4" />
                    Reject
                  </Button>
                </div>
              </CardContent>
            </Card>
          )}

          {/* Timeline */}
          <Card>
            <CardHeader>
              <CardTitle>Timeline</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="text-sm space-y-3">
                <div className="flex gap-3">
                  <div className="w-2 h-2 rounded-full bg-green-500 mt-1 flex-shrink-0" />
                  <div>
                    <p className="font-medium">Created</p>
                    <p className="text-xs text-muted-foreground">
                      {new Date(po.createdAt).toLocaleString()}
                    </p>
                  </div>
                </div>
                <div className="flex gap-3">
                  <div className="w-2 h-2 rounded-full bg-blue-500 mt-1 flex-shrink-0" />
                  <div>
                    <p className="font-medium">Last Updated</p>
                    <p className="text-xs text-muted-foreground">
                      {new Date(po.updatedAt).toLocaleString()}
                    </p>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>

      {/* Approval Dialog */}
      <ApprovalConfirmationDialog
        open={showApprovalDialog}
        documentId={po.id}
        documentType="PURCHASE_ORDER"
        documentNumber={po.documentNumber}
        vendor={po.metadata?.vendorName || ''}
        amount={`K ${(po.metadata?.totalAmount || 0).toLocaleString()}`}
        stageNumber={po.currentStage}
        totalStages={4}
        stageName="Auditor Review"
        onApprove={handleApproveSubmit}
        onCancel={() => setShowApprovalDialog(false)}
      />
    </div>
  )
}
