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

// GetPurchaseOrders retrieves all purchase orders with pagination and filtering
func GetPurchaseOrders(c *fiber.Ctx) error {
	logger := logging.FromContext(c)
	logger.Info("get_purchase_orders_request")

	// Get organization context from tenant middleware
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Organization context required",
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
		"operation":      "get_purchase_orders",
		"page":           page,
		"limit":          limit,
		"status":         status,
		"vendor_id":      vendorID,
		"organizationID": tenant.OrganizationID,
	})

	// Determine document visibility scope for this user
	scope := utils.GetDocumentScope(db, tenant.UserID, tenant.UserRole, tenant.OrganizationID)

	// SECURITY: Always filter by organization ID first
	query := db.Where("organization_id = ?", tenant.OrganizationID)

	// Apply document scope (procurement users see all POs; limited users see own + involved)
	query = scope.ApplyToQuery(query, "created_by", "purchase_order", "")

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if vendorID != "" {
		query = query.Where("vendor_id = ?", vendorID)
	}

	var total int64
	if err := query.Model(&models.PurchaseOrder{}).Count(&total).Error; err != nil {
		logging.LogError(c, err, "failed_to_count_purchase_orders", map[string]interface{}{
			"error_type": "database_error",
		})
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
		logging.LogError(c, err, "failed_to_fetch_purchase_orders", map[string]interface{}{
			"error_type": "database_error",
			"offset":     offset,
			"limit":      limit,
		})
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

	logger.Info("purchase_orders_retrieved_successfully")

	return utils.SendPaginatedSuccess(c, responses, "Purchase orders retrieved successfully", page, limit, total)
}

// CreatePurchaseOrder creates a new purchase order
func CreatePurchaseOrder(c *fiber.Ctx) error {
	logger := logging.FromContext(c)
	logger.Info("create_purchase_order_request")

	// Get organization context from tenant middleware
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Organization context required",
			"error":   err.Error(),
		})
	}

	var req types.CreatePurchaseOrderRequest

	if err := c.BodyParser(&req); err != nil {
		logging.LogError(c, err, "invalid_request_body", map[string]interface{}{
			"error_type": "parsing_error",
		})
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	// Add operation context
	logging.AddFieldsToRequest(c, map[string]interface{}{
		"operation":       "create_purchase_order",
		"vendor_id":       req.VendorID,
		"total_amount":    req.TotalAmount,
		"currency":        req.Currency,
		"items_count":     len(req.Items),
		"organization_id": tenant.OrganizationID,
	})

	if req.VendorID == "" {
		logging.LogWarn(c, "vendor_id_missing")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Vendor ID is required",
		})
	}
	if len(req.Items) == 0 {
		logging.LogWarn(c, "no_items_provided")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "At least one item is required",
		})
	}
	// Validate items have positive quantities
	for _, item := range req.Items {
		if item.Quantity <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "All items must have positive quantities",
			})
		}
	}
	if req.TotalAmount <= 0 {
		logging.LogWarn(c, "invalid_total_amount")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Total amount must be greater than 0",
		})
	}
	// Validate delivery date is not in the past
	if !req.DeliveryDate.Time.IsZero() && req.DeliveryDate.Time.Before(time.Now().Truncate(24*time.Hour)) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Delivery date cannot be in the past",
		})
	}

	// Verify vendor exists and belongs to the same organization
	var vendor models.Vendor
	if err := config.DB.Where("id = ? AND organization_id = ?", req.VendorID, tenant.OrganizationID).First(&vendor).Error; err != nil {
		logging.LogError(c, err, "vendor_not_found", map[string]interface{}{
			"vendor_id":       req.VendorID,
			"organization_id": tenant.OrganizationID,
			"error_type":      "validation_error",
		})
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Vendor not found",
		})
	}

	// Generate PO number
	documentNumber := utils.GenerateDocumentNumber("PO")
	orderID := uuid.New().String()

	logging.AddFieldToRequest(c, "document_number", documentNumber)
	logging.AddFieldToRequest(c, "order_id", orderID)

	order := models.PurchaseOrder{
		ID:                orderID,
		OrganizationID:    tenant.OrganizationID, // SECURITY FIX: Set organization ID
		DocumentNumber:    documentNumber,
		VendorID:          req.VendorID,
		Status:            "draft",
		TotalAmount:       req.TotalAmount,
		Currency:          req.Currency,
		DeliveryDate:      req.DeliveryDate.Time,
		ApprovalStage:     0,
		LinkedRequisition: req.LinkedRequisition,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	order.Items = datatypes.NewJSONType(req.Items)
	order.ApprovalHistory = datatypes.NewJSONType([]types.ApprovalRecord{})

	if err := config.DB.Create(&order).Error; err != nil {
		logging.LogError(c, err, "failed_to_create_purchase_order", map[string]interface{}{
			"error_type":      "database_error",
			"document_number": documentNumber,
		})
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to create purchase order",
			"error":   err.Error(),
		})
	}

	config.DB.Preload("Vendor").First(&order)

	go utils.SyncDocument(config.DB, "PURCHASE_ORDER", order.ID)

	logger.Info("purchase_order_created_successfully")

	return c.Status(fiber.StatusCreated).JSON(types.DetailResponse{
		Success: true,
		Data:    modelToPurchaseOrderResponse(order),
	})
}

// GetPurchaseOrder retrieves a single purchase order by ID
func GetPurchaseOrder(c *fiber.Ctx) error {
	// Set cache control headers to ensure fresh data for PDF generation
	c.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Set("Pragma", "no-cache")
	c.Set("Expires", "0")

	logger := logging.FromContext(c)
	logger.Info("get_purchase_order_request")

	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Organization context required",
		})
	}

	id := c.Params("id")
	if id == "" {
		logging.LogWarn(c, "purchase_order_id_missing")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Purchase Order ID is required",
		})
	}

	logging.AddFieldsToRequest(c, map[string]interface{}{
		"operation": "get_purchase_order",
		"order_id":  id,
	})

	var order models.PurchaseOrder
	if err := config.DB.
		Preload("Vendor").
		Where("id = ? AND organization_id = ?", id, tenant.OrganizationID).
		First(&order).Error; err != nil {
		logging.LogError(c, err, "purchase_order_not_found", map[string]interface{}{
			"order_id":   id,
			"error_type": "not_found",
		})
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Purchase order not found",
		})
	}

	logger.Info("purchase_order_retrieved_successfully")

	return c.JSON(types.DetailResponse{
		Success: true,
		Data:    modelToPurchaseOrderResponse(order),
	})
}

// UpdatePurchaseOrder updates an existing purchase order
func UpdatePurchaseOrder(c *fiber.Ctx) error {
	logger := logging.FromContext(c)
	logger.Info("update_purchase_order_request")

	id := c.Params("id")
	if id == "" {
		logging.LogWarn(c, "purchase_order_id_missing")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Purchase Order ID is required",
		})
	}

	var req types.UpdatePurchaseOrderRequest
	if err := c.BodyParser(&req); err != nil {
		logging.LogError(c, err, "invalid_request_body", map[string]interface{}{
			"error_type": "parsing_error",
		})
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	logging.AddFieldsToRequest(c, map[string]interface{}{
		"operation": "update_purchase_order",
		"order_id":  id,
	})

	var order models.PurchaseOrder
	if err := config.DB.Where("id = ?", id).First(&order).Error; err != nil {
		logging.LogError(c, err, "purchase_order_not_found", map[string]interface{}{
			"order_id":   id,
			"error_type": "not_found",
		})
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Purchase order not found",
		})
	}

	if order.Status != "draft" && order.Status != "pending" {
		logging.LogWarn(c, "invalid_status_for_update", map[string]interface{}{
			"current_status":  order.Status,
			"document_number": order.DocumentNumber,
		})
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": fmt.Sprintf("Cannot update purchase order in %s status", order.Status),
		})
	}

	// Track changes for logging
	changes := make(map[string]interface{})

	if req.VendorID != "" {
		changes["vendor_id"] = map[string]string{"from": order.VendorID, "to": req.VendorID}
		order.VendorID = req.VendorID
	}
	if len(req.Items) > 0 {
		changes["items_count"] = len(req.Items)
		order.Items = datatypes.NewJSONType(req.Items)
	}
	if req.TotalAmount > 0 {
		changes["total_amount"] = map[string]float64{"from": order.TotalAmount, "to": req.TotalAmount}
		order.TotalAmount = req.TotalAmount
	}
	if req.Currency != "" {
		changes["currency"] = map[string]string{"from": order.Currency, "to": req.Currency}
		order.Currency = req.Currency
	}
	if !req.DeliveryDate.Time.IsZero() {
		changes["delivery_date"] = req.DeliveryDate.Time
		order.DeliveryDate = req.DeliveryDate.Time
	}

	order.UpdatedAt = time.Now()

	if err := config.DB.Save(&order).Error; err != nil {
		logging.LogError(c, err, "failed_to_update_purchase_order", map[string]interface{}{
			"error_type":      "database_error",
			"document_number": order.DocumentNumber,
			"changes":         changes,
		})
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to update purchase order",
			"error":   err.Error(),
		})
	}

	config.DB.Preload("Vendor").First(&order)

	go utils.SyncDocument(config.DB, "PURCHASE_ORDER", order.ID)

	logger.Info("purchase_order_updated_successfully")

	return c.JSON(types.DetailResponse{
		Success: true,
		Data:    modelToPurchaseOrderResponse(order),
	})
}

// DeletePurchaseOrder deletes a purchase order
func DeletePurchaseOrder(c *fiber.Ctx) error {
	logger := logging.FromContext(c)
	logger.Info("delete_purchase_order_request")

	id := c.Params("id")
	if id == "" {
		logging.LogWarn(c, "purchase_order_id_missing")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Purchase Order ID is required",
		})
	}

	logging.AddFieldsToRequest(c, map[string]interface{}{
		"operation": "delete_purchase_order",
		"order_id":  id,
	})

	var order models.PurchaseOrder
	if err := config.DB.Where("id = ?", id).First(&order).Error; err != nil {
		logging.LogError(c, err, "purchase_order_not_found", map[string]interface{}{
			"order_id":   id,
			"error_type": "not_found",
		})
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Purchase order not found",
		})
	}

	if order.Status != "draft" {
		logging.LogWarn(c, "invalid_status_for_deletion", map[string]interface{}{
			"current_status":  order.Status,
			"document_number": order.DocumentNumber,
		})
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Only draft purchase orders can be deleted",
		})
	}

	if err := config.DB.Delete(&order).Error; err != nil {
		logging.LogError(c, err, "failed_to_delete_purchase_order", map[string]interface{}{
			"error_type":      "database_error",
			"document_number": order.DocumentNumber,
		})
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to delete purchase order",
			"error":   err.Error(),
		})
	}

	logger.Info("purchase_order_deleted_successfully")

	return c.JSON(types.MessageResponse{
		Success: true,
		Message: "Purchase order deleted successfully",
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
		DocumentNumber:    order.DocumentNumber,
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

// SubmitPurchaseOrder submits a purchase order for approval using the workflow system
func SubmitPurchaseOrder(c *fiber.Ctx) error {
	logger := logging.FromContext(c)
	logger.Info("submit_purchase_order_request")

	id := c.Params("id")
	if id == "" {
		logging.LogWarn(c, "purchase_order_id_missing")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Purchase Order ID is required",
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

	logging.AddFieldsToRequest(c, map[string]interface{}{
		"operation": "submit_purchase_order",
		"order_id":  id,
	})

	// Get existing purchase order
	var order models.PurchaseOrder
	if err := config.DB.Where("id = ? AND organization_id = ?", id, organizationID).First(&order).Error; err != nil {
		logging.LogError(c, err, "purchase_order_not_found", map[string]interface{}{
			"order_id": id,
		})
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Purchase Order not found",
		})
	}

	// Check if purchase order is in draft status
	if order.Status != "draft" {
		logging.LogWarn(c, "invalid_purchase_order_status_for_submission", map[string]interface{}{
			"current_status": order.Status,
		})
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": fmt.Sprintf("Cannot submit purchase order in %s status", order.Status),
		})
	}

	// Get workflow execution service from context
	workflowExecutionService := c.Locals("workflowExecutionService").(*services.WorkflowExecutionService)

	// Assign workflow to the purchase order
	assignment, err := workflowExecutionService.AssignWorkflowToDocumentWithID(
		c.Context(), organizationID, order.ID, "purchase_order", submitReq.WorkflowID, userID,
	)
	if err != nil {
		logging.LogError(c, err, "workflow_assignment_failed", map[string]interface{}{
			"order_id": id,
		})
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to assign workflow to purchase order",
			"error":   err.Error(),
		})
	}

	// Update purchase order status to pending
	order.Status = "pending"
	order.UpdatedAt = time.Now()

	// Add action history entry for submission
	var actionHistory []types.ActionHistoryEntry
	actionHistory = order.ActionHistory.Data()

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
			Comments:        "Purchase order submitted for approval",
			ActionType:      "SUBMIT",
			PreviousStatus:  "draft",
			NewStatus:       "pending",
		})
		order.ActionHistory = datatypes.NewJSONType(actionHistory)
	}

	// Save purchase order
	if err := config.DB.Save(&order).Error; err != nil {
		logging.LogError(c, err, "purchase_order_update_failed", map[string]interface{}{
			"order_id": id,
		})
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to update purchase order status",
			"error":   err.Error(),
		})
	}

	// Preload vendor
	config.DB.Preload("Vendor").First(&order)

	go utils.SyncDocument(config.DB, "PURCHASE_ORDER", order.ID)

	logging.AddFieldsToRequest(c, map[string]interface{}{
		"order_id":      order.ID,
		"workflow_id":   assignment.WorkflowID,
		"assignment_id": assignment.ID,
	})
	logger.Info("purchase_order_submitted_successfully")

	return c.JSON(types.DetailResponse{
		Success: true,
		Data: fiber.Map{
			"purchaseOrder": modelToPurchaseOrderResponse(order),
			"workflow": fiber.Map{
				"assignmentId": assignment.ID,
				"workflowId":   assignment.WorkflowID,
				"currentStage": assignment.CurrentStage,
				"status":       assignment.Status,
			},
		},
	})
}
