// utils/csrf.go
package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
)

// GenerateSecureCSRFToken creates a URL-safe, base64 encoded random token
func GenerateSecureCSRFToken() (string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("failed to generate CSRF token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(tokenBytes), nil
}

// VerifyCSRFToken securely compares tokens
func VerifyCSRFToken(received, stored string) bool {
	return subtle.ConstantTimeCompare(
		[]byte(received),
		[]byte(stored),
	) == 1
}
