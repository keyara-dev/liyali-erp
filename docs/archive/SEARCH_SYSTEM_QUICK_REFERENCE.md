# Search System - Quick Reference Guide

## 30-Second Summary

The Search page (`/search`) lets users find any document (Requisition, PO, Payment Voucher, GRN) they created or need to approve. Users fill a form with filters (doc number, type, status, date range), click search, and get a paginated table of results with view/download actions for each row.

---

## How to Use (User Perspective)

### Basic Search

```
1. Click "Search Transactions" in sidebar
2. See form with fields:
   - Document Number (type to search)
   - Document Type (dropdown: All/Requisitions/Purchase Orders/Payment Vouchers/GRNs)
   - Status (dropdown: All/Draft/Submitted/In Approval/Approved/Rejected/Reversed)
   - Start Date (date picker)
   - End Date (date picker)
3. Fill in desired filters
4. Click "Search" button
5. Results appear below in table with pagination
```

### View Document

```
1. Find document in results table
2. Click "View" button
3. Browser navigates to document detail page
   - Requisition → /requisitions/{id}
   - PO → /purchase-orders/{id}
   - Payment Voucher → /payment-vouchers/{id}
   - GRN → /grn/{id}
```

### Download Document

```
1. Find document in results table
2. Click "Download" button
3. Spinner shows while generating
4. PDF file downloads to computer as "{documentNumber}.pdf"
```

### Reset Search

```
1. Click "Reset" button
2. All form fields clear
3. Results show everything (all user's documents + pending approvals)
```

### Browse Next Page

```
1. View search results
2. At bottom: "Page 1 of 5" with [< Previous] [Next >] buttons
3. Click "Next" to go to page 2
4. Results update to show items 11-20
5. Click "Previous" to go back
```

---

## Technical Flow (Developer Perspective)

### Component Files

| File | Type | Purpose |
|------|------|---------|
| `search/page.tsx` | Server | Authentication check, metadata, pass userId to client |
| `search-client.tsx` | Client | Main orchestrator, manages filters and refresh trigger |
| `search-form.tsx` | Client | Form UI with 5 input fields, Reset/Search buttons |
| `transaction-results.tsx` | Client | Table rendering, pagination, document grid |
| `download-button.tsx` | Client | Download PDF action button |
| `_actions/search.ts` | Server | `searchDocuments()` and `downloadDocumentPDF()` functions |

### Data Flow

```
User fills form → SearchForm state updates
     ↓
User clicks Search → handleSearch() calls onSearch()
     ↓
SearchClient.handleSearch updates parent state:
  • setFilters(newFilters)
  • setRefreshTrigger(++counter)
     ↓
TransactionResults useEffect fires (refreshTrigger changed)
     ↓
Calls searchDocuments(filters, page, limit) (Server Action)
     ↓
Server processes:
  1. getDocumentsByCreator(userId) - all docs user created
  2. getPendingApprovals(userRole) - all docs awaiting approval
  3. Combine + deduplicate
  4. Apply filters (documentNumber, type, status, date range)
  5. Sort by date (newest first)
  6. Paginate
     ↓
Returns { success: true, data: { data: [...], pagination: {...} } }
     ↓
setDocuments() + setPagination() in TransactionResults
     ↓
Table re-renders with results
     ↓
User sees: [Document # | Type | Status | Created | Actions] table
```

---

## State Variables

### SearchClient
```typescript
filters           // Current search filters
refreshTrigger    // Counter to trigger useEffect in results
isSearching       // Button loading state
```

### SearchForm
```typescript
documentNumber    // User's text input
documentType      // Dropdown selection
status            // Dropdown selection
startDate         // Date picker value
endDate           // Date picker value
```

### TransactionResults
```typescript
documents         // Array of search results
pagination        // { page, limit, total, totalPages }
isLoading         // Loading spinner state
sorting           // TanStack table sorting state
columnFilters     // TanStack table column filter state
columnVisibility  // TanStack table visibility state
```

---

## Filter Types

### Document Number
- **Type:** Text input
- **Match:** Case-insensitive substring (contains)
- **Example:** Type "REQ" matches "REQ-2024-001", "REQ-2024-002", etc.

### Document Type
- **Type:** Dropdown
- **Options:** ALL, REQUISITION, PURCHASE_ORDER, PAYMENT_VOUCHER, GOODS_RECEIVED_NOTE
- **Match:** Exact match (or ALL = no filter)

### Status
- **Type:** Dropdown
- **Options:** ALL, DRAFT, SUBMITTED, IN_REVIEW, APPROVED, REJECTED, REVERSED
- **Match:** Exact match (or ALL = no filter)

### Start Date
- **Type:** Date picker
- **Range:** Inclusive (00:00:00 on that day)
- **Example:** Jan 1, 2024 includes all events from Jan 1 00:00:00 onwards

### End Date
- **Type:** Date picker
- **Range:** Inclusive (23:59:59 on that day)
- **Example:** Jan 31, 2024 includes all events through Jan 31 23:59:59

---

## Server Actions

### searchDocuments(filters, page, limit)

**Input:**
```typescript
filters: {
  documentNumber: string,      // "REQ", "PO-2024", etc.
  documentType: 'ALL' | DocType,
  status: 'ALL' | DocStatus,
  startDate: string,           // "2024-01-01"
  endDate: string              // "2024-12-31"
},
page: number,
limit: number
```

**Process:**
1. Fetch from documentStore where `createdBy === userId`
2. Fetch from documentStore where `status === 'IN_REVIEW'` AND user's role is in approvers
3. Merge arrays
4. Deduplicate by ID
5. Filter by each criterion
6. Sort by createdAt DESC
7. Slice by page/limit
8. Return with pagination metadata

**Output:**
```typescript
{
  success: true,
  data: {
    data: [WorkflowDocument[], ...],
    pagination: {
      page: 1,
      limit: 10,
      total: 47,
      totalPages: 5
    }
  }
}
```

### downloadDocumentPDF(documentId)

**Input:** `documentId: string`

**Process:**
1. Verify session
2. Call getDocument(documentId) to verify exists
3. Generate mock download URL: `/api/documents/{id}/download`
4. Return success response

**Output:**
```typescript
{
  success: true,
  data: {
    downloadUrl: "/api/documents/{documentId}/download"
  }
}
```

---

## Table Columns

| Column | Data Type | Sortable | Content |
|--------|-----------|----------|---------|
| Document # | String | ✅ | REQ-2024-001, PO-2024-0042, etc. |
| Type | String | ❌ | "Requisition", "Purchase Order", "Payment Voucher", "GRN" |
| Status | Badge | ❌ | Colored badge (Draft=outline, Approved=green, Rejected=red) |
| Created | DateTime | ✅ | "12/4/2024 2:30 PM" |
| Actions | Buttons | ❌ | [View] [Download] |

---

## Pagination

**Default:** 10 documents per page

**Example:**
- Total documents: 47
- Page 1: Shows 1-10 of 47 documents
- Page 2: Shows 11-20 of 47 documents
- Page 3: Shows 21-30 of 47 documents
- Page 4: Shows 31-40 of 47 documents
- Page 5: Shows 41-47 of 47 documents

**Button States:**
- Page 1: Previous button DISABLED
- Page 5: Next button DISABLED
- Any page while loading: Both buttons DISABLED

---

## URL Mappings (View Action)

When clicking View button on a result:

```javascript
const typeSlug = {
  'REQUISITION': 'requisitions',
  'PURCHASE_ORDER': 'purchase-orders',
  'PAYMENT_VOUCHER': 'payment-vouchers',
  'GOODS_RECEIVED_NOTE': 'grn'
}[doc.type];

router.push(`/${typeSlug}/${doc.id}`);
```

**Examples:**
- Requisition ID "req-123" → `/requisitions/req-123`
- Purchase Order ID "po-456" → `/purchase-orders/po-456`
- Payment Voucher ID "pv-789" → `/payment-vouchers/pv-789`
- GRN ID "grn-101" → `/grn/grn-101`

---

## Status Badge Colors

```typescript
DRAFT         → outline (gray border)
SUBMITTED     → secondary (gray fill)
IN_REVIEW     → default (blue)
APPROVED      → default (blue)
REJECTED      → destructive (red)
REVERSED      → secondary (gray)
```

---

## Common Scenarios

### Scenario: Find all approved POs from Q1 2024

```
1. Document Type: Purchase Orders
2. Status: Approved
3. Start Date: 2024-01-01
4. End Date: 2024-03-31
5. Click Search
→ Results: All POs approved between Jan 1 and Mar 31, 2024
```

### Scenario: Search for specific requisition by number

```
1. Document Number: REQ-2024-0042
2. Click Search
→ Results: Document with that exact number (if exists)
   OR: Multiple documents containing that substring
```

### Scenario: Find all documents awaiting my approval

```
1. Leave all filters as "ALL"
2. Click Search
→ Results:
   - All documents I created (regardless of status)
   - All documents awaiting MY approval (based on my role)
```

### Scenario: Browse all docs with pagination

```
1. Click Search (with all filters as ALL)
2. See first 10 results
3. Click Next → see next 10 results
4. Continue until reach last page
```

---

## Storage & Data Sources

### Data Sources Combined

**Source 1: getDocumentsByCreator()**
```
Query: documentStore.filter(doc => doc.createdBy === userId)
Returns: ALL documents created by this user (any status)
Purpose: Users see "my submissions"
```

**Source 2: getPendingApprovals()**
```
Query: documentStore.filter(doc =>
  doc.status === 'IN_REVIEW' &&
  approversStore[doc.id].some(a => a.role === userRole && a.stepOrder === doc.currentStage)
)
Returns: Documents awaiting this user's approval at current stage
Purpose: Users see "documents waiting for me"
```

**Combination:**
```
allDocuments = [...createdDocs, ...pendingDocs]
Deduplicate by ID (Map)
Result: Union of both sources
```

---

## Edge Cases

### No Results Found

```
Condition: filters return empty array
Display: "No documents found" with icon
Action: User adjusts filters and tries again
```

### Invalid Date Range

```
Example: Start Date = Dec 31, 2024; End Date = Jan 1, 2024
Result: No results (start > end)
Fix: Swap dates or adjust
```

### Very Large Result Set

```
Scenario: User leaves all filters as "ALL"
Result: All their documents + pending approvals
Pagination: Spreads across multiple pages
Memory: Each page fetched on demand (no lazy loading in current impl)
```

### Session Expired

```
If: User's auth token expired
Then: searchDocuments() returns 401 Unauthorized
Display: User redirected to /login
```

---

## Performance Notes

### Current (MVP)

- **In-memory storage:** All documents kept in JavaScript Map
- **Fetch strategy:** getDocumentsByCreator fetches up to 1000 at once
- **Filter timing:** All filtering done in-memory (O(n) complexity)
- **Sort:** O(n log n) by createdAt
- **Pagination:** Client-side array slicing (no DB offset/limit)

### Production Improvements

- Use proper database (PostgreSQL, MongoDB, etc.)
- Add indexes: createdBy, status, type, createdAt
- Push filtering to database layer
- Implement server-side pagination with LIMIT/OFFSET
- Add Redis caching for frequent searches
- Implement full-text search index for documentNumber

---

## Testing Checklist

- [ ] Search by document number substring
- [ ] Search by exact document type
- [ ] Search by status
- [ ] Search by date range (inclusive boundaries)
- [ ] Multi-filter search (all 5 filters together)
- [ ] Reset button clears all fields
- [ ] Pagination: Next/Previous buttons work
- [ ] View button navigates to correct document type page
- [ ] Download button initiates PDF download
- [ ] No results message displays when appropriate
- [ ] Sorting by Document # column
- [ ] Sorting by Created date column
- [ ] Table shows correct columns and data
- [ ] Status badges have correct colors
- [ ] Document type labels display correctly

---

## File Structure

```
src/
├── app/
│   ├── (private)/
│   │   └── (main)/
│   │       └── search/
│   │           ├── page.tsx                    (Server)
│   │           └── _components/
│   │               ├── search-client.tsx       (Client orchestrator)
│   │               ├── search-form.tsx         (Form UI)
│   │               ├── transaction-results.tsx (Table + pagination)
│   │               └── download-button.tsx     (Download action)
│   │
│   └── _actions/
│       ├── search.ts                           (Server actions)
│       └── workflow.ts                         (Data fetching helpers)
│
└── types/
    └── workflow.ts                             (Type definitions)
```

---

## Key Dependencies

- **TanStack React Table** - Table rendering, sorting, filtering
- **React DatePicker** - Date field input
- **Lucide React** - Icons (Search, Download, Eye, etc.)
- **Next.js** - Router, Server Actions
- **Sonner** - Toast notifications (for errors)

---

## Summary

✅ **What it does:** Search across all document types with flexible filtering
✅ **Who uses it:** Any logged-in user
✅ **When used:** When finding specific documents or browsing history
✅ **Key features:** Dual data source, pagination, quick actions
✅ **Data layer:** In-memory storage (MVP), ready for database migration
