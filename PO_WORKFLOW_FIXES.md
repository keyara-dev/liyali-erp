# Purchase Order Workflow - Implementation Fixes

## Overview

This document outlines the fixes needed to ensure smooth PO approval workflow and PV generation readiness.

---

## Fix 1: Add Automatic GRN Workflow Assignment (HIGH PRIORITY)

### Problem

When GRN is auto-created from approved PO, it's created in DRAFT status but not automatically submitted to workflow. Finance team must manually find and submit the GRN.

### Solution

Modify `CreateGRNFromPurchaseOrder` to optionally auto-assign workflow after GRN creation.

### Implementation

**File**: `backend/services/document_automation_service.go`

**Add to AutomationConfig struct**:

```go
type AutomationConfig struct {
    AutoCreatePOFromRequisition bool
    AutoCreateGRNFromPO         bool
    AutoCreatePVFromGRN         bool
    RequireApprovalForAuto      bool
    AutoSubmitGRNToWorkflow     bool  // NEW: Auto-submit created GRN to workflow
    AutoSubmitPVToWorkflow      bool  // NEW: Auto-submit created PV to workflow
}
```

**Modify CreateGRNFromPurchaseOrder method** (after GRN creation):

```go
// After successful GRN creation, before returning result:

// Auto-submit GRN to workflow if configured
if config.AutoSubmitGRNToWorkflow && s.workflowService != nil {
    // Get default workflow for GRN
    defaultWorkflow, err := s.workflowService.GetDefaultWorkflow(ctx, grn.OrganizationID, "grn")
    if err == nil && defaultWorkflow != nil && defaultWorkflow.IsActive {
        // Assign workflow to GRN
        assignment, err := s.workflowExecutionService.AssignWorkflowToDocumentWithID(
            ctx,
            grn.OrganizationID,
            grn.ID,
            "grn",
            defaultWorkflow.ID.String(),
            "system", // System user for automation
        )
        if err != nil {
            fmt.Printf("Warning: Failed to auto-assign workflow to GRN %s: %v\n", grn.DocumentNumber, err)
            // Don't fail GRN creation, just log the error
        } else {
            // Update GRN status to PENDING
            s.db.Model(&grn).Updates(map[string]interface{}{
                "status":     "PENDING",
                "updated_at": time.Now(),
            })

            // Log audit event
            if s.auditService != nil {
                s.auditService.LogEvent(ctx, "system", grn.OrganizationID,
                    "grn_auto_submitted", "grn", grn.ID,
                    fmt.Sprintf("GRN %s automatically submitted to workflow %s",
                        grn.DocumentNumber, defaultWorkflow.Name),
                    "DRAFT", "PENDING",
                )
            }

            fmt.Printf("[Automation] GRN %s auto-submitted to workflow (assignment: %s)\n",
                grn.DocumentNumber, assignment.ID)
        }
    }
}
```

**Add required service dependencies**:

```go
type DocumentAutomationService struct {
    db                       *gorm.DB
    auditService             *AuditService
    notificationSvc          *NotificationService
    workflowService          *WorkflowService          // NEW
    workflowExecutionService *WorkflowExecutionService // NEW
}

func NewDocumentAutomationService(
    db *gorm.DB,
    auditService *AuditService,
    notificationSvc *NotificationService,
    workflowService *WorkflowService,          // NEW
    workflowExecutionService *WorkflowExecutionService, // NEW
) *DocumentAutomationService {
    return &DocumentAutomationService{
        db:                       db,
        auditService:             auditService,
        notificationSvc:          notificationSvc,
        workflowService:          workflowService,
        workflowExecutionService: workflowExecutionService,
    }
}
```

**Update service initialization** in `main.go`:

```go
// Initialize document automation service with workflow services
documentAutomationService := services.NewDocumentAutomationService(
    config.DB,
    auditService,
    notificationService,
    workflowService,          // NEW
    workflowExecutionService, // NEW
)
```

**Update GetDefaultAutomationConfig**:

```go
func (s *DocumentAutomationService) GetDefaultAutomationConfig() AutomationConfig {
    return AutomationConfig{
        AutoCreatePOFromRequisition: true,
        AutoCreateGRNFromPO:         true,
        AutoCreatePVFromGRN:         true,
        RequireApprovalForAuto:      false,
        AutoSubmitGRNToWorkflow:     true,  // NEW: Enable by default
        AutoSubmitPVToWorkflow:      false, // NEW: Disabled by default (manual review)
    }
}
```

---

## Fix 2: Add Automatic PV Workflow Assignment (MEDIUM PRIORITY)

### Problem

PV created in DRAFT status but not automatically submitted to workflow. Finance team must manually submit PV.

### Solution

Modify `CreatePaymentVoucherFromPO` handler to accept optional workflowId and auto-submit.

### Implementation

**File**: `backend/handlers/document_extras_handler.go`

**Modify CreatePaymentVoucherFromPO handler** (after PV creation):

```go
// After successful PV creation, before returning response:

// Auto-submit PV to workflow if workflowId provided
if req.WorkflowID != "" {
    workflowExecutionService := c.Locals("workflowExecutionService").(*services.WorkflowExecutionService)

    assignment, err := workflowExecutionService.AssignWorkflowToDocumentWithID(
        c.Context(),
        tenant.OrganizationID,
        voucher.ID,
        "payment_voucher",
        req.WorkflowID,
        tenant.UserID,
    )

    if err != nil {
        logging.LogError(c, err, "pv_workflow_assignment_failed", map[string]interface{}{
            "pv_id":       voucher.ID,
            "workflow_id": req.WorkflowID,
        })
        // Don't fail PV creation, just log the error
        // Return PV in DRAFT status for manual submission
    } else {
        // Update PV status to PENDING
        voucher.Status = "PENDING"
        voucher.UpdatedAt = time.Now()
        config.DB.Save(&voucher)

        // Add action history entry
        pvSubmitHistory := voucher.ActionHistory.Data()
        pvSubmitHistory = append(pvSubmitHistory, types.ActionHistoryEntry{
            ID:              uuid.New().String(),
            Action:          "SUBMIT",
            ActionType:      "SUBMIT",
            PerformedBy:     tenant.UserID,
            PerformedByName: pvFromPOUser.Name,
            PerformedByRole: pvFromPOUser.Role,
            Timestamp:       time.Now(),
            PerformedAt:     time.Now(),
            Comments:        "Payment voucher auto-submitted for approval",
            PreviousStatus:  "DRAFT",
            NewStatus:       "PENDING",
        })
        voucher.ActionHistory = datatypes.NewJSONType(pvSubmitHistory)
        config.DB.Save(&voucher)

        go services.LogDocumentEvent(config.DB, services.DocumentEvent{
            OrganizationID: tenant.OrganizationID,
            DocumentID:     voucher.ID,
            DocumentType:   "payment_voucher",
            UserID:         tenant.UserID,
            ActorName:      pvFromPOUser.Name,
            ActorRole:      pvFromPOUser.Role,
            Action:         "submitted",
            Details:        map[string]interface{}{
                "documentNumber": voucher.DocumentNumber,
                "workflowId":     req.WorkflowID,
                "assignmentId":   assignment.ID,
            },
        })

        logger.Info("pv_auto_submitted_to_workflow")
    }
}
```

**Frontend Update**: Add workflowId to PV creation request

**File**: `frontend/src/types/payment-voucher.ts`

```typescript
export interface CreatePaymentVoucherFromPORequest {
  purchaseOrderId: string;
  purchaseOrderDocumentNumber: string;
  // ... existing fields ...
  workflowId?: string; // NEW: Optional workflow ID for auto-submission
  linkedGRNDocumentNumber?: string;
}
```

**File**: `frontend/src/components/create-pv-from-po-dialog.tsx` (or similar)

```typescript
// Add workflow selector to the form
const [selectedWorkflowId, setSelectedWorkflowId] = useState<string>("");

// In the submit handler:
const handleSubmit = async () => {
  await createPaymentVoucherFromPO({
    purchaseOrderId: po.id,
    purchaseOrderDocumentNumber: po.documentNumber,
    // ... other fields ...
    workflowId: selectedWorkflowId, // NEW: Pass selected workflow
  });
};
```

---

## Fix 3: Add Audit Logging for Status Changes (LOW PRIORITY)

### Problem

Status changes not logged in audit trail (only in action history).

### Solution

Add audit service calls in `updateDocumentStatus` method.

### Implementation

**File**: `backend/services/workflow_execution_service.go`

**Modify updateDocumentStatus method**:

```go
func (s *WorkflowExecutionService) updateDocumentStatus(tx *gorm.DB, entityType, entityID, newStatus string) error {
    var oldStatus string
    var err error

    switch entityType {
    case "REQUISITION", "requisition":
        var req models.Requisition
        if err := tx.Where("id = ?", entityID).First(&req).Error; err == nil {
            oldStatus = req.Status
        }
        err = tx.Model(&models.Requisition{}).Where("id = ?", entityID).Update("status", newStatus).Error

    case "BUDGET", "budget":
        var budget models.Budget
        if err := tx.Where("id = ?", entityID).First(&budget).Error; err == nil {
            oldStatus = budget.Status
        }
        err = tx.Model(&models.Budget{}).Where("id = ?", entityID).Update("status", newStatus).Error

    case "PURCHASE_ORDER", "purchase_order":
        var po models.PurchaseOrder
        if err := tx.Where("id = ?", entityID).First(&po).Error; err == nil {
            oldStatus = po.Status
        }
        err = tx.Model(&models.PurchaseOrder{}).Where("id = ?", entityID).Update("status", newStatus).Error

    case "PAYMENT_VOUCHER", "payment_voucher":
        var pv models.PaymentVoucher
        if err := tx.Where("id = ?", entityID).First(&pv).Error; err == nil {
            oldStatus = pv.Status
        }
        err = tx.Model(&models.PaymentVoucher{}).Where("id = ?", entityID).Update("status", newStatus).Error

    case "GRN", "grn":
        var grn models.GoodsReceivedNote
        if err := tx.Where("id = ?", entityID).First(&grn).Error; err == nil {
            oldStatus = grn.Status
        }
        err = tx.Model(&models.GoodsReceivedNote{}).Where("id = ?", entityID).Update("status", newStatus).Error

    default:
        return fmt.Errorf("unsupported entity type: %s", entityType)
    }

    if err != nil {
        return err
    }

    // Log audit event for status change
    if s.auditService != nil && oldStatus != newStatus {
        go func(eType, eID, old, new string) {
            ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
            defer cancel()

            s.auditService.LogEvent(ctx, "system", "",
                "document_status_changed",
                strings.ToLower(eType),
                eID,
                fmt.Sprintf("Status changed from %s to %s via workflow", old, new),
                old,
                new,
            )
        }(entityType, entityID, oldStatus, newStatus)
    }

    // Keep the generic documents index in sync after every status change
    go utils.SyncDocument(s.db, entityType, entityID)
    return nil
}
```

---

## Fix 4: Add Payment-First PV Auto-Creation (MEDIUM PRIORITY)

### Problem

Payment-first flow requires manual PV creation even when PO is approved.

### Solution

Add automation trigger for payment-first PV creation in post-approval automation.

### Implementation

**File**: `backend/services/workflow_execution_service.go`

**Modify triggerPostApprovalAutomation method** (PURCHASE_ORDER case):

```go
case "PURCHASE_ORDER", "purchase_order":
    // Get the approved purchase order
    var po models.PurchaseOrder
    if err := s.db.Where("id = ?", entityID).First(&po).Error; err != nil {
        return fmt.Errorf("failed to get purchase order: %w", err)
    }

    // Determine effective procurement flow
    effectiveFlow := po.ProcurementFlow
    if effectiveFlow == "" {
        // Get organization default
        orgSvc := NewOrganizationService(s.db)
        orgSettings, _ := orgSvc.GetOrganizationSettings(po.OrganizationID)
        if orgSettings != nil && orgSettings.ProcurementFlow != "" {
            effectiveFlow = orgSettings.ProcurementFlow
        } else {
            effectiveFlow = "goods_first" // Default
        }
    }

    // Handle based on procurement flow
    if effectiveFlow == "payment_first" {
        // Payment-first: Auto-create PV directly from PO (no GRN required)
        if config.AutoCreatePVFromPO { // NEW config flag
            result, err := s.automationService.CreatePaymentVoucherFromPO(ctx, &po, config)
            if err != nil {
                return fmt.Errorf("failed to create payment voucher: %w", err)
            }

            if !result.Success {
                return fmt.Errorf("payment voucher creation failed: %s", result.Error)
            }

            // Update PO with auto-created PV info
            autoCreatedPV := map[string]interface{}{
                "id":      result.DocumentID,
                "created": true,
            }

            if result.CreatedDocument != nil {
                if pv, ok := result.CreatedDocument.(models.PaymentVoucher); ok {
                    autoCreatedPV["documentNumber"] = pv.DocumentNumber
                    autoCreatedPV["amount"] = pv.Amount
                }
            }

            autoCreatedJSON, _ := datatypes.NewJSONType(autoCreatedPV).MarshalJSON()
            s.db.Model(&po).Updates(map[string]interface{}{
                "automation_used": true,
                "auto_created_pv": datatypes.JSON(autoCreatedJSON),
            })
        }
    } else {
        // Goods-first: Auto-create GRN from PO (existing logic)
        if config.AutoCreateGRNFromPO {
            // Validate automation prerequisites
            if err := s.automationService.ValidateAutomationPrerequisites("purchase_order", &po); err != nil {
                return fmt.Errorf("automation prerequisites not met: %w", err)
            }

            // Create GRN
            result, err := s.automationService.CreateGRNFromPurchaseOrder(ctx, &po, config)
            if err != nil {
                return fmt.Errorf("failed to create GRN: %w", err)
            }

            if !result.Success {
                return fmt.Errorf("GRN creation failed: %s", result.Error)
            }

            // Update PO with auto-created GRN info
            autoCreatedGRN := map[string]interface{}{
                "id":      result.DocumentID,
                "created": true,
            }

            if result.CreatedDocument != nil {
                if grn, ok := result.CreatedDocument.(*models.GoodsReceivedNote); ok {
                    autoCreatedGRN["documentNumber"] = grn.DocumentNumber
                }
            }

            autoCreatedJSON, _ := datatypes.NewJSONType(autoCreatedGRN).MarshalJSON()
            s.db.Model(&po).Updates(map[string]interface{}{
                "automation_used":  true,
                "auto_created_grn": datatypes.JSON(autoCreatedJSON),
            })
        }
    }
```

**Add new config flag**:

```go
type AutomationConfig struct {
    AutoCreatePOFromRequisition bool
    AutoCreateGRNFromPO         bool
    AutoCreatePVFromGRN         bool
    AutoCreatePVFromPO          bool  // NEW: For payment-first flow
    RequireApprovalForAuto      bool
    AutoSubmitGRNToWorkflow     bool
    AutoSubmitPVToWorkflow      bool
}
```

**Add CreatePaymentVoucherFromPO method to DocumentAutomationService**:

```go
// CreatePaymentVoucherFromPO automatically creates a PV from an approved PO (payment-first flow)
func (s *DocumentAutomationService) CreatePaymentVoucherFromPO(
    ctx context.Context,
    purchaseOrder *models.PurchaseOrder,
    config AutomationConfig,
) (*AutomationResult, error) {
    if !config.AutoCreatePVFromPO {
        return &AutomationResult{
            Success: false,
            Error:   fmt.Errorf("automatic PV creation from PO is disabled"),
        }, nil
    }

    if strings.ToUpper(purchaseOrder.Status) != "APPROVED" {
        return &AutomationResult{
            Success: false,
            Error:   fmt.Errorf("purchase order must be approved to create payment voucher"),
        }, nil
    }

    // Check if PV already exists for this PO
    var existingPV models.PaymentVoucher
    if err := s.db.
        Where("linked_po = ? AND organization_id = ? AND UPPER(status) != 'CANCELLED'",
            purchaseOrder.DocumentNumber, purchaseOrder.OrganizationID).
        First(&existingPV).Error; err == nil {
        return &AutomationResult{
            Success: false,
            Error: fmt.Errorf("payment voucher %s already exists for PO %s",
                existingPV.DocumentNumber, purchaseOrder.DocumentNumber),
        }, nil
    }

    // Generate PV document number
    documentNumber := utils.GenerateDocumentNumber("PV")
    invoiceRef := "INV-" + purchaseOrder.DocumentNumber

    // Create Payment Voucher
    paymentVoucher := models.PaymentVoucher{
        ID:             uuid.New().String(),
        DocumentNumber: documentNumber,
        VendorID:       purchaseOrder.VendorID,
        InvoiceNumber:  invoiceRef,
        Status:         "DRAFT",
        Amount:         purchaseOrder.TotalAmount,
        Currency:       purchaseOrder.Currency,
        PaymentMethod:  "bank_transfer",
        LinkedPO:       purchaseOrder.DocumentNumber,
        LinkedGRN:      "", // No GRN for payment-first flow
        ApprovalStage:  0,
        CreatedAt:      time.Now(),
        UpdatedAt:      time.Now(),
        OrganizationID: purchaseOrder.OrganizationID,
        BudgetCode:     purchaseOrder.BudgetCode,
        CostCenter:     purchaseOrder.CostCenter,
        ProjectCode:    purchaseOrder.ProjectCode,
        Title:          purchaseOrder.Title,
        Description:    purchaseOrder.Description,
        Department:     purchaseOrder.Department,
        DepartmentID:   purchaseOrder.DepartmentID,
        CreatedBy:      "system", // System-created
    }

    // Initialize empty approval history
    paymentVoucher.ApprovalHistory = datatypes.NewJSONType([]types.ApprovalRecord{})

    // Initialize action history
    pvInitialHistory := []types.ActionHistoryEntry{
        {
            ID:          uuid.New().String(),
            Action:      "CREATED_FROM_PO",
            PerformedBy: "system",
            Timestamp:   time.Now(),
            Metadata: map[string]interface{}{
                "linkedDocNumber": purchaseOrder.DocumentNumber,
                "linkedDocType":   "purchase_order",
                "flow":            "payment_first",
                "automated":       true,
            },
        },
        {
            ID:          uuid.New().String(),
            Action:      "CREATE",
            ActionType:  "CREATE",
            PerformedBy: "system",
            Timestamp:   time.Now(),
            PerformedAt: time.Now(),
            Comments:    "Payment voucher auto-created from approved purchase order (payment-first flow)",
            NewStatus:   "DRAFT",
        },
    }
    paymentVoucher.ActionHistory = datatypes.NewJSONType(pvInitialHistory)

    // Save to database
    if err := s.db.Create(&paymentVoucher).Error; err != nil {
        return &AutomationResult{
            Success: false,
            Error:   fmt.Errorf("failed to create payment voucher: %w", err),
        }, nil
    }

    // Log audit event
    if s.auditService != nil {
        details := fmt.Sprintf("Auto-created PV %s from approved PO %s (payment-first flow)",
            documentNumber, purchaseOrder.DocumentNumber)
        s.auditService.LogEvent(ctx, "system", purchaseOrder.OrganizationID,
            "pv_auto_created", "payment_voucher", paymentVoucher.ID, details, "", "")
    }

    // Send notification to finance team
    if s.notificationSvc != nil {
        event := NotificationEvent{
            Type:         "document_created",
            DocumentID:   paymentVoucher.ID,
            DocumentType: "payment_voucher",
            Action:       "auto_created",
            ActorID:      "system",
            Details:      fmt.Sprintf("Payment Voucher %s was automatically created from approved PO %s",
                documentNumber, purchaseOrder.DocumentNumber),
            Timestamp:    time.Now(),
        }
        s.notificationSvc.HandleWorkflowEvent(event)
    }

    return &AutomationResult{
        Success:         true,
        CreatedDocument: paymentVoucher,
        DocumentType:    "payment_voucher",
        DocumentID:      paymentVoucher.ID,
    }, nil
}
```

---

## Implementation Priority

### Phase 1: Critical Fixes (Implement Immediately)

1. ✅ Fix 1: Automatic GRN workflow assignment
2. ✅ Fix 3: Audit logging for status changes

### Phase 2: Enhancement Fixes (Next Sprint)

3. ✅ Fix 2: Automatic PV workflow assignment
4. ✅ Fix 4: Payment-first PV auto-creation

### Phase 3: Future Improvements

5. Approval deadline enforcement
6. Notification retry mechanism
7. Partial approval handling

---

## Testing After Implementation

### Test 1: GRN Auto-Submission

```
1. Enable AutoSubmitGRNToWorkflow = true
2. Create and approve PO
3. Verify:
   ✅ GRN auto-created
   ✅ GRN status = PENDING (not DRAFT)
   ✅ WorkflowAssignment created for GRN
   ✅ First approval task created
   ✅ Notification sent to approver
```

### Test 2: PV Auto-Submission

```
1. Create PV from PO with workflowId
2. Verify:
   ✅ PV created
   ✅ PV status = PENDING (not DRAFT)
   ✅ WorkflowAssignment created for PV
   ✅ First approval task created
   ✅ Action history includes SUBMIT entry
```

### Test 3: Payment-First PV Auto-Creation

```
1. Set org procurement_flow = "payment_first"
2. Enable AutoCreatePVFromPO = true
3. Create and approve PO
4. Verify:
   ✅ PV auto-created (no GRN)
   ✅ PV.LinkedPO = PO document number
   ✅ PV.LinkedGRN = empty
   ✅ PV status = DRAFT
   ✅ Notification sent to finance
```

### Test 4: Audit Trail

```
1. Approve PO through workflow
2. Check audit_events table
3. Verify:
   ✅ Event: document_status_changed
   ✅ Old value: PENDING
   ✅ New value: APPROVED
   ✅ Entity type: purchase_order
```

---

## Rollback Plan

If any fix causes issues:

```bash
# Revert specific fix
git checkout HEAD~1 -- backend/services/document_automation_service.go
git checkout HEAD~1 -- backend/services/workflow_execution_service.go

# Rebuild and restart
cd backend
go build
./backend
```

---

## Configuration

Add to organization settings or environment variables:

```env
# Automation flags
AUTO_SUBMIT_GRN_TO_WORKFLOW=true
AUTO_SUBMIT_PV_TO_WORKFLOW=false
AUTO_CREATE_PV_FROM_PO=true
```

Or in database `organization_settings`:

```json
{
  "automationConfig": {
    "autoSubmitGRNToWorkflow": true,
    "autoSubmitPVToWorkflow": false,
    "autoCreatePVFromPO": true
  }
}
```
