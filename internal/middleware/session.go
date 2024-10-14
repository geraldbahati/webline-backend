package middleware

// // Constants for session duration
// const sessionDuration = 30 * 24 * time.Hour // 30 days

// // Helper function to set session cookies
// func setSessionCookies(w http.ResponseWriter, sessionID string, csrfToken string, expires time.Time, secure bool, sameSite http.SameSite) {
// 	// Set the session_id cookie
// 	http.SetCookie(w, &http.Cookie{
// 		Name:     "session_id",
// 		Value:    sessionID,
// 		Expires:  expires,
// 		HttpOnly: true,
// 		Secure:   secure,
// 		Path:     "/",
// 		SameSite: sameSite,
// 	})

// 	// Set the csrf_token cookie
// 	http.SetCookie(w, &http.Cookie{
// 		Name:     "csrf_token",
// 		Value:    csrfToken,
// 		Expires:  expires,
// 		HttpOnly: false, // Accessible via JavaScript
// 		Secure:   secure,
// 		Path:     "/",
// 		SameSite: sameSite,
// 	})
// }

// // Helper function to clear session cookies
// func clearSessionCookies(w http.ResponseWriter, secure bool, sameSite http.SameSite) {
// 	// Clear the session_id cookie
// 	http.SetCookie(w, &http.Cookie{
// 		Name:     "session_id",
// 		Value:    "",
// 		Expires:  time.Unix(0, 0),
// 		HttpOnly: true,
// 		Secure:   secure,
// 		Path:     "/",
// 		SameSite: sameSite,
// 	})

// 	// Clear the csrf_token cookie
// 	http.SetCookie(w, &http.Cookie{
// 		Name:     "csrf_token",
// 		Value:    "",
// 		Expires:  time.Unix(0, 0),
// 		HttpOnly: false,
// 		Secure:   secure,
// 		Path:     "/",
// 		SameSite: sameSite,
// 	})
// }

// // Session middleware manages user sessions and CSRF tokens.
// func Session(logger *zap.Logger, sessionService i.SessionService) func(http.Handler) http.Handler {
// 	return func(next http.Handler) http.Handler {
// 		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 			ctx := r.Context()
// 			logger.Info("Session middleware invoked")

// 			// Determine if running in production
// 			isProduction := os.Getenv("ENV") == "production"

// 			// Set Secure and SameSite attributes based on environment
// 			secure := isProduction
// 			var sameSite http.SameSite
// 			if isProduction {
// 				sameSite = http.SameSiteLaxMode
// 			} else {
// 				// In development, omit the SameSite attribute to allow cross-origin cookies
// 				sameSite = http.SameSiteDefaultMode
// 			}

// 			expirationTime := time.Now().Add(sessionDuration)
// 			var sessionID string
// 			var session model.Session
// 			var err error

// 			cookie, err := r.Cookie("session_id")
// 			if err != nil || cookie.Value == "" {
// 				// No existing session, create a new one
// 				session, err = sessionService.CreateSession(ctx, nil, expirationTime) // nil for Guest userID
// 				if err != nil {
// 					logger.Error("Failed to create session", zap.Error(err))
// 					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
// 					return
// 				}
// 				sessionID = session.SessionID.String()

// 				// Set the session_id and csrf_token cookies
// 				setSessionCookies(w, sessionID, session.CSRFToken, session.ExpiresAt, secure, sameSite)
// 			} else {
// 				// Existing session found, validate it
// 				sessionID = cookie.Value
// 				session, err = sessionService.GetSessionBySessionID(ctx, sessionID)
// 				if err != nil {
// 					if errors.Is(err, app_errors.NewSessionNotFoundError()) {
// 						// Session not found, create a new one
// 						logger.Warn("Session not found, creating new session", zap.String("sessionID", sessionID))
// 						session, err = sessionService.CreateSession(ctx, nil, expirationTime)
// 						if err != nil {
// 							logger.Error("Failed to create session", zap.Error(err))
// 							http.Error(w, "Internal Server Error", http.StatusInternalServerError)
// 							return
// 						}
// 						sessionID = session.SessionID.String()

// 						// Set the new session_id and csrf_token cookies
// 						setSessionCookies(w, sessionID, session.CSRFToken, session.ExpiresAt, secure, sameSite)
// 					} else {
// 						// Some other error occurred
// 						logger.Error("Failed to get session by ID", zap.Error(err), zap.String("sessionID", sessionID))
// 						http.Error(w, "Internal Server Error", http.StatusInternalServerError)
// 						return
// 					}
// 				} else {
// 					if session.ExpiresAt.Before(time.Now()) {
// 						// Session expired, delete it and create a new one
// 						err := sessionService.DeleteSessionBySessionID(ctx, sessionID)
// 						if err != nil {
// 							logger.Warn("Failed to delete expired session", zap.Error(err), zap.String("sessionID", sessionID))
// 						}

// 						// Clear the session_id and csrf_token cookies
// 						clearSessionCookies(w, secure, sameSite)

// 						// Create a new session
// 						session, err = sessionService.CreateSession(ctx, nil, expirationTime)
// 						if err != nil {
// 							logger.Error("Failed to create session", zap.Error(err))
// 							http.Error(w, "Internal Server Error", http.StatusInternalServerError)
// 							return
// 						}
// 						sessionID = session.SessionID.String()

// 						// Set the new session_id and csrf_token cookies
// 						setSessionCookies(w, sessionID, session.CSRFToken, session.ExpiresAt, secure, sameSite)
// 					} else {
// 						// Session is valid, update last_activity
// 						err = sessionService.UpdateSessionLastActivity(ctx, sessionID)
// 						if err != nil {
// 							logger.Error("Failed to update session last_activity", zap.Error(err), zap.String("sessionID", sessionID))
// 							http.Error(w, "Internal Server Error", http.StatusInternalServerError)
// 							return
// 						}
// 					}
// 				}
// 			}

// 			// Check for authenticated user
// 			userID, userAuthenticated := GetUserID(ctx)
// 			if userAuthenticated {
// 				err := sessionService.LinkSessionToUser(ctx, sessionID, userID)
// 				if err != nil {
// 					logger.Error("Failed to link session to user", zap.Error(err), zap.String("sessionID", sessionID), zap.Stringer("userID", userID))
// 					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
// 					return
// 				}
// 			}

// 			// Add session to the context
// 			ctx = context.WithValue(ctx, SessionKey, session)
// 			ctx = context.WithValue(ctx, SessionIDKey, sessionID)
// 			next.ServeHTTP(w, r.WithContext(ctx))
// 		})
// 	}
// }

// // GetSessionFromContext retrieves the Session object from the context.
// // Returns the session and a boolean indicating whether it was found.
// func GetSessionFromContext(ctx context.Context) (model.Session, bool) {
// 	session, ok := ctx.Value(SessionKey).(model.Session)
// 	return session, ok
// }

// // GetSessionID retrieves the session ID from the context.
// // Returns the session ID and a boolean indicating whether it was found.
// func GetSessionID(ctx context.Context) (string, bool) {
// 	sessionID, ok := ctx.Value(SessionIDKey).(string)
// 	return sessionID, ok
// }
