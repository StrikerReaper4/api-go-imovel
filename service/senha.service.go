package service

import (
	"apiGo/model"
	"apiGo/repository"
	"apiGo/utils"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/mail"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	senhaMinima    = 8
	validadeReset  = 30 * time.Minute
	frontendPadrao = "https://bretasimoveis.cloud"
)

// Erros "de usuário" — o controller devolve a mensagem como 400.
var (
	ErrSenhaCurta    = fmt.Errorf("a senha precisa ter pelo menos %d caracteres", senhaMinima)
	ErrSenhaAtual    = errors.New("senha atual incorreta")
	ErrEmailInvalido = errors.New("email inválido")
	ErrEmailEmUso    = errors.New("já existe um usuário com esse email")
	ErrNomeVazio     = errors.New("informe o nome")
	ErrTokenInvalido = errors.New("link inválido ou expirado — peça um novo")
)

func IsUserError(err error) bool {
	for _, e := range []error{ErrSenhaCurta, ErrSenhaAtual, ErrEmailInvalido, ErrEmailEmUso, ErrNomeVazio, ErrTokenInvalido} {
		if errors.Is(err, e) {
			return true
		}
	}
	return false
}

func hashSenha(senha string) (string, error) {
	if len(senha) < senhaMinima {
		return "", ErrSenhaCurta
	}
	b, err := bcrypt.GenerateFromPassword([]byte(senha), 10)
	return string(b), err
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// CreateAdminService cria um novo admin (só pode ser chamado por um admin logado).
func CreateAdminService(p model.Pessoa) (model.Pessoa, error) {
	p.Nome = strings.TrimSpace(p.Nome)
	p.Email = strings.ToLower(strings.TrimSpace(p.Email))
	if p.Nome == "" {
		return model.Pessoa{}, ErrNomeVazio
	}
	if _, err := mail.ParseAddress(p.Email); err != nil {
		return model.Pessoa{}, ErrEmailInvalido
	}
	if _, err := repository.FindByEmail(p.Email); err == nil {
		return model.Pessoa{}, ErrEmailEmUso
	}

	hash, err := hashSenha(p.Senha)
	if err != nil {
		return model.Pessoa{}, err
	}
	p.Senha = hash
	p.Role = "admin"

	id, err := repository.InsertRepository(p)
	if err != nil {
		return model.Pessoa{}, err
	}
	p.Id = id
	p.Senha = ""
	return p, nil
}

// ChangePasswordService troca a senha de quem está logado, conferindo a atual.
func ChangePasswordService(userId int, senhaAtual, novaSenha string) error {
	p, err := repository.FindById(userId)
	if err != nil {
		return ErrSenhaAtual
	}
	if bcrypt.CompareHashAndPassword([]byte(p.Senha), []byte(senhaAtual)) != nil {
		return ErrSenhaAtual
	}
	hash, err := hashSenha(novaSenha)
	if err != nil {
		return err
	}
	return repository.UpdateSenha(userId, hash)
}

// ForgotPasswordService envia o link de redefinição se o email for de um admin.
// Para não revelar quais emails existem, o controller sempre responde igual.
func ForgotPasswordService(email string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	p, err := repository.FindByEmail(email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}
	if p.Role != "admin" {
		return nil
	}

	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return err
	}
	token := hex.EncodeToString(buf)

	if err := repository.InsertPasswordReset(hashToken(token), p.Id, time.Now().Add(validadeReset)); err != nil {
		return err
	}

	front := os.Getenv("FRONTEND_URL")
	if front == "" {
		front = frontendPadrao
	}
	link := strings.TrimRight(front, "/") + "/redefinir-senha?token=" + url.QueryEscape(token)

	html := fmt.Sprintf(`
<div style="font-family:Arial,sans-serif;max-width:480px;margin:auto;color:#0f3e58">
  <h2>Redefinição de senha</h2>
  <p>Olá, %s.</p>
  <p>Recebemos um pedido para redefinir a senha do painel da Bretas Imóveis.
     Clique no botão abaixo para criar uma nova senha. O link vale por 30 minutos.</p>
  <p style="text-align:center;margin:32px 0">
    <a href="%s" style="background:#0f3e58;color:#fff;padding:12px 24px;border-radius:24px;text-decoration:none;font-weight:bold">
      Criar nova senha
    </a>
  </p>
  <p style="font-size:12px;color:#666">Se você não pediu isso, ignore este email — sua senha continua a mesma.</p>
</div>`, htmlEscape(p.Nome), link)

	if err := utils.SendEmail(p.Email, "Redefinição de senha — Bretas Imóveis", html); err != nil {
		log.Println("Erro ao enviar email de recuperação:", err)
		return err
	}
	return nil
}

// ResetPasswordService troca a senha usando o token recebido por email.
func ResetPasswordService(token, novaSenha string) error {
	if token == "" {
		return ErrTokenInvalido
	}
	// Valida a senha antes de consumir o token, para não queimar o link à toa.
	hash, err := hashSenha(novaSenha)
	if err != nil {
		return err
	}
	pessoaId, err := repository.ConsumePasswordReset(hashToken(token))
	if err != nil {
		return ErrTokenInvalido
	}
	return repository.UpdateSenha(pessoaId, hash)
}

func htmlEscape(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;").Replace(s)
}
