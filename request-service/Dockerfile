# =============================================================================
# Request Service Dockerfile
# =============================================================================
# Multi-stage build for optimal image size and security.
# Port: 50059
# =============================================================================

# Build stage
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

COPY shared/ ./shared/
COPY request-service/ ./request-service/

WORKDIR /app/request-service
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.Version=${VERSION:-dev}" \
    -o /app/service ./cmd/server

# Production stage
FROM alpine:3.19

# Install runtime dependencies and grpc_health_probe
RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -g 1000 granula && \
    adduser -u 1000 -G granula -s /bin/sh -D granula && \
    wget -qO/bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/v0.4.28/grpc_health_probe-linux-amd64 && \
    chmod +x /bin/grpc_health_probe

WORKDIR /app

COPY --from=builder /app/service .
COPY --from=builder /app/request-service/migrations ./migrations

RUN chown -R granula:granula /app
USER granula

EXPOSE 50059

# Health check using gRPC health protocol
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ["/bin/grpc_health_probe", "-addr=:50059"]

ENTRYPOINT ["./service"]

