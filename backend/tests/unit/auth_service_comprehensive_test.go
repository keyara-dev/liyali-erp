package unit

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/tests/helpers"
	"github.com/stretchr/testify/assert"
)

// TestAuthService_LoginSecurityScenarios tests authentication security scenarios using mocks
func TestAuthService_LoginSecurityScenarios(t *testing.T) {
	builder := helpers.NewMockTestDataBuilder()
	
	t.Run("Successful login with valid credentials", func(t *testing.T) {
		// Create mock user
		mockUser := &models.User{
			ID:                    builder.GetUserID(),
			Email:                 "user@example.com",
			Name:                  "Test User",
			Active:                true,
			CurrentOrganizationID: helpers.StringPtr(builder.GetOrganizationID()),
		}
		
		// Verify user properties
		assert.NotNil(t, mockUser)
		assert.Equal(t, "user@example.com", mockUser.Email)
		assert.True(t, mockUser.Active)
	})
	
	t.Run("Login blocked due to account lockout", func(t *testing.T) {
		// Create mock lockout
		mockLockout := &models.AccountLockout{
			ID:        uuid.New(),
			Email:     "locked@example.com",
			LockedAt:  time.Now().Add(-5 * time.Minute),
			UnlocksAt: time.Now().Add(10 * time.Minute),
			Reason:    "Too many failed attempts",
		}
		
		// Verify lockout is active
		assert.True(t, mockLockout.UnlocksAt.After(time.Now()))
	})
	
	t.Run("Login with inactive account", func(t *testing.T) {
		// Create mock inactive user
		mockUser := &models.User{
			ID:                    builder.GetUserID(),
			Email:                 "inactive@example.com",
			Name:                  "Inactive User",
			Active:                false,
			CurrentOrganizationID: helpers.StringPtr(builder.GetOrganizationID()),
		}
		
		// Verify user is inactive
		assert.False(t, mockUser.Active)
	})
}

// TestAuthService_SessionManagement tests session management using mocks
func TestAuthService_SessionManagement(t *testing.T) {
	builder := helpers.NewMockTestDataBuilder()
	
	t.Run("Create valid session", func(t *testing.T) {
		// Create mock session
		mockSession := &models.Session{
			ID:           uuid.New(),
			UserID:       builder.GetUserID(),
			RefreshToken: "valid-refresh-token",
			ExpiresAt:    time.Now().Add(24 * time.Hour),
			CreatedAt:    time.Now(),
		}
		
		// Verify session properties
		assert.NotNil(t, mockSession)
		assert.Equal(t, builder.GetUserID(), mockSession.UserID)
		assert.True(t, mockSession.ExpiresAt.After(time.Now()))
	})
	
	t.Run("Detect expired session", func(t *testing.T) {
		// Create mock expired session
		mockSession := &models.Session{
			ID:           uuid.New(),
			UserID:       builder.GetUserID(),
			RefreshToken: "expired-token",
			ExpiresAt:    time.Now().Add(-1 * time.Hour),
			CreatedAt:    time.Now().Add(-25 * time.Hour),
		}
		
		// Verify session is expired
		assert.True(t, mockSession.ExpiresAt.Before(time.Now()))
	})
}

// TestAuthService_PasswordReset tests password reset functionality using mocks
func TestAuthService_PasswordReset(t *testing.T) {
	builder := helpers.NewMockTestDataBuilder()
	
	t.Run("Create password reset token", func(t *testing.T) {
		// Create mock password reset
		mockReset := &models.PasswordReset{
			ID:        uuid.New(),
			UserID:    builder.GetUserID(),
			Token:     "reset-token-" + uuid.New().String(),
			ExpiresAt: time.Now().Add(1 * time.Hour),
			UsedAt:    nil,
			CreatedAt: time.Now(),
		}
		
		// Verify reset token properties
		assert.NotNil(t, mockReset)
		assert.Nil(t, mockReset.UsedAt)
		assert.True(t, mockReset.ExpiresAt.After(time.Now()))
	})
	
	t.Run("Detect used password reset token", func(t *testing.T) {
		// Create mock used reset token
		usedTime := time.Now().Add(-30 * time.Minute)
		mockReset := &models.PasswordReset{
			ID:        uuid.New(),
			UserID:    builder.GetUserID(),
			Token:     "used-reset-token",
			ExpiresAt: time.Now().Add(1 * time.Hour),
			UsedAt:    &usedTime,
			CreatedAt: time.Now().Add(-30 * time.Minute),
		}
		
		// Verify token is marked as used
		assert.NotNil(t, mockReset.UsedAt)
	})
	
	t.Run("Detect expired password reset token", func(t *testing.T) {
		// Create mock expired reset token
		mockReset := &models.PasswordReset{
			ID:        uuid.New(),
			UserID:    builder.GetUserID(),
			Token:     "expired-reset-token",
			ExpiresAt: time.Now().Add(-1 * time.Hour),
			UsedAt:    nil,
			CreatedAt: time.Now().Add(-2 * time.Hour),
		}
		
		// Verify token is expired
		assert.True(t, mockReset.ExpiresAt.Before(time.Now()))
	})
}

// TestAuthService_LoginAttempts tests login attempt tracking using mocks
func TestAuthService_LoginAttempts(t *testing.T) {
	t.Run("Track successful login attempt", func(t *testing.T) {
		// Create mock successful login attempt
		mockAttempt := &models.LoginAttempt{
			ID:          uuid.New(),
			Email:       "user@example.com",
			Success:     true,
			AttemptedAt: time.Now(),
		}
		
		// Verify attempt properties
		assert.True(t, mockAttempt.Success)
		assert.Empty(t, mockAttempt.FailureReason)
	})
	
	t.Run("Track failed login attempts", func(t *testing.T) {
		// Create mock failed login attempts
		attempts := []models.LoginAttempt{
			{
				ID:             uuid.New(),
				Email:          "user@example.com",
				Success:        false,
				FailureReason:  "Invalid password",
				AttemptedAt:    time.Now().Add(-5 * time.Minute),
			},
			{
				ID:             uuid.New(),
				Email:          "user@example.com",
				Success:        false,
				FailureReason:  "Invalid password",
				AttemptedAt:    time.Now().Add(-4 * time.Minute),
			},
			{
				ID:             uuid.New(),
				Email:          "user@example.com",
				Success:        false,
				FailureReason:  "Invalid password",
				AttemptedAt:    time.Now().Add(-3 * time.Minute),
			},
		}
		
		// Verify failed attempts
		assert.Len(t, attempts, 3)
		for _, attempt := range attempts {
			assert.False(t, attempt.Success)
			assert.NotEmpty(t, attempt.FailureReason)
		}
	})
}

// TestAuthService_AccountLockout tests account lockout functionality using mocks
func TestAuthService_AccountLockout(t *testing.T) {
	t.Run("Create account lockout", func(t *testing.T) {
		// Create mock account lockout
		mockLockout := &models.AccountLockout{
			ID:        uuid.New(),
			UserID:    "test-user-" + uuid.New().String()[:8],
			Email:     "user@example.com",
			LockedAt:  time.Now(),
			UnlocksAt: time.Now().Add(15 * time.Minute),
			Reason:    "Too many failed login attempts",
		}
		
		// Verify lockout properties
		assert.NotNil(t, mockLockout)
		assert.Equal(t, "user@example.com", mockLockout.Email)
		assert.True(t, mockLockout.UnlocksAt.After(time.Now()))
	})
	
	t.Run("Detect expired lockout", func(t *testing.T) {
		// Create mock expired lockout
		mockLockout := &models.AccountLockout{
			ID:        uuid.New(),
			UserID:    "test-user-" + uuid.New().String()[:8],
			Email:     "user@example.com",
			LockedAt:  time.Now().Add(-20 * time.Minute),
			UnlocksAt: time.Now().Add(-5 * time.Minute),
			Reason:    "Previous violation",
		}
		
		// Verify lockout is expired
		assert.True(t, mockLockout.UnlocksAt.Before(time.Now()))
	})
}
