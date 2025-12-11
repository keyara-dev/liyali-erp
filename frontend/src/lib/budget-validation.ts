import { Budget, BudgetItem } from '@/types/budget'

/**
 * Calculate total allocated amount across all budget items
 */
export function calculateTotalAllocated(items: BudgetItem[]): number {
  return items.reduce((sum, item) => sum + item.allocatedAmount, 0)
}

/**
 * Calculate total spent amount across all budget items
 */
export function calculateTotalSpent(items: BudgetItem[]): number {
  return items.reduce((sum, item) => sum + item.spentAmount, 0)
}

/**
 * Calculate remaining budget after allocations
 */
export function calculateRemainingBudget(totalAmount: number, items: BudgetItem[]): number {
  const totalAllocated = calculateTotalAllocated(items)
  return totalAmount - totalAllocated
}

/**
 * Calculate remaining budget for a specific item (allocated - spent)
 */
export function calculateItemRemaining(item: BudgetItem): number {
  return item.allocatedAmount - item.spentAmount
}

/**
 * Calculate percentage of budget allocated
 */
export function calculateAllocationPercentage(totalAmount: number, items: BudgetItem[]): number {
  if (totalAmount === 0) return 0
  const totalAllocated = calculateTotalAllocated(items)
  return (totalAllocated / totalAmount) * 100
}

/**
 * Calculate percentage of allocated budget that has been spent
 */
export function calculateSpentPercentage(item: BudgetItem): number {
  if (item.allocatedAmount === 0) return 0
  return (item.spentAmount / item.allocatedAmount) * 100
}

/**
 * Validate a budget item before adding/updating
 */
export interface ValidationResult {
  valid: boolean
  error?: string
}

export function validateBudgetItem(
  item: {
    allocatedAmount: number
    spentAmount?: number
  },
  currentItems: BudgetItem[],
  totalBudget: number,
  excludeItemId?: string
): ValidationResult {
  // Validate allocated amount is positive
  if (item.allocatedAmount <= 0) {
    return {
      valid: false,
      error: 'Allocated amount must be greater than 0',
    }
  }

  // Validate spent amount if provided
  if (item.spentAmount !== undefined) {
    if (item.spentAmount < 0) {
      return {
        valid: false,
        error: 'Spent amount cannot be negative',
      }
    }

    if (item.spentAmount > item.allocatedAmount) {
      return {
        valid: false,
        error: 'Spent amount cannot exceed allocated amount',
      }
    }
  }

  // Calculate total allocated including this item (excluding if editing)
  const otherItemsTotal = currentItems
    .filter((existing) => existing.id !== excludeItemId)
    .reduce((sum, existing) => sum + existing.allocatedAmount, 0)

  const newTotalAllocated = otherItemsTotal + item.allocatedAmount

  // Validate doesn't exceed total budget
  if (newTotalAllocated > totalBudget) {
    const excess = newTotalAllocated - totalBudget
    return {
      valid: false,
      error: `This would exceed your budget by ${excess.toLocaleString('en-US', {
        style: 'currency',
        currency: 'USD',
      })}`,
    }
  }

  return { valid: true }
}

/**
 * Validate budget has items and is properly allocated before submission
 */
export function validateBudgetForSubmission(budget: Budget): ValidationResult {
  // Check has items
  if (!budget.items || budget.items.length === 0) {
    return {
      valid: false,
      error: 'Budget must have at least one item before submission',
    }
  }

  // Check total allocated is greater than 0
  const totalAllocated = calculateTotalAllocated(budget.items)
  if (totalAllocated <= 0) {
    return {
      valid: false,
      error: 'Total allocated amount must be greater than 0',
    }
  }

  // Check not over-allocated
  if (totalAllocated > budget.totalAmount) {
    const excess = totalAllocated - budget.totalAmount
    return {
      valid: false,
      error: `Budget is over-allocated by ${excess.toLocaleString('en-US', {
        style: 'currency',
        currency: 'USD',
      })}`,
    }
  }

  return { valid: true }
}

/**
 * Check if budget is fully allocated
 */
export function isBudgetFullyAllocated(budget: Budget): boolean {
  const totalAllocated = calculateTotalAllocated(budget.items)
  return Math.abs(totalAllocated - budget.totalAmount) < 0.01 // Allow for floating point errors
}

/**
 * Check if budget is over-allocated
 */
export function isBudgetOverAllocated(budget: Budget): boolean {
  const totalAllocated = calculateTotalAllocated(budget.items)
  return totalAllocated > budget.totalAmount
}

/**
 * Check if budget is under-allocated
 */
export function isBudgetUnderAllocated(budget: Budget): boolean {
  const totalAllocated = calculateTotalAllocated(budget.items)
  return totalAllocated < budget.totalAmount
}

/**
 * Get allocation status for UI display
 */
export type AllocationStatus = 'under' | 'full' | 'over'

export function getAllocationStatus(budget: Budget): AllocationStatus {
  const totalAllocated = calculateTotalAllocated(budget.items)

  if (isBudgetOverAllocated(budget)) return 'over'
  if (isBudgetFullyAllocated(budget)) return 'full'
  return 'under'
}
