package handlers

import (
	"log"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/services"
	"github.com/liyali/liyali-gateway/utils"
)

type WorkflowHandler struct {
	workflowService *services.WorkflowService
	validate        *validator.Validate
}

func NewWorkflowHandler(workflowService *services.WorkflowService) *WorkflowHandler {
	return &WorkflowHandler{
		workflowService: workflowService,
		validate:        validator.New(),
	}
}

// GetWorkflows retrieves all workflows with optional filtering
// GET /api/v1/workflows
func (h *WorkflowHandler) GetWorkflows(c *fiber.Ctx) error {
	organizationID := c.Locals("organization_id").(string)

	// Get query parameters
	documentType := c.Query("documentType", "")
	activeOnlyStr := c.Query("activeOnly", "false")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	activeOnly := activeOnlyStr == "true"
	if page < 1 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Get workflows
	workflows, total, err := h.workflowService.ListWorkflows(c.Context(), organizationID, documentType, activeOnly, limit, offset)
	if err != nil {
		log.Printf("Error fetching workflows: %v", err)
		return utils.SendInternalError(c, "Failed to retrieve workflows", err)
	}

	return utils.SendPaginatedSuccess(c, workflows, "Workflows retrieved successfully", page, limit, total)
}

// GetWorkflowByID retrieves a single workflow by ID
// GET /api/v1/workflows/:id
func (h *WorkflowHandler) GetWorkflowByID(c *fiber.Ctx) error {
	organizationID := c.Locals("organization_id").(string)

	// Get workflow ID from params
	workflowIDStr := c.Params("id")
	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		return utils.SendBadRequestError(c, "Invalid workflow ID")
	}

	// Get workflow
	workflow, err := h.workflowService.GetWorkflow(c.Context(), workflowID, organizationID)
	if err != nil {
		log.Printf("Error fetching workflow %s: %v", workflowID, err)
		return utils.SendNotFoundError(c, "Workflow not found")
	}

	return utils.SendSimpleSuccess(c, workflow, "Workflow retrieved successfully")
}

// GetDefaultWorkflow retrieves the default workflow for a document type
// GET /api/v1/workflows/default/:documentType
func (h *WorkflowHandler) GetDefaultWorkflow(c *fiber.Ctx) error {
	organizationID := c.Locals("organization_id").(string)

	// Get document type from params
	documentType := c.Params("documentType")
	if documentType == "" {
		return utils.SendBadRequestError(c, "Document type is required")
	}

	// Get default workflow
	workflow, err := h.workflowService.GetDefaultWorkflow(c.Context(), organizationID, documentType)
	if err != nil {
		log.Printf("Error fetching default workflow for %s: %v", documentType, err)
		return utils.SendNotFoundError(c, "No default workflow found for this document type")
	}

	return utils.SendSimpleSuccess(c, workflow, "Default workflow retrieved successfully")
}

// CreateWorkflow creates a new workflow
// POST /api/v1/workflows
func (h *WorkflowHandler) CreateWorkflow(c *fiber.Ctx) error {
	organizationID := c.Locals("organization_id").(string)
	userID := c.Locals("user_id").(string)

	// Parse request body
	var req services.CreateWorkflowRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendBadRequestError(c, "Invalid request body")
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		return utils.SendBadRequestError(c, "Validation failed: "+err.Error())
	}

	// Validate workflow stages
	if err := h.workflowService.ValidateWorkflowStages(req.Stages); err != nil {
		return utils.SendBadRequestError(c, "Invalid workflow stages: "+err.Error())
	}

	// Create workflow
	workflow, err := h.workflowService.CreateWorkflow(c.Context(), organizationID, userID, req)
	if err != nil {
		log.Printf("Error creating workflow: %v", err)
		return utils.SendInternalError(c, "Failed to create workflow", err)
	}

	return utils.SendCreatedSuccess(c, workflow, "Workflow created successfully")
}

// UpdateWorkflow updates an existing workflow
// PUT /api/v1/workflows/:id
func (h *WorkflowHandler) UpdateWorkflow(c *fiber.Ctx) error {
	organizationID := c.Locals("organization_id").(string)
	userID := c.Locals("user_id").(string)

	// Get workflow ID from params
	workflowIDStr := c.Params("id")
	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		return utils.SendBadRequestError(c, "Invalid workflow ID")
	}

	// Parse request body
	var req services.UpdateWorkflowRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendBadRequestError(c, "Invalid request body")
	}

	// Validate workflow stages if provided
	if req.Stages != nil {
		if err := h.workflowService.ValidateWorkflowStages(req.Stages); err != nil {
			return utils.SendBadRequestError(c, "Invalid workflow stages: "+err.Error())
		}
	}

	// Update workflow
	workflow, err := h.workflowService.UpdateWorkflow(c.Context(), workflowID, organizationID, userID, req)
	if err != nil {
		log.Printf("Error updating workflow %s: %v", workflowID, err)
		return utils.SendInternalError(c, "Failed to update workflow", err)
	}

	return utils.SendSimpleSuccess(c, workflow, "Workflow updated successfully")
}

// ActivateWorkflow activates a workflow
// POST /api/v1/workflows/:id/activate
func (h *WorkflowHandler) ActivateWorkflow(c *fiber.Ctx) error {
	organizationID := c.Locals("organization_id").(string)
	userID := c.Locals("user_id").(string)

	// Get workflow ID from params
	workflowIDStr := c.Params("id")
	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		return utils.SendBadRequestError(c, "Invalid workflow ID")
	}

	// Activate workflow
	workflow, err := h.workflowService.ActivateWorkflow(c.Context(), workflowID, organizationID, userID)
	if err != nil {
		log.Printf("Error activating workflow %s: %v", workflowID, err)
		return utils.SendInternalError(c, "Failed to activate workflow", err)
	}

	return utils.SendSimpleSuccess(c, workflow, "Workflow activated successfully")
}

// DeactivateWorkflow deactivates a workflow
// POST /api/v1/workflows/:id/deactivate
func (h *WorkflowHandler) DeactivateWorkflow(c *fiber.Ctx) error {
	organizationID := c.Locals("organization_id").(string)
	userID := c.Locals("user_id").(string)

	// Get workflow ID from params
	workflowIDStr := c.Params("id")
	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		return utils.SendBadRequestError(c, "Invalid workflow ID")
	}

	// Deactivate workflow
	workflow, err := h.workflowService.DeactivateWorkflow(c.Context(), workflowID, organizationID, userID)
	if err != nil {
		log.Printf("Error deactivating workflow %s: %v", workflowID, err)
		return utils.SendInternalError(c, "Failed to deactivate workflow", err)
	}

	return utils.SendSimpleSuccess(c, workflow, "Workflow deactivated successfully")
}

// DeleteWorkflow deletes a workflow
// DELETE /api/v1/workflows/:id
func (h *WorkflowHandler) DeleteWorkflow(c *fiber.Ctx) error {
	organizationID := c.Locals("organization_id").(string)
	userID := c.Locals("user_id").(string)

	// Get workflow ID from params
	workflowIDStr := c.Params("id")
	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		return utils.SendBadRequestError(c, "Invalid workflow ID")
	}

	// Delete workflow
	if err := h.workflowService.DeleteWorkflow(c.Context(), workflowID, organizationID, userID); err != nil {
		log.Printf("Error deleting workflow %s: %v", workflowID, err)
		return utils.SendInternalError(c, "Failed to delete workflow", err)
	}

	return utils.SendSimpleSuccess(c, nil, "Workflow deleted successfully")
}