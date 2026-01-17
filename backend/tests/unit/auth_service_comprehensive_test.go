package unit

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/services"
	"github.com/liyali/liyali-gateway/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// Comprehensive Authentication Service Tests
// These tests cover critical security scenarios and edge cases

// Mock repositories for testing
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) UpdateLastLogin(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserRepository) UpdatePassword(ctx context.Context, userID, hashedPassword string) error {
	args := m.Called(ctx, userID, hashedPassword)
	return args.Error(0)
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) (*models.User, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, limit, offset int) ([]*models.User, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserRepository) ListByOrganization(ctx context.Context, organizationID string, limit, offset int) ([]*models.User, error) {
	args := m.Called(ctx, organizationID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserRepository) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRepository) CountActive(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRepository) Activate(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) Deactivate(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockSessionRepository struct {
	mock.Mock
}

func (m *MockSessionRepository) Create(ctx context.Context, userID, refreshToken, ipAddress, userAgent string, expiresAt time.Time) (*models.Session, error) {
	args := m.Called(ctx, userID, refreshToken, ipAddress, userAgent, expiresAt)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Session), args.Error(1)
}

func (m *MockSessionRepository) GetByRefreshToken(ctx context.Context, refreshToken string) (*models.Session, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Session), args.Error(1)
}

func (m *MockSessionRepository) GetByUserID(ctx context.Context, userID string) ([]*models.Session, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Session), args.Error(1)
}

func (m *MockSessionRepository) UpdateRefreshToken(ctx context.Context, id uuid.UUID, oldRefreshToken, newRefreshToken string, expiresAt time.Time) (int64, error) {
	args := m.Called(ctx, id, oldRefreshToken, newRefreshToken, expiresAt)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockSessionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSessionRepository) DeleteByRefreshToken(ctx context.Context, refreshToken string) error {
	args := m.Called(ctx, refreshToken)
	return args.Error(0)
}

func (m *MockSessionRepository) DeleteByUserID(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockSessionRepository) DeleteExpired(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockSessionRepository) CountActive(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockSessionRepository) CountUserActive(ctx context.Context, userID string) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

type MockLoginAttemptRepository struct {
	mock.Mock
}

func (m *MockLoginAttemptRepository) Create(ctx context.Context, userID, email, ipAddress, userAgent string, success bool, failureReason string) (*models.LoginAttempt, error) {
	args := m.Called(ctx, userID, email, ipAddress, userAgent, success, failureReason)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LoginAttempt), args.Error(1)
}

func (m *MockLoginAttemptRepository) GetRecentFailedAttempts(ctx context.Context, email string, since time.Time) (int64, error) {
	args := m.Called(ctx, email, since)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockLoginAttemptRepository) GetRecentFailedAttemptsByIP(ctx context.Context, ipAddress string, since time.Time) (int64, error) {
	args := m.Called(ctx, ipAddress, since)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockLoginAttemptRepository) GetByUser(ctx context.Context, userID string, limit, offset int) ([]*models.LoginAttempt, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.LoginAttempt), args.Error(1)
}

func (m *MockLoginAttemptRepository) GetByEmail(ctx context.Context, email string, limit, offset int) ([]*models.LoginAttempt, error) {
	args := m.Called(ctx, email, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.LoginAttempt), args.Error(1)
}

func (m *MockLoginAttemptRepository) DeleteOld(ctx context.Context, before time.Time) error {
	args := m.Called(ctx, before)
	return args.Error(0)
}

type MockAccountLockoutRepository struct {
	mock.Mock
}

func (m *MockAccountLockoutRepository) Create(ctx context.Context, userID, email, ipAddress, reason string, unlocksAt time.Time) (*models.AccountLockout, error) {
	args := m.Called(ctx, userID, email, ipAddress, reason, unlocksAt)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AccountLockout), args.Error(1)
}

func (m *MockAccountLockoutRepository) GetActiveByUserID(ctx context.Context, userID string) (*models.AccountLockout, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AccountLockout), args.Error(1)
}

func (m *MockAccountLockoutRepository) GetActiveByEmail(ctx context.Context, email string) (*models.AccountLockout, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AccountLockout), args.Error(1)
}

func (m *MockAccountLockoutRepository) Unlock(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockAccountLockoutRepository) UnlockByEmail(ctx context.Context, email string) error {
	args := m.Called(ctx, email)
	return args.Error(0)
}

func (m *MockAccountLockoutRepository) GetHistory(ctx context.Context, userID string, limit, offset int) ([]*models.AccountLockout, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.AccountLockout), args.Error(1)
}

func (m *MockAccountLockoutRepository) CleanupExpired(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

type MockPasswordResetRepository struct {
	mock.Mock
}

func (m *MockPasswordResetRepository) Create(ctx context.Context, userID, token string, expiresAt time.Time) (*models.PasswordReset, error) {
	args := m.Called(ctx, userID, token, expiresAt)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PasswordReset), args.Error(1)
}

func (m *MockPasswordResetRepository) GetByToken(ctx context.Context, token string) (*models.PasswordReset, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PasswordReset), args.Error(1)
}

func (m *MockPasswordResetRepository) MarkAsUsed(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPasswordResetRepository) DeleteByUserID(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockPasswordResetRepository) DeleteExpired(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockPasswordResetRepository) DeleteUsed(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestAuthService_Login_SecurityScenarios(t *testing.T) {
	tests := []struct {
		name           string
		email          string
		password       string
		setupMocks     func(*MockUserRepository, *MockSessionRepository, *MockLoginAttemptRepository, *MockAccountLockoutRepository)
		expectedError  error
		expectedResult bool
	}{
		{
			name:     "Successful login with valid credentials",
			email:    "user@example.com",
			password: "validpassword",
			setupMocks: func(userRepo *MockUserRepository, sessionRepo *MockSessionRepository, loginRepo *MockLoginAttemptRepository, lockoutRepo *MockAccountLockoutRepository) {
				user := &models.User{
					ID:       "user-123",
					Email:    "user@example.com",
					Password: utils.HashPassword("validpassword"),
					Active:   true,
				}
				
				// Mock successful user lookup
				userRepo.On("GetByEmail", mock.Anything, "user@example.com").Return(user, nil)
				
				// Mock no recent failed attempts
				loginRepo.On("GetRecentFailedAttempts", mock.Anything, "user@example.com", mock.Anything).Return(int64(0), nil)
				
				// Mock no active lockout
				lockoutRepo.On("GetActiveByUserID", mock.Anything, "user-123").Return(nil, gorm.ErrRecordNotFound)
				
				// Mock successful session creation
				session := &models.Session{
					ID:     uuid.New(),
					UserID: "user-123",
				}
				sessionRepo.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(session, nil)
				
				// Mock successful login attempt recording
				loginRepo.On("Create", mock.Anything, "user-123", "user@example.com", mock.Anything, mock.Anything, true, "").Return(nil, nil)
				
				// Mock last login update
				userRepo.On("UpdateLastLogin", mock.Anything, "user-123").Return(nil)
			},
			expectedError:  nil,
			expectedResult: true,
		},
		{
			name:     "Login blocked due to too many failed attempts",
			email:    "blocked@example.com",
			password: "anypassword",
			setupMocks: func(userRepo *MockUserRepository, sessionRepo *MockSessionRepository, loginRepo *MockLoginAttemptRepository, lockoutRepo *MockAccountLockoutRepository) {
				// Mock too many recent failed attempts
				loginRepo.On("GetRecentFailedAttempts", mock.Anything, "blocked@example.com", mock.Anything).Return(int64(6), nil)
				
				// Mock failed login attempt recording
				loginRepo.On("Create", mock.Anything, "", "blocked@example.com", mock.Anything, mock.Anything, false, "too many failed attempts").Return(nil, nil)
			},
			expectedError:  services.ErrTooManyFailedAttempts,
			expectedResult: false,
		},
		{
			name:     "Login with invalid password triggers lockout",
			email:    "user@example.com",
			password: "wrongpassword",
			setupMocks: func(userRepo *MockUserRepository, sessionRepo *MockSessionRepository, loginRepo *MockLoginAttemptRepository, lockoutRepo *MockAccountLockoutRepository) {
				user := &models.User{
					ID:       "user-123",
					Email:    "user@example.com",
					Password: utils.HashPassword("correctpassword"),
					Active:   true,
				}
				
				// Mock user lookup
				userRepo.On("GetByEmail", mock.Anything, "user@example.com").Return(user, nil)
				
				// Mock no recent failed attempts initially
				loginRepo.On("GetRecentFailedAttempts", mock.Anything, "user@example.com", mock.Anything).Return(int64(0), nil)
				
				// Mock no active lockout
				lockoutRepo.On("GetActiveByUserID", mock.Anything, "user-123").Return(nil, gorm.ErrRecordNotFound)
				
				// Mock failed login attempt recording
				loginRepo.On("Create", mock.Anything, "user-123", "user@example.com", mock.Anything, mock.Anything, false, "invalid password").Return(nil, nil)
				
				// Mock checking for lockout threshold (4 previous attempts + this one = 5)
				loginRepo.On("GetRecentFailedAttempts", mock.Anything, "user@example.com", mock.Anything).Return(int64(4), nil)
				
				// Mock account lockout creation
				lockoutRepo.On("Create", mock.Anything, "user-123", "user@example.com", mock.Anything, "too many failed login attempts", mock.Anything).Return(nil, nil)
			},
			expectedError:  services.ErrInvalidCredentials,
			expectedResult: false,
		},
		{
			name:     "Login with inactive account",
			email:    "inactive@example.com",
			password: "validpassword",
			setupMocks: func(userRepo *MockUserRepository, sessionRepo *MockSessionRepository, loginRepo *MockLoginAttemptRepository, lockoutRepo *MockAccountLockoutRepository) {
				user := &models.User{
					ID:       "user-123",
					Email:    "inactive@example.com",
					Password: utils.HashPassword("validpassword"),
					Active:   false, // Inactive account
				}
				
				// Mock user lookup
				userRepo.On("GetByEmail", mock.Anything, "inactive@example.com").Return(user, nil)
				
				// Mock no recent failed attempts
				loginRepo.On("GetRecentFailedAttempts", mock.Anything, "inactive@example.com", mock.Anything).Return(int64(0), nil)
				
				// Mock no active lockout
				lockoutRepo.On("GetActiveByUserID", mock.Anything, "user-123").Return(nil, gorm.ErrRecordNotFound)
				
				// Mock failed login attempt recording
				loginRepo.On("Create", mock.Anything, "user-123", "inactive@example.com", mock.Anything, mock.Anything, false, "account inactive").Return(nil, nil)
			},
			expectedError:  services.ErrAccountInactive,
			expectedResult: false,
		},
		{
			name:     "Login with locked account",
			email:    "locked@example.com",
			password: "validpassword",
			setupMocks: func(userRepo *MockUserRepository, sessionRepo *MockSessionRepository, loginRepo *MockLoginAttemptRepository, lockoutRepo *MockAccountLockoutRepository) {
				user := &models.User{
					ID:       "user-123",
					Email:    "locked@example.com",
					Password: utils.HashPassword("validpassword"),
					Active:   true,
				}
				
				lockout := &models.AccountLockout{
					ID:        uuid.New(),
					UserID:    "user-123",
					UnlocksAt: time.Now().Add(10 * time.Minute),
				}
				
				// Mock user lookup
				userRepo.On("GetByEmail", mock.Anything, "locked@example.com").Return(user, nil)
				
				// Mock no recent failed attempts
				loginRepo.On("GetRecentFailedAttempts", mock.Anything, "locked@example.com", mock.Anything).Return(int64(0), nil)
				
				// Mock active lockout
				lockoutRepo.On("GetActiveByUserID", mock.Anything, "user-123").Return(lockout, nil)
				
				// Mock failed login attempt recording
				loginRepo.On("Create", mock.Anything, "user-123", "locked@example.com", mock.Anything, mock.Anything, false, "account locked").Return(nil, nil)
			},
			expectedError:  services.ErrAccountLocked,
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			userRepo := &MockUserRepository{}
			sessionRepo := &MockSessionRepository{}
			passwordResetRepo := &MockPasswordResetRepository{}
			loginAttemptRepo := &MockLoginAttemptRepository{}
			lockoutRepo := &MockAccountLockoutRepository{}
			auditService := &services.AuditService{}

			tt.setupMocks(userRepo, sessionRepo, loginAttemptRepo, lockoutRepo)

			// Create auth service
			authService := services.NewAuthService(
				userRepo,
				sessionRepo,
				passwordResetRepo,
				loginAttemptRepo,
				lockoutRepo,
				auditService,
				"test-jwt-secret",
				nil, // db not needed for this test
			)

			// Execute login
			result, err := authService.Login(context.Background(), tt.email, tt.password, "127.0.0.1", "test-agent")

			// Verify results
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				if tt.expectedResult {
					assert.NotNil(t, result)
					assert.NotEmpty(t, result.AccessToken)
					assert.NotEmpty(t, result.RefreshToken)
				}
			}

			// Verify all mocks were called as expected
			userRepo.AssertExpectations(t)
			sessionRepo.AssertExpectations(t)
			loginAttemptRepo.AssertExpectations(t)
			lockoutRepo.AssertExpectations(t)
		})
	}
}

func TestAuthService_RefreshToken_SecurityScenarios(t *testing.T) {
	tests := []struct {
		name           string
		refreshToken   string
		setupMocks     func(*MockUserRepository, *MockSessionRepository)
		expectedError  error
		expectedResult bool
	}{
		{
			name:         "Successful token refresh",
			refreshToken: "valid-refresh-token",
			setupMocks: func(userRepo *MockUserRepository, sessionRepo *MockSessionRepository) {
				session := &models.Session{
					ID:           uuid.New(),
					UserID:       "user-123",
					RefreshToken: "valid-refresh-token",
					ExpiresAt:    time.Now().Add(24 * time.Hour),
				}
				
				user := &models.User{
					ID:     "user-123",
					Email:  "user@example.com",
					Active: true,
				}
				
				// Mock session lookup
				sessionRepo.On("GetByRefreshToken", mock.Anything, "valid-refresh-token").Return(session, nil)
				
				// Mock user lookup
				userRepo.On("GetByID", mock.Anything, "user-123").Return(user, nil)
				
				// Mock token rotation
				sessionRepo.On("UpdateRefreshToken", mock.Anything, session.ID, "valid-refresh-token", mock.Anything, mock.Anything).Return(int64(1), nil)
			},
			expectedError:  nil,
			expectedResult: true,
		},
		{
			name:         "Token reuse detection",
			refreshToken: "reused-token",
			setupMocks: func(userRepo *MockUserRepository, sessionRepo *MockSessionRepository) {
				session := &models.Session{
					ID:           uuid.New(),
					UserID:       "user-123",
					RefreshToken: "reused-token",
					ExpiresAt:    time.Now().Add(24 * time.Hour),
				}
				
				user := &models.User{
					ID:     "user-123",
					Email:  "user@example.com",
					Active: true,
				}
				
				// Mock session lookup
				sessionRepo.On("GetByRefreshToken", mock.Anything, "reused-token").Return(session, nil)
				
				// Mock user lookup
				userRepo.On("GetByID", mock.Anything, "user-123").Return(user, nil)
				
				// Mock token rotation failure (0 rows affected = token already used)
				sessionRepo.On("UpdateRefreshToken", mock.Anything, session.ID, "reused-token", mock.Anything, mock.Anything).Return(int64(0), nil)
			},
			expectedError:  services.ErrTokenReuseDetected,
			expectedResult: false,
		},
		{
			name:         "Expired session",
			refreshToken: "expired-token",
			setupMocks: func(userRepo *MockUserRepository, sessionRepo *MockSessionRepository) {
				session := &models.Session{
					ID:           uuid.New(),
					UserID:       "user-123",
					RefreshToken: "expired-token",
					ExpiresAt:    time.Now().Add(-1 * time.Hour), // Expired
				}
				
				// Mock session lookup
				sessionRepo.On("GetByRefreshToken", mock.Anything, "expired-token").Return(session, nil)
				
				// Mock session deletion
				sessionRepo.On("Delete", mock.Anything, session.ID).Return(nil)
			},
			expectedError:  services.ErrSessionExpired,
			expectedResult: false,
		},
		{
			name:         "Invalid refresh token",
			refreshToken: "invalid-token",
			setupMocks: func(userRepo *MockUserRepository, sessionRepo *MockSessionRepository) {
				// Mock session not found
				sessionRepo.On("GetByRefreshToken", mock.Anything, "invalid-token").Return(nil, gorm.ErrRecordNotFound)
			},
			expectedError:  services.ErrInvalidRefreshToken,
			expectedResult: false,
		},
		{
			name:         "User deactivated during session",
			refreshToken: "valid-token-inactive-user",
			setupMocks: func(userRepo *MockUserRepository, sessionRepo *MockSessionRepository) {
				session := &models.Session{
					ID:           uuid.New(),
					UserID:       "user-123",
					RefreshToken: "valid-token-inactive-user",
					ExpiresAt:    time.Now().Add(24 * time.Hour),
				}
				
				user := &models.User{
					ID:     "user-123",
					Email:  "user@example.com",
					Active: false, // User deactivated
				}
				
				// Mock session lookup
				sessionRepo.On("GetByRefreshToken", mock.Anything, "valid-token-inactive-user").Return(session, nil)
				
				// Mock user lookup
				userRepo.On("GetByID", mock.Anything, "user-123").Return(user, nil)
				
				// Mock session deletion
				sessionRepo.On("Delete", mock.Anything, session.ID).Return(nil)
			},
			expectedError:  services.ErrAccountInactive,
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			userRepo := &MockUserRepository{}
			sessionRepo := &MockSessionRepository{}
			passwordResetRepo := &MockPasswordResetRepository{}
			loginAttemptRepo := &MockLoginAttemptRepository{}
			lockoutRepo := &MockAccountLockoutRepository{}
			auditService := &services.AuditService{}

			tt.setupMocks(userRepo, sessionRepo)

			// Create auth service
			authService := services.NewAuthService(
				userRepo,
				sessionRepo,
				passwordResetRepo,
				loginAttemptRepo,
				lockoutRepo,
				auditService,
				"test-jwt-secret",
				nil,
			)

			// Execute token refresh
			result, err := authService.RefreshToken(context.Background(), tt.refreshToken)

			// Verify results
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				if tt.expectedResult {
					assert.NotNil(t, result)
					assert.NotEmpty(t, result.AccessToken)
					assert.NotEmpty(t, result.RefreshToken)
				}
			}

			// Verify mocks
			userRepo.AssertExpectations(t)
			sessionRepo.AssertExpectations(t)
		})
	}
}

func TestAuthService_PasswordReset_SecurityScenarios(t *testing.T) {
	tests := []struct {
		name          string
		email         string
		setupMocks    func(*MockUserRepository, *MockPasswordResetRepository)
		expectedError error
		expectToken   bool
	}{
		{
			name:  "Successful password reset request",
			email: "user@example.com",
			setupMocks: func(userRepo *MockUserRepository, resetRepo *MockPasswordResetRepository) {
				user := &models.User{
					ID:    "user-123",
					Email: "user@example.com",
				}
				
				// Mock user lookup
				userRepo.On("GetByEmail", mock.Anything, "user@example.com").Return(user, nil)
				
				// Mock reset token creation
				resetRepo.On("Create", mock.Anything, "user-123", mock.Anything, mock.Anything).Return(nil, nil)
			},
			expectedError: nil,
			expectToken:   true,
		},
		{
			name:  "Password reset for non-existent user",
			email: "nonexistent@example.com",
			setupMocks: func(userRepo *MockUserRepository, resetRepo *MockPasswordResetRepository) {
				// Mock user not found
				userRepo.On("GetByEmail", mock.Anything, "nonexistent@example.com").Return(nil, gorm.ErrRecordNotFound)
			},
			expectedError: services.ErrUserNotFound,
			expectToken:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			userRepo := &MockUserRepository{}
			sessionRepo := &MockSessionRepository{}
			passwordResetRepo := &MockPasswordResetRepository{}
			loginAttemptRepo := &MockLoginAttemptRepository{}
			lockoutRepo := &MockAccountLockoutRepository{}
			auditService := &services.AuditService{}

			tt.setupMocks(userRepo, passwordResetRepo)

			// Create auth service
			authService := services.NewAuthService(
				userRepo,
				sessionRepo,
				passwordResetRepo,
				loginAttemptRepo,
				lockoutRepo,
				auditService,
				"test-jwt-secret",
				nil,
			)

			// Execute password reset request
			token, err := authService.CreatePasswordReset(context.Background(), tt.email)

			// Verify results
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				if tt.expectToken {
					assert.NotEmpty(t, token)
				}
			}

			// Verify mocks
			userRepo.AssertExpectations(t)
			passwordResetRepo.AssertExpectations(t)
		})
	}
}

func TestAuthService_JWTValidation_SecurityScenarios(t *testing.T) {
	authService := services.NewAuthService(
		nil, nil, nil, nil, nil, nil,
		"test-jwt-secret-for-validation",
		nil,
	)

	t.Run("Valid JWT token", func(t *testing.T) {
		// Create a valid user for token generation
		user := &models.User{
			ID:    "user-123",
			Email: "user@example.com",
			Name:  "Test User",
			Role:  "user",
		}

		// We can't directly test generateAccessToken since it's private
		// Instead, we'll test through the login flow
		userRepo := &MockUserRepository{}
		sessionRepo := &MockSessionRepository{}
		passwordResetRepo := &MockPasswordResetRepository{}
		loginAttemptRepo := &MockLoginAttemptRepository{}
		lockoutRepo := &MockAccountLockoutRepository{}
		auditService := &services.AuditService{}

		// Setup mocks for successful login
		hashedPassword, _ := utils.HashPassword("testpassword")
		user.Password = hashedPassword
		
		userRepo.On("GetByEmail", mock.Anything, "user@example.com").Return(user, nil)
		loginAttemptRepo.On("GetRecentFailedAttempts", mock.Anything, "user@example.com", mock.Anything).Return(int64(0), nil)
		lockoutRepo.On("GetActiveByUserID", mock.Anything, "user-123").Return(nil, gorm.ErrRecordNotFound)
		
		session := &models.Session{
			ID:     uuid.New(),
			UserID: "user-123",
		}
		sessionRepo.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(session, nil)
		loginAttemptRepo.On("Create", mock.Anything, "user-123", "user@example.com", mock.Anything, mock.Anything, true, "").Return(nil, nil)
		userRepo.On("UpdateLastLogin", mock.Anything, "user-123").Return(nil)

		authService := services.NewAuthService(
			userRepo, sessionRepo, passwordResetRepo, loginAttemptRepo, lockoutRepo,
			auditService, "test-jwt-secret-for-validation", nil,
		)

		// Login to get a valid token
		loginResult, err := authService.Login(context.Background(), "user@example.com", "testpassword", "127.0.0.1", "test-agent")
		assert.NoError(t, err)
		assert.NotEmpty(t, loginResult.AccessToken)

		// Validate the token
		claims, err := authService.ValidateAccessToken(loginResult.AccessToken)
		assert.NoError(t, err)
		assert.NotNil(t, claims)
		assert.Equal(t, "user-123", claims.UserID)
		assert.Equal(t, "user@example.com", claims.Email)
	})

	t.Run("Invalid JWT signature", func(t *testing.T) {
		// Create token with different secret
		userRepo2 := &MockUserRepository{}
		sessionRepo2 := &MockSessionRepository{}
		passwordResetRepo2 := &MockPasswordResetRepository{}
		loginAttemptRepo2 := &MockLoginAttemptRepository{}
		lockoutRepo2 := &MockAccountLockoutRepository{}
		auditService2 := &services.AuditService{}

		differentSecretService := services.NewAuthService(
			userRepo2, sessionRepo2, passwordResetRepo2, loginAttemptRepo2, lockoutRepo2,
			auditService2, "different-secret", nil,
		)

		// Create a token with different secret (we'll use a hardcoded invalid token)
		invalidToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoidXNlci0xMjMiLCJlbWFpbCI6InVzZXJAZXhhbXBsZS5jb20ifQ.invalid_signature"

		// Try to validate with original service (different secret)
		claims, err := authService.ValidateAccessToken(invalidToken)
		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("Malformed JWT token", func(t *testing.T) {
		claims, err := authService.ValidateAccessToken("invalid.jwt.token")
		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("Empty JWT token", func(t *testing.T) {
		claims, err := authService.ValidateAccessToken("")
		assert.Error(t, err)
		assert.Nil(t, claims)
	})
}

func TestAuthService_ConcurrentSessions(t *testing.T) {
	// Test concurrent session management
	userRepo := &MockUserRepository{}
	sessionRepo := &MockSessionRepository{}
	passwordResetRepo := &MockPasswordResetRepository{}
	loginAttemptRepo := &MockLoginAttemptRepository{}
	lockoutRepo := &MockAccountLockoutRepository{}
	auditService := &services.AuditService{}

	authService := services.NewAuthService(
		userRepo,
		sessionRepo,
		passwordResetRepo,
		loginAttemptRepo,
		lockoutRepo,
		auditService,
		"test-jwt-secret",
		nil,
	)

	t.Run("Session cleanup for max sessions", func(t *testing.T) {
		userID := "user-123"
		
		// Mock getting existing sessions (more than max allowed)
		existingSessions := make([]*models.Session, 7) // More than MaxSessionsPerUser (5)
		for i := 0; i < 7; i++ {
			existingSessions[i] = &models.Session{
				ID:        uuid.New(),
				UserID:    userID,
				CreatedAt: time.Now().Add(-time.Duration(i) * time.Hour),
			}
		}
		
		sessionRepo.On("GetByUserID", mock.Anything, userID).Return(existingSessions, nil)
		
		// Mock deletion of old sessions (should delete 2 oldest)
		sessionRepo.On("Delete", mock.Anything, existingSessions[5].ID).Return(nil)
		sessionRepo.On("Delete", mock.Anything, existingSessions[6].ID).Return(nil)

		// Test session cleanup
		authService.CleanupOldSessions(context.Background(), userID)

		sessionRepo.AssertExpectations(t)
	})
}