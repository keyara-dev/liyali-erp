package models

import (
	"encoding/json"
	"time"

	"gorm.io/datatypes"
)

// SubscriptionTier represents a subscription tier with limits and features
type SubscriptionTier struct {
	ID             string         `gorm:"column:id;type:varchar(255);primaryKey" json:"id"`
	Name           string         `gorm:"column:name;type:varchar(255);uniqueIndex;not null" json:"name"`
	DisplayName    string         `gorm:"column:display_name;type:varchar(255);not null" json:"displayName"`
	Description    string         `gorm:"column:description;type:text;not null" json:"description"`
	PriceMonthly   float64        `gorm:"column:price_monthly;type:numeric(10,2);not null;default:0" json:"priceMonthly"`
	PriceYearly    float64        `gorm:"column:price_yearly;type:numeric(10,2);not null;default:0" json:"priceYearly"`
	MaxWorkspaces  int            `gorm:"column:max_workspaces;type:integer;not null;default:1" json:"maxWorkspaces"`
	MaxTeamMembers int            `gorm:"column:max_users;type:integer;not null;default:1" json:"maxTeamMembers"`
	MaxDocuments   int            `gorm:"column:max_documents;type:integer;not null;default:100" json:"maxDocuments"`
	MaxWorkflows   int            `gorm:"column:max_workflows;type:integer;not null;default:1" json:"maxWorkflows"`
	MaxCustomRoles int            `gorm:"column:max_custom_roles;type:integer;not null;default:0" json:"maxCustomRoles"`
	Features       datatypes.JSON `gorm:"column:features;type:jsonb;not null;default:'[]'" json:"features"`
	IsActive       bool           `gorm:"column:is_active;type:boolean;not null;default:true" json:"isActive"`
	SortOrder      int            `gorm:"column:sort_order;type:integer;not null;default:0" json:"sortOrder"`
	CreatedAt      time.Time      `gorm:"column:created_at;type:timestamp with time zone;not null;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt      time.Time      `gorm:"column:updated_at;type:timestamp with time zone;not null;default:CURRENT_TIMESTAMP" json:"updatedAt"`
}

// TableName specifies the table name for GORM
func (SubscriptionTier) TableName() string {
	return "subscription_tiers"
}

// GetFeatureList returns the list of features as strings
func (t *SubscriptionTier) GetFeatureList() ([]string, error) {
	var features []string
	if err := json.Unmarshal(t.Features, &features); err != nil {
		return nil, err
	}
	return features, nil
}

// HasFeature checks if the tier includes a specific feature
func (t *SubscriptionTier) HasFeature(featureName string) bool {
	features, err := t.GetFeatureList()
	if err != nil {
		return false
	}

	for _, f := range features {
		if f == featureName {
			return true
		}
	}
	return false
}

// IsUnlimited checks if a specific limit is unlimited (-1)
func (t *SubscriptionTier) IsUnlimited(limitType string) bool {
	switch limitType {
	case "workspaces":
		return t.MaxWorkspaces == -1
	case "team_members":
		return t.MaxTeamMembers == -1
	case "documents":
		return t.MaxDocuments == -1
	case "workflows":
		return t.MaxWorkflows == -1
	case "custom_roles":
		return t.MaxCustomRoles == -1
	default:
		return false
	}
}

// GetLimit returns the limit value for a specific resource type
func (t *SubscriptionTier) GetLimit(limitType string) int {
	switch limitType {
	case "workspaces":
		return t.MaxWorkspaces
	case "team_members":
		return t.MaxTeamMembers
	case "documents":
		return t.MaxDocuments
	case "workflows":
		return t.MaxWorkflows
	case "custom_roles":
		return t.MaxCustomRoles
	default:
		return 0
	}
}

// Tier name constants
const (
	TierStarter = "starter"
	TierPro     = "pro"
	TierCustom  = "custom"
)

// Tier ID constants
const (
	TierIDStarter = "tier-starter"
	TierIDPro     = "tier-pro"
	TierIDCustom  = "tier-custom"
)

// Unlimited constant
const UnlimitedLimit = -1

// ============================================================================
// REQUEST/RESPONSE MODELS
// ============================================================================

// CreateTierRequest represents a request to create a new tier
type CreateTierRequest struct {
	Name           string   `json:"name" validate:"required,min=2,max=50"`
	DisplayName    string   `json:"displayName" validate:"required,min=2,max=100"`
	Description    string   `json:"description" validate:"required,min=10,max=500"`
	PriceMonthly   float64  `json:"priceMonthly" validate:"min=0"`
	PriceYearly    float64  `json:"priceYearly" validate:"min=0"`
	MaxWorkspaces  int      `json:"maxWorkspaces" validate:"required,min=-1"`
	MaxTeamMembers int      `json:"maxTeamMembers" validate:"required,min=-1"`
	MaxDocuments   int      `json:"maxDocuments" validate:"required,min=-1"`
	MaxWorkflows   int      `json:"maxWorkflows" validate:"required,min=-1"`
	MaxCustomRoles int      `json:"maxCustomRoles" validate:"required,min=-1"`
	Features       []string `json:"features" validate:"required,min=1"`
	IsActive       bool     `json:"isActive"`
	SortOrder      int      `json:"sortOrder" validate:"min=0"`
}

// UpdateTierRequest represents a request to update an existing tier
type UpdateTierRequest struct {
	DisplayName    *string   `json:"displayName,omitempty" validate:"omitempty,min=2,max=100"`
	Description    *string   `json:"description,omitempty" validate:"omitempty,min=10,max=500"`
	PriceMonthly   *float64  `json:"priceMonthly,omitempty" validate:"omitempty,min=0"`
	PriceYearly    *float64  `json:"priceYearly,omitempty" validate:"omitempty,min=0"`
	MaxWorkspaces  *int      `json:"maxWorkspaces,omitempty" validate:"omitempty,min=-1"`
	MaxTeamMembers *int      `json:"maxTeamMembers,omitempty" validate:"omitempty,min=-1"`
	MaxDocuments   *int      `json:"maxDocuments,omitempty" validate:"omitempty,min=-1"`
	MaxWorkflows   *int      `json:"maxWorkflows,omitempty" validate:"omitempty,min=-1"`
	MaxCustomRoles *int      `json:"maxCustomRoles,omitempty" validate:"omitempty,min=-1"`
	Features       *[]string `json:"features,omitempty" validate:"omitempty,min=1"`
	IsActive       *bool     `json:"isActive,omitempty"`
	SortOrder      *int      `json:"sortOrder,omitempty" validate:"omitempty,min=0"`
}

// ChangeTierRequest represents a request to change an organization's tier
type ChangeTierRequest struct {
	NewTier string `json:"newTier" validate:"required,oneof=starter pro custom"`
	Reason  string `json:"reason" validate:"required,min=10"`
}

// TierResponse represents a tier response with computed fields
type TierResponse struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	DisplayName       string    `json:"displayName"`
	Description       string    `json:"description"`
	PriceMonthly      float64   `json:"priceMonthly"`
	PriceYearly       float64   `json:"priceYearly"`
	MaxWorkspaces     int       `json:"maxWorkspaces"`
	MaxTeamMembers    int       `json:"maxTeamMembers"`
	MaxDocuments      int       `json:"maxDocuments"`
	MaxWorkflows      int       `json:"maxWorkflows"`
	MaxCustomRoles    int       `json:"maxCustomRoles"`
	Features          []string  `json:"features"`
	IsActive          bool      `json:"isActive"`
	SortOrder         int       `json:"sortOrder"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
	FeatureCount      int       `json:"featureCount"`
	OrganizationCount int       `json:"organizationCount,omitempty"`
}

// TiersListResponse represents a list of tiers
type TiersListResponse struct {
	Tiers []TierResponse `json:"tiers"`
	Total int            `json:"total"`
}
