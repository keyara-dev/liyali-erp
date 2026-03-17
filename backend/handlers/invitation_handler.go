package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/logging"
	"github.com/liyali/liyali-gateway/middleware"
	"github.com/liyali/liyali-gateway/services"
	"github.com/liyali/liyali-gateway/utils"
)

// ─── Admin-side handlers ─────────────────────────────────────────────────────

// LookupUserByEmail checks whether an email address belongs to an existing platform
// user and whether that user is already a member of the current organisation.
// The frontend uses this before showing the create-vs-invite decision banner.
//
// GET /api/v1/organization/users/lookup?email=
func LookupUserByEmail(c *fiber.Ctx) error {
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return utils.SendUnauthorizedError(c, "Organization context required")
	}

	email := c.Query("email")
	if email == "" {
		return utils.SendBadRequestError(c, "email query parameter is required")
	}

	svc := services.NewUserService(config.DB)
	result, err := svc.LookupUserByEmailForOrg(tenant.OrganizationID, email)
	if err != nil {
		logging.LogError(c, err, "email_lookup_failed", nil)
		return utils.SendInternalError(c, "Failed to look up email", err)
	}

	payload := fiber.Map{
		"exists":               result.User != nil,
		"isOrgMember":          result.IsMember,
		"hasPendingInvitation": false,
	}

	if result.User != nil {
		payload["userId"] = result.User.ID
		payload["name"] = result.User.Name
		payload["email"] = result.User.Email

		// Check for an existing pending invite.
		if !result.IsMember {
			invSvc := services.NewInvitationService(config.DB)
			pending, _ := invSvc.ListOrgInvitations(tenant.OrganizationID)
			for _, inv := range pending {
				if inv.InvitedEmail == email && inv.Status == "pending" {
					payload["hasPendingInvitation"] = true
					break
				}
			}
		}
	}

	return utils.SendSimpleSuccess(c, payload, "Email lookup successful")
}

// SendInvitation creates a new invitation for an existing platform user.
//
// POST /api/v1/organization/invitations
// Body: { email, role?, department_id?, branch_id? }
func SendInvitation(c *fiber.Ctx) error {
	logger := logging.FromContext(c)

	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return utils.SendUnauthorizedError(c, "Organization context required")
	}

	adminID, _ := c.Locals("userID").(string)

	var req struct {
		Email        string  `json:"email"`
		Role         string  `json:"role"`
		DepartmentID *string `json:"department_id"`
		BranchID     *string `json:"branch_id"`
	}
	if err := c.BodyParser(&req); err != nil {
		return utils.SendBadRequestError(c, "Invalid request body")
	}
	if req.Email == "" {
		return utils.SendBadRequestError(c, "email is required")
	}

	// Validate department belongs to this org (if provided).
	if req.DepartmentID != nil && *req.DepartmentID != "" {
		var count int64
		config.DB.Table("organization_departments").
			Where("id = ? AND organization_id = ? AND is_active = true", *req.DepartmentID, tenant.OrganizationID).
			Count(&count)
		if count == 0 {
			return utils.SendBadRequestError(c, "Department not found in this organization")
		}
	}

	// Validate branch belongs to this org (if provided).
	if req.BranchID != nil && *req.BranchID != "" {
		var count int64
		config.DB.Table("organization_branches").
			Where("id = ? AND organization_id = ? AND is_active = true", *req.BranchID, tenant.OrganizationID).
			Count(&count)
		if count == 0 {
			return utils.SendBadRequestError(c, "Branch not found in this organization")
		}
	}

	invSvc := services.NewInvitationService(config.DB)
	inv, err := invSvc.SendInvitation(
		tenant.OrganizationID,
		adminID,
		req.Email,
		req.Role,
		req.DepartmentID,
		req.BranchID,
	)
	if err != nil {
		logging.LogError(c, err, "send_invitation_failed", fiber.Map{"email": req.Email})
		return utils.SendConflictError(c, err.Error())
	}

	// Stub email (non-fatal).
	emailSvc := services.NewEmailService()
	var token string
	if inv.Token != nil {
		token = *inv.Token
	}
	_ = emailSvc.SendInvitationEmail(inv.InvitedEmail, adminID, tenant.OrganizationID, inv.Role, token, inv.ExpiresAt)

	logging.AddFieldsToRequest(c, fiber.Map{
		"invitation_id":   inv.ID,
		"invited_email":   inv.InvitedEmail,
		"organization_id": tenant.OrganizationID,
	})
	logger.Info("invitation_sent")

	// Return the invitation including the token (creation-time only exposure).
	return utils.SendCreatedSuccess(c, fiber.Map{
		"id":           inv.ID,
		"invitedEmail": inv.InvitedEmail,
		"role":         inv.Role,
		"status":       inv.Status,
		"expiresAt":    inv.ExpiresAt,
		"token":        token, // frontend may need this for deep-link construction
	}, "Invitation sent successfully")
}

// ListOrgInvitations returns all invitations for the current organisation.
//
// GET /api/v1/organization/invitations
func ListOrgInvitations(c *fiber.Ctx) error {
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return utils.SendUnauthorizedError(c, "Organization context required")
	}

	invSvc := services.NewInvitationService(config.DB)
	invs, err := invSvc.ListOrgInvitations(tenant.OrganizationID)
	if err != nil {
		logging.LogError(c, err, "list_invitations_failed", nil)
		return utils.SendInternalError(c, "Failed to fetch invitations", err)
	}

	return utils.SendSimpleSuccess(c, invs, "Invitations retrieved successfully")
}

// CancelInvitation cancels a pending invitation (admin action).
//
// DELETE /api/v1/organization/invitations/:id
func CancelInvitation(c *fiber.Ctx) error {
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return utils.SendUnauthorizedError(c, "Organization context required")
	}

	adminID, _ := c.Locals("userID").(string)
	invID := c.Params("id")
	if invID == "" {
		return utils.SendBadRequestError(c, "Invitation ID is required")
	}

	invSvc := services.NewInvitationService(config.DB)
	if err := invSvc.CancelInvitation(invID, adminID, tenant.OrganizationID); err != nil {
		return utils.SendConflictError(c, err.Error())
	}

	return utils.SendSuccess(c, fiber.StatusOK, fiber.Map{"id": invID}, "Invitation cancelled", nil)
}

// ResendInvitation cancels the existing invite and creates a fresh one.
//
// POST /api/v1/organization/invitations/:id/resend
func ResendInvitation(c *fiber.Ctx) error {
	tenant, err := middleware.GetTenantContext(c)
	if err != nil {
		return utils.SendUnauthorizedError(c, "Organization context required")
	}

	adminID, _ := c.Locals("userID").(string)
	invID := c.Params("id")

	// Load the original invitation to copy its parameters.
	invSvc := services.NewInvitationService(config.DB)
	existing, err := invSvc.ListOrgInvitations(tenant.OrganizationID)
	if err != nil {
		return utils.SendInternalError(c, "Failed to load invitation", err)
	}

	var orig *struct {
		email        string
		role         string
		departmentID *string
		branchID     *string
	}
	for _, inv := range existing {
		if inv.ID == invID && inv.OrganizationID == tenant.OrganizationID {
			orig = &struct {
				email        string
				role         string
				departmentID *string
				branchID     *string
			}{
				email:        inv.InvitedEmail,
				role:         inv.Role,
				departmentID: inv.DepartmentID,
				branchID:     inv.BranchID,
			}
			break
		}
	}
	if orig == nil {
		return utils.SendNotFoundError(c, "Invitation not found")
	}

	// SendInvitation auto-cancels any existing pending invite for the same email.
	newInv, err := invSvc.SendInvitation(
		tenant.OrganizationID, adminID,
		orig.email, orig.role,
		orig.departmentID, orig.branchID,
	)
	if err != nil {
		logging.LogError(c, err, "resend_invitation_failed", fiber.Map{"original_id": invID})
		return utils.SendConflictError(c, err.Error())
	}

	emailSvc := services.NewEmailService()
	var token string
	if newInv.Token != nil {
		token = *newInv.Token
	}
	_ = emailSvc.SendInvitationEmail(newInv.InvitedEmail, adminID, tenant.OrganizationID, newInv.Role, token, newInv.ExpiresAt)

	return utils.SendCreatedSuccess(c, fiber.Map{
		"id":           newInv.ID,
		"invitedEmail": newInv.InvitedEmail,
		"status":       newInv.Status,
		"expiresAt":    newInv.ExpiresAt,
	}, "Invitation resent successfully")
}

// ─── Invitee-facing handlers ──────────────────────────────────────────────────

// GetMyPendingInvitations returns pending invitations for the currently logged-in user.
//
// GET /api/v1/invitations/pending
func GetMyPendingInvitations(c *fiber.Ctx) error {
	userID, _ := c.Locals("userID").(string)
	if userID == "" {
		return utils.SendUnauthorizedError(c, "Authentication required")
	}

	invSvc := services.NewInvitationService(config.DB)
	invs, err := invSvc.ListPendingInvitationsForUser(userID)
	if err != nil {
		logging.LogError(c, err, "list_pending_invitations_failed", nil)
		return utils.SendInternalError(c, "Failed to fetch invitations", err)
	}

	return utils.SendSimpleSuccess(c, invs, "Pending invitations retrieved")
}

// AcceptInvitation accepts a pending invitation by token.
//
// POST /api/v1/invitations/:token/accept
func AcceptInvitation(c *fiber.Ctx) error {
	userID, _ := c.Locals("userID").(string)
	if userID == "" {
		return utils.SendUnauthorizedError(c, "Authentication required")
	}

	token := c.Params("token")
	if token == "" {
		return utils.SendBadRequestError(c, "Token is required")
	}

	invSvc := services.NewInvitationService(config.DB)
	if err := invSvc.AcceptInvitation(token, userID); err != nil {
		logging.LogError(c, err, "accept_invitation_failed", fiber.Map{"token": token})
		return utils.SendConflictError(c, err.Error())
	}

	logging.AddFieldsToRequest(c, fiber.Map{"user_id": userID})
	logging.FromContext(c).Info("invitation_accepted")

	return utils.SendSuccess(c, fiber.StatusOK, nil, "Invitation accepted — welcome to the organization!", nil)
}

// DeclineInvitation declines a pending invitation by token.
//
// POST /api/v1/invitations/:token/decline
func DeclineInvitation(c *fiber.Ctx) error {
	userID, _ := c.Locals("userID").(string)
	if userID == "" {
		return utils.SendUnauthorizedError(c, "Authentication required")
	}

	token := c.Params("token")
	if token == "" {
		return utils.SendBadRequestError(c, "Token is required")
	}

	invSvc := services.NewInvitationService(config.DB)
	if err := invSvc.DeclineInvitation(token, userID); err != nil {
		logging.LogError(c, err, "decline_invitation_failed", fiber.Map{"token": token})
		return utils.SendConflictError(c, err.Error())
	}

	return utils.SendSuccess(c, fiber.StatusOK, nil, "Invitation declined", nil)
}
