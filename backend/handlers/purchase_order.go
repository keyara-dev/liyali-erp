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

// GetPurchaseOrders retrieves all purchase orders with pagination and filtering
func GetPurchaseOrders(c *fiber.Ctx) error {
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

	query := db
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if vendorID != "" {
		query = query.Where("vendor_id = ?", vendorID)
	}

	var total int64
	if err := query.Model(&models.PurchaseOrder{}).Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to count purchase orders",
			"error":   err.Error(),
		})
	}

	var orders []models.PurchaseOrder
	offset := (page - 1) * limit
	if err := query.
		Offset(offset).
		Limit(limit).
		Preload("Vendor").
		Order("created_at DESC").
		Find(&orders).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to fetch purchase orders",
			"error":   err.Error(),
		})
	}

	responses := make([]types.PurchaseOrderResponse, 0, len(orders))
	for _, order := range orders {
		responses = append(responses, modelToPurchaseOrderResponse(order))
	}

	return utils.SendPaginatedSuccess(c, responses, "Purchase orders retrieved successfully", page, limit, total)
}

// CreatePurchaseOrder creates a new purchase order
func CreatePurchaseOrder(c *fiber.Ctx) error {
	var req types.CreatePurchaseOrderRequest

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
	if len(req.Items) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "At least one item is required",
		})
	}
	if req.TotalAmount <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Total amount must be greater than 0",
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

	// Generate PO number
	poNumber := fmt.Sprintf("PO-%d-%s", time.Now().Unix(), uuid.New().String()[:8])

	order := models.PurchaseOrder{
		ID:              uuid.New().String(),
		PONumber:        poNumber,
		VendorID:        req.VendorID,
		Status:          "draft",
		TotalAmount:     req.TotalAmount,
		Currency:        req.Currency,
		DeliveryDate:    req.DeliveryDate,
		ApprovalStage:   0,
		LinkedRequisition: req.LinkedRequisition,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	order.Items = datatypes.NewJSONType(req.Items)

	order.ApprovalHistory = datatypes.NewJSONType([]types.ApprovalRecord{})

	if err := config.DB.Create(&order).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to create purchase order",
			"error":   err.Error(),
		})
	}

	config.DB.Preload("Vendor").First(&order)

	return c.Status(fiber.StatusCreated).JSON(types.DetailResponse{
		Success: true,
		Data:    modelToPurchaseOrderResponse(order),
	})
}

// GetPurchaseOrder retrieves a single purchase order by ID
func GetPurchaseOrder(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Purchase Order ID is required",
		})
	}

	var order models.PurchaseOrder
	if err := config.DB.
		Preload("Vendor").
		Where("id = ?", id).
		First(&order).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Purchase order not found",
		})
	}

	return c.JSON(types.DetailResponse{
		Success: true,
		Data:    modelToPurchaseOrderResponse(order),
	})
}

// UpdatePurchaseOrder updates an existing purchase order
func UpdatePurchaseOrder(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Purchase Order ID is required",
		})
	}

	var req types.UpdatePurchaseOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	var order models.PurchaseOrder
	if err := config.DB.Where("id = ?", id).First(&order).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Purchase order not found",
		})
	}

	if order.Status != "draft" && order.Status != "pending" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": fmt.Sprintf("Cannot update purchase order in %s status", order.Status),
		})
	}

	if req.VendorID != "" {
		order.VendorID = req.VendorID
	}
	if len(req.Items) > 0 {
		order.Items = datatypes.NewJSONType(req.Items)
	}
	if req.TotalAmount > 0 {
		order.TotalAmount = req.TotalAmount
	}
	if req.Currency != "" {
		order.Currency = req.Currency
	}
	if !req.DeliveryDate.IsZero() {
		order.DeliveryDate = req.DeliveryDate
	}

	order.UpdatedAt = time.Now()

	if err := config.DB.Save(&order).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to update purchase order",
			"error":   err.Error(),
		})
	}

	config.DB.Preload("Vendor").First(&order)

	return c.JSON(types.DetailResponse{
		Success: true,
		Data:    modelToPurchaseOrderResponse(order),
	})
}

// DeletePurchaseOrder deletes a purchase order
func DeletePurchaseOrder(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Purchase Order ID is required",
		})
	}

	var order models.PurchaseOrder
	if err := config.DB.Where("id = ?", id).First(&order).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Purchase order not found",
		})
	}

	if order.Status != "draft" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Only draft purchase orders can be deleted",
		})
	}

	if err := config.DB.Delete(&order).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to delete purchase order",
			"error":   err.Error(),
		})
	}

	return c.JSON(types.MessageResponse{
		Success: true,
		Message: "Purchase order deleted successfully",
	})
}

// ApprovePurchaseOrder approves a purchase order and optionally auto-creates GRN
func ApprovePurchaseOrder(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Purchase Order ID is required",
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

	var order models.PurchaseOrder
	if err := config.DB.Where("id = ?", id).First(&order).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Purchase order not found",
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
	approvalHistory = order.ApprovalHistory.Data()

	approvalRecord := types.ApprovalRecord{
		ApproverID:   approverID,
		ApproverName: approver.Name,
		Status:       "approved",
		Comments:     req.Comments,
		Signature:    req.Signature,
		ApprovedAt:   time.Now(),
	}
	approvalHistory = append(approvalHistory, approvalRecord)

	order.Status = "approved"
	order.ApprovalStage++
	order.ApprovalHistory = datatypes.NewJSONType([]types.ApprovalRecord{})
	order.UpdatedAt = time.Now()

	if err := config.DB.Save(&order).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to approve purchase order",
			"error":   err.Error(),
		})
	}

	config.DB.Preload("Vendor").First(&order)

	// Auto-create GRN if enabled and prerequisites are met
	var autoCreatedGRN *models.GoodsReceivedNote
	if order.Status == "approved" {
		// Initialize automation service
		auditService := &services.AuditService{}
		notificationService := &services.NotificationService{}
		automationService := services.NewDocumentAutomationService(
			config.DB, auditService, notificationService,
		)

		// Get automation config
		automationConfig := automationService.GetDefaultAutomationConfig()

		// Attempt to auto-create GRN
		result, err := automationService.CreateGRNFromPurchaseOrder(
			c.Context(), &order, automationConfig,
		)

		if err == nil && result.Success {
			if grn, ok := result.CreatedDocument.(models.GoodsReceivedNote); ok {
				autoCreatedGRN = &grn
			}
		}
		// Note: We don't fail the approval if GRN creation fails
		// The PO is still approved, GRN can be created manually
	}

	// Prepare response
	response := types.DetailResponse{
		Success: true,
		Data:    modelToPurchaseOrderResponse(order),
	}

	// Add auto-created GRN to response if available
	if autoCreatedGRN != nil {
		// Convert GRN to response format
		grnResponse := types.GRNResponse{
			ID:           autoCreatedGRN.ID,
			GRNNumber:    autoCreatedGRN.GRNNumber,
			PONumber:     autoCreatedGRN.PONumber,
			Status:       autoCreatedGRN.Status,
			ReceivedDate: autoCreatedGRN.ReceivedDate,
			ReceivedBy:   autoCreatedGRN.ReceivedBy,
			CreatedAt:    autoCreatedGRN.CreatedAt,
			UpdatedAt:    autoCreatedGRN.UpdatedAt,
		}

		// Add GRN to response
		response.Data = fiber.Map{
			"purchaseOrder":   modelToPurchaseOrderResponse(order),
			"autoCreatedGRN":  grnResponse,
			"automationUsed":  true,
		}
	}

	return c.JSON(response)
}

// RejectPurchaseOrder rejects a purchase order
func RejectPurchaseOrder(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Purchase Order ID is required",
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

	var order models.PurchaseOrder
	if err := config.DB.Where("id = ?", id).First(&order).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Purchase order not found",
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
	approvalHistory = order.ApprovalHistory.Data()

	rejectionRecord := types.ApprovalRecord{
		ApproverID:   approverID,
		ApproverName: approver.Name,
		Status:       "rejected",
		Comments:     req.Remarks,
		Signature:    req.Signature,
		ApprovedAt:   time.Now(),
	}
	approvalHistory = append(approvalHistory, rejectionRecord)

	order.Status = "rejected"
	order.ApprovalHistory = datatypes.NewJSONType([]types.ApprovalRecord{})
	order.UpdatedAt = time.Now()

	if err := config.DB.Save(&order).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to reject purchase order",
			"error":   err.Error(),
		})
	}

	config.DB.Preload("Vendor").First(&order)

	return c.JSON(types.DetailResponse{
		Success: true,
		Data:    modelToPurchaseOrderResponse(order),
	})
}

// Helper function to convert model to response
func modelToPurchaseOrderResponse(order models.PurchaseOrder) types.PurchaseOrderResponse {
	var items []types.POItem
	if len(order.Items.Data()) > 0 {
		items = order.Items.Data()
	}

	var approvalHistory []types.ApprovalRecord

	vendorName := ""
	if order.Vendor != nil {
		vendorName = order.Vendor.Name
	}

	return types.PurchaseOrderResponse{
		ID:                order.ID,
		PONumber:          order.PONumber,
		VendorID:          order.VendorID,
		VendorName:        vendorName,
		Status:            order.Status,
		Items:             items,
		TotalAmount:       order.TotalAmount,
		Currency:          order.Currency,
		DeliveryDate:      order.DeliveryDate,
		ApprovalStage:     order.ApprovalStage,
		ApprovalHistory:   approvalHistory,
		LinkedRequisition: order.LinkedRequisition,
		CreatedAt:         order.CreatedAt,
		UpdatedAt:         order.UpdatedAt,
	}
}
