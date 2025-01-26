package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"weblineBackend/pkg/utils"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const bearerScheme = "Bearer"

// User represents the authenticated or guest user
type User struct {
	UserID  uuid.UUID // For authenticated users
	GuestID string    // For guest users
	IsGuest bool
}

// ErrorResponse represents the structure of error messages returned to clients
type ErrorResponse struct {
	Error string `json:"error"`
}

// Auth middleware handles both authenticated and guest users
func Auth(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				logger.Warn("Authorization header is missing")
				writeJSONError(w, http.StatusUnauthorized, "Authorization header is required")
				return
			}

			tokenString, err := getTokenFromHeader(authHeader)
			if err != nil {
				logger.Warn("Invalid authorization header format", zap.Error(err))
				writeJSONError(w, http.StatusUnauthorized, "Invalid authorization header format")
				return
			}

			// Parse and validate the token
			claims, err := utils.ParseToken(tokenString, true)
			if err != nil {
				logger.Warn("Token parsing failed", zap.Error(err))
				writeJSONError(w, http.StatusUnauthorized, "Invalid or expired token")
				return
			}

			// Create a User struct based on the token claims
			user, err := createUserFromClaims(claims)
			if err != nil {
				logger.Warn("Invalid token role", zap.Error(err))
				writeJSONError(w, http.StatusUnauthorized, "Invalid token role")
				return
			}

			// Store the user in the context
			ctx := context.WithValue(r.Context(), UserContextKey, user)
			// Optionally, store userID for easy access
			if !user.IsGuest {
				ctx = context.WithValue(ctx, UserIDKey, user.UserID)
			}

			// After validating the token and creating the user
			session, _, err := GetSessionFromContext(r.Context())
			if err == nil && session.UserID != nil {
				// Verify the session belongs to the authenticated user
				if user.UserID != *session.UserID {
					logger.Warn("Session user ID mismatch",
						zap.String("tokenUserID", user.UserID.String()),
						zap.String("sessionUserID", session.UserID.String()))
					writeJSONError(w, http.StatusUnauthorized, "Invalid session")
					return
				}
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// getTokenFromHeader extracts the token from the Authorization header
func getTokenFromHeader(authHeader string) (string, error) {
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], bearerScheme) {
		return "", errors.New("invalid authorization header format")
	}
	return strings.TrimSpace(parts[1]), nil
}

// createUserFromClaims creates a User struct based on JWT claims
func createUserFromClaims(claims interface{}) (User, error) {
	switch c := claims.(type) {
	case *utils.UserClaims:
		switch c.Role {
		case "guest":
			return User{
				GuestID: c.GuestID,
				IsGuest: true,
			}, nil
		case "user":
			return User{
				UserID:  c.UserID,
				IsGuest: false,
			}, nil
		default:
			return User{}, errors.New("invalid token role")
		}
	case *utils.GuestClaims:
		return User{
			GuestID: c.GuestID.String(),
			IsGuest: true,
		}, nil
	default:
		return User{}, errors.New("invalid claims type")
	}
}

// writeJSONError sends a JSON-formatted error response
func writeJSONError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	resp := ErrorResponse{Error: message}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		// In case of JSON encoding failure, send a plain text response
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// GetUser retrieves the user from the context
func GetUser(ctx context.Context) (User, bool) {
	user, ok := ctx.Value(UserContextKey).(User)
	return user, ok
}

// GetUserID retrieves the user ID from the context
func GetUserID(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(UserIDKey).(uuid.UUID)
	return userID, ok
}

// OptionalAuth middleware optionally parses the Bearer token
func OptionalAuth(logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				tokenString, err := getTokenFromHeader(authHeader)
				if err != nil {
					logger.Warn("Invalid authorization header format", zap.Error(err))
					// Proceed without setting user in context
					next.ServeHTTP(w, r)
					return
				}

				// Parse and validate the token
				claims, err := utils.ParseToken(tokenString, true)
				if err != nil {
					logger.Warn("OptionalAuth: Token parsing failed", zap.Error(err))
					// Proceed without setting user in context
					next.ServeHTTP(w, r)
					return
				}

				// Create a User struct based on the token claims
				user, err := createUserFromClaims(claims)
				if err != nil {
					logger.Warn("OptionalAuth: Invalid token role", zap.Error(err))
					// Proceed without setting user in context
					next.ServeHTTP(w, r)
					return
				}

				// Store the user in the context
				ctx := context.WithValue(r.Context(), UserContextKey, user)

				ctx = context.WithValue(ctx, UserIDKey, user.UserID)

				r = r.WithContext(ctx)
			}

			next.ServeHTTP(w, r)
		})
	}
}
