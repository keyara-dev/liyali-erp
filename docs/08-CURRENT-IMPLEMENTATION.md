# Current Implementation Guide

**Phase**: Phase 11 (localStorage-based)
**Status**: Fully Functional with Client-Side Search
**Last Updated**: 2025-12-12

---

## Overview

The Liyali Gateway currently operates as a **single-user, browser-based application** using localStorage as the data layer. All business logic runs on the client side, with no backend server required for basic functionality.

---

## Architecture

### Data Layer: localStorage

**Storage Keys:**
```
liyali_purchase_orders        // Array of PurchaseOrder objects
liyali_requisitions           // Array of RequisitionForm objects
liyali_payment_vouchers       // Array of PaymentVoucher objects
liyali_goods_received_notes   // Array of GoodsReceivedNote objects
```

**Initialization:**
```typescript
// In frontend/src/lib/storage/init.ts
export function initializeStorage(): void {
  if (typeof window === 'undefined') return;

  if (isStorageInitialized()) {
    console.log('✓ Storage already initialized');
    return;
  }

  // Initialize with seed data on first load
  const purchaseOrders = createSeedPurchaseOrders();
  const requisitions = createSeedRequisitions();
  const paymentVouchers = createSeedPaymentVouchers();
  const goodsReceivedNotes = createSeedGoodsReceivedNotes();

  saveDocuments(STORAGE_KEYS.PURCHASE_ORDERS, purchaseOrders);
  saveDocuments(STORAGE_KEYS.REQUISITIONS, requisitions);
  saveDocuments(STORAGE_KEYS.PAYMENT_VOUCHERS, paymentVouchers);
  saveDocuments(STORAGE_KEYS.GOODS_RECEIVED_NOTES, goodsReceivedNotes);
}
```

---

## Data Access Layer

### Storage Functions

Located in `frontend/src/lib/storage/hooks.ts`:

```typescript
// Get all documents of a type
getPurchaseOrders(): PurchaseOrder[]
getRequisitions(): RequisitionForm[]
getPaymentVouchers(): PaymentVoucher[]
getGoodsReceivedNotes(): GoodsReceivedNote[]

// Get single document
getPurchaseOrderById(id: string): PurchaseOrder | null
getRequisitionById(id: string): RequisitionForm | null
// ... etc for all types

// Save/Update document
savePurchaseOrder(po: PurchaseOrder): PurchaseOrder
saveRequisition(req: RequisitionForm): RequisitionForm
// ... etc for all types

// Delete document
deletePurchaseOrder(id: string): void
deleteRequisition(id: string): void
// ... etc for all types

// Filtering
getPurchaseOrdersByStatus(status: string): PurchaseOrder[]
getRequisitionsByDepartment(dept: string): RequisitionForm[]
getPaymentVouchersByAmount(min: number, max: number): PaymentVoucher[]
getGoodsReceivedNotesByPurchaseOrder(poId: string): GoodsReceivedNote[]
```

### Core Storage API

Located in `frontend/src/lib/storage/storage.ts`:

```typescript
export function getDocuments<T>(storageKey: string): T[]
export function getDocumentById<T>(storageKey: string, id: string): T | null
export function saveDocument<T>(storageKey: string, document: T): T
export function saveDocuments<T>(storageKey: string, documents: T[]): void
export function deleteDocument(storageKey: string, id: string): void
export function clearDocuments(storageKey: string): void
export function clearAllData(): void
export function isStorageInitialized(): boolean
export function getStorageStats(): StorageStats
export function exportStorageAsJSON(): Record<string, any>
```

---

## Component Architecture

### Table Components

All document type tables use the same pattern:

**Purchase Orders Table** (`frontend/src/app/(private)/(main)/purchase-orders/_components/po-table.tsx`)
```typescript
export function PurchaseOrdersTable({ userId, userRole, refreshTrigger }: Props) {
  // 1. Fetch data using React Query
  const { data: pos = [], refetch } =
    usePurchaseOrdersAsWorkflowDocumentsQuery();

  // 2. Refetch on trigger
  useEffect(() => {
    refetch();
  }, [refreshTrigger, refetch]);

  // 3. Map to columns
  const columns: ColumnDef<WorkflowDocument>[] = [
    // Define table columns
  ];

  // 4. Render DataTable
  return (
    <DataTable
      columns={columns}
      data={pos}
      hideSearchBar={false}
      renderRowActions={(po) => (
        // Action buttons: View, Edit, Delete
      )}
    />
  );
}
```

### Search Implementation

**Current Search Page** (`frontend/src/app/(private)/(main)/search/`)

The search functionality is implemented entirely on the client side:

1. **Search Form** (`search-form.tsx`):
   - Date pickers for start/end dates
   - Select dropdowns for type and status
   - Text input for document number
   - Calls parent's `onSearch` callback with filters

2. **Search Results** (`transaction-results.tsx`):
   - Uses `useMutation` to perform client-side search
   - Calls `performSearch()` function which:
     - Retrieves all documents from storage
     - Applies filters (documentNumber, type, status, dates)
     - Sorts by creation date
     - Paginates results
   - Displays results in DataTable
   - Shows loading skeleton during search

3. **Filter Logic**:
```typescript
function performSearch(
  filters: SearchFilters,
  page: number,
  limit: number
) {
  // 1. Get all documents from storage
  const pos = getPurchaseOrders();
  const reqs = getRequisitions();
  const pvs = getPaymentVouchers();
  const grns = getGoodsReceivedNotes();

  // 2. Combine into single array
  const allDocs: WorkflowDocument[] = [
    ...pos.map(convertToWorkflowDocument),
    ...reqs.map(convertToWorkflowDocument),
    ...pvs.map(convertToWorkflowDocument),
    ...grns.map(convertToWorkflowDocument),
  ];

  // 3. Apply filters
  let filtered = allDocs.filter((doc) => {
    // Document number filter (partial match, case-insensitive)
    if (filters.documentNumber &&
        !doc.documentNumber.toLowerCase()
          .includes(filters.documentNumber.toLowerCase())) {
      return false;
    }

    // Type filter
    if (filters.documentType !== "ALL" &&
        doc.type !== filters.documentType) {
      return false;
    }

    // Status filter
    if (filters.status !== "ALL" &&
        doc.status !== filters.status) {
      return false;
    }

    // Date range filters
    if (filters.startDate) {
      const startDate = new Date(filters.startDate);
      if (doc.createdAt < startDate) return false;
    }

    if (filters.endDate) {
      const endDate = new Date(filters.endDate);
      endDate.setHours(23, 59, 59, 999);
      if (doc.createdAt > endDate) return false;
    }

    return true;
  });

  // 4. Sort and paginate
  filtered.sort((a, b) =>
    new Date(b.createdAt).getTime() -
    new Date(a.createdAt).getTime()
  );

  const total = filtered.length;
  const totalPages = Math.ceil(total / limit);
  const skip = (page - 1) * limit;
  const paginatedData = filtered.slice(skip, skip + limit);

  return {
    documents: paginatedData,
    total,
    totalPages,
  };
}
```

---

## Seed Data

### Initialization on App Load

```typescript
// In root layout or app initialization hook
useEffect(() => {
  initializeStorage();  // Initializes localStorage on first load
}, []);
```

### Seed Data Structure

**32 Total Documents** (distributed across 4 types):

- **10 Purchase Orders**: Various statuses (DRAFT, SUBMITTED, IN_REVIEW, APPROVED, REJECTED)
  - Created by 2 different requesters
  - Associated with 8 different vendors
  - Amounts ranging from 3,950 to 125,000 ZMW
  - Dates from 15 days ago to 1 day ago

- **7 Requisitions**: Various statuses and departments
  - Created by 2 different requesters
  - From departments: IT, Finance, Operations, HR
  - Amounts ranging from 35,000 to 200,000 ZMW

- **9 Payment Vouchers**: Various stages
  - From 5 different payees
  - Amounts ranging from 15,000 to 100,000 ZMW

- **6 Goods Received Notes**: Various statuses
  - Linked to specific purchase orders
  - Received quantities match PO quantities
  - Warehouse locations assigned

### Creating Custom Seed Data

To add more seed data, modify `frontend/src/lib/storage/seed-data.ts`:

```typescript
export function createSeedPurchaseOrders(): PurchaseOrder[] {
  const now = new Date();

  return [
    {
      id: `po-${uuidv4()}`,
      type: 'PURCHASE_ORDER',
      documentNumber: 'PO-2024-XXX',
      status: 'DRAFT',
      currentStage: 0,
      createdBy: 'user-1',
      createdByUser: MOCK_USERS.REQUESTER[0],
      createdAt: new Date(now.getTime() - X * 24 * 60 * 60 * 1000),
      updatedAt: new Date(now.getTime() - X * 24 * 60 * 60 * 1000),
      metadata: {
        vendorName: 'Vendor Name',
        vendorId: 'VENDOR-XXX',
        items: [
          {
            id: uuidv4(),
            description: 'Item description',
            quantity: 10,
            unitCost: 1000,
            totalCost: 10000,
          }
        ],
        totalAmount: 10000,
        currency: 'ZMW',
        deliveryDate: new Date(now.getTime() + 30 * 24 * 60 * 60 * 1000),
        specialInstructions: 'Any special instructions',
      },
    }
  ];
}
```

---

## Data Types

### Document Types

All documents extend a base interface:

```typescript
interface WorkflowDocument {
  id: string;
  type: 'PURCHASE_ORDER' | 'REQUISITION' | 'PAYMENT_VOUCHER' | 'GOODS_RECEIVED_NOTE';
  documentNumber: string;
  status: 'DRAFT' | 'SUBMITTED' | 'IN_REVIEW' | 'APPROVED' | 'REJECTED' | 'REVERSED';
  currentStage: number;
  createdBy: string;
  createdByUser?: User;
  createdAt: Date;
  updatedAt: Date;
  metadata: Record<string, any>;
}

interface PurchaseOrder extends WorkflowDocument {
  type: 'PURCHASE_ORDER';
  metadata: {
    vendorName: string;
    vendorId: string;
    items: Array<{
      id: string;
      description: string;
      quantity: number;
      unitCost: number;
      totalCost: number;
    }>;
    totalAmount: number;
    currency: string;
    deliveryDate: Date;
    specialInstructions?: string;
  };
}

interface RequisitionForm extends WorkflowDocument {
  type: 'REQUISITION';
  metadata: {
    department: string;
    amount: number;
    currency: string;
    justification: string;
    items: Array<{
      id: string;
      description: string;
      quantity: number;
      estimatedCost: number;
    }>;
  };
}

interface PaymentVoucher extends WorkflowDocument {
  type: 'PAYMENT_VOUCHER';
  metadata: {
    payeeName: string;
    payeeAccount: string;
    amount: number;
    currency: string;
    invoiceNumber: string;
    paymentDate: string;
    description?: string;
  };
}

interface GoodsReceivedNote extends WorkflowDocument {
  type: 'GOODS_RECEIVED_NOTE';
  metadata: {
    poId: string;
    poNumber: string;
    vendorName: string;
    receivedQuantity: number;
    totalQuantity: number;
    amount: number;
    receivedDate: string;
    warehouseLocation?: string;
  };
}

interface SearchFilters {
  documentNumber: string;
  documentType: 'ALL' | WorkflowDocumentType;
  status: 'ALL' | DocumentStatus;
  startDate: string;
  endDate: string;
}
```

---

## Server Actions

Currently, server actions are **placeholders** since localStorage cannot be accessed from server:

```typescript
// frontend/src/app/_actions/search.ts
export async function searchDocuments(
  filters: SearchFilters,
  page: number = 1,
  limit: number = 10
): Promise<APIResponse<PaginatedResponse<WorkflowDocument>>> {
  const { session } = await verifySession();

  if (!session?.user) {
    return unauthorizedResponse();
  }

  try {
    // Server Actions run on the server and cannot access browser localStorage directly.
    // This endpoint is kept for compatibility but returns empty results.
    // Client-side search using React Query hooks is the recommended approach until
    // proper backend APIs are implemented.

    return {
      success: true,
      message: "Use client-side search for now",
      data: {
        data: [],
        pagination: {
          page,
          limit,
          total: 0,
          totalPages: 0,
        },
      },
      status: 200,
    };
  } catch (error) {
    return handleError(error, "GET", "/search") as any;
  }
}
```

---

## Client-Side State Management

### React Query Hooks

Located in `frontend/src/hooks/use-storage-queries.ts`:

```typescript
export function usePurchaseOrdersAsWorkflowDocumentsQuery() {
  return useQuery({
    queryKey: ['purchase-orders'],
    queryFn: () => {
      const pos = getPurchaseOrders();
      return pos.map(convertToWorkflowDocument);
    },
    staleTime: 0,  // Always consider stale to reflect storage changes
  });
}

// Similar hooks for other document types
export function useRequisitionsAsWorkflowDocumentsQuery() { ... }
export function usePaymentVouchersAsWorkflowDocumentsQuery() { ... }
export function useGrnsAsWorkflowDocumentsQuery() { ... }
```

### Component Integration

```typescript
export function PurchaseOrdersTable({ userId, userRole, refreshTrigger }: Props) {
  // 1. Query data from storage via React Query
  const { data: pos = [], refetch } =
    usePurchaseOrdersAsWorkflowDocumentsQuery();

  // 2. Refetch when refresh is triggered
  useEffect(() => {
    refetch();
  }, [refreshTrigger, refetch]);

  // 3. Use data in component
  return (
    <DataTable
      columns={columns}
      data={pos}
      // ... rest of props
    />
  );
}
```

---

## Data Flow: Creating a Document

1. **User fills form** (`create-po-form.tsx`)
2. **Form validates** (React Hook Form + Zod)
3. **Submits to server action** (`createPurchaseOrder()`)
4. **Server action**:
   - Verifies user session
   - Validates input
   - Calls `savePurchaseOrder()` (storage hook)
   - Returns success response
5. **Client**:
   - Shows success toast
   - Invalidates React Query cache
   - Refetches data (which reads from updated localStorage)
   - Redirects to list view

---

## Data Flow: Search

1. **User enters filters** in search form
2. **Clicks Search button**
3. **Search form calls** `onSearch()` callback with filters
4. **Parent component** (SearchClient):
   - Sets `isSearching = true`
   - Passes filters to TransactionResults
   - Increments refreshTrigger
5. **TransactionResults**:
   - Receives filters and refreshTrigger
   - useEffect triggers mutation on filter/trigger change
   - Mutation calls `performSearch()` function
   - Shows loading skeleton
6. **performSearch()**:
   - Retrieves all 32 documents from storage
   - Filters by documentNumber (partial match)
   - Filters by type (if not "ALL")
   - Filters by status (if not "ALL")
   - Filters by date range
   - Sorts by creation date (newest first)
   - Paginates results
7. **Mutation completes**:
   - `onSuccess()` updates state with results
   - Calls `onSearchComplete()` to clear loading state
   - Displays results in DataTable with pagination

---

## Persistence

### Logout Behavior

**Important**: localStorage is **NOT cleared on logout**

```typescript
// frontend/src/lib/auth.ts
export async function logout() {
  // Clear only session data, NOT localStorage
  // This allows users to see their previously created documents
  // when they log back in

  await signOut({ redirect: true, callbackUrl: '/login' });
}
```

This is intentional - documents created by the user persist across sessions.

### Manual Data Reset

To reset all data for testing:

```typescript
// In browser console:
import { resetStorage } from '@/lib/storage/init';
resetStorage();  // Clears all storage and reinitializes with seed data
```

Or in code:

```typescript
// Somewhere in a debug component or admin panel
import { clearAllData, initializeStorage } from '@/lib/storage';

function resetDataButton() {
  return (
    <button onClick={() => {
      clearAllData();
      initializeStorage();
    }}>
      Reset All Data
    </button>
  );
}
```

---

## Current Limitations

### Single User
- No user isolation of data
- All created documents visible to all "users"
- User ID is just for audit trail

### Single Device
- Data stored only in browser localStorage
- Not synchronized across devices
- Cleared if browser cache is cleared

### Search Limitations
- All filtering happens in client
- No database indexes
- Search on 30+ documents is instant but will slow with larger datasets
- No full-text search capabilities

### No Real Persistence
- Data lost if localStorage quota exceeded
- localStorage has 5-10MB limit
- No backup or recovery mechanism

### No Concurrent Users
- No multi-user approval workflows
- No task assignment or delegation
- No conflict resolution for simultaneous edits

---

## Browser Compatibility

Storage works in all modern browsers:
- ✅ Chrome/Chromium
- ✅ Firefox
- ✅ Safari
- ✅ Edge
- ✅ Mobile browsers (with 5MB localStorage limit)

Not supported:
- ❌ Private/Incognito mode (in some browsers)
- ❌ Disabled cookies/storage

---

## Debugging

### Console Logs

The search component includes detailed logging:

```
🔍 Search starting with filters: {...}
📦 Storage data: { pos: 10, reqs: 7, pvs: 9, grns: 6 }
📄 All documents: 32 [Array(32)]
🔄 Converting document: {...}
✅ Converted document createdAt: 2024-12-01T10:30:00.000Z

🔍 Evaluating PO-2024-001:
  ✓ documentNumber filter: "PO-2024" found in "PO-2024-001"
  ✓ type filter: doc.type="PURCHASE_ORDER" === filters.documentType="ALL"
  ✓ status filter: doc.status="APPROVED" === filters.status="ALL"
  ✅ Document passed all filters: PO-2024-001

🔎 After filtering: 15 documents from 32
📊 Setting search results: {documents: 15, total: 15, totalPages: 2}
```

### Storage Inspector

In browser DevTools:

```javascript
// View all storage
localStorage.getItem('liyali_purchase_orders')
localStorage.getItem('liyali_requisitions')
// etc.

// Get statistics
const { getStorageStats } = await import('@/lib/storage');
getStorageStats();

// Export all data
const { exportStorageAsJSON } = await import('@/lib/storage');
console.log(exportStorageAsJSON());
```

---

## Next Steps for Phase 12

1. **Backend Setup**
   - Create Node.js/Express server
   - Set up PostgreSQL database
   - Implement Prisma ORM

2. **API Implementation**
   - Replace localStorage functions with HTTP calls
   - Implement all endpoints from API-ENDPOINTS.md
   - Add authentication (JWT/OAuth2)

3. **Migration**
   - Migrate localStorage seed data to database
   - Create data import/export tools
   - Update React Query hooks to use API

4. **Testing**
   - Add integration tests
   - Test multi-user workflows
   - Test concurrent approvals

---
