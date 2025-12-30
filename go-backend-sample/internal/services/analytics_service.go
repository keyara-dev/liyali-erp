package services

import (
	"context"
	"time"

	"github.com/cozyCodr/liyali-gateway/internal/db"
	"github.com/cozyCodr/liyali-gateway/internal/repository"
	"github.com/cozyCodr/liyali-gateway/internal/utils"
	"github.com/google/uuid"
)

type AnalyticsService struct {
	documentRepo        repository.DocumentRepository
	approvalTaskRepo    repository.ApprovalTaskRepository
	approvalHistoryRepo repository.ApprovalHistoryRepository
	workflowRepo        repository.WorkflowRepository
}

func NewAnalyticsService(
	documentRepo repository.DocumentRepository,
	approvalTaskRepo repository.ApprovalTaskRepository,
	approvalHistoryRepo repository.ApprovalHistoryRepository,
	workflowRepo repository.WorkflowRepository,
) *AnalyticsService {
	return &AnalyticsService{
		documentRepo:        documentRepo,
		approvalTaskRepo:    approvalTaskRepo,
		approvalHistoryRepo: approvalHistoryRepo,
		workflowRepo:        workflowRepo,
	}
}

// DashboardMetrics represents key metrics for the dashboard
type DashboardMetrics struct {
	TotalDocuments      int64                    `json:"total_documents"`
	DocumentsByStatus   map[string]int64         `json:"documents_by_status"`
	PendingApprovals    int64                    `json:"pending_approvals"`
	OverdueApprovals    int64                    `json:"overdue_approvals"`
	ApprovalsByStatus   map[string]int64         `json:"approvals_by_status"`
	ActiveWorkflows     int64                    `json:"active_workflows"`
	AverageApprovalTime float64                  `json:"average_approval_time_hours"`
	DocumentsByType     map[string]int64         `json:"documents_by_type"`
}

// TrendData represents trend information over a period
type TrendData struct {
	Date              string `json:"date"`
	DocumentsCreated  int64  `json:"documents_created"`
	DocumentsApproved int64  `json:"documents_approved"`
	DocumentsRejected int64  `json:"documents_rejected"`
	ApprovalsCompleted int64 `json:"approvals_completed"`
}

// BottleneckInfo represents workflow bottleneck information
type BottleneckInfo struct {
	DocumentID          uuid.UUID `json:"document_id"`
	DocumentType        string    `json:"document_type"`
	Stage               int32     `json:"stage"`
	PendingTasks        int64     `json:"pending_tasks"`
	AverageTimeInStage  float64   `json:"average_time_in_stage_hours"`
}

// GetDashboardMetrics retrieves comprehensive dashboard metrics
func (s *AnalyticsService) GetDashboardMetrics(ctx context.Context, userID uuid.UUID, userRole string) (*DashboardMetrics, error) {
	metrics := &DashboardMetrics{
		DocumentsByStatus: make(map[string]int64),
		ApprovalsByStatus: make(map[string]int64),
		DocumentsByType:   make(map[string]int64),
	}

	// Get all documents based on role
	var documents []db.Document
	var err error

	if userRole == "ADMIN" || userRole == "MANAGER" {
		// Admins and managers see all documents
		documents, err = s.documentRepo.ListDocuments(ctx, 1000000, 0)
	} else {
		// Regular users see only their documents
		documents, err = s.documentRepo.ListDocumentsByCreator(ctx, userID, 1000000, 0)
	}

	if err != nil {
		return nil, err
	}

	metrics.TotalDocuments = int64(len(documents))

	// Count documents by status and type
	for _, doc := range documents {
		metrics.DocumentsByStatus[doc.Status]++
		metrics.DocumentsByType[doc.DocumentType]++
	}

	// Get pending approval tasks
	pendingTasks, err := s.approvalTaskRepo.ListPendingApprovalTasks(ctx, 1000000, 0)
	if err != nil {
		return nil, err
	}

	// Count approvals by status
	now := time.Now()
	for _, task := range pendingTasks {
		metrics.ApprovalsByStatus[task.Status]++

		if task.Status == "PENDING" {
			metrics.PendingApprovals++

			// Check if overdue
			if task.DueDate.Valid && task.DueDate.Time.Before(now) {
				metrics.OverdueApprovals++
			}
		}
	}

	// Get active workflows
	workflows, err := s.workflowRepo.ListActiveWorkflows(ctx, 1000000, 0)
	if err != nil {
		return nil, err
	}

	metrics.ActiveWorkflows = int64(len(workflows))

	// Calculate average approval time from approval history
	// Get all approval tasks to iterate through their history
	allTasks, err := s.approvalTaskRepo.ListPendingApprovalTasks(ctx, 1000000, 0)
	if err != nil {
		return nil, err
	}

	var totalApprovalTime float64
	var approvalCount int64

	for _, task := range allTasks {
		taskID := utils.PgtypeToUUID(task.ID)
		history, err := s.approvalHistoryRepo.ListApprovalHistoryByTask(ctx, taskID)
		if err != nil {
			continue
		}

		if len(history) < 2 {
			continue
		}

		// Find first and last approval
		var firstTime, lastTime time.Time
		for i, h := range history {
			if i == 0 || h.CreatedAt.Time.Before(firstTime) {
				firstTime = h.CreatedAt.Time
			}
			if i == 0 || h.CreatedAt.Time.After(lastTime) {
				lastTime = h.CreatedAt.Time
			}
		}

		duration := lastTime.Sub(firstTime).Hours()
		totalApprovalTime += duration
		approvalCount++
	}

	if approvalCount > 0 {
		metrics.AverageApprovalTime = totalApprovalTime / float64(approvalCount)
	}

	return metrics, nil
}

// GetTrendData retrieves trend data for the last N days
func (s *AnalyticsService) GetTrendData(ctx context.Context, days int32) ([]TrendData, error) {
	if days <= 0 {
		days = 7 // Default to 7 days
	}
	if days > 90 {
		days = 90 // Max 90 days
	}

	trends := make([]TrendData, days)
	now := time.Now()

	// Get all documents
	documents, err := s.documentRepo.ListDocuments(ctx, 1000000, 0)
	if err != nil {
		return nil, err
	}

	// Initialize trends for each day
	for i := int32(0); i < days; i++ {
		date := now.AddDate(0, 0, -int(days-i-1))
		dateStr := date.Format("2006-01-02")

		trends[i] = TrendData{
			Date: dateStr,
		}

		// Count documents created on this date
		for _, doc := range documents {
			if doc.CreatedAt.Time.Format("2006-01-02") == dateStr {
				trends[i].DocumentsCreated++

				if doc.Status == "APPROVED" && doc.UpdatedAt.Time.Format("2006-01-02") == dateStr {
					trends[i].DocumentsApproved++
				}
				if doc.Status == "REJECTED" && doc.UpdatedAt.Time.Format("2006-01-02") == dateStr {
					trends[i].DocumentsRejected++
				}
			}
		}

		// Count approvals completed on this date
		// We'll check approval history for each pending task
		pendingTasks, err := s.approvalTaskRepo.ListPendingApprovalTasks(ctx, 1000000, 0)
		if err == nil {
			for _, task := range pendingTasks {
				taskID := utils.PgtypeToUUID(task.ID)
				history, err := s.approvalHistoryRepo.ListApprovalHistoryByTask(ctx, taskID)
				if err != nil {
					continue
				}

				for _, h := range history {
					if h.CreatedAt.Time.Format("2006-01-02") == dateStr &&
					   (h.Action == "APPROVED" || h.Action == "REJECTED") {
						trends[i].ApprovalsCompleted++
					}
				}
			}
		}
	}

	return trends, nil
}

// GetBottlenecks identifies workflow bottlenecks
func (s *AnalyticsService) GetBottlenecks(ctx context.Context) ([]BottleneckInfo, error) {
	bottlenecks := []BottleneckInfo{}

	// Get all pending tasks
	tasks, err := s.approvalTaskRepo.ListApprovalTasksByStatus(ctx, "PENDING", 1000000, 0)
	if err != nil {
		return nil, err
	}

	// Group tasks by document and stage
	type docStageKey struct {
		DocumentID uuid.UUID
		Stage      int32
	}

	docStageTasks := make(map[docStageKey][]db.ApprovalTask)
	for _, task := range tasks {
		key := docStageKey{
			DocumentID: utils.PgtypeToUUID(task.DocumentID),
			Stage:      task.CurrentStage,
		}
		docStageTasks[key] = append(docStageTasks[key], task)
	}

	// Calculate bottlenecks
	for key, stageTasks := range docStageTasks {
		if len(stageTasks) < 3 {
			// Only consider stages with 3+ pending tasks as potential bottlenecks
			continue
		}

		// Get document info
		doc, err := s.documentRepo.GetDocumentByID(ctx, key.DocumentID)
		if err != nil {
			continue
		}

		// Calculate average time in this stage
		var totalTime float64
		var count int
		now := time.Now()

		for _, task := range stageTasks {
			timeInStage := now.Sub(task.UpdatedAt.Time).Hours()
			totalTime += timeInStage
			count++
		}

		avgTime := totalTime / float64(count)

		bottlenecks = append(bottlenecks, BottleneckInfo{
			DocumentID:         key.DocumentID,
			DocumentType:       doc.DocumentType,
			Stage:              key.Stage,
			PendingTasks:       int64(len(stageTasks)),
			AverageTimeInStage: avgTime,
		})
	}

	return bottlenecks, nil
}
