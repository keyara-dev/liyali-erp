# PDF Export Enhancements

This document outlines all the optional enhancements implemented for the Liyali PDF export system.

## 1. Inline PDF Preview Functionality ✅

### Overview
Users can now preview PDFs before downloading them using an interactive dialog modal.

### Features
- **Preview Dialog**: Full-screen modal showing PDF pages
- **Page Navigation**: Navigate through multi-page documents
- **Download from Preview**: Download directly from the preview dialog
- **Responsive Design**: Works on all screen sizes

### Files Modified
- `requisition-detail-client.tsx`: Added Preview button and dialog integration
- `po-detail-client.tsx`: Added Preview button and dialog integration
- `pv-detail-client.tsx`: Added Preview button and dialog integration

### New Component
- `src/components/pdf-preview-dialog.tsx`: Reusable PDF preview component using react-pdf

### Dependencies
- `react-pdf@10.2.0`: PDF viewing library
- `pdfjs-dist@5.4.449`: PDF.js worker for rendering

### Usage Example
```typescript
import { PDFPreviewDialog } from '@/components/pdf-preview-dialog'

const [previewOpen, setPreviewOpen] = useState(false)
const [previewBlob, setPreviewBlob] = useState<Blob | null>(null)

const handlePreview = async () => {
  const blob = await getRequisitionPDFBlob(requisition)
  setPreviewBlob(blob)
  setPreviewOpen(true)
}

return (
  <>
    <Button onClick={handlePreview}>Preview PDF</Button>
    {previewBlob && (
      <PDFPreviewDialog
        open={previewOpen}
        onOpenChange={setPreviewOpen}
        pdfBlob={previewBlob}
        fileName="REQ-123456.pdf"
        onDownload={handleExportPDF}
      />
    )}
  </>
)
```

---

## 2. Email PDF as Attachment Feature ✅

### Overview
Send documents via email with the PDF attached directly.

### Features
- **Multiple Recipients**: Send to multiple email addresses
- **CC/BCC Support**: Include cc and bcc recipients
- **Custom Email Body**: Template-based email generation
- **Email Validation**: Validate email addresses before sending
- **Base64 Encoding**: Convert PDFs for API transmission

### Files Created
- `src/lib/pdf/pdf-email.ts`: Email service and utilities

### Available Functions

#### `sendRequisitionPDFEmail(requisition, options)`
Send a requisition as PDF via email
```typescript
const result = await sendRequisitionPDFEmail(requisition, {
  subject: 'Purchase Requisition for Your Review',
  body: 'Please review the attached requisition...',
  recipients: [
    { email: 'manager@company.com', name: 'John Manager' }
  ],
  cc: [
    { email: 'finance@company.com', name: 'Finance Team' }
  ]
})
```

#### `sendPurchaseOrderPDFEmail(purchaseOrder, options)`
Send a purchase order as PDF via email

#### `sendPaymentVoucherPDFEmail(paymentVoucher, options)`
Send a payment voucher as PDF via email

#### `buildDocumentEmailBody(documentType, documentNumber, recipientName, message)`
Generate professional email body template
```typescript
const emailBody = buildDocumentEmailBody(
  'PURCHASE_ORDER',
  'PO-2024-001',
  'John Smith',
  'Please approve this purchase order for budget review.'
)
```

### Helper Functions
- `formatRecipientsDisplay(recipients)`: Format recipient list for display
- `isValidEmail(email)`: Validate email address format

### API Integration
The system calls `/api/email/send-with-attachment` endpoint with:
```json
{
  "recipients": [{ "email": "...", "name": "..." }],
  "cc": [...],
  "bcc": [...],
  "subject": "...",
  "body": "...",
  "attachment": {
    "filename": "REQ-123.pdf",
    "content": "base64-encoded-pdf",
    "contentType": "application/pdf"
  }
}
```

---

## 3. PDF Signature Verification using QR Codes ✅

### Overview
Verify PDF authenticity and integrity using embedded QR codes.

### Features
- **QR Data Decoding**: Parse encoded QR code data
- **Checksum Validation**: Verify document hasn't been tampered with
- **Authenticity Check**: Validate document number and ID match
- **Timestamp Verification**: Check document age
- **Document Comparison**: Compare two QR data objects

### Files Created
- `src/lib/pdf/qr-verification.ts`: QR verification utilities

### QR Code Data Format
```
REQUISITION|REQ-2024-001|uuid-123|2025-12-04T15:30:00Z
```

### Available Functions

#### `decodeQRData(qrString)`
Decode QR code string and return structured data
```typescript
const qrData = decodeQRData(qrCodeString)
if (qrData) {
  console.log(qrData.documentType)     // 'REQUISITION'
  console.log(qrData.documentNumber)   // 'REQ-2024-001'
  console.log(qrData.documentId)       // 'uuid-123'
  console.log(qrData.timestamp)        // Date object
}
```

#### `validateDocumentAuthenticity(qrData, expectedNumber, expectedId)`
Validate document authenticity against expected values
```typescript
const validation = validateDocumentAuthenticity(
  qrData,
  'REQ-2024-001',
  'document-uuid'
)

if (validation.isAuthentic) {
  console.log('Document is authentic!')
} else {
  console.log('Issues found:', validation.issues)
}
```

#### `verifyQRChecksum(documentType, documentNumber, checksum)`
Verify checksum for document integrity

#### `formatQRData(qrData)`
Format QR data for display in UI
```typescript
const formatted = formatQRData(qrData)
// Document Type: REQUISITION
// Document Number: REQ-2024-001
// Document ID: uuid-123
// Created: 12/4/2025, 3:30:00 PM
```

### Security Considerations
- Checksums use simple hash-based validation
- Timestamps must be recent (within 24 hours)
- Document number and ID must match expectations
- QR data is tamper-evident (checksum mismatch detected)

---

## 4. Batch Export Functionality ✅

### Overview
Export multiple documents at once as a ZIP archive with progress tracking.

### Features
- **ZIP Export**: Combine multiple PDFs into one ZIP file
- **Progress Tracking**: Real-time progress updates
- **Error Handling**: Gracefully handle individual document failures
- **Maintains Structure**: Organized file naming in ZIP
- **Memory Efficient**: Streams files to ZIP rather than loading all in memory

### Files Created
- `src/lib/pdf/pdf-batch-export.ts`: Batch export utilities

### Dependencies
- `jszip@3.10.1`: ZIP file creation library

### Available Functions

#### `batchExportRequisitions(requisitions, onProgress)`
Export multiple requisitions as ZIP
```typescript
const result = await batchExportRequisitions(
  [requisition1, requisition2, requisition3],
  (progress) => {
    console.log(`Processing ${progress.current}: ${progress.completed}/${progress.total}`)
  }
)

if (result.success && result.zip) {
  downloadZip(result.zip, 'requisitions-export.zip')
}
```

#### `batchExportPurchaseOrders(purchaseOrders, onProgress)`
Export multiple purchase orders as ZIP

#### `batchExportPaymentVouchers(paymentVouchers, onProgress)`
Export multiple payment vouchers as ZIP

#### `downloadZip(blob, fileName)`
Download ZIP file to user's computer
```typescript
downloadZip(zipBlob, 'documents-export.zip')
```

### Progress Callback Structure
```typescript
interface BatchExportProgress {
  total: number              // Total documents to export
  completed: number          // Documents processed so far
  current: string            // Current document being processed
  status: 'pending' | 'processing' | 'completed' | 'error'
  error?: string             // Error message if failed
}
```

### Batch Export Flow
1. User selects multiple documents
2. Initiates batch export
3. System creates ZIP container
4. For each document:
   - Generates PDF
   - Adds to ZIP
   - Updates progress callback
5. Returns ZIP blob for download
6. User downloads complete archive

### Error Handling
- Individual document failures don't stop the batch
- Failed documents are skipped with error logged
- Overall batch result includes partial success info
- Progress callback updates on errors

---

## 5. Watermark Support ✅

### Overview
Automatically add status-based watermarks to PDFs (DRAFT, APPROVED, PAID, etc.).

### Features
- **Status-Based Watermarks**: Different watermarks for different statuses
- **Customizable Appearance**: Font size, color, opacity, angle
- **SVG Watermarks**: Vector-based for crisp appearance
- **Canvas Watermarks**: Fallback for client-side rendering
- **Color Coding**: Visual indication of document status

### Files Created
- `src/lib/pdf/pdf-watermark.ts`: Watermark utilities and configuration

### Watermark Status Map
| Status | Text | Color | Opacity |
|--------|------|-------|---------|
| DRAFT | DRAFT | Red (#FF6B6B) | 15% |
| SUBMITTED | SUBMITTED | Orange (#FFA500) | 12% |
| IN_REVIEW | IN REVIEW | Yellow (#FFD93D) | 12% |
| APPROVED | APPROVED | Green (#6BCB77) | 10% |
| PAID | PAID | Blue (#4D96FF) | 10% |
| REJECTED | REJECTED | Pink (#FF006E) | 15% |

### Available Functions

#### `getWatermarkByStatus(status)`
Get watermark configuration for a status
```typescript
const watermark = getWatermarkByStatus('DRAFT')
// {
//   text: 'DRAFT',
//   opacity: 0.15,
//   fontSize: 72,
//   color: '#FF6B6B',
//   angle: -45
// }
```

#### `createWatermarkSVG(options)`
Create watermark as SVG data URL
```typescript
const svg = createWatermarkSVG({
  text: 'CONFIDENTIAL',
  opacity: 0.2,
  fontSize: 60,
  color: '#FF0000'
})
// Returns: 'data:image/svg+xml,...'
```

#### `createWatermarkCanvas(options, width, height)`
Create watermark as HTML canvas
```typescript
const canvas = createWatermarkCanvas(
  { text: 'DRAFT', opacity: 0.15 },
  800,
  600
)
document.body.appendChild(canvas)
```

#### `getWatermarkStyle(status)`
Get CSS style object for react-pdf components
```typescript
const style = getWatermarkStyle('DRAFT')
// Can be applied to PDF components
```

#### `hasWatermark(status)`
Check if status has a watermark
```typescript
if (hasWatermark('APPROVED')) {
  // Apply watermark
}
```

### Integration Example
```typescript
import { getWatermarkByStatus } from '@/lib/pdf/pdf-watermark'

// In your PDF template component
const watermark = getWatermarkByStatus(document.status)

{watermark && (
  <View style={getWatermarkSVGStyle(watermark)}>
    <Text>{watermark.text}</Text>
  </View>
)}
```

### Watermark Customization
```typescript
interface WatermarkOptions {
  text: string              // Watermark text
  opacity?: number          // 0-1 (default 0.15)
  fontSize?: number         // In pixels (default 72)
  fontFamily?: string       // Font name (default 'Arial')
  color?: string            // Hex color (default '#CCCCCC')
  angle?: number            // Rotation in degrees (default -45)
}
```

---

## Integration Guide

### Adding Preview to a New Detail Page
1. Import preview dialog and PDF blob getter
2. Add state for `previewOpen` and `previewBlob`
3. Create `handlePreviewPDF` function
4. Add Preview button to UI
5. Add `<PDFPreviewDialog>` at end of render

### Adding Email Export
1. Import email functions
2. Create email options object
3. Call appropriate `sendXXXPDFEmail` function
4. Handle success/error responses

### Adding Batch Export
1. Import batch export function and download helper
2. Collect documents to export
3. Call batch export with progress callback
4. Download ZIP when complete

### Adding Watermarks to PDFs
1. Import watermark utilities in PDF template
2. Get watermark options by status
3. Add watermark view/text to PDF template
4. Customize as needed

---

## Performance Considerations

### PDF Generation
- PDFs are generated on-demand to reduce memory usage
- For batch exports, PDFs are streamed to ZIP container
- Preview uses cached blobs to avoid regeneration

### Memory Management
- Blob URLs are revoked after download
- Canvas watermarks are cleaned up after use
- Batch operations process documents sequentially

### File Size
- Typical single-page PDF: 50-150 KB
- ZIP compression reduces size by ~20-30%
- Batch export of 100 documents: ~5-10 MB

---

## Error Handling

### Common Issues

#### QR Code Not Scanned
- Ensure QR code is visible in PDF
- Check QR code size (minimum 80x80 pixels recommended)
- Verify QR data is correctly encoded

#### Email Send Failure
- Verify email addresses are valid
- Check API endpoint is accessible
- Ensure SMTP credentials are configured

#### Batch Export Incomplete
- Check individual PDFs generate without error
- Verify ZIP library is loaded
- Ensure sufficient disk space

#### Watermark Not Visible
- Check opacity is not 0
- Verify color contrast with background
- Ensure font size is appropriate

---

## Future Enhancements

Potential improvements for future versions:
1. **Digital Signatures**: Add cryptographic signatures to PDFs
2. **Cloud Storage**: Save PDFs to cloud (S3, Google Drive, etc.)
3. **Advanced Watermarking**: Custom watermark images/logos
4. **PDF Compression**: Optimize file size for large batches
5. **Email Templates**: Pre-built email templates for different document types
6. **Bulk Operations**: Approve/reject multiple documents
7. **Archive Management**: Automatic PDF archival with retention policies
8. **Accessibility**: Enhanced PDF accessibility features

---

## Testing Recommendations

### Manual Testing
- [ ] Preview PDF in all supported browsers
- [ ] Test batch export with 1, 10, 100 documents
- [ ] Verify QR codes scan correctly
- [ ] Test email sending with various recipients
- [ ] Check watermarks render correctly in different statuses
- [ ] Test on mobile devices
- [ ] Verify file downloads correctly

### Automated Testing
- [ ] Unit tests for PDF generation
- [ ] Integration tests for email service
- [ ] QR code validation tests
- [ ] Batch export progress tests
- [ ] Watermark style generation tests

---

## Dependencies Summary

| Package | Version | Purpose |
|---------|---------|---------|
| @react-pdf/renderer | 4.3.1 | PDF generation |
| react-pdf | 10.2.0 | PDF viewing/preview |
| pdfjs-dist | 5.4.449 | PDF rendering engine |
| qrcode | 1.5.4 | QR code generation |
| jszip | 3.10.1 | ZIP file creation |

---

## Support & Documentation

For more information:
- See `PDF_DESIGN_GUIDE.md` for template design patterns
- Check `qr-utils.ts` for QR code utility functions
- Review `pdf-export.ts` for core export functions

Last Updated: December 4, 2025
