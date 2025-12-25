package models

import (
	"time"

	"gorm.io/datatypes"
)

// Organization represents a tenant/workspace
type Organization struct {
	ID          string    `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"not null" json:"name"`
	Slug        string    `gorm:"uniqueIndex;not null" json:"slug"`
	Description string    `json:"description,omitempty"`

	// Branding
	LogoURL      string `json:"logoUrl,omitempty"`
	PrimaryColor string `gorm:"default:#0066CC" json:"primaryColor"`

	// Status
	Active bool   `gorm:"default:true" json:"active"`
	Tier   string `gorm:"default:free" json:"tier"` // free, pro, enterprise

	// Relationships
	CreatedBy string `json:"createdBy"`
	Creator   *User  `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// OrganizationSettings stores per-org configuration
type OrganizationSettings struct {
	ID             string `gorm:"primaryKey" json:"id"`
	OrganizationID string `gorm:"uniqueIndex;not null" json:"organizationId"`
	Organization   *Organization `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`

	// Approval Settings
	RequireDigitalSignatures bool   `gorm:"default:true" json:"requireDigitalSignatures"`
	DefaultApprovalChain     string `json:"defaultApprovalChain,omitempty"`

	// Financial Settings
	Currency                  string  `gorm:"default:USD" json:"currency"`
	FiscalYearStart          int     `gorm:"default:1" json:"fiscalYearStart"`
	EnableBudgetValidation   bool    `gorm:"default:true" json:"enableBudgetValidation"`
	BudgetVarianceThreshold  float64 `gorm:"default:5.00" json:"budgetVarianceThreshold"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// OrganizationMember represents user-organization relationship
type OrganizationMember struct {
	ID             string `gorm:"primaryKey" json:"id"`
	OrganizationID string `gorm:"index;not null" json:"organizationId"`
	Organization   *Organization `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	UserID         string `gorm:"index;not null" json:"userId"`
	User           *User  `gorm:"foreignKey:UserID" json:"user,omitempty"`

	// Membership Details
	Role       string `gorm:"not null" json:"role"` // admin, manager, approver, requester, viewer
	Department string `json:"department,omitempty"`
	Title      string `json:"title,omitempty"`

	// Status
	Active     bool       `gorm:"default:true" json:"active"`
	InvitedAt  *time.Time `json:"invitedAt,omitempty"`
	JoinedAt   *time.Time `json:"joinedAt,omitempty"`
	InvitedBy  *string    `json:"invitedBy,omitempty"`

	// Custom permissions override
	CustomPermissions datatypes.JSON `gorm:"type:jsonb" json:"customPermissions,omitempty"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// OrganizationDepartment represents departments within an organization
type OrganizationDepartment struct {
	ID             string `gorm:"primaryKey" json:"id"`
	OrganizationID string `gorm:"index;not null" json:"organizationId"`
	Organization   *Organization `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`

	Name        string  `gorm:"not null" json:"name"`
	Code        string  `json:"code,omitempty"`
	Description string  `json:"description,omitempty"`
	ParentID    *string `json:"parentId,omitempty"`

	Active    bool      `gorm:"default:true" json:"active"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// TableName methods for GORM
func (Organization) TableName() string {
	return "organizations"
}

func (OrganizationSettings) TableName() string {
	return "organization_settings"
}

func (OrganizationMember) TableName() string {
	return "organization_members"
}

func (OrganizationDepartment) TableName() string {
	return "organization_departments"
}

// OrganizationRole represents a custom role defined by organization admin
// Allows organizations to create their own role definitions beyond the default ones
type OrganizationRole struct {
	ID             string    `gorm:"primaryKey" json:"id"`
	OrganizationID string    `gorm:"index;not null" json:"organizationId"`
	Organization   *Organization `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`

	// Role Details
	Name        string `gorm:"not null" json:"name"` // e.g., "Senior Manager", "Budget Controller"
	Description string `json:"description,omitempty"`

	// Role Classification
	IsDefault bool `gorm:"default:false" json:"isDefault"` // Whether this is a default system role
	IsActive  bool `gorm:"default:true" json:"isActive"`

	// Permissions will be managed through PermissionAssignment table
	// Not stored directly here for normalization and flexibility

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// OrganizationPermission represents an available permission in the organization
// These are the granular permissions that can be assigned to roles
type OrganizationPermission struct {
	ID             string    `gorm:"primaryKey" json:"id"`
	OrganizationID string    `gorm:"index;not null" json:"organizationId"`
	Organization   *Organization `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`

	// Permission Details
	Resource    string `gorm:"not null" json:"resource"` // e.g., "requisition", "budget"
	Action      string `gorm:"not null" json:"action"`   // e.g., "create", "approve"
	Description string `json:"description,omitempty"`

	// Status
	IsActive bool `gorm:"default:true" json:"isActive"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// PermissionAssignment represents the mapping between roles and permissions
// A role can have multiple permissions, and a permission can be assigned to multiple roles
type PermissionAssignment struct {
	ID                  string              `gorm:"primaryKey" json:"id"`
	OrganizationRoleID  string              `gorm:"index;not null" json:"organizationRoleId"`
	OrganizationRole    *OrganizationRole   `gorm:"foreignKey:OrganizationRoleID" json:"organizationRole,omitempty"`
	OrganizationPermissionID string         `gorm:"index;not null" json:"organizationPermissionId"`
	OrganizationPermission  *OrganizationPermission `gorm:"foreignKey:OrganizationPermissionId" json:"organizationPermission,omitempty"`

	// Track assignment metadata
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// TableName methods for GORM
func (OrganizationRole) TableName() string {
	return "organization_roles"
}

func (OrganizationPermission) TableName() string {
	return "organization_permissions"
}

func (PermissionAssignment) TableName() string {
	return "permission_assignments"
}
