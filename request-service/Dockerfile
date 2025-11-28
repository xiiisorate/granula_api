# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy shared module first
COPY shared/ ./shared/

# Copy request-service module
COPY request-service/ ./request-service/

# Build
WORKDIR /app/request-service
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/request-service/main .

# Create non-root user
RUN adduser -D -g '' appuser
USER appuser

EXPOSE 50059

CMD ["./main"]

