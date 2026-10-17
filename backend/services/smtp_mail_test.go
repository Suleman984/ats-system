package services

import (
	"net/mail"
	"strings"
	"testing"
)

func TestBuildMIMEMessageBlocksHeaderInjection(t *testing.T) {
	from := &mail.Address{Name: "ATS", Address: "ats@example.com"}
	to := &mail.Address{Address: "c@example.com"}
	msg := string(buildMIMEMessage(from, to, "Hello\r\nBcc: victim@example.com", "<p>hi</p>"))

	headers := msg[:strings.Index(msg, "\r\n\r\n")]
	if strings.Contains(headers, "\r\nBcc:") {
		t.Fatalf("subject injected a header:\n%s", headers)
	}
	if !strings.Contains(headers, "Content-Type: text/html; charset=UTF-8") {
		t.Fatal("missing HTML content type")
	}
}

func TestBuildMIMEMessageEncodesUnicodeSubject(t *testing.T) {
	from := &mail.Address{Address: "ats@example.com"}
	msg := string(buildMIMEMessage(from, from, "Offer – Lahore", "x"))
	if !strings.Contains(msg, "Subject: =?utf-8?q?") {
		t.Fatalf("non-ASCII subject not encoded:\n%s", msg)
	}
}
