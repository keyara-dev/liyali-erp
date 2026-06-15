package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	sqlc "github.com/liyali/liyali-gateway/database/sqlc"
)

// ErrNotImplemented is retained for any repository methods still pending a
// real implementation; the account-lockout and password-reset repos below are
// now fully wired to their sqlc queries.
var ErrNotImplemented = errors.New("repository method not implemented - requires sqlc generation")

type AccountLockoutRepository struct {
	db      *pgxpool.Pool
	queries *sqlc.Queries
}

func NewAccountLockoutRepository(db *pgxpool.Pool) AccountLockoutRepositoryInterface {
	return &AccountLockoutRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

func (r *AccountLockoutRepository) Create(ctx context.Context, userID, email, ipAddress, reason string, unlocksAt time.Time) (*sqlc.AccountLockout, error) {
	var ipPtr *string
	if ipAddress != "" {
		ipPtr = &ipAddress
	}

	lockout, err := r.queries.CreateAccountLockout(ctx, userID, email, ipPtr, reason, pgtype.Timestamptz{
		Time:  unlocksAt,
		Valid: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create account lockout: %w", err)
	}
	return &lockout, nil
}

// GetActiveByUserID returns the active, non-expired lockout for a user, or an
// error (including no-rows) when the account is not locked.
func (r *AccountLockoutRepository) GetActiveByUserID(ctx context.Context, userID string) (*sqlc.AccountLockout, error) {
	lockout, err := r.queries.GetActiveAccountLockout(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &lockout, nil
}

// GetActiveByEmail returns the active, non-expired lockout for an email, or an
// error (including no-rows) when the account is not locked.
func (r *AccountLockoutRepository) GetActiveByEmail(ctx context.Context, email string) (*sqlc.AccountLockout, error) {
	lockout, err := r.queries.GetAccountLockoutByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return &lockout, nil
}

func (r *AccountLockoutRepository) Unlock(ctx context.Context, userID string) error {
	if err := r.queries.UnlockAccount(ctx, userID); err != nil {
		return fmt.Errorf("failed to unlock account: %w", err)
	}
	return nil
}

func (r *AccountLockoutRepository) UnlockByEmail(ctx context.Context, email string) error {
	if err := r.queries.UnlockAccountByEmail(ctx, email); err != nil {
		return fmt.Errorf("failed to unlock account by email: %w", err)
	}
	return nil
}

func (r *AccountLockoutRepository) GetHistory(ctx context.Context, userID string, limit, offset int) ([]*sqlc.AccountLockout, error) {
	lockouts, err := r.queries.GetAccountLockoutHistory(ctx, userID, int32(limit), int32(offset))
	if err != nil {
		return nil, fmt.Errorf("failed to get account lockout history: %w", err)
	}

	result := make([]*sqlc.AccountLockout, len(lockouts))
	for i := range lockouts {
		result[i] = &lockouts[i]
	}
	return result, nil
}

func (r *AccountLockoutRepository) CleanupExpired(ctx context.Context) error {
	if err := r.queries.CleanupExpiredLockouts(ctx); err != nil {
		return fmt.Errorf("failed to cleanup expired lockouts: %w", err)
	}
	return nil
}
