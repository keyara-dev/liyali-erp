# Search System - Visual Flows & Diagrams

## Complete User Journey

```
┌──────────────────────────────────────────────────────────────────────────┐
│                        USER STARTS HERE                                  │
│                  Click "Search Transactions"                             │
└────────────────────────────┬─────────────────────────────────────────────┘
                             │
                             ▼
┌──────────────────────────────────────────────────────────────────────────┐
│                    Page Loads: /search                                   │
│                                                                          │
│  Server checks: Is user logged in?                                      │
│     ├─ NO → Redirect to /login                                          │
│     └─ YES → Continue with userId and userRole                          │
└────────────────────────────┬─────────────────────────────────────────────┘
                             │
                             ▼
┌──────────────────────────────────────────────────────────────────────────┐
│               SearchClient Component Renders                             │
│  - Initializes filter state (all empty/ALL)                             │
│  - Passes to SearchForm and TransactionResults children                 │
│  - Ready for user input                                                 │
└────────────────────────────┬─────────────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────┬─────────────────────────────────┐
│                                         │                                 │
│    USER FILLS SEARCH FORM              │  RESULTS TABLE SHOWS NOTHING    │
│    (State updates in real-time)        │  (Waiting for search click)     │
│                                         │                                 │
│  ┌─────────────────────────────────┐  │  ┌─────────────────────────────┐│
│  │ Document Number: "REQ"          │  │  │ [No results yet]            ││
│  │ Document Type: Requisitions ✓   │  │  │                             ││
│  │ Status: Approved ✓              │  │  │ Results appear after search ││
│  │ Start Date: 2024-01-01 ✓        │  │  └─────────────────────────────┘│
│  │ End Date: 2024-12-31 ✓          │  │                                 │
│  │                                 │  │                                 │
│  │ [Reset]  [Search]  ◀─────────────────────────────────────            │
│  └─────────────────────────────────┘  │                                 │
│                                         │                                 │
└─────────────────────────────────────────┴─────────────────────────────────┘
                             │
                             │ (User clicks Search)
                             ▼
┌──────────────────────────────────────────────────────────────────────────┐
│                   handleSearch() Executes                                │
│                                                                          │
│  1. Prevent default form submission                                     │
│  2. Build SearchFilters object from form state:                        │
│     {                                                                    │
│       documentNumber: 'REQ',                                            │
│       documentType: 'REQUISITION',                                      │
│       status: 'APPROVED',                                               │
│       startDate: '2024-01-01',                                          │
│       endDate: '2024-12-31'                                             │
│     }                                                                    │
│  3. Call onSearch(newFilters) callback                                  │
└────────────────────────────┬─────────────────────────────────────────────┘
                             │
                             ▼
┌──────────────────────────────────────────────────────────────────────────┐
│             SearchClient.handleSearch() Executes                         │
│                                                                          │
│  1. setFilters(newFilters)         ← Updates parent state               │
│  2. setIsSearching(true)           ← Disable buttons, show loading      │
│  3. setRefreshTrigger(++counter)   ← Trigger TransactionResults effect │
└────────────────────────────┬─────────────────────────────────────────────┘
                             │
                             ▼
┌──────────────────────────────────────────────────────────────────────────┐
│         TransactionResults useEffect Hook Fires                          │
│         (Triggered by: filters change OR refreshTrigger change)         │
│                                                                          │
│  Dependency array: [filters, pagination.page, pagination.limit,        │
│                     refreshTrigger]                                      │
│  → Condition met: refreshTrigger changed                                │
│  → Effect executes: Call fetchDocuments()                               │
└────────────────────────────┬─────────────────────────────────────────────┘
                             │
                             ▼
┌──────────────────────────────────────────────────────────────────────────┐
│            searchDocuments() Server Action Called                        │
│                                                                          │
│  Input: {                                                                │
│    filters: { documentNumber, documentType, status, dates },            │
│    page: 1,                                                              │
│    limit: 10                                                             │
│  }                                                                       │
└────────────────────────────┬─────────────────────────────────────────────┘
                             │
                             ▼
┌──────────────────────────────────────────────────────────────────────────┐
│                    SERVER PROCESSING BEGINS                              │
│                                                                          │
│  1. Verify session → if no session, return 401 error                    │
│  2. Fetch created documents:                                            │
│     createdResult = await getDocumentsByCreator(userId, 1, 1000)       │
│  3. Fetch pending approvals:                                            │
│     pendingResult = await getPendingApprovals(userRole)                │
│  4. Combine both arrays:                                                │
│     allDocuments = [...createdDocs, ...pendingDocs]                    │
└────────────────────────────┬─────────────────────────────────────────────┘
                             │
                             ▼
┌──────────────────────────────────────────────────────────────────────────┐
│              DEDUPLICATION                                               │
│                                                                          │
│  Problem: Same document might appear in both lists                      │
│  Solution: Use Map to track by ID                                       │
│                                                                          │
│  const uniqueMap = new Map<string, WorkflowDocument>();                │
│  allDocuments.forEach(doc => uniqueMap.set(doc.id, doc));               │
│  const uniqueDocuments = Array.from(uniqueMap.values());                │
│                                                                          │
│  Before: 27 total (with duplicates)                                     │
│  After:  23 unique documents                                            │
└────────────────────────────┬─────────────────────────────────────────────┘
                             │
                             ▼
┌──────────────────────────────────────────────────────────────────────────┐
│            FILTER APPLICATION (One at a time)                            │
│                                                                          │
│  Start: 23 unique documents                                             │
│                                                                          │
│  Filter 1: documentNumber.includes('REQ')                               │
│  ├─ Keep: REQ-2024-001 ✓                                                │
│  ├─ Keep: REQ-2024-005 ✓                                                │
│  ├─ Remove: PO-2024-001 ✗                                               │
│  ├─ Remove: PV-2024-001 ✗                                               │
│  └─ Result: 15 documents                                                │
│                                                                          │
│  Filter 2: documentType === 'REQUISITION'                               │
│  └─ Result: 15 documents (all are requisitions now)                    │
│                                                                          │
│  Filter 3: status === 'APPROVED'                                        │
│  ├─ Remove: REQ-2024-001 (status=DRAFT) ✗                              │
│  ├─ Keep: REQ-2024-005 (status=APPROVED) ✓                             │
│  └─ Result: 12 documents                                                │
│                                                                          │
│  Filter 4: createdAt >= startDate (2024-01-01 00:00:00)                │
│  └─ Result: 12 documents (all from 2024)                               │
│                                                                          │
│  Filter 5: createdAt <= endDate (2024-12-31 23:59:59)                  │
│  └─ Result: 12 documents (all within year 2024)                        │
│                                                                          │
│  FINAL: 12 documents match all criteria                                 │
└────────────────────────────┬─────────────────────────────────────────────┘
                             │
                             ▼
┌──────────────────────────────────────────────────────────────────────────┐
│                      SORTING                                             │
│                                                                          │
│  filtered.sort((a, b) =>                                                │
│    new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()   │
│  );                                                                      │
│                                                                          │
│  Sort by: createdAt descending (newest first)                           │
│                                                                          │
│  Before sort:                  After sort:                              │
│  1. REQ-2024-001 (Jan 1)  →    1. REQ-2024-005 (Dec 15)                │
│  2. REQ-2024-005 (Dec 15) →    2. REQ-2024-003 (Nov 10)                │
│  3. REQ-2024-003 (Nov 10) →    3. REQ-2024-001 (Jan 1)                │
│  ...                      →    ...                                     │
└────────────────────────────┬─────────────────────────────────────────────┘
                             │
                             ▼
┌──────────────────────────────────────────────────────────────────────────┐
│                      PAGINATION                                          │
│                                                                          │
│  Total documents: 12                                                     │
│  Page requested: 1                                                       │
│  Limit: 10                                                               │
│                                                                          │
│  Calculations:                                                           │
│  totalPages = Math.ceil(12 / 10) = 2 pages                              │
│  skip = (1 - 1) * 10 = 0                                                │
│  paginatedData = documents.slice(0, 10)                                 │
│                                                                          │
│  Result: Items 1-10 returned for page 1                                │
└────────────────────────────┬─────────────────────────────────────────────┘
                             │
                             ▼
┌──────────────────────────────────────────────────────────────────────────┐
│                  RESPONSE RETURNED TO CLIENT                             │
│                                                                          │
│  {                                                                       │
│    success: true,                                                        │
│    message: "Documents search completed",                               │
│    data: {                                                               │
│      data: [                                                             │
│        { id: 'req-123', documentNumber: 'REQ-2024-005', ... },          │
│        { id: 'req-124', documentNumber: 'REQ-2024-003', ... },          │
│        ...                                                               │
│        { id: 'req-132', documentNumber: 'REQ-2024-010', ... }           │
│      ],                                                                  │
│      pagination: {                                                       │
│        page: 1,                                                          │
│        limit: 10,                                                        │
│        total: 12,                                                        │
│        totalPages: 2                                                     │
│      }                                                                   │
│    },                                                                    │
│    status: 200                                                           │
│  }                                                                       │
└────────────────────────────┬─────────────────────────────────────────────┘
                             │
                             ▼
┌──────────────────────────────────────────────────────────────────────────┐
│         TransactionResults Updates State & Renders Table                 │
│                                                                          │
│  const result = await searchDocuments(...);                             │
│  if (result.success) {                                                  │
│    setDocuments(result.data.data);           // 10 items                │
│    setPagination(result.data.pagination);    // page 1 of 2             │
│    setIsLoading(false);                      // Hide spinner            │
│  }                                                                       │
└────────────────────────────┬─────────────────────────────────────────────┘
                             │
                             ▼
┌──────────────────────────────────────────────────────────────────────────┐
│          TABLE RENDERS WITH RESULTS                                      │
│  ┌──────────────────────────────────────────────────────────────────┐   │
│  │ Document # │ Type        │ Status   │ Created      │ Actions     │   │
│  ├──────────────────────────────────────────────────────────────────┤   │
│  │ REQ-2024-0005 │ Requisition │ Approved │ 12/15/2024 │ [View] [DL] │   │
│  │ REQ-2024-0003 │ Requisition │ Approved │ 11/10/2024 │ [View] [DL] │   │
│  │ REQ-2024-0001 │ Requisition │ Approved │ 01/01/2024 │ [View] [DL] │   │
│  │ ...           │             │          │            │             │   │
│  └──────────────────────────────────────────────────────────────────┘   │
│                                                                          │
│  Pagination:                                                             │
│  Showing 1 to 10 of 12 documents                                        │
│  [< Previous] Page 1 of 2 [Next >]                                      │
└──────────────────────────────────────────────────────────────────────────┘
                             │
                             ▼
                    ┌────────────────────┐
                    │ USER SEES RESULTS  │
                    └────────────────────┘
                             │
                ┌────────────┼────────────┐
                │            │            │
                ▼            ▼            ▼
        ┌──────────────┐ ┌──────────┐ ┌──────────────┐
        │ Click View   │ │ Click DL │ │ Click Next   │
        │ → Navigate   │ │ → Download │ │ → Page 2    │
        │   to detail  │ │   PDF    │ │   → More     │
        │   page       │ │          │ │   results    │
        └──────────────┘ └──────────┘ └──────────────┘
                │            │            │
                ▼            ▼            ▼
           (Routes to  (Triggers    (Updates page
            detail     download)    & refetch)
            page)
```

---

## Form Interaction Diagram

```
┌────────────────────────────────────────────────────────────────┐
│              SEARCH FORM COMPONENT                             │
│  ┌─────────────────────────────────────────────────────────┐  │
│  │ 🎨 HEADER: "Search Filters"                            │  │
│  ├─────────────────────────────────────────────────────────┤  │
│  │                                                         │  │
│  │ ROW 1 (Document Number & Type)                         │  │
│  │ ┌──────────────────────┐ ┌──────────────────────────┐ │  │
│  │ │ Document Number      │ │ Document Type          │ │  │
│  │ │ [_____ REQ_____]     │ │ [Requisitions ▼]       │ │  │
│  │ │ Real-time input      │ │ Updated on select      │ │  │
│  │ └──────────────────────┘ └──────────────────────────┘ │  │
│  │                                                         │  │
│  │ ROW 2 (Status & Dates)                                 │  │
│  │ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐   │  │
│  │ │ Status       │ │ Start Date   │ │ End Date     │   │  │
│  │ │ [Approved ▼] │ │ [2024-01-01] │ │ [2024-12-31] │   │  │
│  │ │ Dropdown     │ │ Date Picker  │ │ Date Picker  │   │  │
│  │ └──────────────┘ └──────────────┘ └──────────────┘   │  │
│  │                                                         │  │
│  │ ROW 3 (Action Buttons)                                 │  │
│  │ ┌─────────────────────────────────────────────────┐   │  │
│  │ │        [Reset]           [🔍 Search]           │   │  │
│  │ │     Clears all fields   Shows spinner during    │   │  │
│  │ │                         fetch                   │   │  │
│  │ └─────────────────────────────────────────────────┘   │  │
│  │                                                         │  │
│  └─────────────────────────────────────────────────────────┘  │
│                                                                │
│  STATE CHANGES:                                                │
│  • documentNumber state updates on each keystroke             │
│  • documentType state updates on dropdown select              │
│  • status state updates on dropdown select                    │
│  • startDate/endDate update on date picker select             │
│                                                                │
│  BUTTON STATES:                                                │
│  • Search button:                                              │
│    - Normal state: Can click                                   │
│    - isSearching=true: Disabled, shows "Searching..."         │
│    - isSearching=false: Normal again                           │
│  • Reset button:                                               │
│    - Normal state: Can click                                   │
│    - isSearching=true: Disabled                                │
│    - Normal again when done                                    │
└────────────────────────────────────────────────────────────────┘
```

---

## View Action Flow

```
User clicks "View" button on a row with:
{ id: 'po-123', type: 'PURCHASE_ORDER', documentNumber: 'PO-2024-0042' }
                                    │
                                    ▼
                    ┌───────────────────────────┐
                    │ React Router Location:    │
                    │ /search                   │
                    └───────────────────────────┘
                                    │
                                    ▼
                    ┌───────────────────────────┐
                    │ Check document type      │
                    │ Find URL slug:            │
                    │ PURCHASE_ORDER            │
                    │   ↓                       │
                    │ 'purchase-orders'        │
                    └───────────────────────────┘
                                    │
                                    ▼
                    ┌───────────────────────────┐
                    │ Build URL:                │
                    │ /purchase-orders/po-123   │
                    └───────────────────────────┘
                                    │
                                    ▼
                    ┌───────────────────────────┐
                    │ router.push(url)          │
                    │ Next.js navigates         │
                    └───────────────────────────┘
                                    │
                                    ▼
                    ┌───────────────────────────┐
                    │ [id]/page.tsx loads       │
                    │ Server fetches PO data    │
                    │ Detail page renders       │
                    └───────────────────────────┘

TYPE → URL MAPPING:
┌──────────────────────┬─────────────────────────────┐
│ Document Type        │ URL Slug                    │
├──────────────────────┼─────────────────────────────┤
│ REQUISITION          │ /requisitions/{id}          │
│ PURCHASE_ORDER       │ /purchase-orders/{id}       │
│ PAYMENT_VOUCHER      │ /payment-vouchers/{id}      │
│ GOODS_RECEIVED_NOTE  │ /grn/{id}                   │
└──────────────────────┴─────────────────────────────┘
```

---

## Download Action Flow

```
User clicks "Download" button on row with:
{ id: 'req-456', documentNumber: 'REQ-2024-0001' }
                                    │
                                    ▼
                    ┌───────────────────────────┐
                    │ setIsLoading(true)        │
                    │ Button shows spinner      │
                    │ Button becomes disabled   │
                    └───────────────────────────┘
                                    │
                                    ▼
                    ┌───────────────────────────┐
                    │ Call Server Action:       │
                    │ downloadDocumentPDF()     │
                    │ documentId='req-456'      │
                    └───────────────────────────┘
                                    │
                                    ▼
        ┌───────────────────────────────────────────┐
        │        SERVER PROCESSING                  │
        │                                           │
        │ 1. Verify session                         │
        │    ├─ Not authenticated                   │
        │    │   → return { success: false }        │
        │    │   → Button returns to normal         │
        │    │   → Alert: "Failed to download"      │
        │    │                                      │
        │    └─ Authenticated ✓                    │
        │                                           │
        │ 2. Fetch document metadata                │
        │    await getDocument('req-456')           │
        │    ├─ Not found                           │
        │    │   → return { success: false }        │
        │    │   → Alert: "Failed to download"      │
        │    │                                      │
        │    └─ Found ✓                             │
        │                                           │
        │ 3. Generate download URL                  │
        │    `/api/documents/req-456/download`      │
        │                                           │
        │ 4. Return success response:               │
        │    {                                      │
        │      success: true,                       │
        │      data: {                              │
        │        downloadUrl:                       │
        │        "/api/documents/req-456/download"  │
        │      }                                    │
        │    }                                      │
        └───────────────────────────────────────────┘
                                    │
                                    ▼
                    ┌───────────────────────────┐
                    │ Client receives response  │
                    │ SUCCESS?                  │
                    │   ├─ YES:                 │
                    │   │   ├─ Create <a> tag  │
                    │   │   ├─ Set href=URL    │
                    │   │   ├─ Set download=   │
                    │   │   │  "REQ-2024-0001" │
                    │   │   ├─ Append to DOM   │
                    │   │   ├─ Click           │
                    │   │   ├─ Remove from DOM │
                    │   │   │                  │
                    │   │   └─ BROWSER: Start  │
                    │   │      PDF download    │
                    │   │                      │
                    │   └─ NO:                 │
                    │       ├─ Alert error     │
                    │       └─ Button normal   │
                    └───────────────────────────┘
                                    │
                                    ▼
                    ┌───────────────────────────┐
                    │ setIsLoading(false)       │
                    │ Button returns to normal  │
                    │ Spinner disappears        │
                    │ Button enabled again      │
                    └───────────────────────────┘
                                    │
                                    ▼
                    ┌───────────────────────────┐
                    │ USER DOWNLOADS PDF FILE   │
                    │ Filename: REQ-2024-0001   │
                    │ Location: Downloads/      │
                    └───────────────────────────┘
```

---

## Pagination Flow

```
Initial State: Total 23 documents, limit 10 per page
                                    │
                                    ▼
        ┌───────────────────────────────────────────┐
        │  Calculate Pages:                         │
        │  totalPages = Math.ceil(23 / 10)         │
        │           = Math.ceil(2.3)                │
        │           = 3 pages                       │
        │                                           │
        │  Page distribution:                       │
        │  Page 1: items 1-10                       │
        │  Page 2: items 11-20                      │
        │  Page 3: items 21-23                      │
        └───────────────────────────────────────────┘
                                    │
                                    ▼
        ┌───────────────────────────────────────────┐
        │  INITIAL: Page 1                          │
        │  Showing 1 to 10 of 23 documents          │
        │  [< Prev] Page 1 of 3 [Next >]            │
        │   DISABLED              ENABLED           │
        │  (can't go before 1)  (can go to page 2) │
        └───────────────────────────────────────────┘
                                    │
                    ┌───────────────┴───────────────┐
                    │                               │
                    ▼                               ▼
        ┌──────────────────────┐      ┌──────────────────────┐
        │ User clicks [Next >]  │      │ User clicks [< Prev] │
        │ (on page 1)           │      │ (on page 2)          │
        └──────────────────────┘      └──────────────────────┘
                    │                               │
                    ▼                               ▼
        ┌──────────────────────┐      ┌──────────────────────┐
        │ setPagination({       │      │ setPagination({      │
        │   page: Math.min(     │      │   page: Math.max(    │
        │     2,                │      │     1,               │
        │     3 /* maxPage */   │      │     1 /* minPage */  │
        │   )                   │      │   )                  │
        │ })                    │      │ })                   │
        │                       │      │                      │
        │ → page becomes 2      │      │ → page becomes 1     │
        └──────────────────────┘      └──────────────────────┘
                    │                               │
                    ▼                               ▼
        ┌──────────────────────┐      ┌──────────────────────┐
        │ useEffect triggers   │      │ useEffect triggers   │
        │ (page changed)       │      │ (page changed)       │
        │                      │      │                      │
        │ searchDocuments({    │      │ searchDocuments({    │
        │   filters,           │      │   filters,           │
        │   page: 2,           │      │   page: 1,           │
        │   limit: 10          │      │   limit: 10          │
        │ })                   │      │ })                   │
        └──────────────────────┘      └──────────────────────┘
                    │                               │
                    ▼                               ▼
        ┌──────────────────────┐      ┌──────────────────────┐
        │ Server:              │      │ Server:              │
        │ skip = (2-1)*10 = 10 │      │ skip = (1-1)*10 = 0  │
        │ slice(10, 20)        │      │ slice(0, 10)         │
        │ → Items 11-20        │      │ → Items 1-10         │
        └──────────────────────┘      └──────────────────────┘
                    │                               │
                    ▼                               ▼
        ┌──────────────────────┐      ┌──────────────────────┐
        │ Table updates        │      │ Table updates        │
        │ Shows items 11-20    │      │ Shows items 1-10     │
        │                      │      │                      │
        │ Showing 11 to 20 of  │      │ Showing 1 to 10 of   │
        │ 23 documents         │      │ 23 documents         │
        │                      │      │                      │
        │ [< Prev] Page 2 of 3 │      │ [< Prev] Page 1 of 3 │
        │  ENABLED  ENABLED    │      │  DISABLED  ENABLED   │
        └──────────────────────┘      └──────────────────────┘
                    │                               │
                    ▼                               ▼
        ┌──────────────────────┐      ┌──────────────────────┐
        │ User clicks [Next >] │      │ User clicks [Next >] │
        │ on page 2            │      │ on page 1            │
        └──────────────────────┘      └──────────────────────┘
                    │                               │
                    ▼                               ▼
        ┌──────────────────────┐      ┌──────────────────────┐
        │ page = 3             │      │ page = 2             │
        │ → Page 3 of 3        │      │ → Page 2 of 3        │
        │ → Items 21-23        │      │ → Items 11-20        │
        │ → Next btn DISABLED  │      │ → Both ENABLED       │
        │   (at end)           │      │                      │
        └──────────────────────┘      └──────────────────────┘

SUMMARY:
┌─────────────────────────────────────────────────────────────┐
│ Previous Button:                                            │
│ • DISABLED on page 1 (can't go before first page)          │
│ • ENABLED on page 2+ (can go back)                         │
│ • DISABLED during loading (isLoading=true)                 │
│                                                             │
│ Next Button:                                                │
│ • ENABLED on page < totalPages (can go forward)            │
│ • DISABLED on final page (can't go past last page)         │
│ • DISABLED during loading (isLoading=true)                 │
│                                                             │
│ Page Display:                                               │
│ • Always shows: "Page X of Y"                              │
│ • Updates when page state changes                          │
└─────────────────────────────────────────────────────────────┘
```

---

## Data Combination Diagram

```
TWO DATA SOURCES:

┌────────────────────────────────────┐
│  Query 1: getDocumentsByCreator()  │
│                                    │
│  documentStore.filter(doc =>       │
│    doc.createdBy === userId        │
│  )                                 │
│                                    │
│  Returns:                          │
│  ├─ REQ-2024-001 (created by me)  │
│  ├─ PO-2024-001 (created by me)   │
│  ├─ PV-2024-001 (created by me)   │
│  ├─ GRN-2024-001 (created by me)  │
│  ├─ REQ-2024-005 (created by me)  │
│  └─ ... (18 more docs)             │
│                                    │
│  Total: 22 documents               │
└────────────────────────────────────┘
           │
           │ Combine both
           │
           ▼
┌────────────────────────────────────┐
│  Query 2: getPendingApprovals()    │
│                                    │
│  documentStore.filter(doc =>       │
│    doc.status === 'IN_REVIEW' &&   │
│    approverHasMyRole(doc)          │
│  )                                 │
│                                    │
│  Returns:                          │
│  ├─ REQ-2024-005 (waiting MY appro)│ ← DUPLICATE!
│  ├─ PO-2024-005 (waiting MY appro) │
│  ├─ PV-2024-002 (waiting MY appro) │
│  └─ ... (7 more docs)              │
│                                    │
│  Total: 10 documents               │
└────────────────────────────────────┘
           │
           │ All docs: 22 + 10 = 32
           │ (with duplicates)
           │
           ▼
┌────────────────────────────────────┐
│  DEDUPLICATION STEP:               │
│                                    │
│  const uniqueMap = new Map();      │
│  allDocuments.forEach(doc =>       │
│    uniqueMap.set(doc.id, doc)      │
│  );                                │
│                                    │
│  Before:                           │
│  uniqueMap.set('req-005', {...})   │ ← Set from source 1
│  uniqueMap.set('req-005', {...})   │ ← OVERWRITE with src 2
│  (same key, latest value kept)     │
│                                    │
│  Result: 23 unique documents       │
│  (32 - 9 duplicates = 23)          │
└────────────────────────────────────┘
           │
           ▼
┌────────────────────────────────────┐
│  UNION SET:                        │
│  ├─ REQ-2024-001 (from source 1)  │
│  ├─ REQ-2024-005 (from both)      │
│  ├─ PO-2024-001 (from source 1)   │
│  ├─ PO-2024-005 (from source 2)   │
│  ├─ PV-2024-001 (from source 1)   │
│  ├─ PV-2024-002 (from source 2)   │
│  ├─ GRN-2024-001 (from source 1)  │
│  └─ ... 16 more                    │
│                                    │
│  Total: 23 unique documents        │
│  (ready for filtering)             │
└────────────────────────────────────┘
```

---

## State Management Diagram

```
┌──────────────────────────────────────────────────────────────────┐
│                   COMPONENT STATE TREE                           │
└──────────────────────────────────────────────────────────────────┘
                            │
        ┌───────────────────┴───────────────────┐
        │                                       │
        ▼                                       ▼
    ┌─────────────────┐              ┌──────────────────┐
    │ SearchClient    │              │ TransactionRes   │
    │                 │              │                  │
    │ • filters       │              │ • documents      │
    │ • refreshTrigger│              │ • pagination     │
    │ • isSearching   │              │ • isLoading      │
    │                 │              │ • sorting        │
    └────┬────────────┘              │ • columnFilters  │
         │                           │ • columnVisibility
         │ passes props             │                  │
         │                          └──────────────────┘
         │
    ┌────┴────────────┬────────────────────┐
    │                 │                    │
    ▼                 ▼                    ▼
┌──────────┐  ┌─────────────────┐  ┌──────────────┐
│SearchForm│  │Transaction      │  │ DownloadBtn  │
│          │  │Results          │  │              │
│• docNum  │  │                 │  │ (nested in   │
│• docType │  │• Uses TanStack  │  │  table row)  │
│• status  │  │  React Table    │  │              │
│• startDt │  │• Manages table  │  │ • isLoading  │
│• endDate │  │  state          │  │              │
└──────────┘  └─────────────────┘  └──────────────┘

FLOW:
┌─────────────────────────────────────────────────────┐
│ User Input → SearchForm state updates (real-time)  │
│ User clicks Search → handleSearch()                 │
│   → SearchClient.handleSearch()                     │
│   → setFilters() + setRefreshTrigger(++counter)    │
│   → TransactionResults useEffect fires             │
│   → searchDocuments() server action                 │
│   → setDocuments() + setPagination()                │
│   → Table re-renders with new data                 │
└─────────────────────────────────────────────────────┘
```

---

## Error States

```
ERROR SCENARIOS:

1. NO SESSION
   ┌─────────────────────────┐
   │ User not authenticated  │
   ├─────────────────────────┤
   │ Server returns: 401     │
   │ Client shows: No results│
   │ UI: Table says "No docs"│
   │ Action: Redirect to /login
   └─────────────────────────┘

2. NETWORK ERROR
   ┌─────────────────────────┐
   │ Server unreachable      │
   ├─────────────────────────┤
   │ Caught in try/catch     │
   │ Console error logged    │
   │ Client shows: No results│
   │ UI: "No documents found"│
   │ Action: User can retry  │
   └─────────────────────────┘

3. INVALID DATE RANGE
   ┌─────────────────────────┐
   │ Start > End date        │
   ├─────────────────────────┤
   │ Filter returns 0 matches│
   │ UI: No documents found  │
   │ Action: User adjusts    │
   └─────────────────────────┘

4. EMPTY RESULTS
   ┌─────────────────────────┐
   │ Filters match nothing   │
   ├─────────────────────────┤
   │ documents[] is empty    │
   │ UI: Shows empty state   │
   │ Icon: SearchX (magnifier)
   │ Text: "No documents     │
   │        found"           │
   │ Action: Try new filters │
   └─────────────────────────┘

5. DOWNLOAD FAILED
   ┌─────────────────────────┐
   │ Document not found      │
   │ OR document deleted     │
   ├─────────────────────────┤
   │ getDocument() returns   │
   │ success: false          │
   │ Client shows alert:     │
   │ "Failed to download"    │
   │ Action: Try again later │
   └─────────────────────────┘
```
