package utils

import (
	"fmt"
	"strings"
	"time"
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
    // Create a new email message
    m := gomail.NewMessage()
    m.SetHeader("From", m.FormatAddress(emailConfig.FromEmail, emailConfig.FromName))
    m.SetHeader("To", emailConfig.ToEmail)
    m.SetHeader("Subject", "New Order Notification")

    // Generate order items details
    orderItemsDetails := generateOrderItemsDetails(orderItems)

    // Format the order date
    orderDate := formatOrderDate(orderParams.OrderDate.Format(time.RFC3339))

    // Set email body
    plainTextBody := fmt.Sprintf("A new order has been placed. Order ID: %s, User Email: %s", orderID, orderParams.Email)
    htmlBody := fmt.Sprintf(`
        <!DOCTYPE html>
        <html>
        <head>
            <meta charset="UTF-8">
            <meta name="viewport" content="width=device-width, initial-scale=1.0">
            <title>New Order Notification</title>
        </head>
        <body style="margin:0; padding:0; font-family: Arial, sans-serif; background-color:#f4f4f4;">
            <table width="100%%" cellpadding="0" cellspacing="0">
                <tr>
                    <td align="center" style="padding: 20px 0;">
                        <table width="600" cellpadding="0" cellspacing="0" style="background-color:#ffffff; border-radius:10px; overflow:hidden; box-shadow:0 0 10px rgba(0,0,0,0.1);">
                            <!-- Header -->
                            <tr>
                                <td style="background: linear-gradient(90deg, #6a11cb 0%, #2575fc 100%); color:#ffffff; padding: 20px; text-align:center;">
                                    <h1 style="margin:0; font-size:24px;">New Order Received</h1>
                                    <p style="margin:5px 0 0 0; font-size:14px;">Order ID: %s</p>
                                </td>
                            </tr>
                            <!-- Content -->
                            <tr>
                                <td style="padding: 20px;">
                                    <!-- Customer Details -->
                                    <table width="100%%" cellpadding="0" cellspacing="0">
                                        <tr>
                                            <td width="50%%" style="vertical-align: top;">
                                                <h2 style="font-size:18px; color:#333333;">
                                                    <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" fill="#6a11cb" viewBox="0 0 16 16" style="vertical-align: middle; margin-right:5px;">
                                                        <path d="M8 8a3 3 0 1 0 0-6 3 3 0 0 0 0 6zm2-3a2 2 0 1 1-4 0 2 2 0 0 1 4 0z"/>
                                                        <path fill-rule="evenodd" d="M14 14s-1-4-6-4-6 4-6 4V12a6 6 0 1 1 12 0v2z"/>
                                                    </svg>
                                                    Customer Details
                                                </h2>
                                                <p><strong>Name:</strong> %s %s</p>
                                                <p><strong>Email:</strong> %s</p>
                                                <p><strong>Phone:</strong> %s</p>
                                            </td>
                                            <td width="50%%" style="vertical-align: top;">
                                                <h2 style="font-size:18px; color:#333333;">
                                                    <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" fill="#6a11cb" viewBox="0 0 16 16" style="vertical-align: middle; margin-right:5px;">
                                                        <path d="M3 0a3 3 0 0 0-3 3v10a3 3 0 0 0 3 3h10a3 3 0 0 0 3-3V3a3 3 0 0 0-3-3H3zm10 1a1 1 0 0 1 1 1v10a1 1 0 0 1-1 1H3a1 1 0 0 1-1-1V3a1 1 0 0 1 1-1h10z"/>
                                                        <path d="M4 6h8v2H4V6z"/>
                                                    </svg>
                                                    Shipping Address
                                                </h2>
                                                <p>%s</p>
                                                <p>%s, %s</p>
                                                <p>%s</p>
                                            </td>
                                        </tr>
                                    </table>
                                    <!-- Order Details -->
                                    <h2 style="font-size:18px; color:#333333; margin-top:20px;">
                                        <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" fill="#6a11cb" viewBox="0 0 16 16" style="vertical-align: middle; margin-right:5px;">
                                            <path d="M1 2a1 1 0 0 1 1-1h12a1 1 0 0 1 1 1v10a1 1 0 0 1-1 1H2a1 1 0 0 1-1-1V2zm1 1v9h12V3H2z"/>
                                            <path d="M4 6h8v2H4V6z"/>
                                        </svg>
                                        Order Details
                                    </h2>
                                    <table width="100%%" cellpadding="0" cellspacing="0" style="border-collapse: collapse;">
                                        <thead>
                                            <tr>
                                                <th style="background-color:#6a11cb; color:#ffffff; padding:10px; text-align:left;">Product</th>
                                                <th style="background-color:#6a11cb; color:#ffffff; padding:10px; text-align:center;">Quantity</th>
                                                <th style="background-color:#6a11cb; color:#ffffff; padding:10px; text-align:right;">Price (KES)</th>
                                                <th style="background-color:#6a11cb; color:#ffffff; padding:10px; text-align:right;">Total (KES)</th>
                                            </tr>
                                        </thead>
                                        <tbody>
                                            %s
                                            <tr>
                                                <td colspan="3" style="padding:10px; text-align:right; font-weight:bold;">Subtotal</td>
                                                <td style="padding:10px; text-align:right;">KES %s</td>
                                            </tr>
                                            <tr>
                                                <td colspan="3" style="padding:10px; text-align:right; font-weight:bold;">Discount</td>
                                                <td style="padding:10px; text-align:right;">KES %s</td>
                                            </tr>
                                            <tr>
                                                <td colspan="3" style="padding:10px; text-align:right; font-weight:bold;">Total</td>
                                                <td style="padding:10px; text-align:right; font-weight:bold;">KES %s</td>
                                            </tr>
                                        </tbody>
                                    </table>
                                </td>
                            </tr>
                            <!-- Footer -->
                            <tr>
                                <td style="background-color:#f4f4f4; padding:20px; text-align:center;">
                                    <p style="margin:0; font-size:12px; color:#777777;">Order placed on %s</p>
                                    <a href="%s" style="display:inline-block; margin-top:10px; padding:10px 20px; background-color:#6a11cb; color:#ffffff; text-decoration:none; border-radius:5px; font-weight:bold;">Process Order</a>
                                </td>
                            </tr>
                        </table>
                    </td>
                </tr>
            </table>
        </body>
        </html>`,
        orderID,
        orderParams.FirstName, orderParams.LastName,
        orderParams.Email,
        orderParams.Phone,
        formatShippingAddress(orderParams.City, orderParams.County, orderParams.Country),
        orderItemsDetails,
        formatPrice(orderParams.SubTotal),
        formatPrice(orderParams.DiscountAmount),
        formatPrice(orderParams.GrandTotal),
        orderDate,
        fmt.Sprintf("%s/orders/%s", emailConfig.BackendURL, orderID.String()), // Link to process the order
    )

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
        total := item.Price * float64(item.Quantity)
        sb.WriteString(fmt.Sprintf(`
            <tr>
                <td style="padding:10px; border-bottom:1px solid #dddddd;">%s</td>
                <td style="padding:10px; border-bottom:1px solid #dddddd; text-align:center;">%d</td>
                <td style="padding:10px; border-bottom:1px solid #dddddd; text-align:right;">KES %s</td>
                <td style="padding:10px; border-bottom:1px solid #dddddd; text-align:right;">KES %s</td>
            </tr>`,
            item.ProductName,
            item.Quantity,
            formatPrice(item.Price),
            formatPrice(total)))
    }
    return sb.String()
}

// formatPrice formats a float64 price with commas for better readability.
func formatPrice(price float64) string {
    priceStr := fmt.Sprintf("%.2f", price)
    n := len(priceStr)
    if n <= 6 {
        return priceStr
    }
    var sb strings.Builder
    mod := (n - 3) % 3
    if mod > 0 {
        sb.WriteString(priceStr[:mod])
        if n > mod {
            sb.WriteString(",")
        }
    }
    for i := mod; i < n-3; i += 3 {
        sb.WriteString(priceStr[i : i+3])
        sb.WriteString(",")
    }
    sb.WriteString(priceStr[n-3:])
    return sb.String()
}

// formatShippingAddress formats the shipping address without AddressLine1, AddressLine2, and ZipCode.
func formatShippingAddress(city, state, country string) string {
    return fmt.Sprintf("%s, %s, %s", city, state, country)
}

// formatOrderDate formats the order date string.
func formatOrderDate(orderDate string) string {
    // Assuming orderDate is in RFC3339 format, e.g., "2023-06-15T14:30:00Z"
    t, err := time.Parse(time.RFC3339, orderDate)
    if err != nil {
        return orderDate // Return as is if parsing fails
    }
    return t.Format("January 2, 2006 at 15:04 PM MST")
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

// SendVerificationEmail sends an email containing a verification link.
func SendVerificationEmail(emailConfig *appconfig.Config, userEmail, verificationToken string) error {
	// Construct the verification link
	verificationLink := fmt.Sprintf("%s/auth/verify-email?token=%s", emailConfig.BackendURL, verificationToken)

	// Create a new email message
	m := gomail.NewMessage()
	m.SetHeader("From", m.FormatAddress(emailConfig.FromEmail, emailConfig.FromName))
	m.SetHeader("To", userEmail)
	m.SetHeader("Subject", "Email Verification")

	// Set email body
	plainTextBody := fmt.Sprintf("Please verify your email address by clicking the following link: %s", verificationLink)
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
					<h1>Email Verification</h1>
				</div>
				<div class="content">
					<p>Please verify your email address by clicking the link below:</p>
					<p><a href="%s"><strong>Verify Email</strong></a></p>
				</div>
				<div class="footer">
					<p>&copy; 2024 Webline Technologies Ltd. All rights reserved.</p>
				</div>
			</div>
		</body>
		</html>`, verificationLink)

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

// SendPasswordResetEmail sends an email with a password reset link.
func SendPasswordResetEmail(emailConfig *appconfig.Config, userEmail, resetToken string) error {
	// Construct the password reset link
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", emailConfig.FrontendURL, resetToken)

	// Create a new email message
	m := gomail.NewMessage()
	m.SetHeader("From", m.FormatAddress(emailConfig.FromEmail, emailConfig.FromName))
	m.SetHeader("To", userEmail)
	m.SetHeader("Subject", "Password Reset Request")

	// Set email body
	plainTextBody := fmt.Sprintf("You requested a password reset. Please click the link to reset your password: %s", resetLink)
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
					<h1>Password Reset Request</h1>
				</div>
				<div class="content">
					<p>You requested a password reset. Click the link below to reset your password:</p>
					<p><a href="%s"><strong>Reset Password</strong></a></p>
				</div>
				<div class="footer">
					<p>&copy; 2024 Webline Technologies Ltd. All rights reserved.</p>
				</div>
			</div>
		</body>
		</html>`, resetLink)

	m.SetBody("text/plain", plainTextBody)
	m.AddAlternative("text/html", htmlBody)

	// Configure the SMTP dialer
	dialer := gomail.NewDialer(emailConfig.SMTPHost, emailConfig.SMTPPort, emailConfig.SMTPUsername, emailConfig.SMTPPassword)

	// Send the email
	if err := dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send password reset email: %w", err)
	}
	return nil
}

// SendAdminRequestEmail sends an email with a link to approve an admin role request.
func SendAdminRequestEmail(emailConfig *appconfig.Config, requesterEmail string, approvalToken string) error {
	// Construct the approval link
	approvalLink := fmt.Sprintf("%s/dashboard/approve?token=%s", emailConfig.FrontendURL, approvalToken)

	// Create a new email message
	m := gomail.NewMessage()
	m.SetHeader("From", m.FormatAddress(emailConfig.FromEmail, emailConfig.FromName))
	m.SetHeader("To", emailConfig.ToEmail)
	m.SetHeader("Subject", "Admin Role Request Approval")

	// Set email body
	plainTextBody := fmt.Sprintf("A new request for an admin role has been submitted by %s. Please review and approve the request using the following link: %s", requesterEmail, approvalLink)
	htmlBody := fmt.Sprintf(`
        <!DOCTYPE html>
        <html>
        <head>
            <style>
                body {
                    font-family: Arial, sans-serif;
                    color: #333333;
                    line-height: 1.6;
                }
                .container {
                    width: 80%%;
                    margin: 0 auto;
                    padding: 20px;
                    border: 1px solid #dcdcdc;
                    border-radius: 10px;
                    background-color: #f9f9f9;
                }
                .header {
                    text-align: center;
                    padding: 10px;
                    background-color: #007BFF;
                    color: #ffffff;
                    border-radius: 10px 10px 0 0;
                }
                .content {
                    padding: 20px;
                }
                .content p {
                    margin: 0 0 20px;
                }
                .footer {
                    text-align: center;
                    padding: 10px;
                    color: #999999;
                    font-size: 12px;
                }
                a {
                    color: #007BFF;
                    text-decoration: none;
                }
                a:hover {
                    text-decoration: underline;
                }
            </style>
        </head>
        <body>
            <div class="container">
                <div class="header">
                    <h1>Admin Role Request Approval</h1>
                </div>
                <div class="content">
                    <p>A new request for an admin role has been submitted by <strong>%s</strong>. Please review the details and approve the request if appropriate.</p>
                    <p><a href="%s"><strong>Click here to approve the request</strong></a></p>
                </div>
                <div class="footer">
                    <p>&copy; 2024 Webline Technologies Ltd. All rights reserved.</p>
                </div>
            </div>
        </body>
        </html>`, requesterEmail, approvalLink)

	m.SetBody("text/plain", plainTextBody)
	m.AddAlternative("text/html", htmlBody)

	// Configure the SMTP dialer
	dialer := gomail.NewDialer(emailConfig.SMTPHost, emailConfig.SMTPPort, emailConfig.SMTPUsername, emailConfig.SMTPPassword)

	// Send the email
	if err := dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send admin request email: %w", err)
	}
	return nil
}

// SendAdminRequestApprovedEmail sends an email notifying that the admin request has been approved.
func SendAdminRequestApprovedEmail(emailConfig *appconfig.Config, userEmail string) error {
	// Create a new email message
	m := gomail.NewMessage()
	m.SetHeader("From", m.FormatAddress(emailConfig.FromEmail, emailConfig.FromName))
	m.SetHeader("To", userEmail)
	m.SetHeader("Subject", "Admin Role Request Approved")

	// Set email body
	plainTextBody := "Your request for admin role has been approved. You now have access to admin functionalities on the platform."
	htmlBody := `
		<!DOCTYPE html>
		<html>
		<head>
			<style>
				body {
					font-family: Arial, sans-serif;
					color: #333333;
				}
				.container {
					width: 80%;
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
					<h1>Admin Role Request Approved</h1>
				</div>
				<div class="content">
					<p><strong>Congratulations!</strong></p>
					<p>Your request for admin role has been <strong>approved</strong>. You now have access to admin functionalities on the platform.</p>
					<p>Please log in to your account to start managing the platform's features and settings.</p>
				</div>
				<div class="footer">
					<p>&copy; 2024 Webline Technologies Ltd. All rights reserved.</p>
				</div>
			</div>
		</body>
		</html>`

	m.SetBody("text/plain", plainTextBody)
	m.AddAlternative("text/html", htmlBody)

	// Configure the SMTP dialer
	dialer := gomail.NewDialer(emailConfig.SMTPHost, emailConfig.SMTPPort, emailConfig.SMTPUsername, emailConfig.SMTPPassword)

	// Send the email
	if err := dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send admin request approval email: %w", err)
	}
	return nil
}
