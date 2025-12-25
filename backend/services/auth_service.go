package services

import (
	"crypto/sha256"
	"fmt"
	"log"
	"time"

	"github.com/liyali/liyali-gateway/models"
	"gorm.io/gorm"
)

// AuthService handles authentication security operations
// This includes token blacklisting, login attempt tracking, and account lockout
type AuthService struct {
	db *gorm.DB
}

// NewAuthService creates a new auth service
func NewAuthService(db *gorm.DB) *AuthService {
	return &AuthService{db: db}
}

// Configuration constants
const (
	MaxFailedLoginAttempts = 5
	AccountLockoutDuration = 15 * time.Minute
	TokenBlacklistCleanupAge = 7 * 24 * time.Hour // Delete entries older than 7 days
)

// ============== TOKEN REVOCATION ==============

// BlacklistToken adds a token to the blacklist (for logout)
func (as *AuthService) BlacklistToken(userID, tokenJTI, tokenString string, expiresAt time.Time, reason string) error {
	// Hash the token for storage (never store raw token)
	tokenHash := hashToken(tokenString)

	blacklistedToken := models.NewTokenBlacklist(userID, tokenJTI, tokenHash, expiresAt, reason)

	if err := as.db.Create(blacklistedToken).Error; err != nil {
		log.Printf("Error blacklisting token: %v", err)
		return fmt.Errorf("failed to blacklist token")
	}

	log.Printf("Token blacklisted for user %s: %s", userID, reason)
	return nil
}

// IsTokenBlacklisted checks if a token is in the blacklist
func (as *AuthService) IsTokenBlacklisted(tokenJTI string) bool {
	var count int64
	if err := as.db.Model(&models.TokenBlacklist{}).
		Where("token_jti = ?", tokenJTI).
		Count(&count).Error; err != nil {
		log.Printf("Error checking token blacklist: %v", err)
		return false // Assume not blacklisted on error (allow access)
	}
	return count > 0
}

// RevokeUserTokens revokes all tokens for a user (used for password change, deactivation, etc)
func (as *AuthService) RevokeUserTokens(userID string, reason string) error {
	// This is a simple approach - in production, you might want to track issued tokens
	// For now, we log the revocation request
	log.Printf("Token revocation request for user %s: %s", userID, reason)
	return nil
}

// CleanupExpiredTokens removes blacklist entries that have expired
func (as *AuthService) CleanupExpiredTokens() error {
	// Delete blacklist entries older than the cleanup age
	if err := as.db.Where("expires_at < ?", time.Now().Add(-TokenBlacklistCleanupAge)).
		Delete(&models.TokenBlacklist{}).Error; err != nil {
		log.Printf("Error cleaning up expired tokens: %v", err)
		return fmt.Errorf("failed to cleanup expired tokens")
	}
	return nil
}

// ============== LOGIN ATTEMPT TRACKING ==============

// RecordLoginAttempt records a login attempt (success or failure)
func (as *AuthService) RecordLoginAttempt(userID, email, ipAddress string, success bool, userAgent, reason string) error {
	attempt := models.NewLoginAttempt(userID, email, ipAddress, success, userAgent, reason)

	if err := as.db.Create(attempt).Error; err != nil {
		log.Printf("Error recording login attempt: %v", err)
		return fmt.Errorf("failed to record login attempt")
	}

	return nil
}

// GetRecentFailedAttempts gets failed login attempts for a user in recent time
func (as *AuthService) GetRecentFailedAttempts(email string, since time.Duration) (int64, error) {
	var count int64
	if err := as.db.Model(&models.LoginAttempt{}).
		Where("email = ? AND success = ? AND attempt_at > ?", email, false, time.Now().Add(-since)).
		Count(&count).Error; err != nil {
		log.Printf("Error counting failed login attempts: %v", err)
		return 0, fmt.Errorf("failed to get failed login attempts")
	}
	return count, nil
}

// ============== ACCOUNT LOCKOUT ==============

// LockAccount locks a user account
func (as *AuthService) LockAccount(userID, email, ipAddress, reason string) error {
	lockout := models.NewAccountLockout(userID, email, ipAddress, reason, AccountLockoutDuration)

	if err := as.db.Create(lockout).Error; err != nil {
		log.Printf("Error locking account: %v", err)
		return fmt.Errorf("failed to lock account")
	}

	log.Printf("Account locked for user %s: %s", userID, reason)
	return nil
}

// IsAccountLocked checks if a user account is currently locked
func (as *AuthService) IsAccountLocked(userID string) (bool, error) {
	var lockout models.AccountLockout
	err := as.db.Where("user_id = ? AND active = ? AND unlocks_at > ?", userID, true, time.Now()).
		First(&lockout).Error

	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	if err != nil {
		log.Printf("Error checking account lockout: %v", err)
		return false, fmt.Errorf("failed to check account lockout")
	}

	return true, nil
}

// UnlockAccount unlocks a user account
func (as *AuthService) UnlockAccount(userID string) error {
	if err := as.db.Model(&models.AccountLockout{}).
		Where("user_id = ?", userID).
		Update("active", false).Error; err != nil {
		log.Printf("Error unlocking account: %v", err)
		return fmt.Errorf("failed to unlock account")
	}

	log.Printf("Account unlocked for user %s", userID)
	return nil
}

// GetAccountLockoutStatus gets the current lockout status for an account
func (as *AuthService) GetAccountLockoutStatus(userID string) (*models.AccountLockout, error) {
	var lockout models.AccountLockout
	err := as.db.Where("user_id = ? AND active = ? AND unlocks_at > ?", userID, true, time.Now()).
		First(&lockout).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get lockout status")
	}

	return &lockout, nil
}

// ============== AUDIT LOGGING ==============

// LogAuthEvent logs an authentication-related event
func (as *AuthService) LogAuthEvent(userID, email string, organizationID *string, action string, success bool, details, ipAddress, userAgent string) error {
	status := "success"
	errorMsg := ""
	if !success {
		status = "failure"
	}

	auditLog := models.NewAuditLog(
		userID,
		email,
		organizationID,
		action,
		"authentication",
		userID,
		details,
		ipAddress,
		userAgent,
		status,
		errorMsg,
	)

	if err := as.db.Create(auditLog).Error; err != nil {
		log.Printf("Error logging auth event: %v", err)
		return fmt.Errorf("failed to log auth event")
	}

	return nil
}

// LogPermissionChange logs a permission-related change
func (as *AuthService) LogPermissionChange(userID, email, organizationID string, action string, resourceID string, details, ipAddress, userAgent string) error {
	auditLog := models.NewAuditLog(
		userID,
		email,
		&organizationID,
		action,
		"permission",
		resourceID,
		details,
		ipAddress,
		userAgent,
		"success",
		"",
	)

	if err := as.db.Create(auditLog).Error; err != nil {
		log.Printf("Error logging permission change: %v", err)
		return fmt.Errorf("failed to log permission change")
	}

	return nil
}

// GetAuditLogs retrieves audit logs with filters
func (as *AuthService) GetAuditLogs(userID string, organizationID *string, action string, limit int, offset int) ([]models.AuditLog, int64, error) {
	var logs []models.AuditLog
	var total int64

	query := as.db.Model(&models.AuditLog{})

	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if organizationID != nil && *organizationID != "" {
		query = query.Where("organization_id = ?", organizationID)
	}
	if action != "" {
		query = query.Where("action = ?", action)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count audit logs")
	}

	// Get paginated results
	if err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error; err != nil {
		log.Printf("Error fetching audit logs: %v", err)
		return nil, 0, fmt.Errorf("failed to fetch audit logs")
	}

	return logs, total, nil
}

// CleanupOldAuditLogs deletes old audit log entries
func (as *AuthService) CleanupOldAuditLogs(retentionDays int) error {
	cutoffDate := time.Now().Add(-time.Duration(retentionDays) * 24 * time.Hour)

	if err := as.db.Where("created_at < ?", cutoffDate).
		Delete(&models.AuditLog{}).Error; err != nil {
		log.Printf("Error cleaning up audit logs: %v", err)
		return fmt.Errorf("failed to cleanup audit logs")
	}

	return nil
}

// ============== EMAIL VERIFICATION ==============

// CreateEmailVerification creates a new email verification record
func (as *AuthService) CreateEmailVerification(userID, email, token string) (*models.EmailVerification, error) {
	verification := models.NewEmailVerification(userID, email, token, 24*time.Hour) // 24 hour expiration

	if err := as.db.Create(verification).Error; err != nil {
		log.Printf("Error creating email verification: %v", err)
		return nil, fmt.Errorf("failed to create email verification")
	}

	return verification, nil
}

// VerifyEmail marks an email as verified
func (as *AuthService) VerifyEmail(token string) error {
	var verification models.EmailVerification

	// Find the verification record
	if err := as.db.Where("token = ? AND verified_at IS NULL AND expires_at > ?", token, time.Now()).
		First(&verification).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("verification token not found or expired")
		}
		log.Printf("Error finding verification token: %v", err)
		return fmt.Errorf("failed to verify email")
	}

	// Mark as verified
	now := time.Now()
	if err := as.db.Model(&verification).Update("verified_at", now).Error; err != nil {
		log.Printf("Error verifying email: %v", err)
		return fmt.Errorf("failed to mark email as verified")
	}

	return nil
}

// IsEmailVerified checks if an email has been verified
func (as *AuthService) IsEmailVerified(email string) (bool, error) {
	var count int64
	if err := as.db.Model(&models.EmailVerification{}).
		Where("email = ? AND verified_at IS NOT NULL", email).
		Count(&count).Error; err != nil {
		log.Printf("Error checking email verification status: %v", err)
		return false, fmt.Errorf("failed to check email verification")
	}

	return count > 0, nil
}

// ============== PASSWORD RESET ==============

// CreatePasswordReset creates a new password reset token
func (as *AuthService) CreatePasswordReset(userID, email, token string) (*models.PasswordReset, error) {
	reset := models.NewPasswordReset(userID, email, token, 24*time.Hour) // 24 hour expiration

	if err := as.db.Create(reset).Error; err != nil {
		log.Printf("Error creating password reset: %v", err)
		return nil, fmt.Errorf("failed to create password reset")
	}

	return reset, nil
}

// ValidatePasswordResetToken validates and retrieves a password reset token
func (as *AuthService) ValidatePasswordResetToken(token string) (*models.PasswordReset, error) {
	var reset models.PasswordReset

	if err := as.db.Where("token = ? AND used_at IS NULL AND expires_at > ?", token, time.Now()).
		First(&reset).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("reset token not found or expired")
		}
		log.Printf("Error validating password reset token: %v", err)
		return nil, fmt.Errorf("failed to validate reset token")
	}

	return &reset, nil
}

// MarkPasswordResetUsed marks a password reset token as used
func (as *AuthService) MarkPasswordResetUsed(resetID string) error {
	now := time.Now()
	if err := as.db.Model(&models.PasswordReset{}).
		Where("id = ?", resetID).
		Update("used_at", now).Error; err != nil {
		log.Printf("Error marking password reset as used: %v", err)
		return fmt.Errorf("failed to mark password reset as used")
	}

	return nil
}

// ============== HELPER FUNCTIONS ==============

// hashToken creates a SHA256 hash of a token
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", hash)
}
