# Para tudo que o ambiente local sobe: API (8080), front Vite (5173) e o
# cluster Postgres de desenvolvimento (5440).

$ErrorActionPreference = 'Continue'

# Mata quem estiver escutando na porta, em vez de matar todo processo com o
# nome "node" ou "go" — assim outros projetos abertos nao sao derrubados junto.
function Stop-Porta {
    param([int]$Porta, [string]$Nome)

    $conns = Get-NetTCPConnection -State Listen -LocalPort $Porta -ErrorAction SilentlyContinue
    if (-not $conns) {
        Write-Host "$Nome (porta $Porta): ja parado."
        return
    }

    foreach ($procId in ($conns.OwningProcess | Sort-Object -Unique)) {
        try {
            $proc = Get-Process -Id $procId -ErrorAction Stop
            Stop-Process -Id $procId -Force -ErrorAction Stop
            Write-Host "$Nome (porta $Porta): parado - $($proc.ProcessName) pid $procId."
        } catch {
            Write-Warning "$Nome (porta $Porta): falha ao parar pid $procId - $_"
        }
    }
}

Stop-Porta -Porta 8080 -Nome 'API Go'
Stop-Porta -Porta 5173 -Nome 'Front Vite'

# O banco tem desligamento proprio (checkpoint); nao deve ser morto a forca.
& "$PSScriptRoot\stop-db.ps1"
