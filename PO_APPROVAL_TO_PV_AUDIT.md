# Purchase Order Approval to PV Generation - Complete Audit

## Executive Summary

**Status**: ✅ **WORKFLOW FUNCTIONAL** with identified gaps  
**Date**: 2026-04-20  
**Scope**: PO submission → Approval workflow → PV generation readiness

---

## Workflow Flow Diagram

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        PO APPROVAL WORKFLOW                              │
└─────────────────────────────────────────────────────────────────────────┘

1. PO SUBMISSION
   ┌──────────────┐
   │ PO (DRAFT)   │
   └──────┬───────┘
          │ User clicks "Submit for Approval"
          │ Selects workflow
          ▼
   ┌──────────────────────────────────┐
   │ SubmitPurchaseOrder Handler      │
   │ - Validates PO exists            │
   │ - Checks status = DRAFT          │
   │ - Validates quotation gate       │
   │ - Syncs from linked REQ          │
   └──────┬───────────────────────────┘
          │
          ▼
   ┌──────────────────────────────────┐
   │ AssignWorkflowToDocumentWithID   │
   │ - Creates WorkflowAssignment     │
   │ - Status = IN_PROGRESS           │
   │ - CurrentStage = 1               │
   └──────┬───────────────────────────┘
          │
          ▼
   ┌──────────────────────────────────┐
   │ Create First WorkflowTask        │
   │ - StageNumber = 1                │
   │ - Status = PENDING               │
   │ - AssignedRole = stage.role      │
   │ - DueDate calculated             │
   └──────┬───────────────────────────┘
          │
          ▼
   ┌──────────────┐
   │ PO (PENDING) │ ← Status updated
   └──────────────┘
          │
          │ Notification sent: "approval_required"
          ▼

2. APPROVAL STAGE 1
   ┌──────────────────────────────────┐
   │ Approver 1 receives task         │
   │ - Views PO details               │
   │ - Reviews items, amounts         │
   │ - Clicks "Approve" or "Reject"   │
   └──────┬───────────────────────────┘
          │
          ▼
   ┌──────────────────────────────────┐
   │ ApproveWorkflowTask              │
   │ - Validates user has role        │
   │ - Creates StageApprovalRecord    │
   │ - Checks stage completion        │
   └──────┬───────────────────────────┘
          │
          ├─ Stage Complete? ─┐
          │                   │
          ▼ YES               ▼ NO
   ┌──────────────┐    ┌──────────────┐
   │ Mark task    │    │ Wait for     │
   │ COMPLETED    │    │ other        │
   └──────┬───────┘    │ approvers    │
          │            └──────────────┘
          │
          ├─ Last Stage? ─┐
          │               │
          ▼ NO            ▼ YES
   ┌──────────────┐    ┌──────────────────┐
   │ Create next  │    │ Workflow         │
   │ stage task   │    │ COMPLETED        │
   │ Stage 2      │    └──────┬───────────┘
   └──────────────┘           │
                              ▼
                       ┌──────────────────┐
                       │ PO (APPROVED)    │ ← Final status
                       └──────┬───────────┘
                              │
                              ▼

3. POST-APPROVAL AUTOMATION
   ┌──────────────────────────────────┐
   │ triggerPostApprovalAutomation    │
   │ - Check AutoCreateGRNFromPO flag │
   └──────┬───────────────────────────┘
          │
          ├─ Automation Enabled? ─┐
          │                        │
          ▼ YES                    ▼ NO
   ┌──────────────────┐      ┌──────────────┐
   │ Create GRN       │      │ Manual GRN   │
   │ - Status: DRAFT  │      │ creation     │
   │ - Link to PO     │      └──────────────┘
   └──────┬───────────┘
          │
          ▼
   ┌──────────────────┐
   │ GRN (DRAFT)      │
   └──────┬───────────┘
          │
          │ Finance team submits GRN for approval
          ▼
   ┌──────────────────┐
   │ GRN Workflow     │
   │ (same as PO)     │
   └──────┬───────────┘
          │
          ▼
   ┌──────────────────┐
   │ GRN (APPROVED)   │
   └──────┬───────────┘
          │
          ▼

4. PV GENERATION (GOODS-FIRST FLOW)
   ┌──────────────────────────────────┐
   │ CreatePaymentVoucherFromPO       │
   │ - Requires LinkedGRNDocumentNumber│
   │ - Validates GRN status = APPROVED│
   │ - Validates GRN belongs to PO    │
   └──────┬───────────────────────────┘
          │
          ▼
   ┌──────────────────────────────────┐
   │ Create PV                        │
   │ - Status: DRAFT                  │
   │ - LinkedPO: PO doc number        │
   │ - LinkedGRN: GRN doc number      │
   │ - Amount: PO.TotalAmount         │
   │ - Currency: PO.Currency          │
   │ - Vendor: PO.VendorID            │
   └──────┬───────────────────────────┘
          │
          ▼
   ┌──────────────────┐
   │ PV (DRAFT)       │ ← Ready for finance approval
   └──────────────────┘
          │
          │ Finance team submits PV for approval
          ▼
   ┌──────────────────┐
   │ PV Workflow      │
   │ (same as PO)     │
   └──────┬───────────┘
          │
          ▼
   ┌──────────────────┐
   │ PV (APPROVED)    │ ← Ready for payment
   └──────────────────┘

5. PV GENERATION (PAYMENT-FIRST FLOW)
   ┌──────────────────────────────────┐
   │ CreatePaymentVoucherFromPO       │
   │ - No GRN required                │
   │ - Validates PO status = APPROVED │
   └──────┬───────────────────────────┘
          │
          ▼
   ┌──────────────────────────────────┐
   │ Create PV                        │
   │ - Status: DRAFT                  │
   │ - LinkedPO: PO doc number        │
   │ - LinkedGRN: empty               │
   └──────┬───────────────────────────┘
          │
          ▼
   ┌──────────────────┐
   │ PV (DRAFT)       │ ← Ready for finance approval
   └──────────────────┘
```

---

## Component Analysis

### 1. PO Submission ✅ WORKING

**Handler**: `SubmitPurchaseOrder` (backend/handlers/purchase_order.go:826)

**Validations**:

- ✅ PO exists and belongs to organization
- ✅ PO status is DRAFT
- ✅ Quotation gate: 3 quotations required (unless bypassed)
- ✅ Linked REQ must be APPROVED (if exists)
- ✅ Syncs items/amounts from approved REQ
- ✅ WorkflowID provided in request

**Actions**:

- ✅ Assigns workflow to PO
- ✅ Creates first workflow task
- ✅ Updates PO status to PENDING
- ✅ Adds action history entry
- ✅ Sends notification

**Status**: ✅ **FULLY FUNCTIONAL**

---

### 2. Workflow Assignment ✅ WORKING

**Service**: `WorkflowExecutionService.AssignWorkflowToDocumentWithID`

**Process**:

1. ✅ Validates workflow exists and is active
2. ✅ Validates entity type matches workflow
3. ✅ Creates `WorkflowAssignment` record
   - Status: IN_PROGRESS
   - CurrentStage: 1
   - WorkflowVersion: captured
4. ✅ Creates first `WorkflowTask`
   - StageNumber: 1
   - Status: PENDING
   - AssignedRole: from workflow stage definition
   - DueDate: calculated from timeout or default 7 days
5. ✅ Sends "approval_required" notification

**Status**: ✅ **FULLY FUNCTIONAL**

---

### 3. Approval Task Handling ✅ WORKING

#### Approval Flow

**Handler**: `ApproveTask` (backend/handlers/approval_handler.go:754)  
**Service**: `WorkflowExecutionService.ApproveWorkflowTask`

**Process**:

1. ✅ Validates task exists and is PENDING/CLAIMED
2. ✅ Validates user has required role
3. ✅ Creates `StageApprovalRecord` with signature
4. ✅ Checks stage completion criteria
5. ✅ If stage complete:
   - ✅ Marks task COMPLETED
   - ✅ Adds stage execution to history
   - ✅ Checks if last stage
6. ✅ If last stage:
   - ✅ Marks workflow COMPLETED
   - ✅ Updates PO status to APPROVED
   - ✅ Triggers post-approval automation
7. ✅ If not last stage:
   - ✅ Creates next stage task
   - ✅ Sends notification to next approver

**Optimistic Locking**: ✅ Version field prevents concurrent modifications

**Status**: ✅ **FULLY FUNCTIONAL**

#### Rejection Flow

**Handler**: `RejectTask` (backend/handlers/approval_handler.go:816)  
**Service**: `WorkflowExecutionService.RejectWorkflowTask`

**Rejection Types**:

1. ✅ **Full Rejection**: PO → DRAFT, workflow terminated
2. ✅ **Return to Previous Stage**: PO → REVISION, new task at prev stage
3. ✅ **Return to Draft**: PO → DRAFT, workflow cancelled

**Cascade Logic**:

- ✅ PO rejection → Linked REQ reverted to DRAFT
- ✅ PO rejection → Linked REQ workflow cancelled
- ✅ Action history entries created

**Status**: ✅ **FULLY FUNCTIONAL**

---

### 4. PO Status Transitions ✅ WORKING

**Service**: `WorkflowExecutionService.updateDocumentStatus`

**Status Flow**:

```
DRAFT → PENDING → APPROVED
  ↓       ↓         ↓
  ←─ REJECTED ─────┘
  ↓
  ←─ REVISION (return to prev stage)
```

**Implementation**:

- ✅ Direct SQL update to `purchase_orders.status`
- ✅ Triggers `utils.SyncDocument()` to update generic documents index
- ✅ Action history entries added

**Status**: ✅ **FULLY FUNCTIONAL**

---

### 5. Post-Approval Automation ⚠️ PARTIAL

**Service**: `WorkflowExecutionService.triggerPostApprovalAutomation`

**For Purchase Orders**:

- ✅ Checks `AutoCreateGRNFromPO` flag
- ✅ If enabled: Creates GRN from PO
- ✅ GRN created with status = DRAFT
- ✅ Updates PO with `auto_created_grn` metadata

**Issues Identified**:

1. ⚠️ **No automatic GRN workflow assignment**
   - GRN created but not submitted to workflow
   - Finance team must manually submit GRN
2. ⚠️ **No direct PV auto-creation from PO**
   - PV only auto-created from GRN (goods-first)
   - Payment-first flow requires manual PV creation

**Status**: ⚠️ **PARTIALLY FUNCTIONAL** - Manual steps required

---

### 6. PV Generation from Approved PO ✅ WORKING

#### Goods-First Flow

**Handler**: `CreatePaymentVoucherFromPO` (backend/handlers/document_extras_handler.go:282)

**Requirements**:

- ✅ PO must be APPROVED
- ✅ LinkedGRNDocumentNumber required
- ✅ GRN must be APPROVED
- ✅ GRN must belong to the PO
- ✅ One-to-one constraint: Only one PV per PO

**Process**:

1. ✅ Validates PO exists and is APPROVED
2. ✅ Validates GRN exists and is APPROVED
3. ✅ Validates GRN.PODocumentNumber matches PO
4. ✅ Checks no existing PV for this PO
5. ✅ Creates PV with:
   - Status: DRAFT
   - LinkedPO: PO document number
   - LinkedGRN: GRN document number
   - Amount: PO.TotalAmount
   - Currency: PO.Currency (inherited)
   - Vendor: PO.VendorID
   - Budget fields: BudgetCode, CostCenter, ProjectCode
6. ✅ Adds action history entries
7. ✅ Logs audit event

**Status**: ✅ **FULLY FUNCTIONAL**

#### Payment-First Flow

**Handler**: Same as above, but no GRN required

**Process**:

- ✅ Validates PO is APPROVED
- ✅ Creates PV without GRN link
- ✅ Status: DRAFT

**Status**: ✅ **FULLY FUNCTIONAL**

---

## Identified Gaps & Recommendations

### 🔴 **Gap 1: No Automatic GRN Workflow Assignment**

**Issue**: When GRN is auto-created from approved PO, it's created in DRAFT status but not automatically submitted to workflow.

**Impact**: Finance team must manually:

1. Find the auto-created GRN
2. Submit it for approval
3. Wait for approval before creating PV

**Recommendation**: Add automatic workflow assignment for auto-created GRN

**Fix Location**: `backend/services/document_automation_service.go`

```go
// After creating GRN, automatically assign workflow
if result.Success && result.DocumentID != "" {
    // Get default workflow for GRN
    defaultWorkflow, err := s.workflowService.GetDefaultWorkflow(ctx, organizationID, "grn")
    if err == nil && defaultWorkflow != nil {
        // Assign workflow to GRN
        _, err = s.workflowExecutionService.AssignWorkflowToDocumentWithID(
            ctx, organizationID, result.DocumentID, "grn",
            defaultWorkflow.ID.String(), systemUserID,
        )
        if err != nil {
            log.Printf("Failed to auto-assign workflow to GRN: %v", err)
        }
    }
}
```

---

### 🟡 **Gap 2: No Direct PV Auto-Creation from PO (Payment-First)**

**Issue**: Payment-first flow requires manual PV creation even when PO is approved.

**Impact**: Finance team must manually create PV from approved PO.

**Recommendation**: Add automation flag for payment-first PV creation

**Fix Location**: `backend/services/workflow_execution_service.go`

```go
case "PURCHASE_ORDER", "purchase_order":
    // Check procurement flow
    var po models.PurchaseOrder
    if err := s.db.Where("id = ?", entityID).First(&po).Error; err == nil {
        effectiveFlow := po.ProcurementFlow
        if effectiveFlow == "" {
            // Get org default
            orgSvc := services.NewOrganizationService(s.db)
            orgSettings, _ := orgSvc.GetOrganizationSettings(po.OrganizationID)
            if orgSettings != nil {
                effectiveFlow = orgSettings.ProcurementFlow
            } else {
                effectiveFlow = "goods_first"
            }
        }

        if effectiveFlow == "payment_first" && config.AutoCreatePVFromPO {
            // Auto-create PV directly from PO (payment-first flow)
            result, err := s.automationService.CreatePaymentVoucherFromPO(ctx, &po, config)
            if err == nil && result.Success {
                // Update PO with auto-created PV info
                autoCreatedPV := map[string]interface{}{
                    "id": result.DocumentID,
                    "created": true,
                }
                autoCreatedJSON, _ := datatypes.NewJSONType(autoCreatedPV).MarshalJSON()
                s.db.Model(&po).Updates(map[string]interface{}{
                    "automation_used": true,
                    "auto_created_pv": datatypes.JSON(autoCreatedJSON),
                })
            }
        } else if effectiveFlow == "goods_first" && config.AutoCreateGRNFromPO {
            // Existing goods-first logic
            // ...
        }
    }
```

---

### 🟡 **Gap 3: No Automatic PV Workflow Assignment**

**Issue**: PV created in DRAFT status but not automatically submitted to workflow.

**Impact**: Finance team must manually submit PV for approval.

**Recommendation**: Add automatic workflow assignment for created PV

**Fix Location**: `backend/handlers/document_extras_handler.go`

```go
// After creating PV, optionally auto-submit to workflow
if req.WorkflowID != "" {
    workflowExecutionService := c.Locals("workflowExecutionService").(*services.WorkflowExecutionService)
    _, err := workflowExecutionService.AssignWorkflowToDocumentWithID(
        c.Context(), tenant.OrganizationID, voucher.ID, "payment_voucher",
        req.WorkflowID, tenant.UserID,
    )
    if err != nil {
        log.Printf("Failed to auto-assign workflow to PV: %v", err)
        // Don't fail PV creation, just log the error
    } else {
        // Update PV status to PENDING
        voucher.Status = "PENDING"
        config.DB.Save(&voucher)
    }
}
```

---

### 🟢 **Gap 4: Missing Audit Trail for Status Changes**

**Issue**: `updateDocumentStatus()` updates status but doesn't call audit service.

**Impact**: Status changes not logged in audit trail (only in action history).

**Recommendation**: Add audit logging for status changes

**Fix Location**: `backend/services/workflow_execution_service.go`

```go
func (s *WorkflowExecutionService) updateDocumentStatus(tx *gorm.DB, entityType, entityID, newStatus string) error {
    // Get old status first
    var oldStatus string
    switch entityType {
    case "PURCHASE_ORDER", "purchase_order":
        var po models.PurchaseOrder
        if err := tx.Where("id = ?", entityID).First(&po).Error; err == nil {
            oldStatus = po.Status
        }
        err := tx.Model(&models.PurchaseOrder{}).Where("id = ?", entityID).Update("status", newStatus).Error
        if err != nil {
            return err
        }
    // ... other cases
    }

    // Log audit event
    if s.auditService != nil {
        go s.auditService.LogEvent(context.Background(), "", "",
            "document_status_changed", entityType, entityID,
            fmt.Sprintf("Status changed from %s to %s", oldStatus, newStatus),
            oldStatus, newStatus,
        )
    }

    // Keep the generic documents index in sync
    go utils.SyncDocument(s.db, entityType, entityID)
    return nil
}
```

---

### 🟢 **Gap 5: No Approval Deadline Enforcement**

**Issue**: DueDate calculated but no automatic escalation/rejection on expiry.

**Impact**: Tasks can remain pending indefinitely.

**Recommendation**: Add background worker for deadline enforcement

**Fix Location**: Create new file `backend/workers/approval_deadline_worker.go`

```go
func StartApprovalDeadlineWorker(db *gorm.DB, interval time.Duration) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()

    for range ticker.C {
        // Find overdue tasks
        var overdueTasks []models.WorkflowTask
        db.Where("status IN ('PENDING', 'CLAIMED') AND due_date < ?", time.Now()).
            Find(&overdueTasks)

        for _, task := range overdueTasks {
            // Send escalation notification
            // Or auto-reject based on configuration
            log.Printf("Task %s is overdue (due: %v)", task.ID, task.DueDate)
        }
    }
}
```

---

## Testing Checklist

### End-to-End PO Approval Flow

#### Test 1: Single-Stage Approval

```
1. Create PO in DRAFT status
2. Submit for approval with single-stage workflow
3. Approver approves task
4. Verify:
   ✅ PO status = APPROVED
   ✅ WorkflowAssignment status = COMPLETED
   ✅ WorkflowTask status = COMPLETED
   ✅ StageApprovalRecord created
   ✅ Action history entries added
   ✅ Notification sent
```

#### Test 2: Multi-Stage Approval

```
1. Create PO in DRAFT status
2. Submit for approval with 3-stage workflow
3. Stage 1 approver approves
4. Verify:
   ✅ Stage 1 task = COMPLETED
   ✅ Stage 2 task = PENDING
   ✅ PO status = PENDING (not APPROVED yet)
5. Stage 2 approver approves
6. Verify:
   ✅ Stage 2 task = COMPLETED
   ✅ Stage 3 task = PENDING
7. Stage 3 approver approves
8. Verify:
   ✅ PO status = APPROVED
   ✅ WorkflowAssignment status = COMPLETED
```

#### Test 3: Rejection (Full)

```
1. Create PO in DRAFT status
2. Submit for approval
3. Approver rejects (full rejection)
4. Verify:
   ✅ PO status = DRAFT
   ✅ WorkflowAssignment status = REJECTED
   ✅ WorkflowTask status = REJECTED
   ✅ Linked REQ status = DRAFT (if exists)
   ✅ Action history entries added
```

#### Test 4: Return to Previous Stage

```
1. Create PO in DRAFT status
2. Submit for approval with 2-stage workflow
3. Stage 1 approver approves
4. Stage 2 approver returns to previous stage
5. Verify:
   ✅ PO status = REVISION
   ✅ New task created at Stage 1
   ✅ WorkflowAssignment.CurrentStage = 1
   ✅ Action history entry: "RETURNED_FOR_REVISION"
```

#### Test 5: PV Generation (Goods-First)

```
1. Create and approve PO
2. Create and approve GRN linked to PO
3. Create PV from PO with LinkedGRNDocumentNumber
4. Verify:
   ✅ PV created with status = DRAFT
   ✅ PV.LinkedPO = PO document number
   ✅ PV.LinkedGRN = GRN document number
   ✅ PV.Amount = PO.TotalAmount
   ✅ PV.Currency = PO.Currency
   ✅ Action history entries added
```

#### Test 6: PV Generation (Payment-First)

```
1. Create and approve PO
2. Create PV from PO without GRN
3. Verify:
   ✅ PV created with status = DRAFT
   ✅ PV.LinkedPO = PO document number
   ✅ PV.LinkedGRN = empty
   ✅ PV.Amount = PO.TotalAmount
```

---

## SQL Verification Queries

### Check PO Workflow Status

```sql
SELECT
    po.id,
    po.document_number,
    po.status,
    wa.status as workflow_status,
    wa.current_stage,
    wt.stage_number,
    wt.stage_name,
    wt.status as task_status,
    wt.assigned_role
FROM purchase_orders po
LEFT JOIN workflow_assignments wa ON wa.entity_id = po.id AND wa.entity_type = 'purchase_order'
LEFT JOIN workflow_tasks wt ON wt.workflow_assignment_id = wa.id AND wt.status IN ('PENDING', 'CLAIMED')
WHERE po.document_number = 'PO-XXXX-XXX'
ORDER BY wt.stage_number;
```

### Check Approval History

```sql
SELECT
    sar.stage_number,
    sar.approver_name,
    sar.approver_role,
    sar.action,
    sar.comments,
    sar.approved_at
FROM stage_approval_records sar
JOIN workflow_tasks wt ON wt.id = sar.workflow_task_id
JOIN workflow_assignments wa ON wa.id = wt.workflow_assignment_id
WHERE wa.entity_id = 'po-id-here'
ORDER BY sar.approved_at;
```

### Check PV Creation from PO

```sql
SELECT
    pv.id,
    pv.document_number,
    pv.status,
    pv.linked_po,
    pv.linked_grn,
    pv.amount,
    pv.currency,
    po.document_number as po_doc,
    po.status as po_status
FROM payment_vouchers pv
JOIN purchase_orders po ON po.document_number = pv.linked_po
WHERE pv.linked_po = 'PO-XXXX-XXX';
```

---

## Conclusion

### ✅ **Working Components**

1. PO submission and workflow assignment
2. Multi-stage approval workflow
3. Approval task handling (approve/reject)
4. PO status transitions
5. PV generation from approved PO (both flows)
6. Cascade rejection logic

### ⚠️ **Gaps Requiring Fixes**

1. No automatic GRN workflow assignment (Priority: HIGH)
2. No direct PV auto-creation for payment-first (Priority: MEDIUM)
3. No automatic PV workflow assignment (Priority: MEDIUM)
4. Missing audit trail for status changes (Priority: LOW)
5. No approval deadline enforcement (Priority: LOW)

### 📊 **Overall Assessment**

**Confidence Level**: 90%  
**Workflow Status**: ✅ **FUNCTIONAL** with manual steps  
**PV Generation**: ✅ **READY** after PO approval

The workflow is fully functional for the core approval flow. PV generation works correctly for both goods-first and payment-first flows. The identified gaps are primarily around automation and convenience features that would reduce manual steps for finance teams.

**Recommendation**: Apply Gap 1 fix (automatic GRN workflow assignment) to streamline the goods-first flow. Other gaps can be addressed in future iterations based on user feedback.
