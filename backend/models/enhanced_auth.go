package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// Session represents a user session with refresh token
type Session struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID       string    `gorm:"index;not null" json:"userId"`
	User         *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	RefreshToken string    `gorm:"uniqueIndex;not null" json:"refreshToken"`
	IPAddress    string    `json:"ipAddress"`
	UserAgent    string    `json:"userAgent"`
	ExpiresAt    time.Time `json:"expiresAt"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// PasswordReset represents a password reset token
type PasswordReset struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    string     `gorm:"index;not null" json:"userId"`
	User      *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Token     string     `gorm:"uniqueIndex;not null" json:"token"`
	ExpiresAt time.Time  `json:"expiresAt"`
	UsedAt    *time.Time `json:"usedAt,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
}

// EmailVerification represents an email verification token
type EmailVerification struct {
	ID         uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID     string     `gorm:"index;not null" json:"userId"`
	User       *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Email      string     `gorm:"not null" json:"email"`
	Token      string     `gorm:"uniqueIndex;not null" json:"token"`
	ExpiresAt  time.Time  `json:"expiresAt"`
	VerifiedAt *time.Time `json:"verifiedAt,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
}

// LoginAttempt represents a login attempt for security tracking
type LoginAttempt struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID        *string   `gorm:"index" json:"userId,omitempty"`
	User          *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Email         string    `gorm:"index;not null" json:"email"`
	IPAddress     string    `gorm:"index" json:"ipAddress"`
	UserAgent     string    `json:"userAgent"`
	Success       bool      `gorm:"default:false" json:"success"`
	FailureReason string    `json:"failureReason,omitempty"`
	AttemptedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"attemptedAt"`
}

// AccountLockout represents an account lockout for security
type AccountLockout struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID     string    `gorm:"index;not null" json:"userId"`
	User       *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Email      string    `gorm:"not null" json:"email"`
	IPAddress  string    `json:"ipAddress"`
	Reason     string    `gorm:"not null" json:"reason"`
	LockedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"lockedAt"`
	UnlocksAt  time.Time `gorm:"not null" json:"unlocksAt"`
	Active     bool      `gorm:"default:true" json:"active"`
}

// OrganizationRole represents a custom role within an organization
type OrganizationRole struct {
	ID             uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	OrganizationID string         `gorm:"index;not null" json:"organizationId"`
	Organization   *Organization  `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	Name           string         `gorm:"not null" json:"name"`
	Description    string         `json:"description"`
	IsSystemRole   bool           `gorm:"default:false" json:"isSystemRole"`
	Permissions    datatypes.JSON `gorm:"type:jsonb;default:'[]'" json:"permissions"`
	Active         bool           `gorm:"default:true" json:"active"`
	CreatedBy      *string        `json:"createdBy,omitempty"`
	Creator        *User          `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
}

// UserOrganizationRole represents a user's role assignment within an organization
type UserOrganizationRole struct {
	ID             uuid.UUID         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID         string            `gorm:"index;not null" json:"userId"`
	User           *User             `gorm:"foreignKey:UserID" json:"user,omitempty"`
	OrganizationID string            `gorm:"index;not null" json:"organizationId"`
	Organization   *Organization     `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	RoleID         uuid.UUID         `gorm:"type:uuid;not null" json:"roleId"`
	Role           *OrganizationRole `gorm:"foreignKey:RoleID" json:"role,omitempty"`
	AssignedBy     *string           `json:"assignedBy,omitempty"`
	Assigner       *User             `gorm:"foreignKey:AssignedBy" json:"assigner,omitempty"`
	AssignedAt     time.Time         `gorm:"default:CURRENT_TIMESTAMP" json:"assignedAt"`
	Active         bool              `gorm:"default:true" json:"active"`
}

// Workflow represents a workflow definition for document approvals
type Workflow struct {
	ID             uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	OrganizationID string         `gorm:"index;not null" json:"organizationId"`
	Organization   *Organization  `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	Name           string         `gorm:"not null" json:"name"`
	Description    string         `json:"description"`
	DocumentType   string         `gorm:"index;not null" json:"documentType"`
	Stages         datatypes.JSON `gorm:"type:jsonb;not null;default:'[]'" json:"stages"`
	IsActive       bool           `gorm:"default:true" json:"isActive"`
	IsDefault      bool           `gorm:"default:false" json:"isDefault"`
	CreatedBy      *string        `json:"createdBy,omitempty"`
	Creator        *User          `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
}

// ApprovalTaskEnhanced represents an enhanced approval task with workflow support
type ApprovalTaskEnhanced struct {
	ID             uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	OrganizationID string         `gorm:"index;not null" json:"organizationId"`
	Organization   *Organization  `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	DocumentID     string         `gorm:"index;not null" json:"documentId"`
	DocumentType   string         `gorm:"index;not null" json:"documentType"`
	WorkflowID     *uuid.UUID     `gorm:"type:uuid" json:"workflowId,omitempty"`
	Workflow       *Workflow      `gorm:"foreignKey:WorkflowID" json:"workflow,omitempty"`
	AssignedTo     string         `gorm:"index;not null" json:"assignedTo"`
	Assignee       *User          `gorm:"foreignKey:AssignedTo" json:"assignee,omitempty"`
	AssignedBy     *string        `json:"assignedBy,omitempty"`
	Assigner       *User          `gorm:"foreignKey:AssignedBy" json:"assigner,omitempty"`
	Status         string         `gorm:"default:'PENDING'" json:"status"`
	CurrentStage   int            `gorm:"default:1" json:"currentStage"`
	TotalStages    int            `gorm:"default:1" json:"totalStages"`
	Priority       string         `gorm:"default:'MEDIUM'" json:"priority"`
	DueDate        *time.Time     `json:"dueDate,omitempty"`
	Notes          string         `json:"notes"`
	Metadata       datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"metadata"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
}

// ApprovalHistory represents the history of actions on approval tasks
type ApprovalHistory struct {
	ID        uuid.UUID              `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TaskID    uuid.UUID              `gorm:"type:uuid;index;not null" json:"taskId"`
	Task      *ApprovalTaskEnhanced  `gorm:"foreignKey:TaskID" json:"task,omitempty"`
	UserID    string                 `gorm:"index;not null" json:"userId"`
	User      *User                  `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Action    string                 `gorm:"not null" json:"action"`
	Stage     int                    `gorm:"not null" json:"stage"`
	Comment   string                 `json:"comment"`
	Signature string                 `json:"signature"`
	IPAddress string                 `json:"ipAddress"`
	UserAgent string                 `json:"userAgent"`
	Metadata  datatypes.JSON         `gorm:"type:jsonb;default:'{}'" json:"metadata"`
	CreatedAt time.Time              `json:"createdAt"`
}

// NotificationEnhanced represents an enhanced notification system
type NotificationEnhanced struct {
	ID             uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	OrganizationID string         `gorm:"index;not null" json:"organizationId"`
	Organization   *Organization  `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	UserID         string         `gorm:"index;not null" json:"userId"`
	User           *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Type           string         `gorm:"index;not null" json:"type"`
	Title          string         `gorm:"not null" json:"title"`
	Message        string         `gorm:"not null" json:"message"`
	RelatedID      *uuid.UUID     `gorm:"type:uuid" json:"relatedId,omitempty"`
	RelatedType    string         `json:"relatedType,omitempty"`
	IsRead         bool           `gorm:"default:false" json:"isRead"`
	SentViaEmail   bool           `gorm:"default:false" json:"sentViaEmail"`
	EmailSentAt    *time.Time     `json:"emailSentAt,omitempty"`
	Priority       string         `gorm:"default:'MEDIUM'" json:"priority"`
	Metadata       datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"metadata"`
	ExpiresAt      *time.Time     `json:"expiresAt,omitempty"`
	CreatedAt      time.Time      `json:"createdAt"`
}

// Permission represents a system permission
type Permission struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Resource    string `json:"resource"`
	Action      string `json:"action"`
	Category    string `json:"category"`
}

// RolePermission represents the permissions assigned to a role
type RolePermission struct {
	RoleID       uuid.UUID   `json:"roleId"`
	Role         *OrganizationRole `json:"role,omitempty"`
	PermissionID string      `json:"permissionId"`
	Permission   *Permission `json:"permission,omitempty"`
	GrantedBy    *string     `json:"grantedBy,omitempty"`
	GrantedAt    time.Time   `json:"grantedAt"`
}

// Table names for GORM
func (Session) TableName() string                    { return "sessions" }
func (PasswordReset) TableName() string             { return "password_resets" }
func (EmailVerification) TableName() string         { return "email_verifications" }
func (LoginAttempt) TableName() string              { return "login_attempts" }
func (AccountLockout) TableName() string            { return "account_lockouts" }
func (OrganizationRole) TableName() string          { return "organization_roles" }
func (UserOrganizationRole) TableName() string      { return "user_organization_roles" }
func (Workflow) TableName() string                  { return "workflows" }
func (ApprovalTaskEnhanced) TableName() string      { return "approval_tasks_enhanced" }
func (ApprovalHistory) TableName() string           { return "approval_history" }
func (NotificationEnhanced) TableName() string      { return "notifications_enhanced" }

// Helper methods for Session
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// Helper methods for PasswordReset
func (pr *PasswordReset) IsExpired() bool {
	return time.Now().After(pr.ExpiresAt)
}

func (pr *PasswordReset) IsUsed() bool {
	return pr.UsedAt != nil
}

// Helper methods for EmailVerification
func (ev *EmailVerification) IsExpired() bool {
	return time.Now().After(ev.ExpiresAt)
}

func (ev *EmailVerification) IsVerified() bool {
	return ev.VerifiedAt != nil
}

// Helper methods for AccountLockout
func (al *AccountLockout) IsActive() bool {
	return al.Active && time.Now().Before(al.UnlocksAt)
}

// Helper methods for ApprovalTaskEnhanced
func (at *ApprovalTaskEnhanced) IsOverdue() bool {
	return at.DueDate != nil && time.Now().After(*at.DueDate) && at.Status == "PENDING"
}

func (at *ApprovalTaskEnhanced) IsPending() bool {
	return at.Status == "PENDING" || at.Status == "IN_REVIEW"
}

func (at *ApprovalTaskEnhanced) IsCompleted() bool {
	return at.Status == "APPROVED" || at.Status == "REJECTED"
}

// Helper methods for NotificationEnhanced
func (n *NotificationEnhanced) IsExpired() bool {
	return n.ExpiresAt != nil && time.Now().After(*n.ExpiresAt)
}

func (n *NotificationEnhanced) IsHighPriority() bool {
	return n.Priority == "HIGH" || n.Priority == "URGENT"
}