package repository

import (
	"context"
	"time"

	"github.com/cozyCodr/liyali-gateway/internal/db"
	"github.com/cozyCodr/liyali-gateway/internal/utils"
	"github.com/google/uuid"
)

type AuditLogRepository struct {
	queries *db.Queries
}

func NewAuditLogRepository(queries *db.Queries) *AuditLogRepository {
	return &AuditLogRepository{
		queries: queries,
	}
}

func (r *AuditLogRepository) CreateAuditLog(ctx context.Context, params db.CreateAuditLogParams) (*db.AuditLog, error) {
	log, err := r.queries.CreateAuditLog(ctx, params)
	if err != nil {
		return nil, err
	}
	return &log, nil
}

func (r *AuditLogRepository) GetAuditLogByID(ctx context.Context, id uuid.UUID) (*db.AuditLog, error) {
	log, err := r.queries.GetAuditLogByID(ctx, utils.UUIDToPgtype(id))
	if err != nil {
		return nil, err
	}
	return &log, nil
}

func (r *AuditLogRepository) ListAuditLogs(ctx context.Context, limit, offset int32) ([]db.AuditLog, error) {
	return r.queries.ListAuditLogs(ctx, db.ListAuditLogsParams{
		Limit:  limit,
		Offset: offset,
	})
}

func (r *AuditLogRepository) ListAuditLogsByUser(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]db.AuditLog, error) {
	return r.queries.ListAuditLogsByUser(ctx, db.ListAuditLogsByUserParams{
		UserID: utils.UUIDToPgtype(userID),
		Limit:  limit,
		Offset: offset,
	})
}

func (r *AuditLogRepository) ListAuditLogsByResource(ctx context.Context, resourceType string, resourceID uuid.UUID, limit, offset int32) ([]db.AuditLog, error) {
	return r.queries.ListAuditLogsByResource(ctx, db.ListAuditLogsByResourceParams{
		ResourceType: resourceType,
		ResourceID:   utils.UUIDToPgtype(resourceID),
		Limit:        limit,
		Offset:       offset,
	})
}

func (r *AuditLogRepository) ListAuditLogsByResourceType(ctx context.Context, resourceType string, limit, offset int32) ([]db.AuditLog, error) {
	return r.queries.ListAuditLogsByResourceType(ctx, db.ListAuditLogsByResourceTypeParams{
		ResourceType: resourceType,
		Limit:        limit,
		Offset:       offset,
	})
}

func (r *AuditLogRepository) ListAuditLogsByAction(ctx context.Context, action string, limit, offset int32) ([]db.AuditLog, error) {
	return r.queries.ListAuditLogsByAction(ctx, db.ListAuditLogsByActionParams{
		Action: action,
		Limit:  limit,
		Offset: offset,
	})
}

func (r *AuditLogRepository) DeleteOldAuditLogs(ctx context.Context, before time.Time) error {
	return r.queries.DeleteOldAuditLogs(ctx, utils.TimeToPgtype(before))
}

func (r *AuditLogRepository) CountAuditLogs(ctx context.Context) (int64, error) {
	return r.queries.CountAuditLogs(ctx)
}

func (r *AuditLogRepository) CountAuditLogsByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	return r.queries.CountAuditLogsByUser(ctx, utils.UUIDToPgtype(userID))
}

func (r *AuditLogRepository) CountAuditLogsByResource(ctx context.Context, resourceType string, resourceID uuid.UUID) (int64, error) {
	return r.queries.CountAuditLogsByResource(ctx, db.CountAuditLogsByResourceParams{
		ResourceType: resourceType,
		ResourceID:   utils.UUIDToPgtype(resourceID),
	})
}
