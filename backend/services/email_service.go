package services

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"os"
	"strings"
	"time"

	"github.com/liyali/liyali-gateway/logging"
	"github.com/resend/resend-go/v3"
)

// EmailService handles outbound email delivery via Resend.
// Set RESEND_API_KEY to enable real sends; omit it to run in stub/log-only mode.
type EmailService struct {
	client      *resend.Client
	enabled     bool
	fromAddress string
	frontendURL string
}

func NewEmailService() *EmailService {
	apiKey := os.Getenv("RESEND_API_KEY")
	fromAddress := os.Getenv("EMAIL_FROM")
	frontendURL := os.Getenv("FRONTEND_URL")

	if fromAddress == "" {
		fromAddress = "noreply@liyali.com"
	}

	// Use only the first URL when FRONTEND_URL is comma-separated
	if frontendURL != "" {
		frontendURL = strings.TrimSpace(strings.Split(frontendURL, ",")[0])
	}

	enabled := apiKey != ""

	if !enabled {
		log.Printf("[EmailService] RESEND_API_KEY not set — running in stub mode (emails will be logged, not sent)")
	} else {
		log.Printf("[EmailService] Resend client initialised (from: %s)", fromAddress)
	}

	return &EmailService{
		client:      resend.NewClient(apiKey),
		enabled:     enabled,
		fromAddress: fromAddress,
		frontendURL: frontendURL,
	}
}

// sendEmail dispatches a single HTML email through the Resend API.
func (e *EmailService) sendEmail(to, subject, htmlBody string) error {
	if !e.enabled {
		return fmt.Errorf("email service not enabled (RESEND_API_KEY not configured)")
	}

	params := &resend.SendEmailRequest{
		From:    e.fromAddress,
		To:      []string{to},
		Subject: subject,
		Html:    htmlBody,
	}

	sent, err := e.client.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("resend API error: %w", err)
	}

	log.Printf("[EmailService] email dispatched id=%s to=%s subject=%q", sent.Id, to, subject)
	return nil
}

// SendInvitationEmail sends an invitation email to the invitee.
func (e *EmailService) SendInvitationEmail(
	toEmail, inviterName, orgName, role, token string,
	expiresAt time.Time,
) error {
	if !e.enabled {
		log.Printf("[EmailStub] invitation email → %s | org=%s role=%s token=%s expires=%s",
			toEmail, orgName, role, token, expiresAt.Format(time.RFC3339))
		return fmt.Errorf("email service not enabled - invitation email not sent to %s", toEmail)
	}

	acceptURL := fmt.Sprintf("%s/invitations/%s/accept", e.frontendURL, token)
	declineURL := fmt.Sprintf("%s/invitations/%s/decline", e.frontendURL, token)

	const tmpl = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>You're invited to {{.OrgName}}</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
    <div style="background-color: #f8f9fa; padding: 20px; border-radius: 8px;">
        <h1 style="color: #0066CC; margin-top: 0;">You're Invited!</h1>

        <p>Hello,</p>

        <p><strong>{{.InviterName}}</strong> has invited you to join <strong>{{.OrgName}}</strong> on Liyali Gateway as a <strong>{{.Role}}</strong>.</p>

        <p>This invitation will expire on <strong>{{.ExpiresAt}}</strong>.</p>

        <div style="margin: 30px 0;">
            <a href="{{.AcceptURL}}" style="display: inline-block; padding: 12px 30px; background-color: #0066CC; color: white; text-decoration: none; border-radius: 5px; margin-right: 10px;">Accept Invitation</a>
            <a href="{{.DeclineURL}}" style="display: inline-block; padding: 12px 30px; background-color: #6c757d; color: white; text-decoration: none; border-radius: 5px;">Decline</a>
        </div>

        <p style="color: #666; font-size: 14px;">If the buttons don't work, copy and paste this link into your browser:</p>
        <p style="color: #666; font-size: 12px; word-break: break-all;">{{.AcceptURL}}</p>

        <hr style="border: none; border-top: 1px solid #ddd; margin: 30px 0;">

        <p style="color: #999; font-size: 12px;">This is an automated message from Liyali Gateway. Please do not reply to this email.</p>
    </div>
</body>
</html>`

	t, err := template.New("invitation").Parse(tmpl)
	if err != nil {
		logging.WithError(err).Error("failed to parse invitation email template")
		return fmt.Errorf("failed to parse email template: %w", err)
	}

	data := struct {
		OrgName, InviterName, Role, ExpiresAt, AcceptURL, DeclineURL string
	}{
		OrgName:     orgName,
		InviterName: inviterName,
		Role:        role,
		ExpiresAt:   expiresAt.Format("January 2, 2006 at 3:04 PM MST"),
		AcceptURL:   acceptURL,
		DeclineURL:  declineURL,
	}

	var body bytes.Buffer
	if err := t.Execute(&body, data); err != nil {
		logging.WithError(err).Error("failed to execute invitation email template")
		return fmt.Errorf("failed to render email template: %w", err)
	}

	subject := fmt.Sprintf("You're invited to join %s on Liyali Gateway", orgName)
	if err := e.sendEmail(toEmail, subject, body.String()); err != nil {
		logging.WithFields(map[string]interface{}{
			"recipient": toEmail,
			"org":       orgName,
		}).WithError(err).Error("failed to send invitation email")
		return fmt.Errorf("failed to send invitation email: %w", err)
	}

	return nil
}

// SendPasswordResetEmail sends a password reset email.
func (e *EmailService) SendPasswordResetEmail(toEmail, resetToken string, expiresAt time.Time) error {
	if !e.enabled {
		log.Printf("[EmailStub] password reset email → %s | token=%s expires=%s",
			toEmail, resetToken, expiresAt.Format(time.RFC3339))
		return fmt.Errorf("email service not enabled - password reset email not sent to %s", toEmail)
	}

	resetURL := fmt.Sprintf("%s/reset-password?token=%s", e.frontendURL, resetToken)

	const tmpl = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Reset Your Password</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
    <div style="background-color: #f8f9fa; padding: 20px; border-radius: 8px;">
        <h1 style="color: #0066CC; margin-top: 0;">Reset Your Password</h1>

        <p>Hello,</p>

        <p>We received a request to reset your password for your Liyali Gateway account.</p>

        <p>This password reset link will expire on <strong>{{.ExpiresAt}}</strong>.</p>

        <div style="margin: 30px 0;">
            <a href="{{.ResetURL}}" style="display: inline-block; padding: 12px 30px; background-color: #0066CC; color: white; text-decoration: none; border-radius: 5px;">Reset Password</a>
        </div>

        <p style="color: #666; font-size: 14px;">If the button doesn't work, copy and paste this link into your browser:</p>
        <p style="color: #666; font-size: 12px; word-break: break-all;">{{.ResetURL}}</p>

        <p style="color: #dc3545; margin-top: 30px;"><strong>If you didn't request this password reset, please ignore this email.</strong> Your password will remain unchanged.</p>

        <hr style="border: none; border-top: 1px solid #ddd; margin: 30px 0;">

        <p style="color: #999; font-size: 12px;">This is an automated message from Liyali Gateway. Please do not reply to this email.</p>
    </div>
</body>
</html>`

	t, err := template.New("passwordReset").Parse(tmpl)
	if err != nil {
		logging.WithError(err).Error("failed to parse password reset email template")
		return fmt.Errorf("failed to parse email template: %w", err)
	}

	data := struct{ ResetURL, ExpiresAt string }{
		ResetURL:  resetURL,
		ExpiresAt: expiresAt.Format("January 2, 2006 at 3:04 PM MST"),
	}

	var body bytes.Buffer
	if err := t.Execute(&body, data); err != nil {
		logging.WithError(err).Error("failed to execute password reset email template")
		return fmt.Errorf("failed to render email template: %w", err)
	}

	if err := e.sendEmail(toEmail, "Reset Your Liyali Gateway Password", body.String()); err != nil {
		logging.WithFields(map[string]interface{}{
			"recipient": toEmail,
		}).WithError(err).Error("failed to send password reset email")
		return fmt.Errorf("failed to send password reset email: %w", err)
	}

	return nil
}
