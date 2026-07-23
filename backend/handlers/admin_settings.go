package handlers

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/utils"
)

// SystemSetting represents a system configuration setting
type SystemSetting struct {
	ID           string                 `json:"id" gorm:"primaryKey"`
	Key          string                 `json:"key" gorm:"uniqueIndex;not null"`
	Value        string                 `json:"value"`
	Type         string                 `json:"type"` // string, number, boolean, json, array
	Category     string                 `json:"category"`
	Description  string                 `json:"description"`
	DefaultValue string                 `json:"default_value"`
	IsRequired   bool                   `json:"is_required"`
	IsSecret     bool                   `json:"is_secret"`
	Environment  string                 `json:"environment"` // all, production, staging, development
	Validation   map[string]interface{} `json:"validation" gorm:"type:jsonb"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	CreatedBy    string                 `json:"created_by"`
	UpdatedBy    string                 `json:"updated_by"`
}

// EnvironmentVariable represents an environment variable
type EnvironmentVariable struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Key         string    `json:"key" gorm:"uniqueIndex;not null"`
	Value       string    `json:"value"`
	Environment string    `json:"environment"`
	IsSecret    bool      `json:"is_secret"`
	Description string    `json:"description"`
	IsRequired  bool      `json:"is_required"`
	Category    string    `json:"category"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	CreatedBy   string    `json:"created_by"`
	UpdatedBy   string    `json:"updated_by"`
}

// GetSystemSettings returns all system settings with optional filtering
func GetSystemSettings(c *fiber.Ctx) error {
	db := config.DB

	var settings []SystemSetting
	query := db.Model(&SystemSetting{})

	// Apply filters
	if search := c.Query("search"); search != "" {
		query = query.Where("key ILIKE ? OR description ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	if category := c.Query("category"); category != "" {
		query = query.Where("category = ?", category)
	}
	if environment := c.Query("environment"); environment != "" {
		query = query.Where("environment = ? OR environment = 'all'", environment)
	}
	if settingType := c.Query("type"); settingType != "" {
		query = query.Where("type = ?", settingType)
	}
	if isSecret := c.Query("is_secret"); isSecret != "" {
		query = query.Where("is_secret = ?", isSecret == "true")
	}
	if isRequired := c.Query("is_required"); isRequired != "" {
		query = query.Where("is_required = ?", isRequired == "true")
	}

	if err := query.Order("category, key").Find(&settings).Error; err != nil {
		log.Printf("Error getting system settings: %v", err)
		return utils.SendInternalError(c, "Failed to fetch system settings", err)
	}

	// Hide secret values
	for i := range settings {
		if settings[i].IsSecret {
			settings[i].Value = "***HIDDEN***"
		}
	}

	return utils.SendSimpleSuccess(c, settings, "System settings retrieved successfully")
}

// GetSystemSetting returns a single system setting by ID
func GetSystemSetting(c *fiber.Ctx) error {
	db := config.DB
	id := c.Params("id")

	var setting SystemSetting
	if err := db.First(&setting, "id = ?", id).Error; err != nil {
		return utils.SendNotFound(c, "System setting not found")
	}

	// Hide secret value
	if setting.IsSecret {
		setting.Value = "***HIDDEN***"
	}

	return utils.SendSimpleSuccess(c, setting, "System setting retrieved successfully")
}

// CreateSystemSetting creates a new system setting
func CreateSystemSetting(c *fiber.Ctx) error {
	db := config.DB

	var setting SystemSetting
	if err := c.BodyParser(&setting); err != nil {
		return utils.SendBadRequest(c, "Invalid request body")
	}

	// Set metadata
	setting.ID = utils.GenerateID()
	setting.CreatedAt = time.Now()
	setting.UpdatedAt = time.Now()
	setting.CreatedBy = c.Locals("userID").(string)
	setting.UpdatedBy = c.Locals("userID").(string)

	if err := db.Create(&setting).Error; err != nil {
		log.Printf("Error creating system setting: %v", err)
		return utils.SendInternalError(c, "Failed to create system setting", err)
	}

	// Hide secret value in response
	if setting.IsSecret {
		setting.Value = "***HIDDEN***"
	}

	return utils.SendSimpleSuccess(c, setting, "System setting created successfully")
}

// UpdateSystemSetting updates an existing system setting
func UpdateSystemSetting(c *fiber.Ctx) error {
	db := config.DB
	id := c.Params("id")

	var setting SystemSetting
	if err := db.First(&setting, "id = ?", id).Error; err != nil {
		return utils.SendNotFound(c, "System setting not found")
	}

	var updates SystemSetting
	if err := c.BodyParser(&updates); err != nil {
		return utils.SendBadRequest(c, "Invalid request body")
	}

	// Update allowed fields
	setting.Value = updates.Value
	setting.Description = updates.Description
	setting.IsRequired = updates.IsRequired
	setting.IsSecret = updates.IsSecret
	setting.Environment = updates.Environment
	setting.Validation = updates.Validation
	setting.UpdatedAt = time.Now()
	setting.UpdatedBy = c.Locals("userID").(string)

	if err := db.Save(&setting).Error; err != nil {
		log.Printf("Error updating system setting: %v", err)
		return utils.SendInternalError(c, "Failed to update system setting", err)
	}

	// Hide secret value in response
	if setting.IsSecret {
		setting.Value = "***HIDDEN***"
	}

	return utils.SendSimpleSuccess(c, setting, "System setting updated successfully")
}

// DeleteSystemSetting deletes a system setting
func DeleteSystemSetting(c *fiber.Ctx) error {
	db := config.DB
	id := c.Params("id")

	var setting SystemSetting
	if err := db.First(&setting, "id = ?", id).Error; err != nil {
		return utils.SendNotFound(c, "System setting not found")
	}

	if err := db.Delete(&setting).Error; err != nil {
		log.Printf("Error deleting system setting: %v", err)
		return utils.SendInternalError(c, "Failed to delete system setting", err)
	}

	return utils.SendSimpleSuccess(c, nil, "System setting deleted successfully")
}

// GetEnvironmentVariables returns environment variables
func GetEnvironmentVariables(c *fiber.Ctx) error {
	db := config.DB

	var envVars []EnvironmentVariable
	query := db.Model(&EnvironmentVariable{})

	if environment := c.Query("environment"); environment != "" {
		query = query.Where("environment = ?", environment)
	}

	if err := query.Order("category, key").Find(&envVars).Error; err != nil {
		log.Printf("Error getting environment variables: %v", err)
		return utils.SendInternalError(c, "Failed to fetch environment variables", err)
	}

	// Hide secret values
	for i := range envVars {
		if envVars[i].IsSecret {
			envVars[i].Value = "***HIDDEN***"
		}
	}

	return utils.SendSimpleSuccess(c, envVars, "Environment variables retrieved successfully")
}

// GetSystemHealth returns system health status
func GetSystemHealthStatus(c *fiber.Ctx) error {
	db := config.DB

	// Check database connectivity
	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("Error getting database connection: %v", err)
	}

	dbHealthy := true
	if sqlDB != nil {
		if err := sqlDB.Ping(); err != nil {
			dbHealthy = false
		}
	}

	// Get settings count for validation
	var settingsCount int64
	db.Model(&SystemSetting{}).Count(&settingsCount)

	// Get required settings that are missing values
	var missingRequired int64
	db.Model(&SystemSetting{}).
		Where("is_required = true AND (value = '' OR value IS NULL)").
		Count(&missingRequired)

	status := "healthy"
	score := 100
	checks := []map[string]interface{}{
		{
			"name":         "Database Connection",
			"status":       map[bool]string{true: "pass", false: "fail"}[dbHealthy],
			"message":      map[bool]string{true: "Database is accessible", false: "Database connection failed"}[dbHealthy],
			"last_checked": time.Now(),
		},
		{
			"name":         "Configuration Validation",
			"status":       map[bool]string{true: "pass", false: "fail"}[missingRequired == 0],
			"message":      map[bool]string{true: "All required settings are configured", false: "Some required settings are missing"}[missingRequired == 0],
			"last_checked": time.Now(),
		},
	}

	if !dbHealthy {
		status = "critical"
		score = 0
	} else if missingRequired > 0 {
		status = "warning"
		score = 75
	}

	recommendations := []string{}
	if missingRequired > 0 {
		recommendations = append(recommendations, "Configure missing required settings")
	}
	if !dbHealthy {
		recommendations = append(recommendations, "Check database connectivity")
	}

	health := map[string]interface{}{
		"status":          status,
		"score":           score,
		"checks":          checks,
		"recommendations": recommendations,
		"metrics": map[string]interface{}{
			"total_settings":     settingsCount,
			"missing_required":   missingRequired,
			"database_healthy":   dbHealthy,
		},
	}

	return utils.SendSimpleSuccess(c, health, "System health retrieved successfully")
}

// GetSettingsStats returns statistics about system settings
func GetSettingsStats(c *fiber.Ctx) error {
	db := config.DB

	var total int64
	db.Model(&SystemSetting{}).Count(&total)

	// Get counts by category
	var categoryStats []struct {
		Category string `json:"category"`
		Count    int64  `json:"count"`
	}
	db.Model(&SystemSetting{}).
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
	db.Model(&SystemSetting{}).
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
	db.Model(&SystemSetting{}).
		Select("type, COUNT(*) as count").
		Group("type").
		Scan(&typeStats)

	byType := make(map[string]int64)
	for _, stat := range typeStats {
		byType[stat.Type] = stat.Count
	}

	var secretSettings, requiredSettings, recentlyModified int64
	db.Model(&SystemSetting{}).Where("is_secret = true").Count(&secretSettings)
	db.Model(&SystemSetting{}).Where("is_required = true").Count(&requiredSettings)
	db.Model(&SystemSetting{}).Where("updated_at > ?", time.Now().AddDate(0, 0, -7)).Count(&recentlyModified)

	// Calculate health score based on required settings being set
	healthScore := 100.0
	if requiredSettings > 0 {
		var requiredWithValues int64
		db.Model(&SystemSetting{}).
			Where("is_required = true AND value IS NOT NULL AND value != ''").
			Count(&requiredWithValues)
		healthScore = (float64(requiredWithValues) / float64(requiredSettings)) * 100
	}

	stats := map[string]interface{}{
		"total":              total,
		"by_category":        byCategory,
		"by_environment":     byEnvironment,
		"by_type":            byType,
		"secret_settings":    secretSettings,
		"required_settings":  requiredSettings,
		"recently_modified":  recentlyModified,
		"health_score":       healthScore,
	}

	return utils.SendSimpleSuccess(c, stats, "Settings statistics retrieved successfully")
}