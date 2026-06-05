package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/types"
	"github.com/liyali/liyali-gateway/utils"
	"gorm.io/datatypes"
)

// Automation levels for an auto-created document.
const (
	AutomationManual      = "manual"
	AutomationAutoSubmit  = "auto_submit"
	AutomationAutoApprove = "auto_approve"
)

// resolveAutomationAction maps an automation level + document amount + org cap to
// a concrete action: "draft" (leave as-is), "submit" (send to approval workflow),
// or "approve" (unattended approval). auto_approve only yields "approve" when the
// amount is at/below the cap (cap > 0); otherwise it falls back to "submit".
// Unknown/empty levels are treated as manual.
func resolveAutomationAction(level string, amount, maxAmount float64) string {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case AutomationAutoSubmit:
		return "submit"
	case AutomationAutoApprove:
		if maxAmount > 0 && amount <= maxAmount+0.01 {
			return "approve"
		}
		return "submit"
	default:
		return "draft"
	}
}

// autoCreateFromApprovedPO runs POST-COMMIT after a Purchase Order is approved
// (via triggerPostApprovalAutomation). Mode-aware:
//   - payment_first: auto-create a PV (if enabled) and apply PVAutomationLevel.
//   - goods_first:   optionally create a DRAFT GRN placeholder for the receiver
//     to sign (GRNs require receiver+certifier signatures, so they are never
//     auto-submitted; the PV is auto-created later when the GRN completes).
//
// Best-effort: returns an error only for logging by the caller; never fatal.
func (s *WorkflowExecutionService) autoCreateFromApprovedPO(ctx context.Context, poID string) error {
	var po models.PurchaseOrder
	if err := s.db.Where("id = ?", poID).First(&po).Error; err != nil {
		return nil
	}
	orgID := po.OrganizationID
	settings, err := NewOrganizationService(s.db).GetOrganizationSettings(orgID)
	if err != nil || settings == nil {
		return nil
	}
	flow := utils.ResolveProcurementFlow(po.ProcurementFlow, settings.ProcurementFlow)

	if flow == "payment_first" {
		if !settings.AutoCreatePVFromPO {
			return nil
		}
		pvID, err := s.createDraftPVFromPO(&po)
		if err != nil || pvID == "" {
			return err
		}
		action := resolveAutomationAction(settings.PVAutomationLevel, po.TotalAmount, settings.AutoApproveMaxAmount)
		return s.applyPVAutomation(ctx, pvID, orgID, action)
	}

	// goods_first
	if settings.AutoCreateGRNFromPO {
		return s.createDraftGRNFromPO(&po)
	}
	return nil
}

// createDraftPVFromPO creates a DRAFT PV linked to the PO. Honors the Phase-A
// one-live-PV-per-PO rule: returns ("", nil) when a live PV already exists.
func (s *WorkflowExecutionService) createDraftPVFromPO(po *models.PurchaseOrder) (string, error) {
	var existing int64
	s.db.Model(&models.PaymentVoucher{}).
		Where("linked_po = ? AND organization_id = ? AND UPPER(status) NOT IN ('CANCELLED','REJECTED')",
			po.DocumentNumber, po.OrganizationID).Count(&existing)
	if existing > 0 {
		return "", nil
	}

	// Copy PO lines into PV payment items so the PV isn't an empty-lined doc.
	pvItems := make([]types.PaymentItem, 0)
	for _, it := range po.Items.Data() {
		pvItems = append(pvItems, types.PaymentItem{
			Description: it.Description,
			Amount:      float64(it.Quantity) * it.UnitPrice,
			GLCode:      po.GLCode,
		})
	}

	now := time.Now()
	pv := models.PaymentVoucher{
		ID:             uuid.New().String(),
		OrganizationID: po.OrganizationID,
		DocumentNumber: utils.GenerateDocumentNumber("PV"),
		LinkedPO:       po.DocumentNumber,
		VendorID:       po.VendorID,
		VendorName:     po.VendorName,
		Amount:         po.TotalAmount,
		Currency:       po.Currency,
		Status:         models.StatusDraft,
		CreatedBy:      po.CreatedBy,
		Description:    fmt.Sprintf("Auto-created from approved PO %s (payment-first)", po.DocumentNumber),
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	pv.Items = datatypes.NewJSONType(pvItems)
	pv.ApprovalHistory = datatypes.NewJSONType([]types.ApprovalRecord{})
	pv.ActionHistory = datatypes.NewJSONType([]types.ActionHistoryEntry{{
		ID:          uuid.New().String(),
		Action:      "AUTO_CREATED",
		ActionType:  "CREATE",
		Timestamp:   now,
		PerformedAt: now,
		NewStatus:   models.StatusDraft,
		Comments:    "Auto-created from approved PO via AutoCreatePVFromPO",
	}})
	if err := s.db.Create(&pv).Error; err != nil {
		return "", fmt.Errorf("create draft PV from PO: %w", err)
	}
	go utils.SyncDocumentAs(s.db, "PAYMENT_VOUCHER", pv.ID, "system")
	return pv.ID, nil
}

// createDraftGRNFromPO creates a DRAFT GRN placeholder for the PO's goods, for a
// receiver to record quantities + sign. Skipped if a non-cancelled GRN already
// exists for the PO. GRNs are never auto-submitted (they require signatures).
func (s *WorkflowExecutionService) createDraftGRNFromPO(po *models.PurchaseOrder) error {
	var existing int64
	s.db.Model(&models.GoodsReceivedNote{}).
		Where("po_document_number = ? AND organization_id = ? AND UPPER(status) NOT IN ('CANCELLED','REJECTED')",
			po.DocumentNumber, po.OrganizationID).Count(&existing)
	if existing > 0 {
		return nil
	}

	grnItems := make([]types.GRNItem, 0)
	for _, it := range po.Items.Data() {
		grnItems = append(grnItems, types.GRNItem{
			Description:     it.Description,
			ItemCode:        it.ItemCode,
			QuantityOrdered: it.Quantity,
			QuantityReceived: 0,
			Condition:       "good",
		})
	}

	now := time.Now()
	grn := models.GoodsReceivedNote{
		ID:               uuid.New().String(),
		OrganizationID:   po.OrganizationID,
		DocumentNumber:   utils.GenerateDocumentNumber("GRN"),
		PODocumentNumber: po.DocumentNumber,
		Status:           models.StatusDraft,
		SignoffStatus:    "PENDING_RECEIVER",
		ReceivedDate:     now,
		CreatedBy:        po.CreatedBy,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	grn.Items = datatypes.NewJSONType(grnItems)
	grn.QualityIssues = datatypes.NewJSONType([]types.QualityIssue{})
	grn.ApprovalHistory = datatypes.NewJSONType([]types.ApprovalRecord{})
	grn.ActionHistory = datatypes.NewJSONType([]types.ActionHistoryEntry{{
		ID:          uuid.New().String(),
		Action:      "AUTO_CREATED",
		ActionType:  "CREATE",
		Timestamp:   now,
		PerformedAt: now,
		NewStatus:   models.StatusDraft,
		Comments:    "Draft GRN placeholder auto-created from approved PO — awaiting goods receipt + signatures",
	}})
	if err := s.db.Create(&grn).Error; err != nil {
		return fmt.Errorf("create draft GRN from PO: %w", err)
	}
	go utils.SyncDocumentAs(s.db, "GRN", grn.ID, "system")
	return nil
}

// applyPVAutomation submits or unattended-approves a freshly-created DRAFT PV per
// the resolved action. Runs post-commit (own connection/txn); never nested in the
// approval transaction that triggered it. A missing payment_voucher workflow is a
// no-op (the PV stays DRAFT) with a logged warning.
func (s *WorkflowExecutionService) applyPVAutomation(ctx context.Context, pvID, orgID, action string) error {
	switch action {
	case "submit":
		if _, err := s.AssignWorkflowToDocument(ctx, orgID, pvID, "payment_voucher", "system"); err != nil {
			fmt.Printf("auto-submit PV %s: no workflow assigned (left DRAFT): %v\n", pvID, err)
			return nil
		}
		if err := s.db.Model(&models.PaymentVoucher{}).Where("id = ?", pvID).
			Update("status", models.StatusPending).Error; err != nil {
			return err
		}
		go utils.SyncDocumentAs(s.db, "PAYMENT_VOUCHER", pvID, "system")
		return nil
	case "approve":
		return s.autoApprovePV(ctx, pvID, orgID)
	default:
		return nil
	}
}

// autoApprovePV performs unattended approval: assigns the PV workflow, marks the
// assignment COMPLETED, sets the PV APPROVED, and creates the payment_execution
// task (so it can still be marked paid by finance). If no workflow is configured
// the PV is left DRAFT (we never half-approve a PV that could never be paid).
func (s *WorkflowExecutionService) autoApprovePV(ctx context.Context, pvID, orgID string) error {
	assignment, err := s.AssignWorkflowToDocument(ctx, orgID, pvID, "payment_voucher", "system")
	if err != nil || assignment == nil {
		fmt.Printf("auto-approve PV %s: no workflow assigned (left DRAFT): %v\n", pvID, err)
		return nil
	}

	now := time.Now()
	tx := s.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Close the assignment + any open stage tasks.
	if err := tx.Model(&models.WorkflowAssignment{}).Where("id = ?", assignment.ID).
		Updates(map[string]interface{}{"status": "COMPLETED", "completed_at": now}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Model(&models.WorkflowTask{}).
		Where("workflow_assignment_id = ? AND UPPER(status) IN ('PENDING','CLAIMED')", assignment.ID).
		Updates(map[string]interface{}{"status": "COMPLETED", "completed_at": now}).Error; err != nil {
		tx.Rollback()
		return err
	}
	// Approve the PV + open the payment-execution task for finance.
	if err := tx.Model(&models.PaymentVoucher{}).Where("id = ?", pvID).
		Update("status", models.StatusApproved).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := s.createPaymentExecutionTask(tx, assignment, now); err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		return err
	}
	go utils.SyncDocumentAs(s.db, "PAYMENT_VOUCHER", pvID, "system")
	return nil
}

// ApplyPVAutomationForCompletedGRN is the exported, post-commit entry point used
// by the no-workflow MarkGRNComplete handler path so it honors PVAutomationLevel
// the same way the workflow GRN-completion path does.
func (s *WorkflowExecutionService) ApplyPVAutomationForCompletedGRN(ctx context.Context, grnID string) error {
	return s.applyPVLevelForCompletedGRN(ctx, grnID)
}

// applyPVLevelForCompletedGRN runs POST-COMMIT after a GRN completes. The DRAFT
// PV is auto-created in-transaction by AutoCreatePVFromCompletedGRN; here we
// apply the org's PVAutomationLevel to it (submit / unattended-approve). No-op
// when no auto-created PV exists (feature off, or payment_first).
func (s *WorkflowExecutionService) applyPVLevelForCompletedGRN(ctx context.Context, grnID string) error {
	var grn models.GoodsReceivedNote
	if err := s.db.Where("id = ?", grnID).First(&grn).Error; err != nil {
		return nil
	}
	settings, err := NewOrganizationService(s.db).GetOrganizationSettings(grn.OrganizationID)
	if err != nil || settings == nil {
		return nil
	}
	var pv models.PaymentVoucher
	if err := s.db.Where("linked_grn = ? AND organization_id = ? AND UPPER(status) = 'DRAFT'",
		grn.DocumentNumber, grn.OrganizationID).Order("created_at DESC").First(&pv).Error; err != nil {
		return nil
	}
	action := resolveAutomationAction(settings.PVAutomationLevel, pv.Amount, settings.AutoApproveMaxAmount)
	return s.applyPVAutomation(ctx, pv.ID, grn.OrganizationID, action)
}
