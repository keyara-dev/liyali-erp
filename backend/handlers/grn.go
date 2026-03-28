package handlers

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/middleware"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/services"
	"github.com/liyali/liyali-gateway/types"
	"github.com/liyali/liyali-gateway/utils"
	"gorm.io/datatypes"
)

// GetGRNs retrieves all goods received notes with pagination and filtering
func GetGRNs(c *fiber.Ctx) error {
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
	poDocumentNumber := c.Query("poDocumentNumber")

	// Determine document visibility scope for this user
	scope := utils.GetDocumentScope(db, tenant.UserID, tenant.UserRole, tenant.OrganizationID)

	// Start with organization filter - CRITICAL SECURITY FIX
	query := db.Where("organization_id = ?", tenant.OrganizationID)

	// Apply document scope (procurement users see all GRNs; limited users see own + involved)
	query = scope.ApplyToQuery(query, "created_by", "grn", "received_by")

	if status != "" {
		query = query.Where("UPPER(status) = UPPER(?)", status)
	}
	if poDocumentNumber != "" {
		query = query.Where("po_document_number = ?", poDocumentNumber)
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
	// Get organization context from tenant middleware
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Organization context required",
			"error":   err.Error(),
		})
	}

	var req types.CreateGRNRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	if req.PODocumentNumber == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "PO number is required",
		})
	}
	// Validate PO document number format (should start with "PO-" and be at least 10 characters)
	if len(req.PODocumentNumber) < 10 || req.PODocumentNumber[:3] != "PO-" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid PO document number format",
		})
	}
	if len(req.Items) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "At least one item is required",
		})
	}
	// Validate items have positive quantities
	for _, item := range req.Items {
		if item.QuantityOrdered <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "All items must have positive quantities",
			})
		}
	}
	if req.ReceivedBy == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "ReceivedBy is required",
		})
	}

	// Verify PO exists and belongs to organization
	var po models.PurchaseOrder
	if err := config.DB.Where("document_number = ? AND organization_id = ?", req.PODocumentNumber, tenant.OrganizationID).First(&po).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Purchase order not found",
		})
	}

	// Resolve effective procurement flow: PO override → org default → "goods_first"
	effectiveFlow := po.ProcurementFlow
	if effectiveFlow == "" {
		orgSvc := services.NewOrganizationService(config.DB)
		orgSettings, _ := orgSvc.GetOrganizationSettings(tenant.OrganizationID)
		if orgSettings != nil && orgSettings.ProcurementFlow != "" {
			effectiveFlow = orgSettings.ProcurementFlow
		} else {
			effectiveFlow = "goods_first"
		}
	}

	// One-to-one: reject if any non-cancelled GRN already exists for this PO/PV
	if effectiveFlow == "payment_first" && req.LinkedPV != "" {
		var existingGRN models.GoodsReceivedNote
		if err := config.DB.
			Where("linked_pv = ? AND organization_id = ? AND UPPER(status) != 'CANCELLED'",
				req.LinkedPV, tenant.OrganizationID).
			First(&existingGRN).Error; err == nil {
			return utils.SendConflictError(c, fmt.Sprintf(
				"Goods received note %s already exists for payment voucher %s (status: %s).",
				existingGRN.DocumentNumber, req.LinkedPV, existingGRN.Status))
		}
	} else {
		var existingGRN models.GoodsReceivedNote
		if err := config.DB.
			Where("po_document_number = ? AND organization_id = ? AND UPPER(status) != 'CANCELLED'",
				req.PODocumentNumber, tenant.OrganizationID).
			First(&existingGRN).Error; err == nil {
			return utils.SendConflictError(c, fmt.Sprintf(
				"Goods received note %s already exists for purchase order %s (status: %s).",
				existingGRN.DocumentNumber, req.PODocumentNumber, existingGRN.Status))
		}
	}

	// Payment-first enforcement: require an approved PV before goods can be received
	var linkedPVDoc *models.PaymentVoucher
	if effectiveFlow == "payment_first" {
		if req.LinkedPV == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "A linked payment voucher document number is required for payment-first procurement flow",
			})
		}
		var pv models.PaymentVoucher
		if err := config.DB.Where("document_number = ? AND organization_id = ?", req.LinkedPV, tenant.OrganizationID).First(&pv).Error; err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Linked payment voucher not found",
			})
		}
		if strings.ToUpper(pv.Status) != "APPROVED" && strings.ToUpper(pv.Status) != "PAID" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Linked payment voucher must be approved or paid before receiving goods (payment-first flow)",
			})
		}
		if pv.LinkedPO != po.DocumentNumber {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Linked payment voucher does not belong to the selected purchase order",
			})
		}
		linkedPVDoc = &pv
	}

	// Generate GRN number
	documentNumber := utils.GenerateDocumentNumber("GRN")

	linkedPVDocNum := ""
	if linkedPVDoc != nil {
		linkedPVDocNum = linkedPVDoc.DocumentNumber
	}

	// Build initial action history — chain origin
	var grnInitialHistory []types.ActionHistoryEntry
	if linkedPVDoc != nil {
		grnInitialHistory = append(grnInitialHistory, types.ActionHistoryEntry{
			ID:          uuid.New().String(),
			Action:      "CREATED_FROM_PV",
			PerformedBy: tenant.UserID,
			Timestamp:   time.Now(),
			Metadata: map[string]interface{}{
				"linkedDocNumber": linkedPVDoc.DocumentNumber,
				"linkedDocType":   "payment_voucher",
				"flow":            "payment_first",
			},
		})
	} else {
		grnInitialHistory = append(grnInitialHistory, types.ActionHistoryEntry{
			ID:          uuid.New().String(),
			Action:      "CREATED_FROM_PO",
			PerformedBy: tenant.UserID,
			Timestamp:   time.Now(),
			Metadata: map[string]interface{}{
				"linkedDocNumber": po.DocumentNumber,
				"linkedDocType":   "purchase_order",
				"flow":            "goods_first",
			},
		})
	}

	grn := models.GoodsReceivedNote{
		ID:                uuid.New().String(),
		OrganizationID:    tenant.OrganizationID,
		DocumentNumber:    documentNumber,
		PODocumentNumber:  req.PODocumentNumber,
		Status: "DRAFT",
		ReceivedDate:      time.Now(),
		ReceivedBy:        req.ReceivedBy,
		ApprovalStage:     0,
		LinkedPV:          linkedPVDocNum,
		WarehouseLocation: req.WarehouseLocation,
		Notes:             req.Notes,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	grn.Items = datatypes.NewJSONType(req.Items)

	emptyQuality := []types.QualityIssue{}
	grn.QualityIssues = datatypes.NewJSONType(emptyQuality)

	emptyHistory := []types.ApprovalRecord{}
	grn.ApprovalHistory = datatypes.NewJSONType(emptyHistory)
	grn.ActionHistory = datatypes.NewJSONType(grnInitialHistory)

	if err := config.DB.Create(&grn).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to create GRN",
			"error":   err.Error(),
		})
	}

	// Record GRN_CREATED on the parent document for chain traceability
	grnCreatedEntry := types.ActionHistoryEntry{
		ID:          uuid.New().String(),
		Action:      "GRN_CREATED",
		PerformedBy: tenant.UserID,
		Timestamp:   time.Now(),
		Metadata: map[string]interface{}{
			"linkedDocNumber": grn.DocumentNumber,
			"linkedDocType":   "grn",
			"flow":            effectiveFlow,
		},
	}
	if linkedPVDoc != nil {
		pvHistory := linkedPVDoc.ActionHistory.Data()
		pvHistory = append(pvHistory, grnCreatedEntry)
		linkedPVDoc.ActionHistory = datatypes.NewJSONType(pvHistory)
		config.DB.Save(linkedPVDoc)
	} else {
		poHistory := po.ActionHistory.Data()
		poHistory = append(poHistory, grnCreatedEntry)
		po.ActionHistory = datatypes.NewJSONType(poHistory)
		config.DB.Save(&po)
	}

	go utils.SyncDocument(config.DB, "GRN", grn.ID)

	return c.Status(fiber.StatusCreated).JSON(types.DetailResponse{
		Success: true,
		Data:    modelToGRNResponse(grn),
	})
}

// GetGRN retrieves a single GRN by ID
func GetGRN(c *fiber.Ctx) error {
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
			"message": "GRN ID is required",
		})
	}

	var grn models.GoodsReceivedNote
	// SECURITY FIX: Filter by organization ID
	if err := config.DB.Where("id = ? AND organization_id = ?", id, tenant.OrganizationID).First(&grn).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "GRN not found",
		})
	}

	response := modelToGRNResponse(grn)
	if liveHistory := utils.GetDocumentApprovalHistory(config.DB, grn.ID, "grn"); len(liveHistory) > 0 {
		response.ApprovalHistory = liveHistory
	}
	return c.JSON(types.DetailResponse{
		Success: true,
		Data:    response,
	})
}

// UpdateGRN updates an existing GRN
func UpdateGRN(c *fiber.Ctx) error {
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
	// SECURITY FIX: Filter by organization ID
	if err := config.DB.Where("id = ? AND organization_id = ?", id, tenant.OrganizationID).First(&grn).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "GRN not found",
		})
	}

	if strings.ToUpper(grn.Status) != "DRAFT" && strings.ToUpper(grn.Status) != "PENDING" {
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

	go utils.SyncDocument(config.DB, "GRN", grn.ID)

	return c.JSON(types.DetailResponse{
		Success: true,
		Data:    modelToGRNResponse(grn),
	})
}

// DeleteGRN deletes a GRN
func DeleteGRN(c *fiber.Ctx) error {
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
			"message": "GRN ID is required",
		})
	}

	var grn models.GoodsReceivedNote
	// SECURITY FIX: Filter by organization ID
	if err := config.DB.Where("id = ? AND organization_id = ?", id, tenant.OrganizationID).First(&grn).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "GRN not found",
		})
	}

	if strings.ToUpper(grn.Status) != "DRAFT" {
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

// Helper function to convert model to response
func modelToGRNResponse(grn models.GoodsReceivedNote) types.GRNResponse {
	var items []types.GRNItem
	items = grn.Items.Data()

	var qualityIssues []types.QualityIssue
	qualityIssues = grn.QualityIssues.Data()

	var approvalHistory []types.ApprovalRecord
	approvalHistory = grn.ApprovalHistory.Data()

	var actionHistory []types.ActionHistoryEntry
	actionHistory = grn.ActionHistory.Data()

	return types.GRNResponse{
		ID:               grn.ID,
		DocumentNumber:   grn.DocumentNumber,
		PODocumentNumber: grn.PODocumentNumber,
		Status:           grn.Status,
		ReceivedDate:     grn.ReceivedDate,
		ReceivedBy:       grn.ReceivedBy,
		Items:            items,
		QualityIssues:    qualityIssues,
		ApprovalStage:    grn.ApprovalStage,
		ApprovalHistory:  approvalHistory,
		ActionHistory:    actionHistory,
		LinkedPV:         grn.LinkedPV,
		CreatedAt:        grn.CreatedAt,
		UpdatedAt:        grn.UpdatedAt,
	}
}

// SubmitGRN submits a GRN for approval using the workflow system
func SubmitGRN(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "GRN ID is required",
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

	// Get existing GRN
	var grn models.GoodsReceivedNote
	if err := config.DB.Where("id = ? AND organization_id = ?", id, organizationID).First(&grn).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "GRN not found",
		})
	}

	// Check if GRN is in draft status
	if strings.ToUpper(grn.Status) != "DRAFT" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": fmt.Sprintf("Cannot submit GRN in %s status", grn.Status),
		})
	}

	// Gate: linked PO must still be APPROVED before GRN can be submitted
	if grn.PODocumentNumber != "" {
		var linkedPO models.PurchaseOrder
		if err := config.DB.
			Where("document_number = ? AND organization_id = ?", grn.PODocumentNumber, organizationID).
			First(&linkedPO).Error; err != nil {
			return utils.SendBadRequestError(c, "Linked purchase order not found")
		}
		if strings.ToUpper(linkedPO.Status) != "APPROVED" {
			return utils.SendBadRequestError(c, fmt.Sprintf(
				"Cannot submit GRN: linked PO %s is in %s status and must be APPROVED.",
				grn.PODocumentNumber, linkedPO.Status))
		}
	}

	// Gate: payment-first — linked PV must still be APPROVED or PAID
	if grn.LinkedPV != "" {
		var linkedPV models.PaymentVoucher
		if err := config.DB.
			Where("document_number = ? AND organization_id = ?", grn.LinkedPV, organizationID).
			First(&linkedPV).Error; err != nil {
			return utils.SendBadRequestError(c, "Linked payment voucher not found")
		}
		pvStatus := strings.ToUpper(linkedPV.Status)
		if pvStatus != "APPROVED" && pvStatus != "PAID" {
			return utils.SendBadRequestError(c, fmt.Sprintf(
				"Cannot submit GRN: linked PV %s is in %s status and must be APPROVED or PAID.",
				grn.LinkedPV, linkedPV.Status))
		}
	}

	// Get workflow execution service from context
	workflowExecutionService := c.Locals("workflowExecutionService").(*services.WorkflowExecutionService)

	// Assign workflow to the GRN
	assignment, err := workflowExecutionService.AssignWorkflowToDocumentWithID(
		c.Context(), organizationID, grn.ID, "grn", submitReq.WorkflowID, userID,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to assign workflow to GRN",
			"error":   err.Error(),
		})
	}

	// Update GRN status to pending
	grn.Status = "PENDING"
	grn.UpdatedAt = time.Now()

	// Add action history entry for submission
	var actionHistory []types.ActionHistoryEntry
	actionHistory = grn.ActionHistory.Data()

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
			Comments:        "GRN submitted for approval",
			ActionType:      "SUBMIT",
			PreviousStatus:  "DRAFT",
			NewStatus:       "PENDING",
		})
		grn.ActionHistory = datatypes.NewJSONType(actionHistory)
	}

	// Save GRN
	if err := config.DB.Save(&grn).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to update GRN status",
			"error":   err.Error(),
		})
	}

	// Preload purchase order and vendor
	config.DB.Preload("PurchaseOrder").Preload("Vendor").First(&grn)

	go utils.SyncDocument(config.DB, "GRN", grn.ID)

	return c.JSON(types.DetailResponse{
		Success: true,
		Data: fiber.Map{
			"grn": modelToGRNResponse(grn),
			"workflow": fiber.Map{
				"assignmentId": assignment.ID,
				"workflowId":   assignment.WorkflowID,
				"currentStage": assignment.CurrentStage,
				"status":       assignment.Status,
			},
		},
	})
}
