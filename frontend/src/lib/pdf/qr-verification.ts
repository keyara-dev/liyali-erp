/**
 * QR Code verification utilities
 * Decodes and verifies QR code data from PDFs
 */

export interface QRCodeData {
  documentType: 'REQUISITION' | 'PURCHASE_ORDER' | 'PAYMENT_VOUCHER'
  documentNumber: string
  documentId: string
  timestamp: Date
  isValid: boolean
}

/**
 * Decode QR code data from string
 * Format: TYPE|NUMBER|ID|TIMESTAMP|CHECKSUM
 */
export function decodeQRData(qrString: string): QRCodeData | null {
  try {
    const parts = qrString.split('|')
    if (parts.length < 4) {
      return null
    }

    const [documentType, documentNumber, documentId, timestampStr] = parts

    // Validate document type
    const validTypes = ['REQUISITION', 'PURCHASE_ORDER', 'PAYMENT_VOUCHER']
    if (!validTypes.includes(documentType)) {
      return null
    }

    // Parse timestamp
    const timestamp = new Date(timestampStr)
    if (isNaN(timestamp.getTime())) {
      return null
    }

    return {
      documentType: documentType as 'REQUISITION' | 'PURCHASE_ORDER' | 'PAYMENT_VOUCHER',
      documentNumber,
      documentId,
      timestamp,
      isValid: true,
    }
  } catch (error) {
    console.error('Error decoding QR data:', error)
    return null
  }
}

/**
 * Format QR data for display
 */
export function formatQRData(qrData: QRCodeData): string {
  const typeLabel = qrData.documentType.replace(/_/g, ' ')
  return `
Document Type: ${typeLabel}
Document Number: ${qrData.documentNumber}
Document ID: ${qrData.documentId}
Created: ${qrData.timestamp.toLocaleString()}
  `.trim()
}

/**
 * Verify QR code checksum (simple hash-based validation)
 */
export function verifyQRChecksum(
  documentType: string,
  documentNumber: string,
  checksum: string
): boolean {
  try {
    const data = `${documentType}|${documentNumber}`
    const calculated = simpleHash(data)
    return calculated === checksum
  } catch (error) {
    console.error('Error verifying checksum:', error)
    return false
  }
}

/**
 * Generate QR checksum using simple hash
 */
export function generateQRChecksum(
  documentType: string,
  documentNumber: string
): string {
  const data = `${documentType}|${documentNumber}`
  return simpleHash(data)
}

/**
 * Simple hash function for checksums
 */
function simpleHash(str: string): string {
  let hash = 0
  for (let i = 0; i < str.length; i++) {
    const char = str.charCodeAt(i)
    hash = (hash << 5) - hash + char
    hash = hash & hash // Convert to 32bit integer
  }
  return Math.abs(hash).toString(16).toUpperCase()
}

/**
 * Validate document authenticity using QR data
 */
export function validateDocumentAuthenticity(
  qrData: QRCodeData,
  expectedDocumentNumber: string,
  expectedDocumentId: string
): {
  isAuthentic: boolean
  issues: string[]
} {
  const issues: string[] = []

  // Check document number matches
  if (qrData.documentNumber !== expectedDocumentNumber) {
    issues.push('Document number mismatch')
  }

  // Check document ID matches
  if (qrData.documentId !== expectedDocumentId) {
    issues.push('Document ID mismatch')
  }

  // Check if document is recent (within 24 hours)
  const now = new Date()
  const hoursDiff =
    (now.getTime() - qrData.timestamp.getTime()) / (1000 * 60 * 60)
  if (hoursDiff > 24) {
    issues.push('Document is older than 24 hours')
  }

  return {
    isAuthentic: issues.length === 0,
    issues,
  }
}

/**
 * Compare two QR data objects for equality
 */
export function compareQRData(qr1: QRCodeData, qr2: QRCodeData): boolean {
  return (
    qr1.documentType === qr2.documentType &&
    qr1.documentNumber === qr2.documentNumber &&
    qr1.documentId === qr2.documentId &&
    qr1.timestamp.getTime() === qr2.timestamp.getTime()
  )
}
