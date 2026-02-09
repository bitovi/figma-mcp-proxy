# PowerShell script to restart FigmaProxy application
# Run this directly on the Windows server or via SSM

Write-Host "Stopping FigmaProxy scheduled task..." -ForegroundColor Yellow
Stop-ScheduledTask -TaskName 'FigmaProxy' -ErrorAction SilentlyContinue
Start-Sleep -Seconds 3

Write-Host "Updating code from repository..." -ForegroundColor Yellow
Set-Location "C:\figma-mcp-proxy"
& "C:\Program Files\Git\cmd\git.exe" fetch
& "C:\Program Files\Git\cmd\git.exe" pull

Write-Host "Rebuilding application..." -ForegroundColor Yellow
& "C:\Program Files\Go\bin\go.exe" build -o figma-proxy.exe main.go

Write-Host "Starting FigmaProxy scheduled task..." -ForegroundColor Yellow
Start-ScheduledTask -TaskName 'FigmaProxy' -ErrorAction Stop
Start-Sleep -Seconds 3

$task = Get-ScheduledTask -TaskName 'FigmaProxy'
$taskInfo = Get-ScheduledTaskInfo -TaskName 'FigmaProxy'
Write-Host "Task State: $($task.State)" -ForegroundColor Cyan
Write-Host "Last Run Time: $($taskInfo.LastRunTime)" -ForegroundColor Cyan
Write-Host "Last Task Result: $($taskInfo.LastTaskResult)" -ForegroundColor Cyan

if ($task.State -eq 'Running') {
    Write-Host "FigmaProxy restarted successfully!" -ForegroundColor Green
} else {
    Write-Host "WARNING: Task may not be running properly" -ForegroundColor Red
    Write-Host "You can check logs at: C:\figma-mcp-proxy\logs\proxy.log" -ForegroundColor Yellow
    exit 1
}
