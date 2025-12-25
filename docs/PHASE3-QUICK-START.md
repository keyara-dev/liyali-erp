# Phase 3 Quick Start Guide

**Status**: ✅ Production Ready
**Commit**: `f639e03` feat: Phase 3 - Permission-Based Authorization System Implementation

---

## What is Phase 3?

Phase 3 implements **permission-based authorization** across your entire system:
- ✅ Every API endpoint checks permissions
- ✅ Frontend UI guards conditionally render based on permissions
- ✅ 5 hardcoded roles with defined permissions
- ✅ Ready to extend with Phase 3.5 custom roles

---

## Quick Overview

### Roles & Permissions

| Role | Can Do |
|------|--------|
| **admin** | Everything |
| **approver** | View, create, approve/reject documents |
| **requester** | Create and view their requisitions |
| **finance** | Manage budgets and payments, approve docs |
| **viewer** | Read-only access |

### How It Works

**Backend:**
```
User makes API request
    ↓
AuthMiddleware extracts role from JWT
    ↓
RequirePermission middleware checks role permissions
    ↓
PermissionService.HasPermission() returns true/false
    ↓
200 OK or 403 Forbidden
```

**Frontend:**
```
Component renders
    ↓
usePermissions() hook gets user role
    ↓
<PermissionGuard> checks permission
    ↓
Show or hide UI element
```

---

## Using Permissions in Your Code

### Backend: Protecting an Endpoint

All endpoints are already protected! Example from `routes.go`:

```go
// Only users with "requisition:approve" permission can access
requisitions.Post("/:id/approve",
  middleware.RequirePermission(config.DB, "requisition", "approve"),
  handlers.ApproveRequisition)
```

To add permissions to a new endpoint:

```go
// Single permission
myRouter.Post("/action",
  middleware.RequirePermission(config.DB, "resource", "action"),
  myHandler)

// Multiple permissions (all required)
myRouter.Post("/admin-action",
  middleware.RequirePermission(config.DB, "resource1", "action1", "resource2", "action2"),
  myHandler)

// Multiple permissions (any required)
myRouter.Post("/flexible-action",
  middleware.RequirePermissionOr(config.DB, "resource1", "action1", "resource2", "action2"),
  myHandler)
```

### Frontend: Conditional UI

#### Using Guard Components (Recommended)

```tsx
import { PermissionGuard, AdminGuard } from "@/components/auth/permission-guard";

export function RequisitionActions() {
  return (
    <div>
      {/* Show button only if user can approve */}
      <PermissionGuard resource="requisition" action="approve">
        <button onClick={handleApprove}>Approve</button>
      </PermissionGuard>

      {/* Show section only for admins */}
      <AdminGuard>
        <AdminPanel />
      </AdminGuard>

      {/* Show if user has ANY of these permissions */}
      <AnyPermissionGuard
        permissions={[
          { resource: "requisition", action: "approve" },
          { resource: "requisition", action: "reject" }
        ]}
      >
        <button>Take Action</button>
      </AnyPermissionGuard>
    </div>
  );
}
```

#### Using Hook (for logic)

```tsx
import { usePermissions } from "@/hooks/use-permissions";

export function Dashboard() {
  const {
    hasPermission,
    isAdmin,
    isApprover,
    userRole
  } = usePermissions();

  return (
    <div>
      {hasPermission("requisition", "create") && (
        <button>Create Requisition</button>
      )}

      {isAdmin() && (
        <button>Admin Settings</button>
      )}

      <p>Your role: {userRole}</p>
    </div>
  );
}
```

---

## Testing

### Test Backend Permissions

```bash
# Run unit tests
cd backend
go test ./services -v -run TestPermissionService

# Expected: 30+ test cases passing
```

### Test Endpoints with curl

```bash
# 1. Register and login
TOKEN=$(curl -s -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@test.com","password":"test123"}' \
  | jq -r '.data.token')

# 2. Try endpoint as admin (should work)
curl -X GET http://localhost:3000/api/v1/requisitions \
  -H "Authorization: Bearer $TOKEN"
# → 200 OK

# 3. Try as requester (should fail for approve)
REQUESTER_TOKEN=$(curl -s -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"requester@test.com","password":"test123"}' \
  | jq -r '.data.token')

curl -X POST http://localhost:3000/api/v1/requisitions/{id}/approve \
  -H "Authorization: Bearer $REQUESTER_TOKEN"
# → 403 Forbidden
```

### Test Frontend Components

```tsx
// In your component test file
import { render, screen } from "@testing-library/react";
import { PermissionGuard } from "@/components/auth/permission-guard";
import { usePermissions } from "@/hooks/use-permissions";

jest.mock("@/hooks/use-permissions");

test("shows button when user has permission", () => {
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
```

---

## Common Scenarios

### Scenario 1: Make a button visible only to admins

```tsx
<AdminGuard>
  <button onClick={deleteEverything}>Delete All</button>
</AdminGuard>
```

### Scenario 2: Hide/show menu items based on permissions

```tsx
<nav>
  <PermissionGuard resource="requisition" action="view">
    <a href="/requisitions">Requisitions</a>
  </PermissionGuard>

  <PermissionGuard resource="budget" action="view">
    <a href="/budgets">Budgets</a>
  </PermissionGuard>

  <AdminGuard>
    <a href="/admin">Admin Panel</a>
  </AdminGuard>
</nav>
```

### Scenario 3: Different form fields based on role

```tsx
<form>
  <input name="title" /> {/* Always shown */}

  <PermissionGuard resource="requisition" action="edit">
    <input name="department" /> {/* Only for editors */}
  </PermissionGuard>

  <AdminGuard>
    <select name="priority"> {/* Only for admins */}
      <option>Low</option>
      <option>High</option>
    </select>
  </AdminGuard>
</form>
```

### Scenario 4: Call API with permission check

```tsx
const { hasPermission } = usePermissions();

function handleApprove() {
  if (!hasPermission("requisition", "approve")) {
    alert("You don't have permission to approve");
    return;
  }

  // Make API call
  approveRequisition(id);
}
```

---

## Understanding the Permission Matrix

### Resources
- `requisition`, `budget`, `purchase_order`, `payment_voucher`, `grn`
- `vendor`, `category`, `organization`, `analytics`, `audit_log`

### Actions
- `view` - See the resource
- `create` - Create new resource
- `edit` - Modify existing resource
- `delete` - Remove resource
- `approve` - Approve for workflows
- `reject` - Reject workflows

### Full Format

Permission = `resource:action`

Examples:
- `requisition:approve` - Can approve requisitions
- `budget:edit` - Can edit budgets
- `organization:manage_users` - Can manage organization users

---

## Troubleshooting

### Issue: Button showing when it shouldn't

**Check:**
1. User's role is correctly set in database: `SELECT role FROM users WHERE id=?`
2. Role name matches exactly (case-insensitive): `admin`, `approver`, `requester`, `finance`, `viewer`
3. Permission check is correct: `"requisition"` not `"Requisition"`

### Issue: 403 Forbidden on API call

**Check:**
1. User role is in JWT token: Look at token in jwt.io
2. Organization context is set: Should be in request headers or middleware
3. Required permission exists in `RolePermissions` mapping

### Issue: Frontend guard not responding to changes

**Force refresh:**
```tsx
const queryClient = useQueryClient();
queryClient.invalidateQueries({ queryKey: ["session"] });
```

---

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────┐
│                    Frontend (React)                      │
├─────────────────────────────────────────────────────────┤
│  usePermissions() Hook                                   │
│  ├── hasPermission(resource, action)                    │
│  ├── isAdmin(), isApprover(), etc.                      │
│  └── userRole                                            │
├─────────────────────────────────────────────────────────┤
│  <PermissionGuard>, <AdminGuard>, etc.                   │
│  (Conditionally render UI)                              │
└─────────────────────────────────────────────────────────┘
                          ↓
            API Calls with Authorization Header
                          ↓
┌─────────────────────────────────────────────────────────┐
│                   Backend (Go/Fiber)                     │
├─────────────────────────────────────────────────────────┤
│  AuthMiddleware                                          │
│  (Extract role from JWT)                                │
├─────────────────────────────────────────────────────────┤
│  RequirePermission(resource, action)                     │
│  (Check if user has permission)                         │
├─────────────────────────────────────────────────────────┤
│  PermissionService.HasPermission()                       │
│  (Look up in RolePermissions map)                        │
├─────────────────────────────────────────────────────────┤
│  Handler (Only if permission granted)                    │
│  or 403 Forbidden (If denied)                            │
└─────────────────────────────────────────────────────────┘
```

---

## Files Reference

### Backend
- `backend/services/permission_service.go` - Permission checking logic
- `backend/middleware/middleware.go` - RequirePermission middleware
- `backend/routes/routes.go` - Permission checks on all endpoints

### Frontend
- `frontend/src/hooks/use-permissions.ts` - usePermissions hook
- `frontend/src/components/auth/permission-guard.tsx` - Guard components

### Documentation
- `docs/PHASE3-IMPLEMENTATION-COMPLETE.md` - Full technical details
- `docs/PHASE3-BACKEND-TESTING-GUIDE.md` - Testing guide with examples
- `docs/PHASE3-FRONTEND-IMPLEMENTATION-GUIDE.md` - Frontend integration guide

---

## What's Next?

### Option A: Deploy Phase 3 Now
Phase 3 is production-ready. All endpoints are protected. Ready to deploy!

### Option B: Extend to Phase 3.5
Phase 3.5 will allow organization admins to:
- Create custom roles
- Assign permissions to roles
- Define workflows with role requirements

Database models are already in place. Just need to add:
1. API endpoints for role management
2. Frontend UI for role configuration

See `backend/services/role_management_service.go` for the ready-to-use service.

---

## Need Help?

- **How do I add a permission to a role?**
  See `RolePermissions` map in `permission_service.go`

- **How do I protect a new endpoint?**
  Add `middleware.RequirePermission(config.DB, "resource", "action")` to route

- **How do I hide a button from non-admins?**
  Wrap with `<AdminGuard>` component

- **How do I implement Phase 3.5?**
  Use `RoleManagementService` in `role_management_service.go`

---

## Summary

✅ Phase 3 provides complete, working permission-based authorization
✅ Every endpoint protected
✅ Frontend UI responds to permissions
✅ No breaking changes
✅ Ready for production
✅ Easy to extend with Phase 3.5

**Start using it today!** 🚀
