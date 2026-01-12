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

type SessionRepository struct {
	db      *pgxpool.Pool
	queries *sqlc.Queries
}

func NewSessionRepository(db *pgxpool.Pool) SessionRepositoryInterface {
	return &SessionRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

func (r *SessionRepository) Create(ctx context.Context, userID, refreshToken, ipAddress, userAgent string, expiresAt time.Time) (*sqlc.Session, error) {
	var ipAddr, userAg *string
	if ipAddress != "" {
		ipAddr = &ipAddress
	}
	if userAgent != "" {
		userAg = &userAgent
	}

	expiresAtPgType := pgtype.Timestamp{
		Time:  expiresAt,
		Valid: true,
	}

	session, err := r.queries.CreateSession(ctx, userID, refreshToken, ipAddr, userAg, expiresAtPgType)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &session, nil
}

func (r *SessionRepository) GetByRefreshToken(ctx context.Context, refreshToken string) (*sqlc.Session, error) {
	session, err := r.queries.GetSessionByRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get session by refresh token: %w", err)
	}

	return &session, nil
}

func (r *SessionRepository) GetByUserID(ctx context.Context, userID string) ([]*sqlc.Session, error) {
	sessions, err := r.queries.GetSessionsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions by user ID: %w", err)
	}

	// Convert []Session to []*Session
	result := make([]*sqlc.Session, len(sessions))
	for i := range sessions {
		result[i] = &sessions[i]
	}

	return result, nil
}

func (r *SessionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	pgUUID := pgtype.UUID{
		Bytes: id,
		Valid: true,
	}

	err := r.queries.DeleteSession(ctx, pgUUID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

func (r *SessionRepository) DeleteByRefreshToken(ctx context.Context, refreshToken string) error {
	err := r.queries.DeleteSessionByRefreshToken(ctx, refreshToken)
	if err != nil {
		return fmt.Errorf("failed to delete session by refresh token: %w", err)
	}

	return nil
}

func (r *SessionRepository) DeleteByUserID(ctx context.Context, userID string) error {
	err := r.queries.DeleteSessionsByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to delete sessions by user ID: %w", err)
	}

	return nil
}

func (r *SessionRepository) DeleteExpired(ctx context.Context) error {
	err := r.queries.DeleteExpiredSessions(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete expired sessions: %w", err)
	}

	return nil
}

func (r *SessionRepository) CountActive(ctx context.Context) (int64, error) {
	count, err := r.queries.CountActiveSessions(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count active sessions: %w", err)
	}

	return count, nil
}

func (r *SessionRepository) CountUserActive(ctx context.Context, userID string) (int64, error) {
	count, err := r.queries.CountUserActiveSessions(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to count user active sessions: %w", err)
	}

	return count, nil
}

// UpdateRefreshToken updates the refresh token for a session with old token verification
func (r *SessionRepository) UpdateRefreshToken(ctx context.Context, id uuid.UUID, oldRefreshToken, newRefreshToken string, expiresAt time.Time) (int64, error) {
	rowsAffected, err := r.queries.UpdateSessionRefreshToken(ctx, pgtype.UUID{
		Bytes: id,
		Valid: true,
	}, newRefreshToken, pgtype.Timestamp{
		Time:  expiresAt,
		Valid: true,
	}, oldRefreshToken)

	if err != nil {
		return 0, err
	}

	return rowsAffected, nil
}
