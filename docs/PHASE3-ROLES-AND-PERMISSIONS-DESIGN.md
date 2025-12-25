# Phase 3+: Custom Roles and Permissions Management

## 🎯 Problem Statement

Currently, roles are **hardcoded globally** (admin, approver, requester, finance, viewer). We need a system where:

1. **System/Default Roles** exist for all organizations (predefined)
2. **Organization Admins** can create custom roles
3. **Organization Admins** can assign permissions to roles
4. **Roles are organization-scoped** (different org can have different role definitions)
5. **Users maintain global role** + **organization-specific role** (as currently designed)

---

## 🏗️ Current Architecture

### Current State
```
User Model
├─ role: "admin" | "approver" | "requester" | "finance" | "viewer"  (GLOBAL)
└─ CurrentOrganizationID (which org user is in)

OrganizationMember Model
├─ role: "admin" | "approver" | "requester" | "finance" | "viewer"  (ORG-SPECIFIC)
└─ (inherits global role mappings)
```

### Problem
- Roles are hardcoded everywhere
- Organization admins can't create new roles
- Can't customize permissions per organization
- Tight coupling between role names and permissions

---

## ✅ Proposed Solution

### New Architecture
```
User Model (Unchanged)
├─ role: "admin" | "approver" | "requester" | "finance" | "viewer"  (GLOBAL)
└─ CurrentOrganizationID

Organization Model (Enhanced)
└─ NEW: Roles and permissions managed per organization

NEW OrganizationRole Model
├─ id: UUID
├─ organizationId: UUID
├─ name: "Requester" | "Manager" | "Custom Role Name"  (ORG-SPECIFIC)
├─ description: string
├─ isDefault: boolean  (true for system roles like "Requester")
├─ permissions: []
├─ memberCount: int
├─ createdBy: User ID
├─ createdAt: timestamp

NEW OrganizationPermission Model
├─ id: UUID
├─ organizationId: UUID  (can override system permissions)
├─ resource: "requisition" | "budget" | "vendor" | ...
├─ action: "create" | "read" | "update" | "delete" | "approve" | ...
├─ description: string
├─ createdBy: User ID

NEW PermissionAssignment Model
├─ id: UUID
├─ organizationRoleId: UUID  (which role)
├─ organizationPermissionId: UUID  (which permission)
├─ assignedBy: User ID
├─ assignedAt: timestamp
```

---

## 📊 Data Flow

### System Initialization
```
When app starts:
  1. Load system roles (hardcoded/seeded)
  2. Load system permissions (hardcoded/seeded)
  3. For each organization:
     - Create default roles if not exist
     - Create default permissions if not exist
     - Create default permission assignments
```

### Organization Admin Creates Custom Role
```
Organization Admin
    ↓
POST /api/v1/organizations/{orgId}/roles
  {
    "name": "Senior Approver",
    "description": "Can approve high-value requisitions",
    "permissions": ["requisition:read", "requisition:approve"]
  }
    ↓
Backend:
  1. Verify user is org admin
  2. Validate permission names exist
  3. Create OrganizationRole
  4. Create PermissionAssignment entries
  5. Return created role
    ↓
Response:
  {
    "id": "role_xyz",
    "name": "Senior Approver",
    "organizationId": "org_123",
    "permissions": [...]
  }
    ↓
Organization Admin can now:
  - Assign this role to members
  - View members with this role
  - Edit role permissions
  - Delete role (if no members)
```

### Organization Admin Assigns Member to Role
```
Organization Admin
    ↓
PATCH /api/v1/organizations/{orgId}/members/{memberId}
  {
    "roleId": "role_xyz"  (OrganizationRole ID, not User.role)
  }
    ↓
Backend:
  1. Verify user is org admin
  2. Verify target user is org member
  3. Verify role exists in organization
  4. Update OrganizationMember.roleId = role_xyz
  5. Return updated member
    ↓
Now when user requests data:
  1. Get user's global role (User.role)
  2. Get organization context (X-Organization-ID header)
  3. Look up OrganizationMember.roleId
  4. Look up OrganizationRole permissions
  5. Check permission against request
```

---

## 🗂️ Database Design

### New Tables

#### organization_roles
```sql
CREATE TABLE organization_roles (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL REFERENCES organizations(id),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    is_default BOOLEAN DEFAULT false,  -- true for system roles (Requester, Approver, etc.)
    is_active BOOLEAN DEFAULT true,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP,  -- soft delete

    UNIQUE(organization_id, name),  -- roles unique per org
    INDEX(organization_id),
    INDEX(created_by)
);
```

#### organization_permissions
```sql
CREATE TABLE organization_permissions (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL REFERENCES organizations(id),
    resource VARCHAR(100) NOT NULL,  -- requisition, budget, vendor, etc.
    action VARCHAR(100) NOT NULL,    -- create, read, update, delete, approve, etc.
    description TEXT,
    is_default BOOLEAN DEFAULT false,  -- true for system permissions
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP,  -- soft delete

    UNIQUE(organization_id, resource, action),
    INDEX(organization_id),
    INDEX(resource, action)
);
```

#### role_permissions
```sql
CREATE TABLE role_permissions (
    id UUID PRIMARY KEY,
    role_id UUID NOT NULL REFERENCES organization_roles(id),
    permission_id UUID NOT NULL REFERENCES organization_permissions(id),
    assigned_by UUID NOT NULL REFERENCES users(id),
    assigned_at TIMESTAMP DEFAULT NOW(),

    UNIQUE(role_id, permission_id),
    INDEX(role_id),
    INDEX(permission_id)
);
```

### Modified Tables

#### organization_members
```sql
-- Add new column:
ALTER TABLE organization_members
ADD COLUMN organization_role_id UUID REFERENCES organization_roles(id);

-- Old column (keep for backward compatibility during migration):
-- role VARCHAR(255)  -- DEPRECATED: use organization_role_id instead

-- Note: During migration, map existing role strings to organization_roles
```

---

## 🎯 Implementation Strategy

### Approach: Hybrid (Backward Compatible)

**Phase 3** (Current):
- Keep hardcoded role permissions
- Works with existing User.role
- Works with existing OrganizationMember.role strings
- Good for MVP

**Phase 3.5** (Optional Before Phase 4):
- Create OrganizationRole and OrganizationPermission tables
- Migrate existing roles to database
- Keep backward compatibility
- Allow custom roles

**Phase 4**:
- Full custom role management UI
- Migrate all role checks to use OrganizationRole
- Remove hardcoded role mappings

---

## 🔄 Two-Tier Role System

### Tier 1: Global User Role (System-wide)
```
User.role = "admin" | "approver" | "requester" | "finance" | "viewer"

Purpose: System-wide designation, used for:
- Super admin functionality (across all orgs)
- Default permission assignment (if org role not set)
- User type in system (approver by trade, requester by trade, etc.)

Example:
- User = "approver" globally
- Works in Org A as "approver"
- Works in Org B as "senior approver" (custom role)
- Works in Org C as "requester" (demoted temporarily)
```

### Tier 2: Organization Role (Organization-specific)
```
OrganizationMember.organizationRoleId → OrganizationRole

Purpose: Per-organization designation, used for:
- Specific permissions in that organization
- Different role names per organization
- Custom permissions per organization

Example:
- Same user in different orgs has different roles:
  - In Org A: "Senior Manager" (can do more)
  - In Org B: "Team Lead" (can do less)
  - In Org C: "Custom Role" (organization-specific)
```

---

## 🔐 Permission Resolution Algorithm

When a request comes in with org context:

```
function CanUserDo(userId, orgId, resource, action):
    1. Get user from database
    2. Verify user is member of organization
    3. Get OrganizationMember for user in org

    IF OrganizationMember.organizationRoleId is set:
        4. Get OrganizationRole by ID
        5. Get PermissionAssignments for that role
        6. Check if (resource, action) in permissions
        7. Return result

    ELSE (backward compatibility):
        4. Get OrganizationMember.role string (old format)
        5. Use hardcoded RolePermissions mapping
        6. Check if (resource, action) in permissions
        7. Return result
```

---

## 📋 Organization Admin Capabilities

### Views/Screens Needed

#### 1. Role Management Dashboard
```
/organizations/{orgId}/settings/roles

- List all roles in organization
  ├─ System roles (cannot delete)
  │  ├─ Admin (System)
  │  ├─ Requester (System)
  │  ├─ Approver (System)
  │  └─ ...
  └─ Custom roles (can edit/delete)
     ├─ Senior Manager
     ├─ Project Lead
     └─ ...

- Create New Role
  ├─ Role name
  ├─ Description
  ├─ Select permissions (multi-select)
  └─ Create

- Edit Role
  ├─ Change name
  ├─ Change description
  ├─ Add/remove permissions
  ├─ View members with this role
  └─ Save / Delete
```

#### 2. Permission Management Dashboard
```
/organizations/{orgId}/settings/permissions

- View all permissions available in organization
  ├─ System permissions (cannot modify)
  │  ├─ requisition:create
  │  ├─ requisition:read
  │  ├─ requisition:update
  │  ├─ requisition:approve
  │  └─ ...
  └─ Custom permissions (can add/modify)
     └─ (organization-specific permissions)

- Create Custom Permission
  ├─ Resource
  ├─ Action
  ├─ Description
  └─ Create

- View which roles have this permission
```

#### 3. Member Management (Enhanced)
```
/organizations/{orgId}/settings/members

- List all members
  ├─ Member name
  ├─ Email
  ├─ Current role
  ├─ Department
  ├─ Status
  └─ Actions (edit/remove)

- Edit Member
  ├─ Select role from dropdown
  │  (shows all OrganizationRoles in org)
  ├─ Or select global role (deprecated)
  └─ Save

- Bulk Assign Roles
  ├─ Select multiple members
  ├─ Assign same role to all
  └─ Confirm
```

---

## 🔌 API Endpoints (Phase 3.5+)

### Role Management

```
GET    /api/v1/organizations/{orgId}/roles
       └─ List all roles in organization

POST   /api/v1/organizations/{orgId}/roles
       ├─ name: string
       ├─ description: string
       └─ permissions: [permissionId, ...]

GET    /api/v1/organizations/{orgId}/roles/{roleId}
       └─ Get single role with permissions

PATCH  /api/v1/organizations/{orgId}/roles/{roleId}
       ├─ name: string (optional)
       ├─ description: string (optional)
       └─ permissions: [permissionId, ...] (optional)

DELETE /api/v1/organizations/{orgId}/roles/{roleId}
       └─ Delete role (only if no members assigned)
```

### Permission Management

```
GET    /api/v1/organizations/{orgId}/permissions
       └─ List all permissions in organization

POST   /api/v1/organizations/{orgId}/permissions
       ├─ resource: string
       ├─ action: string
       └─ description: string

GET    /api/v1/organizations/{orgId}/permissions/{permissionId}
       └─ Get single permission with role assignments

PATCH  /api/v1/organizations/{orgId}/permissions/{permissionId}
       └─ Update permission details

DELETE /api/v1/organizations/{orgId}/permissions/{permissionId}
       └─ Delete custom permission
```

### Member Role Assignment

```
PATCH  /api/v1/organizations/{orgId}/members/{memberId}
       ├─ organizationRoleId: string (NEW)
       └─ role: string (DEPRECATED)

GET    /api/v1/organizations/{orgId}/members/{memberId}
       └─ Include organizationRoleId and role info
```

---

## 🔒 Authorization for Role Management

### Who can manage roles?
```
CREATE role:
  ✅ Super admin (across all orgs)
  ✅ Organization admin (for their org only)
  ❌ Regular members

EDIT role:
  ✅ Super admin
  ✅ Organization admin
  ❌ Role creator (unless they're also admin)
  ❌ Regular members

ASSIGN role to member:
  ✅ Super admin
  ✅ Organization admin
  ❌ Manager or approver
  ❌ Regular members

DELETE role:
  ✅ Super admin
  ✅ Organization admin (if no members assigned)
  ❌ Others
```

---

## 🧪 Testing Strategy

### Unit Tests

**RoleService Tests**:
- `CreateRole` - with valid permissions
- `CreateRole` - with invalid permissions
- `CreateRole` - duplicate name in same org
- `UpdateRole` - add permissions
- `UpdateRole` - remove permissions
- `UpdateRole` - system role (should fail)
- `DeleteRole` - with no members
- `DeleteRole` - with members assigned (should fail)

**PermissionService Tests**:
- `CreatePermission` - custom permission
- `CreatePermission` - duplicate (should fail)
- `GetRolePermissions` - from OrganizationRole
- `GetRolePermissions` - from legacy role string
- `HasPermission` - using OrganizationRole
- `HasPermission` - using legacy role

### Integration Tests

**Role Management Flow**:
1. Admin creates new role "Senior Manager"
2. Admin assigns permissions to role
3. Admin assigns member to role
4. Verify member has permissions
5. Admin removes permission from role
6. Verify member no longer has permission
7. Admin tries to delete role (should fail - has members)
8. Admin removes member from role
9. Admin deletes role successfully

**Permission Assignment Flow**:
1. Admin assigns member to "Requester" role
2. Member tries to create requisition (✓ can)
3. Member tries to approve requisition (✗ cannot)
4. Admin creates custom role "Finance Approver"
5. Admin adds "requisition:approve" permission
6. Admin assigns member to new role
7. Member tries to approve requisition (✓ can)

---

## 📈 Phasing Strategy

### Phase 3: Current (Hardcoded Roles)
```
Backend:
  ✅ PermissionService with hardcoded mapping
  ✅ RequirePermission middleware
  ✅ Handler permission checks

Frontend:
  ✅ usePermissions hook
  ✅ PermissionGuard components
  ✅ Component updates

No custom role management yet
```

### Phase 3.5: Optional (Database Roles - NOT Required for Phase 3)
```
Backend:
  - Create OrganizationRole model
  - Create OrganizationPermission model
  - Migrate existing roles to database
  - Update PermissionService to check database
  - Add role management endpoints
  - Keep backward compatibility with hardcoded roles

Frontend:
  - Add role management UI
  - Add permission assignment UI
  - Show available roles when assigning members

Database migration:
  - Seed system roles for all existing orgs
  - Seed system permissions
  - Migrate existing OrganizationMember.role → organizationRoleId
```

### Phase 4: Advanced (Custom Permissions)
```
Full custom role and permission system
- Complete UI for role management
- Custom permissions UI
- Advanced permission logic (conditions, etc.)
- Role inheritance
- Department-based role assignment
```

---

## 🎓 Key Design Decisions

### 1. Keep User.role Global
**Why**:
- Represents user's type system-wide
- Used for super admin checks
- Backward compatible
- Easier migration path

### 2. Use OrganizationRole for Specific Org Roles
**Why**:
- Each org can have different role definitions
- Allows custom roles
- Scalable system

### 3. Hardcoded Roles in Phase 3
**Why**:
- Simpler to implement
- No database overhead for role lookups
- Sufficient for MVP
- Easy migration to database later

### 4. Backward Compatible Approach
**Why**:
- Don't break existing code
- Can migrate gradually
- Run both systems in parallel
- Easy rollback if issues

---

## 🚀 Recommendation

### For Phase 3 (Now)
✅ Use hardcoded roles (admin, approver, requester, finance, viewer)
✅ No database role management
✅ Simpler, faster implementation
✅ Works for MVP

### For Phase 3.5 (Optional, before Phase 4)
- Add database role models
- Allow custom roles per organization
- Admin can create and manage roles
- Still backward compatible

### For Phase 4
- Full UI for role management
- Advanced permission features
- Remove backward compatibility (hardcoded roles)

---

## 📊 Implementation Timeline

### Phase 3 (4-6 hours)
- Hardcoded permission service
- Permission middleware
- Component updates
- Testing

**Cost**: Low
**Risk**: Low
**User Impact**: None (internal refactor)

### Phase 3.5 (8-12 hours - Optional)
- Database models for roles
- API endpoints for role management
- Migration script
- Basic role management UI

**Cost**: Medium
**Risk**: Medium
**User Impact**: High (orgs can customize roles)

### Phase 4 (12-16 hours)
- Advanced role management
- Custom permissions UI
- Permission conditions
- Role inheritance

**Cost**: High
**Risk**: High
**User Impact**: Very High (powerful customization)

---

## ✨ Summary

**Phase 3 Answer: Roles are hardcoded, organization admins can't create roles yet**

But we've designed the system to support:
1. **Phase 3.5**: Database-driven roles per organization
2. **Phase 4**: Full custom role and permission management

**Current Flow**:
```
System (hardcoded) → Phase 3
  → Add to database (Phase 3.5)
  → Full customization (Phase 4)
```

This is the recommended MVP approach: start simple, add complexity when needed.

---

## 🔗 Related Documents

- PHASE3-IMPLEMENTATION-PLAN.md (current)
- PHASE3-PERMISSION-MAPPING.md (what permissions exist)
- PHASE3-ROADMAP.md (visual overview)

---

**Next Steps**:
1. Implement Phase 3 with hardcoded roles
2. After Phase 3 is validated, plan Phase 3.5
3. Then consider Phase 4 for advanced features

