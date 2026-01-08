package handlers

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/services"
	"github.com/liyali/liyali-gateway/types"
	"github.com/liyali/liyali-gateway/utils"
	"gorm.io/datatypes"
)

// GetGRNs retrieves all goods received notes with pagination and filtering
func GetGRNs(c *fiber.Ctx) error {
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
	poNumber := c.Query("poNumber")

	query := db
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if poNumber != "" {
		query = query.Where("po_number = ?", poNumber)
	}

	var total int64
	if err := query.Model(&models.GoodsReceivedNote{}).Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to count GRNs",
			"error":   err.Error(),
		})
	}

	var grns []models.GoodsReceivedNote
	offset := (page - 1) * limit
	if err := query.
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&grns).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to fetch GRNs",
			"error":   err.Error(),
		})
	}

	responses := make([]types.GRNResponse, 0, len(grns))
	for _, grn := range grns {
		responses = append(responses, modelToGRNResponse(grn))
	}

	return utils.SendPaginatedSuccess(c, responses, "GRNs retrieved successfully", page, limit, total)
}

// CreateGRN creates a new goods received note
func CreateGRN(c *fiber.Ctx) error {
	var req types.CreateGRNRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	if req.PONumber == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "PO number is required",
		})
	}
	if len(req.Items) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "At least one item is required",
		})
	}
	if req.ReceivedBy == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "ReceivedBy is required",
		})
	}

	// Verify PO exists
	var po models.PurchaseOrder
	if err := config.DB.Where("po_number = ?", req.PONumber).First(&po).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Purchase order not found",
		})
	}

	// Generate GRN number
	grnNumber := utils.GenerateGRNNumber()

	grn := models.GoodsReceivedNote{
		ID:          uuid.New().String(),
		GRNNumber:   grnNumber,
		PONumber:    req.PONumber,
		Status:      "draft",
		ReceivedDate: time.Now(),
		ReceivedBy:  req.ReceivedBy,
		ApprovalStage: 0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	grn.Items = datatypes.NewJSONType(req.Items)

	emptyQuality := []types.QualityIssue{}
	grn.QualityIssues = datatypes.NewJSONType(emptyQuality)

	emptyHistory := []types.ApprovalRecord{}
	grn.ApprovalHistory = datatypes.NewJSONType(emptyHistory)

	if err := config.DB.Create(&grn).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to create GRN",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(types.DetailResponse{
		Success: true,
		Data:    modelToGRNResponse(grn),
	})
}

// GetGRN retrieves a single GRN by ID
func GetGRN(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "GRN ID is required",
		})
	}

	var grn models.GoodsReceivedNote
	if err := config.DB.Where("id = ?", id).First(&grn).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "GRN not found",
		})
	}

	return c.JSON(types.DetailResponse{
		Success: true,
		Data:    modelToGRNResponse(grn),
	})
}

// UpdateGRN updates an existing GRN
func UpdateGRN(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "GRN ID is required",
		})
	}

	var req types.UpdateGRNRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	var grn models.GoodsReceivedNote
	if err := config.DB.Where("id = ?", id).First(&grn).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "GRN not found",
		})
	}

	if grn.Status != "draft" && grn.Status != "pending" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": fmt.Sprintf("Cannot update GRN in %s status", grn.Status),
		})
	}

	if len(req.Items) > 0 {
		grn.Items = datatypes.NewJSONType(req.Items)
	}
	if req.ReceivedBy != "" {
		grn.ReceivedBy = req.ReceivedBy
	}
	if len(req.QualityIssues) > 0 {
		grn.QualityIssues = datatypes.NewJSONType(req.QualityIssues)
	}

	grn.UpdatedAt = time.Now()

	if err := config.DB.Save(&grn).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to update GRN",
			"error":   err.Error(),
		})
	}

	return c.JSON(types.DetailResponse{
		Success: true,
		Data:    modelToGRNResponse(grn),
	})
}

// DeleteGRN deletes a GRN
func DeleteGRN(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "GRN ID is required",
		})
	}

	var grn models.GoodsReceivedNote
	if err := config.DB.Where("id = ?", id).First(&grn).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "GRN not found",
		})
	}

	if grn.Status != "draft" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Only draft GRNs can be deleted",
		})
	}

	if err := config.DB.Delete(&grn).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to delete GRN",
			"error":   err.Error(),
		})
	}

	return c.JSON(types.MessageResponse{
		Success: true,
		Message: "GRN deleted successfully",
	})
}

// ApproveGRN approves a GRN and optionally auto-creates Payment Voucher
func ApproveGRN(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "GRN ID is required",
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

	var grn models.GoodsReceivedNote
	if err := config.DB.Where("id = ?", id).First(&grn).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "GRN not found",
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
	approvalHistory = grn.ApprovalHistory.Data()

	approvalRecord := types.ApprovalRecord{
		ApproverID:   approverID,
		ApproverName: approver.Name,
		Status:       "approved",
		Comments:     req.Comments,
		Signature:    req.Signature,
		ApprovedAt:   time.Now(),
	}
	approvalHistory = append(approvalHistory, approvalRecord)

	grn.Status = "approved"
	grn.ApprovalStage++
	grn.ApprovalHistory = datatypes.NewJSONType(approvalHistory)
	grn.UpdatedAt = time.Now()

	if err := config.DB.Save(&grn).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to approve GRN",
			"error":   err.Error(),
		})
	}

	// Auto-create Payment Voucher if enabled and prerequisites are met
	var autoCreatedPV *models.PaymentVoucher
	if grn.Status == "approved" {
		// Initialize automation service
		auditService := &services.AuditService{}
		notificationService := &services.NotificationService{}
		automationService := services.NewDocumentAutomationService(
			config.DB, auditService, notificationService,
		)

		// Get automation config
		automationConfig := automationService.GetDefaultAutomationConfig()

		// Attempt to auto-create Payment Voucher
		result, err := automationService.CreatePaymentVoucherFromGRN(
			c.Context(), &grn, automationConfig,
		)

		if err == nil && result.Success {
			if pv, ok := result.CreatedDocument.(models.PaymentVoucher); ok {
				autoCreatedPV = &pv
			}
		}
		// Note: We don't fail the approval if PV creation fails
		// The GRN is still approved, PV can be created manually
	}

	// Prepare response
	response := types.DetailResponse{
		Success: true,
		Data:    modelToGRNResponse(grn),
	}

	// Add auto-created PV to response if available
	if autoCreatedPV != nil {
		// Convert PV to response format
		pvResponse := types.PaymentVoucherResponse{
			ID:            autoCreatedPV.ID,
			VoucherNumber: autoCreatedPV.VoucherNumber,
			VendorID:      autoCreatedPV.VendorID,
			InvoiceNumber: autoCreatedPV.InvoiceNumber,
			Status:        autoCreatedPV.Status,
			Amount:        autoCreatedPV.Amount,
			Currency:      autoCreatedPV.Currency,
			PaymentMethod: autoCreatedPV.PaymentMethod,
			LinkedPO:      autoCreatedPV.LinkedPO,
			ApprovalStage: autoCreatedPV.ApprovalStage,
			CreatedAt:     autoCreatedPV.CreatedAt,
			UpdatedAt:     autoCreatedPV.UpdatedAt,
		}

		// Add PV to response
		response.Data = fiber.Map{
			"grn":             modelToGRNResponse(grn),
			"autoCreatedPV":   pvResponse,
			"automationUsed":  true,
		}
	}

	return c.JSON(response)
}

// RejectGRN rejects a GRN
func RejectGRN(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "GRN ID is required",
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

	var grn models.GoodsReceivedNote
	if err := config.DB.Where("id = ?", id).First(&grn).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "GRN not found",
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
	approvalHistory = grn.ApprovalHistory.Data()

	rejectionRecord := types.ApprovalRecord{
		ApproverID:   approverID,
		ApproverName: approver.Name,
		Status:       "rejected",
		Comments:     req.Remarks,
		Signature:    req.Signature,
		ApprovedAt:   time.Now(),
	}
	approvalHistory = append(approvalHistory, rejectionRecord)

	grn.Status = "rejected"
	grn.ApprovalHistory = datatypes.NewJSONType(approvalHistory)
	grn.UpdatedAt = time.Now()

	if err := config.DB.Save(&grn).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to reject GRN",
			"error":   err.Error(),
		})
	}

	return c.JSON(types.DetailResponse{
		Success: true,
		Data:    modelToGRNResponse(grn),
	})
}

// Helper function to convert model to response
func modelToGRNResponse(grn models.GoodsReceivedNote) types.GRNResponse {
	var items []types.GRNItem
	items = grn.Items.Data()

	var qualityIssues []types.QualityIssue
	qualityIssues = grn.QualityIssues.Data()

	var approvalHistory []types.ApprovalRecord
	approvalHistory = grn.ApprovalHistory.Data()

	return types.GRNResponse{
		ID:              grn.ID,
		GRNNumber:       grn.GRNNumber,
		PONumber:        grn.PONumber,
		Status:          grn.Status,
		ReceivedDate:    grn.ReceivedDate,
		ReceivedBy:      grn.ReceivedBy,
		Items:           items,
		QualityIssues:   qualityIssues,
		ApprovalStage:   grn.ApprovalStage,
		ApprovalHistory: approvalHistory,
		CreatedAt:       grn.CreatedAt,
		UpdatedAt:       grn.UpdatedAt,
	}
}
