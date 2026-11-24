-- Executar conectado ao banco bretasimoveisdb.
-- Esquema derivado de model/models.go e das queries em repository/.
--
-- A tabela password_resets NÃO está aqui de propósito: a própria API a cria no
-- boot, em repository.EnsurePasswordResetTable().

CREATE TABLE IF NOT EXISTS pessoas (
	id    SERIAL PRIMARY KEY,
	nome  TEXT NOT NULL,
	email TEXT NOT NULL UNIQUE,
	senha TEXT NOT NULL,
	role  TEXT NOT NULL DEFAULT 'usuario'
);

-- E-mail é consultado com LOWER(email) em FindByEmail; o índice abaixo permite
-- que essa busca use índice em vez de varrer a tabela.
CREATE UNIQUE INDEX IF NOT EXISTS pessoas_email_lower_idx ON pessoas (LOWER(email));

CREATE TABLE IF NOT EXISTS imoveis (
	id          SERIAL PRIMARY KEY,
	tipo        TEXT    NOT NULL,
	rua         TEXT    NOT NULL DEFAULT '',
	numero      TEXT    NOT NULL DEFAULT '',
	bairro      TEXT    NOT NULL DEFAULT '',
	cidade      TEXT    NOT NULL DEFAULT '',
	estado      TEXT    NOT NULL DEFAULT '',
	cep         TEXT    NOT NULL DEFAULT '',
	pais        TEXT    NOT NULL DEFAULT '',
	area        INTEGER NOT NULL DEFAULT 0,
	quartos     INTEGER NOT NULL DEFAULT 0,
	banheiros   INTEGER NOT NULL DEFAULT 0,
	suites      INTEGER NOT NULL DEFAULT 0,
	vagas       INTEGER NOT NULL DEFAULT 0,
	andar       INTEGER NOT NULL DEFAULT 0,
	valor       INTEGER NOT NULL DEFAULT 0,
	situacao    TEXT    NOT NULL DEFAULT '',
	disponivel  BOOLEAN NOT NULL DEFAULT TRUE,
	descricao   TEXT    NOT NULL DEFAULT '',
	-- imagem: bytes de cada arquivo enviado em multipart (pq.Array -> bytea[]).
	-- O recorte imagem[1:1] da listagem (filter.Thumb) exige um array de verdade.
	-- Aceita NULL de proposito: InsertImovelRepository grava NULL quando nenhuma
	-- imagem e enviada, e o Scan da listagem converte NULL em slice vazio.
	imagem      BYTEA[],
	imagem_type TEXT[],
	id_pessoa   INTEGER NOT NULL REFERENCES pessoas(id) ON DELETE CASCADE
);

-- Colunas mais usadas em buildImovelWhere.
CREATE INDEX IF NOT EXISTS imoveis_filtro_idx
	ON imoveis (pais, situacao, tipo, estado, cidade, valor);
