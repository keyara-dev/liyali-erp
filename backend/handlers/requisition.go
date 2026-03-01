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

	// Extract and normalize pagination parameters
	page, pageSize := utils.NormalizePaginationParams(
		c.QueryInt("page", 1),
		c.QueryInt("page_size", 10),
	)

	// Extract filter parameters
	status := c.Query("status")
	department := c.Query("department")
	priority := c.Query("priority")

	// Get organization ID from context (set by auth middleware)
	organizationID := c.Locals("organizationID").(string)

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

	// Build query with organization filter
	query := db.Where("organization_id = ?", organizationID)
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

	// Validate CategoryID if provided
	if req.CategoryID != nil && *req.CategoryID != "" {
		var category models.Category
		if err := config.DB.Where("id = ? AND organization_id = ?", *req.CategoryID, organizationID).First(&category).Error; err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Category not found in your organization",
			})
		}
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

	metadataBytes, _ := json.Marshal(metadataMap)
	metadata := datatypes.JSON(metadataBytes)

	requisition := models.Requisition{
		ID:                uuid.New().String(),
		OrganizationID:    organizationID, // Add organization ID
		DocumentNumber:    documentNumber,
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
		BudgetCode:      req.BudgetCode,
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
			NewStatus:       "draft",
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

	// Get organization ID from context (set by auth middleware)
	organizationID := c.Locals("organizationID").(string)

	var requisition models.Requisition

	// Try to find by ID first (UUID), then by requisition number, filtered by organization
	err := config.DB.
		Preload("Requester").
		Preload("Category").
		Preload("PreferredVendor").
		Where("organization_id = ? AND (id = ? OR document_number = ?)", organizationID, id, id).
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
	if req.IsEstimate != nil {
		requisition.IsEstimate = *req.IsEstimate
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
	var metadataMap map[string]interface{}
	if len(req.Metadata) > 0 {
		if err := json.Unmarshal(req.Metadata, &metadataMap); err == nil {
			if val, ok := metadataMap["requestedFor"].(string); ok {
				requestedFor = val
			}
			if val, ok := metadataMap["otherCategoryText"].(string); ok {
				otherCategoryText = val
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
	if requisition.Status != "pending" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": fmt.Sprintf("Cannot withdraw requisition in %s status. Only pending requisitions can be withdrawn.", requisition.Status),
		})
	}

	// Check if there is an active workflow task that is claimed
	var workflowTask models.WorkflowTask
	err := config.DB.Where("entity_id = ? AND entity_type = ? AND status IN (?, ?)",
		id, "requisition", "pending", "claimed").First(&workflowTask).Error

	if err == nil {
		// Task exists - check if it's claimed
		if workflowTask.Status == "claimed" && workflowTask.ClaimedBy != nil {
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
	requisition.Status = "draft"
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
		NewStatus:       "draft",
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
	if requisition.Status != "draft" {
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
		requisition.Status = "pending"
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
			PreviousStatus:  "draft",
			NewStatus:       "pending",
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
