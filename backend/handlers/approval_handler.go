package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/services"
	"github.com/liyali/liyali-gateway/types"
	"github.com/liyali/liyali-gateway/utils"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ApproverInfo represents an approver
type ApproverInfo struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type ApprovalHandler struct {
	validate *validator.Validate
}

func NewApprovalHandler() *ApprovalHandler {
	return &ApprovalHandler{
		validate: validator.New(),
	}
}

// Request/Response Types
type ApproveTaskRequest struct {
	Signature       string `json:"signature" validate:"required"`
	Comment         string `json:"comment"`
	ExpectedVersion int    `json:"expectedVersion"`
}

type RejectTaskRequest struct {
	Signature       string `json:"signature"`
	Reason          string `json:"reason" validate:"required"`
	ExpectedVersion int    `json:"expectedVersion"`
	RejectionType   string `json:"rejectionType"`  // "reject" (default), "return_to_draft", or "return_to_previous_stage"
	ReturnToStage   int    `json:"returnToStage"`   // Unused — kept for API compatibility
}

type ReassignTaskRequest struct {
	NewUserID string `json:"newUserId" validate:"required"`
	Reason    string `json:"reason"`
}

type ClaimTaskRequest struct {
	// No additional fields needed for claiming
}

type UnclaimTaskRequest struct {
	// No additional fields needed for unclaiming
}

type BulkApproveRequest struct {
	TaskIDs   []string `json:"taskIds" validate:"required,min=1"`
	Signature string   `json:"signature" validate:"required"`
	Comment   string   `json:"comment"`
}

type BulkRejectRequest struct {
	TaskIDs   []string `json:"taskIds" validate:"required,min=1"`
	Signature string   `json:"signature" validate:"required"`
	Reason    string   `json:"reason" validate:"required"`
}

type BulkReassignRequest struct {
	TaskIDs   []string `json:"taskIds" validate:"required,min=1"`
	NewUserID string   `json:"newUserId" validate:"required"`
	Reason    string   `json:"reason"`
}

type BulkOperationResponse struct {
	SuccessCount int      `json:"successCount"`
	FailureCount int      `json:"failureCount"`
	SuccessIDs   []string `json:"successIds"`
	Errors       []string `json:"errors,omitempty"`
}

// GetTaskStats retrieves task statistics for the current user
// GET /api/v1/approvals/stats
func (h *ApprovalHandler) GetTaskStats(c *fiber.Ctx) error {
	db := config.DB
	organizationID := c.Locals("organizationID").(string)
	userID := c.Locals("userID").(string)

	// Get user role for role-based filtering
	var user models.User
	if err := db.Where("id = ?", userID).First(&user).Error; err != nil {
		return utils.SendUnauthorizedError(c, "User not found")
	}

	// Build permission filter - same logic as approval permissions
	// Users can see tasks if:
	// 1. Task is assigned directly to them (assigned_user_id = userID)
	// 2. Task is assigned to their role (assigned_role = user.Role)
	// 3. User has a built-in approver role (admin, approver, finance, manager, supervisor, department_head)
	// 4. User has an organization role with approval permissions
	approverRoles := []string{"admin", "approver", "finance", "manager", "supervisor", "department_head"}
	isBuiltInApprover := false
	for _, role := range approverRoles {
		if strings.EqualFold(user.Role, role) {
			isBuiltInApprover = true
			break
		}
	}

	// Check for org role with approval permissions
	hasOrgApprovalPermission := false
	approvalPermissions := []string{"requisition.approve", "approval.approve", "budget.approve", "purchase_order.approve", "payment_voucher.approve", "grn.approve"}
	var userOrgRoles []models.UserOrganizationRole
	if err := db.Where("user_id = ? AND organization_id = ? AND active = ?", userID, organizationID, true).Find(&userOrgRoles).Error; err == nil && len(userOrgRoles) > 0 {
		for _, userOrgRole := range userOrgRoles {
			var orgRole models.OrganizationRole
			if err := db.Where("id = ? AND active = ?", userOrgRole.RoleID, true).First(&orgRole).Error; err != nil {
				continue
			}
			var permissions []string
			if err := json.Unmarshal(orgRole.Permissions, &permissions); err != nil {
				continue
			}
			for _, perm := range permissions {
				for _, approvalPerm := range approvalPermissions {
					if strings.EqualFold(perm, approvalPerm) {
						hasOrgApprovalPermission = true
						break
					}
				}
				if hasOrgApprovalPermission {
					break
				}
			}
			if hasOrgApprovalPermission {
				break
			}
		}
	}

	// Build the permission condition
	// If user is a built-in approver or has org approval permissions, they can see ALL tasks in the organization
	// Otherwise, they can only see tasks assigned to them or their role
	var permissionCondition string
	var permissionArgs []interface{}

	if isBuiltInApprover || hasOrgApprovalPermission {
		// User can see all tasks in the organization
		permissionCondition = "organization_id = ?"
		permissionArgs = []interface{}{organizationID}
	} else {
		// User can only see tasks assigned to them, their role name, or their custom org role UUIDs
		orgRoleUUIDs := make([]string, 0, len(userOrgRoles))
		for _, uor := range userOrgRoles {
			orgRoleUUIDs = append(orgRoleUUIDs, uor.RoleID.String())
		}
		if len(orgRoleUUIDs) > 0 {
			permissionCondition = "organization_id = ? AND (assigned_user_id = ? OR LOWER(assigned_role) = LOWER(?) OR assigned_role IN (?))"
			permissionArgs = []interface{}{organizationID, userID, user.Role, orgRoleUUIDs}
		} else {
			permissionCondition = "organization_id = ? AND (assigned_user_id = ? OR LOWER(assigned_role) = LOWER(?))"
			permissionArgs = []interface{}{organizationID, userID, user.Role}
		}
	}

	// Count total tasks
	var totalTasks int64
	db.Table("workflow_tasks").Where(permissionCondition, permissionArgs...).Count(&totalTasks)

	// Count pending tasks
	var pendingTasks int64
	db.Table("workflow_tasks").
		Where(permissionCondition, permissionArgs...).
		Where("LOWER(status) = LOWER(?)", "pending").
		Count(&pendingTasks)

	// Count completed tasks
	var completedTasks int64
	db.Table("workflow_tasks").
		Where(permissionCondition, permissionArgs...).
		Where("LOWER(status) IN (LOWER(?), LOWER(?))", "completed", "approved").
		Count(&completedTasks)

	// Count overdue tasks (pending tasks past due date)
	var overdueTasks int64
	db.Table("workflow_tasks").
		Where(permissionCondition, permissionArgs...).
		Where("LOWER(status) = LOWER(?)", "pending").
		Where("due_date IS NOT NULL AND due_date < ?", time.Now()).
		Count(&overdueTasks)

	// Count high priority tasks (high or urgent)
	var highPriorityTasks int64
	db.Table("workflow_tasks").
		Where(permissionCondition, permissionArgs...).
		Where("LOWER(priority) IN (LOWER(?), LOWER(?))", "high", "urgent").
		Where("LOWER(status) = LOWER(?)", "pending").
		Count(&highPriorityTasks)

	// Count by entity type
	type EntityTypeCount struct {
		EntityType string
		Count      int64
	}
	var byType []EntityTypeCount
	db.Table("workflow_tasks").
		Select("entity_type, COUNT(*) as count").
		Where(permissionCondition, permissionArgs...).
		Where("LOWER(status) = LOWER(?)", "pending").
		Group("entity_type").
		Scan(&byType)

	byTypeMap := make(map[string]int64)
	for _, item := range byType {
		byTypeMap[item.EntityType] = item.Count
	}

	// Count by priority
	type PriorityCount struct {
		Priority string
		Count    int64
	}
	var byPriority []PriorityCount
	db.Table("workflow_tasks").
		Select("priority, COUNT(*) as count").
		Where(permissionCondition, permissionArgs...).
		Where("LOWER(status) = LOWER(?)", "pending").
		Group("priority").
		Scan(&byPriority)

	byPriorityMap := make(map[string]int64)
	for _, item := range byPriority {
		byPriorityMap[strings.ToUpper(item.Priority)] = item.Count
	}

	stats := map[string]interface{}{
		"totalTasks":         totalTasks,
		"pendingTasks":       pendingTasks,
		"completedTasks":     completedTasks,
		"overdueTasks":       overdueTasks,
		"highPriorityTasks":  highPriorityTasks,
		"byType":             byTypeMap,
		"byPriority":         byPriorityMap,
	}

	return utils.SendSimpleSuccess(c, stats, "Task statistics retrieved successfully")
}

// GetMyPendingCount returns the count of pending approval tasks for the current user
// GET /api/v1/approvals/my-pending-count
func (h *ApprovalHandler) GetMyPendingCount(c *fiber.Ctx) error {
	db := config.DB
	organizationID := c.Locals("organizationID").(string)
	userID := c.Locals("userID").(string)

	// Get user role for role-based filtering
	var user models.User
	if err := db.Where("id = ?", userID).First(&user).Error; err != nil {
		return utils.SendSimpleSuccess(c, map[string]interface{}{"count": 0}, "Pending approval count retrieved successfully")
	}

	// Build permission filter — same logic as GetTaskStats
	approverRoles := []string{"admin", "approver", "finance", "manager", "supervisor", "department_head"}
	isBuiltInApprover := false
	for _, role := range approverRoles {
		if strings.EqualFold(user.Role, role) {
			isBuiltInApprover = true
			break
		}
	}

	// Check for org role with approval permissions
	hasOrgApprovalPermission := false
	approvalPermissions := []string{"requisition.approve", "approval.approve", "budget.approve", "purchase_order.approve", "payment_voucher.approve", "grn.approve"}
	var userOrgRoles []models.UserOrganizationRole
	if err := db.Where("user_id = ? AND organization_id = ? AND active = ?", userID, organizationID, true).Find(&userOrgRoles).Error; err == nil && len(userOrgRoles) > 0 {
		for _, userOrgRole := range userOrgRoles {
			var orgRole models.OrganizationRole
			if err := db.Where("id = ? AND active = ?", userOrgRole.RoleID, true).First(&orgRole).Error; err != nil {
				continue
			}
			var permissions []string
			if err := json.Unmarshal(orgRole.Permissions, &permissions); err != nil {
				continue
			}
			for _, perm := range permissions {
				for _, approvalPerm := range approvalPermissions {
					if strings.EqualFold(perm, approvalPerm) {
						hasOrgApprovalPermission = true
						break
					}
				}
				if hasOrgApprovalPermission {
					break
				}
			}
			if hasOrgApprovalPermission {
				break
			}
		}
	}

	// Build the permission condition
	var permissionCondition string
	var permissionArgs []interface{}

	if isBuiltInApprover || hasOrgApprovalPermission {
		permissionCondition = "organization_id = ?"
		permissionArgs = []interface{}{organizationID}
	} else {
		// Non-approvers can only see tasks assigned to them, their role name, or their custom org role UUIDs
		orgRoleUUIDs := make([]string, 0, len(userOrgRoles))
		for _, uor := range userOrgRoles {
			orgRoleUUIDs = append(orgRoleUUIDs, uor.RoleID.String())
		}
		if len(orgRoleUUIDs) > 0 {
			permissionCondition = "organization_id = ? AND (assigned_user_id = ? OR LOWER(assigned_role) = LOWER(?) OR assigned_role IN (?))"
			permissionArgs = []interface{}{organizationID, userID, user.Role, orgRoleUUIDs}
		} else {
			permissionCondition = "organization_id = ? AND (assigned_user_id = ? OR LOWER(assigned_role) = LOWER(?))"
			permissionArgs = []interface{}{organizationID, userID, user.Role}
		}
	}

	// Count pending tasks
	var pendingCount int64
	db.Table("workflow_tasks").
		Where(permissionCondition, permissionArgs...).
		Where("LOWER(status) = LOWER(?)", "pending").
		Count(&pendingCount)

	return utils.SendSimpleSuccess(c, map[string]interface{}{"count": pendingCount}, "Pending approval count retrieved successfully")
}

// GetApprovalTasks retrieves workflow tasks with pagination and filtering
func (h *ApprovalHandler) GetApprovalTasks(c *fiber.Ctx) error {
	db := config.DB
	organizationID := c.Locals("organizationID").(string)
	userID := c.Locals("userID").(string)

	// Extract query parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	status := c.Query("status", "")
	documentType := c.Query("document_type", "")
	priority := c.Query("priority", "")
	assignedToMe := c.Query("assigned_to_me", "false") == "true"
	viewAll := c.Query("view_all", "false") == "true"

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Get user role for permission checks
	var user models.User
	if err := db.Where("id = ?", userID).First(&user).Error; err != nil {
		return utils.SendUnauthorizedError(c, "User not found")
	}

	// Auto-expire stale claims — reset expired claimed tasks back to pending
	db.Table("workflow_tasks").
		Where("organization_id = ? AND status = ? AND claim_expiry < ?",
			organizationID, "claimed", time.Now()).
		Updates(map[string]interface{}{
			"claimed_by":   nil,
			"claimed_at":   nil,
			"claim_expiry": nil,
			"status":       "pending",
		})

	// Build query for workflow_tasks
	query := db.Table("workflow_tasks").Where("organization_id = ?", organizationID)

	// Build permission filter - same logic as approval permissions
	if viewAll {
		// Dashboard/transparency mode: return all org tasks regardless of role.
		// Action-level permissions are enforced when the user tries to claim/approve.
	} else if assignedToMe {
		// Only show tasks assigned specifically to this user, their role name, or their custom org role UUIDs
		var assignedToMeOrgRoles []models.UserOrganizationRole
		db.Where("user_id = ? AND organization_id = ? AND active = ?", userID, organizationID, true).Find(&assignedToMeOrgRoles)
		orgRoleUUIDs := make([]string, 0, len(assignedToMeOrgRoles))
		for _, uor := range assignedToMeOrgRoles {
			orgRoleUUIDs = append(orgRoleUUIDs, uor.RoleID.String())
		}
		if len(orgRoleUUIDs) > 0 {
			query = query.Where("(assigned_user_id = ? OR LOWER(assigned_role) = LOWER(?) OR assigned_role IN (?))", userID, user.Role, orgRoleUUIDs)
		} else {
			query = query.Where("(assigned_user_id = ? OR LOWER(assigned_role) = LOWER(?))", userID, user.Role)
		}
	} else {
		// Show all tasks the user can approve based on permissions
		// Users can see tasks if:
		// 1. Task is assigned directly to them (assigned_user_id = userID)
		// 2. Task is assigned to their role (assigned_role = user.Role)
		// 3. User has a built-in approver role (admin, approver, finance, manager, supervisor, department_head)
		// 4. User has an organization role with approval permissions
		approverRoles := []string{"admin", "approver", "finance", "manager", "supervisor", "department_head"}
		isBuiltInApprover := false
		for _, role := range approverRoles {
			if strings.EqualFold(user.Role, role) {
				isBuiltInApprover = true
				break
			}
		}

		// Check for org role with approval permissions
		hasOrgApprovalPermission := false
		approvalPermissions := []string{"requisition.approve", "approval.approve", "budget.approve", "purchase_order.approve", "payment_voucher.approve", "grn.approve"}
		var userOrgRoles []models.UserOrganizationRole
		if err := db.Where("user_id = ? AND organization_id = ? AND active = ?", userID, organizationID, true).Find(&userOrgRoles).Error; err == nil && len(userOrgRoles) > 0 {
			for _, userOrgRole := range userOrgRoles {
				var orgRole models.OrganizationRole
				if err := db.Where("id = ? AND active = ?", userOrgRole.RoleID, true).First(&orgRole).Error; err != nil {
					continue
				}
				var permissions []string
				if err := json.Unmarshal(orgRole.Permissions, &permissions); err != nil {
					continue
				}
				for _, perm := range permissions {
					for _, approvalPerm := range approvalPermissions {
						if strings.EqualFold(perm, approvalPerm) {
							hasOrgApprovalPermission = true
							break
						}
					}
					if hasOrgApprovalPermission {
						break
					}
				}
				if hasOrgApprovalPermission {
					break
				}
			}
		}

		// If user is NOT a built-in approver and doesn't have org approval permissions,
		// they can only see tasks assigned to them, their role name, or their custom org role UUIDs
		if !isBuiltInApprover && !hasOrgApprovalPermission {
			orgRoleUUIDs := make([]string, 0, len(userOrgRoles))
			for _, uor := range userOrgRoles {
				orgRoleUUIDs = append(orgRoleUUIDs, uor.RoleID.String())
			}
			if len(orgRoleUUIDs) > 0 {
				query = query.Where("(assigned_user_id = ? OR LOWER(assigned_role) = LOWER(?) OR assigned_role IN (?))", userID, user.Role, orgRoleUUIDs)
			} else {
				query = query.Where("(assigned_user_id = ? OR LOWER(assigned_role) = LOWER(?))", userID, user.Role)
			}
		}
		// Otherwise, they can see all tasks in the organization (no additional filter needed)
	}

	if status != "" {
		query = query.Where("LOWER(status) = LOWER(?)", status)
	}
	if documentType != "" {
		query = query.Where("LOWER(entity_type) = LOWER(?)", documentType)
	}
	if priority != "" {
		query = query.Where("LOWER(priority) = LOWER(?)", priority)
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Get tasks with pagination
	var tasks []models.WorkflowTask
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&tasks).Error; err != nil {
		log.Printf("Error fetching workflow tasks: %v", err)
		return utils.SendInternalError(c, "Failed to fetch workflow tasks", err)
	}

	// Populate computed fields for each task
	for i := range tasks {
		if err := h.populateWorkflowTaskFields(db, &tasks[i]); err != nil {
			log.Printf("Error populating computed fields for task %s: %v", tasks[i].ID, err)
			// Continue with other tasks even if one fails
		}
	}

	// Debug: Verify task fields before sending
	for _, task := range tasks {
		if task.DueAt != nil {
			log.Printf("[DEBUG] Task %s FINAL: priority=%s, dueAt=%v", task.ID, task.Priority, *task.DueAt)
		} else {
			log.Printf("[DEBUG] Task %s FINAL: priority=%s, dueAt=NIL (this is a bug!)", task.ID, task.Priority)
		}
	}

	return utils.SendPaginatedSuccess(c, tasks, "Workflow tasks retrieved successfully", page, limit, total)
}

// populateWorkflowTaskFields populates the computed fields for a WorkflowTask
func (h *ApprovalHandler) populateWorkflowTaskFields(db *gorm.DB, task *models.WorkflowTask) error {
	// Map basic fields for frontend compatibility
	task.DocumentID = task.EntityID
	task.DocumentType = task.EntityType
	task.Stage = task.StageNumber
	task.DueAt = task.DueDate
	
	// Set assigned user fields
	if task.AssignedUserID != nil {
		task.AssignedTo = *task.AssignedUserID
		task.ApproverID = *task.AssignedUserID
		
		// Get user name if available
		if task.AssignedUser != nil {
			task.ApproverName = task.AssignedUser.Name
		} else {
			var user models.User
			if err := db.Where("id = ?", *task.AssignedUserID).First(&user).Error; err == nil {
				task.ApproverName = user.Name
			}
		}
	} else if task.ClaimedBy != nil {
		task.AssignedTo = *task.ClaimedBy
		task.ApproverID = *task.ClaimedBy

		// Get claimer name if available
		if task.Claimer != nil {
			task.ApproverName = task.Claimer.Name
			task.ClaimerName = task.Claimer.Name
		} else {
			var user models.User
			if err := db.Where("id = ?", *task.ClaimedBy).First(&user).Error; err == nil {
				task.ApproverName = user.Name
				task.ClaimerName = user.Name
			}
		}
	}

	// Populate document-specific fields based on entity type
	switch task.EntityType {
	case "requisition":
		var req models.Requisition
		if err := db.Where("id = ?", task.EntityID).First(&req).Error; err == nil {
			task.DocumentNumber = req.DocumentNumber
			task.Title = req.Title + " - Approval Required"
			task.TaskType = "REQUISITION_APPROVAL"
		}
	case "purchase_order":
		var po models.PurchaseOrder
		if err := db.Where("id = ?", task.EntityID).First(&po).Error; err == nil {
			task.DocumentNumber = po.DocumentNumber
			task.Title = po.Title + " - Approval Required"
			task.TaskType = "PURCHASE_ORDER_APPROVAL"
		}
	case "payment_voucher":
		var pv models.PaymentVoucher
		if err := db.Where("id = ?", task.EntityID).First(&pv).Error; err == nil {
			task.DocumentNumber = pv.DocumentNumber
			task.Title = pv.Title + " - Approval Required"
			task.TaskType = "PAYMENT_VOUCHER_APPROVAL"
		}
	case "budget":
		var budget models.Budget
		if err := db.Where("id = ?", task.EntityID).First(&budget).Error; err == nil {
			task.DocumentNumber = budget.BudgetCode
			task.Title = budget.Name + " - Approval Required"
			task.TaskType = "BUDGET_APPROVAL"
		}
	case "goods_received_note":
		var grn models.GoodsReceivedNote
		if err := db.Where("id = ?", task.EntityID).First(&grn).Error; err == nil {
			task.DocumentNumber = grn.DocumentNumber
			task.Title = "GRN " + grn.DocumentNumber + " - Confirmation Required"
			task.TaskType = "GOODS_RECEIVED_NOTE_CONFIRMATION"
		}
	}

	// Set default values if not populated
	if task.TaskType == "" {
		task.TaskType = "APPROVAL"
	}
	if task.Title == "" {
		task.Title = "Approval Required"
	}

	// Set default due date if not set (for tasks created before the default due date feature)
	if task.DueDate == nil {
		// Default due date: 7 days from task creation
		defaultDueDate := task.CreatedAt.Add(7 * 24 * time.Hour)
		task.DueDate = &defaultDueDate
		task.DueAt = &defaultDueDate
		log.Printf("[DEBUG] Task %s: Set default due date to %v (created at %v)", task.ID, defaultDueDate, task.CreatedAt)
	} else {
		log.Printf("[DEBUG] Task %s: Has due date %v", task.ID, *task.DueDate)
	}

	// Ensure priority is set
	if task.Priority == "" {
		task.Priority = "medium"
	}

	// Set importance based on priority
	switch strings.ToLower(task.Priority) {
	case "high", "urgent":
		task.Importance = "high"
	case "low":
		task.Importance = "low"
	default:
		task.Importance = "medium"
	}

	// Get workflow name from workflow assignment
	var assignment models.WorkflowAssignment
	if err := db.Where("id = ?", task.WorkflowAssignmentID).First(&assignment).Error; err == nil {
		task.WorkflowID = assignment.WorkflowID.String()
		var workflow models.Workflow
		if err := db.Where("id = ?", assignment.WorkflowID).First(&workflow).Error; err == nil {
			task.WorkflowName = workflow.Name
		}
	}

	if task.WorkflowName == "" {
		task.WorkflowName = "Standard Approval Workflow"
	}

	// Resolve AssignedRoleName: if AssignedRole is a UUID look it up; otherwise it's already a name.
	if task.AssignedRole != nil && *task.AssignedRole != "" {
		if _, parseErr := uuid.Parse(*task.AssignedRole); parseErr == nil {
			var orgRole models.OrganizationRole
			if db.Where("id = ?", *task.AssignedRole).First(&orgRole).Error == nil {
				task.AssignedRoleName = orgRole.Name
			}
		} else {
			task.AssignedRoleName = *task.AssignedRole // already a plain name
		}
	}

	return nil
}

// GetApprovalTask retrieves a single workflow task with full details
func (h *ApprovalHandler) GetApprovalTask(c *fiber.Ctx) error {
	taskID := c.Params("id")
	if taskID == "" {
		return utils.SendBadRequestError(c, "Task ID is required")
	}

	db := config.DB
	organizationID := c.Locals("organizationID").(string)
	userID := c.Locals("userID").(string)

	var task models.WorkflowTask
	query := db.Where("id = ? AND organization_id = ?", taskID, organizationID)
	
	// Check if user can access this task (either assigned to them or they have admin role)
	var user models.User
	if err := db.Where("id = ?", userID).First(&user).Error; err == nil {
		if user.Role != "admin" {
			// Non-admin users can only see tasks assigned to them, their role name, or their custom org role UUIDs
			var userTaskOrgRoles []models.UserOrganizationRole
			db.Where("user_id = ? AND organization_id = ? AND active = ?", userID, organizationID, true).Find(&userTaskOrgRoles)
			orgRoleUUIDs := make([]string, 0, len(userTaskOrgRoles))
			for _, uor := range userTaskOrgRoles {
				orgRoleUUIDs = append(orgRoleUUIDs, uor.RoleID.String())
			}
			if len(orgRoleUUIDs) > 0 {
				query = query.Where("(assigned_user_id = ? OR LOWER(assigned_role) = LOWER(?) OR assigned_role IN (?))", userID, user.Role, orgRoleUUIDs)
			} else {
				query = query.Where("(assigned_user_id = ? OR LOWER(assigned_role) = LOWER(?))", userID, user.Role)
			}
		}
	}

	if err := query.First(&task).Error; err != nil {
		log.Printf("Error fetching workflow task %s: %v", taskID, err)
		return utils.SendNotFoundError(c, "Workflow task not found or access denied")
	}

	// Populate computed fields
	if err := h.populateWorkflowTaskFields(db, &task); err != nil {
		log.Printf("Error populating computed fields for task %s: %v", taskID, err)
	}

	return utils.SendSimpleSuccess(c, task, "Workflow task retrieved successfully")
}

// ClaimTask claims a workflow task for exclusive access
// POST /api/v1/approvals/tasks/:id/claim
func (h *ApprovalHandler) ClaimTask(c *fiber.Ctx) error {
	taskID := c.Params("id")
	if taskID == "" {
		return utils.SendBadRequestError(c, "Task ID is required")
	}

	userID := c.Locals("userID").(string)

	workflowExecutionService := c.Locals("workflowExecutionService").(*services.WorkflowExecutionService)

	err := workflowExecutionService.ClaimWorkflowTask(c.Context(), taskID, userID)
	if err != nil {
		log.Printf("Error claiming workflow task %s: %v", taskID, err)
		return c.Status(fiber.StatusConflict).JSON(types.ErrorResponse{
			Error:   "Claim failed",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(types.SuccessResponse{
		Message: "Task claimed successfully",
		Data:    map[string]interface{}{"taskId": taskID, "claimedBy": userID},
	})
}

// UnclaimTask releases a claimed task
// POST /api/v1/approvals/tasks/:id/unclaim
func (h *ApprovalHandler) UnclaimTask(c *fiber.Ctx) error {
	taskID := c.Params("id")
	if taskID == "" {
		return utils.SendBadRequestError(c, "Task ID is required")
	}

	userID := c.Locals("userID").(string)

	workflowExecutionService := c.Locals("workflowExecutionService").(*services.WorkflowExecutionService)

	err := workflowExecutionService.UnclaimWorkflowTask(c.Context(), taskID, userID)
	if err != nil {
		log.Printf("Error unclaiming workflow task %s: %v", taskID, err)
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{
			Error:   "Unclaim failed",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(types.SuccessResponse{
		Message: "Task unclaimed successfully",
		Data:    map[string]interface{}{"taskId": taskID},
	})
}

// ApproveTask marks a task as approved and moves to next stage
func (h *ApprovalHandler) ApproveTask(c *fiber.Ctx) error {
	taskID := c.Params("id")
	userID := c.Locals("userID").(string)

	var req ApproveTaskRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{
			Error:   "Invalid request body",
			Message: "Failed to parse approval request",
		})
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{
			Error:   "Validation failed",
			Message: err.Error(),
		})
	}

	// Get workflow execution service
	workflowExecutionService := c.Locals("workflowExecutionService").(*services.WorkflowExecutionService)

	// Use workflow system to approve the task with version control
	var err error
	if req.ExpectedVersion > 0 {
		err = workflowExecutionService.ApproveWorkflowTaskWithVersion(c.Context(), taskID, userID, req.Signature, req.Comment, req.ExpectedVersion)
	} else {
		err = workflowExecutionService.ApproveWorkflowTask(c.Context(), taskID, userID, req.Signature, req.Comment)
	}
	
	if err != nil {
		log.Printf("Error approving workflow task: %v", err)
		
		// Handle specific error types
		if contains(err.Error(), "version") || contains(err.Error(), "modified by another user") {
			return c.Status(fiber.StatusConflict).JSON(types.ErrorResponse{
				Error:   "Concurrent modification",
				Message: err.Error(),
			})
		}
		
		if contains(err.Error(), "claimed by another user") || contains(err.Error(), "claim has expired") {
			return c.Status(fiber.StatusConflict).JSON(types.ErrorResponse{
				Error:   "Task claim issue",
				Message: err.Error(),
			})
		}
		
		return c.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse{
			Error:   "Approval failed",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(types.SuccessResponse{
		Message: "Task approved successfully",
		Data:    map[string]interface{}{"taskId": taskID},
	})
}

// RejectTask marks a task as rejected and returns document to draft
func (h *ApprovalHandler) RejectTask(c *fiber.Ctx) error {
	taskID := c.Params("id")
	userID := c.Locals("userID").(string)

	var req RejectTaskRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{
			Error:   "Invalid request body",
			Message: "Failed to parse rejection request",
		})
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{
			Error:   "Validation failed",
			Message: err.Error(),
		})
	}

	// Get workflow execution service
	workflowExecutionService := c.Locals("workflowExecutionService").(*services.WorkflowExecutionService)

	// Use workflow system to reject the task with version control
	var err error
	if req.ExpectedVersion > 0 {
		err = workflowExecutionService.RejectWorkflowTaskWithVersion(c.Context(), taskID, userID, req.Signature, req.Reason, req.ExpectedVersion, req.RejectionType, req.ReturnToStage)
	} else {
		err = workflowExecutionService.RejectWorkflowTask(c.Context(), taskID, userID, req.Signature, req.Reason, req.RejectionType, req.ReturnToStage)
	}
	
	if err != nil {
		log.Printf("Error rejecting workflow task: %v", err)
		
		// Handle specific error types
		if strings.Contains(err.Error(), "version") || strings.Contains(err.Error(), "modified by another user") {
			return c.Status(fiber.StatusConflict).JSON(types.ErrorResponse{
				Error:   "Concurrent modification",
				Message: err.Error(),
			})
		}
		
		if strings.Contains(err.Error(), "claimed by another user") || strings.Contains(err.Error(), "claim has expired") {
			return c.Status(fiber.StatusConflict).JSON(types.ErrorResponse{
				Error:   "Task claim issue",
				Message: err.Error(),
			})
		}
		
		return c.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse{
			Error:   "Rejection failed",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(types.SuccessResponse{
		Message: "Task rejected successfully",
		Data:    map[string]interface{}{"taskId": taskID},
	})
}

// ReassignTask reassigns workflow task to different approver
func (h *ApprovalHandler) ReassignTask(c *fiber.Ctx) error {
	taskID := c.Params("id")
	organizationID := c.Locals("organizationID").(string)
	userID := c.Locals("userID").(string)

	var req ReassignTaskRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{
			Error:   "Invalid request body",
			Message: "Failed to parse reassignment request",
		})
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{
			Error:   "Validation failed",
			Message: err.Error(),
		})
	}

	db := config.DB

	// Get the workflow task
	var task models.WorkflowTask
	if err := db.Where("id = ? AND organization_id = ?", taskID, organizationID).First(&task).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(types.ErrorResponse{
			Error:   "Task not found",
			Message: "Workflow task not found",
		})
	}

	// Check if task is in pending status
	if task.Status != "pending" && task.Status != "claimed" {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{
			Error:   "Invalid task status",
			Message: "Task is not in pending or claimed status",
		})
	}

	// Get current user
	var user models.User
	if err := db.Where("id = ?", userID).First(&user).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(types.ErrorResponse{
			Error:   "User not found",
			Message: "Current user not found",
		})
	}

	// Built-in roles that can reassign tasks
	reassignRoles := []string{"admin", "approver", "finance", "manager", "supervisor", "department_head"}

	// Reassignment-related permissions to check for in organization roles
	reassignPermissions := []string{
		"approval.reassign", "requisition.reassign", "workflow.reassign",
		"approval.manage", "workflow.manage",
	}

	// Helper function to check if user has any organization role with reassign permissions
	checkOrgRoleReassignPermissions := func() bool {
		var userOrgRoles []models.UserOrganizationRole
		if err := db.Where("user_id = ? AND organization_id = ? AND active = ?",
			userID, organizationID, true).Find(&userOrgRoles).Error; err != nil || len(userOrgRoles) == 0 {
			return false
		}

		for _, userOrgRole := range userOrgRoles {
			var orgRole models.OrganizationRole
			if err := db.Where("id = ? AND active = ?", userOrgRole.RoleID, true).First(&orgRole).Error; err != nil {
				continue
			}

			// Parse permissions from JSON
			var permissions []string
			if err := json.Unmarshal(orgRole.Permissions, &permissions); err != nil {
				continue
			}

			// Check if any reassign permission exists
			for _, perm := range permissions {
				for _, reassignPerm := range reassignPermissions {
					if strings.EqualFold(perm, reassignPerm) {
						log.Printf("[DEBUG] User has organization role '%s' with reassign permission '%s'", orgRole.Name, perm)
						return true
					}
				}
			}
		}
		return false
	}

	// Check if user has permission to reassign
	hasReassignPermission := false

	// Check 1: Built-in role check
	for _, role := range reassignRoles {
		if strings.EqualFold(user.Role, role) {
			hasReassignPermission = true
			log.Printf("[DEBUG] User has built-in role '%s' - can reassign", user.Role)
			break
		}
	}

	// Check 2: Organization role with reassign permissions
	if !hasReassignPermission {
		hasReassignPermission = checkOrgRoleReassignPermissions()
	}

	if !hasReassignPermission {
		return c.Status(fiber.StatusForbidden).JSON(types.ErrorResponse{
			Error:   "Insufficient permissions",
			Message: "You don't have permission to reassign tasks",
		})
	}

	// Store previous assignee info for notification
	var previousAssigneeName string
	var previousAssigneeID string
	if task.AssignedUserID != nil {
		previousAssigneeID = *task.AssignedUserID
		var prevUser models.User
		if err := db.Where("id = ?", previousAssigneeID).First(&prevUser).Error; err == nil {
			previousAssigneeName = prevUser.Name
		}
	} else if task.ClaimedBy != nil {
		previousAssigneeID = *task.ClaimedBy
		var prevUser models.User
		if err := db.Where("id = ?", previousAssigneeID).First(&prevUser).Error; err == nil {
			previousAssigneeName = prevUser.Name
		}
	}

	// Validate that the new user exists and is active
	var newUser models.User
	if err := db.Where("id = ? AND active = ? AND current_organization_id = ?", req.NewUserID, true, organizationID).First(&newUser).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{
			Error:   "Invalid assignee",
			Message: "Target user not found or inactive",
		})
	}

	// Update task assignment - clear any existing claim and assign to new user
	// This ensures ONLY the assigned user can approve/reject
	task.AssignedUserID = &req.NewUserID
	task.AssignedRole = nil // Clear role assignment since we're assigning to a specific user
	task.ClaimedBy = nil
	task.ClaimedAt = nil
	task.ClaimExpiry = nil
	task.Version += 1 // Increment version for optimistic locking
	task.UpdatedBy = &userID

	if err := db.Save(&task).Error; err != nil {
		log.Printf("Error reassigning workflow task: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to reassign task",
		})
	}

	// Create audit log entry
	changes := datatypes.NewJSONType(map[string]interface{}{
		"previousAssignee":     previousAssigneeID,
		"previousAssigneeName": previousAssigneeName,
		"newAssignee":          req.NewUserID,
		"newAssigneeName":      newUser.Name,
		"reason":               req.Reason,
		"reassignedBy":         user.Name,
		"reassignedById":       userID,
	})

	auditLog := models.AuditLog{
		ID:           fmt.Sprintf("audit-%d", time.Now().UnixNano()),
		DocumentID:   task.EntityID,
		DocumentType: task.EntityType,
		UserID:       userID,
		Action:       "reassign",
		Changes:      changes,
		CreatedAt:    time.Now(),
	}

	if err := db.Create(&auditLog).Error; err != nil {
		log.Printf("Error creating audit log for task reassignment: %v", err)
		// Don't fail the request for audit log errors
	}

	// Add action history entry to the document
	if err := addReassignmentActionHistory(db, task.EntityType, task.EntityID, userID, user.Name, previousAssigneeName, newUser.Name, req.Reason); err != nil {
		log.Printf("Warning: failed to add action history entry for reassignment: %v", err)
		// Don't fail the request for action history errors
	}

	// Create in-app notification for the new assignee
	notification := models.Notification{
		ID:                 uuid.New().String(),
		OrganizationID:     organizationID,
		RecipientID:        req.NewUserID,
		Type:               "task_reassigned",
		DocumentID:         task.EntityID,
		DocumentType:       task.EntityType,
		EntityID:           task.EntityID,
		EntityType:         task.EntityType,
		Subject:            fmt.Sprintf("Task Reassigned: %s Approval", task.EntityType),
		Body:               fmt.Sprintf("A %s approval task has been reassigned to you by %s. Stage: %s. Reason: %s", task.EntityType, user.Name, task.StageName, req.Reason),
		Message:            fmt.Sprintf("Task reassigned to you by %s", user.Name),
		RelatedUserID:      userID,
		RelatedUserName:    user.Name,
		ReassignmentReason: req.Reason,
		Importance:         "HIGH",
		Sent:               false,
		IsRead:             false,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := db.Create(&notification).Error; err != nil {
		log.Printf("Error creating reassignment notification: %v", err)
		// Don't fail the request for notification errors
	} else {
		log.Printf("[DEBUG] Created reassignment notification for user %s", newUser.Name)
	}

	// If there was a previous assignee (different from the new one), notify them too
	if previousAssigneeID != "" && previousAssigneeID != req.NewUserID {
		prevNotification := models.Notification{
			ID:                 uuid.New().String(),
			OrganizationID:     organizationID,
			RecipientID:        previousAssigneeID,
			Type:               "task_reassigned_away",
			DocumentID:         task.EntityID,
			DocumentType:       task.EntityType,
			EntityID:           task.EntityID,
			EntityType:         task.EntityType,
			Subject:            fmt.Sprintf("Task Reassigned: %s Approval", task.EntityType),
			Body:               fmt.Sprintf("A %s approval task has been reassigned from you to %s by %s. Reason: %s", task.EntityType, newUser.Name, user.Name, req.Reason),
			Message:            fmt.Sprintf("Task reassigned to %s by %s", newUser.Name, user.Name),
			RelatedUserID:      userID,
			RelatedUserName:    user.Name,
			ReassignmentReason: req.Reason,
			Importance:         "MEDIUM",
			Sent:               false,
			IsRead:             false,
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		}

		if err := db.Create(&prevNotification).Error; err != nil {
			log.Printf("Error creating notification for previous assignee: %v", err)
		}
	}

	return c.Status(fiber.StatusOK).JSON(types.SuccessResponse{
		Message: "Task reassigned successfully",
		Data: map[string]interface{}{
			"taskId":              task.ID,
			"newAssignee":         newUser.Name,
			"newAssigneeId":       req.NewUserID,
			"previousAssignee":    previousAssigneeName,
			"previousAssigneeId":  previousAssigneeID,
			"reassignedBy":        user.Name,
			"reason":              req.Reason,
		},
	})
}

// addReassignmentActionHistory adds an action history entry for task reassignment
func addReassignmentActionHistory(db *gorm.DB, entityType, entityID, performedByID, performedByName, previousAssignee, newAssignee, reason string) error {
	entry := types.ActionHistoryEntry{
		ID:              uuid.New().String(),
		Action:          "TASK_REASSIGNED",
		PerformedBy:     performedByID,
		PerformedByName: performedByName,
		PerformedAt:     time.Now(),
		Timestamp:       time.Now(),
		Comments:        fmt.Sprintf("Task reassigned from %s to %s. Reason: %s", previousAssignee, newAssignee, reason),
	}

	switch strings.ToLower(entityType) {
	case "requisition":
		var requisition models.Requisition
		if err := db.First(&requisition, "id = ?", entityID).Error; err != nil {
			return fmt.Errorf("failed to find requisition: %w", err)
		}

		// Get current action history
		currentHistory := requisition.ActionHistory.Data()
		if currentHistory == nil {
			currentHistory = []types.ActionHistoryEntry{}
		}

		// Append new entry
		currentHistory = append(currentHistory, entry)
		requisition.ActionHistory = datatypes.NewJSONType(currentHistory)

		if err := db.Model(&requisition).Update("action_history", requisition.ActionHistory).Error; err != nil {
			return fmt.Errorf("failed to update action history: %w", err)
		}

	// Add other document types as needed
	default:
		log.Printf("Action history not configured for entity type: %s", entityType)
	}

	return nil
}

// GetApprovalHistory retrieves approval history for a document
func (h *ApprovalHandler) GetApprovalHistory(c *fiber.Ctx) error {
	documentID := c.Params("documentId")
	organizationID := c.Locals("organizationID").(string) // Fixed: was "organizationId"

	if documentID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{
			Error:   "Invalid request",
			Message: "Document ID is required",
		})
	}

	db := config.DB

	// First, try to find the actual document ID if a requisition number was provided
	var actualDocumentID string
	var requisition models.Requisition
	
	// Try to find requisition by ID or document_number
	err := db.Where("id = ? OR document_number = ?", documentID, documentID).
		First(&requisition).Error
	
	if err == nil {
		// Found requisition, use its actual ID
		actualDocumentID = requisition.ID
	} else {
		// Assume it's already a valid document ID
		actualDocumentID = documentID
	}

	// Get workflow tasks history instead of legacy approval tasks
	var history []models.WorkflowTask
	if err := db.Where("entity_id = ? AND organization_id = ?", actualDocumentID, organizationID).
		Order("created_at ASC").Find(&history).Error; err != nil {
		log.Printf("Error fetching workflow task history: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to fetch workflow task history",
		})
	}

	return c.Status(fiber.StatusOK).JSON(history)
}

// BulkApprove approves multiple tasks at once
// POST /api/v1/approvals/bulk/approve
func (h *ApprovalHandler) BulkApprove(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	var req BulkApproveRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendBadRequestError(c, "Invalid request body")
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		return utils.SendBadRequestError(c, "Validation failed: "+err.Error())
	}

	// Get workflow execution service
	workflowExecutionService := c.Locals("workflowExecutionService").(*services.WorkflowExecutionService)

	var successIDs []string
	var errors []string

	// Process each task through workflow system
	for _, taskID := range req.TaskIDs {
		err := workflowExecutionService.ApproveWorkflowTask(c.Context(), taskID, userID, req.Signature, req.Comment)
		if err != nil {
			errors = append(errors, "Task "+taskID+": "+err.Error())
			continue
		}
		successIDs = append(successIDs, taskID)
	}

	return utils.SendSimpleSuccess(c, BulkOperationResponse{
		SuccessCount: len(successIDs),
		FailureCount: len(errors),
		SuccessIDs:   successIDs,
		Errors:       errors,
	}, "Bulk approval completed")
}

// BulkReject rejects multiple tasks at once
// POST /api/v1/approvals/bulk/reject
func (h *ApprovalHandler) BulkReject(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	var req BulkRejectRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendBadRequestError(c, "Invalid request body")
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		return utils.SendBadRequestError(c, "Validation failed: "+err.Error())
	}

	// Get workflow execution service
	workflowExecutionService := c.Locals("workflowExecutionService").(*services.WorkflowExecutionService)

	var successIDs []string
	var errors []string

	// Process each task through workflow system
	for _, taskID := range req.TaskIDs {
		err := workflowExecutionService.RejectWorkflowTask(c.Context(), taskID, userID, req.Signature, req.Reason, "reject", 0)
		if err != nil {
			errors = append(errors, "Task "+taskID+": "+err.Error())
			continue
		}
		successIDs = append(successIDs, taskID)
	}

	return utils.SendSimpleSuccess(c, BulkOperationResponse{
		SuccessCount: len(successIDs),
		FailureCount: len(errors),
		SuccessIDs:   successIDs,
		Errors:       errors,
	}, "Bulk rejection completed")
}

// BulkReassign reassigns multiple tasks at once
// POST /api/v1/approvals/bulk/reassign
func (h *ApprovalHandler) BulkReassign(c *fiber.Ctx) error {
	organizationID := c.Locals("organizationID").(string)
	userID := c.Locals("userID").(string)

	var req BulkReassignRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendBadRequestError(c, "Invalid request body")
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		return utils.SendBadRequestError(c, "Validation failed: "+err.Error())
	}

	db := config.DB
	var successIDs []string
	var errors []string

	// Process each task
	for _, taskID := range req.TaskIDs {
		// Get the workflow task
		var task models.WorkflowTask
		if err := db.Where("id = ? AND organization_id = ?", taskID, organizationID).First(&task).Error; err != nil {
			errors = append(errors, "Task "+taskID+": not found")
			continue
		}

		// Check if task is in pending status
		if task.Status != "pending" {
			errors = append(errors, "Task "+taskID+": not in pending status")
			continue
		}

		// Update task assignment
		task.AssignedUserID = &req.NewUserID
		task.UpdatedBy = &userID

		if err := db.Save(&task).Error; err != nil {
			errors = append(errors, "Task "+taskID+": failed to reassign")
			continue
		}

		successIDs = append(successIDs, taskID)
	}

	return utils.SendSimpleSuccess(c, BulkOperationResponse{
		SuccessCount: len(successIDs),
		FailureCount: len(errors),
		SuccessIDs:   successIDs,
		Errors:       errors,
	}, "Bulk reassignment completed")
}

// GetOverdueTasks retrieves tasks that are past their due date
// GET /api/v1/approvals/tasks/overdue
func (h *ApprovalHandler) GetOverdueTasks(c *fiber.Ctx) error {
	organizationID := c.Locals("organizationID").(string) // Fixed: was "organizationId"

	// Get query parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	db := config.DB

	// Get overdue workflow tasks (tasks with due_date in the past and still pending)
	var tasks []models.WorkflowTask
	if err := db.Where("organization_id = ? AND status = ? AND created_at < ?", 
		organizationID, "pending", time.Now()).
		Offset(offset).Limit(limit).Order("created_at ASC").Find(&tasks).Error; err != nil {
		log.Printf("Error fetching overdue tasks: %v", err)
		return utils.SendInternalError(c, "Failed to retrieve overdue tasks", err)
	}

	// Get total count
	var total int64
	db.Model(&models.WorkflowTask{}).Where("organization_id = ? AND status = ? AND due_date < ?", 
		organizationID, "pending", time.Now()).Count(&total)

	return utils.SendPaginatedSuccess(c, tasks, "Overdue tasks retrieved successfully", page, limit, total)
}

// GetApprovalWorkflowStatus retrieves the current approval workflow status for a document
// GET /api/v1/documents/{documentId}/approval-status
func (h *ApprovalHandler) GetApprovalWorkflowStatus(c *fiber.Ctx) error {
	documentID := c.Params("documentId")
	organizationID := c.Locals("organizationID").(string)
	userID := c.Locals("userID").(string)

	if documentID == "" {
		return utils.SendBadRequestError(c, "Document ID is required")
	}

	db := config.DB

	// First, try to find the actual document ID if a requisition number was provided
	var actualDocumentID string
	var requisition models.Requisition
	
	// Try to find requisition by ID or document_number
	err := db.Where("id = ? OR document_number = ?", documentID, documentID).
		First(&requisition).Error
	
	if err == nil {
		// Found requisition, use its actual ID
		actualDocumentID = requisition.ID
	} else {
		// Assume it's already a valid document ID
		actualDocumentID = documentID
	}

	// Get workflow execution service from handler registry
	workflowExecutionService := c.Locals("workflowExecutionService").(*services.WorkflowExecutionService)

	// Get workflow status with detailed stage progress
	workflowStatus, err := workflowExecutionService.GetWorkflowStatus(c.Context(), organizationID, actualDocumentID)
	if err != nil {
		log.Printf("Error fetching workflow status: %v", err)
		return utils.SendInternalError(c, "Failed to fetch workflow status", err)
	}

	// If no workflow is assigned, return basic status
	if workflowStatus.Status == "no_workflow" {
		return c.JSON(types.DetailResponse{
			Success: true,
			Data: map[string]interface{}{
				"currentStage":  0,
				"totalStages":   0,
				"status":        "no_workflow",
				"canApprove":    false,
				"canReject":     false,
				"stageProgress": []interface{}{},
			},
		})
	}

	// Get pending workflow tasks to determine if user can approve
	pendingTasks, err := workflowExecutionService.GetPendingWorkflowTasks(c.Context(), organizationID, actualDocumentID)
	if err != nil {
		log.Printf("Error fetching pending tasks: %v", err)
		// Continue without failing, just set canApprove to false
		pendingTasks = []models.WorkflowTask{}
	}

	canApprove := false
	canReject := false
	nextApprover := ""

	if len(pendingTasks) > 0 {
		currentTask := pendingTasks[0]

		// Built-in roles that have approval permissions
		approverRoles := []string{"admin", "approver", "finance", "manager", "supervisor", "department_head"}

		// Approval-related permissions to check for in organization roles
		approvalPermissions := []string{
			"requisition.approve", "approval.approve", "budget.approve",
			"purchase_order.approve", "payment_voucher.approve", "grn.approve",
		}

		// Helper function to check if user has any organization role with approval permissions
		checkOrgRoleApprovalPermissions := func() bool {
			var userOrgRoles []models.UserOrganizationRole
			if err := db.Where("user_id = ? AND organization_id = ? AND active = ?",
				userID, organizationID, true).Find(&userOrgRoles).Error; err != nil || len(userOrgRoles) == 0 {
				return false
			}

			for _, userOrgRole := range userOrgRoles {
				var orgRole models.OrganizationRole
				if err := db.Where("id = ? AND active = ?", userOrgRole.RoleID, true).First(&orgRole).Error; err != nil {
					continue
				}

				// Parse permissions from JSON
				var permissions []string
				if err := json.Unmarshal(orgRole.Permissions, &permissions); err != nil {
					continue
				}

				// Check if any approval permission exists
				for _, perm := range permissions {
					for _, approvalPerm := range approvalPermissions {
						if strings.EqualFold(perm, approvalPerm) {
							log.Printf("[DEBUG] User has organization role '%s' with approval permission '%s' - canApprove = true", orgRole.Name, perm)
							return true
						}
					}
				}
			}
			return false
		}

		// Get user role to check if they can approve
		var user models.User
		if err := db.Where("id = ?", userID).First(&user).Error; err == nil {
			// PRIORITY 1: If task is assigned to a specific user (after reassignment), ONLY that user can approve
			if currentTask.AssignedUserID != nil && *currentTask.AssignedUserID != "" {
				log.Printf("[DEBUG] Task is assigned to specific user: %s, checking if current user %s matches", *currentTask.AssignedUserID, userID)
				if *currentTask.AssignedUserID == userID {
					log.Printf("[DEBUG] User is the specifically assigned user - canApprove = true")
					canApprove = true
					canReject = true
				} else {
					log.Printf("[DEBUG] User is NOT the specifically assigned user - canApprove = false")
					canApprove = false
					canReject = false
				}
			} else if currentTask.AssignedRole != nil {
				// PRIORITY 2: Check role-based permissions (when task is assigned to a role, not a specific user)
				assignedRole := *currentTask.AssignedRole
				log.Printf("[DEBUG] Checking approval permission - User: %s, UserRole: %s, AssignedRole: %s", userID, user.Role, assignedRole)

				// Check if assignedRole is a UUID (custom organization role)
				if _, parseErr := uuid.Parse(assignedRole); parseErr == nil {
					// It's a UUID - check if user has this organization role
					log.Printf("[DEBUG] AssignedRole is a UUID, checking user_organization_roles table")
					var userOrgRole models.UserOrganizationRole
					if err := db.Where("user_id = ? AND organization_id = ? AND role_id = ? AND active = ?",
						userID, organizationID, assignedRole, true).First(&userOrgRole).Error; err == nil {
						log.Printf("[DEBUG] User has the exact organization role - canApprove = true")
						canApprove = true
						canReject = true
					} else {
						log.Printf("[DEBUG] User does NOT have the exact organization role: %v", err)

						// Fallback 1: Check if user has a built-in approver role
						for _, approverRole := range approverRoles {
							if strings.EqualFold(user.Role, approverRole) {
								log.Printf("[DEBUG] User has built-in approver role '%s' - canApprove = true", user.Role)
								canApprove = true
								canReject = true
								break
							}
						}

						// Fallback 2: Check if user has any organization role with approval permissions
						if !canApprove && checkOrgRoleApprovalPermissions() {
							canApprove = true
							canReject = true
						}
					}
				} else {
					// It's a built-in role name - check user.Role directly (case-insensitive)
					log.Printf("[DEBUG] AssignedRole is a built-in role name, comparing with user.Role")
					if strings.EqualFold(user.Role, assignedRole) {
						log.Printf("[DEBUG] User role matches (case-insensitive) - canApprove = true")
						canApprove = true
						canReject = true
					} else {
						// Fallback 1: Check if user has a built-in approver role
						for _, approverRole := range approverRoles {
							if strings.EqualFold(user.Role, approverRole) {
								log.Printf("[DEBUG] User has built-in approver role '%s' - canApprove = true", user.Role)
								canApprove = true
								canReject = true
								break
							}
						}

						// Fallback 2: Check if user has any organization role with approval permissions
						if !canApprove && checkOrgRoleApprovalPermissions() {
							canApprove = true
							canReject = true
						}

						if !canApprove {
							log.Printf("[DEBUG] User role '%s' does not match assigned role '%s' and has no approval permissions", user.Role, assignedRole)
						}
					}
				}
			}
		}

		// Get next approver name
		// PRIORITY 1: If task is assigned to a specific user, show that user's name
		if currentTask.AssignedUserID != nil && *currentTask.AssignedUserID != "" {
			var assignedUser models.User
			if err := db.Where("id = ?", *currentTask.AssignedUserID).First(&assignedUser).Error; err == nil {
				nextApprover = assignedUser.Name
			} else {
				nextApprover = "Assigned User"
			}
		} else if currentTask.AssignedRole != nil {
			// PRIORITY 2: Role-based assignment
			assignedRole := *currentTask.AssignedRole
			roleDisplayName := assignedRole

			// Check if assignedRole is a UUID (custom organization role)
			if _, parseErr := uuid.Parse(assignedRole); parseErr == nil {
				// It's a UUID - look up the organization role name
				var orgRole models.OrganizationRole
				if err := db.Where("id = ?", assignedRole).First(&orgRole).Error; err == nil {
					roleDisplayName = orgRole.Name
				}

				// Find a user with this custom role
				var userOrgRole models.UserOrganizationRole
				if err := db.Preload("User").Where("organization_id = ? AND role_id = ? AND active = ?",
					organizationID, assignedRole, true).First(&userOrgRole).Error; err == nil && userOrgRole.User != nil {
					nextApprover = userOrgRole.User.Name
				} else {
					nextApprover = fmt.Sprintf("Any %s", roleDisplayName)
				}
			} else {
				// It's a built-in role name - use existing logic
				var approver models.User
				if err := db.Where("current_organization_id = ? AND role = ? AND active = ?",
					organizationID, assignedRole, true).First(&approver).Error; err == nil {
					nextApprover = approver.Name
				} else {
					nextApprover = fmt.Sprintf("Any %s", roleDisplayName)
				}
			}
		}
	}

	// Update the workflow status response with user permissions
	workflowStatus.CanApprove = canApprove
	workflowStatus.CanReject = canReject
	if nextApprover != "" {
		workflowStatus.NextApprover = nextApprover
	}

	return c.JSON(types.DetailResponse{
		Success: true,
		Data:    workflowStatus,
	})
}

// GetAvailableApprovers retrieves available approvers for a document type and stage
// GET /api/v1/approvals/available-approvers?documentType=...&stage=...
func (h *ApprovalHandler) GetAvailableApprovers(c *fiber.Ctx) error {
	organizationIDInterface := c.Locals("organizationID")
	if organizationIDInterface == nil {
		return utils.SendUnauthorizedError(c, "Organization ID not found in context")
	}
	
	organizationID, ok := organizationIDInterface.(string)
	if !ok {
		return utils.SendUnauthorizedError(c, "Invalid organization ID in context")
	}
	
	documentType := c.Query("documentType")
	entityID := c.Query("entityId") // Optional: specific entity ID to get workflow-specific approvers

	if documentType == "" {
		return utils.SendBadRequestError(c, "Document type is required")
	}

	db := config.DB

	// If entityId is provided, try to get workflow-specific approvers
	if entityID != "" {
		workflowExecutionService := c.Locals("workflowExecutionService").(*services.WorkflowExecutionService)
		
		workflowApprovers, err := workflowExecutionService.GetAvailableApproversForWorkflow(c.Context(), organizationID, entityID)
		if err == nil && len(workflowApprovers) > 0 {
			return utils.SendSuccess(c, fiber.StatusOK, workflowApprovers, "Available approvers retrieved successfully", nil)
		}
		// If workflow approvers not found, fall back to role-based approach
	}

	// Fallback to role-based approach for document type
	var roleFilters []string
	switch documentType {
	case "REQUISITION", "requisition":
		roleFilters = []string{"manager", "supervisor", "department_head", "finance"}
	case "PURCHASE_ORDER", "purchase_order":
		roleFilters = []string{"procurement", "finance", "admin"}
	case "PAYMENT_VOUCHER", "payment_voucher":
		roleFilters = []string{"finance", "accountant", "admin"}
	case "BUDGET", "budget":
		roleFilters = []string{"finance", "admin", "executive"}
	default:
		roleFilters = []string{"manager", "admin"}
	}

	// Execute query
	var approvers []ApproverInfo

	queryErr := db.Table("users").
		Select("users.id, users.name, users.email, users.role").
		Where("users.current_organization_id = ? AND users.active = ?", organizationID, true).
		Where("users.role IN ?", roleFilters).
		Find(&approvers).Error
		
	if queryErr != nil {
		log.Printf("Error fetching available approvers: %v", queryErr)
		return utils.SendInternalError(c, "Failed to fetch available approvers", queryErr)
	}

	return utils.SendSuccess(c, fiber.StatusOK, approvers, "Available approvers retrieved successfully", nil)
}

// Helper function for string contains check
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}