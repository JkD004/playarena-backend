package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type ResendRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Html    string   `json:"html"`
}

// SendEmail sends an email using the Resend API (Bypasses SMTP ports)
func SendEmail(to string, subject string, body string) error {
	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("RESEND_API_KEY is not set")
	}

	// For testing without a domain, you MUST use this sender:
	// Once you verify your domain on Resend, you can change this to "support@yourdomain.com"
	fromEmail := "onboarding@resend.dev" 

	reqBody := ResendRequest{
		From:    fromEmail,
		To:      []string{to}, // In testing, 'to' must be YOUR email used to signup
		Subject: subject,
		Html:    body,
	}

	jsonBody, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to Resend: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("resend API failed with status: %d", resp.StatusCode)
	}

	return nil
}