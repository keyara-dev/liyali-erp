# Storage Architecture Overview

## System Design

The storage system follows a **layered architecture** with clear separation of concerns:

```
┌────────────────────────────────────────────────────────────┐
│                    COMPONENTS LAYER                         │
│              (React Components in pages/                    │
│              components)                                    │
│                                                              │
│ Examples:                                                    │
│ - purchase-orders-table.tsx                                 │
│ - requisitions-page.tsx                                     │
│ - payment-vouchers-form.tsx                                 │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌────────────────────────────────────────────────────────────┐
│                  REACT QUERY LAYER                          │
│         (Caching, refetching, state management)            │
│                                                              │
│ Location: src/hooks/use-storage-queries.ts                 │
│                                                              │
│ Hooks:                                                       │
│ - usePurchaseOrdersQuery()                                  │
│ - useRequisitionsQuery()                                    │
│ - usePaymentVouchersQuery()                                 │
│ - usePurchaseOrdersByCreatorQuery()                         │
│ - ... and more                                              │
│                                                              │
│ Benefits:                                                    │
│ - Automatic caching (5 min)                                 │
│ - Background refetching                                     │
│ - Query invalidation                                        │
│ - Built-in loading/error states                             │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌────────────────────────────────────────────────────────────┐
│              STORAGE HOOKS LAYER                            │
│       (Domain-specific helper functions)                    │
│                                                              │
│ Location: src/lib/storage/hooks.ts                          │
│                                                              │
│ Functions:                                                   │
│ - getPurchaseOrders()                                       │
│ - getPurchaseOrderById(id)                                  │
│ - savePurchaseOrder(po)                                     │
│ - deletePurchaseOrder(id)                                   │
│ - getPurchaseOrdersByStatus()                               │
│ - getPurchaseOrdersByCreator()                              │
│ - getRequisitions()                                         │
│ - ... 30+ more                                              │
│                                                              │
│ Benefits:                                                    │
│ - Type-safe                                                 │
│ - Semantic naming                                           │
│ - Pre-built filtering                                       │
│ - Easy to mock for testing                                  │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌────────────────────────────────────────────────────────────┐
│              CORE STORAGE LAYER                             │
│    (Generic CRUD operations for all document types)        │
│                                                              │
│ Location: src/lib/storage/storage.ts                        │
│                                                              │
│ Generic Functions:                                           │
│ - getDocuments<T>(storageKey): T[]                          │
│ - getDocumentById<T>(storageKey, id): T | null             │
│ - saveDocument<T>(storageKey, doc): T                       │
│ - deleteDocument(storageKey, id): void                      │
│ - clearDocuments(storageKey): void                          │
│ - clearAllData(): void                                      │
│                                                              │
│ Constants:                                                   │
│ - STORAGE_KEYS.PURCHASE_ORDERS                              │
│ - STORAGE_KEYS.REQUISITIONS                                 │
│ - STORAGE_KEYS.PAYMENT_VOUCHERS                             │
│                                                              │
│ Utilities:                                                   │
│ - isStorageInitialized()                                    │
│ - getStorageStats()                                         │
│ - exportStorageAsJSON()                                     │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌────────────────────────────────────────────────────────────┐
│            BROWSER LOCALSTORAGE                             │
│        (Single Source of Truth)                             │
│                                                              │
│ Storage Keys:                                                │
│ - liyali_purchase_orders                                    │
│ - liyali_requisitions                                       │
│ - liyali_payment_vouchers                                   │
│                                                              │
│ Data Format: JSON strings                                   │
│ Persistence: Until manually cleared                         │
└────────────────────────────────────────────────────────────┘
```

## Initialization Flow

```
┌─────────────────────────────────────────────────────────────┐
│  APP STARTUP                                                 │
│  (Next.js hydration)                                         │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│  RootLayout (src/app/layout.tsx)                             │
│  Renders Providers component                                │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│  Providers (src/app/providers.tsx)                           │
│  Returns StorageInitializer wrapper                          │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│  StorageInitializer (src/app/providers.tsx)                 │
│  Calls useInitializeStorage() hook                           │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│  useInitializeStorage() (src/hooks/use-initialize-storage)  │
│  Runs initializeStorage() in useEffect                       │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│  initializeStorage() (src/lib/storage/init.ts)              │
│  1. Check if already initialized                            │
│  2. If not, seed data                                       │
└────────────────────┬────────────────────────────────────────┘
                     │
        ┌────────────┼────────────┐
        ▼            ▼            ▼
   ┌────────┐   ┌──────────┐   ┌────────────┐
   │  seed  │   │  seed    │   │  seed      │
   │  POs   │   │  Reqs    │   │  PVs       │
   └────┬───┘   └────┬─────┘   └────┬───────┘
        │            │              │
        └────────────┼──────────────┘
                     │
                     ▼
        ┌────────────────────────────────┐
        │  localStorage ready            │
        │  - liyali_purchase_orders      │
        │  - liyali_requisitions         │
        │  - liyali_payment_vouchers     │
        └────────────────────────────────┘
```

## Data Flow: Reading

```
Component Renders
      │
      ▼
Calls usePurchaseOrdersQuery()
      │
      ▼
React Query checks cache
      │
      ├─ If cached & not stale
      │  └─> Return cached data
      │
      └─ If stale or empty
         └─> Call queryFn
             │
             └─> getPurchaseOrders()
                 │
                 └─> getDocuments(STORAGE_KEYS.PURCHASE_ORDERS)
                     │
                     └─> localStorage.getItem('liyali_purchase_orders')
                         │
                         ├─ Parse JSON
                         └─> Return array
                             │
                             └─> Component updates
```

## Data Flow: Writing

```
User Action (Create/Update/Delete)
      │
      ▼
Call savePurchaseOrder(updatedPO)
      │
      ├─> Validate data
      │
      ├─> Get current documents
      │   └─> getDocuments(STORAGE_KEYS.PURCHASE_ORDERS)
      │
      ├─> Update array (add/update/delete)
      │
      ├─> Write back to localStorage
      │   └─> localStorage.setItem('liyali_purchase_orders', JSON.stringify(docs))
      │
      └─> Return updated document
          │
          └─> Component updates
              │
              └─> React Query invalidates cache (optional)
                  │
                  └─> Automatic refetch on next access
```

## Storage Keys & Structure

### Purchase Orders
```
localStorage key: liyali_purchase_orders
Structure: Array<PurchaseOrder>

PurchaseOrder {
  id: string (UUID)
  type: 'PURCHASE_ORDER'
  documentNumber: string (PO-2024-001)
  status: 'DRAFT' | 'SUBMITTED' | 'IN_REVIEW' | 'APPROVED' | 'REJECTED'
  currentStage: number
  createdBy: string (userId)
  createdByUser: User
  createdAt: Date
  updatedAt: Date
  metadata: {
    vendorName: string
    vendorId: string
    items: Array<{
      id: string
      description: string
      quantity: number
      unitCost: number
      totalCost: number
    }>
    totalAmount: number
    currency: string
    deliveryDate: Date
    specialInstructions?: string
  }
}
```

### Requisitions
```
localStorage key: liyali_requisitions
Structure: Array<RequisitionForm>

RequisitionForm {
  id: string (UUID)
  type: 'REQUISITION'
  documentNumber: string (REQ-2024-001)
  status: 'DRAFT' | 'SUBMITTED' | 'APPROVED' | 'REJECTED'
  currentStage: number
  createdBy: string (userId)
  createdByUser: User
  createdAt: Date
  updatedAt: Date
  metadata: {
    department: string
    requestedFor: string
    items: Array<{
      id: string
      itemDescription: string
      quantity: number
      estimatedCost: number
    }>
    justification: string
    budgetCode: string
  }
}
```

### Payment Vouchers
```
localStorage key: liyali_payment_vouchers
Structure: Array<PaymentVoucher>

PaymentVoucher {
  id: string (UUID)
  type: 'PAYMENT_VOUCHER'
  documentNumber: string (PV-2024-001)
  status: 'DRAFT' | 'SUBMITTED' | 'IN_REVIEW' | 'APPROVED'
  currentStage: number
  createdBy: string (userId)
  createdByUser: User
  createdAt: Date
  updatedAt: Date
  metadata: {
    payeeName: string
    payeeId: string
    amount: number
    currency: string
    reason: string
    accountCode: string
    department: string
  }
}
```

## File Locations

### Storage Module
- **Core Operations**: `src/lib/storage/storage.ts` (150 lines)
- **Initialization**: `src/lib/storage/init.ts` (60 lines)
- **Helper Hooks**: `src/lib/storage/hooks.ts` (200 lines)
- **Seed Data**: `src/lib/storage/seed-data.ts` (350 lines)
- **Exports**: `src/lib/storage/index.ts` (50 lines)
- **Documentation**: `src/lib/storage/README.md` (500+ lines)

### Integration
- **React Query Hooks**: `src/hooks/use-storage-queries.ts` (200 lines)
- **Initialization Hook**: `src/hooks/use-initialize-storage.ts` (20 lines)
- **Providers Integration**: `src/app/providers.tsx` (Updated)

### Documentation
- **Architecture**: `docs/STORAGE-ARCHITECTURE.md` (This file)
- **Setup Guide**: `docs/STORAGE-SYSTEM-SETUP.md`
- **Quick Start**: `docs/STORAGE-QUICK-START.md`

## Performance Characteristics

### Read Operations
- **Time Complexity**: O(n) where n = number of documents
- **Speed**: < 1ms for 100 documents
- **Caching**: React Query caches results (5 min default)

### Write Operations
- **Time Complexity**: O(n) where n = number of documents
- **Speed**: < 2ms for 100 documents
- **Side Effects**: Optional React Query invalidation

### Storage Size
- **Per Document**: ~500-1000 bytes (typical)
- **localStorage Limit**: ~5-10MB (browser dependent)
- **Capacity**: Can handle 10,000+ documents

## Upgrade Path to Backend APIs

### Phase 1: Prepare (1 hour)
- Create REST API endpoints
- Test endpoints with Postman/Thunder Client

### Phase 2: Update Hooks (2 hours)
- Update `use-storage-queries.ts` queryFn functions
- Add API error handling

### Phase 3: Update Storage Layer (1 hour)
- Update `src/lib/storage/hooks.ts` to call APIs
- Add error boundaries

### Phase 4: Test (2 hours)
- Test all CRUD operations
- Test error scenarios
- Test network failures

### Phase 5: Cleanup (30 mins)
- Delete `/src/lib/storage/` folder
- Delete `use-initialize-storage.ts`
- Delete old storage hooks
- Remove from providers

**Total Time: ~6-7 hours**

## Testing Strategies

### Unit Tests
```typescript
// Test storage functions
import { savePurchaseOrder, getPurchaseOrderById } from '@/lib/storage';

test('should save and retrieve purchase order', () => {
  const po = createMockPO();
  const saved = savePurchaseOrder(po);
  const retrieved = getPurchaseOrderById(saved.id);
  expect(retrieved).toEqual(saved);
});
```

### Integration Tests
```typescript
// Test React Query hooks
import { render, screen } from '@testing-library/react';
import { QueryClientProvider } from '@tanstack/react-query';

test('should load purchase orders', async () => {
  const wrapper = ({ children }) => (
    <QueryClientProvider client={queryClient}>
      {children}
    </QueryClientProvider>
  );

  const { } = render(<PurchaseOrderList />, { wrapper });
  // assertions
});
```

### E2E Tests
```typescript
// Test complete user flow
test('user creates and views purchase order', async () => {
  await page.goto('/purchase-orders');
  await page.click('button:has-text("Create")');
  // fill form
  await page.click('button:has-text("Save")');
  // verify appears in list
});
```

## Monitoring & Debugging

### Browser DevTools
```javascript
// View all storage
console.log(JSON.parse(localStorage.getItem('liyali_purchase_orders')));

// Check stats
import { getStorageStats } from '@/lib/storage';
console.log(getStorageStats());

// Export all data
import { exportStorageAsJSON } from '@/lib/storage';
console.log(exportStorageAsJSON());
```

### React Query DevTools
- Automatically included in `app/providers.tsx`
- Click devtools icon to see:
  - Query cache status
  - Stale time remaining
  - Background fetches
  - Query history

### Console Logging
- Storage initialization logs to console
- All operations log on error
- Enable debug mode in React Query

## Summary

This architecture provides:
- ✅ Clean separation of concerns
- ✅ Type safety throughout
- ✅ Caching and performance
- ✅ Easy testing
- ✅ Clear upgrade path
- ✅ Great DX for developers
- ✅ Single source of truth
- ✅ No backend dependency during development

