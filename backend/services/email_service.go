package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"ats-backend/config"
)

var mailClient = &http.Client{Timeout: 15 * time.Second}

// SendEmail delivers through SendGrid (default), or Resend / SMTP when EMAIL_PROVIDER says so.
func SendEmail(to, subject, htmlBody string) error {
	switch strings.ToLower(os.Getenv("EMAIL_PROVIDER")) {
	case "resend":
		return sendViaResend(to, subject, htmlBody)
	case "smtp":
		return sendViaSMTP(to, subject, htmlBody)
	default:
		return sendViaSendGrid(to, subject, htmlBody)
	}
}

func sendViaResend(to, subject, htmlBody string) error {
	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		return errors.New("RESEND_API_KEY not set")
	}
	// onboarding@resend.dev works without a verified domain, but only delivers to your own address.
	from := config.GetEnv("RESEND_FROM_EMAIL", "onboarding@resend.dev")

	payload := map[string]any{"from": from, "to": []string{to}, "subject": subject, "html": htmlBody}
	return postJSON("https://api.resend.com/emails", apiKey, payload)
}

func sendViaSendGrid(to, subject, htmlBody string) error {
	apiKey := os.Getenv("SENDGRID_API_KEY")
	from := os.Getenv("SENDGRID_FROM_EMAIL")
	if apiKey == "" || from == "" {
		return errors.New("SENDGRID_API_KEY and SENDGRID_FROM_EMAIL must be set")
	}

	payload := map[string]any{
		"personalizations": []map[string]any{{"to": []map[string]string{{"email": to}}}},
		"from":             map[string]string{"email": from},
		"subject":          subject,
		"content":          []map[string]string{{"type": "text/html", "value": htmlBody}},
	}
	return postJSON("https://api.sendgrid.com/v3/mail/send", apiKey, payload)
}

func postJSON(endpoint, apiKey string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := mailClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		detail, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("email provider returned %d: %s", resp.StatusCode, detail)
	}
	return nil
}

// renderEmail escapes every caller-supplied value; bodyHTML must already be safe.
func renderEmail(greeting, color, bodyHTML, ctaLabel, ctaURL string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html><head><meta charset="UTF-8"></head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
  <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
    <h2 style="color: %s;">%s</h2>
    %s
    <p style="text-align: center; margin: 20px 0;">
      <a href="%s" style="background-color: %s; color: white; padding: 12px 24px; text-decoration: none; border-radius: 5px; display: inline-block;">%s</a>
    </p>
    <p>Best regards,<br>The Hiring Team</p>
  </div>
</body></html>`,
		color, html.EscapeString(greeting), bodyHTML, html.EscapeString(ctaURL), color, html.EscapeString(ctaLabel))
}

// statusURL pre-fills the candidate portal so links in every email open the application directly.
func statusURL(email, applicationID string) string {
	return config.FrontendURL() + "/application-status?" + url.Values{"email": {email}, "applicationId": {applicationID}}.Encode()
}

func SendConfirmationEmail(to, name, jobTitle, applicationID string) error {
	link := statusURL(to, applicationID)
	body := fmt.Sprintf(`<p>Thank you for applying to the <strong>%s</strong> position.</p>
    <p>We have received your application and will review it shortly. You can check its status anytime.</p>
    <p style="font-size: 12px; color: #666;"><strong>Application ID:</strong> %s<br>
    You'll need this along with your email to check your status later.</p>`,
		html.EscapeString(jobTitle), html.EscapeString(applicationID))

	return SendEmail(to, "Application Received - "+jobTitle,
		renderEmail("Hello "+name+",", "#2563eb", body, "Check Application Status", link))
}

func SendShortlistEmail(to, name, jobTitle, applicationID string) error {
	body := fmt.Sprintf(`<p>Great news! You have been <strong style="color: #10b981;">shortlisted</strong> for the <strong>%s</strong> position.</p>
    <p>Our team will be in touch soon about the next steps.</p>`, html.EscapeString(jobTitle))

	return SendEmail(to, "Congratulations! You've been Shortlisted",
		renderEmail("Congratulations "+name+"!", "#10b981", body, "View Application Status", statusURL(to, applicationID)))
}

func SendRejectionEmail(to, name, jobTitle, applicationID string) error {
	body := fmt.Sprintf(`<p>Thank you for your interest in the <strong>%s</strong> position.</p>
    <p>After careful consideration, we have decided to move forward with other candidates at this time.
    We encourage you to apply for future opportunities.</p>`, html.EscapeString(jobTitle))

	return SendEmail(to, "Application Update - "+jobTitle,
		renderEmail("Hello "+name+",", "#2563eb", body, "View Application Status", statusURL(to, applicationID)))
}

var stageEmails = map[string]struct{ subject, heading, body string }{
	"interview": {"Interview invitation - %s", "Good news, %s!", `<p>You've moved to the <strong>interview</strong> stage for the <strong>%s</strong> position.</p>
    <p>The hiring team will contact you shortly to schedule a time.</p>`},
	"offer": {"Offer update - %s", "Congratulations %s!", `<p>We're pleased to let you know we're preparing an <strong>offer</strong> for the <strong>%s</strong> position.</p>
    <p>The hiring team will be in touch with the details.</p>`},
	"hired": {"Welcome aboard - %s", "Welcome aboard, %s!", `<p>You've been <strong>hired</strong> for the <strong>%s</strong> position. We're excited to have you join.</p>`},
}

// SendStageEmail notifies a candidate that they advanced to interview, offer or hired.
func SendStageEmail(to, name, jobTitle, applicationID, stage string) error {
	tpl, ok := stageEmails[stage]
	if !ok {
		return nil
	}
	body := fmt.Sprintf(tpl.body, html.EscapeString(jobTitle))
	return SendEmail(to, fmt.Sprintf(tpl.subject, jobTitle),
		renderEmail(fmt.Sprintf(tpl.heading, name), "#10b981", body, "View Application Status", statusURL(to, applicationID)))
}

func SendNewMessageEmail(to, name, jobTitle, applicationID, message string) error {
	body := fmt.Sprintf(`<p>You have a new message about your application for <strong>%s</strong>.</p>
    <div style="background-color: #f3f4f6; padding: 15px; border-radius: 5px; margin: 20px 0; white-space: pre-wrap;">%s</div>
    <p>You can reply from your application status portal.</p>`, html.EscapeString(jobTitle), html.EscapeString(message))

	return SendEmail(to, "New Message About Your Application - "+jobTitle,
		renderEmail("Hello "+name+",", "#2563eb", body, "View Application Status", statusURL(to, applicationID)))
}

// SendCandidateReplyEmail tells a recruiter that a candidate wrote back.
func SendCandidateReplyEmail(to, candidateName, jobTitle, message string) error {
	body := fmt.Sprintf(`<p><strong>%s</strong> sent a message about the <strong>%s</strong> position.</p>
    <div style="background-color: #f3f4f6; padding: 15px; border-radius: 5px; margin: 20px 0; white-space: pre-wrap;">%s</div>`,
		html.EscapeString(candidateName), html.EscapeString(jobTitle), html.EscapeString(message))

	return SendEmail(to, "New candidate message - "+jobTitle,
		renderEmail("New message from "+candidateName, "#2563eb", body, "Open Applications", config.FrontendURL()+"/admin/dashboard/applications"))
}

func SendPasswordResetEmail(to, name, link string) error {
	body := `<p>We received a request to reset your password. The link below works once and expires in 1 hour.</p>
    <p style="font-size: 12px; color: #666;">If you didn't ask for this, you can ignore this email; your password won't change.</p>`

	return SendEmail(to, "Reset your password",
		renderEmail("Hello "+name+",", "#2563eb", body, "Reset Password", link))
}

func SendTeamInviteEmail(to, inviterName, companyName, role, link string) error {
	body := fmt.Sprintf(`<p><strong>%s</strong> invited you to join <strong>%s</strong> on the ATS platform as a <strong>%s</strong>.</p>
    <p style="font-size: 12px; color: #666;">The invite link expires in 7 days.</p>`,
		html.EscapeString(inviterName), html.EscapeString(companyName), html.EscapeString(role))

	return SendEmail(to, "You're invited to join "+companyName,
		renderEmail("You've been invited", "#2563eb", body, "Accept Invite", link))
}

func SendVerificationEmail(to, name, link string) error {
	body := `<p>Confirm your email address to finish setting up your ATS account. The link expires in 24 hours.</p>
    <p style="font-size: 12px; color: #666;">If you didn't sign up, you can ignore this email.</p>`
	return SendEmail(to, "Verify your email",
		renderEmail("Welcome, "+name+"!", "#2563eb", body, "Verify Email", link))
}

// SendAccountExistsEmail answers a sign-up attempt for an email that already has an account.
func SendAccountExistsEmail(to, name string) error {
	body := `<p>Someone tried to create a new account with this email address, but you already have one.</p>
    <p>If that was you, just log in, or reset your password if you've forgotten it. Otherwise you can ignore this email.</p>`
	return SendEmail(to, "You already have an account",
		renderEmail("Hello "+name+",", "#2563eb", body, "Log In", config.FrontendURL()+"/admin/login"))
}
