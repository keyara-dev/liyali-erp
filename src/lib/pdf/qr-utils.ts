/**
 * QR Code utility functions for PDF documents
 * Generates QR codes that encode document information for tracking
 */

/**
 * Generate QR code data for a document
 * This creates a URL or string that encodes document information
 */
export function generateDocumentQRData(
  documentType: 'REQUISITION' | 'PURCHASE_ORDER' | 'PAYMENT_VOUCHER',
  documentNumber: string,
  documentId: string,
  timestamp: Date
): string {
  // Format: TYPE|NUMBER|ID|TIMESTAMP|CHECKSUM
  const data = `${documentType}|${documentNumber}|${documentId}|${timestamp.toISOString()}`
  return data
}

/**
 * Generate a unique tracking code for a document
 * Format: TYPE-DOCNUM-HASH-TIMESTAMP
 */
export function generateTrackingCode(
  documentType: string,
  documentNumber: string
): string {
  const typePrefix = documentType === 'REQUISITION' ? 'REQ' :
                     documentType === 'PURCHASE_ORDER' ? 'PO' :
                     documentType === 'PAYMENT_VOUCHER' ? 'PV' : 'DOC'

  // Create a simple hash from document number
  const hash = Math.abs(documentNumber.split('').reduce((acc, char) => {
    return ((acc << 5) - acc) + char.charCodeAt(0)
  }, 0)).toString(16).substring(0, 6).toUpperCase()

  const timestamp = new Date().getTime().toString(36).toUpperCase()

  return `${typePrefix}-${hash}-${timestamp}`
}

/**
 * Generate QR code URL using a free QR service
 * We use qr-server.com which doesn't require API key
 */
export function getQRCodeUrl(data: string, size: number = 200): string {
  // Encode the data for URL
  const encodedData = encodeURIComponent(data)
  // Using QR Server API - free, no authentication needed
  return `https://api.qrserver.com/v1/create-qr-code/?size=${size}x${size}&data=${encodedData}`
}

/**
 * Create a local QR code as data URL
 * For offline/embedded QR codes in PDFs
 */
export async function generateQRCodeDataUrl(data: string, size: number = 200): Promise<string> {
  try {
    const QRCode = require('qrcode')
    // Generate QR code as data URL
    const dataUrl = await QRCode.toDataURL(data, {
      errorCorrectionLevel: 'H',
      type: 'image/png',
      quality: 0.95,
      margin: 1,
      width: size,
    })
    return dataUrl
  } catch (error) {
    console.error('Error generating QR code:', error)
    // Fallback to online QR service
    return getQRCodeUrl(data, size)
  }
}

/**
 * Format tracking information for display
 */
export function formatTrackingInfo(
  documentNumber: string,
  documentId: string,
  status: string,
  createdDate: Date
): string {
  return [
    `Document: ${documentNumber}`,
    `ID: ${documentId}`,
    `Status: ${status}`,
    `Created: ${createdDate.toLocaleDateString()}`,
  ].join(' | ')
}
