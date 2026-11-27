package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/0x24CaptainParrot/gomarket/internal/pkg/services"
)

type ctxKey string

const (
	userIDKey ctxKey = "userID"
)

func AuthUserMiddleware(authService services.Authorization) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var token string

			str := r.Header.Get("Authorization")
			if strings.Contains(str, "Bearer") {
				token = strings.TrimPrefix(str, "Bearer ")
			}

			claims, err := authService.ValidateAccessToken(token)
			if err != nil {
				writeAuthError(w, "unauthorized")
				return
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, userIDKey, claims.UserID)
			h.ServeHTTP(w, r.WithContext(ctx))

		})
	}
}

func writeAuthError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(map[string]string{
		"error":   "authentication failed",
		"message": message,
	})
}
