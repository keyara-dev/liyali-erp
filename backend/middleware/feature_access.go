package middleware

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/liyali/liyali-gateway/cache"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/models"
)

// RequireFeature checks if the organization has access to a specific feature
// Must be used after TenantMiddleware
func RequireFeature(featureName string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get organization ID from context (set by TenantMiddleware)
		orgID, ok := c.Locals("organizationID").(string)
		if !ok || orgID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Organization context required",
			})
		}

		// Check feature access
		hasAccess, err := checkFeatureAccess(orgID, featureName)
		if err != nil {
			log.Printf("Error checking feature access for org %s, feature %s: %v", orgID, featureName, err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to check feature access",
			})
		}

		if !hasAccess {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"message": fmt.Sprintf("Feature '%s' is not available in your subscription tier. Please upgrade to access this feature.", featureName),
				"feature": featureName,
			})
		}

		return c.Next()
	}
}

// checkFeatureAccess checks if an organization has access to a feature
// Uses cache with 5-minute TTL for performance
func checkFeatureAccess(orgID, featureName string) (bool, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("feature:%s:%s", orgID, featureName)
	appCache := cache.GetCache()

	if cached, found := appCache.Get(cacheKey); found {
		if hasAccess, ok := cached.(bool); ok {
			return hasAccess, nil
		}
	}

	// Cache miss - query database
	hasAccess, err := queryFeatureAccess(orgID, featureName)
	if err != nil {
		return false, err
	}

	// Cache result for 5 minutes
	appCache.Set(cacheKey, hasAccess, 5*time.Minute)

	return hasAccess, nil
}

// queryFeatureAccess queries the database for feature access
func queryFeatureAccess(orgID, featureName string) (bool, error) {
	db := config.DB

	// Get organization's subscription tier
	var org struct {
		SubscriptionTier string
	}
	if err := db.Table("organizations").
		Select("subscription_tier").
		Where("id = ?", orgID).
		First(&org).Error; err != nil {
		return false, fmt.Errorf("organization not found: %w", err)
	}

	// Get tier features
	var tier models.SubscriptionTier
	if err := db.Where("name = ? AND is_active = ?", org.SubscriptionTier, true).
		First(&tier).Error; err != nil {
		return false, fmt.Errorf("subscription tier not found: %w", err)
	}

	// Parse features JSON
	features, err := tier.GetFeatureList()
	if err != nil {
		return false, fmt.Errorf("failed to parse tier features: %w", err)
	}

	// Check if feature is in tier
	for _, f := range features {
		if f == featureName {
			return true, nil
		}
	}

	// Check for organization-specific overrides
	var override models.OrganizationLimitOverride
	err = db.Where("organization_id = ? AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)", orgID).
		First(&override).Error

	if err == nil && override.Features != nil {
		var overrideFeatures []string
		if err := json.Unmarshal(override.Features, &overrideFeatures); err == nil {
			for _, f := range overrideFeatures {
				if f == featureName {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

// CheckFeatureAccess is a helper function to check feature access without middleware
// Useful for conditional logic in handlers
func CheckFeatureAccess(orgID, featureName string) (bool, error) {
	return checkFeatureAccess(orgID, featureName)
}

// GetOrganizationFeatures returns all features available to an organization
func GetOrganizationFeatures(orgID string) ([]string, error) {
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

	// Get tier features
	var tier models.SubscriptionTier
	if err := db.Where("name = ? AND is_active = ?", org.SubscriptionTier, true).
		First(&tier).Error; err != nil {
		return nil, fmt.Errorf("subscription tier not found: %w", err)
	}

	// Parse features JSON
	features, err := tier.GetFeatureList()
	if err != nil {
		return nil, fmt.Errorf("failed to parse tier features: %w", err)
	}

	// Add override features if they exist
	var override models.OrganizationLimitOverride
	err = db.Where("organization_id = ? AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)", orgID).
		First(&override).Error

	if err == nil && override.Features != nil {
		var overrideFeatures []string
		if err := json.Unmarshal(override.Features, &overrideFeatures); err == nil {
			// Merge features (avoid duplicates)
			featureMap := make(map[string]bool)
			for _, f := range features {
				featureMap[f] = true
			}
			for _, f := range overrideFeatures {
				featureMap[f] = true
			}

			// Convert back to slice
			allFeatures := make([]string, 0, len(featureMap))
			for f := range featureMap {
				allFeatures = append(allFeatures, f)
			}
			return allFeatures, nil
		}
	}

	return features, nil
}
