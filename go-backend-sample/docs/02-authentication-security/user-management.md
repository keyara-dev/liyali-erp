# User Management & Role Development Plan

**Project**: Liyali Gateway Backend
**Purpose**: Comprehensive plan for user accounts, roles, and permissions
**Status**: Planning Phase
**Last Updated**: December 23, 2025

---

## 📋 Table of Contents

1. [Overview](#overview)
2. [User Roles & Hierarchy](#user-roles--hierarchy)
3. [Permission Matrix](#permission-matrix)
4. [User Lifecycle](#user-lifecycle)
5. [Authentication Flow](#authentication-flow)
6. [Database Schema](#database-schema)
7. [API Endpoints](#api-endpoints)
8. [Development Phases](#development-phases)
9. [Testing Strategy](#testing-strategy)
10. [Security Considerations](#security-considerations)

---

## 🎯 Overview

### What We're Building

A complete user management system that supports:
- **7 distinct user roles** with different permission levels
- **OAuth 2.0 authentication** (Entra ID, Google, GitHub)
- **Role-based access control (RBAC)** at API and UI levels
- **Department-based organization**
- **Session management** with JWT tokens
- **Audit trail** for all user actions

### Goals

- [ ] Every user has a clear role with defined permissions
- [ ] Users can only access features allowed by their role
- [ ] Authentication is secure and uses industry standards
- [ ] User management is simple for administrators
- [ ] All user actions are logged for compliance

---

## 👥 User Roles & Hierarchy

### The 7 User Roles

```
                    ┌─────────────┐
                    │    ADMIN    │ (System Administrator)
                    └──────┬──────┘
                           │
        ┌──────────────────┼──────────────────┐
        │                  │                  │
   ┌────▼─────┐     ┌─────▼──────┐    ┌─────▼────────┐
   │   CFO    │     │ DIRECTOR   │    │ COMPLIANCE   │
   └────┬─────┘     └─────┬──────┘    │   OFFICER    │
        │                 │            └──────────────┘
        └────────┬────────┘
                 │
        ┌────────┼────────┐
        │                 │
   ┌────▼─────────┐  ┌───▼──────────────┐
   │  FINANCE     │  │   DEPARTMENT     │
   │  OFFICER     │  │    MANAGER       │
   └──────────────┘  └──────────────────┘
                           │
                     ┌─────▼─────┐
                     │   USER    │ (Regular Employee)
                     └───────────┘
```

---

### 1. ADMIN (System Administrator)

**Purpose**: Manage the entire system

**Capabilities**:
- ✅ Full system access
- ✅ User management (create, update, delete, assign roles)
- ✅ Workflow management (create, edit, delete workflows)
- ✅ System configuration
- ✅ View all documents and approvals
- ✅ Override any approval
- ✅ Access audit logs
- ✅ Generate system reports

**Use Cases**:
- IT administrators
- System owners
- Technical support staff

**Restrictions**:
- Should not handle day-to-day approvals (delegate to appropriate roles)
- Must follow change management procedures for system changes

**Example Users**:
- John Admin - IT Department Head
- Sarah SysAdmin - Technical Support Lead

---

### 2. CFO (Chief Financial Officer)

**Purpose**: Final financial authority

**Capabilities**:
- ✅ Approve/reject at Stage 3 (final stage)
- ✅ View all financial documents
- ✅ Access all analytics and reports
- ✅ Override rejections (with audit trail)
- ✅ Set budget limits
- ✅ Approve high-value transactions (>K100,000)
- ✅ Review compliance reports

**Use Cases**:
- Final approval for large purchases
- Budget allocation decisions
- Financial policy enforcement

**Restrictions**:
- Cannot modify workflows (use ADMIN role)
- Cannot delete approved documents

**Example Users**:
- Robert CFO - Chief Financial Officer
- Mary Deputy-CFO - Deputy Chief Financial Officer

---

### 3. DIRECTOR

**Purpose**: Department or division leadership

**Capabilities**:
- ✅ Approve/reject at Stage 3
- ✅ View all documents from their division
- ✅ Access division analytics
- ✅ Assign department managers
- ✅ Set division budgets
- ✅ Escalate issues to CFO

**Use Cases**:
- Division heads
- Regional managers
- Senior leadership

**Restrictions**:
- Limited to their division/department
- Cannot override CFO decisions

**Example Users**:
- James Director - Sales Director
- Lisa Director - Operations Director

---

### 4. FINANCE_OFFICER

**Purpose**: Financial review and verification

**Capabilities**:
- ✅ Approve/reject at Stage 2
- ✅ Verify budget availability
- ✅ Check GL codes
- ✅ Review vendor compliance
- ✅ Generate financial reports
- ✅ Request additional documentation
- ✅ View all requisitions and POs

**Use Cases**:
- Finance department staff
- Budget analysts
- Financial controllers

**Restrictions**:
- Cannot approve above budget limits without CFO approval
- Cannot modify budgets (view only)

**Example Users**:
- Peter Finance - Senior Finance Officer
- Jane Finance - Budget Analyst

---

### 5. DEPARTMENT_MANAGER

**Purpose**: First-line approval for department requests

**Capabilities**:
- ✅ Approve/reject at Stage 1
- ✅ Submit requisitions on behalf of team
- ✅ View department documents
- ✅ Reassign within department
- ✅ Access department analytics
- ✅ Manage department users

**Use Cases**:
- Department heads
- Team leads
- Project managers

**Restrictions**:
- Limited to their department
- Cannot approve above department budget without escalation
- Cannot skip approval stages

**Example Users**:
- Mike Manager - Sales Department Manager
- Susan Manager - IT Department Manager

---

### 6. COMPLIANCE_OFFICER

**Purpose**: Audit and compliance monitoring

**Capabilities**:
- ✅ View all documents (read-only)
- ✅ Access complete audit logs
- ✅ Generate compliance reports
- ✅ Flag non-compliant transactions
- ✅ Request document reviews
- ⚠️ Cannot approve/reject (observer role)

**Use Cases**:
- Internal auditors
- Compliance team
- Risk management

**Restrictions**:
- Read-only access (cannot modify)
- Cannot approve or reject
- Cannot create documents

**Example Users**:
- David Compliance - Chief Compliance Officer
- Rachel Audit - Internal Auditor

---

### 7. USER (Regular Employee)

**Purpose**: Submit requests and view own documents

**Capabilities**:
- ✅ Create requisitions
- ✅ View own documents
- ✅ Edit draft documents
- ✅ Submit for approval
- ✅ View approval status
- ⚠️ Cannot approve (must be escalated)

**Use Cases**:
- Regular employees
- Contractors
- Temporary staff

**Restrictions**:
- Cannot view other users' documents
- Cannot approve any stage
- Limited to basic CRUD operations

**Example Users**:
- Tom Employee - Sales Representative
- Alice Employee - Marketing Coordinator

---

## 🔐 Permission Matrix

### Complete Permission Table

| Permission | ADMIN | CFO | DIRECTOR | FINANCE | MANAGER | COMPLIANCE | USER |
|------------|-------|-----|----------|---------|---------|------------|------|
| **Documents** |
| Create Document | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ✅ |
| View Own Documents | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ✅ |
| View All Documents | ✅ | ✅ | ✅ | ✅ | ⚠️ Dept | ✅ | ❌ |
| Edit Draft Documents | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ✅ |
| Delete Documents | ✅ | ❌ | ❌ | ❌ | ⚠️ Own | ❌ | ⚠️ Own |
| Submit for Approval | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ✅ |
| **Approvals** |
| Approve Stage 1 | ✅ | ❌ | ❌ | ❌ | ✅ | ❌ | ❌ |
| Approve Stage 2 | ✅ | ❌ | ❌ | ✅ | ❌ | ❌ | ❌ |
| Approve Stage 3 | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ |
| Reject Any Stage | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ |
| Reassign Tasks | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ |
| Bulk Approve | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ |
| Override Decisions | ✅ | ⚠️ Limited | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Workflows** |
| View Workflows | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Create Workflows | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| Edit Workflows | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| Delete Workflows | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Users** |
| View Users | ✅ | ✅ | ⚠️ Dept | ⚠️ Dept | ⚠️ Dept | ✅ | ❌ |
| Create Users | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| Edit Users | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ⚠️ Own |
| Delete Users | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| Assign Roles | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Analytics** |
| View Own Analytics | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| View Dept Analytics | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ |
| View All Analytics | ✅ | ✅ | ✅ | ✅ | ❌ | ✅ | ❌ |
| Export Reports | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ |
| **Audit** |
| View Audit Logs | ✅ | ⚠️ Limited | ❌ | ❌ | ❌ | ✅ | ❌ |
| Export Audit Logs | ✅ | ❌ | ❌ | ❌ | ❌ | ✅ | ❌ |
| **System** |
| System Configuration | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| Backup/Restore | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |

**Legend**:
- ✅ = Full access
- ❌ = No access
- ⚠️ = Limited/conditional access

---

## 🔄 User Lifecycle

### 1. User Creation

**Process**:
```
1. ADMIN receives new employee information
   ↓
2. ADMIN creates user account via admin panel
   ↓
3. System sends email with OAuth login instructions
   ↓
4. User receives welcome email
   ↓
5. User clicks OAuth login link
   ↓
6. User authenticates via Entra ID/Google/GitHub
   ↓
7. System creates session and JWT token
   ↓
8. User redirected to dashboard
```

**Required Information**:
- Email address (primary identifier)
- Full name
- Role assignment
- Department
- Manager (if applicable)
- Start date

**API Endpoint**: `POST /api/users`

---

### 2. Role Assignment

**Process**:
```
1. ADMIN assesses user responsibilities
   ↓
2. ADMIN assigns appropriate role
   ↓
3. System grants permissions automatically
   ↓
4. User receives role confirmation email
   ↓
5. Audit log records role assignment
```

**Role Change Scenarios**:
- **Promotion**: USER → MANAGER → FINANCE → DIRECTOR → CFO
- **Lateral Move**: MANAGER (Sales) → MANAGER (IT)
- **Special Assignment**: MANAGER → COMPLIANCE (temporary audit role)

**API Endpoint**: `PUT /api/users/:id/role`

---

### 3. Active User Management

**Daily Activities**:
- Login/logout tracking
- Session management
- Permission checks
- Activity logging

**Periodic Reviews**:
- Quarterly access reviews (compliance requirement)
- Annual role reassessment
- Inactive user detection (90 days)

**API Endpoints**:
- `GET /api/users/me` - Current user profile
- `POST /api/users/:id/deactivate` - Deactivate user
- `POST /api/users/:id/reactivate` - Reactivate user

---

### 4. User Offboarding

**Process**:
```
1. HR notifies ADMIN of employee departure
   ↓
2. ADMIN deactivates user account
   ↓
3. System revokes all sessions/tokens
   ↓
4. User cannot login
   ↓
5. Documents remain assigned to user (read-only)
   ↓
6. After 90 days: ADMIN reviews for deletion
   ↓
7. If approved: ADMIN deletes user (soft delete)
   ↓
8. Audit logs preserved permanently
```

**API Endpoint**: `DELETE /api/users/:id`

---

## 🔑 Authentication Flow

### OAuth 2.0 Login Flow

```
┌─────────────┐
│   Browser   │
└──────┬──────┘
       │ 1. Click "Login with Microsoft"
       ↓
┌──────────────┐
│   Frontend   │
└──────┬───────┘
       │ 2. GET /api/auth/login?provider=entra_id
       ↓
┌──────────────┐
│   Backend    │
└──────┬───────┘
       │ 3. Redirect to Entra ID OAuth
       ↓
┌──────────────┐
│  Entra ID    │ (Microsoft)
└──────┬───────┘
       │ 4. User authenticates
       ↓
       │ 5. Callback with auth code
       ↓
┌──────────────┐
│   Backend    │
└──────┬───────┘
       │ 6. Exchange code for access token
       │ 7. Fetch user profile from Entra ID
       │ 8. Check if user exists in DB
       │ 9. Create/update user record
       │ 10. Generate JWT token
       │ 11. Create session record
       ↓
┌──────────────┐
│   Frontend   │
└──────┬───────┘
       │ 12. Store JWT in localStorage
       │ 13. Redirect to dashboard
       ↓
┌─────────────┐
│  Dashboard  │
└─────────────┘
```

### Session Management

**JWT Token Structure**:
```json
{
  "user_id": "uuid",
  "email": "john.manager@company.com",
  "name": "John Manager",
  "role": "DEPARTMENT_MANAGER",
  "department": "Sales",
  "permissions": [
    "approve_stage_1",
    "view_department_docs",
    "create_requisitions"
  ],
  "iat": 1735567200,
  "exp": 1735570800
}
```

**Token Expiration**:
- Access Token: 1 hour
- Refresh Token: 8 hours
- Idle Timeout: 1 hour of inactivity

**Session Storage**:
```sql
sessions
├── id (uuid)
├── user_id (uuid, foreign key)
├── token (jwt token hash)
├── expires_at (timestamp)
├── created_at (timestamp)
└── last_activity (timestamp)
```

---

## 💾 Database Schema

### Users Table

```sql
CREATE TABLE users (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email             VARCHAR(255) UNIQUE NOT NULL,
  name              VARCHAR(255) NOT NULL,
  role              VARCHAR(50) NOT NULL,  -- ADMIN, CFO, DIRECTOR, etc.
  department        VARCHAR(100),
  manager_id        UUID REFERENCES users(id),
  is_active         BOOLEAN DEFAULT true,
  last_login        TIMESTAMP,
  created_at        TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at        TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  created_by        UUID REFERENCES users(id),

  -- Indexes
  INDEX idx_email (email),
  INDEX idx_role (role),
  INDEX idx_department (department),
  INDEX idx_is_active (is_active),

  -- Constraints
  CHECK (role IN ('ADMIN', 'CFO', 'DIRECTOR', 'FINANCE_OFFICER',
                  'DEPARTMENT_MANAGER', 'COMPLIANCE_OFFICER', 'USER'))
);
```

### User Profiles Table (Extended Info)

```sql
CREATE TABLE user_profiles (
  user_id           UUID PRIMARY KEY REFERENCES users(id),
  phone             VARCHAR(20),
  employee_id       VARCHAR(50),
  job_title         VARCHAR(100),
  office_location   VARCHAR(100),
  start_date        DATE,
  end_date          DATE,
  notes             TEXT,
  avatar_url        TEXT,

  created_at        TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at        TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Sessions Table

```sql
CREATE TABLE sessions (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id           UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token             VARCHAR(500) UNIQUE NOT NULL,
  expires_at        TIMESTAMP NOT NULL,
  last_activity     TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  ip_address        VARCHAR(45),
  user_agent        TEXT,
  created_at        TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

  INDEX idx_user_id (user_id),
  INDEX idx_token (token),
  INDEX idx_expires_at (expires_at)
);
```

### User Audit Log

```sql
CREATE TABLE user_audit_log (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id           UUID REFERENCES users(id),
  action            VARCHAR(50) NOT NULL,  -- LOGIN, LOGOUT, ROLE_CHANGE, etc.
  details           JSONB,
  ip_address        VARCHAR(45),
  user_agent        TEXT,
  performed_by      UUID REFERENCES users(id),
  timestamp         TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

  INDEX idx_user_id (user_id),
  INDEX idx_action (action),
  INDEX idx_timestamp (timestamp)
);
```

---

## 🔌 API Endpoints

### User Management

#### GET /api/users
List all users (ADMIN only)

**Query Parameters**:
- `role` - Filter by role
- `department` - Filter by department
- `is_active` - Filter by active status
- `page`, `limit` - Pagination

**Response**:
```json
{
  "success": true,
  "data": {
    "users": [
      {
        "id": "uuid",
        "email": "john.manager@company.com",
        "name": "John Manager",
        "role": "DEPARTMENT_MANAGER",
        "department": "Sales",
        "is_active": true,
        "last_login": "2025-12-23T09:00:00Z"
      }
    ],
    "total": 45,
    "page": 1,
    "limit": 20
  }
}
```

---

#### GET /api/users/me
Get current user profile

**Response**:
```json
{
  "success": true,
  "data": {
    "user": {
      "id": "uuid",
      "email": "john.manager@company.com",
      "name": "John Manager",
      "role": "DEPARTMENT_MANAGER",
      "department": "Sales",
      "manager": {
        "id": "uuid",
        "name": "Mary Director"
      },
      "permissions": [
        "approve_stage_1",
        "view_department_docs",
        "create_requisitions",
        "reassign_tasks"
      ],
      "profile": {
        "phone": "+260-123-456-789",
        "job_title": "Sales Department Manager",
        "office_location": "Lusaka Office"
      }
    }
  }
}
```

---

#### POST /api/users
Create new user (ADMIN only)

**Request**:
```json
{
  "email": "new.employee@company.com",
  "name": "New Employee",
  "role": "USER",
  "department": "Sales",
  "manager_id": "uuid",
  "profile": {
    "phone": "+260-123-456-789",
    "job_title": "Sales Representative",
    "start_date": "2025-01-01"
  }
}
```

**Response**:
```json
{
  "success": true,
  "data": {
    "user": {
      "id": "new_uuid",
      "email": "new.employee@company.com",
      "name": "New Employee",
      "role": "USER",
      "department": "Sales",
      "is_active": true,
      "created_at": "2025-12-23T10:00:00Z"
    }
  }
}
```

---

#### PUT /api/users/:id
Update user (ADMIN only, or self for limited fields)

**Request**:
```json
{
  "name": "Updated Name",
  "department": "Marketing",
  "profile": {
    "phone": "+260-987-654-321"
  }
}
```

---

#### PUT /api/users/:id/role
Change user role (ADMIN only)

**Request**:
```json
{
  "role": "DEPARTMENT_MANAGER",
  "reason": "Promotion to department manager"
}
```

**Response**:
```json
{
  "success": true,
  "data": {
    "user": {
      "id": "uuid",
      "role": "DEPARTMENT_MANAGER",
      "previous_role": "USER",
      "changed_at": "2025-12-23T10:00:00Z"
    }
  }
}
```

---

#### POST /api/users/:id/deactivate
Deactivate user (ADMIN only)

**Request**:
```json
{
  "reason": "Employee left company",
  "effective_date": "2025-12-31"
}
```

---

#### DELETE /api/users/:id
Delete user (ADMIN only, soft delete)

**Response**: 204 No Content

---

### Role & Permission Management

#### GET /api/roles
List all available roles

**Response**:
```json
{
  "success": true,
  "data": {
    "roles": [
      {
        "name": "ADMIN",
        "display_name": "System Administrator",
        "description": "Full system access",
        "permission_count": 50
      },
      {
        "name": "CFO",
        "display_name": "Chief Financial Officer",
        "description": "Final financial authority",
        "permission_count": 35
      }
      // ... other roles
    ]
  }
}
```

---

#### GET /api/roles/:role/permissions
Get permissions for a specific role

**Response**:
```json
{
  "success": true,
  "data": {
    "role": "DEPARTMENT_MANAGER",
    "permissions": [
      {
        "name": "approve_stage_1",
        "description": "Approve documents at stage 1"
      },
      {
        "name": "view_department_docs",
        "description": "View all department documents"
      },
      {
        "name": "create_requisitions",
        "description": "Create purchase requisitions"
      }
    ]
  }
}
```

---

#### POST /api/users/:id/check-permission
Check if user has specific permission

**Request**:
```json
{
  "permission": "approve_stage_1"
}
```

**Response**:
```json
{
  "success": true,
  "data": {
    "has_permission": true,
    "permission": "approve_stage_1",
    "granted_by": "DEPARTMENT_MANAGER"
  }
}
```

---

## 📅 Development Phases

### Phase 1: User Model & Database (2 days)

**Tasks**:
- [ ] Create users table with all fields
- [ ] Create user_profiles table
- [ ] Create sessions table
- [ ] Create user_audit_log table
- [ ] Define GORM models
- [ ] Create migrations
- [ ] Seed initial admin user

**Deliverables**:
- Database schema complete
- Models working with GORM
- 1 admin user seeded

---

### Phase 2: Authentication Setup (3 days)

**Tasks**:
- [ ] Set up OAuth 2.0 providers (Entra ID, Google, GitHub)
- [ ] Implement JWT token generation
- [ ] Create auth middleware
- [ ] Implement session management
- [ ] Create login/logout endpoints
- [ ] Test authentication flow

**Deliverables**:
- OAuth login working
- JWT tokens issued
- Sessions tracked

---

### Phase 3: Role & Permission System (3 days)

**Tasks**:
- [ ] Define permission constants
- [ ] Create permission checking middleware
- [ ] Implement role-based access control
- [ ] Create permission checking functions
- [ ] Add permission checks to existing endpoints
- [ ] Test permission enforcement

**Deliverables**:
- RBAC middleware functional
- All endpoints protected
- Permission tests passing

---

### Phase 4: User Management APIs (2 days)

**Tasks**:
- [ ] Create user CRUD endpoints
- [ ] Implement user listing with filters
- [ ] Add role assignment endpoint
- [ ] Create user profile endpoints
- [ ] Implement user search
- [ ] Test all endpoints

**Deliverables**:
- 10+ user management endpoints
- Admin panel can manage users
- Tests passing

---

### Phase 5: User Audit & Logging (1 day)

**Tasks**:
- [ ] Implement user action logging
- [ ] Create audit log endpoints
- [ ] Add login/logout tracking
- [ ] Implement role change logging
- [ ] Test audit trail completeness

**Deliverables**:
- Complete audit trail
- Audit log API working
- Compliance requirements met

---

## 🧪 Testing Strategy

### Unit Tests

**User Model Tests**:
```go
func TestUserCreation(t *testing.T) {
    user := &User{
        Email: "test@example.com",
        Name: "Test User",
        Role: "USER",
    }

    err := db.Create(user).Error
    assert.NoError(t, err)
    assert.NotEmpty(t, user.ID)
}

func TestRoleValidation(t *testing.T) {
    user := &User{
        Email: "test@example.com",
        Role: "INVALID_ROLE",
    }

    err := db.Create(user).Error
    assert.Error(t, err)
}
```

---

### Integration Tests

**Permission Check Tests**:
```go
func TestManagerCanApproveStage1(t *testing.T) {
    user := createTestUser("DEPARTMENT_MANAGER")
    hasPermission := checkPermission(user, "approve_stage_1")
    assert.True(t, hasPermission)
}

func TestUserCannotApprove(t *testing.T) {
    user := createTestUser("USER")
    hasPermission := checkPermission(user, "approve_stage_1")
    assert.False(t, hasPermission)
}
```

---

### E2E Tests

**Complete User Journey**:
```go
func TestCompleteUserLifecycle(t *testing.T) {
    // 1. Admin creates user
    user := adminCreateUser(t, "new@example.com", "USER")

    // 2. User logs in
    token := userLogin(t, "new@example.com")

    // 3. User creates requisition
    req := userCreateRequisition(t, token)

    // 4. Admin promotes user to manager
    adminUpdateUserRole(t, user.ID, "DEPARTMENT_MANAGER")

    // 5. User can now approve
    canApprove := checkPermission(user, "approve_stage_1")
    assert.True(t, canApprove)

    // 6. Admin deactivates user
    adminDeactivateUser(t, user.ID)

    // 7. User cannot login
    _, err := userLogin(t, "new@example.com")
    assert.Error(t, err)
}
```

---

## 🔒 Security Considerations

### Password Security
- ✅ Use OAuth 2.0 (no passwords stored)
- ✅ Fallback: bcrypt hashing (cost 12)
- ✅ Password reset via email

### Token Security
- ✅ JWT tokens with 1-hour expiration
- ✅ Refresh tokens with 8-hour expiration
- ✅ Token revocation on logout
- ✅ Session invalidation on role change

### Permission Security
- ✅ Check permissions at API level (not just UI)
- ✅ Deny by default
- ✅ Audit all permission checks
- ✅ Log all failed access attempts

### Data Security
- ✅ Encrypt sensitive data at rest
- ✅ HTTPS only in production
- ✅ Rate limiting on auth endpoints
- ✅ Account lockout after failed attempts

---

## 📊 Success Metrics

### User Management
- [ ] 100% of users have valid roles
- [ ] 0 users with excessive permissions
- [ ] < 5% inactive users (cleanup)
- [ ] 100% audit coverage

### Authentication
- [ ] < 1 second login time
- [ ] 99.9% authentication uptime
- [ ] 0 security breaches
- [ ] < 0.1% failed login rate

### Permissions
- [ ] 100% API endpoints protected
- [ ] 0 unauthorized access incidents
- [ ] < 50ms permission check time
- [ ] 100% permission tests passing

---

## 📚 References

- [OAuth 2.0 Specification](https://oauth.net/2/)
- [JWT Best Practices](https://tools.ietf.org/html/rfc8725)
- [RBAC Implementation Guide](https://csrc.nist.gov/projects/role-based-access-control)

---

**Last Updated**: December 23, 2025
**Version**: 1.0
**Status**: Ready for Implementation
**Next Step**: Begin Phase 1 - User Model & Database
