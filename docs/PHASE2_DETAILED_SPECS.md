# Phase 2: Detailed Specifications - Purchase Order, GRN & Payment Voucher

**Status**: Ready for Implementation
**Estimated Effort**: 25-30 hours
**Priority**: HIGH - Critical path for workflow completion

---

## Overview

Phase 2 completes the requisition-to-payment workflow by implementing three critical document types and their workflows:

1. **Purchase Order (PO)** - Created from approved requisition
2. **Goods Received Note (GRN)** - Tracks goods receipt and inspection
3. **Payment Voucher (PV)** - Final approval and payment authorization

---

## Part A: Purchase Order Workflow

### Business Process Flow

```
Requisition Final Approval
  ↓
Auto-Create Purchase Order (linked to requisition)
  ↓
PO Status: SUBMITTED → IN_APPROVAL
  ↓
Principal Officer Reviews & Approves PO
  ↓
PO Status: APPROVED
  ↓
Notify Stores Officer: Goods can be received
  ↓
Trigger GRN Workflow
```

### Data Model

```typescript
interface PurchaseOrder extends WorkflowDocument {
  type: 'PURCHASE_ORDER'

  // Links
  requisitionId: string  // Reference to requisition that created this

  // PO-specific fields
  poNumber: string       // e.g., "PO-2024-001234"
  vendorId: string
  vendorName: string
  vendorContact: string
  vendorEmail?: string

  // Items (copied from requisition)
  items: Array<{
    id: string
    description: string
    quantity: number
    unitCost: number
    totalCost: number
    specifications?: string
  }>

  // Delivery
  deliveryType: 'Standard' | 'Express' | 'Pick-up'
  deliveryAddress: string
  deliveryDate?: Date

  // Cost
  subtotal: number
  tax?: number
  totalAmount: number
  currency: string

  // Terms
  paymentTerms?: string
  specialInstructions?: string

  // Status tracking
  status: 'DRAFT' | 'SUBMITTED' | 'IN_APPROVAL' | 'APPROVED' | 'REJECTED'
  currentStage: 1  // Only 1 stage for PO

  // Approval
  approvedBy?: User
  approvedAt?: Date

  // Additional
  metadata: {
    requisitionNumber: string
    createdFrom: 'REQUISITION'
    autoCreated: true
  }
}
```

### Implementation Tasks

#### A.1 Create PO List Page
**Files**:
- `src/app/workflows/purchase-orders/page.tsx` (server component)
- `src/app/workflows/purchase-orders/_components/po-client.tsx` (client orchestrator)
- `src/app/workflows/purchase-orders/_components/po-table.tsx` (data table)

**Features**:
```typescript
// po-table.tsx columns
const columns: ColumnDef<PurchaseOrder>[] = [
  {
    accessorKey: 'poNumber',
    header: ({ column }) => (
      <Button variant="ghost" onClick={() => column.toggleSorting(...)}>
        PO Number <ArrowUpDown />
      </Button>
    ),
  },
  {
    accessorKey: 'vendorName',
    header: 'Vendor',
  },
  {
    accessorKey: 'totalAmount',
    header: ({ column }) => (
      <Button variant="ghost" onClick={() => column.toggleSorting(...)}>
        Amount <ArrowUpDown />
      </Button>
    ),
    cell: ({ row }) => formatCurrency(row.original.totalAmount),
  },
  {
    accessorKey: 'status',
    header: 'Status',
    cell: ({ row }) => <Badge variant={statusVariant}>{row.original.status}</Badge>,
  },
  {
    accessorKey: 'requisitionNumber',
    header: 'From Requisition',
    cell: ({ row }) => (
      <Link href={`/workflows/requisitions/${row.original.requisitionId}`}>
        {row.original.metadata.requisitionNumber}
      </Link>
    ),
  },
  {
    id: 'actions',
    cell: ({ row }) => (
      <Button
        variant="ghost"
        onClick={() => router.push(`/workflows/purchase-orders/${row.original.id}`)}
      >
        View Details
      </Button>
    ),
  },
]
```

**Functionality**:
- [ ] Display all POs created from requisitions
- [ ] Sort by PO number, amount, vendor, date
- [ ] Filter by vendor, status
- [ ] Show link to originating requisition
- [ ] 10 items per page with pagination
- [ ] Loading and empty states

#### A.2 Create PO Detail Page
**Files**:
- `src/app/workflows/purchase-orders/[id]/page.tsx`
- `src/app/workflows/purchase-orders/_components/po-detail-client.tsx`

**Layout**:
```
Header Section
├── Back button
├── PO Number + Status Badge
├── Vendor info
└── Approval Status (1-stage: Principal Officer)

Main Content (3-column)
├── Column 1: PO Details
│   ├── Vendor Information
│   ├── Items List
│   ├── Cost Summary
│   └── Delivery Info
├── Column 2: Requisition Link
│   ├── Original Requisition Number (clickable)
│   ├── Requisition Status
│   └── Department/Requester Info
└── Sidebar: Approval Section
    ├── Current Status (SUBMITTED/IN_APPROVAL/APPROVED/REJECTED)
    ├── Approval History
    └── Approval Action Panel (if approver)
```

**Approval Action Panel** (for Principal Officer only):
- [ ] Show "This PO needs your approval" if IN_APPROVAL
- [ ] Display requisition it was created from
- [ ] Approve button
- [ ] Reject button (with reason)
- [ ] Optional comments field
- [ ] Show audit trail of how requisition was approved

#### A.3 Update Requisition Detail Page
**File**: `src/app/workflows/requisitions/_components/requisition-detail-client.tsx`

**Add to detail page** (after final approval):
```typescript
// Show connected PO info (if requisition is APPROVED)
if (requisition.status === 'APPROVED') {
  return (
    <Alert>
      <AlertDescription>
        Purchase Order Created: {poNumber}
        <Button variant="link">View PO</Button>
      </AlertDescription>
    </Alert>
  )
}
```

#### A.4 Server Actions for PO
**File**: `src/app/_actions/workflow.ts`

**Add/Update functions**:
```typescript
// Called automatically when requisition approved at stage 4
async function autoCreatePurchaseOrder(requisitionId: string) {
  const requisition = documentStore.get(requisitionId)

  const poData = {
    poNumber: generatePONumber(),
    requisitionId,
    vendorName: requisition.metadata.supplierInfo?.name,
    vendorContact: requisition.metadata.supplierInfo?.contact,
    items: requisition.metadata.items,
    deliveryType: requisition.metadata.deliveryType,
    deliveryAddress: requisition.metadata.deliveryAddress,
    totalAmount: requisition.totalAmount,
    metadata: {
      requisitionNumber: requisition.documentNumber,
      createdFrom: 'REQUISITION',
      autoCreated: true
    }
  }

  return createWorkflowDocument('PURCHASE_ORDER', poData)
}

// PO approval (only Principal Officer, single stage)
async function approvePurchaseOrder(poId: string, comments?: string) {
  const po = documentStore.get(poId)

  // Update status
  po.status = 'APPROVED'
  po.approvedAt = new Date()

  // Log approval
  approvalLogsStore.get(poId)?.push({
    id: generateId(),
    approver: currentUser,
    action: 'APPROVED',
    timestamp: new Date(),
    comments
  })

  // Notify stores officer to receive goods
  // This triggers GRN workflow

  return successResponse(po)
}

// PO rejection
async function rejectPurchaseOrder(poId: string, reason: string) {
  const po = documentStore.get(poId)
  po.status = 'REJECTED'

  // Log rejection
  // Re-send to requisition creator with reason

  return successResponse(po)
}

// Get PO by requisition ID
async function getPOByRequisitionId(requisitionId: string) {
  return Array.from(documentStore.values()).find(
    doc => doc.type === 'PURCHASE_ORDER' &&
            doc.metadata.requisitionId === requisitionId
  )
}

// Get all POs
async function getPurchaseOrders(page: number, pageSize: number) {
  const pos = Array.from(documentStore.values())
    .filter(doc => doc.type === 'PURCHASE_ORDER')
    .sort((a, b) => b.createdAt.getTime() - a.createdAt.getTime())

  return paginatedResponse(pos, page, pageSize)
}
```

---

## Part B: Goods Received Note (GRN) Workflow

### Business Process Flow

```
PO Approved
  ↓
Stores Officer Notified: Goods can be received
  ↓
Goods Arrived at Warehouse
  ↓
Stores Officer Creates GRN
  ├── References PO
  ├── Confirms items received
  ├── Adds delivery inspection notes
  └── Signs off
  ↓
GRN Status: COMPLETED
  ↓
Trigger Payment Voucher Creation
```

### Data Model

```typescript
interface GoodsReceivedNote {
  id: string
  grNumber: string         // e.g., "GRN-2024-005678"

  // Links
  poId: string             // Reference to Purchase Order
  requisitionId: string    // Reference to original requisition

  // Receipt Details
  receivedDate: Date
  receivedBy: User         // Stores Officer
  receivedFrom: string     // Vendor name (copied from PO)

  // Items Received
  items: Array<{
    poItemId: string       // Links to PO item
    description: string
    quantityOrdered: number
    quantityReceived: number
    condition: 'Good' | 'Damaged' | 'Partial'
    notes?: string
  }>

  // Inspection
  inspectionNotes: string
  damageReport?: string
  discrepancies?: Array<{
    item: string
    issue: string
    resolved: boolean
  }>

  // Sign-off
  signedBy: User
  signedAt: Date

  // Status
  status: 'DRAFT' | 'COMPLETED' | 'REJECTED'

  // Attachments
  attachments: Attachment[]

  // For voucher generation
  readyForPayment: boolean
  paymentVoucherId?: string
}
```

### Implementation Tasks

#### B.1 Create GRN Form Page
**Files**:
- `src/app/workflows/grn/page.tsx`
- `src/app/workflows/grn/_components/grn-client.tsx`
- `src/app/workflows/grn/_components/grn-form.tsx`
- `src/app/workflows/grn/_components/grn-list.tsx`

**GRN Form** (src/app/workflows/grn/_components/grn-form.tsx):
```typescript
interface GRNFormProps {
  poId: string
  onSubmit: (grn: GoodsReceivedNote) => void
}

// Form sections:
// 1. Select PO (if not pre-filled)
// 2. Confirm Items Received
//    ├── For each item in PO
//    │   ├── Qty Received (vs Qty Ordered)
//    │   ├── Condition (Good/Damaged/Partial)
//    │   └── Item Notes
//    └── Add Discrepancies if any qty mismatch
// 3. Inspection Notes (textarea)
// 4. Damage Report (optional)
// 5. Upload Documents (delivery slip, inspection docs)
// 6. Signature (Stores Officer - current user)
// 7. Submit button
```

**Features**:
- [ ] Auto-load PO details
- [ ] Show items from PO
- [ ] QR code scan for PO (optional future)
- [ ] Item-by-item receipt confirmation
- [ ] Condition assessment for each item
- [ ] Discrepancy tracking
- [ ] Document upload
- [ ] Digital signature
- [ ] Validation before submit

#### B.2 GRN List & History
**File**: `src/app/workflows/grn/_components/grn-list.tsx`

**Functionality**:
- [ ] List all GRNs created
- [ ] Filter by status, PO, date range
- [ ] Show items received count vs ordered
- [ ] Highlight discrepancies
- [ ] Sort by date, PO number
- [ ] Status badge (DRAFT, COMPLETED, REJECTED)

#### B.3 Server Actions for GRN
**File**: `src/app/_actions/workflow.ts`

```typescript
// Create GRN
async function createGRN(grnData: {
  poId: string
  items: Array<{ poItemId: string, quantityReceived: number, ... }>
  inspectionNotes: string
  damageReport?: string
  attachments?: File[]
}) {
  const po = documentStore.get(grnData.poId)

  const grn: GoodsReceivedNote = {
    id: generateId(),
    grNumber: generateGRNNumber(),
    poId: grnData.poId,
    requisitionId: po.metadata.requisitionId,
    receivedDate: new Date(),
    receivedBy: currentUser,
    items: processItemsReceived(grnData.items),
    inspectionNotes: grnData.inspectionNotes,
    signedBy: currentUser,
    signedAt: new Date(),
    status: 'COMPLETED',
    readyForPayment: !hasDiscrepancies(grnData.items),
    attachments: []
  }

  grnStore.set(grn.id, grn)

  // Trigger Payment Voucher creation
  if (grn.readyForPayment) {
    await autoCreatePaymentVoucher(grn)
  }

  return successResponse(grn)
}

// Get GRN by PO
async function getGRNByPO(poId: string) {
  return Array.from(grnStore.values()).find(grn => grn.poId === poId)
}

// Get all GRNs
async function getAllGRNs(page: number, pageSize: number) {
  const grns = Array.from(grnStore.values())
    .sort((a, b) => b.signedAt.getTime() - a.signedAt.getTime())

  return paginatedResponse(grns, page, pageSize)
}
```

---

## Part C: Payment Voucher Workflow

### Business Process Flow

```
GRN Completed
  ↓
Auto-Create Payment Voucher (linked to GRN & PO)
  ↓
Payment Voucher Status: SUBMITTED → IN_APPROVAL
  ↓
3-Stage Approval Process:
├── Stage 1: Director Finance Reviews & Approves
├── Stage 2: Accountant Reviews & Approves
└── Stage 3: Principal Officer Final Approval
  ↓
Payment Voucher Status: APPROVED
  ↓
Auto-Generate Payment Reference & QR Code
  ↓
Initiate Payment to Vendor
  ↓
Notify Stakeholders: Payment Approved & Initiated
  ↓
System Logs Actions & Updates Dashboards
  ↓
End of Workflow
```

### Data Model

```typescript
interface PaymentVoucher extends WorkflowDocument {
  type: 'PAYMENT_VOUCHER'

  // Links
  grnId: string            // Reference to GRN
  poId: string             // Reference to PO
  requisitionId: string    // Reference to original requisition

  // Voucher Details
  voucherNumber: string    // e.g., "PV-2024-009012"

  // Vendor/Payment Info
  vendorId: string
  vendorName: string
  vendorEmail: string

  // Bank Account (where payment goes)
  bankInfo: {
    accountNumber: string
    accountName: string
    bankCode: string
    bankName: string
    swiftCode?: string
    branchCode?: string
  }

  // Cost Details
  grossAmount: number
  tax: number
  netAmount: number
  deductions?: Array<{
    type: string           // 'WITHHOLDING_TAX', 'PENALTY', etc.
    amount: number
  }>
  currency: string

  // Financial Coding
  voteCode: string         // e.g., "4100.03.01.02" (cost center coding)
  costCenter: string
  glAccount: string

  // Support Documents
  supportingDocs: Array<{
    type: 'INVOICE' | 'GRN' | 'PO' | 'OTHER'
    fileName: string
    fileId: string
    uploadedAt: Date
  }>

  // Payment Authorization
  approvalStages: Array<{
    stageNumber: 1 | 2 | 3
    stageName: 'Director Finance' | 'Accountant' | 'Principal Officer'
    assignedTo: User
    status: 'PENDING' | 'APPROVED' | 'REJECTED'
    approvedAt?: Date
    approvedBy?: User
    comments?: string
  }>

  // Payment Details
  paymentMethod: 'BANK_TRANSFER' | 'CHEQUE' | 'CASH'
  paymentReference?: string
  qrCode?: string
  qrCodeData?: {
    vendorCode: string
    amount: number
    date: Date
    reference: string
  }

  // Status tracking
  status: 'DRAFT' | 'SUBMITTED' | 'IN_APPROVAL' | 'APPROVED' | 'REJECTED' | 'PAID'
  currentStage: 1 | 2 | 3

  // Timestamps
  createdAt: Date
  approvedAt?: Date
  paidAt?: Date
}
```

### Implementation Tasks

#### C.1 Create Payment Voucher List Page
**Files**:
- `src/app/workflows/payment-vouchers/page.tsx`
- `src/app/workflows/payment-vouchers/_components/pv-client.tsx`
- `src/app/workflows/payment-vouchers/_components/pv-table.tsx`

**Table Columns**:
```typescript
const columns: ColumnDef<PaymentVoucher>[] = [
  { accessorKey: 'voucherNumber', header: 'Voucher #' },
  { accessorKey: 'vendorName', header: 'Vendor' },
  { accessorKey: 'netAmount', header: 'Amount', cell: formatCurrency },
  { accessorKey: 'voteCode', header: 'Vote Code' },
  {
    accessorKey: 'status',
    header: 'Status',
    cell: ({ row }) => <Badge>{row.original.status}</Badge>
  },
  {
    accessorKey: 'currentStage',
    header: 'Stage',
    cell: ({ row }) => `${row.original.currentStage}/3`
  },
  {
    id: 'actions',
    cell: ({ row }) => <ViewButton id={row.original.id} />
  }
]
```

**Functionality**:
- [ ] List all payment vouchers
- [ ] Filter by status, vendor, stage
- [ ] Sort by voucher number, amount, date
- [ ] Show stage progress (1/3, 2/3, 3/3)
- [ ] Link to originating PO and GRN
- [ ] 10 items per page with pagination

#### C.2 Create Payment Voucher Detail Page
**Files**:
- `src/app/workflows/payment-vouchers/[id]/page.tsx`
- `src/app/workflows/payment-vouchers/_components/pv-detail-client.tsx`

**Layout**:
```
Header
├── Voucher Number + Status Badge
├── Stage Progress (1/3, 2/3, 3/3)
├── Current Approver (if IN_APPROVAL)
└── Links to PO, GRN, Requisition

Main Content (2-column)
├── Column 1: Voucher Details
│   ├── Vendor Information
│   │   ├── Vendor Name
│   │   ├── Account Details
│   │   └── Payment Method
│   ├── Cost Details
│   │   ├── Gross Amount
│   │   ├── Tax
│   │   ├── Deductions
│   │   └── Net Amount
│   ├── Financial Coding
│   │   ├── Vote Code
│   │   ├── Cost Center
│   │   └── GL Account
│   └── Supporting Documents
│       ├── Invoice
│       ├── GRN
│       ├── PO
│       └── Other docs
│
└── Sidebar: 3-Stage Approval
    ├── Stage 1: Director Finance
    │   ├── Status (Pending/Approved/Rejected)
    │   ├── Date
    │   └── Approver Name
    ├── Stage 2: Accountant
    ├── Stage 3: Principal Officer
    └── Approval Action Panel (if current stage)
```

#### C.3 Payment Voucher Approval Workflow
**File**: `src/app/workflows/payment-vouchers/_components/pv-approval-panel.tsx`

**Stage 1: Director Finance**
- [ ] Show PO and GRN summary
- [ ] Display cost breakdown
- [ ] Show vote code and GL account
- [ ] Approve/Reject buttons
- [ ] Optional comments
- [ ] View supporting documents

**Stage 2: Accountant**
- [ ] Verify bank account info:
  - [ ] Validate account number format
  - [ ] Verify vendor against approved vendors list
  - [ ] Check for duplicate payments
- [ ] Review deductions and tax
- [ ] Verify GL coding
- [ ] Approve/Reject buttons
- [ ] Comments field

**Stage 3: Principal Officer (Final)**
- [ ] Review previous approvals
- [ ] Final authorization check
- [ ] Approve/Reject button
- [ ] Generate Payment Reference & QR Code on approval
- [ ] Initiate payment process

#### C.4 Payment Voucher Creation (Auto from GRN)
**Files**: Server action in `src/app/_actions/workflow.ts`

```typescript
// Called automatically when GRN is completed
async function autoCreatePaymentVoucher(grn: GoodsReceivedNote) {
  const po = documentStore.get(grn.poId)

  const pv: PaymentVoucher = {
    id: generateId(),
    type: 'PAYMENT_VOUCHER',
    voucherNumber: generatePVNumber(),
    grnId: grn.id,
    poId: po.id,
    requisitionId: po.metadata.requisitionId,

    vendorId: po.vendorId,
    vendorName: po.vendorName,
    vendorEmail: po.vendorEmail,

    // Will be filled during accounting review
    bankInfo: {
      accountNumber: '',
      accountName: '',
      bankCode: '',
      bankName: '',
    },

    grossAmount: po.totalAmount,
    tax: calculateTax(po.totalAmount),
    netAmount: po.totalAmount - tax,
    currency: po.currency,

    // Will be filled during approval
    voteCode: '',
    costCenter: '',
    glAccount: '',

    supportingDocs: [
      {
        type: 'PO',
        fileName: `${po.poNumber}.pdf`,
        fileId: po.id,
        uploadedAt: new Date()
      },
      {
        type: 'GRN',
        fileName: `${grn.grNumber}.pdf`,
        fileId: grn.id,
        uploadedAt: new Date()
      }
    ],

    approvalStages: [
      {
        stageNumber: 1,
        stageName: 'Director Finance',
        assignedTo: getDirectorFinanceUser(),
        status: 'PENDING'
      },
      {
        stageNumber: 2,
        stageName: 'Accountant',
        assignedTo: getAccountantUser(),
        status: 'PENDING'
      },
      {
        stageNumber: 3,
        stageName: 'Principal Officer',
        assignedTo: getPrincipalOfficerUser(),
        status: 'PENDING'
      }
    ],

    paymentMethod: 'BANK_TRANSFER',
    status: 'SUBMITTED',
    currentStage: 1,

    createdAt: new Date(),
    approvalLog: []
  }

  paymentVoucherStore.set(pv.id, pv)
  return pv
}

// PV Approval - each stage
async function approvePaymentVoucher(
  pvId: string,
  stageNumber: 1 | 2 | 3,
  data?: {
    bankInfo?: PaymentVoucher['bankInfo']
    voteCode?: string
    costCenter?: string
    glAccount?: string
  },
  comments?: string
) {
  const pv = paymentVoucherStore.get(pvId)
  const stage = pv.approvalStages.find(s => s.stageNumber === stageNumber)

  // Verify current approver
  if (stage.assignedTo.id !== currentUser.id) {
    return forbiddenResponse('Not assigned to this stage')
  }

  // Update stage-specific data
  if (stageNumber === 2 && data?.bankInfo) {
    pv.bankInfo = data.bankInfo
    pv.voteCode = data.voteCode
    pv.costCenter = data.costCenter
    pv.glAccount = data.glAccount
  }

  // Mark stage as approved
  stage.status = 'APPROVED'
  stage.approvedAt = new Date()
  stage.approvedBy = currentUser
  stage.comments = comments

  // Move to next stage or mark approved
  if (stageNumber < 3) {
    pv.currentStage = stageNumber + 1
    pv.status = 'IN_APPROVAL'
  } else {
    // Final approval
    pv.status = 'APPROVED'
    pv.approvedAt = new Date()

    // Generate payment reference & QR code
    pv.paymentReference = generatePaymentReference()
    pv.qrCodeData = {
      vendorCode: pv.vendorId,
      amount: pv.netAmount,
      date: new Date(),
      reference: pv.paymentReference
    }
    // In real implementation, generate actual QR code image
    pv.qrCode = generateQRCode(pv.qrCodeData)

    // Notify stakeholders
    await notifyStakeholders('PAYMENT_APPROVED', {
      vendor: pv.vendorName,
      amount: pv.netAmount,
      reference: pv.paymentReference
    })
  }

  // Log approval
  approvalLogsStore.get(pvId)?.push({
    id: generateId(),
    approver: currentUser,
    action: 'APPROVED',
    timestamp: new Date(),
    comments,
    stage: stageNumber
  })

  return successResponse(pv)
}

// PV Rejection at any stage
async function rejectPaymentVoucher(
  pvId: string,
  reason: string
) {
  const pv = paymentVoucherStore.get(pvId)
  pv.status = 'REJECTED'

  // Log rejection
  approvalLogsStore.get(pvId)?.push({
    action: 'REJECTED',
    timestamp: new Date(),
    comments: reason
  })

  // Notify requester/accountant

  return successResponse(pv)
}
```

#### C.5 QR Code Generation
**File**: `src/lib/qr-code.ts` (new utility file)

```typescript
import QRCode from 'qrcode'

export async function generateQRCode(data: {
  vendorCode: string
  amount: number
  date: Date
  reference: string
}): Promise<string> {
  const qrString = JSON.stringify({
    vc: data.vendorCode,
    amt: data.amount,
    dt: data.date.toISOString(),
    ref: data.reference
  })

  // Generate QR code as data URL
  return QRCode.toDataURL(qrString)
}
```

---

## Database Schema (When Migrating)

```sql
-- Purchase Orders
CREATE TABLE purchase_orders (
  id UUID PRIMARY KEY,
  po_number VARCHAR UNIQUE,
  requisition_id UUID REFERENCES requisitions(id),
  vendor_id VARCHAR,
  vendor_name VARCHAR,
  total_amount DECIMAL,
  delivery_type VARCHAR,
  status VARCHAR,
  current_stage INT,
  approved_at TIMESTAMP,
  created_at TIMESTAMP,
  updated_at TIMESTAMP
);

-- Goods Received Notes
CREATE TABLE goods_received_notes (
  id UUID PRIMARY KEY,
  gr_number VARCHAR UNIQUE,
  po_id UUID REFERENCES purchase_orders(id),
  received_date TIMESTAMP,
  received_by UUID,
  inspection_notes TEXT,
  status VARCHAR,
  signed_at TIMESTAMP,
  created_at TIMESTAMP
);

-- Payment Vouchers
CREATE TABLE payment_vouchers (
  id UUID PRIMARY KEY,
  voucher_number VARCHAR UNIQUE,
  grn_id UUID REFERENCES goods_received_notes(id),
  po_id UUID REFERENCES purchase_orders(id),
  vendor_name VARCHAR,
  bank_account VARCHAR,
  net_amount DECIMAL,
  vote_code VARCHAR,
  status VARCHAR,
  current_stage INT,
  payment_reference VARCHAR,
  qr_code TEXT,
  approved_at TIMESTAMP,
  paid_at TIMESTAMP,
  created_at TIMESTAMP
);

-- Payment Voucher Approval Stages
CREATE TABLE pv_approval_stages (
  id UUID PRIMARY KEY,
  payment_voucher_id UUID REFERENCES payment_vouchers(id),
  stage_number INT,
  stage_name VARCHAR,
  assigned_to UUID,
  status VARCHAR,
  approved_at TIMESTAMP,
  comments TEXT
);

-- Approval Logs (extended)
ALTER TABLE approval_logs ADD COLUMN stage_number INT;
ALTER TABLE approval_logs ADD COLUMN document_type VARCHAR;
```

---

## Testing Scenarios

### Scenario 1: Complete Happy Path
```
1. Create Requisition
2. Submit → Approvals (4 stages)
3. Requisition Approved → Auto-create PO
4. PO Approval by Principal Officer
5. Create GRN with items received
6. GRN Completed → Auto-create Payment Voucher
7. PV Stage 1: Director Finance Approves
8. PV Stage 2: Accountant Approves (with bank info)
9. PV Stage 3: Principal Officer Final Approval
10. Payment Reference Generated
11. QR Code Generated
12. Payment Initiated
13. Stakeholders Notified
14. Dashboards Updated
```

### Scenario 2: Items with Discrepancies
```
1. Requisition → PO → GRN
2. GRN: Items received with discrepancies
   - Qty ordered: 100, Qty received: 95
   - Some items marked 'Damaged'
3. GRN Status: COMPLETED (but with flag)
4. Payment Voucher created
5. PV shows discrepancies in support docs
6. Approvers can see issues and approve anyway
```

### Scenario 3: Rejection at PV Stage 2
```
1. PV goes to Director Finance → Approves
2. PV goes to Accountant → Rejects (invalid bank account)
3. PV Status: REJECTED
4. Notification sent to accounts/vendor
5. Accounts corrects info
6. Resubmit PV
7. Continues from Stage 2 (Accountant)
```

---

## File Structure

```
src/
├── app/
│   └── workflows/
│       ├── purchase-orders/
│       │   ├── page.tsx
│       │   ├── [id]/
│       │   │   └── page.tsx
│       │   └── _components/
│       │       ├── po-client.tsx
│       │       ├── po-table.tsx
│       │       ├── po-detail-client.tsx
│       │       └── po-approval-panel.tsx
│       │
│       ├── grn/
│       │   ├── page.tsx
│       │   └── _components/
│       │       ├── grn-client.tsx
│       │       ├── grn-form.tsx
│       │       ├── grn-list.tsx
│       │       └── grn-detail.tsx
│       │
│       └── payment-vouchers/
│           ├── page.tsx
│           ├── [id]/
│           │   └── page.tsx
│           └── _components/
│               ├── pv-client.tsx
│               ├── pv-table.tsx
│               ├── pv-detail-client.tsx
│               ├── pv-approval-panel.tsx
│               └── pv-stage-sections/
│                   ├── stage1-director.tsx
│                   ├── stage2-accountant.tsx
│                   └── stage3-principal.tsx
│
├── lib/
│   └── qr-code.ts
│
└── _actions/
    ├── workflow.ts (add PO/GRN/PV functions)
    └── notifications.ts (new)
```

---

## Implementation Order

1. **Week 1**: Purchase Order (list, detail, approval, auto-creation)
2. **Week 2**: Goods Received Note (form, list, auto-voucher creation)
3. **Week 3**: Payment Voucher (list, detail, 3-stage approval, QR code)
4. **Week 4**: Testing, refinement, notifications

---

## Estimated Effort

| Component | Est. Hours | Notes |
|-----------|-----------|-------|
| PO Implementation | 8 | List, detail, approval, auto-create |
| GRN Implementation | 8 | Form, list, item receipt tracking |
| PV Implementation | 12 | List, detail, 3-stage approval, QR |
| Testing & Refinement | 4 | Integration testing |
| **Total** | **32 hours** | **~4 weeks part-time** |

---

**Status**: Ready for development
**Next Step**: Begin with Purchase Order implementation following the specifications above
