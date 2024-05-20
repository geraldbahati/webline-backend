package utils

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"log"
	"net/smtp"
	"strings"
	"time"
)

var jwtResetPasswordSecret = []byte("LsM_Zx8sZHuBeC4ty9Aau3x1JuCw-jqeDLS5MxU9bvg=")

func SendResetPasswordEmail(userID uuid.UUID, email string) error {
	// generate reset password token
	resetPasswordToken, err := generateResetPasswordToken(userID)
	if err != nil {
		return err
	}

	// generate reset password link
	resetPasswordLink := fmt.Sprintf("http://localhost:8000/reset-password?token=%s", resetPasswordToken)

	// send email
	err = sendEmail(email, resetPasswordLink)
	if err != nil {
		return err
	}

	return nil
}

func generateResetPasswordToken(userID uuid.UUID) (string, error) {
	// create claims
	claims := UserClaims{
		UserId: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // expires in 24 hours
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID.String(),
		},
	}

	// create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// sign token
	return token.SignedString(jwtResetPasswordSecret)
}

func VerifyResetPasswordToken(tokenString string) (uuid.UUID, error) {
	// parse token
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtResetPasswordSecret, nil
	})

	if err != nil {
		return uuid.UUID{}, err
	}

	// check if token is valid
	if !token.Valid {

		return uuid.UUID{}, errors.New("invalid token")
	}

	// get claims
	claims, ok := token.Claims.(*UserClaims)
	if !ok {
		return uuid.UUID{}, errors.New("invalid token claims")
	}

	return claims.UserId, nil
}

func sendEmail(email string, resetPasswordLink string) error {
	// smtp server configuration
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"
	smtpEmail := "journeytoharvard@gmail.com"
	smtpPassword := "wdib qixk yomq mmpt"

	// email header
	headers := []string{
		"From: " + smtpEmail,
		"To: " + email,
		"Subject: Reset Your Password",
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=\"utf-8\"",
	}
	header := strings.Join(headers, "\r\n")
	log.Printf(resetPasswordLink)
	// email body
	body := `
        <html>
        <head>
            <style>
                body { background-color: #f0f0f0; font-family: Arial, sans-serif; }
                .container { background-color: #fff; padding: 20px; margin: 10px auto; width: 80%; max-width: 600px; }
                .button { background-color: #007bff; color: #ffffff; padding: 10px; text-decoration: none; border-radius: 5px; }
            </style>
        </head>
        <body>
            <div class="container">
                <h2>Reset Your Password</h2>
                <p>Please click the link below to reset your password for your e-commerce account.</p>
                <a href="` + resetPasswordLink + `" class="button">Reset Password</a>
                <p>If you did not request a password reset, please ignore this email.</p>
            </div>
        </body>
        </html>
    `
	message := []byte(header + "\r\n\r\n" + body)

	// send email
	auth := smtp.PlainAuth("", smtpEmail, smtpPassword, smtpHost)
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, smtpEmail, []string{email}, message)
	if err != nil {
		return err
	}
	return nil
}
