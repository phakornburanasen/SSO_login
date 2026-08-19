@echo off
echo Building API Gateway...
cd Go
go build -o ../api_gateway_go.exe ./cmd/gateway/main.go
if %errorlevel% neq 0 (
    echo Build failed!
    pause
    exit /b %errorlevel%
)
echo Build successful!
cd ..
pause
