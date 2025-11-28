# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy shared module first
COPY shared/ /app/shared/

# Initialize shared module
WORKDIR /app/shared
RUN go mod tidy

# Copy service module
WORKDIR /app/services/auth-service
COPY services/auth-service/go.mod ./
RUN go mod download || true

# Copy source code
COPY services/auth-service/ ./

# Tidy and build
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o /auth-service ./cmd/main.go

# Runtime stage
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /auth-service .

EXPOSE 50051

CMD ["./auth-service"]
