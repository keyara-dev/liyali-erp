package handlers

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
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
		total, ids, err = utils.ListDocumentIDsFallback(config.DB, "purchase_orders", utils.DocumentListFilters{
			OrganizationID:    tenant.OrganizationID,
			Status:            status,
			RefField:          "vendor_id",
			RefValue:          vendorID,
			HideDirectPayment: scope.HideDirectPayment,
		}, scope, limit, int(offset))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to fetch purchase orders",
				"error":   err.Error(),
			})
		}
	} else if scope.CanViewAll || scope.IsProcurement {
		total, err = config.Queries.CountPurchaseOrdersAll(ctx, db.CountPurchaseOrdersAllParams{
			OrganizationID:    tenant.OrganizationID,
			Column2:           status,
			Column3:           vendorID,
			HideDirectPayment: scope.HideDirectPayment,
		})
		if err != nil {
			logging.LogError(c, err, "failed_to_count_purchase_orders", map[string]interface{}{"error_type": "database_error"})
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to count purchase orders",
				"error":   err.Error(),
			})
		}
		ids, err = config.Queries.ListPurchaseOrderIDsAll(ctx, db.ListPurchaseOrderIDsAllParams{
			OrganizationID:    tenant.OrganizationID,
			Column2:           status,
			Column3:           vendorID,
			HideDirectPayment: scope.HideDirectPayment,
			Limit:             int32(limit),
			Offset:            offset,
		})
		if err != nil {
			logging.LogError(c, err, "failed_to_fetch_purchase_orders", map[string]interface{}{"error_type": "database_error", "offset": offset, "limit": limit})
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to fetch purchase orders",
				"error":   err.Error(),
			})
		}
	} else {
		total, err = config.Queries.CountPurchaseOrdersLimited(ctx, db.CountPurchaseOrdersLimitedParams{
			OrganizationID: tenant.OrganizationID,
			Column2:        status,
			Column3:        vendorID,
			CreatedBy:      &scope.UserID,
			Lower:          scope.UserRole,
			Column6:        orgRoleIDs,
		})
		if err != nil {
			logging.LogError(c, err, "failed_to_count_purchase_orders", map[string]interface{}{"error_type": "database_error"})
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to count purchase orders",
				"error":   err.Error(),
			})
		}
		ids, err = config.Queries.ListPurchaseOrderIDsLimited(ctx, db.ListPurchaseOrderIDsLimitedParams{
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
			logging.LogError(c, err, "failed_to_fetch_purchase_orders", map[string]interface{}{"error_type": "database_error", "offset": offset, "limit": limit})
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to fetch purchase orders",
				"error":   err.Error(),
			})
		}
	}

	var orders []models.PurchaseOrder
	if len(ids) > 0 {
		if err := config.DB.
			Where("id IN ?", ids).
			Preload("Vendor").
			Order("created_at DESC").
			Find(&orders).Error; err != nil {
			logging.LogError(c, err, "failed_to_fetch_purchase_orders", map[string]interface{}{"error_type": "database_error"})
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to fetch purchase orders",
				"error":   err.Error(),
			})
		}
	}

	responses := make([]types.PurchaseOrderResponse, 0, len(orders))
	for _, order := range orders {
		responses = append(responses, modelToPurchaseOrderResponse(order))
	}

	// Batch-enrich responses with linked PV info (single query, not N+1).
	// Skip when sqlc Queries handle is unavailable (test harness) — enrichment
	// is best-effort decoration and is safe to omit.
	if len(responses) > 0 && config.Queries != nil {
		poDocNumbers := make([]string, len(responses))
		for i, r := range responses {
			poDocNumbers[i] = r.DocumentNumber
		}
		pvRows, _ := config.Queries.GetLinkedPVsForPurchaseOrders(ctx, poDocNumbers, tenant.OrganizationID)
		pvMap := make(map[string]db.GetLinkedPVsForPurchaseOrdersRow, len(pvRows))
		for _, r := range pvRows {
			if r.LinkedPo != nil {
				pvMap[*r.LinkedPo] = r
			}
		}
		for i, r := range responses {
			if row, ok := pvMap[r.DocumentNumber]; ok {
				pvStatus := ""
				if row.Status != nil {
					pvStatus = *row.Status
				}
				responses[i].LinkedPV = &types.LinkedPVSummary{
					ID:             row.ID,
					DocumentNumber: row.DocumentNumber,
					Status:         pvStatus,
				}
			}
		}
	}

	logger.Info("purchase_orders_retrieved_successfully")

	// Resolve creator into a {id,name,email,role} object.
	if len(responses) > 0 {
		ids := make([]string, 0, len(responses))
		for _, r := range responses {
			ids = append(ids, r.CreatedBy)
		}
		users := utils.ResolveUserRefs(config.DB, tenant.OrganizationID, ids)
		for i := range responses {
			if u, ok := users[responses[i].CreatedBy]; ok {
				creator := u
				responses[i].Creator = &creator
			}
		}
	}

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

	// Verify vendor exists if provided
	var vendorIDPtr *string
	if req.VendorID != "" {
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
		vendorIDPtr = &req.VendorID
	}

	// Generate PO number
	documentNumber := utils.GenerateDocumentNumber("PO")
	orderID := uuid.New().String()

	var createUser models.User
	config.DB.Where("id = ?", tenant.UserID).First(&createUser)

	logging.AddFieldToRequest(c, "document_number", documentNumber)
	logging.AddFieldToRequest(c, "order_id", orderID)

	order := models.PurchaseOrder{
		ID:                orderID,
		OrganizationID:    tenant.OrganizationID,
		DocumentNumber:    documentNumber,
		VendorID:          vendorIDPtr,
		VendorName:        req.VendorName,
		Status:            models.StatusDraft,
		TotalAmount:       req.TotalAmount,
		Currency:          req.Currency,
		DeliveryDate:      req.DeliveryDate.Time,
		ApprovalStage:     0,
		LinkedRequisition: req.LinkedRequisition,
		ProcurementFlow:   req.ProcurementFlow,
		EstimatedCost:     req.EstimatedCost,
		CreatedBy:         tenant.UserID,
		Title:             req.Title,
		Description:       req.Description,
		Department:        req.Department,
		DepartmentID:      req.DepartmentID,
		Priority:          req.Priority,
		BudgetCode:        req.BudgetCode,
		CostCenter:        req.CostCenter,
		ProjectCode:       req.ProjectCode,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if len(req.Metadata) > 0 {
		if metaBytes, err := json.Marshal(req.Metadata); err == nil {
			order.Metadata = datatypes.JSON(metaBytes)
		}
	}

	order.Items = datatypes.NewJSONType(req.Items)
	order.ApprovalHistory = datatypes.NewJSONType([]types.ApprovalRecord{})
	createNow := time.Now()
	order.ActionHistory = datatypes.NewJSONType([]types.ActionHistoryEntry{{
		ID:              uuid.New().String(),
		Action:          "CREATE",
		ActionType:      "CREATE",
		PerformedBy:     tenant.UserID,
		PerformedByName: createUser.Name,
		PerformedByRole: createUser.Role,
		Timestamp:       createNow,
		PerformedAt:     createNow,
		Comments:        "Purchase order created",
		NewStatus:       models.StatusDraft,
	}})

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

	go utils.SyncDocumentAs(config.DB, "PURCHASE_ORDER", order.ID, tenant.UserID)
	go services.LogDocumentEvent(config.DB, services.DocumentEvent{
		OrganizationID: tenant.OrganizationID,
		DocumentID:     order.ID,
		DocumentType:   "purchase_order",
		UserID:         tenant.UserID,
		ActorName:      createUser.Name,
		ActorRole:      tenant.UserRole,
		Action:         "created",
		Details:        map[string]interface{}{"documentNumber": order.DocumentNumber},
	})

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

	// Scope to what the caller is actually allowed to see. Previously this
	// endpoint only filtered by organization_id, so any user with the UUID
	// could view any PO in their org — bypassing the list endpoint's scope.
	// 404 (not 403) on scope miss keeps document existence private.
	scope := utils.GetDocumentScope(config.DB, tenant.UserID, tenant.UserRole, tenant.OrganizationID)
	query := config.DB.
		Preload("Vendor").
		Where("id = ? AND organization_id = ?", id, tenant.OrganizationID)
	query = scope.ApplyToQuery(query, "created_by", "purchase_order", "")

	var order models.PurchaseOrder
	if err := query.First(&order).Error; err != nil {
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

	response := modelToPurchaseOrderResponse(order)
	if liveHistory := utils.GetDocumentApprovalHistory(config.DB, order.ID, "purchase_order"); len(liveHistory) > 0 {
		response.ApprovalHistory = liveHistory
	}
	return c.JSON(types.DetailResponse{
		Success: true,
		Data:    response,
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

	tenant, terr := middleware.GetTenantContext(c)
	if terr != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Organization context required",
		})
	}

	// SECURITY: scope to org + owner/involvement — prevents cross-tenant (IDOR)
	// and cross-user edits of a purchase order.
	scope := utils.GetDocumentScope(config.DB, tenant.UserID, tenant.UserRole, tenant.OrganizationID)
	loadQuery := config.DB.Where("id = ? AND organization_id = ?", id, tenant.OrganizationID)
	loadQuery = scope.ApplyToQuery(loadQuery, "created_by", "purchase_order", "")

	var order models.PurchaseOrder
	if err := loadQuery.First(&order).Error; err != nil {
		logging.LogError(c, err, "purchase_order_not_found", map[string]interface{}{
			"order_id":   id,
			"error_type": "not_found",
		})
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Purchase order not found",
		})
	}

	// Metadata-only updates (quotations, attachments, bypass fields) are allowed on any status
	isMetadataOnly := len(req.Metadata) > 0 &&
		req.VendorID == "" &&
		len(req.Items) == 0 && req.TotalAmount == 0 &&
		req.Currency == "" && req.DeliveryDate.Time.IsZero()
	if strings.ToUpper(order.Status) != "DRAFT" && strings.ToUpper(order.Status) != "PENDING" && !isMetadataOnly {
		logging.LogWarn(c, "invalid_status_for_update", map[string]interface{}{
			"current_status":  order.Status,
			"document_number": order.DocumentNumber,
		})
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": fmt.Sprintf("Cannot update purchase order in %s status", order.Status),
		})
	}

	// Capture old values BEFORE making changes for audit trail
	oldValues := map[string]interface{}{
		"title":        order.Title,
		"description":  order.Description,
		"department":   order.Department,
		"departmentId": order.DepartmentID,
		"priority":     order.Priority,
		"budgetCode":   order.BudgetCode,
		"costCenter":   order.CostCenter,
		"projectCode":  order.ProjectCode,
		"vendorId":     "",
		"totalAmount":  order.TotalAmount,
		"currency":     order.Currency,
		"deliveryDate": order.DeliveryDate,
	}
	if order.VendorID != nil {
		oldValues["vendorId"] = *order.VendorID
	}

	// Track changes for logging
	changes := make(map[string]interface{})

	if req.VendorID != "" {
		fromVendorID := ""
		if order.VendorID != nil {
			fromVendorID = *order.VendorID
		}
		if fromVendorID != req.VendorID {
			changes["vendorId"] = map[string]string{"old": fromVendorID, "new": req.VendorID}
		}
		order.VendorID = &req.VendorID
	}
	// Persist vendor name when a vendor change is part of this update.
	// Guard prevents metadata-only updates (quotation list saves) from
	// clobbering the stored name to empty.
	if req.VendorID != "" || req.VendorName != "" {
		if order.VendorName != req.VendorName {
			changes["vendorName"] = map[string]string{"old": order.VendorName, "new": req.VendorName}
		}
		order.VendorName = req.VendorName
	}
	if len(req.Items) > 0 {
		oldItems := order.Items.Data()
		changes["items"] = map[string]interface{}{
			"old": oldItems,
			"new": req.Items,
		}
		changes["itemsCount"] = map[string]int{"old": len(oldItems), "new": len(req.Items)}
		order.Items = datatypes.NewJSONType(req.Items)
	}
	if req.TotalAmount > 0 && req.TotalAmount != order.TotalAmount {
		changes["totalAmount"] = map[string]float64{"old": order.TotalAmount, "new": req.TotalAmount}
		order.TotalAmount = req.TotalAmount
	}
	if req.Currency != "" && req.Currency != order.Currency {
		changes["currency"] = map[string]string{"old": order.Currency, "new": req.Currency}
		order.Currency = req.Currency
	}
	if !req.DeliveryDate.Time.IsZero() && !req.DeliveryDate.Time.Equal(order.DeliveryDate) {
		changes["deliveryDate"] = map[string]interface{}{
			"old": order.DeliveryDate.Format(time.RFC3339),
			"new": req.DeliveryDate.Time.Format(time.RFC3339),
		}
		order.DeliveryDate = req.DeliveryDate.Time
	}
	if len(req.Metadata) > 0 {
		// Deep-merge incoming metadata with existing — never wipe keys like
		// quotations, attachments, selectedQuotationFileUrl that other parts
		// of the UI manage independently.
		existingMeta := map[string]interface{}{}
		if len(order.Metadata) > 0 {
			_ = json.Unmarshal(order.Metadata, &existingMeta)
		}
		for k, v := range req.Metadata {
			existingMeta[k] = v
		}
		if metaBytes, err := json.Marshal(existingMeta); err == nil {
			order.Metadata = datatypes.JSON(metaBytes)
			changes["metadata"] = "updated"
		}
	}
	if req.QuotationGateOverridden != nil {
		order.QuotationGateOverridden = *req.QuotationGateOverridden
	}
	if req.BypassJustification != "" {
		order.BypassJustification = req.BypassJustification
	}
	if req.Title != "" && req.Title != order.Title {
		changes["title"] = map[string]string{"old": order.Title, "new": req.Title}
		order.Title = req.Title
	}
	if req.Description != "" && req.Description != order.Description {
		changes["description"] = map[string]string{"old": order.Description, "new": req.Description}
		order.Description = req.Description
	}
	if req.Department != "" && req.Department != order.Department {
		changes["department"] = map[string]string{"old": order.Department, "new": req.Department}
		order.Department = req.Department
	}
	if req.DepartmentID != "" && req.DepartmentID != order.DepartmentID {
		changes["departmentId"] = map[string]string{"old": order.DepartmentID, "new": req.DepartmentID}
		order.DepartmentID = req.DepartmentID
	}
	if req.Priority != "" && req.Priority != order.Priority {
		changes["priority"] = map[string]string{"old": order.Priority, "new": req.Priority}
		order.Priority = req.Priority
	}
	if req.BudgetCode != "" && req.BudgetCode != order.BudgetCode {
		changes["budgetCode"] = map[string]string{"old": order.BudgetCode, "new": req.BudgetCode}
		order.BudgetCode = req.BudgetCode
	}
	if req.CostCenter != "" && req.CostCenter != order.CostCenter {
		changes["costCenter"] = map[string]string{"old": order.CostCenter, "new": req.CostCenter}
		order.CostCenter = req.CostCenter
	}
	if req.ProjectCode != "" && req.ProjectCode != order.ProjectCode {
		changes["projectCode"] = map[string]string{"old": order.ProjectCode, "new": req.ProjectCode}
		order.ProjectCode = req.ProjectCode
	}

	orgID, _ := c.Locals("organizationID").(string)
	actorID, _ := c.Locals("userID").(string)
	actorRole, _ := c.Locals("userRole").(string)
	var updateUser models.User
	config.DB.Where("id = ?", actorID).First(&updateUser)
	if len(changes) > 0 {
		updateNow := time.Now()
		existingHistory := order.ActionHistory.Data()
		existingHistory = append(existingHistory, types.ActionHistoryEntry{
			ID:              uuid.New().String(),
			Action:          "UPDATE",
			ActionType:      "UPDATE",
			PerformedBy:     actorID,
			PerformedByName: updateUser.Name,
			PerformedByRole: actorRole,
			Timestamp:       updateNow,
			PerformedAt:     updateNow,
			Comments:        "Purchase order updated",
			NewStatus:       order.Status,
		})
		order.ActionHistory = datatypes.NewJSONType(existingHistory)
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

	go utils.SyncDocumentAs(config.DB, "PURCHASE_ORDER", order.ID, actorID)

	if len(changes) > 0 {
		// Create snapshot of current state after changes
		snapshot := services.CreateDocumentSnapshot(order)

		// Log the audit event with changes and snapshot for full transparency
		go services.LogDocumentEvent(config.DB, services.DocumentEvent{
			OrganizationID: orgID,
			DocumentID:     order.ID,
			DocumentType:   "purchase_order",
			UserID:         actorID,
			ActorName:      updateUser.Name,
			ActorRole:      actorRole,
			Action:         "updated",
			Changes:        changes,
			Snapshot:       snapshot,
			Details: map[string]interface{}{
				"documentNumber": order.DocumentNumber,
				"updateType":     "manual_edit",
			},
		})
	}

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

	tenant, terr := middleware.GetTenantContext(c)
	if terr != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Organization context required",
		})
	}

	// SECURITY: scope to org + owner/involvement (cross-tenant + cross-user guard).
	scope := utils.GetDocumentScope(config.DB, tenant.UserID, tenant.UserRole, tenant.OrganizationID)
	loadQuery := config.DB.Where("id = ? AND organization_id = ?", id, tenant.OrganizationID)
	loadQuery = scope.ApplyToQuery(loadQuery, "created_by", "purchase_order", "")

	var order models.PurchaseOrder
	if err := loadQuery.First(&order).Error; err != nil {
		logging.LogError(c, err, "purchase_order_not_found", map[string]interface{}{
			"order_id":   id,
			"error_type": "not_found",
		})
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Purchase order not found",
		})
	}

	if strings.ToUpper(order.Status) != "DRAFT" {
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

// WithdrawPurchaseOrder lets the creator pull a PENDING purchase order back to
// DRAFT, cancelling its in-flight workflow. Mirrors the payment voucher flow.
func WithdrawPurchaseOrder(c *fiber.Ctx) error {
	logger := logging.FromContext(c)
	logger.Info("withdraw_purchase_order_request")

	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Purchase Order ID is required",
		})
	}

	tenant, terr := middleware.GetTenantContext(c)
	if terr != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Organization context required",
		})
	}

	var order models.PurchaseOrder
	if err := config.DB.Where("id = ? AND organization_id = ?", id, tenant.OrganizationID).First(&order).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Purchase order not found",
		})
	}

	// Only the creator (or a privileged role) may withdraw.
	scope := utils.GetDocumentScope(config.DB, tenant.UserID, tenant.UserRole, tenant.OrganizationID)
	if order.CreatedBy != "" && order.CreatedBy != tenant.UserID && !scope.CanViewAll {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Only the creator can withdraw this purchase order",
		})
	}

	// Only PENDING purchase orders can be withdrawn.
	if strings.ToUpper(order.Status) != "PENDING" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": fmt.Sprintf("Cannot withdraw purchase order in %s status. Only pending purchase orders can be withdrawn.", order.Status),
		})
	}

	// Block withdrawal once an approver has actively claimed the review task.
	var task models.WorkflowTask
	if err := config.DB.Where(
		"entity_id = ? AND entity_type = ? AND UPPER(status) IN (?, ?)",
		id, "purchase_order", "PENDING", "CLAIMED",
	).First(&task).Error; err == nil {
		if strings.ToUpper(task.Status) == "CLAIMED" && task.ClaimedBy != nil {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"success": false,
				"message": "Cannot withdraw purchase order. It is currently being reviewed by an approver.",
			})
		}
	}

	tx := config.DB.Begin()
	if tx.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to start transaction",
		})
	}

	if err := tx.Where("entity_id = ? AND entity_type = ?", id, "purchase_order").
		Delete(&models.WorkflowTask{}).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to remove workflow tasks",
		})
	}
	if err := tx.Where("entity_id = ? AND entity_type = ?", id, "purchase_order").
		Delete(&models.WorkflowAssignment{}).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to remove workflow assignments",
		})
	}

	order.Status = "DRAFT"
	order.ApprovalStage = 0
	if err := tx.Save(&order).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to update purchase order status",
		})
	}

	if err := tx.Commit().Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to commit changes",
		})
	}

	logger.Info("purchase_order_withdrawn_successfully")
	return c.JSON(types.DetailResponse{
		Success: true,
		Data:    modelToPurchaseOrderResponse(order),
	})
}

// Helper function to convert model to response
// convertReqItemsToPOItems maps RequisitionItems to POItems for sync on submission.
func convertReqItemsToPOItems(reqItems []types.RequisitionItem) []types.POItem {
	poItems := make([]types.POItem, 0, len(reqItems))
	for _, ri := range reqItems {
		id := ""
		if ri.ID != nil {
			id = *ri.ID
		}
		unit := ""
		if ri.Unit != nil {
			unit = *ri.Unit
		}
		notes := ""
		if ri.Notes != nil {
			notes = *ri.Notes
		}
		category := ""
		if ri.Category != nil {
			category = *ri.Category
		}
		poItems = append(poItems, types.POItem{
			ID:          id,
			Description: ri.Description,
			Quantity:    ri.Quantity,
			UnitPrice:   ri.UnitPrice,
			Amount:      ri.Amount,
			TotalPrice:  ri.Amount,
			Unit:        unit,
			Notes:       notes,
			Category:    category,
		})
	}
	return poItems
}

// buildItemChanges compares two slices of POItem by index and returns a slice
// of per-item change maps. Each map contains: itemId, field, old, new.
// Only fields that differ between the old and new item are included.
// Returns an empty slice when there are no changes.
func buildItemChanges(oldItems, newItems []types.POItem) []map[string]interface{} {
	changes := []map[string]interface{}{}

	minLen := len(oldItems)
	if len(newItems) < minLen {
		minLen = len(newItems)
	}

	for i := 0; i < minLen; i++ {
		old := oldItems[i]
		nw := newItems[i]

		// Determine a stable item identifier: prefer the ID field, fall back to index.
		itemID := old.ID
		if itemID == "" {
			itemID = nw.ID
		}
		if itemID == "" {
			itemID = fmt.Sprintf("%d", i)
		}

		if old.Description != nw.Description {
			changes = append(changes, map[string]interface{}{
				"itemId": itemID,
				"field":  "description",
				"old":    old.Description,
				"new":    nw.Description,
			})
		}
		if old.Quantity != nw.Quantity {
			changes = append(changes, map[string]interface{}{
				"itemId": itemID,
				"field":  "quantity",
				"old":    old.Quantity,
				"new":    nw.Quantity,
			})
		}
		if old.UnitPrice != nw.UnitPrice {
			changes = append(changes, map[string]interface{}{
				"itemId": itemID,
				"field":  "unitPrice",
				"old":    old.UnitPrice,
				"new":    nw.UnitPrice,
			})
		}
		// Compare TotalPrice; fall back to Amount when TotalPrice is zero.
		oldTotal := old.TotalPrice
		if oldTotal == 0 {
			oldTotal = old.Amount
		}
		newTotal := nw.TotalPrice
		if newTotal == 0 {
			newTotal = nw.Amount
		}
		if oldTotal != newTotal {
			changes = append(changes, map[string]interface{}{
				"itemId": itemID,
				"field":  "totalPrice",
				"old":    oldTotal,
				"new":    newTotal,
			})
		}
	}

	return changes
}

func modelToPurchaseOrderResponse(order models.PurchaseOrder) types.PurchaseOrderResponse {
	var items []types.POItem
	if len(order.Items.Data()) > 0 {
		items = order.Items.Data()
	}

	var approvalHistory []types.ApprovalRecord

	vendorID := ""
	if order.VendorID != nil {
		vendorID = *order.VendorID
	}
	vendorName := order.VendorName              // stored fallback
	var vendorResp *types.VendorResponse
	if order.Vendor != nil {
		vendorName = order.Vendor.Name           // canonical wins when relation present
		vr := modelToVendorResponse(*order.Vendor)
		vendorResp = &vr
	}

	actionHistory := order.ActionHistory.Data()

	srcReqID := ""
	if order.SourceRequisitionId != nil {
		srcReqID = *order.SourceRequisitionId
	}

	// Unmarshal metadata JSONB into map
	var metadata map[string]interface{}
	if len(order.Metadata) > 0 {
		_ = json.Unmarshal(order.Metadata, &metadata)
	}

	return types.PurchaseOrderResponse{
		ID:                      order.ID,
		OrganizationID:          order.OrganizationID,
		DocumentNumber:          order.DocumentNumber,
		VendorID:                vendorID,
		VendorName:              vendorName,
		Vendor:                  vendorResp,
		Status:                  order.Status,
		Items:                   items,
		TotalAmount:             order.TotalAmount,
		Currency:                order.Currency,
		DeliveryDate:            order.DeliveryDate,
		ApprovalStage:           order.ApprovalStage,
		ApprovalHistory:         approvalHistory,
		ActionHistory:           actionHistory,
		LinkedRequisition:       order.LinkedRequisition,
		SourceRequisitionId:     srcReqID,
		ProcurementFlow:         order.ProcurementFlow,
		Metadata:                metadata,
		EstimatedCost:           order.EstimatedCost,
		AutomationUsed:          order.AutomationUsed,
		QuotationGateOverridden: order.QuotationGateOverridden,
		BypassJustification:     order.BypassJustification,
		// Add missing fields that are stored in DB but not returned
		Title:                   order.Title,
		Description:             order.Description,
		Department:              order.Department,
		DepartmentID:            order.DepartmentID,
		Priority:                order.Priority,
		BudgetCode:              order.BudgetCode,
		CostCenter:              order.CostCenter,
		ProjectCode:             order.ProjectCode,
		CreatedBy:               order.CreatedBy,
		CreatedAt:               order.CreatedAt,
		UpdatedAt:               order.UpdatedAt,
	}
}

// metaFloat extracts a float from a metadata value that may be a JSON number or
// a numeric string (the frontend stores taxRate/deliveryCost either way).
func metaFloat(v interface{}) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case string:
		if f, err := strconv.ParseFloat(strings.TrimSpace(n), 64); err == nil {
			return f
		}
	}
	return 0
}

// recomputePOTotal returns the authoritative PO total: Σ(quantity × unitPrice)
// plus tax (taxRate% of the subtotal) and deliveryCost read from the PO
// metadata. Mirrors the frontend items-editor computation so server and client
// agree, while never trusting a client-sent total.
func recomputePOTotal(items []types.POItem, meta datatypes.JSON) float64 {
	var subtotal float64
	for _, it := range items {
		subtotal += float64(it.Quantity) * it.UnitPrice
	}
	var taxRate, deliveryCost float64
	if len(meta) > 0 {
		var m map[string]interface{}
		if json.Unmarshal(meta, &m) == nil {
			taxRate = metaFloat(m["taxRate"])
			deliveryCost = metaFloat(m["deliveryCost"])
		}
	}
	tax := 0.0
	if taxRate > 0 {
		tax = math.Round((subtotal*taxRate/100.0)*100) / 100
	}
	return subtotal + tax + deliveryCost
}

// UpdatePurchaseOrderItems updates only the line items and total amount of a DRAFT purchase order
func UpdatePurchaseOrderItems(c *fiber.Ctx) error {
	logger := logging.FromContext(c)
	logger.Info("update_purchase_order_items_request")

	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Organization context required",
			"error":   err.Error(),
		})
	}

	id := c.Params("id")

	var req types.UpdatePurchaseOrderItemsRequest
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

	if len(req.Items) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "At least one item is required",
		})
	}
	for _, item := range req.Items {
		if item.Quantity <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "All items must have positive quantities",
			})
		}
	}
	if req.TotalAmount <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Total amount must be greater than 0",
		})
	}

	// SECURITY: scope to org + owner/involvement — prevents cross-tenant (IDOR)
	// and cross-user edits of a purchase order's line items. Mirrors UpdatePurchaseOrder.
	scope := utils.GetDocumentScope(config.DB, tenant.UserID, tenant.UserRole, tenant.OrganizationID)
	loadQuery := config.DB.Where("id = ? AND organization_id = ?", id, tenant.OrganizationID)
	loadQuery = scope.ApplyToQuery(loadQuery, "created_by", "purchase_order", "")

	var order models.PurchaseOrder
	if err := loadQuery.First(&order).Error; err != nil {
		logging.LogError(c, err, "purchase_order_not_found", map[string]interface{}{
			"order_id":   id,
			"error_type": "not_found",
		})
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Purchase order not found",
		})
	}

	if strings.ToUpper(order.Status) != "DRAFT" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Line items can only be edited on DRAFT purchase orders",
		})
	}

	oldItems := order.Items.Data()
	oldTotalAmount := order.TotalAmount

	order.Items = datatypes.NewJSONType(req.Items)
	// Recompute the total server-side from the submitted line items plus the PO's
	// existing tax/delivery, so the stored total can never diverge from the items.
	// The client-sent totalAmount is treated as a hint, not the source of truth.
	order.TotalAmount = recomputePOTotal(req.Items, order.Metadata)

	actorID := tenant.UserID
	actorRole := tenant.UserRole
	var actorUser models.User
	config.DB.Where("id = ?", actorID).First(&actorUser)

	updateNow := time.Now()
	existingHistory := order.ActionHistory.Data()
	existingHistory = append(existingHistory, types.ActionHistoryEntry{
		ID:              uuid.New().String(),
		Action:          "UPDATE",
		ActionType:      "UPDATE",
		PerformedBy:     actorID,
		PerformedByName: actorUser.Name,
		PerformedByRole: actorRole,
		Timestamp:       updateNow,
		PerformedAt:     updateNow,
		Comments:        "Line items updated",
		NewStatus:       order.Status,
	})
	order.ActionHistory = datatypes.NewJSONType(existingHistory)

	order.UpdatedAt = time.Now()

	if err := config.DB.Save(&order).Error; err != nil {
		logging.LogError(c, err, "failed_to_update_purchase_order_items", map[string]interface{}{
			"error_type":      "database_error",
			"document_number": order.DocumentNumber,
		})
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to update purchase order items",
			"error":   err.Error(),
		})
	}

	itemChanges := buildItemChanges(oldItems, req.Items)

	changes := map[string]interface{}{
		"items": map[string]interface{}{
			"old": oldItems,
			"new": req.Items,
		},
		"itemChanges": itemChanges,
	}
	if req.TotalAmount != oldTotalAmount {
		changes["totalAmount"] = map[string]float64{
			"old": oldTotalAmount,
			"new": req.TotalAmount,
		}
	}

	snapshot := services.CreateDocumentSnapshot(order)

	go services.LogDocumentEvent(config.DB, services.DocumentEvent{
		OrganizationID: tenant.OrganizationID,
		DocumentID:     order.ID,
		DocumentType:   "purchase_order",
		UserID:         actorID,
		ActorName:      actorUser.Name,
		ActorRole:      actorRole,
		Action:         "line_items_updated",
		Changes:        changes,
		Snapshot:       snapshot,
	})

	config.DB.Preload("Vendor").First(&order)

	logger.Info("purchase_order_items_updated_successfully")

	return c.JSON(types.DetailResponse{
		Success: true,
		Data:    modelToPurchaseOrderResponse(order),
	})
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
		"operation":       "submit_purchase_order",
		"order_id":        id,
		"organization_id": organizationID,
		"user_id":         userID,
	})

	// Get existing purchase order
	var order models.PurchaseOrder
	if err := config.DB.Where("id = ? AND organization_id = ? AND deleted_at IS NULL", id, organizationID).First(&order).Error; err != nil {
		logging.LogError(c, err, "purchase_order_not_found", map[string]interface{}{
			"order_id":        id,
			"organization_id": organizationID,
			"user_id":         userID,
			"error_detail":    err.Error(),
		})
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Purchase Order not found",
		})
	}

	// Check if purchase order is in draft status
	if strings.ToUpper(order.Status) != "DRAFT" {
		logging.LogWarn(c, "invalid_purchase_order_status_for_submission", map[string]interface{}{
			"current_status": order.Status,
		})
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": fmt.Sprintf("Cannot submit purchase order in %s status", order.Status),
		})
	}

	// Quotation gate: require 3 quotations unless auto-PO or bypass approved
	if !order.AutomationUsed {
		var quotations []types.Quotation
		if len(order.Metadata) > 0 {
			var meta map[string]interface{}
			if err := json.Unmarshal(order.Metadata, &meta); err == nil {
				if rawQ, ok := meta["quotations"]; ok {
					if qBytes, err := json.Marshal(rawQ); err == nil {
						_ = json.Unmarshal(qBytes, &quotations)
					}
				}
			}
		}
		quotationCount := len(quotations)
		if quotationCount < 3 {
			if !order.QuotationGateOverridden || order.BypassJustification == "" {
				return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
					"success": false,
					"error":   "quotation_required",
					"message": fmt.Sprintf("At least 3 quotations are required before submission. Currently %d attached.", quotationCount),
					"count":   quotationCount,
				})
			}
			// Bypass approved — add to action history and log audit event
			var bypassUser models.User
			config.DB.Where("id = ?", userID).First(&bypassUser)
			bypassTime := time.Now()
			bypassHistory := order.ActionHistory.Data()
			bypassHistory = append(bypassHistory, types.ActionHistoryEntry{
				ID:              uuid.New().String(),
				Action:          "QUOTATION_GATE_BYPASSED",
				ActionType:      "QUOTATION_GATE_BYPASSED",
				PerformedBy:     userID,
				PerformedByName: bypassUser.Name,
				PerformedByRole: bypassUser.Role,
				Timestamp:       bypassTime,
				PerformedAt:     bypassTime,
				Comments:        order.BypassJustification,
				NewStatus:       order.Status,
			})
			order.ActionHistory = datatypes.NewJSONType(bypassHistory)
			go services.LogDocumentEvent(config.DB, services.DocumentEvent{
				OrganizationID: organizationID,
				DocumentID:     order.ID,
				DocumentType:   "purchase_order",
				UserID:         userID,
				ActorName:      bypassUser.Name,
				ActorRole:      func() string { r, _ := c.Locals("userRole").(string); return r }(),
				Action:         "quotation_gate_bypassed",
				Details: map[string]interface{}{
					"justification":  order.BypassJustification,
					"quotationCount": quotationCount,
				},
			})
		}
	}

	// Gate + sync: if linked to a REQ, it must be APPROVED before PO can be submitted
	if order.SourceRequisitionId != nil && *order.SourceRequisitionId != "" {
		var req models.Requisition
		if err := config.DB.First(&req, "id = ? AND organization_id = ?", *order.SourceRequisitionId, organizationID).Error; err != nil {
			return utils.SendBadRequestError(c, "Linked requisition not found")
		}
		if strings.ToUpper(req.Status) != "APPROVED" {
			return utils.SendBadRequestError(c, fmt.Sprintf(
				"Cannot submit PO: linked requisition %s is in %s status and must be APPROVED first.",
				req.DocumentNumber, req.Status))
		}

		// Sync items and amounts from the approved REQ
		reqItems := req.Items.Data()
		poItems := convertReqItemsToPOItems(reqItems)
		order.Items = datatypes.NewJSONType(poItems)
		order.TotalAmount = req.TotalAmount
		order.Currency = req.Currency
		if req.PreferredVendorID != nil && *req.PreferredVendorID != "" {
			order.VendorID = req.PreferredVendorID
		}
		order.UpdatedAt = time.Now()
		if err := config.DB.Save(&order).Error; err != nil {
			return utils.SendInternalError(c, "Failed to sync PO data from requisition", err)
		}
	}

	// Get workflow execution service from context
	workflowExecutionService := c.Locals("workflowExecutionService").(*services.WorkflowExecutionService)

	// Status transition + workflow assignment must be atomic: either both persist
	// or neither does. Otherwise we risk orphan docs (PENDING with no workflow)
	// or orphan workflows (assignment with doc still DRAFT).
	tx := config.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	assignment, err := workflowExecutionService.AssignWorkflowToDocumentWithIDTx(
		c.Context(), tx, organizationID, order.ID, "purchase_order", submitReq.WorkflowID, userID,
	)
	if err != nil {
		tx.Rollback()
		logging.LogError(c, err, "workflow_assignment_failed", map[string]interface{}{
			"order_id": id,
		})
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to assign workflow to purchase order",
			"error":   err.Error(),
		})
	}

	order.Status = models.StatusPending
	order.UpdatedAt = time.Now()

	// Use the open tx to avoid deadlock under single-conn DB pools.
	var user models.User
	_ = tx.Where("id = ?", userID).First(&user).Error
	submitTime := time.Now()
	actionHistory := order.ActionHistory.Data()
	actionHistory = append(actionHistory, types.ActionHistoryEntry{
		ID:              uuid.New().String(),
		Action:          "SUBMIT",
		ActionType:      "SUBMIT",
		PerformedBy:     userID,
		PerformedByName: user.Name,
		PerformedByRole: user.Role,
		Timestamp:       submitTime,
		PerformedAt:     submitTime,
		Comments:        "Purchase order submitted for approval",
		PreviousStatus:  models.StatusDraft,
		NewStatus:       models.StatusPending,
	})
	order.ActionHistory = datatypes.NewJSONType(actionHistory)

	if err := tx.Save(&order).Error; err != nil {
		tx.Rollback()
		logging.LogError(c, err, "purchase_order_update_failed", map[string]interface{}{
			"order_id": id,
		})
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to update purchase order status",
			"error":   err.Error(),
		})
	}

	if err := tx.Commit().Error; err != nil {
		logging.LogError(c, err, "purchase_order_submit_commit_failed", map[string]interface{}{
			"order_id": id,
		})
		return utils.SendInternalError(c, "Failed to submit purchase order", err)
	}

	// Preload vendor
	config.DB.Preload("Vendor").First(&order)

	go utils.SyncDocumentAs(config.DB, "PURCHASE_ORDER", order.ID, userID)
	go services.LogDocumentEvent(config.DB, services.DocumentEvent{
		OrganizationID: organizationID,
		DocumentID:     order.ID,
		DocumentType:   "purchase_order",
		UserID:         userID,
		ActorName:      user.Name,
		ActorRole:      user.Role,
		Action:         "submitted",
		Details:        map[string]interface{}{"documentNumber": order.DocumentNumber, "workflowId": submitReq.WorkflowID},
	})

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
