package utils

import (
	"fmt"
	"os"

	gomail "gopkg.in/gomail.v2"
)

func SendEmail(email string, subject string, validationCode string) error {

	msg := gomail.NewMessage()
	msg.SetHeader("From", os.Getenv("EMAIL"))
	msg.SetHeader("To", email)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/html", fmt.Sprintf("This is your code <b>%s</b>, <br> Go quickly this code expires soon, loomies are waiting for you!", validationCode))

	n := gomail.NewDialer("smtp.gmail.com", 587, os.Getenv("EMAIL"), os.Getenv("EMAIL_PASSWORD"))

	// Send the email
	if err := n.DialAndSend(msg); err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
