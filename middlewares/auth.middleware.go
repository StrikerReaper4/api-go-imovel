package middlewares

import (
	"apiGo/utils"
	"context"
	"net/http"
	"strings"
)

type ctxKey string

// UserIdKey guarda no contexto da requisição o id de quem está logado.
const UserIdKey ctxKey = "user_id"

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		authReader := r.Header.Get("Authorization")

		if authReader == "" {
			http.Error(w, "token não fornecido", http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(authReader, "Bearer ")

		claims, err := utils.ValidateToken(tokenStr)

		if err != nil {
			http.Error(w, "token inválido", http.StatusUnauthorized)
			return
		}

		role, ok := claims["role"].(string)

		if !ok {
			http.Error(w, "role inválida", http.StatusUnauthorized)
			return
		}

		if role != "admin" {
			http.Error(w, "acesso inválido", http.StatusForbidden)
			return
		}

		// jwt decodifica números como float64
		userId, _ := claims["user_id"].(float64)
		ctx := context.WithValue(r.Context(), UserIdKey, int(userId))

		next(w, r.WithContext(ctx))
	}
}
