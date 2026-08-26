# Stop Caddy reverse proxy

$proc = Get-Process caddy -ErrorAction SilentlyContinue
if (!$proc) { Write-Host "Caddy is not running"; exit 0 }

Stop-Process -Name caddy -Force
Start-Sleep -Seconds 1
$check = Get-Process caddy -ErrorAction SilentlyContinue
if ($check) { Write-Warning "Caddy may still be running (PID $($check.Id))" }
else { Write-Host "Caddy stopped" }
