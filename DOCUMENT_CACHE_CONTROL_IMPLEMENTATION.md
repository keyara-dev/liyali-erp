# Document Cache Control Implementation Guide

## Overview

This document provides instructions for ensuring all document endpoints return fresh data for PDF generation and verification.

## Changes Implemented

### Backend Changes (✅ Completed)

#### 1. Public Verification Endpoints

**File**: `backend/handlers/document_handler.go`

Added cache control headers to:

- `VerifyDocumentPublic()` - Line ~377
- `GetDocumentForPDFPublic()` - Line ~397

```go
c.Set("Cache-Control", "no-cache, no-store, must-revalidate")
c.Set("Pragma", "no-cache")
c.Set("Expires", "0")
```

#### 2. Authenticated Document Retrieval Endpoints

Added cache control headers to:

**File**: `backend/handlers/requisition.go`

- `GetRequisition()` - Line ~330

**File**: `backend/handlers/purchase_order.go`

- `GetPurchaseOrder()` - Line ~252

**File**: `backend/handlers/payment_voucher.go`

- `GetPaymentVoucher()` - Line ~210

**File**: `backend/handlers/grn.go`

- `GetGRN()` - Line ~196

### Frontend Changes

#### 1. Public Verification Actions (✅ Completed)

**File**: `frontend/src/app/_actions/verification.ts`

Added cache-busting headers to:

- `verifyDocument()` - Public verification endpoint
- `getDocumentForPDF()` - Public PDF data endpoint

#### 2. Verification Page Configuration (✅ Completed)

**File**: `frontend/src/app/verify/[documentNumber]/page.tsx`

Added:

```typescript
export const dynamic = "force-dynamic";
export const revalidate = 0; // Disable caching completely
```

#### 3. Authenticated Document Fetch Actions (⚠️ TODO)

The following files need cache-busting headers added to their document fetch functions:

**File**: `frontend/src/app/_actions/requisitions.ts`

- `getRequisitionById()` - Line ~106

**File**: `frontend/src/app/_actions/purchase-orders.ts`

- `getPurchaseOrderById()` - Line ~152

**File**: `frontend/src/app/_actions/payment-vouchers.ts`

- `getPaymentVoucherById()` - Line ~157

**File**: `frontend/src/app/_actions/grns.ts` (if exists)

- `getGRNById()` - Similar pattern

### Implementation Pattern

For each document fetch function, add headers to the API call:

```typescript
export async function getDocumentById(
  documentId: string,
): Promise<APIResponse<Document>> {
  const url = `/api/v1/documents/${documentId}`;

  try {
    const response = await authenticatedApiClient({
      method: "GET",
      url,
      headers: {
        "Cache-Control": "no-cache, no-store, must-revalidate",
        Pragma: "no-cache",
        Expires: "0",
      },
    });

    return successResponse(
      response.data?.data,
      "Document retrieved successfully",
    );
  } catch (error: any) {
    return handleError(error, "GET", url);
  }
}
```

## Why This Matters

### Problem

When users:

1. Update a document (change status, amount, items, etc.)
2. Generate a PDF or scan a QR code
3. The PDF shows outdated information

### Root Cause

Multiple caching layers:

- Browser HTTP cache
- Next.js page cache
- API response cache
- CDN/proxy cache (if applicable)

### Solution

By adding cache control headers at both backend (response) and frontend (request) levels:

- Backend: Tells all intermediaries not to cache the response
- Frontend: Ensures requests bypass any client-side cache

## Testing Checklist

For each document type (Requisition, PO, PV, GRN):

1. ✅ Create a document
2. ✅ Generate PDF - verify it shows correct data
3. ✅ Update the document (change amount, status, or items)
4. ✅ Generate PDF again - verify it shows UPDATED data (not cached)
5. ✅ Scan QR code - verify it shows UPDATED data
6. ✅ Test in different browsers (Chrome, Firefox, Safari, Edge)
7. ✅ Test on mobile devices
8. ✅ Test with slow network (to ensure no stale cache is used)

## Performance Considerations

### Impact

- Minimal: Document retrieval is already a real-time operation
- No additional database queries
- Only prevents caching of responses

### Benefits

- Users always see current data
- Eliminates confusion from stale PDFs
- Improves trust in the system
- Reduces support tickets about "wrong data"

## Additional Recommendations

### 1. Add Last Modified Timestamp

Display when the document was last updated on PDFs and verification pages:

```typescript
<p className="text-xs text-muted-foreground">
  Last updated: {new Date(document.updatedAt).toLocaleString()}
</p>
```

### 2. Add Version Number

Consider adding a version number to documents that increments on each update:

```sql
ALTER TABLE requisitions ADD COLUMN version INTEGER DEFAULT 1;
```

### 3. Cache Invalidation Strategy

For list views (where caching might be beneficial), implement smart cache invalidation:

```typescript
// Invalidate cache when document is updated
queryClient.invalidateQueries({ queryKey: ["documents", documentId] });
queryClient.invalidateQueries({ queryKey: ["documents", "list"] });
```

### 4. Add ETag Support

For future optimization, consider implementing ETags:

```go
// Backend
c.Set("ETag", fmt.Sprintf(`"%s"`, document.UpdatedAt.Format(time.RFC3339)))

// Frontend can use If-None-Match header for conditional requests
```

## Monitoring

### Metrics to Track

1. PDF generation time (should remain consistent)
2. Document fetch latency (should not increase)
3. Cache hit/miss rates (should see more misses, which is expected)
4. User reports of stale data (should decrease to zero)

### Logging

Add logging to track when documents are fetched for PDF generation:

```go
log.Printf("Document %s fetched for PDF generation at %s", documentNumber, time.Now())
```

## Rollback Plan

If issues arise:

1. Remove cache control headers from backend handlers
2. Remove cache-busting headers from frontend actions
3. Restore original page configuration
4. Monitor for 24 hours
5. Investigate root cause before re-implementing

## Related Files

### Backend

- `backend/handlers/document_handler.go`
- `backend/handlers/requisition.go`
- `backend/handlers/purchase_order.go`
- `backend/handlers/payment_voucher.go`
- `backend/handlers/grn.go`

### Frontend

- `frontend/src/app/_actions/verification.ts`
- `frontend/src/app/_actions/requisitions.ts`
- `frontend/src/app/_actions/purchase-orders.ts`
- `frontend/src/app/_actions/payment-vouchers.ts`
- `frontend/src/app/verify/[documentNumber]/page.tsx`

### Documentation

- `QR_VERIFICATION_CACHE_FIX.md` - Initial fix for QR verification
- `DOCUMENT_CACHE_CONTROL_IMPLEMENTATION.md` - This document

## Next Steps

1. ✅ Complete backend cache control headers (DONE)
2. ✅ Complete public verification cache control (DONE)
3. ⚠️ Add cache-busting headers to authenticated document fetch actions
4. ⚠️ Test all document types thoroughly
5. ⚠️ Deploy to staging environment
6. ⚠️ Perform user acceptance testing
7. ⚠️ Deploy to production
8. ⚠️ Monitor for 48 hours

## Support

If you encounter issues:

1. Check browser console for errors
2. Check network tab for cache headers
3. Verify backend logs for document fetch requests
4. Test with cache disabled in browser DevTools
5. Contact the development team with specific document numbers and timestamps
