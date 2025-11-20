package services

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

// SMSProvider represents the SMS service provider
type SMSProvider string

const (
	ProviderTwilio SMSProvider = "twilio"
)

// SendSMS sends SMS notification to candidate
// Currently supports Twilio. Returns error if SMS sending fails.
func SendSMS(to, message string) error {
	provider := getSMSProvider()
	
	switch provider {
	case ProviderTwilio:
		return sendSMSViaTwilio(to, message)
	default:
		// If SMS is not configured, log and return nil (don't fail the operation)
		log.Printf("SMS not configured. Skipping SMS to %s", to)
		return nil
	}
}

// getSMSProvider returns the configured SMS provider
func getSMSProvider() SMSProvider {
	provider := strings.ToLower(os.Getenv("SMS_PROVIDER"))
	if provider == string(ProviderTwilio) {
		return ProviderTwilio
	}
	return "" // No provider configured
}

// sendSMSViaTwilio sends SMS using Twilio API
func sendSMSViaTwilio(to, message string) error {
	accountSID := os.Getenv("TWILIO_ACCOUNT_SID")
	authToken := os.Getenv("TWILIO_AUTH_TOKEN")
	fromNumber := os.Getenv("TWILIO_FROM_NUMBER")

	if accountSID == "" || authToken == "" || fromNumber == "" {
		log.Printf("Twilio credentials not configured. Skipping SMS to %s", to)
		return nil // Don't fail if SMS is not configured
	}

	// Format phone number (remove any non-digit characters except +)
	to = formatPhoneNumber(to)
	if to == "" {
		return fmt.Errorf("invalid phone number: %s", to)
	}

	log.Printf("Attempting to send SMS via Twilio to %s from %s", to, fromNumber)

	// Twilio API endpoint
	url := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", accountSID)

	// Prepare form data
	data := map[string]string{
		"From": fromNumber,
		"To":   to,
		"Body": message,
	}

	// Convert to form data
	formData := make([]string, 0, len(data))
	for k, v := range data {
		formData = append(formData, fmt.Sprintf("%s=%s", k, v))
	}
	body := strings.Join(formData, "&")

	req, err := http.NewRequest("POST", url, bytes.NewBufferString(body))
	if err != nil {
		return err
	}

	// Set basic auth header
	req.SetBasicAuth(accountSID, authToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Twilio API error: %v", err)
		return err
	}
	defer resp.Body.Close()

	// Read response body for debugging
	responseBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("Twilio SMS sending failed: Status %d, Response: %s", resp.StatusCode, string(responseBody))
		return fmt.Errorf("SMS sending failed: %d - %s", resp.StatusCode, string(responseBody))
	}

	log.Printf("SMS sent successfully via Twilio to %s", to)
	return nil
}

// formatPhoneNumber formats phone number for Twilio (E.164 format)
func formatPhoneNumber(phone string) string {
	// Remove all non-digit characters except +
	formatted := ""
	for _, char := range phone {
		if char >= '0' && char <= '9' || char == '+' {
			formatted += string(char)
		}
	}

	// If no + prefix and starts with 0, assume local format and return empty (needs country code)
	if !strings.HasPrefix(formatted, "+") {
		// For now, return as-is. User should provide E.164 format
		// In production, you might want to add country code detection
		return formatted
	}

	return formatted
}

// SendStatusUpdateSMS sends SMS notification when application status changes
func SendStatusUpdateSMS(phone, candidateName, jobTitle, status string) error {
	if phone == "" {
		return nil // No phone number, skip SMS
	}

	var message string
	switch status {
	case "cv_viewed":
		message = fmt.Sprintf("Hi %s, your application for %s has been reviewed. We'll be in touch soon!", candidateName, jobTitle)
	case "shortlisted":
		message = fmt.Sprintf("Congratulations %s! You've been shortlisted for %s. We'll contact you soon with next steps.", candidateName, jobTitle)
	case "rejected":
		message = fmt.Sprintf("Hi %s, thank you for applying to %s. We've decided to move forward with other candidates at this time.", candidateName, jobTitle)
	case "under_review":
		message = fmt.Sprintf("Hi %s, your application for %s is under review. We'll update you within 5 business days.", candidateName, jobTitle)
	case "interview_scheduled":
		message = fmt.Sprintf("Hi %s, great news! We'd like to schedule an interview for %s. Check your email for details.", candidateName, jobTitle)
	default:
		// Don't send SMS for other statuses
		return nil
	}

	return SendSMS(phone, message)
}

