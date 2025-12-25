# Phase 3.5 Custom Role Management - Usage Guide

**Status**: ✅ **PHASE 3.5 IMPLEMENTATION COMPLETE AND READY TO USE**

**Date**: 2025-12-25

---

## Table of Contents

1. [Overview](#overview)
2. [System Architecture](#system-architecture)
3. [Backend API Reference](#backend-api-reference)
4. [Frontend Usage](#frontend-usage)
5. [Role Management Workflows](#role-management-workflows)
6. [Permission Management](#permission-management)
7. [Testing Guide](#testing-guide)
8. [Troubleshooting](#troubleshooting)
9. [Migration from Phase 3](#migration-from-phase-3)
10. [Best Practices](#best-practices)

---

## Overview

Phase 3.5 enables organization administrators to create custom roles and assign permissions to them. This allows fine-grained access control tailored to each organization's specific needs.

### Key Features

- ✅ Create custom roles per organization
- ✅ Assign/revoke permissions to/from roles
- ✅ System default roles protected from deletion
- ✅ Database-driven permissions (fallback to Phase 3 hardcoded roles)
- ✅ Real-time permission management UI
- ✅ Multi-tenancy support

### System Default Roles

These roles are built-in and cannot be deleted:

| Role | Description | Default Permissions |
|------|-------------|-------------------|
| **admin** | Full system access | All permissions |
| **approver** | Can approve workflows | Approve, view, edit requisitions |
| **requester** | Can create requisitions | View, create, edit requisitions |
| **finance** | Manages budgets and payments | All budget, payment, approve permissions |
| **viewer** | Read-only access | View all resources |

---

## System Architecture

### Backend Flow

```
Request
  ↓
Authentication Middleware (verify JWT token)
  ↓
Tenant Middleware (extract organization ID)
  ↓
Permission Middleware (check authorization)
  ↓
Role Management Handler
  ↓
Service Layer (RoleManagementService)
  ↓
Database (OrganizationRole, OrganizationPermission, PermissionAssignment)
```

### Permission Resolution

When a user makes an API request:

1. **Extract** user role and organization from JWT token
2. **Check Database** for organization-specific custom roles/permissions
3. **Fall Back** to Phase 3 hardcoded permission mapping if not found
4. **Authorize** based on resolved permissions
5. **Return** 200 OK or 403 Forbidden

### Data Models

```go
// Organization custom role
OrganizationRole {
  ID: string              // UUID
  OrganizationID: string  // Which organization owns this role
  Name: string            // e.g., "Approver", "Manager"
  Description: string     // What this role does
  IsDefault: bool         // true = system role, cannot delete
  IsActive: bool          // Can be deactivated without deleting
  CreatedAt: timestamp
  UpdatedAt: timestamp
}

// Available permission
OrganizationPermission {
  ID: string              // UUID
  OrganizationID: string  // Which organization owns this permission
  Resource: string        // e.g., "requisition", "budget"
  Action: string          // e.g., "approve", "create"
  Description: string     // What this permission allows
  IsActive: bool          // Can be deactivated
  CreatedAt: timestamp
  UpdatedAt: timestamp
}

// Role-to-permission mapping
PermissionAssignment {
  ID: string                      // UUID
  OrganizationRoleID: string      // Which role
  OrganizationPermissionID: string // Which permission
  CreatedAt: timestamp
}
```

---

## Backend API Reference

### List Organization Roles

```http
GET /api/v1/organization/roles
Authorization: Bearer {token}
X-Organization-ID: {orgId}

Response 200:
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "organizationId": "uuid",
      "name": "Manager",
      "description": "Manages team",
      "isDefault": false,
      "isActive": true,
      "createdAt": "2025-12-25T10:00:00Z"
    }
  ],
  "message": "Roles retrieved successfully"
}
```

### Create Role

```http
POST /api/v1/organization/roles
Authorization: Bearer {token}
Content-Type: application/json
X-Organization-ID: {orgId}

Request:
{
  "name": "Department Head",
  "description": "Manages department operations"
}

Response 201:
{
  "success": true,
  "data": {
    "id": "new-uuid",
    "organizationId": "uuid",
    "name": "Department Head",
    "description": "Manages department operations",
    "isDefault": false,
    "isActive": true,
    "createdAt": "2025-12-25T10:00:00Z"
  },
  "message": "Role created successfully"
}
```

### Update Role

```http
PUT /api/v1/organization/roles/{roleId}
Authorization: Bearer {token}
Content-Type: application/json
X-Organization-ID: {orgId}

Request:
{
  "name": "Senior Department Head",
  "description": "Updated description"
}

Response 200:
{
  "success": true,
  "data": { /* updated role */ },
  "message": "Role updated successfully"
}
```

### Delete Role

```http
DELETE /api/v1/organization/roles/{roleId}
Authorization: Bearer {token}
X-Organization-ID: {orgId}

Response 200:
{
  "success": true,
  "message": "Role deleted successfully"
}

Error 400:
{
  "success": false,
  "error": "cannot delete system default roles (admin, approver, requester, finance, viewer)"
}
```

### List Available Permissions

```http
GET /api/v1/organization/permissions
Authorization: Bearer {token}
X-Organization-ID: {orgId}

Response 200:
{
  "success": true,
  "data": [
    {
      "id": "perm-uuid",
      "organizationId": "org-uuid",
      "resource": "requisition",
      "action": "approve",
      "description": "Approve requisitions",
      "isActive": true
    },
    // ... more permissions
  ],
  "message": "Permissions retrieved successfully"
}
```

### Get Role Permissions

```http
GET /api/v1/organization/roles/{roleId}/permissions
Authorization: Bearer {token}
X-Organization-ID: {orgId}

Response 200:
{
  "success": true,
  "data": [
    {
      "id": "perm-uuid",
      "organizationId": "org-uuid",
      "resource": "requisition",
      "action": "approve",
      "description": "Approve requisitions",
      "isActive": true
    }
  ],
  "message": "Role permissions retrieved successfully"
}
```

### Assign Permission to Role

```http
POST /api/v1/organization/roles/{roleId}/permissions/{permissionId}
Authorization: Bearer {token}
X-Organization-ID: {orgId}

Response 201:
{
  "success": true,
  "data": {
    "id": "assignment-uuid",
    "organizationRoleId": "role-uuid",
    "organizationPermissionId": "perm-uuid",
    "createdAt": "2025-12-25T10:00:00Z"
  },
  "message": "Permission assigned successfully"
}
```

### Remove Permission from Role

```http
DELETE /api/v1/organization/roles/{roleId}/permissions/{permissionId}
Authorization: Bearer {token}
X-Organization-ID: {orgId}

Response 200:
{
  "success": true,
  "message": "Permission removed successfully"
}
```

---

## Frontend Usage

### 1. Server Actions (TSX)

The frontend uses server actions to communicate with the backend:

```typescript
// src/app/_actions/roles.ts

// Get all roles
const rolesResponse = await getRolesAction();
const roles = rolesResponse.data; // Array of roles

// Create role
const newRole = await createRoleAction("Manager", "Manages operations");

// Update role
const updated = await updateRoleAction(roleId, "Senior Manager", "Updated desc");

// Delete role
await deleteRoleAction(roleId);

// Get available permissions
const permsResponse = await getAvailablePermissionsAction();
const permissions = permsResponse.data;

// Get role's current permissions
const rolePerms = await getRolePermissionsAction(roleId);

// Assign permission
await assignPermissionAction(roleId, permissionId);

// Remove permission
await removePermissionAction(roleId, permissionId);
```

### 2. Role Management Page

Access the role management interface at `/admin/roles`:

```
┌─────────────────────────────────────┐
│   Organization Roles                │
│   Create and manage custom roles    │
│                    [Create Role]    │
├─────────────────────────────────────┤
│ Name  │ Description │ Status│Actions│
├───────┼─────────────┼───────┼───────┤
│Admin  │ Full access │ Active│ P  -  │
│Manager│ Dept mgmt   │ Active│ P E D │
│Viewer │ Read-only   │ Active│ P  -  │
└─────────────────────────────────────┘

P = Manage Permissions
E = Edit (not available for system roles)
D = Delete (not available for system roles)
```

### 3. Role Modal

Click "Create Role" or "Edit" to open the role modal:

```
┌──────────────────────────┐
│ Create Role              │
├──────────────────────────┤
│ Role Name *              │
│ [____________________]   │
│                          │
│ Description *            │
│ [____________________]   │
│ [____________________]   │
│ [____________________]   │
│                          │
│        [Cancel] [Create] │
└──────────────────────────┘
```

**Validation Rules:**
- Role name: minimum 3 characters
- Description: minimum 10 characters

### 4. Permissions Modal

Click "Permissions" for a role to manage its permissions:

```
┌────────────────────────────────────────┐
│ Permissions for Manager                │
│ Select which permissions this role has │
├────────────────────────────────────────┤
│ [Search permissions...]                │
├────────────────────────────────────────┤
│ REQUISITION                            │
│ ☑ View requisitions                    │
│ ☑ Create requisitions                  │
│ ☑ Edit requisitions                    │
│ ☐ Delete requisitions                  │
│ ☑ Approve requisitions                 │
│ ☐ Reject requisitions                  │
│                                        │
│ BUDGET                                 │
│ ☑ View budgets                         │
│ ☐ Create budgets                       │
│                                        │
│                        [Done]          │
└────────────────────────────────────────┘
```

**Features:**
- Search/filter by resource or action
- Real-time checkbox updates
- Grouped by resource
- Auto-save on toggle

---

## Role Management Workflows

### Workflow 1: Create Custom Role with Permissions

```typescript
// Step 1: Create the role
const role = await createRoleAction("Department Manager", "Manages departmental operations");
const roleId = role.data.id;

// Step 2: Get available permissions
const perms = await getAvailablePermissionsAction();

// Step 3: Assign selected permissions
const requisitionApprove = perms.data.find(p => p.resource === "requisition" && p.action === "approve");
const budgetView = perms.data.find(p => p.resource === "budget" && p.action === "view");

await assignPermissionAction(roleId, requisitionApprove.id);
await assignPermissionAction(roleId, budgetView.id);

// Step 4: Verify permissions
const rolePerms = await getRolePermissionsAction(roleId);
console.log("Role has", rolePerms.data.length, "permissions");
```

### Workflow 2: Modify Existing Role

```typescript
// Get the role
const rolesResponse = await getRolesAction();
const role = rolesResponse.data.find(r => r.id === roleId);

// Update basic info
const updated = await updateRoleAction(roleId, "New Name", "New description");

// Get current permissions
const currentPerms = await getRolePermissionsAction(roleId);

// Add new permission
const newPerm = perms.data.find(p => p.resource === "vendor" && p.action === "create");
await assignPermissionAction(roleId, newPerm.id);

// Remove permission
const approveReq = currentPerms.data.find(p => p.action === "approve");
await removePermissionAction(roleId, approveReq.id);
```

### Workflow 3: Assign User to Custom Role

```typescript
// Note: User assignment to roles happens in organization member management
// This is a Phase 3.5 extension that references custom roles

// When updating an organization member:
const response = await updateOrganizationMember(memberId, {
  role: roleId, // Can now reference custom role UUID instead of just "approver"
});
```

---

## Permission Management

### Permission Naming Convention

Permissions follow a `resource:action` naming pattern:

| Resource | Actions | Example |
|----------|---------|---------|
| requisition | view, create, edit, delete, approve, reject | requisition:approve |
| budget | view, create, edit, delete, approve, reject | budget:view |
| purchase_order | view, create, edit, delete, approve, reject | purchase_order:create |
| payment_voucher | view, create, edit, delete, approve, reject | payment_voucher:approve |
| grn | view, create, edit, delete | grn:view |
| vendor | view, create, edit, delete | vendor:create |
| category | view, create, edit, delete | category:edit |
| organization | view, edit, manage_users, manage_workflows | organization:manage_workflows |
| analytics | view | analytics:view |
| audit_log | view | audit_log:view |

### Available Permissions

The system automatically creates these default permissions when an organization is initialized:

```
REQUISITION (6 permissions)
  - requisition:view
  - requisition:create
  - requisition:edit
  - requisition:delete
  - requisition:approve
  - requisition:reject

BUDGET (6 permissions)
  - budget:view
  - budget:create
  - budget:edit
  - budget:delete
  - budget:approve
  - budget:reject

PURCHASE_ORDER (6 permissions)
  [Similar structure]

PAYMENT_VOUCHER (6 permissions)
  [Similar structure]

GRN (4 permissions)
  - grn:view
  - grn:create
  - grn:edit
  - grn:delete

VENDOR (4 permissions)
  [Similar structure]

CATEGORY (4 permissions)
  [Similar structure]

ORGANIZATION (4 permissions)
  - organization:view
  - organization:edit
  - organization:manage_users
  - organization:manage_workflows

ANALYTICS (1 permission)
  - analytics:view

AUDIT_LOG (1 permission)
  - audit_log:view
```

---

## Testing Guide

### Unit Tests

Unit tests for role management service:

```bash
cd backend
go test ./services -v -run TestRoleManagement
```

**Test Coverage:**
- ✅ CreateOrganizationRole
- ✅ UpdateOrganizationRole
- ✅ DeleteOrganizationRole
- ✅ GetOrganizationRole
- ✅ GetOrganizationRoles
- ✅ AssignPermissionToRole
- ✅ RemovePermissionFromRole
- ✅ GetRolePermissions
- ✅ CreateOrganizationPermission
- ✅ GetOrganizationPermissions
- ✅ System default role protection
- ✅ Role management workflow

### Integration Tests

Test the API endpoints:

```bash
go test ./handlers -v -run TestRoles
```

**Test Cases:**
- List roles
- Create role (valid & invalid)
- Update role
- Delete role (custom & protected)
- Get permissions
- Assign permission
- Remove permission

### Manual Testing with cURL

```bash
# Set variables
TOKEN="your-jwt-token"
ORG_ID="your-org-id"
BASE_URL="http://localhost:3000/api/v1"

# 1. List roles
curl -X GET "$BASE_URL/organization/roles" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Organization-ID: $ORG_ID"

# 2. Create role
curl -X POST "$BASE_URL/organization/roles" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Organization-ID: $ORG_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Department Head",
    "description": "Manages department operations"
  }'

# 3. Get available permissions
curl -X GET "$BASE_URL/organization/permissions" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Organization-ID: $ORG_ID"

# 4. Assign permission to role
ROLE_ID="from-create-response"
PERM_ID="from-permissions-response"

curl -X POST "$BASE_URL/organization/roles/$ROLE_ID/permissions/$PERM_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Organization-ID: $ORG_ID"

# 5. Get role permissions
curl -X GET "$BASE_URL/organization/roles/$ROLE_ID/permissions" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Organization-ID: $ORG_ID"

# 6. Remove permission
curl -X DELETE "$BASE_URL/organization/roles/$ROLE_ID/permissions/$PERM_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Organization-ID: $ORG_ID"

# 7. Update role
curl -X PUT "$BASE_URL/organization/roles/$ROLE_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Organization-ID: $ORG_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Senior Department Head",
    "description": "Updated description"
  }'

# 8. Delete role
curl -X DELETE "$BASE_URL/organization/roles/$ROLE_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Organization-ID: $ORG_ID"
```

### Frontend Testing

Test the role management UI:

```typescript
// Test creating a role
1. Navigate to /admin/roles
2. Click "Create Role"
3. Enter name: "Test Manager" (min 3 chars)
4. Enter description: "Test description" (min 10 chars)
5. Click "Create"
6. Verify role appears in list

// Test managing permissions
1. Click "Permissions" for the created role
2. Search for "requisition"
3. Check "requisition:approve"
4. Verify permission is assigned (checkbox remains checked)
5. Uncheck permission
6. Verify permission is removed

// Test editing role
1. Click "Edit" for custom role
2. Change name and description
3. Click "Update"
4. Verify changes reflected in list

// Test deleting role
1. Click "Delete" for custom role
2. Confirm deletion
3. Verify role removed from list

// Test protected default roles
1. Observe "admin", "approver", etc. have no Edit/Delete buttons
2. Verify "Permissions" button is available
3. Attempt API delete on default role (should fail)
```

---

## Troubleshooting

### Problem: Cannot delete a custom role

**Cause:** Role name matches a system default (admin, approver, requester, finance, viewer)

**Solution:**
- Delete the role through the API to see exact error message
- Check if role name is one of the system defaults
- Rename the role before creating to avoid conflicts

### Problem: Permission not appearing for role after assignment

**Cause:** Permission might be inactive or assignment failed silently

**Solution:**
```bash
# Check if permission exists and is active
curl -X GET "$BASE_URL/organization/permissions" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Organization-ID: $ORG_ID" | jq '.data[] | select(.resource=="requisition" and .action=="approve")'

# Check if assignment exists
curl -X GET "$BASE_URL/organization/roles/$ROLE_ID/permissions" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Organization-ID: $ORG_ID"

# Try reassigning
curl -X POST "$BASE_URL/organization/roles/$ROLE_ID/permissions/$PERM_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Organization-ID: $ORG_ID"
```

### Problem: User cannot perform action despite having permission

**Cause:**
1. User role references old role ID
2. Role doesn't have permission assigned
3. Permission is inactive

**Solution:**
1. Verify user's role ID: Check organization member record
2. Verify role has permission: Use "Get role permissions" API
3. Check permission is active: Use "Get permissions" API
4. Verify user's organization matches permission's organization

### Problem: "Cannot modify default system roles" error

**Cause:** Attempting to update a system default role

**Solution:**
- System default roles (admin, approver, requester, finance, viewer) cannot be modified
- Create a new custom role instead if you need different configuration
- Use the custom role for organization-specific requirements

---

## Migration from Phase 3

Phase 3 used hardcoded role-to-permission mappings. Phase 3.5 adds database-driven custom roles.

### Backward Compatibility

✅ **Phase 3 still works!**

The permission resolution process:
1. First checks database for custom roles in the organization
2. If custom role not found, falls back to Phase 3 hardcoded mappings
3. Users can continue using system roles without custom role setup

### Migration Steps

**Option A: Keep Phase 3 Hardcoded Roles**
- No action needed
- System continues working as before
- All 5 system roles (admin, approver, requester, finance, viewer) always available

**Option B: Migrate to Phase 3.5 Custom Roles**

```typescript
// 1. Keep system roles, add custom roles
await createRoleAction("Department Manager", "Custom role for departments");

// 2. Assign permissions as needed
const perms = await getAvailablePermissionsAction();
// ... assign permissions ...

// 3. Update users to reference custom roles
// (In organization member management)
```

**Option C: Hybrid Approach**
- Use system roles for most users
- Use custom roles for special cases
- Both types coexist and work seamlessly

---

## Best Practices

### 1. Role Design

✅ **Do:**
- Create roles based on job functions
- Name roles clearly (e.g., "Department Manager", not "DM")
- Provide descriptive descriptions
- Follow least-privilege principle

❌ **Don't:**
- Create too many similar roles (consolidate when possible)
- Grant all permissions to custom roles
- Create personal roles for individual users

### 2. Permission Assignment

✅ **Do:**
- Regularly audit role permissions
- Document why each permission is assigned
- Test permissions before assigning to users
- Use permission grouping (related permissions together)

❌ **Don't:**
- Assign permissions "just in case"
- Create duplicate roles with same permissions
- Forget to remove permissions when role scope changes

### 3. Security

✅ **Do:**
- Protect system roles from modification
- Keep default roles unchanged unless necessary
- Audit role and permission changes
- Review user role assignments regularly

❌ **Don't:**
- Rename system roles
- Manually delete protected roles (API will prevent this)
- Assign admin role to untrusted users

### 4. Maintenance

✅ **Do:**
- Regularly review and clean up unused roles
- Document role-to-permission mappings
- Keep descriptions updated
- Test role changes before rollout

❌ **Don't:**
- Create test roles in production
- Leave inactive permissions assigned
- Forget to test all role combinations

---

## Summary

Phase 3.5 provides complete custom role management while maintaining backward compatibility with Phase 3 hardcoded roles. Use the API endpoints and UI to create, manage, and assign permissions for your organization's specific needs.

**Key Takeaways:**
- ✅ System roles protected from deletion
- ✅ Database-driven custom permissions
- ✅ Real-time permission assignment
- ✅ Full backward compatibility with Phase 3
- ✅ Comprehensive testing and documentation
- ✅ Ready for production deployment

---

## Additional Resources

- [Phase 3 Implementation](PHASE3-IMPLEMENTATION-COMPLETE.md)
- [Phase 3.5 Implementation Plan](PHASE3.5-IMPLEMENTATION-PLAN.md)
- [Backend Testing Guide](PHASE3-BACKEND-TESTING-GUIDE.md)
- [Frontend Implementation Guide](PHASE3-FRONTEND-IMPLEMENTATION-GUIDE.md)

