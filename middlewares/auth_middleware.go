package middlewares

import (
	"context"
	"net/http"
	"strings"

	"github.com/SauravSuresh/todoapp/utils"
)

type contextKey string

const userIDKey contextKey = "user_id"

func AuthenticationMiddelware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authentication")
		if tokenString == "" {
			http.Error(w, "Missing Authentication token", http.StatusUnauthorized)
		}
		tokenParts := strings.Split(tokenString, "")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			http.Error(w, "Invalid authentication token", http.StatusUnauthorized)
			return
		}
		tokenString = tokenParts[1]
		claims, err := utils.VerifyToken(tokenString)
		if err != nil {
			http.Error(w, "Invalid authentication token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, claims["user_id"])
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserID(r *http.Request) (interface{}, bool) {
	userID := r.Context().Value(userIDKey)
	return userID, userID != nil
}
