
# Offline-First Architecture

The Liyali Gateway frontend is built with an offline-first approach, ensuring users can continue working even without internet connectivity. Data is synchronized when the connection is restored.

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    Offline-First Layers                    │
├─────────────────────────────────────────────────────────────┤
│ 1. UI Layer (React Components)                             │
│    ↕ Always responsive, shows offline status               │
│                                                             │
│ 2. Query Layer (TanStack Query)                            │
│    ↕ Caches API responses, serves stale data offline       │
│                                                             │
│ 3. Storage Layer (localStorage)                            │
│    ↕ Persists data locally, queues offline operations      │
│                                                             │
│ 4. Sync Layer (Background Sync)                            │
│    ↕ Processes queue when online, resolves conflicts       │
│                                                             │
│ 5. API Layer (Server Actions/HTTP)                         │
│    ↕ Communicates with backend when available              │
└─────────────────────────────────────────────────────────────┘
```

## Storage Management

### Unified Storage API

All data operations go through a centralized storage system:

```typescript
// src/lib/storage/storage.ts
export const STORAGE_KEYS = {
  PURCHASE_ORDERS: 'liyali_purchase_orders',
  REQUISITIONS: 'liyali_requisitions',
  PAYMENT_VOUCHERS: 'liyali_payment_vouchers',
  GOODS_RECEIVED_NOTES: 'liyali_goods_received_notes',
  OFFLINE_QUEUE: 'liyali_offline_queue',
  SYNC_STATUS: 'liyali_sync_status',
} as const;

// Generic document operations with offline support
export function getDocuments<T>(storageKey: string): T[] {
  if (typeof window === 'undefined') return [];

  try {
    const stored = localStorage.getItem(storageKey);
    if (!stored) return [];
    const parsed = JSON.parse(stored);
    return Array.isArray(parsed) ? parsed : [];
  } catch (error) {
    console.error(`Failed to load documents from ${storageKey}:`, error);
    return [];
  }
}

export function saveDocument<T extends { id: string }>(
  storageKey: string,
  document: T
): T {
  try {
    if (typeof window === 'undefined') return document;

    const documents = getDocuments<T>(storageKey);
    const index = documents.findIndex((doc) => doc.id === document.id);

    if (index >= 0) {
      documents[index] = { ...documents[index], ...document };
    } else {
      documents.push(document);
    }

    localStorage.setItem(storageKey, JSON.stringify(documents));
    
    // Mark as modified for sync
    markForSync(storageKey, document.id, 'UPDATE');
    
    return document;
  } catch (error) {
    console.error(`Failed to save document to ${storageKey}:`, error);
    throw error;
  }
}
```

### Offline Queue System

Operations performed while offline are queued for later execution:

```typescript
// src/lib/storage/offline-queue.ts
export interface QueuedOperation {
  id: string;
  type: 'CREATE' | 'UPDATE' | 'DELETE';
  entity: string;
  data: any;
  timestamp: number;
  retryCount: number;
  maxRetries: number;
}

export class OfflineQueue {
  private queue: QueuedOperation[] = [];
  private processing = false;

  constructor() {
    this.loadQueue();
    this.setupOnlineListener();
  }

  // Add operation to queue
  enqueue(operation: Omit<QueuedOperation, 'id' | 'timestamp' | 'retryCount'>) {
    const queuedOp: QueuedOperation = {
      ...operation,
      id: crypto.randomUUID(),
      timestamp: Date.now(),
      retryCount: 0,
      maxRetries: 3,
    };

    this.queue.push(queuedOp);
    this.saveQueue();

    // Process immediately if online
    if (navigator.onLine) {
      this.processQueue();
    }
  }

  // Process all queued operations
  async processQueue() {
    if (this.processing || !navigator.onLine) return;

    this.processing = true;

    try {
      const operations = [...this.queue];
      
      for (const operation of operations) {
        try {
          await this.executeOperation(operation);
          this.removeFromQueue(operation.id);
        } catch (error) {
          console.error('Failed to execute queued operation:', error);
          
          operation.retryCount++;
          
          if (operation.retryCount >= operation.maxRetries) {
            console.error('Max retries reached for operation:', operation);
            this.removeFromQueue(operation.id);
          }
        }
      }
    } finally {
      this.processing = false;
      this.saveQueue();
    }
  }

  private async executeOperation(operation: QueuedOperation) {
    switch (operation.entity) {
      case 'requisitions':
        return this.executeRequisitionOperation(operation);
      case 'purchase-orders':
        return this.executePurchaseOrderOperation(operation);
      case 'payment-vouchers':
        return this.executePaymentVoucherOperation(operation);
      default:
        throw new Error(`Unknown entity type: ${operation.entity}`);
    }
  }

  private setupOnlineListener() {
    window.addEventListener('online', () => {
      console.log('Connection restored, processing offline queue');
      this.processQueue();
    });
  }

  private loadQueue() {
    try {
      const stored = localStorage.getItem(STORAGE_KEYS.OFFLINE_QUEUE);
      if (stored) {
        this.queue = JSON.parse(stored);
      }
    } catch (error) {
      console.error('Failed to load offline queue:', error);
      this.queue = [];
    }
  }

  private saveQueue() {
    try {
      localStorage.setItem(STORAGE_KEYS.OFFLINE_QUEUE, JSON.stringify(this.queue));
    } catch (error) {
      console.error('Failed to save offline queue:', error);
    }
  }
}

// Global queue instance
export const offlineQueue = new OfflineQueue();
```

## Query Cache Integration

### Offline-Aware Queries

TanStack Query is configured to serve stale data when offline:

```typescript
// src/hooks/use-offline-queries.ts
export function useOfflineRequisitions() {
  return useQuery({
    queryKey: [QUERY_KEYS.REQUISITIONS.ALL],
    queryFn: async () => {
      if (navigator.onLine) {
        // Try API first when online
        try {
          const response = await getRequisitions();
          if (response.success) {
            // Cache in localStorage
            saveDocuments(STORAGE_KEYS.REQUISITIONS, response.data);
            return response.data;
          }
        } catch (error) {
          console.warn('API failed, falling back to localStorage:', error);
        }
      }

     