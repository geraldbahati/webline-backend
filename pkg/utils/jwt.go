// internal/utils/token.go

package utils

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// UserClaims represents the JWT claims for a user or guest
type UserClaims struct {
	UserID  uuid.UUID `json:"userId,omitempty"`
	GuestID string    `json:"guestId,omitempty"`
	Email   string    `json:"email,omitempty"`
	Role    string    `json:"role,omitempty"` // "user" or "guest"
	jwt.RegisteredClaims
}

// GuestClaims represents the JWT claims for guest users
type GuestClaims struct {
	GuestID uuid.UUID `json:"guestId"`
	Role    string    `json:"role"` // "guest"
	jwt.RegisteredClaims
}

// Initialize secret keys from environment variables
var (
	jwtAccessSecret  = []byte(getEnv("JWT_ACCESS_SECRET", "default_access_secret"))
	jwtRefreshSecret = []byte(getEnv("JWT_REFRESH_SECRET", "default_refresh_secret"))
	jwtGuestSecret   = []byte(getEnv("JWT_GUEST_SECRET", "default_guest_secret"))
	jwtVerifySecret  = []byte(getEnv("JWT_VERIFY_SECRET", "default_verify_secret"))
)

// getEnv fetches environment variables or returns a default value
func getEnv(key, defaultVal string) string {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		if defaultVal == "" {
			log.Fatalf("Environment variable %s is required but not set", key)
		}
		return defaultVal
	}
	return value
}

// GenerateTokens generates access and refresh tokens for authenticated users
func GenerateTokens(userID uuid.UUID, email string) (accessToken string, refreshToken string, refreshTokenExpiry time.Time, err error) {
	accessToken, err = generateToken(userID, uuid.Nil, "user", jwtAccessSecret, 15*time.Minute)
	if err != nil {
		return "", "", time.Time{}, err
	}

	refreshToken, refreshTokenExpiry, err = generateTokenWithExpiry(userID, uuid.Nil, "user", jwtRefreshSecret, 30*24*time.Hour)
	if err != nil {
		return "", "", time.Time{}, err
	}

	return accessToken, refreshToken, refreshTokenExpiry, nil
}

// GenerateGuestToken generates an access token for guest users
func GenerateGuestToken() (string, error) {
	guestID := uuid.New()
	return generateToken(uuid.Nil, guestID, "guest", jwtGuestSecret, 30*24*time.Hour)
}

// generateToken creates a JWT token with the specified claims and duration
func generateToken(userID uuid.UUID, guestID uuid.UUID, role string, secret []byte, duration time.Duration) (string, error) {
	if len(secret) == 0 {
		return "", errors.New("secret key is not set")
	}

	now := time.Now()
	var claims jwt.Claims

	if role == "user" {
		claims = UserClaims{
			UserID: userID,
			Email:  "", // Populate as needed
			Role:   role,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
				IssuedAt:  jwt.NewNumericDate(now),
				NotBefore: jwt.NewNumericDate(now),
				Subject:   userID.String(),
				ID:        uuid.NewString(),
			},
		}
	} else if role == "guest" {
		claims = GuestClaims{
			GuestID: guestID,
			Role:    role,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
				IssuedAt:  jwt.NewNumericDate(now),
				NotBefore: jwt.NewNumericDate(now),
				Subject:   guestID.String(),
				ID:        uuid.NewString(),
			},
		}
	} else {
		return "", errors.New("invalid role")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

// generateTokenWithExpiry returns a token and its expiration time
func generateTokenWithExpiry(userID uuid.UUID, guestID uuid.UUID, role string, secret []byte, duration time.Duration) (string, time.Time, error) {
	token, err := generateToken(userID, guestID, role, secret, duration)
	if err != nil {
		return "", time.Time{}, err
	}

	return token, time.Now().Add(duration), nil
}

// ParseToken parses the token and returns the claims
func ParseToken(tokenString string, isAccessToken bool) (interface{}, error) {
	// Step 1: Parse without verification to extract the role
	unverifiedClaims := jwt.MapClaims{}
	_, _, err := new(jwt.Parser).ParseUnverified(tokenString, &unverifiedClaims)
	if err != nil {
		return nil, fmt.Errorf("error parsing token unverified: %w", err)
	}

	role, ok := unverifiedClaims["role"].(string)
	if !ok {
		return nil, errors.New("role claim not found or invalid")
	}

	// Step 2: Select the appropriate secret based on role and token type
	secret := getSecret(isAccessToken, role)
	if len(secret) == 0 {
		return nil, errors.New("secret key is not set for the token role")
	}

	// Step 3: Parse and verify the token with the correct secret
	var claims interface{}
	switch role {
	case "user":
		claims = &UserClaims{}
	case "guest":
		claims = &GuestClaims{}
	default:
		return nil, errors.New("invalid role in token")
	}

	token, err := jwt.ParseWithClaims(tokenString, claims.(jwt.Claims), func(token *jwt.Token) (interface{}, error) {
		// Ensure the signing method is HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	}, jwt.WithValidMethods([]string{"HS256"}))

	if err != nil {
		return nil, fmt.Errorf("error parsing token: %w", err)
	}

	if !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	return claims, nil
}

// getSecret returns the appropriate secret based on the token type and role
func getSecret(isAccessToken bool, role string) []byte {
	switch role {
	case "user":
		if isAccessToken {
			return jwtAccessSecret
		}
		return jwtRefreshSecret
	case "guest":
		if isAccessToken {
			return jwtGuestSecret
		}
		// Assuming you have a separate secret for guest refresh tokens
		// If not, you can reuse jwtGuestSecret or define a new one
		return jwtGuestSecret
	default:
		return nil
	}
}

// GenerateGuestTokens generates access and refresh tokens for guest users
func GenerateGuestTokens() (accessToken string, refreshToken string, refreshTokenExpiry time.Time, err error) {
	guestID := uuid.New()

	accessToken, err = generateToken(uuid.Nil, guestID, "guest", jwtGuestSecret, 30*24*time.Hour)
	if err != nil {
		return "", "", time.Time{}, err
	}

	refreshToken, refreshTokenExpiry, err = generateTokenWithExpiry(uuid.Nil, guestID, "guest", jwtGuestSecret, 60*24*time.Hour)
	if err != nil {
		return "", "", time.Time{}, err
	}

	return accessToken, refreshToken, refreshTokenExpiry, nil
}

// RefreshToken generates a new access token using a valid refresh token
func RefreshToken(refreshToken string) (string, error) {
	claims, err := ParseToken(refreshToken, false)
	if err != nil {
		return "", err
	}

	switch c := claims.(type) {
	case *UserClaims:
		if c.Role != "user" || c.UserID == uuid.Nil {
			return "", errors.New("invalid refresh token claims for user")
		}
		return generateToken(c.UserID, uuid.Nil, "user", jwtAccessSecret, 15*time.Minute)
	case *GuestClaims:
		if c.Role != "guest" || c.GuestID == uuid.Nil {
			return "", errors.New("invalid refresh token claims for guest")
		}
		return generateToken(uuid.Nil, c.GuestID, "guest", jwtGuestSecret, 15*time.Minute)
	default:
		return "", errors.New("invalid token claims type")
	}
}

// ValidateToken validates the token without parsing the claims
func ValidateToken(tokenString string, isAccessToken bool) error {
	_, err := ParseToken(tokenString, isAccessToken)
	return err
}

// IsTokenExpired checks if the token is expired
func IsTokenExpired(tokenString string, isAccessToken bool) (bool, error) {
	claims, err := ParseToken(tokenString, isAccessToken)
	if err != nil {
		return true, err
	}

	var exp time.Time
	switch c := claims.(type) {
	case *UserClaims:
		exp = c.ExpiresAt.Time
	case *GuestClaims:
		exp = c.ExpiresAt.Time
	default:
		return true, errors.New("invalid claims type")
	}

	return time.Now().After(exp), nil
}

// GenerateVerificationToken creates a token for email verification
func GenerateVerificationToken(email string) (string, time.Time, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(now.Add(1 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		Subject:   email,
		ID:        uuid.NewString(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtVerifySecret)
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, claims.ExpiresAt.Time, nil
}

// ParseEmailVerificationToken parses the email verification token and returns the email
func ParseEmailVerificationToken(tokenString string) (string, time.Time, error) {
	var claims jwt.RegisteredClaims
	if len(jwtVerifySecret) == 0 {
		return "", time.Time{}, errors.New("secret key is not set")
	}

	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		// Ensure the signing method is HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtVerifySecret, nil
	}, jwt.WithValidMethods([]string{"HS256"}))

	if err != nil {
		return "", time.Time{}, fmt.Errorf("error parsing token: %w", err)
	}

	if !token.Valid {
		return "", time.Time{}, jwt.ErrSignatureInvalid
	}

	return claims.Subject, claims.ExpiresAt.Time, nil
}
