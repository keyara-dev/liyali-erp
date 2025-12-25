# Phase 3B Frontend Implementation Guide

## Overview

Phase 3B implements permission-based access control on the frontend through:

1. **usePermissions Hook** - Check user permissions in any component
2. **Permission Guard Components** - Conditionally render UI based on permissions
3. **Integration with Existing RBAC** - Works alongside existing role-based checks
4. **Fallback Handling** - Graceful degradation for permission-denied scenarios

---

## New Frontend Components and Hooks

### 1. usePermissions Hook

Location: `frontend/src/hooks/use-permissions.ts`

The hook provides comprehensive permission checking capabilities.

#### Basic Usage

```typescript
import { usePermissions } from "@/hooks/use-permissions";

export function RequisitionActions() {
  const { hasPermission, isAdmin, userRole } = usePermissions();

  return (
    <div>
      {/* Show approval button only if user can approve */}
      {hasPermission("requisition", "approve") && (
        <button onClick={handleApprove}>Approve</button>
      )}

      {/* Show admin panel only for admins */}
      {isAdmin() && (
        <AdminPanel />
      )}

      {/* Current role display */}
      <p>Your role: {userRole}</p>
    </div>
  );
}
```

#### Available Methods

```typescript
const {
  // Check single permission
  hasPermission(resource: string, action: string): boolean

  // Check all permissions required
  hasAllPermissions(permissions: PermissionCheck[]): boolean

  // Check any permission from list
  hasAnyPermission(permissions: PermissionCheck[]): boolean

  // Get all user permissions
  getPermissions(): PermissionCheck[]

  // Role-specific checks
  isAdmin(): boolean
  isApprover(): boolean
  isRequester(): boolean
  isFinance(): boolean

  // Current state
  userRole: string | null
  isLoading: boolean
  error: Error | null
} = usePermissions();
```

### 2. Permission Guard Components

Location: `frontend/src/components/auth/permission-guard.tsx`

Guard components provide declarative permission checking.

#### PermissionGuard

Check a single permission:

```tsx
import { PermissionGuard } from "@/components/auth/permission-guard";

export function RequisitionCard({ requisition }) {
  return (
    <div>
      <h3>{requisition.title}</h3>

      {/* Only show approve button if user can approve */}
      <PermissionGuard resource="requisition" action="approve">
        <button onClick={() => approveRequisition(requisition.id)}>
          Approve
        </button>
      </PermissionGuard>

      {/* With fallback message */}
      <PermissionGuard
        resource="requisition"
        action="delete"
        fallback={<p className="text-muted">You cannot delete this</p>}
      >
        <button onClick={() => deleteRequisition(requisition.id)}>
          Delete
        </button>
      </PermissionGuard>
    </div>
  );
}
```

#### MultiPermissionGuard

Check that user has ALL permissions:

```tsx
import { MultiPermissionGuard } from "@/components/auth/permission-guard";

export function AdvancedRequisitionActions() {
  return (
    <MultiPermissionGuard
      permissions={[
        { resource: "requisition", action: "create" },
        { resource: "category", action: "view" },
        { resource: "vendor", action: "view" }
      ]}
      fallback={<p>Insufficient permissions for advanced actions</p>}
    >
      <CreateAdvancedRequisitionForm />
    </MultiPermissionGuard>
  );
}
```

#### AnyPermissionGuard

Check that user has AT LEAST ONE permission:

```tsx
import { AnyPermissionGuard } from "@/components/auth/permission-guard";

export function DocumentActions() {
  return (
    <AnyPermissionGuard
      permissions={[
        { resource: "requisition", action: "approve" },
        { resource: "requisition", action: "reject" }
      ]}
      fallback={<p>You cannot take action on this document</p>}
    >
      <ActionButtons />
    </AnyPermissionGuard>
  );
}
```

#### RoleGuard

Check user role directly:

```tsx
import { RoleGuard } from "@/components/auth/permission-guard";

export function Dashboard() {
  return (
    <div>
      <h1>Dashboard</h1>

      <RoleGuard role="admin">
        <AdminDashboard />
      </RoleGuard>

      <RoleGuard role="approver">
        <ApprovalQueue />
      </RoleGuard>

      <RoleGuard role="requester">
        <RequisitionForm />
      </RoleGuard>
    </div>
  );
}
```

#### AdminGuard

Convenience guard for admin-only content:

```tsx
import { AdminGuard } from "@/components/auth/permission-guard";

export function Settings() {
  return (
    <AdminGuard
      fallback={<p>Admin access required</p>}
    >
      <AdminSettings />
    </AdminGuard>
  );
}
```

---

## Integration Examples

### Example 1: Requisition List with Conditional Actions

```tsx
"use client";

import { usePermissions } from "@/hooks/use-permissions";
import { PermissionGuard } from "@/components/auth/permission-guard";
import { Table } from "@/components/ui/table";

export function RequisitionsList({ requisitions }) {
  const { hasPermission } = usePermissions();

  return (
    <Table>
      <thead>
        <tr>
          <th>Title</th>
          <th>Amount</th>
          <th>Status</th>
          {(hasPermission("requisition", "approve") ||
            hasPermission("requisition", "reject") ||
            hasPermission("requisition", "delete")) && <th>Actions</th>}
        </tr>
      </thead>
      <tbody>
        {requisitions.map((req) => (
          <tr key={req.id}>
            <td>{req.title}</td>
            <td>${req.totalAmount}</td>
            <td>{req.status}</td>
            <td>
              {/* Approve button - only for users with approve permission */}
              <PermissionGuard
                resource="requisition"
                action="approve"
                fallback={null}
              >
                <button
                  onClick={() => handleApprove(req.id)}
                  className="btn-sm btn-success"
                >
                  Approve
                </button>
              </PermissionGuard>

              {/* Reject button - only for users with reject permission */}
              <PermissionGuard
                resource="requisition"
                action="reject"
                fallback={null}
              >
                <button
                  onClick={() => handleReject(req.id)}
                  className="btn-sm btn-danger"
                >
                  Reject
                </button>
              </PermissionGuard>

              {/* Delete button - only for admins */}
              <PermissionGuard
                resource="requisition"
                action="delete"
                fallback={null}
              >
                <button
                  onClick={() => handleDelete(req.id)}
                  className="btn-sm btn-outline-danger"
                >
                  Delete
                </button>
              </PermissionGuard>
            </td>
          </tr>
        ))}
      </tbody>
    </Table>
  );
}
```

### Example 2: Conditional Navigation Menu

```tsx
"use client";

import { usePermissions, RoleGuard, AdminGuard, PermissionGuard } from "@/hooks/use-permissions";
import Link from "next/link";

export function Navigation() {
  const { hasPermission, userRole } = usePermissions();

  return (
    <nav>
      <ul>
        {/* Home - always visible */}
        <li>
          <Link href="/">Home</Link>
        </li>

        {/* Requisitions - visible if user can view */}
        <PermissionGuard resource="requisition" action="view">
          <li>
            <Link href="/requisitions">Requisitions</Link>
          </li>
        </PermissionGuard>

        {/* Budgets - visible if user can view */}
        <PermissionGuard resource="budget" action="view">
          <li>
            <Link href="/budgets">Budgets</Link>
          </li>
        </PermissionGuard>

        {/* Admin Panel - admin only */}
        <AdminGuard>
          <li>
            <Link href="/admin">Admin Panel</Link>
          </li>
        </AdminGuard>

        {/* Approvals - visible if user can approve */}
        <PermissionGuard
          resource="requisition"
          action="approve"
        >
          <li>
            <Link href="/approvals">My Approvals</Link>
          </li>
        </PermissionGuard>
      </ul>
    </nav>
  );
}
```

### Example 3: Form with Permission-Based Fields

```tsx
"use client";

import { usePermissions } from "@/hooks/use-permissions";
import { PermissionGuard, MultiPermissionGuard } from "@/components/auth/permission-guard";

export function RequisitionForm() {
  const { hasPermission } = usePermissions();

  return (
    <form>
      {/* Title - requester can enter */}
      <PermissionGuard resource="requisition" action="create">
        <div>
          <label>Title</label>
          <input type="text" name="title" />
        </div>
      </PermissionGuard>

      {/* Amount - only visible to finance and admins */}
      <MultiPermissionGuard
        permissions={[
          { resource: "budget", action: "view" },
          { resource: "requisition", action: "create" }
        ]}
      >
        <div>
          <label>Budget Amount</label>
          <input type="number" name="amount" />
        </div>
      </MultiPermissionGuard>

      {/* Priority - admin only */}
      <PermissionGuard resource="requisition" action="edit">
        <div>
          <label>Priority Level</label>
          <select name="priority">
            <option>Low</option>
            <option>Medium</option>
            <option>High</option>
          </select>
        </div>
      </PermissionGuard>

      <button type="submit">Submit</button>
    </form>
  );
}
```

---

## Migration Guide: From RBAC to Permissions

The existing RBAC system (in `lib/rbac.ts`) continues to work. The new permission system complements it:

### Before (RBAC only)
```tsx
import { hasPermission } from "@/lib/rbac";
import type { UserRole } from "@/types/workflow";

export function RequisitionActions({ userRole }: { userRole: UserRole }) {
  if (hasPermission(userRole, "approve_document")) {
    return <ApproveButton />;
  }
  return null;
}
```

### After (With Permissions)
```tsx
import { PermissionGuard } from "@/components/auth/permission-guard";

export function RequisitionActions() {
  return (
    <PermissionGuard resource="requisition" action="approve">
      <ApproveButton />
    </PermissionGuard>
  );
}
```

### Both (During Migration)
You can use both systems together:

```tsx
import { usePermissions } from "@/hooks/use-permissions";
import { hasPermission } from "@/lib/rbac";

export function Component() {
  const { hasPermission: hasNewPermission, userRole } = usePermissions();
  const hasOldPermission = hasPermission(userRole as UserRole, "approve_document");

  // Use one or both
  return (
    <>
      {hasOldPermission && <OldStyleComponent />}
      {hasNewPermission("requisition", "approve") && <NewStyleComponent />}
    </>
  );
}
```

---

## Testing Permission Components

### Unit Tests

```typescript
import { render, screen } from "@testing-library/react";
import { PermissionGuard } from "@/components/auth/permission-guard";
import { usePermissions } from "@/hooks/use-permissions";

// Mock the hook
jest.mock("@/hooks/use-permissions");

describe("PermissionGuard", () => {
  it("shows content when user has permission", () => {
    (usePermissions as jest.Mock).mockReturnValue({
      hasPermission: () => true,
      isLoading: false,
    });

    render(
      <PermissionGuard resource="requisition" action="approve">
        <button>Approve</button>
      </PermissionGuard>
    );

    expect(screen.getByText("Approve")).toBeInTheDocument();
  });

  it("hides content when user lacks permission", () => {
    (usePermissions as jest.Mock).mockReturnValue({
      hasPermission: () => false,
      isLoading: false,
    });

    render(
      <PermissionGuard resource="requisition" action="approve">
        <button>Approve</button>
      </PermissionGuard>
    );

    expect(screen.queryByText("Approve")).not.toBeInTheDocument();
  });

  it("shows fallback when permission denied", () => {
    (usePermissions as jest.Mock).mockReturnValue({
      hasPermission: () => false,
      isLoading: false,
    });

    render(
      <PermissionGuard
        resource="requisition"
        action="approve"
        fallback={<p>No permission</p>}
      >
        <button>Approve</button>
      </PermissionGuard>
    );

    expect(screen.getByText("No permission")).toBeInTheDocument();
  });
});
```

---

## Permission Matrix for Frontend

### Requisition Resource
- `view` - Can see requisitions
- `create` - Can create new requisitions
- `edit` - Can edit requisitions
- `delete` - Can delete requisitions
- `approve` - Can approve requisitions
- `reject` - Can reject requisitions

### Budget Resource
- `view` - Can see budgets
- `create` - Can create budgets
- `edit` - Can edit budgets
- `delete` - Can delete budgets
- `approve` - Can approve budgets
- `reject` - Can reject budgets

### Purchase Order, Payment Voucher, GRN Resources
Similar permissions: view, create, edit, delete, approve, reject

### Organization Resource
- `view` - Can view organization
- `edit` - Can edit settings
- `manage_users` - Can manage users
- `manage_workflows` - Can manage workflows

### Analytics & Audit Resources
- `view` - Can view dashboards/logs

---

## Common Patterns

### Pattern 1: Action Button with Permission Check
```tsx
<PermissionGuard
  resource="requisition"
  action="approve"
  fallback={
    <button disabled title="You don't have permission to approve">
      Approve
    </button>
  }
>
  <button onClick={handleApprove}>Approve</button>
</PermissionGuard>
```

### Pattern 2: Conditional Field Visibility
```tsx
<PermissionGuard resource="requisition" action="edit">
  <FormField name="department" label="Department" />
</PermissionGuard>
```

### Pattern 3: Role-Based Sections
```tsx
<div className="grid grid-cols-1 md:grid-cols-2 gap-4">
  <RoleGuard role="approver">
    <ApprovalQueue />
  </RoleGuard>

  <RoleGuard role="finance">
    <FinanceMetrics />
  </RoleGuard>

  <RoleGuard role="admin">
    <SystemSettings />
  </RoleGuard>
</div>
```

### Pattern 4: Loading State
```tsx
<PermissionGuard
  resource="requisition"
  action="approve"
  loadingFallback={<Spinner />}
>
  <ApproveButton />
</PermissionGuard>
```

---

## Troubleshooting

### Issue: Permission checks always return false

**Solution**: Ensure `useSession()` is working and returning user data with role:

```typescript
const { user, isLoading, error } = useSession();
console.log("User:", user);
console.log("Role:", user?.role);
```

### Issue: Components showing when they shouldn't

**Solution**: Check that:
1. User's role is set correctly in the database
2. Role name matches exactly (case-insensitive, but consistent)
3. Backend and frontend permission definitions match

### Issue: Permission guards not responding to auth changes

**Solution**: The hook uses React Query caching. You can invalidate manually:

```typescript
const queryClient = useQueryClient();

function handleLogout() {
  queryClient.invalidateQueries({ queryKey: ["session"] });
  // Now usePermissions will re-evaluate
}
```

---

## Best Practices

1. **Always check permissions before API calls** - Both frontend and backend check permissions
2. **Use fallback UI** - Provide graceful degradation for permission-denied cases
3. **Loading states** - Show loading indicator while permissions are being fetched
4. **Group related permissions** - Use `MultiPermissionGuard` for related checks
5. **Don't hardcode role names** - Use the constants from the hook
6. **Test permission scenarios** - Include tests for both permitted and denied cases

