# =============================================================================
# Notification Service - Dockerfile
# =============================================================================
# Multi-stage build for Notification Service
# =============================================================================

FROM golang:1.22-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates tzdata

COPY shared/ ./shared/
WORKDIR /app/shared
RUN go mod tidy

WORKDIR /app/notification-service
COPY notification-service/go.mod notification-service/go.sum* ./
RUN go mod download || true

COPY notification-service/ ./
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /notification-service ./cmd/main.go

FROM alpine:3.19

# Install runtime dependencies and grpc_health_probe
RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -g 1000 granula && \
    adduser -u 1000 -G granula -s /bin/sh -D granula && \
    wget -qO/bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/v0.4.28/grpc_health_probe-linux-amd64 && \
    chmod +x /bin/grpc_health_probe

WORKDIR /app
COPY --from=builder /notification-service .
RUN chown -R granula:granula /app

USER granula
EXPOSE 50060

# Health check using gRPC health protocol
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ["/bin/grpc_health_probe", "-addr=:50060"]

CMD ["./notification-service"]
