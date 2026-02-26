package models

import (
	"time"

	"gorm.io/datatypes"
)

// OrganizationLimitOverride represents custom limits for specific organizations
type OrganizationLimitOverride struct {
	ID             string         `gorm:"type:varchar(255);primaryKey" json:"id"`
	OrganizationID string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"organizationId"`
	MaxWorkspaces  *int           `gorm:"type:integer" json:"maxWorkspaces,omitempty"`
	MaxTeamMembers *int           `gorm:"type:integer" json:"maxTeamMembers,omitempty"`
	MaxDocuments   *int           `gorm:"type:integer" json:"maxDocuments,omitempty"`
	MaxWorkflows   *int           `gorm:"type:integer" json:"maxWorkflows,omitempty"`
	MaxCustomRoles     *int           `gorm:"type:integer" json:"maxCustomRoles,omitempty"`
	MaxRequisitions    *int           `gorm:"type:integer" json:"maxRequisitions,omitempty"`
	MaxBudgets         *int           `gorm:"type:integer" json:"maxBudgets,omitempty"`
	MaxPurchaseOrders  *int           `gorm:"type:integer" json:"maxPurchaseOrders,omitempty"`
	MaxPaymentVouchers *int           `gorm:"type:integer" json:"maxPaymentVouchers,omitempty"`
	MaxGRNs            *int           `gorm:"type:integer" json:"maxGRNs,omitempty"`
	MaxDepartments     *int           `gorm:"type:integer" json:"maxDepartments,omitempty"`
	MaxVendors         *int           `gorm:"type:integer" json:"maxVendors,omitempty"`
	Features           datatypes.JSON `gorm:"type:jsonb" json:"features,omitempty"`
	Reason         string         `gorm:"type:text;not null" json:"reason"`
	AdminUserID    string         `gorm:"type:varchar(255);not null" json:"adminUserId"`
	ExpiresAt      *time.Time     `gorm:"type:timestamp with time zone" json:"expiresAt,omitempty"`
	CreatedAt      time.Time      `gorm:"type:timestamp with time zone;not null;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt      time.Time      `gorm:"type:timestamp with time zone;not null;default:CURRENT_TIMESTAMP" json:"updatedAt"`
}

// TableName specifies the table name for GORM
func (OrganizationLimitOverride) TableName() string {
	return "organization_limit_overrides"
}

// IsExpired checks if the override has expired
func (o *OrganizationLimitOverride) IsExpired() bool {
	if o.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*o.ExpiresAt)
}

// IsActive checks if the override is currently active
func (o *OrganizationLimitOverride) IsActive() bool {
	return !o.IsExpired()
}

// ============================================================================
// REQUEST/RESPONSE MODELS
// ============================================================================

// OverrideLimitsRequest represents a request to override organization limits
type OverrideLimitsRequest struct {
	MaxWorkspaces      *int      `json:"maxWorkspaces,omitempty" validate:"omitempty,min=-1"`
	MaxTeamMembers     *int      `json:"maxTeamMembers,omitempty" validate:"omitempty,min=-1"`
	MaxDocuments       *int      `json:"maxDocuments,omitempty" validate:"omitempty,min=-1"`
	MaxWorkflows       *int      `json:"maxWorkflows,omitempty" validate:"omitempty,min=-1"`
	MaxCustomRoles     *int      `json:"maxCustomRoles,omitempty" validate:"omitempty,min=-1"`
	MaxRequisitions    *int      `json:"maxRequisitions,omitempty" validate:"omitempty,min=-1"`
	MaxBudgets         *int      `json:"maxBudgets,omitempty" validate:"omitempty,min=-1"`
	MaxPurchaseOrders  *int      `json:"maxPurchaseOrders,omitempty" validate:"omitempty,min=-1"`
	MaxPaymentVouchers *int      `json:"maxPaymentVouchers,omitempty" validate:"omitempty,min=-1"`
	MaxGRNs            *int      `json:"maxGRNs,omitempty" validate:"omitempty,min=-1"`
	MaxDepartments     *int      `json:"maxDepartments,omitempty" validate:"omitempty,min=-1"`
	MaxVendors         *int      `json:"maxVendors,omitempty" validate:"omitempty,min=-1"`
	Features           *[]string `json:"features,omitempty"`
	Reason             string    `json:"reason" validate:"required,min=10"`
	ExpiresAt          *string   `json:"expiresAt,omitempty"` // ISO 8601 format
}

// EffectiveLimits represents the effective limits for an organization (tier + overrides)
type EffectiveLimits struct {
	OrganizationID     string `json:"organizationId"`
	TierName           string `json:"tierName"`
	MaxWorkspaces      int    `json:"maxWorkspaces"`
	MaxTeamMembers     int    `json:"maxTeamMembers"`
	MaxDocuments       int    `json:"maxDocuments"`
	MaxWorkflows       int    `json:"maxWorkflows"`
	MaxCustomRoles     int    `json:"maxCustomRoles"`
	MaxRequisitions    int    `json:"maxRequisitions"`
	MaxBudgets         int    `json:"maxBudgets"`
	MaxPurchaseOrders  int    `json:"maxPurchaseOrders"`
	MaxPaymentVouchers int    `json:"maxPaymentVouchers"`
	MaxGRNs            int    `json:"maxGRNs"`
	MaxDepartments     int    `json:"maxDepartments"`
	MaxVendors         int    `json:"maxVendors"`
	HasOverrides       bool   `json:"hasOverrides"`
}

// OrganizationUsage represents current resource usage for an organization
type OrganizationUsage struct {
	OrganizationID         string  `json:"organizationId"`
	CurrentWorkspaces      int     `json:"currentWorkspaces"`
	CurrentTeamMembers     int     `json:"currentTeamMembers"`
	CurrentDocuments       int     `json:"currentDocuments"`
	CurrentWorkflows       int     `json:"currentWorkflows"`
	CurrentCustomRoles     int     `json:"currentCustomRoles"`
	CurrentRequisitions    int     `json:"currentRequisitions"`
	CurrentBudgets         int     `json:"currentBudgets"`
	CurrentPurchaseOrders  int     `json:"currentPurchaseOrders"`
	CurrentPaymentVouchers int     `json:"currentPaymentVouchers"`
	CurrentGRNs            int     `json:"currentGRNs"`
	CurrentDepartments     int     `json:"currentDepartments"`
	CurrentVendors         int     `json:"currentVendors"`
	WorkspacesPercent      float64 `json:"workspacesPercent"`
	TeamMembersPercent     float64 `json:"teamMembersPercent"`
	DocumentsPercent       float64 `json:"documentsPercent"`
	WorkflowsPercent       float64 `json:"workflowsPercent"`
	CustomRolesPercent     float64 `json:"customRolesPercent"`
	RequisitionsPercent    float64 `json:"requisitionsPercent"`
	BudgetsPercent         float64 `json:"budgetsPercent"`
	PurchaseOrdersPercent  float64 `json:"purchaseOrdersPercent"`
	PaymentVouchersPercent float64 `json:"paymentVouchersPercent"`
	GRNsPercent            float64 `json:"grnsPercent"`
	DepartmentsPercent     float64 `json:"departmentsPercent"`
	VendorsPercent         float64 `json:"vendorsPercent"`
}

// LimitsWithUsage combines effective limits with current usage
type LimitsWithUsage struct {
	Limits EffectiveLimits   `json:"limits"`
	Usage  OrganizationUsage `json:"usage"`
}
