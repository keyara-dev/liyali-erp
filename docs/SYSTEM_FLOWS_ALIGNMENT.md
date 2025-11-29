# System Flows Alignment - Updated Implementation Plan

**Date**: 2024-11-29
**Status**: Updated based on system flow diagrams
**Priority**: Critical - Must align implementation with business processes

---

## Overview

Your system flow diagrams reveal a more complex approval structure than initially planned. This document updates the implementation plan to match your actual business processes.

---

## Key Findings from System Flows

### 1. Dashboard Structure (More Complex)
**Reporting & Dashboards branch includes**:
- Transaction Volume by Department
- Pending Approvals Summary
- Average Approval Time
- Budget vs Actual Analysis

**Search & Transactions branch includes**:
- Filter by Date Range
- Search by Reference Number
- Download Transaction PDFs
- Access Attached Documents
- QR Code Verification

**User Access Management branch includes**:
- Role Assignment
- Add/Remove Users
- Multi-Factor Authentication Setup
- Access Logs & Activity History

### 2. Purchase Order Approval Flow (More Stages than Planned)
**Current Plan**: 1 stage (Principal Officer only)
**Actual Requirement**: 4-stage approval
1. **Dept Head Approval** (with rejection path back to Procurement Officer)
2. **Auditor Approval** (with reversal option)
3. **Director Finance Approval** (with reversal option)
4. **Principal Officer Approval** (final, with reversal option)

**Rejection Handling**: At ANY stage, "Reversal" sends back to Procurement Officer for correction

### 3. Payment Voucher Approval Flow (More Complex)
**Current Plan**: 3-stage approval
**Actual Requirement**: 4-stage approval with more complex reversals
1. **Accountant Generates PV** (from GRN)
2. **Dept Head Approval** (can reverse)
3. **Auditor Approval** (can reverse)
4. **Director Finance Approval** (can reverse)
5. **Principal Officer Final Approval**
   - Generates QR Code & Reference Number
   - Creates Payment Voucher

**Reversal Paths**:
- Any stage can reverse back to Accountant for correction
- If Principal Officer reverses, goes back to Accountant
- Final approved status: "Final Approved Voucher with QR Code & Ref No"

**Post-Approval**:
- Audit Log Updated & Transaction Stored
- System creates immutable record

### 4. Additional Features Identified
**Transaction Management**:
- Filter by Date Range
- Search by Reference Number
- Download PDFs
- Access Attached Documents
- QR Code Verification (scan to validate payment)

**Reporting Requirements**:
- Transaction Volume by Department
- Pending Approvals Summary
- Average Approval Time
- Budget vs Actual Analysis

**User Management**:
- Role Assignment
- Add/Remove Users
- Multi-Factor Authentication Setup
- Access Logs & Activity History

---

## Updated Phase 2 Specifications

### Part A: Purchase Order (UPDATED - More Complex)

#### Approval Stages (4 instead of 1)
```
PO Created from Requisition
  ↓
Stage 1: Department Head Review
  ├── Can Approve → Stage 2
  └── Can Reverse → Back to Procurement Officer
  ↓
Stage 2: Auditor Review
  ├── Can Approve → Stage 3
  └── Can Reverse → Back to Procurement Officer
  ↓
Stage 3: Director Finance Review
  ├── Can Approve → Stage 4
  └── Can Reverse → Back to Procurement Officer
  ↓
Stage 4: Principal Officer Final
  ├── Can Approve → PO Final Approved
  └── Can Reverse → Back to Procurement Officer
  ↓
PO Final Approved
```

#### Updated Data Model
```typescript
interface PurchaseOrder extends WorkflowDocument {
  type: 'PURCHASE_ORDER'

  // Links
  requisitionId: string

  // Approval Stages (4 instead of 1)
  approvalStages: Array<{
    stageNumber: 1 | 2 | 3 | 4
    stageName: 'Department Head' | 'Auditor' | 'Director Finance' | 'Principal Officer'
    assignedTo: User
    status: 'PENDING' | 'APPROVED' | 'REVERSED'
    approvedAt?: Date
    approvedBy?: User
    comments?: string
    reversedAt?: Date
    reversalReason?: string
  }>

  currentStage: 1 | 2 | 3 | 4
  // ... other fields from original
}
```

#### Updated Server Actions
```typescript
// PO approval - any of 4 stages
async function approvePurchaseOrder(
  poId: string,
  stageNumber: 1 | 2 | 3 | 4,
  comments?: string
) {
  const po = documentStore.get(poId)
  const stage = po.approvalStages.find(s => s.stageNumber === stageNumber)

  // Verify current approver
  if (stage.assignedTo.id !== currentUser.id) {
    return forbiddenResponse()
  }

  // Mark stage as approved
  stage.status = 'APPROVED'
  stage.approvedAt = new Date()
  stage.approvedBy = currentUser
  stage.comments = comments

  // Move to next stage
  if (stageNumber < 4) {
    po.currentStage = stageNumber + 1
    po.status = 'IN_APPROVAL'
  } else {
    // Final approval
    po.status = 'APPROVED'
    po.approvedAt = new Date()
  }

  // Log approval
  approvalLogsStore.get(poId)?.push({
    approver: currentUser,
    action: 'APPROVED',
    timestamp: new Date(),
    stage: stageNumber,
    comments
  })

  return successResponse(po)
}

// PO reversal - any stage back to Procurement Officer
async function reversePurchaseOrder(
  poId: string,
  stageNumber: 1 | 2 | 3 | 4,
  reason: string
) {
  const po = documentStore.get(poId)
  const stage = po.approvalStages.find(s => s.stageNumber === stageNumber)

  stage.status = 'REVERSED'
  stage.reversedAt = new Date()
  stage.reversalReason = reason

  // Reset to initial state for Procurement Officer
  po.currentStage = 1  // Back to Department Head
  po.status = 'IN_APPROVAL'

  // Reset all subsequent stages
  for (let i = stageNumber; i <= 4; i++) {
    const s = po.approvalStages.find(st => st.stageNumber === i)
    if (s && s.stageNumber > stageNumber) {
      s.status = 'PENDING'
      s.approvedAt = undefined
      s.approvedBy = undefined
      s.comments = undefined
    }
  }

  // Log reversal
  approvalLogsStore.get(poId)?.push({
    approver: currentUser,
    action: 'REVERSED',
    timestamp: new Date(),
    stage: stageNumber,
    reason
  })

  return successResponse(po)
}
```

---

### Part C: Payment Voucher (UPDATED - 4 Stages + Pre-Approval)

#### Complete Flow
```
GRN Completed
  ↓
Accountant Generates Payment Voucher
  ├── Fills vendor, amount, bank info
  └── Sets initial status: GENERATED
  ↓
Stage 1: Department Head Approval
  ├── Can Approve → Stage 2
  └── Can Reverse → Back to Accountant
  ↓
Stage 2: Auditor Approval
  ├── Can Approve → Stage 3
  └── Can Reverse → Back to Accountant
  ↓
Stage 3: Director Finance Approval
  ├── Can Approve → Stage 4
  └── Can Reverse → Back to Accountant
  ↓
Stage 4: Principal Officer Final Approval
  ├── Can Approve → Generate QR & Reference → Final Approved
  └── Can Reverse → Back to Accountant
  ↓
Final Approved Voucher with QR Code & Reference Number
  ↓
Audit Log Updated & Transaction Stored
```

#### Updated Data Model
```typescript
interface PaymentVoucher extends WorkflowDocument {
  type: 'PAYMENT_VOUCHER'

  // Generated by accountant
  generatedBy: User
  generatedAt: Date

  // Approval Stages (4 stages)
  approvalStages: Array<{
    stageNumber: 1 | 2 | 3 | 4
    stageName: 'Department Head' | 'Auditor' | 'Director Finance' | 'Principal Officer'
    assignedTo: User
    status: 'PENDING' | 'APPROVED' | 'REVERSED'
    approvedAt?: Date
    approvedBy?: User
    comments?: string
    reversedAt?: Date
    reversalReason?: string
  }>

  currentStage: 1 | 2 | 3 | 4

  // Payment Details
  bankInfo: {
    accountNumber: string
    accountName: string
    bankCode: string
    bankName: string
  }

  // Final approval outputs
  qrCode?: string
  paymentReference?: string
  qrCodeData?: {
    vendorCode: string
    amount: number
    date: Date
    reference: string
  }

  // Status progression
  status: 'GENERATED' | 'IN_APPROVAL' | 'APPROVED' | 'REVERSED'

  // Audit
  auditLogId?: string
}
```

#### Updated Server Actions
```typescript
// PV approval - any of 4 stages
async function approvePaymentVoucher(
  pvId: string,
  stageNumber: 1 | 2 | 3 | 4,
  comments?: string
) {
  const pv = paymentVoucherStore.get(pvId)
  const stage = pv.approvalStages.find(s => s.stageNumber === stageNumber)

  // Verify current approver
  if (stage.assignedTo.id !== currentUser.id) {
    return forbiddenResponse()
  }

  stage.status = 'APPROVED'
  stage.approvedAt = new Date()
  stage.approvedBy = currentUser
  stage.comments = comments

  if (stageNumber < 4) {
    pv.currentStage = stageNumber + 1
    pv.status = 'IN_APPROVAL'
  } else {
    // Final approval - generate QR and reference
    pv.status = 'APPROVED'
    pv.approvedAt = new Date()

    pv.paymentReference = generatePaymentReference()
    pv.qrCodeData = {
      vendorCode: pv.vendorId,
      amount: pv.netAmount,
      date: new Date(),
      reference: pv.paymentReference
    }
    pv.qrCode = await generateQRCode(pv.qrCodeData)

    // Create audit log entry
    pv.auditLogId = createAuditLog({
      documentId: pvId,
      documentType: 'PAYMENT_VOUCHER',
      action: 'FINAL_APPROVED',
      timestamp: new Date(),
      approver: currentUser,
      details: {
        reference: pv.paymentReference,
        qrCode: pv.qrCode
      }
    })
  }

  approvalLogsStore.get(pvId)?.push({
    approver: currentUser,
    action: 'APPROVED',
    timestamp: new Date(),
    stage: stageNumber,
    comments
  })

  return successResponse(pv)
}

// PV reversal - any stage back to Accountant
async function reversePaymentVoucher(
  pvId: string,
  stageNumber: 1 | 2 | 3 | 4,
  reason: string
) {
  const pv = paymentVoucherStore.get(pvId)
  const stage = pv.approvalStages.find(s => s.stageNumber === stageNumber)

  stage.status = 'REVERSED'
  stage.reversedAt = new Date()
  stage.reversalReason = reason

  // Reset to initial stage
  pv.currentStage = 1
  pv.status = 'IN_APPROVAL'

  // Reset all subsequent stages
  for (let i = stageNumber; i <= 4; i++) {
    const s = pv.approvalStages.find(st => st.stageNumber === i)
    if (s && s.stageNumber > stageNumber) {
      s.status = 'PENDING'
      s.approvedAt = undefined
      s.approvedBy = undefined
      s.comments = undefined
    }
  }

  // Clear final approval data if it exists
  if (stageNumber === 4) {
    pv.qrCode = undefined
    pv.paymentReference = undefined
    pv.qrCodeData = undefined
    pv.approvedAt = undefined
  }

  approvalLogsStore.get(pvId)?.push({
    approver: currentUser,
    action: 'REVERSED',
    timestamp: new Date(),
    stage: stageNumber,
    reason
  })

  return successResponse(pv)
}
```

---

## Additional Features to Implement

### 1. Transaction Search & Filtering
**Server Actions**:
```typescript
async function searchTransactions(criteria: {
  dateRange?: { from: Date, to: Date }
  referenceNumber?: string
  vendorName?: string
  status?: string
  department?: string
}) {
  // Filter transactions based on criteria
  // Return paginated results
}

async function downloadTransactionPDF(pvId: string) {
  // Generate PDF with:
  // - Voucher details
  // - Approval signatures
  // - QR code
  // - Reference number
  // - All supporting docs
}

async function verifyQRCode(qrData: string) {
  // Scan QR code
  // Verify payment voucher exists
  // Return voucher details for verification
}
```

### 2. Dashboard Enhancements

**New Dashboard Components**:
- Transaction Volume by Department (chart)
- Pending Approvals Summary (table by stage)
- Average Approval Time (metrics)
- Budget vs Actual Analysis (chart)
- Access Logs & Activity History (audit trail)

### 3. User Access Management

**New Admin Features**:
- Role Assignment interface
- Add/Remove Users
- Multi-Factor Authentication Setup
- Access Logs & Activity History
- User permission verification

---

## Updated Phase 2 Timeline

### Week 2: Enhanced Purchase Order (10 hours - increased from 8)
- 4-stage approval workflow (not 1)
- Reversal handling at each stage
- Updated detail and approval UI
- Testing all approval paths

### Week 3: Enhanced Payment Voucher (20 hours - increased from 16)
- Accountant generation step
- 4-stage approval workflow (not 3)
- Reversal handling
- QR code and reference generation
- Audit log integration
- Testing all approval paths

### Week 4-5: Additional Features (15 hours - new)
- Transaction search and filtering
- PDF generation and download
- QR code verification
- Dashboard enhancements
- User access management

**New Total**: Phase 2 now estimated at **45 hours** (increased from 32)

---

## Updated Overall Timeline

| Phase | Task | Original | Updated | Change |
|-------|------|----------|---------|--------|
| 1 | Requisition | 12h | 12h | - |
| 2A | Purchase Order | 8h | 10h | +2h |
| 2B | GRN | 8h | 8h | - |
| 2C | Payment Voucher | 16h | 20h | +4h |
| 2 | Features | - | 15h | +15h |
| 3 | Notifications | 10h | 10h | - |
| 4 | Polish | 15h | 15h | - |
| **TOTAL** | | **69h** | **90h** | +21h |

**New Timeline**: 5-7 weeks (vs 4-6 weeks)

---

## Key Implementation Differences

### Approval Pattern: Reversal vs Rejection
Your flows use **"Reversal"** which is different from **"Rejection"**:

| Aspect | Rejection | Reversal |
|--------|-----------|----------|
| **Who uses it** | Current approver rejects | Current or previous approvers reverse |
| **Where it goes** | Back to creator/previous stage | Back to specific handler (Accountant/Procurement) |
| **Purpose** | Reject the entire document | Correct information/clarification |
| **Status** | REJECTED | Back to IN_APPROVAL |
| **Data loss** | Entire document may be discarded | Information preserved for correction |

**Implementation**: Use "REVERSED" status, not "REJECTED" for these multi-stage approvals

---

## Updated Approval Stage Assignments

### Purchase Order Stages
1. **Stage 1: Department Head**
   - User role: DEPARTMENT_MANAGER
   - From: Head of originating department

2. **Stage 2: Auditor**
   - User role: AUDITOR (new role needed)
   - From: Internal audit department

3. **Stage 3: Director Finance**
   - User role: DIRECTOR_FINANCE
   - From: Finance leadership

4. **Stage 4: Principal Officer**
   - User role: PRINCIPAL_OFFICER
   - From: Executive leadership

### Payment Voucher Stages
1. **Stage 1: Department Head**
   - User role: DEPARTMENT_MANAGER

2. **Stage 2: Auditor**
   - User role: AUDITOR

3. **Stage 3: Director Finance**
   - User role: DIRECTOR_FINANCE

4. **Stage 4: Principal Officer**
   - User role: PRINCIPAL_OFFICER

**New Roles Needed**:
- AUDITOR (with audit permissions)
- PRINCIPAL_OFFICER (if not already defined)
- Update DIRECTOR_FINANCE role definition

---

## Updated Mock Data Requirements

**Additional Users Needed**:
```typescript
// Auditors (2)
{
  id: 'user-aud-1',
  name: 'Robert Banda',
  email: 'robert@example.com',
  department: 'Internal Audit',
  role: 'AUDITOR'
}

// Principal Officer
{
  id: 'user-po-1',
  name: 'Margaret Phiri',
  email: 'margaret@example.com',
  department: 'Executive',
  role: 'PRINCIPAL_OFFICER'
}

// Department Heads (multiple departments)
// - Operations
// - Finance
// - HR
// - Procurement
```

---

## Critical Implementation Notes

### 1. Reversal vs Rejection
- Use "REVERSED" status for multi-stage flows
- Reversals don't delete data, just reset approval state
- Reversals go back to previous handler, not creator

### 2. QR Code Verification
- QR code must be scannable and verifiable
- Contains: vendor code, amount, date, reference
- Used for payment verification at bank/payer

### 3. Audit Trail
- Must log every stage transition
- Must log reversals with reason
- Must be immutable
- Final approved voucher creates immutable audit log entry

### 4. Dashboard Requirements
- Show transaction volume, not just pending items
- Show approval time metrics
- Show budget analysis
- Show access logs for compliance

### 5. Search Functionality
- By date range
- By reference number
- By vendor
- By department
- By status

---

## Updated Success Criteria

### By End of Week 1 (Phase 1)
- ✅ Requisition 4-stage workflow complete
- ✅ PO auto-created and linked
- ✅ Accountant role functional

### By End of Week 2 (Phase 2A - PO)
- ✅ PO 4-stage approval working
- ✅ Reversal handling at each stage
- ✅ All approvers can review and act
- ✅ PO sent back to Procurement Officer on reversal

### By End of Week 3 (Phase 2B/C - GRN + PV)
- ✅ GRN creation working
- ✅ PV auto-generated from GRN
- ✅ PV 4-stage approval working
- ✅ Reversals working with proper handling

### By End of Week 4-5 (Phase 2 Features)
- ✅ QR code generation and display
- ✅ Payment reference generation
- ✅ Transaction search working
- ✅ PDF downloads working
- ✅ Dashboard enhanced
- ✅ User access management

### By End of Week 6-7 (Phase 3)
- ✅ Notifications system
- ✅ All 4 approvers notified at their stages
- ✅ Reversal notifications
- ✅ Complete system 100% feature ready

---

## Files Needing New/Updated Roles

```typescript
// src/lib/rbac.ts - Add these roles
const AUDITOR_PERMISSIONS = [
  'approve_document',
  'reject_document',  // Actually "reverse"
  'view_audit_log',
  'view_attachments',
  'add_comments'
]

const PRINCIPAL_OFFICER_PERMISSIONS = [
  'approve_document',
  'reject_document',
  'view_audit_log',
  'view_attachments',
  'manage_approvers'
]

const DIRECTOR_FINANCE_PERMISSIONS = [
  'approve_document',
  'reject_document',
  'view_audit_log',
  'view_attachments',
  'add_comments'
]

// src/lib/mock-data.ts - Add these users
// Auditors
// Principal Officer
// Department Heads for each department
```

---

## Conclusion

Your system flows reveal a more sophisticated approval process than initially planned:

**Key Changes**:
1. **PO**: 1 stage → 4 stages with reversals
2. **PV**: 3 stages → 4 stages with reversals (plus accountant generation)
3. **New Features**: Search, QR verification, PDF download, audit logs
4. **Timeline**: 69h → 90h total (+21 hours for added complexity)

**Recommended Approach**:
1. Keep Phase 1 as-is (requisition enhancement)
2. Update Phase 2A to handle 4-stage PO approval with reversals
3. Update Phase 2C to handle 4-stage PV approval with reversals + accountant generation
4. Add Phase 2D for search, verification, and dashboard features
5. Keep Phase 3 for notifications
6. Phase 4 optional polish

---

**Status**: Implementation plan updated to match system flows
**Next**: Update MASTER_IMPLEMENTATION_PLAN.md with these changes
