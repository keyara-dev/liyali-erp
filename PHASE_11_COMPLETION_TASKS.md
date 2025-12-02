# Phase 11 Completion Tasks - Polish & Polish E2E

**Current Status**: Phase 11 Core Complete, UX Polish Needed
**Estimated Effort**: 1 day
**Priority**: High - Improves user experience significantly

---

## Quick Start: What to Do Next

You have **3 options**:

### Option A: Ship as-is to Phase 12 (2 hours)

- ✅ Core workflows work
- ⚠️ No user feedback (toasts)
- ⚠️ No advanced form validation
- ✅ Data persists correctly
- **Use when**: You want to move to Phase 12 database work ASAP

### Option B: Quick Polish (4 hours)

- ✅ Add toast notifications
- ✅ Add basic error boundaries
- ✅ Better loading states
- ⚠️ Skip form validation library
- **Use when**: You want better UX but need to move quickly

### Option C: Complete Polish (8 hours)

- ✅ Add toast notifications
- ✅ Add form validation (Zod + React Hook Form)
- ✅ Add error boundaries
- ✅ Add E2E tests
- ✅ Add loading skeletons
- **Use when**: You want production-quality code before Phase 12

---

## Task 1: Add Toast Notifications (30 min) ⏱️

### Why

Users need feedback when they approve/reject tasks. Currently: no confirmation message.

### How

**Step 1: Install Sonner**

```bash
npm install sonner
```

**Step 2: Add toast provider to layout**

```typescript
// src/app/layout.tsx
import { Toaster } from 'sonner'

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html>
      <body>
        {children}
        <Toaster position="top-right" />
      </body>
    </html>
  )
}
```

**Step 3: Update approval-action-panel.tsx**

```typescript
import { toast } from 'sonner'

// In handleApproveSubmit
const handleApproveSubmit = async (signature: string, remarks?: string) => {
  try {
    await approveMutation.mutateAsync({
      assignmentId: task.id,
      stageNumber: task.stageIndex || 0,
      approvingUserId: task.approverUserId || "",
      signature,
      comments: remarks,
    });

    // ✅ ADD THIS
    toast.success('Task approved successfully!', {
      description: `${task.entityType} #${task.entityNumber} approved`,
    })

    setApproveModalOpen(false);
    onApprovalComplete?.();
  } catch (error) {
    // ✅ ADD THIS
    toast.error('Failed to approve task', {
      description: error instanceof Error ? error.message : 'Unknown error',
    })
    console.error("Approval failed:", error);
  }
};

// Similar for handleRejectSubmit
const handleRejectSubmit = async (signature: string, reason?: string) => {
  try {
    await rejectMutation.mutateAsync({...});

    toast.success('Task rejected', {
      description: `Rejection reason: ${reason}`,
    })

    setApproveModalOpen(false);
    onApprovalComplete?.();
  } catch (error) {
    toast.error('Failed to reject task')
    console.error("Rejection failed:", error);
  }
};
```

**Step 4: Test**

1. Go to tasks → Approvals
2. Click a task
3. Approve with signature
4. ✅ Should see green "Task approved successfully!" toast
5. ✅ Click "Rejection" button
6. Should see toast for rejection too

---

## Task 2: Add Form Validation (1 hour) ⏱️

### Why

More robust form checking. Currently only checks signature exists.

### How

**Step 1: Install Dependencies**

```bash
npm install react-hook-form zod @hookform/resolvers
```

**Step 2: Create validation schemas**

```typescript
// src/lib/validation-schemas.ts
import { z } from "zod";

export const approveSchema = z.object({
  signature: z.string().min(1, "Signature is required"),
  remarks: z.string().optional(),
});

export const rejectSchema = z.object({
  signature: z.string().min(1, "Signature is required"),
  reason: z
    .string()
    .min(1, "Rejection reason is required")
    .min(10, "Please provide a detailed reason (at least 10 characters)"),
});

export const reassignSchema = z.object({
  newApproverId: z.string().min(1, "Please select an approver"),
  reason: z.string().optional(),
});
```

**Step 3: Update NotificationActionModal**

```typescript
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { approveSchema, rejectSchema } from '@/lib/validation-schemas'

// Inside component
const form = useForm({
  resolver: zodResolver(selectedAction === 'approve' ? approveSchema : rejectSchema),
  defaultValues: {
    signature: '',
    remarks: '',
    reason: '',
  },
})

// Update form field
<form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
  <div>
    <SignatureCanvas
      onSignatureChange={(sig) => form.setValue('signature', sig)}
      isRequired={true}
    />
    {form.formState.errors.signature && (
      <span className="text-red-500 text-sm mt-1">
        {form.formState.errors.signature.message}
      </span>
    )}
  </div>

  <div>
    <Textarea
      placeholder="Enter remarks..."
      value={form.watch('remarks')}
      onChange={(e) => form.setValue('remarks', e.target.value)}
    />
  </div>

  <Button
    type="submit"
    disabled={form.formState.isSubmitting}
  >
    {form.formState.isSubmitting ? 'Processing...' : 'Submit'}
  </Button>
</form>
```

**Step 4: Test**

1. Click Approve → Try to submit without signature
2. ✅ Should see: "Signature is required"
3. Click Reject → Try to submit without reason
4. ✅ Should see: "Rejection reason is required"
5. Enter reason < 10 characters
6. ✅ Should see: "Please provide a detailed reason"

---

## Task 3: Add Error Boundaries (30 min) ⏱️

### Why

Catch unexpected errors and show user-friendly message instead of white screen.

### How

**Step 1: Create error boundary component**

```typescript
// src/components/error-boundary.tsx
'use client'

import { useEffect, useState } from 'react'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { AlertTriangle, RefreshCw } from 'lucide-react'
import { Button } from '@/components/ui/button'

interface ErrorBoundaryProps {
  children: React.ReactNode
  fallback?: (error: Error) => React.ReactNode
}

export function ErrorBoundary({ children, fallback }: ErrorBoundaryProps) {
  const [error, setError] = useState<Error | null>(null)

  useEffect(() => {
    const handleError = (event: ErrorEvent) => {
      setError(event.error)
    }

    const handleUnhandledRejection = (event: PromiseRejectionEvent) => {
      setError(new Error(event.reason?.message || 'Unknown error'))
    }

    window.addEventListener('error', handleError)
    window.addEventListener('unhandledrejection', handleUnhandledRejection)

    return () => {
      window.removeEventListener('error', handleError)
      window.removeEventListener('unhandledrejection', handleUnhandledRejection)
    }
  }, [])

  if (error) {
    return (
      fallback?.(error) || (
        <div className="p-6">
          <Alert variant="destructive">
            <AlertTriangle className="h-4 w-4" />
            <AlertTitle>Something went wrong</AlertTitle>
            <AlertDescription>
              {error.message || 'An unexpected error occurred'}
            </AlertDescription>
          </Alert>
          <Button
            onClick={() => {
              setError(null)
              window.location.reload()
            }}
            className="mt-4"
          >
            <RefreshCw className="mr-2 h-4 w-4" />
            Reload Page
          </Button>
        </div>
      )
    )
  }

  return children
}
```

**Step 2: Wrap components**

```typescript
// src/app/(private)/(main)/tasks/page.tsx
import { ErrorBoundary } from '@/components/error-boundary'

export default function TasksPage() {
  return (
    <ErrorBoundary>
      <TasksClient userId={userId} userRole={userRole} />
    </ErrorBoundary>
  )
}
```

**Step 3: Test**

1. Throw error intentionally to verify boundary catches it
2. Error should show in alert, not crash entire app

---

## Task 4: Add Loading Skeletons (1 hour) ⏱️

### Why

Better visual feedback while data loads. Currently shows nothing while loading.

### How

**Step 1: Create skeleton component**

```typescript
// src/components/ui/skeleton.tsx (already exists with shadcn)
import { Skeleton } from '@/components/ui/skeleton'

// Usage
export function ApprovalCardSkeleton() {
  return (
    <div className="space-y-4">
      {Array.from({ length: 3 }).map((_, i) => (
        <div key={i} className="space-y-3">
          <Skeleton className="h-20 w-full" />
          <Skeleton className="h-4 w-3/4" />
          <Skeleton className="h-4 w-1/2" />
        </div>
      ))}
    </div>
  )
}
```

**Step 2: Update ApprovalsList component**

```typescript
// src/app/(private)/(main)/tasks/_components/approvals-list.tsx
import { useGetApprovalTasks } from '@/hooks/use-approval-queries'
import { ApprovalCardSkeleton } from '@/components/ui/skeleton'

export function ApprovalsList({ userId }: { userId: string }) {
  const { data: tasks, isLoading, error } = useGetApprovalTasks()

  if (isLoading) {
    return <ApprovalCardSkeleton />
  }

  if (error) {
    return (
      <Alert variant="destructive">
        <AlertCircle className="h-4 w-4" />
        <AlertDescription>Failed to load approval tasks</AlertDescription>
      </Alert>
    )
  }

  if (!tasks || tasks.length === 0) {
    return (
      <Alert>
        <Info className="h-4 w-4" />
        <AlertDescription>No pending approvals</AlertDescription>
      </Alert>
    )
  }

  return (
    <div className="space-y-4">
      {tasks.map((task) => (
        <ApprovalCard key={task.id} task={task} />
      ))}
    </div>
  )
}
```

---

## Task 5: Add E2E Tests (2-4 hours) ⏱️

### Why

Automated testing ensures workflows work before shipping to Phase 12.

### How

**Step 1: Setup Jest & React Testing Library**

```bash
npm install --save-dev @testing-library/react @testing-library/jest-dom jest @types/jest
npx jest --init
```

**Step 2: Create test file**

```typescript
// __tests__/approval-workflow.test.tsx
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { TasksClient } from '@/app/(private)/(main)/tasks/_components/tasks-client'

describe('Approval Workflow E2E', () => {
  beforeEach(() => {
    // Clear localStorage
    localStorage.clear()
    // Initialize mock data
    localStorage.setItem('approval_tasks_v1', JSON.stringify({
      'task-1': {
        id: 'task-1',
        entityType: 'REQUISITION',
        entityNumber: 'REQ-2024-001',
        status: 'pending',
        // ... other fields
      },
    }))
  })

  test('Complete approval workflow end-to-end', async () => {
    const user = userEvent.setup()

    // 1. Render component
    render(<TasksClient userId="user-1" userRole="APPROVER" />)

    // 2. Click Approvals tab
    const approvalsTab = screen.getByRole('tab', { name: /approvals/i })
    await user.click(approvalsTab)

    // 3. Verify tasks loaded
    await waitFor(() => {
      expect(screen.getByText(/REQ-2024-001/i)).toBeInTheDocument()
    })

    // 4. Click task
    const taskCard = screen.getByText(/REQ-2024-001/i)
    await user.click(taskCard)

    // 5. Navigate to approval page (in real test, mock router)
    // 6. Verify signature canvas
    expect(screen.getByText(/Digital Signature/i)).toBeInTheDocument()

    // 7. Draw signature (mock canvas)
    const canvas = screen.getByRole('img', { hidden: true })
    // Mock drawing on canvas...

    // 8. Add remarks
    const remarksInput = screen.getByPlaceholderText(/remarks/i)
    await user.type(remarksInput, 'Approved for procurement')

    // 9. Submit approval
    const submitBtn = screen.getByRole('button', { name: /submit approval/i })
    await user.click(submitBtn)

    // 10. Verify success
    await waitFor(() => {
      expect(screen.getByText(/Task approved successfully/i)).toBeInTheDocument()
    })

    // 11. Verify localStorage updated
    const stored = JSON.parse(localStorage.getItem('approval_tasks_v1') || '{}')
    expect(stored['task-1'].status).toBe('approved')
  })

  test('Reject task with required reason', async () => {
    const user = userEvent.setup()

    // Test rejection flow
    // 1. Click reject
    // 2. Try submit without reason → error
    // 3. Add reason
    // 4. Submit → success
  })

  test('Bulk approve tasks', async () => {
    const user = userEvent.setup()

    // Test bulk operations
    // 1. Check multiple items
    // 2. Click Approve All
    // 3. Verify all tasks approved
  })

  test('Data persists across page refresh', async () => {
    // 1. Approve task
    // 2. Simulate page refresh (clear component, remount)
    // 3. Verify task still approved from localStorage
  })
})
```

**Step 3: Run tests**

```bash
npm test
```

---

## Implementation Checklist

### Toast Notifications

- [ ] Install sonner: `npm install sonner`
- [ ] Add Toaster to layout.tsx
- [ ] Add toast.success() after approve
- [ ] Add toast.error() on error
- [ ] Test: Approve task → see green toast
- [ ] Test: Error → see red toast

### Form Validation

- [ ] Install: `npm install react-hook-form zod @hookform/resolvers`
- [ ] Create validation-schemas.ts
- [ ] Update NotificationActionModal with useForm
- [ ] Add error messages below fields
- [ ] Test: Submit without signature → error
- [ ] Test: Submit with reason < 10 chars → error

### Error Boundaries

- [ ] Create error-boundary.tsx
- [ ] Wrap TasksClient in error boundary
- [ ] Test: Throw error → see alert, not crash

### Loading Skeletons

- [ ] Create ApprovalCardSkeleton component
- [ ] Update ApprovalsList to show skeleton while loading
- [ ] Test: Page loading → see skeleton, then content

### E2E Tests

- [ ] Setup Jest and Testing Library
- [ ] Create test files
- [ ] Write 4-5 test cases
- [ ] Run: `npm test`
- [ ] Verify all tests pass

---

## Testing Your Work

### Manual Testing Script

```
1. Clear browser cache/localStorage
2. Go to http://localhost:3000/workflows/tasks
3. Click "Approvals" tab
4. Click any task card
5. Test Approval:
   a. Click Approve button
   b. Try to submit without signature → should error
   c. Draw signature on canvas
   d. Add remarks (optional)
   e. Click Submit Approval
   f. ✅ Should see green toast: "Task approved successfully!"
   g. Go back to list → task shows as approved

6. Test Rejection:
   a. Click another task
   b. Click Reject button
   c. Try to submit without reason → should error
   d. Add reason < 10 chars → should error
   e. Add proper reason
   f. Draw signature
   g. Click Submit Reject
   h. ✅ Should see green toast: "Task rejected"

7. Persistence:
   a. Refresh page (F5)
   b. ✅ Previously approved/rejected tasks still show as such
   c. Close browser completely
   d. Reopen browser, go to page
   e. ✅ Data still persists

8. Analytics:
   a. Go to Admin Reports → Analytics tab
   b. ✅ Metrics updated (1 approved, 1 rejected, 1 pending)

9. Bulk Operations:
   a. Go to Approvals tab
   b. Check boxes on 2 tasks
   c. Click "Approve All"
   d. Dialog shows count
   e. Click Approve
   f. ✅ Green toast shows "Successfully approved 2 tasks"

```

---

## Summary

**What you're adding**:

1. ✅ Toast notifications on success/error
2. ✅ Form validation with better error messages
3. ✅ Error boundaries to catch crashes
4. ✅ Loading skeletons for better UX
5. ✅ E2E tests to verify everything works

**Time estimate**: 1 day (4-8 hours depending on depth)

**Result**: Production-ready E2E workflows ready for Phase 12 database integration

---

**Next**: After completing these tasks, ready for Phase 12 - Database Integration!
