package utils

import (
	"errors"
	"fmt"
	"mime"
	"net/smtp"
	"os"
	"strings"
)

// ErrEmailNaoConfigurado indica que faltam as variáveis SMTP no servidor.
var ErrEmailNaoConfigurado = errors.New("envio de email não configurado")

// SendEmail envia um email HTML pelo Gmail (ou outro SMTP).
//
// Variáveis de ambiente:
//   SMTP_USER  conta remetente, ex: imoveisbretas@gmail.com
//   SMTP_PASS  "senha de app" do Google (NÃO é a senha normal da conta)
//   SMTP_HOST  opcional, padrão smtp.gmail.com
//   SMTP_PORT  opcional, padrão 587
func SendEmail(to, subject, html string) error {
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")
	if user == "" || pass == "" {
		return ErrEmailNaoConfigurado
	}
	host := os.Getenv("SMTP_HOST")
	if host == "" {
		host = "smtp.gmail.com"
	}
	port := os.Getenv("SMTP_PORT")
	if port == "" {
		port = "587"
	}

	// Evita injeção de cabeçalhos via quebra de linha.
	if strings.ContainsAny(to, "\r\n") {
		return errors.New("destinatário inválido")
	}

	msg := strings.Join([]string{
		"From: Bretas Imóveis <" + user + ">",
		"To: " + to,
		"Subject: " + mime.QEncoding.Encode("utf-8", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=UTF-8",
		"",
		html,
	}, "\r\n")

	auth := smtp.PlainAuth("", user, pass, host)
	if err := smtp.SendMail(host+":"+port, auth, user, []string{to}, []byte(msg)); err != nil {
		return fmt.Errorf("falha ao enviar email: %w", err)
	}
	return nil
}
