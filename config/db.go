package config

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func Connect() {
	// Em produção a conexão vem da variável DATABASE_URL (definida no systemd),
	// para a senha do banco não ficar no código.
	// O valor abaixo é só um fallback para não quebrar quem ainda não configurou.
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Println("⚠️  DATABASE_URL não definida — usando conexão padrão do código")
		dsn = "user=admin password=admin host=147.79.81.199 port=5432 dbname=bretasimoveisdb sslmode=disable"
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("❌ Erro ao abrir conexão: %v", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("❌ Erro ao conectar ao banco: %v", err)
	}

	DB = db
}
