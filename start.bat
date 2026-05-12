@echo off
:: Axis Dev Launcher - starts engine + GUI in one click
:: Usage: double-click or run from terminal

set ROOT=%~dp0
cd /d "%ROOT%"

echo [Axis] Building...
go build -o axis-dev.exe ./cmd/axis/ 2>nul
if errorlevel 1 (
    echo [Axis] Build failed. Run 'go build -o axis-dev.exe ./cmd/axis/' manually.
    pause
    exit /b 1
)

echo [Axis] Starting engine...
start "Axis Engine" /B axis-dev.exe start

:: Wait for runtime.json to appear
:wait_runtime
timeout /t 1 /nobreak >nul
if not exist ".axis\runtime.json" goto wait_runtime

echo [Axis] Engine running.

echo [Axis] Starting GUI on port 3000...
start "Axis GUI" /B tools\axis-gui\axis-gui.exe --port 3000 --root "%ROOT%"

timeout /t 1 /nobreak >nul
echo.
echo ============================================
echo   Axis Engine:  see .axis/runtime.json
echo   Axis GUI:     http://localhost:3000
echo ============================================
echo.
echo Press Ctrl+C or close this window to stop.
pause >nul
