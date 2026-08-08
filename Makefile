APP_NAME    := myservice
BUILD_DIR   := ./tmp
MAIN        := ./main.go

# === Development ===
.PHONY: dev
dev: ## Start with hot reload (Air)
	@air -c .air.toml

.PHONY: build
build: ## Build binary (with otelc instrumentation)
	@otelc go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN)

.PHONY: build-plain
build-plain: ## Build binary without instrumentation (fast, for dev)
	@go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN)

.PHONY: run
run: build ## Build and run HTTP server
	@$(BUILD_DIR)/$(APP_NAME) serve http

# === Testing ===
.PHONY: test-unit
test-unit: ## Run unit tests only
	@go test -race -short ./internal/domain/...

.PHONY: test-integration
test-integration: ## Run integration tests (requires Docker)
	@go test -race -count=1 ./internal/adapter/...

.PHONY: test-e2e
test-e2e: ## Run E2E tests (requires Docker)
	@go test -race -count=1 ./tests/e2e/...

.PHONY: test-all
test-all: ## Run all tests
	@go test -race -coverprofile=coverage.out ./...

.PHONY: test-cover
test-cover: test-all ## Show coverage report
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# === Code Quality ===
.PHONY: lint
lint: ## Run linters
	@golangci-lint run ./...

.PHONY: fmt
fmt: ## Format code
	@gofumpt -w .
	@goimports -w -local github.com/nuzirwan/go-boilerplate .

.PHONY: vet
vet: ## Run go vet
	@go vet ./...

# === Mocks ===
.PHONY: mocks
mocks: ## Generate mocks with mockgen
	@go generate ./...

# === Migrations ===
.PHONY: migrate-up
migrate-up: build ## Run migrations up
	@$(BUILD_DIR)/$(APP_NAME) migrate up

.PHONY: migrate-down
migrate-down: build ## Rollback last migration
	@$(BUILD_DIR)/$(APP_NAME) migrate down

.PHONY: migrate-create
migrate-create: ## Create new migration (usage: make migrate-create name=create_users)
	@migrate create -ext sql -dir migrations -seq $(name)

# === Docker ===
.PHONY: docker-build
docker-build: ## Build Docker image
	@docker build -t $(APP_NAME):latest -f deployments/Dockerfile .

.PHONY: docker-destroy
docker-destroy: ## Stop infra and remove volumes
	@docker compose -f deployments/docker-compose.yml down -v

# === Proto (gRPC) ===
.PHONY: proto
proto: ## Generate gRPC code from proto files
	@protoc --go_out=. --go-grpc_out=. api/proto/product/v1/product.proto

# === Setup ===
.PHONY: init
init: ## Install dev dependencies
	@go install github.com/air-verse/air@latest
	@go install mvdan.cc/gofumpt@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install go.uber.org/mock/mockgen@latest
	@go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@go install go.opentelemetry.io/auto/otelc@latest
	@echo "Dev dependencies installed"

# === Cleanup ===
.PHONY: clean
clean: ## Remove build artifacts
	@rm -rf $(BUILD_DIR) coverage.out coverage.html

# === Help ===
.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
