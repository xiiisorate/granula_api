# =============================================================================
# Auth Service - Dockerfile
# =============================================================================
# Multi-stage build for Auth Service
#
# Build: docker build -t granula-auth-service -f auth-service/Dockerfile .
# Run:   docker run -p 50051:50051 granula-auth-service
# =============================================================================

# Stage 1: Build
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy shared module first
COPY shared/ ./shared/

# Initialize shared module
WORKDIR /app/shared
RUN go mod tidy

# Copy service module
WORKDIR /app/auth-service
COPY auth-service/go.mod auth-service/go.sum* ./
RUN go mod download || true

# Copy source code
COPY auth-service/ ./

# Tidy and build
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /auth-service ./cmd/server/main.go

# Stage 2: Runtime
FROM alpine:3.19

# Install runtime dependencies and grpc_health_probe
RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -g 1000 granula && \
    adduser -u 1000 -G granula -s /bin/sh -D granula && \
    wget -qO/bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/v0.4.28/grpc_health_probe-linux-amd64 && \
    chmod +x /bin/grpc_health_probe

WORKDIR /app

# Copy binary
COPY --from=builder /auth-service .

# Set ownership
RUN chown -R granula:granula /app

USER granula

EXPOSE 50051

# Health check using gRPC health protocol
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ["/bin/grpc_health_probe", "-addr=:50051"]

CMD ["./auth-service"]
