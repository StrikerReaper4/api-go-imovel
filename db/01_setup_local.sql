-- Executar como superusuário no cluster de desenvolvimento.
-- Cria o papel e o banco usados localmente pela API.

DO $$
BEGIN
	IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'imoveisgo') THEN
		CREATE ROLE imoveisgo LOGIN PASSWORD 'admin123';
	ELSE
		ALTER ROLE imoveisgo LOGIN PASSWORD 'admin123';
	END IF;
END
$$;

SELECT 'CREATE DATABASE bretasimoveisdb OWNER imoveisgo'
WHERE NOT EXISTS (SELECT 1 FROM pg_database WHERE datname = 'bretasimoveisdb')
\gexec
