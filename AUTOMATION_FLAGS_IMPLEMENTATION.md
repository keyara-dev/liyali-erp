# Opt-In Automation Flags Implementation Guide

## Overview

This guide implements organization-level opt-in flags for automatic document workflow submission. Each organization can independently enable/disable automation features based on their workflow preferences.

---

## Architecture

### Flag Storage

Automation flags are stored in the `organization_settings` table, allowing per-organization configuration.

### Flag Hierarchy

```
Organization Settings (Database)
    ↓
AutomationConfig (Runtime)
    ↓
Workflow Execution (Conditional Logic)
```

---

## Implementation Steps

### Step 1: Update Database Schema

**File**: `backend/database/migrations/XXX_add_automation_flags.up.sql`

```sql
-- Add automation flags to organization_settings table
ALTER TABLE organization_settings
ADD COLUMN IF NOT EXISTS auto_submit_grn_to_workflow BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS auto_submit_pv_to_workflow BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS auto_create_pv_from_po BOOLEAN DEFAULT FALSE;

-- Add comment for documentation
COMMENT ON COLUMN organization_settings.auto_submit_grn_to_workflow IS
'When enabled, auto-created GRNs are automatically submitted to workflow instead of staying in DRAFT';

COMMENT ON COLUMN organization_settings.auto_submit_pv_to_workflow IS
'When enabled, created PVs are automatically submitted to workflow instead of staying in DRAFT';

COMMENT ON COLUMN organization_settings.auto_create_pv_from_po IS
'When enabled, PVs are automatically created from approved POs in payment-first flow';
```

**File**: `backend/database/migrations/XXX_add_automation_flags.down.sql`

```sql
-- Rollback automation flags
ALTER TABLE organization_settings
DROP COLUMN IF EXISTS auto_submit_grn_to_workflow,
DROP COLUMN IF EXISTS auto_submit_pv_to_workflow,
DROP COLUMN IF EXISTS auto_create_pv_from_po;
```

---

### Step 2: Update OrganizationSettings Model

**File**: `backend/models/organization.go`

```go
// OrganizationSettings stores per-org configuration
type OrganizationSettings struct {
    ID             string `gorm:"primaryKey" json:"id"`
    OrganizationID string `gorm:"uniqueIndex;not null" json:"organizationId"`

    // Existing fields
    ProcurementFlow string `json:"procurementFlow"` // "goods_first" or "payment_first"
    // ... other existing fields ...

    // NEW: Automation flags (opt-in)
    AutoSubmitGRNToWorkflow bool `gorm:"column:auto_submit_grn_to_workflow;default:false" json:"autoSubmitGRNToWorkflow"`
    AutoSubmitPVToWorkflow  bool `gorm:"column:auto_submit_pv_to_workflow;default:false" json:"autoSubmitPVToWorkflow"`
    AutoCreatePVFromPO      bool `gorm:"column:auto_create_pv_from_po;default:false" json:"autoCreatePVFromPO"`

    CreatedAt time.Time `json:"createdAt"`
    UpdatedAt time.Time `json:"updatedAt"`
}
```

---

### Step 3: Update AutomationConfig to Use Org Settings

**File**: `backend/services/document_automation_service.go`

```go
// GetAutomationConfigForOrg retrieves automation config based on org settings
func (s *DocumentAutomationService) GetAutomationConfigForOrg(orgID string) AutomationConfig {
    // Get organization settings
    var orgSettings models.OrganizationSettings
    if err := s.db.Where("organization_id = ?", orgID).First(&orgSettings).Error; err != nil {
        // Return default config if settings not found
        return s.GetDefaultAutomationConfig()
    }

    // Build config from org settings
    return AutomationConfig{
        AutoCreatePOFromRequisition: true,  // Always enabled (existing behavior)
        AutoCreateGRNFromPO:         true,  // Always enabled (existing behavior)
        AutoCreatePVFromGRN:         true,  // Always enabled (existing behavior)
        RequireApprovalForAuto:      false, // Always disabled (existing behavior)

        // NEW: Opt-in flags from org settings
        AutoSubmitGRNToWorkflow: orgSettings.AutoSubmitGRNToWorkflow,
        AutoSubmitPVToWorkflow:  orgSettings.AutoSubmitPVToWorkflow,
        AutoCreatePVFromPO:      orgSettings.AutoCreatePVFromPO,
    }
}

// GetDefaultAutomationConfig returns default config (all opt-in flags disabled)
func (s *DocumentAutomationService) GetDefaultAutomationConfig() AutomationConfig {
    return AutomationConfig{
        AutoCreatePOFromRequisition: true,
        AutoCreateGRNFromPO:         true,
        AutoCreatePVFromGRN:         true,
        RequireApprovalForAuto:      false,

        // Opt-in flags default to FALSE (manual submission)
        AutoSubmitGRNToWorkflow: false,
        AutoSubmitPVToWorkflow:  false,
        AutoCreatePVFromPO:      false,
    }
}
```

---

### Step 4: Update Workflow Execution Service

**File**: `backend/services/workflow_execution_service.go`

**Modify `triggerPostApprovalAutomation` method**:

```go
func (s *WorkflowExecutionService) triggerPostApprovalAutomation(ctx context.Context, entityType, entityID string) error {
    if s.automationService == nil {
        return nil
    }

    // Get organization ID from entity
    var organizationID string
    switch entityType {
    case "REQUISITION", "requisition":
        var req models.Requisition
        if err := s.db.Where("id = ?", entityID).First(&req).Error; err == nil {
            organizationID = req.OrganizationID
        }
    case "PURCHASE_ORDER", "purchase_order":
        var po models.PurchaseOrder
        if err := s.db.Where("id = ?", entityID).First(&po).Error; err == nil {
            organizationID = po.OrganizationID
        }
    case "GRN", "grn":
        var grn models.GoodsReceivedNote
        if err := s.db.Where("id = ?", entityID).First(&grn).Error; err == nil {
            organizationID = grn.OrganizationID
        }
    }

    // Get org-specific automation config
    config := s.automationService.GetAutomationConfigForOrg(organizationID)

    switch entityType {
    case "REQUISITION", "requisition":
        // ... existing requisition logic ...

    case "PURCHASE_ORDER", "purchase_order":
        var po models.PurchaseOrder
        if err := s.db.Where("id = ?", entityID).First(&po).Error; err != nil {
            return fmt.Errorf("failed to get purchase order: %w", err)
        }

        // Determine effective procurement flow
        effectiveFlow := po.ProcurementFlow
        if effectiveFlow == "" {
            orgSvc := NewOrganizationService(s.db)
            orgSettings, _ := orgSvc.GetOrganizationSettings(po.OrganizationID)
            if orgSettings != nil && orgSettings.ProcurementFlow != "" {
                effectiveFlow = orgSettings.ProcurementFlow
            } else {
                effectiveFlow = "goods_first"
            }
        }

        if effectiveFlow == "payment_first" && config.AutoCreatePVFromPO {
            // Payment-first: Auto-create PV directly from PO
            result, err := s.automationService.CreatePaymentVoucherFromPO(ctx, &po, config)
            if err != nil {
                return fmt.Errorf("failed to create payment voucher: %w", err)
            }
            if result.Success {
                // Update PO with auto-created PV info
                autoCreatedPV := map[string]interface{}{
                    "id":      result.DocumentID,
                    "created": true,
                }
                if result.CreatedDocument != nil {
                    if pv, ok := result.CreatedDocument.(models.PaymentVoucher); ok {
                        autoCreatedPV["documentNumber"] = pv.DocumentNumber
                    }
                }
                autoCreatedJSON, _ := datatypes.NewJSONType(autoCreatedPV).MarshalJSON()
                s.db.Model(&po).Updates(map[string]interface{}{
                    "automation_used": true,
                    "auto_created_pv": datatypes.JSON(autoCreatedJSON),
                })
            }
        } else if effectiveFlow == "goods_first" && config.AutoCreateGRNFromPO {
            // Goods-first: Auto-create GRN from PO
            result, err := s.automationService.CreateGRNFromPurchaseOrder(ctx, &po, config)
            if err != nil {
                return fmt.Errorf("failed to create GRN: %w", err)
            }
            if result.Success {
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

    case "GRN", "grn":
        // ... existing GRN logic ...
    }

    return nil
}
```

---

### Step 5: Update CreateGRNFromPurchaseOrder

**File**: `backend/services/document_automation_service.go`

**Add workflow assignment logic after GRN creation**:

```go
func (s *DocumentAutomationService) CreateGRNFromPurchaseOrder(
    ctx context.Context,
    purchaseOrder *models.PurchaseOrder,
    config AutomationConfig,
) (*AutomationResult, error) {
    // ... existing GRN creation logic ...

    // Save to database
    if err := s.db.Create(&grn).Error; err != nil {
        return &AutomationResult{
            Success: false,
            Error:   fmt.Errorf("failed to create GRN: %w", err),
        }, nil
    }

    // NEW: Auto-submit GRN to workflow if enabled
    if config.AutoSubmitGRNToWorkflow && s.workflowService != nil && s.workflowExecutionService != nil {
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

    // Log audit event
    if s.auditService != nil {
        details := fmt.Sprintf("Auto-created GRN %s from approved PO %s", documentNumber, purchaseOrder.DocumentNumber)
        s.auditService.LogEvent(ctx, "system", grn.OrganizationID, "grn_auto_created", "grn", grn.ID, details, "", "")
    }

    // Send notification
    if s.notificationSvc != nil {
        status := "DRAFT"
        if config.AutoSubmitGRNToWorkflow {
            status = "PENDING (auto-submitted)"
        }
        event := NotificationEvent{
            Type:         "document_created",
            DocumentID:   grn.ID,
            DocumentType: "grn",
            Action:       "auto_created",
            ActorID:      "system",
            Details:      fmt.Sprintf("GRN %s was automatically created from PO %s (Status: %s)", documentNumber, purchaseOrder.DocumentNumber, status),
            Timestamp:    time.Now(),
        }
        s.notificationSvc.HandleWorkflowEvent(event)
    }

    return &AutomationResult{
        Success:         true,
        CreatedDocument: &grn,
        DocumentType:    "grn",
        DocumentID:      grn.ID,
    }, nil
}
```

---

### Step 6: Update CreatePaymentVoucherFromPO Handler

**File**: `backend/handlers/document_extras_handler.go`

**Add workflow assignment logic after PV creation**:

```go
func CreatePaymentVoucherFromPO(c *fiber.Ctx) error {
    // ... existing PV creation logic ...

    if err := config.DB.Create(&voucher).Error; err != nil {
        logging.LogError(c, err, "create_pv_from_po_failed", nil)
        return utils.SendInternalError(c, "Failed to create payment voucher", err)
    }

    // NEW: Auto-submit PV to workflow if enabled and workflowId provided
    if req.WorkflowID != "" {
        // Get org settings to check if auto-submit is enabled
        orgSvc := services.NewOrganizationService(config.DB)
        orgSettings, _ := orgSvc.GetOrganizationSettings(tenant.OrganizationID)

        autoSubmitEnabled := false
        if orgSettings != nil {
            autoSubmitEnabled = orgSettings.AutoSubmitPVToWorkflow
        }

        if autoSubmitEnabled {
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
            } else {
                // Update PV status to PENDING
                voucher.Status = "PENDING"
                voucher.UpdatedAt = time.Now()

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
                    Details: map[string]interface{}{
                        "documentNumber": voucher.DocumentNumber,
                        "workflowId":     req.WorkflowID,
                        "assignmentId":   assignment.ID,
                        "autoSubmitted":  true,
                    },
                })

                logger.Info("pv_auto_submitted_to_workflow")
            }
        }
    }

    // ... rest of handler ...
}
```

---

### Step 7: Update Organization Settings API

**File**: `backend/handlers/organization.go`

**Update `UpdateOrganizationSettings` to accept new flags**:

```go
// UpdateOrganizationSettings updates organization settings
// PUT /api/v1/organization/settings
func UpdateOrganizationSettings(c *fiber.Ctx) error {
    tenant, err := middleware.GetTenantContext(c)
    if err != nil {
        return utils.SendUnauthorizedError(c, "Organization context required")
    }

    var req struct {
        ProcurementFlow         string `json:"procurementFlow"`
        // ... other existing fields ...

        // NEW: Automation flags
        AutoSubmitGRNToWorkflow *bool `json:"autoSubmitGRNToWorkflow"`
        AutoSubmitPVToWorkflow  *bool `json:"autoSubmitPVToWorkflow"`
        AutoCreatePVFromPO      *bool `json:"autoCreatePVFromPO"`
    }

    if err := c.BodyParser(&req); err != nil {
        return utils.SendBadRequestError(c, "Invalid request body")
    }

    orgService := services.NewOrganizationService(config.DB)

    // Get existing settings
    settings, err := orgService.GetOrganizationSettings(tenant.OrganizationID)
    if err != nil {
        return utils.SendInternalError(c, "Failed to get organization settings", err)
    }

    // Update fields
    if req.ProcurementFlow != "" {
        settings.ProcurementFlow = req.ProcurementFlow
    }

    // Update automation flags if provided
    if req.AutoSubmitGRNToWorkflow != nil {
        settings.AutoSubmitGRNToWorkflow = *req.AutoSubmitGRNToWorkflow
    }
    if req.AutoSubmitPVToWorkflow != nil {
        settings.AutoSubmitPVToWorkflow = *req.AutoSubmitPVToWorkflow
    }
    if req.AutoCreatePVFromPO != nil {
        settings.AutoCreatePVFromPO = *req.AutoCreatePVFromPO
    }

    settings.UpdatedAt = time.Now()

    if err := config.DB.Save(settings).Error; err != nil {
        return utils.SendInternalError(c, "Failed to update organization settings", err)
    }

    // Log audit event
    go services.LogDocumentEvent(config.DB, services.DocumentEvent{
        OrganizationID: tenant.OrganizationID,
        DocumentID:     settings.ID,
        DocumentType:   "organization_settings",
        UserID:         tenant.UserID,
        Action:         "updated",
        Details: map[string]interface{}{
            "autoSubmitGRNToWorkflow": settings.AutoSubmitGRNToWorkflow,
            "autoSubmitPVToWorkflow":  settings.AutoSubmitPVToWorkflow,
            "autoCreatePVFromPO":      settings.AutoCreatePVFromPO,
        },
    })

    return utils.SendSuccess(c, settings, "Organization settings updated successfully")
}
```

---

### Step 8: Frontend - Organization Settings UI

**File**: `frontend/src/types/organization.ts`

```typescript
export interface OrganizationSettings {
  id: string;
  organizationId: string;
  procurementFlow: "goods_first" | "payment_first" | "";

  // NEW: Automation flags
  autoSubmitGRNToWorkflow: boolean;
  autoSubmitPVToWorkflow: boolean;
  autoCreatePVFromPO: boolean;

  createdAt: Date;
  updatedAt: Date;
}

export interface UpdateOrganizationSettingsRequest {
  procurementFlow?: "goods_first" | "payment_first" | "";

  // NEW: Automation flags
  autoSubmitGRNToWorkflow?: boolean;
  autoSubmitPVToWorkflow?: boolean;
  autoCreatePVFromPO?: boolean;
}
```

**File**: `frontend/src/components/organization-settings-form.tsx` (or similar)

```typescript
export function OrganizationSettingsForm() {
  const [settings, setSettings] = useState<OrganizationSettings | null>(null);

  return (
    <form>
      {/* Existing fields */}

      <div className="space-y-4">
        <h3 className="text-lg font-semibold">Automation Settings</h3>
        <p className="text-sm text-muted-foreground">
          Configure automatic workflow submission for documents
        </p>

        <div className="space-y-3">
          {/* GRN Auto-Submit */}
          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <label className="text-sm font-medium">
                Auto-submit GRN to Workflow
              </label>
              <p className="text-xs text-muted-foreground">
                When enabled, auto-created GRNs are automatically submitted for approval
                instead of staying in DRAFT status
              </p>
            </div>
            <Switch
              checked={settings?.autoSubmitGRNToWorkflow ?? false}
              onCheckedChange={(checked) =>
                handleUpdate({ autoSubmitGRNToWorkflow: checked })
              }
            />
          </div>

          {/* PV Auto-Submit */}
          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <label className="text-sm font-medium">
                Auto-submit PV to Workflow
              </label>
              <p className="text-xs text-muted-foreground">
                When enabled, created PVs are automatically submitted for approval
                instead of staying in DRAFT status
              </p>
            </div>
            <Switch
              checked={settings?.autoSubmitPVToWorkflow ?? false}
              onCheckedChange={(checked) =>
                handleUpdate({ autoSubmitPVToWorkflow: checked })
              }
            />
          </div>

          {/* PV Auto-Create (Payment-First) */}
          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <label className="text-sm font-medium">
                Auto-create PV from PO (Payment-First)
              </label>
              <p className="text-xs text-muted-foreground">
                When enabled and procurement flow is "payment-first", PVs are automatically
                created when POs are approved
              </p>
            </div>
            <Switch
              checked={settings?.autoCreatePVFromPO ?? false}
              onCheckedChange={(checked) =>
                handleUpdate({ autoCreatePVFromPO: checked })
              }
              disabled={settings?.procurementFlow !== "payment_first"}
            />
          </div>
        </div>

        <Alert>
          <InfoIcon className="h-4 w-4" />
          <AlertDescription>
            These settings control automation behavior for your organization.
            Disabling these flags allows for manual review before workflow submission.
          </AlertDescription>
        </Alert>
      </div>
    </form>
  );
}
```

---

## Usage Examples

### Example 1: Enable GRN Auto-Submission

**Admin Console**:

1. Navigate to Organization Settings
2. Toggle "Auto-submit GRN to Workflow" ON
3. Save settings

**Result**:

- When PO is approved, GRN is auto-created
- GRN is automatically submitted to workflow (status: PENDING)
- First approval task created
- Notification sent to approver

### Example 2: Disable All Automation (Manual Review)

**Admin Console**:

1. Navigate to Organization Settings
2. Toggle all automation flags OFF
3. Save settings

**Result**:

- GRN auto-created but stays in DRAFT
- PV created but stays in DRAFT
- Finance team manually reviews and submits

### Example 3: Payment-First with Auto-PV

**Admin Console**:

1. Set Procurement Flow: "Payment-First"
2. Toggle "Auto-create PV from PO" ON
3. Toggle "Auto-submit PV to Workflow" ON
4. Save settings

**Result**:

- When PO is approved, PV is auto-created
- PV is automatically submitted to workflow
- No GRN required

---

## Testing Checklist

### Test 1: GRN Auto-Submit Enabled

```
✅ Enable autoSubmitGRNToWorkflow
✅ Approve PO
✅ Verify GRN created with status = PENDING
✅ Verify workflow assignment created
✅ Verify first approval task created
✅ Verify notification sent
```

### Test 2: GRN Auto-Submit Disabled

```
✅ Disable autoSubmitGRNToWorkflow
✅ Approve PO
✅ Verify GRN created with status = DRAFT
✅ Verify NO workflow assignment
✅ Verify finance can manually submit
```

### Test 3: PV Auto-Submit Enabled

```
✅ Enable autoSubmitPVToWorkflow
✅ Create PV from PO with workflowId
✅ Verify PV status = PENDING
✅ Verify workflow assignment created
✅ Verify action history includes SUBMIT entry
```

### Test 4: Payment-First PV Auto-Create

```
✅ Set procurementFlow = "payment_first"
✅ Enable autoCreatePVFromPO
✅ Approve PO
✅ Verify PV auto-created
✅ Verify PV.LinkedGRN = empty
✅ Verify notification sent to finance
```

---

## Migration Guide

### For Existing Organizations

**Default Behavior**: All automation flags default to `FALSE` (manual submission)

**Migration Steps**:

1. Run database migration to add columns
2. All existing orgs will have flags = FALSE
3. Admins can opt-in per organization
4. No breaking changes to existing workflows

### Recommended Settings

**Conservative (Manual Review)**:

```json
{
  "autoSubmitGRNToWorkflow": false,
  "autoSubmitPVToWorkflow": false,
  "autoCreatePVFromPO": false
}
```

**Moderate (Semi-Automated)**:

```json
{
  "autoSubmitGRNToWorkflow": true,
  "autoSubmitPVToWorkflow": false,
  "autoCreatePVFromPO": false
}
```

**Aggressive (Fully Automated)**:

```json
{
  "autoSubmitGRNToWorkflow": true,
  "autoSubmitPVToWorkflow": true,
  "autoCreatePVFromPO": true
}
```

---

## Benefits

### 1. Flexibility

- Each organization controls their automation level
- Can enable/disable per feature
- No code changes required

### 2. Safety

- Defaults to manual (safe)
- Opt-in prevents surprises
- Can revert anytime

### 3. Scalability

- High-volume orgs can enable full automation
- Low-volume orgs can keep manual review
- Adapts to business needs

### 4. Auditability

- All automation actions logged
- Settings changes tracked
- Clear audit trail

---

## Summary

This implementation provides:

✅ **Organization-level opt-in flags** for automation  
✅ **Database-backed configuration** (persistent)  
✅ **Admin UI** for easy management  
✅ **Backward compatible** (defaults to manual)  
✅ **Fully auditable** (all actions logged)  
✅ **Flexible** (enable/disable per feature)

Organizations can now choose their automation level based on their workflow preferences and compliance requirements.
