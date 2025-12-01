# Liyali Gateway - Features Documentation

## Overview

Liyali Gateway is a comprehensive workflow management and document approval system built with Next.js 16, React, and TypeScript. It streamlines business processes by providing tools for budget management, task management, requisition handling, purchase order approval, payment vouchers, and goods received notes.

## Core Features

### 1. Budget Management Module

#### Overview
Complete budget lifecycle management with multi-stage approval workflows.

#### Key Features
- **Budget Creation**: Create budgets with department allocation, fiscal year, and currency (ZMW/USD)
- **Budget Items**: Add detailed line items with category, description, and allocated amounts
- **Budget Tracking**: Real-time spending visualization with progress indicators
- **Multi-Stage Approvals**: Configurable approval chains with role-based access
- **Budget Details Page**: Tabbed interface organizing information across:
  - Overview tab: Budget summary cards (Total, Spent, Remaining)
  - Information tab: Detailed budget metadata
  - Items tab: Budget line items with management
  - Approvals tab: Approval chain history

#### Data Structure
```typescript
interface Budget {
  id: string
  budgetNumber: string // e.g., "BDG-2024-001"
  name: string
  description?: string
  department: string
  departmentId: string
  fiscalYear: string // e.g., "2024"
  totalAmount: number
  currency: string // "ZMW" or "USD"
  items: BudgetItem[]
  status: BudgetStatus // DRAFT, SUBMITTED, IN_APPROVAL, APPROVED, REJECTED
  createdBy: string
  createdAt: Date
  updatedAt: Date
  approvalChain?: ApprovalRecord[]
  currentApprovalStage?: number
  totalApprovalStages?: number
}
```

#### Files
- `src/types/budget.ts` - Type definitions
- `src/app/_actions/budgets.ts` - Server actions
- `src/app/(private)/workflows/budgets/page.tsx` - Budget list page
- `src/app/(private)/workflows/budgets/[id]/page.tsx` - Budget detail page
- `src/app/(private)/workflows/budgets/_components/` - Budget components
  - `budgets-client.tsx` - Main client component
  - `budgets-table.tsx` - Budget list table with pagination
  - `create-budget-dialog.tsx` - Budget creation form
  - `budget-detail-client.tsx` - Budget detail view with tabs
  - `budget-items-table.tsx` - Budget items table
  - `approval-chain-panel.tsx` - Approval history visualization
  - `budget-approval-action-panel.tsx` - Approval/rejection interface

### 2. Tasks Management Module

#### Overview
Manages pending workflow tasks and actions required from users.

#### Key Features
- **Task Dashboard**: View all pending, in-progress, and completed tasks
- **Task Statistics**: Quick overview with pending count, overdue tasks, urgent items
- **Task Filtering**: Filter tasks by status (All, Pending, In Progress)
- **Task Types**: Budget approvals, requisition approvals, PO approvals, payment vouchers, GRN confirmations
- **Quick Actions**: Direct access to approval documents from task list
- **Pagination**: Customizable page sizes (10, 20, 30, 40, 50 items)

#### Data Structure
```typescript
interface Task {
  id: string
  type: TaskType // BUDGET_APPROVAL, REQUISITION_APPROVAL, etc.
  title: string
  description: string
  assignedTo: string
  assignedRole: string
  status: TaskStatus // PENDING, IN_PROGRESS, COMPLETED
  priority: TaskPriority // URGENT, HIGH, MEDIUM, LOW
  documentType: string
  documentId: string
  documentNumber: string
  createdAt: Date
  dueAt: Date
  actionUrl: string // Route to perform action
}
```

#### Files
- `src/types/tasks.ts` - Type definitions
- `src/app/_actions/tasks.ts` - Server actions
- `src/app/(private)/workflows/tasks/page.tsx` - Tasks page
- `src/app/(private)/workflows/tasks/_components/` - Task components
  - `tasks-client.tsx` - Main client component with filtering
  - `tasks-table.tsx` - Tasks table with pagination
  - `task-stats-cards.tsx` - Task statistics display

### 3. Settings & Profile Management

#### Overview
Comprehensive user account and preference management.

#### Key Features
- **Account Settings**: Edit profile name, email, department
- **Password Management**: Secure password change with validation
- **Preferences**: Language, theme, timezone, notification settings
- **Session Management**: View and revoke active login sessions
- **Error Feedback**: Clear error and success messages for all operations

#### Files
- `src/types/auth.ts` - Authentication types
- `src/app/_actions/settings.ts` - Settings server actions
- `src/app/(private)/settings/page.tsx` - Settings page
- `src/app/(private)/settings/_components/` - Settings components
  - `settings-client.tsx` - Main settings interface with tabs
  - `account-settings.tsx` - Profile management
  - `change-password.tsx` - Password change form
  - `general-settings.tsx` - Preferences and notifications
  - `sessions-management.tsx` - Active session management

### 4. Approval Workflow System

#### Overview
Comprehensive approval and signature system for document authorization.

#### Key Features
- **Digital Signatures**: Canvas-based signature capture and storage (base64 PNG)
- **Signature Requirements**: Mandatory for all approvals
- **Remarks/Rejection Feedback**: Required detailed remarks for rejections, optional for approvals
- **Approval History**: Complete audit trail showing:
  - Approver identity and role
  - Action timestamp
  - Comments and remarks
  - Digital signature image preview
  - Approval stage number
- **Multi-Stage Approvals**: Support for sequential and role-based approval chains

#### Approval Flow
1. **Document Submitted**: Document enters approval workflow
2. **Approval Pending**: Assigned approver reviews document
3. **Action Required**: Approver can:
   - **Approve**: Add optional comments + required signature
   - **Reject**: Add required remarks + optional comments
4. **Approval Recorded**: Action stored with full audit trail
5. **Next Stage/Completion**: Move to next approver or completion

#### Files
- `src/types/workflow.ts` - Workflow type definitions with approval fields
- `src/types/budget.ts` - Budget approval record types
- `src/components/ui/signature-canvas.tsx` - Signature drawing component
- `src/app/(private)/workflows/requisitions/_components/approval-action-panel.tsx` - Requisition approval
- `src/app/(private)/workflows/budgets/[id]/_components/budget-approval-action-panel.tsx` - Budget approval
- `src/app/(private)/workflows/budgets/[id]/_components/approval-chain-panel.tsx` - Approval history display

### 5. Dashboard & Analytics (Existing)

#### Overview
Real-time metrics and workflow insights.

#### Key Features
- **KPI Cards**: Key performance indicators
- **Approval Time Chart**: Average approval time analytics
- **Recent Documents**: Quick access to recent items
- **System Health**: Resource usage and system status

### 6. User Interface Components

#### Reusable Components
- **Custom Pagination**: Page navigation with size selector
- **Status Badge**: Colored status indicators with role-based styling
- **Signature Canvas**: Digital signature capture
- **Tabs**: Tabbed content navigation
- **Cards**: Section containers
- **Dialogs**: Modal forms and confirmations
- **Forms**: Input fields with validation

#### Navigation
- **Sidebar Navigation**: Role-aware menu with sections:
  - Workflows (Dashboard, Tasks, Search, Requisitions, POs, Payment Vouchers, GRNs, Budgets)
  - Admin (Reports, User Management, Activity Logs)
  - Compliance & Monitoring
  - Settings
- **Header**: User menu with profile and logout
- **Responsive Design**: Mobile, tablet, and desktop support

### 7. Authentication & Authorization (Existing)

#### Overview
Session-based authentication with role-based access control.

#### Supported Roles
- REQUESTER
- DEPARTMENT_MANAGER
- FINANCE_OFFICER
- DIRECTOR
- CFO
- COMPLIANCE_OFFICER
- ADMIN

## Technical Stack

### Frontend
- **Framework**: Next.js 16 with Turbopack
- **UI Library**: React 19
- **Styling**: Tailwind CSS
- **Components**: shadcn/ui
- **Tables**: TanStack React Table v8
- **Icons**: Lucide React
- **Notifications**: Sonner
- **Language**: TypeScript

### Backend
- **Runtime**: Node.js
- **Framework**: Next.js API Routes
- **Server Actions**: Next.js Server Components
- **Data Handling**: In-memory (mock data)
- **Authentication**: Next.js Auth (session-based)

### Data Storage (Mock Implementation)
Currently using in-memory mock data. Production ready for:
- PostgreSQL
- MongoDB
- Firebase
- Any REST/GraphQL API

## Security Features

- **Server Actions**: Secure server-side operations
- **Digital Signatures**: Audit trail with signature verification
- **Session Management**: Active session tracking and revocation
- **Role-Based Access**: Permission-based feature access
- **Audit Logging**: Complete action history with timestamps
- **Input Validation**: Form and server-side validation

## Performance Optimizations

- **React Caching**: Server action caching with React cache()
- **Code Splitting**: Route-based code splitting
- **Pagination**: Efficient data loading with customizable page sizes
- **TanStack Table**: Headless table for minimal re-renders
- **Lazy Loading**: Components and images loaded on demand

## Accessibility

- **Semantic HTML**: Proper heading hierarchy and ARIA labels
- **Keyboard Navigation**: Full keyboard support
- **Color Contrast**: WCAG AA compliant colors
- **Form Labels**: Associated labels for all inputs
- **Error Messages**: Clear, descriptive validation feedback

## Browser Support

- Chrome/Edge (latest 2 versions)
- Firefox (latest 2 versions)
- Safari (latest 2 versions)
- Mobile browsers (iOS Safari, Chrome Mobile)

## Internationalization Ready

- Language selection (English, Spanish, French, Portuguese)
- Timezone support
- Currency support (ZMW, USD)
- Date/Time localization

---

**Last Updated**: 2025-11-30
**Version**: 1.0.0
**Status**: Active Development
