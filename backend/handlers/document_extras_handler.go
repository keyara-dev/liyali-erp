package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/logging"
	"github.com/liyali/liyali-gateway/middleware"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/services"
	"github.com/liyali/liyali-gateway/types"
	"github.com/liyali/liyali-gateway/utils"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ============================================================================
// PURCHASE ORDER — FROM REQUISITION
// POST /api/v1/purchase-orders/from-requisition
// ============================================================================

// CreatePurchaseOrderFromRequisition creates a PO pre-populated from an approved requisition.
func CreatePurchaseOrderFromRequisition(c *fiber.Ctx) error {
	logger := logging.FromContext(c)
	logger.Info("create_po_from_requisition_request")

	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return utils.SendUnauthorizedError(c, "Organization context required")
	}

	var req struct {
		RequisitionID             string         `json:"requisitionId"`
		RequisitionDocumentNumber string         `json:"requisitionDocumentNumber"`
		Title                     string         `json:"title"`
		Description               string         `json:"description"`
		VendorID                  string         `json:"vendorId"`
		VendorName                string         `json:"vendorName"`
		Department                string         `json:"department"`
		DepartmentID              string         `json:"departmentId"`
		RequiredByDate            *time.Time     `json:"requiredByDate"`
		Priority                  string         `json:"priority"`
		Items                     []types.POItem `json:"items"`
		TotalAmount               float64        `json:"totalAmount"`
		Currency                  string         `json:"currency"`
		BudgetCode                string         `json:"budgetCode"`
		CostCenter                string         `json:"costCenter"`
		ProjectCode               string         `json:"projectCode"`
		WorkflowID                string         `json:"workflowId"`
		// "" = inherit from org, "goods_first" or "payment_first" to override per-PO
		ProcurementFlow string `json:"procurementFlow"`
	}

	if err := c.BodyParser(&req); err != nil {
		return utils.SendBadRequestError(c, "Invalid request body")
	}
	if req.RequisitionID == "" {
		return utils.SendBadRequestError(c, "requisitionId is required")
	}
	if len(req.Items) == 0 {
		return utils.SendBadRequestError(c, "At least one item is required")
	}
	if req.TotalAmount <= 0 {
		return utils.SendBadRequestError(c, "totalAmount must be greater than 0")
	}
	if req.Currency == "" {
		req.Currency = "ZMW"
	}

	// One-to-one: reject if any non-cancelled PO already exists for this REQ
	var existingPO models.PurchaseOrder
	if err := config.DB.
		Where("source_requisition_id = ? AND organization_id = ? AND UPPER(status) != 'CANCELLED'",
			req.RequisitionID, tenant.OrganizationID).
		First(&existingPO).Error; err == nil {
		return utils.SendConflictError(c, fmt.Sprintf(
			"Purchase order %s already exists for this requisition (status: %s).",
			existingPO.DocumentNumber, existingPO.Status))
	}

	// Load requisition (with preferred vendor) to compare for audit trail
	var requisition models.Requisition
	config.DB.Preload("PreferredVendor").Where("id = ? AND organization_id = ?", req.RequisitionID, tenant.OrganizationID).First(&requisition)

	// Enforce currency inheritance: PO must use the same currency as the source REQ
	if requisition.Currency != "" {
		req.Currency = requisition.Currency
	}

	// Verify vendor belongs to this org if provided
	var vendorIDPtr *string
	var vendor *models.Vendor
	if req.VendorID != "" {
		var v models.Vendor
		if err := config.DB.Where("id = ? AND organization_id = ?", req.VendorID, tenant.OrganizationID).First(&v).Error; err != nil {
			return utils.SendBadRequestError(c, "Vendor not found")
		}
		vendorIDPtr = &req.VendorID
		vendor = &v
	}

	documentNumber := utils.GenerateDocumentNumber("PO")
	orderID := uuid.New().String()

	var poFromReqUser models.User
	config.DB.Where("id = ?", tenant.UserID).First(&poFromReqUser)

	// Build PO metadata: copy REQ's attachments (tagged fromRequisition) + REQ's quotations
	poMetadata := map[string]interface{}{}
	if len(requisition.Metadata) > 0 {
		var reqMeta map[string]interface{}
		if err := json.Unmarshal(requisition.Metadata, &reqMeta); err == nil {
			// Copy attachments with fromRequisition flag
			if rawAttachments, ok := reqMeta["attachments"]; ok {
				if attachSlice, ok2 := rawAttachments.([]interface{}); ok2 {
					tagged := make([]interface{}, 0, len(attachSlice))
					for _, a := range attachSlice {
						if aMap, ok3 := a.(map[string]interface{}); ok3 {
							aMap["fromRequisition"] = true
							tagged = append(tagged, aMap)
						}
					}
					poMetadata["attachments"] = tagged
				}
			}
			// Copy quotations as-is
			if quotations, ok := reqMeta["quotations"]; ok {
				poMetadata["quotations"] = quotations
			}
		}
	}

	// Snapshot the vendor's compliance fields at PO creation time for audit
	// purposes. WARN-ONLY: a missing ZRA TPIN / PACRA number is surfaced via
	// complianceWarnings but never blocks PO creation.
	if vendor != nil {
		poMetadata["vendorCompliance"] = map[string]interface{}{
			"zraTpin":        vendor.ZraTpin,
			"pacraRegNumber": vendor.PacraRegNumber,
			"taxId":          vendor.TaxID,
			"snapshotAt":     time.Now().Format(time.RFC3339),
		}
		if w := vendorComplianceWarnings(vendor); len(w) > 0 {
			poMetadata["complianceWarnings"] = w
		}
	}

	// Set EstimatedCost from REQ total when REQ is marked as estimate
	estimatedCost := 0.0
	if requisition.IsEstimate {
		estimatedCost = requisition.TotalAmount
	}

	order := models.PurchaseOrder{
		ID:                orderID,
		OrganizationID:    tenant.OrganizationID,
		DocumentNumber:    documentNumber,
		VendorID:          vendorIDPtr,
		Status:            models.StatusDraft,
		TotalAmount:       req.TotalAmount,
		Currency:          req.Currency,
		ApprovalStage:     0,
		LinkedRequisition: req.RequisitionDocumentNumber,
		Title:             req.Title,
		Description:       req.Description,
		Department:        req.Department,
		DepartmentID:      req.DepartmentID,
		Priority:          req.Priority,
		BudgetCode:        req.BudgetCode,
		CostCenter:        req.CostCenter,
		ProjectCode:       req.ProjectCode,
		ProcurementFlow:   req.ProcurementFlow,
		EstimatedCost:     estimatedCost,
		CreatedBy:         tenant.UserID,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if len(poMetadata) > 0 {
		if metaBytes, err := json.Marshal(poMetadata); err == nil {
			order.Metadata = datatypes.JSON(metaBytes)
		}
	}

	if req.RequiredByDate != nil {
		order.RequiredByDate = req.RequiredByDate
	}
	if req.RequisitionID != "" {
		order.SourceRequisitionId = &req.RequisitionID
	}

	order.Items = datatypes.NewJSONType(req.Items)
	order.ApprovalHistory = datatypes.NewJSONType([]types.ApprovalRecord{})

	// Build initial action history — log vendor change + chain link
	var initialHistory []types.ActionHistoryEntry

	// Always record the chain origin
	if requisition.DocumentNumber != "" {
		initialHistory = append(initialHistory, types.ActionHistoryEntry{
			ID:          uuid.New().String(),
			Action:      "CREATED_FROM_REQUISITION",
			PerformedBy: tenant.UserID,
			Timestamp:   time.Now(),
			Metadata: map[string]interface{}{
				"linkedDocNumber": requisition.DocumentNumber,
				"linkedDocType":   "requisition",
			},
		})
	}

	reqPreferredVendorID := ""
	if requisition.PreferredVendorID != nil {
		reqPreferredVendorID = *requisition.PreferredVendorID
	}
	if req.VendorID != reqPreferredVendorID && reqPreferredVendorID != "" {
		oldVendorName := ""
		if requisition.PreferredVendor != nil {
			oldVendorName = requisition.PreferredVendor.Name
		}
		initialHistory = append(initialHistory, types.ActionHistoryEntry{
			ID:          uuid.New().String(),
			Action:      "VENDOR_CHANGED",
			PerformedBy: tenant.UserID,
			Timestamp:   time.Now(),
			ChangedFields: map[string]interface{}{
				"vendor": map[string]interface{}{
					"from": oldVendorName,
					"to":   req.VendorName,
				},
			},
		})
	}
	poFromReqNow := time.Now()
	initialHistory = append(initialHistory, types.ActionHistoryEntry{
		ID:              uuid.New().String(),
		Action:          "CREATE",
		ActionType:      "CREATE",
		PerformedBy:     tenant.UserID,
		PerformedByName: poFromReqUser.Name,
		PerformedByRole: poFromReqUser.Role,
		Timestamp:       poFromReqNow,
		PerformedAt:     poFromReqNow,
		Comments:        "Purchase order created from requisition",
		NewStatus:       models.StatusDraft,
	})
	order.ActionHistory = datatypes.NewJSONType(initialHistory)

	if err := config.DB.Create(&order).Error; err != nil {
		logging.LogError(c, err, "create_po_from_requisition_failed", nil)
		return utils.SendInternalError(c, "Failed to create purchase order", err)
	}

	// Record PO_CREATED on the source requisition for full chain traceability
	if req.RequisitionID != "" {
		reqHistory := requisition.ActionHistory.Data()
		reqHistory = append(reqHistory, types.ActionHistoryEntry{
			ID:          uuid.New().String(),
			Action:      "PO_CREATED",
			PerformedBy: tenant.UserID,
			Timestamp:   time.Now(),
			Metadata: map[string]interface{}{
				"linkedDocNumber": order.DocumentNumber,
				"linkedDocType":   "purchase_order",
			},
		})
		requisition.ActionHistory = datatypes.NewJSONType(reqHistory)
		config.DB.Save(&requisition)
	}

	config.DB.Preload("Vendor").First(&order)
	go utils.SyncDocumentAs(config.DB, "PURCHASE_ORDER", order.ID, tenant.UserID)
	go services.LogDocumentEvent(config.DB, services.DocumentEvent{
		OrganizationID: tenant.OrganizationID,
		DocumentID:     order.ID,
		DocumentType:   "purchase_order",
		UserID:         tenant.UserID,
		ActorName:      poFromReqUser.Name,
		ActorRole:      poFromReqUser.Role,
		Action:         "created",
		Details:        map[string]interface{}{"documentNumber": order.DocumentNumber, "sourceRequisition": req.RequisitionDocumentNumber},
	})

	logger.Info("po_from_requisition_created")
	return utils.SendCreatedSuccess(c, modelToPurchaseOrderResponse(order), "Purchase order created from requisition successfully")
}

// ============================================================================
// PAYMENT VOUCHER — FROM PURCHASE ORDER
// POST /api/v1/payment-vouchers/from-po
// ============================================================================

// CreatePaymentVoucherFromPO creates a PV pre-populated from an approved PO.
func CreatePaymentVoucherFromPO(c *fiber.Ctx) error {
	logger := logging.FromContext(c)
	logger.Info("create_pv_from_po_request")

	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return utils.SendUnauthorizedError(c, "Organization context required")
	}

	var req struct {
		PurchaseOrderID             string              `json:"purchaseOrderId"`
		PurchaseOrderDocumentNumber string              `json:"purchaseOrderDocumentNumber"`
		Title                       string              `json:"title"`
		Description                 string              `json:"description"`
		VendorID                    string              `json:"vendorId"`
		VendorName                  string              `json:"vendorName"`
		Department                  string              `json:"department"`
		DepartmentID                string              `json:"departmentId"`
		Items                       []types.PaymentItem `json:"items"`
		TotalAmount                 float64             `json:"totalAmount"`
		Currency                    string              `json:"currency"`
		BudgetCode                  string              `json:"budgetCode"`
		CostCenter                  string              `json:"costCenter"`
		ProjectCode                 string              `json:"projectCode"`
		SourceRequisitionID         string              `json:"sourceRequisitionId"`
		WorkflowID                  string              `json:"workflowId"`
		// Goods-first flow: required GRN document number (e.g. "GRN-20240101-001")
		LinkedGRNDocumentNumber string `json:"linkedGRNDocumentNumber"`
	}

	if err := c.BodyParser(&req); err != nil {
		return utils.SendBadRequestError(c, "Invalid request body")
	}
	if req.PurchaseOrderID == "" {
		return utils.SendBadRequestError(c, "purchaseOrderId is required")
	}
	if req.TotalAmount <= 0 {
		return utils.SendBadRequestError(c, "totalAmount must be greater than 0")
	}
	if req.Currency == "" {
		req.Currency = "ZMW"
	}

	// Load the PO with Vendor preload (needed for audit trail comparison)
	var po models.PurchaseOrder
	if err := config.DB.Preload("Vendor").Where("id = ? AND organization_id = ?", req.PurchaseOrderID, tenant.OrganizationID).First(&po).Error; err != nil {
		return utils.SendBadRequestError(c, "Purchase order not found")
	}

	// Enforce currency inheritance: PV must use the same currency as the linked PO
	if po.Currency != "" {
		req.Currency = po.Currency
	}

	// Enforce the shared PV-creation gate: PO must be APPROVED, no live duplicate
	// PV may exist, the amount is capped at the PO total (and received value in
	// goods-first), and in goods-first the linked GRN must be APPROVED or
	// COMPLETED. Single source of truth shared with the manual + auto paths.
	if msg, code := validateProcurementPVGate(config.DB, tenant.OrganizationID, po.DocumentNumber, req.LinkedGRNDocumentNumber, req.TotalAmount); code != 0 {
		return c.Status(code).JSON(fiber.Map{"success": false, "message": msg})
	}

	// Resolve effective flow + load the linked GRN (goods-first) purely for
	// audit-trail wiring below — all gating was done by the validator above.
	orgDefaultFlow := ""
	if strings.TrimSpace(po.ProcurementFlow) == "" {
		orgSvc := services.NewOrganizationService(config.DB)
		if s, _ := orgSvc.GetOrganizationSettings(tenant.OrganizationID); s != nil {
			orgDefaultFlow = s.ProcurementFlow
		}
	}
	effectiveFlow := utils.ResolveProcurementFlow(po.ProcurementFlow, orgDefaultFlow)

	// Load the linked GRN for audit wiring only when it exists AND belongs to
	// this PO — never persist a foreign/unvalidated GRN reference onto the PV.
	var linkedGRN *models.GoodsReceivedNote
	if req.LinkedGRNDocumentNumber != "" {
		var grn models.GoodsReceivedNote
		if err := config.DB.Where("document_number = ? AND organization_id = ?", req.LinkedGRNDocumentNumber, tenant.OrganizationID).First(&grn).Error; err == nil &&
			grn.PODocumentNumber == po.DocumentNumber {
			linkedGRN = &grn
		}
	}

	// Verify vendor belongs to this org if provided
	var vendorIDPtr *string
	if req.VendorID != "" {
		var vendor models.Vendor
		if err := config.DB.Where("id = ? AND organization_id = ?", req.VendorID, tenant.OrganizationID).First(&vendor).Error; err != nil {
			return utils.SendBadRequestError(c, "Vendor not found")
		}
		vendorIDPtr = &req.VendorID
	}

	documentNumber := utils.GenerateDocumentNumber("PV")
	invoiceRef := "INV-" + po.DocumentNumber

	var pvFromPOUser models.User
	config.DB.Where("id = ?", tenant.UserID).First(&pvFromPOUser)

	// Derive from the validated record (empty unless a PO-owned GRN loaded), so
	// payment_first PVs don't carry a bogus GRN link.
	linkedGRNDocNum := ""
	if linkedGRN != nil {
		linkedGRNDocNum = linkedGRN.DocumentNumber
	}

	voucher := models.PaymentVoucher{
		ID:             uuid.New().String(),
		OrganizationID: tenant.OrganizationID,
		DocumentNumber: documentNumber,
		VendorID:       vendorIDPtr,
		InvoiceNumber:  invoiceRef,
		Status:         models.StatusDraft,
		Amount:         req.TotalAmount,
		Currency:       req.Currency,
		PaymentMethod:  "bank_transfer",
		Description:    req.Description,
		ApprovalStage:  0,
		// Use the canonical, server-loaded PO document number so the duplicate-PV
		// guard (which keys on linked_po) stays reliable even if the client sent a
		// mismatched/empty purchaseOrderDocumentNumber.
		LinkedPO:  po.DocumentNumber,
		LinkedGRN: linkedGRNDocNum,
		Title:          req.Title,
		Department:     req.Department,
		DepartmentID:   req.DepartmentID,
		BudgetCode:     req.BudgetCode,
		CostCenter:     req.CostCenter,
		ProjectCode:    req.ProjectCode,
		CreatedBy:      tenant.UserID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if len(req.Items) > 0 {
		voucher.Items = datatypes.NewJSONType(req.Items)
	}
	voucher.ApprovalHistory = datatypes.NewJSONType([]types.ApprovalRecord{})

	// Build initial action history — chain origin + vendor change if applicable
	var pvInitialHistory []types.ActionHistoryEntry

	// Record which document this PV was created from
	if linkedGRN != nil {
		pvInitialHistory = append(pvInitialHistory, types.ActionHistoryEntry{
			ID:          uuid.New().String(),
			Action:      "CREATED_FROM_GRN",
			PerformedBy: tenant.UserID,
			Timestamp:   time.Now(),
			Metadata: map[string]interface{}{
				"linkedDocNumber": linkedGRN.DocumentNumber,
				"linkedDocType":   "grn",
				"flow":            "goods_first",
			},
		})
	} else {
		pvInitialHistory = append(pvInitialHistory, types.ActionHistoryEntry{
			ID:          uuid.New().String(),
			Action:      "CREATED_FROM_PO",
			PerformedBy: tenant.UserID,
			Timestamp:   time.Now(),
			Metadata: map[string]interface{}{
				"linkedDocNumber": po.DocumentNumber,
				"linkedDocType":   "purchase_order",
				"flow":            "payment_first",
			},
		})
	}

	poVendorID := ""
	if po.VendorID != nil {
		poVendorID = *po.VendorID
	}
	if req.VendorID != poVendorID && poVendorID != "" {
		oldVendorName := ""
		if po.Vendor != nil {
			oldVendorName = po.Vendor.Name
		}
		pvInitialHistory = append(pvInitialHistory, types.ActionHistoryEntry{
			ID:          uuid.New().String(),
			Action:      "VENDOR_CHANGED",
			PerformedBy: tenant.UserID,
			Timestamp:   time.Now(),
			ChangedFields: map[string]interface{}{
				"vendor": map[string]interface{}{
					"from": oldVendorName,
					"to":   req.VendorName,
				},
			},
		})
	}
	pvFromPONow := time.Now()
	pvInitialHistory = append(pvInitialHistory, types.ActionHistoryEntry{
		ID:              uuid.New().String(),
		Action:          "CREATE",
		ActionType:      "CREATE",
		PerformedBy:     tenant.UserID,
		PerformedByName: pvFromPOUser.Name,
		PerformedByRole: pvFromPOUser.Role,
		Timestamp:       pvFromPONow,
		PerformedAt:     pvFromPONow,
		Comments:        "Payment voucher created from purchase order",
		NewStatus:       models.StatusDraft,
	})
	voucher.ActionHistory = datatypes.NewJSONType(pvInitialHistory)

	if err := config.DB.Create(&voucher).Error; err != nil {
		logging.LogError(c, err, "create_pv_from_po_failed", nil)
		return utils.SendInternalError(c, "Failed to create payment voucher", err)
	}

	// Record PV_CREATED on the parent document (GRN for goods_first, PO for payment_first)
	pvCreatedEntry := types.ActionHistoryEntry{
		ID:          uuid.New().String(),
		Action:      "PV_CREATED",
		PerformedBy: tenant.UserID,
		Timestamp:   time.Now(),
		Metadata: map[string]interface{}{
			"linkedDocNumber": voucher.DocumentNumber,
			"linkedDocType":   "payment_voucher",
			"flow":            effectiveFlow,
		},
	}
	if linkedGRN != nil {
		grnHistory := linkedGRN.ActionHistory.Data()
		grnHistory = append(grnHistory, pvCreatedEntry)
		linkedGRN.ActionHistory = datatypes.NewJSONType(grnHistory)
		config.DB.Save(linkedGRN)
	} else {
		poHistory := po.ActionHistory.Data()
		poHistory = append(poHistory, pvCreatedEntry)
		po.ActionHistory = datatypes.NewJSONType(poHistory)
		config.DB.Save(&po)
	}

	config.DB.Preload("Vendor").First(&voucher)
	go utils.SyncDocumentAs(config.DB, "PAYMENT_VOUCHER", voucher.ID, tenant.UserID)
	go services.LogDocumentEvent(config.DB, services.DocumentEvent{
		OrganizationID: tenant.OrganizationID,
		DocumentID:     voucher.ID,
		DocumentType:   "payment_voucher",
		UserID:         tenant.UserID,
		ActorName:      pvFromPOUser.Name,
		ActorRole:      pvFromPOUser.Role,
		Action:         "created",
		Details:        map[string]interface{}{"documentNumber": voucher.DocumentNumber, "sourcePO": req.PurchaseOrderDocumentNumber},
	})

	logger.Info("pv_from_po_created")
	return utils.SendCreatedSuccess(c, modelToPaymentVoucherResponse(voucher), "Payment voucher created from purchase order successfully")
}

// ============================================================================
// PAYMENT VOUCHER — MARK PAID
// POST /api/v1/payment-vouchers/:id/mark-paid
// ============================================================================

// MarkPaymentVoucherPaid marks an approved PV as paid by completing its
// payment_execution workflow task. The task is auto-created when the PV's
// final approval stage completes, so callers no longer need to look it up
// themselves. Amount mismatches are still rejected here rather than deep in
// the workflow so the error surface stays specific.
//
// Behind the scenes this delegates to ApproveWorkflowTask, which means the
// PAID transition inherits the full workflow plumbing: claim check, optimistic
// locking, signed ActionHistory entry, SyncDocument, notification event.
func MarkPaymentVoucherPaid(c *fiber.Ctx) error {
	logger := logging.FromContext(c)
	logger.Info("mark_pv_paid_request")

	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return utils.SendUnauthorizedError(c, "Organization context required")
	}

	id := c.Params("id")
	if id == "" {
		return utils.SendBadRequestError(c, "Payment voucher ID is required")
	}

	var req struct {
		PaidAmount      float64    `json:"paidAmount"`
		PaidDate        *time.Time `json:"paidDate"` // accepted but not used; payment date is server-side now
		ReferenceNumber string     `json:"referenceNumber"`
		Comments        string     `json:"comments"`
		Signature       string     `json:"signature"`
	}

	if err := c.BodyParser(&req); err != nil {
		return utils.SendBadRequestError(c, "Invalid request body")
	}
	if req.PaidAmount <= 0 {
		return utils.SendBadRequestError(c, "paidAmount must be greater than 0")
	}
	if req.ReferenceNumber == "" {
		return utils.SendBadRequestError(c, "referenceNumber is required")
	}
	if req.Signature == "" {
		return utils.SendBadRequestError(c, "signature is required to execute payment")
	}

	var voucher models.PaymentVoucher
	if err := config.DB.Where("id = ? AND organization_id = ?", id, tenant.OrganizationID).First(&voucher).Error; err != nil {
		return utils.SendNotFoundError(c, "Payment voucher not found")
	}

	if strings.ToUpper(voucher.Status) != models.StatusApproved {
		return utils.SendBadRequestError(c, "Only approved payment vouchers can be marked as paid")
	}

	// Amount mismatch validation — paidAmount must match the approved voucher amount.
	// Kept at the handler edge so clients get a specific amount_mismatch error
	// code rather than a generic workflow error.
	const amountTolerance = 0.01 // allow for floating point rounding
	if req.PaidAmount < voucher.Amount-amountTolerance || req.PaidAmount > voucher.Amount+amountTolerance {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"success":        false,
			"error":          "amount_mismatch",
			"message":        fmt.Sprintf("Paid amount (%.2f) does not match the approved voucher amount (%.2f). Please enter the exact approved amount.", req.PaidAmount, voucher.Amount),
			"approvedAmount": voucher.Amount,
			"paidAmount":     req.PaidAmount,
		})
	}

	// Find the open payment_execution task created when the PV was fully approved.
	var task models.WorkflowTask
	if err := config.DB.Where(
		"entity_id = ? AND organization_id = ? AND kind = ? AND UPPER(status) IN ('PENDING','CLAIMED')",
		voucher.ID, tenant.OrganizationID, models.TaskKindPaymentExecution,
	).First(&task).Error; err != nil {
		return utils.SendBadRequestError(c, "No pending payment execution task found for this voucher. The PV may not yet be fully approved, or it may already be PAID.")
	}

	userID := c.Locals("userID").(string)

	// The workflow service enforces claim rules. If the task is unclaimed, auto-claim
	// it to the caller for convenience so a single button press can execute payment.
	if task.ClaimedBy == nil {
		claimNow := time.Now()
		claimExpiry := claimNow.Add(30 * time.Minute)
		config.DB.Model(&task).Updates(map[string]interface{}{
			"status":       "CLAIMED",
			"claimed_by":   userID,
			"claimed_at":   claimNow,
			"claim_expiry": claimExpiry,
			"version":      task.Version + 1,
		})
		task.Version++ // keep local copy in sync for the version-locked approve
	}

	workflowExecutionService := c.Locals("workflowExecutionService").(*services.WorkflowExecutionService)

	commentsWithRef := req.Comments
	if req.ReferenceNumber != "" {
		commentsWithRef = fmt.Sprintf("[Ref: %s] %s", req.ReferenceNumber, req.Comments)
	}

	if err := workflowExecutionService.ApproveWorkflowTaskWithVersion(
		c.Context(), task.ID, userID, req.Signature, commentsWithRef, task.Version,
	); err != nil {
		logging.LogError(c, err, "mark_pv_paid_failed", nil)
		return utils.SendInternalError(c, "Failed to mark payment voucher as paid", err)
	}

	// Reload for response
	if err := config.DB.Where("id = ?", voucher.ID).First(&voucher).Error; err != nil {
		return utils.SendInternalError(c, "Failed to reload payment voucher", err)
	}

	logger.Info("pv_marked_paid")
	return utils.SendSimpleSuccess(c, modelToPaymentVoucherResponse(voucher), "Payment voucher marked as paid successfully")
}

// ============================================================================
// STATS ENDPOINTS
// ============================================================================

// GetRequisitionStats returns count summaries for requisitions in the org.
// GET /api/v1/requisitions/stats
func GetRequisitionStats(c *fiber.Ctx) error {
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return utils.SendUnauthorizedError(c, "Organization context required")
	}

	db := config.DB
	scope := utils.GetDocumentScope(db, tenant.UserID, tenant.UserRole, tenant.OrganizationID)
	base := db.Model(&models.Requisition{}).Where("organization_id = ?", tenant.OrganizationID)
	base = scope.ApplyToQuery(base, "requester_id", "requisition", "")

	stats := fiber.Map{}
	for _, status := range []string{"draft", "pending", "approved", "rejected", "completed", "cancelled"} {
		var count int64
		base.Session(&gorm.Session{}).Where("UPPER(status) = ?", strings.ToUpper(status)).Count(&count)
		stats[status] = count
	}

	var total int64
	base.Session(&gorm.Session{}).Count(&total)
	stats["total"] = total

	return utils.SendSimpleSuccess(c, stats, "Requisition statistics retrieved successfully")
}

// GetPurchaseOrderStats returns count summaries for POs in the org.
// GET /api/v1/purchase-orders/stats
func GetPurchaseOrderStats(c *fiber.Ctx) error {
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return utils.SendUnauthorizedError(c, "Organization context required")
	}

	db := config.DB
	scope := utils.GetDocumentScope(db, tenant.UserID, tenant.UserRole, tenant.OrganizationID)
	base := db.Model(&models.PurchaseOrder{}).Where("organization_id = ?", tenant.OrganizationID)
	base = scope.ApplyToQuery(base, "created_by", "purchase_order", "")

	stats := fiber.Map{}
	for _, status := range []string{"draft", "pending", "approved", "rejected", "fulfilled", "completed", "cancelled"} {
		var count int64
		base.Session(&gorm.Session{}).Where("UPPER(status) = ?", strings.ToUpper(status)).Count(&count)
		stats[status] = count
	}

	var total int64
	base.Session(&gorm.Session{}).Count(&total)
	stats["total"] = total

	return utils.SendSimpleSuccess(c, stats, "Purchase order statistics retrieved successfully")
}

// GetPaymentVoucherStats returns count summaries for PVs in the org.
// GET /api/v1/payment-vouchers/stats
func GetPaymentVoucherStats(c *fiber.Ctx) error {
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return utils.SendUnauthorizedError(c, "Organization context required")
	}

	db := config.DB
	scope := utils.GetDocumentScope(db, tenant.UserID, tenant.UserRole, tenant.OrganizationID)
	base := db.Model(&models.PaymentVoucher{}).Where("organization_id = ?", tenant.OrganizationID)
	base = scope.ApplyToQuery(base, "vendor_id", "payment_voucher", "")

	stats := fiber.Map{}
	for _, status := range []string{"draft", "pending", "approved", "rejected", "paid", "completed", "cancelled"} {
		var count int64
		base.Session(&gorm.Session{}).Where("UPPER(status) = ?", strings.ToUpper(status)).Count(&count)
		stats[status] = count
	}

	var total int64
	base.Session(&gorm.Session{}).Count(&total)
	stats["total"] = total

	return utils.SendSimpleSuccess(c, stats, "Payment voucher statistics retrieved successfully")
}

// ============================================================================
// DEPARTMENT HEADS LIST
// GET /api/v1/users/department-heads/list
// ============================================================================

// GetDepartmentHeadsList returns organization members with roles that can act as approvers/HODs.
func GetDepartmentHeadsList(c *fiber.Ctx) error {
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return utils.SendUnauthorizedError(c, "Organization context required")
	}

	departmentID := c.Query("department_id")
	roleID := c.Query("role_id")
	isActiveStr := c.Query("is_active")
	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("page_size", 50)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	// Roles eligible to be department heads / approvers
	eligibleRoles := []string{"admin", "approver", "finance"}

	query := config.DB.Table("users u").
		Select("u.id, u.name, u.email, u.role, u.position, om.department_id").
		Joins("JOIN organization_members om ON om.user_id = u.id").
		Where("om.organization_id = ? AND om.active = true AND u.role IN ?", tenant.OrganizationID, eligibleRoles)

	if departmentID != "" {
		query = query.Where("om.department_id = ?", departmentID)
	}
	if roleID != "" {
		query = query.Where("u.role = ?", roleID)
	}
	if isActiveStr == "true" {
		query = query.Where("u.active = true")
	} else if isActiveStr == "false" {
		query = query.Where("u.active = false")
	}

	var total int64
	query.Count(&total)

	type HODUser struct {
		ID           string `json:"id"`
		Name         string `json:"name"`
		Email        string `json:"email"`
		Role         string `json:"role"`
		Position     string `json:"position"`
		DepartmentID string `json:"departmentId"`
	}

	var users []HODUser
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Scan(&users).Error; err != nil {
		return utils.SendInternalError(c, "Failed to retrieve department heads", err)
	}

	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)
	return utils.SendSuccess(c, fiber.StatusOK, users, "Department heads retrieved successfully", &types.PaginationMeta{
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    int64(page) < totalPages,
		HasPrev:    page > 1,
	})
}

// ============================================================================
// SIGNATURE VALIDATION
// POST /api/v1/approvals/validate-signature
// ============================================================================

// ValidateSignature checks that a submitted digital signature is non-empty and well-formed.
func ValidateSignature(c *fiber.Ctx) error {
	var req struct {
		Signature string `json:"signature"`
		UserID    string `json:"userId"`
	}

	if err := c.BodyParser(&req); err != nil {
		return utils.SendBadRequestError(c, "Invalid request body")
	}
	if req.Signature == "" {
		return utils.SendBadRequestError(c, "Signature is required")
	}

	// Accept both plain base64 and data-URI format (data:image/png;base64,...)
	raw := req.Signature
	if idx := strings.Index(raw, "base64,"); idx != -1 {
		raw = raw[idx+7:]
	}

	_, decodeErr := base64.StdEncoding.DecodeString(raw)
	if decodeErr != nil {
		// Try URL-safe variant
		_, decodeErr = base64.URLEncoding.DecodeString(raw)
	}

	if decodeErr != nil {
		return utils.SendSimpleSuccess(c, fiber.Map{
			"valid":   false,
			"message": "Signature is not valid base64 encoded data",
		}, "Signature validation completed")
	}

	return utils.SendSimpleSuccess(c, fiber.Map{
		"valid":   true,
		"message": "Signature is valid",
	}, "Signature validation completed")
}

// ============================================================================
// APPROVER WORKLOAD
// GET /api/v1/approvals/approver-workload/:approverId
// ============================================================================

// GetApproverWorkload returns pending task count and basic stats for a given approver.
func GetApproverWorkload(c *fiber.Ctx) error {
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return utils.SendUnauthorizedError(c, "Organization context required")
	}

	approverID := c.Params("approverId")
	if approverID == "" {
		return utils.SendBadRequestError(c, "Approver ID is required")
	}

	db := config.DB

	// Count pending tasks assigned to this approver in this org
	var pendingCount int64
	db.Table("workflow_tasks wt").
		Joins("JOIN workflow_assignments wa ON wa.id = wt.workflow_assignment_id").
		Where("wt.assigned_to = ? AND UPPER(wt.status) = 'PENDING' AND wa.organization_id = ?", approverID, tenant.OrganizationID).
		Count(&pendingCount)

	// Count tasks completed this month
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	var completedThisMonth int64
	db.Table("workflow_tasks wt").
		Joins("JOIN workflow_assignments wa ON wa.id = wt.workflow_assignment_id").
		Where("wt.assigned_to = ? AND UPPER(wt.status) IN ? AND wt.updated_at >= ? AND wa.organization_id = ?",
			approverID, []string{"APPROVED", "REJECTED"}, startOfMonth, tenant.OrganizationID).
		Count(&completedThisMonth)

	// Count overdue tasks (past due_date and still pending)
	var overdueTasks int64
	db.Table("workflow_tasks wt").
		Joins("JOIN workflow_assignments wa ON wa.id = wt.workflow_assignment_id").
		Where("wt.assigned_to = ? AND UPPER(wt.status) = 'PENDING' AND wt.due_date < ? AND wa.organization_id = ?",
			approverID, now, tenant.OrganizationID).
		Count(&overdueTasks)

	return utils.SendSimpleSuccess(c, fiber.Map{
		"pendingCount":        pendingCount,
		"averageResponseTime": 0, // would require time-series aggregation
		"completedThisMonth":  completedThisMonth,
		"overdueTasks":        overdueTasks,
	}, "Approver workload retrieved successfully")
}

// GRN /confirm endpoint removed: workflow approval now auto-transitions the
// GRN to COMPLETED via cascadeGRNApprovalToPO + status update in
// workflow_execution_service.go. The skip-workflow path uses MarkGRNComplete.

// ============================================================================
// PAYMENT VOUCHER — MARK PAID WITH PROOF OF PAYMENT (direct-payment flow)
// POST /api/v1/payment-vouchers/:id/mark-paid-with-pop
// ============================================================================

// allowedPOPExtensions lists the file extensions accepted as proof of payment.
var allowedPOPExtensions = map[string]bool{
	".pdf":  true,
	".jpg":  true,
	".jpeg": true,
	".png":  true,
}

const maxPOPSizeBytes = 10 * 1024 * 1024 // 10 MB

// MarkPaidWithPOP marks an APPROVED PaymentVoucher as PAID by storing an
// uploaded proof-of-payment file (PDF/JPG/PNG, ≤ 10 MB).
//
// Unlike the workflow-task-based MarkPaymentVoucherPaid this handler is
// purpose-built for the direct_payment routing where there are no approval
// stages — the file is the audit artefact. The file bytes are stored as
// base64 in the proof_of_payment JSON column (no external storage dependency).
//
// Business rules:
//   - Caller must be finance or admin.
//   - PV must be in APPROVED status (409 otherwise).
//   - popFile multipart field is required (400 otherwise).
//   - File must be ≤ 10 MB and have an allowed extension.
//   - On success: status → PAID, proof_of_payment populated, paid_at / paid_by set.
//   - If PV routing_type is direct_payment and metadata.sourceReqID is set,
//     the linked requisition status is cascaded to COMPLETED.
func MarkPaidWithPOP(c *fiber.Ctx) error {
	logger := logging.FromContext(c)
	logger.Info("mark_paid_with_pop_request")

	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return utils.SendUnauthorizedError(c, "Organization context required")
	}

	// Authorization is enforced by the route guard, which requires the dedicated
	// "payment_voucher.pay" permission (separate from "approve" for separation of
	// duties). Held by built-in admin/super_admin/finance AND any custom org role
	// the tenant grants it — so custom "payments" roles can disburse without being
	// hard-coded here, and approvers cannot pay what they approved.

	id := c.Params("id")
	if id == "" {
		return utils.SendBadRequestError(c, "Payment voucher ID is required")
	}

	// Load PV.
	var voucher models.PaymentVoucher
	if err := config.DB.Where("id = ? AND organization_id = ?", id, tenant.OrganizationID).First(&voucher).Error; err != nil {
		return utils.SendNotFoundError(c, "Payment voucher not found")
	}

	// Must be APPROVED — return 409 Conflict so callers can distinguish
	// "wrong state" from "bad request".
	if strings.ToUpper(voucher.Status) != models.StatusApproved {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"success": false,
			"message": fmt.Sprintf("Payment voucher must be in APPROVED status to mark as paid (current: %s)", voucher.Status),
		})
	}

	// Optional amount-match: when the client supplies paidAmount it must equal
	// the approved voucher amount (within 0.01). Mirrors MarkPaymentVoucherPaid
	// so the proof-of-payment path enforces the same financial-integrity check.
	if pa := strings.TrimSpace(c.FormValue("paidAmount")); pa != "" {
		parsed, perr := strconv.ParseFloat(pa, 64)
		if perr != nil {
			return utils.SendBadRequestError(c, "paidAmount must be a number")
		}
		if parsed < voucher.Amount-0.01 || parsed > voucher.Amount+0.01 {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
				"success":        false,
				"error":          "amount_mismatch",
				"message":        fmt.Sprintf("Paid amount (%.2f) does not match the approved voucher amount (%.2f). Please enter the exact approved amount.", parsed, voucher.Amount),
				"approvedAmount": voucher.Amount,
				"paidAmount":     parsed,
			})
		}
	}

	// Require popFile field.
	fileHeader, err := c.FormFile("popFile")
	if err != nil || fileHeader == nil {
		return utils.SendBadRequestError(c, "popFile is required (PDF, JPG, or PNG, max 10 MB)")
	}

	// Validate size.
	if fileHeader.Size > maxPOPSizeBytes {
		return utils.SendBadRequestError(c, fmt.Sprintf("popFile exceeds maximum size of 10 MB (got %.2f MB)", float64(fileHeader.Size)/1024/1024))
	}

	// Validate extension.
	dotIdx := strings.LastIndex(fileHeader.Filename, ".")
	ext := ""
	if dotIdx >= 0 {
		ext = strings.ToLower(fileHeader.Filename[dotIdx:])
	}
	if !allowedPOPExtensions[ext] {
		return utils.SendBadRequestError(c, "popFile must be a PDF, JPG, or PNG file")
	}

	// Read file bytes and encode as base64 for DB storage.
	f, err := fileHeader.Open()
	if err != nil {
		return utils.SendInternalError(c, "Failed to open uploaded file", err)
	}
	defer f.Close()

	var buf = make([]byte, fileHeader.Size)
	if _, err := io.ReadFull(f, buf); err != nil {
		return utils.SendInternalError(c, "Failed to read uploaded file", err)
	}
	encoded := base64.StdEncoding.EncodeToString(buf)

	// Optional paidDate (RFC3339 or date-only).
	var paidAt time.Time
	if paidDateStr := c.FormValue("paidDate"); paidDateStr != "" {
		if t, parseErr := time.Parse("2006-01-02", paidDateStr); parseErr == nil {
			paidAt = t
		} else if t, parseErr := time.Parse(time.RFC3339, paidDateStr); parseErr == nil {
			paidAt = t
		}
	}
	if paidAt.IsZero() {
		paidAt = time.Now()
	}

	notes := c.FormValue("notes")

	// Build proof-of-payment payload.
	popID := uuid.New().String()
	popPayload := map[string]interface{}{
		"id":          popID,
		"fileName":    fileHeader.Filename,
		"mimeType":    fileHeader.Header.Get("Content-Type"),
		"sizeBytes":   fileHeader.Size,
		"dataBase64":  encoded,
		"uploadedAt":  time.Now().Format(time.RFC3339),
		"uploadedBy":  tenant.UserID,
	}
	popBytes, _ := json.Marshal(popPayload)

	userID := tenant.UserID
	now := time.Now()

	// Persist in a transaction.
	err = config.DB.Transaction(func(tx *gorm.DB) error {
		// Load actor for action history.
		var actor models.User
		tx.Where("id = ?", userID).First(&actor)

		voucher.Status = models.StatusPaid
		voucher.ProofOfPayment = datatypes.JSON(popBytes)
		voucher.PaidAt = &paidAt
		voucher.PaidBy = &userID
		voucher.UpdatedAt = now

		actionHistory := voucher.ActionHistory.Data()
		actionHistory = append(actionHistory, types.ActionHistoryEntry{
			ID:              uuid.New().String(),
			Action:          "MARK_PAID",
			ActionType:      "MARK_PAID",
			PerformedBy:     userID,
			PerformedByName: actor.Name,
			PerformedByRole: actor.Role,
			Timestamp:       now,
			PerformedAt:     now,
			Comments:        notes,
			PreviousStatus:  models.StatusApproved,
			NewStatus:       models.StatusPaid,
		})
		voucher.ActionHistory = datatypes.NewJSONType(actionHistory)

		if err := tx.Save(&voucher).Error; err != nil {
			return fmt.Errorf("update voucher: %w", err)
		}

		// Close any open payment_execution task for this PV so the proof-of-
		// payment path doesn't leave a dangling task (procurement PVs get one at
		// final approval). Idempotent — direct_payment PVs have none.
		tx.Model(&models.WorkflowTask{}).
			Where("entity_id = ? AND organization_id = ? AND kind = ? AND UPPER(status) IN ('PENDING','CLAIMED')",
				voucher.ID, voucher.OrganizationID, models.TaskKindPaymentExecution).
			Updates(map[string]interface{}{
				"status":       "COMPLETED",
				"completed_at": now,
				"updated_by":   userID,
				"updated_at":   now,
			})

		// Cascade requisition to COMPLETED for direct_payment chain.
		if strings.EqualFold(voucher.RoutingType, models.RoutingTypeDirectPayment) && len(voucher.Metadata) > 0 {
			var meta map[string]interface{}
			if json.Unmarshal(voucher.Metadata, &meta) == nil {
				if srcReqID, ok := meta["sourceReqID"].(string); ok && srcReqID != "" {
					tx.Model(&models.Requisition{}).
						Where("id = ? AND organization_id = ?", srcReqID, voucher.OrganizationID).
						Updates(map[string]interface{}{
							"status":     models.StatusCompleted,
							"updated_at": now,
						})
				}
			}
		}

		// Cascade parent PO to COMPLETED when delivery is full + all PVs PAID
		// (closes both procurement flows: goods_first and payment_first).
		if wfSvc, ok := c.Locals("workflowExecutionService").(*services.WorkflowExecutionService); ok && wfSvc != nil {
			if err := wfSvc.CascadePVPaidToPO(tx, voucher.ID); err != nil {
				return fmt.Errorf("cascade PV→PO: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		logging.LogError(c, err, "mark_paid_with_pop_failed", nil)
		return utils.SendInternalError(c, "Failed to mark payment voucher as paid", err)
	}

	go utils.SyncDocumentAs(config.DB, "PAYMENT_VOUCHER", voucher.ID, userID)

	logger.Info("pv_marked_paid_with_pop")
	return utils.SendSimpleSuccess(c, modelToPaymentVoucherResponse(voucher), "Payment voucher marked as paid successfully")
}

// ============================================================================
// PAYMENT VOUCHER — RECOVER FROM PO
// POST /api/v1/payment-vouchers/recover-from-po/:poId
// ============================================================================

// RecoverPVFromPO creates a draft PaymentVoucher for an existing direct_payment
// PurchaseOrder whose auto-creation failed. Idempotent: if a PV already exists
// for the PO, returns 200 with the existing PV instead of 201.
//
// Only admin or finance callers may trigger recovery.
func RecoverPVFromPO(c *fiber.Ctx) error {
	logger := logging.FromContext(c)
	logger.Info("recover_pv_from_po_request")

	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return utils.SendUnauthorizedError(c, "Organization context required")
	}

	// Role gate: only admin or finance (super_admin is not a gateway role).
	userRole, _ := c.Locals("userRole").(string)
	if userRole != "admin" && userRole != "finance" {
		return utils.SendForbiddenError(c, "Only admin or finance users can recover a payment voucher from a PO")
	}

	poID := c.Params("poId")
	if poID == "" {
		return utils.SendBadRequestError(c, "PO ID is required")
	}

	// Load PO scoped to org.
	var po models.PurchaseOrder
	if err := config.DB.Where("id = ? AND organization_id = ?", poID, tenant.OrganizationID).First(&po).Error; err != nil {
		return utils.SendNotFoundError(c, "Purchase order not found")
	}

	// Reject non-direct_payment POs.
	if !strings.EqualFold(po.RoutingType, models.RoutingTypeDirectPayment) {
		return utils.SendBadRequestError(c, fmt.Sprintf(
			"PO %s has routing_type %q — recovery is only supported for direct_payment POs",
			po.DocumentNumber, po.RoutingType,
		))
	}

	// Idempotency check: return existing PV if one already exists.
	var existingPV models.PaymentVoucher
	if err := config.DB.
		Where("linked_po = ? AND organization_id = ? AND UPPER(status) != ?",
			po.DocumentNumber, tenant.OrganizationID, models.StatusCancelled).
		First(&existingPV).Error; err == nil {
		// PV already exists — return 200.
		return utils.SendSimpleSuccess(c, modelToPaymentVoucherResponse(existingPV), "Payment voucher already exists for this PO")
	}

	// Create draft PV mirroring autoCreateDraftPV logic.
	pvDocNum := utils.GenerateDocumentNumber("PV")
	now := time.Now()

	metaPayload := map[string]interface{}{
		"autoCreated": false,
		"recoveredBy": tenant.UserID,
	}
	metaBytes, _ := json.Marshal(metaPayload)

	vendorName := po.VendorName
	if vendorName == "" {
		vendorName = "Direct Payment"
	}

	pv := models.PaymentVoucher{
		ID:             uuid.New().String(),
		DocumentNumber: pvDocNum,
		OrganizationID: tenant.OrganizationID,
		Status:         models.StatusDraft,
		CreatedBy:      tenant.UserID,
		LinkedPO:       po.DocumentNumber,
		RoutingType:    models.RoutingTypeDirectPayment,
		VendorName:     vendorName,
		Amount:         po.TotalAmount,
		Currency:       po.Currency,
		Metadata:       datatypes.JSON(metaBytes),
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	pv.ApprovalHistory = datatypes.NewJSONType([]types.ApprovalRecord{})
	pv.ActionHistory = datatypes.NewJSONType([]types.ActionHistoryEntry{{
		ID:          uuid.New().String(),
		Action:      "RECOVER",
		ActionType:  "CREATE",
		PerformedBy: tenant.UserID,
		Timestamp:   now,
		PerformedAt: now,
		Comments:    "Draft PV recovered from PO " + po.DocumentNumber,
		NewStatus:   models.StatusDraft,
	}})

	if err := config.DB.Create(&pv).Error; err != nil {
		logging.LogError(c, err, "recover_pv_from_po_failed", nil)
		return utils.SendInternalError(c, "Failed to create payment voucher", err)
	}

	logger.Info("pv_recovered_from_po")
	return c.Status(fiber.StatusCreated).JSON(types.DetailResponse{
		Success: true,
		Data:    modelToPaymentVoucherResponse(pv),
	})
}
