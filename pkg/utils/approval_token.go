package utils

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"time"
)

var (
	jwtApprovalTokenSecret = []byte("\xf4\xab\xd3\x7f\xe2\x9c\x8a\xb7\xc1\xde\x13\x8f\xa4\xe5\x92\xbc\xfa\x4e\x21\x5a\x3d\x89\x76\x4f\x1e\x06\x95\x62\xbf\x27\xad")
)

type ApprovalTokenClaims struct {
	RequestID uuid.UUID `json:"requestID"`
	jwt.RegisteredClaims
}

// GenerateApprovalToken generates a JWT for approval purposes.
func GenerateApprovalToken(requestID uuid.UUID, duration time.Duration) (string, time.Time, error) {
	now := time.Now()
	claims := ApprovalTokenClaims{
		RequestID: requestID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   requestID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(jwtApprovalTokenSecret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, claims.ExpiresAt.Time, nil
}

// ParseApprovalToken parses the approval token and returns the claims.
func ParseApprovalToken(tokenString string) (*ApprovalTokenClaims, error) {
	var claims ApprovalTokenClaims
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtApprovalTokenSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("error parsing token: %w", err)
	}

	if !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	return &claims, nil
}

// IsApprovalTokenExpired checks if the approval token has expired.
func IsApprovalTokenExpired(tokenString string) bool {
	claims, err := ParseApprovalToken(tokenString)
	if err != nil {
		return true
	}

	return claims.ExpiresAt.Time.Before(time.Now())
}
