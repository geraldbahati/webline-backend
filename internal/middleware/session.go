// middleware/session.go
package middleware

import (
	"context"
	"net/http"
	"os"
	"time"
	"weblineBackend/internal/model"
	"weblineBackend/internal/services/i"

	"go.uber.org/zap"
)

// Session is the middleware responsible for managing user sessions and CSRF tokens.
func Session(logger *zap.Logger, sessionService i.SessionService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			logger.Info("Session middleware")
			cookie, err := r.Cookie("session_id")
			var sessionID string
			var session model.Session

			secure := true
			if os.Getenv("ENV") != "production" {
				secure = false
			}

			if err != nil || cookie.Value == "" {
				// **No existing session, create a new one**
				session, err = sessionService.CreateSession(ctx, nil, time.Now().Add(30*24*time.Hour)) // nil for Guest userID
				if err != nil {
					logger.Error("Failed to create session", zap.Error(err))
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				}
				sessionID = session.SessionID.String()

				// **Set the session_id cookie**
				http.SetCookie(w, &http.Cookie{
					Name:     "session_id",
					Value:    sessionID,
					Expires:  session.ExpiresAt,
					HttpOnly: true,
					Secure:   secure, // Ensure HTTPS in production
					Path:     "/",
					SameSite: http.SameSiteLaxMode,
				})

				// **Set the csrf_token cookie**
				http.SetCookie(w, &http.Cookie{
					Name:     "csrf_token",
					Value:    session.CSRFToken,
					Expires:  session.ExpiresAt,
					HttpOnly: false,  // Accessible via JavaScript
					Secure:   secure, // Ensure HTTPS in production
					Path:     "/",
					SameSite: http.SameSiteLaxMode,
				})
			} else {
				// **Existing session found, validate it**
				sessionID = cookie.Value
				session, err = sessionService.GetSessionBySessionID(ctx, sessionID)
				if err != nil {
					logger.Error("Failed to get session by ID", zap.Error(err))
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				}

				if session.ExpiresAt.Before(time.Now()) {
					// **Session expired, delete it and create a new one**
					_ = sessionService.DeleteSessionBySessionID(ctx, sessionID)

					// **Clear the session_id and csrf_token cookies**
					clearCookie := func(name string) {
						http.SetCookie(w, &http.Cookie{
							Name:     name,
							Value:    "",
							Expires:  time.Unix(0, 0),
							HttpOnly: true,
							Secure:   secure,
							Path:     "/",
							SameSite: http.SameSiteLaxMode,
						})
					}
					clearCookie("session_id")
					clearCookie("csrf_token")

					// **Create a new session**
					session, err = sessionService.CreateSession(ctx, nil, time.Now().Add(30*24*time.Hour))
					if err != nil {
						http.Error(w, "Failed to create session", http.StatusInternalServerError)
						return
					}
					sessionID = session.SessionID.String()

					// **Set the new session_id and csrf_token cookies**
					http.SetCookie(w, &http.Cookie{
						Name:     "session_id",
						Value:    sessionID,
						Expires:  session.ExpiresAt,
						HttpOnly: true,
						Secure:   secure, // Ensure HTTPS in production
						Path:     "/",
						SameSite: http.SameSiteLaxMode,
					})
					http.SetCookie(w, &http.Cookie{
						Name:     "csrf_token",
						Value:    session.CSRFToken,
						Expires:  session.ExpiresAt,
						HttpOnly: false, // Accessible via JavaScript
						Secure:   secure,
						Path:     "/",
						SameSite: http.SameSiteLaxMode,
					})
				} else {
					// **Session is valid, update last_activity**
					err = sessionService.UpdateSessionLastActivity(ctx, sessionID)
					if err != nil {
						http.Error(w, "Failed to update session", http.StatusInternalServerError)
						return
					}
				}
			}

			// **Check for authenticated user**
			userID, userAuthenticated := GetUserID(ctx)
			if userAuthenticated {
				err := sessionService.LinkSessionToUser(ctx, sessionID, userID)
				if err != nil {
					logger.Error("Failed to link session to user", zap.Error(err))
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				}
			}

			// **Add session to the context**
			ctx = context.WithValue(ctx, SessionKey, session)
			ctx = context.WithValue(ctx, SessionIDKey, sessionID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetSessionFromContext retrieves the Session object from the context.
// Returns the session and a boolean indicating whether it was found.
func GetSessionFromContext(ctx context.Context) (model.Session, bool) {
	session, ok := ctx.Value(SessionKey).(model.Session)
	return session, ok
}

// GetSessionID retrieves the session ID from the context.
// Returns the session ID and a boolean indicating whether it was found.
func GetSessionID(ctx context.Context) (string, bool) {
	sessionID, ok := ctx.Value(SessionIDKey).(string)
	return sessionID, ok
}
