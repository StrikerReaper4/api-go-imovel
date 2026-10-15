package controller

import (
	"apiGo/middlewares"
	"apiGo/model"
	"apiGo/service"
	"apiGo/utils"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// responde 400 com a mensagem para erros de validação e 500 genérico no resto
func writeServiceError(w http.ResponseWriter, err error) {
	if service.IsUserError(err) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"erro": err.Error()})
		return
	}
	log.Println("Erro interno:", err)
	writeJSON(w, http.StatusInternalServerError, map[string]string{"erro": "erro interno, tente novamente"})
}

func onlyPost(w http.ResponseWriter, r *http.Request) bool {
	if r.Method != http.MethodPost {
		http.Error(w, "método não permitido", http.StatusMethodNotAllowed)
		return false
	}
	return true
}

// POST /admin/criar  (admin logado) — { nome, email, senha }
func CreateAdmin(w http.ResponseWriter, r *http.Request) {
	if !onlyPost(w, r) {
		return
	}
	var p model.Pessoa
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"erro": "dados inválidos"})
		return
	}
	created, err := service.CreateAdminService(p)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

// POST /admin/alterar-senha  (admin logado) — { senha_atual, nova_senha }
func ChangePassword(w http.ResponseWriter, r *http.Request) {
	if !onlyPost(w, r) {
		return
	}
	var body struct {
		SenhaAtual string `json:"senha_atual"`
		NovaSenha  string `json:"nova_senha"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"erro": "dados inválidos"})
		return
	}
	userId, _ := r.Context().Value(middlewares.UserIdKey).(int)
	if err := service.ChangePasswordService(userId, body.SenhaAtual, body.NovaSenha); err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"mensagem": "senha alterada"})
}

// POST /senha/esqueci — { email }
// Sempre responde igual para não revelar quais emails são de admin.
func ForgotPassword(w http.ResponseWriter, r *http.Request) {
	if !onlyPost(w, r) {
		return
	}
	var body struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"erro": "dados inválidos"})
		return
	}
	if err := service.ForgotPasswordService(body.Email); err != nil {
		if errors.Is(err, utils.ErrEmailNaoConfigurado) {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"erro": "envio de email ainda não configurado no servidor"})
			return
		}
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"mensagem": "se o email for de um admin, o link foi enviado"})
}

// POST /senha/redefinir — { token, nova_senha }
func ResetPassword(w http.ResponseWriter, r *http.Request) {
	if !onlyPost(w, r) {
		return
	}
	var body struct {
		Token     string `json:"token"`
		NovaSenha string `json:"nova_senha"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"erro": "dados inválidos"})
		return
	}
	if err := service.ResetPasswordService(body.Token, body.NovaSenha); err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"mensagem": "senha redefinida"})
}
