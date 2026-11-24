# Sobe o cluster Postgres de desenvolvimento (separado do PostgreSQL 18 do
# sistema, que continua na porta 5433 e nao e tocado por este script).

$ErrorActionPreference = 'Stop'

$bin     = 'C:\Program Files\PostgreSQL\18\bin'
$dataDir = 'C:\Users\user\pgdata_imoveis'

# Se ja esta escutando, nao ha o que fazer. Chamar pg_ctl start de novo so
# gera "outro servidor pode estar executando" e falha ao abrir o log.
$up = Test-NetConnection -ComputerName 127.0.0.1 -Port 5440 -InformationLevel Quiet -WarningAction SilentlyContinue
if ($up) {
    Write-Host "Cluster de desenvolvimento ja esta no ar em 127.0.0.1:5440."
    return
}

# Log com nome unico por execucao: se outro processo ainda segurar o arquivo
# anterior, o start nao falha com "Permission denied".
$log = Join-Path $dataDir ("server-{0}.log" -f (Get-Date -Format 'yyyyMMdd-HHmmss'))

& "$bin\pg_ctl.exe" -D $dataDir -l $log -w start

Write-Host "Log: $log"
