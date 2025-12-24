# Multi-Tenancy Refactor Plan - Liyali Gateway

**Document**: Multi-Tenant Architecture Implementation Plan (Slack-like model)
**Status**: Planning Phase
**Created**: 2025-12-15
**Target Phase**: Phase 13 (Post Phase 12 Database Integration)
**Complexity**: High
**Estimated Effort**: 400-600 hours (6-10 weeks, 2-3 full-time developers)

---

## Executive Summary

This document outlines a comprehensive plan to transform Liyali Gateway from a single-tenant system into a **multi-tenant SaaS platform** following the **Slack organizational model**:

### Multi-Tenancy Model (Slack-like)
```
Organizations (Companies/Institutions)
├── Each organization has its own isolated data
├── Users can belong to multiple organizations
├── User activity is audited per organization
├── Organization settings and configurations
└── Each user has different roles per organization
```

### Current State
- ✅ Authentication: JWT-based, working
- ✅ Authorization: Role-based (5 roles), basic
- ✅ Database: Normalized schema, well-designed
- ❌ Multi-tenancy: **Completely absent**
- ❌ Data isolation: **No organization scoping**
- ❌ Organization model: **Does not exist**
- ❌ User-organization relationship: **Not modeled**

### Target State
- ✅ Full multi-tenant data isolation
- ✅ User can belong to multiple organizations
- ✅ Organization-level settings and customization
- ✅ Complete audit trails per organization
- ✅ Role management per user per organization
- ✅ Automatic query filtering by organization context
- ✅ Workspace switcher UI (like Slack)
- ✅ Organization invitation and member management

---

## Part 1: Architecture Overview

### 1.1 Data Isolation Model

```
┌─────────────────────────────────────────────────────┐
│ Liyali Gateway (SaaS Platform)                      │
├─────────────────────────────────────────────────────┤
│                                                       │
│  Org A (Company X)    Org B (Ministry Y)            │
│  ├─ Users            ├─ Users                        │
│  ├─ Requisitions    ├─ Requisitions                  │
│  ├─ POs             ├─ POs                           │
│  ├─ Vendors         ├─ Vendors                       │
│  └─ Settings        └─ Settings                      │
│                                                       │
│  User Sarah:        User Rajesh:                     │
│  ├─ Org A (Admin)   ├─ Org B (Finance)              │
│  └─ Org B (Member)  └─ Org A (Approver)             │
│                                                       │
└─────────────────────────────────────────────────────┘
```

### 1.2 Key Principles

1. **Complete Data Isolation**: No data leakage between organizations
2. **User Portability**: One user can access multiple organizations
3. **Role Per Organization**: Different roles in different organizations
4. **Audit Everything**: All actions tied to organization context
5. **Tenant Context**: Every query automatically scoped to tenant
6. **Backward Compatibility**: Existing data preserved during migration

### 1.3 User-Organization Relationships

```
Users (Global)
├─ User 1 (Sarah)
│  ├─ OrganizationMembership → Org A (role: Admin)
│  ├─ OrganizationMembership → Org B (role: Approver)
│  └─ OrganizationMembership → Org C (role: Viewer)
│
├─ User 2 (John)
│  ├─ OrganizationMembership → Org A (role: Requester)
│  └─ OrganizationMembership → Org B (role: Admin)
│
└─ User 3 (Maria)
   └─ OrganizationMembership → Org B (role: Finance)
```

---

## Part 2: Database Schema Changes

### 2.1 New Tables to Create

#### Organizations Table
```sql
CREATE TABLE organizations (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

  -- Basic Info
  name VARCHAR(255) NOT NULL,
  slug VARCHAR(100) UNIQUE NOT NULL,  -- For URLs: acme-corp
  description TEXT,

  -- Branding
  logo_url VARCHAR(255),
  website_url VARCHAR(255),
  primary_color VARCHAR(7) DEFAULT '#0066CC',

  -- Organization Details
  org_type VARCHAR(50),  -- company, government, ngo, educational
  industry VARCHAR(100),
  country VARCHAR(100),
  tax_id VARCHAR(50) UNIQUE,

  -- Status & Metadata
  active BOOLEAN DEFAULT true,
  tier VARCHAR(50) DEFAULT 'free',  -- free, pro, enterprise
  subscription_status VARCHAR(50) DEFAULT 'active',  -- active, inactive, suspended

  -- Timestamps
  created_by UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

  -- Indexing
  INDEX idx_slug (slug),
  INDEX idx_active (active),
  INDEX idx_created_by (created_by),
  UNIQUE idx_tax_id (tax_id)
);

-- Organization Settings
CREATE TABLE organization_settings (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  organization_id UUID NOT NULL UNIQUE REFERENCES organizations(id) ON DELETE CASCADE,

  -- Approval Settings
  default_approval_chain VARCHAR(255),
  require_digital_signatures BOOLEAN DEFAULT true,
  signature_type VARCHAR(50) DEFAULT 'digital',  -- digital, biometric, verbal

  -- Financial Settings
  currency VARCHAR(3) DEFAULT 'USD',
  fiscal_year_start INT DEFAULT 1,  -- Month (1-12)
  enable_budget_validation BOOLEAN DEFAULT true,
  budget_variance_threshold DECIMAL(5,2) DEFAULT 5.00,  -- % variance tolerance

  -- Security Settings
  password_policy_enabled BOOLEAN DEFAULT true,
  min_password_length INT DEFAULT 8,
  require_2fa BOOLEAN DEFAULT false,
  session_timeout_minutes INT DEFAULT 30,

  -- Audit Settings
  enable_audit_logging BOOLEAN DEFAULT true,
  audit_retention_days INT DEFAULT 365,

  -- Email Settings
  email_notifications_enabled BOOLEAN DEFAULT true,
  sender_email VARCHAR(255),
  sender_name VARCHAR(255),

  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

  INDEX idx_org_id (organization_id)
);

-- Organization Members (User-Organization Relationship)
CREATE TABLE organization_members (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

  -- Membership Details
  role VARCHAR(50) NOT NULL,  -- admin, manager, approver, requester, viewer
  department VARCHAR(100),     -- Optional department within org
  title VARCHAR(255),          -- Job title

  -- Status
  active BOOLEAN DEFAULT true,
  invited_at TIMESTAMP,
  joined_at TIMESTAMP,
  invited_by UUID REFERENCES users(id) ON DELETE SET NULL,

  -- Permissions Override
  custom_permissions JSONB,     -- Per-user permission overrides

  timestamps
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

  -- Constraints
  UNIQUE(organization_id, user_id),
  INDEX idx_org_id (organization_id),
  INDEX idx_user_id (user_id),
  INDEX idx_org_user (organization_id, user_id),
  INDEX idx_active (active)
);

-- Organization Departments (Optional but recommended)
CREATE TABLE organization_departments (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

  name VARCHAR(255) NOT NULL,
  code VARCHAR(50),
  description TEXT,
  parent_id UUID REFERENCES organization_departments(id) ON DELETE CASCADE,

  budget_allocation DECIMAL(15,2),
  budget_fiscal_year INT,

  active BOOLEAN DEFAULT true,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

  UNIQUE(organization_id, code),
  INDEX idx_org_id (organization_id),
  INDEX idx_parent_id (parent_id)
);
```

#### Audit & Activity Tables
```sql
-- Audit Log (Enhanced with Organization Context)
CREATE TABLE audit_logs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

  -- What happened
  action VARCHAR(100) NOT NULL,
  resource_type VARCHAR(50),
  resource_id UUID,

  -- Who did it
  user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  user_email VARCHAR(255),
  user_ip_address VARCHAR(45),

  -- Changes
  old_values JSONB,
  new_values JSONB,
  changes_summary TEXT,

  -- Context
  user_agent TEXT,
  session_id VARCHAR(255),

  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

  INDEX idx_org_id (organization_id),
  INDEX idx_user_id (user_id),
  INDEX idx_action (action),
  INDEX idx_resource (resource_type, resource_id),
  INDEX idx_created_at (created_at)
);

-- Activity Feed (Denormalized for quick UI rendering)
CREATE TABLE organization_activity_feed (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

  -- Activity Details
  activity_type VARCHAR(50),
  actor_id UUID REFERENCES users(id) ON DELETE SET NULL,
  actor_name VARCHAR(255),
  action_description TEXT,

  -- Related Document
  document_type VARCHAR(50),
  document_id UUID,
  document_number VARCHAR(100),

  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

  INDEX idx_org_id (organization_id),
  INDEX idx_created_at (created_at)
);

-- User Activity Within Organization
CREATE TABLE user_activity_logs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

  -- Activity
  action VARCHAR(100),
  description TEXT,

  -- Timestamps
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

  INDEX idx_org_user (organization_id, user_id),
  INDEX idx_created_at (created_at)
);
```

### 2.2 Modified Existing Tables

#### Add organization_id to All Business Tables

```sql
-- Requisitions
ALTER TABLE requisitions ADD COLUMN organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE;
ALTER TABLE requisitions ADD INDEX idx_org_id (organization_id);
ALTER TABLE requisitions ADD INDEX idx_org_status (organization_id, status);

-- Purchase Orders
ALTER TABLE purchase_orders ADD COLUMN organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE;
ALTER TABLE purchase_orders ADD INDEX idx_org_id (organization_id);

-- Payment Vouchers
ALTER TABLE payment_vouchers ADD COLUMN organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE;
ALTER TABLE payment_vouchers ADD INDEX idx_org_id (organization_id);

-- Goods Received Notes
ALTER TABLE goods_received_notes ADD COLUMN organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE;
ALTER TABLE goods_received_notes ADD INDEX idx_org_id (organization_id);

-- Budgets
ALTER TABLE budgets ADD COLUMN organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE;
ALTER TABLE budgets ADD UNIQUE(organization_id, department, fiscal_year);

-- Categories
ALTER TABLE categories ADD COLUMN organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE;
ALTER TABLE categories ADD INDEX idx_org_id (organization_id);

-- Vendors (Make org-specific)
ALTER TABLE vendors ADD COLUMN organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE;
ALTER TABLE vendors ADD INDEX idx_org_id (organization_id);

-- Approval Tasks
ALTER TABLE approval_tasks ADD COLUMN organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE;
ALTER TABLE approval_tasks ADD INDEX idx_org_id (organization_id);

-- Notifications
ALTER TABLE notifications ADD COLUMN organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE;
ALTER TABLE notifications ADD INDEX idx_org_id (organization_id);

-- Attachment/Files
ALTER TABLE attachments ADD COLUMN organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE;
```

#### Enhance Users Table
```sql
-- Users Table Enhancements
ALTER TABLE users ADD COLUMN current_organization_id UUID REFERENCES organizations(id) ON DELETE SET NULL;
ALTER TABLE users ADD COLUMN preferences JSONB;  -- UI preferences, theme, etc
ALTER TABLE users ADD COLUMN is_super_admin BOOLEAN DEFAULT false;  -- Global platform admin
ALTER TABLE users ADD COLUMN deleted_at TIMESTAMP;  -- Soft delete
```

### 2.3 Schema Relationships Diagram

```
users (Global)
├─ id, email, name, role
├─ current_organization_id → organizations.id
└─ ⬇ many-to-many via organization_members

organizations
├─ id, name, slug, tier
├─ ⬇ one-to-one (organization_settings)
├─ ⬇ one-to-many (organization_members)
├─ ⬇ one-to-many (organization_departments)
└─ ⬇ one-to-many (ALL business tables via organization_id)

organization_members
├─ user_id → users.id
├─ organization_id → organizations.id
├─ role, department, custom_permissions
└─ ⬇ A user can have multiple memberships

organization_departments
├─ organization_id → organizations.id
├─ parent_id → organization_departments.id (hierarchical)
└─ Budget tracking

business_documents (requisitions, POs, PVs, GRNs)
├─ organization_id → organizations.id
├─ created_by → users.id
└─ ⬇ All scoped by organization

audit_logs
├─ organization_id → organizations.id
└─ Tracks all activity per organization
```

---

## Part 3: Backend Implementation Strategy

### 3.1 Middleware & Context Layer

#### Organization Context Middleware
```go
package middleware

// TenantContext holds the current organization context
type TenantContext struct {
    OrganizationID string
    UserID         string
    UserRole       string
    UserDepartment string
}

// ExtractTenantContext extracts org context from JWT token + request
func ExtractTenantContext(c *fiber.Ctx) (*TenantContext, error) {
    // 1. Get user from token
    userID := c.Locals("userID")

    // 2. Get organization from:
    //    a. X-Organization-ID header (for API calls)
    //    b) Authorization Bearer token payload
    //    c) Database lookup from users.current_organization_id
    orgID := c.Get("X-Organization-ID")
    if orgID == "" {
        orgID = getOrgIDFromToken(c)
    }
    if orgID == "" {
        return nil, errors.New("no organization context")
    }

    // 3. Verify user belongs to this organization
    membership, err := getUserOrganizationMembership(userID, orgID)
    if err != nil {
        return nil, errors.New("user not a member of organization")
    }

    return &TenantContext{
        OrganizationID: orgID,
        UserID:        userID,
        UserRole:      membership.Role,
        UserDepartment: membership.Department,
    }, nil
}

// TenantMiddleware enforces tenant context on all protected routes
func TenantMiddleware(c *fiber.Ctx) error {
    tenant, err := ExtractTenantContext(c)
    if err != nil {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "error": "Organization context required",
        })
    }

    c.Locals("tenant", tenant)
    return c.Next()
}
```

#### Automatic Query Scoping
```go
package db

// WithTenant adds automatic organization filtering to queries
func WithTenant(db *gorm.DB, tenant *middleware.TenantContext) *gorm.DB {
    return db.Where("organization_id = ?", tenant.OrganizationID)
}

// Usage in handlers:
// All queries automatically scoped to org
query := WithTenant(db, tenantCtx)
var requisitions []models.Requisition
query.Find(&requisitions)  // Only gets this org's requisitions
```

### 3.2 Handler Refactoring Pattern

#### Before (Single Tenant)
```go
func GetRequisitions(c *fiber.Ctx) error {
    var requisitions []models.Requisition

    query := db.DB
    if status := c.Query("status"); status != "" {
        query = query.Where("status = ?", status)
    }

    query.Find(&requisitions)
    return c.JSON(requisitions)
}
```

#### After (Multi Tenant)
```go
func GetRequisitions(c *fiber.Ctx) error {
    tenant := c.Locals("tenant").(*middleware.TenantContext)

    var requisitions []models.Requisition

    query := db.DB.Where("organization_id = ?", tenant.OrganizationID)

    if status := c.Query("status"); status != "" {
        query = query.Where("status = ?", status)
    }

    if err := query.Find(&requisitions).Error; err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": err.Error(),
        })
    }

    return c.JSON(requisitions)
}
```

### 3.3 Service Layer Refactoring

#### Organization Service
```go
package services

type OrganizationService struct {
    db *gorm.DB
}

// CreateOrganization creates a new organization
func (s *OrganizationService) CreateOrganization(
    ctx context.Context,
    req CreateOrgRequest,
    createdBy string,
) (*models.Organization, error) {
    org := &models.Organization{
        Name:      req.Name,
        Slug:      slug.Make(req.Name),
        CreatedBy: createdBy,
        Active:    true,
    }

    if err := s.db.Create(org).Error; err != nil {
        return nil, err
    }

    // Auto-create settings
    settings := &models.OrganizationSettings{
        OrganizationID: org.ID,
    }
    s.db.Create(settings)

    // Add creator as admin
    s.AddOrganizationMember(ctx, org.ID, createdBy, "admin")

    return org, nil
}

// AddOrganizationMember adds a user to an organization
func (s *OrganizationService) AddOrganizationMember(
    ctx context.Context,
    orgID string,
    userID string,
    role string,
) error {
    membership := &models.OrganizationMember{
        OrganizationID: orgID,
        UserID:        userID,
        Role:          role,
        JoinedAt:      time.Now(),
    }

    return s.db.Create(membership).Error
}

// InviteOrganizationMember invites a user via email
func (s *OrganizationService) InviteOrganizationMember(
    ctx context.Context,
    orgID string,
    email string,
    role string,
    invitedBy string,
) error {
    // Get or create user by email
    var user models.User
    if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            // Create user with random password
            user = models.User{
                Email: email,
                // Generate random password, send invite email
            }
            s.db.Create(&user)
        }
    }

    // Add to organization
    return s.AddOrganizationMember(ctx, orgID, user.ID, role)
}

// RemoveOrganizationMember removes a user from organization
func (s *OrganizationService) RemoveOrganizationMember(
    ctx context.Context,
    orgID string,
    userID string,
) error {
    return s.db.Where("organization_id = ? AND user_id = ?", orgID, userID).
        Delete(&models.OrganizationMember{}).Error
}

// GetOrganizationMembers returns all members of an organization
func (s *OrganizationService) GetOrganizationMembers(
    ctx context.Context,
    orgID string,
) ([]models.OrganizationMember, error) {
    var members []models.OrganizationMember
    err := s.db.Where("organization_id = ?", orgID).
        Preload("User").
        Find(&members).Error
    return members, err
}
```

#### Enhanced User Service
```go
package services

type UserService struct {
    db *gorm.DB
}

// GetUserOrganizations returns all organizations a user belongs to
func (s *UserService) GetUserOrganizations(userID string) ([]models.Organization, error) {
    var organizations []models.Organization
    err := s.db.
        Joins("INNER JOIN organization_members ON organizations.id = organization_members.organization_id").
        Where("organization_members.user_id = ? AND organization_members.active = ?", userID, true).
        Find(&organizations).Error
    return organizations, err
}

// SetCurrentOrganization sets the user's active workspace
func (s *UserService) SetCurrentOrganization(userID, orgID string) error {
    // Verify user is member of org
    var membership models.OrganizationMember
    if err := s.db.Where("user_id = ? AND organization_id = ?", userID, orgID).
        First(&membership).Error; err != nil {
        return errors.New("user not member of organization")
    }

    return s.db.Model(&models.User{}).
        Where("id = ?", userID).
        Update("current_organization_id", orgID).Error
}

// GetUserRole returns user's role in an organization
func (s *UserService) GetUserRole(userID, orgID string) (string, error) {
    var membership models.OrganizationMember
    if err := s.db.Where("user_id = ? AND organization_id = ?", userID, orgID).
        First(&membership).Error; err != nil {
        return "", err
    }
    return membership.Role, nil
}
```

### 3.4 API Routes Structure

#### Organization Management Routes
```go
// Routes: /api/v1/organizations

// Organization CRUD
GET    /api/v1/organizations              → List user's organizations
GET    /api/v1/organizations/:id          → Get organization details
POST   /api/v1/organizations              → Create new organization
PUT    /api/v1/organizations/:id          → Update organization
DELETE /api/v1/organizations/:id          → Delete organization (soft delete)

// Organization Settings
GET    /api/v1/organizations/:id/settings → Get org settings
PUT    /api/v1/organizations/:id/settings → Update org settings

// Organization Members
GET    /api/v1/organizations/:id/members  → List members
POST   /api/v1/organizations/:id/members  → Add member
POST   /api/v1/organizations/:id/invite   → Invite member by email
PUT    /api/v1/organizations/:id/members/:userId → Update member role
DELETE /api/v1/organizations/:id/members/:userId → Remove member

// Organization Departments
GET    /api/v1/organizations/:id/departments → List departments
POST   /api/v1/organizations/:id/departments → Create department
PUT    /api/v1/organizations/:id/departments/:deptId → Update
DELETE /api/v1/organizations/:id/departments/:deptId → Delete

// Workspace Switching
POST   /api/v1/organizations/:id/switch   → Set as current organization
```

#### Example: Create Organization Handler
```go
func CreateOrganization(c *fiber.Ctx) error {
    userID := c.Locals("userID").(string)

    var req struct {
        Name        string `json:"name" validate:"required,min=1,max=255"`
        Slug        string `json:"slug" validate:"required,min=1,max=100"`
        Description string `json:"description"`
    }

    if err := c.BindJSON(&req); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid request",
        })
    }

    orgService := services.NewOrganizationService(db.DB)
    org, err := orgService.CreateOrganization(c.Context(), req, userID)

    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": err.Error(),
        })
    }

    return c.Status(fiber.StatusCreated).JSON(org)
}
```

---

## Part 4: Frontend Implementation Strategy

### 4.1 Organization Context & State Management

#### Organization Context Provider
```typescript
// hooks/use-organization-context.ts

interface Organization {
  id: string;
  name: string;
  slug: string;
  logo_url?: string;
  tier: 'free' | 'pro' | 'enterprise';
}

interface OrganizationContextType {
  currentOrganization: Organization | null;
  userOrganizations: Organization[];
  setCurrentOrganization: (orgId: string) => Promise<void>;
  switchWorkspace: (orgId: string) => Promise<void>;
  isLoading: boolean;
  error: string | null;
}

export const useOrganizationContext = (): OrganizationContextType => {
  return useContext(OrganizationContext);
};
```

#### Workspace Switcher Component
```typescript
// components/workspace-switcher.tsx

export function WorkspaceSwitcher() {
  const {
    currentOrganization,
    userOrganizations,
    switchWorkspace
  } = useOrganizationContext();

  return (
    <Popover>
      <PopoverTrigger asChild>
        <Button variant="outline" className="w-full justify-between">
          <span className="flex items-center gap-2">
            <span className="text-2xl">
              {currentOrganization?.logo_url ? (
                <img
                  src={currentOrganization.logo_url}
                  alt={currentOrganization.name}
                  className="w-6 h-6 rounded"
                />
              ) : (
                <span className="bg-muted rounded p-1">
                  {currentOrganization?.name?.[0]?.toUpperCase()}
                </span>
              )}
            </span>
            <span>{currentOrganization?.name}</span>
          </span>
          <ChevronDown className="h-4 w-4" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-56">
        <div className="space-y-2">
          <p className="text-xs font-semibold text-muted-foreground mb-2">
            WORKSPACES
          </p>
          {userOrganizations.map((org) => (
            <button
              key={org.id}
              onClick={() => switchWorkspace(org.id)}
              className={cn(
                "w-full text-left px-2 py-2 rounded flex items-center gap-2",
                currentOrganization?.id === org.id
                  ? "bg-primary text-primary-foreground"
                  : "hover:bg-muted"
              )}
            >
              {org.logo_url ? (
                <img
                  src={org.logo_url}
                  alt={org.name}
                  className="w-4 h-4 rounded"
                />
              ) : (
                <span className="w-4 h-4 rounded bg-muted-foreground text-white text-xs flex items-center justify-center">
                  {org.name[0].toUpperCase()}
                </span>
              )}
              <span className="flex-1">{org.name}</span>
              {currentOrganization?.id === org.id && (
                <Check className="h-4 w-4" />
              )}
            </button>
          ))}
          <Separator className="my-2" />
          <button className="w-full text-left px-2 py-2 rounded hover:bg-muted text-sm">
            Create new workspace
          </button>
          <button className="w-full text-left px-2 py-2 rounded hover:bg-muted text-sm">
            Browse other workspaces
          </button>
        </div>
      </PopoverContent>
    </Popover>
  );
}
```

### 4.2 Updated API Hooks

#### React Query Hooks with Organization Context
```typescript
// hooks/api/use-requisitions.ts

export function useRequisitionsQuery(organizationId?: string) {
  const { currentOrganization } = useOrganizationContext();
  const orgId = organizationId || currentOrganization?.id;

  return useQuery({
    queryKey: ['requisitions', orgId],
    queryFn: async () => {
      const response = await fetch(`/api/v1/requisitions`, {
        headers: {
          'X-Organization-ID': orgId!,
          'Authorization': `Bearer ${getToken()}`,
        },
      });
      if (!response.ok) throw new Error('Failed to fetch');
      return response.json();
    },
    enabled: !!orgId,
  });
}

export function useCreateRequisitionMutation(organizationId?: string) {
  const { currentOrganization } = useOrganizationContext();
  const orgId = organizationId || currentOrganization?.id;
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: CreateRequisitionRequest) => {
      const response = await fetch(`/api/v1/requisitions`, {
        method: 'POST',
        headers: {
          'X-Organization-ID': orgId!,
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${getToken()}`,
        },
        body: JSON.stringify(data),
      });
      if (!response.ok) throw new Error('Failed to create');
      return response.json();
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['requisitions', orgId] });
    },
  });
}
```

### 4.3 Layout Updates

#### Updated App Layout with Workspace Switcher
```typescript
// app/(private)/layout.tsx

export default function PrivateLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <OrganizationProvider>
      <div className="flex h-screen">
        {/* Sidebar */}
        <aside className="w-64 border-r bg-muted/40 p-4">
          <div className="mb-8">
            <WorkspaceSwitcher />
          </div>

          <nav className="space-y-2">
            <SidebarLink href="/requisitions" icon={FileText}>
              Requisitions
            </SidebarLink>
            <SidebarLink href="/purchase-orders" icon={ShoppingCart}>
              Purchase Orders
            </SidebarLink>
            <SidebarLink href="/payment-vouchers" icon={CreditCard}>
              Payment Vouchers
            </SidebarLink>
            <SidebarLink href="/grn" icon={Package}>
              Goods Received
            </SidebarLink>
            <SidebarLink href="/analytics" icon={BarChart3}>
              Analytics
            </SidebarLink>
            <SidebarLink href="/settings" icon={Settings}>
              Settings
            </SidebarLink>
          </nav>
        </aside>

        {/* Main Content */}
        <main className="flex-1 overflow-auto">
          <Header />
          <div className="p-6">
            {children}
          </div>
        </main>
      </div>
    </OrganizationProvider>
  );
}
```

---

## Part 5: Implementation Roadmap

### Phase 13: Multi-Tenancy Implementation (10-12 weeks)

#### Sprint 1-2: Database & Schema (Weeks 1-2)
**Duration**: 60 hours
**Team**: 1 DB Engineer + 1 Backend Dev

- ✅ Create new organizations table
- ✅ Create organization_members table
- ✅ Create organization_settings table
- ✅ Create organization_departments table
- ✅ Enhance audit_logs with organization context
- ✅ Add organization_id to all business tables
- ✅ Create migration scripts
- ✅ Update GORM models
- ✅ Add foreign key constraints
- ✅ Create indexes for query performance

**Deliverables**:
- Migration scripts for dev/staging/prod
- Updated GORM models
- Database documentation

#### Sprint 3-4: Backend API & Services (Weeks 3-4)
**Duration**: 100 hours
**Team**: 2 Backend Developers

**Week 3: Core Organization Services**
- ✅ OrganizationService: Create, Read, Update, Delete
- ✅ OrganizationMemberService: Add/remove members, invitations
- ✅ AuthService: Organization context extraction
- ✅ TenantMiddleware: Enforce org scoping

**Week 4: Handler Refactoring**
- ✅ Update all existing handlers for org context
- ✅ Organization management endpoints
- ✅ Member management endpoints
- ✅ Department management endpoints
- ✅ Settings management endpoints

**Deliverables**:
- 15+ new API endpoints
- TenantContext extraction
- Automatic query scoping

#### Sprint 5: Authentication & Authorization (Week 5)
**Duration**: 50 hours
**Team**: 1-2 Backend Developers

- ✅ Update JWT payload with organization context
- ✅ Implement per-organization role management
- ✅ Update permission checking system
- ✅ Add organization-specific audit logging
- ✅ User-organization invitation workflow
- ✅ Workspace switching logic

#### Sprint 6-7: Frontend Refactoring (Weeks 6-7)
**Duration**: 80 hours
**Team**: 1-2 Frontend Developers

- ✅ OrganizationContext & Provider
- ✅ WorkspaceSwitcher component
- ✅ Update all API hooks with org context
- ✅ Add X-Organization-ID header to requests
- ✅ Layout updates with workspace switcher
- ✅ Settings page for organization management
- ✅ Member management UI
- ✅ Department management UI

#### Sprint 8: Testing & QA (Week 8)
**Duration**: 60 hours
**Team**: 1 QA Engineer + Developers

- ✅ Unit tests for org services
- ✅ Integration tests for API endpoints
- ✅ Multi-org data isolation tests
- ✅ User-org relationship tests
- ✅ Audit logging tests
- ✅ Cross-org access prevention tests

#### Sprint 9-10: Data Migration & Validation (Weeks 9-10)
**Duration**: 100 hours
**Team**: 1 Backend Dev + 1 QA

- ✅ Data migration strategy
- ✅ Create default organization for existing data
- ✅ Migrate all existing documents to org
- ✅ Create seed organizations
- ✅ Validate data integrity
- ✅ Rollback procedures

#### Sprint 11: Staging & UAT (Week 11)
**Duration**: 40 hours
**Team**: 1 QA + Product

- ✅ Deploy to staging
- ✅ User acceptance testing
- ✅ Performance testing with org scoping
- ✅ Security audit

#### Sprint 12: Production Rollout (Week 12)
**Duration**: 30 hours
**Team**: DevOps + Backend Leads

- ✅ Final production deployment
- ✅ Monitor for issues
- ✅ User training/documentation
- ✅ Rollback plan ready

### Estimated Effort Summary
| Component | Hours | Duration |
|-----------|-------|----------|
| Database Schema | 60 | 2 weeks |
| Backend Services | 100 | 2 weeks |
| Authentication/Auth | 50 | 1 week |
| Frontend Refactoring | 80 | 2 weeks |
| Testing & QA | 60 | 1 week |
| Data Migration | 100 | 2 weeks |
| Staging & UAT | 40 | 1 week |
| Rollout & Monitoring | 30 | 1 week |
| **TOTAL** | **520** | **12 weeks** |

**Team Size**: 4-5 people (2 backend, 1-2 frontend, 1 QA, 1 DevOps)
**Timeline**: 12 weeks (3 months) for full implementation

---

## Part 6: Data Migration Strategy

### 6.1 Migration Approach

#### Step 1: Create Default Organization
```sql
-- For all existing data, create a default organization
INSERT INTO organizations (name, slug, created_by, active)
VALUES ('Legacy System', 'legacy-system', (SELECT id FROM users LIMIT 1), true);

-- Store the org ID for use in next steps
SET @legacy_org_id = LAST_INSERT_ID();
```

#### Step 2: Migrate Existing Data
```sql
-- Add organization_id to all business documents
UPDATE requisitions SET organization_id = @legacy_org_id WHERE organization_id IS NULL;
UPDATE purchase_orders SET organization_id = @legacy_org_id WHERE organization_id IS NULL;
UPDATE payment_vouchers SET organization_id = @legacy_org_id WHERE organization_id IS NULL;
-- ... etc for all tables

-- Add all existing users as members of legacy org
INSERT INTO organization_members (organization_id, user_id, role, joined_at)
SELECT @legacy_org_id, id, role, NOW() FROM users;
```

#### Step 3: Verify Migration
```sql
-- Count documents per organization
SELECT organization_id, COUNT(*) FROM requisitions GROUP BY organization_id;

-- Verify all users in organization
SELECT COUNT(*) FROM organization_members WHERE organization_id = @legacy_org_id;

-- Verify no orphaned records
SELECT COUNT(*) FROM requisitions WHERE organization_id IS NULL;
```

### 6.2 Rollback Plan

If issues found, rollback approach:

```sql
-- Remove migration
ALTER TABLE requisitions DROP COLUMN organization_id;
DELETE FROM organization_members;
DELETE FROM organizations WHERE slug = 'legacy-system';
```

---

## Part 7: Risk Assessment & Mitigation

### Critical Risks

| Risk | Impact | Probability | Mitigation |
|------|--------|------------|-----------|
| Data corruption during migration | Critical | Low | Test on staging, full backup, dry-run |
| Query performance degradation | High | Medium | Add proper indexes, query optimization |
| Existing user experience break | High | Medium | Feature flags, phased rollout |
| Cross-org data leakage | Critical | Low | Mandatory org filtering on all queries |
| Token payload too large | Medium | Low | Use org ID shorthand in JWT |

### Mitigation Strategies

1. **Comprehensive Testing**
   - Unit tests for all org services
   - Integration tests for cross-org isolation
   - Stress testing with large datasets
   - Security penetration testing

2. **Phased Rollout**
   - Staging environment rollout first
   - Internal user testing
   - Beta group testing
   - Full production rollout with monitoring

3. **Database Safety**
   - Full backup before migration
   - Dry-run migration on copy
   - Rollback scripts prepared
   - Point-in-time recovery plan

4. **Query Optimization**
   - Profile queries before/after
   - Add necessary indexes
   - Cache frequently accessed orgs
   - Use query hints where needed

---

## Part 8: Success Criteria

### Functional Requirements
- ✅ Users can create new organizations
- ✅ Users can be invited to organizations
- ✅ Users can belong to multiple organizations
- ✅ Users can switch between organizations
- ✅ All data properly scoped by organization
- ✅ Organization settings configurable
- ✅ Department management functional
- ✅ Complete audit trail per organization

### Performance Requirements
- ✅ Query response time < 200ms with org scoping
- ✅ No performance degradation vs single-tenant
- ✅ Full database backup/restore < 5 min

### Security Requirements
- ✅ Zero cross-org data leakage
- ✅ Org context mandatory on all requests
- ✅ Audit logs complete and tamper-proof
- ✅ Authentication properly isolated

### User Experience
- ✅ Workspace switcher intuitive
- ✅ No friction in switching orgs
- ✅ All features work same way in each org
- ✅ Settings persist per org

---

## Part 9: Resource & Cost Estimation

### Team Composition (Ideal)
- 1 Principal Architect (oversight)
- 2 Backend Developers
- 1 Frontend Developer
- 1 Database Engineer
- 1 QA Engineer
- 1 DevOps Engineer
**Total**: 7 people for 12 weeks

### Cost Estimation (at $80/hour blended rate)
- **Development**: 400 hours × $80 = $32,000
- **QA/Testing**: 80 hours × $75 = $6,000
- **DevOps/Infrastructure**: 50 hours × $100 = $5,000
- **Total Labor**: **$43,000**

### Infrastructure Costs
- Staging environment: $500/month
- Database migration tools: $0-1,000
- Monitoring tools: $500/month
- **Total**: ~$2,000-3,000 for 3-month project

### **Total Project Cost**: $45,000-46,000

---

## Part 10: Dependencies & Prerequisites

### Before Starting
1. ✅ Phase 12 must be complete (PostgreSQL backend)
2. ✅ Current test coverage > 80%
3. ✅ All handlers refactored to use service layer
4. ✅ API documentation complete
5. ✅ DevOps pipeline ready for staging deployments

### External Dependencies
1. Email service for invitations (SendGrid, AWS SES)
2. Slug generation library (github.com/gosimple/slug)
3. UUID library (already in use)
4. Additional monitoring for multi-tenant isolation

---

## Part 11: Documentation Needs

### During Implementation
1. **Architecture Design Document**
   - Org model details
   - Query scoping patterns
   - Authorization model
   - Audit trail design

2. **API Documentation**
   - Organization endpoints
   - Member management
   - Auth with org context
   - Error handling

3. **Developer Guide**
   - How to add org scoping to new features
   - Query patterns with organizations
   - Testing multi-org scenarios
   - Migration considerations

4. **Operations Guide**
   - Monitoring org isolation
   - Debugging org-related issues
   - Scaling for multiple orgs
   - Backup/restore procedures

### User Documentation
1. **Admin Guide**
   - Creating organizations
   - Managing members
   - Configuring settings
   - Managing departments

2. **User Guide**
   - Switching workspaces
   - Understanding roles
   - Activity/audit trail
   - Settings per organization

---

## Part 12: Post-Implementation

### Phase 14+ Enhancements
1. **Organization Hierarchy**
   - Parent-child organizations
   - Shared resources between orgs
   - Consolidated reporting

2. **Advanced Features**
   - Organization-level workflows
   - Custom approvals per org
   - Org-specific integrations
   - Branding per organization

3. **Enterprise Features**
   - SSO per organization
   - SAML integration
   - Custom domain per org
   - White-label support

4. **Analytics**
   - Organization billing metrics
   - Usage analytics per org
   - User activity heatmaps
   - Cost per document

---

## Conclusion

This multi-tenancy refactor will transform Liyali Gateway from a single-tenant system into a **SaaS-ready platform** capable of supporting multiple organizations while maintaining complete data isolation and audit trails.

**Key Achievements**:
- ✅ Full organizational segregation
- ✅ User-organization flexibility (Slack-like)
- ✅ Complete audit trails per org
- ✅ Workspace switching (excellent UX)
- ✅ Foundation for SaaS business model

**Timeline**: 12 weeks with 4-5 person team
**Effort**: ~520 hours of development
**Cost**: ~$45,000-46,000
**Risk**: Medium (well-scoped, proven patterns)

The architecture is designed for **scalability from 1 to 1,000+ organizations** without significant refactoring.

---

**Document Status**: ✅ COMPLETE
**Last Updated**: 2025-12-15
**Next Review**: Phase 12 completion
**Related Documents**: 12-MISSING-FEATURES-GAP-ANALYSIS.md, 09-FUTURE-ENHANCEMENTS.md

