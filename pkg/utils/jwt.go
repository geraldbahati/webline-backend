package utils

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	jwtAccessSecret  = []byte("eK_sS2AgDstNYrh0Bx5LK3nPx-z1h2l_ZdjchgQjvyA=")
	jwtRefreshSecret = []byte("bS4RqAvfuWhiAjZJ_104wBUcDAbp4cEt2ChP1IYskI8=")
)

type UserClaims struct {
	UserId uuid.UUID `json:"userId"`
	Email  string    `json:"email"`
	jwt.RegisteredClaims
}

// GenerateTokens generates access and refresh tokens
func GenerateTokens(userId uuid.UUID, email string) (string, string, time.Time, error) {
	accessToken, err := generateToken(userId, email, jwtAccessSecret, 24*time.Hour)
	if err != nil {
		return "", "", time.Time{}, err
	}

	refreshToken, expireTime, err := generateTokenWithExpiry(userId, email, jwtRefreshSecret, 90*24*time.Hour)
	if err != nil {
		return "", "", time.Time{}, err
	}

	return accessToken, refreshToken, expireTime, nil
}

func generateToken(userId uuid.UUID, email string, secret []byte, duration time.Duration) (string, error) {
	now := time.Now()
	claims := UserClaims{
		UserId: userId,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   userId.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func generateTokenWithExpiry(userId uuid.UUID, email string, secret []byte, duration time.Duration) (string, time.Time, error) {
	token, err := generateToken(userId, email, secret, duration)
	if err != nil {
		return "", time.Time{}, err
	}

	return token, time.Now().Add(duration), nil
}

// ParseToken parses the token and returns the claims
func ParseToken(tokenString string, isAccessToken bool) (*UserClaims, error) {
	var claims UserClaims
	secret := getSecret(isAccessToken)

	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("error parsing token: %w", err)
	}

	if !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	return &claims, nil
}

// RefreshToken generates a new access token
func RefreshToken(refreshToken string) (string, error) {
	claims, err := ParseToken(refreshToken, false)
	if err != nil {
		return "", err
	}

	if claims.UserId == uuid.Nil || claims.Email == "" {
		return "", errors.New("invalid token claims")
	}

	return generateToken(claims.UserId, claims.Email, jwtAccessSecret, 24*time.Hour)
}

// ValidateToken validates the token
func ValidateToken(tokenString string, isAccessToken bool) error {
	secret := getSecret(isAccessToken)

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil {
		return err
	}

	if !token.Valid {
		return jwt.ErrSignatureInvalid
	}

	return nil
}

// IsTokenExpired checks if the token is expired
func IsTokenExpired(tokenString string, isAccessToken bool) bool {
	claims, err := ParseToken(tokenString, isAccessToken)
	if err != nil {
		return true
	}
	return time.Now().After(claims.ExpiresAt.Time)
}

// getSecret returns the appropriate secret based on the token type
func getSecret(isAccessToken bool) []byte {
	if isAccessToken {
		return jwtAccessSecret
	}
	return jwtRefreshSecret
}
