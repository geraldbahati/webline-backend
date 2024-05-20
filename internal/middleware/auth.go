package middleware

import (
	"net/http"
	"weblineBackend/pkg/utils"
)

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const bearerSchema = "Bearer "

		// get authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// check if authorization header is valid
		if len(authHeader) < len(bearerSchema) || authHeader[:len(bearerSchema)] != bearerSchema {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// get token
		token := authHeader[len(bearerSchema):]
		claims, err := utils.ParseToken(token, true)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// set user id in context
		ctx := utils.SetUserIdInContext(r.Context(), claims.UserId)

		next.ServeHTTP(w, r.WithContext(ctx))

	})
}
