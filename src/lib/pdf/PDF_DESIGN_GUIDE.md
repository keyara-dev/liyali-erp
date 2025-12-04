# PDF Export Design Guide

This document explains the PDF templates and styles for Requisitions, Purchase Orders, and Payment Vouchers.

## Overview

The PDF export system uses `@react-pdf/renderer` to generate standardized, professional PDFs for all document types. All three document templates follow a consistent design pattern while accommodating document-specific requirements.

## Architecture

### Files

1. **pdf-styles.ts** - Centralized style definitions for all PDFs
2. **requisition-pdf.tsx** - Requisition PDF component
3. **purchase-order-pdf.tsx** - Purchase Order PDF component
4. **payment-voucher-pdf.tsx** - Payment Voucher PDF component
5. **pdf-export.ts** - Export utility functions
6. **PDF_DESIGN_GUIDE.md** - This file

### Component Structure

Each PDF is built using the following sections:

```
Document
└── Page
    ├── Header (Company info + Document title)
    ├── Status Badges (Status + Priority)
    ├── Key Information Section
    ├── Details Section
    ├── Line Items Section
    ├── Financial Information (if applicable)
    ├── Approval Chain Section
    ├── Source Documents (if applicable)
    ├── Footer
    └── Page Numbers
```

## Visual Design

### Color Scheme

- **Primary**: `#1e40af` (Blue) - Headers, borders, important text
- **Backgrounds**:
  - Light gray: `#f3f4f6` - Section headers
  - Very light: `#f9fafb` - Info boxes
  - Light blue: `#dbeafe` - Status badges

- **Status Colors**:
  - Draft: Gray (`#f3f4f6`)
  - Submitted: Blue (`#dbeafe`)
  - In Review: Amber (`#fef3c7`)
  - Approved: Green (`#dcfce7`)
  - Rejected: Red (`#fee2e2`)
  - Paid: Emerald (`#d1fae5`)

- **Priority Colors**:
  - Urgent: Red (`#fee2e2`)
  - High: Amber (`#fed7aa`)
  - Medium/Low: Blue (`#dbeafe`)

### Typography

- **Font Family**: Helvetica
- **Font Sizes**:
  - Company Name: 20pt (bold)
  - Document Title: 16pt (bold)
  - Section Titles: 12pt (bold)
  - Content: 10pt (normal)
  - Labels: 9pt (bold, uppercase)
  - Footers: 8-9pt

### Layout

- **Page Margins**: 40px (all sides)
- **Section Spacing**: 20px margin-bottom
- **Column Layout**:
  - 2-column: 50% width each
  - 3-column: 33.33% width each
  - 4-column: 25% width each

## Document Sections

### 1. Header Section

```
┌─────────────────────────────────────┐
│ Liyali          REQUISITION          │
│ Finance System   REQ-2024-0001       │
└─────────────────────────────────────┘
```

- Left: Company name and subtitle
- Right: Document type and document number
- Bottom border: 2px blue

### 2. Status Section

Shows document status and priority as colored badges:

```
[DRAFT] [HIGH]
```

Provides quick visual identification of document state.

### 3. Information Boxes

Standard format for key data:

```
┌──────────────────┐
│ REQUESTED BY     │
│ John Doe         │
│ Manager          │
└──────────────────┘
```

Used for:
- Requester information
- Dates
- IDs
- Currency
- Amounts

### 4. Line Items Table

```
┌─────┬──────────────┬──────────┬─────┬──────────┬──────────┐
│  #  │ Description  │ Category │ Qty │ Unit     │ Total    │
│     │              │          │     │ Price    │          │
├─────┼──────────────┼──────────┼─────┼──────────┼──────────┤
│  1  │ Item A       │ Office   │ 5   │ 100.00   │ 500.00   │
├─────┼──────────────┼──────────┼─────┼──────────┼──────────┤
│  2  │ Item B       │ Office   │ 3   │ 200.00   │ 600.00   │
└─────┴──────────────┴──────────┴─────┴──────────┴──────────┘
```

Standard table with:
- Header row: Blue background, bold text
- Data rows: Alternating white/gray for readability
- Borders: 1px gray

### 5. Totals Section

```
                        TOTAL AMOUNT:
                        ZMW 1,100.00
```

- Right-aligned
- Large, bold text
- Blue color
- Top border for visual separation

### 6. Approval Chain Section

Shows approval stages and status:

```
Stage 1: Finance Manager Review
john.doe@company.com
Status: APPROVED - Approved on 2024-01-15

Stage 2: Director Approval
jane.smith@company.com
Status: PENDING
```

Each stage shows:
- Stage number and name
- Assigned person
- Current status
- Action date (if actioned)

### 7. Source Documents

For linked documents:

```
SOURCE REQUISITION
REQ-2024-0001
```

Highlighted box with left border for quick identification.

### 8. Footer

```
Generated on 1/15/2024 at 2:30:45 PM
This is a system-generated document. No signature is required.
```

- Centered text
- Light gray color
- Top border separator

## Document-Specific Content

### Requisition PDF

**Unique Sections**:
- Request Information (requester, department, dates, priority)
- Requisition Details (title, description, budget code)
- Line Items with quantities and unit prices
- Approval Chain (3 stages)

**Key Fields**:
- requisitionNumber
- status
- priority
- department
- items (with itemNumber, description, category, quantity, unitPrice)
- approvalChain (with stageName, assignedTo, status, actionTakenAt)

### Purchase Order PDF

**Unique Sections**:
- Vendor & Order Information (vendor, department, requester, dates)
- Order Details (title, description)
- Line Items
- Financial Information (budget code, cost center, project code)
- Approval Chain (4 stages)
- Source Requisition (if linked)

**Key Fields**:
- poNumber
- status
- priority
- vendorName
- requiredByDate
- items (PO items with descriptions and pricing)
- approvalChain (4 stages for PO approval)
- sourceRequisitionNumber

### Payment Voucher PDF

**Unique Sections**:
- Vendor & Payment Information (vendor, payment method, due date, bank details)
- Voucher Details (title, description)
- Line Items (mapped from PO)
- Financial Information (budget, cost center, taxes)
- Payment Confirmation (if PAID status)
- Approval Chain (3 stages)
- Source Documents (PO and Requisition links)

**Key Fields**:
- pvNumber
- status
- priority
- paymentMethod
- bankDetails (bankName, accountName, accountNumber)
- items (with descriptions and pricing)
- approvalChain (3 stages for PV approval)
- sourcePurchaseOrderNumber
- sourceRequisitionNumber
- paidAmount, paidDate, referenceNumber

## Usage

### Export Functions

```typescript
// Export and download immediately
await exportRequisitionPDF(requisition)
await exportPurchaseOrderPDF(purchaseOrder)
await exportPaymentVoucherPDF(paymentVoucher)

// Get blob for custom handling
const blob = await getRequisitionPDFBlob(requisition)

// Get object URL for preview/preview in iframe
const url = await getRequisitionPDFUrl(requisition)
```

### Implementation in Components

```typescript
import { exportPurchaseOrderPDF } from '@/lib/pdf/pdf-export'

export function PODetailClient({ purchaseOrder }) {
  const handleDownloadPDF = async () => {
    try {
      await exportPurchaseOrderPDF(purchaseOrder)
    } catch (error) {
      console.error('PDF export failed:', error)
    }
  }

  return (
    <Button onClick={handleDownloadPDF} className="gap-2">
      <Download className="h-4 w-4" />
      Download PDF
    </Button>
  )
}
```

## Customization

### Modifying Styles

Edit `pdf-styles.ts` to change:
- Color scheme
- Font sizes
- Spacing
- Borders
- Backgrounds

### Modifying Templates

Each PDF template (requisition-pdf.tsx, purchase-order-pdf.tsx, payment-voucher-pdf.tsx) can be customized:
- Add/remove sections
- Reorder sections
- Change content layout
- Adjust spacing

### Adding Sections

To add a new section:

```typescript
<View style={pdfStyles.section}>
  <Text style={pdfStyles.sectionTitle}>NEW SECTION</Text>
  {/* Content here */}
</View>
```

### Conditional Content

Sections can be conditionally rendered:

```typescript
{paymentVoucher.status === 'PAID' && (
  <View style={pdfStyles.highlighted}>
    <Text>Payment Confirmation Section</Text>
  </View>
)}
```

## Reference Template

The provided PO template PDF shows:
- Professional header with company branding space
- Status and priority indicators
- Organized information in boxed sections
- Clear line items table
- Prominent totals
- Approval signatures area
- Footer with generation info

Our implementation follows this pattern while being fully programmatic and data-driven.

## Best Practices

1. **Keep sections consistent** across all three document types
2. **Use meaningful colors** for status and priority
3. **Ensure readability** with adequate spacing and font sizes
4. **Include all critical information** in the PDF output
5. **Test exports** with various document states (DRAFT, APPROVED, etc.)
6. **Handle errors gracefully** in export functions
7. **Use file names** that include document number and timestamp

## Future Enhancements

1. **Multi-page documents** for large line item lists
2. **Page numbers** (Page X of Y)
3. **Company logo** and header image
4. **Signature capture** for approved documents
5. **Watermarks** for draft vs. final versions
6. **Localization** for different date/currency formats
7. **QR codes** for document verification
8. **Email integration** to send PDFs directly
9. **Print optimization** for better printing results
10. **Template selection** for different department requirements
