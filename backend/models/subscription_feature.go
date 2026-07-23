package models

import "time"

// SubscriptionFeature represents a feature that can be assigned to tiers
type SubscriptionFeature struct {
	ID          string    `gorm:"type:varchar(255);primaryKey" json:"id"`
	Name        string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"name"`
	DisplayName string    `gorm:"type:varchar(255);not null" json:"displayName"`
	Description string    `gorm:"type:text;not null" json:"description"`
	Category    string    `gorm:"type:varchar(100);not null" json:"category"`
	IsActive    bool      `gorm:"type:boolean;not null;default:true" json:"isActive"`
	CreatedAt   time.Time `gorm:"type:timestamp with time zone;not null;default:CURRENT_TIMESTAMP" json:"createdAt"`
}

// TableName specifies the table name for GORM
func (SubscriptionFeature) TableName() string {
	return "subscription_features"
}

// Feature category constants
const (
	FeatureCategoryCore          = "core"
	FeatureCategoryWorkflow      = "workflow"
	FeatureCategorySecurity      = "security"
	FeatureCategoryIntegration   = "integration"
	FeatureCategoryAnalytics     = "analytics"
	FeatureCategorySupport       = "support"
	FeatureCategoryCustomization = "customization"
)

// Feature name constants (for code reference)
// Note: Some constants are defined in subscription.go for backward compatibility
const (
	// Core Features (STARTER)
	FeatureDocumentManagement   = "document_management"
	FeatureBasicWorkflows       = "basic_workflows"
	FeatureInAppNotifications   = "in_app_notifications"
	FeatureStandardReports      = "standard_reports"
	FeatureUserManagement       = "user_management"
	FeatureDepartmentManagement = "department_management"
	FeatureVendorManagement     = "vendor_management"
	FeatureBudgetTracking       = "budget_tracking"
	FeatureMobileWebAccess      = "mobile_web_access"
	FeatureEmailSupport         = "email_support"

	// PRO Features
	FeatureAdvancedWorkflows  = "advanced_workflows"
	FeatureEmailNotifications = "email_notifications"
	// FeatureCustomRoles defined in subscription.go
	// FeatureAdvancedAnalytics defined in subscription.go
	FeatureDataExport = "data_export"
	// FeaturePrioritySupport defined in subscription.go
	FeatureAuditLogs90Days   = "audit_logs_90_days"
	FeatureMultiCurrency     = "multi_currency"
	FeatureAdvancedReporting = "advanced_reporting"
	FeatureWorkflowTemplates = "workflow_templates"

	// CUSTOM Features
	FeatureWebhooks = "webhooks"
	FeatureCustomFields            = "custom_fields"
	FeatureBulkOperations          = "bulk_operations"
	// FeatureAPIAccess defined in subscription.go
	FeatureDedicatedSupportManager = "dedicated_support_manager"
	// FeatureSLAGuarantees defined in subscription.go
	FeatureAuditLogsUnlimited   = "audit_logs_unlimited"
	FeatureCustomDevelopment    = "custom_development"
	FeatureProfessionalServices = "professional_services"
	FeatureDedicatedSupport     = "dedicated_support"
	FeatureCustomTraining       = "custom_training"

	// Future Features (CUSTOM - not yet implemented)
	FeatureWhiteLabeling = "white_labeling"
	// FeatureCustomIntegrations defined in subscription.go
	FeatureOnPremiseDeployment = "on_premise_deployment"
	FeatureAdvancedCompliance  = "advanced_compliance"
	// FeatureSSO defined in subscription.go
	FeatureAdvancedSecurity = "advanced_security"
)

// ============================================================================
// REQUEST/RESPONSE MODELS
// ============================================================================

// CreateFeatureRequest represents a request to create a new feature
type CreateFeatureRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	DisplayName string `json:"displayName" validate:"required,min=2,max=100"`
	Description string `json:"description" validate:"required,min=10,max=500"`
	Category    string `json:"category" validate:"required,oneof=core workflow security integration analytics support customization"`
	IsActive    bool   `json:"isActive"`
}

// UpdateFeatureRequest represents a request to update an existing feature
type UpdateFeatureRequest struct {
	DisplayName *string `json:"displayName,omitempty" validate:"omitempty,min=2,max=100"`
	Description *string `json:"description,omitempty" validate:"omitempty,min=10,max=500"`
	Category    *string `json:"category,omitempty" validate:"omitempty,oneof=core workflow security integration analytics support customization"`
	IsActive    *bool   `json:"isActive,omitempty"`
}

// FeatureResponse represents a feature response with computed fields
type FeatureResponse struct {
	SubscriptionFeature
	TierCount int      `json:"tierCount,omitempty"`
	Tiers     []string `json:"tiers,omitempty"`
}

// FeaturesListResponse represents a list of features
type FeaturesListResponse struct {
	Features []FeatureResponse `json:"features"`
	Total    int               `json:"total"`
}

// FeaturesByCategory represents features grouped by category
type FeaturesByCategory struct {
	Category string                `json:"category"`
	Features []SubscriptionFeature `json:"features"`
}
