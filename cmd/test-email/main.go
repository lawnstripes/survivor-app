package main

import (
	"fmt"
	"log"
	"os"

	"survivor-app/internal/email"

	"github.com/joho/godotenv"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run ./cmd/test-email user@email.com")
		os.Exit(1)
	}

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	if os.Getenv("EMAIL_API_KEY") == "" {
		log.Fatal("Resend API key not loaded")
	}

	sender := email.NewResendEmailSender(os.Getenv("EMAIL_DOMAIN"), os.Getenv("EMAIL_API_KEY"))
	if sender != nil {
		sender.SendPasswordReset(os.Args[1], "THIS IS A FAKE TOKEN THAT WOULD RESET YOUR PASSWORD")
	}
}
