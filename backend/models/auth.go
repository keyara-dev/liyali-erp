package models

import (
	"time"

	"github.com/google/uuid"
)

// TokenBlacklist represents a blacklisted JWT token
// Used for logout/token revocation
type TokenBlacklist struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	UserID    string    `gorm:"index" json:"userId"`
	TokenJTI  string    `gorm:"index;uniqueIndex" json:"tokenJti"` // JWT ID claim for uniqueness
	TokenHash string    `gorm:"index" json:"tokenHash"`             // Hash of token for verification
	ExpiresAt time.Time `gorm:"index" json:"expiresAt"`             // Token expiration time (auto-cleanup)
	RevokedAt time.Time `gorm:"index" json:"revokedAt"`             // When token was blacklisted
	Reason    string    `json:"reason"`                             // Why was it blacklisted (logout, password-change, etc)
}

// LoginAttempt tracks login attempts for brute force protection
type LoginAttempt struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	UserID    string    `gorm:"index" json:"userId"`
	Email     string    `gorm:"index" json:"email"`
	IPAddress string    `gorm:"index" json:"ipAddress"`
	Success   bool      `json:"success"`        // true if login succeeded
	AttemptAt time.Time `gorm:"index" json:"attemptAt"`
	UserAgent string    `json:"userAgent"`      // Browser/client info
	Reason    string    `json:"reason,omitempty"` // Error reason if failed
}

// AccountLockout tracks when an account is locked due to failed attempts
type AccountLockout struct {
	ID        string     `gorm:"primaryKey" json:"id"`
	UserID    string     `gorm:"index;uniqueIndex" json:"userId"` // Only one active lockout per user
	Email     string     `gorm:"index" json:"email"`
	LockedAt  time.Time  `json:"lockedAt"`
	UnlocksAt time.Time  `json:"unlocksAt"`      // When account automatically unlocks
	Reason    string     `json:"reason"`          // Why locked (e.g., "5 failed attempts")
	IPAddress string     `json:"ipAddress"`       // IP that triggered lockout
	Active    bool       `json:"active"`          // false if unlock is processed
}

// AuditLog tracks important security and administrative events
type AuditLog struct {
	ID            string    `gorm:"primaryKey" json:"id"`
	UserID        string    `gorm:"index" json:"userId"`
	Email         string    `gorm:"index" json:"email"`
	OrganizationID *string  `gorm:"index" json:"organizationId"`
	Action        string    `gorm:"index" json:"action"` // login, logout, register, password_change, permission_grant, etc
	Resource      string    `gorm:"index" json:"resource"` // user, role, permission, organization, etc
	ResourceID    string    `json:"resourceId"`       // ID of affected resource
	Details       string    `gorm:"type:text" json:"details"` // JSON details of what changed
	IPAddress     string    `json:"ipAddress"`
	UserAgent     string    `json:"userAgent"`
	Status        string    `json:"status"`           // success, failure
	ErrorMessage  string    `json:"errorMessage,omitempty"`
	CreatedAt     time.Time `gorm:"index" json:"createdAt"`
}

// EmailVerification tracks email verification tokens for registration
type EmailVerification struct {
	ID        string     `gorm:"primaryKey" json:"id"`
	UserID    string     `gorm:"index" json:"userId"`
	Email     string     `gorm:"index" json:"email"`
	Token     string     `gorm:"uniqueIndex" json:"token"` // Verification token
	VerifiedAt *time.Time `json:"verifiedAt"`              // null if not yet verified
	ExpiresAt time.Time  `gorm:"index" json:"expiresAt"`
	CreatedAt time.Time  `json:"createdAt"`
}

// PasswordReset tracks password reset tokens
type PasswordReset struct {
	ID        string     `gorm:"primaryKey" json:"id"`
	UserID    string     `gorm:"index" json:"userId"`
	Email     string     `gorm:"index" json:"email"`
	Token     string     `gorm:"uniqueIndex" json:"token"` // Reset token
	UsedAt    *time.Time `json:"usedAt"`                   // null if not used yet
	ExpiresAt time.Time  `gorm:"index" json:"expiresAt"`
	CreatedAt time.Time  `json:"createdAt"`
}

// NewTokenBlacklist creates a new token blacklist entry
func NewTokenBlacklist(userID, tokenJTI, tokenHash string, expiresAt time.Time, reason string) *TokenBlacklist {
	return &TokenBlacklist{
		ID:        uuid.New().String(),
		UserID:    userID,
		TokenJTI:  tokenJTI,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
		RevokedAt: time.Now(),
		Reason:    reason,
	}
}

// NewLoginAttempt creates a new login attempt record
func NewLoginAttempt(userID, email, ipAddress string, success bool, userAgent, reason string) *LoginAttempt {
	return &LoginAttempt{
		ID:        uuid.New().String(),
		UserID:    userID,
		Email:     email,
		IPAddress: ipAddress,
		Success:   success,
		AttemptAt: time.Now(),
		UserAgent: userAgent,
		Reason:    reason,
	}
}

// NewAccountLockout creates a new account lockout record
func NewAccountLockout(userID, email, ipAddress, reason string, duration time.Duration) *AccountLockout {
	now := time.Now()
	return &AccountLockout{
		ID:        uuid.New().String(),
		UserID:    userID,
		Email:     email,
		LockedAt:  now,
		UnlocksAt: now.Add(duration),
		Reason:    reason,
		IPAddress: ipAddress,
		Active:    true,
	}
}

// NewAuditLog creates a new audit log entry
func NewAuditLog(userID, email string, organizationID *string, action, resource, resourceID string, details, ipAddress, userAgent, status, errorMsg string) *AuditLog {
	return &AuditLog{
		ID:            uuid.New().String(),
		UserID:        userID,
		Email:         email,
		OrganizationID: organizationID,
		Action:        action,
		Resource:      resource,
		ResourceID:    resourceID,
		Details:       details,
		IPAddress:     ipAddress,
		UserAgent:     userAgent,
		Status:        status,
		ErrorMessage:  errorMsg,
		CreatedAt:     time.Now(),
	}
}

// NewEmailVerification creates a new email verification record
func NewEmailVerification(userID, email, token string, expiresIn time.Duration) *EmailVerification {
	return &EmailVerification{
		ID:        uuid.New().String(),
		UserID:    userID,
		Email:     email,
		Token:     token,
		ExpiresAt: time.Now().Add(expiresIn),
		CreatedAt: time.Now(),
	}
}

// NewPasswordReset creates a new password reset record
func NewPasswordReset(userID, email, token string, expiresIn time.Duration) *PasswordReset {
	return &PasswordReset{
		ID:        uuid.New().String(),
		UserID:    userID,
		Email:     email,
		Token:     token,
		ExpiresAt: time.Now().Add(expiresIn),
		CreatedAt: time.Now(),
	}
}
