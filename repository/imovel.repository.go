package repository

import (
	"apiGo/config"
	"apiGo/model"
	"fmt"
	"log"
	"strings"

	"github.com/lib/pq"
)

func InsertImovelRepository(im model.Imovel) (int, error) {
	query := `
	INSERT INTO imoveis (
		tipo, rua, numero, bairro, cidade, estado, cep, pais,
		area, quartos, banheiros, suites, vagas, andar,
		valor, situacao, disponivel, descricao, imagem, imagem_type, id_pessoa
	) VALUES (
		$1,$2,$3,$4,$5,$6,$7,$8,
		$9,$10,$11,$12,$13,$14,
		$15,$16,$17,$18,$19,$20,$21
	) RETURNING id;
	`

	var lastId int
	err := config.DB.QueryRow(query,
		im.Tipo, im.Rua, im.Numero, im.Bairro, im.Cidade, im.Estado, im.Cep,
		im.Pais, im.Area, im.Quartos, im.Banheiros, im.Suites, im.Vagas, im.Andar,
		im.Valor, im.Situacao, im.Disponivel, im.Descricao, pq.Array(im.Imagem), pq.Array(im.ImagemType), im.IdPessoa,
	).Scan(&lastId)

	if err != nil {
		log.Println("Erro ao inserir imóvel:", err)
		return 0, err
	}

	return lastId, nil
}

func FilterImovelRepository(filter model.Filtro) ([]model.Imovel, error) {
	args := []interface{}{}
	query := `SELECT id, tipo, rua, numero, bairro, cidade, estado, cep, pais,
                     area, quartos, banheiros, vagas, andar,
                     valor, situacao, disponivel, descricao, imagem, imagem_type, id_pessoa
              FROM imoveis WHERE 1=1`

	paramIndex := 1

	if filter.Id != 0 {
		query += fmt.Sprintf(" AND id = $%d", paramIndex)
		args = append(args, filter.Id)
		paramIndex++
	}
	if filter.Situacao != "" {
		query += fmt.Sprintf(" AND situacao = $%d", paramIndex)
		args = append(args, filter.Situacao)
		paramIndex++
	}
	if filter.Tipo != "" {
		query += fmt.Sprintf(" AND tipo = $%d", paramIndex)
		args = append(args, filter.Tipo)
		paramIndex++
	}
	if filter.Estado != "" {
		query += fmt.Sprintf(" AND estado = $%d", paramIndex)
		args = append(args, filter.Estado)
		paramIndex++
	}
	if filter.Cidade != "" {
		query += fmt.Sprintf(" AND cidade = $%d", paramIndex)
		args = append(args, filter.Cidade)
		paramIndex++
	}
	if filter.De > 0 && filter.Ate > 0 {
		query += fmt.Sprintf(" AND valor BETWEEN $%d AND $%d", paramIndex, paramIndex+1)
		args = append(args, filter.De, filter.Ate)
		paramIndex += 2
	}
	if filter.Quartos != 0 {
		query += fmt.Sprintf(" AND quartos = $%d", paramIndex)
		args = append(args, filter.Quartos)
		paramIndex++
	}
	if filter.Vagas != 0 {
		query += fmt.Sprintf(" AND vagas = $%d", paramIndex)
		args = append(args, filter.Vagas)
		paramIndex++
	}
	if filter.Banheiros != 0 {
		query += fmt.Sprintf(" AND banheiros = $%d", paramIndex)
		args = append(args, filter.Banheiros)
		paramIndex++
	}

	rows, err := config.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	imoveis := []model.Imovel{}

	for rows.Next() {
		var imovel model.Imovel
		err := rows.Scan(
			&imovel.Id,
			&imovel.Tipo,
			&imovel.Rua,
			&imovel.Numero,
			&imovel.Bairro,
			&imovel.Cidade,
			&imovel.Estado,
			&imovel.Cep,
			&imovel.Pais,
			&imovel.Area,
			&imovel.Quartos,
			&imovel.Banheiros,
			&imovel.Vagas,
			&imovel.Andar,
			&imovel.Valor,
			&imovel.Situacao,
			&imovel.Disponivel,
			&imovel.Descricao,
			pq.Array(&imovel.Imagem),
			pq.Array(&imovel.ImagemType),
			&imovel.IdPessoa,
		)
		if err != nil {
			return nil, err
		}
		
		if imovel.ImagemType == nil {
			imovel.ImagemType = []string{}
		}
		if imovel.Imagem == nil {
			imovel.Imagem = [][]byte{}
		}
		imoveis = append(imoveis, imovel)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return imoveis, nil
}

func DeleteImovelRepository(id int) (int, error) {
	result, err := config.DB.Exec("DELETE FROM imoveis WHERE id = $1", id)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return int(rowsAffected), nil
}

func UpdateImovelRepository(imovel model.AtualizarImovel) (int, error) {
	fields := []string{}
	args := []interface{}{}
	paramIndex := 1

	if imovel.Estado != "" {
		fields = append(fields, fmt.Sprintf("estado = $%d", paramIndex))
		args = append(args, imovel.Estado)
		paramIndex++
	}
	if imovel.Cidade != "" {
		fields = append(fields, fmt.Sprintf("cidade = $%d", paramIndex))
		args = append(args, imovel.Cidade)
		paramIndex++
	}
	if imovel.Bairro != "" {
		fields = append(fields, fmt.Sprintf("bairro = $%d", paramIndex))
		args = append(args, imovel.Bairro)
		paramIndex++
	}
	if imovel.Rua != ""{
		fields = append(fields, fmt.Sprintf("rua = $%d", paramIndex))
		args = append(args, imovel.Rua)
		paramIndex++
	}

	if imovel.Pais != ""{
		fields = append(fields, fmt.Sprintf("pais = $%d", paramIndex))
		args = append(args, imovel.Pais)
		paramIndex++
	}

	if imovel.Situacao != "" {
		fields = append(fields, fmt.Sprintf("situacao = $%d", paramIndex))
		args = append(args, imovel.Situacao)
		paramIndex++
	}

	if imovel.Cep != ""{
		fields = append(fields, fmt.Sprintf("cep = $%d", paramIndex))
		args = append(args, imovel.Cep)
		paramIndex++
	}

	if imovel.Numero != "" {
		fields = append(fields, fmt.Sprintf("numero = $%d", paramIndex))
		args = append(args, imovel.Numero)
		paramIndex++
	}
	if imovel.Vagas != 0 {
		fields = append(fields, fmt.Sprintf("vagas = $%d", paramIndex))
		args = append(args, imovel.Vagas)
		paramIndex++
	}

	if imovel.Tipo != "" {
		fields = append(fields, fmt.Sprintf("tipo = $%d", paramIndex))
		args = append(args, imovel.Tipo)
		paramIndex++
	}
	if imovel.Valor != 0 {
		fields = append(fields, fmt.Sprintf("valor = $%d", paramIndex))
		args = append(args, imovel.Valor)
		paramIndex++
	}
	if imovel.Quartos != 0 {
		fields = append(fields, fmt.Sprintf("quartos = $%d", paramIndex))
		args = append(args, imovel.Quartos)
		paramIndex++
	}
	if imovel.Banheiros != 0 {
		fields = append(fields, fmt.Sprintf("banheiros = $%d", paramIndex))
		args = append(args, imovel.Banheiros)
		paramIndex++
	}
	if imovel.Area != 0 {
		fields = append(fields, fmt.Sprintf("area = $%d", paramIndex))
		args = append(args, imovel.Area)
		paramIndex++
	}
	if imovel.Descricao != "" {
		fields = append(fields, fmt.Sprintf("descricao = $%d", paramIndex))
		args = append(args, imovel.Descricao)
		paramIndex++
	}
	if len(imovel.Imagem) > 0 {
		fields = append(fields, fmt.Sprintf("imagem = $%d", paramIndex))
		args = append(args, pq.Array(imovel.Imagem))
		paramIndex++
	}
	if len(imovel.ImagemType) > 0 {
		fields = append(fields, fmt.Sprintf("imagem_type = $%d", paramIndex))
		args = append(args, pq.Array(imovel.ImagemType))
		paramIndex++
	}
	if len(fields) == 0 {
		log.Printf("⚠️ Nenhum campo foi enviado para atualização do imóvel ID %d\n", imovel.IdImovel)
		return 0, fmt.Errorf("nenhum campo para atualizar")
	}

	query := fmt.Sprintf("UPDATE imoveis SET %s WHERE id = $%d", strings.Join(fields, ", "), paramIndex)
	args = append(args, imovel.IdImovel)

	result, err := config.DB.Exec(query, args...)
	if err != nil {
		return 0, err
	}

	rowsAffected, _ := result.RowsAffected()
	return int(rowsAffected), nil
}
