package repository

import (
	"context"
	"time"

	"github.com/cozyCodr/liyali-gateway/internal/db"
	"github.com/cozyCodr/liyali-gateway/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type PasswordResetRepository struct {
	queries *db.Queries
}

func NewPasswordResetRepository(queries *db.Queries) *PasswordResetRepository {
	return &PasswordResetRepository{
		queries: queries,
	}
}

// CreatePasswordReset creates a new password reset record
func (r *PasswordResetRepository) CreatePasswordReset(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) (*db.PasswordReset, error) {
	reset, err := r.queries.CreatePasswordReset(ctx, db.CreatePasswordResetParams{
		UserID: utils.UUIDToPgtype(userID),
		Token:  token,
		ExpiresAt: pgtype.Timestamp{
			Time:  expiresAt,
			Valid: true,
		},
	})
	if err != nil {
		return nil, err
	}
	return &reset, nil
}

// GetPasswordResetByToken retrieves a password reset by token
func (r *PasswordResetRepository) GetPasswordResetByToken(ctx context.Context, token string) (*db.PasswordReset, error) {
	reset, err := r.queries.GetPasswordResetByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	return &reset, nil
}

// MarkPasswordResetAsUsed marks a password reset as used
func (r *PasswordResetRepository) MarkPasswordResetAsUsed(ctx context.Context, id uuid.UUID) error {
	return r.queries.MarkPasswordResetAsUsed(ctx, utils.UUIDToPgtype(id))
}

// DeletePasswordResetsByUserID deletes all password resets for a user
func (r *PasswordResetRepository) DeletePasswordResetsByUserID(ctx context.Context, userID uuid.UUID) error {
	return r.queries.DeletePasswordResetsByUserID(ctx, utils.UUIDToPgtype(userID))
}

// DeleteExpiredPasswordResets removes all expired password resets
func (r *PasswordResetRepository) DeleteExpiredPasswordResets(ctx context.Context) error {
	return r.queries.DeleteExpiredPasswordResets(ctx)
}

// DeleteUsedPasswordResets removes all used password resets
func (r *PasswordResetRepository) DeleteUsedPasswordResets(ctx context.Context) error {
	return r.queries.DeleteUsedPasswordResets(ctx)
}
