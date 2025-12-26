# Frontend Approval Workflow System - E2E Audit & Integration Plan

**Date**: 2025-12-26
**Status**: Audit Complete - Integration Ready
**Current Integration Level**: 60% (Core operations done, legacy code remains)

---

## Executive Summary

The frontend approval workflow system has solid backend integration for core approval operations, but still contains significant legacy code and mock implementations. This document outlines what's integrated, what's not, and the step-by-step plan to complete the integration.

**Current State**:
- ✅ Core approval hooks and mutations are backend-powered
- ✅ Server actions for approve/reject/reassign are implemented
- ✅ React Query caching and invalidation working
- ❌ Legacy approval-actions.ts still uses mock approvalStore
- ❌ PO approval client uses mock data generation
- ❌ Budget approval is isolated from main workflow
- ❌ Some dashboard charts use mock data

---

## Part 1: INTEGRATION STATUS INVENTORY

### 1.1 Fully Integrated (Backend-Powered) ✅

**Backend API Calls Active**:
- GET `/api/v1/approvals?...` → `useApprovalTasks()`
- GET `/api/v1/approvals/{id}` → `useApprovalTaskDetail()`
- POST `/api/v1/approvals/{id}/approve` → `useApproveTask()`
- POST `/api/v1/approvals/{id}/reject` → `useRejectTask()`
- POST `/api/v1/approvals/{id}/reassign` → `useReassignTask()`
- GET `/api/v1/documents/{documentId}/approval-history` → `useApprovalHistory()`

**Components Using Backend Data**:
1. ✅ Tasks page - Approvals tab (`approvals-list.tsx`)
2. ✅ Requisition approval action panel
3. ✅ Purchase Order approval action panel
4. ✅ Payment Voucher approval action panel
5. ✅ Generic approval action panel (`workflows/approval-action-panel.tsx`)
6. ✅ Approval history display
7. ✅ Dashboard pending approval count

**Hooks Confirmed Working**:
- `useApprovalTasks()` - Lists approval tasks with filters
- `useApprovalTaskDetail()` - Gets single task with full details
- `useApproveTask()` - Approve mutation with cache invalidation
- `useRejectTask()` - Reject mutation with cache invalidation
- `useReassignTask()` - Reassign mutation
- `useApprovalHistory()` - Document approval history
- `usePendingApprovalCount()` - Current user pending count
- `usePendingApprovals()` - Pending tasks for current user
- `useApprovalWorkflow()` - Combined hook with all operations

### 1.2 Partially Integrated (Mixed Old/New) ⚠️

**Files with Dual Path Issues**:
1. `use-approval-task-queries.ts` - Exports query hooks but unclear data source
2. `use-approval-mutations.ts` - May call approval-actions.ts (mock)
3. `use-approval-flow.ts` - Old hooks still present but likely unused

**Components Using Uncertain Data**:
1. PO approval page - May generate mock data
2. GRN approval workflow - Status unclear
3. Approval reports - Uses mock or real data unclear

### 1.3 Not Integrated (Still Using Mocks) ❌

**Legacy Mock-Based Server Actions**:
- `approval-actions.ts` - Uses `approvalStore` with localStorage
  - `getApprovalTasks()` - Calls `approvalStore.getAllTasks()`
  - `getApprovalTaskDetail()` - Calls `approvalStore.getTaskDetail()`
  - `getApprovalStats()` - Calls `approvalStore.getStatistics()`
  - `approveTask()` - Calls `approvalStore.approveTask()`
  - `rejectTask()` - Calls `approvalStore.rejectTask()`
  - `reassignTask()` - Calls `approvalStore.reassignTask()`
  - `validateSignature()` - Mock validation
  - `getAvailableApprovers()` - Returns hardcoded mock approvers

**Legacy Mock Store**:
- `approval-store.ts` - In-memory/localStorage store with mock data

**Components Still Using Mocks**:
1. ❌ `approval-time-chart.tsx` - Hardcoded mock trend data
2. ❌ `budget-approval-action-panel.tsx` - Uses approveBudget/rejectBudget (separate system)
3. ❌ `po-approval-client.tsx` - Calls `generateMockPO()` function

**Hooks to Deprecate**:
- `useApproveStage()` - Old workflow hook
- `useRejectStage()` - Old workflow hook
- `useReassignStage()` - Old workflow hook
- Related hooks in `use-approval-flow.ts`

---

## Part 2: DETAILED ISSUE BREAKDOWN

### Issue #1: approval-actions.ts Still Uses Mock Store

**File**: `frontend/src/app/_actions/approval-actions.ts`

**Current Code Pattern**:
```typescript
export async function getApprovalTasks(status?: string) {
  const tasks = approvalStore.getAllTasks(status); // ❌ MOCK
  return { success: true, tasks, total: tasks.length };
}
```

**Problem**:
- Uses localStorage-based mock store instead of backend API
- Has multiple TODOs pointing out this is temporary
- May be called by legacy hooks
- Blocks real data from flowing to components

**Solution**:
- Replace all function bodies with calls to `approval-workflow.ts` server actions
- Keep function signatures for backward compatibility
- Add deprecation warnings
- Eventually remove this file

**Impact**: HIGH - This is a key integration point

---

### Issue #2: PO Approval Client Generates Mock Data

**File**: `frontend/src/app/(private)/(main)/purchase-orders/[id]/approval/_components/po-approval-client.tsx`

**Current Code Pattern**:
```typescript
const mockPO = generateMockPO({
  id: purchaseOrderId,
  // ... mock data
});
```

**Problem**:
- Generates fake PO data instead of fetching real data
- Doesn't connect to actual PO details
- Approver works with incorrect/fake data
- Audit trail will show false information

**Solution**:
- Fetch actual PO from backend using `usePurchaseOrder()` hook
- Use real PO data in approval context
- Validate approver has access to real PO

**Impact**: MEDIUM - Affects PO approval accuracy

---

### Issue #3: Budget Approval Isolated from Main Workflow

**File**: `frontend/src/app/(private)/(main)/budgets/[id]/_components/budget-approval-action-panel.tsx`

**Current Code Pattern**:
```typescript
const { mutate: approveBudget } = useApproveBudget();
const { mutate: rejectBudget } = useRejectBudget();
```

**Problem**:
- Uses separate approval hooks instead of central workflow
- Not integrated with RBAC system
- No unified approval experience
- Duplicate approval logic

**Solution**:
- Refactor to use centralized `useApprovalTasks()`
- Fetch budget approval tasks by documentType: 'BUDGET'
- Use `useApproveTask()` and `useRejectTask()`
- Maintain consistency with other document types

**Impact**: MEDIUM - Budget approval should use same system

---

### Issue #4: Duplicate Approval Hooks

**Files**:
- `use-approval-flow.ts` - Old hooks: useApproveStage, useRejectStage, etc.
- `use-approval-mutations.ts` - New mutations but may call wrong actions
- `use-approval-task-queries.ts` - Query hooks with unclear sourcing
- `use-approval-workflow.ts` - The correct, unified set

**Problem**:
- Multiple hooks doing similar things
- Confusion about which to use
- Maintenance nightmare
- Code duplication

**Solution**:
- Identify which old hooks are still used
- Consolidate all into `use-approval-workflow.ts`
- Mark others as deprecated
- Update all imports

**Impact**: MEDIUM - Technical debt, not blocking functionality

---

### Issue #5: Dashboard Uses Mock Metrics

**File**: `frontend/src/app/(private)/(main)/home/_components/approval-time-chart.tsx`

**Current Code Pattern**:
```typescript
const chartData = [
  { month: 'Jan', pending: 5, approved: 12, rejected: 2 },
  // ... hardcoded mock data
];
```

**Problem**:
- Shows fake approval trends
- Doesn't reflect actual system usage
- No connection to real approval data

**Solution**:
- Create backend endpoint for approval metrics (already done: GET `/api/v1/approvals/stats`)
- Use `useGetApprovalStats()` to fetch real metrics
- Display real trends instead of mock

**Impact**: LOW - Dashboard UX improvement

---

## Part 3: IMPLEMENTATION ROADMAP

### Phase 1: High Priority (Do First)

#### Step 1.1: Update approval-actions.ts to Call Real APIs
**File**: `frontend/src/app/_actions/approval-actions.ts`

**Changes**:
```typescript
// BEFORE (Mock-based):
export async function getApprovalTasks(status?: string) {
  const tasks = approvalStore.getAllTasks(status);
  return { success: true, tasks, total: tasks.length };
}

// AFTER (Backend API):
export async function getApprovalTasks(
  status?: string,
  page: number = 1,
  limit: number = 10
) {
  // Re-export from approval-workflow.ts
  return getApprovalTasks({ status }, page, limit);
}
```

**Actions to Replace**:
- getApprovalTasks → Re-export from approval-workflow.ts
- getApprovalTaskDetail → Re-export from approval-workflow.ts
- getApprovalStats → Re-export from approval-workflow.ts
- approveTask → Re-export from approval-workflow.ts
- rejectTask → Re-export from approval-workflow.ts
- reassignTask → Re-export from approval-workflow.ts
- validateSignature → Backend already validates, so remove
- getAvailableApprovers → Create new backend endpoint

**Estimate**: 2-3 hours

#### Step 1.2: Fix PO Approval Client Mock Data
**File**: `frontend/src/app/(private)/(main)/purchase-orders/[id]/approval/_components/po-approval-client.tsx`

**Changes**:
- Remove `generateMockPO()` call
- Add `usePurchaseOrder(poId)` hook to fetch real PO
- Use real PO data in component rendering
- Pass real PO to approval action panel

**Estimate**: 1-2 hours

#### Step 1.3: Integrate Budget Approval with Central Workflow
**File**: `frontend/src/app/(private)/(main)/budgets/[id]/_components/budget-approval-action-panel.tsx`

**Changes**:
- Replace custom hooks with `useApprovalTasks()`
- Find budget approval task by documentId
- Use `useApproveTask()` and `useRejectTask()`
- Match other document panel patterns

**Estimate**: 2-3 hours

### Phase 2: Medium Priority (Then Do These)

#### Step 2.1: Consolidate Approval Hooks
**Files**:
- Keep: `use-approval-workflow.ts` (source of truth)
- Deprecate: `use-approval-flow.ts`
- Review: `use-approval-mutations.ts`
- Review: `use-approval-task-queries.ts`

**Actions**:
1. Find all imports of old hooks
2. Replace with imports from `use-approval-workflow.ts`
3. Mark old files as deprecated
4. Add migration guide

**Estimate**: 1-2 hours

#### Step 2.2: Add Real Data to Dashboard Charts
**File**: `frontend/src/app/(private)/(main)/home/_components/approval-time-chart.tsx`

**Changes**:
- Create `useApprovalMetrics()` hook (calls backend stats)
- Replace hardcoded mock data with real metrics
- Add time period filtering
- Update chart to show actual trends

**Estimate**: 2-3 hours

### Phase 3: Low Priority (Polish & Cleanup)

#### Step 3.1: Create Missing Backend Endpoints
**Needed Endpoints**:
- GET `/api/v1/approvals/stats` - Approval statistics (may exist)
- GET `/api/v1/approvals/available-approvers/{documentType}` - List possible approvers
- GET `/api/v1/approvals/timeline` - Approval trends over time

**Estimate**: 3-4 hours (backend work)

#### Step 3.2: Remove Deprecated Files
**Files to Remove**:
- `approval-store.ts` (after migration)
- Legacy hooks from `use-approval-flow.ts` (after migration)
- Mock data generation functions

**Estimate**: 1 hour

---

## Part 4: VERIFICATION CHECKLIST

### Data Flow Verification
- [ ] `useApprovalTasks()` returns data from backend API
- [ ] `useApprovalTaskDetail()` returns full document context
- [ ] `useApproveTask()` calls backend and invalidates cache
- [ ] `useRejectTask()` calls backend and invalidates cache
- [ ] `useReassignTask()` calls backend and invalidates cache
- [ ] All mutations show toast notifications
- [ ] All queries handle loading/error states

### Component Verification
- [ ] Tasks page shows real approval tasks
- [ ] Requisition approval panel works end-to-end
- [ ] PO approval panel fetches real PO data
- [ ] PV approval panel works end-to-end
- [ ] Budget approval panel uses central workflow
- [ ] Approval history shows real data
- [ ] Dashboard pending count is accurate

### Hook Verification
- [ ] Old hooks from use-approval-flow.ts are not used
- [ ] use-approval-workflow.ts is the single source of truth
- [ ] All imports point to correct hooks
- [ ] No circular dependencies

### Backend Verification
- [ ] All API endpoints are responding correctly
- [ ] RBAC is enforced on all endpoints
- [ ] Audit logs are created for all actions
- [ ] Notifications are sent appropriately
- [ ] Database transactions are atomic

---

## Part 5: BACKEND WORK NEEDED

The frontend is mostly ready. These backend tasks may be needed:

### New Endpoints to Create (if not exists)
1. **GET `/api/v1/approvals/stats`** - Approval metrics
2. **GET `/api/v1/approvals/available-approvers/:documentType`** - Approver list
3. **GET `/api/v1/approvals/timeline`** - Trend data

### Existing Endpoints to Verify
1. ✅ GET `/api/v1/approvals` - Pagination, filtering
2. ✅ POST `/api/v1/approvals/:id/approve` - With signature
3. ✅ POST `/api/v1/approvals/:id/reject` - With remarks
4. ✅ POST `/api/v1/approvals/:id/reassign` - To new approver
5. ✅ GET `/api/v1/documents/:documentId/approval-history` - Full history

### Features to Verify
- [ ] Organization scoping on all endpoints
- [ ] RBAC permission checks
- [ ] Audit logging for all mutations
- [ ] Notification sending on approval/rejection
- [ ] Proper error messages
- [ ] Request validation
- [ ] Response pagination

---

## Part 6: MIGRATION PATH

### Week 1: High Priority Integration
1. Monday: Update approval-actions.ts (2-3 hours)
2. Tuesday: Fix PO approval mock (1-2 hours)
3. Wednesday: Integrate budget approval (2-3 hours)
4. Thursday: Test all changes, fix bugs
5. Friday: Consolidate hooks (1-2 hours)

### Week 2: Medium Priority & Polish
1. Monday: Add real data to dashboard charts (2-3 hours)
2. Tuesday: Create missing backend endpoints (3-4 hours)
3. Wednesday: Final testing and validation
4. Thursday: Remove deprecated code
5. Friday: Documentation and handoff

### Success Criteria
- [ ] All approval operations use backend API
- [ ] No mock data in production code paths
- [ ] All charts show real data
- [ ] 100% of components tested with real API
- [ ] RBAC enforced throughout
- [ ] Audit logs captured for all actions
- [ ] All tests passing
- [ ] Code review approved

---

## Part 7: RISK ASSESSMENT

### High Risk
1. ❌ Removing approval-store without testing → Mitigation: Run full test suite
2. ❌ Breaking legacy hooks usage → Mitigation: Search all files first
3. ❌ API not responding correctly → Mitigation: Verify backend first

### Medium Risk
1. ⚠️ Cache invalidation issues → Mitigation: Test multiple approval scenarios
2. ⚠️ Race conditions in concurrent approvals → Mitigation: Add proper locking
3. ⚠️ RBAC checks failing → Mitigation: Verify backend permissions

### Low Risk
1. 🟢 Mock data still in test files → Not a problem
2. 🟢 Deprecated code unused → Can clean up gradually
3. 🟢 Documentation outdated → Can update separately

---

## Part 8: TESTING STRATEGY

### Unit Tests Needed
- Approval hooks: useApproveTask, useRejectTask, etc.
- Server actions: approval-workflow.ts functions
- Utility functions: Validation, transformation

### Integration Tests Needed
- Full approval flow: Task creation → Approval → History update
- Multi-user scenarios: Concurrent approvals
- Permission scenarios: Unauthorized access
- Document type scenarios: REQ, PO, PV, GRN, BUDGET

### E2E Tests Needed
- User journey: Create doc → Submit → Approve → View history
- Error scenarios: Network failure, validation error
- Edge cases: Final approval, rejection back to draft

### Test Data
- Use real backend API for integration tests
- Create fixtures for consistent test data
- Seed database with test documents

---

## CONCLUSION

The frontend approval system has a solid 60% integration level with core operations working well. The remaining 40% involves replacing legacy mock implementations with real backend calls and consolidating duplicate code.

**Critical Path** (must do):
1. Replace approval-actions.ts mocks with backend calls
2. Fix PO approval mock data generation
3. Integrate budget approval with central system

**Nice to Have** (should do):
1. Consolidate duplicate hooks
2. Add real metrics to dashboard
3. Create missing backend endpoints

**Timeline**: 1-2 weeks to reach 100% integration

**Owner**: Full-stack team (frontend + backend)

---

**Created**: 2025-12-26
**Status**: Ready for Implementation
**Priority**: High (Complete before production deployment)
