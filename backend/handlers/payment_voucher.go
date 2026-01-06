package handlers

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/logging"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/types"
	"github.com/liyali/liyali-gateway/utils"
	"gorm.io/datatypes"
)

// GetPaymentVouchers retrieves all payment vouchers with pagination and filtering
func GetPaymentVouchers(c *fiber.Ctx) error {
	logger := logging.FromContext(c)
	logger.Info("get_payment_vouchers_request")

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
		"operation":  "get_payment_vouchers",
		"page":       page,
		"limit":      limit,
		"status":     status,
		"vendor_id":  vendorID,
	})

	query := db
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

	// Verify vendor exists
	var vendor models.Vendor
	if err := config.DB.Where("id = ?", req.VendorID).First(&vendor).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Vendor not found",
		})
	}

	// Generate voucher number
	voucherNumber := fmt.Sprintf("PV-%d-%s", time.Now().Unix(), uuid.New().String()[:8])

	voucher := models.PaymentVoucher{
		ID:            uuid.New().String(),
		VoucherNumber: voucherNumber,
		VendorID:      req.VendorID,
		InvoiceNumber: req.InvoiceNumber,
		Status:        "draft",
		Amount:        req.Amount,
		Currency:      req.Currency,
		PaymentMethod: req.PaymentMethod,
		GLCode:        req.GLCode,
		Description:   req.Description,
		ApprovalStage: 0,
		LinkedPO:      req.LinkedPO,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
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
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Payment Voucher ID is required",
		})
	}

	var voucher models.PaymentVoucher
	if err := config.DB.
		Preload("Vendor").
		Where("id = ?", id).
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
	if err := config.DB.Where("id = ?", id).First(&voucher).Error; err != nil {
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
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Payment Voucher ID is required",
		})
	}

	var voucher models.PaymentVoucher
	if err := config.DB.Where("id = ?", id).First(&voucher).Error; err != nil {
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

// ApprovePaymentVoucher approves a payment voucher
func ApprovePaymentVoucher(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Payment Voucher ID is required",
		})
	}

	var req types.ApproveDocumentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	if req.Signature == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Signature is required",
		})
	}

	var voucher models.PaymentVoucher
	if err := config.DB.Where("id = ?", id).First(&voucher).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Payment voucher not found",
		})
	}

	approverID := c.Locals("userID").(string)
	var approver models.User
	if err := config.DB.Where("id = ?", approverID).First(&approver).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Approver not found",
		})
	}

	var approvalHistory []types.ApprovalRecord
	approvalHistory = voucher.ApprovalHistory.Data()

	approvalRecord := types.ApprovalRecord{
		ApproverID:   approverID,
		ApproverName: approver.Name,
		Status:       "approved",
		Comments:     req.Comments,
		Signature:    req.Signature,
		ApprovedAt:   time.Now(),
	}
	approvalHistory = append(approvalHistory, approvalRecord)

	voucher.Status = "approved"
	voucher.ApprovalStage++
	voucher.ApprovalHistory = datatypes.NewJSONType([]types.ApprovalRecord{})
	voucher.UpdatedAt = time.Now()

	if err := config.DB.Save(&voucher).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to approve payment voucher",
			"error":   err.Error(),
		})
	}

	config.DB.Preload("Vendor").First(&voucher)

	return c.JSON(types.DetailResponse{
		Success: true,
		Data:    modelToPaymentVoucherResponse(voucher),
	})
}

// RejectPaymentVoucher rejects a payment voucher
func RejectPaymentVoucher(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Payment Voucher ID is required",
		})
	}

	var req types.RejectDocumentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	if req.Remarks == "" || len(req.Remarks) < 10 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Remarks must be at least 10 characters",
		})
	}
	if req.Signature == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Signature is required",
		})
	}

	var voucher models.PaymentVoucher
	if err := config.DB.Where("id = ?", id).First(&voucher).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Payment voucher not found",
		})
	}

	approverID := c.Locals("userID").(string)
	var approver models.User
	if err := config.DB.Where("id = ?", approverID).First(&approver).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Approver not found",
		})
	}

	var approvalHistory []types.ApprovalRecord
	approvalHistory = voucher.ApprovalHistory.Data()

	rejectionRecord := types.ApprovalRecord{
		ApproverID:   approverID,
		ApproverName: approver.Name,
		Status:       "rejected",
		Comments:     req.Remarks,
		Signature:    req.Signature,
		ApprovedAt:   time.Now(),
	}
	approvalHistory = append(approvalHistory, rejectionRecord)

	voucher.Status = "rejected"
	voucher.ApprovalHistory = datatypes.NewJSONType([]types.ApprovalRecord{})
	voucher.UpdatedAt = time.Now()

	if err := config.DB.Save(&voucher).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to reject payment voucher",
			"error":   err.Error(),
		})
	}

	config.DB.Preload("Vendor").First(&voucher)

	return c.JSON(types.DetailResponse{
		Success: true,
		Data:    modelToPaymentVoucherResponse(voucher),
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
		VoucherNumber:   voucher.VoucherNumber,
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
