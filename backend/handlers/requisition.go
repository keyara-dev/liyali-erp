package handlers

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/logging"
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

	db := config.DB

	// Extract pagination parameters
	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("page_size", 10)

	// Extract filter parameters
	status := c.Query("status")
	department := c.Query("department")
	priority := c.Query("priority")

	// Add query parameters to context
	logging.AddFieldsToRequest(c, map[string]interface{}{
		"operation":  "get_requisitions",
		"page":       page,
		"page_size":  pageSize,
		"status":     status,
		"department": department,
		"priority":   priority,
	})

	// Build query
	query := db
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if department != "" {
		query = query.Where("department = ?", department)
	}
	if priority != "" {
		query = query.Where("priority = ?", priority)
	}

	// Get total count
	var total int64
	if err := query.Model(&models.Requisition{}).Count(&total).Error; err != nil {
		return utils.SendInternalError(c, "Failed to count requisitions", err)
	}

	// Fetch paginated results
	var requisitions []models.Requisition
	offset := (page - 1) * pageSize
	if err := query.
		Offset(offset).
		Limit(pageSize).
		Preload("Requester").
		Preload("Category").
		Preload("PreferredVendor").
		Order("created_at DESC").
		Find(&requisitions).Error; err != nil {
		return utils.SendInternalError(c, "Failed to fetch requisitions", err)
	}

	// Convert to response format
	responses := make([]types.RequisitionResponse, 0, len(requisitions))
	for _, req := range requisitions {
		responses = append(responses, modelToRequisitionResponse(req))
	}

	// Calculate pagination
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

	// Get user details
	var user models.User
	if err := config.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "User not found",
		})
	}

	// Validate CategoryID if provided
	if req.CategoryID != nil && *req.CategoryID != "" {
		var category models.Category
		if err := config.DB.Where("id = ?", *req.CategoryID).First(&category).Error; err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Category not found",
			})
		}
	}

	// Validate PreferredVendorID if provided
	if req.PreferredVendorID != nil && *req.PreferredVendorID != "" {
		var vendor models.Vendor
		if err := config.DB.Where("id = ?", *req.PreferredVendorID).First(&vendor).Error; err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Preferred vendor not found",
			})
		}
	}

	// Create requisition
	reqNumber := utils.GenerateRequisitionNumber()
	
	// Prepare metadata for additional fields
	metadataMap := map[string]interface{}{}
	if req.RequestedFor != "" {
		metadataMap["requestedFor"] = req.RequestedFor
	}
	if req.OtherCategoryText != "" {
		metadataMap["otherCategoryText"] = req.OtherCategoryText
	}
	
	metadataBytes, _ := json.Marshal(metadataMap)
	metadata := datatypes.JSON(metadataBytes)
	
	requisition := models.Requisition{
		ID:                uuid.New().String(),
		REQNumber:         reqNumber,
		RequesterId:       userID,
		Title:             req.Title,
		Description:       req.Description,
		Department:        req.Department,
		DepartmentId:      req.DepartmentId,
		Status:            "draft",
		Priority:          req.Priority,
		TotalAmount:       req.TotalAmount,
		Currency:          req.Currency,
		CategoryID:        req.CategoryID,
		PreferredVendorID: req.PreferredVendorID,
		IsEstimate:        req.IsEstimate,
		ApprovalStage:     0,
		
		// Business requirement fields
		BudgetCode:        req.BudgetCode,
		CostCenter:        req.CostCenter,
		ProjectCode:       req.ProjectCode,
		RequiredByDate:    req.RequiredByDate,
		CreatedBy:         userID,           // From token
		CreatedByName:     user.Name,        // From authenticated user
		CreatedByRole:     user.Role,        // From authenticated user
		RequestedBy:       userID,
		RequestedByName:   user.Name,
		RequestedByRole:   user.Role,
		RequisitionNumber: reqNumber,
		RequestedDate:     time.Now(),
		Metadata:          metadata,
		
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	requisition.Items = datatypes.NewJSONType(req.Items)

	// Initialize empty approval history
	requisition.ApprovalHistory = datatypes.NewJSONType([]types.ApprovalRecord{})

	// Initialize action history with creation entry
	actionHistory := []types.ActionHistoryEntry{
		{
			ID:               uuid.New().String(),
			Action:           "CREATE",
			PerformedBy:      userID,
			PerformedByName:  user.Name,
			PerformedByRole:  user.Role,
			Timestamp:        time.Now(),
			Comments:         "Requisition created",
			ActionType:       "CREATE",
			NewStatus:        "draft",
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

	return c.Status(fiber.StatusCreated).JSON(types.DetailResponse{
		Success: true,
		Data:    modelToRequisitionResponse(requisition),
	})
}

// GetRequisition retrieves a single requisition by ID
func GetRequisition(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Requisition ID is required",
		})
	}

	var requisition models.Requisition
	
	// Try to find by ID first (UUID), then by requisition number
	err := config.DB.
		Preload("Requester").
		Preload("Category").
		Preload("PreferredVendor").
		Where("id = ? OR req_number = ? OR requisition_number = ?", id, id, id).
		First(&requisition).Error
		
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Requisition not found",
		})
	}

	return c.JSON(types.DetailResponse{
		Success: true,
		Data:    modelToRequisitionResponse(requisition),
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

	// Get existing requisition
	var requisition models.Requisition
	if err := config.DB.Where("id = ?", id).First(&requisition).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Requisition not found",
		})
	}

	// Check if requisition is in a state that allows editing (draft or pending)
	if requisition.Status != "draft" && requisition.Status != "pending" {
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
			if err := config.DB.Where("id = ?", *req.CategoryID).First(&category).Error; err != nil {
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
			if err := config.DB.Where("id = ?", *req.PreferredVendorID).First(&vendor).Error; err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"message": "Preferred vendor not found",
				})
			}
		}
		requisition.PreferredVendorID = req.PreferredVendorID
	}
	if req.IsEstimate != nil {
		requisition.IsEstimate = *req.IsEstimate
	}

	// Add action history entry for update
	var actionHistory []types.ActionHistoryEntry
	actionHistory = requisition.ActionHistory.Data()
	
	// Get user info for action history
	userID := c.Locals("userID").(string)
	var user models.User
	if err := config.DB.Where("id = ?", userID).First(&user).Error; err == nil {
		actionHistory = append(actionHistory, types.ActionHistoryEntry{
			ID:               uuid.New().String(),
			Action:           "UPDATE",
			PerformedBy:      userID,
			PerformedByName:  user.Name,
			PerformedByRole:  user.Role,
			Timestamp:        time.Now(),
			Comments:         "Requisition updated",
			ActionType:       "UPDATE",
			PreviousStatus:   requisition.Status,
			NewStatus:        requisition.Status,
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

	var requisition models.Requisition
	if err := config.DB.Where("id = ?", id).First(&requisition).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Requisition not found",
		})
	}

	// Only allow deletion of draft requisitions
	if requisition.Status != "draft" {
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

// ApproveRequisition approves a requisition and optionally auto-creates PO
func ApproveRequisition(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Requisition ID is required",
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

	// Get existing requisition
	var requisition models.Requisition
	if err := config.DB.Where("id = ?", id).First(&requisition).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Requisition not found",
		})
	}

	// Get approver info
	approverID := c.Locals("userID").(string)
	var approver models.User
	if err := config.DB.Where("id = ?", approverID).First(&approver).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Approver not found",
		})
	}

	// Unmarshal existing approval history
	var approvalHistory []types.ApprovalRecord
	approvalHistory = requisition.ApprovalHistory.Data()

	// Add new approval record
	approvalRecord := types.ApprovalRecord{
		ApproverID:   approverID,
		ApproverName: approver.Name,
		Status:       "approved",
		Comments:     req.Comments,
		Signature:    req.Signature,
		ApprovedAt:   time.Now(),
	}
	approvalHistory = append(approvalHistory, approvalRecord)

	// Add action history entry
	var actionHistory []types.ActionHistoryEntry
	actionHistory = requisition.ActionHistory.Data()
	actionHistory = append(actionHistory, types.ActionHistoryEntry{
		ID:               uuid.New().String(),
		Action:           "APPROVE",
		PerformedBy:      approverID,
		PerformedByName:  approver.Name,
		PerformedByRole:  approver.Role,
		Timestamp:        time.Now(),
		Comments:         req.Comments,
		ActionType:       "APPROVE",
		PreviousStatus:   requisition.Status,
		NewStatus:        "approved",
	})

	// Update requisition
	requisition.Status = "approved"
	requisition.ApprovalStage++
	requisition.ApprovalHistory = datatypes.NewJSONType(approvalHistory)
	requisition.ActionHistory = datatypes.NewJSONType(actionHistory)
	requisition.UpdatedAt = time.Now()

	if err := config.DB.Save(&requisition).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to approve requisition",
			"error":   err.Error(),
		})
	}

	// Preload requester
	config.DB.Preload("Requester").First(&requisition)

	// Auto-create Purchase Order if enabled and prerequisites are met
	var autoCreatedPO *models.PurchaseOrder
	if requisition.Status == "approved" {
		// Initialize automation service
		auditService := &services.AuditService{}
		notificationService := &services.NotificationService{}
		automationService := services.NewDocumentAutomationService(
			config.DB, auditService, notificationService,
		)

		// Get automation config (could be from organization settings in the future)
		automationConfig := automationService.GetDefaultAutomationConfig()

		// Attempt to auto-create PO
		result, err := automationService.CreatePurchaseOrderFromRequisition(
			c.Context(), &requisition, automationConfig,
		)

		if err == nil && result.Success {
			if po, ok := result.CreatedDocument.(models.PurchaseOrder); ok {
				autoCreatedPO = &po
			}
		}
		// Note: We don't fail the approval if PO creation fails
		// The requisition is still approved, PO can be created manually
	}

	// Prepare response
	response := types.DetailResponse{
		Success: true,
		Data:    modelToRequisitionResponse(requisition),
	}

	// Add auto-created PO to response if available
	if autoCreatedPO != nil {
		// Convert PO to response format
		poResponse := types.PurchaseOrderResponse{
			ID:                autoCreatedPO.ID,
			PONumber:          autoCreatedPO.PONumber,
			VendorID:          autoCreatedPO.VendorID,
			Status:            autoCreatedPO.Status,
			TotalAmount:       autoCreatedPO.TotalAmount,
			Currency:          autoCreatedPO.Currency,
			DeliveryDate:      autoCreatedPO.DeliveryDate,
			ApprovalStage:     autoCreatedPO.ApprovalStage,
			LinkedRequisition: autoCreatedPO.LinkedRequisition,
			CreatedAt:         autoCreatedPO.CreatedAt,
			UpdatedAt:         autoCreatedPO.UpdatedAt,
		}

		// Add PO to response
		response.Data = fiber.Map{
			"requisition":      modelToRequisitionResponse(requisition),
			"autoCreatedPO":    poResponse,
			"automationUsed":   true,
		}
	}

	return c.JSON(response)
}

// RejectRequisition rejects a requisition
func RejectRequisition(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Requisition ID is required",
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

	// Get existing requisition
	var requisition models.Requisition
	if err := config.DB.Where("id = ?", id).First(&requisition).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Requisition not found",
		})
	}

	// Get approver info
	approverID := c.Locals("userID").(string)
	var approver models.User
	if err := config.DB.Where("id = ?", approverID).First(&approver).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Approver not found",
		})
	}

	// Unmarshal existing approval history
	var approvalHistory []types.ApprovalRecord
	approvalHistory = requisition.ApprovalHistory.Data()

	// Add new rejection record
	rejectionRecord := types.ApprovalRecord{
		ApproverID:   approverID,
		ApproverName: approver.Name,
		Status:       "rejected",
		Comments:     req.Remarks,
		Signature:    req.Signature,
		ApprovedAt:   time.Now(),
	}
	approvalHistory = append(approvalHistory, rejectionRecord)

	// Add action history entry
	var actionHistory []types.ActionHistoryEntry
	actionHistory = requisition.ActionHistory.Data()
	actionHistory = append(actionHistory, types.ActionHistoryEntry{
		ID:               uuid.New().String(),
		Action:           "REJECT",
		PerformedBy:      approverID,
		PerformedByName:  approver.Name,
		PerformedByRole:  approver.Role,
		Timestamp:        time.Now(),
		Comments:         req.Remarks,
		Remarks:          req.Remarks,
		ActionType:       "REJECT",
		PreviousStatus:   requisition.Status,
		NewStatus:        "rejected",
	})

	// Update requisition
	requisition.Status = "rejected"
	requisition.ApprovalHistory = datatypes.NewJSONType(approvalHistory)
	requisition.ActionHistory = datatypes.NewJSONType(actionHistory)
	requisition.UpdatedAt = time.Now()

	if err := config.DB.Save(&requisition).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to reject requisition",
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

	// Get existing requisition
	var requisition models.Requisition
	if err := config.DB.Where("id = ?", id).First(&requisition).Error; err != nil {
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

	requesterName := ""
	if req.Requester != nil {
		requesterName = req.Requester.Name
	}

	categoryName := ""
	if req.Category != nil {
		categoryName = req.Category.Name
	}

	preferredVendorName := ""
	if req.PreferredVendor != nil {
		preferredVendorName = req.PreferredVendor.Name
	}

	// Extract metadata fields
	var requestedFor, otherCategoryText string
	if len(req.Metadata) > 0 {
		var metadata map[string]interface{}
		if err := json.Unmarshal(req.Metadata, &metadata); err == nil {
			if val, ok := metadata["requestedFor"].(string); ok {
				requestedFor = val
			}
			if val, ok := metadata["otherCategoryText"].(string); ok {
				otherCategoryText = val
			}
		}
	}

	// Get action history
	var actionHistory []types.ActionHistoryEntry
	actionHistory = req.ActionHistory.Data()

	return types.RequisitionResponse{
		ID:                  req.ID,
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
		IsEstimate:          req.IsEstimate,
		ApprovalStage:       req.ApprovalStage,
		ApprovalHistory:     approvalHistory,
		
		// Business requirement fields
		BudgetCode:          req.BudgetCode,
		CostCenter:          req.CostCenter,
		ProjectCode:         req.ProjectCode,
		RequiredByDate:      req.RequiredByDate,
		RequestedFor:        requestedFor,
		OtherCategoryText:   otherCategoryText,
		
		// Action history for frontend
		ActionHistory:       actionHistory,
		
		CreatedAt:           req.CreatedAt,
		UpdatedAt:           req.UpdatedAt,
	}
}
