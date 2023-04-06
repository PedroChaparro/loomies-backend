package utils

import (
	"fmt"
	"os"

	"github.com/PedroChaparro/loomies-backend/configuration"
	gomail "gopkg.in/gomail.v2"
)

// SendEmail sends an email to the specified address usint the specified subject and body
func SendEmail(email string, subject string, validationCode string) error {
	// Prevent sending emaails on test environment
	if configuration.GetRunningEnvironment() == "TESTING" {
		return nil
	}

	sender := os.Getenv("EMAIL_MAIL")
	msg := gomail.NewMessage()
	msg.SetHeader("From", sender)
	msg.SetHeader("To", email)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/html", fmt.Sprintf("This is your code: <b>%s</b>, <br> Hurry up! This code expires soon, Loomies are waiting for you!", validationCode))

	n := gomail.NewDialer("smtp.gmail.com", 587, sender, os.Getenv("EMAIL_PASSWORD"))

	// Send the email
	if err := n.DialAndSend(msg); err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}
