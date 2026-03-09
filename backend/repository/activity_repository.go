package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/liyali/liyali-gateway/models"
	"gorm.io/gorm"
)

// ActivityRepository handles persistence for user_activity_logs
type ActivityRepository struct {
	db *gorm.DB
}

// NewActivityRepository creates a new ActivityRepository
func NewActivityRepository(db *gorm.DB) *ActivityRepository {
	return &ActivityRepository{db: db}
}

// Create inserts a new activity log entry
func (r *ActivityRepository) Create(ctx context.Context, log *models.UserActivityLog) error {
	if err := r.db.WithContext(ctx).Create(log).Error; err != nil {
		return fmt.Errorf("activity_repository: create failed: %w", err)
	}
	return nil
}

// GetByUserID returns paginated activity logs for a user with optional filters
func (r *ActivityRepository) GetByUserID(
	ctx context.Context,
	userID string,
	filters models.ActivityFilters,
) ([]*models.UserActivityLog, int64, error) {
	page, limit := normalizePagination(filters.Page, filters.Limit)

	q := r.db.WithContext(ctx).
		Model(&models.UserActivityLog{}).
		Where("user_id = ?", userID)

	q = applyActivityFilters(q, filters)

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("activity_repository: count failed: %w", err)
	}

	var logs []*models.UserActivityLog
	offset := (page - 1) * limit
	if err := q.Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&logs).Error; err != nil {
		return nil, 0, fmt.Errorf("activity_repository: query failed: %w", err)
	}

	return logs, total, nil
}

// GetStatistics calculates activity statistics for a user over the given number of days
func (r *ActivityRepository) GetStatistics(
	ctx context.Context,
	userID string,
	days int,
) (*models.ActivityStatistics, error) {
	since := time.Now().AddDate(0, 0, -days)

	type countByType struct {
		ActionType string
		Count      int64
	}

	var rows []countByType
	if err := r.db.WithContext(ctx).
		Model(&models.UserActivityLog{}).
		Select("action_type, COUNT(*) as count").
		Where("user_id = ? AND created_at >= ?", userID, since).
		Group("action_type").
		Order("count DESC").
		Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("activity_repository: stats by type failed: %w", err)
	}

	type countByDay struct {
		Day   string
		Count int64
	}

	var dayRows []countByDay
	if err := r.db.WithContext(ctx).
		Model(&models.UserActivityLog{}).
		Select("DATE(created_at) as day, COUNT(*) as count").
		Where("user_id = ? AND created_at >= ?", userID, since).
		Group("day").
		Order("day ASC").
		Scan(&dayRows).Error; err != nil {
		return nil, fmt.Errorf("activity_repository: stats by day failed: %w", err)
	}

	var lastActivity *time.Time
	if err := r.db.WithContext(ctx).
		Model(&models.UserActivityLog{}).
		Select("MAX(created_at)").
		Where("user_id = ?", userID).
		Scan(&lastActivity).Error; err != nil {
		return nil, fmt.Errorf("activity_repository: last activity failed: %w", err)
	}

	stats := &models.ActivityStatistics{
		ActionsByType:    make(map[string]int64),
		ActionsByDay:     make(map[string]int64),
		LastActivityTime: lastActivity,
	}

	for _, r := range rows {
		stats.TotalActions += r.Count
		stats.ActionsByType[r.ActionType] = r.Count
		if stats.MostCommonAction == "" {
			stats.MostCommonAction = r.ActionType
		}
	}

	for _, d := range dayRows {
		stats.ActionsByDay[d.Day] = d.Count
	}

	if days > 0 {
		stats.AveragePerDay = float64(stats.TotalActions) / float64(days)
	}

	return stats, nil
}

// DeleteOlderThan removes activity logs older than the given cutoff time, returns rows deleted
func (r *ActivityRepository) DeleteOlderThan(ctx context.Context, cutoff time.Time) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("created_at < ?", cutoff).
		Delete(&models.UserActivityLog{})
	if result.Error != nil {
		return 0, fmt.Errorf("activity_repository: cleanup failed: %w", result.Error)
	}
	return result.RowsAffected, nil
}

// normalizePagination enforces sensible defaults and max-limit
func normalizePagination(page, limit int) (int, int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	return page, limit
}

// applyActivityFilters adds WHERE clauses for the optional filter fields
func applyActivityFilters(q *gorm.DB, f models.ActivityFilters) *gorm.DB {
	if f.StartDate != nil {
		q = q.Where("created_at >= ?", f.StartDate)
	}
	if f.EndDate != nil {
		q = q.Where("created_at <= ?", f.EndDate)
	}
	if f.ActionType != "" {
		q = q.Where("action_type = ?", f.ActionType)
	}
	if f.ResourceType != "" {
		q = q.Where("resource_type = ?", f.ResourceType)
	}
	if f.Search != "" {
		pattern := "%" + f.Search + "%"
		q = q.Where("action_type ILIKE ? OR resource_type ILIKE ? OR resource_id ILIKE ?",
			pattern, pattern, pattern)
	}
	return q
}
