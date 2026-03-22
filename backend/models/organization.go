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
	Tagline      string `json:"tagline,omitempty"`
	PrimaryColor string `gorm:"default:#0066CC" json:"primaryColor"`

	// Status
	Active bool   `gorm:"default:true" json:"active"`
	Tier   string `gorm:"column:subscription_tier;default:starter" json:"tier"` // starter, pro, custom

	// Subscription & Trial (columns exist in DB; must be declared here to avoid GORM silent drops)
	SubscriptionStatus string     `gorm:"column:subscription_status;default:trial" json:"subscriptionStatus"`
	TrialStartDate     *time.Time `gorm:"column:trial_start_date" json:"trialStartDate,omitempty"`
	TrialEndDate       *time.Time `gorm:"column:trial_end_date" json:"trialEndDate,omitempty"`
	GracePeriodEndsAt  *time.Time `gorm:"column:grace_period_ends_at" json:"gracePeriodEndsAt,omitempty"`

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

	// Procurement Flow: "goods_first" (receive goods before payment) or "payment_first" (pay before receiving goods)
	ProcurementFlow string `gorm:"default:goods_first" json:"procurementFlow"`

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
	BranchID     *string `gorm:"index" json:"branchId,omitempty"`     // Foreign key to OrganizationBranch
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

	IsActive bool `gorm:"column:is_active;default:true" json:"is_active"`
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

// OrganizationBranch represents a physical branch/office location of an organization
type OrganizationBranch struct {
	ID             string    `gorm:"primaryKey" json:"id"`
	OrganizationID string    `gorm:"index;not null" json:"organizationId"`
	Name           string    `gorm:"not null" json:"name"`
	Code           string    `json:"code"`
	ProvinceID     string    `json:"provinceId,omitempty"`
	TownID         string    `json:"townId,omitempty"`
	Address        string    `json:"address,omitempty"`
	ManagerID      *string   `json:"managerId,omitempty"`
	IsActive       bool      `gorm:"column:is_active;default:true" json:"isActive"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

func (OrganizationBranch) TableName() string {
	return "organization_branches"
}

// Province represents a Zambian province (global reference data)
type Province struct {
	ID   string `gorm:"primaryKey" json:"id"`
	Name string `gorm:"not null;unique" json:"name"`
	Code string `gorm:"not null;unique" json:"code"`
}

func (Province) TableName() string { return "provinces" }

// Town represents a Zambian town/district (global reference data)
type Town struct {
	ID         string `gorm:"primaryKey" json:"id"`
	ProvinceID string `gorm:"index;not null" json:"provinceId"`
	Name       string `gorm:"not null" json:"name"`
	Code       string `json:"code,omitempty"`
}

func (Town) TableName() string { return "towns" }

// OrganizationInvitation represents a pending invitation for an existing platform
// user to join an organization.  The token field is intentionally hidden from
// list responses (json:"-") and only returned at creation time.
type OrganizationInvitation struct {
	ID             string     `gorm:"primaryKey" json:"id"`
	OrganizationID string     `gorm:"index;not null" json:"organizationId"`
	Organization   *Organization `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`

	// InvitedUserID is nil only when the invitee has no platform account yet.
	InvitedUserID  *string    `gorm:"index" json:"invitedUserId,omitempty"`
	InvitedUser    *User      `gorm:"foreignKey:InvitedUserID" json:"invitedUser,omitempty"`

	InvitedEmail   string     `gorm:"not null" json:"invitedEmail"`
	InvitedBy      string     `gorm:"not null" json:"invitedBy"`
	InvitedByUser  *User      `gorm:"foreignKey:InvitedBy" json:"invitedByUser,omitempty"`

	Role           string     `gorm:"not null;default:requester" json:"role"`
	DepartmentID   *string    `json:"departmentId,omitempty"`
	BranchID       *string    `json:"branchId,omitempty"`

	Status         string     `gorm:"not null;default:pending" json:"status"`
	Token          *string    `gorm:"uniqueIndex" json:"-"` // returned only at creation
	ExpiresAt      time.Time  `json:"expiresAt"`
	AcceptedAt     *time.Time `json:"acceptedAt,omitempty"`
	DeclinedAt     *time.Time `json:"declinedAt,omitempty"`

	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}

func (OrganizationInvitation) TableName() string {
	return "organization_invitations"
}

// Note: OrganizationRole, OrganizationPermission, and PermissionAssignment
// have been moved to enhanced_auth.go for the new RBAC system
