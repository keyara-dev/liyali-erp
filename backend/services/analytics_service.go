package services

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/liyali/liyali-gateway/logging"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/types"
	"gorm.io/gorm"
)

// AnalyticsService handles analytics calculations
type AnalyticsService struct {
	db    *gorm.DB
	cache *CacheService
}

// NewAnalyticsService creates a new analytics service
func NewAnalyticsService(db *gorm.DB) *AnalyticsService {
	return &AnalyticsService{
		db:    db,
		cache: NewCacheService(time.Minute * 15), // 15-minute cache for analytics
	}
}

// GetRequisitionMetrics calculates all requisition metrics with caching
func (s *AnalyticsService) GetRequisitionMetrics(params types.AnalyticsQueryParams) (*types.RequisitionMetricsResponse, error) {
	// Generate cache key based on parameters
	cacheKey := s.generateCacheKey(params)
	
	// Try to get from cache first
	if cached, found := s.cache.Get(cacheKey); found {
		if metrics, ok := cached.(*types.RequisitionMetricsResponse); ok {
			return metrics, nil
		}
	}

	// Build query with organization filter (required for multi-tenancy)
	query := s.db
	if params.OrganizationID != "" {
		query = query.Where("organization_id = ?", params.OrganizationID)
	}
	if params.StartDate != nil {
		query = query.Where("created_at >= ?", params.StartDate)
	}
	if params.EndDate != nil {
		query = query.Where("created_at <= ?", params.EndDate)
	}
	if params.Department != "" {
		query = query.Where("department = ?", params.Department)
	}

	// Get status counts
	statusCounts, err := s.getStatusCounts(query)
	if err != nil {
		logging.WithFields(map[string]interface{}{
			"operation": "get_status_counts",
		}).WithError(err).Error("failed_to_get_status_counts")
		statusCounts = make(map[string]int64)
	}

	// Get rejection rate
	rejectionRate, err := s.calculateRejectionRate(query)
	if err != nil {
		logging.WithFields(map[string]interface{}{
			"operation": "calculate_rejection_rate",
		}).WithError(err).Error("failed_to_calculate_rejection_rate")
		rejectionRate = 0
	}

	// Get rejections over time
	rejectionsOverTime, err := s.getRejectionsOverTime(query, params.Period)
	if err != nil {
		logging.WithFields(map[string]interface{}{
			"operation": "get_rejections_over_time",
			"period":    params.Period,
		}).WithError(err).Error("failed_to_get_rejections_over_time")
		rejectionsOverTime = []types.RejectionTimeData{}
	}

	// Get rejection reasons
	rejectionReasons, err := s.getRejectionReasons(query)
	if err != nil {
		logging.WithFields(map[string]interface{}{
			"operation": "get_rejection_reasons",
		}).WithError(err).Error("failed_to_get_rejection_reasons")
		rejectionReasons = []types.RejectionReason{}
	}

	// Get top rejecting approvers
	topRejectingApprovers, err := s.getTopRejectingApprovers(query)
	if err != nil {
		logging.WithFields(map[string]interface{}{
			"operation": "get_top_rejecting_approvers",
		}).WithError(err).Error("failed_to_get_top_rejecting_approvers")
		topRejectingApprovers = []types.ApproverStats{}
	}

	// Get total requisitions from status counts to avoid extra query
	var total int64
	for _, count := range statusCounts {
		total += count
	}

	result := &types.RequisitionMetricsResponse{
		StatusCounts:          statusCounts,
		RejectionRate:         rejectionRate,
		RejectionsOverTime:    rejectionsOverTime,
		RejectionReasons:      rejectionReasons,
		TopRejectingApprovers: topRejectingApprovers,
		TotalRequisitions:     total,
		Period:                params.Period,
	}

	// Cache the result
	s.cache.Set(cacheKey, result)

	return result, nil
}

// generateCacheKey creates a unique cache key for analytics parameters
func (s *AnalyticsService) generateCacheKey(params types.AnalyticsQueryParams) string {
	data, _ := json.Marshal(params)
	hash := fmt.Sprintf("%x", md5.Sum(data))
	return s.cache.AnalyticsKey(params.OrganizationID, hash)
}

// getStatusCounts counts requisitions by status - optimized with single query
func (s *AnalyticsService) getStatusCounts(query *gorm.DB) (map[string]int64, error) {
	var results []struct {
		Status string
		Count  int64
	}

	// Use the new composite index for better performance
	if err := query.Model(&models.Requisition{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&results).Error; err != nil {
		return nil, err
	}

	statusCounts := make(map[string]int64)
	for _, result := range results {
		statusCounts[result.Status] = result.Count
	}

	return statusCounts, nil
}

// calculateRejectionRate calculates overall rejection rate - optimized single query
func (s *AnalyticsService) calculateRejectionRate(query *gorm.DB) (float64, error) {
	var result struct {
		TotalCount    int64
		RejectedCount int64
	}

	// Single query to get both counts using conditional aggregation
	if err := query.Model(&models.Requisition{}).
		Select(`
			COUNT(*) as total_count,
			COUNT(CASE WHEN UPPER(status) = 'REJECTED' THEN 1 END) as rejected_count
		`).
		Scan(&result).Error; err != nil {
		return 0, err
	}

	if result.TotalCount == 0 {
		return 0, nil
	}

	return float64(result.RejectedCount) / float64(result.TotalCount) * 100, nil
}

// getRejectionsOverTime groups rejections by time period
func (s *AnalyticsService) getRejectionsOverTime(query *gorm.DB, period string) ([]types.RejectionTimeData, error) {
	if period == "" {
		period = "daily"
	}

	var dateFormat string
	switch period {
	case "weekly":
		dateFormat = "2006-W02" // ISO week format
	case "monthly":
		dateFormat = "2006-01"
	default: // daily
		dateFormat = "2006-01-02"
	}

	var requisitions []models.Requisition
	if err := query.Find(&requisitions).Error; err != nil {
		return nil, err
	}

	// Group by date
	timeGroupMap := make(map[string]*types.RejectionTimeData)

	for _, req := range requisitions {
		dateStr := req.CreatedAt.Format(dateFormat)

		if _, exists := timeGroupMap[dateStr]; !exists {
			timeGroupMap[dateStr] = &types.RejectionTimeData{
				Date:       dateStr,
				Rejections: 0,
				Total:      0,
				Rate:       0,
			}
		}

		timeGroupMap[dateStr].Total++
		if strings.ToUpper(req.Status) == "REJECTED" {
			timeGroupMap[dateStr].Rejections++
		}
	}

	// Calculate rates and convert to slice
	results := make([]types.RejectionTimeData, 0, len(timeGroupMap))
	for _, data := range timeGroupMap {
		if data.Total > 0 {
			data.Rate = float64(data.Rejections) / float64(data.Total) * 100
		}
		results = append(results, *data)
	}

	return results, nil
}

// getRejectionReasons extracts reasons from approval_history JSONB - optimized query
func (s *AnalyticsService) getRejectionReasons(query *gorm.DB) ([]types.RejectionReason, error) {
	var requisitions []models.Requisition

	// Use the partial index for rejected requisitions
	if err := query.Where("UPPER(status) = ?", "REJECTED").
		Select("approval_history").
		Find(&requisitions).Error; err != nil {
		return nil, err
	}

	reasonCounts := make(map[string]int64)
	var totalRejections int64

	for _, req := range requisitions {
		// Get the raw JSON data from JSONType
		approvalHistoryBytes, err := req.ApprovalHistory.MarshalJSON()
		if err != nil || len(approvalHistoryBytes) == 0 {
			continue
		}
		
		var approvalRecords []types.ApprovalRecord
		if err := json.Unmarshal(approvalHistoryBytes, &approvalRecords); err != nil {
			continue
		}

		for _, record := range approvalRecords {
			if strings.ToUpper(record.Status) == "REJECTED" {
				reason := strings.TrimSpace(record.Comments)
				if reason == "" {
					reason = "No reason provided"
				}
				reasonCounts[reason]++
				totalRejections++
			}
		}
	}

	// Convert to response format
	results := make([]types.RejectionReason, 0, len(reasonCounts))
	for reason, count := range reasonCounts {
		percentage := float64(0)
		if totalRejections > 0 {
			percentage = float64(count) / float64(totalRejections) * 100
		}

		results = append(results, types.RejectionReason{
			Reason:     reason,
			Count:      count,
			Percentage: percentage,
		})
	}

	return results, nil
}

// getTopRejectingApprovers identifies approvers with highest rejection rates
func (s *AnalyticsService) getTopRejectingApprovers(query *gorm.DB) ([]types.ApproverStats, error) {
	var requisitions []models.Requisition

	if err := query.Preload("Requester").Find(&requisitions).Error; err != nil {
		return nil, err
	}

	approverStatsMap := make(map[string]*types.ApproverStats)

	for _, req := range requisitions {
		// Get the raw JSON data from JSONType
		approvalHistoryBytes, err := req.ApprovalHistory.MarshalJSON()
		if err != nil || len(approvalHistoryBytes) == 0 {
			continue
		}
		
		var approvalRecords []types.ApprovalRecord
		if err := json.Unmarshal(approvalHistoryBytes, &approvalRecords); err != nil {
			continue
		}

		for _, record := range approvalRecords {
			approverID := record.ApproverID
			approverName := record.ApproverName

			if _, exists := approverStatsMap[approverID]; !exists {
				approverStatsMap[approverID] = &types.ApproverStats{
					ApproverID:    approverID,
					ApproverName:  approverName,
					Rejections:    0,
					Approvals:     0,
					RejectionRate: 0,
				}
			}

			if strings.ToUpper(record.Status) == "REJECTED" {
				approverStatsMap[approverID].Rejections++
			} else if strings.ToUpper(record.Status) == "APPROVED" {
				approverStatsMap[approverID].Approvals++
			}
		}
	}

	// Convert to slice and calculate rates
	results := make([]types.ApproverStats, 0, len(approverStatsMap))
	for _, stats := range approverStatsMap {
		total := stats.Rejections + stats.Approvals
		if total > 0 {
			stats.RejectionRate = float64(stats.Rejections) / float64(total) * 100
		}
		results = append(results, *stats)
	}

	return results, nil
}
