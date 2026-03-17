package services

import (
	"log"
	"os"
	"time"
)

// EmailService stubs outbound email delivery.
// Set EMAIL_ENABLED=true and configure SMTP/SendGrid env vars to enable real sends.
type EmailService struct {
	enabled bool
}

func NewEmailService() *EmailService {
	return &EmailService{
		enabled: os.Getenv("EMAIL_ENABLED") == "true",
	}
}

// SendInvitationEmail sends (or stubs) an invitation email to the invitee.
// Failures are non-fatal — the invitation is already persisted in the DB.
func (e *EmailService) SendInvitationEmail(
	toEmail, inviterName, orgName, role, token string,
	expiresAt time.Time,
) error {
	if !e.enabled {
		log.Printf("[EmailStub] invitation email → %s | org=%s role=%s token=%s expires=%s",
			toEmail, orgName, role, token, expiresAt.Format(time.RFC3339))
		return nil
	}
	// TODO: replace stub with SMTP / SendGrid / Resend implementation.
	// Build and send an HTML email with Accept/Decline CTA buttons whose hrefs
	// include the token, e.g.:
	//   {FRONTEND_URL}/invitations/{token}/accept
	//   {FRONTEND_URL}/invitations/{token}/decline
	log.Printf("[EmailService] (not implemented) invitation email → %s", toEmail)
	return nil
}
