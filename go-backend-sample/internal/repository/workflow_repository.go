package repository

import (
	"context"

	"github.com/cozyCodr/liyali-gateway/internal/db"
	"github.com/cozyCodr/liyali-gateway/internal/utils"
	"github.com/google/uuid"
)

type WorkflowRepository struct {
	queries *db.Queries
}

func NewWorkflowRepository(queries *db.Queries) *WorkflowRepository {
	return &WorkflowRepository{
		queries: queries,
	}
}

func (r *WorkflowRepository) CreateWorkflow(ctx context.Context, params db.CreateWorkflowParams) (*db.Workflow, error) {
	workflow, err := r.queries.CreateWorkflow(ctx, params)
	if err != nil {
		return nil, err
	}
	return &workflow, nil
}

func (r *WorkflowRepository) GetWorkflowByID(ctx context.Context, id uuid.UUID) (*db.Workflow, error) {
	workflow, err := r.queries.GetWorkflowByID(ctx, utils.UUIDToPgtype(id))
	if err != nil {
		return nil, err
	}
	return &workflow, nil
}

func (r *WorkflowRepository) ListWorkflows(ctx context.Context, limit, offset int32) ([]db.Workflow, error) {
	return r.queries.ListWorkflows(ctx, db.ListWorkflowsParams{
		Limit:  limit,
		Offset: offset,
	})
}

func (r *WorkflowRepository) ListActiveWorkflows(ctx context.Context, limit, offset int32) ([]db.Workflow, error) {
	return r.queries.ListActiveWorkflows(ctx, db.ListActiveWorkflowsParams{
		Limit:  limit,
		Offset: offset,
	})
}

func (r *WorkflowRepository) ListWorkflowsByDocumentType(ctx context.Context, documentType string, limit, offset int32) ([]db.Workflow, error) {
	return r.queries.ListWorkflowsByDocumentType(ctx, db.ListWorkflowsByDocumentTypeParams{
		DocumentType: documentType,
		Limit:        limit,
		Offset:       offset,
	})
}

func (r *WorkflowRepository) ListActiveWorkflowsByDocumentType(ctx context.Context, documentType string, limit, offset int32) ([]db.Workflow, error) {
	return r.queries.ListActiveWorkflowsByDocumentType(ctx, db.ListActiveWorkflowsByDocumentTypeParams{
		DocumentType: documentType,
		Limit:        limit,
		Offset:       offset,
	})
}

func (r *WorkflowRepository) GetDefaultWorkflowByDocumentType(ctx context.Context, documentType string) (*db.Workflow, error) {
	workflow, err := r.queries.GetDefaultWorkflowByDocumentType(ctx, documentType)
	if err != nil {
		return nil, err
	}
	return &workflow, nil
}

func (r *WorkflowRepository) UpdateWorkflow(ctx context.Context, params db.UpdateWorkflowParams) (*db.Workflow, error) {
	workflow, err := r.queries.UpdateWorkflow(ctx, params)
	if err != nil {
		return nil, err
	}
	return &workflow, nil
}

func (r *WorkflowRepository) ActivateWorkflow(ctx context.Context, id uuid.UUID) (*db.Workflow, error) {
	workflow, err := r.queries.ActivateWorkflow(ctx, utils.UUIDToPgtype(id))
	if err != nil {
		return nil, err
	}
	return &workflow, nil
}

func (r *WorkflowRepository) DeactivateWorkflow(ctx context.Context, id uuid.UUID) (*db.Workflow, error) {
	workflow, err := r.queries.DeactivateWorkflow(ctx, utils.UUIDToPgtype(id))
	if err != nil {
		return nil, err
	}
	return &workflow, nil
}

func (r *WorkflowRepository) DeleteWorkflow(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeleteWorkflow(ctx, utils.UUIDToPgtype(id))
}

func (r *WorkflowRepository) CountWorkflows(ctx context.Context) (int64, error) {
	return r.queries.CountWorkflows(ctx)
}

func (r *WorkflowRepository) CountActiveWorkflows(ctx context.Context) (int64, error) {
	return r.queries.CountActiveWorkflows(ctx)
}

func (r *WorkflowRepository) CountWorkflowsByDocumentType(ctx context.Context, documentType string) (int64, error) {
	return r.queries.CountWorkflowsByDocumentType(ctx, documentType)
}
