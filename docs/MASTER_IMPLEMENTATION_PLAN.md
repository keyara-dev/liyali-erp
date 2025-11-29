# Master Implementation Plan - Complete Workflow System

**Created**: 2024-11-29
**Status**: Ready for Execution
**Total Duration**: 4-6 weeks (45-60 hours)
**Priority**: Phase 1 Immediate, Phase 2 High, Phase 3 Medium

---

## Executive Summary

Complete roadmap for building the requisition-to-payment workflow system with all four document types and supporting infrastructure.

### System Components
1. **Phase 1**: Enhance Requisition Workflow (12 hours)
2. **Phase 2**: Purchase Order, GRN, Payment Voucher (32 hours)
3. **Phase 3**: Notifications & Dashboard (10 hours)
4. **Phase 4**: Polish & Advanced Features (15 hours)

### Current Status
- ✅ Core infrastructure complete
- ✅ Requisition basic workflow done (80%)
- ⚠️ Requisition Stage 4 enhancement needed
- ❌ PO, GRN, Payment Voucher not started

**Total Progress**: 43% → Target 100% in 6 weeks

---

## Phase 1: Enhance Requisition Workflow (IMMEDIATE - Week 1)

**Duration**: 12 hours
**Goal**: Complete requisition workflow to match business flows

### Tasks

#### 1.1 Add Stage Indicators (2 hours)
**Files**: `requisition-detail-client.tsx`

- [ ] Display stage progress (Stage X of 4)
- [ ] Show stage names
- [ ] Show current approver
- [ ] Show next approver
- [ ] Color-coded progress bar

**Acceptance Criteria**:
- User sees "Stage 2 of 4: Principal Officer" clearly
- Current approver name displayed
- Next approver shows in requirements section

#### 1.2 Enhance Procurement Stage (3 hours)
**Files**: `approval-action-panel.tsx`

- [ ] Add supplier info fields:
  - [ ] Supplier name
  - [ ] Supplier contact
  - [ ] Supplier code
- [ ] Add delivery type selector (Standard/Express/Pickup)
- [ ] Add special notes field
- [ ] Show fields ONLY for Procurement Officer at Stage 4

**Acceptance Criteria**:
- Procurement officer sees supplier form fields
- Can select delivery type
- Fields validated before approval
- Data saved with approval record

#### 1.3 Auto-Create Purchase Order (4 hours)
**Files**: `src/app/_actions/workflow.ts`

- [ ] Detect final stage approval for requisition
- [ ] Create PO document with:
  - [ ] Link to requisition
  - [ ] Copy items, costs, supplier info
  - [ ] Set PO number
  - [ ] Auto-assign to Principal Officer
  - [ ] Set status to SUBMITTED
- [ ] Log transition in audit trail
- [ ] Show link on requisition detail page

**Acceptance Criteria**:
- When requisition approved at stage 4, PO created automatically
- PO visible in new purchase-orders section
- Requisition shows link to created PO
- Audit trail logs the creation

#### 1.4 Add Accountant Role (2 hours)
**Files**: `src/lib/mock-data.ts`, `src/lib/rbac.ts`

- [ ] Add accountant user to mock data
- [ ] Create ACCOUNTANT role
- [ ] Assign permissions:
  - [ ] view_draft
  - [ ] approve_document
  - [ ] reject_document
  - [ ] view_audit_log
  - [ ] add_comments
  - [ ] view_attachments

**Acceptance Criteria**:
- Accountant user exists in system
- Can login with accountant role
- Has proper permissions
- Shows in approver assignments

#### 1.5 Update UI Components (1 hour)
**Files**: Various component files

- [ ] Update status colors for consistency
- [ ] Add hover effects to action buttons
- [ ] Improve empty states
- [ ] Add loading skeletons

### Success Criteria for Phase 1
- ✅ All 4 requisition stages visible in UI
- ✅ Procurement officer can add supplier info
- ✅ PO auto-created on final approval
- ✅ Stage progress clearly displayed
- ✅ Accountant role functional
- ✅ All roles have proper permissions
- ✅ No UI errors or warnings
- ✅ Audit trail complete

### Testing Checklist
- [ ] Create requisition with all items
- [ ] Submit and approve through 3 stages
- [ ] At stage 4, add supplier info and delivery type
- [ ] Approve requisition
- [ ] Verify PO created automatically
- [ ] Verify requisition shows PO link
- [ ] Check all stage transitions in audit log

---

## Phase 2: PO, GRN & Payment Voucher (High Priority - Weeks 2-4)

**Duration**: 32 hours
**Goal**: Complete purchase-to-payment workflow

### Part A: Purchase Order (8 hours)

#### 2A.1 Create PO Pages (4 hours)
**Files**:
- `src/app/workflows/purchase-orders/page.tsx`
- `src/app/workflows/purchase-orders/_components/po-client.tsx`
- `src/app/workflows/purchase-orders/_components/po-table.tsx`

**Deliverables**:
- [ ] PO list page with table
- [ ] Sort by PO #, vendor, amount
- [ ] Filter by status
- [ ] Link to originating requisition
- [ ] "View Details" button

**Acceptance Criteria**:
- PO list page loads and displays all POs
- Can sort and filter
- Shows requisition link
- Professional UI matching requisitions page

#### 2A.2 Create PO Detail & Approval (4 hours)
**Files**:
- `src/app/workflows/purchase-orders/[id]/page.tsx`
- `src/app/workflows/purchase-orders/_components/po-detail-client.tsx`

**Deliverables**:
- [ ] PO detail page layout
- [ ] Show vendor info, items, costs
- [ ] Show link to requisition
- [ ] Approval section (Principal Officer only)
- [ ] Approve/Reject buttons with comments

**Acceptance Criteria**:
- Detail page shows all PO information
- Principal Officer can approve
- Comments are captured
- Approval updates status
- No unauthorized users can approve

### Part B: Goods Received Note (8 hours)

#### 2B.1 Create GRN Form (5 hours)
**Files**:
- `src/app/workflows/grn/page.tsx`
- `src/app/workflows/grn/_components/grn-form.tsx`
- `src/app/workflows/grn/_components/grn-list.tsx`

**Deliverables**:
- [ ] GRN form with PO selector
- [ ] Item-by-item receipt fields:
  - [ ] Qty received vs ordered
  - [ ] Condition (Good/Damaged/Partial)
  - [ ] Item notes
- [ ] Inspection notes text area
- [ ] Document upload
- [ ] Digital signature
- [ ] Submit button

**Acceptance Criteria**:
- Form loads with PO data
- Can confirm items received
- Quantity mismatches flagged
- Documents uploaded
- Signature captured
- GRN saved on submit

#### 2B.2 Create GRN List & Auto-Trigger PV (3 hours)
**Files**: GRN components + workflow actions

**Deliverables**:
- [ ] GRN list page with status
- [ ] Show qty ordered vs received
- [ ] Highlight discrepancies
- [ ] Auto-create payment voucher on GRN completion

**Acceptance Criteria**:
- GRN list displays properly
- Discrepancies highlighted
- Payment voucher auto-created when GRN completed
- Link to PV shown in GRN detail

### Part C: Payment Voucher (16 hours)

#### 2C.1 Create PV List & Detail Pages (4 hours)
**Files**:
- `src/app/workflows/payment-vouchers/page.tsx`
- `src/app/workflows/payment-vouchers/[id]/page.tsx`
- `src/app/workflows/payment-vouchers/_components/pv-table.tsx`
- `src/app/workflows/payment-vouchers/_components/pv-detail-client.tsx`

**Deliverables**:
- [ ] PV list with sorting/filtering
- [ ] Show stage progress (1/3, 2/3, 3/3)
- [ ] PV detail page layout
- [ ] Display cost breakdown
- [ ] Show supporting documents
- [ ] Links to PO and GRN

**Acceptance Criteria**:
- List shows all PVs with proper info
- Detail page displays everything
- Stage progress visible
- Links to related documents work

#### 2C.2 Implement 3-Stage Approval (10 hours)
**Files**: `pv-approval-panel.tsx` + stage-specific components

**Stage 1: Director Finance (2 hours)**
- [ ] Review cost and coding
- [ ] Approve/Reject with comments
- [ ] Auto-advance to Stage 2

**Stage 2: Accountant (4 hours)**
- [ ] Show bank account form
- [ ] Show vote code field
- [ ] Validate account details
- [ ] Approve/Reject with comments
- [ ] Auto-advance to Stage 3

**Stage 3: Principal Officer (4 hours)**
- [ ] Final approval review
- [ ] Generate payment reference on approval
- [ ] Generate QR code
- [ ] Notify stakeholders
- [ ] Mark as APPROVED

**Acceptance Criteria**:
- Stage 1: Director Finance approves, advances to Stage 2
- Stage 2: Accountant fills bank info, approves, advances to Stage 3
- Stage 3: Principal Officer approves, payment reference generated
- QR code visible on approved vouchers
- All stages can reject with reason

#### 2C.3 QR Code & Payment Reference (2 hours)
**Files**: `src/lib/qr-code.ts` + PV workflow actions

**Deliverables**:
- [ ] Generate payment reference number
- [ ] Generate QR code with payment info
- [ ] Display QR code on approved vouchers
- [ ] QR code contains vendor code, amount, date, reference

**Acceptance Criteria**:
- Payment reference generated automatically
- QR code generated on final approval
- QR code displays on PV detail page
- QR code contains correct data

### Success Criteria for Phase 2
- ✅ Purchase Order workflow complete
- ✅ Goods Received Note workflow complete
- ✅ Payment Voucher list and detail pages
- ✅ 3-stage PV approval working
- ✅ Bank info validated
- ✅ Payment reference and QR code generated
- ✅ All links between documents working
- ✅ No UI errors
- ✅ Professional UI throughout

### Testing Checklist
- [ ] Create requisition through all stages
- [ ] PO auto-created and approvable
- [ ] Create GRN with items received
- [ ] PV auto-created from GRN
- [ ] PV goes through 3-stage approval
- [ ] Payment reference and QR generated
- [ ] All links (Req → PO → GRN → PV) work
- [ ] Rejection at any stage works
- [ ] Rejection sends back to appropriate stage

---

## Phase 3: Notifications & Dashboard (Medium Priority - Week 5)

**Duration**: 10 hours
**Goal**: User visibility into pending approvals

### Tasks

#### 3.1 Notification System (4 hours)
**Files**: `src/app/_actions/notifications.ts`, `src/lib/notifications.ts`

**Deliverables**:
- [ ] Create notification data model
- [ ] Create notification server actions:
  - [ ] createNotification()
  - [ ] getNotifications()
  - [ ] markAsRead()
  - [ ] deleteNotification()
- [ ] Notification types:
  - [ ] ASSIGNMENT - "Document assigned to you"
  - [ ] APPROVAL - "Document approved and advanced"
  - [ ] REJECTION - "Document rejected with reason"
  - [ ] CREATION - "New document created"

**Acceptance Criteria**:
- Notifications created on events
- Users can see their notifications
- Can mark as read
- Can delete notifications

#### 3.2 Notification Bell in Header (3 hours)
**Files**: Header component in layout

**Deliverables**:
- [ ] Bell icon with unread count
- [ ] Dropdown with recent notifications
- [ ] Click to view document
- [ ] Mark as read from dropdown
- [ ] Clear all option

**Acceptance Criteria**:
- Bell shows unread count
- Dropdown shows notifications
- Can navigate to document
- Unread count updates

#### 3.3 Dashboard Page (3 hours)
**Files**:
- `src/app/dashboard/page.tsx`
- `src/app/dashboard/_components/pending-approvals.tsx`
- `src/app/dashboard/_components/workflow-stats.tsx`

**Deliverables**:
- [ ] Dashboard layout
- [ ] Pending Approvals widget:
  - [ ] Show items awaiting current user's approval
  - [ ] Sort by date
  - [ ] Link to document
- [ ] Statistics cards:
  - [ ] Pending approvals count
  - [ ] Approved this month
  - [ ] Rejected this month
  - [ ] Average approval time
- [ ] Recent Documents
- [ ] Workflow Status Summary

**Acceptance Criteria**:
- Dashboard loads quickly
- Shows accurate pending approvals
- Statistics update correctly
- Links navigate to documents

### Success Criteria for Phase 3
- ✅ Notification system working
- ✅ Bell icon in header showing unread count
- ✅ Dashboard with pending items
- ✅ Accurate statistics
- ✅ Professional appearance

---

## Phase 4: Polish & Advanced Features (Lower Priority - Week 6)

**Duration**: 15 hours
**Goal**: Complete system with extras

### Optional Enhancements

#### 4.1 Budget Memo as Separate Stage (6 hours)
- Separate approval workflow before requisition
- Link memo to requisition creation

#### 4.2 Quotation Management (4 hours)
- Upload supplier quotations
- Compare quotes
- Track evaluation

#### 4.3 SLA Tracking (3 hours)
- Set approval SLA by stage
- Track overdue approvals
- Dashboard alerts

#### 4.4 Bulk Operations (2 hours)
- Bulk approve documents
- Bulk reject with reason
- Export reports

---

## Implementation Schedule

### Week 1: Phase 1 (12 hours)
```
Mon-Tue: Task 1.1-1.2 (Stage indicators, procurement fields)
Wed:     Task 1.3 (Auto-create PO)
Thu:     Task 1.4-1.5 (Accountant role, UI polish)
Fri:     Testing & refinement
```

### Week 2: Phase 2A (8 hours)
```
Mon-Tue: Task 2A.1-2A.2 (PO list and detail pages)
Wed-Thu: PO approval workflow
Fri:     Testing
```

### Week 3: Phase 2B (8 hours)
```
Mon-Tue: Task 2B.1 (GRN form)
Wed-Thu: Task 2B.2 (GRN list, auto-PV)
Fri:     Testing
```

### Week 4: Phase 2C (16 hours)
```
Mon-Wed: Task 2C.1-2C.2 (PV pages and Stage 1)
Thu:     Task 2C.2 (Stage 2-3 approval)
Fri:     Task 2C.3 (QR code, reference)
         Testing & integration
```

### Week 5: Phase 3 (10 hours)
```
Mon-Tue: Task 3.1-3.2 (Notifications)
Wed-Thu: Task 3.3 (Dashboard)
Fri:     Testing
```

### Week 6: Phase 4 (Optional)
```
Polish, bug fixes, advanced features
```

---

## Success Metrics

### By End of Phase 1
- [ ] Requisition workflow complete and tested
- [ ] All 4 stages visible and working
- [ ] Procurement officer stage enhanced
- [ ] PO auto-created and linked
- [ ] Accountant role functional
- [ ] No outstanding bugs

### By End of Phase 2
- [ ] Complete requisition → PO → GRN → Payment Voucher flow
- [ ] All document types functional
- [ ] 3-stage payment voucher approval working
- [ ] Payment reference and QR code generated
- [ ] All links between documents working
- [ ] No outstanding bugs

### By End of Phase 3
- [ ] Users notified of pending approvals
- [ ] Dashboard shows pending items and statistics
- [ ] Professional notification system
- [ ] System 80%+ feature complete

### By End of Phase 4
- [ ] All planned features complete
- [ ] System 100% feature complete
- [ ] Professional Polish
- [ ] Ready for production

---

## Effort Summary

| Phase | Component | Hours | Status |
|-------|-----------|-------|--------|
| 1 | Requisition Enhancement | 12 | Not Started |
| 2 | Purchase Order | 8 | Not Started |
| 2 | Goods Received Note | 8 | Not Started |
| 2 | Payment Voucher | 16 | Not Started |
| 3 | Notifications | 10 | Not Started |
| 4 | Polish & Extras | 15 | Not Started |
| **TOTAL** | | **69 hours** | |

**Timeline**: 4-6 weeks (10-15 hours/week)

---

## Key Files to Create

### New Directories
```
src/app/workflows/purchase-orders/
src/app/workflows/grn/
src/app/workflows/payment-vouchers/
src/app/dashboard/
```

### New Files
```
PO: 8 files
GRN: 6 files
PV: 10 files
Dashboard: 4 files
Notifications: 2 files
Utils: 2 files

Total: 32 new component files + server actions
```

### Files to Modify
```
src/app/_actions/workflow.ts - Add 15+ new functions
src/app/_actions/notifications.ts - New notification actions
src/lib/mock-data.ts - Add more mock users, extend data
src/lib/rbac.ts - Verify roles complete
src/app/workflows/requisitions/_components/requisition-detail-client.tsx
```

---

## Risk Mitigation

### Technical Risks
- **Large refactor scope**: Break into phases, test each phase
- **Data consistency**: Ensure linking works (Req → PO → GRN → PV)
- **Approval logic complexity**: Extensive testing for multi-stage flows

### Mitigation Strategies
1. Weekly testing of completed features
2. Integration tests for document linking
3. Clear separation of concerns
4. Comprehensive error handling
5. Detailed audit trail for debugging

---

## Quality Assurance

### Unit Testing
- Server action functions
- Role-based access control
- Data validation
- Approval transitions

### Integration Testing
- Complete workflows (all phases)
- Document linking
- Notification triggering
- Status updates

### Manual Testing
- Happy path (approval flow)
- Rejection scenarios
- Edge cases (discrepancies, overdue)
- UI/UX polish

### Acceptance Criteria
- All documented tests pass
- No console errors
- All features work as specified
- Professional UI/UX
- Performance acceptable (< 2s page load)

---

## Documentation to Create

- [ ] API documentation for new actions
- [ ] User guide for each workflow
- [ ] Admin guide for configuration
- [ ] Technical documentation for future developers

---

## Deployment Checklist

Before going live:
- [ ] All tests pass
- [ ] No outstanding bugs
- [ ] Stakeholders sign off
- [ ] Training materials ready
- [ ] Backup and recovery plan
- [ ] Performance tested under load
- [ ] Security audit completed
- [ ] User documentation complete

---

## Getting Started

### Immediate Next Steps
1. Review this master plan with team
2. Start Phase 1 implementation
3. Set up code review process
4. Schedule daily standups
5. Create tracking board for progress

### Resources Needed
- Frontend developer (40% for 6 weeks)
- Backend support (testing server actions)
- QA tester (testing each phase)
- Product owner (acceptance testing)

---

## Success Definition

**The system is successful when:**
1. User can create requisition → approve through 4 stages
2. Purchase Order auto-created and approved
3. Goods Received Note created with discrepancy tracking
4. Payment Voucher created and approved through 3 stages
5. Payment reference and QR code generated
6. Users notified of pending approvals
7. Dashboard shows accurate pending items
8. All workflow steps logged immutably
9. System is performant and bug-free
10. Ready for production deployment

---

**Created**: 2024-11-29
**Status**: Ready for Implementation
**Next**: Begin Phase 1 - Week 1
**Owner**: Development Team
**Questions**: Refer to PHASE2_DETAILED_SPECS.md for detailed specifications
