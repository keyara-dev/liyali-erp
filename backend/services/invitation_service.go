package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/liyali/liyali-gateway/models"
	"gorm.io/gorm"
)

// InvitationService manages organization invitations for existing platform users.
type InvitationService struct {
	db *gorm.DB
}

func NewInvitationService(db *gorm.DB) *InvitationService {
	return &InvitationService{db: db}
}

// generateToken returns a 32-byte random hex string for accept/decline links.
func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// SendInvitation creates a new invitation for an existing platform user to join an org.
// If a pending invitation already exists for the same org+email it is cancelled first,
// allowing admins to re-invite after a decline or expiry without manual cleanup.
func (s *InvitationService) SendInvitation(
	orgID, invitedByID, invitedEmail, role string,
	departmentID, branchID *string,
) (*models.OrganizationInvitation, error) {
	if orgID == "" || invitedByID == "" || invitedEmail == "" {
		return nil, errors.New("orgID, invitedByID and invitedEmail are required")
	}
	if role == "" {
		role = "requester"
	}

	// Resolve the invitee's user ID (must have a global account for this flow).
	var invitedUser models.User
	if err := s.db.Where("email = ? AND deleted_at IS NULL", invitedEmail).
		First(&invitedUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("no platform account found for this email address")
		}
		return nil, fmt.Errorf("failed to look up invitee: %w", err)
	}

	// Block if already an active member.
	var memberCount int64
	s.db.Table("organization_members").
		Where("organization_id = ? AND user_id = ? AND active = true", orgID, invitedUser.ID).
		Count(&memberCount)
	if memberCount > 0 {
		return nil, errors.New("this user is already a member of your organization")
	}

	// Cancel any existing pending invitation for the same org+email so there is
	// never more than one active invite at a time.
	s.db.Model(&models.OrganizationInvitation{}).
		Where("organization_id = ? AND invited_email = ? AND UPPER(status) = 'PENDING'", orgID, invitedEmail).
		Update("status", "CANCELLED")

	token, err := generateToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate invitation token: %w", err)
	}

	inv := &models.OrganizationInvitation{
		ID:             uuid.New().String(),
		OrganizationID: orgID,
		InvitedUserID:  &invitedUser.ID,
		InvitedEmail:   invitedEmail,
		InvitedBy:      invitedByID,
		Role:           role,
		DepartmentID:   departmentID,
		BranchID:       branchID,
		Status:         "PENDING",
		Token:          &token,
		ExpiresAt:      time.Now().Add(72 * time.Hour),
	}

	if err := s.db.Create(inv).Error; err != nil {
		return nil, fmt.Errorf("failed to create invitation: %w", err)
	}

	// Create in-app notification for the invitee.
	if err := s.createInvitationNotification(inv); err != nil {
		// Non-fatal — invitation is already persisted.
		log.Printf("[InvitationService] notification creation failed for invitation %s: %v", inv.ID, err)
	}

	return inv, nil
}

// AcceptInvitation accepts a pending invitation by token.
// The calling user must be the invited user.
func (s *InvitationService) AcceptInvitation(token, acceptingUserID string) error {
	inv, err := s.loadPendingByToken(token)
	if err != nil {
		return err
	}

	// Security: only the invited user may accept.
	if inv.InvitedUserID == nil || *inv.InvitedUserID != acceptingUserID {
		return errors.New("this invitation was not sent to your account")
	}

	// Add user to the organization.
	orgSvc := NewOrganizationService(s.db)
	if err := orgSvc.AddMemberWithDepartment(inv.OrganizationID, acceptingUserID, inv.Role, inv.DepartmentID); err != nil {
		return fmt.Errorf("failed to add user to organization: %w", err)
	}

	// Set branch if present.
	if inv.BranchID != nil && *inv.BranchID != "" {
		s.db.Table("organization_members").
			Where("organization_id = ? AND user_id = ?", inv.OrganizationID, acceptingUserID).
			Update("branch_id", inv.BranchID)
	}

	now := time.Now()
	return s.db.Model(inv).Updates(map[string]interface{}{
		"status":      "ACCEPTED",
		"accepted_at": now,
	}).Error
}

// DeclineInvitation declines a pending invitation by token.
func (s *InvitationService) DeclineInvitation(token, decliningUserID string) error {
	inv, err := s.loadPendingByToken(token)
	if err != nil {
		return err
	}

	if inv.InvitedUserID == nil || *inv.InvitedUserID != decliningUserID {
		return errors.New("this invitation was not sent to your account")
	}

	now := time.Now()
	return s.db.Model(inv).Updates(map[string]interface{}{
		"status":      "DECLINED",
		"declined_at": now,
	}).Error
}

// CancelInvitation cancels a pending invitation by ID (admin action).
func (s *InvitationService) CancelInvitation(invitationID, adminUserID, orgID string) error {
	var inv models.OrganizationInvitation
	if err := s.db.Where("id = ? AND organization_id = ?", invitationID, orgID).
		First(&inv).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("invitation not found")
		}
		return fmt.Errorf("failed to load invitation: %w", err)
	}
	if strings.ToUpper(inv.Status) != "PENDING" {
		return fmt.Errorf("invitation cannot be cancelled (current status: %s)", inv.Status)
	}
	return s.db.Model(&inv).Update("status", "CANCELLED").Error
}

// ListOrgInvitations returns all invitations for the given organization, newest first.
func (s *InvitationService) ListOrgInvitations(orgID string) ([]models.OrganizationInvitation, error) {
	var invs []models.OrganizationInvitation
	err := s.db.
		Preload("InvitedUser").
		Preload("InvitedByUser").
		Where("organization_id = ?", orgID).
		Order("created_at DESC").
		Find(&invs).Error
	return invs, err
}

// ListPendingInvitationsForUser returns active pending invitations for the given user.
func (s *InvitationService) ListPendingInvitationsForUser(userID string) ([]models.OrganizationInvitation, error) {
	// Expire stale ones first so the list is always accurate.
	s.expireForUser(userID)

	var invs []models.OrganizationInvitation
	err := s.db.
		Preload("Organization").
		Preload("InvitedByUser").
		Where("invited_user_id = ? AND UPPER(status) = 'PENDING'", userID).
		Order("created_at DESC").
		Find(&invs).Error
	return invs, err
}

// ExpireStaleInvitations bulk-expires all overdue pending invitations.
// Called from a background goroutine in main.go.
func (s *InvitationService) ExpireStaleInvitations() error {
	result := s.db.Model(&models.OrganizationInvitation{}).
		Where("UPPER(status) = 'PENDING' AND expires_at < ?", time.Now()).
		Update("status", "EXPIRED")
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected > 0 {
		log.Printf("[InvitationExpiry] expired %d stale invitations", result.RowsAffected)
	}
	return nil
}

// --- helpers ---

// loadPendingByToken loads an invitation by token, enforcing that it is still
// pending and has not expired.
func (s *InvitationService) loadPendingByToken(token string) (*models.OrganizationInvitation, error) {
	if token == "" {
		return nil, errors.New("token is required")
	}
	var inv models.OrganizationInvitation
	if err := s.db.Where("token = ?", token).First(&inv).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invitation not found")
		}
		return nil, fmt.Errorf("failed to load invitation: %w", err)
	}
	if strings.ToUpper(inv.Status) == "EXPIRED" || time.Now().After(inv.ExpiresAt) {
		// Mark expired if not already.
		if strings.ToUpper(inv.Status) == "PENDING" {
			s.db.Model(&inv).Update("status", "EXPIRED")
		}
		return nil, errors.New("this invitation has expired — ask the admin to send a new one")
	}
	if strings.ToUpper(inv.Status) != "PENDING" {
		return nil, fmt.Errorf("invitation is no longer valid (status: %s)", inv.Status)
	}
	return &inv, nil
}

// expireForUser eagerly expires overdue invitations for a specific user before
// listing, so the invitee never sees expired items as pending.
func (s *InvitationService) expireForUser(userID string) {
	s.db.Model(&models.OrganizationInvitation{}).
		Where("invited_user_id = ? AND UPPER(status) = 'PENDING' AND expires_at < ?", userID, time.Now()).
		Update("status", "EXPIRED")
}

// createInvitationNotification inserts an in-app notification for the invitee.
func (s *InvitationService) createInvitationNotification(inv *models.OrganizationInvitation) error {
	if inv.InvitedUserID == nil {
		return nil // no platform user to notify
	}

	// Resolve org name and inviter name for the notification body.
	var org models.Organization
	var inviter models.User
	s.db.Select("name").Where("id = ?", inv.OrganizationID).First(&org)
	s.db.Select("name").Where("id = ?", inv.InvitedBy).First(&inviter)

	orgName := org.Name
	if orgName == "" {
		orgName = "an organization"
	}
	inviterName := inviter.Name
	if inviterName == "" {
		inviterName = "An admin"
	}

	token := ""
	if inv.Token != nil {
		token = *inv.Token
	}

	notification := &models.Notification{
		ID:             uuid.New().String(),
		OrganizationID: inv.OrganizationID,
		RecipientID:    *inv.InvitedUserID,
		Type:           "org_invitation",
		DocumentID:     inv.ID,
		DocumentType:   "invitation",
		EntityID:       inv.ID,
		EntityType:     "invitation",
		Subject:        fmt.Sprintf("You have been invited to join %s", orgName),
		Body: fmt.Sprintf(
			"%s has invited you to join %s as %s. This invitation expires in 72 hours.",
			inviterName, orgName, inv.Role,
		),
		RelatedUserID:   inv.InvitedBy,
		RelatedUserName: inviterName,
		Importance:      "HIGH",
		Sent:            false,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Embed accept/decline token in QuickAction so the frontend can render buttons.
	if token != "" {
		notification.Message = token // reuse Message field as token carrier for quick-action rendering
	}

	return s.db.Create(notification).Error
}
