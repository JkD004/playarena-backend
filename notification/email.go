package notification

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"os"
)

// SendEmail sends a simple email using Gmail SMTP (Supports 587 and 465)
func SendEmail(to string, subject string, body string) error {
	from := os.Getenv("SMTP_EMAIL")
	password := os.Getenv("SMTP_PASSWORD")
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")

	if from == "" || password == "" {
		return fmt.Errorf("SMTP credentials not set in environment")
	}

	// Address
	addr := fmt.Sprintf("%s:%s", host, port)

	// Headers
	headers := "MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\"\r\n" +
		"From: " + from + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n\r\n" +
		body + "\r\n"

	// AUTHENTICATION
	auth := smtp.PlainAuth("", from, password, host)

	// LOGIC FOR PORT 465 (SSL)
	if port == "465" {
		// Create SSL connection
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         host,
		}
		
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("failed to dial tls: %v", err)
		}
		
		c, err := smtp.NewClient(conn, host)
		if err != nil {
			return fmt.Errorf("failed to create smtp client: %v", err)
		}
		
		// Auth
		if err = c.Auth(auth); err != nil {
			return fmt.Errorf("failed to auth: %v", err)
		}
		
		// To
		if err = c.Mail(from); err != nil {
			return fmt.Errorf("failed to set sender: %v", err)
		}
		if err = c.Rcpt(to); err != nil {
			return fmt.Errorf("failed to set recipient: %v", err)
		}
		
		// Data
		w, err := c.Data()
		if err != nil {
			return fmt.Errorf("failed to get data writer: %v", err)
		}
		
		_, err = w.Write([]byte(headers))
		if err != nil {
			return fmt.Errorf("failed to write body: %v", err)
		}
		
		err = w.Close()
		if err != nil {
			return fmt.Errorf("failed to close writer: %v", err)
		}
		
		return c.Quit()
	}

	// LOGIC FOR PORT 587 (Standard)
	return smtp.SendMail(addr, auth, from, []string{to}, []byte(headers))
}