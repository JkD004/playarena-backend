// pkg/email/email.go
package email

import (
	"fmt"
	"net/smtp"
	"os"
)

// SendEmail sends a simple email notification
func SendEmail(to string, subject string, body string) error {
	// 1. Get Config from Env
	from := os.Getenv("SMTP_EMAIL")
	password := os.Getenv("SMTP_PASSWORD")
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")

	// 2. Setup Auth
	auth := smtp.PlainAuth("", from, password, host)

	// 3. Compose Message (Simple Text + HTML compatible headers)
	// We use standard MIME headers so it looks professional
	msg := []byte("To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\"\r\n" +
		"\r\n" +
		body + "\r\n")

	// 4. Send
	addr := fmt.Sprintf("%s:%s", host, port)
	err := smtp.SendMail(addr, auth, from, []string{to}, msg)
	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}