package email

import (
	"bytes"
	"context"
	"fmt"
	"log"

	"survivor-app/view"

	"github.com/resend/resend-go/v3"
)

type EmailSender interface {
	SendPasswordReset(toEmail, token string) error
}

type LogEmailSender struct{}

func NewLogEmailSender() *LogEmailSender {
	return &LogEmailSender{}
}

func (s *LogEmailSender) SendPasswordReset(toEmail, token string) error {
	var emailBody bytes.Buffer
	// A context is needed for rendering. A background context is fine here.
	err := view.PasswordResetEmail("localhost:8080", token).Render(context.Background(), &emailBody)
	if err != nil {
		log.Printf("failed to render email template for logging: %v", err)
		// Fallback to old logging if template fails
		resetURL := fmt.Sprintf("http://localhost:8080/reset-password?token=%s", token)
		log.Printf("---- MOCK PASSWORD RESET EMAIL (TEMPLATE RENDER FAILED) ----\nTo: %s\nSubject: Reset Your Password\nBody: Click the link to reset your password: %s\n---------------------------------", toEmail, resetURL)
		return err
	}

	log.Printf("---- MOCK PASSWORD RESET EMAIL ----\nTo: %s\nSubject: Reset Your Survivor App Password\nBody:\n%s\n---------------------------------", toEmail, emailBody.String())
	return nil
}

type ResendEmailSender struct {
	APIKey string
	Domain string
}

func NewResendEmailSender(domain string, apiKey string) *ResendEmailSender {
	return &ResendEmailSender{
		APIKey: apiKey,
		Domain: domain,
	}
}

func (s *ResendEmailSender) SendPasswordReset(toEmail, token string) error {
	client := resend.NewClient(s.APIKey)

	var emailBody bytes.Buffer
	err := view.PasswordResetEmail(s.Domain, token).Render(context.Background(), &emailBody)
	if err != nil {
		log.Printf("failed to render email template: %v", err)
		return err
	}

	params := &resend.SendEmailRequest{
		From:    "passwordreset@blueelm-survivor.com",
		To:      []string{toEmail},
		Subject: "Reset Your Survivor App Password",
		Html:    emailBody.String(),
	}

	sent, err := client.Emails.Send(params)
	if err != nil {
		log.Printf("failed to send email: %v", err)
		return err
	}

	log.Printf("email sent: %s", sent.Id)
	return nil
}
