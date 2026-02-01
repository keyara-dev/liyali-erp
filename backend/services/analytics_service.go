package services

import (
	"encoding/json"
	"strings"

	"github.com/liyali/liyali-gateway/logging"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/types"
	"gorm.io/gorm"
)

// AnalyticsService handles analytics calculations
type AnalyticsService struct {
	db *gorm.DB
}

// NewAnalyticsService creates a new analytics service
func NewAnalyticsService(db *gorm.DB) *AnalyticsService {
	return &AnalyticsService{db: db}
}

// GetRequisitionMetrics calculates all requisition metrics
func (s *AnalyticsService) GetRequisitionMetrics(params types.AnalyticsQueryParams) (*types.RequisitionMetricsResponse, error) {
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

	// Get total requisitions
	var total int64
	query.Model(&models.Requisition{}).Count(&total)

	return &types.RequisitionMetricsResponse{
		StatusCounts:          statusCounts,
		RejectionRate:         rejectionRate,
		RejectionsOverTime:    rejectionsOverTime,
		RejectionReasons:      rejectionReasons,
		TopRejectingApprovers: topRejectingApprovers,
		TotalRequisitions:     total,
		Period:                params.Period,
	}, nil
}

// getStatusCounts counts requisitions by status
func (s *AnalyticsService) getStatusCounts(query *gorm.DB) (map[string]int64, error) {
	var results []struct {
		Status string
		Count  int64
	}

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

// calculateRejectionRate calculates overall rejection rate
func (s *AnalyticsService) calculateRejectionRate(query *gorm.DB) (float64, error) {
	var totalCount int64
	var rejectedCount int64

	if err := query.Model(&models.Requisition{}).Count(&totalCount).Error; err != nil {
		return 0, err
	}

	if err := query.Model(&models.Requisition{}).Where("status = ?", "rejected").Count(&rejectedCount).Error; err != nil {
		return 0, err
	}

	if totalCount == 0 {
		return 0, nil
	}

	return float64(rejectedCount) / float64(totalCount) * 100, nil
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
		if req.Status == "rejected" {
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

// getRejectionReasons extracts reasons from approval_history JSONB
func (s *AnalyticsService) getRejectionReasons(query *gorm.DB) ([]types.RejectionReason, error) {
	var requisitions []models.Requisition

	if err := query.Where("status = ?", "rejected").Find(&requisitions).Error; err != nil {
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
			if record.Status == "rejected" {
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

			if record.Status == "rejected" {
				approverStatsMap[approverID].Rejections++
			} else if record.Status == "approved" {
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
