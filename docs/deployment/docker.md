# Docker (Микросервисная архитектура)

## Dockerfile (Template для всех сервисов)

```dockerfile
# =============================================================================
# Build stage
# =============================================================================
FROM golang:1.22-alpine AS builder

ARG SERVICE_NAME

RUN apk add --no-cache git ca-certificates tzdata protoc protobuf-dev

WORKDIR /app

# Копируем shared модуль
COPY shared/ ./shared/

# Копируем модуль сервиса
COPY ${SERVICE_NAME}/ ./${SERVICE_NAME}/

WORKDIR /app/${SERVICE_NAME}

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.Version=${VERSION:-dev}" \
    -o /app/service \
    ./cmd/server

# =============================================================================
# Production stage
# =============================================================================
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -g 1000 granula && \
    adduser -u 1000 -G granula -s /bin/sh -D granula

WORKDIR /app

COPY --from=builder /app/service /app/service
COPY --from=builder /app/${SERVICE_NAME}/migrations /app/migrations 2>/dev/null || true

RUN chown -R granula:granula /app
USER granula

EXPOSE 50051

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:50051/health || exit 1

ENTRYPOINT ["/app/service"]
```

## Dockerfile для API Gateway

```dockerfile
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app
COPY shared/ ./shared/
COPY api-gateway/ ./api-gateway/

WORKDIR /app/api-gateway
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /app/gateway \
    ./cmd/server

FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -g 1000 granula && \
    adduser -u 1000 -G granula -s /bin/sh -D granula

WORKDIR /app
COPY --from=builder /app/gateway /app/gateway
RUN chown -R granula:granula /app
USER granula

EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

ENTRYPOINT ["/app/gateway"]
```

## docker-compose.yml (Microservices Development)

```yaml
version: '3.8'

services:
  # ==========================================================================
  # INFRASTRUCTURE
  # ==========================================================================
  
  postgres:
    image: postgres:16-alpine
    ports:
      - "5432:5432"
    environment:
      POSTGRES_USER: granula
      POSTGRES_PASSWORD: granula_secret
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./scripts/init-databases.sql:/docker-entrypoint-initdb.d/init.sql:ro
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U granula"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - granula-network

  mongodb:
    image: mongo:7
    ports:
      - "27017:27017"
    volumes:
      - mongodb_data:/data/db
    networks:
      - granula-network

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    command: redis-server --appendonly yes
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - granula-network

  minio:
    image: minio/minio
    ports:
      - "9000:9000"
      - "9001:9001"
    command: server /data --console-address ":9001"
    environment:
      MINIO_ROOT_USER: minioadmin
      MINIO_ROOT_PASSWORD: minioadmin
    volumes:
      - minio_data:/data
    networks:
      - granula-network

  # ==========================================================================
  # API GATEWAY
  # ==========================================================================
  
  api-gateway:
    build:
      context: .
      dockerfile: api-gateway/Dockerfile
    ports:
      - "8080:8080"
    environment:
      APP_ENV: development
      APP_PORT: "8080"
      REDIS_URL: redis://redis:6379
      JWT_SECRET: dev-jwt-secret-change-in-production
      # gRPC service addresses
      AUTH_SERVICE_ADDR: auth-service:50051
      USER_SERVICE_ADDR: user-service:50052
      WORKSPACE_SERVICE_ADDR: workspace-service:50053
      FLOOR_PLAN_SERVICE_ADDR: floor-plan-service:50054
      SCENE_SERVICE_ADDR: scene-service:50055
      BRANCH_SERVICE_ADDR: branch-service:50056
      AI_SERVICE_ADDR: ai-service:50057
      COMPLIANCE_SERVICE_ADDR: compliance-service:50058
      REQUEST_SERVICE_ADDR: request-service:50059
      NOTIFICATION_SERVICE_ADDR: notification-service:50060
    depends_on:
      - auth-service
      - redis
    networks:
      - granula-network

  # ==========================================================================
  # CORE SERVICES (Developer 1)
  # ==========================================================================
  
  auth-service:
    build:
      context: .
      dockerfile: auth-service/Dockerfile
    ports:
      - "50051:50051"
    environment:
      GRPC_PORT: "50051"
      POSTGRES_DSN: postgres://granula:granula_secret@postgres:5432/auth_db?sslmode=disable
      REDIS_URL: redis://redis:6379
      JWT_SECRET: dev-jwt-secret-change-in-production
      JWT_ACCESS_TTL: 15m
      JWT_REFRESH_TTL: 168h
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    networks:
      - granula-network

  user-service:
    build:
      context: .
      dockerfile: user-service/Dockerfile
    ports:
      - "50052:50052"
    environment:
      GRPC_PORT: "50052"
      POSTGRES_DSN: postgres://granula:granula_secret@postgres:5432/users_db?sslmode=disable
      S3_ENDPOINT: minio:9000
      S3_ACCESS_KEY: minioadmin
      S3_SECRET_KEY: minioadmin
      S3_BUCKET: avatars
      S3_USE_SSL: "false"
    depends_on:
      postgres:
        condition: service_healthy
    networks:
      - granula-network

  workspace-service:
    build:
      context: .
      dockerfile: workspace-service/Dockerfile
    ports:
      - "50053:50053"
    environment:
      GRPC_PORT: "50053"
      POSTGRES_DSN: postgres://granula:granula_secret@postgres:5432/workspaces_db?sslmode=disable
      REDIS_URL: redis://redis:6379
    depends_on:
      postgres:
        condition: service_healthy
    networks:
      - granula-network

  request-service:
    build:
      context: .
      dockerfile: request-service/Dockerfile
    ports:
      - "50059:50059"
    environment:
      GRPC_PORT: "50059"
      POSTGRES_DSN: postgres://granula:granula_secret@postgres:5432/requests_db?sslmode=disable
      REDIS_URL: redis://redis:6379
    depends_on:
      postgres:
        condition: service_healthy
    networks:
      - granula-network

  notification-service:
    build:
      context: .
      dockerfile: notification-service/Dockerfile
    ports:
      - "50060:50060"
    environment:
      GRPC_PORT: "50060"
      POSTGRES_DSN: postgres://granula:granula_secret@postgres:5432/notifications_db?sslmode=disable
      REDIS_URL: redis://redis:6379
      SMTP_HOST: mailhog
      SMTP_PORT: "1025"
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    networks:
      - granula-network

  # ==========================================================================
  # AI/3D SERVICES (Developer 2)
  # ==========================================================================
  
  floor-plan-service:
    build:
      context: .
      dockerfile: floor-plan-service/Dockerfile
    ports:
      - "50054:50054"
    environment:
      GRPC_PORT: "50054"
      POSTGRES_DSN: postgres://granula:granula_secret@postgres:5432/floor_plans_db?sslmode=disable
      S3_ENDPOINT: minio:9000
      S3_ACCESS_KEY: minioadmin
      S3_SECRET_KEY: minioadmin
      S3_BUCKET: floor-plans
      S3_USE_SSL: "false"
      AI_SERVICE_ADDR: ai-service:50057
      SCENE_SERVICE_ADDR: scene-service:50055
    depends_on:
      postgres:
        condition: service_healthy
      minio:
        condition: service_started
    networks:
      - granula-network

  scene-service:
    build:
      context: .
      dockerfile: scene-service/Dockerfile
    ports:
      - "50055:50055"
    environment:
      GRPC_PORT: "50055"
      MONGODB_URI: mongodb://mongodb:27017
      MONGODB_DATABASE: scenes_db
      COMPLIANCE_SERVICE_ADDR: compliance-service:50058
      REDIS_URL: redis://redis:6379
    depends_on:
      - mongodb
    networks:
      - granula-network

  branch-service:
    build:
      context: .
      dockerfile: branch-service/Dockerfile
    ports:
      - "50056:50056"
    environment:
      GRPC_PORT: "50056"
      MONGODB_URI: mongodb://mongodb:27017
      MONGODB_DATABASE: branches_db
      SCENE_SERVICE_ADDR: scene-service:50055
    depends_on:
      - mongodb
    networks:
      - granula-network

  ai-service:
    build:
      context: .
      dockerfile: ai-service/Dockerfile
    ports:
      - "50057:50057"
    environment:
      GRPC_PORT: "50057"
      MONGODB_URI: mongodb://mongodb:27017
      MONGODB_DATABASE: ai_db
      OPENROUTER_API_KEY: ${OPENROUTER_API_KEY}
      OPENROUTER_MODEL: anthropic/claude-sonnet-4
      BRANCH_SERVICE_ADDR: branch-service:50056
    depends_on:
      - mongodb
    networks:
      - granula-network

  compliance-service:
    build:
      context: .
      dockerfile: compliance-service/Dockerfile
    ports:
      - "50058:50058"
    environment:
      GRPC_PORT: "50058"
      POSTGRES_DSN: postgres://granula:granula_secret@postgres:5432/compliance_db?sslmode=disable
    depends_on:
      postgres:
        condition: service_healthy
    networks:
      - granula-network

  # ==========================================================================
  # DEV TOOLS
  # ==========================================================================
  
  mailhog:
    image: mailhog/mailhog
    ports:
      - "1025:1025"
      - "8025:8025"
    networks:
      - granula-network

  # ==========================================================================
  # Redis
  # ==========================================================================
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - granula-network
    restart: unless-stopped

  # ==========================================================================
  # MinIO (S3-compatible storage)
  # ==========================================================================
  minio:
    image: minio/minio:latest
    ports:
      - "9000:9000"
      - "9001:9001"
    environment:
      - MINIO_ROOT_USER=minioadmin
      - MINIO_ROOT_PASSWORD=minioadmin
    volumes:
      - minio_data:/data
    command: server /data --console-address ":9001"
    networks:
      - granula-network
    restart: unless-stopped

  # MinIO bucket initialization
  minio-init:
    image: minio/mc:latest
    depends_on:
      - minio
    entrypoint: >
      /bin/sh -c "
      sleep 5;
      mc alias set myminio http://minio:9000 minioadmin minioadmin;
      mc mb myminio/granula --ignore-existing;
      mc anonymous set download myminio/granula/avatars;
      mc anonymous set download myminio/granula/models;
      mc anonymous set download myminio/granula/previews;
      exit 0;
      "
    networks:
      - granula-network

volumes:
  postgres_data:
  mongodb_data:
  redis_data:
  minio_data:

networks:
  granula-network:
    driver: bridge
```

## docker-compose.prod.yml

```yaml
version: '3.8'

services:
  api:
    image: ghcr.io/granula/api:${VERSION:-latest}
    ports:
      - "8080:8080"
    env_file:
      - .env.production
    deploy:
      replicas: 3
      resources:
        limits:
          cpus: '2'
          memory: 2G
        reservations:
          cpus: '0.5'
          memory: 512M
      restart_policy:
        condition: on-failure
        delay: 5s
        max_attempts: 3
      update_config:
        parallelism: 1
        delay: 10s
        failure_action: rollback
    healthcheck:
      test: ["CMD", "wget", "--spider", "-q", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
    logging:
      driver: "json-file"
      options:
        max-size: "100m"
        max-file: "5"
    networks:
      - granula-network
      - traefik-public
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.granula-api.rule=Host(`api.granula.ru`)"
      - "traefik.http.routers.granula-api.tls=true"
      - "traefik.http.routers.granula-api.tls.certresolver=letsencrypt"
      - "traefik.http.services.granula-api.loadbalancer.server.port=8080"

networks:
  granula-network:
    external: true
  traefik-public:
    external: true
```

## Makefile

```makefile
.PHONY: help build run test lint docker-build docker-push docker-up docker-down migrate

# Variables
VERSION ?= $(shell git describe --tags --always --dirty)
DOCKER_REGISTRY ?= ghcr.io/granula
IMAGE_NAME ?= api

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the application
	go build -ldflags="-w -s -X main.Version=$(VERSION)" -o bin/server ./cmd/server

run: ## Run the application
	go run ./cmd/server

test: ## Run tests
	go test -v -race -cover ./...

test-coverage: ## Run tests with coverage report
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

lint: ## Run linter
	golangci-lint run ./...

docker-build: ## Build Docker image
	docker build -t $(DOCKER_REGISTRY)/$(IMAGE_NAME):$(VERSION) -f deployments/docker/Dockerfile .
	docker tag $(DOCKER_REGISTRY)/$(IMAGE_NAME):$(VERSION) $(DOCKER_REGISTRY)/$(IMAGE_NAME):latest

docker-push: ## Push Docker image
	docker push $(DOCKER_REGISTRY)/$(IMAGE_NAME):$(VERSION)
	docker push $(DOCKER_REGISTRY)/$(IMAGE_NAME):latest

docker-up: ## Start Docker Compose
	docker-compose -f deployments/docker/docker-compose.yml up -d

docker-down: ## Stop Docker Compose
	docker-compose -f deployments/docker/docker-compose.yml down

docker-logs: ## Show Docker logs
	docker-compose -f deployments/docker/docker-compose.yml logs -f api

migrate-up: ## Run database migrations up
	migrate -path migrations/postgres -database "$(POSTGRES_DSN)" up

migrate-down: ## Run database migrations down
	migrate -path migrations/postgres -database "$(POSTGRES_DSN)" down 1

migrate-create: ## Create new migration (usage: make migrate-create NAME=migration_name)
	migrate create -ext sql -dir migrations/postgres -seq $(NAME)

generate-mocks: ## Generate mocks
	mockgen -source=internal/domain/repository/user.go -destination=internal/mocks/repository/user_mock.go
	mockgen -source=internal/domain/repository/workspace.go -destination=internal/mocks/repository/workspace_mock.go
	# ... add more as needed

swagger: ## Generate Swagger documentation
	swag init -g cmd/server/main.go -o docs/swagger

clean: ## Clean build artifacts
	rm -rf bin/
	rm -rf coverage.out coverage.html
```

## CI/CD (GitHub Actions)

```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    
    services:
      postgres:
        image: postgres:16-alpine
        env:
          POSTGRES_USER: test
          POSTGRES_PASSWORD: test
          POSTGRES_DB: test
        ports:
          - 5432:5432
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
      
      mongodb:
        image: mongo:7
        ports:
          - 27017:27017
      
      redis:
        image: redis:7-alpine
        ports:
          - 6379:6379
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'
          cache: true
      
      - name: Install dependencies
        run: go mod download
      
      - name: Run linter
        uses: golangci/golangci-lint-action@v4
        with:
          version: latest
      
      - name: Run tests
        env:
          POSTGRES_HOST: localhost
          POSTGRES_PORT: 5432
          POSTGRES_USER: test
          POSTGRES_PASSWORD: test
          POSTGRES_DB: test
          MONGODB_URI: mongodb://localhost:27017
          REDIS_URL: redis://localhost:6379
        run: go test -v -race -coverprofile=coverage.out ./...
      
      - name: Upload coverage
        uses: codecov/codecov-action@v4
        with:
          files: coverage.out

  build:
    needs: test
    runs-on: ubuntu-latest
    if: github.event_name == 'push'
    
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3
      
      - name: Login to GitHub Container Registry
        uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      
      - name: Build and push
        uses: docker/build-push-action@v5
        with:
          context: .
          file: deployments/docker/Dockerfile
          push: true
          tags: |
            ghcr.io/granula/api:${{ github.sha }}
            ghcr.io/granula/api:latest
          cache-from: type=gha
          cache-to: type=gha,mode=max
```

