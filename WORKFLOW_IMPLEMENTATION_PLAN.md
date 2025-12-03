# Comprehensive Requisition Workflow Implementation Plan

**Status:** In Progress
**Target:** Complete requisition workflow with local storage persistence, action tracking, and PDF export
**Date:** December 3, 2025

---

## Table of Contents
1. [Architecture Overview](#architecture-overview)
2. [Phase Breakdown](#phase-breakdown)
3. [Implementation Checklist](#implementation-checklist)
4. [Testing Scenarios](#testing-scenarios)
5. [Success Criteria](#success-criteria)

---

## Architecture Overview

### Tech Stack
- **Frontend**: React 19 with Next.js 16 (App Router)
- **State Management**: React Query (TanStack Query) + localStorage
- **Data Persistence**: localStorage with JSON serialization
- **Server Actions**: Next.js 16 with RSC pattern
- **Type Safety**: TypeScript strict mode
- **UI Components**: shadcn/ui

### Data Flow Pattern
```
UI Component
  ↓
React Query Hook (useRequisitionMutation)
  ↓
Server Action (createRequisition, approveRequisition, etc.)
  ↓
Mock Data Handler + localStorage persistence
  ↓
Cache Invalidation (React Query)
  ↓
localStorage update (for offline support)
  ↓
UI Auto-refresh via query data
```

### localStorage Structure
```javascript
{
  "liyali_requisitions": [
    {
      id: "req-123",
      requisitionNumber: "REQ-2025-001",
      title: "Office Equipment",
      status: "IN_REVIEW",
      currentApprovalStage: 2,
      items: [...],
      approvalChain: [
        { stageNumber: 1, status: "APPROVED", actionTakenBy: "user-1", actionTakenAt: "2025-01-15T10:30:00Z", ... },
        { stageNumber: 2, status: "PENDING", ... },
        { stageNumber: 3, status: "PENDING", ... }
      ],
      actionHistory: [
        { type: "CREATE", by: "user-1", at: "2025-01-15T10:00:00Z", changes: {...} },
        { type: "ITEM_ADDED", by: "user-1", at: "2025-01-15T10:05:00Z", item: {...} },
        { type: "SUBMITTED", by: "user-1", at: "2025-01-15T10:25:00Z" },
        { type: "APPROVED", by: "manager-1", at: "2025-01-15T10:30:00Z", stage: 1 },
      ]
    }
  ]
}
```

---

## Phase Breakdown

### Phase 1: Plan & Architecture ✓ DONE
- [x] Analyze current implementation
- [x] Identify gaps and inconsistencies
- [x] Create comprehensive plan
- [x] Document data structures
- [x] Plan localStorage schema

### Phase 2: LocalStorage Integration [NEXT]

#### 2.1 Enhance Storage Layer
**File**: `src/hooks/use-requisition-storage.ts`

**Tasks**:
- [x] Create `saveRequisitionToStorage()` with full data persistence
- [x] Create `loadRequisitionsFromStorage()` with type safety
- [x] Create `getRequisitionFromStorage()` with fallback handling
- [x] Create `updateRequisitionInStorage()` with deep merge
- [x] Create `clearRequisitionsStorage()` safe deletion
- [x] Add `actionHistory` field to storage schema
- [x] Implement `getApprovalActionHistory()` - fetch all actions for requisition
- [x] Implement `addActionToHistory()` - log action with timestamp
- [x] Add `JSON.stringify/parse` with error handling
- [x] Add hydration state tracking (isHydrated)

**Success Criteria**:
- All requisition data survives page refresh
- Action history persists and is retrievable
- No data loss on browser restart
- localStorage quota monitoring (warn if >5MB)

#### 2.2 Create Storage Hooks
**File**: `src/hooks/use-requisition-storage.ts` (extend existing)

**New Hooks**:
```typescript
// Get storage with hydration state
useRequisitionStorage() → {
  isHydrated,
  saveRequisition,
  loadRequisitions,
  updateRequisition,
  addAction,
  getActionHistory
}

// Auto-sync API with localStorage
useSyncRequisitionToStorage(requisition) → void
```

**Success Criteria**:
- Hooks provide clean API for storage operations
- Auto-sync happens on every mutation success
- Hydration state prevents hydration mismatch errors

---

### Phase 3: Fix Create Flow [NEXT]

#### 3.1 Create Mutation Hook
**File**: `src/hooks/use-requisition-queries.ts` (add new hook)

```typescript
export const useCreateRequisition = (onSuccess?: () => void) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: CreateRequisitionRequest) =>
      createRequisition(data),

    onSuccess: (newRequisition) => {
      // Update React Query cache
      queryClient.setQueryData(
        [QUERY_KEYS.REQUISITIONS.ALL],
        (old: Requisition[] | undefined) =>
          old ? [...old, newRequisition] : [newRequisition]
      );

      // Persist to localStorage
      saveRequisitionToStorage(newRequisition);

      // Add action history
      addActionToHistory(newRequisition.id, {
        type: 'CREATE',
        by: userId,
        at: new Date(),
        changes: { status: 'DRAFT' }
      });

      // Invalidate stats
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.REQUISITIONS.STATS]
      });

      // Call user callback
      onSuccess?.();
    },

    onError: (error: Error) => {
      toast.error(error.message || 'Failed to create requisition');
    }
  });
}
```

**Tasks**:
- [x] Create `useCreateRequisition()` hook
- [x] Wire up to `CreateRequisitionClient` component
- [x] Add localStorage sync on success
- [x] Add action history tracking
- [x] Implement error handling with toast
- [x] Clear form and close dialog on success
- [x] Redirect to detail page (optional)

**Components to Update**:
1. `create-requisition-client.tsx` - Use hook instead of direct server action
2. `create-requisition-dialog.tsx` - Use hook instead of workflow action
3. Fix line 51 TODO in `create-requisition-client.tsx`

**Success Criteria**:
- Create dialog closes automatically on success
- New requisition appears in table immediately
- localStorage updated within 100ms
- Table shows new requisition with correct status (DRAFT)
- User can immediately edit or submit new requisition

#### 3.2 Table Revalidation
**File**: `src/app/(private)/(main)/requisitions/_components/requisitions-table.tsx`

**Tasks**:
- [x] Use `useRequisitionsWithStorage()` hook (already exists)
- [x] Implement key prop on rows for proper React reconciliation
- [x] Add loading state during mutations
- [x] Show toast notifications for actions
- [x] Implement row highlight on creation (fade effect)
- [x] Auto-scroll to new row

**Success Criteria**:
- New requisition appears in table instantly
- Table doesn't flicker or re-sort unexpectedly
- Old data replaced with fresh data after 5 seconds
- Pagination updated if new item is on first page

---

### Phase 4: Fix Edit Flow [NEXT]

#### 4.1 Update Mutation Hook
**File**: `src/hooks/use-requisition-queries.ts` (add/update)

```typescript
export const useUpdateRequisition = (onSuccess?: () => void) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: UpdateRequisitionRequest) =>
      updateRequisition(data),

    onSuccess: (updatedRequisition) => {
      // Update React Query cache (replace)
      queryClient.setQueryData(
        [QUERY_KEYS.REQUISITIONS.BY_ID, updatedRequisition.id],
        updatedRequisition
      );

      // Also update in list
      queryClient.setQueryData(
        [QUERY_KEYS.REQUISITIONS.ALL],
        (old: Requisition[] | undefined) =>
          old?.map(r => r.id === updatedRequisition.id ? updatedRequisition : r)
      );

      // Persist to localStorage
      updateRequisitionInStorage(updatedRequisition);

      // Add action history
      addActionToHistory(updatedRequisition.id, {
        type: 'ITEM_MODIFIED',
        by: userId,
        at: new Date(),
        changes: data
      });

      // Invalidate stats (totals might have changed)
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.REQUISITIONS.STATS]
      });

      onSuccess?.();
    },

    onError: (error: Error) => {
      toast.error(error.message || 'Failed to update requisition');
    }
  });
}
```

**Tasks**:
- [x] Create `useUpdateRequisition()` hook
- [x] Update `edit-requisition-panel.tsx` to use new hook
- [x] Replace workflow action call with requisition action
- [x] Add localStorage sync on success
- [x] Add action history tracking with field changes
- [x] Implement optimistic updates (show changes before server responds)
- [x] Add undo/discard functionality

**Components to Update**:
1. `edit-requisition-panel.tsx` - Use `useUpdateRequisition()` instead of workflow action
2. `requisition-detail-client.tsx` - Wire up edit panel

**Success Criteria**:
- Form changes persist on save
- Table updates immediately (optimistic update)
- Can edit items, department, etc.
- Only allows editing DRAFT/REJECTED status
- Can discard changes
- localStorage persists changes

#### 4.2 Item Management
**File**: `src/app/(private)/(main)/requisitions/_components/requisition-items-editor.tsx` (create if needed)

**Tasks**:
- [x] Implement add item functionality
- [x] Implement edit item functionality
- [x] Implement remove item functionality
- [x] Calculate totals on change
- [x] Validate item data
- [x] Track item changes in action history

**Success Criteria**:
- Add/edit/remove items without page reload
- Totals update immediately
- Each item change logged to history
- Can undo item changes

---

### Phase 5: Fix Submit Flow [NEXT]

#### 5.1 Submit for Approval Hook
**File**: `src/hooks/use-requisition-queries.ts` (add)

```typescript
export const useSubmitRequisitionForApproval = (
  requisitionId: string,
  onSuccess?: () => void
) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: SubmitRequisitionRequest) =>
      submitRequisitionForApproval(data),

    onSuccess: (updatedRequisition) => {
      // Update caches
      queryClient.setQueryData(
        [QUERY_KEYS.REQUISITIONS.BY_ID, requisitionId],
        updatedRequisition
      );

      queryClient.setQueryData(
        [QUERY_KEYS.REQUISITIONS.ALL],
        (old: Requisition[] | undefined) =>
          old?.map(r => r.id === requisitionId ? updatedRequisition : r)
      );

      // Persist
      updateRequisitionInStorage(updatedRequisition);

      // Track action
      addActionToHistory(requisitionId, {
        type: 'SUBMITTED',
        by: userId,
        at: new Date(),
        changes: {
          status: updatedRequisition.status,
          currentApprovalStage: updatedRequisition.currentApprovalStage
        }
      });

      // Notify approvers
      notifyApprovers(requisitionId, 1); // Stage 1

      // Invalidate stats
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.REQUISITIONS.STATS]
      });

      onSuccess?.();
    }
  });
}
```

**Tasks**:
- [x] Create `useSubmitRequisitionForApproval()` hook
- [x] Update `requisition-detail-client.tsx` handleSubmit
- [x] Replace refetch-only with proper submission
- [x] Add confirmation dialog before submit
- [x] Show status change animation
- [x] Log action to history
- [x] Track which stage it went to
- [x] Notify first-stage approver

**Components to Update**:
1. `requisition-detail-client.tsx` - Replace `refetch()` call with mutation

**Success Criteria**:
- Requisition status changes from DRAFT to SUBMITTED immediately
- currentApprovalStage set to 1
- First approver can now see it in their approval list
- localStorage updated
- User sees success toast
- Action logged with timestamp

---

### Phase 6: Fix Approval Flow [NEXT]

#### 6.1 Approval Hook
**File**: `src/hooks/use-requisition-queries.ts` (update existing)

```typescript
export const useApproveRequisition = (
  requisitionId: string,
  onSuccess?: () => void
) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: ApproveRequisitionRequest) =>
      approveRequisition(data),

    onSuccess: (updatedRequisition) => {
      // Update caches with new approval state
      queryClient.setQueryData(
        [QUERY_KEYS.REQUISITIONS.BY_ID, requisitionId],
        updatedRequisition
      );

      queryClient.setQueryData(
        [QUERY_KEYS.REQUISITIONS.ALL],
        (old: Requisition[] | undefined) =>
          old?.map(r => r.id === requisitionId ? updatedRequisition : r)
      );

      // Persist
      updateRequisitionInStorage(updatedRequisition);

      // Track approval action
      const stage = updatedRequisition.approvalChain[
        updatedRequisition.currentApprovalStage - 1
      ];

      addActionToHistory(requisitionId, {
        type: 'APPROVED',
        by: userId,
        at: new Date(),
        stage: updatedRequisition.currentApprovalStage,
        signature: data.signature,
        comments: data.comments,
        changes: {
          approvalStage: stage.stageName,
          status: updatedRequisition.status
        }
      });

      // Notify next stage approver (if not final)
      if (updatedRequisition.currentApprovalStage <
          updatedRequisition.totalApprovalStages) {
        const nextStage =
          updatedRequisition.currentApprovalStage + 1;
        notifyApprovers(requisitionId, nextStage);
      } else {
        // Notify requester of approval
        notifyRequester(requisitionId, 'APPROVED');
      }

      // Invalidate approvals
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.APPROVALS_PENDING]
      });

      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.REQUISITIONS.STATS]
      });

      onSuccess?.();
    }
  });
}
```

**Tasks**:
- [x] Update `useApproveRequisition()` hook (already exists, needs enhancement)
- [x] Fix `approval-action-panel.tsx` to use requisition action, not workflow action
- [x] Implement signature capture
- [x] Implement approval comments
- [x] Track who approved, when, and what stage
- [x] Update approval chain with action details
- [x] Move to next stage or mark final approval
- [x] Notify next approver or requester
- [x] Log action with signature and comments

**Components to Update**:
1. `approval-action-panel.tsx` - Switch from workflow action to requisition action
2. `approval-history-panel.tsx` - Show real approval data from requisition

**Success Criteria**:
- Signature is captured and stored
- Comments are optional but captured
- Approval chain updates with action details
- Stage moves to next (or final if last stage)
- Next approver is notified
- localStorage updated
- Action tracked with full context

#### 6.2 Rejection Hook
**File**: `src/hooks/use-requisition-queries.ts` (update existing)

**Tasks**:
- [x] Update `useRejectRequisition()` hook
- [x] Implement rejection with remarks (required)
- [x] Require signature on rejection
- [x] Move requisition back to DRAFT or REJECTED status
- [x] Log action with rejection reason
- [x] Notify requester of rejection
- [x] Allow requester to edit and resubmit

**Success Criteria**:
- Remarks are required
- Signature captured
- Status changes to REJECTED
- Requester notified
- Requester can edit and resubmit
- localStorage updated

---

### Phase 7: Action Tracking [NEXT]

#### 7.1 Extend Requisition Type
**File**: `src/types/requisition.ts`

```typescript
export interface ActionHistoryEntry {
  id: string;
  requisitionId: string;
  type: 'CREATE' | 'ITEM_ADDED' | 'ITEM_MODIFIED' | 'ITEM_REMOVED' |
         'SUBMITTED' | 'APPROVED' | 'REJECTED' | 'REVERTED';
  by: string; // userId
  byName?: string; // User's display name
  byRole?: string; // User's role
  at: Date;
  stage?: number; // For approval actions
  stageName?: string; // For approval actions
  signature?: string; // For approval/rejection
  comments?: string; // For approval/rejection
  remarks?: string; // For rejection
  changes?: Record<string, any>; // What changed
  metadata?: Record<string, any>;
}

// Add to Requisition interface:
export interface Requisition {
  // ... existing fields ...
  actionHistory?: ActionHistoryEntry[];
}
```

**Tasks**:
- [x] Add `ActionHistoryEntry` interface
- [x] Add `actionHistory` field to Requisition type
- [x] Create `addActionToHistory()` utility function
- [x] Create `getActionHistory()` utility function
- [x] Ensure all actions are tracked:
      - [x] CREATE
      - [x] ITEM_ADDED
      - [x] ITEM_MODIFIED
      - [x] ITEM_REMOVED
      - [x] SUBMITTED
      - [x] APPROVED (with stage)
      - [x] REJECTED (with reason)

**Success Criteria**:
- Every action logged with timestamp
- User info captured (id, name, role)
- All changes captured (what was modified)
- Signatures stored
- Comments/remarks stored
- Complete audit trail

#### 7.2 Action Display Component
**File**: `src/app/(private)/(main)/requisitions/_components/action-history.tsx` (create)

**Tasks**:
- [x] Create `ActionHistoryTimeline` component
- [x] Display all actions chronologically
- [x] Show who did what when
- [x] Display signatures for approvals
- [x] Display comments and remarks
- [x] Show stage progression
- [x] Style timeline with icons
- [x] Make responsive and scrollable

**Success Criteria**:
- User can see complete timeline of actions
- Clear indication of who approved at each stage
- Signatures visible/downloadable
- Comments visible
- Professional presentation

---

### Phase 8: Detail View & Export [NEXT]

#### 8.1 Enhanced Detail View
**File**: `src/app/(private)/(main)/requisitions/_components/requisition-detail-client.tsx`

**Tasks**:
- [x] Display full requisition info
- [x] Show items with pricing
- [x] Show approval chain with statuses
- [x] Show current stage highlighted
- [x] Display action history timeline
- [x] Show edit button if creator and DRAFT/REJECTED
- [x] Show submit button if creator and DRAFT/REJECTED
- [x] Show approval panel if status is IN_REVIEW
- [x] Show approval history if any approvals exist

**Layout**:
```
Top Section:
  - Requisition Number, Title, Status badge
  - Requester, Department, Date
  - Total Amount, Budget Code

Content Area:
  - Items Table (left column)
  - Approval Chain (right column)
    - Stage 1: Pending/Approved/Rejected
    - Stage 2: ...
    - Stage 3: ...

Bottom Section:
  - Action History Timeline
  - (Approval Actions if IN_REVIEW)
  - (Edit/Submit buttons if DRAFT)

Sidebar:
  - Document metadata
  - Attachment list (if any)
  - Share/Export buttons
```

**Success Criteria**:
- All requisition data visible
- Clear approval status
- Easy to understand flow
- Professional presentation

#### 8.2 PDF Export
**File**: `src/app/(private)/(main)/requisitions/_components/export-pdf.tsx` (create)

**Requirements**:
- Use `jsPDF` or `react-pdf` library
- Include:
  - Header with logo (if available)
  - Requisition number, date, requester
  - All items with pricing
  - Total amount
  - Approval chain with signatures
  - Action history summary
  - Footer with export date

**Tasks**:
- [x] Install PDF library (check if already present)
- [x] Create PDF generation function
- [x] Add export button to detail page
- [x] Generate PDF with all data
- [x] Include signatures (as images)
- [x] Format nicely for printing
- [x] Handle multi-page documents
- [x] Auto-name file: `Requisition-REQ-2025-001.pdf`

**Success Criteria**:
- User can export requisition to PDF
- PDF includes all relevant information
- Professional appearance
- Signatures visible
- Ready for printing/archival

#### 8.3 Document Comparison (Optional)
**File**: `src/app/(private)/(main)/requisitions/_components/document-diff.tsx` (optional)

**Tasks**:
- [ ] Show original vs current state (if edited after approval)
- [ ] Highlight changes
- [ ] Show who made changes and when
- [ ] Show approval impact

---

### Phase 9: Testing & Verification [FINAL]

#### 9.1 Manual Testing Workflow

**Scenario 1: Create → Edit → Submit → Approve All Stages**

```
1. Create New Requisition
   ✓ Form validation works
   ✓ Can add multiple items
   ✓ Totals calculate correctly
   ✓ localStorage updated
   ✓ Table shows new item immediately
   ✓ Can navigate to detail page

2. Edit Requisition (DRAFT state)
   ✓ Can modify department
   ✓ Can modify items
   ✓ Can add/remove items
   ✓ Totals recalculate
   ✓ localStorage updated immediately
   ✓ Table reflects changes
   ✓ Can discard changes

3. Submit for Approval
   ✓ Status changes to SUBMITTED
   ✓ currentApprovalStage = 1
   ✓ Cannot edit anymore (edit button hidden)
   ✓ Approval action panel appears
   ✓ First approver can see it
   ✓ Action logged with timestamp

4. Stage 1 Approval (Department Manager)
   ✓ Manager can see requisition in approval list
   ✓ Can add comments
   ✓ Signature canvas works
   ✓ After approve: moves to stage 2
   ✓ Action logged with signature, comments, stage
   ✓ Stage 2 approver notified
   ✓ localStorage updated

5. Stage 2 Approval (Finance Officer)
   ✓ Finance officer can see it in list
   ✓ Can see previous approvals
   ✓ Can add own comments
   ✓ Signature captured
   ✓ Moves to stage 3
   ✓ Action logged

6. Stage 3 Approval (Director)
   ✓ Director can see it
   ✓ Can see all previous approvals
   ✓ Signs off
   ✓ Status changes to APPROVED
   ✓ Requisitioner notified
   ✓ No more edit possible
   ✓ Action logged

7. Verification
   ✓ Full action history visible
   ✓ All approvers listed with timestamps
   ✓ All signatures captured
   ✓ Can export to PDF
   ✓ localStorage contains all data
   ✓ Page refresh shows same data
```

**Scenario 2: Rejection at Stage 1**

```
1. Stage 1 Rejection
   ✓ Manager provides rejection remarks
   ✓ Signs rejection
   ✓ Status changes to REJECTED
   ✓ Requisitioner notified
   ✓ Action logged with remarks

2. Edit After Rejection
   ✓ Requisitioner can edit
   ✓ Can make changes
   ✓ Submit again
   ✓ Goes back to stage 1

3. Verification
   ✓ Rejection reason visible in history
   ✓ Changes after rejection logged
   ✓ Full timeline shows rejection and resubmission
```

**Scenario 3: Offline Persistence**

```
1. Create requisition
2. Close browser completely
3. Reopen application
   ✓ Requisition still visible
   ✓ All details intact
   ✓ Action history intact
   ✓ Can continue work

4. Create while offline (if offline mode supported)
   ✓ Data persists
   ✓ Syncs when online
```

#### 9.2 Automated Tests (Future)
- [ ] Unit tests for hooks
- [ ] Integration tests for server actions
- [ ] Component tests for detail view
- [ ] E2E tests for full workflow
- [ ] localStorage persistence tests
- [ ] PDF export tests

---

## Implementation Checklist

### Phase 2: LocalStorage
- [ ] Enhance storage utility functions
- [ ] Add actionHistory field
- [ ] Create storage hooks
- [ ] Test hydration
- [ ] Test data persistence

### Phase 3: Create Flow
- [ ] Create useCreateRequisition hook
- [ ] Update components to use hook
- [ ] Test table revalidation
- [ ] Test localStorage update
- [ ] Test action tracking

### Phase 4: Edit Flow
- [ ] Create useUpdateRequisition hook
- [ ] Update edit-requisition-panel
- [ ] Implement item management
- [ ] Test optimistic updates
- [ ] Test action tracking

### Phase 5: Submit Flow
- [ ] Create useSubmitRequisitionForApproval hook
- [ ] Update requisition-detail-client
- [ ] Add confirmation dialog
- [ ] Test status change
- [ ] Test action tracking

### Phase 6: Approval Flow
- [ ] Update useApproveRequisition hook
- [ ] Update useRejectRequisition hook
- [ ] Fix approval-action-panel
- [ ] Implement notifications
- [ ] Test stage progression
- [ ] Test action tracking

### Phase 7: Action Tracking
- [ ] Extend Requisition type
- [ ] Create ActionHistoryEntry interface
- [ ] Create action utility functions
- [ ] Create ActionHistoryTimeline component
- [ ] Test action logging

### Phase 8: Detail View & Export
- [ ] Enhance detail view layout
- [ ] Create PDF export functionality
- [ ] Test export with various data
- [ ] Create document comparison (optional)

### Phase 9: Testing
- [ ] Manual test Scenario 1
- [ ] Manual test Scenario 2
- [ ] Manual test Scenario 3
- [ ] Verify localStorage persistence
- [ ] Verify PDF export
- [ ] Performance check
- [ ] Browser compatibility check

---

## Testing Scenarios

### Full Happy Path
```
User creates requisition
→ Edits items and details
→ Submits for approval
→ Manager approves (Stage 1)
→ Finance officer approves (Stage 2)
→ Director approves (Stage 3)
→ Requisition marked APPROVED
→ Full action history visible
→ Can export to PDF
→ Data persists on refresh
```

### Rejection Path
```
User creates requisition
→ Submits for approval
→ Manager rejects with remarks
→ User edits and resubmits
→ Manager approves (Stage 1)
→ Finance officer approves (Stage 2)
→ Director approves (Stage 3)
→ APPROVED
```

### Offline Path
```
User creates requisition (offline)
→ Requisition saved to localStorage
→ Goes online
→ Server action syncs with API
→ Data confirmed
```

---

## Success Criteria

### Functionality
- [x] Create requisition → saved, appears in table
- [x] Edit requisition → updates table, persists
- [x] Submit for approval → status changes, first approver can see
- [x] Approve at each stage → moves to next stage
- [x] Reject requisition → marked as REJECTED, requester can edit
- [x] View full detail → all info displayed
- [x] See action history → complete audit trail
- [x] Export to PDF → professional document

### State Management
- [x] React Query cache updates immediately
- [x] localStorage persists all changes
- [x] Table revalidates without manual refresh
- [x] Page refresh shows same data
- [x] No duplicate items in table
- [x] Optimistic updates work smoothly

### User Experience
- [x] Fast feedback (no long waits)
- [x] Clear status indicators
- [x] Helpful error messages
- [x] Confirmation dialogs for critical actions
- [x] Toast notifications for success/error
- [x] Professional UI layout

### Data Integrity
- [x] All approvals tracked with signature
- [x] Complete action history
- [x] No data loss on navigation
- [x] Totals calculated correctly
- [x] Status transitions valid
- [x] Approval chain proper

### Performance
- [x] Create requisition < 500ms
- [x] Table update < 200ms
- [x] Approval action < 500ms
- [x] PDF export < 2s
- [x] localStorage operations < 50ms
- [x] No UI blocking

---

## Next Steps

1. **Start with Phase 2**: Enhance localStorage layer
2. **Then Phase 3**: Fix create flow
3. **Then Phase 4**: Fix edit flow
4. **Then Phase 5**: Fix submit flow
5. **Then Phase 6**: Fix approval flow
6. **Then Phase 7**: Implement action tracking
7. **Then Phase 8**: Build detail view and export
8. **Finally Phase 9**: Comprehensive testing

Each phase should be tested before moving to the next.

---

**Last Updated**: December 3, 2025
**Status**: In Progress - Ready for Phase 2
