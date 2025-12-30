package repository

import (
	"context"

	"github.com/cozyCodr/liyali-gateway/internal/db"
	"github.com/cozyCodr/liyali-gateway/internal/utils"
	"github.com/google/uuid"
)

type ApprovalHistoryRepository struct {
	queries *db.Queries
}

func NewApprovalHistoryRepository(queries *db.Queries) *ApprovalHistoryRepository {
	return &ApprovalHistoryRepository{
		queries: queries,
	}
}

func (r *ApprovalHistoryRepository) CreateApprovalHistoryEntry(ctx context.Context, params db.CreateApprovalHistoryEntryParams) (*db.ApprovalHistory, error) {
	entry, err := r.queries.CreateApprovalHistoryEntry(ctx, params)
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

func (r *ApprovalHistoryRepository) GetApprovalHistoryByID(ctx context.Context, id uuid.UUID) (*db.ApprovalHistory, error) {
	entry, err := r.queries.GetApprovalHistoryByID(ctx, utils.UUIDToPgtype(id))
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

func (r *ApprovalHistoryRepository) ListApprovalHistoryByTask(ctx context.Context, taskID uuid.UUID) ([]db.ApprovalHistory, error) {
	return r.queries.ListApprovalHistoryByTask(ctx, utils.UUIDToPgtype(taskID))
}

func (r *ApprovalHistoryRepository) ListApprovalHistoryByUser(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]db.ApprovalHistory, error) {
	return r.queries.ListApprovalHistoryByUser(ctx, db.ListApprovalHistoryByUserParams{
		UserID: utils.UUIDToPgtype(userID),
		Limit:  limit,
		Offset: offset,
	})
}

func (r *ApprovalHistoryRepository) ListApprovalHistoryByAction(ctx context.Context, action string, limit, offset int32) ([]db.ApprovalHistory, error) {
	return r.queries.ListApprovalHistoryByAction(ctx, db.ListApprovalHistoryByActionParams{
		Action: action,
		Limit:  limit,
		Offset: offset,
	})
}

func (r *ApprovalHistoryRepository) GetLatestApprovalHistoryByTask(ctx context.Context, taskID uuid.UUID) (*db.ApprovalHistory, error) {
	entry, err := r.queries.GetLatestApprovalHistoryByTask(ctx, utils.UUIDToPgtype(taskID))
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

func (r *ApprovalHistoryRepository) DeleteApprovalHistory(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeleteApprovalHistory(ctx, utils.UUIDToPgtype(id))
}

func (r *ApprovalHistoryRepository) CountApprovalHistoryByTask(ctx context.Context, taskID uuid.UUID) (int64, error) {
	return r.queries.CountApprovalHistoryByTask(ctx, utils.UUIDToPgtype(taskID))
}

func (r *ApprovalHistoryRepository) CountApprovalHistoryByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	return r.queries.CountApprovalHistoryByUser(ctx, utils.UUIDToPgtype(userID))
}
