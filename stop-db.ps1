# Para o cluster Postgres de desenvolvimento (porta 5440).
# NAO toca no servico PostgreSQL do sistema, que roda na 5433.

$ErrorActionPreference = 'Stop'

$bin     = 'C:\Program Files\PostgreSQL\18\bin'
$dataDir = 'C:\Users\user\pgdata_imoveis'

$up = Test-NetConnection -ComputerName 127.0.0.1 -Port 5440 -InformationLevel Quiet -WarningAction SilentlyContinue
if (-not $up) {
    Write-Host "Cluster de desenvolvimento ja esta parado."
    return
}

# -m fast encerra as conexoes abertas e faz checkpoint, sem esperar os clientes
# desconectarem sozinhos (que e o comportamento do modo smart, o padrao).
& "$bin\pg_ctl.exe" -D $dataDir -m fast -w stop

Write-Host "Cluster de desenvolvimento parado."
