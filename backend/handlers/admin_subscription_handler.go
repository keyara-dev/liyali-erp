package handlers

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/utils"
)

// Admin Subscription Management Handlers
// These handlers provide comprehensive subscription management for administrators

// SubscriptionTier represents a subscription tier
type SubscriptionTier struct {
	ID               string    `json:"id" gorm:"primaryKey"`
	Name             string    `json:"name" gorm:"uniqueIndex"`
	DisplayName      string    `json:"display_name"`
	Description      string    `json:"description"`
	PriceMonthly     float64   `json:"price_monthly"`
	PriceYearly      float64   `json:"price_yearly"`
	MaxUsers         int       `json:"max_users"`
	MaxOrganizations *int      `json:"max_organizations,omitempty"`
	StorageLimitGB   int       `json:"storage_limit_gb"`
	Features         string    `json:"features"` // JSON array stored as string
	IsActive         bool      `json:"is_active"`
	SortOrder        int       `json:"sort_order"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// SubscriptionFeature represents a subscription feature
type SubscriptionFeature struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"uniqueIndex"`
	DisplayName string    `json:"display_name"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}

// GetAllSubscriptionTiers returns all subscription tiers for admin management
func GetAllSubscriptionTiers(c *fiber.Ctx) error {
	db := config.DB

	var tiers []SubscriptionTier
	if err := db.Order("sort_order ASC, created_at ASC").Find(&tiers).Error; err != nil {
		log.Printf("Error getting subscription tiers: %v", err)
		return utils.SendInternalError(c, "Failed to retrieve subscription tiers", err)
	}

	return utils.SendSimpleSuccess(c, tiers, "Subscription tiers retrieved successfully")
}

// GetSubscriptionTierByID returns a specific subscription tier
func GetSubscriptionTierByID(c *fiber.Ctx) error {
	db := config.DB
	tierID := c.Params("id")

	var tier SubscriptionTier
	if err := db.First(&tier, "id = ?", tierID).Error; err != nil {
		log.Printf("Error getting subscription tier %s: %v", tierID, err)
		return utils.SendNotFound(c, "Subscription tier not found")
	}

	return utils.SendSimpleSuccess(c, tier, "Subscription tier retrieved successfully")
}

// CreateSubscriptionTier creates a new subscription tier
func CreateSubscriptionTier(c *fiber.Ctx) error {
	db := config.DB

	var request struct {
		Name             string   `json:"name" validate:"required,min=2,max=50"`
		DisplayName      string   `json:"display_name" validate:"required,min=2,max=100"`
		Description      string   `json:"description" validate:"required,min=10,max=500"`
		PriceMonthly     float64  `json:"price_monthly" validate:"required,min=0"`
		PriceYearly      float64  `json:"price_yearly" validate:"required,min=0"`
		MaxUsers         int      `json:"max_users" validate:"required,min=1"`
		MaxOrganizations *int     `json:"max_organizations,omitempty"`
		StorageLimitGB   int      `json:"storage_limit_gb" validate:"required,min=1"`
		Features         []string `json:"features" validate:"required"`
		IsActive         bool     `json:"is_active"`
		SortOrder        int      `json:"sort_order" validate:"min=0"`
	}

	if err := c.BodyParser(&request); err != nil {
		return utils.SendBadRequest(c, "Invalid request body")
	}

	// Convert features array to JSON string
	featuresJSON := `["` + request.Features[0]
	for i := 1; i < len(request.Features); i++ {
		featuresJSON += `","` + request.Features[i]
	}
	featuresJSON += `"]`

	tier := SubscriptionTier{
		ID:               utils.GenerateID(),
		Name:             request.Name,
		DisplayName:      request.DisplayName,
		Description:      request.Description,
		PriceMonthly:     request.PriceMonthly,
		PriceYearly:      request.PriceYearly,
		MaxUsers:         request.MaxUsers,
		MaxOrganizations: request.MaxOrganizations,
		StorageLimitGB:   request.StorageLimitGB,
		Features:         featuresJSON,
		IsActive:         request.IsActive,
		SortOrder:        request.SortOrder,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := db.Create(&tier).Error; err != nil {
		log.Printf("Error creating subscription tier: %v", err)
		return utils.SendInternalError(c, "Failed to create subscription tier", err)
	}

	return utils.SendSimpleSuccess(c, tier, "Subscription tier created successfully")
}

// UpdateSubscriptionTier updates an existing subscription tier
func UpdateSubscriptionTier(c *fiber.Ctx) error {
	db := config.DB
	tierID := c.Params("id")

	var tier SubscriptionTier
	if err := db.First(&tier, "id = ?", tierID).Error; err != nil {
		return utils.SendNotFound(c, "Subscription tier not found")
	}

	var request struct {
		Name             *string   `json:"name,omitempty"`
		DisplayName      *string   `json:"display_name,omitempty"`
		Description      *string   `json:"description,omitempty"`
		PriceMonthly     *float64  `json:"price_monthly,omitempty"`
		PriceYearly      *float64  `json:"price_yearly,omitempty"`
		MaxUsers         *int      `json:"max_users,omitempty"`
		MaxOrganizations *int      `json:"max_organizations,omitempty"`
		StorageLimitGB   *int      `json:"storage_limit_gb,omitempty"`
		Features         []string  `json:"features,omitempty"`
		IsActive         *bool     `json:"is_active,omitempty"`
		SortOrder        *int      `json:"sort_order,omitempty"`
	}

	if err := c.BodyParser(&request); err != nil {
		return utils.SendBadRequest(c, "Invalid request body")
	}

	// Update fields if provided
	if request.Name != nil {
		tier.Name = *request.Name
	}
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
	if request.MaxUsers != nil {
		tier.MaxUsers = *request.MaxUsers
	}
	if request.MaxOrganizations != nil {
		tier.MaxOrganizations = request.MaxOrganizations
	}
	if request.StorageLimitGB != nil {
		tier.StorageLimitGB = *request.StorageLimitGB
	}
	if len(request.Features) > 0 {
		featuresJSON := `["` + request.Features[0]
		for i := 1; i < len(request.Features); i++ {
			featuresJSON += `","` + request.Features[i]
		}
		featuresJSON += `"]`
		tier.Features = featuresJSON
	}
	if request.IsActive != nil {
		tier.IsActive = *request.IsActive
	}
	if request.SortOrder != nil {
		tier.SortOrder = *request.SortOrder
	}

	tier.UpdatedAt = time.Now()

	if err := db.Save(&tier).Error; err != nil {
		log.Printf("Error updating subscription tier: %v", err)
		return utils.SendInternalError(c, "Failed to update subscription tier", err)
	}

	return utils.SendSimpleSuccess(c, tier, "Subscription tier updated successfully")
}

// DeleteSubscriptionTier deletes a subscription tier
func DeleteSubscriptionTier(c *fiber.Ctx) error {
	db := config.DB
	tierID := c.Params("id")

	var tier SubscriptionTier
	if err := db.First(&tier, "id = ?", tierID).Error; err != nil {
		return utils.SendNotFound(c, "Subscription tier not found")
	}

	// Check if tier is in use by any organizations
	var orgCount int64
	db.Table("organizations").Where("subscription_tier = ?", tier.Name).Count(&orgCount)
	if orgCount > 0 {
		return utils.SendBadRequest(c, "Cannot delete tier that is in use by organizations")
	}

	if err := db.Delete(&tier).Error; err != nil {
		log.Printf("Error deleting subscription tier: %v", err)
		return utils.SendInternalError(c, "Failed to delete subscription tier", err)
	}

	return utils.SendSimpleSuccess(c, nil, "Subscription tier deleted successfully")
}

// GetAllSubscriptionFeatures returns all subscription features
func GetAllSubscriptionFeatures(c *fiber.Ctx) error {
	db := config.DB

	var features []SubscriptionFeature
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

	feature := SubscriptionFeature{
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

	var feature SubscriptionFeature
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

	var feature SubscriptionFeature
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

	// Get organizations with trial information
	query := `
		SELECT 
			o.id,
			o.name,
			o.trial_start_date,
			o.trial_end_date,
			CASE 
				WHEN o.trial_end_date IS NULL THEN 0
				ELSE CAST(julianday(o.trial_end_date) - julianday('now') AS INTEGER)
			END as days_remaining,
			CASE 
				WHEN o.trial_end_date IS NULL THEN 'no_trial'
				WHEN o.trial_end_date < datetime('now') THEN 'expired'
				ELSE 'active'
			END as status,
			COALESCE(o.subscription_tier, 'basic') as subscription_tier,
			COALESCE(o.subscription_status, 'trial') as subscription_status
		FROM organizations o
		ORDER BY 
			CASE 
				WHEN o.trial_end_date IS NULL THEN 3
				WHEN o.trial_end_date < datetime('now') THEN 1
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

// ChangeOrganizationTier allows admin to change an organization's subscription tier
func ChangeOrganizationTier(c *fiber.Ctx) error {
	db := config.DB
	orgID := c.Params("id")

	var request struct {
		NewTier    string `json:"new_tier" validate:"required"`
		Reason     string `json:"reason" validate:"required,min=10,max=500"`
		OverrideLimits bool `json:"override_limits"`
	}

	if err := c.BodyParser(&request); err != nil {
		return utils.SendBadRequest(c, "Invalid request body")
	}

	// Verify organization exists
	var org struct {
		ID                 string `json:"id"`
		Name               string `json:"name"`
		SubscriptionTier   string `json:"subscription_tier"`
		SubscriptionStatus string `json:"subscription_status"`
	}

	if err := db.Table("organizations").
		Select("id, name, COALESCE(subscription_tier, 'basic') as subscription_tier, COALESCE(subscription_status, 'trial') as subscription_status").
		Where("id = ?", orgID).
		First(&org).Error; err != nil {
		return utils.SendNotFound(c, "Organization not found")
	}

	// Verify new tier exists
	var tier SubscriptionTier
	if err := db.First(&tier, "name = ? AND is_active = ?", request.NewTier, true).Error; err != nil {
		return utils.SendBadRequest(c, "Invalid or inactive subscription tier")
	}

	// Update organization tier
	updates := map[string]interface{}{
		"subscription_tier":   request.NewTier,
		"subscription_status": "active",
		"updated_at":         time.Now(),
	}

	if err := db.Table("organizations").Where("id = ?", orgID).Updates(updates).Error; err != nil {
		log.Printf("Error updating organization tier: %v", err)
		return utils.SendInternalError(c, "Failed to update organization tier", err)
	}

	// Log the tier change for audit purposes
	auditLog := map[string]interface{}{
		"id":             utils.GenerateID(),
		"organization_id": orgID,
		"action":         "tier_change",
		"old_value":      org.SubscriptionTier,
		"new_value":      request.NewTier,
		"reason":         request.Reason,
		"admin_user_id":  c.Locals("userID"),
		"created_at":     time.Now(),
	}

	db.Table("admin_audit_logs").Create(auditLog)

	response := map[string]interface{}{
		"organization_id": orgID,
		"organization_name": org.Name,
		"old_tier":       org.SubscriptionTier,
		"new_tier":       request.NewTier,
		"reason":         request.Reason,
		"changed_by":     c.Locals("userID"),
		"changed_at":     time.Now(),
	}

	return utils.SendSimpleSuccess(c, response, "Organization tier changed successfully")
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

// OverrideOrganizationLimits allows admin to override subscription limits for an organization
func OverrideOrganizationLimits(c *fiber.Ctx) error {
	db := config.DB
	orgID := c.Params("id")

	var request struct {
		MaxUsers       *int    `json:"max_users,omitempty"`
		StorageLimitGB *int    `json:"storage_limit_gb,omitempty"`
		Features       []string `json:"features,omitempty"`
		Reason         string  `json:"reason" validate:"required,min=10,max=500"`
		ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	}

	if err := c.BodyParser(&request); err != nil {
		return utils.SendBadRequest(c, "Invalid request body")
	}

	// Verify organization exists
	var org struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	if err := db.Table("organizations").
		Select("id, name").
		Where("id = ?", orgID).
		First(&org).Error; err != nil {
		return utils.SendNotFound(c, "Organization not found")
	}

	// Create or update organization override record
	override := map[string]interface{}{
		"id":              utils.GenerateID(),
		"organization_id": orgID,
		"max_users":       request.MaxUsers,
		"storage_limit_gb": request.StorageLimitGB,
		"reason":          request.Reason,
		"admin_user_id":   c.Locals("userID"),
		"expires_at":      request.ExpiresAt,
		"created_at":      time.Now(),
		"updated_at":      time.Now(),
	}

	if len(request.Features) > 0 {
		featuresJSON := `["` + request.Features[0]
		for i := 1; i < len(request.Features); i++ {
			featuresJSON += `","` + request.Features[i]
		}
		featuresJSON += `"]`
		override["features"] = featuresJSON
	}

	// Use UPSERT logic - delete existing and create new
	db.Table("organization_limit_overrides").Where("organization_id = ?", orgID).Delete(nil)
	
	if err := db.Table("organization_limit_overrides").Create(override).Error; err != nil {
		log.Printf("Error creating organization limit override: %v", err)
		return utils.SendInternalError(c, "Failed to create limit override", err)
	}

	// Log the override for audit purposes
	auditLog := map[string]interface{}{
		"id":             utils.GenerateID(),
		"organization_id": orgID,
		"action":         "limit_override",
		"details":        request,
		"reason":         request.Reason,
		"admin_user_id":  c.Locals("userID"),
		"created_at":     time.Now(),
	}

	db.Table("admin_audit_logs").Create(auditLog)

	response := map[string]interface{}{
		"organization_id":   orgID,
		"organization_name": org.Name,
		"overrides":         request,
		"reason":            request.Reason,
		"created_by":        c.Locals("userID"),
		"created_at":        time.Now(),
		"expires_at":        request.ExpiresAt,
	}

	return utils.SendSimpleSuccess(c, response, "Organization limits overridden successfully")
}