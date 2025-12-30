package unit

import (
	"context"
	"testing"
	"time"

	"github.com/cozyCodr/liyali-gateway/internal/db"
	"github.com/cozyCodr/liyali-gateway/internal/services"
	"github.com/cozyCodr/liyali-gateway/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// Mock UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(ctx context.Context, params db.CreateUserParams) (*db.User, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*db.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByEmail(ctx context.Context, email string) (*db.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*db.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*db.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*db.User), args.Error(1)
}

func (m *MockUserRepository) UpdateUser(ctx context.Context, params db.UpdateUserParams) (*db.User, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*db.User), args.Error(1)
}

func (m *MockUserRepository) UpdateUserPassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	args := m.Called(ctx, id, passwordHash)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateUserLastLogin(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) IncrementFailedLoginAttempts(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) ResetFailedLoginAttempts(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) LockUserAccount(ctx context.Context, id uuid.UUID, lockedUntil time.Time) error {
	args := m.Called(ctx, id, mock.Anything)
	return args.Error(0)
}

func (m *MockUserRepository) DeactivateUser(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) ActivateUser(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) ListUsers(ctx context.Context, limit, offset int32) ([]db.User, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]db.User), args.Error(1)
}

func (m *MockUserRepository) ListUsersByRole(ctx context.Context, role string) ([]db.User, error) {
	args := m.Called(ctx, role)
	return args.Get(0).([]db.User), args.Error(1)
}

func (m *MockUserRepository) ListUsersByDepartment(ctx context.Context, department string) ([]db.User, error) {
	args := m.Called(ctx, department)
	return args.Get(0).([]db.User), args.Error(1)
}

func (m *MockUserRepository) CountUsers(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRepository) CountActiveUsers(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

// Mock SessionRepository
type MockSessionRepository struct {
	mock.Mock
}

func (m *MockSessionRepository) CreateSession(ctx context.Context, userID uuid.UUID, refreshToken, ipAddress, userAgent string, expiresAt time.Time) (*db.Session, error) {
	args := m.Called(ctx, userID, refreshToken, ipAddress, userAgent, mock.Anything)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*db.Session), args.Error(1)
}

func (m *MockSessionRepository) GetSessionByRefreshToken(ctx context.Context, refreshToken string) (*db.Session, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*db.Session), args.Error(1)
}

func (m *MockSessionRepository) GetSessionsByUserID(ctx context.Context, userID uuid.UUID) ([]db.Session, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]db.Session), args.Error(1)
}

func (m *MockSessionRepository) DeleteSession(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSessionRepository) DeleteSessionByRefreshToken(ctx context.Context, refreshToken string) error {
	args := m.Called(ctx, refreshToken)
	return args.Error(0)
}

func (m *MockSessionRepository) DeleteSessionsByUserID(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockSessionRepository) DeleteExpiredSessions(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockSessionRepository) CountActiveSessions(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockSessionRepository) CountUserActiveSessions(ctx context.Context, userID uuid.UUID) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

// Mock PasswordResetRepository
type MockPasswordResetRepository struct {
	mock.Mock
}

func (m *MockPasswordResetRepository) CreatePasswordReset(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) (*db.PasswordReset, error) {
	args := m.Called(ctx, userID, token, mock.Anything)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*db.PasswordReset), args.Error(1)
}

func (m *MockPasswordResetRepository) GetPasswordResetByToken(ctx context.Context, token string) (*db.PasswordReset, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*db.PasswordReset), args.Error(1)
}

func (m *MockPasswordResetRepository) MarkPasswordResetAsUsed(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPasswordResetRepository) DeletePasswordResetsByUserID(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockPasswordResetRepository) DeleteExpiredPasswordResets(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockPasswordResetRepository) DeleteUsedPasswordResets(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Tests

func TestRegisterUser_Success(t *testing.T) {
	userRepo := new(MockUserRepository)
	sessionRepo := new(MockSessionRepository)
	passwordResetRepo := new(MockPasswordResetRepository)

	authService := services.NewAuthService(userRepo, sessionRepo, passwordResetRepo, "test-secret")

	ctx := context.Background()
	email := "test@example.com"
	password := "password123"
	name := "Test User"
	role := "REQUESTER"
	department := "Engineering"

	// Mock: user doesn't exist
	userRepo.On("GetUserByEmail", ctx, email).Return(nil, assert.AnError)

	// Mock: create user succeeds
	userID := uuid.New()
	createdUser := &db.User{
		ID:       utils.UUIDToPgtype(userID),
		Email:    email,
		Name:     name,
		Role:     role,
		IsActive: utils.BoolToPgtype(true),
	}
	userRepo.On("CreateUser", ctx, mock.Anything).Return(createdUser, nil)

	// Execute
	user, err := authService.RegisterUser(ctx, email, password, name, role, department)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, email, user.Email)
	assert.Equal(t, name, user.Name)
	assert.Equal(t, role, user.Role)
	userRepo.AssertExpectations(t)
}

func TestRegisterUser_EmailExists(t *testing.T) {
	userRepo := new(MockUserRepository)
	sessionRepo := new(MockSessionRepository)
	passwordResetRepo := new(MockPasswordResetRepository)

	authService := services.NewAuthService(userRepo, sessionRepo, passwordResetRepo, "test-secret")

	ctx := context.Background()
	email := "test@example.com"

	// Mock: user already exists
	existingUser := &db.User{Email: email}
	userRepo.On("GetUserByEmail", ctx, email).Return(existingUser, nil)

	// Execute
	user, err := authService.RegisterUser(ctx, email, "password", "Test", "REQUESTER", "Eng")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, services.ErrEmailAlreadyExists, err)
	assert.Nil(t, user)
	userRepo.AssertExpectations(t)
}

func TestLogin_Success(t *testing.T) {
	userRepo := new(MockUserRepository)
	sessionRepo := new(MockSessionRepository)
	passwordResetRepo := new(MockPasswordResetRepository)

	authService := services.NewAuthService(userRepo, sessionRepo, passwordResetRepo, "test-secret")

	ctx := context.Background()
	email := "test@example.com"
	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	userID := uuid.New()
	mockUser := &db.User{
		ID:           utils.UUIDToPgtype(userID),
		Email:        email,
		PasswordHash: string(hashedPassword),
		Name:         "Test User",
		Role:         "REQUESTER",
		IsActive:     utils.BoolToPgtype(true),
		LockedUntil: pgtype.Timestamp{
			Valid: false,
		},
		FailedLoginAttempts: utils.Int32ToPgtype(0),
	}

	// Mocks
	userRepo.On("GetUserByEmail", ctx, email).Return(mockUser, nil)
	userRepo.On("ResetFailedLoginAttempts", ctx, userID).Return(nil)
	userRepo.On("UpdateUserLastLogin", ctx, userID).Return(nil)
	sessionRepo.On("CreateSession", ctx, userID, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&db.Session{}, nil)

	// Execute
	accessToken, refreshToken, user, err := authService.Login(ctx, email, password, "127.0.0.1", "test-agent")

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)
	assert.NotNil(t, user)
	assert.Equal(t, email, user.Email)
}

func TestLogin_InvalidPassword(t *testing.T) {
	userRepo := new(MockUserRepository)
	sessionRepo := new(MockSessionRepository)
	passwordResetRepo := new(MockPasswordResetRepository)

	authService := services.NewAuthService(userRepo, sessionRepo, passwordResetRepo, "test-secret")

	ctx := context.Background()
	email := "test@example.com"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.DefaultCost)

	userID := uuid.New()
	mockUser := &db.User{
		ID:           utils.UUIDToPgtype(userID),
		Email:        email,
		PasswordHash: string(hashedPassword),
		IsActive:     utils.BoolToPgtype(true),
		LockedUntil: pgtype.Timestamp{
			Valid: false,
		},
		FailedLoginAttempts: utils.Int32ToPgtype(0),
	}

	// Mocks
	userRepo.On("GetUserByEmail", ctx, email).Return(mockUser, nil)
	userRepo.On("IncrementFailedLoginAttempts", ctx, userID).Return(nil)

	// Execute
	_, _, _, err := authService.Login(ctx, email, "wrong-password", "127.0.0.1", "test-agent")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, services.ErrInvalidCredentials, err)
	userRepo.AssertExpectations(t)
}

func TestGenerateAndValidateAccessToken(t *testing.T) {
	userRepo := new(MockUserRepository)
	sessionRepo := new(MockSessionRepository)
	passwordResetRepo := new(MockPasswordResetRepository)

	authService := services.NewAuthService(userRepo, sessionRepo, passwordResetRepo, "test-secret")

	userID := uuid.New()
	email := "test@example.com"
	role := "ADMIN"

	// Generate token
	token, err := authService.GenerateAccessToken(userID, email, role)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Validate token
	claims, err := authService.ValidateAccessToken(token)
	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
	assert.Equal(t, role, claims.Role)
}

func TestValidateAccessToken_Invalid(t *testing.T) {
	userRepo := new(MockUserRepository)
	sessionRepo := new(MockSessionRepository)
	passwordResetRepo := new(MockPasswordResetRepository)

	authService := services.NewAuthService(userRepo, sessionRepo, passwordResetRepo, "test-secret")

	// Invalid token
	claims, err := authService.ValidateAccessToken("invalid-token")
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestCreatePasswordReset_Success(t *testing.T) {
	userRepo := new(MockUserRepository)
	sessionRepo := new(MockSessionRepository)
	passwordResetRepo := new(MockPasswordResetRepository)

	authService := services.NewAuthService(userRepo, sessionRepo, passwordResetRepo, "test-secret")

	ctx := context.Background()
	email := "test@example.com"
	userID := uuid.New()

	mockUser := &db.User{
		ID:    utils.UUIDToPgtype(userID),
		Email: email,
	}

	resetID := uuid.New()
	mockReset := &db.PasswordReset{
		ID:     utils.UUIDToPgtype(resetID),
		UserID: utils.UUIDToPgtype(userID),
	}

	// Mocks
	userRepo.On("GetUserByEmail", ctx, email).Return(mockUser, nil)
	passwordResetRepo.On("CreatePasswordReset", ctx, userID, mock.Anything, mock.Anything).Return(mockReset, nil)

	// Execute
	token, err := authService.CreatePasswordReset(ctx, email)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	userRepo.AssertExpectations(t)
	passwordResetRepo.AssertExpectations(t)
}
