# =============================================================================
# API Gateway - Dockerfile
# =============================================================================
# Multi-stage build for API Gateway (HTTP server)
# =============================================================================

FROM golang:1.23-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates tzdata

COPY shared/ ./shared/
WORKDIR /app/shared
RUN go mod tidy

WORKDIR /app/api-gateway
COPY api-gateway/go.mod api-gateway/go.sum* ./
RUN go mod download || true

COPY api-gateway/ ./
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /api-gateway ./cmd/main.go

FROM alpine:3.19

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata curl && \
    addgroup -g 1000 granula && \
    adduser -u 1000 -G granula -s /bin/sh -D granula

WORKDIR /app
COPY --from=builder /api-gateway .
COPY api-gateway/docs/ ./docs/
RUN chown -R granula:granula /app

USER granula
EXPOSE 3001

# Health check using HTTP (API Gateway is HTTP server)
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:3001/health || exit 1

CMD ["./api-gateway"]
