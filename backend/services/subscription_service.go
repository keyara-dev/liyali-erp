package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/liyali/liyali-gateway/logging"
)

// SubscriptionService handles subscription-related business logic
type SubscriptionService struct {
	db     *pgxpool.Pool
	logger *logging.Logger
}

// NewSubscriptionService creates a new subscription service
func NewSubscriptionService(db *pgxpool.Pool, logger *logging.Logger) *SubscriptionService {
	return &SubscriptionService{
		db:     db,
		logger: logger,
	}
}

// FeatureDetail represents full feature information
type FeatureDetail struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

// SubscriptionPlan represents a subscription plan
type SubscriptionPlan struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Slug           string                 `json:"slug"`
	DisplayName    string                 `json:"displayName"`
	Description    string                 `json:"description"`
	PriceMonthly   float64                `json:"priceMonthly"`
	PriceYearly    float64                `json:"priceYearly"`
	MaxWorkspaces  int32                  `json:"maxWorkspaces"`
	MaxTeamMembers int32                  `json:"maxTeamMembers"`
	MaxDocuments   int32                  `json:"maxDocuments"`
	MaxWorkflows   int32                  `json:"maxWorkflows"`
	MaxCustomRoles int32                  `json:"maxCustomRoles"`
	Features       []string               `json:"features"`
	FeatureDetails []FeatureDetail        `json:"featureDetails"`
	IsActive       bool                   `json:"isActive"`
	SortOrder      int32                  `json:"sortOrder"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt      time.Time              `json:"createdAt"`
	UpdatedAt      time.Time              `json:"updatedAt"`
}

// OrganizationTrialStatus represents trial status information
type OrganizationTrialStatus struct {
	OrganizationID     string    `json:"organizationId"`
	SubscriptionStatus string    `json:"subscriptionStatus"`
	TrialStartDate     *time.Time `json:"trialStartDate,omitempty"`
	TrialEndDate       *time.Time `json:"trialEndDate,omitempty"`
	GracePeriodEndsAt  *time.Time `json:"gracePeriodEndsAt,omitempty"`
	PlanSlug           string    `json:"planSlug"`
	PlanName           string    `json:"planName"`
	DaysRemaining      int       `json:"daysRemaining"`
	IsExpired          bool      `json:"isExpired"`
	IsActive           bool      `json:"isActive"`
	InGracePeriod      bool      `json:"inGracePeriod"`
}

// FeatureAccessResult represents feature access check result
type FeatureAccessResult struct {
	Feature   string `json:"feature"`
	HasAccess bool   `json:"hasAccess"`
}

// GetAllSubscriptionPlans retrieves all active subscription plans from database
// Now queries the NEW subscription_tiers table with full feature details
func (s *SubscriptionService) GetAllSubscriptionPlans() ([]SubscriptionPlan, error) {
	// Create a simple logger for service operations
	logger := &logging.Logger{}
	
	logger.Info("Retrieving subscription plans from subscription_tiers table")

	// First, get the basic tier information
	query := `
		SELECT
			id,
			name,
			display_name,
			description,
			price_monthly,
			price_yearly,
			max_workspaces,
			max_team_members,
			max_documents,
			max_workflows,
			max_custom_roles,
			features,
			is_active,
			sort_order,
			created_at,
			updated_at
		FROM subscription_tiers
		WHERE is_active = true
		ORDER BY sort_order ASC
	`

	rows, err := s.db.Query(context.Background(), query)
	if err != nil {
		logger.Error("Failed to query subscription tiers: " + err.Error())
		return nil, fmt.Errorf("failed to query subscription tiers: %w", err)
	}
	defer rows.Close()

	var plans []SubscriptionPlan
	for rows.Next() {
		var plan SubscriptionPlan
		var featuresJSON []byte

		err := rows.Scan(
			&plan.ID,
			&plan.Name,
			&plan.DisplayName,
			&plan.Description,
			&plan.PriceMonthly,
			&plan.PriceYearly,
			&plan.MaxWorkspaces,
			&plan.MaxTeamMembers,
			&plan.MaxDocuments,
			&plan.MaxWorkflows,
			&plan.MaxCustomRoles,
			&featuresJSON,
			&plan.IsActive,
			&plan.SortOrder,
			&plan.CreatedAt,
			&plan.UpdatedAt,
		)
		if err != nil {
			logger.Error("Failed to scan subscription tier: " + err.Error())
			return nil, fmt.Errorf("failed to scan subscription tier: %w", err)
		}

		// Parse feature names array
		if len(featuresJSON) > 0 {
			if err := json.Unmarshal(featuresJSON, &plan.Features); err != nil {
				logger.Error("Failed to parse tier features JSON: " + err.Error())
				// Don't fail, just use empty array
				plan.Features = []string{}
			}
		} else {
			plan.Features = []string{}
		}

		// Initialize empty feature details array
		plan.FeatureDetails = []FeatureDetail{}

		// Try to fetch feature details if we have feature names
		if len(plan.Features) > 0 {
			featureQuery := `
				SELECT id, name, display_name, description, category
				FROM subscription_features
				WHERE name = ANY($1)
				ORDER BY category, display_name
			`
			featureRows, err := s.db.Query(context.Background(), featureQuery, plan.Features)
			if err == nil {
				defer featureRows.Close()
				for featureRows.Next() {
					var detail FeatureDetail
					if err := featureRows.Scan(&detail.ID, &detail.Name, &detail.DisplayName, &detail.Description, &detail.Category); err == nil {
						plan.FeatureDetails = append(plan.FeatureDetails, detail)
					}
				}
			}
		}

		// Convert tier name to uppercase slug format for backward compatibility
		// starter -> STARTER_PLAN, pro -> PRO_PLAN, custom -> ENTERPRISE
		switch plan.Name {
		case "starter":
			plan.Slug = "STARTER_PLAN"
		case "pro":
			plan.Slug = "PRO_PLAN"
		case "enterprise":
			plan.Slug = "ENTERPRISE"
		default:
			plan.Slug = plan.Name
		}

		plans = append(plans, plan)
	}

	if err := rows.Err(); err != nil {
		logger.Error("Error iterating subscription tiers: " + err.Error())
		return nil, fmt.Errorf("error iterating subscription tiers: %w", err)
	}

	logger.Info("Retrieved subscription plans from subscription_tiers table")

	return plans, nil
}

// GetOrganizationTrialStatus retrieves trial status for an organization
func (s *SubscriptionService) GetOrganizationTrialStatus(organizationID string) (*OrganizationTrialStatus, error) {
	logger := &logging.Logger{}
	
	logger.Info("Getting organization trial status")

	// Query organization subscription details from the view
	query := `
		SELECT 
			organization_id,
			organization_name,
			subscription_status,
			trial_start_date,
			trial_end_date,
			grace_period_ends_at,
			plan_name,
			plan_slug,
			trial_days_remaining,
			trial_expired,
			in_grace_period
		FROM organization_subscription_details 
		WHERE organization_id = $1
	`

	var status OrganizationTrialStatus
	var orgName string
	var trialStart, trialEnd, gracePeriodEnd sql.NullTime

	err := s.db.QueryRow(context.Background(), query, organizationID).Scan(
		&status.OrganizationID,
		&orgName,
		&status.SubscriptionStatus,
		&trialStart,
		&trialEnd,
		&gracePeriodEnd,
		&status.PlanName,
		&status.PlanSlug,
		&status.DaysRemaining,
		&status.IsExpired,
		&status.InGracePeriod,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			logger.Warn("Organization not found")
			return nil, fmt.Errorf("organization not found")
		}
		logger.Error("Failed to get organization trial status")
		return nil, fmt.Errorf("failed to get trial status: %w", err)
	}

	// Set optional time fields
	if trialStart.Valid {
		status.TrialStartDate = &trialStart.Time
	}
	if trialEnd.Valid {
		status.TrialEndDate = &trialEnd.Time
	}
	if gracePeriodEnd.Valid {
		status.GracePeriodEndsAt = &gracePeriodEnd.Time
	}

	// Check the subscription_tier — pro/custom overrides trial state entirely
	var orgTier string
	tierQuery := `SELECT COALESCE(subscription_tier, 'starter') FROM organizations WHERE id = $1`
	_ = s.db.QueryRow(context.Background(), tierQuery, organizationID).Scan(&orgTier)

	if orgTier == "pro" || orgTier == "enterprise" {
		status.IsExpired = false
		status.InGracePeriod = false
		status.IsActive = true
		if status.SubscriptionStatus != "active" {
			status.SubscriptionStatus = "active"
		}
		logger.Info("Retrieved organization trial status")
		return &status, nil
	}

	// Determine if trial is active
	now := time.Now()
	if status.SubscriptionStatus == "trial" && status.TrialEndDate != nil {
		status.IsActive = now.Before(*status.TrialEndDate)
	} else {
		status.IsActive = status.SubscriptionStatus == "active"
	}

	logger.Info("Retrieved organization trial status")

	return &status, nil
}

// CheckFeatureAccess checks if an organization has access to a specific feature
func (s *SubscriptionService) CheckFeatureAccess(organizationID, featureName string) (*FeatureAccessResult, error) {
	logger := &logging.Logger{}
	
	logger.Info("Checking feature access")

	// Use the stored function to check feature access
	query := `SELECT organization_has_feature($1, $2)`
	
	var hasAccess bool
	err := s.db.QueryRow(context.Background(), query, organizationID, featureName).Scan(&hasAccess)
	if err != nil {
		logger.Error("Failed to check feature access")
		return nil, fmt.Errorf("failed to check feature access: %w", err)
	}

	result := &FeatureAccessResult{
		Feature:   featureName,
		HasAccess: hasAccess,
	}

	logger.Info("Feature access checked")

	return result, nil
}

// UpgradeOrganization upgrades an organization to a paid tier and clears the trial
func (s *SubscriptionService) UpgradeOrganization(organizationID string, request map[string]interface{}) (map[string]interface{}, error) {
	logger := &logging.Logger{}

	logger.Info("Processing organization upgrade")

	targetPlanSlug, _ := request["targetPlanSlug"].(string)

	// Map plan slug to subscription_tier value
	tierMap := map[string]string{
		"PRO_PLAN":    "pro",
		"ENTERPRISE":  "enterprise",
		"STARTER_PLAN": "starter",
	}
	newTier, ok := tierMap[targetPlanSlug]
	if !ok {
		newTier = targetPlanSlug
	}

	// Upgrade the org: set tier + clear trial for paid plans
	updates := `UPDATE organizations SET subscription_tier = $2, updated_at = CURRENT_TIMESTAMP`
	if newTier == "pro" || newTier == "enterprise" {
		updates += `, subscription_status = 'active', trial_end_date = NULL, grace_period_ends_at = NULL`
	}
	updates += ` WHERE id = $1`

	_, err := s.db.Exec(context.Background(), updates, organizationID, newTier)
	if err != nil {
		logger.Error("Failed to upgrade organization: " + err.Error())
		return nil, fmt.Errorf("failed to upgrade organization: %w", err)
	}

	// Audit log
	metadata, _ := json.Marshal(map[string]interface{}{
		"target_plan_slug": targetPlanSlug,
		"new_tier":         newTier,
	})
	_, _ = s.db.Exec(context.Background(),
		`INSERT INTO subscription_audit_logs (organization_id, action, metadata) VALUES ($1, $2, $3)`,
		organizationID, "upgraded", metadata,
	)

	logger.Info("Organization upgraded successfully")

	return map[string]interface{}{
		"organizationId": organizationID,
		"newTier":        newTier,
		"status":         "active",
		"timestamp":      time.Now(),
	}, nil
}

// ExtendOrganizationTrial extends the trial period for an organization (admin only)
func (s *SubscriptionService) ExtendOrganizationTrial(organizationID string, daysToAdd int, reason string, performedBy string) error {
	logger := &logging.Logger{}
	
	logger.Info("Extending organization trial")

	// Update the organization's grace period
	query := `
		UPDATE organizations 
		SET grace_period_ends_at = COALESCE(grace_period_ends_at, trial_end_date, CURRENT_TIMESTAMP) + INTERVAL '%d days',
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`
	
	_, err := s.db.Exec(context.Background(), fmt.Sprintf(query, daysToAdd), organizationID)
	if err != nil {
		logger.Error("Failed to extend organization trial")
		return fmt.Errorf("failed to extend trial: %w", err)
	}

	// Create audit log entry
	auditQuery := `
		INSERT INTO subscription_audit_logs (
			organization_id, action, metadata, performed_by
		) VALUES ($1, $2, $3, $4)
	`
	
	metadata := map[string]interface{}{
		"days_added": daysToAdd,
		"reason":     reason,
		"action_type": "trial_extension",
	}
	metadataJSON, _ := json.Marshal(metadata)
	
	_, err = s.db.Exec(context.Background(), auditQuery, organizationID, "trial_extended", metadataJSON, performedBy)
	if err != nil {
		logger.Warn("Failed to create audit log for trial extension")
		// Don't fail the operation if audit logging fails
	}

	logger.Info("Organization trial extended successfully")

	return nil
}

// ResetOrganizationTrial resets the trial period for an organization (admin only)
func (s *SubscriptionService) ResetOrganizationTrial(organizationID string, trialDays int, reason string, performedBy string) error {
	logger := &logging.Logger{}
	
	logger.Info("Resetting organization trial")

	// Calculate new trial dates
	now := time.Now()
	trialStart := now
	trialEnd := now.AddDate(0, 0, trialDays)

	// Update the organization with new trial dates and reset status
	query := `
		UPDATE organizations 
		SET trial_start_date = $2,
		    trial_end_date = $3,
		    subscription_status = 'trial',
		    grace_period_ends_at = NULL,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`
	
	_, err := s.db.Exec(context.Background(), query, organizationID, trialStart, trialEnd)
	if err != nil {
		logger.Error("Failed to reset organization trial")
		return fmt.Errorf("failed to reset trial: %w", err)
	}

	// Create audit log entry
	auditQuery := `
		INSERT INTO subscription_audit_logs (
			organization_id, action, metadata, performed_by
		) VALUES ($1, $2, $3, $4)
	`
	
	metadata := map[string]interface{}{
		"trial_days": trialDays,
		"reason":     reason,
		"action_type": "trial_reset",
		"new_trial_start": trialStart.Format(time.RFC3339),
		"new_trial_end": trialEnd.Format(time.RFC3339),
	}
	metadataJSON, _ := json.Marshal(metadata)
	
	_, err = s.db.Exec(context.Background(), auditQuery, organizationID, "trial_reset", metadataJSON, performedBy)
	if err != nil {
		logger.Warn("Failed to create audit log for trial reset")
		// Don't fail the operation if audit logging fails
	}

	logger.Info("Organization trial reset successfully")

	return nil
}

// GetOrganizationSubscriptionDetails retrieves comprehensive subscription details
func (s *SubscriptionService) GetOrganizationSubscriptionDetails(organizationID string) (map[string]interface{}, error) {
	logger := &logging.Logger{}
	
	logger.Info("Getting organization subscription details")

	// Get trial status
	trialStatus, err := s.GetOrganizationTrialStatus(organizationID)
	if err != nil {
		return nil, err
	}

	// Get subscription plans
	plans, err := s.GetAllSubscriptionPlans()
	if err != nil {
		return nil, err
	}

	// Find current plan
	var currentPlan *SubscriptionPlan
	for _, plan := range plans {
		if plan.Slug == trialStatus.PlanSlug {
			currentPlan = &plan
			break
		}
	}

	// Get user count for the organization
	userCountQuery := `
		SELECT COUNT(*) 
		FROM organization_members 
		WHERE organization_id = $1 AND active = true
	`
	
	var currentUserCount int
	err = s.db.QueryRow(context.Background(), userCountQuery, organizationID).Scan(&currentUserCount)
	if err != nil {
		logger.Warn("Failed to get user count")
		currentUserCount = 0
	}

	// Build response
	response := map[string]interface{}{
		"trialStatus": trialStatus,
		"plan":        currentPlan,
		"planLimits": map[string]interface{}{
			"organizationId":    organizationID,
			"maxUsersAllowed":   currentPlan.MaxTeamMembers,
			"planMaxUsers":      currentPlan.MaxTeamMembers,
			"planMetadata":      currentPlan.Metadata,
			"currentUserCount":  currentUserCount,
			"canAddUsers":       currentPlan.MaxTeamMembers == -1 || currentUserCount < int(currentPlan.MaxTeamMembers),
		},
	}

	logger.Info("Retrieved organization subscription details")

	return response, nil
}