package config

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func Connect() {
	dsn := "postgresql://admin:UiE6EVgThtpXBDRUkf2xiBJCYFLYP1Cn@dpg-d419off5r7bs739aa2ug-a.oregon-postgres.render.com/imovel_q1jv"

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
