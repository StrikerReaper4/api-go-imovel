package middlewares

import (
	"apiGo/repository"
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

		if !strings.EqualFold(strings.TrimSpace(role), "admin") {
			http.Error(w, "acesso inválido", http.StatusForbidden)
			return
		}

		// jwt decodifica números como float64
		userId, _ := claims["user_id"].(float64)

		// O token vale por horas; confere no banco se a conta ainda existe e
		// ainda é admin. Sem isso, um admin excluído continuaria mexendo no
		// catálogo até o token expirar.
		p, err := repository.FindById(int(userId))
		if err != nil || !strings.EqualFold(strings.TrimSpace(p.Role), "admin") {
			http.Error(w, "conta sem acesso", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIdKey, int(userId))

		next(w, r.WithContext(ctx))
	}
}
