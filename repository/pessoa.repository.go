package repository

import (
	"apiGo/config"
	"apiGo/model"
	"log"
	"time"
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
		WHERE LOWER(email) = LOWER($1)
	`, email)
	log.Println("Usuario encontrado",p)
	err := row.Scan(&p.Id, &p.Nome, &p.Email, &p.Senha, &p.Role)
	if err != nil {
		return model.Pessoa{}, err
	}

	return p, nil
}

// Busca pessoa pelo id
func FindById(id int) (model.Pessoa, error) {
	var p model.Pessoa
	err := config.DB.QueryRow(`
		SELECT id, nome, email, senha, role
		FROM pessoas
		WHERE id = $1
	`, id).Scan(&p.Id, &p.Nome, &p.Email, &p.Senha, &p.Role)
	if err != nil {
		return model.Pessoa{}, err
	}
	return p, nil
}

// Atualiza a senha (já em hash) de uma pessoa
func UpdateSenha(id int, senhaHash string) error {
	_, err := config.DB.Exec(`UPDATE pessoas SET senha = $1 WHERE id = $2`, senhaHash, id)
	return err
}

// ── Recuperação de senha ──────────────────────────────────────────────────────
// Guardamos só o hash (sha256) do token: quem tiver acesso ao banco não consegue
// usar os links que estão na caixa de entrada do admin.

// EnsurePasswordResetTable cria a tabela na primeira execução.
func EnsurePasswordResetTable() error {
	_, err := config.DB.Exec(`
		CREATE TABLE IF NOT EXISTS password_resets (
			token_hash TEXT PRIMARY KEY,
			pessoa_id  INTEGER NOT NULL REFERENCES pessoas(id) ON DELETE CASCADE,
			expires_at TIMESTAMPTZ NOT NULL,
			used       BOOLEAN NOT NULL DEFAULT FALSE
		)
	`)
	return err
}

func InsertPasswordReset(tokenHash string, pessoaId int, expiresAt time.Time) error {
	// Um pedido novo invalida os anteriores da mesma pessoa.
	if _, err := config.DB.Exec(`DELETE FROM password_resets WHERE pessoa_id = $1`, pessoaId); err != nil {
		return err
	}
	_, err := config.DB.Exec(`
		INSERT INTO password_resets (token_hash, pessoa_id, expires_at)
		VALUES ($1, $2, $3)
	`, tokenHash, pessoaId, expiresAt)
	return err
}

// ConsumePasswordReset marca o token como usado e devolve o id da pessoa.
// Falha se o token não existir, já tiver sido usado ou estiver expirado.
func ConsumePasswordReset(tokenHash string) (int, error) {
	var pessoaId int
	err := config.DB.QueryRow(`
		UPDATE password_resets SET used = TRUE
		WHERE token_hash = $1 AND used = FALSE AND expires_at > NOW()
		RETURNING pessoa_id
	`, tokenHash).Scan(&pessoaId)
	return pessoaId, err
}

// ── Administradores ──────────────────────────────────────────────────────────

// ListAdmins devolve os admins sem a senha.
func ListAdmins() ([]model.Pessoa, error) {
	rows, err := config.DB.Query(`
		SELECT id, nome, email, role
		FROM pessoas
		WHERE LOWER(TRIM(role)) = 'admin'
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	admins := []model.Pessoa{}
	for rows.Next() {
		var p model.Pessoa
		if err := rows.Scan(&p.Id, &p.Nome, &p.Email, &p.Role); err != nil {
			return nil, err
		}
		admins = append(admins, p)
	}
	return admins, rows.Err()
}

// DeleteAdminTransferindo apaga o admin `alvo` e, antes, passa os imóveis dele
// para `novoDono`. Sem isso o ON DELETE CASCADE de imoveis.id_pessoa apagaria
// todos os imóveis cadastrados por ele. Tudo numa transação: ou faz as duas
// coisas, ou nenhuma.
func DeleteAdminTransferindo(alvo, novoDono int) (int64, error) {
	tx, err := config.DB.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	res, err := tx.Exec(`UPDATE imoveis SET id_pessoa = $1 WHERE id_pessoa = $2`, novoDono, alvo)
	if err != nil {
		return 0, err
	}
	movidos, _ := res.RowsAffected()

	if _, err := tx.Exec(`DELETE FROM pessoas WHERE id = $1 AND LOWER(TRIM(role)) = 'admin'`, alvo); err != nil {
		return 0, err
	}
	return movidos, tx.Commit()
}
