@echo off
:: Axis Dev Launcher - starts engine + GUI in one click
:: Usage: double-click or run from terminal

set ROOT=%~dp0
cd /d "%ROOT%"

echo [Axis] Building engine...
go build -o axis-dev.exe ./cmd/axis/ 2>nul
if errorlevel 1 (
    echo [Axis] Engine build failed.
    pause
    exit /b 1
)

echo [Axis] Building GUI...
go build -o tools\axis-gui\axis-gui.exe ./tools/axis-gui/ 2>nul
if errorlevel 1 (
    echo [Axis] GUI build failed.
    pause
    exit /b 1
)

:: Kill stale processes
taskkill /F /IM axis-dev.exe >nul 2>&1
taskkill /F /IM axis-gui.exe >nul 2>&1
timeout /t 1 /nobreak >nul

echo [Axis] Starting engine...
start "" /MIN axis-dev.exe start

:: Wait for runtime.json
:wait_runtime
timeout /t 1 /nobreak >nul
if not exist ".axis\runtime.json" goto wait_runtime

echo [Axis] Engine running.
echo [Axis] Starting GUI on port 3000...
start "" /MIN tools\axis-gui\axis-gui.exe --port 3000 --root "%ROOT%"

timeout /t 2 /nobreak >nul
echo.
echo ============================================
echo   Axis Engine:  running (see .axis/runtime.json)
echo   Axis GUI:     http://localhost:3000
echo ============================================
echo.
echo Both processes run in minimized windows.
echo Close them manually or run: taskkill /F /IM axis-dev.exe ^& taskkill /F /IM axis-gui.exe
