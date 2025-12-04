# Data Source Architecture - Where Does Search Data Come From?

## TL;DR

**The search results come from IN-MEMORY MOCK DATA stored in JavaScript Maps.**

The data is NOT coming from localStorage, a database, or external API. It's entirely simulated within the application's memory during runtime.

---

## Architecture Overview

```
┌──────────────────────────────────────────────────────────────┐
│           SEARCH DATA SOURCE ARCHITECTURE                    │
└──────────────────────────────────────────────────────────────┘
                             │
            ┌────────────────┼────────────────┐
            │                │                │
            ▼                ▼                ▼
    ┌──────────────┐  ┌──────────────┐  ┌──────────────┐
    │  Mock Data   │  │  In-Memory   │  │  Runtime    │
    │  Generator   │  │  Maps        │  │  Store      │
    │              │  │  (No DB)     │  │  (No Persist)
    └──────────────┘  └──────────────┘  └──────────────┘
         │                  │                  │
         └──────────────────┼──────────────────┘
                            │
                    ┌───────▼────────┐
                    │ Search Query   │
                    │ Results        │
                    └────────────────┘
```

---

## Data Flow Stack

### Layer 1: Mock Data Generation

**File:** `src/lib/mock-data.ts`

This file contains factory functions that CREATE fake data objects:

```typescript
export function createMockRequisitionForm(options?: Partial<RequisitionForm>): RequisitionForm {
  return {
    id: uuidv4(),
    documentNumber: generateDocumentNumber('REQUISITION'),
    type: 'REQUISITION',
    status: options?.status || 'DRAFT',
    // ... all fields populated with random/templated values
  };
}

export function createMockPurchaseOrder(options?: Partial<PurchaseOrder>): PurchaseOrder {
  return {
    id: uuidv4(),
    documentNumber: generateDocumentNumber('PURCHASE_ORDER'),
    type: 'PURCHASE_ORDER',
    status: options?.status || 'DRAFT',
    // ... all fields populated with random/templated values
  };
}

export function createMockPaymentVoucher(options?: Partial<PaymentVoucher>): PaymentVoucher {
  return {
    id: uuidv4(),
    documentNumber: generateDocumentNumber('PAYMENT_VOUCHER'),
    type: 'PAYMENT_VOUCHER',
    status: options?.status || 'DRAFT',
    // ... all fields populated with random/templated values
  };
}
```

**Also contains:** Mock user profiles, mock approval chain data, mock attachments

---

### Layer 2: In-Memory Maps (The "Database")

**File:** `src/lib/workflow-stores.ts`

These are simple JavaScript Maps that hold all documents in RAM:

```typescript
// THE ENTIRE "DATABASE" - just Maps in memory!
export const documentStore = new Map<string, WorkflowDocument>();
export const approversStore = new Map<string, Approver[]>();
export const approvalLogsStore = new Map<string, ApprovalLogEntry[]>();
export const attachmentsStore = new Map<string, Attachment[]>();

export let isInitialized = false;
```

**Key Characteristics:**
- ✅ No external database
- ✅ No localStorage persistence
- ✅ No file-based storage
- ✅ Pure in-memory JavaScript data structure
- ❌ **Data is LOST when app restarts**
- ❌ **Not shared between multiple instances**
- ❌ **Single-server only**

---

### Layer 3: Initialization with Sample Data

**File:** `src/lib/workflow-initialization.ts`

When the app starts, this runs automatically:

```typescript
export function initializeSampleData() {
  if (isInitialized) return;  // Only run once

  const statuses: DocumentStatus[] =
    ['DRAFT', 'SUBMITTED', 'IN_REVIEW', 'APPROVED', 'REJECTED', 'REVERSED'];

  const documentTypes: WorkflowDocumentType[] =
    ['REQUISITION', 'PURCHASE_ORDER', 'PAYMENT_VOUCHER'];

  // Create 25 sample documents
  for (let i = 0; i < 25; i++) {
    const status = statuses[i % statuses.length];
    const type = documentTypes[i % documentTypes.length];
    const daysAgo = Math.floor(i / 2);
    const createdDate = new Date(Date.now() - daysAgo * 24 * 60 * 60 * 1000);

    let doc: WorkflowDocument;

    // Create mock doc based on type
    switch (type) {
      case 'PURCHASE_ORDER':
        doc = createMockPurchaseOrder({
          status,
          currentStage: status === 'DRAFT' ? 0 : Math.min(i % 4, 3),
          createdAt: createdDate,
          updatedAt: createdDate,
          createdBy: MOCK_USERS.REQUESTER[i % MOCK_USERS.REQUESTER.length].id,
        });
        break;
      case 'PAYMENT_VOUCHER':
        doc = createMockPaymentVoucher({
          status,
          currentStage: status === 'DRAFT' ? 0 : Math.min(i % 4, 3),
          createdAt: createdDate,
          updatedAt: createdDate,
          createdBy: MOCK_USERS.REQUESTER[i % MOCK_USERS.REQUESTER.length].id,
        });
        break;
      default:
        doc = createMockRequisitionForm({
          status,
          currentStage: status === 'DRAFT' ? 0 : Math.min(i % 4, 3),
          createdAt: createdDate,
          updatedAt: createdDate,
          createdBy: MOCK_USERS.REQUESTER[i % MOCK_USERS.REQUESTER.length].id,
        });
    }

    // Store in Map
    documentStore.set(doc.id, doc);
  }

  isInitialized = true;
}

// THIS RUNS AUTOMATICALLY ON MODULE IMPORT
initializeSampleData();
```

**What This Does:**
1. Creates 25 sample documents
2. Varies the status (DRAFT, SUBMITTED, IN_REVIEW, APPROVED, REJECTED, REVERSED)
3. Varies the type (REQUISITION, PURCHASE_ORDER, PAYMENT_VOUCHER)
4. Sets dates going back in time (25 days of history)
5. Assigns random creators from MOCK_USERS
6. Stores each document in the Map

**When Does It Run:**
- On first import of `workflow-initialization.ts` module
- Which happens when the app starts
- Runs ONCE (isInitialized flag prevents re-initialization)
- Populates with the same 25 documents every startup

---

### Layer 4: Server Actions That Query the Maps

**File:** `src/app/_actions/workflow.ts` and `src/app/_actions/search.ts`

These "server actions" query the in-memory Maps:

```typescript
// Example: getDocumentsByCreator()
export async function getDocumentsByCreator(
  userId: string,
  page: number = 1,
  limit: number = 10
): Promise<APIResponse<PaginatedResponse<WorkflowDocument>>> {
  const session = await auth();
  if (!session?.user) return unauthorizedResponse();

  try {
    // Query the in-memory Map
    const documents = Array.from(documentStore.values()).filter(
      (doc) => doc.createdBy === userId
    );

    // Paginate
    const total = documents.length;
    const totalPages = Math.ceil(total / limit);
    const start = (page - 1) * limit;
    const paginatedDocs = documents.slice(start, start + limit);

    return {
      success: true,
      message: "Documents retrieved successfully",
      data: {
        data: paginatedDocs,
        pagination: { page, limit, total, totalPages },
      },
      status: 200,
    };
  } catch (error) {
    console.error("Error fetching documents:", error);
    return handleError(error, "GET", `/workflows/documents`);
  }
}
```

**What's Happening:**
1. `Array.from(documentStore.values())` - Get all documents from the Map
2. `.filter(doc => doc.createdBy === userId)` - Filter in memory
3. Return results

---

## Complete Data Journey

```
┌─────────────────────────────────────────────────────────────────┐
│  Step 1: APP STARTS                                             │
│                                                                 │
│  index.ts imports workflow-initialization.ts                   │
│         ↓                                                       │
│  initializeSampleData() executes                               │
│         ↓                                                       │
│  Creates 25 mock documents using factory functions             │
│         ↓                                                       │
│  Stores each in documentStore Map<id, doc>                     │
│                                                                 │
│  Result: documentStore now contains 25 documents               │
└─────────────────────────────────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│  Step 2: USER NAVIGATES TO /search                             │
│                                                                 │
│  Page loads, SearchClient component renders                    │
│  User sees empty search form                                   │
└─────────────────────────────────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│  Step 3: USER CLICKS SEARCH                                    │
│                                                                 │
│  searchDocuments() server action called                        │
│         ↓                                                       │
│  Query 1: getDocumentsByCreator(userId, 1, 1000)              │
│  → Array.from(documentStore.values()).filter(...)             │
│  → Returns ~8 docs created by this user (from 25 total)       │
│         ↓                                                       │
│  Query 2: getPendingApprovals(userRole)                        │
│  → Array.from(documentStore.values()).filter(...)             │
│  → Returns ~5 docs awaiting approval from this user            │
│         ↓                                                       │
│  Combine: 8 + 5 = 13 docs (with dedup)                        │
│         ↓                                                       │
│  Apply filters → ~12 docs match                                │
│         ↓                                                       │
│  Sort by date → newest first                                   │
│         ↓                                                       │
│  Paginate → first 10 returned                                  │
│         ↓                                                       │
│  Return to client                                              │
└─────────────────────────────────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│  Step 4: TABLE RENDERS WITH RESULTS                            │
│                                                                 │
│  User sees 10 documents in table                               │
│  All data came from in-memory documentStore Map                │
│  No database, no API, no localStorage                          │
└─────────────────────────────────────────────────────────────────┘
```

---

## Visual Data Source

```
CURRENT STATE (MVP):

User's Browser
  │
  ├─ SearchPage (Server Component)
  │   └─ SearchClient (Client Component)
  │       ├─ SearchForm (UI)
  │       └─ TransactionResults (Table)
  │
Next.js Server (Same Process)
  │
  ├─ workflow.ts (Server Actions)
  │   └─ Query: documentStore.values().filter(...)
  │
  └─ MEMORY (JavaScript Heap)
      └─ documentStore Map<string, WorkflowDocument>
          ├─ 'req-1': { id, documentNumber: 'REQ-...', ... }
          ├─ 'po-2': { id, documentNumber: 'PO-...', ... }
          ├─ 'pv-3': { id, documentNumber: 'PV-...', ... }
          └─ ... (25 total documents)

NO DATABASE ✗
NO LOCAL STORAGE ✗
NO EXTERNAL API ✗
JUST IN-MEMORY JAVASCRIPT ✓
```

---

## Sample Data (What Gets Created)

### Created on App Start

The `initializeSampleData()` function creates 25 documents with:

**Document Statuses (cycle through):**
- DRAFT
- SUBMITTED
- IN_REVIEW
- APPROVED
- REJECTED
- REVERSED

**Document Types (cycle through):**
- REQUISITION
- PURCHASE_ORDER
- PAYMENT_VOUCHER

**Creators (random from MOCK_USERS):**
```typescript
MOCK_USERS = {
  REQUESTER: [
    { id: 'user-req-1', name: 'John Mwale', ... },
    { id: 'user-req-2', name: 'Sarah Banda', ... },
  ],
  DEPARTMENT_MANAGER: [
    { id: 'user-dm-1', name: 'James Chileshe', ... },
    { id: 'user-dm-2', name: 'Maria Chiyanda', ... },
  ],
  FINANCE_OFFICER: [
    { id: 'user-fo-1', name: 'Paul Nkosi', ... },
    { id: 'user-fo-2', name: 'Grace Mvula', ... },
  ],
  DIRECTOR: [
    { id: 'user-dir-1', name: 'David Moyo', ... },
  ],
  CFO: [
    { id: 'user-cfo-1', name: 'Catherine Phiri', ... },
  ],
  COMPLIANCE_OFFICER: [
    { id: 'user-co-1', name: 'Victor Zulu', ... },
  ],
  ADMIN: [
    { id: 'user-admin-1', name: 'Admin User', ... },
  ],
}
```

**Example Generated Document:**

```javascript
{
  id: "550e8400-e29b-41d4-a716-446655440000",
  documentNumber: "REQ-2024-0001",
  type: "REQUISITION",
  status: "DRAFT",
  createdBy: "user-req-1",
  createdAt: Date(2024-12-04),
  updatedAt: Date(2024-12-04),
  title: "Office Supplies",
  description: "Request for office supplies",
  items: [
    {
      id: "item-1",
      description: "Printer Paper",
      quantity: 10,
      unitPrice: 50,
      totalPrice: 500,
      // ... more fields
    }
  ],
  currentStage: 0,
  status: "DRAFT",
  approvalChain: [...],
  // ... many more fields
}
```

---

## Key Characteristics

### ✅ What Works

| Feature | Status | How |
|---------|--------|-----|
| Search by filters | ✅ Works | Queries in-memory Map |
| View document | ✅ Works | Retrieves from Map by ID |
| Pagination | ✅ Works | Client-side array slicing |
| Download | ✅ Works | Generates mock URL |
| Create new document | ✅ Works | Adds to Map |
| Update document | ✅ Works | Updates Map value |
| Delete document | ✅ Works | Removes from Map |
| Approvals | ✅ Works | Stored in approversStore Map |

### ❌ What Doesn't Persist

| Feature | Status | Reason |
|---------|--------|--------|
| Data survives refresh | ❌ Lost | In-memory only |
| Data survives restart | ❌ Lost | Not persisted anywhere |
| Multi-server sync | ❌ No | Each server has own Map |
| Concurrent users | ⚠️ Conflicts | No locking mechanism |
| Backup/Export | ❌ No | No persistence layer |

---

## Comparison: Mock Data vs. Database

```
                    MOCK DATA (Current)     DATABASE (Production)
┌─────────────────────────────────────────────────────────────────┐
│ Location          │ JavaScript Memory     │ PostgreSQL/MongoDB   │
│ Persistence       │ Runtime only          │ Disk + Snapshots     │
│ Data loss on      │ App restart           │ Never (persistent)   │
│ restart?          │                       │                      │
├─────────────────────────────────────────────────────────────────┤
│ Multi-instance    │ Each has own copy     │ Shared database      │
│ support?          │ (no sync)             │ (all sync)           │
├─────────────────────────────────────────────────────────────────┤
│ Concurrent        │ Race conditions       │ ACID transactions    │
│ writes?           │ possible              │                      │
├─────────────────────────────────────────────────────────────────┤
│ Backups?          │ None                  │ Automated dumps      │
├─────────────────────────────────────────────────────────────────┤
│ Query speed       │ Fast (in-memory)      │ Depends on indexes   │
├─────────────────────────────────────────────────────────────────┤
│ Scale to millions │ No (heap limited)     │ Yes (with scaling)   │
│ of records?       │                       │                      │
├─────────────────────────────────────────────────────────────────┤
│ Complex queries   │ Code in JS            │ SQL/aggregations     │
│ (aggregations)?   │                       │                      │
├─────────────────────────────────────────────────────────────────┤
│ Perfect for       │ MVP development       │ Production use       │
│                   │ Testing               │ Multi-server         │
│                   │ Demos                 │ Long-term storage    │
└─────────────────────────────────────────────────────────────────┘
```

---

## How Data Flows Through a Search

```
User Input (Form)
    │
    ▼
searchDocuments() Server Action
    │
    ├─ Query 1: Array.from(documentStore.values())
    │   │           ↓
    │   │   Filter by createdBy === userId
    │   │           ↓
    │   └─ Results: ~8 documents
    │
    ├─ Query 2: Array.from(documentStore.values())
    │   │           ↓
    │   │   Filter by status === 'IN_REVIEW'
    │   │   Filter by approvers[].role === userRole
    │   │           ↓
    │   └─ Results: ~5 documents
    │
    ├─ Combine: Merge arrays → 13 documents
    │
    ├─ Deduplicate: Use Map to remove duplicates → 12 documents
    │
    ├─ Filter: Apply user's search criteria
    │   - documentNumber: contains "REQ"
    │   - documentType: === 'REQUISITION'
    │   - status: === 'APPROVED'
    │   - dates: within range
    │           ↓
    │   Results: 7 documents match all criteria
    │
    ├─ Sort: by createdAt descending → newest first
    │
    ├─ Paginate: slice(0, 10) → first 10 results
    │           ↓
    │   Returns: 7 documents (all fit on page 1)
    │
    ▼
Return to Client (UI)
    │
    ▼
TransactionResults Component
    │
    ├─ setDocuments([...7 results...])
    ├─ setPagination({ page: 1, total: 7, totalPages: 1 })
    │
    ▼
Table Renders
    │
    ├─ Header: Document # | Type | Status | Created | Actions
    └─ Row 1: REQ-2024-XXX | Requisition | Approved | 12/4/2024 | [View] [DL]
      Row 2: REQ-2024-YYY | Requisition | Approved | 12/3/2024 | [View] [DL]
      ...
```

---

## Where Is Data Actually Stored (Files)

```
PROJECT ROOT
├── src/
│   ├── lib/
│   │   ├── workflow-stores.ts          ← Declares the Maps
│   │   ├── workflow-initialization.ts  ← Initializes sample data
│   │   └── mock-data.ts                ← Factory functions
│   │
│   ├── app/
│   │   └── _actions/
│   │       ├── workflow.ts             ← Queries the Maps
│   │       └── search.ts               ← Search queries
│   │
│   └── (no database files or configs)
│
└── (no .env with DB credentials)
```

---

## Summary

| Aspect | Answer |
|--------|--------|
| **Where does search data come from?** | In-memory JavaScript Map called `documentStore` |
| **Is it from localStorage?** | No |
| **Is it from a database?** | No |
| **Is it from a file?** | No |
| **Is it mocked?** | Yes, completely mocked |
| **How much data?** | 25 hardcoded sample documents |
| **Does it persist?** | No, lost on app restart |
| **Can multiple servers share it?** | No, each has its own copy |
| **Is this production-ready?** | No, this is MVP/demo architecture |
| **What would replace this?** | Real database (PostgreSQL, MongoDB, etc.) |

---

## Next Steps for Production

To move from mock data to real data:

1. **Add Database**
   - PostgreSQL, MongoDB, or other DBMS
   - Create schema/collections for documents

2. **Replace In-Memory Maps with Database Queries**
   - Change: `documentStore.get(id)` → `database.documents.findById(id)`
   - Change: `Array.from(documentStore.values()).filter(...)` → `database.documents.find({...})`

3. **Implement Persistence**
   - Data survives restarts
   - Multiple servers can access same data

4. **Remove Initialization Code**
   - Delete `workflow-initialization.ts`
   - Delete `mock-data.ts` factory functions
   - Stop auto-creating fake documents

5. **Add Data Validation**
   - Schema validation at DB layer
   - Query optimization with indexes

6. **Implement Transactions**
   - ACID compliance
   - Concurrent write handling

This would be a complete architectural change, but the server actions (`workflow.ts`) and search logic would remain mostly the same - they'd just query a database instead of in-memory Maps.
