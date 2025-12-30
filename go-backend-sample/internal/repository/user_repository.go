package repository

import (
	"context"
	"time"

	"github.com/cozyCodr/liyali-gateway/internal/db"
	"github.com/cozyCodr/liyali-gateway/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type UserRepository struct {
	queries *db.Queries
}

func NewUserRepository(queries *db.Queries) *UserRepository {
	return &UserRepository{
		queries: queries,
	}
}

// CreateUser creates a new user
func (r *UserRepository) CreateUser(ctx context.Context, params db.CreateUserParams) (*db.User, error) {
	user, err := r.queries.CreateUser(ctx, params)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByID retrieves a user by ID
func (r *UserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*db.User, error) {
	user, err := r.queries.GetUserByID(ctx, utils.UUIDToPgtype(id))
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByEmail retrieves a user by email
func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*db.User, error) {
	user, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUser updates user details
func (r *UserRepository) UpdateUser(ctx context.Context, params db.UpdateUserParams) (*db.User, error) {
	user, err := r.queries.UpdateUser(ctx, params)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUserPassword updates user password hash
func (r *UserRepository) UpdateUserPassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	return r.queries.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
		ID:           utils.UUIDToPgtype(id),
		PasswordHash: passwordHash,
	})
}

// UpdateUserLastLogin updates the last login timestamp
func (r *UserRepository) UpdateUserLastLogin(ctx context.Context, id uuid.UUID) error {
	return r.queries.UpdateUserLastLogin(ctx, utils.UUIDToPgtype(id))
}

// IncrementFailedLoginAttempts increments failed login attempts
func (r *UserRepository) IncrementFailedLoginAttempts(ctx context.Context, id uuid.UUID) error {
	return r.queries.IncrementFailedLoginAttempts(ctx, utils.UUIDToPgtype(id))
}

// ResetFailedLoginAttempts resets failed login attempts to 0
func (r *UserRepository) ResetFailedLoginAttempts(ctx context.Context, id uuid.UUID) error {
	return r.queries.ResetFailedLoginAttempts(ctx, utils.UUIDToPgtype(id))
}

// LockUserAccount locks a user account until specified time
func (r *UserRepository) LockUserAccount(ctx context.Context, id uuid.UUID, lockedUntil time.Time) error {
	return r.queries.LockUserAccount(ctx, db.LockUserAccountParams{
		ID: utils.UUIDToPgtype(id),
		LockedUntil: pgtype.Timestamp{
			Time:  lockedUntil,
			Valid: true,
		},
	})
}

// DeactivateUser deactivates a user account
func (r *UserRepository) DeactivateUser(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeactivateUser(ctx, utils.UUIDToPgtype(id))
}

// ActivateUser activates a user account
func (r *UserRepository) ActivateUser(ctx context.Context, id uuid.UUID) error {
	return r.queries.ActivateUser(ctx, utils.UUIDToPgtype(id))
}

// DeleteUser permanently deletes a user
func (r *UserRepository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeleteUser(ctx, utils.UUIDToPgtype(id))
}

// ListUsers lists users with pagination
func (r *UserRepository) ListUsers(ctx context.Context, limit, offset int32) ([]db.User, error) {
	return r.queries.ListUsers(ctx, db.ListUsersParams{
		Limit:  limit,
		Offset: offset,
	})
}

// ListUsersByRole lists users by role
func (r *UserRepository) ListUsersByRole(ctx context.Context, role string) ([]db.User, error) {
	return r.queries.ListUsersByRole(ctx, role)
}

// ListUsersByDepartment lists users by department
func (r *UserRepository) ListUsersByDepartment(ctx context.Context, department string) ([]db.User, error) {
	return r.queries.ListUsersByDepartment(ctx, utils.StringToPgtype(department))
}

// CountUsers returns total number of users
func (r *UserRepository) CountUsers(ctx context.Context) (int64, error) {
	return r.queries.CountUsers(ctx)
}

// CountActiveUsers returns number of active users
func (r *UserRepository) CountActiveUsers(ctx context.Context) (int64, error) {
	return r.queries.CountActiveUsers(ctx)
}
