package middleware

import (
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/liyali/liyali-gateway/cache"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/models"
)

// CheckLimit validates if an organization can create more resources of a specific type
// Must be used after TenantMiddleware
func CheckLimit(resourceType string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get organization ID from context (set by TenantMiddleware)
		orgID, ok := c.Locals("organizationID").(string)
		if !ok || orgID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Organization context required",
			})
		}

		// Get effective limits (tier + overrides)
		limits, err := getEffectiveLimits(orgID)
		if err != nil {
			log.Printf("Error getting effective limits for org %s: %v", orgID, err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to check resource limits",
			})
		}

		// Get current usage
		usage, err := getCurrentUsage(orgID, resourceType)
		if err != nil {
			log.Printf("Error getting current usage for org %s, resource %s: %v", orgID, resourceType, err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to check current usage",
			})
		}

		// Get limit for this resource type
		var limit int
		switch resourceType {
		case "workspace":
			limit = limits.MaxWorkspaces
		case "team_member":
			limit = limits.MaxTeamMembers
		case "document":
			limit = limits.MaxDocuments
		case "workflow":
			limit = limits.MaxWorkflows
		case "custom_role":
			limit = limits.MaxCustomRoles
		case "requisition":
			limit = limits.MaxRequisitions
		case "budget":
			limit = limits.MaxBudgets
		case "purchase_order":
			limit = limits.MaxPurchaseOrders
		case "payment_voucher":
			limit = limits.MaxPaymentVouchers
		case "grn":
			limit = limits.MaxGRNs
		case "department":
			limit = limits.MaxDepartments
		case "vendor":
			limit = limits.MaxVendors
		default:
			// Unknown resource type, allow by default
			return c.Next()
		}

		// -1 means unlimited
		if limit == models.UnlimitedLimit {
			return c.Next()
		}

		// Check if at limit
		if usage >= limit {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": fmt.Sprintf("You have reached your %s limit (%d/%d). Please upgrade your subscription to create more.", 
					resourceType, usage, limit),
				"resourceType": resourceType,
				"currentUsage": usage,
				"limit":        limit,
			})
		}

		return c.Next()
	}
}

// getEffectiveLimits returns the effective limits for an organization (tier + overrides)
// Uses cache with 5-minute TTL for performance
func getEffectiveLimits(orgID string) (*models.EffectiveLimits, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("limits:%s", orgID)
	appCache := cache.GetCache()

	if cached, found := appCache.Get(cacheKey); found {
		if limits, ok := cached.(*models.EffectiveLimits); ok {
			return limits, nil
		}
	}

	// Cache miss - query database
	limits, err := queryEffectiveLimits(orgID)
	if err != nil {
		return nil, err
	}

	// Cache result for 5 minutes
	appCache.Set(cacheKey, limits, 5*time.Minute)

	return limits, nil
}

// queryEffectiveLimits queries the database for effective limits
func queryEffectiveLimits(orgID string) (*models.EffectiveLimits, error) {
	db := config.DB

	// Get organization's subscription tier
	var org struct {
		SubscriptionTier string
	}
	if err := db.Table("organizations").
		Select("subscription_tier").
		Where("id = ?", orgID).
		First(&org).Error; err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}

	// Get tier limits
	var tier models.SubscriptionTier
	if err := db.Where("name = ? AND is_active = ?", org.SubscriptionTier, true).
		First(&tier).Error; err != nil {
		return nil, fmt.Errorf("subscription tier not found: %w", err)
	}

	// Start with tier limits
	limits := &models.EffectiveLimits{
		OrganizationID:     orgID,
		TierName:           tier.Name,
		MaxWorkspaces:      tier.MaxWorkspaces,
		MaxTeamMembers:     tier.MaxTeamMembers,
		MaxDocuments:       tier.MaxDocuments,
		MaxWorkflows:       tier.MaxWorkflows,
		MaxCustomRoles:     tier.MaxCustomRoles,
		MaxRequisitions:    tier.MaxRequisitions,
		MaxBudgets:         tier.MaxBudgets,
		MaxPurchaseOrders:  tier.MaxPurchaseOrders,
		MaxPaymentVouchers: tier.MaxPaymentVouchers,
		MaxGRNs:            tier.MaxGRNs,
		MaxDepartments:     tier.MaxDepartments,
		MaxVendors:         tier.MaxVendors,
		HasOverrides:       false,
	}

	// Check for overrides
	var override models.OrganizationLimitOverride
	err := db.Where("organization_id = ? AND (expires_at IS NULL OR expires_at > NOW())", orgID).
		First(&override).Error

	if err == nil {
		// Apply overrides (NULL means use tier default)
		limits.HasOverrides = true

		if override.MaxWorkspaces != nil {
			limits.MaxWorkspaces = *override.MaxWorkspaces
		}
		if override.MaxTeamMembers != nil {
			limits.MaxTeamMembers = *override.MaxTeamMembers
		}
		if override.MaxDocuments != nil {
			limits.MaxDocuments = *override.MaxDocuments
		}
		if override.MaxWorkflows != nil {
			limits.MaxWorkflows = *override.MaxWorkflows
		}
		if override.MaxCustomRoles != nil {
			limits.MaxCustomRoles = *override.MaxCustomRoles
		}
		if override.MaxRequisitions != nil {
			limits.MaxRequisitions = *override.MaxRequisitions
		}
		if override.MaxBudgets != nil {
			limits.MaxBudgets = *override.MaxBudgets
		}
		if override.MaxPurchaseOrders != nil {
			limits.MaxPurchaseOrders = *override.MaxPurchaseOrders
		}
		if override.MaxPaymentVouchers != nil {
			limits.MaxPaymentVouchers = *override.MaxPaymentVouchers
		}
		if override.MaxGRNs != nil {
			limits.MaxGRNs = *override.MaxGRNs
		}
		if override.MaxDepartments != nil {
			limits.MaxDepartments = *override.MaxDepartments
		}
		if override.MaxVendors != nil {
			limits.MaxVendors = *override.MaxVendors
		}
	}

	return limits, nil
}

// getCurrentUsage returns the current usage count for a resource type
func getCurrentUsage(orgID, resourceType string) (int, error) {
	db := config.DB
	var count int64

	switch resourceType {
	case "workspace":
		// For now, workspaces = organizations, so always 1
		// Future: implement multi-workspace support
		count = 1

	case "team_member":
		err := db.Table("organization_members").
			Where("organization_id = ? AND active = ?", orgID, true).
			Count(&count).Error
		if err != nil {
			return 0, fmt.Errorf("failed to count team members: %w", err)
		}

	case "document":
		// Count all active documents (not deleted)
		err := db.Table("documents").
			Where("organization_id = ? AND deleted_at IS NULL", orgID).
			Count(&count).Error
		if err != nil {
			return 0, fmt.Errorf("failed to count documents: %w", err)
		}

	case "workflow":
		err := db.Table("workflows").
			Where("organization_id = ? AND is_active = ?", orgID, true).
			Count(&count).Error
		if err != nil {
			return 0, fmt.Errorf("failed to count workflows: %w", err)
		}

	case "custom_role":
		err := db.Table("organization_roles").
			Where("organization_id = ? AND is_system_role = ? AND active = ?", orgID, false, true).
			Count(&count).Error
		if err != nil {
			return 0, fmt.Errorf("failed to count custom roles: %w", err)
		}

	case "requisition":
		err := db.Table("requisitions").
			Where("organization_id = ? AND deleted_at IS NULL", orgID).
			Count(&count).Error
		if err != nil {
			return 0, fmt.Errorf("failed to count requisitions: %w", err)
		}

	case "budget":
		err := db.Table("budgets").
			Where("organization_id = ? AND deleted_at IS NULL", orgID).
			Count(&count).Error
		if err != nil {
			return 0, fmt.Errorf("failed to count budgets: %w", err)
		}

	case "purchase_order":
		err := db.Table("purchase_orders").
			Where("organization_id = ? AND deleted_at IS NULL", orgID).
			Count(&count).Error
		if err != nil {
			return 0, fmt.Errorf("failed to count purchase orders: %w", err)
		}

	case "payment_voucher":
		err := db.Table("payment_vouchers").
			Where("organization_id = ? AND deleted_at IS NULL", orgID).
			Count(&count).Error
		if err != nil {
			return 0, fmt.Errorf("failed to count payment vouchers: %w", err)
		}

	case "grn":
		err := db.Table("goods_received_notes").
			Where("organization_id = ? AND deleted_at IS NULL", orgID).
			Count(&count).Error
		if err != nil {
			return 0, fmt.Errorf("failed to count GRNs: %w", err)
		}

	case "department":
		err := db.Table("organization_departments").
			Where("organization_id = ? AND is_active = ?", orgID, true).
			Count(&count).Error
		if err != nil {
			return 0, fmt.Errorf("failed to count departments: %w", err)
		}

	case "vendor":
		err := db.Table("vendors").
			Where("organization_id = ? AND deleted_at IS NULL", orgID).
			Count(&count).Error
		if err != nil {
			return 0, fmt.Errorf("failed to count vendors: %w", err)
		}

	default:
		return 0, fmt.Errorf("unknown resource type: %s", resourceType)
	}

	return int(count), nil
}

// GetEffectiveLimits is a helper function to get effective limits without middleware
// Useful for displaying limits in UI
func GetEffectiveLimits(orgID string) (*models.EffectiveLimits, error) {
	return getEffectiveLimits(orgID)
}

// GetCurrentUsage is a helper function to get current usage without middleware
// Useful for displaying usage in UI
func GetCurrentUsage(orgID, resourceType string) (int, error) {
	return getCurrentUsage(orgID, resourceType)
}

// GetOrganizationUsage returns usage for all resource types
func GetOrganizationUsage(orgID string) (*models.OrganizationUsage, error) {
	limits, err := getEffectiveLimits(orgID)
	if err != nil {
		return nil, err
	}

	usage := &models.OrganizationUsage{
		OrganizationID: orgID,
	}

	// Get usage for each resource type
	if count, err := getCurrentUsage(orgID, "workspace"); err == nil {
		usage.CurrentWorkspaces = count
		if limits.MaxWorkspaces > 0 {
			usage.WorkspacesPercent = float64(count) / float64(limits.MaxWorkspaces) * 100
		}
	}

	if count, err := getCurrentUsage(orgID, "team_member"); err == nil {
		usage.CurrentTeamMembers = count
		if limits.MaxTeamMembers > 0 {
			usage.TeamMembersPercent = float64(count) / float64(limits.MaxTeamMembers) * 100
		}
	}

	if count, err := getCurrentUsage(orgID, "document"); err == nil {
		usage.CurrentDocuments = count
		if limits.MaxDocuments > 0 {
			usage.DocumentsPercent = float64(count) / float64(limits.MaxDocuments) * 100
		}
	}

	if count, err := getCurrentUsage(orgID, "workflow"); err == nil {
		usage.CurrentWorkflows = count
		if limits.MaxWorkflows > 0 {
			usage.WorkflowsPercent = float64(count) / float64(limits.MaxWorkflows) * 100
		}
	}

	if count, err := getCurrentUsage(orgID, "custom_role"); err == nil {
		usage.CurrentCustomRoles = count
		if limits.MaxCustomRoles > 0 {
			usage.CustomRolesPercent = float64(count) / float64(limits.MaxCustomRoles) * 100
		}
	}

	if count, err := getCurrentUsage(orgID, "requisition"); err == nil {
		usage.CurrentRequisitions = count
		if limits.MaxRequisitions > 0 {
			usage.RequisitionsPercent = float64(count) / float64(limits.MaxRequisitions) * 100
		}
	}

	if count, err := getCurrentUsage(orgID, "budget"); err == nil {
		usage.CurrentBudgets = count
		if limits.MaxBudgets > 0 {
			usage.BudgetsPercent = float64(count) / float64(limits.MaxBudgets) * 100
		}
	}

	if count, err := getCurrentUsage(orgID, "purchase_order"); err == nil {
		usage.CurrentPurchaseOrders = count
		if limits.MaxPurchaseOrders > 0 {
			usage.PurchaseOrdersPercent = float64(count) / float64(limits.MaxPurchaseOrders) * 100
		}
	}

	if count, err := getCurrentUsage(orgID, "payment_voucher"); err == nil {
		usage.CurrentPaymentVouchers = count
		if limits.MaxPaymentVouchers > 0 {
			usage.PaymentVouchersPercent = float64(count) / float64(limits.MaxPaymentVouchers) * 100
		}
	}

	if count, err := getCurrentUsage(orgID, "grn"); err == nil {
		usage.CurrentGRNs = count
		if limits.MaxGRNs > 0 {
			usage.GRNsPercent = float64(count) / float64(limits.MaxGRNs) * 100
		}
	}

	if count, err := getCurrentUsage(orgID, "department"); err == nil {
		usage.CurrentDepartments = count
		if limits.MaxDepartments > 0 {
			usage.DepartmentsPercent = float64(count) / float64(limits.MaxDepartments) * 100
		}
	}

	if count, err := getCurrentUsage(orgID, "vendor"); err == nil {
		usage.CurrentVendors = count
		if limits.MaxVendors > 0 {
			usage.VendorsPercent = float64(count) / float64(limits.MaxVendors) * 100
		}
	}

	return usage, nil
}
