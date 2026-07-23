package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	sqlc "github.com/liyali/liyali-gateway/database/sqlc"
)

type LoginAttemptRepository struct {
	db      *pgxpool.Pool
	queries *sqlc.Queries
}

func NewLoginAttemptRepository(db *pgxpool.Pool) LoginAttemptRepositoryInterface {
	return &LoginAttemptRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

func (r *LoginAttemptRepository) Create(ctx context.Context, userID, email, ipAddress, userAgent string, success bool, failureReason string) (*sqlc.LoginAttempt, error) {
	var userIDPtr, ipAddrPtr, userAgPtr, failureReasonPtr *string
	
	if userID != "" {
		userIDPtr = &userID
	}
	if ipAddress != "" {
		ipAddrPtr = &ipAddress
	}
	if userAgent != "" {
		userAgPtr = &userAgent
	}
	if failureReason != "" {
		failureReasonPtr = &failureReason
	}

	params := sqlc.CreateLoginAttemptParams{
		UserID:        userIDPtr,
		Email:         email,
		IpAddress:     ipAddrPtr,
		UserAgent:     userAgPtr,
		Success:       success,
		FailureReason: failureReasonPtr,
	}

	attempt, err := r.queries.CreateLoginAttempt(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create login attempt: %w", err)
	}

	return &attempt, nil
}

func (r *LoginAttemptRepository) GetRecentFailedAttempts(ctx context.Context, email string, since time.Time) (int64, error) {
	sinceTimestamp := pgtype.Timestamptz{
		Time:  since,
		Valid: true,
	}

	count, err := r.queries.GetRecentFailedAttempts(ctx, email, sinceTimestamp)
	if err != nil {
		return 0, fmt.Errorf("failed to get recent failed attempts: %w", err)
	}

	return count, nil
}

func (r *LoginAttemptRepository) GetRecentFailedAttemptsByIP(ctx context.Context, ipAddress string, since time.Time) (int64, error) {
	var ipAddrPtr *string
	if ipAddress != "" {
		ipAddrPtr = &ipAddress
	}

	sinceTimestamp := pgtype.Timestamptz{
		Time:  since,
		Valid: true,
	}

	count, err := r.queries.GetRecentFailedAttemptsByIP(ctx, ipAddrPtr, sinceTimestamp)
	if err != nil {
		return 0, fmt.Errorf("failed to get recent failed attempts by IP: %w", err)
	}

	return count, nil
}

func (r *LoginAttemptRepository) GetByUser(ctx context.Context, userID string, limit, offset int) ([]*sqlc.LoginAttempt, error) {
	var userIDPtr *string
	if userID != "" {
		userIDPtr = &userID
	}

	attempts, err := r.queries.GetLoginAttemptsByUser(ctx, userIDPtr, int32(limit), int32(offset))
	if err != nil {
		return nil, fmt.Errorf("failed to get login attempts by user: %w", err)
	}

	// Convert []LoginAttempt to []*LoginAttempt
	result := make([]*sqlc.LoginAttempt, len(attempts))
	for i := range attempts {
		result[i] = &attempts[i]
	}

	return result, nil
}

func (r *LoginAttemptRepository) GetByEmail(ctx context.Context, email string, limit, offset int) ([]*sqlc.LoginAttempt, error) {
	attempts, err := r.queries.GetLoginAttemptsByEmail(ctx, email, int32(limit), int32(offset))
	if err != nil {
		return nil, fmt.Errorf("failed to get login attempts by email: %w", err)
	}

	// Convert []LoginAttempt to []*LoginAttempt
	result := make([]*sqlc.LoginAttempt, len(attempts))
	for i := range attempts {
		result[i] = &attempts[i]
	}

	return result, nil
}

func (r *LoginAttemptRepository) DeleteOld(ctx context.Context, before time.Time) error {
	beforeTimestamp := pgtype.Timestamptz{
		Time:  before,
		Valid: true,
	}

	err := r.queries.DeleteOldLoginAttempts(ctx, beforeTimestamp)
	if err != nil {
		return fmt.Errorf("failed to delete old login attempts: %w", err)
	}

	return nil
}