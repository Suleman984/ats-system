package services

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var smsClient = &http.Client{Timeout: 15 * time.Second}

// SendSMS is a no-op unless SMS_PROVIDER=twilio and the Twilio credentials are set.
func SendSMS(to, message string) error {
	if !strings.EqualFold(os.Getenv("SMS_PROVIDER"), "twilio") {
		return nil
	}
	sid := os.Getenv("TWILIO_ACCOUNT_SID")
	token := os.Getenv("TWILIO_AUTH_TOKEN")
	from := os.Getenv("TWILIO_FROM_NUMBER")
	if sid == "" || token == "" || from == "" {
		return nil
	}

	to = normalizePhone(to)
	if to == "" {
		return fmt.Errorf("invalid phone number")
	}

	form := url.Values{"From": {from}, "To": {to}, "Body": {message}}
	endpoint := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", url.PathEscape(sid))
	req, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.SetBasicAuth(sid, token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := smsClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		detail, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("twilio returned %d: %s", resp.StatusCode, detail)
	}
	return nil
}

// normalizePhone keeps digits and a leading +; Twilio expects E.164.
func normalizePhone(phone string) string {
	var b strings.Builder
	for i, r := range strings.TrimSpace(phone) {
		if (r >= '0' && r <= '9') || (r == '+' && i == 0) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func SendStatusUpdateSMS(phone, candidateName, jobTitle, status string) error {
	if phone == "" {
		return nil
	}
	var message string
	switch status {
	case "cv_viewed":
		message = fmt.Sprintf("Hi %s, your application for %s has been reviewed. We'll be in touch soon!", candidateName, jobTitle)
	case "shortlisted":
		message = fmt.Sprintf("Congratulations %s! You've been shortlisted for %s. We'll contact you soon with next steps.", candidateName, jobTitle)
	case "rejected":
		message = fmt.Sprintf("Hi %s, thank you for applying to %s. We've decided to move forward with other candidates at this time.", candidateName, jobTitle)
	default:
		return nil
	}
	return SendSMS(phone, message)
}
