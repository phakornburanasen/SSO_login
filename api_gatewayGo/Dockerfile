# =========================
# Build Stage
# =========================
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy Go module file
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux \
    go build -o /sso-login-gateway ./cmd/server


# =========================
# Runtime Stage
# =========================
FROM alpine:3.19

WORKDIR /app

# Install CA certificates and wget
# ca-certificates is useful for HTTPS/LDAP/External API connections
RUN apk add --no-cache \
    ca-certificates \
    wget

# Copy binary from builder
COPY --from=builder /sso-login-gateway .

# Copy example environment file
# Production should normally use environment variables or env_file
COPY .env.example .env.example

# API Gateway port
EXPOSE 18000

# Health Check
HEALTHCHECK \
    --interval=30s \
    --timeout=5s \
    --start-period=10s \
    --retries=3 \
    CMD wget --no-verbose \
        --tries=1 \
        --spider \
        http://127.0.0.1:18000/health \
        || exit 1

# Start API Gateway
CMD ["./sso-login-gateway"]