package handlers

import (
	"fmt"
	"log"
	"math"
	"regexp"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/utils"
	"golang.org/x/crypto/bcrypt"
)

// AdminGetAllOrganizations returns all organizations with filters and pagination
func AdminGetAllOrganizations(c *fiber.Ctx) error {
	db := config.DB

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	search := c.Query("search")
	status := c.Query("status")
	subscriptionTier := c.Query("subscription_tier")
	trialStatus := c.Query("trial_status")
	sortBy := c.Query("sort_by", "created_at")
	sortOrder := c.Query("sort_order", "desc")

	page, limit = utils.NormalizePaginationParams(page, limit)
	offset := (page - 1) * limit

	// Build query
	query := db.Table("organizations").
		Select(`organizations.id, organizations.name,
			COALESCE(organizations.slug, '') as domain,
			organizations.created_at, organizations.updated_at,
			CASE WHEN organizations.active = true THEN 'active' ELSE 'suspended' END as status,
			COALESCE(organizations.subscription_tier, organizations.tier, 'basic') as subscription_tier,
			CASE
				WHEN COALESCE(organizations.subscription_tier, organizations.tier, 'starter') IN ('pro', 'enterprise') THEN 'subscribed'
				WHEN COALESCE(organizations.subscription_status, 'trial') = 'active' THEN 'subscribed'
				WHEN organizations.trial_end_date IS NOT NULL AND organizations.trial_end_date < CURRENT_TIMESTAMP THEN 'expired'
				ELSE COALESCE(organizations.subscription_status, 'trial')
			END as trial_status,
			organizations.trial_start_date,
			CASE
				WHEN COALESCE(organizations.subscription_tier, organizations.tier, 'starter') IN ('pro', 'enterprise') THEN NULL
				ELSE organizations.trial_end_date
			END as trial_end_date,
			(SELECT COUNT(*) FROM organization_members WHERE organization_members.organization_id = organizations.id AND organization_members.active = true) as user_count`)

	// Apply filters
	if search != "" {
		searchTerm := "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(organizations.name) LIKE ? OR LOWER(organizations.slug) LIKE ?", searchTerm, searchTerm)
	}

	if status != "" && status != "all" {
		if status == "active" {
			query = query.Where("organizations.active = ?", true)
		} else if status == "suspended" {
			query = query.Where("organizations.active = ?", false)
		}
	}

	if subscriptionTier != "" {
		query = query.Where("COALESCE(organizations.subscription_tier, organizations.tier, 'basic') = ?", subscriptionTier)
	}

	if trialStatus != "" {
		query = query.Where("COALESCE(organizations.subscription_status, 'trial') = ?", trialStatus)
	}

	// Count total
	var total int64
	countQuery := db.Table("organizations")
	if search != "" {
		searchTerm := "%" + strings.ToLower(search) + "%"
		countQuery = countQuery.Where("LOWER(name) LIKE ? OR LOWER(slug) LIKE ?", searchTerm, searchTerm)
	}
	if status != "" && status != "all" {
		if status == "active" {
			countQuery = countQuery.Where("active = ?", true)
		} else if status == "suspended" {
			countQuery = countQuery.Where("active = ?", false)
		}
	}
	countQuery.Count(&total)

	// Apply sorting
	allowedSorts := map[string]string{
		"name":       "organizations.name",
		"created_at": "organizations.created_at",
		"user_count": "user_count",
	}
	sortCol, ok := allowedSorts[sortBy]
	if !ok {
		sortCol = "organizations.created_at"
	}
	if sortOrder != "asc" {
		sortOrder = "desc"
	}
	query = query.Order(sortCol + " " + sortOrder)

	// Execute
	var organizations []map[string]interface{}
	if err := query.Offset(offset).Limit(limit).Find(&organizations).Error; err != nil {
		log.Printf("Error getting organizations: %v", err)
		return utils.SendInternalError(c, "Failed to retrieve organizations", err)
	}

	// Calculate days remaining for trial orgs
	for i := range organizations {
		if endDate, ok := organizations[i]["trial_end_date"]; ok && endDate != nil {
			if t, ok := endDate.(time.Time); ok {
				days := int(time.Until(t).Hours() / 24)
				if days < 0 {
					days = 0
				}
				organizations[i]["days_remaining"] = days
			}
		}
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	response := map[string]interface{}{
		"organizations": organizations,
		"total":         total,
		"page":          page,
		"limit":         limit,
		"totalPages":    totalPages,
	}

	return utils.SendSimpleSuccess(c, response, "Organizations retrieved successfully")
}

// AdminGetOrganizationStatistics returns organization statistics
func AdminGetOrganizationStatistics(c *fiber.Ctx) error {
	db := config.DB

	var totalOrgs, activeOrgs, suspendedOrgs, trialOrgs, createdThisMonth int64

	db.Table("organizations").Count(&totalOrgs)
	db.Table("organizations").Where("active = ?", true).Count(&activeOrgs)
	db.Table("organizations").Where("active = ?", false).Count(&suspendedOrgs)
	db.Table("organizations").Where("COALESCE(subscription_status, 'trial') = ?", "trial").Count(&trialOrgs)

	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	db.Table("organizations").Where("created_at >= ?", thirtyDaysAgo).Count(&createdThisMonth)

	var totalUsersAcrossOrgs int64
	db.Table("organization_members").Where("active = ?", true).Count(&totalUsersAcrossOrgs)

	// Trials expiring soon (within 7 days)
	sevenDaysFromNow := time.Now().AddDate(0, 0, 7)
	var trialsExpiringSoon int64
	db.Table("organizations").
		Where("trial_end_date IS NOT NULL AND trial_end_date <= ? AND trial_end_date > ?", sevenDaysFromNow, time.Now()).
		Count(&trialsExpiringSoon)

	// Top organizations by user count
	var topOrgs []map[string]interface{}
	db.Table("organization_members").
		Select("organization_id, COUNT(*) as user_count").
		Where("active = ?", true).
		Group("organization_id").
		Order("user_count DESC").
		Limit(5).
		Find(&topOrgs)

	// Enrich with org names
	for i := range topOrgs {
		if orgID, ok := topOrgs[i]["organization_id"].(string); ok {
			var orgName string
			db.Table("organizations").Where("id = ?", orgID).Pluck("name", &orgName)
			topOrgs[i]["organization_name"] = orgName
		}
	}

	stats := map[string]interface{}{
		"total_organizations":              totalOrgs,
		"active_organizations":             activeOrgs,
		"suspended_organizations":          suspendedOrgs,
		"trial_organizations":              trialOrgs,
		"organizations_created_this_month": createdThisMonth,
		"total_users_across_organizations": totalUsersAcrossOrgs,
		"trials_expiring_soon":             trialsExpiringSoon,
		"top_organizations_by_users":       topOrgs,
	}

	return utils.SendSimpleSuccess(c, stats, "Organization statistics retrieved successfully")
}

// AdminGetOrganizationById returns a single organization by ID
func AdminGetOrganizationById(c *fiber.Ctx) error {
	db := config.DB
	orgID := c.Params("id")

	var org map[string]interface{}
	err := db.Table("organizations").
		Select(`organizations.id, organizations.name,
			COALESCE(organizations.slug, '') as domain,
			organizations.description,
			organizations.created_at, organizations.updated_at,
			CASE WHEN organizations.active = true THEN 'active' ELSE 'suspended' END as status,
			COALESCE(organizations.subscription_tier, organizations.tier, 'basic') as subscription_tier,
			COALESCE(organizations.subscription_status, 'trial') as trial_status,
			organizations.trial_start_date, organizations.trial_end_date`).
		Where("organizations.id = ?", orgID).
		First(&org).Error

	if err != nil {
		return utils.SendNotFound(c, "Organization not found")
	}

	// Get user count
	var userCount int64
	db.Table("organization_members").Where("organization_id = ? AND active = ?", orgID, true).Count(&userCount)
	org["user_count"] = userCount

	// Calculate days remaining
	if endDate, ok := org["trial_end_date"]; ok && endDate != nil {
		if t, ok := endDate.(time.Time); ok {
			days := int(time.Until(t).Hours() / 24)
			if days < 0 {
				days = 0
			}
			org["days_remaining"] = days
		}
	}

	// Get settings
	var settings map[string]interface{}
	db.Table("organization_settings").Where("organization_id = ?", orgID).First(&settings)
	if settings != nil {
		org["settings"] = settings
	}

	// Add contact_info and billing_info stubs (frontend expects these nested objects)
	org["contact_info"] = map[string]interface{}{
		"email":   nil,
		"phone":   nil,
		"address": nil,
		"city":    nil,
		"country": nil,
	}
	org["billing_info"] = map[string]interface{}{
		"billing_email":   nil,
		"payment_method":  nil,
		"billing_address": nil,
	}

	return utils.SendSimpleSuccess(c, org, "Organization retrieved successfully")
}

// AdminCreateOrganization creates a new organization with admin user
func AdminCreateOrganization(c *fiber.Ctx) error {
	db := config.DB

	var request struct {
		Name             string `json:"name"`
		Domain           string `json:"domain"`
		Description      string `json:"description"`
		AdminUserID      string `json:"admin_user_id"`
		AdminName        string `json:"admin_name"`
		AdminEmail       string `json:"admin_email"`
		SubscriptionTier string `json:"subscription_tier"`
		TrialDays        int    `json:"trial_days"`
		MaxUsers         int    `json:"max_users"`
	}

	if err := c.BodyParser(&request); err != nil {
		return utils.SendBadRequest(c, "Invalid request body")
	}

	if request.Name == "" {
		return utils.SendBadRequest(c, "Organization name is required")
	}
	if request.AdminUserID == "" && request.AdminEmail == "" {
		return utils.SendBadRequest(c, "Either admin_user_id or admin_email is required")
	}

	// Generate slug from name
	slug := strings.ToLower(request.Name)
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	slug = reg.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")

	if request.Domain != "" {
		slug = request.Domain
	}

	// Check slug uniqueness
	var existingCount int64
	db.Table("organizations").Where("slug = ?", slug).Count(&existingCount)
	if existingCount > 0 {
		slug = fmt.Sprintf("%s-%s", slug, utils.GenerateID()[:6])
	}

	tier := request.SubscriptionTier
	if tier == "" {
		tier = "basic"
	}

	trialDays := request.TrialDays
	if trialDays == 0 {
		trialDays = 30
	}

	now := time.Now()
	orgID := utils.GenerateID()

	// Create organization
	org := map[string]interface{}{
		"id":                  orgID,
		"name":                request.Name,
		"slug":                slug,
		"description":         request.Description,
		"active":              true,
		"tier":                tier,
		"subscription_tier":   tier,
		"subscription_status": "trial",
		"trial_start_date":    now,
		"trial_end_date":      now.AddDate(0, 0, trialDays),
		"created_by":          c.Locals("userID"),
		"created_at":          now,
		"updated_at":          now,
	}

	if err := db.Table("organizations").Create(org).Error; err != nil {
		log.Printf("Error creating organization: %v", err)
		return utils.SendInternalError(c, "Failed to create organization", err)
	}

	// Resolve admin user — prefer admin_user_id, fall back to email lookup/create
	var resolvedUserID string

	if request.AdminUserID != "" {
		// Verify the user exists
		var userCount int64
		db.Table("users").Where("id = ?", request.AdminUserID).Count(&userCount)
		if userCount == 0 {
			return utils.SendNotFound(c, "Admin user not found")
		}
		resolvedUserID = request.AdminUserID
	} else {
		// Email-based lookup or create
		adminName := request.AdminName
		if adminName == "" {
			adminName = "Admin"
		}

		db.Table("users").Where("email = ?", request.AdminEmail).Pluck("id", &resolvedUserID)

		if resolvedUserID == "" {
			tempPassword := utils.GenerateID()[:12]
			hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(tempPassword), bcrypt.DefaultCost)

			userID := utils.GenerateID()
			user := map[string]interface{}{
				"id":                      userID,
				"email":                   request.AdminEmail,
				"name":                    adminName,
				"password":                string(hashedPassword),
				"role":                    "admin",
				"active":                  true,
				"current_organization_id": orgID,
				"created_at":              now,
				"updated_at":              now,
			}
			db.Table("users").Create(user)
			resolvedUserID = userID
		}
	}

	// Add user as org admin member
	member := map[string]interface{}{
		"id":              utils.GenerateID(),
		"organization_id": orgID,
		"user_id":         resolvedUserID,
		"role":            "admin",
		"active":          true,
		"joined_at":       now,
		"created_at":      now,
		"updated_at":      now,
	}
	db.Table("organization_members").Create(member)

	// Create default settings
	maxUsers := request.MaxUsers
	if maxUsers <= 0 {
		maxUsers = 50
	}
	settings := map[string]interface{}{
		"id":              utils.GenerateID(),
		"organization_id": orgID,
		"currency":        "USD",
		"max_users":       maxUsers,
		"created_at":      now,
		"updated_at":      now,
	}
	db.Table("organization_settings").Create(settings)

	// Audit log
	auditLog := map[string]interface{}{
		"id":              utils.GenerateID(),
		"organization_id": orgID,
		"action":          "organization_created",
		"admin_user_id":   c.Locals("userID"),
		"created_at":      now,
	}
	db.Table("admin_audit_logs").Create(auditLog)

	org["user_count"] = 1
	org["status"] = "active"

	return utils.SendCreatedSuccess(c, org, "Organization created successfully")
}

// AdminUpdateOrganization updates an organization
func AdminUpdateOrganization(c *fiber.Ctx) error {
	db := config.DB
	orgID := c.Params("id")

	// Verify org exists
	var existingCount int64
	db.Table("organizations").Where("id = ?", orgID).Count(&existingCount)
	if existingCount == 0 {
		return utils.SendNotFound(c, "Organization not found")
	}

	var request struct {
		Name             *string                `json:"name,omitempty"`
		Domain           *string                `json:"domain,omitempty"`
		SubscriptionTier *string                `json:"subscription_tier,omitempty"`
		Settings         map[string]interface{} `json:"settings,omitempty"`
		ContactInfo      map[string]interface{} `json:"contact_info,omitempty"`
	}

	if err := c.BodyParser(&request); err != nil {
		return utils.SendBadRequest(c, "Invalid request body")
	}

	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}

	if request.Name != nil {
		updates["name"] = *request.Name
	}
	if request.Domain != nil {
		updates["slug"] = *request.Domain
	}
	if request.SubscriptionTier != nil {
		updates["subscription_tier"] = *request.SubscriptionTier
		updates["tier"] = *request.SubscriptionTier
	}

	if err := db.Table("organizations").Where("id = ?", orgID).Updates(updates).Error; err != nil {
		log.Printf("Error updating organization: %v", err)
		return utils.SendInternalError(c, "Failed to update organization", err)
	}

	// Audit log
	auditLog := map[string]interface{}{
		"id":              utils.GenerateID(),
		"organization_id": orgID,
		"action":          "organization_updated",
		"admin_user_id":   c.Locals("userID"),
		"created_at":      time.Now(),
	}
	db.Table("admin_audit_logs").Create(auditLog)

	return utils.SendSimpleSuccess(c, map[string]interface{}{"id": orgID}, "Organization updated successfully")
}

// AdminUpdateOrganizationStatus updates an organization's status
func AdminUpdateOrganizationStatus(c *fiber.Ctx) error {
	db := config.DB
	orgID := c.Params("id")

	var request struct {
		Status string `json:"status"`
		Reason string `json:"reason"`
	}

	if err := c.BodyParser(&request); err != nil {
		return utils.SendBadRequest(c, "Invalid request body")
	}

	reqStatus := strings.ToUpper(request.Status)
	if reqStatus != "ACTIVE" && reqStatus != "SUSPENDED" && reqStatus != "PENDING" {
		return utils.SendBadRequest(c, "Invalid status. Must be 'active', 'suspended', or 'pending'")
	}

	active := reqStatus == "ACTIVE"

	if err := db.Table("organizations").Where("id = ?", orgID).Updates(map[string]interface{}{
		"active":     active,
		"updated_at": time.Now(),
	}).Error; err != nil {
		return utils.SendInternalError(c, "Failed to update organization status", err)
	}

	// Audit log
	auditLog := map[string]interface{}{
		"id":              utils.GenerateID(),
		"organization_id": orgID,
		"action":          "organization_status_changed",
		"old_value":       "",
		"new_value":       request.Status,
		"reason":          request.Reason,
		"admin_user_id":   c.Locals("userID"),
		"created_at":      time.Now(),
	}
	db.Table("admin_audit_logs").Create(auditLog)

	return utils.SendSimpleSuccess(c, map[string]interface{}{
		"id":     orgID,
		"status": request.Status,
	}, "Organization status updated successfully")
}

// AdminGetOrganizationUsers returns users belonging to an organization
func AdminGetOrganizationUsers(c *fiber.Ctx) error {
	db := config.DB
	orgID := c.Params("id")

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)
	page, limit = utils.NormalizePaginationParams(page, limit)
	offset := (page - 1) * limit

	var total int64
	db.Table("organization_members").Where("organization_id = ?", orgID).Count(&total)

	var users []map[string]interface{}
	err := db.Table("organization_members").
		Select(`organization_members.id, users.name, users.email,
			organization_members.role,
			CASE WHEN organization_members.active = true THEN 'active' ELSE 'suspended' END as status,
			organization_members.joined_at, users.last_login,
			CASE WHEN organization_members.role = 'admin' THEN true ELSE false END as is_admin`).
		Joins("LEFT JOIN users ON users.id = organization_members.user_id").
		Where("organization_members.organization_id = ?", orgID).
		Offset(offset).Limit(limit).
		Order("organization_members.created_at DESC").
		Find(&users).Error

	if err != nil {
		log.Printf("Error getting organization users: %v", err)
		return utils.SendInternalError(c, "Failed to retrieve organization users", err)
	}

	response := map[string]interface{}{
		"users": users,
		"total": total,
		"page":  page,
		"limit": limit,
	}

	return utils.SendSimpleSuccess(c, response, "Organization users retrieved successfully")
}

// AdminGetOrganizationActivity returns activity logs for an organization
func AdminGetOrganizationActivity(c *fiber.Ctx) error {
	db := config.DB
	orgID := c.Params("id")

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)
	page, limit = utils.NormalizePaginationParams(page, limit)
	offset := (page - 1) * limit

	var total int64
	db.Table("admin_audit_logs").Where("organization_id = ?", orgID).Count(&total)

	var activities []map[string]interface{}
	db.Table("admin_audit_logs").
		Where("organization_id = ?", orgID).
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&activities)

	// Map fields to match frontend OrganizationActivity interface
	for i := range activities {
		activities[i]["timestamp"] = activities[i]["created_at"]
		if _, ok := activities[i]["description"]; !ok {
			activities[i]["description"] = activities[i]["action"]
		}
	}

	response := map[string]interface{}{
		"activities": activities,
		"total":      total,
		"page":       page,
		"limit":      limit,
	}

	return utils.SendSimpleSuccess(c, response, "Organization activity retrieved successfully")
}

// AdminGetOrgTrialStatus returns the trial status for an organization
func AdminGetOrgTrialStatus(c *fiber.Ctx) error {
	db := config.DB
	orgID := c.Params("id")

	var org map[string]interface{}
	err := db.Table("organizations").
		Select(`id, name,
			COALESCE(subscription_status, 'trial') as trial_status,
			trial_start_date, trial_end_date,
			COALESCE(subscription_tier, tier, 'basic') as subscription_tier`).
		Where("id = ?", orgID).
		First(&org).Error

	if err != nil {
		return utils.SendNotFound(c, "Organization not found")
	}

	// Calculate days remaining
	if endDate, ok := org["trial_end_date"]; ok && endDate != nil {
		if t, ok := endDate.(time.Time); ok {
			days := int(time.Until(t).Hours() / 24)
			if days < 0 {
				days = 0
			}
			org["days_remaining"] = days
		}
	}

	return utils.SendSimpleSuccess(c, org, "Trial status retrieved successfully")
}

// AdminGetOrgSubscription returns subscription details for an organization
func AdminGetOrgSubscription(c *fiber.Ctx) error {
	db := config.DB
	orgID := c.Params("id")

	var org map[string]interface{}
	err := db.Table("organizations").
		Select(`id, name,
			COALESCE(subscription_tier, tier, 'basic') as subscription_tier,
			COALESCE(subscription_status, 'trial') as subscription_status,
			trial_start_date, trial_end_date`).
		Where("id = ?", orgID).
		First(&org).Error

	if err != nil {
		return utils.SendNotFound(c, "Organization not found")
	}

	// Get tier details
	tierName := "basic"
	if t, ok := org["subscription_tier"].(string); ok {
		tierName = t
	}
	var tier models.SubscriptionTier
	db.First(&tier, "name = ?", tierName)

	// Get settings
	var settings map[string]interface{}
	db.Table("organization_settings").Where("organization_id = ?", orgID).First(&settings)

	// Get override if exists
	var override map[string]interface{}
	db.Table("organization_limit_overrides").Where("organization_id = ?", orgID).First(&override)

	subscription := map[string]interface{}{
		"organization_id":    orgID,
		"tier":               tier,
		"subscription_status": org["subscription_status"],
		"trial_start_date":   org["trial_start_date"],
		"trial_end_date":     org["trial_end_date"],
		"settings":           settings,
		"override":           override,
	}

	return utils.SendSimpleSuccess(c, subscription, "Subscription details retrieved successfully")
}

// AdminDeleteOrganization soft deletes an organization
func AdminDeleteOrganization(c *fiber.Ctx) error {
	db := config.DB
	orgID := c.Params("id")

	// Verify org exists
	var existingCount int64
	db.Table("organizations").Where("id = ?", orgID).Count(&existingCount)
	if existingCount == 0 {
		return utils.SendNotFound(c, "Organization not found")
	}

	now := time.Now()

	// Deactivate all members
	db.Table("organization_members").Where("organization_id = ?", orgID).Updates(map[string]interface{}{
		"active":     false,
		"updated_at": now,
	})

	// Soft delete organization
	if err := db.Table("organizations").Where("id = ?", orgID).Updates(map[string]interface{}{
		"active":     false,
		"deleted_at": now,
		"updated_at": now,
	}).Error; err != nil {
		return utils.SendInternalError(c, "Failed to delete organization", err)
	}

	// Audit log
	auditLog := map[string]interface{}{
		"id":              utils.GenerateID(),
		"organization_id": orgID,
		"action":          "organization_deleted",
		"admin_user_id":   c.Locals("userID"),
		"created_at":      now,
	}
	db.Table("admin_audit_logs").Create(auditLog)

	return utils.SendSimpleSuccess(c, nil, "Organization deleted successfully")
}

// AdminResetOrganizationTrial resets the trial for an organization back to a fresh start.
// Accepts optional trial_days (defaults to 14). Clears grace_period_ends_at.
func AdminResetOrganizationTrial(c *fiber.Ctx) error {
	db := config.DB
	orgID := c.Params("id")

	var req struct {
		TrialDays int    `json:"trial_days"`
		Reason    string `json:"reason"`
	}
	_ = c.BodyParser(&req)
	if req.TrialDays <= 0 {
		req.TrialDays = 14
	}

	var existingCount int64
	db.Table("organizations").Where("id = ?", orgID).Count(&existingCount)
	if existingCount == 0 {
		return utils.SendNotFound(c, "Organization not found")
	}

	now := time.Now()
	trialEnd := now.AddDate(0, 0, req.TrialDays)
	performedBy, _ := c.Locals("userID").(string)

	if err := db.Table("organizations").Where("id = ?", orgID).Updates(map[string]interface{}{
		"subscription_status":  "trial",
		"trial_start_date":     now,
		"trial_end_date":       trialEnd,
		"grace_period_ends_at": nil,
		"updated_at":           now,
	}).Error; err != nil {
		return utils.SendInternalError(c, "Failed to reset trial", err)
	}

	// Write to generic audit log
	db.Table("admin_audit_logs").Create(map[string]interface{}{
		"id":              utils.GenerateID(),
		"action":          "organization_trial_reset",
		"admin_user_id":   performedBy,
		"organization_id": orgID,
		"description":     fmt.Sprintf("Trial reset to %d days. Reason: %s", req.TrialDays, req.Reason),
		"created_at":      now,
	})

	// Write to subscription audit log for history tracking
	newStatus := "trial"
	db.Table("subscription_audit_logs").Create(map[string]interface{}{
		"organization_id": orgID,
		"action":          "trial_reset",
		"new_status":      &newStatus,
		"metadata":        map[string]interface{}{"trial_days": req.TrialDays, "reason": req.Reason, "new_trial_start": now.Format(time.RFC3339), "new_trial_end": trialEnd.Format(time.RFC3339)},
		"performed_by":    &performedBy,
		"performed_at":    now,
	})

	return utils.SendSimpleSuccess(c, map[string]interface{}{
		"organization_id": orgID,
		"trial_start":     now,
		"trial_end":       trialEnd,
		"trial_days":      req.TrialDays,
	}, "Trial reset successfully")
}

// AdminExtendOrganizationTrial extends the trial period for an organization.
func AdminExtendOrganizationTrial(c *fiber.Ctx) error {
	db := config.DB
	orgID := c.Params("id")

	var req struct {
		DaysToAdd int    `json:"days_to_add"`
		Reason    string `json:"reason"`
	}
	if err := c.BodyParser(&req); err != nil {
		return utils.SendBadRequest(c, "Invalid request body")
	}
	if req.DaysToAdd <= 0 {
		return utils.SendBadRequest(c, "days_to_add must be a positive number")
	}

	var org map[string]interface{}
	if err := db.Table("organizations").Select("id, trial_end_date").Where("id = ?", orgID).First(&org).Error; err != nil {
		return utils.SendNotFound(c, "Organization not found")
	}

	now := time.Now()
	var newEnd time.Time
	if endDate, ok := org["trial_end_date"]; ok && endDate != nil {
		if t, ok := endDate.(time.Time); ok && t.After(now) {
			newEnd = t.AddDate(0, 0, req.DaysToAdd)
		} else {
			newEnd = now.AddDate(0, 0, req.DaysToAdd)
		}
	} else {
		newEnd = now.AddDate(0, 0, req.DaysToAdd)
	}

	if err := db.Table("organizations").Where("id = ?", orgID).Updates(map[string]interface{}{
		"subscription_status": "trial",
		"trial_end_date":      newEnd,
		"updated_at":          now,
	}).Error; err != nil {
		return utils.SendInternalError(c, "Failed to extend trial", err)
	}

	performedBy, _ := c.Locals("userID").(string)

	db.Table("admin_audit_logs").Create(map[string]interface{}{
		"id":              utils.GenerateID(),
		"action":          "organization_trial_extended",
		"admin_user_id":   performedBy,
		"organization_id": orgID,
		"description":     fmt.Sprintf("Trial extended by %d days. Reason: %s", req.DaysToAdd, req.Reason),
		"created_at":      now,
	})

	// Write to subscription audit log for history tracking
	newStatus := "trial"
	db.Table("subscription_audit_logs").Create(map[string]interface{}{
		"organization_id": orgID,
		"action":          "trial_extended",
		"new_status":      &newStatus,
		"metadata":        map[string]interface{}{"days_added": req.DaysToAdd, "reason": req.Reason, "new_trial_end": newEnd.Format(time.RFC3339)},
		"performed_by":    &performedBy,
		"performed_at":    now,
	})

	return utils.SendSimpleSuccess(c, map[string]interface{}{
		"organization_id": orgID,
		"new_trial_end":   newEnd,
		"days_added":      req.DaysToAdd,
	}, "Trial extended successfully")
}
