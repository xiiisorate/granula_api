# =============================================================================
# Floor Plan Service Dockerfile
# =============================================================================
# Floor Plan Service handles floor plan upload, storage, and processing.
# Integrates with AI Service for recognition and Scene Service for scene creation.
#
# Port: 50054
# Database: PostgreSQL (floor_plans_db)
# Storage: MinIO/S3
#
# =============================================================================

# =============================================================================
# Stage 1: Build
# =============================================================================
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

COPY shared/ ./shared/
COPY floorplan-service/ ./floorplan-service/

WORKDIR /app/floorplan-service
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
COPY --from=builder /app/floorplan-service/migrations ./migrations

RUN chown -R granula:granula /app
USER granula

EXPOSE 50054

# Health check using gRPC health protocol
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ["/bin/grpc_health_probe", "-addr=:50054"]

LABEL maintainer="Granula Team"
LABEL description="Floor Plan Service - Upload, storage, and processing"
LABEL version="${VERSION:-dev}"

ENTRYPOINT ["./service"]

