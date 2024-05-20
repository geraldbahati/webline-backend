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
	UserId   uuid.UUID `json:"userId"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	Role     string    `json:"role"`
	jwt.RegisteredClaims
}

// GenerateTokens generates access and refresh tokens
func GenerateTokens(userId uuid.UUID, username, email, role string) (string, string, time.Time, error) {
	accessToken, err := generateToken(userId, username, email, role, jwtAccessSecret, 24*time.Hour)
	if err != nil {
		return "", "", time.Time{}, err
	}

	refreshToken, expireTime, err := generateTokenWithExpiry(userId, username, email, role, jwtRefreshSecret, 90*24*time.Hour)
	if err != nil {
		return "", "", time.Time{}, err
	}

	return accessToken, refreshToken, expireTime, nil
}

func generateToken(userId uuid.UUID, username, email, role string, secret []byte, duration time.Duration) (string, error) {
	claims := UserClaims{
		UserId:   userId,
		Username: username,
		Email:    email,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userId.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func generateTokenWithExpiry(userId uuid.UUID, username, email, role string, secret []byte, duration time.Duration) (string, time.Time, error) {
	expireTime := time.Now().Add(duration)
	token, err := generateToken(userId, username, email, role, secret, duration)
	if err != nil {
		return "", time.Time{}, err
	}

	return token, expireTime, nil
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

	if claims.UserId == uuid.Nil || claims.Username == "" || claims.Email == "" || claims.Role == "" {
		return "", errors.New("invalid token claims")
	}

	return generateToken(claims.UserId, claims.Username, claims.Email, claims.Role, jwtAccessSecret, 24*time.Hour)
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
	secret := getSecret(isAccessToken)

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil {
		return true
	}

	if !token.Valid {
		return true
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		exp := int64(claims["exp"].(float64))
		return time.Now().Unix() > exp
	}

	return false
}

// getSecret returns the appropriate secret based on the token type
func getSecret(isAccessToken bool) []byte {
	if isAccessToken {
		return jwtAccessSecret
	}
	return jwtRefreshSecret
}
