# Requisitions Management System

## Overview

A complete requisitions management system has been built with all the features you requested. This allows users to create, manage, and track equipment/supply requisition requests through a multi-stage approval workflow.

## Features Implemented

### 1. Requisitions List Page (`/workflows/requisitions`)
- **Table View**: Display all requisitions created by the current user
- **Columns**:
  - Document Number (auto-generated)
  - Requested For
  - Department
  - Total Items
  - Status (DRAFT, SUBMITTED, IN_APPROVAL, APPROVED, REJECTED)
  - Created Date
  - Actions (View Details)

- **Status Badges**: Color-coded for quick identification
  - DRAFT: Gray
  - SUBMITTED: Blue
  - IN_APPROVAL: Yellow
  - APPROVED: Green
  - REJECTED: Red

### 2. Create Requisition Dialog
Accessible via "Create Requisition" button on the list page.

**Features**:
- Modal form with scrollable content
- **Basic Information Section**:
  - Department (required)
  - Requested For (required)
  - Budget Code (required)
  - Justification (required, textarea)

- **Items Management**:
  - Add multiple items dynamically
  - For each item:
    - Description (required)
    - Quantity (required)
    - Estimated Unit Cost (required)
    - Total Cost (auto-calculated)
  - Remove items with delete button
  - Visual feedback with dashed border when empty

- **Summary**:
  - Total Estimated Cost (displayed at bottom in blue)
  - Displays total in ZMW currency format

- **Form Validation**:
  - All required fields validated before submission
  - Items must have complete details
  - Error toasts for validation failures

### 3. Requisition Detail Page (`/workflows/requisitions/[id]`)

#### Header Section
- Back button for navigation
- Document number and status badge
- Creation timestamp
- Submit for Approval button (only for creator in DRAFT/REJECTED status)

#### Main Content Area (Left Side)

**Requisition Details Card**:
- Department
- Requested For
- Justification
- Budget Code
- Current Approval Stage

**Requisition Items Card**:
- List view of all items
- For each item:
  - Item number
  - Item description
  - Quantity
  - Unit cost
  - Total cost (highlighted in blue)
- Grand total at bottom

**Edit Requisition Panel** (only for creator in DRAFT/REJECTED status):
- "Edit Requisition" button to toggle edit mode
- When editing:
  - All fields become editable
  - Items can be added/removed/modified
  - Save button persists changes
  - Cancel reverts to view mode

#### Sidebar (Right Side)

**Approval History Panel**:
Tabbed interface with two tabs:

1. **Approval Log Tab**:
   - Timeline of all approval actions
   - For each log entry:
     - Approver name
     - Action (APPROVED, REJECTED, COMMENTED, REASSIGNED)
     - Timestamp
     - Comments (if provided)
   - Color-coded backgrounds:
     - Green for approvals
     - Red for rejections
     - Gray for other actions
   - Scrollable with max height

2. **Approvers Tab**:
   - List of all assigned approvers
   - For each approver:
     - Name
     - Approval Stage
     - Status (PENDING, APPROVED, REJECTED, SKIPPED)
   - Status badge with appropriate styling

3. **Approval Action Panel** (only when status is IN_APPROVAL):
   - "Approve" button (green)
   - "Reject" button (red)
   - When action selected:
     - Comments field (optional for approve, required for reject)
     - "Add Supporting Documents" button (opens dialog)
     - Confirmation buttons
     - Cancel to go back

## Workflow Status & Behavior

### Status Transitions

```
DRAFT
  ↓ (Click "Submit for Approval")
SUBMITTED → IN_APPROVAL
             ↓
      (Approvers act)
             ↓
      Can Approve → APPROVED (Final)
      Can Reject → REJECTED
                      ↓
                   (Back to DRAFT for editing)
```

### Editing Rules

**Can Edit**:
- Creator only
- Status must be DRAFT or REJECTED
- All fields editable
- Items fully modifiable

**Cannot Edit**:
- Non-creators cannot edit
- Once submitted and in approval workflow
- After approval or other final statuses

### Approver Actions

**During IN_APPROVAL Status**:
- Approvers can view all document details
- Can add comments
- Can upload supporting documents
- Can approve (moves to next stage or final approval)
- Can reject (returns to DRAFT for creator to revise)
- Actions logged immutably with timestamp and approver details

## Component Structure

```
requisitions/
├── page.tsx                                 # Main list page
├── [id]/page.tsx                           # Detail page
└── _components/
    ├── index.ts                            # Exports
    ├── requisitions-client.tsx             # List page client logic
    ├── requisitions-table.tsx              # Table display
    ├── create-requisition-dialog.tsx       # Create form dialog
    ├── requisition-detail-client.tsx       # Detail page client logic
    ├── approval-history-panel.tsx          # Approval history & approvers tabs
    ├── approval-action-panel.tsx           # Approve/reject actions
    └── edit-requisition-panel.tsx          # Edit requisition form
```

## Data Model Integration

All components use the mocked server actions:

**Create/Update**:
- `createWorkflowDocument('REQUISITION', formData)`
- `updateDocumentDraft(documentId, formData)`
- `submitDocument(documentId)`

**Approval**:
- `approveDocument(documentId, comments)`
- `rejectDocument(documentId, reason)`

**Retrieval**:
- `getDocumentsByCreator(userId)`
- `getDocument(documentId)`
- `getApprovalLog(documentId)`
- `getDocumentApprovers(documentId)`

## Key Features

✅ **Complete Workflow Management**:
- Draft creation and editing
- Submission for approval
- Multi-stage approval tracking
- Status-based visibility controls

✅ **User-Friendly UI**:
- Clean, organized layout
- Color-coded status badges
- Responsive design
- Tabbed interfaces for organization

✅ **Data Validation**:
- Required field validation
- Item details validation
- User feedback via toast notifications

✅ **Audit Trail**:
- Immutable approval logs
- Timestamps for all actions
- Approver tracking
- Comments and attachments support

✅ **Role-Based Access**:
- Only creator can edit/submit
- Approvers can only act during IN_APPROVAL
- Full RBAC integration ready

✅ **Approval Controls**:
- Attach supporting documents
- Add comments/remarks
- Approve or reject with reasons
- Track all changes in audit log

## Usage Flow

### For Requesters

1. **Create Requisition**:
   - Click "Create Requisition" on list page
   - Fill in basic info (department, requested for, justification, budget code)
   - Add items with description, quantity, and estimated cost
   - Submit form to create draft

2. **Edit Draft**:
   - Click "View Details" on requisition
   - Click "Edit Requisition" button
   - Modify any fields or items
   - Click "Save Changes"

3. **Submit for Approval**:
   - Click "Submit for Approval" button on detail page
   - Document moves to IN_APPROVAL status
   - Approvers are auto-assigned based on workflow

4. **Monitor Progress**:
   - Check Approval Log tab to see who approved/rejected
   - View Approvers tab for current stage

### For Approvers

1. **View Pending Approvals**:
   - Navigate to requisitions page (filtered for IN_APPROVAL status)
   - Click "View Details" on a requisition needing approval

2. **Review & Act**:
   - Review all details and items
   - Click "Approve" or "Reject"
   - Add comments/remarks
   - (Optional) Upload supporting documents
   - Confirm action

3. **Track Decisions**:
   - Approval automatically logged
   - Remarks captured in audit trail
   - Next approver auto-assigned if not final stage

## Future Enhancements

- File upload for attachments
- Email notifications to approvers
- Bulk approval actions
- Advanced filtering and search
- Export to PDF
- Approval history reports
- Dashboard analytics

## Testing

All components work with mocked server actions. No database required.

Test the complete flow:
1. Create a requisition
2. Submit for approval
3. Approve/reject as different approvers
4. Check audit trail
5. Edit and resubmit if rejected

## Notes

- All data uses mocked in-memory storage
- Dates are in local browser timezone
- Currency is ZMW with 2 decimal places
- Colors follow consistent design patterns
- Fully responsive and mobile-friendly
