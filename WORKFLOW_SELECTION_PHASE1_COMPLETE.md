# Workflow Selection - Phase 1 Complete ✅

## Summary

Phase 1 (Update Types & Actions) has been successfully completed. All document types now have `workflowId` in their submit request types, and all submit actions send `workflowId` to the backend.

## Changes Made

### 1. Updated Submit Request Types

Added `workflowId: string` (REQUIRED) to all submit request interfaces:

#### ✅ Budget (`frontend/src/types/budget.ts`)

```typescript
export interface SubmitBudgetRequest {
  budgetId: string;
  workflowId: string; // REQUIRED - Workflow to use for approval
  submittedBy: string;
  submittedByRole: string;
  submittingUserId?: string;
  comments?: string;
}
```

#### ✅ Requisition (`frontend/src/types/requisition.ts`)

```typescript
export interface SubmitRequisitionRequest {
  requisitionId: string;
  workflowId: string; // REQUIRED - Workflow to use for approval
  submittedBy: string;
  submittedByName: string;
  submittedByRole: string;
  comments?: string;
}
```

#### ✅ Purchase Order (`frontend/src/types/purchase-order.ts`)

```typescript
export interface SubmitPurchaseOrderRequest {
  purchaseOrderId: string;
  poId?: string;
  workflowId: string; // REQUIRED - Workflow to use for approval
  submittingUserId: string;
  submittedBy?: string;
  submittedByName: string;
  submittedByRole: string;
  comments?: string;
}
```

#### ✅ Payment Voucher (`frontend/src/types/payment-voucher.ts`)

```typescript
export interface SubmitPaymentVoucherRequest {
  paymentVoucherId: string;
  pvId?: string;
  workflowId: string; // REQUIRED - Workflow to use for approval
  submittingUserId: string;
  submittedBy?: string;
  submittedByName: string;
  submittedByRole: string;
  comments?: string;
}
```

#### ✅ GRN (`frontend/src/types/goods-received-note.ts`)

```typescript
export interface SubmitGRNRequest {
  grnId: string;
  workflowId: string; // REQUIRED - Workflow to use for approval
  submittedBy: string;
  submittedByName: string;
  submittedByRole: string;
  comments?: string;
}
```

Also added missing `ApproveGRNRequest` and `RejectGRNRequest` interfaces.

### 2. Updated Submit Actions

Updated all submit actions to send `workflowId` to the backend:

#### ✅ Budget (`frontend/src/app/_actions/budgets.ts`)

```typescript
export async function submitBudgetForApproval(
  request: SubmitBudgetRequest,
): Promise<APIResponse<Budget | null>> {
  const response = await authenticatedApiClient({
    method: "POST",
    url: `/api/v1/budgets/${request.budgetId}/submit`,
    data: {
      workflowId: request.workflowId, // REQUIRED by backend
      submittingUserId: request.submittingUserId,
      comments: request.comments,
    },
  });
  // ...
}
```

#### ✅ Requisition (`frontend/src/app/_actions/requisitions.ts`)

```typescript
export async function submitRequisitionForApproval(
  data: SubmitRequisitionRequest,
): Promise<APIResponse<Requisition>> {
  const response = await authenticatedApiClient({
    method: "POST",
    url: `/api/v1/requisitions/${data.requisitionId}/submit`,
    data: {
      workflowId: data.workflowId, // REQUIRED by backend
      comments: data.comments,
      submittedBy: data.submittedBy,
      submittedByName: data.submittedByName,
      submittedByRole: data.submittedByRole,
    },
  });
  // ...
}
```

#### ✅ Purchase Order (`frontend/src/app/_actions/purchase-orders.ts`)

```typescript
export async function submitPurchaseOrderForApproval(
  data: SubmitPurchaseOrderRequest,
): Promise<APIResponse<PurchaseOrder>> {
  const response = await authenticatedApiClient({
    method: "POST",
    url: `/api/v1/purchase-orders/${data.poId}/submit`,
    data: {
      workflowId: data.workflowId, // REQUIRED by backend
      comments: data.comments,
      submittedBy: data.submittedBy,
      submittedByName: data.submittedByName,
      submittedByRole: data.submittedByRole,
    },
  });
  // ...
}
```

#### ✅ Payment Voucher (`frontend/src/app/_actions/payment-vouchers.ts`)

```typescript
export async function submitPaymentVoucherForApproval(
  data: SubmitPaymentVoucherRequest,
): Promise<APIResponse<PaymentVoucher>> {
  const response = await authenticatedApiClient({
    method: "POST",
    url: `/api/v1/payment-vouchers/${data.pvId}/submit`,
    data: {
      workflowId: data.workflowId, // REQUIRED by backend
      pvId: data.pvId,
      submittedBy: data.submittedBy,
      submittedByName: data.submittedByName,
      submittedByRole: data.submittedByRole,
      comments: data.comments,
    },
  });
  // ...
}
```

#### ✅ GRN (`frontend/src/app/_actions/grn-actions.ts`)

```typescript
export async function submitGRNForApproval(data: {
  grnId: string;
  workflowId: string;
  submittedBy: string;
  submittedByName: string;
  submittedByRole: string;
  comments?: string;
}): Promise<APIResponse<GoodsReceivedNote>> {
  const response = await authenticatedApiClient({
    method: "POST",
    url: `/api/v1/grns/${data.grnId}/submit`,
    data: {
      workflowId: data.workflowId, // REQUIRED by backend
      submittedBy: data.submittedBy,
      submittedByName: data.submittedByName,
      submittedByRole: data.submittedByRole,
      comments: data.comments,
    },
  });
  // ...
}
```

## Validation

All files passed TypeScript diagnostics with no errors:

- ✅ `frontend/src/types/budget.ts`
- ✅ `frontend/src/types/requisition.ts`
- ✅ `frontend/src/types/purchase-order.ts`
- ✅ `frontend/src/types/payment-voucher.ts`
- ✅ `frontend/src/types/goods-received-note.ts`
- ✅ `frontend/src/app/_actions/budgets.ts`
- ✅ `frontend/src/app/_actions/requisitions.ts`
- ✅ `frontend/src/app/_actions/purchase-orders.ts`
- ✅ `frontend/src/app/_actions/payment-vouchers.ts`
- ✅ `frontend/src/app/_actions/grn-actions.ts`

## Impact

### Breaking Changes ⚠️

These changes introduce breaking changes to the submit functions. Any code currently calling these functions will need to be updated to include `workflowId`.

### Files That Will Need Updates

The following files will need to be updated in subsequent phases:

- All submit dialog components
- All document detail pages
- All submit mutation hooks

## Next Steps

### Phase 2: Create Workflow Selector Component

- Create `WorkflowSelector` component
- Create `WorkflowDetails` component (optional)
- Test with different entity types

### Phase 3: Update Submit Dialogs

- Update Budget submit dialog (already exists)
- Create/Update Requisition submit dialog
- Create/Update Purchase Order submit dialog
- Create/Update Payment Voucher submit dialog
- Create/Update GRN submit dialog

### Phase 4: Update Detail Pages

- Update all document detail pages to pass workflowId

### Phase 5: Update Hooks

- Update submit mutation hooks to accept workflowId

## Notes

1. The backend is ready and requires `workflowId` in all submit requests
2. The workflow hooks (`use-workflow-queries.ts`) are already implemented
3. All submit actions now send `workflowId` to match backend expectations
4. GRN submit action was created (didn't exist before)
5. GRN approve/reject request types were added (were missing)

## Time Taken

Phase 1: ~1.5 hours (as estimated)

## Status

✅ **COMPLETE** - Ready to proceed to Phase 2
