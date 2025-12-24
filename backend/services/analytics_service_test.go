package services

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/types"
)

// TestGetStatusCounts tests status counting
func TestGetStatusCounts(t *testing.T) {
	if config.DB == nil {
		t.Skip("Database not initialized")
	}

	// Cleanup
	config.DB.Exec("TRUNCATE requisitions CASCADE")

	// Create test requisitions with different statuses
	statuses := []string{"draft", "pending", "approved", "rejected"}
	for _, status := range statuses {
		for i := 0; i < 2; i++ {
			req := models.Requisition{
				ID:          uuid.New().String(),
				RequesterID: uuid.New().String(),
				Title:       "Test Req",
				Description: "Test Description",
				Department:  "Finance",
				Status:      status,
				Priority:    "high",
				TotalAmount: 100.0,
				Currency:    "USD",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}
			config.DB.Create(&req)
		}
	}

	service := NewAnalyticsService(config.DB)
	params := types.AnalyticsQueryParams{}
	metrics, err := service.GetRequisitionMetrics(params)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if metrics == nil {
		t.Fatal("metrics should not be nil")
	}

	// Verify counts
	expected := int64(2)
	for status, count := range metrics.StatusCounts {
		if count != expected {
			t.Errorf("status %s: expected count %d, got %d", status, expected, count)
		}
	}
}

// TestCalculateRejectionRate tests rejection rate calculation
func TestCalculateRejectionRate(t *testing.T) {
	if config.DB == nil {
		t.Skip("Database not initialized")
	}

	// Cleanup
	config.DB.Exec("TRUNCATE requisitions CASCADE")

	// Create 10 requisitions: 8 approved, 2 rejected
	for i := 0; i < 8; i++ {
		req := models.Requisition{
			ID:          uuid.New().String(),
			RequesterID: uuid.New().String(),
			Title:       "Test Req",
			Description: "Test Description",
			Department:  "Finance",
			Status:      "approved",
			Priority:    "high",
			TotalAmount: 100.0,
			Currency:    "USD",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		config.DB.Create(&req)
	}

	for i := 0; i < 2; i++ {
		req := models.Requisition{
			ID:          uuid.New().String(),
			RequesterID: uuid.New().String(),
			Title:       "Test Req",
			Description: "Test Description",
			Department:  "Finance",
			Status:      "rejected",
			Priority:    "high",
			TotalAmount: 100.0,
			Currency:    "USD",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		config.DB.Create(&req)
	}

	service := NewAnalyticsService(config.DB)
	params := types.AnalyticsQueryParams{}
	metrics, err := service.GetRequisitionMetrics(params)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Expected rejection rate: 2/10 = 20%
	expectedRate := 20.0
	if metrics.RejectionRate != expectedRate {
		t.Errorf("expected rejection rate %f, got %f", expectedRate, metrics.RejectionRate)
	}
}

// TestGetRejectionsOverTime tests time-series rejection data
func TestGetRejectionsOverTime(t *testing.T) {
	if config.DB == nil {
		t.Skip("Database not initialized")
	}

	// Cleanup
	config.DB.Exec("TRUNCATE requisitions CASCADE")

	// Create requisitions on different dates
	now := time.Now()
	for i := 0; i < 3; i++ {
		date := now.AddDate(0, 0, -i)
		for j := 0; j < 2; j++ {
			status := "approved"
			if j == 1 {
				status = "rejected"
			}
			req := models.Requisition{
				ID:          uuid.New().String(),
				RequesterID: uuid.New().String(),
				Title:       "Test Req",
				Description: "Test Description",
				Department:  "Finance",
				Status:      status,
				Priority:    "high",
				TotalAmount: 100.0,
				Currency:    "USD",
				CreatedAt:   date,
				UpdatedAt:   date,
			}
			config.DB.Create(&req)
		}
	}

	service := NewAnalyticsService(config.DB)
	params := types.AnalyticsQueryParams{
		Period: "daily",
	}
	metrics, err := service.GetRequisitionMetrics(params)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(metrics.RejectionsOverTime) == 0 {
		t.Fatal("rejections over time should not be empty")
	}

	// Verify that each date has correct counts
	for _, data := range metrics.RejectionsOverTime {
		if data.Total != 2 {
			t.Errorf("expected total 2, got %d", data.Total)
		}
		if data.Rejections != 1 {
			t.Errorf("expected rejections 1, got %d", data.Rejections)
		}
		if data.Rate != 50.0 {
			t.Errorf("expected rate 50, got %f", data.Rate)
		}
	}
}

// TestGetRejectionReasons tests rejection reason extraction
func TestGetRejectionReasons(t *testing.T) {
	if config.DB == nil {
		t.Skip("Database not initialized")
	}

	// Cleanup
	config.DB.Exec("TRUNCATE requisitions CASCADE")

	// Create a rejected requisition with approval history
	approvalHistory := []types.ApprovalRecord{
		{
			Status:      "rejected",
			ApproverID:  uuid.New().String(),
			ApproverName: "John Approver",
			Comments:    "Budget exceeded",
		},
	}
	historyJSON, _ := json.Marshal(approvalHistory)

	req := models.Requisition{
		ID:              uuid.New().String(),
		RequesterID:     uuid.New().String(),
		Title:           "Test Req",
		Description:     "Test Description",
		Department:      "Finance",
		Status:          "rejected",
		Priority:        "high",
		TotalAmount:     100.0,
		Currency:        "USD",
		ApprovalHistory: historyJSON,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	config.DB.Create(&req)

	service := NewAnalyticsService(config.DB)
	params := types.AnalyticsQueryParams{}
	metrics, err := service.GetRequisitionMetrics(params)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(metrics.RejectionReasons) == 0 {
		t.Fatal("rejection reasons should not be empty")
	}

	// Verify rejection reason
	if metrics.RejectionReasons[0].Reason != "Budget exceeded" {
		t.Errorf("expected reason 'Budget exceeded', got '%s'", metrics.RejectionReasons[0].Reason)
	}
	if metrics.RejectionReasons[0].Count != 1 {
		t.Errorf("expected count 1, got %d", metrics.RejectionReasons[0].Count)
	}
}

// TestGetTopRejectingApprovers tests approver statistics
func TestGetTopRejectingApprovers(t *testing.T) {
	if config.DB == nil {
		t.Skip("Database not initialized")
	}

	// Cleanup
	config.DB.Exec("TRUNCATE requisitions CASCADE")

	approverID := uuid.New().String()

	// Create requisitions with approval history
	for i := 0; i < 5; i++ {
		status := "approved"
		if i > 2 {
			status = "rejected"
		}

		approvalHistory := []types.ApprovalRecord{
			{
				Status:       status,
				ApproverID:   approverID,
				ApproverName: "Jane Approver",
				Comments:     "Checked",
			},
		}
		historyJSON, _ := json.Marshal(approvalHistory)

		req := models.Requisition{
			ID:              uuid.New().String(),
			RequesterID:     uuid.New().String(),
			Title:           "Test Req",
			Description:     "Test Description",
			Department:      "Finance",
			Status:          status,
			Priority:        "high",
			TotalAmount:     100.0,
			Currency:        "USD",
			ApprovalHistory: historyJSON,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}
		config.DB.Create(&req)
	}

	service := NewAnalyticsService(config.DB)
	params := types.AnalyticsQueryParams{}
	metrics, err := service.GetRequisitionMetrics(params)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(metrics.TopRejectingApprovers) == 0 {
		t.Fatal("top rejecting approvers should not be empty")
	}

	// Verify approver stats
	approver := metrics.TopRejectingApprovers[0]
	if approver.ApproverID != approverID {
		t.Errorf("expected approver ID %s, got %s", approverID, approver.ApproverID)
	}
	if approver.Rejections != 2 {
		t.Errorf("expected rejections 2, got %d", approver.Rejections)
	}
	if approver.Approvals != 3 {
		t.Errorf("expected approvals 3, got %d", approver.Approvals)
	}

	// Verify rejection rate: 2/5 = 40%
	expectedRate := 40.0
	if approver.RejectionRate != expectedRate {
		t.Errorf("expected rejection rate %f, got %f", expectedRate, approver.RejectionRate)
	}
}

// TestAnalyticsWithDateRange tests filtering by date range
func TestAnalyticsWithDateRange(t *testing.T) {
	if config.DB == nil {
		t.Skip("Database not initialized")
	}

	// Cleanup
	config.DB.Exec("TRUNCATE requisitions CASCADE")

	// Create requisitions on different dates
	now := time.Now()
	oldDate := now.AddDate(0, 0, -10)
	recentDate := now.AddDate(0, 0, -1)

	// Old requisition
	oldReq := models.Requisition{
		ID:          uuid.New().String(),
		RequesterID: uuid.New().String(),
		Title:       "Old Req",
		Description: "Old",
		Department:  "Finance",
		Status:      "approved",
		Priority:    "high",
		TotalAmount: 100.0,
		Currency:    "USD",
		CreatedAt:   oldDate,
		UpdatedAt:   oldDate,
	}
	config.DB.Create(&oldReq)

	// Recent requisition
	recentReq := models.Requisition{
		ID:          uuid.New().String(),
		RequesterID: uuid.New().String(),
		Title:       "Recent Req",
		Description: "Recent",
		Department:  "Finance",
		Status:      "rejected",
		Priority:    "high",
		TotalAmount: 100.0,
		Currency:    "USD",
		CreatedAt:   recentDate,
		UpdatedAt:   recentDate,
	}
	config.DB.Create(&recentReq)

	// Query for recent requisitions only
	service := NewAnalyticsService(config.DB)
	startDate := now.AddDate(0, 0, -5)
	params := types.AnalyticsQueryParams{
		StartDate: &startDate,
		EndDate:   &now,
	}
	metrics, err := service.GetRequisitionMetrics(params)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should only have recent requisition
	if metrics.TotalRequisitions != 1 {
		t.Errorf("expected 1 total requisition, got %d", metrics.TotalRequisitions)
	}
	if metrics.StatusCounts["rejected"] != 1 {
		t.Errorf("expected 1 rejected, got %d", metrics.StatusCounts["rejected"])
	}
}
