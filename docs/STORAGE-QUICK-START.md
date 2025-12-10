# Storage System Quick Start

## What's New?

LocalStorage is now your **single source of truth** for all data:
- Purchase Orders (5 examples with different statuses)
- Requisitions (4 examples)
- Payment Vouchers (4 examples)

## Accessing Data in Components

### Using React Query (Recommended)

```typescript
import { usePurchaseOrdersQuery } from '@/hooks/use-storage-queries';

export function MyComponent() {
  const { data: orders = [], isLoading } = usePurchaseOrdersQuery();

  if (isLoading) return <p>Loading...</p>;

  return (
    <ul>
      {orders.map((order) => (
        <li key={order.id}>{order.documentNumber}</li>
      ))}
    </ul>
  );
}
```

### Using Storage Hooks Directly

```typescript
import { getPurchaseOrders, savePurchaseOrder } from '@/lib/storage';

// Get all orders
const orders = getPurchaseOrders();

// Get a specific order
const order = getPurchaseOrderById('po-123');

// Save/update an order
savePurchaseOrder(updatedOrder);

// Delete an order
deletePurchaseOrder('po-123');
```

## Common Tasks

### Get Orders by Status

```typescript
import { getPurchaseOrdersByStatus } from '@/lib/storage';

const draftOrders = getPurchaseOrdersByStatus('DRAFT');
const approvedOrders = getPurchaseOrdersByStatus('APPROVED');
```

### Get Orders by User

```typescript
import { getPurchaseOrdersByCreator } from '@/lib/storage';

const myOrders = getPurchaseOrdersByCreator(userId);
```

### Get Requisitions by Department

```typescript
import { getRequisitionsByDepartment } from '@/lib/storage';

const opsReqs = getRequisitionsByDepartment('Operations');
```

### Filter Custom

```typescript
import { filterPurchaseOrders } from '@/lib/storage';

const expensiveOrders = filterPurchaseOrders(
  (po) => po.metadata?.totalAmount > 50000
);
```

### Save New Document

```typescript
import { savePurchaseOrder } from '@/lib/storage';
import { v4 as uuidv4 } from 'uuid';

const newPO = {
  id: `po-${uuidv4()}`,
  type: 'PURCHASE_ORDER',
  documentNumber: 'PO-2024-001',
  status: 'DRAFT',
  currentStage: 0,
  createdBy: userId,
  createdAt: new Date(),
  updatedAt: new Date(),
  metadata: {
    vendorName: 'Acme Corp',
    vendorId: 'VENDOR-123',
    items: [...],
    totalAmount: 10000,
    currency: 'ZMW',
  },
};

const saved = savePurchaseOrder(newPO);
```

## All Available Queries

### Purchase Orders
```typescript
usePurchaseOrdersQuery()
usePurchaseOrdersByCreatorQuery(userId)
usePurchaseOrdersAsWorkflowDocumentsQuery(userId?)
```

### Requisitions
```typescript
useRequisitionsQuery()
useRequisitionsByCreatorQuery(userId)
useRequisitionsAsWorkflowDocumentsQuery(userId?)
```

### Payment Vouchers
```typescript
usePaymentVouchersQuery()
usePaymentVouchersByCreatorQuery(userId)
usePaymentVouchersAsWorkflowDocumentsQuery(userId?)
```

## All Available Hooks

```typescript
// Purchase Orders
getPurchaseOrders()
getPurchaseOrderById(id)
savePurchaseOrder(po)
deletePurchaseOrder(id)
filterPurchaseOrders(predicate)
getPurchaseOrdersByStatus(status)
getPurchaseOrdersByCreator(userId)

// Requisitions
getRequisitions()
getRequisitionById(id)
saveRequisition(req)
deleteRequisition(id)
filterRequisitions(predicate)
getRequisitionsByStatus(status)
getRequisitionsByCreator(userId)
getRequisitionsByDepartment(dept)

// Payment Vouchers
getPaymentVouchers()
getPaymentVoucherById(id)
savePaymentVoucher(pv)
deletePaymentVoucher(id)
filterPaymentVouchers(predicate)
getPaymentVouchersByStatus(status)
getPaymentVouchersByCreator(userId)
getPaymentVouchersByAmount(min, max)

// Bulk
getAllDocuments()
getDocumentsByStatus(status)
getDocumentsByCreator(userId)
```

## Development Tools

### View All Data
```typescript
import { exportStorageAsJSON } from '@/lib/storage';

const allData = exportStorageAsJSON();
console.log(allData);
```

### Get Stats
```typescript
import { getStorageStats } from '@/lib/storage';

const stats = getStorageStats();
// { purchaseOrders: 5, requisitions: 4, paymentVouchers: 4, total: 13 }
```

### Reset Everything
```typescript
import { resetStorage } from '@/lib/storage';

// Clears all data and reinitializes with seed data
resetStorage();
```

### Clear One Type
```typescript
import { clearDocuments, STORAGE_KEYS } from '@/lib/storage';

clearDocuments(STORAGE_KEYS.PURCHASE_ORDERS);
clearDocuments(STORAGE_KEYS.REQUISITIONS);
clearDocuments(STORAGE_KEYS.PAYMENT_VOUCHERS);
```

## File Structure

```
src/lib/storage/          ← All storage code here
├── storage.ts           ← Core CRUD operations
├── init.ts             ← Initialization
├── hooks.ts            ← 30+ helper functions
├── seed-data.ts        ← Seed generators
├── index.ts            ← Barrel exports
└── README.md           ← Full documentation

src/hooks/
└── use-storage-queries.ts   ← React Query hooks

src/hooks/use-initialize-storage.ts  ← Auto-init hook
```

## Debugging

### Check if Initialized
```typescript
import { isStorageInitialized } from '@/lib/storage';

if (isStorageInitialized()) {
  console.log('Storage has been initialized');
}
```

### Check Raw localStorage
```javascript
// In browser DevTools console:
console.log(JSON.parse(localStorage.getItem('liyali_purchase_orders')));
console.log(JSON.parse(localStorage.getItem('liyali_requisitions')));
console.log(JSON.parse(localStorage.getItem('liyali_payment_vouchers')));
```

### Monitor Changes
```typescript
import { getAllDocuments } from '@/lib/storage';

// This gets fresh data every time you call it
console.log(getAllDocuments());
```

## Common Patterns

### In a Server Action

```typescript
'use server';

import { savePurchaseOrder, getPurchaseOrderById } from '@/lib/storage';

export async function updatePurchaseOrder(id: string, updates: Partial<PurchaseOrder>) {
  const existing = getPurchaseOrderById(id);
  if (!existing) {
    throw new Error('PO not found');
  }

  const updated = {
    ...existing,
    ...updates,
    updatedAt: new Date(),
  };

  return savePurchaseOrder(updated);
}
```

### In a React Component

```typescript
'use client';

import { usePurchaseOrdersQuery } from '@/hooks/use-storage-queries';

export function DashboardPOs() {
  const { data: orders = [], isLoading, refetch } = usePurchaseOrdersQuery();

  return (
    <div>
      <button onClick={() => refetch()}>Refresh</button>
      {orders.map((order) => (
        <div key={order.id}>{order.documentNumber}</div>
      ))}
    </div>
  );
}
```

### Creating New Document

```typescript
'use client';

import { savePurchaseOrder } from '@/lib/storage';
import { v4 as uuidv4 } from 'uuid';

async function handleCreate(formData: CreatePOForm) {
  const newPO = {
    id: `po-${uuidv4()}`,
    type: 'PURCHASE_ORDER' as const,
    documentNumber: generatePONumber(),
    status: 'DRAFT' as const,
    currentStage: 0,
    createdBy: session.user.id,
    createdByUser: session.user,
    createdAt: new Date(),
    updatedAt: new Date(),
    metadata: {
      vendorName: formData.vendor,
      vendorId: formData.vendorId,
      items: formData.items,
      totalAmount: formData.total,
      currency: 'ZMW',
      deliveryDate: new Date(formData.deliveryDate),
    },
  };

  const saved = savePurchaseOrder(newPO);
  console.log('Created:', saved);
}
```

## Full Documentation

See [`/src/lib/storage/README.md`](../src/lib/storage/README.md) for complete API reference.

## Next Steps

1. ✅ Purchase Orders - Already showing in table
2. 🔜 Update Requisitions table to use `useRequisitionsAsWorkflowDocumentsQuery()`
3. 🔜 Update Payment Vouchers table to use `usePaymentVouchersAsWorkflowDocumentsQuery()`
4. 🔜 Test creating/updating/deleting documents
5. 🔜 Implement backend API endpoints
6. 🔜 Swap storage calls with API calls
7. 🔜 Delete `/lib/storage` folder

