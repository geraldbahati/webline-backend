package middleware

// // CSRF middleware protects against Cross-Site Request Forgery attacks.
// func CSRF(logger *zap.Logger) func(http.Handler) http.Handler {
// 	return func(next http.Handler) http.Handler {
// 		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 			// Only protect state-changing methods
// 			if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
// 				next.ServeHTTP(w, r)
// 				return
// 			}

// 			// Retrieve the session from the context
// 			session, ok := GetSessionFromContext(r.Context())
// 			if !ok {
// 				logger.Warn("Session not found in context")
// 				http.Error(w, "Invalid session", http.StatusUnauthorized)
// 				return
// 			}

// 			// Get CSRF token from the request header
// 			csrfToken := r.Header.Get("X-CSRF-Token")
// 			if csrfToken == "" {
// 				logger.Warn("CSRF token missing in request")
// 				http.Error(w, "CSRF token missing", http.StatusForbidden)
// 				return
// 			}

// 			// Compare the CSRF token with the one in the session
// 			if csrfToken != session.CSRFToken {
// 				logger.Warn("CSRF token mismatch", zap.String("expected", session.CSRFToken), zap.String("received", csrfToken))
// 				http.Error(w, "Invalid CSRF token", http.StatusForbidden)
// 				return
// 			}

// 			// CSRF token is valid, proceed to the next handler
// 			next.ServeHTTP(w, r)
// 		})
// 	}
// }
