# =============================================================================
# Granula API Makefile (Microservices)
# =============================================================================

.PHONY: help build run test lint docker-build docker-push docker-up docker-down migrate clean proto

# Variables
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
DOCKER_REGISTRY ?= ghcr.io/granula

# Services
SERVICES = api-gateway auth-service user-service workspace-service floor-plan-service \
           scene-service branch-service ai-service compliance-service request-service notification-service

# Go parameters
GOCMD = go
GOBUILD = $(GOCMD) build
GOTEST = $(GOCMD) test
GOGET = $(GOCMD) get
GOMOD = $(GOCMD) mod

# Build flags
LDFLAGS = -ldflags "-w -s -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

# =============================================================================
# Help
# =============================================================================

help: ## Show this help
	@echo "Granula API (Microservices) - Development Commands"
	@echo ""
	@echo "Services: $(SERVICES)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-25s\033[0m %s\n", $$1, $$2}'

# =============================================================================
# Proto Generation
# =============================================================================

proto: ## Generate Go code from proto files
	@echo "Generating proto files..."
	@find shared/proto -name "*.proto" -exec protoc \
		--go_out=shared/gen --go_opt=paths=source_relative \
		--go-grpc_out=shared/gen --go-grpc_opt=paths=source_relative \
		-I shared/proto {} \;
	@echo "Proto generation complete"

proto-clean: ## Clean generated proto files
	rm -rf shared/gen/*

# =============================================================================
# Development
# =============================================================================

run-gateway: ## Run API Gateway
	cd api-gateway && $(GOCMD) run ./cmd/server

run-auth: ## Run Auth Service
	cd auth-service && $(GOCMD) run ./cmd/server

run-user: ## Run User Service
	cd user-service && $(GOCMD) run ./cmd/server

run-workspace: ## Run Workspace Service
	cd workspace-service && $(GOCMD) run ./cmd/server

run-floor-plan: ## Run Floor Plan Service
	cd floor-plan-service && $(GOCMD) run ./cmd/server

run-scene: ## Run Scene Service
	cd scene-service && $(GOCMD) run ./cmd/server

run-branch: ## Run Branch Service
	cd branch-service && $(GOCMD) run ./cmd/server

run-ai: ## Run AI Service
	cd ai-service && $(GOCMD) run ./cmd/server

run-compliance: ## Run Compliance Service
	cd compliance-service && $(GOCMD) run ./cmd/server

run-request: ## Run Request Service
	cd request-service && $(GOCMD) run ./cmd/server

run-notification: ## Run Notification Service
	cd notification-service && $(GOCMD) run ./cmd/server

# =============================================================================
# Build
# =============================================================================

build: build-all ## Build all services

build-all: ## Build all services
	@for service in $(SERVICES); do \
		echo "Building $$service..."; \
		cd $$service && CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) -o ../bin/$$service ./cmd/server && cd ..; \
	done

build-service: ## Build specific service (usage: make build-service SERVICE=auth-service)
	@echo "Building $(SERVICE)..."
	cd $(SERVICE) && CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) -o ../bin/$(SERVICE) ./cmd/server

clean: ## Clean build artifacts
	rm -rf bin/
	rm -rf coverage.out coverage.html
	rm -rf tmp/
	rm -rf shared/gen/

# =============================================================================
# Testing
# =============================================================================

test: ## Run tests
	$(GOTEST) -v -race ./...

test-short: ## Run tests (short mode)
	$(GOTEST) -v -short ./...

test-coverage: ## Run tests with coverage report
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

test-integration: ## Run integration tests
	$(GOTEST) -v -tags=integration ./...

benchmark: ## Run benchmarks
	$(GOTEST) -bench=. -benchmem ./...

# =============================================================================
# Code Quality
# =============================================================================

lint: ## Run linter
	golangci-lint run ./...

lint-fix: ## Run linter with auto-fix
	golangci-lint run --fix ./...

fmt: ## Format code
	$(GOCMD) fmt ./...
	goimports -w .

vet: ## Run go vet
	$(GOCMD) vet ./...

check: fmt vet lint test ## Run all checks

# =============================================================================
# Dependencies
# =============================================================================

deps: ## Download dependencies
	$(GOMOD) download

deps-update: ## Update dependencies
	$(GOMOD) tidy
	$(GOGET) -u ./...

deps-verify: ## Verify dependencies
	$(GOMOD) verify

# =============================================================================
# Code Generation
# =============================================================================

generate: ## Run go generate
	$(GOCMD) generate ./...

mocks: ## Generate mocks
	mockgen -source=internal/domain/repository/user.go -destination=internal/mocks/repository/user_mock.go
	mockgen -source=internal/domain/repository/workspace.go -destination=internal/mocks/repository/workspace_mock.go
	mockgen -source=internal/domain/repository/scene.go -destination=internal/mocks/repository/scene_mock.go
	mockgen -source=internal/domain/repository/branch.go -destination=internal/mocks/repository/branch_mock.go

swagger: ## Generate Swagger documentation
	swag init -g cmd/server/main.go -o docs/swagger --parseInternal

# =============================================================================
# Database
# =============================================================================

migrate-up: ## Run migrations up
	migrate -path migrations/postgres -database "$(POSTGRES_DSN)" up

migrate-down: ## Run migrations down (1 step)
	migrate -path migrations/postgres -database "$(POSTGRES_DSN)" down 1

migrate-down-all: ## Run all migrations down
	migrate -path migrations/postgres -database "$(POSTGRES_DSN)" down

migrate-create: ## Create new migration (usage: make migrate-create NAME=migration_name)
	migrate create -ext sql -dir migrations/postgres -seq $(NAME)

migrate-force: ## Force migration version (usage: make migrate-force VERSION=1)
	migrate -path migrations/postgres -database "$(POSTGRES_DSN)" force $(VERSION)

migrate-version: ## Show current migration version
	migrate -path migrations/postgres -database "$(POSTGRES_DSN)" version

seed: ## Seed database with test data
	$(GOCMD) run ./scripts/seed/main.go

# =============================================================================
# Docker (Microservices)
# =============================================================================

docker-build: ## Build all Docker images
	@for service in $(SERVICES); do \
		echo "Building $$service image..."; \
		docker build \
			--build-arg VERSION=$(VERSION) \
			--build-arg SERVICE_NAME=$$service \
			-t $(DOCKER_REGISTRY)/$$service:$(VERSION) \
			-f $$service/Dockerfile . ; \
		docker tag $(DOCKER_REGISTRY)/$$service:$(VERSION) $(DOCKER_REGISTRY)/$$service:latest; \
	done

docker-build-service: ## Build specific service image (usage: make docker-build-service SERVICE=auth-service)
	docker build \
		--build-arg VERSION=$(VERSION) \
		-t $(DOCKER_REGISTRY)/$(SERVICE):$(VERSION) \
		-f $(SERVICE)/Dockerfile .
	docker tag $(DOCKER_REGISTRY)/$(SERVICE):$(VERSION) $(DOCKER_REGISTRY)/$(SERVICE):latest

docker-push: ## Push all Docker images
	@for service in $(SERVICES); do \
		docker push $(DOCKER_REGISTRY)/$$service:$(VERSION); \
		docker push $(DOCKER_REGISTRY)/$$service:latest; \
	done

docker-push-service: ## Push specific service image (usage: make docker-push-service SERVICE=auth-service)
	docker push $(DOCKER_REGISTRY)/$(SERVICE):$(VERSION)
	docker push $(DOCKER_REGISTRY)/$(SERVICE):latest

docker-up: ## Start all services with Docker Compose
	docker-compose -f docker-compose.yml up -d

docker-up-build: ## Build and start all services
	docker-compose -f docker-compose.yml up -d --build

docker-down: ## Stop all services
	docker-compose -f docker-compose.yml down

docker-logs: ## Show logs for all services
	docker-compose -f docker-compose.yml logs -f

docker-logs-service: ## Show logs for specific service (usage: make docker-logs-service SERVICE=auth-service)
	docker-compose -f docker-compose.yml logs -f $(SERVICE)

docker-ps: ## Show running containers
	docker-compose -f docker-compose.yml ps

docker-clean: ## Clean Docker resources
	docker-compose -f docker-compose.yml down -v --rmi local

docker-restart: ## Restart specific service (usage: make docker-restart SERVICE=auth-service)
	docker-compose -f docker-compose.yml restart $(SERVICE)

# =============================================================================
# Infrastructure
# =============================================================================

infra-up: ## Start only infrastructure (DB, Redis, MinIO)
	docker-compose -f docker-compose.yml up -d postgres mongodb redis minio mailhog

infra-down: ## Stop infrastructure
	docker-compose -f docker-compose.yml stop postgres mongodb redis minio mailhog

infra-logs: ## Show infrastructure logs
	docker-compose -f docker-compose.yml logs -f postgres mongodb redis minio

# =============================================================================
# Migrations (per service)
# =============================================================================

migrate-auth-up: ## Run auth-service migrations
	migrate -path auth-service/migrations -database "postgres://granula:granula_secret@localhost:5432/auth_db?sslmode=disable" up

migrate-user-up: ## Run user-service migrations
	migrate -path user-service/migrations -database "postgres://granula:granula_secret@localhost:5432/users_db?sslmode=disable" up

migrate-workspace-up: ## Run workspace-service migrations
	migrate -path workspace-service/migrations -database "postgres://granula:granula_secret@localhost:5432/workspaces_db?sslmode=disable" up

migrate-floor-plan-up: ## Run floor-plan-service migrations
	migrate -path floor-plan-service/migrations -database "postgres://granula:granula_secret@localhost:5432/floor_plans_db?sslmode=disable" up

migrate-compliance-up: ## Run compliance-service migrations
	migrate -path compliance-service/migrations -database "postgres://granula:granula_secret@localhost:5432/compliance_db?sslmode=disable" up

migrate-request-up: ## Run request-service migrations
	migrate -path request-service/migrations -database "postgres://granula:granula_secret@localhost:5432/requests_db?sslmode=disable" up

migrate-notification-up: ## Run notification-service migrations
	migrate -path notification-service/migrations -database "postgres://granula:granula_secret@localhost:5432/notifications_db?sslmode=disable" up

migrate-all-up: migrate-auth-up migrate-user-up migrate-workspace-up migrate-floor-plan-up migrate-compliance-up migrate-request-up migrate-notification-up ## Run all migrations

# =============================================================================
# Tools Installation
# =============================================================================

tools: ## Install development tools
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/golang/mock/mockgen@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/cosmtrek/air@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@echo "Note: Also install protoc compiler from https://github.com/protocolbuffers/protobuf/releases"

# =============================================================================
# CI/CD
# =============================================================================

ci: deps lint test build ## Run CI pipeline

release: ## Create release
	@echo "Creating release $(VERSION)"
	git tag -a $(VERSION) -m "Release $(VERSION)"
	git push origin $(VERSION)

