# Authentication, Authorization, Permissions & User Creation - Deep Dive

**Last Updated**: 2025-12-25
**Status**: Phases 2-4 Implementation Complete
**Scope**: Complete end-to-end authentication, authorization, and permission system

---

## Table of Contents

1. [System Overview](#system-overview)
2. [User Creation & Organization Flow](#user-creation--organization-flow)
3. [Authentication System](#authentication-system)
4. [Authorization & Permissions](#authorization--permissions)
5. [Security Implementation](#security-implementation)
6. [Data Flows](#data-flows)
7. [File References](#file-references)

---

## System Overview

Liyali Gateway implements a **complete authentication and authorization system** with three key layers:

```
┌─────────────────────────────────────────────────────────────┐
│                    USER AUTHENTICATION                      │
│            (Who are you? Login/Password Verification)       │
│                                                             │
│  Frontend: Login Form → Backend: bcrypt + JWT Token        │
│  Result: Access token stored in HTTP-only cookie           │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│              ORGANIZATION CONTEXT (Multi-Tenancy)           │
│         (Which organization are you working in?)            │
│                                                             │
│  JWT includes org ID → TenantMiddleware validates member   │
│  Result: All queries scoped to organization                │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│           AUTHORIZATION & PERMISSIONS (RBAC)                │
│              (What are you allowed to do?)                  │
│                                                             │
│  RequirePermission Middleware → Check role permissions     │
│  Result: Access granted/denied based on role or custom     │
└─────────────────────────────────────────────────────────────┘
```

---

## User Creation & Organization Flow

### Complete Registration Process

When a user signs up, the system performs these operations:

```
┌─────────────────────────────────────────────────────────────┐
│  USER SUBMITS REGISTRATION FORM (Frontend)                 │
│                                                             │
│  Input Fields:                                              │
│  - Email (required)                                         │
│  - Password (required, 8+ chars, mixed case + digit)       │
│  - Name (required)                                          │
│  - Role (optional, default: "requester")                   │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│  BACKEND VALIDATION (auth.go Register handler)              │
│                                                             │
│  Checks:                                                    │
│  1. All required fields present                             │
│  2. Email format valid                                      │
│  3. Email not already registered                            │
│  4. Password strength (8+ chars, UC, LC, digit)            │
│  5. Role is valid (admin|approver|requester|finance|viewer)│
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│  SECURITY: PASSWORD HASHING (utils/password.go)             │
│                                                             │
│  Algorithm: bcrypt                                          │
│  Cost Factor: 10 (DefaultCost)                              │
│  Process:                                                   │
│  - Generate random salt                                     │
│  - Hash password with salt                                  │
│  - Store only hash (never store plaintext)                  │
│  Result: $2b$10$... (bcrypt hash)                          │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│  USER RECORD CREATION (models/models.go)                    │
│                                                             │
│  Table: users                                               │
│  Columns:                                                   │
│  - id: UUID (unique identifier)                             │
│  - email: string (unique index)                             │
│  - name: string                                             │
│  - role: string (system role)                              │
│  - password_hash: bcrypt hash                              │
│  - active: boolean = true                                   │
│  - current_organization_id: NULL (set below)              │
│  - is_super_admin: boolean = false                         │
│  - created_at, updated_at: timestamps                      │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│  AUTO-CREATE PERSONAL ORGANIZATION (Phase 2 Feature)        │
│                                                             │
│  Automatic Process:                                         │
│  1. Create Organization record                             │
│     - id: UUID                                              │
│     - name: User's name (e.g., "John Smith")              │
│     - slug: URL-safe slug (e.g., "john-smith")            │
│     - description: "Personal Organization"                 │
│     - tier: "free"                                          │
│     - created_by: User ID                                   │
│  2. Create OrganizationSettings with defaults              │
│     - Currency: USD                                         │
│     - Fiscal year start: January 1                         │
│     - Budget validation: enabled                           │
│  3. Create OrganizationMember relationship                 │
│     - organization_id: New organization                    │
│     - user_id: New user                                     │
│     - role: "admin" (org-level admin)                      │
│     - active: true                                          │
│     - joined_at: NOW                                        │
│  4. Update User.CurrentOrganizationID                      │
│     - Set to newly created organization ID                 │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│  GENERATE AUTHENTICATION TOKEN (utils/jwt.go)               │
│                                                             │
│  JWT Structure (HS256 with HMAC-SHA256):                    │
│  {                                                          │
│    "header": {                                              │
│      "alg": "HS256",                                        │
│      "typ": "JWT"                                           │
│    },                                                       │
│    "payload": {                                             │
│      "sub": "user-uuid",           // Subject (user ID)    │
│      "email": "user@example.com",                          │
│      "name": "User Name",                                   │
│      "role": "requester",           // System role          │
│      "currentOrgId": "org-uuid",   // Multi-tenancy scope  │
│      "jti": "token-uuid",           // JWT ID (for revoke)  │
│      "iss": "liyali-gateway",      // Issuer               │
│      "exp": 1735046400,            // Expires 24 hours     │
│      "iat": 1734960000,            // Issued at            │
│      "nbf": 1734960000             // Not before           │
│    },                                                       │
│    "signature": "base64url(HMAC-SHA256(...))"              │
│  }                                                          │
│                                                             │
│  Secret: Read from JWT_SECRET env var (32+ chars)         │
│  Expiration: 24 hours from creation                        │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│  RETURN RESPONSE TO FRONTEND                                │
│                                                             │
│  Response Body:                                             │
│  {                                                          │
│    "success": true,                                         │
│    "message": "User created successfully",                 │
│    "data": {                                                │
│      "token": "eyJhbGc...",     // JWT token               │
│      "user": {                                              │
│        "id": "user-uuid",                                   │
│        "email": "user@example.com",                        │
│        "name": "User Name",                                 │
│        "role": "requester"                                  │
│      },                                                     │
│      "organization": {           // Auto-created org        │
│        "id": "org-uuid",                                    │
│        "name": "User Name",                                │
│        "slug": "user-name",                                │
│        "tier": "free",                                      │
│        "active": true                                       │
│      }                                                      │
│    }                                                        │
│  }                                                          │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│  FRONTEND: CREATE ENCRYPTED SESSION (lib/auth.ts)           │
│                                                             │
│  Session Creation:                                          │
│  1. Extract token from response                             │
│  2. Encrypt token using jose library                        │
│  3. Store in HTTP-only cookie named 'session'              │
│     - Not accessible via JavaScript (XSS protection)        │
│     - Only sent over HTTPS in production                    │
│     - 24-hour expiration (synced with JWT)                 │
│  4. Store metadata in localStorage:                         │
│     - user_id                                               │
│     - organization_id                                       │
│     - role                                                  │
│  5. Context update (OrganizationContext)                    │
│     - Fetch user's organizations                            │
│     - Set current organization                              │
│     - Initialize UI with org switcher                       │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│  REDIRECT TO DASHBOARD                                      │
│                                                             │
│  Browser:                                                   │
│  - Redirect to /home                                        │
│  - Session cookie included with every request              │
│  - User sees dashboard with:                                │
│    * Organization switcher showing "User Name"             │
│    * Full access to organization                           │
│    * Workflow modules (requisitions, budgets, etc.)        │
│    * Role-based menu (what they can access)                │
└─────────────────────────────────────────────────────────────┘
```

### Key Outcome: What Was Created

After registration completes:

| Entity | Details |
|--------|---------|
| **User Record** | Email, hashed password, system role, org reference |
| **Organization** | Named after user, with settings, members, and roles |
| **Membership** | User added as admin to their personal organization |
| **JWT Token** | 24-hour token with all user/org context |
| **Session Cookie** | Encrypted, HTTP-only, synced with JWT |
| **Audit Trail** | CreatedAt timestamp for both user and organization |

---

## Authentication System

### Login Process (Detailed)

```
STEP 1: USER SUBMITS LOGIN
┌─────────────────────────────────────────────────────────────┐
│  Frontend: Login Form (login-form.tsx)                      │
│  Input: { email, password }                                │
│  Action: Calls loginAction(email, password)                │
└─────────────────────────────────────────────────────────────┘
                            ↓
STEP 2: BACKEND RECEIVES LOGIN REQUEST
┌─────────────────────────────────────────────────────────────┐
│  Backend: handlers/auth.go Login() [Line 18-114]            │
│                                                             │
│  1. Parse LoginRequest                                      │
│     - Validate email and password provided                  │
│  2. Query database for user by email                        │
│     - If not found: return 400 "Invalid credentials"       │
│  3. Get user's role (from system role in User.Role)        │
│  4. Get user's current organization                         │
│     - From User.CurrentOrganizationID                       │
└─────────────────────────────────────────────────────────────┘
                            ↓
STEP 3: PASSWORD VERIFICATION
┌─────────────────────────────────────────────────────────────┐
│  Security: utils/password.go VerifyPassword()               │
│                                                             │
│  Process:                                                   │
│  1. Retrieve hashed password from User.PasswordHash         │
│  2. Call bcrypt.CompareHashAndPassword()                   │
│  3. Bcrypt performs:                                        │
│     - Extracts salt from stored hash                        │
│     - Hash provided password with that salt                 │
│     - Compare hashes in constant time                       │
│  4. Result: true or false                                   │
│                                                             │
│  If failed:                                                 │
│  - Record LoginAttempt (failed)                             │
│  - Return 401 "Invalid credentials"                         │
│  - Increment failure counter                                │
│  - Check if account should be locked (Phase 4A.2)          │
└─────────────────────────────────────────────────────────────┘
                            ↓
STEP 4: ACCOUNT STATUS CHECK
┌─────────────────────────────────────────────────────────────┐
│  Verify user can login:                                     │
│                                                             │
│  - Check User.Active == true                                │
│  - Check not in AccountLockout table (Phase 4A.2)          │
│  - Check email verified (if required, Phase 4B.1)          │
│                                                             │
│  If blocked: return 401 with reason                        │
└─────────────────────────────────────────────────────────────┘
                            ↓
STEP 5: RECORD SUCCESSFUL LOGIN
┌─────────────────────────────────────────────────────────────┐
│  Audit Trail: LoginAttempt record (models/auth.go)          │
│                                                             │
│  Create LoginAttempt:                                       │
│  {                                                          │
│    id: UUID,                                                │
│    user_id: user.ID,                                        │
│    email: user.Email,                                       │
│    ip_address: request.IP(),                                │
│    user_agent: request.Header("User-Agent"),               │
│    success: true,                                           │
│    attempt_at: NOW,                                         │
│    reason: ""                                               │
│  }                                                          │
│                                                             │
│  Update User.LastLogin = NOW                                │
└─────────────────────────────────────────────────────────────┘
                            ↓
STEP 6: GENERATE JWT TOKEN
┌─────────────────────────────────────────────────────────────┐
│  File: utils/jwt.go GenerateToken()                         │
│                                                             │
│  Token Creation:                                            │
│  - Generate unique JTI (jwt.New().String())                 │
│  - Create CustomClaims with:                                │
│    * Subject: user.ID                                       │
│    * Email: user.Email                                      │
│    * Name: user.Name                                        │
│    * Role: user.Role                                        │
│    * CurrentOrgID: user.CurrentOrganizationID              │
│    * JTI: unique ID for revocation                         │
│    * ExpiresAt: now + 24 hours                              │
│    * IssuedAt: now                                          │
│    * Issuer: "liyali-gateway"                              │
│  - Sign with HS256 using JWT_SECRET                         │
│  - Return: token string + TokenInfo metadata                │
└─────────────────────────────────────────────────────────────┘
                            ↓
STEP 7: RETURN SUCCESS RESPONSE
┌─────────────────────────────────────────────────────────────┐
│  Response to Frontend:                                      │
│  {                                                          │
│    "success": true,                                         │
│    "message": "Login successful",                           │
│    "data": {                                                │
│      "token": "eyJhbGc...",   // JWT token                  │
│      "user": {                                              │
│        "id": "user-uuid",                                   │
│        "email": "user@example.com",                        │
│        "name": "User Name",                                 │
│        "role": "requester"    // System role               │
│      },                                                     │
│      "organization": {                                      │
│        "id": "org-uuid",                                    │
│        "name": "Organization Name",                        │
│        "slug": "org-slug"                                   │
│      }                                                      │
│    }                                                        │
│  }                                                          │
└─────────────────────────────────────────────────────────────┘
                            ↓
STEP 8: FRONTEND SESSION CREATION
┌─────────────────────────────────────────────────────────────┐
│  File: lib/auth.ts createAuthSession()                      │
│                                                             │
│  Process:                                                   │
│  1. Extract token from response                             │
│  2. Encrypt token using jose.jwtEncrypt()                   │
│  3. Create AuthSession object:                              │
│     {                                                       │
│       access_token: encrypted_jwt,                          │
│       role: response.user.role,                             │
│       user_id: response.user.id,                            │
│       organization_id: response.organization.id             │
│     }                                                       │
│  4. Store in HTTP-only cookie                               │
│     - Cookie name: 'session'                                │
│     - Max age: 86400 (24 hours)                             │
│     - HTTP only: true (JavaScript can't access)             │
│     - Secure: true (HTTPS only in production)               │
│     - SameSite: 'lax' (CSRF protection)                     │
│  5. Update localStorage with metadata:                      │
│     - localStorage.set('user_id', user_id)                  │
│     - localStorage.set('organization_id', org_id)           │
│     - localStorage.set('role', role)                        │
│  6. Invalidate all queries (React Query)                    │
│     - Clear cache to fetch fresh data                       │
└─────────────────────────────────────────────────────────────┘
                            ↓
STEP 9: ORGANIZATION CONTEXT INITIALIZATION
┌─────────────────────────────────────────────────────────────┐
│  File: contexts/organization-context.tsx                    │
│                                                             │
│  On Mount:                                                  │
│  1. Fetch user's organizations                              │
│     - Call /api/v1/organizations (auth required)            │
│     - Returns all orgs user is member of                    │
│  2. Load current organization:                              │
│     - Check localStorage for 'current_org_id'               │
│     - If not found, use first organization                  │
│     - Store in localStorage                                 │
│  3. Set OrganizationContext:                                │
│     - currentOrganization: selected org                     │
│     - userOrganizations: all user's orgs                    │
│     - switchWorkspace(orgId): switch function               │
│  4. Provide to all child components                         │
└─────────────────────────────────────────────────────────────┘
                            ↓
STEP 10: REDIRECT TO DASHBOARD
┌─────────────────────────────────────────────────────────────┐
│  Frontend: Login flow complete                              │
│                                                             │
│  - Redirect to /home                                        │
│  - Session cookie automatically included on all requests    │
│  - Organization context available                           │
│  - User can now make authenticated API calls                │
│  - Dashboard shows user's organization and role             │
└─────────────────────────────────────────────────────────────┘
```

### Token Management

#### Token Generation
```go
// File: utils/jwt.go GenerateToken()
// Creates new 24-hour token with:
// - Unique JTI per token
// - User context (id, email, name, role)
// - Organization context (currentOrgId)
// - HS256 signature with JWT_SECRET
```

#### Token Validation
```go
// File: utils/jwt.go ValidateToken()
// Checks:
// - Signature is valid (prevents tampering)
// - Token not expired
// - Algorithm is HS256 (prevents algorithm switching)
// - Returns CustomClaims if valid
```

#### Token Refresh
```go
// File: utils/jwt.go RefreshToken()
// Extends expiration:
// - Takes existing claims
// - Issues new token with same claims
// - New expiration: 24 hours from now
// - Same JTI (preserves token identity)
```

#### Token Revocation (Phase 4A.1)
```go
// File: services/auth_service.go
// BlacklistToken(): Add token JTI to blacklist
// IsTokenBlacklisted(): Check before accepting token
// RevokeUserTokens(): Logout all sessions for user
```

---

## Authorization & Permissions

### Role Hierarchy

Liyali Gateway uses a **two-level role system**:

```
LEVEL 1: SYSTEM ROLE (in User record)
┌─────────────────────────────────────────────────────────────┐
│  Assigned to user at registration                           │
│  Used as default for all organizations                      │
│  Can be overridden per organization (Phase 3.5)             │
│                                                             │
│  Default System Roles:                                      │
│  ├─ admin       (39 permissions - full access)              │
│  ├─ approver    (18 permissions - workflow management)      │
│  ├─ requester   (8 permissions - creation only)             │
│  ├─ finance     (16 permissions - financial focus)          │
│  └─ viewer      (8 permissions - read-only)                 │
└─────────────────────────────────────────────────────────────┘

LEVEL 2: ORGANIZATION ROLE (in OrganizationMember record)
┌─────────────────────────────────────────────────────────────┐
│  Can be different per organization                          │
│  Overrides system role within that organization             │
│  Examples: "Budget Reviewer", "Senior Manager", "Intern"    │
│  Can have custom permissions (Phase 3.5)                    │
│                                                             │
│  Models:                                                    │
│  - OrganizationRole (role definition)                       │
│  - PermissionAssignment (role → permission mapping)         │
│  - OrganizationPermission (org's permission set)            │
└─────────────────────────────────────────────────────────────┘
```

### Permission Structure

```
Permission = Resource + Action

Example Permissions:
┌──────────────────────────────────────────────────┐
│  Resource: requisition     Action: view          │
│  Resource: requisition     Action: create        │
│  Resource: requisition     Action: edit          │
│  Resource: requisition     Action: approve       │
│  Resource: requisition     Action: reject        │
│  Resource: budget          Action: manage        │
│  Resource: organization    Action: manage_users  │
└──────────────────────────────────────────────────┘

Resources:
- requisition, budget, purchase_order, payment_voucher
- grn (goods received notes)
- vendor, category, organization
- analytics, audit_log

Actions:
- view, create, edit, delete
- approve, reject, reassign
- manage_users, manage_workflows
```

### Permission Checking Flow

```
API REQUEST ARRIVES AT PROTECTED ENDPOINT
┌─────────────────────────────────────────────────────────────┐
│  Example: POST /api/v1/requisitions/:id/approve             │
│           [RequirePermission(db, "requisition", "approve")]  │
└─────────────────────────────────────────────────────────────┘
                            ↓
MIDDLEWARE LAYER 1: AUTHENTICATION
┌─────────────────────────────────────────────────────────────┐
│  File: middleware/middleware.go AuthMiddleware()            │
│                                                             │
│  Step 1: Extract JWT from Authorization header              │
│          "Authorization: Bearer eyJhbGc..."                 │
│  Step 2: Call utils.ValidateToken(token)                    │
│  Step 3: Check signature and expiration                     │
│  Step 4: If valid:                                          │
│          c.Locals("userID", claims.Subject)                 │
│          c.Locals("userRole", claims.Role)                  │
│  Step 5: If invalid, return 401 Unauthorized                │
└─────────────────────────────────────────────────────────────┘
                            ↓
MIDDLEWARE LAYER 2: TENANT CONTEXT
┌─────────────────────────────────────────────────────────────┐
│  File: middleware/tenant.go TenantMiddleware()              │
│                                                             │
│  Step 1: Get organization ID from:                          │
│          a) X-Organization-ID header (explicit switch)       │
│          b) User.CurrentOrganizationID (from JWT)            │
│  Step 2: Query OrganizationMember table                      │
│          WHERE user_id = ? AND organization_id = ?           │
│  Step 3: Verify membership:                                  │
│          - User is member of organization                    │
│          - Membership is active (Active = true)             │
│          - User hasn't been removed                          │
│  Step 4: If valid:                                          │
│          c.Locals("organizationID", org.ID)                 │
│  Step 5: If invalid, return 403 Forbidden                   │
└─────────────────────────────────────────────────────────────┘
                            ↓
MIDDLEWARE LAYER 3: PERMISSION CHECK
┌─────────────────────────────────────────────────────────────┐
│  File: middleware/middleware.go RequirePermission()         │
│  Parameters: resource="requisition", action="approve"       │
│                                                             │
│  Step 1: Get userID, userRole, organizationID from context  │
│  Step 2: Call PermissionService.HasPermission()             │
│          (userID, organizationID, userRole,                 │
│           resource, action)                                 │
│  Step 3: Two-Phase Permission Check (see below)             │
│  Step 4: If allowed, call c.Next()                          │
│  Step 5: If denied, return 403 Forbidden with message       │
└─────────────────────────────────────────────────────────────┘
                            ↓
PERMISSION SERVICE: TWO-PHASE CHECK
┌─────────────────────────────────────────────────────────────┐
│  File: services/permission_service.go HasPermission()       │
│                                                             │
│  PHASE 1: CHECK CUSTOM PERMISSIONS (Phase 3.5+)             │
│  ────────────────────────────────────────────────────────   │
│  Query database:                                            │
│  1. Find OrganizationRole by name and org_id                │
│  2. Join to PermissionAssignment table                       │
│  3. Join to OrganizationPermission table                     │
│  4. Check if (resource, action) pair exists                  │
│                                                             │
│  SQL Query Pattern:                                         │
│  SELECT op.* FROM organization_permissions op              │
│  INNER JOIN permission_assignments pa                       │
│    ON op.id = pa.permission_id                              │
│  INNER JOIN organization_roles r                            │
│    ON r.id = pa.organization_role_id                        │
│  WHERE r.name = ?                                           │
│    AND r.organization_id = ?                                │
│    AND op.resource = ?                                      │
│    AND op.action = ?                                        │
│                                                             │
│  If Found:                                                  │
│    └─ Return TRUE (permission granted)                      │
│                                                             │
│  PHASE 2: FALL BACK TO HARDCODED PERMISSIONS (Phase 3)      │
│  ────────────────────────────────────────────────────────   │
│  If custom permission not found:                            │
│  1. Get hardcoded RolePermissions map                        │
│  2. Look up role in map                                      │
│  3. Check if permission exists in role's permission list     │
│                                                             │
│  Example:                                                   │
│  RolePermissions["approver"] = [                            │
│    Permission{Resource: "requisition", Action: "approve"},   │
│    Permission{Resource: "requisition", Action: "reject"},    │
│    Permission{Resource: "budget", Action: "approve"},        │
│    ...                                                       │
│  ]                                                          │
│                                                             │
│  If Found:                                                  │
│    └─ Return TRUE (permission granted)                      │
│                                                             │
│  If Not Found:                                              │
│    └─ Return FALSE (permission denied)                      │
└─────────────────────────────────────────────────────────────┘
                            ↓
DECISION: ALLOW OR DENY
┌─────────────────────────────────────────────────────────────┐
│  If HasPermission() returns TRUE:                            │
│  ├─ Call handler function                                   │
│  ├─ Process request                                         │
│  └─ Return response to client                               │
│                                                             │
│  If HasPermission() returns FALSE:                          │
│  ├─ Return 403 Forbidden                                    │
│  ├─ Log denial in AuditLog (Phase 4A.3)                     │
│  └─ Client shows error message                              │
└─────────────────────────────────────────────────────────────┘
```

### Permission Matrix (Phase 3 - Hardcoded)

#### Admin Role (39 permissions)
- **All resources**: requisition, budget, purchase_order, payment_voucher, grn, vendor, category
- **All actions**: view, create, edit, delete, approve, reject, reassign, manage_*
- **Scope**: Full system access
- **Cannot**: Delete system roles, perform super-admin functions

#### Approver Role (18 permissions)
- **Core**: Approval workflow management
- **Can Create**: Requisitions, Budgets, Purchase Orders, Payment Vouchers
- **Can Approve/Reject**: All document types
- **Can View**: All resources
- **Cannot**: Delete documents, manage organization

#### Requester Role (8 permissions - Most Restricted)
- **Can Create**: Requisitions only
- **Can View**: Requisitions, budgets, analytics
- **Can Edit**: Own requisitions
- **Cannot**: Delete, approve, or modify others' documents

#### Finance Role (16 permissions)
- **Focus**: Financial operations
- **Can Create/Manage**: Budgets, payment vouchers, GRNs
- **Can Approve/Reject**: Budget and payment items
- **Can View**: Requisitions, analytics
- **Cannot**: Approve requisitions directly

#### Viewer Role (8 permissions - Read Only)
- **Can**: View all resources and analytics
- **Cannot**: Create, edit, delete, or approve anything

### Custom Roles (Phase 3.5)

```
ORGANIZATION ADMIN CREATES CUSTOM ROLE
┌─────────────────────────────────────────────────────────────┐
│  Endpoint: POST /api/v1/organizations/:id/roles              │
│  Request Body:                                               │
│  {                                                           │
│    "name": "Budget Reviewer",                                │
│    "description": "Reviews budgets before final approval"    │
│  }                                                           │
└─────────────────────────────────────────────────────────────┘
                            ↓
ROLE CREATED IN DATABASE
┌─────────────────────────────────────────────────────────────┐
│  Table: organization_roles                                   │
│  {                                                           │
│    "id": "uuid",                                             │
│    "organization_id": "org-uuid",                            │
│    "name": "Budget Reviewer",                                │
│    "description": "...",                                     │
│    "is_default": false,     // Not a system role            │
│    "is_active": true,                                        │
│    "created_at": NOW                                         │
│  }                                                           │
└─────────────────────────────────────────────────────────────┘
                            ↓
ASSIGN PERMISSIONS TO ROLE
┌─────────────────────────────────────────────────────────────┐
│  Endpoint: POST /api/v1/organizations/:id/roles/:roleId/    │
│            permissions/:permissionId                         │
│                                                              │
│  Process:                                                    │
│  1. Verify permission belongs to organization               │
│  2. Create PermissionAssignment record:                      │
│     {                                                        │
│       "id": "uuid",                                          │
│       "organization_role_id": "role-uuid",                   │
│       "permission_id": "perm-uuid"                           │
│     }                                                        │
│  3. Link established: Role → Permission                      │
│  4. Now users with this role have this permission            │
└─────────────────────────────────────────────────────────────┘
                            ↓
ASSIGN USERS TO CUSTOM ROLE
┌─────────────────────────────────────────────────────────────┐
│  Update OrganizationMember record:                           │
│  {                                                           │
│    "id": "uuid",                                             │
│    "organization_id": "org-uuid",                            │
│    "user_id": "user-uuid",                                   │
│    "role": "Budget Reviewer",   // Custom role              │
│    "active": true                                            │
│  }                                                           │
│                                                              │
│  User now has:                                              │
│  - System role (from User.Role): approver                    │
│  - Organization role: Budget Reviewer (custom perms)         │
└─────────────────────────────────────────────────────────────┘
                            ↓
PERMISSION LOOKUP AT REQUEST TIME
┌─────────────────────────────────────────────────────────────┐
│  PermissionService.HasPermission()                           │
│                                                              │
│  Phase 1: Query custom permissions                           │
│  - Find OrganizationRole "Budget Reviewer"                   │
│  - Find attached permissions                                 │
│  - If found and matches (budget, manage), return TRUE        │
│                                                              │
│  Phase 2: Fall back to hardcoded                             │
│  - Find system role "approver"                               │
│  - Check hardcoded permissions                               │
│  - Use as backup                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## Security Implementation

### Password Security

```
PASSWORD HASHING PROCESS
┌─────────────────────────────────────────────────────────────┐
│  File: utils/password.go HashPassword()                      │
│                                                             │
│  Input: "MyPassword123!"                                    │
│                                                             │
│  Bcrypt Process:                                            │
│  1. Generate random salt (16 bytes)                         │
│  2. Cost factor: 10 (DefaultCost)                           │
│     - Means: 2^10 = 1024 iterations                         │
│     - Takes ~100ms to hash                                  │
│     - Makes brute force slow                                │
│  3. Hash password with salt                                 │
│  4. Combine cost + salt + hash                              │
│                                                             │
│  Output: $2b$10$SALT...HASH                                │
│          └─ Version 2b (current Bcrypt)                     │
│             └─ Cost factor 10                               │
│                └─ Salt (22 chars)                           │
│                   └─ Hash (31 chars)                        │
│                                                             │
│  Database Storage:                                          │
│  users.password_hash = "$2b$10$..."                        │
│  (Never store plaintext password)                           │
└─────────────────────────────────────────────────────────────┘

PASSWORD VERIFICATION PROCESS
┌─────────────────────────────────────────────────────────────┐
│  File: utils/password.go VerifyPassword()                    │
│                                                             │
│  Input:                                                     │
│  - Hash: "$2b$10$..." (from database)                       │
│  - Password: "MyPassword123!" (from login form)             │
│                                                             │
│  Bcrypt Process:                                            │
│  1. Extract salt from stored hash                           │
│  2. Hash provided password with extracted salt              │
│  3. Compare two hashes in CONSTANT TIME                     │
│     - Prevents timing attack                                │
│     - Attacker can't tell how close guess is                │
│  4. Return: true or false                                   │
│                                                             │
│  Security Properties:                                       │
│  - One-way: Can't reverse password from hash                │
│  - Unique: Same password gives different hash (due to salt) │
│  - Slow: Takes ~100ms (brute force protection)              │
│  - Timing-safe: Comparison resists timing attacks           │
└─────────────────────────────────────────────────────────────┘

PASSWORD STRENGTH VALIDATION
┌─────────────────────────────────────────────────────────────┐
│  File: utils/password.go ValidatePasswordStrength()         │
│                                                             │
│  Requirements:                                              │
│  ✓ Minimum 8 characters                                      │
│  ✓ At least one UPPERCASE letter (A-Z)                      │
│  ✓ At least one lowercase letter (a-z)                      │
│  ✓ At least one digit (0-9)                                 │
│                                                             │
│  Examples:                                                  │
│  ✗ "password"       (no uppercase, no digit)                │
│  ✗ "PASSWORD123"    (no lowercase)                          │
│  ✗ "Pass123"        (only 7 characters)                     │
│  ✓ "MyPassword123"  (meets all requirements)                │
│                                                             │
│  Enforced at:                                               │
│  - Registration (backend validation)                         │
│  - Password reset (backend validation)                       │
│  - Optional: Frontend preview (UX only)                      │
└─────────────────────────────────────────────────────────────┘
```

### JWT Token Security

```
JWT STRUCTURE
┌─────────────────────────────────────────────────────────────┐
│  Header.Payload.Signature                                   │
│                                                             │
│  HEADER: Base64URL(JSON)                                    │
│  {                                                          │
│    "alg": "HS256",    // Algorithm                          │
│    "typ": "JWT"       // Token type                         │
│  }                                                          │
│                                                             │
│  PAYLOAD: Base64URL(JSON)                                   │
│  {                                                          │
│    "sub": "user-uuid",                    // User ID         │
│    "email": "user@example.com",                             │
│    "name": "User Name",                                     │
│    "role": "requester",          // System role             │
│    "currentOrgId": "org-uuid",  // Tenant scope             │
│    "jti": "token-uuid",          // Token ID (for revoke)    │
│    "iss": "liyali-gateway",     // Issuer                   │
│    "exp": 1735046400,           // Expiration (24h)         │
│    "iat": 1734960000,           // Issued at                │
│    "nbf": 1734960000            // Not before               │
│  }                                                          │
│                                                             │
│  SIGNATURE: Base64URL(HMAC-SHA256)                          │
│  HMAC-SHA256(                                               │
│    base64url(header) + "." + base64url(payload),            │
│    JWT_SECRET_KEY                                           │
│  )                                                          │
└─────────────────────────────────────────────────────────────┘

JWT VALIDATION
┌─────────────────────────────────────────────────────────────┐
│  File: utils/jwt.go ValidateToken()                          │
│                                                             │
│  Security Checks:                                           │
│  1. Parse JWT string                                        │
│  2. Verify signature:                                       │
│     - Recalculate HMAC with JWT_SECRET                      │
│     - Compare with provided signature                       │
│     - If mismatch: Token tampered with → Reject             │
│  3. Verify algorithm:                                       │
│     - Check alg == "HS256"                                  │
│     - Prevent algorithm switching (e.g., HS256 → none)      │
│  4. Verify expiration:                                      │
│     - Check exp > now                                       │
│     - If expired: Reject                                    │
│  5. Verify NotBefore:                                       │
│     - Check nbf < now                                       │
│  6. Extract and return claims                               │
│                                                             │
│  Attack Prevention:                                         │
│  - Tampering: Signature validation detects changes          │
│  - Replay: Expiration prevents old tokens                   │
│  - Algorithm switching: Force HS256                         │
│  - Clock skew: Can add leeway if needed                      │
│  - Missing signature: Reject tokens with "none" algorithm   │
└─────────────────────────────────────────────────────────────┘

JWT SECRET MANAGEMENT
┌─────────────────────────────────────────────────────────────┐
│  File: utils/jwt.go (line 16)                                │
│                                                             │
│  Secret Source: Environment variable                        │
│  secret := os.Getenv("JWT_SECRET")                          │
│                                                             │
│  Requirements:                                              │
│  ✓ Must be 32+ characters (256 bits minimum)                 │
│  ✓ Random and unpredictable                                 │
│  ✓ Unique per environment (dev != staging != prod)          │
│  ✓ Never hardcoded                                          │
│  ✓ Never committed to version control                       │
│  ✓ Stored in secrets manager (e.g., AWS Secrets Manager)    │
│                                                             │
│  Generation Example:                                        │
│  openssl rand -base64 32                                    │
│  $ gzJK7ZK/g8d3X8K4jK+mL7Q=                                 │
│                                                             │
│  If compromised:                                            │
│  - Rotate secret immediately                                │
│  - All existing tokens become invalid                        │
│  - Force users to re-login                                  │
└─────────────────────────────────────────────────────────────┘
```

### Session Security

```
HTTP-ONLY COOKIE STORAGE
┌─────────────────────────────────────────────────────────────┐
│  File: lib/auth.ts createAuthSession()                       │
│                                                             │
│  Cookie Properties:                                         │
│  {                                                          │
│    name: "session",                                         │
│    value: encrypted_jwt,    // Encrypted JWT value          │
│    httpOnly: true,          // Can't be accessed by JS      │
│    secure: true,            // Only over HTTPS              │
│    sameSite: "lax",         // CSRF protection              │
│    maxAge: 86400,           // 24 hours                      │
│    path: "/"                // Available site-wide           │
│  }                                                          │
│                                                             │
│  Security Properties:                                       │
│  - JavaScript can't access cookie (XSS protection)          │
│  - Only sent to server (prevents exfiltration)              │
│  - Encrypted in transmission (HTTPS required)               │
│  - Sampled after site switch (CSRF prevention)              │
│  - Expires after 24 hours (limits exposure window)          │
│                                                             │
│  Attack Scenarios Prevented:                                │
│  X: XSS Attack                                              │
│    - Attacker injects script                                │
│    - Can't access httpOnly cookie                           │
│    - Can't steal token                                      │
│                                                             │
│  X: CSRF Attack                                             │
│    - Attacker forges request from another site              │
│    - sameSite='lax' prevents cookie from being sent         │
│    - Request rejected                                       │
│                                                             │
│  X: Token Theft                                             │
│    - Token only in cookie                                   │
│    - Only sent over HTTPS                                   │
│    - Attacker can't intercept unencrypted                   │
└─────────────────────────────────────────────────────────────┘

ENCRYPTION & DECRYPTION
┌─────────────────────────────────────────────────────────────┐
│  File: lib/auth.ts                                           │
│  Library: jose (JSON Object Signing and Encryption)         │
│                                                             │
│  Encryption (Store in cookie):                              │
│  1. Take raw JWT                                            │
│  2. Encrypt using jose.jwtEncrypt()                         │
│     - Algorithm: A128CBC-HS256 (symmetric)                  │
│     - Key: Derived from session secret                      │
│  3. Encrypted JWT stored in cookie                          │
│                                                             │
│  Decryption (Retrieve from cookie):                         │
│  1. Get encrypted value from cookie                         │
│  2. Decrypt using jose.jwtDecrypt()                         │
│     - Returns original JWT                                  │
│  3. Validate JWT signature                                  │
│  4. Extract claims                                          │
│                                                             │
│  Two Layers of Security:                                    │
│  - Layer 1: Encryption (confidentiality)                    │
│  - Layer 2: JWT signature (authenticity)                    │
└─────────────────────────────────────────────────────────────┘
```

### Organization Isolation

```
TENANT MIDDLEWARE: VERIFY MEMBERSHIP
┌─────────────────────────────────────────────────────────────┐
│  File: middleware/tenant.go TenantMiddleware()              │
│                                                             │
│  For EVERY organization-scoped request:                     │
│                                                             │
│  Step 1: Get organization ID from:                          │
│          - X-Organization-ID header (explicit switch)        │
│          - Or JWT currentOrgId (implicit, default)          │
│                                                             │
│  Step 2: Query OrganizationMember table:                     │
│          WHERE user_id = ? AND organization_id = ?          │
│          AND active = true                                  │
│                                                             │
│  Step 3: If no record found:                                │
│          - Return 403 Forbidden                             │
│          - "User not member of organization"                │
│                                                             │
│  Step 4: If record found:                                   │
│          - Extract user's role in this org                  │
│          - Store in context                                 │
│          - Allow request to proceed                         │
│                                                             │
│  Purpose:                                                   │
│  - Prevents users from accessing orgs they don't belong to │
│  - Ensures all org-scoped data is accessed correctly        │
│  - Enables multi-tenancy security                          │
└─────────────────────────────────────────────────────────────┘

DATABASE-LEVEL ISOLATION
┌─────────────────────────────────────────────────────────────┐
│  All organization-scoped queries filter by organization_id  │
│                                                             │
│  Example Query:                                             │
│  db.Where("organization_id = ?", orgID).                   │
│     Find(&requisitions)                                     │
│                                                             │
│  Query Pattern:                                             │
│  SELECT * FROM requisitions                                │
│  WHERE organization_id = '12345'                            │
│         ↑                     ↑                              │
│         Required filter       Organization scope            │
│                                                             │
│  Defense:                                                   │
│  - Even if permission check fails, query is scoped          │
│  - Even if middleware is bypassed, data is scoped           │
│  - Defense in depth: Multiple layers                        │
│                                                             │
│  What This Prevents:                                        │
│  X: User trying to query another organization's data        │
│    - Middleware catches it first                            │
│    - Query returns no results anyway                        │
│                                                             │
│  X: Permission check bypassed somehow                       │
│    - Query still scoped to org                              │
│    - No data leakage                                        │
└─────────────────────────────────────────────────────────────┘

ORGANIZATION SWITCHING
┌─────────────────────────────────────────────────────────────┐
│  User can work in multiple organizations:                   │
│                                                             │
│  Default:                                                   │
│  - User.CurrentOrganizationID determines org scope          │
│  - JWT includes currentOrgId                                │
│  - All requests use this org                                │
│                                                             │
│  Explicit Switch:                                           │
│  1. Send X-Organization-ID: org-uuid header                 │
│  2. TenantMiddleware uses header instead                    │
│  3. Membership verified                                     │
│  4. Request scoped to that org                              │
│                                                             │
│  Frontend Switcher:                                         │
│  - OrganizationContext provides switchWorkspace()          │
│  - Updates User.CurrentOrganizationID                       │
│  - Invalidates all queries                                  │
│  - Refreshes UI for new organization                        │
└─────────────────────────────────────────────────────────────┘
```

---

## Data Flows

### Complete Request Flow (Authentication → Authorization → Data Access)

```
USER MAKES API REQUEST
├─ Include JWT in Authorization header
├─ Include X-Organization-ID header (optional)
└─ Send to protected endpoint

         ↓

BACKEND RECEIVES REQUEST
    └─ Fiber framework

         ↓

MIDDLEWARE CHAIN (Left to Right)
┌─────────────────────────────────────────────────────────────┐

1️⃣  CORS MIDDLEWARE
    ├─ Check request origin
    └─ Allow if matches FRONTEND_URL

2️⃣  AUTH MIDDLEWARE (utils/jwt.go)
    ├─ Parse Authorization header
    ├─ Extract JWT token
    ├─ Call ValidateToken()
    ├─ Verify signature with JWT_SECRET
    ├─ Check expiration
    ├─ Extract CustomClaims
    ├─ Set c.Locals("userID", claims.Subject)
    ├─ Set c.Locals("userRole", claims.Role)
    └─ If fails: Return 401 Unauthorized

3️⃣  TENANT MIDDLEWARE (middleware/tenant.go)
    ├─ Get organization ID:
    │  ├─ From X-Organization-ID header, OR
    │  └─ From JWT.currentOrgId (fallback)
    ├─ Query: SELECT * FROM organization_members
    │          WHERE user_id = ? AND organization_id = ?
    ├─ Verify active membership
    ├─ Set c.Locals("organizationID", orgID)
    └─ If fails: Return 403 Forbidden

4️⃣  PERMISSION MIDDLEWARE (middleware/middleware.go)
    ├─ Get userID, userRole, organizationID from context
    ├─ Create PermissionService
    ├─ Call HasPermission(userID, orgID, role, resource, action)
    │  ├─ PHASE 1: Check database (custom permissions)
    │  │   └─ Query OrganizationRole → PermissionAssignment
    │  └─ PHASE 2: Check hardcoded (system permissions)
    │      └─ Look up role in RolePermissions map
    ├─ If granted: Continue to handler
    ├─ Log permission decision to AuditLog
    └─ If denied: Return 403 Forbidden

└─────────────────────────────────────────────────────────────┘

         ↓

HANDLER FUNCTION EXECUTES
    ├─ Access user context: c.Locals("userID")
    ├─ Access org context: c.Locals("organizationID")
    ├─ Query database (automatically scoped by org)
    └─ Return data response

         ↓

RESPONSE SENT TO FRONTEND
    ├─ Status 200 OK
    ├─ Body: Organization-scoped data
    └─ Cookie: Session refreshed (optional)
```

### Organization Switching Flow

```
USER CLICKS ORGANIZATION SWITCHER
    ├─ Current: "My Organization"
    └─ User selects: "Client Project"

         ↓

FRONTEND: OrganizationContext
    ├─ Call switchWorkspace("client-project-org-id")
    ├─ Update User.CurrentOrganizationID
    ├─ Store in localStorage
    ├─ Invalidate React Query cache
    └─ Refresh all components

         ↓

NEXT REQUEST TO BACKEND
    ├─ JWT still has old currentOrgId (unchanged)
    ├─ Send X-Organization-ID: client-project-org-id
    └─ TenantMiddleware uses header instead

         ↓

ORGANIZATION ISOLATION IN ACTION
    ├─ Verify user is member of "Client Project" org
    ├─ Query data filtered by "Client Project" org ID
    ├─ Return data for selected organization
    └─ User can't access previous organization data
```

---

## File References

### Authentication Files

| File | Lines | Purpose |
|------|-------|---------|
| `backend/handlers/auth.go` | 1-397 | Login, register, token endpoints |
| `backend/utils/jwt.go` | 1-100+ | JWT generation and validation |
| `backend/utils/password.go` | 1-80+ | Password hashing and verification |
| `backend/models/auth.go` | - | Auth-related models |
| `backend/services/auth_service.go` | 1-400+ | Token blacklist, login tracking |
| `frontend/src/app/_actions/auth.ts` | 1-400+ | Server actions for auth |
| `frontend/src/lib/auth.ts` | 1-200+ | Session encryption/decryption |

### Authorization & Permissions Files

| File | Lines | Purpose |
|------|-------|---------|
| `backend/services/permission_service.go` | 1-300+ | Permission checking logic |
| `backend/services/role_management_service.go` | 1-350+ | Custom role CRUD |
| `backend/middleware/middleware.go` | 1-300+ | Permission enforcement |
| `backend/models/organization.go` | - | Role and permission models |
| `frontend/src/hooks/use-permissions.ts` | 1-150+ | Frontend permission hooks |
| `frontend/src/components/auth/permission-guard.tsx` | - | Permission guard component |

### Organization & Multi-Tenancy Files

| File | Lines | Purpose |
|------|-------|---------|
| `backend/models/models.go` | - | User model structure |
| `backend/models/organization.go` | - | Organization models |
| `backend/services/organization_service.go` | 1-200+ | Organization CRUD |
| `backend/handlers/organizations.go` | 1-300+ | Organization endpoints |
| `backend/middleware/tenant.go` | 1-100+ | Tenant isolation middleware |
| `frontend/src/contexts/organization-context.tsx` | 1-200+ | Organization context |

---

## Security Best Practices Implemented

✅ **Password Security**
- Bcrypt hashing with cost factor 10
- Password strength requirements (8+ chars, mixed case, digit)
- Never stored in plaintext
- Never logged

✅ **JWT Token Security**
- HS256 signature with strong secret (32+ chars)
- 24-hour expiration
- Unique JTI per token for revocation
- Algorithm validation (prevent switching)

✅ **Session Security**
- HTTP-only cookies (XSS protection)
- Encrypted tokens
- 24-hour expiration
- SameSite attribute (CSRF protection)

✅ **Organization Isolation**
- TenantMiddleware verifies membership
- Database-level filtering by org_id
- Defense in depth (multiple layers)
- X-Organization-ID header for explicit switching

✅ **Permission Validation**
- Two-phase checking (custom + hardcoded)
- Per-route enforcement
- Audit logging
- Explicit 403 Forbidden responses

✅ **Audit Trail**
- LoginAttempt logging
- AuditLog for permission changes
- Timestamps on all operations
- IP address and user agent tracking

---

## Summary

The Liyali Gateway authentication, authorization, and permissions system is a comprehensive, multi-layered implementation that provides:

1. **Secure Authentication** - Bcrypt + JWT with proper token management
2. **Flexible Authorization** - Dual-level roles (system + organization)
3. **Fine-grained Permissions** - Resource+Action permission model
4. **Multi-Tenancy** - Complete organization isolation
5. **Audit Trail** - Security event logging
6. **Security Hardening** - Session encryption, CSRF protection, XSS prevention

The system scales from simple role-based access (Phase 3) to custom organization-specific roles and permissions (Phase 3.5), with token revocation and advanced security features in Phase 4.

