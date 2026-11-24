# Sobe a API apontando para o cluster Postgres de desenvolvimento local.
# Nada aqui serve para producao: a DATABASE_URL e o JWT_SECRET sao locais.

$ErrorActionPreference = 'Stop'

# O instalador do Go so altera o PATH de shells abertos DEPOIS da instalacao.
# Num PowerShell que ja estava aberto, 'go' nao e encontrado — entao resolvemos
# o executavel na mao antes de usar.
$go = (Get-Command go -ErrorAction SilentlyContinue).Source
if (-not $go) {
    foreach ($p in @("$env:ProgramFiles\Go\bin\go.exe", "$env:LOCALAPPDATA\Programs\Go\bin\go.exe")) {
        if (Test-Path $p) { $go = $p; break }
    }
}
if (-not $go) {
    Write-Error "go.exe nao encontrado. Instale com: winget install GoLang.Go"
    return
}
Write-Host "Usando $go"

$env:DATABASE_URL = 'postgres://imoveisgo:admin123@127.0.0.1:5440/bretasimoveisdb?sslmode=disable'
$env:JWT_SECRET   = 'dev-only-nao-usar-em-producao'
$env:FRONTEND_URL = 'http://localhost:5173'
$env:PORT         = '8080'

# SMTP nao configurado: /senha/esqueci responde 503.
# Para testar o fluxo completo, defina SMTP_HOST/SMTP_PORT/SMTP_USER/SMTP_PASS.

# Avisa cedo se o banco nao estiver no ar, em vez de deixar a API dar log.Fatal.
$tcp = Test-NetConnection -ComputerName 127.0.0.1 -Port 5440 -InformationLevel Quiet -WarningAction SilentlyContinue
if (-not $tcp) {
    Write-Warning "Nada escutando em 127.0.0.1:5440. Rode start-db.ps1 primeiro."
}

Set-Location $PSScriptRoot
& $go run .
