package middleware

import (
	"context"
	"net/http"
	"os"
	"time"
	"weblineBackend/internal/model"
	"weblineBackend/internal/services/i"
	"weblineBackend/pkg/utils"

	"go.uber.org/zap"
)

const (
	sessionCookieName = "webline_session"
	csrfCookieName    = "webline_csrf"
	sessionDuration   = 30 * 24 * time.Hour
)

type sessionContextKey string

var (
	SessionKey   = sessionContextKey("session")
	SessionIDKey = sessionContextKey("sessionID")
	CSRFTokenKey = sessionContextKey("csrfToken")
)

func Session(logger *zap.Logger, sessionService i.SessionService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			secure := isProduction()
			sameSite := getSameSitePolicy()

			// 1. Try to get existing valid session
			sessionID, session, csrfToken, validSession := validateExistingSession(r, sessionService)

			// 2. Create new session if needed
			if !validSession {
				var err error
				sessionID, session, csrfToken, err = createNewSession(w, r, sessionService, secure, sameSite)
				if err != nil {
					logger.Error("Session creation failed", zap.Error(err))
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				}
			}

			// 3. Handle CSRF protection for state-changing methods
			if shouldValidateCSRF(r) {
				if !validateCSRFToken(r, session.CSRFToken) {
					logger.Warn("CSRF validation failed")
					http.Error(w, "Forbidden", http.StatusForbidden)
					return
				}
			}

			// 4. Update context with session information
			ctx = context.WithValue(ctx, SessionKey, session)
			ctx = context.WithValue(ctx, SessionIDKey, sessionID)
			ctx = context.WithValue(ctx, CSRFTokenKey, csrfToken)

			// 5. Set cookies in a browser-compatible way
			setSessionCookies(w, sessionID, csrfToken, secure, sameSite)

			// 6. Proceed with request
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
		return http.SameSiteLaxMode
	}
	return http.SameSiteNoneMode // For modern local development with HTTPS
}

func validateExistingSession(r *http.Request, service i.SessionService) (string, model.Session, string, bool) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return "", model.Session{}, "", false
	}

	session, err := service.GetSessionBySessionID(r.Context(), cookie.Value)
	if err != nil || session.ExpiresAt.Before(time.Now()) {
		return "", model.Session{}, "", false
	}

	csrfCookie, err := r.Cookie(csrfCookieName)
	if err != nil || csrfCookie.Value != session.CSRFToken {
		return "", model.Session{}, "", false
	}

	return cookie.Value, session, csrfCookie.Value, true
}

// Add to session middleware function
func createNewSession(w http.ResponseWriter, r *http.Request, service i.SessionService, secure bool, sameSite http.SameSite) (string, model.Session, string, error) {
	csrfToken, err := utils.GenerateSecureCSRFToken()
	if err != nil {
		return "", model.Session{}, "", err
	}

	session, err := service.CreateSession(r.Context(), nil, time.Now().Add(sessionDuration), csrfToken)
	if err != nil {
		return "", model.Session{}, "", err
	}

	return session.SessionID.String(), session, csrfToken, nil
}

func setSessionCookies(w http.ResponseWriter, sessionID, csrfToken string, secure bool, sameSite http.SameSite) {
	expires := time.Now().Add(sessionDuration)

	// Session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    sessionID,
		Expires:  expires,
		HttpOnly: true,
		Secure:   secure,
		Path:     "/",
		SameSite: sameSite,
	})

	// CSRF cookie
	http.SetCookie(w, &http.Cookie{
		Name:     csrfCookieName,
		Value:    csrfToken,
		Expires:  expires,
		HttpOnly: false, // Accessible via JavaScript
		Secure:   secure,
		Path:     "/",
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
