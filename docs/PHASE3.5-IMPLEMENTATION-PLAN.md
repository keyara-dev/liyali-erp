# Phase 3.5 Implementation Plan

**Status**: Planning
**Effort**: 12-16 hours
**Goal**: Enable organization admins to create and customize roles with permissions

---

## Overview

Phase 3.5 extends Phase 3 by allowing **organization admins** to:
- ✅ Create custom roles (e.g., "Senior Manager", "Budget Controller", "Approval Clerk")
- ✅ Assign permissions to custom roles
- ✅ Manage which users have which roles
- ✅ Update and delete custom roles (system default roles are protected)

### Architecture Change

**Phase 3:**
```
User.Role (hardcoded: admin, approver, requester, finance, viewer)
    ↓
RolePermissions[role] (hardcoded mapping)
    ↓
Permission granted/denied
```

**Phase 3.5:**
```
User.Role + OrganizationMember.Role
    ↓
Check database first:
  OrganizationRole → PermissionAssignment → OrganizationPermission
    ↓
Fall back to hardcoded if not found
    ↓
Permission granted/denied
```

---

## Implementation Tasks

### Phase 3.5A: Backend API Endpoints (6 hours)

#### Task 3.5A.1: Role Management Endpoints (3 hours)

Create handlers in `backend/handlers/roles.go`:

```go
// GET /api/v1/organization/roles
// Returns all roles for the organization
func GetOrganizationRoles(c fiber.Ctx) error {
  organizationID := c.Locals("organizationID").(string)
  svc := services.NewRoleManagementService(config.DB)
  roles, err := svc.GetOrganizationRoles(organizationID)
  return utils.SendSuccess(c, fiber.StatusOK, roles, "Roles retrieved", nil)
}

// POST /api/v1/organization/roles
// Create a new role
func CreateRole(c fiber.Ctx) error {
  var req CreateRoleRequest
  c.BindJSON(&req)
  organizationID := c.Locals("organizationID").(string)

  svc := services.NewRoleManagementService(config.DB)
  role, err := svc.CreateOrganizationRole(organizationID, req.Name, req.Description)
  return utils.SendSuccess(c, fiber.StatusCreated, role, "Role created", nil)
}

// PUT /api/v1/organization/roles/:roleId
// Update a role
func UpdateRole(c fiber.Ctx) error {
  roleID := c.Params("roleId")
  var req UpdateRoleRequest
  c.BindJSON(&req)

  svc := services.NewRoleManagementService(config.DB)
  role, err := svc.UpdateOrganizationRole(roleID, req.Name, req.Description)
  return utils.SendSuccess(c, fiber.StatusOK, role, "Role updated", nil)
}

// DELETE /api/v1/organization/roles/:roleId
// Delete a role (only custom roles, not system defaults)
func DeleteRole(c fiber.Ctx) error {
  roleID := c.Params("roleId")
  svc := services.NewRoleManagementService(config.DB)
  err := svc.DeleteOrganizationRole(roleID)
  return utils.SendSuccess(c, fiber.StatusOK, nil, "Role deleted", nil)
}

// GET /api/v1/organization/roles/:roleId/permissions
// Get all permissions assigned to a role
func GetRolePermissions(c fiber.Ctx) error {
  roleID := c.Params("roleId")
  svc := services.NewRoleManagementService(config.DB)
  perms, err := svc.GetRolePermissions(roleID)
  return utils.SendSuccess(c, fiber.StatusOK, perms, "Permissions retrieved", nil)
}
```

Update `backend/routes/routes.go`:
```go
// Organization role management (admin only)
orgRoles := tenant.Group("/organization/roles")
orgRoles.Get("/",
  middleware.RequirePermission(config.DB, "organization", "manage_workflows"),
  handlers.GetOrganizationRoles)
orgRoles.Post("/",
  middleware.RequirePermission(config.DB, "organization", "manage_workflows"),
  handlers.CreateRole)
orgRoles.Put("/:roleId",
  middleware.RequirePermission(config.DB, "organization", "manage_workflows"),
  handlers.UpdateRole)
orgRoles.Delete("/:roleId",
  middleware.RequirePermission(config.DB, "organization", "manage_workflows"),
  handlers.DeleteRole)
orgRoles.Get("/:roleId/permissions",
  middleware.RequirePermission(config.DB, "organization", "manage_workflows"),
  handlers.GetRolePermissions)
```

#### Task 3.5A.2: Permission Assignment Endpoints (2 hours)

```go
// POST /api/v1/organization/roles/:roleId/permissions/:permissionId
// Assign permission to role
func AssignPermissionToRole(c fiber.Ctx) error {
  roleID := c.Params("roleId")
  permissionID := c.Params("permissionId")

  svc := services.NewRoleManagementService(config.DB)
  assignment, err := svc.AssignPermissionToRole(roleID, permissionID)
  return utils.SendSuccess(c, fiber.StatusCreated, assignment, "Permission assigned", nil)
}

// DELETE /api/v1/organization/roles/:roleId/permissions/:permissionId
// Remove permission from role
func RemovePermissionFromRole(c fiber.Ctx) error {
  roleID := c.Params("roleId")
  permissionID := c.Params("permissionId")

  svc := services.NewRoleManagementService(config.DB)
  err := svc.RemovePermissionFromRole(roleID, permissionID)
  return utils.SendSuccess(c, fiber.StatusOK, nil, "Permission removed", nil)
}

// GET /api/v1/organization/permissions
// Get all available permissions for the organization
func GetOrganizationPermissions(c fiber.Ctx) error {
  organizationID := c.Locals("organizationID").(string)
  svc := services.NewRoleManagementService(config.DB)
  perms, err := svc.GetOrganizationPermissions(organizationID)
  return utils.SendSuccess(c, fiber.StatusOK, perms, "Permissions retrieved", nil)
}
```

Add routes:
```go
// Permission management
permissions := tenant.Group("/organization/permissions")
permissions.Get("/",
  middleware.RequirePermission(config.DB, "organization", "manage_workflows"),
  handlers.GetOrganizationPermissions)

// Permission assignment
orgRoles.Post("/:roleId/permissions/:permissionId",
  middleware.RequirePermission(config.DB, "organization", "manage_workflows"),
  handlers.AssignPermissionToRole)
orgRoles.Delete("/:roleId/permissions/:permissionId",
  middleware.RequirePermission(config.DB, "organization", "manage_workflows"),
  handlers.RemovePermissionFromRole)
```

#### Task 3.5A.3: Update PermissionService to Check Database (1 hour)

Modify `backend/services/permission_service.go`:

```go
// HasPermission now checks database first, falls back to hardcoded
func (ps *PermissionService) HasPermission(userID, organizationID, role, resource, action string) bool {
  // First check custom permissions in database
  customPermissions, err := ps.getCustomPermissions(userID, organizationID, role)
  if err == nil && len(customPermissions) > 0 {
    return ps.permissionExists(customPermissions, resource, action)
  }

  // Fall back to hardcoded role permissions
  return ps.checkRolePermission(role, resource, action)
}

// Implement getCustomPermissions to query database
func (ps *PermissionService) getCustomPermissions(userID, organizationID, roleName string) ([]Permission, error) {
  var permissions []Permission

  // Find the organization role by name
  var orgRole models.OrganizationRole
  if err := ps.db.Where("organization_id = ? AND name = ?", organizationID, roleName).
    First(&orgRole).Error; err != nil {
    return nil, fmt.Errorf("role not found in database")
  }

  // Get all permissions for this role
  var orgPerms []models.OrganizationPermission
  if err := ps.db.
    Joins("INNER JOIN permission_assignments ON permission_assignments.organization_permission_id = organization_permissions.id").
    Where("permission_assignments.organization_role_id = ?", orgRole.ID).
    Find(&orgPerms).Error; err != nil {
    return nil, fmt.Errorf("failed to fetch permissions")
  }

  // Convert to Permission structs
  for _, perm := range orgPerms {
    permissions = append(permissions, Permission{
      Resource: perm.Resource,
      Action:   perm.Action,
    })
  }

  return permissions, nil
}
```

### Phase 3.5B: Frontend Role Management UI (6-8 hours)

#### Task 3.5B.1: Create Role List Page (2 hours)

`frontend/src/app/admin/roles/page.tsx`:

```tsx
"use client";

import { useQuery, useMutation } from "@tanstack/react-query";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Table } from "@/components/ui/table";
import { Card } from "@/components/ui/card";
import { RoleModal } from "./role-modal";
import { getRolesAction, deleteRoleAction } from "@/app/_actions/roles";
import { AdminGuard } from "@/components/auth/permission-guard";

export default function RolesPage() {
  const [showModal, setShowModal] = useState(false);
  const [selectedRole, setSelectedRole] = useState(null);

  const { data: roles, isLoading, refetch } = useQuery({
    queryKey: ["organization-roles"],
    queryFn: () => getRolesAction(),
    staleTime: 5 * 60 * 1000,
  });

  const deleteMutation = useMutation({
    mutationFn: (roleId: string) => deleteRoleAction(roleId),
    onSuccess: () => refetch(),
  });

  return (
    <AdminGuard>
      <div className="space-y-6">
        <div className="flex justify-between items-center">
          <h1>Organization Roles</h1>
          <Button onClick={() => {
            setSelectedRole(null);
            setShowModal(true);
          }}>
            Create Role
          </Button>
        </div>

        <Card>
          <Table>
            <thead>
              <tr>
                <th>Role Name</th>
                <th>Description</th>
                <th>Permissions</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {roles?.map((role) => (
                <tr key={role.id}>
                  <td>{role.name}</td>
                  <td>{role.description}</td>
                  <td>
                    <Button
                      variant="link"
                      onClick={() => {
                        // Show permissions modal
                      }}
                    >
                      View ({role.permissions?.length || 0})
                    </Button>
                  </td>
                  <td>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => {
                        setSelectedRole(role);
                        setShowModal(true);
                      }}
                    >
                      Edit
                    </Button>
                    {!role.isDefault && (
                      <Button
                        variant="destructive"
                        size="sm"
                        onClick={() => deleteMutation.mutate(role.id)}
                      >
                        Delete
                      </Button>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </Table>
        </Card>

        <RoleModal
          role={selectedRole}
          open={showModal}
          onClose={() => {
            setShowModal(false);
            refetch();
          }}
        />
      </div>
    </AdminGuard>
  );
}
```

#### Task 3.5B.2: Create Role Modal Component (2 hours)

`frontend/src/app/admin/roles/role-modal.tsx`:

```tsx
"use client";

import { useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { createRoleAction, updateRoleAction } from "@/app/_actions/roles";

interface RoleModalProps {
  role?: any;
  open: boolean;
  onClose: () => void;
}

export function RoleModal({ role, open, onClose }: RoleModalProps) {
  const [name, setName] = useState(role?.name || "");
  const [description, setDescription] = useState(role?.description || "");

  const createMutation = useMutation({
    mutationFn: () => createRoleAction(name, description),
    onSuccess: onClose,
  });

  const updateMutation = useMutation({
    mutationFn: () => updateRoleAction(role.id, name, description),
    onSuccess: onClose,
  });

  const handleSubmit = () => {
    if (role?.id) {
      updateMutation.mutate();
    } else {
      createMutation.mutate();
    }
  };

  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{role ? "Edit Role" : "Create Role"}</DialogTitle>
        </DialogHeader>

        <div className="space-y-4">
          <div>
            <label>Role Name</label>
            <Input
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="e.g., Senior Manager"
            />
          </div>

          <div>
            <label>Description</label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="What is this role for?"
              className="w-full border rounded p-2"
            />
          </div>

          <div className="flex gap-2 justify-end">
            <Button variant="outline" onClick={onClose}>
              Cancel
            </Button>
            <Button onClick={handleSubmit}>
              {role ? "Update" : "Create"}
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
```

#### Task 3.5B.3: Create Permission Assignment Component (2 hours)

`frontend/src/app/admin/roles/permissions-modal.tsx`:

```tsx
"use client";

import { useQuery, useMutation } from "@tanstack/react-query";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Checkbox } from "@/components/ui/checkbox";
import { Button } from "@/components/ui/button";
import {
  getRolePermissionsAction,
  getAvailablePermissionsAction,
  assignPermissionAction,
  removePermissionAction,
} from "@/app/_actions/roles";

interface PermissionsModalProps {
  roleId: string;
  open: boolean;
  onClose: () => void;
}

export function PermissionsModal({ roleId, open, onClose }: PermissionsModalProps) {
  const { data: assignedPerms } = useQuery({
    queryKey: ["role-permissions", roleId],
    queryFn: () => getRolePermissionsAction(roleId),
    enabled: open,
  });

  const { data: availablePerms } = useQuery({
    queryKey: ["available-permissions"],
    queryFn: () => getAvailablePermissionsAction(),
    enabled: open,
  });

  const assignMutation = useMutation({
    mutationFn: (permId: string) => assignPermissionAction(roleId, permId),
  });

  const removeMutation = useMutation({
    mutationFn: (permId: string) => removePermissionAction(roleId, permId),
  });

  const hasPermission = (permId: string) => {
    return assignedPerms?.some((p) => p.id === permId);
  };

  // Group permissions by resource
  const groupedPerms = availablePerms?.reduce((acc, perm) => {
    if (!acc[perm.resource]) acc[perm.resource] = [];
    acc[perm.resource].push(perm);
    return acc;
  }, {});

  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>Assign Permissions</DialogTitle>
        </DialogHeader>

        <div className="space-y-6 max-h-96 overflow-y-auto">
          {Object.entries(groupedPerms || {}).map(([resource, perms]: [string, any[]]) => (
            <div key={resource}>
              <h3 className="font-semibold capitalize">{resource}</h3>
              <div className="space-y-2 ml-4">
                {perms.map((perm) => (
                  <label key={perm.id} className="flex items-center gap-2 cursor-pointer">
                    <Checkbox
                      checked={hasPermission(perm.id)}
                      onChange={(e) => {
                        if (e.target.checked) {
                          assignMutation.mutate(perm.id);
                        } else {
                          removeMutation.mutate(perm.id);
                        }
                      }}
                    />
                    <span>{perm.action}</span>
                    <span className="text-sm text-gray-500">{perm.description}</span>
                  </label>
                ))}
              </div>
            </div>
          ))}
        </div>

        <div className="flex justify-end">
          <Button onClick={onClose}>Done</Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
```

#### Task 3.5B.4: Create Server Actions (2 hours)

`frontend/src/app/_actions/roles.ts`:

```typescript
"use server";

import { apiCall } from "@/utils/api";

export async function getRolesAction() {
  return apiCall("/organization/roles", { method: "GET" });
}

export async function createRoleAction(name: string, description: string) {
  return apiCall("/organization/roles", {
    method: "POST",
    body: { name, description },
  });
}

export async function updateRoleAction(roleId: string, name: string, description: string) {
  return apiCall(`/organization/roles/${roleId}`, {
    method: "PUT",
    body: { name, description },
  });
}

export async function deleteRoleAction(roleId: string) {
  return apiCall(`/organization/roles/${roleId}`, { method: "DELETE" });
}

export async function getRolePermissionsAction(roleId: string) {
  return apiCall(`/organization/roles/${roleId}/permissions`, { method: "GET" });
}

export async function getAvailablePermissionsAction() {
  return apiCall("/organization/permissions", { method: "GET" });
}

export async function assignPermissionAction(roleId: string, permissionId: string) {
  return apiCall(`/organization/roles/${roleId}/permissions/${permissionId}`, {
    method: "POST",
  });
}

export async function removePermissionAction(roleId: string, permissionId: string) {
  return apiCall(`/organization/roles/${roleId}/permissions/${permissionId}`, {
    method: "DELETE",
  });
}
```

### Phase 3.5C: Testing & Documentation (2 hours)

#### Task 3.5C.1: Write Tests

Update `backend/services/role_management_service_test.go`:

```go
func TestRoleManagement(t *testing.T) {
  db := setupTestDB()
  svc := NewRoleManagementService(db)
  orgID := "test-org"

  // Test: Create role
  role, err := svc.CreateOrganizationRole(orgID, "Manager", "Manages team")
  assert.NoError(t, err)
  assert.NotEmpty(t, role.ID)

  // Test: Cannot delete system default role
  err = svc.DeleteOrganizationRole("admin")
  assert.Error(t, err) // Should fail

  // Test: Can delete custom role
  err = svc.DeleteOrganizationRole(role.ID)
  assert.NoError(t, err)
}
```

#### Task 3.5C.2: Update Documentation

Create `docs/PHASE3.5-USAGE-GUIDE.md`:

```markdown
# Phase 3.5 Usage Guide

## Admin Creating Custom Roles

1. Go to Admin Panel → Organization Roles
2. Click "Create Role"
3. Enter role name and description
4. Click "Create"
5. Click "View Permissions"
6. Check/uncheck permissions
7. Click "Done"

## Using Custom Roles

When creating a user, instead of assigning one of the 5 system roles, org admins can now assign custom roles with specific permission sets.

## Permission Precedence

System default roles (admin, approver, requester, finance, viewer) still exist and work as Phase 3. But custom roles in a specific organization override them.
```

---

## Implementation Order

1. **Day 1 (6 hours)**: Backend API endpoints
   - Create role handlers
   - Create permission handlers
   - Update PermissionService
   - Add routes

2. **Day 2-3 (6-8 hours)**: Frontend UI
   - Role list page
   - Role create/edit modal
   - Permission assignment modal
   - Server actions

3. **Day 3-4 (2 hours)**: Testing & documentation
   - Unit tests
   - Integration tests
   - Documentation updates

---

## Files to Create

### Backend
- `backend/handlers/roles.go` (250 lines)
- `backend/services/role_management_service_test.go` (150 lines)

### Frontend
- `frontend/src/app/admin/roles/page.tsx` (100 lines)
- `frontend/src/app/admin/roles/role-modal.tsx` (100 lines)
- `frontend/src/app/admin/roles/permissions-modal.tsx` (150 lines)
- `frontend/src/app/_actions/roles.ts` (80 lines)

### Documentation
- `docs/PHASE3.5-USAGE-GUIDE.md`

---

## Files to Modify

### Backend
- `backend/routes/routes.go` (Add 15 new route definitions)
- `backend/services/permission_service.go` (Update getCustomPermissions implementation)

### Frontend
- `frontend/src/app/admin/layout.tsx` (Add roles link to menu)

---

## Testing the Implementation

### Backend Test
```bash
# Test role creation
curl -X POST http://localhost:3000/api/v1/organization/roles \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Senior Manager",
    "description": "Senior level manager"
  }'

# Test permission assignment
curl -X POST http://localhost:3000/api/v1/organization/roles/ROLE_ID/permissions/PERM_ID \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

### Frontend Test
```tsx
// In admin panel
1. Navigate to /admin/roles
2. Create new role "Test Manager"
3. Click View Permissions
4. Assign "requisition:approve" permission
5. Save
6. Assign user to this role
7. Verify user can approve but not delete
```

---

## Success Criteria

✅ Organization admins can create custom roles
✅ Admins can assign permissions to roles
✅ Custom roles appear in user assignment dropdown
✅ Users with custom roles have correct permissions
✅ System default roles are protected
✅ Permission checks work with database-driven roles
✅ All tests pass
✅ Documentation updated

---

## Notes

- Phase 3 hardcoded roles still work and are still available
- Custom roles are per-organization (multi-tenant)
- Database has default permissions pre-populated via InitializeDefaultPermissionsForOrganization()
- System default roles (admin, approver, requester, finance, viewer) cannot be deleted
- Permission fallback to Phase 3 hardcoded mappings if custom role not found

