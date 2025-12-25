package services

import (
	"errors"
	"time"

	"github.com/gosimple/slug"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"github.com/liyali/liyali-gateway/models"
)

type OrganizationService struct {
	db *gorm.DB
}

func NewOrganizationService(db *gorm.DB) *OrganizationService {
	return &OrganizationService{db: db}
}

// CreateOrganization creates a new organization
func (s *OrganizationService) CreateOrganization(name, description, createdBy string) (*models.Organization, error) {
	if name == "" {
		return nil, errors.New("organization name is required")
	}

	if createdBy == "" {
		return nil, errors.New("creator user ID is required")
	}

	org := &models.Organization{
		ID:          uuid.New().String(),
		Name:        name,
		Slug:        slug.Make(name),
		Description: description,
		Active:      true,
		Tier:        "free",
		CreatedBy:   createdBy,
	}

	if err := s.db.Create(org).Error; err != nil {
		return nil, err
	}

	// Auto-create settings
	settings := &models.OrganizationSettings{
		ID:             uuid.New().String(),
		OrganizationID: org.ID,
	}
	if err := s.db.Create(settings).Error; err != nil {
		// Log but don't fail - settings are optional
		return org, nil
	}

	// Add creator as admin
	now := time.Now()
	member := &models.OrganizationMember{
		ID:             uuid.New().String(),
		OrganizationID: org.ID,
		UserID:         createdBy,
		Role:           "admin",
		Active:         true,
		JoinedAt:       &now,
	}
	if err := s.db.Create(member).Error; err != nil {
		return org, nil // Don't fail, but log warning
	}

	// Set as current organization for creator
	s.db.Model(&models.User{}).
		Where("id = ?", createdBy).
		Update("current_organization_id", org.ID)

	return org, nil
}

// GetOrganization retrieves an organization by ID
func (s *OrganizationService) GetOrganization(orgID string) (*models.Organization, error) {
	if orgID == "" {
		return nil, errors.New("organization ID is required")
	}

	var org models.Organization
	if err := s.db.
		Preload("Creator").
		Where("id = ? AND active = ?", orgID, true).
		First(&org).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("organization not found")
		}
		return nil, err
	}

	return &org, nil
}

// GetUserOrganizations returns all organizations a user belongs to
func (s *OrganizationService) GetUserOrganizations(userID string) ([]models.Organization, error) {
	if userID == "" {
		return nil, errors.New("user ID is required")
	}

	var orgs []models.Organization

	err := s.db.
		Joins("INNER JOIN organization_members ON organizations.id = organization_members.organization_id").
		Where("organization_members.user_id = ? AND organization_members.active = ? AND organizations.active = ?",
			userID, true, true).
		Distinct("organizations.*").
		Find(&orgs).Error

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	return orgs, nil
}

// AddMember adds a user to an organization
func (s *OrganizationService) AddMember(orgID, userID, role string) error {
	if orgID == "" || userID == "" {
		return errors.New("organization ID and user ID are required")
	}

	if role == "" {
		role = "requester" // Default role
	}

	// Check if user already a member
	var existing models.OrganizationMember
	err := s.db.Where(
		"organization_id = ? AND user_id = ?",
		orgID, userID,
	).First(&existing).Error

	if err == nil {
		// User already exists, just activate if inactive
		if !existing.Active {
			return s.db.Model(&existing).Update("active", true).Error
		}
		return errors.New("user is already a member of this organization")
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	now := time.Now()
	member := &models.OrganizationMember{
		ID:             uuid.New().String(),
		OrganizationID: orgID,
		UserID:         userID,
		Role:           role,
		Active:         true,
		JoinedAt:       &now,
	}

	return s.db.Create(member).Error
}

// RemoveMember removes a user from an organization (soft delete)
func (s *OrganizationService) RemoveMember(orgID, userID string) error {
	if orgID == "" || userID == "" {
		return errors.New("organization ID and user ID are required")
	}

	// Don't allow removing the last admin
	var adminCount int64
	s.db.Model(&models.OrganizationMember{}).
		Where("organization_id = ? AND role = ? AND active = ?", orgID, "admin", true).
		Where("user_id != ?", userID).
		Count(&adminCount)

	if adminCount == 0 {
		return errors.New("cannot remove the last admin from organization")
	}

	return s.db.Model(&models.OrganizationMember{}).
		Where("organization_id = ? AND user_id = ?", orgID, userID).
		Update("active", false).Error
}

// GetOrganizationMembers returns all members of an organization
func (s *OrganizationService) GetOrganizationMembers(orgID string) ([]models.OrganizationMember, error) {
	if orgID == "" {
		return nil, errors.New("organization ID is required")
	}

	var members []models.OrganizationMember
	if err := s.db.
		Preload("User").
		Where("organization_id = ? AND active = ?", orgID, true).
		Find(&members).Error; err != nil {
		return nil, err
	}

	return members, nil
}

// SwitchOrganization sets user's current organization
func (s *OrganizationService) SwitchOrganization(userID, orgID string) error {
	if userID == "" || orgID == "" {
		return errors.New("user ID and organization ID are required")
	}

	// Verify user is member of this organization
	var member models.OrganizationMember
	if err := s.db.Where(
		"user_id = ? AND organization_id = ? AND active = ?",
		userID, orgID, true,
	).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user is not a member of this organization")
		}
		return err
	}

	// Verify organization exists and is active
	var org models.Organization
	if err := s.db.Where("id = ? AND active = ?", orgID, true).
		First(&org).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("organization not found or is inactive")
		}
		return err
	}

	// Update user's current organization
	return s.db.Model(&models.User{}).
		Where("id = ?", userID).
		Update("current_organization_id", orgID).Error
}

// UpdateOrganizationSettings updates organization configuration
func (s *OrganizationService) UpdateOrganizationSettings(orgID string, settings *models.OrganizationSettings) error {
	if orgID == "" {
		return errors.New("organization ID is required")
	}

	// Only update writable fields
	return s.db.Model(&models.OrganizationSettings{}).
		Where("organization_id = ?", orgID).
		Updates(map[string]interface{}{
			"require_digital_signatures": settings.RequireDigitalSignatures,
			"default_approval_chain":     settings.DefaultApprovalChain,
			"currency":                   settings.Currency,
			"fiscal_year_start":          settings.FiscalYearStart,
			"enable_budget_validation":   settings.EnableBudgetValidation,
			"budget_variance_threshold":  settings.BudgetVarianceThreshold,
		}).Error
}

// GetOrganizationSettings retrieves organization settings
func (s *OrganizationService) GetOrganizationSettings(orgID string) (*models.OrganizationSettings, error) {
	if orgID == "" {
		return nil, errors.New("organization ID is required")
	}

	var settings models.OrganizationSettings
	if err := s.db.
		Where("organization_id = ?", orgID).
		First(&settings).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Return default settings if not found
			return &models.OrganizationSettings{
				ID:             uuid.New().String(),
				OrganizationID: orgID,
				Currency:       "USD",
				FiscalYearStart: 1,
			}, nil
		}
		return nil, err
	}

	return &settings, nil
}

// UpdateOrganization updates organization details
func (s *OrganizationService) UpdateOrganization(orgID string, name, description string) error {
	if orgID == "" {
		return errors.New("organization ID is required")
	}

	if name == "" {
		return errors.New("organization name is required")
	}

	return s.db.Model(&models.Organization{}).
		Where("id = ?", orgID).
		Updates(map[string]interface{}{
			"name":        name,
			"slug":        slug.Make(name),
			"description": description,
		}).Error
}
