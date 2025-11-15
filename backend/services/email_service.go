package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

// EmailProvider represents the email service provider
type EmailProvider string

const (
	ProviderResend  EmailProvider = "resend"
	ProviderSendGrid EmailProvider = "sendgrid"
)

// isResendVerifiedDomain checks if email uses Resend's test domain
func isResendVerifiedDomain(email string) bool {
	return strings.HasSuffix(email, "@resend.dev")
}

type EmailRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
}

// getEmailProvider returns the configured email provider (default: sendgrid)
func getEmailProvider() EmailProvider {
	provider := strings.ToLower(os.Getenv("EMAIL_PROVIDER"))
	if provider == string(ProviderResend) {
		return ProviderResend
	}
	// Default to SendGrid (better for testing - no restrictions)
	return ProviderSendGrid
}

// sendEmailViaSendGrid sends email using SendGrid API
func sendEmailViaSendGrid(to, subject, htmlBody string) error {
	apiKey := os.Getenv("SENDGRID_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("SENDGRID_API_KEY not set")
	}

	fromEmail := os.Getenv("SENDGRID_FROM_EMAIL")
	if fromEmail == "" {
		// SendGrid requires a verified sender email
		// User must verify a single sender email in SendGrid dashboard
		// This can be a personal Gmail/Outlook/etc. - no domain needed!
		return fmt.Errorf("SENDGRID_FROM_EMAIL must be set to a verified sender email. Verify a single sender in SendGrid dashboard (Settings → Sender Authentication → Verify a Single Sender)")
	}

	log.Printf("Attempting to send email via SendGrid to %s from %s", to, fromEmail)

	// SendGrid API format
	type SendGridContent struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	}
	type SendGridPersonalization struct {
		To []map[string]string `json:"to"`
	}
	type SendGridEmail struct {
		Personalizations []SendGridPersonalization `json:"personalizations"`
		From             map[string]string          `json:"from"`
		Subject          string                     `json:"subject"`
		Content          []SendGridContent          `json:"content"`
	}

	emailReq := SendGridEmail{
		Personalizations: []SendGridPersonalization{
			{
				To: []map[string]string{{"email": to}},
			},
		},
		From: map[string]string{
			"email": fromEmail,
		},
		Subject: subject,
		Content: []SendGridContent{
			{
				Type:  "text/html",
				Value: htmlBody,
			},
		},
	}

	jsonData, err := json.Marshal(emailReq)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://api.sendgrid.com/v3/mail/send", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("SendGrid API error: %v", err)
		return err
	}
	defer resp.Body.Close()

	// Read response body for debugging
	body, _ := io.ReadAll(resp.Body)
	
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		log.Printf("SendGrid email sending failed: Status %d, Response: %s", resp.StatusCode, string(body))
		return fmt.Errorf("email sending failed: %d - %s", resp.StatusCode, string(body))
	}

	log.Printf("Email sent successfully via SendGrid to %s", to)
	return nil
}

// sendEmailViaResend sends email using Resend API
func sendEmailViaResend(to, subject, htmlBody string) error {
	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("RESEND_API_KEY not set")
	}

	fromEmail := os.Getenv("RESEND_FROM_EMAIL")
	// Resend requires verified domains. For testing, use their default domain
	// If using a custom email, it must be verified in Resend dashboard
	if fromEmail == "" || !isResendVerifiedDomain(fromEmail) {
		// Use Resend's default domain for testing (no verification needed)
		fromEmail = "onboarding@resend.dev"
		log.Printf("WARNING: Using Resend test domain. Set RESEND_FROM_EMAIL to a verified domain for production")
	}
	
	log.Printf("Attempting to send email via Resend to %s from %s", to, fromEmail)

	emailReq := EmailRequest{
		From:    fromEmail,
		To:      []string{to},
		Subject: subject,
		HTML:    htmlBody,
	}

	jsonData, err := json.Marshal(emailReq)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Resend API error: %v", err)
		return err
	}
	defer resp.Body.Close()

	// Read response body for debugging
	body, _ := io.ReadAll(resp.Body)
	
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		log.Printf("Resend email sending failed: Status %d, Response: %s", resp.StatusCode, string(body))
		return fmt.Errorf("email sending failed: %d - %s", resp.StatusCode, string(body))
	}

	log.Printf("Email sent successfully via Resend to %s", to)
	return nil
}

// sendEmail sends email using the configured provider
func sendEmail(to, subject, htmlBody string) error {
	provider := getEmailProvider()
	
	switch provider {
	case ProviderResend:
		return sendEmailViaResend(to, subject, htmlBody)
	case ProviderSendGrid:
		return sendEmailViaSendGrid(to, subject, htmlBody)
	default:
		return fmt.Errorf("unknown email provider: %s", provider)
	}
}

func SendConfirmationEmail(to, name, jobTitle string) error {
	subject := "Application Received - " + jobTitle
	html := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<meta charset="UTF-8">
		</head>
		<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
			<div style="max-width: 600px; margin: 0 auto; padding: 20px;">
				<h2 style="color: #2563eb;">Hello %s,</h2>
				<p>Thank you for applying to the <strong>%s</strong> position.</p>
				<p>We have received your application and will review it shortly.</p>
				<p>You will hear from us soon!</p>
				<br>
				<p>Best regards,<br>The Hiring Team</p>
			</div>
		</body>
		</html>
	`, name, jobTitle)

	return sendEmail(to, subject, html)
}

func SendShortlistEmail(to, name string) error {
	subject := "Congratulations! You've been shortlisted"
	html := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<meta charset="UTF-8">
		</head>
		<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
			<div style="max-width: 600px; margin: 0 auto; padding: 20px;">
				<h2 style="color: #16a34a;">Hello %s,</h2>
				<p>Great news! Your application has been shortlisted.</p>
				<p>We were impressed with your qualifications and would like to invite you for an interview.</p>
				<p>We will contact you shortly with further details.</p>
				<br>
				<p>Best regards,<br>The Hiring Team</p>
			</div>
		</body>
		</html>
	`, name)

	return sendEmail(to, subject, html)
}

func SendRejectionEmail(to, name string) error {
	subject := "Application Update"
	html := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<meta charset="UTF-8">
		</head>
		<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
			<div style="max-width: 600px; margin: 0 auto; padding: 20px;">
				<h2 style="color: #333;">Hello %s,</h2>
				<p>Thank you for your interest in our company and for taking the time to apply.</p>
				<p>After careful consideration, we regret to inform you that we will not be moving forward with your application at this time.</p>
				<p>We encourage you to apply for future opportunities that match your skills and experience.</p>
				<p>We wish you the best in your job search.</p>
				<br>
				<p>Best regards,<br>The Hiring Team</p>
			</div>
		</body>
		</html>
	`, name)

	return sendEmail(to, subject, html)
}

