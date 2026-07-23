package handlers

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/config"
	db "github.com/liyali/liyali-gateway/database/sqlc"
	"github.com/liyali/liyali-gateway/logging"
	"github.com/liyali/liyali-gateway/middleware"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/services"
	"github.com/liyali/liyali-gateway/types"
	"github.com/liyali/liyali-gateway/utils"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// validateProcurementPVGate enforces the rules for creating a Payment Voucher
// against a linked Purchase Order (and GRN in goods-first flow):
//   - the linked PO must exist and be APPROVED or FULFILLED (FULFILLED means
//     goods are fully delivered but the balance is still outstanding — a
//     partially-paid PO parks there and must still be able to receive the
//     remaining PV(s); see CascadePVPaidToPO, which accepts the same pair);
//   - in goods-first, an APPROVED or COMPLETED GRN must back the PV and the
//     amount may not exceed the *remaining* received value (received value
//     minus what earlier live PVs already committed);
//   - the amount may not push the PO's running committed total past its
//     total — multiple partial PVs against the same PO are allowed as long as
//     their sum stays within the PO's remaining balance. CANCELLED/REJECTED
//     PVs never consume budget, so a fresh PV can always retry a failed one.
//
// Returns ("", 0) when valid, otherwise (message, httpStatus). It is the single
// source of truth shared by the manual, from-PO, and auto-create PV paths.
// `tx` is any query handle (config.DB or a transaction). The PO row is locked
// FOR UPDATE below, so callers creating a PV should pass a transaction and run
// the gate + the PV insert inside it: that's what makes concurrent requests
// against the same PO serialize instead of racing past the remaining-balance
// check now that the DB unique index capping a PO at one live PV is gone
// (migration 023). SQLite (test harness) ignores the lock — single connection,
// no real concurrency to race.
func validateProcurementPVGate(tx *gorm.DB, orgID, linkedPO, linkedGRN string, amount float64) (string, int) {
	if linkedPO == "" {
		return "", 0
	}

	var po models.PurchaseOrder
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("document_number = ? AND organization_id = ?", linkedPO, orgID).
		First(&po).Error; err != nil {
		return "Linked purchase order not found", fiber.StatusBadRequest
	}
	poStatus := strings.ToUpper(po.Status)
	if poStatus != "APPROVED" && poStatus != models.StatusFulfilled {
		return fmt.Sprintf("Cannot create PV: linked PO %s is in %s status and must be APPROVED or FULFILLED first.", linkedPO, po.Status), fiber.StatusBadRequest
	}

	// Committed = Σ amount of every live (non CANCELLED/REJECTED) PV already
	// linked to this PO. Multiple partial PVs are fine as long as the new one
	// doesn't push the running total past the PO (or GRN) ceiling below.
	sum, err := services.ComputePOPaymentSummary(tx, orgID, linkedPO)
	if err != nil {
		return "Failed to compute PO payment summary", fiber.StatusInternalServerError
	}

	// Resolve effective flow: PO override → org default → "goods_first".
	orgDefaultFlow := ""
	if strings.TrimSpace(po.ProcurementFlow) == "" {
		orgSvc := services.NewOrganizationService(tx)
		if s, _ := orgSvc.GetOrganizationSettings(orgID); s != nil {
			orgDefaultFlow = s.ProcurementFlow
		}
	}
	effectiveFlow := utils.ResolveProcurementFlow(po.ProcurementFlow, orgDefaultFlow)

	if effectiveFlow == "goods_first" {
		if linkedGRN == "" {
			return fmt.Sprintf("Cannot create PV for PO %s: goods-first flow requires an APPROVED GRN first. Link the GRN via linkedGRN.", linkedPO), fiber.StatusBadRequest
		}
		var grn models.GoodsReceivedNote
		if err := tx.Where("document_number = ? AND organization_id = ?", linkedGRN, orgID).First(&grn).Error; err != nil {
			return "Linked GRN not found", fiber.StatusBadRequest
		}
		if grn.PODocumentNumber != linkedPO {
			return fmt.Sprintf("GRN %s is not linked to PO %s", linkedGRN, linkedPO), fiber.StatusBadRequest
		}
		// COMPLETED is the terminal GRN state (workflow auto-advances past
		// APPROVED). Both states satisfy the "goods received" gate.
		grnStatus := strings.ToUpper(grn.Status)
		if grnStatus != "APPROVED" && grnStatus != "COMPLETED" {
			return fmt.Sprintf("Cannot create PV: linked GRN %s is in %s status and must be APPROVED or COMPLETED first.", linkedGRN, grn.Status), fiber.StatusBadRequest
		}

		// Over-invoicing guard: PV amount may not exceed the *remaining*
		// received value, computed PO-wide (not just against the single
		// linkedGRN above) as Σ over every confirmed GRN linked to this PO of
		// (grnItem.quantityReceived × poItem.unitPrice), matched by
		// description, minus what's already committed. PO-wide is required
		// for multi-delivery POs: a second/third GRN's received value on its
		// own is usually already fully "committed" by an earlier PV, which
		// would wrongly block that delivery's PV if only that one GRN's value
		// were considered. Only APPROVED/COMPLETED GRNs count — the same bar
		// the specific linkedGRN must meet above. A DRAFT/PENDING GRN is an
		// unverified delivery claim; letting it raise the payment ceiling
		// would allow invoicing against goods no approver has confirmed.
		poItems := po.Items.Data()
		unitPriceByDesc := make(map[string]float64, len(poItems))
		for _, pi := range poItems {
			unitPriceByDesc[pi.Description] = pi.UnitPrice
		}
		var poGRNs []models.GoodsReceivedNote
		if err := tx.Where("po_document_number = ? AND organization_id = ? AND UPPER(status) IN ?",
			linkedPO, orgID, []string{"APPROVED", "COMPLETED"}).Find(&poGRNs).Error; err != nil {
			return "Failed to compute received value", fiber.StatusInternalServerError
		}
		var receivedValue float64
		for _, g := range poGRNs {
			for _, gi := range g.Items.Data() {
				receivedValue += float64(gi.QuantityReceived) * unitPriceByDesc[gi.Description]
			}
		}
		if amount > receivedValue-sum.Committed+0.01 {
			return fmt.Sprintf("PV amount %.2f exceeds remaining received value %.2f across %d GRN(s) on PO %s (received value %.2f, already committed %.2f across %d voucher(s)). Adjust the invoice amount to match goods actually received.",
				amount, receivedValue-sum.Committed, len(poGRNs), linkedPO, receivedValue, sum.Committed, sum.LivePVs), fiber.StatusBadRequest
		}
	}

	// Final backstop: the sum of every live PV against an approved PO can
	// never exceed that PO's total. This remaining-balance cap replaces the
	// old "one live PV per PO" rule — it's what makes partial payments
	// (multiple PVs against a single PO) possible.
	remaining := po.TotalAmount - sum.Committed
	if amount > remaining+0.01 {
		return fmt.Sprintf("PV amount %.2f exceeds remaining balance %.2f on PO %s (PO total %.2f, already committed %.2f across %d voucher(s)).",
			amount, remaining, linkedPO, po.TotalAmount, sum.Committed, sum.LivePVs), fiber.StatusBadRequest
	}
	return "", 0
}

// pvGateError carries an HTTP status + message out of a config.DB.Transaction
// closure so the outer HTTP handler can turn it back into the original
// response after the transaction rolls back. Used by the two PV-creation
// handlers that wrap validateProcurementPVGate + the PV insert in a
// transaction (see the gate's doc comment for why).
type pvGateError struct {
	status int
	msg    string
}

func (e *pvGateError) Error() string { return e.msg }

// GetPaymentVouchers retrieves all payment vouchers with pagination and filtering
func GetPaymentVouchers(c *fiber.Ctx) error {
	logger := logging.FromContext(c)
	logger.Info("get_payment_vouchers_request")

	// Get organization context from tenant middleware
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Organization context required",
			"error":   err.Error(),
		})
	}

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	status := c.Query("status")
	vendorID := c.Query("vendorId")
	hasPoP := c.Query("hasProofOfPayment")

	// Add query parameters to context
	logging.AddFieldsToRequest(c, map[string]interface{}{
		"operation":            "get_payment_vouchers",
		"page":                 page,
		"limit":                limit,
		"status":               status,
		"vendor_id":            vendorID,
		"has_proof_of_payment": hasPoP,
		"organization_id":      tenant.OrganizationID,
	})

	scope := utils.GetDocumentScope(config.DB, tenant.UserID, tenant.UserRole, tenant.OrganizationID)

	ctx := c.Context()
	offset := int32((page - 1) * limit)
	orgRoleIDs := scope.OrgRoleIDs
	if orgRoleIDs == nil {
		orgRoleIDs = []string{}
	}

	var total int64
	var ids []string

	if config.Queries == nil {
		total, ids, err = utils.ListDocumentIDsFallback(config.DB, "payment_vouchers", utils.DocumentListFilters{
			OrganizationID:    tenant.OrganizationID,
			Status:            status,
			RefField:          "vendor_id",
			RefValue:          vendorID,
			HideDirectPayment: scope.HideDirectPayment,
		}, scope, limit, int(offset))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to fetch payment vouchers",
				"error":   err.Error(),
			})
		}
	} else {
		switch {
		case scope.CanViewAll:
		total, err = config.Queries.CountPaymentVouchersAll(ctx, db.CountPaymentVouchersAllParams{
			OrganizationID:    tenant.OrganizationID,
			Column2:           status,
			Column3:           vendorID,
			HideDirectPayment: scope.HideDirectPayment,
		})
		if err != nil {
			logging.LogError(c, err, "failed_to_count_payment_vouchers", map[string]interface{}{"error_type": "database_error"})
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to count payment vouchers",
				"error":   err.Error(),
			})
		}
		ids, err = config.Queries.ListPaymentVoucherIDsAll(ctx, db.ListPaymentVoucherIDsAllParams{
			OrganizationID:    tenant.OrganizationID,
			Column2:           status,
			Column3:           vendorID,
			HideDirectPayment: scope.HideDirectPayment,
			Limit:             int32(limit),
			Offset:            offset,
		})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to fetch payment vouchers",
				"error":   err.Error(),
			})
		}
	case scope.IsProcurement:
		total, err = config.Queries.CountPaymentVouchersProcurement(ctx, db.CountPaymentVouchersProcurementParams{
			OrganizationID:    tenant.OrganizationID,
			Column2:           status,
			Column3:           vendorID,
			HideDirectPayment: scope.HideDirectPayment,
		})
		if err != nil {
			logging.LogError(c, err, "failed_to_count_payment_vouchers", map[string]interface{}{"error_type": "database_error"})
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to count payment vouchers",
				"error":   err.Error(),
			})
		}
		ids, err = config.Queries.ListPaymentVoucherIDsProcurement(ctx, db.ListPaymentVoucherIDsProcurementParams{
			OrganizationID:    tenant.OrganizationID,
			Column2:           status,
			Column3:           vendorID,
			HideDirectPayment: scope.HideDirectPayment,
			Limit:             int32(limit),
			Offset:            offset,
		})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to fetch payment vouchers",
				"error":   err.Error(),
			})
		}
	default:
		total, err = config.Queries.CountPaymentVouchersLimited(ctx, db.CountPaymentVouchersLimitedParams{
			OrganizationID: tenant.OrganizationID,
			Column2:        status,
			Column3:        vendorID,
			CreatedBy:      &scope.UserID,
			Lower:          scope.UserRole,
			Column6:        orgRoleIDs,
		})
		if err != nil {
			logging.LogError(c, err, "failed_to_count_payment_vouchers", map[string]interface{}{"error_type": "database_error"})
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to count payment vouchers",
				"error":   err.Error(),
			})
		}
		ids, err = config.Queries.ListPaymentVoucherIDsLimited(ctx, db.ListPaymentVoucherIDsLimitedParams{
			OrganizationID: tenant.OrganizationID,
			Column2:        status,
			Column3:        vendorID,
			CreatedBy:      &scope.UserID,
			Lower:          scope.UserRole,
			Column6:        orgRoleIDs,
			Limit:          int32(limit),
			Offset:         offset,
		})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to fetch payment vouchers",
				"error":   err.Error(),
			})
		}
		}
	}

	var vouchers []models.PaymentVoucher
	if len(ids) > 0 {
		q := config.DB.Where("id IN ?", ids).Preload("Vendor").Order("created_at DESC")
		// Filter by proof_of_payment presence when hasProofOfPayment query param is provided
		if hasPoP == "false" {
			q = q.Where("proof_of_payment IS NULL OR proof_of_payment::text = 'null' OR proof_of_payment::text = '{}'")
		} else if hasPoP == "true" {
			q = q.Where("proof_of_payment IS NOT NULL AND proof_of_payment::text != 'null' AND proof_of_payment::text != '{}'")
		}
		if err := q.Find(&vouchers).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to fetch payment vouchers",
				"error":   err.Error(),
			})
		}
	}

	responses := make([]types.PaymentVoucherResponse, 0, len(vouchers))
	for _, voucher := range vouchers {
		responses = append(responses, modelToPaymentVoucherResponse(voucher))
	}

	// Resolve creator / paid-by into {id,name,email,role} objects.
	if len(responses) > 0 {
		ids := make([]string, 0, len(responses)*2)
		for _, r := range responses {
			ids = append(ids, r.CreatedBy)
			if r.PaidBy != nil {
				ids = append(ids, *r.PaidBy)
			}
		}
		users := utils.ResolveUserRefs(config.DB, tenant.OrganizationID, ids)
		for i := range responses {
			if u, ok := users[responses[i].CreatedBy]; ok {
				creator := u
				responses[i].Creator = &creator
			}
			if responses[i].PaidBy != nil {
				if u, ok := users[*responses[i].PaidBy]; ok {
					paidBy := u
					responses[i].PaidByUser = &paidBy
				}
			}
		}
	}

	return utils.SendPaginatedSuccess(c, responses, "Payment vouchers retrieved successfully", page, limit, total)
}

// CreatePaymentVoucher creates a new payment voucher
func CreatePaymentVoucher(c *fiber.Ctx) error {
	// Get organization context from tenant middleware
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Organization context required",
			"error":   err.Error(),
		})
	}

	var req types.CreatePaymentVoucherRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	if req.InvoiceNumber == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invoice number is required",
		})
	}
	if req.Amount <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Amount must be greater than 0",
		})
	}
	if req.Description == "" || len(req.Description) < 10 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Description is required and must be at least 10 characters",
		})
	}

	// Verify vendor exists if provided
	var vendorIDPtr *string
	if req.VendorID != "" {
		var vendor models.Vendor
		if err := config.DB.Where("id = ? AND organization_id = ?", req.VendorID, tenant.OrganizationID).First(&vendor).Error; err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Vendor not found",
			})
		}
		vendorIDPtr = &req.VendorID
	}

	// Enforce the PV-creation gate (PO APPROVED, remaining-balance cap,
	// goods-first GRN) and the PV insert inside one transaction: the gate
	// locks the PO row FOR UPDATE, so this must run in the same transaction
	// as the insert for the lock to actually serialize concurrent requests
	// against the same PO. Shared gate with the from-PO and auto-create PV
	// paths.
	var voucher models.PaymentVoucher
	var pvCreateUser models.User
	txErr := config.DB.Transaction(func(tx *gorm.DB) error {
		if msg, code := validateProcurementPVGate(tx, tenant.OrganizationID, req.LinkedPO, req.LinkedGRN, req.Amount); code != 0 {
			return &pvGateError{status: code, msg: msg}
		}

		// When PO-linked, load its total + currency once: currency is inherited
		// (a PV can't drift to a different/empty currency than its PO) and the
		// total + already-committed amount snapshot the payment schedule for the
		// PV metadata built below.
		var poTotal, committedBefore float64
		if req.LinkedPO != "" {
			var lpo models.PurchaseOrder
			if err := tx.Select("currency", "total_amount").
				Where("document_number = ? AND organization_id = ?", req.LinkedPO, tenant.OrganizationID).
				First(&lpo).Error; err == nil {
				poTotal = lpo.TotalAmount
				if strings.TrimSpace(req.Currency) == "" && lpo.Currency != "" {
					req.Currency = lpo.Currency
				}
			}
			if sum, err := services.ComputePOPaymentSummary(tx, tenant.OrganizationID, req.LinkedPO); err == nil {
				committedBefore = sum.Committed
			}
		}

		// Generate voucher number
		documentNumber := utils.GenerateDocumentNumber("PV")

		tx.Where("id = ?", tenant.UserID).First(&pvCreateUser)

		voucher = models.PaymentVoucher{
			ID:             uuid.New().String(),
			OrganizationID: tenant.OrganizationID,
			DocumentNumber: documentNumber,
			VendorID:       vendorIDPtr,
			VendorName:     req.VendorName,
			InvoiceNumber:  req.InvoiceNumber,
			Status:         models.StatusDraft,
			Amount:         req.Amount,
			Currency:       req.Currency,
			PaymentMethod:  req.PaymentMethod,
			GLCode:         req.GLCode,
			Description:    req.Description,
			ApprovalStage:  0,
			LinkedPO:       req.LinkedPO,
			LinkedGRN:      req.LinkedGRN,
			CreatedBy:      tenant.UserID,
			Title:          req.Title,
			Department:     req.Department,
			DepartmentID:   req.DepartmentID,
			Priority:       req.Priority,
			BudgetCode:     req.BudgetCode,
			CostCenter:     req.CostCenter,
			ProjectCode:    req.ProjectCode,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		voucher.ApprovalHistory = datatypes.NewJSONType([]types.ApprovalRecord{})
		pvCreateNow := time.Now()
		voucher.ActionHistory = datatypes.NewJSONType([]types.ActionHistoryEntry{{
			ID:              uuid.New().String(),
			Action:          "CREATE",
			ActionType:      "CREATE",
			PerformedBy:     tenant.UserID,
			PerformedByName: pvCreateUser.Name,
			PerformedByRole: pvCreateUser.Role,
			Timestamp:       pvCreateNow,
			PerformedAt:     pvCreateNow,
			Comments:        "Payment voucher created",
			NewStatus:       models.StatusDraft,
		}})

		voucher.Metadata = buildPVCreationMetadata(req.PaymentType, req.Narration, req.Amount, poTotal, committedBefore)

		if err := tx.Create(&voucher).Error; err != nil {
			return err
		}
		return nil
	})
	if txErr != nil {
		if ge, ok := txErr.(*pvGateError); ok {
			return c.Status(ge.status).JSON(fiber.Map{"success": false, "message": ge.msg})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to create payment voucher",
			"error":   txErr.Error(),
		})
	}

	config.DB.Preload("Vendor").First(&voucher)

	go utils.SyncDocumentAs(config.DB, "PAYMENT_VOUCHER", voucher.ID, tenant.UserID)
	go services.LogDocumentEvent(config.DB, services.DocumentEvent{
		OrganizationID: tenant.OrganizationID,
		DocumentID:     voucher.ID,
		DocumentType:   "payment_voucher",
		UserID:         tenant.UserID,
		ActorName:      pvCreateUser.Name,
		ActorRole:      tenant.UserRole,
		Action:         "created",
		Details:        map[string]interface{}{"documentNumber": voucher.DocumentNumber},
	})

	// Optional auto-submit: when the org has enabled AutoSubmitPVToWorkflow,
	// hand the PV straight to the default workflow on creation.
	if pvOrgSettings, _ := services.NewOrganizationService(config.DB).
		GetOrganizationSettings(tenant.OrganizationID); pvOrgSettings != nil && pvOrgSettings.AutoSubmitPVToWorkflow {
		if wfSvc, ok := c.Locals("workflowExecutionService").(*services.WorkflowExecutionService); ok && wfSvc != nil {
			if defaultWF, _ := services.NewWorkflowService(nil, nil, config.DB).
				GetDefaultWorkflowForEntity(tenant.OrganizationID, "payment_voucher"); defaultWF != nil {
				autoTx := config.DB.Begin()
				autoSubmitOK := false
				if autoTx.Error != nil {
					// Couldn't open the transaction — skip auto-submit; the PV
					// stays DRAFT and the caller can submit it manually.
				} else if _, err := wfSvc.AssignWorkflowToDocumentWithIDTx(
					c.Context(), autoTx, tenant.OrganizationID, voucher.ID, "payment_voucher",
					defaultWF.ID.String(), tenant.UserID,
				); err != nil {
					autoTx.Rollback()
				} else if updErr := autoTx.Model(&models.PaymentVoucher{}).
					Where("id = ?", voucher.ID).
					Update("status", models.StatusPending).Error; updErr != nil {
					// Status flip failed: roll back so we don't leave a workflow
					// assigned to a PV that's still DRAFT, and don't report success.
					autoTx.Rollback()
				} else if commitErr := autoTx.Commit().Error; commitErr == nil {
					autoSubmitOK = true
					_ = config.DB.Where("id = ?", voucher.ID).First(&voucher).Error
				}

				// Post-commit: append a system-actor audit entry so the
				// trail reflects that no human signed off this submission.
				if autoSubmitOK {
					pvAutoNow := time.Now()
					pvAutoHistory := voucher.ActionHistory.Data()
					pvAutoHistory = append(pvAutoHistory, types.ActionHistoryEntry{
						ID:              uuid.New().String(),
						Action:          "AUTO_SUBMIT",
						ActionType:      "SUBMIT",
						PerformedBy:     "system",
						PerformedByName: "System (auto-submit)",
						PerformedByRole: "system",
						Timestamp:       pvAutoNow,
						PerformedAt:     pvAutoNow,
						PreviousStatus:  models.StatusDraft,
						NewStatus:       models.StatusPending,
						Comments:        "Auto-submitted via AutoSubmitPVToWorkflow org setting",
						Metadata: map[string]interface{}{
							"triggeredBy":    tenant.UserID,
							"orgSettingFlag": "AutoSubmitPVToWorkflow",
						},
					})
					voucher.ActionHistory = datatypes.NewJSONType(pvAutoHistory)
					_ = config.DB.Model(&voucher).
						Where("id = ?", voucher.ID).
						Update("action_history", voucher.ActionHistory).Error
				}
			}
		}
	}

	return c.Status(fiber.StatusCreated).JSON(types.DetailResponse{
		Success: true,
		Data:    modelToPaymentVoucherResponse(voucher),
	})
}

// GetPaymentVoucher retrieves a single payment voucher by ID
func GetPaymentVoucher(c *fiber.Ctx) error {
	// Set cache control headers to ensure fresh data for PDF generation
	c.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Set("Pragma", "no-cache")
	c.Set("Expires", "0")

	// Get organization context from tenant middleware
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Organization context required",
			"error":   err.Error(),
		})
	}

	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Payment Voucher ID is required",
		})
	}

	// Org scope + role/ownership scope. Detail endpoint now mirrors the list
	// endpoint's access policy so a user without visibility in the list can't
	// reach the doc by guessing/sharing the UUID. ApplyToQuery is a no-op for
	// privileged and procurement users.
	scope := utils.GetDocumentScope(config.DB, tenant.UserID, tenant.UserRole, tenant.OrganizationID)
	query := config.DB.
		Preload("Vendor").
		Where("id = ? AND organization_id = ?", id, tenant.OrganizationID)
	query = scope.ApplyToQuery(query, "created_by", "payment_voucher", "")

	var voucher models.PaymentVoucher
	if err := query.First(&voucher).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Payment voucher not found",
		})
	}

	response := modelToPaymentVoucherResponse(voucher)
	if liveHistory := utils.GetDocumentApprovalHistory(config.DB, voucher.ID, "payment_voucher"); len(liveHistory) > 0 {
		response.ApprovalHistory = liveHistory
	}
	return c.JSON(types.DetailResponse{
		Success: true,
		Data:    response,
	})
}

// UpdatePaymentVoucher updates an existing payment voucher
func UpdatePaymentVoucher(c *fiber.Ctx) error {
	// Get organization context from tenant middleware
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Organization context required",
			"error":   err.Error(),
		})
	}

	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Payment Voucher ID is required",
		})
	}

	var req types.UpdatePaymentVoucherRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	// Scope to org + owner/involvement so a user can only edit their own PV.
	scope := utils.GetDocumentScope(config.DB, tenant.UserID, tenant.UserRole, tenant.OrganizationID)
	loadQuery := config.DB.Where("id = ? AND organization_id = ?", id, tenant.OrganizationID)
	loadQuery = scope.ApplyToQuery(loadQuery, "created_by", "payment_voucher", "")
	var voucher models.PaymentVoucher
	if err := loadQuery.First(&voucher).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Payment voucher not found",
		})
	}

	// Metadata-only updates (e.g. supporting-document attachments) are allowed
	// on any status — mirrors the PO carve-out in purchase_order.go
	// UpdatePurchaseOrder. isMetadataOnly requires every other field on the
	// request to be absent/zero; a single non-metadata field falls through to
	// the status guard below.
	//
	// SECURITY: this expression must enumerate EVERY non-metadata field of
	// types.UpdatePaymentVoucherRequest. Adding a field to that struct without
	// adding it here silently lets requests carrying it bypass the status
	// guard on approved PVs. Keep in lockstep with the struct definition.
	isMetadataOnly := len(req.Metadata) > 0 &&
		req.VendorID == "" &&
		req.VendorName == "" &&
		req.InvoiceNumber == "" &&
		req.Amount == 0 &&
		req.Currency == "" &&
		req.PaymentMethod == "" &&
		req.GLCode == "" &&
		req.Description == "" &&
		req.Title == "" &&
		req.Department == "" &&
		req.DepartmentID == "" &&
		req.Priority == "" &&
		req.BudgetCode == "" &&
		req.CostCenter == "" &&
		req.ProjectCode == "" &&
		req.Items == nil

	if strings.ToUpper(voucher.Status) != "DRAFT" && strings.ToUpper(voucher.Status) != "PENDING" && !isMetadataOnly {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": fmt.Sprintf("Cannot update payment voucher in %s status", voucher.Status),
		})
	}

	if req.VendorID != "" {
		voucher.VendorID = &req.VendorID
	}
	if req.VendorID != "" || req.VendorName != "" {
		voucher.VendorName = req.VendorName
	}
	if req.InvoiceNumber != "" {
		voucher.InvoiceNumber = req.InvoiceNumber
	}
	if req.Amount > 0 {
		voucher.Amount = req.Amount
	}
	if req.Currency != "" {
		voucher.Currency = req.Currency
	}
	if req.PaymentMethod != "" {
		voucher.PaymentMethod = req.PaymentMethod
	}
	if req.GLCode != "" {
		voucher.GLCode = req.GLCode
	}
	if req.Description != "" {
		voucher.Description = req.Description
	}
	if req.Title != "" {
		voucher.Title = req.Title
	}
	if req.Department != "" {
		voucher.Department = req.Department
	}
	if req.DepartmentID != "" {
		voucher.DepartmentID = req.DepartmentID
	}
	if req.Priority != "" {
		voucher.Priority = req.Priority
	}
	if req.BudgetCode != "" {
		voucher.BudgetCode = req.BudgetCode
	}
	if req.CostCenter != "" {
		voucher.CostCenter = req.CostCenter
	}
	if req.ProjectCode != "" {
		voucher.ProjectCode = req.ProjectCode
	}
	if len(req.Metadata) > 0 {
		// Deep-merge incoming metadata with existing — never wipe keys other
		// parts of the system manage independently (e.g. B4's paymentType /
		// narration written at PV creation). Mirrors the PO carve-out in
		// purchase_order.go UpdatePurchaseOrder.
		existingMeta := map[string]interface{}{}
		if len(voucher.Metadata) > 0 {
			_ = json.Unmarshal(voucher.Metadata, &existingMeta)
		}
		for k, v := range req.Metadata {
			existingMeta[k] = v
		}
		if metaBytes, err := json.Marshal(existingMeta); err == nil {
			voucher.Metadata = datatypes.JSON(metaBytes)
		}
	}

	// Line-item edits are only allowed while the voucher is still a DRAFT.
	// Once submitted, the items become part of the approval record and must
	// not change underneath approvers.
	if req.Items != nil {
		if strings.ToUpper(voucher.Status) != "DRAFT" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"message": "Line items can only be edited while the payment voucher is in DRAFT status",
			})
		}
		items := *req.Items
		voucher.Items = datatypes.NewJSONType(items)

		// Keep the headline Amount in sync with the line items so that
		// totals on the PV detail, list, and PDF stay consistent.
		var itemsTotal float64
		for _, it := range items {
			itemsTotal += it.Amount
		}
		voucher.Amount = itemsTotal
	}

	var pvUpdateUser models.User
	config.DB.Where("id = ?", tenant.UserID).First(&pvUpdateUser)
	pvUpdateNow := time.Now()
	pvUpdateHistory := voucher.ActionHistory.Data()
	pvUpdateHistory = append(pvUpdateHistory, types.ActionHistoryEntry{
		ID:              uuid.New().String(),
		Action:          "UPDATE",
		ActionType:      "UPDATE",
		PerformedBy:     tenant.UserID,
		PerformedByName: pvUpdateUser.Name,
		PerformedByRole: pvUpdateUser.Role,
		Timestamp:       pvUpdateNow,
		PerformedAt:     pvUpdateNow,
		Comments:        "Payment voucher updated",
		NewStatus:       voucher.Status,
	})
	voucher.ActionHistory = datatypes.NewJSONType(pvUpdateHistory)
	voucher.UpdatedAt = pvUpdateNow

	if err := config.DB.Save(&voucher).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to update payment voucher",
			"error":   err.Error(),
		})
	}

	config.DB.Preload("Vendor").First(&voucher)

	go utils.SyncDocumentAs(config.DB, "PAYMENT_VOUCHER", voucher.ID, tenant.UserID)
	go services.LogDocumentEvent(config.DB, services.DocumentEvent{
		OrganizationID: tenant.OrganizationID,
		DocumentID:     voucher.ID,
		DocumentType:   "payment_voucher",
		UserID:         tenant.UserID,
		ActorName:      pvUpdateUser.Name,
		ActorRole:      tenant.UserRole,
		Action:         "updated",
		Details:        map[string]interface{}{"documentNumber": voucher.DocumentNumber},
	})

	return c.JSON(types.DetailResponse{
		Success: true,
		Data:    modelToPaymentVoucherResponse(voucher),
	})
}

// DeletePaymentVoucher deletes a payment voucher
func DeletePaymentVoucher(c *fiber.Ctx) error {
	// Get organization context from tenant middleware
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Organization context required",
			"error":   err.Error(),
		})
	}

	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Payment Voucher ID is required",
		})
	}

	// Scope to org + owner/involvement so a user can only delete their own PV.
	scope := utils.GetDocumentScope(config.DB, tenant.UserID, tenant.UserRole, tenant.OrganizationID)
	loadQuery := config.DB.Where("id = ? AND organization_id = ?", id, tenant.OrganizationID)
	loadQuery = scope.ApplyToQuery(loadQuery, "created_by", "payment_voucher", "")
	var voucher models.PaymentVoucher
	if err := loadQuery.First(&voucher).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Payment voucher not found",
		})
	}

	if strings.ToUpper(voucher.Status) != "DRAFT" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Only draft payment vouchers can be deleted",
		})
	}

	if err := config.DB.Delete(&voucher).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to delete payment voucher",
			"error":   err.Error(),
		})
	}

	return c.JSON(types.MessageResponse{
		Success: true,
		Message: "Payment voucher deleted successfully",
	})
}

// buildPVCreationMetadata records how a PV maps onto its linked PO's payment
// schedule: whether it is a full or partial payment, the free-text narration
// explaining the amount, and a snapshot of the PO total + amount already
// committed across earlier PVs at creation time. Stored in
// PaymentVoucher.Metadata (JSONB) for the partial-payment UI and audit trail.
// poTotal<=0 means the PV isn't PO-linked, so the amount can't be classified
// against a PO total and defaults to "full". Shared by the manual and from-PO
// creation paths so both persist an identical metadata shape.
func buildPVCreationMetadata(explicitType, narration string, amount, poTotal, committedBefore float64) datatypes.JSON {
	paymentType := strings.ToLower(strings.TrimSpace(explicitType))
	if paymentType != "full" && paymentType != "partial" {
		paymentType = "full"
		if poTotal > 0 && amount < poTotal-0.01 {
			paymentType = "partial"
		}
	}
	meta := map[string]interface{}{"paymentType": paymentType}
	if n := strings.TrimSpace(narration); n != "" {
		meta["narration"] = n
	}
	if poTotal > 0 {
		meta["poTotalAtCreation"] = poTotal
		meta["committedBefore"] = committedBefore
	}
	b, _ := json.Marshal(meta)
	return datatypes.JSON(b)
}

// modelToPaymentVoucherResponse converts a PaymentVoucher model to its API response.
func modelToPaymentVoucherResponse(voucher models.PaymentVoucher) types.PaymentVoucherResponse {
	var approvalHistory []types.ApprovalRecord
	if len(voucher.ApprovalHistory.Data()) > 0 {
		approvalHistory = voucher.ApprovalHistory.Data()
	}

	vendorID := ""
	if voucher.VendorID != nil {
		vendorID = *voucher.VendorID
	}
	vendorName := voucher.VendorName // stored fallback
	var vendorResp *types.VendorResponse
	if voucher.Vendor != nil {
		vendorName = voucher.Vendor.Name // canonical wins when relation present
		vr := modelToVendorResponse(*voucher.Vendor)
		vendorResp = &vr
	}

	actionHistory := voucher.ActionHistory.Data()

	// Unmarshal bank details
	var bankDetails interface{}
	if len(voucher.BankDetails) > 0 {
		_ = json.Unmarshal(voucher.BankDetails, &bankDetails)
	}

	// Unmarshal proof of payment
	var proofOfPayment interface{}
	if len(voucher.ProofOfPayment) > 0 {
		_ = json.Unmarshal(voucher.ProofOfPayment, &proofOfPayment)
	}

	// Unmarshal generic metadata (paymentType, narration, ...)
	var metadata map[string]interface{}
	if len(voucher.Metadata) > 0 {
		_ = json.Unmarshal(voucher.Metadata, &metadata)
	}

	items := voucher.Items.Data()

	return types.PaymentVoucherResponse{
		ID:                   voucher.ID,
		OrganizationID:       voucher.OrganizationID,
		DocumentNumber:       voucher.DocumentNumber,
		VendorID:             vendorID,
		VendorName:           vendorName,
		Vendor:               vendorResp,
		InvoiceNumber:        voucher.InvoiceNumber,
		Status:               voucher.Status,
		Amount:               voucher.Amount,
		Currency:             voucher.Currency,
		PaymentMethod:        voucher.PaymentMethod,
		GLCode:               voucher.GLCode,
		Description:          voucher.Description,
		ApprovalStage:        voucher.ApprovalStage,
		ApprovalHistory:      approvalHistory,
		ActionHistory:        actionHistory,
		LinkedPO:             voucher.LinkedPO,
		LinkedGRN:            voucher.LinkedGRN,
		Title:                voucher.Title,
		Department:           voucher.Department,
		DepartmentID:         voucher.DepartmentID,
		Priority:             voucher.Priority,
		BudgetCode:           voucher.BudgetCode,
		CostCenter:           voucher.CostCenter,
		ProjectCode:          voucher.ProjectCode,
		CreatedBy:            voucher.CreatedBy,
		RequestedByName:      voucher.RequestedByName,
		RequestedDate:        voucher.RequestedDate,
		SubmittedAt:          voucher.SubmittedAt,
		ApprovedAt:           voucher.ApprovedAt,
		PaidDate:             voucher.PaidDate,
		PaymentDueDate:       voucher.PaymentDueDate,
		TaxAmount:            voucher.TaxAmount,
		WithholdingTaxAmount: voucher.WithholdingTaxAmount,
		PaidAmount:           voucher.PaidAmount,
		BankDetails:          bankDetails,
		Items:                items,
		RoutingType:          voucher.RoutingType,
		ProofOfPayment:       proofOfPayment,
		PaidAt:               voucher.PaidAt,
		PaidBy:               voucher.PaidBy,
		Metadata:             metadata,
		CreatedAt:            voucher.CreatedAt,
		UpdatedAt:            voucher.UpdatedAt,
	}
}

// SubmitPaymentVoucher submits a payment voucher for approval using the workflow system
func SubmitPaymentVoucher(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Payment Voucher ID is required",
		})
	}

	// Get organization ID and user ID from context
	organizationID := c.Locals("organizationID").(string)
	userID := c.Locals("userID").(string)

	var submitReq types.SubmitDocumentRequest
	if err := c.BodyParser(&submitReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
		})
	}
	if submitReq.WorkflowID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "workflowId is required",
		})
	}

	// Get existing payment voucher
	var voucher models.PaymentVoucher
	if err := config.DB.Where("id = ? AND organization_id = ?", id, organizationID).First(&voucher).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Payment Voucher not found",
		})
	}

	// Scope gate: only the PV owner, a privileged role, or a user with an
	// assigned/claimed workflow task on this PV may submit it.
	userRole := strings.ToLower(c.Locals("userRole").(string))
	scope := utils.GetDocumentScope(config.DB, userID, userRole, organizationID)
	if !scope.CanViewAll && !scope.IsProcurement {
		isOwner := strings.EqualFold(voucher.CreatedBy, userID)
		if !isOwner {
			var taskCount int64
			if err := config.DB.Table("workflow_tasks").
				Where("entity_id = ? AND entity_type = ? AND organization_id = ?", id, "payment_voucher", organizationID).
				Where("assigned_user_id = ? OR claimed_by = ?", userID, userID).
				Count(&taskCount).Error; err != nil {
				return utils.SendInternalError(c, "failed to verify task assignment", err)
			}
			if taskCount == 0 {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"success": false,
					"message": "You do not have permission to submit this payment voucher",
				})
			}
		}
	}

	// Check if payment voucher is in draft status
	if strings.ToUpper(voucher.Status) != "DRAFT" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": fmt.Sprintf("Cannot submit payment voucher in %s status", voucher.Status),
		})
	}

	// Gate: if linked to a PO, it must still be APPROVED or FULFILLED (goods
	// fully delivered, balance still outstanding) before PV can be submitted.
	if voucher.LinkedPO != "" {
		var linkedPO models.PurchaseOrder
		if err := config.DB.
			Where("document_number = ? AND organization_id = ?", voucher.LinkedPO, organizationID).
			First(&linkedPO).Error; err != nil {
			return utils.SendBadRequestError(c, "Linked purchase order not found")
		}
		linkedPOStatus := strings.ToUpper(linkedPO.Status)
		if linkedPOStatus != "APPROVED" && linkedPOStatus != models.StatusFulfilled {
			return utils.SendBadRequestError(c, fmt.Sprintf(
				"Cannot submit PV: linked PO %s is in %s status and must be APPROVED or FULFILLED.",
				voucher.LinkedPO, linkedPO.Status))
		}
	}

	// Gate: goods-first flow — linked GRN must still be APPROVED before PV can be submitted
	if voucher.LinkedGRN != "" {
		var linkedGRN models.GoodsReceivedNote
		if err := config.DB.
			Where("document_number = ? AND organization_id = ?", voucher.LinkedGRN, organizationID).
			First(&linkedGRN).Error; err != nil {
			return utils.SendBadRequestError(c, "Linked goods received note not found")
		}
		grnStatusSub := strings.ToUpper(linkedGRN.Status)
		if grnStatusSub != "APPROVED" && grnStatusSub != "COMPLETED" {
			return utils.SendBadRequestError(c, fmt.Sprintf(
				"Cannot submit PV: linked GRN %s is in %s status and must be APPROVED or COMPLETED.",
				voucher.LinkedGRN, linkedGRN.Status))
		}
	}

	// Get workflow execution service from context
	workflowExecutionService := c.Locals("workflowExecutionService").(*services.WorkflowExecutionService)

	// Atomic submit: status change + workflow assignment in a single transaction.
	tx := config.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	assignment, err := workflowExecutionService.AssignWorkflowToDocumentWithIDTx(
		c.Context(), tx, organizationID, voucher.ID, "payment_voucher", submitReq.WorkflowID, userID,
	)
	if err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to assign workflow to payment voucher",
			"error":   err.Error(),
		})
	}

	voucher.Status = models.StatusPending
	voucher.UpdatedAt = time.Now()

	// Use the open tx to avoid deadlock under single-conn DB pools.
	var user models.User
	_ = tx.Where("id = ?", userID).First(&user).Error
	actionHistory := voucher.ActionHistory.Data()
	actionHistory = append(actionHistory, types.ActionHistoryEntry{
		ID:              uuid.New().String(),
		Action:          "SUBMIT",
		PerformedBy:     userID,
		PerformedByName: user.Name,
		PerformedByRole: user.Role,
		Timestamp:       time.Now(),
		Comments:        "Payment voucher submitted for approval",
		ActionType:      "SUBMIT",
		PreviousStatus:  models.StatusDraft,
		NewStatus:       models.StatusPending,
	})
	voucher.ActionHistory = datatypes.NewJSONType(actionHistory)

	if err := tx.Save(&voucher).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to update payment voucher status",
			"error":   err.Error(),
		})
	}

	if err := tx.Commit().Error; err != nil {
		return utils.SendInternalError(c, "Failed to submit payment voucher", err)
	}

	go utils.SyncDocumentAs(config.DB, "PAYMENT_VOUCHER", voucher.ID, userID)
	go services.LogDocumentEvent(config.DB, services.DocumentEvent{
		OrganizationID: organizationID,
		DocumentID:     voucher.ID,
		DocumentType:   "payment_voucher",
		UserID:         userID,
		ActorName:      user.Name,
		ActorRole:      user.Role,
		Action:         "submitted",
		Details:        map[string]interface{}{"documentNumber": voucher.DocumentNumber},
	})

	return c.JSON(types.DetailResponse{
		Success: true,
		Data: fiber.Map{
			"paymentVoucher": modelToPaymentVoucherResponse(voucher),
			"workflow": fiber.Map{
				"assignmentId": assignment.ID,
				"workflowId":   assignment.WorkflowID,
				"currentStage": assignment.CurrentStage,
				"status":       assignment.Status,
			},
		},
	})
}

// WithdrawPaymentVoucher withdraws a payment voucher from approval workflow
func WithdrawPaymentVoucher(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Payment Voucher ID is required",
		})
	}

	// Get organization ID and user ID from context
	organizationID := c.Locals("organizationID").(string)
	userID := c.Locals("userID").(string)

	// Get existing payment voucher
	var voucher models.PaymentVoucher
	if err := config.DB.Where("id = ? AND organization_id = ?", id, organizationID).First(&voucher).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Payment Voucher not found",
		})
	}

	// Only the creator may withdraw. Prefer the persisted CreatedBy; fall back to
	// the CREATE action-history entry for legacy vouchers that predate the field.
	// (Auto-created PVs set CreatedBy but leave the history PerformedBy blank, so
	// the old history-only logic made them un-withdrawable.)
	creatorID := voucher.CreatedBy
	if creatorID == "" {
		for _, action := range voucher.ActionHistory.Data() {
			if strings.ToUpper(action.ActionType) == "CREATE" {
				creatorID = action.PerformedBy
				break
			}
		}
	}

	if creatorID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Only the creator can withdraw this payment voucher",
		})
	}

	// Check if payment voucher is in a state that can be withdrawn (pending)
	if strings.ToUpper(voucher.Status) != "PENDING" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": fmt.Sprintf("Cannot withdraw payment voucher in %s status. Only pending payment vouchers can be withdrawn.", voucher.Status),
		})
	}

	// Check if there is an active workflow task that is claimed
	var workflowTask models.WorkflowTask
	err := config.DB.Where("entity_id = ? AND entity_type = ? AND UPPER(status) IN (?, ?)",
		id, "payment_voucher", "PENDING", "CLAIMED").First(&workflowTask).Error

	if err == nil {
		// Task exists - check if it's claimed
		if strings.ToUpper(workflowTask.Status) == "CLAIMED" && workflowTask.ClaimedBy != nil {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"success": false,
				"message": "Cannot withdraw payment voucher. It is currently being reviewed by an approver.",
			})
		}
	}

	// Start a transaction to ensure all changes are atomic
	tx := config.DB.Begin()
	if tx.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to start transaction",
			"error":   tx.Error.Error(),
		})
	}

	// Delete the workflow task(s) for this payment voucher
	if err := tx.Where("entity_id = ? AND entity_type = ?", id, "payment_voucher").
		Delete(&models.WorkflowTask{}).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to remove workflow tasks",
			"error":   err.Error(),
		})
	}

	// Delete the workflow assignment(s) for this payment voucher
	if err := tx.Where("entity_id = ? AND entity_type = ?", id, "payment_voucher").
		Delete(&models.WorkflowAssignment{}).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to remove workflow assignments",
			"error":   err.Error(),
		})
	}

	// Update payment voucher status back to draft and reset approval fields
	previousStatus := voucher.Status
	voucher.Status = models.StatusDraft
	voucher.ApprovalStage = 0
	voucher.UpdatedAt = time.Now()

	// Clear approval history since we're reverting to draft
	voucher.ApprovalHistory = datatypes.NewJSONType([]types.ApprovalRecord{})

	// Add action history entry for withdrawal
	actionHistory := voucher.ActionHistory.Data()

	// Get user info for action history
	performerName := "Unknown User"
	performerRole := "unknown"
	var user models.User
	if err := tx.Where("id = ?", userID).First(&user).Error; err == nil {
		performerName = user.Name
		performerRole = user.Role
	}

	actionHistory = append(actionHistory, types.ActionHistoryEntry{
		ID:              uuid.New().String(),
		Action:          "WITHDRAW",
		PerformedBy:     userID,
		PerformedByName: performerName,
		PerformedByRole: performerRole,
		Timestamp:       time.Now(),
		Comments:        "Payment voucher withdrawn by creator",
		ActionType:      "WITHDRAW",
		PreviousStatus:  previousStatus,
		NewStatus:       models.StatusDraft,
	})
	voucher.ActionHistory = datatypes.NewJSONType(actionHistory)

	// Save payment voucher changes
	if err := tx.Save(&voucher).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to update payment voucher status",
			"error":   err.Error(),
		})
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to commit changes",
			"error":   err.Error(),
		})
	}

	// Preload vendor for response
	config.DB.Preload("Vendor").First(&voucher)

	go services.LogDocumentEvent(config.DB, services.DocumentEvent{
		OrganizationID: organizationID,
		DocumentID:     voucher.ID,
		DocumentType:   "payment_voucher",
		UserID:         userID,
		ActorName:      user.Name,
		ActorRole:      user.Role,
		Action:         "withdrawn",
		Details:        map[string]interface{}{"documentNumber": voucher.DocumentNumber},
	})

	return c.JSON(fiber.Map{
		"success": true,
		"data":    modelToPaymentVoucherResponse(voucher),
		"message": "Payment voucher withdrawn successfully. You can now edit and re-submit it.",
	})
}
