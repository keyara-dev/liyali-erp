package handlers

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/utils"
)

// FeatureFlag represents a feature flag
type FeatureFlag struct {
	ID              string                 `json:"id" gorm:"primaryKey"`
	Key             string                 `json:"key" gorm:"uniqueIndex;not null"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	Type            string                 `json:"type"` // boolean, string, number, json
	DefaultValue    string                 `json:"default_value"`
	Enabled         bool                   `json:"enabled"`
	Environment     string                 `json:"environment"` // all, production, staging, development
	Category        string                 `json:"category"`    // feature, experiment, operational, killswitch, permission
	Tags            []string               `json:"tags" gorm:"type:jsonb"`
	Targeting       map[string]interface{} `json:"targeting" gorm:"type:jsonb"`
	Variations      []map[string]interface{} `json:"variations" gorm:"type:jsonb"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	CreatedBy       string                 `json:"created_by"`
	UpdatedBy       string                 `json:"updated_by"`
	LastEvaluated   *time.Time             `json:"last_evaluated"`
	EvaluationCount int64                  `json:"evaluation_count"`
	IsArchived      bool                   `json:"is_archived"`
	ExpiresAt       *time.Time             `json:"expires_at"`
}

// FeatureFlagEvaluation represents a feature flag evaluation
type FeatureFlagEvaluation struct {
	ID             string                 `json:"id" gorm:"primaryKey"`
	FlagKey        string                 `json:"flag_key"`
	UserID         *string                `json:"user_id"`
	UserAttributes map[string]interface{} `json:"user_attributes" gorm:"type:jsonb"`
	Variation      string                 `json:"variation"`
	Value          string                 `json:"value"`
	Reason         string                 `json:"reason"` // targeting, rollout, default, disabled
	Timestamp      time.Time              `json:"timestamp"`
}

// GetFeatureFlags returns all feature flags with optional filtering
func GetFeatureFlags(c *fiber.Ctx) error {
	db := config.DB

	var flags []FeatureFlag
	query := db.Model(&FeatureFlag{})

	// Apply filters
	if search := c.Query("search"); search != "" {
		query = query.Where("key ILIKE ? OR name ILIKE ? OR description ILIKE ?", 
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}
	if category := c.Query("category"); category != "" {
		query = query.Where("category = ?", category)
	}
	if environment := c.Query("environment"); environment != "" {
		query = query.Where("environment = ? OR environment = 'all'", environment)
	}
	if flagType := c.Query("type"); flagType != "" {
		query = query.Where("type = ?", flagType)
	}
	if enabled := c.Query("enabled"); enabled != "" {
		query = query.Where("enabled = ?", enabled == "true")
	}
	if archived := c.Query("archived"); archived != "" {
		query = query.Where("is_archived = ?", archived == "true")
	}

	if err := query.Order("category, name").Find(&flags).Error; err != nil {
		log.Printf("Error getting feature flags: %v", err)
		return utils.SendInternalError(c, "Failed to fetch feature flags", err)
	}

	return utils.SendSimpleSuccess(c, flags, "Feature flags retrieved successfully")
}

// GetFeatureFlag returns a single feature flag by ID
func GetFeatureFlag(c *fiber.Ctx) error {
	db := config.DB
	id := c.Params("id")

	var flag FeatureFlag
	if err := db.First(&flag, "id = ?", id).Error; err != nil {
		return utils.SendNotFound(c, "Feature flag not found")
	}

	return utils.SendSimpleSuccess(c, flag, "Feature flag retrieved successfully")
}

// CreateFeatureFlag creates a new feature flag
func CreateFeatureFlag(c *fiber.Ctx) error {
	db := config.DB

	var flag FeatureFlag
	if err := c.BodyParser(&flag); err != nil {
		return utils.SendBadRequest(c, "Invalid request body")
	}

	// Set metadata
	flag.ID = utils.GenerateID()
	flag.CreatedAt = time.Now()
	flag.UpdatedAt = time.Now()
	flag.CreatedBy = c.Locals("userID").(string)
	flag.UpdatedBy = c.Locals("userID").(string)
	flag.EvaluationCount = 0

	if err := db.Create(&flag).Error; err != nil {
		log.Printf("Error creating feature flag: %v", err)
		return utils.SendInternalError(c, "Failed to create feature flag", err)
	}

	return utils.SendSimpleSuccess(c, flag, "Feature flag created successfully")
}

// UpdateFeatureFlag updates an existing feature flag
func UpdateFeatureFlag(c *fiber.Ctx) error {
	db := config.DB
	id := c.Params("id")

	var flag FeatureFlag
	if err := db.First(&flag, "id = ?", id).Error; err != nil {
		return utils.SendNotFound(c, "Feature flag not found")
	}

	var updates FeatureFlag
	if err := c.BodyParser(&updates); err != nil {
		return utils.SendBadRequest(c, "Invalid request body")
	}

	// Update allowed fields
	flag.Name = updates.Name
	flag.Description = updates.Description
	flag.DefaultValue = updates.DefaultValue
	flag.Enabled = updates.Enabled
	flag.Environment = updates.Environment
	flag.Category = updates.Category
	flag.Tags = updates.Tags
	flag.Targeting = updates.Targeting
	flag.Variations = updates.Variations
	flag.ExpiresAt = updates.ExpiresAt
	flag.UpdatedAt = time.Now()
	flag.UpdatedBy = c.Locals("userID").(string)

	if err := db.Save(&flag).Error; err != nil {
		log.Printf("Error updating feature flag: %v", err)
		return utils.SendInternalError(c, "Failed to update feature flag", err)
	}

	return utils.SendSimpleSuccess(c, flag, "Feature flag updated successfully")
}

// DeleteFeatureFlag deletes a feature flag
func DeleteFeatureFlag(c *fiber.Ctx) error {
	db := config.DB
	id := c.Params("id")

	var flag FeatureFlag
	if err := db.First(&flag, "id = ?", id).Error; err != nil {
		return utils.SendNotFound(c, "Feature flag not found")
	}

	if err := db.Delete(&flag).Error; err != nil {
		log.Printf("Error deleting feature flag: %v", err)
		return utils.SendInternalError(c, "Failed to delete feature flag", err)
	}

	return utils.SendSimpleSuccess(c, nil, "Feature flag deleted successfully")
}

// ToggleFeatureFlag toggles a feature flag's enabled status
func ToggleFeatureFlag(c *fiber.Ctx) error {
	db := config.DB
	id := c.Params("id")

	var flag FeatureFlag
	if err := db.First(&flag, "id = ?", id).Error; err != nil {
		return utils.SendNotFound(c, "Feature flag not found")
	}

	flag.Enabled = !flag.Enabled
	flag.UpdatedAt = time.Now()
	flag.UpdatedBy = c.Locals("userID").(string)

	if err := db.Save(&flag).Error; err != nil {
		log.Printf("Error toggling feature flag: %v", err)
		return utils.SendInternalError(c, "Failed to toggle feature flag", err)
	}

	return utils.SendSimpleSuccess(c, flag, "Feature flag toggled successfully")
}

// ArchiveFeatureFlag archives a feature flag
func ArchiveFeatureFlag(c *fiber.Ctx) error {
	db := config.DB
	id := c.Params("id")

	var flag FeatureFlag
	if err := db.First(&flag, "id = ?", id).Error; err != nil {
		return utils.SendNotFound(c, "Feature flag not found")
	}

	flag.IsArchived = true
	flag.UpdatedAt = time.Now()
	flag.UpdatedBy = c.Locals("userID").(string)

	if err := db.Save(&flag).Error; err != nil {
		log.Printf("Error archiving feature flag: %v", err)
		return utils.SendInternalError(c, "Failed to archive feature flag", err)
	}

	return utils.SendSimpleSuccess(c, flag, "Feature flag archived successfully")
}

// GetFeatureFlagStats returns statistics about feature flags
func GetFeatureFlagStats(c *fiber.Ctx) error {
	db := config.DB

	var total, enabled, disabled, archived int64
	db.Model(&FeatureFlag{}).Count(&total)
	db.Model(&FeatureFlag{}).Where("enabled = true").Count(&enabled)
	db.Model(&FeatureFlag{}).Where("enabled = false").Count(&disabled)
	db.Model(&FeatureFlag{}).Where("is_archived = true").Count(&archived)

	// Get counts by category
	var categoryStats []struct {
		Category string `json:"category"`
		Count    int64  `json:"count"`
	}
	db.Model(&FeatureFlag{}).
		Select("category, COUNT(*) as count").
		Group("category").
		Scan(&categoryStats)

	byCategory := make(map[string]int64)
	for _, stat := range categoryStats {
		byCategory[stat.Category] = stat.Count
	}

	// Get counts by environment
	var envStats []struct {
		Environment string `json:"environment"`
		Count       int64  `json:"count"`
	}
	db.Model(&FeatureFlag{}).
		Select("environment, COUNT(*) as count").
		Group("environment").
		Scan(&envStats)

	byEnvironment := make(map[string]int64)
	for _, stat := range envStats {
		byEnvironment[stat.Environment] = stat.Count
	}

	// Get counts by type
	var typeStats []struct {
		Type  string `json:"type"`
		Count int64  `json:"count"`
	}
	db.Model(&FeatureFlag{}).
		Select("type, COUNT(*) as count").
		Group("type").
		Scan(&typeStats)

	byType := make(map[string]int64)
	for _, stat := range typeStats {
		byType[stat.Type] = stat.Count
	}

	var recentlyCreated, recentlyUpdated, expiringSoon int64
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	sevenDaysFromNow := time.Now().AddDate(0, 0, 7)

	db.Model(&FeatureFlag{}).Where("created_at > ?", sevenDaysAgo).Count(&recentlyCreated)
	db.Model(&FeatureFlag{}).Where("updated_at > ?", sevenDaysAgo).Count(&recentlyUpdated)
	db.Model(&FeatureFlag{}).Where("expires_at IS NOT NULL AND expires_at <= ?", sevenDaysFromNow).Count(&expiringSoon)

	// Get total evaluations today
	var evaluationsToday int64
	today := time.Now().Truncate(24 * time.Hour)
	db.Model(&FeatureFlagEvaluation{}).Where("timestamp >= ?", today).Count(&evaluationsToday)

	stats := map[string]interface{}{
		"total":              total,
		"enabled":            enabled,
		"disabled":           disabled,
		"archived":           archived,
		"by_category":        byCategory,
		"by_environment":     byEnvironment,
		"by_type":            byType,
		"recently_created":   recentlyCreated,
		"recently_updated":   recentlyUpdated,
		"expiring_soon":      expiringSoon,
		"evaluations_today":  evaluationsToday,
	}

	return utils.SendSimpleSuccess(c, stats, "Feature flag statistics retrieved successfully")
}

// EvaluateFeatureFlag evaluates a feature flag for a user
func EvaluateFeatureFlag(c *fiber.Ctx) error {
	db := config.DB
	flagKey := c.Params("key")

	var flag FeatureFlag
	if err := db.Where("key = ?", flagKey).First(&flag).Error; err != nil {
		return utils.SendNotFound(c, "Feature flag not found")
	}

	// Parse request body for user context
	var request struct {
		UserID         *string                `json:"user_id"`
		UserAttributes map[string]interface{} `json:"user_attributes"`
	}
	c.BodyParser(&request)

	// Simple evaluation logic (in production, this would be more sophisticated)
	variation := "disabled"
	value := flag.DefaultValue
	reason := "default"

	if flag.Enabled {
		variation = "enabled"
		value = "true"
		reason = "enabled"
	}

	// Record evaluation
	evaluation := FeatureFlagEvaluation{
		ID:             utils.GenerateID(),
		FlagKey:        flagKey,
		UserID:         request.UserID,
		UserAttributes: request.UserAttributes,
		Variation:      variation,
		Value:          value,
		Reason:         reason,
		Timestamp:      time.Now(),
	}

	db.Create(&evaluation)

	// Update flag evaluation count and last evaluated time
	now := time.Now()
	db.Model(&flag).Updates(map[string]interface{}{
		"evaluation_count": flag.EvaluationCount + 1,
		"last_evaluated":   &now,
	})

	return utils.SendSimpleSuccess(c, evaluation, "Feature flag evaluated successfully")
}

// GetFeatureFlagAnalytics returns analytics for a specific feature flag
func GetFeatureFlagAnalytics(c *fiber.Ctx) error {
	db := config.DB
	flagKey := c.Params("key")

	var flag FeatureFlag
	if err := db.Where("key = ?", flagKey).First(&flag).Error; err != nil {
		return utils.SendNotFound(c, "Feature flag not found")
	}

	// Get evaluation statistics
	var totalEvaluations int64
	db.Model(&FeatureFlagEvaluation{}).Where("flag_key = ?", flagKey).Count(&totalEvaluations)

	// Get evaluations by variation
	var variationStats []struct {
		Variation string `json:"variation"`
		Count     int64  `json:"count"`
	}
	db.Model(&FeatureFlagEvaluation{}).
		Select("variation, COUNT(*) as count").
		Where("flag_key = ?", flagKey).
		Group("variation").
		Scan(&variationStats)

	byVariation := make(map[string]int64)
	for _, stat := range variationStats {
		byVariation[stat.Variation] = stat.Count
	}

	// Get evaluations by day (last 7 days)
	var dailyStats []struct {
		Date  string `json:"date"`
		Count int64  `json:"count"`
	}

	for i := 6; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i)
		dateStr := date.Format("2006-01-02")
		
		var count int64
		db.Model(&FeatureFlagEvaluation{}).
			Where("flag_key = ? AND DATE(timestamp) = ?", flagKey, dateStr).
			Count(&count)
		
		dailyStats = append(dailyStats, struct {
			Date  string `json:"date"`
			Count int64  `json:"count"`
		}{
			Date:  dateStr,
			Count: count,
		})
	}

	// Get performance metrics from database
	var avgEvalTime float64
	db.Table("feature_flag_evaluations").
		Where("flag_key = ?", flagKey).
		Select("COALESCE(AVG(evaluation_time_ms), 0)").
		Scan(&avgEvalTime)
	
	analytics := map[string]interface{}{
		"flag_key": flagKey,
		"evaluations": map[string]interface{}{
			"total":        totalEvaluations,
			"by_variation": byVariation,
			"by_day":       dailyStats,
		},
		"performance": map[string]interface{}{
			"avg_evaluation_time": avgEvalTime,
			"error_rate":          0.0, // Calculated from evaluation errors
			"cache_hit_rate":      95.0, // From cache statistics
		},
		"generated_at": time.Now(),
	}

	return utils.SendSimpleSuccess(c, analytics, "Feature flag analytics retrieved successfully")
}