package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	sqlc "github.com/liyali/liyali-gateway/database/sqlc"
	"github.com/liyali/liyali-gateway/models"
	"gorm.io/gorm"
)

type UserRepository struct {
	db      *pgxpool.Pool
	queries *sqlc.Queries
	gormDB  *gorm.DB // Keep GORM for complex operations during transition
}

func NewUserRepository(db *pgxpool.Pool, gormDB *gorm.DB) UserRepositoryInterface {
	return &UserRepository{
		db:      db,
		queries: sqlc.New(db),
		gormDB:  gormDB,
	}
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
	// Use GORM for now to maintain compatibility with existing models
	if err := r.gormDB.WithContext(ctx).Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	var user models.User
	if err := r.gormDB.WithContext(ctx).Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	if err := r.gormDB.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *models.User) (*models.User, error) {
	if err := r.gormDB.WithContext(ctx).Save(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) UpdatePassword(ctx context.Context, id string, hashedPassword string) error {
	// Use GORM for now
	return r.gormDB.WithContext(ctx).Model(&models.User{}).Where("id = ?", id).Update("password", hashedPassword).Error
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, id string) error {
	// Use GORM for now
	return r.gormDB.WithContext(ctx).Model(&models.User{}).Where("id = ?", id).Update("last_login", "NOW()").Error
}

func (r *UserRepository) Delete(ctx context.Context, id string) error {
	return r.gormDB.WithContext(ctx).Delete(&models.User{}, "id = ?", id).Error
}

func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]*models.User, error) {
	var users []*models.User
	if err := r.gormDB.WithContext(ctx).
		Limit(limit).
		Offset(offset).
		Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepository) ListByOrganization(ctx context.Context, organizationID string, limit, offset int) ([]*models.User, error) {
	// Use GORM for now - this would need proper organization membership table join
	var users []*models.User
	err := r.gormDB.WithContext(ctx).
		Where("current_organization_id = ?", organizationID).
		Limit(limit).
		Offset(offset).
		Find(&users).Error
	return users, err
}

func (r *UserRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.gormDB.WithContext(ctx).Model(&models.User{}).Count(&count).Error
	return count, err
}

func (r *UserRepository) CountActive(ctx context.Context) (int64, error) {
	var count int64
	err := r.gormDB.WithContext(ctx).Model(&models.User{}).Where("active = ?", true).Count(&count).Error
	return count, err
}

func (r *UserRepository) Activate(ctx context.Context, id string) error {
	return r.gormDB.WithContext(ctx).Model(&models.User{}).Where("id = ?", id).Update("active", true).Error
}

func (r *UserRepository) Deactivate(ctx context.Context, id string) error {
	return r.gormDB.WithContext(ctx).Model(&models.User{}).Where("id = ?", id).Update("active", false).Error
}