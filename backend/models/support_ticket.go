package models

import (
	"time"

	"gorm.io/datatypes"
)

// SupportTicket represents a customer support ticket managed from the admin console.
type SupportTicket struct {
	ID string `gorm:"primaryKey" json:"id"`

	TicketNumber string `gorm:"column:ticket_number;uniqueIndex;not null" json:"ticket_number"`

	OrganizationID *string       `gorm:"column:organization_id;index" json:"organization_id,omitempty"`
	Organization   *Organization `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`

	UserID *string `gorm:"column:user_id;index" json:"user_id,omitempty"`
	User   *User   `gorm:"foreignKey:UserID" json:"user,omitempty"`

	CreatedByAdminID *string `gorm:"column:created_by_admin_id;index" json:"created_by_admin_id,omitempty"`
	CreatedByAdmin   *User   `gorm:"foreignKey:CreatedByAdminID" json:"created_by_admin,omitempty"`

	AssignedToAdminID *string `gorm:"column:assigned_to_admin_id;index" json:"assigned_to_admin_id,omitempty"`
	AssignedToAdmin   *User   `gorm:"foreignKey:AssignedToAdminID" json:"assigned_to_admin,omitempty"`

	Source            string         `gorm:"column:source;not null;default:manual" json:"source"`
	Category          string         `gorm:"column:category;not null;default:general" json:"category"`
	Priority          string         `gorm:"column:priority;not null;default:medium" json:"priority"`
	Status            string         `gorm:"column:status;not null;default:open" json:"status"`
	Subject           string         `gorm:"column:subject;not null" json:"subject"`
	Description       string         `gorm:"column:description;not null" json:"description"`
	InternalNotes     string         `gorm:"column:internal_notes;not null;default:''" json:"internal_notes"`
	ExternalReference string         `gorm:"column:external_reference;not null;default:''" json:"external_reference"`
	ResolutionSummary string         `gorm:"column:resolution_summary;not null;default:''" json:"resolution_summary"`
	Metadata          datatypes.JSON `gorm:"type:jsonb;column:metadata" json:"metadata,omitempty"`

	ResolvedAt *time.Time `gorm:"column:resolved_at" json:"resolved_at,omitempty"`
	ClosedAt   *time.Time `gorm:"column:closed_at" json:"closed_at,omitempty"`

	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}
