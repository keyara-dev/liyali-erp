# Requisition-to-Purchase Order Integration Guide

## Overview

This document outlines how the Requisition workflow integrates with and creates Purchase Orders (POs). The relationship follows this flow:

```
REQUISITION (Approved) → CREATE PO → PURCHASE ORDER (Draft) → PO Approval Workflow
```

## Document Structure

### Requisition Workflow
- **Type**: `Requisition` (internal document type)
- **Status**: DRAFT → SUBMITTED → IN_REVIEW → APPROVED → (triggers PO creation)
- **Approvals**: 3-stage approval chain (Department Manager → Finance Officer → Director)
- **Data Storage**: Mock in-memory array + localStorage persistence
- **Action Tracking**: Complete audit trail with timestamps and user info

### Purchase Order Workflow
- **Type**: `PurchaseOrder` (extends `WorkflowDocument`)
- **Status**: DRAFT → SUBMITTED → IN_REVIEW → APPROVED
- **Approvals**: 4-stage approval chain
- **Data Storage**: Currently uses `WorkflowDocument` store (in-memory maps)
- **Connection Point**: References the source Requisition ID

## Relationship Architecture

### 1. **Requisition Approval Complete**
When a Requisition reaches APPROVED status:
```typescript
// From: src/app/_actions/requisitions.ts
if (allApproved) {
  requisition.status = 'APPROVED';
  requisition.approvedAt = new Date();

  // Log approval action with full details
  requisition.actionHistory.push({
    actionType: 'APPROVE',
    stageNumber: 3,
    stageName: 'Director',
    // ... complete user and signature tracking
  });

  // NEXT STEP: Create PO from this approved requisition
  // await createPurchaseOrderFromRequisition(requisition);
}
```

### 2. **Purchase Order Creation from Requisition**
When a Requisition is APPROVED, a Purchase Order should be created with:

**Source Data Mapping:**
```typescript
Interface PurchaseOrder Linking {
  // Link back to source requisition
  sourceRequisitionId: string;           // req-123
  sourceRequisitionNumber: string;       // REQ-2024-001

  // Copy from requisition
  vendorId: string;                      // From requisition metadata or user selection
  vendorName: string;                    // Extracted or user-selected
  items: PurchaseOrderItem[];            // From requisition.items
  totalAmount: number;                   // From requisition.totalAmount
  currency: string;                      // From requisition.currency (ZMW)

  // New metadata
  poNumber: string;                      // Auto-generated: PO-2024-001
  status: 'DRAFT';                       // Always starts as DRAFT
  createdFromRequisition: boolean;       // true - indicates PO source

  // Approval chain (4 stages for PO)
  approvalChain: [
    { stageNumber: 1, stageName: 'Procurement Manager', ... },
    { stageNumber: 2, stageName: 'Finance Manager', ... },
    { stageNumber: 3, stageName: 'Vendor Compliance', ... },
    { stageNumber: 4, stageName: 'Director', ... }
  ];
}
```

### 3. **Linking & Traceability**

Users should be able to:
1. **See Source**: Open a PO and view the original requisition details
2. **Track Approvals**: See approval history flowing from requisition through PO
3. **Audit Trail**: View all actions from requisition creation → approval → PO creation → PO approvals

### 4. **Data Flow Diagram**

```
┌─────────────────────────────────────────────────────────────────┐
│                    REQUISITION LIFECYCLE                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  1. CREATE (User)                                              │
│     ├─ actionHistory.push({type: 'CREATE'})                   │
│     └─ localStorage: liyali_requisitions                       │
│                                                                 │
│  2. SUBMIT (User)                                              │
│     ├─ status: DRAFT → SUBMITTED                              │
│     ├─ currentApprovalStage: 0 → 1                            │
│     ├─ actionHistory.push({type: 'SUBMIT'})                   │
│     └─ localStorage: updated requisition                       │
│                                                                 │
│  3. APPROVE Stage 1 (Dept Manager)                            │
│     ├─ approvalChain[0].status = 'APPROVED'                   │
│     ├─ status: SUBMITTED → IN_REVIEW                          │
│     ├─ currentApprovalStage: 1 → 2                            │
│     ├─ actionHistory.push({type: 'APPROVE', stageNumber: 1})  │
│     └─ localStorage: updated with action                       │
│                                                                 │
│  4. APPROVE Stage 2 (Finance Officer)                         │
│     ├─ approvalChain[1].status = 'APPROVED'                   │
│     ├─ status: IN_REVIEW → IN_REVIEW                          │
│     ├─ currentApprovalStage: 2 → 3                            │
│     ├─ actionHistory.push({type: 'APPROVE', stageNumber: 2})  │
│     └─ localStorage: updated with action                       │
│                                                                 │
│  5. APPROVE Stage 3 (Director) ← ALL APPROVED                 │
│     ├─ approvalChain[2].status = 'APPROVED'                   │
│     ├─ status: IN_REVIEW → APPROVED                           │
│     ├─ approvedAt: new Date()                                 │
│     ├─ actionHistory.push({type: 'APPROVE', stageNumber: 3})  │
│     ├─ localStorage: final requisition saved                   │
│     └─ **TRIGGER: createPurchaseOrderFromRequisition()**      │
│                                                                 │
└────────────────────────────────────────────────────────────────┘
                            ↓
┌────────────────────────────────────────────────────────────────┐
│                    PURCHASE ORDER CREATED                      │
├────────────────────────────────────────────────────────────────┤
│                                                                │
│  PO Details:                                                  │
│  ├─ poNumber: 'PO-2024-001'                                  │
│  ├─ status: 'DRAFT'                                          │
│  ├─ sourceRequisitionId: 'req-123'                           │
│  ├─ sourceRequisitionNumber: 'REQ-2024-001'                  │
│  ├─ items: [...from requisition.items]                       │
│  ├─ totalAmount: requisition.totalAmount                     │
│  ├─ actionHistory: [                                         │
│  │    { type: 'CREATE', fromRequisition: true, ... }         │
│  │  ]                                                        │
│  └─ approvalChain: 4-stage PO approval workflow              │
│                                                               │
│  Storage: WorkflowDocument store                            │
│  (via documentStore.set(po.id, po))                         │
│                                                               │
└────────────────────────────────────────────────────────────────┘
                            ↓
┌────────────────────────────────────────────────────────────────┐
│           PURCHASE ORDER APPROVAL WORKFLOW                    │
├────────────────────────────────────────────────────────────────┤
│                                                                │
│  Stage 1: Procurement Manager → APPROVED                      │
│  Stage 2: Finance Manager → APPROVED                          │
│  Stage 3: Vendor Compliance → APPROVED                        │
│  Stage 4: Director → APPROVED                                 │
│                                                                │
│  Final Status: PO → APPROVED                                  │
│                                                                │
└────────────────────────────────────────────────────────────────┘
```

## Implementation Checklist

### Phase 1: Type System Enhancements
- [ ] Extend `PurchaseOrder` type to include:
  - `sourceRequisitionId`: string
  - `sourceRequisitionNumber`: string
  - `createdFromRequisition`: boolean

- [ ] Add to `ActionHistoryEntry`:
  - `sourceDocumentId`: string (optional, for linking actions)
  - `sourceDocumentType`: 'REQUISITION' | 'PO' | etc.

### Phase 2: PO Creation Logic
- [ ] Create `createPurchaseOrderFromRequisition()` server action
  - Input: Approved Requisition
  - Process: Convert items, create approval chain, set links
  - Output: New PurchaseOrder
  - Trigger: Called when requisition reaches APPROVED status

- [ ] Add to `src/app/_actions/purchase-orders.ts`:
  ```typescript
  export async function createPurchaseOrderFromRequisition(
    requisition: Requisition
  ): Promise<APIResponse<PurchaseOrder>>
  ```

### Phase 3: Linking & UI
- [ ] Update PO detail view to show:
  - "Created from Requisition" section
  - Link to source requisition
  - Original requisition approval history

- [ ] Update Requisition detail view to show:
  - "Generated Purchase Orders" section
  - Links to all POs created from this requisition
  - PO approval status

### Phase 4: LocalStorage for PO
- [ ] Create `src/hooks/use-purchase-order-storage.ts`:
  - Similar to requisition storage
  - PO-specific persistence
  - Action history tracking

- [ ] Create `src/types/purchase-order.ts`:
  - Define PurchaseOrder standalone type (not just extension)
  - Include all linking fields
  - Include action history

### Phase 5: Approval Workflow Alignment
- [ ] Create `src/hooks/use-purchase-order-queries.ts`:
  - `useApprovePO()`
  - `useRejectPO()`
  - `useSubmitPOForApproval()`
  - Same structure as requisition hooks

- [ ] Implement action tracking:
  - APPROVE action with stage info
  - REJECT action with remarks
  - Link back to requisition

### Phase 6: Cascade Effects
- [ ] When PO is approved:
  - Update requisition to show "PO Approved"
  - Display PO status in requisition detail

- [ ] When PO is rejected:
  - Update requisition to show "PO Rejected"
  - Allow regeneration or manual adjustment

## Data Persistence Strategy

### Requisitions Storage
```javascript
localStorage {
  'liyali_requisitions': [
    {
      id: 'req-123',
      requisitionNumber: 'REQ-2024-001',
      status: 'APPROVED',
      actionHistory: [...],        // Complete audit trail
      approvalChain: [...],        // All stage approvals
      relatedPurchaseOrders: [     // NEW: Link to POs
        'po-456',
        'po-789'
      ]
    }
  ]
}
```

### Purchase Orders Storage
```javascript
// In WorkflowDocument store (maps)
documentStore.set('po-456', {
  id: 'po-456',
  poNumber: 'PO-2024-001',
  status: 'DRAFT',
  sourceRequisitionId: 'req-123',        // Link back to requisition
  sourceRequisitionNumber: 'REQ-2024-001',
  items: [...],
  actionHistory: [
    {
      type: 'CREATE',
      sourceDocument: 'REQUISITION',
      sourceDocumentId: 'req-123',
      ...
    }
  ],
  approvalChain: [
    { stageNumber: 1, stageName: 'Procurement Manager', ... },
    ...
  ]
});
```

## API Response Structure

### When Requisition is Approved
```typescript
{
  success: true,
  message: 'Requisition approved and Purchase Order created',
  data: {
    requisition: { /* updated requisition */ },
    purchaseOrder: { /* newly created PO */ },
    createdPONumber: 'PO-2024-001'
  }
}
```

## Business Logic Flow

### Automatic PO Creation Trigger
```typescript
// In approveRequisition()
if (allApproved) {
  requisition.status = 'APPROVED';
  requisition.approvedAt = new Date();

  // Log the approval
  requisition.actionHistory.push({...});

  // CRITICAL: Create PO from approved requisition
  const poResult = await createPurchaseOrderFromRequisition(requisition);

  if (poResult.success) {
    // Link PO back to requisition
    requisition.relatedPurchaseOrders = [
      ...(requisition.relatedPurchaseOrders || []),
      poResult.data.id
    ];

    // Save updated requisition
    mockRequisitions[index] = requisition;
  }
}
```

## Key Design Principles

1. **Single Source of Truth**: Requisition is the authoritative source for item/amount data
2. **Complete Audit Trail**: Every action tracked from requisition creation through PO approval
3. **Transparent Linking**: Users can navigate between requisition and related POs
4. **Approval Separation**: Requisition and PO have independent approval workflows
5. **Persistent Storage**: Both use similar localStorage strategies for offline support
6. **Action Tracking**: Both maintain comprehensive action history

## Status Matrix

| Requisition Status | PO Status  | What Happens |
|-------------------|-----------|--------------|
| DRAFT | N/A | No PO yet |
| SUBMITTED | N/A | No PO yet |
| IN_REVIEW | N/A | No PO yet |
| APPROVED | DRAFT | PO auto-created |
| APPROVED | IN_REVIEW | Requisition complete, PO in progress |
| APPROVED | APPROVED | Entire procurement chain complete |
| REJECTED | N/A | Requisition goes back to draft |

## Testing Scenarios

### Scenario 1: Complete Requisition → PO Flow
1. Create requisition with 3 items
2. Submit for approval
3. Approve at stage 1 (Dept Manager)
4. Approve at stage 2 (Finance Officer)
5. Approve at stage 3 (Director)
6. **Verify**: PO automatically created
7. **Verify**: PO has same items, total amount, requisition links
8. **Verify**: Action history shows creation from requisition

### Scenario 2: PO Approval Workflow
1. Complete requisition approval (creates PO)
2. Submit PO for approval
3. Approve through all 4 PO stages
4. **Verify**: Status changes at each stage
5. **Verify**: Action history tracks each approval
6. **Verify**: Requisition shows linked PO status

### Scenario 3: PO Rejection & Regeneration
1. Complete requisition approval (creates PO)
2. Submit PO for approval
3. Reject at stage 2 (Finance Manager)
4. **Verify**: PO goes back to DRAFT
5. **Verify**: Requisition still shows as APPROVED
6. **Option**: Create new PO with adjusted details

## Implementation Timeline

1. **Phase 1**: Type extensions (1-2 hours)
2. **Phase 2**: PO creation logic (2-3 hours)
3. **Phase 3**: UI linking & detail views (3-4 hours)
4. **Phase 4**: PO storage & hooks (2-3 hours)
5. **Phase 5**: Approval workflows (2-3 hours)
6. **Phase 6**: Testing & refinement (2-3 hours)

**Total**: ~12-18 hours of implementation

## Future Enhancements

1. **Batch PO Creation**: Create multiple POs from single requisition
2. **Partial PO**: PO for subset of requisition items
3. **Budget Tracking**: Link requisition amount to budget
4. **Vendor Management**: Track vendor selection process
5. **Receipt Matching**: Match receipts to PO items
6. **Payment Integration**: Link PO approval to payment processing
