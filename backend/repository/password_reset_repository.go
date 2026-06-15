package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	sqlc "github.com/liyali/liyali-gateway/database/sqlc"
)

type PasswordResetRepository struct {
	db      *pgxpool.Pool
	queries *sqlc.Queries
}

func NewPasswordResetRepository(db *pgxpool.Pool) PasswordResetRepositoryInterface {
	return &PasswordResetRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

func (r *PasswordResetRepository) Create(ctx context.Context, userID, token string, expiresAt time.Time) (*sqlc.PasswordReset, error) {
	reset, err := r.queries.CreatePasswordReset(ctx, userID, token, pgtype.Timestamptz{
		Time:  expiresAt,
		Valid: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create password reset: %w", err)
	}
	return &reset, nil
}

func (r *PasswordResetRepository) GetByToken(ctx context.Context, token string) (*sqlc.PasswordReset, error) {
	// The query already filters to unused, non-expired rows; a no-rows result
	// surfaces as an error the caller maps to "invalid or expired token".
	reset, err := r.queries.GetPasswordResetByToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get password reset by token: %w", err)
	}
	return &reset, nil
}

func (r *PasswordResetRepository) MarkAsUsed(ctx context.Context, id uuid.UUID) error {
	if err := r.queries.MarkPasswordResetAsUsed(ctx, pgtype.UUID{Bytes: id, Valid: true}); err != nil {
		return fmt.Errorf("failed to mark password reset as used: %w", err)
	}
	return nil
}

func (r *PasswordResetRepository) DeleteByUserID(ctx context.Context, userID string) error {
	if err := r.queries.DeletePasswordResetsByUserID(ctx, userID); err != nil {
		return fmt.Errorf("failed to delete password resets by user: %w", err)
	}
	return nil
}

func (r *PasswordResetRepository) DeleteExpired(ctx context.Context) error {
	if err := r.queries.DeleteExpiredPasswordResets(ctx); err != nil {
		return fmt.Errorf("failed to delete expired password resets: %w", err)
	}
	return nil
}

func (r *PasswordResetRepository) DeleteUsed(ctx context.Context) error {
	if err := r.queries.DeleteUsedPasswordResets(ctx); err != nil {
		return fmt.Errorf("failed to delete used password resets: %w", err)
	}
	return nil
}
