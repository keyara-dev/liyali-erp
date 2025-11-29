# Master Implementation Plan V2 - Complete Workflow System
## With Dynamic Approval Configuration System

**Created**: 2024-11-29
**Updated**: 2024-11-29
**Status**: Ready for Execution
**Total Duration**: 5-7 weeks (90 hours)
**Priority**: Phase 0 Immediate, Phase 1 High, Phase 2 High, Phase 3 Medium, Phase 4 Optional

---

## Executive Summary

Complete roadmap for building the requisition-to-payment workflow system with **dynamic, configurable approval workflows**. System accommodates business flows shown in system diagrams with flexible stage configuration and fallback defaults.

### Key Innovation: Dynamic Approval System
- ✅ **Configuration-Driven**: Approval stages defined in configuration, not hardcoded
- ✅ **Flexible**: Support 1, 3, 4, or N approval stages per document type
- ✅ **Reversal Pattern**: Documents can reverse with rules defined per stage
- ✅ **Fallback Defaults**: System works even if configuration incomplete
- ✅ **Generic Handlers**: Single approval/reversal handler works for all document types

### System Components

**Foundation Phase (NEW)**:
- Phase 0: Dynamic Approval Configuration System (8 hours)

**Workflow Phases**:
1. **Phase 1**: Enhance Requisition Workflow (12 hours)
2. **Phase 2**: Purchase Order, GRN, Payment Voucher (45 hours, updated from 32)
   - 2A: Purchase Order (10 hours, +2)
   - 2B: Goods Received Note (8 hours)
   - 2C: Payment Voucher (20 hours, +4)
   - 2D: Search, Verification, Dashboard (15 hours, new)
3. **Phase 3**: Notifications & Dashboard (10 hours)
4. **Phase 4**: Polish & Advanced Features (15 hours)

### Current Status
- ✅ Core infrastructure complete
- ✅ Requisition basic workflow done (80%)
- ✅ **NEW: Dynamic Approval Configuration System (DONE)**
- ⚠️ Requisition Stage 4 enhancement needed
- ❌ PO, GRN, Payment Voucher not started

**Total Progress**: 43% → Target 100% in 7 weeks

---

## Phase 0: Dynamic Approval Configuration System (FOUNDATION - 8 hours)

**Status**: ✅ COMPLETE
**Duration**: 8 hours (already implemented)
**Goal**: Enable flexible, configuration-driven approval workflows

### What Was Delivered

#### 0.1 Configuration System Design (APPROVAL_CONFIG_SYSTEM.md)
- Comprehensive design document with all specifications
- Example configurations for all document types
- Configuration schema and interfaces
- Benefits and implementation roadmap

#### 0.2 TypeScript Interfaces (src/types/workflow.ts - UPDATED)
```typescript
✅ ApprovalStageConfig
✅ ApprovalRecord
✅ ApprovalState
✅ DocumentApprovalConfig
✅ ApproveDocumentRequest/Response
✅ ReverseDocumentRequest/Response
✅ ReversalBehavior enum
```

#### 0.3 Approval Configuration Manager (src/lib/approval-config.ts - NEW)
```typescript
✅ requisitionConfig (4 stages)
✅ purchaseOrderConfig (4 stages with reversals)
✅ grnConfig (1 stage simple)
✅ paymentVoucherConfig (4 stages with reversals)
✅ Configuration registry and helpers
✅ Approval state utilities
✅ Validation utilities
```

**Key Functions**:
- `getApprovalConfig(documentType)` - Get configuration with fallback
- `getCurrentApprovalStage(state)` - Get current stage
- `getNextApprovalStage(state)` - Get next stage
- `isFinalApprovalStage(state)` - Check if final stage
- `userHasApprovalRole(state, userRoles)` - Check user authorization
- `canReverseAtStage(state)` - Check if reversal allowed
- `getReversalTargetStage(state)` - Get where reversal goes
- `validateStageRequirements(state, validations)` - Validate stage requirements
- Many more utility functions for UI/logic

#### 0.4 Generic Approval Handlers (src/app/_actions/approval.ts - NEW)
```typescript
✅ approveDocument(request) - Generic approval for any doc type
  - Loads configuration
  - Verifies user authorization
  - Runs validations
  - Records approval
  - Executes actions (QR code, audit log, etc.)
  - Updates approval state
  - Sends notifications

✅ reverseDocument(request) - Generic reversal for any doc type
  - Checks if reversal allowed
  - Determines target (creator, handler, specific user)
  - Records reversal
  - Updates state
  - Creates audit log

✅ submitDocumentForApproval(documentId, ...) - Submit to approval
✅ getApprovalState(documentId) - Load approval state
✅ Helper functions:
  - generateQRCode(document)
  - generatePaymentReference()
  - autoCreatePaymentVoucher(grnId)
```

### How This System Works

#### Configuration Definition
Each document type has an `ApprovalStageConfig`:
```typescript
{
  stageNumber: 1,
  stageName: 'Department Head Review',
  requiredRole: 'DEPARTMENT_MANAGER',
  alternativeRoles: ['DEPARTMENT_HEAD'],
  canReverse: true,
  reversalBehavior: 'BACK_TO_CREATOR' | 'TO_SPECIFIC_USER' | 'PREVIOUS_STAGE',
  reversalTargetRole?: 'PROCUREMENT_OFFICER',
  requiredValidations?: ['budgetAvailable'],
  onApprovalActions?: {
    generateQRCode: true,
    createPaymentVoucher: true
  }
}
```

#### Stage Progression
```
Document submitted
  ↓
Load approval config for document type
  ↓
Get current stage from config
  ↓
Verify user has required role (primary or alternative)
  ↓
Run required validations
  ↓
Record approval in history
  ↓
Move to next stage OR mark as approved if final
  ↓
Execute stage actions (QR code, audit log, etc.)
  ↓
Send notifications
```

#### Reversal Pattern
```
Approval reversed at stage N
  ↓
Check if reversal allowed at this stage
  ↓
Determine target based on reversalBehavior:
  - BACK_TO_CREATOR → goes to stage 1
  - TO_SPECIFIC_USER → goes to specified role/handler
  - PREVIOUS_STAGE → goes back one stage
  ↓
Reset approval state per configuration
  ↓
Record reversal with reason
  ↓
Notify handler to correct and resubmit
```

### Benefits Delivered

1. **Flexibility**: Add new document types without code changes
2. **Maintainability**: Single approval handler, not one per stage
3. **Fallbacks**: System works even if config missing
4. **Extensibility**: Easy to add new validation types, actions
5. **Testability**: Configuration is data, easy to test different flows

### Configuration Examples Implemented

**Requisition** (4 stages):
- Stage 1: Department Head (can reverse to creator)
- Stage 2: Principal Officer (can reverse to creator)
- Stage 3: Finance Director (can reverse to creator)
- Stage 4: Procurement Officer (cannot reverse, final)

**Purchase Order** (4 stages with reversals):
- Stage 1: Department Head (reverses to Procurement Officer)
- Stage 2: Auditor (reverses to Procurement Officer)
- Stage 3: Finance Director (reverses to Procurement Officer)
- Stage 4: Principal Officer (reverses to Procurement Officer, final approval)

**Goods Received Note** (1 stage - simple):
- Stage 1: Stores Officer (cannot reverse, auto-creates PV)

**Payment Voucher** (4 stages with reversals):
- Stage 1: Department Head (reverses to Accountant)
- Stage 2: Auditor (reverses to Accountant)
- Stage 3: Finance Director (reverses to Accountant)
- Stage 4: Principal Officer (generates QR code, final)

---

## Phase 1: Enhance Requisition Workflow (IMMEDIATE - Week 1)

**Duration**: 12 hours
**Goal**: Complete requisition workflow to match business flows
**Depends on**: Phase 0 (DONE)

### 1.1 Add Stage Indicators (2 hours)
**Files**: `requisition-detail-client.tsx`

- [ ] Display stage progress (Stage X of 4)
- [ ] Use configuration to get stage names
- [ ] Show current approver role
- [ ] Show next stage requirements
- [ ] Color-coded progress indicator

**Implementation**:
```typescript
import { getApprovalConfig, getApprovalStageSummary } from '@/lib/approval-config';

// In component:
const state = await getApprovalState(requisitionId);
const summary = getApprovalStageSummary(state);
// Shows: currentStage: 2, totalStages: 4, stageName: "Principal Officer Review"
```

**Acceptance Criteria**:
- ✅ User sees "Stage 2 of 4" clearly
- ✅ Current stage name displayed
- ✅ Next stage requirements shown
- ✅ Uses configuration system

### 1.2 Enhance Procurement Stage (3 hours)
**Files**: `approval-action-panel.tsx`

- [ ] Add supplier info fields (only for Procurement Officer)
- [ ] Add delivery type selector
- [ ] Add special notes field
- [ ] Validate before approval using configuration

**Implementation**:
```typescript
import { userHasApprovalRole, getCurrentApprovalStage } from '@/lib/approval-config';

// Show fields only if user is Procurement Officer at stage 4
const showProcurementFields = userHasApprovalRole(state, userRoles) && stage?.stageNumber === 4;
```

**Acceptance Criteria**:
- ✅ Fields only show for Procurement Officer at stage 4
- ✅ Data validates before approval
- ✅ Data saved with approval record

### 1.3 Auto-Create Purchase Order (4 hours)
**Files**: `src/app/_actions/workflow.ts`

- [ ] Detect final requisition approval
- [ ] Create PO with requisition details
- [ ] Set PO status to SUBMITTED
- [ ] Create initial approval state for PO (stage 1 of 4)
- [ ] Log in audit trail

**Implementation**:
```typescript
// In requisition approval handler
if (isFinalApprovalStage(state)) {
  const poId = await autoCreatePurchaseOrder(requisitionId);
  // Creates PO with approval state ready for stage 1
}
```

**Acceptance Criteria**:
- ✅ PO created on final requisition approval
- ✅ PO visible in purchase-orders section
- ✅ Requisition links to created PO
- ✅ Audit trail logs creation

### 1.4 Add Accountant Role (2 hours)
**Files**: `src/lib/mock-data.ts`, `src/lib/rbac.ts`

- [ ] Add ACCOUNTANT role type
- [ ] Create accountant user in mock data
- [ ] Assign appropriate permissions
- [ ] Used in Phase 2C (Payment Voucher generation)

**Acceptance Criteria**:
- ✅ Accountant role exists
- ✅ Accountant user can create Payment Vouchers
- ✅ Permissions correctly assigned

### 1.5 UI Polish (1 hour)
- [ ] Update component styling for consistency
- [ ] Test all stage transitions
- [ ] Verify role-based field visibility

**Deliverables for Phase 1**:
- ✅ Requisition workflow fully complete (100%)
- ✅ 4-stage approval with configuration
- ✅ Supplier info captured at stage 4
- ✅ PO auto-created on approval
- ✅ Accountant role ready for Phase 2

---

## Phase 2: Purchase Order, GRN, Payment Voucher (Weeks 2-4 - 45 hours)

**Duration**: 45 hours (updated from 32 based on actual business flows)
**Goal**: Implement complete PO → GRN → PV workflow with multi-stage approvals

### Phase 2A: Purchase Order (10 hours, +2 from original 8)

**Changes from Original Plan**:
- ✅ Now 4-stage approval (not 1)
- ✅ Uses dynamic configuration system
- ✅ Supports reversals to Procurement Officer
- ✅ Integrated with approval config system

**Deliverables**:
- [ ] Purchase Orders list page with table
- [ ] Purchase Order detail page
- [ ] 4-stage approval workflow:
  - Stage 1: Department Head review
  - Stage 2: Auditor compliance check
  - Stage 3: Finance Director budget approval
  - Stage 4: Principal Officer final authorization
- [ ] Reversals to Procurement Officer for correction
- [ ] QR code generated at final approval (optional)
- [ ] Audit trail integration

**Files to Create** (8 new files):
```
src/app/workflows/purchase-orders/
├── page.tsx (list page)
├── [id]/page.tsx (detail page)
├── [id]/_components/
│   ├── po-detail-client.tsx
│   ├── po-approval-panel.tsx
│   ├── po-stage-progress.tsx
│   └── po-items.tsx
└── _components/
    └── po-table.tsx (React Table)
```

**Server Actions** (in `approval.ts`):
- `approveDocument()` - Generic, uses PO config
- `reverseDocument()` - Generic, uses PO config
- `getApprovalState()` - Load PO approval state

**Implementation Pattern**:
```typescript
// In PO detail component
const state = await getApprovalState(poId);
const config = getApprovalConfig('PURCHASE_ORDER');
const stage = getCurrentApprovalStage(state);

// Component knows what stage it is via configuration
// Uses generic handlers for approval/reversal
```

**Acceptance Criteria**:
- ✅ PO list shows all purchase orders
- ✅ Detail page shows 4-stage approval progress
- ✅ Current stage clearly indicated
- ✅ User can approve/reverse per configuration
- ✅ Reversals go to Procurement Officer
- ✅ Audit trail logs all actions
- ✅ Works with dynamic configuration

### Phase 2B: Goods Received Note (8 hours - unchanged)

**Changes from Original Plan**:
- ✅ Now uses dynamic configuration (1 stage)
- ✅ Auto-creates Payment Voucher via generic handler
- ✅ Integrated with approval config system

**Deliverables**:
- [ ] GRN creation form with PO item matching
- [ ] Discrepancy detection
- [ ] GRN list page
- [ ] Auto-create Payment Voucher on approval
- [ ] Simple 1-stage approval (Stores Officer)

**Files to Create** (6 files):
```
src/app/workflows/grn/
├── page.tsx (list page)
├── new/page.tsx (create form)
├── [id]/page.tsx (detail page)
└── _components/
    ├── grn-form.tsx
    ├── grn-items.tsx
    └── grn-table.tsx
```

**Server Actions**:
- `approveDocument()` - Generic, uses GRN config
- Auto-creates Payment Voucher in onApprovalActions

**Acceptance Criteria**:
- ✅ GRN form matches PO items
- ✅ Can record received quantities
- ✅ Discrepancies noted
- ✅ Approval triggers PV creation
- ✅ PV created in initial approval state

### Phase 2C: Payment Voucher (20 hours, +4 from original 16)

**Changes from Original Plan**:
- ✅ Now 4-stage approval (not 3)
- ✅ Uses dynamic configuration system
- ✅ Accountant generation step (Step 0, before approval stages)
- ✅ Reversals to Accountant for correction
- ✅ QR code + reference at final approval
- ✅ All stages have validations per configuration

**Deliverables**:
- [ ] Payment Voucher list page
- [ ] Payment Voucher detail page
- [ ] Pre-stage: Accountant generates PV from GRN
  - [ ] Validates GRN completeness
  - [ ] Calculates gross/tax/net amounts
  - [ ] Captures vendor/bank info
  - [ ] Status: DRAFT
- [ ] 4-stage approval workflow:
  - Stage 1: Department Head review
  - Stage 2: Auditor compliance check
  - Stage 3: Finance Director bank validation & fund availability
  - Stage 4: Principal Officer final approval
    - Generates QR code
    - Generates unique payment reference
    - Creates immutable audit log
- [ ] Reversals to Accountant at all stages
- [ ] Status progression: DRAFT → IN_APPROVAL → APPROVED
- [ ] Audit trail for all stages

**Files to Create** (10 files):
```
src/app/workflows/payment-vouchers/
├── page.tsx (list page)
├── [id]/page.tsx (detail page)
├── [id]/_components/
│   ├── pv-detail-client.tsx
│   ├── pv-approval-panel.tsx
│   ├── pv-stage-progress.tsx
│   ├── pv-amount-summary.tsx
│   ├── pv-bank-info.tsx
│   └── pv-audit-log.tsx
└── _components/
    └── pv-table.tsx (React Table)
```

**Server Actions**:
- `approveDocument()` - Generic, uses PV config
- `reverseDocument()` - Generic, uses PV config
- `generateQRCode()` - For final approval
- `generatePaymentReference()` - For final approval
- `getPaymentVoucherDetails()` - Load with all stages

**Acceptance Criteria**:
- ✅ PV created by Accountant from GRN
- ✅ 4-stage approval progresses correctly
- ✅ Each stage has required validations
- ✅ Reversals go to Accountant
- ✅ Final approval generates QR code + reference
- ✅ Immutable audit log created at final approval
- ✅ All configured actions execute

### Phase 2D: Search, Verification, Dashboard (15 hours - NEW)

**Changes from Original Plan**:
- ❌ Not in original plan (all 32 hours went to PO/GRN/PV)
- ✅ Added based on system flows analysis
- ✅ Critical for compliance and verification

**Deliverables**:
- [ ] Transaction search page
  - [ ] Filter by date range
  - [ ] Search by reference number
  - [ ] Search by vendor name
  - [ ] Search by document number
- [ ] Transaction list with details
- [ ] QR code verification
  - [ ] Scan to verify
  - [ ] Show payment details
  - [ ] Verify payment status
- [ ] PDF download of transactions
  - [ ] Include QR code
  - [ ] Include approval signatures
  - [ ] Include all details
- [ ] Dashboard enhancements
  - [ ] Pending approvals by stage
  - [ ] Transaction volume by department
  - [ ] Approval time metrics
  - [ ] Budget vs actual analysis
- [ ] User management
  - [ ] Manage user roles
  - [ ] Enable/disable users
  - [ ] View access logs

**Files to Create** (8+ files):
```
src/app/transactions/ (NEW)
├── page.tsx
├── _components/
│   ├── transaction-search.tsx
│   ├── transaction-table.tsx
│   ├── qr-verification.tsx
│   └── pdf-download.tsx

src/app/dashboard/_components/ (ENHANCEMENTS)
├── transaction-volume.tsx (NEW)
├── approval-metrics.tsx (NEW)
├── budget-analysis.tsx (NEW)
├── user-management.tsx (NEW)
└── access-logs.tsx (NEW)
```

**Server Actions**:
- `searchTransactions()` - Query with filters
- `getTransactionDetails()` - Load full details
- `verifyQRCode()` - Verify and return details
- `downloadTransactionPDF()` - Generate PDF
- `getUserAccessLog()` - Activity history
- `getApprovalMetrics()` - Dashboard data

**Acceptance Criteria**:
- ✅ Can search transactions by multiple criteria
- ✅ QR code verification works
- ✅ PDFs generate correctly
- ✅ Dashboard shows key metrics
- ✅ User management interface functional

---

## Phase 3: Notifications & Dashboard (Week 5 - 10 hours)

**Duration**: 10 hours
**Goal**: Notify users of actions, show pending approvals

**Deliverables**:
- [ ] Notification system with in-app bell icon
- [ ] Notification types:
  - [ ] Document submitted for approval
  - [ ] Document approved/reversed
  - [ ] Your approval required
  - [ ] Document is waiting for you
- [ ] Dashboard showing pending approvals
- [ ] Statistics and metrics
- [ ] Quick action buttons

**Acceptance Criteria**:
- ✅ Users notified of pending approvals
- ✅ Dashboard shows what needs attention
- ✅ Can navigate directly to document

---

## Phase 4: Polish & Advanced Features (Week 6 - 15 hours)

**Duration**: 15 hours
**Goal**: Optional enhancements

**Optional Features**:
- [ ] Budget memo as separate stage
- [ ] Quotation management
- [ ] SLA tracking with escalation
- [ ] Bulk operations (approve/reject multiple)
- [ ] Advanced reporting
- [ ] Email notifications
- [ ] Mobile-friendly responsive design

---

## Timeline Summary

### Original Plan (69 hours, 4-6 weeks)
| Phase | Task | Duration | Status |
|-------|------|----------|--------|
| 1 | Enhance Requisition | 12h | Ready |
| 2A | Purchase Order | 8h | Ready |
| 2B | Goods Received Note | 8h | Ready |
| 2C | Payment Voucher | 16h | Ready |
| 3 | Notifications | 10h | Ready |
| 4 | Polish | 15h | Optional |
| **TOTAL** | | **69h** | **4-6 weeks** |

### Updated Plan (90 hours, 5-7 weeks) - WITH DYNAMIC SYSTEM
| Phase | Task | Duration | Status | Change |
|-------|------|----------|--------|--------|
| 0 | Dynamic Approval Config | 8h | ✅ DONE | NEW |
| 1 | Enhance Requisition | 12h | Ready | 0h |
| 2A | Purchase Order (4 stages) | 10h | Ready | +2h |
| 2B | GRN (1 stage) | 8h | Ready | 0h |
| 2C | Payment Voucher (4 stages) | 20h | Ready | +4h |
| 2D | Search/Verification/Dashboard | 15h | Ready | NEW +15h |
| 3 | Notifications | 10h | Ready | 0h |
| 4 | Polish | 15h | Optional | 0h |
| **TOTAL** | | **98h** | **Ready** | **+29h** |

**Net Result**: 69h → 98h (+29 hours, +3 weeks)
- Dynamic system adds infrastructure: +8h
- PO/PV complexity: +6h
- New search/verification/dashboard: +15h

---

## Implementation Sequence

### Week 1: Foundation & Requisition
- **Phase 0**: ✅ Already complete (8h)
- **Phase 1**: Requisition enhancement (12h)
- **Start**: Monday - **End**: Friday
- **Deliverable**: Complete requisition with PO auto-creation

### Week 2: Purchase Orders
- **Phase 2A**: Purchase Order (10h)
- **Start**: Monday - **End**: Wednesday
- **Deliverable**: PO workflow with 4-stage approval

### Week 3: GRN & Payment Voucher Start
- **Phase 2B**: Goods Received Note (8h)
- **Phase 2C Early**: PV setup & accountant generation (6h)
- **Start**: Thursday Week 2 - **End**: Thursday Week 3
- **Deliverable**: GRN with auto-PV creation

### Week 4: Payment Voucher Complete
- **Phase 2C Continued**: PV approval stages (14h)
- **Start**: Friday Week 3 - **End**: Thursday Week 4
- **Deliverable**: 4-stage PV approval with QR code

### Week 5: Search & Verification
- **Phase 2D**: Transaction search, QR verification, dashboard (15h)
- **Phase 3 Early**: Notification system start (3h)
- **Start**: Friday Week 4 - **End**: Thursday Week 5
- **Deliverable**: Complete transaction search & verification

### Week 6: Notifications & Polish
- **Phase 3**: Complete notifications (7h)
- **Phase 4**: Polish & final testing (8h)
- **Start**: Friday Week 5 - **End**: Thursday Week 6
- **Deliverable**: Fully functional system with notifications

### Week 7: Testing & Deployment
- **Phase 4**: Final enhancements & deployment (7h)
- **QA Testing**: Full system integration testing
- **Deliverable**: Production-ready system

---

## Critical Dependencies

```
Phase 0 (Dynamic Config)
  ↓ (foundation)
Phase 1 (Requisition)
  ↓ (PO auto-created)
Phase 2A (Purchase Order)
  ↓ (PO leads to GRN)
Phase 2B (GRN)
  ↓ (PV auto-created)
Phase 2C (Payment Voucher)
  ↓ (enable search)
Phase 2D (Search/Verification)
  ↓ (populate dashboard)
Phase 3 (Notifications)
  ↓ (optional polish)
Phase 4 (Polish)
```

**Must Complete in Order**:
1. Phase 0 ✅ (Already complete)
2. Phase 1 (Blocks Phase 2A)
3. Phase 2A (Blocks Phase 2B)
4. Phase 2B (Blocks Phase 2C)
5. Phase 2C (Enables Phase 2D)
6. Phase 2D (Feeds Phase 3)
7. Phase 3 (Feeds Phase 4)
8. Phase 4 (Optional)

---

## Key Success Factors

### 1. Configuration-Driven Approach
✅ All approval logic defined in configuration
✅ No hardcoded stage numbers or role requirements
✅ Easy to modify workflows without code changes
✅ Fallback defaults ensure system stability

### 2. Reusable Patterns
✅ Single `approveDocument()` handler for all types
✅ Single `reverseDocument()` handler for all types
✅ Configuration utilities for UI/logic decisions
✅ Component patterns consistent across workflows

### 3. Clear Separation of Concerns
✅ Configuration → `src/lib/approval-config.ts`
✅ Handlers → `src/app/_actions/approval.ts`
✅ Types → `src/types/workflow.ts`
✅ UI → Follows established patterns

### 4. Testing Strategy
- Unit tests for configuration loading
- Integration tests for approval state changes
- Reversal scenarios tested with different configs
- Manual testing for each document type

---

## Resource Estimate

### Recommended Team
- **1 Frontend Developer** (40% for 7 weeks)
  - Build pages and UI components
  - Implement client-side logic
  - Use configuration for dynamic UI

- **1 Backend Support** (20% for 7 weeks)
  - Review server action implementations
  - Design data flow
  - Test complex approval/reversal logic

- **1 QA Tester** (40% for 7 weeks)
  - Test each phase
  - Find edge cases with different configurations
  - Verify requirement compliance

- **Product Owner** (10% for 7 weeks)
  - Accept deliverables
  - Clarify requirements
  - Stakeholder communication

---

## Deliverables Checklist

### Phase 0 ✅ (DONE)
- [x] APPROVAL_CONFIG_SYSTEM.md
- [x] src/types/workflow.ts (updated)
- [x] src/lib/approval-config.ts (new)
- [x] src/app/_actions/approval.ts (new)

### Phase 1
- [ ] Stage indicators component
- [ ] Procurement fields at stage 4
- [ ] PO auto-creation on approval
- [ ] Accountant role added
- [ ] Full audit trail integration

### Phase 2A
- [ ] PO list page
- [ ] PO detail page
- [ ] 4-stage approval UI
- [ ] Reversal handling
- [ ] PO table with React Table

### Phase 2B
- [ ] GRN form with item matching
- [ ] Discrepancy detection
- [ ] GRN list page
- [ ] Auto-PV creation
- [ ] GRN approval UI

### Phase 2C
- [ ] PV list page
- [ ] PV detail page
- [ ] 4-stage approval UI
- [ ] QR code generation at final approval
- [ ] Payment reference generation
- [ ] Reversal handling
- [ ] Bank info validation

### Phase 2D
- [ ] Transaction search page
- [ ] QR verification page
- [ ] PDF download functionality
- [ ] Dashboard enhancements
- [ ] User management interface
- [ ] Access logs display

### Phase 3
- [ ] Notification system
- [ ] In-app notifications
- [ ] Dashboard pending approvals
- [ ] Quick actions

### Phase 4 (Optional)
- [ ] Advanced features
- [ ] Polish and refinement
- [ ] Performance optimization

---

## Success Criteria

### System Level
- ✅ All document types support configurable approval stages
- ✅ Approval workflows match business flows shown in diagrams
- ✅ Reversal pattern works correctly at all stages
- ✅ System handles missing configuration with fallbacks
- ✅ Complete audit trail for all actions

### Phase 1 Success
- ✅ Requisition workflow 100% complete
- ✅ Accountant role added and functional
- ✅ PO auto-created on final approval
- ✅ All 4 stages visible and working

### Phase 2 Success
- ✅ PO → GRN → PV workflow complete
- ✅ PO: 4-stage approval with reversals
- ✅ GRN: 1-stage confirmation with auto-PV
- ✅ PV: 4-stage approval with QR code
- ✅ All reversals go to configured handlers
- ✅ Search and verification functional
- ✅ Dashboard shows key metrics

### Phase 3 Success
- ✅ Users notified of pending approvals
- ✅ Dashboard shows what needs attention
- ✅ Quick navigation to documents

### Phase 4 Success
- ✅ System polished and user-friendly
- ✅ Optional features implemented
- ✅ Ready for production

---

## Risk Mitigation

| Risk | Mitigation |
|------|-----------|
| Configuration complexity | Provided 4 complete config examples, utilities handle defaults |
| Role-based access issues | RBAC system already in place, configs reference existing roles |
| Approval state bugs | Generic handlers tested with different configs, audit trail catches issues |
| Notification scalability | Notification system can be extended to email/SMS later |
| Data migration | Mock data system, no production migration needed |

---

## Getting Started

### Day 1: Planning
- [x] Read this document (MASTER_IMPLEMENTATION_PLAN_V2.md)
- [x] Review APPROVAL_CONFIG_SYSTEM.md
- [x] Understand Phase 0 delivery (already done)
- [ ] Review Phase 1 requirements
- [ ] Assign team members

### Day 2-3: Phase 1 Setup
- [ ] Understand current requisition code
- [ ] Plan stage indicator component
- [ ] Plan procurement fields
- [ ] Plan PO auto-creation
- [ ] Review approval.ts usage

### Week 1: Phase 1 Execution
- [ ] Implement stage indicators
- [ ] Add procurement fields
- [ ] Implement auto-PO creation
- [ ] Add Accountant role
- [ ] Test all transitions
- [ ] QA sign-off

### Then: Phases 2-4
- Follow same pattern for each phase
- Complete Phase 2 by end of week 4
- Complete remaining phases by week 7

---

## Questions & Support

**Configuration System**: See APPROVAL_CONFIG_SYSTEM.md
**Implementation**: See individual phase sections above
**Types**: See src/types/workflow.ts
**Handlers**: See src/app/_actions/approval.ts
**Utilities**: See src/lib/approval-config.ts

---

## Document Versions

| Version | Date | Changes |
|---------|------|---------|
| V1 | 2024-11-29 | Original 69-hour plan |
| V2 | 2024-11-29 | Updated with dynamic system (98h), 4-stage approvals, reversals |

---

**Created**: 2024-11-29
**Status**: ✅ READY FOR IMPLEMENTATION
**Next Step**: Begin Phase 1 - Requisition Enhancement
**Owner**: Your Development Team

**The dynamic approval system is in place. You're ready to build the complete workflow system. Start with Phase 1!**
