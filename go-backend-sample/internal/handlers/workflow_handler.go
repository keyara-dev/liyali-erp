package handlers

import (
	"strconv"

	"github.com/cozyCodr/liyali-gateway/internal/db"
	"github.com/cozyCodr/liyali-gateway/internal/repository"
	"github.com/cozyCodr/liyali-gateway/internal/utils"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type WorkflowHandler struct {
	workflowRepo repository.WorkflowRepositoryInterface
}

func NewWorkflowHandler(workflowRepo repository.WorkflowRepositoryInterface) *WorkflowHandler {
	return &WorkflowHandler{
		workflowRepo: workflowRepo,
	}
}

// Request/Response Types
type CreateWorkflowRequest struct {
	Name         string          `json:"name" validate:"required"`
	Description  string          `json:"description"`
	DocumentType string          `json:"documentType" validate:"required"`
	Stages       []byte          `json:"stages" validate:"required"` // JSONB
	IsActive     bool            `json:"isActive"`
}

type UpdateWorkflowRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Stages      []byte `json:"stages"`
}

// GetWorkflows retrieves all workflows with optional filtering
// GET /api/workflows
func (h *WorkflowHandler) GetWorkflows(c fiber.Ctx) error {
	// Get query parameters
	documentType := c.Query("documentType", "")
	activeOnlyStr := c.Query("activeOnly", "false")
	limitStr := c.Query("limit", "20")
	offsetStr := c.Query("offset", "0")

	activeOnly := activeOnlyStr == "true"
	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	var workflows []db.Workflow
	var err error

	// Filter based on query parameters
	if documentType != "" && activeOnly {
		workflows, err = h.workflowRepo.ListActiveWorkflowsByDocumentType(c.Context(), documentType, int32(limit), int32(offset))
	} else if documentType != "" {
		workflows, err = h.workflowRepo.ListWorkflowsByDocumentType(c.Context(), documentType, int32(limit), int32(offset))
	} else if activeOnly {
		workflows, err = h.workflowRepo.ListActiveWorkflows(c.Context(), int32(limit), int32(offset))
	} else {
		workflows, err = h.workflowRepo.ListWorkflows(c.Context(), int32(limit), int32(offset))
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve workflows",
		})
	}

	// Get total count
	var total int64
	if documentType != "" {
		total, _ = h.workflowRepo.CountWorkflowsByDocumentType(c.Context(), documentType)
	} else {
		total, _ = h.workflowRepo.CountWorkflows(c.Context())
	}

	return c.JSON(fiber.Map{
		"workflows": workflows,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
	})
}

// GetWorkflowByID retrieves a single workflow by ID
// GET /api/workflows/:id
func (h *WorkflowHandler) GetWorkflowByID(c fiber.Ctx) error {
	// Get workflow ID from params
	workflowIDStr := c.Params("id")
	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid workflow ID",
		})
	}

	// Get workflow
	workflow, err := h.workflowRepo.GetWorkflowByID(c.Context(), workflowID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Workflow not found",
		})
	}

	return c.JSON(fiber.Map{
		"workflow": workflow,
	})
}

// GetDefaultWorkflow retrieves the default workflow for a document type
// GET /api/workflows/default/:documentType
func (h *WorkflowHandler) GetDefaultWorkflow(c fiber.Ctx) error {
	// Get document type from params
	documentType := c.Params("documentType")
	if documentType == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Document type is required",
		})
	}

	// Get default workflow
	workflow, err := h.workflowRepo.GetDefaultWorkflowByDocumentType(c.Context(), documentType)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "No default workflow found for this document type",
		})
	}

	return c.JSON(fiber.Map{
		"workflow": workflow,
	})
}

// CreateWorkflow creates a new workflow
// POST /api/workflows
func (h *WorkflowHandler) CreateWorkflow(c fiber.Ctx) error {
	// Get user ID from context
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	// Parse request body
	var req CreateWorkflowRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate request
	if req.Name == "" || req.DocumentType == "" || req.Stages == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Name, document type, and stages are required",
		})
	}

	// Validate document type
	validTypes := map[string]bool{
		"REQUISITION":     true,
		"BUDGET":          true,
		"PURCHASE_ORDER":  true,
		"PAYMENT_VOUCHER": true,
		"GRN":             true,
	}
	if !validTypes[req.DocumentType] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid document type",
		})
	}

	// Create workflow
	workflow, err := h.workflowRepo.CreateWorkflow(c.Context(), db.CreateWorkflowParams{
		Name:         req.Name,
		Description:  pgtype.Text{String: req.Description, Valid: req.Description != ""},
		DocumentType: req.DocumentType,
		Stages:       req.Stages,
		IsActive:     pgtype.Bool{Bool: req.IsActive, Valid: true},
		CreatedBy:    utils.UUIDToPgtype(userID),
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create workflow",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":  "Workflow created successfully",
		"workflow": workflow,
	})
}

// UpdateWorkflow updates an existing workflow
// PUT /api/workflows/:id
func (h *WorkflowHandler) UpdateWorkflow(c fiber.Ctx) error {
	// Get workflow ID from params
	workflowIDStr := c.Params("id")
	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid workflow ID",
		})
	}

	// Parse request body
	var req UpdateWorkflowRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Get existing workflow to preserve fields not being updated
	existing, err := h.workflowRepo.GetWorkflowByID(c.Context(), workflowID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Workflow not found",
		})
	}

	// Prepare update params
	name := req.Name
	if name == "" {
		name = existing.Name
	}

	description := req.Description
	if description == "" {
		description = utils.PgtypeToString(existing.Description)
	}

	stages := req.Stages
	if stages == nil {
		stages = existing.Stages
	}

	// Update workflow
	workflow, err := h.workflowRepo.UpdateWorkflow(c.Context(), db.UpdateWorkflowParams{
		ID:          utils.UUIDToPgtype(workflowID),
		Name:        name,
		Description: pgtype.Text{String: description, Valid: description != ""},
		Stages:      stages,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update workflow",
		})
	}

	return c.JSON(fiber.Map{
		"message":  "Workflow updated successfully",
		"workflow": workflow,
	})
}

// ActivateWorkflow activates a workflow
// POST /api/workflows/:id/activate
func (h *WorkflowHandler) ActivateWorkflow(c fiber.Ctx) error {
	// Get workflow ID from params
	workflowIDStr := c.Params("id")
	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid workflow ID",
		})
	}

	// Activate workflow
	workflow, err := h.workflowRepo.ActivateWorkflow(c.Context(), workflowID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to activate workflow",
		})
	}

	return c.JSON(fiber.Map{
		"message":  "Workflow activated successfully",
		"workflow": workflow,
	})
}

// DeactivateWorkflow deactivates a workflow
// POST /api/workflows/:id/deactivate
func (h *WorkflowHandler) DeactivateWorkflow(c fiber.Ctx) error {
	// Get workflow ID from params
	workflowIDStr := c.Params("id")
	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid workflow ID",
		})
	}

	// Deactivate workflow
	workflow, err := h.workflowRepo.DeactivateWorkflow(c.Context(), workflowID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to deactivate workflow",
		})
	}

	return c.JSON(fiber.Map{
		"message":  "Workflow deactivated successfully",
		"workflow": workflow,
	})
}

// DeleteWorkflow deletes a workflow
// DELETE /api/workflows/:id
func (h *WorkflowHandler) DeleteWorkflow(c fiber.Ctx) error {
	// Get workflow ID from params
	workflowIDStr := c.Params("id")
	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid workflow ID",
		})
	}

	// Delete workflow
	err = h.workflowRepo.DeleteWorkflow(c.Context(), workflowID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete workflow",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Workflow deleted successfully",
	})
}
