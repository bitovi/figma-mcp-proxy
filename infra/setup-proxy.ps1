param(
    [string]$ApiKey,
    [string]$ExternalDnsName
)

Set-ExecutionPolicy Bypass -Scope Process -Force
[System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072

# Install dependencies
if (!(Test-Path 'C:\ProgramData\chocolatey\bin\choco.exe')) { 
    iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))
}
$env:Path = [System.Environment]::GetEnvironmentVariable('Path','Machine') + ';' + [System.Environment]::GetEnvironmentVariable('Path','User')
& 'C:\ProgramData\chocolatey\bin\choco.exe' install git golang -y --force
$env:Path = [System.Environment]::GetEnvironmentVariable('Path','Machine') + ';' + [System.Environment]::GetEnvironmentVariable('Path','User')

# Firewall
New-NetFirewallRule -DisplayName "Allow App Port 3846" -Direction Inbound -Action Allow -Protocol TCP -LocalPort 3846 -ErrorAction SilentlyContinue

# Remove old task
Unregister-ScheduledTask -TaskName 'FigmaProxy' -Confirm:$false -ErrorAction SilentlyContinue
Start-Sleep -Seconds 2

# Clone and build
cd C:\
if (Test-Path 'C:\figma-mcp-proxy') { 
    Remove-Item -Recurse -Force 'C:\figma-mcp-proxy' -ErrorAction SilentlyContinue 
}
& 'C:\Program Files\Git\cmd\git.exe' clone https://github.com/bitovi/figma-mcp-proxy.git
cd C:\figma-mcp-proxy
& 'C:\Program Files\Go\bin\go.exe' build -o figma-proxy.exe main.go

# Create logs directory
New-Item -ItemType Directory -Force -Path 'C:\figma-mcp-proxy\logs' | Out-Null

# Register scheduled task
$action = New-ScheduledTaskAction `
    -Execute 'powershell.exe' `
    -Argument "-NoProfile -ExecutionPolicy Bypass -Command `"Set-Location 'C:\figma-mcp-proxy'; `$env:API_KEY='$ApiKey'; `$env:EXTERNAL_DNS_NAME='$ExternalDnsName'; `$env:FIGMA_OPEN_DELAY_SECONDS='15'; .\figma-proxy.exe *>> C:\figma-mcp-proxy\logs\proxy.log 2>&1`"" `
    -WorkingDirectory 'C:\figma-mcp-proxy'

$trigger = New-ScheduledTaskTrigger -AtStartup
$settings = New-ScheduledTaskSettingsSet -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries -ExecutionTimeLimit (New-TimeSpan -Days 0)
$principal = New-ScheduledTaskPrincipal -UserId 'Administrator' -LogonType Interactive -RunLevel Highest

Register-ScheduledTask -TaskName 'FigmaProxy' -Action $action -Trigger $trigger -Settings $settings -Principal $principal -ErrorAction Stop | Out-Null
Start-ScheduledTask -TaskName 'FigmaProxy' -ErrorAction Stop

Write-Host "FigmaProxy scheduled task registered and started successfully"
