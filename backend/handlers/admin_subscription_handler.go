package handlers

import (
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/utils"
)

// Admin Subscription Management Handlers
// These handlers provide comprehensive subscription management for administrators

// GetAllSubscriptionTiers returns all subscription tiers for admin management
func GetAllSubscriptionTiers(c *fiber.Ctx) error {
	db := config.DB

	var tiers []models.SubscriptionTier
	if err := db.Order("sort_order ASC, created_at ASC").Find(&tiers).Error; err != nil {
		log.Printf("Error getting subscription tiers: %v", err)
		return utils.SendInternalError(c, "Failed to retrieve subscription tiers", err)
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

	return utils.SendSimpleSuccess(c, responses, "Subscription tiers retrieved successfully")
}

// GetSubscriptionTierByID returns a specific subscription tier
func GetSubscriptionTierByID(c *fiber.Ctx) error {
	db := config.DB
	tierID := c.Params("id")

	var tier models.SubscriptionTier
	if err := db.First(&tier, "id = ?", tierID).Error; err != nil {
		log.Printf("Error getting subscription tier %s: %v", tierID, err)
		return utils.SendNotFound(c, "Subscription tier not found")
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

	return utils.SendSimpleSuccess(c, response, "Subscription tier retrieved successfully")
}

// CreateSubscriptionTier creates a new subscription tier
func CreateSubscriptionTier(c *fiber.Ctx) error {
	var request models.CreateTierRequest
	if err := c.BodyParser(&request); err != nil {
		return utils.SendBadRequest(c, "Invalid request body")
	}

	// Use the handler from subscription_tier_handler.go
	return CreateTier(c)
}

// UpdateSubscriptionTier updates an existing subscription tier
func UpdateSubscriptionTier(c *fiber.Ctx) error {
	// Use the handler from subscription_tier_handler.go
	return UpdateTier(c)
}

// DeleteSubscriptionTier deletes a subscription tier
func DeleteSubscriptionTier(c *fiber.Ctx) error {
	db := config.DB
	tierID := c.Params("id")

	var tier models.SubscriptionTier
	if err := db.First(&tier, "id = ?", tierID).Error; err != nil {
		return utils.SendNotFound(c, "Subscription tier not found")
	}

	// Prevent deleting if less than 3 tiers would remain
	var tierCount int64
	db.Model(&models.SubscriptionTier{}).Where("is_active = ?", true).Count(&tierCount)
	if tierCount <= 3 {
		return utils.SendBadRequest(c, "Cannot delete tier: minimum 3 tiers required")
	}

	// Check if tier is in use by any organizations
	var orgCount int64
	db.Table("organizations").Where("subscription_tier = ?", tier.Name).Count(&orgCount)
	if orgCount > 0 {
		return utils.SendBadRequest(c, fmt.Sprintf("Cannot delete tier that is in use by %d organizations", orgCount))
	}

	if err := db.Delete(&tier).Error; err != nil {
		log.Printf("Error deleting subscription tier: %v", err)
		return utils.SendInternalError(c, "Failed to delete subscription tier", err)
	}

	return utils.SendSimpleSuccess(c, nil, "Subscription tier deleted successfully")
}

// GetAllSubscriptionFeatures returns all subscription features
// GetAllSubscriptionFeatures returns all subscription features
func GetAllSubscriptionFeatures(c *fiber.Ctx) error {
	db := config.DB

	var features []models.SubscriptionFeature
	if err := db.Order("category ASC, name ASC").Find(&features).Error; err != nil {
		log.Printf("Error getting subscription features: %v", err)
		return utils.SendInternalError(c, "Failed to retrieve subscription features", err)
	}

	return utils.SendSimpleSuccess(c, features, "Subscription features retrieved successfully")
}

// CreateSubscriptionFeature creates a new subscription feature
func CreateSubscriptionFeature(c *fiber.Ctx) error {
	db := config.DB

	var request struct {
		Name        string `json:"name" validate:"required,min=2,max=50"`
		DisplayName string `json:"display_name" validate:"required,min=2,max=100"`
		Description string `json:"description" validate:"required,min=10,max=500"`
		Category    string `json:"category" validate:"required,min=2,max=50"`
		IsActive    bool   `json:"is_active"`
	}

	if err := c.BodyParser(&request); err != nil {
		return utils.SendBadRequest(c, "Invalid request body")
	}

	feature := models.SubscriptionFeature{
		ID:          utils.GenerateID(),
		Name:        request.Name,
		DisplayName: request.DisplayName,
		Description: request.Description,
		Category:    request.Category,
		IsActive:    request.IsActive,
		CreatedAt:   time.Now(),
	}

	if err := db.Create(&feature).Error; err != nil {
		log.Printf("Error creating subscription feature: %v", err)
		return utils.SendInternalError(c, "Failed to create subscription feature", err)
	}

	return utils.SendSimpleSuccess(c, feature, "Subscription feature created successfully")
}

// UpdateSubscriptionFeature updates an existing subscription feature
func UpdateSubscriptionFeature(c *fiber.Ctx) error {
	db := config.DB
	featureID := c.Params("id")

	var feature models.SubscriptionFeature
	if err := db.First(&feature, "id = ?", featureID).Error; err != nil {
		return utils.SendNotFound(c, "Subscription feature not found")
	}

	var request struct {
		Name        *string `json:"name,omitempty"`
		DisplayName *string `json:"display_name,omitempty"`
		Description *string `json:"description,omitempty"`
		Category    *string `json:"category,omitempty"`
		IsActive    *bool   `json:"is_active,omitempty"`
	}

	if err := c.BodyParser(&request); err != nil {
		return utils.SendBadRequest(c, "Invalid request body")
	}

	// Update fields if provided
	if request.Name != nil {
		feature.Name = *request.Name
	}
	if request.DisplayName != nil {
		feature.DisplayName = *request.DisplayName
	}
	if request.Description != nil {
		feature.Description = *request.Description
	}
	if request.Category != nil {
		feature.Category = *request.Category
	}
	if request.IsActive != nil {
		feature.IsActive = *request.IsActive
	}

	if err := db.Save(&feature).Error; err != nil {
		log.Printf("Error updating subscription feature: %v", err)
		return utils.SendInternalError(c, "Failed to update subscription feature", err)
	}

	return utils.SendSimpleSuccess(c, feature, "Subscription feature updated successfully")
}

// DeleteSubscriptionFeature deletes a subscription feature
func DeleteSubscriptionFeature(c *fiber.Ctx) error {
	db := config.DB
	featureID := c.Params("id")

	var feature models.SubscriptionFeature
	if err := db.First(&feature, "id = ?", featureID).Error; err != nil {
		return utils.SendNotFound(c, "Subscription feature not found")
	}

	if err := db.Delete(&feature).Error; err != nil {
		log.Printf("Error deleting subscription feature: %v", err)
		return utils.SendInternalError(c, "Failed to delete subscription feature", err)
	}

	return utils.SendSimpleSuccess(c, nil, "Subscription feature deleted successfully")
}

// GetTrialOrganizations returns organizations with trial status for admin management
func GetTrialOrganizations(c *fiber.Ctx) error {
	db := config.DB

	var organizations []struct {
		ID                string     `json:"id"`
		Name              string     `json:"name"`
		TrialStartDate    *time.Time `json:"trial_start_date"`
		TrialEndDate      *time.Time `json:"trial_end_date"`
		DaysRemaining     int        `json:"days_remaining"`
		Status            string     `json:"status"`
		UserCount         int64      `json:"user_count"`
		SubscriptionTier  string     `json:"subscription_tier"`
		SubscriptionStatus string    `json:"subscription_status"`
	}

	// Get organizations with trial information (PostgreSQL syntax)
	query := `
		SELECT 
			o.id,
			o.name,
			o.trial_start_date,
			o.trial_end_date,
			CASE 
				WHEN o.trial_end_date IS NULL THEN 0
				ELSE CAST(EXTRACT(DAY FROM (o.trial_end_date - CURRENT_TIMESTAMP)) AS INTEGER)
			END as days_remaining,
			CASE 
				WHEN o.trial_end_date IS NULL THEN 'no_trial'
				WHEN o.trial_end_date < CURRENT_TIMESTAMP THEN 'expired'
				ELSE 'active'
			END as status,
			COALESCE(o.subscription_tier, 'starter') as subscription_tier,
			COALESCE(o.subscription_status, 'trial') as subscription_status
		FROM organizations o
		ORDER BY 
			CASE 
				WHEN o.trial_end_date IS NULL THEN 3
				WHEN o.trial_end_date < CURRENT_TIMESTAMP THEN 1
				ELSE 2
			END,
			o.trial_end_date ASC
	`

	if err := db.Raw(query).Scan(&organizations).Error; err != nil {
		log.Printf("Error getting trial organizations: %v", err)
		return utils.SendInternalError(c, "Failed to retrieve trial organizations", err)
	}

	// Get user count for each organization
	for i := range organizations {
		var userCount int64
		db.Table("users").Where("organization_id = ?", organizations[i].ID).Count(&userCount)
		organizations[i].UserCount = userCount
	}

	return utils.SendSimpleSuccess(c, organizations, "Trial organizations retrieved successfully")
}


// GetSubscriptionAnalytics returns comprehensive subscription analytics for admin
func GetSubscriptionAnalytics(c *fiber.Ctx) error {
	db := config.DB

	// Get revenue metrics from payments table
	var monthlyRevenue, yearlyRevenue float64
	
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	db.Table("payments").
		Where("payment_status = ? AND paid_at >= ?", "completed", thirtyDaysAgo).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&monthlyRevenue)
	
	oneYearAgo := time.Now().AddDate(-1, 0, 0)
	db.Table("payments").
		Where("payment_status = ? AND paid_at >= ?", "completed", oneYearAgo).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&yearlyRevenue)
	
	// Calculate growth
	sixtyDaysAgo := time.Now().AddDate(0, 0, -60)
	var previousMonthRevenue float64
	db.Table("payments").
		Where("payment_status = ? AND paid_at >= ? AND paid_at < ?", "completed", sixtyDaysAgo, thirtyDaysAgo).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&previousMonthRevenue)
	
	revenueGrowth := 0.0
	if previousMonthRevenue > 0 {
		revenueGrowth = ((monthlyRevenue - previousMonthRevenue) / previousMonthRevenue) * 100
	}
	
	revenue := map[string]interface{}{
		"monthly":  monthlyRevenue,
		"yearly":   yearlyRevenue,
		"growth":   revenueGrowth,
		"trend":    "up",
	}

	// Get subscription counts
	var totalSubs, newThisMonth int64
	db.Table("organizations").Where("subscription_status != ?", "trial").Count(&totalSubs)
	
	db.Table("organizations").
		Where("subscription_status != ? AND created_at > ?", "trial", thirtyDaysAgo).
		Count(&newThisMonth)

	// Get churned count from subscription events
	var churnedThisMonth int64
	db.Table("subscription_events").
		Where("event_type = ? AND created_at >= ?", "subscription_cancelled", thirtyDaysAgo).
		Count(&churnedThisMonth)
	
	// Calculate conversion rate from trial to paid
	var totalTrialConversions, totalTrialsStarted int64
	db.Table("subscription_events").
		Where("event_type = ?", "trial_converted").
		Count(&totalTrialConversions)
	
	db.Table("subscription_events").
		Where("event_type = ?", "trial_started").
		Count(&totalTrialsStarted)
	
	conversionRate := 0.0
	if totalTrialsStarted > 0 {
		conversionRate = (float64(totalTrialConversions) / float64(totalTrialsStarted)) * 100
	}

	subscriptions := map[string]interface{}{
		"total":              totalSubs,
		"new_this_month":     newThisMonth,
		"churned_this_month": churnedThisMonth,
		"conversion_rate":    conversionRate,
	}

	// Get tier distribution with revenue
	var tierStats []struct {
		Name  string `json:"name"`
		Count int64  `json:"count"`
	}

	db.Table("organizations").
		Select("COALESCE(subscription_tier, 'basic') as name, COUNT(*) as count").
		Where("subscription_status != ?", "trial").
		Group("name").
		Scan(&tierStats)

	// Get revenue by tier
	var tierRevenue []struct {
		Tier    string  `json:"tier"`
		Revenue float64 `json:"revenue"`
	}
	
	db.Table("payments").
		Select("subscription_tier as tier, SUM(amount) as revenue").
		Where("payment_status = ? AND paid_at >= ?", "completed", thirtyDaysAgo).
		Group("subscription_tier").
		Scan(&tierRevenue)
	
	// Merge tier data
	revenueMap := make(map[string]float64)
	for _, tr := range tierRevenue {
		revenueMap[tr.Tier] = tr.Revenue
	}
	
	tiers := []map[string]interface{}{}
	for _, stat := range tierStats {
		percentage := 0.0
		if totalSubs > 0 {
			percentage = (float64(stat.Count) / float64(totalSubs)) * 100
		}
		
		tiers = append(tiers, map[string]interface{}{
			"name":       stat.Name,
			"count":      stat.Count,
			"revenue":    revenueMap[stat.Name],
			"percentage": percentage,
		})
	}

	// Get trial analytics
	var activeTrials int64
	db.Table("organizations").Where("subscription_status = ?", "trial").Count(&activeTrials)
	
	var convertedThisMonth, expiredThisMonth int64
	db.Table("subscription_events").
		Where("event_type = ? AND created_at >= ?", "trial_converted", thirtyDaysAgo).
		Count(&convertedThisMonth)
	
	db.Table("subscription_events").
		Where("event_type = ? AND created_at >= ?", "trial_expired", thirtyDaysAgo).
		Count(&expiredThisMonth)
	
	// Calculate trial conversion rate for this month
	trialConversionRate := 0.0
	totalTrialsThisMonth := convertedThisMonth + expiredThisMonth
	if totalTrialsThisMonth > 0 {
		trialConversionRate = (float64(convertedThisMonth) / float64(totalTrialsThisMonth)) * 100
	}

	trials := map[string]interface{}{
		"active":                activeTrials,
		"converted_this_month":  convertedThisMonth,
		"expired_this_month":    expiredThisMonth,
		"conversion_rate":       trialConversionRate,
	}

	// Calculate key metrics
	var totalUsers int64
	db.Table("users").Count(&totalUsers)

	arpu := 0.0
	if totalUsers > 0 {
		arpu = monthlyRevenue / float64(totalUsers)
	}
	
	churnRate := 0.0
	if totalSubs > 0 {
		churnRate = (float64(churnedThisMonth) / float64(totalSubs)) * 100
	}
	
	ltv := 0.0
	if churnRate > 0 {
		ltv = arpu / (churnRate / 100)
	} else {
		ltv = arpu * 12
	}

	metrics := map[string]interface{}{
		"mrr":        monthlyRevenue,
		"arr":        monthlyRevenue * 12,
		"arpu":       arpu,
		"ltv":        ltv,
		"churn_rate": churnRate,
	}

	analytics := map[string]interface{}{
		"revenue":       revenue,
		"subscriptions": subscriptions,
		"tiers":         tiers,
		"trials":        trials,
		"metrics":       metrics,
		"generated_at":  time.Now(),
	}

	return utils.SendSimpleSuccess(c, analytics, "Subscription analytics retrieved successfully")
}

