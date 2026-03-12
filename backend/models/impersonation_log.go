package models

import "time"

// ImpersonationLog records every impersonation event for audit and security purposes.
// Only super_admin users can view this log via the admin console.
type ImpersonationLog struct {
	ID                string     `gorm:"primaryKey" json:"id"`
	ImpersonatorID    string     `gorm:"not null" json:"impersonator_id"`
	ImpersonatorEmail string     `gorm:"not null" json:"impersonator_email"`
	TargetID          string     `gorm:"not null" json:"target_id"`
	TargetEmail       string     `gorm:"not null" json:"target_email"`
	// ImpersonationType: "platform_user" or "admin_user"
	ImpersonationType string     `gorm:"not null" json:"impersonation_type"`
	TokenJTI          string     `gorm:"not null" json:"token_jti"`
	Reason            string     `json:"reason,omitempty"`
	ExpiresAt         time.Time  `gorm:"not null" json:"expires_at"`
	Revoked           bool       `gorm:"default:false" json:"revoked"`
	RevokedAt         *time.Time `json:"revoked_at,omitempty"`
	RevokedBy         *string    `json:"revoked_by,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
}

func (ImpersonationLog) TableName() string { return "impersonation_logs" }
