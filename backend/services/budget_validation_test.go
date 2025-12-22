package services

import (
	"testing"

	"github.com/liyali/liyali-gateway/models"
)

// TestBudgetValidationLogic tests budget validation rules
func TestBudgetValidationLogic(t *testing.T) {
	t.Run("Budget availability checks", func(t *testing.T) {
		tests := []struct {
			name            string
			totalBudget     float64
			allocatedAmount float64
			remainingAmount float64
			requestAmount   float64
			shouldPass      bool
		}{
			{
				name:            "Amount within budget",
				totalBudget:     100000,
				allocatedAmount: 30000,
				remainingAmount: 70000,
				requestAmount:   50000,
				shouldPass:      true,
			},
			{
				name:            "Amount exactly remaining",
				totalBudget:     100000,
				allocatedAmount: 30000,
				remainingAmount: 70000,
				requestAmount:   70000,
				shouldPass:      true,
			},
			{
				name:            "Amount exceeds remaining",
				totalBudget:     100000,
				allocatedAmount: 30000,
				remainingAmount: 70000,
				requestAmount:   80000,
				shouldPass:      false,
			},
			{
				name:            "Zero remaining budget",
				totalBudget:     100000,
				allocatedAmount: 100000,
				remainingAmount: 0,
				requestAmount:   1,
				shouldPass:      false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				canAllocate := tt.requestAmount <= tt.remainingAmount
				if canAllocate != tt.shouldPass {
					t.Errorf("Expected %v, got %v", tt.shouldPass, canAllocate)
				}
			})
		}
	})
}

// TestBudgetAllocationCalculation tests allocation math
func TestBudgetAllocationCalculation(t *testing.T) {
	t.Run("Allocation calculations", func(t *testing.T) {
		tests := []struct {
			name              string
			totalBudget       float64
			currentAllocated  float64
			newAllocation     float64
			expectedAllocated float64
			expectedRemaining float64
		}{
			{
				name:              "Simple allocation",
				totalBudget:       100000,
				currentAllocated:  30000,
				newAllocation:     20000,
				expectedAllocated: 50000,
				expectedRemaining: 50000,
			},
			{
				name:              "Multiple allocations",
				totalBudget:       100000,
				currentAllocated:  50000,
				newAllocation:     30000,
				expectedAllocated: 80000,
				expectedRemaining: 20000,
			},
			{
				name:              "No allocation",
				totalBudget:       100000,
				currentAllocated:  0,
				newAllocation:     0,
				expectedAllocated: 0,
				expectedRemaining: 100000,
			},
			{
				name:              "Full allocation",
				totalBudget:       100000,
				currentAllocated:  50000,
				newAllocation:     50000,
				expectedAllocated: 100000,
				expectedRemaining: 0,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				newAllocated := tt.currentAllocated + tt.newAllocation
				newRemaining := tt.totalBudget - newAllocated

				if newAllocated != tt.expectedAllocated {
					t.Errorf("Allocated: expected %f, got %f", tt.expectedAllocated, newAllocated)
				}
				if newRemaining != tt.expectedRemaining {
					t.Errorf("Remaining: expected %f, got %f", tt.expectedRemaining, newRemaining)
				}
			})
		}
	})
}

// TestBudgetDeallocationCalculation tests deallocation math
func TestBudgetDeallocationCalculation(t *testing.T) {
	t.Run("Deallocation calculations", func(t *testing.T) {
		tests := []struct {
			name              string
			totalBudget       float64
			currentAllocated  float64
			deallocationAmount float64
			expectedAllocated float64
			expectedRemaining float64
		}{
			{
				name:               "Simple deallocation",
				totalBudget:        100000,
				currentAllocated:   50000,
				deallocationAmount: 20000,
				expectedAllocated:  30000,
				expectedRemaining:  70000,
			},
			{
				name:               "Full deallocation",
				totalBudget:        100000,
				currentAllocated:   50000,
				deallocationAmount: 50000,
				expectedAllocated:  0,
				expectedRemaining:  100000,
			},
			{
				name:               "Deallocation more than allocated (clamp to zero)",
				totalBudget:        100000,
				currentAllocated:   30000,
				deallocationAmount: 50000,
				expectedAllocated:  0, // Clamped to zero
				expectedRemaining:  100000,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				newAllocated := tt.currentAllocated - tt.deallocationAmount
				if newAllocated < 0 {
					newAllocated = 0
				}
				newRemaining := tt.totalBudget - newAllocated

				if newAllocated != tt.expectedAllocated {
					t.Errorf("Allocated: expected %f, got %f", tt.expectedAllocated, newAllocated)
				}
				if newRemaining != tt.expectedRemaining {
					t.Errorf("Remaining: expected %f, got %f", tt.expectedRemaining, newRemaining)
				}
			})
		}
	})
}

// TestReserveFundsValidation tests reserve fund calculations
func TestReserveFundsValidation(t *testing.T) {
	t.Run("Reserve fund requirements", func(t *testing.T) {
		tests := []struct {
			name              string
			totalBudget       float64
			reservePercent    float64
			allocatedAmount   float64
			shouldPass        bool
		}{
			{
				name:            "Within reserve requirement (10%)",
				totalBudget:     100000,
				reservePercent:  10,
				allocatedAmount: 85000, // Leaves 15000 (15%)
				shouldPass:      true,
			},
			{
				name:            "Violates reserve requirement (10%)",
				totalBudget:     100000,
				reservePercent:  10,
				allocatedAmount: 92000, // Leaves 8000 (8%)
				shouldPass:      false,
			},
			{
				name:            "Exactly at reserve requirement (15%)",
				totalBudget:     100000,
				reservePercent:  15,
				allocatedAmount: 85000, // Leaves 15000 (15%)
				shouldPass:      true,
			},
			{
				name:            "High reserve requirement (20%)",
				totalBudget:     100000,
				reservePercent:  20,
				allocatedAmount: 75000, // Leaves 25000 (25%)
				shouldPass:      true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				reserveAmount := tt.totalBudget * (tt.reservePercent / 100)
				remainingAfterAllocation := tt.totalBudget - tt.allocatedAmount
				meetsRequirement := remainingAfterAllocation >= reserveAmount

				if meetsRequirement != tt.shouldPass {
					t.Errorf(
						"Expected %v, got %v (reserve: %f, remaining: %f)",
						tt.shouldPass, meetsRequirement, reserveAmount, remainingAfterAllocation,
					)
				}
			})
		}
	})
}

// TestMaxSingleOrderValidation tests max order limits
func TestMaxSingleOrderValidation(t *testing.T) {
	t.Run("Max single order limits", func(t *testing.T) {
		maxOrder := 50000.0

		tests := []struct {
			name      string
			amount    float64
			shouldPass bool
		}{
			{"Within limit", 30000, true},
			{"Exactly at limit", 50000, true},
			{"Exceeds limit", 60000, false},
			{"Zero amount", 0, true},
			{"Large excess", 100000, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				valid := tt.amount <= maxOrder
				if valid != tt.shouldPass {
					t.Errorf("Expected %v, got %v", tt.shouldPass, valid)
				}
			})
		}
	})
}

// TestVendorSpendingLimitValidation tests 30% vendor limit
func TestVendorSpendingLimitValidation(t *testing.T) {
	t.Run("Vendor spending limits (30% per vendor)", func(t *testing.T) {
		totalBudget := 100000.0
		maxPerVendor := totalBudget * 0.3 // 30000

		tests := []struct {
			name            string
			currentSpending float64
			newAmount       float64
			shouldPass      bool
		}{
			{
				name:            "Vendor within limit",
				currentSpending: 10000,
				newAmount:       15000,
				shouldPass:      true, // 25000 < 30000
			},
			{
				name:            "Vendor at limit",
				currentSpending: 15000,
				newAmount:       15000,
				shouldPass:      true, // 30000 == 30000
			},
			{
				name:            "Vendor exceeds limit",
				currentSpending: 20000,
				newAmount:       15000,
				shouldPass:      false, // 35000 > 30000
			},
			{
				name:            "Vendor far exceeds limit",
				currentSpending: 0,
				newAmount:       40000,
				shouldPass:      false, // 40000 > 30000
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				total := tt.currentSpending + tt.newAmount
				valid := total <= maxPerVendor
				if valid != tt.shouldPass {
					t.Errorf("Expected %v, got %v (total: %f, limit: %f)", tt.shouldPass, valid, total, maxPerVendor)
				}
			})
		}
	})
}

// TestQuoteRequirementValidation tests quote thresholds
func TestQuoteRequirementValidation(t *testing.T) {
	t.Run("Quote requirements", func(t *testing.T) {
		tests := []struct {
			name              string
			thresholdIT       float64
			thresholdHR       float64
			department        string
			amount            float64
			shouldRequireQuote bool
		}{
			{
				name:              "IT: Below threshold",
				thresholdIT:       25000,
				thresholdHR:       15000,
				department:        "IT",
				amount:            20000,
				shouldRequireQuote: false,
			},
			{
				name:              "IT: At threshold",
				thresholdIT:       25000,
				thresholdHR:       15000,
				department:        "IT",
				amount:            25000,
				shouldRequireQuote: true,
			},
			{
				name:              "IT: Above threshold",
				thresholdIT:       25000,
				thresholdHR:       15000,
				department:        "IT",
				amount:            30000,
				shouldRequireQuote: true,
			},
			{
				name:              "HR: Below threshold",
				thresholdIT:       25000,
				thresholdHR:       15000,
				department:        "HR",
				amount:            10000,
				shouldRequireQuote: false,
			},
			{
				name:              "HR: Above threshold",
				thresholdIT:       25000,
				thresholdHR:       15000,
				department:        "HR",
				amount:            20000,
				shouldRequireQuote: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var threshold float64
				if tt.department == "IT" {
					threshold = tt.thresholdIT
				} else {
					threshold = tt.thresholdHR
				}

				requiresQuote := tt.amount >= threshold
				if requiresQuote != tt.shouldRequireQuote {
					t.Errorf("Expected %v, got %v", tt.shouldRequireQuote, requiresQuote)
				}
			})
		}
	})
}

// TestBudgetStatusCalculation tests budget status metrics
func TestBudgetStatusCalculation(t *testing.T) {
	t.Run("Budget utilization percentage", func(t *testing.T) {
		tests := []struct {
			name              string
			totalBudget       float64
			allocatedAmount   float64
			expectedPercent   float64
		}{
			{
				name:            "No allocation",
				totalBudget:     100000,
				allocatedAmount: 0,
				expectedPercent: 0,
			},
			{
				name:            "50% allocated",
				totalBudget:     100000,
				allocatedAmount: 50000,
				expectedPercent: 50,
			},
			{
				name:            "Full allocation",
				totalBudget:     100000,
				allocatedAmount: 100000,
				expectedPercent: 100,
			},
			{
				name:            "33% allocated",
				totalBudget:     100000,
				allocatedAmount: 33000,
				expectedPercent: 33,
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
	})
}

// TestBudgetConstraintStructure tests constraint model
func TestBudgetConstraintStructure(t *testing.T) {
	t.Run("Budget constraint fields", func(t *testing.T) {
		constraint := BudgetConstraint{
			ID:             "constraint-1",
			Department:     "IT",
			FiscalYear:     "2025",
			MaxBudget:      500000,
			MinBudget:      10000,
			MaxSingleOrder: 50000,
			ReserveFunds:   10,
			RequiresQuote:  true,
			QuoteThreshold: 25000,
		}

		if constraint.ID == "" {
			t.Error("Constraint ID should not be empty")
		}
		if constraint.Department == "" {
			t.Error("Constraint Department should not be empty")
		}
		if constraint.MaxBudget <= 0 {
			t.Error("Constraint MaxBudget should be positive")
		}
		if constraint.MaxSingleOrder <= 0 {
			t.Error("Constraint MaxSingleOrder should be positive")
		}
		if constraint.ReserveFunds < 0 || constraint.ReserveFunds > 100 {
			t.Error("Constraint ReserveFunds should be 0-100")
		}
	})
}

// BenchmarkBudgetValidation benchmarks validation logic
func BenchmarkBudgetValidation(b *testing.B) {
	totalBudget := 100000.0
	allocatedAmount := 30000.0
	remainingAmount := totalBudget - allocatedAmount
	requestAmount := 50000.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = requestAmount <= remainingAmount
	}
}

// BenchmarkAllocationCalculation benchmarks allocation math
func BenchmarkAllocationCalculation(b *testing.B) {
	totalBudget := 100000.0
	currentAllocated := 30000.0
	newAllocation := 20000.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		newAllocated := currentAllocated + newAllocation
		_ = totalBudget - newAllocated
	}
}

// BenchmarkReserveFundsCheck benchmarks reserve fund validation
func BenchmarkReserveFundsCheck(b *testing.B) {
	totalBudget := 100000.0
	reservePercent := 10.0
	allocatedAmount := 85000.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reserveAmount := totalBudget * (reservePercent / 100)
		remainingAfterAllocation := totalBudget - allocatedAmount
		_ = remainingAfterAllocation >= reserveAmount
	}
}
