# Role-Based Navigation & Page Structure Design

**Date**: 2024-11-29
**Status**: Design Specification
**Purpose**: Define side navigation and page structure with role-based visibility
**Key Requirement**: "Each user will see only what their role-permissions allow them to see on the platform"

---

## Overview

The navigation system must be **dynamic and role-aware**. Each user sees different menu items based on their assigned roles and permissions. This applies to:

1. **Side Navigation Items** - Which links appear in the sidebar
2. **Page Accessibility** - Which pages each role can access
3. **Feature Visibility** - Which features/buttons show on pages
4. **Dashboard Content** - What widgets appear on dashboard
5. **Action Items** - What actions (approve, reverse, etc.) are available

---

## System Architecture

### Navigation Data Structure

```typescript
// src/types/navigation.ts

export type NavigationItem = {
  id: string
  label: string
  icon: React.ReactNode
  href: string
  requiredRoles: string[] // Any of these roles can see it
  requiredPermissions?: string[] // AND these permissions
  children?: NavigationItem[]
  badge?: {
    label: string
    color: 'red' | 'yellow' | 'green' | 'blue'
  }
  divider?: boolean
}

export type NavigationConfig = {
  mainNav: NavigationItem[]
  userMenu: NavigationItem[]
  bottomNav?: NavigationItem[]
}
```

### Role-Based Visibility Logic

```typescript
// src/lib/nav-visibility.ts

export function canAccessNavItem(
  item: NavigationItem,
  userRoles: string[],
  userPermissions: string[]
): boolean {
  // User must have at least one of the required roles
  const hasRole = item.requiredRoles.length === 0 ||
    item.requiredRoles.some(role => userRoles.includes(role))

  // User must have ALL required permissions (if specified)
  const hasPermissions = !item.requiredPermissions ||
    item.requiredPermissions.every(perm => userPermissions.includes(perm))

  return hasRole && hasPermissions
}

export function filterNavItems(
  items: NavigationItem[],
  userRoles: string[],
  userPermissions: string[]
): NavigationItem[] {
  return items
    .filter(item => canAccessNavItem(item, userRoles, userPermissions))
    .map(item => ({
      ...item,
      children: item.children
        ? filterNavItems(item.children, userRoles, userPermissions)
        : undefined
    }))
    .filter(item => !item.children || item.children.length > 0)
}
```

---

## Complete Navigation Configuration

### All Available Roles

```typescript
const AVAILABLE_ROLES = {
  REQUESTER: 'Requester',
  DEPARTMENT_MANAGER: 'Department Manager',
  AUDITOR: 'Auditor/Compliance Officer',
  FINANCE_OFFICER: 'Finance Officer',
  DIRECTOR_FINANCE: 'Finance Director',
  PRINCIPAL_OFFICER: 'Principal Officer / Executive',
  ACCOUNTANT: 'Accountant',
  ADMIN: 'System Administrator'
}
```

### Navigation Definition

```typescript
// src/lib/navigation-config.ts

export const navigationConfig: NavigationConfig = {
  mainNav: [
    // ========================================================================
    // DASHBOARD SECTION - Everyone can see
    // ========================================================================
    {
      id: 'dashboard',
      label: 'Dashboard',
      icon: LayoutDashboard,
      href: '/dashboard',
      requiredRoles: [
        'REQUESTER',
        'DEPARTMENT_MANAGER',
        'AUDITOR',
        'FINANCE_OFFICER',
        'DIRECTOR_FINANCE',
        'PRINCIPAL_OFFICER',
        'ACCOUNTANT',
        'ADMIN'
      ],
    },

    // ========================================================================
    // REQUISITION SECTION - Requesters & Approvers
    // ========================================================================
    {
      id: 'requisitions-section',
      label: 'Requisitions',
      icon: FileText,
      href: '/workflows/requisitions',
      requiredRoles: [
        'REQUESTER',
        'DEPARTMENT_MANAGER',
        'PRINCIPAL_OFFICER',
        'DIRECTOR_FINANCE',
        'ADMIN'
      ],
      children: [
        {
          id: 'req-my-requests',
          label: 'My Requisitions',
          icon: FileText,
          href: '/workflows/requisitions?filter=created-by-me',
          requiredRoles: ['REQUESTER', 'ADMIN']
        },
        {
          id: 'req-to-approve',
          label: 'Pending Approvals',
          icon: CheckCircle2,
          href: '/workflows/requisitions?filter=pending-approval',
          requiredRoles: [
            'DEPARTMENT_MANAGER',
            'PRINCIPAL_OFFICER',
            'DIRECTOR_FINANCE',
            'ADMIN'
          ],
          badge: {
            label: 'Dynamic Count',
            color: 'red'
          }
        },
        {
          id: 'req-all',
          label: 'All Requisitions',
          icon: List,
          href: '/workflows/requisitions',
          requiredRoles: [
            'DEPARTMENT_MANAGER',
            'PRINCIPAL_OFFICER',
            'DIRECTOR_FINANCE',
            'ADMIN'
          ]
        }
      ]
    },

    // ========================================================================
    // PURCHASE ORDER SECTION - Approvers & Procurement
    // ========================================================================
    {
      id: 'po-section',
      label: 'Purchase Orders',
      icon: ShoppingCart,
      href: '/workflows/purchase-orders',
      requiredRoles: [
        'DEPARTMENT_MANAGER',
        'AUDITOR',
        'DIRECTOR_FINANCE',
        'PRINCIPAL_OFFICER',
        'ADMIN'
      ],
      children: [
        {
          id: 'po-to-approve',
          label: 'Pending Approvals',
          icon: CheckCircle2,
          href: '/workflows/purchase-orders?filter=pending-approval',
          requiredRoles: [
            'DEPARTMENT_MANAGER',
            'AUDITOR',
            'DIRECTOR_FINANCE',
            'PRINCIPAL_OFFICER',
            'ADMIN'
          ],
          badge: {
            label: 'Dynamic Count',
            color: 'red'
          }
        },
        {
          id: 'po-in-process',
          label: 'In Process',
          icon: Clock,
          href: '/workflows/purchase-orders?filter=in-approval',
          requiredRoles: [
            'DEPARTMENT_MANAGER',
            'AUDITOR',
            'DIRECTOR_FINANCE',
            'PRINCIPAL_OFFICER',
            'ADMIN'
          ]
        },
        {
          id: 'po-approved',
          label: 'Approved',
          icon: CheckCircle,
          href: '/workflows/purchase-orders?filter=approved',
          requiredRoles: [
            'DEPARTMENT_MANAGER',
            'DIRECTOR_FINANCE',
            'PRINCIPAL_OFFICER',
            'ADMIN'
          ]
        },
        {
          id: 'po-all',
          label: 'All POs',
          icon: List,
          href: '/workflows/purchase-orders',
          requiredRoles: [
            'DEPARTMENT_MANAGER',
            'AUDITOR',
            'DIRECTOR_FINANCE',
            'PRINCIPAL_OFFICER',
            'ADMIN'
          ]
        }
      ]
    },

    // ========================================================================
    // GOODS RECEIVED NOTE SECTION - Stores & Approvers
    // ========================================================================
    {
      id: 'grn-section',
      label: 'Goods Received',
      icon: Package,
      href: '/workflows/grn',
      requiredRoles: [
        'ADMIN',
        'FINANCE_OFFICER',
        'DIRECTOR_FINANCE',
        'ACCOUNTANT'
      ],
      children: [
        {
          id: 'grn-new',
          label: 'Create GRN',
          icon: Plus,
          href: '/workflows/grn/new',
          requiredRoles: ['ADMIN', 'FINANCE_OFFICER']
        },
        {
          id: 'grn-pending',
          label: 'Pending Receipt',
          icon: Clock,
          href: '/workflows/grn?filter=pending',
          requiredRoles: ['ADMIN', 'FINANCE_OFFICER']
        },
        {
          id: 'grn-completed',
          label: 'Completed',
          icon: CheckCircle,
          href: '/workflows/grn?filter=completed',
          requiredRoles: ['ADMIN', 'FINANCE_OFFICER', 'DIRECTOR_FINANCE']
        }
      ]
    },

    // ========================================================================
    // PAYMENT VOUCHER SECTION - Finance & Approvers
    // ========================================================================
    {
      id: 'pv-section',
      label: 'Payment Vouchers',
      icon: DollarSign,
      href: '/workflows/payment-vouchers',
      requiredRoles: [
        'FINANCE_OFFICER',
        'DEPARTMENT_MANAGER',
        'AUDITOR',
        'DIRECTOR_FINANCE',
        'PRINCIPAL_OFFICER',
        'ACCOUNTANT',
        'ADMIN'
      ],
      children: [
        {
          id: 'pv-draft',
          label: 'Draft',
          icon: Edit,
          href: '/workflows/payment-vouchers?filter=draft',
          requiredRoles: ['ACCOUNTANT', 'FINANCE_OFFICER', 'ADMIN']
        },
        {
          id: 'pv-to-approve',
          label: 'For Approval',
          icon: CheckCircle2,
          href: '/workflows/payment-vouchers?filter=pending-approval',
          requiredRoles: [
            'DEPARTMENT_MANAGER',
            'AUDITOR',
            'DIRECTOR_FINANCE',
            'PRINCIPAL_OFFICER',
            'ADMIN'
          ],
          badge: {
            label: 'Dynamic Count',
            color: 'red'
          }
        },
        {
          id: 'pv-approved',
          label: 'Approved',
          icon: CheckCircle,
          href: '/workflows/payment-vouchers?filter=approved',
          requiredRoles: [
            'FINANCE_OFFICER',
            'DIRECTOR_FINANCE',
            'PRINCIPAL_OFFICER',
            'ADMIN'
          ]
        },
        {
          id: 'pv-all',
          label: 'All Vouchers',
          icon: List,
          href: '/workflows/payment-vouchers',
          requiredRoles: [
            'FINANCE_OFFICER',
            'DEPARTMENT_MANAGER',
            'AUDITOR',
            'DIRECTOR_FINANCE',
            'PRINCIPAL_OFFICER',
            'ACCOUNTANT',
            'ADMIN'
          ]
        }
      ]
    },

    // ========================================================================
    // SEARCH & TRANSACTIONS - Finance & Compliance
    // ========================================================================
    {
      id: 'transactions-section',
      label: 'Transactions',
      icon: Search,
      href: '/transactions',
      requiredRoles: [
        'FINANCE_OFFICER',
        'DIRECTOR_FINANCE',
        'PRINCIPAL_OFFICER',
        'AUDITOR',
        'ADMIN'
      ],
      children: [
        {
          id: 'trans-search',
          label: 'Search Transactions',
          icon: Search,
          href: '/transactions?tab=search',
          requiredRoles: [
            'FINANCE_OFFICER',
            'DIRECTOR_FINANCE',
            'AUDITOR',
            'ADMIN'
          ]
        },
        {
          id: 'trans-verify',
          label: 'Verify QR Code',
          icon: QrCode,
          href: '/transactions?tab=verify',
          requiredRoles: [
            'FINANCE_OFFICER',
            'DIRECTOR_FINANCE',
            'PRINCIPAL_OFFICER',
            'AUDITOR',
            'ADMIN'
          ]
        },
        {
          id: 'trans-reports',
          label: 'Reports',
          icon: BarChart3,
          href: '/transactions?tab=reports',
          requiredRoles: [
            'DIRECTOR_FINANCE',
            'PRINCIPAL_OFFICER',
            'AUDITOR',
            'ADMIN'
          ]
        }
      ]
    },

    // ========================================================================
    // ADMIN SECTION - System Administrators Only
    // ========================================================================
    {
      id: 'admin-divider',
      divider: true,
      label: '',
      icon: null,
      href: '',
      requiredRoles: ['ADMIN']
    },
    {
      id: 'admin-section',
      label: 'Administration',
      icon: Settings,
      href: '/admin',
      requiredRoles: ['ADMIN'],
      children: [
        {
          id: 'admin-users',
          label: 'Users',
          icon: Users,
          href: '/admin/users',
          requiredRoles: ['ADMIN']
        },
        {
          id: 'admin-roles',
          label: 'Roles & Permissions',
          icon: Shield,
          href: '/admin/roles',
          requiredRoles: ['ADMIN']
        },
        {
          id: 'admin-workflows',
          label: 'Workflow Config',
          icon: Settings,
          href: '/admin/workflows',
          requiredRoles: ['ADMIN']
        },
        {
          id: 'admin-audit',
          label: 'Audit Logs',
          icon: FileText,
          href: '/admin/audit-logs',
          requiredRoles: ['ADMIN']
        },
        {
          id: 'admin-access',
          label: 'Access Logs',
          icon: Eye,
          href: '/admin/access-logs',
          requiredRoles: ['ADMIN']
        }
      ]
    }
  ],

  // User menu (top right avatar dropdown)
  userMenu: [
    {
      id: 'profile',
      label: 'My Profile',
      icon: User,
      href: '/profile',
      requiredRoles: [
        'REQUESTER',
        'DEPARTMENT_MANAGER',
        'AUDITOR',
        'FINANCE_OFFICER',
        'DIRECTOR_FINANCE',
        'PRINCIPAL_OFFICER',
        'ACCOUNTANT',
        'ADMIN'
      ]
    },
    {
      id: 'my-approvals',
      label: 'My Pending Approvals',
      icon: CheckCircle2,
      href: '/dashboard?tab=pending-approvals',
      requiredRoles: [
        'DEPARTMENT_MANAGER',
        'AUDITOR',
        'FINANCE_OFFICER',
        'DIRECTOR_FINANCE',
        'PRINCIPAL_OFFICER',
        'ADMIN'
      ]
    },
    {
      id: 'settings',
      label: 'Settings',
      icon: Settings,
      href: '/settings',
      requiredRoles: [
        'REQUESTER',
        'DEPARTMENT_MANAGER',
        'AUDITOR',
        'FINANCE_OFFICER',
        'DIRECTOR_FINANCE',
        'PRINCIPAL_OFFICER',
        'ACCOUNTANT',
        'ADMIN'
      ]
    },
    {
      id: 'logout',
      label: 'Logout',
      icon: LogOut,
      href: '/logout',
      requiredRoles: [
        'REQUESTER',
        'DEPARTMENT_MANAGER',
        'AUDITOR',
        'FINANCE_OFFICER',
        'DIRECTOR_FINANCE',
        'PRINCIPAL_OFFICER',
        'ACCOUNTANT',
        'ADMIN'
      ]
    }
  ]
}
```

---

## Page Structure by Role

### REQUESTER (Role: REQUESTER)

**Available Pages**:
- Dashboard
- Requisitions
  - My Requisitions (filter: created-by-me)
  - All Requisitions (read-only)

**Features**:
- View dashboard with own statistics
- Create new requisition
- View requisitions they created
- View approval status
- Upload attachments
- View audit trail

**Not Accessible**:
- Purchase Orders
- Goods Received Note
- Payment Vouchers
- Transactions/Search
- Admin section

**Navigation**:
```
Dashboard
├── Requisitions
│   ├── My Requisitions
│   └── All Requisitions (read-only)
```

---

### DEPARTMENT MANAGER (Role: DEPARTMENT_MANAGER)

**Available Pages**:
- Dashboard
- Requisitions (all, with approval actions)
- Purchase Orders (approval and view)
- Payment Vouchers (approval and view)

**Features**:
- View requisitions needing approval
- Approve/reverse/reject requisitions at Stage 1
- View approved requisitions
- View created POs
- Approve/reverse/reject POs at Stage 1
- Approve/reverse/reject PVs at Stage 1

**Not Accessible**:
- GRN (no role)
- Transactions (no role)
- Admin section

**Navigation**:
```
Dashboard
├── Requisitions
│   ├── Pending Approvals
│   └── All Requisitions
├── Purchase Orders
│   ├── Pending Approvals
│   ├── In Process
│   ├── Approved
│   └── All POs
├── Payment Vouchers
│   ├── For Approval
│   ├── Approved
│   └── All Vouchers
```

---

### AUDITOR (Role: AUDITOR / COMPLIANCE_OFFICER)

**Available Pages**:
- Dashboard
- Purchase Orders (approval and view)
- Payment Vouchers (approval and view)
- Transactions (search and view)

**Features**:
- Review compliance in POs
- Approve/reverse/reject POs at Stage 2
- Review compliance in Payment Vouchers
- Approve/reverse/reject PVs at Stage 2
- Search transactions
- Generate compliance reports

**Not Accessible**:
- Requisitions (not an approver, read-only access might be added later)
- GRN
- Admin section

**Navigation**:
```
Dashboard
├── Purchase Orders
│   ├── Pending Approvals
│   ├── In Process
│   └── All POs
├── Payment Vouchers
│   ├── For Approval
│   ├── Approved
│   └── All Vouchers
├── Transactions
│   ├── Search Transactions
│   ├── Verify QR Code
│   └── Reports
```

---

### FINANCE OFFICER (Role: FINANCE_OFFICER)

**Available Pages**:
- Dashboard
- GRN (create, list, view)
- Payment Vouchers (create/update, approval, view)
- Transactions (search, verify)

**Features**:
- Create GRNs for received goods
- Mark goods as received
- Create Payment Vouchers (or hand off to Accountant)
- View payment vouchers
- Search transactions
- Download reports and PDFs

**Not Accessible**:
- Requisitions
- Purchase Orders (no approval role)
- Admin section

**Navigation**:
```
Dashboard
├── Goods Received
│   ├── Create GRN
│   ├── Pending Receipt
│   └── Completed
├── Payment Vouchers
│   ├── Draft
│   ├── All Vouchers
├── Transactions
│   ├── Search Transactions
│   └── Verify QR Code
```

---

### ACCOUNTANT (Role: ACCOUNTANT)

**Available Pages**:
- Dashboard
- Payment Vouchers (create/generate, manage, view)

**Features**:
- Generate Payment Vouchers from GRNs
- Fill in vendor and bank details
- Calculate taxes and totals
- Submit to approval chain
- View vouchers in various states
- See approval progress

**Not Accessible**:
- Requisitions
- Purchase Orders
- GRN
- Transactions
- Admin section

**Navigation**:
```
Dashboard
├── Payment Vouchers
│   ├── Draft (vouchers being worked on)
│   ├── For Approval (awaiting approval)
│   ├── Approved (completed)
│   └── All Vouchers
```

---

### FINANCE DIRECTOR (Role: DIRECTOR_FINANCE)

**Available Pages**:
- Dashboard
- Requisitions (approval, all)
- Purchase Orders (approval, all)
- GRN (view all)
- Payment Vouchers (approval, all)
- Transactions (search, verify, reports)

**Features**:
- Approve requisitions at Stage 3
- Approve POs at Stage 3
- Approve PVs at Stage 3
- View all GRNs
- Search transactions
- View financial reports
- Verify payments via QR code

**Not Accessible**:
- Admin section (unless also admin)

**Navigation**:
```
Dashboard
├── Requisitions
│   ├── Pending Approvals
│   └── All Requisitions
├── Purchase Orders
│   ├── Pending Approvals
│   ├── In Process
│   ├── Approved
│   └── All POs
├── Goods Received
│   └── Completed (view only)
├── Payment Vouchers
│   ├── For Approval
│   ├── Approved
│   └── All Vouchers
├── Transactions
│   ├── Search Transactions
│   ├── Verify QR Code
│   └── Reports
```

---

### PRINCIPAL OFFICER (Role: PRINCIPAL_OFFICER)

**Available Pages**:
- Dashboard
- Requisitions (approval, all)
- Purchase Orders (final approval, all)
- Payment Vouchers (final approval, all)
- Transactions (search, verify, reports)

**Features**:
- Final approval of requisitions at Stage 4
- Final approval of POs at Stage 4
- Final approval of PVs at Stage 4 (generates QR code)
- Executive-level reporting and verification
- Verify payments

**Not Accessible**:
- GRN
- Admin section (unless also admin)

**Navigation**:
```
Dashboard
├── Requisitions
│   ├── Pending Approvals
│   └── All Requisitions
├── Purchase Orders
│   ├── Pending Approvals
│   ├── In Process
│   ├── Approved
│   └── All POs
├── Payment Vouchers
│   ├── For Approval
│   ├── Approved
│   └── All Vouchers
├── Transactions
│   ├── Verify QR Code
│   └── Reports
```

---

### SYSTEM ADMINISTRATOR (Role: ADMIN)

**Available Pages**:
- Dashboard (system dashboard)
- All workflow pages (Requisitions, PO, GRN, PV)
- Transactions (all features)
- Admin section (full access)

**Features**:
- View all documents in system
- Manage users (create, edit, delete)
- Manage roles and permissions
- Configure approval workflows
- View audit logs
- View access logs
- Approve/reject any document (override)
- Emergency actions

**Available Actions**:
- Everything (administrative access)

**Navigation**:
```
Dashboard
├── Requisitions (all)
├── Purchase Orders (all)
├── Goods Received (all)
├── Payment Vouchers (all)
├── Transactions (all)
├── Admin
│   ├── Users
│   ├── Roles & Permissions
│   ├── Workflow Config
│   ├── Audit Logs
│   └── Access Logs
```

---

## Page Structure Examples

### Requisitions Page (/workflows/requisitions)

**Common Elements**:
- Header with "Requisitions" title
- Search/filter bar
- Table of requisitions

**Role-Specific Rendering**:

**REQUESTER**:
- Shows only: Filter button (filters to own requisitions)
- Shows actions: View, Edit (if draft)
- No approve/reject buttons
- Shows status: Draft, Submitted, Approved, Rejected

**DEPARTMENT_MANAGER / PRINCIPAL_OFFICER / DIRECTOR_FINANCE**:
- Shows: Filter dropdown (All, Pending Approval, Approved, Rejected)
- Shows columns: Document#, Status, Current Stage, Created By, Created Date, Actions
- Shows actions: View, Approve, Reverse, Comments, Audit Trail
- Approve/Reverse buttons only show if:
  - Current stage is their role's stage
  - Status is IN_APPROVAL
  - They have the approval permission

**ADMIN**:
- Shows all filters and actions
- Shows columns: Document#, Status, Current Stage, Created By, Created Date, Actions
- Show additional actions: Edit, Delete, Reset Workflow
- View any user's data without restriction

### Example: Purchase Order Detail Page

**Structure**:
```
PO Header
├── Document Info (PO Number, Date, Vendor)
├── Status Badge (Draft / In Approval / Approved / Rejected)
├── Stage Progress (Stage X of 4)
└── Approval Timeline

Content Tabs
├── Details (items, totals, vendor info)
├── Approvals (history and current stage)
├── Attachments
└── Audit Log

Actions Panel
├── [IF Stage 1 & Department Manager] Approve/Reverse/Reject buttons
├── [IF Stage 2 & Auditor] Approve/Reverse/Reject buttons
├── [IF Stage 3 & Finance Director] Approve/Reverse/Reject buttons
├── [IF Stage 4 & Principal Officer] Approve/Reverse buttons
├── [IF Admin] Any action
└── Comments field (if required by stage)
```

### Example: Payment Voucher Detail Page

**Structure**:
```
PV Header
├── Document Info (Voucher Number, Vendor, Amount)
├── Status Badge & Progress
└── Stage Indicator

Content Sections
├── Amount Summary
│   ├── Gross Amount
│   ├── Tax
│   └── Net Amount
├── Bank Information
│   ├── Account Number
│   ├── Account Name
│   └── Bank Code
├── Approval Progress
│   ├── Stage 1: Department Head (status, date, comments)
│   ├── Stage 2: Auditor (status, date, comments)
│   ├── Stage 3: Finance Director (status, date, comments)
│   └── Stage 4: Principal Officer (status, date, comments, QR Code)
└── Audit Trail

Actions Panel
├── [IF Draft & Accountant] Submit for Approval
├── [IF Stage N & Assigned Role] Approve/Reverse/Reject
├── [IF Final & Principal Officer] Approve (generates QR)
└── [IF Admin] Override any action

Special Elements
├── [IF Stage 3 & Finance Director] Bank validation checkbox
├── [IF Stage 4 & Principal Officer] QR Code display (after approval)
├── [IF Stage 4 Complete] Download PDF button
```

---

## Implementation Details

### Side Navigation Component

```typescript
// src/components/sidebar.tsx

import { filterNavItems } from '@/lib/nav-visibility'
import { navigationConfig } from '@/lib/navigation-config'
import { useCurrentUser } from '@/hooks/use-current-user'

export function Sidebar() {
  const { user } = useCurrentUser()

  // Get user's roles and permissions
  const userRoles = user?.roleIds || []
  const userPermissions = user?.permissions || []

  // Filter navigation to only items user can see
  const visibleNav = filterNavItems(
    navigationConfig.mainNav,
    userRoles,
    userPermissions
  )

  return (
    <nav className="space-y-4">
      {visibleNav.map(item => (
        <NavItem key={item.id} item={item} />
      ))}
    </nav>
  )
}
```

### Dynamic Page Titles

```typescript
// src/lib/page-titles.ts

export const pageTitles: Record<string, Record<string, string>> = {
  '/workflows/requisitions': {
    DEFAULT: 'Requisitions',
    REQUESTER: 'My Requisitions',
    DEPARTMENT_MANAGER: 'Requisitions to Review',
    PRINCIPAL_OFFICER: 'Executive Requisitions'
  },
  '/workflows/purchase-orders': {
    DEFAULT: 'Purchase Orders',
    DEPARTMENT_MANAGER: 'POs to Review',
    AUDITOR: 'POs for Audit'
  }
}
```

### Role-Based Feature Visibility

```typescript
// src/hooks/use-can-perform-action.ts

export function useCanPerformAction(action: string): boolean {
  const { user } = useCurrentUser()

  const actionMap: Record<string, string[]> = {
    'approve-requisition': ['DEPARTMENT_MANAGER', 'PRINCIPAL_OFFICER', 'DIRECTOR_FINANCE', 'ADMIN'],
    'reverse-po': ['DEPARTMENT_MANAGER', 'AUDITOR', 'DIRECTOR_FINANCE', 'PRINCIPAL_OFFICER', 'ADMIN'],
    'create-grn': ['FINANCE_OFFICER', 'ADMIN'],
    'approve-pv': ['DEPARTMENT_MANAGER', 'AUDITOR', 'DIRECTOR_FINANCE', 'PRINCIPAL_OFFICER', 'ADMIN'],
    'generate-qr': ['PRINCIPAL_OFFICER', 'ADMIN'],
    'manage-users': ['ADMIN'],
  }

  const allowedRoles = actionMap[action] || []
  return allowedRoles.some(role => user?.roleIds?.includes(role))
}
```

---

## Navigation Badges

### Dynamic Notification Badges

```typescript
// src/lib/nav-badges.ts

export async function getNavBadges(userId: string): Promise<Record<string, string>> {
  const pendingRequisitions = await getPendingRequisitions(userId)
  const pendingPOs = await getPendingPurchaseOrders(userId)
  const pendingPVs = await getPendingPaymentVouchers(userId)

  return {
    'req-to-approve': String(pendingRequisitions.count),
    'po-to-approve': String(pendingPOs.count),
    'pv-to-approve': String(pendingPVs.count),
  }
}

// Usage in component:
// Shows red badge with number of pending approvals
{item.badge && (
  <Badge variant="destructive">{badgeCount}</Badge>
)}
```

---

## Mobile Navigation

On mobile devices, use:
- Hamburger menu that opens sidebar
- Bottom navigation bar (optional)
- Responsive design that stacks navigation items

```typescript
// Mobile bottom navigation for frequently used items
const mobileBottomNav: NavigationItem[] = [
  { id: 'dashboard', label: 'Dashboard', ... },
  { id: 'pending', label: 'Pending', ... }, // Shows pending approvals
  { id: 'profile', label: 'Profile', ... }
]
```

---

## Implementation Checklist

### Phase 0 (New)
- [ ] Create `src/types/navigation.ts`
- [ ] Create `src/lib/navigation-config.ts` with complete config
- [ ] Create `src/lib/nav-visibility.ts` with filtering logic
- [ ] Create `src/hooks/use-can-perform-action.ts`
- [ ] Create `src/lib/nav-badges.ts` for dynamic counts
- [ ] Create `src/lib/page-titles.ts` for dynamic titles

### Phase 1 Integration
- [ ] Update `src/components/sidebar.tsx` to use new navigation
- [ ] Update requisitions page to respect roles
- [ ] Add navigation badges for requisitions

### Phase 2A Integration (PO)
- [ ] Ensure PO pages only accessible to approved roles
- [ ] Add PO navigation section
- [ ] Add PO approval badges

### Phase 2B Integration (GRN)
- [ ] Ensure GRN pages only accessible to Finance Officer/Admin
- [ ] Add GRN navigation section

### Phase 2C Integration (PV)
- [ ] Ensure PV pages only accessible to approved roles
- [ ] Add PV navigation section
- [ ] Add PV approval badges

### Phase 2D Integration (Transactions)
- [ ] Add Transactions navigation section
- [ ] Ensure only Finance/Auditor/Admin see search

### Phase 3+ Integration
- [ ] Update dashboard to show role-specific widgets
- [ ] Update notifications based on role
- [ ] Admin pages security

---

## Security Considerations

### Frontend Checks
- Hide UI elements based on roles (UX improvement)
- Disable buttons that user can't click
- Hide menu items user can't access

### Backend Validation
- **CRITICAL**: Always validate permissions on server
- Check user roles on every server action
- Never trust frontend role checks alone
- Log unauthorized access attempts

```typescript
// src/app/_actions/approval.ts - ALWAYS validate

export async function approveDocument(request: ApproveDocumentRequest) {
  // 1. Get current user
  const user = await getCurrentUser()

  // 2. Load document and approval state
  const state = store.approvalStates.get(request.documentId)

  // 3. Get current stage from config
  const stage = getCurrentApprovalStage(state)

  // 4. VALIDATE: User must have required role
  const userRoles = user?.roleIds || []
  if (!userRoles.includes(stage.requiredRole)) {
    // Log unauthorized attempt
    logUnauthorizedAccess(user.id, 'approveDocument', request.documentId)

    // Return error
    return {
      success: false,
      error: 'UNAUTHORIZED'
    }
  }

  // 5. Only then proceed with approval
  // ...
}
```

---

## Summary

This navigation system ensures:

✅ **Visibility**: Each user sees only what they can access
✅ **Usability**: Navigation is intuitive for each role
✅ **Security**: Backend validates every action
✅ **Flexibility**: Easy to add new roles or modify access
✅ **Maintainability**: Configuration-driven, not hardcoded
✅ **Scalability**: Supports any number of roles and pages

Every role has clear visibility into:
- Which pages they can access
- Which actions they can perform
- What documents they need to review
- Their current pending approvals

---

**Created**: 2024-11-29
**Status**: Ready for Implementation
**Next Step**: Implement in Phase 0 expansion (2-3 hours)
