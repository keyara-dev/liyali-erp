'use client'

import { WorkflowDocument } from '@/types/workflow'

interface GRNPDFProps {
  grn: WorkflowDocument
}

/**
 * GoodsReceivedNotePDF Component (Placeholder)
 * When @react-pdf/renderer is installed, this can be replaced with proper PDF rendering
 */
export function GoodsReceivedNotePDF({ grn }: GRNPDFProps) {
  return null
}

/**
 * Generate and download the PDF for a GRN
 * Currently a placeholder - install @react-pdf/renderer to enable PDF generation
 */
export async function generateGrnPDF(grn: WorkflowDocument) {
  try {
    const filename = `GRN-${grn.documentNumber}-${new Date().getTime()}.pdf`
    console.log(`PDF generation requested for: ${filename}`)
    console.log('To enable PDF download, install @react-pdf/renderer package')
    // TODO: Implement actual PDF generation with @react-pdf/renderer
  } catch (error) {
    console.error('Error generating GRN PDF:', error)
    throw error
  }
}
