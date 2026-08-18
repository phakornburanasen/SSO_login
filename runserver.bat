@echo off
chcp 65001 >nul
setlocal

REM ============================================================
REM  SSO Login - Run All Services (Windows)
REM  Starts:
REM    - api_gatewayGo  : http://localhost:18000
REM    - backend        : http://localhost:12010
REM    - frontend (dev) : http://localhost:3000
REM ============================================================

set ROOT=%~dp0
set GATEWAY_DIR=%ROOT%api_gatewayGo
set BACKEND_DIR=%ROOT%backend
set FRONTEND_DIR=%ROOT%frontend

set GATEWAY_BIN=%GATEWAY_DIR%\bin\sso-login-gateway.exe
set BACKEND_BIN=%BACKEND_DIR%\bin\sso-login-server.exe

echo.
echo ============================================================
echo   SSO Login - Starting all services
echo   - api_gatewayGo  : http://localhost:18000
echo   - backend        : http://localhost:12010
echo   - frontend (dev) : http://localhost:3000
echo ============================================================
echo.

REM ---- Build gateway if missing ----
if not exist "%GATEWAY_BIN%" (
    echo [INFO] Building api_gatewayGo ...
    pushd "%GATEWAY_DIR%"
    call go build -o bin\sso-login-gateway.exe .\cmd\server
    if errorlevel 1 (
        echo [ERROR] api_gatewayGo build failed.
        popd
        pause
        exit /b 1
    )
    popd
    echo [OK] api_gatewayGo built successfully.
)

REM ---- Build backend if missing ----
if not exist "%BACKEND_BIN%" (
    echo [INFO] Building backend ...
    pushd "%BACKEND_DIR%"
    call go build -o bin\sso-login-server.exe .\cmd\server
    if errorlevel 1 (
        echo [ERROR] backend build failed.
        popd
        pause
        exit /b 1
    )
    popd
    echo [OK] backend built successfully.
)

REM ---- Start api_gatewayGo (port 18000) ----
echo [INFO] Starting api_gatewayGo on :18000 ...
start "SSO api_gatewayGo (18000)" cmd /k "cd /d %GATEWAY_DIR% && bin\sso-login-gateway.exe"

REM ---- Start backend (port 12010) ----
echo [INFO] Starting backend on :12010 ...
start "SSO Backend (12010)" cmd /k "cd /d %BACKEND_DIR% && bin\sso-login-server.exe"

REM ---- Start frontend dev server (port 3000) ----
if exist "%FRONTEND_DIR%\package.json" (
    echo [INFO] Starting frontend on :3000 ...
    start "SSO Frontend (3000)" cmd /k "cd /d %FRONTEND_DIR% && npm run dev"
) else (
    echo [WARN] Frontend directory not found, skipping.
)

echo.
echo ============================================================
echo   All services started in separate windows.
echo   Close the corresponding window to stop each service.
echo ============================================================
echo.
pause
endlocal
