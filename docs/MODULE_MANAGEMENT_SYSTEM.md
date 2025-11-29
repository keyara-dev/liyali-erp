# Module Management System
## Dynamic Feature Access Control

**Date**: 2024-11-29
**Status**: Design Specification
**Purpose**: Define and manage modules that control navigation and feature access
**Key Requirement**: "Modules will need to be assigned to users for them to see nav items and page contents"

---

## Overview

A **Module** is a logical grouping of related features and pages. Users can only access modules they've been assigned. This adds another layer of control beyond roles and permissions:

```
Role → Determines what you can DO (approve, reject, etc.)
Permission → Granular actions (view_draft, add_comments)
Module → Determines what you can SEE and ACCESS
```

### Three-Layer Access Control

```
Layer 1: Authentication
  ↓ (User logged in?)
Layer 2: Role-Based Access Control (RBAC)
  ↓ (Do they have the right role?)
Layer 3: Module-Based Access Control (MBAC)
  ↓ (Do they have access to this module?)
Layer 4: Permission-Based Access Control (PBAC)
  ↓ (Do they have the specific permission?)
Feature/Action Available
```

---

## Module Definition

### What is a Module?

A module is a collection of related pages and features:

```typescript
export type Module = {
  id: string
  name: string
  description: string
  icon: React.ReactNode
  permissions: string[] // Minimum permissions needed
  requiredRoles?: string[] // Optional: specific roles that need this
  pages: string[] // Pages included in this module
  features: string[] // Features included in this module
  order: number // Display order
  isCore: boolean // Core modules can't be disabled
  status: 'ACTIVE' | 'INACTIVE' | 'BETA'
  createdAt: Date
  createdBy: string
}

export type UserModuleAssignment = {
  userId: string
  moduleId: string
  assignedAt: Date
  assignedBy: string
  expiresAt?: Date // Optional: temporary access
  status: 'ACTIVE' | 'SUSPENDED' | 'EXPIRED'
}
```

---

## Core Modules Definition

### Module 1: Dashboard (Core)
```typescript
{
  id: 'dashboard',
  name: 'Dashboard',
  description: 'Main dashboard with overview and metrics',
  icon: LayoutDashboard,
  permissions: ['view_dashboard'],
  pages: ['/dashboard', '/dashboard/*'],
  features: [
    'view-pending-approvals',
    'view-statistics',
    'view-notifications'
  ],
  isCore: true,
  status: 'ACTIVE'
}
```

**Who Gets It**: Everyone
**Can See**: Dashboard link, Dashboard content
**Can't See Without**: Nothing (everyone has dashboard)

---

### Module 2: Requisitions Management
```typescript
{
  id: 'requisitions',
  name: 'Requisitions Management',
  description: 'Create, manage, and approve purchase requisitions',
  icon: FileText,
  permissions: ['view_requisitions'],
  requiredRoles: ['REQUESTER', 'DEPARTMENT_MANAGER', 'PRINCIPAL_OFFICER', 'DIRECTOR_FINANCE', 'ADMIN'],
  pages: [
    '/workflows/requisitions',
    '/workflows/requisitions/new',
    '/workflows/requisitions/[id]'
  ],
  features: [
    'create-requisition',
    'view-requisitions',
    'approve-requisition',
    'reject-requisition',
    'reverse-requisition',
    'add-requisition-comments',
    'upload-attachments'
  ],
  order: 1,
  isCore: false,
  status: 'ACTIVE'
}
```

**Who Gets It**: Requisition users
**Can See**: Requisitions nav item, Requisitions list/detail pages
**Can't See Without**: Item and pages hidden

---

### Module 3: Purchase Orders
```typescript
{
  id: 'purchase-orders',
  name: 'Purchase Orders',
  description: 'Manage purchase orders through approval workflow',
  icon: ShoppingCart,
  permissions: ['view_purchase_orders'],
  requiredRoles: ['DEPARTMENT_MANAGER', 'AUDITOR', 'DIRECTOR_FINANCE', 'PRINCIPAL_OFFICER', 'ADMIN'],
  pages: [
    '/workflows/purchase-orders',
    '/workflows/purchase-orders/[id]'
  ],
  features: [
    'view-purchase-orders',
    'approve-purchase-order',
    'reverse-purchase-order',
    'view-po-approvals'
  ],
  order: 2,
  isCore: false,
  status: 'ACTIVE'
}
```

**Who Gets It**: Approval chain users
**Can See**: PO nav item, PO pages
**Stages Visible**: Only their approval stage

---

### Module 4: Goods Received Note
```typescript
{
  id: 'goods-received-note',
  name: 'Goods Received Notes',
  description: 'Record goods receipt and create payment vouchers',
  icon: Package,
  permissions: ['view_grn'],
  requiredRoles: ['FINANCE_OFFICER', 'ADMIN'],
  pages: [
    '/workflows/grn',
    '/workflows/grn/new',
    '/workflows/grn/[id]'
  ],
  features: [
    'create-grn',
    'view-grn',
    'complete-grn',
    'record-discrepancies'
  ],
  order: 3,
  isCore: false,
  status: 'ACTIVE'
}
```

**Who Gets It**: Stores/Finance staff
**Can See**: GRN nav item and pages
**Can't See Without**: Hidden from other users

---

### Module 5: Payment Vouchers
```typescript
{
  id: 'payment-vouchers',
  name: 'Payment Vouchers',
  description: 'Create and approve payment vouchers for payment processing',
  icon: DollarSign,
  permissions: ['view_payment_vouchers'],
  requiredRoles: ['FINANCE_OFFICER', 'ACCOUNTANT', 'DEPARTMENT_MANAGER', 'AUDITOR', 'DIRECTOR_FINANCE', 'PRINCIPAL_OFFICER', 'ADMIN'],
  pages: [
    '/workflows/payment-vouchers',
    '/workflows/payment-vouchers/[id]'
  ],
  features: [
    'view-payment-vouchers',
    'create-payment-voucher',
    'approve-payment-voucher',
    'reverse-payment-voucher',
    'generate-qr-code',
    'validate-bank-info'
  ],
  order: 4,
  isCore: false,
  status: 'ACTIVE'
}
```

**Who Gets It**: Finance team and approvers
**Can See**: PV nav item and pages
**Features**: Create (Accountant), Approve (all roles), QR (PO only)

---

### Module 6: Transaction Search & Verification
```typescript
{
  id: 'transactions',
  name: 'Transactions',
  description: 'Search transactions, verify QR codes, and download reports',
  icon: Search,
  permissions: ['search_transactions'],
  requiredRoles: ['FINANCE_OFFICER', 'DIRECTOR_FINANCE', 'PRINCIPAL_OFFICER', 'AUDITOR', 'ADMIN'],
  pages: [
    '/transactions',
    '/transactions/[id]'
  ],
  features: [
    'search-transactions',
    'filter-by-date',
    'filter-by-reference',
    'filter-by-vendor',
    'verify-qr-code',
    'download-pdf',
    'view-transaction-details'
  ],
  order: 5,
  isCore: false,
  status: 'ACTIVE'
}
```

**Who Gets It**: Finance and compliance users
**Can See**: Transactions nav item and search page
**Can Do**: Search, verify, download

---

### Module 7: Reporting & Analytics
```typescript
{
  id: 'reporting',
  name: 'Reporting & Analytics',
  description: 'View financial reports, dashboards, and analytics',
  icon: BarChart3,
  permissions: ['view_reports'],
  requiredRoles: ['DIRECTOR_FINANCE', 'PRINCIPAL_OFFICER', 'AUDITOR', 'ADMIN'],
  pages: [
    '/dashboard/reports',
    '/dashboard/analytics',
    '/dashboard/budget-analysis'
  ],
  features: [
    'view-transaction-volume',
    'view-approval-metrics',
    'view-budget-analysis',
    'generate-reports',
    'export-data'
  ],
  order: 6,
  isCore: false,
  status: 'ACTIVE'
}
```

**Who Gets It**: Management and finance leadership
**Can See**: Reports section in dashboard
**Can Do**: View/generate reports

---

### Module 8: User Management (Admin Only)
```typescript
{
  id: 'user-management',
  name: 'User Management',
  description: 'Manage users, roles, permissions, and module assignments',
  icon: Users,
  permissions: ['manage_users', 'manage_roles', 'manage_modules'],
  requiredRoles: ['ADMIN'],
  pages: [
    '/admin/users',
    '/admin/users/[id]',
    '/admin/roles',
    '/admin/modules',
    '/admin/permissions',
    '/admin/audit-logs',
    '/admin/access-logs'
  ],
  features: [
    'create-user',
    'edit-user',
    'delete-user',
    'create-role',
    'edit-role',
    'delete-role',
    'assign-role-to-user',
    'assign-permission-to-role',
    'assign-module-to-user',
    'view-audit-logs',
    'view-access-logs'
  ],
  order: 7,
  isCore: true,
  status: 'ACTIVE'
}
```

**Who Gets It**: Administrators only
**Can See**: Admin nav section
**Can Do**: Create/edit users, roles, modules, permissions

---

### Module 9: Workflow Configuration (Admin Only)
```typescript
{
  id: 'workflow-config',
  name: 'Workflow Configuration',
  description: 'Configure approval workflows and stages',
  icon: Settings,
  permissions: ['manage_workflows'],
  requiredRoles: ['ADMIN'],
  pages: [
    '/admin/workflows',
    '/admin/workflows/[type]'
  ],
  features: [
    'view-workflow-config',
    'edit-workflow-stages',
    'configure-reversals',
    'configure-validations',
    'configure-actions',
    'set-sla-times'
  ],
  order: 8,
  isCore: true,
  status: 'ACTIVE'
}
```

**Who Gets It**: Administrators only
**Can See**: Workflow Config admin page
**Can Do**: Modify approval workflows

---

## Module Assignment Workflow

### Step 1: Create User
```
1. Admin creates new user
   - Name
   - Email
   - Password
   - Department
   - Assign roles (DEPARTMENT_MANAGER, AUDITOR, etc.)
```

### Step 2: Assign Role
```
2. Admin assigns role to user
   - User now has role's permissions
   - Role determines what actions user can perform
   - Example: DEPARTMENT_MANAGER can approve at stage 1
```

### Step 3: Assign Modules
```
3. Admin assigns modules to user
   - User can now access module pages
   - User can see module nav items
   - User can use module features
   - Example: Assign "Requisitions Management" module
```

**Result**: User can see and access Requisitions pages

### Step 4: Verify Access
```
4. User logs in
   - Sees only nav items for assigned modules
   - Can only access pages in assigned modules
   - Can only perform actions in assigned modules
   - Cannot see other modules' content
```

---

## Admin Panel: Module Management Tab

### User Management Section

```
/admin/users

[Users List]
├── User Name | Email | Roles | Modules | Status | Actions

[Actions]
├── Create User
├── Edit User
│   ├── Basic Info (name, email)
│   ├── Roles Assignment (checkboxes)
│   ├── Module Assignment (checkboxes)
│   └── Status (Active/Inactive/Suspended)
└── Delete User
```

### User Edit Page Structure

```
/admin/users/[user-id]

┌─────────────────────────────────────┐
│ USER DETAILS                        │
├─────────────────────────────────────┤
│ Name: [___________]                 │
│ Email: [________________]           │
│ Department: [__________]            │
│ Status: [Active ▼]                  │
│ Created: 2024-11-29                 │
│ Last Login: 2024-11-29 14:30        │
└─────────────────────────────────────┘

┌─────────────────────────────────────┐
│ ROLE ASSIGNMENTS                    │
├─────────────────────────────────────┤
│ ☑ REQUESTER                         │
│ ☑ DEPARTMENT_MANAGER                │
│ ☐ AUDITOR                           │
│ ☑ FINANCE_OFFICER                   │
│ ☐ ACCOUNTANT                        │
│ ☐ DIRECTOR_FINANCE                  │
│ ☐ PRINCIPAL_OFFICER                 │
│ ☐ ADMIN                             │
└─────────────────────────────────────┘

┌─────────────────────────────────────┐
│ MODULE ASSIGNMENTS                  │
├─────────────────────────────────────┤
│ ☑ Dashboard (core)                  │
│ ☑ Requisitions Management           │
│ ☐ Purchase Orders                   │
│ ☑ Payment Vouchers                  │
│ ☐ Goods Received Notes              │
│ ☐ Transactions                      │
│ ☐ Reporting & Analytics             │
│ ☐ User Management (admin)           │
│ ☐ Workflow Configuration (admin)    │
│                                     │
│ [Save] [Cancel]                     │
└─────────────────────────────────────┘
```

### Module Management Tab

```
/admin/modules

[Modules List]
├── Module Name | Description | Status | Users | Actions

[Module Details]
├── Name: [________________]
├── Description: [_______________________]
├── Status: [Active ▼]
├── Pages Included:
│   └── ☑ /workflows/requisitions
│       ☑ /workflows/requisitions/new
│       ☑ /workflows/requisitions/[id]
│       ...
├── Features Included:
│   └── ☑ create-requisition
│       ☑ view-requisitions
│       ☑ approve-requisition
│       ...
├── Permissions Required:
│   └── view_requisitions
├── Users Assigned: (count)
│   └── [John Doe]
│       [Jane Smith]
│       [See all (5)]
│
└── [Save] [Delete] [Assign to Users]
```

### Role Management Tab

```
/admin/roles

[Roles List]
├── Role Name | Description | Users | Permissions | Actions

[Role Details]
├── Name: [DEPARTMENT_MANAGER]
├── Description: [Manages department requisitions]
├── Users with this role: (count)
├── Permissions:
│   └── ☑ view_draft
│       ☑ edit_draft
│       ☑ submit_document
│       ☑ approve_document
│       ☑ view_attachments
│       ☑ add_comments
│       ...
└── [Save] [Edit Permissions]
```

### Permission Assignment Tab

```
/admin/permissions

[Permissions List]
├── Permission | Description | Roles Using It | Status

[Permission Details]
├── ID: view_requisitions
├── Name: View Requisitions
├── Description: View requisition documents
├── Roles Using:
│   └── ☑ REQUESTER (assigned)
│       ☑ DEPARTMENT_MANAGER (assigned)
│       ☐ AUDITOR (not assigned)
│
├── [Assign to Roles] [Create New Permission]
└── [Save]
```

---

## Implementation: TypeScript Interfaces

```typescript
// src/types/modules.ts

export type Module = {
  id: string
  name: string
  description: string
  icon?: string
  permissions: string[]
  requiredRoles?: string[]
  pages: string[]
  features: string[]
  order: number
  isCore: boolean
  status: 'ACTIVE' | 'INACTIVE' | 'BETA'
  createdAt: Date
  createdBy: string
  updatedAt?: Date
  updatedBy?: string
}

export type UserModuleAssignment = {
  id: string
  userId: string
  moduleId: string
  assignedAt: Date
  assignedBy: string
  expiresAt?: Date
  status: 'ACTIVE' | 'SUSPENDED' | 'EXPIRED'
}

export type UserWithModules = User & {
  modules: Module[]
  moduleIds: string[]
}
```

---

## Implementation: Configuration

```typescript
// src/lib/modules-config.ts

export const modulesConfig: Module[] = [
  {
    id: 'dashboard',
    name: 'Dashboard',
    description: 'Main dashboard with overview and metrics',
    permissions: ['view_dashboard'],
    pages: ['/dashboard'],
    features: ['view-pending-approvals', 'view-statistics'],
    order: 0,
    isCore: true,
    status: 'ACTIVE',
    createdAt: new Date(),
    createdBy: 'system'
  },
  // ... other modules
]

// Module registry
const moduleRegistry = new Map<string, Module>()
modulesConfig.forEach(module => {
  moduleRegistry.set(module.id, module)
})

// Get module by ID
export function getModule(moduleId: string): Module | null {
  return moduleRegistry.get(moduleId) || null
}

// Get all modules
export function getAllModules(): Module[] {
  return Array.from(moduleRegistry.values())
}

// Check if user has module
export function userHasModule(
  userModules: string[],
  moduleId: string
): boolean {
  return userModules.includes(moduleId)
}

// Get user's accessible modules
export function getUserModules(
  userModules: string[]
): Module[] {
  return userModules
    .map(moduleId => getModule(moduleId))
    .filter((m): m is Module => m !== null)
}

// Get user's accessible pages
export function getUserAccessiblePages(
  userModules: string[]
): string[] {
  return getUserModules(userModules)
    .flatMap(m => m.pages)
}

// Check if user can access page
export function canAccessPage(
  page: string,
  userModules: string[]
): boolean {
  const accessiblePages = getUserAccessiblePages(userModules)
  return accessiblePages.some(p => {
    // Exact match or wildcard match
    if (p === page) return true
    if (p.endsWith('/*')) {
      const basePath = p.replace('/*', '')
      return page.startsWith(basePath)
    }
    return false
  })
}
```

---

## Implementation: Server Actions

```typescript
// src/app/_actions/modules.ts

'use server'

import { Module, UserModuleAssignment } from '@/types/modules'
import { store } from '@/lib/mock-data'

// Get user's modules
export async function getUserModules(userId: string): Promise<Module[]> {
  const assignments = store.userModuleAssignments.values()
    .filter(a => a.userId === userId && a.status === 'ACTIVE')

  const modules: Module[] = []
  for (const assignment of assignments) {
    const module = store.modules.get(assignment.moduleId)
    if (module) {
      modules.push(module)
    }
  }

  return modules
}

// Assign module to user
export async function assignModuleToUser(
  userId: string,
  moduleId: string,
  adminUserId: string
): Promise<{ success: boolean; error?: string }> {
  // Verify admin
  const admin = store.users.get(adminUserId)
  if (!admin?.roleIds.includes('ADMIN')) {
    return { success: false, error: 'Unauthorized' }
  }

  // Create assignment
  const assignment: UserModuleAssignment = {
    id: `assignment-${Date.now()}`,
    userId,
    moduleId,
    assignedAt: new Date(),
    assignedBy: adminUserId,
    status: 'ACTIVE'
  }

  store.userModuleAssignments.set(assignment.id, assignment)

  return { success: true }
}

// Remove module from user
export async function removeModuleFromUser(
  userId: string,
  moduleId: string,
  adminUserId: string
): Promise<{ success: boolean; error?: string }> {
  // Verify admin
  const admin = store.users.get(adminUserId)
  if (!admin?.roleIds.includes('ADMIN')) {
    return { success: false, error: 'Unauthorized' }
  }

  // Find and remove assignment
  for (const [key, assignment] of store.userModuleAssignments) {
    if (assignment.userId === userId && assignment.moduleId === moduleId) {
      store.userModuleAssignments.delete(key)
      break
    }
  }

  return { success: true }
}

// List all modules
export async function listModules(): Promise<Module[]> {
  return Array.from(store.modules.values())
}

// Create module
export async function createModule(
  module: Module,
  adminUserId: string
): Promise<{ success: boolean; moduleId?: string; error?: string }> {
  // Verify admin
  const admin = store.users.get(adminUserId)
  if (!admin?.roleIds.includes('ADMIN')) {
    return { success: false, error: 'Unauthorized' }
  }

  const moduleId = `module-${Date.now()}`
  const newModule: Module = {
    ...module,
    id: moduleId,
    createdAt: new Date(),
    createdBy: adminUserId
  }

  store.modules.set(moduleId, newModule)

  return { success: true, moduleId }
}

// Update module
export async function updateModule(
  moduleId: string,
  updates: Partial<Module>,
  adminUserId: string
): Promise<{ success: boolean; error?: string }> {
  // Verify admin
  const admin = store.users.get(adminUserId)
  if (!admin?.roleIds.includes('ADMIN')) {
    return { success: false, error: 'Unauthorized' }
  }

  const module = store.modules.get(moduleId)
  if (!module) {
    return { success: false, error: 'Module not found' }
  }

  const updated: Module = {
    ...module,
    ...updates,
    id: moduleId, // Don't allow changing ID
    updatedAt: new Date(),
    updatedBy: adminUserId
  }

  store.modules.set(moduleId, updated)

  return { success: true }
}
```

---

## Implementation: Middleware

```typescript
// src/middleware.ts

import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'
import { getUserModules, canAccessPage } from '@/lib/modules'

export async function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl

  // Get user from session/token
  const user = request.user // From auth middleware

  // Allow public pages
  if (pathname === '/login' || pathname === '/') {
    return NextResponse.next()
  }

  // Get user's modules
  const userModules = await getUserModules(user.id)
  const moduleIds = userModules.map(m => m.id)

  // Check access
  if (!canAccessPage(pathname, moduleIds)) {
    // Redirect to dashboard or unauthorized page
    return NextResponse.redirect(new URL('/dashboard', request.url))
  }

  return NextResponse.next()
}

export const config = {
  matcher: [
    '/workflows/:path*',
    '/transactions/:path*',
    '/dashboard/:path*',
    '/admin/:path*'
  ]
}
```

---

## Navigation Integration

```typescript
// Updated src/lib/navigation-config.ts

export function getNavigationForUser(
  userRoles: string[],
  userModules: string[]
): NavigationItem[] {
  return navigationConfig.mainNav.filter(item => {
    // Check if item's required roles match
    const hasRole = item.requiredRoles.length === 0 ||
      item.requiredRoles.some(role => userRoles.includes(role))

    // Check if user has the module for this item
    const hasModule = !item.moduleId ||
      userModules.includes(item.moduleId)

    return hasRole && hasModule
  })
}

// Usage:
const visibleNav = getNavigationForUser(
  user.roleIds,
  user.moduleIds
)
```

---

## User Story Example

### Scenario: New Finance Officer Joining

**Before Module Assignment**:
- User created with FINANCE_OFFICER role
- Can see: Dashboard only
- Cannot see: Requisitions, POs, GRN, Payment Vouchers nav items
- Cannot access: Any workflow pages

**Step 1: Assign Roles**
- Admin assigns: FINANCE_OFFICER role
- User now has: Finance permissions
- Can do: View and manage financial data (in theory)
- But still cannot see pages

**Step 2: Assign Modules**
- Admin assigns: "Goods Received Notes" module
- User now sees: GRN nav item
- User can access: /workflows/grn pages
- User can create: GRNs

**Step 3: Assign Additional Modules**
- Admin assigns: "Payment Vouchers" module
- User now sees: PV nav item
- User can access: /workflows/payment-vouchers pages
- User can: Create/manage PVs

**Step 4: Grant Audit Access**
- Admin assigns: "Transactions" module
- User now sees: Transactions nav item
- User can: Search transactions, verify QR codes

**Result**: Finance officer has exactly the access needed, nothing more

---

## Benefits of Module System

✅ **Granular Control**: Assign specific modules to specific users
✅ **Navigation Clarity**: Users see only relevant menu items
✅ **Feature Access**: Control what users can do
✅ **Easy Onboarding**: Assign modules, not individual permissions
✅ **Flexible Organization**: Can group related features
✅ **Scalable**: Add new modules without affecting existing ones
✅ **Security**: Multiple layers of access control
✅ **Compliance**: Audit who has access to what
✅ **User Experience**: Cleaner interface for each role

---

## Implementation Timeline

### Phase 0 Expansion (Admin System)
- [ ] Create Module types in workflow.ts
- [ ] Create modules-config.ts with all modules
- [ ] Create module server actions
- [ ] Create module middleware

### Phase 1 (Integration)
- [ ] Update sidebar to use module-aware navigation
- [ ] Update navigation filtering
- [ ] Test module access control

### Phase 2A (Purchase Orders)
- [ ] Add PO module to modules-config
- [ ] Assign to appropriate users

### Phase 2B (Goods Received Note)
- [ ] Add GRN module to modules-config
- [ ] Assign to finance staff

### Phase 2C (Payment Vouchers)
- [ ] Add PV module to modules-config
- [ ] Assign to finance and approvers

### Phase 2D (Transactions)
- [ ] Add Transactions module to modules-config
- [ ] Assign to finance/auditors

### Admin Implementation
- [ ] Create /admin/modules page
- [ ] Create /admin/modules/[id] edit page
- [ ] Create module assignment UI in user edit
- [ ] Create module list in module management tab

---

## Database Schema (Conceptual)

```sql
-- Modules table
CREATE TABLE modules (
  id VARCHAR(255) PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  description TEXT,
  icon VARCHAR(255),
  permissions JSON,
  required_roles JSON,
  pages JSON,
  features JSON,
  order INT,
  is_core BOOLEAN DEFAULT false,
  status VARCHAR(50),
  created_at TIMESTAMP,
  created_by VARCHAR(255),
  updated_at TIMESTAMP,
  updated_by VARCHAR(255)
);

-- User module assignments
CREATE TABLE user_module_assignments (
  id VARCHAR(255) PRIMARY KEY,
  user_id VARCHAR(255) NOT NULL,
  module_id VARCHAR(255) NOT NULL,
  assigned_at TIMESTAMP,
  assigned_by VARCHAR(255),
  expires_at TIMESTAMP NULL,
  status VARCHAR(50),
  FOREIGN KEY (user_id) REFERENCES users(id),
  FOREIGN KEY (module_id) REFERENCES modules(id),
  UNIQUE KEY unique_user_module (user_id, module_id)
);

-- Index for fast lookups
CREATE INDEX idx_user_modules ON user_module_assignments(user_id);
CREATE INDEX idx_module_users ON user_module_assignments(module_id);
```

---

## Mock Data Structure

```typescript
// src/lib/mock-data.ts additions

// Store for modules
export const modules = new Map<string, Module>([
  ['dashboard', { id: 'dashboard', name: 'Dashboard', ... }],
  ['requisitions', { id: 'requisitions', name: 'Requisitions', ... }],
  ['purchase-orders', { id: 'purchase-orders', name: 'Purchase Orders', ... }],
  ['goods-received-note', { id: 'goods-received-note', name: 'GRN', ... }],
  ['payment-vouchers', { id: 'payment-vouchers', name: 'Payment Vouchers', ... }],
  ['transactions', { id: 'transactions', name: 'Transactions', ... }],
  ['reporting', { id: 'reporting', name: 'Reporting', ... }],
  ['user-management', { id: 'user-management', name: 'User Management', ... }],
  ['workflow-config', { id: 'workflow-config', name: 'Workflow Config', ... }]
])

// Store for user module assignments
export const userModuleAssignments = new Map<string, UserModuleAssignment>([
  // Example: John (requester) has dashboard and requisitions
  ['assign-1', {
    id: 'assign-1',
    userId: 'user-john',
    moduleId: 'dashboard',
    assignedAt: new Date(),
    assignedBy: 'system',
    status: 'ACTIVE'
  }],
  ['assign-2', {
    id: 'assign-2',
    userId: 'user-john',
    moduleId: 'requisitions',
    assignedAt: new Date(),
    assignedBy: 'system',
    status: 'ACTIVE'
  }],
  // Example: Finance officer has GRN and PV modules
  ['assign-3', {
    id: 'assign-3',
    userId: 'user-finance',
    moduleId: 'goods-received-note',
    assignedAt: new Date(),
    assignedBy: 'system',
    status: 'ACTIVE'
  }],
  // ... more assignments
])
```

---

## Summary

The Module Management System provides:

✅ **Three-Layer Access Control**: Authentication → Roles → Modules → Permissions
✅ **9 Core Modules**: Dashboard, Requisitions, PO, GRN, PV, Transactions, Reporting, User Mgmt, Workflow Config
✅ **Admin Panel**: Complete CRUD for users, roles, modules, and assignments
✅ **Dynamic Navigation**: Sidebar updates based on assigned modules
✅ **Flexible Assignment**: Temporary access, expiration dates, status management
✅ **Comprehensive Audit**: Know who has access to what
✅ **Security First**: Validated at every level

---

**Created**: 2024-11-29
**Status**: Ready for Implementation
**Next Step**: Add to Phase 0 Expansion (Admin System)
**Effort**: 2-3 hours for admin UI + integration
