# Workflow System - Complete Guide

**Status**: Production Ready
**Last Updated**: December 5, 2025
**Scope**: Workflow management, custom workflows, workflow builder, and implementation

---

## Table of Contents

1. [Overview](#overview)
2. [Core Workflows](#core-workflows)
3. [Workflow Builder](#workflow-builder)
4. [Custom Workflows](#custom-workflows)
5. [Implementation Details](#implementation-details)
6. [Admin Management](#admin-management)
7. [Architecture](#architecture)

---

## Overview

The Liyali Gateway implements a flexible workflow system that processes 5 document types through multi-stage approval workflows. Administrators can create custom approval workflows, and the system automatically routes documents through the configured approval chain.

### Key Capabilities
- ✅ **5 Predefined Workflows**: Requisition, Budget, Purchase Order, Payment Voucher, GRN
- ✅ **Custom Workflow Creation**: Drag-and-drop builder for custom workflows
- ✅ **Multi-stage Approvals**: 2-3 stage workflows with configurable roles
- ✅ **Digital Signatures**: Approval capture with signature and timestamp
- ✅ **Bulk Operations**: Approve/Reject/Reassign multiple items at once
- ✅ **Complete Audit Trail**: All actions logged with timestamps and user info

---

## Core Workflows

### 1. Requisition Workflow (2 Stages)

**Document Type**: REQUISITION

**Stages**:
1. **Department Manager Review** (DEPARTMENT_MANAGER)
   - Can approve, reject, reassign
   - Reviews total cost and items

2. **Finance Approval** (FINANCE_OFFICER)
   - Can approve or reject
   - Final approval before PO creation

**Route**: `/requisitions`
**PDF Export**: ✅ Government-compliant PDF with QR code

### 2. Budget Workflow (3 Stages)

**Document Type**: BUDGET

**Stages**:
1. **Department Head** (DEPARTMENT_HEAD)
   - Reviews and approves budget

2. **Finance Manager** (FINANCE_MANAGER)
   - Allocates budget codes

3. **CFO Review** (CFO)
   - Final approval

**Route**: `/budgets`
**PDF Export**: ✅ Available

### 3. Purchase Order Workflow (2 Stages)

**Document Type**: PURCHASE_ORDER

**Stages**:
1. **Procurement Officer** (PROCUREMENT_OFFICER)
   - Verifies vendor and pricing

2. **Approving Officer** (APPROVING_OFFICER)
   - Final approval before sending to vendor

**Route**: `/purchase-orders`
**Features**: Vendor tracking, GL codes
**PDF Export**: ✅ Government-compliant PDF with signatures

### 4. Payment Voucher Workflow (3 Stages)

**Document Type**: PAYMENT_VOUCHER

**Stages**:
1. **Approving Officer** (APPROVING_OFFICER)
   - Verifies invoice against PO

2. **Finance Officer** (FINANCE_OFFICER)
   - Checks budget allocation and GL codes

3. **Bank Officer** (BANK_OFFICER)
   - Processes payment

**Route**: `/payment-vouchers`
**Features**: Payment method, bank details, invoice tracking
**PDF Export**: ✅ Government-compliant PDF with watermarks

### 5. GRN Workflow (2 Stages)

**Document Type**: GRN (Goods Received Note)

**Stages**:
1. **Warehouse Officer** (WAREHOUSE_OFFICER)
   - Records items received
   - Checks against PO

2. **Quality Officer** (QUALITY_OFFICER)
   - Confirms quality and item counts
   - Logs any variances

**Route**: `/grn`
**Features**: Item matching, variance tracking, quality issues
**PDF Export**: ✅ Available

---

## Workflow Builder

### What Is It?

A **visual, drag-and-drop interface** for administrators to create custom multi-stage approval workflows. Located at `/admin/workflows`.

### How to Create a Workflow

#### Step 1: Navigate to Workflow Creation
- Go to `/admin/workflows`
- Click "Create New Workflow" button

#### Step 2: Fill Workflow Details
```
Workflow Name:        "Standard Purchase Order Approval"
Description:          "2-stage PO approval with vendor verification"
Document Type:        PURCHASE_ORDER (dropdown)
Is Default:          ☐ (Check to use as default for this document type)
```

#### Step 3: Add Approval Stages
- Click "Add Stage" button
- Configure stage in modal dialog:

```
Stage Name:           "Procurement Review"
Approver Role:        PROCUREMENT_OFFICER (dropdown)
Required Approvals:   1
Permissions:
  ☑ Can Reject:       Yes
  ☑ Can Reassign:     Yes
```

#### Step 4: Manage Stages
- **Reorder**: Drag stages by handle to reorder
- **Edit**: Click edit icon to modify stage
- **Delete**: Click delete icon to remove stage
- **Maximum**: Up to 5 stages per workflow

#### Step 5: Submit
- Click "Create Workflow"
- Validation runs
- Redirects to workflow list on success

### Workflow Builder Components

| Component | Purpose | Key Features |
|-----------|---------|--------------|
| **WorkflowBuilder** | Orchestrator component | State management, validation, submission |
| **WorkflowDetailsForm** | Top-level form | Name, description, document type |
| **StageForm** | Stage editor modal | Stage configuration, validation |
| **StageItem** | Draggable stage card | Drag handle, edit/delete actions |
| **WorkflowsClient** | List view | Display, filter, edit, delete workflows |

### Technical Details

**State Management**:
```typescript
const [formData, setFormData] = useState<WorkflowFormData>()
const [showStageDialog, setShowStageDialog] = useState(false)
const [editingStageId, setEditingStageId] = useState(null)
const [stageErrors, setStageErrors] = useState({})
const [formErrors, setFormErrors] = useState({})
```

**Drag-and-Drop Library**: `@dnd-kit/core` with sortable extension

**Validation**:
- Workflow name required
- At least 1 stage required
- Unique stage names
- Valid approver roles
- All stages must have required approvals > 0

---

## Custom Workflows

### Creating Custom Workflows

1. **Start** at `/admin/workflows/create`
2. **Define** workflow name and document type
3. **Build** approval stages with drag-and-drop
4. **Configure** each stage with:
   - Approver role
   - Number of approvals needed
   - Permissions (reject, reassign)
5. **Submit** and start using

### Using Custom Workflows

Once created, custom workflows automatically apply to:
- Document creation forms (if marked as default)
- Approval routing logic
- PDF generation (uses actual approval chain)

### Examples

#### Example 1: Simple 2-Stage Requisition
```
Stage 1: Manager Review (DEPARTMENT_MANAGER)
  - 1 approval required
  - Can reject and reassign

Stage 2: Finance Approval (FINANCE_OFFICER)
  - 1 approval required
  - Can reject only
```

#### Example 2: Complex 4-Stage PO
```
Stage 1: Procurement (PROCUREMENT_OFFICER)
  - 1 approval required

Stage 2: Budget Check (FINANCE_MANAGER)
  - 1 approval required

Stage 3: Director Review (DIRECTOR)
  - 1 approval required

Stage 4: Final Approval (CFO)
  - 1 approval required
  - Can only approve (no reject)
```

---

## Implementation Details

### Folder Structure

```
src/
├── app/(private)/
│   ├── admin/
│   │   └── workflows/                    # Workflow Management Admin
│   │       ├── page.tsx                  # List workflows
│   │       ├── create/
│   │       │   └── _components/
│   │       │       └── create-workflow-client.tsx
│   │       └── [id]/edit/
│   │           └── _components/
│   │               └── edit-workflow-client.tsx
│   │
│   └── (main)/
│       ├── requisitions/                 # Requisition execution
│       ├── purchase-orders/              # PO execution
│       ├── payment-vouchers/             # PV execution
│       ├── budgets/                      # Budget execution
│       └── grn/                          # GRN execution
│
├── components/
│   ├── workflows/
│   │   ├── approval-action-panel.tsx     # Approval modal
│   │   ├── approval-flow-display.tsx     # Approval chain display
│   │   └── ...other workflow components
│   └── ...other components
│
├── types/
│   ├── workflow.ts                       # Workflow types
│   └── custom-workflow.ts                # Custom workflow types
│
├── app/_actions/
│   ├── workflows.ts                      # Custom workflow actions
│   ├── requisitions.ts                   # Requisition actions
│   ├── purchase-orders.ts                # PO actions
│   ├── payment-vouchers.ts               # PV actions
│   ├── budgets.ts                        # Budget actions
│   └── grn.ts                            # GRN actions
│
└── hooks/
    ├── use-workflows.ts                  # Custom workflow hooks
    ├── use-requisition-queries.ts        # Requisition queries
    └── ...other workflow hooks
```

### Data Models

#### WorkflowFormData
```typescript
interface WorkflowFormData {
  name: string                            // "Standard Requisition Approval"
  description: string                     // "2-stage approval process"
  documentType: 'REQUISITION' | 'BUDGET' | 'PURCHASE_ORDER' | 'PAYMENT_VOUCHER' | 'GRN'
  stages: WorkflowStage[]
  isDefault: boolean                      // Use as default for this doc type
}

interface WorkflowStage {
  id: string                              // "stage-1733328400000"
  order: number                           // 1, 2, 3...
  name: string                            // "Department Manager Review"
  approverRole: string                    // "DEPARTMENT_MANAGER"
  requiredApprovals: number               // 1, 2, ...
  canReject: boolean
  canReassign: boolean
}
```

#### ApprovalChain
```typescript
interface ApprovalChainItem {
  stageNumber: number
  stageName: string
  approverRole: string
  status: 'PENDING' | 'APPROVED' | 'REJECTED'
  assignedTo: string                      // User ID
  actionTakenBy?: string                  // User ID
  actionTakenAt?: string                  // ISO timestamp
  signature?: string                      // Base64 signature
  comments?: string
}
```

### Server Actions

Key workflow-related server actions:

```typescript
// Workflow Management
createCustomWorkflow(data: WorkflowFormData)
updateCustomWorkflow(id: string, data: WorkflowFormData)
deleteCustomWorkflow(id: string)
getCustomWorkflows()
getCustomWorkflow(id: string)

// Approval Operations
submitDocumentForApproval(documentId: string, documentType: string)
approveDocument(documentId: string, stageNumber: number, signature: string)
rejectDocument(documentId: string, stageNumber: number, reason: string)
reassignApproval(documentId: string, stageNumber: number, newApprover: string)

// Bulk Operations
bulkApprove(documentIds: string[], signature: string)
bulkReject(documentIds: string[], reason: string)
bulkReassign(documentIds: string[], newApprover: string)
```

---

## Admin Management

### Access

- **Route**: `/admin/workflows`
- **Required Role**: ADMIN
- **Permissions**: Create, Read, Update, Delete workflows

### List View Features

| Feature | Description |
|---------|-------------|
| **Sort** | By name, document type, status, last updated |
| **Filter** | By status (ACTIVE, DEPRECATED) |
| **Actions** | Edit, Duplicate, Delete, Set as Default |
| **View** | Workflow name, document type, # stages, status |

### Permissions

Only ADMIN users can:
- Create workflows
- Edit workflows
- Delete workflows
- Set default workflows

All users can:
- View assigned tasks
- Approve/reject according to workflow stages
- Request reassignment

---

## Architecture

### Data Flow

```
User submits document
    ↓
System fetches workflow (custom or default)
    ↓
Creates approval chain from stages
    ↓
Routes to first stage approver
    ↓
Approver reviews and signs
    ↓
System checks if all required approvals obtained
    ↓
Routes to next stage OR completes
    ↓
Document marked as APPROVED/REJECTED
```

### State Persistence

- **Storage**: localStorage (Phase 11)
- **Planned**: PostgreSQL (Phase 13)
- **Cache**: React Query with auto-refresh

### Audit Trail

Every action is logged:
- Document creation
- Stage approvals
- Rejections with reason
- Reassignments
- Signature captures
- Timestamps and user info

All accessible via:
- `/admin/logs` - Activity logs
- Document detail pages - Action history
- Audit export - Full transaction history

---

## Integration with PDF System

The workflow system integrates with PDF exports:

### Dynamic Approval Signatures

PDFs automatically include:
- Current approval chain
- Actual approver names and roles
- Approval status and dates
- Signature placeholders

### Status Watermarks

PDFs include status-based watermarks:
- DRAFT (Red)
- SUBMITTED (Orange)
- IN_REVIEW (Yellow)
- APPROVED (Green)
- PAID (Blue)
- REJECTED (Pink)

### QR Code Integration

PDFs embed QR codes with:
- Document type and number
- Document ID
- Timestamp
- Checksum for verification

---

## Best Practices

### When Creating Workflows

1. **Keep stages logical**: Order stages in business process sequence
2. **Clear permissions**: Decide who can reject vs. approve
3. **Clear names**: Use descriptive stage names like "Manager Review" not "Stage 1"
4. **Test before deploying**: Create in dev and test approval flow
5. **Document purpose**: Add description for other admins

### When Executing Workflows

1. **Complete details**: Fill all required fields before submitting
2. **Review before approval**: Check all details before approving
3. **Provide feedback**: Add comments when rejecting
4. **Timely approval**: Monitor pending approvals daily

---

## Troubleshooting

### Issue: Custom workflow not being used

**Solution**: Check if workflow is marked as "default" for that document type.

### Issue: Stage not appearing in approval chain

**Solution**: Verify stage order is correct and no stages were deleted.

### Issue: Approval stuck in pending

**Solution**: Check if approver role matches configured stage role. Reassign if needed.

---

## See Also

- [APPROVAL-GUIDE.md](APPROVAL-GUIDE.md) - How to approve workflows
- [WORKFLOW_IMPLEMENTATION_PLAN.md](WORKFLOW_IMPLEMENTATION_PLAN.md) - Technical implementation details
- [PDF_ENHANCEMENTS_SUMMARY.md](PDF_ENHANCEMENTS_SUMMARY.md) - PDF generation with workflows
- [REQUISITION_TO_PO_INTEGRATION.md](REQUISITION_TO_PO_INTEGRATION.md) - Multi-document workflow flow

---

**Version**: 1.0 | **Status**: Production Ready | **Maintained**: Yes
