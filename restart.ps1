# PowerShell script to restart FigmaProxy application
# Run this directly on the Windows server or via SSM

Write-Host "Stopping FigmaProxy service..." -ForegroundColor Yellow
& "C:\ProgramData\chocolatey\bin\nssm.exe" stop FigmaProxy -ErrorAction SilentlyContinue
Start-Sleep -Seconds 5

Write-Host "Updating code from repository..." -ForegroundColor Yellow
Set-Location "C:\figma-mcp-proxy"
& "C:\Program Files\Git\cmd\git.exe" fetch
& "C:\Program Files\Git\cmd\git.exe" pull

Write-Host "Rebuilding application..." -ForegroundColor Yellow
& "C:\Program Files\Go\bin\go.exe" build -o figma-proxy.exe main.go

Write-Host "Starting FigmaProxy service..." -ForegroundColor Yellow
& "C:\ProgramData\chocolatey\bin\nssm.exe" start FigmaProxy
Start-Sleep -Seconds 3

$status = & "C:\ProgramData\chocolatey\bin\nssm.exe" status FigmaProxy
Write-Host "Service status: $status" -ForegroundColor Cyan

if ($status -eq "SERVICE_RUNNING") {
    Write-Host "FigmaProxy restarted successfully!" -ForegroundColor Green
} else {
    Write-Host "WARNING: Service may not be running properly" -ForegroundColor Red
    exit 1
}
