# Rebuild CCTokenViz and start it again: stops the running copy (a running exe blocks the
# build), builds build\bin\CCTokenViz.exe, and launches it. Pin that exe to the taskbar and
# every relaunch picks up the newest build. Run from anywhere: powershell scripts\relaunch.ps1
$root = Split-Path -Parent $PSScriptRoot
$exe = Join-Path $root "build\bin\CCTokenViz.exe"

Get-Process CCTokenViz -ErrorAction SilentlyContinue | Stop-Process -Force
Start-Sleep -Milliseconds 300

Push-Location $root
try {
    wails build
    if ($LASTEXITCODE -ne 0) { Write-Error "Build failed"; exit 1 }
    git checkout -- frontend/wailsjs/runtime 2>$null  # wails rewrites these on every build; keep the tree clean
} finally {
    Pop-Location
}

Start-Process $exe
Write-Host "Started $exe"
