# Search System - Deep Dive Audit

## Executive Summary

The **Search System** is a comprehensive transaction search interface that allows users to find requisitions, purchase orders, payment vouchers, and goods received notes (GRNs) across the entire platform. It combines client-side filtering UI with server-side search logic, real-time table pagination, and document preview/download capabilities.

**Key Features:**
- ✅ Multi-document type search (4 document types supported)
- ✅ Multi-filter support (document number, type, status, date range)
- ✅ Real-time pagination (customizable page size)
- ✅ Sortable columns (document number, created date)
- ✅ View/Download actions for each result
- ✅ Deduplication (handles documents appearing in multiple queries)
- ✅ User-scoped results (documents created by user + pending approvals)

---

## How It Works: End-to-End Flow

### 1. User Navigation to Search Page

```
Browser URL: http://localhost:3001/search
↓
Next.js Router matches: src/app/(private)/(main)/search/page.tsx
↓
Server Component checks authentication via getCurrentUser()
↓
If not authenticated → Redirect to /login
If authenticated → Pass userId and userRole to SearchClient
```

**Code Reference:** [search/page.tsx](src/app/(private)/(main)/search/page.tsx:1-20)

### 2. Initial Page Render

```
SearchClient Component Initializes
↓
State Variables Created:
  • filters: {
      documentNumber: '',
      documentType: 'ALL',
      status: 'ALL',
      startDate: '',
      endDate: ''
    }
  • refreshTrigger: 0           (counter for re-triggering searches)
  • isSearching: false          (button loading state)
↓
Page Header Rendered: "Search Transactions"
↓
SearchForm Component Rendered (empty state with no results yet)
↓
TransactionResults Component Rendered (with useEffect hook waiting for filters)
```

**Code Reference:** [search-client.tsx](src/app/(private)/(main)/search/_components/search-client.tsx:14-50)

### 3. User Fills Search Form

User interacts with SearchForm component:

```
┌─────────────────────────────────────────────────────────────┐
│                    Search Filters Card                      │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│ Row 1:                                                      │
│  📝 Document Number Input (e.g., "REQ-2024-001")           │
│  📋 Document Type Dropdown (Req/PO/PV/GRN/All)            │
│                                                              │
│ Row 2:                                                      │
│  📊 Status Dropdown (Draft/Submitted/In Review/Approved)   │
│  📅 Start Date Picker                                       │
│  📅 End Date Picker                                        │
│                                                              │
│ Row 3:                                                      │
│  [Reset Button]  [Search Button with icon]                │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

**State Updates (Real-time):**
```typescript
// As user types in document number
setDocumentNumber('REQ-2024')  // Updates state immediately

// As user selects from dropdown
setDocumentType('REQUISITION')  // Updates state

// As user selects dates
setStartDate(new Date('2024-01-01'))
setEndDate(new Date('2024-12-31'))
```

**Code Reference:** [search-form.tsx](src/app/(private)/(main)/search/_components/search-form.tsx:39-177)

### 4. User Clicks Search Button

```
handleSubmit() function executes:
↓
event.preventDefault() - prevents form submission
↓
Build SearchFilters object:
{
  documentNumber: 'REQ-2024',
  documentType: 'REQUISITION',
  status: 'APPROVED',
  startDate: '2024-01-01',
  endDate: '2024-12-31'
}
↓
Call onSearch callback with filters
↓
SearchClient's handleSearch executes:
  1. setFilters(newFilters)        // Update parent filters state
  2. setIsSearching(true)          // Show loading in form
  3. setRefreshTrigger(++counter)  // Trigger useEffect in results
```

**Code Reference:** [search-form.tsx](src/app/(private)/(main)/search/_components/search-form.tsx:46-55)

### 5. Server-Side Search Execution

The TransactionResults component useEffect detects filter change:

```typescript
useEffect(() => {
  fetchDocuments();
}, [filters, pagination.page, pagination.limit, refreshTrigger]);
```

This calls the server action:

```
searchDocuments(filters, page=1, limit=10)  // Server Action
↓
Server verifies session/authentication
↓
Fetch documents from TWO sources:
  1. getDocumentsByCreator(userId, 1, 1000)  // All docs user created
  2. getPendingApprovals(userRole)            // All docs awaiting this user's approval
↓
Combine both arrays into allDocuments[]
↓
Remove duplicates using Map<id, doc>
↓
Apply filters to uniqueDocuments:
  - documentNumber: Case-insensitive partial match
  - documentType: Exact match or 'ALL'
  - status: Exact match or 'ALL'
  - startDate: createdAt >= startDate
  - endDate: createdAt <= endDate (with 23:59:59)
↓
Sort by createdAt descending (newest first)
↓
Paginate: Calculate totalPages, extract slice [skip:skip+limit]
↓
Return APIResponse with data and pagination metadata
```

**Code Reference:** [search.ts](src/app/_actions/search.ts:8-102)

### 6. Data Fetching Details

#### 6a. getDocumentsByCreator()

```typescript
// Signature
export async function getDocumentsByCreator(
  userId: string,
  page: number = 1,
  limit: number = 10
): Promise<APIResponse<PaginatedResponse<WorkflowDocument>>>

// Implementation
documents = Array.from(documentStore.values()).filter(
  (doc) => doc.createdBy === userId
);
// Returns paginated results
```

**Code Reference:** [workflow.ts](src/app/_actions/workflow.ts:242-279)

**Storage:** `documentStore` is a `Map<string, WorkflowDocument>` (in-memory)

**Content:** All document types (Requisition, PurchaseOrder, PaymentVoucher, GoodsReceivedNote)

#### 6b. getPendingApprovals()

```typescript
// Signature
export async function getPendingApprovals(
  userRole: string
): Promise<APIResponse<WorkflowDocument[]>>

// Implementation
pendingDocs = Array.from(documentStore.values())
  .filter((doc) => doc.status === "IN_REVIEW")
  .filter((doc) => {
    const approvers = approversStore.get(doc.id) || [];
    return approvers.some(
      (a) => a.stepOrder === doc.currentStage && a.role === userRole
    );
  });
```

**Code Reference:** [workflow.ts](src/app/_actions/workflow.ts:482-513)

**Logic:** Find docs where:
1. Status is "IN_REVIEW"
2. Current approver for current stage has the user's role

**Example:** If you're a CFO and document needs CFO approval at step 2, it shows up

### 7. Filtering Logic (Client-Side Post-Processing)

Once search results come back from server, they're further processed:

```typescript
// Deduplication by ID
const uniqueMap = new Map<string, WorkflowDocument>();
allDocuments.forEach(doc => uniqueMap.set(doc.id, doc));
const uniqueDocuments = Array.from(uniqueMap.values());

// Filter application
let filtered = uniqueDocuments.filter((doc) => {
  // Document number: case-insensitive substring match
  if (filters.documentNumber &&
      !doc.documentNumber.toLowerCase().includes(
        filters.documentNumber.toLowerCase()
      )) {
    return false;
  }

  // Document type: exact match or 'ALL'
  if (filters.documentType !== 'ALL' &&
      doc.type !== filters.documentType) {
    return false;
  }

  // Status: exact match or 'ALL'
  if (filters.status !== 'ALL' &&
      doc.status !== filters.status) {
    return false;
  }

  // Start date: inclusive
  if (filters.startDate) {
    const startDate = new Date(filters.startDate);
    if (doc.createdAt < startDate) {
      return false;
    }
  }

  // End date: inclusive (entire day)
  if (filters.endDate) {
    const endDate = new Date(filters.endDate);
    endDate.setHours(23, 59, 59, 999);
    if (doc.createdAt > endDate) {
      return false;
    }
  }

  return true;
});

// Sort by newest first
filtered.sort((a, b) =>
  new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
);

// Pagination slicing
const skip = (page - 1) * limit;
const paginatedData = filtered.slice(skip, skip + limit);
```

**Code Reference:** [search.ts](src/app/_actions/search.ts:40-77)

### 8. Results Table Rendering

Once data is received in TransactionResults:

```typescript
setDocuments(result.data.data);
setPagination(result.data.pagination);
setIsLoading(false);
```

TanStack React Table is initialized with columns:

| Column | Type | Sortable | Behavior |
|--------|------|----------|----------|
| **Document #** | String | ✅ Yes | Clickable, shows doc number |
| **Type** | String | ❌ No | Maps REQUISITION→"Requisition" |
| **Status** | Badge | ❌ No | Colored badge (Draft=outline, Approved=default, Rejected=destructive) |
| **Created** | Date | ✅ Yes | Shows date + time, formatted locale |
| **Actions** | Buttons | ❌ No | View button + Download button |

**Code Reference:** [transaction-results.tsx](src/app/(private)/(main)/search/_components/transaction-results.tsx:128-227)

#### 8a. Status Color Mapping

```typescript
const STATUS_COLORS: Record<string, string> = {
  DRAFT: "outline",
  SUBMITTED: "secondary",
  IN_REVIEW: "default",
  APPROVED: "default",
  REJECTED: "destructive",
  REVERSED: "secondary",
};
```

#### 8b. Actions Column

When user clicks "View":

```typescript
// Determine URL based on document type
const typeSlug = {
  REQUISITION: "requisitions",
  PURCHASE_ORDER: "purchase-orders",
  PAYMENT_VOUCHER: "payment-vouchers",
  GOODS_RECEIVED_NOTE: "grn",
}[doc.type];

// Navigate to document details page
router.push(`/${typeSlug}/${doc.id}`);
```

**Resulting URLs:**
- Requisition → `/requisitions/{id}`
- Purchase Order → `/purchase-orders/{id}`
- Payment Voucher → `/payment-vouchers/{id}`
- GRN → `/grn/{id}`

**Code Reference:** [transaction-results.tsx](src/app/(private)/(main)/search/_components/transaction-results.tsx:200-206)

### 9. Download Flow

When user clicks "Download" button on a result:

```
DownloadButton Component
↓
User clicks Download button
↓
handleDownload() executes:
  1. setIsLoading(true)        // Show spinner
  2. Call downloadDocumentPDF(documentId)  // Server action
↓
Server Action Execution:
  • Verify session
  • Call getDocument(documentId) to verify doc exists
  • Generate mock download URL: `/api/documents/{id}/download`
  • Return { success: true, downloadUrl: "..." }
↓
Client receives response:
  1. Create temporary <a> element
  2. Set href = downloadUrl
  3. Set download = "{documentNumber}.pdf"
  4. Append to DOM, click(), remove from DOM
↓
Browser triggers PDF download
↓
setIsLoading(false) - button returns to normal
```

**Code Reference:** [download-button.tsx](src/app/(private)/(main)/search/_components/download-button.tsx:13-56)

**Note:** Currently returns mock download URL. In production, would need actual PDF generation.

### 10. Pagination Controls

Below results table:

```
Showing 1 to 10 of 47 documents
[< Previous] Page 1 of 5 [Next >]
```

**Pagination Logic:**

```typescript
// User clicks "Next" button
setPagination((p) => ({
  ...p,
  page: Math.min(p.page + 1, p.totalPages)  // Increment but don't exceed max
}));

// User clicks "Previous" button
setPagination((p) => ({
  ...p,
  page: Math.max(p.page - 1, 1)  // Decrement but not below 1
}));

// This triggers useEffect because pagination.page changed
// useEffect calls fetchDocuments() again with new page number
```

**Code Reference:** [transaction-results.tsx](src/app/(private)/(main)/search/_components/transaction-results.tsx:305-350)

**Disabled States:**
- Previous button disabled when: `page === 1 || isLoading`
- Next button disabled when: `page >= totalPages || isLoading`

---

## Data Flow Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                     Search Page Load                            │
│  src/app/(private)/(main)/search/page.tsx (Server)             │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ▼
        ┌────────────────────────────────────┐
        │     SearchClient (Client)          │
        │  - Manages filters state           │
        │  - Passes to children              │
        └─────┬──────────────────┬───────────┘
              │                  │
        ┌─────▼──────┐    ┌─────▼──────────────────┐
        │SearchForm  │    │ TransactionResults     │
        │(Input UI)  │    │  (Results Table)       │
        └─────┬──────┘    └──────────┬─────────────┘
              │                      │
              │ onClick Search       │ useEffect (filters change)
              │                      │
              ▼                      ▼
        ┌──────────────────────────────────────┐
        │  searchDocuments() (Server Action)   │
        │  src/app/_actions/search.ts         │
        └────────┬─────────────────────────────┘
                 │
    ┌────────────┼────────────┐
    │            │            │
    ▼            ▼            ▼
┌─────────────────────────────────────┐
│  getDocumentsByCreator()            │
│  - Query documentStore by userId    │
│  - Returns user's created docs      │
└─────────────────────────────────────┘

┌─────────────────────────────────────┐
│  getPendingApprovals()              │
│  - Query documentStore by role      │
│  - Filter by status=IN_REVIEW       │
│  - Check approvers for user's role  │
└─────────────────────────────────────┘
    │
    │ Combine results
    │
    ▼
┌──────────────────────────────────────────┐
│  Apply Filters (client-side logic)       │
│  - Dedup by ID                           │
│  - Filter by documentNumber (substring)  │
│  - Filter by type (exact)                │
│  - Filter by status (exact)              │
│  - Filter by date range                  │
│  - Sort by createdAt desc                │
│  - Paginate: limit & page offset         │
└─────────────┬──────────────────────────┘
              │
              ▼
    ┌────────────────────────┐
    │ Return APIResponse     │
    │ - data: [documents]    │
    │ - pagination: {...}    │
    └────────────────────────┘
              │
              ▼
        ┌──────────────────────────┐
        │  TransactionResults      │
        │  - setDocuments()        │
        │  - setPagination()       │
        │  - Render table          │
        └──────────────────────────┘
```

---

## Component Hierarchy

```
SearchPage (Server Component)
├── metadata: { title: "Search Transactions" }
├── getCurrentUser() → redirect if not authenticated
└── <SearchClient userId={user.id} userRole={user.role} />
    │
    ├── <PageHeader title="Search Transactions" />
    │
    ├── <SearchForm onSearch={handleSearch} isSearching={isSearching} />
    │   ├── <Input /> - Document Number
    │   ├── <SelectField /> - Document Type
    │   ├── <SelectField /> - Status
    │   ├── <DatePicker /> - Start Date
    │   ├── <DatePicker /> - End Date
    │   └── <Button /> - Submit/Reset
    │
    └── <TransactionResults filters={filters} refreshTrigger={refreshTrigger} />
        ├── useEffect → searchDocuments() (Server Action)
        ├── <Table>
        │   ├── Headers: Document #, Type, Status, Created, Actions
        │   └── Rows: Mapped from documents[]
        │       ├── DocumentNumber (sortable, clickable)
        │       ├── Type (badge)
        │       ├── Status (colored badge)
        │       ├── CreatedAt (date/time)
        │       └── Actions
        │           ├── <Button onClick={router.push(`/{type}/{id}`)} />
        │           └── <DownloadButton documentId={id} />
        │
        └── Pagination
            ├── [< Previous] [Page X of Y] [Next >]
            └── Showing X to Y of Z documents
```

---

## State Management

### SearchClient State

```typescript
const [filters, setFilters] = useState<SearchFilters>({
  documentNumber: '',
  documentType: 'ALL',
  status: 'ALL',
  startDate: '',
  endDate: '',
});

const [refreshTrigger, setRefreshTrigger] = useState(0);
// Counter that increments when search button clicked
// Used to trigger TransactionResults useEffect

const [isSearching, setIsSearching] = useState(false);
// Passed to SearchForm to disable buttons during search
```

### TransactionResults State

```typescript
const [documents, setDocuments] = useState<WorkflowDocument[]>([]);
// Array of search results

const [pagination, setPagination] = useState({
  page: 1,
  limit: 10,
  total: 0,
  totalPages: 1,
});
// Pagination metadata from server

const [isLoading, setIsLoading] = useState(false);
// Shows loading spinner while fetching

const [sorting, setSorting] = useState<SortingState>([]);
// TanStack table state for column sorting

const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
// TanStack table state for column filtering

const [columnVisibility, setColumnVisibility] = useState<VisibilityState>({});
// TanStack table state for column visibility
```

### SearchForm State

```typescript
const [documentNumber, setDocumentNumber] = useState("");
const [documentType, setDocumentType] = useState("ALL");
const [status, setStatus] = useState("ALL");
const [startDate, setStartDate] = useState<Date | undefined>(undefined);
const [endDate, setEndDate] = useState<Date | undefined>(undefined);
```

---

## Type Definitions

### SearchFilters

```typescript
interface SearchFilters {
  documentNumber: string;      // Partial match, case-insensitive
  documentType: 'ALL' | WorkflowDocumentType;  // REQUISITION | PURCHASE_ORDER | PAYMENT_VOUCHER | GOODS_RECEIVED_NOTE
  status: 'ALL' | DocumentStatus;  // DRAFT | SUBMITTED | IN_REVIEW | APPROVED | REJECTED | REVERSED
  startDate: string;           // ISO date string (YYYY-MM-DD)
  endDate: string;             // ISO date string (YYYY-MM-DD)
}
```

### WorkflowDocument

```typescript
interface WorkflowDocument {
  id: string;
  documentNumber: string;      // REQ-2024-0001, PO-2024-0001, etc.
  type: WorkflowDocumentType;
  status: DocumentStatus;
  createdBy: string;           // User ID who created this document
  createdAt: Date;
  updatedAt: Date;
  // ... additional fields specific to document type
}
```

### PaginatedResponse

```typescript
interface PaginatedResponse<T> {
  data: T[];
  pagination: {
    page: number;
    limit: number;
    total: number;
    totalPages: number;
  };
}
```

### APIResponse

```typescript
interface APIResponse<T> {
  success: boolean;
  message: string;
  data?: T;
  status: number;
  statusText: string;
}
```

---

## Key Features

### 1. Dual Data Source Search

Search results combine two independent queries:

| Source | Query | Records | Purpose |
|--------|-------|---------|---------|
| **Created Documents** | `documentStore.filter(doc => doc.createdBy === userId)` | All docs user created | Users see their own submissions |
| **Pending Approvals** | `documentStore.filter(doc => doc.status === IN_REVIEW && approvers includes user's role)` | All docs awaiting user | Users see what needs their approval |

**Benefit:** Single search shows both "my documents" and "documents waiting for me"

### 2. Deduplication

```typescript
// If a document appears in both queries (created by user AND awaiting their approval)
// it should only appear once in results

const uniqueMap = new Map<string, WorkflowDocument>();
allDocuments.forEach(doc => uniqueMap.set(doc.id, doc));
// Map automatically deduplicates by key (id)

const uniqueDocuments = Array.from(uniqueMap.values());
```

### 3. Date Range Filtering

```typescript
// Start date: Inclusive (at midnight)
if (filters.startDate) {
  const startDate = new Date(filters.startDate);  // 00:00:00
  if (doc.createdAt < startDate) return false;
}

// End date: Inclusive (through 23:59:59)
if (filters.endDate) {
  const endDate = new Date(filters.endDate);
  endDate.setHours(23, 59, 59, 999);  // Extend to end of day
  if (doc.createdAt > endDate) return false;
}
```

**Example:** If user selects Jan 1 to Jan 31:
- Includes all docs from Jan 1 00:00:00 through Jan 31 23:59:59
- Excludes docs from Feb 1 onwards

### 4. Case-Insensitive Substring Matching

```typescript
// User can type "req" and find "REQ-2024-001"
if (filters.documentNumber &&
    !doc.documentNumber.toLowerCase().includes(
      filters.documentNumber.toLowerCase()
    )) {
  return false;
}
```

### 5. Sorting

```typescript
// Results always sorted newest first
filtered.sort((a, b) =>
  new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
);

// Table columns also support sorting (TanStack React Table)
// Clicking "Document #" or "Created" toggles ASC/DESC
```

### 6. Pagination

Default: 10 results per page

```typescript
const skip = (page - 1) * limit;
const paginatedData = filtered.slice(skip, skip + limit);
```

Users can navigate pages without re-fetching all data (slicing already in-memory)

---

## User Interaction Scenarios

### Scenario 1: Search by Document Type

**User Action:** Select "Requisitions" from type dropdown and click Search

```
1. SearchForm state updates: documentType = 'REQUISITION'
2. User clicks Search
3. handleSearch() called
4. SearchClient.handleSearch updates parent filters
5. refreshTrigger increments
6. TransactionResults useEffect fires
7. searchDocuments() called on server
8. Results filtered to only REQUISITION type
9. Table renders only requisitions
```

**Result:** User sees list of all requisitions they created or need to approve

---

### Scenario 2: Filter by Date Range

**User Action:** Select "2024-01-01" to "2024-12-31" and click Search

```
1. setStartDate(2024-01-01), setEndDate(2024-12-31)
2. Click Search
3. Server receives: startDate='2024-01-01', endDate='2024-12-31'
4. Filtering logic:
   - startDate boundary: Jan 1 00:00:00
   - endDate boundary: Dec 31 23:59:59 (extended by setHours)
5. Only docs created within 2024 returned
```

**Result:** User sees documents from the entire year 2024

---

### Scenario 3: Multi-Filter Search

**User Action:**
- Document Number: "REQ"
- Type: "Requisition"
- Status: "Approved"
- Date: Jan 1 - Dec 31, 2024

```
1. All filters set in form
2. Click Search
3. Filters passed to searchDocuments()
4. Filtering chain:
   a. documentNumber filter: .includes('REQ') → subset1
   b. documentType filter: === 'REQUISITION' → subset2
   c. status filter: === 'APPROVED' → subset3
   d. startDate filter: >= 2024-01-01 → subset4
   e. endDate filter: <= 2024-12-31 23:59:59 → subset5
5. Final result: intersection of all conditions
```

**Result:** Only approved requisitions from 2024 with "REQ" in number

---

### Scenario 4: Pagination

**User Action:** Click "Next" button after first search returns 15 results

```
1. First search: page=1, limit=10
   - Returns 10 results
   - pagination.totalPages = 2 (15 results ÷ 10 per page = 1.5 → 2 pages)

2. User clicks Next
3. setPagination(p => ({ ...p, page: 2 }))
4. useEffect re-runs with page=2
5. searchDocuments() called with page=2, limit=10
6. Server skips first 10: slice(10, 20)
7. Returns results 11-15
8. Table re-renders with new data
9. Pagination shows "Page 2 of 2"
10. Next button now disabled (page >= totalPages)
```

**Result:** User sees documents 11-15 on page 2

---

### Scenario 5: View Document Details

**User Action:** Click "View" button on a purchase order result

```
1. Row data: { id: 'po-123', type: 'PURCHASE_ORDER', ... }
2. Click View button
3. onClick handler executes:
   const typeSlug = 'purchase-orders'
   router.push('/purchase-orders/po-123')
4. Browser navigates to /purchase-orders/po-123
5. Next.js loads: src/app/(private)/(main)/purchase-orders/[id]/page.tsx
6. Server fetches PO details with ID 'po-123'
7. Detail page renders
```

**Result:** User sees full purchase order details page

---

### Scenario 6: Download Document PDF

**User Action:** Click "Download" button on a requisition result

```
1. DownloadButton component, documentId='req-456'
2. setIsLoading(true) → button shows spinner
3. Call downloadDocumentPDF('req-456')
4. Server action executes:
   a. Verify session
   b. Call getDocument('req-456')
   c. If found: return { success: true, downloadUrl: '/api/documents/req-456/download' }
5. Client receives response
6. Create <a href="/api/documents/req-456/download" download="REQ-2024-0001.pdf">
7. Simulate click → browser triggers download
8. setIsLoading(false) → button returns to normal
```

**Result:** PDF file "REQ-2024-0001.pdf" downloads to user's computer

---

## Reset Button Behavior

**User Action:** Click "Reset" button in search form

```
1. handleReset() executes
2. Clear all form state:
   - setDocumentNumber('')
   - setDocumentType('ALL')
   - setStatus('ALL')
   - setStartDate(undefined)
   - setEndDate(undefined)
3. Call onSearch with all empty/default values
4. SearchClient.handleSearch receives empty filters
5. setFilters to all defaults
6. refreshTrigger increments
7. TransactionResults useEffect fires
8. searchDocuments() called with empty filters
9. All documents returned (no filtering)
10. Results show all user's documents + pending approvals
```

**Result:** Search form is cleared, results show everything

---

## Storage & Data Layer

### Document Storage

```typescript
// In-memory Map (simulates database)
const documentStore = new Map<string, WorkflowDocument>();

// Documents added by:
// - createRequisition() → adds to documentStore
// - createPurchaseOrder() → adds to documentStore
// - createPaymentVoucher() → adds to documentStore
// - createGoodsReceivedNote() → adds to documentStore
```

### Approvers Storage

```typescript
const approversStore = new Map<string, Approver[]>();

// Each approval relationship:
// Map Key: documentId
// Map Value: Array of { stepOrder, role, userId, status, ... }
```

### Approval Logs Storage

```typescript
const approvalLogsStore = new Map<string, ApprovalLogEntry[]>();

// Audit trail of all approvals:
// Map Key: documentId
// Map Value: Array of { approver, action, timestamp, comments }
```

---

## Performance Considerations

### 1. Memory Usage

**Concern:** Fetching up to 1000 documents with `limit=1000` in getDocumentsByCreator

```typescript
const createdResult = await getDocumentsByCreator(session.user.id, 1, 1000);
```

**Mitigation:** Using in-memory storage, this is acceptable for MVP with limited users

**Production:** Would paginate properly and use database with indexing

### 2. Search Complexity

**Time Complexity:** O(n) where n = total documents

```typescript
Array.from(documentStore.values()).filter(...)  // O(n)
allDocuments.forEach(...)                         // O(n)
filtered.filter(...)                              // O(n)
filtered.sort(...)                                // O(n log n)
filtered.slice(...)                               // O(m) where m = page size
```

**Optimization:** With database, could use indexes on createdBy, status, type, createdAt

### 3. Re-rendering

**TanStack React Table** provides built-in optimizations:
- Only re-renders changed rows
- Memoizes column definitions
- Efficient sorting/filtering

---

## Error Handling

### Missing User Session

```typescript
const session = await auth();
if (!session?.user) {
  return unauthorizedResponse();  // Returns 401 error
}
```

### Document Not Found (Download)

```typescript
const result = await getDocument(documentId);
if (!result.success || !result.data) {
  return {
    success: false,
    message: 'Document not found',
  };
}
```

### Network/Server Errors

```typescript
try {
  const result = await searchDocuments(filters, ...);
  if (result.success) {
    // Use results
  } else {
    // Handle error from server
  }
} catch (error) {
  console.error("Failed to fetch documents:", error);
  // UI shows "No documents found"
}
```

---

## Testing Scenarios

### Unit Tests

```typescript
test('searchDocuments filters by documentNumber', () => {
  const filters = { documentNumber: 'REQ', ... };
  const result = await searchDocuments(filters, 1, 10);
  result.data.data.forEach(doc => {
    expect(doc.documentNumber).toContain('REQ');
  });
});

test('searchDocuments deduplicates results', () => {
  // Mock: document appears in both created + pending
  const result = await searchDocuments(filters, 1, 1000);
  const ids = result.data.data.map(d => d.id);
  const uniqueIds = new Set(ids);
  expect(ids.length).toBe(uniqueIds.size);
});

test('date range filtering is inclusive', () => {
  const filters = {
    startDate: '2024-01-01',
    endDate: '2024-01-31',
  };
  const result = await searchDocuments(filters, 1, 1000);
  result.data.data.forEach(doc => {
    const docDate = new Date(doc.createdAt);
    expect(docDate.getTime()).toBeGreaterThanOrEqual(
      new Date('2024-01-01').getTime()
    );
    expect(docDate.getTime()).toBeLessThanOrEqual(
      new Date('2024-01-31 23:59:59').getTime()
    );
  });
});
```

### Integration Tests

```typescript
test('complete search flow: form → server → table', async () => {
  // 1. Render SearchClient
  const { getByRole, getByDisplayValue } = render(<SearchClient ... />);

  // 2. Fill form
  fireEvent.change(getByRole('combobox', { name: /document type/i }), {
    target: { value: 'REQUISITION' },
  });

  // 3. Click search
  fireEvent.click(getByRole('button', { name: /search/i }));

  // 4. Wait for results
  await waitFor(() => {
    expect(getByRole('table')).toBeInTheDocument();
  });

  // 5. Verify table has requisitions only
  const rows = getByRole('table').querySelectorAll('tbody tr');
  rows.forEach(row => {
    expect(row).toHaveTextContent('Requisition');
  });
});
```

### E2E Tests

```typescript
test('search for approved requisitions from 2024', async () => {
  await page.goto('/search');

  // Fill form
  await page.fill('input[placeholder*="Document Number"]', 'REQ');
  await page.selectOption('select[name="documentType"]', 'REQUISITION');
  await page.selectOption('select[name="status"]', 'APPROVED');
  await page.fill('input[name="startDate"]', '2024-01-01');
  await page.fill('input[name="endDate"]', '2024-12-31');

  // Search
  await page.click('button:has-text("Search")');

  // Verify results
  const rows = await page.$$('table tbody tr');
  expect(rows.length).toBeGreaterThan(0);

  // Click view on first result
  await rows[0].$('button:has-text("View")').click();

  // Should navigate to detail page
  expect(page.url()).toContain('/requisitions/');
});
```

---

## Production Readiness Checklist

- [ ] Replace in-memory Map with proper database
- [ ] Add query indexes on createdBy, status, type, createdAt fields
- [ ] Implement proper pagination at database level (not fetching 1000 records)
- [ ] Add caching layer (Redis) for frequently searched filters
- [ ] Implement actual PDF generation and download
- [ ] Add rate limiting to prevent search abuse
- [ ] Add search result caching with invalidation logic
- [ ] Implement full-text search for document number (instead of substring)
- [ ] Add analytics: track common search patterns
- [ ] Add advanced filters (document amount, approver name, etc.)
- [ ] Implement saved searches / search history
- [ ] Add bulk actions (export, batch approval, etc.)
- [ ] Implement search result filtering by department/division
- [ ] Add role-based search visibility

---

## Summary

The Search System is a well-architected feature that:

✅ **Dual Data Source:** Combines user-created documents with pending approvals
✅ **Flexible Filtering:** Supports 5 independent filter dimensions
✅ **User-Friendly UI:** Gradient card design, clear form layout, intuitive controls
✅ **Efficient Pagination:** Client-side slicing of in-memory results
✅ **Deduplication:** Prevents duplicate results when doc appears in multiple queries
✅ **Document Navigation:** Quick links to view/download any result
✅ **Responsive Table:** TanStack React Table for sorting/filtering/pagination

🚀 Ready for: Database integration, advanced filtering, search analytics
