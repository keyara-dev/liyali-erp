# Development Guide

## Adding a New Workflow Type

### Step 1: Create Type Definition

Add to `src/types/tasks.ts`:

```typescript
interface NewWorkflowTask extends ApprovalTask {
  specificField: string;
  // ... other fields
}
```

### Step 2: Create Pages

Create folder structure:

```
src/app/(private)/(main)/new-workflow/[id]/
├── page.tsx
├── _components/
│   └── new-workflow-detail-client.tsx
└── approval/
    ├── page.tsx
    └── _components/
        └── new-workflow-approval-client.tsx
```

### Step 3: Create Detail Component

```typescript
// src/app/(private)/(main)/new-workflow/[id]/_components/new-workflow-detail-client.tsx

'use client'

import { ApprovalActionPanel } from '@/components/workflows'

export function NewWorkflowDetailClient({ id }: { id: string }) {
  return (
    <div className="space-y-6">
      <h1>New Workflow Details</h1>
      {/* Display fields */}
    </div>
  )
}
```

### Step 4: Create Approval Component

```typescript
// Similar pattern but with ApprovalActionPanel
// Uses useApproveTaskMutation, useRejectTaskMutation, etc.
```

### Step 5: Update Stores

Add mock data to `src/lib/approval-store.ts`:

```typescript
const INITIAL_TASKS = [
  // ... existing tasks
  {
    id: "new-1",
    entityType: "NEW_WORKFLOW",
    entityNumber: "NEW-2024-001",
    // ... other fields
  },
];
```

### Step 6: Update Routes

Add to navigation in `src/components/layout/sidebar/nav-main.tsx`

## Server Action Pattern

All server actions follow this structure:

```typescript
"use server";

interface ActionRequest {
  field1: string;
  field2: number;
}

interface ActionResponse {
  success: boolean;
  data?: any;
  error?: string;
}

export async function myServerAction(
  request: ActionRequest
): Promise<ActionResponse> {
  try {
    // 1. Validate
    if (!request.field1) {
      return { success: false, error: "field1 required" };
    }

    // 2. Process (Phase 11: store, Phase 12: database)
    const result = approvalStore.operation(request);

    // 3. Log
    console.log("[ACTION]", result);

    // 4. Return
    return { success: true, data: result };
  } catch (error) {
    console.error("[ERROR]", error);
    return { success: false, error: "Operation failed" };
  }
}
```

## React Query Hook Pattern

### Query Hook

```typescript
export function useMyQuery(id: string) {
  return useQuery({
    queryKey: ["namespace", id],
    queryFn: async () => {
      const result = await getMyData({ id });
      if (!result.success) throw new Error(result.error);
      return result.data;
    },
    refetchInterval: 30000, // Auto-refresh every 30s
  });
}
```

### Mutation Hook

```typescript
export function useMyMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: myServerAction,
    onSuccess: (data) => {
      // Invalidate related queries
      queryClient.invalidateQueries({
        queryKey: ["namespace"],
      });

      // Show feedback
      toast.success("Success!");
    },
    onError: (error) => {
      toast.error(error.message);
    },
  });
}
```

## Component Structure

### Page Component (Server)

```typescript
// src/app/(private)/(main)/type/[id]/page.tsx

import { getWorkflowData } from '@/actions'
import { DetailClient } from './_components/detail-client'

export default async function Page({ params }: { params: { id: string } }) {
  const data = await getWorkflowData(params.id)

  return <DetailClient data={data} />
}
```

### Detail Component (Client)

```typescript
'use client'

export function DetailClient({ data }: { data: any }) {
  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle>{data.name}</CardTitle>
        </CardHeader>
        <CardContent>
          {/* Display details */}
        </CardContent>
      </Card>
    </div>
  )
}
```

### Approval Component (Client)

```typescript
'use client'

import { ApprovalActionPanel } from '@/components/workflows'
import { useApproveTaskMutation } from '@/hooks'

export function ApprovalClient({ taskId }: { taskId: string }) {
  const approveMutation = useApproveTaskMutation()

  const handleApprove = async (signature: string, remarks: string) => {
    await approveMutation.mutateAsync({
      taskId,
      signature,
      remarks,
    })
  }

  return (
    <ApprovalActionPanel
      onApprove={handleApprove}
      isLoading={approveMutation.isPending}
    />
  )
}
```

## Type Definition Pattern

```typescript
// src/types/tasks.ts

interface WorkflowTask {
  id: string;
  entityId: string;
  entityType: "REQUISITION" | "BUDGET" | "PO" | "PV" | "GRN";
  status: "pending" | "approved" | "rejected";
  stageName: string;
  stageIndex: number;
  // ... more fields
}

// For specific workflows
interface PurchaseOrderTask extends WorkflowTask {
  vendorName: string;
  totalAmount: number;
  itemCount: number;
}
```

## Testing Your Changes

### Manual Testing

1. Add route to sidebar navigation
2. Create sample data in approval-store.ts
3. Navigate to page
4. Test all actions (approve, reject, reassign)
5. Check localStorage for updates
6. Refresh page - data should persist

### Build Testing

```bash
npm run build        # Check for TypeScript errors
npm run type-check   # Verify types
npm run dev          # Test in development
```

## Phase 12 Migration Notes

When migrating to database:

1. Server actions stay the same
2. Replace `approvalStore.operation()` with Prisma queries
3. Update TODO comments in server actions
4. Test with real database

Example migration:

```typescript
// Phase 11 (current)
const result = approvalStore.approveTask(request);

// Phase 12 (database)
const result = await db.approvalTask.update({
  where: { id: request.taskId },
  data: { status: "approved" },
});
```

## Debugging Tips

1. **Open DevTools**: F12 → Application → Local Storage
2. **Check Console**: F12 → Console for errors
3. **Inspect Network**: F12 → Network tab
4. **React DevTools**: Browser extension for component inspection
5. **Check Logs**: Server action console.log output
