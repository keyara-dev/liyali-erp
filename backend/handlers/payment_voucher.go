package handlers

import (
	"fmt"
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

	db := config.DB

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

	// Determine document visibility scope for this user
	scope := utils.GetDocumentScope(db, tenant.UserID, tenant.UserRole, tenant.OrganizationID)

	// Start with organization filter - CRITICAL SECURITY FIX
	query := db.Where("organization_id = ?", tenant.OrganizationID)

	// Apply document scope
	if scope.IsProcurement {
		// Procurement users only see PVs generated from a PO (procurement chain)
		query = query.Where("linked_po != ''")
	} else {
		query = scope.ApplyToQuery(query, "created_by", "payment_voucher", "")
	}

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if vendorID != "" {
		query = query.Where("vendor_id = ?", vendorID)
	}

	var total int64
	if err := query.Model(&models.PaymentVoucher{}).Count(&total).Error; err != nil {
		logging.LogError(c, err, "failed_to_count_payment_vouchers", map[string]interface{}{
			"error_type": "database_error",
		})
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to count payment vouchers",
			"error":   err.Error(),
		})
	}

	var vouchers []models.PaymentVoucher
	offset := (page - 1) * limit
	if err := query.
		Offset(offset).
		Limit(limit).
		Preload("Vendor").
		Order("created_at DESC").
		Find(&vouchers).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to fetch payment vouchers",
			"error":   err.Error(),
		})
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

	if req.VendorID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Vendor ID is required",
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

	// Verify vendor exists and belongs to organization - SECURITY FIX
	var vendor models.Vendor
	if err := config.DB.Where("id = ? AND organization_id = ?", req.VendorID, tenant.OrganizationID).First(&vendor).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Vendor not found",
		})
	}

	// Generate voucher number
	documentNumber := utils.GenerateDocumentNumber("PV")

	voucher := models.PaymentVoucher{
		ID:             uuid.New().String(),
		OrganizationID: tenant.OrganizationID, // SECURITY FIX: Set organization ID
		DocumentNumber: documentNumber,
		VendorID:       req.VendorID,
		InvoiceNumber:  req.InvoiceNumber,
		Status:         "draft",
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

	if err := config.DB.Create(&voucher).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to create payment voucher",
			"error":   err.Error(),
		})
	}

	config.DB.Preload("Vendor").First(&voucher)

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

	return c.JSON(types.DetailResponse{
		Success: true,
		Data:    modelToPaymentVoucherResponse(voucher),
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

	if voucher.Status != "draft" && voucher.Status != "pending" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": fmt.Sprintf("Cannot update payment voucher in %s status", voucher.Status),
		})
	}

	if req.VendorID != "" {
		voucher.VendorID = req.VendorID
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

	voucher.UpdatedAt = time.Now()

	if err := config.DB.Save(&voucher).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to update payment voucher",
			"error":   err.Error(),
		})
	}

	config.DB.Preload("Vendor").First(&voucher)

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

	if voucher.Status != "draft" {
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

	vendorName := ""
	if voucher.Vendor != nil {
		vendorName = voucher.Vendor.Name
	}

	return types.PaymentVoucherResponse{
		ID:              voucher.ID,
		DocumentNumber:  voucher.DocumentNumber,
		VendorID:        voucher.VendorID,
		VendorName:      vendorName,
		InvoiceNumber:   voucher.InvoiceNumber,
		Status:          voucher.Status,
		Amount:          voucher.Amount,
		Currency:        voucher.Currency,
		PaymentMethod:   voucher.PaymentMethod,
		GLCode:          voucher.GLCode,
		Description:     voucher.Description,
		ApprovalStage:   voucher.ApprovalStage,
		ApprovalHistory: approvalHistory,
		LinkedPO:        voucher.LinkedPO,
		CreatedAt:       voucher.CreatedAt,
		UpdatedAt:       voucher.UpdatedAt,
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
	if voucher.Status != "draft" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": fmt.Sprintf("Cannot submit payment voucher in %s status", voucher.Status),
		})
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
	voucher.Status = "pending"
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
			PreviousStatus:  "draft",
			NewStatus:       "pending",
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
