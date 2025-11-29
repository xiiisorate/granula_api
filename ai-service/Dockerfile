# =============================================================================
# AI Service Dockerfile
# =============================================================================
# AI Service handles floor plan recognition, design variant generation,
# and conversational AI chat using OpenRouter API.
#
# Port: 50057
# Database: MongoDB (ai_db)
# External: OpenRouter API
#
# =============================================================================

# =============================================================================
# Stage 1: Build
# =============================================================================
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copy shared module
COPY shared/ ./shared/

# Copy service module
COPY ai-service/ ./ai-service/

# Build
WORKDIR /app/ai-service
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.Version=${VERSION:-dev}" \
    -o /app/service \
    ./cmd/server

# =============================================================================
# Stage 2: Production
# =============================================================================
FROM alpine:3.19

# Install runtime dependencies and grpc_health_probe
RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -g 1000 granula && \
    adduser -u 1000 -G granula -s /bin/sh -D granula && \
    wget -qO/bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/v0.4.28/grpc_health_probe-linux-amd64 && \
    chmod +x /bin/grpc_health_probe

WORKDIR /app

COPY --from=builder /app/service .

RUN chown -R granula:granula /app
USER granula

# gRPC port
EXPOSE 50057

# Health check using gRPC health protocol
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ["/bin/grpc_health_probe", "-addr=:50057"]

LABEL maintainer="Granula Team"
LABEL description="AI Service - Floor plan recognition, generation, and chat"
LABEL version="${VERSION:-dev}"

ENTRYPOINT ["./service"]

