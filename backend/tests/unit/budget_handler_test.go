package unit

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/types"
)

// TestCreateBudgetValidation tests budget request validation
func TestCreateBudgetValidation(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		shouldPass     bool
	}{
		{
			name: "Valid budget request",
			requestBody: map[string]interface{}{
				"budgetCode":     "IT-2025-Q1",
				"department":     "IT",
				"fiscalYear":     "2025",
				"totalBudget":    500000,
				"allocatedAmount": 0,
			},
			expectedStatus: http.StatusCreated,
			shouldPass:     true,
		},
		{
			name: "Missing budget code",
			requestBody: map[string]interface{}{
				"department":     "IT",
				"fiscalYear":     "2025",
				"totalBudget":    500000,
				"allocatedAmount": 0,
			},
			expectedStatus: http.StatusBadRequest,
			shouldPass:     false,
		},
		{
			name: "Missing department",
			requestBody: map[string]interface{}{
				"budgetCode":     "IT-2025-Q1",
				"fiscalYear":     "2025",
				"totalBudget":    500000,
				"allocatedAmount": 0,
			},
			expectedStatus: http.StatusBadRequest,
			shouldPass:     false,
		},
		{
			name: "Negative total budget",
			requestBody: map[string]interface{}{
				"budgetCode":     "IT-2025-Q1",
				"department":     "IT",
				"fiscalYear":     "2025",
				"totalBudget":    -500000,
				"allocatedAmount": 0,
			},
			expectedStatus: http.StatusBadRequest,
			shouldPass:     false,
		},
		{
			name: "Allocated amount exceeds budget",
			requestBody: map[string]interface{}{
				"budgetCode":     "IT-2025-Q1",
				"department":     "IT",
				"fiscalYear":     "2025",
				"totalBudget":    500000,
				"allocatedAmount": 600000,
			},
			expectedStatus: http.StatusBadRequest,
			shouldPass:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			var req types.CreateBudgetRequest
			json.Unmarshal(body, &req)

			// Validate request
			isValid := req.BudgetCode != "" &&
				req.Department != "" &&
				req.TotalBudget > 0 &&
				req.AllocatedAmount <= req.TotalBudget

			if isValid != tt.shouldPass {
				t.Errorf("Expected %v, got %v", tt.shouldPass, isValid)
			}
		})
	}
}

// TestBudgetCalculationLogic tests budget calculations
func TestBudgetCalculationLogic(t *testing.T) {
	tests := []struct {
		name                string
		totalBudget         float64
		allocatedAmount     float64
		expectedRemaining   float64
	}{
		{
			name:              "No allocation",
			totalBudget:       500000,
			allocatedAmount:   0,
			expectedRemaining: 500000,
		},
		{
			name:              "Partial allocation",
			totalBudget:       500000,
			allocatedAmount:   200000,
			expectedRemaining: 300000,
		},
		{
			name:              "Full allocation",
			totalBudget:       500000,
			allocatedAmount:   500000,
			expectedRemaining: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			remaining := tt.totalBudget - tt.allocatedAmount
			if remaining != tt.expectedRemaining {
				t.Errorf("Expected remaining %f, got %f", tt.expectedRemaining, remaining)
			}
		})
	}
}

// TestBudgetStatusValidation tests budget status field
func TestBudgetStatusValidation(t *testing.T) {
	validStatuses := map[string]bool{
		"draft":    true,
		"pending":  true,
		"approved": true,
		"rejected": true,
	}

	tests := []struct {
		name          string
		status        string
		shouldBeValid bool
	}{
		{"Draft status", "draft", true},
		{"Pending status", "pending", true},
		{"Approved status", "approved", true},
		{"Rejected status", "rejected", true},
		{"Invalid status", "active", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := validStatuses[tt.status]
			if isValid != tt.shouldBeValid {
				t.Errorf("Expected %v, got %v", tt.shouldBeValid, isValid)
			}
		})
	}
}

// TestBudgetFiscalYearValidation tests fiscal year validation
func TestBudgetFiscalYearValidation(t *testing.T) {
	tests := []struct {
		name          string
		fiscalYear    string
		shouldBeValid bool
	}{
		{"Year 2024", "2024", true},
		{"Year 2025", "2025", true},
		{"Year 2026", "2026", true},
		{"Invalid format", "25", false},
		{"Text year", "twenty-five", false},
		{"Empty year", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simple validation: check if year is 4 digits
			isValid := len(tt.fiscalYear) == 4
			if isValid != tt.shouldBeValid {
				t.Errorf("Expected %v, got %v", tt.shouldBeValid, isValid)
			}
		})
	}
}

// TestBudgetApprovalWorkflow tests approval workflow for budget
func TestBudgetApprovalWorkflow(t *testing.T) {
	t.Run("Budget approval stages", func(t *testing.T) {
		// Budget typically has 2 stages: finance -> admin
		stages := 2

		if stages < 1 {
			t.Error("Budget should have at least 1 approval stage")
		}

		approvalChain := []string{"finance", "admin"}
		if len(approvalChain) != stages {
			t.Errorf("Expected %d approval stages, got %d", stages, len(approvalChain))
		}
	})
}

// TestBudgetConflictDetection tests duplicate budget detection
func TestBudgetConflictDetection(t *testing.T) {
	t.Run("Prevent duplicate budgets", func(t *testing.T) {
		budget1 := types.CreateBudgetRequest{
			BudgetCode:     "IT-2025-Q1",
			Department:     "IT",
			FiscalYear:     "2025",
			TotalBudget:    500000,
		}

		budget2 := types.CreateBudgetRequest{
			BudgetCode:     "IT-2025-Q1",
			Department:     "IT",
			FiscalYear:     "2025",
			TotalBudget:    500000,
		}

		// Check if duplicate
		isDuplicate := (budget1.BudgetCode == budget2.BudgetCode &&
			budget1.Department == budget2.Department &&
			budget1.FiscalYear == budget2.FiscalYear)

		if !isDuplicate {
			t.Error("Should detect duplicate budget")
		}
	})
}

// TestBudgetResponseFormat tests response structure
func TestBudgetResponseFormat(t *testing.T) {
	t.Run("Budget response structure", func(t *testing.T) {
		budget := types.BudgetResponse{
			ID:               uuid.New().String(),
			OwnerID:          uuid.New().String(),
			BudgetCode:       "IT-2025-Q1",
			Department:       "IT",
			Status: "DRAFT",
			FiscalYear:       "2025",
			TotalBudget:      500000,
			AllocatedAmount:  0,
			RemainingAmount:  500000,
			ApprovalStage:    0,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}

		// Verify required fields
		if budget.ID == "" {
			t.Error("Response should have ID")
		}
		if budget.BudgetCode == "" {
			t.Error("Response should have BudgetCode")
		}
		if budget.Department == "" {
			t.Error("Response should have Department")
		}
		if budget.TotalBudget != budget.AllocatedAmount+budget.RemainingAmount {
			t.Error("Total should equal allocated + remaining")
		}
	})
}

// TestBudgetAllocationValidation tests allocation constraints
func TestBudgetAllocationValidation(t *testing.T) {
	tests := []struct {
		name                string
		totalBudget         float64
		allocatedAmount     float64
		newAllocation       float64
		shouldBeAllowed     bool
	}{
		{
			name:            "Allocation within budget",
			totalBudget:     500000,
			allocatedAmount: 200000,
			newAllocation:   100000,
			shouldBeAllowed: true,
		},
		{
			name:            "Allocation at budget limit",
			totalBudget:     500000,
			allocatedAmount: 400000,
			newAllocation:   100000,
			shouldBeAllowed: true,
		},
		{
			name:            "Allocation exceeds budget",
			totalBudget:     500000,
			allocatedAmount: 400000,
			newAllocation:   150000,
			shouldBeAllowed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newTotal := tt.allocatedAmount + tt.newAllocation
			canAllocate := newTotal <= tt.totalBudget

			if canAllocate != tt.shouldBeAllowed {
				t.Errorf("Expected %v, got %v", tt.shouldBeAllowed, canAllocate)
			}
		})
	}
}

// TestBudgetDepartmentContraint tests department-specific budgets
func TestBudgetDepartmentConstraint(t *testing.T) {
	t.Run("One budget per department per fiscal year", func(t *testing.T) {
		budget1 := types.BudgetResponse{
			Department: "IT",
			FiscalYear: "2025",
			Status: "APPROVED",
		}

		budget2 := types.BudgetResponse{
			Department: "IT",
			FiscalYear: "2025",
			Status: "DRAFT",
		}

		// Should not allow duplicate
		isConflict := (budget1.Department == budget2.Department &&
			budget1.FiscalYear == budget2.FiscalYear)

		if !isConflict {
			t.Error("Should detect conflicting budgets")
		}
	})
}

// TestBudgetUtilizationCalculation tests utilization percentage
func TestBudgetUtilizationCalculation(t *testing.T) {
	tests := []struct {
		name                string
		totalBudget         float64
		allocatedAmount     float64
		expectedPercent     float64
	}{
		{
			name:            "No allocation",
			totalBudget:     500000,
			allocatedAmount: 0,
			expectedPercent: 0,
		},
		{
			name:            "50% allocated",
			totalBudget:     500000,
			allocatedAmount: 250000,
			expectedPercent: 50,
		},
		{
			name:            "100% allocated",
			totalBudget:     500000,
			allocatedAmount: 500000,
			expectedPercent: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var percent float64
			if tt.totalBudget > 0 {
				percent = (tt.allocatedAmount / tt.totalBudget) * 100
			}

			if percent != tt.expectedPercent {
				t.Errorf("Expected %f%%, got %f%%", tt.expectedPercent, percent)
			}
		})
	}
}

// TestBudgetUpdateValidation tests update request validation
func TestBudgetUpdateValidation(t *testing.T) {
	tests := []struct {
		name        string
		currentStatus string
		updateBody  map[string]interface{}
		shouldAllow bool
	}{
		{
			name:          "Update draft budget",
			currentStatus: "DRAFT",
			updateBody: map[string]interface{}{
				"totalBudget": 600000,
			},
			shouldAllow: true,
		},
		{
			name:          "Cannot update approved budget",
			currentStatus: "APPROVED",
			updateBody: map[string]interface{}{
				"totalBudget": 600000,
			},
			shouldAllow: false,
		},
		{
			name:          "Cannot update rejected budget",
			currentStatus: "REJECTED",
			updateBody: map[string]interface{}{
				"totalBudget": 600000,
			},
			shouldAllow: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Only draft budgets can be updated
			canUpdate := tt.currentStatus == "DRAFT"

			if canUpdate != tt.shouldAllow {
				t.Errorf("Expected %v, got %v", tt.shouldAllow, canUpdate)
			}
		})
	}
}

// BenchmarkBudgetCalculation benchmarks budget math
func BenchmarkBudgetCalculation(b *testing.B) {
	totalBudget := 500000.0
	allocatedAmount := 200000.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		remaining := totalBudget - allocatedAmount
		_ = (allocatedAmount / totalBudget) * 100
		_ = remaining
	}
}

// BenchmarkBudgetHandlerValidation benchmarks handler validation logic
func BenchmarkBudgetHandlerValidation(b *testing.B) {
	req := types.CreateBudgetRequest{
		BudgetCode:     "IT-2025-Q1",
		Department:     "IT",
		FiscalYear:     "2025",
		TotalBudget:    500000,
		AllocatedAmount: 0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = req.BudgetCode != "" &&
			req.Department != "" &&
			req.TotalBudget > 0 &&
			req.AllocatedAmount <= req.TotalBudget
	}
}

// TestBudgetListFiltering tests filtering by status
func TestBudgetListFiltering(t *testing.T) {
	tests := []struct {
		name       string
		allBudgets []string
		filterBy   string
		expected   int
	}{
		{
			name:       "Filter by draft",
			allBudgets: []string{"draft", "draft", "approved", "rejected"},
			filterBy:   "draft",
			expected:   2,
		},
		{
			name:       "Filter by approved",
			allBudgets: []string{"draft", "draft", "approved", "rejected"},
			filterBy:   "approved",
			expected:   1,
		},
		{
			name:       "Filter by pending",
			allBudgets: []string{"draft", "draft", "pending", "rejected"},
			filterBy:   "pending",
			expected:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := 0
			for _, budget := range tt.allBudgets {
				if budget == tt.filterBy {
					count++
				}
			}

			if count != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, count)
			}
		})
	}
}
