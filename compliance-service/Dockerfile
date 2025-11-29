# =============================================================================
# Compliance Service Dockerfile
# =============================================================================
# Compliance Service checks floor plans and scene modifications against
# Russian building codes (SNiP) and Housing Code (ZhK RF).
#
# Port: 50058
# Database: PostgreSQL (compliance_db)
#
# =============================================================================

# =============================================================================
# Stage 1: Build
# =============================================================================
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy shared module first (for better caching)
COPY shared/ ./shared/

# Copy service module
COPY compliance-service/ ./compliance-service/

# Build the service
WORKDIR /app/compliance-service
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

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/service .

# Copy migrations
COPY --from=builder /app/compliance-service/migrations ./migrations

# Set ownership
RUN chown -R granula:granula /app

# Switch to non-root user
USER granula

# Expose gRPC port
EXPOSE 50058

# Health check using gRPC health protocol
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ["/bin/grpc_health_probe", "-addr=:50058"]

# Labels for container metadata
LABEL maintainer="Granula Team"
LABEL description="Compliance Service - Russian building codes validation"
LABEL version="${VERSION:-dev}"

# Run the service
ENTRYPOINT ["./service"]

