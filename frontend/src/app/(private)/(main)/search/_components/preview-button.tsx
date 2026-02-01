'use client'

import { useState } from 'react'
import dynamic from 'next/dynamic'
import { Button } from '@/components/ui/button'
import { Eye, Loader2 } from 'lucide-react'
import { getRequisitionById } from '@/app/_actions/requisitions'
import { getPurchaseOrderById } from '@/app/_actions/purchase-orders'
import { getPaymentVoucherById } from '@/app/_actions/payment-vouchers'
import { getGRNAction } from '@/app/_actions/grn-actions'
import {
  getRequisitionPDFBlob,
  getPurchaseOrderPDFBlob,
  getPaymentVoucherPDFBlob,
  getGrnPDFBlob,
  exportRequisitionPDF,
  exportPurchaseOrderPDF,
  exportPaymentVoucherPDF,
  exportGrnPDF,
  downloadBlob,
} from '@/lib/pdf/pdf-export'

// Dynamic import to avoid SSR issues with react-pdf
const PDFPreviewDialog = dynamic(
  () => import('@/components/modals/pdf-preview-dialog').then((mod) => mod.PDFPreviewDialog),
  { ssr: false }
)

interface PreviewButtonProps {
  documentId: string
  documentNumber: string
  documentType: string
}

export function PreviewButton({ documentId, documentNumber, documentType }: PreviewButtonProps) {
  const [isLoading, setIsLoading] = useState(false)
  const [isPreviewOpen, setIsPreviewOpen] = useState(false)
  const [pdfBlob, setPdfBlob] = useState<Blob | null>(null)
  const [documentData, setDocumentData] = useState<any>(null)

  // Normalize document type to uppercase
  const normalizedType = documentType?.toUpperCase() || ''

  // Check if preview is supported for this document type
  const isPreviewSupported = ['REQUISITION', 'PURCHASE_ORDER', 'PAYMENT_VOUCHER', 'GRN', 'GOODS_RECEIVED_NOTE'].includes(normalizedType)

  const fetchDocumentAndGeneratePDF = async (): Promise<{ blob: Blob | null; data: any }> => {
    try {
      let docData: any = null
      let blob: Blob | null = null

      switch (normalizedType) {
        case 'REQUISITION': {
          const result = await getRequisitionById(documentId)
          if (result.success && result.data) {
            docData = result.data
            blob = await getRequisitionPDFBlob(result.data)
          }
          break
        }
        case 'PURCHASE_ORDER':
        case 'PO': {
          const result = await getPurchaseOrderById(documentId)
          if (result.success && result.data) {
            docData = result.data
            blob = await getPurchaseOrderPDFBlob(result.data)
          }
          break
        }
        case 'PAYMENT_VOUCHER':
        case 'PV': {
          const result = await getPaymentVoucherById(documentId)
          if (result.success && result.data) {
            docData = result.data
            blob = await getPaymentVoucherPDFBlob(result.data)
          }
          break
        }
        case 'GRN':
        case 'GOODS_RECEIVED_NOTE': {
          const result = await getGRNAction(documentId)
          if (result.success && result.data) {
            docData = result.data
            blob = await getGrnPDFBlob(result.data)
          }
          break
        }
        default:
          console.error('Unsupported document type for PDF:', normalizedType)
          return { blob: null, data: null }
      }

      return { blob, data: docData }
    } catch (error) {
      console.error('Error generating PDF:', error)
      return { blob: null, data: null }
    }
  }

  const handlePreview = async () => {
    setIsLoading(true)
    try {
      const { blob, data } = await fetchDocumentAndGeneratePDF()
      if (blob) {
        setPdfBlob(blob)
        setDocumentData(data)
        setIsPreviewOpen(true)
      } else {
        alert('Failed to generate PDF preview')
      }
    } catch (error) {
      console.error('Error previewing PDF:', error)
      alert('Failed to preview document')
    } finally {
      setIsLoading(false)
    }
  }

  const handleDownload = async () => {
    if (!documentData) return

    try {
      switch (normalizedType) {
        case 'REQUISITION':
          await exportRequisitionPDF(documentData)
          break
        case 'PURCHASE_ORDER':
        case 'PO':
          await exportPurchaseOrderPDF(documentData)
          break
        case 'PAYMENT_VOUCHER':
        case 'PV':
          await exportPaymentVoucherPDF(documentData)
          break
        case 'GRN':
        case 'GOODS_RECEIVED_NOTE':
          await exportGrnPDF(documentData)
          break
        default:
          if (pdfBlob) {
            downloadBlob(pdfBlob, `${documentNumber}.pdf`)
          }
      }
    } catch (error) {
      console.error('Error downloading PDF:', error)
      // Fallback: download the blob directly if export fails
      if (pdfBlob) {
        downloadBlob(pdfBlob, `${documentNumber}.pdf`)
      }
    }
  }

  const handleClosePreview = (open: boolean) => {
    if (!open) {
      setIsPreviewOpen(false)
      setPdfBlob(null)
    }
  }

  if (!isPreviewSupported) {
    return null
  }

  return (
    <>
      <Button
        variant="outline"
        size="sm"
        onClick={handlePreview}
        disabled={isLoading}
        className="gap-1"
      >
        {isLoading ? (
          <Loader2 className="h-4 w-4 animate-spin" />
        ) : (
          <Eye className="h-4 w-4" />
        )}
        {isLoading ? 'Loading...' : 'Preview'}
      </Button>

      {pdfBlob && (
        <PDFPreviewDialog
          open={isPreviewOpen}
          onOpenChange={handleClosePreview}
          pdfBlob={pdfBlob}
          fileName={`${documentNumber}.pdf`}
          onDownload={handleDownload}
        />
      )}
    </>
  )
}
