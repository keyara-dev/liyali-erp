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
)

// GetRequisitions retrieves all requisitions with pagination and filtering
func GetRequisitions(c *fiber.Ctx) error {
	logger := logging.FromContext(c)
	logger.Info("get_requisitions_request")

	// Extract and normalize pagination parameters
	page, pageSize := utils.NormalizePaginationParams(
		c.QueryInt("page", 1),
		c.QueryInt("page_size", 10),
	)

	// Extract filter parameters
	status := c.Query("status")
	department := c.Query("department")
	priority := c.Query("priority")

	// Get tenant context (organization + user identity)
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Organization context required",
		})
	}
	organizationID := tenant.OrganizationID

	// Add query parameters to context
	logging.AddFieldsToRequest(c, map[string]interface{}{
		"operation":      "get_requisitions",
		"page":           page,
		"page_size":      pageSize,
		"status":         status,
		"department":     department,
		"priority":       priority,
		"organizationID": organizationID,
	})

	scope := utils.GetDocumentScope(config.DB, tenant.UserID, tenant.UserRole, organizationID)

	ctx := c.Context()
	offset := int32((page - 1) * pageSize)
	orgRoleIDs := scope.OrgRoleIDs
	if orgRoleIDs == nil {
		orgRoleIDs = []string{}
	}

	var total int64
	var ids []string

	switch {
	case scope.CanViewAll:
		total, err = config.Queries.CountRequisitionsAll(ctx, db.CountRequisitionsAllParams{
			OrganizationID:    organizationID,
			Column2:           status,
			Column3:           department,
			Column4:           priority,
			HideDirectPayment: scope.HideDirectPayment,
		})
		if err != nil {
			return utils.SendInternalError(c, "Failed to count requisitions", err)
		}
		ids, err = config.Queries.ListRequisitionIDsAll(ctx, db.ListRequisitionIDsAllParams{
			OrganizationID:    organizationID,
			Column2:           status,
			Column3:           department,
			Column4:           priority,
			HideDirectPayment: scope.HideDirectPayment,
			Limit:             int32(pageSize),
			Offset:            offset,
		})
		if err != nil {
			return utils.SendInternalError(c, "Failed to fetch requisitions", err)
		}
	case scope.IsProcurement:
		total, err = config.Queries.CountRequisitionsProcurement(ctx, db.CountRequisitionsProcurementParams{
			OrganizationID:    organizationID,
			Column2:           status,
			Column3:           department,
			Column4:           priority,
			HideDirectPayment: scope.HideDirectPayment,
		})
		if err != nil {
			return utils.SendInternalError(c, "Failed to count requisitions", err)
		}
		ids, err = config.Queries.ListRequisitionIDsProcurement(ctx, db.ListRequisitionIDsProcurementParams{
			OrganizationID:    organizationID,
			Column2:           status,
			Column3:           department,
			Column4:           priority,
			HideDirectPayment: scope.HideDirectPayment,
			Limit:             int32(pageSize),
			Offset:            offset,
		})
		if err != nil {
			return utils.SendInternalError(c, "Failed to fetch requisitions", err)
		}
	default:
		total, err = config.Queries.CountRequisitionsLimited(ctx, db.CountRequisitionsLimitedParams{
			OrganizationID: organizationID,
			Column2:        status,
			Column3:        department,
			Column4:        priority,
			RequesterID:    scope.UserID,
			Lower:          scope.UserRole,
			Column7:        orgRoleIDs,
		})
		if err != nil {
			return utils.SendInternalError(c, "Failed to count requisitions", err)
		}
		ids, err = config.Queries.ListRequisitionIDsLimited(ctx, db.ListRequisitionIDsLimitedParams{
			OrganizationID: organizationID,
			Column2:        status,
			Column3:        department,
			Column4:        priority,
			RequesterID:    scope.UserID,
			Lower:          scope.UserRole,
			Column7:        orgRoleIDs,
			Limit:          int32(pageSize),
			Offset:         offset,
		})
		if err != nil {
			return utils.SendInternalError(c, "Failed to fetch requisitions", err)
		}
	}

	var requisitions []models.Requisition
	if len(ids) > 0 {
		if err := config.DB.
			Where("id IN ?", ids).
			Preload("Requester").
			Preload("Category").
			Preload("PreferredVendor").
			Order("created_at DESC").
			Find(&requisitions).Error; err != nil {
			return utils.SendInternalError(c, "Failed to fetch requisitions", err)
		}
	}

	// Convert to response format
	responses := make([]types.RequisitionResponse, 0, len(requisitions))
	for _, req := range requisitions {
		responses = append(responses, modelToRequisitionResponse(req))
	}

	// Batch-enrich responses with linked PO info (single query, not N+1)
	if len(responses) > 0 {
		reqIDs := make([]string, len(responses))
		for i, r := range responses {
			reqIDs[i] = r.ID
		}
		poRows, _ := config.Queries.GetLinkedPOsForRequisitions(ctx, reqIDs, organizationID)
		poMap := make(map[string]db.GetLinkedPOsForRequisitionsRow, len(poRows))
		for _, r := range poRows {
			if r.SourceRequisitionID != nil {
				poMap[*r.SourceRequisitionID] = r
			}
		}
		for i, r := range responses {
			if row, ok := poMap[r.ID]; ok {
				poStatus := ""
				if row.Status != nil {
					poStatus = *row.Status
				}
				responses[i].LinkedPO = &types.LinkedPOSummary{
					ID:             row.ID,
					DocumentNumber: row.DocumentNumber,
					Status:         poStatus,
				}
			}
		}
	}

	pagination := utils.CalculatePagination(page, pageSize, total)
	return utils.SendSuccess(c, fiber.StatusOK, responses, "Requisitions retrieved successfully", pagination)
}

// CreateRequisition creates a new requisition
func CreateRequisition(c *fiber.Ctx) error {
	var req types.CreateRequisitionRequest

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	// Validate required fields
	if req.Title == "" || len(req.Title) < 3 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Title is required and must be at least 3 characters",
		})
	}
	if req.Description == "" || len(req.Description) < 10 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Description is required and must be at least 10 characters",
		})
	}
	if len(req.Items) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "At least one item is required",
		})
	}
	// Validate items have positive quantities and valid descriptions
	for _, item := range req.Items {
		if item.Description == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "All items must have descriptions",
			})
		}
		if item.Quantity <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "All items must have positive quantities",
			})
		}
		if item.UnitPrice <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "All items must have positive unit prices",
			})
		}
	}
	if req.TotalAmount <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Total amount must be greater than 0",
		})
	}

	// Get authenticated user
	userID := c.Locals("userID").(string)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "User ID not found in token",
		})
	}

	// Get organization ID from context (set by auth middleware)
	organizationID := c.Locals("organizationID").(string)
	if organizationID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Organization ID not found in token",
		})
	}

	// Get user details
	var user models.User
	if err := config.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "User not found",
		})
	}

	// Validate CategoryID if provided — capture category name for denormalized storage
	resolvedCategoryName := ""
	if req.CategoryID != nil && *req.CategoryID != "" {
		var category models.Category
		if err := config.DB.Where("id = ? AND organization_id = ?", *req.CategoryID, organizationID).First(&category).Error; err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Category not found in your organization",
			})
		}
		resolvedCategoryName = category.Name
	}

	// Validate PreferredVendorID if provided
	if req.PreferredVendorID != nil && *req.PreferredVendorID != "" {
		var vendor models.Vendor
		if err := config.DB.Where("id = ? AND organization_id = ?", *req.PreferredVendorID, organizationID).First(&vendor).Error; err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Preferred vendor not found in your organization",
			})
		}
	}

	// Create requisition
	documentNumber := utils.GenerateDocumentNumber("REQ")

	// Prepare metadata — start with any incoming metadata, then overlay known fields
	metadataMap := map[string]interface{}{}
	for k, v := range req.Metadata {
		metadataMap[k] = v
	}
	if req.RequestedFor != "" {
		metadataMap["requestedFor"] = req.RequestedFor
	}
	if req.OtherCategoryText != "" {
		metadataMap["otherCategoryText"] = req.OtherCategoryText
	}
	// Store the resolved category name so the PDF can display it even if the
	// category record is later deleted or the preload fails.
	if resolvedCategoryName != "" {
		metadataMap["categoryName"] = resolvedCategoryName
	}

	metadataBytes, _ := json.Marshal(metadataMap)
	metadata := datatypes.JSON(metadataBytes)

	requisition := models.Requisition{
		ID:                uuid.New().String(),
		OrganizationID:    organizationID, // Add organization ID
		DocumentNumber:    documentNumber,
		RequesterId:       userID,
		RequesterName:     user.Name, // Stored in created_by_name — fallback for when Requester preload fails
		Title:             req.Title,
		Description:       req.Description,
		Department:        req.Department,
		DepartmentId:      req.DepartmentId,
		Status:            models.StatusDraft,
		Priority:          req.Priority,
		TotalAmount:       req.TotalAmount,
		Currency:          req.Currency,
		CategoryID:        req.CategoryID,
		PreferredVendorID:   req.PreferredVendorID,
		PreferredVendorName: req.PreferredVendorName, // persist manual ad-hoc vendor name
		IsEstimate:          req.IsEstimate,
		ApprovalStage:     0,

		// Business requirement fields
		BudgetCode:      req.BudgetCode,
		SourceOfFunds:   req.SourceOfFunds,
		CostCenter:      req.CostCenter,
		ProjectCode:     req.ProjectCode,
		RequiredByDate:  req.RequiredByDate,
		CreatedBy:       userID,    // From token
		CreatedByName:   user.Name, // From authenticated user
		CreatedByRole:   user.Role, // From authenticated user
		RequestedBy:     userID,
		RequestedByName: user.Name,
		RequestedByRole: user.Role,
		RequestedDate:   time.Now(),
		Metadata:        metadata,

		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	requisition.Items = datatypes.NewJSONType(req.Items)

	// Initialize empty approval history
	requisition.ApprovalHistory = datatypes.NewJSONType([]types.ApprovalRecord{})

	// Initialize action history with creation entry
	actionHistory := []types.ActionHistoryEntry{
		{
			ID:              uuid.New().String(),
			Action:          "CREATE",
			PerformedBy:     userID,
			PerformedByName: user.Name,
			PerformedByRole: user.Role,
			Timestamp:       time.Now(),
			Comments:        "Requisition created",
			ActionType:      "CREATE",
			NewStatus:       models.StatusDraft,
		},
	}
	requisition.ActionHistory = datatypes.NewJSONType(actionHistory)

	// Save to database
	if err := config.DB.Create(&requisition).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to create requisition",
			"error":   err.Error(),
		})
	}

	// Preload requester, category, and vendor
	config.DB.Preload("Requester").Preload("Category").Preload("PreferredVendor").First(&requisition)

	go utils.SyncDocumentAs(config.DB, "REQUISITION", requisition.ID, userID)
	go services.LogDocumentEvent(config.DB, services.DocumentEvent{
		OrganizationID: organizationID,
		DocumentID:     requisition.ID,
		DocumentType:   "requisition",
		UserID:         userID,
		ActorName:      user.Name,
		ActorRole:      user.Role,
		Action:         "created",
		Details:        map[string]interface{}{"documentNumber": requisition.DocumentNumber, "title": requisition.Title},
	})

	return c.Status(fiber.StatusCreated).JSON(types.DetailResponse{
		Success: true,
		Data:    modelToRequisitionResponse(requisition),
	})
}

// GetRequisition retrieves a single requisition by ID
func GetRequisition(c *fiber.Ctx) error {
	// Set cache control headers to ensure fresh data for PDF generation
	c.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Set("Pragma", "no-cache")
	c.Set("Expires", "0")

	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Requisition ID is required",
		})
	}

	// Get organization ID + user context for access scoping
	organizationID := c.Locals("organizationID").(string)
	userID, _ := c.Locals("userID").(string)
	userRole, _ := c.Locals("userRole").(string)

	// Org + role/ownership scope. Detail now mirrors list visibility: a user
	// who shouldn't see the requisition in the list can't reach it via the
	// UUID either. Requester + assigned approvers + privileged/procurement
	// all still see their expected rows via ApplyToQuery.
	scope := utils.GetDocumentScope(config.DB, userID, userRole, organizationID)
	query := config.DB.
		Preload("Requester").
		Preload("Category").
		Preload("PreferredVendor").
		Where("organization_id = ? AND (id = ? OR document_number = ?)", organizationID, id, id)
	query = scope.ApplyToQuery(query, "requester_id", "requisition", "")

	var requisition models.Requisition
	if err := query.First(&requisition).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Requisition not found",
		})
	}

	response := modelToRequisitionResponse(requisition)
	if liveHistory := utils.GetDocumentApprovalHistory(config.DB, requisition.ID, "requisition"); len(liveHistory) > 0 {
		response.ApprovalHistory = liveHistory
	}
	return c.JSON(types.DetailResponse{
		Success: true,
		Data:    response,
	})
}

// UpdateRequisition updates an existing requisition
func UpdateRequisition(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Requisition ID is required",
		})
	}

	var req types.UpdateRequisitionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	// Get organization ID from context
	organizationID := c.Locals("organizationID").(string)

	// Get existing requisition - SECURITY FIX: filter by organization_id
	var requisition models.Requisition
	if err := config.DB.Where("id = ? AND organization_id = ?", id, organizationID).First(&requisition).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Requisition not found",
		})
	}

	// Check if requisition is in a state that allows editing
	// Exception: metadata-only updates (e.g. adding quotations) are always allowed
	statusUpper := strings.ToUpper(requisition.Status)
	isMetadataOnly := req.Metadata != nil && req.Title == "" && req.Description == "" &&
		req.Department == "" && req.Priority == "" && len(req.Items) == 0 &&
		req.TotalAmount == 0 && req.Currency == ""
	if statusUpper != "DRAFT" && statusUpper != "PENDING" && !isMetadataOnly {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": fmt.Sprintf("Cannot update requisition in %s status", requisition.Status),
		})
	}

	// Update fields (only if provided)
	if req.Title != "" {
		requisition.Title = req.Title
	}
	if req.Description != "" {
		requisition.Description = req.Description
	}
	if req.Department != "" {
		requisition.Department = req.Department
	}
	if req.Priority != "" {
		requisition.Priority = req.Priority
	}
	if len(req.Items) > 0 {
		requisition.Items = datatypes.NewJSONType(req.Items)
	}
	if req.TotalAmount > 0 {
		requisition.TotalAmount = req.TotalAmount
	}
	if req.Currency != "" {
		requisition.Currency = req.Currency
	}
	if req.CategoryID != nil {
		// Validate category if provided
		if *req.CategoryID != "" {
			var category models.Category
			if err := config.DB.Where("id = ? AND organization_id = ?", *req.CategoryID, organizationID).First(&category).Error; err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"message": "Category not found",
				})
			}
		}
		requisition.CategoryID = req.CategoryID
	}
	if req.PreferredVendorID != nil {
		// Validate vendor if provided
		if *req.PreferredVendorID != "" {
			var vendor models.Vendor
			if err := config.DB.Where("id = ? AND organization_id = ?", *req.PreferredVendorID, organizationID).First(&vendor).Error; err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"message": "Preferred vendor not found",
				})
			}
		}
		requisition.PreferredVendorID = req.PreferredVendorID
	}
	// Persist preferred vendor name when supplied (covers manual ad-hoc vendors).
	if req.PreferredVendorName != "" || (req.PreferredVendorID != nil && *req.PreferredVendorID == "") {
		requisition.PreferredVendorName = req.PreferredVendorName
	}
	if req.IsEstimate != nil {
		requisition.IsEstimate = *req.IsEstimate
	}
	if req.SourceOfFunds != "" {
		requisition.SourceOfFunds = req.SourceOfFunds
	}

	// Merge incoming metadata with existing metadata
	if req.Metadata != nil {
		existingMeta := map[string]interface{}{}
		if len(requisition.Metadata) > 0 {
			json.Unmarshal(requisition.Metadata, &existingMeta)
		}
		for k, v := range req.Metadata {
			existingMeta[k] = v
		}
		metadataBytes, _ := json.Marshal(existingMeta)
		requisition.Metadata = datatypes.JSON(metadataBytes)
	}

	// Add action history entry for update
	var actionHistory []types.ActionHistoryEntry
	actionHistory = requisition.ActionHistory.Data()

	// Get user info for action history
	userID := c.Locals("userID").(string)
	var user models.User
	if err := config.DB.Where("id = ?", userID).First(&user).Error; err == nil {
		actionHistory = append(actionHistory, types.ActionHistoryEntry{
			ID:              uuid.New().String(),
			Action:          "UPDATE",
			PerformedBy:     userID,
			PerformedByName: user.Name,
			PerformedByRole: user.Role,
			Timestamp:       time.Now(),
			Comments:        "Requisition updated",
			ActionType:      "UPDATE",
			PreviousStatus:  requisition.Status,
			NewStatus:       requisition.Status,
		})
		requisition.ActionHistory = datatypes.NewJSONType(actionHistory)
	}

	requisition.UpdatedAt = time.Now()

	// Save changes
	if err := config.DB.Save(&requisition).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to update requisition",
			"error":   err.Error(),
		})
	}

	// Preload requester, category, and vendor
	config.DB.Preload("Requester").Preload("Category").Preload("PreferredVendor").First(&requisition)

	// Sync quotations to linked PO (if metadata["quotations"] changed)
	if req.Metadata != nil {
		if _, hasQuotations := req.Metadata["quotations"]; hasQuotations {
			go func(reqID string, orgID string, newMeta []byte) {
				var po models.PurchaseOrder
				if err := config.DB.Where("source_requisition_id = ? AND organization_id = ?", reqID, orgID).
					First(&po).Error; err != nil {
					return // no linked PO
				}
				poMeta := map[string]interface{}{}
				if len(po.Metadata) > 0 {
					json.Unmarshal(po.Metadata, &poMeta)
				}
				var reqMeta map[string]interface{}
				if err := json.Unmarshal(newMeta, &reqMeta); err == nil {
					if q, ok := reqMeta["quotations"]; ok {
						poMeta["quotations"] = q
					}
				}
				if metaBytes, err := json.Marshal(poMeta); err == nil {
					po.Metadata = datatypes.JSON(metaBytes)
					config.DB.Save(&po)
				}
			}(requisition.ID, organizationID, requisition.Metadata)
		}
	}

	actorID, _ := c.Locals("userID").(string)
	actorRole, _ := c.Locals("userRole").(string)
	var reqUpdateUser models.User
	config.DB.Where("id = ?", actorID).First(&reqUpdateUser)

	go utils.SyncDocumentAs(config.DB, "REQUISITION", requisition.ID, actorID)
	go services.LogDocumentEvent(config.DB, services.DocumentEvent{
		OrganizationID: organizationID,
		DocumentID:     requisition.ID,
		DocumentType:   "requisition",
		UserID:         actorID,
		ActorName:      reqUpdateUser.Name,
		ActorRole:      actorRole,
		Action:         "updated",
		Details:        map[string]interface{}{"documentNumber": requisition.DocumentNumber},
	})

	return c.JSON(types.DetailResponse{
		Success: true,
		Data:    modelToRequisitionResponse(requisition),
	})
}

// DeleteRequisition deletes a requisition (soft delete)
func DeleteRequisition(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Requisition ID is required",
		})
	}

	// Get organization ID from context
	organizationID := c.Locals("organizationID").(string)

	var requisition models.Requisition
	if err := config.DB.Where("id = ? AND organization_id = ?", id, organizationID).First(&requisition).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Requisition not found",
		})
	}

	// Only allow deletion of draft requisitions
	if strings.ToUpper(requisition.Status) != "DRAFT" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Only draft requisitions can be deleted",
		})
	}

	// Hard delete
	if err := config.DB.Delete(&requisition).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to delete requisition",
			"error":   err.Error(),
		})
	}

	return c.JSON(types.MessageResponse{
		Success: true,
		Message: "Requisition deleted successfully",
	})
}

// ReassignRequisition reassigns a requisition to a different approver
func ReassignRequisition(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Requisition ID is required",
		})
	}

	var req types.ReassignDocumentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	if req.NewApproverID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "New approver ID is required",
		})
	}

	// Verify new approver exists
	var newApprover models.User
	if err := config.DB.Where("id = ?", req.NewApproverID).First(&newApprover).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "New approver not found",
		})
	}

	// Get organization ID from context
	organizationID := c.Locals("organizationID").(string)

	// Get existing requisition - SECURITY FIX: filter by organization_id
	var requisition models.Requisition
	if err := config.DB.Where("id = ? AND organization_id = ?", id, organizationID).First(&requisition).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Requisition not found",
		})
	}

	requisition.UpdatedAt = time.Now()

	if err := config.DB.Save(&requisition).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to reassign requisition",
			"error":   err.Error(),
		})
	}

	// Preload requester
	config.DB.Preload("Requester").First(&requisition)

	return c.JSON(types.DetailResponse{
		Success: true,
		Data:    modelToRequisitionResponse(requisition),
	})
}

// Helper function to convert model to response
func modelToRequisitionResponse(req models.Requisition) types.RequisitionResponse {
	var items []types.RequisitionItem
	items = req.Items.Data()

	var approvalHistory []types.ApprovalRecord
	approvalHistory = req.ApprovalHistory.Data()

	// Use preloaded name when available; fall back to the denormalized DB column
	// (created_by_name) which is set at creation time and survives user deletion.
	requesterName := req.RequesterName
	if req.Requester != nil && req.Requester.Name != "" {
		requesterName = req.Requester.Name
	}

	// Use preloaded category name; fall back to the name stored in metadata at
	// creation time (survives category deletion).
	categoryName := ""
	if req.Category != nil {
		categoryName = req.Category.Name
	}

	preferredVendorName := req.PreferredVendorName // stored fallback (covers ad-hoc vendors)
	var preferredVendorResp *types.VendorResponse
	if req.PreferredVendor != nil {
		preferredVendorName = req.PreferredVendor.Name // canonical vendor record wins
		vr := modelToVendorResponse(*req.PreferredVendor)
		preferredVendorResp = &vr
	}

	// Extract metadata fields
	var requestedFor, otherCategoryText string
	var metadataMap map[string]interface{}
	if len(req.Metadata) > 0 {
		if err := json.Unmarshal(req.Metadata, &metadataMap); err == nil {
			if val, ok := metadataMap["requestedFor"].(string); ok {
				requestedFor = val
			}
			if val, ok := metadataMap["otherCategoryText"].(string); ok {
				otherCategoryText = val
			}
			// Fall back to the category name stored at creation time when the
			// Category record can no longer be preloaded (e.g. category deleted).
			if categoryName == "" {
				if val, ok := metadataMap["categoryName"].(string); ok {
					categoryName = val
				}
			}
		}
	}

	// Get action history
	var actionHistory []types.ActionHistoryEntry
	actionHistory = req.ActionHistory.Data()

	return types.RequisitionResponse{
		ID:                  req.ID,
		DocumentNumber:      req.DocumentNumber,
		RequesterID:         req.RequesterId,
		RequesterName:       requesterName,
		Title:               req.Title,
		Description:         req.Description,
		Department:          req.Department,
		Status:              req.Status,
		Priority:            req.Priority,
		Items:               items,
		TotalAmount:         req.TotalAmount,
		Currency:            req.Currency,
		CategoryID:          req.CategoryID,
		CategoryName:        categoryName,
		PreferredVendorID:   req.PreferredVendorID,
		PreferredVendorName: preferredVendorName,
		PreferredVendor:     preferredVendorResp,
		IsEstimate:          req.IsEstimate,
		ApprovalStage:       req.ApprovalStage,
		ApprovalHistory:     approvalHistory,

		// Business requirement fields
		BudgetCode:        req.BudgetCode,
		CostCenter:        req.CostCenter,
		ProjectCode:       req.ProjectCode,
		RequiredByDate:    req.RequiredByDate,
		RequestedFor:      requestedFor,
		OtherCategoryText: otherCategoryText,

		// Full metadata for frontend (e.g. attachments)
		Metadata: metadataMap,

		// Action history for frontend
		ActionHistory: actionHistory,

		CreatedAt: req.CreatedAt,
		UpdatedAt: req.UpdatedAt,
	}
}

// WithdrawRequisition allows the requester to withdraw a submitted requisition
// The requisition must be in pending status and not claimed by any approver
func WithdrawRequisition(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Requisition ID is required",
		})
	}

	// Get organization ID and user ID from context
	organizationID := c.Locals("organizationID").(string)
	userID := c.Locals("userID").(string)

	// Get existing requisition
	var requisition models.Requisition
	if err := config.DB.Where("id = ? AND organization_id = ?", id, organizationID).First(&requisition).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Requisition not found",
		})
	}

	// Verify that the current user is the requester (only the submitter can withdraw)
	if requisition.RequesterId != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Only the requester can withdraw this requisition",
		})
	}

	// Check if requisition is in a state that can be withdrawn (pending)
	if strings.ToUpper(requisition.Status) != "PENDING" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": fmt.Sprintf("Cannot withdraw requisition in %s status. Only pending requisitions can be withdrawn.", requisition.Status),
		})
	}

	// Check if there is an active workflow task that is claimed
	var workflowTask models.WorkflowTask
	err := config.DB.Where("entity_id = ? AND entity_type = ? AND UPPER(status) IN (?, ?)",
		id, "requisition", "PENDING", "CLAIMED").First(&workflowTask).Error

	if err == nil {
		// Task exists - check if it's claimed
		if strings.ToUpper(workflowTask.Status) == "CLAIMED" && workflowTask.ClaimedBy != nil {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"success": false,
				"message": "Cannot withdraw requisition. It is currently being reviewed by an approver.",
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

	// Delete the workflow task(s) for this requisition
	if err := tx.Where("entity_id = ? AND entity_type = ?", id, "requisition").
		Delete(&models.WorkflowTask{}).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to remove workflow tasks",
			"error":   err.Error(),
		})
	}

	// Delete the workflow assignment(s) for this requisition
	if err := tx.Where("entity_id = ? AND entity_type = ?", id, "requisition").
		Delete(&models.WorkflowAssignment{}).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to remove workflow assignments",
			"error":   err.Error(),
		})
	}

	// Update requisition status back to draft and reset approval fields
	previousStatus := requisition.Status
	requisition.Status = models.StatusDraft
	requisition.ApprovalStage = 0
	requisition.UpdatedAt = time.Now()

	// Clear approval history since we're reverting to draft
	requisition.ApprovalHistory = datatypes.NewJSONType([]types.ApprovalRecord{})

	// Add action history entry for withdrawal
	var actionHistory []types.ActionHistoryEntry
	actionHistory = requisition.ActionHistory.Data()
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
		Comments:        "Requisition withdrawn by requester",
		ActionType:      "WITHDRAW",
		PreviousStatus:  previousStatus,
		NewStatus:       models.StatusDraft,
	})
	requisition.ActionHistory = datatypes.NewJSONType(actionHistory)

	// Save requisition changes
	if err := tx.Save(&requisition).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to update requisition status",
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

	// Preload requester, category, and vendor for response
	config.DB.Preload("Requester").Preload("Category").Preload("PreferredVendor").First(&requisition)

	go services.LogDocumentEvent(config.DB, services.DocumentEvent{
		OrganizationID: organizationID,
		DocumentID:     requisition.ID,
		DocumentType:   "requisition",
		UserID:         userID,
		ActorName:      user.Name,
		ActorRole:      user.Role,
		Action:         "withdrawn",
		Details:        map[string]interface{}{"documentNumber": requisition.DocumentNumber},
	})

	return c.JSON(fiber.Map{
		"success": true,
		"data":    modelToRequisitionResponse(requisition),
		"message": "Requisition withdrawn successfully. You can now edit and re-submit it.",
	})
}

// SubmitRequisition submits a requisition for approval using the workflow system
func SubmitRequisition(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Requisition ID is required",
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

	// Get existing requisition
	var requisition models.Requisition
	if err := config.DB.Where("id = ? AND organization_id = ?", id, organizationID).First(&requisition).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Requisition not found",
		})
	}

	// Check if requisition is in draft status
	if strings.ToUpper(requisition.Status) != "DRAFT" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": fmt.Sprintf("Cannot submit requisition in %s status", requisition.Status),
		})
	}

	// Get workflow execution service from handler registry
	// This will be passed from the route handler
	workflowExecutionService := c.Locals("workflowExecutionService").(*services.WorkflowExecutionService)

	// Use routing-aware submission that handles both procurement and accounting paths
	routingResult, err := workflowExecutionService.SubmitRequisitionWithRouting(
		c.Context(), organizationID, requisition.ID, submitReq.WorkflowID, userID, &requisition,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to submit requisition",
			"error":   err.Error(),
		})
	}

	// If auto-approved, the requisition status was already updated by the routing service.
	// Otherwise, update to "pending" and add action history.
	if !routingResult.AutoApproved {
		requisition.Status = models.StatusPending
		requisition.UpdatedAt = time.Now()

		// Add action history entry for submission
		var actionHistory []types.ActionHistoryEntry
		actionHistory = requisition.ActionHistory.Data()
		if actionHistory == nil {
			actionHistory = []types.ActionHistoryEntry{}
		}

		// Get user info for action history
		performerName := "Unknown User"
		performerRole := "unknown"
		var user models.User
		if err := config.DB.Where("id = ?", userID).First(&user).Error; err == nil {
			performerName = user.Name
			performerRole = user.Role
		}

		actionHistory = append(actionHistory, types.ActionHistoryEntry{
			ID:              uuid.New().String(),
			Action:          "SUBMIT",
			PerformedBy:     userID,
			PerformedByName: performerName,
			PerformedByRole: performerRole,
			Timestamp:       time.Now(),
			Comments:        "Requisition submitted for approval",
			ActionType:      "SUBMIT",
			PreviousStatus:  models.StatusDraft,
			NewStatus:       models.StatusPending,
		})
		requisition.ActionHistory = datatypes.NewJSONType(actionHistory)

		// Save requisition
		if err := config.DB.Save(&requisition).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to update requisition status",
				"error":   err.Error(),
			})
		}
	} else {
		// Reload the auto-approved requisition to get the latest state
		if err := config.DB.Where("id = ? AND organization_id = ?", requisition.ID, organizationID).First(&requisition).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to reload requisition after auto-approval",
				"error":   err.Error(),
			})
		}
	}

	// Preload requester, category, and vendor
	config.DB.Preload("Requester").Preload("Category").Preload("PreferredVendor").First(&requisition)

	go utils.SyncDocumentAs(config.DB, "REQUISITION", requisition.ID, userID)
	go services.LogDocumentEvent(config.DB, services.DocumentEvent{
		OrganizationID: organizationID,
		DocumentID:     requisition.ID,
		DocumentType:   "requisition",
		UserID:         userID,
		Action:         "submitted",
		Details:        map[string]interface{}{"routingPath": routingResult.RoutingPath, "autoApproved": routingResult.AutoApproved},
	})

	// Build response with routing information
	responseData := fiber.Map{
		"requisition": modelToRequisitionResponse(requisition),
		"routing": fiber.Map{
			"path":         routingResult.RoutingPath,
			"autoApproved": routingResult.AutoApproved,
		},
	}

	// Include workflow assignment info if available
	if routingResult.Assignment != nil {
		responseData["workflow"] = fiber.Map{
			"assignmentId": routingResult.Assignment.ID,
			"workflowId":   routingResult.Assignment.WorkflowID,
			"currentStage": routingResult.Assignment.CurrentStage,
			"status":       routingResult.Assignment.Status,
		}
	}

	// Include auto-created PO info if available
	if routingResult.AutoCreatedPO != nil && routingResult.AutoCreatedPO.Success {
		poID := routingResult.AutoCreatedPO.DocumentID
		if routingResult.AutoCreatedPOID != "" {
			poID = routingResult.AutoCreatedPOID
		}
		responseData["autoCreatedPO"] = fiber.Map{
			"id": poID,
		}
	}

	return c.JSON(types.DetailResponse{
		Success: true,
		Data:    responseData,
	})
}

// GetRequisitionChain returns the full document chain for a requisition
// GET /api/v1/:orgId/requisitions/:id/chain
func GetRequisitionChain(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Requisition ID is required",
		})
	}

	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Organization context required",
		})
	}
	orgID := tenant.OrganizationID

	// Verify requisition exists and belongs to org
	var req models.Requisition
	if err := config.DB.Where("id = ? AND organization_id = ?", id, orgID).First(&req).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Requisition not found",
		})
	}

	// Build chain using document linking service
	dls := services.NewDocumentLinkingService(config.DB)
	rawChain, err := dls.GetDocumentRelationshipChain(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to retrieve document chain",
			"error":   err.Error(),
		})
	}

	chain := fiber.Map{
		"requisitionId":     id,
		"requisitionStatus": req.Status,
	}

	// Fetch PO status if PO exists
	if poID, ok := rawChain["poId"].(string); ok && poID != "" {
		chain["poId"] = poID
		chain["poDocumentNumber"] = rawChain["poDocumentNumber"]
		var po models.PurchaseOrder
		if err := config.DB.Where("id = ? AND organization_id = ?", poID, orgID).First(&po).Error; err == nil {
			chain["poStatus"] = po.Status
		}

		// Look up PV linked to this PO's document number
		if poDocNum, ok := rawChain["poDocumentNumber"].(string); ok && poDocNum != "" {
			var pv models.PaymentVoucher
			if err := config.DB.Where("linked_po = ? AND organization_id = ?", poDocNum, orgID).First(&pv).Error; err == nil {
				chain["pvId"] = pv.ID
				chain["pvDocumentNumber"] = pv.DocumentNumber
				chain["pvStatus"] = pv.Status
			}
		}
	}

	// Fetch GRN status if GRN exists
	if grnID, ok := rawChain["grnId"].(string); ok && grnID != "" {
		chain["grnId"] = grnID
		chain["grnDocumentNumber"] = rawChain["grnDocumentNumber"]
		var grn models.GoodsReceivedNote
		if err := config.DB.Where("id = ? AND organization_id = ?", grnID, orgID).First(&grn).Error; err == nil {
			chain["grnStatus"] = grn.Status
		}
	}

	// Detect routing type from workflow assignment
	routingType := "procurement"
	var wa models.WorkflowAssignment
	if err := config.DB.Preload("Workflow").
		Where("entity_id = ? AND entity_type = ? AND organization_id = ?", id, "requisition", orgID).
		First(&wa).Error; err == nil && wa.Workflow != nil {
		var wfConditions models.WorkflowConditions
		if jsonErr := json.Unmarshal(wa.Workflow.Conditions, &wfConditions); jsonErr == nil {
			if strings.EqualFold(wfConditions.RoutingType, "accounting") {
				routingType = "accounting"
			}
		}
	}
	chain["routingType"] = routingType

	return c.JSON(fiber.Map{
		"success": true,
		"data":    chain,
	})
}

// GetRequisitionAuditTrail returns merged audit logs across all documents in the chain
// GET /api/v1/:orgId/requisitions/:id/audit-trail
// Access: admin, super_admin, manager, finance only
func GetRequisitionAuditTrail(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Requisition ID is required",
		})
	}

	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Organization context required",
		})
	}
	orgID := tenant.OrganizationID

	// Enforce role restriction
	allowedRoles := []string{"admin", "super_admin", "manager", "finance"}
	callerRole := strings.ToLower(tenant.UserRole)
	allowed := false
	for _, r := range allowedRoles {
		if callerRole == r {
			allowed = true
			break
		}
	}
	if !allowed {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Access restricted to admin, manager, and finance roles",
		})
	}

	// Verify requisition exists and belongs to org
	var req models.Requisition
	if err := config.DB.Where("id = ? AND organization_id = ?", id, orgID).First(&req).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Requisition not found",
		})
	}

	// Get document chain to collect all related doc IDs
	dls := services.NewDocumentLinkingService(config.DB)
	rawChain, err := dls.GetDocumentRelationshipChain(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to retrieve document chain",
			"error":   err.Error(),
		})
	}

	docIDs := []string{id}
	docLabels := map[string]string{id: "Requisition"}

	if poID, ok := rawChain["poId"].(string); ok && poID != "" {
		docIDs = append(docIDs, poID)
		docLabels[poID] = "Purchase Order"

		// Also look up PV
		if poDocNum, ok := rawChain["poDocumentNumber"].(string); ok && poDocNum != "" {
			var pv models.PaymentVoucher
			if err := config.DB.Where("linked_po = ? AND organization_id = ?", poDocNum, orgID).First(&pv).Error; err == nil {
				docIDs = append(docIDs, pv.ID)
				docLabels[pv.ID] = "Payment Voucher"
			}
		}
	}
	if grnID, ok := rawChain["grnId"].(string); ok && grnID != "" {
		docIDs = append(docIDs, grnID)
		docLabels[grnID] = "Goods Received Note"
	}

	// Fetch all audit logs for the collected doc IDs
	var auditLogs []models.AuditLog
	if err := config.DB.
		Where("document_id IN ?", docIDs).
		Order("created_at ASC").
		Find(&auditLogs).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to fetch audit logs",
			"error":   err.Error(),
		})
	}

	responses := make([]map[string]interface{}, 0, len(auditLogs))
	for _, al := range auditLogs {
		responses = append(responses, map[string]interface{}{
			"id":           al.ID,
			"documentId":   al.DocumentID,
			"documentType": al.DocumentType,
			"documentLabel": docLabels[al.DocumentID],
			"userId":       al.UserID,
			"action":       al.Action,
			"changes":      al.Changes,
			"createdAt":    al.CreatedAt,
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    responses,
	})
}
