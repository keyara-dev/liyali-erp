# Logout Storage Cleanup Implementation

**Date:** February 8, 2026  
**Status:** ✅ IMPLEMENTED

---

## Overview

Implemented comprehensive localStorage cleanup when users log out to ensure no organizational data persists between sessions. This prevents data leakage and ensures a clean state for the next user.

---

## What Gets Cleared

### 1. Organization Data

- `current-organization-id` - The currently selected organization

### 2. Document Storage

- `liyali-requisitions` - All requisition drafts
- `liyali-purchase-orders` - All purchase order drafts
- `liyali-payment-vouchers` - All payment voucher drafts
- `liyali-goods-received-notes` - All GRN drafts
- `liyali-budgets` - All budget drafts
- `liyali-requisition-action-history` - Action history

### 3. Permission Cache

- All keys starting with `permissions_*`
- All keys starting with `permissions_expiry_*`

---

## Implementation

### Centralized Utility Function

Created `frontend/src/lib/storage/clear-storage.ts` with three functions:

```typescript
// Clear everything on logout
clearOrganizationalData();

// Clear only documents (for org switching)
clearDocumentStorage();

// Clear only permissions (for role changes)
clearPermissionCache();
```

### Integration Points

#### 1. Manual Logout (`useLogout` hook)

**File:** `frontend/src/hooks/use-organization-mutations.ts`

```typescript
export function useLogout() {
  const mutation = useMutation({
    mutationFn: async () => {
      await logoutAction();
    },
    onSuccess: () => {
      // Clear all organizational data from localStorage
      if (typeof window !== "undefined") {
        try {
          localStorage.removeItem("current-organization-id");

          const storageKeys = [
            "liyali-requisitions",
            "liyali-purchase-orders",
            "liyali-payment-vouchers",
            "liyali-goods-received-notes",
            "liyali-budgets",
            "liyali-requisition-action-history",
          ];

          storageKeys.forEach((key) => localStorage.removeItem(key));

          // Clear permission cache
          const allKeys = Object.keys(localStorage);
          const permissionKeys = allKeys.filter(
            (key) =>
              key.startsWith("permissions_") ||
              key.startsWith("permissions_expiry_"),
          );
          permissionKeys.forEach((key) => localStorage.removeItem(key));

          console.log("✅ Cleared all organizational data from localStorage");
        } catch (error) {
          console.error("Failed to clear localStorage on logout:", error);
        }
      }

      router.push("/login");
    },
  });
}
```

#### 2. Session Timeout (`handleUserLogOut` in screen-lock)

**File:** `frontend/src/components/base/screen-lock.tsx`

```typescript
const handleUserLogOut = useCallback(async () => {
  // ... existing logout logic ...

  // Clear all organizational data from localStorage
  if (typeof window !== "undefined") {
    try {
      localStorage.removeItem("current-organization-id");

      const storageKeys = [
        "liyali-requisitions",
        "liyali-purchase-orders",
        "liyali-payment-vouchers",
        "liyali-goods-received-notes",
        "liyali-budgets",
        "liyali-requisition-action-history",
      ];

      storageKeys.forEach((key) => localStorage.removeItem(key));

      // Clear permission cache
      const allKeys = Object.keys(localStorage);
      const permissionKeys = allKeys.filter(
        (key) =>
          key.startsWith("permissions_") ||
          key.startsWith("permissions_expiry_"),
      );
      permissionKeys.forEach((key) => localStorage.removeItem(key));

      logger.info("✅ Cleared all organizational data from localStorage");
    } catch (error) {
      logger.error("Failed to clear localStorage on logout", error);
    }
  }

  window.location.replace("/login");
}, [broadcastState, setIsIdle]);
```

---

## Benefits

### 1. Security

- No sensitive organizational data persists after logout
- Prevents data leakage between users on shared devices
- Ensures clean state for each session

### 2. Data Integrity

- Prevents stale data from previous sessions
- Avoids confusion with cached permissions
- Ensures fresh data load on next login

### 3. User Experience

- Clean slate for each user
- No unexpected data from previous sessions
- Proper organization context on login

---

## Testing Scenarios

### Test 1: Manual Logout

1. Login as user A
2. Create some draft documents
3. Switch organizations
4. Logout
5. Check localStorage - should be empty of org data

### Test 2: Session Timeout

1. Login as user A
2. Create some draft documents
3. Wait for session timeout (or trigger idle timeout)
4. Check localStorage - should be empty of org data

### Test 3: Multiple Users

1. Login as user A (Org 1)
2. Create drafts
3. Logout
4. Login as user B (Org 2)
5. Verify no data from user A is visible

---

## Future Enhancements

### 1. Selective Clearing on Organization Switch

Currently, switching organizations doesn't clear localStorage. Consider:

- Clearing only document storage (not permissions)
- Keeping user preferences
- Implementing organization-scoped storage keys

### 2. Encrypted Storage

For sensitive data, consider:

- Encrypting localStorage data
- Using sessionStorage for temporary data
- Implementing secure storage patterns

### 3. Storage Quota Management

- Monitor localStorage usage
- Implement cleanup strategies for old data
- Add warnings when approaching quota limits

---

## Related Files

- `frontend/src/lib/storage/clear-storage.ts` - Utility functions
- `frontend/src/hooks/use-organization-mutations.ts` - Manual logout
- `frontend/src/components/base/screen-lock.tsx` - Session timeout
- `frontend/src/stores/organization-store.ts` - Organization state
- `frontend/src/lib/storage/storage.ts` - Document storage
- `frontend/src/hooks/use-permissions.ts` - Permission cache

---

## Verification

To verify the implementation is working:

```javascript
// Before logout
console.log(localStorage.getItem("current-organization-id"));
console.log(localStorage.getItem("liyali-requisitions"));
console.log(
  Object.keys(localStorage).filter((k) => k.startsWith("permissions_")),
);

// After logout - all should be null/empty
```

---

**Status:** Implementation complete and ready for testing! ✅
