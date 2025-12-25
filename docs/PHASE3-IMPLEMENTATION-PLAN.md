# Phase 3 Implementation Plan: Permission-Based Access Control

## 🎯 Objective

Implement **Permission-Based Access Control (PBAC)** to replace role-based checks with fine-grained permission validation. This enables flexible, scalable authorization that decouples role names from capabilities.

---

## 📊 Project Scope

- **Duration**: 4-6 hours (estimated)
- **Complexity**: Medium-High
- **Risk Level**: Medium (affects all authorization logic)
- **Status**: Planning
- **Predecessor**: Phase 2 ✅ Complete

---

## 🎓 Core Concept

### Current State (Role-Based)
```typescript
// Frontend - Old way
if (userRole === "requester") {
  // Can create requisition
}

// Backend - Old way
if role == "requester" {
  // Handle request
}
```

### New State (Permission-Based)
```typescript
// Frontend - New way
if (hasPermission("create_requisition")) {
  // Can create requisition
}

// Backend - New way
if hasPermission("create_requisition") {
  // Handle request
}
```

### Benefits
✅ Decouple role names from capabilities
✅ Enable fine-grained access control
✅ Support custom permissions per organization
✅ Easier to audit and modify permissions
✅ Foundation for future role inheritance

---

## 📋 Implementation Breakdown

### Phase 3A: Backend Permission Service (1.5-2 hours)

#### Task 3A.1: Create Permissions Service
**Time**: 45 minutes
**File**: `backend/services/permissions.go` (new)

**Implementation**:
```go
package services

// Permission represents a capability in the system
type Permission struct {
    Resource string // "requisition", "budget", "organization", etc.
    Action   string // "create", "read", "update", "delete", "approve"
}

// RolePermissions maps roles to their allowed permissions
var RolePermissions = map[string][]Permission{
    "requester": {
        {Resource: "requisition", Action: "create"},
        {Resource: "requisition", Action: "read"},
        {Resource: "requisition", Action: "update"},
        {Resource: "draft", Action: "create"},
    },
    "approver": {
        {Resource: "requisition", Action: "read"},
        {Resource: "requisition", Action: "approve"},
        {Resource: "requisition", Action: "reject"},
    },
    "finance": {
        {Resource: "requisition", Action: "read"},
        {Resource: "budget", Action: "create"},
        {Resource: "budget", Action: "read"},
        {Resource: "budget", Action: "update"},
        {Resource: "vendor", Action: "manage"},
        {Resource: "payment", Action: "approve"},
    },
    "viewer": {
        {Resource: "requisition", Action: "read"},
        {Resource: "budget", Action: "read"},
        {Resource: "report", Action: "read"},
    },
    "admin": {
        // Admins have all permissions
    },
}

// PermissionService provides permission checking
type PermissionService struct{}

func NewPermissionService() *PermissionService {
    return &PermissionService{}
}

// HasPermission checks if a role has a specific permission
func (ps *PermissionService) HasPermission(role, resource, action string) bool {
    // Admins have all permissions
    if role == "admin" {
        return true
    }

    permissions := RolePermissions[role]
    if permissions == nil {
        return false
    }

    for _, perm := range permissions {
        if perm.Resource == resource && perm.Action == action {
            return true
        }
    }
    return false
}

// GetRolePermissions returns all permissions for a role
func (ps *PermissionService) GetRolePermissions(role string) []Permission {
    if role == "admin" {
        // Return all possible permissions
        return getAllPermissions()
    }
    return RolePermissions[role]
}

// getAllPermissions returns all permissions in the system
func getAllPermissions() []Permission {
    var all []Permission
    for _, perms := range RolePermissions {
        all = append(all, perms...)
    }
    return all
}
```

**Checklist**:
- [ ] Service properly handles all roles
- [ ] Admin role checks work
- [ ] Method signatures clear and testable
- [ ] Well-documented with comments

---

#### Task 3A.2: Create RequirePermission Middleware
**Time**: 45 minutes
**File**: `backend/middleware/permission.go` (new)

**Implementation**:
```go
package middleware

import (
    "github.com/gofiber/fiber/v3"
    "github.com/liyali/liyali-gateway/services"
)

// RequirePermission returns a middleware that checks if user has permission
func RequirePermission(resource, action string) fiber.Handler {
    return func(c fiber.Ctx) error {
        // Get user from context (set by auth middleware)
        userRole, ok := c.Locals("userRole").(string)
        if !ok {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                "success": false,
                "message": "User role not found in context",
            })
        }

        // Check permission
        permService := services.NewPermissionService()
        if !permService.HasPermission(userRole, resource, action) {
            return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
                "success": false,
                "message": "Insufficient permissions",
                "required": resource + ":" + action,
            })
        }

        // Continue to next handler
        return c.Next()
    }
}

// RequireAnyPermission checks if user has any of the permissions
func RequireAnyPermission(permissions []struct {
    resource string
    action   string
}) fiber.Handler {
    return func(c fiber.Ctx) error {
        userRole, ok := c.Locals("userRole").(string)
        if !ok {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                "success": false,
                "message": "User role not found in context",
            })
        }

        permService := services.NewPermissionService()
        for _, perm := range permissions {
            if permService.HasPermission(userRole, perm.resource, perm.action) {
                return c.Next()
            }
        }

        return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
            "success": false,
            "message": "Insufficient permissions",
        })
    }
}
```

**Checklist**:
- [ ] Middleware properly checks permissions
- [ ] Returns correct HTTP status (403 Forbidden)
- [ ] Error messages clear
- [ ] Works with auth middleware context

---

#### Task 3A.3: Update Route Handlers with Permissions
**Time**: 30 minutes
**File**: Multiple handler files

**Changes**:

**Before (Role Check)**:
```go
// backend/handlers/requisition.go
func CreateRequisition(c fiber.Ctx) error {
    userRole := c.Locals("userRole").(string)

    if userRole != "requester" && userRole != "admin" {
        return c.Status(fiber.StatusForbidden).JSON(...)
    }

    // Create requisition...
}
```

**After (Permission Check)**:
```go
// backend/handlers/requisition.go
func CreateRequisition(c fiber.Ctx) error {
    permService := services.NewPermissionService()
    userRole := c.Locals("userRole").(string)

    if !permService.HasPermission(userRole, "requisition", "create") {
        return c.Status(fiber.StatusForbidden).JSON(...)
    }

    // Create requisition...
}
```

**Or Using Middleware** (Cleaner):
```go
// backend/routes/routes.go
app.Post("/requisitions",
    middleware.RequirePermission("requisition", "create"),
    handlers.CreateRequisition,
)
```

**Handlers to Update**:
- [ ] CreateRequisition - "requisition:create"
- [ ] GetRequisitions - "requisition:read"
- [ ] UpdateRequisition - "requisition:update"
- [ ] ApproveRequisition - "requisition:approve"
- [ ] RejectRequisition - "requisition:reject"
- [ ] CreateBudget - "budget:create"
- [ ] ManageVendors - "vendor:manage"
- [ ] And others...

**Checklist**:
- [ ] All handlers check permissions
- [ ] Consistent permission naming
- [ ] Permission checks before any business logic
- [ ] Clear error messages

---

### Phase 3B: Frontend Permission Utilities (1.5-2 hours)

#### Task 3B.1: Create Permission Utilities Hook
**Time**: 45 minutes
**File**: `frontend/src/hooks/use-permissions.ts` (new)

**Implementation**:
```typescript
import { useCallback } from 'react';
import { useSessionContext } from '@/contexts/session-context';

interface Permission {
  resource: string;
  action: string;
}

// Role to permissions mapping (mirrors backend)
const ROLE_PERMISSIONS: Record<string, Permission[]> = {
  requester: [
    { resource: 'requisition', action: 'create' },
    { resource: 'requisition', action: 'read' },
    { resource: 'requisition', action: 'update' },
    { resource: 'draft', action: 'create' },
  ],
  approver: [
    { resource: 'requisition', action: 'read' },
    { resource: 'requisition', action: 'approve' },
    { resource: 'requisition', action: 'reject' },
  ],
  finance: [
    { resource: 'requisition', action: 'read' },
    { resource: 'budget', action: 'create' },
    { resource: 'budget', action: 'read' },
    { resource: 'budget', action: 'update' },
    { resource: 'vendor', action: 'manage' },
    { resource: 'payment', action: 'approve' },
  ],
  viewer: [
    { resource: 'requisition', action: 'read' },
    { resource: 'budget', action: 'read' },
    { resource: 'report', action: 'read' },
  ],
  admin: [], // Admins have all permissions (checked separately)
};

export function usePermissions() {
  const { session } = useSessionContext();
  const userRole = session?.role || '';

  // Check single permission
  const hasPermission = useCallback(
    (resource: string, action: string): boolean => {
      // Admins have all permissions
      if (userRole === 'admin') {
        return true;
      }

      const permissions = ROLE_PERMISSIONS[userRole] || [];
      return permissions.some(
        (p) => p.resource === resource && p.action === action
      );
    },
    [userRole]
  );

  // Check multiple permissions (all must be true)
  const hasAllPermissions = useCallback(
    (permissions: Permission[]): boolean => {
      return permissions.every((p) =>
        hasPermission(p.resource, p.action)
      );
    },
    [hasPermission]
  );

  // Check if user has any of the permissions
  const hasAnyPermission = useCallback(
    (permissions: Permission[]): boolean => {
      return permissions.some((p) =>
        hasPermission(p.resource, p.action)
      );
    },
    [hasPermission]
  );

  // Get all permissions for current role
  const getPermissions = useCallback((): Permission[] => {
    if (userRole === 'admin') {
      // Return all permissions
      const all: Permission[] = [];
      Object.values(ROLE_PERMISSIONS).forEach((perms) => {
        all.push(...perms);
      });
      return all;
    }
    return ROLE_PERMISSIONS[userRole] || [];
  }, [userRole]);

  return {
    hasPermission,
    hasAllPermissions,
    hasAnyPermission,
    getPermissions,
    userRole,
  };
}
```

**Checklist**:
- [ ] Hook properly checks permissions
- [ ] Consistent with backend permissions
- [ ] Memoized for performance
- [ ] Well-documented

---

#### Task 3B.2: Create Permission Guard Components
**Time**: 45 minutes
**File**: `frontend/src/components/permission-guard.tsx` (new)

**Implementation**:
```typescript
import { ReactNode } from 'react';
import { usePermissions } from '@/hooks/use-permissions';

interface PermissionGuardProps {
  resource: string;
  action: string;
  children: ReactNode;
  fallback?: ReactNode;
}

/**
 * Conditionally renders children if user has the specified permission
 * Usage: <PermissionGuard resource="requisition" action="create">...</PermissionGuard>
 */
export function PermissionGuard({
  resource,
  action,
  children,
  fallback,
}: PermissionGuardProps) {
  const { hasPermission } = usePermissions();

  if (!hasPermission(resource, action)) {
    return fallback || null;
  }

  return <>{children}</>;
}

interface MultiPermissionGuardProps {
  permissions: Array<{ resource: string; action: string }>;
  mode: 'all' | 'any'; // all = AND, any = OR
  children: ReactNode;
  fallback?: ReactNode;
}

/**
 * Conditionally renders children if user has specified permissions
 * mode="all": requires ALL permissions (AND)
 * mode="any": requires ANY permission (OR)
 */
export function MultiPermissionGuard({
  permissions,
  mode = 'all',
  children,
  fallback,
}: MultiPermissionGuardProps) {
  const { hasAllPermissions, hasAnyPermission } = usePermissions();

  const hasAccess =
    mode === 'all'
      ? hasAllPermissions(permissions)
      : hasAnyPermission(permissions);

  if (!hasAccess) {
    return fallback || null;
  }

  return <>{children}</>;
}
```

**Checklist**:
- [ ] Components render correctly
- [ ] Fallback UI works
- [ ] Both guard components implemented
- [ ] Usage is intuitive

---

#### Task 3B.3: Update Components to Use Permissions
**Time**: 30 minutes
**Files**: Multiple component files

**Changes**:

**Before (Role Check)**:
```typescript
import { useSessionContext } from '@/contexts/session-context';

function RequisitionActions() {
  const { session } = useSessionContext();

  return (
    <div>
      {(session?.role === 'requester' || session?.role === 'admin') && (
        <button>Create Requisition</button>
      )}
      {(session?.role === 'approver' || session?.role === 'admin') && (
        <button>Approve</button>
      )}
    </div>
  );
}
```

**After (Permission Check)**:
```typescript
import { PermissionGuard } from '@/components/permission-guard';

function RequisitionActions() {
  return (
    <div>
      <PermissionGuard resource="requisition" action="create">
        <button>Create Requisition</button>
      </PermissionGuard>

      <PermissionGuard resource="requisition" action="approve">
        <button>Approve</button>
      </PermissionGuard>
    </div>
  );
}
```

**Or Using Hook**:
```typescript
import { usePermissions } from '@/hooks/use-permissions';

function RequisitionActions() {
  const { hasPermission } = usePermissions();

  return (
    <div>
      {hasPermission('requisition', 'create') && (
        <button>Create Requisition</button>
      )}
      {hasPermission('requisition', 'approve') && (
        <button>Approve</button>
      )}
    </div>
  );
}
```

**Components to Update**:
- [ ] RequisitionList - read permission
- [ ] RequisitionForm - create/update permissions
- [ ] RequisitionApprovalModal - approve permission
- [ ] BudgetManager - budget permissions
- [ ] VendorManager - vendor management
- [ ] Navigation/Menu - show/hide items based on permissions
- [ ] And others...

**Checklist**:
- [ ] All role checks replaced with permission checks
- [ ] Using PermissionGuard or usePermissions hook
- [ ] Consistent permission naming
- [ ] Improved readability

---

### Phase 3C: Integration & Testing (2-2.5 hours)

#### Task 3C.1: Permission Mapping Document
**Time**: 30 minutes
**File**: `docs/PHASE3-PERMISSION-MAPPING.md` (new)

**Content**:
```markdown
# Permission Mapping - Phase 3

## Resource: requisition

| Permission | Role | Description |
|-----------|------|-------------|
| requisition:create | requester, admin | Can create new requisitions |
| requisition:read | requester, approver, finance, viewer, admin | Can view requisitions |
| requisition:update | requester, admin | Can update own requisitions |
| requisition:approve | approver, admin | Can approve requisitions |
| requisition:reject | approver, admin | Can reject requisitions |

## Resource: budget

| Permission | Role | Description |
|-----------|------|-------------|
| budget:create | finance, admin | Can create budgets |
| budget:read | finance, viewer, admin | Can view budgets |
| budget:update | finance, admin | Can modify budgets |

... (and so on for all resources)
```

---

#### Task 3C.2: Unit Tests
**Time**: 45 minutes
**Files**:
- `backend/services/permissions_test.go` (new)
- `frontend/src/hooks/__tests__/use-permissions.test.ts` (new)

**Backend Tests**:
```go
func TestHasPermission(t *testing.T) {
    ps := NewPermissionService()

    tests := []struct {
        role       string
        resource   string
        action     string
        expected   bool
    }{
        // Requester
        {"requester", "requisition", "create", true},
        {"requester", "budget", "create", false},
        // Approver
        {"approver", "requisition", "approve", true},
        {"approver", "requisition", "create", false},
        // Admin
        {"admin", "requisition", "create", true},
        {"admin", "budget", "create", true},
        {"admin", "anything", "anything", true},
    }

    for _, tt := range tests {
        t.Run(tt.role+":"+tt.resource+":"+tt.action, func(t *testing.T) {
            result := ps.HasPermission(tt.role, tt.resource, tt.action)
            if result != tt.expected {
                t.Errorf("Expected %v, got %v", tt.expected, result)
            }
        })
    }
}
```

**Frontend Tests**:
```typescript
import { renderHook } from '@testing-library/react';
import { usePermissions } from '@/hooks/use-permissions';

describe('usePermissions', () => {
  it('should return true for admin with any permission', () => {
    // Mock session with admin role
    const { result } = renderHook(() => usePermissions());
    expect(result.current.hasPermission('requisition', 'create')).toBe(true);
  });

  it('should return true for requester creating requisition', () => {
    // Mock session with requester role
    const { result } = renderHook(() => usePermissions());
    expect(result.current.hasPermission('requisition', 'create')).toBe(true);
  });

  it('should return false for requester creating budget', () => {
    const { result } = renderHook(() => usePermissions());
    expect(result.current.hasPermission('budget', 'create')).toBe(false);
  });
});
```

**Checklist**:
- [ ] Backend permission tests passing
- [ ] Frontend permission tests passing
- [ ] All roles tested
- [ ] Edge cases covered

---

#### Task 3C.3: Integration Testing
**Time**: 45 minutes
**File**: `docs/PHASE3-INTEGRATION-TESTING.md` (new)

**Test Scenarios**:
1. **Backend API Permission Checks**
   - Create requisition (requester with permission ✓)
   - Create requisition (finance without permission ✗)
   - Approve requisition (approver with permission ✓)
   - Approve requisition (requester without permission ✗)

2. **Frontend Permission Guards**
   - Permission guard shows component when allowed
   - Permission guard hides component when denied
   - usePermissions hook works correctly

3. **Cross-Layer Consistency**
   - Frontend permissions match backend permissions
   - Backend and frontend agree on access

**Test Execution**:
- [ ] All backend API tests passing
- [ ] All frontend component tests passing
- [ ] Cross-layer consistency verified
- [ ] No regressions from Phase 2

---

### Phase 3D: Documentation (30 minutes)

#### Task 3D.1: Create Implementation Guide
**File**: `docs/PHASE3-IMPLEMENTATION-GUIDE.md`

**Contains**:
- How to add new permissions
- How to check permissions in handlers
- How to use permission guards in components
- How to add permission checks to routes
- Examples and best practices

#### Task 3D.2: Update Main Documentation
**File**: `docs/PHASE3-COMPLETION-SUMMARY.md`

**Contains**:
- What was implemented
- How permissions work
- How to test
- Migration guide (old role checks → new permission checks)

---

## 🎯 Success Criteria

### Backend ✅
- [ ] PermissionService properly checks all permissions
- [ ] RequirePermission middleware works
- [ ] All handlers use permission checks
- [ ] Routes use permission middleware (or inline checks)
- [ ] Unit tests for permissions passing (100% coverage)
- [ ] Permission denied returns 403 Forbidden
- [ ] Error messages clear

### Frontend ✅
- [ ] usePermissions hook works correctly
- [ ] PermissionGuard component renders conditionally
- [ ] All role checks replaced with permissions
- [ ] Components respect permissions
- [ ] Unit tests passing
- [ ] No regressions from Phase 2

### Integration ✅
- [ ] Backend and frontend permissions aligned
- [ ] End-to-end tests verify access control
- [ ] All existing features work with new permissions
- [ ] No breaking changes (backward compatible)
- [ ] Documentation complete

### Code Quality ✅
- [ ] Follows established patterns
- [ ] Well-documented and commented
- [ ] Type-safe (TypeScript/Go)
- [ ] No unused code
- [ ] Consistent naming

---

## 📅 Implementation Timeline

| Phase | Task | Est. Time | Notes |
|-------|------|-----------|-------|
| 3A | Create PermissionService | 45 min | Core business logic |
| 3A | Create RequirePermission middleware | 45 min | Route protection |
| 3A | Update handlers with permissions | 30 min | Applies everywhere |
| **3A Total** | | **2 hours** | |
| 3B | Create usePermissions hook | 45 min | Frontend utilities |
| 3B | Create PermissionGuard components | 45 min | Conditional rendering |
| 3B | Update components to use permissions | 30 min | UI consistency |
| **3B Total** | | **2 hours** | |
| 3C | Permission mapping document | 30 min | Reference guide |
| 3C | Unit tests (backend + frontend) | 45 min | Quality assurance |
| 3C | Integration testing | 45 min | Cross-layer validation |
| **3C Total** | | **2 hours** | |
| 3D | Implementation guide | 20 min | How-to docs |
| 3D | Completion summary | 10 min | Summary |
| **3D Total** | | **30 min** | |
| **TOTAL** | | **6.5 hours** | Flexible based on scope |

---

## 🔄 Implementation Order

**Recommended sequence:**

1. **Start Backend** (2 hours)
   - Create PermissionService
   - Create RequirePermission middleware
   - Update handlers

2. **Run Backend Tests** (30 min)
   - Verify all permissions working
   - Test middleware

3. **Start Frontend** (2 hours)
   - Create usePermissions hook
   - Create PermissionGuard components
   - Update components

4. **Run Frontend Tests** (30 min)
   - Verify all permissions working
   - Test components

5. **Integration Testing** (1 hour)
   - Backend + Frontend together
   - End-to-end flows

6. **Documentation** (30 min)
   - Update all guides
   - Create summary

---

## 🚀 Future Enhancements

### Phase 4 (Post Phase 3)
1. **Database-Driven Permissions**
   - Store permissions in database
   - Allow custom permissions per organization
   - Role inheritance and composition

2. **Fine-Grained ACL**
   - Object-level permissions (user can edit their own requisition only)
   - Attribute-based access control (ABAC)
   - Policy-based access control (PBAC)

3. **Audit Trail**
   - Log all permission checks
   - Track permission grants/revokes
   - Generate access reports

4. **Dynamic Permissions**
   - Permissions change based on context
   - Temporary permission elevation
   - Time-based permissions

---

## 📝 Key Design Decisions

### 1. Hardcoded vs Database Permissions
**Decision**: Start with hardcoded (MVP), move to database in Phase 4
**Rationale**: Faster to implement, easier to test, sufficient for current needs

### 2. Role-Based vs Permission-Based
**Decision**: Keep roles as the primary concept, use permissions internally
**Rationale**: Users understand roles, permissions are implementation detail

### 3. Service vs Middleware Approach
**Decision**: Both - PermissionService for business logic, middleware for routes
**Rationale**: Flexibility - can use either approach depending on context

### 4. Frontend vs Backend Permission Checks
**Decision**: Check on both sides
**Rationale**: Frontend for UX, backend for security (frontend can be bypassed)

---

## 🔒 Security Considerations

### Backend Security
- [ ] Always check permissions server-side (never trust frontend)
- [ ] Check at route level and handler level (defense in depth)
- [ ] Return 403 for permission denied (not 404 - don't leak info)
- [ ] Log permission denials for audit trail

### Frontend Security
- [ ] Permission checks are for UX only
- [ ] Hiding buttons doesn't prevent access (user can call API directly)
- [ ] Never store sensitive data in permission decisions
- [ ] Always validate backend response

### Testing Security
- [ ] Test unauthorized access attempts
- [ ] Test with each role
- [ ] Test permission boundary cases
- [ ] Test missing permission scenarios

---

## 🔗 Dependencies

### From Phase 1-2
- ✅ Authentication working
- ✅ Multi-tenancy context available
- ✅ Role system in place
- ✅ Organization membership verified

### Required for Phase 3
- None - Phase 3 builds on Phase 2

### Enables Future Phases
- Phase 4: Advanced permission management
- Future: Custom roles per organization

---

## ⚠️ Risk Mitigation

### Risk: Breaking Existing Authorization
**Mitigation**:
- Keep old role checks initially (parallel implementation)
- Gradually migrate one handler at a time
- Comprehensive testing before removing old checks
- Easy rollback if issues found

### Risk: Frontend/Backend Permission Mismatch
**Mitigation**:
- Keep permissions in sync between frontend and backend
- Unit tests verify mapping
- Integration tests verify behavior
- Documentation clearly outlines changes

### Risk: Performance Impact
**Mitigation**:
- Permission checks are lightweight (in-memory lookups)
- No database queries required
- Memoized in frontend (useCallback)
- Monitor performance post-implementation

---

## 📋 Deployment Strategy

### Pre-Deployment
- [ ] All unit tests passing
- [ ] All integration tests passing
- [ ] Code review completed
- [ ] Documentation updated

### Deployment Steps
1. Deploy backend changes (PermissionService + handlers)
2. Deploy frontend changes (hooks + components)
3. Run smoke tests
4. Monitor logs for errors
5. Verify with manual E2E testing

### Post-Deployment
- Monitor permission denied errors (403 responses)
- Check for any authorization regressions
- Verify performance metrics
- Collect user feedback

### Rollback Plan
- If critical issue found, revert both backend and frontend
- Keep old role-checking code commented (for quick restoration)
- Have database backup ready

---

## 📞 Support & Troubleshooting

### Common Issues

**Issue**: Permission denied (403) when should be allowed
- Check: Permission name matches exactly (case-sensitive)
- Check: Role correctly assigned to user
- Check: Frontend and backend permissions aligned
- Check: Organization membership verified

**Issue**: Permission allowed when should be denied
- Check: Admin role accidentally assigned
- Check: Fallback to true logic error
- Check: Test with correct role

**Issue**: Performance degradation
- Check: Permission checks not in loops
- Check: useCallback properly implemented
- Check: No N+1 permission queries

---

## ✨ Summary

Phase 3 transitions from role-based to permission-based access control, enabling:
- ✅ Fine-grained authorization
- ✅ More maintainable code
- ✅ Foundation for advanced features
- ✅ Better security model
- ✅ Easier to audit and modify

This is a significant architectural improvement that positions the system for future scalability and flexibility.

---

**Status**: Ready for Implementation
**Predecessor**: Phase 2 ✅ Complete
**Next Phase**: Phase 4 - Advanced Permission Management
**Estimated Start**: After Phase 2 validation complete

