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
	Tier   string `gorm:"default:starter" json:"tier"` // starter, pro, enterprise

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
	Role         string  `gorm:"not null" json:"role"` // admin, manager, approver, requester, viewer
	RoleID       string  `gorm:"-" json:"roleId,omitempty"` // Computed field for role ID
	RoleName     string  `gorm:"-" json:"roleName,omitempty"` // Computed field for role name
	Department   string  `json:"department,omitempty"`
	DepartmentID *string `gorm:"index" json:"departmentId,omitempty"` // Foreign key to OrganizationDepartment
	Title        string  `json:"title,omitempty"`

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
	ManagerName string  `json:"manager_name,omitempty"`
	ParentID    *string `json:"parentId,omitempty"`

	Active    bool      `gorm:"column:is_active;default:true" json:"active"`
	IsActive  bool      `gorm:"column:is_active;default:true" json:"is_active"`
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

// Note: OrganizationRole, OrganizationPermission, and PermissionAssignment 
// have been moved to enhanced_auth.go for the new RBAC system
