package repository

import (
	"context"
	"time"

	"github.com/cozyCodr/liyali-gateway/internal/db"
	"github.com/cozyCodr/liyali-gateway/internal/utils"
	"github.com/google/uuid"
)

type NotificationRepository struct {
	queries *db.Queries
}

func NewNotificationRepository(queries *db.Queries) *NotificationRepository {
	return &NotificationRepository{
		queries: queries,
	}
}

func (r *NotificationRepository) CreateNotification(ctx context.Context, params db.CreateNotificationParams) (*db.Notification, error) {
	notification, err := r.queries.CreateNotification(ctx, params)
	if err != nil {
		return nil, err
	}
	return &notification, nil
}

func (r *NotificationRepository) GetNotificationByID(ctx context.Context, id uuid.UUID) (*db.Notification, error) {
	notification, err := r.queries.GetNotificationByID(ctx, utils.UUIDToPgtype(id))
	if err != nil {
		return nil, err
	}
	return &notification, nil
}

func (r *NotificationRepository) ListNotificationsByUser(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]db.Notification, error) {
	return r.queries.ListNotificationsByUser(ctx, db.ListNotificationsByUserParams{
		UserID: utils.UUIDToPgtype(userID),
		Limit:  limit,
		Offset: offset,
	})
}

func (r *NotificationRepository) ListUnreadNotificationsByUser(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]db.Notification, error) {
	return r.queries.ListUnreadNotificationsByUser(ctx, db.ListUnreadNotificationsByUserParams{
		UserID: utils.UUIDToPgtype(userID),
		Limit:  limit,
		Offset: offset,
	})
}

func (r *NotificationRepository) ListNotificationsByType(ctx context.Context, notificationType string, limit, offset int32) ([]db.Notification, error) {
	return r.queries.ListNotificationsByType(ctx, db.ListNotificationsByTypeParams{
		Type:   notificationType,
		Limit:  limit,
		Offset: offset,
	})
}

func (r *NotificationRepository) ListNotificationsByRelatedID(ctx context.Context, relatedID uuid.UUID) ([]db.Notification, error) {
	return r.queries.ListNotificationsByRelatedID(ctx, utils.UUIDToPgtype(relatedID))
}

func (r *NotificationRepository) MarkNotificationAsRead(ctx context.Context, id uuid.UUID) (*db.Notification, error) {
	notification, err := r.queries.MarkNotificationAsRead(ctx, utils.UUIDToPgtype(id))
	if err != nil {
		return nil, err
	}
	return &notification, nil
}

func (r *NotificationRepository) MarkAllNotificationsAsRead(ctx context.Context, userID uuid.UUID) error {
	return r.queries.MarkAllNotificationsAsRead(ctx, utils.UUIDToPgtype(userID))
}

func (r *NotificationRepository) DeleteNotification(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeleteNotification(ctx, utils.UUIDToPgtype(id))
}

func (r *NotificationRepository) DeleteOldNotifications(ctx context.Context, before time.Time) error {
	return r.queries.DeleteOldNotifications(ctx, utils.TimeToPgtype(before))
}

func (r *NotificationRepository) CountUnreadNotificationsByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	return r.queries.CountUnreadNotificationsByUser(ctx, utils.UUIDToPgtype(userID))
}

func (r *NotificationRepository) CountNotificationsByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	return r.queries.CountNotificationsByUser(ctx, utils.UUIDToPgtype(userID))
}

func (r *NotificationRepository) CountNotificationsByType(ctx context.Context, notificationType string) (int64, error) {
	return r.queries.CountNotificationsByType(ctx, notificationType)
}
