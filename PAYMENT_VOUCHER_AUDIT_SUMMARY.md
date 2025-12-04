# Payment Voucher "Create" Button - Complete Explanation

## TL;DR (30 Seconds)

When a user clicks the **"Create Payment Voucher"** button:

1. ✅ Form validates that a purchase order is selected
2. ✅ Client collects PO details and user information
3. ✅ Sends data to `createPaymentVoucher()` server action
4. ✅ Server generates unique PV number (e.g., PV-2024-1001)
5. ✅ Server initializes 3-stage approval chain (all PENDING)
6. ✅ Server stores payment voucher in DRAFT status
7. ✅ Client shows success toast
8. ✅ React Query cache invalidated (triggers refetch)
9. ✅ User redirected to `/payment-vouchers` list
10. ✅ New payment voucher appears with DRAFT status

**Result**: A new payment voucher is created in DRAFT status, ready to be edited or submitted for approval.

---

## Complete Step-by-Step Breakdown

### STEP 1: User Interface Setup

**Component**: `PVCreateClient` (Client Component)
**Location**: `src/app/(private)/(main)/payment-vouchers/create/_components/pv-create-client.tsx`

The component renders with:
- Dropdown selector for approved purchase orders
- PO details preview (when selected)
- Line items table (when selected)
- "Create Payment Voucher" button (when PO selected)

### STEP 2: Load Approved POs

**Hook**: `usePurchaseOrders()`
**When**: Component mounts

The component fetches all purchase orders using React Query, then filters to show only:
```javascript
approvedPOs = purchaseOrders.filter(po => po.status === 'APPROVED')
```

Only **APPROVED** purchase orders can be converted to payment vouchers.

### STEP 3: User Selects a PO

**Event Handler**: `handleSelectPO(poId)`

When user selects a PO from dropdown:
```javascript
const po = approvedPOs.find(p => p.id === poId)
setSelectedPO(po)
```

The component re-renders showing:
- Vendor information
- Department
- Total amount
- Line items table
- "Create Payment Voucher" button (now enabled)

### STEP 4: User Clicks "Create Payment Voucher"

**Event Handler**: `handleCreatePV()`
**Button**: Green button with checkmark icon

When clicked:

#### 4.1 Validation
```javascript
if (!selectedPO) {
  toast.error('Please select a purchase order')
  return
}
```

#### 4.2 Build Request Object
From the selected PO and user context, create `CreatePaymentVoucherRequest`:

```javascript
const pvData: CreatePaymentVoucherRequest = {
  title: `Payment for ${selectedPO.poNumber}`,
  description: selectedPO.description,
  vendorId: selectedPO.vendorId,
  vendorName: selectedPO.vendorName,
  department: selectedPO.department,
  departmentId: selectedPO.departmentId,
  paymentDueDate: new Date(new Date().getTime() + 30 * 24 * 60 * 60 * 1000),  // +30 days
  priority: 'MEDIUM',
  paymentMethod: 'BANK_TRANSFER',
  items: selectedPO.items.map(item => ({
    poItemId: item.id,
    itemNumber: item.itemNumber,
    description: item.description,
    category: item.category,
    quantity: item.quantity,
    unitPrice: item.unitPrice,
    unit: item.unit,
    totalPrice: item.totalPrice,
    notes: item.notes
  })),
  budgetCode: selectedPO.budgetCode,
  costCenter: selectedPO.costCenter,
  projectCode: selectedPO.projectCode,
  createdBy: userId,
  createdByName: userName,
  createdByRole: userRole,
  sourcePurchaseOrderId: selectedPO.id,
  sourcePurchaseOrderNumber: selectedPO.poNumber,
  sourceRequisitionId: selectedPO.sourceRequisitionId,
  sourceRequisitionNumber: selectedPO.sourceRequisitionNumber
}
```

#### 4.3 Set Loading State
```javascript
setIsCreating(true)
// Button text changes to "Creating..."
// Button becomes disabled
```

#### 4.4 Call React Query Mutation
```javascript
await savePVMutation.mutateAsync(pvData)
```

---

### STEP 5: Server-Side Processing

**Server Action**: `createPaymentVoucher(data)`
**Location**: `src/app/_actions/payment-vouchers.ts`
**Lines**: 163-261

#### 5.1 Generate Unique PV Number
```javascript
function generatePVNumber(): string {
  const year = new Date().getFullYear()  // 2024
  pvCounter++                            // Global counter increments
  return `PV-${year}-${pvCounter.toString().padStart(4, '0')}`
}
// Result: "PV-2024-1001"
```

#### 5.2 Generate Unique PV ID
```javascript
const pvId = `pv-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`
// Result: "pv-1733328400000-abc123xyz"
```

#### 5.3 Initialize 3-Stage Approval Chain
```javascript
function initializePVApprovalChain(): PVApprovalRecord[] {
  return [
    {
      stageNumber: 1,
      stageName: 'Finance Manager Review',
      assignedTo: 'finance-manager-1',
      assignedRole: 'FINANCE_MANAGER',
      status: 'PENDING'
    },
    {
      stageNumber: 2,
      stageName: 'Approval Authority Review',
      assignedTo: 'approval-authority-1',
      assignedRole: 'APPROVAL_AUTHORITY',
      status: 'PENDING'
    },
    {
      stageNumber: 3,
      stageName: 'Director Approval',
      assignedTo: 'director-1',
      assignedRole: 'DIRECTOR',
      status: 'PENDING'
    }
  ]
}
```

All three stages start as **PENDING** (not yet active).

#### 5.4 Build Complete PaymentVoucher Object
```javascript
const paymentVoucher: PaymentVoucher = {
  id: pvId,
  pvNumber: 'PV-2024-1001',
  title: 'Payment for PO-2024-001',
  description: '...',
  vendorId: 'vendor-123',
  vendorName: 'Supplier Inc',
  department: 'Finance',
  departmentId: 'dept-456',

  // Requestor (creator)
  requestedBy: userId,
  requestedByName: userName,
  requestedByRole: userRole,
  requestedDate: now,

  // Payment Details
  paymentDueDate: (30 days from now),
  priority: 'MEDIUM',
  paymentMethod: 'BANK_TRANSFER',
  status: 'DRAFT',  // ← Key: Initial status

  // Line Items (copied from PO)
  items: [
    {
      id: 'pvi-pvId-0',
      pvId: pvId,
      poItemId: 'po-item-1',
      itemNumber: 1,
      description: 'Item description',
      quantity: 10,
      unitPrice: 100,
      totalPrice: 1000,
      // ... more fields
    }
    // More items...
  ],

  // Financial Summary
  totalAmount: 5000,
  currency: 'ZMW',

  // Approval Chain (Not active yet)
  approvalChain: [
    { stageNumber: 1, status: 'PENDING', ... },
    { stageNumber: 2, status: 'PENDING', ... },
    { stageNumber: 3, status: 'PENDING', ... }
  ],
  currentApprovalStage: 1,
  totalApprovalStages: 3,

  // Action History (Creation recorded)
  actionHistory: [
    {
      id: 'pva-pvId-create',
      actionType: 'CREATE',
      performedBy: userId,
      performedByName: userName,
      performedByRole: userRole,
      performedAt: now,
      newStatus: 'DRAFT',
      metadata: {
        sourcePurchaseOrderId: po.id,
        sourcePurchaseOrderNumber: 'PO-2024-001',
        autoCreated: false
      }
    }
  ],

  // PO Linking (Full Traceability)
  sourcePurchaseOrderId: 'po-456',
  sourcePurchaseOrderNumber: 'PO-2024-001',
  createdFromPurchaseOrder: true,

  // Requisition Linking
  sourceRequisitionId: 'req-789',
  sourceRequisitionNumber: 'REQ-2024-001',

  // Financial Codes
  budgetCode: 'BUDGET-001',
  costCenter: 'CC-123',
  projectCode: 'PROJ-456',

  // Timestamps
  createdAt: now,
  updatedAt: now
}
```

#### 5.5 Store in Mock Storage
```javascript
mockPaymentVouchers.push(paymentVoucher)
```

(In production: Would save to database)

#### 5.6 Return Success Response
```javascript
return {
  success: true,
  data: paymentVoucher,
  message: 'Payment voucher created successfully'
}
```

---

### STEP 6: Client-Side Mutation Success Handler

**Hook**: `useSavePaymentVoucher()`
**When**: Server returns `success: true`

#### 6.1 Show Success Toast
```javascript
toast.success('Payment voucher created successfully')
```

User sees confirmation message at bottom right.

#### 6.2 Invalidate React Query Cache
```javascript
queryClient.invalidateQueries({
  queryKey: [QUERY_KEYS.PAYMENT_VOUCHERS.ALL]
})
queryClient.invalidateQueries({
  queryKey: [QUERY_KEYS.PAYMENT_VOUCHERS.STATS]
})
```

This triggers queries to be marked as stale and refetched (will fetch updated list from server).

#### 6.3 Call onSuccess Callback
```javascript
onSuccess?.()  // Callback from parent component
```

In `PVCreateClient`, the callback is:
```javascript
const savePVMutation = useSavePaymentVoucher(() => {
  router.push('/payment-vouchers')  // Navigate to list
})
```

---

### STEP 7: Redirect and Display Result

**Navigation**: `router.push('/payment-vouchers')`

1. Button loading state ends: `setIsCreating(false)`
2. Browser navigates to `/payment-vouchers`
3. Payment vouchers list page loads
4. React Query fetches latest list (includes new PV)
5. List renders with new PV at top

**User sees**:
```
Payment Vouchers
────────────────────────────────────
PV-2024-1001 | Supplier Inc | DRAFT | Edit | Delete  ← NEW
PV-2024-1000 | Hardware Ltd | APPROVED | View
PV-2024-999  | Office Sup   | PAID | View
```

---

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                         USER BROWSER                             │
│                                                                  │
│  PVCreateClient Component (React)                              │
│  ├─ usePurchaseOrders() → Fetch approved POs                   │
│  ├─ Select PO from dropdown → Display details                  │
│  ├─ Click "Create" button                                      │
│  │   └─ handleCreatePV()                                       │
│  │       ├─ Validate PO selected                               │
│  │       ├─ Build CreatePaymentVoucherRequest                  │
│  │       └─ savePVMutation.mutateAsync(pvData)                 │
│  │                                                              │
│  └─ React Query Mutation (useSavePaymentVoucher)              │
│      ├─ Call server action: createPaymentVoucher()            │
│      ├─ Receive response                                       │
│      ├─ onSuccess: Show toast                                 │
│      ├─ onSuccess: Invalidate cache                           │
│      ├─ onSuccess: Redirect                                   │
│      └─ Component unmounts                                    │
│                                                                  │
└──────────────────────┬──────────────────────────────────────────┘
                       │ HTTP POST (Network Request)
                       ↓
┌─────────────────────────────────────────────────────────────────┐
│                      SERVER (Node.js)                            │
│                                                                  │
│  Server Action: createPaymentVoucher()                          │
│  ├─ Generate PV Number: "PV-2024-1001"                        │
│  ├─ Generate PV ID: "pv-1733328400000-abc123xyz"              │
│  ├─ Initialize 3-stage approval chain                          │
│  ├─ Map PO items to PV items                                   │
│  ├─ Create action history entry                                │
│  ├─ Store: mockPaymentVouchers.push(paymentVoucher)            │
│  └─ Return: { success: true, data: paymentVoucher, ... }      │
│                                                                  │
└──────────────────────┬──────────────────────────────────────────┘
                       │ HTTP 200 OK (Response)
                       ↓
┌─────────────────────────────────────────────────────────────────┐
│                      CLIENT (Continued)                          │
│                                                                  │
│  Mutation onSuccess Handler                                     │
│  ├─ toast.success("Payment voucher created successfully")       │
│  ├─ queryClient.invalidateQueries(...)                         │
│  ├─ router.push('/payment-vouchers')                           │
│  └─ Redirect to list page                                      │
│                                                                  │
│  Payment Vouchers List Page Loads                              │
│  ├─ usePaymentVouchers() fetches list                          │
│  ├─ List includes new PV with DRAFT status                    │
│  └─ User sees: PV-2024-1001 | DRAFT | Edit                   │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## Payment Voucher Status After Creation

### Current State
```
Status: DRAFT
Approval Stage: 1 (but not active)
Can Edit: YES
Can Delete: YES
Can Submit for Approval: YES
```

### What Happens Next
1. **Option A**: User can EDIT the PV
   - Change due date
   - Change priority
   - Add/remove items
   - Update vendor details

2. **Option B**: User can DELETE the PV
   - Only possible while DRAFT
   - Removes from system

3. **Option C**: User can SUBMIT for approval
   - Status changes to SUBMITTED
   - Approval chain activates
   - Finance Manager notified for Stage 1
   - PV moves to IN_REVIEW status

---

## Complete Request/Response Cycle

### CLIENT SENDS
```json
{
  "title": "Payment for PO-2024-001",
  "description": "...",
  "vendorId": "vendor-123",
  "vendorName": "Supplier Inc",
  "department": "Finance",
  "departmentId": "dept-456",
  "paymentDueDate": "2025-01-03",
  "priority": "MEDIUM",
  "paymentMethod": "BANK_TRANSFER",
  "items": [
    {
      "poItemId": "po-item-1",
      "itemNumber": 1,
      "description": "Item A",
      "quantity": 10,
      "unitPrice": 100,
      "totalPrice": 1000
    }
  ],
  "budgetCode": "BUDGET-001",
  "costCenter": "CC-123",
  "projectCode": "PROJ-456",
  "createdBy": "user-id-123",
  "createdByName": "John Doe",
  "createdByRole": "FINANCE_OFFICER",
  "sourcePurchaseOrderId": "po-456",
  "sourcePurchaseOrderNumber": "PO-2024-001",
  "sourceRequisitionId": "req-789",
  "sourceRequisitionNumber": "REQ-2024-001"
}
```

### SERVER RETURNS
```json
{
  "success": true,
  "message": "Payment voucher created successfully",
  "data": {
    "id": "pv-1733328400000-abc123xyz",
    "pvNumber": "PV-2024-1001",
    "status": "DRAFT",
    "title": "Payment for PO-2024-001",
    "vendorName": "Supplier Inc",
    "department": "Finance",
    "paymentDueDate": "2025-01-03",
    "priority": "MEDIUM",
    "paymentMethod": "BANK_TRANSFER",
    "items": [...],
    "totalAmount": 5000,
    "currency": "ZMW",
    "approvalChain": [
      {
        "stageNumber": 1,
        "stageName": "Finance Manager Review",
        "status": "PENDING",
        "assignedTo": "finance-manager-1"
      },
      {
        "stageNumber": 2,
        "stageName": "Approval Authority Review",
        "status": "PENDING",
        "assignedTo": "approval-authority-1"
      },
      {
        "stageNumber": 3,
        "stageName": "Director Approval",
        "status": "PENDING",
        "assignedTo": "director-1"
      }
    ],
    "actionHistory": [
      {
        "actionType": "CREATE",
        "performedBy": "user-id-123",
        "performedByName": "John Doe",
        "performedByRole": "FINANCE_OFFICER",
        "newStatus": "DRAFT"
      }
    ],
    "sourcePurchaseOrderNumber": "PO-2024-001",
    "sourceRequisitionNumber": "REQ-2024-001",
    "createdAt": "2024-12-04T10:30:00Z",
    "updatedAt": "2024-12-04T10:30:00Z"
  }
}
```

---

## Code Flow Summary

```
handleCreatePV()
  ├─ Validate selectedPO exists
  ├─ setIsCreating(true)
  ├─ Build CreatePaymentVoucherRequest object
  └─ savePVMutation.mutateAsync(pvData)
     │
     ├─ SERVER RECEIVES REQUEST
     │  ├─ createPaymentVoucher(data)
     │  ├─ Generate PV number & ID
     │  ├─ Initialize approval chain
     │  ├─ Map PO items
     │  ├─ Create action history
     │  ├─ Store in mock storage
     │  └─ Return success response
     │
     └─ CLIENT RECEIVES RESPONSE
        ├─ toast.success("created successfully")
        ├─ queryClient.invalidateQueries(...)
        ├─ router.push('/payment-vouchers')
        └─ Page redirects
```

---

## Key Files and Functions

| File | Function | Purpose |
|------|----------|---------|
| `pv-create-client.tsx` | `handleCreatePV()` | Main handler for button click |
| `pv-create-client.tsx` | `handleSelectPO()` | Handle PO selection |
| `payment-vouchers.ts` | `createPaymentVoucher()` | Server action to create PV |
| `payment-vouchers.ts` | `generatePVNumber()` | Generate unique PV number |
| `payment-vouchers.ts` | `initializePVApprovalChain()` | Create 3-stage approval chain |
| `use-payment-voucher-queries.ts` | `useSavePaymentVoucher()` | React Query mutation |
| `payment-voucher.ts` | `CreatePaymentVoucherRequest` | Request DTO type |
| `payment-voucher.ts` | `PaymentVoucher` | Main PV interface |

---

## Performance Notes

- **Data Fetching**: POs loaded async with React Query (staleTime: 5 min)
- **Mutation**: Instant (mock storage, no DB latency)
- **Cache Invalidation**: Both ALL and STATS queries refreshed
- **Redirect**: Happens after mutation completes
- **UI Feedback**: Loading state shown, success toast displayed

---

## Security Considerations

✅ Authentication: User must be logged in (server-side check)
✅ Authorization: User role captured in action history
✅ Data Validation: Required fields validated on both client & server
✅ Audit Trail: All actions logged with user info and timestamp
✅ Data Traceability: Links to source PO and requisition
✅ Error Handling: Errors don't expose sensitive information

---

## Related Documentation

📄 **PAYMENT_VOUCHER_CREATE_AUDIT.md** - Detailed technical audit
📄 **PAYMENT_VOUCHER_QUICK_REFERENCE.md** - Quick reference guide
📄 **payment-vouchers.ts** - Server action implementation
📄 **pv-create-client.tsx** - Client component implementation
📄 **payment-voucher.ts** - Type definitions

---

## Summary

When the "Create Payment Voucher" button is clicked, a complete workflow is initiated:

1. Form validates the selection
2. Data is collected and formatted
3. Server generates unique identifiers
4. 3-stage approval chain is initialized
5. Payment voucher is created in DRAFT status
6. Audit trail is recorded
7. User gets success confirmation
8. User is redirected to the vouchers list
9. New voucher appears with full traceability

The system is designed for **security**, **auditability**, and **workflow efficiency**.
