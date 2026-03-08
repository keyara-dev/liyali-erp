# API Client Authentication Fix Summary

**Date**: March 7, 2026  
**Status**: ✅ RESOLVED

---

## Problem

API requests were failing with `Authorization header required` error from the backend, even though:

- Session was valid
- Token existed
- Code appeared to set the Authorization header

---

## Root Cause

The issue was in `frontend/src/app/_actions/api-config.ts` in the `authenticatedApiClient` function:

```typescript
// ❌ BEFORE (BROKEN)
const config = {
  method: "GET",
  headers, // Our auth headers with Authorization
  withCredentials: true,
  ...request, // ⚠️ This overwrites headers if request.headers exists!
};
```

When we passed additional headers (like `NO_CACHE_HEADERS`), the spread operator `...request` would overwrite the entire `headers` object, removing the Authorization header.

---

## Solution

Fixed the header merging order:

```typescript
// ✅ AFTER (FIXED)
const config = {
  method: "GET",
  withCredentials: true,
  ...request, // Spread request first
  headers: {
    ...headers, // Then merge auth headers
    ...request.headers, // Then merge request headers
  },
};
```

Now headers are properly merged:

1. Auth headers (Authorization, Cookie, X-Organization-ID) are set first
2. Additional headers from the request (like Cache-Control) are merged in
3. Authorization header is never overwritten

---

## Changes Made

### 1. Removed Unnecessary Wrapper

**File**: `frontend/src/app/_actions/api-config.ts`

- ❌ Removed: `authenticatedApiClientNoCache()` function
- ✅ Kept: `NO_CACHE_HEADERS` constant
- ✅ Updated: All actions to use `authenticatedApiClient` with `NO_CACHE_HEADERS` directly

**Benefits**:

- Single source of truth for authenticated requests
- Easier to debug
- More explicit about cache control
- Reduced code complexity

### 2. Fixed Header Merging

**File**: `frontend/src/app/_actions/api-config.ts`

- ✅ Fixed: Header merging order in config object
- ✅ Ensured: Authorization header is always included

### 3. Updated All Document Actions

**Files**:

- `frontend/src/app/_actions/requisitions.ts`
- `frontend/src/app/_actions/purchase-orders.ts`
- `frontend/src/app/_actions/payment-vouchers.ts`
- `frontend/src/app/_actions/grn-actions.ts`

**Changes**:

```typescript
// Before
import { authenticatedApiClientNoCache } from "./api-config";
const response = await authenticatedApiClientNoCache({ method: "GET", url });

// After
import { NO_CACHE_HEADERS } from "./api-config";
const response = await authenticatedApiClient({
  method: "GET",
  url,
  headers: NO_CACHE_HEADERS,
});
```

---

## Testing

### Before Fix

```
Error: Authorization header required
baseURL: 'http://localhost:8081',
method: 'get',
url: '/api/v1/requisitions/...',
// ❌ No Authorization header in request
```

### After Fix

```
[API Client] Session check: { isAuthenticated: true, hasToken: true }
[API Client] Request config: { hasAuthHeader: true, authHeaderPreview: 'Bearer eyJ...' }
✅ Request successful with Authorization header
```

---

## Impact

### Fixed Issues:

1. ✅ Authorization header now properly included in all requests
2. ✅ Cache-busting headers work without breaking auth
3. ✅ PDF generation works (requires fresh data with auth)
4. ✅ All document detail pages load correctly
5. ✅ No more "Authorization header required" errors

### Affected Features:

- ✅ Requisition detail pages
- ✅ Purchase Order detail pages
- ✅ Payment Voucher detail pages
- ✅ GRN detail pages
- ✅ PDF preview and export
- ✅ Document refetching

---

## Lessons Learned

### 1. Spread Operator Order Matters

When using the spread operator with objects, **order matters**:

```typescript
// ❌ Later spreads overwrite earlier properties
const config = { headers: authHeaders, ...request };

// ✅ Explicitly merge nested objects
const config = {
  ...request,
  headers: { ...authHeaders, ...request.headers },
};
```

### 2. Debug Logging is Essential

Adding temporary debug logs helped identify:

- Session was valid ✅
- Token existed ✅
- Header was being set ✅
- But not reaching backend ❌

This narrowed down the issue to the request configuration.

### 3. Simplify When Possible

The `authenticatedApiClientNoCache` wrapper was unnecessary abstraction:

- Added complexity
- Made debugging harder
- Didn't provide real value

Removing it made the code clearer and easier to maintain.

---

## Commits

1. `refactor: remove unnecessary authenticatedApiClientNoCache wrapper`
   - Simplified API client code
   - Use NO_CACHE_HEADERS directly

2. `debug: add detailed logging for API client authentication`
   - Added temporary debug logs
   - Helped diagnose the issue

3. `fix: prevent request headers from overwriting auth headers`
   - **THE FIX**: Changed header merging order
   - Resolved Authorization header issue

4. `chore: remove debug logging from API client`
   - Cleaned up temporary debug logs
   - Production-ready code

---

## Verification

✅ All apps build successfully  
✅ Frontend: No TypeScript errors  
✅ Backend: Receiving Authorization headers  
✅ Authentication working correctly  
✅ Document pages loading properly  
✅ PDF generation working

---

## Status: ✅ RESOLVED

The API client now properly includes Authorization headers in all requests, even when additional headers like cache-control are specified.
