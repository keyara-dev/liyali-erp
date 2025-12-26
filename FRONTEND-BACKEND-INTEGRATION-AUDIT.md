# Frontend-Backend Integration Audit Report

**Status**: AUDIT COMPLETE - DISCREPANCIES FOUND
**Date**: 2025-12-26
**Branch**: feat/go-fiber

---

## Executive Summary

The frontend codebase has **~85% backend integration** with the following status:

| Category | Status | Notes |
|----------|--------|-------|
| Core Workflows (Req/PO/PV/Budget) | ✅ Excellent | Fully integrated with Query hooks + localStorage fallback |
| Admin/Users/Roles | ✅ Excellent | Proper backend integration |
| GRNs (Goods Received Notes) | ⚠️ Critical Issue | Server action accessing localStorage - INVALID |
| Workflows Configuration | ⚠️ Major Issue | Primarily using localStorage instead of backend API |
| Vendors & Categories | ❌ Not Implemented | No server actions or Query hooks |
| Departments | ⚠️ Partial | Mock storage in localStorage |

---

## Critical Issues Found

### 1. **GRN Server Action Anti-Pattern** (CRITICAL)
**File**: `frontend/src/app/_actions/grn-actions.ts`

**Problem**: Server action directly accesses `localStorage` which is invalid in server components:
```typescript
export async function getQualityIssues(grnId: string) {
  const data = localStorage.getItem(STORAGE_KEY)  // ❌ INVALID - Server can't access client storage
  return JSON.parse(data)
}
```

**Questions for you**:
- [ ] Does the backend API have GRN endpoints? (GET, POST, PUT, DELETE /api/v1/grns/*)
- [ ] Should GRN quality issues be persisted to the backend database?
- [ ] Is there a GRN approval workflow that needs backend tracking?
- [ ] What's the complete GRN data model on the backend?

**What needs fixing**:
- Create proper server action that calls backend API instead of localStorage
- Create `use-grn-queries.ts` hook for React Query integration
- Remove localStorage access from server components

---

### 2. **Workflows Configuration Using localStorage as Primary** (MAJOR)
**File**: `frontend/src/app/(private)/admin/workflows/_components/workflows-client.tsx`

**Problem**: Workflows loaded from localStorage on component mount:
```typescript
useEffect(() => {
  const workflows = localStorage.getItem('workflows')
  setWorkflows(JSON.parse(workflows))  // ❌ Should use React Query hook
}, [])
```

**Questions for you**:
- [ ] Does the backend have workflow configuration endpoints?
- [ ] Should workflow definitions be stored in the backend database?
- [ ] Are workflows organization-scoped or global?
- [ ] Is there a workflow publish/activate flow?

**What needs fixing**:
- Create `use-workflow-queries.ts` with:
  - `useWorkflows()` - Fetch all workflows
  - `useWorkflowById(id)` - Fetch specific workflow
  - `useSaveWorkflow()` - Create/update mutation
  - `usePublishWorkflow()` - Publish workflow
- Update workflows-client.tsx to use the Query hook

---

### 3. **Missing Resource Management** (MISSING)

**Vendors**:
- **Current**: Referenced in POs and Requisitions but no dedicated management
- **Backend**: Do you have `/api/v1/vendors` endpoints?
- **Need**: Server actions + Query hooks for vendor CRUD

**Categories**:
- **Current**: Only inline mentions, no management UI
- **Backend**: Do you have `/api/v1/categories` endpoints?
- **Need**: Server actions + Query hooks for category CRUD

**Departments**:
- **Current**: Mock storage in `/lib/mock-departments.ts`
- **Backend**: Should departments come from the backend?
- **Need**: Server actions + Query hooks if backend has department data

---

## What's Working Well

### ✅ Fully Integrated Resources (85% of app)

**1. Requisitions**
- ✅ Backend API: `/api/v1/requisitions`
- ✅ Server actions: `getRequisitions()`, `createRequisition()`, `updateRequisition()`, etc.
- ✅ Query hook: `useRequisitions()`, `useRequisitionById()`
- ✅ localStorage fallback for offline
- ✅ Approval workflow integration

**2. Purchase Orders**
- ✅ Backend API: `/api/v1/purchase-orders`
- ✅ Server actions: Full CRUD
- ✅ Query hooks: `usePurchaseOrders()`, `usePurchaseOrderById()`
- ✅ Auto-linked from Requisitions
- ✅ localStorage fallback

**3. Payment Vouchers**
- ✅ Backend API: `/api/v1/payment-vouchers`
- ✅ Server actions: Full CRUD + mark as paid
- ✅ Query hooks: `usePaymentVouchers()`, `usePaymentVoucherById()`
- ✅ Approval workflow
- ✅ localStorage fallback

**4. Budgets**
- ✅ Backend API: `/api/v1/budgets`
- ✅ Server actions: Full CRUD
- ✅ Query hooks: `useBudgets()`, `useBudgetById()`
- ✅ Approval workflow
- ✅ localStorage fallback

**5. Users, Roles, Permissions**
- ✅ Backend API: `/api/v1/users`, `/api/v1/roles`, `/api/v1/permissions`
- ✅ Server actions: Full management
- ✅ Query hooks: Complete
- ✅ Admin UI: Full CRUD

**6. Tasks & Approvals**
- ✅ Backend API: `/api/v1/tasks`, `/api/v1/approvals`
- ✅ Server actions: Full integration
- ✅ Query hooks: Complete
- ✅ Workflow integration

**7. Organizations**
- ✅ Backend API: `/api/v1/organizations`
- ✅ Server actions: Fetch and switch
- ✅ Context + mutations: `useSelectOrganization()`, `useLogout()`
- ✅ Multi-tenancy: Full isolation

**8. Dashboard & Analytics**
- ✅ Backend API: `/api/v1/dashboard/metrics`
- ✅ Server actions: `getDashboardMetrics()`
- ✅ Components: Fully integrated

---

## localStorage Usage Status

### Appropriate Uses (Offline-First Strategy):
- ✅ `current-organization-id` - UI state for organization switching
- ✅ `screen_lock_state` - Multi-tab synchronization
- ✅ Fallback caching for all major resources (Req/PO/PV/Budget)

### Inappropriate Uses (NEEDS FIX):
- ❌ `workflows` - Should fetch from backend API
- ❌ `app_grns` - Should fetch from backend API
- ⚠️ `recently_used_workflows` - Could stay in localStorage but workflows should come from backend

---

## Current Architecture Patterns

### Pattern 1: Well-Implemented (Requisitions, POs, PVs, Budgets)
```
Backend API
    ↓
Server Action (getRequisitions, etc)
    ↓
React Query Hook (useRequisitions)
    ↓
Component (fetch via hook)
    ↓
localStorage fallback on error
```

### Pattern 2: Partially-Implemented (GRNs)
```
❌ Server Action accessing localStorage directly (INVALID)
    ↓
❌ No React Query hooks
    ↓
❌ Components using localStorage as primary
```

### Pattern 3: Partially-Implemented (Workflows)
```
❌ localStorage as primary source
    ↓
⚠️ Fallback to server action exists
    ↓
❌ No React Query hooks
```

---

## Questions for Clarification

### About GRNs:

1. **GRN Backend API**: Does your backend have these endpoints?
   - `GET /api/v1/grns` - List GRNs
   - `GET /api/v1/grns/{id}` - Get specific GRN
   - `POST /api/v1/grns` - Create GRN
   - `PUT /api/v1/grns/{id}` - Update GRN
   - `POST /api/v1/grns/{id}/quality-issues` - Add quality issue
   - `POST /api/v1/grns/{id}/confirm` - Confirm GRN

2. **GRN Data Persistence**: Should quality issues and GRN data be:
   - Persisted to the backend database?
   - Tracked for audit logging?
   - Part of the approval workflow?

3. **GRN Workflow**: Is GRN part of the requisition/PO/PV flow or standalone?

### About Workflows:

4. **Workflow Backend API**: Does your backend have:
   - `GET /api/v1/workflows` - List workflow templates
   - `GET /api/v1/workflows/{id}` - Get specific workflow
   - `POST /api/v1/workflows` - Create workflow template
   - `PUT /api/v1/workflows/{id}` - Update workflow template
   - `DELETE /api/v1/workflows/{id}` - Delete workflow template

5. **Workflow Scope**: Are workflows:
   - Organization-scoped or global?
   - User-created or predefined?
   - Stored in database or configuration?

6. **Workflow Status**: Should workflow definitions be:
   - Persisted to backend database?
   - Tracked in audit logs?
   - Have versioning support?

### About Missing Resources:

7. **Vendors**:
   - Does backend have `/api/v1/vendors` endpoint?
   - Should vendors be managed by admin users?
   - Are they organization-scoped?

8. **Categories**:
   - Does backend have `/api/v1/categories` endpoint?
   - Are they used for item categorization in requisitions?
   - Are they organization-scoped?

9. **Departments**:
   - Should departments come from backend?
   - Are they used only for user assignment or in other workflows?
   - Organization-scoped or global?

### About Architecture:

10. **Phase 12 Completion**: Is the expectation:
    - All data fetching uses React Query hooks?
    - All mutations use mutation hooks?
    - localStorage only for offline fallback and UI state?
    - No direct API calls from components?

11. **Offline-First Strategy**: Should we:
    - Keep localStorage fallback for all major resources?
    - Implement sync when user comes back online?
    - Show "offline mode" indicators in UI?

---

## Action Plan (Once You Clarify)

### Immediate (Critical):
1. [ ] Fix GRN server actions to call backend API instead of localStorage
2. [ ] Create `use-grn-queries.ts` with full Query hook integration
3. [ ] Update GRN pages to use the Query hooks

### High Priority:
4. [ ] Create `use-workflow-queries.ts` for workflow management
5. [ ] Update workflows admin page to use Query hooks
6. [ ] Fix workflow selector to use hooks instead of localStorage

### Medium Priority:
7. [ ] Create vendor management (server actions + Query hooks)
8. [ ] Create category management (server actions + Query hooks)
9. [ ] Fix department management to use backend if available

### Testing & Documentation:
10. [ ] Update HOOKS-IMPLEMENTATION-GUIDE.md with GRN/Workflow hooks
11. [ ] Add Query hook examples for new resources
12. [ ] Update FRONTEND-REFACTORING-SUMMARY.md with findings

---

## Dependency Notes

- GRN fixes depend on clarifying if backend has GRN API endpoints
- Workflow fixes depend on clarifying if backend has workflow API endpoints
- Vendor/Category fixes depend on whether these should be backend-managed
- Department fixes depend on whether backend has department data

---

**Status**: ✅ Ready for your input
**Next**: Please clarify the questions above so we can proceed with fixes
**Estimated Time to Fix**: 2-3 days once backend API endpoints are confirmed

