package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/logging"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/repository"
	"gorm.io/gorm"
)

// WorkflowService handles workflow business logic
type WorkflowService struct {
	workflowRepo repository.WorkflowRepositoryInterface
	auditService *AuditService
	db           *gorm.DB
}

// CreateWorkflowRequest represents a workflow creation request
type CreateWorkflowRequest struct {
	Name        string                     `json:"name" validate:"required"`
	Description string                     `json:"description"`
	EntityType  string                     `json:"entityType" validate:"required"` // Changed from DocumentType
	Stages      []models.WorkflowStage     `json:"stages" validate:"required"`
	Conditions  *models.WorkflowConditions `json:"conditions"`
	IsDefault   bool                       `json:"isDefault"`
}

// UpdateWorkflowRequest represents a workflow update request
type UpdateWorkflowRequest struct {
	Name        *string                    `json:"name"`
	Description *string                    `json:"description"`
	Stages      []models.WorkflowStage     `json:"stages"`
	Conditions  *models.WorkflowConditions `json:"conditions"`
	IsDefault   *bool                      `json:"isDefault"`
}

// WorkflowListFilter represents filters for listing workflows
type WorkflowListFilter struct {
	EntityType string `json:"entityType"`
	IsActive   *bool  `json:"isActive"`
	IsDefault  *bool  `json:"isDefault"`
}

// NewWorkflowService creates a new workflow service
func NewWorkflowService(workflowRepo repository.WorkflowRepositoryInterface, auditService *AuditService, db *gorm.DB) *WorkflowService {
	return &WorkflowService{
		workflowRepo: workflowRepo,
		auditService: auditService,
		db:           db,
	}
}

// CreateWorkflow creates a new workflow
func (s *WorkflowService) CreateWorkflow(ctx context.Context, organizationID, userID string, req CreateWorkflowRequest) (*models.Workflow, error) {
	// Validate request
	if err := s.validateCreateRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// If this is set as default, unset other defaults for the same entity type
	if req.IsDefault {
		if err := s.unsetDefaultWorkflows(tx, organizationID, req.EntityType); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to unset existing defaults: %w", err)
		}
	}

	// Create workflow
	workflow := &models.Workflow{
		ID:             uuid.New(),
		OrganizationID: organizationID,
		Name:           req.Name,
		Description:    req.Description,
		DocumentType:   req.EntityType, // Set both for compatibility
		EntityType:     req.EntityType,
		Version:        1,
		IsActive:       true,
		IsDefault:      req.IsDefault,
		CreatedBy:      userID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Set stages
	if err := workflow.SetStages(req.Stages); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to set stages: %w", err)
	}

	// Set conditions
	if err := workflow.SetConditions(req.Conditions); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to set conditions: %w", err)
	}

	// Save workflow
	if err := tx.Create(workflow).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create workflow: %w", err)
	}

	// Create default workflow record if needed
	if req.IsDefault {
		defaultRecord := &models.WorkflowDefault{
			ID:                     uuid.New().String(),
			OrganizationID:         organizationID,
			EntityType:             req.EntityType,
			DefaultWorkflowID:      workflow.ID,
			DefaultWorkflowVersion: workflow.Version,
			SetBy:                  userID,
			SetAt:                  time.Now(),
		}

		if err := tx.Create(defaultRecord).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to create default workflow record: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Load computed fields
	s.loadComputedFields(workflow)

	// Log audit event
	if s.auditService != nil {
		details := fmt.Sprintf("Created workflow '%s' for entity type '%s' with %d stages", req.Name, req.EntityType, len(req.Stages))
		s.auditService.LogEvent(ctx, userID, organizationID, "workflow_created", "workflow", workflow.ID.String(), details, "", "")
	}

	return workflow, nil
}

// GetWorkflow retrieves a workflow by ID
func (s *WorkflowService) GetWorkflow(ctx context.Context, id uuid.UUID, organizationID string) (*models.Workflow, error) {
	var workflow models.Workflow
	
	if err := s.db.Where("id = ? AND organization_id = ?", id.String(), organizationID).First(&workflow).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("workflow not found")
		}
		return nil, fmt.Errorf("failed to retrieve workflow: %w", err)
	}

	// Load computed fields
	s.loadComputedFields(&workflow)

	return &workflow, nil
}

// GetWorkflowByStringID retrieves a workflow by string ID (for frontend compatibility)
func (s *WorkflowService) GetWorkflowByStringID(ctx context.Context, id string, organizationID string) (*models.Workflow, error) {
	// Try to parse as UUID first
	workflowID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid workflow ID format")
	}
	
	return s.GetWorkflow(ctx, workflowID, organizationID)
}

// UpdateWorkflow updates an existing workflow (creates new version)
func (s *WorkflowService) UpdateWorkflow(ctx context.Context, id uuid.UUID, organizationID, userID string, req UpdateWorkflowRequest) (*models.Workflow, error) {
	// Get existing workflow
	existing, err := s.GetWorkflow(ctx, id, organizationID)
	if err != nil {
		return nil, err
	}

	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create new version
	newVersion := &models.Workflow{
		ID:             uuid.New(),
		OrganizationID: organizationID,
		Name:           existing.Name,
		Description:    existing.Description,
		DocumentType:   existing.EntityType, // Set both for compatibility
		EntityType:     existing.EntityType,
		Version:        existing.Version + 1,
		IsActive:       true,
		IsDefault:      existing.IsDefault,
		Conditions:     existing.Conditions,
		Stages:         existing.Stages,
		CreatedBy:      existing.CreatedBy,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Apply updates
	if req.Name != nil {
		newVersion.Name = *req.Name
	}
	if req.Description != nil {
		newVersion.Description = *req.Description
	}
	if req.Stages != nil {
		if err := newVersion.SetStages(req.Stages); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to set stages: %w", err)
		}
	}
	if req.Conditions != nil {
		if err := newVersion.SetConditions(req.Conditions); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to set conditions: %w", err)
		}
	}
	if req.IsDefault != nil {
		newVersion.IsDefault = *req.IsDefault

		// If setting as default, unset other defaults
		if *req.IsDefault {
			if err := s.unsetDefaultWorkflows(tx, organizationID, newVersion.EntityType); err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("failed to unset existing defaults: %w", err)
			}
		}
	}

	// Validate new version
	if err := newVersion.Validate(); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Deactivate old version
	if err := tx.Model(existing).Update("is_active", false).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to deactivate old version: %w", err)
	}

	// Save new version
	if err := tx.Create(newVersion).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create new version: %w", err)
	}

	// Update default workflow record if needed
	if newVersion.IsDefault {
		if err := s.updateDefaultWorkflow(tx, organizationID, newVersion); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to update default workflow: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Load computed fields
	s.loadComputedFields(newVersion)

	// Log audit event
	if s.auditService != nil {
		details := fmt.Sprintf("Updated workflow '%s'", newVersion.Name)
		s.auditService.LogEvent(ctx, userID, organizationID, "workflow_updated", "workflow", newVersion.ID.String(), details, "", "")
	}

	return newVersion, nil
}

// ListWorkflows retrieves workflows with filtering and pagination
func (s *WorkflowService) ListWorkflows(ctx context.Context, organizationID string, entityType string, activeOnly bool, limit, offset int) ([]*models.Workflow, int64, error) {
	var workflows []*models.Workflow
	var total int64

	query := s.db.Model(&models.Workflow{}).Where("organization_id = ?", organizationID)

	// Apply filters
	if entityType != "" {
		query = query.Where("entity_type = ?", entityType)
	}
	if activeOnly {
		query = query.Where("is_active = ?", true)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count workflows: %w", err)
	}

	// Get workflows with pagination
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&workflows).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list workflows: %w", err)
	}

	// Load computed fields for each workflow
	for _, workflow := range workflows {
		s.loadComputedFields(workflow)
	}

	return workflows, total, nil
}

// GetWorkflows retrieves workflows with optional filters (for frontend compatibility)
func (s *WorkflowService) GetWorkflows(ctx context.Context, organizationID string, filter WorkflowListFilter) ([]models.Workflow, error) {
	var workflows []models.Workflow

	query := s.db.Model(&models.Workflow{}).Where("organization_id = ?", organizationID)

	// Apply filters
	if filter.EntityType != "" {
		query = query.Where("entity_type = ?", filter.EntityType)
	}
	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}
	if filter.IsDefault != nil {
		query = query.Where("is_default = ?", *filter.IsDefault)
	}

	// Order by creation date (newest first)
	query = query.Order("created_at DESC")

	if err := query.Find(&workflows).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve workflows: %w", err)
	}

	// Load computed fields for each workflow
	for i := range workflows {
		s.loadComputedFields(&workflows[i])
	}

	return workflows, nil
}

// DeleteWorkflow deletes a workflow
func (s *WorkflowService) DeleteWorkflow(ctx context.Context, id uuid.UUID, organizationID, userID string) error {
	// Get existing workflow for audit logging
	existing, err := s.GetWorkflow(ctx, id, organizationID)
	if err != nil {
		return fmt.Errorf("workflow not found: %w", err)
	}

	// Check if workflow is in use
	var assignmentCount int64
	if err := s.db.Model(&models.WorkflowAssignment{}).
		Where("workflow_id = ? AND organization_id = ?", existing.ID, organizationID).
		Count(&assignmentCount).Error; err != nil {
		return fmt.Errorf("failed to check workflow usage: %w", err)
	}

	if assignmentCount > 0 {
		return fmt.Errorf("cannot delete workflow: it is currently in use by %d assignments", assignmentCount)
	}

	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Soft delete workflow
	now := time.Now()
	if err := tx.Model(existing).Update("deleted_at", &now).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete workflow: %w", err)
	}

	// Delete default workflow record if it exists
	if err := tx.Where("default_workflow_id = ? AND organization_id = ?", existing.ID, organizationID).
		Delete(&models.WorkflowDefault{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete default workflow record: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Log audit event
	if s.auditService != nil {
		details := fmt.Sprintf("Deleted workflow '%s'", existing.Name)
		s.auditService.LogEvent(ctx, userID, organizationID, "workflow_deleted", "workflow", existing.ID.String(), details, "", "")
	}

	return nil
}

// GetDefaultWorkflow retrieves the default workflow for an entity type
func (s *WorkflowService) GetDefaultWorkflow(ctx context.Context, organizationID, entityType string) (*models.Workflow, error) {
	var defaultRecord models.WorkflowDefault

	if err := s.db.Where("organization_id = ? AND entity_type = ?", organizationID, entityType).
		First(&defaultRecord).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no default workflow found for entity type: %s", entityType)
		}
		return nil, fmt.Errorf("failed to retrieve default workflow: %w", err)
	}

	return s.GetWorkflow(ctx, defaultRecord.DefaultWorkflowID, organizationID)
}

// ActivateWorkflow activates a workflow
func (s *WorkflowService) ActivateWorkflow(ctx context.Context, id uuid.UUID, organizationID, userID string) (*models.Workflow, error) {
	workflow, err := s.GetWorkflow(ctx, id, organizationID)
	if err != nil {
		return nil, err
	}

	if err := s.db.Model(workflow).Update("is_active", true).Error; err != nil {
		return nil, fmt.Errorf("failed to activate workflow: %w", err)
	}

	workflow.IsActive = true

	// Log audit event
	if s.auditService != nil {
		details := fmt.Sprintf("Activated workflow '%s'", workflow.Name)
		s.auditService.LogEvent(ctx, userID, organizationID, "workflow_activated", "workflow", workflow.ID.String(), details, "", "")
	}

	return workflow, nil
}

// DeactivateWorkflow deactivates a workflow
func (s *WorkflowService) DeactivateWorkflow(ctx context.Context, id uuid.UUID, organizationID, userID string) (*models.Workflow, error) {
	workflow, err := s.GetWorkflow(ctx, id, organizationID)
	if err != nil {
		return nil, err
	}

	if err := s.db.Model(workflow).Update("is_active", false).Error; err != nil {
		return nil, fmt.Errorf("failed to deactivate workflow: %w", err)
	}

	workflow.IsActive = false

	// Log audit event
	if s.auditService != nil {
		details := fmt.Sprintf("Deactivated workflow '%s'", workflow.Name)
		s.auditService.LogEvent(ctx, userID, organizationID, "workflow_deactivated", "workflow", workflow.ID.String(), details, "", "")
	}

	return workflow, nil
}

// DuplicateWorkflow creates a copy of an existing workflow
func (s *WorkflowService) DuplicateWorkflow(ctx context.Context, id uuid.UUID, organizationID, userID, newName string) (*models.Workflow, error) {
	// Get existing workflow
	existing, err := s.GetWorkflow(ctx, id, organizationID)
	if err != nil {
		return nil, err
	}

	// Get stages and conditions
	stages, err := existing.GetStages()
	if err != nil {
		return nil, fmt.Errorf("failed to get stages: %w", err)
	}

	conditions, err := existing.GetConditions()
	if err != nil {
		return nil, fmt.Errorf("failed to get conditions: %w", err)
	}

	// Create duplicate
	req := CreateWorkflowRequest{
		Name:        newName,
		Description: existing.Description + " (Copy)",
		EntityType:  existing.EntityType,
		Stages:      stages,
		Conditions:  conditions,
		IsDefault:   false, // Duplicates are never default
	}

	return s.CreateWorkflow(ctx, organizationID, userID, req)
}

// SetDefaultWorkflow sets a workflow as the default for an entity type
func (s *WorkflowService) SetDefaultWorkflow(ctx context.Context, organizationID, entityType, workflowId, userID string) error {
	// Parse workflow ID as UUID
	workflowUUID, err := uuid.Parse(workflowId)
	if err != nil {
		return fmt.Errorf("invalid workflow ID format")
	}
	
	// Verify workflow exists and is active
	workflow, err := s.GetWorkflow(ctx, workflowUUID, organizationID)
	if err != nil {
		return err
	}

	if !workflow.IsActive {
		return fmt.Errorf("cannot set inactive workflow as default")
	}

	if workflow.EntityType != entityType {
		return fmt.Errorf("workflow entity type mismatch")
	}

	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Unset existing defaults
	if err := s.unsetDefaultWorkflows(tx, organizationID, entityType); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to unset existing defaults: %w", err)
	}

	// Update workflow to be default
	if err := tx.Model(workflow).Update("is_default", true).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update workflow: %w", err)
	}

	// Create/update default workflow record
	defaultRecord := &models.WorkflowDefault{
		ID:                     uuid.New().String(),
		OrganizationID:         organizationID,
		EntityType:             entityType,
		DefaultWorkflowID:      workflowUUID,
		DefaultWorkflowVersion: workflow.Version,
		SetBy:                  userID,
		SetAt:                  time.Now(),
	}

	if err := tx.Create(defaultRecord).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create default workflow record: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ResolveWorkflowForEntity finds the appropriate workflow for an entity
func (s *WorkflowService) ResolveWorkflowForEntity(ctx context.Context, organizationID, entityType string, document interface{}) (*models.Workflow, error) {
	// Get all active workflows for the entity type
	workflows, err := s.GetWorkflows(ctx, organizationID, WorkflowListFilter{
		EntityType: entityType,
		IsActive:   &[]bool{true}[0],
	})
	if err != nil {
		return nil, err
	}

	// Find matching workflow based on conditions
	for _, workflow := range workflows {
		conditions, err := workflow.GetConditions()
		if err != nil {
			continue // Skip workflows with invalid conditions
		}

		if conditions == nil || conditions.MatchesDocument(document) {
			return &workflow, nil
		}
	}

	// Fall back to default workflow
	return s.GetDefaultWorkflow(ctx, organizationID, entityType)
}

// GetWorkflowStages parses and returns the stages from a workflow
func (s *WorkflowService) GetWorkflowStages(workflow *models.Workflow) ([]models.WorkflowStage, error) {
	return workflow.GetStages()
}

// ValidateWorkflowStages validates workflow stages
func (s *WorkflowService) ValidateWorkflowStages(stages []models.WorkflowStage) error {
	if len(stages) == 0 {
		return fmt.Errorf("workflow must have at least one stage")
	}

	stageNumbers := make(map[int]bool)
	for i, stage := range stages {
		expectedNumber := i + 1
		if stage.StageNumber != expectedNumber {
			return fmt.Errorf("stage %d: stage number should be %d, got %d", i+1, expectedNumber, stage.StageNumber)
		}
		
		if err := stage.Validate(); err != nil {
			return fmt.Errorf("stage %d validation failed: %w", i+1, err)
		}
		
		if stageNumbers[stage.StageNumber] {
			return fmt.Errorf("duplicate stage number: %d", stage.StageNumber)
		}
		stageNumbers[stage.StageNumber] = true
	}

	return nil
}

// GetWorkflowUsageCount returns the number of times a workflow has been used
func (s *WorkflowService) GetWorkflowUsageCount(ctx context.Context, organizationID, workflowId string) (int64, error) {
	// Parse workflow ID as UUID
	workflowUUID, err := uuid.Parse(workflowId)
	if err != nil {
		return 0, fmt.Errorf("invalid workflow ID format")
	}
	
	var count int64
	err = s.db.Model(&models.WorkflowAssignment{}).
		Where("workflow_id = ? AND organization_id = ?", workflowUUID, organizationID).
		Count(&count).Error
	return count, err
}

// Helper methods
func (s *WorkflowService) validateCreateRequest(req CreateWorkflowRequest) error {
	if req.Name == "" {
		return fmt.Errorf("workflow name is required")
	}
	if req.EntityType == "" {
		return fmt.Errorf("entity type is required")
	}
	if len(req.Stages) == 0 {
		return fmt.Errorf("workflow must have at least one stage")
	}

	// Validate stages
	return s.ValidateWorkflowStages(req.Stages)
}

func (s *WorkflowService) unsetDefaultWorkflows(tx *gorm.DB, organizationID, entityType string) error {
	// Update all workflows of this entity type to not be default
	if err := tx.Model(&models.Workflow{}).
		Where("organization_id = ? AND entity_type = ? AND is_default = ?", organizationID, entityType, true).
		Update("is_default", false).Error; err != nil {
		return err
	}

	// Delete existing default workflow records
	return tx.Where("organization_id = ? AND entity_type = ?", organizationID, entityType).
		Delete(&models.WorkflowDefault{}).Error
}

func (s *WorkflowService) updateDefaultWorkflow(tx *gorm.DB, organizationID string, workflow *models.Workflow) error {
	// Delete existing default record
	if err := tx.Where("organization_id = ? AND entity_type = ?", organizationID, workflow.EntityType).
		Delete(&models.WorkflowDefault{}).Error; err != nil {
		return err
	}

	// Create new default record
	defaultRecord := &models.WorkflowDefault{
		ID:                     uuid.New().String(),
		OrganizationID:         organizationID,
		EntityType:             workflow.EntityType,
		DefaultWorkflowID:      workflow.ID,
		DefaultWorkflowVersion: workflow.Version,
		SetBy:                  workflow.CreatedBy,
		SetAt:                  time.Now(),
	}

	return tx.Create(defaultRecord).Error
}

func (s *WorkflowService) loadComputedFields(workflow *models.Workflow) {
	// Load total stages
	stages, err := workflow.GetStages()
	if err == nil {
		workflow.TotalStages = len(stages)
	}

	// Load usage count
	var count int64
	s.db.Model(&models.WorkflowAssignment{}).
		Where("workflow_id = ? AND organization_id = ?", workflow.ID, workflow.OrganizationID).
		Count(&count)
	workflow.UsageCount = int(count)
}