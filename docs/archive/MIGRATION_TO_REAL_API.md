# Migration Guide: From Mock Data to Real API

## Overview

This document provides a step-by-step migration path from the current mock/server action implementation to a real backend API.

**Current State**: Server actions + in-memory documentStore
**Target State**: Real HTTP API + React Query caching

**Effort**: ~4-6 hours
**Risk**: Low (architecture is already API-ready)

---

## Phase 1: Create HTTP Client (1 hour)

### Step 1.1: Install Dependencies

```bash
npm install axios  # Or fetch if using native
```

### Step 1.2: Create API Client

```typescript
// src/lib/api/http-client.ts
import axios from 'axios';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:3001/api';

export const httpClient = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Add request interceptor for auth tokens
httpClient.interceptors.request.use((config) => {
  const token = localStorage.getItem('auth_token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Add response interceptor for error handling
httpClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // Handle unauthorized - redirect to login
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);
```

### Step 1.3: Update API Client

```typescript
// src/lib/api/client.ts - REPLACE the entire file

import { httpClient } from './http-client';
import { APIResponse } from '@/types';
import { Requisition, RequisitionStats } from '@/types/requisition';
import { PurchaseOrder, PurchaseOrderStats } from '@/types/purchase-order';
import { PaymentVoucher, PaymentVoucherStats } from '@/types/payment-voucher';

export const apiClient = {
  requisitions: {
    getAll: async (): Promise<APIResponse<Requisition[]>> => {
      const { data } = await httpClient.get('/requisitions');
      return data;
    },

    getById: async (id: string): Promise<APIResponse<Requisition>> => {
      const { data } = await httpClient.get(`/requisitions/${id}`);
      return data;
    },

    create: async (data) => {
      const { data: response } = await httpClient.post('/requisitions', data);
      return response;
    },

    update: async (id: string, data) => {
      const { data: response } = await httpClient.put(`/requisitions/${id}`, data);
      return response;
    },

    submit: async (id: string, data) => {
      const { data: response } = await httpClient.post(`/requisitions/${id}/submit`, data);
      return response;
    },

    approve: async (id: string, data) => {
      const { data: response } = await httpClient.post(`/requisitions/${id}/approve`, data);
      return response;
    },

    reject: async (id: string, data) => {
      const { data: response } = await httpClient.post(`/requisitions/${id}/reject`, data);
      return response;
    },

    delete: async (id: string) => {
      const { data: response } = await httpClient.delete(`/requisitions/${id}`);
      return response;
    },

    getStats: async (): Promise<APIResponse<RequisitionStats>> => {
      const { data } = await httpClient.get('/requisitions/stats');
      return data;
    },
  },

  purchaseOrders: {
    // Similar pattern for POs
  },

  paymentVouchers: {
    // Similar pattern for PVs
  },
};
```

---

## Phase 2: Update Server Actions (1.5 hours)

### Step 2.1: Replace documentStore with API

```typescript
// src/app/_actions/requisitions.ts - MODIFY

import { cache } from 'react';
import { apiClient } from '@/lib/api/client';

export const getRequisitions = cache(async () => {
  try {
    // REMOVE: const workflowDocs = Array.from(documentStore.values());
    // REPLACE WITH:
    return await apiClient.requisitions.getAll();
  } catch (error) {
    console.error('Failed to fetch requisitions:', error);
    // Fallback to localStorage cache
    try {
      const cached = localStorage.getItem('liyali_requisitions_cache');
      return cached
        ? {
            success: true,
            data: JSON.parse(cached),
            message: 'Loaded from cache (offline)',
          }
        : { success: false, data: null, message: 'Failed to fetch requisitions' };
    } catch {
      return { success: false, data: null, message: 'Failed to fetch requisitions' };
    }
  }
});

export async function createRequisition(data) {
  try {
    // REMOVE: const requisition = { ...data, id: `req-${Date.now()}` };
    // REPLACE WITH:
    return await apiClient.requisitions.create(data);
  } catch (error) {
    return {
      success: false,
      message: error instanceof Error ? error.message : 'Failed to create requisition',
      data: null,
    };
  }
}

// Apply same pattern to update, submit, approve, reject, delete
```

### Step 2.2: Delete Mock Data

```typescript
// REMOVE from requisitions.ts:
// let mockRequisitions: Requisition[] = [ ... ]

// REMOVE from purchase-orders.ts:
// let mockPurchaseOrders: PurchaseOrder[] = [ ... ]

// REMOVE from payment-vouchers.ts:
// let mockPaymentVouchers: PaymentVoucher[] = [ ... ]
```

---

## Phase 3: Update Offline Queue Processor (1 hour)

```typescript
// src/hooks/use-offline-queue-processor.ts - MODIFY processQueue function

const processQueue = async () => {
  // ... existing code ...

  for (const operation of operations) {
    try {
      await updateOperationStatus(operation.id, 'processing');

      // REPLACE: TODO comment with actual execution
      let result;

      switch (operation.entity) {
        case 'requisition':
          switch (operation.type) {
            case 'CREATE':
              result = await apiClient.requisitions.create(operation.data);
              break;
            case 'UPDATE':
              result = await apiClient.requisitions.update(
                operation.entityId!,
                operation.data
              );
              break;
            case 'APPROVE':
              result = await apiClient.requisitions.approve(
                operation.entityId!,
                operation.data
              );
              break;
            case 'REJECT':
              result = await apiClient.requisitions.reject(
                operation.entityId!,
                operation.data
              );
              break;
            case 'SUBMIT':
              result = await apiClient.requisitions.submit(
                operation.entityId!,
                operation.data
              );
              break;
            case 'DELETE':
              result = await apiClient.requisitions.delete(operation.entityId!);
              break;
          }
          break;

        case 'purchase-order':
          // Similar pattern for POs
          break;

        case 'payment-voucher':
          // Similar pattern for PVs
          break;
      }

      if (result?.success) {
        await updateOperationStatus(operation.id, 'completed', result.data);
        await removeOperation(operation.id);
        successCount++;
      } else {
        throw new Error(result?.message || 'Unknown error');
      }
    } catch (error) {
      // ... existing error handling ...
    }
  }
};
```

---

## Phase 4: Testing (1.5 hours)

### Step 4.1: Integration Test

```typescript
// Test offline → online sync
describe('Offline Sync', () => {
  it('should queue operations when offline and sync when online', async () => {
    // 1. Go offline
    simulateOffline();

    // 2. Create requisition (should queue)
    const result = await createRequisition({ title: 'Test' });
    expect(result.queued).toBe(true);

    // 3. Verify operation in queue
    const ops = await getPendingOperations();
    expect(ops).toHaveLength(1);

    // 4. Come online
    simulateOnline();

    // 5. Wait for processing
    await waitFor(() => {
      const ops = getPendingOperations();
      expect(ops).toHaveLength(0);
    });
  });
});
```

### Step 4.2: API Error Handling

```typescript
// Test error recovery
it('should retry failed operations', async () => {
  // Mock API to fail once, then succeed
  mock.onPost('/requisitions').replyOnce(500).replyOnce(201, { success: true });

  // Queue operation while offline
  await queueOperation('CREATE', 'requisition', data);

  // Come online - first attempt fails, second succeeds
  await processQueue();

  // Should retry and eventually succeed
  expect(operations).toHaveLength(0);
});
```

### Step 4.3: Performance Testing

```typescript
// Measure sync time
const startTime = performance.now();
await processQueue();
const duration = performance.now() - startTime;

console.log(`Queue processing took ${duration}ms for ${ops.length} operations`);
// Should be <5s for 10 operations
```

---

## Phase 5: Cleanup (30 minutes)

### Step 5.1: Remove documentStore

```bash
# Delete files no longer needed
rm src/lib/workflow-stores.ts
rm -rf src/lib/storage/  # If only used for mock data
```

### Step 5.2: Remove Mock Utilities

```typescript
// Remove from use-initialize-storage.ts:
// import { initializeWorkflowStores } from '@/lib/workflow-stores';
// initializeWorkflowStores();

// Keep only:
initializeStorage(); // For localStorage cache initialization
```

### Step 5.3: Environment Configuration

Create `.env.local`:

```env
NEXT_PUBLIC_API_URL=http://localhost:3001/api
NEXT_PUBLIC_API_TIMEOUT=30000
```

---

## Verification Checklist

### ✅ Before Going to Production

- [ ] All CRUD operations work with real API
- [ ] Offline queue processes successfully
- [ ] Cache invalidation works correctly
- [ ] Error handling for network failures
- [ ] Retry logic for failed operations
- [ ] localStorage fallback when API fails
- [ ] Performance acceptable (<2s load time)
- [ ] Authentication working (token refresh)
- [ ] Logging configured for debugging
- [ ] UI updated with offline indicator

### ✅ Post-Migration Monitoring

- [ ] Monitor API response times
- [ ] Track cache hit rates
- [ ] Monitor offline queue size
- [ ] Verify token refresh working
- [ ] Check error logs for API failures
- [ ] User feedback on sync experience

---

## Rollback Plan

If issues occur:

### Quick Rollback (< 5 minutes)

```bash
# Revert server actions to use documentStore
git checkout HEAD -- src/app/_actions/

# Revert API client
git checkout HEAD -- src/lib/api/

# Restart app
npm run dev
```

### Full Rollback

```bash
git checkout HEAD~1  # Previous commit before migration
npm install
npm run dev
```

---

## Post-Migration Optimizations

### 1. API Response Caching

```typescript
// Cache entire responses in localStorage
staleTime: 5 * 60 * 1000,  // Already configured in providers.tsx
```

### 2. Background Sync

```typescript
// Already configured with offline queue processor
useOfflineQueueProcessor(); // Runs automatically when online
```

### 3. Request Batching (Optional)

```typescript
// For high-traffic scenarios
import { BatchRequestManager } from '@/lib/api/batch';

// Batch multiple requests into single API call
const batch = new BatchRequestManager();
batch.add(async () => await apiClient.requisitions.getById('1'));
batch.add(async () => await apiClient.requisitions.getById('2'));
await batch.execute();
```

---

## Troubleshooting

### Issue: Operations not syncing

```typescript
// Check queue status
const stats = await getQueueStats();
console.log('Queue stats:', stats);

// Check network
console.log('Online:', navigator.onLine);

// Manually trigger sync
await useOfflineQueueProcessor();
```

### Issue: Stale data being served

```typescript
// Force fresh data
queryClient.refetchQueries({ queryKey: queryKeys.requisitions.all() });

// Clear cache
localStorage.removeItem('liyali_requisitions_cache');
```

### Issue: API errors not caught

```typescript
// Check error interceptor
httpClient.interceptors.response.handlers;

// Add detailed logging
httpClient.defaults.onError = (error) => {
  console.error('[API Error]', error.response?.status, error.message);
};
```

---

## Support

For issues during migration:

1. Check offline queue status: `getQueueStats()`
2. Review error logs: Console + Server logs
3. Test API directly: `curl http://localhost:3001/api/requisitions`
4. Verify auth token: `localStorage.getItem('auth_token')`
5. Clear all caches: `localStorage.clear()` + `queryClient.clear()`

---

**Estimated Total Time**: 4-6 hours
**Complexity**: Low (architecture supports it)
**Breaking Changes**: None (same API surface)

Good luck with the migration! 🚀
