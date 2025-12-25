# Phase 3: Permission-Based Access Control - Roadmap

**Status**: 📋 **PLANNED - READY FOR IMPLEMENTATION**
**Duration**: 4-6 hours
**Complexity**: Medium-High
**Prerequisite**: Phase 2 ✅ Complete

---

## 🎯 What is Phase 3?

**Transform authorization from:**
```
Role-Based          →      Permission-Based
if (role == "requester")  if (hasPermission("requisition:create"))
```

---

## 📊 Quick Overview

### Before Phase 3
```go
// Backend - Tight coupling to roles
if role == "requester" || role == "admin" {
    // Create requisition
}

// Frontend - Multiple conditions
{userRole === 'requester' || userRole === 'admin' && <CreateButton />}
```

### After Phase 3
```go
// Backend - Decoupled from role names
if permService.HasPermission(role, "requisition", "create") {
    // Create requisition
}

// Frontend - Clear intent
<PermissionGuard resource="requisition" action="create">
    <CreateButton />
</PermissionGuard>
```

---

## 🏗️ Architecture Overview

```
┌─────────────────────────────────────────────────────┐
│                   FRONTEND                           │
├─────────────────────────────────────────────────────┤
│ usePermissions Hook                                  │
│ ├─ hasPermission(resource, action)                  │
│ ├─ hasAllPermissions([...])                         │
│ └─ hasAnyPermission([...])                          │
│                                                      │
│ PermissionGuard Component                           │
│ └─ Conditional rendering based on permission        │
│                                                      │
│ All Components                                       │
│ └─ Use permissions instead of role checks           │
└─────────────────────────────────────────────────────┘
                        ↓↑ API Calls
┌─────────────────────────────────────────────────────┐
│                   BACKEND                            │
├─────────────────────────────────────────────────────┤
│ PermissionService                                   │
│ ├─ RolePermissions mapping (Go)                     │
│ ├─ HasPermission(role, resource, action)            │
│ └─ GetRolePermissions(role)                         │
│                                                      │
│ RequirePermission Middleware                        │
│ ├─ Check user permission                            │
│ ├─ Return 403 if denied                             │
│ └─ Call next handler if allowed                     │
│                                                      │
│ Route Definitions                                   │
│ └─ POST /requisitions - RequirePermission(...)      │
│                                                      │
│ All Handlers                                        │
│ └─ Use permissions instead of role checks           │
└─────────────────────────────────────────────────────┘
```

---

## 📋 4 Phases of Implementation

### Phase 3A: Backend Core (2 hours)
```
┌──────────────────────────┐
│ PermissionService        │ 45 min
│ • Hardcoded mapping      │
│ • HasPermission method   │
│ • GetRolePermissions     │
└──────────────────────────┘
           ↓
┌──────────────────────────┐
│ RequirePermission        │ 45 min
│ Middleware               │
│ • Check permission       │
│ • Return 403 if denied   │
│ • Continue if allowed    │
└──────────────────────────┘
           ↓
┌──────────────────────────┐
│ Update Handlers          │ 30 min
│ • CreateRequisition      │
│ • ApproveRequisition     │
│ • ManageBudget           │
│ • etc...                 │
└──────────────────────────┘
```

### Phase 3B: Frontend Implementation (2 hours)
```
┌──────────────────────────┐
│ usePermissions Hook      │ 45 min
│ • hasPermission()        │
│ • hasAllPermissions()    │
│ • hasAnyPermission()     │
│ • getPermissions()       │
└──────────────────────────┘
           ↓
┌──────────────────────────┐
│ PermissionGuard          │ 45 min
│ Components               │
│ • <PermissionGuard>      │
│ • <MultiPermissionGuard> │
│ • Conditional render     │
└──────────────────────────┘
           ↓
┌──────────────────────────┐
│ Update Components        │ 30 min
│ • RequisitionList        │
│ • RequisitionForm        │
│ • BudgetManager          │
│ • etc...                 │
└──────────────────────────┘
```

### Phase 3C: Integration & Testing (2 hours)
```
┌──────────────────────────┐
│ Permission Mapping Doc   │ 30 min
│ • All permissions listed │
│ • Resource + Action      │
│ • Role assignments       │
└──────────────────────────┘
           ↓
┌──────────────────────────┐
│ Unit Tests               │ 45 min
│ • Backend tests          │
│ • Frontend tests         │
│ • 100% coverage          │
└──────────────────────────┘
           ↓
┌──────────────────────────┐
│ Integration Testing      │ 45 min
│ • Backend + Frontend     │
│ • End-to-end flows      │
│ • Permission boundaries  │
└──────────────────────────┘
```

### Phase 3D: Documentation (30 min)
```
┌──────────────────────────┐
│ Implementation Guide     │ 20 min
│ • How to add permissions │
│ • Examples & patterns    │
│ • Best practices         │
└──────────────────────────┘
           ↓
┌──────────────────────────┐
│ Completion Summary       │ 10 min
│ • What was implemented   │
│ • Success criteria met   │
│ • Next steps             │
└──────────────────────────┘
```

---

## 📊 Permission Matrix Example

```
Resource: requisition

                    Requester  Approver  Finance  Viewer  Admin
requisition:create     ✅         ✅        ❌       ❌      ✅
requisition:read       ✅         ✅        ✅       ✅      ✅
requisition:update     ✅         ❌        ❌       ❌      ✅
requisition:approve    ❌         ✅        ❌       ❌      ✅
requisition:reject     ❌         ✅        ❌       ❌      ✅

Resource: budget

                    Requester  Approver  Finance  Viewer  Admin
budget:create          ❌         ❌        ✅       ❌      ✅
budget:read            ❌         ❌        ✅       ✅      ✅
budget:update          ❌         ❌        ✅       ❌      ✅

Resource: organization

                    Requester  Approver  Finance  Viewer  Admin
org:add_member         ❌         ❌        ❌       ❌      ✅
org:remove_member      ❌         ❌        ❌       ❌      ✅
org:manage_roles       ❌         ❌        ❌       ❌      ✅
```

---

## 💻 Code Examples

### Backend - PermissionService
```go
permService := services.NewPermissionService()
if !permService.HasPermission(userRole, "requisition", "create") {
    return c.Status(fiber.StatusForbidden).JSON(...)
}
```

### Backend - Middleware
```go
app.Post("/requisitions",
    middleware.RequirePermission("requisition", "create"),
    handlers.CreateRequisition,
)
```

### Frontend - Hook
```typescript
const { hasPermission } = usePermissions();

if (hasPermission('requisition', 'create')) {
    // Show create button
}
```

### Frontend - Component Guard
```typescript
<PermissionGuard resource="requisition" action="create">
    <button>Create Requisition</button>
</PermissionGuard>
```

---

## ✅ Success Metrics

| Category | Metric | Target |
|----------|--------|--------|
| **Code Coverage** | Unit tests | 100% |
| **Permission Checks** | Handlers using permissions | 100% |
| **Component Updates** | Components using guards | 100% |
| **Documentation** | Complete and current | ✅ |
| **Test Results** | All passing | ✅ |
| **Performance** | No regression | ✅ |

---

## 🔄 Dependencies & Prerequisites

### Must Have (From Phase 1-2)
- ✅ Authentication working
- ✅ Multi-tenancy implemented
- ✅ Role system in place
- ✅ Organization membership verified

### Creates Foundation For
- Phase 4: Custom permissions per organization
- Future: Advanced RBAC features
- Future: Audit logging
- Future: Dynamic permissions

---

## 🚀 Recommended Implementation Flow

```
Day 1 (Morning) - Backend
  └─ PermissionService (45 min)
  └─ RequirePermission middleware (45 min)
  └─ Update handlers (30 min)
  └─ Backend tests (45 min)

Day 1 (Afternoon) - Frontend
  └─ usePermissions hook (45 min)
  └─ PermissionGuard components (45 min)
  └─ Update components (30 min)
  └─ Frontend tests (45 min)

Day 2 (Morning) - Integration
  └─ Integration testing (45 min)
  └─ Fix any issues (30 min)
  └─ E2E testing (30 min)

Day 2 (Afternoon) - Wrap-up
  └─ Documentation (30 min)
  └─ Final testing (30 min)
  └─ Deploy to staging (30 min)
```

---

## 📚 Documentation Files

After Phase 3 completion, you'll have:

1. **PHASE3-IMPLEMENTATION-PLAN.md** ← You are here
   - Detailed task breakdown
   - Code examples
   - Success criteria

2. **PHASE3-PERMISSION-MAPPING.md**
   - All permissions listed
   - Role assignments
   - Resource/action matrix

3. **PHASE3-IMPLEMENTATION-GUIDE.md**
   - How to add new permissions
   - How to use in handlers
   - How to use in components

4. **PHASE3-INTEGRATION-TESTING.md**
   - Test scenarios
   - Test cases
   - Verification checklist

5. **PHASE3-COMPLETION-SUMMARY.md**
   - What was implemented
   - Success criteria met
   - Next steps

---

## 🎯 Key Outcomes

### Code Quality
✅ **Cleaner Authorization Logic**
- Remove multiple role checks
- Replace with single permission check
- More maintainable code

✅ **Better Separation of Concerns**
- Business logic separated from auth
- Permission service handles all checks
- Guards handle UI rendering

### Architecture
✅ **More Flexible**
- Easy to add new permissions
- Easy to change role permissions
- Foundation for future enhancements

✅ **More Secure**
- Centralized permission checking
- Easier to audit authorization
- Consistent across all handlers

### Development
✅ **Easier to Extend**
- Adding new feature = define permission + use in component/handler
- Changing access = update permission mapping
- Testing = verify permission checks

---

## ⚡ Quick Facts

| Aspect | Details |
|--------|---------|
| **Total Time** | 4-6 hours |
| **Backend Work** | ~2 hours |
| **Frontend Work** | ~2 hours |
| **Testing** | ~1 hour |
| **New Files** | ~6 files |
| **Modified Files** | ~15 files |
| **Lines of Code** | ~500-700 lines |
| **Breaking Changes** | None (backward compatible) |
| **Database Changes** | None (not yet) |

---

## 🔐 Security Benefits

✅ **Centralized Authorization**
- All permission checks in one place
- Consistent enforcement
- Easier to audit

✅ **Defense in Depth**
- Frontend guards (UX)
- Route middleware (first defense)
- Handler checks (second defense)

✅ **Clear Permission Model**
- What can do what is explicit
- Easier to find security issues
- Easier to add security tests

---

## 📈 Future Enhancements (Phase 4+)

### Immediate Next Steps
1. Store permissions in database
2. Allow custom permissions per organization
3. Add role inheritance
4. Support department-based access

### Long Term Vision
1. Attribute-based access control (ABAC)
2. Policy-based access control (PBAC)
3. Time-based permissions
4. Temporary elevation
5. Fine-grained object permissions

---

## 🎓 Learning Outcomes

After Phase 3, you'll understand:

✅ **Permission-Based Access Control (PBAC)**
- How it differs from RBAC
- When to use each approach
- Trade-offs and benefits

✅ **Middleware Patterns**
- How to create reusable middleware
- Composition and chaining
- Error handling

✅ **Frontend Authorization**
- Client-side permission checks
- Conditional rendering patterns
- UX for restricted features

✅ **Testing Authorization**
- Unit testing permissions
- Integration testing workflows
- Security testing

---

## 📞 Getting Help

### Questions During Implementation?

1. **Review** PHASE3-IMPLEMENTATION-PLAN.md (detailed guide)
2. **Check** PHASE3-PERMISSION-MAPPING.md (what permissions exist)
3. **Look** at examples in this roadmap
4. **Reference** Phase 1-2 patterns for consistency

### Common Patterns

**Adding New Permission**:
1. Define in RolePermissions map
2. Update permission mapping doc
3. Use in handlers
4. Use in components
5. Add tests

**Checking Permission**:
- Backend: `if !permService.HasPermission(...)`
- Frontend: `<PermissionGuard>` or `usePermissions()`

---

## ✨ Summary

**Phase 3 transforms the authorization system from role-based to permission-based**, enabling:

- ✅ More flexible access control
- ✅ Cleaner, more maintainable code
- ✅ Better security model
- ✅ Foundation for advanced features
- ✅ Easier to extend and modify

**Ready to implement?** Start with PHASE3-IMPLEMENTATION-PLAN.md

---

**Status**: 📋 **PLANNED - READY FOR IMPLEMENTATION**
**Estimated Duration**: 4-6 hours
**Recommended Start**: After Phase 2 validation complete
**Expected Completion**: Within 1 day of full-time work

---

*For detailed implementation steps, see [PHASE3-IMPLEMENTATION-PLAN.md](PHASE3-IMPLEMENTATION-PLAN.md)*

