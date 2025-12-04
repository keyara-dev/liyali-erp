# Payment Voucher Creation - Quick Reference

## What Happens When User Clicks "Create Payment Voucher"

### The Complete Flow in 60 Seconds

```
USER CLICKS "CREATE PAYMENT VOUCHER"
        ↓
Check if PO selected?
        ├─ NO → Show error toast, stay on page
        └─ YES ↓
        ↓
Build data object from:
        • Selected PO (vendor, items, budget)
        • User info (userId, name, role)
        • Defaults (30-day due date, MEDIUM priority)
        ↓
Send to server action: createPaymentVoucher()
        ↓
SERVER GENERATES:
        • Unique ID: pv-1733328400000-abc123xyz
        • Unique Number: PV-2024-1001
        • 3-Stage Approval Chain (all PENDING)
        • Action History (CREATE entry)
        ↓
SERVER STORES:
        • Payment voucher in mock storage
        • Status: DRAFT
        • Ready to edit or submit
        ↓
CLIENT RECEIVES SUCCESS:
        • Show "Payment voucher created successfully"
        • Refresh payment vouchers list
        • Redirect to /payment-vouchers
        ↓
USER SEES:
        • New payment voucher in DRAFT status
        • Ready to edit, delete, or submit
```

---

## Key Data Created

### Payment Voucher Object
```javascript
{
  id: "pv-1733328400000-abc123xyz",
  pvNumber: "PV-2024-1001",
  status: "DRAFT",

  // From PO
  vendorName: "Supplier Inc",
  department: "Finance",
  items: [...],  // Copied from PO
  totalAmount: 5000,

  // Approval (Not Active Yet)
  approvalChain: [
    { stageName: "Finance Manager Review", status: "PENDING" },
    { stageName: "Approval Authority Review", status: "PENDING" },
    { stageName: "Director Approval", status: "PENDING" }
  ],

  // Traceability
  sourcePurchaseOrderNumber: "PO-2024-001",
  sourceRequisitionNumber: "REQ-2024-001",

  // Created By
  requestedBy: "user-id",
  requestedByName: "John Doe",
  requestedByRole: "FINANCE_OFFICER"
}
```

---

## UI Flow

```
STEP 1: SELECT PO
┌─────────────────────────────────┐
│ Select Purchase Order *         │
│ ┌─────────────────────────────┐ │
│ │ PO-2024-001 - Supplier Inc  │ │  ← Click to select
│ │ (ZMW 5,000)                 │ │
│ └─────────────────────────────┘ │
└─────────────────────────────────┘

         ↓

STEP 2: REVIEW PO DETAILS
┌─────────────────────────────────┐
│ Purchase Order Details          │
├─────────────────────────────────┤
│ Vendor: Supplier Inc            │
│ Department: Finance             │
│ Total Amount: ZMW 5,000         │
│ Status: APPROVED                │
└─────────────────────────────────┘

         ↓

STEP 3: REVIEW LINE ITEMS
┌─────────────────────────────────┐
│ Line Items                      │
├─────────────────────────────────┤
│ # | Description | Qty | Total   │
│ 1 | Item A      | 10  | 1,000   │
│ 2 | Item B      | 20  | 4,000   │
├─────────────────────────────────┤
│ Total: ZMW 5,000                │
└─────────────────────────────────┘

         ↓

STEP 4: CREATE
┌─────────────────────────────────┐
│ [✓ Create Payment Voucher]      │  ← Click
│ [Cancel]                        │
└─────────────────────────────────┘

         ↓ (Button shows "Creating...")

STEP 5: SUCCESS
┌─────────────────────────────────┐
│ ✓ Payment voucher created       │
│   successfully                  │
│                                 │
│ Redirecting to payments...      │
└─────────────────────────────────┘

         ↓

PAYMENT VOUCHERS LIST
┌─────────────────────────────────┐
│ Payment Vouchers                │
├─────────────────────────────────┤
│ PV-2024-1001 | DRAFT | EDIT     │  ← New PV
│ PV-2024-1000 | APPROVED | VIEW  │
│ PV-2024-999  | PAID | VIEW      │
└─────────────────────────────────┘
```

---

## Approval Chain Created

When PV is created, a 3-stage approval chain is automatically initialized:

```
┌────────────────────────────────────────────────────┐
│                  APPROVAL CHAIN                     │
├────────────────────────────────────────────────────┤
│                                                    │
│  Stage 1: Finance Manager Review                  │
│  └─ Assigned to: finance-manager-1               │
│  └─ Status: PENDING (waiting for input)          │
│                                                    │
│  Stage 2: Approval Authority Review              │
│  └─ Assigned to: approval-authority-1            │
│  └─ Status: PENDING (waiting)                    │
│                                                    │
│  Stage 3: Director Approval                      │
│  └─ Assigned to: director-1                      │
│  └─ Status: PENDING (waiting)                    │
│                                                    │
└────────────────────────────────────────────────────┘

Note: Approval chain only activates when PV is SUBMITTED
      from DRAFT status
```

---

## Comparison: PO → PV Auto-Population

### What Gets Copied From PO
```
Vendor Info:
  • vendorId
  • vendorName

Department Info:
  • department
  • departmentId

Financial Codes:
  • budgetCode
  • costCenter
  • projectCode

Line Items:
  • All items from PO
  • Quantities
  • Prices
  • Total amount
  • Currency

Traceability:
  • sourcePurchaseOrderId
  • sourcePurchaseOrderNumber
  • sourceRequisitionId (from PO)
  • sourceRequisitionNumber (from PO)
```

### What Gets Set to Defaults
```
Priority: "MEDIUM"
Payment Method: "BANK_TRANSFER"
Due Date: +30 days from creation
Status: "DRAFT"
```

---

## Server Operations (Line-by-Line)

```typescript
// 1. Generate unique number
PV Number = "PV-2024-1001"  (auto-incremented)

// 2. Generate unique ID
PV ID = "pv-1733328400000-abc123xyz"  (timestamp + random)

// 3. Initialize approval chain with 3 stages
approvalChain = [
  { stageNumber: 1, status: "PENDING" },
  { stageNumber: 2, status: "PENDING" },
  { stageNumber: 3, status: "PENDING" }
]

// 4. Map PO items to PV items
items = po.items.map(item => ({
  id: "pvi-...",
  poItemId: item.id,
  description: item.description,
  quantity: item.quantity,
  unitPrice: item.unitPrice,
  totalPrice: item.totalPrice
}))

// 5. Create action history entry
actionHistory = [{
  actionType: "CREATE",
  performedBy: userId,
  performedAt: now,
  newStatus: "DRAFT"
}]

// 6. Store in mock storage
mockPaymentVouchers.push(paymentVoucher)

// 7. Return success response
return { success: true, data: paymentVoucher, message: "..." }
```

---

## Files Involved

| File | Role | Key Function |
|------|------|------|
| `create/page.tsx` | Server Component | Auth check + initial setup |
| `pv-create-client.tsx` | Client Component | UI form + PO selection |
| `payment-vouchers.ts` | Server Action | `createPaymentVoucher()` |
| `use-payment-voucher-queries.ts` | React Query Hook | `useSavePaymentVoucher()` |
| `payment-voucher.ts` | Type Definitions | Interfaces & types |

---

## Error Scenarios

### Scenario 1: No PO Selected
```
User clicks "Create" without selecting PO
        ↓
Client validation fails
        ↓
Toast: "Please select a purchase order"
        ↓
User stays on create page
```

### Scenario 2: No Approved POs Available
```
User visits /payment-vouchers/create
        ↓
usePurchaseOrders() returns only non-approved POs
        ↓
Filtered list is empty
        ↓
Alert shown: "No approved purchase orders available..."
        ↓
User cannot create PV
```

### Scenario 3: Server Error
```
User clicks "Create"
        ↓
Server returns { success: false, message: "..." }
        ↓
Toast: Error message displayed
        ↓
User stays on create page and can retry
```

---

## Status Journey

```
DRAFT
  ↓ [User clicks "Submit for Approval"]
SUBMITTED
  ↓ [Finance Manager approves]
IN_REVIEW
  ↓ [Approval Authority approves]
IN_REVIEW
  ↓ [Director approves]
APPROVED
  ↓ [Accounting marks paid]
PAID
```

---

## Quick Facts

| Aspect | Value |
|--------|-------|
| **Initial Status** | DRAFT |
| **Approval Stages** | 3 (Finance Manager → Approval Authority → Director) |
| **Default Priority** | MEDIUM |
| **Default Payment Method** | BANK_TRANSFER |
| **Default Due Date** | +30 days |
| **PV Number Format** | PV-YYYY-NNNN (e.g., PV-2024-1001) |
| **Can Edit While DRAFT** | YES |
| **Can Delete While DRAFT** | YES |
| **Can Submit** | When status is DRAFT |
| **Created In** | Mock storage (in-memory) |

---

## Real-World Example

### Scenario
Finance Officer creates PV from approved PO

### Steps
1. Navigate to `/payment-vouchers/create`
2. System shows 2 approved POs:
   - PO-2024-001 (Supplier Inc, ZMW 5,000)
   - PO-2024-002 (Hardware Ltd, ZMW 2,500)
3. Select "PO-2024-001 - Supplier Inc (ZMW 5,000)"
4. Details display:
   - Vendor: Supplier Inc
   - Department: Finance
   - 2 line items (Item A, Item B)
   - Total: ZMW 5,000
5. Click "Create Payment Voucher"
6. System creates:
   - PV-2024-1001
   - 3-stage approval chain
   - Action history (CREATE)
7. System redirects to payment vouchers list
8. Finance Officer sees new PV:
   - PV-2024-1001 | DRAFT | [Edit] [Delete]

### What Happens Next
- Finance Officer can edit details (while DRAFT)
- Finance Officer can submit for approval
- System sends to Finance Manager (Stage 1)
- Finance Manager reviews and approves or rejects
- If approved, moves to Approval Authority (Stage 2)
- And so on...

---

## Benefits of This Design

✅ **Full Traceability**: Links to both PO and Requisition
✅ **Automated Workflows**: 3-stage approval chain auto-initialized
✅ **Audit Trail**: Every action recorded with user info
✅ **Data Integrity**: All PO data preserved in PV
✅ **Flexibility**: Can edit before submission
✅ **Security**: All operations logged
✅ **Efficiency**: No manual re-entry of data

---

## To Summarize

When the user clicks "Create Payment Voucher":

1. **Validates** that a PO is selected
2. **Collects** PO details and user info
3. **Sends** to server
4. **Server generates** unique ID and number
5. **Server initializes** 3-stage approval chain
6. **Server stores** the complete PV
7. **Returns** success response
8. **Shows** success toast
9. **Invalidates** cache
10. **Redirects** to payment vouchers list

The payment voucher is now in DRAFT status, ready for editing or submission to the approval workflow.
