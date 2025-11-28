# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy shared module first
COPY shared/ ./shared/

# Copy workspace-service module
COPY workspace-service/ ./workspace-service/

# Build
WORKDIR /app/workspace-service
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/workspace-service/main .

# Create non-root user
RUN adduser -D -g '' appuser
USER appuser

EXPOSE 50053

CMD ["./main"]

