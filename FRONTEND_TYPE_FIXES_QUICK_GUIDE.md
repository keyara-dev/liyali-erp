# Frontend Type Fixes - Quick Implementation Guide

## Overview
This guide provides step-by-step instructions to fix the 2 critical type issues and 4 high-priority import issues identified in the audit.

**Total Fix Time**: ~17 minutes
**Difficulty**: Easy
**Risk Level**: Low (all changes are safe)

---

## 🔴 CRITICAL FIX #1: UserRole Type Mismatch in auth.ts

### File: `frontend/src/lib/auth.ts`

### Current Code (BROKEN)
```typescript
// Lines 1-30
'use server'
import "server-only";

import { SignJWT, jwtVerify } from "jose";
import { cookies } from "next/headers";

import type { AuthSession, User, UserType, AuthUser } from "@/types";
import { SESSION_CONFIG } from "@/lib/session-config";
import {
  AUTH_SESSION,
  USER_SESSION,
  PERMISSIONS_SESSION,
  SCREEN_LOCK_SESSION,
} from "@/lib/constants";

// ============================================================================
// TYPES
// ============================================================================

// Re-export AuthUser from types for backward compatibility
export type { AuthUser } from "@/types";

export type UserRole =                    // ❌ DELETE THIS ENTIRE TYPE
  | "requester"
  | "approver" 
  | "finance"
  | "admin"
  | "viewer";
```

### Fixed Code (CORRECT)
```typescript
// Lines 1-30
'use server'
import "server-only";

import { SignJWT, jwtVerify } from "jose";
import { cookies } from "next/headers";

import type { AuthSession, User, UserType, AuthUser, UserRole } from "@/types";  // ✅ ADD UserRole HERE
import { SESSION_CONFIG } from "@/lib/session-config";
import {
  AUTH_SESSION,
  USER_SESSION,
  PERMISSIONS_SESSION,
  SCREEN_LOCK_SESSION,
} from "@/lib/constants";

// ============================================================================
// TYPES
// ============================================================================

// Re-export AuthUser from types for backward compatibility
export type { AuthUser } from "@/types";

// ✅ REMOVED: export type UserRole = ... (no longer needed)
```

### Changes Required
1. **Line 7**: Add `UserRole` to the import from `@/types`
   - Change: `import type { AuthSession, User, UserType, AuthUser } from "@/types";`
   - To: `import type { AuthSession, User, UserType, AuthUser, UserRole } from "@/types";`

2. **Lines 25-29**: Delete the local `UserRole` type definition
   - Delete these lines entirely:
   ```typescript
   export type UserRole =
     | "requester"
     | "approver" 
     | "finance"
     | "admin"
     | "viewer";
   ```

### Verification
After making changes, verify:
```bash
# Check for TypeScript errors
npx tsc --noEmit

# Should show no errors related to UserRole
```

---

## 🔴 CRITICAL FIX #2: ApprovalRecord Import in approval-chain-panel.tsx

### File: `frontend/src/app/(private)/(main)/budgets/[id]/_components/approval-chain-panel.tsx`

### Current Code (WRONG)
```typescript
// Lines 1-10
"use client";

import { Button } from "@/components";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { ApprovalRecord } from "@/types/budget";  // ❌ WRONG MODULE
import {
  CheckCircle2,
  ClipboardListIcon,
  Clock,
  Plus,
  XCircle,
} from "lucide-react";
```

### Fixed Code (CORRECT)
```typescript
// Lines 1-10
"use client";

import { Button } from "@/components";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { ApprovalRecord } from "@/types";  // ✅ CORRECT MODULE
import {
  CheckCircle2,
  ClipboardListIcon,
  Clock,
  Plus,
  XCircle,
} from "lucide-react";
```

### Changes Required
1. **Line 5**: Change import source
   - Change: `import { ApprovalRecord } from "@/types/budget";`
   - To: `import { ApprovalRecord } from "@/types";`

### Verification
After making changes, verify:
```bash
# Check for TypeScript errors
npx tsc --noEmit

# Should show no errors related to ApprovalRecord
```

---

## 🟡 HIGH PRIORITY FIX #3: User Import in use-session.ts

### File: `frontend/src/hooks/use-session.ts`

### Current Code (WRONG)
```typescript
// Lines 1-15
import { useCallback, useEffect, useState } from "react";
import {
  checkIsAdminAction,
} from "@/app/_actions/session";
import type { User } from "@/types/auth";  // ❌ WRONG MODULE

export interface SessionData {
  user: User | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  isAdmin: boolean;
  checkIsAdmin: () => Promise<boolean>;
}
```

### Fixed Code (CORRECT)
```typescript
// Lines 1-15
import { useCallback, useEffect, useState } from "react";
import {
  checkIsAdminAction,
} from "@/app/_actions/session";
import type { User } from "@/types";  // ✅ CORRECT MODULE

export interface SessionData {
  user: User | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  isAdmin: boolean;
  checkIsAdmin: () => Promise<boolean>;
}
```

### Changes Required
1. **Line 5**: Change import source
   - Change: `import type { User } from "@/types/auth";`
   - To: `import type { User } from "@/types";`

---

## 🟡 HIGH PRIORITY FIX #4: User Import in use-permissions.ts

### File: `frontend/src/hooks/use-permissions.ts`

### Current Code (WRONG)
```typescript
// Lines 1-10
import { useMemo } from "react";
import { useSession } from "./use-session";
import type { User } from "@/types/auth";  // ❌ WRONG MODULE

/**
 * Hook to check user permissions
 */
export function usePermissions() {
```

### Fixed Code (CORRECT)
```typescript
// Lines 1-10
import { useMemo } from "react";
import { useSession } from "./use-session";
import type { User } from "@/types";  // ✅ CORRECT MODULE

/**
 * Hook to check user permissions
 */
export function usePermissions() {
```

### Changes Required
1. **Line 3**: Change import source
   - Change: `import type { User } from "@/types/auth";`
   - To: `import type { User } from "@/types";`

---

## 🟡 HIGH PRIORITY FIX #5: User Import in user-actions.ts

### File: `frontend/src/app/_actions/user-actions.ts`

### Current Code (WRONG)
```typescript
// Lines 1-15
"use server";

import { revalidatePath } from "next/cache";
import type { APIResponse } from "@/types";
import authenticatedApiClient, {
  handleError,
  successResponse,
  badRequestResponse,
} from "./api-config";
import { User, UserType } from "@/types/auth";  // ❌ WRONG MODULE
```

### Fixed Code (CORRECT)
```typescript
// Lines 1-15
"use server";

import { revalidatePath } from "next/cache";
import type { APIResponse } from "@/types";
import authenticatedApiClient, {
  handleError,
  successResponse,
  badRequestResponse,
} from "./api-config";
import { User, UserType } from "@/types";  // ✅ CORRECT MODULE
```

### Changes Required
1. **Line 10**: Change import source
   - Change: `import { User, UserType } from "@/types/auth";`
   - To: `import { User, UserType } from "@/types";`

---

## 🟡 HIGH PRIORITY FIX #6: User Import in session.ts

### File: `frontend/src/app/_actions/session.ts`

### Current Code (WRONG)
```typescript
// Lines 1-10
"use server";

import { getCurrentUser, hasRole, isAdmin, verifySession } from "@/lib/auth";
import type { User } from "@/types/auth";  // ❌ WRONG MODULE

/**
 * Check if current user is admin
 */
export async function checkIsAdminAction(): Promise<boolean> {
```

### Fixed Code (CORRECT)
```typescript
// Lines 1-10
"use server";

import { getCurrentUser, hasRole, isAdmin, verifySession } from "@/lib/auth";
import type { User } from "@/types";  // ✅ CORRECT MODULE

/**
 * Check if current user is admin
 */
export async function checkIsAdminAction(): Promise<boolean> {
```

### Changes Required
1. **Line 4**: Change import source
   - Change: `import type { User } from "@/types/auth";`
   - To: `import type { User } from "@/types";`

---

## Implementation Checklist

### Step 1: Fix Critical Issues (5 minutes)
- [ ] Fix `frontend/src/lib/auth.ts` (UserRole type)
- [ ] Fix `frontend/src/app/(private)/(main)/budgets/[id]/_components/approval-chain-panel.tsx` (ApprovalRecord import)

### Step 2: Fix High Priority Issues (10 minutes)
- [ ] Fix `frontend/src/hooks/use-session.ts` (User import)
- [ ] Fix `frontend/src/hooks/use-permissions.ts` (User import)
- [ ] Fix `frontend/src/app/_actions/user-actions.ts` (User import)
- [ ] Fix `frontend/src/app/_actions/session.ts` (User import)

### Step 3: Verification (2 minutes)
- [ ] Run TypeScript compiler: `npx tsc --noEmit`
- [ ] Run ESLint: `npx eslint frontend/src --ext .ts,.tsx`
- [ ] Check for any remaining errors

### Step 4: Testing (Optional but Recommended)
- [ ] Test with users having extended roles (department_manager, finance_manager, etc.)
- [ ] Test approval workflow
- [ ] Test budget approval chain display
- [ ] Verify no console errors

---

## Automated Fix Script

If you prefer to automate these changes, you can use this script:

```bash
#!/bin/bash

# Fix 1: auth.ts - Add UserRole to import
sed -i 's/import type { AuthSession, User, UserType, AuthUser } from "@\/types";/import type { AuthSession, User, UserType, AuthUser, UserRole } from "@\/types";/' frontend/src/lib/auth.ts

# Fix 1: auth.ts - Remove local UserRole type (lines 25-29)
sed -i '25,29d' frontend/src/lib/auth.ts

# Fix 2: approval-chain-panel.tsx
sed -i 's/import { ApprovalRecord } from "@\/types\/budget";/import { ApprovalRecord } from "@\/types";/' frontend/src/app/\(private\)/\(main\)/budgets/\[id\]/_components/approval-chain-panel.tsx

# Fix 3: use-session.ts
sed -i 's/import type { User } from "@\/types\/auth";/import type { User } from "@\/types";/' frontend/src/hooks/use-session.ts

# Fix 4: use-permissions.ts
sed -i 's/import type { User } from "@\/types\/auth";/import type { User } from "@\/types";/' frontend/src/hooks/use-permissions.ts

# Fix 5: user-actions.ts
sed -i 's/import { User, UserType } from "@\/types\/auth";/import { User, UserType } from "@\/types";/' frontend/src/app/_actions/user-actions.ts

# Fix 6: session.ts
sed -i 's/import type { User } from "@\/types\/auth";/import type { User } from "@\/types";/' frontend/src/app/_actions/session.ts

echo "All fixes applied!"
npx tsc --noEmit
```

---

## Verification Commands

After applying fixes, run these commands to verify:

```bash
# Check TypeScript compilation
npx tsc --noEmit

# Check ESLint
npx eslint frontend/src --ext .ts,.tsx

# Check for remaining import issues
grep -r "from \"@/types/auth\"" frontend/src --include="*.ts" --include="*.tsx" | grep -v "node_modules"
grep -r "from \"@/types/budget\"" frontend/src --include="*.ts" --include="*.tsx" | grep -v "node_modules"

# Verify auth.ts doesn't have local UserRole
grep -n "export type UserRole" frontend/src/lib/auth.ts
```

---

## Expected Results

After applying all fixes:

✅ **No TypeScript errors**
```
$ npx tsc --noEmit
# (no output = success)
```

✅ **No ESLint errors related to imports**
```
$ npx eslint frontend/src --ext .ts,.tsx
# (no import-related errors)
```

✅ **All imports use centralized types**
```
$ grep -r "from \"@/types\"" frontend/src | wc -l
# (should show many results)
```

✅ **No problematic imports remain**
```
$ grep -r "from \"@/types/auth\"" frontend/src --include="*.ts" --include="*.tsx" | grep -v "node_modules"
# (should show no results)
```

---

## Rollback Instructions

If you need to rollback any changes:

```bash
# Restore from git
git checkout frontend/src/lib/auth.ts
git checkout frontend/src/app/\(private\)/\(main\)/budgets/\[id\]/_components/approval-chain-panel.tsx
git checkout frontend/src/hooks/use-session.ts
git checkout frontend/src/hooks/use-permissions.ts
git checkout frontend/src/app/_actions/user-actions.ts
git checkout frontend/src/app/_actions/session.ts
```

---

## Support

If you encounter any issues:

1. **TypeScript errors**: Run `npx tsc --noEmit` to see detailed errors
2. **Import not found**: Verify the file exists at `frontend/src/types/index.ts`
3. **Type mismatch**: Check that `UserRole` is properly exported from `@/types/core.ts`

---

## Summary

| Fix | File | Change | Time |
|-----|------|--------|------|
| 1 | auth.ts | Add UserRole to import, remove local type | 2 min |
| 2 | approval-chain-panel.tsx | Change ApprovalRecord import | 1 min |
| 3 | use-session.ts | Change User import | 1 min |
| 4 | use-permissions.ts | Change User import | 1 min |
| 5 | user-actions.ts | Change User import | 1 min |
| 6 | session.ts | Change User import | 1 min |
| **Total** | **6 files** | **6 changes** | **~7 min** |

**Verification**: 2-3 minutes
**Total Time**: ~10 minutes

---

**Status**: Ready to implement
**Risk Level**: Low (all changes are safe and reversible)
**Confidence**: 95%
