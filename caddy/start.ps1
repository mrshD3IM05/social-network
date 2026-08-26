# Start Caddy reverse proxy
# Frontend: Next.js on 127.0.0.1:3000
# Backend : Go API on 127.0.0.1:8080

$ErrorActionPreference = "Stop"

$caddy = "C:\Users\ELECTRO MARSLI\Desktop\git\social-network\caddy\caddy.exe"
$config = "C:\Users\ELECTRO MARSLI\Desktop\git\social-network\caddy\Caddyfile"

if (!(Test-Path $caddy)) { Write-Error "caddy.exe not found at $caddy"; exit 1 }

$existing = Get-Process caddy -ErrorAction SilentlyContinue
if ($existing) { Write-Host "Caddy is already running (PID $($existing.Id))"; exit 0 }

Start-Process -FilePath $caddy -ArgumentList "run", "--config", $config -WindowStyle Hidden
Start-Sleep -Seconds 1
$proc = Get-Process caddy -ErrorAction SilentlyContinue
if ($proc) { Write-Host "Caddy started (PID $($proc.Id))" }
else { Write-Error "Caddy failed to start" }
