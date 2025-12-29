// notification/email.go
package notification

import (
	"fmt"
	"net/smtp"
	"os"
)

// SendEmail sends a simple email using Gmail SMTP
func SendEmail(to string, subject string, body string) error {
	// 1. Read credentials from .env
	from := os.Getenv("SMTP_EMAIL")
	password := os.Getenv("SMTP_PASSWORD") // This must be the App Password
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")

	// Basic validation to prevent crashes if env vars are missing
	if from == "" || password == "" || host == "" {
		return fmt.Errorf("SMTP credentials not set in environment")
	}

	// 2. Setup Authentication
	auth := smtp.PlainAuth("", from, password, host)

	// 3. Compose Message (MIME headers allow HTML content)
	msg := []byte("To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\"\r\n" +
		"\r\n" +
		body + "\r\n")

	// 4. Send Email
	addr := fmt.Sprintf("%s:%s", host, port)
	err := smtp.SendMail(addr, auth, from, []string{to}, msg)
	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}