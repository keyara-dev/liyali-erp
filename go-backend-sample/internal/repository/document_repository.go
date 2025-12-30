package repository

import (
	"context"

	"github.com/cozyCodr/liyali-gateway/internal/db"
	"github.com/cozyCodr/liyali-gateway/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type DocumentRepository struct {
	queries *db.Queries
}

func NewDocumentRepository(queries *db.Queries) *DocumentRepository {
	return &DocumentRepository{
		queries: queries,
	}
}

func (r *DocumentRepository) CreateDocument(ctx context.Context, params db.CreateDocumentParams) (*db.Document, error) {
	document, err := r.queries.CreateDocument(ctx, params)
	if err != nil {
		return nil, err
	}
	return &document, nil
}

func (r *DocumentRepository) GetDocumentByID(ctx context.Context, id uuid.UUID) (*db.Document, error) {
	document, err := r.queries.GetDocumentByID(ctx, utils.UUIDToPgtype(id))
	if err != nil {
		return nil, err
	}
	return &document, nil
}

func (r *DocumentRepository) GetDocumentByNumber(ctx context.Context, documentNumber string) (*db.Document, error) {
	document, err := r.queries.GetDocumentByNumber(ctx, documentNumber)
	if err != nil {
		return nil, err
	}
	return &document, nil
}

func (r *DocumentRepository) ListDocuments(ctx context.Context, limit, offset int32) ([]db.Document, error) {
	return r.queries.ListDocuments(ctx, db.ListDocumentsParams{
		Limit:  limit,
		Offset: offset,
	})
}

func (r *DocumentRepository) ListDocumentsByType(ctx context.Context, documentType string, limit, offset int32) ([]db.Document, error) {
	return r.queries.ListDocumentsByType(ctx, db.ListDocumentsByTypeParams{
		DocumentType: documentType,
		Limit:        limit,
		Offset:       offset,
	})
}

func (r *DocumentRepository) ListDocumentsByStatus(ctx context.Context, status string, limit, offset int32) ([]db.Document, error) {
	return r.queries.ListDocumentsByStatus(ctx, db.ListDocumentsByStatusParams{
		Status: status,
		Limit:  limit,
		Offset: offset,
	})
}

func (r *DocumentRepository) ListDocumentsByCreator(ctx context.Context, creatorID uuid.UUID, limit, offset int32) ([]db.Document, error) {
	return r.queries.ListDocumentsByCreator(ctx, db.ListDocumentsByCreatorParams{
		CreatedBy: utils.UUIDToPgtype(creatorID),
		Limit:     limit,
		Offset:    offset,
	})
}

func (r *DocumentRepository) ListDocumentsByDepartment(ctx context.Context, department string, limit, offset int32) ([]db.Document, error) {
	return r.queries.ListDocumentsByDepartment(ctx, db.ListDocumentsByDepartmentParams{
		Department: pgtype.Text{String: department, Valid: true},
		Limit:      limit,
		Offset:     offset,
	})
}

func (r *DocumentRepository) ListDocumentsByTypeAndStatus(ctx context.Context, documentType, status string, limit, offset int32) ([]db.Document, error) {
	return r.queries.ListDocumentsByTypeAndStatus(ctx, db.ListDocumentsByTypeAndStatusParams{
		DocumentType: documentType,
		Status:       status,
		Limit:        limit,
		Offset:       offset,
	})
}

func (r *DocumentRepository) ListDocumentsByWorkflow(ctx context.Context, workflowID uuid.UUID, limit, offset int32) ([]db.Document, error) {
	return r.queries.ListDocumentsByWorkflow(ctx, db.ListDocumentsByWorkflowParams{
		WorkflowID: utils.UUIDToPgtype(workflowID),
		Limit:      limit,
		Offset:     offset,
	})
}

func (r *DocumentRepository) UpdateDocument(ctx context.Context, params db.UpdateDocumentParams) (*db.Document, error) {
	document, err := r.queries.UpdateDocument(ctx, params)
	if err != nil {
		return nil, err
	}
	return &document, nil
}

func (r *DocumentRepository) UpdateDocumentStatus(ctx context.Context, params db.UpdateDocumentStatusParams) (*db.Document, error) {
	document, err := r.queries.UpdateDocumentStatus(ctx, params)
	if err != nil {
		return nil, err
	}
	return &document, nil
}

func (r *DocumentRepository) SubmitDocument(ctx context.Context, id uuid.UUID) (*db.Document, error) {
	document, err := r.queries.SubmitDocument(ctx, utils.UUIDToPgtype(id))
	if err != nil {
		return nil, err
	}
	return &document, nil
}

func (r *DocumentRepository) ApproveDocument(ctx context.Context, id uuid.UUID) (*db.Document, error) {
	document, err := r.queries.ApproveDocument(ctx, utils.UUIDToPgtype(id))
	if err != nil {
		return nil, err
	}
	return &document, nil
}

func (r *DocumentRepository) RejectDocument(ctx context.Context, id uuid.UUID) (*db.Document, error) {
	document, err := r.queries.RejectDocument(ctx, utils.UUIDToPgtype(id))
	if err != nil {
		return nil, err
	}
	return &document, nil
}

func (r *DocumentRepository) DeleteDocument(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeleteDocument(ctx, utils.UUIDToPgtype(id))
}

func (r *DocumentRepository) CountDocuments(ctx context.Context) (int64, error) {
	return r.queries.CountDocuments(ctx)
}

func (r *DocumentRepository) CountDocumentsByType(ctx context.Context, documentType string) (int64, error) {
	return r.queries.CountDocumentsByType(ctx, documentType)
}

func (r *DocumentRepository) CountDocumentsByStatus(ctx context.Context, status string) (int64, error) {
	return r.queries.CountDocumentsByStatus(ctx, status)
}

func (r *DocumentRepository) CountDocumentsByCreator(ctx context.Context, creatorID uuid.UUID) (int64, error) {
	return r.queries.CountDocumentsByCreator(ctx, utils.UUIDToPgtype(creatorID))
}
