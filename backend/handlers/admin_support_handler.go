package handlers

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/utils"
)

// ============================================================================
// Support — Documents (platform-wide view for super_admin)
// ============================================================================

// AdminGetSupportDocuments returns a platform-wide list of documents for support.
// Optional query params: org_id, user_id, type, status, search, page, limit
func AdminGetSupportDocuments(c *fiber.Ctx) error {
	db := config.DB

	query := db.Model(&models.Document{})

	if orgID := c.Query("org_id"); orgID != "" {
		query = query.Where("organization_id = ?", orgID)
	}
	if userID := c.Query("user_id"); userID != "" {
		query = query.Where("created_by = ?", userID)
	}
	if docType := c.Query("type"); docType != "" {
		query = query.Where("document_type = ?", docType)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if search := c.Query("search"); search != "" {
		pattern := "%" + search + "%"
		query = query.Where("title ILIKE ? OR document_number ILIKE ?", pattern, pattern)
	}

	var total int64
	query.Count(&total)

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit

	var docs []models.Document
	if err := query.
		Preload("Organization").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&docs).Error; err != nil {
		return utils.SendInternalError(c, "Failed to fetch documents", err)
	}

	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Documents retrieved successfully",
		"data":    docs,
		"pagination": fiber.Map{
			"total":       total,
			"page":        page,
			"page_size":   limit,
			"total_pages": totalPages,
			"has_next":    page < totalPages,
			"has_prev":    page > 1,
		},
	})
}

// AdminGetSupportDocument returns a single document by ID for support diagnosis.
func AdminGetSupportDocument(c *fiber.Ctx) error {
	db := config.DB
	docID := c.Params("id")

	var doc models.Document
	if err := db.
		Preload("Organization").
		Preload("Creator").
		Preload("Workflow").
		First(&doc, "id = ?", docID).Error; err != nil {
		return utils.SendNotFound(c, "Document not found")
	}

	return utils.SendSimpleSuccess(c, doc, "Document retrieved successfully")
}

// ============================================================================
// Support — Workflow Tasks (platform-wide view for super_admin)
// ============================================================================

// AdminGetSupportWorkflowTasks returns a platform-wide list of workflow tasks for support.
// Optional query params: org_id, entity_id, status, stalled, page, limit
func AdminGetSupportWorkflowTasks(c *fiber.Ctx) error {
	db := config.DB

	query := db.Model(&models.WorkflowTask{})

	if orgID := c.Query("org_id"); orgID != "" {
		query = query.Where("organization_id = ?", orgID)
	}
	if entityID := c.Query("entity_id"); entityID != "" {
		query = query.Where("entity_id = ?", entityID)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	// stalled = claimed tasks where claim_expiry has passed
	if c.Query("stalled") == "true" {
		query = query.Where("status = 'claimed' AND claimed_at < ?", time.Now().Add(-30*time.Minute))
	}

	var total int64
	query.Count(&total)

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit

	var tasks []models.WorkflowTask
	if err := query.
		Preload("Organization").
		Preload("AssignedUser").
		Preload("Claimer").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&tasks).Error; err != nil {
		return utils.SendInternalError(c, "Failed to fetch workflow tasks", err)
	}

	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Workflow tasks retrieved successfully",
		"data":    tasks,
		"pagination": fiber.Map{
			"total":       total,
			"page":        page,
			"page_size":   limit,
			"total_pages": totalPages,
			"has_next":    page < totalPages,
			"has_prev":    page > 1,
		},
	})
}

// AdminGetSupportWorkflowTask returns a single workflow task with full details.
func AdminGetSupportWorkflowTask(c *fiber.Ctx) error {
	db := config.DB
	taskID := c.Params("id")

	var task models.WorkflowTask
	if err := db.
		Preload("Organization").
		Preload("AssignedUser").
		Preload("Claimer").
		Preload("WorkflowAssignment").
		First(&task, "id = ?", taskID).Error; err != nil {
		return utils.SendNotFound(c, "Workflow task not found")
	}

	return utils.SendSimpleSuccess(c, task, "Workflow task retrieved successfully")
}

// AdminReassignWorkflowTask reassigns a stuck/claimed workflow task to a different user.
// Body: { new_assignee_id: string, reason: string }
func AdminReassignWorkflowTask(c *fiber.Ctx) error {
	db := config.DB
	taskID := c.Params("id")
	callerID, _ := c.Locals("userID").(string)

	var req struct {
		NewAssigneeID string `json:"new_assignee_id"`
		Reason        string `json:"reason"`
	}
	if err := c.BodyParser(&req); err != nil {
		return utils.SendBadRequest(c, "Invalid request body")
	}
	if req.NewAssigneeID == "" {
		return utils.SendBadRequest(c, "new_assignee_id is required")
	}

	var task models.WorkflowTask
	if err := db.First(&task, "id = ?", taskID).Error; err != nil {
		return utils.SendNotFound(c, "Workflow task not found")
	}

	// Verify the new assignee exists
	var userCount int64
	db.Table("users").Where("id = ? AND deleted_at IS NULL", req.NewAssigneeID).Count(&userCount)
	if userCount == 0 {
		return utils.SendNotFound(c, "New assignee user not found")
	}

	now := time.Now()
	if err := db.Table("workflow_tasks").Where("id = ?", taskID).Updates(map[string]interface{}{
		"assignment_type":  "specific_user",
		"assigned_user_id": req.NewAssigneeID,
		"status":           "pending",
		"claimed_by":       nil,
		"claimed_at":       nil,
		"updated_at":       now,
	}).Error; err != nil {
		return utils.SendInternalError(c, "Failed to reassign workflow task", err)
	}

	db.Table("admin_audit_logs").Create(map[string]interface{}{
		"id":            utils.GenerateID(),
		"action":        "support_workflow_task_reassigned",
		"admin_user_id": callerID,
		"new_value":     fmt.Sprintf("task:%s → user:%s", taskID, req.NewAssigneeID),
		"description":   fmt.Sprintf("Reason: %s", req.Reason),
		"created_at":    now,
	})

	return utils.SendSimpleSuccess(c, map[string]interface{}{"id": taskID}, "Workflow task reassigned successfully")
}

// AdminResetWorkflowTask resets a stuck/expired claimed task back to pending so eligible users can claim it.
// Body: { reason: string }
func AdminResetWorkflowTask(c *fiber.Ctx) error {
	db := config.DB
	taskID := c.Params("id")
	callerID, _ := c.Locals("userID").(string)

	var req struct {
		Reason string `json:"reason"`
	}
	_ = c.BodyParser(&req)

	var task models.WorkflowTask
	if err := db.First(&task, "id = ?", taskID).Error; err != nil {
		return utils.SendNotFound(c, "Workflow task not found")
	}

	now := time.Now()
	if err := db.Table("workflow_tasks").Where("id = ?", taskID).Updates(map[string]interface{}{
		"status":     "pending",
		"claimed_by": nil,
		"claimed_at": nil,
		"updated_at": now,
	}).Error; err != nil {
		return utils.SendInternalError(c, "Failed to reset workflow task", err)
	}

	db.Table("admin_audit_logs").Create(map[string]interface{}{
		"id":            utils.GenerateID(),
		"action":        "support_workflow_task_reset",
		"admin_user_id": callerID,
		"new_value":     taskID,
		"description":   fmt.Sprintf("Task reset to pending. Reason: %s", req.Reason),
		"created_at":    now,
	})

	return utils.SendSimpleSuccess(c, map[string]interface{}{"id": taskID}, "Workflow task reset to pending successfully")
}
