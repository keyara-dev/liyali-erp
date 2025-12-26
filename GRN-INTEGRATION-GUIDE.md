# GRN (Goods Received Notes) Integration Guide

**Status**: ✅ COMPLETE
**Date**: 2025-12-26
**Branch**: feat/go-fiber

---

## Overview

The GRN module has been fully integrated with the backend API using React Query hooks and server actions. All GRN operations now properly call the backend instead of using localStorage as the primary source.

---

## Architecture

### Data Flow

```
Backend API (/api/v1/grns/*)
    ↓
Server Actions (grn-actions.ts)
    ├─ authenticatedApiClient for API calls
    ├─ Error handling via handleError()
    └─ Response wrapping via successResponse()
    ↓
React Query Hooks (use-grn-queries.ts)
    ├─ Query hooks: useGRNs(), useGRNById()
    ├─ Mutation hooks: useCreateGRN(), useUpdateGRN(), etc.
    ├─ Toast notifications (sonner)
    └─ Automatic query invalidation
    ↓
Components
    └─ useGRNById(), useUpdateGRN(), etc.
```

---

## Files Changed

### 1. **Server Actions** (`frontend/src/app/_actions/grn-actions.ts`)

**All functions now call the backend API using `authenticatedApiClient`:**

#### Query Functions
- `getGRNAction(grnId)` - GET `/api/v1/grns/{id}`
- `getGRNsAction(page, limit, filters)` - GET `/api/v1/grns?page=...&limit=...`

#### Create/Update Functions
- `createGRNAction(poNumber, items, receivedBy, ...)` - POST `/api/v1/grns`
- `updateGRNAction(grnId, updates)` - PUT `/api/v1/grns/{id}`

#### Quality Issue Functions
- `addQualityIssueToGRN(grnId, issue)` - Adds to qualityIssues array
- `removeQualityIssueFromGRN(grnId, issueId)` - Removes from qualityIssues array
- `updateQualityIssueInGRN(grnId, issueId, updates)` - Updates in qualityIssues array

#### Approval Workflow Functions
- `approveGRNAction(grnId, signature, comments)` - POST `/api/v1/grns/{id}/approve`
- `rejectGRNAction(grnId, signature, remarks)` - POST `/api/v1/grns/{id}/reject`
- `confirmGRNAction(grnId)` - POST `/api/v1/grns/{id}/confirm`

#### Delete Function
- `deleteGRNAction(grnId)` - DELETE `/api/v1/grns/{id}` (DRAFT only)

**Return Type**: All functions return `APIResponse<T>` with proper error handling.

---

### 2. **Query Hooks** (`frontend/src/hooks/use-grn-queries.ts`)

#### Query Hooks (Read-only)

**`useGRNs(page, limit, filters)`**
```typescript
const { data: grns, isLoading, error, refetch } = useGRNs(1, 10, {
  status: 'DRAFT',
  poNumber: 'PO-123'
});
```
- Fetches all GRNs with pagination
- 5-minute cache time
- Supports filtering by status and PO number

**`useGRNById(grnId, initialData)`**
```typescript
const { data: grn, isLoading, error } = useGRNById(grnId);
```
- Fetches a single GRN
- Optional initial data from server
- Enabled only when grnId is provided

#### Mutation Hooks (Write operations)

**`useCreateGRN(onSuccess)`**
```typescript
const { mutateAsync: createGRN, isPending, error } = useCreateGRN((grn) => {
  console.log('Created:', grn);
});

await createGRN({
  poNumber: 'PO-123',
  items: [...],
  receivedBy: 'user-id',
  warehouseLocation: 'Warehouse A',
  notes: 'Optional notes'
});
```

**`useUpdateGRN(grnId, onSuccess)`**
```typescript
const { mutateAsync: updateGRN, isPending } = useUpdateGRN(grnId);

await updateGRN({
  items: [...],
  qualityIssues: [...],
  notes: 'Updated notes'
});
```

**`useApproveGRN(grnId, onSuccess)`**
```typescript
const { mutateAsync: approveGRN, isPending } = useApproveGRN(grnId);

await approveGRN({
  signature: 'user-signature-data',
  comments: 'Approved as is'
});
```

**`useRejectGRN(grnId, onSuccess)`**
```typescript
const { mutateAsync: rejectGRN, isPending } = useRejectGRN(grnId);

await rejectGRN({
  signature: 'user-signature-data',
  remarks: 'Items do not match PO specifications' // min 10 chars
});
```

**`useConfirmGRN(grnId, onSuccess)`**
```typescript
const { mutateAsync: confirmGRN, isPending } = useConfirmGRN(grnId);

await confirmGRN();
```

**`useDeleteGRN(grnId, onSuccess)`**
```typescript
const { mutateAsync: deleteGRN, isPending } = useDeleteGRN(grnId);

await deleteGRN();
```

#### Quality Issue Mutations (`use-grn-mutations.ts`)

**`useAddQualityIssueMutation(grnId, onSuccess)`**
```typescript
const { addIssue, isPending, error } = useAddQualityIssueMutation(grnId);

await addIssue({
  itemId: 'item-123',
  description: 'Damaged corner',
  severity: 'MEDIUM'
});
```

**`useRemoveQualityIssueMutation(grnId, onSuccess)`**
```typescript
const { removeIssue, isPending } = useRemoveQualityIssueMutation(grnId);

await removeIssue('issue-123');
```

**`useUpdateQualityIssueMutation(grnId, onSuccess)`**
```typescript
const { updateIssue, isPending } = useUpdateQualityIssueMutation(grnId);

await updateIssue({
  issueId: 'issue-123',
  updates: { severity: 'HIGH' }
});
```

---

## Key Features

### ✅ Automatic Query Management
- Toast notifications on success/error
- Automatic query invalidation
- Cache updates for optimistic UI

### ✅ Error Handling
- Consistent error responses via `handleError()`
- User-friendly error messages
- Network error detection

### ✅ Authentication
- Uses `authenticatedApiClient` for automatic token injection
- Includes organization header if available
- Handles session verification

### ✅ Type Safety
- Full TypeScript support
- `GoodsReceivedNote` interface exported
- `APIResponse<T>` for all server actions

---

## GRN Data Model

```typescript
interface GoodsReceivedNote {
  id: string;
  grnNumber: string;
  poNumber: string;
  status: 'DRAFT' | 'SUBMITTED' | 'CONFIRMED' | 'REJECTED' | 'APPROVED';
  warehouseLocation: string;
  receivedDate: string;
  receivedBy: string;
  approvedBy?: string;
  items: GRNItem[];
  qualityIssues: QualityIssue[];
  notes?: string;
  currentStage: number;
  stageName: string;
  createdAt: string;
  updatedAt: string;
}

interface GRNItem {
  id: string;
  itemNumber: number;
  description: string;
  poQuantity: number;
  receivedQuantity: number;
  unit: string;
  variance: number;
  damage: number;
  damageNotes?: string;
  condition: 'GOOD' | 'DAMAGED' | 'PARTIAL';
}

interface QualityIssue {
  id: string;
  itemId: string;
  description: string;
  severity: 'LOW' | 'MEDIUM' | 'HIGH';
}
```

---

## GRN Workflow States

```
DRAFT → SUBMITTED → APPROVED → CONFIRMED
                  ↓
                REJECTED
```

### Status Descriptions
- **DRAFT**: Initial state, can be edited or deleted
- **SUBMITTED**: Submitted for approval, cannot be edited
- **APPROVED**: Approved by authorized user
- **CONFIRMED**: Receipt confirmed, workflow complete
- **REJECTED**: Rejected during approval, cannot proceed

---

## Usage Examples

### Example 1: Creating a GRN from a PO

```typescript
'use client';

import { useCreateGRN } from '@/hooks/use-grn-queries';
import { useSession } from '@/hooks/use-session';

export function CreateGRNButton({ poNumber, items }) {
  const { user } = useSession();
  const { mutateAsync: createGRN, isPending } = useCreateGRN((newGrn) => {
    console.log('GRN created:', newGrn.grnNumber);
    // Navigate to detail page, show success, etc.
  });

  const handleCreate = async () => {
    try {
      await createGRN({
        poNumber,
        items,
        receivedBy: user?.id || '',
        warehouseLocation: 'Main Warehouse'
      });
    } catch (error) {
      console.error('Failed to create GRN:', error);
    }
  };

  return (
    <button onClick={handleCreate} disabled={isPending}>
      {isPending ? 'Creating...' : 'Create GRN'}
    </button>
  );
}
```

### Example 2: Viewing GRN Details with Quality Issues

```typescript
'use client';

import { useGRNById, useUpdateGRN } from '@/hooks/use-grn-queries';
import { useAddQualityIssueMutation } from '@/hooks/use-grn-mutations';

export function GRNDetailPage({ grnId }: { grnId: string }) {
  const { data: grn, isLoading } = useGRNById(grnId);
  const { addIssue, isPending: isAddingIssue } = useAddQualityIssueMutation(grnId);

  if (isLoading) return <div>Loading...</div>;

  const handleAddQualityIssue = async () => {
    await addIssue({
      itemId: 'item-1',
      description: 'Item damaged in transit',
      severity: 'HIGH'
    });
  };

  return (
    <div>
      <h1>{grn?.grnNumber}</h1>
      <p>PO: {grn?.poNumber}</p>
      <p>Status: {grn?.status}</p>

      <h2>Items ({grn?.items.length})</h2>
      {grn?.items.map(item => (
        <div key={item.id}>
          <p>{item.description}</p>
          <p>Received: {item.receivedQuantity} / {item.poQuantity}</p>
        </div>
      ))}

      <h2>Quality Issues ({grn?.qualityIssues.length})</h2>
      {grn?.qualityIssues.map(issue => (
        <div key={issue.id}>
          <p>{issue.description} ({issue.severity})</p>
        </div>
      ))}

      <button onClick={handleAddQualityIssue} disabled={isAddingIssue}>
        Add Quality Issue
      </button>
    </div>
  );
}
```

### Example 3: Approval Workflow

```typescript
'use client';

import { useGRNById, useApproveGRN, useRejectGRN } from '@/hooks/use-grn-queries';
import { useState } from 'react';

export function GRNApprovalPage({ grnId }: { grnId: string }) {
  const { data: grn } = useGRNById(grnId);
  const { mutateAsync: approveGRN, isPending: isApproving } = useApproveGRN(grnId);
  const { mutateAsync: rejectGRN, isPending: isRejecting } = useRejectGRN(grnId);
  const [signature, setSignature] = useState('');
  const [remarks, setRemarks] = useState('');

  const handleApprove = async () => {
    await approveGRN({
      signature,
      comments: 'Approved'
    });
  };

  const handleReject = async () => {
    await rejectGRN({
      signature,
      remarks
    });
  };

  return (
    <div>
      <h1>Review GRN: {grn?.grnNumber}</h1>

      <div>
        <input
          type="text"
          placeholder="Enter your signature"
          value={signature}
          onChange={(e) => setSignature(e.target.value)}
        />
      </div>

      <button onClick={handleApprove} disabled={isApproving || !signature}>
        Approve
      </button>

      <div>
        <textarea
          placeholder="Rejection remarks (min 10 characters)"
          value={remarks}
          onChange={(e) => setRemarks(e.target.value)}
        />
      </div>

      <button onClick={handleReject} disabled={isRejecting || remarks.length < 10 || !signature}>
        Reject
      </button>
    </div>
  );
}
```

---

## Migration from localStorage

### Before (localStorage-based)
```typescript
// Old pattern - used localStorage directly
const grns = JSON.parse(localStorage.getItem('app_grns') || '[]');
```

### After (Backend API)
```typescript
// New pattern - uses React Query hooks
const { data: grns, isLoading } = useGRNs(1, 10);
```

---

## Backend API Endpoints

The following backend endpoints are used:

| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/api/v1/grns` | List GRNs with pagination |
| GET | `/api/v1/grns/{id}` | Get single GRN |
| POST | `/api/v1/grns` | Create new GRN |
| PUT | `/api/v1/grns/{id}` | Update GRN |
| DELETE | `/api/v1/grns/{id}` | Delete draft GRN |
| POST | `/api/v1/grns/{id}/approve` | Approve GRN |
| POST | `/api/v1/grns/{id}/reject` | Reject GRN |
| POST | `/api/v1/grns/{id}/confirm` | Confirm receipt |

---

## Query Keys

All GRN queries use the following keys:

```typescript
QUERY_KEYS.GRN.ALL        // 'grn-all' - for useGRNs()
QUERY_KEYS.GRN.BY_ID      // 'grn-by-id' - for useGRNById()
QUERY_KEYS.GRN.BY_USER    // 'grn-by-user' - reserved for future use
```

---

## Error Handling

All mutations include built-in error handling with toast notifications:

```typescript
// Automatic error handling:
// ❌ Network errors → "Please check your internet connection"
// ❌ 404 errors → "Resource not found"
// ❌ Validation errors → Server message displayed
```

---

## Future Enhancements

### TODO
- [ ] Offline sync - sync localStorage changes to backend when online
- [ ] GRN list pagination UI component
- [ ] GRN approval workflow UI
- [ ] Quality issues bulk management
- [ ] GRN PDF export/download
- [ ] GRN search and filtering

---

## Testing

### Test the integration:

1. **Create a GRN**
   ```bash
   POST /api/v1/grns
   Body: { poNumber, items, receivedBy }
   ```

2. **Fetch a GRN**
   ```bash
   GET /api/v1/grns/{id}
   ```

3. **Update GRN quality issues**
   ```bash
   PUT /api/v1/grns/{id}
   Body: { qualityIssues: [...] }
   ```

4. **Approve GRN**
   ```bash
   POST /api/v1/grns/{id}/approve
   Body: { signature, comments }
   ```

---

## Summary

✅ **Completed**
- GRN server actions fully integrated with backend API
- React Query hooks for all CRUD operations
- Toast notifications for user feedback
- Automatic query invalidation
- Full TypeScript support
- Error handling and validation
- Authentication via authenticatedApiClient

✅ **Key Benefits**
- No more localStorage for GRN data
- Real-time backend synchronization
- Automatic caching and performance optimization
- Consistent error handling
- Reusable hooks across components

📝 **Pattern Established**
All future GRN development should follow the:
Server Action → React Query Hook → Component pattern

---

**Maintained By**: Claude Code
**Status**: ✅ Production Ready
**Last Updated**: 2025-12-26

