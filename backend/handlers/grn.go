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

	scope := utils.GetDocumentScope(config.DB, tenant.UserID, tenant.UserRole, tenant.OrganizationID)

	ctx := c.Context()
	offset := int32((page - 1) * limit)
	orgRoleIDs := scope.OrgRoleIDs
	if orgRoleIDs == nil {
		orgRoleIDs = []string{}
	}

	var total int64
	var ids []string

	if scope.CanViewAll || scope.IsProcurement {
		total, err = config.Queries.CountGRNsAll(ctx, tenant.OrganizationID, status, poDocumentNumber)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to count GRNs",
				"error":   err.Error(),
			})
		}
		ids, err = config.Queries.ListGRNIDsAll(ctx, tenant.OrganizationID, status, poDocumentNumber, int32(limit), offset)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to fetch GRNs",
				"error":   err.Error(),
			})
		}
	} else {
		total, err = config.Queries.CountGRNsLimited(ctx, db.CountGRNsLimitedParams{
			OrganizationID: tenant.OrganizationID,
			Column2:        status,
			Column3:        poDocumentNumber,
			CreatedBy:      &scope.UserID,
			Lower:          scope.UserRole,
			Column6:        orgRoleIDs,
		})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to count GRNs",
				"error":   err.Error(),
			})
		}
		ids, err = config.Queries.ListGRNIDsLimited(ctx, db.ListGRNIDsLimitedParams{
			OrganizationID: tenant.OrganizationID,
			Column2:        status,
			Column3:        poDocumentNumber,
			CreatedBy:      &scope.UserID,
			Lower:          scope.UserRole,
			Column6:        orgRoleIDs,
			Limit:          int32(limit),
			Offset:         offset,
		})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to fetch GRNs",
				"error":   err.Error(),
			})
		}
	}

	var grns []models.GoodsReceivedNote
	if len(ids) > 0 {
		if err := config.DB.
			Where("id IN ?", ids).
			Order("created_at DESC").
			Find(&grns).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to fetch GRNs",
				"error":   err.Error(),
			})
		}
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

	// Goods-first: the PO must be APPROVED before goods can be received against it.
	// Payment-first enforces the PV-approval gate further down; no PO-status gate there.
	if effectiveFlow != "payment_first" && strings.ToUpper(po.Status) != "APPROVED" {
		return utils.SendBadRequestError(c, fmt.Sprintf(
			"Cannot create GRN: linked PO %s is in %s status and must be APPROVED first.",
			po.DocumentNumber, po.Status))
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
		CreatedBy:         tenant.UserID,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	grn.Items = datatypes.NewJSONType(req.Items)

	emptyQuality := []types.QualityIssue{}
	grn.QualityIssues = datatypes.NewJSONType(emptyQuality)

	emptyHistory := []types.ApprovalRecord{}
	grn.ApprovalHistory = datatypes.NewJSONType(emptyHistory)
	var grnCreateUser models.User
	config.DB.Where("id = ?", tenant.UserID).First(&grnCreateUser)
	grnCreateNow := time.Now()
	grnInitialHistory = append(grnInitialHistory, types.ActionHistoryEntry{
		ID:              uuid.New().String(),
		Action:          "CREATE",
		ActionType:      "CREATE",
		PerformedBy:     tenant.UserID,
		PerformedByName: grnCreateUser.Name,
		PerformedByRole: grnCreateUser.Role,
		Timestamp:       grnCreateNow,
		PerformedAt:     grnCreateNow,
		Comments:        "GRN created",
		NewStatus:       "DRAFT",
	})
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
	go services.LogDocumentEvent(config.DB, services.DocumentEvent{
		OrganizationID: tenant.OrganizationID,
		DocumentID:     grn.ID,
		DocumentType:   "grn",
		UserID:         tenant.UserID,
		ActorName:      grnCreateUser.Name,
		ActorRole:      tenant.UserRole,
		Action:         "created",
		Details:        map[string]interface{}{"documentNumber": grn.DocumentNumber},
	})

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

	var grnUpdateUser models.User
	config.DB.Where("id = ?", tenant.UserID).First(&grnUpdateUser)
	grnUpdateNow := time.Now()
	grnUpdateHistory := grn.ActionHistory.Data()
	grnUpdateHistory = append(grnUpdateHistory, types.ActionHistoryEntry{
		ID:              uuid.New().String(),
		Action:          "UPDATE",
		ActionType:      "UPDATE",
		PerformedBy:     tenant.UserID,
		PerformedByName: grnUpdateUser.Name,
		PerformedByRole: grnUpdateUser.Role,
		Timestamp:       grnUpdateNow,
		PerformedAt:     grnUpdateNow,
		Comments:        "GRN updated",
		NewStatus:       grn.Status,
	})
	grn.ActionHistory = datatypes.NewJSONType(grnUpdateHistory)
	grn.UpdatedAt = grnUpdateNow

	if err := config.DB.Save(&grn).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to update GRN",
			"error":   err.Error(),
		})
	}

	go utils.SyncDocument(config.DB, "GRN", grn.ID)
	go services.LogDocumentEvent(config.DB, services.DocumentEvent{
		OrganizationID: tenant.OrganizationID,
		DocumentID:     grn.ID,
		DocumentType:   "grn",
		UserID:         tenant.UserID,
		ActorName:      grnUpdateUser.Name,
		ActorRole:      tenant.UserRole,
		Action:         "updated",
		Details:        map[string]interface{}{"documentNumber": grn.DocumentNumber},
	})

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
	
	// Unmarshal metadata
	var metadata map[string]interface{}
	if len(grn.Metadata) > 0 {
		_ = json.Unmarshal(grn.Metadata, &metadata)
	}
	
	// Unmarshal autoCreatedPV
	var autoCreatedPV interface{}
	if len(grn.AutoCreatedPV) > 0 {
		_ = json.Unmarshal(grn.AutoCreatedPV, &autoCreatedPV)
	}

	return types.GRNResponse{
		ID:                grn.ID,
		OrganizationID:    grn.OrganizationID,
		DocumentNumber:    grn.DocumentNumber,
		PODocumentNumber:  grn.PODocumentNumber,
		Status:            grn.Status,
		ReceivedDate:      grn.ReceivedDate,
		ReceivedBy:        grn.ReceivedBy,
		Items:             items,
		QualityIssues:     qualityIssues,
		ApprovalStage:     grn.ApprovalStage,
		ApprovalHistory:   approvalHistory,
		ActionHistory:     actionHistory,
		LinkedPV:          grn.LinkedPV,
		BudgetCode:        grn.BudgetCode,
		CostCenter:        grn.CostCenter,
		ProjectCode:       grn.ProjectCode,
		CreatedBy:         grn.CreatedBy,
		OwnerID:           grn.OwnerID,
		WarehouseLocation: grn.WarehouseLocation,
		Notes:             grn.Notes,
		CurrentStage:      grn.CurrentStage,
		StageName:         grn.StageName,
		ApprovedBy:        grn.ApprovedBy,
		AutomationUsed:    grn.AutomationUsed,
		AutoCreatedPV:     autoCreatedPV,
		Metadata:          metadata,
		CreatedAt:         grn.CreatedAt,
		UpdatedAt:         grn.UpdatedAt,
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

	// Atomic submit: status change + workflow assignment in a single transaction.
	tx := config.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	assignment, err := workflowExecutionService.AssignWorkflowToDocumentWithIDTx(
		c.Context(), tx, organizationID, grn.ID, "grn", submitReq.WorkflowID, userID,
	)
	if err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to assign workflow to GRN",
			"error":   err.Error(),
		})
	}

	grn.Status = "PENDING"
	grn.UpdatedAt = time.Now()

	var user models.User
	_ = config.DB.Where("id = ?", userID).First(&user).Error
	actionHistory := grn.ActionHistory.Data()
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

	if err := tx.Save(&grn).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to update GRN status",
			"error":   err.Error(),
		})
	}

	if err := tx.Commit().Error; err != nil {
		return utils.SendInternalError(c, "Failed to submit GRN", err)
	}

	// Preload purchase order and vendor
	config.DB.Preload("PurchaseOrder").Preload("Vendor").First(&grn)

	go utils.SyncDocument(config.DB, "GRN", grn.ID)
	go services.LogDocumentEvent(config.DB, services.DocumentEvent{
		OrganizationID: organizationID,
		DocumentID:     grn.ID,
		DocumentType:   "grn",
		UserID:         userID,
		ActorName:      user.Name,
		ActorRole:      user.Role,
		Action:         "submitted",
		Details:        map[string]interface{}{"documentNumber": grn.DocumentNumber},
	})

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
