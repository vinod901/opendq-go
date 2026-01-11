.PHONY: help build run test clean docker-up docker-down ent-generate deps

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

deps: ## Download dependencies
	go mod download
	go mod tidy

build: ## Build the server binary
	go build -o opendq-server ./cmd/server

run: ## Run the server
	go run ./cmd/server

test: ## Run tests
	go test -v ./...

test-coverage: ## Run tests with coverage
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

clean: ## Clean build artifacts
	rm -f opendq-server
	rm -f coverage.out coverage.html

docker-up: ## Start Docker Compose services
	docker-compose up -d

docker-down: ## Stop Docker Compose services
	docker-compose down

docker-logs: ## View Docker Compose logs
	docker-compose logs -f

ent-generate: ## Generate Ent code
	go generate ./ent

fmt: ## Format code
	go fmt ./...

lint: ## Run linter
	golangci-lint run

vet: ## Run go vet
	go vet ./...

mod-update: ## Update dependencies
	go get -u ./...
	go mod tidy

dev: docker-up ## Start development environment
	@echo "Waiting for services to be ready..."
	@sleep 10
	@echo "Development environment ready!"
	@echo "PostgreSQL: localhost:5432"
	@echo "OpenFGA: http://localhost:8081"
	@echo "Keycloak: http://localhost:8180 (admin/admin)"
	@echo "Marquez: http://localhost:5000"
	@echo "Marquez Web: http://localhost:3001"

install-tools: ## Install development tools
	go install entgo.io/ent/cmd/ent@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
