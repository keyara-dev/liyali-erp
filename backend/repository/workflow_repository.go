package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/liyali/liyali-gateway/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// WorkflowRepositoryInterface defines the contract for workflow repository
type WorkflowRepositoryInterface interface {
	// Basic CRUD operations
	Create(ctx context.Context, organizationID, name, description, documentType string, stages datatypes.JSON, isActive bool, createdBy string) (*models.Workflow, error)
	GetByID(ctx context.Context, id uuid.UUID, organizationID string) (*models.Workflow, error)
	Update(ctx context.Context, id uuid.UUID, organizationID, name, description string, stages datatypes.JSON) (*models.Workflow, error)
	Delete(ctx context.Context, id uuid.UUID, organizationID string) error
	
	// List operations
	List(ctx context.Context, organizationID string, limit, offset int) ([]*models.Workflow, error)
	ListActive(ctx context.Context, organizationID string, limit, offset int) ([]*models.Workflow, error)
	ListByDocumentType(ctx context.Context, organizationID, documentType string, limit, offset int) ([]*models.Workflow, error)
	ListActiveByDocumentType(ctx context.Context, organizationID, documentType string, limit, offset int) ([]*models.Workflow, error)
	
	// Special operations
	GetDefaultByDocumentType(ctx context.Context, organizationID, documentType string) (*models.Workflow, error)
	Activate(ctx context.Context, id uuid.UUID, organizationID string) (*models.Workflow, error)
	Deactivate(ctx context.Context, id uuid.UUID, organizationID string) (*models.Workflow, error)
	
	// Count operations
	Count(ctx context.Context, organizationID string) (int64, error)
	CountActive(ctx context.Context, organizationID string) (int64, error)
	CountByDocumentType(ctx context.Context, organizationID, documentType string) (int64, error)
}

// WorkflowRepository implements WorkflowRepositoryInterface
type WorkflowRepository struct {
	db    *gorm.DB
	pgxDB *pgxpool.Pool
}

// NewWorkflowRepository creates a new workflow repository
func NewWorkflowRepository(pgxDB *pgxpool.Pool, db *gorm.DB) WorkflowRepositoryInterface {
	return &WorkflowRepository{
		db:    db,
		pgxDB: pgxDB,
	}
}

// Create creates a new workflow
func (r *WorkflowRepository) Create(ctx context.Context, organizationID, name, description, documentType string, stages datatypes.JSON, isActive bool, createdBy string) (*models.Workflow, error) {
	workflow := &models.Workflow{
		ID:             uuid.New(),
		OrganizationID: organizationID,
		Name:           name,
		Description:    description,
		DocumentType:   documentType,
		Stages:         stages,
		IsActive:       isActive,
		CreatedBy:      &createdBy,
	}

	if err := r.db.WithContext(ctx).Create(workflow).Error; err != nil {
		return nil, err
	}

	return workflow, nil
}

// GetByID retrieves a workflow by ID
func (r *WorkflowRepository) GetByID(ctx context.Context, id uuid.UUID, organizationID string) (*models.Workflow, error) {
	var workflow models.Workflow
	err := r.db.WithContext(ctx).
		Where("id = ? AND organization_id = ?", id, organizationID).
		Preload("Creator").
		First(&workflow).Error
	
	if err != nil {
		return nil, err
	}
	
	return &workflow, nil
}

// Update updates a workflow
func (r *WorkflowRepository) Update(ctx context.Context, id uuid.UUID, organizationID, name, description string, stages datatypes.JSON) (*models.Workflow, error) {
	var workflow models.Workflow
	err := r.db.WithContext(ctx).
		Where("id = ? AND organization_id = ?", id, organizationID).
		First(&workflow).Error
	
	if err != nil {
		return nil, err
	}

	workflow.Name = name
	workflow.Description = description
	workflow.Stages = stages

	if err := r.db.WithContext(ctx).Save(&workflow).Error; err != nil {
		return nil, err
	}

	return &workflow, nil
}

// Delete deletes a workflow
func (r *WorkflowRepository) Delete(ctx context.Context, id uuid.UUID, organizationID string) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND organization_id = ?", id, organizationID).
		Delete(&models.Workflow{}).Error
}

// List retrieves workflows with pagination
func (r *WorkflowRepository) List(ctx context.Context, organizationID string, limit, offset int) ([]*models.Workflow, error) {
	var workflows []*models.Workflow
	err := r.db.WithContext(ctx).
		Where("organization_id = ?", organizationID).
		Preload("Creator").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&workflows).Error
	
	return workflows, err
}

// ListActive retrieves active workflows with pagination
func (r *WorkflowRepository) ListActive(ctx context.Context, organizationID string, limit, offset int) ([]*models.Workflow, error) {
	var workflows []*models.Workflow
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND is_active = ?", organizationID, true).
		Preload("Creator").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&workflows).Error
	
	return workflows, err
}

// ListByDocumentType retrieves workflows by document type with pagination
func (r *WorkflowRepository) ListByDocumentType(ctx context.Context, organizationID, documentType string, limit, offset int) ([]*models.Workflow, error) {
	var workflows []*models.Workflow
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND document_type = ?", organizationID, documentType).
		Preload("Creator").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&workflows).Error
	
	return workflows, err
}

// ListActiveByDocumentType retrieves active workflows by document type with pagination
func (r *WorkflowRepository) ListActiveByDocumentType(ctx context.Context, organizationID, documentType string, limit, offset int) ([]*models.Workflow, error) {
	var workflows []*models.Workflow
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND document_type = ? AND is_active = ?", organizationID, documentType, true).
		Preload("Creator").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&workflows).Error
	
	return workflows, err
}

// GetDefaultByDocumentType retrieves the default workflow for a document type
func (r *WorkflowRepository) GetDefaultByDocumentType(ctx context.Context, organizationID, documentType string) (*models.Workflow, error) {
	var workflow models.Workflow
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND document_type = ? AND is_active = ?", organizationID, documentType, true).
		Preload("Creator").
		Order("created_at DESC").
		First(&workflow).Error
	
	if err != nil {
		return nil, err
	}
	
	return &workflow, nil
}

// Activate activates a workflow
func (r *WorkflowRepository) Activate(ctx context.Context, id uuid.UUID, organizationID string) (*models.Workflow, error) {
	var workflow models.Workflow
	err := r.db.WithContext(ctx).
		Where("id = ? AND organization_id = ?", id, organizationID).
		First(&workflow).Error
	
	if err != nil {
		return nil, err
	}

	workflow.IsActive = true
	if err := r.db.WithContext(ctx).Save(&workflow).Error; err != nil {
		return nil, err
	}

	return &workflow, nil
}

// Deactivate deactivates a workflow
func (r *WorkflowRepository) Deactivate(ctx context.Context, id uuid.UUID, organizationID string) (*models.Workflow, error) {
	var workflow models.Workflow
	err := r.db.WithContext(ctx).
		Where("id = ? AND organization_id = ?", id, organizationID).
		First(&workflow).Error
	
	if err != nil {
		return nil, err
	}

	workflow.IsActive = false
	if err := r.db.WithContext(ctx).Save(&workflow).Error; err != nil {
		return nil, err
	}

	return &workflow, nil
}

// Count counts total workflows
func (r *WorkflowRepository) Count(ctx context.Context, organizationID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Workflow{}).
		Where("organization_id = ?", organizationID).
		Count(&count).Error
	
	return count, err
}

// CountActive counts active workflows
func (r *WorkflowRepository) CountActive(ctx context.Context, organizationID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Workflow{}).
		Where("organization_id = ? AND is_active = ?", organizationID, true).
		Count(&count).Error
	
	return count, err
}

// CountByDocumentType counts workflows by document type
func (r *WorkflowRepository) CountByDocumentType(ctx context.Context, organizationID, documentType string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Workflow{}).
		Where("organization_id = ? AND document_type = ?", organizationID, documentType).
		Count(&count).Error
	
	return count, err
}