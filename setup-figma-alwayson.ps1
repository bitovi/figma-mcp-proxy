# ============================================================================
# Figma Always-On Setup Script
# ============================================================================

# Configuration
$FigmaPath = 'C:\Users\Administrator\AppData\Local\Figma\app-125.11.6\Figma.exe'
$FigmaWorkDir = 'C:\figma'
$WatchdogScript = "$FigmaWorkDir\watchdog.ps1"
$AdminPassword = Read-Host "Enter Administrator password" -AsSecureString
$AdminPasswordPlain = [System.Runtime.InteropServices.Marshal]::PtrToStringAuto([System.Runtime.InteropServices.Marshal]::SecureStringToBSTR($AdminPassword))

Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "Figma Always-On Setup" -ForegroundColor Cyan
Write-Host "========================================`n" -ForegroundColor Cyan

# ============================================================================
# Step 1: Setup Auto-Login
# ============================================================================
Write-Host "[1/5] Configuring auto-login..." -ForegroundColor Yellow

$RegPath = "HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Winlogon"
Set-ItemProperty -Path $RegPath -Name "AutoAdminLogon" -Value "1"
Set-ItemProperty -Path $RegPath -Name "DefaultUsername" -Value "Administrator"
Set-ItemProperty -Path $RegPath -Name "DefaultPassword" -Value $AdminPasswordPlain
Set-ItemProperty -Path $RegPath -Name "AutoLogonCount" -Value "999999"

Write-Host "Auto-login configured" -ForegroundColor Green

# ============================================================================
# Step 2: Disable Lock Screen and Sleep
# ============================================================================
Write-Host "`n[2/5] Disabling lock screen and sleep..." -ForegroundColor Yellow

# Disable lock screen
$PersonalizationPath = "HKLM:\SOFTWARE\Policies\Microsoft\Windows\Personalization"
if (-not (Test-Path $PersonalizationPath)) {
    New-Item -Path $PersonalizationPath -Force | Out-Null
}
Set-ItemProperty -Path $PersonalizationPath -Name "NoLockScreen" -Value 1 -Force

# Prevent sleep
powercfg /change standby-timeout-ac 0
powercfg /change standby-timeout-dc 0
powercfg /change monitor-timeout-ac 0
powercfg /change hibernate-timeout-ac 0

# Disable screen saver
$ScreenSaverPath = "HKCU:\Control Panel\Desktop"
Set-ItemProperty -Path $ScreenSaverPath -Name "ScreenSaveActive" -Value "0" -Force
Set-ItemProperty -Path $ScreenSaverPath -Name "ScreenSaveTimeOut" -Value "0" -Force

# Disable password on wake
powercfg /setacvalueindex SCHEME_CURRENT SUB_NONE CONSOLELOCK 0
powercfg /setactive SCHEME_CURRENT

Write-Host "Lock screen and sleep disabled" -ForegroundColor Green

# ============================================================================
# Step 3: Setup Auto-Start Figma on Login
# ============================================================================
Write-Host "`n[3/5] Setting up Figma auto-start..." -ForegroundColor Yellow

# Remove existing task if it exists
Unregister-ScheduledTask -TaskName "FigmaAlwaysOn" -Confirm:$false -ErrorAction SilentlyContinue

# Create scheduled task to start Figma at login
$action = New-ScheduledTaskAction -Execute $FigmaPath
$trigger = New-ScheduledTaskTrigger -AtLogOn -User 'Administrator'
$settings = New-ScheduledTaskSettingsSet -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries -RestartCount 3 -RestartInterval (New-TimeSpan -Minutes 1)
$principal = New-ScheduledTaskPrincipal -UserId 'Administrator' -LogonType Interactive -RunLevel Highest

Register-ScheduledTask -TaskName "FigmaAlwaysOn" -Action $action -Trigger $trigger -Settings $settings -Principal $principal | Out-Null

Write-Host "Figma auto-start configured" -ForegroundColor Green

# ============================================================================
# Step 4: Create Watchdog Script
# ============================================================================
Write-Host "`n[4/5] Creating watchdog script..." -ForegroundColor Yellow

# Create figma directory if it doesn't exist
New-Item -ItemType Directory -Force -Path $FigmaWorkDir | Out-Null
New-Item -ItemType Directory -Force -Path "$FigmaWorkDir\logs" | Out-Null

# Create watchdog script
$watchdogContent = @'
# Figma Watchdog Script
# Monitors Figma process and restarts if not running

$logFile = "C:\figma\logs\watchdog.log"
$figmaPath = 'C:\Users\Administrator\AppData\Local\Figma\app-125.11.6\Figma.exe'

function Write-Log {
    param([string]$message)
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    $logMessage = "$timestamp - $message"
    Add-Content -Path $logFile -Value $logMessage
    Write-Host $logMessage
}

Write-Log "Watchdog started"

while ($true) {
    $figma = Get-Process -Name "Figma" -ErrorAction SilentlyContinue
    
    if (-not $figma) {
        Write-Log "Figma not running, starting..."
        try {
            Start-Process $figmaPath -ErrorAction Stop
            Write-Log "Figma started successfully"
        } catch {
            Write-Log "ERROR: Failed to start Figma - $($_.Exception.Message)"
        }
    }
    
    Start-Sleep -Seconds 30
}
'@

Set-Content -Path $WatchdogScript -Value $watchdogContent -Force

Write-Host "Watchdog script created at: $WatchdogScript" -ForegroundColor Green

# ============================================================================
# Step 5: Schedule Watchdog
# ============================================================================
Write-Host "`n[5/5] Scheduling watchdog..." -ForegroundColor Yellow

# Remove existing task if it exists
Unregister-ScheduledTask -TaskName "FigmaWatchdog" -Confirm:$false -ErrorAction SilentlyContinue

# Create scheduled task for watchdog
$action = New-ScheduledTaskAction -Execute 'powershell.exe' -Argument "-WindowStyle Hidden -ExecutionPolicy Bypass -File `"$WatchdogScript`""
$trigger = New-ScheduledTaskTrigger -AtStartup
$settings = New-ScheduledTaskSettingsSet -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries -ExecutionTimeLimit (New-TimeSpan -Days 0)
$principal = New-ScheduledTaskPrincipal -UserId 'Administrator' -LogonType Interactive -RunLevel Highest

Register-ScheduledTask -TaskName "FigmaWatchdog" -Action $action -Trigger $trigger -Settings $settings -Principal $principal | Out-Null

Write-Host "Watchdog scheduled" -ForegroundColor Green

# ============================================================================
# Summary
# ============================================================================
Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "Setup Complete!" -ForegroundColor Green
Write-Host "========================================`n" -ForegroundColor Cyan

Write-Host "Next steps:" -ForegroundColor Yellow
Write-Host "1. Reboot the system (auto-login will occur)"
Write-Host "2. Log into Figma via browser"
Write-Host "3. Start the MCP server in Figma"
Write-Host "4. Figma will now restart automatically if closed"
Write-Host ""
Write-Host "Watchdog logs: $FigmaWorkDir\logs\watchdog.log" -ForegroundColor Cyan
Write-Host ""
Write-Host "To manually start watchdog now: Start-ScheduledTask -TaskName 'FigmaWatchdog'" -ForegroundColor Gray

# Clear password from memory
$AdminPasswordPlain = $null
[System.GC]::Collect()