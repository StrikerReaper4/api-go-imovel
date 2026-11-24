-- Dados de desenvolvimento local. NÃO usar em produção.
-- Admin de teste: admin@local.test / admin123 (hash bcrypt abaixo).
-- Role 'admin' é obrigatória: AuthMiddleware recusa qualquer outra.

INSERT INTO pessoas (nome, email, senha, role)
VALUES ('Admin Local', 'admin@local.test',
        '$2a$10$IVsfuIZ5rSvXOY/q96bIpOya68U5R85rqHfSHvfX6sCfKXXp3ZXSC', 'admin')
ON CONFLICT (email) DO UPDATE SET senha = EXCLUDED.senha, role = EXCLUDED.role;

-- Imóveis de exemplo (sem imagens) só para a home e os filtros terem conteúdo.
INSERT INTO imoveis (tipo, rua, numero, bairro, cidade, estado, cep, pais,
                     area, quartos, banheiros, suites, vagas, andar,
                     valor, situacao, disponivel, descricao, id_pessoa)
SELECT v.tipo, v.rua, v.numero, v.bairro, v.cidade, v.estado, v.cep, v.pais,
       v.area, v.quartos, v.banheiros, v.suites, v.vagas, v.andar,
       v.valor, v.situacao, TRUE, v.descricao, p.id
FROM (VALUES
	('Casa',        'Rua das Acácias', '120', 'Centro',             'Campo Grande', 'MS', '79002-363', 'Brasil', 180, 3, 2, 1, 2, 0, 650000, 'venda', 'Casa térrea reformada, quintal amplo.'),
	('Apartamento', 'Av. Afonso Pena', '850', 'Jardim dos Estados', 'Campo Grande', 'MS', '79020-000', 'Brasil',  95, 2, 2, 1, 1, 7, 420000, 'venda', 'Apartamento com vista livre e lazer completo.'),
	('Terreno',     'Rua Projetada',   '0',   'Tiradentes',         'Campo Grande', 'MS', '79040-000', 'Brasil', 400, 0, 0, 0, 0, 0, 180000, 'venda', 'Terreno plano, pronto para construir.')
) AS v(tipo, rua, numero, bairro, cidade, estado, cep, pais,
       area, quartos, banheiros, suites, vagas, andar, valor, situacao, descricao)
CROSS JOIN (SELECT id FROM pessoas WHERE email = 'admin@local.test') p
WHERE NOT EXISTS (SELECT 1 FROM imoveis);
