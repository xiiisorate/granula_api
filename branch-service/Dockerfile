# =============================================================================
# Branch Service Dockerfile
# =============================================================================
# Branch Service manages design branches/variants for scenes,
# enabling version control and A/B comparison of design alternatives.
#
# Port: 50056
# Database: MongoDB (branches_db)
#
# =============================================================================

# =============================================================================
# Stage 1: Build
# =============================================================================
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

COPY shared/ ./shared/
COPY branch-service/ ./branch-service/

WORKDIR /app/branch-service
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

EXPOSE 50056

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:50056/health || exit 1

LABEL maintainer="Granula Team"
LABEL description="Branch Service - Design versioning and branching"
LABEL version="${VERSION:-dev}"

ENTRYPOINT ["./service"]

