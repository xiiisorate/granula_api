# =============================================================================
# API Gateway - Dockerfile
# =============================================================================
# Multi-stage build for API Gateway
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
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /api-gateway ./cmd/main.go

FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -g 1000 granula && \
    adduser -u 1000 -G granula -s /bin/sh -D granula

WORKDIR /app
COPY --from=builder /api-gateway .
COPY api-gateway/docs/ ./docs/
RUN chown -R granula:granula /app

USER granula
EXPOSE 8080
CMD ["./api-gateway"]
