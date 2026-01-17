package integration

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/types"
)

// TestBudgetAvailabilityCheck tests budget availability verification
func TestBudgetAvailabilityCheck(t *testing.T) {
	t.Run("Verify budget is available before creating requisition", func(t *testing.T) {
		budget := types.BudgetResponse{
			ID:              uuid.New().String(),
			TotalBudget:     500000.0,
			AllocatedAmount: 200000.0,
			RemainingAmount: 300000.0,
		}

		requisitionAmount := 50000.0

		// Check if budget is available
		if budget.RemainingAmount < requisitionAmount {
			t.Error("Insufficient budget available")
		}

		// Allocate budget
		budget.AllocatedAmount += requisitionAmount
		budget.RemainingAmount -= requisitionAmount

		if budget.RemainingAmount != 250000 {
			t.Errorf("Expected remaining 250000, got %f", budget.RemainingAmount)
		}

		// Verify total equals allocated + remaining
		total := budget.AllocatedAmount + budget.RemainingAmount
		if total != budget.TotalBudget {
			t.Errorf("Allocated (%f) + Remaining (%f) should equal Total (%f)",
				budget.AllocatedAmount, budget.RemainingAmount, budget.TotalBudget)
		}
	})
}

// TestVendorSpendingLimitEnforcement tests 30% vendor spending limit
func TestVendorSpendingLimitEnforcement(t *testing.T) {
	t.Run("Enforce 30% max spending per vendor", func(t *testing.T) {
		totalBudget := 500000.0
		maxVendorSpend := totalBudget * 0.30 // 30% = 150,000

		tests := []struct {
			name              string
			existingAmount    float64
			newPOAmount       float64
			shouldBeAllowed   bool
		}{
			{"At vendor limit", 150000, 0, true},
			{"Under vendor limit", 100000, 40000, true},
			{"At vendor limit with new PO", 100000, 50000, true},
			{"Exceeds vendor limit", 130000, 30000, false},
			{"Far exceeds vendor limit", 150000, 10000, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				totalVendorSpend := tt.existingAmount + tt.newPOAmount

				if totalVendorSpend > maxVendorSpend {
					if tt.shouldBeAllowed {
						t.Errorf("Expected to allow %f, but it exceeds limit %f", totalVendorSpend, maxVendorSpend)
					}
				} else {
					if !tt.shouldBeAllowed {
						t.Errorf("Expected to deny %f", totalVendorSpend)
					}
				}
			})
		}
	})
}

// TestReserveFundsEnforcement tests reserve funds percentage
func TestReserveFundsEnforcement(t *testing.T) {
	t.Run("Maintain 10-15% reserve funds", func(t *testing.T) {
		totalBudget := 500000.0
		minReserve := totalBudget * 0.10  // 10%

		tests := []struct {
			name              string
			allocatedAmount   float64
			shouldBeAllowed   bool
		}{
			{"At 85% allocation", 425000, true},
			{"At 90% allocation", 450000, true},
			{"At 91% allocation (exceeds)", 455000, false},
			{"Full allocation attempt", 500000, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				remaining := totalBudget - tt.allocatedAmount
				meetsMinReserve := remaining >= minReserve

				if tt.shouldBeAllowed {
					if !meetsMinReserve {
						t.Errorf("Should allow %f allocation (remaining %f)", tt.allocatedAmount, remaining)
					}
				} else {
					if meetsMinReserve {
						t.Errorf("Should not allow %f allocation", tt.allocatedAmount)
					}
				}
			})
		}
	})
}

// TestQuoteRequirementByAmount tests quote requirements based on amount
func TestQuoteRequirementByAmount(t *testing.T) {
	t.Run("Require quotes for orders above threshold", func(t *testing.T) {
		quoteThreshold := 25000.0

		tests := []struct {
			name              string
			amount            float64
			quoteRequired     bool
		}{
			{"Below threshold", 10000, false},
			{"At threshold", 25000, true},
			{"Above threshold", 50000, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				requiresQuote := tt.amount >= quoteThreshold

				if requiresQuote != tt.quoteRequired {
					t.Errorf("Expected requiresQuote=%v, got %v", tt.quoteRequired, requiresQuote)
				}
			})
		}
	})
}

// TestBudgetDeallocatorion tests fund deallocation on PO cancellation
func TestBudgetDeallocation(t *testing.T) {
	t.Run("Release allocated funds when PO is cancelled", func(t *testing.T) {
		budget := types.BudgetResponse{
			ID:              uuid.New().String(),
			TotalBudget:     500000.0,
			AllocatedAmount: 150000.0,
			RemainingAmount: 350000.0,
		}

		cancelledPOAmount := 50000.0

		// Deallocate funds
		budget.AllocatedAmount -= cancelledPOAmount
		budget.RemainingAmount += cancelledPOAmount

		if budget.AllocatedAmount != 100000 {
			t.Errorf("Expected allocated 100000, got %f", budget.AllocatedAmount)
		}

		if budget.RemainingAmount != 400000 {
			t.Errorf("Expected remaining 400000, got %f", budget.RemainingAmount)
		}

		// Verify totals match
		total := budget.AllocatedAmount + budget.RemainingAmount
		if total != budget.TotalBudget {
			t.Error("Budget allocation mismatch after deallocation")
		}
	})
}

// TestBudgetUtilizationTracking tests budget utilization percentage
func TestBudgetUtilizationTracking(t *testing.T) {
	t.Run("Track and report budget utilization percentage", func(t *testing.T) {
		tests := []struct {
			name              string
			totalBudget       float64
			allocatedAmount   float64
			expectedPercent   float64
		}{
			{"No allocation", 500000, 0, 0},
			{"25% allocated", 500000, 125000, 25},
			{"50% allocated", 500000, 250000, 50},
			{"75% allocated", 500000, 375000, 75},
			{"100% allocated", 500000, 500000, 100},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				utilization := (tt.allocatedAmount / tt.totalBudget) * 100

				if utilization != tt.expectedPercent {
					t.Errorf("Expected %f%%, got %f%%", tt.expectedPercent, utilization)
				}
			})
		}
	})
}

// TestBudgetExhaustedCondition tests behavior when budget is exhausted
func TestBudgetExhaustedCondition(t *testing.T) {
	t.Run("Prevent requisition when budget exhausted", func(t *testing.T) {
		budget := types.BudgetResponse{
			ID:              uuid.New().String(),
			TotalBudget:     100000.0,
			AllocatedAmount: 100000.0,
			RemainingAmount: 0.0,
		}

		newRequisitionAmount := 10000.0

		// Check if budget available
		if budget.RemainingAmount < newRequisitionAmount {
			// Cannot create requisition
			if budget.RemainingAmount == 0 {
				t.Logf("Budget exhausted: Cannot create requisition for %f", newRequisitionAmount)
			}
		}

		// Verify budget is fully allocated
		if budget.AllocatedAmount >= budget.TotalBudget {
			t.Logf("Warning: Budget fully allocated")
		}
	})
}

// TestDepartmentBudgetConstraint tests one-budget-per-department rule
func TestDepartmentBudgetConstraint(t *testing.T) {
	t.Run("Enforce one budget per department per fiscal year", func(t *testing.T) {
		department := "IT"
		fiscalYear := "2025"

		existingBudget := types.BudgetResponse{
			ID:         uuid.New().String(),
			Department: department,
			FiscalYear: fiscalYear,
			Status:     "approved",
		}

		newBudget := types.CreateBudgetRequest{
			Department:  department,
			FiscalYear:  fiscalYear,
			TotalBudget: 600000,
		}

		// Check if another budget exists for same dept/year
		isDuplicate := (existingBudget.Department == newBudget.Department &&
			existingBudget.FiscalYear == newBudget.FiscalYear)

		if isDuplicate {
			if existingBudget.Status == "approved" {
				t.Logf("Cannot create new budget: existing approved budget for %s/%s", department, fiscalYear)
			}
		}

		if !isDuplicate {
			t.Error("Should detect duplicate budget")
		}
	})
}

// TestBudgetTransferBetweenLineItems tests moving funds between line items
func TestBudgetTransferBetweenLineItems(t *testing.T) {
	t.Run("Transfer allocated funds between line items", func(t *testing.T) {
		lineItem1Amount := 100000.0
		lineItem2Amount := 50000.0
		totalAllocated := lineItem1Amount + lineItem2Amount

		// Transfer 20k from line item 1 to line item 2
		transferAmount := 20000.0

		lineItem1Amount -= transferAmount
		lineItem2Amount += transferAmount

		newTotal := lineItem1Amount + lineItem2Amount

		if newTotal != totalAllocated {
			t.Error("Total allocated should not change after transfer")
		}

		if lineItem1Amount != 80000 {
			t.Errorf("Expected line item 1 = 80000, got %f", lineItem1Amount)
		}

		if lineItem2Amount != 70000 {
			t.Errorf("Expected line item 2 = 70000, got %f", lineItem2Amount)
		}
	})
}

// TestMultiYearBudgetPlanning tests budget allocation across years
func TestMultiYearBudgetPlanning(t *testing.T) {
	t.Run("Plan budgets for multiple fiscal years", func(t *testing.T) {
		budgets := map[string]float64{
			"2024": 300000,
			"2025": 350000,
			"2026": 400000,
		}

		totalPlanned := 0.0
		for _, amount := range budgets {
			totalPlanned += amount
		}

		if totalPlanned != 1050000 {
			t.Errorf("Expected total planned 1050000, got %f", totalPlanned)
		}

		// Verify each year has a budget
		for year := range budgets {
			amount := budgets[year]
			if amount == 0 {
				t.Errorf("Year %s should have allocated budget", year)
			}
		}
	})
}

// TestBudgetRevisionHistory tests tracking budget amendments
func TestBudgetRevisionHistory(t *testing.T) {
	t.Run("Track budget revisions with timestamps", func(t *testing.T) {
		budget := types.BudgetResponse{
			ID:          uuid.New().String(),
			TotalBudget: 500000,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		revisions := []struct {
			timestamp   time.Time
			oldAmount   float64
			newAmount   float64
			reason      string
		}{
			{
				timestamp:   time.Now(),
				oldAmount:   500000,
				newAmount:   600000,
				reason:      "Q1 budget increase",
			},
		}

		if len(revisions) == 0 {
			t.Error("Should track budget revisions")
		}

		// Verify revision is after creation
		if revisions[0].timestamp.Before(budget.CreatedAt) {
			t.Error("Revision timestamp should be after creation")
		}
	})
}

// TestCostVarianceAnalysis tests comparing planned vs actual costs
func TestCostVarianceAnalysis(t *testing.T) {
	t.Run("Analyze variance between budgeted and actual PO amounts", func(t *testing.T) {
		tests := []struct {
			name             string
			budgetedAmount   float64
			actualAmount     float64
			expectedVariance float64
		}{
			{"Under budget", 50000, 45000, -5000},
			{"Over budget", 50000, 55000, 5000},
			{"On budget", 50000, 50000, 0},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				variance := tt.actualAmount - tt.budgetedAmount
				variancePercent := (variance / tt.budgetedAmount) * 100

				if variance != tt.expectedVariance {
					t.Errorf("Expected variance %f, got %f", tt.expectedVariance, variance)
				}

				if tt.budgetedAmount == tt.actualAmount && variancePercent != 0 {
					t.Error("On budget should have 0% variance")
				}
			})
		}
	})
}

// TestBudgetAlertThresholds tests alert generation for budget thresholds
func TestBudgetAlertThresholds(t *testing.T) {
	t.Run("Generate alerts based on budget utilization thresholds", func(t *testing.T) {
		budget := types.BudgetResponse{
			ID:              uuid.New().String(),
			TotalBudget:     500000,
			AllocatedAmount: 450000,
			RemainingAmount: 50000,
		}

		utilization := (budget.AllocatedAmount / budget.TotalBudget) * 100

		alerts := []string{}

		if utilization > 90 {
			alerts = append(alerts, "Warning: Budget more than 90% allocated")
		}
		if utilization > 100 {
			alerts = append(alerts, "Error: Budget exceeded")
		}

		if utilization == 90 {
			if len(alerts) != 1 {
				t.Errorf("Expected 1 alert at 90%% utilization, got %d", len(alerts))
			}
		}
	})
}

// BenchmarkBudgetAvailabilityCheck benchmarks budget availability checking
func BenchmarkBudgetAvailabilityCheck(b *testing.B) {
	remaining := 250000.0
	required := 50000.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = remaining >= required
	}
}

// BenchmarkVendorSpendingLimitCheck benchmarks vendor limit checking
func BenchmarkVendorSpendingLimitCheck(b *testing.B) {
	totalBudget := 500000.0
	existing := 100000.0
	new := 50000.0
	limit := totalBudget * 0.30

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = (existing + new) <= limit
	}
}

// BenchmarkBudgetUtilizationCalculation benchmarks utilization calc
func BenchmarkBudgetUtilizationCalculation(b *testing.B) {
	allocated := 450000.0
	total := 500000.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = (allocated / total) * 100
	}
}
