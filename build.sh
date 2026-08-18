#!/bin/bash

# ============================================================
#  SSO Login - Build for Linux (native)
#  Output:
#    - backend/bin/sso-login-server
#    - api_gatewayGo/bin/sso-login-gateway
# ============================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GATEWAY_DIR="$SCRIPT_DIR/api_gatewayGo"
BACKEND_DIR="$SCRIPT_DIR/backend"

echo ""
echo "============================================================"
echo "  SSO Login - Build for Linux (amd64)"
echo "============================================================"
echo ""

# ---- Build backend ----
echo "[INFO] Building backend ..."
cd "$BACKEND_DIR"
mkdir -p bin
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/sso-login-server ./cmd/server
echo "[OK] backend built: $BACKEND_DIR/bin/sso-login-server"

# ---- Build gateway ----
echo "[INFO] Building api_gatewayGo ..."
cd "$GATEWAY_DIR"
mkdir -p bin
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/sso-login-gateway ./cmd/server
echo "[OK] api_gatewayGo built: $GATEWAY_DIR/bin/sso-login-gateway"

echo ""
echo "============================================================"
echo "  Build complete!"
echo ""
echo "  Files:"
echo "    backend/bin/sso-login-server"
echo "    api_gatewayGo/bin/sso-login-gateway"
echo ""
echo "  Run:"
echo "    cd backend && ./bin/sso-login-server"
echo "    cd api_gatewayGo && ./bin/sso-login-gateway"
echo "============================================================"
echo ""
