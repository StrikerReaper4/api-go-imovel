package config

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func Connect() {
	dsn := "user=admin password=admin host=147.79.81.199 port=5432 dbname=bretasimoveisdb sslmode=disable"

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
