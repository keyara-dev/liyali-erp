# Workflow Management System - Implementation Guide

## Overview

This document describes the complete workflow management system architecture, including the newly implemented admin UI for creating and managing custom approval workflows.

## Folder Structure

### Core Workflow Infrastructure

```
src/
├── app/(private)/
│   ├── admin/
│   │   ├── workflows/                    # Workflow Management Admin Panel (NEW)
│   │   │   ├── page.tsx                  # List workflows
│   │   │   ├── _components/
│   │   │   │   ├── workflows-client.tsx  # Main workflows list component
│   │   │   │   ├── workflow-builder.tsx  # Drag-and-drop workflow builder
│   │   │   │   ├── workflow-details-form.tsx  # Workflow basic info form
│   │   │   │   ├── stage-form.tsx        # Individual stage configuration
│   │   │   │   └── stage-item.tsx        # Stage card with drag handle
│   │   │   ├── create/
│   │   │   │   ├── page.tsx              # Create workflow page
│   │   │   │   └── _components/
│   │   │   │       └── create-workflow-client.tsx  # Create workflow component
│   │   │   └── [id]/
│   │   │       └── edit/
│   │   │           ├── page.tsx          # Edit workflow page
│   │   │           └── _components/
│   │   │               └── edit-workflow-client.tsx  # Edit workflow component
│   │   │
│   │   ├── reports/                      # Admin Reports
│   │   ├── users/                        # User Management
│   │   └── logs/                         # Activity Logs
│   │
│   ├── (main)/                           # Main application routes group
│   │   ├── home/                         # Dashboard/Home page
│   │   ├── requisitions/                 # Purchase Requisitions
│   │   ├── purchase-orders/              # Purchase Orders
│   │   ├── payment-vouchers/             # Payment Vouchers
│   │   ├── grn/                          # Goods Received Notes
│   │   ├── budgets/                      # Budget Management
│   │   ├── tasks/                        # Task Management
│   │   ├── search/                       # Transaction Search
│   │   └── notifications/                # Notifications
│   │
│   ├── settings/                         # User Settings
│   ├── verification/                     # QR Verification
│   └── compliance/                       # Compliance & Monitoring
│
├── components/
│   ├── layout/
│   │   └── sidebar/
│   │       └── nav-main.tsx              # Navigation menu
│   ├── workflows/                        # Shared workflow components
│   │   ├── approval-action-panel.tsx
│   │   ├── approval-flow-display.tsx
│   │   └── ... (other workflow UI components)
│   └── ... (other UI components)
│
├── types/
│   └── custom-workflow.ts                # Workflow TypeScript types
│
├── app/_actions/
│   └── workflows.ts                      # Server-side workflow actions
│
└── hooks/
    └── use-workflows.ts                  # React Query hooks for workflows
```

## New Pages and Routes

### Workflow Management Admin Panel

#### 1. Workflows List Page

- **Route:** `/admin/workflows`
- **File:** `src/app/(private)/admin/workflows/page.tsx`
- **Component:** `WorkflowsClient`
- **Features:**
  - View all created workflows in a sortable table
  - Filter by status (ACTIVE, DEPRECATED)
  - Quick actions: Edit, Duplicate, Delete
  - Create new workflow button
  - Shows workflow details: name, document type, stages, status, last updated

#### 2. Create Workflow Page

- **Route:** `/admin/workflows/create`
- **File:** `src/app/(private)/admin/workflows/create/page.tsx`
- **Component:** `CreateWorkflowClient` → `WorkflowBuilder`
- **Features:**
  - Multi-step form for creating workflows
  - Configure workflow basic details (name, description, document type)
  - Drag-and-drop interface for adding/ordering approval stages
  - Each stage configurable with:
    - Name and description
    - Approver role (from predefined list)
    - Number of required approvals
    - Permissions (can reject, can reassign)

#### 3. Edit Workflow Page

- **Route:** `/admin/workflows/[id]/edit`
- **File:** `src/app/(private)/admin/workflows/[id]/edit/page.tsx`
- **Component:** `EditWorkflowClient` → `WorkflowBuilder`
- **Features:**
  - Same as create, but loads existing workflow data
  - Modify all workflow aspects
  - Update and save changes

### Main Application Routes

The following routes are for document workflow execution (not definition):

| Route                             | Purpose             | File                                                               |
| --------------------------------- | ------------------- | ------------------------------------------------------------------ |
| `/home`                           | Dashboard/Home      | `src/app/(private)/(main)/home/page.tsx`                           |
| `/requisitions`                   | Requisition list    | `src/app/(private)/(main)/requisitions/page.tsx`                   |
| `/requisitions/create`            | Create requisition  | `src/app/(private)/(main)/requisitions/create/page.tsx`            |
| `/requisitions/[id]`              | Requisition details | `src/app/(private)/(main)/requisitions/[id]/page.tsx`              |
| `/purchase-orders`                | PO list             | `src/app/(private)/(main)/purchase-orders/page.tsx`                |
| `/purchase-orders/[id]`           | PO details          | `src/app/(private)/(main)/purchase-orders/[id]/page.tsx`           |
| `/purchase-orders/[id]/approval`  | PO approval         | `src/app/(private)/(main)/purchase-orders/[id]/approval/page.tsx`  |
| `/payment-vouchers`               | PV list             | `src/app/(private)/(main)/payment-vouchers/page.tsx`               |
| `/payment-vouchers/[id]`          | PV details          | `src/app/(private)/(main)/payment-vouchers/[id]/page.tsx`          |
| `/payment-vouchers/[id]/approval` | PV approval         | `src/app/(private)/(main)/payment-vouchers/[id]/approval/page.tsx` |
| `/grn`                            | GRN list            | `src/app/(private)/(main)/grn/page.tsx`                            |
| `/grn/[id]`                       | GRN details         | `src/app/(private)/(main)/grn/[id]/page.tsx`                       |
| `/grn/[id]/confirmation`          | GRN confirmation    | `src/app/(private)/(main)/grn/[id]/confirmation/page.tsx`          |
| `/budgets`                        | Budget list         | `src/app/(private)/(main)/budgets/page.tsx`                        |
| `/budgets/[id]`                   | Budget details      | `src/app/(private)/(main)/budgets/[id]/page.tsx`                   |
| `/budgets/[id]/approval`          | Budget approval     | `src/app/(private)/(main)/budgets/[id]/approval/page.tsx`          |
| `/tasks`                          | Task list           | `src/app/(private)/(main)/tasks/page.tsx`                          |
| `/search`                         | Search documents    | `src/app/(private)/(main)/search/page.tsx`                         |
| `/notifications`                  | View notifications  | `src/app/(private)/(main)/notifications/page.tsx`                  |

## Component Architecture

### WorkflowBuilder (Main Container)

**File:** `src/app/(private)/admin/workflows/_components/workflow-builder.tsx`

Orchestrates the complete workflow creation/editing experience:

- Manages form state for workflow data
- Handles drag-and-drop stage ordering
- Validates all inputs
- Manages stage dialog modal
- Submits to parent component via callback

**Props:**

```typescript
interface WorkflowBuilderProps {
  onSubmit: (data: WorkflowFormData) => Promise<void>;
  isSubmitting: boolean;
  mode: "create" | "edit";
  initialData?: WorkflowFormData;
}
```

### WorkflowDetailsForm

**File:** `src/app/(private)/admin/workflows/_components/workflow-details-form.tsx`

Handles basic workflow information:

- Workflow name (required)
- Description
- Document type selection
- Default workflow toggle

### StageForm

**File:** `src/app/(private)/admin/workflows/_components/stage-form.tsx`

Modal form for adding/editing individual stages:

- Stage name and description
- Approver role selection
- Required approvals count
- Permission toggles (reject, reassign)
- Input validation with error display

### StageItem (Draggable)

**File:** `src/app/(private)/admin/workflows/_components/stage-item.tsx`

Visual representation of a single stage:

- Shows stage number, name, description
- Displays approver role and requirements
- Drag handle for reordering
- Edit and delete buttons
- Integrated with dnd-kit for drag-and-drop

### WorkflowsClient (List Page)

**File:** `src/app/(private)/admin/workflows/_components/workflows-client.tsx`

Displays all created workflows:

- Searchable/filterable table
- Duplicate workflow functionality
- Delete with confirmation dialog
- Navigate to create/edit pages
- Status badge display

## Data Types

### WorkflowFormData

```typescript
interface WorkflowFormData {
  name: string;
  description: string;
  documentType: string;
  stages: WorkflowStage[];
  isDefault: boolean;
}
```

### WorkflowStage

```typescript
interface WorkflowStage {
  id: string;
  order: number;
  name: string;
  description: string;
  approverRole: string;
  requiredApprovals: number;
  canReject: boolean;
  canReassign: boolean;
}
```

### Supported Document Types

- `REQUISITION` - Purchase requisitions
- `PURCHASE_ORDER` - Purchase orders
- `PAYMENT_VOUCHER` - Payment vouchers
- `GOODS_RECEIVED_NOTE` - Goods receipt notes
- `BUDGET` - Budget documents

### Supported Approver Roles

- `DEPARTMENT_MANAGER` - Department manager
- `FINANCE_OFFICER` - Finance officer
- `CFO` - Chief Financial Officer
- `WAREHOUSE_MANAGER` - Warehouse manager
- `PROCUREMENT_OFFICER` - Procurement officer
- `ADMIN` - Administrator

## Key Features

### 1. Drag-and-Drop Stage Ordering

- Uses `@dnd-kit` library for accessible drag-and-drop
- Stages automatically renumbered when order changes
- Visual feedback during dragging

### 2. Form Validation

- Real-time error display
- Required field validation
- Stage validation before adding
- Workflow-level validation before submit

### 3. Workflow Duplication

- Quick way to create similar workflows
- Copies all stages and configuration
- Creates new workflow with "(Copy)" suffix
- Useful for creating variations

### 4. Admin-Only Access

- Routes protected with authentication
- Admin role check on both server and client
- Non-admin users redirected to dashboard

### 5. Mock Data

- All components use mock data for demonstration
- Ready for integration with server actions
- TODO comments marking where to connect to actual API

## Integration Points

### Server Actions (To Be Implemented)

The following server actions need to be connected:

1. **Create Workflow**
   - File: `src/app/_actions/workflows.ts`
   - Function: `createWorkflow()`
   - Called from: `CreateWorkflowClient.handleSubmit()`

2. **Update Workflow**
   - File: `src/app/_actions/workflows.ts`
   - Function: `updateWorkflowAction()`
   - Called from: `EditWorkflowClient.handleSubmit()`

3. **Fetch Workflow**
   - File: `src/app/_actions/workflows.ts`
   - Function: `getWorkflowAction()`
   - Called from: `EditWorkflowClient.useEffect()`

4. **Delete Workflow**
   - File: `src/app/_actions/workflows.ts`
   - Function: `deleteWorkflow()` (needs creation)
   - Called from: `WorkflowsClient.handleDelete()`

### React Query Hooks (Already Implemented)

Available in `src/hooks/use-workflows.ts`:

- `useWorkflows()` - Fetch all workflows
- `useWorkflow()` - Fetch single workflow
- `useCreateWorkflow()` - Create mutation
- `useUpdateWorkflow()` - Update mutation

## Navigation Integration

### Sidebar Navigation

The workflow management route has been added to the sidebar navigation:

**File:** `src/components/layout/sidebar/nav-main.tsx`

**Admin Section:**

- **Title:** Workflow Management
- **Icon:** GitBranch
- **Route:** `/admin/workflows`

**Main Navigation Groups:**

1. **Main** - Dashboard, Tasks, Search, and document lists
2. **Budget Management** - Budget operations
3. **Admin** - Reports, Users, Logs, Workflows
4. **Compliance & Monitoring** - Tracking and monitoring
5. **Settings** - User preferences

### Route Structure Changes

The application uses the following route organization:

**Route Group: `(main)`**

- Encapsulates main application workflow document routes
- Keeps document workflows separate from admin/settings areas
- Routes accessible to users with appropriate permissions

**Route Group: `(private)`**

- Wrapper for all authenticated routes
- Protects all child routes with authentication middleware
- Requires active session

**Admin Routes: `/admin/*`**

- Separate admin section for configuration and management
- Admin-only access (role-based authorization)
- Includes: reports, users, logs, workflows

## Database Schema (Backend)

The backend is prepared to handle:

- Custom workflows storage
- Stage definitions
- Workflow assignments to documents
- Execution history
- Stage approvals tracking

See `src/types/custom-workflow.ts` for complete type definitions.

## Route Reference Guide

### Important Route Mapping

When navigating or creating links in the application, use these routes:

**Note:** The folder structure was reorganized to group routes logically:

- Main workflow documents moved to `(main)` route group
- Original `/workflows/*` routes map to `/(main)/*` in folder structure
- Sidebar navigation automatically handles the routing

**Example Route Mappings:**

```
OLD: /workflows/requisitions  →  NEW: /requisitions (via (main) group)
OLD: /home     →  NEW: /home (via (main) group)
OLD: /workflows/purchase-orders/[id]  →  NEW: /purchase-orders/[id]
OLD: /workflows/tasks         →  NEW: /tasks
```

The sidebar and all components automatically use the correct routes.

## User Workflow

### Creating a New Workflow

1. Navigate to: **Admin → Workflow Management**
2. Click: **Create Workflow** button
3. Fill in workflow details:
   - Name (required)
   - Description
   - Document type (required)
4. Add approval stages:
   - Click **Add Stage**
   - Configure stage details
   - Set approver role and permissions
   - Click **Add Stage** or **Update Stage**
5. Reorder stages by dragging
6. Review workflow
7. Click **Create Workflow**

### Editing a Workflow

1. Navigate to: **Admin → Workflow Management**
2. Find workflow in the list
3. Click **Edit** button
4. Modify workflow details or stages
5. Click **Update Workflow**

### Duplicating a Workflow

1. Navigate to: **Admin → Workflow Management**
2. Find workflow in the list
3. Click **Duplicate** button
4. New workflow created with "(Copy)" suffix
5. Edit as needed

## Testing

### Components to Test

1. **WorkflowsClient** - List display, actions
2. **WorkflowBuilder** - Form state, validation
3. **WorkflowDetailsForm** - Input validation
4. **StageForm** - Stage creation/editing
5. **StageItem** - Drag-and-drop functionality

### Mock Data

All components use mock data defined in respective files. Replace with real API calls when backend is ready.

## Future Enhancements

1. **Workflow Templates**
   - Pre-built templates for common workflows
   - Quick-start configurations

2. **Workflow Preview**
   - Visual preview of workflow execution
   - Show how documents flow through stages

3. **Advanced Permissions**
   - Conditional stage visibility
   - Role-based stage assignment
   - Dynamic stage routing

4. **Workflow Analytics**
   - Average approval times per stage
   - Bottleneck identification
   - Rejection rate analysis

5. **Version Control**
   - Workflow versioning
   - Rollback capability
   - Change history

## Next Steps

1. **Connect Server Actions**
   - Implement missing server actions in `src/app/_actions/workflows.ts`
   - Update component callbacks to use real data

2. **Database Integration**
   - Implement database queries for workflows
   - Set up data persistence

3. **Testing**
   - Unit tests for components
   - Integration tests for workflows
   - E2E tests for user workflows

4. **Error Handling**
   - Enhanced error messages
   - User feedback improvements
   - Retry mechanisms

5. **Documentation**
   - API documentation
   - Admin user guide
   - Developer guide for extending workflows

## File Summary

### New Files Created

- `src/app/(private)/admin/workflows/page.tsx`
- `src/app/(private)/admin/workflows/_components/workflows-client.tsx`
- `src/app/(private)/admin/workflows/_components/workflow-builder.tsx`
- `src/app/(private)/admin/workflows/_components/workflow-details-form.tsx`
- `src/app/(private)/admin/workflows/_components/stage-form.tsx`
- `src/app/(private)/admin/workflows/_components/stage-item.tsx`
- `src/app/(private)/admin/workflows/create/page.tsx`
- `src/app/(private)/admin/workflows/create/_components/create-workflow-client.tsx`
- `src/app/(private)/admin/workflows/[id]/edit/page.tsx`
- `src/app/(private)/admin/workflows/[id]/edit/_components/edit-workflow-client.tsx`

### Modified Files

- `src/components/layout/sidebar/nav-main.tsx` - Added Workflow Management menu item

### Existing Files (Already Implemented)

- `src/types/custom-workflow.ts`
- `src/app/_actions/workflows.ts`
- `src/hooks/use-workflows.ts`
- `src/lib/approval-config.ts`

## Support and Questions

For questions about the workflow system, see:

- Type definitions: `src/types/custom-workflow.ts`
- Server actions: `src/app/_actions/workflows.ts`
- React hooks: `src/hooks/use-workflows.ts`
- Approval configuration: `src/lib/approval-config.ts`
