package handlers

import (
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
)

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

	// Add query parameters to context
	logging.AddFieldsToRequest(c, map[string]interface{}{
		"operation":       "get_payment_vouchers",
		"page":            page,
		"limit":           limit,
		"status":          status,
		"vendor_id":       vendorID,
		"organization_id": tenant.OrganizationID,
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

	switch {
	case scope.CanViewAll:
		total, err = config.Queries.CountPaymentVouchersAll(ctx, tenant.OrganizationID, status, vendorID)
		if err != nil {
			logging.LogError(c, err, "failed_to_count_payment_vouchers", map[string]interface{}{"error_type": "database_error"})
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to count payment vouchers",
				"error":   err.Error(),
			})
		}
		ids, err = config.Queries.ListPaymentVoucherIDsAll(ctx, tenant.OrganizationID, status, vendorID, int32(limit), offset)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to fetch payment vouchers",
				"error":   err.Error(),
			})
		}
	case scope.IsProcurement:
		total, err = config.Queries.CountPaymentVouchersProcurement(ctx, tenant.OrganizationID, status, vendorID)
		if err != nil {
			logging.LogError(c, err, "failed_to_count_payment_vouchers", map[string]interface{}{"error_type": "database_error"})
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to count payment vouchers",
				"error":   err.Error(),
			})
		}
		ids, err = config.Queries.ListPaymentVoucherIDsProcurement(ctx, tenant.OrganizationID, status, vendorID, int32(limit), offset)
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

	var vouchers []models.PaymentVoucher
	if len(ids) > 0 {
		if err := config.DB.
			Where("id IN ?", ids).
			Preload("Vendor").
			Order("created_at DESC").
			Find(&vouchers).Error; err != nil {
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

	// One-to-one guard + PO APPROVED gate
	if req.LinkedPO != "" {
		var linkedPO models.PurchaseOrder
		if err := config.DB.
			Where("document_number = ? AND organization_id = ?", req.LinkedPO, tenant.OrganizationID).
			First(&linkedPO).Error; err != nil {
			return utils.SendBadRequestError(c, "Linked purchase order not found")
		}
		if strings.ToUpper(linkedPO.Status) != "APPROVED" {
			return utils.SendBadRequestError(c, fmt.Sprintf(
				"Cannot create PV: linked PO %s is in %s status and must be APPROVED first.",
				req.LinkedPO, linkedPO.Status))
		}
		var existingPV models.PaymentVoucher
		if err := config.DB.
			Where("linked_po = ? AND organization_id = ? AND UPPER(status) != 'CANCELLED'",
				req.LinkedPO, tenant.OrganizationID).
			First(&existingPV).Error; err == nil {
			return utils.SendConflictError(c, fmt.Sprintf(
				"Payment voucher %s already exists for PO %s (status: %s).",
				existingPV.DocumentNumber, req.LinkedPO, existingPV.Status))
		}
	}

	// Generate voucher number
	documentNumber := utils.GenerateDocumentNumber("PV")

	var pvCreateUser models.User
	config.DB.Where("id = ?", tenant.UserID).First(&pvCreateUser)

	voucher := models.PaymentVoucher{
		ID:             uuid.New().String(),
		OrganizationID: tenant.OrganizationID, // SECURITY FIX: Set organization ID
		DocumentNumber: documentNumber,
		VendorID:       vendorIDPtr,
		InvoiceNumber:  req.InvoiceNumber,
		Status: "DRAFT",
		Amount:         req.Amount,
		Currency:       req.Currency,
		PaymentMethod:  req.PaymentMethod,
		GLCode:         req.GLCode,
		Description:    req.Description,
		ApprovalStage:  0,
		LinkedPO:       req.LinkedPO,
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
		NewStatus:       "DRAFT",
	}})

	if err := config.DB.Create(&voucher).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to create payment voucher",
			"error":   err.Error(),
		})
	}

	config.DB.Preload("Vendor").First(&voucher)

	go utils.SyncDocument(config.DB, "PAYMENT_VOUCHER", voucher.ID)
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

	var voucher models.PaymentVoucher
	// SECURITY FIX: Filter by organization ID
	if err := config.DB.
		Preload("Vendor").
		Where("id = ? AND organization_id = ?", id, tenant.OrganizationID).
		First(&voucher).Error; err != nil {
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

	var voucher models.PaymentVoucher
	// SECURITY FIX: Filter by organization ID
	if err := config.DB.Where("id = ? AND organization_id = ?", id, tenant.OrganizationID).First(&voucher).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Payment voucher not found",
		})
	}

	if strings.ToUpper(voucher.Status) != "DRAFT" && strings.ToUpper(voucher.Status) != "PENDING" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": fmt.Sprintf("Cannot update payment voucher in %s status", voucher.Status),
		})
	}

	if req.VendorID != "" {
		voucher.VendorID = &req.VendorID
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

	go utils.SyncDocument(config.DB, "PAYMENT_VOUCHER", voucher.ID)
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

	var voucher models.PaymentVoucher
	// SECURITY FIX: Filter by organization ID
	if err := config.DB.Where("id = ? AND organization_id = ?", id, tenant.OrganizationID).First(&voucher).Error; err != nil {
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

// Helper function to convert model to response
func modelToPaymentVoucherResponse(voucher models.PaymentVoucher) types.PaymentVoucherResponse {
	var approvalHistory []types.ApprovalRecord
	if len(voucher.ApprovalHistory.Data()) > 0 {
		approvalHistory = voucher.ApprovalHistory.Data()
	}

	vendorID := ""
	if voucher.VendorID != nil {
		vendorID = *voucher.VendorID
	}
	vendorName := ""
	if voucher.Vendor != nil {
		vendorName = voucher.Vendor.Name
	}

	actionHistory := voucher.ActionHistory.Data()
	
	// Unmarshal bank details
	var bankDetails interface{}
	if len(voucher.BankDetails) > 0 {
		_ = json.Unmarshal(voucher.BankDetails, &bankDetails)
	}
	
	items := voucher.Items.Data()

	return types.PaymentVoucherResponse{
		ID:                   voucher.ID,
		OrganizationID:       voucher.OrganizationID,
		DocumentNumber:       voucher.DocumentNumber,
		VendorID:             vendorID,
		VendorName:           vendorName,
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

	// Check if payment voucher is in draft status
	if strings.ToUpper(voucher.Status) != "DRAFT" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": fmt.Sprintf("Cannot submit payment voucher in %s status", voucher.Status),
		})
	}

	// Gate: if linked to a PO, it must still be APPROVED before PV can be submitted
	if voucher.LinkedPO != "" {
		var linkedPO models.PurchaseOrder
		if err := config.DB.
			Where("document_number = ? AND organization_id = ?", voucher.LinkedPO, organizationID).
			First(&linkedPO).Error; err != nil {
			return utils.SendBadRequestError(c, "Linked purchase order not found")
		}
		if strings.ToUpper(linkedPO.Status) != "APPROVED" {
			return utils.SendBadRequestError(c, fmt.Sprintf(
				"Cannot submit PV: linked PO %s is in %s status and must be APPROVED.",
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
		if strings.ToUpper(linkedGRN.Status) != "APPROVED" {
			return utils.SendBadRequestError(c, fmt.Sprintf(
				"Cannot submit PV: linked GRN %s is in %s status and must be APPROVED.",
				voucher.LinkedGRN, linkedGRN.Status))
		}
	}

	// Get workflow execution service from context
	workflowExecutionService := c.Locals("workflowExecutionService").(*services.WorkflowExecutionService)

	// Assign workflow to the payment voucher
	assignment, err := workflowExecutionService.AssignWorkflowToDocumentWithID(
		c.Context(), organizationID, voucher.ID, "payment_voucher", submitReq.WorkflowID, userID,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to assign workflow to payment voucher",
			"error":   err.Error(),
		})
	}

	// Update payment voucher status to pending
	voucher.Status = "PENDING"
	voucher.UpdatedAt = time.Now()

	// Add action history entry for submission
	var actionHistory []types.ActionHistoryEntry
	actionHistory = voucher.ActionHistory.Data()

	// Get user info for action history
	var user models.User
	if err := config.DB.Where("id = ?", userID).First(&user).Error; err == nil {
		actionHistory = append(actionHistory, types.ActionHistoryEntry{
			ID:              uuid.New().String(),
			Action:          "SUBMIT",
			PerformedBy:     userID,
			PerformedByName: user.Name,
			PerformedByRole: user.Role,
			Timestamp:       time.Now(),
			Comments:        "Payment voucher submitted for approval",
			ActionType:      "SUBMIT",
			PreviousStatus:  "DRAFT",
			NewStatus:       "PENDING",
		})
		voucher.ActionHistory = datatypes.NewJSONType(actionHistory)
	}

	// Save payment voucher
	if err := config.DB.Save(&voucher).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to update payment voucher status",
			"error":   err.Error(),
		})
	}

	go utils.SyncDocument(config.DB, "PAYMENT_VOUCHER", voucher.ID)
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

	// Verify that the current user is the creator (only the submitter can withdraw)
	// Note: PaymentVoucher doesn't have a CreatedBy field, so we check the first action history entry
	var actionHistory []types.ActionHistoryEntry
	actionHistory = voucher.ActionHistory.Data()
	if actionHistory == nil || len(actionHistory) == 0 {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Cannot determine payment voucher creator",
		})
	}

	// Find the CREATE action to determine the creator
	creatorID := ""
	for _, action := range actionHistory {
		if strings.ToUpper(action.ActionType) == "CREATE" {
			creatorID = action.PerformedBy
			break
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
	voucher.Status = "DRAFT"
	voucher.ApprovalStage = 0
	voucher.UpdatedAt = time.Now()

	// Clear approval history since we're reverting to draft
	voucher.ApprovalHistory = datatypes.NewJSONType([]types.ApprovalRecord{})

	// Add action history entry for withdrawal
	if actionHistory == nil {
		actionHistory = []types.ActionHistoryEntry{}
	}

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
		NewStatus:       "DRAFT",
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
