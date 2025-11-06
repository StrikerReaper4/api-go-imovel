package repository

import (
	"apiGo/config"
	"apiGo/model"
	"log"
)

// Insere uma nova pessoa no banco
func InsertRepository(p model.Pessoa) (int, error) {
	query := `
		INSERT INTO pessoas (nome, email, senha, role)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	var lastId int
	err := config.DB.QueryRow(query, p.Nome, p.Email, p.Senha, p.Role).Scan(&lastId)
	if err != nil {
		return 0, err
	}

	return lastId, nil
}

// Busca pessoa pelo e-mail
func FindByEmail(email string) (model.Pessoa, error) {
	var p model.Pessoa
	row := config.DB.QueryRow(`
		SELECT id, nome, email, senha, role
		FROM pessoas
		WHERE email = $1
	`, email)
	log.Println("Usuario encontrado",p)
	err := row.Scan(&p.Id, &p.Nome, &p.Email, &p.Senha, &p.Role)
	if err != nil {
		return model.Pessoa{}, err
	}

	return p, nil
}
