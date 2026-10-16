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

// SendEmail delivers through SendGrid (default) or Resend when EMAIL_PROVIDER=resend.
func SendEmail(to, subject, htmlBody string) error {
	if strings.EqualFold(os.Getenv("EMAIL_PROVIDER"), "resend") {
		return sendViaResend(to, subject, htmlBody)
	}
	return sendViaSendGrid(to, subject, htmlBody)
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

func statusURL() string {
	return config.FrontendURL() + "/application-status"
}

func SendConfirmationEmail(to, name, jobTitle, applicationID string) error {
	link := statusURL() + "?" + url.Values{"email": {to}, "applicationId": {applicationID}}.Encode()
	body := fmt.Sprintf(`<p>Thank you for applying to the <strong>%s</strong> position.</p>
    <p>We have received your application and will review it shortly. You can check its status anytime.</p>
    <p style="font-size: 12px; color: #666;"><strong>Application ID:</strong> %s<br>
    You'll need this along with your email to check your status later.</p>`,
		html.EscapeString(jobTitle), html.EscapeString(applicationID))

	return SendEmail(to, "Application Received - "+jobTitle,
		renderEmail("Hello "+name+",", "#2563eb", body, "Check Application Status", link))
}

func SendShortlistEmail(to, name, jobTitle string) error {
	body := fmt.Sprintf(`<p>Great news! You have been <strong style="color: #10b981;">shortlisted</strong> for the <strong>%s</strong> position.</p>
    <p>Our team will be in touch soon about the next steps.</p>`, html.EscapeString(jobTitle))

	return SendEmail(to, "Congratulations! You've been Shortlisted",
		renderEmail("Congratulations "+name+"!", "#10b981", body, "View Application Status", statusURL()))
}

func SendRejectionEmail(to, name, jobTitle string) error {
	body := fmt.Sprintf(`<p>Thank you for your interest in the <strong>%s</strong> position.</p>
    <p>After careful consideration, we have decided to move forward with other candidates at this time.
    We encourage you to apply for future opportunities.</p>`, html.EscapeString(jobTitle))

	return SendEmail(to, "Application Update - "+jobTitle,
		renderEmail("Hello "+name+",", "#2563eb", body, "View Application Status", statusURL()))
}

func SendNewMessageEmail(to, name, jobTitle, message string) error {
	body := fmt.Sprintf(`<p>You have a new message about your application for <strong>%s</strong>.</p>
    <div style="background-color: #f3f4f6; padding: 15px; border-radius: 5px; margin: 20px 0; white-space: pre-wrap;">%s</div>
    <p>You can reply from your application status portal.</p>`, html.EscapeString(jobTitle), html.EscapeString(message))

	return SendEmail(to, "New Message About Your Application - "+jobTitle,
		renderEmail("Hello "+name+",", "#2563eb", body, "View Application Status", statusURL()))
}
