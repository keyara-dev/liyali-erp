package services

import (
	"fmt"
	"log"

	"github.com/liyali/liyali-gateway/models"
	"gorm.io/gorm"
)

// BudgetConstraint represents budget rules and limits
type BudgetConstraint struct {
	ID              string  `gorm:"primaryKey" json:"id"`
	Department      string  `json:"department"`
	FiscalYear      string  `json:"fiscalYear"`
	MaxBudget       float64 `json:"maxBudget"`
	MinBudget       float64 `json:"minBudget"`
	MaxSingleOrder  float64 `json:"maxSingleOrder"`
	ReserveFunds    float64 `json:"reserveFunds"` // Percentage
	RequiresQuote   bool    `json:"requiresQuote"`
	QuoteThreshold  float64 `json:"quoteThreshold"`
	CreatedAt       string  `json:"createdAt"`
}

// BudgetValidationService handles budget constraint checking
type BudgetValidationService struct {
	db *gorm.DB
}

// NewBudgetValidationService creates a new budget validation service
func NewBudgetValidationService(db *gorm.DB) *BudgetValidationService {
	return &BudgetValidationService{db: db}
}

// ValidateBudgetForRequisition checks if a requisition amount is within budget
func (bvs *BudgetValidationService) ValidateBudgetForRequisition(
	department, fiscalYear string,
	amount float64,
) (bool, string, error) {
	// Get department budget
	var budget models.Budget
	if err := bvs.db.Where(
		"department = ? AND fiscal_year = ? AND status = ?",
		department, fiscalYear, "approved",
	).First(&budget).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, fmt.Sprintf("No approved budget found for %s in %s", department, fiscalYear), nil
		}
		return false, "", err
	}

	// Check if amount exceeds remaining budget
	if amount > budget.RemainingAmount {
		return false, fmt.Sprintf(
			"Amount %.2f exceeds remaining budget %.2f",
			amount, budget.RemainingAmount,
		), nil
	}

	// Get budget constraint
	constraint, err := bvs.getBudgetConstraint(department, fiscalYear)
	if err != nil {
		log.Printf("Warning: Could not get budget constraint: %v", err)
		return true, "", nil // Proceed if no constraint found
	}

	// Check max single order
	if amount > constraint.MaxSingleOrder {
		return false, fmt.Sprintf(
			"Amount %.2f exceeds max single order limit %.2f",
			amount, constraint.MaxSingleOrder,
		), nil
	}

	// Check if quotes are required
	if constraint.RequiresQuote && amount >= constraint.QuoteThreshold {
		return true, "Quotes required for amounts >= threshold", nil
	}

	return true, "", nil
}

// ValidateBudgetForPurchaseOrder checks if a PO amount is within budget and constraints
func (bvs *BudgetValidationService) ValidateBudgetForPurchaseOrder(
	department, fiscalYear string,
	amount float64,
	vendorID string,
) (bool, string, error) {
	// First validate basic budget constraint
	valid, msg, err := bvs.ValidateBudgetForRequisition(department, fiscalYear, amount)
	if err != nil || !valid {
		return valid, msg, err
	}

	// Check vendor-specific constraints (single vendor max per period)
	vendorTotal, err := bvs.getVendorPOTotal(vendorID, department, fiscalYear)
	if err != nil {
		log.Printf("Warning: Could not get vendor PO total: %v", err)
		return true, "", nil
	}

	// If vendor exceeds 30% of budget, flag for approval
	var budget models.Budget
	bvs.db.Where(
		"department = ? AND fiscal_year = ? AND status = ?",
		department, fiscalYear, "approved",
	).First(&budget)

	maxPerVendor := budget.TotalBudget * 0.3
	if vendorTotal+amount > maxPerVendor {
		return true, fmt.Sprintf(
			"Vendor total will exceed 30%% of budget (current: %.2f, new: %.2f, limit: %.2f)",
			vendorTotal, amount, maxPerVendor,
		), nil // Allow but flag
	}

	return true, "", nil
}

// ValidateBudgetAllocation checks if allocated amount doesn't exceed total budget
func (bvs *BudgetValidationService) ValidateBudgetAllocation(
	budget *models.Budget,
	additionalAllocation float64,
) (bool, string, error) {
	newTotal := budget.AllocatedAmount + additionalAllocation
	if newTotal > budget.TotalBudget {
		return false, fmt.Sprintf(
			"Allocation %.2f would exceed total budget %.2f (current allocation: %.2f)",
			additionalAllocation, budget.TotalBudget, budget.AllocatedAmount,
		), nil
	}

	// Check reserve funds requirement
	constraint, err := bvs.getBudgetConstraint(budget.Department, budget.FiscalYear)
	if err == nil {
		reserveAmount := budget.TotalBudget * (constraint.ReserveFunds / 100)
		remainingAfterAllocation := budget.TotalBudget - newTotal

		if remainingAfterAllocation < reserveAmount {
			return false, fmt.Sprintf(
				"Allocation would violate reserve fund requirement of %.2f",
				reserveAmount,
			), nil
		}
	}

	return true, "", nil
}

// AllocateBudget allocates funds from a budget to a requisition
func (bvs *BudgetValidationService) AllocateBudget(
	budgetID string,
	allocationAmount float64,
	requisitionID string,
) error {
	// Get budget
	var budget models.Budget
	if err := bvs.db.First(&budget, "id = ?", budgetID).Error; err != nil {
		return err
	}

	// Validate allocation
	valid, msg, err := bvs.ValidateBudgetAllocation(&budget, allocationAmount)
	if err != nil {
		return err
	}
	if !valid {
		return fmt.Errorf("allocation not allowed: %s", msg)
	}

	// Update budget allocation
	newAllocated := budget.AllocatedAmount + allocationAmount
	newRemaining := budget.TotalBudget - newAllocated

	if err := bvs.db.Model(&budget).
		Updates(map[string]interface{}{
			"allocated_amount": newAllocated,
			"remaining_amount": newRemaining,
		}).Error; err != nil {
		return err
	}

	// Log the allocation
	log.Printf("Allocated %.2f to budget %s for requisition %s", allocationAmount, budgetID, requisitionID)

	return nil
}

// DeallocateBudget releases allocated funds (e.g., when requisition is rejected)
func (bvs *BudgetValidationService) DeallocateBudget(
	budgetID string,
	allocationAmount float64,
	requisitionID string,
) error {
	// Get budget
	var budget models.Budget
	if err := bvs.db.First(&budget, "id = ?", budgetID).Error; err != nil {
		return err
	}

	// Update budget allocation (decrease)
	newAllocated := budget.AllocatedAmount - allocationAmount
	if newAllocated < 0 {
		newAllocated = 0
	}
	newRemaining := budget.TotalBudget - newAllocated

	if err := bvs.db.Model(&budget).
		Updates(map[string]interface{}{
			"allocated_amount": newAllocated,
			"remaining_amount": newRemaining,
		}).Error; err != nil {
		return err
	}

	log.Printf("Deallocated %.2f from budget %s for requisition %s", allocationAmount, budgetID, requisitionID)

	return nil
}

// GetBudgetStatus returns detailed budget status
func (bvs *BudgetValidationService) GetBudgetStatus(budgetID string) (map[string]interface{}, error) {
	var budget models.Budget
	if err := bvs.db.First(&budget, "id = ?", budgetID).Error; err != nil {
		return nil, err
	}

	utilizationPercent := 0.0
	if budget.TotalBudget > 0 {
		utilizationPercent = (budget.AllocatedAmount / budget.TotalBudget) * 100
	}

	return map[string]interface{}{
		"budgetId":              budget.ID,
		"department":            budget.Department,
		"fiscalYear":            budget.FiscalYear,
		"totalBudget":           budget.TotalBudget,
		"allocatedAmount":       budget.AllocatedAmount,
		"remainingAmount":       budget.RemainingAmount,
		"utilizationPercent":    utilizationPercent,
		"status":                budget.Status,
		"canAllocateMore":       budget.RemainingAmount > 0,
	}, nil
}

// GetBudgetsByDepartment returns all budgets for a department
func (bvs *BudgetValidationService) GetBudgetsByDepartment(
	department string,
) ([]models.Budget, error) {
	var budgets []models.Budget
	if err := bvs.db.Where(
		"department = ? AND status = ?",
		department, "approved",
	).Order("fiscal_year DESC").Find(&budgets).Error; err != nil {
		return nil, err
	}
	return budgets, nil
}

// getBudgetConstraint retrieves constraint rules for a department
func (bvs *BudgetValidationService) getBudgetConstraint(
	department, fiscalYear string,
) (*BudgetConstraint, error) {
	var constraint BudgetConstraint
	if err := bvs.db.Where(
		"department = ? AND fiscal_year = ?",
		department, fiscalYear,
	).First(&constraint).Error; err != nil {
		return nil, err
	}
	return &constraint, nil
}

// getVendorPOTotal calculates total PO amount for a vendor in a period
func (bvs *BudgetValidationService) getVendorPOTotal(
	vendorID, department, fiscalYear string,
) (float64, error) {
	var total float64
	if err := bvs.db.Model(&models.PurchaseOrder{}).
		Where("vendor_id = ?", vendorID).
		Where("status IN ?", []string{"approved", "fulfilled", "completed"}).
		Select("SUM(total_amount)").
		Row().
		Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

// CreateDefaultBudgetConstraints creates default constraints
func (bvs *BudgetValidationService) CreateDefaultBudgetConstraints() error {
	constraints := []BudgetConstraint{
		{
			ID:             "constraint-it-2025",
			Department:     "IT",
			FiscalYear:     "2025",
			MaxBudget:      500000,
			MinBudget:      10000,
			MaxSingleOrder: 50000,
			ReserveFunds:   10,
			RequiresQuote:  true,
			QuoteThreshold: 25000,
		},
		{
			ID:             "constraint-hr-2025",
			Department:     "HR",
			FiscalYear:     "2025",
			MaxBudget:      300000,
			MinBudget:      5000,
			MaxSingleOrder: 30000,
			ReserveFunds:   15,
			RequiresQuote:  true,
			QuoteThreshold: 15000,
		},
		{
			ID:             "constraint-ops-2025",
			Department:     "Operations",
			FiscalYear:     "2025",
			MaxBudget:      750000,
			MinBudget:      20000,
			MaxSingleOrder: 100000,
			ReserveFunds:   10,
			RequiresQuote:  true,
			QuoteThreshold: 50000,
		},
	}

	for _, constraint := range constraints {
		var count int64
		if err := bvs.db.Model(&BudgetConstraint{}).
			Where("id = ?", constraint.ID).
			Count(&count).Error; err != nil {
			return err
		}

		if count == 0 {
			if err := bvs.db.Create(&constraint).Error; err != nil {
				log.Printf("Error creating budget constraint: %v", err)
				return err
			}
		}
	}

	return nil
}
