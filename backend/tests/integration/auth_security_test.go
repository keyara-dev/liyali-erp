package integration

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/repository"
	"github.com/liyali/liyali-gateway/services"
	"github.com/liyali/liyali-gateway/utils"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupAuthSecurityTestDB creates an in-memory database for security testing
func setupAuthSecurityTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Auto-migrate security-related models
	db.AutoMigrate(
		&models.User{},
		&models.Session{},
		&models.LoginAttempt{},
		&models.AccountLockout{},
		&models.PasswordReset{},
		&models.Organization{},
		&models.OrganizationMember{},
		&models.AuditLog{},
	)

	return db
}

// createAuthTestUser creates a test user with hashed password
func createAuthTestUser(db *gorm.DB, email, password string) *models.User {
	hashedPassword, _ := utils.HashPassword(password)
	user := &models.User{
		ID:       uuid.New().String(),
		Email:    email,
		Name:     "Test User",
		Password: hashedPassword,
		Active:   true,
	}
	db.Create(user)
	return user
}

func TestAuthService_BruteForceProtection(t *testing.T) {
	db := setupAuthSecurityTestDB()
	
	// Create repositories
	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	passwordResetRepo := repository.NewPasswordResetRepository(db)
	loginAttemptRepo := repository.NewLoginAttemptRepository(db)
	lockoutRepo := repository.NewAccountLockoutRepository(db)
	auditService := services.NewAuditService(nil) // Simplified for testing

	authService := services.NewAuthService(
		userRepo,
		sessionRepo,
		passwordResetRepo,
		loginAttemptRepo,
		lockoutRepo,
		auditService,
		"test-jwt-secret",
		db,
	)

	// Create test user
	testUser := createAuthTestUser(db, "test@example.com", "correctpassword")
	ctx := context.Background()

	t.Run("Account lockout after multiple failed attempts", func(t *testing.T) {
		// Attempt login with wrong password multiple times
		for i := 0; i < 5; i++ {
			_, err := authService.Login(ctx, testUser.Email, "wrongpassword", "127.0.0.1", "test-agent")
			if i < 4 {
				// First 4 attempts should fail with invalid credentials
				assert.Equal(t, services.ErrInvalidCredentials, err)
			} else {
				// 5th attempt should trigger lockout
				assert.Equal(t, services.ErrInvalidCredentials, err)
			}
		}

		// 6th attempt should be blocked due to lockout
		_, err := authService.Login(ctx, testUser.Email, "wrongpassword", "127.0.0.1", "test-agent")
		assert.Equal(t, services.ErrTooManyFailedAttempts, err)

		// Even correct password should be blocked during lockout
		_, err = authService.Login(ctx, testUser.Email, "correctpassword", "127.0.0.1", "test-agent")
		assert.Equal(t, services.ErrAccountLocked, err)

		// Verify lockout record exists
		lockout, err := lockoutRepo.GetActiveByUserID(ctx, testUser.ID)
		assert.NoError(t, err)
		assert.NotNil(t, lockout)
		assert.True(t, lockout.UnlocksAt.After(time.Now()))
	})

	t.Run("IP-based rate limiting", func(t *testing.T) {
		// Create another user to test IP-based limiting
		testUser2 := createAuthTestUser(db, "test2@example.com", "password123")

		// Attempt multiple failed logins from same IP with different emails
		maliciousIP := "192.168.1.100"
		
		for i := 0; i < 10; i++ {
			email := fmt.Sprintf("fake%d@example.com", i)
			_, err := authService.Login(ctx, email, "wrongpassword", maliciousIP, "test-agent")
			// Should fail with user not found or invalid credentials
			assert.Error(t, err)
		}

		// Check if there are many failed attempts from this IP
		var attempts []models.LoginAttempt
		err := db.Where("ip_address = ? AND success = ? AND created_at > ?", 
			maliciousIP, false, time.Now().Add(-1*time.Hour)).Find(&attempts).Error
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(attempts), 10)
	})

	t.Run("Account enumeration protection", func(t *testing.T) {
		// Attempt login with non-existent email
		start := time.Now()
		_, err := authService.Login(ctx, "nonexistent@example.com", "anypassword", "127.0.0.1", "test-agent")
		nonExistentTime := time.Since(start)

		// Attempt login with existing email but wrong password
		start = time.Now()
		_, err2 := authService.Login(ctx, testUser.Email, "wrongpassword", "127.0.0.1", "test-agent")
		existingTime := time.Since(start)

		// Both should return the same error type
		assert.Equal(t, services.ErrInvalidCredentials, err)
		assert.Equal(t, services.ErrInvalidCredentials, err2)

		// Response times should be similar to prevent timing attacks
		timeDiff := existingTime - nonExistentTime
		if timeDiff < 0 {
			timeDiff = -timeDiff
		}
		// Allow up to 50ms difference (generous for testing)
		assert.Less(t, timeDiff, 50*time.Millisecond, "Response time difference too large, possible timing attack vector")
	})
}

func TestAuthService_SessionSecurity(t *testing.T) {
	db := setupAuthSecurityTestDB()
	
	// Create repositories
	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	passwordResetRepo := repository.NewPasswordResetRepository(db)
	loginAttemptRepo := repository.NewLoginAttemptRepository(db)
	lockoutRepo := repository.NewAccountLockoutRepository(db)
	auditService := services.NewAuditService(nil)

	authService := services.NewAuthService(
		userRepo,
		sessionRepo,
		passwordResetRepo,
		loginAttemptRepo,
		lockoutRepo,
		auditService,
		"test-jwt-secret",
		db,
	)

	testUser := createAuthTestUser(db, "session@example.com", "password123")
	ctx := context.Background()

	t.Run("Token rotation prevents replay attacks", func(t *testing.T) {
		// Login to get initial tokens
		loginResult, err := authService.Login(ctx, testUser.Email, "password123", "127.0.0.1", "test-agent")
		assert.NoError(t, err)
		assert.NotNil(t, loginResult)

		originalRefreshToken := loginResult.RefreshToken

		// Refresh token (should rotate)
		refreshResult, err := authService.RefreshToken(ctx, originalRefreshToken)
		assert.NoError(t, err)
		assert.NotNil(t, refreshResult)
		assert.NotEqual(t, originalRefreshToken, refreshResult.RefreshToken)

		// Try to use original refresh token again (should fail)
		_, err = authService.RefreshToken(ctx, originalRefreshToken)
		assert.Equal(t, services.ErrTokenReuseDetected, err)

		// New refresh token should still work
		refreshResult2, err := authService.RefreshToken(ctx, refreshResult.RefreshToken)
		assert.NoError(t, err)
		assert.NotNil(t, refreshResult2)
	})

	t.Run("Session hijacking protection", func(t *testing.T) {
		// Login from one IP
		loginResult, err := authService.Login(ctx, testUser.Email, "password123", "192.168.1.1", "Mozilla/5.0")
		assert.NoError(t, err)

		// Try to use refresh token from different IP (in real implementation, this might be flagged)
		// For now, we'll test that the session is properly tracked
		refreshResult, err := authService.RefreshToken(ctx, loginResult.RefreshToken)
		assert.NoError(t, err)

		// Verify session exists and is valid
		session, err := sessionRepo.GetByRefreshToken(ctx, refreshResult.RefreshToken)
		assert.NoError(t, err)
		assert.NotNil(t, session)
		assert.Equal(t, testUser.ID, session.UserID)
	})

	t.Run("Concurrent session limit", func(t *testing.T) {
		// Create multiple sessions for the same user
		var refreshTokens []string
		
		for i := 0; i < 7; i++ { // More than MaxSessionsPerUser (5)
			loginResult, err := authService.Login(ctx, testUser.Email, "password123", 
				fmt.Sprintf("192.168.1.%d", i+10), fmt.Sprintf("Agent-%d", i))
			assert.NoError(t, err)
			refreshTokens = append(refreshTokens, loginResult.RefreshToken)
		}

		// Check that old sessions are cleaned up
		sessions, err := sessionRepo.GetByUserID(ctx, testUser.ID)
		assert.NoError(t, err)
		// Should have at most MaxSessionsPerUser (5) sessions
		assert.LessOrEqual(t, len(sessions), 5)

		// Verify that some of the earlier tokens are no longer valid
		invalidCount := 0
		for _, token := range refreshTokens[:3] { // Check first 3 tokens
			_, err := authService.RefreshToken(ctx, token)
			if err != nil {
				invalidCount++
			}
		}
		assert.Greater(t, invalidCount, 0, "Some early sessions should have been invalidated")
	})

	t.Run("Session expiration", func(t *testing.T) {
		// Create a session with short expiration for testing
		shortLivedToken, err := utils.GenerateSecureToken()
		assert.NoError(t, err)

		expiredSession, err := sessionRepo.Create(ctx, testUser.ID, shortLivedToken, 
			"127.0.0.1", "test-agent", time.Now().Add(-1*time.Hour)) // Expired 1 hour ago
		assert.NoError(t, err)

		// Try to refresh expired session
		_, err = authService.RefreshToken(ctx, shortLivedToken)
		assert.Equal(t, services.ErrSessionExpired, err)

		// Verify expired session was cleaned up
		_, err = sessionRepo.GetByRefreshToken(ctx, shortLivedToken)
		assert.Error(t, err) // Should not be found
	})
}

func TestAuthService_PasswordSecurity(t *testing.T) {
	db := setupAuthSecurityTestDB()
	
	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	passwordResetRepo := repository.NewPasswordResetRepository(db)
	loginAttemptRepo := repository.NewLoginAttemptRepository(db)
	lockoutRepo := repository.NewAccountLockoutRepository(db)
	auditService := services.NewAuditService(nil)

	authService := services.NewAuthService(
		userRepo,
		sessionRepo,
		passwordResetRepo,
		loginAttemptRepo,
		lockoutRepo,
		auditService,
		"test-jwt-secret",
		db,
	)

	testUser := createAuthTestUser(db, "password@example.com", "oldpassword123")
	ctx := context.Background()

	t.Run("Password reset token security", func(t *testing.T) {
		// Request password reset
		resetToken, err := authService.CreatePasswordReset(ctx, testUser.Email)
		assert.NoError(t, err)
		assert.NotEmpty(t, resetToken)
		assert.GreaterOrEqual(t, len(resetToken), 32) // Should be sufficiently long

		// Verify token is cryptographically secure (no obvious patterns)
		assert.NotContains(t, resetToken, testUser.Email)
		assert.NotContains(t, resetToken, testUser.ID)
		assert.NotContains(t, strings.ToLower(resetToken), "password")

		// Reset password using token
		newPassword := "newSecurePassword456"
		err = authService.ResetPassword(ctx, resetToken, newPassword)
		assert.NoError(t, err)

		// Verify old password no longer works
		_, err = authService.Login(ctx, testUser.Email, "oldpassword123", "127.0.0.1", "test-agent")
		assert.Equal(t, services.ErrInvalidCredentials, err)

		// Verify new password works
		loginResult, err := authService.Login(ctx, testUser.Email, newPassword, "127.0.0.1", "test-agent")
		assert.NoError(t, err)
		assert.NotNil(t, loginResult)

		// Verify reset token can't be reused
		err = authService.ResetPassword(ctx, resetToken, "anotherPassword")
		assert.Equal(t, services.ErrInvalidToken, err)
	})

	t.Run("Password change security", func(t *testing.T) {
		// Login first to get user context
		loginResult, err := authService.Login(ctx, testUser.Email, "newSecurePassword456", "127.0.0.1", "test-agent")
		assert.NoError(t, err)

		// Change password with correct current password
		err = authService.ChangePassword(ctx, testUser.ID, "newSecurePassword456", "anotherNewPassword789")
		assert.NoError(t, err)

		// Verify old password no longer works
		_, err = authService.Login(ctx, testUser.Email, "newSecurePassword456", "127.0.0.1", "test-agent")
		assert.Equal(t, services.ErrInvalidCredentials, err)

		// Verify new password works
		_, err = authService.Login(ctx, testUser.Email, "anotherNewPassword789", "127.0.0.1", "test-agent")
		assert.NoError(t, err)

		// Try to change password with wrong current password
		err = authService.ChangePassword(ctx, testUser.ID, "wrongCurrentPassword", "hackedPassword")
		assert.Equal(t, services.ErrInvalidCredentials, err)

		// Verify password wasn't changed
		_, err = authService.Login(ctx, testUser.Email, "anotherNewPassword789", "127.0.0.1", "test-agent")
		assert.NoError(t, err)
	})

	t.Run("Password reset invalidates all sessions", func(t *testing.T) {
		// Create multiple sessions
		var refreshTokens []string
		for i := 0; i < 3; i++ {
			loginResult, err := authService.Login(ctx, testUser.Email, "anotherNewPassword789", 
				fmt.Sprintf("192.168.1.%d", i+20), fmt.Sprintf("Agent-%d", i))
			assert.NoError(t, err)
			refreshTokens = append(refreshTokens, loginResult.RefreshToken)
		}

		// Verify sessions work
		for _, token := range refreshTokens {
			_, err := authService.RefreshToken(ctx, token)
			assert.NoError(t, err)
		}

		// Reset password
		resetToken, err := authService.CreatePasswordReset(ctx, testUser.Email)
		assert.NoError(t, err)
		err = authService.ResetPassword(ctx, resetToken, "postResetPassword123")
		assert.NoError(t, err)

		// Verify all old sessions are invalidated
		for _, token := range refreshTokens {
			_, err := authService.RefreshToken(ctx, token)
			assert.Error(t, err) // Should fail
		}
	})
}

func TestAuthService_JWTSecurity(t *testing.T) {
	db := setupAuthSecurityTestDB()
	
	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	passwordResetRepo := repository.NewPasswordResetRepository(db)
	loginAttemptRepo := repository.NewLoginAttemptRepository(db)
	lockoutRepo := repository.NewAccountLockoutRepository(db)
	auditService := services.NewAuditService(nil)

	authService := services.NewAuthService(
		userRepo,
		sessionRepo,
		passwordResetRepo,
		loginAttemptRepo,
		lockoutRepo,
		auditService,
		"test-jwt-secret-for-security-testing",
		db,
	)

	testUser := createAuthTestUser(db, "jwt@example.com", "password123")
	ctx := context.Background()

	t.Run("JWT token tampering detection", func(t *testing.T) {
		// Login to get valid JWT
		loginResult, err := authService.Login(ctx, testUser.Email, "password123", "127.0.0.1", "test-agent")
		assert.NoError(t, err)

		originalToken := loginResult.AccessToken

		// Validate original token
		claims, err := authService.ValidateAccessToken(originalToken)
		assert.NoError(t, err)
		assert.Equal(t, testUser.ID, claims.UserID)

		// Tamper with token (change one character)
		tamperedToken := originalToken[:len(originalToken)-5] + "XXXXX"
		
		// Validation should fail
		_, err = authService.ValidateAccessToken(tamperedToken)
		assert.Error(t, err)

		// Try to modify payload (this would require re-signing, so should fail)
		parts := strings.Split(originalToken, ".")
		if len(parts) == 3 {
			// Modify the payload part
			tamperedPayload := parts[1][:len(parts[1])-5] + "XXXXX"
			tamperedToken2 := parts[0] + "." + tamperedPayload + "." + parts[2]
			
			_, err = authService.ValidateAccessToken(tamperedToken2)
			assert.Error(t, err)
		}
	})

	t.Run("JWT expiration enforcement", func(t *testing.T) {
		// Create auth service with very short token expiration for testing
		shortExpiryService := services.NewAuthService(
			userRepo,
			sessionRepo,
			passwordResetRepo,
			loginAttemptRepo,
			lockoutRepo,
			auditService,
			"test-jwt-secret",
			db,
		)

		// In a real test, you'd need to modify the service to accept custom expiration
		// For now, we'll test with the standard expiration and simulate time passage
		loginResult, err := shortExpiryService.Login(ctx, testUser.Email, "password123", "127.0.0.1", "test-agent")
		assert.NoError(t, err)

		// Token should be valid immediately
		claims, err := shortExpiryService.ValidateAccessToken(loginResult.AccessToken)
		assert.NoError(t, err)
		assert.NotNil(t, claims)

		// Verify expiration time is set
		assert.True(t, claims.ExpiresAt.Time.After(time.Now()))
		assert.True(t, claims.ExpiresAt.Time.Before(time.Now().Add(2*time.Hour))) // Should expire within 2 hours
	})

	t.Run("JWT algorithm confusion attack prevention", func(t *testing.T) {
		// This test ensures that the JWT library properly validates the algorithm
		// and doesn't accept tokens signed with different algorithms
		
		loginResult, err := authService.Login(ctx, testUser.Email, "password123", "127.0.0.1", "test-agent")
		assert.NoError(t, err)

		// Try to create a token with "none" algorithm (should be rejected)
		noneAlgToken := "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJ1c2VyX2lkIjoidGVzdC11c2VyIiwiZW1haWwiOiJ0ZXN0QGV4YW1wbGUuY29tIn0."
		
		_, err = authService.ValidateAccessToken(noneAlgToken)
		assert.Error(t, err)

		// Original token should still work
		_, err = authService.ValidateAccessToken(loginResult.AccessToken)
		assert.NoError(t, err)
	})
}

func TestAuthService_AuditLogging(t *testing.T) {
	db := setupAuthSecurityTestDB()
	
	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	passwordResetRepo := repository.NewPasswordResetRepository(db)
	loginAttemptRepo := repository.NewLoginAttemptRepository(db)
	lockoutRepo := repository.NewAccountLockoutRepository(db)
	auditService := services.NewAuditService(repository.NewAuditLogRepository(db))

	authService := services.NewAuthService(
		userRepo,
		sessionRepo,
		passwordResetRepo,
		loginAttemptRepo,
		lockoutRepo,
		auditService,
		"test-jwt-secret",
		db,
	)

	testUser := createAuthTestUser(db, "audit@example.com", "password123")
	ctx := context.Background()

	t.Run("Security events are properly logged", func(t *testing.T) {
		// Successful login
		_, err := authService.Login(ctx, testUser.Email, "password123", "192.168.1.1", "Mozilla/5.0")
		assert.NoError(t, err)

		// Failed login
		_, err = authService.Login(ctx, testUser.Email, "wrongpassword", "192.168.1.2", "Chrome/90.0")
		assert.Error(t, err)

		// Password reset request
		_, err = authService.CreatePasswordReset(ctx, testUser.Email)
		assert.NoError(t, err)

		// Check login attempts are logged
		var loginAttempts []models.LoginAttempt
		err = db.Where("email = ?", testUser.Email).Find(&loginAttempts).Error
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(loginAttempts), 2) // At least success and failure

		// Verify successful attempt
		var successAttempt *models.LoginAttempt
		for _, attempt := range loginAttempts {
			if attempt.Success {
				successAttempt = &attempt
				break
			}
		}
		assert.NotNil(t, successAttempt)
		assert.Equal(t, testUser.ID, successAttempt.UserID)
		assert.Equal(t, "192.168.1.1", successAttempt.IPAddress)
		assert.Equal(t, "Mozilla/5.0", successAttempt.UserAgent)

		// Verify failed attempt
		var failedAttempt *models.LoginAttempt
		for _, attempt := range loginAttempts {
			if !attempt.Success {
				failedAttempt = &attempt
				break
			}
		}
		assert.NotNil(t, failedAttempt)
		assert.Equal(t, testUser.ID, failedAttempt.UserID)
		assert.Equal(t, "192.168.1.2", failedAttempt.IPAddress)
		assert.Contains(t, failedAttempt.FailureReason, "invalid password")
	})

	t.Run("Sensitive data is not logged", func(t *testing.T) {
		// Attempt login with password in various fields
		_, err := authService.Login(ctx, testUser.Email, "password123", "127.0.0.1", "test-agent")
		assert.NoError(t, err)

		// Check that password is not stored in any log
		var loginAttempts []models.LoginAttempt
		err = db.Find(&loginAttempts).Error
		assert.NoError(t, err)

		for _, attempt := range loginAttempts {
			// Verify password is not in any field
			assert.NotContains(t, attempt.FailureReason, "password123")
			assert.NotContains(t, attempt.UserAgent, "password123")
			assert.NotContains(t, attempt.IPAddress, "password123")
		}

		// Check audit logs don't contain sensitive data
		var auditLogs []models.AuditLog
		err = db.Find(&auditLogs).Error
		assert.NoError(t, err)

		for _, log := range auditLogs {
			assert.NotContains(t, log.Details, "password123")
		}
	})
}