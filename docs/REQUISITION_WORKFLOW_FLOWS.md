# Requisition Workflow Flows - Analysis & Implementation Map

## Overview

The requisition system follows a comprehensive multi-stage approval workflow with different roles and decision points. This document maps the visual flows to the system implementation.

---

## Flow 1: User Login & Dashboard Access

```
User Login
  → Authentication (Password + MFA)
    → Access Granted
      → Role-Based Dashboard Loaded
        → Create Budget Memo / Requisition
```

### Current Implementation Status
- ✅ **User Login**: Handled by NextAuth authentication
- ✅ **MFA**: Supported by NextAuth configuration
- ✅ **Access Granted**: Session validation in server pages
- ✅ **Role-Based Dashboard**: Permissions system in place
- ✅ **Create Button**: Present on requisitions list page

### Related Code
- Authentication: `src/auth.ts` + `src/auth.config.ts`
- Dashboard redirect: `src/app/workflows/requisitions/page.tsx`
- Role-based access: RBAC system in `src/lib/rbac.ts`

---

## Flow 2: Requisition Creation & Budget Memo Process

```
Create Budget Memo
  → Enter User Details, Justification, Budget Line, Attach Docs
    → Submit Memo for Requisition Creation
      → [If Rejected] → Back to Budget Memo entry (loop)
      → [If Approved] → Create Requisition
        → Add Items, Specs, Cost, Justification
          → Submit Requisition to Head of Department
```

### Current Implementation Status

#### Budget Memo Section
- ⚠️ **Status**: Partially implemented
- **What's Done**:
  - Create requisition dialog exists
  - Form captures: department, requestedFor, budgetCode, justification
  - Item collection with add/remove functionality

- **What's Missing**:
  - Budget memo as separate entity before full requisition
  - Memo submission/approval workflow
  - Memo rejection feedback loop
  - Docs/attachments at memo stage

#### Requisition Creation
- ✅ **Implemented**:
  - Dialog form with fields for department, requestedFor, justification
  - Dynamic item addition (description, quantity, cost)
  - Auto-calculated totals
  - Submit for approval functionality

- **Enhancement Needed**:
  - More detailed item specifications (specs field)
  - Better justification capture
  - Pre-memo workflow option

### Related Code
- Dialog: `src/app/workflows/requisitions/_components/create-requisition-dialog.tsx`
- Server action: `createWorkflowDocument()` in `src/app/_actions/workflow.ts`

### Recommended Enhancement
Consider creating a "budget memo" as initial draft stage, then promoting to full requisition after memo approval.

---

## Flow 3: Multi-Stage Approval Workflow

### Stage 1: Head of Department Review

```
Submit Requisition to Head of Department
  → Head of Department Review
    → [Approved] → Continue to Stage 2
    → [Rejected] → Return to requisition creation (loop)
```

**Current Implementation**:
- ✅ Status: SUBMITTED → IN_APPROVAL
- ✅ Auto-assignment of first approver (Head of Department)
- ✅ Approval action panel exists
- ✅ Rejection returns to DRAFT/REJECTED
- ✅ Immutable audit trail logs all actions

### Stage 2: Principal Officer Review

```
Principal Officer Reviews & Approves Memo/Requisition
  → [Approved] → Continue to Stage 3
  → [Rejected] → Return to original requester (with feedback)
```

**Current Implementation**:
- ✅ Auto-progression to next stage on approval
- ✅ Comment/remark capability during approval
- ✅ Rejection handling with reason capture

### Stage 3: Director Finance Review

```
Director Finance Reviews & Approves Memo/Requisition for Financing
  → [Approved] → Forward to Procurement Officer
  → [Rejected] → Return to requester
```

**Current Implementation**:
- ✅ Multi-stage progression system ready
- ✅ Role-based approver assignment
- ✅ Forwarding/routing capability

### Stage 4: Procurement Officer Review

```
Procurement Officer Reviews
  → Add Supplier Info & Upload Compliance Docs
    → Attach Evaluation Report & Quotations
      → Choose Delivery Type
        → Purchase Order Approval
          → 2. Principal Officer Approval (final PO approval)
```

**Current Implementation**:
- ⚠️ **Status**: Partially implemented
- **What's Done**:
  - Attachment upload functionality exists
  - Comments/remarks during approval
  - Role-based access control

- **What's Missing**:
  - Procurement-specific fields (supplier info, compliance docs)
  - Evaluation report attachment workflow
  - Delivery type selection
  - Transition from requisition to Purchase Order creation
  - Final PO approval stage

### Related Code
- Approval action: `src/app/workflows/requisitions/_components/approval-action-panel.tsx`
- Server actions: `approveDocument()`, `rejectDocument()`, `uploadAttachment()` in `src/app/_actions/workflow.ts`
- Audit trail: Handled by immutable ApprovalLogEntry

---

## Flow 4: Payment Voucher Workflow

```
After PO Approval
  → Notify Procurement Officer: PO Approved
    → [Before Payment] → Stores Officer Receives Items
                          → Create GRN, Add Delivery & Inspection Notes, Sign
                          → Notify Accountant: Ready for Voucher
    → [After Payment] → Accountant Proceeds to Voucher Generation
                         → Review Docs, Validate Bank Info, Select Vote Code
                           → Generate Payment Voucher with QR & Reference Number
                             → Payment Voucher Approval (3-stage):
                                 1. Director Finance Approval
                                 2. Accountant Approval
                                 3. Principal Officer Final Approval
                                   → Notify Stakeholders: Payment Approved
                                     → System Logs Actions, Updates Dashboards
                                       → End of Workflow
```

**Current Implementation Status**:
- ⚠️ **Status**: Framework ready, workflow-specific features needed
- **What's Done**:
  - Basic payment voucher document type exists
  - Multi-stage approval system
  - Attachment/document upload
  - Approval workflow with comments
  - Audit trail logging

- **What's Missing**:
  - GRN (Goods Received Note) creation
  - Delivery inspection tracking
  - Bank info validation
  - Vote code selection
  - QR code generation
  - Payment voucher-specific fields
  - Post-approval notifications
  - Dashboard updates after completion

### Related Code
- Document creation: `createWorkflowDocument()` supports PAYMENT_VOUCHER type
- Server actions: All approval workflow actions apply
- Status handling: Document state management ready

---

## Implementation Status Summary

### ✅ Complete - Core Foundation
1. User authentication with role-based access
2. Requisition creation with items
3. Multi-stage approval workflow
4. Approval/rejection actions with comments
5. Attachment upload capability
6. Immutable audit trail
7. Auto-stage progression
8. Status tracking (DRAFT, SUBMITTED, IN_APPROVAL, APPROVED, REJECTED)

### ⚠️ Partial - Workflow-Specific Features
1. **Budget Memo Stage**: Create as separate approval stage before full requisition
2. **Procurement Officer Stage**: Add supplier/compliance doc handling
3. **Purchase Order Creation**: From requisition after final approval
4. **GRN Management**: Goods received note creation and tracking
5. **Payment Voucher Details**: Bank info, vote code, QR generation
6. **Notifications**: System should notify users at key stages

### ❌ Not Yet Implemented - Additional Features
1. **Delivery Type Selection**: Choose delivery method during procurement
2. **Evaluation Reports**: Store and manage supplier evaluation docs
3. **Quotations**: Collect and compare supplier quotations
4. **QR Code Generation**: For payment vouchers
5. **Dashboard Updates**: Reflect workflow completion
6. **Stakeholder Notifications**: Email/in-app notifications at stages

---

## Approval Roles Mapping

Based on the flows, the system requires:

| Stage | Role | Document Type | Action |
|-------|------|---------------|--------|
| 1 | Head of Department | Requisition | Approve/Reject |
| 2 | Principal Officer | Requisition/Memo | Approve/Reject |
| 3 | Director Finance | Requisition | Approve/Reject |
| 4 | Procurement Officer | Requisition | Add Docs, Approve |
| 5 | Principal Officer | Purchase Order | Final Approve |
| 6 | Director Finance | Payment Voucher | Approve/Reject |
| 7 | Accountant | Payment Voucher | Approve/Reject |
| 8 | Principal Officer | Payment Voucher | Final Approve |

### Current System Status
- ✅ Role-based assignment system ready
- ✅ Custom role creation supported
- ⚠️ Need to verify all roles defined in mock data
- ⚠️ May need to add "Accountant" role to default roles

---

## Document Type Workflows

### Requisition Form Workflow
```
DRAFT → SUBMITTED → IN_APPROVAL (4 stages) → APPROVED
                 ↓
              REJECTED → DRAFT (back to creator)
```

**Stages in IN_APPROVAL**:
1. Head of Department
2. Principal Officer
3. Director Finance
4. Procurement Officer (add docs, evaluate)

**On Final Approval**:
- Create Purchase Order document
- Transition to PO workflow

### Purchase Order Workflow
```
Created from Requisition → IN_APPROVAL → APPROVED
                        ↓
                      REJECTED
```

**Stage**:
1. Principal Officer (final approval)

**On Approval**:
- Notify Procurement Officer
- Proceed to delivery/GRN phase

### Payment Voucher Workflow
```
Created after GRN/Goods Receipt → IN_APPROVAL (3 stages) → APPROVED
                                ↓
                              REJECTED
```

**Stages in IN_APPROVAL**:
1. Director Finance
2. Accountant
3. Principal Officer (final)

**On Approval**:
- Send payment
- Notify stakeholders
- Update dashboards
- End workflow

---

## Key UI Features Needed

### 1. Requisition Detail Page
Currently implemented but needs:
- [ ] Display budget memo reference (if exists)
- [ ] Show current approval stage indicator
- [ ] Display next approver info
- [ ] Multi-stage progress indicator
- [ ] Supplier info section (for procurement stage)
- [ ] Evaluation docs section

### 2. Approval Action Panel
Currently implemented but needs:
- [ ] Stage-specific fields based on role
- [ ] For Procurement Officer: supplier info form, compliance docs
- [ ] For Finance roles: vote code selection, bank info validation
- [ ] Better comment/remark capture
- [ ] Document upload with purpose selection

### 3. Status Indicators
Currently implemented but needs:
- [ ] Current stage number display (Stage 1/4, etc.)
- [ ] Approver timeline/history
- [ ] Next approver name and role
- [ ] Expected completion date
- [ ] SLA indicator (if applicable)

### 4. Dashboard/Notifications
Not yet implemented:
- [ ] Pending approvals by role
- [ ] Workflow completion stats
- [ ] In-progress documents by stage
- [ ] Email/in-app notifications

---

## Next Steps (Priority Order)

### Phase 1: Improve Requisition Workflow (High Priority)
1. Add budget memo as separate document type/approval stage
2. Add "Accountant" role to default roles
3. Enhance procurement officer stage with supplier/evaluation fields
4. Add current stage indicator to detail page
5. Create Purchase Order automatically on final approval

### Phase 2: Payment Voucher Workflow (Medium Priority)
1. Implement GRN document creation
2. Add payment voucher specific fields
3. Implement 3-stage payment voucher approval
4. Add vote code and bank info handling
5. Create QR code generation for vouchers

### Phase 3: Notifications & Dashboard (Medium Priority)
1. Add notification system (email/in-app)
2. Create dashboard with pending items
3. Add workflow completion tracking
4. Display metrics and statistics

### Phase 4: Polish & Enhancement (Low Priority)
1. Add delivery type selection
2. Implement evaluation report management
3. Add quotation comparison
4. Add SLA tracking
5. Add bulk operations for admins

---

## Data Model Extensions

### Budget Memo (New)
```typescript
interface BudgetMemo {
  id: string
  requisitionId?: string  // Links to requisition after approval
  creatorId: string
  department: string
  budgetLine: string
  justification: string
  attachments: Attachment[]
  status: 'DRAFT' | 'SUBMITTED' | 'APPROVED' | 'REJECTED'
  approvalLog: ApprovalLogEntry[]
  createdAt: Date
  updatedAt: Date
}
```

### Requisition Extension
```typescript
interface RequisitionForm {
  // ... existing fields
  budgetMemoId?: string    // Links to approved memo
  procurementNotes?: string
  supplierInfo?: {
    name: string
    contact: string
    compliance: Attachment[]
  }
  evaluationReport?: Attachment
  deliveryType?: 'Standard' | 'Express' | 'Pick-up'
  currentStage: number  // 1-4
  nextApprover?: User
}
```

### Payment Voucher Extension
```typescript
interface PaymentVoucher {
  // ... existing fields
  purchaseOrderId: string
  grnId: string
  bankInfo: {
    accountNumber: string
    bankCode: string
    accountName: string
  }
  voteCode: string
  qrCode?: string
  paymentReference?: string
  currentStage: number  // 1-3
}
```

---

## Testing Workflow Scenarios

### Scenario 1: Happy Path (Requisition → PO → Payment)
1. Create requisition
2. Submit for approval
3. Each approver approves → auto-progresses
4. Final approval creates PO
5. PO approval triggers payment workflow
6. GRN → Payment Voucher
7. Payment voucher approvals
8. Final notification

### Scenario 2: Rejection at Each Stage
1. Requisition rejected at stage 2
2. Sent back to creator as REJECTED
3. Creator edits and resubmits
4. Process continues

### Scenario 3: Procurement-Specific Actions
1. Requisition reaches procurement officer
2. Officer adds supplier info, docs
3. Officer adds evaluation report
4. Approves with stage-specific data
5. Continues to next stage

---

## Flow Alignment Checklist

- [x] User login and authentication
- [x] Role-based dashboard access
- [x] Create requisition form
- [ ] Budget memo as separate stage
- [x] Submit requisition to approvers
- [x] Multi-stage approval workflow
- [x] Approval/rejection with comments
- [x] Add attachments during approval
- [ ] Procurement-specific fields
- [ ] Auto-create Purchase Order on final approval
- [ ] Purchase Order approval stage
- [ ] GRN creation and tracking
- [ ] Payment Voucher creation
- [ ] Payment Voucher 3-stage approval
- [ ] Stakeholder notifications
- [ ] System logs and dashboard updates

---

## Notes for Development

1. **Budget Memo**: Consider making this a workflow option (memo-only vs direct requisition)
2. **Role Assignments**: Verify all roles are properly assigned in mock data
3. **Stage Progression**: Ensure correct approver assignment based on role/document type
4. **Notifications**: Plan notification system early (email, in-app, dashboard)
5. **Audit Trail**: Ensure all stage transitions are logged immutably
6. **Rejection Flow**: Ensure rejected documents clearly show reason and return to appropriate creator

---

**Last Updated**: 2024-11-29
**Flow Diagrams Reference**: Provided flow images show complete requisition, PO, and payment voucher workflows
**Status**: Analysis complete, implementation priorities identified
