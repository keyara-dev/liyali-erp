/**
 * Batch PDF export utilities
 * Handle exporting multiple documents at once
 */

import {
  getRequisitionPDFBlob,
  getPurchaseOrderPDFBlob,
  getPaymentVoucherPDFBlob,
} from './pdf-export'
import { Requisition } from '@/types/requisition'
import { PurchaseOrder } from '@/types/purchase-order'
import { PaymentVoucher } from '@/types/payment-voucher'

export interface BatchExportProgress {
  total: number
  completed: number
  current: string
  status: 'pending' | 'processing' | 'completed' | 'error'
  error?: string
}

export interface BatchExportResult {
  fileName: string
  success: boolean
  error?: string
}

/**
 * Export multiple requisitions as individual PDFs in a zip file
 */
export async function batchExportRequisitions(
  requisitions: Requisition[],
  onProgress?: (progress: BatchExportProgress) => void
): Promise<{ success: boolean; message: string; zip?: Blob }> {
  try {
    const JSZip = (await import('jszip')).default

    if (!JSZip) {
      throw new Error('JSZip library not available')
    }

    const zip = new JSZip()
    const total = requisitions.length

    for (let i = 0; i < requisitions.length; i++) {
      const requisition = requisitions[i]

      onProgress?.({
        total,
        completed: i,
        current: `REQ-${requisition.requisitionNumber}`,
        status: 'processing',
      })

      try {
        const blob = await getRequisitionPDFBlob(requisition)
        const fileName = `REQ-${requisition.requisitionNumber}.pdf`
        zip.file(fileName, blob)
      } catch (error) {
        console.error(`Error exporting requisition ${requisition.id}:`, error)
        onProgress?.({
          total,
          completed: i + 1,
          current: `REQ-${requisition.requisitionNumber}`,
          status: 'error',
          error: `Failed to export: ${requisition.requisitionNumber}`,
        })
      }
    }

    const zipBlob = await zip.generateAsync({ type: 'blob' })

    onProgress?.({
      total,
      completed: total,
      current: 'Complete',
      status: 'completed',
    })

    return {
      success: true,
      message: `Successfully exported ${total} requisitions`,
      zip: zipBlob,
    }
  } catch (error) {
    const message = error instanceof Error ? error.message : 'Unknown error'
    return {
      success: false,
      message: `Batch export failed: ${message}`,
    }
  }
}

/**
 * Export multiple purchase orders as individual PDFs in a zip file
 */
export async function batchExportPurchaseOrders(
  purchaseOrders: PurchaseOrder[],
  onProgress?: (progress: BatchExportProgress) => void
): Promise<{ success: boolean; message: string; zip?: Blob }> {
  try {
    const JSZip = (await import('jszip')).default

    if (!JSZip) {
      throw new Error('JSZip library not available')
    }

    const zip = new JSZip()
    const total = purchaseOrders.length

    for (let i = 0; i < purchaseOrders.length; i++) {
      const po = purchaseOrders[i]

      onProgress?.({
        total,
        completed: i,
        current: `PO-${po.poNumber}`,
        status: 'processing',
      })

      try {
        const blob = await getPurchaseOrderPDFBlob(po)
        const fileName = `PO-${po.poNumber}.pdf`
        zip.file(fileName, blob)
      } catch (error) {
        console.error(`Error exporting purchase order ${po.id}:`, error)
        onProgress?.({
          total,
          completed: i + 1,
          current: `PO-${po.poNumber}`,
          status: 'error',
          error: `Failed to export: ${po.poNumber}`,
        })
      }
    }

    const zipBlob = await zip.generateAsync({ type: 'blob' })

    onProgress?.({
      total,
      completed: total,
      current: 'Complete',
      status: 'completed',
    })

    return {
      success: true,
      message: `Successfully exported ${total} purchase orders`,
      zip: zipBlob,
    }
  } catch (error) {
    const message = error instanceof Error ? error.message : 'Unknown error'
    return {
      success: false,
      message: `Batch export failed: ${message}`,
    }
  }
}

/**
 * Export multiple payment vouchers as individual PDFs in a zip file
 */
export async function batchExportPaymentVouchers(
  paymentVouchers: PaymentVoucher[],
  onProgress?: (progress: BatchExportProgress) => void
): Promise<{ success: boolean; message: string; zip?: Blob }> {
  try {
    const JSZip = (await import('jszip')).default

    if (!JSZip) {
      throw new Error('JSZip library not available')
    }

    const zip = new JSZip()
    const total = paymentVouchers.length

    for (let i = 0; i < paymentVouchers.length; i++) {
      const pv = paymentVouchers[i]

      onProgress?.({
        total,
        completed: i,
        current: `PV-${pv.pvNumber}`,
        status: 'processing',
      })

      try {
        const blob = await getPaymentVoucherPDFBlob(pv)
        const fileName = `PV-${pv.pvNumber}.pdf`
        zip.file(fileName, blob)
      } catch (error) {
        console.error(`Error exporting payment voucher ${pv.id}:`, error)
        onProgress?.({
          total,
          completed: i + 1,
          current: `PV-${pv.pvNumber}`,
          status: 'error',
          error: `Failed to export: ${pv.pvNumber}`,
        })
      }
    }

    const zipBlob = await zip.generateAsync({ type: 'blob' })

    onProgress?.({
      total,
      completed: total,
      current: 'Complete',
      status: 'completed',
    })

    return {
      success: true,
      message: `Successfully exported ${total} payment vouchers`,
      zip: zipBlob,
    }
  } catch (error) {
    const message = error instanceof Error ? error.message : 'Unknown error'
    return {
      success: false,
      message: `Batch export failed: ${message}`,
    }
  }
}

/**
 * Download blob as zip file
 */
export function downloadZip(blob: Blob, fileName: string): void {
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = fileName
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
}
