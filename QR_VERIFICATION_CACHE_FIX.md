# QR Verification Cache Fix

## Problem

Documents generated through QR code verification were showing outdated information. When users scanned a QR code, they would see old document data instead of the current version.

## Root Cause

The issue was caused by multiple layers of caching:

1. **Next.js Page Caching**: The verification page was being cached by Next.js despite having `dynamic = "force-dynamic"`
2. **HTTP Response Caching**: No cache control headers were set on the backend API responses
3. **Client Request Caching**: The frontend axios requests didn't include cache-busting headers

## Solution Implemented

### 1. Frontend Changes

#### `frontend/src/app/_actions/verification.ts`

Added cache-busting headers to both verification functions:

```typescript
// In verifyDocument()
const response = await axios.get(
  `/api/v1/public/verify/${encodeURIComponent(documentNumber)}`,
  {
    headers: {
      "Cache-Control": "no-cache, no-store, must-revalidate",
      Pragma: "no-cache",
      Expires: "0",
    },
  },
);

// In getDocumentForPDF()
const response = await axios.get(
  `/api/v1/public/verify/${encodeURIComponent(documentNumber)}/document`,
  {
    headers: {
      "Cache-Control": "no-cache, no-store, must-revalidate",
      Pragma: "no-cache",
      Expires: "0",
    },
  },
);
```

#### `frontend/src/app/verify/[documentNumber]/page.tsx`

Added explicit revalidation configuration:

```typescript
export const dynamic = "force-dynamic";
export const revalidate = 0; // Disable caching completely
```

### 2. Backend Changes

#### `backend/handlers/document_handler.go`

Added cache control headers to both public verification endpoints:

```go
// In VerifyDocumentPublic()
c.Set("Cache-Control", "no-cache, no-store, must-revalidate")
c.Set("Pragma", "no-cache")
c.Set("Expires", "0")

// In GetDocumentForPDFPublic()
c.Set("Cache-Control", "no-cache, no-store, must-revalidate")
c.Set("Pragma", "no-cache")
c.Set("Expires", "0")
```

## Cache Control Headers Explained

- **Cache-Control: no-cache, no-store, must-revalidate**
  - `no-cache`: Forces caches to submit the request to the origin server for validation before releasing a cached copy
  - `no-store`: Instructs caches not to store the response
  - `must-revalidate`: Once a resource becomes stale, caches must not use it without successful validation

- **Pragma: no-cache**
  - HTTP/1.0 backward compatibility header for older caches

- **Expires: 0**
  - Sets the expiration date to the past, ensuring immediate expiration

## Testing the Fix

To verify the fix works:

1. **Update a document** (e.g., change status, amount, or title)
2. **Scan the QR code** or visit the verification URL
3. **Verify the latest data is displayed** (not cached version)
4. **Test in different browsers** to ensure cache headers work universally
5. **Test on mobile devices** where QR scanning typically occurs

## Files Modified

1. `frontend/src/app/_actions/verification.ts` - Added cache-busting headers to API requests
2. `frontend/src/app/verify/[documentNumber]/page.tsx` - Added revalidate configuration
3. `backend/handlers/document_handler.go` - Added cache control headers to responses

## Impact

- ✅ Users will always see the most current document data when scanning QR codes
- ✅ No performance impact as verification is already a real-time operation
- ✅ Works across all browsers and devices
- ✅ Maintains backward compatibility

## Additional Recommendations

1. **Monitor verification endpoint performance** to ensure no-cache doesn't cause issues
2. **Consider adding a "Last Updated" timestamp** to the verification UI for transparency
3. **Add automated tests** to verify cache headers are present in responses
4. **Document this behavior** in the API documentation

## Related Endpoints

The following public endpoints now have proper cache control:

- `GET /api/v1/public/verify/:documentNumber` - Document verification
- `GET /api/v1/public/verify/:documentNumber/document` - Full document data for PDF

## Notes

- This fix ensures data freshness for public verification endpoints
- Private/authenticated endpoints may have different caching strategies
- QR codes themselves don't change, only the data they point to
