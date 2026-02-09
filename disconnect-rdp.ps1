# Disconnect from RDP while keeping session active
# Run this script instead of closing the RDP window

Write-Host "Finding current RDP session..." -ForegroundColor Yellow

try {
    # Get current session ID
    $sessions = qwinsta
    $currentSession = $sessions | Where-Object { $_ -match 'Active' -and $_ -match 'rdp' }
    
    if ($currentSession) {
        # Extract session ID (typically column 3)
        $sessionId = ($currentSession -split '\s+')[2]
        
        Write-Host "Current session ID: $sessionId" -ForegroundColor Cyan
        Write-Host "Transferring session to console..." -ForegroundColor Yellow
        
        # Transfer to console session
        tscon $sessionId /dest:console
        
        Write-Host "Session transferred successfully!" -ForegroundColor Green
        Write-Host "You will be disconnected, but the session remains active." -ForegroundColor Cyan
    } else {
        Write-Host "No active RDP session found. Are you connected via RDP?" -ForegroundColor Red
        exit 1
    }
} catch {
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "`nFalling back to manual method:" -ForegroundColor Yellow
    Write-Host "1. Run: query session" -ForegroundColor Gray
    Write-Host "2. Find your session ID" -ForegroundColor Gray
    Write-Host "3. Run: tscon <ID> /dest:console" -ForegroundColor Gray
    exit 1
}
