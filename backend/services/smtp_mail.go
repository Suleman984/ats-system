package services

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"os"
	"strings"
	"time"
)

// SMTP delivery (EMAIL_PROVIDER=smtp), e.g. Gmail with an app password:
// SMTP_HOST=smtp.gmail.com SMTP_PORT=587 SMTP_USERNAME=you@gmail.com SMTP_PASSWORD=<app password>.
// SMTP_FROM defaults to SMTP_USERNAME; Gmail rewrites any other address to the account's own.

const smtpTimeout = 20 * time.Second

func sendViaSMTP(to, subject, htmlBody string) error {
	host, user, pass := os.Getenv("SMTP_HOST"), os.Getenv("SMTP_USERNAME"), os.Getenv("SMTP_PASSWORD")
	if host == "" || user == "" || pass == "" {
		return errors.New("SMTP_HOST, SMTP_USERNAME and SMTP_PASSWORD must be set")
	}
	port := os.Getenv("SMTP_PORT")
	if port == "" {
		port = "587"
	}
	from := os.Getenv("SMTP_FROM")
	if from == "" {
		from = user
	}
	fromAddr, err := mail.ParseAddress(from)
	if err != nil {
		return fmt.Errorf("SMTP_FROM: %w", err)
	}
	toAddr, err := mail.ParseAddress(to)
	if err != nil {
		return fmt.Errorf("recipient: %w", err)
	}

	client, err := dialSMTP(host, port)
	if err != nil {
		return err
	}
	defer client.Close()

	// Port 587 must upgrade to TLS before credentials are sent; 465 is TLS from the start.
	if port != "465" {
		if ok, _ := client.Extension("STARTTLS"); !ok {
			return errors.New("SMTP server does not offer STARTTLS; refusing to send credentials in clear text")
		}
		if err := client.StartTLS(&tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}); err != nil {
			return fmt.Errorf("starttls: %w", err)
		}
	}
	if err := client.Auth(smtp.PlainAuth("", user, pass, host)); err != nil {
		return fmt.Errorf("smtp auth: %w", err)
	}
	if err := client.Mail(fromAddr.Address); err != nil {
		return err
	}
	if err := client.Rcpt(toAddr.Address); err != nil {
		return err
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(buildMIMEMessage(fromAddr, toAddr, subject, htmlBody)); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return client.Quit()
}

func dialSMTP(host, port string) (*smtp.Client, error) {
	addr := net.JoinHostPort(host, port)
	dialer := &net.Dialer{Timeout: smtpTimeout}
	var conn net.Conn
	var err error
	if port == "465" {
		conn, err = tls.DialWithDialer(dialer, "tcp", addr, &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12})
	} else {
		conn, err = dialer.Dial("tcp", addr)
	}
	if err != nil {
		return nil, fmt.Errorf("smtp connect: %w", err)
	}
	_ = conn.SetDeadline(time.Now().Add(smtpTimeout))
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		conn.Close()
		return nil, err
	}
	return client, nil
}

// buildMIMEMessage writes headers safely: CR/LF can't sneak in (subjects include recruiter-typed job
// titles), and non-ASCII subjects are RFC 2047 encoded.
func buildMIMEMessage(from, to *mail.Address, subject, htmlBody string) []byte {
	subject = strings.NewReplacer("\r", " ", "\n", " ").Replace(subject)
	id := make([]byte, 12)
	_, _ = rand.Read(id)
	domain := "localhost"
	if at := strings.LastIndex(from.Address, "@"); at >= 0 {
		domain = from.Address[at+1:]
	}

	var b strings.Builder
	fmt.Fprintf(&b, "From: %s\r\n", from.String())
	fmt.Fprintf(&b, "To: %s\r\n", to.String())
	fmt.Fprintf(&b, "Subject: %s\r\n", mime.QEncoding.Encode("utf-8", subject))
	fmt.Fprintf(&b, "Date: %s\r\n", time.Now().Format(time.RFC1123Z))
	fmt.Fprintf(&b, "Message-ID: <%s@%s>\r\n", hex.EncodeToString(id), domain)
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	b.WriteString("Content-Transfer-Encoding: 8bit\r\n\r\n")
	b.WriteString(strings.ReplaceAll(strings.ReplaceAll(htmlBody, "\r\n", "\n"), "\n", "\r\n"))
	return []byte(b.String())
}
