# API Gateway Go

Go version of the existing Python gateway. It keeps the same public routes:

- `GET /api/health`
- `/api/{service}/{endpoint}` for HTTP and WebSocket traffic

The gateway reads `service.json`, so adding or removing a backend service does not require a code change.

## Run

```powershell
cd Go
go run ./cmd/gateway
```

By default it listens on `:8000` and searches for config in this order:

1. `-config` flag
2. `GATEWAY_CONFIG`
3. `SERVICE_CONFIG`
4. `./service.json`
5. `../service.json`

Examples:

```powershell
go run ./cmd/gateway -addr :8000 -config ..\service.json
$env:PORT = "9000"; go run ./cmd/gateway
```

## Service Config

The current Python format still works:

```json
{
  "O365": "http://127.0.0.1:5000",
  "Reports": "http://127.0.0.1:5100"
}
```

An extended format is also supported when you need to disable a service without deleting it:

```json
{
  "services": {
    "O365": { "url": "http://127.0.0.1:5000", "enabled": true },
    "Reports": { "url": "http://127.0.0.1:5100", "enabled": false }
  }
}
```

## Behavior

- Proxies all HTTP methods that reach `/api/{service}/...`.
- Preserves query strings, headers, cookies, request body, uploads, and response status/body/headers.
- Supports browser CORS preflight requests.
- Supports WebSocket upgrade requests through the same route.
- Hot-reloads service config by checking the file every second by default.

## Build And Test

```powershell
cd Go
go test ./...
go build -o gateway.exe ./cmd/gateway
.\gateway.exe -config ..\service.json
```
