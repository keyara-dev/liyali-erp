# RBAC System

**Role-Based Access Control with 7 roles and 35+ permissions**

---

## Overview

Liyali Gateway implements a comprehensive Role-Based Access Control (RBAC) system with 7 user roles and 35+ granular permissions. Each role has specific capabilities aligned with organizational hierarchy.

---

## User Roles

### 1. ADMIN
**Full system access**

- Manage all users
- Configure workflows
- Access all documents
- View all analytics
- Delete records
- System configuration

**Typical Users**: System administrators, IT staff

---

### 2. CFO (Chief Financial Officer)
**Executive-level approvals**

- Approve at Stage 3 (final approval)
- View all documents
- Access financial analytics
- Reassign tasks
- Reject requests

**Typical Users**: Chief Financial Officer, Finance Director

---

### 3. DIRECTOR
**Senior management approvals**

- Approve at Stage 3
- View departmental documents
- Access analytics
- Reassign tasks
- Reject requests

**Typical Users**: Department Directors, Senior Managers

---

### 4. FINANCE_OFFICER
**Financial processing and validation**

- Approve at Stage 2
- Process payments
- Verify budgets
- View financial documents
- Reassign tasks

**Typical Users**: Finance Officers, Accountants

---

### 5. DEPARTMENT_MANAGER
**Departmental approvals**

- Approve at Stage 1
- Submit requests
- View departmental documents
- Reassign within department
- Reject requests

**Typical Users**: Department Managers, Team Leads

---

### 6. COMPLIANCE_OFFICER
**Audit and compliance monitoring**

- View all documents (read-only)
- Access audit logs
- Generate compliance reports
- View analytics
- **Cannot approve/reject**

**Typical Users**: Compliance Officers, Auditors

---

### 7. REQUESTER
**Basic request submission**

- Submit new requests
- View own requests
- View request status
- **Cannot approve/reject**
- **Cannot view others' requests**

**Typical Users**: Employees, Staff members

---

## Permission Matrix

| Permission | ADMIN | CFO | DIRECTOR | FINANCE | MANAGER | COMPLIANCE | REQUESTER |
|-----------|-------|-----|----------|---------|---------|------------|-----------|
| **Approvals** |
| Approve Stage 1 | ✓ | ✓ | ✓ | ✓ | ✓ | ✗ | ✗ |
| Approve Stage 2 | ✓ | ✓ | ✓ | ✓ | ✗ | ✗ | ✗ |
| Approve Stage 3 | ✓ | ✓ | ✓ | ✗ | ✗ | ✗ | ✗ |
| Reject Tasks | ✓ | ✓ | ✓ | ✓ | ✓ | ✗ | ✗ |
| Reassign Tasks | ✓ | ✓ | ✓ | ✓ | ✓ | ✗ | ✗ |
| Bulk Approve | ✓ | ✓ | ✓ | ✓ | ✓ | ✗ | ✗ |
| **Documents** |
| Create Documents | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| View All Documents | ✓ | ✓ | ✓ | ✗ | ✗ | ✓ | ✗ |
| View Own Documents | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| Update Documents | ✓ | ✓ | ✓ | ✓ | ✓ | ✗ | ✓* |
| Delete Documents | ✓ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ |
| Export Documents | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| **Workflows** |
| Create Workflows | ✓ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ |
| Update Workflows | ✓ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ |
| Delete Workflows | ✓ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ |
| View Workflows | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| **Users** |
| Create Users | ✓ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ |
| Update Users | ✓ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ |
| Deactivate Users | ✓ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ |
| View All Users | ✓ | ✓ | ✓ | ✗ | ✗ | ✓ | ✗ |
| **Analytics** |
| View Analytics | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✗ |
| View Reports | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✗ |
| **Audit** |
| View Audit Logs | ✓ | ✓ | ✗ | ✗ | ✗ | ✓ | ✗ |

*Requester can only update own documents before submission

---

## Permission List

### Approval Permissions
```go
const (
    PermApproveStage1        = "approval:approve:stage1"
    PermApproveStage2        = "approval:approve:stage2"
    PermApproveStage3        = "approval:approve:stage3"
    PermReject               = "approval:reject"
    PermReassign             = "approval:reassign"
    PermBulkApprove          = "approval:bulk_approve"
    PermBulkReject           = "approval:bulk_reject"
    PermViewAllApprovals     = "approval:view_all"
    PermViewOwnApprovals     = "approval:view_own"
)
```

### Document Permissions
```go
const (
    PermCreateDocument       = "document:create"
    PermViewAllDocuments     = "document:view_all"
    PermViewOwnDocuments     = "document:view_own"
    PermUpdateDocument       = "document:update"
    PermDeleteDocument       = "document:delete"
    PermExportDocument       = "document:export"
)
```

### Workflow Permissions
```go
const (
    PermCreateWorkflow       = "workflow:create"
    PermUpdateWorkflow       = "workflow:update"
    PermDeleteWorkflow       = "workflow:delete"
    PermViewWorkflow         = "workflow:view"
)
```

### User Permissions
```go
const (
    PermCreateUser           = "user:create"
    PermUpdateUser           = "user:update"
    PermDeactivateUser       = "user:deactivate"
    PermViewAllUsers         = "user:view_all"
    PermViewOwnProfile       = "user:view_own"
)
```

### Analytics Permissions
```go
const (
    PermViewAnalytics        = "analytics:view"
    PermViewReports          = "analytics:reports"
    PermExportAnalytics      = "analytics:export"
)
```

### Audit Permissions
```go
const (
    PermViewAuditLogs        = "audit:view"
    PermExportAuditLogs      = "audit:export"
)
```

---

## Middleware Usage

### Require Authentication

Protect any endpoint with JWT authentication:

```go
// Protect route - requires valid JWT token
auth.Get("/me", authMiddleware.Authenticate, authHandler.GetCurrentUser)
```

### Require Specific Permission

```go
// Only users with approval permission can approve
app.Post("/api/approvals/:id/approve",
    authMiddleware.Authenticate,
    rbacMiddleware.RequirePermission(rbac.PermApproveStage1),
    approvalHandler.Approve,
)
```

### Require Any Permission

```go
// User needs at least one of these permissions
app.Get("/api/documents",
    authMiddleware.Authenticate,
    rbacMiddleware.RequireAnyPermission(
        rbac.PermViewAllDocuments,
        rbac.PermViewOwnDocuments,
    ),
    documentHandler.List,
)
```

### Require All Permissions

```go
// User needs all these permissions
app.Delete("/api/workflows/:id",
    authMiddleware.Authenticate,
    rbacMiddleware.RequireAllPermissions(
        rbac.PermDeleteWorkflow,
        rbac.PermViewWorkflow,
    ),
    workflowHandler.Delete,
)
```

### Require Specific Role

```go
// Only admins can create users
app.Post("/api/users",
    authMiddleware.Authenticate,
    rbacMiddleware.RequireRole(rbac.RoleAdmin),
    userHandler.Create,
)
```

### Require Any Role

```go
// CFO or Director can approve stage 3
app.Post("/api/approvals/:id/approve-stage3",
    authMiddleware.Authenticate,
    rbacMiddleware.RequireAnyRole(rbac.RoleCFO, rbac.RoleDirector),
    approvalHandler.ApproveStage3,
)
```

---

## Programmatic Permission Checks

### Check in Handler

```go
func (h *DocumentHandler) Update(c fiber.Ctx) error {
    userID, _ := middleware.GetUserID(c)
    role, _ := middleware.GetUserRole(c)
    documentID := c.Params("id")

    // Get document
    doc, err := h.documentService.GetByID(c.Context(), documentID)
    if err != nil {
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
            "error": "Document not found",
        })
    }

    // Check permission
    if role != rbac.RoleAdmin && doc.CreatedBy != userID {
        return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
            "error": "Permission denied: can only update own documents",
        })
    }

    // Continue with update...
}
```

### Check Permission

```go
// Check if user has permission
if rbac.HasPermission(role, rbac.PermApproveStage2) {
    // Allow approval
}

// Check if user has any permission
if rbac.HasAnyPermission(role, rbac.PermViewAllDocuments, rbac.PermViewOwnDocuments) {
    // Allow view
}
```

---

## Role Assignment

### During Registration

```go
// User specifies role during registration
POST /api/auth/register
{
  "email": "manager@example.com",
  "password": "SecurePass123!",
  "name": "John Manager",
  "role": "DEPARTMENT_MANAGER",  // Must be valid role
  "department": "Finance"
}
```

### Role Validation

```go
// Valid roles
var validRoles = []string{
    "ADMIN",
    "CFO",
    "DIRECTOR",
    "FINANCE_OFFICER",
    "DEPARTMENT_MANAGER",
    "COMPLIANCE_OFFICER",
    "REQUESTER",
}

// Validate role
func IsValidRole(role string) bool {
    for _, validRole := range validRoles {
        if role == validRole {
            return true
        }
    }
    return false
}
```

---

## Implementation Examples

### Example 1: Approval with Stage Check

```go
func (h *ApprovalHandler) Approve(c fiber.Ctx) error {
    role, _ := middleware.GetUserRole(c)
    taskID := c.Params("id")

    // Get task to determine stage
    task, _ := h.approvalService.GetTask(c.Context(), taskID)

    // Check permission based on stage
    var hasPermission bool
    switch task.CurrentStage {
    case 1:
        hasPermission = rbac.HasPermission(role, rbac.PermApproveStage1)
    case 2:
        hasPermission = rbac.HasPermission(role, rbac.PermApproveStage2)
    case 3:
        hasPermission = rbac.HasPermission(role, rbac.PermApproveStage3)
    }

    if !hasPermission {
        return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
            "error": "Permission denied for this stage",
        })
    }

    // Proceed with approval...
}
```

### Example 2: Document Access Control

```go
func (h *DocumentHandler) List(c fiber.Ctx) error {
    userID, _ := middleware.GetUserID(c)
    role, _ := middleware.GetUserRole(c)

    var documents []db.Document

    if rbac.HasPermission(role, rbac.PermViewAllDocuments) {
        // Can view all documents
        documents, _ = h.documentService.ListAll(c.Context())
    } else if rbac.HasPermission(role, rbac.PermViewOwnDocuments) {
        // Can only view own documents
        documents, _ = h.documentService.ListByUser(c.Context(), userID)
    } else {
        return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
            "error": "Permission denied",
        })
    }

    return c.JSON(documents)
}
```

---

## Testing RBAC

### Unit Test

```go
func TestHasPermission(t *testing.T) {
    // ADMIN should have all permissions
    assert.True(t, rbac.HasPermission(rbac.RoleAdmin, rbac.PermApproveStage3))

    // REQUESTER should not have approval permission
    assert.False(t, rbac.HasPermission(rbac.RoleRequester, rbac.PermApproveStage1))

    // COMPLIANCE_OFFICER should view all but not approve
    assert.True(t, rbac.HasPermission(rbac.RoleComplianceOfficer, rbac.PermViewAllDocuments))
    assert.False(t, rbac.HasPermission(rbac.RoleComplianceOfficer, rbac.PermApproveStage1))
}
```

### Integration Test

```go
func TestApproveEndpoint_Forbidden(t *testing.T) {
    app := setupTestApp(t)
    defer teardownTestApp()

    // Register as REQUESTER
    user := registerUser(t, app, "requester@example.com", "password", "REQUESTER")

    // Login
    accessToken := login(t, app, "requester@example.com", "password")

    // Try to approve (should fail - no permission)
    req := httptest.NewRequest(http.MethodPost, "/api/approvals/123/approve", nil)
    req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

    resp, _ := app.Test(req)

    // Assert forbidden
    assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}
```

---

## Common Patterns

### Pattern 1: View All vs View Own

```go
// Check if user can view all or just own
if rbac.HasPermission(role, rbac.PermViewAllDocuments) {
    // Return all documents
    return allDocuments
} else if rbac.HasPermission(role, rbac.PermViewOwnDocuments) {
    // Filter to user's documents
    return filterByUser(allDocuments, userID)
} else {
    return forbiddenError
}
```

### Pattern 2: Hierarchical Permissions

```go
// Higher stages include lower stages
// If you can approve stage 3, you can approve stage 1 and 2
if rbac.HasPermission(role, rbac.PermApproveStage3) {
    // Can approve any stage
} else if rbac.HasPermission(role, rbac.PermApproveStage2) {
    // Can approve stage 1 and 2 only
} else if rbac.HasPermission(role, rbac.PermApproveStage1) {
    // Can approve stage 1 only
}
```

### Pattern 3: Admin Override

```go
// Admins bypass all checks
if role == rbac.RoleAdmin {
    // Allow operation
    return proceed()
}

// Otherwise check specific permission
if !rbac.HasPermission(role, requiredPermission) {
    return forbiddenError
}
```

---

## Related Pages

- [Authentication Flow](./authentication.md) - Login and JWT
- [User Management](./user-management.md) - Complete role guide
- [Security Best Practices](./security.md) - Additional security

---

**Files**:
- `internal/rbac/permissions.go` - Permission definitions and role mappings
- `internal/middleware/rbac_middleware.go` - Permission checking middleware

**Last Updated**: December 25, 2025
