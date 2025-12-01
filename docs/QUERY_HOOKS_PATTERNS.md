# Query and Mutation Hooks Patterns

This document outlines the standardized patterns for creating and using React Query hooks throughout the Liyali Gateway application.

## Table of Contents

1. [Overview](#overview)
2. [Query Key Management](#query-key-management)
3. [Query Hooks Pattern](#query-hooks-pattern)
4. [Mutation Hooks Pattern](#mutation-hooks-pattern)
5. [File Organization](#file-organization)
6. [Examples](#examples)

---

## Overview

We use **React Query** (TanStack Query) for managing server state and caching. All query hooks follow a consistent pattern to ensure:

- **Single Source of Truth**: All QUERY_KEYS are centralized in `src/lib/constants.ts`
- **SSR Support**: Hooks accept `initialData` from server components
- **Consistent Mutations**: Create/Update operations are combined with conditional logic
- **Error Handling**: Standardized toast notifications for user feedback
- **Cache Invalidation**: Automatic query invalidation on mutations

---

## Query Key Management

### Location
All QUERY_KEYS are defined in **`src/lib/constants.ts`**

### Structure
Query keys are organized hierarchically by feature:

```typescript
export const QUERY_KEYS = {
  // Feature Group
  FEATURE_NAME: {
    ALL: "feature-all",           // All items
    BY_ID: "feature-by-id",       // Single item
    BY_USER: "feature-by-user",   // Filtered by user
    STATS: "feature-stats",       // Statistics
  },
};
```

### Adding New Query Keys

When implementing a new feature:

1. Add the feature to `QUERY_KEYS` in `constants.ts`
2. Reference keys using `QUERY_KEYS.FEATURE_NAME.KEY` in hook files
3. Update documentation when adding new keys

---

## Query Hooks Pattern

### Basic Query Hook

```typescript
'use client';

import { useQuery } from '@tanstack/react-query';
import { QUERY_KEYS } from '@/lib/constants';
import { fetchData } from '@/app/_actions/feature';
import { DataType } from '@/types/feature';

/**
 * Fetch all items
 * Static data - rarely changes
 *
 * @param initialData - Optional initial data from server component
 * @returns Query result with items array
 *
 * @example
 * const { data: items } = useItems(initialItems)
 */
export const useItems = (initialData?: DataType[]) =>
  useQuery({
    queryKey: [QUERY_KEYS.FEATURE.ALL],
    queryFn: async () => {
      const response = await fetchData();
      return response.success ? response.data : [];
    },
    initialData,
    staleTime: Infinity, // Never refetch
  });
```

### Query with Parameters

```typescript
/**
 * Fetch items by ID
 *
 * @param id - Item ID to fetch
 * @returns Query result with single item
 *
 * @example
 * const { data: item } = useItemById(id)
 */
export const useItemById = (id: string) =>
  useQuery({
    queryKey: [QUERY_KEYS.FEATURE.BY_ID, id],
    queryFn: async () => {
      const response = await fetchById(id);
      if (!response.success) throw new Error(response.message);
      return response.data;
    },
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
```

### Live Data Query (Frequently Updated)

```typescript
/**
 * Fetch live task data
 * Updates frequently - refetches every 30 seconds
 *
 * @param userId - User ID
 * @param initialData - Optional initial data from server
 * @returns Query result with auto-refetch
 *
 * @example
 * const { data: tasks } = useLiveTasks(userId, initialTasks)
 */
export const useLiveTasks = (userId: string, initialData?: Task[]) =>
  useQuery({
    queryKey: [QUERY_KEYS.TASKS.BY_USER, userId],
    queryFn: async () => {
      const response = await getTasks(userId);
      return response.success ? response.data : [];
    },
    initialData,
    staleTime: 0,                  // Always stale
    refetchInterval: 30 * 1000,    // Refetch every 30 seconds
  });
```

### Key Points for Query Hooks

- **Always use `QUERY_KEYS` from constants**
- **Include JSDoc comments** with @param, @returns, @example
- **Accept `initialData`** for SSR support
- **Handle errors** by throwing them so React Query captures them
- **Use appropriate `staleTime`**:
  - `Infinity`: Static data (departments, roles, etc.)
  - `5 * 60 * 1000`: Standard data (budgets, requisitions)
  - `0`: Live data (tasks, activities)

---

## Mutation Hooks Pattern

### Combined Create/Update Mutation

```typescript
'use client';

import { useMutation, useQueryClient } from '@tanstack/react-query';
import { QUERY_KEYS } from '@/lib/constants';
import { createItem, updateItem } from '@/app/_actions/feature';
import { CreateRequest, UpdateRequest } from '@/types/feature';
import { toast } from 'sonner';

/**
 * Create or update item mutation
 * Handles both create (no ID) and update (with ID) operations
 *
 * @param onSuccess - Callback after successful mutation
 * @returns Mutation object with mutate and mutateAsync
 *
 * @example
 * const saveMutation = useSaveItem(() => {
 *   queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.FEATURE.ALL] })
 * })
 *
 * // Create
 * await saveMutation.mutateAsync({ name: 'New Item' })
 *
 * // Update
 * await saveMutation.mutateAsync({ id: 'item-1', name: 'Updated Item' })
 */
export const useSaveItem = (onSuccess?: () => void) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: CreateRequest | (UpdateRequest & { id?: string })) => {
      // Determine if creating or updating based on presence of ID
      const response = 'id' in data && data.id
        ? await updateItem(data as UpdateRequest)
        : await createItem(data as CreateRequest);

      if (!response.success) {
        throw new Error(response.message);
      }
      return response;
    },
    onSuccess: (response) => {
      const isUpdate = (response.data as any)?.id;
      toast.success(isUpdate ? 'Item updated successfully' : 'Item created successfully');

      // Invalidate related queries
      queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.FEATURE.ALL] });
      queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.FEATURE.STATS] });

      onSuccess?.();
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to save item');
    },
  });
};
```

### Specific Action Mutation

```typescript
/**
 * Approve item mutation
 *
 * @param itemId - Item ID to approve
 * @param onSuccess - Callback after successful approval
 * @returns Mutation object
 *
 * @example
 * const approveMutation = useApproveItem(itemId)
 * await approveMutation.mutateAsync({
 *   approvingUserId: userId,
 *   approvingUserRole: 'APPROVER',
 *   signature: signatureDataUrl,
 *   comments: 'Approved'
 * })
 */
export const useApproveItem = (itemId: string, onSuccess?: () => void) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: Omit<ApproveRequest, 'itemId'>) => {
      const response = await approveItem({ itemId, ...data });

      if (!response.success) {
        throw new Error(response.message);
      }
      return response;
    },
    onSuccess: () => {
      toast.success('Item approved successfully');

      // Invalidate all related queries
      queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.FEATURE.BY_ID, itemId] });
      queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.FEATURE.ALL] });
      queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.APPROVALS_PENDING] });

      onSuccess?.();
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to approve item');
    },
  });
};
```

### Delete Mutation

```typescript
/**
 * Delete item mutation
 *
 * @param itemId - Item ID to delete
 * @param onSuccess - Callback after successful deletion
 * @returns Mutation object
 *
 * @example
 * const deleteMutation = useDeleteItem(itemId)
 * await deleteMutation.mutateAsync()
 */
export const useDeleteItem = (itemId: string, onSuccess?: () => void) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async () => {
      const response = await deleteItem(itemId);

      if (!response.success) {
        throw new Error(response.message);
      }
      return response;
    },
    onSuccess: () => {
      toast.success('Item deleted successfully');

      queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.FEATURE.ALL] });
      queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.FEATURE.STATS] });

      onSuccess?.();
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to delete item');
    },
  });
};
```

### Key Points for Mutation Hooks

- **Always invalidate related queries** after successful mutation
- **Use conditional logic** to combine create/update when appropriate
- **Include error handling** with `try-catch` or `throw`
- **Provide success/error toast notifications**
- **Omit IDs from parameters** when IDs are already known (e.g., `useApproveItem(id)`)
- **Always include JSDoc comments**

---

## File Organization

### Structure

```
src/hooks/
├── use-query-data.ts              # Base hooks (useQueryData, useServerData, etc.)
├── use-budget-queries.ts          # Budget queries and mutations
├── use-task-queries.ts            # Task queries only
├── use-requisition-queries.ts     # Requisition queries and mutations
└── use-[feature]-queries.ts       # One file per feature
```

### Naming Conventions

- **File names**: `use-[feature]-queries.ts`
- **Query hooks**: `use[Feature]` or `use[Feature]ById`, `use[Feature]ByUser`
- **Mutation hooks**: `use[Action][Feature]` (e.g., `useSaveBudget`, `useApproveBudget`)
- **Convenience hooks**: `usePending[Feature]`, `useCompleted[Feature]`

---

## Examples

### Example 1: Using Query Hook in Component

```typescript
'use client';

import { useBudgets } from '@/hooks/use-budget-queries';
import { Spinner } from '@/components/ui/spinner';
import { ErrorMessage } from '@/components/ui/error-message';

interface BudgetsClientProps {
  userId: string;
  initialBudgets?: Budget[];
}

export function BudgetsClient({ userId, initialBudgets }: BudgetsClientProps) {
  // Server component passes initialBudgets, client component uses them
  const { data: budgets, isLoading, error } = useBudgets(userId, initialBudgets);

  if (isLoading) return <Spinner />;
  if (error) return <ErrorMessage message={error.message} />;

  return (
    <div className="space-y-4">
      {budgets?.map((budget) => (
        <BudgetCard key={budget.id} budget={budget} />
      ))}
    </div>
  );
}
```

### Example 2: Using Mutation Hook in Component

```typescript
'use client';

import { useSaveBudget } from '@/hooks/use-budget-queries';
import { useRouter } from 'next/navigation';

export function BudgetForm() {
  const router = useRouter();
  const saveMutation = useSaveBudget(() => {
    router.push('/budgets');
  });

  const handleSubmit = async (formData: CreateBudgetRequest) => {
    try {
      // Create new budget
      await saveMutation.mutateAsync(formData);
    } catch (error) {
      // Error handled by mutation's onError
    }
  };

  return (
    <form onSubmit={handleSubmit}>
      {/* Form fields */}
      <button
        type="submit"
        disabled={saveMutation.isPending}
      >
        {saveMutation.isPending ? 'Saving...' : 'Save Budget'}
      </button>
    </form>
  );
}
```

### Example 3: Server Component with Initial Data

```typescript
// app/(private)/budgets/page.tsx
import { getBudgets } from '@/app/_actions/budgets';
import { BudgetsClient } from './_components/budgets-client';

export default async function BudgetsPage() {
  const session = await auth();

  // Fetch initial data on server
  const result = await getBudgets(session.user.id);

  return (
    <BudgetsClient
      userId={session.user.id}
      initialBudgets={result.success ? result.data : undefined}
    />
  );
}
```

---

## Best Practices

### Do's ✅

- ✅ Keep all QUERY_KEYS in `constants.ts`
- ✅ Use JSDoc comments for all hooks
- ✅ Include examples in comments
- ✅ Provide `initialData` support for SSR
- ✅ Invalidate related queries after mutations
- ✅ Use appropriate stale times
- ✅ Handle errors with user-friendly messages
- ✅ Combine create/update when logic is similar
- ✅ Create one hook file per feature

### Don'ts ❌

- ❌ Don't hardcode query keys in components
- ❌ Don't forget to invalidate queries after mutations
- ❌ Don't forget error handling
- ❌ Don't mix hooks from multiple features in one file
- ❌ Don't forget JSDoc comments
- ❌ Don't create unnecessary hooks for one-time operations
- ❌ Don't forget to return errors properly

---

## Common Query Key Patterns

```typescript
// All items
[QUERY_KEYS.FEATURE.ALL]

// Specific item
[QUERY_KEYS.FEATURE.BY_ID, id]

// Filtered by user
[QUERY_KEYS.FEATURE.BY_USER, userId]

// With multiple parameters
[QUERY_KEYS.FEATURE.BY_USER, userId, status]

// Statistics
[QUERY_KEYS.FEATURE.STATS, userId]
```

---

## Invalidation Patterns

### After Create/Update
```typescript
queryClient.invalidateQueries({
  queryKey: [QUERY_KEYS.FEATURE.ALL]
});
queryClient.invalidateQueries({
  queryKey: [QUERY_KEYS.FEATURE.STATS]
});
```

### After Approval
```typescript
queryClient.invalidateQueries({
  queryKey: [QUERY_KEYS.FEATURE.BY_ID, itemId]
});
queryClient.invalidateQueries({
  queryKey: [QUERY_KEYS.APPROVALS_PENDING]
});
```

### After Delete
```typescript
queryClient.invalidateQueries({
  queryKey: [QUERY_KEYS.FEATURE.ALL]
});
queryClient.invalidateQueries({
  queryKey: [QUERY_KEYS.FEATURE.STATS]
});
```

---

## Migration Guide

### Before (Old Pattern)
```typescript
const { data } = useQuery({
  queryKey: ['budgets'],  // ❌ Hardcoded key
  queryFn: () => getBudgets(),
  staleTime: 5 * 60 * 1000,
});
```

### After (New Pattern)
```typescript
const { data } = useBudgets(userId, initialBudgets);  // ✅ Uses QUERY_KEYS, supports SSR
```

---

## Troubleshooting

### Query Not Updating After Mutation

**Problem**: Data doesn't reflect after mutation even though it succeeded.

**Solution**: Ensure `queryClient.invalidateQueries()` is called in the mutation's `onSuccess`.

### Initial Data Not Showing

**Problem**: Initial data from server component is not displayed.

**Solution**: Make sure you're passing `initialData` to the hook and the data structure matches.

### Too Many Refetches

**Problem**: Queries are refetching too frequently.

**Solution**: Adjust `staleTime` value. For static data, use `Infinity`.

---

## References

- [React Query Documentation](https://tanstack.com/query/latest)
- [QUERY_KEYS Constants](../src/lib/constants.ts)
- [Use Query Data Hook](../src/hooks/use-query-data.ts)
- [Budget Queries Example](../src/hooks/use-budget-queries.ts)

---

**Last Updated**: 2025-11-30
**Version**: 1.0.0
