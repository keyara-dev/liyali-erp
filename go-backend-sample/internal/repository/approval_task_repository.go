package repository

import (
	"context"

	"github.com/cozyCodr/liyali-gateway/internal/db"
	"github.com/cozyCodr/liyali-gateway/internal/utils"
	"github.com/google/uuid"
)

type ApprovalTaskRepository struct {
	queries *db.Queries
}

func NewApprovalTaskRepository(queries *db.Queries) *ApprovalTaskRepository {
	return &ApprovalTaskRepository{
		queries: queries,
	}
}

func (r *ApprovalTaskRepository) CreateApprovalTask(ctx context.Context, params db.CreateApprovalTaskParams) (*db.ApprovalTask, error) {
	task, err := r.queries.CreateApprovalTask(ctx, params)
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *ApprovalTaskRepository) GetApprovalTaskByID(ctx context.Context, id uuid.UUID) (*db.ApprovalTask, error) {
	task, err := r.queries.GetApprovalTaskByID(ctx, utils.UUIDToPgtype(id))
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *ApprovalTaskRepository) ListApprovalTasksByAssignee(ctx context.Context, assigneeID uuid.UUID, limit, offset int32) ([]db.ApprovalTask, error) {
	return r.queries.ListApprovalTasksByAssignee(ctx, db.ListApprovalTasksByAssigneeParams{
		AssignedTo: utils.UUIDToPgtype(assigneeID),
		Limit:      limit,
		Offset:     offset,
	})
}

func (r *ApprovalTaskRepository) ListApprovalTasksByStatus(ctx context.Context, status string, limit, offset int32) ([]db.ApprovalTask, error) {
	return r.queries.ListApprovalTasksByStatus(ctx, db.ListApprovalTasksByStatusParams{
		Status: status,
		Limit:  limit,
		Offset: offset,
	})
}

func (r *ApprovalTaskRepository) ListApprovalTasksByAssigneeAndStatus(ctx context.Context, assigneeID uuid.UUID, status string, limit, offset int32) ([]db.ApprovalTask, error) {
	return r.queries.ListApprovalTasksByAssigneeAndStatus(ctx, db.ListApprovalTasksByAssigneeAndStatusParams{
		AssignedTo: utils.UUIDToPgtype(assigneeID),
		Status:     status,
		Limit:      limit,
		Offset:     offset,
	})
}

func (r *ApprovalTaskRepository) ListApprovalTasksByDocument(ctx context.Context, documentID uuid.UUID) ([]db.ApprovalTask, error) {
	return r.queries.ListApprovalTasksByDocument(ctx, utils.UUIDToPgtype(documentID))
}

func (r *ApprovalTaskRepository) ListPendingApprovalTasks(ctx context.Context, limit, offset int32) ([]db.ApprovalTask, error) {
	return r.queries.ListPendingApprovalTasks(ctx, db.ListPendingApprovalTasksParams{
		Limit:  limit,
		Offset: offset,
	})
}

func (r *ApprovalTaskRepository) ListOverdueApprovalTasks(ctx context.Context, limit, offset int32) ([]db.ApprovalTask, error) {
	return r.queries.ListOverdueApprovalTasks(ctx, db.ListOverdueApprovalTasksParams{
		Limit:  limit,
		Offset: offset,
	})
}

func (r *ApprovalTaskRepository) UpdateApprovalTaskStatus(ctx context.Context, params db.UpdateApprovalTaskStatusParams) (*db.ApprovalTask, error) {
	task, err := r.queries.UpdateApprovalTaskStatus(ctx, params)
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *ApprovalTaskRepository) UpdateApprovalTaskStage(ctx context.Context, params db.UpdateApprovalTaskStageParams) (*db.ApprovalTask, error) {
	task, err := r.queries.UpdateApprovalTaskStage(ctx, params)
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *ApprovalTaskRepository) ReassignApprovalTask(ctx context.Context, params db.ReassignApprovalTaskParams) (*db.ApprovalTask, error) {
	task, err := r.queries.ReassignApprovalTask(ctx, params)
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *ApprovalTaskRepository) UpdateApprovalTaskNotes(ctx context.Context, params db.UpdateApprovalTaskNotesParams) (*db.ApprovalTask, error) {
	task, err := r.queries.UpdateApprovalTaskNotes(ctx, params)
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *ApprovalTaskRepository) DeleteApprovalTask(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeleteApprovalTask(ctx, utils.UUIDToPgtype(id))
}

func (r *ApprovalTaskRepository) CountApprovalTasksByAssignee(ctx context.Context, assigneeID uuid.UUID) (int64, error) {
	return r.queries.CountApprovalTasksByAssignee(ctx, utils.UUIDToPgtype(assigneeID))
}

func (r *ApprovalTaskRepository) CountApprovalTasksByStatus(ctx context.Context, status string) (int64, error) {
	return r.queries.CountApprovalTasksByStatus(ctx, status)
}

func (r *ApprovalTaskRepository) CountPendingApprovalTasksByAssignee(ctx context.Context, assigneeID uuid.UUID) (int64, error) {
	return r.queries.CountPendingApprovalTasksByAssignee(ctx, utils.UUIDToPgtype(assigneeID))
}
