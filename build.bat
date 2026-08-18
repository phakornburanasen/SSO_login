@echo off
chcp 65001 >nul
setlocal

REM ============================================================
REM  SSO Login - Build for Linux (cross-compile from Windows)
REM  Output:
REM    - backend/bin/sso-login-server       (linux/amd64)
REM    - api_gatewayGo/bin/sso-login-gateway (linux/amd64)
REM ============================================================

set ROOT=%~dp0
set GATEWAY_DIR=%ROOT%api_gatewayGo
set BACKEND_DIR=%ROOT%backend

echo.
echo ============================================================
echo   SSO Login - Build for Linux (amd64)
echo ============================================================
echo.

REM ---- Build backend for Linux ----
echo [INFO] Building backend for linux/amd64 ...
pushd "%BACKEND_DIR%"
if not exist "bin" mkdir bin
set CGO_ENABLED=0
set GOOS=linux
set GOARCH=amd64
call go build -o bin\sso-login-server .\cmd\server
if errorlevel 1 (
    echo [ERROR] backend build failed.
    popd
    pause
    exit /b 1
)
popd
echo [OK] backend built: %BACKEND_DIR%\bin\sso-login-server

REM ---- Build gateway for Linux ----
echo [INFO] Building api_gatewayGo for linux/amd64 ...
pushd "%GATEWAY_DIR%"
if not exist "bin" mkdir bin
set CGO_ENABLED=0
set GOOS=linux
set GOARCH=amd64
call go build -o bin\sso-login-gateway .\cmd\server
if errorlevel 1 (
    echo [ERROR] api_gatewayGo build failed.
    popd
    pause
    exit /b 1
)
popd
echo [OK] api_gatewayGo built: %GATEWAY_DIR%\bin\sso-login-gateway

echo.
echo ============================================================
echo   Build complete!
echo.
echo   Files:
echo     backend\bin\sso-login-server          (linux/amd64)
echo     api_gatewayGo\bin\sso-login-gateway   (linux/amd64)
echo.
echo   Deploy to Linux:
echo     1. Copy bin files to target machine
echo     2. Copy .env files to target machine
echo     3. chmod +x sso-login-server sso-login-gateway
echo     4. Run: ./sso-login-server ^& ./sso-login-gateway
echo ============================================================
echo.
pause
endlocal
