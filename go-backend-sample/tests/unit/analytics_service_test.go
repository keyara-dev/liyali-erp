package unit

import (
	"context"
	"testing"
	"time"

	"github.com/cozyCodr/liyali-gateway/internal/db"
	"github.com/cozyCodr/liyali-gateway/internal/services"
	"github.com/cozyCodr/liyali-gateway/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock repositories
type MockDocumentRepository struct {
	mock.Mock
}

func (m *MockDocumentRepository) ListDocuments(ctx context.Context, limit, offset int32) ([]db.Document, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]db.Document), args.Error(1)
}

func (m *MockDocumentRepository) ListDocumentsByCreator(ctx context.Context, creatorID uuid.UUID, limit, offset int32) ([]db.Document, error) {
	args := m.Called(ctx, creatorID, limit, offset)
	return args.Get(0).([]db.Document), args.Error(1)
}

func (m *MockDocumentRepository) GetDocumentByID(ctx context.Context, id uuid.UUID) (*db.Document, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*db.Document), args.Error(1)
}

type MockApprovalTaskRepository struct {
	mock.Mock
}

func (m *MockApprovalTaskRepository) ListPendingApprovalTasks(ctx context.Context, limit, offset int32) ([]db.ApprovalTask, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]db.ApprovalTask), args.Error(1)
}

func (m *MockApprovalTaskRepository) ListApprovalTasksByStatus(ctx context.Context, status string, limit, offset int32) ([]db.ApprovalTask, error) {
	args := m.Called(ctx, status, limit, offset)
	return args.Get(0).([]db.ApprovalTask), args.Error(1)
}

type MockApprovalHistoryRepository struct {
	mock.Mock
}

func (m *MockApprovalHistoryRepository) ListApprovalHistoryByTask(ctx context.Context, taskID uuid.UUID) ([]db.ApprovalHistory, error) {
	args := m.Called(ctx, taskID)
	return args.Get(0).([]db.ApprovalHistory), args.Error(1)
}

type MockWorkflowRepository struct {
	mock.Mock
}

func (m *MockWorkflowRepository) ListActiveWorkflows(ctx context.Context, limit, offset int32) ([]db.Workflow, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]db.Workflow), args.Error(1)
}

func TestAnalyticsService_GetDashboardMetrics(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("success - admin user", func(t *testing.T) {
		// Setup mocks
		mockDocRepo := new(MockDocumentRepository)
		mockTaskRepo := new(MockApprovalTaskRepository)
		mockHistoryRepo := new(MockApprovalHistoryRepository)
		mockWorkflowRepo := new(MockWorkflowRepository)

		// Mock data
		documents := []db.Document{
			{
				ID:           utils.UUIDToPgtype(uuid.New()),
				DocumentType: "REQUISITION",
				Status:       "DRAFT",
			},
			{
				ID:           utils.UUIDToPgtype(uuid.New()),
				DocumentType: "BUDGET",
				Status:       "SUBMITTED",
			},
			{
				ID:           utils.UUIDToPgtype(uuid.New()),
				DocumentType: "REQUISITION",
				Status:       "APPROVED",
			},
		}

		pendingTasks := []db.ApprovalTask{
			{
				ID:     utils.UUIDToPgtype(uuid.New()),
				Status: "PENDING",
				DueDate: pgtype.Timestamp{
					Time:  time.Now().Add(24 * time.Hour),
					Valid: true,
				},
			},
		}

		workflows := []db.Workflow{
			{
				ID:       utils.UUIDToPgtype(uuid.New()),
				IsActive: pgtype.Bool{Bool: true, Valid: true},
			},
			{
				ID:       utils.UUIDToPgtype(uuid.New()),
				IsActive: pgtype.Bool{Bool: true, Valid: true},
			},
		}

		// Set expectations
		mockDocRepo.On("ListDocuments", ctx, int32(1000000), int32(0)).Return(documents, nil)
		mockTaskRepo.On("ListPendingApprovalTasks", ctx, int32(1000000), int32(0)).Return(pendingTasks, nil)
		mockWorkflowRepo.On("ListActiveWorkflows", ctx, int32(1000000), int32(0)).Return(workflows, nil)

		// Create service
		service := services.NewAnalyticsService(*mockDocRepo, *mockTaskRepo, *mockHistoryRepo, *mockWorkflowRepo)

		// Execute
		metrics, err := service.GetDashboardMetrics(ctx, userID, "ADMIN")

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, metrics)
		assert.Equal(t, int64(3), metrics.TotalDocuments)
		assert.Equal(t, int64(1), metrics.DocumentsByStatus["DRAFT"])
		assert.Equal(t, int64(1), metrics.DocumentsByStatus["SUBMITTED"])
		assert.Equal(t, int64(1), metrics.DocumentsByStatus["APPROVED"])
		assert.Equal(t, int64(2), metrics.DocumentsByType["REQUISITION"])
		assert.Equal(t, int64(1), metrics.DocumentsByType["BUDGET"])
		assert.Equal(t, int64(1), metrics.PendingApprovals)
		assert.Equal(t, int64(0), metrics.OverdueApprovals)
		assert.Equal(t, int64(2), metrics.ActiveWorkflows)

		mockDocRepo.AssertExpectations(t)
		mockTaskRepo.AssertExpectations(t)
		mockWorkflowRepo.AssertExpectations(t)
	})

	t.Run("success - regular user", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockTaskRepo := new(MockApprovalTaskRepository)
		mockHistoryRepo := new(MockApprovalHistoryRepository)
		mockWorkflowRepo := new(MockWorkflowRepository)

		documents := []db.Document{
			{
				ID:           utils.UUIDToPgtype(uuid.New()),
				DocumentType: "REQUISITION",
				Status:       "DRAFT",
			},
		}

		mockDocRepo.On("ListDocumentsByCreator", ctx, userID, int32(1000000), int32(0)).Return(documents, nil)
		mockTaskRepo.On("ListPendingApprovalTasks", ctx, int32(1000000), int32(0)).Return([]db.ApprovalTask{}, nil)
		mockWorkflowRepo.On("ListActiveWorkflows", ctx, int32(1000000), int32(0)).Return([]db.Workflow{}, nil)

		service := services.NewAnalyticsService(*mockDocRepo, *mockTaskRepo, *mockHistoryRepo, *mockWorkflowRepo)

		metrics, err := service.GetDashboardMetrics(ctx, userID, "USER")

		assert.NoError(t, err)
		assert.NotNil(t, metrics)
		assert.Equal(t, int64(1), metrics.TotalDocuments)

		mockDocRepo.AssertExpectations(t)
	})

	t.Run("detects overdue tasks", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockTaskRepo := new(MockApprovalTaskRepository)
		mockHistoryRepo := new(MockApprovalHistoryRepository)
		mockWorkflowRepo := new(MockWorkflowRepository)

		overdueTasks := []db.ApprovalTask{
			{
				ID:     utils.UUIDToPgtype(uuid.New()),
				Status: "PENDING",
				DueDate: pgtype.Timestamp{
					Time:  time.Now().Add(-24 * time.Hour), // Overdue
					Valid: true,
				},
			},
			{
				ID:     utils.UUIDToPgtype(uuid.New()),
				Status: "PENDING",
				DueDate: pgtype.Timestamp{
					Time:  time.Now().Add(-48 * time.Hour), // Overdue
					Valid: true,
				},
			},
		}

		mockDocRepo.On("ListDocuments", ctx, int32(1000000), int32(0)).Return([]db.Document{}, nil)
		mockTaskRepo.On("ListPendingApprovalTasks", ctx, int32(1000000), int32(0)).Return(overdueTasks, nil)
		mockWorkflowRepo.On("ListActiveWorkflows", ctx, int32(1000000), int32(0)).Return([]db.Workflow{}, nil)

		service := services.NewAnalyticsService(*mockDocRepo, *mockTaskRepo, *mockHistoryRepo, *mockWorkflowRepo)

		metrics, err := service.GetDashboardMetrics(ctx, userID, "ADMIN")

		assert.NoError(t, err)
		assert.Equal(t, int64(2), metrics.PendingApprovals)
		assert.Equal(t, int64(2), metrics.OverdueApprovals)
	})
}

func TestAnalyticsService_GetTrendData(t *testing.T) {
	ctx := context.Background()

	t.Run("success - 7 days", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockTaskRepo := new(MockApprovalTaskRepository)
		mockHistoryRepo := new(MockApprovalHistoryRepository)
		mockWorkflowRepo := new(MockWorkflowRepository)

		documents := []db.Document{
			{
				ID:           utils.UUIDToPgtype(uuid.New()),
				DocumentType: "REQUISITION",
				Status:       "APPROVED",
				CreatedAt:    pgtype.Timestamp{Time: time.Now(), Valid: true},
				UpdatedAt:    pgtype.Timestamp{Time: time.Now(), Valid: true},
			},
		}

		mockDocRepo.On("ListDocuments", ctx, int32(1000000), int32(0)).Return(documents, nil)
		mockTaskRepo.On("ListPendingApprovalTasks", ctx, int32(1000000), int32(0)).Return([]db.ApprovalTask{}, nil)

		service := services.NewAnalyticsService(*mockDocRepo, *mockTaskRepo, *mockHistoryRepo, *mockWorkflowRepo)

		trends, err := service.GetTrendData(ctx, 7)

		assert.NoError(t, err)
		assert.Len(t, trends, 7)
		assert.Equal(t, int64(1), trends[6].DocumentsCreated) // Today is the last index
	})

	t.Run("defaults to 7 days if invalid", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockTaskRepo := new(MockApprovalTaskRepository)
		mockHistoryRepo := new(MockApprovalHistoryRepository)
		mockWorkflowRepo := new(MockWorkflowRepository)

		mockDocRepo.On("ListDocuments", ctx, int32(1000000), int32(0)).Return([]db.Document{}, nil)
		mockTaskRepo.On("ListPendingApprovalTasks", ctx, int32(1000000), int32(0)).Return([]db.ApprovalTask{}, nil)

		service := services.NewAnalyticsService(*mockDocRepo, *mockTaskRepo, *mockHistoryRepo, *mockWorkflowRepo)

		trends, err := service.GetTrendData(ctx, 0)

		assert.NoError(t, err)
		assert.Len(t, trends, 7)
	})

	t.Run("caps at 90 days", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockTaskRepo := new(MockApprovalTaskRepository)
		mockHistoryRepo := new(MockApprovalHistoryRepository)
		mockWorkflowRepo := new(MockWorkflowRepository)

		mockDocRepo.On("ListDocuments", ctx, int32(1000000), int32(0)).Return([]db.Document{}, nil)
		mockTaskRepo.On("ListPendingApprovalTasks", ctx, int32(1000000), int32(0)).Return([]db.ApprovalTask{}, nil)

		service := services.NewAnalyticsService(*mockDocRepo, *mockTaskRepo, *mockHistoryRepo, *mockWorkflowRepo)

		trends, err := service.GetTrendData(ctx, 200)

		assert.NoError(t, err)
		assert.Len(t, trends, 90)
	})
}

func TestAnalyticsService_GetBottlenecks(t *testing.T) {
	ctx := context.Background()

	t.Run("identifies bottlenecks", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockTaskRepo := new(MockApprovalTaskRepository)
		mockHistoryRepo := new(MockApprovalHistoryRepository)
		mockWorkflowRepo := new(MockWorkflowRepository)

		docID := uuid.New()

		// Create 5 pending tasks for the same document/stage (bottleneck)
		pendingTasks := make([]db.ApprovalTask, 5)
		for i := 0; i < 5; i++ {
			pendingTasks[i] = db.ApprovalTask{
				ID:           utils.UUIDToPgtype(uuid.New()),
				DocumentID:   utils.UUIDToPgtype(docID),
				Status:       "PENDING",
				CurrentStage: 2,
				UpdatedAt:    pgtype.Timestamp{Time: time.Now().Add(-48 * time.Hour), Valid: true},
			}
		}

		document := &db.Document{
			ID:           utils.UUIDToPgtype(docID),
			DocumentType: "BUDGET",
			Status:       "IN_REVIEW",
		}

		mockTaskRepo.On("ListApprovalTasksByStatus", ctx, "PENDING", int32(1000000), int32(0)).Return(pendingTasks, nil)
		mockDocRepo.On("GetDocumentByID", ctx, docID).Return(document, nil)

		service := services.NewAnalyticsService(*mockDocRepo, *mockTaskRepo, *mockHistoryRepo, *mockWorkflowRepo)

		bottlenecks, err := service.GetBottlenecks(ctx)

		assert.NoError(t, err)
		assert.Len(t, bottlenecks, 1)
		assert.Equal(t, docID, bottlenecks[0].DocumentID)
		assert.Equal(t, "BUDGET", bottlenecks[0].DocumentType)
		assert.Equal(t, int32(2), bottlenecks[0].Stage)
		assert.Equal(t, int64(5), bottlenecks[0].PendingTasks)
		assert.Greater(t, bottlenecks[0].AverageTimeInStage, 40.0) // Should be ~48 hours
	})

	t.Run("ignores stages with less than 3 tasks", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockTaskRepo := new(MockApprovalTaskRepository)
		mockHistoryRepo := new(MockApprovalHistoryRepository)
		mockWorkflowRepo := new(MockWorkflowRepository)

		// Only 2 tasks - should not be considered a bottleneck
		pendingTasks := []db.ApprovalTask{
			{
				ID:           utils.UUIDToPgtype(uuid.New()),
				DocumentID:   utils.UUIDToPgtype(uuid.New()),
				Status:       "PENDING",
				CurrentStage: 1,
			},
			{
				ID:           utils.UUIDToPgtype(uuid.New()),
				DocumentID:   utils.UUIDToPgtype(uuid.New()),
				Status:       "PENDING",
				CurrentStage: 1,
			},
		}

		mockTaskRepo.On("ListApprovalTasksByStatus", ctx, "PENDING", int32(1000000), int32(0)).Return(pendingTasks, nil)

		service := services.NewAnalyticsService(*mockDocRepo, *mockTaskRepo, *mockHistoryRepo, *mockWorkflowRepo)

		bottlenecks, err := service.GetBottlenecks(ctx)

		assert.NoError(t, err)
		assert.Len(t, bottlenecks, 0)
	})

	t.Run("no bottlenecks with no pending tasks", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockTaskRepo := new(MockApprovalTaskRepository)
		mockHistoryRepo := new(MockApprovalHistoryRepository)
		mockWorkflowRepo := new(MockWorkflowRepository)

		mockTaskRepo.On("ListApprovalTasksByStatus", ctx, "PENDING", int32(1000000), int32(0)).Return([]db.ApprovalTask{}, nil)

		service := services.NewAnalyticsService(*mockDocRepo, *mockTaskRepo, *mockHistoryRepo, *mockWorkflowRepo)

		bottlenecks, err := service.GetBottlenecks(ctx)

		assert.NoError(t, err)
		assert.Len(t, bottlenecks, 0)
	})
}
