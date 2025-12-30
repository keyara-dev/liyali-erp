package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/repository"
	"gorm.io/datatypes"
)

// WorkflowService handles workflow business logic
type WorkflowService struct {
	workflowRepo repository.WorkflowRepositoryInterface
	auditService *AuditService
}

// WorkflowStage represents a single stage in a workflow
type WorkflowStage struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Approvers   []string `json:"approvers"`   // User IDs who can approve at this stage
	RequiredApprovals int `json:"requiredApprovals"` // Number of approvals needed
	Order       int      `json:"order"`       // Stage order
	Conditions  map[string]interface{} `json:"conditions,omitempty"` // Conditions for this stage
}

// CreateWorkflowRequest represents a workflow creation request
type CreateWorkflowRequest struct {
	Name         string          `json:"name" validate:"required"`
	Description  string          `json:"description"`
	DocumentType string          `json:"documentType" validate:"required"`
	Stages       []WorkflowStage `json:"stages" validate:"required"`
	IsActive     bool            `json:"isActive"`
}

// UpdateWorkflowRequest represents a workflow update request
type UpdateWorkflowRequest struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Stages      []WorkflowStage `json:"stages"`
}

// NewWorkflowService creates a new workflow service
func NewWorkflowService(workflowRepo repository.WorkflowRepositoryInterface, auditService *AuditService) *WorkflowService {
	return &WorkflowService{
		workflowRepo: workflowRepo,
		auditService: auditService,
	}
}

// CreateWorkflow creates a new workflow
func (s *WorkflowService) CreateWorkflow(ctx context.Context, organizationID, userID string, req CreateWorkflowRequest) (*models.Workflow, error) {
	// Validate document type
	validTypes := map[string]bool{
		"REQUISITION":     true,
		"BUDGET":          true,
		"PURCHASE_ORDER":  true,
		"PAYMENT_VOUCHER": true,
		"GRN":             true,
	}
	
	if !validTypes[req.DocumentType] {
		return nil, fmt.Errorf("invalid document type: %s", req.DocumentType)
	}

	// Validate stages
	if len(req.Stages) == 0 {
		return nil, fmt.Errorf("workflow must have at least one stage")
	}

	// Convert stages to JSON
	stagesJSON, err := json.Marshal(req.Stages)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal stages: %w", err)
	}

	// Create workflow
	workflow, err := s.workflowRepo.Create(ctx, organizationID, req.Name, req.Description, req.DocumentType, datatypes.JSON(stagesJSON), req.IsActive, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to create workflow: %w", err)
	}

	// Log audit event
	if s.auditService != nil {
		details := fmt.Sprintf("Created workflow '%s' for document type '%s' with %d stages", req.Name, req.DocumentType, len(req.Stages))
		s.auditService.LogEvent(ctx, userID, organizationID, "workflow_created", "workflow", workflow.ID.String(), details, "", "")
	}

	return workflow, nil
}

// GetWorkflow retrieves a workflow by ID
func (s *WorkflowService) GetWorkflow(ctx context.Context, id uuid.UUID, organizationID string) (*models.Workflow, error) {
	return s.workflowRepo.GetByID(ctx, id, organizationID)
}

// UpdateWorkflow updates a workflow
func (s *WorkflowService) UpdateWorkflow(ctx context.Context, id uuid.UUID, organizationID, userID string, req UpdateWorkflowRequest) (*models.Workflow, error) {
	// Get existing workflow
	existing, err := s.workflowRepo.GetByID(ctx, id, organizationID)
	if err != nil {
		return nil, fmt.Errorf("workflow not found: %w", err)
	}

	// Prepare update data
	name := req.Name
	if name == "" {
		name = existing.Name
	}

	description := req.Description
	if description == "" {
		description = existing.Description
	}

	var stagesJSON datatypes.JSON
	if req.Stages != nil {
		// Validate stages
		if len(req.Stages) == 0 {
			return nil, fmt.Errorf("workflow must have at least one stage")
		}

		stagesBytes, err := json.Marshal(req.Stages)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal stages: %w", err)
		}
		stagesJSON = datatypes.JSON(stagesBytes)
	} else {
		stagesJSON = existing.Stages
	}

	// Update workflow
	workflow, err := s.workflowRepo.Update(ctx, id, organizationID, name, description, stagesJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to update workflow: %w", err)
	}

	// Log audit event
	if s.auditService != nil {
		details := fmt.Sprintf("Updated workflow '%s'", workflow.Name)
		s.auditService.LogEvent(ctx, userID, organizationID, "workflow_updated", "workflow", workflow.ID.String(), details, "", "")
	}

	return workflow, nil
}

// DeleteWorkflow deletes a workflow
func (s *WorkflowService) DeleteWorkflow(ctx context.Context, id uuid.UUID, organizationID, userID string) error {
	// Get existing workflow for audit logging
	existing, err := s.workflowRepo.GetByID(ctx, id, organizationID)
	if err != nil {
		return fmt.Errorf("workflow not found: %w", err)
	}

	// Delete workflow
	if err := s.workflowRepo.Delete(ctx, id, organizationID); err != nil {
		return fmt.Errorf("failed to delete workflow: %w", err)
	}

	// Log audit event
	if s.auditService != nil {
		details := fmt.Sprintf("Deleted workflow '%s'", existing.Name)
		s.auditService.LogEvent(ctx, userID, organizationID, "workflow_deleted", "workflow", existing.ID.String(), details, "", "")
	}

	return nil
}

// ListWorkflows retrieves workflows with filtering and pagination
func (s *WorkflowService) ListWorkflows(ctx context.Context, organizationID string, documentType string, activeOnly bool, limit, offset int) ([]*models.Workflow, int64, error) {
	var workflows []*models.Workflow
	var total int64
	var err error

	// Apply filters
	if documentType != "" && activeOnly {
		workflows, err = s.workflowRepo.ListActiveByDocumentType(ctx, organizationID, documentType, limit, offset)
		if err == nil {
			total, err = s.workflowRepo.CountByDocumentType(ctx, organizationID, documentType)
		}
	} else if documentType != "" {
		workflows, err = s.workflowRepo.ListByDocumentType(ctx, organizationID, documentType, limit, offset)
		if err == nil {
			total, err = s.workflowRepo.CountByDocumentType(ctx, organizationID, documentType)
		}
	} else if activeOnly {
		workflows, err = s.workflowRepo.ListActive(ctx, organizationID, limit, offset)
		if err == nil {
			total, err = s.workflowRepo.CountActive(ctx, organizationID)
		}
	} else {
		workflows, err = s.workflowRepo.List(ctx, organizationID, limit, offset)
		if err == nil {
			total, err = s.workflowRepo.Count(ctx, organizationID)
		}
	}

	if err != nil {
		return nil, 0, fmt.Errorf("failed to list workflows: %w", err)
	}

	return workflows, total, nil
}

// GetDefaultWorkflow retrieves the default workflow for a document type
func (s *WorkflowService) GetDefaultWorkflow(ctx context.Context, organizationID, documentType string) (*models.Workflow, error) {
	workflow, err := s.workflowRepo.GetDefaultByDocumentType(ctx, organizationID, documentType)
	if err != nil {
		return nil, fmt.Errorf("no default workflow found for document type %s: %w", documentType, err)
	}

	return workflow, nil
}

// ActivateWorkflow activates a workflow
func (s *WorkflowService) ActivateWorkflow(ctx context.Context, id uuid.UUID, organizationID, userID string) (*models.Workflow, error) {
	workflow, err := s.workflowRepo.Activate(ctx, id, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to activate workflow: %w", err)
	}

	// Log audit event
	if s.auditService != nil {
		details := fmt.Sprintf("Activated workflow '%s'", workflow.Name)
		s.auditService.LogEvent(ctx, userID, organizationID, "workflow_activated", "workflow", workflow.ID.String(), details, "", "")
	}

	return workflow, nil
}

// DeactivateWorkflow deactivates a workflow
func (s *WorkflowService) DeactivateWorkflow(ctx context.Context, id uuid.UUID, organizationID, userID string) (*models.Workflow, error) {
	workflow, err := s.workflowRepo.Deactivate(ctx, id, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to deactivate workflow: %w", err)
	}

	// Log audit event
	if s.auditService != nil {
		details := fmt.Sprintf("Deactivated workflow '%s'", workflow.Name)
		s.auditService.LogEvent(ctx, userID, organizationID, "workflow_deactivated", "workflow", workflow.ID.String(), details, "", "")
	}

	return workflow, nil
}

// GetWorkflowStages parses and returns the stages from a workflow
func (s *WorkflowService) GetWorkflowStages(workflow *models.Workflow) ([]WorkflowStage, error) {
	var stages []WorkflowStage
	if err := json.Unmarshal(workflow.Stages, &stages); err != nil {
		return nil, fmt.Errorf("failed to parse workflow stages: %w", err)
	}
	return stages, nil
}

// ValidateWorkflowStages validates workflow stages
func (s *WorkflowService) ValidateWorkflowStages(stages []WorkflowStage) error {
	if len(stages) == 0 {
		return fmt.Errorf("workflow must have at least one stage")
	}

	orderMap := make(map[int]bool)
	for _, stage := range stages {
		if stage.Name == "" {
			return fmt.Errorf("stage name is required")
		}
		if len(stage.Approvers) == 0 {
			return fmt.Errorf("stage '%s' must have at least one approver", stage.Name)
		}
		if stage.RequiredApprovals <= 0 {
			return fmt.Errorf("stage '%s' must require at least one approval", stage.Name)
		}
		if stage.RequiredApprovals > len(stage.Approvers) {
			return fmt.Errorf("stage '%s' requires more approvals than available approvers", stage.Name)
		}
		if orderMap[stage.Order] {
			return fmt.Errorf("duplicate stage order: %d", stage.Order)
		}
		orderMap[stage.Order] = true
	}

	return nil
}