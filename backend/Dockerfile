# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /sso-login-server ./cmd/server

# Run stage
FROM alpine:3.19

WORKDIR /app

# Copy binary from builder
COPY --from=builder /sso-login-server .

# Copy .env file (optional - can be mounted)
COPY .env.example .env

# Expose port
EXPOSE 12080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:12080/health || exit 1

CMD ["./sso-login-server"]
