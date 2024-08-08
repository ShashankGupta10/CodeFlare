package auth

import (
	"net/http"
	"strings"
	"time"

	"codeflare/internal/config"

	"github.com/dgrijalva/jwt-go"
)

var cfg = config.Load()

func GenerateToken(userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})
	return token.SignedString([]byte(cfg.JWTSecret))
}

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing authorization header", http.StatusUnauthorized)
			return
		}

		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 {
			http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
			return
		}

		token, err := jwt.Parse(bearerToken[1], func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	}
}
