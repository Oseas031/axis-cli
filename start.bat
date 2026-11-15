@echo off
:: Axis Dev Launcher - single-binary runtime + built-in UI
:: Usage: double-click or run from terminal

set ROOT=%~dp0
cd /d "%ROOT%"

echo [Axis] Building single-binary engine + UI...
go build -o axis-dev.exe ./cmd/axis/
if errorlevel 1 (
    echo [Axis] Build failed.
    pause
    exit /b 1
)

taskkill /F /IM axis-dev.exe >nul 2>&1
timeout /t 1 /nobreak >nul

echo [Axis] Starting runtime and opening Observatory UI...
axis-dev.exe gui
