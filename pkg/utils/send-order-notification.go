package utils

import (
	"fmt"
	"strings"
	"weblineBackend/internal/appconfig"
	"weblineBackend/internal/model"

	"github.com/google/uuid"
	"gopkg.in/gomail.v2"
)

type OrderItem struct {
	ProductName string
	Quantity    int32
	Price       float64
}

// SendOrderNotification sends an email notification for a new order.
func SendOrderNotification(emailConfig *appconfig.Config, orderID uuid.UUID, orderParams *model.CreateOrderParams, orderItems []OrderItem) error {
	if orderParams.PaymentOption == "delivery" {
		orderParams.PaymentOption = "Pay On Delivery"
	} else {
		orderParams.PaymentOption = "Pay Now - Mpesa"
	}
	// Create a new email message
	m := gomail.NewMessage()
	m.SetHeader("From", m.FormatAddress(emailConfig.FromEmail, emailConfig.FromName))
	m.SetHeader("To", emailConfig.ToEmail)
	m.SetHeader("Subject", "New Order Notification")

	// Generate order items details
	orderItemsDetails := generateOrderItemsDetails(orderItems)

	// Set email body
	plainTextBody := fmt.Sprintf("A new order has been placed. Order ID: %s, User Email: %s", orderID, orderParams.Email)
	htmlBody := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<style>
				body {
					font-family: Arial, sans-serif;
					color: #333333;
				}
				.container {
					width: 80%%;
					margin: auto;
					padding: 20px;
					border: 1px solid #dcdcdc;
					border-radius: 10px;
					background-color: #f9f9f9;
				}
				.header {
					text-align: center;
					padding: 10px 0;
					background-color: #007BFF;
					color: #ffffff;
					border-radius: 10px 10px 0 0;
				}
				.content {
					padding: 20px;
				}
				.footer {
					text-align: center;
					padding: 10px 0;
					color: #999999;
					font-size: 12px;
				}
				strong {
					color: #007BFF;
				}
				.details {
					margin: 20px 0;
				}
				.details th, .details td {
					padding: 8px 12px;
					text-align: left;
				}
				.details th {
					background-color: #007BFF;
					color: #ffffff;
				}
				.details tbody tr:nth-child(odd) {
					background-color: #f2f2f2;
				}
			</style>
		</head>
		<body>
			<div class="container">
				<div class="header">
					<h1>New Order Notification</h1>
				</div>
				<div class="content">
					<p><strong>A new order has been placed.</strong></p>
					<p>Order ID: <strong>%s</strong></p>
					<p>Customer: <strong>%s %s</strong></p>
					<p>Email: <strong>%s</strong></p>
					<p>Phone: <strong>%s</strong></p>
					<p>Address: <strong>%s, %s, %s, %s</strong></p>
					<p>Shipping Option: <strong>%s</strong></p>
					<p>Payment Method: <strong>%s</strong></p>
					<p>Total: <strong>KES %s</strong></p>
					<div class="details">
						<h3>Order Items</h3>
						<table>
							<thead>
								<tr>
									<th>Product Name</th>
									<th>Quantity</th>
									<th>Price (KES)</th>
								</tr>
							</thead>
							<tbody>
								%s
							</tbody>
						</table>
					</div>
				</div>
				<div class="footer">
					<p>&copy; 2024 Webline Technologies Ltd. All rights reserved.</p>
				</div>
			</div>
		</body>
		</html>`, orderID, orderParams.FirstName, orderParams.LastName, orderParams.Email, orderParams.Phone, orderParams.StreetAddress, orderParams.City, orderParams.State, orderParams.Country, orderParams.ShippingOption, orderParams.PaymentOption, formatPrice(orderParams.Total), orderItemsDetails)

	m.SetBody("text/plain", plainTextBody)
	m.AddAlternative("text/html", htmlBody)

	// Configure the SMTP dialer
	dialer := gomail.NewDialer(emailConfig.SMTPHost, emailConfig.SMTPPort, emailConfig.SMTPUsername, emailConfig.SMTPPassword)

	// Send the email
	if err := dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}

// generateOrderItemsDetails creates the HTML rows for the order items table.
func generateOrderItemsDetails(orderItems []OrderItem) string {
	var sb strings.Builder
	for _, item := range orderItems {
		sb.WriteString(fmt.Sprintf(`
			<tr>
				<td>%s</td>
				<td>%d</td>
				<td>KES %s</td>
			</tr>`, item.ProductName, item.Quantity, formatPrice(item.Price)))
	}
	return sb.String()
}

// formatPrice formats a float64 price with commas for better readability.
func formatPrice(price float64) string {
	priceStr := fmt.Sprintf("%.0f", price)
	n := len(priceStr)
	if n <= 3 {
		return priceStr
	}
	var sb strings.Builder
	mod := n % 3
	if mod > 0 {
		sb.WriteString(priceStr[:mod])
		if n > mod {
			sb.WriteString(",")
		}
	}
	for i := mod; i < n; i += 3 {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(priceStr[i : i+3])
	}
	return sb.String()
}

// SendInquiryEmail sends an email with the user's product inquiry.
func SendInquiryEmail(emailConfig *appconfig.Config, productName, userEmail, userMessage string) error {
	// Create a new email message
	m := gomail.NewMessage()
	m.SetHeader("From", m.FormatAddress(emailConfig.FromEmail, emailConfig.FromName))
	m.SetHeader("To", emailConfig.ToEmail)
	m.SetHeader("Subject", fmt.Sprintf("Product Inquiry: %s", productName))

	// Set email body
	plainTextBody := fmt.Sprintf("You have received an inquiry for the product '%s'.\n\nFrom: %s\n\nMessage:\n%s", productName, userEmail, userMessage)
	htmlBody := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<style>
				body {
					font-family: Arial, sans-serif;
					color: #333333;
				}
				.container {
					width: 80%%;
					margin: auto;
					padding: 20px;
					border: 1px solid #dcdcdc;
					border-radius: 10px;
					background-color: #f9f9f9;
				}
				.header {
					text-align: center;
					padding: 10px 0;
					background-color: #007BFF;
					color: #ffffff;
					border-radius: 10px 10px 0 0;
				}
				.content {
					padding: 20px;
				}
				.footer {
					text-align: center;
					padding: 10px 0;
					color: #999999;
					font-size: 12px;
				}
				strong {
					color: #007BFF;
				}
			</style>
		</head>
		<body>
			<div class="container">
				<div class="header">
					<h1>Product Inquiry</h1>
				</div>
				<div class="content">
					<p>You have received an inquiry for the product '<strong>%s</strong>'.</p>
					<p><strong>From:</strong> %s</p>
					<p><strong>Message:</strong></p>
					<p>%s</p>
				</div>
				<div class="footer">
					<p>&copy; 2024 Webline Technologies Ltd. All rights reserved.</p>
				</div>
			</div>
		</body>
		</html>`, productName, userEmail, userMessage)

	m.SetBody("text/plain", plainTextBody)
	m.AddAlternative("text/html", htmlBody)

	// Configure the SMTP dialer
	dialer := gomail.NewDialer(emailConfig.SMTPHost, emailConfig.SMTPPort, emailConfig.SMTPUsername, emailConfig.SMTPPassword)

	// Send the email
	if err := dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}

// SendOrderCancellationNotification sends an email notification for a cancelled order.
func SendOrderCancellationNotification(emailConfig *appconfig.Config, orderNumber string, reason string) error {
	// Create a new email message
	m := gomail.NewMessage()
	m.SetHeader("From", m.FormatAddress(emailConfig.FromEmail, emailConfig.FromName))
	m.SetHeader("To", emailConfig.ToEmail)
	m.SetHeader("Subject", "Order Cancellation Notification")

	// Set email body
	plainTextBody := fmt.Sprintf("Order %s has been cancelled.\n\nReason: %s", orderNumber, reason)
	htmlBody := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<style>
				body {
					font-family: Arial, sans-serif;
					color: #333333;
				}
				.container {
					width: 80%%;
					margin: auto;
					padding: 20px;
					border: 1px solid #dcdcdc;
					border-radius: 10px;
					background-color: #f9f9f9;
				}
				.header {
					text-align: center;
					padding: 10px 0;
					background-color: #FF0000;
					color: #ffffff;
					border-radius: 10px 10px 0 0;
				}
				.content {
					padding: 20px;
				}
				.footer {
					text-align: center;
					padding: 10px 0;
					color: #999999;
					font-size: 12px;
				}
				strong {
					color: #FF0000;
				}
			</style>
		</head>
		<body>
			<div class="container">
				<div class="header">
					<h1>Order Cancellation Notification</h1>
				</div>
				<div class="content">
					<p>Order <strong>%s</strong> has been cancelled.</p>
					<p>Reason: <strong>%s</strong></p>
				</div>
				<div class="footer">
					<p>&copy; 2024 Webline Technologies Ltd. All rights reserved.</p>
				</div>
			</div>
		</body>
		</html>`, orderNumber, reason)

	m.SetBody("text/plain", plainTextBody)
	m.AddAlternative("text/html", htmlBody)

	// Configure the SMTP dialer
	dialer := gomail.NewDialer(emailConfig.SMTPHost, emailConfig.SMTPPort, emailConfig.SMTPUsername, emailConfig.SMTPPassword)

	// Send the email
	if err := dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}

// SendOrderPaymentMethodChangeNotification sends an email notification for an order payment method change.
func SendOrderPaymentMethodChangeNotification(emailConfig *appconfig.Config, orderNumber string, method string) error {
	// Create a new email message
	m := gomail.NewMessage()
	m.SetHeader("From", m.FormatAddress(emailConfig.FromEmail, emailConfig.FromName))
	m.SetHeader("To", emailConfig.ToEmail)
	m.SetHeader("Subject", "Order Payment Method Change Notification")

	// Set email body
	plainTextBody := fmt.Sprintf("The payment method for order %s has been changed to %s.", orderNumber, method)
	htmlBody := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<style>
				body {
					font-family: Arial, sans-serif;
					color: #333333;
				}
				.container {
					width: 80%%;
					margin: auto;
					padding: 20px;
					border: 1px solid #dcdcdc;
					border-radius: 10px;
					background-color: #f9f9f9;
				}
				.header {
					text-align: center;
					padding: 10px 0;
					background-color: #007BFF;
					color: #ffffff;
					border-radius: 10px 10px 0 0;
				}
				.content {
					padding: 20px;
				}
				.footer {
					text-align: center;
					padding: 10px 0;
					color: #999999;
					font-size: 12px;
				}
				strong {
					color: #007BFF;
				}
			</style>
		</head>
		<body>
			<div class="container">
				<div class="header">
					<h1>Order Payment Method Change Notification</h1>
				</div>
				<div class="content">
					<p>The payment method for order <strong>%s</strong> has been changed to <strong>%s</strong>.</p>
				</div>
				<div class="footer">
					<p>&copy; 2024 Webline Technologies Ltd. All rights reserved.</p>
				</div>
			</div>
		</body>
		</html>`, orderNumber, method)

	m.SetBody("text/plain", plainTextBody)
	m.AddAlternative("text/html", htmlBody)

	// Configure the SMTP dialer
	dialer := gomail.NewDialer(emailConfig.SMTPHost, emailConfig.SMTPPort, emailConfig.SMTPUsername, emailConfig.SMTPPassword)

	// Send the email
	if err := dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}
