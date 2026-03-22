package models

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
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

// OrganizationRole represents a role (global system role or org-specific custom role)
// System roles have OrganizationID = nil (global), custom roles have OrganizationID set
type OrganizationRole struct {
	ID             uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	OrganizationID *string        `gorm:"index" json:"organizationId"` // nil for global system roles
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
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID string         `gorm:"index;not null" json:"organizationId"`
	Organization   *Organization  `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	Name           string         `gorm:"not null" json:"name"`
	Description    string         `json:"description"`
	DocumentType   string         `gorm:"index;not null" json:"documentType"` // Keep original field name for compatibility
	EntityType     string         `gorm:"index;not null" json:"entityType"` // "requisition", "purchase_order", "grn", "payment_voucher"
	Version        int            `gorm:"default:1" json:"version"`
	IsActive       bool           `gorm:"default:true" json:"isActive"`
	IsDefault      bool           `gorm:"default:false" json:"isDefault"`
	Conditions     datatypes.JSON `gorm:"type:jsonb" json:"conditions,omitempty"`
	Stages         datatypes.JSON `gorm:"type:jsonb;not null" json:"stages"`
	CreatedBy      string         `gorm:"not null" json:"createdBy"`
	Creator        *User          `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
	DeletedAt      gorm.DeletedAt `json:"deletedAt" gorm:"index"`
	
	// Computed fields for frontend compatibility
	TotalStages int `gorm:"-" json:"totalStages"`
	UsageCount  int `gorm:"-" json:"usageCount"`
}

// WorkflowStage represents a single stage in a workflow
type WorkflowStage struct {
	StageNumber       int    `json:"stageNumber"`
	StageName         string `json:"stageName"`
	Description       string `json:"description,omitempty"`
	RequiredRole      string `json:"requiredRole"`
	RequiredApprovals int    `json:"requiredApprovals"`
	TimeoutHours      *int   `json:"timeoutHours,omitempty"`
	CanReject         bool   `json:"canReject"`
	CanReassign       bool   `json:"canReassign"`
	
	// NEW: Enhanced approval support
	RequiredApprovalCount int    `json:"requiredApprovalCount"` // Default: 1, for multiple approvals
	ApprovalType          string `json:"approvalType"`          // "any", "all", "majority", "quorum"
	QuorumCount           *int   `json:"quorumCount,omitempty"` // For quorum-based approval
	AllowSelfApproval     bool   `json:"allowSelfApproval"`     // Can creator approve their own document
	RequireUnanimous      bool   `json:"requireUnanimous"`      // All qualified users must approve
	EscalationUserID      *string `json:"escalationUserId,omitempty"`
	
	// NEW: Assignment strategy support
	AssignmentStrategy    string   `json:"assignmentStrategy"`    // "role", "round_robin", "specific_user", "user_group"
	AssignedUserIDs       []string `json:"assignedUserIds,omitempty"`
	AssignedGroupID       *string  `json:"assignedGroupId,omitempty"`

	// Computed: resolved role name for display (populated by loadComputedFields)
	RequiredRoleName string `json:"requiredRoleName,omitempty"`
}

// Validate validates the workflow stage configuration
func (ws *WorkflowStage) Validate() error {
	if ws.StageNumber <= 0 {
		return fmt.Errorf("stage number must be positive")
	}
	if ws.StageName == "" {
		return fmt.Errorf("stage name is required")
	}
	if ws.RequiredRole == "" && ws.AssignmentStrategy != "specific_user" {
		return fmt.Errorf("required role is required unless using specific user assignment")
	}
	if ws.RequiredApprovalCount <= 0 {
		ws.RequiredApprovalCount = 1 // Default to 1
	}
	if ws.ApprovalType == "" {
		ws.ApprovalType = "any" // Default to any
	}
	if ws.ApprovalType == "quorum" && ws.QuorumCount == nil {
		return fmt.Errorf("quorum count is required for quorum-based approval")
	}
	if ws.AssignmentStrategy == "" {
		ws.AssignmentStrategy = "role" // Default to role-based
	}
	return nil
}

// WorkflowConditions defines when a workflow should be applied and routing behavior
type WorkflowConditions struct {
	AmountRange  *AmountRange           `json:"amountRange,omitempty"`
	Departments  []string               `json:"departments,omitempty"`
	Priority     []string               `json:"priority,omitempty"`
	Categories   []string               `json:"categories,omitempty"`
	CustomFields map[string]interface{} `json:"customFields,omitempty"`

	// Routing behavior
	RoutingType    string `json:"routingType,omitempty"`    // "procurement" (default) or "accounting"
	AutoApprove    bool   `json:"autoApprove,omitempty"`    // Skip workflow stages when criteria met
	AutoGeneratePO bool   `json:"autoGeneratePO,omitempty"` // Auto-create PO after approval
	AutoApprovePO  bool   `json:"autoApprovePO,omitempty"`  // Create PO as "approved" not "draft"

	// Auto-approval criteria (only evaluated when AutoApprove is true)
	AutoApprovalMaxAmount  *float64 `json:"autoApprovalMaxAmount,omitempty"`  // Max amount for auto-approval
	AutoApprovalCategories []string `json:"autoApprovalCategories,omitempty"` // e.g. ["petty_cash","stationery"]
}

// AmountRange defines monetary range conditions
type AmountRange struct {
	Min *float64 `json:"min,omitempty"`
	Max *float64 `json:"max,omitempty"`
}

// WorkflowAssignment tracks workflow execution for specific entities
type WorkflowAssignment struct {
	ID                string         `gorm:"primaryKey" json:"id"`
	OrganizationID    string         `gorm:"index;not null" json:"organizationId"`
	Organization      *Organization  `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	EntityID          string         `gorm:"not null;index" json:"entityId"`
	EntityType        string         `gorm:"not null" json:"entityType"`
	WorkflowID        uuid.UUID      `gorm:"type:uuid;not null;index" json:"workflowId"`
	Workflow          *Workflow      `gorm:"foreignKey:WorkflowID" json:"workflow,omitempty"`
	WorkflowVersion   int            `gorm:"not null" json:"workflowVersion"`
	CurrentStage      int            `gorm:"default:0" json:"currentStage"`
	Status            string         `gorm:"default:'IN_PROGRESS'" json:"status"` // "IN_PROGRESS", "COMPLETED", "REJECTED", "CANCELLED"
	StageHistory      datatypes.JSON `gorm:"type:jsonb" json:"stageHistory"`
	AssignedAt        time.Time      `json:"assignedAt"`
	AssignedBy        string         `gorm:"not null" json:"assignedBy"`
	Assigner          *User          `gorm:"foreignKey:AssignedBy" json:"assigner,omitempty"`
	CompletedAt       *time.Time     `json:"completedAt,omitempty"`
	CreatedAt         time.Time      `json:"createdAt"`
	UpdatedAt         time.Time      `json:"updatedAt"`
}

// StageExecution represents the execution of a single workflow stage
type StageExecution struct {
	StageNumber  int       `json:"stageNumber"`
	StageName    string    `json:"stageName"`
	ApproverID   string    `json:"approverId"`
	ApproverName string    `json:"approverName"`
	ApproverRole string    `json:"approverRole"`
	Action       string    `json:"action"` // "approved", "rejected", "reassigned"
	Comments     string    `json:"comments,omitempty"`
	Signature    string    `json:"signature,omitempty"`
	ExecutedAt   time.Time `json:"executedAt"`
}

// WorkflowTask represents a pending approval task
type WorkflowTask struct {
	ID                   string             `gorm:"primaryKey" json:"id"`
	OrganizationID       string             `gorm:"index;not null" json:"organizationId"`
	Organization         *Organization      `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	WorkflowAssignmentID string             `gorm:"not null;index" json:"workflowAssignmentId"`
	WorkflowAssignment   *WorkflowAssignment `gorm:"foreignKey:WorkflowAssignmentID" json:"workflowAssignment,omitempty"`
	EntityID             string             `gorm:"not null;index" json:"entityId"`
	EntityType           string             `gorm:"not null" json:"entityType"`
	StageNumber          int                `gorm:"not null" json:"stageNumber"`
	StageName            string             `gorm:"not null" json:"stageName"`
	
	// Assignment details
	AssignmentType string  `gorm:"default:'role'" json:"assignmentType"` // "role", "specific_user"
	AssignedRole   *string `json:"assignedRole,omitempty"`
	AssignedUserID *string `json:"assignedUserId,omitempty"`
	AssignedUser   *User   `gorm:"foreignKey:AssignedUserID" json:"assignedUser,omitempty"`
	
	// Task lifecycle
	Status      string     `gorm:"default:'PENDING'" json:"status"` // "PENDING", "CLAIMED", "COMPLETED", "EXPIRED"
	Priority    string     `gorm:"default:'MEDIUM'" json:"priority"` // "LOW", "MEDIUM", "HIGH", "URGENT"
	CreatedAt   time.Time  `json:"createdAt"`
	ClaimedAt   *time.Time `json:"claimedAt,omitempty"`
	ClaimedBy   *string    `json:"claimedBy,omitempty"`
	Claimer     *User      `gorm:"foreignKey:ClaimedBy" json:"claimer,omitempty"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
	DueDate     *time.Time `json:"dueDate,omitempty"`
	
	// NEW: Optimistic locking and enhanced claiming
	Version      int        `gorm:"default:1;not null" json:"version"`
	UpdatedBy    *string    `json:"updatedBy,omitempty"`
	ClaimExpiry  *time.Time `json:"claimExpiry,omitempty"`

	// Frontend compatibility fields - computed at runtime
	DocumentID       string     `gorm:"-" json:"documentId,omitempty"`       // Maps to EntityID
	DocumentType     string     `gorm:"-" json:"documentType,omitempty"`     // Maps to EntityType
	DocumentNumber   string     `gorm:"-" json:"documentNumber,omitempty"`   // Computed from document reference
	ApproverID       string     `gorm:"-" json:"approverId,omitempty"`       // Maps to AssignedUserID or ClaimedBy
	AssignedTo       string     `gorm:"-" json:"assignedTo,omitempty"`       // Maps to AssignedUserID or ClaimedBy
	ApproverName     string     `gorm:"-" json:"approverName,omitempty"`     // Computed from AssignedUser.Name
	ClaimerName      string     `gorm:"-" json:"claimerName,omitempty"`      // Computed from Claimer.Name
	AssignedRoleName string     `gorm:"-" json:"assignedRoleName,omitempty"` // Human-readable name for AssignedRole (resolved from UUID or plain name)
	TaskType         string     `gorm:"-" json:"taskType,omitempty"`         // Computed task type for UI display
	Title            string     `gorm:"-" json:"title,omitempty"`            // Computed human-readable task title
	WorkflowID       string     `gorm:"-" json:"workflowId,omitempty"`       // Computed workflow ID for the task
	WorkflowName     string     `gorm:"-" json:"workflowName,omitempty"`     // Computed workflow name for the task
	Importance       string     `gorm:"-" json:"importance,omitempty"`       // Computed from priority
	Stage            int        `gorm:"-" json:"stage,omitempty"`            // Maps to StageNumber
	DueAt            *time.Time `gorm:"-" json:"dueAt,omitempty"`            // Maps to DueDate
}

// StageApprovalRecord tracks individual approvals per stage for multiple approval support
type StageApprovalRecord struct {
	ID               string    `gorm:"primaryKey" json:"id"`
	OrganizationID   string    `gorm:"index;not null" json:"organizationId"`
	Organization     *Organization `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	WorkflowTaskID   string    `gorm:"index;not null" json:"workflowTaskId"`
	WorkflowTask     *WorkflowTask `gorm:"foreignKey:WorkflowTaskID" json:"workflowTask,omitempty"`
	StageNumber      int       `gorm:"not null" json:"stageNumber"`
	ApproverID       string    `gorm:"not null" json:"approverId"`
	Approver         *User     `gorm:"foreignKey:ApproverID" json:"approver,omitempty"`
	ApproverName     string    `json:"approverName"`
	ApproverRole     string    `json:"approverRole"`
	Action           string    `json:"action"` // "approved", "rejected"
	Comments         string    `json:"comments"`
	Signature        string    `json:"signature"`
	ApprovedAt       time.Time `json:"approvedAt"`
	IPAddress        string    `json:"ipAddress"`
	UserAgent        string    `json:"userAgent"`
	CreatedAt        time.Time `json:"createdAt"`
}

// TaskAssignmentHistory tracks round-robin assignment history
type TaskAssignmentHistory struct {
	ID             string    `gorm:"primaryKey" json:"id"`
	OrganizationID string    `gorm:"index;not null" json:"organizationId"`
	Organization   *Organization `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	Role           string    `gorm:"index;not null" json:"role"`
	AssignedUserID string    `gorm:"not null" json:"assignedUserId"`
	AssignedUser   *User     `gorm:"foreignKey:AssignedUserID" json:"assignedUser,omitempty"`
	AssignedAt     time.Time `json:"assignedAt"`
}

// WorkflowDefault tracks default workflows for entity types
type WorkflowDefault struct {
	ID                     string    `gorm:"primaryKey" json:"id"`
	OrganizationID         string    `gorm:"index;not null" json:"organizationId"`
	Organization           *Organization `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	EntityType             string    `gorm:"uniqueIndex:idx_org_entity_default" json:"entityType"`
	DefaultWorkflowID      uuid.UUID `gorm:"type:uuid;not null" json:"defaultWorkflowId"`
	DefaultWorkflow        *Workflow `gorm:"foreignKey:DefaultWorkflowID" json:"defaultWorkflow,omitempty"`
	DefaultWorkflowVersion int       `gorm:"not null" json:"defaultWorkflowVersion"`
	SetBy                  string    `gorm:"not null" json:"setBy"`
	Setter                 *User     `gorm:"foreignKey:SetBy" json:"setter,omitempty"`
	SetAt                  time.Time `json:"setAt"`
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
func (WorkflowAssignment) TableName() string        { return "workflow_assignments" }
func (WorkflowTask) TableName() string              { return "workflow_tasks" }
func (WorkflowDefault) TableName() string           { return "workflow_defaults" }
func (StageApprovalRecord) TableName() string       { return "stage_approval_records" }
func (TaskAssignmentHistory) TableName() string     { return "task_assignment_history" }
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

// Helper methods for Workflow
func (w *Workflow) GetStages() ([]WorkflowStage, error) {
	var stages []WorkflowStage
	if err := json.Unmarshal(w.Stages, &stages); err != nil {
		return nil, fmt.Errorf("failed to unmarshal stages: %w", err)
	}
	return stages, nil
}

func (w *Workflow) SetStages(stages []WorkflowStage) error {
	stagesJSON, err := json.Marshal(stages)
	if err != nil {
		return fmt.Errorf("failed to marshal stages: %w", err)
	}
	w.Stages = stagesJSON
	w.TotalStages = len(stages)
	return nil
}

func (w *Workflow) GetConditions() (*WorkflowConditions, error) {
	if w.Conditions == nil {
		return nil, nil
	}
	
	var conditions WorkflowConditions
	if err := json.Unmarshal(w.Conditions, &conditions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal conditions: %w", err)
	}
	return &conditions, nil
}

func (w *Workflow) SetConditions(conditions *WorkflowConditions) error {
	if conditions == nil {
		w.Conditions = nil
		return nil
	}
	
	conditionsJSON, err := json.Marshal(conditions)
	if err != nil {
		return fmt.Errorf("failed to marshal conditions: %w", err)
	}
	w.Conditions = conditionsJSON
	return nil
}

func (w *Workflow) Validate() error {
	if w.Name == "" {
		return fmt.Errorf("workflow name is required")
	}
	if w.EntityType == "" {
		return fmt.Errorf("entity type is required")
	}

	// Allow 0 stages only if auto-approve is enabled on the workflow conditions
	if len(w.Stages) == 0 {
		conditions, _ := w.GetConditions()
		if conditions != nil && conditions.AutoApprove {
			return nil
		}
		return fmt.Errorf("workflow must have at least one stage (or enable auto-approval)")
	}
	return nil
}

// IsAccountingWorkflow returns true if this workflow is configured for the accounting/direct-payment path.
func (w *Workflow) IsAccountingWorkflow() bool {
	conditions, err := w.GetConditions()
	if err != nil || conditions == nil {
		return false
	}
	return strings.EqualFold(conditions.RoutingType, "accounting")
}

// SupportsAutoApproval returns true if this workflow supports auto-approval.
func (w *Workflow) SupportsAutoApproval() bool {
	conditions, err := w.GetConditions()
	if err != nil || conditions == nil {
		return false
	}
	return conditions.AutoApprove
}

// Helper methods for WorkflowAssignment
func (wa *WorkflowAssignment) GetStageHistory() ([]StageExecution, error) {
	var history []StageExecution
	if wa.StageHistory == nil {
		return history, nil
	}
	
	if err := json.Unmarshal(wa.StageHistory, &history); err != nil {
		return nil, fmt.Errorf("failed to unmarshal stage history: %w", err)
	}
	return history, nil
}

func (wa *WorkflowAssignment) AddStageExecution(execution StageExecution) error {
	history, err := wa.GetStageHistory()
	if err != nil {
		return err
	}
	
	history = append(history, execution)
	
	historyJSON, err := json.Marshal(history)
	if err != nil {
		return fmt.Errorf("failed to marshal stage history: %w", err)
	}
	
	wa.StageHistory = historyJSON
	return nil
}

func (wa *WorkflowAssignment) IsCompleted() bool {
	return strings.ToUpper(wa.Status) == "COMPLETED"
}

func (wa *WorkflowAssignment) IsInProgress() bool {
	return strings.ToUpper(wa.Status) == "IN_PROGRESS"
}

func (wa *WorkflowAssignment) IsRejected() bool {
	return strings.ToUpper(wa.Status) == "REJECTED"
}

// Helper methods for WorkflowTask
func (wt *WorkflowTask) IsPending() bool {
	return strings.ToUpper(wt.Status) == "PENDING"
}

func (wt *WorkflowTask) IsClaimed() bool {
	return strings.ToUpper(wt.Status) == "CLAIMED"
}

func (wt *WorkflowTask) IsCompleted() bool {
	return strings.ToUpper(wt.Status) == "COMPLETED"
}

func (wt *WorkflowTask) IsOverdue() bool {
	return wt.DueDate != nil && time.Now().After(*wt.DueDate) && !wt.IsCompleted()
}

// Helper methods for WorkflowConditions

// documentData is an internal representation used for condition matching.
type documentData struct {
	TotalAmount float64
	Department  string
	Priority    string
	CategoryID  string
}

// extractDocumentData extracts comparable fields from various document types.
func extractDocumentData(document interface{}) (documentData, bool) {
	switch d := document.(type) {
	case *Requisition:
		catID := ""
		if d.CategoryID != nil {
			catID = *d.CategoryID
		}
		return documentData{
			TotalAmount: d.TotalAmount,
			Department:  d.Department,
			Priority:    d.Priority,
			CategoryID:  catID,
		}, true
	case *PurchaseOrder:
		return documentData{
			TotalAmount: d.TotalAmount,
			Department:  d.Department,
			Priority:    d.Priority,
		}, true
	case *PaymentVoucher:
		return documentData{
			TotalAmount: d.Amount,
		}, true
	default:
		return documentData{}, false
	}
}

// stringSliceContains checks if a slice contains a string (case-insensitive).
func stringSliceContains(slice []string, val string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, val) {
			return true
		}
	}
	return false
}

// MatchesDocument checks whether a document satisfies this condition set.
func (wc *WorkflowConditions) MatchesDocument(document interface{}) bool {
	if wc == nil {
		return true
	}

	doc, ok := extractDocumentData(document)
	if !ok {
		return true // Unknown type, match by default
	}

	if wc.AmountRange != nil {
		if wc.AmountRange.Min != nil && doc.TotalAmount < *wc.AmountRange.Min {
			return false
		}
		if wc.AmountRange.Max != nil && doc.TotalAmount > *wc.AmountRange.Max {
			return false
		}
	}

	if len(wc.Departments) > 0 && doc.Department != "" {
		if !stringSliceContains(wc.Departments, doc.Department) {
			return false
		}
	}

	if len(wc.Priority) > 0 && doc.Priority != "" {
		if !stringSliceContains(wc.Priority, doc.Priority) {
			return false
		}
	}

	if len(wc.Categories) > 0 && doc.CategoryID != "" {
		if !stringSliceContains(wc.Categories, doc.CategoryID) {
			return false
		}
	}

	return true
}

// MeetsAutoApprovalCriteria checks if a document qualifies for auto-approval
// under this workflow's conditions. Returns false if AutoApprove is not enabled.
func (wc *WorkflowConditions) MeetsAutoApprovalCriteria(totalAmount float64, categoryID string) bool {
	if wc == nil || !wc.AutoApprove {
		return false
	}

	if wc.AutoApprovalMaxAmount != nil && totalAmount > *wc.AutoApprovalMaxAmount {
		return false
	}

	if len(wc.AutoApprovalCategories) > 0 {
		if !stringSliceContains(wc.AutoApprovalCategories, categoryID) {
			return false
		}
	}

	return true
}