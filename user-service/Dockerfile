# =============================================================================
# User Service - Dockerfile
# =============================================================================
# Multi-stage build for User Service
# =============================================================================

FROM golang:1.23-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates tzdata

COPY shared/ ./shared/
WORKDIR /app/shared
RUN go mod tidy

WORKDIR /app/user-service
COPY user-service/go.mod user-service/go.sum* ./
RUN go mod download || true

COPY user-service/ ./
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /user-service ./cmd/main.go

FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -g 1000 granula && \
    adduser -u 1000 -G granula -s /bin/sh -D granula

WORKDIR /app
COPY --from=builder /user-service .
RUN chown -R granula:granula /app

USER granula
EXPOSE 50052
CMD ["./user-service"]
