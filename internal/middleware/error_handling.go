package middleware

import (
	"log"
	"net/http"
	"weblineBackend/internal/app_errors"
)

func ErrorHandling(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Recover from panics
		defer func() {
			if r := recover(); r != nil {
				var err error
				switch t := r.(type) {
				case string:
					err = app_errors.NewInternalError(t, nil)
				case error:
					err = app_errors.NewInternalError(t.Error(), t)
				default:
					err = app_errors.NewInternalError("unknown error", nil)
				}
				log.Printf("panic: %v", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}
