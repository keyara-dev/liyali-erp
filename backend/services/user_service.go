package services

import (
	"fmt"
	"time"

	"gorm.io/gorm"
	"github.com/liyali/liyali-gateway/models"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

// UserExistsInOrganization checks if a user exists and belongs to the organization
func (s *UserService) UserExistsInOrganization(organizationID, userID string) (bool, error) {
	var count int64
	
	// Check if user exists and is a member of the organization
	if err := s.db.Table("organization_members").
		Where("organization_id = ? AND user_id = ? AND active = ?", organizationID, userID, true).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}

	return count > 0, nil
}

// GetUserByEmail gets a user by email (for checking if user already exists).
// Returns the user if the email is taken globally — the UNIQUE constraint on
// users.email makes cross-org re-use impossible, so any match is a conflict.
func (s *UserService) GetUserByEmail(organizationID, email string) (*models.User, error) {
	var user models.User
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // email is free
		}
		return nil, fmt.Errorf("failed to check user by email: %w", err)
	}
	return &user, nil // email already taken — caller should return 409
}

// EmailLookupResult holds the result of a per-org email lookup.
type EmailLookupResult struct {
	User     *models.User // nil if no global account exists
	IsMember bool         // true if the user is already an active member of orgID
}

// LookupUserByEmailForOrg checks whether an email address belongs to an existing
// platform user, and whether that user is already a member of the given org.
// All three cases are distinguishable from the returned result:
//
//	result.User == nil                → email free, safe to create
//	result.User != nil && IsMember   → already a member, block creation
//	result.User != nil && !IsMember  → has a global account, offer invite flow
func (s *UserService) LookupUserByEmailForOrg(orgID, email string) (*EmailLookupResult, error) {
	var user models.User
	if err := s.db.Where("email = ? AND deleted_at IS NULL", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return &EmailLookupResult{}, nil
		}
		return nil, fmt.Errorf("failed to look up user by email: %w", err)
	}

	var memberCount int64
	s.db.Table("organization_members").
		Where("organization_id = ? AND user_id = ? AND active = true", orgID, user.ID).
		Count(&memberCount)

	return &EmailLookupResult{User: &user, IsMember: memberCount > 0}, nil
}

// AssignUserToDepartment assigns a user to a department
func (s *UserService) AssignUserToDepartment(organizationID, userID, departmentID string) error {
	// For now, we'll use the organization_members table to store department assignment
	// In a more complex system, you might want a separate user_departments table
	
	result := s.db.Table("organization_members").
		Where("organization_id = ? AND user_id = ?", organizationID, userID).
		Updates(map[string]interface{}{
			"department_id": departmentID,
			"updated_at":    time.Now(),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to assign user to department: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found in organization")
	}

	return nil
}

// GetUserDepartment retrieves the department assigned to a user
func (s *UserService) GetUserDepartment(organizationID, userID string) (*models.OrganizationDepartment, error) {
	var department models.OrganizationDepartment
	
	// Join organization_members with organization_departments
	if err := s.db.Table("organization_departments").
		Select("organization_departments.*").
		Joins("JOIN organization_members ON organization_members.department_id = organization_departments.id").
		Where("organization_members.organization_id = ? AND organization_members.user_id = ? AND organization_members.active = ?", 
			organizationID, userID, true).
		First(&department).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // User has no department assigned
		}
		return nil, fmt.Errorf("failed to fetch user department: %w", err)
	}

	return &department, nil
}

// RemoveUserFromDepartment removes a user from their current department
func (s *UserService) RemoveUserFromDepartment(organizationID, userID string) error {
	result := s.db.Table("organization_members").
		Where("organization_id = ? AND user_id = ?", organizationID, userID).
		Updates(map[string]interface{}{
			"department_id": nil,
			"updated_at":    time.Now(),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to remove user from department: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found in organization")
	}

	return nil
}

// GetDepartmentUsers retrieves all users in a specific department
func (s *UserService) GetDepartmentUsers(organizationID, departmentID string) ([]interface{}, error) {
	var users []struct {
		ID           string    `json:"id"`
		Email        string    `json:"email"`
		Name         string    `json:"name"`
		Role         string    `json:"role"`
		Active       bool      `json:"active"`
		JoinedAt     time.Time `json:"joined_at"`
		DepartmentID *string   `json:"department_id"`
	}

	if err := s.db.Table("users").
		Select("users.id, users.email, users.name, organization_members.role, users.active, organization_members.joined_at, organization_members.department_id").
		Joins("JOIN organization_members ON users.id = organization_members.user_id").
		Where("organization_members.organization_id = ? AND organization_members.department_id = ? AND organization_members.active = ?", 
			organizationID, departmentID, true).
		Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch department users: %w", err)
	}

	// Convert to interface slice
	result := make([]interface{}, len(users))
	for i, user := range users {
		result[i] = user
	}

	return result, nil
}

// GetUsersByOrganization retrieves all users in an organization with their department info
func (s *UserService) GetUsersByOrganization(organizationID string, page, pageSize int) ([]interface{}, int64, error) {
	var users []struct {
		ID             string    `json:"id"`
		Email          string    `json:"email"`
		Name           string    `json:"name"`
		Role           string    `json:"role"`
		Active         bool      `json:"active"`
		JoinedAt       time.Time `json:"joined_at"`
		DepartmentID   *string   `json:"department_id"`
		DepartmentName *string   `json:"department_name"`
		DepartmentCode *string   `json:"department_code"`
	}
	var total int64

	offset := (page - 1) * pageSize

	// Count total users in organization
	if err := s.db.Table("organization_members").
		Where("organization_id = ? AND active = ?", organizationID, true).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Get users with department info
	if err := s.db.Table("users").
		Select(`users.id, users.email, users.name, organization_members.role, users.active, 
				organization_members.joined_at, organization_members.department_id,
				organization_departments.name as department_name, organization_departments.code as department_code`).
		Joins("JOIN organization_members ON users.id = organization_members.user_id").
		Joins("LEFT JOIN organization_departments ON organization_members.department_id = organization_departments.id").
		Where("organization_members.organization_id = ? AND organization_members.active = ?", organizationID, true).
		Order("users.name ASC").
		Limit(pageSize).
		Offset(offset).
		Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch users: %w", err)
	}

	// Convert to interface slice
	result := make([]interface{}, len(users))
	for i, user := range users {
		result[i] = user
	}

	return result, total, nil
}

// CreateUserInOrganization creates a new user and adds them to an organization
func (s *UserService) CreateUserInOrganization(organizationID string, email, name, password, role string, departmentID *string) (*models.User, error) {
	// Start a transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create the user
	user := models.User{
		ID:       fmt.Sprintf("user_%d", time.Now().UnixNano()),
		Email:    email,
		Name:     name,
		Password: password, // Should be hashed before calling this function
		Role:     role,
		Active:   true,
	}

	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Add user to organization
	orgMember := struct {
		ID             string    `gorm:"primaryKey"`
		OrganizationID string    `gorm:"index;not null"`
		UserID         string    `gorm:"index;not null"`
		Role           string    `gorm:"not null"`
		DepartmentID   *string   `gorm:"index"`
		Active         bool      `gorm:"default:true"`
		JoinedAt       time.Time `gorm:"default:CURRENT_TIMESTAMP"`
		CreatedAt      time.Time
		UpdatedAt      time.Time
	}{
		ID:             fmt.Sprintf("member_%d", time.Now().UnixNano()),
		OrganizationID: organizationID,
		UserID:         user.ID,
		Role:           role,
		DepartmentID:   departmentID,
		Active:         true,
		JoinedAt:       time.Now(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := tx.Table("organization_members").Create(&orgMember).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to add user to organization: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &user, nil
}