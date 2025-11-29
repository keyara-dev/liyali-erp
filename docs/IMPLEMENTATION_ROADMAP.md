# Implementation Roadmap - Workflow System

## Current Status
✅ **Completed**: Core infrastructure, requisition creation & list, multi-stage approval, attachment uploads, immutable audit trails

## Phase 1: Enhance Requisition Workflow (IMMEDIATE - HIGH PRIORITY)

### Goal
Map requisition flow exactly to the business process shown in flows

### Tasks

#### 1.1 Add Stage Indicators to Detail Page
**File**: `src/app/workflows/requisitions/_components/requisition-detail-client.tsx`

- [ ] Display "Stage X of 4" indicator
- [ ] Show stage name (Department Head, Principal Officer, Director Finance, Procurement)
- [ ] Display current approver information
- [ ] Show next approver info
- [ ] Color-coded stage progression indicator

**Implementation**:
```typescript
interface ApprovalStage {
  stageNumber: number
  stageName: string
  assignedTo: User
  status: 'PENDING' | 'APPROVED' | 'REJECTED'
  completedAt?: Date
}

// Display stage timeline
const stages: ApprovalStage[] = [
  { stageNumber: 1, stageName: 'Head of Department', ... },
  { stageNumber: 2, stageName: 'Principal Officer', ... },
  { stageNumber: 3, stageName: 'Director Finance', ... },
  { stageNumber: 4, stageName: 'Procurement Officer', ... },
]
```

#### 1.2 Enhance Procurement Officer Stage
**File**: `src/app/workflows/requisitions/_components/approval-action-panel.tsx`

- [ ] Detect when current stage is "Procurement Officer"
- [ ] Show form for supplier information:
  - [ ] Supplier name
  - [ ] Supplier contact
  - [ ] Supplier code/ID
- [ ] Allow compliance documents upload
- [ ] Add evaluation report upload area
- [ ] Add "Delivery Type" selection (Standard/Express/Pick-up)
- [ ] Show these fields ONLY for procurement officer role

**Implementation**:
```typescript
interface ProcurementData {
  supplierName: string
  supplierContact: string
  supplierCode: string
  complianceDocs: Attachment[]
  evaluationReport: Attachment
  deliveryType: 'Standard' | 'Express' | 'Pick-up'
  procurementNotes: string
}

// Conditionally show procurement fields
if (currentRole === 'PROCUREMENT_OFFICER' && currentStage === 4) {
  return <ProcurementFieldsSection />
}
```

#### 1.3 Auto-Create Purchase Order on Final Approval
**File**: `src/app/_actions/workflow.ts`

- [ ] When requisition reaches final approval (stage 4), add logic to:
  - [ ] Create new PURCHASE_ORDER document
  - [ ] Link to original requisition ID
  - [ ] Copy relevant fields (items, supplier info, delivery type)
  - [ ] Set status to SUBMITTED (needs PO approval)
  - [ ] Auto-assign to Principal Officer for final approval
  - [ ] Log this transition in audit trail

**Code Location**:
```typescript
async function approveDocument(documentId: string, comments?: string) {
  // ... existing approval logic ...

  if (isLastStage && document.type === 'REQUISITION') {
    // Auto-create Purchase Order
    const poResult = await createWorkflowDocument('PURCHASE_ORDER', {
      requisitionId: documentId,
      vendorName: document.metadata.supplierInfo?.name,
      items: document.metadata.items,
      totalAmount: document.totalAmount,
      deliveryType: document.metadata.deliveryType,
      specialInstructions: document.metadata.procurementNotes
    })

    // Log the transition
    approvalLogsStore.get(documentId)?.push({
      action: 'SYSTEM_ACTION',
      message: `Purchase Order ${poResult.data.documentNumber} created`
    })
  }
}
```

#### 1.4 Update Requisitions Table Status Display
**File**: `src/app/workflows/requisitions/_components/requisitions-table.tsx`

- [ ] Show current stage in table (optional new column)
- [ ] Or enhance status badge to show stage info: "In Review (Stage 2/4)"
- [ ] Update empty state message to mention stages

#### 1.5 Add "Accountant" Role
**File**: `src/lib/mock-data.ts`

- [ ] Add accountant user to mock data
- [ ] Add ACCOUNTANT role definition
- [ ] Add to RBAC system in `src/lib/rbac.ts`

**Mock Data**:
```typescript
{
  id: 'user-acc-1',
  name: 'Francis Muleya',
  email: 'francis@example.com',
  department: 'Finance',
  role: 'ACCOUNTANT',
  image: null
}
```

**RBAC**:
```typescript
const ACCOUNTANT_PERMISSIONS = [
  'view_draft',
  'view_audit_log',
  'approve_document',
  'reject_document',
  'view_attachments',
  'add_comments',
  'add_attachments'
]
```

---

## Phase 2: Purchase Order & Payment Voucher Workflow (HIGH PRIORITY - Follow Phase 1)

### Goal
Complete the PO and Payment Voucher workflows shown in flows

### Tasks

#### 2.1 Create Purchase Order Page Structure
**New Files**:
- `src/app/workflows/purchase-orders/page.tsx`
- `src/app/workflows/purchase-orders/_components/purchase-orders-client.tsx`
- `src/app/workflows/purchase-orders/_components/purchase-orders-table.tsx`
- `src/app/workflows/purchase-orders/_components/po-detail-client.tsx`

**Implementation**: Follow same pattern as requisitions (use UI template pattern)

- [ ] List all POs created (from requisition approvals)
- [ ] Show PO status, supplier, amount, stage
- [ ] Detail page with requisition reference
- [ ] PO approval action (Principal Officer only)
- [ ] Link back to requisition

#### 2.2 Create Payment Voucher Page Structure
**New Files**:
- `src/app/workflows/payment-vouchers/page.tsx`
- `src/app/workflows/payment-vouchers/_components/payment-vouchers-client.tsx`
- `src/app/workflows/payment-vouchers/_components/payment-vouchers-table.tsx`
- `src/app/workflows/payment-vouchers/_components/pv-detail-client.tsx`
- `src/app/workflows/payment-vouchers/_components/pv-approval-panel.tsx`

**Implementation**: Follow same pattern as requisitions

#### 2.3 GRN (Goods Received Note) Management
**New Files**:
- `src/app/workflows/grn/page.tsx`
- `src/app/workflows/grn/_components/grn-form.tsx`
- Server action: `createGRN()` in `src/app/_actions/workflow.ts`

**GRN Process**:
- [ ] Triggered after PO approval
- [ ] Stores Officer receives goods
- [ ] Creates GRN with:
  - [ ] PO reference
  - [ ] Items received
  - [ ] Delivery inspection notes
  - [ ] Signature/sign-off
- [ ] Links to Payment Voucher creation

#### 2.4 Payment Voucher Specific Fields
**File**: `src/app/_actions/workflow.ts`

- [ ] When Payment Voucher created, capture:
  - [ ] Bank account info (account number, bank code, account name)
  - [ ] Vote code (financial/budgetary code)
  - [ ] Reference number
  - [ ] QR code (generated)

#### 2.5 Payment Voucher 3-Stage Approval
**File**: `src/app/_actions/workflow.ts`

- [ ] Stage 1: Director Finance approval
- [ ] Stage 2: Accountant approval
- [ ] Stage 3: Principal Officer final approval
- [ ] Each stage can approve/reject with comments
- [ ] On final approval: trigger payment notification

---

## Phase 3: Notifications & Dashboard (MEDIUM PRIORITY)

### Goal
Notify users of workflow events and provide dashboard visibility

### Tasks

#### 3.1 Create Notification System
**New Files**:
- `src/lib/notifications.ts` - Notification service
- Server action: `createNotification()` in new file `src/app/_actions/notifications.ts`

**Notification Triggers**:
- [ ] Document submitted for approval
- [ ] Document approved (auto-progressed)
- [ ] Document rejected (with reason)
- [ ] Document assigned to you for approval
- [ ] Purchase Order created
- [ ] Payment Voucher created
- [ ] Payment approved

**Implementation**:
```typescript
interface Notification {
  id: string
  userId: string
  type: 'ASSIGNMENT' | 'APPROVAL' | 'REJECTION' | 'CREATION'
  documentId: string
  documentType: string
  message: string
  read: boolean
  createdAt: Date
}
```

#### 3.2 Create Dashboard
**New Files**:
- `src/app/dashboard/page.tsx` (main dashboard)
- `src/app/dashboard/_components/pending-approvals-card.tsx`
- `src/app/dashboard/_components/submitted-documents-card.tsx`
- `src/app/dashboard/_components/workflow-stats.tsx`

**Dashboard Sections**:
- [ ] Pending Approvals (for current user's role)
- [ ] Your Submitted Documents
- [ ] Workflow Statistics:
  - [ ] Total requisitions this month
  - [ ] Average approval time
  - [ ] Pending by stage
  - [ ] Approved/Rejected counts

#### 3.3 Notification UI Component
**New Files**:
- `src/app/_components/notifications-bell.tsx`
- `src/app/_components/notification-dropdown.tsx`

**Features**:
- [ ] Bell icon in header
- [ ] Show unread count
- [ ] Dropdown with recent notifications
- [ ] Mark as read
- [ ] Click to view document
- [ ] Clear all option

---

## Phase 4: Polish & Advanced Features (LOWER PRIORITY)

### Tasks

#### 4.1 Budget Memo as Optional Stage
- [ ] Create budget memo document type
- [ ] Allow creating just memo or memo+requisition
- [ ] Memo approval before requisition creation
- [ ] Link memo to requisition

#### 4.2 Quotation & Evaluation Management
- [ ] Upload supplier quotations
- [ ] Compare quotations
- [ ] Store evaluation reports
- [ ] Track evaluations per requisition

#### 4.3 SLA Tracking
- [ ] Set approval SLA by stage
- [ ] Track if approval is overdue
- [ ] Display SLA status on detail page
- [ ] Dashboard alert for overdue items

#### 4.4 Bulk Operations
- [ ] Bulk approve similar documents
- [ ] Bulk reject with reason
- [ ] Bulk reassign approvers
- [ ] Export reports (CSV/PDF)

#### 4.5 Report Generation
- [ ] Requisition approval time report
- [ ] Cost by department report
- [ ] Approval bottleneck analysis
- [ ] Spending trend analysis

---

## Database Model Updates (When Ready)

### New Collections/Tables Needed

```typescript
// Budget Memo
CREATE TABLE budget_memos (
  id UUID PRIMARY KEY,
  creator_id UUID,
  department VARCHAR,
  budget_line VARCHAR,
  justification TEXT,
  status VARCHAR, // DRAFT, SUBMITTED, APPROVED, REJECTED
  created_at TIMESTAMP,
  updated_at TIMESTAMP
)

// Goods Received Note (GRN)
CREATE TABLE goods_received_notes (
  id UUID PRIMARY KEY,
  po_id UUID REFERENCES purchase_orders(id),
  received_by UUID,
  items JSON,
  delivery_notes TEXT,
  signature TEXT,
  created_at TIMESTAMP
)

// Notifications
CREATE TABLE notifications (
  id UUID PRIMARY KEY,
  user_id UUID,
  document_id UUID,
  type VARCHAR,
  message TEXT,
  read BOOLEAN,
  created_at TIMESTAMP
)

// Approval History (rename/extend from approval_logs)
CREATE TABLE approval_history (
  id UUID PRIMARY KEY,
  document_id UUID,
  approver_id UUID,
  stage_number INT,
  action VARCHAR, // APPROVED, REJECTED, COMMENTED
  comments TEXT,
  data JSON, // stage-specific data
  timestamp TIMESTAMP
)
```

---

## Testing Strategy

### Unit Tests
- [ ] Stage progression logic
- [ ] Role-based access control
- [ ] Approval workflow transitions
- [ ] Notification triggering

### Integration Tests
- [ ] Complete requisition flow (all 4 stages)
- [ ] Rejection and resubmission flow
- [ ] PO creation from requisition
- [ ] Payment voucher workflow
- [ ] Notification delivery

### Manual Testing Scenarios
1. **Happy Path**: Req → All approvals → PO → Payment → Complete
2. **Rejection at Stage 2**: Reject, return to creator, resubmit, proceed
3. **Procurement Stage**: Add supplier info, evaluation docs, delivery type
4. **Payment Process**: GRN → Payment Voucher → 3-stage approval
5. **Multiple Requisitions**: Ensure proper filtering and assignment

---

## Performance Considerations

- [ ] Index approval_logs by document_id for fast retrieval
- [ ] Cache pending approvals by role
- [ ] Pagination for large document lists
- [ ] Lazy-load approval history
- [ ] Background job for auto-notifications
- [ ] Archive old documents/logs quarterly

---

## Security Considerations

- [ ] Verify approver can approve at their stage only
- [ ] Prevent approvers from approving their own documents
- [ ] Log all access to sensitive documents
- [ ] Require MFA for final approvals (CFO, Director, etc.)
- [ ] Audit trail immutability verification
- [ ] Encrypt attachment downloads

---

## Success Criteria

### Phase 1 (Requisition Enhancement)
- ✅ All 4-stage approval workflow visible and functional
- ✅ Procurement officer specific fields captured and stored
- ✅ Purchase Order auto-created on final approval
- ✅ Stage information displayed on detail page
- ✅ All roles properly assigned in system

### Phase 2 (PO & Payment)
- ✅ PO approval workflow complete
- ✅ GRN creation and tracking working
- ✅ Payment Voucher 3-stage approval working
- ✅ Bank info and vote code captured
- ✅ End-to-end workflow functional

### Phase 3 (Notifications)
- ✅ Users notified of assignments/approvals
- ✅ Dashboard shows pending items
- ✅ Statistics accurate and updated
- ✅ Notification bell shows unread count

### Phase 4 (Polish)
- ✅ All optional features working
- ✅ Reports generating correctly
- ✅ Performance acceptable
- ✅ No security vulnerabilities

---

## Time Estimates

| Phase | Component | Estimated Time |
|-------|-----------|-----------------|
| 1.1 | Stage indicators | 2-3 hours |
| 1.2 | Procurement fields | 2-3 hours |
| 1.3 | Auto-create PO | 3-4 hours |
| 1.4 | Table updates | 1-2 hours |
| 1.5 | Add Accountant role | 30 mins |
| **Phase 1 Total** | **Requisition Enhancement** | **~12 hours** |
| 2.1-2.2 | PO & Payment pages | 6-8 hours |
| 2.3 | GRN management | 4-5 hours |
| 2.4-2.5 | PV approval workflow | 3-4 hours |
| **Phase 2 Total** | **PO & Payment Workflow** | **~15-20 hours** |
| 3.1-3.3 | Dashboard & Notifications | 8-10 hours |
| **Phase 3 Total** | **Notifications** | **~8-10 hours** |
| 4 | Polish features | 10-15 hours |

**Total**: ~45-55 hours (1-2 weeks full-time development)

---

## Getting Started

### Immediate Next Steps
1. Run through Phase 1 tasks in order
2. Test each stage thoroughly
3. Verify all roles and permissions work
4. Get feedback on UI/UX before Phase 2

### Code Style Guidelines
- Follow existing patterns in `src/app/workflows/`
- Use UI template components from `docs/ui-templates/`
- Keep server actions focused and reusable
- Document complex workflow logic with comments
- Test rejection scenarios at each stage

---

**Last Updated**: 2024-11-29
**Priority**: Phase 1 should start immediately
**Owner**: Development Team
