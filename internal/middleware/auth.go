package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"weblineBackend/pkg/utils"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	// BearerScheme defines the authorization scheme.
	BearerScheme = "Bearer"
)

// ErrorResponse represents the structure of error messages returned to clients.
type ErrorResponse struct {
	Error string `json:"error"`
}

// extractBearerToken extracts the token from the Authorization header.
// Returns the token and a boolean indicating success.
func extractBearerToken(authHeader string) (string, bool) {
	if !strings.HasPrefix(authHeader, BearerScheme+" ") {
		return "", false
	}
	return strings.TrimPrefix(authHeader, BearerScheme+" "), true
}

// writeJSONError sends a JSON-formatted error response with the specified status code and message.
func writeJSONError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	resp := ErrorResponse{Error: message}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		// In case of JSON encoding failure, log the error and send a plain text response.
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// parseToken extracts the user ID from the token.
// Returns the user ID and any parsing error encountered.
func parseToken(token string) (uuid.UUID, error) {
	claims, err := utils.ParseToken(token, true)
	if err != nil {
		return uuid.Nil, err
	}
	return claims.UserId, nil
}

// Auth is a middleware that enforces authentication by validating the Bearer token.
// It sets the user ID in the request context upon successful validation.
func Auth(logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				logger.Warn("Authorization header is missing")
				writeJSONError(w, http.StatusUnauthorized, "Authorization header is required")
				return
			}

			token, ok := extractBearerToken(authHeader)
			if !ok {
				logger.Warn("Authorization header format is invalid", zap.String("header", authHeader))
				writeJSONError(w, http.StatusUnauthorized, "Invalid authorization header format")
				return
			}

			userID, err := parseToken(token)
			if err != nil {
				logger.Warn("Token parsing failed", zap.Error(err))
				writeJSONError(w, http.StatusUnauthorized, "Invalid or expired token")
				return
			}

			// Store the user ID in the context
			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuth is a middleware that optionally parses the Bearer token.
// It sets the user ID in the context if the token is valid but does not enforce authentication.
func OptionalAuth(logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")

			if token, ok := extractBearerToken(authHeader); ok {
				userID, err := parseToken(token)
				if err != nil {
					logger.Warn("OptionalAuth: Token parsing failed", zap.Error(err))
				} else {
					// Store the user ID in the context
					ctx := context.WithValue(r.Context(), UserIDKey, userID)
					r = r.WithContext(ctx)
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetUserID retrieves the user ID from the context.
// Returns the user ID and a boolean indicating whether it was found.
func GetUserID(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(UserIDKey).(uuid.UUID)
	return userID, ok
}
