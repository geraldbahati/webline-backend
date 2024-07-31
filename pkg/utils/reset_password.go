package utils

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	jwtPasswordResetSecret = []byte("\xaa\xdec*\xa7c\xcd\xa6\xe4\xd8l\xc7\x9e\xe7\xa7\xb4\xef\x155\xa4\x898\x18\x9f(\x07\xa3\x0b\x85})")
)

type PasswordResetClaims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

// GeneratePasswordResetToken generates a JWT for password reset purposes.
func GeneratePasswordResetToken(email string, duration time.Duration) (string, time.Time, error) {
	now := time.Now()
	claims := PasswordResetClaims{
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   email,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(jwtPasswordResetSecret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, claims.ExpiresAt.Time, nil
}

// ParsePasswordResetToken parses the password reset token and returns the claims.
func ParsePasswordResetToken(tokenString string) (*PasswordResetClaims, error) {
	var claims PasswordResetClaims
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtPasswordResetSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("error parsing token: %w", err)
	}

	if !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	return &claims, nil
}

// IsPasswordResetTokenExpired checks if the password reset token is expired.
func IsPasswordResetTokenExpired(tokenString string) (bool, error) {
	claims, err := ParsePasswordResetToken(tokenString)
	if err != nil {
		return true, err
	}
	return time.Now().After(claims.ExpiresAt.Time), nil
}
