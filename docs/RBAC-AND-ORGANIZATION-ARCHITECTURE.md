# RBAC and Organization Architecture Design

## Overview

This document outlines how Role-Based Access Control (RBAC) and organization management work in the Liyali Gateway system, addressing the shift from `user_type` to `role` and permission-based access.

---

## 1. Core Architecture Change: user_type → role

### Previous Model (Deprecated)
- Used `user_type` field to determine user capabilities
- Hard-coded roles tied to user record globally
- Couldn't have different roles in different organizations
- Limited flexibility for organization-specific permissions

### New Model (Current)
- `role` field in User model = global role designation
- `OrganizationMember.role` = organization-specific role
- **Users can have different roles in different organizations**
- Permissions derived from both global role AND organization membership

### Implementation Details

**User Model** (`backend/models/models.go`):
```go
type User struct {
    ID                    string
    Email                 string
    Name                  string
    Role                  string  // Global role: admin, approver, requester, finance, viewer
    CurrentOrganizationID *string // Currently active organization
    IsSuperAdmin          bool    // Can operate across all orgs (future use)
}
```

**OrganizationMember Model** (`backend/models/organization.go`):
```go
type OrganizationMember struct {
    ID             string  // UUID
    OrganizationID string  // Which org
    UserID         string  // Which user
    Role           string  // Role in THIS organization
    Department     string  // Department within org
    Title          string  // Job title
    Active         bool    // Active membership
    JoinedAt       *time.Time
    CustomPermissions JSON // Future extensibility for custom permissions
}
```

**Effective Role Determination:**

When a request arrives with `X-Organization-ID` header:
1. Extract user from JWT token
2. Look up `OrganizationMember` for (user, organization) pair
3. Use `OrganizationMember.role` as the effective role for that organization
4. If user not in organization → deny access (403 Forbidden)

---

## 2. Permission Model: From Roles to Permissions

### Why Permissions Matter

Instead of checking `if role == "requester"`, we check `if has_permission("create_requisition")`.

**Benefits:**
- Decouples role names from capabilities
- Enables fine-grained access control
- Allows custom permissions per organization (via `CustomPermissions` JSON)
- Future-proof for complex workflows

### Permission Definition

**Permission Structure:**
```
{
  resource: "requisition",      // What entity/object
  action: "create|read|update|delete|approve",  // What action
  description: "Can create requisitions"
}
```

**Example Permissions Matrix:**

| Permission | Requester | Approver | Finance | Viewer | Admin |
|-----------|-----------|----------|---------|--------|-------|
| create_requisition | ✓ | ✓ | ✗ | ✗ | ✓ |
| view_requisition | ✓ | ✓ | ✓ | ✓ | ✓ |
| approve_requisition | ✗ | ✓ | ✗ | ✗ | ✓ |
| create_budget | ✗ | ✗ | ✓ | ✗ | ✓ |
| add_org_member | ✗ | ✗ | ✗ | ✗ | ✓ |
| manage_vendors | ✗ | ✗ | ✓ | ✗ | ✓ |

### Implementation Approach

**Option A: Hardcoded Role Permissions (Current - Recommended for MVP)**

```go
// backend/services/permissions.go
var RolePermissions = map[string][]string{
    "requester": {
        "create_requisition",
        "view_requisition",
        "create_draft",
    },
    "approver": {
        "view_requisition",
        "approve_requisition",
        "reject_requisition",
    },
    "finance": {
        "create_budget",
        "view_budget",
        "manage_vendors",
        "view_payment_voucher",
        "approve_payment_voucher",
    },
    "viewer": {
        "view_requisition",
        "view_budget",
        "view_reports",
    },
    "admin": {
        "*", // All permissions
    },
}

func HasPermission(userRole string, permission string) bool {
    if userRole == "admin" {
        return true
    }
    perms := RolePermissions[userRole]
    for _, p := range perms {
        if p == permission {
            return true
        }
    }
    return false
}
```

**Option B: Database-Driven Permissions (Future Enhancement)**

```go
type Permission struct {
    ID          string
    Resource    string
    Action      string
    Description string
}

type RolePermission struct {
    RoleID       string
    PermissionID string
}

// More flexible but requires database lookups
// Use caching to avoid N+1 queries
```

**Option C: Per-Organization Custom Permissions (Advanced)**

```go
type OrganizationMember struct {
    // ... other fields ...
    CustomPermissions json.RawMessage // Override permissions for this user
}

// Check custom perms first, fall back to role permissions
```

---

## 3. User Registration & Organization Onboarding Flow

### Challenge
When a new user registers, they have no organization membership. How do they get into an organization?

### Proposed Solution: Three Scenarios

#### Scenario A: Self-Service Organization Creation
New user creates their own organization upon signup.

**Flow:**
```
1. User signs up (email, password, name, role)
2. User created in database with default role (requester)
3. Redirect to "Create Organization" page
4. User creates new organization (name, slug, logo, etc.)
5. System automatically adds user as admin to new org
6. Set as CurrentOrganizationID
7. User can now access org-scoped resources
8. Later: Can invite others to organization
```

**Implementation:**

```typescript
// Frontend: frontend/src/app/(auth)/signup/_components/signup-form.tsx
export function SignupForm() {
  const handleSignup = async (formData) => {
    // Call backend /api/v1/auth/register
    const result = await registerAction({
      email: formData.email,
      password: formData.password,
      name: formData.name,
      role: "requester", // Default role
    });

    if (result.success) {
      // Redirect to organization creation
      router.push("/onboarding/create-organization");
    }
  };
}
```

```go
// Backend: handlers/auth.go - Register handler
func Register(c *fiber.Ctx) error {
    // ... existing validation ...

    newUser := models.User{
        ID:     utils.GenerateUserID(),
        Email:  req.Email,
        Name:   req.Name,
        Role:   "requester", // Default role - can be enhanced later
        Active: true,
    }

    // Note: NO currentOrgID yet - user will create/join org on next step
    config.DB.Create(&newUser)

    return c.JSON(types.SuccessResponse{
        Success: true,
        Message: "User registered successfully",
        Token:   token,
        User:    newUser,
    })
}
```

```go
// Backend: handlers/organization.go - Create Organization
func CreateOrganization(c *fiber.Ctx) error {
    userID := c.Locals("userID").(string)

    var req struct {
        Name  string `json:"name"`
        Slug  string `json:"slug"`
        Logo  string `json:"logo"`
    }
    c.BodyParser(&req)

    // Create org
    org := models.Organization{
        ID:    utils.GenerateOrgID(),
        Name:  req.Name,
        Slug:  req.Slug,
        Logo:  req.Logo,
    }
    config.DB.Create(&org)

    // Add creator as admin member
    member := models.OrganizationMember{
        ID:             utils.GenerateID(),
        OrganizationID: org.ID,
        UserID:         userID,
        Role:           "admin", // Creator is admin
        Active:         true,
        JoinedAt:       time.Now(),
    }
    config.DB.Create(&member)

    // Set as current organization
    config.DB.Model(&models.User{}).
        Where("id = ?", userID).
        Update("current_organization_id", org.ID)

    return c.JSON(org)
}
```

---

#### Scenario B: Organization Invitation Link
Admin invites new user via email with pre-filled organization.

**Flow:**
```
1. Org admin clicks "Invite Member"
2. Admin enters email and selects role
3. System generates invite token (JWT with org + email + role)
4. Email sent with signup link: /signup?token=xxx
5. New user clicks link, pre-filled form shows org + role
6. User signs up
7. System verifies token, auto-adds user to organization
8. User can immediately access org resources
```

**Implementation:**

```go
// Backend: handlers/organization.go - Create Invite
func InviteMember(c *fiber.Ctx) error {
    userID := c.Locals("userID").(string)
    orgID := c.Locals("organizationId").(string)

    var req struct {
        Email string `json:"email"`
        Role  string `json:"role"`
    }
    c.BodyParser(&req)

    // Create invite token (valid for 7 days)
    inviteToken, _ := utils.GenerateInviteToken(orgID, req.Email, req.Role)

    // Store invite in database
    invite := models.OrganizationInvite{
        ID:             utils.GenerateID(),
        OrganizationID: orgID,
        Email:          req.Email,
        Role:           req.Role,
        Token:          inviteToken,
        ExpiresAt:      time.Now().AddDate(0, 0, 7),
        CreatedBy:      userID,
    }
    config.DB.Create(&invite)

    // Send email with signup link
    signupLink := fmt.Sprintf("http://localhost:3001/signup?inviteToken=%s", inviteToken)
    // sendInviteEmail(req.Email, signupLink)

    return c.JSON(invite)
}
```

```typescript
// Frontend: frontend/src/app/(auth)/signup/_components/signup-form.tsx
export function SignupForm() {
  const searchParams = useSearchParams();
  const inviteToken = searchParams.get("inviteToken");

  useEffect(() => {
    if (inviteToken) {
      // Decode token to get org + role info
      const inviteData = decodeInviteToken(inviteToken);
      // Pre-fill form with org name, role
      setOrgName(inviteData.orgName);
      setRole(inviteData.role);
    }
  }, [inviteToken]);

  const handleSignup = async (formData) => {
    const result = await registerAction({
      email: formData.email,
      password: formData.password,
      name: formData.name,
      role: inviteToken ? inviteData.role : "requester",
      inviteToken, // Pass token to backend
    });

    if (result.success) {
      // Invite link auto-adds user, redirect to dashboard
      router.push("/home");
    }
  };
}
```

```go
// Backend: handlers/auth.go - Register with Invite
func Register(c *fiber.Ctx) error {
    var req struct {
        Email       string `json:"email"`
        Password    string `json:"password"`
        Name        string `json:"name"`
        Role        string `json:"role"`
        InviteToken string `json:"invite_token"`
    }
    c.BodyParser(&req)

    // Create user (as before)
    newUser := models.User{
        ID:     utils.GenerateUserID(),
        Email:  req.Email,
        Name:   req.Name,
        Role:   req.Role,
        Active: true,
    }
    config.DB.Create(&newUser)

    // If invite token provided, auto-add to organization
    if req.InviteToken != "" {
        claims, _ := utils.ValidateInviteToken(req.InviteToken)
        orgID := claims.OrganizationID

        // Verify invite still valid
        var invite models.OrganizationInvite
        config.DB.Where("token = ? AND email = ? AND expires_at > ?",
            req.InviteToken, req.Email, time.Now()).
            First(&invite)

        if invite.ID != "" {
            // Add user to organization
            member := models.OrganizationMember{
                ID:             utils.GenerateID(),
                OrganizationID: orgID,
                UserID:         newUser.ID,
                Role:           invite.Role,
                Active:         true,
                JoinedAt:       time.Now(),
            }
            config.DB.Create(&member)

            // Set as current org
            config.DB.Model(&newUser).
                Update("current_organization_id", orgID)

            // Mark invite as used
            config.DB.Model(&invite).Update("used_at", time.Now())
        }
    }

    return c.JSON(types.SuccessResponse{...})
}
```

---

#### Scenario C: Default Organization
Create a default "Personal" organization for each user on signup.

**Flow:**
```
1. User signs up (email, password, name)
2. User created with role = "requester"
3. Automatic: Create "Personal" org for user
4. Automatic: Add user as admin of Personal org
5. Set as CurrentOrganizationID
6. User can immediately start using app
7. Later: Can be invited to other orgs
```

**Implementation:**

```go
// Backend: handlers/auth.go - Register with Auto-Org
func Register(c *fiber.Ctx) error {
    var req struct {
        Email    string `json:"email"`
        Password string `json:"password"`
        Name     string `json:"name"`
    }
    c.BodyParser(&req)

    // Create user
    newUser := models.User{
        ID:     utils.GenerateUserID(),
        Email:  req.Email,
        Name:   req.Name,
        Role:   "requester",
        Active: true,
    }
    config.DB.Create(&newUser)

    // AUTO-CREATE personal organization
    personalOrg := models.Organization{
        ID:    utils.GenerateOrgID(),
        Name:  fmt.Sprintf("%s's Organization", req.Name),
        Slug:  strings.ToLower(strings.ReplaceAll(req.Email, "@", "-")),
        Type:  "personal", // Mark as personal org
    }
    config.DB.Create(&personalOrg)

    // Add user as admin of personal org
    member := models.OrganizationMember{
        ID:             utils.GenerateID(),
        OrganizationID: personalOrg.ID,
        UserID:         newUser.ID,
        Role:           "admin", // User is admin of their own org
        Active:         true,
        JoinedAt:       time.Now(),
    }
    config.DB.Create(&member)

    // Set as current organization
    newUser.CurrentOrganizationID = &personalOrg.ID
    config.DB.Save(&newUser)

    // Generate token WITH org context
    token, _ := utils.GenerateToken(
        newUser.ID,
        newUser.Email,
        newUser.Name,
        newUser.Role,
        &personalOrg.ID, // Include org in token
    )

    return c.JSON(types.SuccessResponse{
        Success: true,
        Message: "User registered successfully",
        Token:   token,
        User:    newUser,
    })
}
```

**Recommended: Scenario C (Default Organization)**

**Reasoning:**
- ✅ Simplest for new users - immediate access
- ✅ No intermediate screens needed
- ✅ Doesn't require email verification
- ✅ Users can immediately create content
- ✅ Can invite others later, or join other orgs
- ✅ Reduces friction in onboarding

---

## 4. Organization Member Lifecycle

### Adding a Member to Organization

**Admin-Only Action** (org.go: line 146)

```go
func AddOrganizationMember(c *fiber.Ctx) error {
    tenant := middleware.GetTenantContext(c)

    // Check if user is org admin
    if tenant.UserRole != "admin" {
        return c.Status(403).JSON(fiber.Map{
            "error": "Only admins can add members",
        })
    }

    var req struct {
        UserID string `json:"user_id"` // Or email to invite
        Role   string `json:"role"`    // approver, requester, finance, viewer
    }
    c.BodyParser(&req)

    // Validate role
    validRoles := []string{"admin", "approver", "requester", "finance", "viewer"}
    if !contains(validRoles, req.Role) {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid role"})
    }

    // Add member
    member := models.OrganizationMember{
        ID:             utils.GenerateID(),
        OrganizationID: tenant.OrganizationID,
        UserID:         req.UserID,
        Role:           req.Role,
        Active:         true,
        JoinedAt:       time.Now(),
    }
    config.DB.Create(&member)

    return c.JSON(member)
}
```

### Removing a Member

**Admin-Only, Cannot Remove Last Admin** (org.go: lines 191-195)

```go
func RemoveOrganizationMember(c *fiber.Ctx) error {
    tenant := middleware.GetTenantContext(c)
    userIDToRemove := c.Params("userId")

    // Only admins can remove
    if tenant.UserRole != "admin" {
        return c.Status(403).JSON(fiber.Map{"error": "Only admins can remove members"})
    }

    // Cannot remove last admin
    var adminCount int64
    config.DB.Model(&models.OrganizationMember{}).
        Where("organization_id = ? AND role = ? AND active = true",
            tenant.OrganizationID, "admin").
        Count(&adminCount)

    if adminCount <= 1 {
        return c.Status(400).JSON(fiber.Map{
            "error": "Cannot remove the last admin",
        })
    }

    // Soft delete (set active = false)
    config.DB.Model(&models.OrganizationMember{}).
        Where("organization_id = ? AND user_id = ?",
            tenant.OrganizationID, userIDToRemove).
        Update("active", false)

    return c.JSON(fiber.Map{"success": true})
}
```

### Member Status Transitions

```
User Registration
    ↓
Create/Invite to Organization
    ↓
Add as OrganizationMember (active=true)
    ↓
Can access org-scoped resources
    ↓
[Member can be removed (active=false)]
    ↓
Can be reactivated
```

---

## 5. Permission Checking in Handlers

### Pattern 1: Role-Based Check (Current)

```go
func CreateRequisition(c *fiber.Ctx) error {
    tenant := middleware.GetTenantContext(c)

    // Only requesters and admins can create requisitions
    if tenant.UserRole != "requester" && tenant.UserRole != "admin" {
        return c.Status(403).JSON(fiber.Map{
            "error": "Only requesters can create requisitions",
        })
    }

    // ... create requisition ...
    return c.JSON(requisition)
}
```

### Pattern 2: Permission-Based Check (Future)

```go
func CreateRequisition(c *fiber.Ctx) error {
    tenant := middleware.GetTenantContext(c)
    userRole := tenant.UserRole

    // Check if user has permission
    if !permissions.HasPermission(userRole, "create_requisition") {
        return c.Status(403).JSON(fiber.Map{
            "error": "You don't have permission to create requisitions",
        })
    }

    // ... create requisition ...
    return c.JSON(requisition)
}
```

### Pattern 3: Middleware for Permission Checks (Best)

```go
// Middleware: RequirePermission
func RequirePermission(permission string) fiber.Handler {
    return func(c *fiber.Ctx) error {
        tenant := middleware.GetTenantContext(c)

        if !permissions.HasPermission(tenant.UserRole, permission) {
            return c.Status(403).JSON(fiber.Map{
                "error": fmt.Sprintf("Permission denied: %s", permission),
            })
        }

        return c.Next()
    }
}

// Usage in routes:
tenant.Post("/requisitions",
    middleware.RequirePermission("create_requisition"),
    handlers.CreateRequisition)
```

---

## 6. Frontend Authorization Integration

### Session Structure

**Updated AuthSession** (`frontend/src/types/auth.ts`):

```typescript
export interface AuthSession {
  access_token: string;           // JWT token
  role: UserType;                 // Global role (requester, approver, etc.)
  user_id?: string;               // User ID
  organization_id?: string;       // Current organization context
  organization?: {
    id: string;
    name: string;
    slug: string;
  };
  user?: User;                    // Cached user object
  change_password?: boolean;
  expiresAt?: Date | string;
}
```

### Permission Checking in Components

```typescript
// frontend/src/lib/permissions.ts
export const ROLE_PERMISSIONS = {
  requester: [
    "create_requisition",
    "view_requisition",
    "create_draft",
  ],
  approver: [
    "view_requisition",
    "approve_requisition",
    "reject_requisition",
  ],
  finance: [
    "create_budget",
    "view_budget",
    "manage_vendors",
    "view_payment_voucher",
    "approve_payment_voucher",
  ],
  viewer: [
    "view_requisition",
    "view_budget",
    "view_reports",
  ],
  admin: ["*"], // All permissions
};

export async function hasPermission(
  requiredPermission: string
): Promise<boolean> {
  const { session } = await verifySession();
  if (!session?.role) return false;

  const role = session.role as keyof typeof ROLE_PERMISSIONS;
  const permissions = ROLE_PERMISSIONS[role] || [];

  // Admin has all permissions
  if (permissions.includes("*")) return true;

  return permissions.includes(requiredPermission);
}

export async function canCreateRequisition(): Promise<boolean> {
  return hasPermission("create_requisition");
}

export async function canApproveRequisition(): Promise<boolean> {
  return hasPermission("approve_requisition");
}
```

### Component Usage

```typescript
// frontend/src/components/requisition-button.tsx
"use client";

import { useQuery } from "@tanstack/react-query";
import { canCreateRequisition } from "@/lib/permissions";

export function CreateRequisitionButton() {
  const { data: canCreate } = useQuery({
    queryKey: ["permissions", "create_requisition"],
    queryFn: canCreateRequisition,
  });

  if (!canCreate) {
    return null; // Don't show button if no permission
  }

  return (
    <Button onClick={() => router.push("/requisitions/new")}>
      Create Requisition
    </Button>
  );
}
```

---

## 7. Organization Context Flow (Complete)

### Request Flow with Organization Context

```
User Makes Request (with X-Organization-ID header or cookie)
    ↓
AuthMiddleware: Validates JWT Token
    ├─ Extracts userID, email, role from token
    ├─ Stores in context.locals["userID"]
    └─ Stores in context.locals["userRole"] (global role)
    ↓
TenantMiddleware: Resolves Organization Context
    ├─ Gets X-Organization-ID from header (or User.CurrentOrganizationID)
    ├─ Looks up OrganizationMember (userID, orgID)
    ├─ Extracts role from membership
    ├─ Verifies membership is active
    └─ Stores in context.locals["tenant"] (TenantContext)
    ↓
Handler: Uses TenantContext for Authorization & Filtering
    ├─ tenant.UserRole = role in THIS org (may differ from global role)
    ├─ tenant.OrganizationID = current org
    ├─ Checks role/permissions for org-specific resources
    └─ All queries filtered by organization_id
    ↓
Response: Org-scoped data returned to user
```

### Example Request

```
Request:
  GET /api/v1/requisitions
  Header: X-Organization-ID: org-123
  Auth: Bearer <JWT token with user-456>

AuthMiddleware:
  ✓ Token valid
  ✓ userID = user-456
  ✓ global role = "requester"

TenantMiddleware:
  ✓ Look up OrganizationMember where userID=user-456, orgID=org-123
  ✓ Found: role="requester" in org-123
  ✓ Member is active=true
  ✓ Store in context

Handler (GetRequisitions):
  ✓ Check permission: "view_requisition" with role="requester" ✓
  ✓ Query: SELECT * FROM requisitions
           WHERE organization_id = 'org-123'
           AND (created_by = 'user-456' OR assigned_to = 'user-456')
  ✓ Return results

Response:
  ✓ Only requisitions from org-123 that user can see
```

---

## 8. Implementation Roadmap

### Phase 1: User Registration & Default Organization (MVP)
- [x] Backend: Register endpoint creates default personal org (Scenario C)
- [ ] Frontend: Update signup form and login flow
- [ ] Frontend: Organization selector when user has multiple orgs
- [ ] Backend: Test user registration with org creation

### Phase 2: Permission Model (Quick)
- [ ] Backend: Add `permissions.go` with role → permission mapping
- [ ] Backend: Add permission check middleware
- [ ] Backend: Update handlers to use permission checks
- [ ] Frontend: Add permission checking functions

### Phase 3: Organization Invitations (Enhancement)
- [ ] Backend: Create invitation endpoints (invite token generation)
- [ ] Backend: Update register to handle invite tokens
- [ ] Frontend: Invitation acceptance flow
- [ ] Email: Send invitation emails (if applicable)

### Phase 4: Advanced Member Management (Future)
- [ ] Backend: Custom permissions per user (JSON override)
- [ ] Backend: Department-based access control
- [ ] Frontend: Admin panel for member management
- [ ] Backend: Audit logging for org changes

---

## 9. Frontend Architecture Changes

### Updated File Structure

```
frontend/src/
├── lib/
│   ├── auth.ts                    [Updated]
│   ├── permissions.ts             [NEW]
│   └── organization.ts            [NEW]
├── app/
│   ├── (auth)/
│   │   ├── signup/
│   │   │   └── page.tsx           [Updated]
│   │   └── login/
│   │       └── page.tsx           [Updated]
│   ├── onboarding/
│   │   └── create-organization/   [NEW]
│   └── (dashboard)/
│       ├── settings/
│       │   └── members/           [NEW]
│       └── requisitions/
│           └── new/
│               └── page.tsx       [Updated - permission check]
└── contexts/
    ├── auth-context.tsx           [Update with org]
    └── permissions-context.tsx    [NEW]
```

### Updated Components

**Login/Signup Flow:**
- Signup → Auto-create personal org → Login → Dashboard

**Organization Switcher:**
- Show all orgs user belongs to
- Quick switch between organizations
- Update `CurrentOrganizationID` and `X-Organization-ID` header

**Member Management:**
- Admin-only section to add/remove members
- Invite users via email
- Manage member roles

**Permission Checks:**
- Show/hide buttons based on permissions
- Disable actions user can't perform
- Display permission-denied messages

---

## 10. Key Decisions Summary

| Decision | Rationale |
|----------|-----------|
| **Default Organization (Scenario C)** | Reduces friction, immediate access, simplest to implement |
| **Role in User vs OrganizationMember** | Allows different roles per org, more flexible |
| **Permission-Based Access** | More maintainable than hard-coded role checks |
| **X-Organization-ID Header** | Consistent with standard multi-tenant patterns |
| **Soft Delete for Members** | Allows deactivation without data loss |
| **Admin-only Member Management** | Prevents unauthorized access changes |

---

## 11. Security Considerations

### Data Isolation
- ✅ All queries filtered by organization_id
- ✅ Cannot access data from other organizations
- ✅ Organization membership verified before access
- ✅ Unique constraint: (organization_id, user_id)

### Permission Validation
- ✅ Check at middleware level for critical operations
- ✅ Check in handler for business logic
- ✅ Never trust client-side permissions alone
- ✅ Always verify user is active member

### Token Security
- ✅ JWT includes user role and org context
- ✅ 24-hour expiration at backend
- ✅ 30-minute frontend session expiration
- ✅ Token refresh before expiration

### Audit Trail
- ✅ Log all member additions/removals
- ✅ Log all permission-denied attempts
- ✅ Log all org switches
- ✅ Track created_by for all resources

---

## 12. Testing Checklist

- [ ] User can sign up and auto-create personal org
- [ ] User can create requisition in their org
- [ ] User cannot access other org's requisitions
- [ ] Admin can add member to org
- [ ] New member can access org-scoped resources with correct role
- [ ] User cannot remove last admin from org
- [ ] Permissions enforced: requester cannot approve, etc.
- [ ] Organization switch updates CurrentOrganizationID
- [ ] X-Organization-ID header respected in queries
- [ ] Invitation tokens expire after 7 days
- [ ] Invited user auto-added to organization

---

## Conclusion

This architecture provides:
1. **Flexibility**: Different roles per organization
2. **Security**: Multi-level authorization checks
3. **Simplicity**: Clear permission model
4. **Scalability**: Database-driven permissions (future)
5. **User Experience**: Auto-org creation on signup, no intermediate screens

The default organization approach (Scenario C) is recommended for MVP and can be enhanced with invitation workflows later.
