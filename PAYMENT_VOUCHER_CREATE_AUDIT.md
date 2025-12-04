# Payment Voucher Creation Flow - Complete Audit

## Executive Summary

When a user clicks the **"Create Payment Voucher"** button, a multi-step process is triggered that:
1. Collects an approved purchase order from a dropdown
2. Displays PO details and line items for review
3. Initializes a 3-stage approval chain
4. Creates the payment voucher in DRAFT status
5. Redirects to the payment vouchers list

---

## User Journey: Step-by-Step Walkthrough

### Step 1: User Navigates to Create Page
```
Route: GET /payment-vouchers/create
↓
Server Component: CreatePaymentVoucherPage (page.tsx)
  ├─ Authenticates user via auth()
  ├─ Redirects to /login if not authenticated
  ├─ Passes userId, userName, userRole to PVCreateClient
  └─ Renders PVCreateClient component
```

### Step 2: Component Loads and Fetches Data
```
Client Component: PVCreateClient
  ├─ usePurchaseOrders() hook fetches all purchase orders
  │  └─ Filters to show only: status === 'APPROVED'
  │
  ├─ useSavePaymentVoucher() hook initialized
  │  └─ Sets up React Query mutation for create/update
  │
  └─ Renders UI with:
     ├─ Header: "Create Payment Voucher"
     ├─ PO Selector dropdown (disabled if no approved POs)
     └─ Empty state alert if no approved POs available
```

### Step 3: User Selects a Purchase Order
```
UI Event: handleSelectPO(poId)
  ├─ Finds matching PO from approvedPOs array
  ├─ Updates local state: setSelectedPO(po)
  ├─ Triggers re-render
  └─ UI displays:
     ├─ PO Details card:
     │  ├─ Vendor name
     │  ├─ Department
     │  ├─ Total amount
     │  └─ Status badge
     │
     ├─ Line Items card:
     │  ├─ Table with PO items
     │  ├─ Columns: #, Description, Category, Qty, Unit Price, Total
     │  └─ Total amount summary
     │
     └─ Action buttons:
        ├─ "Create Payment Voucher" button (enabled)
        └─ "Cancel" button
```

### Step 4: User Clicks "Create Payment Voucher"
```
UI Event: handleCreatePV()
  ├─ Validation Check:
  │  └─ if (!selectedPO) → toast.error("Please select a purchase order")
  │
  ├─ setIsCreating(true) → Button shows "Creating..."
  │
  ├─ Build CreatePaymentVoucherRequest object:
  │  │
  │  └─ From selectedPO and user context:
  │     ├─ title: "Payment for {PO.poNumber}"
  │     ├─ description: selectedPO.description
  │     ├─ vendorId: selectedPO.vendorId
  │     ├─ vendorName: selectedPO.vendorName
  │     ├─ department: selectedPO.department
  │     ├─ departmentId: selectedPO.departmentId
  │     ├─ paymentDueDate: (30 days from now)
  │     ├─ priority: 'MEDIUM' (default)
  │     ├─ paymentMethod: 'BANK_TRANSFER' (default)
  │     ├─ items: [
  │     │    {
  │     │      poItemId: item.id,
  │     │      itemNumber: item.itemNumber,
  │     │      description: item.description,
  │     │      category: item.category,
  │     │      quantity: item.quantity,
  │     │      unitPrice: item.unitPrice,
  │     │      unit: item.unit,
  │     │      totalPrice: item.totalPrice,
  │     │      notes: item.notes
  │     │    }
  │     │  ]
  │     ├─ budgetCode: selectedPO.budgetCode
  │     ├─ costCenter: selectedPO.costCenter
  │     ├─ projectCode: selectedPO.projectCode
  │     ├─ createdBy: userId
  │     ├─ createdByName: userName
  │     ├─ createdByRole: userRole
  │     ├─ sourcePurchaseOrderId: selectedPO.id
  │     ├─ sourcePurchaseOrderNumber: selectedPO.poNumber
  │     ├─ sourceRequisitionId: selectedPO.sourceRequisitionId
  │     └─ sourceRequisitionNumber: selectedPO.sourceRequisitionNumber
  │
  └─ Call useSavePaymentVoucher mutation
     └─ savePVMutation.mutateAsync(pvData)
```

---

## Server-Side Processing: createPaymentVoucher()

### Function Location
```
src/app/_actions/payment-vouchers.ts
Lines: 163-261
```

### Processing Steps

#### Step 1: Generate Unique PV Number
```typescript
function generatePVNumber(): string {
  const year = new Date().getFullYear()  // 2024
  pvCounter++                             // Increment counter
  return `PV-${year}-${pvCounter.toString().padStart(4, '0')}`
  // Result: "PV-2024-1001"
}
```

#### Step 2: Generate Unique PV ID
```typescript
const pvId = `pv-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`
// Result: "pv-1733328400000-abc123xyz"
```

#### Step 3: Initialize 3-Stage Approval Chain
```typescript
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

#### Step 4: Build PaymentVoucher Object
```typescript
const paymentVoucher: PaymentVoucher = {
  // Identity
  id: pvId,
  pvNumber: "PV-2024-1001",
  title: "Payment for PO-2024-001",
  description: "...",

  // Vendor & Department
  vendorId: "vendor-123",
  vendorName: "Supplier Inc",
  department: "Finance",
  departmentId: "dept-456",

  // Requestor Info
  requestedBy: userId,
  requestedByName: "John Doe",
  requestedByRole: "FINANCE_OFFICER",
  requestedDate: now,

  // Payment Details
  paymentDueDate: (30 days from now),
  priority: "MEDIUM",
  paymentMethod: "BANK_TRANSFER",
  status: "DRAFT",  // ← Initial status

  // Line Items (mapped from PO)
  items: [
    {
      id: "pvi-...",
      pvId: pvId,
      poItemId: poItem.id,
      itemNumber: 1,
      description: "Item description",
      quantity: 10,
      unitPrice: 100,
      totalPrice: 1000,
      ...
    }
  ],

  // Financial Summary
  totalAmount: (sum of all item totals),
  currency: "ZMW",

  // Approval Chain (3 stages, all PENDING)
  approvalChain: initializePVApprovalChain(),
  currentApprovalStage: 1,
  totalApprovalStages: 3,

  // Action History (creation recorded)
  actionHistory: [
    {
      id: "pva-...-create",
      actionType: "CREATE",
      performedBy: userId,
      performedByName: "John Doe",
      performedByRole: "FINANCE_OFFICER",
      performedAt: now,
      newStatus: "DRAFT",
      metadata: {
        sourcePurchaseOrderId: po.id,
        sourcePurchaseOrderNumber: "PO-2024-001",
        autoCreated: false
      }
    }
  ],

  // PO Linking (Full Traceability)
  sourcePurchaseOrderId: po.id,
  sourcePurchaseOrderNumber: "PO-2024-001",
  createdFromPurchaseOrder: true,

  // Requisition Linking
  sourceRequisitionId: po.sourceRequisitionId,
  sourceRequisitionNumber: po.sourceRequisitionNumber,

  // Financial Metadata
  budgetCode: "BUDGET-001",
  costCenter: "CC-123",
  projectCode: "PROJ-456",

  // Timestamps
  createdAt: now,
  updatedAt: now
}
```

#### Step 5: Store in Mock Storage
```typescript
mockPaymentVouchers.push(paymentVoucher)
```

#### Step 6: Return Success Response
```typescript
return {
  success: true,
  data: paymentVoucher,
  message: 'Payment voucher created successfully'
}
```

---

## Client-Side: React Query Mutation

### Mutation Hook: useSavePaymentVoucher()

```typescript
return useMutation({
  mutationFn: async (data: CreatePaymentVoucherRequest) => {
    // Calls server action
    const response = await createPaymentVoucher(data)

    if (!response.success) {
      throw new Error(response.message)
    }
    return response
  },

  onSuccess: (response) => {
    // Show success toast
    toast.success("Payment voucher created successfully")

    // Invalidate queries to trigger refetch
    queryClient.invalidateQueries({
      queryKey: [QUERY_KEYS.PAYMENT_VOUCHERS.ALL]
    })
    queryClient.invalidateQueries({
      queryKey: [QUERY_KEYS.PAYMENT_VOUCHERS.STATS]
    })

    // Call parent callback
    onSuccess?.()  // → router.push('/payment-vouchers')
  },

  onError: (error: Error) => {
    toast.error(error.message || "Failed to save payment voucher")
  }
})
```

### Handler Integration
```typescript
const savePVMutation = useSavePaymentVoucher(() => {
  router.push('/payment-vouchers')  // ← Redirect on success
})

// Called when user clicks "Create"
await savePVMutation.mutateAsync(pvData)
```

---

## Complete Data Flow Diagram

```
┌──────────────────────────────────────────────────────────────────────────┐
│                    USER CLICKS "CREATE PAYMENT VOUCHER"                  │
└──────────────────────────────────────────────────────────────────────────┘
                                    ↓
                           handleCreatePV()
                                    ↓
                    ┌───────────────┴────────────────┐
                    │                                │
            Validation Check            Build CreatePaymentVoucherRequest
            (selectedPO exists?)                   ↓
                    │              Collect all necessary data
                    ├─ YES ↓
                    │  setIsCreating(true)
                    │  Button shows "Creating..."
                    │
                    ├─ NO ↓
                    │  toast.error("Please select...")
                    │  return (exit)
                    │
                    └─ Continue ↓
                          ↓
           savePVMutation.mutateAsync(pvData)
                          ↓
        ┌─────────────────────────────────────┐
        │    SERVER ACTION TRIGGERED          │
        │  createPaymentVoucher(pvData)       │
        └─────────────────────────────────────┘
                          ↓
           ┌──────────────────────────────────┐
           │  Generate Unique PV Number       │
           │  PV-2024-1001                    │
           └──────────────────────────────────┘
                          ↓
           ┌──────────────────────────────────┐
           │  Generate Unique PV ID           │
           │  pv-1733328400000-abc123xyz      │
           └──────────────────────────────────┘
                          ↓
           ┌──────────────────────────────────┐
           │  Initialize 3-Stage Approval     │
           │  Chain (all PENDING):            │
           │  1. Finance Manager              │
           │  2. Approval Authority           │
           │  3. Director                     │
           └──────────────────────────────────┘
                          ↓
           ┌──────────────────────────────────┐
           │  Map PO Items to PV Items        │
           │  Create line item records        │
           └──────────────────────────────────┘
                          ↓
           ┌──────────────────────────────────┐
           │  Create Action History Entry     │
           │  actionType: "CREATE"            │
           │  status: "DRAFT"                 │
           └──────────────────────────────────┘
                          ↓
           ┌──────────────────────────────────┐
           │  Store in mockPaymentVouchers    │
           └──────────────────────────────────┘
                          ↓
           ┌──────────────────────────────────┐
           │  Return Success Response         │
           │  {                               │
           │    success: true,                │
           │    data: paymentVoucher,         │
           │    message: "Created..."         │
           │  }                               │
           └──────────────────────────────────┘
                          ↓
        ┌─────────────────────────────────────┐
        │     CLIENT-SIDE MUTATION onSuccess  │
        └─────────────────────────────────────┘
                          ↓
           ┌──────────────────────────────────┐
           │  Show Toast: "Created success"   │
           │  setIsCreating(false)            │
           │  Button returns to normal        │
           └──────────────────────────────────┘
                          ↓
           ┌──────────────────────────────────┐
           │  Invalidate React Query Keys:    │
           │  - PAYMENT_VOUCHERS.ALL          │
           │  - PAYMENT_VOUCHERS.STATS        │
           │  (Triggers refetch from server)  │
           └──────────────────────────────────┘
                          ↓
           ┌──────────────────────────────────┐
           │  Call onSuccess Callback:        │
           │  router.push('/payment-vouchers')│
           └──────────────────────────────────┘
                          ↓
        ┌─────────────────────────────────────┐
        │     REDIRECT TO PAYMENT VOUCHERS    │
        │     LIST PAGE (/payment-vouchers)   │
        └─────────────────────────────────────┘
                          ↓
                    USER SEES
            Payment Vouchers List with
            Newly Created PV in DRAFT status
```

---

## Payment Voucher Created: Initial State

### Status & Workflow Position
```
Status: DRAFT
Current Approval Stage: 1
Next Action: SUBMIT for approval
```

### What Gets Created
```
PaymentVoucher {
  id: "pv-1733328400000-abc123xyz"
  pvNumber: "PV-2024-1001"
  status: "DRAFT"

  // Approval Chain Initialized (Not Active Yet)
  approvalChain: [
    {
      stageNumber: 1,
      stageName: "Finance Manager Review",
      status: "PENDING",  // Not started
      assignedTo: "finance-manager-1"
    },
    {
      stageNumber: 2,
      stageName: "Approval Authority Review",
      status: "PENDING",
      assignedTo: "approval-authority-1"
    },
    {
      stageNumber: 3,
      stageName: "Director Approval",
      status: "PENDING",
      assignedTo: "director-1"
    }
  ]

  // Action History Shows Creation
  actionHistory: [
    {
      actionType: "CREATE",
      performedBy: userId,
      performedAt: now,
      newStatus: "DRAFT"
    }
  ]

  // Full Traceability to Source PO
  sourcePurchaseOrderId: "po-456",
  sourcePurchaseOrderNumber: "PO-2024-001",
  sourceRequisitionId: "req-789",
  sourceRequisitionNumber: "REQ-2024-001"
}
```

---

## Payment Voucher Lifecycle Flow

```
┌─────────────┐
│   DRAFT     │  ← Created here
└──────┬──────┘
       │ User clicks "Submit for Approval"
       ↓
┌─────────────────┐
│   SUBMITTED     │  ← Approval chain begins
└──────┬──────────┘
       │ Finance Manager reviews & approves
       ↓
┌────────────────────┐
│   IN_REVIEW        │  ← Moved to stage 2
└──────┬─────────────┘
       │ Approval Authority reviews & approves
       ↓
┌────────────────────┐
│   IN_REVIEW        │  ← Moved to stage 3 (if still needed)
└──────┬─────────────┘
       │ Director approves (final stage)
       ↓
┌──────────────┐
│   APPROVED   │  ← All 3 stages approved
└──────┬───────┘
       │ Accounting marks payment made
       ↓
┌─────────────┐
│   PAID      │  ← Final status
└─────────────┘
```

---

## File Structure

```
src/app/(private)/(main)/payment-vouchers/
├── create/
│   ├── page.tsx                      [Server component - Auth + redirect]
│   └── _components/
│       └── pv-create-client.tsx      [Client component - Form UI]
│
└── [id]/
    ├── page.tsx
    ├── approval/
    │   └── page.tsx
    └── _components/
        └── pv-detail-client.tsx

src/app/_actions/
└── payment-vouchers.ts              [Server actions - createPaymentVoucher]

src/hooks/
└── use-payment-voucher-queries.ts   [React Query hooks]

src/types/
└── payment-voucher.ts               [Type definitions]
```

---

## Key Components & Responsibilities

### 1. CreatePaymentVoucherPage (Server Component)
- **Location**: `create/page.tsx`
- **Responsibility**: Authentication & initial server setup
- **Actions**:
  - Authenticates user via `auth()`
  - Redirects to login if not authenticated
  - Passes session data to client component

### 2. PVCreateClient (Client Component)
- **Location**: `create/_components/pv-create-client.tsx`
- **Responsibility**: UI rendering & form handling
- **State**:
  - `selectedPO`: Currently selected purchase order
  - `isCreating`: Button loading state
- **Data Fetching**:
  - `usePurchaseOrders()`: Fetch approved POs
  - `useSavePaymentVoucher()`: Create/update mutation
- **Handlers**:
  - `handleSelectPO()`: Update selected PO
  - `handleCreatePV()`: Build & submit data

### 3. createPaymentVoucher Server Action
- **Location**: `src/app/_actions/payment-vouchers.ts`
- **Responsibility**: Business logic & data persistence
- **Operations**:
  - Generate unique PV number
  - Generate unique PV ID
  - Initialize 3-stage approval chain
  - Map PO items to PV items
  - Create action history entry
  - Store in mock storage (or database)

### 4. useSavePaymentVoucher React Query Hook
- **Location**: `src/hooks/use-payment-voucher-queries.ts`
- **Responsibility**: Manage mutation state & UI updates
- **Operations**:
  - Call server action with data
  - Handle success (show toast, invalidate queries, redirect)
  - Handle errors (show error toast, keep on page)

---

## Error Handling

### Client-Side Validation
```typescript
// Check PO is selected
if (!selectedPO) {
  toast.error('Please select a purchase order')
  return
}
```

### Server-Side Validation
```typescript
// Verify no critical fields are missing
if (!data.vendorName || !data.items) {
  return { success: false, message: '...' }
}
```

### Error Recovery
- Toast notification shown to user
- Form remains open for corrections
- No redirect on error (user stays on create page)
- User can select different PO or try again

---

## Data Mapping: PO → PV

### Fields Auto-Populated from PO
```
PO Field              →  PV Field
───────────────────────────────────
vendorId              →  vendorId
vendorName            →  vendorName
department            →  department
departmentId          →  departmentId
budgetCode            →  budgetCode
costCenter            →  costCenter
projectCode           →  projectCode
items[]               →  items[]
totalAmount           →  totalAmount
currency              →  currency
sourceRequisitionId   →  sourceRequisitionId
sourceRequisitionNumber → sourceRequisitionNumber
```

### Fields Generated for PV
```
Field                   Value                  Purpose
──────────────────────────────────────────────────────
pvNumber               PV-2024-1001            Unique identifier
status                 "DRAFT"                 Initial workflow state
priority               "MEDIUM"                Default priority
paymentMethod          "BANK_TRANSFER"         Default payment method
paymentDueDate         +30 days                Default due date
approvalChain          [3 stages]              Approval workflow
currentApprovalStage   1                       First stage
actionHistory          [CREATE entry]          Audit trail
```

---

## Performance Characteristics

| Aspect | Details |
|--------|---------|
| **Data Fetching** | Approved POs loaded async via useQuery |
| **Rendering** | Only APPROVED purchase orders shown |
| **Filtering** | Client-side filter: `status === 'APPROVED'` |
| **Mutation Time** | Instant (mock storage, no API latency) |
| **Cache Invalidation** | React Query refetch after creation |
| **UI Feedback** | Loading state, success/error toasts |

---

## Security Considerations

### Authentication
- ✅ User must be authenticated (server-side check)
- ✅ Session validated before accessing page

### Authorization
- ✅ User role captured for action history
- ✅ All actions traceable to user

### Data Validation
- ✅ Required fields validated on client & server
- ✅ PO filtering ensures only approved POs used
- ✅ Error messages don't expose sensitive data

### Audit Trail
- ✅ Creation action logged with user info
- ✅ All future actions tracked in actionHistory
- ✅ Timestamps recorded at every step

---

## Testing Scenarios

### Happy Path
1. User navigates to `/payment-vouchers/create` ✓
2. Approved POs load in dropdown ✓
3. User selects a PO ✓
4. PO details display ✓
5. User clicks "Create Payment Voucher" ✓
6. Payment voucher created in DRAFT status ✓
7. User redirected to `/payment-vouchers` ✓
8. New PV visible in list ✓

### Error Scenarios
1. No approved POs available → Destructive alert shown
2. No PO selected → Error toast, form stays open
3. Network error during create → Error toast, user can retry
4. Server validation fails → Error toast, user can modify or cancel

### Edge Cases
1. Multiple users creating PVs simultaneously → Each gets unique pvNumber
2. Creating PV from PO with many items → All items mapped correctly
3. PO with special characters in description → Properly escaped and stored
4. Creating PV from PO created long ago → Requisition link preserved

---

## Summary: What Happens When Button is Clicked

```
User Clicks "Create Payment Voucher"
        ↓
[Client] Validate PO selected
        ↓
[Client] Build CreatePaymentVoucherRequest
        ↓
[Client] Call useSavePaymentVoucher mutation
        ↓
[Server] createPaymentVoucher() executes
  ├─ Generate PV-2024-1001
  ├─ Generate unique pvId
  ├─ Initialize 3-stage approval chain
  ├─ Map PO items → PV items
  ├─ Create action history
  ├─ Store in mock storage
  └─ Return success response
        ↓
[Client] onSuccess callback fires
  ├─ Toast: "Payment voucher created successfully"
  ├─ Invalidate React Query cache
  └─ Redirect to /payment-vouchers
        ↓
User Sees Payment Vouchers List
  └─ New PV appears with DRAFT status
        ↓
User Can Now:
  • Edit the PV (while DRAFT)
  • Delete the PV (while DRAFT)
  • Submit for approval (move to SUBMITTED)
```

---

## Code References

| File | Lines | Purpose |
|------|-------|---------|
| [create/page.tsx](src/app/(private)/(main)/payment-vouchers/create/page.tsx) | 1-24 | Server component |
| [pv-create-client.tsx](src/app/(private)/(main)/payment-vouchers/create/_components/pv-create-client.tsx) | 1-308 | Client form component |
| [payment-vouchers.ts (createPaymentVoucher)](src/app/_actions/payment-vouchers.ts#L163-L261) | 163-261 | Server action |
| [payment-voucher-queries.ts (useSavePaymentVoucher)](src/hooks/use-payment-voucher-queries.ts#L122-L165) | 122-165 | React Query mutation |
| [payment-voucher.ts (types)](src/types/payment-voucher.ts#L148-L178) | 148-178 | CreatePaymentVoucherRequest type |
