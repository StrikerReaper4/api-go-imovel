package config

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func Connect() {
	// A conexão vem SEMPRE de DATABASE_URL (systemd em produção, run-local.ps1
	// em desenvolvimento). Não há fallback embutido de propósito: um fallback
	// apontando para produção faz qualquer máquina sem a variável definida
	// gravar no banco real — e deixa a senha versionada no repositório.
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL não definida. " +
			"Em desenvolvimento use api-go-imovel/run-local.ps1; " +
			"em produção defina a variável na unit do systemd. " +
			"Formato: postgres://usuario:senha@host:porta/banco?sslmode=disable")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Erro ao abrir conexão: %v", err)
	}

	if err := db.Ping(); err != nil {
		// A mensagem do driver pode conter a senha; por isso o log mostra só o
		// host/banco, que vêm do próprio ambiente.
		log.Fatalf("Erro ao conectar ao banco: %v", err)
	}

	DB = db
}
