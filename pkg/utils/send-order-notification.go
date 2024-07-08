package utils

import (
	"fmt"
	"weblineBackend/internal/appconfig"

	"github.com/google/uuid"
	"gopkg.in/gomail.v2"
)

func SendOrderNotification(emailConfig *appconfig.Config, orderID uuid.UUID, userEmail string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", m.FormatAddress(emailConfig.FromEmail, emailConfig.FromName))
	m.SetHeader("To", emailConfig.ToEmail)
	m.SetHeader("Subject", "New Order Notification")
	m.SetBody("text/plain", fmt.Sprintf("A new order has been placed. Order ID: %s, User Email: %s", orderID, userEmail))
	m.SetBody("text/html", fmt.Sprintf("<strong>A new order has been placed.</strong><br>Order ID: %s<br>User Email: %s", orderID, userEmail))

	d := gomail.NewDialer(emailConfig.SMTPHost, emailConfig.SMTPPort, emailConfig.SMTPUsername, emailConfig.SMTPPassword)

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}
