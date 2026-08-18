@echo off
chcp 65001 >nul
setlocal

REM ============================================================
REM  SSO Login - Stop All Services
REM  Kills:
REM    - api_gatewayGo (18000)
REM    - backend (12010)
REM    - frontend dev server (3000)
REM ============================================================

echo.
echo [INFO] Stopping SSO services...

for /f "tokens=5" %%P in ('netstat -aon ^| findstr ":18000" ^| findstr "LISTENING"') do (
    echo [INFO] Killing api_gatewayGo PID %%P ...
    taskkill /F /PID %%P >nul 2>&1
)

for /f "tokens=5" %%P in ('netstat -aon ^| findstr ":12010" ^| findstr "LISTENING"') do (
    echo [INFO] Killing backend PID %%P ...
    taskkill /F /PID %%P >nul 2>&1
)

for /f "tokens=5" %%P in ('netstat -aon ^| findstr ":3000"  ^| findstr "LISTENING"') do (
    echo [INFO] Killing frontend PID %%P ...
    taskkill /F /PID %%P >nul 2>&1
)

REM Also kill by process name (catches any stragglers)
taskkill /F /IM sso-login-gateway.exe >nul 2>&1
taskkill /F /IM sso-login-server.exe  >nul 2>&1
taskkill /F /IM node.exe               >nul 2>&1

echo [OK] All SSO services stopped.
echo.
pause
endlocal
