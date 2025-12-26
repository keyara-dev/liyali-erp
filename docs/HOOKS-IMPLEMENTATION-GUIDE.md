# Frontend Hooks Implementation Guide

**Status**: ✅ COMPLETE
**Last Updated**: 2025-12-26
**Pattern**: React Query Mutations & Hooks

---

## Overview

This guide documents the transformation of the frontend architecture to use React Query hooks and mutations instead of manual data fetching and state management. This ensures consistency, reusability, and better separation of concerns across the application.

---

## 🎯 Core Principle

**"No useEffect for data fetching. No direct server action calls in components."**

All data fetching, mutations, and server actions should be wrapped in specialized hooks (Query or Mutation).

---

## ✅ Implemented Hooks

### Organization & Logout Mutations

**File**: `frontend/src/hooks/use-organization-mutations.ts`

#### `useSelectOrganization()`
```typescript
const { selectOrganization, isPending, error } = useSelectOrganization();

// Usage in component
await selectOrganization(orgId);
```
- Switches to selected organization
- Auto-navigates to `/home` on success
- Integrates with `OrganizationContext`
- Used in: Welcome page, Workspace switcher

#### `useLogout()`
```typescript
const { logout, isPending, error } = useLogout();

// Usage in component
await logout();
```
- Clears session and authentication
- Auto-navigates to `/login` on success
- Used in: User menu, Nav user, Welcome page
- Replaces direct `/api/auth/signout` links

---

### Authentication Mutations

**File**: `frontend/src/hooks/use-auth-mutations.ts`

#### `useLoginMutation()`
```typescript
const { login, isPending, error } = useLoginMutation();

// Usage
try {
  const result = await login({ email, password });
  if (!result.success) {
    setError(result.message);
  }
} catch (err) {
  setError(err.message);
}
```
- Handles user login with email/password
- Auto-navigates to `/welcome` on success
- Manual error handling in component (caller decides what to do)
- Used in: Login form

#### `useSignupMutation()`
```typescript
const { signup, isPending, error } = useSignupMutation();

// Usage
const result = await signup({
  email,
  name,
  password,
  role: 'requester'
});
```
- Creates new user account
- Auto-navigates to `/welcome` on success
- Used in: Signup component

#### `useSendResetEmailMutation()`
```typescript
const { sendResetEmail, isPending, error } = useSendResetEmailMutation();

// Usage
await sendResetEmail({ email: userEmail });
```
- Sends password reset email
- No auto-navigation (caller handles redirect)
- Used in: Forgot password flow

#### `useResetPasswordMutation()`
```typescript
const { resetPassword, isPending, error } = useResetPasswordMutation();

// Usage
await resetPassword({ token, newPassword });
```
- Resets password with token
- Auto-navigates to `/login` on success
- Used in: Reset password flow

#### `useChangePasswordMutation()`
```typescript
const { changePassword, isPending, error } = useChangePasswordMutation();

// Usage
await changePassword({
  oldPassword,
  newPassword,
  confirmPassword
});
```
- Changes password for authenticated user
- Requires current password verification
- Used in: Account settings

---

## 📋 Components Updated

### 1. Login Form
**File**: `frontend/src/app/(auth)/login/_components/login-form.tsx`

**Before**:
```typescript
const [isLoading, setIsLoading] = useState(false);
const result = await loginAction(email, password);
// Manual router.push and state management
```

**After**:
```typescript
const { login, isPending } = useLoginMutation();
const result = await login({ email, password });
// Auto-handles navigation and loading state
```

**Benefits**:
- Cleaner component code
- Automatic loading state (`isPending`)
- Consistent error handling pattern
- No manual router navigation needed

---

### 2. Signup Component
**File**: `frontend/src/app/(auth)/_components/signup.tsx`

**Before**:
```typescript
const [loading, setLoading] = useState(false);
const result = await createNewAccount({...});
setLoading(false); // Manual state management
```

**After**:
```typescript
const { signup, isPending } = useSignupMutation();
const result = await signup({...});
// isPending automatically managed
```

**Benefits**:
- Removed manual loading state
- Cleaner disabled states on form fields
- Consistent mutation pattern

---

### 3. User Menu
**File**: `frontend/src/components/layout/header/user-menu.tsx`

**Before**:
```typescript
<Link href="/api/auth/signout">
  <LogOut />
  Log out
</Link>
```

**After**:
```typescript
const { logout, isPending } = useLogout();

<DropdownMenuItem onClick={() => logout()} disabled={isPending}>
  <LogOut />
  {isPending ? "Logging out..." : "Log out"}
</DropdownMenuItem>
```

**Benefits**:
- Proper mutation flow instead of simple link
- Loading state feedback
- Consistent with other logout implementations

---

### 4. Nav User (Sidebar)
**File**: `frontend/src/components/layout/sidebar/nav-user.tsx`

**Before**:
```typescript
<Link href="/api/auth/signout">
  <LogOutIcon />
  Log out
</Link>
```

**After**:
```typescript
const { logout, isPending } = useLogout();

<DropdownMenuItem onClick={() => logout()} disabled={isPending}>
  <LogOutIcon />
  {isPending ? "Logging out..." : "Log out"}
</DropdownMenuItem>
```

**Benefits**:
- Unified logout experience across UI
- Loading feedback during logout
- Proper React Query integration

---

### 5. Welcome Page
**File**: `frontend/src/app/welcome/page.tsx`

**Components**:
- Uses `useSelectOrganization()` for org switching
- Uses `useLogout()` for logout button
- Uses `useSession()` for user data (not useEffect)

**Benefits**:
- No useEffect data fetching
- Consistent mutation patterns
- Cleaner component logic

---

## 🎭 Hook Usage Patterns

### Pattern 1: Mutation with Success Navigation
```typescript
// Used by: Login, Signup, ResetPassword, SelectOrganization, Logout
const { mutationFn, isPending, error } = useMutationHook();

const handleAction = async () => {
  try {
    const result = await mutationFn(data);
    if (!result.success) {
      setError(result.message);
    }
  } catch (err) {
    setError(err.message);
  }
};
```

### Pattern 2: Mutation with Custom Response Handling
```typescript
// Used by: ChangePassword, SendResetEmail
const { mutationFn, isPending, error } = useMutationHook();

const handleAction = async () => {
  try {
    const result = await mutationFn(data);
    // Component handles response (no auto-navigation)
    if (result.success) {
      setSuccess(result.message);
      // Custom navigation or state updates
    }
  } catch (err) {
    setError(err.message);
  }
};
```

### Pattern 3: Query Hook (Not Yet Implemented)
```typescript
// Future pattern for data fetching
const { data, isLoading, error } = useQueryHook();

// Auto-fetches on mount, handles caching, refetching
```

---

## 🔄 Server Actions Integration

### Authentication Actions
**File**: `frontend/src/app/_actions/auth.ts`

Each server action is wrapped in a hook:

| Server Action | Hook | Auto-Navigation |
|---|---|---|
| `loginAction()` | `useLoginMutation()` | `/welcome` |
| `createNewAccount()` | `useSignupMutation()` | `/welcome` |
| `sendResetEmail()` | `useSendResetEmailMutation()` | None |
| `resetPassword()` | `useResetPasswordMutation()` | `/login` |
| `changePassword()` | `useChangePasswordMutation()` | None |
| `logoutAction()` | `useLogout()` | `/login` |

### Organization Actions
**File**: `frontend/src/app/_actions/organizations.ts`

| Server Action | Hook | Auto-Navigation |
|---|---|---|
| `switchOrganization()` | `useSelectOrganization()` | `/home` |

---

## 📊 State Management

### Before (Manual)
```typescript
const [loading, setLoading] = useState(false);
const [error, setError] = useState("");
const [data, setData] = useState(null);

// In handler:
setLoading(true);
try {
  const result = await serverAction();
  if (result.success) {
    setData(result.data);
  } else {
    setError(result.message);
  }
} finally {
  setLoading(false);
}
```

### After (React Query)
```typescript
const { mutationFn, isPending, error } = useMutationHook();
const [error, setError] = useState(""); // Only for display errors

// In handler:
try {
  const result = await mutationFn(data);
  if (!result.success) {
    setError(result.message);
  }
} catch (err) {
  setError(err.message);
}

// isPending is automatically managed by React Query
```

**Benefits**:
- Reduced boilerplate
- Automatic loading state
- Consistent error handling
- Automatic caching and refetching

---

## 🚀 Implementation Checklist

### ✅ Completed

- [x] `useLoginMutation()` - Login flow
- [x] `useSignupMutation()` - Registration flow
- [x] `useSendResetEmailMutation()` - Reset email
- [x] `useResetPasswordMutation()` - Password reset
- [x] `useChangePasswordMutation()` - Change password
- [x] `useSelectOrganization()` - Organization switching
- [x] `useLogout()` - Logout flow
- [x] Login form refactored
- [x] Signup form refactored
- [x] User menu refactored
- [x] Nav user refactored
- [x] Welcome page uses hooks

### ⏳ To Do (Future)

- [ ] `useDashboardMetricsQuery()` - Dashboard data
- [ ] `useTasksQuery()` - Tasks list
- [ ] `useDashboardStatsQuery()` - Stats
- [ ] Refactor approval flows to use mutations
- [ ] Refactor workflow queries
- [ ] Add error boundary for hook errors

---

## 📝 Writing a New Hook

### Template
```typescript
'use client';

import { useMutation } from '@tanstack/react-query';
import { useRouter } from 'next/navigation';
import { serverAction } from '@/app/_actions/...';

/**
 * Hook for handling [action description]
 * Manages [flow description] with [automatic behavior]
 *
 * @returns {Object} Object with mutation function, isPending, and error
 *
 * @example
 * ```typescript
 * const { actionFn, isPending, error } = useActionHook();
 * await actionFn({ param1, param2 });
 * ```
 */
export function useActionHook() {
  const router = useRouter(); // If navigation needed

  const mutation = useMutation({
    mutationFn: async (data: InputType) => {
      return await serverAction(data);
    },
    onSuccess: (data) => {
      if (data.success) {
        // Optional: Auto-navigate
        router.push('/destination');
      }
    },
    onError: (error) => {
      console.error('Action failed:', error);
    },
  });

  return {
    actionFn: mutation.mutateAsync,
    isPending: mutation.isPending,
    error: mutation.error,
  };
}
```

### Key Points
1. Always use `useMutation` for mutations
2. Always expose `isPending` (not `isLoading`)
3. Always expose `mutateAsync` (not `mutate`)
4. Document with JSDoc
5. Include usage example
6. Handle errors consistently

---

## 🎯 Benefits Summary

| Aspect | Before | After |
|---|---|---|
| Code Duplication | High (manual state in each component) | Low (centralized in hooks) |
| Loading State | Manual (setIsLoading) | Automatic (isPending) |
| Error Handling | Inconsistent | Consistent pattern |
| Navigation | Manual (router.push) | Automatic (onSuccess) |
| Caching | None | Built-in (React Query) |
| Refetching | Manual | Automatic with invalidation |
| Type Safety | Partial | Full (TypeScript generics) |
| Testing | Difficult | Easier (isolated hooks) |
| Reusability | Limited | High (use in multiple components) |

---

## 🔗 Related Documentation

- [React Query Documentation](https://tanstack.com/query/latest)
- [Server Actions Guide](./06-DEVELOPMENT-GUIDE.md#server-actions)
- [Authentication Architecture](./AUTH-PERMISSIONS-SYSTEM-DEEP-DIVE.md)
- [Component Structure](./05-CODE-STRUCTURE.md)

---

## 📞 Usage Examples

### Login Form Usage
```typescript
import { useLoginMutation } from '@/hooks/use-auth-mutations';

export function LoginForm() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const { login, isPending } = useLoginMutation();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    try {
      const result = await login({ email, password });
      if (!result.success) {
        setError(result.message || "Login failed");
      }
    } catch (err: any) {
      setError(err.message || "An error occurred");
    }
  };

  return (
    <form onSubmit={handleSubmit}>
      <Input
        value={email}
        onChange={(e) => setEmail(e.target.value)}
        disabled={isPending}
      />
      <Input
        type="password"
        value={password}
        onChange={(e) => setPassword(e.target.value)}
        disabled={isPending}
      />
      <Button type="submit" disabled={isPending} isLoading={isPending}>
        Sign In
      </Button>
    </form>
  );
}
```

### Welcome Page Organization Selection
```typescript
import { useSelectOrganization } from '@/hooks/use-organization-mutations';

export default function WelcomePage() {
  const { selectOrganization, isPending } = useSelectOrganization();
  const { userOrganizations } = useOrganizationContext();

  const handleSelectOrg = async (orgId: string) => {
    await selectOrganization(orgId);
    // Auto-navigates to /home on success
  };

  return (
    <div>
      {userOrganizations.map(org => (
        <button
          key={org.id}
          onClick={() => handleSelectOrg(org.id)}
          disabled={isPending}
        >
          {org.name}
        </button>
      ))}
    </div>
  );
}
```

---

## ✨ Next Steps

1. **Query Hooks**: Implement `useDashboardMetricsQuery()` and other data fetching hooks
2. **Approval Mutations**: Convert approval flows to use hooks
3. **Error Boundaries**: Add error boundaries around hook usage
4. **Testing**: Add unit tests for hooks
5. **Documentation**: Update component-specific documentation

---

## 🎓 Summary

The hooks implementation provides:
- ✅ Consistent mutation patterns across the app
- ✅ Automatic state management (loading, errors)
- ✅ Reusable authentication flows
- ✅ Built-in caching and refetching
- ✅ Better separation of concerns
- ✅ Cleaner component code
- ✅ Foundation for future query hooks

This becomes the **standard pattern going forward** for all data mutations and async operations.

---

**Maintained By**: Claude Code
**Status**: ✅ Complete and Production Ready
**Pattern**: React Query Mutations (following TanStack Query best practices)
