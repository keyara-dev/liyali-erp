package repository

import (
	"context"
	"time"

	"github.com/cozyCodr/liyali-gateway/internal/db"
	"github.com/cozyCodr/liyali-gateway/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type SessionRepository struct {
	queries *db.Queries
}

func NewSessionRepository(queries *db.Queries) *SessionRepository {
	return &SessionRepository{
		queries: queries,
	}
}

// CreateSession creates a new session
func (r *SessionRepository) CreateSession(ctx context.Context, userID uuid.UUID, refreshToken, ipAddress, userAgent string, expiresAt time.Time) (*db.Session, error) {
	session, err := r.queries.CreateSession(ctx, db.CreateSessionParams{
		UserID:       utils.UUIDToPgtype(userID),
		RefreshToken: refreshToken,
		IpAddress:    utils.StringToPgtype(ipAddress),
		UserAgent:    utils.StringToPgtype(userAgent),
		ExpiresAt: pgtype.Timestamp{
			Time:  expiresAt,
			Valid: true,
		},
	})
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// GetSessionByRefreshToken retrieves a session by refresh token
func (r *SessionRepository) GetSessionByRefreshToken(ctx context.Context, refreshToken string) (*db.Session, error) {
	session, err := r.queries.GetSessionByRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// GetSessionsByUserID retrieves all active sessions for a user
func (r *SessionRepository) GetSessionsByUserID(ctx context.Context, userID uuid.UUID) ([]db.Session, error) {
	return r.queries.GetSessionsByUserID(ctx, utils.UUIDToPgtype(userID))
}

// DeleteSession deletes a session by ID
func (r *SessionRepository) DeleteSession(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeleteSession(ctx, utils.UUIDToPgtype(id))
}

// DeleteSessionByRefreshToken deletes a session by refresh token
func (r *SessionRepository) DeleteSessionByRefreshToken(ctx context.Context, refreshToken string) error {
	return r.queries.DeleteSessionByRefreshToken(ctx, refreshToken)
}

// DeleteSessionsByUserID deletes all sessions for a user
func (r *SessionRepository) DeleteSessionsByUserID(ctx context.Context, userID uuid.UUID) error {
	return r.queries.DeleteSessionsByUserID(ctx, utils.UUIDToPgtype(userID))
}

// DeleteExpiredSessions removes all expired sessions
func (r *SessionRepository) DeleteExpiredSessions(ctx context.Context) error {
	return r.queries.DeleteExpiredSessions(ctx)
}

// CountActiveSessions returns the number of active sessions
func (r *SessionRepository) CountActiveSessions(ctx context.Context) (int64, error) {
	return r.queries.CountActiveSessions(ctx)
}

// CountUserActiveSessions returns the number of active sessions for a user
func (r *SessionRepository) CountUserActiveSessions(ctx context.Context, userID uuid.UUID) (int64, error) {
	return r.queries.CountUserActiveSessions(ctx, utils.UUIDToPgtype(userID))
}
