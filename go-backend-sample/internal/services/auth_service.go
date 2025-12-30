package services

import (
	"context"
	"errors"
	"time"

	"github.com/cozyCodr/liyali-gateway/internal/db"
	"github.com/cozyCodr/liyali-gateway/internal/repository"
	"github.com/cozyCodr/liyali-gateway/internal/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	AccessTokenDuration  = 1 * time.Hour
	RefreshTokenDuration = 8 * time.Hour
	PasswordResetExpiry  = 1 * time.Hour
	MaxFailedAttempts    = 5
	AccountLockDuration  = 15 * time.Minute
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrAccountLocked      = errors.New("account is locked due to too many failed login attempts")
	ErrAccountInactive    = errors.New("account is inactive")
	ErrEmailNotVerified   = errors.New("email not verified")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrEmailAlreadyExists = errors.New("email already exists")
)

type JWTClaims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Role   string    `json:"role"`
	jwt.RegisteredClaims
}

type AuthService struct {
	userRepo          repository.UserRepositoryInterface
	sessionRepo       repository.SessionRepositoryInterface
	passwordResetRepo repository.PasswordResetRepositoryInterface
	jwtSecret         []byte
}

func NewAuthService(
	userRepo repository.UserRepositoryInterface,
	sessionRepo repository.SessionRepositoryInterface,
	passwordResetRepo repository.PasswordResetRepositoryInterface,
	jwtSecret string,
) *AuthService {
	return &AuthService{
		userRepo:          userRepo,
		sessionRepo:       sessionRepo,
		passwordResetRepo: passwordResetRepo,
		jwtSecret:         []byte(jwtSecret),
	}
}

// RegisterUser creates a new user account
func (s *AuthService) RegisterUser(ctx context.Context, email, password, name, role, department string) (*db.User, error) {
	// Check if user already exists
	_, err := s.userRepo.GetUserByEmail(ctx, email)
	if err == nil {
		return nil, ErrEmailAlreadyExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create user
	user, err := s.userRepo.CreateUser(ctx, db.CreateUserParams{
		Email:         email,
		PasswordHash:  string(hashedPassword),
		Name:          name,
		Role:          role,
		Department:    utils.StringToPgtype(department),
		IsActive:      utils.BoolToPgtype(true),
		EmailVerified: utils.BoolToPgtype(false),
	})
	if err != nil {
		return nil, err
	}

	return user, nil
}

// Login authenticates a user and returns access and refresh tokens
func (s *AuthService) Login(ctx context.Context, email, password, ipAddress, userAgent string) (accessToken, refreshToken string, user *db.User, err error) {
	// Get user by email
	userRecord, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", "", nil, ErrInvalidCredentials
	}

	// Check if account is locked
	if userRecord.LockedUntil.Valid && time.Now().Before(userRecord.LockedUntil.Time) {
		return "", "", nil, ErrAccountLocked
	}

	// Check if account is active
	if !utils.PgtypeToBool(userRecord.IsActive) {
		return "", "", nil, ErrAccountInactive
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(userRecord.PasswordHash), []byte(password))
	if err != nil {
		// Increment failed login attempts
		userID := utils.PgtypeToUUID(userRecord.ID)
		_ = s.userRepo.IncrementFailedLoginAttempts(ctx, userID)

		// Lock account if max attempts reached
		if utils.PgtypeToInt32(userRecord.FailedLoginAttempts)+1 >= MaxFailedAttempts {
			lockUntil := time.Now().Add(AccountLockDuration)
			_ = s.userRepo.LockUserAccount(ctx, userID, lockUntil)
		}

		return "", "", nil, ErrInvalidCredentials
	}

	// Reset failed login attempts
	userID := utils.PgtypeToUUID(userRecord.ID)
	_ = s.userRepo.ResetFailedLoginAttempts(ctx, userID)

	// Update last login
	_ = s.userRepo.UpdateUserLastLogin(ctx, userID)

	// Generate access token
	accessToken, err = s.GenerateAccessToken(userID, userRecord.Email, userRecord.Role)
	if err != nil {
		return "", "", nil, err
	}

	// Generate refresh token
	refreshToken, err = s.GenerateRefreshToken()
	if err != nil {
		return "", "", nil, err
	}

	// Create session
	_, err = s.sessionRepo.CreateSession(ctx, userID, refreshToken, ipAddress, userAgent, time.Now().Add(RefreshTokenDuration))
	if err != nil {
		return "", "", nil, err
	}

	return accessToken, refreshToken, userRecord, nil
}

// GenerateAccessToken creates a JWT access token
func (s *AuthService) GenerateAccessToken(userID uuid.UUID, email, role string) (string, error) {
	claims := JWTClaims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(AccessTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

// GenerateRefreshToken creates a random refresh token
func (s *AuthService) GenerateRefreshToken() (string, error) {
	return uuid.New().String(), nil
}

// ValidateAccessToken validates and parses a JWT access token
func (s *AuthService) ValidateAccessToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// RefreshAccessToken generates a new access token from a refresh token
func (s *AuthService) RefreshAccessToken(ctx context.Context, refreshToken string) (string, error) {
	// Get session by refresh token
	session, err := s.sessionRepo.GetSessionByRefreshToken(ctx, refreshToken)
	if err != nil {
		return "", ErrInvalidToken
	}

	// Get user
	userID := utils.PgtypeToUUID(session.UserID)
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return "", ErrUserNotFound
	}

	// Check if account is active
	if !utils.PgtypeToBool(user.IsActive) {
		return "", ErrAccountInactive
	}

	// Generate new access token
	accessToken, err := s.GenerateAccessToken(userID, user.Email, user.Role)
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

// Logout invalidates a refresh token
func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	return s.sessionRepo.DeleteSessionByRefreshToken(ctx, refreshToken)
}

// CreatePasswordReset generates a password reset token
func (s *AuthService) CreatePasswordReset(ctx context.Context, email string) (string, error) {
	// Get user by email
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", ErrUserNotFound
	}

	// Generate reset token
	token := uuid.New().String()

	// Create password reset record
	userID := utils.PgtypeToUUID(user.ID)
	_, err = s.passwordResetRepo.CreatePasswordReset(ctx, userID, token, time.Now().Add(PasswordResetExpiry))
	if err != nil {
		return "", err
	}

	return token, nil
}

// ResetPassword resets a user's password using a reset token
func (s *AuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
	// Get password reset record
	resetRecord, err := s.passwordResetRepo.GetPasswordResetByToken(ctx, token)
	if err != nil {
		return ErrInvalidToken
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Update user password
	userID := utils.PgtypeToUUID(resetRecord.UserID)
	err = s.userRepo.UpdateUserPassword(ctx, userID, string(hashedPassword))
	if err != nil {
		return err
	}

	// Mark reset token as used
	resetID := utils.PgtypeToUUID(resetRecord.ID)
	err = s.passwordResetRepo.MarkPasswordResetAsUsed(ctx, resetID)
	if err != nil {
		return err
	}

	// Delete all sessions for user (force re-login)
	_ = s.sessionRepo.DeleteSessionsByUserID(ctx, userID)

	return nil
}

// ChangePassword changes a user's password (requires current password)
func (s *AuthService) ChangePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error {
	// Get user
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return ErrUserNotFound
	}

	// Verify current password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword))
	if err != nil {
		return ErrInvalidCredentials
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Update user password
	err = s.userRepo.UpdateUserPassword(ctx, userID, string(hashedPassword))
	if err != nil {
		return err
	}

	// Delete all sessions except current one
	_ = s.sessionRepo.DeleteSessionsByUserID(ctx, userID)

	return nil
}
