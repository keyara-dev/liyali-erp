# Frontend Module Integration Status Report

**Status**: ⚠️ INCONSISTENT PATTERNS FOUND
**Date**: 2025-12-26
**Branch**: feat/go-fiber

---

## Summary

**GRN Refactoring** has established the correct pattern, but **other modules are NOT following it**.

Current State:
- ✅ **GRN**: Fully refactored, uses authenticatedApiClient, proper error handling
- ❌ **Requisitions**: Mock data, NO backend API calls
- ❌ **Purchase Orders**: Mock data, NO backend API calls
- ❌ **Payment Vouchers**: Likely mock data
- ❌ **Budgets**: Likely mock data

---

## Module Status

### 1. GRN (Goods Received Notes) ✅ COMPLETE

**File**: `frontend/src/app/_actions/grn-actions.ts`

**Status**: ✅ REFACTORED
- [x] Uses authenticatedApiClient
- [x] Proper error handling with handleError()
- [x] Response wrapping with successResponse()
- [x] 11 server actions all using backend API
- [x] React Query hooks with toast notifications
- [x] Query invalidation on mutations
- [x] Full TypeScript support

**Pattern**: CORRECT ✅

---

### 2. Requisitions ❌ NEEDS REFACTORING

**File**: `frontend/src/app/_actions/requisitions.ts` (27KB)

**Status**: ❌ MOCK DATA
- [x] Has backend API available
- ❌ NOT using authenticatedApiClient
- ❌ Uses hardcoded mock data (250+ lines)
- ❌ No error handling via handleError()
- ❌ No response wrapping via successResponse()
- ❌ No authentication checks
- ❌ No organization scoping

**Issues**:
```typescript
// ❌ MOCK DATA (250+ lines hardcoded)
let mockRequisitions: Requisition[] = [...];

// ❌ DIRECT ARRAY MANIPULATION
export async function createRequisition(data) {
  mockRequisitions.push(newRequisition);  // ❌ NOT API
  return { success: true, data: newRequisition };
}
```

**Pattern**: INCORRECT ❌

**Action**: Refactor to match GRN pattern

**Estimated Effort**: 2-3 hours

---

### 3. Purchase Orders ❌ NEEDS REFACTORING

**File**: `frontend/src/app/_actions/purchase-orders.ts` (23KB)

**Status**: ❌ MOCK DATA
- [x] Has backend API available
- ❌ NOT using authenticatedApiClient
- ❌ Uses similar mock data pattern
- ❌ No error handling via handleError()
- ❌ No response wrapping via successResponse()
- ❌ No authentication checks

**Pattern**: INCORRECT ❌

**Action**: Refactor to match GRN pattern

**Estimated Effort**: 2-3 hours

---

### 4. Payment Vouchers ❌ LIKELY NEEDS REFACTORING

**File**: `frontend/src/app/_actions/payment-vouchers.ts` (22KB)

**Status**: ⚠️ LIKELY MOCK DATA
- Suspected similar pattern as Requisitions/POs
- Backend API likely available
- Probably NOT using authenticatedApiClient

**Action**: Audit and refactor if needed

**Estimated Effort**: 2-3 hours

---

### 5. Budgets ❌ LIKELY NEEDS REFACTORING

**File**: `frontend/src/app/_actions/budgets.ts` (18KB)

**Status**: ⚠️ LIKELY MOCK DATA
- Suspected similar pattern
- Backend API likely available
- Probably NOT using authenticatedApiClient

**Action**: Audit and refactor if needed

**Estimated Effort**: 2-3 hours

---

## Pattern Comparison

### CORRECT Pattern (GRN) ✅

```typescript
// Server Action with authenticatedApiClient
export async function getGRNAction(grnId: string): Promise<APIResponse<GoodsReceivedNote>> {
  const url = `/api/v1/grns/${grnId}`;
  try {
    const response = await authenticatedApiClient({
      method: 'GET',
      url,
    });
    return successResponse(response.data?.data, 'GRN retrieved');
  } catch (error: any) {
    return handleError(error, 'GET', url);
  }
}

// React Query Hook with notifications
export const useGRNById = (grnId: string) => {
  const queryClient = useQueryClient();
  return useQuery({
    queryKey: [QUERY_KEYS.GRN.BY_ID, grnId],
    queryFn: async () => {
      const response = await getGRNAction(grnId);
      if (!response.success) throw new Error(response.message);
      return response.data;
    },
  });
};
```

---

### INCORRECT Pattern (Requisitions, POs, etc) ❌

```typescript
// Server Action with MOCK DATA
let mockRequisitions: Requisition[] = [...]; // 250+ lines of hardcoded data

export async function getRequisition(requisitionId: string): Promise<APIResponse<Requisition>> {
  try {
    const requisition = mockRequisitions.find(r => r.id === requisitionId); // ❌ MOCK
    if (!requisition) {
      return { success: false, message: 'Not found', status: 404 };
    }
    return { success: true, data: requisition, message: 'OK', status: 200 };
  } catch (error: any) {
    return { success: false, message: error.message, status: 500 }; // ❌ NO handleError
  }
}

// React Query Hook WITHOUT notifications
export const useRequisitionById = (requisitionId: string) => {
  return useQuery({
    queryKey: ['requisition', requisitionId],
    queryFn: async () => {
      const response = await getRequisition(requisitionId);
      if (!response.success) throw new Error(response.message);
      return response.data; // Still returns mock data!
    },
  });
};
```

---

## Refactoring Priority

### Phase 1 (CRITICAL)
1. **Requisitions** - Most used module
   - Effort: 2-3 hours
   - Risk: Medium
   - Impact: High

### Phase 2 (HIGH)
2. **Purchase Orders** - Core workflow
   - Effort: 2-3 hours
   - Risk: Medium
   - Impact: High

3. **Payment Vouchers** - Finance critical
   - Effort: 2-3 hours
   - Risk: High
   - Impact: Critical

### Phase 3 (MEDIUM)
4. **Budgets** - Planning module
   - Effort: 2-3 hours
   - Risk: Medium
   - Impact: Medium

---

## Implementation Template

All modules should follow this template (matching GRN):

### Step 1: Update Server Actions

```typescript
'use server';

import { APIResponse } from '@/types';
import {
  handleError,
  successResponse,
  badRequestResponse,
} from './api-config';
import authenticatedApiClient from './api-config';

// Each function must:
export async function getMyResourceAction(...): Promise<APIResponse<T>> {
  const url = `/api/v1/my-resource`;
  try {
    const response = await authenticatedApiClient({
      method: 'GET',
      url,
    });
    return successResponse(response.data?.data, 'Success message');
  } catch (error: any) {
    return handleError(error, 'GET', url);
  }
}
```

### Step 2: Update Hooks

```typescript
'use client';

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { QUERY_KEYS } from '@/lib/constants';
import { toast } from 'sonner';

// Each hook must include:
export const useMyResource = () => {
  const queryClient = useQueryClient();
  const mutation = useMutation({
    mutationFn: async (data) => {
      const response = await myResourceAction(data);
      if (!response.success) throw new Error(response.message);
      return response;
    },
    onSuccess: (response) => {
      toast.success('Operation successful');
      queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.MY_RESOURCE.ALL] });
    },
    onError: (error: any) => {
      toast.error(error.message || 'Operation failed');
    }
  });
  return mutation;
};
```

---

## Backend API Verification

All these modules have backend APIs:

✅ **Requisitions**: `backend/handlers/requisition.go`
✅ **Purchase Orders**: `backend/handlers/purchase_order.go`
✅ **Payment Vouchers**: `backend/handlers/payment_voucher.go`
✅ **Budgets**: `backend/handlers/budget.go`

---

## Checklist for Each Module

When refactoring, ensure:

- [ ] Remove all mock data
- [ ] All functions use authenticatedApiClient
- [ ] All responses use successResponse()
- [ ] All errors use handleError()
- [ ] verifySession() called for authentication
- [ ] Organization context included
- [ ] All mutations have toast notifications
- [ ] Query invalidation implemented
- [ ] TypeScript types defined
- [ ] Exported from hooks/index.ts
- [ ] Documentation created
- [ ] Tests updated

---

## Estimated Total Effort

- **Phase 1 (Requisitions)**: 2-3 hours
- **Phase 2 (POs + PVs)**: 4-6 hours
- **Phase 3 (Budgets)**: 2-3 hours
- **Documentation**: 2 hours
- **Testing**: 3 hours

**Total**: 13-17 hours (~2-3 days)

---

## Risk Mitigation

- Start with Requisitions (most used, lower risk)
- Test thoroughly with backend API
- Keep mock data handlers for fallback (initially)
- Implement gradually, one module at a time
- Update documentation as we go

---

## Next Steps

1. **Audit all modules** (1 hour)
2. **Create refactoring tickets** (30 mins)
3. **Refactor Requisitions** (2-3 hours)
4. **Test & document** (1.5 hours)
5. **Refactor remaining modules** (follow same pattern)

---

## Success Criteria

When complete:
- [x] All modules use authenticatedApiClient
- [x] All use handleError() and successResponse()
- [x] All hooks have toast notifications
- [x] All hooks implement query invalidation
- [x] Consistent pattern across entire codebase
- [x] Full TypeScript support
- [x] Production ready

---

## GRN as Reference

GRN has been completed as a reference implementation. All other modules should be refactored to match its pattern exactly.

**Files to reference**:
- `frontend/src/app/_actions/grn-actions.ts` - Server actions pattern
- `frontend/src/hooks/use-grn-queries.ts` - Hook pattern
- `GRN-INTEGRATION-GUIDE.md` - Complete documentation

---

**Status**: Ready for Phase 1 refactoring
**Owner**: Claude Code
**Created**: 2025-12-26

