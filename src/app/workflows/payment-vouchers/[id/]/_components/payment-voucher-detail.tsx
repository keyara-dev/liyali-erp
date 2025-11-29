'use client'

import { useState, useEffect } from 'react'
import { Download, ChevronLeft, CheckCircle2, XCircle, Link as LinkIcon } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { WorkflowDocument } from '@/types/workflow'
import { ApprovalConfirmationDialog } from '@/components/approval-confirmation-dialog'
import { ApprovalHistory } from '@/components/approval-history'
import { generatePaymentVoucherPDF } from '@/lib/pdf-generators/payment-voucher-pdf'
import Link from 'next/link'

interface PaymentVoucherDetailProps {
  pvId: string
  userId: string
  userRole: string
}

export function PaymentVoucherDetail({
  pvId,
  userId,
  userRole,
}: PaymentVoucherDetailProps) {
  const [pv, setPv] = useState<WorkflowDocument | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [showApprovalDialog, setShowApprovalDialog] = useState(false)
  const [isDownloadingPDF, setIsDownloadingPDF] = useState(false)

  useEffect(() => {
    loadPaymentVoucher()
  }, [pvId])

  const loadPaymentVoucher = async () => {
    setIsLoading(true)
    try {
      // Mock data - will be replaced with API call
      const mockPV: WorkflowDocument = {
        id: pvId,
        type: 'PAYMENT_VOUCHER',
        documentNumber: 'PV-2024-001',
        status: 'IN_APPROVAL',
        currentStage: 2,
        createdBy: 'user-accountant',
        createdAt: new Date('2024-11-25'),
        updatedAt: new Date('2024-11-29'),
        metadata: {
          grnId: 'grn-1',
          poId: 'po-1',
          requisitionId: 'req-1',
          vendorName: 'Broadway Ventures',
          vendorId: 'vendor-1',
          grossAmount: 7500.00,
          tax: 1125.00,
          netAmount: 6375.00,
          paymentMethod: 'BANK_TRANSFER',
          bankInfo: {
            accountNumber: '123456789',
            accountName: 'Broadway Ventures Ltd',
            bankCode: 'ZANACO',
            bankName: 'Zambia National Commercial Bank',
          },
        },
      }
      setPv(mockPV)
    } catch (error) {
      console.error('Error loading payment voucher:', error)
    } finally {
      setIsLoading(false)
    }
  }

  const handleDownloadPDF = async () => {
    setIsDownloadingPDF(true)
    try {
      if (pv) {
        // Call PDF generator with @react-pdf/renderer
        await generatePaymentVoucherPDF(pv)
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
      console.log('Approving PV with data:', data)
      // Refresh the page after approval
      loadPaymentVoucher()
      setShowApprovalDialog(false)
    } catch (error) {
      console.error('Error approving PV:', error)
    }
  }

  if (isLoading) {
    return <div className="text-center py-8">Loading...</div>
  }

  if (!pv) {
    return <div className="text-center py-8 text-red-600">Payment voucher not found</div>
  }

  const canApprove = pv.status === 'IN_APPROVAL'
  const statusVariant =
    pv.status === 'APPROVED'
      ? 'default'
      : pv.status === 'PAID'
        ? 'secondary'
        : pv.status === 'REJECTED'
          ? 'destructive'
          : pv.status === 'IN_APPROVAL'
            ? 'secondary'
            : 'outline'

  const stageNames = [
    'Accountant Generation',
    'Department Head Review',
    'Auditor Review',
    'Finance Director Review',
    'Principal Officer Approval',
  ]

  return (
    <div className="space-y-6">
      {/* Header with back button */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Link href="/workflows/payment-vouchers">
            <Button variant="ghost" size="sm" className="gap-2">
              <ChevronLeft className="h-4 w-4" />
              Back to Payment Vouchers
            </Button>
          </Link>
          <div>
            <h1 className="text-2xl font-bold">{pv.documentNumber}</h1>
            <p className="text-sm text-muted-foreground">
              Created on {new Date(pv.createdAt).toLocaleDateString()}
            </p>
          </div>
        </div>
        <div className="flex items-center gap-3">
          <Badge variant={statusVariant}>{pv.status}</Badge>
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

      {/* Linked Documents Section */}
      <Card className="border-blue-200 bg-blue-50 dark:bg-blue-950/20">
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-blue-900 dark:text-blue-100">
            <LinkIcon className="h-5 w-5" />
            Linked Documents
          </CardTitle>
          <CardDescription className="text-blue-800 dark:text-blue-200">
            Trace the full workflow from requisition through payment
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            {/* Requisition Link */}
            {pv.metadata?.requisitionId && (
              <div className="flex items-center justify-between bg-white dark:bg-slate-900 p-3 rounded border">
                <div>
                  <p className="text-sm text-muted-foreground">Requisition</p>
                  <p className="font-medium">REQ-2024-001</p>
                </div>
                <Link href={`/workflows/requisitions/${pv.metadata.requisitionId}`}>
                  <Button variant="outline" size="sm">
                    View Requisition
                  </Button>
                </Link>
              </div>
            )}

            {/* Purchase Order Link */}
            {pv.metadata?.poId && (
              <div className="flex items-center justify-between bg-white dark:bg-slate-900 p-3 rounded border">
                <div>
                  <p className="text-sm text-muted-foreground">Purchase Order</p>
                  <p className="font-medium">PO-2024-001</p>
                </div>
                <Link href={`/workflows/purchase-orders/${pv.metadata.poId}`}>
                  <Button variant="outline" size="sm">
                    View PO
                  </Button>
                </Link>
              </div>
            )}

            {/* GRN Link */}
            {pv.metadata?.grnId && (
              <div className="flex items-center justify-between bg-white dark:bg-slate-900 p-3 rounded border">
                <div>
                  <p className="text-sm text-muted-foreground">Goods Received Note</p>
                  <p className="font-medium">GRN-2024-001</p>
                </div>
                <Link href={`/workflows/grn/${pv.metadata.grnId}`}>
                  <Button variant="outline" size="sm">
                    View GRN
                  </Button>
                </Link>
              </div>
            )}
          </div>
        </CardContent>
      </Card>

      {/* Main content grid */}
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
        {/* Left column - PV Details (2/3 width) */}
        <div className="lg:col-span-2 space-y-6">
          {/* Vendor & Bank Information */}
          <Card>
            <CardHeader>
              <CardTitle>Vendor & Bank Information</CardTitle>
            </CardHeader>
            <CardContent className="grid grid-cols-2 gap-6">
              <div>
                <p className="text-sm text-muted-foreground">Vendor Name</p>
                <p className="font-medium">{pv.metadata?.vendorName}</p>
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Vendor ID</p>
                <p className="font-medium">{pv.metadata?.vendorId}</p>
              </div>
              <div className="col-span-2 border-t pt-4">
                <p className="text-sm text-muted-foreground mb-2">Bank Details</p>
                <div className="grid grid-cols-2 gap-4 text-sm">
                  <div>
                    <span className="text-muted-foreground">Bank Name:</span>
                    <p className="font-medium">{pv.metadata?.bankInfo?.bankName}</p>
                  </div>
                  <div>
                    <span className="text-muted-foreground">Bank Code:</span>
                    <p className="font-medium">{pv.metadata?.bankInfo?.bankCode}</p>
                  </div>
                  <div>
                    <span className="text-muted-foreground">Account Name:</span>
                    <p className="font-medium">{pv.metadata?.bankInfo?.accountName}</p>
                  </div>
                  <div>
                    <span className="text-muted-foreground">Account No:</span>
                    <p className="font-medium">{pv.metadata?.bankInfo?.accountNumber}</p>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Amount Details */}
          <Card>
            <CardHeader>
              <CardTitle>Amount Summary</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <div className="flex justify-between items-center border-b pb-2">
                <span className="text-muted-foreground">Gross Amount</span>
                <span className="font-medium">K {(pv.metadata?.grossAmount || 0).toLocaleString()}</span>
              </div>
              <div className="flex justify-between items-center border-b pb-2">
                <span className="text-muted-foreground">Tax (15%)</span>
                <span className="font-medium">K {(pv.metadata?.tax || 0).toLocaleString()}</span>
              </div>
              <div className="flex justify-between items-center pt-2 text-lg font-bold">
                <span>Net Amount</span>
                <span className="text-green-600">
                  K {(pv.metadata?.netAmount || 0).toLocaleString()}
                </span>
              </div>
            </CardContent>
          </Card>

          {/* Approval History */}
          <Card>
            <CardHeader>
              <CardTitle>Approval History</CardTitle>
              <CardDescription>
                {pv.currentStage > 0
                  ? `Stage ${pv.currentStage} of 4 - ${stageNames[pv.currentStage] || 'Unknown'}`
                  : 'Draft - Not yet submitted'}
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
                    {pv.currentStage === 0 ? 'Draft' : `Stage ${pv.currentStage}`}
                  </span>
                  {pv.currentStage > 0 && (
                    <span className="text-sm">{stageNames[pv.currentStage]}</span>
                  )}
                </div>
              </div>

              <div className="pt-4 border-t space-y-2">
                <p className="text-sm text-muted-foreground">Document Status</p>
                <Badge variant={statusVariant} className="w-full justify-center py-2">
                  {pv.status}
                </Badge>
              </div>

              {/* Payment Reference if approved */}
              {pv.metadata?.paymentReference && (
                <div className="pt-4 border-t space-y-2">
                  <p className="text-sm text-muted-foreground">Payment Reference</p>
                  <p className="font-mono text-sm font-medium bg-muted p-2 rounded">
                    {pv.metadata.paymentReference}
                  </p>
                </div>
              )}
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
                  This voucher is waiting for your approval at{' '}
                  {stageNames[pv.currentStage] || 'this stage'}.
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
                      {new Date(pv.createdAt).toLocaleString()}
                    </p>
                  </div>
                </div>
                <div className="flex gap-3">
                  <div className="w-2 h-2 rounded-full bg-blue-500 mt-1 flex-shrink-0" />
                  <div>
                    <p className="font-medium">Last Updated</p>
                    <p className="text-xs text-muted-foreground">
                      {new Date(pv.updatedAt).toLocaleString()}
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
        documentId={pv.id}
        documentType="PAYMENT_VOUCHER"
        documentNumber={pv.documentNumber}
        vendor={pv.metadata?.vendorName || ''}
        amount={`K ${(pv.metadata?.netAmount || 0).toLocaleString()}`}
        stageNumber={pv.currentStage}
        totalStages={4}
        stageName={stageNames[pv.currentStage] || 'Unknown'}
        onApprove={handleApproveSubmit}
        onCancel={() => setShowApprovalDialog(false)}
      />
    </div>
  )
}
