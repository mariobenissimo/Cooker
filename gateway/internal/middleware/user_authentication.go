package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"gitub.com/mariobenissimo/apiGateway/internal/models"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var auth = r.Header.Get("Authorization") //Grab the token from the header
		if auth == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		tknStr := strings.Split(auth, "Bearer ")[1]
		tknStr = strings.TrimSpace(tknStr)
		if tknStr == "" {
			//Token is missing, returns with error code 403 Unauthorized
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		claims := &models.Claims{}

		tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte("secret"), nil
		})
		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if !tkn.Valid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		// ??
		k := models.FavContextKey("user")
		ctx := context.WithValue(r.Context(), k, tknStr)
		// forward request
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
