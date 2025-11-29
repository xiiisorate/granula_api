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
FROM golang:1.22-alpine AS builder

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

RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -g 1000 granula && \
    adduser -u 1000 -G granula -s /bin/sh -D granula

WORKDIR /app

COPY --from=builder /app/service .

RUN chown -R granula:granula /app
USER granula

# gRPC port
EXPOSE 50057

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:50057/health || exit 1

LABEL maintainer="Granula Team"
LABEL description="AI Service - Floor plan recognition, generation, and chat"
LABEL version="${VERSION:-dev}"

ENTRYPOINT ["./service"]

