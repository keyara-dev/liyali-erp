# Liyali Gateway - Complete Implementation Plan

## Project Overview

Building a comprehensive Workflow Management System for Mitete Town Council with requisition-to-payment workflow, admin functions, reporting, and user management.

---

## Architecture & Patterns

### Established Patterns (From Existing Code)

1. **Page Structure**:
   - `src/app/[module]/page.tsx` - Server component with auth
   - `src/app/[module]/_components/[module]-client.tsx` - Client wrapper
   - `src/app/[module]/_components/[module]-table.tsx` - React Table component
   - `src/app/[module]/[id]/page.tsx` - Detail server component
   - `src/app/[module]/[id]/_components/[module]-detail.tsx` - Detail client component

2. **Data Layer**:
   - `src/app/_actions/[feature].ts` - Server actions
   - `src/lib/mock-data.ts` - Mock data generators
   - In-memory Maps for data storage
   - Types in `src/types/workflow.ts`

3. **UI Components**:
   - Tailwind CSS v4 with custom color system
   - Primary: Blue (#0c54e7), Secondary: Green (#10b981), Accent: Amber (#f59e0b)
   - React Table for data tables
   - Shadcn/ui components (Button, Card, Badge, etc.)

4. **Authentication & Authorization**:
   - NextAuth with role-based access control
   - 7 roles: REQUESTER, DEPARTMENT_MANAGER, FINANCE_OFFICER, DIRECTOR, CFO, COMPLIANCE_OFFICER, ADMIN

---

## Implementation Phases

### Phase 1: High Priority - User Facing (Weeks 1-2)

#### 1.1 Search & Past Transactions Page
**Route**: `/workflows/search`

**Components**:
- `src/app/workflows/search/page.tsx` - Server page with auth
- `src/app/workflows/search/_components/search-client.tsx` - Client wrapper
- `src/app/workflows/search/_components/search-form.tsx` - Search form with filters
- `src/app/workflows/search/_components/transaction-results.tsx` - Results table
- `src/app/workflows/search/_components/download-button.tsx` - PDF download

**Features**:
- Search by reference number (REQ-*, PO-*, GRN-*, PV-*)
- Filter by date range
- Filter by status
- Filter by document type
- Display results in table (Document #, Type, Amount, Status, Date)
- Download PDFs for selected transactions
- QR code verification badge

**Data Actions**:
- Create `searchDocuments()` in `src/app/_actions/search.ts`
- Search across all document types
- Return paginated results
- Support filtering and sorting

**Mock Data**:
- Generate 20-30 past transactions across all types
- Mix of all statuses (DRAFT, SUBMITTED, IN_APPROVAL, APPROVED, REJECTED)
- Various date ranges

---

#### 1.2 Dashboard with Key Metrics
**Route**: `/dashboard`

**Components**:
- `src/app/dashboard/page.tsx` - Server page with auth
- `src/app/dashboard/_components/dashboard-client.tsx` - Client wrapper
- `src/app/dashboard/_components/metrics-cards.tsx` - KPI cards
- `src/app/dashboard/_components/quick-actions.tsx` - Action buttons
- `src/app/dashboard/_components/recent-activity.tsx` - Recent transactions
- `src/app/dashboard/_components/workflow-status-chart.tsx` - Status breakdown chart
- `src/app/dashboard/_components/approval-time-chart.tsx` - Approval timeline chart

**Metrics to Display**:
- Total Pending Approvals (number)
- Total Submitted This Month (number)
- Average Approval Time (hours)
- Success Rate (%)
- Chart 1: Documents by Status (pie/donut)
- Chart 2: Approval Trends (line chart - last 30 days)
- Quick Actions: Create Requisition, View Pending, Download Reports
- Recent Activity: Last 5 transactions

**Data Actions**:
- Create `getDashboardMetrics()` in `src/app/_actions/dashboard.ts`
- Calculate metrics from document store
- Return role-specific filtered data

**Color Scheme**:
- Use new color system for charts
- Primary blue for submitted
- Secondary green for approved
- Accent amber for pending
- Destructive red for rejected

---

#### 1.3 Requisition Creation Form
**Route**: `/workflows/requisitions/create`

**Components**:
- `src/app/workflows/requisitions/create/page.tsx` - Server page with auth
- `src/app/workflows/requisitions/create/_components/create-form.tsx` - Form component
- `src/app/workflows/requisitions/create/_components/item-input.tsx` - Repeatable item inputs
- `src/app/workflows/requisitions/create/_components/form-preview.tsx` - Preview before submit

**Form Fields**:
- Department (dropdown from mock users)
- Requested For (text input)
- Budget Code (dropdown/text)
- Justification (textarea)
- Items (repeatable):
  - Item Description
  - Quantity
  - Estimated Unit Cost
  - Total Cost (calculated)
- Total Estimated Cost (calculated)
- Submit button

**Features**:
- Form validation
- Automatic total calculation
- Save as draft
- Submit for approval
- Confirmation dialog before submit
- Success toast notification
- Redirect to detail page after submit

**Data Actions**:
- Update `createWorkflowDocument()` in `src/app/_actions/workflow.ts`
- Generate REQ-YYYY-XXX document number
- Create mock approvers based on approval config
- Store in document store

---

### Phase 2: Medium Priority - Admin Functions (Weeks 3-4)

#### 2.1 Reporting & Dashboards
**Route**: `/admin/reporting`

**Components**:
- `src/app/admin/reporting/page.tsx` - Server page with auth (ADMIN/DIRECTOR only)
- `src/app/admin/reporting/_components/reporting-client.tsx` - Client wrapper
- `src/app/admin/reporting/_components/transaction-volume.tsx` - Volume by department chart
- `src/app/admin/reporting/_components/pending-approvals.tsx` - Pending summary table
- `src/app/admin/reporting/_components/approval-time-analysis.tsx` - Approval time metrics
- `src/app/admin/reporting/_components/budget-analysis.tsx` - Budget vs actual comparison

**Reports**:
1. Transaction Volume by Department (bar chart)
2. Pending Approvals Summary (table)
3. Average Approval Time (line chart)
4. Budget vs Actual Analysis (comparison chart)

**Data Actions**:
- Create `getReportingData()` in `src/app/_actions/admin.ts`
- Aggregate data by department
- Calculate approval metrics
- Budget tracking

---

#### 2.2 User Access Management
**Route**: `/admin/users`

**Components**:
- `src/app/admin/users/page.tsx` - Server page with auth
- `src/app/admin/users/_components/users-client.tsx` - Client wrapper
- `src/app/admin/users/_components/users-table.tsx` - Users table
- `src/app/admin/users/_components/add-user-dialog.tsx` - Add user form
- `src/app/admin/users/_components/role-assignment.tsx` - Role assignment
- `src/app/admin/users/_components/mfa-setup.tsx` - MFA configuration

**Features**:
- List all users (table)
- Add new users (form dialog)
- Assign/change roles
- Enable/disable MFA
- View access logs
- Delete users

**Columns**:
- Name
- Email
- Department
- Role
- Status (Active/Inactive)
- MFA Enabled
- Last Login
- Actions

**Data Actions**:
- Create `getUsers()`, `createUser()`, `updateUserRole()`, `enableMFA()` in `src/app/_actions/user-management.ts`
- Store users (in mock store for now)

---

#### 2.3 Activity Logs & History
**Route**: `/admin/logs`

**Components**:
- `src/app/admin/logs/page.tsx` - Server page with auth
- `src/app/admin/logs/_components/logs-client.tsx` - Client wrapper
- `src/app/admin/logs/_components/logs-table.tsx` - Activity log table
- `src/app/admin/logs/_components/log-filters.tsx` - Filter options
- `src/app/admin/logs/_components/log-detail-modal.tsx` - Detailed view

**Features**:
- View all user activities
- Filter by:
  - User
  - Action type (APPROVED, REJECTED, COMMENTED, etc.)
  - Document type
  - Date range
- Display columns:
  - Timestamp
  - User
  - Action
  - Document
  - Details
  - IP Address (if available)
- Export logs to CSV

**Data Actions**:
- Create `getActivityLogs()` in `src/app/_actions/admin.ts`
- Query approval logs from store
- Support filtering and pagination

---

### Phase 3: Low Priority - Future Features (Week 5+)

#### 3.1 Compliance Tracking
**Route**: `/admin/compliance`

- Government regulation alignment
- Compliance audit trails
- Future features placeholder

#### 3.2 Real-Time Monitoring
**Route**: `/admin/monitoring`

- Live transaction tracking
- Status updates in real-time
- Performance metrics

#### 3.3 QR Code Verification
**Integrated into**:
- Document detail pages
- Search results
- PDF documents

---

## File Structure Summary

```
src/
├── app/
│   ├── dashboard/                           [Phase 1]
│   │   ├── page.tsx
│   │   └── _components/
│   │       ├── dashboard-client.tsx
│   │       ├── metrics-cards.tsx
│   │       ├── quick-actions.tsx
│   │       ├── recent-activity.tsx
│   │       ├── workflow-status-chart.tsx
│   │       └── approval-time-chart.tsx
│   ├── workflows/
│   │   ├── search/                          [Phase 1]
│   │   │   ├── page.tsx
│   │   │   └── _components/
│   │   │       ├── search-client.tsx
│   │   │       ├── search-form.tsx
│   │   │       ├── transaction-results.tsx
│   │   │       └── download-button.tsx
│   │   ├── requisitions/
│   │   │   ├── create/                      [Phase 1]
│   │   │   │   ├── page.tsx
│   │   │   │   └── _components/
│   │   │   │       ├── create-form.tsx
│   │   │   │       ├── item-input.tsx
│   │   │   │       └── form-preview.tsx
│   │   │   ├── [existing files...]
│   ├── admin/                               [Phase 2]
│   │   ├── reporting/
│   │   │   ├── page.tsx
│   │   │   └── _components/
│   │   │       ├── reporting-client.tsx
│   │   │       ├── transaction-volume.tsx
│   │   │       ├── pending-approvals.tsx
│   │   │       ├── approval-time-analysis.tsx
│   │   │       └── budget-analysis.tsx
│   │   ├── users/
│   │   │   ├── page.tsx
│   │   │   └── _components/
│   │   │       ├── users-client.tsx
│   │   │       ├── users-table.tsx
│   │   │       ├── add-user-dialog.tsx
│   │   │       ├── role-assignment.tsx
│   │   │       └── mfa-setup.tsx
│   │   └── logs/
│   │       ├── page.tsx
│   │       └── _components/
│   │           ├── logs-client.tsx
│   │           ├── logs-table.tsx
│   │           ├── log-filters.tsx
│   │           └── log-detail-modal.tsx
│   └── _actions/
│       ├── dashboard.ts                     [New - Phase 1]
│       ├── search.ts                        [New - Phase 1]
│       └── admin.ts                         [New - Phase 2]
│
└── lib/
    ├── mock-data.ts                         [Update - Add more data]
    └── search-helpers.ts                    [New - Phase 1]
```

---

## Data & Storage

### Existing Mock Data
- Users (7 roles × 2-3 users per role)
- Documents (REQ, PO, GRN, PV)
- Approval logs
- Approvers

### New Mock Data Needed
- 20-30 past transactions (for search/dashboard)
- Department data
- Budget data
- Activity logs

### Data Stores (In-Memory Maps)
```typescript
documentStore: Map<string, WorkflowDocument>
approversStore: Map<string, Approver[]>
approvalLogsStore: Map<string, ApprovalLogEntry[]>
attachmentsStore: Map<string, Attachment[]>
usersStore: Map<string, User>                    [New - Phase 2]
activityLogsStore: Map<string, ActivityLog>     [New - Phase 2]
```

---

## Database & State Management

**Current**: In-memory Maps (suitable for demo/MVP)

**Future Migration**:
- PostgreSQL or similar
- Prisma ORM
- Real file storage for attachments
- Redis for caching

---

## Security & Authorization

### Role-Based Access Control
```
REQUESTER
  ✓ Create requisitions
  ✓ View own documents
  ✓ Edit drafts

DEPARTMENT_MANAGER
  ✓ Approve requisitions
  ✓ View department documents

FINANCE_OFFICER
  ✓ Create POs and GRNs
  ✓ Manage payment vouchers

DIRECTOR
  ✓ Approve high-value items
  ✓ View all documents

CFO
  ✓ Final approval on all documents
  ✓ Access all reports

COMPLIANCE_OFFICER
  ✓ View audit logs
  ✓ Access compliance reports

ADMIN
  ✓ Full system access
  ✓ User management
  ✓ Configuration
```

### Route Protection
- All routes check authentication
- Admin routes check role permissions
- Enforce via middleware and server actions

---

## UI/UX Patterns

### Color System
- **Primary (Blue #0c54e7)**: Main actions, buttons, links
- **Secondary (Green #10b981)**: Success, approved states
- **Accent (Amber #f59e0b)**: Warnings, pending states
- **Destructive (Red)**: Errors, rejections

### Status Badges
```
DRAFT        → Gray background
SUBMITTED    → Blue background
IN_APPROVAL  → Amber background
APPROVED     → Green background
REJECTED     → Red background
```

### Tables
- React Table with sorting/filtering
- Pagination (10 items per page)
- Status badges with colors
- Action menus (View, Download, Approve, Reject)
- Hover effects on rows

### Forms
- Clear labels and help text
- Validation feedback
- Required field indicators
- Submit/Cancel buttons
- Auto-save drafts (future)

### Charts
- Recharts or Chart.js
- Use color system for series
- Responsive design
- Legend on all charts

---

## Testing Strategy

### Unit Tests
- Helper functions
- Data transformations
- Validation logic

### Integration Tests
- Server actions
- Data store operations
- Role-based access

### E2E Tests
- User workflows
- Form submissions
- Approval chains

### Manual Testing
- Light/dark mode
- Mobile responsiveness
- Accessibility (WCAG AA+)
- Color contrast

---

## Development Checklist

### Phase 1
- [ ] Search page structure
- [ ] Search form and filters
- [ ] Transaction results table
- [ ] Dashboard layout
- [ ] Metric cards
- [ ] Charts (2 types)
- [ ] Requisition create form
- [ ] Form validation
- [ ] Mock data generation
- [ ] Testing

### Phase 2
- [ ] Admin reporting pages
- [ ] User management pages
- [ ] Activity logs
- [ ] Data aggregation
- [ ] Admin authorization

### Phase 3
- [ ] Compliance module
- [ ] Real-time updates
- [ ] QR code generation

---

## Known Limitations & Future Work

### Current MVP Limitations
- In-memory data storage (resets on restart)
- No file uploads/attachments
- No real email notifications
- No digital signature verification
- No QR code generation
- Limited reporting options

### Future Enhancements
- Database integration
- File storage (AWS S3, etc.)
- Email notifications
- SMS alerts
- Advanced analytics
- ML-based approval recommendations
- API integrations
- Mobile app

---

## Success Criteria

✅ All Phase 1 pages created and functional
✅ Search works across all document types
✅ Dashboard shows accurate metrics
✅ Requisition creation works end-to-end
✅ All pages have proper authorization
✅ Color system consistently applied
✅ Responsive design (mobile/tablet/desktop)
✅ WCAG AA+ accessibility
✅ No console errors or warnings
✅ Performance: pages load <3 seconds

---

## Timeline Estimate

| Phase | Duration | Status |
|-------|----------|--------|
| Phase 1 (High Priority) | 1-2 weeks | Starting |
| Phase 2 (Medium Priority) | 1-2 weeks | Pending |
| Phase 3 (Low Priority) | 1+ weeks | Future |

---

**Document Version**: 1.0
**Created**: November 29, 2024
**Status**: Ready for Implementation
