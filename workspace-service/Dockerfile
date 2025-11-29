# =============================================================================
# Workspace Service Dockerfile
# =============================================================================
# Multi-stage build for optimal image size and security.
#
# Build: docker build -t granula/workspace-service .
# Run:   docker run -p 50053:50053 granula/workspace-service
#
# =============================================================================

# =============================================================================
# Stage 1: Build
# =============================================================================
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy shared module first (for better caching)
COPY shared/ ./shared/

# Copy service module
COPY workspace-service/ ./workspace-service/

# Build the service
WORKDIR /app/workspace-service
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.Version=${VERSION:-dev}" \
    -o /app/service \
    ./cmd/server

# =============================================================================
# Stage 2: Production
# =============================================================================
FROM alpine:3.19

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -g 1000 granula && \
    adduser -u 1000 -G granula -s /bin/sh -D granula

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/service .

# Copy migrations
COPY --from=builder /app/workspace-service/migrations ./migrations

# Set ownership
RUN chown -R granula:granula /app

# Switch to non-root user
USER granula

# Expose gRPC port
EXPOSE 50053

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:50053/health || exit 1

# Run the service
ENTRYPOINT ["./service"]

