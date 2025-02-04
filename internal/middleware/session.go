package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"
	"weblineBackend/internal/model"
	"weblineBackend/internal/services/i"
	"weblineBackend/pkg/utils"

	"go.uber.org/zap"
)

const (
	sessionCookieName     = "webline_session"
	csrfCookieName        = "webline_csrf"
	sessionDuration       = 30 * 24 * time.Hour
	UserTypeGuest         = "guest"
	UserTypeAuthenticated = "authenticated"
)

func Session(logger *zap.Logger, sessionService i.SessionService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			secure := isProduction()
			sameSite := getSameSitePolicy()

			logger.Debug("Session middleware", zap.String("secure", fmt.Sprintf("%t", secure)))

			// 1. Check if user is authenticated via Auth middleware
			logger.Debug("Checking if user is authenticated")
			user, isUserAuthenticated := GetUser(ctx)
			logger.Debug("User authenticated", zap.String("isUserAuthenticated", fmt.Sprintf("%t", isUserAuthenticated)))

			// 2. Try to get existing valid session
			logger.Debug("Validating existing session")
			sessionID, session, csrfToken, validSession := validateExistingSession(r, sessionService, logger)
			logger.Debug("Session validated", zap.String("sessionID", sessionID))

			logger.Debug("Session", zap.Any("session", session))
			logger.Debug("CSRF token", zap.String("csrfToken", csrfToken))
			logger.Debug("Valid session", zap.String("validSession", fmt.Sprintf("%t", validSession)))

			// 3. Create new session if needed
			if !validSession {
				logger.Debug("Creating new session")
				var err error
				sessionID, session, csrfToken, err = createNewSession(w, r, sessionService, secure, sameSite, logger)
				if err != nil {
					logger.Error("Session creation failed", zap.Error(err))
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				}
				logger.Debug("New session created", zap.String("sessionID", sessionID))
			}

			// 4. Link session to user if authenticated
			if isUserAuthenticated && session.UserID == nil {
				logger.Debug("Linking session to user")
				err := sessionService.LinkSessionToUser(ctx, sessionID, user.UserID)
				if err != nil {
					logger.Warn("Failed to link session to user",
						zap.Error(err),
						zap.String("sessionID", sessionID),
						zap.String("userID", user.UserID.String()))
				}
				logger.Debug("Session linked to user", zap.String("sessionID", sessionID))
			}

			// 5. Set user type in context
			logger.Debug("Setting user type in context")
			userType := UserTypeGuest
			if isUserAuthenticated {
				userType = UserTypeAuthenticated
			}
			ctx = context.WithValue(ctx, userTypeContextKey, userType)
			logger.Debug("User type set in context", zap.String("userType", userType))

			// 6. Handle CSRF protection for state-changing methods
			logger.Debug("Validating CSRF token")
			if shouldValidateCSRF(r) {
				if !validateCSRFToken(r, session.CSRFToken) {
					logger.Debug("CSRF validation failed", zap.String("csrfToken", csrfToken))
					logger.Warn("CSRF validation failed")
					http.Error(w, "Forbidden", http.StatusForbidden)
					return
				}
				logger.Debug("CSRF validation passed")
			}

			// 7. Update context with session information
			logger.Debug("Updating context with session information")
			ctx = context.WithValue(ctx, sessionContextKey, &session)
			ctx = context.WithValue(ctx, sessionIDContextKey, sessionID)
			ctx = context.WithValue(ctx, csrfTokenContextKey, csrfToken)

			// 8. Set cookies in a browser-compatible way
			logger.Debug("Setting cookies")
			setSessionCookies(w, sessionID, csrfToken, secure, sameSite, logger)

			// 9. Proceed with request
			logger.Debug("Proceeding with request")
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Helper functions
func isProduction() bool {
	return os.Getenv("ENV") == "production"
}

func getSameSitePolicy() http.SameSite {
	if isProduction() {
		// Use SameSite=None for cross-origin requests over HTTPS.
		return http.SameSiteNoneMode
	}
	// In development, use Lax (or Strict) to ensure the cookie is sent over HTTP.
	return http.SameSiteLaxMode
}

func validateExistingSession(r *http.Request, service i.SessionService, logger *zap.Logger) (string, model.Session, string, bool) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		logger.Error("Error getting session cookie", zap.Error(err))
		return "", model.Session{}, "", false
	}

	logger.Debug("Validating existing session", zap.String("sessionID", cookie.Value))

	session, err := service.GetSessionBySessionID(r.Context(), cookie.Value)
	if err != nil || session.ExpiresAt.Before(time.Now()) {
		logger.Error("Session validation failed", zap.Error(err))
		return "", model.Session{}, "", false
	}

	logger.Debug("Session validated", zap.String("sessionID", cookie.Value))

	csrfCookie, err := r.Cookie(csrfCookieName)
	if err != nil || csrfCookie.Value != session.CSRFToken {
		logger.Error("CSRF validation failed", zap.Error(err))
		return "", model.Session{}, "", false
	}

	logger.Debug("CSRF validation passed", zap.String("csrfToken", csrfCookie.Value))

	return cookie.Value, session, csrfCookie.Value, true
}

// Add to session middleware function
func createNewSession(w http.ResponseWriter, r *http.Request, service i.SessionService, secure bool, sameSite http.SameSite, logger *zap.Logger) (string, model.Session, string, error) {
	logger.Debug("Creating new session")
	csrfToken, err := utils.GenerateSecureCSRFToken()
	if err != nil {
		logger.Error("Failed to generate CSRF token", zap.Error(err))
		return "", model.Session{}, "", err
	}

	session, err := service.CreateSession(r.Context(), nil, time.Now().Add(sessionDuration), csrfToken)
	if err != nil {
		logger.Error("Failed to create session", zap.Error(err))
		return "", model.Session{}, "", err
	}

	logger.Debug("New session created", zap.String("sessionID", session.SessionID.String()))

	return session.SessionID.String(), session, csrfToken, nil
}

func setSessionCookies(w http.ResponseWriter, sessionID, csrfToken string, secure bool, sameSite http.SameSite, logger *zap.Logger) {
	logger.Debug("Setting session cookies")
	expires := time.Now().Add(sessionDuration)

	// Session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    sessionID,
		Path:     "/",
		Expires:  expires,
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	})
	logger.Debug("Session cookie set")

	// CSRF cookie
	http.SetCookie(w, &http.Cookie{
		Name:     csrfCookieName,
		Value:    csrfToken,
		Path:     "/",
		Expires:  expires,
		HttpOnly: false, // Accessible via JavaScript
		Secure:   secure,
		SameSite: sameSite,
	})
	logger.Debug("CSRF cookie set")
	logger.Debug("Cookies set")
}

func shouldValidateCSRF(r *http.Request) bool {
	// Validate CSRF for non-GET/HEAD/OPTIONS requests
	return r.Method == http.MethodPost ||
		r.Method == http.MethodPut ||
		r.Method == http.MethodPatch ||
		r.Method == http.MethodDelete
}

// Update CSRF validation
func validateCSRFToken(r *http.Request, expectedToken string) bool {
	receivedToken := r.Header.Get("X-CSRF-Token")
	if receivedToken == "" {
		receivedToken = r.FormValue("csrf_token")
	}

	// Remove direct logger.Debug call since logger is undefined
	return utils.VerifyCSRFToken(receivedToken, expectedToken)
}

func GetUserType(ctx context.Context) string {
	if userType, ok := ctx.Value(userTypeContextKey).(string); ok {
		return userType
	}
	return UserTypeGuest // Default to guest if not set
}

func GetSessionFromContext(ctx context.Context) (*model.Session, error) {
	sess, ok := ctx.Value(sessionContextKey).(*model.Session)
	if !ok || sess == nil {
		return nil, errors.New("session not found in context")
	}
	if sess.IsExpired() {
		return nil, errors.New("session has expired")
	}
	// Optionally log if the session is near expiry.
	if time.Until(sess.ExpiresAt) < 10*time.Minute {
		// Ideally, pass in a logger; here we simply note that in production you might log this.
		// Example: logger.Warn("Session is about to expire", zap.Duration("time_left", time.Until(sess.ExpiresAt)))
	}
	return sess, nil
}

func GetCookieSecureFlag() bool {
	// In production, we want Secure cookies; otherwise, not.
	return isProduction()
}

func GetCookieSameSitePolicy() http.SameSite {
	if isProduction() {
		// Use SameSite=None for cross-origin requests over HTTPS.
		return http.SameSiteNoneMode
	}
	// In development, use Lax (or Strict) to ensure the cookie is sent over HTTP.
	return http.SameSiteLaxMode
}
