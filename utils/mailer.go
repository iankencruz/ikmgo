package utils

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"strconv"

	"gopkg.in/gomail.v2"
)

// SendEmail uses GoMail to send an HTML email using a provided template.
func SendEmail(from, to, subject, tmplPath string, data interface{}) error {
	// 1) Parse the HTML template file
	t, err := template.ParseFiles(tmplPath)
	if err != nil {
		return fmt.Errorf("failed to parse email template: %w", err)
	}

	// 2) Execute the template with 'data'
	var body bytes.Buffer
	if err := t.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to execute email template: %w", err)
	}

	// 3) Create a new GoMail message
	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body.String())

	// 4) Parse the SMTP port (which is a string in .env) into an integer
	portStr := os.Getenv("SMTP_PORT")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("invalid SMTP_PORT %q: %w", portStr, err)
	}

	// 5) Build a new GoMail dialer with your SMTP settings
	dialer := gomail.NewDialer(
		os.Getenv("SMTP_HOST"), // e.g. "smtp.gmail.com"
		port,                   // e.g. 587
		os.Getenv("SMTP_USER"), // your email address or username
		os.Getenv("SMTP_PASS"), // your password or app-specific password
	)

	// Optional: If you need SSL/TLS, you can configure:
	// dialer.SSL = true
	// or dialer.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	// 6) Dial and send the message
	if err := dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
