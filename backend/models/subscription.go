package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// SubscriptionPlan represents admin-managed subscription plans
type SubscriptionPlan struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name         string         `gorm:"not null" json:"name"`
	Slug         string         `gorm:"uniqueIndex;not null" json:"slug"` // STARTER_PLAN, PRO_PLAN, ENTERPRISE
	Description  string         `json:"description"`
	PriceMonthly float64        `gorm:"default:0.00" json:"priceMonthly"`
	PriceYearly  float64        `gorm:"default:0.00" json:"priceYearly"`
	Features     datatypes.JSON `gorm:"type:jsonb;not null;default:'[]'" json:"features"`
	MaxUsers     int            `gorm:"default:50" json:"maxUsers"`
	IsActive     bool           `gorm:"default:true" json:"isActive"`
	SortOrder    int            `gorm:"default:0" json:"sortOrder"`
	Metadata     datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"metadata"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
}

// FeatureFlag represents feature flags that control access based on subscription plans
type FeatureFlag struct {
	ID                 uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name               string         `gorm:"uniqueIndex;not null" json:"name"`
	Description        string         `json:"description"`
	PlanRequirements   datatypes.JSON `gorm:"type:jsonb;not null;default:'[]'" json:"planRequirements"`
	IsTrialAllowed     bool           `gorm:"default:false" json:"isTrialAllowed"`
	IsEnterpriseOnly   bool           `gorm:"default:false" json:"isEnterpriseOnly"`
	IsActive           bool           `gorm:"default:true" json:"isActive"`
	CreatedAt          time.Time      `json:"createdAt"`
	UpdatedAt          time.Time      `json:"updatedAt"`
}

// OrganizationSubscription represents active subscriptions for organizations
type OrganizationSubscription struct {
	ID                     uuid.UUID         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	OrganizationID         string            `gorm:"uniqueIndex;not null" json:"organizationId"`
	Organization           *Organization     `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	PlanID                 uuid.UUID         `gorm:"type:uuid;not null" json:"planId"`
	Plan                   *SubscriptionPlan `gorm:"foreignKey:PlanID" json:"plan,omitempty"`
	StripeSubscriptionID   *string           `json:"stripeSubscriptionId,omitempty"`
	Status                 string            `gorm:"not null;default:'trial'" json:"status"` // trial, active, past_due, canceled, expired
	CurrentPeriodStart     *time.Time        `json:"currentPeriodStart,omitempty"`
	CurrentPeriodEnd       *time.Time        `json:"currentPeriodEnd,omitempty"`
	CancelAtPeriodEnd      bool              `gorm:"default:false" json:"cancelAtPeriodEnd"`
	PaymentFailedCount     int               `gorm:"default:0" json:"paymentFailedCount"`
	LastPaymentFailedAt    *time.Time        `json:"lastPaymentFailedAt,omitempty"`
	CreatedAt              time.Time         `json:"createdAt"`
	UpdatedAt              time.Time         `json:"updatedAt"`
}

// SubscriptionAuditLog represents audit trail for subscription changes
type SubscriptionAuditLog struct {
	ID             uuid.UUID         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	OrganizationID string            `gorm:"not null" json:"organizationId"`
	Organization   *Organization     `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	Action         string            `gorm:"not null" json:"action"` // trial_started, upgraded, downgraded, payment_failed, etc.
	OldPlanID      *uuid.UUID        `gorm:"type:uuid" json:"oldPlanId,omitempty"`
	OldPlan        *SubscriptionPlan `gorm:"foreignKey:OldPlanID" json:"oldPlan,omitempty"`
	NewPlanID      *uuid.UUID        `gorm:"type:uuid" json:"newPlanId,omitempty"`
	NewPlan        *SubscriptionPlan `gorm:"foreignKey:NewPlanID" json:"newPlan,omitempty"`
	OldStatus      *string           `json:"oldStatus,omitempty"`
	NewStatus      *string           `json:"newStatus,omitempty"`
	Metadata       datatypes.JSON    `gorm:"type:jsonb;default:'{}'" json:"metadata"`
	PerformedBy    *string           `json:"performedBy,omitempty"` // User ID or 'system'
	PerformedAt    time.Time         `gorm:"default:CURRENT_TIMESTAMP" json:"performedAt"`
}

// Subscription status constants
const (
	SubscriptionStatusTrial    = "trial"
	SubscriptionStatusActive   = "active"
	SubscriptionStatusPastDue  = "past_due"
	SubscriptionStatusCanceled = "canceled"
	SubscriptionStatusExpired  = "expired"
)

// Plan slug constants
const (
	PlanSlugStarter    = "STARTER_PLAN"
	PlanSlugPro        = "PRO_PLAN"
	PlanSlugEnterprise = "ENTERPRISE"
)

// ============================================================================
// COMPUTED MODELS (NOT STORED IN DATABASE)
// ============================================================================

// TrialStatus represents the trial status of an organization
type TrialStatus struct {
	OrganizationID     string    `json:"organizationId"`
	SubscriptionStatus string    `json:"subscriptionStatus"`
	TrialStartDate     *time.Time `json:"trialStartDate"`
	TrialEndDate       *time.Time `json:"trialEndDate"`
	GracePeriodEndsAt  *time.Time `json:"gracePeriodEndsAt"`
	PlanSlug           string    `json:"planSlug"`
	PlanName           string    `json:"planName"`
	DaysRemaining      int       `json:"daysRemaining"`
	IsExpired          bool      `json:"isExpired"`
	IsActive           bool      `json:"isActive"`
	InGracePeriod      bool      `json:"inGracePeriod"`
}

// PlanLimits represents the limits and usage for an organization's plan
type PlanLimits struct {
	OrganizationID   string                 `json:"organizationId"`
	MaxUsersAllowed  int                    `json:"maxUsersAllowed"`
	PlanMaxUsers     int                    `json:"planMaxUsers"`
	PlanMetadata     map[string]interface{} `json:"planMetadata"`
	CurrentUserCount int                    `json:"currentUserCount"`
	CanAddUsers      bool                   `json:"canAddUsers"`
}

// SubscriptionAnalytics represents subscription analytics data
type SubscriptionAnalytics struct {
	TrialCount        int     `json:"trialCount"`
	ActiveCount       int     `json:"activeCount"`
	PastDueCount      int     `json:"pastDueCount"`
	CanceledCount     int     `json:"canceledCount"`
	ExpiredCount      int     `json:"expiredCount"`
	ExpiredTrials     int     `json:"expiredTrials"`
	TrialsEndingSoon  int     `json:"trialsEndingSoon"`
	ConversionRate    float64 `json:"conversionRate"`
}

// PlanDistribution represents the distribution of organizations across plans
type PlanDistribution struct {
	PlanName           string  `json:"planName"`
	PlanSlug           string  `json:"planSlug"`
	OrganizationCount  int     `json:"organizationCount"`
	Percentage         float64 `json:"percentage"`
}

// ============================================================================
// REQUEST/RESPONSE MODELS
// ============================================================================

// UpgradeRequest represents a request to upgrade an organization
type UpgradeRequest struct {
	TargetPlanSlug       string  `json:"targetPlanSlug" validate:"required,oneof=PRO_PLAN ENTERPRISE"`
	PaymentMethodID      *string `json:"paymentMethodId,omitempty"`
	BillingCycle         string  `json:"billingCycle" validate:"required,oneof=monthly yearly"`
	PromoCode            *string `json:"promoCode,omitempty"`
}

// DowngradeRequest represents a request to downgrade an organization
type DowngradeRequest struct {
	TargetPlanSlug string `json:"targetPlanSlug" validate:"required,oneof=STARTER_PLAN PRO_PLAN"`
	Reason         string `json:"reason,omitempty"`
	Feedback       string `json:"feedback,omitempty"`
}

// ExtendTrialRequest represents a request to extend trial period (admin only)
type ExtendTrialRequest struct {
	DaysToAdd int    `json:"daysToAdd" validate:"required,min=1,max=30"`
	Reason    string `json:"reason" validate:"required"`
}

// SubscriptionResponse represents a subscription response
type SubscriptionResponse struct {
	Organization       *Organization              `json:"organization"`
	Subscription       *OrganizationSubscription  `json:"subscription"`
	Plan               *SubscriptionPlan          `json:"plan"`
	TrialStatus        *TrialStatus               `json:"trialStatus"`
	AvailableFeatures  []FeatureFlag              `json:"availableFeatures"`
	PlanLimits         *PlanLimits                `json:"planLimits"`
}

// PlansResponse represents the response for getting all plans
type PlansResponse struct {
	Plans []SubscriptionPlan `json:"plans"`
}

// AnalyticsResponse represents subscription analytics response
type AnalyticsResponse struct {
	Analytics        *SubscriptionAnalytics `json:"analytics"`
	PlanDistribution []PlanDistribution     `json:"planDistribution"`
}

// ============================================================================
// WEBHOOK MODELS
// ============================================================================

// StripeWebhookEvent represents a Stripe webhook event
type StripeWebhookEvent struct {
	ID               string                 `json:"id"`
	Type             string                 `json:"type"`
	Data             map[string]interface{} `json:"data"`
	Created          int64                  `json:"created"`
	LiveMode         bool                   `json:"livemode"`
	PendingWebhooks  int                    `json:"pending_webhooks"`
	Request          map[string]interface{} `json:"request"`
}

// PaymentIntentData represents Stripe payment intent data
type PaymentIntentData struct {
	ID                   string                 `json:"id"`
	Amount               int64                  `json:"amount"`
	Currency             string                 `json:"currency"`
	Status               string                 `json:"status"`
	SubscriptionID       *string                `json:"subscription,omitempty"`
	Metadata             map[string]interface{} `json:"metadata"`
}

// SubscriptionData represents Stripe subscription data
type SubscriptionData struct {
	ID                 string                 `json:"id"`
	Status             string                 `json:"status"`
	CustomerID         string                 `json:"customer"`
	CurrentPeriodStart int64                  `json:"current_period_start"`
	CurrentPeriodEnd   int64                  `json:"current_period_end"`
	CancelAtPeriodEnd  bool                   `json:"cancel_at_period_end"`
	Metadata           map[string]interface{} `json:"metadata"`
}

// ============================================================================
// FEATURE ACCESS HELPERS
// ============================================================================

// HasFeature checks if a plan has a specific feature
func (p *SubscriptionPlan) HasFeature(featureName string) bool {
	if p.Metadata == nil {
		return false
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal(p.Metadata, &metadata); err != nil {
		return false
	}

	if value, exists := metadata[featureName]; exists {
		if boolValue, ok := value.(bool); ok {
			return boolValue
		}
	}

	return false
}

// GetFeatureList returns the list of features as strings
func (p *SubscriptionPlan) GetFeatureList() []string {
	var features []string
	if err := json.Unmarshal(p.Features, &features); err != nil {
		return []string{}
	}
	return features
}

// IsUnlimited checks if the plan has unlimited users
func (p *SubscriptionPlan) IsUnlimited() bool {
	return p.MaxUsers == -1
}

// ============================================================================
// TRIAL STATUS HELPERS
// ============================================================================

// IsTrialActive checks if the trial is currently active
func (t *TrialStatus) IsTrialActive() bool {
	return t.SubscriptionStatus == SubscriptionStatusTrial && 
		   t.TrialEndDate != nil && 
		   time.Now().Before(*t.TrialEndDate)
}

// IsTrialExpired checks if the trial has expired
func (t *TrialStatus) IsTrialExpired() bool {
	return t.SubscriptionStatus == SubscriptionStatusTrial && 
		   t.TrialEndDate != nil && 
		   time.Now().After(*t.TrialEndDate)
}

// IsInGracePeriod checks if the organization is in grace period
func (t *TrialStatus) IsInGracePeriod() bool {
	return t.GracePeriodEndsAt != nil && time.Now().Before(*t.GracePeriodEndsAt)
}

// GetTrialDaysRemaining returns the number of days remaining in trial
func (t *TrialStatus) GetTrialDaysRemaining() int {
	if t.TrialEndDate == nil || !t.IsTrialActive() {
		return 0
	}
	
	duration := t.TrialEndDate.Sub(time.Now())
	days := int(duration.Hours() / 24)
	if days < 0 {
		return 0
	}
	return days
}

// GetGracePeriodDaysRemaining returns the number of days remaining in grace period
func (t *TrialStatus) GetGracePeriodDaysRemaining() int {
	if !t.IsInGracePeriod() {
		return 0
	}
	
	duration := t.GracePeriodEndsAt.Sub(time.Now())
	days := int(duration.Hours() / 24)
	if days < 0 {
		return 0
	}
	return days
}
const (
	FeatureCustomRoles          = "custom_roles"
	FeatureOfflineCapabilities  = "offline_capabilities"
	FeatureAPIAccess            = "api_access"
	FeaturePrioritySupport      = "priority_support"
	FeatureDedicatedInstance    = "dedicated_instance"
	FeatureCustomIntegrations   = "custom_integrations"
	FeatureSLAGuarantees        = "sla_guarantees"
	FeatureAdvancedAnalytics    = "advanced_analytics"
	FeatureUnlimitedUsers       = "unlimited_users"
	FeatureDedicatedSuccessManager = "dedicated_success_manager"
)

// Audit action constants
const (
	AuditActionTrialStarted     = "trial_started"
	AuditActionTrialExtended    = "trial_extended"
	AuditActionTrialReset       = "trial_reset"
	AuditActionTrialExpired     = "trial_expired"
	AuditActionUpgraded         = "upgraded"
	AuditActionDowngraded       = "downgraded"
	AuditActionPaymentFailed    = "payment_failed"
	AuditActionPaymentSucceeded = "payment_succeeded"
	AuditActionCanceled         = "canceled"
	AuditActionReactivated      = "reactivated"
	AuditActionGracePeriodSet   = "grace_period_set"
	AuditActionGracePeriodEnded = "grace_period_ended"
)

// TableName overrides for GORM
func (SubscriptionPlan) TableName() string {
	return "subscription_plans"
}

func (FeatureFlag) TableName() string {
	return "feature_flags"
}

func (OrganizationSubscription) TableName() string {
	return "organization_subscriptions"
}

func (SubscriptionAuditLog) TableName() string {
	return "subscription_audit_logs"
}