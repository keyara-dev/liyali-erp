# Phase 3 Extended: Complete Authorization & Workflow System Roadmap

**Status**: 📋 PLANNED
**Scope**: Answers the key question: "Who creates roles and how do workflows work?"

---

## 🎯 The Complete Picture

You asked an excellent question: **"Who creates the roles and assigns permissions?"**

The answer spans multiple phases:

### Phase 3 (Now): Fixed Roles, Permission-Based Authorization
```
Hardcoded Roles:
├─ admin
├─ approver
├─ requester
├─ finance
└─ viewer

Hardcoded Permissions:
├─ requisition:create
├─ requisition:approve
├─ budget:manage
└─ ...
```

✅ **Admin** manages the system
❌ **Organization admins** cannot create custom roles
❌ **Organization admins** cannot customize permissions

---

### Phase 3.5 (Optional): Database-Driven Custom Roles
```
Organization Admins Can:
├─ Create custom roles ("Senior Manager", "Finance Approver")
├─ Assign permissions to roles
├─ Create workflows with stages
├─ Assign roles to workflow stages
└─ Modify everything via UI (no coding)

System Still Supports:
├─ Default system roles (backward compatible)
├─ User.role (global designation)
└─ OrganizationMember.role (org-specific)
```

✅ **Organization admins** can create roles
✅ **Organization admins** can assign permissions
✅ **Organization admins** can create workflows

---

### Phase 4: Advanced Features
```
Organization Admins Can Also:
├─ Role inheritance and composition
├─ Custom permission conditions
├─ Attribute-based access control (ABAC)
├─ Time-based temporary roles
├─ Custom approval workflows
└─ Bulk permission management
```

✅ **Full flexibility** for organization customization

---

## 🏗️ Complete Architecture

```
SYSTEM LAYER (Fixed, Global)
├─ User.role (admin, approver, requester, finance, viewer)
├─ System Permissions (hardcoded in Phase 3, in DB in Phase 3.5+)
└─ Super Admin Functions

ORGANIZATION LAYER (Configurable by Admin)
├─ OrganizationRole (e.g., "Senior Manager")
├─ OrganizationPermission (e.g., "requisition:approve")
├─ RolePermissionAssignment (role has permission)
└─ WorkflowStage (approval workflow stages)

DOCUMENT LAYER (Runtime)
├─ Document created
├─ Routed through WorkflowStages
├─ Each stage checks user role
├─ User must have required role to approve
└─ Document progresses or is rejected
```

---

## 🔄 Three Key Components

### 1. ROLES (Who is what?)
```
Before Phase 3.5:
  User has global role: "approver"
  User has org role: "approver" (string)
  ❌ Cannot create custom roles

After Phase 3.5:
  User has global role: "approver"
  User has org role: <OrganizationRole ID>
  ✅ Can have "Senior Manager" role in Org A
  ✅ Can have "Team Lead" role in Org B
  ✅ Can have custom "Vendor Approver" role
```

### 2. PERMISSIONS (What can they do?)
```
Before Phase 3.5:
  Role "approver" can: [requisition:approve, requisition:read, ...]
  Defined in code
  ❌ Cannot customize

After Phase 3.5:
  Role "Senior Manager" can: [requisition:approve, budget:approve, vendor:approve, ...]
  Defined in database
  ✅ Can customize per organization
  ✅ Can add/remove permissions anytime
```

### 3. WORKFLOWS (How do documents flow?)
```
Before Phase 3.5:
  Requisition has approvalStage: 1, 2, 3, ...
  Workflow is implicit
  ❌ Cannot customize

After Phase 3.5:
  Organization defines workflow:
    Stage 1: Manager approval (role: "Manager")
    Stage 2: Finance approval (role: "Finance Approver")
    Stage 3: Director approval (role: "Director", if amount > $50k)
  ✅ Can define in UI
  ✅ Can modify anytime
  ✅ Can add conditional stages (based on amount, department, etc.)
```

---

## 📊 Implementation Phases Detailed

### Phase 3: Permissions System (4-6 hours)
**Cost**: Low
**Risk**: Low
**Status**: Ready to implement now

**What gets done**:
```
Backend:
  ✅ PermissionService with hardcoded mapping
  ✅ RequirePermission middleware
  ✅ Handler permission checks
  ✅ Tests for all permissions

Frontend:
  ✅ usePermissions hook
  ✅ PermissionGuard components
  ✅ Component updates for permission checks
  ✅ Tests

Result: Cleaner, more maintainable authorization
```

**What doesn't change**:
- Roles still hardcoded (admin, approver, etc.)
- Organization admins cannot create roles
- Workflows are implicit (magic numbers)

---

### Phase 3.5: Custom Roles (Optional, Before Phase 4) (12-16 hours)
**Cost**: Medium
**Risk**: Medium
**Status**: Planned for after Phase 3

**Prerequisites**:
- Phase 3 complete and validated

**What gets done**:
```
Database:
  ✅ OrganizationRole table
  ✅ OrganizationPermission table
  ✅ RolePermission assignment table
  ✅ Migration script

Backend:
  ✅ Role CRUD endpoints
  ✅ Permission CRUD endpoints
  ✅ Permission assignment endpoints
  ✅ Update PermissionService to check DB
  ✅ Backward compatibility with hardcoded roles
  ✅ Tests

Frontend:
  ✅ Role management UI
  ✅ Permission assignment UI
  ✅ Show available roles when assigning members
  ✅ Tests

Result: Organization admins can create custom roles
```

**What changes**:
- Admins can create roles like "Senior Manager"
- Admins can assign permissions to roles
- Admins can assign users to custom roles

---

### Phase 3.5+: Workflow System (Optional, After Phase 3.5) (16-20 hours)
**Cost**: High
**Risk**: Medium-High
**Status**: Advanced feature, planned later

**Prerequisites**:
- Phase 3.5 complete (custom roles working)

**What gets done**:
```
Database:
  ✅ WorkflowTemplate table
  ✅ WorkflowStage table
  ✅ WorkflowApproval table
  ✅ Condition support (min amount, departments)

Backend:
  ✅ Workflow engine
  ✅ Stage routing logic
  ✅ Role requirement checking
  ✅ Condition evaluation
  ✅ Approval endpoints
  ✅ Rejection handling
  ✅ Tests

Frontend:
  ✅ Workflow builder UI
  ✅ Approval task board
  ✅ Stage visualization
  ✅ Condition builder
  ✅ Tests

Result: Organization admins can create approval workflows
```

**What changes**:
- Admins can create workflows with multiple stages
- Admins assign roles to stages
- Documents route through stages automatically
- Users see approval tasks personalized to them

---

### Phase 4: Advanced Features (20+ hours)
**Cost**: High
**Risk**: High
**Status**: Future enhancement

**What could be added**:
```
✅ Role inheritance and composition
✅ Dynamic permission conditions (if amount > X)
✅ Temporary role elevation
✅ Department-based access
✅ Attribute-based access control (ABAC)
✅ Custom approval logic
✅ Approval notifications
✅ Audit logging
✅ Workflow analytics
✅ Bulk operations
```

---

## 🔗 How They Connect

### Scenario: Organization Creates Workflow

**Step 1: Admin Creates Role** (Phase 3.5)
```
Admin: "Create role 'Finance Approver'"
System:
  - Creates OrganizationRole with name "Finance Approver"
  - ID: role_xyz
  - Organization: Org ABC
```

**Step 2: Admin Assigns Permissions** (Phase 3.5)
```
Admin: "Give 'Finance Approver' these permissions"
  - requisition:approve
  - budget:manage
System:
  - Creates RolePermission entries
  - Now: finance_approver can do these things
```

**Step 3: Admin Creates Workflow** (Phase 3.5+)
```
Admin: "Create 'Requisition Approval' workflow"
  Stage 1: Manager Review (role: Manager)
  Stage 2: Finance Review (role: Finance Approver)
  Stage 3: Director Approval (role: Director, if amount > $50k)
System:
  - Creates WorkflowTemplate
  - Creates WorkflowStages
  - Stores role requirements
```

**Step 4: User Creates Document**
```
User creates Requisition ($25,000)
System:
  - Looks up workflow for "requisition"
  - Evaluates stages (conditions met?)
  - Stage 1: Manager (✓)
  - Stage 2: Finance Approver (✓)
  - Stage 3: Director (✗ amount < $50k)
  - Routes to Stage 1
```

**Step 5: Manager Approves**
```
Manager reviews document
Manager: "Approve"
System:
  - Checks: Is manager a "Manager" role? ✓
  - Marks stage as approved
  - Routes to Stage 2 (Finance Approver)
```

**Step 6: Finance Approver Approves**
```
Finance Approver reviews document
Finance Approver: "Approve"
System:
  - Checks: Is user a "Finance Approver" role? ✓
  - Marks stage as approved
  - No more stages
  - Document status: APPROVED
```

---

## 📈 Feature Timeline

```
NOW (Phase 3 Ready)
  └─ Permission-based authorization
     └─ Cleaner code, hardcoded roles

Q1 (Phase 3.5 - Optional)
  └─ Custom roles
  └─ Custom permissions per org
  └─ Basic workflow support

Q2+ (Phase 4 - Advanced)
  └─ Advanced workflows
  └─ Complex approval logic
  └─ Full customization
```

---

## 🎯 Recommendation

### Immediate Implementation (Phase 3)
✅ **Do this now**: Permission-based authorization
- Clean up code
- Remove hardcoded role checks
- Use permission-based logic
- Easier to maintain

### Short Term (Phase 3.5)
✅ **Do this after Phase 3 validated**: Custom roles
- Allow org admins to create roles
- Better customization
- More flexible

### Medium Term (Phase 3.5+)
✅ **Do this after Phase 3.5 working**: Workflow system
- Custom approval workflows
- Multi-stage processes
- Powerful automation

### Long Term (Phase 4)
✅ **Do this as advanced feature**: Advanced customization
- Complex logic
- Rich automation
- Full flexibility

---

## 🏆 What Org Admins Can Do At Each Phase

### Phase 3 Now
```
❌ Create custom roles
❌ Assign permissions to roles
❌ Create approval workflows
❌ Customize document routing
```

### Phase 3.5 (After This)
```
✅ Create custom roles (e.g., "Senior Manager")
✅ Assign permissions to roles
❌ Create approval workflows (yet)
❌ Customize document routing (yet)
```

### Phase 3.5+ (After That)
```
✅ Create custom roles
✅ Assign permissions to roles
✅ Create approval workflows
✅ Customize document routing
✅ Set conditional stages
```

### Phase 4 (Future)
```
✅ Create custom roles
✅ Assign permissions to roles
✅ Create approval workflows
✅ Customize document routing
✅ Set conditional stages
✅ Create role inheritance
✅ Set dynamic conditions
✅ And much more...
```

---

## 📊 Effort vs Capability Chart

```
Hours    Capability
│
100 ├─ Phase 4: Advanced Customization ████████
    │
80  ├─ Phase 3.5+: Full Workflows ██████
    │
60  ├─ Phase 3.5: Custom Roles ████
    │
40  ├─ Phase 3: Permissions System ██
    │
20  ├─ Phase 2: Auto-Org ██
    │
0   └─────────────────────────────────
    Phase Phase Phase Phase Phase Phase
      1    2    3    3.5  3.5+ 4
```

---

## 📚 Related Documents

1. **PHASE3-IMPLEMENTATION-PLAN.md**
   - How to implement permission system
   - Code examples
   - Testing strategy

2. **PHASE3-ROLES-AND-PERMISSIONS-DESIGN.md**
   - Who creates roles
   - Custom role system design
   - Database schema
   - Phase 3.5 implementation

3. **PHASE3-WORKFLOW-STAGES-AND-ROLES.md**
   - How workflows work
   - Multi-stage approval
   - Admin creates workflows
   - Phase 3.5+ implementation

4. **PHASE3-ROADMAP.md**
   - Visual overview
   - Quick reference
   - Architecture diagrams

---

## ✨ Summary

**Your question: "Who creates the roles and assigns permissions?"**

**Answer**:
- **Phase 3 (Now)**: System admin (hardcoded)
- **Phase 3.5 (Soon)**: Organization admins (via UI)
- **Phase 4 (Future)**: Organization admins with advanced features

**With the workflow system**:
- Organization admins can create approval workflows
- Each workflow stage can require specific roles
- Documents route through stages automatically
- All without any coding!

---

**Next Steps**:
1. Implement Phase 3 (permission system) ← **You are here**
2. Validate Phase 3
3. Plan Phase 3.5 (custom roles)
4. Implement Phase 3.5
5. Plan Phase 3.5+ (workflows)
6. Implement Phase 3.5+

---

**Status**: Phase 3 ready to implement now. Phases 3.5 and beyond are planned and designed, ready to execute when needed.

