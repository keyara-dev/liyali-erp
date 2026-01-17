package integration

import (
	"context"

	"github.com/liyali/liyali-gateway/models"
	"gorm.io/gorm"
)

// SimpleWorkflowRepo is a simple implementation for testing
type SimpleWorkflowRepo struct {
	db *gorm.DB
}

func (r *SimpleWorkflowRepo) GetByID(ctx context.Context, id string) (*models.Workflow, error) {
	var workflow models.Workflow
	err := r.db.Where("id = ?", id).First(&workflow).Error
	return &workflow, err
}

func (r *SimpleWorkflowRepo) GetDefaultByDocumentType(ctx context.Context, organizationID, documentType string) (*models.Workflow, error) {
	var workflow models.Workflow
	err := r.db.Where("organization_id = ? AND document_type = ? AND is_default = ? AND is_active = ?", 
		organizationID, documentType, true, true).First(&workflow).Error
	return &workflow, err
}

func (r *SimpleWorkflowRepo) Create(ctx context.Context, workflow *models.Workflow) error {
	return r.db.Create(workflow).Error
}

func (r *SimpleWorkflowRepo) Update(ctx context.Context, workflow *models.Workflow) error {
	return r.db.Save(workflow).Error
}

func (r *SimpleWorkflowRepo) Delete(ctx context.Context, id string) error {
	return r.db.Delete(&models.Workflow{}, "id = ?", id).Error
}

func (r *SimpleWorkflowRepo) List(ctx context.Context, organizationID string, filters map[string]interface{}) ([]*models.Workflow, error) {
	var workflows []*models.Workflow
	query := r.db.Where("organization_id = ?", organizationID)
	
	for key, value := range filters {
		query = query.Where(key+" = ?", value)
	}
	
	err := query.Find(&workflows).Error
	return workflows, err
}
