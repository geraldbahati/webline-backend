package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"os"
	"time"
	"weblineBackend/internal/model"
	"weblineBackend/internal/services/i"
	"weblineBackend/pkg/utils"

	"github.com/google/uuid"
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

			// 1. Check if user is authenticated via Auth middleware
			userID, isUserAuthenticated := GetUserID(ctx)

			// 2. Try to get existing valid session
			sessionID, session, csrfToken, validSession := validateExistingSession(r, sessionService, logger)

			// 3. Create new session if needed
			if !validSession {
				var err error
				var uid *uuid.UUID
				if isUserAuthenticated {
					uid = &userID
				}
				sessionID, session, csrfToken, err = createNewSession(w, r, sessionService, secure, sameSite, logger, uid)
				if err != nil {
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				}
			}

			// Update last activity before proceeding
			_ = sessionService.UpdateSessionLastActivity(ctx, sessionID)

			// 4. Link session to user if authenticated and session not yet linked
			if isUserAuthenticated && (session.UserID == nil || *session.UserID == uuid.Nil) {
				err := sessionService.LinkSessionToUser(ctx, sessionID, userID)
				if err == nil {
					updatedSession, err := sessionService.GetSessionBySessionID(ctx, sessionID)
					if err == nil {
						session = updatedSession
					}
				}
			}

			// 5. Set user type in context
			userType := UserTypeGuest
			if isUserAuthenticated {
				userType = UserTypeAuthenticated
			}
			ctx = context.WithValue(ctx, userTypeContextKey, userType)

			// 6. Handle CSRF protection for state-changing methods
			if shouldValidateCSRF(r) {
				_, isUserAuthenticated := GetUser(r.Context())
				if !isUserAuthenticated {
					if !validateCSRFToken(r, session.CSRFToken) {
						http.Error(w, "Forbidden", http.StatusForbidden)
						return
					}
				}
			}

			// 7. Update context with session information
			ctx = context.WithValue(ctx, sessionContextKey, &session)
			ctx = context.WithValue(ctx, sessionIDContextKey, sessionID)
			ctx = context.WithValue(ctx, csrfTokenContextKey, csrfToken)

			// 8. Set cookies in a browser-compatible way
			setSessionCookies(w, sessionID, csrfToken, secure, sameSite, logger)

			// 9. Proceed with request
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
		return "", model.Session{}, "", false
	}

	// Decode the cookie value in case it's URL-encoded.
	sessionID, err := url.QueryUnescape(cookie.Value)
	if err != nil {
		return "", model.Session{}, "", false
	}

	session, err := service.GetSessionBySessionID(r.Context(), sessionID)
	if err != nil {
		return "", model.Session{}, "", false
	}
	// Check if the session has expired
	if session.ExpiresAt.Before(time.Now()) {
		return "", model.Session{}, "", false
	}

	csrfCookie, err := r.Cookie(csrfCookieName)
	if err != nil || csrfCookie.Value != session.CSRFToken {
		return "", model.Session{}, "", false
	}

	return sessionID, session, csrfCookie.Value, true
}

// Add to session middleware function
func createNewSession(w http.ResponseWriter, r *http.Request, service i.SessionService, secure bool, sameSite http.SameSite, logger *zap.Logger, uid *uuid.UUID) (string, model.Session, string, error) {
	csrfToken, err := utils.GenerateSecureCSRFToken()
	if err != nil {
		return "", model.Session{}, "", err
	}

	session, err := service.CreateSession(r.Context(), uid, time.Now().Add(sessionDuration), csrfToken)
	if err != nil {
		return "", model.Session{}, "", err
	}

	return session.SessionID.String(), session, csrfToken, nil
}

func setSessionCookies(w http.ResponseWriter, sessionID, csrfToken string, secure bool, sameSite http.SameSite, logger *zap.Logger) {
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
