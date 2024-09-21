package utils

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"strings"
	"time"
	"weblineBackend/internal/appconfig"
	"weblineBackend/internal/model"

	"github.com/google/uuid"
	"gopkg.in/gomail.v2"
)

// OrderItem represents a single item in the order.
type OrderItem struct {
    ProductName string
    Quantity    int
    Price       float64
}

// SendOrderNotification sends an email notification for a new order.
func SendOrderNotification(emailConfig *appconfig.Config, orderID uuid.UUID, orderParams *model.CreateOrderParams, orderItems []OrderItem) error {
    if emailConfig == nil {
        return fmt.Errorf("email configuration is required")
    }

    // Generate order items details
    orderItemsDetails := generateOrderItemsDetails(orderItems)

    // Format the order date
    orderDate, err := formatOrderDate(orderParams.OrderDate)
    if err != nil {
        return fmt.Errorf("failed to format order date: %w", err)
    }

    // Prepare email data
    data := EmailData{
        OrderNumber:      orderParams.OrderNumber,
        OrderID:          orderID.String(),
        FirstName:        orderParams.FirstName,
        LastName:         orderParams.LastName,
        Email:            orderParams.Email,
        Phone:            orderParams.Phone,
        ShippingAddress:  formatShippingAddress(orderParams.City, orderParams.County, orderParams.Country),
        OrderItems:       orderItemsDetails,
        SubTotal:         formatPrice(orderParams.SubTotal),
        DiscountAmount:   formatPrice(orderParams.DiscountAmount),
        GrandTotal:       formatPrice(orderParams.GrandTotal),
        OrderDate:        orderDate,
        ProcessOrderLink: fmt.Sprintf("%s/orders/%s", emailConfig.BackendURL, orderID.String()),
		VAT:              formatPrice(orderParams.VatAmount),
    }

    // Define the email template
    tpl := template.Must(template.New("email").Parse(emailTemplate))

    var tplBuffer bytes.Buffer
    if err := tpl.Execute(&tplBuffer, data); err != nil {
        return fmt.Errorf("failed to execute email template: %w", err)
    }

    // Prepare plain text body
    plainTextBody := fmt.Sprintf(
        "A new order has been placed.\nOrder Number: %s\nUser Email: %s",
        data.OrderNumber,
        data.Email,
    )

    // Create a new email message
    m := gomail.NewMessage()
    m.SetHeader("From", m.FormatAddress(emailConfig.FromEmail, emailConfig.FromName))
    m.SetHeader("To", emailConfig.ToEmail)
    m.SetHeader("Subject", fmt.Sprintf("New Order Notification - Order Number: %s", data.OrderNumber))
    m.SetBody("text/plain", plainTextBody)
    m.AddAlternative("text/html", tplBuffer.String())

    // Optional: Log the email contents for debugging
    log.Println("Plain Text Body:", plainTextBody)
    log.Println("HTML Body:", tplBuffer.String())

    // Configure the SMTP dialer
    dialer := gomail.NewDialer(emailConfig.SMTPHost, emailConfig.SMTPPort, emailConfig.SMTPUsername, emailConfig.SMTPPassword)

    // Send the email
    if err := dialer.DialAndSend(m); err != nil {
        return fmt.Errorf("failed to send email: %w", err)
    }

    return nil
}

// EmailData represents the data structure used in the email template.
type EmailData struct {
    OrderNumber      string
    OrderID          string
    FirstName        string
    LastName         string
    Email            string
    Phone            string
    ShippingAddress  string
    OrderItems       template.HTML
    SubTotal         string
    DiscountAmount   string
    GrandTotal       string
    OrderDate        string
    ProcessOrderLink string
	VAT              string
}

// emailTemplate is the HTML template for the order notification email.
const emailTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>New Order Notification</title>
</head>
<body style="margin:0; padding:0; font-family: Arial, sans-serif; background-color:#f4f4f4;">
    <table width="100%" cellpadding="0" cellspacing="0">
        <tr>
            <td align="center" style="padding: 20px 0;">
                <table width="600" cellpadding="0" cellspacing="0" style="background-color:#ffffff; border-radius:10px; overflow:hidden; box-shadow:0 0 10px rgba(0,0,0,0.1);">
                    <tr>
                        <td style="background: linear-gradient(90deg, #6a11cb 0%, #2575fc 100%); color:#ffffff; padding: 20px; text-align:center;">
                            <h1 style="margin:0; font-size:24px;">New Order Received</h1>
                            <p style="margin:5px 0 0 0; font-size:14px;">Order Number: {{.OrderNumber}}</p>
                        </td>
                    </tr>
                    <tr>
                        <td style="padding: 20px;">
                            <h2 style="font-size:18px; color:#333333;">Customer Details</h2>
                            <p><strong>Name:</strong> {{.FirstName}} {{.LastName}}</p>
                            <p><strong>Email:</strong> {{.Email}}</p>
                            <p><strong>Phone:</strong> {{.Phone}}</p>
                            <h2 style="font-size:18px; color:#333333;">Shipping Address</h2>
                            <p>{{.ShippingAddress}}</p>
                            <h2 style="font-size:18px; color:#333333; margin-top:20px;">Order Details</h2>
                            <table width="100%" cellpadding="0" cellspacing="0" style="border-collapse: collapse;">
                                <thead>
                                    <tr>
                                        <th style="background-color:#6a11cb; color:#ffffff; padding:10px; text-align:left;">Product</th>
                                        <th style="background-color:#6a11cb; color:#ffffff; padding:10px; text-align:center;">Quantity</th>
                                        <th style="background-color:#6a11cb; color:#ffffff; padding:10px; text-align:right;">Price (KES)</th>
                                        <th style="background-color:#6a11cb; color:#ffffff; padding:10px; text-align:right;">Total (KES)</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {{.OrderItems}}
                                    <tr>
                                        <td colspan="3" style="padding:10px; text-align:right; font-weight:bold;">Subtotal</td>
                                        <td style="padding:10px; text-align:right;">KES {{.SubTotal}}</td>
                                    </tr>
									<tr>
                                        <td colspan="3" style="padding:10px; text-align:right; font-weight:bold;">VAT</td>
                                        <td style="padding:10px; text-align:right;">KES {{.VAT}}</td>
                                    </tr>
                                    <tr>
                                        <td colspan="3" style="padding:10px; text-align:right; font-weight:bold;">Discount</td>
                                        <td style="padding:10px; text-align:right;">KES {{.DiscountAmount}}</td>
                                    </tr>
                                    <tr>
                                        <td colspan="3" style="padding:10px; text-align:right; font-weight:bold;">Total</td>
                                        <td style="padding:10px; text-align:right; font-weight:bold;">KES {{.GrandTotal}}</td>
                                    </tr>
                                </tbody>
                            </table>
                        </td>
                    </tr>
                    <tr>
                        <td style="background-color:#f4f4f4; padding:20px; text-align:center;">
                            <p style="margin:0; font-size:12px; color:#777777;">Order placed on {{.OrderDate}}</p>
                            <a href="{{.ProcessOrderLink}}" style="display:inline-block; margin-top:10px; padding:10px 20px; background-color:#6a11cb; color:#ffffff; text-decoration:none; border-radius:5px; font-weight:bold;">Process Order</a>
                        </td>
                    </tr>
                </table>
            </td>
        </tr>
    </table>
</body>
</html>
`

// generateOrderItemsDetails creates the HTML rows for the order items table.
func generateOrderItemsDetails(orderItems []OrderItem) template.HTML {
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
    return template.HTML(sb.String())  // Use template.HTML to prevent escaping
}

// formatPrice formats a float64 price with commas for better readability.
func formatPrice(price float64) string {
    priceStr := fmt.Sprintf("%.2f", price)
    parts := strings.Split(priceStr, ".")
    integerPart := parts[0]
    decimalPart := parts[1]

    var sb strings.Builder
    n := len(integerPart)
    for i, digit := range integerPart {
        sb.WriteByte(byte(digit))
        if (n-i-1)%3 == 0 && i != n-1 {
            sb.WriteByte(',')
        }
    }

    return fmt.Sprintf("%s.%s", sb.String(), decimalPart)
}

// formatShippingAddress formats the shipping address without AddressLine1, AddressLine2, and ZipCode.
func formatShippingAddress(city, state, country string) string {
    return fmt.Sprintf("%s, %s, %s", city, state, country)
}

// formatOrderDate formats the order date string.
func formatOrderDate(orderDate time.Time) (string, error) {
    if orderDate.IsZero() {
        return "", fmt.Errorf("invalid order date")
    }
    return orderDate.Format("January 2, 2006 at 15:04 PM MST"), nil
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
