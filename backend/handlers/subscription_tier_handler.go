package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/cache"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/middleware"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/utils"
	"gorm.io/datatypes"
)

// ============================================================================
// ADMIN TIER MANAGEMENT HANDLERS
// ============================================================================

// GetAllTiers returns all subscription tiers
func GetAllTiers(c *fiber.Ctx) error {
	db := config.DB

	var tiers []models.SubscriptionTier
	if err := db.Order("sort_order ASC").Find(&tiers).Error; err != nil {
		log.Printf("Error getting tiers: %v", err)
		return utils.SendInternalError(c, "Failed to retrieve tiers", err)
	}

	// Build response with computed fields
	responses := make([]models.TierResponse, len(tiers))
	for i, tier := range tiers {
		features, _ := tier.GetFeatureList()
		
		// Count organizations using this tier
		var orgCount int64
		db.Table("organizations").Where("subscription_tier = ?", tier.Name).Count(&orgCount)

		responses[i] = models.TierResponse{
			ID:                tier.ID,
			Name:              tier.Name,
			DisplayName:       tier.DisplayName,
			Description:       tier.Description,
			PriceMonthly:      tier.PriceMonthly,
			PriceYearly:       tier.PriceYearly,
			MaxWorkspaces:     tier.MaxWorkspaces,
			MaxTeamMembers:    tier.MaxTeamMembers,
			MaxDocuments:      tier.MaxDocuments,
			MaxWorkflows:      tier.MaxWorkflows,
			MaxCustomRoles:    tier.MaxCustomRoles,
			Features:          features,
			IsActive:          tier.IsActive,
			SortOrder:         tier.SortOrder,
			CreatedAt:         tier.CreatedAt,
			UpdatedAt:         tier.UpdatedAt,
			FeatureCount:      len(features),
			OrganizationCount: int(orgCount),
		}
	}

	return utils.SendSimpleSuccess(c, responses, "Tiers retrieved successfully")
}

// GetTierByID returns a specific tier
func GetTierByID(c *fiber.Ctx) error {
	db := config.DB
	tierID := c.Params("id")

	var tier models.SubscriptionTier
	if err := db.First(&tier, "id = ?", tierID).Error; err != nil {
		return utils.SendNotFound(c, "Tier not found")
	}

	features, _ := tier.GetFeatureList()
	var orgCount int64
	db.Table("organizations").Where("subscription_tier = ?", tier.Name).Count(&orgCount)

	response := models.TierResponse{
		ID:                tier.ID,
		Name:              tier.Name,
		DisplayName:       tier.DisplayName,
		Description:       tier.Description,
		PriceMonthly:      tier.PriceMonthly,
		PriceYearly:       tier.PriceYearly,
		MaxWorkspaces:     tier.MaxWorkspaces,
		MaxTeamMembers:    tier.MaxTeamMembers,
		MaxDocuments:      tier.MaxDocuments,
		MaxWorkflows:      tier.MaxWorkflows,
		MaxCustomRoles:    tier.MaxCustomRoles,
		Features:          features,
		IsActive:          tier.IsActive,
		SortOrder:         tier.SortOrder,
		CreatedAt:         tier.CreatedAt,
		UpdatedAt:         tier.UpdatedAt,
		FeatureCount:      len(features),
		OrganizationCount: int(orgCount),
	}

	return utils.SendSimpleSuccess(c, response, "Tier retrieved successfully")
}

// CreateTier creates a new subscription tier
func CreateTier(c *fiber.Ctx) error {
	db := config.DB

	var request models.CreateTierRequest
	if err := c.BodyParser(&request); err != nil {
		return utils.SendBadRequest(c, "Invalid request body")
	}

	// Convert features to JSON
	featuresJSON, err := json.Marshal(request.Features)
	if err != nil {
		return utils.SendBadRequest(c, "Invalid features format")
	}

	tier := models.SubscriptionTier{
		ID:             "tier-" + uuid.New().String()[:8],
		Name:           request.Name,
		DisplayName:    request.DisplayName,
		Description:    request.Description,
		PriceMonthly:   request.PriceMonthly,
		PriceYearly:    request.PriceYearly,
		MaxWorkspaces:  request.MaxWorkspaces,
		MaxTeamMembers: request.MaxTeamMembers,
		MaxDocuments:   request.MaxDocuments,
		MaxWorkflows:   request.MaxWorkflows,
		MaxCustomRoles: request.MaxCustomRoles,
		Features:       datatypes.JSON(featuresJSON),
		IsActive:       request.IsActive,
		SortOrder:      request.SortOrder,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := db.Create(&tier).Error; err != nil {
		log.Printf("Error creating tier: %v", err)
		return utils.SendInternalError(c, "Failed to create tier", err)
	}

	// Clear all tier caches
	cache.GetCache().ClearPrefix("feature:")
	cache.GetCache().ClearPrefix("limits:")

	return utils.SendSimpleSuccess(c, tier, "Tier created successfully")
}

// UpdateTier updates an existing tier
func UpdateTier(c *fiber.Ctx) error {
	db := config.DB
	tierID := c.Params("id")

	var tier models.SubscriptionTier
	if err := db.First(&tier, "id = ?", tierID).Error; err != nil {
		return utils.SendNotFound(c, "Tier not found")
	}

	var request models.UpdateTierRequest
	if err := c.BodyParser(&request); err != nil {
		return utils.SendBadRequest(c, "Invalid request body")
	}

	// Update fields if provided
	if request.DisplayName != nil {
		tier.DisplayName = *request.DisplayName
	}
	if request.Description != nil {
		tier.Description = *request.Description
	}
	if request.PriceMonthly != nil {
		tier.PriceMonthly = *request.PriceMonthly
	}
	if request.PriceYearly != nil {
		tier.PriceYearly = *request.PriceYearly
	}
	if request.MaxWorkspaces != nil {
		tier.MaxWorkspaces = *request.MaxWorkspaces
	}
	if request.MaxTeamMembers != nil {
		tier.MaxTeamMembers = *request.MaxTeamMembers
	}
	if request.MaxDocuments != nil {
		tier.MaxDocuments = *request.MaxDocuments
	}
	if request.MaxWorkflows != nil {
		tier.MaxWorkflows = *request.MaxWorkflows
	}
	if request.MaxCustomRoles != nil {
		tier.MaxCustomRoles = *request.MaxCustomRoles
	}
	if request.Features != nil {
		featuresJSON, err := json.Marshal(*request.Features)
		if err != nil {
			return utils.SendBadRequest(c, "Invalid features format")
		}
		tier.Features = datatypes.JSON(featuresJSON)
	}
	if request.IsActive != nil {
		tier.IsActive = *request.IsActive
	}
	if request.SortOrder != nil {
		tier.SortOrder = *request.SortOrder
	}

	tier.UpdatedAt = time.Now()

	if err := db.Save(&tier).Error; err != nil {
		log.Printf("Error updating tier: %v", err)
		return utils.SendInternalError(c, "Failed to update tier", err)
	}

	// Clear all tier caches
	cache.GetCache().ClearPrefix("feature:")
	cache.GetCache().ClearPrefix("limits:")

	return utils.SendSimpleSuccess(c, tier, "Tier updated successfully")
}

// ============================================================================
// ORGANIZATION TIER MANAGEMENT HANDLERS
// ============================================================================

// ChangeOrganizationTier changes an organization's subscription tier
func ChangeOrganizationTier(c *fiber.Ctx) error {
	db := config.DB
	orgID := c.Params("id")

	var request models.ChangeTierRequest
	if err := c.BodyParser(&request); err != nil {
		return utils.SendBadRequest(c, "Invalid request body")
	}

	// Verify tier exists and is active
	var tier models.SubscriptionTier
	if err := db.First(&tier, "name = ? AND is_active = ?", request.NewTier, true).Error; err != nil {
		return utils.SendBadRequest(c, "Invalid or inactive tier")
	}

	// Get current organization
	var org models.Organization
	if err := db.First(&org, "id = ?", orgID).Error; err != nil {
		return utils.SendNotFound(c, "Organization not found")
	}

	oldTier := org.Tier

	// Update tier
	if err := db.Model(&org).Updates(map[string]interface{}{
		"subscription_tier": request.NewTier,
		"updated_at":        time.Now(),
	}).Error; err != nil {
		log.Printf("Error updating organization tier: %v", err)
		return utils.SendInternalError(c, "Failed to update tier", err)
	}

	// Create audit log
	adminUserID := c.Locals("userID")
	if adminUserID == nil {
		adminUserID = "system"
	}

	auditLog := map[string]interface{}{
		"id":              "audit-" + uuid.New().String(),
		"organization_id": orgID,
		"action":          "tier_change",
		"old_value":       oldTier,
		"new_value":       request.NewTier,
		"reason":          request.Reason,
		"admin_user_id":   adminUserID,
		"created_at":      time.Now(),
	}
	db.Table("admin_audit_logs").Create(auditLog)

	// Clear caches for this organization
	cache.GetCache().ClearPrefix(fmt.Sprintf("feature:%s:", orgID))
	cache.GetCache().ClearPrefix(fmt.Sprintf("limits:%s", orgID))

	return utils.SendSimpleSuccess(c, fiber.Map{
		"organization_id": orgID,
		"old_tier":        oldTier,
		"new_tier":        request.NewTier,
	}, "Tier changed successfully")
}

// OverrideOrganizationLimits creates or updates limit overrides for an organization
func OverrideOrganizationLimits(c *fiber.Ctx) error {
	db := config.DB
	orgID := c.Params("id")

	var request models.OverrideLimitsRequest
	if err := c.BodyParser(&request); err != nil {
		return utils.SendBadRequest(c, "Invalid request body")
	}

	// Check if override already exists
	var override models.OrganizationLimitOverride
	err := db.Where("organization_id = ?", orgID).First(&override).Error

	adminUserID := c.Locals("userID")
	if adminUserID == nil {
		adminUserID = "system"
	}

	if err != nil {
		// Create new override
		override = models.OrganizationLimitOverride{
			ID:             "override-" + uuid.New().String()[:8],
			OrganizationID: orgID,
			Reason:         request.Reason,
			AdminUserID:    fmt.Sprintf("%v", adminUserID),
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}
	} else {
		// Update existing override
		override.Reason = request.Reason
		override.AdminUserID = fmt.Sprintf("%v", adminUserID)
		override.UpdatedAt = time.Now()
	}

	// Set override values
	override.MaxWorkspaces = request.MaxWorkspaces
	override.MaxTeamMembers = request.MaxTeamMembers
	override.MaxDocuments = request.MaxDocuments
	override.MaxWorkflows = request.MaxWorkflows
	override.MaxCustomRoles = request.MaxCustomRoles

	if request.Features != nil {
		featuresJSON, err := json.Marshal(*request.Features)
		if err != nil {
			return utils.SendBadRequest(c, "Invalid features format")
		}
		override.Features = datatypes.JSON(featuresJSON)
	}

	if request.ExpiresAt != nil {
		expiresAt, err := time.Parse(time.RFC3339, *request.ExpiresAt)
		if err != nil {
			return utils.SendBadRequest(c, "Invalid expiration date format")
		}
		override.ExpiresAt = &expiresAt
	}

	if err := db.Save(&override).Error; err != nil {
		log.Printf("Error saving override: %v", err)
		return utils.SendInternalError(c, "Failed to save override", err)
	}

	// Clear caches
	cache.GetCache().ClearPrefix(fmt.Sprintf("feature:%s:", orgID))
	cache.GetCache().ClearPrefix(fmt.Sprintf("limits:%s", orgID))

	return utils.SendSimpleSuccess(c, override, "Limits overridden successfully")
}

// ============================================================================
// PUBLIC ORGANIZATION SUBSCRIPTION HANDLERS
// ============================================================================

// GetOrganizationSubscription returns subscription info for an organization
func GetOrganizationSubscription(c *fiber.Ctx) error {
	db := config.DB
	orgID := c.Params("id")

	var org models.Organization
	if err := db.First(&org, "id = ?", orgID).Error; err != nil {
		return utils.SendNotFound(c, "Organization not found")
	}

	var tier models.SubscriptionTier
	if err := db.First(&tier, "name = ?", org.Tier).Error; err != nil {
		return utils.SendNotFound(c, "Subscription tier not found")
	}

	return utils.SendSimpleSuccess(c, fiber.Map{
		"organization_id": orgID,
		"tier":            tier,
	}, "Subscription retrieved successfully")
}

// GetOrganizationFeatures returns available features for an organization
func GetOrganizationFeatures(c *fiber.Ctx) error {
	orgID := c.Params("id")

	features, err := middleware.GetOrganizationFeatures(orgID)
	if err != nil {
		log.Printf("Error getting features for org %s: %v", orgID, err)
		return utils.SendInternalError(c, "Failed to get features", err)
	}

	return utils.SendSimpleSuccess(c, fiber.Map{
		"organization_id": orgID,
		"features":        features,
	}, "Features retrieved successfully")
}

// GetOrganizationLimits returns effective limits for an organization
func GetOrganizationLimits(c *fiber.Ctx) error {
	orgID := c.Params("id")

	limits, err := middleware.GetEffectiveLimits(orgID)
	if err != nil {
		log.Printf("Error getting limits for org %s: %v", orgID, err)
		return utils.SendInternalError(c, "Failed to get limits", err)
	}

	return utils.SendSimpleSuccess(c, limits, "Limits retrieved successfully")
}

// GetOrganizationUsage returns current usage and limits for an organization.
// Supports both admin (orgID from URL param) and tenant (orgID from context) usage.
func GetOrganizationUsage(c *fiber.Ctx) error {
	// Try URL param first (admin routes), then tenant context
	orgID := c.Params("id")
	if orgID == "" {
		if id, ok := c.Locals("organizationID").(string); ok {
			orgID = id
		}
	}
	if orgID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ErrorResponse("Organization ID required"))
	}

	limits, err := middleware.GetEffectiveLimits(orgID)
	if err != nil {
		log.Printf("Error getting limits for org %s: %v", orgID, err)
		return utils.SendInternalError(c, "Failed to get limits", err)
	}

	usage, err := middleware.GetOrganizationUsage(orgID)
	if err != nil {
		log.Printf("Error getting usage for org %s: %v", orgID, err)
		return utils.SendInternalError(c, "Failed to get usage", err)
	}

	result := models.LimitsWithUsage{
		Limits: *limits,
		Usage:  *usage,
	}

	return utils.SendSimpleSuccess(c, result, "Usage retrieved successfully")
}
